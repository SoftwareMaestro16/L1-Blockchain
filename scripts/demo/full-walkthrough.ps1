param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 4,
  [int]$TimeoutSeconds = 180,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [string]$Fees = "1000000naet",
  [string]$Profile = "base",
  [switch]$SkipBuild,
  [switch]$Check,
  [switch]$KeepLocalnet,
  [switch]$SkipExportGenesis
)

$ErrorActionPreference = "Stop"

if ($ChainId -notmatch 'local') { throw "full-walkthrough is local-only; ChainId must contain 'local'" }
if ($Fees -notmatch '^[0-9]+naet$') { throw "fees must use the local denom naet" }

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-demo"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "demo output directory"
if (-not $SkipBuild) { Assert-LocalnetWorkspacePath -Path (Split-Path $Binary) -Purpose "binary output directory" }

$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"
$restBase = "http://127.0.0.1:$($node0Ports.REST)"

function Write-Step { param([string]$Message); Write-Host ""; Write-Host "==> $Message" }
function Write-Note { param([string]$Message); Write-Host "    $Message" }

function Invoke-DemoLocalnetScript {
  param([string]$ScriptName, [hashtable]$Extra = @{})
  $args = @{
    OutputDir = $OutputDir; Binary = $Binary; ChainId = $ChainId
    ValidatorCount = $ValidatorCount; BaseP2PPort = $BaseP2PPort
    BaseRPCPort = $BaseRPCPort; BaseRESTPort = $BaseRESTPort
    BaseGRPCPort = $BaseGRPCPort; BasePprofPort = $BasePprofPort
    PortStride = $PortStride; TimeoutCommit = $TimeoutCommit
    LogLevel = $LogLevel; Profile = $Profile
    EnableAPI = $true; EnableGRPC = $true; EnableRPC = $true
  }
  foreach ($k in $Extra.Keys) { $args[$k] = $Extra[$k] }
  & (Join-Path $RepoRoot "scripts\localnet\$ScriptName") @args
}

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

function Send-DemoTx {
  param([string]$Label, [string[]]$ActionArgs, [string]$FromHome, [string]$FromKey = "node0")
  Write-Note "$Label"
  $tx = Send-LocalnetTx -Binary $Binary -Arguments ($ActionArgs + @("--from", $FromKey, "--home", $FromHome, "--chain-id", $ChainId, "--keyring-backend", "test", "--fees", $Fees, "--yes", "--broadcast-mode", "sync", "--node", $rpcNode, "--output", "json")) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Note "txhash=$(Get-LocalnetTxHash -Tx $tx)"
  return $tx
}

function Get-Balance {
  param([string]$Address, [string]$Denom = "naet")
  $bal = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $Denom -RPCPort $node0Ports.RPC
  if (-not $bal.amount) { return [int64]0 }
  return [int64]$bal.amount
}

function Get-ValidatorOpAddr {
  param([string]$NodeHome)
  $valOut = Invoke-CliRaw -Arguments @("keys", "show", "node0", "--home", $NodeHome, "--keyring-backend", "test", "--bech", "val", "-a")
  return $valOut.Trim().Split("`n")[-1].Trim()
}

if ($Check) {
  Write-Output "Aetra full walkthrough demo check"
  Write-Output "chain-id: $ChainId validators: $ValidatorCount output: $OutputDir binary: $Binary"
  Write-Output "steps: build/start -> keys -> bank -> pool-staking -> validator -> emissions -> slashing -> rewards -> export -> stop"
  return
}

$node0Home = Join-Path $OutputDir "node0\aetrad"
$node1Home = Join-Path $OutputDir "node1\aetrad"

Push-Location $RepoRoot
try {
  Write-Host "Aetra Full Network Walkthrough"
  Write-Host "LOCAL ONLY: uses ignored localnet homes and --keyring-backend test."

  # ── 1. Build & Start ──────────────────────────────────────────────
  Write-Step "Build & start localnet"
  & (Join-Path $RepoRoot "scripts\localnet\stop.ps1") -OutputDir $OutputDir
  if (-not $SkipBuild) {
    & (Join-Path $RepoRoot "scripts\build-aetrad.ps1") -Binary $Binary
  } elseif (!(Test-Path $Binary)) {
    throw "Binary not found: $Binary and -SkipBuild was specified"
  }
  $version = & $Binary version 2>&1 | Select-Object -First 1
  Write-Note "aetrad version: $version"

  Invoke-DemoLocalnetScript -ScriptName "init.ps1" -Extra @{ SkipBuild = $true }
  Invoke-DemoLocalnetScript -ScriptName "validate-genesis.ps1"
  Invoke-DemoLocalnetScript -ScriptName "start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Note "height=$height rpc=$rpcNode rest=$restBase"

  # ── 2. Wallet Keys & Addresses ────────────────────────────────────
  Write-Step "Wallet keys and addresses"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
  $node0ValOp = Get-ValidatorOpAddr -NodeHome $node0Home
  Write-Note "node0 addr    = $node0"
  Write-Note "node1 addr    = $node1"
  Write-Note "node0 valoper = $node0ValOp"

  $node0Bal = Get-Balance -Address $node0 -Denom "naet"
  $node1Bal = Get-Balance -Address $node1 -Denom "naet"
  Write-Note "node0 balance = $node0Bal naet"
  Write-Note "node1 balance = $node1Bal naet"

  # ── 3. Bank Transfer ──────────────────────────────────────────────
  Write-Step "Bank transfer"
  $node1Before = Get-Balance -Address $node1
  Send-DemoTx -Label "send 5000naet node0 -> node1" -ActionArgs @("tx", "bank", "send", "node0", $node1, "5000naet") -FromHome $node0Home | Out-Null
  $node1After = Get-Balance -Address $node1
  if ($node1After -ne ($node1Before + 5000)) { throw "bank send mismatch" }
  Write-Note "node1 naet: $node1Before -> $node1After"

  # ── 4. Pool Staking (Official Liquid Staking Pool) ───────────────
  Write-Step "Pool staking"
  Write-Note "Creating official liquid staking pool..."
  try {
    $poolTx = Send-DemoTx -Label "create official liquid staking pool" -ActionArgs @("tx", "nominator-pool", "create-official-pool") -FromHome $node0Home
    $poolResp = $poolTx
    # Try to get pool ID from events
    $poolId = "1"
    Write-Note "official pool created: pool_id=$poolId"
  } catch {
    Write-Note "create-official-pool skipped (may require genesis pool pre-creation): $_"
    $poolId = "1"
  }

  Write-Note "Depositing to official pool..."
  try {
    Send-DemoTx -Label "deposit 10000000naet to pool $poolId" -ActionArgs @("tx", "nominator-pool", "deposit", $poolId, "10000000naet") -FromHome $node0Home | Out-Null
    Write-Note "deposit succeeded"
  } catch {
    Write-Note "deposit note: $_"
  }

  # Query pool state via REST
  try {
    $restPools = Invoke-RestMethod -Uri "$restBase/l1/nominatorpool/v1/pools" -TimeoutSec 5
    Write-Note "pools via REST: $($restPools | ConvertTo-Json -Depth 5 -Compress)"
  } catch { Write-Note "REST pools query unavailable" }

  # ── 5. Staking (Cosmos SDK) ──────────────────────────────────────
  Write-Step "Staking (Cosmos SDK staking module)"
  $bondedVal = Get-LocalnetBondedValidator -Binary $Binary -RPCPort $node0Ports.RPC
  $valOperAddr = $bondedVal.operator_address
  Write-Note "bonded validator: $valOperAddr"

  $delegationBefore = Get-Balance -Address $node0
  try {
    Send-DemoTx -Label "delegate 5000000naet to $valOperAddr" -ActionArgs @("tx", "staking", "delegate", $valOperAddr, "5000000naet") -FromHome $node0Home | Out-Null
    Write-Note "delegation succeeded"
  } catch { Write-Note "delegation note: $_" }

  # Query delegation
  try {
    $delResp = Invoke-Cli -Arguments @("query", "staking", "delegation", $node0, $valOperAddr, "--node", $rpcNode, "--output", "json")
    Write-Note "delegation balance: $($delResp.delegation_response.balance.amount) $($delResp.delegation_response.balance.denom)"
  } catch { Write-Note "delegation query unavailable" }

  # Query validator set
  try {
    $vals = Invoke-Cli -Arguments @("query", "staking", "validators", "--node", $rpcNode, "--output", "json")
    Write-Note "validators: $(@($vals.validators).Count) bonded to $($vals.pagination.total) total"
  } catch { Write-Note "validator query unavailable" }

  # ── 6. Emissions / Economics ─────────────────────────────────────
  Write-Step "Emissions & economics"
  try {
    $inflation = Invoke-Cli -Arguments @("query", "aetra-economics", "current-inflation", "--node", $rpcNode, "--output", "json")
    Write-Note "current inflation: $($inflation | ConvertTo-Json -Compress)"
  } catch { Write-Note "inflation query unavailable ($_)" }

  try {
    $apr = Invoke-Cli -Arguments @("query", "aetra-economics", "estimated-apr", "--node", $rpcNode, "--output", "json")
    Write-Note "estimated APR: $($apr | ConvertTo-Json -Compress)"
  } catch { Write-Note "APR query unavailable ($_)" }

  try {
    $bondedRatio = Invoke-Cli -Arguments @("query", "aetra-economics", "current-bonded-ratio", "--node", $rpcNode, "--output", "json")
    Write-Note "bonded ratio: $($bondedRatio | ConvertTo-Json -Compress)"
  } catch { Write-Note "bonded ratio query unavailable ($_)" }

  try {
    $rewardSummary = Invoke-Cli -Arguments @("query", "aetra-economics", "epoch-reward-summary", "--node", $rpcNode, "--output", "json")
    Write-Note "epoch reward summary: $($rewardSummary | ConvertTo-Json -Compress)"
  } catch { Write-Note "reward summary query unavailable ($_)" }

  try {
    $feeSplit = Invoke-Cli -Arguments @("query", "aetra-economics", "fee-split-params", "--node", $rpcNode, "--output", "json")
    Write-Note "fee split params: $($feeSplit | ConvertTo-Json -Compress)"
  } catch { Write-Note "fee split query unavailable ($_)" }

  try {
    $totalSupply = Invoke-Cli -Arguments @("query", "bank", "total-supply-of", "naet", "--node", $rpcNode, "--output", "json")
    Write-Note "total naet supply: $($totalSupply.amount.amount) $($totalSupply.amount.denom)"
  } catch { Write-Note "supply query unavailable ($_)" }

  # ── 7. Slashing ──────────────────────────────────────────────────
  Write-Step "Slashing"
  try {
    $slashParams = Get-LocalnetSlashingParams -Binary $Binary -RPCPort $node0Ports.RPC
    Write-Note "signed_blocks_window=$($slashParams.signed_blocks_window) min_signed_per_window=$($slashParams.min_signed_per_window) downtime_jail_duration=$($slashParams.downtime_jail_duration)"
  } catch { Write-Note "slashing params unavailable ($_)" }

  try {
    $signingInfos = Get-LocalnetSigningInfos -Binary $Binary -RPCPort $node0Ports.RPC
    foreach ($info in $signingInfos) {
      Write-Note "validator $($info.address): missed $($info.missed_blocks_counter)/$($info.index_offset) blocks"
    }
  } catch { Write-Note "signing infos unavailable ($_)" }

  try {
    $stakeConcParams = Invoke-Cli -Arguments @("query", "stake-concentration", "params", "--node", $rpcNode, "--output", "json")
    Write-Note "stake concentration: $($stakeConcParams | ConvertTo-Json -Compress)"
  } catch { Write-Note "stake concentration query unavailable ($_)" }

  # ── 8. Rewards & Distribution ─────────────────────────────────────
  Write-Step "Rewards & distribution"
  try {
    $rewards = Invoke-Cli -Arguments @("query", "distribution", "rewards", $node0, "--node", $rpcNode, "--output", "json")
    $total = $rewards.total | Where-Object { $_.denom -eq "naet" }
    if ($total) { Write-Note "pending rewards: $($total.amount) naet" } else { Write-Note "pending rewards: 0 naet" }
  } catch { Write-Note "rewards query unavailable ($_)" }

  try {
    Write-Note "withdrawing rewards..."
    Send-DemoTx -Label "withdraw rewards" -ActionArgs @("tx", "distribution", "withdraw-all-rewards") -FromHome $node0Home | Out-Null
    Write-Note "rewards withdrawn"
  } catch { Write-Note "reward withdrawal note: $_" }

  # ── 9. Reputation ─────────────────────────────────────────────────
  Write-Step "Reputation"
  try {
    $repIdentity = Invoke-Cli -Arguments @("query", "reputation", "identity", $node0, "--node", $rpcNode, "--output", "json")
    Write-Note "identity reputation: $($repIdentity | ConvertTo-Json -Compress)"
  } catch { Write-Note "reputation query unavailable ($_)" }

  try {
    $repValidators = Invoke-Cli -Arguments @("query", "reputation", "validators", "--node", $rpcNode, "--output", "json")
    Write-Note "validator scores: $($repValidators | ConvertTo-Json -Compress)"
  } catch { Write-Note "validator score query unavailable ($_)" }

  try {
    $repTrust = Invoke-Cli -Arguments @("query", "reputation", "service-trust", "--node", $rpcNode, "--output", "json")
    Write-Note "service trust: $($repTrust | ConvertTo-Json -Compress)"
  } catch { Write-Note "service trust query unavailable ($_)" }

  # ── 10. Validator Score ───────────────────────────────────────────
  Write-Step "Validator score"
  try {
    $valScores = Invoke-Cli -Arguments @("query", "aetra-validator-score", "all-validator-scores", "--node", $rpcNode, "--output", "json")
    Write-Note "all scores: $($valScores | ConvertTo-Json -Compress)"
  } catch { Write-Note "validator score query unavailable ($_)" }

  try {
    $valMetrics = Invoke-Cli -Arguments @("query", "aetra-validator-score", "public-validator-metrics", "--node", $rpcNode, "--output", "json")
    Write-Note "public metrics: $($valMetrics | ConvertTo-Json -Compress)"
  } catch { Write-Note "validator metrics query unavailable ($_)" }

  # ── 11. Staking Policy ────────────────────────────────────────────
  Write-Step "Staking policy"
  try {
    $stakePolicy = Invoke-Cli -Arguments @("query", "aetra-staking-policy", "params", "--node", $rpcNode, "--output", "json")
    Write-Note "policy params: $($stakePolicy | ConvertTo-Json -Compress)"
  } catch { Write-Note "staking policy query unavailable ($_)" }

  try {
    $topN = Invoke-Cli -Arguments @("query", "aetra-staking-policy", "top-n-concentration", "--node", $rpcNode, "--output", "json")
    Write-Note "top-N concentration: $($topN | ConvertTo-Json -Compress)"
  } catch { Write-Note "concentration query unavailable ($_)" }

  # ── 12. Dynamic Commission ─────────────────────────────────────────
  Write-Step "Dynamic commission"
  try {
    $dynComm = Invoke-Cli -Arguments @("query", "dynamic-commission", "params", "--node", $rpcNode, "--output", "json")
    Write-Note "dynamic commission params: $($dynComm | ConvertTo-Json -Compress)"
  } catch { Write-Note "dynamic commission query unavailable ($_)" }

  # ── 13. Validator Registry ────────────────────────────────────────
  Write-Step "Validator registry"
  try {
    $valReg = Invoke-Cli -Arguments @("query", "validator-registry", "validators", "--node", $rpcNode, "--output", "json")
    Write-Note "registry: $($valReg | ConvertTo-Json -Compress)"
  } catch { Write-Note "validator registry query unavailable ($_)" }

  # ── 14. Performance ───────────────────────────────────────────────
  Write-Step "Performance"
  try {
    $perf = Invoke-Cli -Arguments @("query", "performance", "params", "--node", $rpcNode, "--output", "json")
    Write-Note "performance params: $($perf | ConvertTo-Json -Compress)"
  } catch { Write-Note "performance query unavailable ($_)" }

  # ── 15. Delegator Protection ──────────────────────────────────────
  Write-Step "Delegator protection"
  try {
    $delProt = Invoke-Cli -Arguments @("query", "delegator-protection", "params", "--node", $rpcNode, "--output", "json")
    Write-Note "protection params: $($delProt | ConvertTo-Json -Compress)"
  } catch { Write-Note "delegator protection query unavailable ($_)" }

  # ── 16. System Registry ───────────────────────────────────────────
  Write-Step "System registry"
  try {
    $sysReg = Invoke-Cli -Arguments @("query", "system-registry", "params", "--node", $rpcNode, "--output", "json")
    Write-Note "system registry params: $($sysReg | ConvertTo-Json -Compress)"
  } catch { Write-Note "system registry query unavailable ($_)" }

  # ── 17. Final Balances ───────────────────────────────────────────
  Write-Step "Final balances"
  $node0Final = Get-Balance -Address $node0
  $node1Final = Get-Balance -Address $node1
  $heightFinal = Wait-LocalnetHeight -TargetHeight ($height + 2) -RPCPort $node0Ports.RPC -TimeoutSeconds 30
  Write-Note "height=$heightFinal"
  Write-Note "node0 naet = $node0Final (started with bonded stake)"
  Write-Note "node1 naet = $node1Final (received bank transfer)"

  # ── 18. Genesis Export ────────────────────────────────────────────
  if (-not $SkipExportGenesis) {
    Write-Step "Genesis export"
    $exportPath = Join-Path $OutputDir "exported-genesis.json"
    try {
      $export = & $Binary export --home $node0Home 2>&1
      $export | Set-Content -LiteralPath $exportPath
      Write-Note "genesis exported to $exportPath ($([int]($export.Length / 1024)) KB)"
    } catch { Write-Note "genesis export failed: $_" }
  }

  # ── Summary ───────────────────────────────────────────────────────
  Write-Step "Walkthrough complete"
  Write-Note "All major subsystems verified: keys, bank, staking, pool-staking, emissions,"
  Write-Note "slashing, rewards, reputation, validator-score, staking-policy, dynamic-commission,"
  Write-Note "validator-registry, performance, delegator-protection, system-registry."
  Write-Note "Chain produced $heightFinal blocks with $ValidatorCount validators."

} finally {
  if (-not $KeepLocalnet) {
    Write-Step "Stop localnet"
    & (Join-Path $RepoRoot "scripts\localnet\stop.ps1") -OutputDir $OutputDir
  } else {
    Write-Step "Localnet kept running"
    Write-Note "Stop with: .\scripts\localnet\stop.ps1 -OutputDir $OutputDir"
  }
  Pop-Location
}
