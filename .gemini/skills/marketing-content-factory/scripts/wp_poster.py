import requests
from requests.auth import HTTPBasicAuth
import sys
import os
import json
from datetime import date

# --- 配置區 ---
# 優先讀取 wp-config.json，其次環境變數
_SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
_CONFIG_PATH = os.path.join(_SCRIPT_DIR, "wp-config.json")

def _load_config():
    """從 wp-config.json 讀取設定，不存在則退回環境變數。"""
    cfg = {}
    if os.path.exists(_CONFIG_PATH):
        with open(_CONFIG_PATH, "r", encoding="utf-8") as f:
            cfg = json.load(f)
    return {
        "url": cfg.get("WP_URL") or os.getenv("WP_URL", ""),
        "user": cfg.get("WP_USER") or os.getenv("WP_USER", ""),
        "password": cfg.get("WP_APP_PWD") or os.getenv("WP_APP_PWD", ""),
    }

_cfg = _load_config()
WP_URL = _cfg["url"]
WP_USER = _cfg["user"]
WP_APP_PWD = _cfg["password"]
# --------------

def post_to_wordpress(title, content, status="draft"):
    """
    使用 WordPress REST API 建立文章
    status: 'publish' (直接發佈), 'draft' (草稿), 'private' (私用), 'pending' (待審核)
    """
    if not WP_URL.startswith("http"):
        print("錯誤: 請設定正確的 WP_URL (需包含 http:// 或 https://)")
        return

    endpoint = f"{WP_URL.rstrip('/')}/wp-json/wp/v2/posts"
    
    payload = {
        "title": title,
        "content": content,
        "status": status
    }
    
    try:
        response = requests.post(
            endpoint,
            json=payload,
            auth=HTTPBasicAuth(WP_USER, WP_APP_PWD),
            timeout=30
        )
        
        if response.status_code == 201:
            data = response.json()
            print("\n✅ 成功！文章已建立。")
            print(f"文章 ID: {data.get('id')}")
            print(f"文章連結: {data.get('link')}")
            print(f"目前狀態: {data.get('status')}")
            # 記錄到 published.json 供內部連結使用
            _append_published(title, data.get('link', ''), data.get('id'))
        else:
            print(f"\n❌ 失敗。狀態碼: {response.status_code}")
            print(f"回應內容: {response.text}")
            
    except Exception as e:
        print(f"\n☢️ 發生錯誤: {str(e)}")


PUBLISHED_JSON = os.path.join(os.path.dirname(os.path.dirname(__file__)), "published.json")

def _append_published(title, url, post_id):
    """將已發布文章資訊追加到 published.json，供寫新文章時查詢內部連結。"""
    entries = []
    if os.path.exists(PUBLISHED_JSON):
        try:
            with open(PUBLISHED_JSON, "r", encoding="utf-8") as f:
                entries = json.load(f)
        except (json.JSONDecodeError, IOError):
            entries = []

    entries.append({
        "title": title,
        "url": url,
        "id": post_id,
        "date": date.today().isoformat(),
    })

    with open(PUBLISHED_JSON, "w", encoding="utf-8") as f:
        json.dump(entries, f, ensure_ascii=False, indent=2)
    print(f"📝 已更新 {PUBLISHED_JSON}")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("--- WordPress REST API Poster ---")
        print("用法: python wp_poster.py <標題> <內容或檔案路徑> [狀態]")
        print("範例: python wp_poster.py '我的文章' './content.html' publish")
        print("\n* 預設狀態為 'draft' (草稿)。")
    else:
        title = sys.argv[1]
        content_input = sys.argv[2]
        
        # 檢查是否為檔案路徑
        if os.path.exists(content_input):
            with open(content_input, 'r', encoding='utf-8') as f:
                content = f.read()
            print(f"自檔案讀取內容: {content_input}")
        else:
            content = content_input
            
        status = sys.argv[3] if len(sys.argv) > 3 else "draft"
        
        print(f"正在發送文章 (狀態: {status}): {title} ...")
        post_to_wordpress(title, content, status)
