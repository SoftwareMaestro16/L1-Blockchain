param(
  [string]$OutputDir = "",
  [string]$Binary = ""
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }
if ($Binary -eq "") { $Binary = Join-Path $RepoRoot "build\orbitalisd.exe" }

$Go = Join-Path $RepoRoot ".work\tools\go1.25.11\go\bin\go.exe"
if (!(Test-Path $Go)) { $Go = "go" }

New-Item -ItemType Directory -Force -Path (Split-Path $Binary) | Out-Null
& $Go build -o $Binary ./cmd/l1d

if (Test-Path $OutputDir) { Remove-Item -LiteralPath $OutputDir -Recurse -Force }
& $Binary testnet init-files `
  --validator-count 3 `
  --output-dir $OutputDir `
  --chain-id orbitalis-local-1 `
  --staking-denom uorb `
  --node-daemon-home orbitalisd `
  --node-dir-prefix node `
  --keyring-backend test `
  --single-host `
  --commit-timeout 1s `
  --minimum-gas-prices 0uorb

$ports = @(
  @{OldP2P=16656; OldRPC=26657; OldAPI=1317; OldGRPC=9090; P2P=26656; RPC=26657; API=1317; GRPC=9090},
  @{OldP2P=16657; OldRPC=26658; OldAPI=1318; OldGRPC=9091; P2P=26756; RPC=26757; API=1318; GRPC=9091},
  @{OldP2P=16658; OldRPC=26659; OldAPI=1319; OldGRPC=9092; P2P=26856; RPC=26857; API=1319; GRPC=9092}
)

for ($i = 0; $i -lt 3; $i++) {
  $nodeHome = Join-Path $OutputDir "node$i\orbitalisd"
  $configToml = Join-Path $nodeHome "config\config.toml"
  $appToml = Join-Path $nodeHome "config\app.toml"
  $p = $ports[$i]

  $config = Get-Content -Raw -LiteralPath $configToml
  $config = $config -replace ":16656", ":26656"
  $config = $config -replace ":16657", ":26756"
  $config = $config -replace ":16658", ":26856"
  $config = $config -replace ":$($p.OldRPC)", ":$($p.RPC)"
  $config = $config -replace 'pprof_laddr = "localhost:\d+"', "pprof_laddr = `"localhost:$(6060 + $i)`""
  Set-Content -LiteralPath $configToml -Value $config

  $app = Get-Content -Raw -LiteralPath $appToml
  $app = $app -replace ":$($p.OldAPI)", ":$($p.API)"
  $app = $app -replace ":$($p.OldGRPC)", ":$($p.GRPC)"
  $app = $app -replace 'minimum-gas-prices = ""', 'minimum-gas-prices = "0uorb"'
  Set-Content -LiteralPath $appToml -Value $app
}

Write-Host "Initialized 3-node localnet at $OutputDir"
Write-Host "node0: p2p 26656, rpc 26657, grpc 9090, rest 1317"
Write-Host "node1: p2p 26756, rpc 26757, grpc 9091, rest 1318"
Write-Host "node2: p2p 26856, rpc 26857, grpc 9092, rest 1319"
