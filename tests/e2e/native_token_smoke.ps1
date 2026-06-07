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
  [string]$SendAmount = "1000naet",
  [string]$Fees = "1000000naet"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for native token transfer smoke"
}
if ($SendAmount -notmatch '^([0-9]+)naet$') {
  throw "SendAmount must be a naet coin string, got $SendAmount"
}
$SendAmountNorb = [int64]$Matches[1]

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$node0Ports = Get-LocalnetPortProfile -Index 0 -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
$rpcNode = "tcp://127.0.0.1:$($node0Ports.RPC)"

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

function Assert-NativeMetadata {
  param([object]$Metadata)

  if ($Metadata.base -ne "naet") {
    throw "metadata base must be naet, got $($Metadata.base)"
  }
  if ($Metadata.display -ne "AET") {
    throw "metadata display must be AET, got $($Metadata.display)"
  }
  if ($Metadata.symbol -ne "AET") {
    throw "metadata symbol must be AET, got $($Metadata.symbol)"
  }
  if ($Metadata.name -ne "Aetra") {
    throw "metadata name must be Aetra, got $($Metadata.name)"
  }

  $baseUnit = @($Metadata.denom_units | Where-Object { $_.denom -eq "naet" })
  if ($baseUnit.Count -ne 1 -or [int]$baseUnit[0].exponent -ne 0) {
    throw "metadata must include naet exponent 0"
  }
  $displayUnit = @($Metadata.denom_units | Where-Object { $_.denom -eq "AET" })
  if ($displayUnit.Count -ne 1 -or [int]$displayUnit[0].exponent -ne 9) {
    throw "metadata must include AET exponent 9"
  }
}

Push-Location $RepoRoot
try {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\init.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\validate-genesis.ps1"
  Invoke-WithCommonLocalnetArgs -ScriptPath ".\scripts\localnet\start.ps1" -Extra @{ NoInit = $true }

  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "localnet reached height $height"
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "validator set contains $ValidatorCount validators"

  $metadata = Get-LocalnetBankMetadata -Binary $Binary -Denom "naet" -RPCPort $node0Ports.RPC
  Assert-NativeMetadata -Metadata $metadata
  Write-Host "bank metadata exposes naet/AET with exponent 9"

  $supply = Get-LocalnetBankSupplyOf -Binary $Binary -Denom "naet" -RPCPort $node0Ports.RPC
  if ($supply.denom -ne "naet" -or [int64]$supply.amount -le 0) {
    throw "native supply must be positive naet, got $($supply.amount)$($supply.denom)"
  }
  Write-Host "native supply query returned $($supply.amount)$($supply.denom)"

  $stakingParams = Get-LocalnetStakingParams -Binary $Binary -RPCPort $node0Ports.RPC
  if ($stakingParams.bond_denom -ne "naet") {
    throw "staking bond denom must be naet, got $($stakingParams.bond_denom)"
  }

  $feesParams = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "fees", "params", "--node", $rpcNode, "--output", "json")
  $allowedFeeDenoms = @($feesParams.params.allowed_fee_denoms)
  if ($allowedFeeDenoms.Count -ne 1 -or $allowedFeeDenoms[0] -ne "naet") {
    throw "fees params must allow only naet, got $($allowedFeeDenoms -join ',')"
  }

  $mintParams = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "mint", "params", "--node", $rpcNode, "--output", "json")
  if ($mintParams.params.mint_denom -ne "naet") {
    throw "mint denom must be naet, got $($mintParams.params.mint_denom)"
  }
  Write-Host "staking, fees, and mint params all use naet"

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $sender = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $recipient = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"

  $senderBalanceBefore = Get-LocalnetBankBalance -Binary $Binary -Address $sender -Denom "naet" -RPCPort $node0Ports.RPC
  $recipientBalanceBefore = Get-LocalnetBankBalance -Binary $Binary -Address $recipient -Denom "naet" -RPCPort $node0Ports.RPC
  if ([int64]$senderBalanceBefore.amount -le 0 -or [int64]$recipientBalanceBefore.amount -le 0) {
    throw "node0 and node1 must start with positive naet balances"
  }

  Send-LocalnetBankTx `
    -Binary $Binary `
    -FromHome $node0Home `
    -FromKey "node0" `
    -ToAddress $recipient `
    -Amount $SendAmount `
    -Fees $Fees `
    -ChainId $ChainId `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds | Out-Null

  $height = Wait-LocalnetHeight -TargetHeight ([int64]$height + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  $recipientBalanceAfter = Get-LocalnetBankBalance -Binary $Binary -Address $recipient -Denom "naet" -RPCPort $node0Ports.RPC
  $expectedRecipient = [int64]$recipientBalanceBefore.amount + $SendAmountNorb
  if ([int64]$recipientBalanceAfter.amount -ne $expectedRecipient) {
    throw "recipient naet balance mismatch: expected $expectedRecipient, got $($recipientBalanceAfter.amount)"
  }

  Write-Host "bank send paid fees in naet and recipient balance increased to $($recipientBalanceAfter.amount)naet"
  Write-Host "native token smoke flow completed at height $height"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
