param(
  [string]$OutputDir = "",
  [string]$Binary = ""
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }
if ($Binary -eq "") { $Binary = Join-Path $RepoRoot "build\orbitalisd.exe" }

if (!(Test-Path $Binary) -or !(Test-Path $OutputDir)) {
  & (Join-Path $PSScriptRoot "init.ps1") -OutputDir $OutputDir -Binary $Binary
}

$pidDir = Join-Path $OutputDir "pids"
$logDir = Join-Path $OutputDir "logs"
New-Item -ItemType Directory -Force -Path $pidDir, $logDir | Out-Null

for ($i = 0; $i -lt 3; $i++) {
  $nodeHome = Join-Path $OutputDir "node$i\orbitalisd"
  $stdout = Join-Path $logDir "node$i.out.log"
  $stderr = Join-Path $logDir "node$i.err.log"
  $proc = Start-Process -FilePath $Binary `
    -ArgumentList @("start", "--home", $nodeHome) `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -WindowStyle Hidden `
    -PassThru
  Set-Content -LiteralPath (Join-Path $pidDir "node$i.pid") -Value $proc.Id
  Write-Host "Started node$i pid=$($proc.Id)"
}

Write-Host "Logs: $logDir"
