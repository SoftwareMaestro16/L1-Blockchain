param(
  [string]$Doc = "docs\architecture\acceptance-test-matrix.md",
  [string]$Catalog = "app\params\acceptance_test_matrix.go",
  [string]$Tests = "app\params\acceptance_test_matrix_test.go"
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
$catalogText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Catalog)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)

foreach ($term in @(
    'Acceptance Test Matrix',
    'Minimum acceptance matrix before public testnet',
    'Base node:',
    'boot single node',
    'boot multi-validator localnet',
    'restart',
    'export/import',
    'state sync or snapshot restore',
    'Staking:',
    'create validator',
    'delegate',
    'redelegate',
    'unbond',
    'withdraw rewards',
    'validator commission update',
    'Anti-centralization:',
    'validator reaches cap',
    'validator exceeds cap',
    'excess stake reward penalty applied',
    'top-N concentration query works',
    'commission floor enforced',
    'Slashing:',
    'downtime tracked',
    'downtime jail',
    'double-sign evidence path where feasible',
    'tombstone behavior',
    'delegator slash accounting',
    'Economics:',
    'inflation update',
    'fee burn',
    'treasury allocation',
    'rewards allocation',
    'APR query',
    'supply invariant',
    'AVM:',
    'upload code',
    'instantiate',
    'execute',
    'query',
    'migrate if enabled',
    'gas exhaustion contained',
    'Governance:',
    'valid param proposal',
    'invalid param proposal',
    'treasury proposal',
    'delayed critical param activation',
    'Observability:',
    'Prometheus metrics',
    'CLI queries',
    'gRPC queries',
    'events indexable',
    'DefaultAetraAcceptanceMatrixEvidence'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "acceptance matrix doc missing: $term"
}

foreach ($term in @(
    'AetraAcceptanceCategoryBaseNode',
    'AetraAcceptanceCategoryStaking',
    'AetraAcceptanceCategoryAntiCentralization',
    'AetraAcceptanceCategorySlashing',
    'AetraAcceptanceCategoryEconomics',
    'AetraAcceptanceCategoryAVM',
    'AetraAcceptanceCategoryGovernance',
    'AetraAcceptanceCategoryObservability',
    'AetraAcceptanceBaseNodeBootSingle',
    'AetraAcceptanceBaseNodeBootMultiValidator',
    'AetraAcceptanceBaseNodeRestart',
    'AetraAcceptanceBaseNodeExportImport',
    'AetraAcceptanceBaseNodeStateSyncSnapshotRestore',
    'AetraAcceptanceStakingCreateValidator',
    'AetraAcceptanceStakingDelegate',
    'AetraAcceptanceStakingRedelegate',
    'AetraAcceptanceStakingUnbond',
    'AetraAcceptanceStakingWithdrawRewards',
    'AetraAcceptanceStakingValidatorCommissionUpdate',
    'AetraAcceptanceAntiCentralizationValidatorReachesCap',
    'AetraAcceptanceAntiCentralizationValidatorExceedsCap',
    'AetraAcceptanceAntiCentralizationRewardPenalty',
    'AetraAcceptanceAntiCentralizationTopNQuery',
    'AetraAcceptanceAntiCentralizationCommissionFloor',
    'AetraAcceptanceSlashingDowntimeTracked',
    'AetraAcceptanceSlashingDowntimeJail',
    'AetraAcceptanceSlashingDoubleSignEvidence',
    'AetraAcceptanceSlashingTombstoneBehavior',
    'AetraAcceptanceSlashingDelegatorAccounting',
    'AetraAcceptanceEconomicsInflationUpdate',
    'AetraAcceptanceEconomicsFeeBurn',
    'AetraAcceptanceEconomicsTreasuryAllocation',
    'AetraAcceptanceEconomicsRewardsAllocation',
    'AetraAcceptanceEconomicsAPRQuery',
    'AetraAcceptanceEconomicsSupplyInvariant',
    'AetraAcceptanceAVMUploadCode',
    'AetraAcceptanceAVMInstantiate',
    'AetraAcceptanceAVMExecute',
    'AetraAcceptanceAVMQuery',
    'AetraAcceptanceAVMMigrateIfEnabled',
    'AetraAcceptanceAVMGasExhaustionContained',
    'AetraAcceptanceGovernanceValidParamProposal',
    'AetraAcceptanceGovernanceInvalidParamProposal',
    'AetraAcceptanceGovernanceTreasuryProposal',
    'AetraAcceptanceGovernanceDelayedCriticalActivation',
    'AetraAcceptanceObservabilityPrometheusMetrics',
    'AetraAcceptanceObservabilityCLIQueries',
    'AetraAcceptanceObservabilityGRPCQueries',
    'AetraAcceptanceObservabilityEventsIndexable',
    'AetraAcceptanceCategoryEvidence',
    'AetraAcceptanceMatrixReport',
    'DefaultAetraAcceptanceMatrixEvidence',
    'ValidateAetraAcceptanceMatrix',
    'BuildAetraAcceptanceMatrixReport'
  )) {
  Assert-Contains -Text $catalogText -Pattern ([regex]::Escape($term)) -Message "acceptance matrix catalog missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraAcceptanceMatrixCoversSection34',
    'TestAetraAcceptanceMatrixRejectsMissingCategory',
    'TestAetraAcceptanceMatrixRejectsMissingDuplicateAndUnexpectedScenarios',
    'TestAetraAcceptanceMatrixRejectsDuplicateCategory'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "acceptance matrix tests missing: $term"
}

Write-Host "acceptance test matrix doc test passed"
