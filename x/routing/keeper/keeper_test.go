package keeper

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestInitGenesisRejectsCorruptedState(t *testing.T) {
	keeper := NewKeeper()
	bad := DefaultGenesis()
	bad.Shards = []ShardConfig{
		{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 1},
		{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 2},
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "duplicate")

	bad = DefaultGenesis()
	bad.Shards = []ShardConfig{
		{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 1},
		{ZoneID: routingtypes.ZoneApplication, ActiveShards: 1},
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "sorted")
}

func TestFeatureDisabledRejectsRoutingTableMutation(t *testing.T) {
	keeper := NewKeeper()
	err := keeper.SetRoutingTable(prototype.DefaultAuthority, 1, []ShardConfig{{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 2}})
	require.ErrorContains(t, err, "disabled")
}

func TestUpdateParamsAuthoritySetRouteAndPagination(t *testing.T) {
	keeper := NewKeeper()
	require.ErrorContains(t, keeper.UpdateParams("4:0000000000000000000000000000000000000000000000000000000000000002", prototype.TestnetParams()), "authority")
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))
	require.NoError(t, keeper.SetRoutingTable(prototype.DefaultAuthority, 9, []ShardConfig{
		{ZoneID: routingtypes.ZoneApplication, ActiveShards: 4},
		{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 8},
	}))

	decision, err := keeper.Route(routingtypes.RouteInput{
		MsgType:	routingtypes.MsgTypeBankSend,
		FeeDenom:	routingtypes.NativeFeeDenom,
		TxHash:		hashBytes("tx"),
		Locality:	routingtypes.Locality{AccountKey: bytes.Repeat([]byte{1}, 20)},
		ActiveShards:	nil,
	})
	require.NoError(t, err)
	require.Equal(t, routingtypes.ZoneFinancial, decision.ZoneID)
	require.Equal(t, uint32(8), decision.ActiveShards)

	first, page, err := keeper.Shards(&prototype.PageRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.NotZero(t, page.NextOffset)

	_, _, err = keeper.Shards(&prototype.PageRequest{Offset: 99, Limit: 1})
	require.ErrorContains(t, err, "offset")
	_, _, err = keeper.Shards(nil)
	require.NoError(t, err)
}

func TestExportImportDeterministicAndMigration(t *testing.T) {
	source := NewKeeper()
	require.NoError(t, source.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))
	require.NoError(t, source.SetRoutingTable(prototype.DefaultAuthority, 1, []ShardConfig{{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 2}}))

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func hashBytes(seed string) []byte {
	sum := sha256.Sum256([]byte(seed))
	return sum[:]
}
