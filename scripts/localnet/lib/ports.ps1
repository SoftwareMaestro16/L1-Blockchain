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
    [string]$NodeDaemonHome = "aetrad"
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
    [string]$MinimumGasPrices = "0naet",
    [string]$LogLevel = "info"
  )

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  for ($i = 0; $i -lt $ValidatorCount; $i++) {
    $nodeHome = Join-Path $resolved "node$i\aetrad"
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
    $app = Set-TomlSectionValue -Content $app -Section "grpc" -Key "address" -Value "`"127.0.0.1:$($p.GRPC)`""
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

function Test-LocalnetTcpPortAvailable {
  param([int]$Port)

  if (Get-Command Get-NetTCPConnection -ErrorAction SilentlyContinue) {
    $existing = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue |
      Where-Object { $_.State -eq "Listen" }
    if ($existing) {
      return $false
    }
  }

  $listener = $null
  try {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Any, $Port)
    $listener.Start()
  } catch {
    return $false
  } finally {
    if ($null -ne $listener) {
      $listener.Stop()
    }
  }

  $v6Listener = $null
  try {
    $v6Listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::IPv6Any, $Port)
    $v6Listener.Start()
    return $true
  } catch {
    return $false
  } finally {
    if ($null -ne $v6Listener) {
      $v6Listener.Stop()
    }
  }

  return $true
}

function Assert-LocalnetPortsAvailable {
  param(
    [int]$ValidatorCount,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseRESTPort = 1317,
    [int]$BaseGRPCPort = 9090,
    [int]$BasePprofPort = 6060,
    [int]$PortStride = 100,
    [bool]$EnableAPI = $true,
    [bool]$EnableGRPC = $true,
    [bool]$EnableRPC = $true
  )

  for ($i = 0; $i -lt $ValidatorCount; $i++) {
    $p = Get-LocalnetPortProfile -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
    $ports = @($p.P2P, $p.Pprof)
    if ($EnableRPC) { $ports += $p.RPC }
    if ($EnableAPI) { $ports += $p.REST }
    if ($EnableGRPC) { $ports += $p.GRPC }

    foreach ($port in $ports) {
      if (-not (Test-LocalnetTcpPortAvailable -Port $port)) {
        throw "Port $port is already in use before starting node$i"
      }
    }
  }
}

function Repair-LocalnetProcessPathEnvironment {
  $pathValue = [Environment]::GetEnvironmentVariable("Path", "Process")
  if ([string]::IsNullOrEmpty($pathValue)) {
    $pathValue = [Environment]::GetEnvironmentVariable("PATH", "Process")
  }
  if (-not [string]::IsNullOrEmpty($pathValue)) {
    [Environment]::SetEnvironmentVariable("PATH", $null, "Process")
    [Environment]::SetEnvironmentVariable("Path", $pathValue, "Process")
  }
}
