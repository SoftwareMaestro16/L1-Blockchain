package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testMsgPaymentRoute() MsgPaymentRoute {
	alice := testAddress(0x71)
	router := testAddress(0x72)
	bob := testAddress(0x73)
	return MsgPaymentRoute{
		RouteID:	HashParts("msg-payment-route"),
		Payer:		alice,
		Payee:		bob,
		Amount:		"100",
		MaxFee:		"5",
		ConditionRoot:	HashParts("msg-payment-route-condition"),
		ExpiryHeight:	100,
		SettlementMode:	ConditionSettlementModePreimage,
		Hops: []PaymentRouteHop{
			{ChannelID: HashParts("msg-payment-route-channel-1"), From: alice, To: router, FeeAmount: "1", TimeoutHeight: 80},
			{ChannelID: HashParts("msg-payment-route-channel-2"), From: router, To: bob, FeeAmount: "2", TimeoutHeight: 90},
		},
	}
}

func admissionForRoute(route MsgPaymentRoute) PaymentRouteAdmission {
	return PaymentRouteAdmission{
		CurrentHeight:	50,
		Commitments: []PaymentRouteCommitment{{
			RouteID:	route.RouteID,
			Committer:	route.Payer,
			CommitmentHash:	ComputePaymentRouteCommitmentHash(route),
			Signed:		true,
			ExpiresHeight:	90,
		}},
		Balances:	[]PaymentRouteBalance{{Participant: route.Payer, Available: "105"}},
		SupportedSettlementModes: []ConditionSettlementMode{
			ConditionSettlementModePreimage,
			ConditionSettlementModeExpiry,
		},
	}
}

func TestMsgPaymentRouteAdmissionValidatesCommitmentBalanceAndPolicy(t *testing.T) {
	route := testMsgPaymentRoute()
	admission := admissionForRoute(route)

	built, err := BuildMsgPaymentRoute(route, admission)
	require.NoError(t, err)
	require.Equal(t, route.RouteID, built.RouteID)

	unsigned := admissionForRoute(route)
	unsigned.Commitments[0].Signed = false
	require.ErrorContains(t, route.Validate(unsigned), "signed or reserved")

	insufficient := admissionForRoute(route)
	insufficient.Balances = []PaymentRouteBalance{{Participant: route.Payer, Available: "104"}}
	require.ErrorContains(t, route.Validate(insufficient), "unavailable")

	unordered := route
	unordered.Hops = append([]PaymentRouteHop{}, route.Hops...)
	unordered.Hops[1].TimeoutHeight = unordered.Hops[0].TimeoutHeight
	require.ErrorContains(t, unordered.Validate(admissionForRoute(unordered)), "timeouts")

	route = testMsgPaymentRoute()
	unsupported := admissionForRoute(route)
	unsupported.SupportedSettlementModes = []ConditionSettlementMode{ConditionSettlementModeExpiry}
	require.ErrorContains(t, route.Validate(unsupported), "settlement mode")
}

func TestPaymentRouteScoringChangesDeterministicallyBetweenEpochs(t *testing.T) {
	route := testMsgPaymentRoute()
	epoch1, err := ApplyPaymentRoutingEpochUpdate(PaymentRouteTableState{}, PaymentRoutingEpochUpdate{
		Epoch:		1,
		CurrentHeight:	50,
		Routes:		[]MsgPaymentRoute{route},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), epoch1.Epoch)
	require.NotEmpty(t, epoch1.RootHash)

	lowScore, err := DeterministicPaymentRouteScore(route, []PaymentRouteCongestionSnapshot{{
		RouteID:	route.RouteID,
		ChannelID:	route.Hops[0].ChannelID,
		HopIndex:	0,
		CongestionBps:	100,
		ObservedHeight:	50,
	}})
	require.NoError(t, err)
	highScore, err := DeterministicPaymentRouteScore(route, []PaymentRouteCongestionSnapshot{{
		RouteID:		route.RouteID,
		ChannelID:		route.Hops[0].ChannelID,
		HopIndex:		0,
		CongestionBps:		2_500,
		PendingMessageCount:	4,
		RetryCount:		2,
		ObservedHeight:		51,
	}})
	require.NoError(t, err)
	require.Greater(t, highScore, lowScore)

	epoch2, err := ApplyPaymentRoutingEpochUpdate(epoch1, PaymentRoutingEpochUpdate{
		Epoch:		2,
		CurrentHeight:	51,
		Routes:		[]MsgPaymentRoute{route},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(2), epoch2.Epoch)
	require.NotEqual(t, epoch1.RootHash, epoch2.RootHash)
}

func TestPaymentRouteSchedulerRetriesAndExpires(t *testing.T) {
	route := testMsgPaymentRoute()
	policy := RouteRetryPolicy{MaxAttempts: 2, AlternateRouteLimit: 2, CongestionRetryDelay: 5}

	first, err := SchedulePaymentRouteDelivery(route, 50, nil, policy)
	require.NoError(t, err)
	require.Len(t, first.Tasks, 2)
	require.Equal(t, uint32(1), first.Tasks[0].Attempt)
	require.Equal(t, uint64(50), first.Tasks[0].DeliverAfterHeight)
	require.Equal(t, PaymentRouteReceiptRetry, first.Receipt.Status)

	second, err := SchedulePaymentRouteDelivery(route, 55, first.Tasks, policy)
	require.NoError(t, err)
	require.Len(t, second.Tasks, 2)
	require.Equal(t, uint32(2), second.Tasks[0].Attempt)
	require.Equal(t, uint64(60), second.Tasks[0].DeliverAfterHeight)

	exhausted, err := SchedulePaymentRouteDelivery(route, 61, append(first.Tasks, second.Tasks...), policy)
	require.NoError(t, err)
	require.Empty(t, exhausted.Tasks)
	require.Equal(t, PaymentRouteReceiptBounced, exhausted.Receipt.Status)
	require.NoError(t, ValidatePaymentRouteBounceConservation(route, exhausted.Receipt))

	expired, err := SchedulePaymentRouteDelivery(route, 101, nil, policy)
	require.NoError(t, err)
	require.Empty(t, expired.Tasks)
	require.Equal(t, PaymentRouteReceiptExpired, expired.Receipt.Status)

	state, err := RecordPaymentRouteReceipt(PaymentRouteTableState{Epoch: 1}, expired.Receipt)
	require.NoError(t, err)
	found, ok := QueryPaymentRouteReceipt(state, route.RouteID, 0)
	require.True(t, ok)
	require.Equal(t, expired.Receipt.ReceiptHash, found.ReceiptHash)
}

func TestPaymentRouteBounceValueConservation(t *testing.T) {
	route := testMsgPaymentRoute()
	receipt, err := BuildPaymentRouteBounceReceipt(route, 101, PaymentRouteReceiptExpired)
	require.NoError(t, err)
	require.Equal(t, "105", receipt.ValueReturned)
	require.NoError(t, ValidatePaymentRouteBounceConservation(route, receipt))

	leaking := receipt
	leaking.ValueReturned = "100"
	leaking.ReceiptHash = ComputePaymentRouteReceiptHash(leaking)
	require.ErrorContains(t, ValidatePaymentRouteBounceConservation(route, leaking), "conservation")
}
