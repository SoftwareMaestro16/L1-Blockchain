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
