package types

import (
	"errors"
	"fmt"
)

const (
	ContractLifecycleActionExecuteExternal		= "execute_external"
	ContractLifecycleActionReceiveInternal		= "receive_internal"
	ContractLifecycleActionReceiveTopUp		= "receive_top_up"
	ContractLifecycleActionPayRentDebt		= "pay_rent_debt"
	ContractLifecycleActionUnfreeze			= "unfreeze"
	ContractLifecycleActionQuery			= "query"
	ContractLifecycleActionProofQuery		= "proof_query"
	ContractLifecycleActionEmitInternalMessage	= "emit_internal_message"
	ContractLifecycleActionUpgradeMigrate		= "upgrade_migrate"
	ContractLifecycleActionArchiveDelete		= "archive_delete"

	ErrContractLifecycle	= "contracts_lifecycle"
)

// ContractLifecycleActionAllowed is the canonical AVM contract lifecycle matrix.
// Frozen contracts preserve code, data, balance and state root while allowing only
// recovery actions. Frozen-limited contracts use the same bounded recovery surface
// for official contracts: no growth or normal execution, but exit/recovery hooks can
// be added as explicit whitelisted actions without reopening normal execution.
func ContractLifecycleActionAllowed(status string, action string) bool {
	switch status {
	case ContractStatusActive:
		switch action {
		case ContractLifecycleActionExecuteExternal,
			ContractLifecycleActionReceiveInternal,
			ContractLifecycleActionReceiveTopUp,
			ContractLifecycleActionPayRentDebt,
			ContractLifecycleActionQuery,
			ContractLifecycleActionProofQuery,
			ContractLifecycleActionEmitInternalMessage,
			ContractLifecycleActionUpgradeMigrate,
			ContractLifecycleActionArchiveDelete:
			return true
		default:
			return false
		}
	case ContractStatusFrozen:
		switch action {
		case ContractLifecycleActionReceiveTopUp,
			ContractLifecycleActionPayRentDebt,
			ContractLifecycleActionUnfreeze,
			ContractLifecycleActionQuery,
			ContractLifecycleActionProofQuery:
			return true
		default:
			return false
		}
	case ContractStatusFrozenLimited:
		switch action {
		case ContractLifecycleActionReceiveTopUp,
			ContractLifecycleActionPayRentDebt,
			ContractLifecycleActionUnfreeze,
			ContractLifecycleActionQuery,
			ContractLifecycleActionProofQuery:
			return true
		default:
			return false
		}
	case ContractStatusArchived:
		switch action {
		case ContractLifecycleActionReceiveTopUp,
			ContractLifecycleActionPayRentDebt,
			ContractLifecycleActionQuery,
			ContractLifecycleActionProofQuery,
			ContractLifecycleActionArchiveDelete:
			return true
		default:
			return false
		}
	case ContractStatusDeleted:
		switch action {
		case ContractLifecycleActionQuery,
			ContractLifecycleActionProofQuery:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func EnsureContractLifecycleAction(contract Contract, action string) error {
	if ContractLifecycleActionAllowed(contract.Status, action) {
		return nil
	}
	errCode := ErrContractLifecycle
	if contract.Status == ContractStatusFrozen || contract.Status == ContractStatusFrozenLimited {
		errCode = ErrAccountFrozen
	}
	return fmt.Errorf("%s: contract status %q does not allow %s", errCode, contract.Status, action)
}

func ValidateDeletedContractTombstone(contract Contract) error {
	if contract.Status != ContractStatusDeleted {
		return nil
	}
	if contract.Balance != 0 {
		return errors.New(ErrContractLifecycle + ": deleted contract cannot retain balance")
	}
	if contract.StorageRentDebt != 0 {
		return errors.New(ErrContractLifecycle + ": deleted contract cannot retain storage rent debt")
	}
	return nil
}
