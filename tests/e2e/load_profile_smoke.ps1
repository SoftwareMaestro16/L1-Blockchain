param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [int]$BaseP2PPort = 30656,
  [int]$BaseRPCPort = 30657,
  [int]$BaseRESTPort = 1717,
  [int]$BaseGRPCPort = 9490,
  [int]$BasePprofPort = 6460,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-load-profile"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride

function Invoke-WithCommonLocalnetArgs {
  param(
    [string]$ScriptPath,
    [hashtable]$Extra = @{}
  )

  $args = @{
    OutputDir      = $OutputDir
    Binary         = $Binary
    ChainId        = $ChainId
    ValidatorCount = $ValidatorCount
    BaseP2PPort    = $BaseP2PPort
    BaseRPCPort    = $BaseRPCPort
    BaseRESTPort   = $BaseRESTPort
    BaseGRPCPort   = $BaseGRPCPort
    BasePprofPort  = $BasePprofPort
    PortStride     = $PortStride
    TimeoutCommit  = $TimeoutCommit
    LogLevel       = $LogLevel
    EnableAPI      = $true
    EnableGRPC     = $true
    EnableRPC      = $true
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function Assert-True {
  param(
    [bool]$Condition,
    [string]$Message
  )

  if (-not $Condition) {
    throw $Message
  }
}

Push-Location $RepoRoot
try {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  $initExtra = @{}
  if (Test-Path -LiteralPath $Binary) {
    $initExtra.SkipBuild = $true
  }
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1" -Extra $initExtra
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null

  $summaryText = & .\scripts\localnet\load-profile.ps1 `
    -OutputDir $OutputDir `
    -Binary $Binary `
    -ChainId $ChainId `
    -RPCPort $node0Ports.RPC `
    -Scenario mixed `
    -Count 3 `
    -RatePerSecond 3 `
    -TimeoutSeconds $TimeoutSeconds `
    -FactorySubdenom "smokeload" `
    -Json
  if ($LASTEXITCODE -ne 0) {
    throw "load-profile exited with $LASTEXITCODE"
  }

  $summary = ($summaryText -join "`n") | ConvertFrom-Json
  Assert-True ($summary.profile -eq "local-minimal-load") "unexpected load profile name"
  Assert-True ($summary.scenario -eq "mixed") "unexpected load scenario"
  Assert-True ([int]$summary.count -eq 3) "load count mismatch"
  Assert-True ([int]$summary.successes -eq 3) "tiny load profile must complete all txs"
  Assert-True ([int]$summary.failures -eq 0) "tiny load profile must record zero failures"
  Assert-True ([int]$summary.blocks_progressed -ge 1) "load profile must observe block progress"
  Assert-True ([int64]$summary.latency_ms.max -gt 0) "load profile must record tx latency"
  Assert-True ($summary.operations.bank_send.success -eq 1) "mixed profile must include one bank send"
  Assert-True ($summary.operations.contract-assets_mint.success -eq 1) "mixed profile must include one contract-assets mint"
  Assert-True ($summary.operations.dex_swap.success -eq 1) "mixed profile must include one DEX swap"

  Write-Host "load profile smoke passed: successes=$($summary.successes) latency_avg_ms=$($summary.latency_ms.avg) blocks=$($summary.blocks_progressed)"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
