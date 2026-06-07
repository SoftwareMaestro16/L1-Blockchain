package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	configvotingkeeper "github.com/sovereign-l1/l1/x/config-voting/keeper"
	configvotingtypes "github.com/sovereign-l1/l1/x/config-voting/types"
)

func TestConfigVotingPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, configvotingtypes.ModuleName)
	require.Contains(t, app.keys, configvotingtypes.StoreKey)
	require.Contains(t, genesis, configvotingtypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), configvotingtypes.ModuleName)

	var gs configvotingkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[configvotingtypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Proposals)
	require.Empty(t, gs.State.Votes)
}
