"""
WordPress.com REST API client — OAuth2 Authorization Code flow.

First run: opens browser, you approve, token cached to ~/.wp_token.json
Later runs: reads cached token, no browser.

Required env vars:
  WP_CLIENT_ID       — from https://developer.wordpress.com/apps/
  WP_CLIENT_SECRET   — from same dashboard (keep private!)
  WP_SITE            — e.g. yoursite.wordpress.com
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
from pathlib import Path
from typing import Any

import requests

REDIRECT_HOST = "127.0.0.1"
REDIRECT_PORT = 8080
REDIRECT_URI = f"http://localhost:{REDIRECT_PORT}/callback"
AUTHORIZE_URL = "https://public-api.wordpress.com/oauth2/authorize"
TOKEN_URL = "https://public-api.wordpress.com/oauth2/token"
TOKEN_FILE = Path.home() / ".wp_token.json"


class _CallbackHandler(http.server.BaseHTTPRequestHandler):
    received_code: str | None = None
    received_error: str | None = None

    def do_GET(self):
        params = urllib.parse.parse_qs(urllib.parse.urlparse(self.path).query)
        if "code" in params:
            _CallbackHandler.received_code = params["code"][0]
            body = b"<h2>OK. You can close this tab.</h2>"
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
        pass  # silence default access log


def _run_oauth_flow(client_id: str, client_secret: str, site: str) -> dict[str, Any]:
    auth_url = f"{AUTHORIZE_URL}?" + urllib.parse.urlencode(
        {
            "client_id": client_id,
            "redirect_uri": REDIRECT_URI,
            "response_type": "code",
            "blog": site,
        }
    )

    server = socketserver.TCPServer((REDIRECT_HOST, REDIRECT_PORT), _CallbackHandler)
    threading.Thread(target=server.serve_forever, daemon=True).start()

    print(f"Opening browser to authorize:\n  {auth_url}\n")
    webbrowser.open(auth_url)
    print("Waiting for callback on http://localhost:8080/callback ... (Ctrl+C to abort)")

    try:
        while _CallbackHandler.received_code is None and _CallbackHandler.received_error is None:
            pass
    finally:
        server.shutdown()
        server.server_close()

    if _CallbackHandler.received_error:
        sys.exit(f"OAuth error: {_CallbackHandler.received_error}")

    code = _CallbackHandler.received_code
    print(f"Got authorization code, exchanging for access token...")

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
        sys.exit(f"Token exchange failed: {resp.status_code} {resp.text}")
    return resp.json()


def get_token(client_id: str, client_secret: str, site: str, *, force: bool = False) -> dict[str, Any]:
    if not force and TOKEN_FILE.exists():
        return json.loads(TOKEN_FILE.read_text())
    data = _run_oauth_flow(client_id, client_secret, site)
    TOKEN_FILE.write_text(json.dumps(data, indent=2))
    TOKEN_FILE.chmod(0o600)
    print(f"Token cached at {TOKEN_FILE}")
    return data


class WPClient:
    def __init__(self, site: str, access_token: str):
        self.base = f"https://public-api.wordpress.com/wp/v2/sites/{site}"
        self.headers = {"Authorization": f"Bearer {access_token}"}

    def _req(self, method: str, path: str, **kw) -> Any:
        url = f"{self.base}{path}"
        r = requests.request(method, url, headers=self.headers, timeout=30, **kw)
        if not r.ok:
            raise RuntimeError(f"{method} {path} → {r.status_code}: {r.text}")
        return r.json() if r.content else None

    def whoami(self) -> dict[str, Any]:
        return self._req("GET", "/users/me", params={"context": "edit"})

    def list_posts(self, per_page: int = 5, status: str = "any") -> list[dict[str, Any]]:
        return self._req("GET", "/posts", params={"per_page": per_page, "status": status})

    def create_post(self, title: str, content: str, status: str = "draft") -> dict[str, Any]:
        return self._req(
            "POST",
            "/posts",
            json={"title": title, "content": content, "status": status},
        )

    def update_post(self, post_id: int, **fields) -> dict[str, Any]:
        return self._req("POST", f"/posts/{post_id}", json=fields)

    def delete_post(self, post_id: int, force: bool = False) -> Any:
        return self._req(
            "DELETE", f"/posts/{post_id}", params={"force": str(force).lower()}
        )


def main() -> None:
    client_id = os.environ.get("WP_CLIENT_ID")
    client_secret = os.environ.get("WP_CLIENT_SECRET")
    site = os.environ.get("WP_SITE")

    if not (client_id and client_secret and site):
        sys.exit("Set required env vars first: WP_CLIENT_ID, WP_CLIENT_SECRET, WP_SITE")

    # 取得參數：標題、內容(或檔案路徑)
    title = sys.argv[1] if len(sys.argv) > 1 else "未命名文章"
    content_input = sys.argv[2] if len(sys.argv) > 2 else "無內容"
    
    if os.path.exists(content_input):
        with open(content_input, "r", encoding="utf-8") as f:
            content = f.read()
    else:
        content = content_input

    token_data = get_token(client_id, client_secret, site)
    client = WPClient(site, token_data["access_token"])

    print(f"正在發布至 {site}...")
    draft = client.create_post(title=title, content=content, status="draft")
    print(f"✅ 成功！文章 ID: {draft['id']} 狀態: {draft['status']}")
    print(f"🔗 連結: {draft.get('link')}")


if __name__ == "__main__":
    main()
