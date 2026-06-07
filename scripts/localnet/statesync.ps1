param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$TargetNodeIndex = 0,
  [int]$TrustHeight = 0,
  [string]$TrustHash = "",
  [switch]$ResetData
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$manifest = Read-LocalnetManifest -OutputDir $OutputDir
if ($null -eq $manifest) {
  throw "localnet manifest not found in $OutputDir"
}
$validatorCount = [int]$manifest.validator_count
if ($validatorCount -lt 2) {
  throw "state sync requires at least two validators"
}
if ($TargetNodeIndex -lt 0 -or $TargetNodeIndex -ge $validatorCount) {
  throw "target node index $TargetNodeIndex out of range for $validatorCount validators"
}

$rpcServers = @()
for ($i = 0; $i -lt $validatorCount -and $rpcServers.Count -lt 2; $i++) {
  if ($i -eq $TargetNodeIndex) {
    continue
  }
  $rpcServers += $manifest.nodes[$i].rpc_url
}

if ($TrustHeight -le 0 -or [string]::IsNullOrWhiteSpace($TrustHash)) {
  $statusJson = Invoke-ExternalChecked -FilePath $Binary -Arguments @("status", "--node", $rpcServers[0], "--output", "json") -FailureMessage "status query failed"
  $status = ($statusJson -join "`n" | ConvertFrom-Json)
  $syncInfo = $status.sync_info
  if ($null -eq $syncInfo) {
    $syncInfo = $status.SyncInfo
  }
  $TrustHeight = [int]$syncInfo.latest_block_height
  $TrustHash = [string]$syncInfo.latest_block_hash
  if ($TrustHeight -le 0 -or [string]::IsNullOrWhiteSpace($TrustHash)) {
    throw "could not derive trust height/hash from $($rpcServers[0])"
  }
}

$nodeHome = Get-NodeHome -OutputDir $OutputDir -Index $TargetNodeIndex
$configToml = Join-Path $nodeHome "config\config.toml"
if (!(Test-Path -LiteralPath $configToml)) {
  throw "target config.toml not found: $configToml"
}

if ($ResetData) {
  $dataDir = Join-Path $nodeHome "data"
  $fullNodeHome = (ConvertTo-AbsolutePath -Path $nodeHome).TrimEnd('\', '/')
  $fullDataDir = (ConvertTo-AbsolutePath -Path $dataDir).TrimEnd('\', '/')
  if (-not $fullDataDir.StartsWith($fullNodeHome, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "refusing to remove unexpected data path: $fullDataDir"
  }
  if (Test-Path -LiteralPath $dataDir) {
    Remove-Item -LiteralPath $dataDir -Recurse -Force
    Write-Host "Removed target data directory $dataDir"
  }
}

$config = Get-Content -Raw -LiteralPath $configToml
$config = $config -replace 'enable = false', 'enable = true'
$config = $config -replace 'rpc_servers = ".*"', "rpc_servers = `"$($rpcServers -join ',')`""
$config = $config -replace 'trust_height = \d+', "trust_height = $TrustHeight"
$config = $config -replace 'trust_hash = ".*"', "trust_hash = `"$TrustHash`""
$config = $config -replace 'trust_period = ".*"', 'trust_period = "168h0m0s"'
Set-Content -LiteralPath $configToml -Value $config

Write-Host "Configured node$TargetNodeIndex state sync"
Write-Host "rpc_servers=$($rpcServers -join ',')"
Write-Host "trust_height=$TrustHeight"
Write-Host "trust_hash=$TrustHash"
