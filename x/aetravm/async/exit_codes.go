package async

import contractstypes "github.com/sovereign-l1/l1/x/contracts/types"

type RuntimeExitCodeSpec struct {
	Code			uint32
	Name			string
	ContractExitCode	uint32
	ContractExitCodeName	string
}

var runtimeExitCodes = []RuntimeExitCodeSpec{
	{Code: ResultOK, Name: "ok", ContractExitCode: contractstypes.ExitCodeOK, ContractExitCodeName: "ok"},
	{Code: ResultNoDestination, Name: "no_destination", ContractExitCode: contractstypes.ExitCodeValidationFailed, ContractExitCodeName: "validation_failed"},
	{Code: ResultExpired, Name: "expired", ContractExitCode: contractstypes.ExitCodeMessageExpired, ContractExitCodeName: "message_expired"},
	{Code: ResultExecutionFailed, Name: "execution_failed", ContractExitCode: contractstypes.ExitCodeExecutionFailed, ContractExitCodeName: "execution_failed"},
	{Code: ResultLimitExceeded, Name: "limit_exceeded", ContractExitCode: contractstypes.ExitCodeQueueLimit, ContractExitCodeName: "queue_limit"},
	{Code: ResultBounceSuppressed, Name: "bounce_suppressed", ContractExitCode: contractstypes.ExitCodeInternalBounce, ContractExitCodeName: "internal_bounce"},
	{Code: ResultRefundSuppressed, Name: "refund_suppressed", ContractExitCode: contractstypes.ExitCodeInternalBounce, ContractExitCodeName: "internal_bounce"},
	{Code: ResultForbiddenHostCall, Name: "forbidden_host_call", ContractExitCode: contractstypes.ExitCodeForbiddenHostCall, ContractExitCodeName: "forbidden_host_call"},

	{Code: ResultInvalidJump, Name: "invalid_jump", ContractExitCode: contractstypes.ExitCodeInvalidJump, ContractExitCodeName: "invalid_jump"},
	{Code: ResultCallStackOverflow, Name: "call_stack_overflow", ContractExitCode: contractstypes.ExitCodeCallStackOverflow, ContractExitCodeName: "call_stack_overflow"},
	{Code: ResultContinuationNotFound, Name: "continuation_not_found", ContractExitCode: contractstypes.ExitCodeContinuationMissing, ContractExitCodeName: "continuation_missing"},
	{Code: ResultRecursionLimitExceeded, Name: "recursion_limit_exceeded", ContractExitCode: contractstypes.ExitCodeRecursionLimitExceeded, ContractExitCodeName: "recursion_limit_exceeded"},
	{Code: ResultInvalidMemoryAccess, Name: "invalid_memory_access", ContractExitCode: contractstypes.ExitCodeInvalidMemoryAccess, ContractExitCodeName: "invalid_memory_access"},
	{Code: ResultNullReference, Name: "null_reference", ContractExitCode: contractstypes.ExitCodeNullReference, ContractExitCodeName: "null_reference"},
	{Code: ResultInvalidChunkReference, Name: "invalid_chunk_reference", ContractExitCode: contractstypes.ExitCodeInvalidChunkReference, ContractExitCodeName: "invalid_chunk_reference"},
	{Code: ResultCorruptedStateObject, Name: "corrupted_state_object", ContractExitCode: contractstypes.ExitCodeCorruptedStateObject, ContractExitCodeName: "corrupted_state_object"},
	{Code: ResultDivisionByZero, Name: "division_by_zero", ContractExitCode: contractstypes.ExitCodeDivisionByZero, ContractExitCodeName: "division_by_zero"},
	{Code: ResultInvalidShift, Name: "invalid_shift", ContractExitCode: contractstypes.ExitCodeInvalidShift, ContractExitCodeName: "invalid_shift"},
	{Code: ResultArithmeticUnderflow, Name: "arithmetic_underflow", ContractExitCode: contractstypes.ExitCodeArithmeticUnderflow, ContractExitCodeName: "arithmetic_underflow"},
	{Code: ResultGasLimitExceeded, Name: "gas_limit_exceeded", ContractExitCode: contractstypes.ExitCodeGasLimitExceeded, ContractExitCodeName: "gas_limit_exceeded"},
	{Code: ResultGasReservationFailed, Name: "gas_reservation_failed", ContractExitCode: contractstypes.ExitCodeGasReservationFailed, ContractExitCodeName: "gas_reservation_failed"},
	{Code: ResultExecutionTimeout, Name: "execution_timeout", ContractExitCode: contractstypes.ExitCodeExecutionTimeout, ContractExitCodeName: "execution_timeout"},
	{Code: ResultStackOverflow, Name: "stack_overflow", ContractExitCode: contractstypes.ExitCodeStackOverflow, ContractExitCodeName: "stack_overflow"},
	{Code: ResultStackUnderflow, Name: "stack_underflow", ContractExitCode: contractstypes.ExitCodeStackUnderflow, ContractExitCodeName: "stack_underflow"},
	{Code: ResultTypeCheckError, Name: "type_check_error", ContractExitCode: contractstypes.ExitCodeTypeCheckError, ContractExitCodeName: "type_check_error"},

	{Code: ResultMessageRoutingFailed, Name: "message_routing_failed", ContractExitCode: contractstypes.ExitCodeRoutingFailed, ContractExitCodeName: "routing_failed"},
	{Code: ResultQueueOverflow, Name: "queue_overflow", ContractExitCode: contractstypes.ExitCodeQueueOverflow, ContractExitCodeName: "queue_overflow"},
	{Code: ResultShardUnavailable, Name: "shard_unavailable", ContractExitCode: contractstypes.ExitCodeShardUnavailable, ContractExitCodeName: "shard_unavailable"},
	{Code: ResultInsufficientBalance, Name: "insufficient_balance", ContractExitCode: contractstypes.ExitCodeInsufficientBalance, ContractExitCodeName: "insufficient_balance"},
	{Code: ResultInsufficientGas, Name: "insufficient_gas", ContractExitCode: contractstypes.ExitCodeInsufficientGas, ContractExitCodeName: "insufficient_gas"},
	{Code: ResultStateCorruption, Name: "state_corruption", ContractExitCode: contractstypes.ExitCodeStateCorruption, ContractExitCodeName: "state_corruption"},
	{Code: ResultStateVersionMismatch, Name: "state_version_mismatch", ContractExitCode: contractstypes.ExitCodeStateVersionMismatch, ContractExitCodeName: "state_version_mismatch"},
	{Code: ResultSnapshotFailure, Name: "snapshot_failure", ContractExitCode: contractstypes.ExitCodeSnapshotFailure, ContractExitCodeName: "snapshot_failure"},
	{Code: ResultExplicitAbort, Name: "explicit_abort", ContractExitCode: contractstypes.ExitCodeExplicitAbort, ContractExitCodeName: "explicit_abort"},
	{Code: ResultAssertionFailed, Name: "assertion_failed", ContractExitCode: contractstypes.ExitCodeAssertionFailed, ContractExitCodeName: "assertion_failed"},
	{Code: ResultAccountStateTooBig, Name: "account_state_too_big", ContractExitCode: contractstypes.ExitCodeAccountStateTooBig, ContractExitCodeName: "account_state_too_big"},
	{Code: ResultStorageRentDebt, Name: "storage_rent_debt", ContractExitCode: contractstypes.ExitCodeStorageRentDebt, ContractExitCodeName: "storage_rent_debt"},
	{Code: ResultInactiveFrozen, Name: "inactive_frozen", ContractExitCode: contractstypes.ExitCodeInactiveFrozen, ContractExitCodeName: "inactive_frozen"},
	{Code: ResultActionBudgetExceeded, Name: "action_budget_exceeded", ContractExitCode: contractstypes.ExitCodeActionBudgetExceeded, ContractExitCodeName: "action_budget_exceeded"},
}

func RuntimeExitCodes() []RuntimeExitCodeSpec {
	return append([]RuntimeExitCodeSpec(nil), runtimeExitCodes...)
}

func RuntimeExitCodeName(code uint32) string {
	for _, spec := range runtimeExitCodes {
		if spec.Code == code {
			return spec.Name
		}
	}
	return "unknown"
}

func ContractExitCodeForRuntime(code uint32, failedPhase string) uint32 {
	switch code {
	case ResultOK:
		return contractstypes.ExitCodeOK
	case ResultNoDestination:
		return contractstypes.ExitCodeValidationFailed
	case ResultExpired:
		return contractstypes.ExitCodeMessageExpired
	case ResultExecutionFailed:
		return contractstypes.ExitCodeExecutionFailed
	case ResultLimitExceeded:
		switch failedPhase {
		case FailedPhaseExecution:
			return contractstypes.ExitCodeOutOfGas
		case FailedPhaseStorage:
			return contractstypes.ExitCodeStorageLimit
		case FailedPhaseValidation:
			return contractstypes.ExitCodeValidationFailed
		case FailedPhaseQueue, FailedPhaseDispatch:
			return contractstypes.ExitCodeQueueLimit
		default:
			return contractstypes.ExitCodeQueueLimit
		}
	case ResultBounceSuppressed, ResultRefundSuppressed:
		return contractstypes.ExitCodeInternalBounce
	case ResultForbiddenHostCall:
		return contractstypes.ExitCodeForbiddenHostCall
	case ResultInvalidJump:
		return contractstypes.ExitCodeInvalidJump
	case ResultCallStackOverflow:
		return contractstypes.ExitCodeCallStackOverflow
	case ResultContinuationNotFound:
		return contractstypes.ExitCodeContinuationNotFound
	case ResultRecursionLimitExceeded:
		return contractstypes.ExitCodeRecursionLimitExceeded
	case ResultInvalidMemoryAccess:
		return contractstypes.ExitCodeInvalidMemoryAccess
	case ResultNullReference:
		return contractstypes.ExitCodeNullReference
	case ResultInvalidChunkReference:
		return contractstypes.ExitCodeInvalidChunkReference
	case ResultCorruptedStateObject:
		return contractstypes.ExitCodeCorruptedStateObject
	case ResultDivisionByZero:
		return contractstypes.ExitCodeDivisionByZero
	case ResultInvalidShift:
		return contractstypes.ExitCodeInvalidShift
	case ResultArithmeticUnderflow:
		return contractstypes.ExitCodeArithmeticUnderflow
	case ResultGasLimitExceeded:
		return contractstypes.ExitCodeGasLimitExceeded
	case ResultGasReservationFailed:
		return contractstypes.ExitCodeGasReservationFailed
	case ResultExecutionTimeout:
		return contractstypes.ExitCodeExecutionTimeout
	case ResultStackOverflow:
		return contractstypes.ExitCodeStackOverflow
	case ResultStackUnderflow:
		return contractstypes.ExitCodeStackUnderflow
	case ResultTypeCheckError:
		return contractstypes.ExitCodeTypeCheckError
	case ResultMessageRoutingFailed:
		return contractstypes.ExitCodeMessageRoutingFailed
	case ResultQueueOverflow:
		return contractstypes.ExitCodeQueueOverflow
	case ResultShardUnavailable:
		return contractstypes.ExitCodeShardUnavailable
	case ResultInsufficientBalance:
		return contractstypes.ExitCodeInsufficientBalance
	case ResultInsufficientGas:
		return contractstypes.ExitCodeInsufficientGas
	case ResultStateCorruption:
		return contractstypes.ExitCodeStateCorruption
	case ResultStateVersionMismatch:
		return contractstypes.ExitCodeStateVersionMismatch
	case ResultSnapshotFailure:
		return contractstypes.ExitCodeSnapshotFailure
	case ResultExplicitAbort:
		return contractstypes.ExitCodeExplicitAbort
	case ResultAssertionFailed:
		return contractstypes.ExitCodeAssertionFailed
	case ResultAccountStateTooBig:
		return contractstypes.ExitCodeAccountStateTooBig
	case ResultStorageRentDebt:
		return contractstypes.ExitCodeStorageRentDebt
	case ResultInactiveFrozen:
		return contractstypes.ExitCodeInactiveFrozen
	case ResultActionBudgetExceeded:
		return contractstypes.ExitCodeActionBudgetExceeded
	default:
		if code >= 100 && code < 200 {
			return contractstypes.ExitCodeContractAbort
		}
		return contractstypes.ExitCodeExecutionFailed
	}
}

func ContractExitCodeNameForRuntime(code uint32, failedPhase string) string {
	return contractstypes.ExitCodeName(ContractExitCodeForRuntime(code, failedPhase))
}
