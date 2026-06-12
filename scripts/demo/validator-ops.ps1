param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$RPCPort = 26657,
  [string]$Fees = "1000000naet",
  [int]$TimeoutSeconds = 60,
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
  param([string]$Label, [string[]]$ActionArgs, [string]$FromKey = "node0")
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
  } catch { throw "Localnet not running on RPC $RPCPort. Start it first with scripts\localnet\start.ps1" }
}

if ($Check) {
  Write-Output "Validator operations helper"
  Write-Output "rpc: $rpcNode home: $nodeHome"
  Write-Output "commands: bonded-validator, delegations, signing-info, unjail, rewards"
  return
}

Assert-Running

$currentAction = ""
try {
  # Show bonded validator
  Write-Host "==> Bonded validator"
  $bonded = Get-LocalnetBondedValidator -Binary $Binary -RPCPort $RPCPort
  $valAddr = $bonded.operator_address
  $valMoniker = $bonded.description.moniker
  Write-Host "  operator=$valAddr moniker=$valMoniker"
  Write-Host "  tokens=$($bonded.tokens) commission=$($bonded.commission.commission_rates.rate)"
  if ($bonded.jailed -eq "true") { Write-Host "  WARNING: validator is JAILED" }

  # Show validator set  
  $currentAction = "validator-set"
  Write-Host "==> Validator set"
  $allVals = Get-LocalnetStakingValidators -Binary $Binary -RPCPort $RPCPort
  foreach ($v in $allVals) {
    $status = $v.status -replace "BOND_STATUS_", ""
    Write-Host "  $($v.operator_address): $($v.description.moniker) status=$status power=$($v.tokens)"
  }

  # Delegations
  $currentAction = "delegations"
  Write-Host "==> Delegations"
  $keys = @("node0", "node1", "node2")
  foreach ($k in $keys) {
    $addr = Invoke-CliRaw -Arguments @("keys", "show", $k, "--home", $nodeHome, "--keyring-backend", "test", "-a")
    $addr = $addr.Trim().Split("`n")[-1].Trim()
    try {
      $del = Invoke-Cli -Arguments @("query", "staking", "delegations", $addr, "--node", $rpcNode, "--output", "json")
      foreach ($d in @($del.delegation_responses)) {
        Write-Host "  $k -> $($d.delegation.validator_address): $($d.balance.amount) $($d.balance.denom)"
      }
    } catch { Write-Host ("  ${k}: no delegations") }
  }

  # Signing info
  $currentAction = "signing-info"
  Write-Host "==> Signing info"
  $infos = Get-LocalnetSigningInfos -Binary $Binary -RPCPort $RPCPort
  foreach ($info in $infos) {
    $missedPct = if ($info.index_offset -gt 0) { [math]::Round([double]$info.missed_blocks_counter / $info.index_offset * 100, 1) } else { 0 }
    Write-Host "  $($info.address): missed $($info.missed_blocks_counter)/$($info.index_offset) ($missedPct%)"
  }

  # Balances
  $currentAction = "balances"
  Write-Host "==> Key balances"
  foreach ($k in $keys) {
    $addr = Invoke-CliRaw -Arguments @("keys", "show", $k, "--home", $nodeHome, "--keyring-backend", "test", "-a")
    $addr = $addr.Trim().Split("`n")[-1].Trim()
    $bal = Get-Balance -Address $addr
    Write-Host ("  ${k}: $bal naet")
  }

  # Pending rewards
  $currentAction = "rewards"
  Write-Host "==> Pending rewards"
  foreach ($k in $keys) {
    $addr = Invoke-CliRaw -Arguments @("keys", "show", $k, "--home", $nodeHome, "--keyring-backend", "test", "-a")
    $addr = $addr.Trim().Split("`n")[-1].Trim()
    try {
      $rewards = Invoke-Cli -Arguments @("query", "distribution", "rewards", $addr, "--node", $rpcNode, "--output", "json")
      $totalNaet = $rewards.total | Where-Object { $_.denom -eq "naet" }
      if ($totalNaet) { Write-Host ("  ${k}: $($totalNaet.amount) naet") }
    } catch { Write-Host ("  ${k}: rewards query failed") }
  }

  # Staking params
  $currentAction = "staking-params"
  Write-Host "==> Staking params"
  $stakingParams = Get-LocalnetStakingParams -Binary $Binary -RPCPort $RPCPort
  Write-Host "  bond_denom=$($stakingParams.bond_denom) max_validators=$($stakingParams.max_validators) max_entries=$($stakingParams.max_entries)"
  Write-Host "  historical_entries=$($stakingParams.historical_entries) unbonding_time=$($stakingParams.unbonding_time)"

  # Slashing params
  $currentAction = "slashing-params"
  Write-Host "==> Slashing params"
  $slashParams = Get-LocalnetSlashingParams -Binary $Binary -RPCPort $RPCPort
  Write-Host "  signed_blocks_window=$($slashParams.signed_blocks_window)"
  Write-Host "  min_signed_per_window=$($slashParams.min_signed_per_window)"
  Write-Host "  downtime_jail_duration=$($slashParams.downtime_jail_duration)"
  Write-Host "  slash_fraction_double_sign=$($slashParams.slash_fraction_double_sign)"
  Write-Host "  slash_fraction_downtime=$($slashParams.slash_fraction_downtime)"

} catch {
  Write-Host "ERROR at '$currentAction': $_"
  throw
}
