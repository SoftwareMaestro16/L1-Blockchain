package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
	taskgrouptypes "github.com/sovereign-l1/l1/x/taskgroups/types"
)

func TestKeeperSubmitsReceiptsAndAggregatesVerification(t *testing.T) {
	group := keeperTestGroup()
	k, err := NewKeeper([]postypes.TaskGroup{group})
	require.NoError(t, err)
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	receiptA, err := taskgrouptypes.NewVerificationReceipt(group, "val-a", objectHash, taskgrouptypes.VerificationResultValid, "sig-a", sdkmath.ZeroInt(), 40)
	require.NoError(t, err)
	receiptB, err := taskgrouptypes.NewVerificationReceipt(group, "val-b", objectHash, taskgrouptypes.VerificationResultInvalid, "sig-b", sdkmath.ZeroInt(), 40)
	require.NoError(t, err)

	require.NoError(t, k.SubmitVerificationReceipt(receiptB))
	require.NoError(t, k.SubmitVerificationReceipt(receiptA))
	set, found := k.VerificationReceiptSet(group.TaskGroupID)
	require.True(t, found)
	require.Len(t, set.Receipts, 2)
	require.Equal(t, receiptA, set.Receipts[0])

	aggregation, found, err := k.AggregateVerificationReceipts(group.TaskGroupID, objectHash, 5_000)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint32(1), aggregation.ValidCount)
	require.Equal(t, uint32(1), aggregation.InvalidCount)
	require.Equal(t, uint32(6_666), aggregation.ParticipationBps)
	require.True(t, aggregation.QuorumReached)
	require.NotNil(t, aggregation.InvalidEvidence)
}

func TestKeeperRejectsUnknownTaskGroupAndDuplicateReceipts(t *testing.T) {
	group := keeperTestGroup()
	k, err := NewKeeper([]postypes.TaskGroup{group})
	require.NoError(t, err)
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	receipt, err := taskgrouptypes.NewVerificationReceipt(group, "val-a", objectHash, taskgrouptypes.VerificationResultValid, "sig-a", sdkmath.ZeroInt(), 40)
	require.NoError(t, err)

	unknown := receipt
	unknown.TaskGroupID = "missing"
	require.ErrorContains(t, k.SubmitVerificationReceipt(unknown), "task group not found")

	require.NoError(t, k.SubmitVerificationReceipt(receipt))
	require.ErrorContains(t, k.SubmitVerificationReceipt(receipt), "duplicate verification receipt")
}

func keeperTestGroup() postypes.TaskGroup {
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
