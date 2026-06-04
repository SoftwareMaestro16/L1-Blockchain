param(
  [int]$SpamCount = 5,
  [int]$TimeoutSeconds = 90,
  [string]$OutputDir = "",
  [string]$Binary = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")
$Binary = Resolve-BinaryPath -Binary $Binary
if ([string]::IsNullOrWhiteSpace($OutputDir)) {
  $OutputDir = Join-Path $RepoRoot ".localnet-adversarial"
}
$OutputDir = Resolve-LocalnetPath -OutputDir $OutputDir

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

function Invoke-OrbitalisRaw {
  param([Parameter(Mandatory = $true)][string[]]$Arguments)

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  }
  finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  return [pscustomobject]@{ ExitCode = $exitCode; Output = $output }
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

function Wait-ForTxResult {
  param(
    [Parameter(Mandatory = $true)][string]$Node,
    [Parameter(Mandatory = $true)][string]$TxHash,
    [int]$TimeoutSeconds = 45
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    Start-Sleep -Seconds 2
    try {
      $query = Invoke-OrbitalisJson -Arguments @("query", "tx", $TxHash, "--node", $Node, "--output", "json")
      if ($query.tx_response) {
        return $query.tx_response
      }
      return $query
    }
    catch {
      continue
    }
  }
  throw "transaction $TxHash was not found within $TimeoutSeconds seconds"
}

function Get-KeyAddress {
  param(
    [Parameter(Mandatory = $true)][string]$NodeHome,
    [Parameter(Mandatory = $true)][string]$Name
  )

  $output = Invoke-ExternalChecked -FilePath $Binary -Arguments @("keys", "show", $Name, "-a", "--home", $NodeHome, "--keyring-backend", "test") -FailureMessage "key lookup failed"
  return (($output | Select-Object -Last 1) -as [string]).Trim()
}

function Assert-TxRejected {
  param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][string[]]$Arguments,
    [string]$Node = "",
    [int]$TimeoutSeconds = 45
  )

  $raw = Invoke-OrbitalisRaw -Arguments $Arguments
  if ($raw.ExitCode -ne 0) {
    Write-Host "adversarial tx rejected: $Name"
    return
  }
  $json = Get-CliJson -Output $raw.Output
  if ($json.code -and [int]$json.code -ne 0) {
    Write-Host "adversarial tx rejected: $Name code=$($json.code)"
    return
  }
  $txHash = ""
  if ($json.txhash) {
    $txHash = [string]$json.txhash
  } elseif ($json.tx_response -and $json.tx_response.txhash) {
    $txHash = [string]$json.tx_response.txhash
  }
  if (-not [string]::IsNullOrWhiteSpace($Node) -and -not [string]::IsNullOrWhiteSpace($txHash)) {
    $txResult = Wait-ForTxResult -Node $Node -TxHash $txHash -TimeoutSeconds $TimeoutSeconds
    if ($txResult.code -and [int]$txResult.code -ne 0) {
      Write-Host "adversarial tx rejected after delivery: $Name code=$($txResult.code)"
      return
    }
  }
  throw "adversarial tx unexpectedly succeeded: $Name"
}

Push-Location $RepoRoot
try {
  & .\scripts\localnet\init.ps1 -OutputDir $OutputDir -Binary $Binary -ValidatorCount 3
  & .\scripts\localnet\start.ps1 -OutputDir $OutputDir -Binary $Binary -ValidatorCount 3 -CleanLogs

  $manifest = Read-LocalnetManifest -OutputDir $OutputDir
  $node = [string]$manifest.nodes[0].rpc_url
  $home0 = Get-NodeHome -OutputDir $OutputDir -Index 0
  $home1 = Get-NodeHome -OutputDir $OutputDir -Index 1
  $chainId = [string]$manifest.chain_id
  $addr1 = Get-KeyAddress -NodeHome $home1 -Name "node1"
  $height = Wait-ForHeight -Node $node -TargetHeight 3 -TimeoutSeconds $TimeoutSeconds

  Assert-TxRejected -Name "malformed broadcast bytes" -Arguments @("tx", "broadcast", "not-a-protobuf-tx", "--node", $node, "--output", "json")

  for ($i = 0; $i -lt $SpamCount; $i++) {
    Assert-TxRejected -Name "wrong fee denom spam $i" -Arguments @(
      "tx", "bank", "send", "node0", $addr1, "1norb",
      "--home", $home0,
      "--chain-id", $chainId,
      "--keyring-backend", "test",
      "--node", $node,
      "--fees", "1uatom",
      "--gas", "200000",
      "--broadcast-mode", "sync",
      "--output", "json",
      "--yes"
    )
  }

  Assert-TxRejected -Name "DEX same-denom pool manipulation" -Arguments @(
    "tx", "dex", "create-pool", "10norb", "10norb",
    "--home", $home0,
    "--from", "node0",
    "--chain-id", $chainId,
    "--keyring-backend", "test",
    "--node", $node,
    "--fees", $script:DefaultFee,
    "--gas", "auto",
    "--gas-adjustment", "1.2",
    "--broadcast-mode", "sync",
    "--output", "json",
    "--yes"
  )

  $nextHeight = Wait-ForHeight -Node $node -TargetHeight ($height + 3) -TimeoutSeconds $TimeoutSeconds
  Write-Host "adversarial localnet smoke passed; height advanced from $height to $nextHeight"
}
finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
