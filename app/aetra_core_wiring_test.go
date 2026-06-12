package app

import (
	"encoding/json"
	"slices"
	"testing"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log/v2"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"

	aetracorekeeper "github.com/sovereign-l1/l1/x/aetracore/keeper"
	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	contractskeeper "github.com/sovereign-l1/l1/x/contracts/keeper"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	networkingkeeper "github.com/sovereign-l1/l1/x/networking/keeper"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
	paymentskeeper "github.com/sovereign-l1/l1/x/payments/keeper"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	schedulerkeeper "github.com/sovereign-l1/l1/x/scheduler/keeper"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAetraCoreWiringGateRegistersPrototypeModulesDisabled(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetraCoreWiringGate())
	require.Equal(t, RoutingExecutionPointAnteAdmissionOnly, AetraCoreRoutingExecutionPoint())

	prototypeModuleNames := AetraCorePrototypeModuleNames()
	prototypeStoreKeys := AetraCorePrototypeStoreKeys()
	require.Len(t, prototypeStoreKeys, len(prototypeModuleNames))
	for i, moduleName := range prototypeModuleNames {
		require.Contains(t, app.ModuleManager.Modules, moduleName)
		require.Contains(t, app.keys, prototypeStoreKeys[i])
		require.Contains(t, genesis, moduleName)
		if IsReservedSystemModuleAccountName(moduleName) {
			require.Contains(t, GetMaccPerms(), moduleName)
			require.Nil(t, GetMaccPerms()[moduleName])
		} else {
			require.NotContains(t, GetMaccPerms(), moduleName)
		}
		_, hasBegin := app.ModuleManager.Modules[moduleName].(appmodule.HasBeginBlocker)
		_, hasEnd := app.ModuleManager.Modules[moduleName].(appmodule.HasEndBlocker)
		require.False(t, hasBegin, moduleName)
		require.False(t, hasEnd, moduleName)
		require.True(t, slices.Contains(app.ModuleManager.OrderBeginBlockers, moduleName), moduleName)
		require.True(t, slices.Contains(app.ModuleManager.OrderEndBlockers, moduleName), moduleName)
	}

	aetherCoreGenesis := decodeJSONGenesis[aetracorekeeper.GenesisState](t, genesis[aetracoretypes.ModuleName])
	require.False(t, aetherCoreGenesis.Params.Enabled)
	require.Empty(t, aetherCoreGenesis.State.ZoneDescriptors)
	require.Empty(t, aetherCoreGenesis.State.ServiceDescriptors)
	require.Empty(t, aetherCoreGenesis.State.ZoneCommitments)
	require.Empty(t, aetherCoreGenesis.State.GlobalRoots)

	loadGenesis := decodeJSONGenesis[loadkeeper.GenesisState](t, genesis[loadtypes.ModuleName])
	require.False(t, loadGenesis.Params.Enabled)
	require.Empty(t, loadGenesis.History)

	routingGenesis := decodeJSONGenesis[routingkeeper.GenesisState](t, genesis[routingtypes.ModuleName])
	require.False(t, routingGenesis.Params.Enabled)
	require.Empty(t, routingGenesis.Shards)

	zonesGenesis := decodeJSONGenesis[zoneskeeper.GenesisState](t, genesis[zonestypes.ModuleName])
	require.False(t, zonesGenesis.Params.Enabled)
	require.Empty(t, zonesGenesis.State.ActiveZones)
	require.Empty(t, zonesGenesis.State.Commitments)

	meshGenesis := decodeJSONGenesis[meshkeeper.GenesisState](t, genesis[meshtypes.ModuleName])
	require.False(t, meshGenesis.Params.Enabled)
	require.Empty(t, meshGenesis.State.Destinations)
	require.Empty(t, meshGenesis.State.ReplayMarkers)

	networkingGenesis := decodeJSONGenesis[networkingkeeper.GenesisState](t, genesis[networkingtypes.ModuleName])
	require.False(t, networkingGenesis.Params.Enabled)
	require.NotEmpty(t, networkingGenesis.State.ChannelPolicies)
	require.Empty(t, networkingGenesis.State.NodeRecords)
	require.Empty(t, networkingGenesis.State.Sessions)

	paymentsGenesis := decodeJSONGenesis[paymentskeeper.GenesisState](t, genesis[paymentstypes.ModuleName])
	require.False(t, paymentsGenesis.Params.Enabled)
	require.Empty(t, paymentsGenesis.State.Channels)
	require.Empty(t, paymentsGenesis.State.Settlements)

	schedulerGenesis := decodeJSONGenesis[schedulerkeeper.GenesisState](t, genesis[schedulertypes.ModuleName])
	require.False(t, schedulerGenesis.Params.Enabled)
	require.Empty(t, schedulerGenesis.State.Jobs)
	require.Empty(t, schedulerGenesis.State.History)

	contractsGenesis := decodeJSONGenesis[contractstypes.GenesisState](t, genesis[contractstypes.ModuleName])
	require.NoError(t, contractsGenesis.Validate())
	require.True(t, contractsGenesis.Params.Enabled)
	require.Empty(t, contractsGenesis.State.Codes)
	require.Equal(t, contractskeeper.DefaultGenesis().StateRoot, contractsGenesis.StateRoot)
}

func TestFeatureDisabledMainnetProfileHasNoActiveProductionShardingBehavior(t *testing.T) {
	app := Setup(t, false)

	_, err := app.LoadKeeper.ApplyMetrics(loadtypes.Metrics{CanonicalMempoolSize: 1})
	require.ErrorContains(t, err, "disabled")
	err = app.RoutingKeeper.SetRoutingTable("4:0000000000000000000000000000000000000000000000000000000000000001", 1, []routingkeeper.ShardConfig{{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 1}})
	require.ErrorContains(t, err, "disabled")
	err = app.ZonesKeeper.RegisterZone(zonestypes.Zone{})
	require.ErrorContains(t, err, "disabled")
	err = app.MeshKeeper.RegisterDestination(meshtypes.MeshDestination{})
	require.ErrorContains(t, err, "disabled")
	err = app.NetworkingKeeper.RegisterNodeRecord(networkingtypes.NodeRecord{}, nil, 1)
	require.ErrorContains(t, err, "disabled")
	err = app.PaymentsKeeper.OpenChannel(paymentstypes.ChannelRecord{})
	require.ErrorContains(t, err, "disabled")
	err = app.AetraCoreKeeper.RegisterZoneDescriptor(aetracoretypes.ZoneDescriptor{})
	require.ErrorContains(t, err, "disabled")
	err = app.SchedulerKeeper.RegisterScheduledJob(schedulertypes.MsgRegisterScheduledJob{
		Authority:	"4:0000000000000000000000000000000000000000000000000000000000000001",
		Job: schedulertypes.ScheduledJob{
			ID:			"disabled",
			OwnerModule:		"aetracore",
			Type:			schedulertypes.JobTypeDelayed,
			NextExecutionHeight:	1,
			MaxGas:			1,
		},
	})
	require.ErrorContains(t, err, "disabled")
}

func TestPrototypeGenesisInitializesRuntimeKeeperState(t *testing.T) {
	app, _ := setup(true, 5)
	genesis := GenesisStateWithSingleValidator(t, app)
	routingGenesis := routingkeeper.DefaultGenesis()
	routingGenesis.Shards = []routingkeeper.ShardConfig{{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 2}}
	rawRoutingGenesis, err := json.Marshal(routingGenesis)
	require.NoError(t, err)
	genesis[routingtypes.ModuleName] = rawRoutingGenesis
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)

	shards, _, err := app.RoutingKeeper.Shards(nil)
	require.NoError(t, err)
	require.Equal(t, []routingkeeper.ShardConfig{{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 2}}, shards)
}

func TestAetraCorePrototypeStateSurvivesRestartWhenDisabled(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = source.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.NoError(t, err)
	_, err = source.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	source.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = source.Commit()
	require.NoError(t, err)

	sourceCtx := source.NewUncachedContext(false, cmtproto.Header{Height: source.LastBlockHeight()})
	sourceLoad, err := source.LoadKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourceRouting, err := source.RoutingKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourceZones, err := source.ZonesKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourceMesh, err := source.MeshKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourceNetworking, err := source.NetworkingKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourcePayments, err := source.PaymentsKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourceAetraCore, err := source.AetraCoreKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)
	sourceScheduler, err := source.SchedulerKeeper.ExportGenesisState(sourceCtx)
	require.NoError(t, err)

	restarted := NewL1App(log.NewNopLogger(), db, true, appOptions)
	restartedCtx := restarted.NewUncachedContext(false, cmtproto.Header{Height: restarted.LastBlockHeight()})
	restartedLoad, err := restarted.LoadKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedRouting, err := restarted.RoutingKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedZones, err := restarted.ZonesKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedMesh, err := restarted.MeshKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedNetworking, err := restarted.NetworkingKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedPayments, err := restarted.PaymentsKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedAetraCore, err := restarted.AetraCoreKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)
	restartedScheduler, err := restarted.SchedulerKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)

	require.Equal(t, sourceAetraCore, restartedAetraCore)
	require.Equal(t, sourceLoad, restartedLoad)
	require.Equal(t, sourceRouting, restartedRouting)
	require.Equal(t, sourceZones, restartedZones)
	require.Equal(t, sourceMesh, restartedMesh)
	require.Equal(t, sourceNetworking, restartedNetworking)
	require.Equal(t, sourcePayments, restartedPayments)
	require.Equal(t, sourceScheduler, restartedScheduler)
}

func decodeJSONGenesis[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var out T
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}
