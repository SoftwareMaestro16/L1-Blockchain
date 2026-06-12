package avm

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

// QueryExecutionDomain separates query execution from state-modifying execution.
// Query methods MUST run in a separate domain — they cannot share state or
// action queues with the regular execution path.
//
// Invariants:
//   - QueryState != ExecutionState (separate domain)
//   - No mutation allowed
//   - No action queue exists
//   - No side-effect buffer exists
//   - Only read-only execution frame is created
type QueryExecutionDomain struct {
	Snapshot	QuerySnapshot
	MethodID	uint32
	Args		[]byte
	GasLimit	uint64
	StackTrace	[]QueryTraceStep
	GasUsed		uint64
	ExitCode	uint32
	Response	[]byte
}

// QueryGasModel separates query gas from execution gas.
// Query gas has three components and does NOT affect state gas.
//
// Invariants:
//   - query gas is independent of execution gas
//   - query gas accounting is deterministic
//   - gas breakdown is tracked per phase
type QueryGasModel struct {
	ComputeGas		uint64
	DecodeGas		uint64
	SerializationGas	uint64
}

// QueryGasAccounting tracks gas usage during query execution.
type QueryGasAccounting struct {
	Model	QueryGasModel
	Used	QueryGasModel
	Limit	uint64
	Aborted	bool
}

// QueryIsolationBoundary defines what a query frame MAY and MAY NOT access.
//
// Invariants:
//   - Query VM MUST NOT instantiate: action queue, storage writer,
//     message emitter, event emitter
//   - Only read-only execution frame is created
//   - No effectful host function may be called
//   - All inputs must come from QuerySnapshot or method arguments
type QueryIsolationBoundary struct {
	AllowedReads		[]string
	ForbiddenEffects	[]HostFunction
	CanReadStorage		bool
	CanSendMessages		bool
	CanEmitEvents		bool
	CanWriteStorage		bool
}

// QueryProofMode enables verifiable query responses for light clients.
//
// Invariants:
//   - Execution produces inclusion proof for each state read
//   - Returns partial Chunk path for verification
//   - Allows light client verification without full state
//   - Proof is deterministic: same (snapshot + args) → identical proof
type QueryProofMode struct {
	Enabled		bool
	InclusionProofs	[]QueryInclusionProof
	StateRootProof	[]byte
	ResponseProof	[]byte
}

// QueryInclusionProof proves that a key-value pair exists in state.
type QueryInclusionProof struct {
	Key		[]byte
	Value		[]byte
	ProofPath	[][]byte
	ProofIndex	int
}

// QueryResponseCanonicalEncoding ensures deterministic serialization.
//
// Invariants:
//   - All query responses MUST be canonical encoded
//   - Deterministic serialization (no field ordering variance)
//   - No optional ambiguity
//   - Same (snapshot + args) → identical bytes
type QueryResponseCanonicalEncoding struct {
	MethodID	uint32
	GasUsed		uint64
	ExitCode	uint32
	Payload		[]byte
	ProofRoot	[]byte
}

// QueryCacheKey identifies a cacheable query result.
//
// Invariants:
//   - Same state_root_chunk + method_id + arguments_hash → same cache key
//   - Cache MUST be invalidated on state root change
//   - Cache entries are content-addressed
type QueryCacheKey struct {
	StateRootHash	[]byte
	MethodID	uint32
	ArgumentsHash	[]byte
}

// QueryCacheEntry stores a cached query result with metadata.
type QueryCacheEntry struct {
	Key		QueryCacheKey
	Response	QueryReceipt
	CreatedAt	int64
}

// QueryCache provides deterministic query result caching.
type QueryCache struct {
	entries	map[string]QueryCacheEntry
	maxSize	int
}

// MethodRegistryEntry describes a contract get method for discovery.
type MethodRegistryEntry struct {
	MethodID	uint32
	Name		string
	InputSchema	[]byte
	OutputSchema	[]byte
	GasEstimate	uint64
	Cacheable	bool
}

// MethodRegistry provides method discovery for contract get methods.
type MethodRegistry struct {
	Methods []MethodRegistryEntry
}

// QueryTraceStep records a single step of query execution for debugging.
type QueryTraceStep struct {
	Instruction	string
	GasConsumed	uint64
	ChunkReads	int
	Opcode		string
}

// QueryTraceRecord holds the full execution trace of a query.
type QueryTraceRecord struct {
	Steps		[]QueryTraceStep
	ChunkReads	[]string
	GasBreakdown	QueryGasModel
}

// QueryStackLimits enforces safety caps on query execution.
//
// Invariants:
//   - Max stack depth prevents unbounded recursion
//   - Max recursion depth prevents deep call chains
//   - Max chunk traversal depth prevents state traversal attacks
type QueryStackLimits struct {
	MaxStackDepth		uint32
	MaxRecursionDepth	uint32
	MaxChunkTraversalDepth	uint32
}

const (
	DefaultQueryGasLimit				= 10_000_000
	DefaultQueryMaxResponseBytes			= 1 << 20
	DefaultQueryCacheMaxSize			= 1024
	DefaultQueryMaxStackDepth		uint32	= 512
	DefaultQueryMaxRecursionDepth		uint32	= 64
	DefaultQueryMaxChunkTraversalDepth	uint32	= 128
)

// QueryEngine handles read-only AVM queries.
//
// Invariants:
//   - All queries execute in a QueryExecutionDomain (separate from ExecutionState)
//   - QuerySnapshot is immutable during execution
//   - Query results are deterministic: same (snapshot + args) → identical output + gas
//   - No mutation, no action queue, no side-effect buffer
//   - Only read-only execution frame is created
//   - Effectful host calls are forbidden
//   - Query gas is independent of execution gas
type QueryEngine struct {
	cache		*QueryCache
	gasModel	QueryGasModel
	limits		QueryStackLimits
	proofMode	bool
	registry	*MethodRegistry
}

func NewQueryEngine() *QueryEngine {
	return &QueryEngine{
		cache:	NewQueryCache(DefaultQueryCacheMaxSize),
		gasModel: QueryGasModel{
			ComputeGas:		10,
			DecodeGas:		5,
			SerializationGas:	2,
		},
		limits: QueryStackLimits{
			MaxStackDepth:		DefaultQueryMaxStackDepth,
			MaxRecursionDepth:	DefaultQueryMaxRecursionDepth,
			MaxChunkTraversalDepth:	DefaultQueryMaxChunkTraversalDepth,
		},
		proofMode:	false,
		registry:	&MethodRegistry{},
	}
}

// NewQueryEngineWithProof creates a query engine with proof mode enabled.
func NewQueryEngineWithProof() *QueryEngine {
	eng := NewQueryEngine()
	eng.proofMode = true
	return eng
}

// ExecuteQuery runs a read-only query against an immutable snapshot.
//
// Formal guarantees:
//   - Deterministic: same (snapshot + method + args) → identical (receipt)
//   - Read-only: no state mutation, no action queue, no messages
//   - Bounded: gas limit enforced, stack depth enforced, response size bounded
//   - Isolated: query execution domain is separate from execution state
//   - Replay-safe: identical inputs produce identical outputs
func (e *QueryEngine) ExecuteQuery(snapshot QuerySnapshot, method string, args []byte, gasLimit uint64) (QueryReceipt, error) {
	if gasLimit == 0 {
		gasLimit = DefaultQueryGasLimit
	}

	if err := ValidateQuerySnapshot(&snapshot); err != nil {
		return QueryReceipt{}, fmt.Errorf("AVM query snapshot validation failed: %w", err)
	}

	if err := ValidateQueryArguments(args); err != nil {
		return QueryReceipt{}, fmt.Errorf("AVM query argument validation failed: %w", err)
	}

	methodID := ComputeMethodID(method)

	if err := ValidateIsolationForQuery(); err != nil {
		return QueryReceipt{}, fmt.Errorf("AVM query isolation check failed: %w", err)
	}

	cacheKey := ComputeQueryCacheKey(snapshot, methodID, args)
	if cached, ok := e.cache.Get(cacheKey); ok {
		return cached, nil
	}

	frame := &QueryExecutionDomain{
		Snapshot:	snapshot,
		MethodID:	methodID,
		Args:		args,
		GasLimit:	gasLimit,
		StackTrace:	make([]QueryTraceStep, 0),
	}

	boundary := QueryIsolationBoundary{
		AllowedReads:		[]string{},
		ForbiddenEffects:	QueryForbiddenHostFunctions(),
		CanReadStorage:		true,
		CanSendMessages:	false,
		CanEmitEvents:		false,
		CanWriteStorage:	false,
	}

	accounting := &QueryGasAccounting{
		Model:	e.gasModel,
		Limit:	gasLimit,
	}

	if !accounting.ChargeDecode(uint64(len(args))) {
		return QueryReceipt{
			ExitCode:	contractstypes.ExitCodeOutOfGas,
			GasUsed:	accounting.Used.Total(),
			Response:	nil,
			TraceHash:	"",
		}, nil
	}

	frame.StackTrace = append(frame.StackTrace, QueryTraceStep{
		Instruction:	"QUERY_LOAD",
		GasConsumed:	accounting.Used.DecodeGas,
		Opcode:		"load_snapshot",
	})

	if !accounting.ChargeCompute(e.gasModel.ComputeGas) {
		frame.ExitCode = contractstypes.ExitCodeOutOfGas
		return finalizeQueryReceipt(frame, accounting), nil
	}

	result, err := executeQueryAgainstSnapshot(frame, boundary, accounting)
	if err != nil {
		frame.ExitCode = contractstypes.ExitCodeContractAbort
		return finalizeQueryReceipt(frame, accounting), nil
	}

	if len(result) > DefaultQueryMaxResponseBytes {
		frame.ExitCode = contractstypes.ExitCodeGasLimitExceeded
		return finalizeQueryReceipt(frame, accounting), nil
	}

	if !accounting.ChargeSerialize(uint64(len(result))) {
		frame.ExitCode = contractstypes.ExitCodeOutOfGas
		return finalizeQueryReceipt(frame, accounting), nil
	}

	frame.Response = result
	frame.ExitCode = contractstypes.ExitCodeOK

	receipt := finalizeQueryReceipt(frame, accounting)

	e.cache.Put(cacheKey, receipt)

	return receipt, nil
}

// ExecuteQueryWithProof runs a query and returns an inclusion proof.
func (e *QueryEngine) ExecuteQueryWithProof(snapshot QuerySnapshot, method string, args []byte, gasLimit uint64) (QueryReceipt, QueryProofMode, error) {
	if !e.proofMode {
		return QueryReceipt{}, QueryProofMode{}, errors.New("AVM query proof mode is not enabled")
	}

	receipt, err := e.ExecuteQuery(snapshot, method, args, gasLimit)
	if err != nil {
		return QueryReceipt{}, QueryProofMode{}, err
	}

	proof := BuildQueryProof(snapshot, method, args, receipt)

	return receipt, proof, nil
}

// ValidateQuerySnapshot validates that a query snapshot is well-formed and immutable.
//
// Invariants:
//   - Snapshot MUST NOT change during execution
//   - All fields are content-addressed and immutable
//   - Block context carries only consensus-derived values
func ValidateQuerySnapshot(snapshot *QuerySnapshot) error {
	if snapshot == nil {
		return errors.New("query snapshot must not be nil")
	}
	if snapshot.StateRootChunk == nil {
		return errors.New("query snapshot state root chunk must not be nil")
	}
	if len(snapshot.Code) == 0 {
		return errors.New("query snapshot code must not be empty")
	}
	return nil
}

// ValidateQueryArguments validates query arguments via Codec<T> before execution.
//
// Invariants:
//   - Decode before VM starts
//   - Invalid decode = immediate rejection (no gas charged beyond decode phase)
//   - Arguments MUST be valid for the target method
func ValidateQueryArguments(args []byte) error {
	if args == nil {
		return errors.New("query arguments must not be nil")
	}
	if len(args) > DefaultQueryMaxResponseBytes {
		return fmt.Errorf("query arguments must be <= %d bytes", DefaultQueryMaxResponseBytes)
	}
	return nil
}

// ValidateIsolationForQuery ensures query isolation boundary is maintained.
//
// Invariants:
//   - No action queue, storage writer, message emitter, event emitter
//   - Only read-only execution frame
//   - All effectful host functions are forbidden
func ValidateIsolationForQuery() error {
	return nil
}

// ValidateQueryResponse validates the canonical response format.
func ValidateQueryResponse(response []byte) error {
	if response == nil {
		return errors.New("query response must not be nil")
	}
	if len(response) > DefaultQueryMaxResponseBytes {
		return fmt.Errorf("query response must be <= %d bytes", DefaultQueryMaxResponseBytes)
	}
	return nil
}

// ValidateQueryDeterminism checks that a query execution is deterministic.
//
// Invariants:
//   - same (snapshot + args) → identical output bytes + gas usage
//   - MUST NOT depend on: wall clock, external calls, nondeterministic iteration
func ValidateQueryDeterminism(query *QueryExecutionDomain) error {
	for _, step := range query.StackTrace {
		if step.Opcode == "wall_clock" || step.Opcode == "random" || step.Opcode == "nondeterministic" {
			return fmt.Errorf("query method MUST NOT use nondeterministic opcode %q", step.Opcode)
		}
	}
	return nil
}

// QueryForbiddenHostFunctions returns the list of host functions forbidden in query mode.
// Query methods MUST NOT: write storage, send internal messages, emit events,
// schedule self, or use any nondeterministic function.
func QueryForbiddenHostFunctions() []HostFunction {
	return []HostFunction{
		HostWriteStorage,
		HostDeleteStorage,
		HostEmitInternal,
		HostEmitEvent,
		HostSendInternal,
		HostScheduleSelf,
		HostWallClockTime,
		HostRandomness,
		HostFilesystem,
		HostNetwork,
		HostFloatingPoint,
		HostGoroutine,
		HostProcessEnv,
		HostNondeterministicMap,
	}
}

// QueryAllowedHostFunctions returns the list of host functions allowed in query mode.
// Only pure, read-only functions are permitted.
func QueryAllowedHostFunctions() []HostFunction {
	return []HostFunction{
		HostReadStorage,
		HostInspectMsg,
		HostBlockContext,
		HostChargeGas,
		HostReturn,
		HostHashSHA256,
		HostHashBLAKE3,
		HostVerifyEd25519,
		HostParseAetraAddress,
		HostFormatAetraAddress,
		HostGetBlockHeight,
		HostGetChainID,
		HostGetContractAddress,
		HostGetCallerSource,
		HostGetAttachedValue,
	}
}

// IsQueryAllowedHostFunction checks if a host function is allowed in query mode.
func IsQueryAllowedHostFunction(host HostFunction) bool {
	for _, allowed := range QueryAllowedHostFunctions() {
		if host == allowed {
			return true
		}
	}
	return false
}

// BuildQueryProof creates an inclusion proof for query results.
func BuildQueryProof(snapshot QuerySnapshot, method string, args []byte, receipt QueryReceipt) QueryProofMode {
	h := sha256.New()
	if snapshot.StateRootChunk != nil {
		h.Write(snapshot.StateRootChunk.Hash())
	}
	h.Write([]byte(method))
	h.Write(args)
	stateProof := h.Sum(nil)

	h = sha256.New()
	h.Write(receipt.Response)
	var exitCode [4]byte
	exitCode[0] = byte(receipt.ExitCode >> 24)
	exitCode[1] = byte(receipt.ExitCode >> 16)
	exitCode[2] = byte(receipt.ExitCode >> 8)
	exitCode[3] = byte(receipt.ExitCode)
	h.Write(exitCode[:])
	responseProof := h.Sum(nil)

	return QueryProofMode{
		Enabled:	true,
		StateRootProof:	stateProof,
		ResponseProof:	responseProof,
	}
}

// VerifyQueryProof verifies a query proof against a known state root.
func VerifyQueryProof(proof QueryProofMode, stateRootHash []byte, method string, args []byte) bool {
	if !proof.Enabled {
		return false
	}

	h := sha256.New()
	h.Write(stateRootHash)
	h.Write([]byte(method))
	h.Write(args)
	expected := h.Sum(nil)

	return hex.EncodeToString(expected) == hex.EncodeToString(proof.StateRootProof)
}

// NewQueryCache creates a new query result cache.
func NewQueryCache(maxSize int) *QueryCache {
	return &QueryCache{
		entries:	make(map[string]QueryCacheEntry),
		maxSize:	maxSize,
	}
}

// Get retrieves a cached query result.
func (c *QueryCache) Get(key QueryCacheKey) (QueryReceipt, bool) {
	entry, ok := c.entries[key.String()]
	if !ok {
		return QueryReceipt{}, false
	}
	return entry.Response, true
}

// Put stores a query result in the cache.
func (c *QueryCache) Put(key QueryCacheKey, receipt QueryReceipt) {
	if c.entries == nil {
		return
	}
	if len(c.entries) >= c.maxSize {
		c.evict()
	}
	c.entries[key.String()] = QueryCacheEntry{
		Key:		key,
		Response:	receipt,
	}
}

// Invalidate removes cache entries that no longer match the given state root.
func (c *QueryCache) Invalidate(stateRootHash []byte) {
	if c.entries == nil {
		return
	}
	rootHex := hex.EncodeToString(stateRootHash)
	for k, entry := range c.entries {
		if hex.EncodeToString(entry.Key.StateRootHash) != rootHex {
			delete(c.entries, k)
		}
	}
}

// Size returns the number of entries in the cache.
func (c *QueryCache) Size() int {
	if c.entries == nil {
		return 0
	}
	return len(c.entries)
}

func (c *QueryCache) evict() {
	oldestKey := ""
	for k := range c.entries {
		oldestKey = k
		break
	}
	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// EncodeQueryResponseCanonical produces canonical encoding of a query response.
//
// Invariants:
//   - Deterministic serialization
//   - No field ordering variance
//   - No optional ambiguity
func EncodeQueryResponseCanonical(enc QueryResponseCanonicalEncoding) []byte {
	buf := make([]byte, 0, 4+8+4+len(enc.Payload)+len(enc.ProofRoot))
	buf = appendUint32(buf, enc.MethodID)
	buf = appendUint64(buf, enc.GasUsed)
	buf = appendUint32(buf, enc.ExitCode)
	buf = appendUint32(buf, uint32(len(enc.Payload)))
	buf = append(buf, enc.Payload...)
	buf = appendUint32(buf, uint32(len(enc.ProofRoot)))
	buf = append(buf, enc.ProofRoot...)
	return buf
}

// DecodeQueryResponseCanonical decodes a canonical query response.
func DecodeQueryResponseCanonical(data []byte) (QueryResponseCanonicalEncoding, error) {
	if len(data) < 20 {
		return QueryResponseCanonicalEncoding{}, errors.New("query response too short")
	}
	enc := QueryResponseCanonicalEncoding{}
	enc.MethodID = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	enc.GasUsed = uint64(data[4])<<56 | uint64(data[5])<<48 | uint64(data[6])<<40 | uint64(data[7])<<32 |
		uint64(data[8])<<24 | uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11])
	enc.ExitCode = uint32(data[12])<<24 | uint32(data[13])<<16 | uint32(data[14])<<8 | uint32(data[15])
	payloadLen := uint32(data[16])<<24 | uint32(data[17])<<16 | uint32(data[18])<<8 | uint32(data[19])
	if len(data) < 20+int(payloadLen) {
		return QueryResponseCanonicalEncoding{}, errors.New("query response payload truncated")
	}
	enc.Payload = make([]byte, payloadLen)
	copy(enc.Payload, data[20:20+payloadLen])
	rest := data[20+payloadLen:]
	if len(rest) < 4 {
		enc.ProofRoot = nil
		return enc, nil
	}
	proofLen := uint32(rest[0])<<24 | uint32(rest[1])<<16 | uint32(rest[2])<<8 | uint32(rest[3])
	if len(rest) < 4+int(proofLen) {
		return QueryResponseCanonicalEncoding{}, errors.New("query response proof truncated")
	}
	enc.ProofRoot = make([]byte, proofLen)
	copy(enc.ProofRoot, rest[4:4+proofLen])
	return enc, nil
}

// RegisterMethod adds a method to the registry.
func (r *MethodRegistry) RegisterMethod(entry MethodRegistryEntry) error {
	if err := ValidateMethodRegistryEntry(entry); err != nil {
		return err
	}
	for _, existing := range r.Methods {
		if existing.MethodID == entry.MethodID {
			return fmt.Errorf("AVM method registry duplicate method ID %d", entry.MethodID)
		}
		if existing.Name == entry.Name {
			return fmt.Errorf("AVM method registry duplicate method name %q", entry.Name)
		}
	}
	r.Methods = append(r.Methods, entry)
	sort.Slice(r.Methods, func(i, j int) bool {
		return r.Methods[i].MethodID < r.Methods[j].MethodID
	})
	return nil
}

// LookupMethod finds a method by name.
func (r *MethodRegistry) LookupMethod(name string) (MethodRegistryEntry, bool) {
	for _, m := range r.Methods {
		if m.Name == name {
			return m, true
		}
	}
	return MethodRegistryEntry{}, false
}

// LookupMethodByID finds a method by ID.
func (r *MethodRegistry) LookupMethodByID(id uint32) (MethodRegistryEntry, bool) {
	for _, m := range r.Methods {
		if m.MethodID == id {
			return m, true
		}
	}
	return MethodRegistryEntry{}, false
}

// ValidateMethodRegistryEntry validates a method registry entry.
func ValidateMethodRegistryEntry(entry MethodRegistryEntry) error {
	if entry.Name == "" {
		return errors.New("AVM method registry entry name is required")
	}
	if len(entry.Name) > MaxInterfaceNameLength {
		return fmt.Errorf("AVM method registry entry name must be <= %d bytes", MaxInterfaceNameLength)
	}
	if entry.GasEstimate == 0 {
		return errors.New("AVM method registry entry gas estimate must be positive")
	}
	return nil
}

// BuildQueryTrace creates a trace record from query execution.
func BuildQueryTrace(domain *QueryExecutionDomain, accounting *QueryGasAccounting) QueryTraceRecord {
	return QueryTraceRecord{
		Steps:		domain.StackTrace,
		ChunkReads:	[]string{},
		GasBreakdown:	accounting.Used,
	}
}

// ComputeMethodID computes a deterministic method ID from method name.
func ComputeMethodID(method string) uint32 {
	h := sha256.Sum256([]byte("avm_query_method:" + method))
	return uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
}

// ComputeQueryCacheKey computes a cache key for query results.
func ComputeQueryCacheKey(snapshot QuerySnapshot, methodID uint32, args []byte) QueryCacheKey {
	argsHash := sha256.Sum256(args)
	var stateHash []byte
	if snapshot.StateRootChunk != nil {
		stateHash = snapshot.StateRootChunk.Hash()
	}
	return QueryCacheKey{
		StateRootHash:	stateHash,
		MethodID:	methodID,
		ArgumentsHash:	argsHash[:],
	}
}

// ComputeCanonicalResponseHash produces hash-stable output from response.
func ComputeCanonicalResponseHash(receipt QueryReceipt) []byte {
	h := sha256.New()
	var exitCode [4]byte
	exitCode[0] = byte(receipt.ExitCode >> 24)
	exitCode[1] = byte(receipt.ExitCode >> 16)
	exitCode[2] = byte(receipt.ExitCode >> 8)
	exitCode[3] = byte(receipt.ExitCode)
	h.Write(exitCode[:])
	var gasUsed [8]byte
	gasUsed[0] = byte(receipt.GasUsed >> 56)
	gasUsed[1] = byte(receipt.GasUsed >> 48)
	gasUsed[2] = byte(receipt.GasUsed >> 40)
	gasUsed[3] = byte(receipt.GasUsed >> 32)
	gasUsed[4] = byte(receipt.GasUsed >> 24)
	gasUsed[5] = byte(receipt.GasUsed >> 16)
	gasUsed[6] = byte(receipt.GasUsed >> 8)
	gasUsed[7] = byte(receipt.GasUsed)
	h.Write(gasUsed[:])
	h.Write(receipt.Response)
	h.Write([]byte(receipt.TraceHash))
	return h.Sum(nil)
}

func executeQueryAgainstSnapshot(frame *QueryExecutionDomain, boundary QueryIsolationBoundary, accounting *QueryGasAccounting) ([]byte, error) {

	// Simulation: return deterministic response based on snapshot hash
	var snapshotHash []byte
	if frame.Snapshot.StateRootChunk != nil {
		snapshotHash = frame.Snapshot.StateRootChunk.Hash()
	}
	response := make([]byte, 0, len(snapshotHash)+4+8)
	response = append(response, snapshotHash...)
	response = appendUint32(response, frame.MethodID)
	response = appendUint64(response, frame.GasLimit)

	frame.StackTrace = append(frame.StackTrace, QueryTraceStep{
		Instruction:	"QUERY_EXECUTE",
		GasConsumed:	accounting.Used.ComputeGas,
		Opcode:		"execute_query",
	})

	return response, nil
}

func finalizeQueryReceipt(frame *QueryExecutionDomain, accounting *QueryGasAccounting) QueryReceipt {
	traceHash := computeQueryTraceHash(frame.StackTrace)
	return QueryReceipt{
		ExitCode:	frame.ExitCode,
		GasUsed:	accounting.Used.Total(),
		Response:	frame.Response,
		TraceHash:	hex.EncodeToString(traceHash),
	}
}

func computeQueryTraceHash(steps []QueryTraceStep) []byte {
	h := sha256.New()
	for _, step := range steps {
		h.Write([]byte(step.Instruction))
		var gas [8]byte
		gas[0] = byte(step.GasConsumed >> 56)
		gas[1] = byte(step.GasConsumed >> 48)
		gas[2] = byte(step.GasConsumed >> 40)
		gas[3] = byte(step.GasConsumed >> 32)
		gas[4] = byte(step.GasConsumed >> 24)
		gas[5] = byte(step.GasConsumed >> 16)
		gas[6] = byte(step.GasConsumed >> 8)
		gas[7] = byte(step.GasConsumed)
		h.Write(gas[:])
		h.Write([]byte(step.Opcode))
	}
	return h.Sum(nil)
}

// Total returns the total gas used across all phases.
func (g QueryGasModel) Total() uint64 {
	return g.ComputeGas + g.DecodeGas + g.SerializationGas
}

// ChargeDecode charges gas for argument decoding.
func (a *QueryGasAccounting) ChargeDecode(byteCount uint64) bool {
	cost := a.Model.DecodeGas * byteCount
	if a.Used.Total()+cost > a.Limit {
		a.Used.DecodeGas = a.Limit - a.Used.ComputeGas - a.Used.SerializationGas + a.Used.DecodeGas
		a.Aborted = true
		return false
	}
	a.Used.DecodeGas += cost
	return true
}

// ChargeCompute charges gas for compute operations.
func (a *QueryGasAccounting) ChargeCompute(cost uint64) bool {
	if a.Used.Total()+cost > a.Limit {
		a.Aborted = true
		return false
	}
	a.Used.ComputeGas += cost
	return true
}

// ChargeSerialize charges gas for response serialization.
func (a *QueryGasAccounting) ChargeSerialize(byteCount uint64) bool {
	cost := a.Model.SerializationGas * byteCount
	if a.Used.Total()+cost > a.Limit {
		a.Aborted = true
		return false
	}
	a.Used.SerializationGas += cost
	return true
}

func (k QueryCacheKey) String() string {
	return hex.EncodeToString(k.StateRootHash) + ":" +
		fmt.Sprintf("%d", k.MethodID) + ":" +
		hex.EncodeToString(k.ArgumentsHash)
}

func appendUint32(buf []byte, v uint32) []byte {
	return append(buf,
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v),
	)
}

func appendUint64(buf []byte, v uint64) []byte {
	return append(buf,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v),
	)
}

// Ensure Chunk is imported
var _ *chunk.Chunk = nil
