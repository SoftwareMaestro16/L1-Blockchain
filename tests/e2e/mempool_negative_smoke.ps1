param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 90,
  [int]$BaseP2PPort = 29656,
  [int]$BaseRPCPort = 29657,
  [int]$BaseRESTPort = 1617,
  [int]$BaseGRPCPort = 9390,
  [int]$BasePprofPort = 6360,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [string]$Fees = "1000000naet",
  [string]$WrongFees = "1000testtoken"
)

$ErrorActionPreference = "Stop"

if ($ValidatorCount -lt 2) {
  throw "ValidatorCount must be at least 2 for mempool negative smoke"
}

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet-mempool-negative"
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

function Convert-TextToJson {
  param(
    [string]$Text,
    [string]$Command
  )

  $jsonStart = $Text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $Text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    return $null
  }
  try {
    return $Text.Substring($jsonStart) | ConvertFrom-Json
  } catch {
    return $null
  }
}

function New-SignedTxArgs {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [AllowEmptyString()][string]$TxFees = $Fees,
    [hashtable]$Extra = @{}
  )

  $args = $ActionArgs + @(
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $rpcNode,
    "--output", "json"
  )
  if (-not [string]::IsNullOrWhiteSpace($TxFees)) {
    $args += @("--fees", $TxFees)
  }
  foreach ($key in $Extra.Keys) {
    $args += @($key, [string]$Extra[$key])
  }
  return $args
}

function Send-SignedTx {
  param(
    [string[]]$ActionArgs,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [AllowEmptyString()][string]$TxFees = $Fees
  )

  return Send-LocalnetTx `
    -Binary $Binary `
    -Arguments (New-SignedTxArgs -ActionArgs $ActionArgs -FromHome $FromHome -FromKey $FromKey -TxFees $TxFees) `
    -RPCPort $node0Ports.RPC `
    -TimeoutSeconds $TimeoutSeconds
}

function Invoke-QueryCliJson {
  param([string[]]$Arguments)

  return Invoke-LocalnetCliJson -Binary $Binary -Arguments ($Arguments + @("--node", $rpcNode, "--output", "json"))
}

function Get-BalanceAmount {
  param(
    [string]$Address,
    [string]$Denom
  )

  $balance = Get-LocalnetBankBalance -Binary $Binary -Address $Address -Denom $Denom -RPCPort $node0Ports.RPC
  if (-not $balance.amount) {
    return [int64]0
  }
  return [int64]$balance.amount
}

function Get-SupplyAmount {
  param([string]$Denom)

  $supply = Get-LocalnetBankSupplyOf -Binary $Binary -Denom $Denom -RPCPort $node0Ports.RPC
  if (-not $supply.amount) {
    return [int64]0
  }
  return [int64]$supply.amount
}

function Get-Contract assetsDenomCount {
  $res = Invoke-QueryCliJson -Arguments @("query", "contract-assets", "denoms", "--limit", "50")
  return @($res.denoms).Count
}

function Get-DexPoolCount {
  $res = Invoke-QueryCliJson -Arguments @("query", "dex", "pools", "--limit", "50")
  return @($res.pools).Count
}

function Invoke-NegativeTx {
  param(
    [string]$Name,
    [string[]]$Arguments,
    [string]$ExpectedText = "",
    [string[]]$AllowedPhases = @("CheckTx", "DeliverTx", "CLI")
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }

  $text = $output -join "`n"
  $json = Convert-TextToJson -Text $text -Command "$Binary $($Arguments -join ' ')"
  if ($null -eq $json) {
    if ($exitCode -eq 0) {
      throw "$Name returned no JSON but exit code was zero"
    }
    if ($ExpectedText -and ($text -notmatch $ExpectedText)) {
      throw "$Name CLI failure did not match '$ExpectedText': $text"
    }
    Assert-True (@($AllowedPhases) -contains "CLI") "$Name failed in CLI, expected one of $($AllowedPhases -join ',')"
    Write-Host "$Name rejected in CLI validation"
    return [pscustomobject]@{ Name = $Name; Phase = "CLI"; Code = $exitCode; Log = $text; TxHash = "" }
  }

  $code = Get-LocalnetTxCode -Tx $json
  if ($code -ne 0) {
    $log = Get-LocalnetTxLog -Tx $json
    if ([string]::IsNullOrWhiteSpace($log)) {
      $log = $text
    }
    if ($ExpectedText -and ($log -notmatch $ExpectedText)) {
      throw "$Name CheckTx log did not match '$ExpectedText': $log"
    }
    Assert-True (@($AllowedPhases) -contains "CheckTx") "$Name failed in CheckTx, expected one of $($AllowedPhases -join ',')"
    Write-Host "$Name rejected in CheckTx: $log"
    return [pscustomobject]@{ Name = $Name; Phase = "CheckTx"; Code = $code; Log = $log; TxHash = (Get-LocalnetTxHash -Tx $json) }
  }

  $txHash = Get-LocalnetTxHash -Tx $json
  if (-not $txHash) {
    throw "$Name returned successful CheckTx response without txhash"
  }
  $tx = Wait-LocalnetTx -Binary $Binary -TxHash $txHash -RPCPort $node0Ports.RPC -TimeoutSeconds $TimeoutSeconds -ExpectFailure
  $log = Get-LocalnetTxLog -Tx $tx
  if ($ExpectedText -and ($log -notmatch $ExpectedText)) {
    throw "$Name DeliverTx log did not match '$ExpectedText': $log"
  }
  Assert-True (@($AllowedPhases) -contains "DeliverTx") "$Name failed in DeliverTx, expected one of $($AllowedPhases -join ',')"
  Write-Host "$Name rejected in DeliverTx: $log"
  return [pscustomobject]@{ Name = $Name; Phase = "DeliverTx"; Code = (Get-LocalnetTxCode -Tx $tx); Log = $log; TxHash = $txHash }
}

function New-SignedReplayTx {
  param(
    [string[]]$GenerateArguments,
    [string]$FromKey,
    [string]$FromHome,
    [string]$WorkDir
  )

  New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
  $unsignedPath = Join-Path $WorkDir "negative-replay-unsigned.json"
  $signedPath = Join-Path $WorkDir "negative-replay-signed.json"
  $unsigned = Invoke-LocalnetCliJson -Binary $Binary -Arguments ($GenerateArguments + @("--generate-only", "--output", "json"))
  [System.IO.File]::WriteAllText($unsignedPath, ($unsigned | ConvertTo-Json -Depth 100), (New-Object System.Text.UTF8Encoding $false))

  $signArgs = @(
    "tx", "sign", $unsignedPath,
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--node", $rpcNode,
    "--output", "json",
    "--output-document", $signedPath
  )
  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $signOutput = & $Binary @signArgs 2>&1
    $signExitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  if ($signExitCode -ne 0) {
    throw "tx sign failed: $($signOutput -join "`n")"
  }
  Assert-True (Test-Path -LiteralPath $signedPath) "signed replay tx not created"
  return $signedPath
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

  $node0Home = Join-Path $OutputDir "node0\aetrad"
  $node1Home = Join-Path $OutputDir "node1\aetrad"
  $node0 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node0Home -KeyName "node0"
  $node1 = Get-LocalnetKeyAddress -Binary $Binary -NodeHome $node1Home -KeyName "node1"
  $factoryDenom = "factory/$node0/negasset"

  Send-SignedTx -ActionArgs @("tx", "contract-assets", "create-denom", "negasset") -FromHome $node0Home | Out-Null
  Send-SignedTx -ActionArgs @("tx", "contract-assets", "mint", "100000000$factoryDenom", $node0) -FromHome $node0Home | Out-Null
  Send-SignedTx -ActionArgs @("tx", "dex", "create-pool", "10000000naet", "10000000$factoryDenom") -FromHome $node0Home | Out-Null
  Write-Host "baseline contract-assets denom and DEX pool created"

  $node1Before = Get-BalanceAmount -Address $node1 -Denom "naet"
  $wrongFee = Invoke-NegativeTx `
    -Name "wrong fee bank send" `
    -Arguments (New-SignedTxArgs -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet") -FromHome $node0Home -TxFees $WrongFees) `
    -ExpectedText "fee denom testtoken not accepted; use naet" `
    -AllowedPhases @("CheckTx")
  Assert-True ($wrongFee.Phase -eq "CheckTx") "wrong fee must be rejected in CheckTx"
  Assert-True ((Get-BalanceAmount -Address $node1 -Denom "naet") -eq $node1Before) "wrong fee bank send changed receiver balance"

  $node0Before = Get-BalanceAmount -Address $node0 -Denom "naet"
  $insufficient = Invoke-NegativeTx `
    -Name "insufficient funds bank send" `
    -Arguments (New-SignedTxArgs -ActionArgs @("tx", "bank", "send", "node1", $node0, "999999999999999999999naet") -FromHome $node1Home -FromKey "node1") `
    -ExpectedText "insufficient|spendable|funds" `
    -AllowedPhases @("DeliverTx")
  Assert-True ($insufficient.Phase -eq "DeliverTx") "insufficient bank send should reach DeliverTx after fee ante"
  Assert-True ((Get-BalanceAmount -Address $node0 -Denom "naet") -eq $node0Before) "insufficient bank send changed receiver balance"

  $node1BeforeStaleSequence = Get-BalanceAmount -Address $node1 -Denom "naet"
  $signedStaleSequence = New-SignedReplayTx `
    -GenerateArguments (New-SignedTxArgs -ActionArgs @("tx", "bank", "send", "node0", $node1, "2naet") -FromHome $node0Home) `
    -FromKey "node0" `
    -FromHome $node0Home `
    -WorkDir (Join-Path $OutputDir "negative-replay")
  Send-SignedTx -ActionArgs @("tx", "bank", "send", "node0", $node1, "1naet") -FromHome $node0Home | Out-Null
  $node1AfterSequenceAdvance = Get-BalanceAmount -Address $node1 -Denom "naet"
  Assert-True ($node1AfterSequenceAdvance -eq ($node1BeforeStaleSequence + 1)) "sequence-advancing tx did not send 1naet"
  $replay = Invoke-NegativeTx `
    -Name "invalid sequence replay" `
    -Arguments @("tx", "broadcast", $signedStaleSequence, "--node", $rpcNode, "--broadcast-mode", "sync", "--output", "json") `
    -ExpectedText "sequence|account sequence|signature verification failed" `
    -AllowedPhases @("CheckTx")
  Assert-True ($replay.Phase -eq "CheckTx") "signed replay must be rejected in CheckTx"
  Assert-True ((Get-BalanceAmount -Address $node1 -Denom "naet") -eq $node1AfterSequenceAdvance) "stale sequence tx changed receiver balance"

  $factorySupplyBefore = Get-SupplyAmount -Denom $factoryDenom
  $node1FactoryBefore = Get-BalanceAmount -Address $node1 -Denom $factoryDenom
  Invoke-NegativeTx `
    -Name "unauthorized contract-assets mint" `
    -Arguments (New-SignedTxArgs -ActionArgs @("tx", "contract-assets", "mint", "1$factoryDenom", $node1) -FromHome $node1Home -FromKey "node1") `
    -ExpectedText "only denom admin can mint" `
    -AllowedPhases @("DeliverTx") | Out-Null
  Assert-True ((Get-SupplyAmount -Denom $factoryDenom) -eq $factorySupplyBefore) "unauthorized mint changed supply"
  Assert-True ((Get-BalanceAmount -Address $node1 -Denom $factoryDenom) -eq $node1FactoryBefore) "unauthorized mint changed node1 factory balance"

  $denomCountBefore = Get-Contract assetsDenomCount
  Invoke-NegativeTx `
    -Name "malformed contract-assets subdenom" `
    -Arguments (New-SignedTxArgs -ActionArgs @("tx", "contract-assets", "create-denom", "!") -FromHome $node0Home) `
    -ExpectedText "subdenom|invalid|3-64" `
    -AllowedPhases @("CLI", "DeliverTx") | Out-Null
  Assert-True ((Get-Contract assetsDenomCount) -eq $denomCountBefore) "malformed subdenom changed contract-assets denom count"

  $poolCountBefore = Get-DexPoolCount
  $poolBefore = (Invoke-QueryCliJson -Arguments @("query", "dex", "pool", "1")).pool | ConvertTo-Json -Depth 20 -Compress
  Invoke-NegativeTx `
    -Name "duplicate DEX pool" `
    -Arguments (New-SignedTxArgs -ActionArgs @("tx", "dex", "create-pool", "1naet", "1$factoryDenom") -FromHome $node0Home) `
    -ExpectedText "pool already exists" `
    -AllowedPhases @("DeliverTx") | Out-Null
  Assert-True ((Get-DexPoolCount) -eq $poolCountBefore) "duplicate pool changed pool count"
  $poolAfter = (Invoke-QueryCliJson -Arguments @("query", "dex", "pool", "1")).pool | ConvertTo-Json -Depth 20 -Compress
  Assert-True ($poolAfter -eq $poolBefore) "duplicate pool changed existing pool state"

  Invoke-NegativeTx `
    -Name "malformed DEX denom" `
    -Arguments (New-SignedTxArgs -ActionArgs @("tx", "dex", "create-pool", "1naet", "1!") -FromHome $node0Home) `
    -ExpectedText "invalid coin|invalid denom|invalid" `
    -AllowedPhases @("CLI", "DeliverTx") | Out-Null
  Assert-True ((Get-DexPoolCount) -eq $poolCountBefore) "malformed DEX denom changed pool count"

  Write-Host "mempool negative smoke completed: wrong fee=$($wrongFee.Phase), insufficient=$($insufficient.Phase), replay=$($replay.Phase)"
} finally {
  & .\scripts\localnet\stop.ps1 -OutputDir $OutputDir
  Pop-Location
}
