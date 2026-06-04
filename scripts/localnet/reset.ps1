param(
  [string]$OutputDir = ""
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }

& (Join-Path $PSScriptRoot "stop.ps1") -OutputDir $OutputDir
if (Test-Path $OutputDir) {
  Remove-Item -LiteralPath $OutputDir -Recurse -Force
  Write-Host "Removed $OutputDir"
}
