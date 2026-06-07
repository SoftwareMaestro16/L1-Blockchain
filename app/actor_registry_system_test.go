package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	actorregistrykeeper "github.com/sovereign-l1/l1/x/actor-registry/keeper"
	actorregistrytypes "github.com/sovereign-l1/l1/x/actor-registry/types"
)

func TestActorRegistryPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, actorregistrytypes.ModuleName)
	require.Contains(t, app.keys, actorregistrytypes.StoreKey)
	require.Contains(t, genesis, actorregistrytypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), actorregistrytypes.ModuleName)

	var gs actorregistrykeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[actorregistrytypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Actors)
	require.Empty(t, gs.State.CodeStore)
}
