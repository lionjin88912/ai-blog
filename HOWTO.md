# HOWTO: AI Blog 部落格 使用指南

## ⬇️ 下載(永遠拿到最新版)

固定連結,不用找哪個檔最新:

- **Windows:** https://github.com/lionjin88912/ai-blog/releases/latest/download/ai-blog-windows-amd64.exe
- 其他平台到 Releases 頁抓:https://github.com/lionjin88912/ai-blog/releases/latest

App 右上角會顯示版本(例 `v1.0.0`),跟 Releases 頁最新版對一下就知道自己是不是最新。

---

## 🚀 推薦入口:Antigravity(agy)

1. 直接雙擊 `ai-blog-windows-amd64.exe`(macOS 執行 `./ai-blog-darwin-arm64`)。**exe 放哪都行**(Downloads、桌面都可),它是純啟動器,不會在旁邊產生任何檔案
2. 首次啟動會自動下載所有工具(Node.js、Antigravity CLI、Python、Git 等),約需 2-3 分鐘
3. 瀏覽器開啟 Web Terminal 後,點畫面上的 **「Launch Antigravity」按鈕**
4. 第一次會問「信任這個資料夾嗎?」→ 選**信任**(skills 才會載入)
5. 在 agy 裡輸入 `/skills` 應看到 5 個能力(marketing-content-factory、persona-writer、tool-status-checker、wordpress-com-rest-api、translate-zh-tw)→ 直接開始說「我要寫一篇行銷文章」

> **所有資料存在固定的使用者資料夾**,不散在 exe 旁邊:
> - Windows:`%LOCALAPPDATA%\ai-blog\`
> - macOS:`~/Library/Application Support/ai-blog/`
>
> 裡面有 skills(`.agents/skills/`,Antigravity 自動載入)、下載的工具(`sandbox/`)、`workspace/`、`GEMINI.md`。**要整包清除就刪這個資料夾;換新版 exe 時這裡的資料原封不動**(skills 演化、帳密、發文紀錄全保留)。

**出問題要回報?** 點網頁右上角「📦 匯出診斷包」,把下載的 zip 傳給工程師即可 — 包內**不含任何密碼**(帳密欄位一律遮罩)。

**要更新系統?** 拿到新版 exe 後,直接跟 Antigravity 說「更新系統」— 只會更新 SOP 與腳本,你的人格、帳密、發文紀錄一律保留(舊 SOP 自動備份到 `.agents/_backup/`)。

---

## Windows

### 啟動

1. 雙擊 `ai-blog-windows-amd64.exe`
2. 首次啟動會自動下載所有工具（Node.js、Gemini CLI、Python、Git 等），約需 2-3 分鐘
3. 下載完成後自動開啟瀏覽器，顯示 Web Terminal

### 啟動 Gemini CLI

在 Web Terminal 裡直接輸入：

```bash
gemini
```

首次使用時，你可以選擇以下任一方式進行驗證：

#### 方法一：使用 Google 帳號（推薦）

直接輸入 `gemini` 後，在選單中選擇 **Sign in with Google**。系統會自動開啟瀏覽器讓你登入 Google 帳號。登入完成後即可開始使用。

*注意：如果你使用的是 Google Workspace 或有付費授權的帳號，可能需要設定專案 ID：`export GOOGLE_CLOUD_PROJECT="YOUR_PROJECT_ID"`*

#### 方法二：使用 API Key

如果你在遠端伺服器或無法使用瀏覽器的環境，可以使用 API Key：

```bash
gemini --api-key YOUR_GEMINI_API_KEY
```

或透過環境變數（在 terminal 裡輸入）：

```bash
export GEMINI_API_KEY=YOUR_GEMINI_API_KEY
gemini
```

### 開始使用(自動準備環境)

Gemini CLI 啟動後,**直接打一句話就好**,例如:

```
如何開始
```

或:

```
我要寫一篇行銷文章
```

Gemini 會在第一次互動時自動完成所有準備工作:
- ✅ 環境檢查(curl / uv / python / cat 版本驗證)
- ✅ 自動建立 `.venv`(如果還沒)
- ✅ 自動掃描並安裝 Python 套件依賴
- ✅ 寫 `TOOLS.md` 記錄狀態
- ✅ 列出可用的能力(skill 清單)
- ✅ 記憶本次環境狀態

完成後直接顯示行銷工廠的 5 項選單,你回一個數字就開始用了。

> 💡 **不需要記任何指令**。如果你是技術人員想手動觸發環境檢查,也可以打 `check env`。

### 圖片分析

在 Web Terminal 裡直接 **Ctrl+V 貼上截圖**,系統會自動儲存並呼叫 Gemini 分析圖片內容。

### CLI 模式

如果不想用瀏覽器，也可以用命令列：

```cmd
ai-blog-windows-amd64.exe shell
```

這會開啟一個 Git Bash 終端（不經過瀏覽器），sandbox 工具都在 PATH 裡。

### 其他指令

```cmd
ai-blog-windows-amd64.exe init      # 設定 API Key 和工作目錄
ai-blog-windows-amd64.exe setup     # 手動下載工具
ai-blog-windows-amd64.exe status    # 顯示已安裝工具版本
ai-blog-windows-amd64.exe web -p 9090  # 指定 port
ai-blog-windows-amd64.exe clean     # 刪除 sandbox 目錄
```

---

## macOS / Linux

### 啟動

```bash
# macOS (Apple Silicon)
chmod +x ai-blog-darwin-arm64
./ai-blog-darwin-arm64

# macOS (Intel)
chmod +x ai-blog-darwin-amd64
./ai-blog-darwin-amd64

# Linux
chmod +x ai-blog-linux-amd64
./ai-blog-linux-amd64
```

首次啟動會自動下載工具，完成後自動開啟瀏覽器。

### 啟動 Gemini CLI

在 Web Terminal 裡輸入：

```bash
gemini
```

首次使用時，你可以選擇以下任一方式進行驗證：

#### 方法一：使用 Google 帳號（推薦）

直接輸入 `gemini` 後，在選單中選擇 **Sign in with Google**。系統會自動開啟瀏覽器讓你登入 Google 帳號。登入完成後即可開始使用。

*注意：如果你使用的是 Google Workspace 或有付費授權的帳號，可能需要設定專案 ID：`export GOOGLE_CLOUD_PROJECT="YOUR_PROJECT_ID"`*

#### 方法二：使用 API Key

設定 API Key 環境變數：

```bash
export GEMINI_API_KEY=YOUR_GEMINI_API_KEY
gemini
```

### 開始使用(自動準備環境)

Gemini CLI 啟動後,**直接打一句話就好**,例如:

```
如何開始
```

或:

```
我要寫一篇行銷文章
```

Gemini 會在第一次互動時自動完成所有準備工作(環境檢查、建立 `.venv`、安裝套件、列出可用能力、記錄環境狀態),完成後直接顯示行銷工廠的 5 項選單,你回一個數字就開始用了。

> 💡 **不需要記任何指令**。如果你是技術人員想手動觸發環境檢查,也可以打 `check env`。

### 使用 zsh 或其他 Shell

預設使用系統的 `$SHELL`（通常是 bash 或 zsh）。也可以手動指定：

```bash
# 用 zsh
./ai-blog-darwin-arm64 web --shell zsh

# 用 fish
./ai-blog-darwin-arm64 web --shell /usr/local/bin/fish

# shell 模式也支援
./ai-blog-darwin-arm64 shell --shell zsh
```

### CLI 模式（不開瀏覽器）

```bash
./ai-blog-darwin-arm64 shell
```

直接在目前的終端開啟一個帶 sandbox PATH 的子 shell。

### 其他指令

```bash
./ai-blog-darwin-arm64 init       # 設定 API Key 和工作目錄
./ai-blog-darwin-arm64 setup      # 手動下載工具
./ai-blog-darwin-arm64 status     # 顯示已安裝工具版本
./ai-blog-darwin-arm64 web -p 9090  # 指定 port
./ai-blog-darwin-arm64 clean      # 刪除 sandbox 目錄
```

---

## 📝 環境裝完了,接下來呢?

如果你是行銷同仁,**直接到 [BOOTSTRAP.md](./BOOTSTRAP.md) 開始用**。

簡而言之,只要在 Gemini CLI 隨便打一句話,例如:

```
如何開始
```

或:

```
我要寫一篇行銷文章
```

Gemini 會自動完成環境準備(第一次需要約 30 秒)→ 列出能力 → 顯示 5 項選單,你回一個數字就開始用了。整個系統設計成「不需要懂技術」,有問題就直接問 Gemini「失敗了幫我看看」即可。

完整使用流程、5 個模組、如何新增寫手人格,都在 BOOTSTRAP.md 有寫。

---

## 🏗️ 系統架構(技術人員看)

```
.agents/skills/
├── marketing-content-factory/        ← 行銷對話入口(L2)
├── persona-writer/                   ← 通用寫手 SOP(L3)
│   ├── scripts/wp_poster.py          ← 共用發布腳本(per-persona 讀取設定)
│   └── personas/
│       ├── _template/                ← 新人格範本(含 wp-config.example.json)
│       └── mrs-lin-slow-travel/      ← 內建林太人格
│           ├── persona.md
│           ├── wp-config.json        ← 林太自己的 WP 連線(.gitignore)
│           └── published.json        ← 林太的發文紀錄(.gitignore)
└── tool-status-checker/              ← 環境檢查(L1)
```

**設計原則:每個人格 = 一個品牌 = 一個 WordPress 部落格**,設定獨立不共用。

新增人格不用寫程式,行銷同仁透過 marketing-content-factory 模組 5 對話 11 題(8 題人格 + 3 題 WordPress)就能生出來。

---

## 常見問題

### Q: Gemini CLI 說找不到 API Key？

你可以選擇以下任一方式：

1. **使用 Google 帳號**：直接輸入 `gemini` 並選擇 **Sign in with Google**。
2. **設定環境變數**：
   ```bash
   export GEMINI_API_KEY=YOUR_KEY
   ```
3. **執行 `ai-blog init`**：透過互動式精靈設定 API Key。

### Q: Windows 上 Gemini 找不到 curl？

確認使用最新版的 `ai-blog`，已內建 Portable Git（含 curl、grep 等 Unix 工具）。

### Q: macOS 提示「無法打開，因為無法驗證開發者」？

```bash
xattr -d com.apple.quarantine ./ai-blog-darwin-arm64
```

### Q: 如何更新工具？

```bash
./ai-blog clean    # 刪除舊的
./ai-blog setup    # 重新下載
```
