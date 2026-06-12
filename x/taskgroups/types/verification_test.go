package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestVerificationReceiptAcceptsAllResultValuesAndHashesDeterministically(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	results := []string{
		VerificationResultValid,
		VerificationResultInvalid,
		VerificationResultAbstain,
		VerificationResultUnavailable,
	}
	for _, result := range results {
		receipt, err := NewVerificationReceipt(group, "val-a", objectHash, result, "sig-"+result, sdkmath.NewInt(10), 40)
		require.NoError(t, err)
		require.Equal(t, group.EpochID, receipt.EpochID)
		require.Equal(t, group.TaskGroupID, receipt.TaskGroupID)
		require.Equal(t, group.WorkloadID, receipt.WorkloadID)
		require.Len(t, ComputeVerificationReceiptHash(receipt), 64)
	}
}

func TestVerificationReceiptRejectsInvalidMembershipSignatureHashAndWindow(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	_, err := NewVerificationReceipt(group, "val-x", objectHash, VerificationResultValid, "sig", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "not assigned")

	_, err = NewVerificationReceipt(group, "val-a", "bad", VerificationResultValid, "sig", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "hex chars")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, "maybe", "sig", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "unsupported verification result")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "", sdkmath.ZeroInt(), 40)
	require.ErrorContains(t, err, "signature")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig", sdkmath.NewInt(-1), 40)
	require.ErrorContains(t, err, "gas or cost")

	_, err = NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig", sdkmath.ZeroInt(), 61)
	require.ErrorContains(t, err, "activity window")
}

func TestVerificationReceiptSetSortsAndRootsReceipts(t *testing.T) {
	group := proposerTestGroup()
	hashA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hashB := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	receiptB, err := NewVerificationReceipt(group, "val-b", hashB, VerificationResultValid, "sig-b", sdkmath.NewInt(20), 40)
	require.NoError(t, err)
	receiptA, err := NewVerificationReceipt(group, "val-a", hashA, VerificationResultInvalid, "sig-a", sdkmath.NewInt(10), 40)
	require.NoError(t, err)

	left, err := NewVerificationReceiptSet(group, []VerificationReceipt{receiptB, receiptA})
	require.NoError(t, err)
	right, err := NewVerificationReceiptSet(group, []VerificationReceipt{receiptA, receiptB})
	require.NoError(t, err)
	require.Equal(t, left.Root, right.Root)
	require.Equal(t, []VerificationReceipt{receiptA, receiptB}, left.Receipts)
	require.NoError(t, left.Validate(group))

	_, err = NewVerificationReceiptSet(group, []VerificationReceipt{receiptA, receiptA})
	require.ErrorContains(t, err, "duplicate verification receipt")
}

func TestCrossDomainProofVerificationCoversConfiguredProofKinds(t *testing.T) {
	group := proposerTestGroup()
	kinds := []string{
		ProofKindZoneRoot,
		ProofKindShardRoot,
		ProofKindMessageRoot,
		ProofKindReceiptRoot,
		ProofKindIdentity,
		ProofKindPaymentSettlement,
	}
	for _, kind := range kinds {
		proof := CrossDomainProof{
			ProofID:	"proof-" + kind,
			WorkloadID:	group.WorkloadID,
			ProofKind:	kind,
			SubjectID:	"subject-1",
			RootHash:	"1111111111111111111111111111111111111111111111111111111111111111",
			ParentRootHash:	"2222222222222222222222222222222222222222222222222222222222222222",
			ProofHash:	"3333333333333333333333333333333333333333333333333333333333333333",
			CreatedHeight:	40,
		}
		require.NoError(t, VerifyCrossDomainProof(group, proof))
	}
}

func TestCrossDomainProofVerificationRejectsMismatchesAndUnconfiguredContractExecution(t *testing.T) {
	group := proposerTestGroup()
	proof := CrossDomainProof{
		ProofID:	"proof-contract",
		WorkloadID:	group.WorkloadID,
		ProofKind:	ProofKindContractExecution,
		SubjectID:	"contract-1",
		RootHash:	"1111111111111111111111111111111111111111111111111111111111111111",
		ParentRootHash:	"2222222222222222222222222222222222222222222222222222222222222222",
		ProofHash:	"3333333333333333333333333333333333333333333333333333333333333333",
		CreatedHeight:	40,
	}
	require.ErrorContains(t, VerifyCrossDomainProof(group, proof), "not configured")

	proof.ContractExecutionRequired = true
	require.NoError(t, VerifyCrossDomainProof(group, proof))

	proof.WorkloadID = "other"
	require.ErrorContains(t, VerifyCrossDomainProof(group, proof), "workload mismatch")

	proof.WorkloadID = group.WorkloadID
	proof.ProofHash = "bad"
	require.ErrorContains(t, VerifyCrossDomainProof(group, proof), "hex chars")
}

func TestAggregateVerificationReceiptsTracksParticipationAndInvalidEvidence(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	receiptA, err := NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig-a", sdkmath.NewInt(10), 40)
	require.NoError(t, err)
	receiptB, err := NewVerificationReceipt(group, "val-b", objectHash, VerificationResultInvalid, "sig-b", sdkmath.NewInt(10), 40)
	require.NoError(t, err)
	receiptC, err := NewVerificationReceipt(group, "val-c", objectHash, VerificationResultAbstain, "sig-c", sdkmath.NewInt(10), 40)
	require.NoError(t, err)
	set, err := NewVerificationReceiptSet(group, []VerificationReceipt{receiptC, receiptB, receiptA})
	require.NoError(t, err)

	aggregation, err := AggregateVerificationReceipts(group, set, objectHash, 6_000)
	require.NoError(t, err)
	require.Equal(t, uint32(1), aggregation.ValidCount)
	require.Equal(t, uint32(1), aggregation.InvalidCount)
	require.Equal(t, uint32(1), aggregation.AbstainCount)
	require.Equal(t, uint32(0), aggregation.UnavailableCount)
	require.Equal(t, uint32(10_000), aggregation.ParticipationBps)
	require.True(t, aggregation.QuorumReached)
	require.NotNil(t, aggregation.InvalidEvidence)
	require.Len(t, aggregation.InvalidEvidence.InvalidReceipts, 1)
	require.Equal(t, receiptB, aggregation.InvalidEvidence.InvalidReceipts[0])
	require.Len(t, aggregation.InvalidEvidence.EvidenceHash, 64)
}

func TestTrackVerifierParticipationIncludesMissingValidatorsAsUnavailable(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	receiptA, err := NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig-a", sdkmath.NewInt(10), 40)
	require.NoError(t, err)
	set, err := NewVerificationReceiptSet(group, []VerificationReceipt{receiptA})
	require.NoError(t, err)

	records, err := TrackVerifierParticipation(group, set, objectHash)
	require.NoError(t, err)
	require.Len(t, records, 3)
	require.Equal(t, "val-a", records[0].ValidatorAddress)
	require.True(t, records[0].Participated)
	require.Len(t, records[0].ReceiptHash, 64)
	require.Equal(t, "val-b", records[1].ValidatorAddress)
	require.False(t, records[1].Participated)
	require.Equal(t, VerificationResultUnavailable, records[1].Result)
	require.Equal(t, "val-c", records[2].ValidatorAddress)
	require.False(t, records[2].Participated)
	require.Equal(t, VerificationResultUnavailable, records[2].Result)
}

func TestInvalidResultEvidenceRejectsNonInvalidReceipts(t *testing.T) {
	group := proposerTestGroup()
	objectHash := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	receipt, err := NewVerificationReceipt(group, "val-a", objectHash, VerificationResultValid, "sig-a", sdkmath.ZeroInt(), 40)
	require.NoError(t, err)

	_, err = BuildInvalidResultEvidence(group, objectHash, []VerificationReceipt{receipt})
	require.ErrorContains(t, err, "only accepts invalid receipts")
}

func TestRequiredVerificationDutiesEnumeratesValidatorDuties(t *testing.T) {
	require.Equal(t, []string{
		VerificationDutyReexecuteStateTransition,
		VerificationDutyValidateCrossDomainProof,
		VerificationDutyVerifyTaskGroup,
		VerificationDutyValidateConsensusOrdering,
		VerificationDutyVerifyMessageReceipt,
		VerificationDutySignValidOutput,
		VerificationDutyRejectInvalidOutput,
		VerificationDutySubmitEvidence,
	}, RequiredVerificationDuties())
}
