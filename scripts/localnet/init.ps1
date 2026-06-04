param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 3
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }
if ($Binary -eq "") { $Binary = Join-Path $RepoRoot "build\orbitalisd.exe" }
if ($ValidatorCount -lt 1) { throw "ValidatorCount must be at least 1" }

$Go = Join-Path $RepoRoot ".work\tools\go1.25.11\go\bin\go.exe"
if (!(Test-Path $Go)) { $Go = "go" }

New-Item -ItemType Directory -Force -Path (Split-Path $Binary) | Out-Null
& $Go build -o $Binary ./cmd/l1d

if (Test-Path $OutputDir) { Remove-Item -LiteralPath $OutputDir -Recurse -Force }
& $Binary testnet init-files `
  --validator-count $ValidatorCount `
  --output-dir $OutputDir `
  --chain-id orbitalis-local-1 `
  --staking-denom norb `
  --node-daemon-home orbitalisd `
  --node-dir-prefix node `
  --keyring-backend test `
  --single-host `
  --commit-timeout 1s `
  --minimum-gas-prices 0norb

function Get-PortProfile {
  param([int]$Index)

  return @{
    OldP2P  = 16656 + $Index
    OldRPC  = 26657 + $Index
    OldAPI  = 1317 + $Index
    OldGRPC = 9090 + $Index
    P2P     = 26656 + (100 * $Index)
    RPC     = 26657 + (100 * $Index)
    API     = 1317 + $Index
    GRPC    = 9090 + $Index
    Pprof   = 6060 + $Index
  }
}

for ($i = 0; $i -lt $ValidatorCount; $i++) {
  $nodeHome = Join-Path $OutputDir "node$i\orbitalisd"
  $configToml = Join-Path $nodeHome "config\config.toml"
  $appToml = Join-Path $nodeHome "config\app.toml"
  $p = Get-PortProfile -Index $i

  $config = Get-Content -Raw -LiteralPath $configToml
  for ($peer = 0; $peer -lt $ValidatorCount; $peer++) {
    $peerPorts = Get-PortProfile -Index $peer
    $config = $config -replace ":$($peerPorts.OldP2P)", ":$($peerPorts.P2P)"
  }
  $config = $config -replace ":$($p.OldRPC)", ":$($p.RPC)"
  $config = $config -replace 'pprof_laddr = "localhost:\d+"', "pprof_laddr = `"localhost:$($p.Pprof)`""
  Set-Content -LiteralPath $configToml -Value $config

  $app = Get-Content -Raw -LiteralPath $appToml
  $app = $app -replace ":$($p.OldAPI)", ":$($p.API)"
  $app = $app -replace ":$($p.OldGRPC)", ":$($p.GRPC)"
  $app = $app -replace 'minimum-gas-prices = ""', 'minimum-gas-prices = "0norb"'
  Set-Content -LiteralPath $appToml -Value $app
}

Write-Host "Initialized $ValidatorCount-node localnet at $OutputDir"
for ($i = 0; $i -lt $ValidatorCount; $i++) {
  $p = Get-PortProfile -Index $i
  Write-Host ("node{0}: p2p {1}, rpc {2}, grpc {3}, rest {4}" -f $i, $p.P2P, $p.RPC, $p.GRPC, $p.API)
}
