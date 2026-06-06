package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	shardingcoordinatorkeeper "github.com/sovereign-l1/l1/x/sharding-coordinator/keeper"
	shardingcoordinatortypes "github.com/sovereign-l1/l1/x/sharding-coordinator/types"
)

func TestShardingCoordinatorPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetherCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, shardingcoordinatortypes.ModuleName)
	require.Contains(t, app.keys, shardingcoordinatortypes.StoreKey)
	require.Contains(t, genesis, shardingcoordinatortypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), shardingcoordinatortypes.ModuleName)

	var gs shardingcoordinatorkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[shardingcoordinatortypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Shards)
	require.Empty(t, gs.State.RebalanceProposals)
}
