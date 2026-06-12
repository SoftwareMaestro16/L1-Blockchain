package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMStateAccessRead	AVMStateAccessMode	= "read"
	AVMStateAccessWrite	AVMStateAccessMode	= "write"

	AVMStateAccessTargetZone	AVMStateAccessTarget	= "zone"
	AVMStateAccessTargetActor	AVMStateAccessTarget	= "actor"
	AVMStateAccessTargetContract	AVMStateAccessTarget	= "contract"
)

type AVMStateAccessMode string
type AVMStateAccessTarget string

type AVMNonceKeeper struct {
	States		[]AVMReplayNonceState
	KeeperRoot	string
}

type AVMReplayTombstoneStore struct {
	ConsumedTombstones	[]AVMAsyncReplayTombstone
	ExpiredNonces		[]AVMExpiredNonceTombstone
	StoreRoot		string
}

type AVMZoneAccessCapability struct {
	SourceZone	zonestypes.ZoneID
	ActorIDOptional	string
	ContractAddress	string
	CapabilityHash	string
}

type AVMStateAccessRequest struct {
	SourceZone	zonestypes.ZoneID
	TargetZone	zonestypes.ZoneID
	Mode		AVMStateAccessMode
	Target		AVMStateAccessTarget
	ActorIDOptional	string
	ContractAddress	string
	StateKey	string
	ProofHash	string
	ReadOnlyProof	bool
	RequestHash	string
}

type AVMSchedulerRetryBound struct {
	MessageID	string
	Attempt		uint32
	MaxAttempts	uint32
}

type AVMSchedulerSafetyCheck struct {
	ZoneID				zonestypes.ZoneID
	Height				uint64
	ReadyMessages			[]AVMAsyncMessage
	ExpiredMessages			[]AVMAsyncMessage
	Budget				zonestypes.ZoneExecutionBudget
	RetryBounds			[]AVMSchedulerRetryBound
	RequireSenderNonceOrdering	bool
	RejectedEarlyMessageIDs		[]string
	RejectedExpiredExecutionIDs	[]string
	SchedulerCheckHash		string
}

func NewAVMNonceKeeper(keeper AVMNonceKeeper) (AVMNonceKeeper, error) {
	keeper = canonicalAVMNonceKeeper(keeper)
	keeper.KeeperRoot = ComputeAVMNonceKeeperRoot(keeper)
	return keeper, keeper.Validate()
}

func (k AVMNonceKeeper) Validate() error {
	k = canonicalAVMNonceKeeper(k)
	seen := make(map[string]struct{}, len(k.States))
	var previous string
	for _, state := range k.States {
		if err := state.Validate(); err != nil {
			return err
		}
		scope := AVMNonceKeeperScope(state)
		if _, found := seen[scope]; found {
			return fmt.Errorf("duplicate AVM nonce keeper scope %q", scope)
		}
		if previous != "" && previous >= scope {
			return errors.New("AVM nonce keeper states must be sorted canonically")
		}
		previous = scope
		seen[scope] = struct{}{}
	}
	if k.KeeperRoot == "" {
		return errors.New("AVM nonce keeper root is required")
	}
	if err := zonestypes.ValidateHash("AVM nonce keeper root", k.KeeperRoot); err != nil {
		return err
	}
	if k.KeeperRoot != ComputeAVMNonceKeeperRoot(k) {
		return errors.New("AVM nonce keeper root mismatch")
	}
	return nil
}

func UpsertAVMNonceKeeperState(keeper AVMNonceKeeper, state AVMReplayNonceState) (AVMNonceKeeper, error) {
	keeper = canonicalAVMNonceKeeper(keeper)
	keeper.KeeperRoot = ComputeAVMNonceKeeperRoot(keeper)
	if err := keeper.Validate(); err != nil {
		return AVMNonceKeeper{}, err
	}
	state = canonicalAVMReplayNonceState(state)
	if err := state.Validate(); err != nil {
		return AVMNonceKeeper{}, err
	}
	scope := AVMNonceKeeperScope(state)
	replaced := false
	for i := range keeper.States {
		if AVMNonceKeeperScope(keeper.States[i]) == scope {
			keeper.States[i] = state
			replaced = true
			break
		}
	}
	if !replaced {
		keeper.States = append(keeper.States, state)
	}
	keeper = canonicalAVMNonceKeeper(keeper)
	keeper.KeeperRoot = ComputeAVMNonceKeeperRoot(keeper)
	return keeper, keeper.Validate()
}

func NewAVMReplayTombstoneStore(store AVMReplayTombstoneStore) (AVMReplayTombstoneStore, error) {
	store = canonicalAVMReplayTombstoneStore(store)
	store.StoreRoot = ComputeAVMReplayTombstoneStoreRoot(store)
	return store, store.Validate()
}

func (s AVMReplayTombstoneStore) Validate() error {
	s = canonicalAVMReplayTombstoneStore(s)
	seenConsumed := make(map[string]struct{}, len(s.ConsumedTombstones))
	for i, tombstone := range s.ConsumedTombstones {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		if _, found := seenConsumed[tombstone.MessageID]; found {
			return fmt.Errorf("duplicate AVM replay tombstone store message %q", tombstone.MessageID)
		}
		if i > 0 && s.ConsumedTombstones[i-1].MessageID >= tombstone.MessageID {
			return errors.New("AVM replay tombstone store consumed entries must be sorted canonically")
		}
		seenConsumed[tombstone.MessageID] = struct{}{}
	}
	seenExpired := make(map[string]struct{}, len(s.ExpiredNonces))
	for i, tombstone := range s.ExpiredNonces {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		scope := AVMReplayNonceScope(tombstone.ChainID, tombstone.SourceZone, tombstone.Sender, tombstone.SenderNonce)
		if _, found := seenExpired[scope]; found {
			return fmt.Errorf("duplicate AVM replay tombstone store expired nonce %q", scope)
		}
		if i > 0 && AVMExpiredNonceTombstoneSortKey(s.ExpiredNonces[i-1]) >= AVMExpiredNonceTombstoneSortKey(tombstone) {
			return errors.New("AVM replay tombstone store expired entries must be sorted canonically")
		}
		seenExpired[scope] = struct{}{}
	}
	if s.StoreRoot == "" {
		return errors.New("AVM replay tombstone store root is required")
	}
	if err := zonestypes.ValidateHash("AVM replay tombstone store root", s.StoreRoot); err != nil {
		return err
	}
	if s.StoreRoot != ComputeAVMReplayTombstoneStoreRoot(s) {
		return errors.New("AVM replay tombstone store root mismatch")
	}
	return nil
}

func NewAVMZoneAccessCapability(capability AVMZoneAccessCapability) (AVMZoneAccessCapability, error) {
	capability = canonicalAVMZoneAccessCapability(capability)
	capability.CapabilityHash = ComputeAVMZoneAccessCapabilityHash(capability)
	return capability, capability.Validate()
}

func (c AVMZoneAccessCapability) Validate() error {
	c = canonicalAVMZoneAccessCapability(c)
	if err := zonestypes.ValidateZoneID(c.SourceZone); err != nil {
		return err
	}
	if c.ActorIDOptional != "" {
		if err := validateEngineToken("AVM zone access actor id", c.ActorIDOptional, MaxActorRuntimeTokenLength); err != nil {
			return err
		}
	}
	if c.ContractAddress != "" {
		if err := validateEngineToken("AVM zone access contract address", c.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
			return err
		}
	}
	if c.CapabilityHash == "" {
		return errors.New("AVM zone access capability hash is required")
	}
	if err := zonestypes.ValidateHash("AVM zone access capability hash", c.CapabilityHash); err != nil {
		return err
	}
	if c.CapabilityHash != ComputeAVMZoneAccessCapabilityHash(c) {
		return errors.New("AVM zone access capability hash mismatch")
	}
	return nil
}

func NewAVMStateAccessRequest(request AVMStateAccessRequest) (AVMStateAccessRequest, error) {
	request = canonicalAVMStateAccessRequest(request)
	request.RequestHash = ComputeAVMStateAccessRequestHash(request)
	return request, request.Validate()
}

func (r AVMStateAccessRequest) Validate() error {
	r = canonicalAVMStateAccessRequest(r)
	if err := zonestypes.ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.TargetZone); err != nil {
		return err
	}
	if !IsAVMStateAccessMode(r.Mode) {
		return fmt.Errorf("invalid AVM state access mode %q", r.Mode)
	}
	if !IsAVMStateAccessTarget(r.Target) {
		return fmt.Errorf("invalid AVM state access target %q", r.Target)
	}
	if r.ActorIDOptional != "" {
		if err := validateEngineToken("AVM state access actor id", r.ActorIDOptional, MaxActorRuntimeTokenLength); err != nil {
			return err
		}
	}
	if r.ContractAddress != "" {
		if err := validateEngineToken("AVM state access contract address", r.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
			return err
		}
	}
	if r.StateKey == "" {
		return errors.New("AVM state access key is required")
	}
	if r.Target == AVMStateAccessTargetActor {
		if !strings.HasPrefix(r.StateKey, "actor/") {
			return errors.New("AVM actor state access key must use actor prefix")
		}
	} else if err := validateAVMStatePrefix("AVM state access key", r.StateKey); err != nil {
		return err
	}
	if r.ProofHash != "" {
		if err := zonestypes.ValidateHash("AVM state access proof hash", r.ProofHash); err != nil {
			return err
		}
	}
	if r.RequestHash == "" {
		return errors.New("AVM state access request hash is required")
	}
	if err := zonestypes.ValidateHash("AVM state access request hash", r.RequestHash); err != nil {
		return err
	}
	if r.RequestHash != ComputeAVMStateAccessRequestHash(r) {
		return errors.New("AVM state access request hash mismatch")
	}
	return nil
}

func ValidateAVMStateIsolationAccess(capability AVMZoneAccessCapability, request AVMStateAccessRequest) error {
	if err := capability.Validate(); err != nil {
		return err
	}
	if err := request.Validate(); err != nil {
		return err
	}
	if request.SourceZone != capability.SourceZone {
		return errors.New("AVM state access source zone capability mismatch")
	}
	if request.SourceZone != request.TargetZone && request.Mode == AVMStateAccessWrite {
		return errors.New("AVM cross-zone writes are prohibited")
	}
	switch request.Target {
	case AVMStateAccessTargetZone:
		expected := AVMZoneRuntimeConfigKey(request.TargetZone)
		if !strings.HasPrefix(request.StateKey, expected) {
			return fmt.Errorf("AVM zone state access must use isolated prefix %q", expected)
		}
	case AVMStateAccessTargetActor:
		expected := ActorStateKeyPrefix(request.ActorIDOptional)
		if request.ActorIDOptional == "" || !strings.HasPrefix(request.StateKey, expected) {
			return fmt.Errorf("AVM actor state access must use isolated prefix %q", expected)
		}
		if request.ActorIDOptional != capability.ActorIDOptional {
			if request.Mode == AVMStateAccessWrite {
				return errors.New("AVM cross-actor mutable state writes are prohibited")
			}
			if !request.ReadOnlyProof || request.ProofHash == "" {
				return errors.New("AVM cross-actor reads require read-only proof")
			}
		}
	case AVMStateAccessTargetContract:
		expected := AVMContractStorageKey(request.ContractAddress, "")
		if request.ContractAddress == "" || !strings.HasPrefix(request.StateKey, expected) {
			return fmt.Errorf("AVM contract storage access must use isolated prefix %q", expected)
		}
		if capability.ContractAddress != "" && request.ContractAddress != capability.ContractAddress && request.Mode == AVMStateAccessWrite {
			return errors.New("AVM contract storage writes are isolated by contract address")
		}
	}
	return nil
}

func NewAVMSchedulerSafetyCheck(queue AVMZoneQueue, messages []AVMAsyncMessage, height uint64, budget zonestypes.ZoneExecutionBudget, retryBounds []AVMSchedulerRetryBound, requireSenderNonceOrdering bool) (AVMSchedulerSafetyCheck, error) {
	selection, err := SelectAVMZoneQueueWork(queue, messages, height, budget)
	if err != nil {
		return AVMSchedulerSafetyCheck{}, err
	}
	check := AVMSchedulerSafetyCheck{
		ZoneID:				queue.ZoneID,
		Height:				height,
		ReadyMessages:			selection.Ready,
		ExpiredMessages:		selection.Expired,
		Budget:				selection.Budget,
		RetryBounds:			append([]AVMSchedulerRetryBound(nil), retryBounds...),
		RequireSenderNonceOrdering:	requireSenderNonceOrdering,
	}
	for _, msg := range messages {
		msg = canonicalAVMAsyncMessage(msg)
		if height < AVMMessageScheduledHeight(msg) {
			check.RejectedEarlyMessageIDs = append(check.RejectedEarlyMessageIDs, msg.ID)
		}
		if height > msg.ExpiryHeight {
			check.RejectedExpiredExecutionIDs = append(check.RejectedExpiredExecutionIDs, msg.ID)
		}
	}
	check = canonicalAVMSchedulerSafetyCheck(check)
	check.SchedulerCheckHash = ComputeAVMSchedulerSafetyCheckHash(check)
	return check, check.Validate()
}

func (c AVMSchedulerSafetyCheck) Validate() error {
	c = canonicalAVMSchedulerSafetyCheck(c)
	if err := zonestypes.ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if c.Height == 0 {
		return errors.New("AVM scheduler safety height must be positive")
	}
	if err := c.Budget.Validate(); err != nil {
		return err
	}
	if c.Budget.MessagesUsed > c.Budget.MaxMessages || c.Budget.GasUsed > c.Budget.MaxGas {
		return errors.New("AVM scheduler queue processing must be bounded")
	}
	seenReady := map[string]struct{}{}
	for _, msg := range c.ReadyMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if msg.DestinationZone != c.ZoneID {
			return errors.New("AVM scheduler ready message zone mismatch")
		}
		if c.Height < AVMMessageScheduledHeight(msg) {
			return errors.New("AVM delayed messages cannot execute early")
		}
		if c.Height > msg.ExpiryHeight {
			return errors.New("AVM expired messages cannot execute")
		}
		if _, found := seenReady[msg.ID]; found {
			return fmt.Errorf("duplicate AVM scheduler ready message %q", msg.ID)
		}
		seenReady[msg.ID] = struct{}{}
	}
	for _, msg := range c.ExpiredMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
		if c.Height <= msg.ExpiryHeight {
			return errors.New("AVM scheduler expired set contains live message")
		}
	}
	if c.RequireSenderNonceOrdering {
		if err := validateAVMSchedulerSenderNonceOrdering(c.ReadyMessages); err != nil {
			return err
		}
	}
	for _, retry := range c.RetryBounds {
		if err := retry.Validate(); err != nil {
			return err
		}
	}
	if err := validateSortedHashes("AVM scheduler rejected early message ids", c.RejectedEarlyMessageIDs); err != nil {
		return err
	}
	if err := validateSortedHashes("AVM scheduler rejected expired execution ids", c.RejectedExpiredExecutionIDs); err != nil {
		return err
	}
	if c.SchedulerCheckHash == "" {
		return errors.New("AVM scheduler safety hash is required")
	}
	if err := zonestypes.ValidateHash("AVM scheduler safety hash", c.SchedulerCheckHash); err != nil {
		return err
	}
	if c.SchedulerCheckHash != ComputeAVMSchedulerSafetyCheckHash(c) {
		return errors.New("AVM scheduler safety hash mismatch")
	}
	return nil
}

func (r AVMSchedulerRetryBound) Validate() error {
	r.MessageID = strings.TrimSpace(r.MessageID)
	if err := zonestypes.ValidateHash("AVM scheduler retry message id", r.MessageID); err != nil {
		return err
	}
	if r.Attempt == 0 {
		return errors.New("AVM scheduler retry attempt must be positive")
	}
	if r.MaxAttempts == 0 {
		return errors.New("AVM scheduler retry max attempts must be positive")
	}
	if r.Attempt > r.MaxAttempts {
		return errors.New("AVM scheduler retry count is bounded")
	}
	return nil
}

func IsAVMStateAccessMode(mode AVMStateAccessMode) bool {
	return mode == AVMStateAccessRead || mode == AVMStateAccessWrite
}

func IsAVMStateAccessTarget(target AVMStateAccessTarget) bool {
	switch target {
	case AVMStateAccessTargetZone, AVMStateAccessTargetActor, AVMStateAccessTargetContract:
		return true
	default:
		return false
	}
}

func AVMNonceKeeperScope(state AVMReplayNonceState) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimSpace(state.ChainID), state.SourceZone, strings.TrimSpace(state.Sender))
}

func ComputeAVMNonceKeeperRoot(keeper AVMNonceKeeper) string {
	keeper = canonicalAVMNonceKeeper(keeper)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-nonce-keeper-v1")
	writeEngineUint64(h, uint64(len(keeper.States)))
	for _, state := range keeper.States {
		writeEnginePart(h, state.StateHash)
		writeEnginePart(h, AVMNonceKeeperScope(state))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMReplayTombstoneStoreRoot(store AVMReplayTombstoneStore) string {
	store = canonicalAVMReplayTombstoneStore(store)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-replay-tombstone-store-v1")
	writeEngineUint64(h, uint64(len(store.ConsumedTombstones)))
	for _, tombstone := range store.ConsumedTombstones {
		writeEnginePart(h, tombstone.MessageID)
		writeEngineUint64(h, tombstone.ConsumedHeight)
	}
	writeEngineUint64(h, uint64(len(store.ExpiredNonces)))
	for _, tombstone := range store.ExpiredNonces {
		writeEnginePart(h, tombstone.TombstoneHash)
		writeEnginePart(h, AVMExpiredNonceTombstoneSortKey(tombstone))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneAccessCapabilityHash(capability AVMZoneAccessCapability) string {
	capability = canonicalAVMZoneAccessCapability(capability)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-access-capability-v1")
	writeEnginePart(h, string(capability.SourceZone))
	writeEnginePart(h, capability.ActorIDOptional)
	writeEnginePart(h, capability.ContractAddress)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStateAccessRequestHash(request AVMStateAccessRequest) string {
	request = canonicalAVMStateAccessRequest(request)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-state-access-request-v1")
	writeEnginePart(h, string(request.SourceZone))
	writeEnginePart(h, string(request.TargetZone))
	writeEnginePart(h, string(request.Mode))
	writeEnginePart(h, string(request.Target))
	writeEnginePart(h, request.ActorIDOptional)
	writeEnginePart(h, request.ContractAddress)
	writeEnginePart(h, request.StateKey)
	writeEnginePart(h, request.ProofHash)
	writeEngineBool(h, request.ReadOnlyProof)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMSchedulerSafetyCheckHash(check AVMSchedulerSafetyCheck) string {
	check = canonicalAVMSchedulerSafetyCheck(check)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-scheduler-safety-v1")
	writeEnginePart(h, string(check.ZoneID))
	writeEngineUint64(h, check.Height)
	writeEngineBool(h, check.RequireSenderNonceOrdering)
	writeEngineUint64(h, check.Budget.MaxGas)
	writeEngineUint64(h, check.Budget.GasUsed)
	writeEngineUint64(h, uint64(check.Budget.MaxMessages))
	writeEngineUint64(h, uint64(check.Budget.MessagesUsed))
	writeEngineUint64(h, uint64(len(check.ReadyMessages)))
	for _, msg := range check.ReadyMessages {
		writeAVMAsyncMessageParts(h, msg)
	}
	writeEngineUint64(h, uint64(len(check.ExpiredMessages)))
	for _, msg := range check.ExpiredMessages {
		writeAVMAsyncMessageParts(h, msg)
	}
	writeEngineUint64(h, uint64(len(check.RetryBounds)))
	for _, retry := range check.RetryBounds {
		writeEnginePart(h, retry.MessageID)
		writeEngineUint64(h, uint64(retry.Attempt))
		writeEngineUint64(h, uint64(retry.MaxAttempts))
	}
	writeEngineUint64(h, uint64(len(check.RejectedEarlyMessageIDs)))
	for _, id := range check.RejectedEarlyMessageIDs {
		writeEnginePart(h, id)
	}
	writeEngineUint64(h, uint64(len(check.RejectedExpiredExecutionIDs)))
	for _, id := range check.RejectedExpiredExecutionIDs {
		writeEnginePart(h, id)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMNonceKeeper(keeper AVMNonceKeeper) AVMNonceKeeper {
	keeper.States = append([]AVMReplayNonceState(nil), keeper.States...)
	for i := range keeper.States {
		keeper.States[i] = canonicalAVMReplayNonceState(keeper.States[i])
	}
	sort.Slice(keeper.States, func(i, j int) bool {
		return AVMNonceKeeperScope(keeper.States[i]) < AVMNonceKeeperScope(keeper.States[j])
	})
	keeper.KeeperRoot = strings.TrimSpace(keeper.KeeperRoot)
	return keeper
}

func canonicalAVMReplayTombstoneStore(store AVMReplayTombstoneStore) AVMReplayTombstoneStore {
	store.ConsumedTombstones = append([]AVMAsyncReplayTombstone(nil), store.ConsumedTombstones...)
	for i := range store.ConsumedTombstones {
		store.ConsumedTombstones[i].MessageID = strings.TrimSpace(store.ConsumedTombstones[i].MessageID)
	}
	sort.Slice(store.ConsumedTombstones, func(i, j int) bool {
		return store.ConsumedTombstones[i].MessageID < store.ConsumedTombstones[j].MessageID
	})
	store.ExpiredNonces = append([]AVMExpiredNonceTombstone(nil), store.ExpiredNonces...)
	for i := range store.ExpiredNonces {
		store.ExpiredNonces[i] = canonicalAVMExpiredNonceTombstone(store.ExpiredNonces[i])
	}
	sort.Slice(store.ExpiredNonces, func(i, j int) bool {
		return AVMExpiredNonceTombstoneSortKey(store.ExpiredNonces[i]) < AVMExpiredNonceTombstoneSortKey(store.ExpiredNonces[j])
	})
	store.StoreRoot = strings.TrimSpace(store.StoreRoot)
	return store
}

func canonicalAVMZoneAccessCapability(capability AVMZoneAccessCapability) AVMZoneAccessCapability {
	capability.ActorIDOptional = strings.TrimSpace(capability.ActorIDOptional)
	capability.ContractAddress = strings.TrimSpace(capability.ContractAddress)
	capability.CapabilityHash = strings.TrimSpace(capability.CapabilityHash)
	return capability
}

func canonicalAVMStateAccessRequest(request AVMStateAccessRequest) AVMStateAccessRequest {
	request.ActorIDOptional = strings.TrimSpace(request.ActorIDOptional)
	request.ContractAddress = strings.TrimSpace(request.ContractAddress)
	request.StateKey = strings.TrimSpace(request.StateKey)
	request.ProofHash = strings.TrimSpace(request.ProofHash)
	request.RequestHash = strings.TrimSpace(request.RequestHash)
	return request
}

func canonicalAVMSchedulerSafetyCheck(check AVMSchedulerSafetyCheck) AVMSchedulerSafetyCheck {
	check.ReadyMessages = append([]AVMAsyncMessage(nil), check.ReadyMessages...)
	for i := range check.ReadyMessages {
		check.ReadyMessages[i] = canonicalAVMAsyncMessage(check.ReadyMessages[i])
	}
	check.ExpiredMessages = append([]AVMAsyncMessage(nil), check.ExpiredMessages...)
	for i := range check.ExpiredMessages {
		check.ExpiredMessages[i] = canonicalAVMAsyncMessage(check.ExpiredMessages[i])
	}
	check.RetryBounds = append([]AVMSchedulerRetryBound(nil), check.RetryBounds...)
	for i := range check.RetryBounds {
		check.RetryBounds[i].MessageID = strings.TrimSpace(check.RetryBounds[i].MessageID)
	}
	sort.Slice(check.RetryBounds, func(i, j int) bool {
		return check.RetryBounds[i].MessageID < check.RetryBounds[j].MessageID
	})
	check.RejectedEarlyMessageIDs = trimSortStrings(check.RejectedEarlyMessageIDs)
	check.RejectedExpiredExecutionIDs = trimSortStrings(check.RejectedExpiredExecutionIDs)
	check.SchedulerCheckHash = strings.TrimSpace(check.SchedulerCheckHash)
	return check
}

func validateAVMSchedulerSenderNonceOrdering(messages []AVMAsyncMessage) error {
	lastNonce := map[string]uint64{}
	for _, msg := range messages {
		scope := fmt.Sprintf("%s/%s", msg.SourceZone, strings.TrimSpace(msg.Source))
		nonce, found := lastNonce[scope]
		if found && msg.SenderNonce <= nonce {
			return errors.New("AVM priority cannot bypass sender nonce ordering")
		}
		lastNonce[scope] = msg.SenderNonce
	}
	return nil
}

func validateSortedHashes(fieldName string, values []string) error {
	for i, value := range values {
		if err := zonestypes.ValidateHash(fieldName, value); err != nil {
			return err
		}
		if i > 0 && values[i-1] >= value {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
	}
	return nil
}

func trimSortStrings(values []string) []string {
	out := append([]string(nil), values...)
	for i := range out {
		out[i] = strings.TrimSpace(out[i])
	}
	sort.Strings(out)
	return out
}
