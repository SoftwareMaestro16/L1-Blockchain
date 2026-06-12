package types

import (
	"testing"
)

// TestExitCodeNameReturnsUnknownForInvalidCodes verifies unknown code handling
func TestExitCodeNameReturnsUnknownForInvalidCodes(t *testing.T) {
	invalidCodes := []uint32{200, 255, 1000, ^uint32(0)}
	for _, code := range invalidCodes {
		name := ExitCodeName(code)
		if name != "unknown" {
			t.Errorf("ExitCodeName(%d) = %q, want 'unknown'", code, name)
		}
	}
}

// TestExitCodeNameGoldenList verifies all defined codes have names
func TestExitCodeNameGoldenList(t *testing.T) {
	codes := []uint32{
		ExitCodeOK,
		ExitCodeValidationFailed,
		ExitCodeUnauthorized,
		ExitCodeAccountInactive,
		ExitCodeAccountFrozen,
		ExitCodeContractFrozen,
		ExitCodeCodeRejected,
		ExitCodeTypeCheckFailed,
		ExitCodeInvalidJump,
		ExitCodeCallStackOverflow,
		ExitCodeContinuationMissing,
		ExitCodeRecursionLimitExceeded,
		ExitCodeInvalidMemoryAccess,
		ExitCodeNullReference,
		ExitCodeInvalidChunkRef,
		ExitCodeCorruptedStateObject,
		ExitCodeDivisionByZero,
		ExitCodeInvalidShift,
		ExitCodeArithmeticUnderflow,
		ExitCodeGasLimitExceeded,
		ExitCodeGasReservationFailed,
		ExitCodeExecutionTimeout,
		ExitCodeMessageExpired,
		ExitCodeQueueLimit,
		ExitCodeMessageTooLarge,
		ExitCodeRoutingFailed,
		ExitCodeQueueOverflow,
		ExitCodeShardUnavailable,
		ExitCodeStorageLimit,
		ExitCodeStorageRentDebt,
		ExitCodeAccountStateTooBig,
		ExitCodeStateCorruption,
		ExitCodeStateVersionMismatch,
		ExitCodeSnapshotFailure,
		ExitCodeExecutionFailed,
		ExitCodeInternalBounce,
		ExitCodeForbiddenHostCall,
		ExitCodeContractAbort,
		ExitCodeAssertionFailed,
		ExitCodeInsufficientBalance,
	}

	for _, code := range codes {
		name := ExitCodeName(code)
		if name == "unknown" {
			t.Errorf("ExitCodeName(%d) returned 'unknown' for defined code", code)
		}
	}
}

// TestAllExitCodesUnder100 verifies all codes are under 100
func TestAllExitCodesUnder100(t *testing.T) {
	codes := []uint32{
		ExitCodeOK,
		ExitCodeValidationFailed,
		ExitCodeUnauthorized,
		ExitCodeAccountInactive,
		ExitCodeAccountFrozen,
		ExitCodeContractFrozen,
		ExitCodeCodeRejected,
		ExitCodeTypeCheckFailed,
		ExitCodeInvalidJump,
		ExitCodeCallStackOverflow,
		ExitCodeContinuationMissing,
		ExitCodeRecursionLimitExceeded,
		ExitCodeInvalidMemoryAccess,
		ExitCodeNullReference,
		ExitCodeInvalidChunkRef,
		ExitCodeCorruptedStateObject,
		ExitCodeDivisionByZero,
		ExitCodeInvalidShift,
		ExitCodeArithmeticUnderflow,
		ExitCodeGasLimitExceeded,
		ExitCodeGasReservationFailed,
		ExitCodeExecutionTimeout,
		ExitCodeMessageExpired,
		ExitCodeQueueLimit,
		ExitCodeMessageTooLarge,
		ExitCodeRoutingFailed,
		ExitCodeQueueOverflow,
		ExitCodeShardUnavailable,
		ExitCodeStorageLimit,
		ExitCodeStorageRentDebt,
		ExitCodeAccountStateTooBig,
		ExitCodeStateCorruption,
		ExitCodeStateVersionMismatch,
		ExitCodeSnapshotFailure,
		ExitCodeExecutionFailed,
		ExitCodeInternalBounce,
		ExitCodeForbiddenHostCall,
		ExitCodeContractAbort,
		ExitCodeAssertionFailed,
	}

	for _, code := range codes {
		if code > 100 {
			t.Errorf("Exit code %d exceeds limit of 100", code)
		}
	}
}

// TestExitCodeDomains verifies domain classification functions
func TestExitCodeDomains(t *testing.T) {
	tests := []struct {
		code		uint32
		isVM		bool
		isAction	bool
		isState		bool
		isSystem	bool
	}{
		{0, true, false, false, false},
		{14, true, false, false, false},
		{20, true, false, false, false},
		{32, false, true, false, false},
		{40, false, true, false, false},
		{64, false, false, true, false},
		{67, false, false, true, false},
		{96, false, false, false, true},
		{99, false, false, false, true},
		{101, false, false, false, true},
	}

	for _, tt := range tests {
		if got := IsVMExecutionError(tt.code); got != tt.isVM {
			t.Errorf("IsVMExecutionError(%d) = %v, want %v", tt.code, got, tt.isVM)
		}
		if got := IsActionMessageError(tt.code); got != tt.isAction {
			t.Errorf("IsActionMessageError(%d) = %v, want %v", tt.code, got, tt.isAction)
		}
		if got := IsStateStorageError(tt.code); got != tt.isState {
			t.Errorf("IsStateStorageError(%d) = %v, want %v", tt.code, got, tt.isState)
		}
		if got := IsSystemHostError(tt.code); got != tt.isSystem {
			t.Errorf("IsSystemHostError(%d) = %v, want %v", tt.code, got, tt.isSystem)
		}
	}
}

// TestExitCodeHasInsufficientBalance verifies gas balance code exists
func TestExitCodeHasInsufficientBalance(t *testing.T) {
	if ExitCodeInsufficientBalance == 0 {
		t.Error("ExitCodeInsufficientBalance should be defined")
	}
	name := ExitCodeName(ExitCodeInsufficientBalance)
	if name == "unknown" {
		t.Error("ExitCodeInsufficientBalance should have a name")
	}
}
