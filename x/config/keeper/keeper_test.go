package keeper

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestGenesisRejectsMalformedAndNondeterministicState(t *testing.T) {
	keeper := NewKeeper()

	bad := DefaultGenesis()
	bad.Version = 99
	require.ErrorContains(t, keeper.InitGenesis(bad), "unsupported")

	bad = DefaultGenesis()
	bad.Params.Authority = "4:0000000000000000000000000000000000000000000000000000000000000000"
	require.ErrorContains(t, keeper.InitGenesis(bad), "zero address")

	bad = DefaultGenesis()
	bad.State.Entries = []types.ConfigEntry{
		entry("runtime/z", "one", 1),
		entry("runtime/a", "two", 1),
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "sorted")

	bad.State.Entries = []types.ConfigEntry{
		entry("runtime/a", "one", 1),
		entry("runtime/a", "two", 1),
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "duplicated")

	bad = DefaultGenesis()
	bad.Params.MaxPendingChanges = 0
	require.ErrorContains(t, keeper.InitGenesis(bad), "max pending")

	bad = DefaultGenesis()
	bad.State.PendingChanges = []types.ConfigChange{
		change("z", "runtime/z", "1"),
		change("a", "runtime/a", "1"),
	}
	require.ErrorContains(t, keeper.InitGenesis(bad), "pending changes must be sorted")
}

func TestUpsertRequiresAuthorityAndRejectsUnsafeFields(t *testing.T) {
	keeper := NewKeeper()

	_, err := keeper.UpsertEntry("4:0000000000000000000000000000000000000000000000000000000000000002", "runtime/max_validators", "100", 1)
	require.ErrorContains(t, err, "governance authority")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, " runtime/max_validators", "100", 1)
	require.ErrorContains(t, err, "canonical")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/max_validators", strings.Repeat("x", int(types.MaxConfigValueBytesV1)+1), 1)
	require.ErrorContains(t, err, "value exceeds")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/max_validators", "100", -1)
	require.ErrorContains(t, err, "height")
}

func TestUpsertMaintainsSortedEntriesAndVersions(t *testing.T) {
	keeper := NewKeeper()

	_, err := keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/z", "z", 1)
	require.NoError(t, err)
	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a", 2)
	require.NoError(t, err)
	updated, err := keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a2", 3)
	require.NoError(t, err)
	require.Equal(t, uint64(2), updated.Version)

	entries, err := keeper.Entries()
	require.NoError(t, err)
	require.Equal(t, []string{"runtime/a", "runtime/z"}, []string{entries[0].Key, entries[1].Key})

	found, ok, err := keeper.Entry("runtime/a")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "a2", found.Value)
	require.Equal(t, int64(3), found.UpdatedHeight)
}

func TestUpdateParamsAndEntryLimit(t *testing.T) {
	keeper := NewKeeper()
	params := types.DefaultParams()
	params.MaxEntries = 1
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, params))

	_, err := keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a", 1)
	require.NoError(t, err)
	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, "runtime/b", "b", 2)
	require.ErrorContains(t, err, "limit")
}

func TestExportImportDeterministicAndMigration(t *testing.T) {
	source := NewKeeper()
	_, err := source.UpsertEntry(prototype.DefaultAuthority, "runtime/b", "b", 2)
	require.NoError(t, err)
	_, err = source.UpsertEntry(prototype.DefaultAuthority, "runtime/a", "a", 1)
	require.NoError(t, err)
	_, err = source.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("change-b", "runtime/b", "b2"),
	}, 3)
	require.NoError(t, err)
	_, err = source.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("change-a", "runtime/a", "a2"),
	}, 4)
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	require.Equal(t, []string{"runtime/a", "runtime/b"}, []string{exported.State.Entries[0].Key, exported.State.Entries[1].Key})
	require.Equal(t, []string{"change-a", "change-b"}, []string{exported.State.PendingChanges[0].ID, exported.State.PendingChanges[1].ID})

	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	require.NoError(t, source.InitGenesisState(ctx, DefaultGenesis()))

	_, err := source.UpsertEntry(prototype.DefaultAuthority, "runtime/persistent", "enabled", 1)
	require.NoError(t, err)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Entries, 1)
	require.Equal(t, "runtime/persistent", exported.State.Entries[0].Key)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	value, found, err := imported.ConfigValue("runtime/persistent")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "enabled", value)
}

func TestConfigChangeLifecycleRequiresAuthority(t *testing.T) {
	keeper := NewKeeper()
	_, err := keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	"4:0000000000000000000000000000000000000000000000000000000000000002",
		Change:		change("change-1", "runtime/max_validators", "100"),
	}, 1)
	require.ErrorContains(t, err, "governance authority")

	submitted, err := keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("change-1", "runtime/max_validators", "100"),
	}, 1)
	require.NoError(t, err)
	require.Equal(t, types.ChangeStatusPending, submitted.Status)

	_, _, err = keeper.ExecuteConfigChange(types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "change-1"}, 2)
	require.ErrorContains(t, err, "approved")

	approved, err := keeper.ApproveConfigChange(types.MsgApproveConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "change-1"}, 2)
	require.NoError(t, err)
	require.Equal(t, types.ChangeStatusApproved, approved.Status)

	_, _, err = keeper.ExecuteConfigChange(types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "change-1"}, approved.ActivationHeight-1)
	require.ErrorContains(t, err, "activation height")

	entry, executed, err := keeper.ExecuteConfigChange(types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "change-1"}, approved.ActivationHeight)
	require.NoError(t, err)
	require.Equal(t, types.ChangeStatusExecuted, executed.Status)
	require.Equal(t, "runtime/max_validators", entry.Key)
	require.Equal(t, "100", entry.Value)

	value, found, err := keeper.ConfigValue("runtime/max_validators")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "100", value)
}

func TestCriticalConfigChangeActivatesOnDeterministicEpoch(t *testing.T) {
	keeper := NewKeeper()
	submitted, err := keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("critical-gas", types.KeyConsensusMaxBlockGas, "1000000"),
	}, 7)
	require.NoError(t, err)
	require.True(t, submitted.Critical)
	require.Equal(t, int64(150), submitted.ActivationHeight)
	require.Equal(t, uint64(3), submitted.ActivationEpoch)

	approved, err := keeper.ApproveConfigChange(types.MsgApproveConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "critical-gas"}, 8)
	require.NoError(t, err)
	_, _, err = keeper.ExecuteConfigChange(types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "critical-gas"}, approved.ActivationHeight-1)
	require.ErrorContains(t, err, "activation height")
	_, executed, err := keeper.ExecuteConfigChange(types.MsgExecuteConfigChange{Authority: prototype.DefaultAuthority, ChangeID: "critical-gas"}, approved.ActivationHeight)
	require.NoError(t, err)
	require.Equal(t, types.ChangeStatusExecuted, executed.Status)
}

func TestInvalidConfigChangeRejectedBeforeExecution(t *testing.T) {
	keeper := NewKeeper()

	_, err := keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("bad-gas", "avm/gas/contract_call", "0"),
	}, 1)
	require.ErrorContains(t, err, "positive")

	_, err = keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("bad-denom", types.KeyFeeBaseDenom, "uatom"),
	}, 1)
	require.ErrorContains(t, err, "base denom")
}

func TestConfigCannotSetUnlimitedBlockGas(t *testing.T) {
	keeper := NewKeeper()
	_, err := keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("bad-block-gas", types.KeyConsensusMaxBlockGas, "1000000001"),
	}, 1)
	require.ErrorContains(t, err, "unlimited block gas")
}

func TestConfigCannotSetZeroStorageRentForNonEmptyStateWithoutConstitutionalRule(t *testing.T) {
	keeper := NewKeeper()
	_, err := keeper.UpsertEntry(prototype.DefaultAuthority, types.KeyStorageContractStateActive, "true", 1)
	require.NoError(t, err)

	_, err = keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		change("zero-rent", types.KeyStorageRentPerByteEpoch, "0"),
	}, 2)
	require.ErrorContains(t, err, "constitutional allowance")

	_, err = keeper.UpsertEntry(prototype.DefaultAuthority, types.KeyConstitutionZeroRentAllow, "true", 2)
	require.NoError(t, err)
	allowed := change("zero-rent", types.KeyStorageRentPerByteEpoch, "0")
	allowed.RequiresConstitutionalException = true
	_, err = keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		allowed,
	}, 3)
	require.NoError(t, err)
}

func TestConfigCannotRemoveRequiredSystemAccountAddresses(t *testing.T) {
	keeper := NewKeeper()
	_, err := keeper.UpsertEntry(prototype.DefaultAuthority, "system/account/fee_collector", prototype.DefaultAuthority, 1)
	require.NoError(t, err)

	deleteChange := change("remove-fee-collector", "system/account/fee_collector", "")
	deleteChange.Operation = types.OperationDelete
	_, err = keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
		Authority:	prototype.DefaultAuthority,
		Change:		deleteChange,
	}, 2)
	require.ErrorContains(t, err, "required system account")
}

func TestPendingConfigChangesAreDeterministicallyOrdered(t *testing.T) {
	keeper := NewKeeper()
	for _, id := range []string{"change-z", "change-a", "change-m"} {
		_, err := keeper.SubmitConfigChange(types.MsgSubmitConfigChange{
			Authority:	prototype.DefaultAuthority,
			Change:		change(id, "runtime/"+id, "1"),
		}, 1)
		require.NoError(t, err)
	}

	pending, err := keeper.PendingConfigChanges()
	require.NoError(t, err)
	require.Equal(t, []string{"change-a", "change-m", "change-z"}, []string{pending[0].ID, pending[1].ID, pending[2].ID})
}

func entry(key string, value string, version uint64) types.ConfigEntry {
	return types.ConfigEntry{
		Key:		key,
		Value:		value,
		Owner:		prototype.DefaultAuthority,
		Version:	version,
		UpdatedHeight:	1,
	}
}

func change(id string, key string, value string) types.ConfigChange {
	return types.ConfigChange{
		ID:				id,
		Key:				key,
		Value:				value,
		Operation:			types.OperationSet,
		Status:				types.ChangeStatusPending,
		SubmittedBy:			prototype.DefaultAuthority,
		CreatedHeight:			1,
		UpdatedHeight:			1,
		ExpectedPreviousVersion:	0,
	}
}
