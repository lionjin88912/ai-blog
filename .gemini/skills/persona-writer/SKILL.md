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

## 📝 標準作業程序 (Standard Operating Procedure)

人格載入完成後,依序執行以下 6 步:

### 0. 環境檢查 (Pre-flight Check)

確認該人格的 WordPress 設定存在:`.gemini/skills/persona-writer/personas/<persona-slug>/wp-config.json`。
- **存在**:該人格的 WordPress 發布功能已就緒,繼續。
- **不存在**:中止流程,告訴使用者「人格『<display_name>』還沒設定 WordPress,請先到 marketing-content-factory 設定」。**不要在這裡引導設定**,設定流程歸入口層管。

> 💡 **每個人格對應一個 WordPress 部落格**,設定獨立不共用。林太發到她的 lifestyle blog,王老闆發到他的美食 blog,各走各的。

### 1. 資料檢索 (Research Phase)

- **工具**:呼叫 `google_web_search`。
- **目標**:依 persona.md 的 `topic` 與當次任務主題,搜尋背景資料、文化典故、交通與相關實用資訊。
- **要求**:蒐集至少 3 個有深度的細節作為寫作素材。**素材的選材傾向必須符合 persona.md 的「適合主題範例」與「寫作禁區」**。

### 2. 視覺獲取 (Visual Acquisition Phase)

分兩步完成,**絕對不要在 shell 指令中內嵌 JSON**(PowerShell 會破壞引號與編碼)。

**Step 1**:使用 `write_file` 工具建立 `request.json`:
```json
{"locations":"KEYWORDS"}
```
將 `KEYWORDS` 替換為相關地點的**英文關鍵字**,多個以逗號分隔。
**關鍵字格式必須是「城市+地點」**,避免模糊搜尋導致圖片與文章不符。
- ✅ 正確:`Taichung Miyahara,Taichung Station,Taichung City Hall`
- ❌ 錯誤:`Miyahara,Station,City Hall`

**Step 2**:執行 shell 指令呼叫 API:
```
curl --location "https://script.google.com/macros/s/AKfycbxlQSTNpSifs9t6gt-0QNYPuE8ui3dXn7O6v7akOby0gwLR6EVBlrb_CQhGSajpYo30/exec" -H "Content-Type: application/json" -d @request.json
```
> Windows 注意:若 `curl` 被 PowerShell alias 攔截,改用 `curl.exe`。

**⚠️ 禁止事項**:不要用 `--data '{"locations":"..."}'` 內嵌 JSON。

**選圖原則**:從 API 回傳的圖片中挑 3-4 張,**選圖標準依 persona.md 的「視覺風格偏好」**(例如林太偏好寧靜雜誌風,王老闆可能偏好熱鬧夜市感)。

### 3. 創作 (Writing Phase)

依 persona.md 的「文章架構偏好」與「溝通風格」撰寫。預設骨架:
- **開頭**:依 persona.md 的開頭風格切入。
- **中段**:結合 Research Phase 的素材,撰寫至少 3 個段落,每段約 200 字。
- **結尾**:依 persona.md 的結尾風格收束。

**視覺整合**:從 `blog-visual-styles` 選用 2-3 種風格(雜誌封面、全景雙圖、圓形提示、背景引言)進行 HTML 排版。具體選用哪種,看 persona.md 的視覺風格偏好。

### 4. SEO 深度優化 (SEO Review Phase)

確保通過 **Yoast SEO** 檢測:

**4.1 焦點關鍵字 (Focus Keyphrase)**
- 從文章主題 + persona.md 的「SEO 焦點關鍵字風格」提取 1 組核心關鍵詞組(2-4 字)。
- 焦點關鍵字必須出現在:
  - SEO 標題開頭、Meta Description、H1 標題、第一段(前 100 字)、至少 2 個 H2、至少 1 張圖片 alt、URL slug 建議
- 關鍵字密度:全文出現 3-6 次,密度約 0.5%-1.5%。

**4.2 SEO 標題與 Meta Description**
- SEO 標題:50-60 字元,焦點關鍵字 + 修飾詞,用 `|` 分隔品牌名。
  範例:`日月潭慢旅｜退休教師的湖畔散策與私房茶席 | 林太的慢活旅行誌`
- Meta Description:120-155 字元,含焦點關鍵字,以行動呼籲結尾。

**4.3 內容結構與可讀性**
- H1 僅一個、H2 至少 3 個、可搭配 H3。
- 每段不超過 150 字。
- 至少 30% 句子使用過渡詞(此外、值得一提的是、不過、換句話說…)。
- **內部連結**:讀取 `personas/<persona-slug>/published.json`(若存在),挑 1-2 篇主題相關的已發布文章,以 WordPress URL 插入連結。**只連結同一人格自己的舊文章**,不跨人格。
- **外部連結**:至少 1 個指向權威來源(景點官網、維基百科)。

**4.4 視覺資產強化**
- 每張 `<img>` 必須有描述性 `alt`,至少 1 張包含焦點關鍵字。
- 圖片下方加 `<figcaption>`,以該人格的口吻撰寫感性註解。

**4.5 結構化資料 (JSON-LD)**
- HTML 末尾加 `application/ld+json`,包含:
  - `BlogPosting`:作者(persona.display_name)、發布日期、主圖片、description
  - `FAQPage`:對應 FAQ 區塊的問答
- **FAQ 區塊**:文章末尾新增以該人格口吻命名的 FAQ 區(如林太用「林太的小叮嚀」),解答 3 個實務問題。

### 5. HTML 存檔 (HTML Storage Phase)

- **工具**:`write_file`。
- **路徑**:`.gemini/skills/persona-writer/personas/<persona-slug>/articles/yyyymmdd-location-topic.html`
- **內容**:完整 HTML 文件(含 `<!DOCTYPE html>`、`<head>`、`<body>`、JSON-LD)。
- **建立資料夾**:如果 `articles/` 不存在,write_file 會自動建立。
- **設計理由**:把產出的草稿 HTML 跟人格其他資料(persona.md / wp-config.json / published.json)放一起,結構乾淨,且 .gitignore 只要一條 `personas/*/articles/` 就涵蓋所有現在與未來的人格。

### 6. 自動發布 (Publishing Phase)

執行 shell 指令(注意 HTML 路徑已改到人格資料夾下):
```
python3 .gemini/skills/persona-writer/scripts/wp_poster.py <persona-slug> "<標題>" ".gemini/skills/persona-writer/personas/<persona-slug>/articles/<filename>.html" draft
```

- 第一個參數**必須**是 persona-slug,腳本會用它定位該人格的 `published.json`。
- 預設狀態 `draft`(草稿),**未經使用者明確同意,絕對不要用 `publish` 直接公開**。
- 發布成功後,腳本會自動更新該人格的 `published.json`。

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

每次任務完成前確認:

- [ ] 我有正確讀取 `personas/<slug>/persona.md` 並遵循其中的設定
- [ ] 我有確認 `personas/<slug>/wp-config.json` 存在(該人格的 WordPress 設定)
- [ ] 寫作口吻、主題選材、SEO 風格都符合該人格
- [ ] 內部連結只連到同一人格自己的舊文章(不跨人格)
- [ ] HTML 結構符合骨架規範
- [ ] 發布指令第一個參數是 persona-slug,且預設用 `draft`
- [ ] 完成後 published.json 有更新到正確人格的資料夾下
- [ ] 該人格的文章發到該人格自己的 blog(從 wp-config.json 的 WP_URL 看)
