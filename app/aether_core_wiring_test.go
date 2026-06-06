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

	aethercorekeeper "github.com/sovereign-l1/l1/x/aethercore/keeper"
	aethercoretypes "github.com/sovereign-l1/l1/x/aethercore/types"
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
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAetherCoreWiringGateRegistersPrototypeModulesDisabled(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.ValidateAetherCoreWiringGate())
	require.Equal(t, RoutingExecutionPointAnteAdmissionOnly, AetherCoreRoutingExecutionPoint())

	for _, moduleName := range AetherCorePrototypeModuleNames() {
		require.Contains(t, app.ModuleManager.Modules, moduleName)
		require.Contains(t, app.keys, moduleName)
		require.Contains(t, genesis, moduleName)
		require.NotContains(t, GetMaccPerms(), moduleName)
		_, hasBegin := app.ModuleManager.Modules[moduleName].(appmodule.HasBeginBlocker)
		_, hasEnd := app.ModuleManager.Modules[moduleName].(appmodule.HasEndBlocker)
		require.False(t, hasBegin, moduleName)
		require.False(t, hasEnd, moduleName)
		require.True(t, slices.Contains(app.ModuleManager.OrderBeginBlockers, moduleName), moduleName)
		require.True(t, slices.Contains(app.ModuleManager.OrderEndBlockers, moduleName), moduleName)
	}

	aetherCoreGenesis := decodeJSONGenesis[aethercorekeeper.GenesisState](t, genesis[aethercoretypes.ModuleName])
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
	err = app.AetherCoreKeeper.RegisterZoneDescriptor(aethercoretypes.ZoneDescriptor{})
	require.ErrorContains(t, err, "disabled")
}

func TestAetherCorePrototypeStateSurvivesRestartWhenDisabled(t *testing.T) {
	db := dbm.NewMemDB()
	appOptions := sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome}
	source := NewL1App(log.NewNopLogger(), db, true, appOptions)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = source.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = source.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   source.LastCommitID().Hash,
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
	sourceAetherCore, err := source.AetherCoreKeeper.ExportGenesisState(sourceCtx)
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
	restartedAetherCore, err := restarted.AetherCoreKeeper.ExportGenesisState(restartedCtx)
	require.NoError(t, err)

	require.Equal(t, sourceAetherCore, restartedAetherCore)
	require.Equal(t, sourceLoad, restartedLoad)
	require.Equal(t, sourceRouting, restartedRouting)
	require.Equal(t, sourceZones, restartedZones)
	require.Equal(t, sourceMesh, restartedMesh)
	require.Equal(t, sourceNetworking, restartedNetworking)
	require.Equal(t, sourcePayments, restartedPayments)
}

func decodeJSONGenesis[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var out T
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}
