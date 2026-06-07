package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	bridgehubkeeper "github.com/sovereign-l1/l1/x/bridge-hub/keeper"
	bridgehubtypes "github.com/sovereign-l1/l1/x/bridge-hub/types"
)

func TestBridgeHubPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, bridgehubtypes.ModuleName)
	require.Contains(t, app.keys, bridgehubtypes.StoreKey)
	require.Contains(t, genesis, bridgehubtypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), bridgehubtypes.ModuleName)

	var gs bridgehubkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[bridgehubtypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Bridges)
	require.Empty(t, gs.State.Events)
}
