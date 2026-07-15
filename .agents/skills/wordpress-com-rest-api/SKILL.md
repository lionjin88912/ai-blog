---
name: wordpress-com-rest-api
description: Use when integrating with WordPress.com hosted sites (*.wordpress.com) via REST API, especially when /wp-json returns 404 on a wordpress.com domain, when Basic Auth with Application Password returns 401 "That API call requires authentication against the correct blog" or 403 "An active access token must be used", or when needing to programmatically read/post content on a free or personal plan wordpress.com blog
---

# WordPress.com REST API Integration

## Overview

WordPress.com hosted sites (`*.wordpress.com`), especially on **free / personal plans**, do **NOT** accept Application Passwords with Basic Auth. They require an **OAuth2 Bearer token** routed through `public-api.wordpress.com`.

This is the opposite of self-hosted WordPress (your own server with `/wp-json/` exposed), which DOES accept Application Passwords via Basic Auth as documented at https://developer.wordpress.org/rest-api/.

**Core rule:** Pick the auth method by site type, not by what the official WordPress REST API docs say.

## When to Use

- Site URL ends in `.wordpress.com`
- `https://<site>/wp-json/` returns **404**
- Authenticated requests fail with `401 "correct blog"` or `403 "active access token must be used"` regardless of which username / password combination you try
- You need to POST/PUT (read-only public data does not need auth)

## When NOT to Use

- **Self-hosted WordPress** (`/wp-json/` returns 200): use Basic Auth + Application Password against `https://<your-site>/wp-json/wp/v2/...`
- **WordPress.com Business / Commerce plan with wp-json exposed**: Application Passwords may also work; test `/wp-json/` first
- **Public read-only access**: no auth needed at all — see Quick Reference

## Quick Reference: Auth Method by Site Type

| Site type | API base | Auth |
|---|---|---|
| Self-hosted WordPress | `https://<site>/wp-json/wp/v2/` | Basic Auth + Application Password |
| `*.wordpress.com` public reads | `https://public-api.wordpress.com/wp/v2/sites/<site>/` | None |
| `*.wordpress.com` writes (any plan) | Same as above | **OAuth2 Bearer token** |
| `*.wordpress.com` Business+ with wp-json open | `https://<site>/wp-json/wp/v2/` | Basic Auth + Application Password |

## Diagnose First (2 curls)

```bash
# 1. Is wp-json exposed on the site itself?
curl -sS -o /dev/null -w "%{http_code}\n" "https://<site>/wp-json/"
#   200 → self-hosted style, use Application Password
#   404 → must use public-api + OAuth2

# 2. Is the public-api proxy serving this site?
curl -sS "https://public-api.wordpress.com/wp/v2/sites/<site>" | head -c 200
#   200 → OAuth2 path is open
```

Empirically observed for a free-plan site (`n8nlife.wordpress.com`): `wp-json/` was 404, `public-api.wordpress.com/wp/v2/sites/.../posts` returned 200 for public reads, and every Basic Auth attempt against authenticated endpoints returned 401/403 regardless of credentials — proving the server rejects the auth **method**, not the credentials.

## OAuth2 Authorization Code Flow (the working method)

### Step 1 — Create OAuth App (one-time)

Go to https://developer.wordpress.com/apps/ → **Create New Application**.

Fill in:
- **Name**: anything (shown to user during authorize)
- **Description**: anything
- **Website URL**: any valid URL (your blog URL is fine)
- **Redirect URLs**: `http://localhost:8080/callback` (one URL per line, keep this exact unless you change the Python client to match)
- **Javascript Origins**: leave blank
- **Type**: **Web** (server-side, can keep secret private)

Save. You get **Client ID** (public) and **Client Secret** (keep private).

### Step 2 — User authorizes via browser (one-time)

Open this URL in a browser (replace `<CLIENT_ID>` and `<site>`):

```
https://public-api.wordpress.com/oauth2/authorize?client_id=<CLIENT_ID>&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&blog=<site>
```

Approve. Browser redirects to `http://localhost:8080/callback?code=XXXX`. If nothing is listening on 8080 you will see "connection refused" — that is fine; copy the `code=XXXX` from the address bar.

**The code expires within minutes — exchange it immediately.**

### Step 3 — Exchange code for access token (curl)

```bash
curl -X POST https://public-api.wordpress.com/oauth2/token \
  -d "client_id=<CLIENT_ID>" \
  -d "client_secret=<CLIENT_SECRET>" \
  -d "code=<CODE>" \
  -d "redirect_uri=http://localhost:8080/callback" \
  -d "grant_type=authorization_code"
```

Response:
```json
{
  "access_token": "...",
  "token_type": "bearer",
  "blog_id": "...",
  "blog_url": "http://yoursite.wordpress.com",
  "scope": ""
}
```

Tokens are **long-lived** (no documented expiry on success). Store securely (file mode 0600 or env var).

### Step 4 — Use the Bearer token

```bash
TOKEN="<access_token>"

# Identity / capabilities check
curl -H "Authorization: Bearer $TOKEN" \
  "https://public-api.wordpress.com/wp/v2/sites/<site>/users/me?context=edit"

# Create draft post
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"Hello","content":"posted via OAuth2","status":"draft"}' \
  "https://public-api.wordpress.com/wp/v2/sites/<site>/posts"
```

Successful POST returns `201` with the new post id and link.

## Python Reference Implementation

`wp_client.py` (in this skill folder) is a working end-to-end client:
- First run: opens browser → catches callback on local server → exchanges code → caches token in `~/.wp_token.json` (mode 0600)
- Subsequent runs: reads cached token, no browser

Setup:
```bash
pip install -r requirements.txt
export WP_CLIENT_ID=139032
export WP_CLIENT_SECRET=...        # from OAuth app dashboard
export WP_SITE=yoursite.wordpress.com
python wp_client.py
```

## Common Errors

| HTTP + body | Cause | Fix |
|---|---|---|
| `401 "That API call requires authentication against the correct blog"` | Basic Auth on `public-api.../wp/v2/sites/.../...` | Method is wrong — switch to OAuth2 Bearer |
| `403 "An active access token must be used"` | Basic Auth on `/rest/v1.1/me` | Same — needs Bearer token |
| `404` on `https://<site>/wp-json/` | Free / personal wordpress.com plan, no Jetpack | Use `public-api.wordpress.com` endpoint |
| `400 invalid_request` on `/oauth2/token` | code expired or `redirect_uri` mismatch | Re-run authorize step; redirect_uri must match the OAuth app's whitelist character-for-character |
| `403 unauthorized_client` on token exchange | Wrong client_id / client_secret pair | Re-check OAuth app settings; secret may have been reset |
| Token suddenly stops working | Owner clicked Disconnect or admin Reset Key | Re-run full OAuth flow |

## Security Notes

- **Client Secret** is server-side only (env var, never in git, logs, or shared transcripts)
- **Access tokens** grant full access at the granted scope — store with file mode 0600
- If a token leaks: revoke at https://wordpress.com/me/security/connected-applications (Disconnect the app), or **Reset Key** in the OAuth app dashboard (invalidates ALL tokens issued under that secret)
- WordPress.com does not expose a public token-revoke REST endpoint — use the dashboard / Reset Key

## Why Application Password Fails Here (one-paragraph explanation for AI Q&A)

WordPress core added Application Passwords in WP 5.6 as a Basic Auth credential targeting the `/wp-json/` REST API on a WordPress install. WordPress.com hosted sites on free/personal plans do not expose `/wp-json/` and route REST traffic through `public-api.wordpress.com`, which is a different auth boundary that only accepts OAuth2 Bearer tokens (or signed requests from Jetpack). So even a syntactically valid Application Password can never authenticate there — the server rejects the method before it ever checks the credential, which is why error messages reference "active access token" rather than "invalid password".
