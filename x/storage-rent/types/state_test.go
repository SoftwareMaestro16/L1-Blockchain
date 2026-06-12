package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRentAccrualAndDebtNeverNegative(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 10
	contract := ContractRentRecord{
		ContractAddress:	"contract-1",
		ActorID:		"actor-1",
		StorageBytes:		20,
		PrepaidRentBalance:	200,
		LastChargedHeight:	1,
		Status:			ContractStatusActive,
	}

	next, due, err := AccrueRent(contract, params, 11)
	require.NoError(t, err)
	require.Equal(t, uint64(100), due)
	require.Equal(t, uint64(100), next.PrepaidRentBalance)
	require.Zero(t, next.RentDebt)

	next, err = ApplyRentPayment(next, 50)
	require.NoError(t, err)
	require.Zero(t, next.RentDebt)
	require.Equal(t, uint64(150), next.PrepaidRentBalance)
}

func TestFrozenOrDeletedContractCannotExecute(t *testing.T) {
	require.True(t, CanExecuteContract(ContractRentRecord{Status: ContractStatusActive}))
	require.False(t, CanExecuteContract(ContractRentRecord{Status: ContractStatusFrozen}))
	require.False(t, CanExecuteContract(ContractRentRecord{Status: ContractStatusDeleted}))
}

func TestUnfreezeRequiresDebtPlusConfiguredBuffer(t *testing.T) {
	params := DefaultStorageRentParams()
	params.UnfreezeBufferBlocks = 3
	contract := ContractRentRecord{StorageBytes: 10, RentDebt: 7}

	required, err := RequiredUnfreezePayment(contract, params)
	require.NoError(t, err)
	require.Equal(t, uint64(37), required)
}

func TestDistributionConservesCoins(t *testing.T) {
	params := DefaultStorageRentParams()
	distribution := BuildDistribution("contract-1", 10, 101, params)

	require.Equal(t, uint64(101), distribution.FeeCollectorAmount+distribution.TreasuryAmount+distribution.BurnAmount)
	require.NoError(t, distribution.Validate())
}
