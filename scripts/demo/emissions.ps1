param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$RPCPort = 26657,
  [int]$TimeoutSeconds = 30,
  [string]$Action = "all",
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

function Assert-Running {
  try {
    Wait-LocalnetRpc -RPCPort $RPCPort -TimeoutSeconds 5 | Out-Null
  } catch { throw "Localnet not running on RPC $RPCPort" }
}

function Get-Balance {
  param([string]$Address, [string]$Denom = "naet")
  $bal = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $Denom -RPCPort $RPCPort
  if (-not $bal.amount) { return [int64]0 }
  return [int64]$bal.amount
}

if ($Check) {
  Write-Output "Emissions & economics query helper"
  Write-Output "actions: inflation, apr, bonded-ratio, fee-split, rewards, supply, burn, treasury, all"
  Write-Output "rpc: $rpcNode"
  return
}

Assert-Running

$results = @{}

switch ($Action.ToLower()) {
  "inflation" {
    Write-Host "==> Current inflation"
    $r = Invoke-Cli -Arguments @("query", "aetra-economics", "current-inflation", "--node", $rpcNode, "--output", "json")
    $results.Inflation = $r
    Write-Host "  $($r | ConvertTo-Json -Compress)"
  }

  "apr" {
    Write-Host "==> Estimated APR"
    $r = Invoke-Cli -Arguments @("query", "aetra-economics", "estimated-apr", "--node", $rpcNode, "--output", "json")
    $results.APR = $r
    Write-Host "  $($r | ConvertTo-Json -Compress)"
  }

  "bonded-ratio" {
    Write-Host "==> Bonded ratio"
    $r = Invoke-Cli -Arguments @("query", "aetra-economics", "current-bonded-ratio", "--node", $rpcNode, "--output", "json")
    $results.BondedRatio = $r
    Write-Host "  $($r | ConvertTo-Json -Compress)"
  }

  "fee-split" {
    Write-Host "==> Fee split params"
    $r = Invoke-Cli -Arguments @("query", "aetra-economics", "fee-split-params", "--node", $rpcNode, "--output", "json")
    $results.FeeSplit = $r
    Write-Host "  $($r | ConvertTo-Json -Compress)"
  }

  "rewards" {
    Write-Host "==> Epoch reward summary"
    $r = Invoke-Cli -Arguments @("query", "aetra-economics", "epoch-reward-summary", "--node", $rpcNode, "--output", "json")
    $results.EpochReward = $r
    Write-Host "  $($r | ConvertTo-Json -Compress)"
  }

  "supply" {
    Write-Host "==> Total supply"
    $r = Invoke-Cli -Arguments @("query", "bank", "total-supply-of", "naet", "--node", $rpcNode, "--output", "json")
    $results.Supply = $r
    Write-Host "  $($r.amount.amount) $($r.amount.denom)"

    $height = Get-LocalnetHeight -RPCPort $RPCPort
    Write-Host "  height=$height"
  }

  "burn" {
    Write-Host "==> Burned supply"
    try {
      $r = Invoke-Cli -Arguments @("query", "burn", "total-burned", "--node", $rpcNode, "--output", "json")
      $results.Burned = $r
      Write-Host "  $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  burn query unavailable ($_)" }
  }

  "treasury" {
    Write-Host "==> Treasury"
    try {
      $r = Invoke-Cli -Arguments @("query", "treasury", "balance", "--node", $rpcNode, "--output", "json")
      $results.Treasury = $r
      Write-Host "  $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  treasury query unavailable ($_)" }

    try {
      $r = Invoke-Cli -Arguments @("query", "fee-collector", "balance", "--node", $rpcNode, "--output", "json")
      $results.FeeCollector = $r
      Write-Host "  fee_collector: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  fee_collector query unavailable ($_)" }
  }

  "all" {
    Write-Host "==> Emissions & Economics (all)"
    try {
      $r = Invoke-Cli -Arguments @("query", "aetra-economics", "current-inflation", "--node", $rpcNode, "--output", "json")
      $results.Inflation = $r; Write-Host "  inflation: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  inflation: N/A" }
    try {
      $r = Invoke-Cli -Arguments @("query", "aetra-economics", "estimated-apr", "--node", $rpcNode, "--output", "json")
      $results.APR = $r; Write-Host "  apr: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  apr: N/A" }
    try {
      $r = Invoke-Cli -Arguments @("query", "aetra-economics", "current-bonded-ratio", "--node", $rpcNode, "--output", "json")
      $results.BondedRatio = $r; Write-Host "  bonded_ratio: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  bonded_ratio: N/A" }
    try {
      $r = Invoke-Cli -Arguments @("query", "aetra-economics", "fee-split-params", "--node", $rpcNode, "--output", "json")
      $results.FeeSplit = $r; Write-Host "  fee_split: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  fee_split: N/A" }
    try {
      $r = Invoke-Cli -Arguments @("query", "aetra-economics", "epoch-reward-summary", "--node", $rpcNode, "--output", "json")
      $results.EpochReward = $r; Write-Host "  epoch_reward: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  epoch_reward: N/A" }
    try {
      $r = Invoke-Cli -Arguments @("query", "bank", "total-supply-of", "naet", "--node", $rpcNode, "--output", "json")
      $results.Supply = $r; Write-Host "  naet_supply: $($r.amount.amount)"
    } catch { Write-Host "  naet_supply: N/A" }
    try {
      $r = Invoke-Cli -Arguments @("query", "burn", "total-burned", "--node", $rpcNode, "--output", "json")
      $results.Burned = $r; Write-Host "  burned: $($r | ConvertTo-Json -Compress)"
    } catch { Write-Host "  burned: N/A" }
  }

  default { throw "Unknown action: $Action. Use: inflation, apr, bonded-ratio, fee-split, rewards, supply, burn, treasury, all" }
}

if ($Json) { $results | ConvertTo-Json -Depth 10 }
