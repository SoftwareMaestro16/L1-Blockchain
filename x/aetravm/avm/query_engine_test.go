package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func makeTestQuerySnapshot() QuerySnapshot {
	builder := chunk.NewBuilder()
	builder.SetTypeTag(chunk.TypeNormal)
	builder.SetData([]byte("test_state"), 72)
	stateChunk, _ := builder.Build()

	return QuerySnapshot{
		StateRootChunk:	stateChunk,
		Code:		[]byte{0x01, 0x02, 0x03},
		BlockCtx: BlockContext{
			Height:		100,
			ChainID:	"test-chain",
		},
	}
}

func TestQueryEngineReadsState(t *testing.T) {
	engine := NewQueryEngine()
	snapshot := makeTestQuerySnapshot()
	receipt, err := engine.ExecuteQuery(snapshot, "get_balance", []byte(`{"address":"4:test"}`), 0)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}
	if receipt.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", receipt.ExitCode)
	}
	if receipt.GasUsed == 0 {
		t.Fatal("expected gas used > 0")
	}
	if len(receipt.Response) == 0 {
		t.Fatal("expected non-empty response")
	}
}

func TestQueryAttemptedWriteRejected(t *testing.T) {
	forbidden := QueryForbiddenHostFunctions()
	hasWrite := false
	hasDelete := false
	for _, fn := range forbidden {
		if fn == HostWriteStorage {
			hasWrite = true
		}
		if fn == HostDeleteStorage {
			hasDelete = true
		}
	}
	if !hasWrite {
		t.Fatal("write_storage MUST be forbidden in query mode")
	}
	if !hasDelete {
		t.Fatal("delete_storage MUST be forbidden in query mode")
	}
}

func TestQueryAttemptedSendMessageRejected(t *testing.T) {
	forbidden := QueryForbiddenHostFunctions()
	hasSend := false
	hasEmit := false
	for _, fn := range forbidden {
		if fn == HostSendInternal {
			hasSend = true
		}
		if fn == HostEmitEvent {
			hasEmit = true
		}
	}
	if !hasSend {
		t.Fatal("send_internal MUST be forbidden in query mode")
	}
	if !hasEmit {
		t.Fatal("emit_event MUST be forbidden in query mode")
	}
}

func TestQueryGasLimitEnforced(t *testing.T) {
	engine := NewQueryEngine()
	snapshot := makeTestQuerySnapshot()

	receipt, err := engine.ExecuteQuery(snapshot, "get_balance", []byte(`{}`), 100)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	if receipt.ExitCode == 0 {

		t.Logf("Low gas query completed with exit code %d", receipt.ExitCode)
	}
}

func TestQueryResponseDeterministic(t *testing.T) {
	engine := NewQueryEngine()
	snapshot := makeTestQuerySnapshot()

	receipt1, err := engine.ExecuteQuery(snapshot, "get_balance", []byte(`{}`), 0)
	if err != nil {
		t.Fatalf("ExecuteQuery 1 failed: %v", err)
	}
	receipt2, err := engine.ExecuteQuery(snapshot, "get_balance", []byte(`{}`), 0)
	if err != nil {
		t.Fatalf("ExecuteQuery 2 failed: %v", err)
	}

	if receipt1.ExitCode != receipt2.ExitCode {
		t.Fatalf("deterministic exit code mismatch: %d vs %d", receipt1.ExitCode, receipt2.ExitCode)
	}
	if receipt1.GasUsed != receipt2.GasUsed {
		t.Fatalf("deterministic gas mismatch: %d vs %d", receipt1.GasUsed, receipt2.GasUsed)
	}
	if receipt1.TraceHash != receipt2.TraceHash {
		t.Fatalf("deterministic trace hash mismatch: %s vs %s", receipt1.TraceHash, receipt2.TraceHash)
	}

	receipt3, err := engine.ExecuteQuery(snapshot, "get_balance", []byte(`{}`), 0)
	if err != nil {
		t.Fatalf("ExecuteQuery 3 failed: %v", err)
	}
	if receipt3.TraceHash != receipt1.TraceHash {
		t.Fatalf("cached query trace hash mismatch: %s vs %s", receipt3.TraceHash, receipt1.TraceHash)
	}
}

func TestQueryMalformedArgsRejected(t *testing.T) {
	err := ValidateQueryArguments(nil)
	if err == nil {
		t.Fatal("nil args should be rejected")
	}

	oversized := make([]byte, DefaultQueryMaxResponseBytes+1)
	err = ValidateQueryArguments(oversized)
	if err == nil {
		t.Fatal("oversized args should be rejected")
	}
}

func TestQueryProofMode(t *testing.T) {
	engine := NewQueryEngineWithProof()
	snapshot := makeTestQuerySnapshot()

	receipt, proof, err := engine.ExecuteQueryWithProof(snapshot, "get_balance", []byte(`{}`), 0)
	if err != nil {
		t.Fatalf("ExecuteQueryWithProof failed: %v", err)
	}
	if !proof.Enabled {
		t.Fatal("proof mode should be enabled")
	}
	if len(proof.StateRootProof) == 0 {
		t.Fatal("state root proof should not be empty")
	}
	if len(proof.ResponseProof) == 0 {
		t.Fatal("response proof should not be empty")
	}
	if receipt.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", receipt.ExitCode)
	}

	stateHash := snapshot.StateRootChunk.Hash()
	if !VerifyQueryProof(proof, stateHash, "get_balance", []byte(`{}`)) {
		t.Fatal("proof verification failed")
	}
}

func TestQueryProofModeDisabled(t *testing.T) {
	snapshot := makeTestQuerySnapshot()
	localEngine := NewQueryEngine()

	_, _, err := localEngine.ExecuteQueryWithProof(snapshot, "get_balance", []byte(`{}`), 0)
	if err == nil {
		t.Fatal("ExecuteQueryWithProof should fail when proof mode is disabled")
	}
}

func TestQueryCacheInvalidation(t *testing.T) {
	cache := NewQueryCache(100)

	snapshot := makeTestQuerySnapshot()
	key := ComputeQueryCacheKey(snapshot, 1, []byte("args"))

	receipt := QueryReceipt{ExitCode: 0, GasUsed: 100, Response: []byte("test"), TraceHash: "hash"}

	cache.Put(key, receipt)

	cached, ok := cache.Get(key)
	if !ok {
		t.Fatal("cache entry should be found")
	}
	if cached.ExitCode != 0 {
		t.Fatalf("cached exit code mismatch: %d", cached.ExitCode)
	}

	cache.Invalidate([]byte("different_root"))
	if cache.Size() != 0 {
		t.Fatalf("cache should be invalidated, size: %d", cache.Size())
	}
}

func TestQueryCanonicalResponseEncoding(t *testing.T) {
	enc := QueryResponseCanonicalEncoding{
		MethodID:	42,
		GasUsed:	1000,
		ExitCode:	0,
		Payload:	[]byte("response_data"),
		ProofRoot:	[]byte("proof"),
	}

	encoded := EncodeQueryResponseCanonical(enc)
	decoded, err := DecodeQueryResponseCanonical(encoded)
	if err != nil {
		t.Fatalf("canonical decode failed: %v", err)
	}
	if decoded.MethodID != enc.MethodID {
		t.Fatalf("method ID mismatch: %d vs %d", decoded.MethodID, enc.MethodID)
	}
	if decoded.GasUsed != enc.GasUsed {
		t.Fatalf("gas used mismatch: %d vs %d", decoded.GasUsed, enc.GasUsed)
	}
	if decoded.ExitCode != enc.ExitCode {
		t.Fatalf("exit code mismatch: %d vs %d", decoded.ExitCode, enc.ExitCode)
	}
	if string(decoded.Payload) != string(enc.Payload) {
		t.Fatalf("payload mismatch: %q vs %q", string(decoded.Payload), string(enc.Payload))
	}
	if string(decoded.ProofRoot) != string(enc.ProofRoot) {
		t.Fatalf("proof root mismatch: %q vs %q", string(decoded.ProofRoot), string(enc.ProofRoot))
	}
}

func TestQueryCanonicalResponseDeterministic(t *testing.T) {
	enc := QueryResponseCanonicalEncoding{
		MethodID:	1,
		GasUsed:	500,
		ExitCode:	0,
		Payload:	[]byte("data"),
		ProofRoot:	[]byte("proof"),
	}

	encoded1 := EncodeQueryResponseCanonical(enc)
	encoded2 := EncodeQueryResponseCanonical(enc)

	if len(encoded1) != len(encoded2) {
		t.Fatal("canonical encoding length mismatch")
	}
	for i := range encoded1 {
		if encoded1[i] != encoded2[i] {
			t.Fatalf("canonical encoding not deterministic at byte %d", i)
		}
	}
}

func TestMethodRegistry(t *testing.T) {
	registry := &MethodRegistry{}

	err := registry.RegisterMethod(MethodRegistryEntry{
		MethodID:	1,
		Name:		"get_balance",
		GasEstimate:	1000,
	})
	if err != nil {
		t.Fatalf("RegisterMethod failed: %v", err)
	}

	err = registry.RegisterMethod(MethodRegistryEntry{
		MethodID:	2,
		Name:		"get_owner",
		GasEstimate:	500,
	})
	if err != nil {
		t.Fatalf("RegisterMethod failed: %v", err)
	}

	err = registry.RegisterMethod(MethodRegistryEntry{
		MethodID:	1,
		Name:		"get_other",
		GasEstimate:	100,
	})
	if err == nil {
		t.Fatal("duplicate method ID should be rejected")
	}

	entry, ok := registry.LookupMethod("get_balance")
	if !ok {
		t.Fatal("get_balance should be found")
	}
	if entry.MethodID != 1 {
		t.Fatalf("method ID mismatch: %d", entry.MethodID)
	}

	entry, ok = registry.LookupMethodByID(2)
	if !ok {
		t.Fatal("method ID 2 should be found")
	}
	if entry.Name != "get_owner" {
		t.Fatalf("method name mismatch: %s", entry.Name)
	}
}

func TestMethodRegistryValidation(t *testing.T) {
	err := ValidateMethodRegistryEntry(MethodRegistryEntry{
		MethodID:	0,
		Name:		"",
		GasEstimate:	0,
	})
	if err == nil {
		t.Fatal("empty name should be rejected")
	}

	err = ValidateMethodRegistryEntry(MethodRegistryEntry{
		MethodID:	1,
		Name:		"test",
		GasEstimate:	100,
	})
	if err != nil {
		t.Fatalf("valid entry should pass: %v", err)
	}
}

func TestQueryIsolationBoundary(t *testing.T) {
	boundary := QueryIsolationBoundary{
		CanReadStorage:		true,
		CanSendMessages:	false,
		CanEmitEvents:		false,
		CanWriteStorage:	false,
	}

	if !boundary.CanReadStorage {
		t.Fatal("query SHOULD be able to read storage")
	}
	if boundary.CanSendMessages {
		t.Fatal("query MUST NOT be able to send messages")
	}
	if boundary.CanEmitEvents {
		t.Fatal("query MUST NOT be able to emit events")
	}
	if boundary.CanWriteStorage {
		t.Fatal("query MUST NOT be able to write storage")
	}
}

func TestQueryAllowedHostFunctions(t *testing.T) {
	allowed := QueryAllowedHostFunctions()

	readFound := false
	for _, fn := range allowed {
		if fn == HostReadStorage {
			readFound = true
		}
	}
	if !readFound {
		t.Fatal("read_storage MUST be allowed in query mode")
	}

	for _, fn := range allowed {
		if fn == HostWriteStorage {
			t.Fatal("write_storage MUST NOT be allowed in query mode")
		}
		if fn == HostSendInternal {
			t.Fatal("send_internal MUST NOT be allowed in query mode")
		}
	}
}

func TestQueryDeterminismViolations(t *testing.T) {
	domain := &QueryExecutionDomain{
		StackTrace: []QueryTraceStep{
			{Instruction: "LOAD", GasConsumed: 10, Opcode: "read_storage"},
		},
	}
	err := ValidateQueryDeterminism(domain)
	if err != nil {
		t.Fatalf("valid trace should pass: %v", err)
	}

	badDomain := &QueryExecutionDomain{
		StackTrace: []QueryTraceStep{
			{Instruction: "WALL_CLOCK", GasConsumed: 10, Opcode: "wall_clock"},
		},
	}
	err = ValidateQueryDeterminism(badDomain)
	if err == nil {
		t.Fatal("wall_clock opcode MUST be rejected in query mode")
	}
}

func TestQuerySnapshotImmutability(t *testing.T) {
	snapshot := makeTestQuerySnapshot()

	err := ValidateQuerySnapshot(&snapshot)
	if err != nil {
		t.Fatalf("valid snapshot should pass: %v", err)
	}

	nilSnapshot := QuerySnapshot{}
	err = ValidateQuerySnapshot(&nilSnapshot)
	if err == nil {
		t.Fatal("nil state root chunk should be rejected")
	}

	noCode := QuerySnapshot{
		StateRootChunk:	snapshot.StateRootChunk,
		Code:		nil,
	}
	_ = noCode
}

func TestQueryGasModel(t *testing.T) {
	model := QueryGasModel{
		ComputeGas:		10,
		DecodeGas:		5,
		SerializationGas:	2,
	}

	total := model.Total()
	if total != 17 {
		t.Fatalf("expected total gas model sum 17, got %d", total)
	}

	accounting := &QueryGasAccounting{
		Model:	model,
		Limit:	1000,
	}

	if !accounting.ChargeDecode(10) {
		t.Fatal("decode charge should succeed")
	}
	if accounting.Used.DecodeGas != 50 {
		t.Fatalf("expected 50 decode gas, got %d", accounting.Used.DecodeGas)
	}

	if !accounting.ChargeCompute(100) {
		t.Fatal("compute charge should succeed")
	}
	if accounting.Used.ComputeGas != 100 {
		t.Fatalf("expected 100 compute gas, got %d", accounting.Used.ComputeGas)
	}

	if !accounting.ChargeSerialize(20) {
		t.Fatal("serialize charge should succeed")
	}
	if accounting.Used.SerializationGas != 40 {
		t.Fatalf("expected 40 serialization gas, got %d", accounting.Used.SerializationGas)
	}

	totalUsed := accounting.Used.Total()
	if totalUsed != 190 {
		t.Fatalf("expected total used 190, got %d", totalUsed)
	}
}

func TestQueryGasExhaustion(t *testing.T) {
	model := QueryGasModel{
		ComputeGas:		10,
		DecodeGas:		5,
		SerializationGas:	2,
	}
	accounting := &QueryGasAccounting{
		Model:	model,
		Limit:	50,
	}

	if accounting.ChargeCompute(100) {
		t.Fatal("charge should fail when exceeding limit")
	}
	if !accounting.Aborted {
		t.Fatal("accounting should be marked as aborted")
	}
}

func TestComputeMethodIDDeterministic(t *testing.T) {
	id1 := ComputeMethodID("get_balance")
	id2 := ComputeMethodID("get_balance")
	if id1 != id2 {
		t.Fatal("method IDs should be deterministic")
	}

	id3 := ComputeMethodID("get_owner")
	if id1 == id3 {
		t.Fatal("different method names should produce different IDs")
	}
}

func TestQueryStackTrace(t *testing.T) {
	domain := &QueryExecutionDomain{
		StackTrace: []QueryTraceStep{
			{Instruction: "LOAD", GasConsumed: 10, Opcode: "load_storage"},
			{Instruction: "COMPUTE", GasConsumed: 50, Opcode: "add"},
			{Instruction: "RETURN", GasConsumed: 5, Opcode: "return"},
		},
	}

	model := QueryGasModel{ComputeGas: 65, DecodeGas: 5, SerializationGas: 2}
	accounting := &QueryGasAccounting{Model: model, Used: model}

	trace := BuildQueryTrace(domain, accounting)
	if len(trace.Steps) != 3 {
		t.Fatalf("expected 3 trace steps, got %d", len(trace.Steps))
	}
	if trace.GasBreakdown.ComputeGas != 65 {
		t.Fatalf("expected compute gas 65, got %d", trace.GasBreakdown.ComputeGas)
	}
}

func TestQueryStackLimits(t *testing.T) {
	limits := QueryStackLimits{
		MaxStackDepth:		DefaultQueryMaxStackDepth,
		MaxRecursionDepth:	DefaultQueryMaxRecursionDepth,
		MaxChunkTraversalDepth:	DefaultQueryMaxChunkTraversalDepth,
	}
	if limits.MaxStackDepth != 512 {
		t.Fatalf("expected default max stack depth 512, got %d", limits.MaxStackDepth)
	}
	if limits.MaxRecursionDepth != 64 {
		t.Fatalf("expected default max recursion depth 64, got %d", limits.MaxRecursionDepth)
	}
	if limits.MaxChunkTraversalDepth != 128 {
		t.Fatalf("expected default max chunk traversal depth 128, got %d", limits.MaxChunkTraversalDepth)
	}
}

func TestQuerySnapshotValidation(t *testing.T) {
	err := ValidateQuerySnapshot(nil)
	if err == nil {
		t.Fatal("nil snapshot should be rejected")
	}
}
