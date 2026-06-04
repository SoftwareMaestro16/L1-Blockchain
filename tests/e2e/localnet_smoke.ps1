param(
  [string[]]$ValidatorCounts = @("3", "5", "10"),
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [string]$OutputDir = "",
  [string]$Binary = "",
  [switch]$SkipTxFlow,
  [switch]$SkipNegativeTests
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")
$Binary = Resolve-BinaryPath -Binary $Binary

$ResolvedValidatorCounts = @()
foreach ($rawCount in $ValidatorCounts) {
  foreach ($part in ([string]$rawCount).Split(",")) {
    $trimmed = $part.Trim()
    if ([string]::IsNullOrWhiteSpace($trimmed)) {
      continue
    }
    $ResolvedValidatorCounts += [int]$trimmed
  }
}
if ($ResolvedValidatorCounts.Count -eq 0) {
  throw "at least one validator count is required"
}

function Get-CliJson {
  param([Parameter(Mandatory = $true)]$Output)

  $text = ($Output | ForEach-Object { "$_" }) -join "`n"
  $objectStart = $text.IndexOf("{")
  $arrayStart = $text.IndexOf("[")
  $jsonStart = -1
  if ($objectStart -ge 0 -and $arrayStart -ge 0) {
    $jsonStart = [Math]::Min($objectStart, $arrayStart)
  } elseif ($objectStart -ge 0) {
    $jsonStart = $objectStart
  } elseif ($arrayStart -ge 0) {
    $jsonStart = $arrayStart
  }
  if ($jsonStart -lt 0) {
    throw "CLI output did not contain JSON: $text"
  }
  return ($text.Substring($jsonStart) | ConvertFrom-Json)
}

function Invoke-OrbitalisJson {
  param([Parameter(Mandatory = $true)][string[]]$Arguments)

  $output = Invoke-ExternalChecked -FilePath $Binary -Arguments $Arguments -FailureMessage "orbitalisd command failed"
  return Get-CliJson -Output $output
}

function Wait-ForHeight {
  param(
    [Parameter(Mandatory = $true)][string]$Node,
    [int]$TargetHeight,
    [int]$TimeoutSeconds
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  $height = 0
  while ((Get-Date) -lt $deadline) {
    Start-Sleep -Seconds 2
    try {
      $block = Invoke-OrbitalisJson -Arguments @("query", "block", "--node", $Node, "--output", "json")
      $heightValue = $block.header.height
      if (-not $heightValue -and $block.block) {
        $heightValue = $block.block.header.height
      }
      if ($heightValue) {
        $height = [int]$heightValue
        if ($height -ge $TargetHeight) {
          return $height
        }
      }
    }
    catch {
      continue
    }
  }
  throw "localnet did not reach height $TargetHeight within $TimeoutSeconds seconds; last height $height"
}

function Invoke-Tx {
  param(
    [Parameter(Mandatory = $true)][string[]]$Arguments,
    [Parameter(Mandatory = $true)][string]$NodeHome,
    [string]$From = "node0",
    [Parameter(Mandatory = $true)][string]$Node,
    [Parameter(Mandatory = $true)][string]$ChainId,
    [int]$TimeoutSeconds
  )

  $txArgs = $Arguments + @(
    "--home", $NodeHome,
    "--from", $From,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--node", $Node,
    "--fees", $script:DefaultFee,
    "--gas", "500000",
    "--broadcast-mode", "sync",
    "--output", "json",
    "--yes"
  )
  $currentHeight = Wait-ForHeight -Node $Node -TargetHeight 1 -TimeoutSeconds $TimeoutSeconds
  $res = Invoke-OrbitalisJson -Arguments $txArgs
  if ($res.code -and [int]$res.code -ne 0) {
    throw "tx broadcast failed with code $($res.code): $($res.raw_log)"
  }
  if ([string]::IsNullOrWhiteSpace([string]$res.txhash)) {
    throw "tx broadcast response missing txhash"
  }
  Wait-ForHeight -Node $Node -TargetHeight ($currentHeight + 2) -TimeoutSeconds $TimeoutSeconds | Out-Null
  return $res
}

function Get-KeyAddress {
  param(
    [Parameter(Mandatory = $true)][string]$NodeHome,
    [Parameter(Mandatory = $true)][string]$Name
  )

  $output = Invoke-ExternalChecked -FilePath $Binary -Arguments @("keys", "show", $Name, "-a", "--home", $NodeHome, "--keyring-backend", "test") -FailureMessage "key lookup failed"
  return (($output | Select-Object -Last 1) -as [string]).Trim()
}

function Get-BalanceAmount {
  param(
    [Parameter(Mandatory = $true)][string]$Node,
    [Parameter(Mandatory = $true)][string]$Address,
    [Parameter(Mandatory = $true)][string]$Denom
  )

  $balances = Invoke-OrbitalisJson -Arguments @("query", "bank", "balances", $Address, "--node", $Node, "--output", "json")
  foreach ($coin in $balances.balances) {
    if ($coin.denom -eq $Denom) {
      return [Int64]$coin.amount
    }
  }
  return [Int64]0
}

function Assert-CommandFails {
  param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][scriptblock]$Command
  )

  try {
    & $Command
  }
  catch {
    Write-Host "negative check passed: $Name"
    return
  }
  throw "negative check failed: $Name"
}

function Assert-MetricsEndpoint {
  param(
    [Parameter(Mandatory = $true)][string]$Url,
    [Parameter(Mandatory = $true)][string[]]$ExpectedNames
  )

  $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 10
  if ($response.StatusCode -ne 200) {
    throw "metrics endpoint $Url returned status $($response.StatusCode)"
  }
  $text = [string]$response.Content
  foreach ($name in $ExpectedNames) {
    if (!$text.Contains($name)) {
      throw "metrics endpoint $Url missing metric $name"
    }
  }
}

function Invoke-ModuleSmoke {
  param(
    [Parameter(Mandatory = $true)]$Manifest,
    [Parameter(Mandatory = $true)][string]$OutputDir,
    [int]$TimeoutSeconds
  )

  $chainId = [string]$Manifest.chain_id
  $node = [string]$Manifest.nodes[0].rpc_url
  $home0 = Get-NodeHome -OutputDir $OutputDir -Index 0
  $home1 = Get-NodeHome -OutputDir $OutputDir -Index 1
  $addr0 = Get-KeyAddress -NodeHome $home0 -Name "node0"
  $addr1 = Get-KeyAddress -NodeHome $home1 -Name "node1"

  $before = Get-BalanceAmount -Node $node -Address $addr1 -Denom "norb"
  Invoke-Tx -Arguments @("tx", "bank", "send", "node0", $addr1, "10000000norb") -NodeHome $home0 -Node $node -ChainId $chainId -TimeoutSeconds $TimeoutSeconds | Out-Null
  $after = Get-BalanceAmount -Node $node -Address $addr1 -Denom "norb"
  if ($after -le $before) {
    throw "bank send did not increase recipient balance: before=$before after=$after"
  }

  $subdenom = "smoke"
  $factoryDenom = "factory/$addr0/$subdenom"
  Invoke-Tx -Arguments @("tx", "tokenfactory", "create-denom", $subdenom) -NodeHome $home0 -Node $node -ChainId $chainId -TimeoutSeconds $TimeoutSeconds | Out-Null
  Invoke-Tx -Arguments @("tx", "tokenfactory", "mint", "300000000$factoryDenom", $addr0) -NodeHome $home0 -Node $node -ChainId $chainId -TimeoutSeconds $TimeoutSeconds | Out-Null
  Invoke-OrbitalisJson -Arguments @("query", "tokenfactory", "denom", $factoryDenom, "--node", $node, "--output", "json") | Out-Null

  Invoke-Tx -Arguments @("tx", "dex", "create-pool", "100000000norb", "100000000$factoryDenom") -NodeHome $home0 -Node $node -ChainId $chainId -TimeoutSeconds $TimeoutSeconds | Out-Null
  $pools = Invoke-OrbitalisJson -Arguments @("query", "dex", "pools", "--node", $node, "--output", "json")
  if (!$pools.pools -or $pools.pools.Count -lt 1) {
    throw "DEX create-pool did not create a pool"
  }
  $poolID = [string]$pools.pools[0].id
  Invoke-OrbitalisJson -Arguments @("query", "dex", "pool", $poolID, "--node", $node, "--output", "json") | Out-Null
  Invoke-Tx -Arguments @("tx", "dex", "swap-exact-in", $poolID, "1000000norb", $factoryDenom, "1") -NodeHome $home0 -Node $node -ChainId $chainId -TimeoutSeconds $TimeoutSeconds | Out-Null
  Invoke-OrbitalisJson -Arguments @("query", "dex", "pools", "--node", $node, "--output", "json") | Out-Null

  Invoke-OrbitalisJson -Arguments @("query", "fees", "params", "--node", $node, "--output", "json") | Out-Null
  Invoke-OrbitalisJson -Arguments @("query", "fees", "accounting", "--node", $node, "--output", "json") | Out-Null
  Invoke-OrbitalisJson -Arguments @("query", "fees", "module-balances", "--node", $node, "--output", "json") | Out-Null

  if ($Manifest.nodes[0].app_metrics_url) {
    Assert-MetricsEndpoint -Url ([string]$Manifest.nodes[0].app_metrics_url) -ExpectedNames @(
      "orbitalis_dex_swaps_total",
      "orbitalis_fees_accepted_total",
      "orbitalis_dex_pool_count"
    )
  }
  Write-Host "tx/module smoke passed on $($Manifest.validator_count)-validator localnet"
}

function Run-LocalnetSmoke {
  param([int]$ValidatorCount)

  if ($ValidatorCount -lt 2) {
    throw "localnet smoke requires at least two validators"
  }

  $runOutputDir = $OutputDir
  if ([string]::IsNullOrWhiteSpace($runOutputDir)) {
    if ($ResolvedValidatorCounts.Count -eq 1 -and $ValidatorCount -eq 3) {
      $runOutputDir = Join-Path $RepoRoot ".localnet"
    } else {
      $runOutputDir = Join-Path $RepoRoot ".localnet-smoke-$ValidatorCount"
    }
  }

  Push-Location $RepoRoot
  try {
    & .\scripts\localnet\init.ps1 -OutputDir $runOutputDir -Binary $Binary -ValidatorCount $ValidatorCount
    if (!$SkipNegativeTests) {
      Assert-CommandFails -Name "reset refuses repository root" -Command { & .\scripts\localnet\reset.ps1 -OutputDir $RepoRoot }
      Assert-CommandFails -Name "start rejects mismatched validator count" -Command { & .\scripts\localnet\start.ps1 -OutputDir $runOutputDir -Binary $Binary -ValidatorCount ($ValidatorCount + 1) }
    }

    & .\scripts\localnet\start.ps1 -OutputDir $runOutputDir -Binary $Binary -ValidatorCount $ValidatorCount -CleanLogs
    $manifest = Read-LocalnetManifest -OutputDir $runOutputDir
    $node = [string]$manifest.nodes[0].rpc_url
    $height = Wait-ForHeight -Node $node -TargetHeight $MinHeight -TimeoutSeconds $TimeoutSeconds
    Write-Host "$ValidatorCount-validator localnet reached height $height"
    if ($manifest.nodes[0].app_metrics_url) {
      Assert-MetricsEndpoint -Url ([string]$manifest.nodes[0].app_metrics_url) -ExpectedNames @(
        "orbitalis_block_height",
        "orbitalis_block_processing_seconds",
        "orbitalis_localnet_health",
        "orbitalis_process_memory_bytes"
      )
    }

    if (!$SkipTxFlow) {
      Invoke-ModuleSmoke -Manifest $manifest -OutputDir $runOutputDir -TimeoutSeconds $TimeoutSeconds
    }

    & .\scripts\localnet\stop.ps1 -OutputDir $runOutputDir
    & .\scripts\localnet\start.ps1 -OutputDir $runOutputDir -Binary $Binary -ValidatorCount $ValidatorCount
    $restartHeight = Wait-ForHeight -Node $node -TargetHeight ($height + 1) -TimeoutSeconds $TimeoutSeconds
    Write-Host "$ValidatorCount-validator localnet restart preserved state and reached height $restartHeight"
  }
  finally {
    & .\scripts\localnet\stop.ps1 -OutputDir $runOutputDir
    Pop-Location
  }
}

foreach ($count in $ResolvedValidatorCounts) {
  Run-LocalnetSmoke -ValidatorCount $count
}
