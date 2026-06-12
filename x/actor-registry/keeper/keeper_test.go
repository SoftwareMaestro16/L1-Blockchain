package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/actor-registry/types"
	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisIsDisabledAndValid(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Actors)
}

func TestRegisterActor(t *testing.T) {
	k := enabledKeeper(t, codeHash("code-a"))
	actor, err := k.RegisterActor(registerMsg("owner-a", codeHash("code-a"), "salt-a", 10))
	require.NoError(t, err)
	require.Equal(t, types.DeriveActorID("owner-a", codeHash("code-a"), "salt-a"), actor.ActorID)
	require.Equal(t, types.DeriveContractAddress(actor.ActorID), actor.ContractAddress)
	require.Equal(t, types.ActorStatusActive, actor.Status)
	require.Equal(t, uint64(1), actor.LogicalTime)

	stored, found, err := k.Actor(actor.ActorID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, actor, stored)
	byOwner, err := k.ActorsByOwner("owner-a")
	require.NoError(t, err)
	require.Equal(t, []types.ActorRecord{actor}, byOwner)
	byCode, err := k.ActorsByCodeHash(codeHash("code-a"))
	require.NoError(t, err)
	require.Equal(t, []types.ActorRecord{actor}, byCode)
}

func TestDuplicateActorRejected(t *testing.T) {
	k := enabledKeeper(t, codeHash("code-a"))
	msg := registerMsg("owner-a", codeHash("code-a"), "salt-a", 10)
	_, err := k.RegisterActor(msg)
	require.NoError(t, err)
	_, err = k.RegisterActor(msg)
	require.ErrorContains(t, err, "already registered")
}

func TestCodeHashMustExistInAVMCodeStore(t *testing.T) {
	k := enabledKeeper(t, codeHash("code-a"))
	_, err := k.RegisterActor(registerMsg("owner-a", codeHash("missing"), "salt-a", 10))
	require.ErrorContains(t, err, "code hash")
}

func TestFreezeUnfreezeLifecycle(t *testing.T) {
	k := enabledKeeper(t, codeHash("code-a"))
	actor, err := k.RegisterActor(registerMsg("owner-a", codeHash("code-a"), "salt-a", 10))
	require.NoError(t, err)

	frozen, err := k.FreezeActor(types.MsgFreezeActor{Authority: prototype.DefaultAuthority, ActorID: actor.ActorID, Height: 11})
	require.NoError(t, err)
	require.Equal(t, types.ActorStatusFrozen, frozen.Status)
	require.False(t, types.CanExecuteNormalMessage(frozen))

	active, err := k.UnfreezeActor(types.MsgUnfreezeActor{Authority: prototype.DefaultAuthority, ActorID: actor.ActorID, Height: 12})
	require.NoError(t, err)
	require.Equal(t, types.ActorStatusActive, active.Status)
	require.True(t, types.CanExecuteNormalMessage(active))
	require.Greater(t, active.LogicalTime, frozen.LogicalTime)
}

func TestDeleteLifecycle(t *testing.T) {
	k := enabledKeeper(t, codeHash("code-a"))
	actor, err := k.RegisterActor(registerMsg("owner-a", codeHash("code-a"), "salt-a", 10))
	require.NoError(t, err)

	deleted, err := k.DeleteActor(types.MsgDeleteActor{Authority: prototype.DefaultAuthority, ActorID: actor.ActorID, Height: 13})
	require.NoError(t, err)
	require.Equal(t, types.ActorStatusDeleted, deleted.Status)
	require.False(t, types.CanReceiveValue(deleted, k.ExportGenesis().RegistryParams))
	_, err = k.FreezeActor(types.MsgFreezeActor{Authority: prototype.DefaultAuthority, ActorID: actor.ActorID, Height: 14})
	require.ErrorContains(t, err, "active")
}

func TestMigrationUpdatesCodeHashAndLogicalTime(t *testing.T) {
	k := enabledKeeper(t, codeHash("code-a"), codeHash("code-b"))
	actor, err := k.RegisterActor(registerMsg("owner-a", codeHash("code-a"), "salt-a", 10))
	require.NoError(t, err)

	migrated, err := k.MigrateActor(types.MsgMigrateActor{
		Authority:	prototype.DefaultAuthority,
		ActorID:	actor.ActorID,
		NewCodeHash:	codeHash("code-b"),
		Height:		20,
		LogicalTime:	9,
	})
	require.NoError(t, err)
	require.Equal(t, types.ActorStatusMigrated, migrated.Status)
	require.Equal(t, codeHash("code-b"), migrated.CodeHash)
	require.Equal(t, uint64(9), migrated.LogicalTime)
	require.Equal(t, actor.ActorID, migrated.MigratedFrom)

	_, err = k.UpdateActorCode(types.MsgUpdateActorCode{
		Authority:	prototype.DefaultAuthority,
		ActorID:	migrated.ActorID,
		CodeHash:	codeHash("code-a"),
		Height:		21,
		LogicalTime:	8,
	})
	require.ErrorContains(t, err, "monotonically")
}

func TestExportImportPreservesLogicalTime(t *testing.T) {
	source := enabledKeeper(t, codeHash("code-a"), codeHash("code-b"))
	actor, err := source.RegisterActor(registerMsg("owner-a", codeHash("code-a"), "salt-a", 10))
	require.NoError(t, err)
	updated, err := source.UpdateActorCode(types.MsgUpdateActorCode{Authority: prototype.DefaultAuthority, ActorID: actor.ActorID, CodeHash: codeHash("code-b"), Height: 20})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	restored, found, err := target.Actor(updated.ActorID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, updated.LogicalTime, restored.LogicalTime)
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.State.CodeStore = append(gs.State.CodeStore, types.CodeRecord{CodeHash: codeHash("code-a"), RegisteredBy: prototype.DefaultAuthority, RegisteredAt: 1})
	require.NoError(t, source.InitGenesisState(ctx, gs))

	actor, err := source.RegisterActor(registerMsg("owner-a", codeHash("code-a"), "salt-a", 10))
	require.NoError(t, err)
	frozen, err := source.FreezeActor(types.MsgFreezeActor{Authority: prototype.DefaultAuthority, ActorID: actor.ActorID, Height: 11})
	require.NoError(t, err)
	require.Equal(t, types.ActorStatusFrozen, frozen.Status)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Actors, 1)
	require.Equal(t, types.ActorStatusFrozen, exported.State.Actors[0].Status)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	restored, found, err := imported.Actor(actor.ActorID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, frozen, restored)
}

func enabledKeeper(t *testing.T, codeHashes ...string) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	for _, hash := range codeHashes {
		gs.State.CodeStore = append(gs.State.CodeStore, types.CodeRecord{CodeHash: hash, RegisteredBy: prototype.DefaultAuthority, RegisteredAt: 1})
	}
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func registerMsg(owner, codeHashValue, salt string, height uint64) types.MsgRegisterActor {
	return types.MsgRegisterActor{
		Authority:	prototype.DefaultAuthority,
		Owner:		owner,
		CodeHash:	codeHashValue,
		Salt:		salt,
		Height:		height,
		Capabilities:	[]string{"call", "storage"},
	}
}

func codeHash(seed string) string {
	return types.DefaultRoot(seed)
}
