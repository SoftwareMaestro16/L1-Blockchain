package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/evidence/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestValidEvidenceAccepted(t *testing.T) {
	k := NewKeeper()
	record := submitEvidence(t, &k, "evidence-valid", types.EvidenceTypeConsensus, true, 1)

	_, err := k.VoteEvidence(types.MsgVoteEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Voter:		rawAddress("44"),
		Accept:		true,
		VotingPowerBps:	7_000,
		Height:		2,
	})
	require.NoError(t, err)

	finalized, err := k.FinalizeEvidence(types.MsgFinalizeEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusAccepted, finalized.Status)
	require.True(t, finalized.SlashDecision.Applied)
	require.True(t, finalized.RewardDecision.Paid)
	require.Len(t, k.SlashEvents(), 1)
	require.Len(t, k.ReporterRewards(), 1)
}

func TestMalformedEvidenceRejected(t *testing.T) {
	k := NewKeeper()
	_, err := k.SubmitEvidence(types.MsgSubmitEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"bad",
		EvidenceType:		"unknown",
		AccusedValidator:	rawAddress("11"),
		Reporter:		rawAddress("22"),
		ProofPayloadHash:	proofHash("bad"),
		PayloadSizeBytes:	128,
		Height:			1,
	})
	require.ErrorContains(t, err, "unsupported evidence type")

	_, err = k.SubmitEvidence(types.MsgSubmitEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"bad-hash",
		EvidenceType:		types.EvidenceTypeFraud,
		AccusedValidator:	rawAddress("11"),
		Reporter:		rawAddress("22"),
		ProofPayloadHash:	"not hex",
		PayloadSizeBytes:	128,
		Height:			1,
	})
	require.ErrorContains(t, err, "proof payload hash")
}

func TestDuplicateEvidenceRejected(t *testing.T) {
	k := NewKeeper()
	record := submitEvidence(t, &k, "evidence-duplicate", types.EvidenceTypeFraud, false, 1)

	_, err := k.SubmitEvidence(validSubmit("evidence-duplicate", types.EvidenceTypeFraud, proofHash("other"), 2))
	require.ErrorContains(t, err, "duplicate evidence id")

	_, err = k.SubmitEvidence(validSubmit("evidence-other", types.EvidenceTypeFraud, record.ProofPayloadHash, 2))
	require.ErrorContains(t, err, "duplicate proof payload hash")
}

func TestExpiredEvidenceIgnored(t *testing.T) {
	k := NewKeeper()
	gs := k.ExportGenesis()
	gs.Params.EvidenceTTLBlocks = 2
	require.NoError(t, k.InitGenesis(gs))
	record := submitEvidence(t, &k, "evidence-expired", types.EvidenceTypeMissedBlock, false, 10)

	finalized, err := k.FinalizeEvidence(types.MsgFinalizeEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		13,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusExpired, finalized.Status)
	require.Empty(t, k.SlashEvents())
	require.Empty(t, k.ReporterRewards())

	_, err = k.CancelExpiredEvidence(types.MsgCancelExpiredEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		14,
	})
	require.ErrorContains(t, err, "pending")
}

func TestSlashEventUpdatesRegistryAndTombstoneIsIrreversible(t *testing.T) {
	k := NewKeeper()
	record := submitEvidence(t, &k, "evidence-critical", types.EvidenceTypeConsensus, false, 1)

	finalized, err := k.FinalizeEvidence(types.MsgFinalizeEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		2,
	})
	require.NoError(t, err)
	require.True(t, finalized.SlashDecision.Tombstone)

	require.Equal(t, []types.SlashEvent{{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		FractionBps:		finalized.SlashDecision.FractionBps,
		Tombstone:		true,
		Height:			2,
	}}, k.SlashEvents())
	require.Equal(t, []types.RegistryUpdate{{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		Status:			types.RegistryStatusTombstoned,
		Height:			2,
	}}, k.RegistryUpdates())
	require.Contains(t, k.TombstonedValidators(), record.AccusedValidator)

	_, err = k.FinalizeEvidence(types.MsgFinalizeEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		3,
	})
	require.ErrorContains(t, err, "only be finalized once")
	require.Contains(t, k.TombstonedValidators(), record.AccusedValidator)
}

func TestReporterRewardPaidOnce(t *testing.T) {
	k := NewKeeper()
	record := submitEvidence(t, &k, "evidence-reward", types.EvidenceTypePerformance, false, 1)

	_, err := k.FinalizeEvidence(types.MsgFinalizeEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		2,
	})
	require.NoError(t, err)
	require.Len(t, k.ReporterRewards(), 1)
	require.True(t, k.ReporterRewards()[0].Paid)

	_, err = k.FinalizeEvidence(types.MsgFinalizeEvidence{
		Authority:	prototype.DefaultAuthority,
		EvidenceID:	record.EvidenceID,
		Height:		3,
	})
	require.ErrorContains(t, err, "only be finalized once")
	require.Len(t, k.ReporterRewards(), 1)
}

func TestExportImportPreservesPendingEvidence(t *testing.T) {
	source := NewKeeper()
	submitEvidence(t, &source, "evidence-z", types.EvidenceTypeFraud, true, 1)
	submitEvidence(t, &source, "evidence-a", types.EvidenceTypeMissedBlock, false, 1)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	require.Len(t, exported.State.Evidence, 2)
	require.Equal(t, "evidence-a", exported.State.Evidence[0].EvidenceID)

	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
	require.Len(t, target.PendingEvidence(), 2)
}

func TestEvidenceProcessingCannotPanicOnInvalidPayload(t *testing.T) {
	k := NewKeeper()
	require.NotPanics(t, func() {
		_, err := k.SubmitEvidence(types.MsgSubmitEvidence{
			Authority:		prototype.DefaultAuthority,
			EvidenceID:		"bad-payload",
			EvidenceType:		types.EvidenceTypeFraud,
			AccusedValidator:	rawAddress("11"),
			Reporter:		rawAddress("22"),
			ProofPayloadHash:	"   ",
			PayloadSizeBytes:	128,
			Height:			1,
		})
		require.Error(t, err)
	})
}

func TestDoubleSignEvidencePipelineJailsTombstonesFreezesStakeAndCallsHooks(t *testing.T) {
	k := NewKeeper()
	hooks := &recordingSlashingHooks{}
	k.SetSlashingIntegrationHooks(hooks)
	validator := rawAddress("33")

	record, err := k.ProcessDoubleSignEvidence(types.MsgSubmitDoubleSignEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"double-sign-1",
		AccusedValidator:	validator,
		Reporter:		rawAddress("22"),
		VoteAHash:		proofHash("vote-a"),
		VoteBHash:		proofHash("vote-b"),
		InfractionHeight:	9,
		Height:			10,
		ValidatorStake:		1_000_000,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusAccepted, record.Status)
	require.Equal(t, types.EvidenceTypeDoubleSign, record.EvidenceType)
	require.True(t, record.SlashDecision.Tombstone)

	events := k.SlashEvents()
	require.Len(t, events, 1)
	require.Equal(t, types.SlashingReasonDoubleSign, events[0].Reason)
	require.True(t, events[0].Tombstone)
	require.Equal(t, uint64(1), events[0].OffenseCount)
	require.Equal(t, uint64(50_000), events[0].FrozenStake)
	require.Contains(t, k.TombstonedValidators(), validator)
	require.Len(t, k.JailRecords(), 1)
	require.True(t, k.JailRecords()[0].Active)
	require.True(t, k.JailRecords()[0].Tombstone)
	require.Len(t, k.FrozenStakes(), 1)
	require.Equal(t, uint64(50_000), k.FrozenStakes()[0].Amount)
	require.Zero(t, k.FrozenStakes()[0].ReleaseHeight)
	require.ErrorContains(t, k.ValidateActiveSetInvariant([]string{validator}), "jailed validator")

	require.Equal(t, []string{
		"slash:double-sign-1",
		"jail:double-sign-1:true:0",
		"freeze:double-sign-1:50000:0",
		"reputation:double-sign-1:double-sign",
		"insurance:double-sign-1:50000",
	}, hooks.calls)
	require.Len(t, k.IntegrationEvents(), 3)
}

func TestDowntimeEvidencePipelineSupportsRepeatedOffenseAndUnjail(t *testing.T) {
	k := NewKeeper()
	validator := rawAddress("44")

	first, err := k.ProcessDowntimeEvidence(types.MsgSubmitDowntimeEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"downtime-1",
		AccusedValidator:	validator,
		Reporter:		rawAddress("22"),
		MissedBlocks:		51,
		WindowBlocks:		100,
		InfractionHeight:	5,
		Height:			6,
		ValidatorStake:		1_000_000,
	})
	require.NoError(t, err)
	require.Equal(t, types.EvidenceTypeDowntime, first.EvidenceType)
	require.False(t, first.SlashDecision.Tombstone)
	require.Equal(t, types.DefaultMinSlashFractionBps, first.SlashDecision.FractionBps)
	require.Equal(t, uint64(1), k.OffenseCounters()[0].Count)
	firstJailUntil := uint64(6 + types.DefaultDowntimeFirstJailBlocks)
	require.Equal(t, firstJailUntil, k.JailRecords()[0].JailedUntilHeight)
	require.ErrorContains(t, k.UnjailValidator(types.MsgUnjailValidator{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Height:			firstJailUntil - 1,
	}), "before jail period")
	require.NoError(t, k.UnjailValidator(types.MsgUnjailValidator{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Height:			firstJailUntil,
	}))
	require.NoError(t, k.ValidateActiveSetInvariant([]string{validator}))

	second, err := k.ProcessDowntimeEvidence(types.MsgSubmitDowntimeEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"downtime-2",
		AccusedValidator:	validator,
		Reporter:		rawAddress("22"),
		MissedBlocks:		60,
		WindowBlocks:		100,
		InfractionHeight:	firstJailUntil + 1,
		Height:			firstJailUntil + 2,
		ValidatorStake:		1_000_000,
	})
	require.NoError(t, err)
	require.Equal(t, types.DefaultMinSlashFractionBps*types.DefaultDowntimeRepeatMultiplier, second.SlashDecision.FractionBps)
	require.Equal(t, uint64(2), k.OffenseCounters()[0].Count)
	require.Equal(t, uint64(firstJailUntil+2+types.DefaultDowntimeRepeatJailBlocks), k.JailRecords()[1].JailedUntilHeight)
	require.ErrorContains(t, k.ValidateActiveSetInvariant([]string{validator}), "jailed validator")
}

func TestInvalidObjectiveEvidenceRejectedSafely(t *testing.T) {
	k := NewKeeper()
	hash := proofHash("same")
	_, err := k.ProcessDoubleSignEvidence(types.MsgSubmitDoubleSignEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"bad-double-sign",
		AccusedValidator:	rawAddress("33"),
		Reporter:		rawAddress("22"),
		VoteAHash:		hash,
		VoteBHash:		hash,
		InfractionHeight:	9,
		Height:			10,
		ValidatorStake:		1_000_000,
	})
	require.ErrorContains(t, err, "distinct")
	require.Empty(t, k.SlashEvents())

	_, err = k.ProcessDowntimeEvidence(types.MsgSubmitDowntimeEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		"bad-downtime",
		AccusedValidator:	rawAddress("33"),
		Reporter:		rawAddress("22"),
		MissedBlocks:		101,
		WindowBlocks:		100,
		InfractionHeight:	9,
		Height:			10,
		ValidatorStake:		1_000_000,
	})
	require.ErrorContains(t, err, "missed/window")
	require.Empty(t, k.SlashEvents())
}

func submitEvidence(t *testing.T, k *Keeper, id string, evidenceType string, review bool, height uint64) types.EvidenceRecord {
	t.Helper()
	record, err := k.SubmitEvidence(validSubmit(id, evidenceType, proofHash(id), height, withReview(review)))
	require.NoError(t, err)
	return record
}

func validSubmit(id string, evidenceType string, hash string, height uint64, opts ...func(*types.MsgSubmitEvidence)) types.MsgSubmitEvidence {
	msg := types.MsgSubmitEvidence{
		Authority:		prototype.DefaultAuthority,
		EvidenceID:		id,
		EvidenceType:		evidenceType,
		AccusedValidator:	rawAddress("11"),
		Reporter:		rawAddress("22"),
		ProofPayloadHash:	hash,
		PayloadSizeBytes:	128,
		Height:			height,
	}
	for _, opt := range opts {
		opt(&msg)
	}
	return msg
}

func withReview(review bool) func(*types.MsgSubmitEvidence) {
	return func(msg *types.MsgSubmitEvidence) {
		msg.RequiresReview = review
	}
}

func proofHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func rawAddress(hexByte string) string {
	return "4:000000000000000000000000" + fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte)
}

type recordingSlashingHooks struct {
	calls []string
}

func (h *recordingSlashingHooks) RecordSlashingEvent(event types.SlashEvent) error {
	h.calls = append(h.calls, "slash:"+event.EvidenceID)
	return nil
}

func (h *recordingSlashingHooks) JailValidator(_ string, tombstone bool, jailUntilHeight uint64, evidenceID string) error {
	h.calls = append(h.calls, fmt.Sprintf("jail:%s:%t:%d", evidenceID, tombstone, jailUntilHeight))
	return nil
}

func (h *recordingSlashingHooks) UnjailValidator(_ string, height uint64) error {
	h.calls = append(h.calls, fmt.Sprintf("unjail:%d", height))
	return nil
}

func (h *recordingSlashingHooks) FreezeStake(_ string, evidenceID string, amount uint64, releaseHeight uint64) error {
	h.calls = append(h.calls, fmt.Sprintf("freeze:%s:%d:%d", evidenceID, amount, releaseHeight))
	return nil
}

func (h *recordingSlashingHooks) ApplyReputationPenalty(_ string, evidenceID string, evidenceType string, _ uint64) error {
	h.calls = append(h.calls, "reputation:"+evidenceID+":"+evidenceType)
	return nil
}

func (h *recordingSlashingHooks) ApplyInsuranceSlash(_ string, evidenceID string, amount uint64, _ uint64) error {
	h.calls = append(h.calls, fmt.Sprintf("insurance:%s:%d", evidenceID, amount))
	return nil
}
