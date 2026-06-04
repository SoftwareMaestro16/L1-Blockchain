param(
  [ValidateSet("Smoke", "Full")]
  [string]$Profile = "Smoke",
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "orbitalis-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 4,
  [int]$TimeoutSeconds = 120,
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
  [string]$Node = "",
  [string]$Fees = "1000000norb",
  [string]$WrongFees = "1000testtoken",
  [string]$FactorySubdenom = "acceptgold",
  [string]$DelegationAmount = "5000000norb",
  [switch]$SkipBuild,
  [switch]$KeepLogsOnFailure
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for prototype acceptance"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")
. (Join-Path $RepoRoot "tests\e2e\prototype_acceptance_helpers.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\orbitalisd.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "acceptance localnet output directory"
if (-not $SkipBuild) {
  Assert-LocalnetWorkspacePath -Path (Split-Path $Binary) -Purpose "acceptance binary output directory"
}

$node0Ports = Get-LocalnetPortProfile `
  -Index 0 `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BasePprofPort $BasePprofPort `
  -PortStride $PortStride
$rpcNode = if ([string]::IsNullOrWhiteSpace($Node)) { "tcp://127.0.0.1:$($node0Ports.RPC)" } else { $Node }

$ctx = [pscustomobject]@{
  RepoRoot       = $RepoRoot
  OutputDir      = $OutputDir
  Binary         = $Binary
  ChainId        = $ChainId
  ValidatorCount = $ValidatorCount
  MinHeight      = $MinHeight
  TimeoutSeconds = $TimeoutSeconds
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
  Node0Ports     = $node0Ports
  RpcNode        = $rpcNode
  GrpcAddr       = "127.0.0.1:$($node0Ports.GRPC)"
  RestBase       = "http://127.0.0.1:$($node0Ports.REST)"
  Fees           = $Fees
}

$failure = $null

Push-Location $RepoRoot
try {
  Write-AcceptanceStep "profile=$Profile validators=$ValidatorCount chain-id=$ChainId"

  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir

  if (-not $SkipBuild) {
    Write-AcceptanceStep "build orbitalisd"
    Invoke-AcceptanceBuild -Context $ctx
  } elseif (!(Test-Path -LiteralPath $Binary)) {
    throw "Binary not found at $Binary and -SkipBuild was specified"
  }

  Write-AcceptanceStep "reset/init localnet"
  Invoke-AcceptanceLocalnetScript -Context $ctx -ScriptName "init.ps1" -Extra @{ SkipBuild = $true }
  Invoke-AcceptanceLocalnetScript -Context $ctx -ScriptName "validate-genesis.ps1"

  Write-AcceptanceStep "start validators"
  Invoke-AcceptanceLocalnetScript -Context $ctx -ScriptName "start.ps1" -Extra @{ NoInit = $true }
  Wait-LocalnetRpc -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  $height = Wait-LocalnetHeight -TargetHeight $MinHeight -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Wait-LocalnetValidators -ExpectedCount $ValidatorCount -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  if ($ValidatorCount -gt 1) {
    Wait-LocalnetPeers -ExpectedMinPeers 1 -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds | Out-Null
  }
  & .\scripts\localnet\health.ps1 -OutputDir $OutputDir -ValidatorCount $ValidatorCount -TimeoutSeconds $TimeoutSeconds | Out-Null
  Write-Host "localnet healthy at height $height"

  $node0Home = Join-Path $OutputDir "node0\orbitalisd"
  $node1Home = Join-Path $OutputDir "node1\orbitalisd"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"

  Write-AcceptanceStep "query base state"
  $status = Invoke-LocalnetRpc -RPCPort $node0Ports.RPC -Path "/status"
  if ($status.result.node_info.network -ne $ChainId) {
    throw "RPC status network mismatch: $($status.result.node_info.network)"
  }
  $latestBlock = Invoke-AcceptanceQueryCliJson -Context $ctx -Arguments @("query", "block")
  if (-not ($latestBlock.header.height -or $latestBlock.block.header.height)) {
    throw "query block did not return a block height"
  }
  $metadata = Get-LocalnetBankMetadata -Binary $Binary -Denom "norb" -RPCPort $node0Ports.RPC
  Assert-AcceptanceNativeMetadata -Metadata $metadata
  $feesParams = Invoke-AcceptanceQueryGrpcJson -Context $ctx -Arguments @("query", "fees", "params")
  Assert-AcceptanceFeesParams -Params $feesParams.params
  if ($EnableAPI) {
    $restNode = Invoke-AcceptanceRestJson -Context $ctx -Path "/cosmos/base/tendermint/v1beta1/node_info"
    if ($restNode.default_node_info.network -ne $ChainId) {
      throw "REST node_info network mismatch: $($restNode.default_node_info.network)"
    }
  }
  Write-Host "base CLI/gRPC/REST queries passed"

  Write-AcceptanceStep "bank send"
  $node1Before = Get-AcceptanceBalanceAmount -Context $ctx -Address $node1 -Denom "norb"
  Send-AcceptanceTx -Context $ctx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1000norb") -FromHome $node0Home | Out-Null
  $node1After = Get-AcceptanceBalanceAmount -Context $ctx -Address $node1 -Denom "norb"
  if ($node1After -ne ($node1Before + 1000)) {
    throw "bank send did not increase node1 balance by 1000norb: before=$node1Before after=$node1After"
  }
  Write-Host "bank send updated node1 balance to $($node1After)norb"

  Write-AcceptanceStep "fees policy"
  Send-AcceptanceTx `
    -Context $ctx `
    -ActionArgs @("tx", "bank", "send", "node0", $node1, "1norb") `
    -FromHome $node0Home `
    -Fees $WrongFees `
    -ExpectFailure `
    -ExpectedLog "fee denom testtoken not accepted; use norb" | Out-Null
  Write-Host "wrong fee denom rejected"

  Write-AcceptanceStep "tokenfactory create/mint/query"
  Send-AcceptanceTx -Context $ctx -ActionArgs @("tx", "tokenfactory", "create-denom", $FactorySubdenom) -FromHome $node0Home | Out-Null
  $factoryDenom = "factory/$node0/$FactorySubdenom"
  $tfMeta = Invoke-AcceptanceQueryGrpcJson -Context $ctx -Arguments @("query", "tokenfactory", "denom", $factoryDenom)
  if ($tfMeta.metadata.admin -ne $node0) {
    throw "tokenfactory admin mismatch"
  }
  Send-AcceptanceTx -Context $ctx -ActionArgs @("tx", "tokenfactory", "mint", "100000000$factoryDenom", $node0) -FromHome $node0Home | Out-Null
  $factoryBalance = Get-AcceptanceBalanceAmount -Context $ctx -Address $node0 -Denom $factoryDenom
  if ($factoryBalance -lt 100000000) {
    throw "factory balance after mint too low: $factoryBalance"
  }
  if ($EnableAPI) {
    $tfRest = Invoke-AcceptanceRestJson -Context $ctx -Path "/l1/tokenfactory/v1/denom/$factoryDenom"
    if ($tfRest.metadata.admin -ne $node0) {
      throw "REST tokenfactory admin mismatch"
    }
  }
  Write-Host "factory denom $factoryDenom minted to node0"

  Write-AcceptanceStep "DEX create pool/swap/query"
  Send-AcceptanceTx -Context $ctx -ActionArgs @("tx", "dex", "create-pool", "10000000norb", "10000000$factoryDenom") -FromHome $node0Home | Out-Null
  $pool = Invoke-AcceptanceQueryGrpcJson -Context $ctx -Arguments @("query", "dex", "pool", "1")
  if ($pool.pool.lp_denom -ne "lp/1") {
    throw "DEX pool 1 returned unexpected lp denom $($pool.pool.lp_denom)"
  }
  if ($Profile -eq "Full") {
    Send-AcceptanceTx `
      -Context $ctx `
      -ActionArgs @("tx", "dex", "swap-exact-in", "1", "100000norb", $factoryDenom, "1000000") `
      -FromHome $node0Home `
      -ExpectFailure `
      -ExpectedLog "amount out below minimum" | Out-Null
    Write-Host "DEX slippage guard rejected excessive min_amount_out"
  }
  $factoryBeforeSwap = Get-AcceptanceBalanceAmount -Context $ctx -Address $node0 -Denom $factoryDenom
  Send-AcceptanceTx -Context $ctx -ActionArgs @("tx", "dex", "swap-exact-in", "1", "100000norb", $factoryDenom, "1") -FromHome $node0Home | Out-Null
  $factoryAfterSwap = Get-AcceptanceBalanceAmount -Context $ctx -Address $node0 -Denom $factoryDenom
  if ($factoryAfterSwap -le $factoryBeforeSwap) {
    throw "factory balance did not increase after DEX swap: before=$factoryBeforeSwap after=$factoryAfterSwap"
  }
  if ($EnableAPI) {
    $poolRest = Invoke-AcceptanceRestJson -Context $ctx -Path "/l1/dex/v1/pools/1"
    if ($poolRest.pool.lp_denom -ne "lp/1") {
      throw "REST DEX pool query returned unexpected lp denom"
    }
  }
  Write-Host "DEX swap increased factory balance from $factoryBeforeSwap to $factoryAfterSwap"

  Write-AcceptanceStep "PoS delegation/slashing queries"
  $stakingParams = Get-LocalnetStakingParams -Binary $Binary -RPCPort $node0Ports.RPC
  if ($stakingParams.bond_denom -ne "norb") {
    throw "staking bond denom must be norb, got $($stakingParams.bond_denom)"
  }
  $validators = @(Get-LocalnetStakingValidators -Binary $Binary -RPCPort $node0Ports.RPC)
  if ($validators.Count -ne $ValidatorCount) {
    throw "staking validators query returned $($validators.Count), expected $ValidatorCount"
  }
  foreach ($validator in $validators) {
    Assert-AcceptanceBondedValidator -Validator $validator
  }
  $selectedValidator = @($validators | Sort-Object -Property operator_address | Select-Object -First 1)[0]
  $slashingParams = Get-LocalnetSlashingParams -Binary $Binary -RPCPort $node0Ports.RPC
  if ([int64]$slashingParams.signed_blocks_window -le 0) {
    throw "slashing signed_blocks_window must be positive"
  }
  $beforePower = Get-LocalnetTotalVotingPower -RPCPort $node0Ports.RPC
  Send-LocalnetDelegateTx `
    -Binary $Binary `
    -FromHome $node0Home `
    -FromKey "node0" `
    -ValidatorAddress $selectedValidator.operator_address `
    -Amount $DelegationAmount `
    -Fees $Fees `
    -ChainId $ChainId `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds | Out-Null
  $delegation = Get-LocalnetDelegation -Binary $Binary -DelegatorAddress $node0 -ValidatorAddress $selectedValidator.operator_address -RPCPort $node0Ports.RPC
  $delegationBalance = if ($delegation.delegation_response.balance) { $delegation.delegation_response.balance } else { $delegation.balance }
  if ($delegationBalance.denom -ne "norb" -or [int64]$delegationBalance.amount -lt 5000000) {
    throw "delegation query returned unexpected balance $($delegationBalance.amount)$($delegationBalance.denom)"
  }
  $afterPower = Wait-LocalnetTotalVotingPowerGreater -PreviousPower $beforePower -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
  Write-Host "delegation increased voting power from $beforePower to $afterPower"

  if ($Profile -eq "Full") {
    Write-AcceptanceStep "full profile restart/health"
    $heightBeforeRestart = Get-LocalnetHeight -RPCPort $node0Ports.RPC
    & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
    Invoke-AcceptanceLocalnetScript -Context $ctx -ScriptName "start.ps1" -Extra @{ NoInit = $true }
    $restartHeight = Wait-LocalnetHeight -TargetHeight ($heightBeforeRestart + 1) -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds
    & .\scripts\localnet\health.ps1 -OutputDir $OutputDir -ValidatorCount $ValidatorCount -TimeoutSeconds $TimeoutSeconds | Out-Null
    Write-Host "restart preserved chain progress: $heightBeforeRestart->$restartHeight"
  }

  $height = Get-LocalnetHeight -RPCPort $node0Ports.RPC
  Write-Host "prototype acceptance $Profile passed at height $height"
} catch {
  $failure = $_
  Write-Host "prototype acceptance failed: $($failure.Exception.Message)"
  Invoke-AcceptanceDiagnostics -Context $ctx -Reason $Profile
  if (-not $KeepLogsOnFailure) {
    Write-Host "diagnostic bundle collected; pass -KeepLogsOnFailure to preserve localnet output after failure"
  }
  throw
} finally {
  try {
    & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
    if ($failure -and -not $KeepLogsOnFailure) {
      & .\scripts\localnet\reset.ps1 -OutputDir $OutputDir
    }
  } finally {
    Pop-Location
  }
}
