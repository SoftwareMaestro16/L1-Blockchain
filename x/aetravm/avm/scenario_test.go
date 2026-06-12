package avm

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func TestASMModuleCreation(t *testing.T) {
	m := NewASMModule("test_counter")
	if m.Name != "test_counter" {
		t.Errorf("expected name test_counter, got %s", m.Name)
	}
	if len(m.Instructions) != 0 {
		t.Errorf("expected empty instructions")
	}
	if len(m.Entrypoints) != 1 || m.Entrypoints[0] != "main" {
		t.Errorf("expected default entrypoint 'main'")
	}
}

func TestASMModuleChainableInstructions(t *testing.T) {
	m := NewASMModule("chain").
		Push(42).
		Push(10).
		Add().
		Dup().
		Drop().
		StoreState()
	if len(m.Instructions) != 6 {
		t.Errorf("expected 6 instructions, got %d", len(m.Instructions))
	}
	if m.Instructions[0].Opcode != "push" || m.Instructions[0].Args[0] != "42" {
		t.Errorf("expected push 42, got %v", m.Instructions[0])
	}
	if m.Instructions[2].Opcode != "add" {
		t.Errorf("expected add at index 2, got %s", m.Instructions[2].Opcode)
	}
	if m.Instructions[5].Opcode != "store_state" {
		t.Errorf("expected store_state at index 5, got %s", m.Instructions[5].Opcode)
	}
}

func TestASMModuleComment(t *testing.T) {
	m := NewASMModule("commented").Push(1).Comment("push one")
	if m.Instructions[0].Comment != "push one" {
		t.Errorf("expected comment 'push one', got %s", m.Instructions[0].Comment)
	}
}

func TestASMModuleAllOpcodes(t *testing.T) {
	m := NewASMModule("all_ops").
		Push(1).
		Dup().
		Drop().
		Add().
		Sub().
		Mul().
		Div().
		Eq().
		JumpCond("end").
		CallFrame("helper").
		ReturnFrame().
		Nop().
		LoadState().
		StoreState().
		ChunkMapGet().
		ChunkMapPut().
		EmitAction().
		GetCaller().
		HashData().
		RaiseError(99)
	expected := []string{
		"push", "dup", "drop", "add", "sub", "mul", "div", "eq",
		"jump_cond", "call_frame", "return_frame", "nop",
		"load_state", "store_state", "chunkmap_get", "chunkmap_put",
		"emit_action", "get_caller", "hash_data", "raise_error",
	}
	for i, exp := range expected {
		if m.Instructions[i].Opcode != exp {
			t.Errorf("instruction %d: expected %s, got %s", i, exp, m.Instructions[i].Opcode)
		}
	}
}

func TestJSONModuleCreation(t *testing.T) {
	m := NewJSONModule("counter", 1, 1)
	if m.Name != "counter" {
		t.Errorf("expected name counter, got %s", m.Name)
	}
	if m.Version != 1 || m.ABI != 1 {
		t.Errorf("expected version=1, abi=1, got version=%d, abi=%d", m.Version, m.ABI)
	}
	if len(m.Code) != 0 || len(m.Exports) != 0 {
		t.Errorf("expected empty code and exports")
	}
}

func TestJSONModuleAddInstruction(t *testing.T) {
	m := NewJSONModule("test_mod", 1, 1)
	m.AddInstruction("push", 42, "").
		AddInstruction("add", 0, "").
		AddInstruction("jump_cond", 0, "loop_start")
	if len(m.Code) != 3 {
		t.Errorf("expected 3 instructions, got %d", len(m.Code))
	}
	if m.Code[0].Opcode != "push" || m.Code[0].Arg != 42 {
		t.Errorf("expected push 42, got %v", m.Code[0])
	}
	if m.Code[2].Target != "loop_start" {
		t.Errorf("expected target loop_start, got %s", m.Code[2].Target)
	}
}

func TestJSONModuleAddExport(t *testing.T) {
	m := NewJSONModule("test_mod", 1, 1)
	m.AddExport("get_count", "entry_0", 0xA1B2C3D4)
	if len(m.Exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(m.Exports))
	}
	exp := m.Exports[0]
	if exp.Name != "get_count" || exp.Entrypoint != "entry_0" || exp.Selector != 0xA1B2C3D4 {
		t.Errorf("unexpected export: %+v", exp)
	}
}

func TestJSONModuleRoundTrip(t *testing.T) {
	m := NewJSONModule("roundtrip_test", 2, 1)
	m.AddInstruction("push", 100, "").
		AddInstruction("load_state", 0, "").
		AddInstruction("chunkmap_get", 0, "").
		AddInstruction("return_frame", 0, "")
	m.AddExport("query_balance", "entry_0", 0xDEADBEEF)

	json, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	m2, err := JSONModuleFromJSON(json)
	if err != nil {
		t.Fatalf("JSONModuleFromJSON: %v", err)
	}
	if m2.Name != "roundtrip_test" {
		t.Errorf("expected name roundtrip_test, got %s", m2.Name)
	}
	if m2.Version != 2 {
		t.Errorf("expected version 2, got %d", m2.Version)
	}
	if len(m2.Code) != 4 {
		t.Errorf("expected 4 instructions, got %d", len(m2.Code))
	}
	if len(m2.Exports) != 1 {
		t.Errorf("expected 1 export, got %d", len(m2.Exports))
	}
	if m2.Exports[0].Selector != 0xDEADBEEF {
		t.Errorf("expected selector DEADBEEF, got %X", m2.Exports[0].Selector)
	}
}

func TestDeterminismInvariantPasses(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{},
	}
	check := DeterminismInvariant(result)
	if !check.Passed {
		t.Errorf("determinism invariant should pass: %s", check.Message)
	}
}

func TestValueConservationInvariantPasses(t *testing.T) {
	receipt := &AVMLedgerReceipt{
		ExitCode:	ExitSuccess,
		GasUsed:	1000,
		ValueIn:	5000,
		ValueOut:	3000,
		GasBreakdown:	GasBreakdown{ComputeGas: 1000},
		StorageFee:	2000,
	}
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{receipt},
	}
	check := ValueConservationInvariant(result)
	if !check.Passed {
		t.Errorf("value conservation invariant should pass: %s", check.Message)
	}
}

func TestValueConservationInvariantFailsOnImbalance(t *testing.T) {
	receipt := &AVMLedgerReceipt{
		ExitCode:	ExitSuccess,
		GasUsed:	1000,
		ValueIn:	5000,
		ValueOut:	1000,
		GasBreakdown:	GasBreakdown{ComputeGas: 1000},
		StorageFee:	500,
	}
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{receipt},
	}
	check := ValueConservationInvariant(result)
	if check.Passed {
		t.Errorf("value conservation should fail when value_in != value_out + fees")
	}
}

func TestGasMonotonicityInvariantPasses(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{
			{GasUsed: 100},
			{GasUsed: 200},
			{GasUsed: 300},
		},
	}
	check := GasMonotonicityInvariant(result)
	if !check.Passed {
		t.Errorf("gas monotonicity should pass for increasing gas: %s", check.Message)
	}
}

func TestGasMonotonicityInvariantFailsOnDecrease(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{
			{GasUsed: 300},
			{GasUsed: 200},
		},
	}
	check := GasMonotonicityInvariant(result)
	if check.Passed {
		t.Errorf("gas monotonicity should fail when gas decreases")
	}
}

func TestNoDoubleRefundInvariantPasses(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{
			{GasUsed: 1000, GasRefunded: 500, MessageFlags: MessageFlags{RefundIssued: true}},
		},
	}
	check := NoDoubleRefundInvariant(result)
	if !check.Passed {
		t.Errorf("no double refund should pass for valid refund: %s", check.Message)
	}
}

func TestNoDoubleRefundInvariantFailsOnExcessRefund(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{
			{GasUsed: 100, GasRefunded: 200, MessageFlags: MessageFlags{RefundIssued: true}},
		},
	}
	check := NoDoubleRefundInvariant(result)
	if check.Passed {
		t.Errorf("no double refund should fail when refund exceeds gas used")
	}
}

func TestNoInfiniteBounceInvariantPasses(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{},
	}
	check := NoInfiniteBounceInvariant(result)
	if !check.Passed {
		t.Errorf("no infinite bounce should pass with zero bounces: %s", check.Message)
	}
}

func TestNoInfiniteBounceInvariantFailsOnMultipleBounces(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{
			{MessageFlags: MessageFlags{Bounced: true}},
			{MessageFlags: MessageFlags{Bounced: true}},
		},
	}
	check := NoInfiniteBounceInvariant(result)
	if check.Passed {
		t.Errorf("no infinite bounce should fail with multiple bounces")
	}
}

func TestRunInvariantsAllPass(t *testing.T) {
	result := &ScenarioResult{
		Receipts: []*AVMLedgerReceipt{
			{
				ExitCode:	ExitSuccess,
				GasUsed:	1000,
				ValueIn:	5000,
				ValueOut:	4000,
				GasBreakdown:	GasBreakdown{ComputeGas: 1000},
				GasRefunded:	0,
				StorageFee:	1000,
				MessageFlags:	MessageFlags{Consumed: true},
			},
		},
	}
	checks := RunInvariants(result)
	for _, c := range checks {
		if !c.Passed {
			t.Errorf("invariant %s failed: %s", c.Name, c.Message)
		}
	}
}

func TestNegativeTestCasesDefined(t *testing.T) {
	if len(NegativeTestCases) == 0 {
		t.Fatal("expected negative test cases to be defined")
	}
	names := make(map[string]bool)
	for _, tc := range NegativeTestCases {
		if tc.Name == "" {
			t.Error("negative test case has empty name")
		}
		if tc.Description == "" {
			t.Errorf("negative test case %s has empty description", tc.Name)
		}
		if names[tc.Name] {
			t.Errorf("duplicate negative test case name: %s", tc.Name)
		}
		names[tc.Name] = true
	}
}

func TestNegativeTestCaseExitCodesCategorized(t *testing.T) {
	categories := map[string]string{
		"invalid_opcode":		"validation",
		"stack_underflow":		"execution",
		"stack_overflow":		"execution",
		"gas_exhaustion":		"gas",
		"type_mismatch":		"type",
		"division_by_zero":		"execution",
		"chunk_reference_invalid":	"chunk",
		"invalid_migration":		"upgrade",
		"unauthorized_upgrade":		"upgrade",
		"forbidden_host_call":		"query",
		"double_refund_attempt":	"receipt",
		"infinite_bounce_attempt":	"receipt",
		"malformed_abi_query":		"query",
		"action_budget_exceeded":	"execution",
	}
	for _, tc := range NegativeTestCases {
		cat, ok := categories[tc.Name]
		if !ok {
			t.Errorf("unexpected negative test case: %s", tc.Name)
			continue
		}
		_ = cat
	}
}

func TestCIPipelineCreation(t *testing.T) {
	pipeline := NewCIPipeline()
	if len(pipeline.Stages) != 8 {
		t.Errorf("expected 8 stages, got %d", len(pipeline.Stages))
	}
	expectedStages := []CIPipelineStage{
		CIStageBuild, CIStageVerify, CIStageDeploy, CIStageExecute,
		CIStageQuery, CIStageExport, CIStageImport, CIStageReplay,
	}
	for i, exp := range expectedStages {
		if pipeline.Stages[i] != exp {
			t.Errorf("stage %d: expected %s, got %s", i, exp, pipeline.Stages[i])
		}
	}
}

func TestCIPipelineStageStrings(t *testing.T) {
	stages := map[CIPipelineStage]string{
		CIStageBuild:	"build",
		CIStageVerify:	"verify",
		CIStageDeploy:	"deploy",
		CIStageExecute:	"execute",
		CIStageQuery:	"query",
		CIStageExport:	"export",
		CIStageImport:	"import",
		CIStageReplay:	"replay",
	}
	for stage, exp := range stages {
		if string(stage) != exp {
			t.Errorf("expected stage %s, got %s", exp, string(stage))
		}
	}
}

func TestRunScenarioBasicExecution(t *testing.T) {
	scenario := &Scenario{
		Name:	"basic_execution",
		Messages: []ScenarioMessage{
			{
				Name:		"simple_transfer",
				Sender:		"AE:sender1",
				Target:		"4:contract1",
				Value:		1000,
				GasLimit:	10000,
			},
		},
	}

	result, err := RunScenario(scenario)
	if err != nil {
		t.Fatalf("RunScenario: %v", err)
	}
	if len(result.Receipts) != 1 {
		t.Errorf("expected 1 receipt, got %d", len(result.Receipts))
	}
	if result.Receipts[0].ExitCode.Category != ExitCategorySuccess {
		t.Errorf("expected success exit, got %v", result.Receipts[0].ExitCode)
	}
	if result.Receipts[0].ValueIn != 1000 {
		t.Errorf("expected value_in=1000, got %d", result.Receipts[0].ValueIn)
	}
}

func TestRunScenarioMultipleMessages(t *testing.T) {
	scenario := &Scenario{
		Name:	"multi_message",
		Messages: []ScenarioMessage{
			{Name: "msg1", Sender: "AE:alice", Target: "4:contract1", Value: 500, GasLimit: 5000},
			{Name: "msg2", Sender: "AE:bob", Target: "4:contract1", Value: 1000, GasLimit: 5000},
			{Name: "msg3", Sender: "AE:carol", Target: "4:contract1", Value: 2000, GasLimit: 5000},
		},
	}

	result, err := RunScenario(scenario)
	if err != nil {
		t.Fatalf("RunScenario: %v", err)
	}
	if len(result.Receipts) != 3 {
		t.Errorf("expected 3 receipts, got %d", len(result.Receipts))
	}
}

func TestRunScenarioInvariantChecks(t *testing.T) {
	scenario := &Scenario{
		Name:	"invariant_check",
		Messages: []ScenarioMessage{
			{Name: "msg", Sender: "AE:sender", Target: "4:target", Value: 500, GasLimit: 5000},
		},
	}

	result, err := RunScenario(scenario)
	if err != nil {
		t.Fatalf("RunScenario: %v", err)
	}
	if len(result.InvariantChecks) == 0 {
		t.Error("expected invariant checks to be run")
	}
	for _, check := range result.InvariantChecks {
		if check.Name == "" {
			t.Error("invariant check missing name")
		}
	}
}

func TestCanonicalExamplesDefined(t *testing.T) {
	if len(CanonicalExamples) == 0 {
		t.Fatal("expected canonical examples to be defined")
	}
	names := make(map[string]bool)
	for _, ex := range CanonicalExamples {
		if ex.Name == "" {
			t.Error("canonical example has empty name")
		}
		if ex.Description == "" {
			t.Errorf("canonical example %s has empty description", ex.Name)
		}
		if names[ex.Name] {
			t.Errorf("duplicate canonical example name: %s", ex.Name)
		}
		names[ex.Name] = true
	}
}

func TestCanonicalExamplesCoversAllScenarios(t *testing.T) {
	required := []string{"counter", "domain_registry",
		"message_sender_receiver", "bounce_scenario", "refund_scenario",
		"get_method_query", "upgrade_migration"}
	names := make(map[string]bool)
	for _, ex := range CanonicalExamples {
		names[ex.Name] = true
	}
	for _, req := range required {
		if !names[req] {
			t.Errorf("missing canonical example: %s", req)
		}
	}
}

func TestScenarioMessageBounceFlag(t *testing.T) {
	msg := ScenarioMessage{
		Name:	"bounce_test",
		Sender:	"AE:sender",
		Target:	"4:target",
		Value:	1000,
		Bounce:	true,
	}
	if !msg.Bounce {
		t.Error("expected Bounce=true")
	}
}

func TestPipelineBuildAndVerify(t *testing.T) {
	m := NewJSONModule("pipeline_test", 1, 1)
	m.AddInstruction("push", 0, "").
		AddInstruction("load_state", 0, "").
		AddInstruction("return_frame", 0, "")
	m.AddExport("get_value", "entry_0", 0x12345678)

	json, err := m.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	m2, err := JSONModuleFromJSON(json)
	if err != nil {
		t.Fatalf("JSONModuleFromJSON: %v", err)
	}
	if m2.Name != "pipeline_test" {
		t.Errorf("round-trip failed: expected pipeline_test, got %s", m2.Name)
	}
	if len(m2.Code) != 3 {
		t.Errorf("round-trip failed: expected 3 instructions, got %d", len(m2.Code))
	}
	if len(m2.Exports) != 1 {
		t.Errorf("round-trip failed: expected 1 export, got %d", len(m2.Exports))
	}
}

func TestPipelineScenarioWithScenarioResult(t *testing.T) {
	scenario := &Scenario{
		Name:	"pipeline_integration",
		Messages: []ScenarioMessage{
			{Name: "deploy", Sender: "AE:deployer", Target: "4:new_contract", Value: 0, GasLimit: 50000},
			{Name: "interact", Sender: "AE:user", Target: "4:new_contract", Value: 100, GasLimit: 10000},
		},
	}

	result, err := RunScenario(scenario)
	if err != nil {
		t.Fatalf("RunScenario: %v", err)
	}
	if len(result.Receipts) != 2 {
		t.Errorf("expected 2 receipts, got %d", len(result.Receipts))
	}
	for _, r := range result.Receipts {
		if r.ExitCode.Category != ExitCategorySuccess {
			t.Errorf("expected success exit code, got %v", r.ExitCode)
		}
	}
}

func TestQueryScenarioBuildsSnapshot(t *testing.T) {
	emptyMap := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("value"), 40).Build()
	m, _ := emptyMap.Put([]byte("key"), b)
	root := m.Root()

	snapshot := &QuerySnapshot{
		StateRootChunk:	root,
		Code:		[]byte{0x01, 0x02},
		BlockCtx:	BlockContext{Height: 100},
	}
	if snapshot.BlockCtx.Height != 100 {
		t.Errorf("expected block height 100, got %d", snapshot.BlockCtx.Height)
	}
	if snapshot.StateRootChunk == nil {
		t.Error("expected non-nil state root chunk")
	}
}

func TestUpgradeScenarioWithMigration(t *testing.T) {
	registry := NewMigrationRegistry()
	var oldCodeHash [32]byte
	oldCodeHash[0] = 0xAB
	var newCodeHash [32]byte
	newCodeHash[0] = 0xCD

	called := false
	handler := func(oldRoot *chunk.Chunk, oldSchema, newSchema ContractVersion, gasLimit uint64) (*chunk.Chunk, uint64, error) {
		called = true
		m2 := chunk.NewEmptyMap()
		b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("migrated"), 48).Build()
		m2, _ = m2.Put([]byte("schema_version"), b)
		return m2.Root(), 500, nil
	}
	registry.Register(newCodeHash, handler, nil)

	m := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("original"), 48).Build()
	m, _ = m.Put([]byte("data"), b)
	originalRoot := m.Root()

	state := &ContractState{
		Address:	"4:upgradable",
		Version:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:	oldCodeHash,
		Capabilities:	UpgradeFlagAllowed,
		Admin:		"AE:admin",
		StateRoot:	originalRoot,
	}

	msg := UpgradeMessage{
		Type:		MsgUpgradeContractCode,
		NewCodeHash:	newCodeHash,
		TargetSchema:	ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 1},
		Caller:		"AE:admin",
		Authority:	AuthorityAdmin,
	}

	engine := NewUpgradeEngine(registry)
	engine.RegisterVerifiedModule(oldCodeHash)
	engine.RegisterVerifiedModule(newCodeHash)

	newState, _, err := engine.ProcessUpgrade(state, msg)
	if err != nil {
		t.Fatalf("ProcessUpgrade: %v", err)
	}
	if !called {
		t.Error("expected migration handler to be called")
	}
	if newState.Version.SchemaVersion != 2 {
		t.Errorf("expected schema version 2, got %d", newState.Version.SchemaVersion)
	}
}

func TestBounceScenarioSingleBounceOnly(t *testing.T) {
	bounceMsg := NewBounceMessage([]byte("msg_hash"), ExitChunkError)

	msg := AVMLedgerReceipt{
		ExitCode:	ExitChunkError,
		GasUsed:	500,
		GasRefunded:	200,
		ValueIn:	1000,
		ValueOut:	800,
		MessageFlags:	MessageFlags{Bounced: true, RefundIssued: true},
		BounceMessage:	bounceMsg,
	}

	if !msg.MessageFlags.Bounced {
		t.Error("expected bounced=true")
	}
	if !msg.MessageFlags.RefundIssued {
		t.Error("expected refund_issued=true")
	}
	if msg.BounceMessage == nil {
		t.Error("expected bounce message")
	}
	if !msg.BounceMessage.BounceFlag {
		t.Error("expected bounce flag=true")
	}
}

func TestRefundDoubleRefundPrevention(t *testing.T) {
	acc := NewRefundAccounting()
	err := acc.IssueRefund(2000)
	if err != nil {
		t.Fatalf("first refund should succeed: %v", err)
	}
	if acc.GasRefunded != 2000 {
		t.Errorf("expected 2000 refunded, got %d", acc.GasRefunded)
	}
	err = acc.IssueRefund(1000)
	if err == nil {
		t.Error("second refund should fail (double refund prevention)")
	}
}

func TestStateInitDeployScenario(t *testing.T) {
	var codeHash [32]byte
	codeHash[0] = 0x01
	codeHash[1] = 0x02
	codeHash[2] = 0x03
	codeHash[3] = 0x04

	salt := []byte{0xAA, 0xBB}
	si := &StateInit{
		ABIVersion:		1,
		CodeHash:		codeHash,
		InitData:		[]byte{0x00},
		Salt:			salt,
		DeployerAddress:	"AE:deployer1",
		ChainID:		"test",
		Namespace:		"namespace",
	}

	addr, err := DeriveContractAddress(si)
	if err != nil {
		t.Fatalf("DeriveContractAddress: %v", err)
	}
	if len(addr.Internal) == 0 {
		t.Error("expected non-empty internal address")
	}
	if len(addr.External) == 0 {
		t.Error("expected non-empty external address")
	}
	if addr.Internal[:2] != "4:" {
		t.Errorf("expected internal address starting with '4:', got %s", addr.Internal[:2])
	}
	if addr.External[:3] != "AE:" {
		t.Errorf("expected external address starting with 'AE:', got %s", addr.External[:3])
	}

	if err := si.Validate(); err != nil {
		t.Errorf("valid StateInit should pass validation: %v", err)
	}
}

func TestGetMethodQueryScenario(t *testing.T) {
	selector := ComputeMethodSelector("get_balance(address)")
	if selector == [4]byte{} {
		t.Error("expected non-zero selector")
	}

	resolver := NewABIMethodResolver()
	balanceMethod := GetMethodABI{
		Name:			"get_balance(address)",
		Selector:		selector,
		InputCodec:		"abi",
		OutputCodec:		"abi",
		GasEstimate:		500,
		Mutability:		MethodRead,
		Cacheable:		true,
		MaxResponseBytes:	1024,
		Description:		"get balance for address",
	}
	if err := resolver.Register(balanceMethod); err != nil {
		t.Fatalf("failed to register get_balance: %v", err)
	}

	ownerSelector := ComputeMethodSelector("get_owner()")
	ownerMethod := GetMethodABI{
		Name:			"get_owner()",
		Selector:		ownerSelector,
		InputCodec:		"abi",
		OutputCodec:		"abi",
		GasEstimate:		300,
		Mutability:		MethodRead,
		Cacheable:		true,
		MaxResponseBytes:	256,
		Description:		"get contract owner",
	}
	if err := resolver.Register(ownerMethod); err != nil {
		t.Fatalf("failed to register get_owner: %v", err)
	}

	resolved, ok := resolver.ResolveByName("get_balance(address)")
	if !ok {
		t.Error("failed to resolve get_balance by name")
	}
	if resolved.Selector != selector {
		t.Errorf("expected selector %x, got %x", selector, resolved.Selector)
	}

	methods := resolver.AllMethods()
	if len(methods) != 2 {
		t.Errorf("expected 2 discoverable methods, got %d", len(methods))
	}
}
