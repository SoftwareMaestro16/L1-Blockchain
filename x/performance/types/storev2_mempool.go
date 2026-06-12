package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	MaxStoreV2RangeLimit	= uint32(1_024)
	MaxStoreV2ValueBytes	= uint32(1 << 20)
	MaxMempoolTxBytes	= uint32(1 << 20)
	SystemMempoolZoneID	= "system"
	SystemMempoolShardID	= "system-shard"
	DefaultMempoolMaxGas	= uint64(10_000_000)
	DefaultMempoolLaneCap	= uint32(50_000)
)

type StoreV2RecordKind string

const (
	StoreV2RecordDomain	StoreV2RecordKind	= "domain"
	StoreV2RecordContract	StoreV2RecordKind	= "contract"
	StoreV2RecordChannel	StoreV2RecordKind	= "channel"
	StoreV2RecordPool	StoreV2RecordKind	= "pool"
	StoreV2RecordAccount	StoreV2RecordKind	= "account"
)

type StoreV2BenchmarkOperation string

const (
	StoreV2BenchDirectBalanceRead		StoreV2BenchmarkOperation	= "direct_balance_read"
	StoreV2BenchDirectIdentityResolution	StoreV2BenchmarkOperation	= "direct_identity_resolution"
	StoreV2BenchRecursiveIdentityResolve	StoreV2BenchmarkOperation	= "recursive_identity_resolution"
	StoreV2BenchContractStorageReadWrite	StoreV2BenchmarkOperation	= "contract_storage_read_write"
	StoreV2BenchMessageEnqueueDequeue	StoreV2BenchmarkOperation	= "message_enqueue_dequeue"
	StoreV2BenchPaymentChannelSettle	StoreV2BenchmarkOperation	= "payment_channel_settle"
	StoreV2BenchProofGeneration		StoreV2BenchmarkOperation	= "proof_generation"
)

type StoreV2ObjectRecord struct {
	ZoneID		string
	ShardID		string
	Kind		StoreV2RecordKind
	ObjectKey	string
	ValueHash	string
	Version		uint64
	UpdatedHeight	uint64
	SizeBytes	uint32
	RecordHash	string
}

type StoreV2KVField struct {
	ZoneID		string
	ShardID		string
	ObjectKey	string
	FieldPath	string
	ValueHash	string
	Version		uint64
	UpdatedHeight	uint64
	FieldHash	string
}

type StoreV2ShardState struct {
	ZoneID		string
	ShardID		string
	Records		[]StoreV2ObjectRecord
	KVFields	[]StoreV2KVField
	RootHash	string
}

type StoreV2ProofEntry struct {
	Key		string
	Hash		string
	Version		uint64
	Kind		string
	LeafHash	string
}

type StoreV2PrefixProof struct {
	ZoneID		string
	ShardID		string
	Prefix		string
	StartAfter	string
	Limit		uint32
	Entries		[]StoreV2ProofEntry
	ProofRoot	string
}

type StoreV2ShardRoot struct {
	ZoneID		string
	ShardID		string
	RootHash	string
}

type StoreV2ZoneRoot struct {
	ZoneID		string
	ShardRoots	[]StoreV2ShardRoot
	ZoneRootHash	string
}

type StoreV2BenchmarkResult struct {
	Operation		StoreV2BenchmarkOperation
	Samples			uint64
	Operations		uint64
	MaxRangeLimit		uint32
	ObservedRootHash	string
}

type MempoolMessageClass string

const (
	MempoolClassAccount	MempoolMessageClass	= "account"
	MempoolClassIdentity	MempoolMessageClass	= "identity"
	MempoolClassContract	MempoolMessageClass	= "contract"
	MempoolClassPayment	MempoolMessageClass	= "payment"
	MempoolClassDEX		MempoolMessageClass	= "dex"
	MempoolClassMessage	MempoolMessageClass	= "message"
	MempoolClassSystem	MempoolMessageClass	= "system"
)

type SeparatedMempoolTx struct {
	TxID		string
	Sender		string
	TargetZoneID	string
	TargetShardID	string
	TargetObject	string
	RouteKey	string
	MessageClass	MempoolMessageClass
	FeeAmount	string
	GasWanted	uint64
	SizeBytes	uint32
	CreatedHeight	uint64
	ExpiryHeight	uint64
	PreResolved	bool
	TxHash		string
}

type MempoolSeparationLimits struct {
	MaxPerSender		uint32
	MaxPerTargetObject	uint32
	MaxLaneSize		uint32
	MaxTxBytes		uint32
	MaxGasWanted		uint64
}

type MempoolLane struct {
	ZoneID		string
	ShardID		string
	MessageClass	MempoolMessageClass
	Transactions	[]SeparatedMempoolTx
	LaneHash	string
}

type SeparatedMempoolSnapshot struct {
	Height		uint64
	Lanes		[]MempoolLane
	RootHash	string
}

func BuildStoreV2ShardState(zoneID, shardID string, records []StoreV2ObjectRecord, fields []StoreV2KVField) (StoreV2ShardState, error) {
	state := StoreV2ShardState{
		ZoneID:		strings.TrimSpace(zoneID),
		ShardID:	strings.TrimSpace(shardID),
		Records:	normalizeStoreV2Records(records),
		KVFields:	normalizeStoreV2Fields(fields),
	}
	if err := state.ValidateWithoutRoot(); err != nil {
		return StoreV2ShardState{}, err
	}
	state.RootHash = ComputeStoreV2ShardRoot(state)
	return state, state.Validate()
}

func (s StoreV2ShardState) ValidateWithoutRoot() error {
	if err := validateExecutionToken("Store v2 shard zone id", s.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("Store v2 shard id", s.ShardID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(s.Records)+len(s.KVFields))
	for _, record := range s.Records {
		if err := record.Validate(); err != nil {
			return err
		}
		if record.ZoneID != s.ZoneID || record.ShardID != s.ShardID {
			return errors.New("Store v2 record zone or shard mismatch")
		}
		key := StoreV2ObjectKey(record)
		if _, found := seen[key]; found {
			return errors.New("Store v2 duplicate object key")
		}
		seen[key] = struct{}{}
	}
	for _, field := range s.KVFields {
		if err := field.Validate(); err != nil {
			return err
		}
		if field.ZoneID != s.ZoneID || field.ShardID != s.ShardID {
			return errors.New("Store v2 field zone or shard mismatch")
		}
		key := StoreV2FieldKey(field)
		if _, found := seen[key]; found {
			return errors.New("Store v2 duplicate field key")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func (s StoreV2ShardState) Validate() error {
	if err := s.ValidateWithoutRoot(); err != nil {
		return err
	}
	if err := validateHexHash("Store v2 shard root", s.RootHash); err != nil {
		return err
	}
	if s.RootHash != ComputeStoreV2ShardRoot(s) {
		return errors.New("Store v2 shard root mismatch")
	}
	return nil
}

func (r StoreV2ObjectRecord) Normalize() StoreV2ObjectRecord {
	r.ZoneID = strings.TrimSpace(r.ZoneID)
	r.ShardID = strings.TrimSpace(r.ShardID)
	r.ObjectKey = strings.TrimSpace(r.ObjectKey)
	r.ValueHash = normalizeLowerHex(r.ValueHash)
	r.RecordHash = normalizeLowerHex(r.RecordHash)
	return r
}

func (r StoreV2ObjectRecord) Validate() error {
	record := r.Normalize()
	if err := validateExecutionToken("Store v2 record zone id", record.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("Store v2 record shard id", record.ShardID); err != nil {
		return err
	}
	if !IsStoreV2RecordKind(record.Kind) {
		return errors.New("Store v2 record kind is unsupported")
	}
	if err := validateStoreV2Path("Store v2 object key", record.ObjectKey); err != nil {
		return err
	}
	if err := validateHexHash("Store v2 record value hash", record.ValueHash); err != nil {
		return err
	}
	if record.Version == 0 || record.UpdatedHeight == 0 {
		return errors.New("Store v2 record version and height must be positive")
	}
	if record.SizeBytes > MaxStoreV2ValueBytes {
		return errors.New("Store v2 record value exceeds max size")
	}
	if record.RecordHash != ComputeStoreV2RecordHash(record) {
		return errors.New("Store v2 record hash mismatch")
	}
	return nil
}

func (f StoreV2KVField) Normalize() StoreV2KVField {
	f.ZoneID = strings.TrimSpace(f.ZoneID)
	f.ShardID = strings.TrimSpace(f.ShardID)
	f.ObjectKey = strings.TrimSpace(f.ObjectKey)
	f.FieldPath = strings.TrimSpace(f.FieldPath)
	f.ValueHash = normalizeLowerHex(f.ValueHash)
	f.FieldHash = normalizeLowerHex(f.FieldHash)
	return f
}

func (f StoreV2KVField) Validate() error {
	field := f.Normalize()
	if err := validateExecutionToken("Store v2 field zone id", field.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("Store v2 field shard id", field.ShardID); err != nil {
		return err
	}
	if err := validateStoreV2Path("Store v2 field object key", field.ObjectKey); err != nil {
		return err
	}
	if err := validateStoreV2Path("Store v2 field path", field.FieldPath); err != nil {
		return err
	}
	if err := validateHexHash("Store v2 field value hash", field.ValueHash); err != nil {
		return err
	}
	if field.Version == 0 || field.UpdatedHeight == 0 {
		return errors.New("Store v2 field version and height must be positive")
	}
	if field.FieldHash != ComputeStoreV2FieldHash(field) {
		return errors.New("Store v2 field hash mismatch")
	}
	return nil
}

func StoreV2BoundedRangeScan(state StoreV2ShardState, prefix, startAfter string, limit uint32) ([]StoreV2ProofEntry, error) {
	if err := state.Validate(); err != nil {
		return nil, err
	}
	prefix = strings.TrimSpace(prefix)
	startAfter = strings.TrimSpace(startAfter)
	if prefix == "" {
		return nil, errors.New("Store v2 range prefix is required")
	}
	if limit == 0 || limit > MaxStoreV2RangeLimit {
		return nil, errors.New("Store v2 range limit is out of bounds")
	}
	entries := storeV2ProofEntries(state)
	out := make([]StoreV2ProofEntry, 0, limit)
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Key, prefix) {
			continue
		}
		if startAfter != "" && entry.Key <= startAfter {
			continue
		}
		out = append(out, entry)
		if uint32(len(out)) == limit {
			break
		}
	}
	return out, nil
}

func GenerateStoreV2PrefixProof(state StoreV2ShardState, prefix, startAfter string, limit uint32) (StoreV2PrefixProof, error) {
	entries, err := StoreV2BoundedRangeScan(state, prefix, startAfter, limit)
	if err != nil {
		return StoreV2PrefixProof{}, err
	}
	proof := StoreV2PrefixProof{
		ZoneID:		state.ZoneID,
		ShardID:	state.ShardID,
		Prefix:		strings.TrimSpace(prefix),
		StartAfter:	strings.TrimSpace(startAfter),
		Limit:		limit,
		Entries:	entries,
	}
	proof.ProofRoot = ComputeStoreV2PrefixProofRoot(proof)
	return proof, proof.Validate(state.RootHash)
}

func (p StoreV2PrefixProof) Validate(shardRoot string) error {
	if err := validateExecutionToken("Store v2 proof zone id", p.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("Store v2 proof shard id", p.ShardID); err != nil {
		return err
	}
	if p.Prefix == "" {
		return errors.New("Store v2 proof prefix is required")
	}
	if p.Limit == 0 || p.Limit > MaxStoreV2RangeLimit {
		return errors.New("Store v2 proof limit is out of bounds")
	}
	if err := validateHexHash("Store v2 proof shard root", normalizeLowerHex(shardRoot)); err != nil {
		return err
	}
	if err := validateHexHash("Store v2 proof root", p.ProofRoot); err != nil {
		return err
	}
	previous := ""
	for _, entry := range p.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if !strings.HasPrefix(entry.Key, p.Prefix) {
			return errors.New("Store v2 proof entry prefix mismatch")
		}
		if p.StartAfter != "" && entry.Key <= p.StartAfter {
			return errors.New("Store v2 proof entry violates cursor")
		}
		if previous != "" && previous >= entry.Key {
			return errors.New("Store v2 proof entries must be sorted canonically")
		}
		previous = entry.Key
	}
	if p.ProofRoot != ComputeStoreV2PrefixProofRoot(p) {
		return errors.New("Store v2 proof root mismatch")
	}
	return nil
}

func (e StoreV2ProofEntry) Validate() error {
	if e.Key == "" || e.Kind == "" {
		return errors.New("Store v2 proof entry key and kind are required")
	}
	if err := validateHexHash("Store v2 proof entry hash", e.Hash); err != nil {
		return err
	}
	if e.Version == 0 {
		return errors.New("Store v2 proof entry version must be positive")
	}
	if e.LeafHash != ComputeStoreV2ProofEntryHash(e) {
		return errors.New("Store v2 proof entry leaf hash mismatch")
	}
	return nil
}

func BuildStoreV2ZoneRoot(zoneID string, shards []StoreV2ShardState) (StoreV2ZoneRoot, error) {
	zoneID = strings.TrimSpace(zoneID)
	if err := validateExecutionToken("Store v2 zone root zone id", zoneID); err != nil {
		return StoreV2ZoneRoot{}, err
	}
	roots := make([]StoreV2ShardRoot, 0, len(shards))
	for _, shard := range shards {
		if err := shard.Validate(); err != nil {
			return StoreV2ZoneRoot{}, err
		}
		if shard.ZoneID != zoneID {
			return StoreV2ZoneRoot{}, errors.New("Store v2 zone root shard zone mismatch")
		}
		roots = append(roots, StoreV2ShardRoot{ZoneID: shard.ZoneID, ShardID: shard.ShardID, RootHash: shard.RootHash})
	}
	sort.SliceStable(roots, func(i, j int) bool {
		return roots[i].ShardID < roots[j].ShardID
	})
	root := StoreV2ZoneRoot{ZoneID: zoneID, ShardRoots: roots}
	root.ZoneRootHash = ComputeStoreV2ZoneRootHash(root)
	return root, root.Validate()
}

func (z StoreV2ZoneRoot) Validate() error {
	if err := validateExecutionToken("Store v2 zone root zone id", z.ZoneID); err != nil {
		return err
	}
	previous := ""
	for _, shard := range z.ShardRoots {
		if shard.ZoneID != z.ZoneID {
			return errors.New("Store v2 zone root shard zone mismatch")
		}
		if err := validateExecutionToken("Store v2 zone root shard id", shard.ShardID); err != nil {
			return err
		}
		if err := validateHexHash("Store v2 zone shard root", shard.RootHash); err != nil {
			return err
		}
		if previous != "" && previous >= shard.ShardID {
			return errors.New("Store v2 zone shard roots must be sorted canonically")
		}
		previous = shard.ShardID
	}
	if z.ZoneRootHash != ComputeStoreV2ZoneRootHash(z) {
		return errors.New("Store v2 zone root hash mismatch")
	}
	return nil
}

func RequiredStoreV2Benchmarks() []StoreV2BenchmarkOperation {
	return []StoreV2BenchmarkOperation{
		StoreV2BenchDirectBalanceRead,
		StoreV2BenchDirectIdentityResolution,
		StoreV2BenchRecursiveIdentityResolve,
		StoreV2BenchContractStorageReadWrite,
		StoreV2BenchMessageEnqueueDequeue,
		StoreV2BenchPaymentChannelSettle,
		StoreV2BenchProofGeneration,
	}
}

func ValidateStoreV2BenchmarkCoverage(results []StoreV2BenchmarkResult) error {
	seen := make(map[StoreV2BenchmarkOperation]struct{}, len(results))
	for _, result := range results {
		if result.Samples == 0 || result.Operations == 0 {
			return errors.New("Store v2 benchmark result requires samples and operations")
		}
		if result.MaxRangeLimit > MaxStoreV2RangeLimit {
			return errors.New("Store v2 benchmark range limit exceeds bound")
		}
		if result.ObservedRootHash != "" {
			if err := validateHexHash("Store v2 benchmark observed root", result.ObservedRootHash); err != nil {
				return err
			}
		}
		seen[result.Operation] = struct{}{}
	}
	for _, required := range RequiredStoreV2Benchmarks() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("Store v2 benchmark %s is required", required)
		}
	}
	return nil
}

func BuildSeparatedMempoolSnapshot(height uint64, txs []SeparatedMempoolTx, limits MempoolSeparationLimits) (SeparatedMempoolSnapshot, error) {
	if height == 0 {
		return SeparatedMempoolSnapshot{}, errors.New("mempool separation height must be positive")
	}
	limits = limits.Normalize()
	if err := limits.Validate(); err != nil {
		return SeparatedMempoolSnapshot{}, err
	}
	normalized := make([]SeparatedMempoolTx, 0, len(txs))
	perSender := make(map[string]uint32)
	perTarget := make(map[string]uint32)
	for _, tx := range txs {
		tx = tx.NormalizeTarget()
		if err := tx.Validate(height, limits); err != nil {
			return SeparatedMempoolSnapshot{}, err
		}
		perSender[tx.Sender]++
		if perSender[tx.Sender] > limits.MaxPerSender {
			return SeparatedMempoolSnapshot{}, errors.New("mempool separation per-sender limit exceeded")
		}
		target := tx.TargetZoneID + "/" + tx.TargetShardID + "/" + tx.TargetObject
		perTarget[target]++
		if perTarget[target] > limits.MaxPerTargetObject {
			return SeparatedMempoolSnapshot{}, errors.New("mempool separation per-target object limit exceeded")
		}
		normalized = append(normalized, tx)
	}
	lanes := buildMempoolLanes(normalized, limits)
	snapshot := SeparatedMempoolSnapshot{Height: height, Lanes: lanes}
	snapshot.RootHash = ComputeSeparatedMempoolRoot(snapshot)
	return snapshot, snapshot.Validate(limits)
}

func (l MempoolSeparationLimits) Normalize() MempoolSeparationLimits {
	if l.MaxPerSender == 0 {
		l.MaxPerSender = 512
	}
	if l.MaxPerTargetObject == 0 {
		l.MaxPerTargetObject = 256
	}
	if l.MaxLaneSize == 0 {
		l.MaxLaneSize = DefaultMempoolLaneCap
	}
	if l.MaxTxBytes == 0 {
		l.MaxTxBytes = MaxMempoolTxBytes
	}
	if l.MaxGasWanted == 0 {
		l.MaxGasWanted = DefaultMempoolMaxGas
	}
	return l
}

func (l MempoolSeparationLimits) Validate() error {
	if l.MaxPerSender == 0 || l.MaxPerTargetObject == 0 || l.MaxLaneSize == 0 {
		return errors.New("mempool separation limits must be positive")
	}
	if l.MaxTxBytes == 0 || l.MaxTxBytes > MaxMempoolTxBytes {
		return errors.New("mempool separation tx size limit is out of bounds")
	}
	if l.MaxGasWanted == 0 {
		return errors.New("mempool separation gas limit must be positive")
	}
	return nil
}

func (tx SeparatedMempoolTx) NormalizeTarget() SeparatedMempoolTx {
	tx.TxID = strings.TrimSpace(tx.TxID)
	tx.Sender = strings.TrimSpace(tx.Sender)
	tx.TargetZoneID = strings.TrimSpace(tx.TargetZoneID)
	tx.TargetShardID = strings.TrimSpace(tx.TargetShardID)
	tx.TargetObject = strings.TrimSpace(tx.TargetObject)
	tx.RouteKey = strings.TrimSpace(tx.RouteKey)
	tx.FeeAmount = strings.TrimSpace(tx.FeeAmount)
	tx.TxHash = normalizeLowerHex(tx.TxHash)
	if tx.TargetZoneID == "" {
		tx.TargetZoneID = SystemMempoolZoneID
	}
	if tx.TargetShardID == "" || (tx.TargetObject == "" && !tx.PreResolved) {
		tx.TargetShardID = SystemMempoolShardID
		tx.MessageClass = MempoolClassSystem
		if tx.TargetObject == "" {
			tx.TargetObject = "unknown-target"
		}
	}
	if tx.MessageClass == "" {
		tx.MessageClass = MempoolClassAccount
	}
	if tx.RouteKey == "" {
		tx.RouteKey = tx.TargetObject
	}
	return tx
}

func (tx SeparatedMempoolTx) Validate(currentHeight uint64, limits MempoolSeparationLimits) error {
	item := tx.NormalizeTarget()
	if item.TxID == "" || item.Sender == "" {
		return errors.New("mempool separation tx id and sender are required")
	}
	if err := validateExecutionToken("mempool separation target zone id", item.TargetZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("mempool separation target shard id", item.TargetShardID); err != nil {
		return err
	}
	if err := validateStoreV2Path("mempool separation target object", item.TargetObject); err != nil {
		return err
	}
	if err := validateStoreV2Path("mempool separation route key", item.RouteKey); err != nil {
		return err
	}
	if !IsMempoolMessageClass(item.MessageClass) {
		return errors.New("mempool separation message class is unsupported")
	}
	if _, err := parsePerformanceNonNegativeInt("mempool separation fee", item.FeeAmount); err != nil {
		return err
	}
	if item.GasWanted == 0 || item.GasWanted > limits.MaxGasWanted {
		return errors.New("mempool separation gas wanted is out of bounds")
	}
	if item.SizeBytes == 0 || item.SizeBytes > limits.MaxTxBytes {
		return errors.New("mempool separation tx size is out of bounds")
	}
	if item.ExpiryHeight == 0 || item.ExpiryHeight < currentHeight {
		return errors.New("mempool separation tx is expired")
	}
	if item.TargetShardID != SystemMempoolShardID && !item.PreResolved && item.RouteKey == "" {
		return errors.New("mempool separation unknown target must be pre-resolved or routed to system shard")
	}
	expected := ComputeSeparatedMempoolTxHash(item)
	if item.TxHash != "" && item.TxHash != expected {
		return errors.New("mempool separation tx hash mismatch")
	}
	return nil
}

func (s SeparatedMempoolSnapshot) Validate(limits MempoolSeparationLimits) error {
	if s.Height == 0 {
		return errors.New("mempool separation snapshot height must be positive")
	}
	previous := ""
	for _, lane := range s.Lanes {
		if err := lane.Validate(s.Height, limits); err != nil {
			return err
		}
		key := lane.ZoneID + "/" + lane.ShardID + "/" + string(lane.MessageClass)
		if previous != "" && previous >= key {
			return errors.New("mempool separation lanes must be sorted canonically")
		}
		previous = key
	}
	if s.RootHash != ComputeSeparatedMempoolRoot(s) {
		return errors.New("mempool separation root hash mismatch")
	}
	return nil
}

func (l MempoolLane) Validate(height uint64, limits MempoolSeparationLimits) error {
	if err := validateExecutionToken("mempool lane zone id", l.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("mempool lane shard id", l.ShardID); err != nil {
		return err
	}
	if !IsMempoolMessageClass(l.MessageClass) {
		return errors.New("mempool lane message class is unsupported")
	}
	if uint32(len(l.Transactions)) > limits.MaxLaneSize {
		return errors.New("mempool lane exceeds max size")
	}
	for i, tx := range l.Transactions {
		if tx.TargetZoneID != l.ZoneID || tx.TargetShardID != l.ShardID || tx.MessageClass != l.MessageClass {
			return errors.New("mempool lane transaction target mismatch")
		}
		if err := tx.Validate(height, limits); err != nil {
			return err
		}
		if i > 0 && compareMempoolTx(l.Transactions[i-1], tx) > 0 {
			return errors.New("mempool lane transactions must be sorted canonically")
		}
	}
	if l.LaneHash != ComputeMempoolLaneHash(l) {
		return errors.New("mempool lane hash mismatch")
	}
	return nil
}

func IsStoreV2RecordKind(kind StoreV2RecordKind) bool {
	switch kind {
	case StoreV2RecordDomain, StoreV2RecordContract, StoreV2RecordChannel, StoreV2RecordPool, StoreV2RecordAccount:
		return true
	default:
		return false
	}
}

func IsMempoolMessageClass(class MempoolMessageClass) bool {
	switch class {
	case MempoolClassAccount, MempoolClassIdentity, MempoolClassContract, MempoolClassPayment, MempoolClassDEX, MempoolClassMessage, MempoolClassSystem:
		return true
	default:
		return false
	}
}

func StoreV2ObjectKey(record StoreV2ObjectRecord) string {
	record = record.Normalize()
	return fmt.Sprintf("storev2/object/%s/%s/%s/%s", record.ZoneID, record.ShardID, record.Kind, record.ObjectKey)
}

func StoreV2FieldKey(field StoreV2KVField) string {
	field = field.Normalize()
	return fmt.Sprintf("storev2/kv/%s/%s/%s/%s", field.ZoneID, field.ShardID, field.ObjectKey, field.FieldPath)
}

func ComputeStoreV2RecordHash(record StoreV2ObjectRecord) string {
	record = record.Normalize()
	return hashStrings("storev2-record", StoreV2ObjectKey(record), record.ValueHash, fmt.Sprintf("%020d", record.Version), fmt.Sprintf("%020d", record.UpdatedHeight), fmt.Sprintf("%020d", record.SizeBytes))
}

func ComputeStoreV2FieldHash(field StoreV2KVField) string {
	field = field.Normalize()
	return hashStrings("storev2-field", StoreV2FieldKey(field), field.ValueHash, fmt.Sprintf("%020d", field.Version), fmt.Sprintf("%020d", field.UpdatedHeight))
}

func ComputeStoreV2ShardRoot(state StoreV2ShardState) string {
	entries := storeV2ProofEntries(state)
	parts := []string{"storev2-shard-root", state.ZoneID, state.ShardID}
	for _, entry := range entries {
		parts = append(parts, entry.LeafHash)
	}
	return hashStrings(parts...)
}

func ComputeStoreV2ProofEntryHash(entry StoreV2ProofEntry) string {
	return hashStrings("storev2-proof-entry", entry.Key, entry.Hash, fmt.Sprintf("%020d", entry.Version), entry.Kind)
}

func ComputeStoreV2PrefixProofRoot(proof StoreV2PrefixProof) string {
	parts := []string{"storev2-prefix-proof", proof.ZoneID, proof.ShardID, proof.Prefix, proof.StartAfter, fmt.Sprintf("%020d", proof.Limit)}
	for _, entry := range proof.Entries {
		parts = append(parts, entry.LeafHash)
	}
	return hashStrings(parts...)
}

func ComputeStoreV2ZoneRootHash(root StoreV2ZoneRoot) string {
	parts := []string{"storev2-zone-root", root.ZoneID}
	for _, shard := range root.ShardRoots {
		parts = append(parts, shard.ShardID, shard.RootHash)
	}
	return hashStrings(parts...)
}

func ComputeSeparatedMempoolTxHash(tx SeparatedMempoolTx) string {
	tx = tx.NormalizeTarget()
	return hashStrings("separated-mempool-tx", tx.TxID, tx.Sender, tx.TargetZoneID, tx.TargetShardID, tx.TargetObject, tx.RouteKey, string(tx.MessageClass), tx.FeeAmount, fmt.Sprintf("%020d", tx.GasWanted), fmt.Sprintf("%020d", tx.SizeBytes), fmt.Sprintf("%020d", tx.CreatedHeight), fmt.Sprintf("%020d", tx.ExpiryHeight))
}

func ComputeMempoolLaneHash(lane MempoolLane) string {
	parts := []string{"separated-mempool-lane", lane.ZoneID, lane.ShardID, string(lane.MessageClass)}
	for _, tx := range lane.Transactions {
		parts = append(parts, tx.TxHash)
	}
	return hashStrings(parts...)
}

func ComputeSeparatedMempoolRoot(snapshot SeparatedMempoolSnapshot) string {
	parts := []string{"separated-mempool-root", fmt.Sprintf("%020d", snapshot.Height)}
	for _, lane := range snapshot.Lanes {
		parts = append(parts, lane.LaneHash)
	}
	return hashStrings(parts...)
}

func normalizeStoreV2Records(records []StoreV2ObjectRecord) []StoreV2ObjectRecord {
	out := make([]StoreV2ObjectRecord, len(records))
	for i, record := range records {
		out[i] = record.Normalize()
		if out[i].RecordHash == "" {
			out[i].RecordHash = ComputeStoreV2RecordHash(out[i])
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return StoreV2ObjectKey(out[i]) < StoreV2ObjectKey(out[j])
	})
	return out
}

func normalizeStoreV2Fields(fields []StoreV2KVField) []StoreV2KVField {
	out := make([]StoreV2KVField, len(fields))
	for i, field := range fields {
		out[i] = field.Normalize()
		if out[i].FieldHash == "" {
			out[i].FieldHash = ComputeStoreV2FieldHash(out[i])
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return StoreV2FieldKey(out[i]) < StoreV2FieldKey(out[j])
	})
	return out
}

func storeV2ProofEntries(state StoreV2ShardState) []StoreV2ProofEntry {
	entries := make([]StoreV2ProofEntry, 0, len(state.Records)+len(state.KVFields))
	for _, record := range normalizeStoreV2Records(state.Records) {
		entry := StoreV2ProofEntry{Key: StoreV2ObjectKey(record), Hash: record.RecordHash, Version: record.Version, Kind: string(record.Kind)}
		entry.LeafHash = ComputeStoreV2ProofEntryHash(entry)
		entries = append(entries, entry)
	}
	for _, field := range normalizeStoreV2Fields(state.KVFields) {
		entry := StoreV2ProofEntry{Key: StoreV2FieldKey(field), Hash: field.FieldHash, Version: field.Version, Kind: "kv"}
		entry.LeafHash = ComputeStoreV2ProofEntryHash(entry)
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})
	return entries
}

func buildMempoolLanes(txs []SeparatedMempoolTx, limits MempoolSeparationLimits) []MempoolLane {
	byLane := make(map[string][]SeparatedMempoolTx)
	keys := make([]string, 0)
	for _, tx := range txs {
		tx = tx.NormalizeTarget()
		tx.TxHash = ComputeSeparatedMempoolTxHash(tx)
		key := tx.TargetZoneID + "/" + tx.TargetShardID + "/" + string(tx.MessageClass)
		if _, found := byLane[key]; !found {
			keys = append(keys, key)
		}
		byLane[key] = append(byLane[key], tx)
	}
	sort.Strings(keys)
	lanes := make([]MempoolLane, 0, len(keys))
	for _, key := range keys {
		transactions := byLane[key]
		sort.SliceStable(transactions, func(i, j int) bool {
			return compareMempoolTx(transactions[i], transactions[j]) < 0
		})
		if uint32(len(transactions)) > limits.MaxLaneSize {
			transactions = transactions[:limits.MaxLaneSize]
		}
		parts := strings.Split(key, "/")
		lane := MempoolLane{ZoneID: parts[0], ShardID: parts[1], MessageClass: MempoolMessageClass(parts[2]), Transactions: transactions}
		lane.LaneHash = ComputeMempoolLaneHash(lane)
		lanes = append(lanes, lane)
	}
	return lanes
}

func compareMempoolTx(left, right SeparatedMempoolTx) int {
	left = left.NormalizeTarget()
	right = right.NormalizeTarget()
	if left.MessageClass == MempoolClassMessage || left.MessageClass == MempoolClassSystem {
		if left.ExpiryHeight != right.ExpiryHeight {
			return compareUint64(left.ExpiryHeight, right.ExpiryHeight)
		}
	}
	leftFee, _ := parsePerformanceNonNegativeInt("left fee", left.FeeAmount)
	rightFee, _ := parsePerformanceNonNegativeInt("right fee", right.FeeAmount)
	if !leftFee.Equal(rightFee) {
		if leftFee.GT(rightFee) {
			return -1
		}
		return 1
	}
	if left.ExpiryHeight != right.ExpiryHeight {
		return compareUint64(left.ExpiryHeight, right.ExpiryHeight)
	}
	if left.CreatedHeight != right.CreatedHeight {
		return compareUint64(left.CreatedHeight, right.CreatedHeight)
	}
	if left.TxID < right.TxID {
		return -1
	}
	if left.TxID > right.TxID {
		return 1
	}
	return 0
}

func validateStoreV2Path(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > 160 {
		return fmt.Errorf("%s is too long", field)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains unsupported character", field)
	}
	return nil
}

func addStoreV2Amount(valueHash string, delta uint64) string {
	current := sdkmath.NewInt(int64(delta))
	return hashStrings(valueHash, current.String())
}
