package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestInitGenesisRejectsCorruptedState(t *testing.T) {
	keeper := NewKeeper()
	bad := DefaultGenesis()
	bad.Version = 99
	require.ErrorContains(t, keeper.InitGenesis(bad), "unsupported")

	bad = DefaultGenesis()
	bad.History = []loadtypes.Result{
		{EMA: loadtypes.EMAState{WindowHeight: 2}},
		{EMA: loadtypes.EMAState{WindowHeight: 1}},
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "sorted")
}

func TestFeatureDisabledRejectsLoadMutation(t *testing.T) {
	keeper := NewKeeper()
	_, err := keeper.ApplyMetrics(loadtypes.Metrics{CanonicalMempoolSize: 1})
	require.ErrorContains(t, err, "disabled")
}

func TestUpdateParamsAuthorityAndApplyMetrics(t *testing.T) {
	keeper := NewKeeper()
	require.ErrorContains(t, keeper.UpdateParams("4:0000000000000000000000000000000000000000000000000000000000000002", prototype.TestnetParams()), "authority")
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))

	result, err := keeper.ApplyMetrics(loadtypes.Metrics{
		CanonicalMempoolSize:		10_000,
		UsedBlockGas:			20_000_000,
		AverageInclusionDelayBlocks:	5,
		FailedTxCount:			1,
		TotalTxCount:			10,
		ExecutionStepCount:		20_000_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.EMA.WindowHeight)
}

func TestHistoryPaginationAndNilRequest(t *testing.T) {
	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))
	for i := 0; i < 3; i++ {
		_, err := keeper.ApplyMetrics(loadtypes.Metrics{CanonicalMempoolSize: uint64(i + 1)})
		require.NoError(t, err)
	}

	first, page, err := keeper.History(&prototype.PageRequest{Limit: 2})
	require.NoError(t, err)
	require.Len(t, first, 2)
	require.Equal(t, uint64(2), page.NextOffset)

	next, _, err := keeper.History(&prototype.PageRequest{Offset: page.NextOffset, Limit: 2})
	require.NoError(t, err)
	require.Len(t, next, 1)

	_, _, err = keeper.History(&prototype.PageRequest{Offset: 99, Limit: 1})
	require.ErrorContains(t, err, "offset")
	_, _, err = keeper.History(nil)
	require.NoError(t, err)
}

func TestExportImportDeterministicAndMigration(t *testing.T) {
	source := NewKeeper()
	require.NoError(t, source.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))
	_, err := source.ApplyMetrics(loadtypes.Metrics{CanonicalMempoolSize: 1})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}
