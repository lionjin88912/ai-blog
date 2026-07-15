---
name: dot-skill-distiller
description: 透過 dot-skill 框架提取個人特質、工作專業與決策模型，並將其轉化為可執行的 AI 技能（SKILL）。
---

# Dot-Skill Distiller (Persona & Expertise Extraction)

此技能使代理能夠使用 `dot-skill` 框架來「蒸餾」特定人物（同事、導師、公眾人物）的專業知識、溝通風格與決策框架。

## 核心流程 (Workflows)

### 1. 資料收集與攝入 (Data Ingestion)
當使用者想要複刻某個人的專業技能或風格時：
1.  **來源分析**：識別數據來源（Slack/飛書日誌、PDF 文件、電子郵件、訪談紀錄）。
2.  **工具運用**：呼叫對應的腳本進行自動化採集（如 `feishu_auto_collector.py` 或 `slack_ingest.py`）。

### 2. 人格與專業分析 (Analysis Pipeline)
將收集到的數據轉化為結構化的技能描述：
1.  **Persona 提取**：分析語氣、慣用詞彙與情緒反應模式。
2.  **Work Skill 提取**：分析其技術標準、工作流程與決策規則（Playbook）。
3.  **六維度建檔**：涵蓋作品集、訪談紀錄、公開言論、重大決策、工作準則及他人評價。

### 3. 生成與部署 SKILL.md
根據分析結果，產生符合 AgentSkills 標準的檔案：
1.  使用 `persona_builder.md` 與 `work_analyzer.md` 模板。
2.  輸出的 `SKILL.md` 應包含：
    *   **Persona 層**：定義「我是誰」以及「我如何說話」。
    *   **Work 層**：定義「我如何工作」以及「我的專業規則」。

### 4. 進化與修正 (Refinement)
1.  **反饋迭代**：使用者提供修正建議（例如：「如果是他，他不會這樣回答」）。
2.  **版本控制**：記錄技能的變動歷史，確保人格穩定性。

## 核心分類 (Character Families)
*   **Colleague (同事)**：側重技術標準與工作姿勢（Work Posture）。
*   **Relationship (關係)**：側重情感觸發點與表達 DNA。
*   **Celebrity (名人)**：側重其心智模型與決策啟發法。

## 最佳實踐 (Best Practices)
*   **去中心化數據**：優先收集真實的對話與決策紀錄，而非僅僅是個人簡介。
*   **層次分離**：確保人格（Style）與專業（Skill）在檔案中是分離的，以便獨立調優。
*   **隱私保護**：在蒸餾過程中務必遵守隱私規範，去除敏感資訊。

## 錯誤處理 (Troubleshooting)
*   **人格崩壞 (Persona Decay)**：若 AI 回應不符合預期，應回溯至數據分析階段，補足更多原始紀錄。
*   **數據不足**：當數據量太小時，優先建立 Work Skill 層，而非強行模擬 Persona。
