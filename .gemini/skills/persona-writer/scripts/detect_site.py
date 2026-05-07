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


def _classify_seo_plugin(namespaces: list[str]) -> str | None:
    """Map a list of REST namespaces to a known SEO plugin slug.

    Verified live against rankmath.com (the canonical source) — the actual
    namespace is `rankmath/v1` with NO hyphen. The hyphenated form is
    accepted as a defensive alias in case any fork/version uses it.

    Only what we actually integrate is named; everything else returns None.
    """
    ns_set = {n.lower() for n in namespaces}
    if "rankmath/v1" in ns_set or "rank-math/v1" in ns_set:
        return "rankmath"
    if "yoast/v1" in ns_set:
        return "yoast"
    return None


def detect_site(url: str) -> dict[str, Any]:
    parsed = urllib.parse.urlparse(url if "://" in url else f"https://{url}")
    host = (parsed.hostname or url).lower().strip().rstrip("/")
    # Honor caller's scheme so plain-HTTP internal sites are probed correctly
    scheme = parsed.scheme if parsed.scheme in ("http", "https") else "https"
    is_wpcom_subdomain = host.endswith(".wordpress.com")

    info: dict[str, Any] = {
        "host": host,
        "type": "unknown",
        "plan": None,
        "is_atomic": None,
        "wp_json_exposed": False,
        "namespaces": [],
        "seo_plugin": None,
        "recommended_auth": None,
        "api_base": None,
        "notes": [],
    }

    # Probe 1: /wp-json/ on the site itself (most reliable signal)
    try:
        r = requests.get(f"{scheme}://{host}/wp-json/", timeout=15, allow_redirects=True)
        if r.ok and "application/json" in r.headers.get("Content-Type", ""):
            info["wp_json_exposed"] = True
            try:
                root = r.json()
                ns = root.get("namespaces") or []
                if isinstance(ns, list):
                    info["namespaces"] = [str(n) for n in ns]
                    info["seo_plugin"] = _classify_seo_plugin(info["namespaces"])
            except (ValueError, TypeError):
                pass
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
