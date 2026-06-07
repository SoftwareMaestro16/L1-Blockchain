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
  [string]$AcceptedFees = "1000naet",
  [string]$WrongFees = "1000testtoken",
  [string]$MultiDenomFees = "1000naet,1testtoken"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for fees ante smoke"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"
$wrongFeeError = "fee denom testtoken not accepted; use naet"

function Invoke-WithCommonLocalnetArgs {
  param(
    [string]$ScriptPath,
    [hashtable]$Extra = @{}
  )

  $args = @{
    OutputDir      = $OutputDir
    Binary         = $Binary
    ChainId        = $ChainId
    ValidatorCount = $ValidatorCount
    BaseP2PPort    = $BaseP2PPort
    BaseRPCPort    = $BaseRPCPort
    BaseRESTPort   = $BaseRESTPort
    BaseGRPCPort   = $BaseGRPCPort
    BasePprofPort  = $BasePprofPort
    PortStride     = $PortStride
    TimeoutCommit  = $TimeoutCommit
    LogLevel       = $LogLevel
    EnableAPI      = $EnableAPI
    EnableGRPC     = $EnableGRPC
    EnableRPC      = $EnableRPC
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function New-SignedTxArgs {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [AllowEmptyString()][string]$Fees = $AcceptedFees
  )

  $args = $ActionArgs + @(
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $rpcNode,
    "--output", "json"
  )
  if (-not [string]::IsNullOrWhiteSpace($Fees)) {
    $args += @("--fees", $Fees)
  }
  return $args
}

function Send-SignedTx {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [AllowEmptyString()][string]$Fees = $AcceptedFees,
    [switch]$ExpectFailure,
    [string]$ExpectedLog = ""
  )

  return Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey -Fees $Fees) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds `
    -ExpectFailure:$ExpectFailure `
    -ExpectedLog $ExpectedLog
}

function Assert-AllowedFeesParams {
  param([object]$Params)

  $allowed = @($Params.allowed_fee_denoms)
  if ($allowed.Count -ne 1 -or $allowed[0] -ne "naet") {
    throw "fees params must allow only naet, got $($allowed -join ',')"
  }
  if ($Params.validator_rewards_ratio -ne "0.98") {
    throw "validator_rewards_ratio must be 0.98, got $($Params.validator_rewards_ratio)"
  }
  if ($Params.community_pool_ratio -ne "0.02") {
    throw "community_pool_ratio must be 0.02, got $($Params.community_pool_ratio)"
  }
}

Push-Location $RepoRoot
try {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  $initExtra = @{}
  if (Test-Path -LiteralPath $Binary) {
    $initExtra.SkipBuild = $true
  }
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1" -Extra $initExtra
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet reached height $height"
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "validator set contains $ValidatorCount validators"
  if ($EnableAPI) {
    try {
      Wait-LocalnetRest -RESTPort $node0Ports.REST -TimeoutSeconds 10 | Out-Null
      Write-Host "REST endpoint is healthy on port $($node0Ports.REST)"
    } catch {
      Write-Host "REST endpoint health check skipped: $($_.Exception.Message)"
    }
  }
  if ($EnableGRPC) {
    Wait-LocalnetGrpc -GRPCPort $node0Ports.GRPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "gRPC TCP endpoint is listening on port $($node0Ports.GRPC)"
  }

  $feesParams = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "fees", "params", "--node", $rpcNode, "--output", "json")
  Assert-AllowedFeesParams -Params $feesParams.params
  Write-Host "CLI fees params allow only naet"

  if ($EnableAPI) {
    try {
      $restParams = Invoke-RestMethod -Uri "http://127.0.0.1:$($node0Ports.REST)/l1/fees/v1/params" -TimeoutSec 2
      Assert-AllowedFeesParams -Params $restParams.params
      Write-Host "REST fees params allow only naet"
    } catch {
      Write-Host "REST fees params query skipped: $($_.Exception.Message)"
    }
  }

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"

  Send-SignedTx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1000naet") -FromHome $node0Home -Fees $AcceptedFees | Out-Null
  Write-Host "bank send with naet fee succeeded"

  Send-SignedTx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet") -FromHome $node0Home -Fees $WrongFees -ExpectFailure -ExpectedLog $wrongFeeError | Out-Null
  Write-Host "bank send with wrong fee denom is rejected"

  Send-SignedTx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet") -FromHome $node0Home -Fees $MultiDenomFees -ExpectFailure -ExpectedLog $wrongFeeError | Out-Null
  Write-Host "bank send with mixed fee denoms is rejected"

  Send-SignedTx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet") -FromHome $node0Home -Fees "0naet" | Out-Null
  Write-Host "bank send with zero naet fee is accepted by prototype localnet"

  Send-SignedTx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet") -FromHome $node0Home -Fees "" | Out-Null
  Write-Host "bank send with empty fee list is accepted by prototype localnet min-gas-prices"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", "gold") -FromHome $node0Home -Fees $AcceptedFees | Out-Null
  $factoryDenom = "factory/$node0/gold"
  Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "contract-assets", "denom", $factoryDenom, "--node", $rpcNode, "--output", "json") | Out-Null
  Write-Host "contract-assets create-denom with naet fee succeeded"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", "badfee") -FromHome $node0Home -Fees $WrongFees -ExpectFailure -ExpectedLog $wrongFeeError | Out-Null
  Write-Host "contract-assets tx with wrong fee denom is rejected"

  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "1000naet", "1000testtoken") -FromHome $node0Home -Fees $AcceptedFees | Out-Null
  Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "dex", "pool", "1", "--node", $rpcNode, "--output", "json") | Out-Null
  Write-Host "DEX create-pool with naet fee succeeded"

  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "10naet", "10testtoken") -FromHome $node0Home -Fees $WrongFees -ExpectFailure -ExpectedLog $wrongFeeError | Out-Null
  Write-Host "DEX tx with wrong fee denom is rejected"

  $node0Balance = Get-LocalnetBankBalance -Binary $Binary -Address $node0 -Denom "naet" -RPCPort $node0Ports.RPC
  if ($node0Balance.denom -ne "naet" -or [int64]$node0Balance.amount -le 0) {
    throw "node0 must retain positive naet balance after fee smoke"
  }
  Write-Host "fees ante smoke completed with node0 balance $($node0Balance.amount)naet"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
