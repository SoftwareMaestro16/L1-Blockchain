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

Push-Location $RepoRoot
try {
  foreach ($validators in $profiles) {
    $outputDir = Resolve-LocalnetPath -Path ".localnet-public-preflight-$validators" -DefaultRelativePath ".localnet-public-preflight-$validators"
    $baseOffset = switch ($validators) {
      3 { 3000 }
      5 { 4000 }
      10 { 5000 }
      default { 6000 }
    }
    Write-Host "Running public testnet preflight: validators=$validators output=$outputDir"

    & .\tests\e2e\prototype_acceptance.ps1 `
      -Profile Full `
      -OutputDir $outputDir `
      -Binary $Binary `
      -ChainId $ChainId `
      -ValidatorCount $validators `
      -TimeoutSeconds $TimeoutSeconds `
      -BaseP2PPort (26656 + $baseOffset) `
      -BaseRPCPort (26657 + $baseOffset) `
      -BaseRESTPort (1317 + $baseOffset) `
      -BaseGRPCPort (9090 + $baseOffset) `
      -BasePprofPort (6060 + $baseOffset) `
      -SkipBuild

    if (-not $SkipCosmWasmDisabledCheck) {
      & .\tests\e2e\cosmwasm_smoke.ps1 `
        -Binary $Binary `
        -Node "tcp://127.0.0.1:$((26657 + $baseOffset))"
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
