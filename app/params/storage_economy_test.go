package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestStorageFeesAreDeterministicFromStateDelta(t *testing.T) {
	params := DefaultStorageEconomyParams()
	params.StateWriteFeePerByteNaet = sdkmath.NewInt(4)
	params.StateUpdateFeePerByteNaet = sdkmath.NewInt(1)

	write, err := ComputeStorageFee(StorageFeeInput{
		OwnerID:	"acct1",
		Class:		StorageClassAccount,
		Operation:	StorageOperationWrite,
		CurrentBytes:	100,
		DeltaBytes:	25,
		Params:		params,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(100), write.FeeNaet)
	require.True(t, write.RefundNaet.IsZero())
	require.Equal(t, int64(125), write.NewFootprintBytes)
	require.Len(t, write.Events, 1)
	require.Equal(t, StorageFeeEventWrite, write.Events[0].Type)

	update, err := ComputeStorageFee(StorageFeeInput{
		ContractID:	"contract1",
		Class:		StorageClassContract,
		Operation:	StorageOperationUpdate,
		CurrentBytes:	125,
		DeltaBytes:	-10,
		Params:		params,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(10), update.FeeNaet)
	require.Equal(t, int64(115), update.NewFootprintBytes)
	require.Len(t, update.Events, 1)
	require.Equal(t, StorageFeeEventUpdate, update.Events[0].Type)
}

func TestDeleteRefundCannotExceedOriginalCostAfterDecayAndCap(t *testing.T) {
	params := DefaultStorageEconomyParams()
	params.DeleteRefundRatioBps = 8_000
	params.DeleteRefundCapBps = 4_000
	params.DeleteRefundDecayBpsPerPeriod = 1_000

	deleted, err := ComputeStorageFee(StorageFeeInput{
		ContractID:		"contract2",
		Class:			StorageClassContract,
		Operation:		StorageOperationDelete,
		CurrentBytes:		100,
		DeletedBytes:		20,
		OriginalCostNaet:	sdkmath.NewInt(1_000),
		StorageAgePeriods:	3,
		Params:			params,
	})
	require.NoError(t, err)
	require.True(t, deleted.FeeNaet.IsZero())
	require.Equal(t, sdkmath.NewInt(400), deleted.RefundCapNaet)
	require.Equal(t, sdkmath.NewInt(400), deleted.RefundNaet)
	require.Equal(t, int64(3_000), deleted.RefundDecayBps)
	require.Equal(t, sdkmath.NewInt(-400), deleted.NetChargeNaet)
	require.Equal(t, int64(80), deleted.NewFootprintBytes)
	require.Len(t, deleted.Events, 2)
	require.Equal(t, StorageFeeEventDelete, deleted.Events[0].Type)
	require.Equal(t, StorageFeeEventRefund, deleted.Events[1].Type)
}

func TestStorageFootprintIsQueryableForAccountsAndContracts(t *testing.T) {
	query, err := QueryStorageFootprint(StorageFootprintQueryInput{
		Records: []StorageFootprintRecord{
			{OwnerID: "acct1", Class: StorageClassAccount, Bytes: 100, PrepaidBalanceNaet: sdkmath.NewInt(1_000)},
			{OwnerID: "acct1", ContractID: "contract1", Class: StorageClassContract, Bytes: 250, PrepaidBalanceNaet: sdkmath.NewInt(2_000)},
			{OwnerID: "acct2", Class: StorageClassProtocolCritical, Bytes: 500, PrepaidBalanceNaet: sdkmath.NewInt(3_000), ConsensusCritical: true},
		},
		OwnerID:	"acct1",
	})
	require.NoError(t, err)
	require.Len(t, query.Records, 2)
	require.Equal(t, int64(350), query.TotalBytes)
	require.Equal(t, int64(100), query.AccountBytes)
	require.Equal(t, int64(250), query.ContractBytes)
	require.Equal(t, sdkmath.NewInt(3_000), query.TotalPrepaidNaet)

	contract, err := QueryStorageFootprint(StorageFootprintQueryInput{
		Records:	query.Records,
		ContractID:	"contract1",
	})
	require.NoError(t, err)
	require.Len(t, contract.Records, 1)
	require.Equal(t, int64(250), contract.TotalBytes)
}

func TestStorageRentStatusWarningAndRecoveryPath(t *testing.T) {
	params := DefaultStorageEconomyParams()
	params.RentRatePerBytePeriodNaet = sdkmath.NewInt(1)
	params.RentPeriodBlocks = 10
	params.WarningPeriodsBeforeExhaustion = 2

	status, err := ComputeStorageRentStatus(StorageRentInput{
		Record: StorageFootprintRecord{
			OwnerID:		"acct1",
			Class:			StorageClassAccount,
			Bytes:			100,
			PrepaidBalanceNaet:	sdkmath.NewInt(500),
			LastRentHeight:		100,
		},
		CurrentHeight:	130,
		Params:		params,
	})
	require.NoError(t, err)
	require.Equal(t, StorageRentStatusWarning, status.Status)
	require.Equal(t, uint64(3), status.PeriodsElapsed)
	require.Equal(t, uint64(5), status.PeriodsCovered)
	require.Equal(t, uint64(2), status.PeriodsUntilExhaustion)
	require.True(t, status.CanExecute)
	require.False(t, status.Frozen)
}

func TestStorageRentExhaustionStatesAreDeterministic(t *testing.T) {
	params := DefaultStorageEconomyParams()
	params.RentRatePerBytePeriodNaet = sdkmath.NewInt(1)
	params.RentPeriodBlocks = 10
	params.FreezeGracePeriods = 1
	params.CleanupGracePeriods = 2
	record := StorageFootprintRecord{
		ContractID:		"contract1",
		Class:			StorageClassContract,
		Bytes:			100,
		PrepaidBalanceNaet:	sdkmath.NewInt(100),
		LastRentHeight:		100,
	}

	frozen, err := ComputeStorageRentStatus(StorageRentInput{Record: record, CurrentHeight: 120, Params: params})
	require.NoError(t, err)
	require.Equal(t, StorageRentStatusFrozen, frozen.Status)
	require.True(t, frozen.Frozen)
	require.False(t, frozen.CanExecute)
	require.Equal(t, sdkmath.NewInt(200), frozen.RecoveryRequiredNaet)

	limited, err := ComputeStorageRentStatus(StorageRentInput{Record: record, CurrentHeight: 130, Params: params})
	require.NoError(t, err)
	require.Equal(t, StorageRentStatusLimitedExec, limited.Status)
	require.True(t, limited.LimitedExecution)
	require.True(t, limited.CanExecute)

	cleanup, err := ComputeStorageRentStatus(StorageRentInput{Record: record, CurrentHeight: 150, Params: params})
	require.NoError(t, err)
	require.Equal(t, StorageRentStatusCleanupEligible, cleanup.Status)
	require.True(t, cleanup.CleanupEligible)
	require.False(t, cleanup.CanExecute)
}

func TestStorageRentCannotDeleteConsensusCriticalStateAccidentally(t *testing.T) {
	status, err := ComputeStorageRentStatus(StorageRentInput{
		Record: StorageFootprintRecord{
			OwnerID:		"module-account",
			Class:			StorageClassProtocolCritical,
			Bytes:			10_000,
			LastRentHeight:		1,
			ConsensusCritical:	true,
		},
		CurrentHeight:	1_000_000,
		Params:		DefaultStorageEconomyParams(),
	})
	require.NoError(t, err)
	require.Equal(t, StorageRentStatusExempt, status.Status)
	require.True(t, status.CanExecute)
	require.True(t, status.ConsensusCriticalProtected)
	require.False(t, status.CleanupEligible)

	fee, err := ComputeStorageFee(StorageFeeInput{
		OwnerID:	"module-account",
		Class:		StorageClassProtocolCritical,
		Operation:	StorageOperationDelete,
		CurrentBytes:	10_000,
		DeletedBytes:	1_000,
		Params:		DefaultStorageEconomyParams(),
	})
	require.NoError(t, err)
	require.True(t, fee.FeeNaet.IsZero())
	require.True(t, fee.RefundNaet.IsZero())
	require.Empty(t, fee.Events)
}
