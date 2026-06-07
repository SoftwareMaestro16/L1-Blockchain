package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraStakingPolicySpecCoversResponsibilities(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraStakingPolicyModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 15, report.Required)
	require.NoError(t, ValidateAetraStakingPolicySpec(evidence))
}

func TestAetraStakingPolicySpecRejectsMissingStakeAndPowerResponsibilities(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.CalculatesRawValidatorStake = false
	evidence.CalculatesEffectiveValidatorStake = false
	evidence.CalculatesOverflowStake = false
	evidence.EnforcesOrExposesEffectiveVotingPowerCap = false
	evidence.CalculatesOverflowRewardMultiplier = false

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityRawStake)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityEffectiveStake)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityOverflowStake)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityEffectiveVotingPowerCap)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityOverflowRewardMultiplier)
	require.Error(t, ValidateAetraStakingPolicySpec(evidence))
}

func TestAetraStakingPolicySpecRejectsMissingCommissionConcentrationAndSafetyResponsibilities(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.ExposesDelegationConcentrationWarnings = false
	evidence.EnforcesCommissionFloor = false
	evidence.EnforcesMaxCommission = false
	evidence.EnforcesMaxCommissionChangeRate = false
	evidence.ExposesTopNConcentrationMetrics = false
	evidence.ValidatesGovernanceParamChanges = false
	evidence.EmitsCapOverflowCommissionPolicyEvents = false
	evidence.RemainsDeterministicAndExportImportSafe = false

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityDelegationConcentrationWarning)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityCommissionFloor)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityMaxCommission)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityMaxCommissionChangeRate)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityTopNConcentrationMetrics)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityGovernanceParamValidation)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityPolicyChangeEvents)
	require.Contains(t, report.Failed, AetraStakingPolicyResponsibilityDeterministicExportImport)
	require.Error(t, ValidateAetraStakingPolicySpec(evidence))
}

func TestAetraStakingPolicySpecRejectsMissingPurposeAndCentralModuleRole(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.PurposeEffectivePowerOverflowCommissionAntiConcentration = false
	evidence.CentralAntiCentralizationModule = false

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyPurposeEffectivePowerOverflowCommissionAntiConcentration)
	require.Contains(t, report.Failed, AetraStakingPolicyCentralAntiCentralizationModule)
}

func TestAetraStakingPolicySpecRejectsWrongModuleIdentity(t *testing.T) {
	evidence := DefaultAetraStakingPolicySpecEvidence()
	evidence.ModuleName = "x/staking-policy"

	report := BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)

	evidence.ModuleName = ""
	report = BuildAetraStakingPolicySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
}

func TestDefaultAetraStakingPolicyStateSpecCoversSuggestedState(t *testing.T) {
	evidence := DefaultAetraStakingPolicyStateSpecEvidence()

	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraStakingPolicyModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 31, report.Required)
	require.NoError(t, ValidateAetraStakingPolicyStateSpec(evidence))
}

func TestAetraStakingPolicyStateSpecRejectsMissingParamsFields(t *testing.T) {
	evidence := DefaultAetraStakingPolicyStateSpecEvidence()
	evidence.ParamsFields = removeString(evidence.ParamsFields,
		AetraStakingPolicyStateParamMaxValidatorsSoftTarget,
		AetraStakingPolicyStateParamValidatorPowerCapBps,
		AetraStakingPolicyStateParamTop10TargetBps,
		AetraStakingPolicyStateParamMinSelfBond,
	)

	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyStateParams+"."+AetraStakingPolicyStateParamMaxValidatorsSoftTarget+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateParams+"."+AetraStakingPolicyStateParamValidatorPowerCapBps+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateParams+"."+AetraStakingPolicyStateParamTop10TargetBps+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateParams+"."+AetraStakingPolicyStateParamMinSelfBond+":missing")
	require.Error(t, ValidateAetraStakingPolicyStateSpec(evidence))
}

func TestAetraStakingPolicyStateSpecRejectsMissingValidatorPolicyAndSnapshotFields(t *testing.T) {
	evidence := DefaultAetraStakingPolicyStateSpecEvidence()
	evidence.ValidatorPolicyFields = removeString(evidence.ValidatorPolicyFields,
		AetraStakingPolicyStateValidatorRawBondedTokens,
		AetraStakingPolicyStateValidatorIsOverCap,
		AetraStakingPolicyStateValidatorLastCommissionChangeTime,
	)
	evidence.ConcentrationSnapshotFields = removeString(evidence.ConcentrationSnapshotFields,
		AetraStakingPolicyStateSnapshotBondedRatio,
		AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate,
	)

	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyStateValidatorPolicy+"."+AetraStakingPolicyStateValidatorRawBondedTokens+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateValidatorPolicy+"."+AetraStakingPolicyStateValidatorIsOverCap+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateValidatorPolicy+"."+AetraStakingPolicyStateValidatorLastCommissionChangeTime+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateConcentrationSnapshot+"."+AetraStakingPolicyStateSnapshotBondedRatio+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyStateConcentrationSnapshot+"."+AetraStakingPolicyStateSnapshotNakamotoCoefficientEstimate+":missing")
}

func TestAetraStakingPolicyStateSpecRejectsUnsafeDecimalAccounting(t *testing.T) {
	evidence := DefaultAetraStakingPolicyStateSpecEvidence()
	evidence.IntegerBasisPointsOrSDKDecimals = false
	evidence.AvoidsFloatingPoint = false

	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyStateIntegerBpsOrSDKDecimal)
	require.Contains(t, report.Failed, AetraStakingPolicyStateNoFloatingPoint)
	require.Error(t, ValidateAetraStakingPolicyStateSpec(evidence))
}

func TestAetraStakingPolicyStateSpecRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraStakingPolicyStateSpecEvidence()
	evidence.ModuleName = "x/other"
	evidence.ParamsFields = append(evidence.ParamsFields, AetraStakingPolicyStateParamTop20TargetBps)
	evidence.ValidatorPolicyFields = append(evidence.ValidatorPolicyFields, "FloatingPointRewardShare")

	report := BuildAetraStakingPolicyStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	require.Contains(t, report.Failed, AetraStakingPolicyStateParams+"."+AetraStakingPolicyStateParamTop20TargetBps+":duplicate")
	require.Contains(t, report.Failed, AetraStakingPolicyStateValidatorPolicy+".FloatingPointRewardShare:unexpected")

	evidence = DefaultAetraStakingPolicyStateSpecEvidence()
	evidence.ModuleName = ""
	report = BuildAetraStakingPolicyStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
}

func TestDefaultAetraStakingPolicyParameterRulesMatchSection223(t *testing.T) {
	rules := DefaultAetraStakingPolicyParameterRuleSet()

	require.Equal(t, int64(100), rules.ValidatorPowerCapBps.MinBps)
	require.Equal(t, int64(500), rules.ValidatorPowerCapBps.MaxBps)
	require.Equal(t, int64(200), rules.ValidatorPowerCapBps.RecommendedMin)
	require.Equal(t, int64(300), rules.ValidatorPowerCapBps.RecommendedMax)

	require.Equal(t, int64(0), rules.OverflowRewardMultiplierBps.MinBps)
	require.Equal(t, int64(10_000), rules.OverflowRewardMultiplierBps.MaxBps)
	require.Equal(t, int64(0), rules.OverflowRewardMultiplierBps.RecommendedMin)
	require.Equal(t, int64(3_000), rules.OverflowRewardMultiplierBps.RecommendedMax)

	require.Equal(t, int64(0), rules.CommissionFloorBps.MinBps)
	require.Equal(t, int64(1_000), rules.CommissionFloorBps.MaxBps)
	require.Equal(t, int64(300), rules.CommissionFloorBps.RecommendedMin)
	require.Equal(t, int64(500), rules.CommissionFloorBps.RecommendedMax)

	require.Equal(t, int64(0), rules.CommissionMaxBps.MinBps)
	require.Equal(t, int64(3_000), rules.CommissionMaxBps.MaxBps)
	require.Equal(t, int64(1_500), rules.CommissionMaxBps.RecommendedMin)
	require.Equal(t, int64(2_000), rules.CommissionMaxBps.RecommendedMax)

	require.Equal(t, int64(1), rules.CommissionMaxDailyChangeBps.MinBps)
	require.Equal(t, int64(500), rules.CommissionMaxDailyChangeBps.MaxBps)
	require.Equal(t, int64(50), rules.CommissionMaxDailyChangeBps.RecommendedMin)
	require.Equal(t, int64(100), rules.CommissionMaxDailyChangeBps.RecommendedMax)
}

func TestDefaultAetraStakingPolicyParameterValuesPassValidation(t *testing.T) {
	values := DefaultAetraStakingPolicyParameterValues()

	report := BuildAetraStakingPolicyParameterReport(values)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.NoError(t, ValidateAetraStakingPolicyParameterValues(values))
}

func TestAetraStakingPolicyParameterRulesRejectUnsafeGovernanceValues(t *testing.T) {
	values := DefaultAetraStakingPolicyParameterValues()
	values.ValidatorPowerCapBps = 99
	values.CommissionFloorBps = 400
	values.CommissionMaxBps = 399
	values.OverflowRewardMultiplierBps = 10_001

	report := BuildAetraStakingPolicyParameterReport(values)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyParamValidatorPowerCapBps)
	require.Contains(t, report.Failed, AetraStakingPolicyParamCommissionMaxBps)
	require.Contains(t, report.Failed, AetraStakingPolicyParamOverflowRewardMultiplierBps)
	require.Error(t, ValidateAetraStakingPolicyParameterValues(values))
}

func TestAetraStakingPolicyParameterRulesRejectInvalidTopNAndZeroValidatorTarget(t *testing.T) {
	values := DefaultAetraStakingPolicyParameterValues()
	values.Top10TargetBps = 4_500
	values.Top20TargetBps = 4_000
	values.Top33TargetBps = 10_001
	values.MaxValidatorsSoftTarget = 0

	report := BuildAetraStakingPolicyParameterReport(values)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyParamTop10TargetBps)
	require.Contains(t, report.Failed, AetraStakingPolicyParamTop20TargetBps)
	require.Contains(t, report.Failed, AetraStakingPolicyParamTop33TargetBps)
	require.Contains(t, report.Failed, AetraStakingPolicyParamMaxValidatorsSoftTarget)
}

func TestAetraStakingPolicyParameterRulesRejectNegativeMathValues(t *testing.T) {
	values := DefaultAetraStakingPolicyParameterValues()
	values.CommissionMaxDailyChangeBps = -1

	report := BuildAetraStakingPolicyParameterReport(values)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyParamCommissionMaxDailyChangeBps)
	require.Contains(t, report.Failed, AetraStakingPolicyParamRejectNegativeOrOverflowMath)
}

func TestDefaultAetraStakingPolicyEffectivePowerStage1IsLowConsensusRisk(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEffectivePowerStage1Evidence()

	report := BuildAetraStakingPolicyEffectivePowerReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, AetraStakingPolicyEffectivePowerStage1, report.Stage)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 4, report.Required)
	require.True(t, evidence.CapAffectsRewardCalculation)
	require.True(t, evidence.CapAffectsDelegationWarnings)
	require.False(t, evidence.CapAffectsCometBFTVotingPower)
	require.NoError(t, ValidateAetraStakingPolicyEffectivePowerEvidence(evidence))
}

func TestAetraStakingPolicyEffectivePowerStage1RejectsCometBFTPowerMutation(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEffectivePowerStage1Evidence()
	evidence.CapAffectsCometBFTVotingPower = true

	report := BuildAetraStakingPolicyEffectivePowerReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerCometBFTVotingPower+":stage_1_must_not_touch_cometbft_power")
	require.Error(t, ValidateAetraStakingPolicyEffectivePowerEvidence(evidence))
}

func TestDefaultAetraStakingPolicyEffectivePowerStage2RequiresConsensusIntegrationGates(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEffectivePowerStage2Evidence()

	report := BuildAetraStakingPolicyEffectivePowerReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, AetraStakingPolicyEffectivePowerStage2, report.Stage)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 9, report.Required)
	require.True(t, evidence.CapAffectsCometBFTVotingPower)
	require.NoError(t, ValidateAetraStakingPolicyEffectivePowerEvidence(evidence))
}

func TestAetraStakingPolicyEffectivePowerStage2RejectsMissingValidatorUpdateAndStakeSafety(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEffectivePowerStage2Evidence()
	evidence.ValidatorUpdatesUseCappedPower = false
	evidence.TotalVotingPowerConsistent = false
	evidence.NoValidatorCanExceedCap = false
	evidence.DelegationUnbondingSharesCorrect = false
	evidence.SlashingUsesUnderlyingRawStake = false
	evidence.EvidenceHandlingCorrect = false

	report := BuildAetraStakingPolicyEffectivePowerReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerValidatorUpdatesCapped)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerTotalVotingConsistent)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerNoValidatorExceedsCap)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerSharesCorrect)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerSlashingRawStake)
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerEvidenceHandlingCorrect)
	require.Error(t, ValidateAetraStakingPolicyEffectivePowerEvidence(evidence))
}

func TestAetraStakingPolicyEffectivePowerRejectsMissingScopeStageAndModule(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEffectivePowerStage1Evidence()
	evidence.ModuleName = ""
	evidence.Stage = "stage_3_unknown"
	evidence.DefinesCapScope = false

	report := BuildAetraStakingPolicyEffectivePowerReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, "effective_power_stage_unknown")
	require.Contains(t, report.Failed, AetraStakingPolicyEffectivePowerDefinesCapScope)

	evidence = DefaultAetraStakingPolicyEffectivePowerStage1Evidence()
	evidence.ModuleName = "x/other"
	report = BuildAetraStakingPolicyEffectivePowerReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
}

func TestDefaultAetraStakingPolicyMessageSpecCoversRequiredAndOptionalMessages(t *testing.T) {
	evidence := DefaultAetraStakingPolicyMessageSpecEvidence()

	report := BuildAetraStakingPolicyMessageSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 12, report.Required)
	require.Contains(t, evidence.GovernanceAuthorityMessages, AetraStakingPolicyMessageMsgUpdateStakingPolicyParams)
	require.Contains(t, evidence.GovernanceAuthorityMessages, AetraStakingPolicyMessageMsgUpdateValidatorPowerCapSchedule)
	require.Contains(t, evidence.GovernanceAuthorityMessages, AetraStakingPolicyMessageMsgSetCommissionPolicy)
	require.Contains(t, evidence.OptionalValidatorMessages, AetraStakingPolicyMessageMsgRegisterValidatorIdentity)
	require.Contains(t, evidence.OptionalValidatorMessages, AetraStakingPolicyMessageMsgUpdateValidatorIdentity)
	require.Contains(t, evidence.OptionalValidatorMessages, AetraStakingPolicyMessageMsgAcknowledgeOverCapWarning)
	require.NoError(t, ValidateAetraStakingPolicyMessageSpec(evidence))
}

func TestAetraStakingPolicyMessageSpecRejectsMissingMessagesAndChecks(t *testing.T) {
	evidence := DefaultAetraStakingPolicyMessageSpecEvidence()
	evidence.GovernanceAuthorityMessages = removeString(evidence.GovernanceAuthorityMessages, AetraStakingPolicyMessageMsgSetCommissionPolicy)
	evidence.OptionalValidatorMessages = removeString(evidence.OptionalValidatorMessages, AetraStakingPolicyMessageMsgUpdateValidatorIdentity)
	evidence.ValidateAuthority = false
	evidence.ValidateSigner = false
	evidence.RejectMalformedAddresses = false
	evidence.RejectInvalidParams = false
	evidence.EmitEvents = false
	evidence.CoveredByTests = false

	report := BuildAetraStakingPolicyMessageSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageGovernanceOrAuthorityOnly+"."+AetraStakingPolicyMessageMsgSetCommissionPolicy+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyMessageOptionalValidatorMessages+"."+AetraStakingPolicyMessageMsgUpdateValidatorIdentity+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyMessageValidateAuthority)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageValidateSigner)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageRejectMalformedAddresses)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageRejectInvalidParams)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageEmitEvents)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageCoveredByTests)
	require.Error(t, ValidateAetraStakingPolicyMessageSpec(evidence))
}

func TestAetraStakingPolicyMessageSpecRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraStakingPolicyMessageSpecEvidence()
	evidence.ModuleName = "x/other"
	evidence.GovernanceAuthorityMessages = append(evidence.GovernanceAuthorityMessages, AetraStakingPolicyMessageMsgSetCommissionPolicy)
	evidence.OptionalValidatorMessages = append(evidence.OptionalValidatorMessages, "MsgUnsafeValidatorOverride")

	report := BuildAetraStakingPolicyMessageSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	require.Contains(t, report.Failed, AetraStakingPolicyMessageGovernanceOrAuthorityOnly+"."+AetraStakingPolicyMessageMsgSetCommissionPolicy+":duplicate")
	require.Contains(t, report.Failed, AetraStakingPolicyMessageOptionalValidatorMessages+".MsgUnsafeValidatorOverride:unexpected")
}

func TestDefaultAetraStakingPolicyQuerySpecCoversRequiredQueries(t *testing.T) {
	evidence := DefaultAetraStakingPolicyQuerySpecEvidence()

	report := BuildAetraStakingPolicyQuerySpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 11, report.Required)
	for _, query := range []string{
		AetraStakingPolicyQueryParams,
		AetraStakingPolicyQueryValidatorPolicy,
		AetraStakingPolicyQueryValidatorEffectivePower,
		AetraStakingPolicyQueryValidatorOverflow,
		AetraStakingPolicyQueryTopNConcentration,
		AetraStakingPolicyQueryDelegationWarning,
		AetraStakingPolicyQueryCommissionPolicy,
		AetraStakingPolicyQueryConcentrationSnapshot,
		AetraStakingPolicyQueryNakamotoCoefficient,
	} {
		require.Contains(t, evidence.RequiredQueries, query)
	}
	require.True(t, evidence.StableResponses)
	require.True(t, evidence.IndexerFriendlyResponses)
	require.NoError(t, ValidateAetraStakingPolicyQuerySpec(evidence))
}

func TestAetraStakingPolicyQuerySpecRejectsMissingUnstableAndNonIndexerFriendlyQueries(t *testing.T) {
	evidence := DefaultAetraStakingPolicyQuerySpecEvidence()
	evidence.RequiredQueries = removeString(evidence.RequiredQueries, AetraStakingPolicyQueryValidatorOverflow, AetraStakingPolicyQueryNakamotoCoefficient)
	evidence.StableResponses = false
	evidence.IndexerFriendlyResponses = false

	report := BuildAetraStakingPolicyQuerySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "queries."+AetraStakingPolicyQueryValidatorOverflow+":missing")
	require.Contains(t, report.Failed, "queries."+AetraStakingPolicyQueryNakamotoCoefficient+":missing")
	require.Contains(t, report.Failed, AetraStakingPolicyQueryStableResponses)
	require.Contains(t, report.Failed, AetraStakingPolicyQueryIndexerFriendlyResponses)
	require.Error(t, ValidateAetraStakingPolicyQuerySpec(evidence))
}

func TestAetraStakingPolicyQuerySpecRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraStakingPolicyQuerySpecEvidence()
	evidence.ModuleName = ""
	evidence.RequiredQueries = append(evidence.RequiredQueries, AetraStakingPolicyQueryParams, "Query/UnstableDebug")

	report := BuildAetraStakingPolicyQuerySpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, "queries."+AetraStakingPolicyQueryParams+":duplicate")
	require.Contains(t, report.Failed, "queries.Query/UnstableDebug:unexpected")
}

func TestDefaultAetraStakingPolicyEventSpecCoversRequiredEvents(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEventSpecEvidence()

	report := BuildAetraStakingPolicyEventSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 8, report.Required)
	for _, event := range []string{
		AetraStakingPolicyEventParamsUpdated,
		AetraStakingPolicyEventValidatorOverCap,
		AetraStakingPolicyEventValidatorBackUnderCap,
		AetraStakingPolicyEventCommissionRejected,
		AetraStakingPolicyEventConcentrationSnapshot,
		AetraStakingPolicyEventRewardMultiplierChanged,
	} {
		require.Contains(t, evidence.RequiredEvents, event)
	}
	require.True(t, evidence.StableEventNames)
	require.True(t, evidence.IndexerFriendlyAttributes)
	require.NoError(t, ValidateAetraStakingPolicyEventSpec(evidence))
}

func TestAetraStakingPolicyEventSpecRejectsMissingDuplicateUnexpectedAndUnstableEvents(t *testing.T) {
	evidence := DefaultAetraStakingPolicyEventSpecEvidence()
	evidence.ModuleName = "x/other"
	evidence.RequiredEvents = removeString(evidence.RequiredEvents, AetraStakingPolicyEventCommissionRejected)
	evidence.RequiredEvents = append(evidence.RequiredEvents, AetraStakingPolicyEventParamsUpdated, "aetra.staking_policy.debug_only")
	evidence.StableEventNames = false
	evidence.IndexerFriendlyAttributes = false

	report := BuildAetraStakingPolicyEventSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	require.Contains(t, report.Failed, "events."+AetraStakingPolicyEventCommissionRejected+":missing")
	require.Contains(t, report.Failed, "events."+AetraStakingPolicyEventParamsUpdated+":duplicate")
	require.Contains(t, report.Failed, "events.aetra.staking_policy.debug_only:unexpected")
	require.Contains(t, report.Failed, AetraStakingPolicyEventStableNames)
	require.Contains(t, report.Failed, AetraStakingPolicyEventIndexerFriendlyAttrs)
	require.Error(t, ValidateAetraStakingPolicyEventSpec(evidence))
}

func TestDefaultAetraStakingPolicyInvariantSpecCoversRequiredInvariants(t *testing.T) {
	evidence := DefaultAetraStakingPolicyInvariantSpecEvidence()

	report := BuildAetraStakingPolicyInvariantSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 8, report.Required)
	require.NoError(t, ValidateAetraStakingPolicyInvariantSpec(evidence))
}

func TestAetraStakingPolicyInvariantSpecRejectsMissingSafetyInvariants(t *testing.T) {
	evidence := DefaultAetraStakingPolicyInvariantSpecEvidence()
	evidence.EffectivePowerNeverExceedsCap = false
	evidence.OverflowStakeNeverNegative = false
	evidence.RawStakeConservation = false
	evidence.CommissionWithinFloorAndMax = false
	evidence.CommissionChangeWithinDailyLimit = false
	evidence.TopNDoesNotExceedHundredPercent = false
	evidence.ExportImportPreservesPolicyState = false
	evidence.CoveredByTests = false

	report := BuildAetraStakingPolicyInvariantSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantEffectivePowerCap)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantOverflowNonNegative)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantRawStakeConservation)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantCommissionBounds)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantCommissionDailyChange)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantTopNMaxHundredPercent)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantExportImportPreserves)
	require.Contains(t, report.Failed, AetraStakingPolicyInvariantCoveredByTests)
	require.Error(t, ValidateAetraStakingPolicyInvariantSpec(evidence))
}

func TestAetraStakingPolicyInvariantSpecRejectsWrongModuleIdentity(t *testing.T) {
	evidence := DefaultAetraStakingPolicyInvariantSpecEvidence()
	evidence.ModuleName = ""

	report := BuildAetraStakingPolicyInvariantSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")

	evidence.ModuleName = "x/other"
	report = BuildAetraStakingPolicyInvariantSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
}

func TestDefaultAetraStakingPolicyTestSpecCoversRequiredTests(t *testing.T) {
	evidence := DefaultAetraStakingPolicyTestSpecEvidence()

	report := BuildAetraStakingPolicyTestSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 17, report.Required)
	for _, requiredTest := range []string{
		AetraStakingPolicyTestCapMath100Validators,
		AetraStakingPolicyTestCapMath150Validators,
		AetraStakingPolicyTestCapMath250Validators,
		AetraStakingPolicyTestCapMath300Validators,
		AetraStakingPolicyTestValidatorCrossingCapUpward,
		AetraStakingPolicyTestValidatorCrossingCapDownward,
		AetraStakingPolicyTestDelegationToOverCapValidator,
		AetraStakingPolicyTestRedelegationFromOverCapValidator,
		AetraStakingPolicyTestUnbondingFromOverCapValidator,
		AetraStakingPolicyTestSlashingOverCapValidator,
		AetraStakingPolicyTestCommissionBelowFloorRejected,
		AetraStakingPolicyTestCommissionAboveMaxRejected,
		AetraStakingPolicyTestCommissionDailyJumpRejected,
		AetraStakingPolicyTestGovernanceParamUpdateAccepted,
		AetraStakingPolicyTestGovernanceParamUpdateRejected,
		AetraStakingPolicyTestExportImportOverCapValidators,
		AetraStakingPolicyTestDeterministicConcentrationSnapshot,
	} {
		require.Contains(t, evidence.RequiredTests, requiredTest)
	}
	require.NoError(t, ValidateAetraStakingPolicyTestSpec(evidence))
}

func TestAetraStakingPolicyTestSpecRejectsMissingDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraStakingPolicyTestSpecEvidence()
	evidence.ModuleName = "x/other"
	evidence.RequiredTests = removeString(evidence.RequiredTests, AetraStakingPolicyTestSlashingOverCapValidator)
	evidence.RequiredTests = append(evidence.RequiredTests, AetraStakingPolicyTestCapMath100Validators, "untracked_manual_test")

	report := BuildAetraStakingPolicyTestSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraStakingPolicyModuleName)
	require.Contains(t, report.Failed, "tests."+AetraStakingPolicyTestSlashingOverCapValidator+":missing")
	require.Contains(t, report.Failed, "tests."+AetraStakingPolicyTestCapMath100Validators+":duplicate")
	require.Contains(t, report.Failed, "tests.untracked_manual_test:unexpected")
	require.Error(t, ValidateAetraStakingPolicyTestSpec(evidence))
}

func removeString(values []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !targetSet[value] {
			out = append(out, value)
		}
	}
	return out
}
