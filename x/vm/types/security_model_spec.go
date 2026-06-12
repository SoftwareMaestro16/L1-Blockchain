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
	AVMSecurityAssumptionCometBFTFinality		AVMSecurityAssumption	= "cometbft_finality"
	AVMSecurityAssumptionDeterministicExecution	AVMSecurityAssumption	= "deterministic_execution"
	AVMSecurityAssumptionNoExternalCalls		AVMSecurityAssumption	= "no_external_calls"
	AVMSecurityAssumptionGasBounding		AVMSecurityAssumption	= "gas_bounding"
	AVMSecurityAssumptionReplayProtection		AVMSecurityAssumption	= "replay_protection"
	AVMSecurityAssumptionZoneStateIsolation		AVMSecurityAssumption	= "zone_state_isolation"
	AVMSecurityAssumptionActorStateIsolation	AVMSecurityAssumption	= "actor_state_isolation"
	AVMSecurityAssumptionStoreRootCommitments	AVMSecurityAssumption	= "store_root_commitments"
)

type AVMSecurityAssumption string

type AVMSecurityModel struct {
	Assumptions		[]AVMSecurityAssumption
	StoreRootCommitment	string
	ModelHash		string
}

type AVMReplayProtectionRecord struct {
	ChainID		string
	SourceZone	zonestypes.ZoneID
	Sender		string
	SenderNonce	uint64
	MessageID	string
	CreatedHeight	uint64
	ExpiryHeight	uint64
	DestinationZone	zonestypes.ZoneID
	Destination	string
	PayloadHash	string
	RecordHash	string
}

type AVMExpiredNonceTombstone struct {
	ChainID		string
	SourceZone	zonestypes.ZoneID
	Sender		string
	SenderNonce	uint64
	MessageID	string
	ExpiryHeight	uint64
	TombstoneHash	string
}

type AVMReplayNonceState struct {
	ChainID			string
	SourceZone		zonestypes.ZoneID
	Sender			string
	LastNonce		uint64
	ConsumedTombstones	[]AVMAsyncReplayTombstone
	ExpiredNonceRecords	[]AVMExpiredNonceTombstone
	StateHash		string
}

func DefaultAVMSecurityModel(storeRootCommitment string) (AVMSecurityModel, error) {
	model := AVMSecurityModel{
		Assumptions:		AllAVMSecurityAssumptions(),
		StoreRootCommitment:	strings.TrimSpace(storeRootCommitment),
	}
	model.ModelHash = ComputeAVMSecurityModelHash(model)
	return model, model.Validate()
}

func (m AVMSecurityModel) Validate() error {
	m = canonicalAVMSecurityModel(m)
	if err := zonestypes.ValidateHash("AVM security store root commitment", m.StoreRootCommitment); err != nil {
		return err
	}
	required := AllAVMSecurityAssumptions()
	if len(m.Assumptions) != len(required) {
		return errors.New("AVM security model must declare every security assumption")
	}
	for i, assumption := range m.Assumptions {
		if !IsAVMSecurityAssumption(assumption) {
			return fmt.Errorf("invalid AVM security assumption %q", assumption)
		}
		if i > 0 && m.Assumptions[i-1] >= assumption {
			return errors.New("AVM security assumptions must be sorted canonically")
		}
	}
	requiredSet := make(map[AVMSecurityAssumption]struct{}, len(required))
	for _, assumption := range required {
		requiredSet[assumption] = struct{}{}
	}
	for _, assumption := range m.Assumptions {
		delete(requiredSet, assumption)
	}
	if len(requiredSet) != 0 {
		return errors.New("AVM security model is missing required assumptions")
	}
	if m.ModelHash == "" {
		return errors.New("AVM security model hash is required")
	}
	if err := zonestypes.ValidateHash("AVM security model hash", m.ModelHash); err != nil {
		return err
	}
	if m.ModelHash != ComputeAVMSecurityModelHash(m) {
		return errors.New("AVM security model hash mismatch")
	}
	return nil
}

func NewAVMReplayProtectionRecord(msg AVMAsyncMessage) (AVMReplayProtectionRecord, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMReplayProtectionRecord{}, err
	}
	record := AVMReplayProtectionRecord{
		ChainID:		msg.ChainID,
		SourceZone:		msg.SourceZone,
		Sender:			msg.Source,
		SenderNonce:		msg.SenderNonce,
		MessageID:		msg.ID,
		CreatedHeight:		msg.CreatedHeight,
		ExpiryHeight:		msg.ExpiryHeight,
		DestinationZone:	msg.DestinationZone,
		Destination:		msg.Destination,
		PayloadHash:		msg.PayloadHash,
	}
	record = canonicalAVMReplayProtectionRecord(record)
	record.RecordHash = ComputeAVMReplayProtectionRecordHash(record)
	return record, record.Validate()
}

func (r AVMReplayProtectionRecord) Validate() error {
	r = canonicalAVMReplayProtectionRecord(r)
	if err := validateRouterOptionalToken("AVM replay chain id", r.ChainID, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if r.ChainID == "" {
		return errors.New("AVM replay chain id is required")
	}
	if err := zonestypes.ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM replay sender", r.Sender, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if r.Sender == "" {
		return errors.New("AVM replay sender is required")
	}
	if r.SenderNonce == 0 {
		return errors.New("AVM replay sender nonce must be positive")
	}
	if err := zonestypes.ValidateHash("AVM replay message id", r.MessageID); err != nil {
		return err
	}
	if r.CreatedHeight == 0 {
		return errors.New("AVM replay created height must be positive")
	}
	if r.ExpiryHeight <= r.CreatedHeight {
		return errors.New("AVM replay expiry height must be after creation")
	}
	if err := zonestypes.ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM replay destination", r.Destination, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if r.Destination == "" {
		return errors.New("AVM replay destination is required")
	}
	if err := zonestypes.ValidateHash("AVM replay payload hash", r.PayloadHash); err != nil {
		return err
	}
	if r.RecordHash == "" {
		return errors.New("AVM replay record hash is required")
	}
	if err := zonestypes.ValidateHash("AVM replay record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMReplayProtectionRecordHash(r) {
		return errors.New("AVM replay record hash mismatch")
	}
	return nil
}

func NewAVMReplayNonceState(state AVMReplayNonceState) (AVMReplayNonceState, error) {
	state = canonicalAVMReplayNonceState(state)
	state.StateHash = ComputeAVMReplayNonceStateHash(state)
	return state, state.Validate()
}

func (s AVMReplayNonceState) Validate() error {
	s = canonicalAVMReplayNonceState(s)
	if err := validateRouterOptionalToken("AVM replay state chain id", s.ChainID, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if s.ChainID == "" {
		return errors.New("AVM replay state chain id is required")
	}
	if err := zonestypes.ValidateZoneID(s.SourceZone); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM replay state sender", s.Sender, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if s.Sender == "" {
		return errors.New("AVM replay state sender is required")
	}
	seenConsumed := make(map[string]struct{}, len(s.ConsumedTombstones))
	for i, tombstone := range s.ConsumedTombstones {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		if _, found := seenConsumed[tombstone.MessageID]; found {
			return fmt.Errorf("duplicate AVM replay consumed tombstone %q", tombstone.MessageID)
		}
		seenConsumed[tombstone.MessageID] = struct{}{}
		if i > 0 && s.ConsumedTombstones[i-1].MessageID >= tombstone.MessageID {
			return errors.New("AVM replay consumed tombstones must be sorted canonically")
		}
	}
	seenExpired := make(map[string]struct{}, len(s.ExpiredNonceRecords))
	for i, tombstone := range s.ExpiredNonceRecords {
		if err := tombstone.Validate(); err != nil {
			return err
		}
		if tombstone.ChainID != s.ChainID || tombstone.SourceZone != s.SourceZone || tombstone.Sender != s.Sender {
			return errors.New("AVM replay expired nonce tombstone scope mismatch")
		}
		scope := AVMReplayNonceScope(tombstone.ChainID, tombstone.SourceZone, tombstone.Sender, tombstone.SenderNonce)
		if _, found := seenExpired[scope]; found {
			return fmt.Errorf("duplicate AVM replay expired nonce scope %q", scope)
		}
		seenExpired[scope] = struct{}{}
		if i > 0 && AVMExpiredNonceTombstoneSortKey(s.ExpiredNonceRecords[i-1]) >= AVMExpiredNonceTombstoneSortKey(tombstone) {
			return errors.New("AVM replay expired nonce tombstones must be sorted canonically")
		}
	}
	if s.StateHash == "" {
		return errors.New("AVM replay nonce state hash is required")
	}
	if err := zonestypes.ValidateHash("AVM replay nonce state hash", s.StateHash); err != nil {
		return err
	}
	if s.StateHash != ComputeAVMReplayNonceStateHash(s) {
		return errors.New("AVM replay nonce state hash mismatch")
	}
	return nil
}

func (t AVMExpiredNonceTombstone) Validate() error {
	t = canonicalAVMExpiredNonceTombstone(t)
	if err := validateRouterOptionalToken("AVM expired nonce chain id", t.ChainID, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if t.ChainID == "" {
		return errors.New("AVM expired nonce chain id is required")
	}
	if err := zonestypes.ValidateZoneID(t.SourceZone); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM expired nonce sender", t.Sender, MaxAsyncMessageEndpointLength); err != nil {
		return err
	}
	if t.Sender == "" {
		return errors.New("AVM expired nonce sender is required")
	}
	if t.SenderNonce == 0 {
		return errors.New("AVM expired nonce must be positive")
	}
	if err := zonestypes.ValidateHash("AVM expired nonce message id", t.MessageID); err != nil {
		return err
	}
	if t.ExpiryHeight == 0 {
		return errors.New("AVM expired nonce expiry height must be positive")
	}
	if t.TombstoneHash == "" {
		return errors.New("AVM expired nonce tombstone hash is required")
	}
	if err := zonestypes.ValidateHash("AVM expired nonce tombstone hash", t.TombstoneHash); err != nil {
		return err
	}
	if t.TombstoneHash != ComputeAVMExpiredNonceTombstoneHash(t) {
		return errors.New("AVM expired nonce tombstone hash mismatch")
	}
	return nil
}

func ValidateAVMReplaySubmission(state AVMReplayNonceState, msg AVMAsyncMessage, currentHeight uint64) error {
	state = canonicalAVMReplayNonceState(state)
	if err := state.Validate(); err != nil {
		return err
	}
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return err
	}
	if currentHeight == 0 {
		return errors.New("AVM replay current height must be positive")
	}
	if msg.ChainID != state.ChainID || msg.SourceZone != state.SourceZone || msg.Source != state.Sender {
		return errors.New("AVM replay submission scope mismatch")
	}
	if currentHeight > msg.ExpiryHeight {
		return errors.New("AVM replay submission message is expired")
	}
	for _, tombstone := range state.ConsumedTombstones {
		if tombstone.MessageID == msg.ID {
			return errors.New("AVM replay consumed message tombstone blocks replay")
		}
	}
	nonceScope := AVMReplayNonceScope(msg.ChainID, msg.SourceZone, msg.Source, msg.SenderNonce)
	for _, tombstone := range state.ExpiredNonceRecords {
		if AVMReplayNonceScope(tombstone.ChainID, tombstone.SourceZone, tombstone.Sender, tombstone.SenderNonce) == nonceScope {
			return errors.New("AVM replay expired nonce tombstone blocks resubmission")
		}
	}
	if msg.SenderNonce <= state.LastNonce {
		return errors.New("AVM replay sender nonce must increase")
	}
	return ValidateAVMCrossZoneReplayBinding(msg)
}

func ConsumeAVMReplayMessage(state AVMReplayNonceState, msg AVMAsyncMessage, consumedHeight uint64) (AVMReplayNonceState, AVMAsyncReplayTombstone, error) {
	if consumedHeight == 0 {
		return AVMReplayNonceState{}, AVMAsyncReplayTombstone{}, errors.New("AVM replay consumed height must be positive")
	}
	if err := ValidateAVMReplaySubmission(state, msg, consumedHeight); err != nil {
		return AVMReplayNonceState{}, AVMAsyncReplayTombstone{}, err
	}
	tombstone := AVMAsyncReplayTombstone{MessageID: strings.TrimSpace(msg.ID), ConsumedHeight: consumedHeight}
	state.LastNonce = msg.SenderNonce
	state.ConsumedTombstones = append(state.ConsumedTombstones, tombstone)
	state = canonicalAVMReplayNonceState(state)
	state.StateHash = ComputeAVMReplayNonceStateHash(state)
	return state, tombstone, state.Validate()
}

func ExpireAVMReplayMessage(state AVMReplayNonceState, msg AVMAsyncMessage, expiredHeight uint64) (AVMReplayNonceState, AVMExpiredNonceTombstone, error) {
	if expiredHeight == 0 {
		return AVMReplayNonceState{}, AVMExpiredNonceTombstone{}, errors.New("AVM replay expired height must be positive")
	}
	state = canonicalAVMReplayNonceState(state)
	if err := state.Validate(); err != nil {
		return AVMReplayNonceState{}, AVMExpiredNonceTombstone{}, err
	}
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMReplayNonceState{}, AVMExpiredNonceTombstone{}, err
	}
	if msg.ChainID != state.ChainID || msg.SourceZone != state.SourceZone || msg.Source != state.Sender {
		return AVMReplayNonceState{}, AVMExpiredNonceTombstone{}, errors.New("AVM replay expiration scope mismatch")
	}
	if expiredHeight <= msg.ExpiryHeight {
		return AVMReplayNonceState{}, AVMExpiredNonceTombstone{}, errors.New("AVM replay expiration height must be after message expiry")
	}
	tombstone := AVMExpiredNonceTombstone{
		ChainID:	msg.ChainID,
		SourceZone:	msg.SourceZone,
		Sender:		msg.Source,
		SenderNonce:	msg.SenderNonce,
		MessageID:	msg.ID,
		ExpiryHeight:	msg.ExpiryHeight,
	}
	tombstone = canonicalAVMExpiredNonceTombstone(tombstone)
	tombstone.TombstoneHash = ComputeAVMExpiredNonceTombstoneHash(tombstone)
	if err := tombstone.Validate(); err != nil {
		return AVMReplayNonceState{}, AVMExpiredNonceTombstone{}, err
	}
	if msg.SenderNonce > state.LastNonce {
		state.LastNonce = msg.SenderNonce
	}
	state.ExpiredNonceRecords = append(state.ExpiredNonceRecords, tombstone)
	state = canonicalAVMReplayNonceState(state)
	state.StateHash = ComputeAVMReplayNonceStateHash(state)
	return state, tombstone, state.Validate()
}

func ValidateAVMCrossZoneReplayBinding(msg AVMAsyncMessage) error {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return err
	}
	if msg.SourceZone == msg.DestinationZone {
		return nil
	}
	if msg.ID != DeriveAVMAsyncMessageID(msg) {
		return errors.New("AVM cross-zone replay message id must bind source and destination zones")
	}
	if msg.Source == msg.Destination {
		return errors.New("AVM cross-zone replay source and destination endpoints must differ")
	}
	return nil
}

func AllAVMSecurityAssumptions() []AVMSecurityAssumption {
	items := []AVMSecurityAssumption{
		AVMSecurityAssumptionCometBFTFinality,
		AVMSecurityAssumptionDeterministicExecution,
		AVMSecurityAssumptionNoExternalCalls,
		AVMSecurityAssumptionGasBounding,
		AVMSecurityAssumptionReplayProtection,
		AVMSecurityAssumptionZoneStateIsolation,
		AVMSecurityAssumptionActorStateIsolation,
		AVMSecurityAssumptionStoreRootCommitments,
	}
	sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })
	return items
}

func IsAVMSecurityAssumption(assumption AVMSecurityAssumption) bool {
	switch assumption {
	case AVMSecurityAssumptionCometBFTFinality,
		AVMSecurityAssumptionDeterministicExecution,
		AVMSecurityAssumptionNoExternalCalls,
		AVMSecurityAssumptionGasBounding,
		AVMSecurityAssumptionReplayProtection,
		AVMSecurityAssumptionZoneStateIsolation,
		AVMSecurityAssumptionActorStateIsolation,
		AVMSecurityAssumptionStoreRootCommitments:
		return true
	default:
		return false
	}
}

func AVMReplayNonceScope(chainID string, sourceZone zonestypes.ZoneID, sender string, senderNonce uint64) string {
	return fmt.Sprintf("%s/%s/%s/%020d", strings.TrimSpace(chainID), sourceZone, strings.TrimSpace(sender), senderNonce)
}

func AVMExpiredNonceTombstoneSortKey(tombstone AVMExpiredNonceTombstone) string {
	tombstone = canonicalAVMExpiredNonceTombstone(tombstone)
	return fmt.Sprintf("%s/%020d/%s", AVMReplayNonceScope(tombstone.ChainID, tombstone.SourceZone, tombstone.Sender, tombstone.SenderNonce), tombstone.ExpiryHeight, tombstone.MessageID)
}

func ComputeAVMSecurityModelHash(model AVMSecurityModel) string {
	model = canonicalAVMSecurityModel(model)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-security-model-v1")
	writeEnginePart(h, model.StoreRootCommitment)
	writeEngineUint64(h, uint64(len(model.Assumptions)))
	for _, assumption := range model.Assumptions {
		writeEnginePart(h, string(assumption))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMReplayProtectionRecordHash(record AVMReplayProtectionRecord) string {
	record = canonicalAVMReplayProtectionRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-replay-record-v1")
	writeEnginePart(h, record.ChainID)
	writeEnginePart(h, string(record.SourceZone))
	writeEnginePart(h, record.Sender)
	writeEngineUint64(h, record.SenderNonce)
	writeEnginePart(h, record.MessageID)
	writeEngineUint64(h, record.CreatedHeight)
	writeEngineUint64(h, record.ExpiryHeight)
	writeEnginePart(h, string(record.DestinationZone))
	writeEnginePart(h, record.Destination)
	writeEnginePart(h, record.PayloadHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMExpiredNonceTombstoneHash(tombstone AVMExpiredNonceTombstone) string {
	tombstone = canonicalAVMExpiredNonceTombstone(tombstone)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-expired-nonce-tombstone-v1")
	writeEnginePart(h, tombstone.ChainID)
	writeEnginePart(h, string(tombstone.SourceZone))
	writeEnginePart(h, tombstone.Sender)
	writeEngineUint64(h, tombstone.SenderNonce)
	writeEnginePart(h, tombstone.MessageID)
	writeEngineUint64(h, tombstone.ExpiryHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMReplayNonceStateHash(state AVMReplayNonceState) string {
	state = canonicalAVMReplayNonceState(state)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-replay-nonce-state-v1")
	writeEnginePart(h, state.ChainID)
	writeEnginePart(h, string(state.SourceZone))
	writeEnginePart(h, state.Sender)
	writeEngineUint64(h, state.LastNonce)
	writeEngineUint64(h, uint64(len(state.ConsumedTombstones)))
	for _, tombstone := range state.ConsumedTombstones {
		writeEnginePart(h, tombstone.MessageID)
		writeEngineUint64(h, tombstone.ConsumedHeight)
	}
	writeEngineUint64(h, uint64(len(state.ExpiredNonceRecords)))
	for _, tombstone := range state.ExpiredNonceRecords {
		writeEnginePart(h, tombstone.TombstoneHash)
		writeEnginePart(h, AVMExpiredNonceTombstoneSortKey(tombstone))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMSecurityModel(model AVMSecurityModel) AVMSecurityModel {
	model.StoreRootCommitment = strings.TrimSpace(model.StoreRootCommitment)
	model.Assumptions = append([]AVMSecurityAssumption(nil), model.Assumptions...)
	sort.Slice(model.Assumptions, func(i, j int) bool { return model.Assumptions[i] < model.Assumptions[j] })
	model.ModelHash = strings.TrimSpace(model.ModelHash)
	return model
}

func canonicalAVMReplayProtectionRecord(record AVMReplayProtectionRecord) AVMReplayProtectionRecord {
	record.ChainID = strings.TrimSpace(record.ChainID)
	record.Sender = strings.TrimSpace(record.Sender)
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.Destination = strings.TrimSpace(record.Destination)
	record.PayloadHash = strings.TrimSpace(record.PayloadHash)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMReplayNonceState(state AVMReplayNonceState) AVMReplayNonceState {
	state.ChainID = strings.TrimSpace(state.ChainID)
	state.Sender = strings.TrimSpace(state.Sender)
	state.ConsumedTombstones = append([]AVMAsyncReplayTombstone(nil), state.ConsumedTombstones...)
	for i := range state.ConsumedTombstones {
		state.ConsumedTombstones[i].MessageID = strings.TrimSpace(state.ConsumedTombstones[i].MessageID)
	}
	sort.Slice(state.ConsumedTombstones, func(i, j int) bool {
		return state.ConsumedTombstones[i].MessageID < state.ConsumedTombstones[j].MessageID
	})
	state.ExpiredNonceRecords = append([]AVMExpiredNonceTombstone(nil), state.ExpiredNonceRecords...)
	for i := range state.ExpiredNonceRecords {
		state.ExpiredNonceRecords[i] = canonicalAVMExpiredNonceTombstone(state.ExpiredNonceRecords[i])
	}
	sort.Slice(state.ExpiredNonceRecords, func(i, j int) bool {
		return AVMExpiredNonceTombstoneSortKey(state.ExpiredNonceRecords[i]) < AVMExpiredNonceTombstoneSortKey(state.ExpiredNonceRecords[j])
	})
	state.StateHash = strings.TrimSpace(state.StateHash)
	return state
}

func canonicalAVMExpiredNonceTombstone(tombstone AVMExpiredNonceTombstone) AVMExpiredNonceTombstone {
	tombstone.ChainID = strings.TrimSpace(tombstone.ChainID)
	tombstone.Sender = strings.TrimSpace(tombstone.Sender)
	tombstone.MessageID = strings.TrimSpace(tombstone.MessageID)
	tombstone.TombstoneHash = strings.TrimSpace(tombstone.TombstoneHash)
	return tombstone
}
