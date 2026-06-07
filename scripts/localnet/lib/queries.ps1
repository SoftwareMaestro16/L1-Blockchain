function Get-LocalnetStakingParams {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "staking", "params", "--node", $node, "--output", "json")
  return $result.params
}

function Get-LocalnetStakingValidators {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "staking", "validators", "--node", $node, "--output", "json")
  return @($result.validators)
}

function Get-LocalnetBondedValidator {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  foreach ($validator in @(Get-LocalnetStakingValidators -Binary $Binary -RPCPort $RPCPort)) {
    $status = [string]$validator.status
    if ($status -eq "BOND_STATUS_BONDED" -or $status -eq "3") {
      return $validator
    }
  }
  throw "No bonded staking validator found on RPC $RPCPort"
}

function Get-LocalnetSlashingParams {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "slashing", "params", "--node", $node, "--output", "json")
  return $result.params
}

function Get-LocalnetSigningInfos {
  param(
    [string]$Binary,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "slashing", "signing-infos", "--node", $node, "--output", "json")
  if ($result.info) { return @($result.info) }
  if ($result.signing_infos) { return @($result.signing_infos) }
  return @()
}

function Get-LocalnetBankMetadata {
  param(
    [string]$Binary,
    [string]$Denom = "naet",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "denom-metadata", $Denom, "--node", $node, "--output", "json")
  return $result.metadata
}

function Get-LocalnetBankSupplyOf {
  param(
    [string]$Binary,
    [string]$Denom = "naet",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "total-supply-of", $Denom, "--node", $node, "--output", "json")
  if ($result.amount) { return $result.amount }
  return $result
}

function Get-LocalnetBankBalance {
  param(
    [string]$Binary,
    [string]$Address,
    [string]$Denom = "naet",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "balance", $Address, $Denom, "--node", $node, "--output", "json")
  if ($result.balance) { return $result.balance }
  return $result
}

function Get-LocalnetKeyAddress {
  param(
    [string]$Binary,
    [string]$NodeHome,
    [string]$KeyName
  )

  $output = & $Binary keys show $KeyName -a --home $NodeHome --keyring-backend test 2>&1
  if ($LASTEXITCODE -ne 0) {
    throw "failed to read key $KeyName from $NodeHome`n$($output -join "`n")"
  }
  return (($output | Select-Object -Last 1).ToString().Trim())
}

function Send-LocalnetDelegateTx {
  param(
    [string]$Binary,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [string]$ValidatorAddress,
    [string]$Amount = "5000000naet",
    [string]$Fees = "1000naet",
    [string]$ChainId = "aetra-local-1",
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $tx = Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "tx", "staking", "delegate", $ValidatorAddress, $Amount,
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $node,
    "--output", "json"
  )

  $txHash = $tx.txhash
  if (-not $txHash -and $tx.tx_response) {
    $txHash = $tx.tx_response.txhash
  }
  if (-not $txHash) {
    throw "staking delegate did not return txhash"
  }

  return Wait-LocalnetTx -Binary $Binary -TxHash $txHash -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds
}

function Get-LocalnetDelegation {
  param(
    [string]$Binary,
    [string]$DelegatorAddress,
    [string]$ValidatorAddress,
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  return Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "query", "staking", "delegation", $DelegatorAddress, $ValidatorAddress,
    "--node", $node,
    "--output", "json"
  )
}

function Send-LocalnetBankTx {
  param(
    [string]$Binary,
    [string]$FromHome,
    [string]$FromKey = "node0",
    [string]$ToAddress,
    [string]$Amount = "1000naet",
    [string]$Fees = "1000naet",
    [string]$ChainId = "aetra-local-1",
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $tx = Invoke-LocalnetCliJson -Binary $Binary -Arguments @(
    "tx", "bank", "send", $FromKey, $ToAddress, $Amount,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--fees", $Fees,
    "--yes",
    "--broadcast-mode", "sync",
    "--node", $node,
    "--output", "json"
  )

  $txHash = $tx.txhash
  if (-not $txHash -and $tx.tx_response) {
    $txHash = $tx.tx_response.txhash
  }
  if (-not $txHash) {
    throw "bank send did not return txhash"
  }

  return Wait-LocalnetTx -Binary $Binary -TxHash $txHash -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds
}
