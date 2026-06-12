package types

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestPaymentChannelCloseDisputeFraudAndSettlement(t *testing.T) {
	alice := testAddress(0x11)
	bob := testAddress(0x22)
	channel := signedChannel(t, "channel-main", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "10")
	require.NoError(t, err)

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newerState, bob, 25)
	require.NoError(t, err)

	conflicting := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	proof := FraudProof{
		ProofID:		HashParts("fraud", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			newerState,
		StateB:			conflicting,
		PenaltyAmount:		"25",
		EvidenceHash:		HashParts("evidence", newerState.StateHash, conflicting.StateHash),
	}
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 26)
	require.NoError(t, err)

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 50)
	require.NoError(t, err)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
	require.Equal(t, "315", amountFor(settlement.FinalBalances, alice))
	require.Equal(t, "675", amountFor(settlement.FinalBalances, bob))
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, newerState.Nonce, state.Channels[0].FinalizedNonce)
	require.Empty(t, state.CustodyLocks)
}

func TestPaymentAPISurfaceMessagesQueriesAndSettlementViews(t *testing.T) {
	require.NoError(t, ValidatePaymentAPISurface())
	require.ElementsMatch(t, []PaymentAPIMessageName{
		PaymentAPIMsgOpenChannel,
		PaymentAPIMsgCooperativeClose,
		PaymentAPIMsgUnilateralClose,
		PaymentAPIMsgDisputeClose,
		PaymentAPIMsgFinalizeClose,
		PaymentAPIMsgSubmitCheckpoint,
		PaymentAPIMsgRegisterPromise,
		PaymentAPIMsgResolvePromise,
		PaymentAPIMsgExpirePromise,
		PaymentAPIMsgBatchResolvePromises,
		PaymentAPIMsgOpenVirtualChannel,
		PaymentAPIMsgCloseVirtualChannel,
		PaymentAPIMsgDisputeVirtualChannel,
		PaymentAPIMsgSubmitFraudProof,
		PaymentAPIMsgRegisterRoutingAdvertisement,
	}, RequiredPaymentOnChainMessages())
	require.ElementsMatch(t, []PaymentAPIQueryName{
		PaymentAPIQueryChannel,
		PaymentAPIQueryChannelsByParticipant,
		PaymentAPIQueryPendingClose,
		PaymentAPIQueryFinalizationHeight,
		PaymentAPIQueryCondition,
		PaymentAPIQueryConditionsByChannel,
		PaymentAPIQueryVirtualChannel,
		PaymentAPIQueryChannelCapacity,
		PaymentAPIQueryFeeSchedule,
		PaymentAPIQuerySettlementTombstone,
		PaymentAPIQueryFraudProof,
		PaymentAPIQueryActiveDisputes,
		PaymentAPIQueryPendingFinalizations,
	}, RequiredPaymentQueries())
	require.ElementsMatch(t, []PaymentAPIEventName{
		PaymentAPIEventChannelOpened,
		PaymentAPIEventChannelCheckpointed,
		PaymentAPIEventChannelCloseStarted,
		PaymentAPIEventChannelDisputed,
		PaymentAPIEventChannelFinalized,
		PaymentAPIEventChannelSettled,
		PaymentAPIEventChannelPenalized,
		PaymentAPIEventPromiseRegistered,
		PaymentAPIEventPromiseResolved,
		PaymentAPIEventPromiseExpired,
		PaymentAPIEventVirtualChannelOpened,
		PaymentAPIEventVirtualChannelClosed,
		PaymentAPIEventVirtualChannelDisputed,
		PaymentAPIEventFraudProofAccepted,
		PaymentAPIEventFraudProofRejected,
		PaymentAPIEventRoutingAdvertisementRegistered,
		PaymentAPIEventSettlementFeeCharged,
	}, RequiredPaymentEvents())

	alice := testAddress(0x41)
	bob := testAddress(0x42)
	openReq := ChannelOpenRequest{
		ChainID:			"aetra-test-1",
		Participants:			[]string{alice, bob},
		InitialBalances:		[]Balance{{Participant: alice, Amount: "1000"}, {Participant: bob, Amount: "0"}},
		ChannelType:			ChannelTypeBidirectional,
		Collateral:			"1000",
		CloseDelay:			8,
		ChallengePeriod:		8,
		FeePolicyID:			NativeDenom,
		OpeningFeeDenom:		NativeDenom,
		OpeningFeePaid:			DefaultOpeningFee,
		ConditionalPaymentsSupported:	true,
		OpenHeight:			10,
	}
	state, fraud, result, err := ApplyPaymentAPISurfaceMessage(EmptyState(), EmptyFraudProofVerificationState(), MsgOpenChannel{Signer: alice, Request: openReq})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgOpenChannel, result.MsgName)
	require.Len(t, state.Channels, 1)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventChannelOpened))
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventSettlementFeeCharged))
	channel := state.Channels[0]

	foundChannel, found, err := QueryChannel(state, channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, channel.ChannelID, foundChannel.ChannelID)
	byParticipant, err := QueryChannelsByParticipant(state, alice)
	require.NoError(t, err)
	require.Len(t, byParticipant, 1)
	schedule, err := QueryFeeSchedule(state)
	require.NoError(t, err)
	require.Equal(t, NativeDenom, schedule.Denom)

	base := signedReserveState(t, channel, 2, channel.OpeningStateHash, "40", "0", []Balance{
		{Participant: alice, Amount: "940"},
		{Participant: bob, Amount: "20"},
	})
	state, err = AcceptSignedState(state, channel.ChannelID, base, 19)
	require.NoError(t, err)
	promiseChannel := channel
	promiseChannel.LatestState = base
	preimage := "api-surface-preimage"
	promise := signedPromiseWithHashLock(t, promiseChannel, "api-surface-promise", alice, bob, "20", "1", 2, 40, HashParts(preimage))
	state, fraud, result, err = ApplyPaymentAPISurfaceMessage(state, fraud, MsgRegisterPromise{
		Signer:		alice,
		ChannelID:	channel.ChannelID,
		BaseState:	base,
		Promises:	[]ConditionalPromise{promise},
		CurrentHeight:	20,
	})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgRegisterPromise, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventPromiseRegistered))
	condition, found, err := QueryCondition(state, channel.ChannelID, promise.PromiseID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, promise.Amount, condition.Amount)
	capacity, err := QueryChannelCapacity(state, EmptyLiquidityOptimizationState(), channel.ChannelID, 21)
	require.NoError(t, err)
	require.Equal(t, "20", capacity.PendingConditionAmount)
	require.Equal(t, "980", capacity.AvailableCapacity)

	state, fraud, result, err = ApplyPaymentAPISurfaceMessage(state, fraud, MsgResolvePromise{Request: PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{promise},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	30,
	}})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgResolvePromise, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventPromiseResolved))
	conditions, err := QueryConditionsByChannel(state, channel.ChannelID)
	require.NoError(t, err)
	require.Empty(t, conditions)

	channel = state.Channels[0]
	closeState := signedState(t, channel, 3, channel.LatestState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	state, fraud, result, err = ApplyPaymentAPISurfaceMessage(state, fraud, MsgUnilateralClose{Signer: alice, Request: ChannelCloseRequest{
		ChannelID:	channel.ChannelID,
		ClosingState:	closeState,
		CurrentHeight:	31,
		SettlementFee:	"0",
	}})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgUnilateralClose, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventChannelCloseStarted))
	pending, found, err := QueryPendingClose(state, channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, closeState.StateHash, pending.State.StateHash)
	height, found, err := QueryFinalizationHeight(state, channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(39), height)
	pendingFinalizations, err := QueryPendingFinalizations(state, 32)
	require.NoError(t, err)
	require.Len(t, pendingFinalizations, 1)

	newerState := signedState(t, channel, 4, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})
	state, fraud, result, err = ApplyPaymentAPISurfaceMessage(state, fraud, MsgDisputeClose{Signer: bob, Request: ChannelDisputeRequest{
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	closeState.StateHash,
		NewerState:		newerState,
		CurrentHeight:		32,
		DisputeFeePaid:		"0",
	}})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgDisputeClose, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventChannelDisputed))
	activeDisputes, err := QueryActiveDisputes(state, 33)
	require.NoError(t, err)
	require.Len(t, activeDisputes, 1)

	conflicting := signedState(t, channel, 4, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "510"},
		{Participant: bob, Amount: "490"},
	})
	proofID := HashParts("api-surface-fraud", channel.ChannelID)
	state, fraud, result, err = ApplyPaymentAPISurfaceMessage(state, fraud, MsgSubmitFraudProof{Signer: bob, Submission: FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof: FraudProof{
			ProofID:		proofID,
			ProofType:		FraudProofTypeDoubleSign,
			SubmittedBy:		bob,
			OffendingSigner:	alice,
			StateA:			newerState,
			StateB:			conflicting,
			PenaltyDenom:		NativeDenom,
			PenaltyAmount:		"20",
			EvidenceHash:		HashParts("api-surface-fraud-evidence", newerState.StateHash, conflicting.StateHash),
		},
		CurrentHeight:	33,
		Policy:		FraudPenaltyPolicy{ReporterRewardCap: "3", BurnShareBps: MaxPenaltyRouteBps},
		GasLimit:	100_000_000,
	}})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgSubmitFraudProof, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventFraudProofAccepted))
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventChannelPenalized))
	queriedProof, found, err := QueryFraudProof(state, fraud, proofID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, proofID, queriedProof.Evidence.ProofID)
	require.Equal(t, "3", queriedProof.Reward.Amount)

	state, _, result, err = ApplyPaymentAPISurfaceMessage(state, fraud, MsgFinalizeClose{Signer: alice, Request: FinalSettlementRequest{
		ChannelID:	channel.ChannelID,
		CurrentHeight:	50,
	}})
	require.NoError(t, err)
	require.Equal(t, PaymentAPIMsgFinalizeClose, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventChannelFinalized))
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventChannelSettled))
	tombstone, found, err := QuerySettlementTombstone(state, channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, channel.ChannelID, tombstone.ChannelID)
}

func TestPaymentAPISurfaceVirtualChannelMessagesAndQueries(t *testing.T) {
	alice := testAddress(0x43)
	router := testAddress(0x44)
	bob := testAddress(0x45)
	state, vc, proof := virtualChannelFixture(t, "api-surface-virtual", alice, router, bob, "100", 60)
	base := state.Clone()
	base.VirtualChannels = nil
	base.FeeCharges = nil

	state, fraud, result, err := ApplyPaymentAPISurfaceMessage(base, EmptyFraudProofVerificationState(), MsgOpenVirtualChannel{Signer: alice, ActivationProof: proof})
	require.NoError(t, err)
	require.Empty(t, fraud.EvidenceRecords)
	require.Equal(t, PaymentAPIMsgOpenVirtualChannel, result.MsgName)
	require.Contains(t, paymentEventTypes(state.Events), string(PaymentAPIEventVirtualChannelOpened))

	queried, found, err := QueryVirtualChannel(state, vc.VirtualChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, vc.VirtualChannelID, queried.VirtualChannelID)

	next, fraud, result := PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}
	next, fraud, result, err = ApplyPaymentAPISurfaceMessage(state, EmptyFraudProofVerificationState(), MsgCloseVirtualChannel{
		Signer:			alice,
		VirtualChannelID:	vc.VirtualChannelID,
		CurrentHeight:		40,
	})
	require.NoError(t, err)
	require.Empty(t, fraud.EvidenceRecords)
	require.Equal(t, PaymentAPIMsgCloseVirtualChannel, result.MsgName)
	require.Equal(t, vc.VirtualChannelID, result.VirtualChannelID)
	require.Contains(t, paymentEventTypes(next.Events), string(PaymentAPIEventVirtualChannelClosed))
	_, found, err = QueryVirtualChannel(next, vc.VirtualChannelID)
	require.NoError(t, err)
	require.False(t, found)

	rejected, err := PaymentAPIFraudProofRejectedEvent(FraudProofSubmission{
		ChannelID:	state.Channels[0].ChannelID,
		CurrentHeight:	41,
		Proof: FraudProof{
			ProofID:	HashParts("api-surface-rejected-fraud", state.Channels[0].ChannelID),
			ProofType:	FraudProofTypeDoubleSign,
		},
	}, "malformed evidence")
	require.NoError(t, err)
	require.Equal(t, string(PaymentAPIEventFraudProofRejected), rejected.EventType)
}

func TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria(t *testing.T) {
	report := BuildPaymentImplementationRoadmap()
	require.NoError(t, ValidatePaymentImplementationRoadmap(report))
	require.Equal(t, report.TotalTaskCount, report.CompletedTaskCount)
	require.Equal(t, uint64(49), report.TotalTaskCount)
	require.Equal(t, uint64(22), report.ExitCriteriaCount)

	alice := testAddress(0x46)
	router := testAddress(0x47)
	bob := testAddress(0x48)
	channel := signedChannel(t, "roadmap-base-channel", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "700"},
		{Participant: bob, Amount: "300"},
	})
	promise := signedPromiseWithHashLock(t, channel, "roadmap-promise", alice, bob, "10", "1", 3, 32, HashParts("roadmap-preimage"))
	asyncChannel := signedAsyncChannel(t, "roadmap-async-channel", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 10, 10, "50", 60, alice, bob)
	delta := signedAsyncDelta(t, asyncChannel, "roadmap-delta", alice, bob, "5", 3, 4, 35)
	_, vc, _ := virtualChannelFixture(t, "roadmap-virtual", alice, router, bob, "100", 60)

	vectors, err := BuildPaymentRoadmapCanonicalTestVectors(channel, state, promise, delta, vc)
	require.NoError(t, err)
	require.NoError(t, ValidatePaymentRoadmapCanonicalTestVectors(vectors))
	vectorTypes := make([]string, 0, len(vectors))
	for _, vector := range vectors {
		vectorTypes = append(vectorTypes, vector.ObjectType)
	}
	require.ElementsMatch(t, []string{SignatureObjectState, SignatureObjectPromise, SignatureObjectDelta, SignatureObjectVirtual}, vectorTypes)

	closeState := signedState(t, channel, 3, state.StateHash, []Balance{
		{Participant: alice, Amount: "650"},
		{Participant: bob, Amount: "350"},
	})
	conflicting := signedState(t, channel, 3, state.StateHash, []Balance{
		{Participant: alice, Amount: "640"},
		{Participant: bob, Amount: "360"},
	})
	proof := FraudProof{
		ProofID:		HashParts("roadmap-double-sign-proof", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"20",
		EvidenceHash:		HashParts("roadmap-double-sign-evidence", closeState.StateHash, conflicting.StateHash),
	}
	fraudVectors, err := BuildPaymentRoadmapFraudProofVectors(channel, []FraudProof{proof})
	require.NoError(t, err)
	require.NoError(t, ValidatePaymentRoadmapFraudProofVectors(fraudVectors))
	require.Equal(t, PenaltyClassDoubleSign, fraudVectors[0].PenaltyClass)

	upstream := signedPromiseWithHashLock(t, channel, "roadmap-upstream", alice, bob, "10", "0", 4, 58, HashParts("roadmap-route-preimage"))
	downstream := signedPromiseWithHashLock(t, channel, "roadmap-downstream", alice, bob, "10", "0", 5, 40, HashParts("roadmap-route-preimage"))
	validTimeout := BuildPaymentRoadmapTimeoutOrderingVector(channel, upstream, downstream, DefaultTimeoutMargin)
	require.NoError(t, ValidatePaymentRoadmapTimeoutVector(validTimeout, true))
	unsafeDownstream := signedPromiseWithHashLock(t, channel, "roadmap-downstream-unsafe", alice, bob, "10", "0", 6, 55, HashParts("roadmap-route-preimage"))
	invalidTimeout := BuildPaymentRoadmapTimeoutOrderingVector(channel, upstream, unsafeDownstream, DefaultTimeoutMargin)
	require.NoError(t, ValidatePaymentRoadmapTimeoutVector(invalidTimeout, false))

	conditionalVector, err := BuildPaymentRoadmapConditionalVector(BatchConditionSettlementResult{
		RouteID:	HashParts("roadmap-conditional-route", channel.ChannelID),
		EvidenceHash:	HashParts("roadmap-conditional-result", promise.PromiseID),
		Resolutions: []ConditionResolution{{
			ConditionID:	promise.PromiseID,
			Resolver:	bob,
			Recipient:	bob,
			Amount:		promise.Amount,
			EvidenceHash:	HashParts("roadmap-conditional-resolution", promise.PromiseID),
		}},
		ConditionRootUpdates: []ConditionRootUpdate{{
			ChannelID:	channel.ChannelID,
			Nonce:		state.Nonce + 1,
			ConditionRoot:	ComputeConditionsRoot(nil),
			ConditionCount:	0,
		}},
	}, ConditionSettlementModePreimage, true, ConditionClaimRecord{
		ChainID:	channel.ChainID,
		ChannelID:	channel.ChannelID,
		ConditionID:	promise.PromiseID,
		EvidenceHash:	HashParts("roadmap-preimage-replay-key", promise.PromiseID),
		ResolvedHeight:	40,
		ExpiresHeight:	40 + DefaultReplayHorizon,
	})
	require.NoError(t, err)
	require.NoError(t, conditionalVector.Validate())

	firstPlan, err := AccessPlanForSettlementOperation(SettlementOperation{
		OperationID:	HashParts("roadmap-plan-first"),
		OperationType:	BatchOperationSettle,
		ChannelID:	channel.ChannelID,
		Nonce:		3,
		StateHash:	closeState.StateHash,
	}, 50)
	require.NoError(t, err)
	secondChannel := signedChannel(t, "roadmap-second-channel", "1000", alice, router)
	secondPlan, err := AccessPlanForSettlementOperation(SettlementOperation{
		OperationID:	HashParts("roadmap-plan-second"),
		OperationType:	BatchOperationSettle,
		ChannelID:	secondChannel.ChannelID,
		Nonce:		2,
		StateHash:	secondChannel.LatestState.StateHash,
	}, 50)
	require.NoError(t, err)
	blockPlan, err := BuildPaymentRoadmapBlockSTMPlan([]BlockSTMAccessPlan{firstPlan, secondPlan})
	require.NoError(t, err)
	require.Zero(t, blockPlan.ConflictCount)
	require.True(t, blockPlan.DeferredAccounting)
	require.NotEmpty(t, blockPlan.IndependentGroups)
	require.NoError(t, ValidateHash("roadmap blockstm plan hash", blockPlan.PlanHash))

	edge := ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "1000", FeeDenom: NativeDenom, FeeAmount: "1", Active: true}
	route := ScoredRoute{
		Edges:		[]ChannelEdge{edge},
		Amount:		"100",
		TotalFee:	"1",
		TotalCost:	"101",
		MinCapacity:	"1000",
		ScoreHash:	HashParts("roadmap-route-score", channel.ChannelID),
	}
	attempt := RouteAttempt{
		AttemptID:	HashParts("roadmap-route-attempt", route.ScoreHash),
		From:		alice,
		To:		bob,
		Amount:		"100",
		CurrentHeight:	50,
		Route:		route,
		Success:	true,
	}
	failure := RouteFailureReport{
		ChannelID:	channel.ChannelID,
		From:		alice,
		To:		bob,
		FailureClass:	RouteFailureTimeout,
		Retryable:	true,
		ObservedHeight:	51,
	}
	routingVector, err := BuildPaymentRoadmapRoutingVector(route, attempt, failure)
	require.NoError(t, err)
	require.NoError(t, routingVector.Validate())

	virtualState, virtualBase, activation := virtualChannelFixture(t, "roadmap-phase5-virtual", alice, router, bob, "100", 60)
	endpointUpdate := signedVirtualEndpointUpdate(t, virtualBase, 2, "85", "15")
	virtualState, err = AcceptVirtualChannelUpdate(virtualState, endpointUpdate, 25)
	require.NoError(t, err)
	disputeState := signedVirtualEndpointUpdate(t, endpointUpdate, 3, "80", "20")
	disputeProof, err := BuildVirtualChannelDisputeProof(disputeState, virtualReserveCommitments(activation), router)
	require.NoError(t, err)
	virtualState, err = SubmitVirtualChannelDispute(virtualState, disputeProof, 30)
	require.NoError(t, err)
	opsVirtualState := virtualState
	finalVirtualState := signedVirtualEndpointUpdate(t, disputeState, 4, "75", "25")
	closeProof, err := BuildVirtualCloseProof(finalVirtualState, VirtualCloseModeDisputed, virtualReserveCommitments(activation), router, 31)
	require.NoError(t, err)
	_, _, releases, err := CloseVirtualChannelWithProof(virtualState, closeProof, 31)
	require.NoError(t, err)
	virtualVector, err := BuildPaymentRoadmapVirtualVector(virtualBase, activation, closeProof, disputeProof, releases, endpointUpdate)
	require.NoError(t, err)
	require.NoError(t, virtualVector.Validate())
	require.Len(t, virtualVector.ParentChannelIDs, 2)
	require.Len(t, virtualVector.ReserveReleaseHashes, 2)

	cleanupChannel := signedChannel(t, "roadmap-cleanup-channel", "1000", alice, bob)
	cleanupBase := signedReserveState(t, cleanupChannel, 2, cleanupChannel.OpeningStateHash, "20", "0", []Balance{
		{Participant: alice, Amount: "970"},
		{Participant: bob, Amount: "10"},
	})
	cleanupPromiseChannel := cleanupChannel
	cleanupPromiseChannel.LatestState = cleanupBase
	cleanupPromise := signedLinkedPromise(t, cleanupPromiseChannel, HashParts("roadmap-cleanup-promise", cleanupChannel.ChannelID), alice, bob, "10", "0", 3, 40, HashParts("roadmap-cleanup-preimage"), "", "")
	cleanupConditioned, _, err := BuildConditionRootUpdateFromPromises(cleanupPromiseChannel, cleanupBase, []ConditionalPromise{cleanupPromise}, nil)
	require.NoError(t, err)
	cleanupConditioned = resignState(t, cleanupChannel, cleanupConditioned)
	cleanupState := EmptyState()
	cleanupState, err = OpenChannel(cleanupState, cleanupChannel)
	require.NoError(t, err)
	cleanupState, err = AcceptSignedState(cleanupState, cleanupChannel.ChannelID, cleanupConditioned, 20)
	require.NoError(t, err)
	cleanupState, _, err = EnqueueExpiredPromise(cleanupState, cleanupPromise, alice, 21)
	require.NoError(t, err)
	_, cleanupResult, err := ProcessAsyncExecutionQueues(cleanupState, 41, 0, 1)
	require.NoError(t, err)
	require.NotEmpty(t, cleanupResult.EmittedCompletionIDs)

	layout, err := BuildStoreV2Layout(opsVirtualState)
	require.NoError(t, err)
	require.NoError(t, layout.Validate())
	snapshot, err := BuildAdaptiveSyncSnapshot(opsVirtualState, 50)
	require.NoError(t, err)
	recovery, err := RecoverAdaptiveSyncSafety(snapshot)
	require.NoError(t, err)
	require.Contains(t, recovery.VirtualChannelIDs, virtualBase.VirtualChannelID)
	require.NotEmpty(t, recovery.WatcherReplayEventIDs)
	opsVector, err := BuildPaymentRoadmapOperationsVector(layout, blockPlan, snapshot, recovery, cleanupResult)
	require.NoError(t, err)
	require.NoError(t, opsVector.Validate())
	require.Equal(t, uint64(len(snapshot.WatcherReplayEvents)), opsVector.WatcherReplayCount)
}

func TestPaymentEngineeringBacklogTracksSection19Priorities(t *testing.T) {
	report := BuildPaymentEngineeringBacklog()
	require.NoError(t, ValidatePaymentEngineeringBacklog(report))
	require.Equal(t, uint64(9), report.HighPriorityCount)
	require.Equal(t, uint64(8), report.MediumPriorityCount)
	require.Equal(t, uint64(6), report.LowerPriorityCount)
	require.Equal(t, uint64(23), report.CompleteCount)
	require.Equal(t, uint64(1), report.LocalOnlyCount)
	require.Len(t, report.Items, 23)

	seen := map[PaymentEngineeringBacklogItemID]PaymentEngineeringBacklogItem{}
	for _, item := range report.Items {
		require.Equal(t, PaymentBacklogStatusComplete, item.Status)
		require.NotEmpty(t, item.Evidence)
		require.NoError(t, ValidateHash("payments engineering backlog item hash", item.ItemHash))
		seen[item.ItemID] = item
	}
	localDoc := seen["high_local_payments_doc"]
	require.True(t, localDoc.LocalOnly)
	require.Contains(t, localDoc.Evidence, ".git/info/exclude:/PAYMENTS.md")
	require.Contains(t, localDoc.Evidence, "PAYMENTS.md")
	require.Contains(t, seen["high_blockstm_settlement_analysis"].Evidence, "ProfileBlockSTMConflicts")
	require.Contains(t, seen["medium_capacity_aware_path_search"].Evidence, "SelectPaymentRoute")
	require.Contains(t, seen["lower_route_privacy_packetization"].Evidence, "ForwardingPacket")

	duplicate := report
	duplicate.Items = append(duplicate.Items, report.Items[0])
	duplicate.ReportHash = ComputePaymentEngineeringBacklogReportHash(duplicate)
	require.ErrorContains(t, ValidatePaymentEngineeringBacklog(duplicate), "duplicate payments engineering backlog item")

	missing := report
	missing.Items = missing.Items[:len(missing.Items)-1]
	missing.LowerPriorityCount--
	missing.CompleteCount--
	missing.ReportHash = ComputePaymentEngineeringBacklogReportHash(missing)
	require.ErrorContains(t, ValidatePaymentEngineeringBacklog(missing), "missing payments engineering backlog item")
}

func TestPaymentAcceptanceCriteriaCoverInitialProductionHardening(t *testing.T) {
	report := BuildPaymentAcceptanceReport()
	require.NoError(t, ValidatePaymentAcceptanceReport(report))
	require.Equal(t, uint64(14), report.CriterionCount)
	require.Equal(t, uint64(14), report.SatisfiedCount)
	require.Equal(t, uint64(9), report.DomainCount)
	require.Len(t, report.Criteria, 14)

	seen := map[PaymentAcceptanceCriterionID]PaymentAcceptanceCriterion{}
	domains := map[PaymentAcceptanceDomain]bool{}
	for _, criterion := range report.Criteria {
		require.Equal(t, PaymentAcceptanceStatusSatisfied, criterion.Status)
		require.NotEmpty(t, criterion.Evidence)
		require.NotEmpty(t, criterion.TestNames)
		require.NoError(t, ValidateHash("payments acceptance criterion hash", criterion.ItemHash))
		seen[criterion.CriterionID] = criterion
		domains[criterion.Domain] = true
	}
	for _, domain := range []PaymentAcceptanceDomain{
		PaymentAcceptanceSettlement,
		PaymentAcceptanceFraud,
		PaymentAcceptanceConditional,
		PaymentAcceptanceVirtual,
		PaymentAcceptanceExecution,
		PaymentAcceptanceRecovery,
		PaymentAcceptanceEconomics,
		PaymentAcceptanceSecurity,
		PaymentAcceptanceObservability,
	} {
		require.True(t, domains[domain], "missing acceptance domain %s", domain)
	}
	require.Contains(t, seen["accept_any_participant_unilateral_close"].Evidence, "MsgUnilateralClose")
	require.Contains(t, seen["accept_fraud_proofs_deterministic_bounded_tested"].Evidence, "MeterFraudProofVerification")
	require.Contains(t, seen["accept_observability_liquidity_settlement_dispute_fee_perf"].TestNames, "TestPaymentObservabilityMetricsCoverOperationalSignals")

	duplicate := report
	duplicate.Criteria = append(duplicate.Criteria, report.Criteria[0])
	duplicate.ReportHash = ComputePaymentAcceptanceReportHash(duplicate)
	require.ErrorContains(t, ValidatePaymentAcceptanceReport(duplicate), "duplicate payments acceptance criterion")

	missing := report
	missing.Criteria = missing.Criteria[:len(missing.Criteria)-1]
	missing.CriterionCount--
	missing.SatisfiedCount--
	missing.ReportHash = ComputePaymentAcceptanceReportHash(missing)
	require.ErrorContains(t, ValidatePaymentAcceptanceReport(missing), "missing payments acceptance criterion")
}

func TestRequiredPaymentTestCoverageMatrixCoversUnitAndIntegrationSpecs(t *testing.T) {
	report := BuildRequiredTestCoverageReport()
	require.NoError(t, ValidateRequiredTestCoverageReport(report))
	require.Equal(t, uint64(14), report.UnitCount)
	require.Equal(t, uint64(9), report.IntegrationCount)
	require.Equal(t, uint64(10), report.InvariantCount)
	require.Equal(t, uint64(9), report.FuzzCount)
	require.Equal(t, uint64(9), report.PerformanceCount)
	require.Len(t, report.Entries, 51)

	seen := map[RequiredTestCoverageID]RequiredTestCoverageEntry{}
	for _, entry := range report.Entries {
		require.NotEmpty(t, entry.TestNames)
		require.NotEmpty(t, entry.Evidence)
		require.NoError(t, ValidateHash("coverage entry evidence", entry.EvidenceHash))
		seen[entry.CoverageID] = entry
	}
	require.Equal(t, RequiredTestCoverageUnit, seen["unit_channel_id_generation"].Kind)
	require.Contains(t, seen["unit_channel_id_generation"].Evidence, "HashParts")
	require.Equal(t, RequiredTestCoverageIntegration, seen["integration_parent_dispute_with_virtual_active"].Kind)
	require.Contains(t, seen["integration_parent_dispute_with_virtual_active"].TestNames, "TestParentChannelDisputeWhileVirtualChannelIsActive")
	require.Equal(t, RequiredTestCoverageInvariant, seen["invariant_same_channel_writes_conflict"].Kind)
	require.Contains(t, seen["invariant_same_channel_writes_conflict"].TestNames, "TestBlockSTMConflictProfileDetectsSameChannelConflicts")
	require.Equal(t, RequiredTestCoverageFuzz, seen["fuzz_async_delta_aggregation"].Kind)
	require.Contains(t, seen["fuzz_async_delta_aggregation"].TestNames, "FuzzPaymentRequiredFuzzVectors")
	require.Equal(t, RequiredTestCoveragePerformance, seen["performance_blockstm_conflict_rate_mix"].Kind)
	require.Contains(t, seen["performance_blockstm_conflict_rate_mix"].TestNames, "TestPaymentPerformanceCoverageProfilesPerBlockWorkloads")

	duplicate := report
	duplicate.Entries = append(duplicate.Entries, report.Entries[0])
	duplicate.ReportHash = ComputeRequiredTestCoverageReportHash(duplicate)
	require.ErrorContains(t, ValidateRequiredTestCoverageReport(duplicate), "duplicate payments required test coverage")

	missing := report
	missing.Entries = missing.Entries[:len(missing.Entries)-1]
	missing.UnitCount = 14
	missing.IntegrationCount = 9
	missing.InvariantCount = 10
	missing.FuzzCount = 9
	missing.PerformanceCount = 8
	missing.ReportHash = ComputeRequiredTestCoverageReportHash(missing)
	require.ErrorContains(t, ValidateRequiredTestCoverageReport(missing), "missing payments required test coverage")
}

func TestPaymentPerformanceCoverageProfilesPerBlockWorkloads(t *testing.T) {
	alice := testAddress(0xba)
	router := testAddress(0xbb)
	bob := testAddress(0xbc)
	const blockHeight uint64 = 64
	const opsPerBlock = 4

	openMessages := make([]PaymentChannelModuleMessage, 0, opsPerBlock)
	for i := 0; i < opsPerBlock; i++ {
		opener := testAddress(byte(0xc0 + i*2))
		counterparty := testAddress(byte(0xc1 + i*2))
		openMessages = append(openMessages, MsgOpenChannel{
			Signer:	opener,
			Request: ChannelOpenRequest{
				ChainID:		"aetra-test-1",
				Participants:		[]string{opener, counterparty},
				InitialBalances:	[]Balance{{Participant: opener, Amount: "100"}, {Participant: counterparty, Amount: "0"}},
				ChannelType:		ChannelTypeBidirectional,
				Collateral:		"100",
				CloseDelay:		8,
				ChallengePeriod:	8,
				FeePolicyID:		NativeDenom,
				OpeningFeeDenom:	NativeDenom,
				OpeningFeePaid:		DefaultOpeningFee,
				OpenHeight:		10 + uint64(i),
			},
		})
	}
	openProfile, err := PaymentChannelMessagesConflictProfile(openMessages, blockHeight)
	require.NoError(t, err)
	require.True(t, openProfile.ConflictFree)
	require.Len(t, openProfile.Plans, opsPerBlock)

	state := EmptyState()
	closeMessages := make([]PaymentChannelModuleMessage, 0, opsPerBlock)
	disputeMessages := make([]PaymentChannelModuleMessage, 0, opsPerBlock)
	conditionPlans := make([]BlockSTMAccessPlan, 0, opsPerBlock)
	for i := 0; i < opsPerBlock; i++ {
		channel := signedChannel(t, fmt.Sprintf("perf-channel-%d", i), "1000", alice, bob)
		var openErr error
		state, openErr = OpenChannel(state, channel)
		require.NoError(t, openErr)
		closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "900"}, {Participant: bob, Amount: "100"}})
		newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{{Participant: alice, Amount: "850"}, {Participant: bob, Amount: "150"}})
		closeMessages = append(closeMessages, MsgUnilateralClose{Signer: alice, Request: ChannelCloseRequest{ChannelID: channel.ChannelID, ClosingState: closeState, Submitter: alice, CurrentHeight: blockHeight, SettlementFee: "0"}})
		disputeMessages = append(disputeMessages, MsgDisputeClose{Signer: bob, Request: ChannelDisputeRequest{ChannelID: channel.ChannelID, ClosingStateReference: closeState.StateHash, NewerState: newerState, Submitter: bob, CurrentHeight: blockHeight + 1}})
		conditionPlan, planErr := AccessPlanForConditionResolution(channel.ChannelID, []string{HashParts("perf-condition", channel.ChannelID)}, blockHeight)
		require.NoError(t, planErr)
		conditionPlans = append(conditionPlans, conditionPlan)
	}
	closeProfile, err := PaymentChannelMessagesConflictProfile(closeMessages, blockHeight)
	require.NoError(t, err)
	require.True(t, closeProfile.ConflictFree)
	require.Len(t, closeProfile.Plans, opsPerBlock)
	disputeProfile, err := PaymentChannelMessagesConflictProfile(disputeMessages, blockHeight+1)
	require.NoError(t, err)
	require.True(t, disputeProfile.ConflictFree)
	require.Len(t, disputeProfile.Plans, opsPerBlock)
	conditionProfile := ProfileBlockSTMConflicts(conditionPlans)
	require.True(t, conditionProfile.ConflictFree)
	require.Len(t, conditionProfile.Plans, opsPerBlock)

	cooperativeSettles := 0
	coopState := EmptyState()
	for i := 0; i < opsPerBlock; i++ {
		channel := signedChannel(t, fmt.Sprintf("perf-coop-%d", i), "1000", alice, bob)
		var err error
		coopState, err = OpenChannel(coopState, channel)
		require.NoError(t, err)
		closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "500"}, {Participant: bob, Amount: "500"}})
		coopState, _, err = CooperativeClose(coopState, channel.ChannelID, closeState, alice, blockHeight, "0")
		require.NoError(t, err)
		cooperativeSettles++
	}
	require.Equal(t, opsPerBlock, cooperativeSettles)

	virtualState, _, activation := virtualChannelFixture(t, "perf-virtual", alice, router, bob, "100", 40)
	virtualDisputes := 0
	virtualDisputeHeight := uint64(32)
	for i := 0; i < opsPerBlock; i++ {
		currentVirtual := virtualState.VirtualChannels[0]
		update := signedVirtualEndpointUpdate(t, currentVirtual, currentVirtual.Nonce+uint64(i+1), fmt.Sprintf("%d", 90-i), fmt.Sprintf("%d", 10+i))
		proof, err := BuildVirtualChannelDisputeProof(update, virtualReserveCommitments(activation), router)
		require.NoError(t, err)
		if i == 0 {
			virtualState, err = SubmitVirtualChannelDispute(virtualState, proof, virtualDisputeHeight)
			require.NoError(t, err)
		} else {
			require.NoError(t, ValidateVirtualChannelDisputeProof(proof, virtualState.VirtualChannels[0]))
		}
		virtualDisputes++
	}
	require.Equal(t, opsPerBlock, virtualDisputes)

	mixedPlans := append([]BlockSTMAccessPlan{}, closeProfile.Plans...)
	mixedPlans = append(mixedPlans, disputeProfile.Plans...)
	mixedProfile := ProfileBlockSTMConflicts(mixedPlans)
	require.False(t, mixedProfile.ConflictFree)
	require.NotEmpty(t, mixedProfile.Conflicts)
	conflictRateBps := uint64(len(mixedProfile.Conflicts)) * 10_000 / uint64(len(mixedPlans)*(len(mixedPlans)-1)/2)
	require.Greater(t, conflictRateBps, uint64(0))
	require.LessOrEqual(t, conflictRateBps, uint64(10_000))

	layout, err := BuildStoreV2Layout(state)
	require.NoError(t, err)
	require.Len(t, layout.Channels, opsPerBlock)
	require.Len(t, layout.ParticipantChannels, opsPerBlock*2)
	page, err := QueryStoreV2ParticipantChannels(layout, ParticipantChannelPageRequest{Address: alice, Limit: opsPerBlock})
	require.NoError(t, err)
	require.Equal(t, uint64(opsPerBlock), page.Total)
	storeIndexFootprint := len(layout.Channels) + len(layout.ChannelStates) + len(layout.ParticipantChannels)
	require.GreaterOrEqual(t, storeIndexFootprint, opsPerBlock*4)

	disputeState := EmptyState()
	disputeChannel := signedChannel(t, "perf-snapshot-dispute", "1000", alice, bob)
	disputeState, err = OpenChannel(disputeState, disputeChannel)
	require.NoError(t, err)
	closeState := signedConditionalState(t, disputeChannel, 2, disputeChannel.OpeningStateHash, "25", []Balance{{Participant: alice, Amount: "975"}, {Participant: bob, Amount: "0"}})
	disputeState, err = SubmitClose(disputeState, disputeChannel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	newer := signedConditionalState(t, disputeChannel, 3, closeState.StateHash, "25", []Balance{{Participant: alice, Amount: "950"}, {Participant: bob, Amount: "25"}})
	disputeState, err = DisputeClose(disputeState, disputeChannel.ChannelID, newer, bob, 21)
	require.NoError(t, err)
	snapshot, err := BuildAdaptiveSyncSnapshot(disputeState, 22)
	require.NoError(t, err)
	recovered, err := RecoverAdaptiveSyncSafety(snapshot)
	require.NoError(t, err)
	require.Len(t, snapshot.ActiveDisputes, 1)
	require.Contains(t, recovered.ActiveDisputeChannelIDs, disputeChannel.ChannelID)
	require.NotEmpty(t, recovered.WatcherReplayEventIDs)
}

func TestPaymentObservabilityMetricsCoverOperationalSignals(t *testing.T) {
	alice := testAddress(0xd0)
	router := testAddress(0xd1)
	bob := testAddress(0xd2)
	currentHeight := uint64(29)

	state, vc, _ := virtualChannelFixture(t, "observability-virtual", alice, router, bob, "100", 40)
	conditionChannel, found := state.ChannelByID(vc.ParentChannelIDs[0])
	require.True(t, found)
	conditionChannel.RoutingAdvertised = true
	for i := range state.Channels {
		if state.Channels[i].ChannelID == conditionChannel.ChannelID {
			state.Channels[i] = conditionChannel
			break
		}
	}
	activePromise := signedPromiseWithHashLock(t, conditionChannel, "observability-active", alice, router, "5", "0", 3, 40, HashParts("observability-active-preimage"))
	conditioned, _, err := BuildConditionRootUpdateFromPromises(conditionChannel, conditionChannel.LatestState, []ConditionalPromise{activePromise}, nil)
	require.NoError(t, err)
	conditioned.Nonce = conditionChannel.LatestState.Nonce + 1
	conditioned.PreviousStateHash = conditionChannel.LatestState.StateHash
	conditioned, err = BuildState(conditioned)
	require.NoError(t, err)
	conditioned = resignState(t, conditionChannel, conditioned)
	state, err = AcceptSignedState(state, conditionChannel.ChannelID, conditioned, 24)
	require.NoError(t, err)

	disputeChannel := signedChannel(t, "observability-dispute", "1000", alice, bob)
	state, err = OpenChannel(state, disputeChannel)
	require.NoError(t, err)
	disputeClose := signedState(t, disputeChannel, 2, disputeChannel.OpeningStateHash, []Balance{{Participant: alice, Amount: "900"}, {Participant: bob, Amount: "100"}})
	state, err = SubmitClose(state, disputeChannel.ChannelID, disputeClose, alice, 20, "0")
	require.NoError(t, err)
	newer := signedState(t, disputeChannel, 3, disputeClose.StateHash, []Balance{{Participant: alice, Amount: "850"}, {Participant: bob, Amount: "150"}})
	state, err = DisputeClose(state, disputeChannel.ChannelID, newer, bob, 21)
	require.NoError(t, err)

	settledChannel := signedChannel(t, "observability-settled", "1000", alice, bob)
	state, err = OpenChannel(state, settledChannel)
	require.NoError(t, err)
	settledClose := signedState(t, settledChannel, 2, settledChannel.OpeningStateHash, []Balance{{Participant: alice, Amount: "500"}, {Participant: bob, Amount: "500"}})
	state, _, err = CooperativeClose(state, settledChannel.ChannelID, settledClose, alice, currentHeight, "2")
	require.NoError(t, err)
	state, _, err = RecordSettlementInclusionLatency(state, HashParts("observability-latency"), disputeChannel.ChannelID, SettlementArbitrationDispute, 20, currentHeight, 4)
	require.NoError(t, err)

	state.ConditionClaims = append(state.ConditionClaims,
		ConditionClaimRecord{ChainID: conditionChannel.ChainID, ChannelID: conditionChannel.ChannelID, ConditionID: HashParts("observability-resolved"), EvidenceHash: HashParts("promise-preimage", "observability-resolved"), PreimageHash: HashParts("observability-preimage"), ResolvedHeight: 25, ExpiresHeight: 25 + DefaultReplayHorizon}.Normalize(),
		ConditionClaimRecord{ChainID: conditionChannel.ChainID, ChannelID: conditionChannel.ChannelID, ConditionID: HashParts("observability-expired"), EvidenceHash: HashParts("promise-expiry", "observability-expired"), ResolvedHeight: 26, ExpiresHeight: 26 + DefaultReplayHorizon}.Normalize(),
	)
	rejected, err := PaymentAPIFraudProofRejectedEvent(FraudProofSubmission{
		ChannelID:	disputeChannel.ChannelID,
		CurrentHeight:	27,
		Proof:		FraudProof{ProofID: HashParts("observability-rejected-proof"), ProofType: FraudProofTypeDoubleSign},
	}, "observability rejected")
	require.NoError(t, err)
	state.Events = append(state.Events, rejected)
	state, err = RefreshAsyncExecutionQueues(state, currentHeight)
	require.NoError(t, err)
	state, _, err = EnqueueExpiredPromise(state, activePromise, alice, 28)
	require.NoError(t, err)
	state.FeeCharges = append(state.FeeCharges, PaymentFeeCharge{
		FeeID:		HashParts("observability-routing-fee"),
		FeeClass:	PaymentFeeClassRoutingAdvertisement,
		ChannelID:	conditionChannel.ChannelID,
		ObjectID:	HashParts("observability-routing-advertisement"),
		Payer:		alice,
		Denom:		NativeDenom,
		Amount:		"7",
		RequiredAmount:	"7",
		StorageBytes:	128,
		MultiplierBps:	10_000,
		Height:		28,
	}.Normalize())
	sortPaymentFeeCharges(state.FeeCharges)
	require.NoError(t, state.Validate())

	evidenceID := HashParts("observability-evidence")
	proofID := HashParts("observability-proof")
	canonical := HashParts("observability-canonical")
	evidence := EvidenceRecord{EvidenceID: evidenceID, ChannelID: disputeChannel.ChannelID, ProofID: proofID, ProofType: FraudProofTypeDoubleSign, CanonicalHash: canonical, SubmittedBy: bob, OffendingSigner: alice, SubmittedHeight: 21, ExpiresHeight: 40, GasUsed: 10}.WithHash()
	penalty := PenaltyRecord{PenaltyID: HashParts("observability-penalty"), EvidenceID: evidenceID, ChannelID: disputeChannel.ChannelID, ProofID: proofID, Offender: alice, TotalPenalty: "10", ReporterReward: "3", CounterpartyComp: "2", RecordedHeight: 22}.WithHash()
	reward := ReporterReward{RewardID: HashParts("observability-reward"), EvidenceID: evidenceID, ChannelID: disputeChannel.ChannelID, ProofID: proofID, Reporter: bob, Denom: NativeDenom, Amount: "3", Claimed: true, ClaimedHeight: 23}.WithHash()
	fraud := FraudProofVerificationState{EvidenceRecords: []EvidenceRecord{evidence}, PenaltyRecords: []PenaltyRecord{penalty}, ReporterRewards: []ReporterReward{reward}}
	require.NoError(t, fraud.Validate())

	closeOp := SettlementOperation{OperationID: HashParts("observability-close-op"), OperationType: BatchOperationClose, ChannelID: disputeChannel.ChannelID, Nonce: 2, StateHash: disputeClose.StateHash}
	disputeOp := SettlementOperation{OperationID: HashParts("observability-dispute-op"), OperationType: BatchOperationDispute, ChannelID: disputeChannel.ChannelID, Nonce: 3, StateHash: newer.StateHash}
	closePlan, err := AccessPlanForSettlementOperation(closeOp, currentHeight)
	require.NoError(t, err)
	disputePlan, err := AccessPlanForSettlementOperation(disputeOp, currentHeight)
	require.NoError(t, err)
	profile := ProfileBlockSTMConflicts([]BlockSTMAccessPlan{closePlan, disputePlan})
	require.False(t, profile.ConflictFree)

	metrics, err := BuildPaymentObservabilityMetrics(state, fraud, profile, currentHeight, 1)
	require.NoError(t, err)
	require.NoError(t, metrics.Validate())
	require.Equal(t, uint64(2), metrics.ActiveChannels)
	require.Equal(t, uint64(1), metrics.PendingCloses)
	require.Equal(t, uint64(1), metrics.ActiveDisputes)
	require.Equal(t, uint64(1), metrics.FinalizableChannels)
	require.Equal(t, uint64(1), metrics.SettledChannelsPerBlock)
	require.Equal(t, uint64(19), metrics.AverageChannelLifetime)
	require.Equal(t, "2800", metrics.TotalLockedNaet)
	require.Equal(t, "2800", lockedAmountForType(metrics.LockedNaetByChannelType, ChannelTypeBidirectional))
	require.Equal(t, uint64(1), metrics.ConditionalPromisesActive)
	require.Equal(t, uint64(1), metrics.ConditionalPromisesExpired)
	require.Equal(t, uint64(1), metrics.ConditionalPromisesResolved)
	require.Equal(t, uint64(1), metrics.VirtualChannelsActive)
	require.Equal(t, uint64(1), metrics.RoutingAdvertisementsActive)
	require.Equal(t, uint64(2), metrics.FraudProofsSubmitted)
	require.Equal(t, uint64(1), metrics.FraudProofsAccepted)
	require.Equal(t, uint64(1), metrics.FraudProofsRejected)
	require.Equal(t, uint64(1), metrics.PenaltiesApplied)
	require.Equal(t, uint64(1), metrics.ReporterRewardsPaid)
	require.Equal(t, "13", metrics.SettlementFeesCollected)
	require.Equal(t, DefaultOpeningFee, metrics.ChannelOpenFeeAverage)
	require.Equal(t, uint64(9), metrics.DisputeInclusionLatency)
	require.Equal(t, uint64(1), metrics.ChallengePeriodNearExpiryCount)
	require.Equal(t, uint64(10000), metrics.BlockSTMConflictRateBps)
	require.Greater(t, metrics.StoreV2PaymentModuleReadLatencyOps, uint64(0))
	require.GreaterOrEqual(t, metrics.StoreV2PaymentModuleWriteLatencyOps, metrics.StoreV2PaymentModuleReadLatencyOps)
	require.NoError(t, ValidateHash("observability metrics hash", metrics.MetricsHash))

	alerts, err := EvaluatePaymentObservabilityAlerts(state, metrics, PaymentObservabilityAlertThresholds{
		WindowBlocks:			100,
		HighPendingDisputeCount:	0,
		NearExpiryWithoutFinalize:	0,
		FraudProofRejectionSpike:	0,
		ChannelOpenSpamSpike:		0,
		PromiseExpiryBacklog:		0,
		SettlementQueueBacklog:		0,
		BlockSTMConflictRateBps:	1,
		StoreLatencyOps:		1,
		WatchReplayLagBlocks:		1,
		RoutingGossipSpamRateBps:	1,
	}, PaymentObservabilityExternalSignals{
		WatcherReplayHeight:	20,
		RoutingGossipMessages:	10,
		RoutingGossipRejected:	5,
	})
	require.NoError(t, err)
	require.Len(t, alerts, 10)
	alertTypes := make(map[PaymentObservabilityAlertType]bool, len(alerts))
	for _, alert := range alerts {
		require.NoError(t, alert.Validate())
		alertTypes[alert.AlertType] = true
	}
	for _, alertType := range []PaymentObservabilityAlertType{
		PaymentAlertHighPendingDisputeCount,
		PaymentAlertChallengeNearExpiryWithoutFinalization,
		PaymentAlertFraudProofRejectionSpike,
		PaymentAlertChannelOpenSpamSpike,
		PaymentAlertPromiseExpiryBacklog,
		PaymentAlertSettlementQueueBacklog,
		PaymentAlertBlockSTMConflictRateAboveThreshold,
		PaymentAlertPaymentModuleStoreLatencyAboveThreshold,
		PaymentAlertWatchServiceEventReplayLag,
		PaymentAlertRoutingGossipSpamRateAboveThreshold,
	} {
		require.True(t, alertTypes[alertType], "missing alert %s", alertType)
	}

	reports, err := BuildPaymentObservabilityReports(state, fraud, metrics, 1, currentHeight)
	require.NoError(t, err)
	require.Len(t, reports, 8)
	reportTypes := make(map[PaymentObservabilityReportType]PaymentObservabilityReport, len(reports))
	for _, report := range reports {
		require.NoError(t, report.Validate())
		reportTypes[report.ReportType] = report
	}
	for _, reportType := range []PaymentObservabilityReportType{
		PaymentReportDailyLockedLiquidity,
		PaymentReportDailySettlementVolume,
		PaymentReportDailyRoutingFee,
		PaymentReportDailyDisputeAndFraud,
		PaymentReportDailyStateFootprint,
		PaymentReportWeeklyChannelChurn,
		PaymentReportWeeklyLiquidityConcentration,
		PaymentReportWeeklyPerformance,
	} {
		require.Contains(t, reportTypes, reportType)
	}
	require.Equal(t, "2800", reportTypes[PaymentReportDailyLockedLiquidity].TotalLockedNaet)
	require.Equal(t, "998", reportTypes[PaymentReportDailySettlementVolume].SettlementVolumeNaet)
	require.Equal(t, "7", reportTypes[PaymentReportDailyRoutingFee].RoutingFeesNaet)
	require.Equal(t, uint64(2), reportTypes[PaymentReportDailyDisputeAndFraud].FraudProofsSubmitted)
	require.Greater(t, reportTypes[PaymentReportDailyStateFootprint].StateFootprintRecords, uint64(0))
	require.Equal(t, uint64(4), reportTypes[PaymentReportWeeklyChannelChurn].ChannelOpens)
	require.Equal(t, uint64(1), reportTypes[PaymentReportWeeklyChannelChurn].ChannelSettlements)
	require.Greater(t, reportTypes[PaymentReportWeeklyLiquidityConcentration].LiquidityConcentrationBps, uint64(0))
	require.Equal(t, uint64(10000), reportTypes[PaymentReportWeeklyPerformance].BlockSTMConflictRateBps)
}

func TestPaymentGovernanceParametersValidateChannelAndConditionalBounds(t *testing.T) {
	alice := testAddress(0xe0)
	bob := testAddress(0xe1)

	params := DefaultPaymentGovernanceParams()
	require.NoError(t, params.Validate())
	require.NoError(t, ValidateHash("payments governance params hash", params.ParamsHash))

	params.Channel.MinimumChannelCollateral = "100"
	params.Channel.MaximumChannelCollateral = "1000"
	params.Channel.MinimumChallengePeriod = 4
	params.Channel.MaximumChallengePeriod = 64
	params.Channel.DefaultChallengePeriod = 16
	params.Channel.MinimumCloseDelay = 2
	params.Channel.MaximumCloseDelay = 32
	params.Channel.ChannelOpenBaseFee = "9"
	params.Channel.ChannelStorageFeePerByte = "2"
	params.Channel.ChannelTombstoneRetention = 500
	params.Conditional.MaximumActivePromisesPerChannel = 1
	params.Conditional.MaximumPromiseAmountRatioBps = 2_500
	params.Conditional.MinimumTimeoutMargin = 10
	params.Conditional.MaximumPromiseLifetime = 70
	params.Conditional.BatchResolutionMaximumSize = 1
	params.Conditional.PromiseStorageFee = "3"
	params.Conditional.ExpiredPromiseCleanupLimitPerBlock = 7
	params = params.WithHash()
	require.NoError(t, params.Validate())

	schedule, err := BuildGovernedPaymentFeeSchedule(params)
	require.NoError(t, err)
	require.Equal(t, "9", schedule.ChannelOpenFee)
	require.Equal(t, "2", schedule.StorageByteFee)
	require.Equal(t, "3", schedule.ConditionalPromiseSettlementFee)

	openReq := ChannelOpenRequest{
		ChainID:		"aetra-test-1",
		ChannelID:		HashParts("governance-open", alice, bob),
		Participants:		[]string{alice, bob},
		InitialBalances:	[]Balance{{Participant: alice, Amount: "250"}, {Participant: bob, Amount: "0"}},
		ChannelType:		ChannelTypeBidirectional,
		Collateral:		"250",
		CloseDelay:		8,
		ChallengePeriod:	16,
		FeePolicyID:		NativeDenom,
		OpeningFeeDenom:	NativeDenom,
		OpeningFeePaid:		"9",
		OpenHeight:		10,
	}
	require.NoError(t, ValidateChannelOpenRequestWithGovernance(openReq, params))

	tooSmall := openReq
	tooSmall.Collateral = "99"
	tooSmall.InitialBalances = []Balance{{Participant: alice, Amount: "99"}, {Participant: bob, Amount: "0"}}
	require.ErrorContains(t, ValidateChannelOpenRequestWithGovernance(tooSmall, params), "below governance minimum")

	lowFee := openReq
	lowFee.OpeningFeePaid = "8"
	require.ErrorContains(t, ValidateChannelOpenRequestWithGovernance(lowFee, params), "below governance base fee")

	badChallenge := openReq
	badChallenge.ChallengePeriod = 3
	require.ErrorContains(t, ValidateChannelOpenRequestWithGovernance(badChallenge, params), "challenge period")

	expiry, err := SettlementTombstoneExpiryHeight(100, params)
	require.NoError(t, err)
	require.Equal(t, uint64(600), expiry)

	cleanupLimit, err := ExpiredPromiseCleanupLimit(params)
	require.NoError(t, err)
	require.Equal(t, uint64(7), cleanupLimit)

	channel := signedChannel(t, "governance-conditional", "1000", alice, bob)
	channel.LatestState = signedReserveState(t, channel, 2, channel.OpeningStateHash, "300", "0", []Balance{{Participant: alice, Amount: "700"}, {Participant: bob, Amount: "0"}})
	require.NoError(t, channel.Validate())

	promise := signedPromiseWithHashLock(t, channel, "governance-ok", alice, bob, "200", "10", 2, 30, HashParts("governance-preimage"))
	require.NoError(t, ValidateConditionalPromisesForChannelWithGovernance(channel, []ConditionalPromise{promise}, nil, params))

	second := signedPromiseWithHashLock(t, channel, "governance-too-many", alice, bob, "1", "0", 3, 31, promise.HashLock)
	require.ErrorContains(t, ValidateConditionalPromisesForChannelWithGovernance(channel, []ConditionalPromise{promise, second}, nil, params), "active promises exceed")

	tooLarge := signedPromiseWithHashLock(t, channel, "governance-too-large", alice, bob, "260", "0", 4, 32, promise.HashLock)
	require.ErrorContains(t, ValidateConditionalPromisesForChannelWithGovernance(channel, []ConditionalPromise{tooLarge}, nil, params), "exceeds governance channel ratio")

	tooLong := signedPromiseWithHashLock(t, channel, "governance-too-long", alice, bob, "10", "0", 5, 81, promise.HashLock)
	windowChannel := channel
	windowChannel.LatestState.TimeoutHeight = 120
	require.ErrorContains(t, params.Conditional.ValidatePromiseWindow(windowChannel, tooLong), "lifetime exceeds")
}

func TestPaymentGovernanceParametersValidateVirtualAndFraudPenaltyBounds(t *testing.T) {
	alice := testAddress(0xe2)
	router := testAddress(0xe3)
	bob := testAddress(0xe4)

	state, vc, proof := virtualChannelFixture(t, "governance-virtual", alice, router, bob, "100", 40)
	base := state.Clone()
	base.VirtualChannels = nil
	base.FeeCharges = nil

	params := DefaultPaymentGovernanceParams()
	params.Virtual.MaximumVirtualChannelsPerParentChannel = 2
	params.Virtual.MaximumVirtualChannelDepth = 4
	params.Virtual.MinimumParentTimeoutMargin = 10
	params.Virtual.VirtualChannelAnchorFee = "0"
	params.Virtual.VirtualChannelReservationExpiry = 5
	params.Virtual.MultiSegmentVirtualChannelMaxSegments = 2
	params.FraudPenalty.StaleClosePenalty = "11"
	params.FraudPenalty.SameNonceDoubleSignPenalty = "22"
	params.FraudPenalty.InvalidConditionPenalty = "13"
	params.FraudPenalty.ReplayAttemptPenalty = "14"
	params.FraudPenalty.InvalidFraudProofDeposit = "6"
	params.FraudPenalty.ReporterRewardPercentageBps = 2_000
	params.FraudPenalty.ReporterRewardCap = "5"
	params.FraudPenalty.PenaltyBurnAllocationBps = 3_000
	params.FraudPenalty.SecurityReserveAllocationBps = 4_000
	params.FraudPenalty.CounterpartyCompensationPriority = true
	params = params.WithHash()
	require.NoError(t, params.Validate())

	schedule, err := BuildGovernedPaymentFeeSchedule(params)
	require.NoError(t, err)
	require.Equal(t, "0", schedule.VirtualChannelAnchorFee)

	require.NoError(t, ValidateVirtualChannelWithGovernance(base, vc, params))
	require.NoError(t, ValidateVirtualActivationProofWithGovernance(base, proof, params))
	expiry, err := VirtualChannelReservationExpiryHeight(20, params)
	require.NoError(t, err)
	require.Equal(t, uint64(25), expiry)
	require.Equal(t, uint64(2), VirtualChannelDepth(vc))

	secondVC := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("governance-virtual-second", alice, bob),
		ParentChannelIDs:	vc.ParentChannelIDs,
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		40,
	}, vc.ChainID)
	perParentLimited := params
	perParentLimited.Virtual.MaximumVirtualChannelsPerParentChannel = 1
	perParentLimited = perParentLimited.WithHash()
	require.ErrorContains(t, ValidateVirtualChannelWithGovernance(state, secondVC, perParentLimited), "virtual channel limit")

	depthLimited := params
	depthLimited.Virtual.MaximumVirtualChannelDepth = 1
	depthLimited = depthLimited.WithHash()
	require.ErrorContains(t, ValidateVirtualChannelWithGovernance(base, vc, depthLimited), "depth exceeds")

	segmentLimited := params
	segmentLimited.Virtual.MultiSegmentVirtualChannelMaxSegments = 1
	segmentLimited = segmentLimited.WithHash()
	require.ErrorContains(t, ValidateVirtualChannelWithGovernance(base, vc, segmentLimited), "parent segments exceed")

	feeRequired := params
	feeRequired.Virtual.VirtualChannelAnchorFee = "1"
	feeRequired = feeRequired.WithHash()
	require.ErrorContains(t, ValidateVirtualChannelWithGovernance(base, vc, feeRequired), "anchor fee below")

	timeoutVC := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("governance-virtual-timeout", alice, bob),
		ParentChannelIDs:	vc.ParentChannelIDs,
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		81,
	}, vc.ChainID)
	require.ErrorContains(t, ValidateVirtualChannelWithGovernance(base, timeoutVC, params), "parent timeout margin")

	matrix, err := BuildGovernedPenaltyMatrix(params)
	require.NoError(t, err)
	doubleSign, err := PenaltyMatrixEntryForProof(FraudProofTypeDoubleSign, matrix)
	require.NoError(t, err)
	require.Equal(t, PenaltyClassDoubleSign, doubleSign.Class)
	require.Equal(t, "22", doubleSign.BasePenalty)
	require.Equal(t, "4", doubleSign.ReporterRewardCap)
	require.Equal(t, "22", doubleSign.CounterpartyCompensation)
	require.Equal(t, uint32(3_000), doubleSign.BurnShareBps)
	require.Equal(t, uint32(4_000), doubleSign.SecurityReserveShareBps)
	require.Equal(t, uint32(3_000), doubleSign.CommunityPoolShareBps)

	stale, err := PenaltyMatrixEntryForProof(FraudProofTypeStaleClose, matrix)
	require.NoError(t, err)
	require.Equal(t, "11", stale.BasePenalty)
	invalidCondition, err := PenaltyMatrixEntryForProof(FraudProofTypeInvalidCondition, matrix)
	require.NoError(t, err)
	require.Equal(t, "13", invalidCondition.BasePenalty)
	replay, err := PenaltyMatrixEntryForProof(FraudProofTypeReplayAttempt, matrix)
	require.NoError(t, err)
	require.Equal(t, "14", replay.BasePenalty)

	invalidDeposit := PenaltyMatrixEntry{}
	for _, entry := range matrix {
		if entry.Class == PenaltyClassInvalidFraudProof {
			invalidDeposit = entry
			break
		}
	}
	require.Equal(t, PenaltyClassInvalidFraudProof, invalidDeposit.Class)
	require.Equal(t, "6", invalidDeposit.InvalidProofVerifierCost)

	policy, err := BuildGovernedFraudPenaltyPolicy(params)
	require.NoError(t, err)
	require.Equal(t, "5", policy.ReporterRewardCap)
	require.Equal(t, uint32(10_000), policy.CounterpartyRewardBps)
	require.Equal(t, uint32(3_000), policy.BurnShareBps)
	require.Equal(t, uint32(4_000), policy.SecurityReserveShareBps)
	require.Equal(t, uint32(3_000), policy.CommunityPoolShareBps)
	require.True(t, policy.SecurityReserveHook)

	reward, err := GovernedReporterRewardAmount("100", params)
	require.NoError(t, err)
	require.Equal(t, "5", reward)

	badAlloc := params
	badAlloc.FraudPenalty.PenaltyBurnAllocationBps = 8_000
	badAlloc.FraudPenalty.SecurityReserveAllocationBps = 3_000
	badAlloc = badAlloc.WithHash()
	require.ErrorContains(t, badAlloc.Validate(), "exceed 10000")
}

func TestPaymentGovernanceParametersValidateRoutingAndExecutionBounds(t *testing.T) {
	alice := testAddress(0xe5)
	router := testAddress(0xe6)
	bob := testAddress(0xe7)
	channel := signedChannel(t, "governance-routing-execution", "1000", alice, bob)

	params := DefaultPaymentGovernanceParams()
	params.Routing.RoutingAdvertisementDeposit = "12"
	params.Routing.GossipMessageExpiry = 20
	params.Routing.LiquidityHintExpiry = 10
	params.Routing.MaximumTopologyUpdatesPerPeerWindow = 3
	params.Routing.RouteFailureScoreDecay = 5
	params.Routing.CongestionPenaltyDecay = 7
	params.Routing.CapacityProbeRateLimit = 2
	params.Routing.CapacityProbeWindow = 6
	params.Execution.SettlementBatchMaximumSize = 1
	params.Execution.FinalizationQueueWorkLimitPerBlock = 4
	params.Execution.ExpiredPromiseCleanupWorkLimitPerBlock = 5
	params.Execution.ChannelOpenCongestionFeeMultiplierBps = 30_000
	params.Execution.DisputeCongestionFeeMultiplierBps = 40_000
	params.Execution.StorePruningHorizon = 33
	params = params.WithHash()
	require.NoError(t, params.Validate())

	schedule, err := BuildGovernedPaymentFeeSchedule(params)
	require.NoError(t, err)
	require.Equal(t, "12", schedule.RoutingAdvertisementDeposit)

	routePolicy, err := BuildGovernedRoutePolicy(params)
	require.NoError(t, err)
	require.Equal(t, uint64(10), routePolicy.StaleLiquidityAfter)
	require.Equal(t, uint64(7), routePolicy.DecayHalfLife)

	gossipPolicy, err := BuildGovernedGossipRateLimitPolicy(params)
	require.NoError(t, err)
	require.Equal(t, uint64(20), gossipPolicy.WindowBlocks)
	require.Equal(t, uint32(3), gossipPolicy.MaxTopologyUpdates)

	failurePolicy, err := BuildGovernedRouteFailureScoringPolicy(params)
	require.NoError(t, err)
	score, err := BuildRouteFailureScore(RouteFailureReport{
		ChannelID:	channel.ChannelID,
		From:		alice,
		To:		bob,
		FailureClass:	RouteFailureCongestion,
		Retryable:	true,
		ObservedHeight:	10,
	}, 1, failurePolicy)
	require.NoError(t, err)
	decayed, err := DecayRouteFailureScoreWithGovernance(score, 15, params)
	require.NoError(t, err)
	require.Equal(t, score.ScoreDelta/2, decayed.ScoreDelta)

	gossip, err := BuildGossipMessage(GossipMessage{
		MessageType:		GossipChannelUpdate,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Capacity:		"100",
		FeeDenom:		NativeDenom,
		FeeAmount:		"1",
		ValidAfterHeight:	20,
		ValidUntilHeight:	40,
		Sequence:		1,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateGossipMessageExpiryWithGovernance(gossip, params))
	tooLong := gossip
	tooLong.ValidUntilHeight = 41
	require.ErrorContains(t, ValidateGossipMessageExpiryWithGovernance(tooLong, params), "expiry exceeds")

	liquidity, err := BuildGossipMessage(GossipMessage{
		MessageType:		GossipLiquidityHint,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Capacity:		"100",
		Liquidity:		"90",
		FeeDenom:		NativeDenom,
		FeeAmount:		"1",
		ValidAfterHeight:	20,
		ValidUntilHeight:	30,
		Sequence:		2,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateGossipMessageExpiryWithGovernance(liquidity, params))
	liquidity.ValidUntilHeight = 31
	require.ErrorContains(t, ValidateGossipMessageExpiryWithGovernance(liquidity, params), "expiry exceeds")

	probe := CapacityProbeRequest{From: alice, To: router, Amount: "10", CurrentHeight: 20, MaxHops: 3, BlindedRouteHint: HashParts("governance-probe")}
	existing := []CapacityProbeRequest{
		{From: alice, To: router, Amount: "10", CurrentHeight: 16, MaxHops: 3, BlindedRouteHint: HashParts("governance-probe-1")},
		{From: alice, To: bob, Amount: "10", CurrentHeight: 19, MaxHops: 3, BlindedRouteHint: HashParts("governance-probe-2")},
	}
	require.ErrorContains(t, ValidateCapacityProbeRateLimitWithGovernance(existing, probe, params), "rate limit")
	require.NoError(t, ValidateCapacityProbeRateLimitWithGovernance(existing[:1], probe, params))

	opA := SettlementOperation{OperationID: HashParts("governance-batch-a"), OperationType: BatchOperationClose, ChannelID: channel.ChannelID, Nonce: 1, StateHash: channel.LatestState.StateHash}
	opB := SettlementOperation{OperationID: HashParts("governance-batch-b"), OperationType: BatchOperationDispute, ChannelID: channel.ChannelID, Nonce: 2, StateHash: channel.LatestState.StateHash}
	_, err = NewSettlementBatchWithGovernance(HashParts("governance-batch"), []SettlementOperation{opA, opB}, params)
	require.ErrorContains(t, err, "settlement batch exceeds")
	batch, err := NewSettlementBatchWithGovernance(HashParts("governance-batch-ok"), []SettlementOperation{opA}, params)
	require.NoError(t, err)
	require.Len(t, batch.Operations, 1)

	next, result, err := ProcessAsyncExecutionQueuesWithGovernance(EmptyStateWithChannel(t, channel), 20, params)
	require.NoError(t, err)
	require.NoError(t, next.Validate())
	require.LessOrEqual(t, result.ProcessedFinalizations, params.Execution.FinalizationQueueWorkLimitPerBlock)
	require.LessOrEqual(t, result.ProcessedPromiseExpiries, params.Execution.ExpiredPromiseCleanupWorkLimitPerBlock)

	feeState, err := ApplyGovernedExecutionFeeMultipliers(EmptyState(), 21, 8_000, 9_000, params)
	require.NoError(t, err)
	require.Len(t, feeState.FeeMultipliers, 2)
	require.Equal(t, uint32(30_000), feeMultiplierForClass(feeState, PaymentFeeClassChannelOpen, feeState.FeeSchedule))
	require.Equal(t, uint32(40_000), feeMultiplierForClass(feeState, PaymentFeeClassDispute, feeState.FeeSchedule))

	pruneAfter, err := StorePruneAfterHeightWithGovernance(100, params)
	require.NoError(t, err)
	require.Equal(t, uint64(133), pruneAfter)

	badExecution := params
	badExecution.Execution.SettlementBatchMaximumSize = MaxSettlementBatchOps + 1
	badExecution = badExecution.WithHash()
	require.ErrorContains(t, badExecution.Validate(), "settlement batch max")

	badMultiplier := params
	badMultiplier.Execution.ChannelOpenCongestionFeeMultiplierBps = 1
	badMultiplier = badMultiplier.WithHash()
	require.ErrorContains(t, badMultiplier.Validate(), "channel-open fee multiplier")
}

func TestSettlementArbitrationBoundaryRejectsNonDeterministicInputs(t *testing.T) {
	alice := testAddress(0x12)
	bob := testAddress(0x13)
	channel := signedChannel(t, "settlement-boundary", "1000", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})

	valid := SettlementArbitrationInput{
		Operation:	SettlementArbitrationUnilateralClose,
		ChannelID:	channel.ChannelID,
		SignedState:	closeState,
		CurrentHeight:	20,
	}
	require.NoError(t, valid.ValidateForChannel(channel))

	withRoute := valid
	withRoute.RouteHints = []ChannelEdge{{ChannelID: channel.ChannelID}}
	require.ErrorContains(t, withRoute.ValidateForChannel(channel), "must not select payment routes")

	withGossip := valid
	withGossip.GossipStateHash = HashParts("gossip", channel.ChannelID)
	require.ErrorContains(t, withGossip.ValidateForChannel(channel), "must not trust gossip")

	withLiquidity := valid
	withLiquidity.ExternalLiquidity = []Balance{{Participant: alice, Amount: "1000"}}
	require.ErrorContains(t, withLiquidity.ValidateForChannel(channel), "must not depend on external liquidity")

	withUnsignedBalance := valid
	withUnsignedBalance.UnsignedBalances = []Balance{{Participant: bob, Amount: "1000"}}
	require.ErrorContains(t, withUnsignedBalance.ValidateForChannel(channel), "must not accept unsigned balance")

	withIntent := valid
	withIntent.OffchainIntent = "alice verbally approved this close"
	require.ErrorContains(t, withIntent.ValidateForChannel(channel), "must not infer participant intent")
}

func TestSettlementArbitrationBoundaryRequiresSignedState(t *testing.T) {
	alice := testAddress(0x16)
	bob := testAddress(0x17)
	channel := signedChannel(t, "settlement-signed-state", "1000", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "300"},
		{Participant: bob, Amount: "700"},
	})
	closeState.Signatures = nil

	err := (SettlementArbitrationInput{
		Operation:	SettlementArbitrationUnilateralClose,
		ChannelID:	channel.ChannelID,
		SignedState:	closeState,
		CurrentHeight:	20,
	}).ValidateForChannel(channel)
	require.ErrorContains(t, err, "quorum")
}

func TestUnilateralCloseRequestStoresReasonAndDetachedSignatures(t *testing.T) {
	alice := testAddress(0x18)
	bob := testAddress(0x19)
	channel := signedChannel(t, "unilateral-close-request", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "425"},
		{Participant: bob, Amount: "575"},
	})
	detached := closeState.Signatures
	closeState.Signatures = nil
	state, err = SubmitCloseWithRequest(state, ChannelCloseRequest{
		ChannelID:	channel.ChannelID,
		ClosingState:	closeState,
		Signatures:	detached,
		CloseReason:	CloseReasonUnilateral,
		Submitter:	bob,
		CurrentHeight:	20,
		SettlementFee:	"3",
	})
	require.NoError(t, err)
	require.Equal(t, ChannelStatusPendingClose, state.Channels[0].Status)
	require.Equal(t, CloseReasonUnilateral, state.Channels[0].PendingClose.CloseReason)
	require.Equal(t, uint64(20), state.Channels[0].PendingClose.SubmittedHeight)
	require.Equal(t, uint64(28), state.Channels[0].PendingClose.SettleAfterHeight)
	require.Equal(t, "3", state.Channels[0].PendingClose.SettlementFee)
}

func TestFraudProofInvalidBalanceRoutesPenaltyRemainder(t *testing.T) {
	alice := testAddress(0x1a)
	bob := testAddress(0x1b)
	channel := signedChannel(t, "fraud-penalty-routing", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	invalid := mutateCanonicalState(closeState, func(next *ChannelState) {
		next.Nonce = 3
		next.PreviousStateHash = closeState.StateHash
		next.Balances = []Balance{
			{Participant: alice, Amount: "900"},
			{Participant: bob, Amount: "900"},
		}
		next.BalanceA = "900"
		next.BalanceB = "900"
	})
	invalid = resignState(t, channel, invalid)
	require.ErrorContains(t, invalid.ValidateForChannel(channel, false), "conserve")

	proof := FraudProof{
		ProofID:		HashParts("invalid-balance", channel.ChannelID),
		ProofType:		FraudProofTypeInvalidBalance,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			invalid,
		PenaltyAmount:		"25",
		EvidenceHash:		HashParts("evidence", invalid.StateHash),
	}
	state, err = SubmitFraudProofWithPolicy(state, channel.ChannelID, proof, 21, FraudPenaltyPolicy{
		ReporterRewardCap:		"10",
		BurnShareBps:			5000,
		SecurityReserveShareBps:	2500,
		CommunityPoolShareBps:		2500,
	})
	require.NoError(t, err)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 1)
	require.Equal(t, "10", state.Channels[0].PendingClose.Penalties[0].Amount)
	require.Equal(t, "7", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteBurn))
	require.Equal(t, "3", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteSecurityReserve))
	require.Equal(t, "5", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteCommunityPool))

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 50)
	require.NoError(t, err)
	require.Equal(t, "375", amountFor(settlement.FinalBalances, alice))
	require.Equal(t, "610", amountFor(settlement.FinalBalances, bob))
	require.Len(t, settlement.PenaltyAllocations, 3)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
}

func TestSettlementFinalityTransitionsPendingHeightAndEvents(t *testing.T) {
	alice := testAddress(0x1c)
	bob := testAddress(0x1d)
	channel := signedChannel(t, "settlement-finality", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityOpen, state.Channels[0].Finality)
	require.Equal(t, "channel-finality-transition", state.Events[len(state.Events)-1].EventType)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "410"},
		{Participant: bob, Amount: "590"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPendingClose, state.Channels[0].Finality)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	height, found, err := state.PendingFinalizationHeight(channel.ChannelID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(28), height)

	state, err = AdvanceChannelFinality(state, channel.ChannelID, 27)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPendingClose, state.Channels[0].Finality)

	state, err = AdvanceChannelFinality(state, channel.ChannelID, 28)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityFinalizable, state.Channels[0].Finality)
	require.Equal(t, "channel-finality-transition", state.Events[len(state.Events)-1].EventType)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 28)
	require.NoError(t, err)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
	require.Equal(t, ChannelFinalitySettled, state.Channels[0].Finality)
	require.Empty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
}

func TestDisputeAndPenaltyFinalityTransitionsRetainCollateralUntilSettlement(t *testing.T) {
	alice := testAddress(0x1e)
	bob := testAddress(0x1f)
	channel := signedChannel(t, "dispute-penalty-finality", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newerState, bob, 21)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityInDispute, state.Channels[0].Finality)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	conflicting := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	proof := FraudProof{
		ProofID:		HashParts("finality-proof", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			newerState,
		StateB:			conflicting,
		PenaltyAmount:		"20",
		EvidenceHash:		HashParts("evidence", newerState.StateHash, conflicting.StateHash),
	}
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 22)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPenalized, state.Channels[0].Finality)
	require.NotEmpty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	state, _, err = FraudClose(state, channel.ChannelID, 23)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPenalized, state.Channels[0].Finality)
	require.Empty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
}

func TestLockedCollateralInvariantForEveryFinalityState(t *testing.T) {
	alice := testAddress(0x20)
	bob := testAddress(0x21)
	base := signedChannel(t, "finality-invariant", "1000", alice, bob)
	lock := CustodyLock{ChannelID: base.ChannelID, Denom: NativeDenom, Amount: base.Collateral}

	cases := []struct {
		name		string
		finality	ChannelFinality
		status		ChannelStatus
		locks		[]CustodyLock
	}{
		{"open", ChannelFinalityOpen, ChannelStatusOpen, []CustodyLock{lock}},
		{"pending-close", ChannelFinalityPendingClose, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"in-dispute", ChannelFinalityInDispute, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"pending-condition", ChannelFinalityPendingConditionResolution, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"finalizable", ChannelFinalityFinalizable, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"pending-penalized", ChannelFinalityPenalized, ChannelStatusPendingClose, []CustodyLock{lock}},
		{"expired", ChannelFinalityExpired, ChannelStatusOpen, []CustodyLock{lock}},
		{"settled", ChannelFinalitySettled, ChannelStatusSettled, nil},
		{"settled-penalized", ChannelFinalityPenalized, ChannelStatusSettled, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			channel := base
			channel.Status = tc.status
			channel.Finality = tc.finality
			require.NoError(t, ValidateLockedCollateralForFinality(PaymentsState{
				Channels:	[]ChannelRecord{channel},
				CustodyLocks:	tc.locks,
			}))
		})
	}

	missing := base
	missing.Finality = ChannelFinalityFinalizable
	missing.Status = ChannelStatusPendingClose
	require.ErrorContains(t, ValidateLockedCollateralForFinality(PaymentsState{Channels: []ChannelRecord{missing}}), "retain custody")
}

func TestConditionalPromiseObjectSignatureReserveAndReplayRules(t *testing.T) {
	alice := testAddress(0x22)
	bob := testAddress(0x23)
	channel := signedChannel(t, "promise-object", "1000", alice, bob)
	channel.LatestState = signedReserveState(t, channel, 2, channel.OpeningStateHash, "40", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "60"},
	})
	require.NoError(t, channel.Validate())

	promise := signedPromise(t, channel, "promise-a", alice, bob, "25", "5", 7, 40)
	require.NoError(t, promise.ValidateForChannel(channel))
	require.Equal(t, promise.PromiseID, promise.ToConditionalPayment().ConditionID)
	require.Equal(t, promise.Amount, promise.ToConditionalPayment().Amount)
	require.NoError(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise}, nil))

	duplicate := signedPromise(t, channel, "promise-a", alice, bob, "1", "0", 8, 40)
	require.ErrorContains(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise, duplicate}, nil), "duplicate promise")

	overReserve := signedPromise(t, channel, "promise-b", alice, bob, "10", "1", 9, 40)
	require.ErrorContains(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise, overReserve}, nil), "reserve")

	settled := []ConditionClaimRecord{{
		ChainID:	channel.ChainID,
		ChannelID:	channel.ChannelID,
		ConditionID:	promise.PromiseID,
		EvidenceHash:	HashParts("promise-settled", promise.PromiseID),
		ResolvedHeight:	50,
		ExpiresHeight:	100,
	}}
	require.ErrorContains(t, ValidateConditionalPromisesForChannel(channel, []ConditionalPromise{promise}, settled), "already been settled")

	late := signedPromise(t, channel, "promise-late", alice, bob, "1", "0", 10, channel.LatestState.TimeoutHeight)
	require.ErrorContains(t, late.ValidateForChannel(channel), "dispute window")

	wrongSigner := promise
	sig, err := SignatureForPromise(channel, wrongSigner, bob)
	require.NoError(t, err)
	wrongSigner.Signature = sig
	require.ErrorContains(t, wrongSigner.ValidateForChannel(channel), "signer must be source")
}

func TestHashLockedPreimageRevealResolvesLinkedPromisesAndTracksPreimage(t *testing.T) {
	alice := testAddress(0x24)
	bob := testAddress(0x25)
	channel := signedChannel(t, "promise-reveal", "1000", alice, bob)
	reserveState := signedReserveState(t, channel, 2, channel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	require.NoError(t, channel.Validate())

	preimage := "shared-secret"
	hashLock := HashParts(preimage)
	first := signedPromiseWithHashLock(t, channel, "reveal-a", alice, bob, "20", "1", 7, 40, hashLock)
	second := signedPromiseWithHashLock(t, channel, "reveal-b", alice, bob, "10", "1", 8, 40, hashLock)

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, channel.ChannelID, reserveState, 20)
	require.NoError(t, err)
	freshState := state.Clone()

	state, resolutions, err := RevealPromisePreimage(state, PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{first, second},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	30,
	})
	require.NoError(t, err)
	require.Len(t, resolutions, 2)
	resolutionsByID := map[string]ConditionResolution{}
	for _, resolution := range resolutions {
		resolutionsByID[resolution.ConditionID] = resolution
	}
	require.Equal(t, bob, resolutionsByID[first.PromiseID].Recipient)
	require.Equal(t, bob, resolutionsByID[second.PromiseID].Recipient)
	require.Len(t, state.ConditionClaims, 2)
	claimsByID := map[string]ConditionClaimRecord{}
	for _, claim := range state.ConditionClaims {
		claimsByID[claim.ConditionID] = claim
	}
	require.Equal(t, hashLock, claimsByID[first.PromiseID].PreimageHash)
	require.Equal(t, hashLock, claimsByID[second.PromiseID].PreimageHash)

	_, _, err = RevealPromisePreimage(state, PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{first},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	31,
	})
	require.ErrorContains(t, err, "already been settled")

	_, _, err = RevealPromisePreimage(freshState.Clone(), PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{first},
		Preimage:	"wrong-secret",
		Revealer:	bob,
		CurrentHeight:	30,
	})
	require.ErrorContains(t, err, "does not satisfy hash lock")

	_, _, err = RevealPromisePreimage(freshState.Clone(), PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{first},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	41,
	})
	require.ErrorContains(t, err, "timed out")
}

func TestTimeoutOrderingAndExpiryResolutionReleaseConditionRoot(t *testing.T) {
	alice := testAddress(0x26)
	bob := testAddress(0x27)
	channel := signedChannel(t, "promise-expiry", "1000", alice, bob)
	base := signedReserveState(t, channel, 2, channel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	promiseChannel := channel
	promiseChannel.LatestState = base
	hashLock := HashParts("race-preimage")
	downstreamID := HashParts("promise", channel.ChannelID, "downstream")
	upstreamID := HashParts("promise", channel.ChannelID, "upstream")
	downstream := signedLinkedPromise(t, promiseChannel, downstreamID, alice, bob, "20", "1", 9, 40, hashLock, "", upstreamID)
	upstream := signedLinkedPromise(t, promiseChannel, upstreamID, alice, bob, "15", "1", 10, 70, hashLock, downstreamID, "")

	require.NoError(t, ValidatePromiseTimeoutOrdering(promiseChannel, upstream, downstream, DefaultTimeoutMargin))
	require.NoError(t, ValidatePromiseTimeoutChain(promiseChannel, []ConditionalPromise{downstream, upstream}, DefaultTimeoutMargin))
	require.ErrorContains(t, ValidatePromiseTimeoutOrdering(promiseChannel, upstream, downstream, 4), "margin")
	badUpstream := signedLinkedPromise(t, promiseChannel, HashParts("promise", channel.ChannelID, "bad-upstream"), alice, bob, "15", "1", 11, 50, hashLock, downstreamID, "")
	require.ErrorContains(t, ValidatePromiseTimeoutOrdering(promiseChannel, badUpstream, downstream, DefaultTimeoutMargin), "downstream timeout")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	conditioned, rootUpdate, err := BuildConditionRootUpdateFromPromises(promiseChannel, base, []ConditionalPromise{downstream, upstream}, nil)
	require.NoError(t, err)
	require.Equal(t, uint32(2), rootUpdate.ConditionCount)
	conditioned = resignState(t, channel, conditioned)
	state, err = AcceptSignedState(state, channel.ChannelID, conditioned, 20)
	require.NoError(t, err)

	_, _, _, err = ExpireConditionalPromises(state, PromiseExpiryRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{downstream},
		Resolver:	alice,
		CurrentHeight:	downstream.TimeoutHeight,
	})
	require.ErrorContains(t, err, "has not expired")

	state, resolutions, expiryUpdate, err := ExpireConditionalPromises(state, PromiseExpiryRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{downstream},
		Resolver:	alice,
		CurrentHeight:	downstream.TimeoutHeight + 1,
	})
	require.NoError(t, err)
	require.Len(t, resolutions, 1)
	require.True(t, resolutions[0].Expired)
	require.Equal(t, alice, resolutions[0].Recipient)
	require.Len(t, state.ConditionClaims, 1)
	require.Equal(t, uint32(1), expiryUpdate.ConditionCount)
	require.NotEqual(t, rootUpdate.ConditionRoot, expiryUpdate.ConditionRoot)

	_, _, err = RevealPromisePreimage(state, PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{downstream},
		Preimage:	"race-preimage",
		Revealer:	bob,
		CurrentHeight:	downstream.TimeoutHeight + 1,
	})
	require.ErrorContains(t, err, "timed out")
}

func TestBatchConditionSettlementAtomicallyResolvesChainedPromises(t *testing.T) {
	alice := testAddress(0x28)
	router := testAddress(0x29)
	bob := testAddress(0x2a)
	routeID := HashParts("route", alice, router, bob)
	hashLock := HashParts("atomic-preimage")
	firstChannel := signedChannel(t, "chain-first", "1000", alice, router)
	secondChannel := signedChannel(t, "chain-second", "1000", router, bob)
	firstID := HashParts("promise", routeID, "first")
	secondID := HashParts("promise", routeID, "second")

	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstPromiseChannel := firstChannel
	firstPromiseChannel.LatestState = firstBase
	secondPromiseChannel := secondChannel
	secondPromiseChannel.LatestState = secondBase
	first := signedRoutePromise(t, firstPromiseChannel, firstID, routeID, alice, router, "31", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondPromiseChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, firstChannel)
	require.NoError(t, err)
	state, err = OpenChannel(state, secondChannel)
	require.NoError(t, err)
	firstConditioned, firstRoot, err := BuildConditionRootUpdateFromPromises(firstPromiseChannel, firstBase, []ConditionalPromise{first}, nil)
	require.NoError(t, err)
	secondConditioned, secondRoot, err := BuildConditionRootUpdateFromPromises(secondPromiseChannel, secondBase, []ConditionalPromise{second}, nil)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, firstChannel.ChannelID, resignState(t, firstChannel, firstConditioned), 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, secondChannel.ChannelID, resignState(t, secondChannel, secondConditioned), 20)
	require.NoError(t, err)

	proof := ConditionLinkageProof{
		RouteID:	routeID,
		Promises:	[]ConditionalPromise{first, second},
		Sender:		alice,
		Receiver:	bob,
		Amount:		"30",
		TotalFees:	"1",
		HashLock:	hashLock,
		TimeoutMargin:	DefaultTimeoutMargin,
	}
	state, result, err := BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof:	proof,
		Mode:		ConditionSettlementModePreimage,
		Preimage:	"atomic-preimage",
		Resolver:	bob,
		CurrentHeight:	30,
	})
	require.NoError(t, err)
	require.Len(t, result.Resolutions, 2)
	require.Len(t, result.FeeClaims, 1)
	require.Equal(t, router, result.FeeClaims[0].Recipient)
	require.Equal(t, "1", result.FeeClaims[0].Amount)
	require.Len(t, result.ConditionRootUpdates, 2)
	require.NotContains(t, []string{firstRoot.ConditionRoot, secondRoot.ConditionRoot}, result.ConditionRootUpdates[0].ConditionRoot)
	require.Len(t, state.ConditionClaims, 2)
	require.Equal(t, hashLock, state.ConditionClaims[0].PreimageHash)
	require.Equal(t, hashLock, state.ConditionClaims[1].PreimageHash)

	_, _, err = BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof:	proof,
		Mode:		ConditionSettlementModePreimage,
		Preimage:	"atomic-preimage",
		Resolver:	bob,
		CurrentHeight:	31,
	})
	require.ErrorContains(t, err, "already been settled")
}

func TestConditionalPaymentsModuleMessagesRootsClaimsAndDisputes(t *testing.T) {
	alice := testAddress(0xd1)
	bob := testAddress(0xd2)
	channel := signedChannel(t, "conditional-module", "1000", alice, bob)
	base := signedReserveState(t, channel, 2, channel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	promiseChannel := channel
	promiseChannel.LatestState = base
	preimage := "conditional-module-preimage"
	promise := signedPromiseWithHashLock(t, promiseChannel, "conditional-module-promise", alice, bob, "20", "1", 9, 40, HashParts(preimage))
	state := EmptyStateWithChannel(t, channel)
	state, err := AcceptSignedState(state, channel.ChannelID, base, 19)
	require.NoError(t, err)

	state, snapshot, err := ApplyConditionalPaymentMessage(state, MsgRegisterPromise{
		Signer:		alice,
		ChannelID:	channel.ChannelID,
		BaseState:	base,
		Promises:	[]ConditionalPromise{promise},
		CurrentHeight:	20,
	})
	require.NoError(t, err)
	require.Len(t, snapshot.Promises, 1)
	require.Len(t, snapshot.ConditionRoots, 1)
	require.Len(t, snapshot.PromiseTimeouts, 1)
	require.NoError(t, ValidateReservedBalancesForConditions(state.Channels[0], state.Channels[0].LatestState))

	_, _, err = ApplyConditionalPaymentMessage(state, MsgRegisterPromise{
		Signer:		alice,
		ChannelID:	channel.ChannelID,
		BaseState:	base,
		Promises:	[]ConditionalPromise{promise, promise},
		CurrentHeight:	21,
	})
	require.ErrorContains(t, err, "duplicate promise")

	state, snapshot, err = ApplyConditionalPaymentMessage(state, MsgResolveWithPreimage{Request: PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{promise},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	30,
	}})
	require.NoError(t, err)
	require.Len(t, snapshot.PreimageClaims, 1)
	require.Equal(t, HashParts(preimage), state.ConditionClaims[0].PreimageHash)

	_, _, err = ApplyConditionalPaymentMessage(state, MsgResolveWithPreimage{Request: PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{promise},
		Preimage:	preimage,
		Revealer:	bob,
		CurrentHeight:	31,
	}})
	require.ErrorContains(t, err, "already been settled")

	invalidResolution := ConditionResolution{
		ConditionID:	promise.PromiseID,
		Resolver:	bob,
		Recipient:	alice,
		Amount:		promise.Amount,
		EvidenceHash:	HashParts("invalid-condition-resolution", promise.PromiseID),
	}
	disputed, _, err := ApplyConditionalPaymentMessage(state, MsgDisputeCondition{
		Signer:		alice,
		ChannelID:	channel.ChannelID,
		Promise:	promise,
		Resolution:	invalidResolution,
		Reason:		"wrong-recipient",
		CurrentHeight:	32,
	})
	require.NoError(t, err)
	require.Equal(t, "condition_dispute", disputed.Events[len(disputed.Events)-1].EventType)
}

func TestConditionalPaymentsModuleExpiryAndBatchFinalizeRecords(t *testing.T) {
	alice := testAddress(0xd3)
	router := testAddress(0xd4)
	bob := testAddress(0xd5)
	routeID := HashParts("conditional-module-route", alice, router, bob)
	hashLock := HashParts("conditional-module-batch-preimage")
	firstChannel := signedChannel(t, "conditional-module-first", "1000", alice, router)
	secondChannel := signedChannel(t, "conditional-module-second", "1000", router, bob)
	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstPromiseChannel := firstChannel
	firstPromiseChannel.LatestState = firstBase
	secondPromiseChannel := secondChannel
	secondPromiseChannel.LatestState = secondBase
	firstID := HashParts("conditional-module-promise", "first")
	secondID := HashParts("conditional-module-promise", "second")
	first := signedRoutePromise(t, firstPromiseChannel, firstID, routeID, alice, router, "31", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondPromiseChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, firstChannel)
	require.NoError(t, err)
	state, err = OpenChannel(state, secondChannel)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, firstChannel.ChannelID, firstBase, 19)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, secondChannel.ChannelID, secondBase, 19)
	require.NoError(t, err)
	state, _, err = ApplyConditionalPaymentMessage(state, MsgRegisterPromise{Signer: alice, ChannelID: firstChannel.ChannelID, BaseState: firstBase, Promises: []ConditionalPromise{first}, CurrentHeight: 20})
	require.NoError(t, err)
	state, _, err = ApplyConditionalPaymentMessage(state, MsgRegisterPromise{Signer: router, ChannelID: secondChannel.ChannelID, BaseState: secondBase, Promises: []ConditionalPromise{second}, CurrentHeight: 20})
	require.NoError(t, err)

	proof := ConditionLinkageProof{
		RouteID:	routeID,
		Promises:	[]ConditionalPromise{first, second},
		Sender:		alice,
		Receiver:	bob,
		Amount:		"30",
		TotalFees:	"1",
		HashLock:	hashLock,
		TimeoutMargin:	DefaultTimeoutMargin,
	}
	state, snapshot, err := ApplyConditionalPaymentMessage(state, MsgBatchResolvePromises{Request: BatchConditionSettlementRequest{
		LinkageProof:	proof,
		Mode:		ConditionSettlementModePreimage,
		Preimage:	"conditional-module-batch-preimage",
		Resolver:	bob,
		CurrentHeight:	30,
	}})
	require.NoError(t, err)
	require.Len(t, state.ConditionClaims, 2)
	require.NotEmpty(t, snapshot.ExpiredClaims)
	require.Equal(t, "condition_settlement", state.Events[len(state.Events)-1].EventType)

	record := ConditionSettlementRecordFromBatch(BatchConditionSettlementRequest{
		LinkageProof:	proof,
		Mode:		ConditionSettlementModePreimage,
		Resolver:	bob,
		CurrentHeight:	30,
	}, BatchConditionSettlementResult{
		RouteID:		routeID,
		Resolutions:		[]ConditionResolution{{ConditionID: first.PromiseID, Resolver: bob, Recipient: router, Amount: first.Amount, EvidenceHash: HashParts("finalize-condition", first.PromiseID)}},
		ConditionRootUpdates:	[]ConditionRootUpdate{{ChannelID: first.ChannelID, Nonce: 2, ConditionRoot: state.Channels[0].LatestState.ConditionRoot, ConditionCount: 1, Conditions: state.Channels[0].LatestState.Conditions}},
		EvidenceHash:		HashParts("finalize-condition", routeID),
	})
	state, _, err = ApplyConditionalPaymentMessage(state, MsgFinalizeConditionSettlement{Signer: bob, Settlement: record, CurrentHeight: 31})
	require.NoError(t, err)
	require.Equal(t, "condition_settlement", state.Events[len(state.Events)-1].EventType)
}

func TestBatchConditionSettlementRejectsBrokenRouteInvariants(t *testing.T) {
	alice := testAddress(0x2b)
	router := testAddress(0x2c)
	bob := testAddress(0x2d)
	routeID := HashParts("route-bad", alice, router, bob)
	hashLock := HashParts("bad-route-preimage")
	firstChannel := signedChannel(t, "chain-bad-first", "1000", alice, router)
	secondChannel := signedChannel(t, "chain-bad-second", "1000", router, bob)
	openFirst := firstChannel
	openSecond := secondChannel
	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstChannel.LatestState = firstBase
	secondChannel.LatestState = secondBase
	firstID := HashParts("promise", routeID, "first")
	secondID := HashParts("promise", routeID, "second")
	first := signedRoutePromise(t, firstChannel, firstID, routeID, alice, router, "30", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, openFirst)
	require.NoError(t, err)
	state, err = OpenChannel(state, openSecond)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, openFirst.ChannelID, firstBase, 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, openSecond.ChannelID, secondBase, 20)
	require.NoError(t, err)
	require.ErrorContains(t, ConditionLinkageProof{
		RouteID:	routeID,
		Promises:	[]ConditionalPromise{first, second},
		Sender:		alice,
		Receiver:	bob,
		Amount:		"30",
		TotalFees:	"1",
		HashLock:	hashLock,
		TimeoutMargin:	DefaultTimeoutMargin,
	}.ValidateForState(state, nil), "amount conservation")

	firstConservedID := HashParts("promise", routeID, "first-conserved")
	badTimeoutID := HashParts("promise", routeID, "bad-timeout")
	firstConserved := signedRoutePromise(t, firstChannel, firstConservedID, routeID, alice, router, "31", "0", 12, 70, hashLock, "", badTimeoutID)
	badTimeout := signedRoutePromise(t, secondChannel, badTimeoutID, routeID, router, bob, "30", "1", 11, 60, hashLock, firstConservedID, "")
	require.ErrorContains(t, ConditionLinkageProof{
		RouteID:	routeID,
		Promises:	[]ConditionalPromise{firstConserved, badTimeout},
		Sender:		alice,
		Receiver:	bob,
		Amount:		"30",
		TotalFees:	"1",
		HashLock:	hashLock,
		TimeoutMargin:	DefaultTimeoutMargin,
	}.ValidateForState(state, nil), "downstream timeout")

	partialState, partialResult, err := BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof: ConditionLinkageProof{
			RouteID:			routeID,
			Promises:			[]ConditionalPromise{second},
			Sender:				router,
			Receiver:			bob,
			Amount:				"30",
			TotalFees:			"0",
			HashLock:			hashLock,
			TimeoutMargin:			DefaultTimeoutMargin,
			PartialDispute:			true,
			OffchainResolvedPromiseIDs:	[]string{firstID},
		},
		Mode:		ConditionSettlementModePreimage,
		Preimage:	"bad-route-preimage",
		Resolver:	bob,
		CurrentHeight:	30,
	})
	require.NoError(t, err)
	require.Len(t, partialResult.Resolutions, 1)
	require.Empty(t, partialResult.FeeClaims)
	require.Len(t, partialState.ConditionClaims, 1)
}

func TestBatchConditionSettlementExpiryIsAtomicWithoutFees(t *testing.T) {
	alice := testAddress(0x2e)
	router := testAddress(0x2f)
	bob := testAddress(0x30)
	routeID := HashParts("route-expiry", alice, router, bob)
	hashLock := HashParts("expiry-preimage")
	firstChannel := signedChannel(t, "chain-expiry-first", "1000", alice, router)
	secondChannel := signedChannel(t, "chain-expiry-second", "1000", router, bob)
	firstID := HashParts("promise", routeID, "first")
	secondID := HashParts("promise", routeID, "second")
	firstBase := signedReserveState(t, firstChannel, 2, firstChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: router, Amount: "20"},
	})
	secondBase := signedReserveState(t, secondChannel, 2, secondChannel.OpeningStateHash, "80", "0", []Balance{
		{Participant: router, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	firstPromiseChannel := firstChannel
	firstPromiseChannel.LatestState = firstBase
	secondPromiseChannel := secondChannel
	secondPromiseChannel.LatestState = secondBase
	first := signedRoutePromise(t, firstPromiseChannel, firstID, routeID, alice, router, "31", "0", 9, 70, hashLock, "", secondID)
	second := signedRoutePromise(t, secondPromiseChannel, secondID, routeID, router, bob, "30", "1", 10, 40, hashLock, firstID, "")

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, firstChannel)
	require.NoError(t, err)
	state, err = OpenChannel(state, secondChannel)
	require.NoError(t, err)
	firstConditioned, _, err := BuildConditionRootUpdateFromPromises(firstPromiseChannel, firstBase, []ConditionalPromise{first}, nil)
	require.NoError(t, err)
	secondConditioned, _, err := BuildConditionRootUpdateFromPromises(secondPromiseChannel, secondBase, []ConditionalPromise{second}, nil)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, firstChannel.ChannelID, resignState(t, firstChannel, firstConditioned), 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, secondChannel.ChannelID, resignState(t, secondChannel, secondConditioned), 20)
	require.NoError(t, err)

	_, _, err = BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof: ConditionLinkageProof{
			RouteID:	routeID,
			Promises:	[]ConditionalPromise{first, second},
			Sender:		alice,
			Receiver:	bob,
			Amount:		"30",
			TotalFees:	"1",
			HashLock:	hashLock,
			TimeoutMargin:	DefaultTimeoutMargin,
		},
		Mode:		ConditionSettlementModeExpiry,
		Resolver:	alice,
		CurrentHeight:	70,
	})
	require.ErrorContains(t, err, "has not expired")

	state, result, err := BatchSettleLinkedPromises(state, BatchConditionSettlementRequest{
		LinkageProof: ConditionLinkageProof{
			RouteID:	routeID,
			Promises:	[]ConditionalPromise{first, second},
			Sender:		alice,
			Receiver:	bob,
			Amount:		"30",
			TotalFees:	"1",
			HashLock:	hashLock,
			TimeoutMargin:	DefaultTimeoutMargin,
		},
		Mode:		ConditionSettlementModeExpiry,
		Resolver:	alice,
		CurrentHeight:	71,
	})
	require.NoError(t, err)
	require.Len(t, result.Resolutions, 2)
	require.Empty(t, result.FeeClaims)
	require.True(t, result.Resolutions[0].Expired)
	require.True(t, result.Resolutions[1].Expired)
	require.Len(t, result.ConditionRootUpdates, 2)
	require.Len(t, state.ConditionClaims, 2)
}

func TestDisputeRequestEmitsEventAndAppliesOptionalFraudProof(t *testing.T) {
	alice := testAddress(0x14)
	bob := testAddress(0x15)
	channel := signedChannel(t, "dispute-request", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "350"},
		{Participant: bob, Amount: "650"},
	})
	conflicting := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	state, err = DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	closeState.StateHash,
		NewerState:		newerState,
		FraudProof: FraudProof{
			ProofID:		HashParts("dispute-proof", channel.ChannelID),
			ProofType:		FraudProofTypeDoubleSign,
			SubmittedBy:		bob,
			OffendingSigner:	alice,
			StateA:			newerState,
			StateB:			conflicting,
			PenaltyAmount:		"25",
			EvidenceHash:		HashParts("evidence", newerState.StateHash, conflicting.StateHash),
		},
		Submitter:	bob,
		CurrentHeight:	25,
	})
	require.NoError(t, err)
	require.Equal(t, newerState.StateHash, state.Channels[0].PendingClose.State.StateHash)
	require.Len(t, state.Channels[0].PendingClose.FraudProofs, 1)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 1)
	require.Equal(t, "channel-dispute", state.Events[len(state.Events)-1].EventType)

	_, err = DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	closeState.StateHash,
		NewerState:		newerState,
		Submitter:		bob,
		CurrentHeight:		26,
	})
	require.ErrorContains(t, err, "reference")
}

func TestPaymentStateRejectsNonNaetAndCollateralMismatch(t *testing.T) {
	alice := testAddress(0x31)
	bob := testAddress(0x32)
	channel := signedChannel(t, "bad-denom", "1000", alice, bob)
	channel.Denom = "uatom"
	err := channel.Validate()
	require.ErrorContains(t, err, "naet")

	channel = signedChannel(t, "bad-collateral", "1000", alice, bob)
	badState, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			2,
		PreviousStateHash:	channel.OpeningStateHash,
		TimeoutHeight:		64,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
		Balances: []Balance{
			{Participant: alice, Amount: "999"},
			{Participant: bob, Amount: "0"},
		},
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(badState, signer)
		require.NoError(t, err)
		badState.Signatures = append(badState.Signatures, sig)
	}
	channel.LatestState = badState
	err = channel.LatestState.ValidateForChannel(channel, true)
	require.ErrorContains(t, err, "conserve")
}

func TestChannelOpenLifecycleLocksFeeAndEmitsEvent(t *testing.T) {
	alice := testAddress(0x33)
	bob := testAddress(0x34)
	req := ChannelOpenRequest{
		ChainID:			"aetra-test-1",
		ChannelID:			HashParts("open-lifecycle", alice, bob),
		Participants:			[]string{alice, bob},
		InitialBalances:		[]Balance{{Participant: alice, Amount: "700"}, {Participant: bob, Amount: "300"}},
		ChannelType:			ChannelTypeBidirectional,
		Collateral:			"1000",
		CloseDelay:			8,
		ChallengePeriod:		12,
		FeePolicyID:			NativeDenom,
		OpeningFeeDenom:		NativeDenom,
		OpeningFeePaid:			DefaultOpeningFee,
		RoutingAdvertised:		true,
		ConditionalPaymentsSupported:	true,
		OpenHeight:			11,
	}

	state, event, err := OpenChannelFromRequest(EmptyState(), req)
	require.NoError(t, err)
	require.Len(t, state.Channels, 1)
	require.Len(t, state.CustodyLocks, 1)
	require.Len(t, state.Events, 2)
	require.Equal(t, event, state.Events[0])
	require.Equal(t, "channel-open", event.EventType)
	require.Equal(t, "channel-finality-transition", state.Events[1].EventType)
	require.Equal(t, ChannelFinalityOpen, state.Channels[0].Finality)
	require.Equal(t, req.ChannelID, state.CustodyLocks[0].ChannelID)
	require.Equal(t, "1000", state.CustodyLocks[0].Amount)
	require.Equal(t, DefaultOpeningFee, state.Channels[0].OpeningFeePaid)
	require.True(t, state.Channels[0].RoutingAdvertised)
	require.True(t, state.Channels[0].ConditionalPayments)
	require.Equal(t, uint64(8), state.Channels[0].CloseDelay)
	require.Equal(t, uint64(12), state.Channels[0].DisputePeriod)

	_, _, err = OpenChannelFromRequest(state, req)
	require.ErrorContains(t, err, "already exists")

	badFee := req
	badFee.ChannelID = HashParts("open-bad-fee", alice, bob)
	badFee.OpeningFeePaid = "0"
	_, _, err = OpenChannelFromRequest(EmptyState(), badFee)
	require.ErrorContains(t, err, "opening fee")

	badDelay := req
	badDelay.ChannelID = HashParts("open-bad-delay", alice, bob)
	badDelay.CloseDelay = 0
	_, _, err = OpenChannelFromRequest(EmptyState(), badDelay)
	require.ErrorContains(t, err, "close delay")

	badChallenge := req
	badChallenge.ChannelID = HashParts("open-bad-challenge", alice, bob)
	badChallenge.ChallengePeriod = MaxChallengePeriod + 1
	_, _, err = OpenChannelFromRequest(EmptyState(), badChallenge)
	require.ErrorContains(t, err, "challenge period")

	badBalances := req
	badBalances.ChannelID = HashParts("open-bad-balances", alice, bob)
	badBalances.InitialBalances = []Balance{{Participant: alice, Amount: "999"}, {Participant: bob, Amount: "0"}}
	_, _, err = OpenChannelFromRequest(EmptyState(), badBalances)
	require.ErrorContains(t, err, "sum to collateral")
}

func TestPaymentFeeScheduleChargesStorageAndDynamicMultiplier(t *testing.T) {
	alice := testAddress(0xa0)
	bob := testAddress(0xa1)
	channel := signedChannel(t, "fee-storage-open", "100", alice, bob)
	schedule := DefaultPaymentFeeSchedule()
	schedule.ChannelOpenFee = "2"
	schedule.StorageFeeEnabled = true
	schedule.StorageByteFee = "1"
	state, err := ConfigurePaymentFeeSchedule(EmptyState(), schedule)
	require.NoError(t, err)
	state, err = SetPaymentFeeMultiplier(state, PaymentFeeMultiplier{
		FeeClass:	PaymentFeeClassChannelOpen,
		MultiplierBps:	20_000,
		CongestionBps:	5_000,
		UpdatedHeight:	channel.OpenHeight,
	})
	require.NoError(t, err)

	_, err = OpenChannel(state, channel)
	require.ErrorContains(t, err, "fee below required")
	required, storageBytes, multiplier, err := RequiredPaymentFee(state, PaymentFeeClassChannelOpen, channel)
	require.NoError(t, err)
	require.NotZero(t, storageBytes)
	require.Equal(t, uint32(20_000), multiplier)
	channel.OpeningFeePaid = required
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	require.Len(t, state.FeeCharges, 1)
	require.Equal(t, PaymentFeeClassChannelOpen, state.FeeCharges[0].FeeClass)
	require.Equal(t, required, state.FeeCharges[0].RequiredAmount)
	require.Equal(t, storageBytes, state.FeeCharges[0].StorageBytes)
}

func TestChannelOpenFeeFormulaComponentsAndBounds(t *testing.T) {
	alice := testAddress(0xa0)
	bob := testAddress(0xa1)
	channel := signedChannel(t, "fee-open-formula", "100", alice, bob)
	channel.ConditionalPayments = true
	channel.RoutingAdvertised = true

	schedule := DefaultPaymentFeeSchedule()
	schedule.ChannelOpenFee = "5"
	schedule.ChannelOpenPerParticipantFee = "2"
	schedule.StorageFeeEnabled = true
	schedule.StorageByteFee = "1"
	schedule.ConditionalCapabilitySurcharge = "7"
	schedule.RoutingAdvertisementDeposit = "11"
	schedule.StorageRentPerBlock = "3"
	schedule.RenewalPeriod = 4
	schedule.OpenFeeMin = "1"
	state, err := ConfigurePaymentFeeSchedule(EmptyState(), schedule)
	require.NoError(t, err)

	formula, err := ComputeChannelOpenFeeFormula(state, channel)
	require.NoError(t, err)
	expected := 5 + 2*len(channel.Participants) + int(formula.StorageBytes) + 7 + 11 + 3*4
	require.Equal(t, fmt.Sprintf("%d", expected), formula.TotalFee)
	require.Equal(t, "4", formula.ParticipantFee)
	require.Equal(t, "7", formula.ConditionalSurcharge)
	require.Equal(t, "11", formula.RoutingDeposit)
	require.Equal(t, "12", formula.RentReserve)
	require.Equal(t, uint64(2), formula.ParticipantCount)
	require.NotZero(t, formula.StorageBytes)

	channel.OpeningFeePaid = formula.TotalFee
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	require.Equal(t, formula.TotalFee, state.FeeCharges[0].RequiredAmount)

	boundedSchedule := schedule
	boundedSchedule.ChannelOpenFee = "100"
	boundedSchedule.ChannelOpenPerParticipantFee = "0"
	boundedSchedule.StorageFeeEnabled = false
	boundedSchedule.ConditionalCapabilitySurcharge = "0"
	boundedSchedule.RoutingAdvertisementDeposit = "0"
	boundedSchedule.StorageRentPerBlock = "0"
	boundedSchedule.RenewalPeriod = 0
	boundedSchedule.OpenFeeMax = "20"
	boundedState, err := ConfigurePaymentFeeSchedule(EmptyState(), boundedSchedule)
	require.NoError(t, err)
	boundedState, err = SetPaymentFeeMultiplier(boundedState, PaymentFeeMultiplier{
		FeeClass:	PaymentFeeClassChannelOpen,
		MultiplierBps:	15_000,
		CongestionBps:	2_500,
		UpdatedHeight:	channel.OpenHeight,
	})
	require.NoError(t, err)
	formula, err = ComputeChannelOpenFeeFormula(boundedState, channel)
	require.NoError(t, err)
	require.Equal(t, "20", formula.MaxFee)
	require.Equal(t, uint32(15_000), formula.MultiplierBps)
	require.Equal(t, "30", formula.TotalFee)
}

func TestChannelOpenFeeFormulaPreventsManySmallChannelBypass(t *testing.T) {
	schedule := DefaultPaymentFeeSchedule()
	schedule.ChannelOpenFee = "5"
	schedule.ChannelOpenPerParticipantFee = "1"
	schedule.StorageFeeEnabled = true
	schedule.StorageByteFee = "1"
	schedule.OpenFeeMin = "1"
	state, err := ConfigurePaymentFeeSchedule(EmptyState(), schedule)
	require.NoError(t, err)

	for i := 0; i < 6; i++ {
		alice := testAddress(byte(0xb0 + i*2))
		bob := testAddress(byte(0xb1 + i*2))
		channel := signedChannel(t, fmt.Sprintf("small-open-%d", i), "1", alice, bob)

		_, err = OpenChannel(state, channel)
		require.ErrorContains(t, err, "fee below required")

		formula, err := ComputeChannelOpenFeeFormula(state, channel)
		require.NoError(t, err)
		channel.OpeningFeePaid = formula.TotalFee
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	require.Len(t, state.Channels, 6)
	require.Len(t, state.FeeCharges, 6)
}

func TestPaymentFeeScheduleRejectsCloseDisputeAndRoutingBypass(t *testing.T) {
	alice := testAddress(0xa2)
	bob := testAddress(0xa3)
	channel := signedChannel(t, "fee-bypass", "100", alice, bob)
	schedule := DefaultPaymentFeeSchedule()
	schedule.UnilateralCloseFee = "3"
	schedule.DisputeFee = "4"
	schedule.RoutingAdvertisementFee = "2"
	state, err := ConfigurePaymentFeeSchedule(EmptyState(), schedule)
	require.NoError(t, err)
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	_, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "50", FeeAmount: "1", Active: true})
	require.ErrorContains(t, err, "fee below required")
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "50", FeeAmount: "1", AdvertisementFeePaid: "2", Active: true})
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	_, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.ErrorContains(t, err, "fee below required")
	pending, err := SubmitClose(state, channel.ChannelID, closeState, alice, 20, "3")
	require.NoError(t, err)
	newer := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "45"},
		{Participant: bob, Amount: "55"},
	})
	_, err = DisputeChannel(pending, ChannelDisputeRequest{
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	closeState.StateHash,
		NewerState:		newer,
		Submitter:		bob,
		CurrentHeight:		21,
		DisputeFeePaid:		"0",
	})
	require.ErrorContains(t, err, "fee below required")
	pending, err = DisputeChannel(pending, ChannelDisputeRequest{
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	closeState.StateHash,
		NewerState:		newer,
		Submitter:		bob,
		CurrentHeight:		21,
		DisputeFeePaid:		"4",
	})
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, pending.Channels[0].PendingClose.State.StateHash)
}

func TestFraudProofVerificationFeeRefundsWhenAccepted(t *testing.T) {
	alice := testAddress(0xa4)
	bob := testAddress(0xa5)
	channel := signedChannel(t, "fee-fraud-refund", "100", alice, bob)
	schedule := DefaultPaymentFeeSchedule()
	schedule.FraudProofVerificationFee = "7"
	state, err := ConfigurePaymentFeeSchedule(EmptyState(), schedule)
	require.NoError(t, err)
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	proofID := HashParts("fee-fraud-proof", channel.ChannelID)
	proof := FraudProof{
		ProofID:		proofID,
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		ComputeDisputeProofHash(FraudProof{ProofID: proofID, ProofType: FraudProofTypeDoubleSign, StateA: closeState, StateB: conflicting}),
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"10",
		VerificationFeePaid:	"0",
	}
	_, err = SubmitFraudProof(state, channel.ChannelID, proof, 21)
	require.ErrorContains(t, err, "fee below required")
	proof.VerificationFeePaid = "7"
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 21)
	require.NoError(t, err)
	require.Len(t, state.FeeRefunds, 1)
	require.Equal(t, "7", state.FeeRefunds[0].Amount)
	require.Equal(t, bob, state.FeeRefunds[0].Recipient)
	var fraudFee PaymentFeeCharge
	for _, charge := range state.FeeCharges {
		if charge.FeeClass == PaymentFeeClassFraudProofVerification {
			fraudFee = charge
		}
	}
	require.True(t, fraudFee.Refunded)
	require.Equal(t, state.FeeRefunds[0].FeeID, fraudFee.FeeID)
}

func TestFraudProofVerificationModuleDedupGasPenaltyAndRewardClaim(t *testing.T) {
	alice := testAddress(0xb1)
	bob := testAddress(0xb2)
	channel := signedChannel(t, "fraud-verification-module", "100", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	var err error
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	proofID := HashParts("fraud-module-proof", channel.ChannelID)
	proof := FraudProof{
		ProofID:		proofID,
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		HashParts("fraud-module-evidence", closeState.StateHash, conflicting.StateHash),
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"20",
	}
	module := EmptyFraudProofVerificationState()
	state, module, err = ApplyFraudProofVerificationMessage(state, module, MsgSubmitDoubleSignProof{Input: FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof:		proof,
		CurrentHeight:	21,
		Policy:		FraudPenaltyPolicy{ReporterRewardCap: "4", BurnShareBps: MaxPenaltyRouteBps},
		GasLimit:	100_000_000,
	}})
	require.NoError(t, err)
	require.Len(t, module.EvidenceRecords, 1)
	require.Len(t, module.PenaltyRecords, 1)
	require.Len(t, module.ReporterRewards, 1)
	require.Len(t, module.DoubleSignEvidence, 1)
	require.Equal(t, "4", module.ReporterRewards[0].Amount)
	require.Len(t, state.Channels[0].PendingClose.FraudProofs, 1)

	reordered := proof
	reordered.ProofID = HashParts("fraud-module-proof-reordered", channel.ChannelID)
	reordered.StateA = conflicting
	reordered.StateB = closeState
	reordered.EvidenceHash = HashParts("fraud-module-evidence-reordered", conflicting.StateHash, closeState.StateHash)
	_, _, err = ApplyFraudProofVerificationMessage(state, module, MsgSubmitDoubleSignProof{Input: FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof:		reordered,
		CurrentHeight:	22,
		Policy:		FraudPenaltyPolicy{ReporterRewardCap: "4", BurnShareBps: MaxPenaltyRouteBps},
		GasLimit:	100_000_000,
	}})
	require.ErrorContains(t, err, "duplicate fraud evidence")

	_, _, err = ApplyFraudProofVerificationMessage(state, EmptyFraudProofVerificationState(), MsgSubmitDoubleSignProof{Input: FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof:		reordered,
		CurrentHeight:	22,
		Policy:		FraudPenaltyPolicy{ReporterRewardCap: "4", BurnShareBps: MaxPenaltyRouteBps},
		GasLimit:	1,
	}})
	require.ErrorContains(t, err, "gas limit")

	_, _, err = ApplyFraudProofVerificationMessage(state, EmptyFraudProofVerificationState(), MsgSubmitDoubleSignProof{Input: FraudProofSubmission{
		ChannelID:	channel.ChannelID,
		Proof:		FraudProof{ProofID: HashParts("fraud-module-too-large"), ProofType: FraudProofTypeDoubleSign, SubmittedBy: bob, OffendingSigner: alice, StateA: closeState, StateB: conflicting, EvidenceHash: HashParts("fraud-module-too-large"), PenaltyDenom: NativeDenom, PenaltyAmount: "80"},
		CurrentHeight:	22,
		Policy:		FraudPenaltyPolicy{ReporterRewardCap: "80"},
		GasLimit:	100_000_000,
	}})
	require.ErrorContains(t, err, "exceeds available balance")

	_, module, err = ApplyFraudProofVerificationMessage(state, module, MsgClaimReporterReward{
		RewardID:	module.ReporterRewards[0].RewardID,
		Reporter:	bob,
		CurrentHeight:	23,
	})
	require.NoError(t, err)
	require.True(t, module.ReporterRewards[0].Claimed)
	_, _, err = ApplyFraudProofVerificationMessage(state, module, MsgClaimReporterReward{
		RewardID:	module.ReporterRewards[0].RewardID,
		Reporter:	bob,
		CurrentHeight:	24,
	})
	require.ErrorContains(t, err, "already claimed")
}

func FuzzCanonicalFraudEvidenceHashMalformedInputs(f *testing.F) {
	f.Add("bad-chain", "bad-channel", uint64(0), uint64(0))
	f.Add("aetra-test-1", HashParts("fuzz-channel"), uint64(1), uint64(2))
	f.Fuzz(func(t *testing.T, chainID, channelID string, epoch, nonce uint64) {
		alice := testAddress(0xb3)
		bob := testAddress(0xb4)
		channel := signedChannel(t, "fraud-fuzz", "100", alice, bob)
		channel.ChainID = chainID
		if len(channelID) == 64 {
			channel.ChannelID = channelID
		}
		proof := FraudProof{
			ProofID:		HashParts("fraud-fuzz-proof", channel.ChannelID),
			ProofType:		FraudProofTypeDoubleSign,
			SubmittedBy:		bob,
			OffendingSigner:	alice,
			StateA:			ChannelState{ChainID: chainID, ChannelID: channel.ChannelID, Epoch: epoch, Nonce: nonce, StateHash: HashParts("left", chainID, channel.ChannelID)},
			StateB:			ChannelState{ChainID: chainID, ChannelID: channel.ChannelID, Epoch: epoch, Nonce: nonce, StateHash: HashParts("right", chainID, channel.ChannelID)},
			EvidenceHash:		HashParts("fraud-fuzz-evidence", chainID, channel.ChannelID),
			PenaltyDenom:		NativeDenom,
			PenaltyAmount:		"20",
		}
		require.NoError(t, ValidateHash("fuzz canonical fraud evidence", ComputeCanonicalFraudEvidenceHash(channel, proof)))
	})
}

func FuzzPaymentRequiredFuzzVectors(f *testing.F) {
	for i, seed := range []string{
		"malformed signed state",
		"random nonce ordering",
		"conflicting same nonce",
		"invalid promise links",
		"timeout boundary",
		"batch ordering",
		"duplicate fraud encoding",
		"node queue congestion",
		"async delta aggregation",
	} {
		f.Add(seed, uint64(i), int64(i+1), seed)
	}
	f.Fuzz(func(t *testing.T, seed string, selector uint64, skew int64, reason string) {
		if len(seed) > 64 {
			seed = seed[:64]
		}
		if len(reason) > 128 {
			reason = reason[:128]
		}
		alice := testAddress(0xb5)
		router := testAddress(0xb6)
		bob := testAddress(0xb7)
		amount := fmt.Sprintf("%d", uint64(1+(absInt64(skew)%20)))
		channel := signedChannel(t, "required-fuzz-"+seed, "1000", alice, bob)

		switch selector % 9 {
		case 0:
			malformed := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "900"}, {Participant: bob, Amount: "100"}})
			malformed.SignaturePreimageHash = HashParts("malformed-signature-preimage", seed)
			require.Error(t, malformed.ValidateForChannel(channel, true))
		case 1:
			state := EmptyStateWithChannel(t, channel)
			stale := signedState(t, channel, 1, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "1000"}, {Participant: bob, Amount: "0"}})
			_, err := AcceptSignedState(state, channel.ChannelID, stale, 20)
			require.ErrorContains(t, err, "strictly increase")
		case 2:
			left := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "900"}, {Participant: bob, Amount: "100"}})
			right := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "800"}, {Participant: bob, Amount: "200"}})
			require.NotEqual(t, left.StateHash, right.StateHash)
			proof := FraudProof{ProofID: HashParts("required-fuzz-conflict", seed), ProofType: FraudProofTypeDoubleSign, StateA: left, StateB: right, EvidenceHash: HashParts("required-fuzz-evidence", seed), PenaltyDenom: NativeDenom, PenaltyAmount: amount}
			require.NoError(t, ValidateHash("required fuzz canonical conflict", ComputeCanonicalFraudEvidenceHash(channel, proof)))
		case 3:
			routeID := HashParts("required-fuzz-route", seed)
			hashLock := HashParts("required-fuzz-promise-preimage", seed)
			first := signedRoutePromise(t, channel, HashParts("required-fuzz-promise-first", seed), routeID, alice, router, "20", "0", 3, 70, hashLock, "", "")
			second := signedRoutePromise(t, channel, HashParts("required-fuzz-promise-second", seed), routeID, router, bob, "20", "1", 4, 40, hashLock, HashParts("wrong-previous", seed), "")
			err := (ConditionLinkageProof{RouteID: routeID, Promises: []ConditionalPromise{first, second}, Sender: alice, Receiver: bob, Amount: "20", TotalFees: "1", HashLock: hashLock, TimeoutMargin: DefaultTimeoutMargin}).ValidateForState(EmptyStateWithChannel(t, channel), nil)
			require.Error(t, err)
		case 4:
			base := signedReserveState(t, channel, 2, channel.OpeningStateHash, "20", "0", []Balance{{Participant: alice, Amount: "970"}, {Participant: bob, Amount: "10"}})
			promiseChannel := channel
			promiseChannel.LatestState = base
			state := EmptyStateWithChannel(t, channel)
			state, err := AcceptSignedState(state, channel.ChannelID, base, 20)
			require.NoError(t, err)
			promise := signedPromiseWithHashLock(t, promiseChannel, "required-fuzz-timeout-"+seed, alice, bob, amount, "0", 3, 30, HashParts("required-fuzz-timeout-preimage"))
			_, _, err = RevealPromisePreimage(state, PreimageRevealRequest{ChannelID: channel.ChannelID, Promises: []ConditionalPromise{promise}, Preimage: "required-fuzz-timeout-preimage", Revealer: bob, CurrentHeight: 31})
			require.ErrorContains(t, err, "timed out")
		case 5:
			firstOp := SettlementOperation{OperationID: HashParts("required-fuzz-op-first", seed), OperationType: BatchOperationSettle, ChannelID: channel.ChannelID, Nonce: 1, StateHash: channel.LatestState.StateHash}
			secondChannel := signedChannel(t, "required-fuzz-batch-second-"+seed, "1000", alice, router)
			secondOp := SettlementOperation{OperationID: HashParts("required-fuzz-op-second", seed), OperationType: BatchOperationClose, ChannelID: secondChannel.ChannelID, Nonce: 1, StateHash: secondChannel.LatestState.StateHash}
			batch, err := NewSettlementBatch(HashParts("required-fuzz-batch", seed), []SettlementOperation{secondOp, firstOp})
			require.NoError(t, err)
			tampered := batch
			tampered.RootHash = HashParts("wrong-batch-root", seed)
			require.ErrorContains(t, tampered.Validate(), "root")
		case 6:
			left := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "950"}, {Participant: bob, Amount: "50"}})
			right := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{{Participant: alice, Amount: "940"}, {Participant: bob, Amount: "60"}})
			firstProof := FraudProof{ProofID: HashParts("required-fuzz-proof-a", seed), ProofType: FraudProofTypeDoubleSign, StateA: left, StateB: right, EvidenceHash: HashParts("required-fuzz-proof-evidence", seed), PenaltyDenom: NativeDenom, PenaltyAmount: amount}
			secondProof := firstProof
			secondProof.ProofID = HashParts("required-fuzz-proof-b", seed)
			require.Equal(t, ComputeCanonicalFraudEvidenceHash(channel, firstProof), ComputeCanonicalFraudEvidenceHash(channel, secondProof))
		case 7:
			class := ClassifyRouteFailure(reason)
			require.True(t, IsRouteFailureClass(class))
			_, err := BuildRouteFailureScore(RouteFailureReport{ChannelID: channel.ChannelID, From: alice, To: bob, FailureClass: class, Retryable: true, ObservedHeight: 30}, 1, DefaultRouteFailureScoringPolicy())
			require.NoError(t, err)
		case 8:
			asyncChannel := signedAsyncChannel(t, "required-fuzz-async-"+seed, "1000", []Balance{{Participant: alice, Amount: "1000"}, {Participant: bob, Amount: "0"}}, 10, 10, "100", 70, alice, bob)
			delta := signedAsyncDelta(t, asyncChannel, "required-fuzz-delta-"+seed, alice, bob, amount, 3, 4, 40)
			checkpoint, err := BuildAsyncCheckpointState(asyncChannel, []AsyncPaymentDelta{delta}, 4, 30)
			require.NoError(t, err)
			require.NoError(t, ValidateHash("required fuzz async checkpoint", checkpoint.StateHash))
			_, err = BuildAsyncCheckpointState(asyncChannel, []AsyncPaymentDelta{delta, delta}, 5, 30)
			require.Error(t, err)
		}
	})
}

func TestSettlementGasCostsAndInclusionLatencyMonitoring(t *testing.T) {
	alice := testAddress(0xa4)
	bob := testAddress(0xa5)
	channel := signedChannel(t, "settlement-gas-monitoring", "100", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "45"},
		{Participant: bob, Amount: "55"},
	})

	schedule := DefaultSettlementGasCostSchedule()
	estimate, err := EstimateSettlementMessageGas(SettlementArbitrationInput{
		Operation:		SettlementArbitrationDispute,
		ChannelID:		channel.ChannelID,
		SignedState:		closeState,
		ConditionProofs:	[]ConditionResolution{{ConditionID: HashParts("settlement-gas-condition"), Resolver: bob, Recipient: bob, Amount: "1", EvidenceHash: HashParts("settlement-gas-evidence")}},
	}, schedule)
	require.NoError(t, err)
	require.Equal(t, SettlementArbitrationDispute, estimate.Operation)
	require.Equal(t, schedule.DisputeGas, estimate.BaseGas)
	require.Equal(t, uint64(len(closeState.Signatures))*schedule.PerSignatureGas, estimate.SignatureGas)
	require.Equal(t, schedule.PerConditionGas, estimate.ConditionGas)
	require.NotZero(t, estimate.StateByteGas)
	require.Equal(t, estimate.BaseGas+estimate.SignatureGas+estimate.ConditionGas+estimate.FraudProofGas+estimate.PenaltyAllocationGas+estimate.StateByteGas, estimate.TotalGas)

	state, err := OpenChannel(EmptyState(), channel)
	require.NoError(t, err)
	state, latency, err := RecordSettlementInclusionLatency(state, HashParts("settlement-gas-op"), channel.ChannelID, SettlementArbitrationDispute, 21, 29, 4)
	require.NoError(t, err)
	require.Equal(t, uint64(8), latency.LatencyBlocks)
	require.True(t, latency.Breached)
	require.Len(t, state.InclusionLatencies, 1)
	require.Equal(t, latency.RecordID, state.Export().InclusionLatencies[0].RecordID)
}

func TestFraudProofRefundAccountingAndSecurityReserveHook(t *testing.T) {
	alice := testAddress(0xa8)
	bob := testAddress(0xa9)
	channel := signedChannel(t, "fraud-reserve-hook", "100", alice, bob)
	schedule := DefaultPaymentFeeSchedule()
	schedule.FraudProofVerificationFee = "7"
	state, err := ConfigurePaymentFeeSchedule(EmptyState(), schedule)
	require.NoError(t, err)
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	proofID := HashParts("fraud-reserve-hook-proof", channel.ChannelID)
	proof := FraudProof{
		ProofID:		proofID,
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		ComputeDisputeProofHash(FraudProof{ProofID: proofID, ProofType: FraudProofTypeDoubleSign, StateA: closeState, StateB: conflicting}),
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"10",
		VerificationFeePaid:	"7",
	}
	state, err = SubmitFraudProofWithPolicy(state, channel.ChannelID, proof, 21, FraudPenaltyPolicy{
		ReporterRewardCap:		"4",
		SecurityReserveShareBps:	MaxPenaltyRouteBps,
		SecurityReserveHook:		true,
	})
	require.NoError(t, err)
	require.Len(t, state.FeeRefunds, 1)
	require.Equal(t, "7", state.FeeRefunds[0].Amount)
	require.Equal(t, bob, state.FeeRefunds[0].Recipient)
	require.Len(t, state.SecurityReserveHooks, 1)
	require.Equal(t, PenaltyRouteSecurityReserve, state.SecurityReserveHooks[0].Route)
	require.Equal(t, "6", state.SecurityReserveHooks[0].Amount)
	require.Equal(t, alice, state.SecurityReserveHooks[0].Offender)
	require.Equal(t, "6", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteSecurityReserve))
}

func TestPenaltyMatrixCoversFraudProofCategoriesAndBoundsBalances(t *testing.T) {
	alice := testAddress(0xaa)
	bob := testAddress(0xab)
	channel := signedChannel(t, "penalty-matrix", "100", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "50"},
		{Participant: bob, Amount: "50"},
	})
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	channel.Status = ChannelStatusPendingClose
	channel.Finality = ChannelFinalityPendingClose
	channel.PendingClose = PendingClose{
		Submitter:		alice,
		SubmittedHeight:	20,
		SettleAfterHeight:	36,
		CloseReason:		CloseReasonUnilateral,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		"0",
		State:			closeState,
	}
	for _, proofType := range []FraudProofType{
		FraudProofTypeInvalidClose,
		FraudProofTypeStaleClose,
		FraudProofTypeDoubleSign,
		FraudProofTypeInvalidBalance,
		FraudProofTypeInvalidCondition,
		FraudProofTypeReplayAttempt,
		FraudProofTypeAsyncOverexposure,
	} {
		_, err := PenaltyMatrixEntryForProof(proofType, DefaultPenaltyMatrix())
		require.NoError(t, err)
	}

	proofID := HashParts("penalty-matrix-proof", channel.ChannelID)
	proof := FraudProof{
		ProofID:		proofID,
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		ComputeDisputeProofHash(FraudProof{ProofID: proofID, ProofType: FraudProofTypeDoubleSign, StateA: closeState, StateB: conflicting}),
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"25",
	}
	accounting, err := BuildPenaltyRouteAccounting(channel, proof, DefaultPenaltyMatrix(), FraudPenaltyPolicy{})
	require.NoError(t, err)
	require.Equal(t, PenaltyClassDoubleSign, accounting.Class)
	require.Equal(t, PenaltySourceParticipantBond, accounting.Source)
	require.Equal(t, "6", accounting.CounterpartyComp)
	require.Equal(t, "8", accounting.ReporterReward)
	require.Equal(t, "4", allocationAmountFor(accounting.Allocations, PenaltyRouteSecurityReserve))

	finalBalances, err := applySettlementAdjustments(closeState.Balances, accounting.Penalties, accounting.Allocations, "0", alice)
	require.NoError(t, err)
	for _, balance := range finalBalances {
		amount, err := parseNonNegativeInt("test final balance", balance.Amount)
		require.NoError(t, err)
		require.False(t, amount.IsNegative())
	}

	invalidProofPenalty, err := ComputeInvalidFraudProofSubmissionPenalty(bob, "7", "3")
	require.NoError(t, err)
	require.Equal(t, "3", invalidProofPenalty.ForfeitedAmount)
	require.Equal(t, "4", invalidProofPenalty.RefundAmount)
	invalidProofPenalty, err = ComputeInvalidFraudProofSubmissionPenalty(bob, "2", "3")
	require.NoError(t, err)
	require.Equal(t, "2", invalidProofPenalty.ForfeitedAmount)
	require.Equal(t, "0", invalidProofPenalty.RefundAmount)
}

func TestLiquidityAdvertisementReservationScoreAndDepositPenalty(t *testing.T) {
	alice := testAddress(0xac)
	bob := testAddress(0xad)
	channelID := HashParts("liquidity-ad-channel")
	ad, err := BuildLiquidityAdvertisement(LiquidityAdvertisement{
		ChannelID:		channelID,
		Advertiser:		alice,
		Counterparty:		bob,
		Capacity:		"1000",
		FeeDenom:		NativeDenom,
		BaseFee:		"1",
		ReservationFee:		"3",
		VirtualSetupFee:	"5",
		ReliabilityBps:		9_000,
		ValidUntilHeight:	50,
		DepositAmount:		"9",
		BackedByReservation:	true,
	}, "5")
	require.NoError(t, err)
	require.NotEmpty(t, ad.AdvertisementHash)

	_, err = BuildLiquidityAdvertisement(LiquidityAdvertisement{
		ChannelID:		channelID,
		Advertiser:		alice,
		Counterparty:		bob,
		Capacity:		"1000",
		FeeDenom:		NativeDenom,
		ValidUntilHeight:	50,
		DepositAmount:		"4",
	}, "5")
	require.ErrorContains(t, err, "deposit below required")

	reservation, err := BuildSignedLiquidityReservation(SignedLiquidityReservation{
		AdvertisementID:	ad.AdvertisementID,
		ChainID:		"aetra-test-chain",
		ChannelID:		channelID,
		Reserver:		alice,
		Counterparty:		bob,
		Capacity:		"700",
		FeeAmount:		"3",
		ExpirationHeight:	45,
		Nonce:			1,
	}, alice)
	require.NoError(t, err)
	require.NoError(t, reservation.Validate())
	require.Equal(t, reservation.CommitmentHash, reservation.Signature.CommitmentHash)

	score, err := LiquidityAvailabilityScore(ad, EdgeRoutingStats{SuccessRateBps: 9_000, CongestionBps: 1_000})
	require.NoError(t, err)
	require.Greater(t, score, int64(0))
	store, forfeited, err := ApplyFalseLiquidityAdvertisementPenalty(TopologyStore{}, ad, 60)
	require.NoError(t, err)
	require.Equal(t, "9", forfeited)
	require.Equal(t, -InvalidGossipPenalty, RoutingScoreForEdge(store, ChannelEdge{ChannelID: channelID, From: alice, To: bob, Capacity: "100", FeeAmount: "1", Active: true}))
}

func TestLiquidityOptimizationModuleReservationsForecastsFeesAndDecay(t *testing.T) {
	alice := testAddress(0xae)
	bob := testAddress(0xaf)
	channel := signedChannel(t, "liquidity-optimization", "1000", alice, bob)
	chain := EmptyStateWithChannel(t, channel)
	state := EmptyLiquidityOptimizationState()
	var err error

	state, err = ApplyLiquidityOptimizationMessage(chain, state, MsgSetLiquidityLimits{
		Signer:	alice,
		Limits: LiquidityLimits{
			ChannelID:		channel.ChannelID,
			Participant:		alice,
			MaxReservedCapacity:	"700",
			MinAvailableCapacity:	"100",
			MaxBaseFee:		"5",
			MaxReservationFee:	"3",
			MaxVirtualSetupFee:	"8",
			MaxProportionalBps:	500,
			MaxRebalanceLoad:	10,
		},
		CurrentHeight:	12,
	})
	require.NoError(t, err)

	ad := LiquidityAdvertisement{
		ChannelID:		channel.ChannelID,
		Advertiser:		alice,
		Counterparty:		bob,
		Capacity:		"600",
		FeeDenom:		NativeDenom,
		BaseFee:		"2",
		ReservationFee:		"3",
		VirtualSetupFee:	"5",
		ReliabilityBps:		9_500,
		ValidUntilHeight:	80,
		DepositAmount:		"9",
		BackedByReservation:	true,
	}
	state, err = ApplyLiquidityOptimizationMessage(chain, state, MsgAdvertiseLiquidity{
		Signer:			alice,
		Advertisement:		ad,
		RequiredDeposit:	"5",
		CurrentHeight:		13,
	})
	require.NoError(t, err)
	require.Len(t, state.Positions, 1)
	require.Equal(t, "600", state.Positions[0].AvailableCapacity)
	require.Len(t, state.Forecasts, 1)
	require.Len(t, state.Scores, 1)

	reservation, err := BuildSignedLiquidityReservation(SignedLiquidityReservation{
		AdvertisementID:	state.Positions[0].FeePolicyID,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Reserver:		alice,
		Counterparty:		bob,
		Capacity:		"400",
		FeeAmount:		"3",
		ExpirationHeight:	30,
		Nonce:			1,
	}, alice)
	require.NoError(t, err)
	state, err = ApplyLiquidityOptimizationMessage(chain, state, MsgReserveLiquidity{
		Reservation:	reservation,
		CurrentHeight:	14,
	})
	require.NoError(t, err)
	require.Equal(t, "400", state.Positions[0].ReservedCapacity)
	require.Equal(t, "200", state.Positions[0].AvailableCapacity)
	require.Equal(t, "400", state.Forecasts[0].ReservedCapacity)

	overReserve, err := BuildSignedLiquidityReservation(SignedLiquidityReservation{
		AdvertisementID:	state.Positions[0].FeePolicyID,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Reserver:		alice,
		Counterparty:		bob,
		Capacity:		"300",
		FeeAmount:		"3",
		ExpirationHeight:	30,
		Nonce:			2,
	}, alice)
	require.NoError(t, err)
	_, err = ApplyLiquidityOptimizationMessage(chain, state, MsgReserveLiquidity{Reservation: overReserve, CurrentHeight: 15})
	require.ErrorContains(t, err, "over-reservation")

	policy, err := BuildRoutingFeePolicyUpdate(RoutingFeePolicyUpdate{
		ChainID:			channel.ChainID,
		ChannelID:			channel.ChannelID,
		From:				alice,
		To:				bob,
		FeeDenom:			NativeDenom,
		BaseHopFee:			"2",
		ProportionalFeeBps:		100,
		LiquidityReservationFee:	"3",
		VirtualChannelSetupFee:		"5",
		CongestionSurcharge:		"1",
		FailurePenalty:			"1",
		MaxHopFee:			"5",
		ValidAfterHeight:		16,
		ValidUntilHeight:		50,
		Sequence:			1,
	}, alice)
	require.NoError(t, err)
	state, err = ApplyLiquidityOptimizationMessage(chain, state, MsgUpdateFeePolicy{
		Signer:	alice,
		Policy:	policy,
		Bounds: LiquidityFeePolicyBounds{
			MaxBaseFee:		"5",
			MaxReservationFee:	"3",
			MaxVirtualSetupFee:	"8",
			MaxCongestionFee:	"2",
			MaxFailurePenalty:	"2",
			MaxHopFee:		"8",
			MaxProportionalBps:	500,
			MinValidityWindow:	8,
			MaxValidityWindow:	80,
		},
		CurrentHeight:	16,
	})
	require.NoError(t, err)
	require.Len(t, state.FeePolicies, 1)

	tooHigh := policy
	tooHigh.BaseHopFee = "9"
	tooHigh.PolicyHash = ComputeRoutingFeePolicyHash(tooHigh)
	tooHigh.Signature, err = SignatureForRoutingFeePolicy(tooHigh, alice)
	require.NoError(t, err)
	_, err = ApplyLiquidityOptimizationMessage(chain, state, MsgUpdateFeePolicy{
		Signer:		alice,
		Policy:		tooHigh,
		Bounds:		LiquidityFeePolicyBounds{MaxBaseFee: "5"},
		CurrentHeight:	17,
	})
	require.ErrorContains(t, err, "bounds")

	state, err = ApplyLiquidityOptimizationMessage(chain, state, MsgSubmitRebalanceIntent{
		Signer:	alice,
		Intent: RebalanceIntent{
			ChannelID:		channel.ChannelID,
			Owner:			alice,
			TargetCapacity:		"500",
			MaxSettlementLoad:	8,
			Priority:		1,
			ExpiresHeight:		60,
		},
		CurrentHeight:	18,
	})
	require.NoError(t, err)
	require.Len(t, state.RebalanceIntents, 1)

	_, err = ApplyLiquidityOptimizationMessage(chain, state, MsgSubmitRebalanceIntent{
		Signer:	alice,
		Intent: RebalanceIntent{
			ChannelID:		channel.ChannelID,
			Owner:			alice,
			TargetCapacity:		"500",
			MaxSettlementLoad:	11,
			Priority:		1,
			ExpiresHeight:		60,
		},
		CurrentHeight:	19,
	})
	require.ErrorContains(t, err, "settlement load")

	state, expired, err := ExpireLiquidityReservations(chain, state, 31)
	require.NoError(t, err)
	require.Len(t, expired, 1)
	require.True(t, state.Reservations[0].Released)
	require.Equal(t, "0", state.Positions[0].ReservedCapacity)
	require.Equal(t, "600", state.Positions[0].AvailableCapacity)

	before := state.Scores[0].Score
	state, err = DecayLiquidityScores(state, 13+DefaultGossipTTL, DefaultGossipTTL)
	require.NoError(t, err)
	require.Less(t, state.Scores[0].Score, before)
}

func TestAsyncExecutionFinalizationQueueIsBoundedAndIdempotent(t *testing.T) {
	alice := testAddress(0xa6)
	bob := testAddress(0xa7)
	first := signedChannel(t, "async-finalize-first", "100", alice, bob)
	second := signedChannel(t, "async-finalize-second", "100", alice, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	firstClose := signedState(t, first, 2, first.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	secondClose := signedState(t, second, 2, second.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "45"},
		{Participant: bob, Amount: "55"},
	})
	state, err = SubmitClose(state, first.ChannelID, firstClose, alice, 20, "0")
	require.NoError(t, err)
	state, err = SubmitClose(state, second.ChannelID, secondClose, alice, 20, "0")
	require.NoError(t, err)

	state, result, err := ProcessAsyncExecutionQueues(state, 28, 1, 0)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ProcessedFinalizations)
	require.Len(t, result.CompletedJobIDs, 1)
	require.Len(t, state.Settlements, 1)
	require.Len(t, state.AsyncCompletions, 1)
	require.Contains(t, paymentEventTypes(state.Events), "async-settlement-completion")

	state, result, err = ProcessAsyncExecutionQueues(state, 28, 1, 0)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ProcessedFinalizations)
	require.Len(t, state.Settlements, 2)
	require.Len(t, state.AsyncCompletions, 2)

	state, result, err = ProcessAsyncExecutionQueues(state, 28, 10, 0)
	require.NoError(t, err)
	require.Zero(t, result.ProcessedFinalizations)
	require.Len(t, state.Settlements, 2)
	require.Len(t, state.AsyncCompletions, 2)
	for _, job := range state.AsyncFinalizationQueue {
		require.True(t, job.Completed)
		require.NotEmpty(t, job.SettlementHash)
	}
}

func TestAsyncExecutionExpiredPromiseQueueIsBoundedAndRetriable(t *testing.T) {
	alice := testAddress(0xa8)
	bob := testAddress(0xa9)
	channel := signedChannel(t, "async-promise-expiry", "1000", alice, bob)
	base := signedReserveState(t, channel, 2, channel.OpeningStateHash, "80", "0", []Balance{
		{Participant: alice, Amount: "900"},
		{Participant: bob, Amount: "20"},
	})
	promiseChannel := channel
	promiseChannel.LatestState = base
	firstPromise := signedLinkedPromise(t, promiseChannel, HashParts("async-expiry-promise", channel.ChannelID, "first"), alice, bob, "10", "0", 9, 40, HashParts("async-expiry-one"), "", "")
	secondPromise := signedLinkedPromise(t, promiseChannel, HashParts("async-expiry-promise", channel.ChannelID, "second"), alice, bob, "10", "0", 10, 41, HashParts("async-expiry-two"), "", "")
	conditioned, _, err := BuildConditionRootUpdateFromPromises(promiseChannel, base, []ConditionalPromise{firstPromise, secondPromise}, nil)
	require.NoError(t, err)
	conditioned = resignState(t, channel, conditioned)
	state := EmptyState()
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, channel.ChannelID, conditioned, 20)
	require.NoError(t, err)
	state, firstJob, err := EnqueueExpiredPromise(state, firstPromise, alice, 21)
	require.NoError(t, err)
	state, secondJob, err := EnqueueExpiredPromise(state, secondPromise, alice, 21)
	require.NoError(t, err)
	require.NotEqual(t, firstJob.JobID, secondJob.JobID)

	state, result, err := ProcessAsyncExecutionQueues(state, 41, 0, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ProcessedPromiseExpiries)
	require.Len(t, state.ConditionClaims, 1)
	require.Len(t, state.AsyncCompletions, 1)

	state, result, err = ProcessAsyncExecutionQueues(state, 42, 0, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ProcessedPromiseExpiries)
	require.Len(t, state.ConditionClaims, 2)
	require.Len(t, state.AsyncCompletions, 2)

	state, result, err = ProcessAsyncExecutionQueues(state, 42, 0, 10)
	require.NoError(t, err)
	require.Zero(t, result.ProcessedPromiseExpiries)
	require.Len(t, state.ConditionClaims, 2)
	require.Len(t, state.AsyncCompletions, 2)
	for _, job := range state.AsyncPromiseExpiryQueue {
		require.True(t, job.Completed)
		require.NotEmpty(t, job.ResolutionHash)
	}
}

func TestBidirectionalStateCommitmentIncludesDomainFields(t *testing.T) {
	alice := testAddress(0x37)
	bob := testAddress(0x38)
	channel := signedChannel(t, "bidirectional-domain", "1000", alice, bob)
	channel = channel.Normalize()

	condition := ConditionalPayment{
		ConditionID:	HashParts("condition", channel.ChannelID),
		ConditionType:	ConditionTypeHashLock,
		Payer:		channel.Participants[0],
		Payee:		channel.Participants[1],
		Amount:		"40",
		HashLock:	HashParts("preimage"),
		TimeoutHeight:	88,
		NonceStart:	2,
		NonceEnd:	5,
	}
	state, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			2,
		PreviousStateHash:	channel.OpeningStateHash,
		Balances:		[]Balance{{Participant: channel.Participants[0], Amount: "460"}, {Participant: channel.Participants[1], Amount: "500"}},
		ReserveA:		"25",
		ReserveB:		"15",
		Conditions:		[]ConditionalPayment{condition},
		TimeoutHeight:		96,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	require.Equal(t, ComputeConditionsRoot(state.Conditions), state.PendingConditionsRoot)

	changedTimeout := state
	changedTimeout.TimeoutHeight++
	changedTimeout.StateHash = ""
	changedTimeout.Signatures = nil
	changedTimeout, err = BuildState(changedTimeout)
	require.NoError(t, err)
	require.NotEqual(t, state.StateHash, changedTimeout.StateHash)

	changedReserve := state
	changedReserve.ReserveB = "16"
	changedReserve.StateHash = ""
	changedReserve.Signatures = nil
	changedReserve, err = BuildState(changedReserve)
	require.NoError(t, err)
	require.NotEqual(t, state.StateHash, changedReserve.StateHash)

	badRoot := state
	badRoot.PendingConditionsRoot = HashParts("wrong-root")
	badRoot.StateHash = ComputeStateHash(badRoot)
	for i := range badRoot.Signatures {
		sig, err := SignatureForState(badRoot, badRoot.Signatures[i].Signer)
		require.NoError(t, err)
		badRoot.Signatures[i] = sig
	}
	require.ErrorContains(t, badRoot.ValidateForChannel(channel, true), "conditions root")
}

func TestCanonicalChannelStateIncludesAllStateDomains(t *testing.T) {
	alice := testAddress(0x3b)
	bob := testAddress(0x3c)
	channel := signedChannel(t, "canonical-state", "1000", alice, bob)
	channel = channel.Normalize()

	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "475"},
		{Participant: bob, Amount: "525"},
	})

	require.Equal(t, CurrentAppVersion, state.AppVersion)
	require.Equal(t, ModuleName, state.ModuleName)
	require.Equal(t, ComputeParticipantSetHash(channel.Participants), state.ParticipantSetHash)
	require.Equal(t, "0", state.AccruedFees)
	require.Equal(t, channel.DisputePeriod, state.ChallengePeriod)
	require.Equal(t, ComputeConditionsRoot(state.Conditions), state.ConditionRoot)
	require.Equal(t, state.PendingConditionsRoot, state.ConditionRoot)
	require.Equal(t, uint32(len(state.Conditions)), state.ConditionCount)
	require.Equal(t, ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners), state.RequiredSignerBitmap)
	require.Equal(t, SignatureSchemeEd25519, state.SignatureScheme)
	require.Equal(t, ComputeStateSignaturePreimageHash(state), state.SignaturePreimageHash)

	changedFees := state
	changedFees.AccruedFees = "1"
	changedFees.StateHash = ""
	changedFees.Signatures = nil
	changedFees, err := BuildState(changedFees)
	require.NoError(t, err)
	require.NotEqual(t, state.StateHash, changedFees.StateHash)

	badParticipantSet := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.ParticipantSetHash = HashParts("wrong-participant-set")
	}))
	require.ErrorContains(t, badParticipantSet.ValidateForChannel(channel, true), "participant set")

	badChallenge := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.ChallengePeriod++
	}))
	require.ErrorContains(t, badChallenge.ValidateForChannel(channel, true), "challenge period")

	badScheme := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.SignatureScheme = "secp256k1"
	}))
	require.ErrorContains(t, badScheme.ValidateForChannel(channel, true), "signature scheme")
}

func TestStateHashEncodingVersionAndDomainSeparation(t *testing.T) {
	alice := testAddress(0x3f)
	bob := testAddress(0x40)
	channel := signedChannel(t, "state-hash-version", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	v1Hash, err := ComputeStateHashForEncodingVersion(state, CanonicalEncodingVersion)
	require.NoError(t, err)
	require.Equal(t, state.StateHash, v1Hash)
	_, err = ComputeStateHashForEncodingVersion(state, CanonicalEncodingVersion+1)
	require.ErrorContains(t, err, "unsupported")

	condition := ConditionalPayment{
		ConditionID:	HashParts("promise", channel.ChannelID),
		ConditionType:	ConditionTypeHashLock,
		Payer:		alice,
		Payee:		bob,
		Amount:		"10",
		HashLock:	HashParts("promise-preimage"),
		TimeoutHeight:	64,
		NonceStart:	2,
		NonceEnd:	3,
	}
	async := signedAsyncChannel(t, "domain-async", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	delta := signedAsyncDelta(t, async, "domain-delta", alice, bob, "5", 2, 2, 70)
	conflicting := signedState(t, channel, state.Nonce, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "510"},
		{Participant: bob, Amount: "490"},
	})
	proof := FraudProof{
		ProofID:		HashParts("domain-proof", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			state,
		StateB:			conflicting,
		PenaltyAmount:		"10",
		EvidenceHash:		HashParts("evidence", state.StateHash, conflicting.StateHash),
	}
	vc := VirtualChannel{
		VirtualChannelID:	HashParts("domain-vc", alice, bob),
		ChainID:		channel.ChainID,
		Nonce:			1,
		ParentChannelIDs:	[]string{channel.ChannelID},
		Endpoints:		[]string{alice, bob},
		Capacity:		"100",
		ExpiresHeight:		90,
	}
	vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	vc.StateHash = ComputeVirtualChannelStateHash(vc)

	hashes := []string{
		state.StateHash,
		delta.DeltaHash,
		ComputeConditionalPromiseHash(condition),
		ComputeCooperativeCloseHash(channel.ChainID, channel.ChannelID, state.StateHash, state.Nonce),
		ComputeDisputeProofHash(proof),
		vc.StateHash,
	}
	seen := map[string]struct{}{}
	for _, hash := range hashes {
		require.NotContains(t, seen, hash)
		seen[hash] = struct{}{}
	}
}

func TestStateRejectsUnknownRequiredFields(t *testing.T) {
	alice := testAddress(0x41)
	bob := testAddress(0x42)
	channel := signedChannel(t, "unknown-required-field", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	bad := resignState(t, channel, mutateCanonicalState(state, func(s *ChannelState) {
		s.RequiredFields = append(s.RequiredFields, "future_required_field")
	}))
	require.ErrorContains(t, bad.ValidateForChannel(channel, true), "unknown required field")
}

func TestCommitmentModelBindsChannelDomainAndPayloads(t *testing.T) {
	alice := testAddress(0x59)
	bob := testAddress(0x5a)
	first := signedChannel(t, "commitment-first", "1000", alice, bob)
	second := signedChannel(t, "commitment-second", "1000", alice, bob)

	firstState := signedConditionalState(t, first, 2, first.OpeningStateHash, "25", []Balance{
		{Participant: alice, Amount: "975"},
		{Participant: bob, Amount: "0"},
	})
	secondState := firstState
	secondState.ChannelID = second.ChannelID
	secondState.PreviousStateHash = second.OpeningStateHash
	secondState.ParticipantSetHash = ComputeParticipantSetHash(second.Participants)
	secondState.StateHash = ""
	secondState.Signatures = nil
	var err error
	secondState, err = BuildState(secondState)
	require.NoError(t, err)

	require.NotEqual(t, ComputeOpeningCommitment(first), ComputeOpeningCommitment(second))
	require.NotEqual(t, ComputeBalanceStateCommitment(first, firstState), ComputeBalanceStateCommitment(second, secondState))
	require.NotEqual(t, ComputeConditionRootCommitment(first, firstState), ComputeConditionRootCommitment(second, secondState))
	require.Equal(t, ComputeConditionsRoot(firstState.Conditions), firstState.ConditionRoot)

	asyncFirst := signedAsyncChannel(t, "commitment-async-first", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	asyncSecond := signedAsyncChannel(t, "commitment-async-second", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	delta := signedAsyncDelta(t, asyncFirst, "commitment-delta", alice, bob, "10", 2, 2, 70)
	require.NotEqual(t, ComputeAsyncDeltaRootForChannel(asyncFirst, []AsyncPaymentDelta{delta}), ComputeAsyncDeltaRootForChannel(asyncSecond, []AsyncPaymentDelta{delta}))

	vc := VirtualChannel{
		VirtualChannelID:	HashParts("commitment-vc", alice, bob),
		ChainID:		first.ChainID,
		Nonce:			1,
		ParentChannelIDs:	[]string{first.ChannelID},
		Endpoints:		[]string{alice, bob},
		Capacity:		"100",
		ExpiresHeight:		90,
	}
	vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	changedVC := vc
	changedVC.Capacity = "101"
	require.NotEqual(t, vc.AnchorCommitment, ComputeVirtualChannelAnchor(changedVC))
	changedVC = vc
	changedVC.ExpiresHeight++
	require.NotEqual(t, vc.AnchorCommitment, ComputeVirtualChannelAnchor(changedVC))

	settlement := SettlementRecord{
		ChainID:		first.ChainID,
		ChannelID:		first.ChannelID,
		StateHash:		firstState.StateHash,
		Nonce:			firstState.Nonce,
		FinalBalances:		firstState.Balances,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		"0",
		Penalties:		[]Penalty{{Offender: alice, Recipient: bob, Denom: NativeDenom, Amount: "1"}},
		SettledHeight:		100,
	}
	penaltyRoute := settlement
	penaltyRoute.Penalties = []Penalty{{Offender: bob, Recipient: alice, Denom: NativeDenom, Amount: "1"}}
	otherDomain := settlement
	otherDomain.ChannelID = second.ChannelID
	otherDomain.ChainID = "aetra-test-2"
	require.NotEqual(t, ComputeSettlementResultCommitment(first, settlement), ComputeSettlementResultCommitment(first, penaltyRoute))
	require.NotEqual(t, ComputeSettlementHash(settlement), ComputeSettlementHash(otherDomain))
}

func TestSignatureEnvelopeRejectsReplayAndWrongCommitment(t *testing.T) {
	alice := testAddress(0x5b)
	bob := testAddress(0x5c)
	channel := signedChannel(t, "signature-envelope", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	wrongChain := state
	wrongChain.Signatures = append([]StateSignature(nil), state.Signatures...)
	wrongChain.Signatures[0].ChainID = "other-chain"
	wrongChain.Signatures[0].SignatureHash = ComputeSignatureEnvelopeHash(
		wrongChain.Signatures[0].Signer,
		wrongChain.Signatures[0].ChainID,
		wrongChain.Signatures[0].ChannelID,
		wrongChain.Signatures[0].ObjectType,
		wrongChain.Signatures[0].Version,
		wrongChain.Signatures[0].Nonce,
		wrongChain.Signatures[0].ObjectID,
		wrongChain.Signatures[0].ExpirationHeight,
		wrongChain.Signatures[0].CommitmentHash,
	)
	require.ErrorContains(t, wrongChain.ValidateForChannel(channel, true), "chain id")

	wrongCommitment := state
	wrongCommitment.Signatures = append([]StateSignature(nil), state.Signatures...)
	wrongCommitment.Signatures[0].CommitmentHash = HashParts("wrong-commitment")
	wrongCommitment.Signatures[0].SignatureHash = ComputeSignatureEnvelopeHash(
		wrongCommitment.Signatures[0].Signer,
		wrongCommitment.Signatures[0].ChainID,
		wrongCommitment.Signatures[0].ChannelID,
		wrongCommitment.Signatures[0].ObjectType,
		wrongCommitment.Signatures[0].Version,
		wrongCommitment.Signatures[0].Nonce,
		wrongCommitment.Signatures[0].ObjectID,
		wrongCommitment.Signatures[0].ExpirationHeight,
		wrongCommitment.Signatures[0].CommitmentHash,
	)
	require.ErrorContains(t, wrongCommitment.ValidateForChannel(channel, true), "commitment")

	duplicate := state
	duplicate.Signatures = append([]StateSignature(nil), state.Signatures...)
	duplicate.Signatures = append(duplicate.Signatures, duplicate.Signatures[0])
	duplicate.Signatures = normalizeSignatures(duplicate.Signatures)
	require.ErrorContains(t, duplicate.ValidateForChannel(channel, true), "duplicate")
}

func TestClaimAndDeltaSignatureEnvelopeValidation(t *testing.T) {
	payer := testAddress(0x5d)
	receiver := testAddress(0x5e)
	channel := signedUnidirectionalChannel(t, "claim-envelope", "1000", payer, receiver, false)
	claim := signedUnidirectionalClaim(t, channel, "100", 2, 80, false)

	wrongChannel := claim
	wrongChannel.PayerSignature.ChannelID = HashParts("wrong-channel")
	wrongChannel.PayerSignature.SignatureHash = ComputeSignatureEnvelopeHash(
		wrongChannel.PayerSignature.Signer,
		wrongChannel.PayerSignature.ChainID,
		wrongChannel.PayerSignature.ChannelID,
		wrongChannel.PayerSignature.ObjectType,
		wrongChannel.PayerSignature.Version,
		wrongChannel.PayerSignature.Nonce,
		wrongChannel.PayerSignature.ObjectID,
		wrongChannel.PayerSignature.ExpirationHeight,
		wrongChannel.PayerSignature.CommitmentHash,
	)
	require.ErrorContains(t, wrongChannel.ValidateForChannel(channel), "channel id")

	alice := testAddress(0x5f)
	bob := testAddress(0x60)
	async := signedAsyncChannel(t, "delta-envelope", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)
	delta := signedAsyncDelta(t, async, "delta-envelope", alice, bob, "10", 2, 2, 70)

	wrongType := delta
	wrongType.Signature.ObjectType = SignatureObjectClaim
	wrongType.Signature.SignatureHash = ComputeSignatureEnvelopeHash(
		wrongType.Signature.Signer,
		wrongType.Signature.ChainID,
		wrongType.Signature.ChannelID,
		wrongType.Signature.ObjectType,
		wrongType.Signature.Version,
		wrongType.Signature.Nonce,
		wrongType.Signature.ObjectID,
		wrongType.Signature.ExpirationHeight,
		wrongType.Signature.CommitmentHash,
	)
	require.ErrorContains(t, wrongType.ValidateForChannel(async, 30), "object type")
	require.ErrorContains(t, delta.ValidateForChannel(async, 71), "expired")
}

func TestLocalSignerWriteAheadPreventsDoubleSign(t *testing.T) {
	alice := testAddress(0x61)
	bob := testAddress(0x62)
	channel := signedChannel(t, "signer-wal", "1000", alice, bob)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})

	records, sig, err := SignStateWithWriteAhead(nil, state, alice, SignerIsolationHardware)
	require.NoError(t, err)
	require.Equal(t, alice, sig.Signer)
	require.Len(t, records, 1)
	require.True(t, records[0].Released)
	require.Equal(t, SignerIsolationHardware, records[0].IsolationMode)
	require.Equal(t, ComputeSignedNonceWALHash(records[0]), records[0].WALHash)

	records, _, err = SignStateWithWriteAhead(records, state, alice, SignerIsolationHardware)
	require.NoError(t, err)
	require.Len(t, records, 1)

	next := signedState(t, channel, 3, state.StateHash, []Balance{
		{Participant: alice, Amount: "480"},
		{Participant: bob, Amount: "520"},
	})
	persistence := SignerPersistence{Records: records, IsolationMode: SignerIsolationHardware}
	persistence, _, err = persistence.SignState(next, alice)
	require.NoError(t, err)
	require.Equal(t, uint64(3), persistence.HighestSignedNonce(alice, channel.ChainID, channel.ChannelID, 1))
	_, _, err = SignStateWithWriteAhead(persistence.Records, state, alice, SignerIsolationHardware)
	require.ErrorContains(t, err, "below highest signed nonce")

	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "510"},
		{Participant: bob, Amount: "490"},
	})
	_, _, err = SignStateWithWriteAhead(records, conflicting, alice, SignerIsolationHardware)
	require.ErrorContains(t, err, "same nonce replacement")
}

func TestPaymentSignerAPIEnforcesLimitsPauseAndAuditLog(t *testing.T) {
	alice := testAddress(0x65)
	bob := testAddress(0x66)
	gossipKey := testAddress(0x67)
	channel := signedChannel(t, "signer-api-limits", "1000", alice, bob)
	stateTwo := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "990"},
		{Participant: bob, Amount: "10"},
	})
	stateThree := signedState(t, channel, 3, stateTwo.StateHash, []Balance{
		{Participant: alice, Amount: "970"},
		{Participant: bob, Amount: "30"},
	})

	api, err := NewPaymentSignerAPI(PaymentSignerConfig{
		Signer:			alice,
		KeyRole:		SignerKeyRoleParticipant,
		FundsKey:		alice,
		GossipKey:		gossipKey,
		IsolationMode:		SignerIsolationHardware,
		AutomatedSigning:	true,
		MaxAutomatedAmount:	"25",
		MaxAutomatedPerBlock:	2,
		ChannelLimits: []ChannelSigningLimit{{
			ChannelID:		channel.ChannelID,
			MaxNonce:		3,
			MaxAmount:		"25",
			MaxSignatures:		2,
			ValidUntilHeight:	100,
		}},
	}, SignerPersistence{}, nil)
	require.NoError(t, err)

	api, resp, err := api.SignState(SignStateRequest{
		State:		stateTwo,
		Signer:		alice,
		KeyRole:	SignerKeyRoleParticipant,
		Amount:		"10",
		Automated:	true,
		CurrentHeight:	20,
	})
	require.NoError(t, err)
	require.True(t, resp.WALRecord.Released)
	require.Equal(t, SignerIsolationHardware, resp.WALRecord.IsolationMode)
	require.Equal(t, resp.Signature.SignatureHash, resp.AuditLog.SignatureHash)
	require.Equal(t, ComputeSignedStateAuditHash(resp.AuditLog), resp.AuditLog.AuditHash)
	require.NoError(t, resp.AuditLog.Validate())
	require.Equal(t, uint64(2), api.NonceStore.HighestSignedNonce(alice, channel.ChainID, channel.ChannelID, 1))
	require.Len(t, api.AuditLogs, 1)

	api, _, err = api.SignState(SignStateRequest{
		State:		stateThree,
		Signer:		alice,
		KeyRole:	SignerKeyRoleParticipant,
		Amount:		"20",
		Automated:	true,
		CurrentHeight:	20,
	})
	require.NoError(t, err)

	stateFour := signedState(t, channel, 4, stateThree.StateHash, []Balance{
		{Participant: alice, Amount: "960"},
		{Participant: bob, Amount: "40"},
	})
	signatureLimited := api
	signatureLimited.Config.ChannelLimits[0].MaxNonce = 5
	_, _, err = signatureLimited.SignState(SignStateRequest{
		State:		stateFour,
		Signer:		alice,
		KeyRole:	SignerKeyRoleParticipant,
		Amount:		"10",
		Automated:	true,
		CurrentHeight:	21,
	})
	require.ErrorContains(t, err, "signature limit")

	nonceLimited := api
	nonceLimited.Config.ChannelLimits = []ChannelSigningLimit{{
		ChannelID:		channel.ChannelID,
		MaxNonce:		3,
		MaxAmount:		"25",
		ValidUntilHeight:	100,
	}}
	_, _, err = nonceLimited.SignState(SignStateRequest{
		State:		stateFour,
		Signer:		alice,
		KeyRole:	SignerKeyRoleParticipant,
		Amount:		"10",
		Automated:	true,
		CurrentHeight:	21,
	})
	require.ErrorContains(t, err, "nonce limit")

	amountLimited := api
	amountLimited.Config.ChannelLimits = []ChannelSigningLimit{{
		ChannelID:		channel.ChannelID,
		MaxNonce:		5,
		MaxAmount:		"25",
		ValidUntilHeight:	100,
	}}
	_, _, err = amountLimited.SignState(SignStateRequest{
		State:		stateFour,
		Signer:		alice,
		KeyRole:	SignerKeyRoleParticipant,
		Amount:		"26",
		Automated:	true,
		CurrentHeight:	21,
	})
	require.ErrorContains(t, err, "amount limit")

	pausedConfig, err := EmergencyPauseSigner(api.Config, alice, channel.ChannelID, "compromised key", 25)
	require.NoError(t, err)
	api.Config = pausedConfig
	_, _, err = api.SignState(SignStateRequest{
		State:		stateTwo,
		Signer:		alice,
		KeyRole:	SignerKeyRoleParticipant,
		Amount:		"10",
		Automated:	true,
		CurrentHeight:	26,
	})
	require.ErrorContains(t, err, "emergency paused")
}

func TestPaymentSignerAPISeparatesRoutingAndFundsKeys(t *testing.T) {
	fundsKey := testAddress(0x68)
	routeKey := testAddress(0x69)
	counterparty := testAddress(0x6a)
	channel := signedChannel(t, "signer-routing-separation", "1000", fundsKey, counterparty)
	state := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: fundsKey, Amount: "990"},
		{Participant: counterparty, Amount: "10"},
	})

	_, err := NewPaymentSignerAPI(PaymentSignerConfig{
		Signer:		routeKey,
		KeyRole:	SignerKeyRoleRoutingGossip,
		FundsKey:	fundsKey,
		GossipKey:	fundsKey,
	}, SignerPersistence{}, nil)
	require.ErrorContains(t, err, "separate from funds key")

	routingAPI, err := NewPaymentSignerAPI(PaymentSignerConfig{
		Signer:		routeKey,
		KeyRole:	SignerKeyRoleRoutingGossip,
		FundsKey:	fundsKey,
		GossipKey:	routeKey,
	}, SignerPersistence{}, nil)
	require.NoError(t, err)

	_, envelope, err := routingAPI.SignGossip(GossipMessage{
		MessageType:		GossipNodeAnnouncement,
		ChainID:		channel.ChainID,
		NodeID:			routeKey,
		From:			routeKey,
		ValidAfterHeight:	10,
		ValidUntilHeight:	50,
	}, routeKey, 20)
	require.NoError(t, err)
	require.Equal(t, routeKey, envelope.Signature.Signer)
	require.NoError(t, envelope.ValidateForState(EmptyState(), 20))

	_, _, err = routingAPI.SignState(SignStateRequest{
		State:		state,
		Signer:		routeKey,
		KeyRole:	SignerKeyRoleRoutingGossip,
		Amount:		"10",
		CurrentHeight:	20,
	})
	require.ErrorContains(t, err, "cannot sign channel state")

	fundsAPI, err := NewPaymentSignerAPI(PaymentSignerConfig{
		Signer:		fundsKey,
		KeyRole:	SignerKeyRoleParticipant,
		FundsKey:	fundsKey,
		GossipKey:	routeKey,
	}, SignerPersistence{}, nil)
	require.NoError(t, err)
	_, _, err = fundsAPI.SignGossip(envelope.Message, fundsKey, 20)
	require.ErrorContains(t, err, "requires routing gossip")
}

func TestKeyCompromiseCloseStartsFraudCloseWithLatestState(t *testing.T) {
	alice := testAddress(0x6b)
	bob := testAddress(0x6c)
	channel := signedChannel(t, "key-compromise-close", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	latest := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})

	bad := KeyCompromiseCloseRequest{
		ChannelID:	channel.ChannelID,
		CompromisedKey:	testAddress(0x6d),
		SafeSubmitter:	bob,
		LatestState:	latest,
		CurrentHeight:	20,
		EvidenceHash:	HashParts("key-compromise", "bad", channel.ChannelID),
	}
	_, err = SubmitKeyCompromiseClose(state, bad)
	require.ErrorContains(t, err, "channel participants")

	next, err := SubmitKeyCompromiseClose(state, KeyCompromiseCloseRequest{
		ChannelID:	channel.ChannelID,
		CompromisedKey:	alice,
		SafeSubmitter:	bob,
		LatestState:	latest,
		CurrentHeight:	20,
		SettlementFee:	"0",
		EvidenceHash:	HashParts("key-compromise", channel.ChannelID, alice),
	})
	require.NoError(t, err)
	require.Equal(t, ChannelStatusPendingClose, next.Channels[0].Status)
	require.Equal(t, CloseReasonFraud, next.Channels[0].PendingClose.CloseReason)
	require.Equal(t, bob, next.Channels[0].PendingClose.Submitter)
	require.Equal(t, latest.StateHash, next.Channels[0].PendingClose.State.StateHash)
	require.Equal(t, uint64(20)+channel.DisputePeriod, next.Channels[0].PendingClose.SettleAfterHeight)
}

func TestRollbackVectorsRejectNonceAndPreviousHashRollback(t *testing.T) {
	alice := testAddress(0x67)
	bob := testAddress(0x68)
	channel := signedChannel(t, "rollback-vectors", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	update := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "490"},
		{Participant: bob, Amount: "510"},
	})
	state, err = AcceptSignedState(state, channel.ChannelID, update, 18)
	require.NoError(t, err)

	lowerNonce := signedState(t, channel, 1, "", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	})
	_, err = AcceptSignedState(state, channel.ChannelID, lowerNonce, 19)
	require.ErrorContains(t, err, "strictly increase")

	wrongPrevious := signedState(t, channel, 3, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "480"},
		{Participant: bob, Amount: "520"},
	})
	require.ErrorContains(t, ValidatePreviousHashContinuity(state.Channels[0], wrongPrevious), "previous hash")
	_, err = AcceptSignedState(state, channel.ChannelID, wrongPrevious, 20)
	require.ErrorContains(t, err, "previous hash")
}

func TestDoubleSignFraudAppliesIndependentPenalties(t *testing.T) {
	alice := testAddress(0x63)
	bob := testAddress(0x64)
	channel := signedChannel(t, "both-double-sign", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "600"},
		{Participant: bob, Amount: "400"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	aliceProof := FraudProof{
		ProofID:		HashParts("double-sign-alice", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		PenaltyAmount:		"25",
		EvidenceHash:		HashParts("evidence", "alice", closeState.StateHash, conflicting.StateHash),
	}
	bobProof := aliceProof
	bobProof.ProofID = HashParts("double-sign-bob", channel.ChannelID)
	bobProof.SubmittedBy = alice
	bobProof.OffendingSigner = bob
	bobProof.EvidenceHash = HashParts("evidence", "bob", closeState.StateHash, conflicting.StateHash)

	state, err = SubmitFraudProof(state, channel.ChannelID, aliceProof, 21)
	require.NoError(t, err)
	state, err = SubmitFraudProof(state, channel.ChannelID, bobProof, 22)
	require.NoError(t, err)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 2)
	require.Equal(t, alice, state.Channels[0].PendingClose.Penalties[0].Offender)
	require.Equal(t, bob, state.Channels[0].PendingClose.Penalties[1].Offender)
}

func TestBidirectionalCloseAndUpdateRules(t *testing.T) {
	alice := testAddress(0x39)
	bob := testAddress(0x3a)
	channel := signedChannel(t, "bidirectional-lifecycle", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	update := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "475"},
		{Participant: bob, Amount: "525"},
	})
	oneSignature := update
	oneSignature.Signatures = oneSignature.Signatures[:1]
	require.ErrorContains(t, oneSignature.ValidateForChannel(channel, true), "quorum")

	state, err = AcceptSignedState(state, channel.ChannelID, update, 18)
	require.NoError(t, err)
	_, err = AcceptSignedState(state, channel.ChannelID, update, 19)
	require.ErrorContains(t, err, "strictly increase")

	staleClose := signedState(t, channel, 1, "", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	})
	_, err = SubmitClose(state, channel.ChannelID, staleClose, alice, 20, "0")
	require.ErrorContains(t, err, "latest accepted nonce")

	cooperative := signedState(t, channel, 3, update.StateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	state, settlement, err := CooperativeClose(state, channel.ChannelID, cooperative, bob, 21, "5")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Empty(t, state.Channels[0].PendingClose.State.StateHash)
	require.Equal(t, cooperative.Nonce, settlement.Nonce)
	require.Equal(t, "545", amountFor(settlement.FinalBalances, bob))
}

func TestChannelUpdateLifecycleValidatesOffchainAndRegistersCheckpoint(t *testing.T) {
	alice := testAddress(0x46)
	bob := testAddress(0x47)
	channel := signedChannel(t, "update-lifecycle", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	update := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: channel.Participants[0], Amount: "475"},
		{Participant: channel.Participants[1], Amount: "500"},
	})
	req := ChannelUpdateRequest{
		ChannelID:		channel.ChannelID,
		State:			update,
		ConditionCommitments:	update.Conditions,
		Submitter:		alice,
		CurrentHeight:		18,
	}
	result, err := ValidateOffchainUpdate(channel, req)
	require.NoError(t, err)
	require.True(t, result.ValidatedOffChain)
	require.False(t, result.CheckpointRegistered)

	unchanged, result, err := RegisterUpdateCheckpoint(state, req)
	require.NoError(t, err)
	require.False(t, result.CheckpointRegistered)
	require.Equal(t, channel.LatestState.StateHash, unchanged.Channels[0].LatestState.StateHash)

	req.RegisterCheckpoint = true
	state, result, err = RegisterUpdateCheckpoint(state, req)
	require.NoError(t, err)
	require.True(t, result.CheckpointRegistered)
	require.Equal(t, update.StateHash, state.Channels[0].LatestState.StateHash)

	overReserve := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "30", []Balance{
		{Participant: channel.Participants[0], Amount: "475"},
		{Participant: channel.Participants[1], Amount: "500"},
	})
	_, err = ValidateOffchainUpdate(channel, ChannelUpdateRequest{
		ChannelID:		channel.ChannelID,
		State:			overReserve,
		ConditionCommitments:	overReserve.Conditions,
		Submitter:		alice,
		CurrentHeight:		18,
	})
	require.ErrorContains(t, err, "reserve")
}

func TestAsyncUpdateBatchCanRegisterCheckpoint(t *testing.T) {
	alice := testAddress(0x48)
	bob := testAddress(0x49)
	channel := signedAsyncChannel(t, "async-update-lifecycle", "1000", []Balance{
		{Participant: alice, Amount: "700"},
		{Participant: bob, Amount: "300"},
	}, 4, 8, "100", 90, alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	delta := signedAsyncDelta(t, channel, "update-delta", alice, bob, "40", 2, 2, 80)
	checkpoint, err := BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{delta}, 3, 30)
	require.NoError(t, err)
	checkpoint = signAsyncCheckpoint(t, channel, checkpoint)
	state, result, err := RegisterUpdateCheckpoint(state, ChannelUpdateRequest{
		ChannelID:		channel.ChannelID,
		State:			checkpoint,
		AsyncDeltas:		[]AsyncPaymentDelta{delta},
		RegisterCheckpoint:	true,
		Submitter:		bob,
		CurrentHeight:		30,
	})
	require.NoError(t, err)
	require.True(t, result.CheckpointRegistered)
	require.Equal(t, checkpoint.StateHash, state.Channels[0].LatestState.StateHash)
}

func TestUnidirectionalReceiverCloseUsesSinglePayerSignature(t *testing.T) {
	payer := testAddress(0x3b)
	receiver := testAddress(0x3c)
	channel := signedUnidirectionalChannel(t, "uni-close", "1000", payer, receiver, false)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	claim := signedUnidirectionalClaim(t, channel, "320", 2, 80, false)
	require.Empty(t, claim.ReceiverAckOptional.SignatureHash)
	require.NoError(t, claim.ValidateForChannel(channel))

	badClaim := claim
	badClaim.PayerSignature, err = SignatureForClaim(badClaim, receiver)
	require.NoError(t, err)
	require.ErrorContains(t, badClaim.ValidateForChannel(channel), "payer signature")

	state, settlement, err := ReceiverClose(state, channel.ChannelID, claim, receiver, 30, "5")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, claim.Nonce, state.Channels[0].FinalizedNonce)
	require.Equal(t, claim.StateHash, settlement.StateHash)
	require.Equal(t, "680", amountFor(settlement.FinalBalances, payer))
	require.Equal(t, "315", amountFor(settlement.FinalBalances, receiver))
}

func TestUnidirectionalAcknowledgementModeAndPayerReclaim(t *testing.T) {
	payer := testAddress(0x3d)
	receiver := testAddress(0x3e)
	channel := signedUnidirectionalChannel(t, "uni-ack", "1000", payer, receiver, true)

	claim, err := BuildUnidirectionalClaim(UnidirectionalClaim{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Payer:			payer,
		Receiver:		receiver,
		LockedAmount:		channel.Collateral,
		ClaimedAmount:		"125",
		Nonce:			2,
		ExpirationHeight:	80,
		ExpirationTimestamp:	0,
	})
	require.NoError(t, err)
	claim.PayerSignature, err = SignatureForClaim(claim, payer)
	require.NoError(t, err)
	require.ErrorContains(t, claim.ValidateForChannel(channel), "acknowledgement")
	claim.ReceiverAckOptional, err = SignatureForClaim(claim, receiver)
	require.NoError(t, err)
	require.NoError(t, claim.ValidateForChannel(channel))

	state := EmptyState()
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	_, _, err = PayerReclaim(state, channel.ChannelID, payer, channel.ExpirationHeight+channel.DisputePeriod, "3")
	require.ErrorContains(t, err, "dispute window")
	state, settlement, err := PayerReclaim(state, channel.ChannelID, payer, channel.ExpirationHeight+channel.DisputePeriod+1, "3")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, "997", amountFor(settlement.FinalBalances, payer))
	require.Equal(t, "0", amountFor(settlement.FinalBalances, receiver))
}

func TestUnidirectionalStreamingPaymentHelperFormat(t *testing.T) {
	payer := testAddress(0x3f)
	receiver := testAddress(0x40)
	channel := signedUnidirectionalChannel(t, "uni-stream", "1000", payer, receiver, false)

	claim, err := StreamingClaimForChannel(channel, StreamingPaymentFrame{
		ChannelID:		channel.ChannelID,
		StreamID:		HashParts("stream", channel.ChannelID),
		Payer:			payer,
		Receiver:		receiver,
		PreviousClaimed:	"10",
		RatePerBlock:		"5",
		StartHeight:		20,
		CurrentHeight:		32,
		Nonce:			2,
		ExpirationHeight:	90,
		ExpirationTimestamp:	0,
	})
	require.NoError(t, err)
	require.Equal(t, "70", claim.ClaimedAmount)
	claim.PayerSignature, err = SignatureForClaim(claim, payer)
	require.NoError(t, err)
	require.NoError(t, claim.ValidateForChannel(channel))
}

func TestAsyncCheckpointAggregationExposureExpiryAndProof(t *testing.T) {
	alice := testAddress(0x44)
	bob := testAddress(0x45)
	channel := signedAsyncChannel(t, "async-main", "1000", []Balance{
		{Participant: alice, Amount: "700"},
		{Participant: bob, Amount: "300"},
	}, 4, 8, "100", 90, alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	delta := signedAsyncDelta(t, channel, "delta-1", alice, bob, "40", 2, 2, 80)
	checkpoint, err := BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{delta}, 3, 30)
	require.NoError(t, err)
	checkpoint = signAsyncCheckpoint(t, channel, checkpoint)
	proof := AsyncDeltaDisputeProof{
		ProofID:		HashParts("proof", checkpoint.StateHash),
		ChannelID:		channel.ChannelID,
		CheckpointState:	checkpoint,
		Deltas:			[]AsyncPaymentDelta{delta},
		EvidenceHash:		HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRootForChannel(channel, []AsyncPaymentDelta{delta})),
	}
	require.NoError(t, proof.ValidateForChannel(channel, 30))

	state, err = AcceptAsyncCheckpoint(state, channel.ChannelID, checkpoint, []AsyncPaymentDelta{delta}, bob, 30)
	require.NoError(t, err)
	require.Equal(t, "660", amountFor(state.Channels[0].LatestState.Balances, alice))
	require.Equal(t, "340", amountFor(state.Channels[0].LatestState.Balances, bob))
	require.Equal(t, checkpoint.StateHash, state.Channels[0].LatestState.StateHash)

	badSigner := delta
	badSigner.Signature, err = SignatureForAsyncDelta(badSigner, bob)
	require.NoError(t, err)
	require.ErrorContains(t, badSigner.ValidateForChannel(channel, 30), "sender")

	tooMuch := []AsyncPaymentDelta{
		signedAsyncDelta(t, channel, "delta-2", alice, bob, "60", 3, 3, 80),
		signedAsyncDelta(t, channel, "delta-3", alice, bob, "50", 4, 4, 80),
	}
	_, err = BuildAsyncCheckpointState(channel, tooMuch, 5, 30)
	require.ErrorContains(t, err, "exposure")

	expired := signedAsyncDelta(t, channel, "delta-expired", bob, alice, "10", 3, 3, 31)
	_, err = BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{expired}, 4, 32)
	require.ErrorContains(t, err, "expired")

	badProof := proof
	badProof.Deltas = nil
	badProof.EvidenceHash = HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRootForChannel(channel, nil))
	require.ErrorContains(t, badProof.ValidateForChannel(channel, 30), "signed deltas")
}

func TestAsyncCheckpointRejectsDuplicateDeltaNonce(t *testing.T) {
	alice := testAddress(0x57)
	bob := testAddress(0x58)
	channel := signedAsyncChannel(t, "async-duplicate-nonce", "1000", []Balance{
		{Participant: alice, Amount: "1000"},
		{Participant: bob, Amount: "0"},
	}, 4, 4, "100", 80, alice, bob)

	first := signedAsyncDelta(t, channel, "first", alice, bob, "10", 2, 2, 70)
	second := signedAsyncDelta(t, channel, "second", alice, bob, "15", 2, 2, 70)
	_, err := BuildAsyncCheckpointState(channel, []AsyncPaymentDelta{first, second}, 3, 30)
	require.ErrorContains(t, err, "duplicate async delta nonce")
}

func TestPaymentAssetScopeRejectsNonNaetFeesAndPenalties(t *testing.T) {
	alice := testAddress(0x35)
	bob := testAddress(0x36)
	channel := signedChannel(t, "asset-scope", "1000", alice, bob)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})

	edge := ChannelEdge{
		ChannelID:	channel.ChannelID,
		From:		alice,
		To:		bob,
		Capacity:	"100",
		FeeDenom:	"uatom",
		FeeAmount:	"1",
		Active:		true,
	}
	require.ErrorContains(t, edge.Validate(), "naet")

	pending := PendingClose{
		Submitter:		alice,
		SubmittedHeight:	20,
		SettleAfterHeight:	28,
		SettlementFeeDenom:	"uatom",
		SettlementFee:		"1",
		State:			closeState,
	}
	require.ErrorContains(t, pending.ValidateForChannel(channel), "naet")

	proof := FraudProof{
		ProofID:		HashParts("bad-penalty-denom"),
		ProofType:		FraudProofTypeStaleClose,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB: signedState(t, channel, 3, closeState.StateHash, []Balance{
			{Participant: alice, Amount: "450"},
			{Participant: bob, Amount: "550"},
		}),
		PenaltyDenom:	"uatom",
		PenaltyAmount:	"10",
		EvidenceHash:	HashParts("evidence", "bad-penalty-denom"),
	}
	require.ErrorContains(t, proof.ValidateForChannel(channel), "naet")

	penalty := Penalty{Offender: alice, Recipient: bob, Denom: "uatom", Amount: "10"}
	require.ErrorContains(t, penalty.ValidateForChannel(channel), "naet")

	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		closeState.StateHash,
		Nonce:			closeState.Nonce,
		FinalBalances:		closeState.Balances,
		SettlementFeeDenom:	"uatom",
		SettlementFee:		"0",
		SettledHeight:		40,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	require.ErrorContains(t, settlement.ValidateForChannel(channel), "naet")
}

func TestRoutePaymentAndVirtualChannelUseExistingLiquidity(t *testing.T) {
	alice := testAddress(0x41)
	router := testAddress(0x42)
	bob := testAddress(0x43)
	first := signedChannel(t, "route-1", "700", alice, router)
	second := signedChannel(t, "route-2", "700", bob, router)

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	firstReserve := signedReserveState(t, first, 2, first.OpeningStateHash, "250", "0", []Balance{
		{Participant: first.Participants[0], Amount: "450"},
		{Participant: first.Participants[1], Amount: "0"},
	})
	secondReserve := signedReserveState(t, second, 2, second.OpeningStateHash, "250", "0", []Balance{
		{Participant: second.Participants[0], Amount: "450"},
		{Participant: second.Participants[1], Amount: "0"},
	})
	state, err = AcceptSignedState(state, first.ChannelID, firstReserve, 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, second.ChannelID, secondReserve, 20)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)

	path, err := RoutePayment(state, alice, bob, "250", 10, 4)
	require.NoError(t, err)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{path[0].ChannelID, path[1].ChannelID})

	vc := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("virtual", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"250",
		BalanceA:		"250",
		BalanceB:		"0",
		RoutingFeeAmount:	"2",
		ExpiresHeight:		40,
	}, first.ChainID)
	state, err = OpenVirtualChannel(state, vc)
	require.NoError(t, err)
	require.Len(t, state.VirtualChannels, 1)
	require.Equal(t, "250", state.VirtualChannels[0].BalanceA)

	state, closed, err := CloseVirtualChannel(state, vc.VirtualChannelID, 45)
	require.NoError(t, err)
	require.Equal(t, VirtualChannelStatusSettled, closed.Status)
	require.Empty(t, state.VirtualChannels)
}

func TestVirtualChannelActivationRequiresReservesExpiryAndSignatures(t *testing.T) {
	alice := testAddress(0x6d)
	router := testAddress(0x6e)
	bob := testAddress(0x6f)
	first := signedChannel(t, "vc-rules-first", "1000", alice, router)
	second := signedChannel(t, "vc-rules-second", "1000", router, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)

	unsigned := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-rules", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		40,
	}, first.ChainID)
	_, err = OpenVirtualChannel(state, unsigned)
	require.ErrorContains(t, err, "reserved capacity")

	firstReserve := signedReserveState(t, first, 2, first.OpeningStateHash, "100", "0", []Balance{
		{Participant: first.Participants[0], Amount: "900"},
		{Participant: first.Participants[1], Amount: "0"},
	})
	secondReserve := signedReserveState(t, second, 2, second.OpeningStateHash, "100", "0", []Balance{
		{Participant: second.Participants[0], Amount: "900"},
		{Participant: second.Participants[1], Amount: "0"},
	})
	state, err = AcceptSignedState(state, first.ChannelID, firstReserve, 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, second.ChannelID, secondReserve, 20)
	require.NoError(t, err)

	tooLate := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-too-late", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		firstReserve.TimeoutHeight,
	}, first.ChainID)
	_, err = OpenVirtualChannel(state, tooLate)
	require.ErrorContains(t, err, "parent safety timeout")

	missingSignature := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-missing-sig", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		40,
	}, first.ChainID)
	missingSignature.Signatures = missingSignature.Signatures[:1]
	_, err = OpenVirtualChannel(state, missingSignature)
	require.ErrorContains(t, err, "missing required signature")

	_, err = BuildVirtualChannel(VirtualChannel{
		VirtualChannelID:	HashParts("vc-bad-balances", alice, bob),
		ChainID:		first.ChainID,
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"60",
		BalanceB:		"60",
		ExpiresHeight:		40,
	})
	require.ErrorContains(t, err, "balances")
}

func TestVirtualChannelOpeningRequiresReservationProofAndRouteTimeout(t *testing.T) {
	alice := testAddress(0x70)
	router := testAddress(0x71)
	bob := testAddress(0x72)
	first := signedChannel(t, "vc-proof-first", "900", alice, router)
	second := signedChannel(t, "vc-proof-second", "900", router, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, first.ChannelID, signedReserveState(t, first, 2, first.OpeningStateHash, "150", "0", []Balance{
		{Participant: first.Participants[0], Amount: "750"},
		{Participant: first.Participants[1], Amount: "0"},
	}), 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, second.ChannelID, signedReserveState(t, second, 2, second.OpeningStateHash, "150", "0", []Balance{
		{Participant: second.Participants[0], Amount: "750"},
		{Participant: second.Participants[1], Amount: "0"},
	}), 20)
	require.NoError(t, err)

	vc := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-proof", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		RoutingFeeAmount:	"2",
		ExpiresHeight:		40,
	}, first.ChainID)
	proof := signedVirtualActivationProof(t, vc, router, 80)
	state, err = OpenVirtualChannelWithProof(state, proof)
	require.NoError(t, err)
	require.Len(t, state.VirtualChannels, 1)

	badSignature := proof
	badSignature.VirtualChannel.VirtualChannelID = HashParts("vc-proof-bad-signature", alice, bob)
	badSignature.ProofHash = ""
	badSignature.VirtualChannel = signedVirtualChannel(t, badSignature.VirtualChannel, first.ChainID)
	_, err = OpenVirtualChannelWithProof(state.Clone(), badSignature)
	require.ErrorContains(t, err, "signature channel id mismatch")

	tooLate := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-proof-too-late", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		70,
	}, first.ChainID)
	_, err = OpenVirtualChannelWithProof(state.Clone(), signedVirtualActivationProof(t, tooLate, router, 80))
	require.ErrorContains(t, err, "route timeout")
}

func TestVirtualChannelEndpointUpdatesAndDisputeProof(t *testing.T) {
	alice := testAddress(0x73)
	router := testAddress(0x74)
	bob := testAddress(0x75)
	first := signedChannel(t, "vc-update-first", "900", alice, router)
	second := signedChannel(t, "vc-update-second", "900", router, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	firstReserve := signedReserveState(t, first, 2, first.OpeningStateHash, "120", "0", []Balance{
		{Participant: first.Participants[0], Amount: "780"},
		{Participant: first.Participants[1], Amount: "0"},
	})
	secondReserve := signedReserveState(t, second, 2, second.OpeningStateHash, "120", "0", []Balance{
		{Participant: second.Participants[0], Amount: "780"},
		{Participant: second.Participants[1], Amount: "0"},
	})
	state, err = AcceptSignedState(state, first.ChannelID, firstReserve, 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, second.ChannelID, secondReserve, 20)
	require.NoError(t, err)

	vc := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-update", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		40,
	}, first.ChainID)
	proof := signedVirtualActivationProof(t, vc, router, 80)
	state, err = OpenVirtualChannelWithProof(state, proof)
	require.NoError(t, err)
	parentHashes := []string{state.Channels[0].LatestState.StateHash, state.Channels[1].LatestState.StateHash}

	update := signedVirtualEndpointUpdate(t, vc, 2, "70", "30")
	state, err = AcceptVirtualChannelUpdate(state, update, 25)
	require.NoError(t, err)
	require.Equal(t, uint64(2), state.VirtualChannels[0].Nonce)
	require.Equal(t, "30", state.VirtualChannels[0].BalanceB)
	require.Equal(t, parentHashes, []string{state.Channels[0].LatestState.StateHash, state.Channels[1].LatestState.StateHash})

	_, err = AcceptVirtualChannelUpdate(state, update, 26)
	require.ErrorContains(t, err, "nonce must strictly increase")

	disputeState := signedVirtualEndpointUpdate(t, vc, 3, "60", "40")
	commitments := virtualReserveCommitments(proof)
	dispute, err := BuildVirtualChannelDisputeProof(disputeState, commitments, router)
	require.NoError(t, err)
	state, err = SubmitVirtualChannelDispute(state, dispute, 30)
	require.NoError(t, err)
	require.Equal(t, uint64(3), state.VirtualChannels[0].Nonce)
	require.Equal(t, "40", state.VirtualChannels[0].BalanceB)

	badDispute, err := BuildVirtualChannelDisputeProof(signedVirtualEndpointUpdate(t, vc, 4, "50", "50"), []string{HashParts("wrong-reserve"), commitments[1]}, router)
	require.NoError(t, err)
	_, err = SubmitVirtualChannelDispute(state, badDispute, 31)
	require.ErrorContains(t, err, "reserve commitment mismatch")
}

func TestVirtualChannelCloseProofModesAndTimeoutHierarchy(t *testing.T) {
	alice := testAddress(0x76)
	router := testAddress(0x77)
	bob := testAddress(0x78)
	state, vc, proof := virtualChannelFixture(t, "vc-close", alice, router, bob, "100", 40)

	final := signedVirtualEndpointUpdate(t, vc, 2, "80", "20")
	closeProof, err := BuildVirtualCloseProof(final, VirtualCloseModeCooperative, virtualReserveCommitments(proof), alice, 30)
	require.NoError(t, err)
	next, closed, releases, err := CloseVirtualChannelWithProof(state, closeProof, 30)
	require.NoError(t, err)
	require.Empty(t, next.VirtualChannels)
	require.Equal(t, VirtualChannelStatusSettled, closed.Status)
	require.Len(t, releases, 2)
	require.Equal(t, uint64(30), releases[0].ReleaseHeight)

	state, vc, proof = virtualChannelFixture(t, "vc-expired", alice, router, bob, "100", 40)
	expiredState := signedVirtualEndpointUpdate(t, vc, 1, "100", "0")
	expiredProof, err := BuildVirtualCloseProof(expiredState, VirtualCloseModeExpired, virtualReserveCommitments(proof), bob, 55)
	require.NoError(t, err)
	_, _, _, err = CloseVirtualChannelWithProof(state, expiredProof, 55)
	require.ErrorContains(t, err, "before finalization")
	expiredProof, err = BuildVirtualCloseProof(expiredState, VirtualCloseModeExpired, virtualReserveCommitments(proof), bob, 56)
	require.NoError(t, err)
	_, _, releases, err = CloseVirtualChannelWithProof(state, expiredProof, 56)
	require.NoError(t, err)
	require.Equal(t, uint64(56), releases[0].ReleaseHeight)

	state, vc, proof = virtualChannelFixture(t, "vc-risk", alice, router, bob, "100", 40)
	parentHashes := []string{state.Channels[0].LatestState.StateHash, state.Channels[1].LatestState.StateHash}
	riskState := signedVirtualEndpointUpdate(t, vc, 1, "100", "0")
	riskProof, err := BuildVirtualCloseProof(riskState, VirtualCloseModeIntermediaryRisk, virtualReserveCommitments(proof), router, 25)
	require.NoError(t, err)
	next, _, releases, err = CloseVirtualChannelWithProof(state, riskProof, 25)
	require.NoError(t, err)
	require.Equal(t, uint64(25+DefaultDisputePeriod), releases[0].ReleaseHeight)
	require.Equal(t, parentHashes, []string{next.Channels[0].LatestState.StateHash, next.Channels[1].LatestState.StateHash})
}

func TestVirtualChannelNestedDisputeSimulation(t *testing.T) {
	alice := testAddress(0x79)
	router := testAddress(0x7a)
	bob := testAddress(0x7b)
	state, vc, activation := virtualChannelFixture(t, "vc-nested", alice, router, bob, "100", 40)

	update := signedVirtualEndpointUpdate(t, vc, 2, "70", "30")
	var err error
	state, err = AcceptVirtualChannelUpdate(state, update, 24)
	require.NoError(t, err)

	staleClose, err := BuildVirtualCloseProof(vc, VirtualCloseModeDisputed, virtualReserveCommitments(activation), router, 30)
	require.NoError(t, err)
	_, _, _, err = CloseVirtualChannelWithProof(state, staleClose, 30)
	require.ErrorContains(t, err, "stale")

	newer := signedVirtualEndpointUpdate(t, vc, 3, "65", "35")
	dispute, err := BuildVirtualChannelDisputeProof(newer, virtualReserveCommitments(activation), router)
	require.NoError(t, err)
	state, err = SubmitVirtualChannelDispute(state, dispute, 30)
	require.NoError(t, err)
	require.Equal(t, uint64(3), state.VirtualChannels[0].Nonce)

	disputedClose, err := BuildVirtualCloseProof(newer, VirtualCloseModeDisputed, virtualReserveCommitments(activation), router, 31)
	require.NoError(t, err)
	state, closed, releases, err := CloseVirtualChannelWithProof(state, disputedClose, 31)
	require.NoError(t, err)
	require.Empty(t, state.VirtualChannels)
	require.Equal(t, "35", closed.BalanceB)
	require.Equal(t, uint64(31+DefaultDisputePeriod), releases[0].ReleaseHeight)
}

func TestParentChannelDisputeWhileVirtualChannelIsActive(t *testing.T) {
	alice := testAddress(0x7c)
	router := testAddress(0x7d)
	bob := testAddress(0x7e)
	state, vc, activation := virtualChannelFixture(t, "vc-parent-dispute", alice, router, bob, "100", 40)
	parent := state.Channels[0].Normalize()
	closeState := parent.LatestState.Normalize()
	parentStateHash := parent.LatestState.StateHash
	virtualStateHash := state.VirtualChannels[0].StateHash

	var err error
	state, err = SubmitClose(state, parent.ChannelID, closeState, parent.Participants[0], 25, "0")
	require.NoError(t, err)
	require.Len(t, state.VirtualChannels, 1)
	require.Equal(t, vc.VirtualChannelID, state.VirtualChannels[0].VirtualChannelID)
	require.Equal(t, virtualStateHash, state.VirtualChannels[0].StateHash)
	require.Contains(t, state.VirtualChannels[0].ParentReserveCommitments, activation.ParentReserves[0].ReserveCommitment)

	newerParent := signedReserveState(t, parent, closeState.Nonce+1, closeState.StateHash, "100", "0", []Balance{
		{Participant: parent.Participants[0], Amount: "790"},
		{Participant: parent.Participants[1], Amount: "10"},
	})
	state, err = DisputeClose(state, parent.ChannelID, newerParent, parent.Participants[1], 26)
	require.NoError(t, err)
	disputedParent, found := state.ChannelByID(parent.ChannelID)
	require.True(t, found)
	require.Equal(t, ChannelFinalityInDispute, disputedParent.Finality)
	require.Equal(t, newerParent.StateHash, disputedParent.PendingClose.State.StateHash)
	require.NotEqual(t, parentStateHash, disputedParent.PendingClose.State.StateHash)
	require.Len(t, state.VirtualChannels, 1)
	require.Equal(t, vc.VirtualChannelID, state.VirtualChannels[0].VirtualChannelID)
	require.Equal(t, virtualStateHash, state.VirtualChannels[0].StateHash)
}

func TestVirtualChannelLiquidityAggregationSegmentsAndPartialFailure(t *testing.T) {
	alice := testAddress(0x7c)
	bob := testAddress(0x7d)
	first := signedChannel(t, "vc-agg-first", "100", alice, bob)
	second := signedChannel(t, "vc-agg-second", "100", alice, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, first.ChannelID, signedReserveState(t, first, 2, first.OpeningStateHash, "60", "0", []Balance{
		{Participant: first.Participants[0], Amount: "40"},
		{Participant: first.Participants[1], Amount: "0"},
	}), 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, second.ChannelID, signedReserveState(t, second, 2, second.OpeningStateHash, "40", "0", []Balance{
		{Participant: second.Participants[0], Amount: "60"},
		{Participant: second.Participants[1], Amount: "0"},
	}), 20)
	require.NoError(t, err)

	vc := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts("vc-aggregated", alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Capacity:		"100",
		BalanceA:		"100",
		BalanceB:		"0",
		ExpiresHeight:		40,
	}, first.ChainID)
	reserves := []VirtualParentReserve{
		signedVirtualReserve(t, vc, first.ChannelID, alice, "60", "60"),
		signedVirtualReserve(t, vc, second.ChannelID, bob, "40", "40"),
	}
	activation, err := BuildVirtualActivationProof(vc, reserves, 80)
	require.NoError(t, err)
	activation.AggregatedCapacity = true
	activation.ProofHash = ComputeVirtualActivationProofHash(activation)
	require.NoError(t, ValidateVirtualActivationProof(activation))
	segments := VirtualReserveSegmentsFromProof(activation)
	require.NoError(t, ValidateVirtualReserveSegments(vc, segments))
	settlementProofs, err := BuildVirtualSegmentSettlementProofs(vc, segments)
	require.NoError(t, err)
	require.Len(t, settlementProofs, 2)
	for i := range settlementProofs {
		require.NoError(t, settlementProofs[i].ValidateForSegment(segments[i], vc))
	}

	state, err = OpenVirtualChannelWithProof(state, activation)
	require.NoError(t, err)
	require.Len(t, state.VirtualChannels[0].ParentReserveCommitments, 2)

	missing := activation
	missing.VirtualChannel.VirtualChannelID = HashParts("vc-aggregated-missing", alice, bob)
	missing.VirtualChannel = signedVirtualChannel(t, missing.VirtualChannel, first.ChainID)
	missing.ParentReserves = missing.ParentReserves[:1]
	missing.ProofHash = ComputeVirtualActivationProofHash(missing)
	_, err = OpenVirtualChannelWithProof(state.Clone(), missing)
	require.ErrorContains(t, err, "parent split")
	failure, err := BuildVirtualPartialActivationFailure(missing.VirtualChannel, missing.ParentReserves[0].SegmentID, "second segment unavailable", []string{missing.ParentReserves[0].ReserveCommitment})
	require.NoError(t, err)
	require.NoError(t, failure.ValidateForVirtualChannel(missing.VirtualChannel))
}

func TestScoredRouteSelectionPenalizesFeeStaleLiquidityAndFailures(t *testing.T) {
	alice := testAddress(0x37)
	router := testAddress(0x38)
	bob := testAddress(0x39)
	direct := signedChannel(t, "score-direct", "1000", alice, bob)
	first := signedChannel(t, "score-first", "1000", alice, router)
	second := signedChannel(t, "score-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{direct, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "20", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.HopPenalty = "1"
	policy.StaleLiquidityAfter = 10
	policy.EdgeStats = []EdgeRoutingStats{
		{ChannelID: direct.ChannelID, From: alice, To: bob, SuccessRateBps: 3_000, LiquidityUpdatedHeight: 1, FailureCount: 4, CongestionBps: 7_500, NodeAvailabilityBps: 5_000, TimeoutMargin: 4},
		{ChannelID: first.ChannelID, From: alice, To: router, SuccessRateBps: 10_000, LiquidityUpdatedHeight: 95, NodeAvailabilityBps: 10_000, TimeoutMargin: DefaultTimeoutMargin},
		{ChannelID: second.ChannelID, From: router, To: bob, SuccessRateBps: 10_000, LiquidityUpdatedHeight: 95, NodeAvailabilityBps: 10_000, TimeoutMargin: DefaultTimeoutMargin},
	}
	route, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{
		From:		alice,
		To:		bob,
		Amount:		"100",
		CurrentHeight:	100,
		Policy:		policy,
	})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{route.Edges[0].ChannelID, route.Edges[1].ChannelID})
	require.Equal(t, "2", route.TotalFee)
	require.NotEmpty(t, route.ScoreHash)

	again, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 100, Policy: policy})
	require.NoError(t, err)
	require.Equal(t, route.ScoreHash, again.ScoreHash)
	sim, err := SimulateRoute(route, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 100, Policy: policy})
	require.NoError(t, err)
	require.True(t, sim.Attemptable)
}

func TestScoredRouteSelectionExcludesInsufficientCapacityAndMaxFee(t *testing.T) {
	alice := testAddress(0x3a)
	router := testAddress(0x3b)
	bob := testAddress(0x3c)
	low := signedChannel(t, "score-low-capacity", "1000", alice, bob)
	first := signedChannel(t, "score-cap-first", "1000", alice, router)
	second := signedChannel(t, "score-cap-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{low, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: low.ChannelID, From: alice, To: bob, Capacity: "50", FeeAmount: "0", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "150", FeeAmount: "4", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "150", FeeAmount: "4", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.MaxFeeAmount = "10"
	route, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, Policy: policy})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, "8", route.TotalFee)

	policy.MaxFeeAmount = "3"
	_, err = SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, Policy: policy})
	require.ErrorContains(t, err, "not found")
}

func TestRoutingFeePolicyUpdateAndHopFeeCalculation(t *testing.T) {
	alice := testAddress(0x41)
	router := testAddress(0x42)
	policy, err := BuildRoutingFeePolicyUpdate(RoutingFeePolicyUpdate{
		ChainID:			"aetra-test-chain",
		ChannelID:			HashParts("routing-fee-policy-channel"),
		From:				router,
		To:				alice,
		FeeDenom:			NativeDenom,
		BaseHopFee:			"2",
		ProportionalFeeBps:		250,
		LiquidityReservationFee:	"3",
		VirtualChannelSetupFee:		"5",
		CongestionSurcharge:		"7",
		FailurePenalty:			"11",
		MaxHopFee:			"100",
		ValidAfterHeight:		10,
		ValidUntilHeight:		20,
		Sequence:			1,
	}, router)
	require.NoError(t, err)
	require.NotEmpty(t, policy.PolicyID)
	require.NotEmpty(t, policy.PolicyHash)
	require.NoError(t, policy.ValidateAtHeight(12))
	require.ErrorContains(t, policy.ValidateAtHeight(21), "validity window")

	fee, err := CalculateHopRoutingFee(HopFeeCalculationRequest{
		Amount:				"1000",
		Policy:				policy,
		CurrentHeight:			12,
		IncludeVirtualSetup:		true,
		RepeatedInvalidAttempts:	2,
	})
	require.NoError(t, err)
	require.Equal(t, "2", fee.BaseHopFee)
	require.Equal(t, "25", fee.ProportionalFee)
	require.Equal(t, "3", fee.LiquidityReservationFee)
	require.Equal(t, "5", fee.VirtualChannelSetupFee)
	require.Equal(t, "7", fee.CongestionSurcharge)
	require.Equal(t, "22", fee.FailurePenalty)
	require.Equal(t, "64", fee.TotalFee)
	require.Equal(t, policy.PolicyHash, fee.PolicyHash)

	tooLowMax, err := BuildRoutingFeePolicyUpdate(RoutingFeePolicyUpdate{
		ChainID:			policy.ChainID,
		ChannelID:			policy.ChannelID,
		From:				policy.From,
		To:				policy.To,
		FeeDenom:			NativeDenom,
		BaseHopFee:			"2",
		ProportionalFeeBps:		250,
		LiquidityReservationFee:	"3",
		VirtualChannelSetupFee:		"5",
		CongestionSurcharge:		"7",
		FailurePenalty:			"11",
		MaxHopFee:			"50",
		ValidAfterHeight:		10,
		ValidUntilHeight:		20,
		Sequence:			2,
	}, router)
	require.NoError(t, err)
	_, err = CalculateHopRoutingFee(HopFeeCalculationRequest{
		Amount:				"1000",
		Policy:				tooLowMax,
		CurrentHeight:			12,
		IncludeVirtualSetup:		true,
		RepeatedInvalidAttempts:	2,
	})
	require.ErrorContains(t, err, "exceeds policy maximum")
}

func TestRouteFeeCeilingRejectsMicroFeeOvercharge(t *testing.T) {
	alice := testAddress(0x43)
	router := testAddress(0x44)
	bob := testAddress(0x45)
	policy := DefaultRoutePolicy()
	policy.MaxFeeAmount = "5"
	route := ScoredRoute{
		Edges: []ChannelEdge{{
			ChannelID:	HashParts("fee-ceiling-edge"),
			From:		alice,
			To:		bob,
			Capacity:	"100",
			FeeDenom:	NativeDenom,
			FeeAmount:	"4",
			Active:		true,
		}},
		Amount:		"50",
		TotalFee:	"4",
		TotalCost:	"4",
		MinCapacity:	"100",
	}
	require.NoError(t, ValidateRouteFeeCeiling(route, policy))
	route.TotalFee = "6"
	require.ErrorContains(t, ValidateRouteFeeCeiling(route, policy), "policy ceiling")

	first := signedChannel(t, "fee-ceiling-first", "1000", alice, router)
	second := signedChannel(t, "fee-ceiling-second", "1000", router, bob)
	first.ConditionalPayments = true
	second.ConditionalPayments = true
	routeID := HashParts("fee-ceiling-route")
	hashLock := HashParts("fee-ceiling-preimage")
	secondID := HashParts("fee-ceiling-second-promise")
	firstID := HashParts("fee-ceiling-first-promise")
	firstPromise := signedRoutePromise(t, first, firstID, routeID, alice, router, "58", "0", 9, 70, hashLock, "", secondID)
	secondPromise := signedRoutePromise(t, second, secondID, routeID, router, bob, "50", "8", 10, 40, hashLock, firstID, "")
	proof := ConditionLinkageProof{
		RouteID:	routeID,
		Sender:		alice,
		Receiver:	bob,
		Amount:		"50",
		TotalFees:	"8",
		HashLock:	hashLock,
		Promises:	[]ConditionalPromise{firstPromise, secondPromise},
		TimeoutMargin:	DefaultTimeoutMargin,
	}
	require.ErrorContains(t, ValidateConditionLinkageFeeCeiling(proof, policy), "policy ceiling")
	policy.MaxFeeAmount = "8"
	require.NoError(t, ValidateConditionLinkageFeeCeiling(proof, policy))

	underreported := proof
	underreported.TotalFees = "4"
	require.ErrorContains(t, ValidateConditionLinkageFeeCeiling(underreported, policy), "overcharge")
}

func TestMultiPathSplittingUsesIndependentCapacityAwareRoutes(t *testing.T) {
	alice := testAddress(0x3d)
	r1 := testAddress(0x3e)
	r2 := testAddress(0x3f)
	bob := testAddress(0x40)
	channels := []ChannelRecord{
		signedChannel(t, "split-a", "1000", alice, r1),
		signedChannel(t, "split-b", "1000", r1, bob),
		signedChannel(t, "split-c", "1000", alice, r2),
		signedChannel(t, "split-d", "1000", r2, bob),
	}
	state := EmptyState()
	var err error
	for _, channel := range channels {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	edges := []ChannelEdge{
		{ChannelID: channels[0].ChannelID, From: alice, To: r1, Capacity: "60", FeeAmount: "1", Active: true},
		{ChannelID: channels[1].ChannelID, From: r1, To: bob, Capacity: "60", FeeAmount: "1", Active: true},
		{ChannelID: channels[2].ChannelID, From: alice, To: r2, Capacity: "60", FeeAmount: "2", Active: true},
		{ChannelID: channels[3].ChannelID, From: r2, To: bob, Capacity: "60", FeeAmount: "2", Active: true},
	}
	for _, edge := range edges {
		state, err = RegisterRoutingEdge(state, edge)
		require.NoError(t, err)
	}
	policy := DefaultRoutePolicy()
	policy.EnableMultiPath = true
	policy.MaxSplits = 2
	result, err := SplitPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 30, Policy: policy})
	require.NoError(t, err)
	require.Len(t, result.Parts, 2)
	require.Equal(t, "100", result.TotalAmount)
	require.Equal(t, "6", result.TotalFee)
	require.NotEqual(t, result.Parts[0].Edges[0].ChannelID, result.Parts[1].Edges[0].ChannelID)
	require.NotEmpty(t, result.ScoreHash)
}

func TestCongestionSignalsIncreaseWeightAndReduceMaxPaymentSize(t *testing.T) {
	alice := testAddress(0x57)
	router := testAddress(0x58)
	bob := testAddress(0x59)
	direct := signedChannel(t, "congestion-direct", "1000", alice, bob)
	first := signedChannel(t, "congestion-first", "1000", alice, router)
	second := signedChannel(t, "congestion-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{direct, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy, err = ApplyCongestionSnapshot(policy, CongestionSnapshot{
		ChannelID:			direct.ChannelID,
		From:				alice,
		To:				bob,
		ChannelUpdateFailureRateBps:	9_000,
		PendingConditionCount:		8,
		AvgResolutionLatency:		500,
		RouteRetryCount:		3,
		ReservePressureBps:		8_000,
		NodeQueueDelay:			400,
		LiquidityUpdatedHeight:		10,
		ObservedHeight:			100,
	})
	require.NoError(t, err)

	route, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 100, Policy: policy})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{route.Edges[0].ChannelID, route.Edges[1].ChannelID})

	cappedPolicy := policy
	cappedPolicy.ExcludedChannels = []string{first.ChannelID, second.ChannelID}
	_, err = SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "300", CurrentHeight: 100, Policy: cappedPolicy})
	require.ErrorContains(t, err, "eligible")
}

func TestCongestionPenaltyDecayRestoresRoutePreference(t *testing.T) {
	alice := testAddress(0x5a)
	bob := testAddress(0x5b)
	cheap := signedChannel(t, "decay-cheap", "1000", alice, bob)
	expensive := signedChannel(t, "decay-expensive", "1000", alice, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, cheap)
	require.NoError(t, err)
	state, err = OpenChannel(state, expensive)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: cheap.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: expensive.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "9", Active: true})
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.DecayHalfLife = 10
	policy, err = ApplyCongestionSnapshot(policy, CongestionSnapshot{
		ChannelID:			cheap.ChannelID,
		From:				alice,
		To:				bob,
		ChannelUpdateFailureRateBps:	9_000,
		PendingConditionCount:		10,
		ReservePressureBps:		1_000,
		ObservedHeight:			10,
	})
	require.NoError(t, err)
	congested, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 10, Policy: policy})
	require.NoError(t, err)
	require.Equal(t, expensive.ChannelID, congested.Edges[0].ChannelID)

	decayed := DecayRoutePolicyPenalties(policy, 120)
	recovered, err := SelectPaymentRoute(state, TopologyStore{}, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 120, Policy: decayed})
	require.NoError(t, err)
	require.Equal(t, cheap.ChannelID, recovered.Edges[0].ChannelID)
}

func TestCongestionAwareRetryPolicySelectsAlternateRoute(t *testing.T) {
	alice := testAddress(0x5c)
	router := testAddress(0x5d)
	bob := testAddress(0x5e)
	direct := signedChannel(t, "retry-direct", "1000", alice, bob)
	first := signedChannel(t, "retry-first", "1000", alice, router)
	second := signedChannel(t, "retry-second", "1000", router, bob)
	state := EmptyState()
	var err error
	for _, channel := range []ChannelRecord{direct, first, second} {
		state, err = OpenChannel(state, channel)
		require.NoError(t, err)
	}
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: first.ChannelID, From: alice, To: router, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: second.ChannelID, From: router, To: bob, Capacity: "500", FeeAmount: "2", Active: true})
	require.NoError(t, err)

	result, err := RetryPaymentRoute(state, TopologyStore{}, RouteRetryRequest{
		Selection:	RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 50, Policy: DefaultRoutePolicy()},
		Failures: []RouteFailureReport{{
			ChannelID:	direct.ChannelID,
			From:		alice,
			To:		bob,
			FailureClass:	ClassifyRouteFailure("node queue congestion"),
			Retryable:	true,
			ObservedHeight:	50,
		}},
		Policy:	RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true},
	})
	require.NoError(t, err)
	require.True(t, result.Retryable)
	require.Equal(t, uint32(2), result.Attempts)
	require.Len(t, result.Route.Edges, 2)
	require.Equal(t, []string{first.ChannelID, second.ChannelID}, []string{result.Route.Edges[0].ChannelID, result.Route.Edges[1].ChannelID})
	require.NotEmpty(t, result.PolicyHash)

	exhausted, err := RetryPaymentRoute(state, TopologyStore{}, RouteRetryRequest{
		Selection:	RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 50, Policy: DefaultRoutePolicy()},
		Failures: []RouteFailureReport{
			{ChannelID: direct.ChannelID, From: alice, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 50},
			{ChannelID: first.ChannelID, From: alice, To: router, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 51},
			{ChannelID: second.ChannelID, From: router, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 52},
		},
		Policy:	RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true},
	})
	require.NoError(t, err)
	require.False(t, exhausted.Retryable)
	require.Contains(t, exhausted.Reason, "attempts exhausted")
}

func TestRoutingEngineModuleGossipSearchRetryProbeAndSpamResistance(t *testing.T) {
	alice := testAddress(0xc1)
	router := testAddress(0xc2)
	bob := testAddress(0xc3)
	direct := signedChannel(t, "routing-engine-direct", "500", alice, bob)
	first := signedChannel(t, "routing-engine-first", "500", alice, router)
	second := signedChannel(t, "routing-engine-second", "500", router, bob)
	state := EmptyStateWithChannel(t, direct)
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)

	policy := DefaultRoutePolicy()
	policy.MaxHops = 4
	policy.HopPenalty = "0"
	engine, err := SnapshotRoutingEngineState(TopologyStore{}, policy, GossipRateLimitPolicy{
		WindowBlocks:		8,
		MaxMessagesPerNode:	8,
		MaxMessagesPerChannel:	4,
		MaxTopologyUpdates:	8,
		RejectPenalty:		InvalidGossipPenalty,
	}, DefaultRouteFailureScoringPolicy())
	require.NoError(t, err)

	node := routingEngineEnvelope(t, GossipMessage{
		MessageType:		GossipNodeAnnouncement,
		ChainID:		direct.ChainID,
		NodeID:			router,
		From:			router,
		ValidAfterHeight:	10,
		ValidUntilHeight:	50,
	}, router, 10)
	engine, decision, err := ApplyRoutingEngineMessage(engine, state, MsgGossipNodeAnnouncement{Gossip: node}, 10)
	require.NoError(t, err)
	require.True(t, decision.Allowed)
	require.Len(t, engine.Nodes, 1)

	for _, env := range []SignedGossipEnvelope{
		routingEngineEnvelope(t, GossipMessage{MessageType: GossipChannelUpdate, ChainID: direct.ChainID, ChannelID: direct.ChannelID, NodeID: alice, From: alice, To: bob, Capacity: "500", FeeDenom: NativeDenom, FeeAmount: "9", ValidAfterHeight: 10, ValidUntilHeight: 50, Sequence: 1}, alice, 11),
		routingEngineEnvelope(t, GossipMessage{MessageType: GossipChannelUpdate, ChainID: first.ChainID, ChannelID: first.ChannelID, NodeID: alice, From: alice, To: router, Capacity: "500", FeeDenom: NativeDenom, FeeAmount: "1", ValidAfterHeight: 10, ValidUntilHeight: 50, Sequence: 2}, alice, 11),
		routingEngineEnvelope(t, GossipMessage{MessageType: GossipChannelUpdate, ChainID: second.ChainID, ChannelID: second.ChannelID, NodeID: router, From: router, To: bob, Capacity: "500", FeeDenom: NativeDenom, FeeAmount: "1", ValidAfterHeight: 10, ValidUntilHeight: 50, Sequence: 3}, router, 11),
	} {
		engine, _, err = ApplyRoutingEngineMessage(engine, state, MsgGossipChannelUpdate{Gossip: env}, 11)
		require.NoError(t, err)
	}
	require.Len(t, engine.Topology.Edges, 3)

	liquidity := routingEngineEnvelope(t, GossipMessage{MessageType: GossipLiquidityHint, ChainID: first.ChainID, ChannelID: first.ChannelID, NodeID: alice, From: alice, To: router, Capacity: "500", Liquidity: "450", FeeDenom: NativeDenom, FeeAmount: "1", ValidAfterHeight: 12, ValidUntilHeight: 40, Sequence: 4, Advisory: true}, alice, 12)
	engine, _, err = ApplyRoutingEngineMessage(engine, state, MsgGossipLiquidityHint{Gossip: liquidity}, 12)
	require.NoError(t, err)
	require.Len(t, engine.LiquidityHints, 1)

	feePolicy := routingEngineEnvelope(t, GossipMessage{MessageType: GossipFeePolicyUpdate, ChainID: direct.ChainID, ChannelID: direct.ChannelID, NodeID: alice, From: alice, To: bob, Capacity: "500", FeeDenom: NativeDenom, FeeAmount: "9", MaxFee: "20", ValidAfterHeight: 12, ValidUntilHeight: 40, Sequence: 5}, alice, 12)
	engine, _, err = ApplyRoutingEngineMessage(engine, state, MsgGossipFeePolicyUpdate{Gossip: feePolicy}, 12)
	require.NoError(t, err)
	require.Len(t, engine.FeePolicies, 1)

	engine, route, err := SelectRoutingEnginePath(engine, state, RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, Policy: policy})
	require.NoError(t, err)
	require.Len(t, route.Edges, 2)
	require.Equal(t, "2", route.TotalFee)
	require.Len(t, engine.RouteAttempts, 1)

	probe, err := HandleCapacityProbe(engine, state, CapacityProbeRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 20, MaxHops: 4, BlindedRouteHint: HashParts("probe-hint")}, router)
	require.NoError(t, err)
	require.True(t, probe.Available)
	require.NotEmpty(t, probe.RouteHash)

	engine, retry, err := RetryRoutingEnginePath(engine, state, RouteRetryRequest{
		Selection:	RouteSelectionRequest{From: alice, To: bob, Amount: "100", CurrentHeight: 21, Policy: policy},
		Failures:	[]RouteFailureReport{{ChannelID: first.ChannelID, From: alice, To: router, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 20}},
		Policy:		RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(2), retry.Attempts)
	require.Len(t, retry.Route.Edges, 1)
	require.Equal(t, direct.ChannelID, retry.Route.Edges[0].ChannelID)
	require.NotEmpty(t, engine.RouteAttempts)

	engine, scores, err := ApplyRoutingEngineFailures(engine, []RouteFailureReport{{ChannelID: direct.ChannelID, From: alice, To: bob, FailureClass: RouteFailureLiquidityStale, Retryable: true, ObservedHeight: 22}})
	require.NoError(t, err)
	require.Len(t, scores, 1)
	require.NotEmpty(t, engine.LocalPeerScores)
	require.Negative(t, RoutingScoreForEdge(engine.Topology, ChannelEdge{ChannelID: direct.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "9", Active: true}))

	spamEngine, err := SnapshotRoutingEngineState(TopologyStore{}, policy, GossipRateLimitPolicy{WindowBlocks: 8, MaxMessagesPerNode: 2, MaxMessagesPerChannel: 8, MaxTopologyUpdates: 8, RejectPenalty: InvalidGossipPenalty}, DefaultRouteFailureScoringPolicy())
	require.NoError(t, err)
	for seq := uint64(1); seq <= 2; seq++ {
		env := routingEngineEnvelope(t, GossipMessage{MessageType: GossipNodeAnnouncement, ChainID: direct.ChainID, NodeID: router, From: router, ValidAfterHeight: 30, ValidUntilHeight: 60, Sequence: seq}, router, 30+seq)
		spamEngine, _, err = ApplyRoutingEngineMessage(spamEngine, state, MsgGossipNodeAnnouncement{Gossip: env}, 30+seq)
		require.NoError(t, err)
	}
	spam := routingEngineEnvelope(t, GossipMessage{MessageType: GossipNodeAnnouncement, ChainID: direct.ChainID, NodeID: router, From: router, ValidAfterHeight: 30, ValidUntilHeight: 60, Sequence: 3}, router, 32)
	_, decision, err = ApplyRoutingEngineMessage(spamEngine, state, MsgGossipNodeAnnouncement{Gossip: spam}, 32)
	require.ErrorContains(t, err, "rate limit")
	require.False(t, decision.Allowed)
}

func TestForwardingPacketsExposeOnlyPerHopMetadata(t *testing.T) {
	alice := testAddress(0x67)
	router1 := testAddress(0x68)
	router2 := testAddress(0x69)
	bob := testAddress(0x6a)
	route := ScoredRoute{
		Edges: []ChannelEdge{
			{ChannelID: HashParts("privacy-channel-1"), From: alice, To: router1, Capacity: "500", FeeAmount: "1", Active: true},
			{ChannelID: HashParts("privacy-channel-2"), From: router1, To: router2, Capacity: "500", FeeAmount: "2", Active: true},
			{ChannelID: HashParts("privacy-channel-3"), From: router2, To: bob, Capacity: "500", FeeAmount: "3", Active: true},
		},
		Amount:		"100",
		TotalFee:	"6",
		TotalCost:	"9",
		MinCapacity:	"500",
		ScoreHash:	HashParts("privacy-score"),
	}
	packets, err := BuildForwardingPackets(route, "payment-seed", 7, 100)
	require.NoError(t, err)
	require.Len(t, packets, 3)
	require.Equal(t, alice, packets[0].ForwardingNode)
	require.Equal(t, router1, packets[0].NextNode)
	require.Equal(t, router1, packets[1].ForwardingNode)
	require.Equal(t, router2, packets[1].NextNode)
	require.NotEqual(t, packets[0].RouteID, packets[1].RouteID)
	require.NotEqual(t, packets[1].RouteID, packets[2].RouteID)
	require.NotEqual(t, packets[0].HopPaymentID, packets[1].HopPaymentID)
	require.Equal(t, packets[1].PacketHash, packets[0].NextPacketHash)
	require.Equal(t, packets[2].PacketHash, packets[1].NextPacketHash)
	require.Empty(t, packets[2].NextPacketHash)

	logRecord, err := PrivacySafeForwardingLog(packets[1], 50)
	require.NoError(t, err)
	require.Equal(t, packets[1].PacketID, logRecord.PacketID)
	require.NotEqual(t, packets[1].NextNode, logRecord.NextNodeHash)
	require.NotEqual(t, packets[1].Amount, logRecord.AmountHash)
	require.Equal(t, router1, logRecord.ForwardingNode)
}

func TestForwardingPacketReplayProtectionRejectsReusedIdentifiers(t *testing.T) {
	alice := testAddress(0x6b)
	bob := testAddress(0x6c)
	route := ScoredRoute{
		Edges:		[]ChannelEdge{{ChannelID: HashParts("privacy-replay-channel"), From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true}},
		Amount:		"50",
		TotalFee:	"1",
		TotalCost:	"2",
		MinCapacity:	"500",
		ScoreHash:	HashParts("privacy-replay-score"),
	}
	_, err := DeriveRouteID("seed", 0)
	require.ErrorContains(t, err, "nonce")
	packets, err := BuildForwardingPackets(route, "seed", 1, 80)
	require.NoError(t, err)
	var records []ForwardingPacketReplayRecord
	records, err = RecordForwardingPacket(records, packets[0], 40)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.ErrorContains(t, ValidateForwardingPacket(packets[0], alice, records, 41), "replay")

	reusedRoute := packets[0]
	reusedRoute.HopPaymentID = HashParts("new-hop-payment")
	reusedRoute.NextPacketHash = ""
	reusedRoute.PacketHash = ComputeForwardingPacketHash(reusedRoute)
	reusedRoute.PacketID = HashParts("forwarding-packet-id", reusedRoute.PacketHash)
	require.ErrorContains(t, ValidateForwardingPacket(reusedRoute, alice, records, 41), "route id replay")

	pruned := PruneForwardingReplayRecords(records, 40+DefaultReplayHorizon+1)
	require.Empty(t, pruned)
	require.NoError(t, ValidateForwardingPacket(packets[0], alice, pruned, 41))
}

func TestUntrustedTopologyIsRejectedBeforeRouteUse(t *testing.T) {
	alice := testAddress(0x49)
	router := testAddress(0x4a)
	bob := testAddress(0x4b)
	channel := signedChannel(t, "verified-route", "500", alice, router)

	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	_, err = RegisterRoutingEdge(state, ChannelEdge{
		ChannelID:	HashParts("unknown-channel"),
		From:		alice,
		To:		bob,
		Capacity:	"100",
		FeeAmount:	"1",
		Active:		true,
	})
	require.ErrorContains(t, err, "open channel")

	_, err = RegisterRoutingEdge(state, ChannelEdge{
		ChannelID:	channel.ChannelID,
		From:		alice,
		To:		bob,
		Capacity:	"100",
		FeeAmount:	"1",
		Active:		true,
	})
	require.ErrorContains(t, err, "participants")

	untrusted := state
	untrusted.Edges = append(untrusted.Edges, ChannelEdge{
		ChannelID:	HashParts("gossip-only"),
		From:		alice,
		To:		bob,
		Capacity:	"100",
		FeeAmount:	"1",
		Active:		true,
	})
	_, err = RoutePayment(untrusted, alice, bob, "50", 10, 4)
	require.ErrorContains(t, err, "unknown channel")
}

func TestSignedGossipEnvelopeBuildsLocalTopologyStore(t *testing.T) {
	alice := testAddress(0x31)
	bob := testAddress(0x32)
	channel := signedChannel(t, "gossip-announcement", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	envelope := signedGossipEnvelope(t, GossipMessage{
		MessageType:		GossipChannelAnnouncement,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Capacity:		"500",
		FeeAmount:		"2",
		ValidAfterHeight:	20,
		ValidUntilHeight:	50,
		ReputationDelta:	3,
		Sequence:		1,
	}, alice, 20)

	store, err := ApplyGossipEnvelope(TopologyStore{}, state, envelope, 20)
	require.NoError(t, err)
	require.Len(t, store.Messages, 1)
	require.Len(t, store.Edges, 1)
	require.Equal(t, channel.ChannelID, store.Edges[0].ChannelID)
	require.Equal(t, int64(3), RoutingScoreForEdge(store, store.Edges[0]))

	commitmentOnly := signedGossipEnvelope(t, GossipMessage{
		MessageType:		GossipChannelAnnouncement,
		ChainID:		channel.ChainID,
		NodeID:			bob,
		From:			bob,
		To:			alice,
		Capacity:		"100",
		FeeAmount:		"1",
		ValidAfterHeight:	20,
		ValidUntilHeight:	50,
		ChannelCommitment:	HashParts("verifiable-channel-commitment", bob, alice),
		Sequence:		2,
	}, bob, 20)
	store, err = ApplyGossipEnvelope(store, state, commitmentOnly, 20)
	require.NoError(t, err)
	require.Len(t, store.Messages, 2)
	require.Len(t, store.Edges, 1)
}

func TestGossipExpiryPruningAndInvalidPenaltyAffectLocalScoreOnly(t *testing.T) {
	alice := testAddress(0x33)
	bob := testAddress(0x34)
	channel := signedChannel(t, "gossip-prune", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	envelope := signedGossipEnvelope(t, GossipMessage{
		MessageType:		GossipChannelUpdate,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Capacity:		"400",
		FeeAmount:		"1",
		ValidAfterHeight:	20,
		ValidUntilHeight:	25,
		Sequence:		1,
	}, alice, 20)

	store, err := ApplyGossipEnvelope(TopologyStore{}, state, envelope, 20)
	require.NoError(t, err)
	require.Len(t, store.Edges, 1)
	pruned, err := PruneTopologyStore(store, 26)
	require.NoError(t, err)
	require.Empty(t, pruned.Messages)
	require.Empty(t, pruned.Edges)

	_, err = ApplyGossipEnvelope(TopologyStore{}, state, envelope, 26)
	require.ErrorContains(t, err, "expired")

	invalid := signedGossipEnvelope(t, GossipMessage{
		MessageType:		GossipLiquidityHint,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Liquidity:		"250",
		FeeAmount:		"1",
		ValidAfterHeight:	30,
		ValidUntilHeight:	60,
		Sequence:		2,
		Advisory:		true,
	}, bob, 30)
	penalized, err := ApplyGossipEnvelope(store, state, invalid, 30)
	require.ErrorContains(t, err, "signer must match")
	require.Len(t, penalized.Reputation, 1)
	require.Equal(t, uint64(1), penalized.Reputation[0].InvalidGossip)
	require.Equal(t, -InvalidGossipPenalty, RoutingScoreForEdge(penalized, store.Edges[0]))
	require.Len(t, state.Channels, 1)
	require.Empty(t, state.Edges)
}

func TestFeePolicyGossipRequiresValidityAndMaxFee(t *testing.T) {
	alice := testAddress(0x35)
	bob := testAddress(0x36)
	channel := signedChannel(t, "gossip-fee-policy", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	invalidPolicy := GossipMessage{
		MessageType:		GossipFeePolicyUpdate,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		FeeAmount:		"1",
		ValidAfterHeight:	20,
		ValidUntilHeight:	50,
	}
	_, err := BuildGossipMessage(invalidPolicy)
	require.ErrorContains(t, err, "max fee")

	policy := signedGossipEnvelope(t, GossipMessage{
		MessageType:		GossipFeePolicyUpdate,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Capacity:		"300",
		FeeAmount:		"2",
		MaxFee:			"5",
		ValidAfterHeight:	20,
		ValidUntilHeight:	50,
		Sequence:		1,
	}, alice, 20)
	store, err := ApplyGossipEnvelope(TopologyStore{}, state, policy, 20)
	require.NoError(t, err)
	require.Len(t, store.Edges, 1)
	require.Equal(t, "2", store.Edges[0].FeeAmount)
}

func TestGossipRateLimitRejectsTopologySpamLocally(t *testing.T) {
	alice := testAddress(0xb0)
	bob := testAddress(0xb1)
	channel := signedChannel(t, "gossip-rate-limit", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	policy := GossipRateLimitPolicy{WindowBlocks: 4, MaxMessagesPerNode: 2, MaxMessagesPerChannel: 4, MaxTopologyUpdates: 8, RejectPenalty: InvalidGossipPenalty}
	store := TopologyStore{}

	for seq := uint64(1); seq <= 2; seq++ {
		envelope := signedGossipEnvelope(t, GossipMessage{
			MessageType:		GossipChannelUpdate,
			ChainID:		channel.ChainID,
			ChannelID:		channel.ChannelID,
			NodeID:			alice,
			From:			alice,
			To:			bob,
			Capacity:		"500",
			FeeAmount:		"1",
			ValidAfterHeight:	20,
			ValidUntilHeight:	40,
			Sequence:		seq,
		}, alice, 20)
		var decision GossipRateLimitDecision
		var err error
		store, decision, err = ApplyGossipEnvelopeWithRateLimit(store, state, envelope, 20, policy)
		require.NoError(t, err)
		require.True(t, decision.Allowed)
	}
	require.Len(t, store.Messages, 2)

	spam := signedGossipEnvelope(t, GossipMessage{
		MessageType:		GossipChannelUpdate,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		NodeID:			alice,
		From:			alice,
		To:			bob,
		Capacity:		"500",
		FeeAmount:		"1",
		ValidAfterHeight:	20,
		ValidUntilHeight:	40,
		Sequence:		3,
	}, alice, 20)
	next, decision, err := ApplyGossipEnvelopeWithRateLimit(store, state, spam, 20, policy)
	require.ErrorContains(t, err, "node message rate limit")
	require.False(t, decision.Allowed)
	require.Equal(t, uint32(2), decision.NodeMessages)
	require.Len(t, next.Messages, 2)
	require.Len(t, next.Reputation, 1)
	require.Equal(t, -InvalidGossipPenalty, next.Reputation[0].Score)
	require.Empty(t, state.Edges)
}

func TestRouteFailureScoringReducesLocalRoutingScore(t *testing.T) {
	alice := testAddress(0xb2)
	bob := testAddress(0xb3)
	channel := signedChannel(t, "route-failure-score", "1000", alice, bob)
	edge := ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "500", FeeAmount: "1", Active: true}
	reports := []RouteFailureReport{
		{ChannelID: channel.ChannelID, From: alice, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 30},
		{ChannelID: channel.ChannelID, From: alice, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 31},
		{ChannelID: channel.ChannelID, From: alice, To: bob, FailureClass: RouteFailureCongestion, Retryable: true, ObservedHeight: 32},
	}

	store, scores, err := ApplyRouteFailureScoring(TopologyStore{}, reports, DefaultRouteFailureScoringPolicy())
	require.NoError(t, err)
	require.Len(t, scores, 3)
	require.Equal(t, int64(-30), scores[0].ScoreDelta)
	require.Equal(t, int64(-40), scores[1].ScoreDelta)
	require.Equal(t, int64(-50), scores[2].ScoreDelta)
	require.Equal(t, int64(-120), RoutingScoreForEdge(store, edge))
	require.NotEmpty(t, scores[2].ScoreHash)
}

func TestHighValueRouteRequiresBackedLiquidityProof(t *testing.T) {
	alice := testAddress(0xb4)
	bob := testAddress(0xb5)
	channel := signedChannel(t, "liquidity-proof", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	ad, err := BuildLiquidityAdvertisement(LiquidityAdvertisement{
		ChannelID:		channel.ChannelID,
		Advertiser:		alice,
		Counterparty:		bob,
		Capacity:		"300",
		FeeDenom:		NativeDenom,
		BaseFee:		"1",
		ReservationFee:		"2",
		ReliabilityBps:		9_000,
		ValidUntilHeight:	60,
		DepositAmount:		"10",
		BackedByReservation:	true,
	}, "10")
	require.NoError(t, err)
	reservation, err := BuildSignedLiquidityReservation(SignedLiquidityReservation{
		AdvertisementID:	ad.AdvertisementID,
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Reserver:		alice,
		Counterparty:		bob,
		Capacity:		"250",
		FeeAmount:		"2",
		ExpirationHeight:	55,
		Nonce:			1,
	}, alice)
	require.NoError(t, err)
	proof, err := BuildRouteLiquidityProof(RouteLiquidityProof{
		ChannelID:		channel.ChannelID,
		Amount:			"200",
		HighValueThreshold:	"100",
		RequiredDeposit:	"10",
		CurrentHeight:		40,
		Advertisement:		ad,
		Reservation:		reservation,
	})
	require.NoError(t, err)
	require.NoError(t, VerifyRouteLiquidityProof(state, proof))

	unbackedAd, err := BuildLiquidityAdvertisement(LiquidityAdvertisement{
		ChannelID:		channel.ChannelID,
		Advertiser:		alice,
		Counterparty:		bob,
		Capacity:		"300",
		FeeDenom:		NativeDenom,
		BaseFee:		"1",
		ReservationFee:		"2",
		ReliabilityBps:		9_000,
		ValidUntilHeight:	60,
		DepositAmount:		"10",
	}, "10")
	require.NoError(t, err)
	unbacked := proof
	unbacked.Advertisement = unbackedAd
	unbacked.ProofHash = ComputeRouteLiquidityProofHash(unbacked)
	require.ErrorContains(t, VerifyRouteLiquidityProof(state, unbacked), "requires backed advertisement")

	lowValue := proof
	lowValue.Amount = "50"
	lowValue.ProofHash = ComputeRouteLiquidityProofHash(lowValue)
	require.NoError(t, VerifyRouteLiquidityProof(state, lowValue))
}

func TestTopologySpamSimulationAppliesRateLimitsAndPenalties(t *testing.T) {
	alice := testAddress(0xb6)
	bob := testAddress(0xb7)
	channel := signedChannel(t, "topology-spam", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	policy := GossipRateLimitPolicy{WindowBlocks: 4, MaxMessagesPerNode: 2, MaxMessagesPerChannel: 8, MaxTopologyUpdates: 8, RejectPenalty: InvalidGossipPenalty}
	envelopes := []SignedGossipEnvelope{}
	for seq := uint64(1); seq <= 4; seq++ {
		envelopes = append(envelopes, signedGossipEnvelope(t, GossipMessage{
			MessageType:		GossipChannelUpdate,
			ChainID:		channel.ChainID,
			ChannelID:		channel.ChannelID,
			NodeID:			alice,
			From:			alice,
			To:			bob,
			Capacity:		"500",
			FeeAmount:		"1",
			ValidAfterHeight:	20,
			ValidUntilHeight:	40,
			Sequence:		seq,
		}, alice, 20))
	}

	store, sim, err := SimulateTopologySpam(state, TopologyStore{}, envelopes, 20, policy)
	require.NoError(t, err)
	require.Equal(t, uint32(2), sim.Accepted)
	require.Equal(t, uint32(2), sim.Rejected)
	require.Equal(t, []string{alice}, sim.PenalizedNodes)
	require.Len(t, sim.Decisions, 4)
	require.NotEmpty(t, sim.SimulationHash)
	require.Len(t, store.Messages, 2)
	require.Equal(t, int64(-2*InvalidGossipPenalty), store.Reputation[0].Score)
}

func TestSettlementPrunesRoutingEdgesForAuthoritativeClosedChannel(t *testing.T) {
	alice := testAddress(0x4c)
	bob := testAddress(0x4d)
	channel := signedChannel(t, "settlement-prunes-edge", "500", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "100", FeeAmount: "1", Active: true})
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "250"},
		{Participant: bob, Amount: "250"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	state, _, err = FinalizeSettlement(state, channel.ChannelID, 40)
	require.NoError(t, err)
	require.Empty(t, state.Edges)

	_, err = RoutePayment(state, alice, bob, "50", 41, 4)
	require.ErrorContains(t, err, "route not found")
}

func TestFinalSettlementRequiresResolvedConditionsAndUnlocksCustody(t *testing.T) {
	alice := testAddress(0x55)
	bob := testAddress(0x56)
	channel := signedChannel(t, "condition-settlement", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: channel.Participants[0], Amount: "475"},
		{Participant: channel.Participants[1], Amount: "500"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	_, _, err = FinalizeSettlement(state, channel.ChannelID, 40)
	require.ErrorContains(t, err, "conditions")

	resolution := ConditionResolution{
		ConditionID:	closeState.Conditions[0].ConditionID,
		Resolver:	bob,
		Recipient:	closeState.Conditions[0].Payee,
		Amount:		closeState.Conditions[0].Amount,
		EvidenceHash:	HashParts("condition-resolution", closeState.Conditions[0].ConditionID),
	}
	state, settlement, err := FinalizeSettlementWithRequest(state, FinalSettlementRequest{
		ChannelID:		channel.ChannelID,
		ResolvedConditions:	[]ConditionResolution{resolution},
		CurrentHeight:		40,
	})
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Empty(t, state.CustodyLocks)
	require.Equal(t, "475", amountFor(settlement.FinalBalances, channel.Participants[0]))
	require.Equal(t, "525", amountFor(settlement.FinalBalances, channel.Participants[1]))
	require.Len(t, state.ClosedChannels, 1)
	require.Equal(t, channel.ChannelID, state.ClosedChannels[0].ChannelID)
	require.Len(t, state.ConditionClaims, 1)
	require.Equal(t, resolution.ConditionID, state.ConditionClaims[0].ConditionID)
}

func TestSettlementRejectsReusedConditionAndPreimageClaims(t *testing.T) {
	alice := testAddress(0x65)
	bob := testAddress(0x66)
	channel := signedChannel(t, "condition-replay", "1000", alice, bob)
	closeState := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: alice, Amount: "975"},
		{Participant: bob, Amount: "0"},
	})
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	resolution := ConditionResolution{
		ConditionID:	closeState.Conditions[0].ConditionID,
		Resolver:	alice,
		Recipient:	bob,
		Amount:		"25",
		EvidenceHash:	HashParts("condition-preimage", "shared"),
	}
	reusedCondition := state
	reusedCondition.ConditionClaims = append(reusedCondition.ConditionClaims, ConditionClaimRecord{
		ChainID:	channel.ChainID,
		ChannelID:	channel.ChannelID,
		ConditionID:	resolution.ConditionID,
		EvidenceHash:	HashParts("condition-preimage", "old"),
		ResolvedHeight:	19,
		ExpiresHeight:	19 + DefaultReplayHorizon,
	})
	_, _, err = FinalizeSettlementWithRequest(reusedCondition, FinalSettlementRequest{
		ChannelID:		channel.ChannelID,
		ResolvedConditions:	[]ConditionResolution{resolution},
		CurrentHeight:		40,
	})
	require.ErrorContains(t, err, "condition claim")

	reusedEvidence := state
	reusedEvidence.ConditionClaims = append(reusedEvidence.ConditionClaims, ConditionClaimRecord{
		ChainID:	channel.ChainID,
		ChannelID:	channel.ChannelID,
		ConditionID:	HashParts("other-condition"),
		EvidenceHash:	resolution.EvidenceHash,
		ResolvedHeight:	19,
		ExpiresHeight:	19 + DefaultReplayHorizon,
	})
	_, _, err = FinalizeSettlementWithRequest(reusedEvidence, FinalSettlementRequest{
		ChannelID:		channel.ChannelID,
		ResolvedConditions:	[]ConditionResolution{resolution},
		CurrentHeight:		40,
	})
	require.ErrorContains(t, err, "evidence claim")
}

func TestForcedClosePreservesDisputeWindowAfterTimeout(t *testing.T) {
	alice := testAddress(0x4e)
	bob := testAddress(0x4f)
	channel := signedChannel(t, "forced-close", "500", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	_, err = ForcedClose(state, channel.ChannelID, alice, channel.LatestState.TimeoutHeight, "0")
	require.ErrorContains(t, err, "timeout")

	state, err = ForcedClose(state, channel.ChannelID, alice, channel.LatestState.TimeoutHeight+1, "0")
	require.NoError(t, err)
	require.Equal(t, ChannelStatusPendingClose, state.Channels[0].Status)
	require.Equal(t, channel.LatestState.TimeoutHeight+1+channel.DisputePeriod, state.Channels[0].PendingClose.SettleAfterHeight)

	newer := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "225"},
		{Participant: bob, Amount: "275"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newer, bob, channel.LatestState.TimeoutHeight+2)
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, state.Channels[0].PendingClose.State.StateHash)
}

func TestWatchServiceSubmitsStaleCloseDispute(t *testing.T) {
	alice := testAddress(0x69)
	bob := testAddress(0x6a)
	watch := testAddress(0x6b)
	channel := signedChannel(t, "watch-stale-close", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	stale := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	newer := signedState(t, channel, 3, stale.StateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	state, err = SubmitClose(state, channel.ChannelID, stale, alice, 20, "0")
	require.NoError(t, err)

	state, err = SubmitWatchDispute(state, WatchDisputeSubmission{
		WatchService:		watch,
		Delegator:		bob,
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	stale.StateHash,
		NewerState:		newer,
		CurrentHeight:		21,
		EvidenceHash:		HashParts("watch-dispute", channel.ChannelID, newer.StateHash),
	})
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, state.Channels[0].PendingClose.State.StateHash)

	_, err = SubmitWatchDispute(state, WatchDisputeSubmission{
		WatchService:		watch,
		Delegator:		watch,
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	newer.StateHash,
		NewerState:		newer,
		CurrentHeight:		22,
	})
	require.ErrorContains(t, err, "delegator")
}

func TestValidatorAssistedWatchServiceSubmitsDispute(t *testing.T) {
	alice := testAddress(0x92)
	bob := testAddress(0x93)
	validator := testAddress(0x94)
	service := testAddress(0x95)
	channel := signedChannel(t, "validator-watch-dispute", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = RegisterValidatorPaymentService(state, ValidatorPaymentServiceMetadata{
		ValidatorAddress:	validator,
		ServiceAddress:		service,
		WatchEndpoint:		"https://validator.example/watch",
		RoutingEndpoint:	"https://validator.example/route",
		PublicKey:		"validator-watch-key",
		MinDelegation:		"100",
		CommissionBps:		250,
		Active:			true,
		UpdatedHeight:		10,
	})
	require.NoError(t, err)
	require.Len(t, state.ValidatorPaymentServices, 1)
	require.NotEmpty(t, state.ValidatorPaymentServices[0].MetadataHash)
	state, err = RegisterValidatorWatchService(state, ValidatorWatchRegistration{
		ValidatorAddress:	validator,
		Delegator:		bob,
		RegisteredHeight:	11,
	})
	require.NoError(t, err)

	stale := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	newer := signedState(t, channel, 3, stale.StateHash, []Balance{
		{Participant: alice, Amount: "425"},
		{Participant: bob, Amount: "575"},
	})
	state, err = SubmitClose(state, channel.ChannelID, stale, alice, 20, "0")
	require.NoError(t, err)
	state, err = SubmitValidatorAssistedDispute(state, ValidatorAssistedDisputeSubmission{
		ValidatorAddress:	validator,
		ServiceAddress:		service,
		Delegator:		bob,
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	stale.StateHash,
		NewerState:		newer,
		CurrentHeight:		21,
		EvidenceHash:		HashParts("validator-watch", channel.ChannelID, newer.StateHash),
	})
	require.NoError(t, err)
	require.Equal(t, newer.StateHash, state.Channels[0].PendingClose.State.StateHash)
	require.Contains(t, paymentEventTypes(state.Events), "validator-assisted-dispute")

	_, err = SubmitValidatorAssistedDispute(state, ValidatorAssistedDisputeSubmission{
		ValidatorAddress:	testAddress(0x96),
		ServiceAddress:		service,
		Delegator:		bob,
		ChannelID:		channel.ChannelID,
		ClosingStateReference:	newer.StateHash,
		NewerState:		newer,
		CurrentHeight:		22,
	})
	require.ErrorContains(t, err, "validator service not found")
}

func TestValidatorServicePenaltiesStaySeparateFromSlashing(t *testing.T) {
	alice := testAddress(0x97)
	bob := testAddress(0x98)
	validator := alice
	service := testAddress(0x99)
	channel := signedChannel(t, "validator-penalty-separation", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = RegisterValidatorPaymentService(state, ValidatorPaymentServiceMetadata{
		ValidatorAddress:	validator,
		ServiceAddress:		service,
		WatchEndpoint:		"https://validator.example/watch",
		MinDelegation:		"1",
		Active:			true,
		UpdatedHeight:		10,
	})
	require.NoError(t, err)
	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	proof := FraudProof{
		ProofID:		HashParts("validator-channel-fraud", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	validator,
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		ComputeDisputeProofHash(FraudProof{ProofID: HashParts("validator-channel-fraud", channel.ChannelID), ProofType: FraudProofTypeDoubleSign, StateA: closeState, StateB: conflicting}),
		PenaltyDenom:		NativeDenom,
		PenaltyAmount:		"30",
	}
	state, err = SubmitFraudProofWithPolicy(state, channel.ChannelID, proof, 21, FraudPenaltyPolicy{
		ReporterRewardCap:		"15",
		SecurityReserveShareBps:	7000,
		CommunityPoolShareBps:		3000,
	})
	require.NoError(t, err)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 1)
	require.Equal(t, validator, state.Channels[0].PendingClose.Penalties[0].Offender)
	require.Equal(t, "15", state.Channels[0].PendingClose.Penalties[0].Amount)
	require.Equal(t, "10", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteSecurityReserve))
	require.Equal(t, "5", allocationAmountFor(state.Channels[0].PendingClose.PenaltyAllocations, PenaltyRouteCommunityPool))
}

func TestFraudCloseSettlesAfterAcceptedProof(t *testing.T) {
	alice := testAddress(0x53)
	bob := testAddress(0x54)
	channel := signedChannel(t, "fraud-close", "1000", alice, bob)
	state := EmptyState()

	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "400"},
		{Participant: bob, Amount: "600"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)

	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	proof := FraudProof{
		ProofID:		HashParts("fraud-close-proof", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		PenaltyAmount:		"25",
		EvidenceHash:		HashParts("evidence", closeState.StateHash, conflicting.StateHash),
	}
	state, err = SubmitFraudProof(state, channel.ChannelID, proof, 21)
	require.NoError(t, err)
	state, settlement, err := FraudClose(state, channel.ChannelID, 22)
	require.NoError(t, err)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Equal(t, "375", amountFor(settlement.FinalBalances, alice))
	require.Equal(t, "625", amountFor(settlement.FinalBalances, bob))
}

func TestSettlementBatchRequiresIndependentChannels(t *testing.T) {
	alice := testAddress(0x51)
	bob := testAddress(0x52)
	first := signedChannel(t, "batch-1", "100", alice, bob)
	second := signedChannel(t, "batch-2", "100", alice, bob)
	ops := []SettlementOperation{
		{OperationID: HashParts("op", "second"), OperationType: BatchOperationSettle, ChannelID: second.ChannelID, Nonce: 1, StateHash: second.LatestState.StateHash},
		{OperationID: HashParts("op", "first"), OperationType: BatchOperationClose, ChannelID: first.ChannelID, Nonce: 1, StateHash: first.LatestState.StateHash},
	}

	batch, err := NewSettlementBatch(HashParts("batch"), ops)
	require.NoError(t, err)
	require.Less(t, batch.Operations[0].ChannelID, batch.Operations[1].ChannelID)

	batch.Operations = append(batch.Operations, SettlementOperation{
		OperationID:	HashParts("op", "duplicate"),
		OperationType:	BatchOperationDispute,
		ChannelID:	first.ChannelID,
		Nonce:		2,
		StateHash:	first.LatestState.StateHash,
	})
	batch.RootHash = ComputeBatchRoot(batch.Operations)
	require.ErrorContains(t, batch.Validate(), "independent")
}

func TestBlockSTMAccessPlanUsesPerChannelKeysAndDefersAccounting(t *testing.T) {
	alice := testAddress(0x81)
	bob := testAddress(0x82)
	first := signedChannel(t, "blockstm-first", "100", alice, bob)
	second := signedChannel(t, "blockstm-second", "100", alice, bob)
	firstOp := SettlementOperation{OperationID: HashParts("blockstm-op", "first"), OperationType: BatchOperationSettle, ChannelID: first.ChannelID, Nonce: 1, StateHash: first.LatestState.StateHash}
	secondOp := SettlementOperation{OperationID: HashParts("blockstm-op", "second"), OperationType: BatchOperationSettle, ChannelID: second.ChannelID, Nonce: 1, StateHash: second.LatestState.StateHash}

	firstPlan, err := AccessPlanForSettlementOperation(firstOp, 100)
	require.NoError(t, err)
	secondPlan, err := AccessPlanForSettlementOperation(secondOp, 100)
	require.NoError(t, err)
	require.Contains(t, firstPlan.WriteKeys, PaymentChannelKey(first.ChannelID))
	require.NotContains(t, firstPlan.WriteKeys, PaymentBlockAccumulatorKey(100))
	require.Equal(t, []string{PaymentBlockAccumulatorKey(100)}, firstPlan.AccumulatorKeys)

	profile := ProfileBlockSTMConflicts([]BlockSTMAccessPlan{firstPlan, secondPlan})
	require.True(t, profile.ConflictFree)
	require.True(t, profile.GlobalAccountingDeferred)
	require.Len(t, profile.ParallelizableGroups, 1)
	require.ElementsMatch(t, []string{firstOp.OperationID, secondOp.OperationID}, profile.ParallelizableGroups[0])
}

func TestBlockSTMConflictProfileDetectsSameChannelConflicts(t *testing.T) {
	alice := testAddress(0x83)
	bob := testAddress(0x84)
	channel := signedChannel(t, "blockstm-conflict", "100", alice, bob)
	closeOp := SettlementOperation{OperationID: HashParts("blockstm-conflict", "close"), OperationType: BatchOperationClose, ChannelID: channel.ChannelID, Nonce: 1, StateHash: channel.LatestState.StateHash}
	disputeOp := SettlementOperation{OperationID: HashParts("blockstm-conflict", "dispute"), OperationType: BatchOperationDispute, ChannelID: channel.ChannelID, Nonce: 2, StateHash: channel.LatestState.StateHash}
	closePlan, err := AccessPlanForSettlementOperation(closeOp, 101)
	require.NoError(t, err)
	disputePlan, err := AccessPlanForSettlementOperation(disputeOp, 101)
	require.NoError(t, err)

	profile := ProfileBlockSTMConflicts([]BlockSTMAccessPlan{closePlan, disputePlan})
	require.False(t, profile.ConflictFree)
	require.NotEmpty(t, profile.Conflicts)
	require.Contains(t, blockSTMConflictKeys(profile.Conflicts), PaymentChannelKey(channel.ChannelID))
	require.Len(t, profile.ParallelizableGroups, 2)
}

func TestSettlementBatchGroupingByChannelKey(t *testing.T) {
	alice := testAddress(0x85)
	bob := testAddress(0x86)
	first := signedChannel(t, "batch-group-first", "100", alice, bob)
	second := signedChannel(t, "batch-group-second", "100", alice, bob)
	ops := []SettlementOperation{
		{OperationID: HashParts("batch-group", "first-close"), OperationType: BatchOperationClose, ChannelID: first.ChannelID, Nonce: 1, StateHash: first.LatestState.StateHash},
		{OperationID: HashParts("batch-group", "first-settle"), OperationType: BatchOperationSettle, ChannelID: first.ChannelID, Nonce: 2, StateHash: first.LatestState.StateHash},
		{OperationID: HashParts("batch-group", "second-settle"), OperationType: BatchOperationSettle, ChannelID: second.ChannelID, Nonce: 1, StateHash: second.LatestState.StateHash},
	}

	groups, err := GroupSettlementOperationsByChannelKey("blockstm-group", ops)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	for _, group := range groups {
		require.NoError(t, group.Validate())
		seen := map[string]struct{}{}
		for _, op := range group.Operations {
			_, duplicate := seen[op.ChannelID]
			require.False(t, duplicate)
			seen[op.ChannelID] = struct{}{}
		}
	}
}

func TestPaymentBlockAccumulatorAggregatesAfterSettlementHotPath(t *testing.T) {
	alice := testAddress(0x87)
	bob := testAddress(0x88)
	channel := signedChannel(t, "block-accumulator", "100", alice, bob)
	settlement := SettlementRecord{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		StateHash:		channel.LatestState.StateHash,
		Nonce:			channel.LatestState.Nonce,
		FinalBalances:		channel.LatestState.Balances,
		SettlementFeeDenom:	NativeDenom,
		SettlementFee:		"3",
		PenaltyAllocations:	[]PenaltyAllocation{{Offender: alice, Route: PenaltyRouteCommunityPool, Denom: NativeDenom, Amount: "7"}},
		SettledHeight:		100,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	acc := PaymentBlockAccumulator{BlockHeight: 100}
	acc, err := AccumulatePaymentBlockAccounting(acc, settlement)
	require.NoError(t, err)
	require.NoError(t, acc.Validate())
	require.Equal(t, "3", acc.FeeAmount)
	require.Equal(t, "7", acc.PenaltyAmount)
	require.Equal(t, uint64(1), acc.OperationCount)
}

func TestPaymentChannelModuleMessagesDispatchAnteAndInvariants(t *testing.T) {
	alice := testAddress(0x8b)
	bob := testAddress(0x8c)
	openReq := ChannelOpenRequest{
		ChainID:		"aetra-test-1",
		Participants:		[]string{alice, bob},
		InitialBalances:	[]Balance{{Participant: alice, Amount: "100"}, {Participant: bob, Amount: "0"}},
		ChannelType:		ChannelTypeBidirectional,
		Collateral:		"100",
		CloseDelay:		8,
		ChallengePeriod:	8,
		FeePolicyID:		NativeDenom,
		OpeningFeeDenom:	NativeDenom,
		OpeningFeePaid:		DefaultOpeningFee,
		OpenHeight:		10,
	}
	openMsg := MsgOpenChannel{Signer: alice, Request: openReq}.Normalize()
	state := EmptyState()
	ante, err := ValidatePaymentChannelMessageFee(state, openMsg)
	require.NoError(t, err)
	require.Equal(t, PaymentFeeClassChannelOpen, ante.FeeClass)

	state, result, err := ApplyPaymentChannelMessage(state, openMsg)
	require.NoError(t, err)
	require.Equal(t, PaymentChannelMsgOpenChannel, result.MsgType)
	require.Len(t, state.Channels, 1)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
	channel := state.Channels[0]

	nextState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "90"},
		{Participant: bob, Amount: "10"},
	})
	state, result, err = ApplyPaymentChannelMessage(state, MsgSubmitCheckpoint{
		Signer:	alice,
		Request: ChannelUpdateRequest{
			ChannelID:		channel.ChannelID,
			State:			nextState,
			RegisterCheckpoint:	true,
			Submitter:		alice,
			CurrentHeight:		12,
			CheckpointFeePaid:	"0",
		},
	})
	require.NoError(t, err)
	require.True(t, result.Checkpoint.CheckpointRegistered)
	require.Equal(t, nextState.StateHash, state.Channels[0].LatestState.StateHash)

	closeMsg := MsgUnilateralClose{
		Signer:	alice,
		Request: ChannelCloseRequest{
			ChannelID:	channel.ChannelID,
			ClosingState:	nextState,
			Submitter:	alice,
			CurrentHeight:	20,
			SettlementFee:	"0",
		},
	}.Normalize()
	plan, err := PaymentChannelMessageAccessPlan(closeMsg, 20)
	require.NoError(t, err)
	require.Contains(t, plan.WriteKeys, PaymentPendingCloseIndexKey(channel.ChannelID))
	state, _, err = ApplyPaymentChannelMessage(state, closeMsg)
	require.NoError(t, err)
	require.Equal(t, ChannelStatusPendingClose, state.Channels[0].Status)
	require.NoError(t, ValidateLockedCollateralForFinality(state))

	state, result, err = ApplyPaymentChannelMessage(state, MsgFinalizeClose{
		Signer:	bob,
		Request: FinalSettlementRequest{
			ChannelID:	channel.ChannelID,
			CurrentHeight:	28,
		},
	})
	require.NoError(t, err)
	require.Equal(t, PaymentChannelMsgFinalizeClose, result.MsgType)
	require.Equal(t, ChannelStatusSettled, state.Channels[0].Status)
	require.Empty(t, state.CustodyLocks)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
	require.Len(t, state.ClosedChannels, 1)

	snapshot, err := SnapshotPaymentChannelModuleState(state, 28)
	require.NoError(t, err)
	require.Len(t, snapshot.Channels, 1)
	require.Len(t, snapshot.Participants, 2)
	require.Len(t, snapshot.Configs, 1)
	require.Len(t, snapshot.Settlements, 1)
	require.Len(t, snapshot.SettlementTombstones, 1)
	require.Len(t, snapshot.FeeAccumulators, 1)
}

func TestPaymentChannelModuleBlockSTMProfilesMessageConflicts(t *testing.T) {
	alice := testAddress(0x8d)
	bob := testAddress(0x8e)
	first := signedChannel(t, "msg-blockstm-first", "100", alice, bob)
	second := signedChannel(t, "msg-blockstm-second", "100", alice, bob)
	firstClose := signedState(t, first, 2, first.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	secondClose := signedState(t, second, 2, second.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "70"},
		{Participant: bob, Amount: "30"},
	})

	profile, err := PaymentChannelMessagesConflictProfile([]PaymentChannelModuleMessage{
		MsgUnilateralClose{Signer: alice, Request: ChannelCloseRequest{ChannelID: first.ChannelID, ClosingState: firstClose, Submitter: alice, CurrentHeight: 20}},
		MsgUnilateralClose{Signer: alice, Request: ChannelCloseRequest{ChannelID: second.ChannelID, ClosingState: secondClose, Submitter: alice, CurrentHeight: 20}},
	}, 20)
	require.NoError(t, err)
	require.True(t, profile.ConflictFree)
	require.True(t, profile.GlobalAccountingDeferred)

	profile, err = PaymentChannelMessagesConflictProfile([]PaymentChannelModuleMessage{
		MsgUnilateralClose{Signer: alice, Request: ChannelCloseRequest{ChannelID: first.ChannelID, ClosingState: firstClose, Submitter: alice, CurrentHeight: 20}},
		MsgDisputeClose{Signer: bob, Request: ChannelDisputeRequest{ChannelID: first.ChannelID, ClosingStateReference: firstClose.StateHash, NewerState: firstClose, Submitter: bob, CurrentHeight: 21}},
	}, 21)
	require.NoError(t, err)
	require.False(t, profile.ConflictFree)
	require.NotEmpty(t, profile.Conflicts)
}

func TestStoreV2LayoutUsesSpecifiedPrefixesAndCompactChannelRecords(t *testing.T) {
	alice := testAddress(0x89)
	bob := testAddress(0x8a)
	channel := signedChannel(t, "store-v2-layout", "100", alice, bob)
	channel.RoutingAdvertised = true
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	openLayout, err := BuildStoreV2Layout(state)
	require.NoError(t, err)
	require.Len(t, openLayout.ChannelStates, 1)
	require.False(t, openLayout.ChannelStates[0].SubmittedOnChain)
	require.Equal(t, channel.LatestState.StateHash, openLayout.ChannelStates[0].FullState.StateHash)
	require.Empty(t, openLayout.ChannelStates[0].FullState.Signatures)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "40"},
		{Participant: bob, Amount: "60"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "1")
	require.NoError(t, err)

	layout, err := BuildStoreV2Layout(state)
	require.NoError(t, err)
	require.Equal(t, StoreV2MigrationVersion, layout.Version)
	require.Len(t, layout.Channels, 1)
	require.Equal(t, StoreV2ChannelKey(channel.ChannelID), layout.Channels[0].Key)
	require.True(t, strings.HasPrefix(layout.Channels[0].Key, "payments/channels/"))
	require.Empty(t, layout.Channels[0].Channel.LatestState.Signatures)
	require.Equal(t, StoreV2PendingCloseKey(channel.ChannelID), layout.Channels[0].PendingCloseKey)
	require.Equal(t, PaymentRoutingAdvertisementIndexKey(channel.ChannelID), layout.Channels[0].RoutingAdvertisementKey)
	require.Len(t, layout.ParticipantChannels, 2)
	require.True(t, strings.HasPrefix(layout.ParticipantChannels[0].Key, "payments/participant_channels/"))
	require.Len(t, layout.PendingCloses, 1)
	require.Equal(t, StoreV2PendingCloseKey(channel.ChannelID), layout.PendingCloses[0].Key)

	var submittedFullState bool
	for _, record := range layout.ChannelStates {
		require.True(t, strings.HasPrefix(record.Key, "payments/channel_states/"))
		if record.SubmittedOnChain && record.Nonce == closeState.Nonce {
			submittedFullState = len(record.FullState.Signatures) == len(channel.Participants)
		}
	}
	require.True(t, submittedFullState)
}

func TestStoreV2PrunesExpiredTombstonesAndConditions(t *testing.T) {
	channelID := HashParts("store-v2-prune-channel")
	layout := StoreV2Layout{
		Version:	StoreV2MigrationVersion,
		Conditions: []StoreV2ConditionRecord{
			{Key: StoreV2ConditionKey(HashParts("expired-condition")), Version: StoreV2MigrationVersion, ConditionID: HashParts("expired-condition"), ChannelID: channelID, ExpiresHeight: 10},
			{Key: StoreV2ConditionKey(HashParts("active-condition")), Version: StoreV2MigrationVersion, ConditionID: HashParts("active-condition"), ChannelID: channelID, ExpiresHeight: 30},
			{Key: StoreV2ConditionKey(HashParts("settled-condition")), Version: StoreV2MigrationVersion, ConditionID: HashParts("settled-condition"), ChannelID: channelID, ExpiresHeight: 40, Settled: true},
		},
		SettlementTombstones: []StoreV2SettlementTombstoneRecord{
			{
				Key:		StoreV2SettlementTombstoneKey(HashParts("old-channel")),
				Version:	StoreV2MigrationVersion,
				ChannelID:	HashParts("old-channel"),
				Tombstone:	ClosedChannelTombstone{ChainID: "aetra-test-1", ChannelID: HashParts("old-channel"), FinalizedNonce: 1, StateHash: HashParts("old-state"), ClosedHeight: 5, ExpiresHeight: 15},
			},
			{
				Key:		StoreV2SettlementTombstoneKey(HashParts("kept-channel")),
				Version:	StoreV2MigrationVersion,
				ChannelID:	HashParts("kept-channel"),
				Tombstone:	ClosedChannelTombstone{ChainID: "aetra-test-1", ChannelID: HashParts("kept-channel"), FinalizedNonce: 1, StateHash: HashParts("kept-state"), ClosedHeight: 5, ExpiresHeight: 35},
			},
		},
	}
	layout = layout.Normalize()
	require.NoError(t, layout.Validate())

	pruned, err := PruneStoreV2Layout(layout, 20)
	require.NoError(t, err)
	require.Len(t, pruned.Conditions, 1)
	require.Equal(t, StoreV2ConditionKey(HashParts("active-condition")), pruned.Conditions[0].Key)
	require.Len(t, pruned.SettlementTombstones, 1)
	require.Equal(t, StoreV2SettlementTombstoneKey(HashParts("kept-channel")), pruned.SettlementTombstones[0].Key)
}

func TestStoreV2ParticipantIndexPagination(t *testing.T) {
	alice := testAddress(0x8b)
	bob := testAddress(0x8c)
	first := signedChannel(t, "store-v2-page-first", "100", alice, bob)
	second := signedChannel(t, "store-v2-page-second", "100", alice, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	layout, err := BuildStoreV2Layout(state)
	require.NoError(t, err)

	page, err := QueryStoreV2ParticipantChannels(layout, ParticipantChannelPageRequest{Address: alice, Limit: 1})
	require.NoError(t, err)
	require.Len(t, page.Entries, 1)
	require.Equal(t, uint64(2), page.Total)
	require.Equal(t, uint64(1), page.NextOffset)
	next, err := QueryStoreV2ParticipantChannels(layout, ParticipantChannelPageRequest{Address: alice, Offset: page.NextOffset, Limit: 1})
	require.NoError(t, err)
	require.Len(t, next.Entries, 1)
	require.Zero(t, next.NextOffset)
	require.NotEqual(t, page.Entries[0].ChannelID, next.Entries[0].ChannelID)
}

func TestAdaptiveSyncSnapshotRecoversNodeDuringActiveDispute(t *testing.T) {
	alice := testAddress(0x8d)
	bob := testAddress(0x8e)
	channel := signedChannel(t, "adaptive-sync-dispute", "100", alice, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, channel)
	require.NoError(t, err)
	state, err = RegisterRoutingEdge(state, ChannelEdge{ChannelID: channel.ChannelID, From: alice, To: bob, Capacity: "100", FeeAmount: "1", Active: true})
	require.NoError(t, err)
	withoutRouting := state.Clone()
	withoutRouting.Edges = nil

	closeState := signedConditionalState(t, channel, 2, channel.OpeningStateHash, "25", []Balance{
		{Participant: alice, Amount: "75"},
		{Participant: bob, Amount: "0"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	withoutRouting, err = SubmitClose(withoutRouting, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	newer := signedConditionalState(t, channel, 3, closeState.StateHash, "25", []Balance{
		{Participant: alice, Amount: "70"},
		{Participant: bob, Amount: "5"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newer, bob, 21)
	require.NoError(t, err)
	withoutRouting, err = DisputeClose(withoutRouting, channel.ChannelID, newer, bob, 21)
	require.NoError(t, err)

	snapshot, err := BuildAdaptiveSyncSnapshot(state, 22)
	require.NoError(t, err)
	require.NoError(t, snapshot.Validate())
	noRoutingSnapshot, err := BuildAdaptiveSyncSnapshot(withoutRouting, 22)
	require.NoError(t, err)
	require.Equal(t, noRoutingSnapshot.SnapshotHash, snapshot.SnapshotHash)
	require.True(t, snapshot.ConsensusOnly)
	require.True(t, snapshot.RoutingTopologyExcluded)
	require.Len(t, snapshot.ActiveDisputes, 1)
	require.Equal(t, channel.ChannelID, snapshot.ActiveDisputes[0].ChannelID)
	require.Len(t, snapshot.PendingFinalizations, 1)
	require.Equal(t, state.Channels[0].PendingClose.SettleAfterHeight, snapshot.PendingFinalizations[0].PendingHeight)
	require.Len(t, snapshot.Layout.PendingCloses, 1)
	require.Len(t, snapshot.Layout.Conditions, 1)
	require.NotEmpty(t, snapshot.WatcherReplayEvents)
	require.True(t, strings.HasPrefix(snapshot.WatcherReplayEvents[0].Key, "payments/watcher_replay_events/"))

	recovered, err := RecoverAdaptiveSyncSafety(snapshot)
	require.NoError(t, err)
	require.Equal(t, snapshot.SnapshotHash, recovered.RecoveredFromSnapshotHash)
	require.Contains(t, recovered.ActiveDisputeChannelIDs, channel.ChannelID)
	require.Contains(t, recovered.PendingCloseChannelIDs, channel.ChannelID)
	require.Contains(t, recovered.PendingFinalizationIDs, channel.ChannelID)
	require.Contains(t, recovered.UnresolvedConditionIDs, newer.Conditions[0].ConditionID)
	require.NotEmpty(t, recovered.WatcherReplayEventIDs)
}

func TestAdaptiveSyncSnapshotIncludesVirtualAnchors(t *testing.T) {
	alice := testAddress(0x8f)
	router := testAddress(0x90)
	bob := testAddress(0x91)
	state, vc, _ := virtualChannelFixture(t, "adaptive-sync-virtual", alice, router, bob, "100", 40)

	snapshot, err := BuildAdaptiveSyncSnapshot(state, 30)
	require.NoError(t, err)
	require.Len(t, snapshot.Layout.VirtualChannels, 1)
	require.Equal(t, StoreV2VirtualChannelKey(vc.VirtualChannelID), snapshot.Layout.VirtualChannels[0].Key)
	recovered, err := RecoverAdaptiveSyncSafety(snapshot)
	require.NoError(t, err)
	require.Contains(t, recovered.VirtualChannelIDs, vc.VirtualChannelID)
}

func BenchmarkPaymentChannelOpenAccessPlan(b *testing.B) {
	op := benchmarkSettlementOperation("bench-open", BatchOperationOpen, 1)
	for i := 0; i < b.N; i++ {
		if _, err := AccessPlanForSettlementOperation(op, 100); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentChannelCloseAccessPlan(b *testing.B) {
	op := benchmarkSettlementOperation("bench-close", BatchOperationClose, 1)
	for i := 0; i < b.N; i++ {
		if _, err := AccessPlanForSettlementOperation(op, 100); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentChannelDisputeAccessPlan(b *testing.B) {
	op := benchmarkSettlementOperation("bench-dispute", BatchOperationDispute, 1)
	for i := 0; i < b.N; i++ {
		if _, err := AccessPlanForSettlementOperation(op, 100); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentBatchSettlementGrouping(b *testing.B) {
	ops := make([]SettlementOperation, 128)
	for i := range ops {
		ops[i] = benchmarkSettlementOperation("bench-batch", BatchOperationSettle, uint64(i+1))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := GroupSettlementOperationsByChannelKey("bench-batch", ops); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkSettlementOperation(seed string, opType BatchOperationType, nonce uint64) SettlementOperation {
	channelID := HashParts(seed, "channel", fmt.Sprintf("%020d", nonce))
	stateHash := HashParts(seed, "state", fmt.Sprintf("%020d", nonce))
	return SettlementOperation{
		OperationID:	HashParts(seed, "op", fmt.Sprintf("%020d", nonce), string(opType)),
		OperationType:	opType,
		ChannelID:	channelID,
		Nonce:		nonce,
		StateHash:	stateHash,
	}
}

func blockSTMConflictKeys(conflicts []BlockSTMConflict) []string {
	out := make([]string, 0, len(conflicts))
	for _, conflict := range conflicts {
		out = append(out, conflict.Key)
	}
	return out
}

func TestSecurityThreatModelCoversRequiredThreats(t *testing.T) {
	threats := DefaultThreatModel()
	require.NoError(t, ValidateThreatModelCoverage(threats))
	require.Len(t, threats, 15)

	seen := make(map[SecurityThreat]struct{}, len(threats))
	for _, entry := range threats {
		require.True(t, IsSecurityThreat(entry.Threat))
		require.NotEmpty(t, entry.Controls)
		for _, control := range entry.Controls {
			require.True(t, IsSecurityGuarantee(control))
		}
		seen[entry.Threat] = struct{}{}
	}
	require.Contains(t, seen, SecurityThreatStaleStateClose)
	require.Contains(t, seen, SecurityThreatSameNonceDoubleSign)
	require.Contains(t, seen, SecurityThreatReplayAcrossDomain)
	require.Contains(t, seen, SecurityThreatSettlementBatchConflictAmplify)

	require.ErrorContains(t, ValidateThreatModelCoverage(threats[:len(threats)-1]), "missing payments security threat")
	withDuplicate := append([]ThreatModelEntry{}, threats...)
	withDuplicate[1] = withDuplicate[0]
	require.ErrorContains(t, ValidateThreatModelCoverage(withDuplicate), "duplicate payments security threat")
}

func TestSecurityModelReportTracksCloseDisputeReplayAndCollateralGuarantees(t *testing.T) {
	alice := testAddress(0xf1)
	bob := testAddress(0xf2)
	channel := signedChannel(t, "security-model-report", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	report, err := BuildSecurityModelReport(state)
	require.NoError(t, err)
	require.NoError(t, report.Validate())
	require.Len(t, report.Guarantees, 8)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
	report, err = BuildSecurityModelReport(state)
	require.NoError(t, err)
	require.NoError(t, report.Validate())

	newerState := signedState(t, channel, 3, closeState.StateHash, []Balance{
		{Participant: alice, Amount: "425"},
		{Participant: bob, Amount: "575"},
	})
	state, err = DisputeClose(state, channel.ChannelID, newerState, bob, 21)
	require.NoError(t, err)
	require.Equal(t, newerState.Nonce, state.Channels[0].PendingClose.State.Nonce)
	require.Equal(t, newerState.Nonce, state.Channels[0].DisputedNonce)
	require.NoError(t, ValidateLockedCollateralForFinality(state))
	report, err = BuildSecurityModelReport(state)
	require.NoError(t, err)
	require.NoError(t, report.Validate())

	state, settlement, err := FinalizeSettlement(state, channel.ChannelID, 40)
	require.NoError(t, err)
	require.NoError(t, settlement.ValidateForChannel(state.Channels[0]))
	require.NoError(t, ValidateLockedCollateralForFinality(state))
	report, err = BuildSecurityModelReport(state)
	require.NoError(t, err)
	require.NoError(t, report.Validate())
	require.NotEmpty(t, state.ClosedChannels)
}

func TestSecurityModelReportFailsWhenGuaranteeStateIsBroken(t *testing.T) {
	alice := testAddress(0xf3)
	bob := testAddress(0xf4)
	channel := signedChannel(t, "security-model-broken", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	state.Channels[0].Participants = nil

	_, err := BuildSecurityModelReport(state)
	require.ErrorContains(t, err, string(SecurityGuaranteeUnilateralClose))
}

func TestSecurityModelUsesPenaltyAndConditionEnforcement(t *testing.T) {
	alice := testAddress(0xf5)
	bob := testAddress(0xf6)
	channel := signedChannel(t, "security-model-enforcement", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	entry, err := PenaltyMatrixEntryForProof(FraudProofTypeDoubleSign, DefaultPenaltyMatrix())
	require.NoError(t, err)
	require.Equal(t, PenaltyClassDoubleSign, entry.Class)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "480"},
		{Participant: bob, Amount: "520"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	conflicting := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	state, err = SubmitFraudProof(state, channel.ChannelID, FraudProof{
		ProofID:		HashParts("security-double-sign", channel.ChannelID),
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		bob,
		OffendingSigner:	alice,
		StateA:			closeState,
		StateB:			conflicting,
		PenaltyAmount:		"10",
		EvidenceHash:		HashParts("security-double-sign-evidence", closeState.StateHash, conflicting.StateHash),
	}, 21)
	require.NoError(t, err)
	require.Len(t, state.Channels[0].PendingClose.Penalties, 1)

	promise := signedPromiseWithHashLock(t, channel, "security-condition", alice, bob, "10", "0", 7, 11, HashParts("security-preimage"))
	_, _, err = RevealPromisePreimage(EmptyStateWithChannel(t, channel), PreimageRevealRequest{
		ChannelID:	channel.ChannelID,
		Promises:	[]ConditionalPromise{promise},
		Preimage:	"wrong-preimage",
		Revealer:	bob,
		CurrentHeight:	10,
	})
	require.ErrorContains(t, err, "does not satisfy hash lock")
}

func TestEconomicFinalityRequirementsChallengeSizingAndReport(t *testing.T) {
	alice := testAddress(0xfa)
	bob := testAddress(0xfb)
	channel := signedChannel(t, "economic-finality", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)
	sizing := ChallengePeriodSizing{
		MessagePropagationDelay:	1,
		WatchServiceReaction:		2,
		CongestionBuffer:		1,
		MultiHopTimeoutMargin:		1,
	}

	require.NoError(t, ValidateEconomicFinalityRequirements(DefaultEconomicFinalityRequirements()))
	require.NoError(t, ValidateChallengePeriodSizing(channel.DisputePeriod, sizing))
	require.ErrorContains(t, ValidateChallengePeriodSizing(sizing.TotalRequired(), sizing), "must exceed")

	report, err := BuildEconomicFinalityReport(state, 20, sizing)
	require.NoError(t, err)
	require.NoError(t, report.Validate())
	require.Equal(t, sizing.TotalRequired(), report.RequiredChallengeSize)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "450"},
		{Participant: bob, Amount: "550"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	report, err = BuildEconomicFinalityReport(state, 21, sizing)
	require.NoError(t, err)
	require.NoError(t, report.Validate())

	early := state.Clone()
	early.Channels[0].Finality = ChannelFinalityFinalizable
	_, err = BuildEconomicFinalityReport(early, 21, sizing)
	require.ErrorContains(t, err, string(EconomicFinalityUnilateral))
}

func TestEconomicFinalityReportCoversVirtualAndPenaltySettlement(t *testing.T) {
	alice := testAddress(0xfc)
	router := testAddress(0xfd)
	bob := testAddress(0xfe)
	sizing := ChallengePeriodSizing{
		MessagePropagationDelay:	1,
		WatchServiceReaction:		2,
		CongestionBuffer:		1,
		MultiHopTimeoutMargin:		1,
	}

	state, _, _ := virtualChannelFixture(t, "economic-finality-virtual", alice, router, bob, "100", 40)
	report, err := BuildEconomicFinalityReport(state, 25, sizing)
	require.NoError(t, err)
	require.NoError(t, report.Validate())

	channel := state.Channels[0]
	closeState := signedState(t, channel, 3, channel.LatestState.StateHash, []Balance{
		{Participant: channel.Participants[0], Amount: "450"},
		{Participant: channel.Participants[1], Amount: "450"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, channel.Participants[0], 26, "0")
	require.NoError(t, err)
	conflicting := signedState(t, channel, 3, channel.LatestState.StateHash, []Balance{
		{Participant: channel.Participants[0], Amount: "440"},
		{Participant: channel.Participants[1], Amount: "460"},
	})
	proofID := HashParts("economic-finality-penalty", channel.ChannelID)
	state, err = SubmitFraudProof(state, channel.ChannelID, FraudProof{
		ProofID:		proofID,
		ProofType:		FraudProofTypeDoubleSign,
		SubmittedBy:		channel.Participants[1],
		OffendingSigner:	channel.Participants[0],
		StateA:			closeState,
		StateB:			conflicting,
		EvidenceHash:		ComputeDisputeProofHash(FraudProof{ProofID: proofID, ProofType: FraudProofTypeDoubleSign, StateA: closeState, StateB: conflicting}),
		PenaltyAmount:		"10",
	}, 27)
	require.NoError(t, err)
	require.Equal(t, ChannelFinalityPenalized, state.Channels[0].Finality)
	report, err = BuildEconomicFinalityReport(state, 27, sizing)
	require.NoError(t, err)
	require.NoError(t, report.Validate())
}

func TestDisputePriorityPolicyNearExpiryFraudAndStressInclusion(t *testing.T) {
	first := HashParts("priority-channel-first")
	second := HashParts("priority-channel-second")
	policy := DefaultDisputePriorityPolicy()
	require.NoError(t, policy.Validate())

	normal, err := ComputeDisputeTransactionPriority(policy, DisputePriorityRequest{
		Operation:		SettlementArbitrationDispute,
		ChannelID:		first,
		SubmittedHeight:	20,
		CurrentHeight:		21,
		SettleAfterHeight:	28,
		FeePaid:		"4",
		RequiredFee:		"4",
		EstimatedGas:		10,
	})
	require.NoError(t, err)
	critical, err := ComputeDisputeTransactionPriority(policy, DisputePriorityRequest{
		Operation:		SettlementArbitrationFraudProof,
		ChannelID:		second,
		SubmittedHeight:	20,
		CurrentHeight:		27,
		SettleAfterHeight:	28,
		HasFraudProof:		true,
		FeePaid:		"4",
		RequiredFee:		"4",
		EstimatedGas:		10,
		CongestionBps:		9_000,
	})
	require.NoError(t, err)
	require.True(t, critical.PriorityScore > normal.PriorityScore)
	require.True(t, critical.NearExpiry)
	require.True(t, critical.Deterministic)

	repeated, err := ComputeDisputeTransactionPriority(policy, DisputePriorityRequest{
		Operation:		SettlementArbitrationFraudProof,
		ChannelID:		second,
		SubmittedHeight:	20,
		CurrentHeight:		27,
		SettleAfterHeight:	28,
		HasFraudProof:		true,
		FeePaid:		"4",
		RequiredFee:		"4",
		EstimatedGas:		10,
		CongestionBps:		9_000,
	})
	require.NoError(t, err)
	require.Equal(t, critical.DecisionHash, repeated.DecisionHash)

	stress, err := SimulateDisputeInclusionStress(policy, []DisputePriorityRequest{
		{Operation: SettlementArbitrationDispute, ChannelID: first, SubmittedHeight: 20, CurrentHeight: 21, SettleAfterHeight: 28, FeePaid: "4", RequiredFee: "4", EstimatedGas: 10},
		{Operation: SettlementArbitrationFraudProof, ChannelID: second, SubmittedHeight: 20, CurrentHeight: 27, SettleAfterHeight: 28, HasFraudProof: true, FeePaid: "4", RequiredFee: "4", EstimatedGas: 10, CongestionBps: 9_000},
		{Operation: SettlementArbitrationDispute, ChannelID: first, SubmittedHeight: 20, CurrentHeight: 26, SettleAfterHeight: 28, FeePaid: "0", RequiredFee: "4", EstimatedGas: 10},
	}, 20)
	require.NoError(t, err)
	require.Len(t, stress.Included, 2)
	require.Len(t, stress.Deferred, 1)
	require.Equal(t, "insufficient dispute fee", stress.Deferred[0].DeferredReason)
	require.False(t, stress.ConflictFree)
	require.NotEmpty(t, stress.Conflicts)

	parallel, err := SimulateDisputeInclusionStress(policy, []DisputePriorityRequest{
		{Operation: SettlementArbitrationDispute, ChannelID: first, SubmittedHeight: 20, CurrentHeight: 21, SettleAfterHeight: 28, FeePaid: "4", RequiredFee: "4", EstimatedGas: 10},
		{Operation: SettlementArbitrationFraudProof, ChannelID: second, SubmittedHeight: 20, CurrentHeight: 27, SettleAfterHeight: 28, HasFraudProof: true, FeePaid: "4", RequiredFee: "4", EstimatedGas: 10},
	}, 20)
	require.NoError(t, err)
	require.True(t, parallel.ConflictFree)
	require.Empty(t, parallel.Conflicts)
}

func TestNearExpiryDisputeMonitoringAndValidatorWatcherFormat(t *testing.T) {
	alice := testAddress(0xaa)
	bob := testAddress(0xab)
	validator := testAddress(0xac)
	service := testAddress(0xad)
	channel := signedChannel(t, "near-expiry-monitor", "1000", alice, bob)
	state := EmptyStateWithChannel(t, channel)

	var err error
	state, err = RegisterValidatorPaymentService(state, ValidatorPaymentServiceMetadata{
		ValidatorAddress:	validator,
		ServiceAddress:		service,
		WatchEndpoint:		"https://validator.example/watch",
		PublicKey:		"validator-watch-key",
		MinDelegation:		"10",
		Active:			true,
		UpdatedHeight:		12,
	})
	require.NoError(t, err)
	state, err = RegisterValidatorWatchService(state, ValidatorWatchRegistration{
		ValidatorAddress:	validator,
		Delegator:		bob,
		RegisteredHeight:	13,
	})
	require.NoError(t, err)
	require.Equal(t, service, state.ValidatorWatchRegistries[0].ServiceAddress)
	require.Equal(t, state.ValidatorPaymentServices[0].MetadataHash, state.ValidatorWatchRegistries[0].MetadataHash)

	closeState := signedState(t, channel, 2, channel.OpeningStateHash, []Balance{
		{Participant: alice, Amount: "500"},
		{Participant: bob, Amount: "500"},
	})
	state, err = SubmitClose(state, channel.ChannelID, closeState, alice, 20, "0")
	require.NoError(t, err)
	alerts, err := MonitorNearExpiryDisputes(state, 27, 2)
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	require.Equal(t, channel.ChannelID, alerts[0].ChannelID)
	require.Equal(t, uint64(1), alerts[0].BlocksRemaining)
	require.Equal(t, "critical", alerts[0].Severity)
	require.NotEmpty(t, alerts[0].EvidenceHash)
}

func signedChannel(t *testing.T, salt, collateral, left, right string) ChannelRecord {
	t.Helper()

	channelID := HashParts(salt, left, right)
	channel := ChannelRecord{
		ChainID:		"aetra-test-1",
		ChannelID:		channelID,
		ChannelType:		ChannelTypeBidirectional,
		Participants:		[]string{left, right},
		Denom:			NativeDenom,
		Collateral:		collateral,
		OpenHeight:		10,
		CloseDelay:		8,
		DisputePeriod:		8,
		OpeningFeePaid:		DefaultOpeningFee,
		ConditionalPayments:	true,
		CustodyDenom:		NativeDenom,
		CustodyAmount:		collateral,
		Status:			ChannelStatusOpen,
	}
	openState := signedState(t, channel, 1, "", []Balance{
		{Participant: left, Amount: collateral},
		{Participant: right, Amount: "0"},
	})
	channel.LatestState = openState
	channel.OpeningStateHash = openState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func EmptyStateWithChannel(t *testing.T, channel ChannelRecord) PaymentsState {
	t.Helper()

	state, err := OpenChannel(EmptyState(), channel)
	require.NoError(t, err)
	return state
}

func signedState(t *testing.T, channel ChannelRecord, nonce uint64, previous string, balances []Balance) ChannelState {
	t.Helper()

	state, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			nonce,
		Balances:		balances,
		PreviousStateHash:	previous,
		TimeoutHeight:		channel.OpenHeight + channel.DisputePeriod + nonce,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func signedReserveState(t *testing.T, channel ChannelRecord, nonce uint64, previous, reserveA, reserveB string, balances []Balance) ChannelState {
	t.Helper()

	state, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			nonce,
		Balances:		balances,
		ReserveA:		reserveA,
		ReserveB:		reserveB,
		PreviousStateHash:	previous,
		TimeoutHeight:		channel.OpenHeight + channel.DisputePeriod + 70,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func signedPromise(t *testing.T, channel ChannelRecord, salt, source, destination, amount, fee string, nonce, timeoutHeight uint64) ConditionalPromise {
	t.Helper()

	return signedLinkedPromise(t, channel, HashParts("promise", channel.ChannelID, salt), source, destination, amount, fee, nonce, timeoutHeight, HashParts("promise-preimage", salt), "", "")
}

func signedPromiseWithHashLock(t *testing.T, channel ChannelRecord, salt, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock string) ConditionalPromise {
	t.Helper()

	return signedLinkedPromise(t, channel, HashParts("promise", channel.ChannelID, salt), source, destination, amount, fee, nonce, timeoutHeight, hashLock, "", "")
}

func signedLinkedPromise(t *testing.T, channel ChannelRecord, promiseID, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock, previousID, nextID string) ConditionalPromise {
	t.Helper()

	return signedRoutePromise(t, channel, promiseID, "", source, destination, amount, fee, nonce, timeoutHeight, hashLock, previousID, nextID)
}

func signedRoutePromise(t *testing.T, channel ChannelRecord, promiseID, routeID, source, destination, amount, fee string, nonce, timeoutHeight uint64, hashLock, previousID, nextID string) ConditionalPromise {
	t.Helper()

	promise, err := BuildConditionalPromise(ConditionalPromise{
		PromiseID:			promiseID,
		ChannelID:			channel.ChannelID,
		Source:				source,
		Destination:			destination,
		Amount:				amount,
		Fee:				fee,
		HashLock:			hashLock,
		TimeoutHeight:			timeoutHeight,
		TimeoutTimestamp:		int64(timeoutHeight * 10),
		ConditionType:			ConditionTypeHashLock,
		RouteIDOptional:		routeID,
		PreviousPromiseIDOptional:	previousID,
		NextPromiseIDOptional:		nextID,
		Nonce:				nonce,
	})
	require.NoError(t, err)
	promise.Signature, err = SignatureForPromise(channel, promise, source)
	require.NoError(t, err)
	promise = promise.Normalize()
	return promise
}

func signedGossipEnvelope(t *testing.T, message GossipMessage, signer string, receivedAt uint64) SignedGossipEnvelope {
	t.Helper()

	built, err := BuildGossipMessage(message)
	require.NoError(t, err)
	sig, err := SignatureForGossip(built, signer)
	require.NoError(t, err)
	return SignedGossipEnvelope{
		Message:	built,
		MessageHash:	built.MessageID,
		Signature:	sig,
		ReceivedFrom:	signer,
		ReceivedAt:	receivedAt,
	}.Normalize()
}

func routingEngineEnvelope(t *testing.T, message GossipMessage, signer string, receivedAt uint64) SignedGossipEnvelope {
	t.Helper()

	envelope, err := BuildRoutingGossipEnvelope(message, signer, receivedAt)
	require.NoError(t, err)
	return envelope
}

func signedVirtualChannel(t *testing.T, vc VirtualChannel, chainID string) VirtualChannel {
	t.Helper()

	vc.ChainID = chainID
	vc.Signatures = nil
	built, err := BuildVirtualChannel(vc)
	require.NoError(t, err)
	signers := normalizeAddressSet(append(append([]string{}, built.Endpoints...), built.Intermediaries...))
	for _, signer := range signers {
		sig, err := SignatureForVirtualChannel(built, signer)
		require.NoError(t, err)
		built.Signatures = append(built.Signatures, sig)
	}
	built = built.Normalize()
	require.NoError(t, ValidateVirtualChannelActivation(built))
	return built
}

func signedVirtualActivationProof(t *testing.T, vc VirtualChannel, signer string, routeTimeoutHeight uint64) VirtualActivationProof {
	t.Helper()

	reserves := make([]VirtualParentReserve, 0, len(vc.ParentChannelIDs))
	for _, parentID := range vc.ParentChannelIDs {
		reserve, err := BuildVirtualParentReserve(vc, VirtualParentReserve{
			ParentChannelID:	parentID,
			Capacity:		vc.Capacity,
			FeeAmount:		"1",
		}, signer)
		require.NoError(t, err)
		reserves = append(reserves, reserve)
	}
	proof, err := BuildVirtualActivationProof(vc, reserves, routeTimeoutHeight)
	require.NoError(t, err)
	require.NoError(t, ValidateVirtualActivationProof(proof))
	return proof
}

func virtualChannelFixture(t *testing.T, salt, alice, router, bob, capacity string, expiresHeight uint64) (PaymentsState, VirtualChannel, VirtualActivationProof) {
	t.Helper()

	first := signedChannel(t, salt+"-first", "900", alice, router)
	second := signedChannel(t, salt+"-second", "900", router, bob)
	state := EmptyState()
	var err error
	state, err = OpenChannel(state, first)
	require.NoError(t, err)
	state, err = OpenChannel(state, second)
	require.NoError(t, err)
	firstReserve := signedReserveState(t, first, 2, first.OpeningStateHash, capacity, "0", []Balance{
		{Participant: first.Participants[0], Amount: "800"},
		{Participant: first.Participants[1], Amount: "0"},
	})
	secondReserve := signedReserveState(t, second, 2, second.OpeningStateHash, capacity, "0", []Balance{
		{Participant: second.Participants[0], Amount: "800"},
		{Participant: second.Participants[1], Amount: "0"},
	})
	state, err = AcceptSignedState(state, first.ChannelID, firstReserve, 20)
	require.NoError(t, err)
	state, err = AcceptSignedState(state, second.ChannelID, secondReserve, 20)
	require.NoError(t, err)
	vc := signedVirtualChannel(t, VirtualChannel{
		VirtualChannelID:	HashParts(salt, alice, bob),
		ParentChannelIDs:	[]string{first.ChannelID, second.ChannelID},
		Endpoints:		[]string{alice, bob},
		Intermediaries:		[]string{router},
		Capacity:		capacity,
		BalanceA:		capacity,
		BalanceB:		"0",
		ExpiresHeight:		expiresHeight,
	}, first.ChainID)
	proof := signedVirtualActivationProof(t, vc, router, 80)
	state, err = OpenVirtualChannelWithProof(state, proof)
	require.NoError(t, err)
	return state, vc, proof
}

func signedVirtualReserve(t *testing.T, vc VirtualChannel, parentID, signer, capacity, splitAmount string) VirtualParentReserve {
	t.Helper()

	reserve, err := BuildVirtualParentReserve(vc, VirtualParentReserve{
		ParentChannelID:	parentID,
		Capacity:		capacity,
		SplitAmount:		splitAmount,
		FeeAmount:		"0",
	}, signer)
	require.NoError(t, err)
	return reserve
}

func signedVirtualEndpointUpdate(t *testing.T, vc VirtualChannel, nonce uint64, balanceA, balanceB string) VirtualChannel {
	t.Helper()

	update := vc.Normalize()
	update.Nonce = nonce
	update.BalanceA = balanceA
	update.BalanceB = balanceB
	update.Signatures = nil
	update.AnchorCommitment = ""
	update.StateHash = ""
	built, err := BuildVirtualChannel(update)
	require.NoError(t, err)
	for _, signer := range built.Endpoints {
		sig, err := SignatureForVirtualChannel(built, signer)
		require.NoError(t, err)
		built.Signatures = append(built.Signatures, sig)
	}
	built = built.Normalize()
	require.NoError(t, built.ValidateCore())
	return built
}

func virtualReserveCommitments(proof VirtualActivationProof) []string {
	proof = proof.Normalize()
	out := make([]string, 0, len(proof.ParentReserves))
	for _, reserve := range proof.ParentReserves {
		out = append(out, reserve.ReserveCommitment)
	}
	return out
}

func mutateCanonicalState(state ChannelState, mutate func(*ChannelState)) ChannelState {
	state = state.Normalize()
	state.Signatures = nil
	mutate(&state)
	state.SignaturePreimageHash = ComputeStateSignaturePreimageHash(state)
	state.StateHash = ComputeStateHash(state)
	return state.Normalize()
}

func resignState(t *testing.T, channel ChannelRecord, state ChannelState) ChannelState {
	t.Helper()

	state.Signatures = nil
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	return state.Normalize()
}

func signedConditionalState(t *testing.T, channel ChannelRecord, nonce uint64, previous, conditionAmount string, balances []Balance) ChannelState {
	t.Helper()

	channel = channel.Normalize()
	condition := ConditionalPayment{
		ConditionID:	HashParts("condition", channel.ChannelID, conditionAmount),
		ConditionType:	ConditionTypeHashLock,
		Payer:		channel.Participants[0],
		Payee:		channel.Participants[1],
		Amount:		conditionAmount,
		HashLock:	HashParts("condition-preimage", conditionAmount),
		TimeoutHeight:	channel.OpenHeight + channel.DisputePeriod + nonce + 2,
		NonceStart:	nonce,
		NonceEnd:	nonce + 2,
	}
	state, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			nonce,
		Balances:		balances,
		ReserveA:		"25",
		ReserveB:		"0",
		Conditions:		[]ConditionalPayment{condition},
		PreviousStateHash:	previous,
		TimeoutHeight:		channel.OpenHeight + channel.DisputePeriod + nonce,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func signedUnidirectionalChannel(t *testing.T, salt, collateral, payer, receiver string, ackRequired bool) ChannelRecord {
	t.Helper()

	channel := ChannelRecord{
		ChainID:		"aetra-test-1",
		ChannelID:		HashParts(salt, payer, receiver),
		ChannelType:		ChannelTypeUnidirectional,
		Participants:		[]string{payer, receiver},
		Payer:			payer,
		Receiver:		receiver,
		ReceiverAckRequired:	ackRequired,
		Denom:			NativeDenom,
		Collateral:		collateral,
		OpenHeight:		10,
		CloseDelay:		8,
		DisputePeriod:		8,
		ExpirationHeight:	72,
		ExpirationTimestamp:	0,
		OpeningFeePaid:		DefaultOpeningFee,
		CustodyDenom:		NativeDenom,
		CustodyAmount:		collateral,
		Status:			ChannelStatusOpen,
	}
	openState, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			1,
		Balances:		[]Balance{{Participant: payer, Amount: collateral}, {Participant: receiver, Amount: "0"}},
		TimeoutHeight:		channel.ExpirationHeight,
		TimeoutTimestamp:	channel.ExpirationTimestamp,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
	})
	require.NoError(t, err)
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(openState, signer)
		require.NoError(t, err)
		openState.Signatures = append(openState.Signatures, sig)
	}
	channel.LatestState = openState.Normalize()
	channel.OpeningStateHash = channel.LatestState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func signedUnidirectionalClaim(t *testing.T, channel ChannelRecord, claimed string, nonce, expirationHeight uint64, ack bool) UnidirectionalClaim {
	t.Helper()

	claim, err := BuildUnidirectionalClaim(UnidirectionalClaim{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Payer:			channel.Payer,
		Receiver:		channel.Receiver,
		LockedAmount:		channel.Collateral,
		ClaimedAmount:		claimed,
		Nonce:			nonce,
		ExpirationHeight:	expirationHeight,
		ExpirationTimestamp:	channel.ExpirationTimestamp,
	})
	require.NoError(t, err)
	claim.PayerSignature, err = SignatureForClaim(claim, channel.Payer)
	require.NoError(t, err)
	if ack {
		claim.ReceiverAckOptional, err = SignatureForClaim(claim, channel.Receiver)
		require.NoError(t, err)
	}
	claim = claim.Normalize()
	if !channel.ReceiverAckRequired || ack {
		require.NoError(t, claim.ValidateForChannel(channel))
	}
	return claim
}

func signedAsyncChannel(t *testing.T, salt, collateral string, balances []Balance, sendWindow, receiveWindow uint64, maxUnacked string, expiryHeight uint64, participants ...string) ChannelRecord {
	t.Helper()

	channel := ChannelRecord{
		ChainID:	"aetra-test-1",
		ChannelID:	HashParts(append([]string{salt}, participants...)...),
		ChannelType:	ChannelTypeAsync,
		Participants:	participants,
		Denom:		NativeDenom,
		Collateral:	collateral,
		OpenHeight:	10,
		CloseDelay:	8,
		DisputePeriod:	8,
		OpeningFeePaid:	DefaultOpeningFee,
		CustodyDenom:	NativeDenom,
		CustodyAmount:	collateral,
		Status:		ChannelStatusOpen,
	}
	openState, err := BuildState(ChannelState{
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		ChannelType:		channel.ChannelType,
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			1,
		Balances:		balances,
		CheckpointNonce:	1,
		CheckpointBalances:	balances,
		AsyncUpdateRoot:	ComputeAsyncDeltaRootForChannel(channel, nil),
		AcceptedUpdateRoot:	ComputeAsyncDeltaRootForChannel(channel, nil),
		SendWindow:		sendWindow,
		ReceiveWindow:		receiveWindow,
		MaxUnackedAmount:	maxUnacked,
		ExpiryHeight:		expiryHeight,
		TimeoutHeight:		expiryHeight,
		CloseDelay:		channel.DisputePeriod,
		FeePolicyID:		NativeDenom,
	})
	require.NoError(t, err)
	channel.LatestState = signAsyncCheckpoint(t, channel, openState)
	channel.OpeningStateHash = channel.LatestState.StateHash
	require.NoError(t, channel.Validate())
	return channel.Normalize()
}

func signedAsyncDelta(t *testing.T, channel ChannelRecord, salt, from, to, amount string, nonceStart, nonceEnd, expiryHeight uint64) AsyncPaymentDelta {
	t.Helper()

	delta, err := BuildAsyncDelta(AsyncPaymentDelta{
		UpdateID:	HashParts("async-delta", channel.ChannelID, salt),
		ChainID:	channel.ChainID,
		ChannelID:	channel.ChannelID,
		From:		from,
		To:		to,
		Direction:	AsyncDeltaDirection(from, to),
		Amount:		amount,
		NonceStart:	nonceStart,
		NonceEnd:	nonceEnd,
		ExpiryHeight:	expiryHeight,
	})
	require.NoError(t, err)
	delta.Signature, err = SignatureForAsyncDelta(delta, from)
	require.NoError(t, err)
	require.NoError(t, delta.ValidateForChannel(channel, channel.OpenHeight))
	return delta.Normalize()
}

func signAsyncCheckpoint(t *testing.T, channel ChannelRecord, state ChannelState) ChannelState {
	t.Helper()

	state.Signatures = nil
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		require.NoError(t, err)
		state.Signatures = append(state.Signatures, sig)
	}
	state = state.Normalize()
	require.NoError(t, state.ValidateForChannel(channel, true))
	return state
}

func testAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(bytes20(fill)))
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}

func amountFor(balances []Balance, participant string) string {
	for _, balance := range balances {
		if balance.Participant == participant {
			return balance.Amount
		}
	}
	return ""
}

func lockedAmountForType(values []PaymentLockedByChannelType, channelType ChannelType) string {
	for _, value := range values {
		if value.ChannelType == channelType {
			return value.Amount
		}
	}
	return "0"
}

func absInt64(value int64) int64 {
	if value == -9223372036854775808 {
		return 9223372036854775807
	}
	if value < 0 {
		return -value
	}
	return value
}

func paymentEventTypes(events []PaymentEvent) []string {
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.EventType)
	}
	return types
}

func allocationAmountFor(allocations []PenaltyAllocation, route PenaltyRoute) string {
	for _, allocation := range allocations {
		if allocation.Route == route {
			return allocation.Amount
		}
	}
	return ""
}
