package types

import (
	"errors"
	"fmt"
)

type ExecutionMode string
type DeliveryStatus string
type ReceiptStatus string
type FailureReason string

const (
	ExecutionModeSync	ExecutionMode	= "SYNC"
	ExecutionModeAsync	ExecutionMode	= "ASYNC"

	DeliveryStatusClassified	DeliveryStatus	= "CLASSIFIED"
	DeliveryStatusPrepared		DeliveryStatus	= "PREPARED"
	DeliveryStatusExecuted		DeliveryStatus	= "EXECUTED"
	DeliveryStatusRootCommitted	DeliveryStatus	= "ROOT_COMMITTED"
	DeliveryStatusNextBlockEligible	DeliveryStatus	= "NEXT_BLOCK_ELIGIBLE"
	DeliveryStatusDelivered		DeliveryStatus	= "DELIVERED"
	DeliveryStatusBounced		DeliveryStatus	= "BOUNCED"
	DeliveryStatusRefunded		DeliveryStatus	= "REFUNDED"
	DeliveryStatusFailed		DeliveryStatus	= "FAILED"

	ReceiptStatusSuccess		ReceiptStatus	= "SUCCESS"
	ReceiptStatusBounced		ReceiptStatus	= "BOUNCED"
	ReceiptStatusRefunded		ReceiptStatus	= "REFUNDED"
	ReceiptStatusTerminalFailure	ReceiptStatus	= "TERMINAL_FAILURE"

	FailureReasonNone			FailureReason	= ""
	FailureReasonExecutionFailed		FailureReason	= "EXECUTION_FAILED"
	FailureReasonInvalidDestination		FailureReason	= "INVALID_DESTINATION"
	FailureReasonExpired			FailureReason	= "EXPIRED"
	FailureReasonMissingCommittedRoot	FailureReason	= "MISSING_COMMITTED_ROOT"
	FailureReasonMissingActiveShard		FailureReason	= "MISSING_ACTIVE_SHARD"
	FailureReasonCrossZoneProofRejected	FailureReason	= "CROSS_ZONE_PROOF_REJECTED"
)

type ClassificationInput struct {
	Height			uint64
	TxHash			string
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	PriorityClass		uint32
	AdmissionHeight		uint64
	TxIndex			uint32
	MessageIndex		uint32
}

type ClassifiedTransaction struct {
	Height			uint64
	TxHash			string
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	ExecutionMode		ExecutionMode
	ProposalItem		ProposalItem
	DeliveryStatus		DeliveryStatus
}

type ExecutionResult struct {
	Success		bool
	Code		uint32
	ResultHash	string
}

type ExecutionReceipt struct {
	TxHash			string
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	ExecutionMode		ExecutionMode
	Status			ReceiptStatus
	Reason			FailureReason
	Height			uint64
	Sequence		uint64
	ExecutionCode		uint32
	ResultHash		string
	ReceiptHash		string
}

type CrossZoneDelivery struct {
	TxHash			string
	SourceZone		ZoneID
	SourceShard		ShardID
	DestinationZone		ZoneID
	DestinationShard	ShardID
	CommittedHeight		uint64
	EligibleHeight		uint64
	SourceCommitmentHash	string
	MessageRoot		string
	DeliveryItem		ProposalItem
	Status			DeliveryStatus
	Receipt			ExecutionReceipt
	BounceReceipt		ExecutionReceipt
}

func ClassifyTransaction(state CoreState, input ClassificationInput) (ClassifiedTransaction, error) {
	if input.Height == 0 {
		return ClassifiedTransaction{}, errors.New("aetracore classification height must be positive")
	}
	if err := ValidateHash("aetracore classification tx hash", input.TxHash); err != nil {
		return ClassifiedTransaction{}, err
	}
	if input.AdmissionHeight == 0 {
		input.AdmissionHeight = input.Height
	}
	if input.AdmissionHeight > input.Height {
		return ClassifiedTransaction{}, errors.New("aetracore admission height must not exceed classification height")
	}
	if err := validateZoneAndShard(state, input.SourceZone, input.SourceShard, input.Height); err != nil {
		return ClassifiedTransaction{}, err
	}
	if err := validateZoneAndShard(state, input.DestinationZone, input.DestinationShard, input.Height); err != nil {
		return ClassifiedTransaction{}, err
	}
	mode := ExecutionModeSync
	if input.SourceZone != input.DestinationZone {
		mode = ExecutionModeAsync
	}
	classified := ClassifiedTransaction{
		Height:			input.Height,
		TxHash:			input.TxHash,
		SourceZone:		input.SourceZone,
		SourceShard:		input.SourceShard,
		DestinationZone:	input.DestinationZone,
		DestinationShard:	input.DestinationShard,
		ExecutionMode:		mode,
		ProposalItem: ProposalItem{
			ZoneID:			input.SourceZone,
			ShardID:		input.SourceShard,
			TxHash:			input.TxHash,
			PriorityClass:		input.PriorityClass,
			AdmissionHeight:	input.AdmissionHeight,
			TxIndex:		input.TxIndex,
			MessageIndex:		input.MessageIndex,
		},
		DeliveryStatus:	DeliveryStatusClassified,
	}
	if err := classified.Validate(); err != nil {
		return ClassifiedTransaction{}, err
	}
	return classified, nil
}

func ExecuteSync(classified ClassifiedTransaction, result ExecutionResult, height uint64, sequence uint64) (ExecutionReceipt, error) {
	if classified.ExecutionMode != ExecutionModeSync {
		return ExecutionReceipt{}, errors.New("aetracore sync execution requires same-zone classification")
	}
	return buildExecutionReceipt(classified, result, height, sequence, FailureReasonExecutionFailed, false)
}

func ExecuteAsyncLocal(classified ClassifiedTransaction, result ExecutionResult, height uint64, sequence uint64) (ExecutionReceipt, error) {
	if classified.ExecutionMode != ExecutionModeAsync {
		return ExecutionReceipt{}, errors.New("aetracore async local execution requires cross-zone classification")
	}
	return buildExecutionReceipt(classified, result, height, sequence, FailureReasonExecutionFailed, true)
}

func MarkCrossZoneEligible(state CoreState, classified ClassifiedTransaction, committedHeight uint64) (CrossZoneDelivery, error) {
	if classified.ExecutionMode != ExecutionModeAsync {
		return CrossZoneDelivery{}, errors.New("aetracore cross-zone delivery requires async classification")
	}
	if committedHeight == 0 {
		return CrossZoneDelivery{}, errors.New("aetracore committed height must be positive")
	}
	if err := classified.Validate(); err != nil {
		return CrossZoneDelivery{}, err
	}
	if _, found := state.RootSnapshotAtHeight(committedHeight); !found {
		return CrossZoneDelivery{}, errors.New("aetracore cross-zone delivery requires committed roots")
	}
	commitment, found := state.ZoneCommitmentAtHeight(committedHeight, classified.SourceZone)
	if !found {
		return CrossZoneDelivery{}, errors.New("aetracore cross-zone source commitment is not committed")
	}
	delay := state.Params.CrossZoneFinalityDelay
	if delay == 0 {
		delay = 1
	}
	delivery := CrossZoneDelivery{
		TxHash:			classified.TxHash,
		SourceZone:		classified.SourceZone,
		SourceShard:		classified.SourceShard,
		DestinationZone:	classified.DestinationZone,
		DestinationShard:	classified.DestinationShard,
		CommittedHeight:	committedHeight,
		EligibleHeight:		committedHeight + delay,
		SourceCommitmentHash:	commitment.CommitmentHash,
		MessageRoot:		commitment.OutboxRoot,
		DeliveryItem: ProposalItem{
			ZoneID:			classified.DestinationZone,
			ShardID:		classified.DestinationShard,
			TxHash:			classified.TxHash,
			PriorityClass:		classified.ProposalItem.PriorityClass,
			AdmissionHeight:	committedHeight + delay,
			TxIndex:		classified.ProposalItem.TxIndex,
			MessageIndex:		classified.ProposalItem.MessageIndex,
		},
		Status:	DeliveryStatusNextBlockEligible,
	}
	return delivery, delivery.Validate()
}

func DeliverCrossZone(delivery CrossZoneDelivery, result ExecutionResult, height uint64, sequence uint64) (CrossZoneDelivery, error) {
	if err := delivery.Validate(); err != nil {
		return CrossZoneDelivery{}, err
	}
	if height < delivery.EligibleHeight {
		return CrossZoneDelivery{}, errors.New("aetracore cross-zone delivery is not yet eligible")
	}
	classified := ClassifiedTransaction{
		Height:			delivery.CommittedHeight,
		TxHash:			delivery.TxHash,
		SourceZone:		delivery.SourceZone,
		SourceShard:		delivery.SourceShard,
		DestinationZone:	delivery.DestinationZone,
		DestinationShard:	delivery.DestinationShard,
		ExecutionMode:		ExecutionModeAsync,
		DeliveryStatus:		DeliveryStatusNextBlockEligible,
	}
	receipt, err := buildExecutionReceipt(classified, result, height, sequence, FailureReasonExecutionFailed, false)
	if err != nil {
		return CrossZoneDelivery{}, err
	}
	next := delivery
	next.Receipt = receipt
	switch receipt.Status {
	case ReceiptStatusSuccess:
		next.Status = DeliveryStatusDelivered
	case ReceiptStatusRefunded:
		next.Status = DeliveryStatusRefunded
		next.BounceReceipt = buildBounceReceipt(receipt, height, sequence+1)
	case ReceiptStatusBounced:
		next.Status = DeliveryStatusBounced
		next.BounceReceipt = buildBounceReceipt(receipt, height, sequence+1)
	default:
		next.Status = DeliveryStatusFailed
	}
	return next, next.Validate()
}

func FailCrossZoneDelivery(delivery CrossZoneDelivery, reason FailureReason, height uint64, sequence uint64) (CrossZoneDelivery, error) {
	if err := delivery.Validate(); err != nil {
		return CrossZoneDelivery{}, err
	}
	if height < delivery.EligibleHeight {
		return CrossZoneDelivery{}, errors.New("aetracore cross-zone delivery is not yet eligible")
	}
	if reason == FailureReasonNone {
		return CrossZoneDelivery{}, errors.New("aetracore failure reason is required")
	}
	if !IsFailureReason(reason) {
		return CrossZoneDelivery{}, fmt.Errorf("unknown aetracore delivery failure reason %q", reason)
	}
	result := ExecutionResult{Success: false, Code: 1, ResultHash: hashParts("aetracore-delivery-failure", delivery.TxHash, string(reason))}
	classified := ClassifiedTransaction{
		Height:			delivery.CommittedHeight,
		TxHash:			delivery.TxHash,
		SourceZone:		delivery.SourceZone,
		SourceShard:		delivery.SourceShard,
		DestinationZone:	delivery.DestinationZone,
		DestinationShard:	delivery.DestinationShard,
		ExecutionMode:		ExecutionModeAsync,
	}
	receipt, err := buildExecutionReceiptWithReason(classified, result, height, sequence, reason, false)
	if err != nil {
		return CrossZoneDelivery{}, err
	}
	next := delivery
	next.Receipt = receipt
	if reason == FailureReasonExecutionFailed {
		next.Status = DeliveryStatusRefunded
	} else {
		next.Status = DeliveryStatusBounced
	}
	next.BounceReceipt = buildBounceReceipt(receipt, height, sequence+1)
	return next, next.Validate()
}

func (c ClassifiedTransaction) Validate() error {
	if c.Height == 0 {
		return errors.New("aetracore classified tx height must be positive")
	}
	if err := ValidateHash("aetracore classified tx hash", c.TxHash); err != nil {
		return err
	}
	if err := ValidateZoneID(c.SourceZone); err != nil {
		return err
	}
	if err := ValidateShardID(c.SourceShard); err != nil {
		return err
	}
	if err := ValidateZoneID(c.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(c.DestinationShard); err != nil {
		return err
	}
	if c.ExecutionMode != ExecutionModeSync && c.ExecutionMode != ExecutionModeAsync {
		return fmt.Errorf("unknown aetracore execution mode %q", c.ExecutionMode)
	}
	if c.ExecutionMode == ExecutionModeSync && c.SourceZone != c.DestinationZone {
		return errors.New("aetracore sync execution must stay within one zone")
	}
	if c.ExecutionMode == ExecutionModeAsync && c.SourceZone == c.DestinationZone {
		return errors.New("aetracore async execution requires cross-zone delivery")
	}
	if c.DeliveryStatus != "" && !IsDeliveryStatus(c.DeliveryStatus) {
		return fmt.Errorf("unknown aetracore classified delivery status %q", c.DeliveryStatus)
	}
	if c.ProposalItem.TxHash != "" {
		if c.ProposalItem.TxHash != c.TxHash {
			return errors.New("aetracore classified proposal item tx hash mismatch")
		}
		if c.ProposalItem.ZoneID != c.SourceZone || c.ProposalItem.ShardID != c.SourceShard {
			return errors.New("aetracore classified proposal item route mismatch")
		}
		if c.ProposalItem.AdmissionHeight == 0 || c.ProposalItem.AdmissionHeight > c.Height {
			return errors.New("aetracore classified proposal item admission height is invalid")
		}
		return c.ProposalItem.Validate()
	}
	return nil
}

func (r ExecutionResult) Validate() error {
	return ValidateHash("aetracore execution result hash", r.ResultHash)
}

func (r ExecutionReceipt) Validate() error {
	if err := ValidateHash("aetracore execution receipt tx hash", r.TxHash); err != nil {
		return err
	}
	if err := ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := ValidateShardID(r.SourceShard); err != nil {
		return err
	}
	if err := ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(r.DestinationShard); err != nil {
		return err
	}
	if !IsReceiptStatus(r.Status) {
		return fmt.Errorf("unknown aetracore receipt status %q", r.Status)
	}
	if r.Status == ReceiptStatusSuccess && r.Reason != FailureReasonNone {
		return errors.New("aetracore success receipt must not include failure reason")
	}
	if r.Status != ReceiptStatusSuccess && r.Reason == FailureReasonNone {
		return errors.New("aetracore failed receipt requires failure reason")
	}
	if r.Status != ReceiptStatusSuccess && !IsFailureReason(r.Reason) {
		return fmt.Errorf("unknown aetracore receipt failure reason %q", r.Reason)
	}
	if r.Height == 0 {
		return errors.New("aetracore receipt height must be positive")
	}
	if err := ValidateHash("aetracore receipt result hash", r.ResultHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore receipt hash", r.ReceiptHash); err != nil {
		return err
	}
	if expected := ComputeExecutionReceiptHash(r); r.ReceiptHash != expected {
		return errors.New("aetracore receipt hash mismatch")
	}
	return nil
}

func (d CrossZoneDelivery) Validate() error {
	if err := ValidateHash("aetracore cross-zone tx hash", d.TxHash); err != nil {
		return err
	}
	if err := ValidateZoneID(d.SourceZone); err != nil {
		return err
	}
	if err := ValidateShardID(d.SourceShard); err != nil {
		return err
	}
	if err := ValidateZoneID(d.DestinationZone); err != nil {
		return err
	}
	if err := ValidateShardID(d.DestinationShard); err != nil {
		return err
	}
	if d.SourceZone == d.DestinationZone {
		return errors.New("aetracore cross-zone delivery requires different zones")
	}
	if d.CommittedHeight == 0 || d.EligibleHeight == 0 {
		return errors.New("aetracore cross-zone delivery heights must be positive")
	}
	if d.EligibleHeight < d.CommittedHeight {
		return errors.New("aetracore cross-zone eligible height must not precede committed height")
	}
	if err := ValidateHash("aetracore cross-zone source commitment", d.SourceCommitmentHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore cross-zone message root", d.MessageRoot); err != nil {
		return err
	}
	if d.DeliveryItem.TxHash != "" {
		if d.DeliveryItem.TxHash != d.TxHash {
			return errors.New("aetracore cross-zone delivery item tx hash mismatch")
		}
		if d.DeliveryItem.ZoneID != d.DestinationZone || d.DeliveryItem.ShardID != d.DestinationShard {
			return errors.New("aetracore cross-zone delivery item route mismatch")
		}
		if d.DeliveryItem.AdmissionHeight != d.EligibleHeight {
			return errors.New("aetracore cross-zone delivery item admission height must match eligible height")
		}
		if err := d.DeliveryItem.Validate(); err != nil {
			return err
		}
	}
	if !IsDeliveryStatus(d.Status) {
		return fmt.Errorf("unknown aetracore delivery status %q", d.Status)
	}
	if d.Receipt.ReceiptHash != "" {
		if err := d.Receipt.Validate(); err != nil {
			return err
		}
		if err := d.validateDeliveryReceiptRoute(); err != nil {
			return err
		}
	}
	if d.BounceReceipt.ReceiptHash != "" {
		if err := d.BounceReceipt.Validate(); err != nil {
			return err
		}
		if err := d.validateBounceReceiptRoute(); err != nil {
			return err
		}
	}
	return d.validateDeliveryLifecycle()
}

func (d CrossZoneDelivery) validateDeliveryReceiptRoute() error {
	if d.Receipt.TxHash != d.TxHash {
		return errors.New("aetracore cross-zone receipt tx hash mismatch")
	}
	if d.Receipt.SourceZone != d.SourceZone || d.Receipt.SourceShard != d.SourceShard {
		return errors.New("aetracore cross-zone receipt source route mismatch")
	}
	if d.Receipt.DestinationZone != d.DestinationZone || d.Receipt.DestinationShard != d.DestinationShard {
		return errors.New("aetracore cross-zone receipt destination route mismatch")
	}
	if d.Receipt.ExecutionMode != ExecutionModeAsync {
		return errors.New("aetracore cross-zone receipt must be async")
	}
	return nil
}

func (d CrossZoneDelivery) validateBounceReceiptRoute() error {
	if d.BounceReceipt.SourceZone != d.DestinationZone || d.BounceReceipt.SourceShard != d.DestinationShard {
		return errors.New("aetracore cross-zone bounce source route mismatch")
	}
	if d.BounceReceipt.DestinationZone != d.SourceZone || d.BounceReceipt.DestinationShard != d.SourceShard {
		return errors.New("aetracore cross-zone bounce destination route mismatch")
	}
	if d.BounceReceipt.ExecutionMode != ExecutionModeAsync {
		return errors.New("aetracore cross-zone bounce receipt must be async")
	}
	return nil
}

func (d CrossZoneDelivery) validateDeliveryLifecycle() error {
	hasReceipt := d.Receipt.ReceiptHash != ""
	hasBounce := d.BounceReceipt.ReceiptHash != ""
	switch d.Status {
	case DeliveryStatusNextBlockEligible:
		if hasReceipt || hasBounce {
			return errors.New("aetracore eligible delivery must not include execution receipts")
		}
	case DeliveryStatusDelivered:
		if !hasReceipt || d.Receipt.Status != ReceiptStatusSuccess {
			return errors.New("aetracore delivered cross-zone message requires success receipt")
		}
		if hasBounce {
			return errors.New("aetracore delivered cross-zone message must not include bounce receipt")
		}
	case DeliveryStatusRefunded:
		if !hasReceipt || d.Receipt.Status != ReceiptStatusRefunded {
			return errors.New("aetracore refunded cross-zone message requires refund receipt")
		}
		if !hasBounce || d.BounceReceipt.Status != ReceiptStatusBounced {
			return errors.New("aetracore refunded cross-zone message requires bounce receipt")
		}
	case DeliveryStatusBounced:
		if !hasReceipt || d.Receipt.Status != ReceiptStatusBounced {
			return errors.New("aetracore bounced cross-zone message requires bounced receipt")
		}
		if !hasBounce || d.BounceReceipt.Status != ReceiptStatusBounced {
			return errors.New("aetracore bounced cross-zone message requires bounce receipt")
		}
	}
	return nil
}

func IsDeliveryStatus(status DeliveryStatus) bool {
	switch status {
	case DeliveryStatusClassified, DeliveryStatusPrepared, DeliveryStatusExecuted, DeliveryStatusRootCommitted,
		DeliveryStatusNextBlockEligible, DeliveryStatusDelivered, DeliveryStatusBounced, DeliveryStatusRefunded, DeliveryStatusFailed:
		return true
	default:
		return false
	}
}

func IsReceiptStatus(status ReceiptStatus) bool {
	switch status {
	case ReceiptStatusSuccess, ReceiptStatusBounced, ReceiptStatusRefunded, ReceiptStatusTerminalFailure:
		return true
	default:
		return false
	}
}

func IsFailureReason(reason FailureReason) bool {
	switch reason {
	case FailureReasonNone, FailureReasonExecutionFailed, FailureReasonInvalidDestination, FailureReasonExpired,
		FailureReasonMissingCommittedRoot, FailureReasonMissingActiveShard, FailureReasonCrossZoneProofRejected:
		return true
	default:
		return false
	}
}

func ComputeExecutionReceiptHash(receipt ExecutionReceipt) string {
	return hashParts(
		"aetra-aek-execution-receipt-v1",
		receipt.TxHash,
		string(receipt.SourceZone),
		string(receipt.SourceShard),
		string(receipt.DestinationZone),
		string(receipt.DestinationShard),
		string(receipt.ExecutionMode),
		string(receipt.Status),
		string(receipt.Reason),
		fmt.Sprint(receipt.Height),
		fmt.Sprint(receipt.Sequence),
		fmt.Sprint(receipt.ExecutionCode),
		receipt.ResultHash,
	)
}

func validateZoneAndShard(state CoreState, zoneID ZoneID, shardID ShardID, height uint64) error {
	descriptor, found := state.ZoneDescriptorByID(zoneID)
	if !found {
		return fmt.Errorf("aetracore classification zone %s is not registered", zoneID)
	}
	if !descriptor.Enabled {
		return fmt.Errorf("aetracore classification zone %s is disabled", zoneID)
	}
	if zoneID == ZoneIDAetraCore {
		return nil
	}
	layout, found := state.LatestShardLayout(zoneID, height)
	if !found {
		return fmt.Errorf("aetracore classification zone %s has no active shard layout", zoneID)
	}
	if !layout.HasActiveShard(shardID) {
		return fmt.Errorf("aetracore classification shard %s is not active for zone %s", shardID, zoneID)
	}
	return nil
}

func buildExecutionReceipt(classified ClassifiedTransaction, result ExecutionResult, height uint64, sequence uint64, failureReason FailureReason, terminalAsyncFailure bool) (ExecutionReceipt, error) {
	if !result.Success && failureReason == FailureReasonNone {
		return ExecutionReceipt{}, errors.New("aetracore failed execution requires failure reason")
	}
	return buildExecutionReceiptWithReason(classified, result, height, sequence, failureReason, terminalAsyncFailure)
}

func buildExecutionReceiptWithReason(classified ClassifiedTransaction, result ExecutionResult, height uint64, sequence uint64, failureReason FailureReason, terminalAsyncFailure bool) (ExecutionReceipt, error) {
	if height == 0 {
		return ExecutionReceipt{}, errors.New("aetracore receipt height must be positive")
	}
	if err := classified.Validate(); err != nil {
		return ExecutionReceipt{}, err
	}
	if err := result.Validate(); err != nil {
		return ExecutionReceipt{}, err
	}
	status := ReceiptStatusSuccess
	reason := FailureReasonNone
	if !result.Success {
		reason = failureReason
		if classified.ExecutionMode == ExecutionModeSync || terminalAsyncFailure {
			status = ReceiptStatusTerminalFailure
		} else if reason == FailureReasonExecutionFailed {
			status = ReceiptStatusRefunded
		} else {
			status = ReceiptStatusBounced
		}
	}
	receipt := ExecutionReceipt{
		TxHash:			classified.TxHash,
		SourceZone:		classified.SourceZone,
		SourceShard:		classified.SourceShard,
		DestinationZone:	classified.DestinationZone,
		DestinationShard:	classified.DestinationShard,
		ExecutionMode:		classified.ExecutionMode,
		Status:			status,
		Reason:			reason,
		Height:			height,
		Sequence:		sequence,
		ExecutionCode:		result.Code,
		ResultHash:		result.ResultHash,
	}
	receipt.ReceiptHash = ComputeExecutionReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func buildBounceReceipt(receipt ExecutionReceipt, height uint64, sequence uint64) ExecutionReceipt {
	bounce := ExecutionReceipt{
		TxHash:			hashParts("aetracore-bounce", receipt.TxHash, receipt.ReceiptHash),
		SourceZone:		receipt.DestinationZone,
		SourceShard:		receipt.DestinationShard,
		DestinationZone:	receipt.SourceZone,
		DestinationShard:	receipt.SourceShard,
		ExecutionMode:		ExecutionModeAsync,
		Status:			ReceiptStatusBounced,
		Reason:			receipt.Reason,
		Height:			height,
		Sequence:		sequence,
		ExecutionCode:		receipt.ExecutionCode,
		ResultHash:		hashParts("aetracore-bounce-result", receipt.ReceiptHash),
	}
	bounce.ReceiptHash = ComputeExecutionReceiptHash(bounce)
	return bounce
}

func (s CoreState) ZoneCommitmentAtHeight(height uint64, zoneID ZoneID) (ZoneCommitment, bool) {
	for _, commitment := range s.ZoneCommitments {
		if commitment.Height == height && commitment.ZoneID == zoneID {
			return commitment, true
		}
	}
	return ZoneCommitment{}, false
}

func (s CoreState) RootSnapshotAtHeight(height uint64) (RootSnapshot, bool) {
	for _, snapshot := range s.RootSnapshots {
		if snapshot.Height == height {
			out := snapshot
			out.ProofRoots = append([]ProofRoot(nil), snapshot.ProofRoots...)
			return out, true
		}
	}
	return RootSnapshot{}, false
}
