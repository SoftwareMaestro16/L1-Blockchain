param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 3,
  [string]$ChainId = "aetra-local-1",
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
  [string]$Profile = "base",
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$RepoRoot = Get-LocalnetRepoRoot
$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
if ($ValidatorCount -lt 1) { throw "ValidatorCount must be at least 1" }
if ($PortStride -lt 1) { throw "PortStride must be at least 1" }
Assert-LocalnetProfile -Profile $Profile

if ($SkipBuild) {
  if (!(Test-Path -LiteralPath $Binary)) {
    throw "Binary not found at $Binary and -SkipBuild was specified"
  }
} else {
  & (Join-Path $RepoRoot "scripts\build-aetrad.ps1") -Binary $Binary
}

Remove-LocalnetDirectory -OutputDir $OutputDir
& $Binary testnet init-files `
  --validator-count $ValidatorCount `
  --output-dir $OutputDir `
  --chain-id $ChainId `
  --staking-denom naet `
  --node-daemon-home aetrad `
  --node-dir-prefix node `
  --keyring-backend test `
  --single-host `
  --commit-timeout $TimeoutCommit `
  --minimum-gas-prices 0naet

Set-LocalnetProfileGenesis -OutputDir $OutputDir -Profile $Profile
Write-LocalnetProfileManifest -OutputDir $OutputDir -Profile $Profile -ValidatorCount $ValidatorCount -ChainId $ChainId

Set-LocalnetGeneratedPorts `
  -OutputDir $OutputDir `
  -ValidatorCount $ValidatorCount `
  -BaseP2PPort $BaseP2PPort `
  -BaseRPCPort $BaseRPCPort `
  -BaseRESTPort $BaseRESTPort `
  -BaseGRPCPort $BaseGRPCPort `
  -BasePprofPort $BasePprofPort `
  -PortStride $PortStride `
  -EnableAPI $EnableAPI `
  -EnableGRPC $EnableGRPC `
  -EnableRPC $EnableRPC `
  -MinimumGasPrices "0naet" `
  -LogLevel $LogLevel

Write-Host "Initialized $ValidatorCount-node localnet for $ChainId at $OutputDir"
Write-Host "profile: $Profile"
for ($i = 0; $i -lt $ValidatorCount; $i++) {
  $p = Get-LocalnetPortProfile -Index $i -BaseP2PPort $BaseP2PPort -BaseRPCPort $BaseRPCPort -BaseRESTPort $BaseRESTPort -BaseGRPCPort $BaseGRPCPort -BasePprofPort $BasePprofPort -PortStride $PortStride
  Write-Host ("node{0}: p2p {1}, rpc {2}, grpc {3}, rest {4}" -f $i, $p.P2P, $p.RPC, $p.GRPC, $p.REST)
}
