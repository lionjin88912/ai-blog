---
name: tool-status-checker
description: Cross-platform environment checker. Verifies availability and versions of core sandbox tools (curl, uv, python, cat). Supports Windows (PowerShell), macOS, and Linux (Bash). Use when the user asks to check the environment, verify tools, or troubleshoot missing tools. After checking, always update TOOLS.md with the results.
---

# Tool Status Checker

This skill checks whether all required binary tools are correctly installed and accessible within the sandbox environment. It supports **Windows**, **macOS**, and **Linux**.

## Workflow

1. Detect the current OS.
2. Run the appropriate diagnostic script.
3. Parse the output for each tool's status and version.
4. **Python venv 自動建立**：若 `<project_root>/.venv` 不存在，用 uv 建立並安裝 Python 3.12。
5. **掃描 SKILL scripts 並安裝依賴**：掃描 `.gemini/skills/*/scripts/*.py`，解析 `import` 取得第三方套件，用 uv 安裝。
6. Update `TOOLS.md` at the project root with the results and today's date.

## Step 1: Run Diagnostic Script

### Windows (PowerShell)

```powershell
powershell.exe -ExecutionPolicy Bypass -File <project_root>\.gemini\skills\tool-status-checker\scripts\check_tools.ps1
```

### macOS / Linux (Bash)

```bash
bash <project_root>/.gemini/skills/tool-status-checker/scripts/check_tools.sh
```

## Step 2: Python venv 自動建立

在確認 uv 可用之後，檢查 `<project_root>/.venv` 是否存在。若不存在，自動建立：

### Windows (PowerShell)

```powershell
# 建立 venv（使用 sandbox 的 uv）
sandbox\uv\uv.exe venv .venv --python 3.12
```

### macOS / Linux (Bash)

```bash
# 建立 venv（使用 sandbox 的 uv）
sandbox/uv/uv venv .venv --python 3.12
```

> 若 `.venv` 已存在則跳過此步驟。

## Step 3: 掃描 SKILL scripts 並安裝依賴

掃描 `.gemini/skills/*/scripts/*.py` 中的 `import` 語句，找出第三方套件（排除標準庫），用 uv 安裝到 `.venv`。

### 判斷邏輯

1. 用 `read_file` 讀取每個 `.py` 檔案的前 20 行
2. 提取 `import xxx` 和 `from xxx import` 中的頂層模組名
3. 排除 Python 標準庫模組（`os`, `sys`, `json`, `re`, `datetime`, `time`, `pathlib`, `base64`, `hashlib`, `urllib`, `collections`, `typing`, `dataclasses`, `functools`, `io`, `math`, `string`, `copy`, `csv`, `logging`, `unittest`, `textwrap`, `argparse`, `shutil`, `glob`, `tempfile`, `subprocess`, `configparser`, `itertools`, `contextlib`, `abc`, `enum`, `struct`, `threading`, `multiprocessing`, `socket`, `http`, `email`, `html`, `xml`, `sqlite3`, `zipfile`, `gzip`, `tarfile`, `pprint`, `traceback`, `warnings`, `inspect`, `dis`, `ast`, `secrets`, `uuid`, `decimal`, `fractions`, `statistics`, `random`, `operator`, `signal`, `platform`）
4. 剩餘的模組即為第三方套件，執行安裝

### 安裝指令

**Windows：**
```powershell
sandbox\uv\uv.exe pip install --python .venv\Scripts\python.exe <package1> <package2> ...
```

**macOS / Linux：**
```bash
sandbox/uv/uv pip install --python .venv/bin/python <package1> <package2> ...
```

### 已知的模組名 → 套件名對應

部分模組的 `import` 名稱與 PyPI 套件名不同：

| import 名 | PyPI 套件名 |
| :--- | :--- |
| `cv2` | `opencv-python` |
| `PIL` | `Pillow` |
| `bs4` | `beautifulsoup4` |
| `sklearn` | `scikit-learn` |
| `yaml` | `pyyaml` |
| `dotenv` | `python-dotenv` |
| `gi` | `PyGObject` |

其他模組名通常與 PyPI 套件名相同（如 `requests` → `requests`）。

## Step 4: Update TOOLS.md

After running the script, update `TOOLS.md` at the project root with:
- The current date in the header.
- The detected OS/platform.
- Each tool's status (`[OK]`, `[FAILED]`, or `[NOT FOUND]`) and version string.
- Environment-specific paths (e.g., MINGW64 on Windows, Git sandbox on macOS/Linux).

Use the following format:

```markdown
# Core Tools Status

This file documents the status of core development tools in the sandbox environment, last checked on YYYY-MM-DD.

## Diagnostic Results

| Tool | Status | Version / Detail |
| :--- | :--- | :--- |
| **curl** | [STATUS] | version string |
| **uv** | [STATUS] | version string |
| **python** | [STATUS] | version string |
| **cat** | [STATUS] | version string |

## Environment

- **Platform**: Windows / macOS / Linux
- **MINGW64**: Found/Not Found at `path` (Windows only)
- **Python VENV**: Located at `<project_root>/.venv` (auto-created / already existed)
- **Installed Packages**: list of packages installed via uv
```

## Tools Monitored

| Tool | Windows | macOS / Linux |
| :--- | :--- | :--- |
| **curl** | `curl.exe --version` | `curl --version` |
| **uv** | `sandbox\uv\uv.exe --version` | `sandbox/uv/uv --version` |
| **python** | `sandbox\python\cpython-*\python.exe --version` | `sandbox/python/cpython-*/python3* --version` |
| **cat** | `sandbox\git\usr\bin\cat.exe --version` | `cat --version` |
| **MINGW64** | Directory at `sandbox\git\mingw64` | N/A |

## Troubleshooting

If a tool is reported as **[NOT FOUND]** or **[FAILED]**:
1. Check the `sandbox/` directory structure under the project root.
2. Ensure the paths in the check script match the current project layout.
3. Re-run `gemini-cli setup` to re-download missing tools.
