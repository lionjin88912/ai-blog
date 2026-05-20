---
name: persona-writer
description: 通用寫手框架。本身不限定人格,而是搭配 personas/ 底下的人格設定產生內容。負責資料檢索、視覺獲取、寫作排版、SEO 優化、WordPress 發布的標準作業流程,被 marketing-content-factory 調度使用。
---

# 通用寫手 (Persona Writer)

> 這份 skill 是一個**沒有預設人格的寫作引擎**。所有寫作風格、口吻、主題偏好都不寫死在這裡,而是由 `personas/<persona-slug>/persona.md` 提供。
>
> **使用時必須先指定 persona-slug**(例:`mrs-lin-slow-travel`、`mr-wang-foodie`)。

---

## 🧭 載入流程 (Persona Loading)

接到寫作任務時,**第一件事永遠是載入人格**:

1. 確認任務帶了 `persona-slug`(由 marketing-content-factory 傳入,或使用者明確指定)。
2. 用 `read_file` 讀取 `.gemini/skills/persona-writer/personas/<persona-slug>/persona.md`。
3. 解析其中的:
   - `display_name`(用於對話與文章署名)
   - `topic`(主題傾向)
   - 身份、溝通風格、核心金句、文章架構偏好、視覺風格、SEO 風格、適合主題、寫作禁區
4. 在後續所有產出中,**嚴格遵循這份 persona.md 的設定**,不要用其他人格的口吻。

> ⚠️ 找不到 persona-slug 對應資料夾時,請告知使用者並建議使用 `marketing-content-factory` 模組 5 建立新人格,或檢查名稱是否正確。**絕對不要自己編一個人格繼續執行**。

> 🚫 **`personas/_template/` 不是人格,是新人格的範本**。如果任務的 persona-slug 是 `_template`,直接拒絕執行並告知使用者:「`_template` 是範本,不能拿來寫文章。請用實際人格(例如 `mrs-lin-slow-travel`)或建立新人格。」掃描 personas/ 資料夾列出可用人格時,**永遠排除 `_template`**。

---

## 📝 標準作業程序 (Stage-Based Execution)

> 這份 SKILL 過去是「7 步 SOP(Step 0-6)一氣呵成」,**現在改成 `draft.json` 階段接力**。每次被 `marketing-content-factory` 模組 2 委派任務時,**先讀 draft.json 的 `stage` 欄位**,知道現在該做哪一階段,只執行那一階段的工作,再寫回去,結束。

### 🧭 入口:讀 draft.json

每次任務啟動,**第一件事永遠是**:

1. 任務帶來 `persona-slug` + `draft.json` 路徑(由 `marketing-content-factory` 模組 2 提供)
2. `read_file` 讀 draft.json
3. `read_file` 讀 `.gemini/skills/persona-writer/personas/<persona-slug>/persona.md`
4. 依 draft.json 的 `stage` 欄位跳到對應階段

> ⚠️ 找不到 persona-slug 對應資料夾時,告知使用者並建議用 `marketing-content-factory` 模組 5 建立新人格。**絕對不要自己編一個人格繼續執行**。

> 🚫 `persona-slug == _template` 時直接拒絕:「`_template` 是範本,不能拿來寫文章。請用實際人格(例如 `mrs-lin-slow-travel`)或建立新人格。」

### Stage 狀態機

```
init ──research──▶ research_done ──H1──▶ h1_done ──H3──▶ h3_done
  ──H2──▶ h2_done ──FAQ──▶ faq_done ──全文──▶ full_text_done
  ──自動段──▶ published ──清檔──▶ (draft 刪除)
```

每階段的契約如下:

#### Stage `init` → 跑 research

- **讀 draft**:`topic` / `keywords` / `persona_slug`
- **讀 persona.md**:`topic`(主題傾向)、適合主題範例、寫作禁區
- **做**:`google_web_search` 蒐集 ≥3 個有深度的素材(文化背景、交通、特色、地方典故)。素材必須符合 persona.md 的「適合主題範例」與「寫作禁區」
- **寫 draft**:`research = [<素材 1>, <素材 2>, ...]`(string list,每項 1-2 段) + `stage = "research_done"`
- **失敗**:`google_web_search` 拿不到結果 → 回報 marketing-content-factory:「research 失敗,需要使用者提供」(由模組 2 對使用者收手動資料)

#### Stage `research_done` → 生 H1

- **讀 draft**:`topic` / `keywords` / `research`
- **讀 persona.md**:SEO 焦點關鍵字風格、display_name
- **做**:生 1 個 H1 標題(50-60 字元,焦點關鍵字 = keywords[0] 在開頭,`｜` 分隔 sub,` | ` 連 persona.display_name 結尾)
- **不寫 draft**(這階段的 check 由 marketing-content-factory 對使用者做)。把 H1 交回 marketing-content-factory 對話呈現
- **使用者通過後**(由 marketing-content-factory 觸發):`h1 = "<H1>"` + `stage = "h1_done"`

#### Stage `h1_done` → 生 H3 小標

- **讀 draft**:`h1` / `topic` / `research`
- **讀 persona.md**:文章架構偏好、適合主題範例
- **做**:列出整篇能講的點,**5-8 個 H3 小標**(每個 15-25 字,具體不抽象)。**刻意先發散**,Stage `h3_done` → `h2_done` 才歸納
- **使用者通過後**:`h3_subheadings = [<list>]` + `stage = "h3_done"`

#### Stage `h3_done` → 生 H2 大綱(歸納)

- **讀 draft**:`h3_subheadings`
- **讀 persona.md**:文章架構偏好
- **做**:把 H3 小標歸納成 **2-4 個 H2 章節**,**明示哪些 H3 歸到哪個 H2**(讓使用者驗證歸納邏輯)
- **使用者通過後**:`h2_outline = [{"h2": "<標題>", "h3_indices": [<idx>, ...]}, ...]` + `stage = "h2_done"`

#### Stage `h2_done` → 生 FAQ

- **讀 draft**:`topic` / `h1` / `research` / `keywords`
- **讀 persona.md**:溝通風格、文章架構偏好、適合主題範例
- **做**:3 題實務 FAQ,Q1「決定要不要去」、Q2「節奏與時間」、Q3「該人格族群特化問題」。每 A ~120 字,口吻**完全符合該人格**
- **FAQ 區塊命名**:從 persona.md「文章架構偏好」抽(例:林太是「林太的小叮嚀」)。沒寫則 fallback 為「<display_name>的問與答」
- **使用者通過後**:`faq = [{"q": "...", "a": "..."}, ...]` + `stage = "faq_done"`

#### Stage `faq_done` → 寫全文

- **讀 draft**:所有已確認欄位(`h1` / `h3_subheadings` / `h2_outline` / `faq` / `research`)
- **讀 persona.md**:溝通風格、核心金句、文章架構偏好、適合主題範例、寫作禁區
- **做**:寫純文字段落版本的全文。**不附 HTML 標籤、不插圖 placeholder、不附 JSON-LD**
- **格式**:H1 用 `# 標題`、H2 用 `## ...`、H3 用 `### ...`、FAQ 子題用 `**Q1:**` 粗體
- **長度**:約 1500-2000 字,每段 ≤ 150 字,過渡詞 ≥ 30%
- **使用者通過後**:`full_text = "<完整 markdown>"` + `stage = "full_text_done"`

#### Stage `full_text_done` → 自動結尾段

> 這一階段對應 marketing-content-factory 模組 2 的 Step 8,完整流程寫在那邊。persona-writer 只負責執行,不再對使用者問問題。

按以下順序執行:

**8a 找圖**(英文「城市+地點」關鍵字 → request.json → curl Apps Script API → 挑 3-4 張依 persona.md 視覺風格)
**8b 組完整 HTML**(依下方「🎨 HTML 骨架」規範,含 JSON-LD)
**8c LLM 自評 SEO**(焦點關鍵字密度 0.5-1.5%、各分布點、H1/H2 數量、過渡詞、alt 描述。連修 3 次不過 → 直接發 + 警示)
**8d 存 HTML + 跑 wp_poster.py**(預設 `draft`,**絕不**未經使用者明確同意用 `publish`)
**8e 刪 draft.json**(僅當 wp_poster 印 `✅ 成功`時)

成功後:`stage = "published"` → 刪 draft.json(`os.remove`)

失敗時:
- 8a 圖失敗 → 不擋,8b 用無圖純文字版繼續,結尾警示
- 8c 連 3 次不過 → 直接 8d,結尾警示
- 8d 發布失敗 → `stage` 維持 `full_text_done`,**draft 不刪**,把錯誤丟回 marketing-content-factory 對使用者報告(對照模組 4 FAQ)

### 🚫 絕對禁止

- 不要繞過 stage 機制(例如:從 `init` 直接跳到 `full_text_done`)
- 不要在某階段順便做別階段的事(例如:`h1_done` 階段順手也寫 H3)
- 不要在 `_template` persona-slug 上執行任何工作
- 不要用 `publish`(除非使用者明說「直接公開」並由 marketing-content-factory 在 Step 1 記錄下來)

---

## 🎨 HTML 骨架(必須嚴格遵守)

```html
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>焦點關鍵字｜副標題 | <persona.display_name>的XXX</title>
    <meta name="description" content="120-155字元的描述...">
    <meta name="keywords" content="焦點關鍵字,次要關鍵字1,次要關鍵字2">
</head>
<body>
    <!-- 文章內容放這裡 -->

    <!-- JSON-LD 結構化資料放在 body 結尾 -->
    <script type="application/ld+json">...</script>
</body>
</html>
```

> **`<meta name="keywords">` 的用途**:留作搜尋引擎與**未來自動帶 SEO meta**(尚未上線)用。寫 3-5 個焦點 + 次要關鍵字剛好,不要灌超過 7 個(SEO 反效果)。**第一個應該是文章焦點關鍵字**,跟 SEO 標題、第一段呼應。

**⚠️ 禁止事項**:
- 不要把 `<meta>`、`<title>` 放在 `<body>` 或 `<p>` 裡
- 不要用 `<br>` 換行取代正確的 HTML 標籤結構
- 不要在 `<head>` 裡放任何可見內容
- 內容排版用 `div` 容器與內聯樣式 (inline-style),加入 `clear:both;` 防止跑版

---

## 🗂️ 資料夾結構說明

```
persona-writer/
├── SKILL.md                          ← 你正在看的這個檔案(通用 SOP)
├── scripts/
│   └── wp_poster.py                  ← 共用發布腳本(per-persona 讀取設定)
└── personas/
    ├── _template/
    │   ├── persona.md                ← 新人格範本
    │   └── wp-config.example.json    ← WP 設定範例
    ├── mrs-lin-slow-travel/
    │   ├── persona.md                ← 林太人格設定
    │   ├── wp-config.json            ← 林太的 WordPress 連線(機密,.gitignore)
    │   ├── published.json            ← 林太發文紀錄(.gitignore)
    │   └── articles/                 ← 寫文章時自動產生的 HTML 草稿(.gitignore)
    │       └── yyyymmdd-xxx.html
    └── (未來人格.../)
        ├── persona.md
        ├── wp-config.json            ← 各自獨立的 WP 設定
        ├── published.json
        └── articles/
```

**重點**:每個人格資料夾下都有自己的 `wp-config.json`(對應一個 WordPress 部落格),沒有全域共用設定。

---

## ✅ 給 Gemini 的執行自我檢查

每次階段完成前確認:

- [ ] 我有正確讀取 `personas/<slug>/persona.md` 並遵循其中的設定
- [ ] 我有確認 `personas/<slug>/wp-config.json` 存在(該人格的 WordPress 設定)
- [ ] 寫作口吻、主題選材、SEO 風格都符合該人格
- [ ] 內部連結只連到同一人格自己的舊文章(不跨人格)
- [ ] HTML 結構符合骨架規範
- [ ] 發布指令第一個參數是 persona-slug,且預設用 `draft`
- [ ] 完成後 published.json 有更新到正確人格的資料夾下
- [ ] 該人格的文章發到該人格自己的 blog(從 wp-config.json 的 WP_URL 看)
- [ ] **我有讀 draft.json 的 `stage` 欄位**,且只執行該階段該做的事,沒順便做別階段的事
- [ ] **使用者通過後我有把對應欄位寫回 draft + 推進 `stage`**(不要跳階段)
- [ ] 完成後(stage=published)有 `os.remove` 對應的 draft.json,**只在 wp_poster 印 `✅` 時刪**
