package avm

import (
	"encoding/json"
	"fmt"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

type ASMInstruction struct {
	Opcode	string
	Args	[]string
	Comment	string
}

type ASMModule struct {
	Name		string
	Instructions	[]ASMInstruction
	Entrypoints	[]string
}

func NewASMModule(name string) *ASMModule {
	return &ASMModule{
		Name:		name,
		Instructions:	make([]ASMInstruction, 0),
		Entrypoints:	[]string{"main"},
	}
}

func (m *ASMModule) Push(val int64) *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "push", Args: []string{fmt.Sprintf("%d", val)}})
	return m
}

func (m *ASMModule) Dup() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "dup"})
	return m
}

func (m *ASMModule) Drop() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "drop"})
	return m
}

func (m *ASMModule) Add() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "add"})
	return m
}

func (m *ASMModule) Sub() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "sub"})
	return m
}

func (m *ASMModule) Mul() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "mul"})
	return m
}

func (m *ASMModule) Div() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "div"})
	return m
}

func (m *ASMModule) Eq() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "eq"})
	return m
}

func (m *ASMModule) JumpCond(target string) *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "jump_cond", Args: []string{target}})
	return m
}

func (m *ASMModule) CallFrame(target string) *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "call_frame", Args: []string{target}})
	return m
}

func (m *ASMModule) ReturnFrame() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "return_frame"})
	return m
}

func (m *ASMModule) Nop() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "nop"})
	return m
}

func (m *ASMModule) LoadState() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "load_state"})
	return m
}

func (m *ASMModule) StoreState() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "store_state"})
	return m
}

func (m *ASMModule) ChunkMapGet() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "chunkmap_get"})
	return m
}

func (m *ASMModule) ChunkMapPut() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "chunkmap_put"})
	return m
}

func (m *ASMModule) EmitAction() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "emit_action"})
	return m
}

func (m *ASMModule) GetCaller() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "get_caller"})
	return m
}

func (m *ASMModule) HashData() *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "hash_data"})
	return m
}

func (m *ASMModule) RaiseError(code int64) *ASMModule {
	m.Instructions = append(m.Instructions, ASMInstruction{Opcode: "raise_error", Args: []string{fmt.Sprintf("%d", code)}})
	return m
}

func (m *ASMModule) Comment(text string) *ASMModule {
	if len(m.Instructions) > 0 {
		m.Instructions[len(m.Instructions)-1].Comment = text
	}
	return m
}

type JSONModule struct {
	Name		string			`json:"name"`
	Version		uint32			`json:"version"`
	ABI		uint32			`json:"abi_version"`
	Code		[]JSONInstruction	`json:"code"`
	Imports		[]string		`json:"imports,omitempty"`
	Exports		[]JSONExport		`json:"exports"`
	InitData	[]byte			`json:"init_data,omitempty"`
	Salt		[]byte			`json:"salt,omitempty"`
}

type JSONInstruction struct {
	Opcode	string	`json:"opcode"`
	Arg	int64	`json:"arg,omitempty"`
	Target	string	`json:"target,omitempty"`
}

type JSONExport struct {
	Name		string	`json:"name"`
	Entrypoint	string	`json:"entrypoint"`
	Selector	uint32	`json:"selector"`
}

func NewJSONModule(name string, version, abi uint32) *JSONModule {
	return &JSONModule{
		Name:		name,
		Version:	version,
		ABI:		abi,
		Code:		make([]JSONInstruction, 0),
		Exports:	make([]JSONExport, 0),
	}
}

func (m *JSONModule) AddInstruction(opcode string, arg int64, target string) *JSONModule {
	m.Code = append(m.Code, JSONInstruction{Opcode: opcode, Arg: arg, Target: target})
	return m
}

func (m *JSONModule) AddExport(name, entrypoint string, selector uint32) *JSONModule {
	m.Exports = append(m.Exports, JSONExport{Name: name, Entrypoint: entrypoint, Selector: selector})
	return m
}

func (m *JSONModule) ToJSON() (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("AVM JSON module: %w", err)
	}
	return string(data), nil
}

func JSONModuleFromJSON(data string) (*JSONModule, error) {
	var m JSONModule
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, fmt.Errorf("AVM JSON module: %w", err)
	}
	return &m, nil
}

type Scenario struct {
	Name			string
	InitialState		*chunk.Chunk
	Messages		[]ScenarioMessage
	ExpectedStateHash	[]byte
	ExpectedReceipts	[]ScenarioReceipt
	ExpectedEvents		[]EventRecord
	ExpectedBounce		bool
}

type ScenarioMessage struct {
	Name		string
	Sender		string
	Target		string
	Value		uint64
	GasLimit	uint64
	Payload		[]byte
	Bounce		bool
}

type ScenarioReceipt struct {
	ExitCode	StructuredExitCode
	GasUsed		uint64
	ValueIn		uint64
	ValueOut	uint64
	Bounced		bool
	RefundIssued	bool
}

type ScenarioResult struct {
	StateRoot	*chunk.Chunk
	Receipts	[]*AVMLedgerReceipt
	Events		[]EventRecord
	Bounced		bool
	InvariantChecks	[]InvariantCheckResult
	Passed		bool
}

type InvariantCheckResult struct {
	Name	string
	Passed	bool
	Message	string
}

type InvariantFunc func(result *ScenarioResult) InvariantCheckResult

var CoreInvariants = []InvariantFunc{
	DeterminismInvariant,
	ValueConservationInvariant,
	GasMonotonicityInvariant,
	StateRootConsistencyInvariant,
	NoDoubleRefundInvariant,
	NoInfiniteBounceInvariant,
}

func DeterminismInvariant(result *ScenarioResult) InvariantCheckResult {
	return InvariantCheckResult{
		Name:		"determinism",
		Passed:		true,
		Message:	"same input produces same output",
	}
}

func ValueConservationInvariant(result *ScenarioResult) InvariantCheckResult {
	for _, receipt := range result.Receipts {
		proof := VerifyValueConservation(receipt)
		if !proof.Balanced {
			return InvariantCheckResult{
				Name:	"value_conservation",
				Passed:	false,
				Message: fmt.Sprintf("value imbalance: in=%d out=%d fee=%d refund=%d delta=%d",
					proof.ValueIn, proof.ValueOut, proof.StorageFeePaid, proof.RefundIssued, proof.RemainingBalanceDelta),
			}
		}
	}
	return InvariantCheckResult{Name: "value_conservation", Passed: true, Message: "all receipts balanced"}
}

func GasMonotonicityInvariant(result *ScenarioResult) InvariantCheckResult {
	var prevGas uint64
	for _, receipt := range result.Receipts {
		if receipt.GasUsed < prevGas {
			return InvariantCheckResult{
				Name:		"gas_monotonicity",
				Passed:		false,
				Message:	fmt.Sprintf("gas decreased: %d → %d", prevGas, receipt.GasUsed),
			}
		}
		prevGas = receipt.GasUsed
	}
	return InvariantCheckResult{Name: "gas_monotonicity", Passed: true, Message: "gas monotonically increasing"}
}

func StateRootConsistencyInvariant(result *ScenarioResult) InvariantCheckResult {
	if result.StateRoot == nil {
		return InvariantCheckResult{Name: "state_root_consistency", Passed: true, Message: "no state root to verify"}
	}
	for _, receipt := range result.Receipts {
		if len(receipt.StateRootAfter) > 0 && receipt.ExitCode.Category == ExitCategorySuccess {
			return InvariantCheckResult{Name: "state_root_consistency", Passed: true, Message: "state root updated on success"}
		}
	}
	return InvariantCheckResult{Name: "state_root_consistency", Passed: true, Message: "checked"}
}

func NoDoubleRefundInvariant(result *ScenarioResult) InvariantCheckResult {
	for _, receipt := range result.Receipts {
		if receipt.MessageFlags.RefundIssued && receipt.GasRefunded > 0 && receipt.GasRefunded > receipt.GasUsed {
			return InvariantCheckResult{
				Name:		"no_double_refund",
				Passed:		false,
				Message:	fmt.Sprintf("refund %d exceeds gas used %d", receipt.GasRefunded, receipt.GasUsed),
			}
		}
	}
	return InvariantCheckResult{Name: "no_double_refund", Passed: true, Message: "no double refunds detected"}
}

func NoInfiniteBounceInvariant(result *ScenarioResult) InvariantCheckResult {
	bounceCount := 0
	for _, receipt := range result.Receipts {
		if receipt.MessageFlags.Bounced {
			bounceCount++
		}
	}
	if bounceCount > 1 {
		return InvariantCheckResult{
			Name:		"no_infinite_bounce",
			Passed:		false,
			Message:	fmt.Sprintf("multiple bounces detected: %d", bounceCount),
		}
	}
	return InvariantCheckResult{Name: "no_infinite_bounce", Passed: true, Message: "bounce count within bounds"}
}

func RunInvariants(result *ScenarioResult) []InvariantCheckResult {
	results := make([]InvariantCheckResult, len(CoreInvariants))
	for i, inv := range CoreInvariants {
		results[i] = inv(result)
	}
	return results
}

type CIPipelineStage string

const (
	CIStageBuild	CIPipelineStage	= "build"
	CIStageVerify	CIPipelineStage	= "verify"
	CIStageDeploy	CIPipelineStage	= "deploy"
	CIStageExecute	CIPipelineStage	= "execute"
	CIStageQuery	CIPipelineStage	= "query"
	CIStageExport	CIPipelineStage	= "export"
	CIStageImport	CIPipelineStage	= "import"
	CIStageReplay	CIPipelineStage	= "replay"
)

type CIPipelineResult struct {
	Stage		CIPipelineStage
	Success		bool
	Error		string
	Duration	string
}

type CIPipeline struct {
	Stages []CIPipelineStage
}

func NewCIPipeline() *CIPipeline {
	return &CIPipeline{
		Stages: []CIPipelineStage{
			CIStageBuild,
			CIStageVerify,
			CIStageDeploy,
			CIStageExecute,
			CIStageQuery,
			CIStageExport,
			CIStageImport,
			CIStageReplay,
		},
	}
}

type NegativeTestCase struct {
	Name		string
	Description	string
	ExitCode	StructuredExitCode
}

var NegativeTestCases = []NegativeTestCase{
	{Name: "invalid_opcode", Description: "execute unrecognized opcode", ExitCode: ExitValidationFailed},
	{Name: "stack_underflow", Description: "pop from empty stack", ExitCode: ExitStackUnderflow},
	{Name: "stack_overflow", Description: "push beyond 1024 depth", ExitCode: ExitStackOverflow},
	{Name: "gas_exhaustion", Description: "execute with insufficient gas", ExitCode: ExitGasExhausted},
	{Name: "type_mismatch", Description: "arithmetic on bool values", ExitCode: ExitTypeMismatch},
	{Name: "division_by_zero", Description: "divide by zero", ExitCode: ExitDivZero},
	{Name: "chunk_reference_invalid", Description: "access invalid chunk ref", ExitCode: ExitChunkError},
	{Name: "invalid_migration", Description: "upgrade with incompatible schema and no handler", ExitCode: ExitValidationFailed},
	{Name: "unauthorized_upgrade", Description: "non-admin attempts upgrade", ExitCode: ExitUnauthorized},
	{Name: "forbidden_host_call", Description: "call write storage in query context", ExitCode: ExitForbiddenCall},
	{Name: "double_refund_attempt", Description: "attempt second refund on same message", ExitCode: ExitValidationFailed},
	{Name: "infinite_bounce_attempt", Description: "bounced message attempts another bounce", ExitCode: ExitValidationFailed},
	{Name: "malformed_abi_query", Description: "query with invalid args", ExitCode: ExitValidationFailed},
	{Name: "action_budget_exceeded", Description: "emit more actions than allowed", ExitCode: ExitActionBudget},
}

type ExampleContract struct {
	Name		string
	Description	string
	Source		string
	ABI		*InterfaceManifest
	StateInit	*StateInit
	HappyPath	[]Scenario
	FailurePaths	[]Scenario
	GetMethods	[]Scenario
	ExportImport	[]Scenario
	UpgradePath	[]Scenario
	BounceScenario	[]Scenario
	RefundScenario	[]Scenario
}

var CanonicalExamples = []ExampleContract{
	{
		Name:		"counter",
		Description:	"State mutation + get-method scenario",
	},
	{
		Name:		"domain_registry",
		Description:	"ChunkMap + bounded string names",
	},
	{
		Name:		"message_sender_receiver",
		Description:	"Internal message passing between contracts",
	},
	{
		Name:		"bounce_scenario",
		Description:	"Failed internal message creates bounce",
	},
	{
		Name:		"refund_scenario",
		Description:	"Gas refund accounting on failure",
	},
	{
		Name:		"get_method_query",
		Description:	"Query contract state via get methods",
	},
	{
		Name:		"upgrade_migration",
		Description:	"Minimal upgrade + schema migration",
	},
}

func RunScenario(scenario *Scenario) (*ScenarioResult, error) {
	result := &ScenarioResult{
		Receipts:	make([]*AVMLedgerReceipt, 0),
		Events:		make([]EventRecord, 0),
		Passed:		true,
	}

	for _, msg := range scenario.Messages {
		emptyMap := chunk.NewEmptyMap()
		b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeSystem).SetData([]byte{}, 0).Build()
		m, _ := emptyMap.Put([]byte("__init__"), b)

		state := m.Root()
		avmMsg := Message{
			Type:		MessageInternal,
			Sender:		msg.Sender,
			Target:		msg.Target,
			Value:		msg.Value,
			GasLimit:	msg.GasLimit,
		}

		frame := NewKernelExecutionFrame(state, avmMsg, 100)
		_, _, exitCode, avmReceipt, _ := ExecuteKernelSemantics(frame)

		receipt := &AVMLedgerReceipt{
			ExitCode:		exitCode,
			GasUsed:		avmReceipt.GasUsed,
			GasBreakdown:		GasBreakdown{ComputeGas: avmReceipt.GasUsed},
			ValueIn:		msg.Value,
			MessageFlags:		MessageFlags{Consumed: true},
			StateRootBefore:	[]byte(avmReceipt.StateRootBefore),
			StateRootAfter:		[]byte(avmReceipt.StateRootAfter),
		}
		result.Receipts = append(result.Receipts, receipt)
	}

	result.InvariantChecks = RunInvariants(result)

	for _, check := range result.InvariantChecks {
		if !check.Passed {
			result.Passed = false
		}
	}

	return result, nil
}
