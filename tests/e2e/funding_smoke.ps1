param(
  [string]$OutputDir = ".localnet-funding",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [int]$BaseP2PPort = 27656,
  [int]$BaseRPCPort = 27657,
  [int]$BaseRESTPort = 1417,
  [int]$BaseGRPCPort = 9190,
  [int]$BasePprofPort = 6160,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

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
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function Assert-ScriptFails {
  param(
    [scriptblock]$Script,
    [string]$ExpectedText
  )

  try {
    & $Script
  } catch {
    if ($_.Exception.Message -notmatch [regex]::Escape($ExpectedText)) {
      throw "failure did not contain '$ExpectedText': $($_.Exception.Message)"
    }
    return
  }
  throw "script succeeded but failure was expected"
}

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-funding"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride

Push-Location $RepoRoot
try {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true; EnableAPI = $true; EnableGRPC = $true; EnableRPC = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet reached height $height"

  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $node2Home = Join-Path $OutputDir "node2\aetrad"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
  $node2 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node2Home -KeyName "node2"

  $node1Before = Get-LocalnetBankBalance -Binary $Binary -Address $node1 -Denom "naet" -RPCPort $node0Ports.RPC
  $node2Before = Get-LocalnetBankBalance -Binary $Binary -Address $node2 -Denom "naet" -RPCPort $node0Ports.RPC

  $funding = & .\scripts\localnet\fund.ps1 `
    -OutputDir $OutputDir `
    -Binary $Binary `
    -ChainId $ChainId `
    -RPCPort $node0Ports.RPC `
    -Recipients @($node1) `
    -Transfers @("$node2=777naet") `
    -Amount "1234naet" `
    -Fees "1000000naet" `
    -TimeoutSeconds $TimeoutSeconds `
    -Json | ConvertFrom-Json

  if (@($funding).Count -ne 2) {
    throw "funding flow must return two transfer results"
  }

  $node1After = Get-LocalnetBankBalance -Binary $Binary -Address $node1 -Denom "naet" -RPCPort $node0Ports.RPC
  $node2After = Get-LocalnetBankBalance -Binary $Binary -Address $node2 -Denom "naet" -RPCPort $node0Ports.RPC
  if ([int64]$node1After.amount -ne ([int64]$node1Before.amount + 1234)) {
    throw "node1 funding mismatch"
  }
  if ([int64]$node2After.amount -ne ([int64]$node2Before.amount + 777)) {
    throw "node2 funding mismatch"
  }
  Write-Host "local funding increased node1 and node2 balances"

  Assert-ScriptFails -ExpectedText "non-local chain-id" -Script {
    & .\scripts\localnet\fund.ps1 -OutputDir $OutputDir -Binary $Binary -ChainId "aetra-main-1" -RPCPort $node0Ports.RPC -Recipients @($node1) -Amount "1naet"
  }
  Assert-ScriptFails -ExpectedText "RPC network $ChainId does not match requested local chain-id aetra-local-2" -Script {
    & .\scripts\localnet\fund.ps1 -OutputDir $OutputDir -Binary $Binary -ChainId "aetra-local-2" -RPCPort $node0Ports.RPC -Recipients @($node1) -Amount "1naet"
  }
  Assert-ScriptFails -ExpectedText "key not found" -Script {
    & .\scripts\localnet\fund.ps1 -OutputDir $OutputDir -Binary $Binary -ChainId $ChainId -RPCPort $node0Ports.RPC -FromKey "missing-funder" -Recipients @($node1) -Amount "1naet"
  }

  Write-Host "local funding safety failures passed"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
