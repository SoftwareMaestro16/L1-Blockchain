package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	identityrootkeeper "github.com/sovereign-l1/l1/x/identity-root/keeper"
	identityroottypes "github.com/sovereign-l1/l1/x/identity-root/types"
)

func TestIdentityRootPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, identityroottypes.ModuleName)
	require.Contains(t, app.keys, identityroottypes.StoreKey)
	require.Contains(t, genesis, identityroottypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), identityroottypes.ModuleName)

	var gs identityrootkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[identityroottypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Records)
	require.Equal(t, identityroottypes.DefaultRootNamespace, gs.IdentityParams.RootNamespace)
}
