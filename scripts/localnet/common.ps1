$ErrorActionPreference = "Stop"

function Get-LocalnetRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function Resolve-LocalnetPath {
  param(
    [string]$Path,
    [string]$DefaultRelativePath
  )

  $repoRoot = Get-LocalnetRepoRoot
  if ([string]::IsNullOrWhiteSpace($Path)) {
    $Path = Join-Path $repoRoot $DefaultRelativePath
  } elseif (-not [System.IO.Path]::IsPathRooted($Path)) {
    $Path = Join-Path $repoRoot $Path
  }

  return [System.IO.Path]::GetFullPath($Path)
}

function Assert-LocalnetWorkspacePath {
  param(
    [string]$Path,
    [string]$Purpose = "localnet path"
  )

  $repoRoot = (Get-LocalnetRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar

  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Remove-LocalnetDirectory {
  param([string]$OutputDir)

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  Assert-LocalnetWorkspacePath -Path $resolved -Purpose "delete target"
  if (Test-Path -LiteralPath $resolved) {
    Remove-Item -LiteralPath $resolved -Recurse -Force
  }
}

function Get-LocalnetPortProfile {
  param(
    [int]$Index,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseRESTPort = 1317,
    [int]$BaseGRPCPort = 9090,
    [int]$BasePprofPort = 6060,
    [int]$PortStride = 100
  )

  return @{
    OldP2P  = 16656 + $Index
    OldRPC  = 26657 + $Index
    OldAPI  = 1317 + $Index
    OldGRPC = 9090 + $Index
    P2P     = $BaseP2PPort + ($PortStride * $Index)
    RPC     = $BaseRPCPort + ($PortStride * $Index)
    REST    = $BaseRESTPort + $Index
    GRPC    = $BaseGRPCPort + $Index
    Pprof   = $BasePprofPort + $Index
  }
}

function Get-LocalnetNodes {
  param(
    [string]$OutputDir,
    [string]$NodeDirPrefix = "node",
    [string]$NodeDaemonHome = "orbitalisd"
  )

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  if (!(Test-Path -LiteralPath $resolved)) {
    return @()
  }

  return @(Get-ChildItem -LiteralPath $resolved -Directory -Filter "$NodeDirPrefix*" |
    Where-Object { Test-Path (Join-Path $_.FullName "$NodeDaemonHome\config\genesis.json") } |
    Sort-Object {
      if ($_.Name -match "^$([regex]::Escape($NodeDirPrefix))(\d+)$") {
        [int]$Matches[1]
      } else {
        [int]::MaxValue
      }
    })
}

function Set-TomlSectionValue {
  param(
    [string]$Content,
    [string]$Section,
    [string]$Key,
    [string]$Value
  )

  $sectionEsc = [regex]::Escape($Section)
  $keyEsc = [regex]::Escape($Key)
  $pattern = "(?ms)(\[$sectionEsc\].*?^$keyEsc\s*=\s*)([^\r\n]+)"
  return [regex]::Replace($Content, $pattern, { param($match) $match.Groups[1].Value + $Value }, 1)
}

function Set-LocalnetGeneratedPorts {
  param(
    [string]$OutputDir,
    [int]$ValidatorCount,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseRESTPort = 1317,
    [int]$BaseGRPCPort = 9090,
    [int]$BasePprofPort = 6060,
    [int]$PortStride = 100,
    [bool]$EnableAPI = $true,
    [bool]$EnableGRPC = $true,
    [bool]$EnableRPC = $true,
    [string]$MinimumGasPrices = "0norb",
    [string]$LogLevel = "info"
  )

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  for ($i = 0; $i -lt $ValidatorCount; $i++) {
    $nodeHome = Join-Path $resolved "node$i\orbitalisd"
    $configToml = Join-Path $nodeHome "config\config.toml"
    $appToml = Join-Path $nodeHome "config\app.toml"
    $p = Get-LocalnetPortProfile -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride

    $config = Get-Content -Raw -LiteralPath $configToml
    for ($peer = 0; $peer -lt $ValidatorCount; $peer++) {
      $peerPorts = Get-LocalnetPortProfile -Index $peer -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
      $config = $config -replace ":$($peerPorts.OldP2P)", ":$($peerPorts.P2P)"
    }
    $config = Set-TomlSectionValue -Content $config -Section "p2p" -Key "laddr" -Value "`"tcp://0.0.0.0:$($p.P2P)`""
    $rpcLaddr = if ($EnableRPC) { "`"tcp://0.0.0.0:$($p.RPC)`"" } else { "`"`"" }
    $config = Set-TomlSectionValue -Content $config -Section "rpc" -Key "laddr" -Value $rpcLaddr
    $config = Set-TomlSectionValue -Content $config -Section "rpc" -Key "pprof_laddr" -Value "`"localhost:$($p.Pprof)`""
    $config = $config -replace '(?m)^log_level = ".*"', "log_level = `"$LogLevel`""
    Set-Content -LiteralPath $configToml -Value $config

    $app = Get-Content -Raw -LiteralPath $appToml
    $apiEnable = if ($EnableAPI) { "true" } else { "false" }
    $grpcEnable = if ($EnableGRPC) { "true" } else { "false" }
    $app = Set-TomlSectionValue -Content $app -Section "api" -Key "enable" -Value $apiEnable
    $app = Set-TomlSectionValue -Content $app -Section "api" -Key "address" -Value "`"tcp://0.0.0.0:$($p.REST)`""
    $app = Set-TomlSectionValue -Content $app -Section "grpc" -Key "enable" -Value $grpcEnable
    $app = Set-TomlSectionValue -Content $app -Section "grpc" -Key "address" -Value "`"0.0.0.0:$($p.GRPC)`""
    $app = $app -replace '(?m)^minimum-gas-prices = ".*"', "minimum-gas-prices = `"$MinimumGasPrices`""
    Set-Content -LiteralPath $appToml -Value $app
  }
}

function Test-LocalnetTcpPortOpen {
  param(
    [string]$HostName = "127.0.0.1",
    [int]$Port,
    [int]$TimeoutMilliseconds = 500
  )

  $client = [System.Net.Sockets.TcpClient]::new()
  try {
    $async = $client.BeginConnect($HostName, $Port, $null, $null)
    if (-not $async.AsyncWaitHandle.WaitOne($TimeoutMilliseconds)) {
      return $false
    }
    $client.EndConnect($async)
    return $true
  } catch {
    return $false
  } finally {
    $client.Close()
  }
}

function Assert-LocalnetPortsAvailable {
  param(
    [int]$ValidatorCount,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseRESTPort = 1317,
    [int]$BaseGRPCPort = 9090,
    [int]$PortStride = 100,
    [bool]$EnableAPI = $true,
    [bool]$EnableGRPC = $true,
    [bool]$EnableRPC = $true
  )

  for ($i = 0; $i -lt $ValidatorCount; $i++) {
    $p = Get-LocalnetPortProfile -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -PortStride $PortStride
    $ports = @($p.P2P)
    if ($EnableRPC) { $ports += $p.RPC }
    if ($EnableAPI) { $ports += $p.REST }
    if ($EnableGRPC) { $ports += $p.GRPC }

    foreach ($port in $ports) {
      if (Test-LocalnetTcpPortOpen -Port $port) {
        throw "Port $port is already in use before starting node$i"
      }
    }
  }
}

function Invoke-LocalnetRpc {
  param(
    [int]$RPCPort,
    [string]$Path,
    [int]$TimeoutSeconds = 2
  )

  return Invoke-RestMethod -Uri "http://127.0.0.1:$RPCPort/$Path" -TimeoutSec $TimeoutSeconds
}

function Wait-LocalnetCondition {
  param(
    [scriptblock]$Condition,
    [int]$TimeoutSeconds,
    [string]$Description,
    [int]$PollMilliseconds = 500
  )

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  $lastError = $null
  while ((Get-Date) -lt $deadline) {
    try {
      $result = & $Condition
      if ($result) {
        return $result
      }
    } catch {
      $lastError = $_.Exception.Message
    }
    Start-Sleep -Milliseconds $PollMilliseconds
  }

  if ($lastError) {
    throw "Timed out waiting for $Description; last error: $lastError"
  }
  throw "Timed out waiting for $Description"
}

function Wait-LocalnetRpc {
  param(
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "RPC port $RPCPort" -Condition {
    $status = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "status"
    if ($status.result.node_info.network) { return $status }
    return $null
  }
}

function Get-LocalnetHeight {
  param([int]$RPCPort = 26657)

  $status = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "status"
  return [int64]$status.result.sync_info.latest_block_height
}

function Wait-LocalnetHeight {
  param(
    [int64]$TargetHeight,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "height $TargetHeight on RPC $RPCPort" -Condition {
    $height = Get-LocalnetHeight -RPCPort $RPCPort
    if ($height -ge $TargetHeight) { return $height }
    return $null
  }
}

function Wait-LocalnetValidators {
  param(
    [int]$ExpectedCount,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "$ExpectedCount validators on RPC $RPCPort" -Condition {
    $validators = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "validators?per_page=100"
    $count = @($validators.result.validators).Count
    if ($count -ne $ExpectedCount) { return $null }
    foreach ($validator in @($validators.result.validators)) {
      if ([int64]$validator.voting_power -le 0) { return $null }
    }
    return $validators
  }
}

function Get-LocalnetTotalVotingPower {
  param([int]$RPCPort = 26657)

  $validators = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "validators?per_page=100"
  $total = [int64]0
  foreach ($validator in @($validators.result.validators)) {
    $total += [int64]$validator.voting_power
  }
  return $total
}

function Wait-LocalnetTotalVotingPowerGreater {
  param(
    [int64]$PreviousPower,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "validator voting power greater than $PreviousPower on RPC $RPCPort" -Condition {
    $power = Get-LocalnetTotalVotingPower -RPCPort $RPCPort
    if ($power -gt $PreviousPower) { return $power }
    return $null
  }
}

function Wait-LocalnetPeers {
  param(
    [int]$ExpectedMinPeers,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "at least $ExpectedMinPeers peers on RPC $RPCPort" -Condition {
    $netInfo = Invoke-LocalnetRpc -RPCPort $RPCPort -Path "net_info"
    $peers = [int]$netInfo.result.n_peers
    if ($peers -ge $ExpectedMinPeers) { return $netInfo }
    return $null
  }
}

function Wait-LocalnetRest {
  param(
    [int]$RESTPort = 1317,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "REST endpoint $RESTPort" -Condition {
    $latest = Invoke-RestMethod -Uri "http://127.0.0.1:$RESTPort/cosmos/base/tendermint/v1beta1/blocks/latest" -TimeoutSec 2
    if ($latest.block.header.height) { return $latest }
    return $null
  }
}

function Wait-LocalnetGrpc {
  param(
    [int]$GRPCPort = 9090,
    [int]$TimeoutSeconds = 60
  )

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "gRPC TCP endpoint $GRPCPort" -Condition {
    return (Test-LocalnetTcpPortOpen -Port $GRPCPort -TimeoutMilliseconds 500)
  }
}

function Invoke-LocalnetCliJson {
  param(
    [string]$Binary,
    [string[]]$Arguments
  )

  $output = & $Binary @Arguments 2>&1
  if ($LASTEXITCODE -ne 0) {
    throw "orbitalisd command failed: $Binary $($Arguments -join ' ')`n$($output -join "`n")"
  }

  $text = $output -join "`n"
  $jsonStart = $text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    throw "orbitalisd command did not return JSON: $Binary $($Arguments -join ' ')`n$text"
  }
  return $text.Substring($jsonStart) | ConvertFrom-Json
}

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
    [string]$Denom = "norb",
    [int]$RPCPort = 26657
  )

  $node = "tcp://127.0.0.1:$RPCPort"
  $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "bank", "denom-metadata", $Denom, "--node", $node, "--output", "json")
  return $result.metadata
}

function Get-LocalnetBankSupplyOf {
  param(
    [string]$Binary,
    [string]$Denom = "norb",
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
    [string]$Denom = "norb",
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
    [string]$Amount = "5000000norb",
    [string]$Fees = "1000000norb",
    [string]$ChainId = "orbitalis-local-1",
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

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "staking delegate tx $txHash" -Condition {
    try {
      $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "tx", $txHash, "--node", $node, "--output", "json")
      if ($result.tx_response.code -eq 0 -or $result.code -eq 0) { return $result }
      throw "staking delegate tx failed with code $($result.tx_response.code)$($result.code)"
    } catch {
      return $null
    }
  }
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
    [string]$Amount = "1000norb",
    [string]$Fees = "1000000norb",
    [string]$ChainId = "orbitalis-local-1",
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

  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "bank send tx $txHash" -Condition {
    try {
      $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "tx", $txHash, "--node", $node, "--output", "json")
      if ($result.tx_response.code -eq 0 -or $result.code -eq 0) { return $result }
      throw "bank send tx failed with code $($result.tx_response.code)$($result.code)"
    } catch {
      return $null
    }
  }
}

function Stop-LocalnetProcesses {
  param(
    [string]$OutputDir,
    [string]$PidDir
  )

  if (Test-Path -LiteralPath $PidDir) {
    Get-ChildItem -LiteralPath $PidDir -Filter *.pid | ForEach-Object {
      $pidValue = [int](Get-Content -Raw -LiteralPath $_.FullName)
      $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      if ($proc) {
        Stop-Process -Id $pidValue -Force
        Write-Host "Stopped pid=$pidValue"
      }
      Remove-Item -LiteralPath $_.FullName -Force
    }
  }

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  $escaped = [regex]::Escape($resolved)
  Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
    Where-Object {
      $_.Name -like "orbitalisd*" -and
      $_.CommandLine -match $escaped
    } |
    ForEach-Object {
      $pidValue = [int]$_.ProcessId
      $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      if ($proc) {
        Stop-Process -Id $pidValue -Force
        Write-Host "Stopped orphan localnet pid=$pidValue"
      }
    }
}
