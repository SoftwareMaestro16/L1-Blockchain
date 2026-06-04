param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "orbitalis-local-1",
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
  [string]$DelegationAmount = "5000000norb",
  [string]$Fees = "1000000norb"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\orbitalisd.exe"
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

function Assert-PositiveDecimalString {
  param(
    [object]$Value,
    [string]$Name
  )

  try {
    $number = [decimal]([string]$Value)
  } catch {
    throw "$Name is not a decimal value: $Value"
  }
  if ($number -le 0) {
    throw "$Name must be positive, got $Value"
  }
}

function Assert-BondedStatus {
  param([object]$Validator)

  $status = [string]$Validator.status
  if ($status -ne "BOND_STATUS_BONDED" -and $status -ne "3") {
    throw "validator $($Validator.operator_address) is not bonded: $status"
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
  Write-Host "CometBFT validator set contains $ValidatorCount validators"

  if ($ValidatorCount -gt 1) {
    Wait-LocalnetPeers -ExpectedMinPeers ($ValidatorCount - 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "node0 has at least $($ValidatorCount - 1) peers"
  }

  $stakingParams = Get-LocalnetStakingParams -Binary $Binary -RPCPort $node0Ports.RPC
  if ($stakingParams.bond_denom -ne "norb") {
    throw "staking bond denom must be norb, got $($stakingParams.bond_denom)"
  }
  Write-Host "staking params use bond denom norb"

  $validators = @(Get-LocalnetStakingValidators -Binary $Binary -RPCPort $node0Ports.RPC)
  if ($validators.Count -ne $ValidatorCount) {
    throw "staking validators query returned $($validators.Count), expected $ValidatorCount"
  }
  foreach ($validator in $validators) {
    Assert-BondedStatus -Validator $validator
  }

  $selectedValidator = @($validators | Sort-Object -Property operator_address | Select-Object -First 1)[0]
  $validatorAddress = $selectedValidator.operator_address
  if ([string]::IsNullOrWhiteSpace($validatorAddress)) {
    throw "selected validator is missing operator_address"
  }
  Write-Host "selected bonded validator $validatorAddress"

  Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "staking", "validator", $validatorAddress, "--node", $rpcNode, "--output", "json") | Out-Null
  Write-Host "staking validator query succeeded"

  $slashingParams = Get-LocalnetSlashingParams -Binary $Binary -RPCPort $node0Ports.RPC
  if ([int64]$slashingParams.signed_blocks_window -le 0) {
    throw "signed_blocks_window must be positive, got $($slashingParams.signed_blocks_window)"
  }
  Assert-PositiveDecimalString -Name "min_signed_per_window" -Value $slashingParams.min_signed_per_window
  Assert-PositiveDecimalString -Name "slash_fraction_double_sign" -Value $slashingParams.slash_fraction_double_sign
  Assert-PositiveDecimalString -Name "slash_fraction_downtime" -Value $slashingParams.slash_fraction_downtime
  Write-Host "slashing params are positive"

  Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "slashing signing infos" -Condition {
    $infos = @(Get-LocalnetSigningInfos -Binary $Binary -RPCPort $node0Ports.RPC)
    if ($infos.Count -ge $ValidatorCount) { return $infos }
    return $null
  } | Out-Null
  Write-Host "slashing signing-infos query returned validator records"

  $node0Home = Join-Path $OutputDir "node0\orbitalisd"
  $delegator = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $beforePower = Get-LocalnetTotalVotingPower -RPCPort $node0Ports.RPC
  Send-LocalnetDelegateTx `
    -Binary $Binary `
    -FromHome $node0Home `
    -FromKey "node0" `
    -ValidatorAddress $validatorAddress `
    -Amount $DelegationAmount `
    -Fees $Fees `
    -ChainId $ChainId `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "delegation tx accepted from node0"

  $height = Wait-LocalnetHeight -TargetHeight ([int64]$height + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  $delegation = Get-LocalnetDelegation -Binary $Binary -DelegatorAddress $delegator -ValidatorAddress $validatorAddress -RPCPort $node0Ports.RPC
  $balance = $delegation.delegation_response.balance
  if ($null -eq $balance) {
    $balance = $delegation.balance
  }
  if ($balance.denom -ne "norb") {
    throw "delegation balance denom must be norb, got $($balance.denom)"
  }
  if ([int64]$balance.amount -lt 5000000) {
    throw "delegation balance amount is too small: $($balance.amount)"
  }
  Write-Host "delegation query shows $($balance.amount)$($balance.denom)"

  $afterPower = Wait-LocalnetTotalVotingPowerGreater -PreviousPower $beforePower -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "validator voting power increased from $beforePower to $afterPower"

  $height = Wait-LocalnetHeight -TargetHeight ([int64]$height + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "PoS smoke flow completed at height $height"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
