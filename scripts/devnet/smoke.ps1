param(
  [ValidateSet("1", "4")]
  [string]$Profile = "1",
  [string]$OutputDir = "",
  [string]$Binary = "",
  [switch]$SkipBuild,
  [switch]$SkipBankTx,
  [int]$TimeoutSeconds = 90
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

if ($Profile -eq "4") {
  $validatorCount = 4
  $chainId = "aetra-devnet-4"
  if ([string]::IsNullOrWhiteSpace($OutputDir)) { $OutputDir = ".devnet-4" }
  $baseP2P = 27656
  $baseRPC = 27657
  $baseREST = 1417
  $baseGRPC = 9190
} else {
  $validatorCount = 1
  $chainId = "aetra-devnet-1"
  if ([string]::IsNullOrWhiteSpace($OutputDir)) { $OutputDir = ".devnet-1" }
  $baseP2P = 26656
  $baseRPC = 26657
  $baseREST = 1317
  $baseGRPC = 9090
}

$args = @{
  OutputDir      = $OutputDir
  Binary         = $Binary
  ChainId        = $chainId
  ValidatorCount = $validatorCount
  BaseP2PPort    = $baseP2P
  BaseRPCPort    = $baseRPC
  BaseRESTPort   = $baseREST
  BaseGRPCPort   = $baseGRPC
  TimeoutSeconds = $TimeoutSeconds
  MinHeight      = 3
}
if ($SkipBuild) { $args.SkipBuild = $true }
if ($SkipBankTx -or $validatorCount -eq 1) { $args.SkipBankTx = $true }

& (Join-Path $RepoRoot "tests\e2e\localnet_smoke.ps1") @args
