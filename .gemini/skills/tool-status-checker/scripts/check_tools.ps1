# Tool Status Checker Script for Windows Gemini CLI Environment

function Check-Tool {
    param([string]$name, [string]$command, [string]$versionArgs = "--version")
    Write-Host "Checking $name..." -NoNewline
    try {
        $out = & $command $versionArgs 2>&1
        if ($LASTEXITCODE -eq 0 -or $out -match "version") {
            $ver = ($out | Select-Object -First 1).ToString().Trim()
            Write-Host " [OK] ($ver)" -ForegroundColor Green
        } else {
            Write-Host " [FAILED]" -ForegroundColor Red
        }
    } catch {
        Write-Host " [NOT FOUND]" -ForegroundColor Yellow
    }
}

Write-Host "=== Core Tools Status ==="
Check-Tool "curl" "curl.exe"
Check-Tool "uv" "E:\gh_gemini_ai_cli\sandbox\uv\uv.exe"
Check-Tool "python" "E:\gh_gemini_ai_cli\sandbox\python\cpython-3.12.10-windows-x86_64-none\python.exe"
Check-Tool "cat" "E:\gh_gemini_ai_cli\sandbox\git\usr\bin\cat.exe"

Write-Host "`n=== Environment Check ==="
$mingwPath = "E:\gh_gemini_ai_cli\sandbox\git\mingw64"
if (Test-Path $mingwPath) {
    Write-Host "MINGW64: [FOUND] at $mingwPath" -ForegroundColor Green
} else {
    Write-Host "MINGW64: [NOT FOUND]" -ForegroundColor Red
}
