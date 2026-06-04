param(
  [string]$OutputDir = "",
  [int]$TimeoutSeconds = 10
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -OutputDir $OutputDir
$pidDir = Join-Path $OutputDir "pids"

if (!(Test-Path -LiteralPath $pidDir)) {
  Write-Host "No pid directory found at $pidDir"
  return
}

Get-ChildItem -LiteralPath $pidDir -Filter "*.pid" | ForEach-Object {
  $pidText = (Get-Content -Raw -LiteralPath $_.FullName).Trim()
  if ([string]::IsNullOrWhiteSpace($pidText)) {
    Remove-Item -LiteralPath $_.FullName -Force
    return
  }
  $pidValue = [int]$pidText
  $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
  if ($proc) {
    Stop-Process -Id $pidValue -ErrorAction SilentlyContinue
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
      if (!(Get-Process -Id $pidValue -ErrorAction SilentlyContinue)) {
        break
      }
      Start-Sleep -Milliseconds 250
    }
    if (Get-Process -Id $pidValue -ErrorAction SilentlyContinue) {
      Stop-Process -Id $pidValue -Force -ErrorAction SilentlyContinue
    }
    Write-Host "Stopped pid=$pidValue"
  }
  Remove-Item -LiteralPath $_.FullName -Force
}
