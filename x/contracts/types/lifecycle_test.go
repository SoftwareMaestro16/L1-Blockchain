package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractLifecycleActionMatrix(t *testing.T) {
	tests := []struct {
		status  string
		allowed []string
		blocked []string
	}{
		{
			status: ContractStatusActive,
			allowed: []string{
				ContractLifecycleActionExecuteExternal,
				ContractLifecycleActionReceiveInternal,
				ContractLifecycleActionReceiveTopUp,
				ContractLifecycleActionPayRentDebt,
				ContractLifecycleActionQuery,
				ContractLifecycleActionProofQuery,
				ContractLifecycleActionEmitInternalMessage,
				ContractLifecycleActionUpgradeMigrate,
				ContractLifecycleActionArchiveDelete,
			},
			blocked: []string{ContractLifecycleActionUnfreeze},
		},
		{
			status: ContractStatusFrozen,
			allowed: []string{
				ContractLifecycleActionReceiveTopUp,
				ContractLifecycleActionPayRentDebt,
				ContractLifecycleActionUnfreeze,
				ContractLifecycleActionQuery,
				ContractLifecycleActionProofQuery,
			},
			blocked: []string{
				ContractLifecycleActionExecuteExternal,
				ContractLifecycleActionReceiveInternal,
				ContractLifecycleActionEmitInternalMessage,
				ContractLifecycleActionUpgradeMigrate,
				ContractLifecycleActionArchiveDelete,
			},
		},
		{
			status: ContractStatusFrozenLimited,
			allowed: []string{
				ContractLifecycleActionReceiveTopUp,
				ContractLifecycleActionPayRentDebt,
				ContractLifecycleActionUnfreeze,
				ContractLifecycleActionQuery,
				ContractLifecycleActionProofQuery,
			},
			blocked: []string{
				ContractLifecycleActionExecuteExternal,
				ContractLifecycleActionReceiveInternal,
				ContractLifecycleActionEmitInternalMessage,
				ContractLifecycleActionUpgradeMigrate,
				ContractLifecycleActionArchiveDelete,
			},
		},
		{
			status: ContractStatusArchived,
			allowed: []string{
				ContractLifecycleActionReceiveTopUp,
				ContractLifecycleActionPayRentDebt,
				ContractLifecycleActionQuery,
				ContractLifecycleActionProofQuery,
				ContractLifecycleActionArchiveDelete,
			},
			blocked: []string{
				ContractLifecycleActionExecuteExternal,
				ContractLifecycleActionReceiveInternal,
				ContractLifecycleActionUnfreeze,
				ContractLifecycleActionEmitInternalMessage,
				ContractLifecycleActionUpgradeMigrate,
			},
		},
		{
			status:  ContractStatusDeleted,
			allowed: []string{ContractLifecycleActionQuery, ContractLifecycleActionProofQuery},
			blocked: []string{
				ContractLifecycleActionExecuteExternal,
				ContractLifecycleActionReceiveInternal,
				ContractLifecycleActionReceiveTopUp,
				ContractLifecycleActionPayRentDebt,
				ContractLifecycleActionUnfreeze,
				ContractLifecycleActionEmitInternalMessage,
				ContractLifecycleActionUpgradeMigrate,
				ContractLifecycleActionArchiveDelete,
			},
		},
	}

	for _, tc := range tests {
		for _, action := range tc.allowed {
			require.Truef(t, ContractLifecycleActionAllowed(tc.status, action), "%s should allow %s", tc.status, action)
			require.NoError(t, EnsureContractLifecycleAction(Contract{Status: tc.status}, action))
		}
		for _, action := range tc.blocked {
			require.Falsef(t, ContractLifecycleActionAllowed(tc.status, action), "%s should block %s", tc.status, action)
			require.Error(t, EnsureContractLifecycleAction(Contract{Status: tc.status}, action))
		}
	}
	require.False(t, ContractLifecycleActionAllowed("unknown", ContractLifecycleActionQuery))
}

func TestDeletedContractTombstoneRejectsRemainingBalanceOrDebt(t *testing.T) {
	contract := lifecycleContract(t, ContractStatusDeleted)
	require.NoError(t, contract.Validate(DefaultParams()))

	withBalance := contract
	withBalance.Balance = 1
	require.ErrorContains(t, withBalance.Validate(DefaultParams()), "deleted contract cannot retain balance")

	withDebt := contract
	withDebt.StorageRentDebt = 1
	require.ErrorContains(t, withDebt.Validate(DefaultParams()), "deleted contract cannot retain storage rent debt")
}

func lifecycleContract(t *testing.T, status string) Contract {
	t.Helper()
	owner := contractAPIAddress(0x41)
	raw, err := RawAddressForUserAddress(owner)
	require.NoError(t, err)
	codeHash := stateInitHash("lifecycle-code")
	contract := Contract{
		AddressUser:          owner,
		AddressRaw:           raw,
		CodeID:               codeHash,
		CodeHash:             codeHash,
		Creator:              owner,
		Owner:                owner,
		Admin:                owner,
		Status:               status,
		StorageSchemaVersion: 1,
		LogicalTime:          1,
		CreatedHeight:        1,
		UpdatedHeight:        1,
	}
	contract.StateRoot = ComputeContractStateRoot(contract)
	return contract
}
