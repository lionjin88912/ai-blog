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
