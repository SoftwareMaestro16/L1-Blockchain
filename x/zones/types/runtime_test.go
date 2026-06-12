package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZoneRuntimeStateIsolatesQueuesBudgetsNamespacesAndProofRoot(t *testing.T) {
	zone := testZone(ZoneIDContract, ZoneKindContract, VMPolicyAVM, 1)
	queue := []ZoneMessage{
		testZoneMessage(ZoneIDContract, "contract.execute", 2, 25_000),
		testZoneMessage(ZoneIDContract, "contract.execute", 1, 10_000),
	}
	budget := DefaultZoneExecutionBudget()
	budget.MaxGas = 100_000
	budget.MaxMessages = 4
	gas := DefaultZoneGasPolicy()
	gas.MaxGasPerBlock = 100_000
	gas.MaxGasPerMessage = 50_000
	filter := ZoneMessageFilter{AllowedMessageTypes: []string{"contract.execute"}}

	runtime, err := NewZoneRuntimeState(zone, hash("contract-state"), queue, budget, gas, filter)
	require.NoError(t, err)
	require.Equal(t, ZoneStateNamespace(ZoneIDContract), runtime.StateNamespace)
	require.Equal(t, ZoneQueueNamespace(ZoneIDContract), runtime.QueueNamespace)
	require.Equal(t, ZoneProofNamespace(ZoneIDContract), runtime.ProofNamespace)
	require.Equal(t, ZoneQueryNamespace(ZoneIDContract), runtime.QueryNamespace)
	require.Equal(t, ZoneKVPrefix(ZoneIDContract), runtime.KVPrefix)
	require.Equal(t, ZoneExecutionPipeline(ZoneIDContract), runtime.ExecutionPipeline)
	require.Equal(t, ComputeZoneModuleSetRoot(runtime.ModuleSet), runtime.ModuleSetRoot)
	require.Equal(t, uint64(1), runtime.MessageQueue[0].Sequence)
	require.Equal(t, ComputeZoneMessageRoot(runtime.MessageQueue), runtime.MessageRoot)
	require.Equal(t, ComputeZoneRuntimeProofRoot(runtime), runtime.ProofRoot)

	commitment, err := BuildCommitmentFromRuntime(runtime, 11, "")
	require.NoError(t, err)
	require.Equal(t, runtime.StateRoot, commitment.StateRoot)
	require.Equal(t, runtime.MessageRoot, commitment.MessageRoot)
	require.Equal(t, runtime.ExecutionResultRoot, commitment.ExecutionResultRoot)
}

func TestZoneRuntimeRejectsCrossZoneMessagesNamespaceDriftAndFilterViolations(t *testing.T) {
	zone := testZone(ZoneIDContract, ZoneKindContract, VMPolicyAVM, 1)
	filter := ZoneMessageFilter{AllowedMessageTypes: []string{"contract.execute"}}

	_, err := NewZoneRuntimeState(
		zone,
		hash("contract-state"),
		[]ZoneMessage{testZoneMessage(ZoneIDFinancial, "contract.execute", 1, 1)},
		DefaultZoneExecutionBudget(),
		DefaultZoneGasPolicy(),
		filter,
	)
	require.ErrorContains(t, err, "expected CONTRACT_ZONE")

	_, err = NewZoneRuntimeState(
		zone,
		hash("contract-state"),
		[]ZoneMessage{testZoneMessage(ZoneIDContract, "contract.migrate", 1, 1)},
		DefaultZoneExecutionBudget(),
		DefaultZoneGasPolicy(),
		filter,
	)
	require.ErrorContains(t, err, "not allowed")

	runtime, err := NewZoneRuntimeState(zone, hash("contract-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	runtime.StateNamespace = ZoneStateNamespace(ZoneIDFinancial)
	runtime.ProofRoot = ComputeZoneRuntimeProofRoot(runtime)
	require.ErrorContains(t, runtime.Validate(), "state namespace")

	runtime, err = NewZoneRuntimeState(zone, hash("contract-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	runtime.QueryNamespace = ZoneQueryNamespace(ZoneIDFinancial)
	runtime.ProofRoot = ComputeZoneRuntimeProofRoot(runtime)
	require.ErrorContains(t, runtime.Validate(), "query namespace")

	runtime, err = NewZoneRuntimeState(zone, hash("contract-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	runtime.ModuleSet = []string{"vm:AVM", "kind:CONTRACT"}
	runtime.ModuleSetRoot = ComputeZoneModuleSetRoot(runtime.ModuleSet)
	runtime.ProofRoot = ComputeZoneRuntimeProofRoot(runtime)
	require.ErrorContains(t, runtime.Validate(), "module set")
}

func TestParallelExecutionPlanRejectsOverlappingZonePrefixes(t *testing.T) {
	contract, err := NewZoneRuntimeState(testZone(ZoneIDContract, ZoneKindContract, VMPolicyAVM, 1), hash("contract-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	financial, err := NewZoneRuntimeState(testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1), hash("financial-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)

	plan, err := BuildParallelExecutionPlan(10, []ZoneRuntimeState{financial, contract})
	require.NoError(t, err)
	require.Equal(t, []ZoneID{ZoneIDContract, ZoneIDFinancial}, []ZoneID{plan.Zones[0].ZoneID, plan.Zones[1].ZoneID})

	drifted := financial
	drifted.KVPrefix = contract.KVPrefix + "nested/"
	drifted.ProofRoot = ComputeZoneRuntimeProofRoot(drifted)
	_, err = BuildParallelExecutionPlan(10, []ZoneRuntimeState{contract, drifted})
	require.ErrorContains(t, err, "KV prefix")
}

func TestZoneBudgetAndGasPolicyBounds(t *testing.T) {
	budget := ZoneExecutionBudget{MaxGas: 10, MaxMessages: 1}
	_, err := budget.Consume(11, 1)
	require.ErrorContains(t, err, "gas used")

	gas := DefaultZoneGasPolicy()
	gas.MaxGasPerMessage = gas.MaxGasPerBlock + 1
	require.ErrorContains(t, gas.Validate(), "per-message")
}

func testZoneMessage(zoneID ZoneID, messageType string, sequence uint64, gasLimit uint64) ZoneMessage {
	return ZoneMessage{
		ZoneID:		zoneID,
		MessageType:	messageType,
		Source:		"source",
		Destination:	"destination",
		GasLimit:	gasLimit,
		PayloadHash:	hash(string(zoneID) + messageType + string(rune(sequence))),
		Sequence:	sequence,
	}
}
