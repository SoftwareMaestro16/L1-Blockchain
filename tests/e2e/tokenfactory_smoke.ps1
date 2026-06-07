param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [string]$FactorySubdenom = "gold",
  [string]$Fees = "1000000naet"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for contract-assets smoke"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"
$restBase = "http://127.0.0.1:$($node0Ports.REST)"

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
    EnableAPI      = $EnableAPI
    EnableGRPC     = $EnableGRPC
    EnableRPC      = $EnableRPC
  }
  foreach ($key in $Extra.Keys) {
    $args[$key] = $Extra[$key]
  }
  & $ScriptPath @args
}

function New-SignedTxArgs {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0"
  )

  return $ActionArgs + @(
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $rpcNode,
    "--output", "json"
  )
}

function Send-SignedTx {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [switch]$ExpectFailure,
    [string]$ExpectedLog = ""
  )

  return Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds `
    -ExpectFailure:$ExpectFailure `
    -ExpectedLog $ExpectedLog
}

function Get-FactoryDenomMetadata {
  param([string]$Denom)

  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "contract-assets", "denom", $Denom, "--node", $rpcNode, "--output", "json")
  return $result.metadata
}

function Get-BalanceAmount {
  param(
    [string]$Address,
    [string]$Denom
  )

  $balance = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $Denom -RPCPort $node0Ports.RPC
  if (-not $balance.amount) { return [int64]0 }
  return [int64]$balance.amount
}

function Get-SupplyAmount {
  param([string]$Denom)

  $supply = Get-LocalnetBankSupplyOf -Binary $Binary -Denom $Denom -RPCPort $node0Ports.RPC
  if (-not $supply.amount) { return [int64]0 }
  return [int64]$supply.amount
}

function Assert-FactoryMetadata {
  param(
    [object]$Metadata,
    [string]$Denom,
    [string]$Admin
  )

  if ($Metadata.denom -ne $Denom) {
    throw "factory denom mismatch: expected $Denom, got $($Metadata.denom)"
  }
  if ($Metadata.admin -ne $Admin) {
    throw "factory admin mismatch: expected $Admin, got $($Metadata.admin)"
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
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet reached height $height"
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null

  if ($EnableAPI) {
    Wait-LocalnetRest -RESTPort $node0Ports.REST -TimeoutSeconds $TimeoutSeconds | Out-Null
  }

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
  $factoryDenom = "factory/$node0/$FactorySubdenom"

  $emptyDenoms = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "contract-assets", "denoms", "--node", $rpcNode, "--output", "json")
  if (@($emptyDenoms.denoms).Count -ne 0) {
    throw "fresh localnet contract-assets denoms must be empty"
  }

  $createTx = Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom) -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $createTx -Type "contract-assets_create_denom" -Attributes @{
    denom   = $factoryDenom
    creator = $node0
    admin   = $node0
  } | Out-Null
  $metadata = Get-FactoryDenomMetadata -Denom $factoryDenom
  Assert-FactoryMetadata -Metadata $metadata -Denom $factoryDenom -Admin $node0
  Write-Host "factory denom created with node0 admin"

  if ($EnableAPI) {
    $rest = Invoke-RestMethod -Uri "$restBase/l1/contract-assets/v1/denom/$factoryDenom" -TimeoutSec 5
    Assert-FactoryMetadata -Metadata $rest.metadata -Denom $factoryDenom -Admin $node0
    Write-Host "REST denom query returned matching admin"
  }

  $bankMetadata = Get-LocalnetBankMetadata -Binary $Binary -Denom $factoryDenom -RPCPort $node0Ports.RPC
  if ($bankMetadata.display -eq "AET" -or $bankMetadata.symbol -eq "AET" -or $bankMetadata.base -eq "naet") {
    throw "factory metadata must not spoof native AET/naet"
  }

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom) -FromHome $node0Home -ExpectFailure -ExpectedLog "denom already exists" | Out-Null
  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", "naet") -FromHome $node0Home -ExpectFailure -ExpectedLog "native AET/naet" | Out-Null
  Write-Host "duplicate and native-spoof denom creation rejected"

  $mintTx = Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "1000000$factoryDenom", $node0) -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $mintTx -Type "contract-assets_mint" -Attributes @{
    denom           = $factoryDenom
    sender          = $node0
    amount          = "1000000"
    mint_to_address = $node0
  } | Out-Null
  if ((Get-BalanceAmount -Address $node0 -Denom $factoryDenom) -ne 1000000) {
    throw "node0 factory balance must be 1000000 after mint"
  }
  if ((Get-SupplyAmount -Denom $factoryDenom) -ne 1000000) {
    throw "factory supply must be 1000000 after mint"
  }
  Write-Host "mint updated node0 balance and bank supply"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "burn", "1$factoryDenom", $node1) -FromHome $node0Home -ExpectFailure -ExpectedLog "burn_from_address must match sender" | Out-Null
  $burnTx = Send-SignedTx -ActionArgs @("tx", "contract-assets", "burn", "250000$factoryDenom", $node0) -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $burnTx -Type "contract-assets_burn" -Attributes @{
    denom             = $factoryDenom
    sender            = $node0
    amount            = "250000"
    burn_from_address = $node0
  } | Out-Null
  if ((Get-BalanceAmount -Address $node0 -Denom $factoryDenom) -ne 750000) {
    throw "node0 factory balance must be 750000 after burn"
  }
  if ((Get-SupplyAmount -Denom $factoryDenom) -ne 750000) {
    throw "factory supply must be 750000 after burn"
  }
  Write-Host "burn reduced node0 balance and bank supply"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "change-admin", $factoryDenom, "not-an-address") -FromHome $node0Home -ExpectFailure | Out-Null
  $changeAdminTx = Send-SignedTx -ActionArgs @("tx", "contract-assets", "change-admin", $factoryDenom, $node1) -FromHome $node0Home
  Assert-LocalnetTxEvent -Tx $changeAdminTx -Type "contract-assets_change_admin" -Attributes @{
    denom     = $factoryDenom
    sender    = $node0
    new_admin = $node1
  } | Out-Null
  $metadata = Get-FactoryDenomMetadata -Denom $factoryDenom
  Assert-FactoryMetadata -Metadata $metadata -Denom $factoryDenom -Admin $node1
  Write-Host "admin transferred to node1"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "1$factoryDenom", $node0) -FromHome $node0Home -ExpectFailure -ExpectedLog "only denom admin can mint" | Out-Null
  $newAdminMintTx = Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "100$factoryDenom", $node1) -FromHome $node1Home -FromKey "node1"
  Assert-LocalnetTxEvent -Tx $newAdminMintTx -Type "contract-assets_mint" -Attributes @{
    denom           = $factoryDenom
    sender          = $node1
    amount          = "100"
    mint_to_address = $node1
  } | Out-Null
  if ((Get-BalanceAmount -Address $node1 -Denom $factoryDenom) -ne 100) {
    throw "new admin node1 must receive 100 factory tokens"
  }
  if ((Get-SupplyAmount -Denom $factoryDenom) -ne 750100) {
    throw "factory supply must be 750100 after new admin mint"
  }
  Write-Host "old admin rejected and new admin minted successfully"

  $height = Wait-LocalnetHeight -TargetHeight ([int64]$height + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "contract-assets smoke flow completed at height $height"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
