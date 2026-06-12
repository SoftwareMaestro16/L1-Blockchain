package keeper

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/system-registry/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestGenesisRejectsDuplicateSystemAccounts(t *testing.T) {
	gs := DefaultGenesis()
	config, found := gs.State.Entity(types.ModuleConfig)
	require.True(t, found)
	gs.State.Entities = append(gs.State.Entities, types.SystemEntity{
		ModuleName:		"duplicate-config-account",
		ModuleAccountAddress:	config.ModuleAccountAddress,
		AuthorityAddress:	prototype.DefaultAuthority,
		Status:			types.StatusActive,
		Version:		1,
	})

	require.ErrorContains(t, gs.Validate(), "duplicate module account")
}

func TestRequiredModulesCannotBeRemovedOrPaused(t *testing.T) {
	k := NewKeeper()

	_, _, err := k.PauseSystemEntity(types.MsgPauseSystemEntity{
		Authority:	prototype.DefaultAuthority,
		ModuleName:	types.ModuleConfig,
		Height:		1,
	})
	require.ErrorContains(t, err, "required module")

	config, found, err := k.SystemEntity(types.ModuleConfig)
	require.NoError(t, err)
	require.True(t, found)
	config.Status = types.StatusDeprecated
	_, _, err = k.UpdateSystemEntity(types.MsgUpdateSystemEntity{
		Authority:	prototype.DefaultAuthority,
		Entity:		config,
	})
	require.ErrorContains(t, err, "required module")

	gs := k.ExportGenesis()
	gs.Params.RequiredModules = []string{types.ModuleConfig, types.ModuleConstitution, "missing", types.ModuleName}
	require.ErrorContains(t, gs.Validate(), "required module")
}

func TestPauseResumeEventsAreDeterministic(t *testing.T) {
	first := NewKeeper()
	second := NewKeeper()
	for _, k := range []*Keeper{&first, &second} {
		_, _, err := k.RegisterSystemEntity(types.MsgRegisterSystemEntity{
			Authority:	prototype.DefaultAuthority,
			Entity: types.SystemEntity{
				ModuleName:		"state-metering",
				ModuleAccountAddress:	testAddress(0x55),
				AuthorityAddress:	prototype.DefaultAuthority,
				Status:			types.StatusActive,
				Capabilities:		[]string{"rent-collection", "state-metering"},
				Version:		3,
				Dependencies:		[]string{types.ModuleConstitution},
			},
		})
		require.NoError(t, err)
	}

	pausedFirst, pauseEventFirst, err := first.PauseSystemEntity(types.MsgPauseSystemEntity{
		Authority:				prototype.DefaultAuthority,
		ModuleName:				"state-metering",
		Height:					10,
		AllowPrivilegedCallsWhilePaused:	false,
	})
	require.NoError(t, err)
	pausedSecond, pauseEventSecond, err := second.PauseSystemEntity(types.MsgPauseSystemEntity{
		Authority:				prototype.DefaultAuthority,
		ModuleName:				"state-metering",
		Height:					10,
		AllowPrivilegedCallsWhilePaused:	false,
	})
	require.NoError(t, err)
	require.Equal(t, pausedFirst, pausedSecond)
	require.Equal(t, pauseEventFirst, pauseEventSecond)
	require.Equal(t, types.EventTypePaused, pauseEventFirst.Type)
	allowed, err := first.CanReceivePrivilegedCall("state-metering")
	require.NoError(t, err)
	require.False(t, allowed)

	resumedFirst, resumeEventFirst, err := first.ResumeSystemEntity(types.MsgResumeSystemEntity{
		Authority:	prototype.DefaultAuthority,
		ModuleName:	"state-metering",
		Height:		11,
	})
	require.NoError(t, err)
	resumedSecond, resumeEventSecond, err := second.ResumeSystemEntity(types.MsgResumeSystemEntity{
		Authority:	prototype.DefaultAuthority,
		ModuleName:	"state-metering",
		Height:		11,
	})
	require.NoError(t, err)
	require.Equal(t, resumedFirst, resumedSecond)
	require.Equal(t, resumeEventFirst, resumeEventSecond)
	require.Equal(t, types.EventTypeResumed, resumeEventFirst.Type)
	allowed, err = first.CanReceivePrivilegedCall("state-metering")
	require.NoError(t, err)
	require.True(t, allowed)
}

func TestPausedModulePrivilegedCallsRequireExplicitAllowance(t *testing.T) {
	k := NewKeeper()
	_, _, err := k.RegisterSystemEntity(types.MsgRegisterSystemEntity{
		Authority:	prototype.DefaultAuthority,
		Entity: types.SystemEntity{
			ModuleName:		"latency-oracle",
			ModuleAccountAddress:	testAddress(0x56),
			AuthorityAddress:	prototype.DefaultAuthority,
			Status:			types.StatusActive,
			Version:		1,
			Dependencies:		[]string{types.ModuleConstitution},
		},
	})
	require.NoError(t, err)

	_, _, err = k.PauseSystemEntity(types.MsgPauseSystemEntity{
		Authority:				prototype.DefaultAuthority,
		ModuleName:				"latency-oracle",
		Height:					7,
		AllowPrivilegedCallsWhilePaused:	true,
	})
	require.NoError(t, err)
	allowed, err := k.CanReceivePrivilegedCall("latency-oracle")
	require.NoError(t, err)
	require.True(t, allowed)
}

func TestDependencyGraphCycleIsRejected(t *testing.T) {
	gs := DefaultGenesis()
	gs.State.Entities = append(gs.State.Entities,
		types.SystemEntity{
			ModuleName:		"cycle-a",
			ModuleAccountAddress:	testAddress(0x60),
			AuthorityAddress:	prototype.DefaultAuthority,
			Status:			types.StatusActive,
			Version:		1,
			Dependencies:		[]string{"cycle-b"},
		},
		types.SystemEntity{
			ModuleName:		"cycle-b",
			ModuleAccountAddress:	testAddress(0x61),
			AuthorityAddress:	prototype.DefaultAuthority,
			Status:			types.StatusActive,
			Version:		1,
			Dependencies:		[]string{"cycle-a"},
		},
	)

	require.ErrorContains(t, gs.Validate(), "cycle")
}

func TestExportImportPreservesRegistryOrdering(t *testing.T) {
	source := NewKeeper()
	for _, moduleName := range []string{"custom-election-audit", "custom-emissions-audit"} {
		_, _, err := source.RegisterSystemEntity(types.MsgRegisterSystemEntity{
			Authority:	prototype.DefaultAuthority,
			Entity: types.SystemEntity{
				ModuleName:		moduleName,
				ModuleAccountAddress:	testAddress(byte(len(moduleName))),
				AuthorityAddress:	prototype.DefaultAuthority,
				Status:			types.StatusActive,
				Capabilities:		[]string{"z", "a"},
				Version:		1,
				Dependencies:		[]string{types.ModuleName, types.ModuleConstitution},
			},
		})
		require.NoError(t, err)
	}

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	entities, err := source.SystemEntities()
	require.NoError(t, err)
	require.LessOrEqual(t, entities[0].ModuleName, entities[len(entities)-1].ModuleName)
	emissions, found, err := source.SystemEntity("custom-emissions-audit")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []string{"a", "z"}, emissions.Capabilities)
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())

	graph, err := target.DependencyGraph()
	require.NoError(t, err)
	require.NotEmpty(t, graph)
	require.LessOrEqual(t, graph[0].ModuleName, graph[len(graph)-1].ModuleName)
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	require.NoError(t, source.InitGenesisState(ctx, DefaultGenesis()))

	_, _, err := source.RegisterSystemEntity(types.MsgRegisterSystemEntity{
		Authority:	prototype.DefaultAuthority,
		Entity: types.SystemEntity{
			ModuleName:		"runtime-metering",
			ModuleAccountAddress:	testAddress(0x72),
			AuthorityAddress:	prototype.DefaultAuthority,
			Status:			types.StatusActive,
			Version:		1,
		},
	})
	require.NoError(t, err)
	_, _, err = source.PauseSystemEntity(types.MsgPauseSystemEntity{Authority: prototype.DefaultAuthority, ModuleName: "runtime-metering", Height: 2})
	require.NoError(t, err)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	entity, found := exported.State.Entity("runtime-metering")
	require.True(t, found)
	require.Equal(t, types.StatusPaused, entity.Status)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	allowed, err := imported.CanReceivePrivilegedCall("runtime-metering")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestMaliciousAuthorityCannotUpdateRegistry(t *testing.T) {
	k := NewKeeper()
	_, _, err := k.RegisterSystemEntity(types.MsgRegisterSystemEntity{
		Authority:	"4:0000000000000000000000000000000000000000000000000000000000000002",
		Entity: types.SystemEntity{
			ModuleName:		"malicious",
			ModuleAccountAddress:	testAddress(0x70),
			AuthorityAddress:	prototype.DefaultAuthority,
			Status:			types.StatusActive,
			Version:		1,
		},
	})
	require.ErrorContains(t, err, "governance authority")
}

func testAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(bytesOf(fill)))
}

func bytesOf(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
