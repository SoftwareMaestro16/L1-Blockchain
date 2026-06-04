param(
  [string]$OutputDir = ""
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }
$pidDir = Join-Path $OutputDir "pids"

if (!(Test-Path $pidDir)) {
  Write-Host "No pid directory found at $pidDir"
  exit 0
}

Get-ChildItem -LiteralPath $pidDir -Filter *.pid | ForEach-Object {
  $pidValue = [int](Get-Content -Raw -LiteralPath $_.FullName)
  $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
  if ($proc) {
    Stop-Process -Id $pidValue -Force
    Write-Host "Stopped pid=$pidValue"
  }
  Remove-Item -LiteralPath $_.FullName -Force
}
