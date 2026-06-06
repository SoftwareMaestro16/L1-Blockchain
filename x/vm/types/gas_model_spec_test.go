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
