param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 36656,
  [int]$BaseRPCPort = 36657,
  [int]$BaseRESTPort = 2317,
  [int]$BaseGRPCPort = 9990,
  [int]$BasePprofPort = 7060,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $PSScriptRoot "execution_os_profile_helpers.ps1")
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$result = Invoke-ExecutionOSProfileLocalnet `
  -Profile "identity-prototype" `
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

$report = Invoke-ExecutionOSSmokeReport -Binary $Binary -Profile "identity-prototype"
if ($report.identity.domain -ne "operator.aet") { throw "identity smoke did not register operator.aet" }
if ($report.identity.lifecycle -ne "ACTIVE") { throw "identity smoke domain is not active" }
if ($result.Diagnostics.identity_prototype -ne $true) { throw "identity prototype diagnostics flag is missing" }
Assert-ExecutionOSGateEnabled -Diagnostics $result.Diagnostics -Module "zones"
Assert-ExecutionOSRestartStable -Before $result.Diagnostics -After $result.RestartDiagnostics -Fields @("feature_gates", "identity_prototype", "active_zones")

Write-Host "identity prototype smoke passed"
