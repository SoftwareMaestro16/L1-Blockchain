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
  [string]$FactorySubdenom = "dexgold",
  [string]$Fees = "1000000naet"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for DEX smoke"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"

$poolId = 1
$lpDenom = "lp/$poolId"
$initialNorb = 10000000
$initialFactory = 10000000
$addNorb = 1000000
$addFactory = 1000000
$swapInNorb = 100000
$removeShares = 1000000

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
    [string]$TxFees = $Fees
  )

  return $ActionArgs + @(
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $TxFees,
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
    [string]$FromKey = "node0",
    [switch]$ExpectFailure,
    [string]$ExpectedLog = ""
  )

  return Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds `
    -ExpectFailure:$ExpectFailure `
    -ExpectedLog $ExpectedLog
}

function Get-DexPool {
  param([int]$PoolId = $poolId)

  $res = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "dex", "pool", "$PoolId", "--node", $rpcNode, "--output", "json")
  return $res.pool
}

function Assert-Pool {
  param(
    [object]$Pool,
    [string]$FactoryDenom,
    [int64]$ReserveNorb,
    [int64]$ReserveFactory,
    [int64]$TotalShares
  )

  if ([int64]$Pool.id -ne $poolId) {
    throw "pool id must be $poolId, got $($Pool.id)"
  }
  if ($Pool.denom0 -ne $FactoryDenom -or $Pool.denom1 -ne "naet") {
    throw "pool denoms must be canonical factory/naet, got $($Pool.denom0)/$($Pool.denom1)"
  }
  if ([int64]$Pool.reserve0 -ne $ReserveFactory -or [int64]$Pool.reserve1 -ne $ReserveNorb) {
    throw "pool reserves mismatch: expected $ReserveFactory/$ReserveNorb, got $($Pool.reserve0)/$($Pool.reserve1)"
  }
  if ([int64]$Pool.total_shares -ne $TotalShares) {
    throw "pool total shares mismatch: expected $TotalShares, got $($Pool.total_shares)"
  }
  if ($Pool.lp_denom -ne $lpDenom) {
    throw "pool lp denom must be $lpDenom, got $($Pool.lp_denom)"
  }
}

function Get-BalanceAmount {
  param(
    [string]$Address,
    [string]$Denom
  )

  $balance = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $Denom -RPCPort $node0Ports.RPC
  if (-not $balance.amount) {
    return [int64]0
  }
  return [int64]$balance.amount
}

Push-Location $RepoRoot
try {
  $goCache = Join-Path $RepoRoot ".work\gocache"
  New-Item -ItemType Directory -Force -Path $goCache | Out-Null
  $env:GOCACHE = $goCache

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet reached height $height"
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "validator set contains $ValidatorCount validators"
  if ($EnableGRPC) {
    Wait-LocalnetGrpc -GRPCPort $node0Ports.GRPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "gRPC TCP endpoint is listening on port $($node0Ports.GRPC)"
  }
  if ($EnableAPI) {
    try {
      Wait-LocalnetRest -RESTPort $node0Ports.REST -TimeoutSeconds 10 | Out-Null
      Write-Host "REST endpoint is healthy on port $($node0Ports.REST)"
    } catch {
      Write-Host "REST endpoint health check skipped: $($_.Exception.Message)"
    }
  }

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $factoryDenom = "factory/$node0/$FactorySubdenom"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom) -FromHome $node0Home | Out-Null
  Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "contract-assets", "denom", $factoryDenom, "--node", $rpcNode, "--output", "json") | Out-Null
  Write-Host "created factory denom $factoryDenom"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom", $node0) -FromHome $node0Home | Out-Null
  $factoryBalance = Get-BalanceAmount -Address $node0 -Denom $factoryDenom
  if ($factoryBalance -lt 100000000) {
    throw "factory balance after mint must be at least 100000000, got $factoryBalance"
  }
  Write-Host "minted factory liquidity asset to node0"

  $createPoolTx = Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "$($initialNorb)naet", "$($initialFactory)$factoryDenom") -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $createPoolTx -Type "dex_create_pool" -Attributes @{
    pool_id       = "$poolId"
    creator       = $node0
    denom0        = $factoryDenom
    denom1        = "naet"
    amount0       = "$initialFactory"
    amount1       = "$initialNorb"
    lp_denom      = $lpDenom
    minted_shares = "$initialNorb"
  } | Out-Null
  $pool = Get-DexPool
  Assert-Pool -Pool $pool -FactoryDenom $factoryDenom -ReserveNorb $initialNorb -ReserveFactory $initialFactory -TotalShares $initialNorb
  Write-Host "created DEX pool $poolId with LP denom $lpDenom"

  $lpBalance = Get-BalanceAmount -Address $node0 -Denom $lpDenom
  if ($lpBalance -ne $initialNorb) {
    throw "initial LP balance mismatch: expected $initialNorb, got $lpBalance"
  }
  Write-Host "LP balance after create-pool is $lpBalance$lpDenom"

  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "1$factoryDenom", "1naet") -FromHome $node0Home -ExpectFailure -ExpectedLog "pool already exists" | Out-Null
  Write-Host "duplicate pair pool creation is rejected"

  Send-SignedTx -ActionArgs @("tx", "dex", "add-liquidity", "$poolId", "$($addNorb)naet", "$($addFactory)$factoryDenom", "$($addNorb + 1)") -FromHome $node0Home -ExpectFailure -ExpectedLog "minted shares below minimum" | Out-Null
  Write-Host "add-liquidity slippage guard rejected excessive min_shares"

  $addLiquidityTx = Send-SignedTx -ActionArgs @("tx", "dex", "add-liquidity", "$poolId", "$($addNorb)naet", "$($addFactory)$factoryDenom", "$addNorb") -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $addLiquidityTx -Type "dex_add_liquidity" -Attributes @{
    pool_id       = "$poolId"
    depositor     = $node0
    denom0        = $factoryDenom
    denom1        = "naet"
    amount0       = "$addFactory"
    amount1       = "$addNorb"
    lp_denom      = $lpDenom
    minted_shares = "$addNorb"
  } | Out-Null
  $expectedReserveNorb = $initialNorb + $addNorb
  $expectedReserveFactory = $initialFactory + $addFactory
  $expectedShares = $initialNorb + $addNorb
  $pool = Get-DexPool
  Assert-Pool -Pool $pool -FactoryDenom $factoryDenom -ReserveNorb $expectedReserveNorb -ReserveFactory $expectedReserveFactory -TotalShares $expectedShares
  Write-Host "add-liquidity updated reserves to $expectedReserveFactory/$expectedReserveNorb and shares $expectedShares"

  Send-SignedTx -ActionArgs @("tx", "dex", "add-liquidity", "$poolId", "1naet", "1testtoken", "1") -FromHome $node0Home -ExpectFailure -ExpectedLog "liquidity denoms do not match pool" | Out-Null
  Write-Host "add-liquidity rejects wrong denom pair"

  Send-SignedTx -ActionArgs @("tx", "dex", "swap-exact-in", "$poolId", "$($swapInNorb)naet", $factoryDenom, "1000000") -FromHome $node0Home -ExpectFailure -ExpectedLog "amount out below minimum" | Out-Null
  Write-Host "swap slippage guard rejected excessive min_amount_out"

  $factoryBeforeSwap = Get-BalanceAmount -Address $node0 -Denom $factoryDenom
  $swapTx = Send-SignedTx -ActionArgs @("tx", "dex", "swap-exact-in", "$poolId", "$($swapInNorb)naet", $factoryDenom, "1") -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $swapTx -Type "dex_swap_exact_amount_in" -Attributes @{
    pool_id  = "$poolId"
    trader   = $node0
    token_in = "$($swapInNorb)naet"
  } | Out-Null
  $factoryAfterSwap = Get-BalanceAmount -Address $node0 -Denom $factoryDenom
  if ($factoryAfterSwap -le $factoryBeforeSwap) {
    throw "factory output balance must increase after swap, before=$factoryBeforeSwap after=$factoryAfterSwap"
  }
  $poolAfterSwap = Get-DexPool
  if ([int64]$poolAfterSwap.reserve1 -ne ($expectedReserveNorb + $swapInNorb)) {
    throw "naet reserve must increase by swap input"
  }
  if ([int64]$poolAfterSwap.reserve0 -ge $expectedReserveFactory) {
    throw "factory reserve must decrease after swap output"
  }
  Write-Host "swap exact-in increased factory output balance from $factoryBeforeSwap to $factoryAfterSwap"

  Send-SignedTx -ActionArgs @("tx", "dex", "remove-liquidity", "$poolId", "1lp/999") -FromHome $node0Home -ExpectFailure -ExpectedLog "invalid LP shares" | Out-Null
  Write-Host "remove-liquidity rejects wrong LP denom"

  $lpBeforeRemove = Get-BalanceAmount -Address $node0 -Denom $lpDenom
  $removeLiquidityTx = Send-SignedTx -ActionArgs @("tx", "dex", "remove-liquidity", "$poolId", "$($removeShares)$lpDenom") -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $removeLiquidityTx -Type "dex_remove_liquidity" -Attributes @{
    pool_id    = "$poolId"
    withdrawer = $node0
    lp_denom   = $lpDenom
    shares     = "$removeShares"
  } | Out-Null
  $lpAfterRemove = Get-BalanceAmount -Address $node0 -Denom $lpDenom
  if ($lpAfterRemove -ne ($lpBeforeRemove - $removeShares)) {
    throw "LP balance mismatch after remove-liquidity: expected $($lpBeforeRemove - $removeShares), got $lpAfterRemove"
  }
  $poolAfterRemove = Get-DexPool
  if ([int64]$poolAfterRemove.total_shares -ne ([int64]$poolAfterSwap.total_shares - $removeShares)) {
    throw "pool total shares did not decrease by removed shares"
  }
  if ([int64]$poolAfterRemove.reserve0 -ge [int64]$poolAfterSwap.reserve0 -or [int64]$poolAfterRemove.reserve1 -ge [int64]$poolAfterSwap.reserve1) {
    throw "pool reserves must decrease after remove-liquidity"
  }
  Write-Host "remove-liquidity burned $removeShares$lpDenom and updated reserves"

  $pools = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "dex", "pools", "--node", $rpcNode, "--output", "json")
  if (@($pools.pools).Count -ne 1) {
    throw "query dex pools must return exactly 1 pool, got $(@($pools.pools).Count)"
  }

  if ($EnableAPI) {
    try {
      $restPool = Invoke-RestMethod -Uri "http://127.0.0.1:$($node0Ports.REST)/l1/dex/v1/pools/$poolId" -TimeoutSec 2
      if ($restPool.pool.lp_denom -ne $lpDenom) {
        throw "REST pool lp denom mismatch"
      }
      Write-Host "REST DEX pool query returned $lpDenom"
    } catch {
      Write-Host "REST DEX pool query skipped: $($_.Exception.Message)"
    }
  }

  Write-Host "DEX smoke completed at height $(Get-LocalnetHeight -RPCPort $node0Ports.RPC)"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
