package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	storagerentkeeper "github.com/sovereign-l1/l1/x/storage-rent/keeper"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
)

func TestStorageRentPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, storagerenttypes.ModuleName)
	require.Contains(t, app.keys, storagerenttypes.StoreKey)
	require.Contains(t, genesis, storagerenttypes.ModuleName)
	require.Contains(t, GetMaccPerms(), storagerenttypes.ModuleName)
	require.Nil(t, GetMaccPerms()[storagerenttypes.ModuleName])

	var gs storagerentkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[storagerenttypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Contracts)
	require.Empty(t, gs.State.Distributions)
}
