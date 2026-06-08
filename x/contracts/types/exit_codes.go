package types

const (
	ExitCodeOK uint32 = iota
	ExitCodeValidationFailed
	ExitCodeUnauthorized
	ExitCodeAccountInactive
	ExitCodeAccountFrozen
	ExitCodeContractFrozen
	ExitCodeCodeRejected
	ExitCodeOutOfGas
	ExitCodeStorageLimit
	ExitCodeStorageRentDebt
	ExitCodeMessageExpired
	ExitCodeQueueLimit
	ExitCodeExecutionFailed
	ExitCodeInternalBounce
)

func ExitCodeName(code uint32) string {
	switch code {
	case ExitCodeOK:
		return "ok"
	case ExitCodeValidationFailed:
		return "validation_failed"
	case ExitCodeUnauthorized:
		return "unauthorized"
	case ExitCodeAccountInactive:
		return "account_inactive"
	case ExitCodeAccountFrozen:
		return "account_frozen"
	case ExitCodeContractFrozen:
		return "contract_frozen"
	case ExitCodeCodeRejected:
		return "code_rejected"
	case ExitCodeOutOfGas:
		return "out_of_gas"
	case ExitCodeStorageLimit:
		return "storage_limit"
	case ExitCodeStorageRentDebt:
		return "storage_rent_debt"
	case ExitCodeMessageExpired:
		return "message_expired"
	case ExitCodeQueueLimit:
		return "queue_limit"
	case ExitCodeExecutionFailed:
		return "execution_failed"
	case ExitCodeInternalBounce:
		return "internal_bounce"
	default:
		return "unknown"
	}
}
