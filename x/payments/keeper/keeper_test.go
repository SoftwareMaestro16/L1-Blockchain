package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
)

func TestDefaultGenesisIsDisabledAndValid(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Channels)
	require.Empty(t, gs.State.Settlements)
}

func TestKeeperFeatureGateRejectsPaymentMutationWhenDisabled(t *testing.T) {
	k := NewKeeper()

	err := k.OpenChannel(paymentstypes.ChannelRecord{})
	require.ErrorContains(t, err, "disabled")
}

func TestKeeperPaymentLifecycleWhenEnabled(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x61)
	bob := keeperAddress(0x62)
	channel := keeperSignedChannel(t, "keeper-channel", "500", alice, bob)
	require.NoError(t, k.OpenChannel(channel))

	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "250"},
		{Participant: bob, Amount: "250"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, closeState, alice, 20, "5"))
	height, found, err := k.QueryPendingFinalizationHeight(channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(28), height)
	require.NoError(t, k.AdvanceChannelFinality(channel.ChannelID, 28))
	debug, err := k.QueryStateHash(channel.ChannelID)
	require.NoError(t, err)
	require.Equal(t, paymentstypes.ChannelStatusPendingClose, debug.Status)
	require.Equal(t, closeState.Nonce, debug.PendingNonce)
	require.Equal(t, closeState.StateHash, debug.PendingStateHash)
	require.Equal(t, paymentstypes.ComputeStateHash(closeState), debug.ComputedPendingStateHash)
	require.Equal(t, closeState.Nonce, debug.DisputedNonce)

	settlement, err := k.FinalizeSettlement(channel.ChannelID, 30)
	require.NoError(t, err)
	require.Equal(t, "245", keeperAmountFor(settlement.FinalBalances, alice))
	require.Equal(t, "250", keeperAmountFor(settlement.FinalBalances, bob))

	settlements, page, err := k.Settlements(nil)
	require.NoError(t, err)
	require.Zero(t, page.NextOffset)
	require.Equal(t, []paymentstypes.SettlementRecord{settlement}, settlements)
}

func TestKeeperPaymentChannelModuleMessagesAndReplayGuards(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x71)
	bob := keeperAddress(0x72)
	outsider := keeperAddress(0x73)
	openReq := paymentstypes.ChannelOpenRequest{
		ChainID:		"aetra-test-1",
		Participants:		[]string{alice, bob},
		InitialBalances:	[]paymentstypes.Balance{{Participant: alice, Amount: "100"}, {Participant: bob, Amount: "0"}},
		ChannelType:		paymentstypes.ChannelTypeBidirectional,
		Collateral:		"100",
		CloseDelay:		8,
		ChallengePeriod:	8,
		FeePolicyID:		paymentstypes.NativeDenom,
		OpeningFeeDenom:	paymentstypes.NativeDenom,
		OpeningFeePaid:		paymentstypes.DefaultOpeningFee,
		OpenHeight:		10,
	}
	_, err := k.ValidatePaymentChannelAnte(paymentstypes.MsgOpenChannel{
		Signer:		outsider,
		Request:	openReq,
	})
	require.ErrorContains(t, err, "signer must be participant")

	result, err := k.HandlePaymentChannelMessage(paymentstypes.MsgOpenChannel{
		Signer:		alice,
		Request:	openReq,
	})
	require.NoError(t, err)
	require.Equal(t, paymentstypes.PaymentChannelMsgOpenChannel, result.MsgType)
	require.NoError(t, k.AssertPaymentChannelCollateralInvariant())
	exported := k.ExportGenesis()
	require.Len(t, exported.State.Channels, 1)
	channel := exported.State.Channels[0]

	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "55"},
		{Participant: bob, Amount: "45"},
	})
	plan, err := k.PaymentChannelAccessPlan(paymentstypes.MsgUnilateralClose{
		Signer:	alice,
		Request: paymentstypes.ChannelCloseRequest{
			ChannelID:	channel.ChannelID,
			ClosingState:	closeState,
			Submitter:	alice,
			CurrentHeight:	20,
			SettlementFee:	"0",
		},
	}, 20)
	require.NoError(t, err)
	require.Contains(t, plan.WriteKeys, paymentstypes.PaymentPendingCloseIndexKey(channel.ChannelID))

	_, err = k.HandlePaymentChannelMessage(paymentstypes.MsgUnilateralClose{
		Signer:	alice,
		Request: paymentstypes.ChannelCloseRequest{
			ChannelID:	channel.ChannelID,
			ClosingState:	closeState,
			Submitter:	alice,
			CurrentHeight:	20,
			SettlementFee:	"0",
		},
	})
	require.NoError(t, err)
	require.NoError(t, k.AssertPaymentChannelCollateralInvariant())
	exported = k.ExportGenesis()
	require.Equal(t, paymentstypes.ChannelStatusPendingClose, exported.State.Channels[0].Status)

	result, err = k.HandlePaymentChannelMessage(paymentstypes.MsgFinalizeClose{
		Signer:	bob,
		Request: paymentstypes.FinalSettlementRequest{
			ChannelID:	channel.ChannelID,
			CurrentHeight:	28,
		},
	})
	require.NoError(t, err)
	require.Equal(t, paymentstypes.PaymentChannelMsgFinalizeClose, result.MsgType)
	require.Equal(t, "55", keeperAmountFor(result.Settlement.FinalBalances, alice))
	require.Equal(t, "45", keeperAmountFor(result.Settlement.FinalBalances, bob))
	require.NoError(t, k.AssertPaymentChannelCollateralInvariant())
	require.ErrorContains(t, k.RejectEarlyTombstonePruning(channel.ChannelID, 30), "before replay horizon")
	require.NoError(t, k.RejectEarlyTombstonePruning(channel.ChannelID, 28+paymentstypes.DefaultReplayHorizon))

	snapshot, err := k.PaymentChannelModuleState(28)
	require.NoError(t, err)
	require.Len(t, snapshot.Participants, 2)
	require.Len(t, snapshot.SettlementTombstones, 1)
	accs, page, err := k.ChannelFeeAccumulators(nil, 28)
	require.NoError(t, err)
	require.Zero(t, page.NextOffset)
	require.Len(t, accs, 1)
}

func TestKeeperPaymentAPISurfaceMessagesAndQueries(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x81)
	bob := keeperAddress(0x82)
	openReq := paymentstypes.ChannelOpenRequest{
		ChainID:		"aetra-test-1",
		Participants:		[]string{alice, bob},
		InitialBalances:	[]paymentstypes.Balance{{Participant: alice, Amount: "200"}, {Participant: bob, Amount: "0"}},
		ChannelType:		paymentstypes.ChannelTypeBidirectional,
		Collateral:		"200",
		CloseDelay:		8,
		ChallengePeriod:	8,
		FeePolicyID:		paymentstypes.NativeDenom,
		OpeningFeeDenom:	paymentstypes.NativeDenom,
		OpeningFeePaid:		paymentstypes.DefaultOpeningFee,
		OpenHeight:		10,
	}
	result, err := k.HandlePaymentAPIMessage(paymentstypes.MsgOpenChannel{Signer: alice, Request: openReq})
	require.NoError(t, err)
	require.Equal(t, paymentstypes.PaymentAPIMsgOpenChannel, result.MsgName)
	exported := k.ExportGenesis()
	require.Len(t, exported.State.Channels, 1)
	channelID := exported.State.Channels[0].ChannelID

	channel, found, err := k.QueryChannel(channelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, channelID, channel.ChannelID)
	channels, err := k.QueryChannelsByParticipant(alice)
	require.NoError(t, err)
	require.Len(t, channels, 1)
	feeSchedule, err := k.QueryFeeSchedule()
	require.NoError(t, err)
	require.Equal(t, paymentstypes.NativeDenom, feeSchedule.Denom)
	capacity, err := k.QueryChannelCapacity(channelID, 11)
	require.NoError(t, err)
	require.Equal(t, "200", capacity.AvailableCapacity)

	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "120"},
		{Participant: bob, Amount: "80"},
	})
	result, err = k.HandlePaymentAPIMessage(paymentstypes.MsgUnilateralClose{Signer: alice, Request: paymentstypes.ChannelCloseRequest{
		ChannelID:	channelID,
		ClosingState:	closeState,
		CurrentHeight:	20,
		SettlementFee:	"0",
	}})
	require.NoError(t, err)
	require.Equal(t, paymentstypes.PaymentAPIMsgUnilateralClose, result.MsgName)
	pending, found, err := k.QueryPendingClose(channelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, closeState.StateHash, pending.State.StateHash)
	height, found, err := k.QueryFinalizationHeight(channelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(28), height)
	finalizations, err := k.QueryPendingFinalizations(21)
	require.NoError(t, err)
	require.Len(t, finalizations, 1)
	disputes, err := k.QueryActiveDisputes(21)
	require.NoError(t, err)
	require.Empty(t, disputes)
}

func TestKeeperStoreV2ParticipantChannelPagination(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x63)
	bob := keeperAddress(0x64)
	first := keeperSignedChannel(t, "keeper-store-v2-first", "100", alice, bob)
	second := keeperSignedChannel(t, "keeper-store-v2-second", "100", alice, bob)
	require.NoError(t, k.OpenChannel(first))
	require.NoError(t, k.OpenChannel(second))

	layout, err := k.StoreV2Layout()
	require.NoError(t, err)
	require.Len(t, layout.Channels, 2)
	entries, page, err := k.ParticipantChannels(alice, &prototype.PageRequest{Limit: 1})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, uint64(1), page.NextOffset)
	next, page, err := k.ParticipantChannels(alice, &prototype.PageRequest{Offset: page.NextOffset, Limit: 1})
	require.NoError(t, err)
	require.Len(t, next, 1)
	require.Zero(t, page.NextOffset)
	require.NotEqual(t, entries[0].ChannelID, next[0].ChannelID)
}

func TestKeeperPaymentFeeScheduleBlocksOpenBypass(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x6b)
	bob := keeperAddress(0x6c)
	channel := keeperSignedChannel(t, "keeper-fee-open", "100", alice, bob)
	schedule := paymentstypes.DefaultPaymentFeeSchedule()
	schedule.ChannelOpenFee = "3"
	require.NoError(t, k.ConfigurePaymentFeeSchedule(schedule))
	require.NoError(t, k.SetPaymentFeeMultiplier(paymentstypes.PaymentFeeMultiplier{
		FeeClass:	paymentstypes.PaymentFeeClassChannelOpen,
		MultiplierBps:	20_000,
		CongestionBps:	2_500,
		UpdatedHeight:	channel.OpenHeight,
	}))
	err := k.OpenChannel(channel)
	require.ErrorContains(t, err, "fee below required")
	channel.OpeningFeePaid = "6"
	require.NoError(t, k.OpenChannel(channel))
	exported := k.ExportGenesis()
	require.Len(t, exported.State.FeeCharges, 1)
	require.Equal(t, paymentstypes.PaymentFeeClassChannelOpen, exported.State.FeeCharges[0].FeeClass)
	require.Equal(t, "6", exported.State.FeeCharges[0].RequiredAmount)
}

func TestKeeperAsyncExecutionFinalizesQueuedChannel(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x6d)
	bob := keeperAddress(0x6e)
	channel := keeperSignedChannel(t, "keeper-async-finalize", "100", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, closeState, alice, 20, "0"))
	require.NoError(t, k.RefreshAsyncExecutionQueues(21))
	exported := k.ExportGenesis()
	require.Len(t, exported.State.AsyncFinalizationQueue, 1)
	result, err := k.ProcessAsyncExecutionQueues(28, 1, 0)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ProcessedFinalizations)
	exported = k.ExportGenesis()
	require.Len(t, exported.State.Settlements, 1)
	require.Len(t, exported.State.AsyncCompletions, 1)
	require.True(t, exported.State.AsyncFinalizationQueue[0].Completed)
}

func TestKeeperAdaptiveSyncSnapshotRecovery(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x65)
	bob := keeperAddress(0x66)
	channel := keeperSignedChannel(t, "keeper-adaptive-sync", "100", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, closeState, alice, 20, "0"))
	newer := keeperSignedState(t, channel, 3, closeState.StateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "35"},
		{Participant: bob, Amount: "65"},
	})
	require.NoError(t, k.DisputeClose(channel.ChannelID, newer, bob, 21))

	snapshot, err := k.AdaptiveSyncSnapshot(22)
	require.NoError(t, err)
	require.Len(t, snapshot.ActiveDisputes, 1)
	recovered, err := k.RecoverAdaptiveSyncSafety(snapshot)
	require.NoError(t, err)
	require.Contains(t, recovered.ActiveDisputeChannelIDs, channel.ChannelID)
	require.Contains(t, recovered.PendingFinalizationIDs, channel.ChannelID)
}

func TestKeeperRoutingEngineModuleSelectsRetriesAndProbes(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x74)
	router := keeperAddress(0x75)
	bob := keeperAddress(0x76)
	direct := keeperSignedChannel(t, "keeper-routing-direct", "500", alice, bob)
	first := keeperSignedChannel(t, "keeper-routing-first", "500", alice, router)
	second := keeperSignedChannel(t, "keeper-routing-second", "500", router, bob)
	require.NoError(t, k.OpenChannel(direct))
	require.NoError(t, k.OpenChannel(first))
	require.NoError(t, k.OpenChannel(second))

	policy := paymentstypes.DefaultRoutePolicy()
	policy.MaxHops = 4
	policy.HopPenalty = "0"
	engine, err := k.SnapshotRoutingEngine(paymentstypes.TopologyStore{}, policy, paymentstypes.DefaultGossipRateLimitPolicy(), paymentstypes.DefaultRouteFailureScoringPolicy())
	require.NoError(t, err)

	for _, env := range []paymentstypes.SignedGossipEnvelope{
		keeperRoutingEnvelope(t, paymentstypes.GossipMessage{MessageType: paymentstypes.GossipChannelUpdate, ChainID: direct.ChainID, ChannelID: direct.ChannelID, NodeID: alice, From: alice, To: bob, Capacity: "500", FeeDenom: paymentstypes.NativeDenom, FeeAmount: "9", ValidAfterHeight: 10, ValidUntilHeight: 50, Sequence: 1}, alice, 11),
		keeperRoutingEnvelope(t, paymentstypes.GossipMessage{MessageType: paymentstypes.GossipChannelUpdate, ChainID: first.ChainID, ChannelID: first.ChannelID, NodeID: alice, From: alice, To: router, Capacity: "500", FeeDenom: paymentstypes.NativeDenom, FeeAmount: "1", ValidAfterHeight: 10, ValidUntilHeight: 50, Sequence: 2}, alice, 11),
		keeperRoutingEnvelope(t, paymentstypes.GossipMessage{MessageType: paymentstypes.GossipChannelUpdate, ChainID: second.ChainID, ChannelID: second.ChannelID, NodeID: router, From: router, To: bob, Capacity: "500", FeeDenom: paymentstypes.NativeDenom, FeeAmount: "1", ValidAfterHeight: 10, ValidUntilHeight: 50, Sequence: 3}, router, 11),
	} {
		engine, _, err = k.ApplyRoutingGossip(engine, paymentstypes.MsgGossipChannelUpdate{Gossip: env}, 11)
		require.NoError(t, err)
	}

	engine, route, err := k.SelectRoutingPath(engine, paymentstypes.RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, Policy: policy})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, "2", route.TotalFee)

	probe, err := k.HandleCapacityProbe(engine, paymentstypes.CapacityProbeRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, MaxHops: 4, BlindedRouteHint: paymentstypes.HashParts("keeper-probe")}, router)
	require.NoError(t, err)
	require.True(t, probe.Available)
	require.NotEmpty(t, probe.RouteHash)

	engine, retry, err := k.RetryRoutingPath(engine, paymentstypes.RouteRetryRequest{
		Selection:	paymentstypes.RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 21, Policy: policy},
		Failures:	[]paymentstypes.RouteFailureReport{{ChannelID: first.ChannelID, From: alice, To: router, FailureClass: paymentstypes.RouteFailureCongestion, Retryable: true, ObservedHeight: 20}},
		Policy:		paymentstypes.RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(2), retry.Attempts)
	require.Len(t, retry.Route.Edges, 1)
	require.Equal(t, direct.ChannelID, retry.Route.Edges[0].ChannelID)

	engine, scores, err := k.ApplyRoutingFailures(engine, []paymentstypes.RouteFailureReport{{ChannelID: direct.ChannelID, From: alice, To: bob, FailureClass: paymentstypes.RouteFailureLiquidityStale, Retryable: true, ObservedHeight: 22}})
	require.NoError(t, err)
	require.Len(t, scores, 1)
	require.NotEmpty(t, engine.LocalPeerScores)
}

func TestKeeperConditionalPaymentsModuleMessagesAndInvariants(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x77)
	bob := keeperAddress(0x78)
	channel := keeperSignedChannel(t, "keeper-conditional", "500", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	base := keeperSignedReserveState(t, channel, 2, channel.OpeningStateHash, "40", "0", []paymentstypes.Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "10"},
	})
	require.NoError(t, k.AcceptSignedState(channel.ChannelID, base, 19))
	promiseChannel := channel
	promiseChannel.LatestState = base
	preimage := "keeper-conditional-preimage"
	promise := keeperSignedPromiseWithHashLock(t, promiseChannel, "keeper-conditional-promise", alice, bob, "20", "1", 7, 40, paymentstypes.HashParts(preimage))

	snapshot, err := k.HandleConditionalPaymentMessage(paymentstypes.MsgRegisterPromise{
		Signer:		alice,
		ChannelID:	channel.ChannelID,
		BaseState:	base,
		Promises:	[]paymentstypes.ConditionalPromise{promise},
		CurrentHeight:	20,
	})
	require.NoError(t, err)
	require.Len(t, snapshot.Promises, 1)
	require.Len(t, snapshot.ConditionRoots, 1)
	require.NoError(t, k.AssertConditionalReserveInvariant())

	snapshot, err = k.HandleConditionalPaymentMessage(paymentstypes.MsgResolveWithPreimage{Request: paymentstypes.PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]paymentstypes.ConditionalPromise{promise},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	30,
	}})
	require.NoError(t, err)
	require.Len(t, snapshot.PreimageClaims, 1)
	require.Empty(t, snapshot.Promises)
	require.NoError(t, k.AssertConditionalReserveInvariant())

	_, err = k.HandleConditionalPaymentMessage(paymentstypes.MsgResolveWithPreimage{Request: paymentstypes.PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]paymentstypes.ConditionalPromise{promise},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	31,
	}})
	require.ErrorContains(t, err, "already been settled")
}

func TestKeeperLiquidityOptimizationModuleReservationsForecastsAndDecay(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x79)
	bob := keeperAddress(0x7a)
	channel := keeperSignedChannel(t, "keeper-liquidity-optimization", "500", alice, bob)
	require.NoError(t, k.OpenChannel(channel))

	liquidity, err := k.HandleLiquidityOptimizationMessage(paymentstypes.MsgSetLiquidityLimits{
		Signer:	alice,
		Limits: paymentstypes.LiquidityLimits{
			ChannelID:		channel.ChannelID,
			Participant:		alice,
			MaxReservedCapacity:	"350",
			MinAvailableCapacity:	"50",
			MaxBaseFee:		"5",
			MaxReservationFee:	"3",
			MaxVirtualSetupFee:	"8",
			MaxRebalanceLoad:	8,
		},
		CurrentHeight:	12,
	})
	require.NoError(t, err)
	require.Len(t, liquidity.Limits, 1)

	liquidity, err = k.HandleLiquidityOptimizationMessage(paymentstypes.MsgAdvertiseLiquidity{
		Signer:	alice,
		Advertisement: paymentstypes.LiquidityAdvertisement{
			ChannelID:		channel.ChannelID,
			Advertiser:		alice,
			Counterparty:		bob,
			Capacity:		"300",
			FeeDenom:		paymentstypes.NativeDenom,
			BaseFee:		"2",
			ReservationFee:		"3",
			VirtualSetupFee:	"5",
			ReliabilityBps:		9_000,
			ValidUntilHeight:	80,
			DepositAmount:		"9",
			BackedByReservation:	true,
		},
		RequiredDeposit:	"5",
		CurrentHeight:		13,
	})
	require.NoError(t, err)
	require.Len(t, liquidity.Positions, 1)
	require.Len(t, liquidity.Forecasts, 1)
	require.Len(t, liquidity.Scores, 1)

	reservation, err := paymentstypes.BuildSignedLiquidityReservation(paymentstypes.SignedLiquidityReservation{
		AdvertisementID:	liquidity.Positions[0].FeePolicyID,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Reserver:		alice,
		Counterparty:		bob,
		Capacity:		"200",
		FeeAmount:		"3",
		ExpirationHeight:	30,
		Nonce:			1,
	}, alice)
	require.NoError(t, err)
	liquidity, err = k.HandleLiquidityOptimizationMessage(paymentstypes.MsgReserveLiquidity{
		Reservation:	reservation,
		CurrentHeight:	14,
	})
	require.NoError(t, err)
	require.Equal(t, "200", liquidity.Positions[0].ReservedCapacity)
	require.Equal(t, "100", liquidity.Positions[0].AvailableCapacity)

	forecast, err := k.CapacityForecast(channel.ChannelID, alice, bob, 15, 10)
	require.NoError(t, err)
	require.Equal(t, "200", forecast.ReservedCapacity)

	expired, err := k.ExpireLiquidityReservations(31)
	require.NoError(t, err)
	require.Len(t, expired, 1)
	liquidity, err = k.LiquidityOptimizationState()
	require.NoError(t, err)
	require.True(t, liquidity.Reservations[0].Released)
	require.Equal(t, "0", liquidity.Positions[0].ReservedCapacity)

	before := liquidity.Scores[0].Score
	require.NoError(t, k.DecayLiquidityScores(13+paymentstypes.DefaultGossipTTL, paymentstypes.DefaultGossipTTL))
	liquidity, err = k.LiquidityOptimizationState()
	require.NoError(t, err)
	require.Less(t, liquidity.Scores[0].Score, before)

	exported := k.ExportGenesis()
	require.Len(t, exported.Liquidity.Positions, 1)
	require.Len(t, exported.Liquidity.Reservations, 1)
}

func TestKeeperFraudProofVerificationModuleRecordsRewardsAndDedup(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x7b)
	bob := keeperAddress(0x7c)
	channel := keeperSignedChannel(t, "keeper-fraud-proof-verification", "100", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	closeState := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, closeState, alice, 20, "0"))
	conflicting := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	proof := paymentstypes.FraudProof{
		ProofID:		paymentstypes.HashParts("keeper-fraud-proof", channel.ChannelID),
		ProofType:		paymentstypes.FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		paymentstypes.HashParts("keeper-fraud-proof-evidence", closeState.StateHash, conflicting.StateHash),
		PenaltyDenom:		paymentstypes.NativeDenom,
		PenaltyAmount:		"20",
	}
	fraud, err := k.HandleFraudProofVerificationMessage(paymentstypes.MsgSubmitDoubleSignProof{Input: paymentstypes.FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof:		proof,
		CurrentHeight:	21,
		Policy:		paymentstypes.FraudPenaltyPolicy{ReporterRewardCap: "4", BurnShareBps: paymentstypes.MaxPenaltyRouteBps},
		GasLimit:	100_000_000,
	}})
	require.NoError(t, err)
	require.Len(t, fraud.EvidenceRecords, 1)
	require.Len(t, fraud.PenaltyRecords, 1)
	require.Len(t, fraud.ReporterRewards, 1)
	require.Len(t, fraud.DoubleSignEvidence, 1)
	debug, err := k.QueryStateHash(channel.ChannelID)
	require.NoError(t, err)
	require.Equal(t, paymentstypes.ChannelStatusPendingClose, debug.Status)

	duplicate := proof
	duplicate.ProofID = paymentstypes.HashParts("keeper-fraud-proof-duplicate", channel.ChannelID)
	duplicate.StateA = conflicting
	duplicate.StateB = closeState
	duplicate.EvidenceHash = paymentstypes.HashParts("keeper-fraud-proof-duplicate-evidence", conflicting.StateHash, closeState.StateHash)
	_, err = k.HandleFraudProofVerificationMessage(paymentstypes.MsgSubmitDoubleSignProof{Input: paymentstypes.FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof:		duplicate,
		CurrentHeight:	22,
		Policy:		paymentstypes.FraudPenaltyPolicy{ReporterRewardCap: "4", BurnShareBps: paymentstypes.MaxPenaltyRouteBps},
		GasLimit:	100_000_000,
	}})
	require.ErrorContains(t, err, "duplicate fraud evidence")

	fraud, err = k.HandleFraudProofVerificationMessage(paymentstypes.MsgClaimReporterReward{
		RewardID:	fraud.ReporterRewards[0].RewardID,
		Reporter:	bob,
		CurrentHeight:	23,
	})
	require.NoError(t, err)
	require.True(t, fraud.ReporterRewards[0].Claimed)

	exported := k.ExportGenesis()
	require.Len(t, exported.FraudProofs.EvidenceRecords, 1)
	require.True(t, exported.FraudProofs.ReporterRewards[0].Claimed)
}

func TestKeeperValidatorAssistedWatchDispute(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	alice := keeperAddress(0x67)
	bob := keeperAddress(0x68)
	validator := keeperAddress(0x69)
	service := keeperAddress(0x6a)
	channel := keeperSignedChannel(t, "keeper-validator-watch", "100", alice, bob)
	require.NoError(t, k.OpenChannel(channel))
	require.NoError(t, k.RegisterValidatorPaymentService(paymentstypes.ValidatorPaymentServiceMetadata{
		ValidatorAddress:	validator,
		ServiceAddress:		service,
		WatchEndpoint:		"https://validator.example/watch",
		MinDelegation:		"10",
		Active:			true,
		UpdatedHeight:		10,
	}))
	require.NoError(t, k.RegisterValidatorWatchService(paymentstypes.ValidatorWatchRegistration{
		ValidatorAddress:	validator,
		Delegator:		bob,
		RegisteredHeight:	11,
	}))
	stale := keeperSignedState(t, channel, 2, channel.OpeningStateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	newer := keeperSignedState(t, channel, 3, stale.StateHash, []paymentstypes.Balance{
		{Participant: alice, Amount: "45"},
		{Participant: bob, Amount: "55"},
	})
	require.NoError(t, k.SubmitClose(channel.ChannelID, stale, alice, 20, "0"))
	require.NoError(t, k.SubmitValidatorAssistedDispute(paymentstypes.ValidatorAssistedDisputeSubmission{
		ValidatorAddress:	validator,
		ServiceAddress:		service,
		Delegator:		bob,
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	stale.StateHash,
		NewerState:		newer,
		CurrentHeight:		21,
		EvidenceHash:		paymentstypes.HashParts("keeper-validator-watch", channel.ChannelID, newer.StateHash),
	}))
	debug, err := k.QueryStateHash(channel.ChannelID)
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, debug.PendingStateHash)
}

func keeperSignedChannel(t *testing.T, salt, collateral, left, right string) paymentstypes.ChannelRecord {
	t.Helper()

	channel := paymentstypes.ChannelRecord{
		ChainID:		"aetra-test-1",
		ChannelID:		paymentstypes.HashParts(salt, left, right),
		ChannelType:		paymentstypes.ChannelTypeBidirectional,
		Participants:		[]string{left, right},
		Denom:			paymentstypes.NativeDenom,
		Collateral:		collateral,
		OpenHeight:		10,
		CloseDelay:		8,
		DisputePeriod:		8,
		OpeningFeePaid:		paymentstypes.DefaultOpeningFee,
		ConditionalPayments:	true,
		CustodyDenom:		paymentstypes.NativeDenom,
		CustodyAmount:		collateral,
		Status:			paymentstypes.ChannelStatusOpen,
	}
	openState := keeperSignedState(t, channel, 1, "", []paymentstypes.Balance{
		{Participant: left, Amount: collateral},
		{Participant: right, Amount: "0"},
	})
	channel.LatestState = openState
	channel.OpeningStateHash = openState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func keeperSignedState(t *testing.T, channel paymentstypes.ChannelRecord, nonce uint64, previous string, balances []paymentstypes.Balance) paymentstypes.ChannelState {
	t.Helper()

	state, err := paymentstypes.BuildState(paymentstypes.ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		paymentstypes.CurrentStateVersion,
		Epoch:			1,
		Nonce:			nonce,
		Balances:		balances,
		PreviousStateHash:	previous,
		TimeoutHeight:		channel.OpenHeight + channel.DisputePeriod + nonce,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		paymentstypes.NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := paymentstypes.SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func keeperSignedReserveState(t *testing.T, channel paymentstypes.ChannelRecord, nonce uint64, previous, reserveA, reserveB string, balances []paymentstypes.Balance) paymentstypes.ChannelState {
	t.Helper()

	state, err := paymentstypes.BuildState(paymentstypes.ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		paymentstypes.CurrentStateVersion,
		Epoch:			1,
		Nonce:			nonce,
		Balances:		balances,
		ReserveA:		reserveA,
		ReserveB:		reserveB,
		PreviousStateHash:	previous,
		TimeoutHeight:		channel.OpenHeight + channel.DisputePeriod + 70,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		paymentstypes.NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := paymentstypes.SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func keeperSignedPromiseWithHashLock(t *testing.T, channel paymentstypes.ChannelRecord, salt, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock string) paymentstypes.ConditionalPromise {
	t.Helper()

	promise, err := paymentstypes.BuildConditionalPromise(paymentstypes.ConditionalPromise{
		PromiseID:		paymentstypes.HashParts("keeper-promise", channel.ChannelID, salt),
		ChannelID:		channel.ChannelID,
		Source:			source,
		Destination:		destination,
		Amount:			amount,
		Fee:			fee,
		HashLock:		hashLock,
		TimeoutHeight:		timeoutHeight,
		TimeoutTimestamp:	int64(timeoutHeight * 10),
		ConditionType:		paymentstypes.ConditionTypeHashLock,
		Nonce:			nonce,
	})
	require.NoError(t, err)
	promise.Signature, err = paymentstypes.SignatureForPromise(channel, promise, source)
	require.NoError(t, err)
	return promise.Normalize()
}

func keeperAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(keeperBytes20(fill)))
}

func keeperBytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}

func keeperAmountFor(balances []paymentstypes.Balance, participant string) string {
	for _, balance := range balances {
		if balance.Participant == participant {
			return balance.Amount
		}
	}
	return ""
}

func keeperRoutingEnvelope(t *testing.T, message paymentstypes.GossipMessage, signer string, receivedAt uint64) paymentstypes.SignedGossipEnvelope {
	t.Helper()
	envelope, err := paymentstypes.BuildRoutingGossipEnvelope(message, signer, receivedAt)
	require.NoError(t, err)
	return envelope
}
