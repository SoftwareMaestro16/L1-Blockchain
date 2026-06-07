param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 35656,
  [int]$BaseRPCPort = 35657,
  [int]$BaseRESTPort = 2217,
  [int]$BaseGRPCPort = 9890,
  [int]$BasePprofPort = 6960,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $PSScriptRoot "execution_os_profile_helpers.ps1")
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$result = Invoke-ExecutionOSProfileLocalnet `
  -Profile "mesh-prototype" `
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

$report = Invoke-ExecutionOSSmokeReport -Binary $Binary -Profile "mesh-prototype"
if ($report.mesh.receipt_status -ne "SUCCESS") { throw "mesh cross-zone message did not produce success receipt" }
if ($report.mesh.replay_marker_count -ne 1) { throw "mesh replay marker was not written in simulator" }
Assert-ExecutionOSGateEnabled -Diagnostics $result.Diagnostics -Module "mesh"
Assert-ExecutionOSRestartStable -Before $result.Diagnostics -After $result.RestartDiagnostics -Fields @("feature_gates", "replay_marker_count", "mesh_receipt_count")

Write-Host "mesh prototype smoke passed"
