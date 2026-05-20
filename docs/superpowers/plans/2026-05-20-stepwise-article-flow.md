# Stepwise Article Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `marketing-content-factory` 模組 2 與 `persona-writer` SOP，將「一鍵跑完 2-3 分鐘」改成 n8n 風格的逐節點生成 + 人工 check（H1 → H3 → H2 → FAQ → 全文），中間態存 `draft-*.json` 暫存檔，全文 check 後 silent 跑找圖+組 HTML+發 draft。

**Architecture:** 純 SOP 文件變動，**零 Python 腳本變動**。中間態用一份 per-article 的 JSON 暫存檔追蹤 `stage` 欄位（`init` → `research_done` → `h1_done` → `h3_done` → `h2_done` → `faq_done` → `full_text_done` → `published`），Gemini 用 `read_file`/`write_file` 管理。意圖判斷靠 LLM 語意理解，不寫死關鍵字。

**Tech Stack:** 純 Markdown SOP 文件 (`.gemini/skills/*/SKILL.md`)。中間態 JSON。沿用既有 `wp_poster.py` / `detect_site.py` / `wp_oauth_setup.py` 不改。

**參考文件:** `docs/superpowers/specs/2026-05-20-stepwise-article-flow-design.md` (完整設計 spec，本計畫的每一節都對應到 spec 的某段)

---

## SOP 工作的「測試」概念

本專案沒有測試框架，且此次只改 Gemini 的指引文件（不執行的 prose），所以採取以下取代 TDD：

- **Test = Trace**：每個 Task 完成後，閱讀新 SOP 把代表性對話走一遍，確認流程能銜接、無矛盾、無 placeholder
- **Verify = Read & Trace by hand**：手動 trace 是唯一可靠的驗證方法。沒有 `pytest`、沒有 lint
- **Final test (Task 7) = End-to-end walkthrough**：把所有情境（正常、中斷、放棄、失敗、換人格）口頭走一遍

---

## File Structure

```
docs/superpowers/specs/2026-05-20-stepwise-article-flow-design.md   ← 已存在 (Task 0 前置)
docs/superpowers/plans/2026-05-20-stepwise-article-flow.md          ← 本檔

要修改的檔案 (兩個 SKILL.md)
.gemini/skills/marketing-content-factory/SKILL.md
  └── 模組 2 (line 258-320):整段替換,從約 63 行 → 約 220 行
.gemini/skills/persona-writer/SKILL.md
  ├── SOP 段落 (line 32-134):整段替換,從約 103 行 → 約 150 行
  └── 自我檢查清單 (line 194-206):補 2 條 draft.json 相關項目

不動的檔案
  scripts/wp_poster.py           CLI 簽名與行為都不變
  scripts/detect_site.py
  scripts/wp_oauth_setup.py
  personas/_template/*           範本不變
  .gitignore                     已涵蓋 personas/*/articles/
  GEMINI.md                      路由規則表中「寫文章」routing 仍指向模組 2,不需要改
```

每個 Task 修改的目標檔案與行範圍會在該 Task 開頭明示。

---

## Task 1: 重寫模組 2 - 開頭 + Step 1 (收主題+關鍵字+人格 / 偵測既有 draft)

**目標**:把 `marketing-content-factory/SKILL.md` line 258-320 的整段模組 2 用 Edit 一次替換掉,放入新版的「開頭說明 + Step 1」。後續 Task 2-5 會再用 Edit 把 Step 2-8 一段一段插進去。

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (line 258-320 整段替換)

- [ ] **Step 1: Read 模組 2 確認邊界**

Run: `Read` tool 讀 `marketing-content-factory/SKILL.md` line 256-322,確認:
- line 258 是 `## 📝 模組 2:寫一篇新文章的完整流程`
- line 320 末是 Module 2 的最後一行 (空行或 `---` 之前)
- line 321 是 `## ⌨️ 模組 3:常用指令範例(複製貼上)`

Expected: 邊界與計畫一致。如果不一致,以實際讀到的為準調整 Edit 範圍。

- [ ] **Step 2: 用 Edit 把模組 2 整段替換成新版開頭 + Step 1**

old_string (從 `## 📝 模組 2` 開始到 `---` 結束,完整現存 Module 2):需要讀檔案後完整貼入。下面是替換進去的 new_string:

```markdown
## 📝 模組 2:寫一篇新文章的完整流程

> 目標:從一句話的主題,到 WordPress 草稿,全程透過「逐步引導 + 人工 check」的對話節奏完成。
> 行銷使用者只在 7 個明確節點介入(Step 1 收輸入 + Step 3-7 共 5 個 check),其餘 silent。
> **中間態存在每篇文章自己的 draft.json**(在該人格的 `articles/` 資料夾裡),session 中斷可以接續、發布成功後刪掉。

### 流程總覽

```
Step 1 收主題 + 關鍵字 + 人格   ← 使用者必填
Step 2 資料檢索 (silent)
Step 3 H1 標題     ── check ✋
Step 4 H3 小標     ── check ✋   (先發散)
Step 5 H2 大綱     ── check ✋   (再歸納)
Step 6 FAQ 3 題    ── check ✋
Step 7 全文        ── check ✋   (純文字)
Step 8 自動結尾段 (silent ~30 秒)
       ├── 找圖
       ├── 組 HTML
       ├── SEO 自評
       ├── wp_poster.py 發 draft
       └── 刪 draft.json
```

### Step 1 — 收輸入(包含「偵測既有 draft」分支)

**先掃既有 draft**:在問使用者任何問題前,掃 `.gemini/skills/persona-writer/personas/*/articles/draft-*.json`。

- **掃到 0 份** → 走「新文章」路徑(下方 Step 1A)
- **掃到 ≥1 份** → 走「偵測到既有 draft」路徑(下方 Step 1B)

#### Step 1A — 新文章

**A-1. 收主題與關鍵字群**

問:
> 好的,我們來寫文章 ✍️ 先請你給我兩件事(請分兩行貼,或一句話講完都行):
>
> 1. 文章主題(例:日月潭兩天一夜慢旅、米其林餐廳推薦…)
> 2. 你希望這篇主打的「關鍵字群」(用逗號分隔,例:日月潭慢旅, 湖畔散策, 私房茶席)

驗收條件:
- 主題:必須有,空字串或「隨便」「都行」要追問
- 關鍵字群:必須有 **至少 1 個**;使用者完全沒給時追問「你想讓 Google 搜這篇時用什麼字?隨便給 1-3 個就好」

**A-2. 確認人格**

掃 `.gemini/skills/persona-writer/personas/` 排除 `_template`:
- **只有一位人格**:直接帶他用該位(例:目前只有林太就用林太),口頭跟使用者預告「我會用『林太』寫,沒問題吧?」
- **多位人格 + 使用者話裡有指定**(例「用王老闆寫」):照他說的
- **多位人格 + 使用者沒指定**:列出所有人格的 `display_name` + `topic`,讓他挑

**A-3. 確認該人格的 WordPress 已設定**

檢查 `.gemini/skills/persona-writer/personas/<slug>/wp-config.json` 是否存在。
- 不存在 → 中止 Step 1,引導去模組 1:「『<人格中文名>』還沒設定過 WordPress 部落格,我們先花 2 分鐘設定好」
- 存在 → 繼續 A-4

**A-4. 建立 draft.json**

跟使用者預告:
> 收到主題「<主題>」+ 關鍵字「<關鍵字 1, 2, 3>」+ 用「<人格中文名>」寫。
> 我先幫你建一份草稿筆記、上網查資料(大約 15 秒),然後就會把標題提給你看。

執行 `write_file`:
- **路徑**:`.gemini/skills/persona-writer/personas/<persona-slug>/articles/draft-<YYYYMMDD>-<HHMM>-<topic-slug>.json`
  - `YYYYMMDD-HHMM` 用當下時間,**HHMM 防同主題同日撞檔**
  - `topic-slug`:把主題簡化成 kebab-case 英數(中文主題用拼音 / 簡化英譯),例:`sun-moon-lake`、`tainan-old-city`、`ho-chi-minh-coffee`
- **內容**:
```json
{
  "persona_slug": "<persona-slug>",
  "topic": "<使用者給的主題原文>",
  "keywords": ["<關鍵字 1>", "<關鍵字 2>", "..."],
  "stage": "init",
  "created_at": "<YYYY-MM-DDTHH:MM:SS>",
  "research": null,
  "h1": null,
  "h3_subheadings": null,
  "h2_outline": null,
  "faq": null,
  "full_text": null
}
```

接著進入 Step 2(資料檢索)。

#### Step 1B — 偵測到既有 draft

掃到既有 draft 後,把每份 draft 的 `topic`、`stage`、`created_at`、`persona_slug` 讀出來,對使用者說:

> 你之前還有 N 篇寫到一半的草稿:
>   - 「<topic 1>」(<persona 中文名>,停在 <stage 中文> 階段,<created_at 日期>)
>   - 「<topic 2>」(<persona 中文名>,停在 <stage 中文> 階段,<created_at 日期>)
>   - …
>
> 要繼續哪一篇?還是放棄、開新的?

**stage 中文對照表**(對使用者顯示用):
- `init` → 「剛開始」
- `research_done` → 「資料查完」
- `h1_done` → 「H1 標題已確認」
- `h3_done` → 「H3 小標已確認」
- `h2_done` → 「H2 大綱已確認」
- `faq_done` → 「FAQ 已確認」
- `full_text_done` → 「全文已確認(待發布)」

使用者選擇:
- **「繼續第 X 篇」** → 讀那份 draft,依 `stage` 跳到對應步驟(`init` → Step 2、`research_done` → Step 3、`h1_done` → Step 4、`h3_done` → Step 5、`h2_done` → Step 6、`faq_done` → Step 7、`full_text_done` → Step 8)
- **「放棄全部、開新的」** → 二次確認「我幫你把那 N 份草稿都刪掉,確定嗎?」→ 是 → `os.remove` 刪掉那些 draft → 走 Step 1A
- **「放棄第 X 篇,但繼續第 Y 篇」** → 二次確認後刪 X,走「繼續 Y」流程
- **混合需求**:逐項問清楚,不要批次猜
```

new_string 的完整內容如上 (Step 1A + Step 1B 完整列出)。

> **重要**:`new_string` 結尾**沒有** `---` 分隔符,因為 Task 2 會接著在這後面 Insert Step 2。Edit 後 Module 2 暫時是不完整的,Task 2 補完後才會接 Module 3。

- [ ] **Step 3: 用 trace 驗證 Task 1 的結果**

打開新版的 `marketing-content-factory/SKILL.md` Module 2,口頭走以下情境:

**情境 A (新文章)**:使用者打「幫我用林太寫日月潭」
- Step 1 掃 draft 0 份 → 走 1A
- 1A-1 收到主題「日月潭」+ 沒給關鍵字 → 該追問關鍵字
- 1A-2 多人格但話裡指定林太 → 用林太
- 1A-3 林太 wp-config 存在 → 繼續
- 1A-4 write_file 建 `draft-20260520-1430-sun-moon-lake.json`,stage=init

預期:每一步 SOP 都有明確指示;沒有 placeholder。

**情境 B (有既有 draft)**:同樣訊息,但 articles/ 已有 1 份 draft 停在 h2_done
- Step 1 掃到 1 份 → 走 1B
- 列出該 draft,問「繼續還是放棄」
- 使用者說「繼續」 → 跳到 Step 6 (faq 階段)

預期:1B 的 stage→step 對應表能找到正確的目標 step。

**情境 C (使用者沒指定人格,多人格)**:訊息「我要寫一篇文章」
- 1A-2 應該列出可用人格選單

預期:SOP 有明確的「多人格沒指定」分支。

如果 trace 中發現任何缺漏,直接改 Edit。

- [ ] **Step 4: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "feat(factory): module 2 rewrite — step 1 (collect input + draft resume detection)"
```

---

## Task 2: 加入 Step 2 (silent research) + 共用對話樣板

**目標**:在 Task 1 完成的 Step 1 後面 Insert 兩個新區塊 — 「Step 2 資料檢索」與「共用對話樣板 + 意圖分類」。後者會被 Step 3-7 共同引用。

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (在 Task 1 結尾後追加)

- [ ] **Step 1: Edit 在 Task 1 結尾後 append Step 2 + 共用樣板**

old_string:Task 1 最後一段 Step 1B 結尾的 `- **混合需求**:逐項問清楚,不要批次猜`(這是 Step 1B 在檔案中的最後一句,作為 Edit 的錨點)

new_string(把 old_string 整段保留 + 在後面接這段):

```markdown
- **混合需求**:逐項問清楚,不要批次猜

---

### Step 2 — 資料檢索(silent)

跟使用者預告(只說一次,不報進度):
> 我先查一下「<主題>」的背景資料,15 秒左右回你。

執行:
1. 用 `google_web_search` 依該人格 `persona.md` 的 `topic` 與本次主題 + 關鍵字搜資料
2. 至少蒐集 3 個有深度的素材(文化背景、交通、特色、地方典故)
3. 素材的選材必須符合 persona.md 的「適合主題範例」與「寫作禁區」
4. read_file draft → 把素材寫入 `research`(string list)+ 推進 `stage` 為 `research_done` → write_file 回去

不對使用者顯示 research 細節,**直接接 Step 3 提 H1**。

> 失敗處理:`google_web_search` 拿不到結果(網路問題 / 主題太冷門)時,告知使用者「我這邊查資料卡住,你要不要給我一兩段你自己手邊的資料,我用那個寫?」 → 收到 → 寫入 `research`(視為使用者提供)+ 推進 stage,接 Step 3。

---

### 共用對話樣板 — Step 3-7 通用三段式

**Step 3-7 的每一個 check 階段,都用這個對話格式產出建議:**

```
[1] AI 簡短預告
    例:「H1 標題我提一個版本給你看 ──」

[2] AI 給「一個」建議(粗體或明確標出來,不附理由)
    例:**「日月潭兩天一夜慢旅｜林太的湖畔散策與私房茶席」**

[3] AI 問意見(固定句)
    「這樣可以嗎?還是有什麼想調整的?」
```

> **不給多版本選單**。一次一個版本,使用者沒意見就快速過。要看更多選項的人主動說「換一個」,AI 才重生。

### 共用對話樣板 — 意圖分類(靠 LLM 語意,不寫死關鍵字)

使用者回應後,**用語意理解**分三類,SOP 不列觸發詞清單。

#### A. 接受(前進)

**語意特徵**:表達同意 / 沒意見 / 想繼續;短回應、不含修改建議。
**範例**:「好」「ok」「可以」「就這樣」「沒問題」「下一步」「嗯」「都行」「沒意見」「繼續」「Y」按 Enter 空訊息……
**動作**:read_file draft → 寫入該階段欄位 + 推進 `stage` → write_file → 進下一階段

#### B. 直接給修改版(覆寫接受)

**語意特徵**:使用者貼的內容看起來是要 **替換** AI 剛才的建議,而不是下指令。長度、格式、角色匹配該階段的產出(H1 是短句、H3/H2 是條列、FAQ 是 Q/A、全文是長文)。
**動作**:
1. AI 先二次確認:「收到,用你這版繼續嗎?」
2. 使用者再點頭一次 → 走 A 的存檔動作
3. 為什麼要二次確認:避免「使用者只是想補一個想法」被誤判為覆寫

#### C. 下指令重生 / 修改

**語意特徵**:對 AI 建議「指示要怎麼改」,而非「直接給新版」;含修改方向、風格指示、增刪要求、語氣調整。
**範例**:「再活潑一點」「換個角度」「加上某某主題」「拿掉第 3 點」「再短一點」「能不能更口語」「太正經了」「重生」「換一個」……
**動作**:
1. AI 簡短回應:「好,我朝 OOO 方向再來一版 ──」(明確覆述修改方向)
2. 重新生成,**把使用者的修改方向加到 generation prompt**
3. 再次走「預告 → 給建議 → 問意見」三段式
4. 同階段重生 **上限 5 次**(見下)

#### 模糊回應 — 一律問回去,不要瞎猜

| 使用者回應 | 模糊在哪 | 處理 |
|---|---|---|
| 「日月潭茶香」 | 是給新的 H1?還是補充關鍵字想加進去? | 問:「你是希望把 H1 直接改成『日月潭茶香』嗎?還是想要我加入這個元素改寫?」 |
| 「嗯這個第 3 點不太對」 | 是要改第 3 點?還是要整篇重生? | 問:「我重生第 3 點就好,還是整個 H3 大綱換個方向?」 |
| (使用者貼一段話,不確定是覆寫還是補充想法) | B 還是 C 的差別 | 問:「你給我這段是要直接用,還是希望我參考這個方向重寫?」 |

> **核心原則**:不確定就問,不要動。對行銷使用者來說,被 AI 誤解然後重做一輪比多回一句話更挫折。

#### 重生上限(軟限制)

同階段累積 5 次重生後,AI 主動緩衝:

```
這已經是我試的第 5 個版本了 😅
我覺得目前最接近你想法的版本是這個 ──

<最新一版>

要不要先用這個往下走?之後到 WordPress 後台還是可以再潤稿。
或者你想直接貼一個你心目中的版本給我用?
```

不強制擋,使用者堅持還是可以繼續重生。

#### 跨階段回退

使用者在 H2 階段說「H3 第 3 點不對」也可以:
- 接受「回 H3 修改」這種跨一階段的指令
- **只回退一階段**,不能跨多階段(H2 階段不接受「回 H1 改標題」)
- 跨多階段時引導:「H1 也想改的話,我們乾脆從頭來?還是只改 H3 就好?」
```

> Insert 結束。後續 Task 3 會在「跨階段回退」這段之後 append Step 3-7 的 5 個具體 check 例子。

- [ ] **Step 2: Trace 驗證 Task 2 的結果**

口頭走以下情境:
- Step 1A 結束 (draft.json stage=init) → Step 2 跑 research → draft 更新 stage=research_done。預期:銜接順,research 失敗有 fallback
- Step 3 收到使用者「好」 → 走 A → 推進 stage、進 Step 4。預期:「A 接受」邏輯能套用到 Step 3-7 任一階段
- Step 4 收到使用者「再加一個關於茶廠的點」 → 走 C → 重生 → 再問意見。預期:「C 重生」邏輯能套用
- Step 5 收到使用者貼一段完整 H2 大綱 → 走 B → 二次確認 → 接受。預期:「B 覆寫」邏輯能套用
- 同階段重生 5 次 → 看到緩衝訊息

如果共用樣板無法套用到某個 Step,直接補充樣板。

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "feat(factory): module 2 — step 2 research + shared dialog/intent templates"
```

---

## Task 3: 加入 Step 3-7 (五個 check 階段的具體 prose)

**目標**:在 Task 2 的「跨階段回退」之後 append 五個具體 check 階段。每階段沿用 Task 2 的「通用三段式 + 意圖分類」,只描述該階段的「該生成什麼」「該寫入 draft 哪個欄位」「該推進到哪個 stage」。

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (在 Task 2 結尾後 append)

- [ ] **Step 1: Edit append Step 3-7 五個 check 區塊**

old_string:Task 2 結尾的 `- 跨多階段時引導:「H1 也想改的話,我們乾脆從頭來?還是只改 H3 就好?」`

new_string(保留 old_string,後面接這段):

```markdown
- 跨多階段時引導:「H1 也想改的話,我們乾脆從頭來?還是只改 H3 就好?」

---

### Step 3 — H1 標題 ✋

從 draft.json 讀 `topic` / `keywords` / `research` + 該人格的 SEO 風格,生 **1 個** H1 標題(50-60 字元,**首位放焦點關鍵字**,用 `｜` 分隔 sub,用 ` | ` 連 persona display_name 結尾)。

範例:`日月潭兩天一夜慢旅｜林太的湖畔散策與私房茶席 | 林太的慢活旅行誌`

用通用三段式呈現給使用者。

意圖分類後:
- A 接受 → read_file → `h1 = "<該標題>"` + `stage = "h1_done"` → write_file → 進 Step 4
- B 覆寫 → 二次確認 → 用使用者那版 → A 動作
- C 重生 → 朝指定方向再生 → 重新三段式

---

### Step 4 — H3 小標 ✋(先發散)

從 draft.json 讀 `h1` / `topic` / `research`,**列出整篇能講的點**,5-8 個 H3 小標(每個 15-25 字,具體不抽象)。**這一步刻意先發散,Step 5 才歸納**。

用通用三段式呈現,例:

```
H3 我先把可以講的點都列出來,等你點頭再去歸納大綱 ──

1. 從窗邊茶香說起 — 為什麼想去日月潭
2. 向山遊客中心的清晨光線
3. 湖畔散策:伊達邵到水社的私房節奏
4. 茶廠裡的時光感:阿薩姆紅茶的故事
5. 給熟齡旅人的住宿建議
6. 兩天行程的呼吸節奏

這 6 個點 OK 嗎?要加要減要換?
```

意圖分類後:
- A → `h3_subheadings = [<list>]` + `stage = "h3_done"` → 進 Step 5
- B → 二次確認 → A 動作
- C → 朝指定方向重生(常見:「加一個關於 XXX」「拿掉第 N 點」「換個角度」)

---

### Step 5 — H2 大綱 ✋(歸納上面 H3)

讀 `h3_subheadings`,**歸納成 2-4 個 H2 章節**,**明示哪些 H3 歸到哪個 H2**(這樣使用者能驗證歸納邏輯)。

用通用三段式呈現,例:

```
把剛才的 H3 歸納成 3 個章節 ──

## 一、出發前的心情(對應 H3 第 1 點)
## 二、湖畔散策的私房節奏(對應 H3 第 2、3、4 點)
## 三、慢旅尾聲的安排(對應 H3 第 5、6 點)

這樣分組可以嗎?
```

意圖分類後:
- A → `h2_outline = [{"h2": "<標題>", "h3_indices": [<index>, ...]}, ...]` + `stage = "h2_done"` → 進 Step 6
- B → 二次確認 → A 動作
- C → 朝指定方向重生(常見:「分成 2 個就好」「第 3 章拆成兩個」)

---

### Step 6 — FAQ ✋

讀 `topic` / `h1` / `research`,生 **3 題實務 FAQ**,Q1 是「決定要不要去」、Q2 是「節奏與時間」、Q3 是「該人格族群特化問題」(熟齡、親子、預算等)。每 A 約 120 字,口吻完全符合該人格。

FAQ 區塊在文章末尾的命名:**用該人格的口吻命名**(林太用「林太的小叮嚀」,王老闆可能用「王老闆的呷飯心法」)。命名規則寫在 persona.md 裡,從那邊讀。

用通用三段式呈現。

意圖分類後:
- A → `faq = [{"q": "...", "a": "..."}, ...]` + `stage = "faq_done"` → 進 Step 7
- B → 二次確認 → A 動作
- C → 朝指定方向重生

---

### Step 7 — 全文 ✋

讀 draft 中所有已確認欄位(`h1` / `h3_subheadings` / `h2_outline` / `faq` / `research` / `persona.md`),寫出 **純文字段落版本** 的全文。

**注意**:
- **不要附 HTML 標籤**(沒有 `<h1>`、`<p>`、`<img>`)
- **不要插圖 placeholder**(沒有 `[圖]` 或 `<img>`)
- **不要附 JSON-LD**
- 只給乾淨的段落文字,讓使用者只看內容品質

格式:用空行分段,H1 用 `# 標題` markdown 標明,H2 用 `## 章節名` 標明,H3 用 `### 小標` 標明,FAQ 用 `### Q1: ... A1: ...`。

用通用三段式呈現,例:

```
全文寫好了(約 1850 字),貼給你看 ──

# 日月潭兩天一夜慢旅｜林太的湖畔散策與私房茶席

(開頭段落...)

## 一、出發前的心情

### 從窗邊茶香說起 — 為什麼想去日月潭

(段落...)

(...完整正文...)

### 林太的小叮嚀

### Q1: ...
(...)

這篇方向 OK 嗎?還是哪一段要重寫?
```

意圖分類後:
- A → `full_text = "<完整 markdown 全文>"` + `stage = "full_text_done"` → 進 Step 8 自動結尾段
- B → 二次確認 → A 動作
- C → 朝指定方向重生(常見:「第 2 段太冗」「結尾哲理改一下」「整體再溫暖一點」)
```

- [ ] **Step 2: Trace 驗證 Task 3 的結果**

對每個 Step 3-7,口頭走「使用者說好」「使用者貼新版」「使用者要求修改」三條路:

**Step 3 (H1)**:
- 收到「好」 → write h1 + stage=h1_done。 預期:有明確 schema
- 收到使用者貼一個新 H1 → 二次確認 → 寫入
- 收到「再正式一點」 → 重生 → 再三段式

**Step 4 (H3)**:
- 收到「拿掉第 5 點」 → 重生時只刪該點 + 維持其他

**Step 5 (H2)**:
- 收到「分 2 個就好」 → 重生時改成 2 個 H2 + 重新對應 h3_indices

**Step 6 (FAQ)**:
- 命名規則(「林太的小叮嚀」)從 persona.md 哪裡讀?

  → trace 過程中發現 persona.md `mrs-lin-slow-travel/persona.md` 在「文章架構偏好」段提到「林太的小叮嚀(FAQ)」。SOP 應該交代「從該段抽取命名」,如果 trace 發現 SOP 沒寫,補上(在 Step 6 加一句:「命名從 persona.md 的『文章架構偏好』段抽,若沒寫則 fallback 為『<display_name>的問與答』」)。

**Step 7 (全文)**:
- 收到「結尾改溫暖一點」 → 只改結尾段,不重寫整篇? 或全篇重寫?
  → 預期:重生 = 全篇重寫,因為 LLM 部分改動會破壞一致性。Step 7 該段補一句:「Step 7 的『C 重生』= 整篇重寫,不做局部 patch」

如果 trace 發現上述缺漏,直接 Edit 補進去。

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "feat(factory): module 2 — step 3-7 five check stages (h1/h3/h2/faq/full text)"
```

---

## Task 4: 加入 Step 8 自動結尾段

**目標**:在 Task 3 結尾後 append Step 8 區塊,描述全文 check 過後的自動段(8a-8e)+ 最終回報訊息。

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (在 Task 3 結尾後 append)

- [ ] **Step 1: Edit append Step 8 區塊**

old_string:Task 3 結尾 Step 7 區塊的最後一行 `- C → 朝指定方向重生(常見:「第 2 段太冗」「結尾哲理改一下」「整體再溫暖一點」)`

new_string:

```markdown
- C → 朝指定方向重生(常見:「第 2 段太冗」「結尾哲理改一下」「整體再溫暖一點」)
- 重生 = 整篇重寫,不做局部 patch(LLM 部分改動會破壞一致性)

---

### Step 8 — 自動結尾段(全程 silent,~30 秒)

Step 7 通過 A(使用者說「好」)後,進入此段。**對使用者只說一句**:

> 好,我來幫你組稿、找圖、發到 WordPress 草稿,大約 30 秒 ☕

之後不報「找圖中…」「排版中…」逐項進度,**只有失敗才講話**。

#### 8a — 找圖(內部執行)

從 draft 的 `topic` + `research` 萃取 **英文「城市+地點」** 關鍵字。命名規則:
- ✅ 正確:`Sun Moon Lake,Ita Thao,Xiangshan Visitor Center`(每個都帶城市/地點全名)
- ❌ 錯誤:`Lake,Ita Thao,Center`(模糊)

執行:
1. `write_file` 建 `request.json`:`{"locations":"<英文關鍵字逗號分隔>"}`(注意 **絕對不要在 shell 指令裡內嵌 JSON**)
2. shell:`curl --location "https://script.google.com/macros/s/AKfycbxlQSTNpSifs9t6gt-0QNYPuE8ui3dXn7O6v7akOby0gwLR6EVBlrb_CQhGSajpYo30/exec" -H "Content-Type: application/json" -d @request.json`
   - Windows 注意:若 `curl` 被 PowerShell alias 攔截,改用 `curl.exe`
3. 從 API 回傳挑 **3-4 張**,**選圖原則依 persona.md 的「視覺風格偏好」**

**失敗處理**:API 沒回應 / 圖數不足 4 張 → **不擋發布,改用無圖純文字版本繼續**,在 8d 結束的回報訊息加一行:`⚠️ 這次圖片 API 沒抓到圖,我先發純文字草稿,你可以到後台手動補圖`

#### 8b — 組完整 HTML(內部執行)

從 draft 取出 `h1` / `h2_outline` / `h3_subheadings` / `faq` / `full_text` 組裝。HTML 骨架嚴格依 `.gemini/skills/persona-writer/SKILL.md` 的 HTML 骨架規範(`<!DOCTYPE html>`、`<head>` 含 title/description/keywords meta、`<body>` 含內容 + 圖片、結尾 JSON-LD)。

**JSON-LD**:文末加 `application/ld+json`,含 `BlogPosting`(作者用 persona.display_name)+ `FAQPage`(對應 8a 階段的 FAQ)。
> 注意:`wp_poster.py` 發文前會剝掉 `<script>` 標籤(避免 wp.com 把 JSON-LD 內容洩漏到內文)。**JSON-LD 只存在本地 HTML 檔,沒上 WordPress**,跟現有行為一致。

**視覺整合**:從 `blog-visual-styles` 選用 2-3 種(雜誌封面、全景雙圖、圓形提示、背景引言),具體用哪幾種看 persona.md。每張 `<img>` 必有 `alt`(至少 1 張含焦點關鍵字)+ 圖片下方 `<figcaption>`(以該人格口吻撰寫)。

**內部連結**:讀 `personas/<slug>/published.json`(若存在),挑 1-2 篇主題相關的舊文章插入 link。**只連結同人格自己的舊文章**,不跨人格。

**外部連結**:至少 1 個指向權威來源(景點官網、維基百科)。

#### 8c — SEO 結構自我檢查(LLM 自評,不用程式 lint)

組好 HTML 後,Gemini 自己讀一遍,檢查:

- [ ] 焦點關鍵字密度 0.5–1.5%(從 draft.keywords[0] 取焦點關鍵字)
- [ ] 焦點關鍵字有出現在:title、description、第一段、≥2 個 H2、≥1 張 alt
- [ ] H1 只有一個
- [ ] H2 ≥ 3 個
- [ ] 每段 ≤ 150 字
- [ ] 每張 `<img>` 都有描述性 alt、≥1 張 alt 含焦點關鍵字
- [ ] 過渡詞使用率 ≥ 30%(「此外」「值得一提的是」「不過」「換句話說」…)

**不通過**:Gemini 自己修。**連修 3 次還過不了** → 直接發出去,在 8d 結束的回報加一行:`⚠️ SEO 結構檢查有幾項沒達標,你可以到後台調整`

#### 8d — 存 HTML + 發 wp_poster.py(內部執行)

執行:
1. `write_file`:`.gemini/skills/persona-writer/personas/<slug>/articles/<YYYYMMDD>-<location>-<topic>.html`(用最終定型的 HTML)
2. shell:`python3 .gemini/skills/persona-writer/scripts/wp_poster.py <persona-slug> "<H1 標題>" "<HTML 檔路徑>" draft`

**強制 `draft`**。除非使用者**在 Step 1 就明說「直接公開」**,否則一律 draft。

`wp_poster.py` 內部會:
- 依 `wp-config.json` 的 `auth_method` 走 Application Password 或 OAuth2
- 剝掉 `<script>` 標籤再發
- 成功時自動 append `published.json`(同人格)
- 失敗時印 `❌ 失敗。狀態碼: ...`

**發布失敗**:
- **draft.json 不刪**
- 對使用者回報:`⚠️ 發布到 WordPress 沒通,錯誤訊息:<vstrip>。對照模組 4 FAQ 看怎麼處理`
- draft 留著,使用者排錯後說「再試一次發布」可以從 draft 接續(讀 draft → 跳到 8a 或 8d,依失敗點而定)

#### 8e — 清掉 draft 暫存檔(只有發布成功才做)

**確認 wp_poster 印 `✅ 成功`** 之後執行:
- `os.remove(".gemini/skills/persona-writer/personas/<slug>/articles/draft-<...>.json")`

最終 HTML(`<YYYYMMDD>-<location>-<topic>.html`)**保留**,當本地完成品備份。

#### 最終回報訊息(發布成功後)

格式:

```
✅ 寫好囉!
📄 標題:<H1>
🔗 草稿連結:<WordPress 後台編輯連結>
🖼️ 用了 X 張圖、寫了約 X 字
🎯 焦點關鍵字:<keywords[0]>

點上面連結進去後,你可以再潤稿,確認沒問題就按 WordPress 右上角的「發布」。
```

(若 8a 圖失敗 / 8c SEO 沒達標,**結尾加警示行**)

> **不**貼 HTML 原始碼、JSON-LD、SEO 評分細節給使用者,他們不需要。
```

- [ ] **Step 2: Trace 驗證**

口頭走:
- Step 7 通過 A → Step 8 開頭只說一句「30 秒」→ 8a 找圖成功 → 8b 組 HTML → 8c SEO 自評過 → 8d 發布成功 → 8e 刪 draft → 回報
- 8a 失敗 → 繼續 8b 但不插圖 → 結尾加警示
- 8c 連修 3 次不過 → 直接發 → 結尾加警示
- 8d 失敗 → draft.json 不刪 + 對使用者報錯
- 8e 路徑模式對:`personas/<slug>/articles/draft-<YYYYMMDD>-<HHMM>-<topic-slug>.json`

預期:每分支 SOP 都有具體指示,沒有「視情況」「適當處理」這種 placeholder。

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "feat(factory): module 2 — step 8 auto-bundle (images/html/seo/publish/cleanup)"
```

---

## Task 5: 加入「錯誤與中斷處理」subsection + 結尾整理

**目標**:在 Task 4 結尾後 append 一個「錯誤與中斷處理」彙整 subsection,然後加上分隔線 `---` 收尾,讓 Module 2 跟 Module 3(下一節)正確分開。

**Files:**
- Modify: `.gemini/skills/marketing-content-factory/SKILL.md` (在 Task 4 結尾 append 錯誤處理 + 結尾分隔線)

- [ ] **Step 1: Edit append 錯誤處理 subsection + 結尾**

old_string:Task 4 結尾的 `> **不**貼 HTML 原始碼、JSON-LD、SEO 評分細節給使用者,他們不需要。`

new_string(保留 old_string + append):

```markdown
> **不**貼 HTML 原始碼、JSON-LD、SEO 評分細節給使用者,他們不需要。

---

### 模組 2 錯誤與中斷情境彙整

| 情境 | 處理 |
|---|---|
| 使用者 session 中斷後回來 | Step 1 開頭掃 `articles/draft-*.json` → Step 1B 問「上次寫到 X 階段,要繼續嗎?」 |
| 同人格多份 draft 並存 | Step 1B 列出來讓使用者挑要繼續 / 放棄哪份 |
| 使用者回應模糊(B 還 C?) | 走「模糊回應 - 一律問回去」表格 |
| 同階段重生 ≥5 次 | 走「重生上限」軟限制,主動建議「用最新版往下」,不強制擋 |
| 跨多階段回退(H2 階段要改 H1) | 走「跨階段回退」,引導「乾脆從頭來?還是只改最近一階段?」 |
| 找圖 API 失敗 | 8a 不擋發布,純文字版繼續 + 結尾加警示行 |
| SEO 自評連修 3 次不過 | 8c 直接發 + 結尾加警示行 |
| WordPress 發布失敗 | 8d draft 不刪 + 對使用者報錯 + 對照模組 4 FAQ;使用者說「再試一次發布」可從 draft 接續 |
| 使用者中途說「不寫了 / 放棄」 | 二次確認「確定放棄這篇『<topic>』嗎?草稿會被清掉」→ 是 → `os.remove draft-*.json` |
| 使用者中途要換人格 | 二次確認「之前寫的 H1/H3/... 會作廢,真的要換 OOO 嗎?」→ 是 → 刪舊 draft → 走新的 Step 1A(新 draft) |
| Step 1 前:該人格沒設定 WordPress | Step 1A-3 中止 → 引導模組 1 |
| Step 1 前:人格不存在 | Step 1A-2 列出可用人格、或引導模組 5 新增 |

---
```

> 重點:結尾的 `---` 把 Module 2 跟 Module 3(下一節 `## ⌨️ 模組 3:常用指令範例(複製貼上)`)正確分開。

- [ ] **Step 2: 讀整段模組 2 做最終一次檢查**

Run: `Read` tool 讀 `marketing-content-factory/SKILL.md`,範圍涵蓋整個新模組 2。檢查清單:

- [ ] Module 2 標題到 Module 3 標題之間沒有 placeholder (`TBD` / `TODO` / `視情況` / `適當處理`)
- [ ] 7 個 step 流程從頭到尾連貫(Step 1A/1B → 2 → 3 → 4 → 5 → 6 → 7 → 8a-e)
- [ ] 共用樣板被 Step 3-7 引用且夠用
- [ ] 錯誤處理表格涵蓋設計 spec §7 所有情境
- [ ] 結尾 `---` 把 Module 2 結束、跟 Module 3 隔開

如果發現任一項不符,直接 Edit 修。

- [ ] **Step 3: Commit**

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md
git commit -m "feat(factory): module 2 — error & interruption handling subsection (final)"
```

---

## Task 6: 重寫 persona-writer SKILL.md SOP 段落(stage-based execution)

**目標**:把 `persona-writer/SKILL.md` line 32-134 的 SOP 段落整段替換成「以 draft.json `stage` 欄位驅動」的版本。每個 stage 明訂該讀什麼欄位 / 該寫什麼欄位 / 該推進到哪。

**Files:**
- Modify: `.gemini/skills/persona-writer/SKILL.md` (line 32-134 整段替換,加自我檢查清單 line 194-206 補 2 條)

- [ ] **Step 1: Read 確認邊界**

Read `persona-writer/SKILL.md` line 30-136 確認:
- line 32 `## 📝 標準作業程序 (Standard Operating Procedure)`
- line 135 `## 🎨 HTML 骨架(必須嚴格遵守)`(SOP 結束)

- [ ] **Step 2: Edit 整段替換 SOP**

old_string:`## 📝 標準作業程序 (Standard Operating Procedure)` 開始,到 `## 🎨 HTML 骨架(必須嚴格遵守)` 之前的最後一行(整段 SOP)

new_string:

```markdown
## 📝 標準作業程序 (Stage-Based Execution)

> 這份 SKILL 過去是「6 步 SOP 一氣呵成」,**現在改成 `draft.json` 階段接力**。每次被 `marketing-content-factory` 模組 2 委派任務時,**先讀 draft.json 的 `stage` 欄位**,知道現在該做哪一階段,只執行那一階段的工作,再寫回去,結束。

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
- **格式**:H1 用 `# 標題`、H2 用 `## ...`、H3 用 `### ...`、FAQ 用 `### Q1: ... A1: ...`
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
```

- [ ] **Step 3: Edit 補強自我檢查清單**

old_string(line 194-206 整段):

```markdown
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
```

new_string:

```markdown
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
```

- [ ] **Step 4: Trace 驗證**

口頭走:
- Module 2 委派 → persona-writer 拿到 `persona-slug` + `draft.json 路徑`
- 讀 draft 看到 `stage="init"` → 跳到「init → 跑 research」階段 → 做完更新 stage
- 下次委派 stage=`research_done` → 跳到「生 H1」階段 → 把 H1 字串交回(不寫 draft)
- 使用者 OK → marketing-content-factory 寫 draft + stage h1_done
- ... 走到 `full_text_done` → 委派執行 8a-8e

預期:每階段契約清楚,跟 marketing-content-factory Module 2 對應得上,無欄位名稱不一致。

- [ ] **Step 5: Commit**

```bash
git add .gemini/skills/persona-writer/SKILL.md
git commit -m "refactor(persona-writer): stage-based execution driven by draft.json"
```

---

## Task 7: End-to-end manual walkthrough + final cleanup

**目標**:把所有可預見情境用對話完整 trace 一遍,確認兩份 SKILL.md 改完後使用者體驗完整。發現任何缺漏或不一致,inline 補上。

**Files:** (僅讀取與微調 - 不會做大改)
- Read: 兩份 SKILL.md 最新版本
- Modify (only if issues found): 對應檔案

- [ ] **Step 1: Trace 情境 1 — 全新文章流暢路徑**

開場訊息:「幫我用林太寫日月潭兩天一夜慢旅,關鍵字日月潭慢旅, 湖畔散策, 私房茶席」

預期 SOP 走完:
1. 模組 2 Step 1 掃 0 份 draft → 走 1A
2. 1A-1 收主題+關鍵字 → 1A-2 用林太(明指)→ 1A-3 wp-config 存在 → 1A-4 建 draft
3. Step 2 silent research → 寫 research + research_done
4. Step 3 H1 → 使用者「好」 → h1_done
5. Step 4 H3 → 使用者「好」 → h3_done
6. Step 5 H2 → 使用者「好」 → h2_done
7. Step 6 FAQ → 使用者「好」 → faq_done
8. Step 7 全文 → 使用者「好」 → full_text_done
9. Step 8a 找圖 OK → 8b HTML → 8c SEO 自評 OK → 8d 發 draft OK → 8e 刪 draft → 最終回報

每一步的 SOP 都該有清楚指示。發現缺漏 → Edit 補。

- [ ] **Step 2: Trace 情境 2 — 中斷接續**

寫到 Step 5 (h2_done) 後 session 中斷,下次開場:「我要寫文章」

預期:
1. 模組 2 Step 1 掃到 1 份 draft → 走 1B
2. 列出該 draft「日月潭…,停在 H2 大綱已確認階段」
3. 使用者「繼續」 → 跳到 Step 6 (faq)
4. ... 後續正常走完

預期 SOP 的 stage→step 對應表正確覆蓋每個 stage 值。

- [ ] **Step 3: Trace 情境 3 — 重生修改**

Step 4 (H3) 收到「拿掉第 5 點,加一個關於茶廠的點」

預期:
1. 走 C(下指令重生)
2. AI 簡短覆述方向「好,拿掉住宿建議、加上茶廠的點 ──」
3. 重生 5-7 個 H3(含茶廠)
4. 再次三段式呈現

- [ ] **Step 4: Trace 情境 4 — 直接覆寫**

Step 3 (H1) 收到使用者貼「日月潭慢時光｜林太寫給熟齡旅人的湖畔散策」

預期:
1. 走 B(覆寫接受)
2. AI 二次確認「收到,用你這版繼續嗎?」
3. 使用者「好」 → 走 A 寫入 + 進 Step 4

- [ ] **Step 5: Trace 情境 5 — 中途放棄**

Step 5 (H2) 收到「不寫了,放棄」

預期:
1. 二次確認「確定放棄這篇『日月潭兩天一夜慢旅』嗎?草稿會被清掉」
2. 使用者「確定」 → `os.remove draft-*.json` → 回到一般對話模式(可以選別的模組)

- [ ] **Step 6: Trace 情境 6 — 換人格**

開場「用林太寫日月潭」進行到 Step 5 後,使用者「不對,我想用王老闆寫」

(假設 王老闆/mr-wang-foodie persona 存在)
預期:
1. 二次確認「之前寫的 H1/H3/H2 會作廢,真的換王老闆嗎?」
2. 使用者「對」 → 刪舊 draft → 走新的 1A(王老闆 + 新 draft)

- [ ] **Step 7: Trace 情境 7 — 找圖失敗 / SEO 不過 / 發布失敗**

Step 7 (全文) 通過後進入 8a 連 3 次重試圖片 API 還是失敗:
- 預期:不擋,8b 用無圖純文字版 → 8c → 8d 發 → 最終回報結尾加警示

8c 連修 3 次 SEO 還是有 1 項沒達標:
- 預期:直接 8d → 結尾警示

8d wp_poster 印 `❌ 失敗`:
- 預期:draft 不刪 + 對使用者報錯,引導模組 4 FAQ

- [ ] **Step 8: 修補發現的問題**

如果上述任一情境 trace 不通,直接 Edit 對應 SKILL.md 補上缺漏。修完 commit:

```bash
git add .gemini/skills/marketing-content-factory/SKILL.md .gemini/skills/persona-writer/SKILL.md
git commit -m "fix(factory|persona-writer): trace-test fixes from end-to-end walkthrough"
```

如果 trace 7 個情境都通,**跳過 commit**(無變動)。

- [ ] **Step 9: 最終 git log 確認**

Run: `git log --oneline -8`

Expected: 看到 Task 1-6(可能 7)的 commit 排在一起,沒有奇怪的中間態。

---

## Self-Review

對照 spec `docs/superpowers/specs/2026-05-20-stepwise-article-flow-design.md`:

| Spec 章節 | 對應 Task |
|---|---|
| §1 目標 | 整體計畫 |
| §2 背景 | 整體計畫 |
| §3 設計總覽 | Task 1, 2, 3, 4 |
| §4 狀態管理(draft.json 結構與生命週期) | Task 1 (建立)、Task 2-4 (累加)、Task 4 (清檔) |
| §5 對話互動格式(三段式 + 意圖分類 + 模糊回應 + 重生上限 + 跨階段回退) | Task 2 (共用樣板)、Task 3 (各階段套用) |
| §5.5 各階段具體呈現格式 | Task 3 |
| §6 Step 8 自動結尾段(8a-8e + 最終回報) | Task 4 |
| §7 錯誤與中斷處理 | Task 5(彙整 in factory)、Task 6(persona-writer 端)、Task 7(trace) |
| §8 要改的檔案清單 | Task 1-6 |
| §9 風險與相容性 | 整體計畫 - 純改 SOP、無腳本變動 |
| §10 不在範圍 | 整體計畫 - 不含 SEO meta 自動帶、自動 publish 等 |

✅ Spec 所有章節都對應到至少一個 Task。

**Placeholder scan**:
- ✅ 沒有 `TBD` / `TODO` / `implement later`
- ✅ 沒有 "fill in details" 之類的 vague requirements
- ✅ 每個 Edit 步驟都列了完整的 old_string / new_string 內容(new_string 在 plan 裡完整列出 prose)
- ✅ 每個 trace 步驟都列了具體的「使用者訊息 + 預期 SOP 路徑」

**Type / 一致性 check**:
- ✅ `draft.json` 欄位名稱在 Task 1 (建立)、Task 2 (research)、Task 3 (h1/h3/h2/faq/full_text)、Task 4 (清檔) 全篇使用同一組命名
- ✅ `stage` 值在 Task 1-6 都用同一張表(`init` / `research_done` / ...)
- ✅ `persona-slug` 命名在兩份 SKILL.md 一致
- ✅ Task 1B 的 stage→step 對應表 (`init`→Step 2、`research_done`→Step 3、...) 跟 Task 6 的 stage 狀態機一致

**Scope check**:
- 兩份 SKILL.md 都改在 Module 2 / SOP 段,沒擴散到其他模組
- 無腳本變動
- 單一 implementation plan 涵蓋整個 spec,不需要再分子計畫

Plan 完整、無 placeholder、Spec 全覆蓋。
