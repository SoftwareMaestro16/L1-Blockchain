param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 0,
  [string]$ChainId = "aetra-local-1",
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
  [string]$Profile = "base",
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [switch]$NoInit,
  [switch]$Wait,
  [switch]$CleanLogs,
  [int]$TimeoutSeconds = 60
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$startupStartedAt = Get-Date
$startupTimer = [System.Diagnostics.Stopwatch]::StartNew()
$initElapsedMs = 0
$launchElapsedMs = 0
$waitElapsedMs = 0

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
Assert-LocalnetProfile -Profile $Profile

if (!(Test-Path $Binary) -or !(Test-Path $OutputDir)) {
  if ($NoInit) {
    throw "Binary or output directory is missing and -NoInit was specified: binary=$Binary output=$OutputDir"
  }
  $initArgs = @{
    OutputDir      = $OutputDir
    Binary         = $Binary
    ChainId        = $ChainId
    BaseP2PPort    = $BaseP2PPort
    BaseRPCPort    = $BaseRPCPort
    BaseRESTPort   = $BaseRESTPort
    BaseGRPCPort   = $BaseGRPCPort
    BasePprofPort  = $BasePprofPort
    PortStride     = $PortStride
    TimeoutCommit  = $TimeoutCommit
    LogLevel       = $LogLevel
    Profile        = $Profile
    EnableAPI      = $EnableAPI
    EnableGRPC     = $EnableGRPC
    EnableRPC      = $EnableRPC
  }
  if ($ValidatorCount -gt 0) { $initArgs.ValidatorCount = $ValidatorCount }
  $initTimer = [System.Diagnostics.Stopwatch]::StartNew()
  & (Join-Path $PSScriptRoot "init.ps1") @initArgs
  $initTimer.Stop()
  $initElapsedMs = [int64]$initTimer.ElapsedMilliseconds
}

$nodes = Get-LocalnetNodes -OutputDir $OutputDir

if ($ValidatorCount -gt 0 -and $nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount validator nodes under $OutputDir, found $($nodes.Count)"
}
if ($nodes.Count -lt 1) {
  throw "No validator node directories found under $OutputDir"
}
$actualValidatorCount = $nodes.Count

$pidDir = Join-Path $OutputDir "pids"
$logDir = Join-Path $OutputDir "logs"
New-Item -ItemType Directory -Force -Path $pidDir, $logDir | Out-Null
if ($CleanLogs) {
  Get-ChildItem -LiteralPath $logDir -Filter "*.log" -ErrorAction SilentlyContinue | Remove-Item -Force
}

Get-ChildItem -LiteralPath $pidDir -Filter *.pid -ErrorAction SilentlyContinue | ForEach-Object {
  $pidValue = [int](Get-Content -Raw -LiteralPath $_.FullName)
  $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
  if ($proc) {
    throw "Localnet already appears to be running: $($_.FullName) pid=$pidValue"
  }
  Remove-Item -LiteralPath $_.FullName -Force
}

Assert-LocalnetPortsAvailable `
  -ValidatorCount $actualValidatorCount `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BasePprofPort $BasePprofPort `
  -PortStride $PortStride `
  -EnableAPI $EnableAPI `
  -EnableGRPC $EnableGRPC `
  -EnableRPC $EnableRPC

Repair-LocalnetProcessPathEnvironment

$launchTimer = [System.Diagnostics.Stopwatch]::StartNew()
foreach ($node in $nodes) {
  $nodeName = $node.Name
  $nodeHome = Join-Path $node.FullName "aetrad"
  $stdout = Join-Path $logDir "$nodeName.out.log"
  $stderr = Join-Path $logDir "$nodeName.err.log"
  $proc = Start-Process -FilePath $Binary `
    -ArgumentList @("start", "--home", $nodeHome, "--log_level", $LogLevel) `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -WindowStyle Hidden `
    -PassThru
  Set-Content -LiteralPath (Join-Path $pidDir "$nodeName.pid") -Value $proc.Id
  Write-Host "Started $nodeName pid=$($proc.Id)"
}
$launchTimer.Stop()
$launchElapsedMs = [int64]$launchTimer.ElapsedMilliseconds

Write-Host "Logs: $logDir"

if ($Wait) {
  $waitTimer = [System.Diagnostics.Stopwatch]::StartNew()
  $p = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
  if ($EnableRPC) {
    Wait-LocalnetRpc -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Wait-LocalnetHeight -TargetHeight 1 -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
  if ($EnableAPI) {
    Wait-LocalnetRest -RESTPort $p.REST -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
  if ($EnableGRPC) {
    Wait-LocalnetGrpc -GRPCPort $p.GRPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
  $waitTimer.Stop()
  $waitElapsedMs = [int64]$waitTimer.ElapsedMilliseconds
}

$startupTimer.Stop()
$timing = [ordered]@{
  script                 = "localnet/start.ps1"
  chain_id               = $ChainId
  validators             = $actualValidatorCount
  profile                = $Profile
  wait                   = [bool]$Wait
  started_at_utc         = $startupStartedAt.ToUniversalTime().ToString("o")
  total_ms               = [int64]$startupTimer.ElapsedMilliseconds
  init_ms                = $initElapsedMs
  process_launch_ms      = $launchElapsedMs
  health_wait_ms         = $waitElapsedMs
  output_dir             = $OutputDir
  logs_dir               = $logDir
}
$timingPath = Join-Path $logDir "startup-timing.json"
$timing | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $timingPath
Write-Host "Startup timing: total_ms=$($timing.total_ms) init_ms=$($timing.init_ms) process_launch_ms=$($timing.process_launch_ms) health_wait_ms=$($timing.health_wait_ms)"
Write-Host "Startup timing summary: $timingPath"
