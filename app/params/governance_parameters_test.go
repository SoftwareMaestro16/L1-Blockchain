package params

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGovernanceParameterSpecsCoverSection13(t *testing.T) {
	specs := DefaultGovernanceParameterSpecs()
	report := BuildGovernanceParameterSafetyReport(specs)

	require.Empty(t, report.Failed)
	require.True(t, report.AllBounded)
	require.True(t, report.AllGenesisChecked)
	require.True(t, report.AllEmitEvents)
	require.True(t, report.CriticalProtected)
	require.NoError(t, ValidateGovernanceParameterSpecs(specs))

	keys := map[string]bool{}
	for _, spec := range specs {
		keys[spec.Key] = true
	}
	for key := range requiredGovernanceParameterKeys() {
		require.True(t, keys[key], "missing governed param %s", key)
	}
}

func TestGovernanceControlledModulesCoverSection271(t *testing.T) {
	specs := DefaultGovernanceParameterSpecs()
	byCategory := map[string]int{}
	byKey := map[string]bool{}
	for _, spec := range specs {
		byCategory[spec.Category]++
		byKey[spec.Key] = true
	}

	require.NotZero(t, byCategory["staking_policy"])
	require.NotZero(t, byCategory["economics"])
	require.NotZero(t, byCategory["validator_score"])
	require.NotZero(t, byCategory["slashing"])
	require.NotZero(t, byCategory["vm"])
	require.NotZero(t, byCategory["treasury"])
	require.NotZero(t, byCategory["validator_set_growth"])
	require.NotZero(t, byCategory["consensus"])
	require.True(t, byKey[GovernanceParamValidatorScorePolicy])
	require.True(t, byKey[GovernanceParamValidatorSetGrowth])
	require.True(t, byKey[GovernanceParamBlockGasLimit])
	require.True(t, byKey[GovernanceParamBlockMaxBytes])
}

func TestGovernanceParamSpecsCarrySection272Metadata(t *testing.T) {
	specs := DefaultGovernanceParameterSpecs()
	for _, spec := range specs {
		require.NotEmpty(t, spec.ValueType, spec.Key)
		require.NotEmpty(t, spec.Authority, spec.Key)
		require.NotEmpty(t, spec.EventType, spec.Key)
		require.True(t, spec.InvalidUpdateTest, spec.Key)
		if spec.Critical {
			require.True(t, spec.ApplyEpochDelay, spec.Key)
		}
		switch spec.ValueType {
		case GovernanceValueTypeInteger:
			require.LessOrEqual(t, spec.MinInt, spec.DefaultInt, spec.Key)
			require.LessOrEqual(t, spec.DefaultInt, spec.MaxInt, spec.Key)
		case GovernanceValueTypeEnum:
			require.Contains(t, spec.AllowedValues, spec.DefaultString, spec.Key)
		default:
			require.Fail(t, "unexpected governance value type", spec.Key)
		}
	}
	require.NoError(t, ValidateGovernanceParameterSpecs(specs))
}

func TestGovernanceParamChangeRejectsUnsafeExecution(t *testing.T) {
	change := GovernanceParamChange{
		Value: GovernanceParamValue{
			Key:		GovernanceParamValidatorSetSize,
			IntValue:	AetraValidatorSetMax + 1,
		},
		VotingPeriodBlocks:	GovernanceCriticalVotingPeriodBlocks,
		QuorumBps:		GovernanceCriticalQuorumBps,
		ProposalExecution:	true,
		EmitsEvent:		true,
	}

	require.ErrorContains(t, ValidateGovernanceParamChange(nil, change), "validator_set_size")

	change.Value.IntValue = AetraValidatorSetGenesisMin
	change.ProposalExecution = false
	require.ErrorContains(t, ValidateGovernanceParamChange(nil, change), "proposal execution")

	change.ProposalExecution = true
	change.EmitsEvent = false
	require.ErrorContains(t, ValidateGovernanceParamChange(nil, change), "emit events")
}

func TestCriticalGovernanceParamChangesRequireLongerVotingAndHigherQuorum(t *testing.T) {
	change := GovernanceParamChange{
		Value: GovernanceParamValue{
			Key:		GovernanceParamInflationMax,
			IntValue:	MaxInflationBps,
		},
		VotingPeriodBlocks:	GovernanceDefaultVotingPeriodBlocks,
		QuorumBps:		GovernanceCriticalQuorumBps,
		ProposalExecution:	true,
		EmitsEvent:		true,
	}

	require.ErrorContains(t, ValidateGovernanceParamChange(nil, change), "longer voting period")

	change.VotingPeriodBlocks = GovernanceCriticalVotingPeriodBlocks
	change.QuorumBps = GovernanceDefaultQuorumBps
	require.ErrorContains(t, ValidateGovernanceParamChange(nil, change), "higher quorum")

	change.QuorumBps = GovernanceCriticalQuorumBps
	require.NoError(t, ValidateGovernanceParamChange(nil, change))
}

func TestNonCriticalGovernanceParamChangeUsesNormalVotingBounds(t *testing.T) {
	change := GovernanceParamChange{
		Value: GovernanceParamValue{
			Key:		GovernanceParamCommissionFloor,
			IntValue:	StakingCommissionFloorBps,
		},
		VotingPeriodBlocks:	GovernanceDefaultVotingPeriodBlocks,
		QuorumBps:		GovernanceDefaultQuorumBps,
		ProposalExecution:	true,
		EmitsEvent:		true,
	}

	require.NoError(t, ValidateGovernanceParamChange(nil, change))
}

func TestGovernanceGenesisValidationRejectsInvalidParams(t *testing.T) {
	values := DefaultGovernanceGenesisParams()
	require.NoError(t, ValidateGovernanceGenesisParams(nil, values))

	values[0].IntValue = 99
	require.ErrorContains(t, ValidateGovernanceGenesisParams(nil, values), GovernanceParamValidatorSetSize)

	values = DefaultGovernanceGenesisParams()
	values = values[:len(values)-1]
	require.ErrorContains(t, ValidateGovernanceGenesisParams(nil, values), "genesis governance parameter")
}

func TestGovernanceEnumParamsAreBounded(t *testing.T) {
	change := GovernanceParamChange{
		Value: GovernanceParamValue{
			Key:		GovernanceParamAVMContractUploadPolicy,
			StringValue:	AVMContractUploadGovernanceOnly,
		},
		VotingPeriodBlocks:	GovernanceCriticalVotingPeriodBlocks,
		QuorumBps:		GovernanceCriticalQuorumBps,
		ProposalExecution:	true,
		EmitsEvent:		true,
	}

	require.NoError(t, ValidateGovernanceParamChange(nil, change))

	change.Value.StringValue = "open_upload_for_everyone"
	require.ErrorContains(t, ValidateGovernanceParamChange(nil, change), "allowed policy value")
}

func TestGovernanceSafetyReportDetectsMissingBoundsGenesisAndEvents(t *testing.T) {
	specs := DefaultGovernanceParameterSpecs()
	specs[0].ExecutionBounded = false
	specs[1].GenesisRequired = false
	specs[2].EmitsEvents = false

	report := BuildGovernanceParameterSafetyReport(specs)
	require.NotEmpty(t, report.Failed)
	require.False(t, report.AllBounded)
	require.False(t, report.AllGenesisChecked)
	require.False(t, report.AllEmitEvents)
	require.Error(t, ValidateGovernanceParameterSpecs(specs))
}

func TestGovernanceSafetyReportDetectsMissingSection272Metadata(t *testing.T) {
	specs := DefaultGovernanceParameterSpecs()
	specs[0].Authority = ""
	specs[1].ApplyEpochDelay = false
	specs[2].EventType = ""
	specs[3].InvalidUpdateTest = false
	specs[4].DefaultInt = specs[4].MaxInt + 1

	report := BuildGovernanceParameterSafetyReport(specs)
	require.NotEmpty(t, report.Failed)
	require.Contains(t, report.Failed, specs[0].Key+":authority_missing")
	require.Contains(t, report.Failed, specs[1].Key+":critical_param_must_apply_at_epoch_boundary")
	require.Contains(t, report.Failed, specs[2].Key+":event_type_missing")
	require.Contains(t, report.Failed, specs[3].Key+":invalid_update_test_missing")
	require.Contains(t, report.Failed, specs[4].Key+":default integer value is outside bounds")
	require.Error(t, ValidateGovernanceParameterSpecs(specs))
}

func TestDefaultGovernanceTestingEvidenceCoversSection273(t *testing.T) {
	evidence := DefaultGovernanceTestingEvidence()

	report := BuildGovernanceTestingReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, "x/gov", report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 8, report.Required)
	require.NoError(t, ValidateGovernanceTestingEvidence(evidence))
}

func TestGovernanceTestingEvidenceRejectsMissingRequiredTests(t *testing.T) {
	evidence := DefaultGovernanceTestingEvidence()
	evidence.ModuleName = "x/params"
	evidence.ValidParamProposalExecutes = false
	evidence.InvalidParamRejected = false
	evidence.UnauthorizedAuthority = false
	evidence.EmergencyUnsafeRejected = false
	evidence.EpochDelayedActivation = false
	evidence.EventEmitted = false
	evidence.QueryReflectsNewParams = false
	evidence.ExportImportAfterChange = false

	report := BuildGovernanceTestingReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_x/gov")
	require.Contains(t, report.Failed, GovernanceTestValidParamProposalExecutes)
	require.Contains(t, report.Failed, GovernanceTestInvalidParamRejected)
	require.Contains(t, report.Failed, GovernanceTestUnauthorizedAuthority)
	require.Contains(t, report.Failed, GovernanceTestEmergencyUnsafeRejected)
	require.Contains(t, report.Failed, GovernanceTestEpochDelayedActivation)
	require.Contains(t, report.Failed, GovernanceTestEventEmitted)
	require.Contains(t, report.Failed, GovernanceTestQueryReflectsNewParams)
	require.Contains(t, report.Failed, GovernanceTestExportImportAfterChange)
	require.Error(t, ValidateGovernanceTestingEvidence(evidence))
}

func TestGovernanceParamsSchema185DefaultGenesisExactMatch(t *testing.T) {
	values := governanceValuesByKey(DefaultGovernanceGenesisParams())

	require.Equal(t, int64(AetraValidatorSetGenesisMin), values[GovernanceParamValidatorSetSize].IntValue)
	require.Equal(t, GovernanceDefaultValidatorEntryStake, values[GovernanceParamValidatorEntryStake].IntValue)
	require.Equal(t, GovernanceDefaultPoolBackedSelfStake, values[GovernanceParamPoolBackedSelfStake].IntValue)
	require.Equal(t, GovernanceDefaultPoolBackedPoolStake, values[GovernanceParamPoolBackedPoolStake].IntValue)
	require.Equal(t, GovernanceDefaultPoolMinDeposit, values[GovernanceParamPoolMinDeposit].IntValue)
	require.Equal(t, DirectUserDelegationDisabled, values[GovernanceParamDirectUserDelegation].StringValue)
	require.Equal(t, GovernanceDefaultUnbondingBlocks, values[GovernanceParamUnbondingBlocks].IntValue)
	require.Equal(t, GovernanceDefaultMinTxFeeNaet, values[GovernanceParamMinTxFee].IntValue)
	require.Equal(t, int64(5_000), values[GovernanceParamFeeBurnShare].IntValue)
	require.Equal(t, int64(3_500), values[GovernanceParamFeeRewardShare].IntValue)
	require.Equal(t, int64(1_500), values[GovernanceParamFeeTreasuryShare].IntValue)
	require.Equal(t, GovernanceDefaultStorageRentRate, values[GovernanceParamStorageRentRate].IntValue)
	require.Equal(t, GovernanceDefaultReserveMinRunway, values[GovernanceParamSystemReserveMin].IntValue)
	require.Equal(t, GovernanceDefaultReserveWarningRunway, values[GovernanceParamSystemReserveWarning].IntValue)
	require.Equal(t, GovernanceDefaultReserveCriticalRunway, values[GovernanceParamSystemReserveCritical].IntValue)
	require.NoError(t, ValidateGovernanceGenesisParams(nil, DefaultGovernanceGenesisParams()))
}

func TestGovernanceParamsSchema185RejectsInvalidValidatorCount(t *testing.T) {
	values := DefaultGovernanceGenesisParams()
	setGovernanceIntValue(values, GovernanceParamValidatorSetSize, 99)
	require.ErrorContains(t, ValidateGovernanceGenesisParams(nil, values), GovernanceParamValidatorSetSize)

	values = DefaultGovernanceGenesisParams()
	setGovernanceIntValue(values, GovernanceParamValidatorSetSize, 301)
	require.ErrorContains(t, ValidateGovernanceGenesisParams(nil, values), GovernanceParamValidatorSetSize)
}

func TestGovernanceParamsSchema185RejectsInvalidPoolBackedSplit(t *testing.T) {
	values := DefaultGovernanceGenesisParams()
	setGovernanceIntValue(values, GovernanceParamPoolBackedSelfStake, GovernanceDefaultPoolBackedSelfStake-BaseUnitsPerDisplay)
	require.ErrorContains(t, ValidateGovernanceGenesisParams(nil, values), GovernanceParamPoolBackedSelfStake)

	specs := DefaultGovernanceParameterSpecs()
	relaxGovernanceIntBounds(specs, GovernanceParamPoolBackedSelfStake, 1, GovernanceDefaultValidatorEntryStake)
	relaxGovernanceIntBounds(specs, GovernanceParamPoolBackedPoolStake, 1, GovernanceDefaultValidatorEntryStake)
	values = DefaultGovernanceGenesisParams()
	setGovernanceIntValue(values, GovernanceParamPoolBackedSelfStake, int64(500_000)*BaseUnitsPerDisplay)
	setGovernanceIntValue(values, GovernanceParamPoolBackedPoolStake, int64(500_000)*BaseUnitsPerDisplay)
	require.ErrorContains(t, ValidateGovernanceGenesisParams(specs, values), "pool-backed validator split")
}

func TestGovernanceParamsSchema185RejectsInvalidFeeSplit(t *testing.T) {
	values := DefaultGovernanceGenesisParams()
	setGovernanceIntValue(values, GovernanceParamFeeRewardShare, 3_000)
	require.ErrorContains(t, ValidateGovernanceGenesisParams(nil, values), "fee split")
}

func TestGovernanceParamsSchema185ExportImportStable(t *testing.T) {
	values := DefaultGovernanceGenesisParams()
	bz, err := json.Marshal(values)
	require.NoError(t, err)

	var imported []GovernanceParamValue
	require.NoError(t, json.Unmarshal(bz, &imported))
	require.Equal(t, values, imported)
	require.NoError(t, ValidateGovernanceGenesisParams(nil, imported))
}

func governanceValuesByKey(values []GovernanceParamValue) map[string]GovernanceParamValue {
	byKey := make(map[string]GovernanceParamValue, len(values))
	for _, value := range values {
		byKey[value.Key] = value
	}
	return byKey
}

func setGovernanceIntValue(values []GovernanceParamValue, key string, next int64) {
	for i := range values {
		if values[i].Key == key {
			values[i].IntValue = next
			return
		}
	}
}

func relaxGovernanceIntBounds(specs []GovernanceParameterSpec, key string, minValue int64, maxValue int64) {
	for i := range specs {
		if specs[i].Key == key {
			specs[i].MinInt = minValue
			specs[i].MaxInt = maxValue
			return
		}
	}
}
