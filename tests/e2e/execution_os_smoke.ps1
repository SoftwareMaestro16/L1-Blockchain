param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 33656,
  [int]$BaseRPCPort = 33657,
  [int]$BaseRESTPort = 2017,
  [int]$BaseGRPCPort = 9690,
  [int]$BasePprofPort = 6760,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $PSScriptRoot "execution_os_profile_helpers.ps1")
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$result = Invoke-ExecutionOSProfileLocalnet `
  -Profile "execution-os-sim" `
  -OutputDir $OutputDir `
  -Binary $Binary `
  -ValidatorCount $ValidatorCount `
  -MinHeight $MinHeight `
  -TimeoutSeconds $TimeoutSeconds `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BasePprofPort $BasePprofPort `
  -SkipBuild:$SkipBuild

$report = Invoke-ExecutionOSSmokeReport -Binary $Binary -Profile "execution-os-sim"
if ($report.load.score_bps -le 0) { throw "load score update did not run" }
if ($report.routing.zone_id -ne "FINANCIAL_ZONE") { throw "route tx to zone failed" }
if ($report.sharding.active_shard_count -lt 2) { throw "load did not activate shards" }
if ($report.mesh.receipt_status -ne "SUCCESS") { throw "cross-zone message receipt failed" }
if ($report.identity.domain -ne "operator.aet") { throw ".aet registration smoke failed" }

Assert-ExecutionOSGateEnabled -Diagnostics $result.Diagnostics -Module "load"
Assert-ExecutionOSGateEnabled -Diagnostics $result.Diagnostics -Module "routing"
Assert-ExecutionOSRestartStable -Before $result.Diagnostics -After $result.RestartDiagnostics -Fields @("feature_gates", "active_shards", "active_zones")

Write-Host "execution OS smoke passed"
