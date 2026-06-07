param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetra-local-1",
  [int]$RPCPort = 26657,
  [string]$FromHome = "",
  [string]$FromKey = "node0",
  [string[]]$Recipients = @(),
  [string[]]$Transfers = @(),
  [string]$Amount = "1000000naet",
  [string]$Fees = "1000000naet",
  [int]$TimeoutSeconds = 60,
  [switch]$Json
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

function Assert-LocalFundingTarget {
  param(
    [string]$ChainId,
    [int]$RPCPort,
    [int]$TimeoutSeconds
  )

  if ($ChainId -notmatch '(^|-)local(-|$)') {
    throw "refusing local funding for non-local chain-id: $ChainId"
  }
  if ($RPCPort -le 0 -or $RPCPort -gt 65535) {
    throw "invalid local RPC port: $RPCPort"
  }

  $status = Wait-LocalnetRpc -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds
  $network = [string]$status.result.node_info.network
  if ($network -ne $ChainId) {
    throw "RPC network $network does not match requested local chain-id $ChainId"
  }
  return $network
}

function Convert-TransferSpec {
  param([string]$Spec)

  if ($Spec -notmatch '^([^=]+)=(.+)$') {
    throw "transfer must use address=amount format: $Spec"
  }
  return @{
    Recipient = $Matches[1].Trim()
    Amount    = $Matches[2].Trim()
  }
}

function Get-CoinParts {
  param([string]$Coin)

  if ($Coin -notmatch '^([1-9][0-9]*)([A-Za-z][A-Za-z0-9/:._-]*)$') {
    throw "amount must be a positive SDK coin such as 1000000naet: $Coin"
  }
  return @{
    Amount = [decimal]$Matches[1]
    Denom  = $Matches[2]
  }
}

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$FromHome = Resolve-LocalnetPath -Path $FromHome -DefaultRelativePath "node0\aetrad"
if ([string]::IsNullOrWhiteSpace($FromHome) -or -not (Test-Path -LiteralPath $FromHome)) {
  $FromHome = Join-Path $OutputDir "node0\aetrad"
}

Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "local funding output directory"
Assert-LocalnetWorkspacePath -Path $FromHome -Purpose "local funding key home"
if (-not (Test-Path -LiteralPath $Binary)) {
  throw "binary not found: $Binary"
}
if (-not (Test-Path -LiteralPath $FromHome)) {
  throw "funder home not found: $FromHome"
}

$network = Assert-LocalFundingTarget -ChainId $ChainId -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds

$transferList = New-Object System.Collections.Generic.List[hashtable]
foreach ($recipient in $Recipients) {
  if (-not [string]::IsNullOrWhiteSpace($recipient)) {
    $transferList.Add(@{ Recipient = $recipient.Trim(); Amount = $Amount })
  }
}
foreach ($spec in $Transfers) {
  if (-not [string]::IsNullOrWhiteSpace($spec)) {
    $transferList.Add((Convert-TransferSpec -Spec $spec))
  }
}
if ($transferList.Count -eq 0) {
  throw "provide at least one -Recipients address or -Transfers address=amount entry"
}

$results = @()
foreach ($transfer in $transferList) {
  $recipient = [string]$transfer.Recipient
  $coin = Get-CoinParts -Coin ([string]$transfer.Amount)
  $before = Get-LocalnetBankBalance -Binary $Binary -Address $recipient -Denom $coin.Denom -RPCPort $RPCPort

  $tx = Send-LocalnetBankTx `
    -Binary $Binary `
    -FromHome $FromHome `
    -FromKey $FromKey `
    -ToAddress $recipient `
    -Amount ([string]$transfer.Amount) `
    -Fees $Fees `
    -ChainId $ChainId `
    -RPCPort $RPCPort `
    -TimeoutSeconds $TimeoutSeconds

  $after = Get-LocalnetBankBalance -Binary $Binary -Address $recipient -Denom $coin.Denom -RPCPort $RPCPort
  $beforeAmount = [decimal]$before.amount
  $afterAmount = [decimal]$after.amount
  $expected = $beforeAmount + $coin.Amount
  if ($afterAmount -ne $expected) {
    throw "funding balance mismatch for $recipient $($coin.Denom): before=$beforeAmount after=$afterAmount expected=$expected"
  }

  $results += [pscustomobject]@{
    chain_id = $network
    rpc_port = $RPCPort
    from_key = $FromKey
    recipient = $recipient
    amount = [string]$transfer.Amount
    txhash = (Get-LocalnetTxHash -Tx $tx)
    balance_before = $beforeAmount.ToString()
    balance_after = $afterAmount.ToString()
  }
}

if ($Json) {
  $results | ConvertTo-Json -Depth 20
} else {
  foreach ($result in $results) {
    Write-Host "funded $($result.recipient) with $($result.amount) on $($result.chain_id); balance $($result.balance_before)->$($result.balance_after)"
  }
}
