param(
  [string]$ReadinessScript = "scripts\testnet\public-testnet-readiness-report.ps1",
  [string]$Gates = "docs\public-testnet-production-gates.md",
  [string]$Preparation = "docs\public-testnet-preparation.md",
  [string]$SmokeCommands = "docs\public-testnet-e2e-smoke-commands.md",
  [string]$LongRunningEvidence = "docs\public-testnet-long-running-evidence.md"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$scriptPath = Resolve-RepoPath $ReadinessScript
$json = & $scriptPath -OutputFormat Json -AllowFailures
$report = $json | ConvertFrom-Json

if ($report.status -ne "PASS" -and $report.status -ne "FAIL") {
  throw "readiness report has invalid status: $($report.status)"
}

foreach ($check in $report.checks) {
  if ($check.status -ne "PASS" -and $check.status -ne "FAIL") {
    throw "readiness check $($check.id) has invalid status: $($check.status)"
  }
  if ($check.status -eq "FAIL" -and [string]::IsNullOrWhiteSpace([string]$check.error)) {
    throw "readiness check $($check.id) failed without an error"
  }
}

$ids = @($report.checks | ForEach-Object { $_.id })
foreach ($id in @(
    "avm_runtime_wired",
    "native_account_wired",
    "direct_delegation_disabled",
    "official_pool_staking",
    "storage_rent_enforcement",
    "system_governance_safety",
    "app_invariants_registered",
    "export_import_roundtrip",
    "no_native_asset_modules",
    "docs_match_behavior",
    "localnet_profiles",
    "long_running_evidence",
    "e2e_smoke_commands"
  )) {
  if ($ids -notcontains $id) {
    throw "readiness report missing check id: $id"
  }
}

$gatesText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Gates)
$prepText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Preparation)
$smokeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $SmokeCommands)
$evidenceText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $LongRunningEvidence)

foreach ($term in @(
    "public-testnet-readiness-report.ps1",
    "prototype/spec state blocks public",
    "docs\public-testnet-e2e-smoke-commands.md",
    "docs\public-testnet-long-running-evidence.md"
  )) {
  Assert-Contains -Text $gatesText -Pattern ([regex]::Escape($term)) -Message "production gates doc missing readiness term: $term"
}

foreach ($term in @(
    "official liquid staking pool flows",
    "direct delegation rejection",
    "storage rent recovery behavior",
    "Token, NFT, and DEX-style behavior must be exercised through AVM contracts"
  )) {
  Assert-Contains -Text $prepText -Pattern ([regex]::Escape($term)) -Message "preparation doc missing readiness behavior: $term"
}

foreach ($term in @(
    "public-testnet-preflight.ps1 -ValidatorProfile 3",
    "public-testnet-preflight.ps1 -ValidatorProfile 5",
    "public-testnet-preflight.ps1 -ValidatorProfile 10",
    "export_import_smoke.ps1",
    "pos_smoke.ps1",
    "execution_os_smoke.ps1"
  )) {
  Assert-Contains -Text $smokeText -Pattern ([regex]::Escape($term)) -Message "smoke command doc missing: $term"
}

foreach ($term in @(
    "app_hash",
    "finality_seconds",
    "storage_rent_debt",
    "system_rent_runway",
    "pool_deposit_claim_unbond",
    "incident_count"
  )) {
  Assert-Contains -Text $evidenceText -Pattern ([regex]::Escape($term)) -Message "long-running evidence missing metric: $term"
}

Write-Host "public testnet readiness gate test passed"
