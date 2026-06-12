param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$RPCPort = 26657,
  [string]$Fees = "1000000naet",
  [int]$TimeoutSeconds = 60,
  [string]$FromKey = "node0",
  [string]$Action = "status",
  [string]$Amount = "5000000naet",
  [string]$ValidatorAddress = "",
  [string]$PoolID = "",
  [switch]$SkipPool,
  [switch]$Json,
  [switch]$Check
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$nodeHome = Join-Path $OutputDir "node0\aetrad"
$rpcNode = "tcp://127.0.0.1:$RPCPort"

function Invoke-Cli {
  param([string[]]$Arguments)
  return Invoke-LocalnetCliJson -Binary $Binary -Arguments $Arguments
}

function Invoke-CliRaw {
  param([string[]]$Arguments)
  $prev = $ErrorActionPreference; $ErrorActionPreference = "Continue"
  $output = & $Binary @Arguments 2>&1; $ErrorActionPreference = $prev
  return ($output -join "`n")
}

function Send-Tx {
  param([string]$Label, [string[]]$ActionArgs)
  Write-Host "  $Label"
  $tx = Send-LocalnetTx -Binary $Binary -Arguments ($ActionArgs + @("--from", $FromKey, "--home", $nodeHome, "--chain-id", $ChainId, "--keyring-backend", "test", "--fees", $Fees, "--yes", "--broadcast-mode", "sync", "--node", $rpcNode, "--output", "json")) -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds
  Write-Host "  txhash=$(Get-LocalnetTxHash -Tx $tx)"
  return $tx
}

function Get-Balance {
  param([string]$Address, [string]$Denom = "naet")
  $bal = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $Denom -RPCPort $RPCPort
  if (-not $bal.amount) { return [int64]0 }
  return [int64]$bal.amount
}

function Assert-Running {
  try {
    Wait-LocalnetRpc -RPCPort $RPCPort -TimeoutSeconds 5 | Out-Null
  } catch { throw "Localnet not running on RPC $RPCPort" }
}

function UserAddr {
  param([string]$KeyName)
  return (Invoke-CliRaw -Arguments @("keys", "show", $KeyName, "--home", $nodeHome, "--keyring-backend", "test", "-a")).Trim().Split("`n")[-1].Trim()
}

if ($Check) {
  Write-Output "Staking & pool-staking helper"
  Write-Output "actions: status, delegate, unbond, pool-create, pool-deposit, pool-unbond, pool-rewards, pool-query"
  Write-Output "rpc: $rpcNode"
  return
}

Assert-Running

$userAddr = UserAddr -KeyName $FromKey
$valAddr = if ($ValidatorAddress) { $ValidatorAddress } else {
  $bonded = Get-LocalnetBondedValidator -Binary $Binary -RPCPort $RPCPort
  $bonded.operator_address
}

switch ($Action.ToLower()) {
  "status" {
    Write-Host "==> Staking status"
    $params = Get-LocalnetStakingParams -Binary $Binary -RPCPort $RPCPort
    Write-Host "  bond_denom=$($params.bond_denom) max_validators=$($params.max_validators)"
    $vals = Get-LocalnetStakingValidators -Binary $Binary -RPCPort $RPCPort
    Write-Host "  validators: $(@($vals).Count) total"
    foreach ($v in $vals) {
      $bondedAmt = if ($v.tokens) { $v.tokens } else { $v.delegator_shares }
      Write-Host "    $($v.operator_address) moniker=$($v.description.moniker) tokens=$bondedAmt status=$($v.status)"
    }
    try {
      $poolQuery = Invoke-Cli -Arguments @("query", "nominator-pool", "pools", "--node", $rpcNode, "--output", "json")
      Write-Host "  nominator pools: $(@($poolQuery.pools).Count)"
    } catch { Write-Host "  nominator pools: query unavailable" }
  }

  "delegate" {
    Write-Host "==> Delegate $Amount to $valAddr"
    $tx = Send-Tx -Label "delegate $Amount" -ActionArgs @("tx", "staking", "delegate", $valAddr, $Amount)
    try {
      $del = Invoke-Cli -Arguments @("query", "staking", "delegation", $userAddr, $valAddr, "--node", $rpcNode, "--output", "json")
      Write-Host "  delegation: $($del.delegation_response.balance.amount) $($del.delegation_response.balance.denom)"
    } catch { Write-Host "  delegation query failed" }
  }

  "unbond" {
    Write-Host "==> Unbond $Amount from $valAddr"
    Send-Tx -Label "unbond $Amount" -ActionArgs @("tx", "staking", "unbond", $valAddr, $Amount)
  }

  "pool-create" {
    Write-Host "==> Create official liquid staking pool"
    try {
      $tx = Send-Tx -Label "create official pool" -ActionArgs @("tx", "nominator-pool", "create-official-pool")
    } catch { Write-Host "  create-official-pool note: $_" }
  }

  "pool-deposit" {
    if ([string]::IsNullOrWhiteSpace($PoolID)) { throw "provide -PoolID for pool-deposit action" }
    Write-Host "==> Deposit $Amount to pool $PoolID"
    try {
      $tx = Send-Tx -Label "deposit $Amount to pool $PoolID" -ActionArgs @("tx", "nominator-pool", "deposit", $PoolID, $Amount)
    } catch { Write-Host "  deposit note: $_" }
  }

  "pool-unbond" {
    if ([string]::IsNullOrWhiteSpace($PoolID)) { throw "provide -PoolID for pool-unbond action" }
    Write-Host "==> Request unbond $Amount from pool $PoolID"
    Send-Tx -Label "request unbond $Amount from pool $PoolID" -ActionArgs @("tx", "nominator-pool", "request-unbond")
  }

  "pool-rewards" {
    Write-Host "==> Claim pool rewards"
    try {
      Send-Tx -Label "claim staking rewards" -ActionArgs @("tx", "nominator-pool", "claim-staking-rewards")
    } catch { Write-Host "  claim staking rewards note: $_" }
    try {
      Send-Tx -Label "claim pool rewards" -ActionArgs @("tx", "nominator-pool", "claim-rewards")
    } catch { Write-Host "  claim pool rewards note: $_" }
  }

  "pool-query" {
    Write-Host "==> Pool queries"
    if ($PoolID) {
      try {
        $pool = Invoke-Cli -Arguments @("query", "nominator-pool", "pool", $PoolID, "--node", $rpcNode, "--output", "json")
        Write-Host ("  pool ${PoolID}: $($pool | ConvertTo-Json -Compress)")
      } catch { Write-Host "  pool query failed: $_" }
      try {
        $share = Invoke-Cli -Arguments @("query", "nominator-pool", "pool-share", $PoolID, $userAddr, "--node", $rpcNode, "--output", "json")
        Write-Host "  pool share: $($share | ConvertTo-Json -Compress)"
      } catch { Write-Host "  pool share query failed: $_" }
    } else {
      try {
        $all = Invoke-Cli -Arguments @("query", "nominator-pool", "pools", "--node", $rpcNode, "--output", "json")
        Write-Host "  all pools: $($all | ConvertTo-Json -Compress)"
      } catch { Write-Host "  pools query failed: $_" }
    }
  }

  default { throw "Unknown action: $Action. Use: status, delegate, unbond, pool-create, pool-deposit, pool-unbond, pool-rewards, pool-query" }
}
