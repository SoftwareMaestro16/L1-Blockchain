package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestReporterMarketplaceSettlementReturnsDepositAndCapsRewardByPenalty(t *testing.T) {
	params := DefaultParams()
	params.ReporterRewardBps = 500
	submission := testEvidenceSubmission(t, params, SubmitterRoleReporter, "reporter-a", nil)

	settlement, err := SettleMarketplace(params, submission, true, sdkmath.NewInt(1_000))
	require.NoError(t, err)
	require.Equal(t, submission.DepositNaet, settlement.DepositReturnedNaet)
	require.Equal(t, sdkmath.NewInt(50), settlement.RewardNaet)
	require.Equal(t, submission.DepositNaet.Add(sdkmath.NewInt(50)), settlement.TotalPayoutNaet)
	require.True(t, settlement.DepositBurnedNaet.IsZero())
	require.True(t, settlement.RewardNaet.LTE(settlement.PenaltyAmountNaet))
	require.Len(t, settlement.SettlementHash, postypes.PosHashHexLength)
}

func TestInvalidEvidenceBurnsAndRedirectsDeposit(t *testing.T) {
	params := DefaultParams()
	params.InvalidEvidenceBurnBps = 6_000
	params.InvalidEvidenceRedirectBps = 4_000
	submission := testEvidenceSubmission(t, params, SubmitterRoleReporter, "reporter-a", nil)

	settlement, err := SettleMarketplace(params, submission, false, sdkmath.NewInt(1_000))
	require.NoError(t, err)
	require.True(t, settlement.DepositReturnedNaet.IsZero())
	require.True(t, settlement.RewardNaet.IsZero())
	require.Equal(t, sdkmath.NewInt(60_000_000), settlement.DepositBurnedNaet)
	require.Equal(t, sdkmath.NewInt(40_000_000), settlement.DepositRedirectedNaet)
	require.True(t, settlement.TotalPayoutNaet.IsZero())
}

func TestFishermanCanSubmitFraudProofButCannotDecideOutcome(t *testing.T) {
	params := DefaultParams()
	submission := testEvidenceSubmission(t, params, SubmitterRoleFisherman, "fish-1", func(record postypes.EvidenceRecord) postypes.EvidenceRecord {
		record.Reporter = "fish-1"
		return record
	})
	require.Equal(t, SubmitterRoleFisherman, submission.SubmitterRole)
	require.Equal(t, params.FishermanDepositNaet, submission.DepositNaet)

	settlement, err := SettleMarketplace(params, submission, true, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	require.Equal(t, params.FishermanDepositNaet, settlement.DepositReturnedNaet)
	require.Equal(t, sdkmath.NewInt(500), settlement.RewardNaet)

	_, err = NewDecisionVote(submission.Evidence.EvidenceID, "fish-1", VoterRoleFisherman, true, 1_000, testHash("fish-vote"), 88, []string{"val-000", "val-001"})
	require.ErrorContains(t, err, "fishermen cannot decide")

	vote, err := NewDecisionVote(submission.Evidence.EvidenceID, "val-000", VoterRoleValidator, true, 1_000, testHash("validator-vote"), 88, []string{"val-000", "val-001"})
	require.NoError(t, err)
	require.Equal(t, "val-000", vote.VoterID)
}

func TestDuplicateEvidenceIsRejectedByIDAndPayload(t *testing.T) {
	params := DefaultParams()
	first := testEvidenceSubmission(t, params, SubmitterRoleReporter, "reporter-a", nil)
	record := first.Evidence
	record.Reporter = "reporter-b"

	group := testVerificationGroup(t, record)
	proof := EvidenceProofPayload{
		EvidenceID:		record.EvidenceID,
		ObjectHash:		record.ObjectHash,
		ProofPayloadHash:	record.ProofPayloadHash,
		PayloadSignature:	testHash("proof-duplicate"),
	}
	_, err := SubmitEvidence(params, []EvidenceSubmission{first}, record, proof, "reporter-b", SubmitterRoleReporter, params.ReporterDepositNaet, group)
	require.ErrorContains(t, err, "duplicate evidence id")

	record.EvidenceID = "evidence-other"
	group = testVerificationGroup(t, record)
	proof.EvidenceID = record.EvidenceID
	_, err = SubmitEvidence(params, []EvidenceSubmission{first}, record, proof, "reporter-b", SubmitterRoleReporter, params.ReporterDepositNaet, group)
	require.ErrorContains(t, err, "duplicate evidence object and proof payload")
}

func TestForgedEvidencePayloadIsRejected(t *testing.T) {
	params := DefaultParams()
	record := testEvidenceRecord(t, "evidence-forged", "reporter-a")
	group := testVerificationGroup(t, record)

	_, err := SubmitEvidence(params, nil, record, EvidenceProofPayload{
		EvidenceID:		record.EvidenceID,
		ObjectHash:		testHash("wrong-object"),
		ProofPayloadHash:	record.ProofPayloadHash,
		PayloadSignature:	testHash("proof-forged"),
	}, "reporter-a", SubmitterRoleReporter, params.ReporterDepositNaet, group)
	require.ErrorContains(t, err, "object hash mismatch")

	_, err = SubmitEvidence(params, nil, record, EvidenceProofPayload{
		EvidenceID:		record.EvidenceID,
		ObjectHash:		record.ObjectHash,
		ProofPayloadHash:	record.ProofPayloadHash,
		PayloadSignature:	"not-hex",
	}, "reporter-a", SubmitterRoleReporter, params.ReporterDepositNaet, group)
	require.ErrorContains(t, err, "signature")
}

func testEvidenceSubmission(t *testing.T, params Params, role string, submitter string, mutate func(postypes.EvidenceRecord) postypes.EvidenceRecord) EvidenceSubmission {
	t.Helper()
	record := testEvidenceRecord(t, "evidence-market-1", "reporter-a")
	if mutate != nil {
		record = mutate(record)
		normalized, err := postypes.NewEvidenceRecord(record)
		require.NoError(t, err)
		record = normalized
	}
	group := testVerificationGroup(t, record)
	proof := EvidenceProofPayload{
		EvidenceID:		record.EvidenceID,
		ObjectHash:		record.ObjectHash,
		ProofPayloadHash:	record.ProofPayloadHash,
		PayloadSignature:	testHash("proof-signature"),
	}
	submission, err := SubmitEvidence(params, nil, record, proof, submitter, role, RequiredDeposit(params, role), group)
	require.NoError(t, err)
	require.Len(t, submission.SubmissionHash, postypes.PosHashHexLength)
	return submission
}

func testVerificationGroup(t *testing.T, record postypes.EvidenceRecord) postypes.EvidenceVerificationGroup {
	t.Helper()
	posParams := postypes.DefaultParams()
	validators := testScoredValidators(t, posParams, 6)
	epoch, err := postypes.NewEpochRecord(posParams, record.EpochID, 70, 100, postypes.EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	group, err := SelectVerificationGroup(posParams, epoch, validators, record, 3, 6_700)
	require.NoError(t, err)
	return group
}

func testEvidenceRecord(t *testing.T, evidenceID string, reporter string) postypes.EvidenceRecord {
	t.Helper()
	record, err := postypes.NewEvidenceRecord(postypes.EvidenceRecord{
		EvidenceID:		evidenceID,
		EvidenceType:		postypes.EvidenceTypeInvalidTaskExecutionProof,
		AccusedValidator:	"val-000",
		Reporter:		reporter,
		EpochID:		7,
		ObjectHash:		testHash("object"),
		ProofPayloadHash:	testHash("payload"),
		SubmittedHeight:	70,
	})
	require.NoError(t, err)
	return record
}

func testScoredValidators(t *testing.T, params postypes.Params, count int) []postypes.ScoredValidator {
	t.Helper()
	out := make([]postypes.ScoredValidator, count)
	for i := range out {
		candidate := postypes.Candidate{
			ValidatorID:		"val-" + threeDigits(i),
			SelfStakeNaet:		sdkmath.NewInt(1_000_000_000),
			DelegatedStakeNaet:	sdkmath.ZeroInt(),
			PerformanceScoreBps:	postypes.BasisPoints,
			UptimeFactorBps:	postypes.BasisPoints,
			CommissionBps:		500,
		}
		scored, err := postypes.ScoreCandidate(params, candidate)
		require.NoError(t, err)
		out[i] = scored
	}
	return out
}

func threeDigits(value int) string {
	return fmt.Sprintf("%03d", value)
}

func testHash(parts ...string) string {
	return hashParts(parts...)
}
