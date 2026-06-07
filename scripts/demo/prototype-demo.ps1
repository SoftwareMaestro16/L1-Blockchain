param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 4,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [string]$Fees = "1000000naet",
  [string]$FactorySubdenom = "demogold",
  [switch]$SkipBuild,
  [switch]$Check,
  [switch]$KeepLocalnet
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for the prototype demo"
}
if ($ChainId -notmatch 'local') {
  throw "prototype-demo is local-only; ChainId must contain 'local'"
}
if ($Fees -notmatch '^[0-9]+naet$') {
  throw "demo fees must use the local prototype fee denom naet"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-demo"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "demo localnet output directory"
if (-not $SkipBuild) {
  Assert-LocalnetWorkspacePath -Path (Split-Path $Binary) -Purpose "demo binary output directory"
}

$node0Ports = Get-LocalnetPortProfile `
  -Index 0 `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BasePprofPort $BasePprofPort `
  -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"
$restBase = "http://127.0.0.1:$($node0Ports.REST)"

function Write-DemoStep {
  param([string]$Message)

  Write-Host ""
  Write-Host "==> $Message"
}

function Write-DemoNote {
  param([string]$Message)

  Write-Host "    $Message"
}

function Invoke-DemoLocalnetScript {
  param(
    [string]$ScriptName,
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
  & (Join-Path $RepoRoot "scripts\localnet\$ScriptName") @args
}

function Invoke-DemoQueryCliJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--node", $rpcNode, "--output", "json"))
}

function New-DemoTxArgs {
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

function Send-DemoTx {
  param(
    [string]$Label,
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0"
  )

  Write-DemoNote "$Label"
  $tx = Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-DemoTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds
  Write-DemoNote "txhash=$(Get-LocalnetTxHash -Tx $tx)"
  return $tx
}

function Get-DemoBalanceAmount {
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

function Invoke-DemoCheck {
  Write-Output "Aetra prototype demo check"
  Write-Output "local-only: true"
  Write-Output "chain-id: $ChainId"
  Write-Output "validators: $ValidatorCount"
  Write-Output "output-dir: $OutputDir"
  Write-Output "binary: $Binary"
  Write-Output "rpc: $rpcNode"
  Write-Output "rest: $restBase"
  Write-Output "steps:"
  foreach ($step in @(
      "build aetrad unless -SkipBuild",
      "stop/init/validate/start 3-validator localnet",
      "show height and REST node info",
      "send bank tx in naet",
      "create and mint contract-assets denom",
      "create DEX pool and swap naet for factory token",
      "show REST DEX pool and final balances",
      "stop localnet unless -KeepLocalnet"
    )) {
    Write-Output " - $step"
  }
}

if ($Check) {
  Invoke-DemoCheck
  return
}

$node0Home = $null
$node1Home = $null

Push-Location $RepoRoot
try {
  Write-Host "Aetra prototype guided demo"
  Write-Host "LOCAL ONLY: uses ignored localnet homes and --keyring-backend test. Do not use these keys on public networks."
  Write-Host "Demo wraps tested localnet commands; it is not a substitute for e2e tests."

  Write-DemoStep "Build binary"
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  if (-not $SkipBuild) {
    & .\scripts\build-aetrad.ps1 -Binary $Binary
  } elseif (!(Test-Path -LiteralPath $Binary)) {
    throw "Binary not found at $Binary and -SkipBuild was specified"
  }
  $version = & $Binary version 2>&1
  Write-DemoNote "aetrad version: $($version | Select-Object -First 1)"

  Write-DemoStep "Start localnet"
  Invoke-DemoLocalnetScript -ScriptName "init.ps1" -Extra @{ SkipBuild = $true }
  Invoke-DemoLocalnetScript -ScriptName "validate-genesis.ps1"
  Invoke-DemoLocalnetScript -ScriptName "start.ps1" -Extra @{ NoInit = $true }
  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-DemoNote "height=$height rpc=$rpcNode rest=$restBase"

  Write-DemoStep "Show block and REST node info"
  $block = Invoke-DemoQueryCliJson -Arguments @("query", "block")
  $blockHeight = if ($block.block.header.height) { $block.block.header.height } else { $block.header.height }
  Write-DemoNote "CLI latest block height=$blockHeight"
  $restNode = Invoke-RestMethod -Uri "$restBase/cosmos/base/tendermint/v1beta1/node_info" -TimeoutSec 5
  Write-DemoNote "REST network=$($restNode.default_node_info.network)"

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
  Write-DemoNote "node0=$node0"
  Write-DemoNote "node1=$node1"

  Write-DemoStep "Bank send"
  $node1Before = Get-DemoBalanceAmount -Address $node1 -Denom "naet"
  Send-DemoTx -Label "send 1000naet from node0 to node1" -ActionArgs @("tx", "bank", "send", "node0", $node1, "1000naet") -FromHome $node0Home | Out-Null
  $node1After = Get-DemoBalanceAmount -Address $node1 -Denom "naet"
  if ($node1After -ne ($node1Before + 1000)) {
    throw "bank send balance mismatch: before=$node1Before after=$node1After"
  }
  Write-DemoNote "node1 naet balance: $node1Before -> $node1After"

  Write-DemoStep "Contract assets create and mint"
  Send-DemoTx -Label "create factory denom $FactorySubdenom" -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom) -FromHome $node0Home | Out-Null
  $factoryDenom = "factory/$node0/$FactorySubdenom"
  $tfQuery = Invoke-DemoQueryCliJson -Arguments @("query", "contract-assets", "denom", $factoryDenom)
  if ($tfQuery.metadata.admin -ne $node0) {
    throw "contract-assets admin mismatch"
  }
  Send-DemoTx -Label "mint 100000000$factoryDenom to node0" -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom", $node0) -FromHome $node0Home | Out-Null
  $factoryBalance = Get-DemoBalanceAmount -Address $node0 -Denom $factoryDenom
  Write-DemoNote "factory denom=$factoryDenom balance=$factoryBalance"

  Write-DemoStep "DEX pool and swap"
  Send-DemoTx -Label "create DEX pool with naet and factory token" -ActionArgs @("tx", "dex", "create-pool", "10000000naet", "10000000$factoryDenom") -FromHome $node0Home | Out-Null
  $pool = Invoke-DemoQueryCliJson -Arguments @("query", "dex", "pool", "1")
  if ($pool.pool.lp_denom -ne "lp/1") {
    throw "DEX pool lp denom mismatch"
  }
  $factoryBeforeSwap = Get-DemoBalanceAmount -Address $node0 -Denom $factoryDenom
  Send-DemoTx -Label "swap 100000naet for factory token" -ActionArgs @("tx", "dex", "swap-exact-in", "1", "100000naet", $factoryDenom, "1") -FromHome $node0Home | Out-Null
  $factoryAfterSwap = Get-DemoBalanceAmount -Address $node0 -Denom $factoryDenom
  if ($factoryAfterSwap -le $factoryBeforeSwap) {
    throw "DEX swap did not increase factory balance"
  }
  Write-DemoNote "factory balance after swap: $factoryBeforeSwap -> $factoryAfterSwap"

  Write-DemoStep "REST query and final balances"
  $restPool = Invoke-RestMethod -Uri "$restBase/l1/dex/v1/pools/1" -TimeoutSec 5
  Write-DemoNote "REST pool 1 lp=$($restPool.pool.lp_denom) reserves=$($restPool.pool.reserve0)/$($restPool.pool.reserve1)"
  $node0Norb = Get-DemoBalanceAmount -Address $node0 -Denom "naet"
  $node1Norb = Get-DemoBalanceAmount -Address $node1 -Denom "naet"
  $node0Lp = Get-DemoBalanceAmount -Address $node0 -Denom "lp/1"
  Write-DemoNote "node0 naet=$node0Norb"
  Write-DemoNote "node1 naet=$node1Norb"
  Write-DemoNote "node0 $factoryDenom=$factoryAfterSwap"
  Write-DemoNote "node0 lp/1=$node0Lp"

  Write-DemoStep "Demo complete"
  Write-DemoNote "Aetra local prototype produced blocks, accepted bank/contract-assets/DEX txs, served REST queries, and updated final state."
} finally {
  if (-not $KeepLocalnet) {
    Write-DemoStep "Stop localnet"
    & (Join-Path $RepoRoot "scripts\localnet\stop.ps1") -OutputDir $OutputDir
  } else {
    Write-DemoStep "Localnet kept running"
    Write-DemoNote "Stop it with: .\scripts\localnet\stop.ps1 -OutputDir $OutputDir"
  }
  Pop-Location
}
