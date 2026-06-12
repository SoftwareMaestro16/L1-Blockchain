param(
  [ValidateSet("3", "5", "10", "All")]
  [string]$ValidatorProfile = "All",
  [string]$Binary = "",
  [string]$ChainId = "aetra-testnet-preflight-1",
  [int]$TimeoutSeconds = 180,
  [switch]$SkipBuild,
  [switch]$SkipCosmWasmDisabledCheck
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
. (Join-Path $RepoRoot "scripts\localnet\common.ps1")

$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
if (-not $SkipBuild) {
  & (Join-Path $RepoRoot "scripts\build-aetrad.ps1") -Binary $Binary
} elseif (-not (Test-Path -LiteralPath $Binary)) {
  throw "Binary not found at $Binary and -SkipBuild was specified"
}

$profiles = if ($ValidatorProfile -eq "All") { @(3, 5, 10) } else { @([int]$ValidatorProfile) }

$profilePorts = @{
  3  = @{ P2P = 29656; RPC = 29657; REST = 4317; GRPC = 12090; Pprof = 9060 }
  5  = @{ P2P = 30656; RPC = 30657; REST = 5317; GRPC = 13090; Pprof = 10060 }
  10 = @{ P2P = 31656; RPC = 31657; REST = 6317; GRPC = 14090; Pprof = 11060 }
}

Push-Location $RepoRoot
try {
  foreach ($validators in $profiles) {
    $outputDir = Resolve-LocalnetPath -Path ".localnet-public-preflight-$validators" -DefaultRelativePath ".localnet-public-preflight-$validators"
    $ports = $profilePorts[$validators]
    Write-Host "Running public testnet preflight: validators=$validators output=$outputDir ports=@{$($ports.RPC),$($ports.P2P),$($ports.REST),$($ports.GRPC)}"

    & .\tests\e2e\prototype_acceptance.ps1 `
      -Profile Full `
      -OutputDir $outputDir `
      -Binary $Binary `
      -ChainId $ChainId `
      -ValidatorCount $validators `
      -TimeoutSeconds $TimeoutSeconds `
      -BaseP2PPort $ports.P2P `
      -BaseRPCPort $ports.RPC `
      -BaseRESTPort $ports.REST `
      -BaseGRPCPort $ports.GRPC `
      -BasePprofPort $ports.Pprof `
      -SkipBuild

    if (-not $SkipCosmWasmDisabledCheck) {
      & .\tests\e2e\cosmwasm_smoke.ps1 `
        -Binary $Binary `
        -Node "tcp://127.0.0.1:$($ports.RPC)"
    }
  }
} finally {
  foreach ($validators in $profiles) {
    $outputDir = Resolve-LocalnetPath -Path ".localnet-public-preflight-$validators" -DefaultRelativePath ".localnet-public-preflight-$validators"
    & .\scripts\localnet\stop.ps1 -OutputDir $outputDir
  }
  Pop-Location
}

Write-Host "Public testnet preflight passed for validator profile(s): $($profiles -join ',')"
