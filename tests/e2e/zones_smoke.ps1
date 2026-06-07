param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 34656,
  [int]$BaseRPCPort = 34657,
  [int]$BaseRESTPort = 2117,
  [int]$BaseGRPCPort = 9790,
  [int]$BasePprofPort = 6860,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $PSScriptRoot "execution_os_profile_helpers.ps1")
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$result = Invoke-ExecutionOSProfileLocalnet `
  -Profile "zones-prototype" `
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

Assert-ExecutionOSGateEnabled -Diagnostics $result.Diagnostics -Module "zones"
if (@($result.Diagnostics.active_zones).Count -lt 4) { throw "zones profile did not activate all execution zones" }
Assert-ExecutionOSRestartStable -Before $result.Diagnostics -After $result.RestartDiagnostics -Fields @("feature_gates", "active_zones", "zone_commitment_roots")

Write-Host "zones prototype smoke passed"
