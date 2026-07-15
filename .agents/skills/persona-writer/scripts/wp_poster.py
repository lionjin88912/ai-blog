"""
WordPress REST API Poster (per-persona 設定版)

用法:
  python3 wp_poster.py <persona-slug> <標題> <內容或檔案路徑> [狀態]

範例:
  python3 wp_poster.py mrs-lin-slow-travel "九份慢旅" "./article.html" draft

說明:
  - <persona-slug> 對應 persona-writer/personas/<slug>/ 資料夾名稱
  - WordPress 連線資訊從該人格的 personas/<slug>/wp-config.json 讀取
  - 每個人格對應一個 WordPress 部落格,設定獨立不共用
  - 發布成功後,文章紀錄會寫入該人格資料夾下的 published.json
"""

import requests
from requests.auth import HTTPBasicAuth
import sys
import os
import json
import re
import webbrowser
from datetime import date
from urllib.parse import urlparse

# 解決 Windows 主控台 cp950 編碼不支援 Unicode 字符的錯誤
if sys.platform.startswith('win'):
    try:
        sys.stdout.reconfigure(encoding='utf-8')
        sys.stderr.reconfigure(encoding='utf-8')
    except Exception:
        pass

_SCRIPT_RE = re.compile(r"<script\b[^>]*>.*?</script>", re.DOTALL | re.IGNORECASE)
_HEAD_RE = re.compile(r"<head\b[^>]*>(.*?)</head>", re.DOTALL | re.IGNORECASE)
_TITLE_RE = re.compile(r"<title\b[^>]*>(.*?)</title>", re.DOTALL | re.IGNORECASE)
_META_NAMED_RE = re.compile(
    r'<meta\b[^>]*\bname\s*=\s*["\']([^"\']+)["\'][^>]*\bcontent\s*=\s*["\']([^"\']*)["\'][^>]*/?>',
    re.IGNORECASE,
)


# RESERVED: not currently called. Will be re-wired into post_to_wordpress
# once the companion mu-plugin / Code Snippets bootstrap registers SEO meta
# keys with show_in_rest=true. See commit 55e3ec5 for revert rationale.
# When re-wiring: rank_math_focus_keyword takes only the FIRST comma-split
# entry (Rank Math free version stores a single string; multi-focus is
# Pro-only via rank_math_focus_keyword_2/3/...).
def _extract_seo_from_html(content: str) -> dict[str, str | None]:
    """Pull title / description / keywords out of the article HTML <head>.

    Returns a dict with keys "title", "description", "keywords" — each value
    is a stripped string or None when the corresponding tag is missing /
    empty. The function is permissive: works with partial heads, missing
    tags, single- or double-quoted attributes, and meta tags in either
    name-then-content or content-then-name attribute order (the named regex
    only catches name-first; content-first falls back to the swap pattern).
    """
    seo: dict[str, str | None] = {"title": None, "description": None, "keywords": None}
    if not content:
        return seo

    head_match = _HEAD_RE.search(content)
    head = head_match.group(1) if head_match else content  # fall back to whole doc

    title_match = _TITLE_RE.search(head)
    if title_match:
        title = title_match.group(1).strip()
        seo["title"] = title or None

    for name, value in _META_NAMED_RE.findall(head):
        key = name.strip().lower()
        if key in ("description", "keywords") and seo[key] is None:
            stripped = value.strip()
            seo[key] = stripped or None

    return seo


def _strip_script_tags(content: str) -> str:
    """Remove <script>...</script> blocks before POSTing to WordPress.

    WordPress REST API strips <script> tag boundaries as a security default
    (wp.com is especially strict). When the boundaries are stripped, the JSON
    inside leaks into the visible body. We pre-strip here so the leak never
    happens — JSON-LD belongs in <head> via an SEO plugin, not in post content.

    Limitations (intentional, real risk ~zero for travel / lifestyle prose):
    - Lazy match stops at first `</script>` — `</script>` inside a JS string
      literal would truncate. JSON strings can't contain raw `</script>` so
      this is fine for JSON-LD.
    - HTML entities (`&lt;script&gt;`) are NOT stripped — code-block / docs
      content like `<pre><code>&lt;script&gt;…</code></pre>` survives intact.
    - Only `<script>` is handled. `<style>`, `<iframe>`, etc. are also tag-
      stripped by wp.com but persona-writer doesn't emit them.
    """
    if not content:
        return content
    return _SCRIPT_RE.sub("", content)

# --- 路徑與設定 ---
_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
_SKILL_ROOT = os.path.dirname(_SCRIPT_DIR)  # persona-writer/
_PERSONAS_DIR = os.path.join(_SKILL_ROOT, "personas")


def _persona_dir(persona_slug):
    """回傳指定人格的資料夾路徑。"""
    return os.path.join(_PERSONAS_DIR, persona_slug)


def _config_path(persona_slug):
    """回傳指定人格的 wp-config.json 路徑。"""
    return os.path.join(_persona_dir(persona_slug), "wp-config.json")


def _published_json_path(persona_slug):
    """回傳指定人格的 published.json 路徑。"""
    return os.path.join(_persona_dir(persona_slug), "published.json")


def _persona_exists(persona_slug):
    """檢查人格資料夾是否存在。"""
    return os.path.isdir(_persona_dir(persona_slug))


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

    content = _strip_script_tags(content)

    # Extract only the body content to prevent leading blank spaces and template wrappers from showing up
    body_match = re.search(r"<body\b[^>]*>(.*?)</body>", content, re.DOTALL | re.IGNORECASE)
    if body_match:
        content = body_match.group(1)

    # Clean up leading/trailing whitespaces and newlines
    content = content.strip()

    # Remove any leading <h1> heading (since WordPress already displays the post title automatically)
    content = re.sub(r"^\s*<h1\b[^>]*>.*?</h1>", "", content, flags=re.DOTALL | re.IGNORECASE)
    content = content.strip()

    # Remove any raw markdown headers (like ## or ###) or carriage returns that might have leaked into the HTML
    content = re.sub(r"(?m)^\s*#+\s*", "", content)
    content = re.sub(r"<p>\s*#+\s*", "<p>", content)

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
            f"  python3 .agents/skills/persona-writer/scripts/wp_oauth_setup.py {persona_slug}"
        )
        return

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

    if _looks_like_invalid_token(response):
        print(
            "\n❌ 失敗:Access token 已失效或被撤銷(可能是使用者去後台 Disconnect, "
            "或 client secret 已重置)。\n"
            f"請重新跑授權: python3 .agents/skills/persona-writer/scripts/wp_oauth_setup.py {persona_slug}"
        )
        return

    _report_response(persona_slug, site_url, title, response)


def _looks_like_invalid_token(response):
    """Detect "token revoked / unauthorized" across the variants public-api emits.

    Catches 401 (any body) and 403 with a body that mentions invalid/revoked/
    unauthorized tokens — which is what wp.com returns when scope mismatches
    or the user has Disconnected the OAuth app.
    """
    if response.status_code == 401:
        return True
    if response.status_code == 403:
        body = (response.text or "").lower()
        return any(
            marker in body
            for marker in ("unauthorized", "invalid_token", "authorization_required")
        )
    return False


def _report_response(persona_slug, target_url, title, response):
    if response.status_code == 201:
        data = response.json()
        post_id = data.get("id")
        link = data.get("link", "")
        # 後台編輯 URL pattern 適用所有 WP 站台(自架 / wp.com hosted / Atomic)
        admin_edit_url = f"{target_url.rstrip('/')}/wp-admin/post.php?post={post_id}&action=edit"

        print("\n✅ 成功！文章已建立。")
        print(f"人格: {persona_slug}")
        print(f"目標部落格: {target_url}")
        print(f"文章 ID: {post_id}")
        print(f"後台編輯連結: {admin_edit_url}")
        print(f"前台連結: {link}")
        print(f"目前狀態: {data.get('status')}")
        _append_published(persona_slug, title, link, post_id)

        # 自動開 WordPress 後台編輯分頁讓使用者直接潤稿
        # silent fail:無 GUI / 權限不夠不影響發布成功
        try:
            if webbrowser.open(admin_edit_url):
                print("🌐 已開啟 WordPress 後台編輯分頁。")
        except Exception:
            pass
    else:
        print(f"\n❌ 失敗。狀態碼: {response.status_code}")
        print(f"回應內容: {response.text}")


def _append_published(persona_slug, title, url, post_id):
    """將已發布文章資訊追加到該人格的 published.json,供寫新文章時查詢內部連結。"""
    path = _published_json_path(persona_slug)
    entries = []
    if os.path.exists(path):
        try:
            with open(path, "r", encoding="utf-8") as f:
                entries = json.load(f)
        except (json.JSONDecodeError, IOError):
            entries = []

    entries.append({
        "title": title,
        "url": url,
        "id": post_id,
        "date": date.today().isoformat(),
    })

    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        json.dump(entries, f, ensure_ascii=False, indent=2)
    print(f"📝 已更新 {path}")


if __name__ == "__main__":
    if len(sys.argv) < 4:
        print("--- WordPress REST API Poster (per-persona 設定版) ---")
        print("用法: python3 wp_poster.py <persona-slug> <標題> <內容或檔案路徑> [狀態]")
        print("範例: python3 wp_poster.py mrs-lin-slow-travel '九份慢旅' './article.html' draft")
        print("\n* 預設狀態為 'draft' (草稿)。")
        print("* 每個人格使用 personas/<slug>/wp-config.json 內的設定。")
        sys.exit(1)

    persona_slug = sys.argv[1]
    title = sys.argv[2]
    content_input = sys.argv[3]

    if os.path.exists(content_input):
        with open(content_input, "r", encoding="utf-8") as f:
            content = f.read()
        print(f"自檔案讀取內容: {content_input}")
    else:
        content = content_input

    status = sys.argv[4] if len(sys.argv) > 4 else "draft"

    print(f"正在發送文章 (人格: {persona_slug}, 狀態: {status}): {title} ...")
    post_to_wordpress(persona_slug, title, content, status)
