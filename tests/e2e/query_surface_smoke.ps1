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
  [string]$FactorySubdenom = "querygold",
  [string]$Fees = "1000000naet"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 1) {
  throw "ValidatorCount must be at least 1"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"
$grpcAddr = "127.0.0.1:$($node0Ports.GRPC)"
$restBase = "http://127.0.0.1:$($node0Ports.REST)"

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
    EnableAPI      = $true
    EnableGRPC     = $true
    EnableRPC      = $true
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function Invoke-RestJson {
  param([string]$Path)

  try {
    return Invoke-RestMethod -Uri "$restBase$Path" -TimeoutSec 5
  } catch {
    throw "REST $Path failed: $($_.Exception.Message)"
  }
}

function Assert-RestError {
  param(
    [string]$Path,
    [int]$ExpectedStatus
  )

  try {
    Invoke-RestJson -Path $Path | Out-Null
  } catch {
    $actual = $null
    if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
      if ($null -ne $_.Exception.Response.StatusCode.value__) {
        $actual = [int]$_.Exception.Response.StatusCode.value__
      } else {
        $actual = [int]$_.Exception.Response.StatusCode
      }
    }
    if ($null -eq $actual -and $_.Exception.Message -match "\((\d{3})\)") {
      $actual = [int]$Matches[1]
    }
    if ($actual -eq $ExpectedStatus) {
      return
    }
    throw "REST $Path returned status $actual, expected $ExpectedStatus"
  }
  throw "REST $Path succeeded but status $ExpectedStatus was expected"
}

function Test-HasNextKey {
  param($Response)
  return ($Response.pagination.next_key -or $Response.pagination.nextKey)
}

function New-SignedTxArgs {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0"
  )

  return $ActionArgs + @(
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $rpcNode,
    "--output", "json"
  )
}

function Send-SignedTx {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0"
  )

  return Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds
}

function Invoke-QueryCliJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--node", $rpcNode, "--output", "json"))
}

function Invoke-QueryGrpcJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--grpc-addr", $grpcAddr, "--grpc-insecure", "--node", $rpcNode, "--output", "json"))
}

function Assert-QueryGrpcFailure {
  param(
    [string[]]$Arguments,
    [string]$ExpectedText
  )

  $fullArgs = $Arguments + @("--grpc-addr", $grpcAddr, "--grpc-insecure", "--node", $rpcNode, "--output", "json")
  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @fullArgs 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  if ($exitCode -eq 0) {
    throw "gRPC query succeeded but failure was expected: $Binary $($fullArgs -join ' ')"
  }
  $text = $output -join "`n"
  if ($ExpectedText -and ($text -notmatch [regex]::Escape($ExpectedText))) {
    throw "gRPC query failed, but output did not contain '$ExpectedText': $text"
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
  & .\scripts\localnet\health.ps1 -OutputDir $OutputDir -ValidatorCount $ValidatorCount -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "localnet health check passed for RPC/REST/gRPC"

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"

  $latestBlock = Invoke-QueryCliJson -Arguments @("query", "block")
  if (-not ($latestBlock.header.height -or $latestBlock.block.header.height)) {
    throw "CLI query block must include header.height"
  }

  $balance = Invoke-QueryCliJson -Arguments @("query", "bank", "balance", $node0, "naet")
  if ($balance.balance.denom -ne "naet" -or [int64]$balance.balance.amount -le 0) {
    throw "CLI bank balance must return positive naet balance"
  }
  $balanceGrpc = Invoke-QueryGrpcJson -Arguments @("query", "bank", "balance", $node0, "naet")
  if ($balanceGrpc.balance.denom -ne "naet" -or [int64]$balanceGrpc.balance.amount -le 0) {
    throw "gRPC bank balance must return positive naet balance"
  }

  $validators = Invoke-QueryCliJson -Arguments @("query", "staking", "validators")
  if (@($validators.validators).Count -ne $ValidatorCount) {
    throw "CLI staking validators count mismatch"
  }
  $validatorsGrpc = Invoke-QueryGrpcJson -Arguments @("query", "staking", "validators")
  if (@($validatorsGrpc.validators).Count -ne $ValidatorCount) {
    throw "gRPC staking validators count mismatch"
  }

  $feesParams = Invoke-QueryGrpcJson -Arguments @("query", "fees", "params")
  if (@($feesParams.params.allowed_fee_denoms) -notcontains "naet") {
    throw "gRPC fees params must include naet"
  }
  Write-Host "CLI/gRPC queries returned block, bank, staking, and fees data"

  $emptyDenoms = Invoke-QueryGrpcJson -Arguments @("query", "contract-assets", "denoms")
  if (@($emptyDenoms.denoms).Count -ne 0) {
    throw "fresh localnet contract-assets denoms must be empty"
  }
  if (-not $emptyDenoms.pagination) {
    throw "fresh localnet contract-assets denoms must include pagination"
  }

  $emptyPools = Invoke-QueryGrpcJson -Arguments @("query", "dex", "pools")
  if (@($emptyPools.pools).Count -ne 0) {
    throw "fresh localnet dex pools must be empty"
  }
  if (-not $emptyPools.pagination) {
    throw "fresh localnet dex pools must include pagination"
  }

  $restBlock = Invoke-RestJson -Path "/cosmos/base/tendermint/v1beta1/blocks/latest"
  if (-not $restBlock.block.header.height) {
    throw "REST latest block must include block.header.height"
  }
  $restNode = Invoke-RestJson -Path "/cosmos/base/tendermint/v1beta1/node_info"
  if ($restNode.default_node_info.network -ne $ChainId) {
    throw "REST node_info network mismatch"
  }
  $restBalances = Invoke-RestJson -Path "/cosmos/bank/v1beta1/balances/$node0"
  $restNorb = @($restBalances.balances) | Where-Object { $_.denom -eq "naet" } | Select-Object -First 1
  if (-not $restNorb -or [int64]$restNorb.amount -le 0) {
    throw "REST bank balances must include positive naet"
  }
  $restBalanceByDenom = Invoke-RestJson -Path "/cosmos/bank/v1beta1/balances/$node0/by_denom?denom=naet"
  if ($restBalanceByDenom.balance.denom -ne "naet" -or [int64]$restBalanceByDenom.balance.amount -le 0) {
    throw "REST bank balance by denom must include positive naet"
  }
  $restValidators = Invoke-RestJson -Path "/cosmos/staking/v1beta1/validators"
  if (@($restValidators.validators).Count -ne $ValidatorCount) {
    throw "REST staking validators count mismatch"
  }
  $restFees = Invoke-RestJson -Path "/l1/fees/v1/params"
  if (@($restFees.params.allowed_fee_denoms) -notcontains "naet") {
    throw "REST fees params must include naet"
  }
  $restDenoms = Invoke-RestJson -Path "/l1/contract-assets/v1/denoms"
  if (@($restDenoms.denoms).Count -ne 0) {
    throw "REST contract-assets denoms must be empty on fresh localnet"
  }
  if (-not $restDenoms.pagination) {
    throw "REST contract-assets denoms must include pagination"
  }
  $restPools = Invoke-RestJson -Path "/l1/dex/v1/pools"
  if (@($restPools.pools).Count -ne 0) {
    throw "REST dex pools must be empty on fresh localnet"
  }
  if (-not $restPools.pagination) {
    throw "REST dex pools must include pagination"
  }
  Write-Host "REST base, bank, staking, fees, contract-assets, and dex list queries passed"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom) -FromHome $node0Home | Out-Null
  $factoryDenom = "factory/$node0/$FactorySubdenom"
  $factorySubdenom2 = "$FactorySubdenom-page"
  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", $factorySubdenom2) -FromHome $node0Home | Out-Null
  $factoryDenom2 = "factory/$node0/$factorySubdenom2"

  $tfPageCli = Invoke-QueryGrpcJson -Arguments @("query", "contract-assets", "denoms", "--limit", "1")
  if (@($tfPageCli.denoms).Count -ne 1 -or -not (Test-HasNextKey -Response $tfPageCli)) {
    throw "gRPC contract-assets denoms --limit 1 must return one denom and next_key"
  }
  $tfPageRest = Invoke-RestJson -Path "/l1/contract-assets/v1/denoms?pagination.limit=1"
  if (@($tfPageRest.denoms).Count -ne 1 -or -not (Test-HasNextKey -Response $tfPageRest)) {
    throw "REST contract-assets denoms pagination.limit=1 must return one denom and next_key"
  }

  $tfCli = Invoke-QueryGrpcJson -Arguments @("query", "contract-assets", "denom", $factoryDenom)
  if ($tfCli.metadata.admin -ne $node0) {
    throw "contract-assets denom admin mismatch"
  }

  $tfRest = Invoke-RestJson -Path "/l1/contract-assets/v1/denom/$factoryDenom"
  if ($tfRest.metadata.admin -ne $node0) {
    throw "REST contract-assets denom admin mismatch"
  }
  Assert-RestError -Path "/l1/contract-assets/v1/denom/factory/$node0/missing" -ExpectedStatus 404
  Write-Host "contract-assets gRPC/REST denom queries passed"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom", $node0) -FromHome $node0Home | Out-Null
  Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom2", $node0) -FromHome $node0Home | Out-Null
  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "10000000naet", "10000000$factoryDenom") -FromHome $node0Home | Out-Null
  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "10000000naet", "10000000$factoryDenom2") -FromHome $node0Home | Out-Null

  $poolCli = Invoke-QueryGrpcJson -Arguments @("query", "dex", "pool", "1")
  if ($poolCli.pool.id -ne "1" -and [int64]$poolCli.pool.id -ne 1) {
    throw "DEX pool query must return pool id 1"
  }
  $poolRest = Invoke-RestJson -Path "/l1/dex/v1/pools/1"
  if ($poolRest.pool.lp_denom -ne "lp/1") {
    throw "REST DEX pool must return lp/1"
  }
  $poolsCli = Invoke-QueryGrpcJson -Arguments @("query", "dex", "pools", "--limit", "1")
  if (@($poolsCli.pools).Count -ne 1 -or -not (Test-HasNextKey -Response $poolsCli)) {
    throw "gRPC DEX pools --limit 1 must return one pool and next_key"
  }
  $poolsRest = Invoke-RestJson -Path "/l1/dex/v1/pools"
  if (@($poolsRest.pools).Count -ne 2) {
    throw "REST DEX pools must return two pools after create-pool"
  }
  $poolsPageRest = Invoke-RestJson -Path "/l1/dex/v1/pools?pagination.limit=1"
  if (@($poolsPageRest.pools).Count -ne 1 -or -not (Test-HasNextKey -Response $poolsPageRest)) {
    throw "REST DEX pools pagination.limit=1 must return one pool and next_key"
  }
  Assert-RestError -Path "/l1/dex/v1/pools/0" -ExpectedStatus 400
  Assert-RestError -Path "/l1/dex/v1/pools/999" -ExpectedStatus 404
  Assert-QueryGrpcFailure -Arguments @("query", "dex", "pool", "999") -ExpectedText "pool not found"
  Write-Host "DEX gRPC/REST pool queries passed"

  Write-Host "query surface smoke completed at height $(Get-LocalnetHeight -RPCPort $node0Ports.RPC)"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
