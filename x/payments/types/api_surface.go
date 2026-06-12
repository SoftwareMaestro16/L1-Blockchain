package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type PaymentAPIMessageName string
type PaymentAPIQueryName string
type PaymentAPIEventName string

const (
	PaymentAPIMsgOpenChannel			PaymentAPIMessageName	= "MsgOpenChannel"
	PaymentAPIMsgCooperativeClose			PaymentAPIMessageName	= "MsgCooperativeClose"
	PaymentAPIMsgUnilateralClose			PaymentAPIMessageName	= "MsgUnilateralClose"
	PaymentAPIMsgDisputeClose			PaymentAPIMessageName	= "MsgDisputeClose"
	PaymentAPIMsgFinalizeClose			PaymentAPIMessageName	= "MsgFinalizeClose"
	PaymentAPIMsgSubmitCheckpoint			PaymentAPIMessageName	= "MsgSubmitCheckpoint"
	PaymentAPIMsgRegisterPromise			PaymentAPIMessageName	= "MsgRegisterPromise"
	PaymentAPIMsgResolvePromise			PaymentAPIMessageName	= "MsgResolvePromise"
	PaymentAPIMsgExpirePromise			PaymentAPIMessageName	= "MsgExpirePromise"
	PaymentAPIMsgBatchResolvePromises		PaymentAPIMessageName	= "MsgBatchResolvePromises"
	PaymentAPIMsgOpenVirtualChannel			PaymentAPIMessageName	= "MsgOpenVirtualChannel"
	PaymentAPIMsgCloseVirtualChannel		PaymentAPIMessageName	= "MsgCloseVirtualChannel"
	PaymentAPIMsgDisputeVirtualChannel		PaymentAPIMessageName	= "MsgDisputeVirtualChannel"
	PaymentAPIMsgSubmitFraudProof			PaymentAPIMessageName	= "MsgSubmitFraudProof"
	PaymentAPIMsgRegisterRoutingAdvertisement	PaymentAPIMessageName	= "MsgRegisterRoutingAdvertisement"
)

const (
	PaymentAPIQueryChannel			PaymentAPIQueryName	= "QueryChannel"
	PaymentAPIQueryChannelsByParticipant	PaymentAPIQueryName	= "QueryChannelsByParticipant"
	PaymentAPIQueryPendingClose		PaymentAPIQueryName	= "QueryPendingClose"
	PaymentAPIQueryFinalizationHeight	PaymentAPIQueryName	= "QueryFinalizationHeight"
	PaymentAPIQueryCondition		PaymentAPIQueryName	= "QueryCondition"
	PaymentAPIQueryConditionsByChannel	PaymentAPIQueryName	= "QueryConditionsByChannel"
	PaymentAPIQueryVirtualChannel		PaymentAPIQueryName	= "QueryVirtualChannel"
	PaymentAPIQueryChannelCapacity		PaymentAPIQueryName	= "QueryChannelCapacity"
	PaymentAPIQueryFeeSchedule		PaymentAPIQueryName	= "QueryFeeSchedule"
	PaymentAPIQuerySettlementTombstone	PaymentAPIQueryName	= "QuerySettlementTombstone"
	PaymentAPIQueryFraudProof		PaymentAPIQueryName	= "QueryFraudProof"
	PaymentAPIQueryActiveDisputes		PaymentAPIQueryName	= "QueryActiveDisputes"
	PaymentAPIQueryPendingFinalizations	PaymentAPIQueryName	= "QueryPendingFinalizations"
)

const (
	PaymentAPIEventChannelOpened			PaymentAPIEventName	= "channel_opened"
	PaymentAPIEventChannelCheckpointed		PaymentAPIEventName	= "channel_checkpointed"
	PaymentAPIEventChannelCloseStarted		PaymentAPIEventName	= "channel_close_started"
	PaymentAPIEventChannelDisputed			PaymentAPIEventName	= "channel_disputed"
	PaymentAPIEventChannelFinalized			PaymentAPIEventName	= "channel_finalized"
	PaymentAPIEventChannelSettled			PaymentAPIEventName	= "channel_settled"
	PaymentAPIEventChannelPenalized			PaymentAPIEventName	= "channel_penalized"
	PaymentAPIEventPromiseRegistered		PaymentAPIEventName	= "promise_registered"
	PaymentAPIEventPromiseResolved			PaymentAPIEventName	= "promise_resolved"
	PaymentAPIEventPromiseExpired			PaymentAPIEventName	= "promise_expired"
	PaymentAPIEventVirtualChannelOpened		PaymentAPIEventName	= "virtual_channel_opened"
	PaymentAPIEventVirtualChannelClosed		PaymentAPIEventName	= "virtual_channel_closed"
	PaymentAPIEventVirtualChannelDisputed		PaymentAPIEventName	= "virtual_channel_disputed"
	PaymentAPIEventFraudProofAccepted		PaymentAPIEventName	= "fraud_proof_accepted"
	PaymentAPIEventFraudProofRejected		PaymentAPIEventName	= "fraud_proof_rejected"
	PaymentAPIEventRoutingAdvertisementRegistered	PaymentAPIEventName	= "routing_advertisement_registered"
	PaymentAPIEventSettlementFeeCharged		PaymentAPIEventName	= "settlement_fee_charged"
)

type MsgResolvePromise = MsgResolveWithPreimage
type MsgRegisterRoutingAdvertisement = MsgRegisterChannelAdvertisement

type MsgOpenVirtualChannel struct {
	Signer		string
	VirtualChannel	VirtualChannel
	ActivationProof	VirtualActivationProof
}

type MsgCloseVirtualChannel struct {
	Signer			string
	VirtualChannelID	string
	CloseProof		VirtualCloseProof
	CurrentHeight		uint64
}

type MsgDisputeVirtualChannel struct {
	Signer		string
	Proof		VirtualChannelDisputeProof
	CurrentHeight	uint64
}

type MsgSubmitFraudProof struct {
	Signer		string
	Submission	FraudProofSubmission
}

type PaymentAPISurfaceResult struct {
	MsgName			PaymentAPIMessageName
	ChannelResult		PaymentChannelMessageResult
	ConditionalSnapshot	ConditionalPaymentsModuleState
	VirtualChannelID	string
	ClosedVirtual		VirtualChannel
	VirtualReleases		[]VirtualReserveRelease
	FraudSnapshot		FraudProofVerificationState
}

type ChannelCapacity struct {
	ChannelID		string
	TotalCollateral		string
	PendingConditionAmount	string
	ReservedLiquidity	string
	AvailableCapacity	string
	ParticipantBalances	[]Balance
	ParticipantAvailable	[]Balance
	ConditionCount		uint64
	CapacityHash		string
}

type FraudProofQueryResult struct {
	Proof		FraudProof
	Evidence	EvidenceRecord
	Penalty		PenaltyRecord
	Reward		ReporterReward
	Pending		bool
	Canonical	string
}

func RequiredPaymentOnChainMessages() []PaymentAPIMessageName {
	return []PaymentAPIMessageName{
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
	}
}

func RequiredPaymentQueries() []PaymentAPIQueryName {
	return []PaymentAPIQueryName{
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
	}
}

func RequiredPaymentEvents() []PaymentAPIEventName {
	return []PaymentAPIEventName{
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
	}
}

func ValidatePaymentAPISurface() error {
	seenMessages := map[PaymentAPIMessageName]struct{}{}
	for _, name := range RequiredPaymentOnChainMessages() {
		if strings.TrimSpace(string(name)) == "" {
			return errors.New("payments api message name is empty")
		}
		if _, found := seenMessages[name]; found {
			return fmt.Errorf("payments api duplicate message %q", name)
		}
		seenMessages[name] = struct{}{}
	}
	seenQueries := map[PaymentAPIQueryName]struct{}{}
	for _, name := range RequiredPaymentQueries() {
		if strings.TrimSpace(string(name)) == "" {
			return errors.New("payments api query name is empty")
		}
		if _, found := seenQueries[name]; found {
			return fmt.Errorf("payments api duplicate query %q", name)
		}
		seenQueries[name] = struct{}{}
	}
	seenEvents := map[PaymentAPIEventName]struct{}{}
	for _, name := range RequiredPaymentEvents() {
		if strings.TrimSpace(string(name)) == "" {
			return errors.New("payments api event name is empty")
		}
		if _, found := seenEvents[name]; found {
			return fmt.Errorf("payments api duplicate event %q", name)
		}
		seenEvents[name] = struct{}{}
	}
	return nil
}

func ApplyPaymentAPISurfaceMessage(chain PaymentsState, fraud FraudProofVerificationState, msg interface{}) (PaymentsState, FraudProofVerificationState, PaymentAPISurfaceResult, error) {
	chain = chain.Export()
	fraud = fraud.Export()
	if msg == nil {
		return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, errors.New("payments api message is required")
	}
	switch m := msg.(type) {
	case PaymentChannelModuleMessage:
		next, result, err := ApplyPaymentChannelMessage(chain, m)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		apiResult := PaymentAPISurfaceResult{
			MsgName:	paymentAPINameForChannelMessage(m),
			ChannelResult:	result,
		}
		next, err = appendPaymentAPIEvents(next, paymentAPIEventsForChannelMessage(chain, next, m, apiResult)...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		return next, fraud, apiResult, nil
	case ConditionalPaymentMessage:
		next, snapshot, err := ApplyConditionalPaymentMessage(chain, m)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		apiResult := PaymentAPISurfaceResult{
			MsgName:		paymentAPINameForConditionalMessage(m),
			ConditionalSnapshot:	snapshot,
		}
		next, err = appendPaymentAPIEvents(next, paymentAPIEventsForConditionalMessage(chain, next, m)...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		return next, fraud, apiResult, nil
	case MsgOpenVirtualChannel:
		msg := m.Normalize()
		if err := msg.ValidateBasic(); err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		var next PaymentsState
		var err error
		vc := msg.VirtualChannel.Normalize()
		if msg.ActivationProof.VirtualChannel.VirtualChannelID != "" {
			next, err = OpenVirtualChannelWithProof(chain, msg.ActivationProof)
			vc = msg.ActivationProof.VirtualChannel.Normalize()
		} else {
			next, err = OpenVirtualChannel(chain, vc)
		}
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		apiResult := PaymentAPISurfaceResult{MsgName: PaymentAPIMsgOpenVirtualChannel, VirtualChannelID: vc.VirtualChannelID}
		next, err = appendPaymentAPIEvents(next, paymentAPIEventForVirtualChannel(PaymentAPIEventVirtualChannelOpened, vc, msg.Signer, virtualEventHeight(vc, msg.ActivationProof.RouteTimeoutHeight))...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		next, err = appendPaymentAPIEvents(next, settlementFeeChargedEvents(chain, next)...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		return next, fraud, apiResult, nil
	case MsgCloseVirtualChannel:
		msg := m.Normalize()
		if err := msg.ValidateBasic(); err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		if msg.CloseProof.VirtualChannelID != "" {
			next, closed, releases, err := CloseVirtualChannelWithProof(chain, msg.CloseProof, msg.CurrentHeight)
			if err != nil {
				return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
			}
			apiResult := PaymentAPISurfaceResult{MsgName: PaymentAPIMsgCloseVirtualChannel, VirtualChannelID: closed.VirtualChannelID, ClosedVirtual: closed, VirtualReleases: releases}
			next, err = appendPaymentAPIEvents(next, paymentAPIEventForVirtualChannel(PaymentAPIEventVirtualChannelClosed, closed, msg.Signer, msg.CurrentHeight)...)
			if err != nil {
				return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
			}
			return next, fraud, apiResult, nil
		}
		next, closed, err := CloseVirtualChannel(chain, msg.VirtualChannelID, msg.CurrentHeight)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		apiResult := PaymentAPISurfaceResult{MsgName: PaymentAPIMsgCloseVirtualChannel, VirtualChannelID: closed.VirtualChannelID, ClosedVirtual: closed}
		next, err = appendPaymentAPIEvents(next, paymentAPIEventForVirtualChannel(PaymentAPIEventVirtualChannelClosed, closed, msg.Signer, msg.CurrentHeight)...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		return next, fraud, apiResult, nil
	case MsgDisputeVirtualChannel:
		msg := m.Normalize()
		if err := msg.ValidateBasic(); err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		next, err := SubmitVirtualChannelDispute(chain, msg.Proof, msg.CurrentHeight)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		apiResult := PaymentAPISurfaceResult{MsgName: PaymentAPIMsgDisputeVirtualChannel, VirtualChannelID: msg.Proof.VirtualChannelID}
		next, err = appendPaymentAPIEvents(next, paymentAPIEventForVirtualChannel(PaymentAPIEventVirtualChannelDisputed, msg.Proof.LatestState, msg.Signer, msg.CurrentHeight)...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		return next, fraud, apiResult, nil
	case MsgSubmitFraudProof:
		msg := m.Normalize()
		if err := msg.ValidateBasic(); err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		nextChain, nextFraud, err := applyGenericFraudProofSubmission(chain, fraud, msg.Submission)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		apiResult := PaymentAPISurfaceResult{MsgName: PaymentAPIMsgSubmitFraudProof, FraudSnapshot: nextFraud}
		nextChain, err = appendPaymentAPIEvents(nextChain, paymentAPIEventsForFraudProof(chain, nextChain, msg.Submission)...)
		if err != nil {
			return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, err
		}
		return nextChain, nextFraud, apiResult, nil
	default:
		return PaymentsState{}, FraudProofVerificationState{}, PaymentAPISurfaceResult{}, errors.New("payments api message type is unsupported")
	}
}

func QueryChannel(state PaymentsState, channelID string) (ChannelRecord, bool, error) {
	state = state.Export()
	if err := ValidateHash("payments query channel id", channelID); err != nil {
		return ChannelRecord{}, false, err
	}
	channel, found := state.ChannelByID(channelID)
	return channel, found, nil
}

func QueryChannelsByParticipant(state PaymentsState, participant string) ([]ChannelRecord, error) {
	state = state.Export()
	participant = strings.TrimSpace(participant)
	if err := addressing.ValidateUserAddress("payments query participant", participant); err != nil {
		return nil, err
	}
	channels := []ChannelRecord{}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if containsString(channel.Participants, participant) || channel.Payer == participant || channel.Receiver == participant {
			channels = append(channels, channel)
		}
	}
	sort.SliceStable(channels, func(i, j int) bool { return channels[i].ChannelID < channels[j].ChannelID })
	return channels, nil
}

func QueryPendingClose(state PaymentsState, channelID string) (PendingClose, bool, error) {
	channel, found, err := QueryChannel(state, channelID)
	if err != nil || !found {
		return PendingClose{}, found, err
	}
	if channel.PendingClose.State.StateHash == "" {
		return PendingClose{}, false, nil
	}
	return channel.PendingClose.Normalize(), true, nil
}

func QueryFinalizationHeight(state PaymentsState, channelID string) (uint64, bool, error) {
	return state.Export().PendingFinalizationHeight(channelID)
}

func QueryCondition(state PaymentsState, channelID, conditionID string) (ConditionalPayment, bool, error) {
	conditions, err := QueryConditionsByChannel(state, channelID)
	if err != nil {
		return ConditionalPayment{}, false, err
	}
	conditionID = normalizeHash(conditionID)
	if err := ValidateHash("payments query condition id", conditionID); err != nil {
		return ConditionalPayment{}, false, err
	}
	for _, condition := range conditions {
		condition = condition.Normalize()
		if condition.ConditionID == conditionID {
			return condition, true, nil
		}
	}
	return ConditionalPayment{}, false, nil
}

func QueryConditionsByChannel(state PaymentsState, channelID string) ([]ConditionalPayment, error) {
	channel, found, err := QueryChannel(state, channelID)
	if err != nil || !found {
		if err == nil && !found {
			err = errors.New("payments query channel not found")
		}
		return nil, err
	}
	seen := map[string]struct{}{}
	conditions := []ConditionalPayment{}
	for _, condition := range channel.LatestState.Conditions {
		condition = condition.Normalize()
		if _, found := seen[condition.ConditionID]; !found {
			seen[condition.ConditionID] = struct{}{}
			conditions = append(conditions, condition)
		}
	}
	for _, condition := range channel.PendingClose.State.Conditions {
		condition = condition.Normalize()
		if condition.ConditionID == "" {
			continue
		}
		if _, found := seen[condition.ConditionID]; !found {
			seen[condition.ConditionID] = struct{}{}
			conditions = append(conditions, condition)
		}
	}
	sort.SliceStable(conditions, func(i, j int) bool { return conditions[i].ConditionID < conditions[j].ConditionID })
	return conditions, nil
}

func QueryVirtualChannel(state PaymentsState, virtualChannelID string) (VirtualChannel, bool, error) {
	state = state.Export()
	if err := ValidateHash("payments query virtual channel id", virtualChannelID); err != nil {
		return VirtualChannel{}, false, err
	}
	vc, found := state.VirtualChannelByID(virtualChannelID)
	return vc, found, nil
}

func QueryChannelCapacity(state PaymentsState, liquidity LiquidityOptimizationState, channelID string, currentHeight uint64) (ChannelCapacity, error) {
	channel, found, err := QueryChannel(state, channelID)
	if err != nil {
		return ChannelCapacity{}, err
	}
	if !found {
		return ChannelCapacity{}, errors.New("payments query channel not found")
	}
	total, err := parseNonNegativeInt("payments query channel collateral", channel.Collateral)
	if err != nil {
		return ChannelCapacity{}, err
	}
	pending, err := sumConditions(channel.LatestState.Conditions)
	if err != nil {
		return ChannelCapacity{}, err
	}
	reserved, err := activeReservedCapacityForChannel(liquidity, channel.ChannelID, currentHeight)
	if err != nil {
		return ChannelCapacity{}, err
	}
	available := total.Sub(pending).Sub(reserved)
	if available.IsNegative() {
		available = sdkmath.ZeroInt()
	}
	participantAvailable := []Balance{}
	for _, balance := range channel.LatestState.Balances {
		amount, err := parseNonNegativeInt("payments query participant balance", balance.Amount)
		if err != nil {
			return ChannelCapacity{}, err
		}
		share := amount
		if share.GT(available) {
			share = available
		}
		participantAvailable = append(participantAvailable, Balance{Participant: balance.Participant, Amount: share.String()})
	}
	capacity := ChannelCapacity{
		ChannelID:		channel.ChannelID,
		TotalCollateral:	total.String(),
		PendingConditionAmount:	pending.String(),
		ReservedLiquidity:	reserved.String(),
		AvailableCapacity:	available.String(),
		ParticipantBalances:	normalizeBalances(channel.LatestState.Balances),
		ParticipantAvailable:	normalizeBalances(participantAvailable),
		ConditionCount:		uint64(len(channel.LatestState.Conditions)),
	}
	capacity.CapacityHash = HashParts("payments-query-channel-capacity", capacity.ChannelID, capacity.TotalCollateral, capacity.PendingConditionAmount, capacity.ReservedLiquidity, capacity.AvailableCapacity)
	return capacity, nil
}

func QueryFeeSchedule(state PaymentsState) (PaymentFeeSchedule, error) {
	schedule := state.Export().FeeSchedule.Normalize()
	return schedule, schedule.Validate()
}

func QuerySettlementTombstone(state PaymentsState, channelID string) (ClosedChannelTombstone, bool, error) {
	state = state.Export()
	channelID = normalizeHash(channelID)
	if err := ValidateHash("payments query settlement tombstone channel id", channelID); err != nil {
		return ClosedChannelTombstone{}, false, err
	}
	for _, tombstone := range state.ClosedChannels {
		tombstone = tombstone.Normalize()
		if tombstone.ChannelID == channelID {
			return tombstone, true, nil
		}
	}
	return ClosedChannelTombstone{}, false, nil
}

func QueryFraudProof(state PaymentsState, fraud FraudProofVerificationState, proofID string) (FraudProofQueryResult, bool, error) {
	state = state.Export()
	fraud = fraud.Export()
	proofID = normalizeHash(proofID)
	if err := ValidateHash("payments query fraud proof id", proofID); err != nil {
		return FraudProofQueryResult{}, false, err
	}
	result := FraudProofQueryResult{}
	for _, evidence := range fraud.EvidenceRecords {
		evidence = evidence.Normalize()
		if evidence.ProofID == proofID || evidence.EvidenceID == proofID || evidence.CanonicalHash == proofID {
			result.Evidence = evidence
			result.Canonical = evidence.CanonicalHash
			for _, penalty := range fraud.PenaltyRecords {
				penalty = penalty.Normalize()
				if penalty.EvidenceID == evidence.EvidenceID {
					result.Penalty = penalty
					break
				}
			}
			for _, reward := range fraud.ReporterRewards {
				reward = reward.Normalize()
				if reward.EvidenceID == evidence.EvidenceID {
					result.Reward = reward
					break
				}
			}
			return result, true, nil
		}
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		for _, proof := range channel.PendingClose.FraudProofs {
			proof = proof.Normalize()
			if proof.ProofID == proofID {
				result.Proof = proof
				result.Pending = true
				if proof.EvidenceHash != "" {
					result.Canonical = ComputeCanonicalFraudEvidenceHash(channel, proof)
				}
				return result, true, nil
			}
		}
	}
	return FraudProofQueryResult{}, false, nil
}

func QueryActiveDisputes(state PaymentsState, currentHeight uint64) ([]AdaptiveSyncActiveDisputeIndex, error) {
	snapshot, err := BuildAdaptiveSyncSnapshot(state, currentHeight)
	if err != nil {
		return nil, err
	}
	return append([]AdaptiveSyncActiveDisputeIndex(nil), snapshot.ActiveDisputes...), nil
}

func QueryPendingFinalizations(state PaymentsState, currentHeight uint64) ([]AdaptiveSyncPendingFinalizationIndex, error) {
	snapshot, err := BuildAdaptiveSyncSnapshot(state, currentHeight)
	if err != nil {
		return nil, err
	}
	return append([]AdaptiveSyncPendingFinalizationIndex(nil), snapshot.PendingFinalizations...), nil
}

func (m MsgOpenVirtualChannel) Normalize() MsgOpenVirtualChannel {
	m.Signer = strings.TrimSpace(m.Signer)
	m.VirtualChannel = m.VirtualChannel.Normalize()
	m.ActivationProof = m.ActivationProof.Normalize()
	return m
}

func (m MsgOpenVirtualChannel) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg open virtual signer", msg.Signer); err != nil {
		return err
	}
	vc := msg.VirtualChannel
	if msg.ActivationProof.VirtualChannel.VirtualChannelID != "" {
		vc = msg.ActivationProof.VirtualChannel.Normalize()
	}
	if err := ValidateHash("payments msg open virtual id", vc.VirtualChannelID); err != nil {
		return err
	}
	if len(vc.Endpoints) > 0 && !containsString(vc.Endpoints, msg.Signer) && msg.Signer != vc.EndpointA && msg.Signer != vc.EndpointB {
		return errors.New("payments msg open virtual signer must be endpoint")
	}
	return nil
}

func (m MsgCloseVirtualChannel) Normalize() MsgCloseVirtualChannel {
	m.Signer = strings.TrimSpace(m.Signer)
	m.VirtualChannelID = normalizeOptionalHash(m.VirtualChannelID)
	m.CloseProof = m.CloseProof.Normalize()
	if m.VirtualChannelID == "" && m.CloseProof.VirtualChannelID != "" {
		m.VirtualChannelID = m.CloseProof.VirtualChannelID
	}
	if m.CloseProof.SubmittedBy == "" {
		m.CloseProof.SubmittedBy = m.Signer
	}
	return m
}

func (m MsgCloseVirtualChannel) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg close virtual signer", msg.Signer); err != nil {
		return err
	}
	if err := ValidateHash("payments msg close virtual id", msg.VirtualChannelID); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg close virtual height must be positive")
	}
	if msg.CloseProof.VirtualChannelID != "" && msg.CloseProof.SubmittedBy != msg.Signer {
		return errors.New("payments msg close virtual signer mismatch")
	}
	return nil
}

func (m MsgDisputeVirtualChannel) Normalize() MsgDisputeVirtualChannel {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Proof = m.Proof.Normalize()
	if m.Proof.SubmittedBy == "" {
		m.Proof.SubmittedBy = m.Signer
	}
	return m
}

func (m MsgDisputeVirtualChannel) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg dispute virtual signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg dispute virtual height must be positive")
	}
	if msg.Proof.SubmittedBy != msg.Signer {
		return errors.New("payments msg dispute virtual signer mismatch")
	}
	return ValidateHash("payments msg dispute virtual id", msg.Proof.VirtualChannelID)
}

func (m MsgSubmitFraudProof) Normalize() MsgSubmitFraudProof {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Submission = m.Submission.Normalize()
	if m.Submission.Proof.SubmittedBy == "" {
		m.Submission.Proof.SubmittedBy = m.Signer
	}
	return m
}

func (m MsgSubmitFraudProof) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg submit fraud signer", msg.Signer); err != nil {
		return err
	}
	if msg.Submission.Proof.SubmittedBy != msg.Signer {
		return errors.New("payments msg submit fraud signer mismatch")
	}
	return msg.Submission.ValidateBasic()
}

func paymentAPINameForChannelMessage(msg PaymentChannelModuleMessage) PaymentAPIMessageName {
	switch msg.PaymentChannelType() {
	case PaymentChannelMsgOpenChannel:
		return PaymentAPIMsgOpenChannel
	case PaymentChannelMsgCooperativeClose:
		return PaymentAPIMsgCooperativeClose
	case PaymentChannelMsgUnilateralClose:
		return PaymentAPIMsgUnilateralClose
	case PaymentChannelMsgDisputeClose:
		return PaymentAPIMsgDisputeClose
	case PaymentChannelMsgFinalizeClose:
		return PaymentAPIMsgFinalizeClose
	case PaymentChannelMsgSubmitCheckpoint:
		return PaymentAPIMsgSubmitCheckpoint
	case PaymentChannelMsgRegisterChannelAdvertisement:
		return PaymentAPIMsgRegisterRoutingAdvertisement
	default:
		return PaymentAPIMessageName(msg.PaymentChannelType())
	}
}

func paymentAPINameForConditionalMessage(msg ConditionalPaymentMessage) PaymentAPIMessageName {
	switch msg.ConditionalPaymentType() {
	case ConditionalMsgRegisterPromise:
		return PaymentAPIMsgRegisterPromise
	case ConditionalMsgResolveWithPreimage:
		return PaymentAPIMsgResolvePromise
	case ConditionalMsgExpirePromise:
		return PaymentAPIMsgExpirePromise
	case ConditionalMsgBatchResolvePromises:
		return PaymentAPIMsgBatchResolvePromises
	default:
		return PaymentAPIMessageName(msg.ConditionalPaymentType())
	}
}

func applyGenericFraudProofSubmission(chain PaymentsState, fraud FraudProofVerificationState, submission FraudProofSubmission) (PaymentsState, FraudProofVerificationState, error) {
	switch submission.Normalize().Proof.ProofType {
	case FraudProofTypeStaleClose:
		return ApplyFraudProofVerificationMessage(chain, fraud, MsgSubmitStaleCloseProof{Input: submission})
	case FraudProofTypeDoubleSign:
		return ApplyFraudProofVerificationMessage(chain, fraud, MsgSubmitDoubleSignProof{Input: submission})
	case FraudProofTypeInvalidCondition:
		return ApplyFraudProofVerificationMessage(chain, fraud, MsgSubmitInvalidConditionProof{Input: submission})
	case FraudProofTypeReplayAttempt:
		return ApplyFraudProofVerificationMessage(chain, fraud, MsgSubmitReplayProof{Input: submission})
	case FraudProofTypeAsyncOverexposure:
		return ApplyFraudProofVerificationMessage(chain, fraud, MsgSubmitAsyncOverexposureProof{Input: submission})
	default:
		return PaymentsState{}, FraudProofVerificationState{}, fmt.Errorf("payments generic fraud proof unsupported type %q", submission.Proof.ProofType)
	}
}

func activeReservedCapacityForChannel(state LiquidityOptimizationState, channelID string, currentHeight uint64) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	channelID = normalizeHash(channelID)
	for _, reservation := range state.Normalize().Reservations {
		reservation = reservation.Normalize()
		if reservation.ChannelID != channelID || reservation.Released {
			continue
		}
		if currentHeight > 0 && reservation.ExpirationHeight > 0 && reservation.ExpirationHeight < currentHeight {
			continue
		}
		amount, err := parseNonNegativeInt("payments query reserved liquidity", reservation.Capacity)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(amount)
	}
	return total, nil
}

func appendPaymentAPIEvents(state PaymentsState, events ...PaymentEvent) (PaymentsState, error) {
	if len(events) == 0 {
		return state, nil
	}
	next := state.Clone()
	for _, event := range events {
		event = event.Normalize()
		if event.EventID == "" {
			continue
		}
		next.Events = append(next.Events, event)
	}
	return next, next.Validate()
}

func paymentAPIEventsForChannelMessage(before, after PaymentsState, msg PaymentChannelModuleMessage, result PaymentAPISurfaceResult) []PaymentEvent {
	events := []PaymentEvent{}
	switch m := msg.(type) {
	case MsgOpenChannel:
		req := m.Normalize().Request
		events = append(events, newPaymentAPIEvent(PaymentAPIEventChannelOpened, req.ChannelID, req.OpenHeight,
			PaymentEventAttribute{Key: "channel_type", Value: string(req.ChannelType)},
			PaymentEventAttribute{Key: "collateral", Value: req.Collateral},
			PaymentEventAttribute{Key: "denom", Value: req.OpeningFeeDenom},
		))
	case MsgCooperativeClose:
		req := m.Normalize().Request
		events = append(events,
			newPaymentAPIEvent(PaymentAPIEventChannelFinalized, req.ChannelID, req.CurrentHeight, PaymentEventAttribute{Key: "mode", Value: string(CloseReasonCooperative)}),
			newPaymentAPIEvent(PaymentAPIEventChannelSettled, req.ChannelID, req.CurrentHeight, PaymentEventAttribute{Key: "settlement_hash", Value: result.ChannelResult.Settlement.SettlementHash}),
		)
	case MsgUnilateralClose:
		req := m.Normalize().Request
		events = append(events, newPaymentAPIEvent(PaymentAPIEventChannelCloseStarted, req.ChannelID, req.CurrentHeight,
			PaymentEventAttribute{Key: "close_reason", Value: string(req.CloseReason)},
			PaymentEventAttribute{Key: "state_hash", Value: req.ClosingStateWithSignatures().StateHash},
		))
	case MsgDisputeClose:
		req := m.Normalize().Request
		events = append(events, newPaymentAPIEvent(PaymentAPIEventChannelDisputed, req.ChannelID, req.CurrentHeight,
			PaymentEventAttribute{Key: "submitter", Value: req.Submitter},
			PaymentEventAttribute{Key: "state_hash", Value: req.NewerState.StateHash},
		))
	case MsgFinalizeClose:
		req := m.Normalize().Request
		events = append(events,
			newPaymentAPIEvent(PaymentAPIEventChannelFinalized, req.ChannelID, req.CurrentHeight, PaymentEventAttribute{Key: "mode", Value: "unilateral"}),
			newPaymentAPIEvent(PaymentAPIEventChannelSettled, req.ChannelID, req.CurrentHeight, PaymentEventAttribute{Key: "settlement_hash", Value: result.ChannelResult.Settlement.SettlementHash}),
		)
	case MsgSubmitCheckpoint:
		req := m.Normalize().Request
		events = append(events, newPaymentAPIEvent(PaymentAPIEventChannelCheckpointed, req.ChannelID, req.CurrentHeight,
			PaymentEventAttribute{Key: "state_hash", Value: req.State.StateHash},
			PaymentEventAttribute{Key: "nonce", Value: fmt.Sprintf("%d", req.State.Nonce)},
		))
	case MsgRegisterChannelAdvertisement:
		msg := m.Normalize()
		events = append(events, newPaymentAPIEvent(PaymentAPIEventRoutingAdvertisementRegistered, msg.Advertisement.ChannelID, msg.CurrentHeight,
			PaymentEventAttribute{Key: "advertisement_id", Value: msg.Advertisement.AdvertisementID},
			PaymentEventAttribute{Key: "advertiser", Value: msg.Advertisement.Advertiser},
		))
	}
	events = append(events, settlementFeeChargedEvents(before, after)...)
	return events
}

func paymentAPIEventsForConditionalMessage(before, after PaymentsState, msg ConditionalPaymentMessage) []PaymentEvent {
	_ = before
	_ = after
	events := []PaymentEvent{}
	switch m := msg.(type) {
	case MsgRegisterPromise:
		msg := m.Normalize()
		for _, promise := range msg.Promises {
			events = append(events, newPaymentAPIEvent(PaymentAPIEventPromiseRegistered, msg.ChannelID, msg.CurrentHeight,
				PaymentEventAttribute{Key: "promise_id", Value: promise.PromiseID},
				PaymentEventAttribute{Key: "amount", Value: promise.Amount},
			))
		}
	case MsgResolveWithPreimage:
		req := m.Normalize().Request
		for _, promise := range req.Promises {
			events = append(events, newPaymentAPIEvent(PaymentAPIEventPromiseResolved, req.ChannelID, req.CurrentHeight,
				PaymentEventAttribute{Key: "promise_id", Value: promise.PromiseID},
				PaymentEventAttribute{Key: "revealer", Value: req.Revealer},
			))
		}
	case MsgExpirePromise:
		req := m.Normalize().Request
		for _, promise := range req.Promises {
			events = append(events, newPaymentAPIEvent(PaymentAPIEventPromiseExpired, req.ChannelID, req.CurrentHeight,
				PaymentEventAttribute{Key: "promise_id", Value: promise.PromiseID},
				PaymentEventAttribute{Key: "resolver", Value: req.Resolver},
			))
		}
	case MsgBatchResolvePromises:
		req := m.Normalize().Request
		eventName := PaymentAPIEventPromiseResolved
		if req.Mode == ConditionSettlementModeExpiry {
			eventName = PaymentAPIEventPromiseExpired
		}
		for _, promise := range req.LinkageProof.Promises {
			events = append(events, newPaymentAPIEvent(eventName, promise.ChannelID, req.CurrentHeight,
				PaymentEventAttribute{Key: "promise_id", Value: promise.PromiseID},
				PaymentEventAttribute{Key: "route_id", Value: req.LinkageProof.RouteID},
			))
		}
	}
	return events
}

func paymentAPIEventsForFraudProof(before, after PaymentsState, submission FraudProofSubmission) []PaymentEvent {
	submission = submission.Normalize()
	events := []PaymentEvent{
		newPaymentAPIEvent(PaymentAPIEventFraudProofAccepted, submission.ChannelID, submission.CurrentHeight,
			PaymentEventAttribute{Key: "proof_id", Value: submission.Proof.ProofID},
			PaymentEventAttribute{Key: "proof_type", Value: string(submission.Proof.ProofType)},
			PaymentEventAttribute{Key: "submitted_by", Value: submission.Proof.SubmittedBy},
		),
		newPaymentAPIEvent(PaymentAPIEventChannelPenalized, submission.ChannelID, submission.CurrentHeight,
			PaymentEventAttribute{Key: "proof_id", Value: submission.Proof.ProofID},
			PaymentEventAttribute{Key: "offending_signer", Value: submission.Proof.OffendingSigner},
			PaymentEventAttribute{Key: "penalty_amount", Value: submission.Proof.PenaltyAmount},
		),
	}
	events = append(events, settlementFeeChargedEvents(before, after)...)
	return events
}

func paymentAPIEventForVirtualChannel(name PaymentAPIEventName, vc VirtualChannel, signer string, height uint64) []PaymentEvent {
	vc = vc.Normalize()
	channelID := virtualEventChannelID(vc)
	if channelID == "" || height == 0 {
		return nil
	}
	return []PaymentEvent{newPaymentAPIEvent(name, channelID, height,
		PaymentEventAttribute{Key: "virtual_channel_id", Value: vc.VirtualChannelID},
		PaymentEventAttribute{Key: "signer", Value: strings.TrimSpace(signer)},
		PaymentEventAttribute{Key: "capacity", Value: vc.Capacity},
	)}
}

func PaymentAPIFraudProofRejectedEvent(submission FraudProofSubmission, reason string) (PaymentEvent, error) {
	submission = submission.Normalize()
	if err := ValidateHash("payments rejected fraud proof channel", submission.ChannelID); err != nil {
		return PaymentEvent{}, err
	}
	if submission.CurrentHeight == 0 {
		return PaymentEvent{}, errors.New("payments rejected fraud proof height must be positive")
	}
	event := newPaymentAPIEvent(PaymentAPIEventFraudProofRejected, submission.ChannelID, submission.CurrentHeight,
		PaymentEventAttribute{Key: "proof_id", Value: submission.Proof.ProofID},
		PaymentEventAttribute{Key: "proof_type", Value: string(submission.Proof.ProofType)},
		PaymentEventAttribute{Key: "reason_hash", Value: HashParts("fraud-proof-rejected", reason)},
	)
	return event, event.Validate()
}

func settlementFeeChargedEvents(before, after PaymentsState) []PaymentEvent {
	beforeByID := map[string]struct{}{}
	for _, charge := range before.Export().FeeCharges {
		beforeByID[charge.Normalize().FeeID] = struct{}{}
	}
	events := []PaymentEvent{}
	after = after.Export()
	for _, charge := range after.FeeCharges {
		charge = charge.Normalize()
		if _, found := beforeByID[charge.FeeID]; found {
			continue
		}
		channelID := paymentAPIEventChannelID(after, charge.ChannelID)
		events = append(events, newPaymentAPIEvent(PaymentAPIEventSettlementFeeCharged, channelID, charge.Height,
			PaymentEventAttribute{Key: "fee_id", Value: charge.FeeID},
			PaymentEventAttribute{Key: "fee_class", Value: string(charge.FeeClass)},
			PaymentEventAttribute{Key: "payer", Value: charge.Payer},
			PaymentEventAttribute{Key: "amount", Value: charge.Amount},
			PaymentEventAttribute{Key: "required_amount", Value: charge.RequiredAmount},
			PaymentEventAttribute{Key: "object_id", Value: charge.ObjectID},
		))
	}
	return events
}

func newPaymentAPIEvent(name PaymentAPIEventName, channelID string, height uint64, attrs ...PaymentEventAttribute) PaymentEvent {
	channelID = normalizeHash(channelID)
	normalizedAttrs := normalizePaymentEventAttributes(attrs)
	parts := []string{"payment-api-event", string(name), channelID, fmt.Sprintf("%020d", height)}
	for _, attr := range normalizedAttrs {
		parts = append(parts, attr.Key, attr.Value)
	}
	return PaymentEvent{
		EventID:	HashParts(parts...),
		EventType:	string(name),
		ChannelID:	channelID,
		Height:		height,
		Attributes:	normalizedAttrs,
	}.Normalize()
}

func virtualEventChannelID(vc VirtualChannel) string {
	vc = vc.Normalize()
	for _, channelID := range vc.ParentChannelIDs {
		if channelID != "" {
			return channelID
		}
	}
	return ""
}

func virtualEventHeight(vc VirtualChannel, routeTimeoutHeight uint64) uint64 {
	if routeTimeoutHeight != 0 {
		return routeTimeoutHeight
	}
	return vc.Normalize().ExpiresHeight
}

func paymentAPIEventChannelID(state PaymentsState, objectID string) string {
	objectID = normalizeHash(objectID)
	if _, found := state.ChannelByID(objectID); found {
		return objectID
	}
	if vc, found := state.VirtualChannelByID(objectID); found {
		if channelID := virtualEventChannelID(vc); channelID != "" {
			return channelID
		}
	}
	return objectID
}
