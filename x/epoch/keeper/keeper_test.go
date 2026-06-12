package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	epochtypes "github.com/sovereign-l1/l1/x/epoch/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestBootstrapDefinesCurrentEpochRecordAndQueries(t *testing.T) {
	keeper := NewKeeper()
	params := testEpochParams()
	validators := testScoredValidators(t, params, 3)

	current, err := keeper.Bootstrap(params, 10, 100, 1_000, 25, validators)
	require.NoError(t, err)
	require.Equal(t, uint64(10), current.EpochID)
	require.Equal(t, uint64(100), current.StartHeight)
	require.Equal(t, uint64(124), current.EndHeight)
	require.Equal(t, postypes.EpochPhaseDelegation, current.Phase)

	queried, err := keeper.CurrentEpoch()
	require.NoError(t, err)
	require.Equal(t, current, queried)

	found, ok, err := keeper.Epoch(10)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, current, found)

	_, ok, err = keeper.Epoch(99)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestBeginHookAndPhaseTransitionKeeper(t *testing.T) {
	keeper := bootstrappedKeeper(t)

	begin, err := keeper.BeginEpoch(100, 1_000)
	require.NoError(t, err)
	require.Equal(t, epochtypes.HookEventEpochBegin, begin.Event)
	require.Equal(t, postypes.EpochPhaseDelegation, begin.ToPhase)

	transition, err := keeper.TransitionPhase(101, 1_100, postypes.EpochPhaseElection)
	require.NoError(t, err)
	require.Equal(t, epochtypes.HookEventPhaseTransition, transition.Event)
	require.Equal(t, postypes.EpochPhaseDelegation, transition.FromPhase)
	require.Equal(t, postypes.EpochPhaseElection, transition.ToPhase)

	_, err = keeper.TransitionPhase(102, 1_200, postypes.EpochPhaseActive)
	require.ErrorContains(t, err, "invalid epoch phase transition")

	hooks := keeper.HookLog()
	require.Len(t, hooks, 2)
	require.Equal(t, epochtypes.HookEventEpochBegin, hooks[0].Event)
	require.Equal(t, epochtypes.HookEventPhaseTransition, hooks[1].Event)
}

func TestSyncPhaseByTimeUsesPhaseBoundaries(t *testing.T) {
	keeper := bootstrappedKeeper(t)
	params := keeper.ExportState().Params
	durations := params.EffectivePhaseDurations()

	hooks, err := keeper.SyncPhaseByTime(110, 1_000+durations.DelegationSeconds)
	require.NoError(t, err)
	require.Len(t, hooks, 1)
	require.Equal(t, postypes.EpochPhaseElection, hooks[0].ToPhase)

	hooks, err = keeper.SyncPhaseByTime(111, 1_000+durations.DelegationSeconds+durations.ElectionSeconds+durations.AssignmentSeconds)
	require.NoError(t, err)
	require.Len(t, hooks, 2)
	require.Equal(t, postypes.EpochPhaseAssignment, hooks[0].ToPhase)
	require.Equal(t, postypes.EpochPhaseActive, hooks[1].ToPhase)
}

func TestEndHookClosesEpochAndStartsNextFromCommittedValidators(t *testing.T) {
	keeper := bootstrappedKeeper(t)
	transitionToSettlement(t, &keeper)
	params := keeper.ExportState().Params
	nextValidators := testScoredValidators(t, params, 4)

	closed, err := keeper.EndEpoch(124, 2_000, postypes.EpochSettlementRoots{
		PerformanceRoot:	postypes.PosEmptyRootHash,
		RewardRoot:		postypes.PosEmptyRootHash,
		SlashRoot:		postypes.PosEmptyRootHash,
	}, nextValidators)
	require.NoError(t, err)
	require.Equal(t, postypes.EpochPhaseClosed, closed.Phase)
	require.Equal(t, postypes.SettlementStatusFinalized, closed.SettlementStatus)

	current, err := keeper.CurrentEpoch()
	require.NoError(t, err)
	require.Equal(t, uint64(11), current.EpochID)
	require.Equal(t, uint64(125), current.StartHeight)
	require.Equal(t, uint64(149), current.EndHeight)
	require.Equal(t, postypes.EpochPhaseDelegation, current.Phase)
	expectedHash, err := postypes.ComputeValidatorSetHash(nextValidators)
	require.NoError(t, err)
	require.Equal(t, expectedHash, current.ValidatorSetHash)

	history, page, err := keeper.HistoricalEpochs(&prototype.PageRequest{Limit: 1})
	require.NoError(t, err)
	require.Empty(t, page.NextOffset)
	require.Equal(t, []postypes.EpochRecord{closed}, history)

	hooks := keeper.HookLog()
	require.Equal(t, epochtypes.HookEventEpochEnd, hooks[len(hooks)-2].Event)
	require.Equal(t, epochtypes.HookEventEpochBegin, hooks[len(hooks)-1].Event)
}

func TestEndHookRequiresSettlementPhaseAndFinalityRoots(t *testing.T) {
	keeper := bootstrappedKeeper(t)
	params := keeper.ExportState().Params

	_, err := keeper.EndEpoch(124, 2_000, postypes.EpochSettlementRoots{
		PerformanceRoot:	postypes.PosEmptyRootHash,
		RewardRoot:		postypes.PosEmptyRootHash,
		SlashRoot:		postypes.PosEmptyRootHash,
	}, testScoredValidators(t, params, 3))
	require.ErrorContains(t, err, "settlement phase")

	transitionToSettlement(t, &keeper)
	_, err = keeper.EndEpoch(124, 2_000, postypes.EpochSettlementRoots{
		PerformanceRoot:	postypes.PosEmptyRootHash,
		RewardRoot:		"not-a-root",
		SlashRoot:		postypes.PosEmptyRootHash,
	}, testScoredValidators(t, params, 3))
	require.ErrorContains(t, err, "reward root")
}

func TestDelayedDelegationActivationRuleForEpochKeeper(t *testing.T) {
	params := testEpochParams()
	params.DelegationActivationEpochs = 2

	active, err := postypes.DelegationAffectsElection(params, 10, 11)
	require.NoError(t, err)
	require.False(t, active)

	active, err = postypes.DelegationAffectsElection(params, 10, 12)
	require.NoError(t, err)
	require.True(t, active)
}

func TestExportImportDeterministicAndMigration(t *testing.T) {
	source := bootstrappedKeeper(t)
	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func bootstrappedKeeper(t *testing.T) Keeper {
	t.Helper()
	keeper := NewKeeper()
	params := testEpochParams()
	_, err := keeper.Bootstrap(params, 10, 100, 1_000, 25, testScoredValidators(t, params, 3))
	require.NoError(t, err)
	return keeper
}

func transitionToSettlement(t *testing.T, keeper *Keeper) {
	t.Helper()
	_, err := keeper.TransitionPhase(101, 1_100, postypes.EpochPhaseElection)
	require.NoError(t, err)
	_, err = keeper.TransitionPhase(102, 1_200, postypes.EpochPhaseAssignment)
	require.NoError(t, err)
	_, err = keeper.TransitionPhase(103, 1_300, postypes.EpochPhaseActive)
	require.NoError(t, err)
	_, err = keeper.TransitionPhase(104, 1_400, postypes.EpochPhaseSettlement)
	require.NoError(t, err)
}

func testEpochParams() postypes.Params {
	params := postypes.DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 2
	params.MaxTaskGroupValidators = 4
	return params
}

func testScoredValidators(t *testing.T, params postypes.Params, count int) []postypes.ScoredValidator {
	t.Helper()
	validators := make([]postypes.ScoredValidator, count)
	for i := range validators {
		candidate := postypes.Candidate{
			ValidatorID:		"val-" + string(rune('a'+i)),
			SelfStakeNaet:		sdkmath.NewInt(1_000 + int64(i)),
			DelegatedStakeNaet:	sdkmath.ZeroInt(),
			PerformanceScoreBps:	postypes.BasisPoints,
			UptimeFactorBps:	postypes.BasisPoints,
			CommissionBps:		500,
		}
		scored, err := postypes.ScoreCandidate(params, candidate)
		require.NoError(t, err)
		validators[i] = scored
	}
	return validators
}
