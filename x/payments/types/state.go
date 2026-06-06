package types

import (
	"errors"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

type PaymentsState struct {
	Channels        []ChannelRecord
	Edges           []ChannelEdge
	VirtualChannels []VirtualChannel
	Settlements     []SettlementRecord
	Batches         []SettlementBatch
	CustodyLocks    []CustodyLock
	ClosedChannels  []ClosedChannelTombstone
	ConditionClaims []ConditionClaimRecord
	Events          []PaymentEvent
}

func EmptyState() PaymentsState {
	return PaymentsState{
		Channels:        []ChannelRecord{},
		Edges:           []ChannelEdge{},
		VirtualChannels: []VirtualChannel{},
		Settlements:     []SettlementRecord{},
		Batches:         []SettlementBatch{},
		CustodyLocks:    []CustodyLock{},
		ClosedChannels:  []ClosedChannelTombstone{},
		ConditionClaims: []ConditionClaimRecord{},
		Events:          []PaymentEvent{},
	}
}

func setChannelFinality(channel ChannelRecord, finality ChannelFinality, height uint64, events *[]PaymentEvent) (ChannelRecord, error) {
	channel = channel.Normalize()
	if height == 0 {
		return ChannelRecord{}, errors.New("payments finality transition height must be positive")
	}
	if !IsChannelFinality(finality) {
		return ChannelRecord{}, fmt.Errorf("unknown payments channel finality %q", finality)
	}
	previous := channel.Finality
	if previous == finality {
		return channel, nil
	}
	channel.Finality = finality
	if err := validateChannelFinalityForStatus(channel); err != nil {
		return ChannelRecord{}, err
	}
	if events != nil {
		*events = append(*events, ChannelFinalityTransitionEvent(channel, previous, finality, height))
	}
	return channel.Normalize(), nil
}

func finalityForPendingClose(channel ChannelRecord) ChannelFinality {
	return DerivedChannelFinality(channel)
}

func finalityForSettledChannel(channel ChannelRecord) ChannelFinality {
	channel = channel.Normalize()
	if len(channel.PendingClose.Penalties) > 0 || len(channel.PendingClose.PenaltyAllocations) > 0 {
		return ChannelFinalityPenalized
	}
	return ChannelFinalitySettled
}

func OpenChannelFromRequest(state PaymentsState, req ChannelOpenRequest) (PaymentsState, PaymentEvent, error) {
	channel, err := BuildChannelFromOpenRequest(req)
	if err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	next, event, err := openChannelRecord(state, channel)
	return next, event, err
}

func OpenChannel(state PaymentsState, channel ChannelRecord) (PaymentsState, error) {
	next, _, err := openChannelRecord(state, channel)
	return next, err
}

func openChannelRecord(state PaymentsState, channel ChannelRecord) (PaymentsState, PaymentEvent, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	channel = channel.Normalize()
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments new channel must start open")
	}
	if _, found := state.ChannelByID(channel.ChannelID); found {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments channel already exists")
	}
	if err := (SettlementArbitrationInput{
		Operation: SettlementArbitrationOpen,
		ChannelID: channel.ChannelID,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	if err := channel.LatestState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	if channel.OpeningStateHash == "" {
		channel.OpeningStateHash = channel.LatestState.StateHash
	}
	if channel.OpeningStateHash != channel.LatestState.StateHash {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments opening state hash mismatch")
	}
	channel.FinalizedNonce = 0
	channel.Finality = ChannelFinalityOpen
	if err := channel.Validate(); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	lock := CustodyLock{ChannelID: channel.ChannelID, Denom: NativeDenom, Amount: channel.Collateral}.Normalize()
	if err := lock.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, PaymentEvent{}, err
	}
	if _, found := state.CustodyLockByChannel(channel.ChannelID); found {
		return PaymentsState{}, PaymentEvent{}, errors.New("payments custody lock already exists")
	}
	event := ChannelOpenEvent(channel)
	next := state.Clone()
	next.Channels = append(next.Channels, channel)
	next.CustodyLocks = append(next.CustodyLocks, lock)
	next.Events = append(next.Events, event)
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(channel, "", ChannelFinalityOpen, channel.OpenHeight))
	sortChannels(next.Channels)
	sortCustodyLocks(next.CustodyLocks)
	return next, event, next.Validate()
}

func RegisterRoutingEdge(state PaymentsState, edge ChannelEdge) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	edge = edge.Normalize()
	if err := edge.Validate(); err != nil {
		return PaymentsState{}, err
	}
	channel, found := state.ChannelByID(edge.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments routing edge requires open channel")
	}
	if !containsString(channel.Participants, edge.From) || !containsString(channel.Participants, edge.To) {
		return PaymentsState{}, errors.New("payments routing edge endpoints must be channel participants")
	}
	if _, found := state.EdgeByKey(edge.ChannelID, edge.From, edge.To); found {
		return PaymentsState{}, errors.New("payments routing edge already exists")
	}
	next := state.Clone()
	next.Edges = append(next.Edges, edge)
	sortEdges(next.Edges)
	return next, next.Validate()
}

func AcceptSignedState(state PaymentsState, channelID string, nextState ChannelState, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments state update height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	nextState = nextState.Normalize()
	if err := nextState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, err
	}
	if nextState.Nonce <= channel.LatestState.Nonce {
		return PaymentsState{}, errors.New("payments channel state nonce must strictly increase")
	}
	if err := ValidatePreviousHashContinuity(channel, nextState); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.LatestState = nextState
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func AcceptAsyncCheckpoint(state PaymentsState, channelID string, checkpoint ChannelState, deltas []AsyncPaymentDelta, submitter string, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments async checkpoint height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	if channel.ChannelType != ChannelTypeAsync {
		return PaymentsState{}, errors.New("payments checkpoint requires async channel")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, errors.New("payments async checkpoint submitter must be participant")
	}
	checkpoint = checkpoint.Normalize()
	if err := checkpoint.ValidateForChannel(channel, false); err != nil {
		return PaymentsState{}, err
	}
	if checkpoint.CheckpointNonce <= channel.LatestState.CheckpointNonce {
		return PaymentsState{}, errors.New("payments async checkpoint nonce must increase")
	}
	proof := AsyncDeltaDisputeProof{
		ProofID:         HashParts("async-checkpoint-proof", checkpoint.StateHash),
		ChannelID:       channel.ChannelID,
		CheckpointState: checkpoint,
		Deltas:          deltas,
		EvidenceHash:    HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRootForChannel(channel, deltas)),
	}
	if err := proof.ValidateForChannel(channel, currentHeight); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.LatestState = checkpoint
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func RegisterUpdateCheckpoint(state PaymentsState, req ChannelUpdateRequest) (PaymentsState, ChannelUpdateResult, error) {
	state = state.Export()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, ChannelUpdateResult{}, errors.New("payments channel not found")
	}
	result, err := ValidateOffchainUpdate(channel, req)
	if err != nil {
		return PaymentsState{}, ChannelUpdateResult{}, err
	}
	if !req.Normalize().RegisterCheckpoint {
		return state, result, nil
	}
	var next PaymentsState
	if channel.ChannelType == ChannelTypeAsync || len(req.Normalize().AsyncDeltas) > 0 {
		next, err = AcceptAsyncCheckpoint(state, channel.ChannelID, req.Normalize().State, req.Normalize().AsyncDeltas, req.Normalize().Submitter, req.Normalize().CurrentHeight)
	} else {
		next, err = AcceptSignedState(state, channel.ChannelID, req.Normalize().State, req.Normalize().CurrentHeight)
	}
	if err != nil {
		return PaymentsState{}, ChannelUpdateResult{}, err
	}
	result.CheckpointRegistered = true
	return next, result, nil
}

func RevealPromisePreimage(state PaymentsState, req PreimageRevealRequest) (PaymentsState, []ConditionResolution, error) {
	state = state.Export()
	req = req.Normalize()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, nil, errors.New("payments channel not found")
	}
	if err := req.ValidateForChannel(channel, state.ConditionClaims); err != nil {
		return PaymentsState{}, nil, err
	}
	preimageHash := HashParts(req.Preimage)
	resolutions := make([]ConditionResolution, 0, len(req.Promises))
	next := state.Clone()
	for _, promise := range normalizeConditionalPromises(req.Promises) {
		evidenceHash := HashParts("promise-preimage", promise.PromiseID, preimageHash)
		resolution := ConditionResolution{
			ConditionID:  promise.PromiseID,
			Resolver:     req.Revealer,
			Recipient:    promise.Destination,
			Amount:       promise.Amount,
			Expired:      false,
			EvidenceHash: evidenceHash,
		}.Normalize()
		resolutions = append(resolutions, resolution)
		next.ConditionClaims = append(next.ConditionClaims, ConditionClaimRecord{
			ChainID:        channel.ChainID,
			ChannelID:      channel.ChannelID,
			ConditionID:    promise.PromiseID,
			EvidenceHash:   evidenceHash,
			PreimageHash:   preimageHash,
			ResolvedHeight: req.CurrentHeight,
			ExpiresHeight:  req.CurrentHeight + DefaultReplayHorizon,
		}.Normalize())
	}
	sortConditionClaimRecords(next.ConditionClaims)
	return next, normalizeConditionResolutions(resolutions), next.Validate()
}

func ExpireConditionalPromises(state PaymentsState, req PromiseExpiryRequest) (PaymentsState, []ConditionResolution, ConditionRootUpdate, error) {
	state = state.Export()
	req = req.Normalize()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, nil, ConditionRootUpdate{}, errors.New("payments channel not found")
	}
	if err := req.ValidateForChannel(channel, state.ConditionClaims); err != nil {
		return PaymentsState{}, nil, ConditionRootUpdate{}, err
	}
	_, update, err := BuildConditionRootAfterExpiry(channel.LatestState, req.Promises)
	if err != nil {
		return PaymentsState{}, nil, ConditionRootUpdate{}, err
	}
	resolutions := make([]ConditionResolution, 0, len(req.Promises))
	next := state.Clone()
	for _, promise := range normalizeConditionalPromises(req.Promises) {
		evidenceHash := HashParts("promise-expiry", promise.PromiseID, fmt.Sprintf("%020d", req.CurrentHeight))
		resolution := ConditionResolution{
			ConditionID:  promise.PromiseID,
			Resolver:     req.Resolver,
			Recipient:    promise.Source,
			Amount:       promise.Amount,
			Expired:      true,
			EvidenceHash: evidenceHash,
		}.Normalize()
		resolutions = append(resolutions, resolution)
		next.ConditionClaims = append(next.ConditionClaims, ConditionClaimRecord{
			ChainID:        channel.ChainID,
			ChannelID:      channel.ChannelID,
			ConditionID:    promise.PromiseID,
			EvidenceHash:   evidenceHash,
			ResolvedHeight: req.CurrentHeight,
			ExpiresHeight:  req.CurrentHeight + DefaultReplayHorizon,
		}.Normalize())
	}
	sortConditionClaimRecords(next.ConditionClaims)
	return next, normalizeConditionResolutions(resolutions), update, next.Validate()
}

func SubmitClose(state PaymentsState, channelID string, closingState ChannelState, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, error) {
	return SubmitCloseWithRequest(state, ChannelCloseRequest{
		ChannelID:     channelID,
		ClosingState:  closingState,
		CloseReason:   CloseReasonUnilateral,
		Submitter:     submitter,
		CurrentHeight: currentHeight,
		SettlementFee: settlementFee,
	})
}

func SubmitCloseWithRequest(state PaymentsState, req ChannelCloseRequest) (PaymentsState, error) {
	state = state.Export()
	req = req.Normalize()
	if req.CurrentHeight == 0 {
		return PaymentsState{}, errors.New("payments close height must be positive")
	}
	index, channel, found := state.ChannelIndex(req.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	closingState := req.ClosingStateWithSignatures()
	if err := req.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	pending := PendingClose{
		Submitter:          req.Submitter,
		SubmittedHeight:    req.CurrentHeight,
		SettleAfterHeight:  req.CurrentHeight + channel.DisputePeriod,
		CloseReason:        req.CloseReason,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      req.SettlementFee,
		State:              closingState,
	}
	if err := (SettlementArbitrationInput{
		Operation:     SettlementArbitrationUnilateralClose,
		ChannelID:     channel.ChannelID,
		SignedState:   pending.State,
		CurrentHeight: req.CurrentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if err := pending.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if pending.State.Nonce < channel.FinalizedNonce {
		return PaymentsState{}, errors.New("payments close state nonce is below finalized nonce")
	}
	if pending.State.Nonce < channel.LatestState.Nonce {
		return PaymentsState{}, errors.New("payments close state nonce is below latest accepted nonce")
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusPendingClose
	nextChannel.PendingClose = pending
	nextChannel.LatestState = pending.State
	if nextChannel.DisputedNonce < pending.State.Nonce {
		nextChannel.DisputedNonce = pending.State.Nonce
	}
	next := state.Clone()
	var err error
	nextChannel, err = setChannelFinality(nextChannel, finalityForPendingClose(nextChannel), req.CurrentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	sortChannels(next.Channels)
	return next, next.Validate()
}

func ForcedClose(state PaymentsState, channelID string, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments forced close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, errors.New("payments forced close submitter must be participant")
	}
	timeoutHeight := channel.LatestState.TimeoutHeight
	if channel.ChannelType == ChannelTypeAsync && channel.LatestState.ExpiryHeight != 0 {
		timeoutHeight = channel.LatestState.ExpiryHeight
	}
	if channel.ChannelType == ChannelTypeUnidirectional && channel.ExpirationHeight != 0 {
		timeoutHeight = channel.ExpirationHeight
	}
	if timeoutHeight == 0 || currentHeight <= timeoutHeight {
		return PaymentsState{}, errors.New("payments forced close timeout has not expired")
	}
	pending := PendingClose{
		Submitter:          submitter,
		SubmittedHeight:    currentHeight,
		SettleAfterHeight:  currentHeight + channel.DisputePeriod,
		CloseReason:        CloseReasonTimeout,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      settlementFee,
		State:              channel.LatestState.Normalize(),
	}
	if err := pending.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusPendingClose
	nextChannel.PendingClose = pending
	if nextChannel.DisputedNonce < pending.State.Nonce {
		nextChannel.DisputedNonce = pending.State.Nonce
	}
	next := state.Clone()
	var err error
	nextChannel, err = setChannelFinality(nextChannel, finalityForPendingClose(nextChannel), currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	sortChannels(next.Channels)
	return next, next.Validate()
}

func CooperativeClose(state PaymentsState, channelID string, closingState ChannelState, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not open")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close submitter must be participant")
	}
	closingState = closingState.Normalize()
	if err := (SettlementArbitrationInput{
		Operation:     SettlementArbitrationCooperativeClose,
		ChannelID:     channel.ChannelID,
		SignedState:   closingState,
		CurrentHeight: currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := closingState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if closingState.Nonce < channel.LatestState.Nonce {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close state nonce is below latest accepted nonce")
	}
	finalBalances, err := applySettlementAdjustments(closingState.Balances, nil, nil, settlementFee, submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		StateHash:          closingState.StateHash,
		Nonce:              closingState.Nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      settlementFee,
		SettledHeight:      currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.LatestState = closingState
	nextChannel.PendingClose = PendingClose{}
	next := state.Clone()
	nextChannel.Finality = ChannelFinalitySettled
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, channel.Finality, ChannelFinalitySettled, currentHeight))
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, nil, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	return next, settlement, next.Validate()
}

func ReceiverClose(state PaymentsState, channelID string, claim UnidirectionalClaim, receiver string, currentHeight uint64, settlementFee string) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not open")
	}
	if channel.ChannelType != ChannelTypeUnidirectional {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close requires unidirectional channel")
	}
	receiver = normalizeAddress(receiver)
	if receiver != channel.Receiver {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close submitter must be receiver")
	}
	claim = claim.Normalize()
	if err := (SettlementArbitrationInput{
		Operation:     SettlementArbitrationUnilateralClose,
		ChannelID:     channel.ChannelID,
		Claim:         claim,
		CurrentHeight: currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := claim.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := validateUnidirectionalClaimProgress(channel.LatestClaim, claim); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if currentHeight > claim.ExpirationHeight+channel.DisputePeriod {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments receiver close claim has expired")
	}
	finalBalances, err := finalBalancesForUnidirectionalClaim(channel, claim, settlementFee, receiver)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		StateHash:          claim.StateHash,
		Nonce:              claim.Nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      settlementFee,
		SettledHeight:      currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.LatestClaim = claim
	nextChannel.PendingClose = PendingClose{}
	next := state.Clone()
	nextChannel.Finality = ChannelFinalitySettled
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, channel.Finality, ChannelFinalitySettled, currentHeight))
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, nil, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	return next, settlement, next.Validate()
}

func PayerReclaim(state PaymentsState, channelID string, payer string, currentHeight uint64, settlementFee string) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments payer reclaim height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not open")
	}
	if channel.ChannelType != ChannelTypeUnidirectional {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments payer reclaim requires unidirectional channel")
	}
	payer = normalizeAddress(payer)
	if payer != channel.Payer {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments reclaim submitter must be payer")
	}
	expirationHeight := channel.ExpirationHeight
	claim := channel.LatestClaim.Normalize()
	if !claim.IsZero() {
		expirationHeight = claim.ExpirationHeight
	}
	if currentHeight <= expirationHeight+channel.DisputePeriod {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments reclaim is still in dispute window")
	}
	stateHash := channel.OpeningStateHash
	nonce := channel.LatestState.Nonce
	var finalBalances []Balance
	var err error
	if claim.IsZero() {
		finalBalances, err = applySettlementAdjustments([]Balance{
			{Participant: channel.Payer, Amount: channel.Collateral},
			{Participant: channel.Receiver, Amount: "0"},
		}, nil, nil, settlementFee, payer)
	} else {
		stateHash = claim.StateHash
		nonce = claim.Nonce
		finalBalances, err = finalBalancesForUnidirectionalClaim(channel, claim, settlementFee, payer)
	}
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		StateHash:          stateHash,
		Nonce:              nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      settlementFee,
		SettledHeight:      currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	next := state.Clone()
	var transitionErr error
	nextChannel, transitionErr = setChannelFinality(nextChannel, ChannelFinalityExpired, currentHeight, &next.Events)
	if transitionErr != nil {
		return PaymentsState{}, SettlementRecord{}, transitionErr
	}
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.PendingClose = PendingClose{}
	nextChannel.Finality = ChannelFinalitySettled
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, ChannelFinalityExpired, ChannelFinalitySettled, currentHeight))
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, nil, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	return next, settlement, next.Validate()
}

func DisputeClose(state PaymentsState, channelID string, newerState ChannelState, submitter string, currentHeight uint64) (PaymentsState, error) {
	channel, found := state.Export().ChannelByID(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	return DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:             channelID,
		ClosingStateReference: channel.PendingClose.State.StateHash,
		NewerState:            newerState,
		Submitter:             submitter,
		CurrentHeight:         currentHeight,
	})
}

func DisputeChannel(state PaymentsState, req ChannelDisputeRequest) (PaymentsState, error) {
	state = state.Export()
	req = req.Normalize()
	if req.CurrentHeight == 0 {
		return PaymentsState{}, errors.New("payments dispute height must be positive")
	}
	index, channel, found := state.ChannelIndex(req.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, errors.New("payments channel is not pending close")
	}
	if req.ClosingStateReference != channel.PendingClose.State.StateHash {
		return PaymentsState{}, errors.New("payments dispute closing state reference mismatch")
	}
	if req.CurrentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, errors.New("payments dispute window has closed")
	}
	if err := (SettlementArbitrationInput{
		Operation:       SettlementArbitrationDispute,
		ChannelID:       channel.ChannelID,
		SignedState:     req.NewerState,
		ConditionProofs: req.ConditionProofs,
		CurrentHeight:   req.CurrentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if err := req.NewerState.ValidateForChannel(channel, false); err != nil {
		return PaymentsState{}, err
	}
	if !stateStrongerThan(req.NewerState, channel.PendingClose.State) {
		return PaymentsState{}, errors.New("payments dispute state must be newer or stronger")
	}
	if !containsString(channel.Participants, req.Submitter) {
		return PaymentsState{}, errors.New("payments dispute submitter must be participant")
	}
	if err := rejectReusedConditionClaims(state, channel, req.ConditionProofs); err != nil {
		return PaymentsState{}, err
	}
	if err := validateConditionResolutionsForState(req.NewerState, channel, req.ConditionProofs, false); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.PendingClose.State = req.NewerState
	nextChannel.PendingClose.SubmittedHeight = req.CurrentHeight
	if nextChannel.PendingClose.DisputeCount < MaxDisputeExtensions {
		nextChannel.PendingClose.SettleAfterHeight = req.CurrentHeight + channel.DisputePeriod
		nextChannel.PendingClose.DisputeCount++
	}
	nextChannel.PendingClose.ConditionProofs = mergeConditionResolutions(nextChannel.PendingClose.ConditionProofs, req.ConditionProofs)
	if nextChannel.DisputedNonce < req.NewerState.Nonce {
		nextChannel.DisputedNonce = req.NewerState.Nonce
	}
	if req.FraudProof.ProofID != "" {
		if err := req.FraudProof.ValidateForChannel(channel); err != nil {
			return PaymentsState{}, err
		}
		penalties, allocations, err := BuildFraudPenaltyRouting(channel, req.FraudProof, FraudPenaltyPolicy{})
		if err != nil {
			return PaymentsState{}, err
		}
		nextChannel.PendingClose.FraudProofs = append(nextChannel.PendingClose.FraudProofs, req.FraudProof)
		nextChannel.PendingClose.Penalties = append(nextChannel.PendingClose.Penalties, penalties...)
		nextChannel.PendingClose.PenaltyAllocations = append(nextChannel.PendingClose.PenaltyAllocations, allocations...)
	}
	nextChannel.LatestState = req.NewerState
	next := state.Clone()
	var err error
	nextChannel, err = setChannelFinality(nextChannel, finalityForPendingClose(nextChannel), req.CurrentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	next.Events = append(next.Events, ChannelDisputeEvent(nextChannel, req.Submitter, req.CurrentHeight))
	sortChannels(next.Channels)
	return next, next.Validate()
}

func SubmitWatchDispute(state PaymentsState, submission WatchDisputeSubmission) (PaymentsState, error) {
	state = state.Export()
	submission = submission.Normalize()
	channel, found := state.ChannelByID(submission.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if err := submission.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	return DisputeChannel(state, ChannelDisputeRequest{
		ChannelID:             submission.ChannelID,
		ClosingStateReference: submission.ClosingStateReference,
		NewerState:            submission.NewerState,
		Submitter:             submission.Delegator,
		CurrentHeight:         submission.CurrentHeight,
	})
}

func SubmitFraudProof(state PaymentsState, channelID string, proof FraudProof, currentHeight uint64) (PaymentsState, error) {
	return SubmitFraudProofWithPolicy(state, channelID, proof, currentHeight, FraudPenaltyPolicy{})
}

func SubmitFraudProofWithPolicy(state PaymentsState, channelID string, proof FraudProof, currentHeight uint64, policy FraudPenaltyPolicy) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments fraud proof height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, errors.New("payments fraud proof requires pending close")
	}
	if currentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, errors.New("payments fraud proof window has closed")
	}
	proof = proof.Normalize()
	if err := (SettlementArbitrationInput{
		Operation:     SettlementArbitrationFraudProof,
		ChannelID:     channel.ChannelID,
		FraudProof:    proof,
		CurrentHeight: currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	if err := proof.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	for _, existing := range channel.PendingClose.FraudProofs {
		if existing.ProofID == proof.ProofID {
			return PaymentsState{}, errors.New("payments duplicate fraud proof")
		}
	}
	penalties, allocations, err := BuildFraudPenaltyRouting(channel, proof, policy)
	if err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.PendingClose.FraudProofs = append(nextChannel.PendingClose.FraudProofs, proof)
	nextChannel.PendingClose.Penalties = append(nextChannel.PendingClose.Penalties, penalties...)
	nextChannel.PendingClose.PenaltyAllocations = append(nextChannel.PendingClose.PenaltyAllocations, allocations...)
	next := state.Clone()
	nextChannel, err = setChannelFinality(nextChannel, ChannelFinalityPenalized, currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func FraudClose(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments fraud close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments fraud close requires pending close")
	}
	if len(channel.PendingClose.FraudProofs) == 0 || len(channel.PendingClose.Penalties) == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments fraud close requires accepted proof")
	}
	if err := (SettlementArbitrationInput{
		Operation:       SettlementArbitrationFinalSettlement,
		ChannelID:       channel.ChannelID,
		SignedState:     channel.PendingClose.State,
		ConditionProofs: channel.PendingClose.ConditionProofs,
		CurrentHeight:   currentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := rejectReusedConditionClaims(state, channel, channel.PendingClose.ConditionProofs); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	finalBalances, err := applySettlementAdjustments(channel.PendingClose.State.Balances, channel.PendingClose.Penalties, channel.PendingClose.PenaltyAllocations, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		StateHash:          channel.PendingClose.State.StateHash,
		Nonce:              channel.PendingClose.State.Nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: channel.PendingClose.SettlementFeeDenom,
		SettlementFee:      channel.PendingClose.SettlementFee,
		Penalties:          channel.PendingClose.Penalties,
		PenaltyAllocations: channel.PendingClose.PenaltyAllocations,
		SettledHeight:      currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	next := state.Clone()
	nextChannel, err = setChannelFinality(nextChannel, ChannelFinalityFinalizable, currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	settledFinality := finalityForSettledChannel(nextChannel)
	nextChannel.Finality = settledFinality
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, ChannelFinalityFinalizable, settledFinality, currentHeight))
	nextChannel.PendingClose = PendingClose{}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, channel.PendingClose.ConditionProofs, currentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	sortConditionClaimRecords(next.ConditionClaims)
	return next, settlement, next.Validate()
}

func FinalizeSettlement(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, SettlementRecord, error) {
	return FinalizeSettlementWithRequest(state, FinalSettlementRequest{ChannelID: channelID, CurrentHeight: currentHeight})
}

func FinalizeSettlementWithRequest(state PaymentsState, req FinalSettlementRequest) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	req = req.Normalize()
	if req.CurrentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement height must be positive")
	}
	index, channel, found := state.ChannelIndex(req.ChannelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not pending close")
	}
	if req.CurrentHeight < channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement is still in dispute window")
	}
	resolutions := mergeConditionResolutions(channel.PendingClose.ConditionProofs, req.ResolvedConditions)
	if err := (SettlementArbitrationInput{
		Operation:       SettlementArbitrationFinalSettlement,
		ChannelID:       channel.ChannelID,
		SignedState:     channel.PendingClose.State,
		ConditionProofs: resolutions,
		CurrentHeight:   req.CurrentHeight,
	}).ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := rejectReusedConditionClaims(state, channel, resolutions); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if err := validateConditionResolutionsForState(channel.PendingClose.State, channel, resolutions, true); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	baseBalances, err := settlementBalancesWithConditions(channel.PendingClose.State, channel, resolutions)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	finalBalances, err := applySettlementAdjustments(baseBalances, channel.PendingClose.Penalties, channel.PendingClose.PenaltyAllocations, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		StateHash:          channel.PendingClose.State.StateHash,
		Nonce:              channel.PendingClose.State.Nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: channel.PendingClose.SettlementFeeDenom,
		SettlementFee:      channel.PendingClose.SettlementFee,
		Penalties:          channel.PendingClose.Penalties,
		PenaltyAllocations: channel.PendingClose.PenaltyAllocations,
		SettledHeight:      req.CurrentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	next := state.Clone()
	nextChannel, err = setChannelFinality(nextChannel, ChannelFinalityFinalizable, req.CurrentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	settledFinality := finalityForSettledChannel(nextChannel)
	nextChannel.Finality = settledFinality
	next.Events = append(next.Events, ChannelFinalityTransitionEvent(nextChannel, ChannelFinalityFinalizable, settledFinality, req.CurrentHeight))
	nextChannel.PendingClose = PendingClose{}
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.CustodyLocks = filterCustodyLocksForSettledChannel(next.CustodyLocks, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	appendSettlementReplayRecords(&next, nextChannel, settlement, resolutions, req.CurrentHeight)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	sortClosedChannelTombstones(next.ClosedChannels)
	sortConditionClaimRecords(next.ConditionClaims)
	return next, settlement, next.Validate()
}

func OpenVirtualChannel(state PaymentsState, vc VirtualChannel) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	vc = vc.Normalize()
	if _, found := state.VirtualChannelByID(vc.VirtualChannelID); found {
		return PaymentsState{}, errors.New("payments virtual channel already exists")
	}
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return PaymentsState{}, err
	}
	var parentChainID string
	for _, parentID := range vc.ParentChannelIDs {
		channel, found := state.ChannelByID(parentID)
		if !found || channel.Status != ChannelStatusOpen {
			return PaymentsState{}, errors.New("payments virtual channel requires open parents")
		}
		if parentChainID == "" {
			parentChainID = channel.ChainID
		} else if parentChainID != channel.ChainID {
			return PaymentsState{}, errors.New("payments virtual channel parents must share chain id")
		}
		if !containsString(channel.Participants, vc.Endpoints[0]) && !containsString(channel.Participants, vc.Endpoints[1]) {
			return PaymentsState{}, errors.New("payments virtual channel parent path must touch an endpoint")
		}
		if channelCapacity, err := parsePositiveInt("payments channel collateral", channel.Collateral); err != nil {
			return PaymentsState{}, err
		} else if channelCapacity.LT(capacity) {
			return PaymentsState{}, errors.New("payments virtual channel capacity exceeds parent capacity")
		}
	}
	if vc.ChainID == "" {
		vc.ChainID = parentChainID
		vc.AnchorCommitment = ""
		vc.StateHash = ""
	}
	if vc.ChainID != parentChainID {
		return PaymentsState{}, errors.New("payments virtual channel chain id mismatch")
	}
	if vc.AnchorCommitment == "" {
		vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	}
	if vc.StateHash == "" {
		vc.StateHash = ComputeVirtualChannelStateHash(vc)
	}
	if err := vc.Validate(); err != nil {
		return PaymentsState{}, err
	}
	next := state.Clone()
	next.VirtualChannels = append(next.VirtualChannels, vc)
	sortVirtualChannels(next.VirtualChannels)
	return next, next.Validate()
}

func AddSettlementBatch(state PaymentsState, batch SettlementBatch) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	batch = batch.Normalize()
	if err := batch.Validate(); err != nil {
		return PaymentsState{}, err
	}
	for _, existing := range state.Batches {
		if existing.BatchID == batch.BatchID {
			return PaymentsState{}, errors.New("payments settlement batch already exists")
		}
	}
	for _, op := range batch.Operations {
		channel, found := state.ChannelByID(op.ChannelID)
		if !found {
			return PaymentsState{}, errors.New("payments settlement batch references unknown channel")
		}
		if op.Nonce < channel.FinalizedNonce {
			return PaymentsState{}, errors.New("payments settlement batch operation nonce below finalized nonce")
		}
	}
	next := state.Clone()
	next.Batches = append(next.Batches, batch)
	sortBatches(next.Batches)
	return next, next.Validate()
}

func RoutePayment(state PaymentsState, from, to, amountText string, currentHeight uint64, maxHops int) ([]ChannelEdge, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return nil, err
	}
	amount, err := parsePositiveInt("payments route amount", amountText)
	if err != nil {
		return nil, err
	}
	if maxHops <= 0 || maxHops > MaxRoutingHops {
		maxHops = MaxRoutingHops
	}
	candidates := activeEdgesForAmount(state.Edges, amount, currentHeight)
	sortEdges(candidates)
	type path struct {
		node  string
		edges []ChannelEdge
	}
	queue := []path{{node: from}}
	visitedDepth := map[string]int{from: 0}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if len(current.edges) >= maxHops {
			continue
		}
		for _, edge := range candidates {
			if edge.From != current.node {
				continue
			}
			nextEdges := append([]ChannelEdge(nil), current.edges...)
			nextEdges = append(nextEdges, edge)
			if edge.To == to {
				return nextEdges, nil
			}
			if depth, seen := visitedDepth[edge.To]; seen && depth <= len(nextEdges) {
				continue
			}
			visitedDepth[edge.To] = len(nextEdges)
			queue = append(queue, path{node: edge.To, edges: nextEdges})
		}
	}
	return nil, errors.New("payments route not found")
}

func ImportState(state PaymentsState) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	return state, nil
}

func (s PaymentsState) Export() PaymentsState {
	out := s.Clone()
	sortChannels(out.Channels)
	sortEdges(out.Edges)
	sortVirtualChannels(out.VirtualChannels)
	sortSettlements(out.Settlements)
	sortBatches(out.Batches)
	sortCustodyLocks(out.CustodyLocks)
	sortClosedChannelTombstones(out.ClosedChannels)
	sortConditionClaimRecords(out.ConditionClaims)
	return out
}

func (s PaymentsState) Clone() PaymentsState {
	out := PaymentsState{
		Channels:        make([]ChannelRecord, len(s.Channels)),
		Edges:           make([]ChannelEdge, len(s.Edges)),
		VirtualChannels: make([]VirtualChannel, len(s.VirtualChannels)),
		Settlements:     make([]SettlementRecord, len(s.Settlements)),
		Batches:         make([]SettlementBatch, len(s.Batches)),
		CustodyLocks:    make([]CustodyLock, len(s.CustodyLocks)),
		ClosedChannels:  make([]ClosedChannelTombstone, len(s.ClosedChannels)),
		ConditionClaims: make([]ConditionClaimRecord, len(s.ConditionClaims)),
		Events:          make([]PaymentEvent, len(s.Events)),
	}
	for i, channel := range s.Channels {
		out.Channels[i] = channel.Normalize()
	}
	for i, edge := range s.Edges {
		out.Edges[i] = edge.Normalize()
	}
	for i, vc := range s.VirtualChannels {
		out.VirtualChannels[i] = vc.Normalize()
	}
	for i, settlement := range s.Settlements {
		out.Settlements[i] = settlement.Normalize()
	}
	for i, batch := range s.Batches {
		out.Batches[i] = batch.Normalize()
	}
	for i, lock := range s.CustodyLocks {
		out.CustodyLocks[i] = lock.Normalize()
	}
	for i, tombstone := range s.ClosedChannels {
		out.ClosedChannels[i] = tombstone.Normalize()
	}
	for i, claim := range s.ConditionClaims {
		out.ConditionClaims[i] = claim.Normalize()
	}
	for i, event := range s.Events {
		out.Events[i] = event.Normalize()
	}
	return out
}

func (s PaymentsState) Validate() error {
	if err := validateChannels(s.Channels); err != nil {
		return err
	}
	if err := validateEdges(s.Channels, s.Edges); err != nil {
		return err
	}
	if err := validateVirtualChannels(s.Channels, s.VirtualChannels); err != nil {
		return err
	}
	if err := validateSettlements(s.Channels, s.Settlements); err != nil {
		return err
	}
	if err := validateBatches(s.Channels, s.Batches); err != nil {
		return err
	}
	if err := validateCustodyLocks(s.Channels, s.CustodyLocks); err != nil {
		return err
	}
	if err := ValidateLockedCollateralForFinality(s); err != nil {
		return err
	}
	if err := validateClosedChannelTombstones(s.Channels, s.ClosedChannels); err != nil {
		return err
	}
	if err := validateConditionClaimRecords(s.Channels, s.ConditionClaims); err != nil {
		return err
	}
	return validatePaymentEvents(s.Channels, s.Events)
}

func (s PaymentsState) ChannelByID(channelID string) (ChannelRecord, bool) {
	_, channel, found := s.ChannelIndex(channelID)
	return channel, found
}

func (s PaymentsState) ChannelIndex(channelID string) (int, ChannelRecord, bool) {
	needle := normalizeHash(channelID)
	for i, channel := range s.Channels {
		channel = channel.Normalize()
		if channel.ChannelID == needle {
			return i, channel, true
		}
	}
	return 0, ChannelRecord{}, false
}

func (s PaymentsState) EdgeByKey(channelID, from, to string) (ChannelEdge, bool) {
	channelID = normalizeHash(channelID)
	for _, edge := range s.Edges {
		edge = edge.Normalize()
		if edge.ChannelID == channelID && edge.From == from && edge.To == to {
			return edge, true
		}
	}
	return ChannelEdge{}, false
}

func (s PaymentsState) VirtualChannelByID(id string) (VirtualChannel, bool) {
	needle := normalizeHash(id)
	for _, vc := range s.VirtualChannels {
		vc = vc.Normalize()
		if vc.VirtualChannelID == needle {
			return vc, true
		}
	}
	return VirtualChannel{}, false
}

func (s PaymentsState) StateHashDebug(channelID string) (StateHashDebug, error) {
	channel, found := s.Export().ChannelByID(channelID)
	if !found {
		return StateHashDebug{}, errors.New("payments channel not found")
	}
	debug := StateHashDebug{
		ChannelID:               channel.ChannelID,
		Status:                  channel.Status,
		LatestNonce:             channel.LatestState.Nonce,
		LatestStateHash:         channel.LatestState.StateHash,
		ComputedLatestStateHash: ComputeStateHash(channel.LatestState),
		FinalizedNonce:          channel.FinalizedNonce,
		DisputedNonce:           channel.DisputedNonce,
	}
	if channel.PendingClose.State.StateHash != "" {
		debug.PendingNonce = channel.PendingClose.State.Nonce
		debug.PendingStateHash = channel.PendingClose.State.StateHash
		debug.ComputedPendingStateHash = ComputeStateHash(channel.PendingClose.State)
	}
	return debug, nil
}

func (s PaymentsState) CustodyLockByChannel(channelID string) (CustodyLock, bool) {
	needle := normalizeHash(channelID)
	for _, lock := range s.CustodyLocks {
		lock = lock.Normalize()
		if lock.ChannelID == needle {
			return lock, true
		}
	}
	return CustodyLock{}, false
}

func (s PaymentsState) PendingFinalizationHeight(channelID string) (uint64, bool, error) {
	state := s.Export()
	channel, found := state.ChannelByID(channelID)
	if !found {
		return 0, false, errors.New("payments channel not found")
	}
	height, ok := PendingFinalizationHeightForChannel(channel)
	return height, ok, nil
}

func AdvanceChannelFinality(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments finality advance height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	nextFinality := FinalityAfterPendingClose(channel, currentHeight)
	if nextFinality == channel.Finality {
		return state, nil
	}
	next := state.Clone()
	nextChannel, err := setChannelFinality(channel, nextFinality, currentHeight, &next.Events)
	if err != nil {
		return PaymentsState{}, err
	}
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func ValidateLockedCollateralForFinality(state PaymentsState) error {
	state = state.Export()
	lockByChannel := make(map[string]CustodyLock, len(state.CustodyLocks))
	for _, lock := range state.CustodyLocks {
		lock = lock.Normalize()
		lockByChannel[lock.ChannelID] = lock
	}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		lock, locked := lockByChannel[channel.ChannelID]
		switch channel.Finality {
		case ChannelFinalitySettled, ChannelFinalityPenalized:
			if channel.Status == ChannelStatusSettled {
				if locked {
					return errors.New("payments settled finality must not retain custody lock")
				}
				continue
			}
		}
		if !locked {
			return errors.New("payments unsettled finality must retain custody lock")
		}
		if err := lock.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	return nil
}

func validateChannels(channels []ChannelRecord) error {
	seen := make(map[string]struct{}, len(channels))
	var previous string
	for i, channel := range channels {
		channel = channel.Normalize()
		if err := channel.Validate(); err != nil {
			return err
		}
		if _, found := seen[channel.ChannelID]; found {
			return errors.New("payments duplicate channel")
		}
		seen[channel.ChannelID] = struct{}{}
		if i > 0 && previous >= channel.ChannelID {
			return errors.New("payments channels must be sorted canonically")
		}
		previous = channel.ChannelID
	}
	return nil
}

func validateEdges(channels []ChannelRecord, edges []ChannelEdge) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(edges))
	var previous string
	for i, edge := range edges {
		edge = edge.Normalize()
		if err := edge.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[edge.ChannelID]
		if !found {
			return errors.New("payments routing edge references unknown channel")
		}
		if channel.Status != ChannelStatusOpen {
			return errors.New("payments routing edge references non-open channel")
		}
		if !containsString(channel.Participants, edge.From) || !containsString(channel.Participants, edge.To) {
			return errors.New("payments routing edge endpoints must be channel participants")
		}
		key := edgeKey(edge)
		if _, found := seen[key]; found {
			return errors.New("payments duplicate routing edge")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("payments routing edges must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateVirtualChannels(channels []ChannelRecord, virtualChannels []VirtualChannel) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(virtualChannels))
	var previous string
	for i, vc := range virtualChannels {
		vc = vc.Normalize()
		if err := vc.Validate(); err != nil {
			return err
		}
		for _, parentID := range vc.ParentChannelIDs {
			if _, found := channelByID[parentID]; !found {
				return errors.New("payments virtual channel references unknown parent")
			}
		}
		if _, found := seen[vc.VirtualChannelID]; found {
			return errors.New("payments duplicate virtual channel")
		}
		seen[vc.VirtualChannelID] = struct{}{}
		if i > 0 && previous >= vc.VirtualChannelID {
			return errors.New("payments virtual channels must be sorted canonically")
		}
		previous = vc.VirtualChannelID
	}
	return nil
}

func validateSettlements(channels []ChannelRecord, settlements []SettlementRecord) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(settlements))
	var previous string
	for i, settlement := range settlements {
		settlement = settlement.Normalize()
		channel, found := channelByID[settlement.ChannelID]
		if !found {
			return errors.New("payments settlement references unknown channel")
		}
		if err := settlement.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[settlement.ChannelID]; found {
			return errors.New("payments duplicate settlement")
		}
		seen[settlement.ChannelID] = struct{}{}
		if i > 0 && previous >= settlement.ChannelID {
			return errors.New("payments settlements must be sorted canonically")
		}
		previous = settlement.ChannelID
	}
	return nil
}

func validateBatches(channels []ChannelRecord, batches []SettlementBatch) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(batches))
	var previous string
	for i, batch := range batches {
		batch = batch.Normalize()
		if err := batch.Validate(); err != nil {
			return err
		}
		for _, op := range batch.Operations {
			if _, found := channelByID[op.ChannelID]; !found {
				return errors.New("payments batch references unknown channel")
			}
		}
		if _, found := seen[batch.BatchID]; found {
			return errors.New("payments duplicate batch")
		}
		seen[batch.BatchID] = struct{}{}
		if i > 0 && previous >= batch.BatchID {
			return errors.New("payments batches must be sorted canonically")
		}
		previous = batch.BatchID
	}
	return nil
}

func validateCustodyLocks(channels []ChannelRecord, locks []CustodyLock) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(locks))
	var previous string
	for i, lock := range locks {
		lock = lock.Normalize()
		channel, found := channelByID[lock.ChannelID]
		if !found {
			return errors.New("payments custody lock references unknown channel")
		}
		if channel.Status == ChannelStatusSettled {
			return errors.New("payments settled channel must not retain custody lock")
		}
		if err := lock.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[lock.ChannelID]; found {
			return errors.New("payments duplicate custody lock")
		}
		seen[lock.ChannelID] = struct{}{}
		if i > 0 && previous >= lock.ChannelID {
			return errors.New("payments custody locks must be sorted canonically")
		}
		previous = lock.ChannelID
	}
	for _, channel := range channelByID {
		if channel.Status == ChannelStatusSettled {
			continue
		}
		if _, found := seen[channel.ChannelID]; !found {
			return errors.New("payments channel custody lock is required")
		}
	}
	return nil
}

func validateClosedChannelTombstones(channels []ChannelRecord, tombstones []ClosedChannelTombstone) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(tombstones))
	var previous string
	for i, tombstone := range tombstones {
		tombstone = tombstone.Normalize()
		if err := tombstone.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[tombstone.ChannelID]
		if !found {
			return errors.New("payments tombstone references unknown channel")
		}
		if channel.Status != ChannelStatusSettled {
			return errors.New("payments tombstone requires settled channel")
		}
		if tombstone.ChainID != channel.ChainID || tombstone.FinalizedNonce != channel.FinalizedNonce {
			return errors.New("payments tombstone channel domain mismatch")
		}
		if _, found := seen[tombstone.ChannelID]; found {
			return errors.New("payments duplicate closed channel tombstone")
		}
		seen[tombstone.ChannelID] = struct{}{}
		if i > 0 && previous >= tombstone.ChannelID {
			return errors.New("payments closed channel tombstones must be sorted canonically")
		}
		previous = tombstone.ChannelID
	}
	for _, channel := range channelByID {
		if channel.Status != ChannelStatusSettled {
			continue
		}
		if _, found := seen[channel.ChannelID]; !found {
			return errors.New("payments settled channel tombstone is required")
		}
	}
	return nil
}

func validateConditionClaimRecords(channels []ChannelRecord, claims []ConditionClaimRecord) error {
	channelByID := channelMap(channels)
	seenCondition := make(map[string]struct{}, len(claims))
	seenEvidence := make(map[string]struct{}, len(claims))
	var previous string
	for i, claim := range claims {
		claim = claim.Normalize()
		if err := claim.Validate(); err != nil {
			return err
		}
		channel, found := channelByID[claim.ChannelID]
		if !found {
			return errors.New("payments condition claim references unknown channel")
		}
		if claim.ChainID != channel.ChainID {
			return errors.New("payments condition claim channel domain mismatch")
		}
		conditionKey := conditionClaimKey(claim.ChannelID, claim.ConditionID)
		evidenceKey := conditionEvidenceKey(claim.ChannelID, claim.EvidenceHash)
		if _, found := seenCondition[conditionKey]; found {
			return errors.New("payments duplicate condition claim")
		}
		if _, found := seenEvidence[evidenceKey]; found {
			return errors.New("payments duplicate condition evidence claim")
		}
		seenCondition[conditionKey] = struct{}{}
		seenEvidence[evidenceKey] = struct{}{}
		sortKey := conditionKey + "/" + claim.EvidenceHash
		if i > 0 && previous >= sortKey {
			return errors.New("payments condition claims must be sorted canonically")
		}
		previous = sortKey
	}
	return nil
}

func validatePaymentEvents(channels []ChannelRecord, events []PaymentEvent) error {
	channelByID := channelMap(channels)
	seen := make(map[string]struct{}, len(events))
	openEventByChannel := make(map[string]struct{}, len(channels))
	for _, event := range events {
		event = event.Normalize()
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := channelByID[event.ChannelID]; !found {
			return errors.New("payments event references unknown channel")
		}
		if _, found := seen[event.EventID]; found {
			return errors.New("payments duplicate event")
		}
		seen[event.EventID] = struct{}{}
		if event.EventType == "channel-open" {
			openEventByChannel[event.ChannelID] = struct{}{}
		}
	}
	for _, channel := range channelByID {
		if _, found := openEventByChannel[channel.ChannelID]; !found {
			return errors.New("payments channel-open event is required")
		}
	}
	return nil
}

func applySettlementAdjustments(balances []Balance, penalties []Penalty, allocations []PenaltyAllocation, feeText, feePayer string) ([]Balance, error) {
	amounts := make(map[string]sdkmath.Int, len(balances))
	for _, balance := range normalizeBalances(balances) {
		amount, err := parseNonNegativeInt("payments final balance", balance.Amount)
		if err != nil {
			return nil, err
		}
		amounts[balance.Participant] = amount
	}
	for _, penalty := range normalizePenalties(penalties) {
		amount, err := parsePositiveInt("payments penalty amount", penalty.Amount)
		if err != nil {
			return nil, err
		}
		offenderBalance, found := amounts[penalty.Offender]
		if !found || offenderBalance.LT(amount) {
			return nil, errors.New("payments penalty exceeds offender balance")
		}
		amounts[penalty.Offender] = offenderBalance.Sub(amount)
		amounts[penalty.Recipient] = amounts[penalty.Recipient].Add(amount)
	}
	for _, allocation := range normalizePenaltyAllocations(allocations) {
		amount, err := parsePositiveInt("payments penalty allocation amount", allocation.Amount)
		if err != nil {
			return nil, err
		}
		offenderBalance, found := amounts[allocation.Offender]
		if !found || offenderBalance.LT(amount) {
			return nil, errors.New("payments penalty allocation exceeds offender balance")
		}
		amounts[allocation.Offender] = offenderBalance.Sub(amount)
	}
	fee, err := parseNonNegativeInt("payments settlement fee", feeText)
	if err != nil {
		return nil, err
	}
	if fee.IsPositive() {
		balance, found := amounts[feePayer]
		if !found || balance.LT(fee) {
			return nil, errors.New("payments settlement fee exceeds payer balance")
		}
		amounts[feePayer] = balance.Sub(fee)
	}
	out := make([]Balance, 0, len(amounts))
	for participant, amount := range amounts {
		out = append(out, Balance{Participant: participant, Amount: amount.String()})
	}
	return normalizeBalances(out), nil
}

func settlementBalancesWithConditions(state ChannelState, channel ChannelRecord, resolutions []ConditionResolution) ([]Balance, error) {
	state = state.Normalize()
	if len(state.Conditions) == 0 {
		return state.Balances, nil
	}
	amounts := make(map[string]sdkmath.Int, len(state.Balances))
	for _, balance := range normalizeBalances(state.Balances) {
		amount, err := parseNonNegativeInt("payments settlement base balance", balance.Amount)
		if err != nil {
			return nil, err
		}
		amounts[balance.Participant] = amount
	}
	reserveByParticipant := map[string]sdkmath.Int{}
	if state.ChannelType == ChannelTypeBidirectional {
		reserveA, err := parseNonNegativeInt("payments settlement reserve a", state.ReserveA)
		if err != nil {
			return nil, err
		}
		reserveB, err := parseNonNegativeInt("payments settlement reserve b", state.ReserveB)
		if err != nil {
			return nil, err
		}
		reserveByParticipant[state.ParticipantA] = reserveA
		reserveByParticipant[state.ParticipantB] = reserveB
	}
	resolutionByID := make(map[string]ConditionResolution, len(resolutions))
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		resolutionByID[resolution.ConditionID] = resolution
	}
	for _, condition := range state.Conditions {
		condition = condition.Normalize()
		resolution, found := resolutionByID[condition.ConditionID]
		if !found {
			return nil, errors.New("payments condition is unresolved")
		}
		amount, err := parsePositiveInt("payments condition amount", condition.Amount)
		if err != nil {
			return nil, err
		}
		reserve := reserveByParticipant[condition.Payer]
		if reserve.LT(amount) {
			return nil, errors.New("payments condition exceeds reserved balance")
		}
		reserveByParticipant[condition.Payer] = reserve.Sub(amount)
		recipient := resolution.Recipient
		amounts[recipient] = amounts[recipient].Add(amount)
	}
	for participant, reserve := range reserveByParticipant {
		amounts[participant] = amounts[participant].Add(reserve)
	}
	out := make([]Balance, 0, len(amounts))
	for participant, amount := range amounts {
		if !containsString(channel.Participants, participant) {
			return nil, errors.New("payments settlement condition participant must be in channel")
		}
		out = append(out, Balance{Participant: participant, Amount: amount.String()})
	}
	return normalizeBalances(out), nil
}

func rejectReusedConditionClaims(state PaymentsState, channel ChannelRecord, resolutions []ConditionResolution) error {
	channel = channel.Normalize()
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		conditionKey := conditionClaimKey(channel.ChannelID, resolution.ConditionID)
		evidenceKey := conditionEvidenceKey(channel.ChannelID, resolution.EvidenceHash)
		for _, existing := range state.ConditionClaims {
			existing = existing.Normalize()
			if existing.ChainID != channel.ChainID || existing.ChannelID != channel.ChannelID {
				continue
			}
			if conditionClaimKey(existing.ChannelID, existing.ConditionID) == conditionKey {
				return errors.New("payments condition claim has already been used")
			}
			if conditionEvidenceKey(existing.ChannelID, existing.EvidenceHash) == evidenceKey {
				return errors.New("payments condition evidence claim has already been used")
			}
		}
	}
	return nil
}

func appendSettlementReplayRecords(state *PaymentsState, channel ChannelRecord, settlement SettlementRecord, resolutions []ConditionResolution, height uint64) {
	channel = channel.Normalize()
	settlement = settlement.Normalize()
	tombstone := ClosedChannelTombstone{
		ChainID:        channel.ChainID,
		ChannelID:      channel.ChannelID,
		FinalizedNonce: settlement.Nonce,
		StateHash:      settlement.StateHash,
		ClosedHeight:   height,
		ExpiresHeight:  height + DefaultReplayHorizon,
	}.Normalize()
	state.ClosedChannels = upsertClosedChannelTombstone(state.ClosedChannels, tombstone)
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		state.ConditionClaims = append(state.ConditionClaims, ConditionClaimRecord{
			ChainID:        channel.ChainID,
			ChannelID:      channel.ChannelID,
			ConditionID:    resolution.ConditionID,
			EvidenceHash:   resolution.EvidenceHash,
			ResolvedHeight: height,
			ExpiresHeight:  height + DefaultReplayHorizon,
		}.Normalize())
	}
}

func upsertClosedChannelTombstone(tombstones []ClosedChannelTombstone, next ClosedChannelTombstone) []ClosedChannelTombstone {
	out := make([]ClosedChannelTombstone, 0, len(tombstones)+1)
	replaced := false
	for _, tombstone := range tombstones {
		tombstone = tombstone.Normalize()
		if tombstone.ChannelID == next.ChannelID {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, tombstone)
	}
	if !replaced {
		out = append(out, next)
	}
	sortClosedChannelTombstones(out)
	return out
}

func finalBalancesForUnidirectionalClaim(channel ChannelRecord, claim UnidirectionalClaim, settlementFee, feePayer string) ([]Balance, error) {
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return nil, err
	}
	claimed, err := parseNonNegativeInt("payments claimed amount", claim.ClaimedAmount)
	if err != nil {
		return nil, err
	}
	if claimed.GT(collateral) {
		return nil, errors.New("payments claimed amount exceeds locked collateral")
	}
	return applySettlementAdjustments([]Balance{
		{Participant: channel.Payer, Amount: collateral.Sub(claimed).String()},
		{Participant: channel.Receiver, Amount: claimed.String()},
	}, nil, nil, settlementFee, feePayer)
}

func activeEdgesForAmount(edges []ChannelEdge, amount sdkmath.Int, currentHeight uint64) []ChannelEdge {
	out := make([]ChannelEdge, 0, len(edges))
	for _, edge := range edges {
		edge = edge.Normalize()
		capacity, err := parsePositiveInt("payments routing capacity", edge.Capacity)
		if err != nil {
			continue
		}
		if !edge.Active || capacity.LT(amount) {
			continue
		}
		if edge.ExpiresHeight > 0 && currentHeight > edge.ExpiresHeight {
			continue
		}
		out = append(out, edge)
	}
	return out
}

func filterEdgesForSettledChannel(edges []ChannelEdge, channelID string) []ChannelEdge {
	channelID = normalizeHash(channelID)
	out := make([]ChannelEdge, 0, len(edges))
	for _, edge := range edges {
		if edge.Normalize().ChannelID == channelID {
			continue
		}
		out = append(out, edge)
	}
	return out
}

func filterCustodyLocksForSettledChannel(locks []CustodyLock, channelID string) []CustodyLock {
	channelID = normalizeHash(channelID)
	out := make([]CustodyLock, 0, len(locks))
	for _, lock := range locks {
		if lock.Normalize().ChannelID == channelID {
			continue
		}
		out = append(out, lock)
	}
	return out
}

func stateStrongerThan(candidate, current ChannelState) bool {
	candidate = candidate.Normalize()
	current = current.Normalize()
	if candidate.Nonce > current.Nonce {
		return true
	}
	return candidate.ChannelType == ChannelTypeAsync && candidate.CheckpointNonce > current.CheckpointNonce
}

func mergeConditionResolutions(left, right []ConditionResolution) []ConditionResolution {
	byID := make(map[string]ConditionResolution, len(left)+len(right))
	for _, resolution := range normalizeConditionResolutions(left) {
		byID[resolution.ConditionID] = resolution
	}
	for _, resolution := range normalizeConditionResolutions(right) {
		byID[resolution.ConditionID] = resolution
	}
	out := make([]ConditionResolution, 0, len(byID))
	for _, resolution := range byID {
		out = append(out, resolution)
	}
	return normalizeConditionResolutions(out)
}

func channelMap(channels []ChannelRecord) map[string]ChannelRecord {
	out := make(map[string]ChannelRecord, len(channels))
	for _, channel := range channels {
		channel = channel.Normalize()
		out[channel.ChannelID] = channel
	}
	return out
}

func sortChannels(channels []ChannelRecord) {
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Normalize().ChannelID < channels[j].Normalize().ChannelID
	})
}

func sortEdges(edges []ChannelEdge) {
	sort.SliceStable(edges, func(i, j int) bool {
		return edgeKey(edges[i].Normalize()) < edgeKey(edges[j].Normalize())
	})
}

func sortVirtualChannels(channels []VirtualChannel) {
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Normalize().VirtualChannelID < channels[j].Normalize().VirtualChannelID
	})
}

func sortSettlements(settlements []SettlementRecord) {
	sort.SliceStable(settlements, func(i, j int) bool {
		return settlements[i].Normalize().ChannelID < settlements[j].Normalize().ChannelID
	})
}

func sortBatches(batches []SettlementBatch) {
	sort.SliceStable(batches, func(i, j int) bool {
		return batches[i].Normalize().BatchID < batches[j].Normalize().BatchID
	})
}

func sortCustodyLocks(locks []CustodyLock) {
	sort.SliceStable(locks, func(i, j int) bool {
		return locks[i].Normalize().ChannelID < locks[j].Normalize().ChannelID
	})
}

func sortClosedChannelTombstones(tombstones []ClosedChannelTombstone) {
	sort.SliceStable(tombstones, func(i, j int) bool {
		return tombstones[i].Normalize().ChannelID < tombstones[j].Normalize().ChannelID
	})
}

func sortConditionClaimRecords(claims []ConditionClaimRecord) {
	sort.SliceStable(claims, func(i, j int) bool {
		left := claims[i].Normalize()
		right := claims[j].Normalize()
		return conditionClaimKey(left.ChannelID, left.ConditionID)+"/"+left.EvidenceHash < conditionClaimKey(right.ChannelID, right.ConditionID)+"/"+right.EvidenceHash
	})
}

func conditionClaimKey(channelID, conditionID string) string {
	return normalizeHash(channelID) + "/" + normalizeHash(conditionID)
}

func conditionEvidenceKey(channelID, evidenceHash string) string {
	return normalizeHash(channelID) + "/" + normalizeHash(evidenceHash)
}

func edgeKey(edge ChannelEdge) string {
	return fmt.Sprintf("%s/%s/%s", edge.ChannelID, edge.From, edge.To)
}
