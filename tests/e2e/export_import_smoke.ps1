param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ExportDir = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [int]$BaseP2PPort = 28656,
  [int]$BaseRPCPort = 28657,
  [int]$BaseRESTPort = 1517,
  [int]$BaseGRPCPort = 9290,
  [int]$BasePprofPort = 6260,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [string]$FactorySubdenom = "exportgold",
  [string]$Fees = "1000naet"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for export/import smoke"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-export-import"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$ExportDir = Resolve-LocalnetPath -Path $ExportDir -DefaultRelativePath ".work\genesis\export-import"
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
    EnableAPI      = $true
    EnableGRPC     = $true
    EnableRPC      = $true
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
    [string]$FromKey = "node0"
  )

  $tx = Invoke-LocalnetCliJson -Binary $Binary -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey)
  $broadcastCode = Get-LocalnetTxCode -Tx $tx
  if ($broadcastCode -ne 0) {
    throw "broadcast failed with code $broadcastCode`: $(Get-LocalnetTxLog -Tx $tx)"
  }
  Wait-LocalnetHeightIncreasing -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  return $tx
}

function Invoke-QueryCliJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--node", $rpcNode, "--output", "json"))
}

function Get-LocalnetKeySDKAddress {
  param(
    [string]$NodeHome,
    [string]$KeyName
  )

  $output = & $Binary keys show $KeyName -a --home $NodeHome --keyring-backend test 2>&1
  if ($LASTEXITCODE -ne 0) {
    throw "failed to read SDK key address $KeyName from $NodeHome`n$($output -join "`n")"
  }
  return (($output | Select-Object -Last 1).ToString().Trim())
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

function Assert-CoinInBalance {
  param(
    [object]$Genesis,
    [string[]]$Addresses,
    [string]$Denom,
    [int64]$MinAmount
  )

  $entry = @($Genesis.app_state.bank.balances | Where-Object { $Addresses -contains $_.address } | Select-Object -First 1)
  Assert-True ($entry.Count -eq 1) "missing bank balance entry for $($Addresses -join ' or ')"
  $coin = @($entry[0].coins | Where-Object { $_.denom -eq $Denom } | Select-Object -First 1)
  Assert-True ($coin.Count -eq 1) "missing $Denom balance for $($entry[0].address) in exported genesis"
  Assert-True ([int64]$coin[0].amount -ge $MinAmount) "exported $Denom balance for $($entry[0].address) below $MinAmount"
}

function Assert-NativeCommandFails {
  param(
    [string[]]$Arguments,
    [string]$ExpectedText
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  if ($exitCode -eq 0) {
    throw "command succeeded but failure was expected: $Binary $($Arguments -join ' ')"
  }
  $text = $output -join "`n"
  if ($ExpectedText -and ($text -notmatch $ExpectedText)) {
    throw "command failed, but output did not match '$ExpectedText': $text"
  }
  $global:LASTEXITCODE = 0
}

Push-Location $RepoRoot
try {
  $goCache = Join-Path $RepoRoot ".work\gocache"
  New-Item -ItemType Directory -Force -Path $goCache | Out-Null
  $env:GOCACHE = $goCache

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

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
  $node1Raw = [string](Invoke-LocalnetCliJson -Binary $Binary -Arguments @("address", "convert", $node1)).raw
  $node1SDK = Get-LocalnetKeySDKAddress -NodeHome $node1Home -KeyName "node1"
  $exportUserKey = "export-user"
  $keyOutput = & $Binary keys add $exportUserKey --home $node0Home --keyring-backend test --no-backup 2>&1
  if ($LASTEXITCODE -ne 0) {
    throw "failed to create non-validator export user key"
  }
  $exportUser = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName $exportUserKey
  $exportUserRaw = [string](Invoke-LocalnetCliJson -Binary $Binary -Arguments @("address", "convert", $exportUser)).raw
  $exportUserSDK = Get-LocalnetKeySDKAddress -NodeHome $node0Home -KeyName $exportUserKey

  Send-LocalnetBankTx -Binary $Binary -FromHome $node0Home -FromKey "node0" -ToAddress $node1 -Amount "12345naet" -Fees $Fees -ChainId $ChainId -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Send-LocalnetBankTx -Binary $Binary -FromHome $node0Home -FromKey "node0" -ToAddress $exportUser -Amount "1000naet" -Fees $Fees -ChainId $ChainId -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "bank send flow committed"

  $validator = Get-LocalnetBondedValidator -Binary $Binary -RPCPort $node0Ports.RPC
  Write-Host "direct staking delegation is disabled by policy; export smoke skips obsolete direct delegate tx"
  Write-Host "contract-assets and DEX placeholder tx paths are skipped by export smoke"

  $feesParams = Invoke-QueryCliJson -Arguments @("query", "fees", "params")
  Assert-True (@($feesParams.params.allowed_fee_denoms).Count -eq 1) "fees params must export one allowed fee denom"
  Assert-True (@($feesParams.params.allowed_fee_denoms) -contains "naet") "fees params must include naet"

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  & .\scripts\localnet\export-genesis.ps1 -OutputDir $OutputDir -Binary $Binary -ChainId $ChainId -NodeIndex 0 -ExportDir $ExportDir

  $exportPath = Join-Path $ExportDir "node0-export.json"
  Assert-True (Test-Path -LiteralPath $exportPath) "exported genesis not found at $exportPath"
  $raw = Get-Content -Raw -LiteralPath $exportPath
  Assert-True ($raw -notmatch '(?i)\b(mnemonic|private[_-]?key|priv_validator|node_key|keyring|secret|seed|wallet)\b') "exported genesis contains secret-like material"

  $genesis = $raw | ConvertFrom-Json
  Assert-True ($genesis.chain_id -eq $ChainId) "exported genesis chain-id mismatch"
  Assert-True (@($genesis.app_state.fees.params.allowed_fee_denoms).Count -eq 1) "exported fees allowed denoms count mismatch"
  Assert-True (@($genesis.app_state.fees.params.allowed_fee_denoms) -contains "naet") "exported fees params missing naet"

  Assert-CoinInBalance -Genesis $genesis -Addresses @($node1, $node1Raw, $node1SDK) -Denom "naet" -MinAmount 12345
  Assert-CoinInBalance -Genesis $genesis -Addresses @($exportUser, $exportUserRaw, $exportUserSDK) -Denom "naet" -MinAmount 1000

  Assert-True ($genesis.app_state.staking.params.bond_denom -eq "naet") "exported staking bond denom mismatch"
  $exportUserAddresses = @($exportUser, $exportUserRaw, $exportUserSDK)
  $exportedDelegation = @($genesis.app_state.staking.delegations | Where-Object { $exportUserAddresses -contains $_.delegator_address -and $_.validator_address -eq $validator.operator_address } | Select-Object -First 1)
  Assert-True ($exportedDelegation.Count -eq 0) "exported genesis must not include rejected direct staking delegation"
  Write-Host "exported genesis preserves bank, staking params, and fees state"

  $corruptPath = Join-Path $ExportDir "node0-export-corrupt.json"
  $genesis.app_state.bank.balances[0].coins[0].amount = "not-an-int"
  $genesis | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $corruptPath -Encoding UTF8
  Assert-NativeCommandFails -Arguments @("genesis", "validate-genesis", $corruptPath, "--home", $node0Home) -ExpectedText "amount|invalid|unmarshal|big\.Int|not-an-int"
  Write-Host "corrupted exported genesis is rejected by validate-genesis"

  Write-Host "state export/import smoke completed with export $exportPath"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
