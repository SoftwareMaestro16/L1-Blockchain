package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMGasScheduleDeclaresSection81GasClasses(t *testing.T) {
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)
	require.NoError(t, schedule.Validate())
	require.Equal(t, ComputeAVMGasScheduleHash(schedule), schedule.ScheduleHash)

	classes := make(map[AVMGasClass]uint64)
	for _, budget := range schedule.ClassBudgets {
		classes[budget.Class] = budget.Limit
	}
	for _, class := range []AVMGasClass{
		AVMGasClassExecution,
		AVMGasClassStorage,
		AVMGasClassScheduling,
		AVMGasClassCrossZoneRouting,
		AVMGasClassProofVerification,
		AVMGasClassContinuation,
		AVMGasClassInterfaceIntrospection,
	} {
		require.NotZero(t, classes[class], string(class))
	}

	mutated := schedule
	mutated.BounceGas++
	require.NotEqual(t, schedule.ScheduleHash, ComputeAVMGasScheduleHash(mutated))
}

func TestAVMGasPolicyDefinesSection83ParametersAndDerivedSchedule(t *testing.T) {
	policy, err := DefaultAVMGasPolicy()
	require.NoError(t, err)
	require.NoError(t, policy.Validate())
	require.Equal(t, ComputeAVMGasPolicyHash(policy), policy.PolicyHash)
	require.Positive(t, policy.BaseMessageGas)
	require.Positive(t, policy.PerBytePayloadGas)
	require.Positive(t, policy.StorageReadGas)
	require.Positive(t, policy.StorageWriteGas)
	require.Positive(t, policy.QueueInsertGas)
	require.Positive(t, policy.QueuePopGas)
	require.Positive(t, policy.CrossZoneBaseGas)
	require.Positive(t, policy.ProofVerifyBaseGas)
	require.Positive(t, policy.ContinuationStoreGas)
	require.Positive(t, policy.BounceBaseGas)

	schedule, err := AVMGasScheduleFromPolicy(policy, true, 1_000)
	require.NoError(t, err)
	require.Equal(t, policy.QueueInsertGas, schedule.SchedulingGas)
	require.Equal(t, policy.QueuePopGas, schedule.RetryGas)
	require.Equal(t, policy.CrossZoneBaseGas, schedule.CrossZoneRoutingGas)
	require.Equal(t, policy.ProofVerifyBaseGas, schedule.ProofVerificationGas)
	require.Equal(t, policy.ContinuationStoreGas, schedule.ContinuationGas)
	require.Equal(t, policy.BounceBaseGas, schedule.BounceGas)

	mutated := policy
	mutated.StorageWriteGas++
	require.NotEqual(t, policy.PolicyHash, ComputeAVMGasPolicyHash(mutated))
}

func TestAVMGasPolicyMetersPayloadProofStorageAndQueueCosts(t *testing.T) {
	policy, err := DefaultAVMGasPolicy()
	require.NoError(t, err)
	msg := testAVMGasMessage(t)

	admission, err := AVMMessageAdmissionGas(msg, policy)
	require.NoError(t, err)
	expected := policy.BaseMessageGas +
		policy.PerBytePayloadGas*uint64(len(msg.Payload)) +
		msg.GasLimit +
		policy.QueueInsertGas +
		policy.CrossZoneBaseGas +
		policy.ProofVerifyBaseGas*2 +
		policy.ContinuationStoreGas
	require.Equal(t, expected, admission)

	readGas, err := AVMStorageReadGas(7, policy)
	require.NoError(t, err)
	require.Equal(t, policy.StorageReadGas*7, readGas)
	writeGas, err := AVMStorageWriteGas(11, policy)
	require.NoError(t, err)
	require.Equal(t, policy.StorageWriteGas*11, writeGas)
	proofGas, err := AVMProofVerificationGas(2, policy)
	require.NoError(t, err)
	require.Equal(t, policy.ProofVerifyBaseGas*2, proofGas)
}

func TestAVMAsyncGasReservationEscrowsConsumesRetryBounceAndRefunds(t *testing.T) {
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)
	msg := testAVMGasMessage(t)

	reserve, err := NewAVMAsyncGasReserve(msg, schedule)
	require.NoError(t, err)
	require.NoError(t, reserve.Validate())
	require.Equal(t, msg.ID, reserve.MessageID)
	require.Equal(t, msg.DestinationZone, reserve.ZoneID)
	require.Equal(t, uint64(190), reserve.ReservedGas)
	require.Equal(t, reserve.ReservedGas, reserve.EscrowedGas)
	require.Equal(t, reserve.ReservedGas, reserve.RemainingGas)

	reserve, err = ConsumeAVMReservedGas(reserve, AVMGasClassExecution, 40)
	require.NoError(t, err)
	require.Equal(t, uint64(150), reserve.RemainingGas)

	reserve, err = ChargeAVMRetryGas(reserve, msg.RetryPolicy, schedule)
	require.NoError(t, err)
	require.Equal(t, uint64(140), reserve.RemainingGas)

	reserve, err = ChargeAVMBounceGas(reserve, schedule)
	require.NoError(t, err)
	require.Equal(t, uint64(115), reserve.RemainingGas)

	finalized, err := FinalizeAVMAsyncGasRefund(reserve, schedule)
	require.NoError(t, err)
	require.Equal(t, uint64(115), finalized.RefundGas)
	require.Zero(t, finalized.RemainingGas)
	require.Equal(t, ComputeAVMAsyncGasReserveHash(finalized), finalized.ReserveHash)
}

func TestAVMAsyncGasRefundPolicyCanRetainUnusedGas(t *testing.T) {
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)
	schedule.RefundUnused = false
	schedule.MaxRefundGas = 0
	schedule.ScheduleHash = ComputeAVMGasScheduleHash(schedule)
	msg := testAVMGasMessage(t)

	reserve, err := NewAVMAsyncGasReserve(msg, schedule)
	require.NoError(t, err)
	reserve, err = ConsumeAVMReservedGas(reserve, AVMGasClassExecution, 100)
	require.NoError(t, err)
	finalized, err := FinalizeAVMAsyncGasRefund(reserve, schedule)
	require.NoError(t, err)
	require.Zero(t, finalized.RefundGas)
	require.Equal(t, reserve.RemainingGas, finalized.RemainingGas)
}

func TestAVMAsyncGasRejectsUnescrowedOverconsumptionAndMalformedSchedule(t *testing.T) {
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)
	msg := testAVMGasMessage(t)
	reserve, err := NewAVMAsyncGasReserve(msg, schedule)
	require.NoError(t, err)

	badEscrow := reserve
	badEscrow.EscrowedGas--
	badEscrow.ReserveHash = ComputeAVMAsyncGasReserveHash(badEscrow)
	require.ErrorContains(t, badEscrow.Validate(), "fully escrowed upfront")

	_, err = ConsumeAVMReservedGas(reserve, AVMGasClassExecution, reserve.RemainingGas+1)
	require.ErrorContains(t, err, "exceeds reserved gas")

	badSchedule := schedule
	badSchedule.ClassBudgets = badSchedule.ClassBudgets[:len(badSchedule.ClassBudgets)-1]
	badSchedule.ScheduleHash = ComputeAVMGasScheduleHash(badSchedule)
	require.ErrorContains(t, badSchedule.Validate(), "every gas class")
}

func TestAVMRetryGasSkipsNonePolicy(t *testing.T) {
	schedule, err := DefaultAVMGasSchedule()
	require.NoError(t, err)
	msg := testAVMGasMessage(t)
	msg.RetryPolicy = AVMRetryPolicy{Mode: AVMRetryModeNone, BackoffMode: AVMBackoffModeNone}
	msg, err = NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	reserve, err := NewAVMAsyncGasReserve(msg, schedule)
	require.NoError(t, err)

	next, err := ChargeAVMRetryGas(reserve, msg.RetryPolicy, schedule)
	require.NoError(t, err)
	require.Equal(t, reserve.RemainingGas, next.RemainingGas)
	require.Equal(t, reserve.ReserveHash, next.ReserveHash)
}

func TestAVMZoneAsyncGasMeterConsumesPerZoneBudget(t *testing.T) {
	meter, err := NewAVMZoneAsyncGasMeter(zonestypes.ZoneIDContract, zonestypes.ZoneExecutionBudget{MaxGas: 50, MaxMessages: 2})
	require.NoError(t, err)

	meter, err = ConsumeAVMZoneAsyncGas(meter, AVMGasClassExecution, 20, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(20), meter.Budget.GasUsed)
	require.Equal(t, uint32(1), meter.Budget.MessagesUsed)
	require.Equal(t, ComputeAVMZoneAsyncGasMeterHash(meter), meter.MeterHash)

	_, err = ConsumeAVMZoneAsyncGas(meter, AVMGasClassExecution, 31, 1)
	require.ErrorContains(t, err, "exceeds max gas")
	_, err = ConsumeAVMZoneAsyncGas(meter, AVMGasClassExecution, 1, 2)
	require.ErrorContains(t, err, "messages used exceeds")
}

func TestAVMGasInvariantsRejectLimitBudgetRoutingProofStorageAndFailureDrift(t *testing.T) {
	policy, err := DefaultAVMGasPolicy()
	require.NoError(t, err)
	msg := testAVMGasMessage(t)

	require.NoError(t, ValidateAVMExecutionGasLimit(msg, msg.GasLimit))
	require.ErrorContains(t, ValidateAVMExecutionGasLimit(msg, msg.GasLimit+1), "exceeds message gas limit")

	reserve, err := NewAVMAsyncGasReserveWithPolicy(msg, policy, true, 1_000)
	require.NoError(t, err)
	require.NoError(t, ValidateAVMContractEmissionGasReserve(msg, reserve))
	missingRouting := reserve
	missingRouting.Consumed = filterAVMGasCharges(missingRouting.Consumed, AVMGasClassCrossZoneRouting)
	missingRouting.ReserveHash = ComputeAVMAsyncGasReserveHash(missingRouting)
	require.ErrorContains(t, ValidateAVMContractEmissionGasReserve(msg, missingRouting), "routing gas reserve")
	wrongMsg := msg
	wrongMsg.SenderNonce++
	wrongMsg, err = NewAVMAsyncMessage(wrongMsg)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateAVMContractEmissionGasReserve(wrongMsg, reserve), "id mismatch")

	require.NoError(t, ValidateAVMProofGasMetering(2, policy.ProofVerifyBaseGas*2, policy))
	require.ErrorContains(t, ValidateAVMProofGasMetering(2, policy.ProofVerifyBaseGas, policy), "under-metered")
	require.NoError(t, ValidateAVMStorageWriteGas(3, policy.StorageWriteGas*3, policy))
	require.ErrorContains(t, ValidateAVMStorageWriteGas(3, policy.StorageWriteGas*2, policy), "under-metered")

	receipt := testAVMFailedGasReceipt(t, msg, 9)
	require.NoError(t, ValidateAVMFailedExecutionGas(receipt, 9))
	require.ErrorContains(t, ValidateAVMFailedExecutionGas(receipt, 8), "gas drift")
	executed := receipt
	executed.Status = AVMReceiptStatusExecuted
	executed.ReceiptID = ""
	executed.ReceiptHash = ""
	executed, err = NewAVMExecutionReceipt(executed)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateAVMFailedExecutionGas(executed, 9), "requires failed receipt")
}

func FuzzAVMGasPolicyCostsStayBounded(f *testing.F) {
	f.Add(uint16(0), uint64(1), uint16(1), uint8(1))
	f.Add(uint16(32), uint64(500), uint16(64), uint8(2))
	f.Add(uint16(1023), uint64(99_999), uint16(4095), uint8(4))

	f.Fuzz(func(t *testing.T, payloadSize uint16, gasLimit uint64, storageBytes uint16, proofCount uint8) {
		policy, err := DefaultAVMGasPolicy()
		require.NoError(t, err)
		msg := testAVMAsyncMessage("fuzz-source", zonestypes.ZoneIDApplication, "fuzz-dest", zonestypes.ZoneIDContract, 13, 20)
		msg.Payload = make([]byte, int(payloadSize%1024)+1)
		msg.PayloadHash = ""
		msg.GasLimit = gasLimit%100_000 + 1
		if proofCount%2 == 1 {
			msg.AuthProofOptional = engineHash("fuzz-auth")
		}
		if proofCount%3 == 0 {
			msg.StateProofOptional = engineHash("fuzz-state")
		}
		msg, err = NewAVMAsyncMessage(msg)
		require.NoError(t, err)

		admission, err := AVMMessageAdmissionGas(msg, policy)
		require.NoError(t, err)
		require.GreaterOrEqual(t, admission, msg.GasLimit)

		writeGas, err := AVMStorageWriteGas(uint64(storageBytes%4096)+1, policy)
		require.NoError(t, err)
		require.Positive(t, writeGas)

		proofs := uint32(proofCount%4) + 1
		proofGas, err := AVMProofVerificationGas(proofs, policy)
		require.NoError(t, err)
		require.Equal(t, policy.ProofVerifyBaseGas*uint64(proofs), proofGas)
	})
}

func testAVMGasMessage(t *testing.T) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage("gas-source", zonestypes.ZoneIDApplication, "gas-dest", zonestypes.ZoneIDContract, 9, 10)
	msg.GasLimit = 100
	msg.DestinationActorOptional = "actor-gas"
	msg.AuthProofOptional = engineHash("auth-proof")
	msg.StateProofOptional = engineHash("state-proof")
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}

func testAVMFailedGasReceipt(t *testing.T, msg AVMAsyncMessage, gasUsed uint64) AVMExecutionReceipt {
	t.Helper()
	receipt, err := NewAVMExecutionReceipt(AVMExecutionReceipt{
		MessageID:		msg.ID,
		ZoneID:			msg.DestinationZone,
		Executor:		"gas-executor",
		Status:			AVMReceiptStatusFailed,
		GasUsed:		gasUsed,
		StorageWritten:		1,
		EventsHash:		engineHash("gas-events"),
		OutputMessagesRoot:	engineHash("gas-output"),
		ErrorCodeOptional:	"gas_failed",
		CreatedHeight:		20,
	})
	require.NoError(t, err)
	return receipt
}

func filterAVMGasCharges(charges []AVMGasCharge, excluded AVMGasClass) []AVMGasCharge {
	out := make([]AVMGasCharge, 0, len(charges))
	for _, charge := range charges {
		if charge.Class != excluded {
			out = append(out, charge)
		}
	}
	return out
}
