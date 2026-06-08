param(
  [string]$Doc = "docs\architecture\governance-parameters.md",
  [string]$Policy = "app\params\governance_parameters.go",
  [string]$Tests = "app\params\governance_parameters_test.go"
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
    'Governance and Parameters',
    '27. Governance Specification',
    'Governance must be powerful enough to tune the network',
    '27.1 Governance-Controlled Modules',
    'staking policy params',
    'economics params',
    'validator score params',
    'slashing params within bounds',
    'AVM contract upload policy',
    'treasury spend',
    'validator set growth schedule',
    'block gas/size within safe bounds',
    '`staking_policy`: validator set size, validator power cap',
    'inflation min/max',
    'target bonded ratio',
    'fee split',
    '`validator_score`: validator score policy',
    '`slashing`: double-sign slash, downtime slash, downtime window',
    '`validator_set_growth`: validator set growth schedule',
    '`consensus`: block gas limit and block max bytes',
    '27.2 Param Safety Bounds',
    'type',
    'default value',
    'min value',
    'max value',
    'authority',
    'whether change is immediate or epoch-delayed',
    'event emitted on change',
    'tests for invalid update',
    'Critical params should apply only at epoch boundary',
    'treasury spend policy',
    'params must have min/max validation',
    'unsafe params must be rejected at proposal execution',
    'genesis validation must reject invalid params',
    'parameter changes must emit events',
    'critical changes should use longer voting period or higher quorum',
    'every governed parameter has authority metadata',
    'every governed parameter has default value metadata',
    'every governed parameter declares immediate or epoch-delayed application',
    'critical parameters apply at epoch boundary',
    '27.3 Governance Tests',
    'valid param proposal executes',
    'invalid param proposal rejected',
    'unauthorized authority rejected',
    'emergency unsafe value rejected',
    'epoch-delayed param activation',
    'event emitted',
    'query reflects new params',
    'export/import after param change',
    'BuildGovernanceTestingReport',
    'DefaultGovernanceTestingEvidence',
    'ValidateGovernanceTestingEvidence',
    'unsafe values fail closed',
    'critical params are not activated mid-epoch'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "governance parameters doc missing: $term"
}

foreach ($term in @(
    'GovernanceParameterSpec',
    'GovernanceParamChange',
    'GovernanceParameterSafetyReport',
    'DefaultGovernanceParameterSpecs',
    'DefaultGovernanceGenesisParams',
    'ValidateGovernanceParameterSpecs',
    'ValidateGovernanceParamChange',
    'ValidateGovernanceGenesisParams',
    'GovernanceParamValidatorSetSize',
    'GovernanceParamValidatorPowerCap',
    'GovernanceParamCommissionFloor',
    'GovernanceParamCommissionMax',
    'GovernanceParamCommissionMaxChange',
    'GovernanceParamInflationMin',
    'GovernanceParamInflationMax',
    'GovernanceParamTargetBondedRatio',
    'GovernanceParamFeeBurnShare',
    'GovernanceParamFeeRewardShare',
    'GovernanceParamFeeTreasuryShare',
    'GovernanceParamDoubleSignSlash',
    'GovernanceParamDowntimeSlash',
    'GovernanceParamDowntimeWindow',
    'GovernanceParamAVMContractUploadPolicy',
    'GovernanceParamTreasurySpendPolicy',
    'GovernanceParamValidatorScorePolicy',
    'GovernanceParamValidatorSetGrowth',
    'GovernanceParamBlockGasLimit',
    'GovernanceParamBlockMaxBytes',
    'GovernanceAuthorityGovModule',
    'DefaultInt',
    'DefaultString',
    'Authority',
    'ApplyEpochDelay',
    'EventType',
    'InvalidUpdateTest',
    'GovernanceTestingEvidence',
    'GovernanceTestingReport',
    'DefaultGovernanceTestingEvidence',
    'ValidateGovernanceTestingEvidence',
    'BuildGovernanceTestingReport',
    'GovernanceTestValidParamProposalExecutes',
    'GovernanceTestInvalidParamRejected',
    'GovernanceTestUnauthorizedAuthority',
    'GovernanceTestEmergencyUnsafeRejected',
    'GovernanceTestEpochDelayedActivation',
    'GovernanceTestEventEmitted',
    'GovernanceTestQueryReflectsNewParams',
    'GovernanceTestExportImportAfterChange',
    'GovernanceCriticalVotingPeriodBlocks',
    'GovernanceCriticalQuorumBps'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "governance parameters policy missing: $term"
}

foreach ($term in @(
    'TestDefaultGovernanceParameterSpecsCoverSection13',
    'TestGovernanceControlledModulesCoverSection271',
    'TestGovernanceParamSpecsCarrySection272Metadata',
    'TestGovernanceParamChangeRejectsUnsafeExecution',
    'TestCriticalGovernanceParamChangesRequireLongerVotingAndHigherQuorum',
    'TestNonCriticalGovernanceParamChangeUsesNormalVotingBounds',
    'TestGovernanceGenesisValidationRejectsInvalidParams',
    'TestGovernanceEnumParamsAreBounded',
    'TestGovernanceSafetyReportDetectsMissingBoundsGenesisAndEvents',
    'TestGovernanceSafetyReportDetectsMissingSection272Metadata',
    'TestDefaultGovernanceTestingEvidenceCoversSection273',
    'TestGovernanceTestingEvidenceRejectsMissingRequiredTests'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "governance parameters tests missing: $term"
}

Write-Host "governance parameters doc test passed"
