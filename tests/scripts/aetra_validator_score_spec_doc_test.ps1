param(
  [string]$Doc = "docs\architecture\aetra-validator-score-spec.md",
  [string]$Policy = "app\params\aetra_validator_score_spec.go",
  [string]$Tests = "app\params\aetra_validator_score_spec_test.go",
  [string]$ModuleTypes = "x\aetra-validator-score\types\state.go",
  [string]$ModuleTypesTests = "x\aetra-validator-score\types\state_test.go",
  [string]$ModuleKeeperTests = "x\aetra-validator-score\keeper\keeper_test.go"
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
$moduleTypesText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ModuleTypes)
$moduleTypesTestText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ModuleTypesTests)
$moduleKeeperTestText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $ModuleKeeperTests)

foreach ($term in @(
    'x/aetra-validator-score Module Specification',
    'public accountability without subjective consensus control',
    '24. Module Specification: `x/aetra-validator-score`',
    '24.1 Responsibilities',
    'track validator uptime',
    'track missed block windows',
    'track jail history',
    'track slashing history',
    'track commission behavior',
    'track self-bond ratio',
    'track governance participation',
    'track concentration status',
    'produce public score',
    'expose explorer-friendly queries',
    'Score must not become a subjective censorship mechanism',
    'informational first and reward-affecting only when based on objective chain data',
    'The implementation gate is `app/params/aetra_validator_score_spec.go`',
    'BuildAetraValidatorScoreResponsibilitiesReport',
    'BuildAetraValidatorScoreSubjectiveControlReport',
    'ConsensusOverrideEnabled',
    'ObjectiveRewardModifierEnabled',
    'QueryValidatorScore',
    'QueryPublicValidatorMetrics',
    'QueryAllValidatorScores',
    '24.2 State',
    'Params:',
    'UptimeWindow',
    'UptimeWeightBps',
    'SlashHistoryWeightBps',
    'GovernanceWeightBps',
    'SelfBondWeightBps',
    'ConcentrationWeightBps',
    'MinScore',
    'MaxScore',
    'RewardModifierEnabled',
    'MaxRewardPenaltyBps',
    'ValidatorScore:',
    'OperatorAddress',
    'Score',
    'UptimeScore',
    'SlashScore',
    'GovernanceScore',
    'SelfBondScore',
    'ConcentrationScore',
    'MissedBlocks',
    'SignedBlocks',
    'JailCount',
    'SlashCount',
    'LastUpdatedHeight',
    'BuildAetraValidatorScoreStateSpecReport',
    'Current Implementation Mapping',
    'UptimeWindowBlocks',
    'ScoreWeights',
    'OverallScoreBps',
    'JailEvents',
    'SlashEventCount'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "aetra validator score spec doc missing: $term"
}

foreach ($term in @(
    'AetraValidatorScoreModuleName',
    'AetraValidatorScoreSpecEvidence',
    'AetraValidatorScoreSpecReport',
    'DefaultAetraValidatorScoreSpecEvidence',
    'ValidateAetraValidatorScoreSpec',
    'BuildAetraValidatorScoreSpecReport',
    'AetraValidatorScorePurposePublicAccountability',
    'AetraValidatorScoreResponsibilitiesEvidence',
    'AetraValidatorScoreResponsibilitiesReport',
    'DefaultAetraValidatorScoreResponsibilitiesEvidence',
    'ValidateAetraValidatorScoreResponsibilities',
    'BuildAetraValidatorScoreResponsibilitiesReport',
    'AetraValidatorScoreResponsibilityTrackUptime',
    'AetraValidatorScoreResponsibilityTrackMissedBlockWindows',
    'AetraValidatorScoreResponsibilityTrackJailHistory',
    'AetraValidatorScoreResponsibilityTrackSlashingHistory',
    'AetraValidatorScoreResponsibilityTrackCommissionBehavior',
    'AetraValidatorScoreResponsibilityTrackSelfBondRatio',
    'AetraValidatorScoreResponsibilityTrackGovernanceParticipation',
    'AetraValidatorScoreResponsibilityTrackConcentrationStatus',
    'AetraValidatorScoreResponsibilityProducePublicScore',
    'AetraValidatorScoreResponsibilityExplorerFriendlyQueries',
    'AetraValidatorScoreSubjectiveControlEvidence',
    'AetraValidatorScoreSubjectiveControlReport',
    'DefaultAetraValidatorScoreSubjectiveControlEvidence',
    'ValidateAetraValidatorScoreSubjectiveControl',
    'BuildAetraValidatorScoreSubjectiveControlReport',
    'AetraValidatorScoreGuardNoSubjectiveCensorship',
    'AetraValidatorScoreGuardInformationalFirst',
    'AetraValidatorScoreGuardObjectiveRewardOnly',
    'AetraValidatorScoreGuardConsensusOverrideDisabled',
    'AetraValidatorScoreGuardObjectiveInputsDeterministic',
    'AetraValidatorScoreStateSpecEvidence',
    'AetraValidatorScoreStateSpecReport',
    'DefaultAetraValidatorScoreStateSpecEvidence',
    'ValidateAetraValidatorScoreStateSpec',
    'BuildAetraValidatorScoreStateSpecReport',
    'AetraValidatorScoreStateParams',
    'AetraValidatorScoreStateValidatorScore',
    'AetraValidatorScoreStateParamUptimeWindow',
    'AetraValidatorScoreStateParamUptimeWeightBps',
    'AetraValidatorScoreStateParamSlashHistoryWeightBps',
    'AetraValidatorScoreStateParamGovernanceWeightBps',
    'AetraValidatorScoreStateParamSelfBondWeightBps',
    'AetraValidatorScoreStateParamConcentrationWeightBps',
    'AetraValidatorScoreStateParamMinScore',
    'AetraValidatorScoreStateParamMaxScore',
    'AetraValidatorScoreStateParamRewardModifierEnabled',
    'AetraValidatorScoreStateParamMaxRewardPenaltyBps',
    'AetraValidatorScoreStateScoreOperatorAddress',
    'AetraValidatorScoreStateScoreScore',
    'AetraValidatorScoreStateScoreUptimeScore',
    'AetraValidatorScoreStateScoreSlashScore',
    'AetraValidatorScoreStateScoreGovernanceScore',
    'AetraValidatorScoreStateScoreSelfBondScore',
    'AetraValidatorScoreStateScoreConcentrationScore',
    'AetraValidatorScoreStateScoreMissedBlocks',
    'AetraValidatorScoreStateScoreSignedBlocks',
    'AetraValidatorScoreStateScoreJailCount',
    'AetraValidatorScoreStateScoreSlashCount',
    'AetraValidatorScoreStateScoreLastUpdatedHeight'
  )) {
  Assert-Contains -Text $policyText -Pattern ([regex]::Escape($term)) -Message "aetra validator score spec policy missing: $term"
}

foreach ($term in @(
    'TestDefaultAetraValidatorScoreSpecCoversModulePurpose',
    'TestAetraValidatorScoreSpecRejectsMissingPurpose',
    'TestDefaultAetraValidatorScoreResponsibilitiesCoverSection241',
    'TestAetraValidatorScoreResponsibilitiesRejectMissingRequiredItems',
    'TestDefaultAetraValidatorScoreSubjectiveControlGuards',
    'TestAetraValidatorScoreSubjectiveControlRejectsMissingGuards',
    'TestDefaultAetraValidatorScoreStateSpecCoversSection242',
    'TestAetraValidatorScoreStateSpecRejectsMissingFields',
    'TestAetraValidatorScoreStateSpecRejectsDuplicateUnexpectedAndWrongModule'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "aetra validator score spec tests missing: $term"
}

foreach ($term in @(
    'UptimeScoreBps',
    'MissedBlockScoreBps',
    'JailScoreBps',
    'SlashHistoryScoreBps',
    'CommissionScoreBps',
    'SelfBondScoreBps',
    'GovernanceScoreBps',
    'DecentralizationScoreBps',
    'OverallScoreBps',
    'PublicValidatorMetrics',
    'ConsensusOverrideEnabled',
    'ConsensusOverrideAllowed',
    'ObjectiveRewardModifierEnabled',
    'InformationalOnly'
  )) {
  Assert-Contains -Text $moduleTypesText -Pattern ([regex]::Escape($term)) -Message "aetra validator score module types missing: $term"
}

foreach ($term in @(
    'TestUptimeAccountingScoresSignedWindow',
    'TestMissedBlockWindowRejectsImpossibleCounts',
    'TestSlashHistoryReducesScore',
    'TestJailHistoryReducesScoreAndRewardMultiplier',
    'TestGovernanceParticipationScore',
    'TestCommissionSelfBondAndConcentrationMetricsAffectScore',
    'TestPublicMetricsExposeExplorerFriendlyFields',
    'TestInformationalOnlyModeDisablesRewardEffectAndConsensusOverride',
    'TestDeterministicScoringIndependentOfInputOrder',
    'TestGenesisValidationRejectsConsensusOverrideDrift'
  )) {
  Assert-Contains -Text $moduleTypesTestText -Pattern ([regex]::Escape($term)) -Message "aetra validator score module tests missing: $term"
}

foreach ($term in @(
    'TestKeeperExportImportPreservesScores',
    'TestPublicValidatorMetricsQuery',
    'TestAllScoresQueryIsDeterministic'
  )) {
  Assert-Contains -Text $moduleKeeperTestText -Pattern ([regex]::Escape($term)) -Message "aetra validator score keeper tests missing: $term"
}

Write-Host "aetra validator score spec doc test passed"
