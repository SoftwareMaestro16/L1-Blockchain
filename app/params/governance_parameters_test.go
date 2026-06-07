package params

import (
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
			Key:      GovernanceParamValidatorSetSize,
			IntValue: AetraValidatorSetMax + 1,
		},
		VotingPeriodBlocks: GovernanceCriticalVotingPeriodBlocks,
		QuorumBps:          GovernanceCriticalQuorumBps,
		ProposalExecution:  true,
		EmitsEvent:         true,
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
			Key:      GovernanceParamInflationMax,
			IntValue: MaxInflationBps,
		},
		VotingPeriodBlocks: GovernanceDefaultVotingPeriodBlocks,
		QuorumBps:          GovernanceCriticalQuorumBps,
		ProposalExecution:  true,
		EmitsEvent:         true,
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
			Key:      GovernanceParamCommissionFloor,
			IntValue: StakingCommissionFloorBps,
		},
		VotingPeriodBlocks: GovernanceDefaultVotingPeriodBlocks,
		QuorumBps:          GovernanceDefaultQuorumBps,
		ProposalExecution:  true,
		EmitsEvent:         true,
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
			Key:         GovernanceParamCosmWasmUploadPolicy,
			StringValue: CosmWasmUploadGovernanceOnly,
		},
		VotingPeriodBlocks: GovernanceCriticalVotingPeriodBlocks,
		QuorumBps:          GovernanceCriticalQuorumBps,
		ProposalExecution:  true,
		EmitsEvent:         true,
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
