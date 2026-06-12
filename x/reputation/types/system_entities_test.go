package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReputationMissedBlockPenalty(t *testing.T) {
	state := newTestReputationState(t)
	validator := ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 50})
	state.Validators = []ReputationRecord{validator}

	next, err := ApplyReputationPenalty(state, MsgApplyReputationPenalty{
		Authority:	state.Params.Authority,
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
		Component:	ComponentMissedBlock,
		Reason:		"missed block",
		Epoch:		2,
	})
	require.NoError(t, err)
	record, found := QueryValidatorReputation(next, addr(1))
	require.True(t, found)
	require.Equal(t, uint16(next.Params.MissedBlockPenalty), record.FailedTxPenalty)
	require.Less(t, record.Score, validator.Score)
	require.Len(t, next.PenaltyEvents, 1)
	require.NoError(t, CheckReputationInvariants(next))
}

func TestReputationUptimeReward(t *testing.T) {
	state := newTestReputationState(t)
	state.Validators = []ReputationRecord{ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 10})}

	next, err := ApplyReputationReward(state, MsgApplyReputationReward{
		Authority:	state.Params.Authority,
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
		Component:	ComponentUptime,
		Reason:		"uptime",
		Epoch:		3,
	})
	require.NoError(t, err)
	record, found := QueryValidatorReputation(next, addr(1))
	require.True(t, found)
	require.Equal(t, uint16(next.Params.UptimeReward), record.TxSuccessScore)
	require.Greater(t, record.Score, uint8(10))
	require.Len(t, next.RecoveryEvents, 1)
	require.NoError(t, CheckReputationInvariants(next))
}

func TestReputationSlashingPenaltyAlwaysReducesScore(t *testing.T) {
	state := newTestReputationState(t)
	validator := ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 80})
	state.Validators = []ReputationRecord{validator}

	next, err := ApplyReputationPenalty(state, MsgApplyReputationPenalty{
		Authority:	state.Params.Authority,
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
		Component:	ComponentSlashing,
		Reason:		"double sign",
		Epoch:		4,
	})
	require.NoError(t, err)
	record, found := QueryValidatorReputation(next, addr(1))
	require.True(t, found)
	require.Equal(t, uint16(next.Params.SlashingPenalty), record.SlashPenalty)
	require.Less(t, record.Score, validator.Score)

	badParams := state.Params
	badParams.SlashingPenalty = 0
	_, err = ApplyUpdateReputationParams(state, MsgUpdateReputationParams{Authority: state.Params.Authority, Params: badParams})
	require.ErrorContains(t, err, "slashing penalty")
}

func TestReputationScoreFloorAndCeiling(t *testing.T) {
	state := newTestReputationState(t)

	next, err := ApplyReputationPenalty(state, MsgApplyReputationPenalty{
		Authority:	state.Params.Authority,
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
		Component:	ComponentSpam,
		Amount:		500,
		Reason:		"spam",
		Epoch:		1,
	})
	require.NoError(t, err)
	low, found := QueryValidatorReputation(next, addr(1))
	require.True(t, found)
	require.Equal(t, ScoreMin, low.Score)

	next, err = ApplyReputationReward(next, MsgApplyReputationReward{
		Authority:	next.Params.Authority,
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
		Component:	ComponentUptime,
		Amount:		1000,
		Reason:		"long uptime",
		Epoch:		2,
	})
	require.NoError(t, err)
	high, found := QueryValidatorReputation(next, addr(1))
	require.True(t, found)
	require.Equal(t, ScoreMax, high.Score)
}

func TestReputationDeterministicRecomputation(t *testing.T) {
	stateA := newTestReputationState(t)
	stateB := newTestReputationState(t)
	stateA.Validators = []ReputationRecord{
		ApplyComputedScore(ReputationRecord{Account: addr(2), AgeScore: 20, TxSuccessScore: 10}),
		ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 50, SlashPenalty: 3}),
	}
	stateB.Validators = []ReputationRecord{
		stateA.Validators[1],
		stateA.Validators[0],
	}

	msg := MsgRecomputeReputation{Authority: stateA.Params.Authority, SubjectType: SubjectValidator, Subject: addr(1), Epoch: 8}
	nextA, err := ApplyRecomputeReputation(stateA, msg)
	require.NoError(t, err)
	nextB, err := ApplyRecomputeReputation(stateB, msg)
	require.NoError(t, err)
	require.Equal(t, NormalizeReputationState(nextA).Validators, NormalizeReputationState(nextB).Validators)

	exportA, err := ExportReputationState(nextA)
	require.NoError(t, err)
	exportB, err := ExportReputationState(nextB)
	require.NoError(t, err)
	require.Equal(t, exportA, exportB)
}

func TestReputationExportImportPreservesSnapshots(t *testing.T) {
	state := newTestReputationState(t)
	state.Validators = []ReputationRecord{ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 50})}
	state.Reporters = []ReputationRecord{ApplyComputedScore(ReputationRecord{Account: addr(2), TxSuccessScore: 30})}

	next, snapshot, err := SnapshotReputationEpoch(state, 9)
	require.NoError(t, err)
	require.Equal(t, ComputeReputationSnapshotHash(snapshot), snapshot.SnapshotHash)
	exported, err := ExportReputationState(next)
	require.NoError(t, err)
	imported, err := ImportReputationState(exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
	require.Len(t, imported.Snapshots, 1)

	snapshots, events, err := QueryReputationHistory(imported, ReputationHistoryQuery{
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
	})
	require.NoError(t, err)
	require.Len(t, snapshots, 1)
	require.Empty(t, events)
}

func TestReputationSecurityAuthorizationAndTamperRejection(t *testing.T) {
	state := newTestReputationState(t)
	_, err := ApplyReputationPenalty(state, MsgApplyReputationPenalty{
		Authority:	"4:0000000000000000000000000000000000000000000000000000000000000002",
		SubjectType:	SubjectValidator,
		Subject:	addr(1),
		Component:	ComponentMissedBlock,
		Epoch:		1,
	})
	require.ErrorContains(t, err, "requires authority")

	tampered := state
	tampered.Validators = []ReputationRecord{{
		Account:	addr(1),
		Score:		ScoreMax,
		AgeScore:	1,
	}}
	require.ErrorContains(t, CheckReputationInvariants(tampered), "score mismatch")

	valid, err := ApplyReputationReward(state, MsgApplyReputationReward{
		Authority:	state.Params.Authority,
		SubjectType:	SubjectReporter,
		Subject:	addr(2),
		Component:	ComponentRecovery,
		Epoch:		2,
	})
	require.NoError(t, err)
	_, found := QueryReporterReputation(valid, addr(2))
	require.True(t, found)
	require.NoError(t, CheckReputationInvariants(valid))
}

func newTestReputationState(t *testing.T) ReputationState {
	t.Helper()
	state, err := NewReputationState(DefaultReputationParams())
	require.NoError(t, err)
	return state
}
