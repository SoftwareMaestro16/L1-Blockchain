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
	AVMStoreV2PayloadInlineMaxBytes		= 256
	AVMStoreV2CompactMessageMaxBytes	= 512
)

type AVMStoreV2MessageRecord struct {
	MessageID	string
	MessageKey	string
	PayloadHash	string
	PayloadInline	bool
	PayloadRefKey	string
	CompactBytes	uint32
	RecordHash	string
}

type AVMStoreV2PayloadRecord struct {
	PayloadHash	string
	PayloadKey	string
	PayloadSize	uint32
	Inline		bool
	RecordHash	string
}

type AVMStoreV2DelayedQueueBucket struct {
	ZoneID		zonestypes.ZoneID
	Height		uint64
	BucketKey	string
	MessageIDs	[]string
	BucketHash	string
}

type AVMStoreV2ActorStatePrefix struct {
	ActorID		string
	Prefix		string
	PrefixHash	string
}

type AVMStoreV2ContractStatePrefix struct {
	ContractAddress	string
	Prefix		string
	PrefixHash	string
}

type AVMStoreV2TombstonePruningPlan struct {
	CurrentHeight		uint64
	ProofHorizon		uint64
	RetainAfterHeight	uint64
	PrunableConsumedIDs	[]string
	PrunableExpiredScopes	[]string
	PlanHash		string
}

type AVMStoreV2LayoutStrategy struct {
	MessageRecords		[]AVMStoreV2MessageRecord
	PayloadRecords		[]AVMStoreV2PayloadRecord
	ActorPrefixes		[]AVMStoreV2ActorStatePrefix
	ContractPrefixes	[]AVMStoreV2ContractStatePrefix
	DelayedBuckets		[]AVMStoreV2DelayedQueueBucket
	PruningPlan		AVMStoreV2TombstonePruningPlan
	LayoutRoot		string
}

func NewAVMStoreV2MessageRecord(msg AVMAsyncMessage) (AVMStoreV2MessageRecord, AVMStoreV2PayloadRecord, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMStoreV2MessageRecord{}, AVMStoreV2PayloadRecord{}, err
	}
	inline := len(msg.Payload) <= AVMStoreV2PayloadInlineMaxBytes
	payload := AVMStoreV2PayloadRecord{
		PayloadHash:	msg.PayloadHash,
		PayloadKey:	AVMStoreV2PayloadKey(msg.PayloadHash),
		PayloadSize:	uint32(len(msg.Payload)),
		Inline:		inline,
	}
	payload.RecordHash = ComputeAVMStoreV2PayloadRecordHash(payload)
	if err := payload.Validate(); err != nil {
		return AVMStoreV2MessageRecord{}, AVMStoreV2PayloadRecord{}, err
	}
	record := AVMStoreV2MessageRecord{
		MessageID:	msg.ID,
		MessageKey:	AVMAsyncMessageKey(msg.ID),
		PayloadHash:	msg.PayloadHash,
		PayloadInline:	inline,
		CompactBytes:	AVMStoreV2CompactMessageBytes(msg),
	}
	if !inline {
		record.PayloadRefKey = payload.PayloadKey
	}
	record.RecordHash = ComputeAVMStoreV2MessageRecordHash(record)
	return record, payload, record.Validate()
}

func (r AVMStoreV2MessageRecord) Validate() error {
	r = canonicalAVMStoreV2MessageRecord(r)
	if err := zonestypes.ValidateHash("AVM Store v2 message id", r.MessageID); err != nil {
		return err
	}
	if r.MessageKey != AVMAsyncMessageKey(r.MessageID) {
		return errors.New("AVM Store v2 message key mismatch")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 payload hash", r.PayloadHash); err != nil {
		return err
	}
	if r.CompactBytes == 0 || r.CompactBytes > AVMStoreV2CompactMessageMaxBytes {
		return fmt.Errorf("AVM Store v2 compact message record must be 1..%d bytes", AVMStoreV2CompactMessageMaxBytes)
	}
	if r.PayloadInline && r.PayloadRefKey != "" {
		return errors.New("AVM Store v2 inline payload must not store external payload ref")
	}
	if !r.PayloadInline && r.PayloadRefKey != AVMStoreV2PayloadKey(r.PayloadHash) {
		return errors.New("AVM Store v2 payload must be stored by hash key")
	}
	if r.RecordHash == "" {
		return errors.New("AVM Store v2 message record hash is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 message record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMStoreV2MessageRecordHash(r) {
		return errors.New("AVM Store v2 message record hash mismatch")
	}
	return nil
}

func (r AVMStoreV2PayloadRecord) Validate() error {
	r = canonicalAVMStoreV2PayloadRecord(r)
	if err := zonestypes.ValidateHash("AVM Store v2 payload hash", r.PayloadHash); err != nil {
		return err
	}
	if r.PayloadKey != AVMStoreV2PayloadKey(r.PayloadHash) {
		return errors.New("AVM Store v2 payload key must be hash-addressed")
	}
	if r.Inline && r.PayloadSize > AVMStoreV2PayloadInlineMaxBytes {
		return errors.New("AVM Store v2 inline payload exceeds inline threshold")
	}
	if !r.Inline && r.PayloadSize <= AVMStoreV2PayloadInlineMaxBytes {
		return errors.New("AVM Store v2 external payload should only be used above inline threshold")
	}
	if r.RecordHash == "" {
		return errors.New("AVM Store v2 payload record hash is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 payload record hash", r.RecordHash); err != nil {
		return err
	}
	if r.RecordHash != ComputeAVMStoreV2PayloadRecordHash(r) {
		return errors.New("AVM Store v2 payload record hash mismatch")
	}
	return nil
}

func NewAVMStoreV2ActorStatePrefix(actorID string) (AVMStoreV2ActorStatePrefix, error) {
	prefix := AVMStoreV2ActorStatePrefix{
		ActorID:	strings.TrimSpace(actorID),
		Prefix:		ActorStateKeyPrefix(actorID),
	}
	prefix.PrefixHash = ComputeAVMStoreV2ActorStatePrefixHash(prefix)
	return prefix, prefix.Validate()
}

func (p AVMStoreV2ActorStatePrefix) Validate() error {
	p = canonicalAVMStoreV2ActorStatePrefix(p)
	if err := validateEngineToken("AVM Store v2 actor id", p.ActorID, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if p.Prefix != ActorStateKeyPrefix(p.ActorID) {
		return errors.New("AVM Store v2 actor state prefix must be scoped by actor id")
	}
	if p.PrefixHash == "" {
		return errors.New("AVM Store v2 actor prefix hash is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 actor prefix hash", p.PrefixHash); err != nil {
		return err
	}
	if p.PrefixHash != ComputeAVMStoreV2ActorStatePrefixHash(p) {
		return errors.New("AVM Store v2 actor prefix hash mismatch")
	}
	return nil
}

func NewAVMStoreV2ContractStatePrefix(contractAddress string) (AVMStoreV2ContractStatePrefix, error) {
	prefix := AVMStoreV2ContractStatePrefix{
		ContractAddress:	strings.TrimSpace(contractAddress),
		Prefix:			AVMStatePrefixContractStorage + "/" + strings.TrimSpace(contractAddress) + "/",
	}
	prefix.PrefixHash = ComputeAVMStoreV2ContractStatePrefixHash(prefix)
	return prefix, prefix.Validate()
}

func (p AVMStoreV2ContractStatePrefix) Validate() error {
	p = canonicalAVMStoreV2ContractStatePrefix(p)
	if err := validateEngineToken("AVM Store v2 contract address", p.ContractAddress, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	expected := AVMStatePrefixContractStorage + "/" + p.ContractAddress + "/"
	if p.Prefix != expected {
		return errors.New("AVM Store v2 contract state prefix must be scoped by contract address")
	}
	if err := validateAVMStatePrefix("AVM Store v2 contract prefix", strings.TrimSuffix(p.Prefix, "/")); err != nil {
		return err
	}
	if p.PrefixHash == "" {
		return errors.New("AVM Store v2 contract prefix hash is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 contract prefix hash", p.PrefixHash); err != nil {
		return err
	}
	if p.PrefixHash != ComputeAVMStoreV2ContractStatePrefixHash(p) {
		return errors.New("AVM Store v2 contract prefix hash mismatch")
	}
	return nil
}

func NewAVMStoreV2DelayedQueueBucket(zoneID zonestypes.ZoneID, height uint64, messageIDs []string) (AVMStoreV2DelayedQueueBucket, error) {
	bucket := AVMStoreV2DelayedQueueBucket{
		ZoneID:		zoneID,
		Height:		height,
		BucketKey:	AVMStoreV2DelayedQueueBucketKey(zoneID, height),
		MessageIDs:	trimSortStrings(messageIDs),
	}
	bucket.BucketHash = ComputeAVMStoreV2DelayedQueueBucketHash(bucket)
	return bucket, bucket.Validate()
}

func (b AVMStoreV2DelayedQueueBucket) Validate() error {
	b = canonicalAVMStoreV2DelayedQueueBucket(b)
	if err := zonestypes.ValidateZoneID(b.ZoneID); err != nil {
		return err
	}
	if b.Height == 0 {
		return errors.New("AVM Store v2 delayed queue bucket height must be positive")
	}
	if b.BucketKey != AVMStoreV2DelayedQueueBucketKey(b.ZoneID, b.Height) {
		return errors.New("AVM Store v2 delayed queue bucket key mismatch")
	}
	if len(b.MessageIDs) == 0 {
		return errors.New("AVM Store v2 delayed queue bucket must contain messages")
	}
	if err := validateSortedHashes("AVM Store v2 delayed queue message id", b.MessageIDs); err != nil {
		return err
	}
	if b.BucketHash == "" {
		return errors.New("AVM Store v2 delayed queue bucket hash is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 delayed queue bucket hash", b.BucketHash); err != nil {
		return err
	}
	if b.BucketHash != ComputeAVMStoreV2DelayedQueueBucketHash(b) {
		return errors.New("AVM Store v2 delayed queue bucket hash mismatch")
	}
	return nil
}

func NewAVMStoreV2TombstonePruningPlan(store AVMReplayTombstoneStore, currentHeight, proofHorizon uint64) (AVMStoreV2TombstonePruningPlan, error) {
	store = canonicalAVMReplayTombstoneStore(store)
	if err := store.Validate(); err != nil {
		return AVMStoreV2TombstonePruningPlan{}, err
	}
	if currentHeight == 0 {
		return AVMStoreV2TombstonePruningPlan{}, errors.New("AVM Store v2 pruning current height must be positive")
	}
	retainAfter := uint64(0)
	if currentHeight > proofHorizon {
		retainAfter = currentHeight - proofHorizon
	}
	plan := AVMStoreV2TombstonePruningPlan{
		CurrentHeight:		currentHeight,
		ProofHorizon:		proofHorizon,
		RetainAfterHeight:	retainAfter,
	}
	for _, tombstone := range store.ConsumedTombstones {
		if tombstone.ConsumedHeight < retainAfter {
			plan.PrunableConsumedIDs = append(plan.PrunableConsumedIDs, tombstone.MessageID)
		}
	}
	for _, tombstone := range store.ExpiredNonces {
		if tombstone.ExpiryHeight < retainAfter {
			plan.PrunableExpiredScopes = append(plan.PrunableExpiredScopes, AVMReplayNonceScope(tombstone.ChainID, tombstone.SourceZone, tombstone.Sender, tombstone.SenderNonce))
		}
	}
	plan = canonicalAVMStoreV2TombstonePruningPlan(plan)
	plan.PlanHash = ComputeAVMStoreV2TombstonePruningPlanHash(plan)
	return plan, plan.Validate()
}

func (p AVMStoreV2TombstonePruningPlan) Validate() error {
	p = canonicalAVMStoreV2TombstonePruningPlan(p)
	if p.CurrentHeight == 0 {
		return errors.New("AVM Store v2 pruning current height must be positive")
	}
	expectedRetainAfter := uint64(0)
	if p.CurrentHeight > p.ProofHorizon {
		expectedRetainAfter = p.CurrentHeight - p.ProofHorizon
	}
	if p.RetainAfterHeight != expectedRetainAfter {
		return errors.New("AVM Store v2 pruning retain height must follow proof horizon")
	}
	if err := validateSortedHashes("AVM Store v2 pruning consumed id", p.PrunableConsumedIDs); err != nil {
		return err
	}
	if err := validateSortedStoreScopes("AVM Store v2 pruning expired nonce scope", p.PrunableExpiredScopes); err != nil {
		return err
	}
	if p.PlanHash == "" {
		return errors.New("AVM Store v2 pruning plan hash is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 pruning plan hash", p.PlanHash); err != nil {
		return err
	}
	if p.PlanHash != ComputeAVMStoreV2TombstonePruningPlanHash(p) {
		return errors.New("AVM Store v2 pruning plan hash mismatch")
	}
	return nil
}

func NewAVMStoreV2LayoutStrategy(layout AVMStoreV2LayoutStrategy) (AVMStoreV2LayoutStrategy, error) {
	layout = canonicalAVMStoreV2LayoutStrategy(layout)
	layout.LayoutRoot = ComputeAVMStoreV2LayoutRoot(layout)
	return layout, layout.Validate()
}

func (l AVMStoreV2LayoutStrategy) Validate() error {
	l = canonicalAVMStoreV2LayoutStrategy(l)
	if len(l.MessageRecords) == 0 {
		return errors.New("AVM Store v2 layout requires compact message records")
	}
	for i, record := range l.MessageRecords {
		if err := record.Validate(); err != nil {
			return err
		}
		if i > 0 && l.MessageRecords[i-1].MessageID >= record.MessageID {
			return errors.New("AVM Store v2 message records must be sorted canonically")
		}
	}
	for i, record := range l.PayloadRecords {
		if err := record.Validate(); err != nil {
			return err
		}
		if i > 0 && l.PayloadRecords[i-1].PayloadHash >= record.PayloadHash {
			return errors.New("AVM Store v2 payload records must be sorted canonically")
		}
	}
	for i, prefix := range l.ActorPrefixes {
		if err := prefix.Validate(); err != nil {
			return err
		}
		if i > 0 && l.ActorPrefixes[i-1].ActorID >= prefix.ActorID {
			return errors.New("AVM Store v2 actor prefixes must be sorted canonically")
		}
	}
	for i, prefix := range l.ContractPrefixes {
		if err := prefix.Validate(); err != nil {
			return err
		}
		if i > 0 && l.ContractPrefixes[i-1].ContractAddress >= prefix.ContractAddress {
			return errors.New("AVM Store v2 contract prefixes must be sorted canonically")
		}
	}
	for i, bucket := range l.DelayedBuckets {
		if err := bucket.Validate(); err != nil {
			return err
		}
		if i > 0 && l.DelayedBuckets[i-1].BucketKey >= bucket.BucketKey {
			return errors.New("AVM Store v2 delayed buckets must be sorted canonically")
		}
	}
	if err := l.PruningPlan.Validate(); err != nil {
		return err
	}
	if l.LayoutRoot == "" {
		return errors.New("AVM Store v2 layout root is required")
	}
	if err := zonestypes.ValidateHash("AVM Store v2 layout root", l.LayoutRoot); err != nil {
		return err
	}
	if l.LayoutRoot != ComputeAVMStoreV2LayoutRoot(l) {
		return errors.New("AVM Store v2 layout root mismatch")
	}
	return nil
}

func AVMStoreV2PayloadKey(payloadHash string) string {
	return AVMStatePrefixAsyncMessages + "/payloads/" + strings.TrimSpace(payloadHash)
}

func AVMStoreV2DelayedQueueBucketKey(zoneID zonestypes.ZoneID, height uint64) string {
	return fmt.Sprintf("%s/%s/delayed/%020d", AVMStatePrefixAsyncQueues, zoneID, height)
}

func AVMStoreV2CompactMessageBytes(msg AVMAsyncMessage) uint32 {
	msg = canonicalAVMAsyncMessage(msg)
	size := len(msg.ID) + len(msg.Source) + len(msg.Destination) + len(msg.PayloadHash) + len(msg.PayloadType)
	size += len(msg.RouteHintOptional) + 8*6 + 1
	if size > AVMStoreV2CompactMessageMaxBytes {
		return AVMStoreV2CompactMessageMaxBytes
	}
	return uint32(size)
}

func ComputeAVMStoreV2MessageRecordHash(record AVMStoreV2MessageRecord) string {
	record = canonicalAVMStoreV2MessageRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-message-record-v1")
	writeEnginePart(h, record.MessageID)
	writeEnginePart(h, record.MessageKey)
	writeEnginePart(h, record.PayloadHash)
	writeEngineBool(h, record.PayloadInline)
	writeEnginePart(h, record.PayloadRefKey)
	writeEngineUint64(h, uint64(record.CompactBytes))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2PayloadRecordHash(record AVMStoreV2PayloadRecord) string {
	record = canonicalAVMStoreV2PayloadRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-payload-record-v1")
	writeEnginePart(h, record.PayloadHash)
	writeEnginePart(h, record.PayloadKey)
	writeEngineUint64(h, uint64(record.PayloadSize))
	writeEngineBool(h, record.Inline)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2ActorStatePrefixHash(prefix AVMStoreV2ActorStatePrefix) string {
	prefix = canonicalAVMStoreV2ActorStatePrefix(prefix)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-actor-prefix-v1")
	writeEnginePart(h, prefix.ActorID)
	writeEnginePart(h, prefix.Prefix)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2ContractStatePrefixHash(prefix AVMStoreV2ContractStatePrefix) string {
	prefix = canonicalAVMStoreV2ContractStatePrefix(prefix)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-contract-prefix-v1")
	writeEnginePart(h, prefix.ContractAddress)
	writeEnginePart(h, prefix.Prefix)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2DelayedQueueBucketHash(bucket AVMStoreV2DelayedQueueBucket) string {
	bucket = canonicalAVMStoreV2DelayedQueueBucket(bucket)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-delayed-bucket-v1")
	writeEnginePart(h, string(bucket.ZoneID))
	writeEngineUint64(h, bucket.Height)
	writeEnginePart(h, bucket.BucketKey)
	writeEngineUint64(h, uint64(len(bucket.MessageIDs)))
	for _, id := range bucket.MessageIDs {
		writeEnginePart(h, id)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2TombstonePruningPlanHash(plan AVMStoreV2TombstonePruningPlan) string {
	plan = canonicalAVMStoreV2TombstonePruningPlan(plan)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-tombstone-pruning-v1")
	writeEngineUint64(h, plan.CurrentHeight)
	writeEngineUint64(h, plan.ProofHorizon)
	writeEngineUint64(h, plan.RetainAfterHeight)
	writeEngineUint64(h, uint64(len(plan.PrunableConsumedIDs)))
	for _, id := range plan.PrunableConsumedIDs {
		writeEnginePart(h, id)
	}
	writeEngineUint64(h, uint64(len(plan.PrunableExpiredScopes)))
	for _, scope := range plan.PrunableExpiredScopes {
		writeEnginePart(h, scope)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMStoreV2LayoutRoot(layout AVMStoreV2LayoutStrategy) string {
	layout = canonicalAVMStoreV2LayoutStrategy(layout)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-store-v2-layout-v1")
	writeEngineUint64(h, uint64(len(layout.MessageRecords)))
	for _, record := range layout.MessageRecords {
		writeEnginePart(h, record.RecordHash)
	}
	writeEngineUint64(h, uint64(len(layout.PayloadRecords)))
	for _, record := range layout.PayloadRecords {
		writeEnginePart(h, record.RecordHash)
	}
	writeEngineUint64(h, uint64(len(layout.ActorPrefixes)))
	for _, prefix := range layout.ActorPrefixes {
		writeEnginePart(h, prefix.PrefixHash)
	}
	writeEngineUint64(h, uint64(len(layout.ContractPrefixes)))
	for _, prefix := range layout.ContractPrefixes {
		writeEnginePart(h, prefix.PrefixHash)
	}
	writeEngineUint64(h, uint64(len(layout.DelayedBuckets)))
	for _, bucket := range layout.DelayedBuckets {
		writeEnginePart(h, bucket.BucketHash)
	}
	writeEnginePart(h, layout.PruningPlan.PlanHash)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMStoreV2MessageRecord(record AVMStoreV2MessageRecord) AVMStoreV2MessageRecord {
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.MessageKey = strings.TrimSpace(record.MessageKey)
	record.PayloadHash = strings.TrimSpace(record.PayloadHash)
	record.PayloadRefKey = strings.TrimSpace(record.PayloadRefKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMStoreV2PayloadRecord(record AVMStoreV2PayloadRecord) AVMStoreV2PayloadRecord {
	record.PayloadHash = strings.TrimSpace(record.PayloadHash)
	record.PayloadKey = strings.TrimSpace(record.PayloadKey)
	record.RecordHash = strings.TrimSpace(record.RecordHash)
	return record
}

func canonicalAVMStoreV2ActorStatePrefix(prefix AVMStoreV2ActorStatePrefix) AVMStoreV2ActorStatePrefix {
	prefix.ActorID = strings.TrimSpace(prefix.ActorID)
	prefix.Prefix = strings.TrimSpace(prefix.Prefix)
	prefix.PrefixHash = strings.TrimSpace(prefix.PrefixHash)
	return prefix
}

func canonicalAVMStoreV2ContractStatePrefix(prefix AVMStoreV2ContractStatePrefix) AVMStoreV2ContractStatePrefix {
	prefix.ContractAddress = strings.TrimSpace(prefix.ContractAddress)
	prefix.Prefix = strings.TrimSpace(prefix.Prefix)
	prefix.PrefixHash = strings.TrimSpace(prefix.PrefixHash)
	return prefix
}

func canonicalAVMStoreV2DelayedQueueBucket(bucket AVMStoreV2DelayedQueueBucket) AVMStoreV2DelayedQueueBucket {
	bucket.BucketKey = strings.TrimSpace(bucket.BucketKey)
	bucket.MessageIDs = trimSortStrings(bucket.MessageIDs)
	bucket.BucketHash = strings.TrimSpace(bucket.BucketHash)
	return bucket
}

func canonicalAVMStoreV2TombstonePruningPlan(plan AVMStoreV2TombstonePruningPlan) AVMStoreV2TombstonePruningPlan {
	plan.PrunableConsumedIDs = trimSortStrings(plan.PrunableConsumedIDs)
	plan.PrunableExpiredScopes = trimSortStrings(plan.PrunableExpiredScopes)
	plan.PlanHash = strings.TrimSpace(plan.PlanHash)
	return plan
}

func canonicalAVMStoreV2LayoutStrategy(layout AVMStoreV2LayoutStrategy) AVMStoreV2LayoutStrategy {
	layout.MessageRecords = append([]AVMStoreV2MessageRecord(nil), layout.MessageRecords...)
	for i := range layout.MessageRecords {
		layout.MessageRecords[i] = canonicalAVMStoreV2MessageRecord(layout.MessageRecords[i])
	}
	sort.Slice(layout.MessageRecords, func(i, j int) bool { return layout.MessageRecords[i].MessageID < layout.MessageRecords[j].MessageID })
	layout.PayloadRecords = append([]AVMStoreV2PayloadRecord(nil), layout.PayloadRecords...)
	for i := range layout.PayloadRecords {
		layout.PayloadRecords[i] = canonicalAVMStoreV2PayloadRecord(layout.PayloadRecords[i])
	}
	sort.Slice(layout.PayloadRecords, func(i, j int) bool {
		return layout.PayloadRecords[i].PayloadHash < layout.PayloadRecords[j].PayloadHash
	})
	layout.ActorPrefixes = append([]AVMStoreV2ActorStatePrefix(nil), layout.ActorPrefixes...)
	for i := range layout.ActorPrefixes {
		layout.ActorPrefixes[i] = canonicalAVMStoreV2ActorStatePrefix(layout.ActorPrefixes[i])
	}
	sort.Slice(layout.ActorPrefixes, func(i, j int) bool { return layout.ActorPrefixes[i].ActorID < layout.ActorPrefixes[j].ActorID })
	layout.ContractPrefixes = append([]AVMStoreV2ContractStatePrefix(nil), layout.ContractPrefixes...)
	for i := range layout.ContractPrefixes {
		layout.ContractPrefixes[i] = canonicalAVMStoreV2ContractStatePrefix(layout.ContractPrefixes[i])
	}
	sort.Slice(layout.ContractPrefixes, func(i, j int) bool {
		return layout.ContractPrefixes[i].ContractAddress < layout.ContractPrefixes[j].ContractAddress
	})
	layout.DelayedBuckets = append([]AVMStoreV2DelayedQueueBucket(nil), layout.DelayedBuckets...)
	for i := range layout.DelayedBuckets {
		layout.DelayedBuckets[i] = canonicalAVMStoreV2DelayedQueueBucket(layout.DelayedBuckets[i])
	}
	sort.Slice(layout.DelayedBuckets, func(i, j int) bool { return layout.DelayedBuckets[i].BucketKey < layout.DelayedBuckets[j].BucketKey })
	layout.PruningPlan = canonicalAVMStoreV2TombstonePruningPlan(layout.PruningPlan)
	layout.LayoutRoot = strings.TrimSpace(layout.LayoutRoot)
	return layout
}

func validateSortedStoreScopes(fieldName string, values []string) error {
	for i, value := range values {
		if err := validateRouterOptionalToken(fieldName, value, MaxAVMStateKeySegmentLength*2); err != nil {
			return err
		}
		if value == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
		if i > 0 && values[i-1] >= value {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
	}
	return nil
}
