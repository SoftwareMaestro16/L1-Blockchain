param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [ValidateRange(1, 100)][int]$ValidatorCount = 3,
  [string]$ChainId = "orbitalis-local-1",
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseGRPCPort = 9090,
  [int]$BaseRESTPort = 1317,
  [int]$BasePprofPort = 6060,
  [int]$BaseMetricsPort = 26660,
  [int]$BaseAppMetricsPort = 27660,
  [switch]$DebugSecrets
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0
. (Join-Path $PSScriptRoot "common.ps1")

$RepoRoot = Get-RepoRoot
$OutputDir = Resolve-LocalnetPath -OutputDir $OutputDir
$Binary = Resolve-BinaryPath -Binary $Binary
Assert-SafeLocalnetPath -Path $OutputDir

if (Test-Path -LiteralPath $OutputDir) {
  & (Join-Path $PSScriptRoot "stop.ps1") -OutputDir $OutputDir
  Remove-Item -LiteralPath $OutputDir -Recurse -Force
}

Build-OrbitalisBinary -Binary $Binary
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
Set-PrivateLocalnetDirectory -Path $OutputDir
$logDir = Join-Path $OutputDir "logs"
New-Item -ItemType Directory -Force -Path $logDir | Out-Null

$initArgs = @(
  "testnet", "init-files",
  "--validator-count", "$ValidatorCount",
  "--output-dir", $OutputDir,
  "--chain-id", $ChainId,
  "--staking-denom", "norb",
  "--node-daemon-home", $script:LocalnetNodeHomeName,
  "--node-dir-prefix", "node",
  "--keyring-backend", "test",
  "--single-host",
  "--commit-timeout", "1s",
  "--minimum-gas-prices", $script:DefaultMinGasPrices
)

$initLog = Join-Path $logDir "init.log"
Invoke-ExternalChecked -FilePath $Binary -Arguments $initArgs -LogPath $initLog -FailureMessage "testnet init-files failed" -SuppressFailureOutput:(!$DebugSecrets) -EchoOutput:$DebugSecrets | Out-Null

for ($i = 0; $i -lt $ValidatorCount; $i++) {
  $nodeHome = Get-NodeHome -OutputDir $OutputDir -Index $i
  $configToml = Join-Path $nodeHome "config\config.toml"
  $appToml = Join-Path $nodeHome "config\app.toml"
  $ports = Get-NodePorts -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseGRPCPort $BaseGRPCPort -BaseRESTPort $BaseRESTPort -BasePprofPort $BasePprofPort -BaseMetricsPort $BaseMetricsPort -BaseAppMetricsPort $BaseAppMetricsPort

  $config = Get-Content -Raw -LiteralPath $configToml
  for ($j = 0; $j -lt $ValidatorCount; $j++) {
    $oldP2P = 16656 + $j
    $newP2P = $BaseP2PPort + ($j * 100)
    $config = $config -replace ":$oldP2P", ":$newP2P"
  }
  $oldRPC = 26657 + $i
  $config = $config -replace ":$oldRPC", ":$($ports.rpc)"
  $config = $config -replace 'pprof_laddr = "localhost:\d+"', "pprof_laddr = `"localhost:$($ports.pprof)`""
  $config = $config -replace 'prometheus = false', 'prometheus = true'
  $config = $config -replace 'prometheus_listen_addr = ":[0-9]+"', "prometheus_listen_addr = `":$($ports.metrics)`""
  Set-Content -LiteralPath $configToml -Value $config

  $app = Get-Content -Raw -LiteralPath $appToml
  $oldAPI = 1317 + $i
  $oldGRPC = 9090 + $i
  $app = $app -replace ":$oldAPI", ":$($ports.rest)"
  $app = $app -replace ":$oldGRPC", ":$($ports.grpc)"
  $app = $app -replace 'minimum-gas-prices = ".*"', "minimum-gas-prices = `"$($script:DefaultMinGasPrices)`""
  Set-Content -LiteralPath $appToml -Value $app
}

Write-LocalnetManifest `
  -OutputDir $OutputDir `
  -ValidatorCount $ValidatorCount `
  -ChainId $ChainId `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BasePprofPort $BasePprofPort `
  -BaseMetricsPort $BaseMetricsPort `
  -BaseAppMetricsPort $BaseAppMetricsPort

Write-Host "Initialized $ValidatorCount-node localnet at $OutputDir"
for ($i = 0; $i -lt $ValidatorCount; $i++) {
  $ports = Get-NodePorts -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseGRPCPort $BaseGRPCPort -BaseRESTPort $BaseRESTPort -BasePprofPort $BasePprofPort -BaseMetricsPort $BaseMetricsPort -BaseAppMetricsPort $BaseAppMetricsPort
  Write-Host ("node{0}: p2p {1}, rpc {2}, grpc {3}, rest {4}, comet metrics {5}, app metrics {6}" -f $i, $ports.p2p, $ports.rpc, $ports.grpc, $ports.rest, $ports.metrics, $ports.app_metrics)
}
if (!$DebugSecrets) {
  Write-Host "Init output captured at $initLog; rerun with -DebugSecrets only when mnemonic output is explicitly needed."
}
