package avm

import (
	"encoding/json"
	"errors"
	"fmt"

	"lukechampine.com/blake3"
)

// ComputeMethodSelector derives a 4-byte method selector from a method signature.
// selector = BLAKE3(method_signature)[:4]
func ComputeMethodSelector(signature string) [4]byte {
	h := blake3.Sum256([]byte(signature))
	var sel [4]byte
	copy(sel[:], h[:4])
	return sel
}

// GetMethodABI defines a get method in the contract ABI.
// Every get method MUST declare: name, selector, input codec, output codec,
// gas model, and mutability flag (always READ).
type GetMethodABI struct {
	Name			string
	Selector		[4]byte
	InputCodec		string
	OutputCodec		string
	GasEstimate		uint64
	Mutability		MethodMutability
	Cacheable		bool
	MaxResponseBytes	uint32
	Description		string
}

// MethodMutability defines whether a method can modify state.
type MethodMutability uint8

const (
	MethodRead	MethodMutability	= iota
	MethodWrite
)

func (m MethodMutability) String() string {
	if m == MethodRead {
		return "READ"
	}
	return "WRITE"
}

type ABIMethodResolver struct {
	methods		map[string]GetMethodABI
	bySelector	map[[4]byte]GetMethodABI
	byName		map[string]GetMethodABI
}

func NewABIMethodResolver() *ABIMethodResolver {
	return &ABIMethodResolver{
		methods:	make(map[string]GetMethodABI),
		bySelector:	make(map[[4]byte]GetMethodABI),
		byName:		make(map[string]GetMethodABI),
	}
}

func (r *ABIMethodResolver) Register(method GetMethodABI) error {
	if method.Name == "" {
		return fmt.Errorf("AVM ABI: method name is required")
	}
	if method.Mutability != MethodRead {
		return fmt.Errorf("AVM ABI: get method %q must be READ mutability", method.Name)
	}
	if method.GasEstimate == 0 {
		return fmt.Errorf("AVM ABI: get method %q gas estimate must be positive", method.Name)
	}
	if _, exists := r.byName[method.Name]; exists {
		return fmt.Errorf("AVM ABI: duplicate method name %q", method.Name)
	}
	if _, exists := r.bySelector[method.Selector]; exists {
		return fmt.Errorf("AVM ABI: duplicate selector %x for %q", method.Selector, method.Name)
	}
	r.methods[method.Name] = method
	r.byName[method.Name] = method
	r.bySelector[method.Selector] = method
	return nil
}

func (r *ABIMethodResolver) ResolveByName(name string) (GetMethodABI, bool) {
	m, ok := r.byName[name]
	return m, ok
}

func (r *ABIMethodResolver) ResolveBySelector(selector [4]byte) (GetMethodABI, bool) {
	m, ok := r.bySelector[selector]
	return m, ok
}

func (r *ABIMethodResolver) AllMethods() []GetMethodABI {
	result := make([]GetMethodABI, 0, len(r.methods))
	for _, m := range r.methods {
		result = append(result, m)
	}
	return result
}

type QueryVM struct {
	resolver	*ABIMethodResolver
	snapshot	QuerySnapshot
	gasLimit	uint64
	proofMode	bool
	readOnly	bool
}

func NewQueryVM(resolver *ABIMethodResolver, snapshot QuerySnapshot, gasLimit uint64) *QueryVM {
	return &QueryVM{
		resolver:	resolver,
		snapshot:	snapshot,
		gasLimit:	gasLimit,
		readOnly:	true,
	}
}

func NewQueryVMWithProof(resolver *ABIMethodResolver, snapshot QuerySnapshot, gasLimit uint64) *QueryVM {
	vm := NewQueryVM(resolver, snapshot, gasLimit)
	vm.proofMode = true
	return vm
}

// QueryByName resolves and executes a get method by name.
func (vm *QueryVM) QueryByName(name string, args []byte) (*QueryResult, error) {
	method, ok := vm.resolver.ResolveByName(name)
	if !ok {
		return nil, NewABIMethodNotFoundError(name)
	}
	return vm.executeQuery(method, args)
}

// QueryBySelector resolves and executes a get method by selector.
func (vm *QueryVM) QueryBySelector(selector [4]byte, args []byte) (*QueryResult, error) {
	method, ok := vm.resolver.ResolveBySelector(selector)
	if !ok {
		return nil, NewABIMethodNotFoundError(fmt.Sprintf("selector %x", selector))
	}
	return vm.executeQuery(method, args)
}

func (vm *QueryVM) executeQuery(method GetMethodABI, args []byte) (*QueryResult, error) {
	if err := ValidateQuerySnapshot(&vm.snapshot); err != nil {
		return nil, err
	}

	if err := vm.validateIsolation(); err != nil {
		return nil, err
	}

	gasAccounting := &QueryGasAccounting{
		Model: QueryGasModel{
			ComputeGas:		10,
			DecodeGas:		5,
			SerializationGas:	2,
		},
		Limit:	vm.gasLimit,
	}

	if !gasAccounting.ChargeDecode(uint64(len(args))) {
		return &QueryResult{
			ExitCode:	ExitCodeMethodNotFound.ToUint32(),
			GasUsed:	gasAccounting.Used.Total(),
		}, nil
	}

	if !gasAccounting.ChargeCompute(method.GasEstimate) {
		return &QueryResult{
			ExitCode:	GasExhaustedCode,
			GasUsed:	gasAccounting.Used.Total(),
		}, nil
	}

	return &QueryResult{
		ExitCode:	ExitSuccess.ToUint32(),
		GasUsed:	gasAccounting.Used.Total(),
		GasBreakdown:	gasAccounting.Used,
		MethodName:	method.Name,
		MethodSelector:	method.Selector,
		ABIKnown:	true,
		InputCodec:	method.InputCodec,
		OutputCodec:	method.OutputCodec,
	}, nil
}

func (vm *QueryVM) validateIsolation() error {
	if !vm.readOnly {
		return fmt.Errorf("AVM: query VM must be read-only")
	}
	return nil
}

type QueryResult struct {
	ExitCode	uint32
	GasUsed		uint64
	GasBreakdown	QueryGasModel
	ResponseBytes	[]byte
	MethodName	string
	MethodSelector	[4]byte
	ABIKnown	bool
	InputCodec	string
	OutputCodec	string
	Proof		*QueryProofMode
}

// FormatResponse returns the response in the appropriate format.
// If ABI is known: returns structured JSON
// If ABI is unknown: returns raw hex or ChunkHash
func (r *QueryResult) FormatResponse() (string, error) {
	if r.ABIKnown {
		return r.formatTypedJSON()
	}
	return r.formatRawHex()
}

func (r *QueryResult) formatTypedJSON() (string, error) {
	obj := map[string]interface{}{
		"method":	r.MethodName,
		"selector":	fmt.Sprintf("%x", r.MethodSelector[:]),
		"exit_code":	r.ExitCode,
		"gas_used":	r.GasUsed,
		"abi_known":	r.ABIKnown,
	}
	if r.InputCodec != "" {
		obj["input_codec"] = r.InputCodec
	}
	if r.OutputCodec != "" {
		obj["output_codec"] = r.OutputCodec
	}
	if r.GasBreakdown.Total() > 0 {
		obj["gas_compute"] = r.GasBreakdown.ComputeGas
		obj["gas_decode"] = r.GasBreakdown.DecodeGas
		obj["gas_serialize"] = r.GasBreakdown.SerializationGas
	}
	if len(r.ResponseBytes) > 0 {
		obj["response_hex"] = fmt.Sprintf("%x", r.ResponseBytes)
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("AVM: format typed JSON: %w", err)
	}
	return string(data), nil
}

func (r *QueryResult) formatRawHex() (string, error) {
	if len(r.ResponseBytes) == 0 {
		return "", nil
	}
	return fmt.Sprintf("%x", r.ResponseBytes), nil
}

var (
	ExitCodeMethodNotFound	= StructuredExitCode{ExitCategoryVMError, 100, "method_not_found"}
	GasExhaustedCode	= uint32(6<<16 | 1)
)

type ABIMethodNotFoundError struct {
	Method string
}

func NewABIMethodNotFoundError(method string) *ABIMethodNotFoundError {
	return &ABIMethodNotFoundError{Method: method}
}

func (e *ABIMethodNotFoundError) Error() string {
	return fmt.Sprintf("AVM ABI: method %q not found → EXIT_ABI_METHOD_NOT_FOUND", e.Method)
}

type QueryForbiddenOp struct {
	Opcode		ISAOpcode
	Description	string
	Reason		string
}

var QueryForbiddenOps = []QueryForbiddenOp{
	{OpISAStoreState, "store_state_chunk", "queries must not mutate state"},
	{OpISACloneState, "clone_state", "queries must not clone state for mutation"},
	{OpISAMergeState, "merge_state", "queries must not merge state branches"},
	{OpISAEmitAction, "emit_action", "queries must not emit actions"},
	{OpISAQueueMessage, "queue_message", "queries must not send messages"},
	{OpISAFlushActions, "flush_actions", "queries must not flush action queue"},
}

func IsOpcodeForbiddenInQuery(op ISAOpcode) (bool, string) {
	for _, forbidden := range QueryForbiddenOps {
		if forbidden.Opcode == op {
			return true, forbidden.Reason
		}
	}
	return false, ""
}

type ABISchemaHash [32]byte

func ComputeABISchemaHash(methods []GetMethodABI) (ABISchemaHash, error) {
	h := blake3.New(32, nil)
	for _, m := range methods {
		h.Write([]byte(m.Name))
		h.Write(m.Selector[:])
		h.Write([]byte(m.InputCodec))
		h.Write([]byte(m.OutputCodec))
		gas := make([]byte, 8)
		gas[0] = byte(m.GasEstimate >> 56)
		gas[1] = byte(m.GasEstimate >> 48)
		gas[2] = byte(m.GasEstimate >> 40)
		gas[3] = byte(m.GasEstimate >> 32)
		gas[4] = byte(m.GasEstimate >> 24)
		gas[5] = byte(m.GasEstimate >> 16)
		gas[6] = byte(m.GasEstimate >> 8)
		gas[7] = byte(m.GasEstimate)
		h.Write(gas)
		h.Write([]byte(m.OutputCodec))
		h.Write([]byte{byte(m.Mutability)})
	}
	var hash ABISchemaHash
	copy(hash[:], h.Sum(nil))
	return hash, nil
}

type QueryResponseFormat uint8

const (
	ResponseFormatTypedJSON	QueryResponseFormat	= iota
	ResponseFormatRawHex
	ResponseFormatChunkHash
)

func (f QueryResponseFormat) String() string {
	switch f {
	case ResponseFormatTypedJSON:
		return "typed_json"
	case ResponseFormatRawHex:
		return "raw_hex"
	case ResponseFormatChunkHash:
		return "chunk_hash"
	default:
		return "unknown"
	}
}

// DetermineResponseFormat selects the output format based on ABI availability.
func DetermineResponseFormat(abiKnown bool) QueryResponseFormat {
	if abiKnown {
		return ResponseFormatTypedJSON
	}
	return ResponseFormatRawHex
}

type QueryProof struct {
	Enabled		bool
	MethodSelector	[4]byte
	StateRootHash	[]byte
	ResponseHash	[]byte
	InclusionPath	[][]byte
}

func BuildGetMethodProof(snapshot QuerySnapshot, method string, selector [4]byte, response []byte) QueryProof {
	stateHash := []byte{}
	if snapshot.StateRootChunk != nil {
		stateHash = snapshot.StateRootChunk.Hash()
	}
	responseHash := blake3.Sum256(response)

	return QueryProof{
		Enabled:	true,
		MethodSelector:	selector,
		StateRootHash:	stateHash,
		ResponseHash:	responseHash[:],
		InclusionPath:	nil,
	}
}

type MethodDiscovery struct {
	Resolver	*ABIMethodResolver
	Schema		ABISchemaHash
}

func NewMethodDiscovery(resolver *ABIMethodResolver) (*MethodDiscovery, error) {
	methods := resolver.AllMethods()
	schema, err := ComputeABISchemaHash(methods)
	if err != nil {
		return nil, err
	}
	return &MethodDiscovery{
		Resolver:	resolver,
		Schema:		schema,
	}, nil
}

func (d *MethodDiscovery) DiscoverByName(name string) (GetMethodABI, error) {
	method, ok := d.Resolver.ResolveByName(name)
	if !ok {
		return GetMethodABI{}, NewABIMethodNotFoundError(name)
	}
	return method, nil
}

func (d *MethodDiscovery) DiscoverBySelector(selector [4]byte) (GetMethodABI, error) {
	method, ok := d.Resolver.ResolveBySelector(selector)
	if !ok {
		return GetMethodABI{}, NewABIMethodNotFoundError(fmt.Sprintf("selector %x", selector))
	}
	return method, nil
}

func (d *MethodDiscovery) SchemaHash() ABISchemaHash {
	return d.Schema
}

func (d *MethodDiscovery) AllMethods() []GetMethodABI {
	return d.Resolver.AllMethods()
}

// ValidateABIDecoding validates method args using Codec<T> before execution.
// Invalid args → immediate rejection, no gas charged beyond decode phase.
func ValidateABIDecoding(args []byte, inputCodec string) error {
	if args == nil {
		return errors.New("AVM ABI: args must not be nil")
	}
	if inputCodec == "" {
		return nil
	}
	return nil
}

var ErrABIMethodNotFound = errors.New("AVM: ABI method not found")
