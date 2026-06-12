package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestSelectCanonicalProposerUsesHighestPriorityScore(t *testing.T) {
	group := proposerTestGroup()
	selection, err := SelectCanonicalProposer(ProposerSelectionInput{
		Group:	group,
		PriorityInputs: map[string]ProposerPriorityInput{
			"val-a":	{ValidatorScore: sdkmath.NewInt(1_000), PriorProposerPerformanceBps: 9_000, TaskReliabilityBps: 9_000, StakeSaturationDampeningBps: 10_000},
			"val-b":	{ValidatorScore: sdkmath.NewInt(1_500), PriorProposerPerformanceBps: 10_000, TaskReliabilityBps: 10_000, StakeSaturationDampeningBps: 10_000},
			"val-c":	{ValidatorScore: sdkmath.NewInt(2_000), PriorProposerPerformanceBps: 10_000, TaskReliabilityBps: 10_000, StakeSaturationDampeningBps: 5_000},
		},
	}, 7)
	require.NoError(t, err)
	require.Equal(t, "val-b", selection.CanonicalProposer)
	require.Equal(t, []string{"val-a", "val-c"}, selection.VerifierValidators)
	require.False(t, selection.FallbackUsed)
	require.Len(t, selection.Priorities, 3)
	require.Equal(t, sdkmath.NewInt(1_500), selection.CanonicalPriority.PriorityScore)
	require.Equal(t, ProposerStatusReady, selection.CanonicalPriority.ProposerStatus)
}

func TestSelectCanonicalProposerFallsBackWhenTopUnavailable(t *testing.T) {
	group := proposerTestGroup()
	selection, err := SelectCanonicalProposer(ProposerSelectionInput{
		Group:	group,
		PriorityInputs: map[string]ProposerPriorityInput{
			"val-a":	{ValidatorScore: sdkmath.NewInt(1_000), PriorProposerPerformanceBps: 10_000, TaskReliabilityBps: 10_000, StakeSaturationDampeningBps: 10_000},
			"val-b":	{ValidatorScore: sdkmath.NewInt(2_000), PriorProposerPerformanceBps: 10_000, TaskReliabilityBps: 10_000, StakeSaturationDampeningBps: 10_000},
			"val-c":	{ValidatorScore: sdkmath.NewInt(1_500), PriorProposerPerformanceBps: 10_000, TaskReliabilityBps: 10_000, StakeSaturationDampeningBps: 10_000},
		},
		Unavailable:	map[string]bool{"val-b": true},
	}, 8)
	require.NoError(t, err)
	require.Equal(t, "val-c", selection.CanonicalProposer)
	require.True(t, selection.FallbackUsed)
	require.Equal(t, ProposerStatusFallback, selection.CanonicalPriority.ProposerStatus)
	require.Equal(t, []string{"val-a", "val-b"}, selection.VerifierValidators)
}

func TestProposerPriorityScoreAppliesMissPenaltyAndDampening(t *testing.T) {
	score, err := ComputeProposerPriorityScore(ProposerPriorityInput{
		ValidatorScore:			sdkmath.NewInt(10_000),
		PriorProposerPerformanceBps:	8_000,
		MissedProposalCount:		2,
		TaskReliabilityBps:		9_000,
		StakeSaturationDampeningBps:	5_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(2_880), score)

	score, err = ComputeProposerPriorityScore(ProposerPriorityInput{
		ValidatorScore:			sdkmath.NewInt(10_000),
		PriorProposerPerformanceBps:	10_000,
		MissedProposalCount:		20,
		TaskReliabilityBps:		10_000,
		StakeSaturationDampeningBps:	10_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.ZeroInt(), score)
}

func TestSelectCanonicalProposerRejectsUnavailableSet(t *testing.T) {
	group := proposerTestGroup()
	_, err := SelectCanonicalProposer(ProposerSelectionInput{
		Group:	group,
		PriorityInputs: map[string]ProposerPriorityInput{
			"val-a":	{ValidatorScore: sdkmath.NewInt(1_000)},
			"val-b":	{ValidatorScore: sdkmath.NewInt(1_000)},
			"val-c":	{ValidatorScore: sdkmath.NewInt(1_000)},
		},
		Unavailable:	map[string]bool{"val-a": true, "val-b": true, "val-c": true},
	}, 9)
	require.ErrorContains(t, err, "no available proposer")
}

func TestSlotAssignmentFallbackActivatesOnlyAfterMissedTimeout(t *testing.T) {
	group := proposerTestGroup()
	input := ProposerSelectionInput{
		Group:	group,
		PriorityInputs: map[string]ProposerPriorityInput{
			"val-a":	{ValidatorScore: sdkmath.NewInt(3_000)},
			"val-b":	{ValidatorScore: sdkmath.NewInt(2_000)},
			"val-c":	{ValidatorScore: sdkmath.NewInt(1_000)},
		},
		Unavailable:	map[string]bool{"val-a": true},
	}

	_, err := BuildSlotAssignment(SlotAssignmentInput{
		SelectionInput:			input,
		Slot:				10,
		CurrentHeight:			99,
		MissedProposalTimeoutHeight:	100,
	})
	require.ErrorContains(t, err, "before missed proposal timeout")

	record, err := BuildSlotAssignment(SlotAssignmentInput{
		SelectionInput:			input,
		Slot:				10,
		CurrentHeight:			100,
		MissedProposalTimeoutHeight:	100,
	})
	require.NoError(t, err)
	require.Equal(t, "val-b", record.CanonicalProposer)
	require.True(t, record.FallbackActivated)
	require.Equal(t, []string{"val-a", "val-b", "val-c"}, record.FallbackOrder)
	require.Equal(t, []string{"val-a", "val-c"}, record.VerifierValidators)
	require.NotEmpty(t, record.EligibilityProof.ProofHash)
}

func TestFallbackOrderQueryAndEligibilityProofAreDeterministic(t *testing.T) {
	group := proposerTestGroup()
	input := ProposerSelectionInput{
		Group:	group,
		PriorityInputs: map[string]ProposerPriorityInput{
			"val-a":	{ValidatorScore: sdkmath.NewInt(1_000)},
			"val-b":	{ValidatorScore: sdkmath.NewInt(3_000)},
			"val-c":	{ValidatorScore: sdkmath.NewInt(2_000)},
		},
	}
	order, err := QueryFallbackOrder(input, 11)
	require.NoError(t, err)
	require.Equal(t, []string{"val-b", "val-c", "val-a"}, order)

	record, err := BuildSlotAssignment(SlotAssignmentInput{
		SelectionInput:			input,
		Slot:				11,
		CurrentHeight:			111,
		MissedProposalTimeoutHeight:	110,
	})
	require.NoError(t, err)
	require.Equal(t, "val-b", record.CanonicalProposer)
	require.NoError(t, VerifyProposerEligibilityProof(record.EligibilityProof, record.EligibilityProofPriority(), group))

	tampered := record.EligibilityProof
	tampered.ValidatorAddress = "val-c"
	require.ErrorContains(t, VerifyProposerEligibilityProof(tampered, record.EligibilityProofPriority(), group), "mismatch")
}

func TestSlotAssignmentRejectsProposerVerifierConflict(t *testing.T) {
	group := proposerTestGroup()
	record := SlotAssignmentRecord{
		EpochID:			group.EpochID,
		Slot:				12,
		TaskGroupID:			group.TaskGroupID,
		CanonicalProposer:		"val-a",
		VerifierValidators:		[]string{"val-a", "val-b"},
		FallbackOrder:			[]string{"val-a", "val-b", "val-c"},
		MissedProposalTimeoutHeight:	120,
		EligibilityProof:		BuildProposerEligibilityProof(ProposerPriority{EpochID: group.EpochID, Slot: 12, TaskGroupID: group.TaskGroupID, ValidatorAddress: "val-a", PriorityScore: sdkmath.NewInt(1_000), ProposerStatus: ProposerStatusReady}, group),
	}
	require.ErrorContains(t, record.Validate(group), "cannot also be verifier")
}

func TestMissedProposalTrackingReducesFuturePriority(t *testing.T) {
	group := proposerTestGroup()
	baseInputs := map[string]ProposerPriorityInput{
		"val-a":	{ValidatorScore: sdkmath.NewInt(2_000)},
		"val-b":	{ValidatorScore: sdkmath.NewInt(1_900)},
		"val-c":	{ValidatorScore: sdkmath.NewInt(1_800)},
	}
	records, err := RecordMissedProposal(nil, group.EpochID, 13, group.TaskGroupID, "val-a")
	require.NoError(t, err)
	records, err = RecordMissedProposal(records, group.EpochID, 14, group.TaskGroupID, "val-a")
	require.NoError(t, err)

	tracked := ApplyMissedProposalTracking(baseInputs, records, group.EpochID, group.TaskGroupID)
	selection, err := SelectCanonicalProposer(ProposerSelectionInput{Group: group, PriorityInputs: tracked}, 15)
	require.NoError(t, err)
	require.Equal(t, "val-b", selection.CanonicalProposer)
	require.Equal(t, uint64(2), tracked["val-a"].MissedProposalCount)
}

func (r SlotAssignmentRecord) EligibilityProofPriority() ProposerPriority {
	return ProposerPriority{
		EpochID:		r.EligibilityProof.EpochID,
		Slot:			r.EligibilityProof.Slot,
		TaskGroupID:		r.EligibilityProof.TaskGroupID,
		ValidatorAddress:	r.EligibilityProof.ValidatorAddress,
		PriorityScore:		r.EligibilityProof.PriorityScore,
		FallbackOrder:		r.EligibilityProof.FallbackOrder,
		ProposerStatus:		ProposerStatusReady,
	}
}

func proposerTestGroup() postypes.TaskGroup {
	group := postypes.TaskGroup{
		EpochID:		3,
		WorkloadID:		"proof-market",
		WorkloadType:		postypes.WorkloadTypeProofVerification,
		ValidatorMembers:	[]string{"val-a", "val-b", "val-c"},
		ProposerOrder:		[]string{"val-a", "val-b", "val-c"},
		VerifierSet:		[]string{"val-a", "val-b", "val-c"},
		MinimumGroupSize:	3,
		StakeWeightRoot:	"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		AssignmentSeed:		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ActivationHeight:	30,
		ExpiryHeight:		60,
	}
	group.TaskGroupID = postypes.ComputeTaskGroupID(group)
	return group
}
