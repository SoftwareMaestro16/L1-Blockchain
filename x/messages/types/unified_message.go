package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	MaxTraceIDLength	= 128
	MaxShardIDLength	= 128
	MaxModuleRouteLength	= 128
)

type UnifiedMessageCapability string
type UnifiedExecutionMode string
type UnifiedOrderingClass string
type MessageLifecycleStage string

const (
	MessageCapabilityAccountDelivery	UnifiedMessageCapability	= "account_delivery"
	MessageCapabilityContractCall		UnifiedMessageCapability	= "contract_call"
	MessageCapabilityCrossShard		UnifiedMessageCapability	= "cross_shard_delivery"
	MessageCapabilityCrossZone		UnifiedMessageCapability	= "cross_zone_delivery"
	MessageCapabilityModule			UnifiedMessageCapability	= "module_message"
	MessageCapabilityProofRead		UnifiedMessageCapability	= "proof_backed_read"

	UnifiedExecutionSyncLocal	UnifiedExecutionMode	= "sync_local"
	UnifiedExecutionAsync		UnifiedExecutionMode	= "async"
	UnifiedExecutionDeferred	UnifiedExecutionMode	= "deferred"
	UnifiedExecutionPromise		UnifiedExecutionMode	= "promise"

	UnifiedOrderingUnordered	UnifiedOrderingClass	= "unordered"
	UnifiedOrderingSenderOrdered	UnifiedOrderingClass	= "sender_ordered"
	UnifiedOrderingReceiverOrdered	UnifiedOrderingClass	= "receiver_ordered"
	UnifiedOrderingObjectOrdered	UnifiedOrderingClass	= "object_ordered"
	UnifiedOrderingStrictTraceOrder	UnifiedOrderingClass	= "strict_trace_ordered"

	MessageLifecycleCreated		MessageLifecycleStage	= "created"
	MessageLifecycleRouted		MessageLifecycleStage	= "route_committed"
	MessageLifecycleOutbox		MessageLifecycleStage	= "source_outbox"
	MessageLifecycleCore		MessageLifecycleStage	= "core_finality"
	MessageLifecycleInbox		MessageLifecycleStage	= "destination_inbox"
	MessageLifecycleExecuted	MessageLifecycleStage	= "executed"
	MessageLifecycleReceipt		MessageLifecycleStage	= "receipt_committed"

	MessageLifecycleQueuedInSourceOutbox		MessageLifecycleStage	= "queued_in_source_outbox"
	MessageLifecycleCommittedInMessageRoot		MessageLifecycleStage	= "committed_in_message_root"
	MessageLifecycleEligibleForDelivery		MessageLifecycleStage	= "eligible_for_delivery"
	MessageLifecycleQueuedInDestinationInbox	MessageLifecycleStage	= "queued_in_destination_inbox"
	MessageLifecycleExecutedOrFailed		MessageLifecycleStage	= "executed_or_failed"
	MessageLifecycleBounceOrFinalize		MessageLifecycleStage	= "bounce_or_finalize"
)

type UnifiedMessageRoute struct {
	SourceZoneID		zonestypes.ZoneID
	SourceShardID		string
	DestinationZoneID	zonestypes.ZoneID
	DestinationShardID	string
	ModuleRoute		string
	OrderingClass		UnifiedOrderingClass
	ExecutionMode		UnifiedExecutionMode
	RouteCommitment		string
	CommittedHeight		uint64
	FinalityDelay		uint64
	DeliveryEligibleFrom	uint64
}

type UnifiedMessageMetadata struct {
	Capability		UnifiedMessageCapability
	TraceID			string
	ExecutionMode		UnifiedExecutionMode
	OrderingClass		UnifiedOrderingClass
	SourceShardID		string
	DestinationShardID	string
	ModuleRoute		string
	AuthProofHash		string
	StateProofHash		string
}

type UnifiedMessageObject struct {
	Message		Message
	Capability	UnifiedMessageCapability
	TraceID		string
	ExecutionMode	UnifiedExecutionMode
	OrderingClass	UnifiedOrderingClass
	Route		UnifiedMessageRoute
	AuthProofHash	string
	StateProofHash	string
	LifecycleStage	MessageLifecycleStage
	ObjectHash	string
}

type MessageLifecycleRecord struct {
	MessageID	[]byte
	Stage		MessageLifecycleStage
	Height		uint64
	MessageHash	string
	RouteCommitment	string
	ReceiptHash	[]byte
	RecordHash	string
}

func NewUnifiedMessageObject(msg Message, metadata UnifiedMessageMetadata, params MessageParams) (UnifiedMessageObject, error) {
	if err := msg.Validate(params); err != nil {
		return UnifiedMessageObject{}, err
	}
	metadata = normalizeUnifiedMessageMetadata(metadata)
	object := UnifiedMessageObject{
		Message:	msg.Clone(),
		Capability:	metadata.Capability,
		TraceID:	metadata.TraceID,
		ExecutionMode:	metadata.ExecutionMode,
		OrderingClass:	metadata.OrderingClass,
		AuthProofHash:	metadata.AuthProofHash,
		StateProofHash:	metadata.StateProofHash,
		LifecycleStage:	MessageLifecycleCreated,
	}
	object.Route = UnifiedMessageRoute{
		SourceZoneID:		msg.SourceZone,
		SourceShardID:		metadata.SourceShardID,
		DestinationZoneID:	msg.DestinationZone,
		DestinationShardID:	metadata.DestinationShardID,
		ModuleRoute:		metadata.ModuleRoute,
		OrderingClass:		metadata.OrderingClass,
		ExecutionMode:		metadata.ExecutionMode,
	}
	if err := object.ValidateFormat(params); err != nil {
		return UnifiedMessageObject{}, err
	}
	object.ObjectHash = ComputeUnifiedMessageObjectHash(object)
	return object, object.ValidateHash(params)
}

func CommitUnifiedMessageRoute(object UnifiedMessageObject, route UnifiedMessageRoute, params MessageParams) (UnifiedMessageObject, error) {
	if err := object.ValidateHash(params); err != nil {
		return UnifiedMessageObject{}, err
	}
	route = normalizeUnifiedMessageRoute(route)
	if route.SourceZoneID == "" {
		route.SourceZoneID = object.Message.SourceZone
	}
	if route.DestinationZoneID == "" {
		route.DestinationZoneID = object.Message.DestinationZone
	}
	if route.ExecutionMode == "" {
		route.ExecutionMode = object.ExecutionMode
	}
	if route.OrderingClass == "" {
		route.OrderingClass = object.OrderingClass
	}
	if route.CommittedHeight == 0 {
		return UnifiedMessageObject{}, errors.New("unified message route committed height must be positive")
	}
	route.DeliveryEligibleFrom = route.CommittedHeight + route.FinalityDelay
	route.RouteCommitment = ComputeUnifiedRouteCommitment(route)
	next := object.Clone()
	next.Route = route
	next.LifecycleStage = MessageLifecycleRouted
	next.ObjectHash = ComputeUnifiedMessageObjectHash(next)
	return next, next.ValidateHash(params)
}

func BuildMessageLifecycleRecord(object UnifiedMessageObject, stage MessageLifecycleStage, height uint64, receipt MessageReceipt, params MessageParams) (MessageLifecycleRecord, error) {
	if err := object.ValidateHash(params); err != nil {
		return MessageLifecycleRecord{}, err
	}
	if height == 0 {
		return MessageLifecycleRecord{}, errors.New("message lifecycle record height must be positive")
	}
	if !IsMessageLifecycleStage(stage) {
		return MessageLifecycleRecord{}, fmt.Errorf("unknown message lifecycle stage %q", stage)
	}
	record := MessageLifecycleRecord{
		MessageID:		append([]byte(nil), object.Message.MessageID...),
		Stage:			stage,
		Height:			height,
		MessageHash:		object.ObjectHash,
		RouteCommitment:	object.Route.RouteCommitment,
	}
	if stage == MessageLifecycleReceipt {
		if err := receipt.Validate(); err != nil {
			return MessageLifecycleRecord{}, err
		}
		record.ReceiptHash = append([]byte(nil), receipt.ReceiptHash...)
	}
	record.RecordHash = ComputeMessageLifecycleRecordHash(record)
	return record, record.Validate()
}

func (o UnifiedMessageObject) ValidateHash(params MessageParams) error {
	if err := o.ValidateFormat(params); err != nil {
		return err
	}
	if o.ObjectHash != ComputeUnifiedMessageObjectHash(o) {
		return errors.New("unified message object hash mismatch")
	}
	return nil
}

func (o UnifiedMessageObject) ValidateFormat(params MessageParams) error {
	if err := o.Message.Validate(params); err != nil {
		return err
	}
	if !IsUnifiedMessageCapability(o.Capability) {
		return fmt.Errorf("unknown unified message capability %q", o.Capability)
	}
	if err := validateToken("unified message trace id", o.TraceID, MaxTraceIDLength); err != nil {
		return err
	}
	if !IsUnifiedExecutionMode(o.ExecutionMode) {
		return fmt.Errorf("unknown unified message execution mode %q", o.ExecutionMode)
	}
	if !IsUnifiedOrderingClass(o.OrderingClass) {
		return fmt.Errorf("unknown unified message ordering class %q", o.OrderingClass)
	}
	if err := validateOptionalHash("unified message auth proof hash", o.AuthProofHash); err != nil {
		return err
	}
	if err := validateOptionalHash("unified message state proof hash", o.StateProofHash); err != nil {
		return err
	}
	if err := validateUnifiedMessageCapability(o); err != nil {
		return err
	}
	if o.Route.SourceZoneID != "" || o.Route.DestinationZoneID != "" || o.Route.RouteCommitment != "" || o.Route.CommittedHeight != 0 {
		if err := o.Route.ValidateFormat(); err != nil {
			return err
		}
		if o.Route.SourceZoneID != o.Message.SourceZone || o.Route.DestinationZoneID != o.Message.DestinationZone {
			return errors.New("unified message route zone mismatch")
		}
		if o.Route.ExecutionMode != o.ExecutionMode || o.Route.OrderingClass != o.OrderingClass {
			return errors.New("unified message route execution metadata mismatch")
		}
	}
	if !IsMessageLifecycleStage(o.LifecycleStage) {
		return fmt.Errorf("unknown message lifecycle stage %q", o.LifecycleStage)
	}
	if o.ObjectHash != "" {
		if err := zonestypes.ValidateHash("unified message object hash", o.ObjectHash); err != nil {
			return err
		}
	}
	return nil
}

func (r UnifiedMessageRoute) ValidateHash() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.RouteCommitment != ComputeUnifiedRouteCommitment(r) {
		return errors.New("unified message route commitment mismatch")
	}
	return nil
}

func (r UnifiedMessageRoute) ValidateFormat() error {
	if err := zonestypes.ValidateZoneID(r.SourceZoneID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.DestinationZoneID); err != nil {
		return err
	}
	if r.SourceShardID != "" {
		if err := validateToken("unified message source shard", r.SourceShardID, MaxShardIDLength); err != nil {
			return err
		}
	}
	if r.DestinationShardID != "" {
		if err := validateToken("unified message destination shard", r.DestinationShardID, MaxShardIDLength); err != nil {
			return err
		}
	}
	if r.ModuleRoute != "" {
		if err := validateToken("unified message module route", r.ModuleRoute, MaxModuleRouteLength); err != nil {
			return err
		}
	}
	if !IsUnifiedOrderingClass(r.OrderingClass) {
		return fmt.Errorf("unknown unified message route ordering class %q", r.OrderingClass)
	}
	if !IsUnifiedExecutionMode(r.ExecutionMode) {
		return fmt.Errorf("unknown unified message route execution mode %q", r.ExecutionMode)
	}
	if r.CommittedHeight != 0 && r.DeliveryEligibleFrom != r.CommittedHeight+r.FinalityDelay {
		return errors.New("unified message delivery eligibility height mismatch")
	}
	if r.RouteCommitment != "" {
		return zonestypes.ValidateHash("unified message route commitment", r.RouteCommitment)
	}
	return nil
}

func (r MessageLifecycleRecord) Validate() error {
	if len(r.MessageID) != MessageIDBytes {
		return fmt.Errorf("message lifecycle record id must be %d bytes", MessageIDBytes)
	}
	if !IsMessageLifecycleStage(r.Stage) {
		return fmt.Errorf("unknown message lifecycle stage %q", r.Stage)
	}
	if r.Height == 0 {
		return errors.New("message lifecycle record height must be positive")
	}
	if err := zonestypes.ValidateHash("message lifecycle message hash", r.MessageHash); err != nil {
		return err
	}
	if r.RouteCommitment != "" {
		if err := zonestypes.ValidateHash("message lifecycle route commitment", r.RouteCommitment); err != nil {
			return err
		}
	}
	if r.Stage == MessageLifecycleReceipt && len(r.ReceiptHash) != MessageIDBytes {
		return fmt.Errorf("message lifecycle receipt hash must be %d bytes", MessageIDBytes)
	}
	if err := zonestypes.ValidateHash("message lifecycle record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeMessageLifecycleRecordHash(r) {
		return errors.New("message lifecycle record hash mismatch")
	}
	return nil
}

func (o UnifiedMessageObject) Clone() UnifiedMessageObject {
	o.Message = o.Message.Clone()
	return o
}

func ComputeUnifiedRouteCommitment(route UnifiedMessageRoute) string {
	route = normalizeUnifiedMessageRoute(route)
	return hashParts(
		"aetra-unified-message-route-v1",
		string(route.SourceZoneID),
		route.SourceShardID,
		string(route.DestinationZoneID),
		route.DestinationShardID,
		route.ModuleRoute,
		string(route.OrderingClass),
		string(route.ExecutionMode),
		fmt.Sprint(route.CommittedHeight),
		fmt.Sprint(route.FinalityDelay),
		fmt.Sprint(route.DeliveryEligibleFrom),
	)
}

func ComputeUnifiedMessageObjectHash(object UnifiedMessageObject) string {
	return hashParts(
		"aetra-unified-message-object-v1",
		hex.EncodeToString(object.Message.MessageID),
		string(object.Capability),
		object.TraceID,
		string(object.ExecutionMode),
		string(object.OrderingClass),
		object.Route.RouteCommitment,
		object.AuthProofHash,
		object.StateProofHash,
		string(object.LifecycleStage),
	)
}

func ComputeMessageLifecycleRecordHash(record MessageLifecycleRecord) string {
	return hashParts(
		"aetra-unified-message-lifecycle-record-v1",
		hex.EncodeToString(record.MessageID),
		string(record.Stage),
		fmt.Sprint(record.Height),
		record.MessageHash,
		record.RouteCommitment,
		hex.EncodeToString(record.ReceiptHash),
	)
}

func ComputeMessageLifecycleRoot(records []MessageLifecycleRecord) (string, error) {
	ordered := cloneMessageLifecycleRecords(records)
	sortMessageLifecycleRecords(ordered)
	parts := []string{"aetra-unified-message-lifecycle-root-v1", fmt.Sprint(len(ordered))}
	for _, record := range ordered {
		if err := record.Validate(); err != nil {
			return "", err
		}
		parts = append(parts, record.RecordHash)
	}
	return hashParts(parts...), nil
}

func IsUnifiedMessageCapability(capability UnifiedMessageCapability) bool {
	switch capability {
	case MessageCapabilityAccountDelivery, MessageCapabilityContractCall, MessageCapabilityCrossShard, MessageCapabilityCrossZone, MessageCapabilityModule, MessageCapabilityProofRead:
		return true
	default:
		return false
	}
}

func IsUnifiedExecutionMode(mode UnifiedExecutionMode) bool {
	switch mode {
	case UnifiedExecutionSyncLocal, UnifiedExecutionAsync, UnifiedExecutionDeferred, UnifiedExecutionPromise:
		return true
	default:
		return false
	}
}

func IsUnifiedOrderingClass(class UnifiedOrderingClass) bool {
	switch class {
	case UnifiedOrderingUnordered, UnifiedOrderingSenderOrdered, UnifiedOrderingReceiverOrdered, UnifiedOrderingObjectOrdered, UnifiedOrderingStrictTraceOrder:
		return true
	default:
		return false
	}
}

func IsMessageLifecycleStage(stage MessageLifecycleStage) bool {
	switch stage {
	case MessageLifecycleCreated,
		MessageLifecycleRouted,
		MessageLifecycleOutbox,
		MessageLifecycleCore,
		MessageLifecycleInbox,
		MessageLifecycleExecuted,
		MessageLifecycleReceipt,
		MessageLifecycleQueuedInSourceOutbox,
		MessageLifecycleCommittedInMessageRoot,
		MessageLifecycleEligibleForDelivery,
		MessageLifecycleQueuedInDestinationInbox,
		MessageLifecycleExecutedOrFailed,
		MessageLifecycleBounceOrFinalize:
		return true
	default:
		return false
	}
}

func normalizeUnifiedMessageMetadata(metadata UnifiedMessageMetadata) UnifiedMessageMetadata {
	metadata.TraceID = strings.TrimSpace(metadata.TraceID)
	metadata.SourceShardID = strings.TrimSpace(metadata.SourceShardID)
	metadata.DestinationShardID = strings.TrimSpace(metadata.DestinationShardID)
	metadata.ModuleRoute = strings.TrimSpace(metadata.ModuleRoute)
	metadata.AuthProofHash = strings.ToLower(strings.TrimSpace(metadata.AuthProofHash))
	metadata.StateProofHash = strings.ToLower(strings.TrimSpace(metadata.StateProofHash))
	if metadata.ExecutionMode == "" {
		metadata.ExecutionMode = UnifiedExecutionAsync
	}
	if metadata.OrderingClass == "" {
		metadata.OrderingClass = UnifiedOrderingSenderOrdered
	}
	return metadata
}

func normalizeUnifiedMessageRoute(route UnifiedMessageRoute) UnifiedMessageRoute {
	route.SourceShardID = strings.TrimSpace(route.SourceShardID)
	route.DestinationShardID = strings.TrimSpace(route.DestinationShardID)
	route.ModuleRoute = strings.TrimSpace(route.ModuleRoute)
	route.RouteCommitment = strings.ToLower(strings.TrimSpace(route.RouteCommitment))
	return route
}

func validateUnifiedMessageCapability(object UnifiedMessageObject) error {
	switch object.Capability {
	case MessageCapabilityAccountDelivery:
		if object.Message.Value.IsNegative() || object.Message.FeeLimit.IsNegative() || object.Message.Nonce == 0 {
			return errors.New("account delivery requires sender nonce, value, and fee metadata")
		}
	case MessageCapabilityContractCall:
		if object.TraceID == "" || object.Message.GasLimit == 0 || object.ExecutionMode == "" || object.Message.Opcode == "" {
			return errors.New("contract call requires trace, payload type, gas, and execution mode")
		}
	case MessageCapabilityCrossShard:
		if object.Route.SourceShardID == "" || object.Route.DestinationShardID == "" {
			return errors.New("cross-shard message requires source and destination shards")
		}
	case MessageCapabilityCrossZone:
		if object.Message.SourceZone == object.Message.DestinationZone {
			return errors.New("cross-zone message requires distinct zones")
		}
	case MessageCapabilityModule:
		if object.Route.ModuleRoute == "" {
			return errors.New("module message requires explicit module route")
		}
	case MessageCapabilityProofRead:
		if object.StateProofHash == "" {
			return errors.New("proof-backed read requires committed proof input")
		}
	}
	return nil
}

func validateOptionalHash(fieldName string, value string) error {
	if value == "" {
		return nil
	}
	return zonestypes.ValidateHash(fieldName, value)
}

func cloneMessageLifecycleRecords(records []MessageLifecycleRecord) []MessageLifecycleRecord {
	out := make([]MessageLifecycleRecord, len(records))
	for i, record := range records {
		out[i] = record
		out[i].MessageID = append([]byte(nil), record.MessageID...)
		out[i].ReceiptHash = append([]byte(nil), record.ReceiptHash...)
	}
	return out
}

func sortMessageLifecycleRecords(records []MessageLifecycleRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Height != records[j].Height {
			return records[i].Height < records[j].Height
		}
		if records[i].Stage != records[j].Stage {
			return records[i].Stage < records[j].Stage
		}
		return hex.EncodeToString(records[i].MessageID) < hex.EncodeToString(records[j].MessageID)
	})
}
