package types

// ExitCodeSpec represents an exit code with its name
type ExitCodeSpec struct {
	Code	uint32
	Name	string
}

const (
	// VM Execution Errors (0-31)
	ExitCodeOK			uint32	= 0
	ExitCodeValidationFailed	uint32	= 1
	ExitCodeUnauthorized		uint32	= 2
	ExitCodeAccountInactive		uint32	= 3
	ExitCodeAccountFrozen		uint32	= 4
	ExitCodeContractFrozen		uint32	= 5
	ExitCodeCodeRejected		uint32	= 6
	ExitCodeTypeCheckFailed		uint32	= 7
	ExitCodeOutOfGas		uint32	= 8
	ExitCodeStackOverflow		uint32	= 9
	ExitCodeStackUnderflow		uint32	= 10
	ExitCodeTypeCheckError		uint32	= 11	// Alias for TypeCheckFailed
	ExitCodeInvalidJump		uint32	= 14
	ExitCodeCallStackOverflow	uint32	= 15
	ExitCodeContinuationMissing	uint32	= 16
	ExitCodeContinuationNotFound	uint32	= 16	// Alias for ContinuationMissing
	ExitCodeRecursionLimitExceeded	uint32	= 17

	// Memory / Type Safety Errors (18-21)
	ExitCodeInvalidMemoryAccess	uint32	= 18
	ExitCodeNullReference		uint32	= 19
	ExitCodeInvalidChunkRef		uint32	= 20
	ExitCodeInvalidChunkReference	uint32	= 20	// Alias for InactiveFrozen
	ExitCodeCorruptedStateObject	uint32	= 21

	// Arithmetic Edge Cases (22-24)
	ExitCodeDivisionByZero		uint32	= 22
	ExitCodeInvalidShift		uint32	= 23
	ExitCodeArithmeticUnderflow	uint32	= 24

	// Execution Safety / Gas Edge Cases (25-27)
	ExitCodeGasLimitExceeded	uint32	= 25
	ExitCodeGasReservationFailed	uint32	= 26
	ExitCodeExecutionTimeout	uint32	= 27

	// Action / Message Errors (32-63)
	ExitCodeMessageExpired		uint32	= 32
	ExitCodeQueueLimit		uint32	= 33
	ExitCodeMessageTooLarge		uint32	= 34
	ExitCodeActionBudgetExceeded	uint32	= 35
	ExitCodeRoutingFailed		uint32	= 38
	ExitCodeMessageRoutingFailed	uint32	= 38	// Alias for RoutingFailed
	ExitCodeQueueOverflow		uint32	= 39
	ExitCodeShardUnavailable	uint32	= 40

	// State / Storage Errors (64-95)
	ExitCodeStorageLimit		uint32	= 64
	ExitCodeStorageRentDebt		uint32	= 65
	ExitCodeAccountStateTooBig	uint32	= 66
	ExitCodeStateCorruption		uint32	= 67
	ExitCodeStateVersionMismatch	uint32	= 68
	ExitCodeSnapshotFailure		uint32	= 69

	// System / Host Errors (96-127)
	ExitCodeExecutionFailed		uint32	= 96
	ExitCodeInternalBounce		uint32	= 97
	ExitCodeForbiddenHostCall	uint32	= 98
	ExitCodeContractAbort		uint32	= 99
	ExitCodeAssertionFailed		uint32	= 100
	ExitCodeInsufficientBalance	uint32	= 101
	ExitCodeInsufficientGas		uint32	= 102
	ExitCodeExplicitAbort		uint32	= 103
	ExitCodeInactiveFrozen		uint32	= 104
)

// IsVMExecutionError returns true if code is in VM execution domain (0-31)
func IsVMExecutionError(code uint32) bool {
	return code <= 31
}

// IsActionMessageError returns true if code is in action/message domain (32-63)
func IsActionMessageError(code uint32) bool {
	return code >= 32 && code <= 63
}

// IsStateStorageError returns true if code is in state/storage domain (64-95)
func IsStateStorageError(code uint32) bool {
	return code >= 64 && code <= 95
}

// IsSystemHostError returns true if code is in system/host domain (96+)
func IsSystemHostError(code uint32) bool {
	return code >= 96
}

// ExitCodeName returns the human-readable name for an exit code
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
	case ExitCodeTypeCheckFailed:
		return "type_check_failed"
	case ExitCodeTypeCheckError:
		return "type_check_error"
	case ExitCodeOutOfGas:
		return "out_of_gas"
	case ExitCodeStackOverflow:
		return "stack_overflow"
	case ExitCodeStackUnderflow:
		return "stack_underflow"
	case ExitCodeInvalidJump:
		return "invalid_jump"
	case ExitCodeCallStackOverflow:
		return "call_stack_overflow"
	case ExitCodeContinuationMissing:
		return "continuation_missing"
	case ExitCodeRecursionLimitExceeded:
		return "recursion_limit_exceeded"
	case ExitCodeInvalidMemoryAccess:
		return "invalid_memory_access"
	case ExitCodeNullReference:
		return "null_reference"
	case ExitCodeInvalidChunkRef:
		return "invalid_chunk_reference"
	case ExitCodeCorruptedStateObject:
		return "corrupted_state_object"
	case ExitCodeDivisionByZero:
		return "division_by_zero"
	case ExitCodeInvalidShift:
		return "invalid_shift"
	case ExitCodeArithmeticUnderflow:
		return "arithmetic_underflow"
	case ExitCodeGasLimitExceeded:
		return "gas_limit_exceeded"
	case ExitCodeGasReservationFailed:
		return "gas_reservation_failed"
	case ExitCodeExecutionTimeout:
		return "execution_timeout"
	case ExitCodeMessageExpired:
		return "message_expired"
	case ExitCodeQueueLimit:
		return "queue_limit"
	case ExitCodeMessageTooLarge:
		return "message_too_large"
	case ExitCodeActionBudgetExceeded:
		return "action_budget_exceeded"
	case ExitCodeRoutingFailed:
		return "routing_failed"
	case ExitCodeQueueOverflow:
		return "queue_overflow"
	case ExitCodeShardUnavailable:
		return "shard_unavailable"
	case ExitCodeStorageLimit:
		return "storage_limit"
	case ExitCodeStorageRentDebt:
		return "storage_rent_debt"
	case ExitCodeAccountStateTooBig:
		return "account_state_too_big"
	case ExitCodeStateCorruption:
		return "state_corruption"
	case ExitCodeStateVersionMismatch:
		return "state_version_mismatch"
	case ExitCodeSnapshotFailure:
		return "snapshot_failure"
	case ExitCodeExecutionFailed:
		return "execution_failed"
	case ExitCodeInternalBounce:
		return "internal_bounce"
	case ExitCodeForbiddenHostCall:
		return "forbidden_host_call"
	case ExitCodeContractAbort:
		return "contract_abort"
	case ExitCodeAssertionFailed:
		return "assertion_failed"
	case ExitCodeInsufficientBalance:
		return "insufficient_balance"
	case ExitCodeInsufficientGas:
		return "insufficient_gas"
	case ExitCodeExplicitAbort:
		return "explicit_abort"
	case ExitCodeInactiveFrozen:
		return "inactive_frozen"
	default:
		return "unknown"
	}
}

// CanonicalExitCodes returns the list of all defined exit codes
func CanonicalExitCodes() []ExitCodeSpec {
	return []ExitCodeSpec{
		{ExitCodeOK, "ok"},
		{ExitCodeValidationFailed, "validation_failed"},
		{ExitCodeUnauthorized, "unauthorized"},
		{ExitCodeAccountInactive, "account_inactive"},
		{ExitCodeAccountFrozen, "account_frozen"},
		{ExitCodeContractFrozen, "contract_frozen"},
		{ExitCodeCodeRejected, "code_rejected"},
		{ExitCodeTypeCheckFailed, "type_check_failed"},
		{ExitCodeOutOfGas, "out_of_gas"},
		{ExitCodeStackOverflow, "stack_overflow"},
		{ExitCodeStackUnderflow, "stack_underflow"},
		{ExitCodeInvalidJump, "invalid_jump"},
		{ExitCodeCallStackOverflow, "call_stack_overflow"},
		{ExitCodeContinuationMissing, "continuation_missing"},
		{ExitCodeRecursionLimitExceeded, "recursion_limit_exceeded"},
		{ExitCodeInvalidMemoryAccess, "invalid_memory_access"},
		{ExitCodeNullReference, "null_reference"},
		{ExitCodeInvalidChunkRef, "invalid_chunk_reference"},
		{ExitCodeCorruptedStateObject, "corrupted_state_object"},
		{ExitCodeDivisionByZero, "division_by_zero"},
		{ExitCodeInvalidShift, "invalid_shift"},
		{ExitCodeArithmeticUnderflow, "arithmetic_underflow"},
		{ExitCodeGasLimitExceeded, "gas_limit_exceeded"},
		{ExitCodeGasReservationFailed, "gas_reservation_failed"},
		{ExitCodeExecutionTimeout, "execution_timeout"},
		{ExitCodeMessageExpired, "message_expired"},
		{ExitCodeQueueLimit, "queue_limit"},
		{ExitCodeMessageTooLarge, "message_too_large"},
		{ExitCodeActionBudgetExceeded, "action_budget_exceeded"},
		{ExitCodeRoutingFailed, "routing_failed"},
		{ExitCodeQueueOverflow, "queue_overflow"},
		{ExitCodeShardUnavailable, "shard_unavailable"},
		{ExitCodeStorageLimit, "storage_limit"},
		{ExitCodeStorageRentDebt, "storage_rent_debt"},
		{ExitCodeAccountStateTooBig, "account_state_too_big"},
		{ExitCodeStateCorruption, "state_corruption"},
		{ExitCodeStateVersionMismatch, "state_version_mismatch"},
		{ExitCodeSnapshotFailure, "snapshot_failure"},
		{ExitCodeExecutionFailed, "execution_failed"},
		{ExitCodeInternalBounce, "internal_bounce"},
		{ExitCodeForbiddenHostCall, "forbidden_host_call"},
		{ExitCodeContractAbort, "contract_abort"},
		{ExitCodeAssertionFailed, "assertion_failed"},
		{ExitCodeInsufficientBalance, "insufficient_balance"},
		{ExitCodeInsufficientGas, "insufficient_gas"},
	}
}

// KnownExitCode returns true if the code is recognized
func KnownExitCode(code uint32) bool {
	return ExitCodeName(code) != "unknown"
}
