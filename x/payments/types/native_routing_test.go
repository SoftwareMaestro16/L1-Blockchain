package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNativePaymentGoalsAndAbstractionsCoverSectionNine(t *testing.T) {
	require.NoError(t, ValidateNativePaymentArchitectureDescriptors())
	require.Len(t, NativePaymentGoals(), 6)
	require.Len(t, NativePaymentAbstractions(), 7)

	expectedRoots := map[NativePaymentObject]PaymentCommittedRoot{
		NativePaymentObjectChannel:		PaymentCommittedRootChannel,
		NativePaymentObjectVirtualChannel:	PaymentCommittedRootVirtualChannel,
		NativePaymentObjectCondition:		PaymentCommittedRootCondition,
		NativePaymentObjectRoute:		PaymentCommittedRootRoute,
		NativePaymentObjectReservation:		PaymentCommittedRootReservation,
		NativePaymentObjectSettlementProof:	PaymentCommittedRootSettlementProof,
		NativePaymentObjectReceipt:		PaymentCommittedRootReceipt,
	}
	for _, abstraction := range NativePaymentAbstractions() {
		require.Equal(t, expectedRoots[abstraction.Object], abstraction.CommittedRoot)
		require.NotEmpty(t, abstraction.Purpose)
	}
}

func TestNativePaymentRoutingSnapshotCommitsAllObjectRootsDeterministically(t *testing.T) {
	fixture := nativeRoutingFixture(t)
	snapshot, err := BuildNativePaymentRoutingSnapshot(NativePaymentRoutingSnapshot{
		Height:			77,
		Channels:		[]PaymentChannel{fixture.Channel},
		VirtualChannels:	[]VirtualPaymentChannel{fixture.Virtual},
		Conditions:		[]ConditionalPayment{fixture.Condition},
		Routes:			[]PaymentRoute{fixture.Route},
		Reservations:		[]LiquidityReservation{fixture.Reservation},
		SettlementProofs:	[]SettlementProof{fixture.Proof},
		Receipts:		[]PaymentReceipt{fixture.Receipt},
	})
	require.NoError(t, err)
	require.NoError(t, snapshot.Validate())

	for _, abstraction := range NativePaymentAbstractions() {
		root, found := RootForNativePaymentObject(snapshot, abstraction.Object)
		require.True(t, found, abstraction.Object)
		require.NoError(t, ValidateHash(string(abstraction.CommittedRoot), root))
	}

	reordered, err := BuildNativePaymentRoutingSnapshot(NativePaymentRoutingSnapshot{
		Height:			77,
		Receipts:		[]PaymentReceipt{fixture.Receipt},
		SettlementProofs:	[]SettlementProof{fixture.Proof},
		Reservations:		[]LiquidityReservation{fixture.Reservation},
		Routes:			[]PaymentRoute{fixture.Route},
		Conditions:		[]ConditionalPayment{fixture.Condition},
		VirtualChannels:	[]VirtualPaymentChannel{fixture.Virtual},
		Channels:		[]PaymentChannel{fixture.Channel},
	})
	require.NoError(t, err)
	require.Equal(t, snapshot.StateRoot, reordered.StateRoot)
	require.Equal(t, snapshot.RouteRoot, reordered.RouteRoot)
}

func TestNativePaymentRouteRejectsUncommittedCapacityAndRouteMutation(t *testing.T) {
	route := testMsgPaymentRoute()

	_, err := NewPaymentRouteFromMsg(route, "99")
	require.ErrorContains(t, err, "capacity")

	committed, err := NewPaymentRouteFromMsg(route, "100")
	require.NoError(t, err)
	committed.Hops[0].FeeAmount = "0"
	committed.RouteRoot = ComputeNativePaymentRouteRoot(committed)
	require.ErrorContains(t, committed.Validate(), "commitment")
}

func TestNativePaymentSettlementProofRequiresTypedEvidence(t *testing.T) {
	fixture := nativeRoutingFixture(t)
	fraud := fixture.Proof
	fraud.ProofType = SettlementProofFraud
	fraud.FraudProofHashOptional = ""
	fraud.ProofRoot = ""
	_, err := BuildSettlementProof(fraud)
	require.ErrorContains(t, err, "fraud proof hash")

	fraud.FraudProofHashOptional = HashParts("native-routing-fraud-proof")
	built, err := BuildSettlementProof(fraud)
	require.NoError(t, err)
	require.NoError(t, built.Validate())
}

func TestNativePaymentSnapshotRejectsDuplicateReceipts(t *testing.T) {
	fixture := nativeRoutingFixture(t)
	_, err := BuildNativePaymentRoutingSnapshot(NativePaymentRoutingSnapshot{
		Height:		88,
		Receipts:	[]PaymentReceipt{fixture.Receipt, fixture.Receipt},
	})
	require.ErrorContains(t, err, "duplicate native receipt")
}

type nativeRoutingFixtureSet struct {
	Channel		PaymentChannel
	Virtual		VirtualPaymentChannel
	Condition	ConditionalPayment
	Route		PaymentRoute
	Reservation	LiquidityReservation
	Proof		SettlementProof
	Receipt		PaymentReceipt
}

func nativeRoutingFixture(t *testing.T) nativeRoutingFixtureSet {
	t.Helper()

	alice := testAddress(0x81)
	bob := testAddress(0x82)
	channelRecord := signedChannel(t, "native-routing-channel", "1000", alice, bob)
	channel, err := NewPaymentChannelFromRecord(channelRecord, "financial", 3)
	require.NoError(t, err)

	virtual, err := BuildVirtualPaymentChannel(VirtualPaymentChannel{
		VirtualChannelID:	HashParts("native-routing-virtual"),
		ChainID:		channelRecord.ChainID,
		ZoneID:			"financial",
		ShardID:		4,
		ParentRouteID:		HashParts("native-routing-parent-route"),
		ParentChannelIDs:	[]string{channelRecord.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{testAddress(0x83)},
		Capacity:		"300",
		BalanceA:		"250",
		BalanceB:		"50",
		RoutingFeeAmount:	"2",
		ExpiresHeight:		120,
		Status:			VirtualChannelStatusOpen,
		StateHash:		HashParts("native-routing-virtual-state"),
	})
	require.NoError(t, err)

	condition := ConditionalPayment{
		ConditionID:	HashParts("native-routing-condition"),
		ConditionType:	ConditionTypeHashLock,
		Payer:		alice,
		Payee:		bob,
		Amount:		"25",
		HashLock:	HashParts("native-routing-preimage"),
		TimeoutHeight:	110,
		NonceStart:	2,
		NonceEnd:	4,
	}
	require.NoError(t, condition.Validate())

	routeMsg := testMsgPaymentRoute()
	route, err := NewPaymentRouteFromMsg(routeMsg, "100")
	require.NoError(t, err)

	reservation, err := BuildLiquidityReservation(LiquidityReservation{
		ReservationID:	HashParts("native-routing-reservation"),
		RouteID:	route.RouteID,
		ChannelID:	route.Hops[0].ChannelID,
		Participant:	route.Payer,
		Amount:		"100",
		FeeAmount:	"5",
		ReservedHeight:	50,
		ExpiryHeight:	90,
	})
	require.NoError(t, err)

	proof, err := BuildSettlementProof(SettlementProof{
		ProofID:		HashParts("native-routing-proof"),
		ProofType:		SettlementProofLatestState,
		ChannelID:		channel.ChannelID,
		LatestStateHash:	channel.LatestStateHash,
		SubmittedBy:		alice,
		Height:			76,
	})
	require.NoError(t, err)

	receipt, err := BuildPaymentReceipt(PaymentReceipt{
		PaymentID:	HashParts("native-routing-receipt"),
		RouteID:	route.RouteID,
		ChannelID:	channel.ChannelID,
		Status:		PaymentReceiptSettled,
		Amount:		"100",
		FeeAmount:	"5",
		ValueReturned:	"0",
		Height:		77,
	})
	require.NoError(t, err)

	return nativeRoutingFixtureSet{
		Channel:	channel,
		Virtual:	virtual,
		Condition:	condition,
		Route:		route,
		Reservation:	reservation,
		Proof:		proof,
		Receipt:	receipt,
	}
}
