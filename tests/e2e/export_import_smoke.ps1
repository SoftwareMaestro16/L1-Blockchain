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
  [string]$Fees = "1000000naet"
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

  return Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds
}

function Invoke-QueryCliJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--node", $rpcNode, "--output", "json"))
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
    [string]$Address,
    [string]$Denom,
    [int64]$MinAmount
  )

  $entry = @($Genesis.app_state.bank.balances | Where-Object { $_.address -eq $Address } | Select-Object -First 1)
  Assert-True ($entry.Count -eq 1) "missing bank balance entry for $Address"
  $coin = @($entry[0].coins | Where-Object { $_.denom -eq $Denom } | Select-Object -First 1)
  Assert-True ($coin.Count -eq 1) "missing $Denom balance for $Address in exported genesis"
  Assert-True ([int64]$coin[0].amount -ge $MinAmount) "exported $Denom balance for $Address below $MinAmount"
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
  $factoryDenom = "factory/$node0/$FactorySubdenom"

  Send-LocalnetBankTx -Binary $Binary -FromHome $node0Home -FromKey "node0" -ToAddress $node1 -Amount "12345naet" -Fees $Fees -ChainId $ChainId -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "bank send flow committed"

  $validator = Get-LocalnetBondedValidator -Binary $Binary -RPCPort $node0Ports.RPC
  Send-LocalnetDelegateTx -Binary $Binary -FromHome $node0Home -FromKey "node0" -ValidatorAddress $validator.operator_address -Amount "5000000naet" -Fees $Fees -ChainId $ChainId -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $delegation = Get-LocalnetDelegation -Binary $Binary -DelegatorAddress $node0 -ValidatorAddress $validator.operator_address -RPCPort $node0Ports.RPC
  Assert-True ([int64]$delegation.delegation_response.balance.amount -eq 5000000) "staking delegation query did not preserve expected amount"
  Write-Host "staking delegate flow committed"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", $FactorySubdenom) -FromHome $node0Home | Out-Null
  Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom", $node0) -FromHome $node0Home | Out-Null
  $tfQuery = Invoke-QueryCliJson -Arguments @("query", "contract-assets", "denom", $factoryDenom)
  Assert-True ($tfQuery.metadata.admin -eq $node0) "contract-assets admin query mismatch"
  Write-Host "contract-assets create/mint flow committed for $factoryDenom"

  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "10000000naet", "10000000$factoryDenom") -FromHome $node0Home | Out-Null
  $poolQuery = Invoke-QueryCliJson -Arguments @("query", "dex", "pool", "1")
  Assert-True ($poolQuery.pool.lp_denom -eq "lp/1") "DEX pool query did not return lp/1"
  Assert-True ($poolQuery.pool.denom0 -eq $factoryDenom) "DEX pool denom0 mismatch"
  Assert-True ($poolQuery.pool.denom1 -eq "naet") "DEX pool denom1 mismatch"
  Write-Host "DEX create-pool flow committed"

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

  $exportedDenom = @($genesis.app_state.contract-assets.denoms | Where-Object { $_.denom -eq $factoryDenom } | Select-Object -First 1)
  Assert-True ($exportedDenom.Count -eq 1) "exported contract-assets denom missing"
  Assert-True ($exportedDenom[0].admin -eq $node0) "exported contract-assets admin mismatch"

  $exportedPool = @($genesis.app_state.dex.pools | Where-Object { [int64]$_.id -eq 1 } | Select-Object -First 1)
  Assert-True ($exportedPool.Count -eq 1) "exported DEX pool 1 missing"
  Assert-True ($exportedPool[0].denom0 -eq $factoryDenom) "exported DEX denom0 mismatch"
  Assert-True ($exportedPool[0].denom1 -eq "naet") "exported DEX denom1 mismatch"
  Assert-True ([int64]$exportedPool[0].reserve0 -eq 10000000) "exported DEX reserve0 mismatch"
  Assert-True ([int64]$exportedPool[0].reserve1 -eq 10000000) "exported DEX reserve1 mismatch"
  Assert-True ($exportedPool[0].lp_denom -eq "lp/1") "exported DEX lp denom mismatch"

  Assert-CoinInBalance -Genesis $genesis -Address $node0 -Denom $factoryDenom -MinAmount 90000000
  Assert-CoinInBalance -Genesis $genesis -Address $node0 -Denom "lp/1" -MinAmount 10000000
  Assert-CoinInBalance -Genesis $genesis -Address $node1 -Denom "naet" -MinAmount 12345

  Assert-True ($genesis.app_state.staking.params.bond_denom -eq "naet") "exported staking bond denom mismatch"
  $exportedDelegation = @($genesis.app_state.staking.delegations | Where-Object { $_.delegator_address -eq $node0 -and $_.validator_address -eq $validator.operator_address } | Select-Object -First 1)
  Assert-True ($exportedDelegation.Count -eq 1) "exported staking delegation missing"
  Assert-True ($exportedDelegation[0].shares -match '^5000000(\.0+)?$') "exported staking delegation shares mismatch"
  Write-Host "exported genesis preserves bank, staking, fees, contract-assets, and DEX state"

  $corruptPath = Join-Path $ExportDir "node0-export-corrupt.json"
  $genesis.app_state.dex.pools[0].reserve0 = "not-an-int"
  $genesis | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $corruptPath -Encoding UTF8
  Assert-NativeCommandFails -Arguments @("genesis", "validate-genesis", $corruptPath, "--home", $node0Home) -ExpectedText "reserve0|invalid"
  Write-Host "corrupted exported genesis is rejected by validate-genesis"

  Write-Host "state export/import smoke completed with export $exportPath"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
