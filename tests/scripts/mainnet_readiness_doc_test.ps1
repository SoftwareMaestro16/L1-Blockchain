param(
  [string]$Doc = "docs\architecture\mainnet-readiness.md",
  [string]$Policy = "app\params\mainnet_readiness.go",
  [string]$Tests = "app\params\mainnet_readiness_test.go"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) { return $Path }
  return Join-Path $RepoRoot $Path
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$policyText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Policy)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Mainnet Readiness Criteria',
    'Aetra should not be considered mainnet-ready until all criteria are met',
    'validator set policy implemented and tested',
    'effective power cap implemented and tested',
    'anti-concentration rewards implemented and tested',
    'dynamic inflation implemented and tested',
    'fee burn/treasury/reward split implemented and tested',
    'slashing configured and tested',
    'AVM integrated and tested',
    'export/import stable',
    'state sync/snapshots stable',
    'public testnet has run long enough to observe validator behavior',
    'load tests demonstrate finality target',
    'security audit completed',
    'critical findings fixed',
    'docs complete for validators, delegators, and contract developers',
    'Mainnet readiness is all-or-nothing',
    'completed security audit report',
    'evidence that all critical findings are fixed',
    'published docs for validators, delegators, and contract developers'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "mainnet readiness doc missing: $term"
}

foreach ($term in @(
    'MainnetReadinessEvidence',
    'MainnetReadinessReport',
    'ValidateMainnetReadiness',
    'BuildMainnetReadinessReport',
    'MainnetReadinessValidatorSetPolicy',
    'MainnetReadinessEffectivePowerCap',
    'MainnetReadinessAntiConcentrationRewards',
    'MainnetReadinessDynamicInflation',
    'MainnetReadinessFeeBurnTreasuryRewards',
    'MainnetReadinessSlashing',
    'MainnetReadinessAVM',
    'MainnetReadinessExportImport',
    'MainnetReadinessStateSyncSnapshots',
    'MainnetReadinessPublicTestnetDuration',
    'MainnetReadinessFinalityLoadTests',
    'MainnetReadinessSecurityAudit',
    'MainnetReadinessCriticalFindingsFixed',
    'MainnetReadinessDocsComplete'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "mainnet readiness policy missing: $term"
}

foreach ($term in @(
    'TestMainnetReadinessRequiresAllCriteria',
    'TestMainnetReadinessRejectsMissingConsensusAndEconomicsCriteria',
    'TestMainnetReadinessRejectsMissingOperationalAndSecurityCriteria',
    'completeMainnetReadinessEvidence'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "mainnet readiness tests missing: $term"
}

Write-Host "mainnet readiness doc test passed"
