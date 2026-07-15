$ErrorActionPreference = "Stop"

if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) {
    Write-Error "WSL is not installed or not available in PATH."
    exit 2
}

function Convert-WindowsPathToWsl {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    if ($fullPath -match '^(?<drive>[A-Za-z]):\\(?<rest>.*)$') {
        $drive = $Matches.drive.ToLowerInvariant()
        $rest = $Matches.rest -replace '\\', '/'
        if ([string]::IsNullOrEmpty($rest)) {
            return "/mnt/$drive"
        }
        return "/mnt/$drive/$rest"
    }

    throw "Only local drive paths are supported: $fullPath"
}

$windowsCwd = (Get-Location).Path
$wslCwd = Convert-WindowsPathToWsl -Path $windowsCwd

if (-not $wslCwd) {
    Write-Error "Failed to resolve current directory for WSL: $windowsCwd"
    exit 2
}

wsl --cd "$wslCwd" ".github/hooks/secrets-scanner/scan-secrets.sh"
exit $LASTEXITCODE
