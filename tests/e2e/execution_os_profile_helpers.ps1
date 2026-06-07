$ErrorActionPreference = "Stop"

function Invoke-ExecutionOSProfileLocalnet {
  param(
    [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
    [string]$Profile,
    [string]$OutputDir = "",
    [string]$Binary = "",
    [string]$ChainId = "aetra-local-1",
    [int]$ValidatorCount = 3,
    [int]$MinHeight = 3,
    [int]$TimeoutSeconds = 120,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseRESTPort = 1317,
    [int]$BaseGRPCPort = 9090,
    [int]$BasePprofPort = 6060,
    [int]$PortStride = 100,
    [switch]$SkipBuild
  )

  $RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
  . (Join-Path $RepoRoot "scripts\localnet\common.ps1")

  $OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  $Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
  Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "execution OS smoke output directory"

  $node0Ports = Get-LocalnetPortProfile `
    -Index 0 `
    -BaseP2PPort $BaseP2PPort `
    -BaseRPCPort $BaseRPCPort `
    -BaseRESTPort $BaseRESTPort `
    -BaseGRPCPort $BaseGRPCPort `
    -BasePprofPort $BasePprofPort `
    -PortStride $PortStride

  Push-Location $RepoRoot
  try {
    & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir

    if (-not $SkipBuild) {
      & .\scripts\build-aetrad.ps1 -Binary $Binary
    } elseif (-not (Test-Path -LiteralPath $Binary)) {
      throw "Binary not found at $Binary and -SkipBuild was specified"
    }

    & .\scripts\localnet\init.ps1 `
      -OutputDir $OutputDir `
      -Binary $Binary `
      -ChainId $ChainId `
      -ValidatorCount $ValidatorCount `
      -BaseP2PPort $BaseP2PPort `
      -BaseRPCPort $BaseRPCPort `
      -BaseRESTPort $BaseRESTPort `
      -BaseGRPCPort $BaseGRPCPort `
      -BasePprofPort $BasePprofPort `
      -PortStride $PortStride `
      -Profile $Profile `
      -SkipBuild

    & .\scripts\localnet\validate-genesis.ps1 `
      -OutputDir $OutputDir `
      -Binary $Binary `
      -ChainId $ChainId `
      -ValidatorCount $ValidatorCount `
      -BaseP2PPort $BaseP2PPort `
      -BaseRPCPort $BaseRPCPort `
      -BaseRESTPort $BaseRESTPort `
      -BaseGRPCPort $BaseGRPCPort `
      -BasePprofPort $BasePprofPort `
      -PortStride $PortStride `
      -Profile $Profile

    & .\scripts\localnet\start.ps1 `
      -OutputDir $OutputDir `
      -Binary $Binary `
      -ChainId $ChainId `
      -ValidatorCount $ValidatorCount `
      -BaseP2PPort $BaseP2PPort `
      -BaseRPCPort $BaseRPCPort `
      -BaseRESTPort $BaseRESTPort `
      -BaseGRPCPort $BaseGRPCPort `
      -BasePprofPort $BasePprofPort `
      -PortStride $PortStride `
      -Profile $Profile `
      -NoInit

    Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null

    $diagnostics = & .\scripts\localnet\execution-os-diagnostics.ps1 `
      -OutputDir $OutputDir `
      -Binary $Binary `
      -Profile $Profile `
      -Json | ConvertFrom-Json

    & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir

    & .\scripts\localnet\start.ps1 `
      -OutputDir $OutputDir `
      -Binary $Binary `
      -ChainId $ChainId `
      -ValidatorCount $ValidatorCount `
      -BaseP2PPort $BaseP2PPort `
      -BaseRPCPort $BaseRPCPort `
      -BaseRESTPort $BaseRESTPort `
      -BaseGRPCPort $BaseGRPCPort `
      -BasePprofPort $BasePprofPort `
      -PortStride $PortStride `
      -Profile $Profile `
      -NoInit

    Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Wait-LocalnetHeight -TargetHeight ($MinHeight + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null

    $restartDiagnostics = & .\scripts\localnet\execution-os-diagnostics.ps1 `
      -OutputDir $OutputDir `
      -Binary $Binary `
      -Profile $Profile `
      -Json | ConvertFrom-Json

    return [pscustomobject]@{
      Diagnostics        = $diagnostics
      RestartDiagnostics = $restartDiagnostics
      OutputDir          = $OutputDir
      RPCPort            = $node0Ports.RPC
    }
  } finally {
    & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
    Pop-Location
  }
}

function Invoke-ExecutionOSSmokeReport {
  param(
    [string]$Binary,
    [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
    [string]$Profile
  )

  return (& $Binary execution-os smoke --profile $Profile | ConvertFrom-Json)
}

function Assert-ExecutionOSGateEnabled {
  param(
    [object]$Diagnostics,
    [string]$Module
  )

  $gate = $Diagnostics.feature_gates.$Module
  if ($null -eq $gate -or $gate.enabled -ne $true -or $gate.testnet_profile -ne $true) {
    throw "$Module feature gate is not enabled as a testnet profile"
  }
}

function Assert-ExecutionOSRestartStable {
  param(
    [object]$Before,
    [object]$After,
    [string[]]$Fields
  )

  foreach ($field in $Fields) {
    $left = $Before.$field | ConvertTo-Json -Depth 100 -Compress
    $right = $After.$field | ConvertTo-Json -Depth 100 -Compress
    if ($left -ne $right) {
      throw "restart diagnostics field $field changed: before=$left after=$right"
    }
  }
}
