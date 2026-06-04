param(
  [string]$OutputDir = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -OutputDir $OutputDir
Assert-SafeLocalnetPath -Path $OutputDir

& (Join-Path $PSScriptRoot "stop.ps1") -OutputDir $OutputDir
if (Test-Path -LiteralPath $OutputDir) {
  Remove-Item -LiteralPath $OutputDir -Recurse -Force
  Write-Host "Removed $OutputDir"
}
