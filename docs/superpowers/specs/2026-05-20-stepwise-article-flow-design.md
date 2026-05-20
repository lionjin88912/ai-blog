# Stepwise Article Flow — Design Spec

> 作者:行銷端業務回饋驅動,Gemini 協作整理
> 日期:2026-05-20
> 狀態:Design ready for implementation plan

---

## 1. 目標 (Goal)

把現行 `marketing-content-factory` 模組 2(寫一篇新文章)的「一鍵跑完 2-3 分鐘」流程,改成 **逐步引導 + 每階段人工 check** 的流程,讓行銷同仁能在文章結構成形的過程中即時介入修改,降低拿到草稿才發現大方向錯掉的代價。

## 2. 背景 (Why)

**現況**:模組 2 收完主題與人格後委派給 `persona-writer`,後者連跑「資料檢索 → 視覺獲取 → 寫作 → SEO → HTML → 發布」6 步,使用者全程沒有介入點,只能等最後看草稿連結。

**痛點**:
- 業務端反映想要 n8n 風格的「逐節點生成 + 人工審核」體驗
- 行銷使用者對 H1 / H2 結構有強意見,但目前沒有機會在生成前介入
- 一口氣跑 2-3 分鐘沒進度回饋,使用者不知道在做什麼

**業務端期望流程**(n8n 風格):
1. 輸入主題 + 關鍵字群
2. 生成 H1 標題 → 人工 check
3. 生成 H2 架構 → 人工 check
4. 生成 H3 小標 → 人工 check
5. 生成 FAQ → 人工 check
6. 生成全文 → 人工 check

註:實作時 **H2 與 H3 順序對調**(先 H3 發散小標、再 H2 歸納大綱)是業務端後續確認的正式版本。

## 3. 設計總覽 (Overview)

```
[使用者打開話]
    │
    ▼
Step 1 收主題 + 關鍵字 + 確認人格    ← 唯一強制使用者輸入的階段
    │
    ▼
Step 2 資料檢索 (silent,無 check)
    │
    ▼
Step 3 H1 標題    ── check ✋
    │
    ▼
Step 4 H3 小標    ── check ✋   (先發散)
    │
    ▼
Step 5 H2 大綱    ── check ✋   (再歸納)
    │
    ▼
Step 6 FAQ 3 題   ── check ✋
    │
    ▼
Step 7 全文       ── check ✋   (純文字,不含圖 / HTML)
    │
    ▼
Step 8 自動結尾段 (全程 silent,約 30 秒)
    ├── 8a 找圖
    ├── 8b 組完整 HTML
    ├── 8c SEO 結構自我檢查 (LLM self-review)
    ├── 8d 存 HTML + wp_poster.py 發 draft
    └── 8e 清掉 draft 暫存檔
    │
    ▼
最終回報:標題 + 草稿連結 + 圖數 + 字數 + 焦點關鍵字
```

## 4. 狀態管理 (State Management)

### 4.1 暫存檔位置與命名

```
.gemini/skills/persona-writer/personas/<persona-slug>/articles/
├── draft-<YYYYMMDD>-<HHMM>-<topic-slug>.json   ← 中間態 (寫作中)
└── <YYYYMMDD>-<location>-<topic>.html          ← 最終 HTML (完成品)
```

- `HHMM` 防撞:同日同主題多次寫不會覆蓋
- 與最終 HTML 同資料夾,方便管理
- `.gitignore` 已涵蓋 `personas/*/articles/`,不用改

### 4.2 JSON 結構

```json
{
  "persona_slug": "mrs-lin-slow-travel",
  "topic": "日月潭兩天一夜慢旅",
  "keywords": ["日月潭慢旅", "湖畔散策", "私房茶席"],
  "stage": "h3_done",
  "created_at": "2026-05-20T14:30:00",

  "research": [
    "日月潭原為邵族聖地…",
    "向山遊客中心由建築師…",
    "…"
  ],
  "h1": "日月潭兩天一夜慢旅｜林太的湖畔散策與私房茶席",
  "h3_subheadings": [
    "從窗邊茶香說起 — 為什麼這趟想去日月潭",
    "向山遊客中心的清晨光線",
    "湖畔散步道:伊達邵到水社的私房節奏",
    "茶廠裡的時光感:阿薩姆紅茶的故事",
    "給熟齡旅人的住宿建議",
    "兩天行程的呼吸節奏"
  ],
  "h2_outline": null,
  "faq": null,
  "full_text": null
}
```

### 4.3 stage 欄位的值

| stage 值 | 已完成什麼 | 下一步該做 |
|---|---|---|
| `init` | 主題、關鍵字、人格收齊 | 跑 research |
| `research_done` | research 拿到 | 生 H1 |
| `h1_done` | H1 確認 | 生 H3 小標 |
| `h3_done` | H3 確認 | 生 H2 大綱(歸納) |
| `h2_done` | H2 確認 | 生 FAQ |
| `faq_done` | FAQ 確認 | 生全文 |
| `full_text_done` | 全文確認 | 跑 Step 8 自動結尾段 |
| `published` | 發布完成 | 刪 draft 檔 |

### 4.4 生命週期

1. **建立**:Step 1 收完主題+關鍵字+人格 → write_file 建檔,`stage: "init"`
2. **累加**:每階段使用者點頭 → read_file → 更新該階段欄位 + 推進 stage → write_file
3. **接力**:使用者下次回來,Gemini 掃 `articles/draft-*.json`,問「上次寫到 X 階段,要繼續嗎?」
4. **完成**:Step 8e 發布成功後 `os.remove(draft-*.json)`
5. **失敗**:發布失敗時 **不刪 draft**,留著讓使用者下次接續或除錯

### 4.5 多份 draft 並存時

Gemini 偵測到多份 draft 提示:

```
你之前還有 2 篇寫到一半的草稿:
  - 「日月潭慢旅」(停在 H3 階段,5/20 開始)
  - 「九份散策」(停在全文階段,5/18 開始)

要繼續哪一篇?還是放棄它們、開新的?
```

「放棄」就刪掉對應 draft.json。

## 5. 對話互動格式 (Dialog Pattern)

### 5.1 通用三段式

每個 check 階段(Step 3-7)都用這個格式:

```
[1] AI 簡短預告        例:「H1 標題我提一個版本給你看 ──」
[2] AI 給「一個」建議    粗體 / 明確標出來,不附理由
[3] AI 問意見          固定句:「這樣可以嗎?還是有什麼想調整的?」
```

> **設計取捨**:不給多版本選單(會讓對話變長)。一次一個版本,使用者沒意見就快速過。要看更多選項的人主動說「換一個」,AI 才重生。

### 5.2 意圖判斷靠 LLM,不寫死關鍵字

Gemini 對使用者回應做 **語意理解**,分三類意圖。SOP 只給判準與範例,不列觸發詞清單。

#### A. 接受(前進)
- **語意特徵**:表達同意 / 沒意見 / 想繼續;短回應、不含對內容的修改建議
- **範例**:「好」「ok」「可以」「就這樣」「沒問題」「下一步」「嗯」「都行」「沒意見」「繼續」「Y」按 Enter 空訊息……
- **動作**:read_file → 寫入該階段欄位 + 推進 stage → write_file → 進下一階段

#### B. 直接給修改版(覆寫接受)
- **語意特徵**:使用者貼的內容看起來是要 **替換** AI 剛才的建議,不是下指令
- **範例**:H1 階段貼一段標題格式短句;H3/H2 階段貼條列清單;FAQ 階段貼 Q/A 結構;全文階段貼長文
- **動作**:AI 先二次確認「收到,用你這版繼續嗎?」→ 確認後走 A 的存檔動作

#### C. 下指令重生 / 修改
- **語意特徵**:對 AI 建議「指示要怎麼改」,而非「直接給新版」;含修改方向、風格指示、增刪要求、語氣調整
- **範例**:「再活潑一點」「換個角度」「加上某某主題」「拿掉第 3 點」「再短一點」「能不能更口語」……
- **動作**:
  1. AI 簡短回應:「好,我朝 OOO 方向再來一版 ──」
  2. 重新生成,**明確帶上使用者的修改方向到 prompt**
  3. 再次走「預告 → 給建議 → 問意見」三段式
  4. 同階段重生上限 5 次(見 5.4)

### 5.3 模糊回應一律問回去

| 使用者回應 | 模糊在哪 | 處理 |
|---|---|---|
| 「日月潭茶香」 | 是給新的 H1?還是補充關鍵字想加進去? | 問:「你是希望把 H1 直接改成『日月潭茶香』嗎?還是想要我加入這個元素改寫?」 |
| 「嗯這個第 3 點不太對」 | 是要改第 3 點?還是要整篇重生? | 問:「我重生第 3 點就好,還是整個 H3 大綱換個方向?」 |
| (使用者貼一段話,不確定是覆寫還是補充想法) | B 還是 C 的差別 | 問:「你給我這段是要直接用,還是希望我參考這個方向重寫?」 |

> **核心原則**:不確定就問,不要動。對行銷使用者來說,被 AI 誤解然後重做一輪比多回一句話更挫折。

### 5.4 重生上限(軟限制)

同階段累積 5 次重生後,AI 主動緩衝:

```
這已經是我試的第 5 個版本了 😅
我覺得目前最接近你想法的版本是這個 ──

<最新一版>

要不要先用這個往下走?之後到 WordPress 後台還是可以再潤稿。
或者你想直接貼一個你心目中的版本給我用?
```

不強制擋,使用者堅持還是可以繼續重生。

### 5.5 各階段的具體呈現格式

#### Step 3 — H1 標題
```
Gemini:H1 標題我提一個版本 ──

       「日月潭兩天一夜慢旅｜林太的湖畔散策與私房茶席」

       這樣可以嗎?還是有什麼想調整的?
```

#### Step 4 — H3 小標(先發散)
```
Gemini:H3 我先把可以講的點都列出來,等你點頭再去歸納大綱 ──

       1. 從窗邊茶香說起 — 為什麼想去日月潭
       2. 向山遊客中心的清晨光線
       3. 湖畔散策:伊達邵到水社的私房節奏
       4. 茶廠裡的時光感:阿薩姆紅茶的故事
       5. 給熟齡旅人的住宿建議
       6. 兩天行程的呼吸節奏

       這 6 個點 OK 嗎?要加要減要換?
```

#### Step 5 — H2 大綱(歸納上面 H3)
```
Gemini:把剛才的 H3 歸納成 3 個章節 ──

       ## 一、出發前的心情(對應 H3 第 1 點)
       ## 二、湖畔散策的私房節奏(對應 H3 第 2、3、4 點)
       ## 三、慢旅尾聲的安排(對應 H3 第 5、6 點)

       這樣分組可以嗎?
```

#### Step 6 — FAQ
```
Gemini:文末的「林太的小叮嚀」我準備這 3 題 ──

       Q1:第一次去日月潭,選日月潭住宿還是埔里住宿好?
       A1:(120 字回答…)

       Q2:慢旅節奏的話兩天一夜會不會太趕?
       A2:(120 字回答…)

       Q3:除了 SUP 跟自行車環湖,中高齡有什麼節奏舒適的玩法?
       A3:(120 字回答…)

       這 3 題符合你想答的問題嗎?
```

#### Step 7 — 全文
```
Gemini:全文寫好了(約 1800 字),貼給你看 ──

       <貼出完整正文,純文字段落,不含 HTML / 圖片>

       這篇方向 OK 嗎?還是哪一段要重寫?
```

> **設計取捨**:全文階段不附 HTML、不附圖、不附 JSON-LD。**只看正文純文字內容**。「OK」之後 Step 8 才把 HTML / 圖 / SEO 全部組好。原因:全文時還沒找圖,顯示 HTML 會有「[圖]」placeholder 反而干擾使用者判斷文字品質。

## 6. Step 8 自動結尾段細節 (Auto-Bundle Phase)

全程 silent,**中間不報「找圖中…」「排版中…」逐項進度**,避免訊息洪水。只有失敗才講話。

### 6.1 8a 找圖

- 從 draft.json 的 `topic` + `research` 萃取 **英文「城市+地點」** 關鍵字(例:`Sun Moon Lake,Ita Thao,Xiangshan Visitor Center`)
- `write_file` 建 `request.json`
- `curl` 打 Google Apps Script 圖片 API
- 從 API 回傳挑 3-4 張,**選圖原則依 persona.md 的「視覺風格偏好」**

**失敗處理**:API 沒回應 / 圖數不足 → 不擋發布,改用「無圖純文字版本」繼續往下,回報訊息結尾加一行警示。

### 6.2 8b 組完整 HTML

從 draft.json 取出 `h1`、`h2_outline`、`h3_subheadings`、`faq`、`full_text` 組裝成完整 HTML,結構遵循 `persona-writer/SKILL.md` 既有 HTML 骨架(含 `<head>` meta、JSON-LD 結構化資料)。

> 注意:JSON-LD 放在 HTML 裡,但 `wp_poster.py` 發文前會剝掉 `<script>` 標籤(避免 wp.com 把 JSON 內容洩漏到內文)。**現階段 JSON-LD 只存在本地 HTML 檔,沒上 WordPress**。跟現有行為一致。

### 6.3 8c SEO 結構自我檢查 (LLM self-review)

組好 HTML 後,**AI 自己檢查一遍**(不是用程式 lint,是 LLM 自己 review):

- 焦點關鍵字密度落在 0.5–1.5%
- 焦點關鍵字有出現在:title、description、第一段、≥2 個 H2、≥1 張 alt
- H1 只有一個
- H2 ≥ 3 個
- 每段 ≤ 150 字
- 每張 `<img>` 都有描述性 alt、≥1 張 alt 含焦點關鍵字
- 過渡詞使用率 ≥ 30%

**不通過**:AI 自己修。連修 3 次還過不了 → 直接發出去,回報結尾警示「SEO 結構檢查有幾項沒達標」。

### 6.4 8d 存 HTML + 跑 wp_poster.py

```bash
# Gemini 內部執行
write_file: personas/<slug>/articles/<YYYYMMDD>-<location>-<topic>.html
shell: python3 .gemini/skills/persona-writer/scripts/wp_poster.py \
       <persona-slug> "<H1 標題>" "<HTML 檔路徑>" draft
```

**強制 `draft`**,這條規則不變(除非使用者在 Step 1 就明說「直接公開」)。

`wp_poster.py` 內部邏輯(沒變動):
- 依 wp-config.json 的 `auth_method` 走 Application Password 或 OAuth2
- 剝掉 `<script>` 標籤再發
- 成功時自動 append `published.json`
- 失敗時印 `❌ 失敗。狀態碼: ...`

**發布失敗**:
- AI 不刪 draft.json
- 回報訊息對照模組 4 FAQ 給對應解法
- draft 留著,使用者排錯後可說「再試一次發布」直接從 draft 接續

### 6.5 8e 清掉 draft 暫存檔

**只有發布成功** 才刪:

```python
os.remove(.gemini/skills/persona-writer/personas/<slug>/articles/draft-*.json)
```

最終 HTML 保留,當本地完成品備份。

### 6.6 最終回報訊息

```
✅ 寫好囉!
📄 標題:<H1>
🔗 草稿連結:<WordPress 後台編輯連結>
🖼️ 用了 X 張圖、寫了約 X 字
🎯 焦點關鍵字:<焦點關鍵字>

點上面連結進去後,你可以再潤稿,確認沒問題就按 WordPress 右上角的「發布」。
```

(若 8a 圖失敗 / 8c SEO 沒達標,結尾加警示行)

## 7. 錯誤與中斷處理 (Error & Interruption)

| 情境 | 處理 |
|---|---|
| 使用者 session 中斷後回來 | Gemini 掃 `articles/draft-*.json` → 問「上次寫到 X 階段,要繼續嗎?」 |
| 同人格多份 draft 並存 | 列出來讓使用者挑要繼續哪份 / 放棄哪份 |
| 模糊意圖回應(B 還 C?) | 問回去,不要瞎猜 |
| 同階段重生 ≥5 次 | AI 主動緩衝建議「先用最新版往下」,不強制擋 |
| 跨多階段回退(H2 階段要改 H1) | AI 引導「乾脆從頭來?還是只改最近這階段?」 |
| 找圖 API 失敗 | 不擋發布,純文字版繼續 + 結尾警示 |
| SEO 自評連修 3 次不過 | 直接發出去 + 結尾警示「有幾項沒達標」 |
| WordPress 發布失敗 | draft 不刪,回報錯誤對照模組 4 FAQ;使用者可說「再試一次發布」接續 |
| 使用者中途說「不寫了 / 放棄」 | 二次確認「確定放棄這篇『XXX』嗎?草稿會被清掉」→ 是 → 刪 draft.json |
| 使用者中途要換人格 | 二次確認「之前寫的 H1/H3/... 會作廢,真的要換 OOO 嗎?」→ 是 → 刪舊 draft、開新對話 |
| Step 1 之前:該人格沒設定 WordPress | 中止寫文章流程,引導到模組 1 設定 |
| Step 1 之前:人格不存在 | 列出可用人格、或引導模組 5 新增 |

## 8. 要改的檔案清單 (Files Touched)

```
.gemini/skills/marketing-content-factory/SKILL.md
  ├── 模組 2 (寫一篇新文章):整個重寫
  │   - 拿掉現有 Step A/B/C/D/E (一鍵跑完那段)
  │   - 換成新 Step 1 → 8 (逐步引導 + 自動結尾段)
  │   - 加入「偵測既有 draft」分支
  │
  └── 沿用不動:模組 1、3、4、5、6,以及全域行為準則

.gemini/skills/persona-writer/SKILL.md
  ├── 改:標準作業程序 (SOP) 段落
  │   - 原本「7 步 SOP (Step 0-6) 一氣呵成」改成「draft.json 接力 + 階段呼叫」
  │   - 環境檢查 (原 Step 0) 折入新 Step 1 的人格確認
  │   - 明訂每階段讀什麼 / 寫什麼欄位 / 推進到哪個 stage
  │
  └── 沿用不動:HTML 骨架、人格載入流程、給 Gemini 的自我檢查清單

不動的東西 (零腳本變動)
  ├── scripts/wp_poster.py        ← CLI 簽名不變
  ├── scripts/detect_site.py
  ├── scripts/wp_oauth_setup.py
  ├── personas/_template/         ← 範本不變
  └── .gitignore                  ← 已涵蓋 articles/*
```

**估計改動量**:
- `marketing-content-factory/SKILL.md`:模組 2 約 70 行 → 改成約 180 行
- `persona-writer/SKILL.md`:SOP 段落約 100 行 → 改成約 150 行
- 共約 200 行新文字、零程式碼

## 9. 風險與相容性 (Risks & Compatibility)

- **純改 SOP**(給 Gemini 看的指引文件),不動 Python 腳本,回滾容易
- 跟現有 **per-persona**、**OAuth2**、**SEO meta silent no-op** 三個已落地機制都相容
- 對既有 `mrs-lin-slow-travel` persona 沒 breaking change(draft.json 是新增的,舊資料不受影響)
- 新流程比舊流程 **使用者互動次數變多**(從 2-3 次變成 5-7 次),需要在後續使用者測試中確認接受度

## 10. 不在這次範圍 (Out of Scope)

下列項目不在此 spec 涵蓋,留給未來工作:

- SEO meta 自動帶到 WordPress 後台(等 companion mu-plugin / Code Snippets bootstrap)
- 自動發布(`publish` 而非 `draft`)
- 多人格協作(一篇文章用兩位人格)
- 跨階段任意回退(目前只支援「回退一階段」)
- 草稿版本控制 / 歷史紀錄(draft 是一次性的,沒做歷史保留)
- 自動 commit draft.json(draft 是 .gitignored,不進版本控)
