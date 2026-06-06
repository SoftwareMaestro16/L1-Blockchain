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
		Events:          []PaymentEvent{},
	}
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
	if channel.ChannelType != ChannelTypeAsync && nextState.PreviousStateHash != channel.LatestState.StateHash {
		return PaymentsState{}, errors.New("payments channel state previous hash must match latest state")
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
		EvidenceHash:    HashParts("async-dispute", checkpoint.StateHash, ComputeAsyncDeltaRoot(deltas)),
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

func SubmitClose(state PaymentsState, channelID string, closingState ChannelState, submitter string, currentHeight uint64, settlementFee string) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments close height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return PaymentsState{}, errors.New("payments channel is not open")
	}
	pending := PendingClose{
		Submitter:          submitter,
		SubmittedHeight:    currentHeight,
		SettleAfterHeight:  currentHeight + channel.DisputePeriod,
		SettlementFeeDenom: NativeDenom,
		SettlementFee:      settlementFee,
		State:              closingState.Normalize(),
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
	next := state.Clone()
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
	next := state.Clone()
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
	if err := closingState.ValidateForChannel(channel, true); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	if closingState.Nonce < channel.LatestState.Nonce {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments cooperative close state nonce is below latest accepted nonce")
	}
	finalBalances, err := applySettlementAdjustments(closingState.Balances, nil, settlementFee, submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
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
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
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
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
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
		}, nil, settlementFee, payer)
	} else {
		stateHash = claim.StateHash
		nonce = claim.Nonce
		finalBalances, err = finalBalancesForUnidirectionalClaim(channel, claim, settlementFee, payer)
	}
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
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
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.PendingClose = PendingClose{}
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	return next, settlement, next.Validate()
}

func DisputeClose(state PaymentsState, channelID string, newerState ChannelState, submitter string, currentHeight uint64) (PaymentsState, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments dispute height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, errors.New("payments channel is not pending close")
	}
	if currentHeight > channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, errors.New("payments dispute window has closed")
	}
	newerState = newerState.Normalize()
	if err := newerState.ValidateForChannel(channel, false); err != nil {
		return PaymentsState{}, err
	}
	if newerState.Nonce <= channel.PendingClose.State.Nonce {
		return PaymentsState{}, errors.New("payments dispute state nonce must be newer")
	}
	if !containsString(channel.Participants, submitter) {
		return PaymentsState{}, errors.New("payments dispute submitter must be participant")
	}
	nextChannel := channel
	nextChannel.PendingClose.State = newerState
	nextChannel.PendingClose.SubmittedHeight = currentHeight
	nextChannel.PendingClose.SettleAfterHeight = currentHeight + channel.DisputePeriod
	nextChannel.LatestState = newerState
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	sortChannels(next.Channels)
	return next, next.Validate()
}

func SubmitFraudProof(state PaymentsState, channelID string, proof FraudProof, currentHeight uint64) (PaymentsState, error) {
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
	if err := proof.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	for _, existing := range channel.PendingClose.FraudProofs {
		if existing.ProofID == proof.ProofID {
			return PaymentsState{}, errors.New("payments duplicate fraud proof")
		}
	}
	penalty := Penalty{Offender: proof.OffendingSigner, Recipient: proof.SubmittedBy, Denom: NativeDenom, Amount: proof.PenaltyAmount}.Normalize()
	if err := penalty.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	nextChannel := channel
	nextChannel.PendingClose.FraudProofs = append(nextChannel.PendingClose.FraudProofs, proof)
	nextChannel.PendingClose.Penalties = append(nextChannel.PendingClose.Penalties, penalty)
	next := state.Clone()
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
	finalBalances, err := applySettlementAdjustments(channel.PendingClose.State.Balances, channel.PendingClose.Penalties, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChannelID:          channel.ChannelID,
		StateHash:          channel.PendingClose.State.StateHash,
		Nonce:              channel.PendingClose.State.Nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: channel.PendingClose.SettlementFeeDenom,
		SettlementFee:      channel.PendingClose.SettlementFee,
		Penalties:          channel.PendingClose.Penalties,
		SettledHeight:      currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.PendingClose = PendingClose{}
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	return next, settlement, next.Validate()
}

func FinalizeSettlement(state PaymentsState, channelID string, currentHeight uint64) (PaymentsState, SettlementRecord, error) {
	state = state.Export()
	if currentHeight == 0 {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement height must be positive")
	}
	index, channel, found := state.ChannelIndex(channelID)
	if !found {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel not found")
	}
	if channel.Status != ChannelStatusPendingClose {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments channel is not pending close")
	}
	if currentHeight < channel.PendingClose.SettleAfterHeight {
		return PaymentsState{}, SettlementRecord{}, errors.New("payments settlement is still in dispute window")
	}
	finalBalances, err := applySettlementAdjustments(channel.PendingClose.State.Balances, channel.PendingClose.Penalties, channel.PendingClose.SettlementFee, channel.PendingClose.Submitter)
	if err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	settlement := SettlementRecord{
		ChannelID:          channel.ChannelID,
		StateHash:          channel.PendingClose.State.StateHash,
		Nonce:              channel.PendingClose.State.Nonce,
		FinalBalances:      finalBalances,
		SettlementFeeDenom: channel.PendingClose.SettlementFeeDenom,
		SettlementFee:      channel.PendingClose.SettlementFee,
		Penalties:          channel.PendingClose.Penalties,
		SettledHeight:      currentHeight,
	}
	settlement.SettlementHash = ComputeSettlementHash(settlement)
	if err := settlement.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, SettlementRecord{}, err
	}
	nextChannel := channel
	nextChannel.Status = ChannelStatusSettled
	nextChannel.FinalizedNonce = settlement.Nonce
	nextChannel.PendingClose = PendingClose{}
	next := state.Clone()
	next.Channels[index] = nextChannel.Normalize()
	next.Edges = filterEdgesForSettledChannel(next.Edges, channel.ChannelID)
	next.Settlements = append(next.Settlements, settlement)
	sortChannels(next.Channels)
	sortSettlements(next.Settlements)
	return next, settlement, next.Validate()
}

func OpenVirtualChannel(state PaymentsState, vc VirtualChannel) (PaymentsState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentsState{}, err
	}
	vc = vc.Normalize()
	if vc.AnchorCommitment == "" {
		vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	}
	if err := vc.Validate(); err != nil {
		return PaymentsState{}, err
	}
	if _, found := state.VirtualChannelByID(vc.VirtualChannelID); found {
		return PaymentsState{}, errors.New("payments virtual channel already exists")
	}
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return PaymentsState{}, err
	}
	for _, parentID := range vc.ParentChannelIDs {
		channel, found := state.ChannelByID(parentID)
		if !found || channel.Status != ChannelStatusOpen {
			return PaymentsState{}, errors.New("payments virtual channel requires open parents")
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
		if _, found := seen[channel.ChannelID]; !found {
			return errors.New("payments channel custody lock is required")
		}
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

func applySettlementAdjustments(balances []Balance, penalties []Penalty, feeText, feePayer string) ([]Balance, error) {
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
	}, nil, settlementFee, feePayer)
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

func edgeKey(edge ChannelEdge) string {
	return fmt.Sprintf("%s/%s/%s", edge.ChannelID, edge.From, edge.To)
}
