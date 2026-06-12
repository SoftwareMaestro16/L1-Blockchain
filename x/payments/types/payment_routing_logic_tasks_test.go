package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentRoutingLogicTaskSpecCoversSectionTwelveSix(t *testing.T) {
	require.NoError(t, ValidatePaymentRoutingLogicTaskSpec())

	spec, err := DefaultPaymentRoutingLogicTaskSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Tasks, 8)
	require.NotEmpty(t, spec.Root)

	byID := map[PaymentRoutingLogicTaskID]PaymentRoutingLogicTaskDescriptor{}
	for _, task := range spec.Tasks {
		require.NoError(t, task.Validate())
		byID[task.TaskID] = task
	}

	require.Contains(t, byID[PaymentRoutingTaskRouteTableState].Target, "PaymentRouteTableState")
	require.Contains(t, byID[PaymentRoutingTaskRouteTableState].Evidence, "ComputePaymentRouteTableRoot")
	require.Contains(t, byID[PaymentRoutingTaskEpochUpdateRules].Target, "ApplyPaymentRoutingEpochUpdate")
	require.Contains(t, byID[PaymentRoutingTaskDeterministicPathScoring].Target, "DeterministicPaymentRouteScore")
	require.Contains(t, byID[PaymentRoutingTaskDeliveryScheduler].Target, "SchedulePaymentRouteDelivery")
	require.Contains(t, byID[PaymentRoutingTaskReceiptCommitmentQuery].Evidence, "QueryPaymentRouteReceipt")
	require.Contains(t, byID[PaymentRoutingTaskCongestionEpochTests].Target, "TestPaymentRouteScoringChangesDeterministicallyBetweenEpochs")
	require.Contains(t, byID[PaymentRoutingTaskRetryExpiryTests].Target, "TestPaymentRouteSchedulerRetriesAndExpires")
	require.Contains(t, byID[PaymentRoutingTaskBounceConservationTests].Evidence, "ValidatePaymentRouteBounceConservation")
}

func TestPaymentRoutingLogicTaskSpecRootIsCanonicalAcrossInputOrder(t *testing.T) {
	spec, err := DefaultPaymentRoutingLogicTaskSpec()
	require.NoError(t, err)

	reordered := append([]PaymentRoutingLogicTaskDescriptor(nil), PaymentRoutingLogicTaskDescriptors()...)
	slices.Reverse(reordered)
	reorderedSpec, err := BuildPaymentRoutingLogicTaskSpec(reordered)
	require.NoError(t, err)

	require.Equal(t, spec.Root, reorderedSpec.Root)
	require.Equal(t, spec.Tasks, reorderedSpec.Tasks)
}

func TestPaymentRoutingLogicTaskSpecRejectsDuplicateAndTamperedDescriptors(t *testing.T) {
	duplicate, err := BuildPaymentRoutingLogicTaskSpec([]PaymentRoutingLogicTaskDescriptor{
		PaymentRoutingLogicTaskDescriptors()[0],
		PaymentRoutingLogicTaskDescriptors()[0],
	})
	require.ErrorContains(t, err, "duplicate payments routing task")
	require.Empty(t, duplicate.Root)

	_, err = BuildPaymentRoutingLogicTaskDescriptor(PaymentRoutingLogicTaskDescriptor{
		TaskID:		PaymentRoutingLogicTaskID("unknown"),
		Task:		"unknown",
		Target:		"unknown",
		Enforcement:	"unknown",
		Evidence:	"unknown",
	})
	require.ErrorContains(t, err, "unknown payments routing task")

	tampered := PaymentRoutingLogicTaskDescriptors()[0]
	tampered.Enforcement = strings.ReplaceAll(tampered.Enforcement, "canonical", "live")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}

func TestMsgPaymentRouteSpecFieldsAndValidationsRemainBoundToRouteTasks(t *testing.T) {
	route := testMsgPaymentRoute()
	admission := admissionForRoute(route)

	built, err := BuildMsgPaymentRoute(route, admission)
	require.NoError(t, err)
	require.Equal(t, route.RouteID, built.RouteID)
	require.Equal(t, route.Payer, built.Payer)
	require.Equal(t, route.Payee, built.Payee)
	require.Equal(t, route.Amount, built.Amount)
	require.Equal(t, route.MaxFee, built.MaxFee)
	require.Equal(t, route.Hops, built.Hops)
	require.Equal(t, route.ConditionRoot, built.ConditionRoot)
	require.Equal(t, route.ExpiryHeight, built.ExpiryHeight)
	require.Equal(t, route.SettlementMode, built.SettlementMode)

	reserved := admissionForRoute(route)
	reserved.Commitments[0].Signed = false
	reserved.Commitments[0].Reserved = true
	require.NoError(t, route.Validate(reserved))

	unsignedAndUnreserved := admissionForRoute(route)
	unsignedAndUnreserved.Commitments[0].Signed = false
	unsignedAndUnreserved.Commitments[0].Reserved = false
	require.ErrorContains(t, route.Validate(unsignedAndUnreserved), "signed or reserved")

	unsupported := admissionForRoute(route)
	unsupported.SupportedSettlementModes = []ConditionSettlementMode{ConditionSettlementModeExpiry}
	require.ErrorContains(t, route.Validate(unsupported), "settlement mode")
}
