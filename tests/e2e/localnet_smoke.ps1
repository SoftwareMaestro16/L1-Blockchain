param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [switch]$SkipBuild,
  [switch]$SkipBankTx,
  [switch]$SkipNegativeChecks
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"

function Invoke-WithCommonLocalnetArgs {
  param(
    [string]$ScriptPath,
    [hashtable]$Extra = @{}
  )

  $args = @{
    OutputDir     = $OutputDir
    Binary        = $Binary
    ChainId       = $ChainId
    ValidatorCount = $ValidatorCount
    BaseP2PPort   = $BaseP2PPort
    BaseRPCPort   = $BaseRPCPort
    BaseRESTPort  = $BaseRESTPort
    BaseGRPCPort  = $BaseGRPCPort
    BasePprofPort = $BasePprofPort
    PortStride    = $PortStride
    TimeoutCommit = $TimeoutCommit
    LogLevel      = $LogLevel
    EnableAPI     = $EnableAPI
    EnableGRPC    = $EnableGRPC
    EnableRPC     = $EnableRPC
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function Assert-Throws {
  param(
    [scriptblock]$Script,
    [string]$Name
  )

  try {
    & $Script
  } catch {
    Write-Host "negative check passed: $Name"
    return
  }
  throw "negative check did not fail: $Name"
}

function Test-LocalnetNegativeCases {
  $negativeRoot = Resolve-LocalnetPath -Path ".work\localnet-negative" -DefaultRelativePath ".work\localnet-negative"
  New-Item -ItemType Directory -Force -Path $negativeRoot | Out-Null

  Assert-Throws -Name "invalid validator count" -Script {
    & (Join-Path $RepoRoot "scripts\localnet\init.ps1") `
      -OutputDir (Join-Path $negativeRoot "invalid-count") `
      -Binary $Binary `
      -ValidatorCount 0 `
      -SkipBuild
  }

  Assert-Throws -Name "missing binary with NoInit" -Script {
    & (Join-Path $RepoRoot "scripts\localnet\start.ps1") `
      -OutputDir (Join-Path $negativeRoot "missing-binary") `
      -Binary (Join-Path $negativeRoot "missing\aetrad.exe") `
      -NoInit
  }

  Assert-Throws -Name "height timeout" -Script {
    Wait-LocalnetHeight -TargetHeight 1 -RPCPort ($BaseRPCPort + 9000) -TimeoutSeconds 1 | Out-Null
  }

  $occupiedOutput = Join-Path $negativeRoot "occupied-port"
  $occupiedBaseP2P = $BaseP2PPort + 2000
  $occupiedBaseRPC = $BaseRPCPort + 2000
  $occupiedBaseREST = $BaseRESTPort + 200
  $occupiedBaseGRPC = $BaseGRPCPort + 200
  & (Join-Path $RepoRoot "scripts\localnet\init.ps1") `
    -OutputDir $occupiedOutput `
    -Binary $Binary `
    -ValidatorCount 1 `
    -BaseP2PPort $occupiedBaseP2P `
    -BaseRPCPort $occupiedBaseRPC `
    -BaseRESTPort $occupiedBaseREST `
    -BaseGRPCPort $occupiedBaseGRPC `
    -SkipBuild | Out-Null

  $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Any, $occupiedBaseRPC)
  $listener.ExclusiveAddressUse = $true
  $listener.Start()
  try {
    Assert-Throws -Name "occupied RPC port" -Script {
      & (Join-Path $RepoRoot "scripts\localnet\start.ps1") `
        -OutputDir $occupiedOutput `
        -Binary $Binary `
        -ValidatorCount 1 `
        -BaseP2PPort $occupiedBaseP2P `
        -BaseRPCPort $occupiedBaseRPC `
        -BaseRESTPort $occupiedBaseREST `
        -BaseGRPCPort $occupiedBaseGRPC `
        -NoInit
    }
  } finally {
    $listener.Stop()
    & (Join-Path $RepoRoot "scripts\localnet\reset.ps1") -OutputDir $occupiedOutput
  }

  Remove-LocalnetDirectory -OutputDir $negativeRoot
}

Push-Location $RepoRoot
try {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  $initExtra = @{}
  if ($SkipBuild) { $initExtra.SkipBuild = $true }
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1" -Extra $initExtra
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet reached height $height"

  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "validator set contains $ValidatorCount validators"

  if ($ValidatorCount -gt 1) {
    Wait-LocalnetPeers -ExpectedMinPeers ($ValidatorCount - 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "node0 has at least $($ValidatorCount - 1) peers"
  }

  if ($EnableAPI) {
    Wait-LocalnetRest -RESTPort $node0Ports.REST -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "REST endpoint is healthy on port $($node0Ports.REST)"
  }
  if ($EnableGRPC) {
    Wait-LocalnetGrpc -GRPCPort $node0Ports.GRPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "gRPC TCP endpoint is listening on port $($node0Ports.GRPC)"
  }

  $health = & .\scripts\localnet\health.ps1 `
    -OutputDir $OutputDir `
    -ValidatorCount $ValidatorCount `
    -BaseP2PPort $BaseP2PPort `
    -BaseRPCPort $BaseRPCPort `
    -BaseRESTPort $BaseRESTPort `
    -BaseGRPCPort $BaseGRPCPort `
    -BasePprofPort $BasePprofPort `
    -PortStride $PortStride `
    -EnableAPI $EnableAPI `
    -EnableGRPC $EnableGRPC `
    -EnableRPC $EnableRPC `
    -Json | ConvertFrom-Json
  if ($EnableRPC -and $health.rpc -ne "ok") { throw "health RPC check failed: $($health.rpc)" }
  if ($EnableAPI -and $health.rest -ne "ok") { throw "health REST check failed: $($health.rest)" }
  if ($EnableGRPC -and $health.grpc -ne "ok") { throw "health gRPC check failed: $($health.grpc)" }
  if (@($health.processes).Count -lt $ValidatorCount) { throw "health process list has fewer entries than validators" }
  Write-Host "health summary passed"

  Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "block", "--node", $rpcNode, "--output", "json") | Out-Null
  Write-Host "query block succeeded"

  if (-not $SkipBankTx -and $ValidatorCount -gt 1) {
    $node0Home = Join-Path $OutputDir "node0\aetrad"
    $node1Home = Join-Path $OutputDir "node1\aetrad"
    $recipient = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
    Send-LocalnetBankTx -Binary $Binary -FromHome $node0Home -FromKey "node0" -ToAddress $recipient -ChainId $ChainId -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    $height = Wait-LocalnetHeight -TargetHeight ([int64]$height + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
    Write-Host "bank send tx succeeded and chain reached height $height"
  }

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  $restartHeight = Wait-LocalnetHeight -TargetHeight ([int64]$height + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet restart preserved state and reached height $restartHeight"

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir

  if (-not $SkipNegativeChecks) {
    Test-LocalnetNegativeCases
  }
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
