package types

import (
	"errors"
	"fmt"
)

func EmptyState(params MeshParams) MeshState {
	return MeshState{
		Params:			NormalizeParams(params),
		Destinations:		[]MeshDestination{},
		FinalizedCommitments:	[]FinalizedCommitment{},
		ReplayMarkers:		[]ReplayMarker{},
		Receipts:		[]MeshReceipt{},
		BounceReceipts:		[]BounceReceipt{},
		RefundReceipts:		[]RefundReceipt{},
	}
}

func NewMessage(msg MeshMessage) (MeshMessage, error) {
	msg = msg.Normalize()
	if err := msg.Validate(); err != nil {
		return MeshMessage{}, err
	}
	return msg, nil
}

func RegisterDestination(state MeshState, destination MeshDestination) (MeshState, error) {
	state = normalizeStateParams(state)
	if err := state.Validate(); err != nil {
		return MeshState{}, err
	}
	if err := destination.Validate(); err != nil {
		return MeshState{}, err
	}
	if hasDestination(state, destination.ZoneID, destination.ShardID) {
		return MeshState{}, errors.New("mesh destination already registered")
	}
	next := state.Clone()
	next.Destinations = append(next.Destinations, destination)
	sortDestinations(next.Destinations)
	return next, next.Validate()
}

func AddFinalizedCommitment(state MeshState, commitment FinalizedCommitment) (MeshState, error) {
	state = normalizeStateParams(state)
	if err := state.Validate(); err != nil {
		return MeshState{}, err
	}
	if err := commitment.Validate(); err != nil {
		return MeshState{}, err
	}
	if hasCommitment(state, commitment.ZoneID, commitment.ShardID, commitment.Height) {
		return MeshState{}, errors.New("mesh finalized commitment already exists")
	}
	next := state.Clone()
	next.FinalizedCommitments = append(next.FinalizedCommitments, commitment)
	sortCommitments(next.FinalizedCommitments)
	return next, next.Validate()
}

func ApplyMessage(state MeshState, msg MeshMessage, result ExecutionResult, currentHeight uint64) (MeshState, MeshReceipt, error) {
	state = normalizeStateParams(state)
	if currentHeight == 0 {
		return MeshState{}, MeshReceipt{}, errors.New("mesh current height must be positive")
	}
	if err := state.Validate(); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	msg = msg.Normalize()
	if err := msg.Validate(); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	if hasReplayMarker(state, msg.MessageID) {
		return MeshState{}, MeshReceipt{}, errors.New("mesh message replay detected")
	}
	if hasReceipt(state, msg.MessageID) {
		return MeshState{}, MeshReceipt{}, errors.New("mesh duplicate receipt detected")
	}
	if err := ValidateSourceProof(state, msg, currentHeight); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	if currentHeight > msg.TimeoutHeight {
		return commitFailure(state, msg, currentHeight, FailureReasonExpired)
	}
	if !hasActiveDestination(state, msg.DestinationZone, msg.DestinationShard) {
		return commitFailure(state, msg, currentHeight, FailureReasonInvalidDestination)
	}
	if err := result.Validate(); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	if !result.Success {
		return commitFailureWithResult(state, msg, currentHeight, FailureReasonExecutionFailed, result)
	}
	return commitSuccess(state, msg, currentHeight, result)
}

func ValidateSourceProof(state MeshState, msg MeshMessage, currentHeight uint64) error {
	params := NormalizeParams(state.Params)
	if currentHeight < msg.Finality.Height {
		return errors.New("mesh source finality is in the future")
	}
	if currentHeight > msg.Finality.Height+params.MaxFinalityAge {
		return errors.New("mesh source finality reference is stale")
	}
	if err := msg.Proof.Validate(); err != nil {
		return fmt.Errorf("missing or invalid mesh source proof: %w", err)
	}
	commitment, found := findCommitment(state, msg.SourceZone, msg.SourceShard, msg.Finality.Height)
	if !found {
		return errors.New("mesh finalized source commitment not found")
	}
	if commitment.CommitmentHash != msg.Finality.CommitmentHash {
		return errors.New("mesh finality commitment mismatch")
	}
	if msg.Proof.SourceCommitment != commitment.CommitmentHash {
		return errors.New("mesh proof source commitment mismatch")
	}
	if msg.Proof.MessageRoot != commitment.MessageRoot {
		return errors.New("mesh proof message root mismatch")
	}
	if expected := ComputeProofHash(msg, commitment); msg.Proof.ProofHash != expected {
		return errors.New("mesh proof hash mismatch")
	}
	return nil
}

func CommitReceipt(state MeshState, receipt MeshReceipt) (MeshState, error) {
	state = normalizeStateParams(state)
	if err := state.Validate(); err != nil {
		return MeshState{}, err
	}
	if err := receipt.Validate(); err != nil {
		return MeshState{}, err
	}
	if hasReceipt(state, receipt.MessageID) {
		return MeshState{}, errors.New("mesh duplicate receipt detected")
	}
	next := state.Clone()
	next.Receipts = append(next.Receipts, receipt)
	sortReceipts(next.Receipts)
	return next, next.Validate()
}

func ImportState(state MeshState) (MeshState, error) {
	state = normalizeStateParams(state)
	if err := state.Validate(); err != nil {
		return MeshState{}, err
	}
	return state.Export(), nil
}

func (s MeshState) Export() MeshState {
	out := s.Clone()
	out.Params = NormalizeParams(out.Params)
	sortDestinations(out.Destinations)
	sortCommitments(out.FinalizedCommitments)
	sortReplayMarkers(out.ReplayMarkers)
	sortReceipts(out.Receipts)
	sortBounceReceipts(out.BounceReceipts)
	sortRefundReceipts(out.RefundReceipts)
	return out
}

func (s MeshState) Clone() MeshState {
	out := MeshState{
		CurrentHeight:		s.CurrentHeight,
		Params:			s.Params,
		Destinations:		append([]MeshDestination(nil), s.Destinations...),
		FinalizedCommitments:	append([]FinalizedCommitment(nil), s.FinalizedCommitments...),
		ReplayMarkers:		append([]ReplayMarker(nil), s.ReplayMarkers...),
		Receipts:		append([]MeshReceipt(nil), s.Receipts...),
		BounceReceipts:		append([]BounceReceipt(nil), s.BounceReceipts...),
		RefundReceipts:		make([]RefundReceipt, len(s.RefundReceipts)),
	}
	for i, receipt := range s.RefundReceipts {
		receipt.Recipient = cloneBytes(receipt.Recipient)
		out.RefundReceipts[i] = receipt
	}
	return out
}

func (s MeshState) Validate() error {
	params := NormalizeParams(s.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validateDestinations(s.Destinations); err != nil {
		return err
	}
	if err := validateCommitments(s.FinalizedCommitments); err != nil {
		return err
	}
	if err := validateReplayMarkers(s.ReplayMarkers); err != nil {
		return err
	}
	if err := validateReceipts(s.Receipts); err != nil {
		return err
	}
	if err := validateBounceReceipts(s.BounceReceipts); err != nil {
		return err
	}
	return validateRefundReceipts(s.RefundReceipts)
}

func normalizeStateParams(state MeshState) MeshState {
	state.Params = NormalizeParams(state.Params)
	return state
}

func commitSuccess(state MeshState, msg MeshMessage, height uint64, result ExecutionResult) (MeshState, MeshReceipt, error) {
	receipt := buildReceipt(msg, ReceiptStatusSuccess, FailureReasonNone, height, result)
	return commitDelivery(state, msg, receipt, nil, nil)
}

func commitFailure(state MeshState, msg MeshMessage, height uint64, reason FailureReason) (MeshState, MeshReceipt, error) {
	result := ExecutionResult{
		Success:	false,
		Code:		1,
		ResultHash:	HashParts("mesh-failure", msg.MessageID, string(reason)),
	}
	return commitFailureWithResult(state, msg, height, reason, result)
}

func commitFailureWithResult(state MeshState, msg MeshMessage, height uint64, reason FailureReason, result ExecutionResult) (MeshState, MeshReceipt, error) {
	if msg.Kind != MessageKindNormal {
		receipt := buildReceipt(msg, ReceiptStatusTerminalFailure, reason, height, result)
		return commitDelivery(state, msg, receipt, nil, nil)
	}
	if reason == FailureReasonExecutionFailed {
		receipt := buildReceipt(msg, ReceiptStatusRefunded, reason, height, result)
		refund := buildRefundReceipt(msg, receipt)
		return commitDelivery(state, msg, receipt, nil, &refund)
	}
	receipt := buildReceipt(msg, ReceiptStatusBounced, reason, height, result)
	bounce := buildBounceReceipt(msg, receipt)
	return commitDelivery(state, msg, receipt, &bounce, nil)
}

func commitDelivery(state MeshState, msg MeshMessage, receipt MeshReceipt, bounce *BounceReceipt, refund *RefundReceipt) (MeshState, MeshReceipt, error) {
	if err := receipt.Validate(); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	if hasReplayMarker(state, msg.MessageID) {
		return MeshState{}, MeshReceipt{}, errors.New("mesh message replay detected")
	}
	if hasReceipt(state, msg.MessageID) {
		return MeshState{}, MeshReceipt{}, errors.New("mesh duplicate receipt detected")
	}
	marker := ReplayMarker{
		MessageID:	msg.MessageID,
		ReceiptHash:	receipt.ReceiptHash,
		Reason:		receipt.Reason,
		Height:		receipt.Height,
	}
	if err := marker.Validate(); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	next := state.Clone()
	next.CurrentHeight = maxUint64(next.CurrentHeight, receipt.Height)
	next.ReplayMarkers = append(next.ReplayMarkers, marker)
	next.Receipts = append(next.Receipts, receipt)
	if bounce != nil {
		if err := bounce.Validate(); err != nil {
			return MeshState{}, MeshReceipt{}, err
		}
		next.BounceReceipts = append(next.BounceReceipts, *bounce)
	}
	if refund != nil {
		if err := refund.Validate(); err != nil {
			return MeshState{}, MeshReceipt{}, err
		}
		next.RefundReceipts = append(next.RefundReceipts, *refund)
	}
	sortReplayMarkers(next.ReplayMarkers)
	sortReceipts(next.Receipts)
	sortBounceReceipts(next.BounceReceipts)
	sortRefundReceipts(next.RefundReceipts)
	if err := next.Validate(); err != nil {
		return MeshState{}, MeshReceipt{}, err
	}
	return next, receipt, nil
}

func buildReceipt(msg MeshMessage, status ReceiptStatus, reason FailureReason, height uint64, result ExecutionResult) MeshReceipt {
	receipt := MeshReceipt{
		MessageID:		msg.MessageID,
		SourceZone:		msg.SourceZone,
		SourceShard:		msg.SourceShard,
		DestinationZone:	msg.DestinationZone,
		DestinationShard:	msg.DestinationShard,
		Status:			status,
		Reason:			reason,
		Height:			height,
		Sequence:		msg.Sequence,
		ExecutionCode:		result.Code,
		ResultHash:		result.ResultHash,
	}
	receipt.ReceiptHash = ComputeReceiptHash(receipt)
	return receipt
}

func buildBounceReceipt(msg MeshMessage, receipt MeshReceipt) BounceReceipt {
	bounce := msg
	bounce.SourceZone, bounce.DestinationZone = msg.DestinationZone, msg.SourceZone
	bounce.SourceShard, bounce.DestinationShard = msg.DestinationShard, msg.SourceShard
	bounce.Nonce++
	bounce.Sequence++
	bounce.SourceLogicalTime = receipt.Height
	bounce.Kind = MessageKindBounce
	bounce.ParentMessageID = msg.MessageID
	bounce.MessageID = ComputeMessageID(bounce)
	out := BounceReceipt{
		MessageID:		HashParts("bounce-receipt", msg.MessageID, receipt.ReceiptHash),
		SourceMessageID:	msg.MessageID,
		BounceMessageID:	bounce.MessageID,
		DestinationZone:	bounce.DestinationZone,
		DestinationShard:	bounce.DestinationShard,
		Reason:			receipt.Reason,
		Height:			receipt.Height,
	}
	out.ReceiptHash = ComputeBounceReceiptHash(out)
	return out
}

func buildRefundReceipt(msg MeshMessage, receipt MeshReceipt) RefundReceipt {
	out := RefundReceipt{
		MessageID:		HashParts("refund-receipt", msg.MessageID, receipt.ReceiptHash),
		SourceMessageID:	msg.MessageID,
		Recipient:		cloneBytes(msg.Sender),
		AssetCommitment:	msg.AssetCommitment,
		Reason:			receipt.Reason,
		Height:			receipt.Height,
	}
	out.ReceiptHash = ComputeRefundReceiptHash(out)
	return out
}
