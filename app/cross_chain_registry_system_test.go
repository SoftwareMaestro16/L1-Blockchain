package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	crosschainregistrykeeper "github.com/sovereign-l1/l1/x/cross-chain-registry/keeper"
	crosschainregistrytypes "github.com/sovereign-l1/l1/x/cross-chain-registry/types"
)

func TestCrossChainRegistryPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, crosschainregistrytypes.ModuleName)
	require.Contains(t, app.keys, crosschainregistrytypes.StoreKey)
	require.Contains(t, genesis, crosschainregistrytypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), crosschainregistrytypes.ModuleName)

	var gs crosschainregistrykeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[crosschainregistrytypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Chains)
	require.Empty(t, gs.State.BridgeRoutes)
}
