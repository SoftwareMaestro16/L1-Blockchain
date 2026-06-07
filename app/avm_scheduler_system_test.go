package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	avmschedulerkeeper "github.com/sovereign-l1/l1/x/avm-scheduler/keeper"
	avmschedulertypes "github.com/sovereign-l1/l1/x/avm-scheduler/types"
)

func TestAVMSchedulerPrototypeModuleWiringAndGenesis(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Contains(t, app.ModuleManager.Modules, avmschedulertypes.ModuleName)
	require.Contains(t, app.keys, avmschedulertypes.StoreKey)
	require.Contains(t, genesis, avmschedulertypes.ModuleName)
	require.NotContains(t, GetMaccPerms(), avmschedulertypes.ModuleName)

	var gs avmschedulerkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[avmschedulertypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.ExecutionQueue)
	require.Empty(t, gs.State.ExecutionReceipts)
}
