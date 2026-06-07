param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [ValidateSet("bank", "contract-assets", "dex", "mixed")]
  [string]$Scenario = "mixed",
  [int]$Count = 12,
  [decimal]$RatePerSecond = 2,
  [int]$RPCPort = 26657,
  [int]$TimeoutSeconds = 90,
  [string]$Fees = "1000000naet",
  [string]$FactorySubdenom = "loadasset",
  [switch]$Json
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

if ($Count -lt 1) {
  throw "Count must be at least 1"
}
if ($RatePerSecond -le 0) {
  throw "RatePerSecond must be positive"
}
if ($RPCPort -lt 1 -or $RPCPort -gt 65535) {
  throw "RPCPort must be a valid TCP port"
}
if ($ChainId -notmatch '(^|-)local($|-)' -and $ChainId -notmatch 'local') {
  throw "load-profile is local/test-only; ChainId must contain 'local'"
}
if ($Fees -notmatch '^[0-9]+naet$') {
  throw "Fees must be a naet coin for the local prototype profile"
}

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

if (!(Test-Path -LiteralPath $Binary)) {
  throw "Binary not found at $Binary"
}

$rpcNode = "tcp://127.0.0.1:$RPCPort"
$status = Wait-LocalnetRpc -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds
if ($status.result.node_info.network -ne $ChainId) {
  throw "connected chain-id mismatch: expected $ChainId, got $($status.result.node_info.network)"
}

$node0Home = Join-Path $OutputDir "node0\aetrad"
$node1Home = Join-Path $OutputDir "node1\aetrad"
if (!(Test-Path -LiteralPath $node0Home) -or !(Test-Path -LiteralPath $node1Home)) {
  throw "load-profile requires local node0 and node1 homes under $OutputDir"
}

$node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
$node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
$factoryDenom = "factory/$node0/$FactorySubdenom"
$poolId = 1

function Invoke-QueryCliJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--node", $rpcNode, "--output", "json"))
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

function Send-ProfileTx {
  param(
    [string]$Operation,
    [string[]]$ActionArgs,
    [string]$FromHome = $node0Home,
    [string]$FromKey = "node0"
  )

  $sw = [System.Diagnostics.Stopwatch]::StartNew()
  try {
    $tx = Send-LocalnetTx `
      -Binary $Binary `
      -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey) `
      -RPCPort $RPCPort `
      -TimeoutSeconds $TimeoutSeconds
    $sw.Stop()
    return [ordered]@{
      operation  = $Operation
      ok         = $true
      latency_ms = [int64]$sw.ElapsedMilliseconds
      txhash     = Get-LocalnetTxHash -Tx $tx
      code       = Get-LocalnetTxCode -Tx $tx
      error      = ""
    }
  } catch {
    $sw.Stop()
    $err = $_.Exception.Message -replace '\s+', ' '
    if ($err.Length -gt 300) {
      $err = $err.Substring(0, 300)
    }
    return [ordered]@{
      operation  = $Operation
      ok         = $false
      latency_ms = [int64]$sw.ElapsedMilliseconds
      txhash     = ""
      code       = -1
      error      = $err
    }
  }
}

function Test-QuerySucceeds {
  param([string[]]$Arguments)

  try {
    Invoke-QueryCliJson -Arguments $Arguments | Out-Null
    return $true
  } catch {
    return $false
  }
}

function Ensure-Contract assetsAsset {
  if (-not (Test-QuerySucceeds -Arguments @("query", "contract-assets", "denom", $factoryDenom))) {
    $res = Send-ProfileTx -Operation "setup_contract-assets_create" -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom)
    if (-not $res.ok) {
      throw "failed to create load contract-assets denom: $($res.error)"
    }
  }
  $mint = Send-ProfileTx -Operation "setup_contract-assets_mint" -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom", $node0)
  if (-not $mint.ok) {
    throw "failed to mint load contract-assets denom: $($mint.error)"
  }
}

function Ensure-DexPool {
  Ensure-Contract assetsAsset
  try {
    $pool = Invoke-QueryCliJson -Arguments @("query", "dex", "pool", "$poolId")
    if ($pool.pool.denom0 -eq $factoryDenom -and $pool.pool.denom1 -eq "naet") {
      return
    }
  } catch {
  }
  $res = Send-ProfileTx -Operation "setup_dex_create_pool" -ActionArgs @("tx", "dex", "create-pool", "10000000naet", "10000000$factoryDenom")
  if (-not $res.ok) {
    throw "failed to create load DEX pool: $($res.error)"
  }
}

function Get-ScenarioOperation {
  param([int]$Index)

  if ($Scenario -eq "bank") { return "bank" }
  if ($Scenario -eq "contract-assets") { return "contract-assets" }
  if ($Scenario -eq "dex") { return "dex" }
  switch ($Index % 3) {
    0 { return "bank" }
    1 { return "contract-assets" }
    default { return "dex" }
  }
}

if ($Scenario -eq "contract-assets" -or $Scenario -eq "mixed") {
  Ensure-Contract assetsAsset
}
if ($Scenario -eq "dex" -or $Scenario -eq "mixed") {
  Ensure-DexPool
}

$startHeight = Get-LocalnetHeight -RPCPort $RPCPort
$startedAt = Get-Date
$results = @()
$intervalMs = [int][Math]::Ceiling(1000 / [double]$RatePerSecond)

for ($i = 0; $i -lt $Count; $i++) {
  $iterationStart = [DateTime]::UtcNow
  $op = Get-ScenarioOperation -Index $i
  switch ($op) {
    "bank" {
      $result = Send-ProfileTx -Operation "bank_send" -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet")
    }
    "contract-assets" {
      $result = Send-ProfileTx -Operation "contract-assets_mint" -ActionArgs @("tx", "contract-assets", "mint", "10$factoryDenom", $node0)
    }
    "dex" {
      $result = Send-ProfileTx -Operation "dex_swap" -ActionArgs @("tx", "dex", "swap-exact-in", "$poolId", "1000naet", $factoryDenom, "1")
    }
  }
  $results += [pscustomobject]$result

  $elapsedMs = [int]([DateTime]::UtcNow - $iterationStart).TotalMilliseconds
  $sleepMs = $intervalMs - $elapsedMs
  if ($sleepMs -gt 0 -and $i -lt ($Count - 1)) {
    Start-Sleep -Milliseconds $sleepMs
  }
}

$endedAt = Get-Date
$endHeight = Get-LocalnetHeight -RPCPort $RPCPort
$okResults = @($results | Where-Object { $_.ok })
$failedResults = @($results | Where-Object { -not $_.ok })
$latencies = @($okResults | ForEach-Object { [int64]$_.latency_ms } | Sort-Object)
$avgLatency = if ($latencies.Count -gt 0) { [Math]::Round((($latencies | Measure-Object -Average).Average), 2) } else { $null }
$minLatency = if ($latencies.Count -gt 0) { [int64]$latencies[0] } else { $null }
$maxLatency = if ($latencies.Count -gt 0) { [int64]$latencies[$latencies.Count - 1] } else { $null }
$p95Latency = if ($latencies.Count -gt 0) {
  $idx = [Math]::Ceiling($latencies.Count * 0.95) - 1
  if ($idx -lt 0) { $idx = 0 }
  if ($idx -ge $latencies.Count) { $idx = $latencies.Count - 1 }
  [int64]$latencies[$idx]
} else {
  $null
}

$operationCounts = [ordered]@{}
foreach ($group in ($results | Group-Object operation)) {
  $operationCounts[$group.Name] = [ordered]@{
    total    = $group.Count
    success  = @($group.Group | Where-Object { $_.ok }).Count
    failures = @($group.Group | Where-Object { -not $_.ok }).Count
  }
}

$duration = ($endedAt - $startedAt).TotalSeconds
$summary = [ordered]@{
  profile           = "local-minimal-load"
  scenario          = $Scenario
  chain_id          = $ChainId
  rpc               = "127.0.0.1:$RPCPort"
  count             = $Count
  target_rate_tps   = [double]$RatePerSecond
  duration_seconds  = [Math]::Round($duration, 2)
  observed_tps      = if ($duration -gt 0) { [Math]::Round($Count / $duration, 2) } else { $null }
  start_height      = $startHeight
  end_height        = $endHeight
  blocks_progressed = $endHeight - $startHeight
  successes         = $okResults.Count
  failures          = $failedResults.Count
  failure_rate      = [Math]::Round($failedResults.Count / [double]$Count, 4)
  latency_ms        = [ordered]@{
    min = $minLatency
    avg = $avgLatency
    p95 = $p95Latency
    max = $maxLatency
  }
  operations        = $operationCounts
  failure_samples   = @($failedResults | Select-Object -First 5 operation, code, error)
}

if ($Json) {
  $summary | ConvertTo-Json -Depth 8
} else {
  Write-Host "profile: $($summary.profile)"
  Write-Host "scenario: $Scenario count=$Count target_rate_tps=$([double]$RatePerSecond)"
  Write-Host "height: $startHeight->$endHeight blocks_progressed=$($summary.blocks_progressed)"
  Write-Host "successes: $($summary.successes) failures=$($summary.failures) failure_rate=$($summary.failure_rate)"
  Write-Host "latency_ms: min=$minLatency avg=$avgLatency p95=$p95Latency max=$maxLatency"
  Write-Host "observed_tps: $($summary.observed_tps)"
  foreach ($key in $operationCounts.Keys) {
    $item = $operationCounts[$key]
    Write-Host "operation[$key]: total=$($item.total) success=$($item.success) failures=$($item.failures)"
  }
}
