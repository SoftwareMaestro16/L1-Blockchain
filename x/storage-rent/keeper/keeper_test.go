package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/storage-rent/types"
)

const (
	authority	= prototype.DefaultAuthority
	contract	= "contract-1"
	actor		= "actor-1"
)

func setupKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.RentParams.RentRatePerByteBlock = 1
	gs.RentParams.FreeStorageAllowance = 0
	gs.RentParams.RetentionBlocks = 5
	gs.RentParams.UnfreezeBufferBlocks = 2
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func TestDefaultGenesisDisabled(t *testing.T) {
	gs := DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
}

func TestContractWithPaidRentRemainsActive(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)
	ctx := context.Background()
	_, distribution, err := k.PayStorageRent(ctx, types.MsgPayStorageRent{Payer: "payer", ContractAddress: contract, Amount: 100, Height: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(100), distribution.Amount)

	_, err = k.FreezeExpiredContract(types.MsgFreezeExpiredContract{Authority: authority, ContractAddress: contract, Height: 5})
	require.ErrorContains(t, err, "not expired")
	record, found, err := k.ContractRent(contract)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, types.ContractStatusActive, record.Status)
	require.True(t, types.CanExecuteContract(record))
}

func TestUnpaidRentFreezesAndRejectsExecution(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)

	record, err := k.FreezeExpiredContract(types.MsgFreezeExpiredContract{Authority: authority, ContractAddress: contract, Height: 3})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusFrozen, record.Status)
	require.Equal(t, uint64(20), record.RentDebt)
	require.False(t, types.CanExecuteContract(record))

	frozen, err := k.FrozenContracts()
	require.NoError(t, err)
	require.Len(t, frozen, 1)
}

func TestPayingDebtUnfreezesWithBuffer(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)
	_, err = k.FreezeExpiredContract(types.MsgFreezeExpiredContract{Authority: authority, ContractAddress: contract, Height: 3})
	require.NoError(t, err)

	ctx := context.Background()
	_, _, err = k.UnfreezeContract(ctx, types.MsgUnfreezeContract{Payer: "payer", ContractAddress: contract, Amount: 49, Height: 4})
	require.ErrorContains(t, err, "full debt plus configured buffer")
	record, distribution, err := k.UnfreezeContract(ctx, types.MsgUnfreezeContract{Payer: "payer", ContractAddress: contract, Amount: 50, Height: 4})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusActive, record.Status)
	require.Zero(t, record.RentDebt)
	require.Equal(t, uint64(20), record.PrepaidRentBalance)
	require.Equal(t, uint64(50), distribution.Amount)
}

func TestDeleteAfterRetentionPeriodOnly(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)
	frozen, err := k.FreezeExpiredContract(types.MsgFreezeExpiredContract{Authority: authority, ContractAddress: contract, Height: 3})
	require.NoError(t, err)
	require.Equal(t, uint64(8), frozen.DeletionEligibilityHeight)

	_, err = k.DeleteExpiredContract(types.MsgDeleteExpiredContract{Authority: authority, ContractAddress: contract, Height: 7})
	require.ErrorContains(t, err, "before retention")
	deleted, err := k.DeleteExpiredContract(types.MsgDeleteExpiredContract{Authority: authority, ContractAddress: contract, Height: 8})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusDeleted, deleted.Status)
}

func TestStorageUsageAccountingExactness(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 9, 1)
	require.ErrorContains(t, err, "must match")
}

func TestRentDistributionConservesConfiguredProportions(t *testing.T) {
	k := setupKeeper(t)
	ctx := context.Background()
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)

	_, distribution, err := k.PayStorageRent(ctx, types.MsgPayStorageRent{Payer: "payer", ContractAddress: contract, Amount: 100, Height: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(50), distribution.FeeCollectorAmount)
	require.Equal(t, uint64(40), distribution.TreasuryAmount)
	require.Equal(t, uint64(10), distribution.BurnAmount)
	require.Equal(t, distribution.Amount, distribution.FeeCollectorAmount+distribution.TreasuryAmount+distribution.BurnAmount)
}

func TestExportImportPreservesFreezeAndDeletionQueues(t *testing.T) {
	k := setupKeeper(t)
	_, err := k.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)
	_, err = k.FreezeExpiredContract(types.MsgFreezeExpiredContract{Authority: authority, ContractAddress: contract, Height: 3})
	require.NoError(t, err)

	exported := k.ExportGenesis()
	var imported Keeper
	require.NoError(t, imported.InitGenesis(exported))
	frozen, err := imported.FrozenContracts()
	require.NoError(t, err)
	deletionQueue, err := imported.DeletionQueue()
	require.NoError(t, err)
	require.Len(t, frozen, 1)
	require.Len(t, deletionQueue, 1)
	require.Equal(t, frozen[0].DeletionEligibilityHeight, deletionQueue[0].DeletionEligibilityHeight)
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.RentParams.FreeStorageAllowance = 0
	require.NoError(t, source.InitGenesisState(ctx, gs))

	_, err := source.TrackContractStorageUsage(authority, contract, actor, 10, 10, 1)
	require.NoError(t, err)
	_, distribution, err := source.PayStorageRent(ctx, types.MsgPayStorageRent{Payer: "payer", ContractAddress: contract, Amount: 100, Height: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(100), distribution.Amount)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Contracts, 1)
	require.Len(t, exported.State.Distributions, 1)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	record, found, err := imported.ContractRent(contract)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(100), record.PrepaidRentBalance)
}
