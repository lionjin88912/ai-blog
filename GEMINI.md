# Gemini CLI 全域指令(行銷自動化專案)

> 此檔案會在每次 Gemini 啟動時自動載入,作為全域行為準則。請嚴格遵守以下章節。

---

## 🎬 第一次互動的自動化流程 (First-Touch Auto-Setup)

當使用者**在這個 session 第一次傳訊息給你時**(任何訊息,包含「你好」「如何開始」「幫我寫文章」等),先檢查環境是否準備好,**沒準備好就自動處理,不要要求使用者手動跑指令**。

### Step 1:檢查環境是否就緒

- 看 `./.venv/` 資料夾存在嗎?
- 看 `./TOOLS.md` 檔案存在嗎?

### Step 2A:**環境未就緒** → 執行自動準備流程

如果上述任何一個不存在,這代表使用者剛裝完工具,環境還沒準備好。請依序做這 4 件事:

#### 2A-1. 先親切告知使用者
> 「歡迎!👋 我先幫你準備一下環境,大概 30 秒,完成後就可以開始用了。」

#### 2A-2. 自動呼叫 `tool-status-checker` skill
這個 skill 會:檢查工具版本、建立 `.venv`、掃描 SKILL 腳本依賴、安裝 Python 套件、寫 `TOOLS.md`。
不需要使用者額外指令觸發 — **你直接 activate 它**。

#### 2A-3. 列出已載入的能力給使用者看
等 tool-status-checker 跑完後,列出當前可用的 skill(依你實際看到的為準):

```
📚 已載入的能力:
  ✓ marketing-content-factory  — 行銷內容對話入口(你會直接跟它說話)
  ✓ persona-writer             — 通用寫手 SOP(幕後執行)
  ✓ tool-status-checker        — 環境檢查
  ✓ wordpress-com-rest-api     — WordPress.com REST API 操作
  ✓ translate-zh-tw            — 繁中翻譯
```

#### 2A-4. 把環境狀態記下來(memory)
記住這次 session 的環境狀態,後續對話不需要再驗證。例如:
> 環境記錄:.venv 已建立,TOOLS.md 存在,persona-writer 依賴(requests)已安裝。

#### 2A-5. 進入主流程
做完上述,**不要等使用者再說一次**,直接接著呼叫 `marketing-content-factory` skill,顯示它的 5 項選單給使用者看。讓使用者只需要回一個數字就好。

### Step 2B:**環境已就緒** → 直接進主流程

如果 `.venv/` 跟 `TOOLS.md` 都存在,跳過 Step 2A 全部,直接根據使用者訊息內容路由:

- 訊息跟「行銷/內容/文章/部落格/寫作/發布」有關 → 呼叫 `marketing-content-factory`
- 訊息很模糊(「你好」「能做什麼」「如何開始」) → 呼叫 `marketing-content-factory` 顯示選單
- 訊息明確指定其他事(例:「翻譯這段話」「檢查環境」) → 用對應 skill

---

## 🎯 路由規則 (Routing Rules)

| 使用者訊息類型 | 該用的 skill |
|---|---|
| 「如何開始」「你能做什麼」「help」「你好」 | `marketing-content-factory`(顯示選單) |
| 「設定 WordPress」「綁帳號」「重新設定」「設定林太/王老闆的 WP」 | `marketing-content-factory` 模組 1 |
| 「我要寫文章」「幫我發 blog」「用林太/王老闆寫 XXX」「寫一篇關於 XXX」 | `marketing-content-factory` 模組 2 |
| 「指令範例」「我可以怎麼跟你說」「給我句子範例」 | `marketing-content-factory` 模組 3 |
| 「失敗」「錯誤」「跑不出來」「不能用」 | `marketing-content-factory` 模組 4 |
| 「新增寫手」「建立人格」「我想多一個人格」「我想做美食家」 | `marketing-content-factory` 模組 5 |
| 「修改 XXX 的口吻」「林太太正經」「調整人格」「改活潑一點」 | `marketing-content-factory` 模組 6 |
| 「檢查環境」「check env」「裝套件」 | `tool-status-checker` |

---

## 📋 專案環境規格 (Environment Specs)

| 工具 | 版本要求 | 用途 |
|---|---|---|
| Python | 3.12.10 | 主要執行環境 |
| uv | 0.7.2+ | Python 套件管理 |
| curl | 8.12.1+ | 網路與 API 呼叫 |
| Git | Portable | 版本控制 |

- **跨平台**:Windows / macOS / Linux 三平台
- **Python VENV**:位於 `./.venv`
- **Sandbox**:位於 `./sandbox`(自動下載的工具)

---

## ⚠️ 重要行為準則

1. **發布到 WordPress 一律預設 `draft`**,絕對不要在使用者沒明確說「直接公開」時用 `publish`。
2. **每個人格對應一個 WordPress 部落格**,設定獨立不共用,不要在人格之間借用 wp-config.json。
3. **派任務寫文章時,呼叫 `persona-writer` 並傳入 persona-slug**,不要直接執行 SOP。
4. **WordPress 帳密、應用程式密碼**:可以在對話中收集並寫進 `personas/<slug>/wp-config.json`,但**絕對不要顯示在訊息中**(收到後預覽時用「已收到(出於安全不顯示)」代替)。
5. **使用者不需要懂技術**,任何技術名詞(JSON、API、endpoint…)在對話中都要先翻成白話。
6. **Auto-Git 政策**:只有兩種情況自動 commit,其他狀況不要主動 commit:
   - 模組 5 建立新人格成功後 → `git add personas/<slug>/persona.md` + commit
   - 模組 6 修改現有人格後 → 同上
   - **絕對只 add 指定的 persona.md**,不要 `git add .` 或 `git add -A`(避免帶進 wp-config.json 之類機密)
   - `.git/` 不存在時靜默跳過(不要報錯嚇行銷)

## 🎨 輸出排版規範 (Output Formatting Rules)

當你需要向使用者列出「建議選項」、「操作選單」、「可用的指令」或「範例問法」時，**請嚴格遵守以下格式規範，禁止使用標準的 Markdown 縮排**：

1. **選項列表固定格式**：
   一律使用 **6 個半形空白 + 1 個減號 + 1 個半形空白** (`      - `) 作為開頭。

2. **正確範例 (Must Follow)**：
   ```text
   📝 你可以選擇以下操作：
      - 「幫我用林太寫一篇關於日月潭的兩天一夜文章」
      - 「幫我重新設定 WordPress 連線」
      - 「我之前寫過哪些文章？」
   ```
---

## 📂 專案結構 (Project Structure)

本專案 = 行銷內容自動化(skills)+ AI Sandbox CLI 工具原始碼(sandbox-src),架構如下:

```text
/
├── .agents/skills/        # 全部 agent skills(Antigravity CLI 會自動載入這裡)
│   ├── marketing-content-factory/
│   ├── persona-writer/    # 含 scripts/wp_poster.py 與 personas/<slug>/
│   ├── tool-status-checker/
│   ├── wordpress-com-rest-api/
│   └── translate-zh-tw/
├── sandbox-src/           # ai-sandbox 工具的 Go 原始碼
│   ├── cmd/               # CLI 指令(cobra)
│   ├── internal/
│   │   ├── config/        # 設定檔處理
│   │   ├── toolchain/     # Node/Python/Git/Antigravity CLI/Copilot/uv 下載安裝
│   │   ├── web/           # 網頁終端(含 Launch Antigravity 按鈕)
│   │   └── wizard/        # 互動式安裝選單
│   ├── main.go / go.mod / Makefile / Dockerfile
│   └── docker-compose.yml # 跨平台編譯
├── ai-sandbox-*(.exe/.bat) # 編譯好的執行檔(業務雙擊入口)
├── docs/                  # 專案文件(_archive/ 為歸檔的舊版 skill)
├── GEMINI.md              # 本檔案,agent 啟動時自動載入
└── HOWTO.md / BOOTSTRAP.md / README.md
```

**Skill 位置鐵則:skills 只放 `.agents/skills/` 一個地方。** 不要複製到其他路徑(歷史教訓:`.gemini/antigravity-cli/skills/`、`.antigravitycli/` 都是猜錯路徑的失敗複製,已清除)。
