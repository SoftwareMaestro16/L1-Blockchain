param(
  [string]$OutputDir = "",
  [int]$ValidatorCount = 0,
  [int]$TimeoutSeconds = 60,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [int]$LogTailLines = 20,
  [switch]$Json
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"

$nodes = Get-LocalnetNodes -OutputDir $OutputDir
if ($nodes.Count -lt 1) {
  throw "No localnet node directories found under $OutputDir"
}
if ($ValidatorCount -gt 0 -and $nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount localnet nodes, found $($nodes.Count)"
}

$actualValidatorCount = if ($ValidatorCount -gt 0) { $ValidatorCount } else { $nodes.Count }
$p = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$summary = [ordered]@{
  output_dir       = $OutputDir
  validator_count  = $actualValidatorCount
  rpc_port         = $p.RPC
  rest_port        = $p.REST
  grpc_port        = $p.GRPC
  rpc              = "disabled"
  height           = $null
  height_increase  = $null
  validators       = "unchecked"
  peers            = "unchecked"
  rest             = "disabled"
  grpc             = "disabled"
  processes        = @()
  telemetry        = @()
  recent_logs      = @()
}

foreach ($node in $nodes) {
  $nodeHome = Join-Path $node.FullName "aetrad"
  $summary.telemetry += [ordered]@{
    node      = $node.Name
    telemetry = Get-LocalnetNodeTelemetry -NodeHome $nodeHome
  }
}

$summary.processes = @(Get-LocalnetProcessSnapshot -OutputDir $OutputDir)
$summary.recent_logs = @(Get-LocalnetRecentLogs -OutputDir $OutputDir -TailLines $LogTailLines)

if ($EnableRPC) {
  $status = Wait-LocalnetRpc -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds
  $summary.rpc = "ok"
  $summary.height = [int64]$status.result.sync_info.latest_block_height

  $heightIncrease = Wait-LocalnetHeightIncreasing -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds
  $summary.height_increase = "$($heightIncrease.StartHeight)->$($heightIncrease.CurrentHeight)"

  Wait-LocalnetValidators -ExpectedCount $actualValidatorCount -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $summary.validators = "ok"

  if ($actualValidatorCount -gt 1) {
    Wait-LocalnetPeers -ExpectedMinPeers 1 -RPCPort $p.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    $summary.peers = "ok"
  } else {
    $summary.peers = "single-node"
  }
}

if ($EnableAPI) {
  Wait-LocalnetRest -RESTPort $p.REST -TimeoutSeconds $TimeoutSeconds | Out-Null
  $summary.rest = "ok"
}

if ($EnableGRPC) {
  Wait-LocalnetGrpc -GRPCPort $p.GRPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $summary.grpc = "ok"
}

if ($Json) {
  $summary | ConvertTo-Json -Depth 5
} else {
  foreach ($item in $summary.GetEnumerator()) {
    if ($item.Value -is [array]) {
      Write-Host "$($item.Key): $($item.Value.Count) item(s)"
    } else {
      Write-Host "$($item.Key): $($item.Value)"
    }
  }
}
