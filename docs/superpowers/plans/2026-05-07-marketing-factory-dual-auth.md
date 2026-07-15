# Marketing-Factory Dual-Auth Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the marketing-content-factory + persona-writer pipeline so a marketing user can be onboarded against EITHER a self-hosted WordPress site (Application Password + Basic Auth) OR a WordPress.com hosted site (OAuth2 Bearer token), driven by automatic site-type detection.

**Architecture:** Add a standalone `detect_site()` script that probes a URL and recommends an auth method. Extend `wp_poster.py` to branch internally on a new `auth_method` field in `wp-config.json` while keeping its CLI signature unchanged (so persona-writer/SKILL.md does not need to change). Add a separate interactive `wp_oauth_setup.py` for the one-time browser authorize flow. Rewrite Module 1 of marketing-content-factory/SKILL.md so the Gemini dialog branches on the detection result.

**Tech Stack:** Python 3.12 + `requests`. No new third-party dependencies. Existing `requirements.txt` covers it.

---

## File Structure

| Path | Action | Responsibility |
|---|---|---|
| `.gemini/skills/persona-writer/scripts/detect_site.py` | **CREATE** | Probe URL → return `{type, recommended_auth, api_base, ...}`. Standalone module + CLI (`python detect_site.py <url>`). |
| `.gemini/skills/persona-writer/scripts/wp_poster.py` | **MODIFY** | Add `auth_method` branch. Application Password path unchanged. New OAuth2 path uses `public-api.wordpress.com`. CLI signature unchanged. |
| `.gemini/skills/persona-writer/scripts/wp_oauth_setup.py` | **CREATE** | One-time interactive OAuth flow — opens browser, captures callback on local server, exchanges code → token, writes token into the persona's `wp-config.json`. CLI: `python wp_oauth_setup.py <persona-slug>`. |
| `.gemini/skills/persona-writer/personas/_template/wp-config.example.json` | **MODIFY** | Show both schemas (Application Password vs OAuth2) with `auth_method` field. |
| `.gemini/skills/marketing-content-factory/SKILL.md` | **MODIFY** | Rewrite Module 1 — opens with URL collection + `detect_site` call, branches dialog. Add OAuth-related entries to Module 4 FAQ. |
| `.gemini/skills/persona-writer/SKILL.md` | **NO CHANGE** | wp_poster.py CLI signature unchanged → no edits needed (verified). |

### Schema decision: `wp-config.json`

```json
// Self-hosted / Atomic site
{
  "auth_method": "application_password",
  "WP_URL": "https://your-site.com",
  "WP_USER": "your-login@example.com",
  "WP_APP_PWD": "xxxx xxxx xxxx xxxx xxxx xxxx"
}

// WordPress.com hosted (free / personal / premium)
{
  "auth_method": "oauth2",
  "WP_URL": "https://yoursite.wordpress.com",
  "WP_CLIENT_ID": "139032",
  "WP_CLIENT_SECRET": "...",
  "WP_ACCESS_TOKEN": "...",
  "WP_BLOG_ID": "254693094"
}
```

Backward compat: when `auth_method` is missing AND `WP_APP_PWD` is present, treat as `application_password`. (Existing `mrs-lin-slow-travel/wp-config.json` keeps working without manual edit.)

---

## Task 1: Create `detect_site.py`

**Files:**
- Create: `.gemini/skills/persona-writer/scripts/detect_site.py`

- [ ] **Step 1: Decide expected output for known URLs**

We already empirically verified these on 2026-05-07:

| URL | type | recommended_auth |
|---|---|---|
| `https://liontest-blog.jincoco.site/` | self-hosted | application_password |
| `https://jincoco88912-wxrqt.wordpress.com/` | wordpress.com | oauth2 |
| `https://n8nlife.wordpress.com` | wordpress.com | oauth2 |
| `https://google.com` | unknown | None |

This is the test oracle — the script must produce these classifications.

- [ ] **Step 2: Write `detect_site.py`**

```python
"""
detect_site.py — Probe a WordPress URL and recommend an auth method.

CLI:
    python detect_site.py https://yoursite.com/

Module:
    from detect_site import detect_site
    info = detect_site("https://yoursite.com/")

Returns dict:
    {
      "host": str,
      "type": "wordpress.com" | "wordpress.com-mapped" | "self-hosted" | "unknown",
      "plan": str | None,
      "is_atomic": bool | None,
      "wp_json_exposed": bool,
      "recommended_auth": "oauth2" | "application_password" | None,
      "api_base": str | None,
      "notes": list[str],
    }
"""

from __future__ import annotations

import json
import sys
import urllib.parse
from typing import Any

import requests


def detect_site(url: str) -> dict[str, Any]:
    parsed = urllib.parse.urlparse(url if "://" in url else f"https://{url}")
    host = (parsed.hostname or url).lower().strip().rstrip("/")
    is_wpcom_subdomain = host.endswith(".wordpress.com")

    info: dict[str, Any] = {
        "host": host,
        "type": "unknown",
        "plan": None,
        "is_atomic": None,
        "wp_json_exposed": False,
        "recommended_auth": None,
        "api_base": None,
        "notes": [],
    }

    # Probe 1: /wp-json/ on the site itself (most reliable signal)
    try:
        r = requests.get(f"https://{host}/wp-json/", timeout=15, allow_redirects=True)
        if r.ok and "application/json" in r.headers.get("Content-Type", ""):
            info["wp_json_exposed"] = True
    except requests.RequestException as e:
        info["notes"].append(f"wp-json probe failed: {e}")

    # Probe 2: WordPress.com public-api lookup
    # IMPORTANT: public-api returns 200 with onboarding stubs for ANY domain
    # (e.g. google.com → name="New Onboarding", is_coming_soon=true).
    # Trust it only for *.wordpress.com hosts, OR is_coming_soon=false + jetpack=true.
    pub_data: dict[str, Any] | None = None
    try:
        r = requests.get(
            f"https://public-api.wordpress.com/rest/v1.1/sites/{host}",
            timeout=15,
        )
        if r.ok and "application/json" in r.headers.get("Content-Type", ""):
            pub_data = r.json()
    except requests.RequestException as e:
        info["notes"].append(f"public-api probe failed: {e}")

    if pub_data is not None:
        is_stub = bool(pub_data.get("is_coming_soon")) and not is_wpcom_subdomain
        if is_wpcom_subdomain:
            info["type"] = "wordpress.com"
            plan = pub_data.get("plan")
            if isinstance(plan, dict):
                info["plan"] = plan.get("product_slug")
            info["is_atomic"] = bool(pub_data.get("is_atomic"))
        elif not is_stub and pub_data.get("jetpack"):
            info["is_atomic"] = bool(pub_data.get("is_atomic"))
            info["type"] = (
                "wordpress.com-mapped" if info["is_atomic"] else "self-hosted"
            )
            plan = pub_data.get("plan")
            if isinstance(plan, dict):
                info["plan"] = plan.get("product_slug")

    if info["type"] == "unknown" and info["wp_json_exposed"]:
        info["type"] = "self-hosted"

    if info["type"] in ("wordpress.com", "wordpress.com-mapped"):
        if info["wp_json_exposed"] and info["is_atomic"]:
            info["recommended_auth"] = "application_password"
            info["api_base"] = f"https://{host}/wp-json/wp/v2"
            info["notes"].append(
                "Atomic (Business/Commerce) site — wp-json open, Application Password works"
            )
        else:
            info["recommended_auth"] = "oauth2"
            info["api_base"] = f"https://public-api.wordpress.com/wp/v2/sites/{host}"
            info["notes"].append("WordPress.com hosted — use OAuth2 via public-api")
    elif info["type"] == "self-hosted":
        info["recommended_auth"] = "application_password"
        info["api_base"] = f"https://{host}/wp-json/wp/v2"
        info["notes"].append("Self-hosted WordPress — use Application Password via /wp-json/")
    else:
        info["notes"].append("Could not identify as WordPress; check the URL")

    return info


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python detect_site.py <url>")
        sys.exit(1)
    print(json.dumps(detect_site(sys.argv[1]), indent=2, ensure_ascii=False))
```

- [ ] **Step 3: Verify against 4 known URLs**

Run:
```bash
cd /Users/jincoco/Workspace/wordpress-com-rest-api/gemini_ai_cli
python3 .gemini/skills/persona-writer/scripts/detect_site.py https://liontest-blog.jincoco.site/
python3 .gemini/skills/persona-writer/scripts/detect_site.py https://jincoco88912-wxrqt.wordpress.com/
python3 .gemini/skills/persona-writer/scripts/detect_site.py n8nlife.wordpress.com
python3 .gemini/skills/persona-writer/scripts/detect_site.py google.com
```

Expected:
- liontest-blog → `"type": "self-hosted"`, `"recommended_auth": "application_password"`
- jincoco88912-wxrqt → `"type": "wordpress.com"`, `"recommended_auth": "oauth2"`
- n8nlife → same as above
- google.com → `"type": "unknown"`, `"recommended_auth": null`

- [ ] **Step 4: Commit**

```bash
git add .gemini/skills/persona-writer/scripts/detect_site.py
git commit -m "feat(persona-writer): add detect_site script for URL → auth-method classification"
```

---

## Task 2: Update `wp-config.example.json` template

**Files:**
- Modify: `.gemini/skills/persona-writer/personas/_template/wp-config.example.json`

- [ ] **Step 1: Replace template with two clear examples**

Current contents (single self-hosted example):
```json
{
  "WP_URL": "https://your-wordpress-site.com",
  "WP_USER": "your-email@example.com",
  "WP_APP_PWD": "xxxx xxxx xxxx xxxx xxxx xxxx"
}
```

Replace with the self-hosted form (this file is the actual template that `marketing-content-factory` writes from for `application_password` users; OAuth users get a different shape written by `wp_oauth_setup.py`):

```json
{
  "auth_method": "application_password",
  "WP_URL": "https://your-wordpress-site.com",
  "WP_USER": "your-email@example.com",
  "WP_APP_PWD": "xxxx xxxx xxxx xxxx xxxx xxxx"
}
```

- [ ] **Step 2: Add a second example file showing the OAuth shape (so factory dialog can reference it)**

Create: `.gemini/skills/persona-writer/personas/_template/wp-config.oauth.example.json`

```json
{
  "auth_method": "oauth2",
  "WP_URL": "https://yoursite.wordpress.com",
  "WP_CLIENT_ID": "<from developer.wordpress.com/apps/>",
  "WP_CLIENT_SECRET": "<from same dashboard, keep private>",
  "WP_ACCESS_TOKEN": "<filled in automatically by wp_oauth_setup.py>",
  "WP_BLOG_ID": "<filled in automatically by wp_oauth_setup.py>"
}
```

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/persona-writer/personas/_template/wp-config.example.json \
        .gemini/skills/persona-writer/personas/_template/wp-config.oauth.example.json
git commit -m "feat(template): add auth_method field and OAuth config example"
```

---

## Task 3: Refactor `wp_poster.py` — split into auth_method branches (App-Password path unchanged)

**Files:**
- Modify: `.gemini/skills/persona-writer/scripts/wp_poster.py`

- [ ] **Step 1: Refactor `_load_persona_config` to surface `auth_method`**

Replace the existing `_load_persona_config` with:

```python
def _load_persona_config(persona_slug):
    """Load and normalise the persona's WP config. Returns None if missing.

    Adds an "auth_method" key, defaulting to "application_password" if absent
    (back-compat for configs written before the OAuth feature).
    """
    path = _config_path(persona_slug)
    if not os.path.exists(path):
        return None
    try:
        with open(path, "r", encoding="utf-8") as f:
            cfg = json.load(f)
    except (json.JSONDecodeError, IOError) as e:
        print(f"錯誤: 讀取 {path} 失敗: {e}")
        return None

    auth_method = cfg.get("auth_method")
    if not auth_method:
        # Legacy config (no auth_method field) — infer from contents
        auth_method = "application_password" if cfg.get("WP_APP_PWD") else None

    cfg["auth_method"] = auth_method
    return cfg
```

- [ ] **Step 2: Split `post_to_wordpress` into a dispatcher + Application Password helper**

Replace `post_to_wordpress` with:

```python
def post_to_wordpress(persona_slug, title, content, status="draft"):
    """
    Dispatch to the correct backend based on the persona's auth_method.
    status: 'publish' | 'draft' | 'private' | 'pending'
    """
    if not _persona_exists(persona_slug):
        print(f"錯誤: 找不到人格資料夾 personas/{persona_slug}/")
        print("請先用 marketing-content-factory 模組 5 建立新人格,或確認名稱是否正確。")
        return

    cfg = _load_persona_config(persona_slug)
    if cfg is None:
        print(f"錯誤: 人格「{persona_slug}」尚未設定 WordPress 連線。")
        print("請到 marketing-content-factory 模組 1 設定該人格的 WordPress。")
        return

    auth_method = cfg.get("auth_method")
    if auth_method == "application_password":
        _post_via_app_password(persona_slug, cfg, title, content, status)
    elif auth_method == "oauth2":
        _post_via_oauth2(persona_slug, cfg, title, content, status)
    else:
        print(
            f"錯誤: 人格「{persona_slug}」的 wp-config.json 沒有指定 auth_method, "
            "也無法從欄位推斷。請重新跑模組 1 設定。"
        )


def _post_via_app_password(persona_slug, cfg, title, content, status):
    """Self-hosted WP / Atomic — Basic Auth + Application Password."""
    url = cfg.get("WP_URL", "")
    user = cfg.get("WP_USER", "")
    app_pwd = cfg.get("WP_APP_PWD", "")

    if not url.startswith("http"):
        print("錯誤: 該人格的 WP_URL 不正確,需包含 http:// 或 https://")
        return
    if not (user and app_pwd):
        print("錯誤: 該人格的 wp-config.json 缺 WP_USER 或 WP_APP_PWD。")
        return

    endpoint = f"{url.rstrip('/')}/wp-json/wp/v2/posts"
    payload = {"title": title, "content": content, "status": status}

    try:
        response = requests.post(
            endpoint,
            json=payload,
            auth=HTTPBasicAuth(user, app_pwd),
            timeout=30,
        )
    except Exception as e:
        print(f"\n☢️ 發生錯誤: {str(e)}")
        return

    _report_response(persona_slug, url, title, response)
```

- [ ] **Step 3: Add OAuth2 path (stub that raises clearly until Task 4)**

Add below `_post_via_app_password`:

```python
def _post_via_oauth2(persona_slug, cfg, title, content, status):
    """WordPress.com hosted — OAuth2 Bearer token via public-api.wordpress.com."""
    raise NotImplementedError("OAuth2 publishing arrives in Task 4")
```

- [ ] **Step 4: Extract response handling into `_report_response` (DRY for Task 4)**

Add below `_post_via_oauth2`:

```python
def _report_response(persona_slug, target_url, title, response):
    if response.status_code == 201:
        data = response.json()
        print("\n✅ 成功！文章已建立。")
        print(f"人格: {persona_slug}")
        print(f"目標部落格: {target_url}")
        print(f"文章 ID: {data.get('id')}")
        print(f"文章連結: {data.get('link')}")
        print(f"目前狀態: {data.get('status')}")
        _append_published(persona_slug, title, data.get("link", ""), data.get("id"))
    else:
        print(f"\n❌ 失敗。狀態碼: {response.status_code}")
        print(f"回應內容: {response.text}")
```

- [ ] **Step 5: Verify Application Password path still works (smoke test against existing mrs-lin)**

```bash
python3 .gemini/skills/persona-writer/scripts/wp_poster.py mrs-lin-slow-travel \
  "refactor smoke test - delete me" \
  "<p>Verifying wp_poster refactor preserves Application Password path.</p>" \
  draft
```

Expected: `✅ 成功！文章已建立。` + a draft visible in `https://liontest-blog.jincoco.site/wp-admin/edit.php?post_status=draft`. After confirming, delete the draft from wp-admin.

- [ ] **Step 6: Commit**

```bash
git add .gemini/skills/persona-writer/scripts/wp_poster.py
git commit -m "refactor(wp_poster): branch on auth_method, keep app-password path identical"
```

---

## Task 4: Implement OAuth2 path in `wp_poster.py`

**Files:**
- Modify: `.gemini/skills/persona-writer/scripts/wp_poster.py`

- [ ] **Step 1: Replace the `_post_via_oauth2` stub with the real implementation**

```python
def _post_via_oauth2(persona_slug, cfg, title, content, status):
    """WordPress.com hosted — OAuth2 Bearer token via public-api.wordpress.com."""
    site_url = cfg.get("WP_URL", "")
    access_token = cfg.get("WP_ACCESS_TOKEN", "")

    if not site_url.startswith("http"):
        print("錯誤: 該人格的 WP_URL 不正確,需包含 http:// 或 https://")
        return
    if not access_token:
        print(
            "錯誤: 該人格還沒拿到 access token。請執行:\n"
            f"  python3 .gemini/skills/persona-writer/scripts/wp_oauth_setup.py {persona_slug}"
        )
        return

    # Extract host (e.g. yoursite.wordpress.com) from WP_URL
    from urllib.parse import urlparse
    host = urlparse(site_url).hostname or site_url

    endpoint = f"https://public-api.wordpress.com/wp/v2/sites/{host}/posts"
    payload = {"title": title, "content": content, "status": status}
    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json",
    }

    try:
        response = requests.post(endpoint, json=payload, headers=headers, timeout=30)
    except Exception as e:
        print(f"\n☢️ 發生錯誤: {str(e)}")
        return

    if response.status_code == 401:
        print(
            "\n❌ 失敗:401 未授權。Access token 可能已被撤銷或 client secret 已重置。\n"
            f"請重新跑授權: python3 .gemini/skills/persona-writer/scripts/wp_oauth_setup.py {persona_slug}"
        )
        return

    _report_response(persona_slug, site_url, title, response)
```

- [ ] **Step 2: Verify by code review (no live test yet — needs Task 5 first)**

Read through the full `wp_poster.py` and confirm:
- `post_to_wordpress` dispatches by `auth_method`
- Both branches end at `_report_response` for consistent stdout
- Both branches handle missing config fields with friendly error messages
- Both branches catch network exceptions

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/persona-writer/scripts/wp_poster.py
git commit -m "feat(wp_poster): implement OAuth2 publishing path for wordpress.com sites"
```

---

## Task 5: Create `wp_oauth_setup.py` — interactive OAuth flow

**Files:**
- Create: `.gemini/skills/persona-writer/scripts/wp_oauth_setup.py`

- [ ] **Step 1: Write the script**

```python
"""
wp_oauth_setup.py — One-time interactive OAuth2 setup for a wordpress.com persona.

Reads partial config:
    .gemini/skills/persona-writer/personas/<slug>/wp-config.json
must contain at least: WP_URL, WP_CLIENT_ID, WP_CLIENT_SECRET (auth_method=oauth2).

Flow:
    1. Build authorize URL.
    2. Open default browser to it.
    3. Run a local HTTP server on 127.0.0.1:8080 to catch the callback.
    4. Exchange the returned code for an access token.
    5. Write WP_ACCESS_TOKEN and WP_BLOG_ID back into wp-config.json (mode 0600).

CLI:
    python3 wp_oauth_setup.py <persona-slug>
"""

from __future__ import annotations

import http.server
import json
import os
import socketserver
import sys
import threading
import urllib.parse
import webbrowser

import requests

REDIRECT_HOST = "127.0.0.1"
REDIRECT_PORT = 8080
REDIRECT_URI = f"http://localhost:{REDIRECT_PORT}/callback"
AUTHORIZE_URL = "https://public-api.wordpress.com/oauth2/authorize"
TOKEN_URL = "https://public-api.wordpress.com/oauth2/token"

_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
_SKILL_ROOT = os.path.dirname(_SCRIPT_DIR)
_PERSONAS_DIR = os.path.join(_SKILL_ROOT, "personas")


class _CallbackHandler(http.server.BaseHTTPRequestHandler):
    received_code: str | None = None
    received_error: str | None = None

    def do_GET(self):
        params = urllib.parse.parse_qs(urllib.parse.urlparse(self.path).query)
        if "code" in params:
            _CallbackHandler.received_code = params["code"][0]
            body = b"<h2>OK. You can close this tab and return to Gemini.</h2>"
            self.send_response(200)
        elif "error" in params:
            _CallbackHandler.received_error = params["error"][0]
            body = f"<h2>Error: {params['error'][0]}</h2>".encode()
            self.send_response(400)
        else:
            body = b"<h2>No code in callback.</h2>"
            self.send_response(400)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_args, **_kwargs):
        pass


def _config_path(slug: str) -> str:
    return os.path.join(_PERSONAS_DIR, slug, "wp-config.json")


def main(slug: str) -> None:
    cfg_path = _config_path(slug)
    if not os.path.exists(cfg_path):
        sys.exit(
            f"錯誤: 找不到 {cfg_path}。\n"
            "請先到 marketing-content-factory 模組 1 完成基本資料收集。"
        )
    with open(cfg_path, "r", encoding="utf-8") as f:
        cfg = json.load(f)

    if cfg.get("auth_method") != "oauth2":
        sys.exit("錯誤: 這個人格的 auth_method 不是 'oauth2',不需要跑這個腳本。")

    site_url = cfg.get("WP_URL", "")
    client_id = cfg.get("WP_CLIENT_ID", "")
    client_secret = cfg.get("WP_CLIENT_SECRET", "")
    if not (site_url and client_id and client_secret):
        sys.exit("錯誤: wp-config.json 缺 WP_URL / WP_CLIENT_ID / WP_CLIENT_SECRET。")

    host = urllib.parse.urlparse(site_url).hostname or site_url

    auth_url = f"{AUTHORIZE_URL}?" + urllib.parse.urlencode(
        {
            "client_id": client_id,
            "redirect_uri": REDIRECT_URI,
            "response_type": "code",
            "blog": host,
        }
    )

    server = socketserver.TCPServer((REDIRECT_HOST, REDIRECT_PORT), _CallbackHandler)
    threading.Thread(target=server.serve_forever, daemon=True).start()

    print(f"開啟瀏覽器進行授權:\n  {auth_url}\n")
    print("如果瀏覽器沒自動開啟,複製上面網址手動開。")
    webbrowser.open(auth_url)
    print("等待瀏覽器回呼到 http://localhost:8080/callback ... (Ctrl+C 中止)")

    try:
        while (
            _CallbackHandler.received_code is None
            and _CallbackHandler.received_error is None
        ):
            pass
    finally:
        server.shutdown()
        server.server_close()

    if _CallbackHandler.received_error:
        sys.exit(f"OAuth 錯誤: {_CallbackHandler.received_error}")

    code = _CallbackHandler.received_code
    print("拿到授權碼,換 access token 中...")

    resp = requests.post(
        TOKEN_URL,
        data={
            "client_id": client_id,
            "client_secret": client_secret,
            "code": code,
            "redirect_uri": REDIRECT_URI,
            "grant_type": "authorization_code",
        },
        timeout=30,
    )
    if not resp.ok:
        sys.exit(f"換 token 失敗: {resp.status_code} {resp.text}")
    data = resp.json()

    cfg["WP_ACCESS_TOKEN"] = data["access_token"]
    cfg["WP_BLOG_ID"] = data.get("blog_id", "")
    with open(cfg_path, "w", encoding="utf-8") as f:
        json.dump(cfg, f, indent=2, ensure_ascii=False)
    os.chmod(cfg_path, 0o600)
    print(f"\n✅ Access token 已寫入 {cfg_path} (mode 0600)。")
    print(f"   blog_id={data.get('blog_id')}, blog_url={data.get('blog_url')}")
    print("   後續直接用 wp_poster.py 即可發文,不用再跑這個腳本。")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("用法: python3 wp_oauth_setup.py <persona-slug>")
        sys.exit(1)
    main(sys.argv[1])
```

- [ ] **Step 2: Live test against `jincoco88912-wxrqt.wordpress.com`**

Manual test (one-time, needs human interaction):

a. Create a test persona folder with partial OAuth config:
```bash
mkdir -p .gemini/skills/persona-writer/personas/oauth-test
cat > .gemini/skills/persona-writer/personas/oauth-test/wp-config.json <<'EOF'
{
  "auth_method": "oauth2",
  "WP_URL": "https://jincoco88912-wxrqt.wordpress.com",
  "WP_CLIENT_ID": "139032",
  "WP_CLIENT_SECRET": "<paste fresh secret from developer.wordpress.com/apps/>"
}
EOF
chmod 600 .gemini/skills/persona-writer/personas/oauth-test/wp-config.json
# Also drop a minimal persona.md so wp_poster's _persona_exists passes
echo "# oauth test persona" > .gemini/skills/persona-writer/personas/oauth-test/persona.md
```

b. Run setup:
```bash
python3 .gemini/skills/persona-writer/scripts/wp_oauth_setup.py oauth-test
```

c. Browser opens → approve → return to terminal → see `✅ Access token 已寫入 ...`.

d. Verify token works via the poster:
```bash
python3 .gemini/skills/persona-writer/scripts/wp_poster.py oauth-test \
  "OAuth smoke test - delete me" \
  "<p>OAuth path verification.</p>" \
  draft
```

Expected: `✅ 成功！` + status 201 + draft visible at `https://jincoco88912-wxrqt.wordpress.com/wp-admin/edit.php?post_status=draft`.

e. Cleanup:
```bash
rm -rf .gemini/skills/persona-writer/personas/oauth-test
# Also revoke the token at https://wordpress.com/me/security/connected-applications
```

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/persona-writer/scripts/wp_oauth_setup.py
git commit -m "feat(persona-writer): add wp_oauth_setup.py for one-time OAuth authorize flow"
```

---

## Task 6: Rewrite Module 1 of `marketing-content-factory/SKILL.md`

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (Module 1 section, lines ~49–115)

- [ ] **Step 1: Replace the existing Module 1 with a detect-then-branch version**

Find the current Module 1 (starts at `## 📘 模組 1:設定某個人格的 WordPress 連線`) and replace **the entire module 1 block (from that heading up to the `---` before Module 2)** with:

```markdown
## 📘 模組 1:設定某個人格的 WordPress 連線

> 目標:讓使用者把某個人格的 WordPress 連線資訊填進 `personas/<slug>/wp-config.json`。
> **重要**:**每個人格對應一個 WordPress 部落格,設定獨立不共用**。
> 因應 WordPress 站台型態不同(自架站 vs wordpress.com 託管站),設定流程會分兩種對話路徑,你不需要事先決定走哪條 — Gemini 會用 `detect_site` 幫你判斷。

### 開頭先確認要設定哪個人格

掃描 `personas/` 資料夾,排除 `_template`:
- **只有一個人格**:直接帶他設定那位(通常是林太)
- **多個人格**:問使用者「你要設定哪一位寫手的 WordPress?」並列出選項
- **使用者明確指定**:照他說的(例如「設定王老闆的」就直接設定 mr-wang-foodie)

接下來假設要設定 `<persona-slug>` 那位人格。

### 步驟 1️⃣ 收 URL + 跑偵測

問:
> 把這位寫手要用的 WordPress 網址給我(整段網址,要含 `https://`),例如:
> - 自架站: `https://yourblog.example.com`
> - wordpress.com 站: `https://yourname.wordpress.com`

收到網址後,執行(把 `<URL>` 換成使用者給的):

```
python3 .gemini/skills/persona-writer/scripts/detect_site.py <URL>
```

讀腳本的 JSON 輸出,看 `recommended_auth` 欄位:
- `"application_password"` → **走 Branch A**(自架站 / Atomic 站)
- `"oauth2"` → **走 Branch B**(wordpress.com 託管站)
- `null` (type=`unknown`) → **走 Branch C**(沒辨識到 WordPress)

對使用者只說一句白話:
> 我幫你檢查了一下這個網站,是 [自架的 WordPress / wordpress.com 線上版本],設定方式是 [類型 A / 類型 B]。我帶你一步一步做。

不要把 JSON 細節貼給使用者看。

---

### Branch A — 自架站 / Atomic(Application Password)

**A-1. 教使用者拿應用程式密碼**

如果他說沒有 / 需要教,逐字念:

> 1. 用瀏覽器登入你的 WordPress 後台
> 2. 左邊選單點「使用者」→「個人資料」(或「Profile」)
> 3. 拉到頁面最下面,找到「應用程式密碼 (Application Passwords)」這一區
> 4. 在「新應用程式名稱」隨便打一個名字,例如:`gemini-marketing`
> 5. 按「新增應用程式密碼」
> 6. **畫面會出現一串像 `xxxx xxxx xxxx xxxx xxxx xxxx` 的密碼,把它整串複製起來(包含空格沒關係)**
> 7. 這串密碼只會出現這一次,記得先貼到記事本

**A-2. 收兩項資訊**(網址已經有了)

依序問:
1. 「你的 WordPress 登入帳號 / Email 是?」
2. 「剛才複製的應用程式密碼貼給我」

**A-3. 寫入設定卡**

跟他說「我等一下會把資訊存進『<人格中文名>』的設定卡,可以嗎?」,得到同意後,使用 `write_file` 寫入:

路徑:`.gemini/skills/persona-writer/personas/<persona-slug>/wp-config.json`
內容:
```json
{
  "auth_method": "application_password",
  "WP_URL": "<使用者給的網址>",
  "WP_USER": "<使用者給的帳號>",
  "WP_APP_PWD": "<使用者給的應用程式密碼>"
}
```

**A-4. 跳到「共用驗證步驟」(下方)**

---

### Branch B — wordpress.com 託管站(OAuth2)

**B-1. 跟使用者預告流程**(讓他心裡有底)

> 你的網站是 wordpress.com 線上版本,設定比自架版多一步:
> 我們需要先去 wordpress.com 申請一個「OAuth 應用程式」拿到兩個鑰匙(Client ID、Secret),然後一鍵授權給 Gemini 操作你的部落格。
> 全程約 5 分鐘,我會帶你做。準備好了告訴我「好」。

**B-2. 引導他創 OAuth 應用程式**

> 步驟一:打開瀏覽器到 https://developer.wordpress.com/apps/
> 用你 wordpress.com 的帳號登入,然後點「Create New Application」(建立新的應用程式)。
>
> 步驟二:照下面填表,**逐欄問使用者** — 一次問一格,不要全部攤給他:

| 欄位 | 對使用者怎麼說 | 該填什麼 |
|---|---|---|
| Name | 「應用程式名稱,自己看得懂就好」 | `gemini-<persona-slug>`(例如 `gemini-mrs-lin-slow-travel`)|
| Description | 「描述,隨便寫」 | `Gemini 行銷自動化發文` |
| Website URL | 「網站網址,填你 wordpress.com 站台網址就好」 | 使用者剛才給的 WP_URL |
| Redirect URLs | 「**這個欄位很重要必須填這個**」 | `http://localhost:8080/callback`(逐字念) |
| Javascript Origins | 「這格留空」 | (空) |
| What is 4 + 7? | 驗證機器人的數學題 | `11` |
| Type | 「兩個選項,選 Web」 | Web |

> 填完按 **Create**。會出現一個頁面,上面有 **Client ID**(數字,例如 139032)和 **Client Secret**(一串長亂碼)。
> 把這兩個都複製給我(分兩次貼,一次一個)。

**B-3. 寫入「半完成」設定卡**

收到 Client ID + Client Secret 後,跟他說:
> 收到了 — Client ID 和 Secret 我先記下來(出於安全不會顯示完整內容)。我下一步會幫你開瀏覽器做授權,授權完就可以用了。

使用 `write_file` 寫入(注意 token 兩欄位先空著,會由 wp_oauth_setup 自動補):

路徑:`.gemini/skills/persona-writer/personas/<persona-slug>/wp-config.json`
內容:
```json
{
  "auth_method": "oauth2",
  "WP_URL": "<使用者給的網址>",
  "WP_CLIENT_ID": "<使用者給的 Client ID>",
  "WP_CLIENT_SECRET": "<使用者給的 Client Secret>",
  "WP_ACCESS_TOKEN": "",
  "WP_BLOG_ID": ""
}
```

**B-4. 跑授權腳本**

跟使用者說:
> 我等一下會跑一個腳本,**會自動開瀏覽器**到 wordpress.com 的授權頁,你看到「<人格中文名>」的應用程式想要存取你帳號,**按綠色的 Approve 同意**。
> 然後瀏覽器會跳到「無法連線」的頁面 — 這是正常的,**不要關**,把那個分頁留著就好,Gemini 會自己接到訊號。

執行:
```
python3 .gemini/skills/persona-writer/scripts/wp_oauth_setup.py <persona-slug>
```

腳本會印出授權 URL、開瀏覽器、等回呼、換 token、寫回 wp-config.json。

- 看到 `✅ Access token 已寫入` → 成功,跳到「共用驗證步驟」
- 看到任何錯誤 → 跳到 **模組 4 FAQ → OAuth 相關**

**B-5. 跳到「共用驗證步驟」**

---

### Branch C — 沒辨識到 WordPress

對使用者說:
> 我剛才檢查了那個網址,看起來不是一個可以用 API 操作的 WordPress。可能原因有三個,你看符合哪一種:
>
> 1. 網址打錯了 → 麻煩重新貼正確的網址給我
> 2. 是 WordPress 但站台把 API 關掉了 → 需要請架站工程師打開 REST API
> 3. 根本不是 WordPress(例如是 Wix、Squarespace、Medium…)→ 這個工具暫時沒辦法支援,要請工程師幫忙

請使用者重新給網址,或停下來找工程師處理。

---

### 共用驗證步驟(Branch A 和 B 完成後都要做)

跟他說「我幫『<人格中文名>』發一篇測試草稿確認連線有沒有通,等我 10 秒」,然後執行:

```
python3 .gemini/skills/persona-writer/scripts/wp_poster.py <persona-slug> "連線測試 - 可刪除" "<p>這是 Gemini 自動發送的測試文章,確認連線後可以刪除。</p>" draft
```

- 看到 `✅ 成功`:告訴使用者「『<人格中文名>』的部落格設定完成!以後直接跟我說『用<人格中文名>寫一篇 XXX』就會發到這個部落格。剛才那篇測試草稿你可以登進後台直接刪掉。」
- 看到 `❌ 失敗`:跳到 **模組 4 FAQ** 找對應的錯誤訊息。
```

- [ ] **Step 2: Verify by reading**

Open the modified SKILL.md and trace the dialog as if you were a marketing user with each of:
- A self-hosted URL → does Branch A flow make sense end-to-end?
- A `*.wordpress.com` URL → does Branch B flow get them from "give URL" to "post test draft" without needing to know what OAuth is?
- A non-WP URL (e.g. medium.com) → does Branch C bail out gracefully?

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "feat(factory): rewrite Module 1 with detect_site-driven Branch A/B/C flow"
```

---

## Task 7: Add OAuth FAQ to Module 4 of `marketing-content-factory/SKILL.md`

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (Module 4 section)

- [ ] **Step 1: Locate Module 4**

Search for `## 📘 模組 4` (or whatever heading the FAQ uses — confirm by reading the file).

- [ ] **Step 2: Append OAuth-related FAQ rows**

Add these rows to whatever error table / FAQ format Module 4 uses (preserve existing style):

| 訊息使用者看到 | 真正原因 | 怎麼跟使用者說 / 如何修 |
|---|---|---|
| `❌ 失敗:401 未授權。Access token 可能已被撤銷或 client secret 已重置` | 使用者去 wp.com「已連線應用程式」按了 Disconnect,或在 OAuth app 後台按了 Reset Key | 「鑰匙失效了,我重新跑一次授權即可」→ 執行 `wp_oauth_setup.py <slug>` |
| `OAuth 錯誤: access_denied` | 使用者在授權頁按了拒絕(Decline) | 「你剛才按到拒絕了,我再開一次給你」→ 重跑 `wp_oauth_setup.py <slug>` |
| `OAuth 錯誤: invalid_request` 或 `400 invalid_request` | Client ID 或 Secret 抄錯 / Redirect URLs 沒填 `http://localhost:8080/callback` | 回 OAuth app 設定頁檢查 Redirect URLs 那欄;Client ID/Secret 要使用者重貼一次 |
| `403 unauthorized_client` | OAuth app 的 Type 不是 Web | 重新到 https://developer.wordpress.com/apps/ 開新的應用程式,Type 選 Web |
| `wp_oauth_setup.py` 卡在「等待瀏覽器回呼」久久沒動靜 | 瀏覽器頁開了但使用者忘了按 Approve;或同事網路擋 localhost:8080 | 提醒使用者去檢查瀏覽器是否還在「同意授權」頁面;若是網路問題請工程師看 |

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "docs(factory): add OAuth troubleshooting entries to Module 4 FAQ"
```

---

## Self-Review

Spec coverage check:
- ✅ Detect site type from URL → Task 1
- ✅ Two posting backends (App Password + OAuth2) sharing one CLI → Tasks 3, 4
- ✅ One-time interactive OAuth setup → Task 5
- ✅ wp-config schema supports both → Task 2
- ✅ Module 1 dialog branches automatically based on detection → Task 6
- ✅ Module 4 FAQ covers OAuth-specific errors → Task 7
- ✅ Backward compat for existing `mrs-lin-slow-travel/wp-config.json` → Task 3 Step 1 (legacy inference)
- ✅ AI guides each marketing user through OAuth app creation (per user, no shared secret) → Task 6 Branch B

Type/signature consistency:
- `_load_persona_config` returns full cfg dict (not the previously-trimmed 3-key dict). All callers (`_post_via_app_password`, `_post_via_oauth2`) read with `cfg.get(...)`. Consistent.
- `wp_oauth_setup.py` writes `WP_ACCESS_TOKEN` + `WP_BLOG_ID`; `_post_via_oauth2` reads `WP_ACCESS_TOKEN`. Consistent.
- `detect_site()` returns `recommended_auth` ∈ {`"application_password"`, `"oauth2"`, `None`}. Module 1 Step 1 reads exactly those three values. Consistent.

Placeholder scan: no TBD / "fill in later" / "similar to Task N" — every step has the actual code or command.
