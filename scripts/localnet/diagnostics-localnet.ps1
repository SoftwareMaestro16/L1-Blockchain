param(
  [string]$OutputDir = "",
  [string]$ChainId = "aetra-local-1",
  [int]$BaseRPCPort = 26657,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$PortStride = 100,
  [bool]$EnableRPC = $true,
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [int]$SampleHeight = 3,
  [switch]$Full,
  [switch]$IncludeLogs
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "diagnostics output directory"

$diagnosticDir = Join-Path $OutputDir "diagnostics"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$diagTimestampDir = Join-Path $diagnosticDir $timestamp
New-Item -ItemType Directory -Force -Path $diagTimestampDir | Out-Null

Write-Host "Collecting localnet diagnostics to: $diagTimestampDir"

# Get node count from manifest or count directories
$manifestPath = Join-Path $OutputDir "profile.json"
$validatorCount = 3
if (Test-Path $manifestPath) {
  $manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
  $validatorCount = $manifest.validator_count
}
$nodes = Get-LocalnetNodes -OutputDir $OutputDir
if ($nodes.Count -gt 0) {
  $validatorCount = $nodes.Count
}

# System info (no secrets)
$sysInfo = [ordered]@{
  collected_at_utc = (Get-Date).ToUniversalTime().ToString("o")
  chain_id = $ChainId
  validator_count = $validatorCount
  output_dir = $OutputDir
  host_name = $env:COMPUTERNAME
  user_name = $env:USERNAME
  ps_version = $PSVersionTable.PSVersion.ToString()
  platform = if ($IsWindows) { "windows" } elseif ($IsLinux) { "linux" } elseif ($IsMacOS) { "macos" } else { "unknown" }
  aetra_version = $null
  diag_version = "1.0.0"
}
Set-Content -LiteralPath (Join-Path $diagTimestampDir "system-info.json") -Value ($sysInfo | ConvertTo-Json -Depth 10)

# Query each node
$nodeResults = @()
for ($i = 0; $i -lt $validatorCount; $i++) {
  $p = Get-LocalnetPortProfile -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort 6060 -PortStride $PortStride
  $nodeName = "node$i"
  $nodeResult = [ordered]@{
    node_index = $i
    node_name = $nodeName
    rpc_port = $p.RPC
    rest_port = $p.REST
    grpc_port = $p.GRPC
    p2p_port = $p.P2P
    status = "unknown"
    block_height = $null
    catching_up = $null
    voting_power = $null
    validator_info = $null
    errors = @()
  }

  if ($EnableRPC) {
    try {
      $status = Invoke-RestMethod -Uri "http://localhost:$($p.RPC)/status" -TimeoutSec 5
      $nodeResult.status = "running"
      $nodeResult.block_height = $status.result.sync_info.latest_block_height
      $nodeResult.catching_up = $status.result.sync_info.catching_up
      $nodeResult.voting_power = $status.result.validator_info.voting_power
      $nodeResult.validator_info = @{
        address = $status.result.validator_info.address
        pub_key_type = $status.result.validator_info.pub_key.'@type'
      }
      if ($null -eq $sysInfo.aetra_version) {
        $sysInfo.aetra_version = $status.result.node_info.version
      }
    } catch {
      $nodeResult.status = "unreachable"
      $nodeResult.errors += $_.Exception.Message
    }
  }

  $nodeResults += $nodeResult
}

# Save node status
$nodeResults | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $diagTimestampDir "node-status.json")

# Query validator set
if ($EnableRPC -and $nodeResults[0].status -eq "running") {
  try {
    $validators = Invoke-RestMethod -Uri "http://localhost:$($nodeResults[0].rpc_port)/validators?height=$($nodeResults[0].block_height)" -TimeoutSec 10
    $validators | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $diagTimestampDir "validator-set.json")
  } catch {
    "Failed to query validators: $($_.Exception.Message)" | Set-Content -LiteralPath (Join-Path $diagTimestampDir "validator-set-error.txt")
  }
}

# Sample blocks if requested
if ($Full -and $EnableRPC -and $nodeResults[0].status -eq "running") {
  $blocksDir = Join-Path $diagTimestampDir "blocks"
  New-Item -ItemType Directory -Force -Path $blocksDir | Out-Null
  
  $startHeight = [int]$nodeResults[0].block_height - $SampleHeight
  if ($startHeight -lt 1) { $startHeight = 1 }
  
  for ($h = $startHeight; $h -le $nodeResults[0].block_height; $h++) {
    try {
      $block = Invoke-RestMethod -Uri "http://localhost:$($nodeResults[0].rpc_port)/block?height=$h" -TimeoutSec 5
      $block | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $blocksDir "block-$h.json")
    } catch {
      # Skip failed block queries
    }
  }
}

# Include logs if requested (no secrets - just stderr patterns)
if ($IncludeLogs) {
  $logDir = Join-Path $OutputDir "logs"
  if (Test-Path $logDir) {
    $logErrors = @()
    foreach ($log in Get-ChildItem -LiteralPath $logDir -Filter "*.log" -ErrorAction SilentlyContinue) {
      $content = Get-Content -LiteralPath $log.FullName -Raw
      # Look for error patterns (not secrets)
      $errorLines = $content -split "`n" | Where-Object { $_ -match "ERR|ERROR|PANIC|FAIL" }
      if ($errorLines.Count -gt 0) {
        $logErrors += [ordered]@{
          log_file = $log.Name
          error_count = $errorLines.Count
          errors = $errorLines | Select-Object -First 20
        }
      }
    }
    if ($logErrors.Count -gt 0) {
      $logErrors | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $diagTimestampDir "log-errors.json")
    }
  }
}

# Generate summary
$summary = [ordered]@{
  collected_at_utc = (Get-Date).ToUniversalTime().ToString("o")
  chain_id = $ChainId
  validator_count = $validatorCount
  running_nodes = ($nodeResults | Where-Object { $_.status -eq "running" }).Count
  unreachable_nodes = ($nodeResults | Where-Object { $_.status -eq "unreachable" }).Count
  block_height = $nodeResults[0].block_height
  catching_up = $nodeResults[0].catching_up
  aetra_version = $sysInfo.aetra_version
  diagnostics_dir = $diagTimestampDir
  note = "This diagnostics collection contains no secrets, mnemonics, private keys, or keyring material."
}
$summary | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $diagTimestampDir "summary.json")

Write-Host "Diagnostics collected to: $diagTimestampDir"
Write-Host "Summary: $($summary.running_nodes)/$($summary.validator_count) nodes running, block height=$($summary.block_height)"
Write-Host "No secrets, mnemonics, or private keys were collected."

return $summary