package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestDeterministicAssignmentReproducibleFromCommittedState(t *testing.T) {
	params := taskgroupParams()
	validators := taskgroupValidators(t, params, 4, postypes.ValidatorCapacity{
		MaxTaskGroups:		4,
		SupportedWorkloads:	[]postypes.WorkloadType{postypes.WorkloadTypeProofVerification, postypes.WorkloadTypeDataAvailability},
		ZoneSupport:		[]string{"GLOBAL"},
	})
	epoch, err := postypes.NewEpochRecord(params, 12, 120, 180, postypes.EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	registry, err := NewWorkloadRegistry(params, []WorkloadRegistration{
		{
			WorkloadID:		"proof-market",
			WorkloadType:		postypes.WorkloadTypeProofVerification,
			ZoneID:			"GLOBAL",
			ShardID:		"proof",
			MinimumGroupSize:	2,
			RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
		},
		{
			WorkloadID:		"da-sampling",
			WorkloadType:		postypes.WorkloadTypeDataAvailability,
			ZoneID:			"GLOBAL",
			ShardID:		"da",
			MinimumGroupSize:	2,
			RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
		},
	})
	require.NoError(t, err)

	left, err := DeterministicAssignment(AssignmentInput{
		Params:			params,
		Epoch:			epoch,
		ValidatorSet:		validators,
		Registry:		registry,
		ActivationHeight:	121,
		ExpiryHeight:		181,
	})
	require.NoError(t, err)
	right, err := DeterministicAssignment(AssignmentInput{
		Params:			params,
		Epoch:			epoch,
		ValidatorSet:		validators,
		Registry:		registry,
		ActivationHeight:	121,
		ExpiryHeight:		181,
	})
	require.NoError(t, err)
	require.Equal(t, left.GroupSet.Root, right.GroupSet.Root)
	require.Equal(t, left.GroupSet.Groups, right.GroupSet.Groups)
	require.NoError(t, left.Validate())
}

func TestAssignmentProofQueryReturnsCommittedGroupProof(t *testing.T) {
	params := taskgroupParams()
	validators := taskgroupValidators(t, params, 3, postypes.ValidatorCapacity{MaxTaskGroups: 3})
	epoch, err := postypes.NewEpochRecord(params, 13, 130, 190, postypes.EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	registry, err := NewWorkloadRegistry(params, []WorkloadRegistration{{
		WorkloadID:		"service-validation",
		WorkloadType:		postypes.WorkloadTypeServiceValidation,
		ZoneID:			"GLOBAL",
		ShardID:		"service",
		MinimumGroupSize:	2,
		RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
	}})
	require.NoError(t, err)
	state, err := DeterministicAssignment(AssignmentInput{
		Params:			params,
		Epoch:			epoch,
		ValidatorSet:		validators,
		Registry:		registry,
		ActivationHeight:	131,
		ExpiryHeight:		191,
	})
	require.NoError(t, err)

	proof, found, err := state.QueryAssignmentProof("service-validation")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(13), proof.EpochID)
	require.Equal(t, "service-validation", proof.WorkloadID)
	require.Equal(t, state.GroupSet.Root, proof.TaskGroupRoot)
	require.Len(t, proof.StakeWeightRoot, postypes.PosHashHexLength)
	require.Equal(t, epoch.Seed, proof.AssignmentSeed)
	require.Len(t, proof.ValidatorMembers, 2)
	require.Len(t, proof.ProposerOrder, 2)
	require.NotEmpty(t, proof.VerifierSet)

	_, found, err = state.QueryAssignmentProof("missing")
	require.NoError(t, err)
	require.False(t, found)
}

func TestWorkloadRegistryRejectsBelowMinimumGroupSize(t *testing.T) {
	params := taskgroupParams()
	params.MinTaskGroupValidators = 2
	_, err := NewWorkloadRegistry(params, []WorkloadRegistration{{
		WorkloadID:		"too-small",
		WorkloadType:		postypes.WorkloadTypeProofVerification,
		ZoneID:			"GLOBAL",
		ShardID:		"proof",
		MinimumGroupSize:	1,
		RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
	}})
	require.ErrorContains(t, err, "at least 2")
}

func TestDeterministicAssignmentPreventsValidatorOverload(t *testing.T) {
	params := taskgroupParams()
	validators := taskgroupValidators(t, params, 2, postypes.ValidatorCapacity{
		MaxTaskGroups:		1,
		SupportedWorkloads:	[]postypes.WorkloadType{postypes.WorkloadTypeProofVerification},
		ZoneSupport:		[]string{"GLOBAL"},
	})
	epoch, err := postypes.NewEpochRecord(params, 14, 140, 200, postypes.EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	registry, err := NewWorkloadRegistry(params, []WorkloadRegistration{
		{
			WorkloadID:		"proof-a",
			WorkloadType:		postypes.WorkloadTypeProofVerification,
			ZoneID:			"GLOBAL",
			ShardID:		"proof-a",
			MinimumGroupSize:	1,
			RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
		},
		{
			WorkloadID:		"proof-b",
			WorkloadType:		postypes.WorkloadTypeProofVerification,
			ZoneID:			"GLOBAL",
			ShardID:		"proof-b",
			MinimumGroupSize:	1,
			RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
		},
		{
			WorkloadID:		"proof-c",
			WorkloadType:		postypes.WorkloadTypeProofVerification,
			ZoneID:			"GLOBAL",
			ShardID:		"proof-c",
			MinimumGroupSize:	1,
			RoleRequirements:	[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier},
		},
	})
	require.NoError(t, err)

	_, err = DeterministicAssignment(AssignmentInput{
		Params:			params,
		Epoch:			epoch,
		ValidatorSet:		validators,
		Registry:		registry,
		ActivationHeight:	141,
		ExpiryHeight:		201,
	})
	require.ErrorContains(t, err, "insufficient validators")
}

func taskgroupParams() postypes.Params {
	params := postypes.DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 1
	params.MaxTaskGroupValidators = 4
	return params
}

func taskgroupValidators(t *testing.T, params postypes.Params, count int, capacity postypes.ValidatorCapacity) []postypes.ScoredValidator {
	t.Helper()
	validators := make([]postypes.ScoredValidator, count)
	for i := range validators {
		candidate := postypes.Candidate{
			ValidatorID:		fmt.Sprintf("val-%03d", i),
			SelfStakeNaet:		sdkmath.NewInt(1_000),
			DelegatedStakeNaet:	sdkmath.ZeroInt(),
			PerformanceScoreBps:	postypes.BasisPoints,
			UptimeFactorBps:	postypes.BasisPoints,
			CommissionBps:		500,
			Roles:			[]postypes.ValidatorRole{postypes.ValidatorRoleVerifier, postypes.ValidatorRoleBlockProducer},
			Capacity:		capacity,
		}
		scored, err := postypes.ScoreCandidate(params, candidate)
		require.NoError(t, err)
		validators[i] = scored
	}
	return validators
}
