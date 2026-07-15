@echo off
REM Open Git Bash in a new window with all sandbox tools in PATH.
REM Use this from the web terminal or command line when you need real bash.

set "SCRIPT_DIR=%~dp0"

REM Find sandbox: next to this bat, or in parent directory
if exist "%SCRIPT_DIR%sandbox\git\bin\bash.exe" (
    set "SANDBOX_DIR=%SCRIPT_DIR%sandbox"
) else if exist "%SCRIPT_DIR%..\sandbox\git\bin\bash.exe" (
    set "SANDBOX_DIR=%SCRIPT_DIR%..\sandbox"
) else (
    echo ERROR: Portable Git not found.
    echo Run 'ai-sandbox setup' first.
    exit /b 1
)

set "PATH=%SANDBOX_DIR%\bin;%SANDBOX_DIR%\git\bin;%SANDBOX_DIR%\git\usr\bin;%SANDBOX_DIR%\node;%SANDBOX_DIR%\uv;%PATH%"

start "Git Bash" "%SANDBOX_DIR%\git\bin\bash.exe" --login -i
