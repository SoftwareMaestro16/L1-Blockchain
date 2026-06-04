$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$script:LocalnetNodeHomeName = "orbitalisd"
$script:DefaultChainId = "orbitalis-local-1"
$script:DefaultMinGasPrices = "0norb"
$script:DefaultFee = "1000000norb"

function Get-RepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function ConvertTo-AbsolutePath {
  param(
    [Parameter(Mandatory = $true)][string]$Path,
    [string]$BasePath = (Get-RepoRoot)
  )

  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $BasePath $Path))
}

function Resolve-LocalnetPath {
  param([string]$OutputDir = "")

  $repoRoot = Get-RepoRoot
  if ([string]::IsNullOrWhiteSpace($OutputDir)) {
    return Join-Path $repoRoot ".localnet"
  }
  return ConvertTo-AbsolutePath -Path $OutputDir -BasePath $repoRoot
}

function Assert-SafeLocalnetPath {
  param([Parameter(Mandatory = $true)][string]$Path)

  $repoRoot = (Get-RepoRoot).TrimEnd('\', '/')
  $fullPath = (ConvertTo-AbsolutePath -Path $Path).TrimEnd('\', '/')
  $separator = [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath -eq $repoRoot) {
    throw "refusing to operate on repository root: $fullPath"
  }
  if (-not $fullPath.StartsWith($repoRoot + $separator, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "refusing to operate outside repository root: $fullPath"
  }

  $leaf = Split-Path -Leaf $fullPath
  if (-not $leaf.StartsWith(".localnet", [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "refusing to operate on non-localnet path: $fullPath"
  }
}

function Get-DefaultBinary {
  return Join-Path (Get-RepoRoot) "build\orbitalisd.exe"
}

function Resolve-BinaryPath {
  param([string]$Binary = "")

  if ([string]::IsNullOrWhiteSpace($Binary)) {
    return Get-DefaultBinary
  }
  return ConvertTo-AbsolutePath -Path $Binary
}

function Get-GoBinary {
  $repoRoot = Get-RepoRoot
  $bundled = Join-Path $repoRoot ".work\tools\go1.25.11\go\bin\go.exe"
  if (Test-Path -LiteralPath $bundled) {
    return $bundled
  }
  return "go"
}

function Build-OrbitalisBinary {
  param([Parameter(Mandatory = $true)][string]$Binary)

  $go = Get-GoBinary
  New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Binary) | Out-Null
  Push-Location (Get-RepoRoot)
  try {
    & $go build -o $Binary .\cmd\l1d
    if ($LASTEXITCODE -ne 0) {
      throw "go build failed with exit code $LASTEXITCODE"
    }
  }
  finally {
    Pop-Location
  }
}

function Get-NodePorts {
  param(
    [Parameter(Mandatory = $true)][int]$Index,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseGRPCPort = 9090,
    [int]$BaseRESTPort = 1317,
    [int]$BasePprofPort = 6060,
    [int]$BaseMetricsPort = 26660,
    [int]$BaseAppMetricsPort = 27660
  )

  return [pscustomobject]@{
    index       = $Index
    p2p         = $BaseP2PPort + ($Index * 100)
    rpc         = $BaseRPCPort + ($Index * 100)
    grpc        = $BaseGRPCPort + $Index
    rest        = $BaseRESTPort + $Index
    pprof       = $BasePprofPort + $Index
    metrics     = $BaseMetricsPort + ($Index * 100)
    app_metrics = $BaseAppMetricsPort + ($Index * 100)
  }
}

function Get-NodeHome {
  param(
    [Parameter(Mandatory = $true)][string]$OutputDir,
    [Parameter(Mandatory = $true)][int]$Index
  )

  return Join-Path $OutputDir ("node{0}\{1}" -f $Index, $script:LocalnetNodeHomeName)
}

function Get-ManifestPath {
  param([Parameter(Mandatory = $true)][string]$OutputDir)
  return Join-Path $OutputDir "localnet.json"
}

function Write-LocalnetManifest {
  param(
    [Parameter(Mandatory = $true)][string]$OutputDir,
    [Parameter(Mandatory = $true)][int]$ValidatorCount,
    [Parameter(Mandatory = $true)][string]$ChainId,
    [int]$BaseP2PPort = 26656,
    [int]$BaseRPCPort = 26657,
    [int]$BaseGRPCPort = 9090,
    [int]$BaseRESTPort = 1317,
    [int]$BasePprofPort = 6060,
    [int]$BaseMetricsPort = 26660,
    [int]$BaseAppMetricsPort = 27660
  )

  $nodes = @()
  for ($i = 0; $i -lt $ValidatorCount; $i++) {
    $ports = Get-NodePorts -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseGRPCPort $BaseGRPCPort -BaseRESTPort $BaseRESTPort -BasePprofPort $BasePprofPort -BaseMetricsPort $BaseMetricsPort -BaseAppMetricsPort $BaseAppMetricsPort
    $nodes += [pscustomobject]@{
      name    = "node$i"
      home    = Get-NodeHome -OutputDir $OutputDir -Index $i
      ports   = $ports
      rpc_url = "tcp://127.0.0.1:$($ports.rpc)"
      rest_url = "http://127.0.0.1:$($ports.rest)"
      grpc_url = "127.0.0.1:$($ports.grpc)"
      metrics_url = "http://127.0.0.1:$($ports.metrics)/metrics"
      app_metrics_url = "http://127.0.0.1:$($ports.app_metrics)/metrics"
    }
  }

  $manifest = [pscustomobject]@{
    chain_id = $ChainId
    validator_count = $ValidatorCount
    node_home_name = $script:LocalnetNodeHomeName
    min_gas_prices = $script:DefaultMinGasPrices
    nodes = $nodes
  }
  $manifest | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Get-ManifestPath -OutputDir $OutputDir)
}

function Read-LocalnetManifest {
  param([Parameter(Mandatory = $true)][string]$OutputDir)

  $path = Get-ManifestPath -OutputDir $OutputDir
  if (!(Test-Path -LiteralPath $path)) {
    return $null
  }
  return (Get-Content -Raw -LiteralPath $path | ConvertFrom-Json)
}

function Get-ManifestValidatorCount {
  param(
    [Parameter(Mandatory = $true)][string]$OutputDir,
    [int]$ValidatorCount = 0
  )

  if ($ValidatorCount -gt 0) {
    return $ValidatorCount
  }
  $manifest = Read-LocalnetManifest -OutputDir $OutputDir
  if ($null -ne $manifest) {
    return [int]$manifest.validator_count
  }
  return 3
}

function Invoke-ExternalChecked {
  param(
    [Parameter(Mandatory = $true)][string]$FilePath,
    [Parameter(Mandatory = $true)][string[]]$Arguments,
    [string]$LogPath = "",
    [string]$FailureMessage = "command failed",
    [switch]$EchoOutput,
    [switch]$SuppressFailureOutput
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $FilePath @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  }
  finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  if (-not [string]::IsNullOrWhiteSpace($LogPath)) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $LogPath) | Out-Null
    $output | Set-Content -LiteralPath $LogPath
  }
  if ($EchoOutput) {
    $output | ForEach-Object { Write-Host $_ }
  }
  if ($exitCode -ne 0) {
    if ($SuppressFailureOutput) {
      throw "$FailureMessage (exit $exitCode). See $LogPath"
    }
    $tail = ($output | Select-Object -Last 20) -join "`n"
    throw "$FailureMessage (exit $exitCode):`n$tail"
  }
  return $output
}

function Set-PrivateLocalnetDirectory {
  param([Parameter(Mandatory = $true)][string]$Path)

  if (!(Test-Path -LiteralPath $Path)) {
    return
  }
  if ($env:OS -ne "Windows_NT") {
    & chmod 700 $Path 2>$null
  }
}
