param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 0,
  [switch]$CleanLogs,
  [switch]$Restart
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -OutputDir $OutputDir
$Binary = Resolve-BinaryPath -Binary $Binary
$requestedCount = $ValidatorCount
$ValidatorCount = Get-ManifestValidatorCount -OutputDir $OutputDir -ValidatorCount $ValidatorCount

if (!(Test-Path -LiteralPath $Binary) -or !(Test-Path -LiteralPath $OutputDir)) {
  & (Join-Path $PSScriptRoot "init.ps1") -OutputDir $OutputDir -Binary $Binary -ValidatorCount $ValidatorCount
}

$manifest = Read-LocalnetManifest -OutputDir $OutputDir
if ($null -eq $manifest) {
  throw "localnet manifest not found in $OutputDir; run scripts/localnet/init.ps1 first"
}
if ($requestedCount -gt 0 -and [int]$manifest.validator_count -ne $requestedCount) {
  throw "requested validator count $requestedCount does not match initialized manifest count $($manifest.validator_count)"
}
$ValidatorCount = [int]$manifest.validator_count

$pidDir = Join-Path $OutputDir "pids"
$logDir = Join-Path $OutputDir "logs"
New-Item -ItemType Directory -Force -Path $pidDir, $logDir | Out-Null
if ($CleanLogs) {
  Get-ChildItem -LiteralPath $logDir -Filter "*.log" -ErrorAction SilentlyContinue | Remove-Item -Force
}

$started = @()
try {
  for ($i = 0; $i -lt $ValidatorCount; $i++) {
    $pidFile = Join-Path $pidDir "node$i.pid"
    if (Test-Path -LiteralPath $pidFile) {
      $pidValue = [int](Get-Content -Raw -LiteralPath $pidFile)
      $existing = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      if ($existing) {
        if (!$Restart) {
          Write-Host "node$i already running pid=$pidValue"
          continue
        }
        Stop-Process -Id $pidValue -Force
      }
      Remove-Item -LiteralPath $pidFile -Force
    }

    $nodeHome = Get-NodeHome -OutputDir $OutputDir -Index $i
    if (!(Test-Path -LiteralPath $nodeHome)) {
      throw "node home missing for node${i}: $nodeHome"
    }
    $stdout = Join-Path $logDir "node$i.out.log"
    $stderr = Join-Path $logDir "node$i.err.log"
    $appMetricsPort = $manifest.nodes[$i].ports.app_metrics
    if (-not $appMetricsPort) {
      $appMetricsPort = 27660 + ($i * 100)
    }
    $appMetricsAddr = "127.0.0.1:$appMetricsPort"
    $proc = Start-Process -FilePath $Binary `
      -ArgumentList @("start", "--home", $nodeHome, "--log_level", "info", "--observability-metrics", "--observability-metrics-addr", $appMetricsAddr) `
      -RedirectStandardOutput $stdout `
      -RedirectStandardError $stderr `
      -WindowStyle Hidden `
      -PassThru
    Start-Sleep -Milliseconds 500
    if ($proc.HasExited) {
      throw "node$i exited immediately with code $($proc.ExitCode); see $stderr"
    }
    Set-Content -LiteralPath $pidFile -Value $proc.Id
    $started += $proc.Id
    Write-Host "Started node$i pid=$($proc.Id)"
  }
}
catch {
  foreach ($pidValue in $started) {
    Stop-Process -Id $pidValue -Force -ErrorAction SilentlyContinue
  }
  throw
}

Write-Host "Logs: $logDir"
