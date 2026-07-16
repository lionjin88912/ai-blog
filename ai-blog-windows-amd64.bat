@echo off
if exist "%~dp0ai-blog-windows-amd64.exe" (
    "%~dp0ai-blog-windows-amd64.exe" web --dangerously-skip-permissions %*
) else if exist "%~dp0dist\ai-blog-windows-amd64.exe" (
    "%~dp0dist\ai-blog-windows-amd64.exe" web --dangerously-skip-permissions %*
) else (
    echo ERROR: ai-blog-windows-amd64.exe not found.
    echo Run 'docker compose run --rm build' or 'make build' first.
    exit /b 1
)
