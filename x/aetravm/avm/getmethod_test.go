package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func TestComputeMethodSelector(t *testing.T) {
	sel1 := ComputeMethodSelector("get_balance(address)")
	sel2 := ComputeMethodSelector("get_balance(address)")
	if sel1 != sel2 {
		t.Error("same signature should produce same selector")
	}

	sel3 := ComputeMethodSelector("get_total_supply()")
	if sel1 == sel3 {
		t.Error("different signatures should produce different selectors")
	}

	if len(sel1) != 4 {
		t.Errorf("selector should be 4 bytes, got %d", len(sel1))
	}
}

func TestComputeMethodSelectorDeterministic(t *testing.T) {
	for i := 0; i < 10; i++ {
		sel := ComputeMethodSelector("get_balance(address)")
		if sel != [4]byte{sel[0], sel[1], sel[2], sel[3]} {
			t.Error("selector computation should be deterministic")
			break
		}
	}
}

func TestABIMethodResolverRegisterAndLookup(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("get_balance(address)")

	method := GetMethodABI{
		Name:		"get_balance",
		Selector:	sel,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasEstimate:	1000,
		Mutability:	MethodRead,
	}

	err := resolver.Register(method)
	if err != nil {
		t.Fatalf("register method: %v", err)
	}

	found, ok := resolver.ResolveByName("get_balance")
	if !ok {
		t.Error("should find method by name")
	}
	if found.Name != "get_balance" {
		t.Errorf("expected get_balance, got %s", found.Name)
	}

	found2, ok := resolver.ResolveBySelector(sel)
	if !ok {
		t.Error("should find method by selector")
	}
	if found2.Name != "get_balance" {
		t.Errorf("expected get_balance, got %s", found2.Name)
	}
}

func TestABIMethodResolverUnknownMethod(t *testing.T) {
	resolver := NewABIMethodResolver()

	_, ok := resolver.ResolveByName("nonexistent")
	if ok {
		t.Error("should not find unknown method")
	}

	var sel [4]byte
	_, ok = resolver.ResolveBySelector(sel)
	if ok {
		t.Error("should not find unknown selector")
	}
}

func TestABIMethodResolverDuplicateName(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("get_balance(address)")

	method := GetMethodABI{
		Name:		"get_balance",
		Selector:	sel,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasEstimate:	1000,
		Mutability:	MethodRead,
	}
	resolver.Register(method)

	err := resolver.Register(method)
	if err == nil {
		t.Error("should reject duplicate method name")
	}
}

func TestABIMethodResolverWriteMutabilityRejected(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("transfer(address,uint64)")

	method := GetMethodABI{
		Name:		"transfer",
		Selector:	sel,
		InputCodec:	"address,uint64",
		OutputCodec:	"bool",
		GasEstimate:	5000,
		Mutability:	MethodWrite,
	}

	err := resolver.Register(method)
	if err == nil {
		t.Error("should reject WRITE mutability for get method")
	}
}

func TestABIMethodResolverZeroGasEstimate(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("get_balance(address)")

	method := GetMethodABI{
		Name:		"get_balance",
		Selector:	sel,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasEstimate:	0,
		Mutability:	MethodRead,
	}

	err := resolver.Register(method)
	if err == nil {
		t.Error("should reject zero gas estimate")
	}
}

func testStateRoot() *chunk.Chunk {
	m := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeSystem).SetData([]byte{}, 0).Build()
	m, _ = m.Put([]byte("__init__"), b)
	return m.Root()
}

func TestQueryVMQueryByName(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("get_balance(address)")
	resolver.Register(GetMethodABI{
		Name:		"get_balance",
		Selector:	sel,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasEstimate:	1000,
		Mutability:	MethodRead,
	})

	snapshot := QuerySnapshot{
		StateRootChunk:	testStateRoot(),
		Code:		[]byte("code"),
	}

	vm := NewQueryVM(resolver, snapshot, 100000)
	result, err := vm.QueryByName("get_balance", []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("query by name: %v", err)
	}
	if result.ExitCode != ExitSuccess.ToUint32() {
		t.Errorf("expected success exit code, got %d", result.ExitCode)
	}
	if !result.ABIKnown {
		t.Error("ABI should be known")
	}
	if result.MethodName != "get_balance" {
		t.Errorf("expected method get_balance, got %s", result.MethodName)
	}
}

func TestQueryVMQueryBySelector(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("get_balance(address)")
	resolver.Register(GetMethodABI{
		Name:		"get_balance",
		Selector:	sel,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasEstimate:	1000,
		Mutability:	MethodRead,
	})

	snapshot := QuerySnapshot{
		StateRootChunk:	testStateRoot(),
		Code:		[]byte("code"),
	}

	vm := NewQueryVM(resolver, snapshot, 100000)
	result, err := vm.QueryBySelector(sel, []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("query by selector: %v", err)
	}
	if result.MethodName != "get_balance" {
		t.Errorf("expected method get_balance, got %s", result.MethodName)
	}
}

func TestQueryVMUnknownMethodRejected(t *testing.T) {
	resolver := NewABIMethodResolver()
	snapshot := QuerySnapshot{
		StateRootChunk:	testStateRoot(),
		Code:		[]byte("code"),
	}

	vm := NewQueryVM(resolver, snapshot, 100000)
	_, err := vm.QueryByName("nonexistent", nil)
	if err == nil {
		t.Error("unknown method should be rejected")
	}
}

func TestForbiddenOpsInQuery(t *testing.T) {
	forbiddenOps := []ISAOpcode{
		OpISAStoreState,
		OpISAEmitAction,
		OpISAQueueMessage,
		OpISAFlushActions,
	}

	for _, op := range forbiddenOps {
		isForbidden, reason := IsOpcodeForbiddenInQuery(op)
		if !isForbidden {
			t.Errorf("opcode %v should be forbidden in query context", op)
		}
		if reason == "" {
			t.Errorf("opcode %v should have a reason", op)
		}
	}

	allowedOps := []ISAOpcode{
		OpISALoadState,
		OpISAGetCaller,
		OpISAHashChunk,
		OpISAAdd,
		OpISAEq,
	}

	for _, op := range allowedOps {
		isForbidden, _ := IsOpcodeForbiddenInQuery(op)
		if isForbidden {
			t.Errorf("opcode %v should NOT be forbidden in query context", op)
		}
	}
}

func TestABISchemaHashDeterministic(t *testing.T) {
	methods := []GetMethodABI{
		{
			Name:		"get_balance",
			Selector:	ComputeMethodSelector("get_balance(address)"),
			InputCodec:	"address",
			OutputCodec:	"uint64",
			GasEstimate:	1000,
			Mutability:	MethodRead,
		},
		{
			Name:		"get_total_supply",
			Selector:	ComputeMethodSelector("get_total_supply()"),
			InputCodec:	"",
			OutputCodec:	"uint64",
			GasEstimate:	500,
			Mutability:	MethodRead,
		},
	}

	h1, err := ComputeABISchemaHash(methods)
	if err != nil {
		t.Fatalf("compute hash: %v", err)
	}
	h2, err := ComputeABISchemaHash(methods)
	if err != nil {
		t.Fatalf("compute hash: %v", err)
	}
	if h1 != h2 {
		t.Error("same methods should produce same schema hash")
	}
}

func TestABISchemaHashChangesOnMutation(t *testing.T) {
	methods1 := []GetMethodABI{
		{
			Name:		"get_balance",
			Selector:	ComputeMethodSelector("get_balance(address)"),
			InputCodec:	"address",
			OutputCodec:	"uint64",
			GasEstimate:	1000,
			Mutability:	MethodRead,
		},
	}
	methods2 := []GetMethodABI{
		{
			Name:		"get_balance",
			Selector:	ComputeMethodSelector("get_balance(address)"),
			InputCodec:	"address",
			OutputCodec:	"uint64",
			GasEstimate:	2000,
			Mutability:	MethodRead,
		},
	}

	h1, _ := ComputeABISchemaHash(methods1)
	h2, _ := ComputeABISchemaHash(methods2)
	if h1 == h2 {
		t.Error("different ABI should produce different hash")
	}
}

func TestMethodDiscovery(t *testing.T) {
	resolver := NewABIMethodResolver()
	sel := ComputeMethodSelector("get_balance(address)")
	resolver.Register(GetMethodABI{
		Name:		"get_balance",
		Selector:	sel,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasEstimate:	1000,
		Mutability:	MethodRead,
	})

	discovery, err := NewMethodDiscovery(resolver)
	if err != nil {
		t.Fatalf("create discovery: %v", err)
	}

	methods := discovery.AllMethods()
	if len(methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(methods))
	}

	method, err := discovery.DiscoverByName("get_balance")
	if err != nil {
		t.Fatalf("discover by name: %v", err)
	}
	if method.Name != "get_balance" {
		t.Errorf("expected get_balance, got %s", method.Name)
	}

	_, err = discovery.DiscoverByName("nonexistent")
	if err == nil {
		t.Error("unknown method should not be found")
	}
}

func TestDetermineResponseFormat(t *testing.T) {
	if DetermineResponseFormat(true) != ResponseFormatTypedJSON {
		t.Error("ABI known should use typed JSON")
	}
	if DetermineResponseFormat(false) != ResponseFormatRawHex {
		t.Error("ABI unknown should use raw hex")
	}
}

func TestQueryResultFormatTypedJSON(t *testing.T) {
	result := &QueryResult{
		ExitCode:	0,
		GasUsed:	500,
		MethodName:	"get_balance",
		ABIKnown:	true,
		InputCodec:	"address",
		OutputCodec:	"uint64",
		GasBreakdown:	QueryGasModel{ComputeGas: 300, DecodeGas: 100, SerializationGas: 100},
	}

	formatted, err := result.FormatResponse()
	if err != nil {
		t.Fatalf("format response: %v", err)
	}
	if formatted == "" {
		t.Error("formatted response should not be empty")
	}
}

func TestQueryResultFormatRawHex(t *testing.T) {
	result := &QueryResult{
		ExitCode:	0,
		GasUsed:	500,
		MethodName:	"unknown",
		ABIKnown:	false,
		ResponseBytes:	[]byte{0xDE, 0xAD, 0xBE, 0xEF},
	}

	formatted, err := result.FormatResponse()
	if err != nil {
		t.Fatalf("format response: %v", err)
	}
	if formatted != "deadbeef" {
		t.Errorf("expected 'deadbeef', got '%s'", formatted)
	}
}

func TestBuildGetMethodProof(t *testing.T) {
	sel := ComputeMethodSelector("get_balance(address)")
	snapshot := QuerySnapshot{
		StateRootChunk:	testStateRoot(),
		Code:		[]byte("code"),
	}

	proof := BuildGetMethodProof(snapshot, "get_balance", sel, []byte{1, 2, 3})
	if !proof.Enabled {
		t.Error("proof should be enabled")
	}
	if proof.MethodSelector != sel {
		t.Error("selector mismatch")
	}
	if len(proof.StateRootHash) == 0 {
		t.Error("state root hash should not be empty")
	}
	if len(proof.ResponseHash) == 0 {
		t.Error("response hash should not be empty")
	}
}

func TestABIMethodNotFoundError(t *testing.T) {
	err := NewABIMethodNotFoundError("nonexistent")
	if err == nil {
		t.Error("expected error for unknown method")
	}
	if err.Method != "nonexistent" {
		t.Errorf("expected method 'nonexistent', got '%s'", err.Method)
	}
}

func TestABIDecodingValidation(t *testing.T) {
	err := ValidateABIDecoding(nil, "address")
	if err == nil {
		t.Error("nil args should be rejected")
	}

	err = ValidateABIDecoding([]byte{1, 2, 3}, "")
	if err != nil {
		t.Errorf("empty codec should be OK: %v", err)
	}
}

func TestMethodMutabilityString(t *testing.T) {
	if MethodRead.String() != "READ" {
		t.Errorf("expected READ, got %s", MethodRead.String())
	}
	if MethodWrite.String() != "WRITE" {
		t.Errorf("expected WRITE, got %s", MethodWrite.String())
	}
}

func TestResponseFormatStrings(t *testing.T) {
	if ResponseFormatTypedJSON.String() != "typed_json" {
		t.Errorf("expected typed_json, got %s", ResponseFormatTypedJSON.String())
	}
	if ResponseFormatRawHex.String() != "raw_hex" {
		t.Errorf("expected raw_hex, got %s", ResponseFormatRawHex.String())
	}
}
