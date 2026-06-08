param(
  [string]$Doc = "docs\architecture\testing-requirements.md",
  [string]$Policy = "app\params\testing_requirements.go",
  [string]$Tests = "app\params\testing_requirements_test.go"
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
    'Testing Requirements',
    'Every implementation task must include tests',
    'A feature is not complete without tests',
    'Unit Tests',
    'keeper logic',
    'params validation',
    'math and accounting',
    'cap calculation',
    'slashing policy',
    'reward split',
    'inflation curve',
    'score calculation',
    'Integration Tests',
    'staking + custom staking policy',
    'slashing + validator score',
    'distribution + economics',
    'fee collector + burn + treasury',
    'nomination pool + delegation + unbonding',
    'governance param updates',
    'AVM tx flow',
    'E2E/Localnet Tests',
    'node startup',
    'validator creation',
    'delegation',
    'redelegation',
    'unbonding',
    'downtime scenario',
    'double-sign evidence scenario where feasible',
    'fee burn scenario',
    'AVM instantiate/execute/query',
    'export/import',
    'restart',
    'state sync/snapshot where feasible',
    'Adversarial Tests',
    'concentration attack simulation',
    'validator overflow stake simulation',
    'commission manipulation attempt',
    'invalid params proposal',
    'malformed evidence',
    'jailed validator reward attempt',
    'module account abuse attempt',
    'contract gas exhaustion',
    'contract storage abuse',
    'Performance Tests',
    '100 validator localnet/profile',
    '150-200 validator simulation/profile if feasible',
    'block time under load',
    'finality latency measurement',
    'mempool pressure',
    'AVM execution load',
    'state growth profile'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "testing requirements doc missing: $term"
}

foreach ($term in @(
    'Test Acceptance Rule',
    'No module should be considered production-ready unless',
    'unit tests pass',
    'integration tests pass',
    'genesis validation tests pass',
    'export/import tests pass',
    'deterministic restart tests pass',
    'adversarial tests for the relevant module pass',
    'CI runs the critical subset automatically'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "testing acceptance doc missing: $term"
}

foreach ($term in @(
    'TestingLayerRequirement',
    'FeatureTestingEvidence',
    'TestingRequirementsReport',
    'DefaultTestingRequirements',
    'ValidateTestingRequirements',
    'BuildTestingRequirementsReport',
    'ValidateFeatureTestingEvidence',
    'ModuleProductionReadinessEvidence',
    'ModuleProductionReadinessReport',
    'ValidateModuleProductionReadiness',
    'BuildModuleProductionReadinessReport',
    'ProductionAcceptanceUnitTestsPass',
    'ProductionAcceptanceIntegrationTestsPass',
    'ProductionAcceptanceGenesisValidationPass',
    'ProductionAcceptanceExportImportPass',
    'ProductionAcceptanceDeterministicRestart',
    'ProductionAcceptanceAdversarialModulePass',
    'ProductionAcceptanceCriticalCISubset',
    'TestLayerUnit',
    'TestLayerIntegration',
    'TestLayerE2ELocalnet',
    'TestLayerAdversarial',
    'TestLayerPerformance',
    'TestRequirementKeeperLogic',
    'TestRequirementParamsValidation',
    'TestRequirementMathAccounting',
    'TestRequirementCapCalculation',
    'TestRequirementSlashingPolicy',
    'TestRequirementRewardSplit',
    'TestRequirementInflationCurve',
    'TestRequirementScoreCalculation',
    'TestRequirementStakingCustomPolicy',
    'TestRequirementSlashingValidatorScore',
    'TestRequirementDistributionEconomics',
    'TestRequirementFeeCollectorBurnTreasury',
    'TestRequirementNominationDelegation',
    'TestRequirementGovernanceParamUpdates',
    'TestRequirementAVMTxFlow',
    'TestRequirementDoubleSignEvidence',
    'TestRequirementStateSyncSnapshot',
    'TestRequirementConcentrationAttack',
    'TestRequirementOverflowStake',
    'TestRequirementCommissionManipulation',
    'TestRequirementInvalidParamsProposal',
    'TestRequirementMalformedEvidence',
    'TestRequirementJailedRewardAttempt',
    'TestRequirementModuleAccountAbuse',
    'TestRequirementContractGasExhaustion',
    'TestRequirementContractStorageAbuse',
    'TestRequirementHundredValidatorProfile',
    'TestRequirementTwoHundredValidatorProfile',
    'TestRequirementBlockTimeUnderLoad',
    'TestRequirementFinalityLatencyMeasurement',
    'TestRequirementMempoolPressure',
    'TestRequirementAVMExecutionLoad',
    'TestRequirementStateGrowthProfile'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "testing requirements policy missing: $term"
}

foreach ($term in @(
    'TestDefaultTestingRequirementsCoverSection16Layers',
    'TestTestingRequirementsRejectMissingRequiredTest',
    'TestTestingRequirementsRejectMissingLayerScenarioAndDuplicates',
    'TestTestingRequirementsTreatFeasibleOptionalAsRequiredToImplement',
    'TestFeatureTestingEvidenceRequiresTestsForCompletion',
    'TestFeatureTestingEvidenceRejectsMissingIdentityOrImplementation',
    'TestModuleProductionReadinessRequiresAcceptanceRule',
    'TestModuleProductionReadinessRejectsMissingModuleNameAndCI'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "testing requirements tests missing: $term"
}

Write-Host "testing requirements doc test passed"
