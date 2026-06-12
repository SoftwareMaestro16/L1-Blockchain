package types

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetravm/avm"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestExecutionRouterPlanClassifiesBudgetsAndDispatchesComponentMap(t *testing.T) {
	contract := routerRuntime(t, routerZone(zonestypes.ZoneIDContract, zonestypes.ZoneKindContract, zonestypes.VMPolicyAVM), []string{"contract.execute"})
	application := routerRuntime(t, routerZone(zonestypes.ZoneIDApplication, zonestypes.ZoneKindApplication, zonestypes.VMPolicyAVM), []string{"async.deliver"})
	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	policy := DefaultRuntimePolicy()

	plan, err := BuildExecutionRouterPlan(42, binding, []zonestypes.ZoneRuntimeState{contract, application}, []ExecutionRouterMessage{
		{
			Sequence:	2,
			MsgType:	"contract.execute",
			TargetModule:	"vm",
			TargetContract:	"counter",
			Source:		"wallet",
			Destination:	"contract",
			PayloadHash:	routerHash("contract-payload"),
			GasClass:	RouterGasClassHigh,
			Priority:	10,
			DomainRouteKey:	"counter.execute",
			StakingPower:	1,
			Call: VMCall{
				Runtime:	RuntimeAVM,
				Action:		ActionExternalCall,
				GasLimit:	7,
				Entrypoint:	avm.EntryReceiveExternal,
			},
		},
		{
			Sequence:	1,
			MsgType:	"async.deliver",
			TargetActor:	"counter-worker",
			Source:		"queue",
			Destination:	"actor",
			PayloadHash:	routerHash("async-payload"),
			GasClass:	RouterGasClassStandard,
			DomainRouteKey:	"counter.resume",
			StakingPower:	1,
			Scheduling: RouterSchedulingMetadata{
				DeliverAtBlock:		42,
				DeadlineBlock:		45,
				MaxRetries:		2,
				RetryDelayBlocks:	1,
			},
			Call: VMCall{
				Runtime:	RuntimeAVM,
				Action:		ActionInternalCall,
				GasLimit:	5,
				Entrypoint:	avm.EntryReceiveInternal,
			},
		},
	}, policy)
	require.NoError(t, err)
	require.NoError(t, plan.Validate(policy))
	require.Len(t, plan.Dispatches, 2)
	require.Equal(t, zonestypes.ZoneIDApplication, plan.Dispatches[0].ZoneID)
	require.Equal(t, RouterLaneAsync, plan.Dispatches[0].Lane)
	require.Equal(t, RouterBackendAVMActor, plan.Dispatches[0].Backend)
	require.Equal(t, RouterDispatchModeQueued, plan.Dispatches[0].DispatchMode)
	require.Equal(t, RouterReceiptDeferred, plan.Dispatches[0].ReceiptPolicy)
	require.Equal(t, "counter-worker", plan.Dispatches[0].ExecutionTarget)
	require.Contains(t, plan.Dispatches[0].QueueID, "counter.resume")
	require.Equal(t, RouterGasMeter{Class: RouterGasClassStandard, Limit: 5, Reserved: 5}, plan.Dispatches[0].GasMeter)
	require.Equal(t, zonestypes.ZoneIDContract, plan.Dispatches[1].ZoneID)
	require.Equal(t, RouterLaneSync, plan.Dispatches[1].Lane)
	require.Equal(t, RouterDispatchModeDirect, plan.Dispatches[1].DispatchMode)
	require.Equal(t, RouterReceiptCommit, plan.Dispatches[1].ReceiptPolicy)
	require.Equal(t, "counter", plan.Dispatches[1].ExecutionTarget)
	require.Equal(t, "direct", plan.Dispatches[1].QueueID)
	require.Equal(t, RouterGasMeter{Class: RouterGasClassHigh, Limit: 7, Reserved: 7}, plan.Dispatches[1].GasMeter)
	require.Equal(t, uint64(42), plan.SDKPlan.Dispatches[0].ExecutionHeight)
	require.Equal(t, plan.Dispatches[0].BlockSTMKey, plan.SDKPlan.Dispatches[0].BlockSTMKey)

	outputs := routerOutputsByZone(plan.ZoneOutputs)
	require.Equal(t, uint64(5), outputs[zonestypes.ZoneIDApplication].Budget.GasUsed)
	require.Equal(t, uint32(1), outputs[zonestypes.ZoneIDApplication].Budget.MessagesUsed)
	require.Equal(t, uint64(7), outputs[zonestypes.ZoneIDContract].Budget.GasUsed)
	require.Equal(t, uint32(1), outputs[zonestypes.ZoneIDContract].Budget.MessagesUsed)
	require.NotEmpty(t, plan.PlanRoot)

	mutated := plan
	mutated.Dispatches = append([]ExecutionRouterDispatch(nil), plan.Dispatches...)
	mutated.Dispatches[0].ExecutionHeight++
	require.NotEqual(t, plan.PlanRoot, ComputeExecutionRouterPlanRoot(mutated))
}

func TestExecutionRouterRoutesDesignRulesAndCrossZoneWrites(t *testing.T) {
	require.Equal(t, zonestypes.ZoneIDFinancial, mustClassifyRouterZone(t, "financial.transfer"))
	require.Equal(t, zonestypes.ZoneIDIdentity, mustClassifyRouterZone(t, "identity.issue"))
	require.Equal(t, zonestypes.ZoneIDIdentity, mustClassifyRouterZone(t, "resolver.lookup"))
	require.Equal(t, zonestypes.ZoneIDApplication, mustClassifyRouterZone(t, "scheduler.enqueue"))
	require.Equal(t, zonestypes.ZoneIDApplication, mustClassifyRouterZone(t, "workflow.resume"))
	require.Equal(t, zonestypes.ZoneIDContract, mustClassifyRouterZone(t, "contract.execute"))

	contract := routerRuntime(t, routerZone(zonestypes.ZoneIDContract, zonestypes.ZoneKindContract, zonestypes.VMPolicyAVM), []string{"contract.execute"})
	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	policy := DefaultRuntimePolicy()

	plan, err := BuildExecutionRouterPlan(7, binding, []zonestypes.ZoneRuntimeState{contract}, []ExecutionRouterMessage{{
		Sequence:	1,
		MsgType:	"contract.execute",
		SourceZoneID:	zonestypes.ZoneIDFinancial,
		TargetContract:	"escrow",
		Source:		"financial-module",
		Destination:	"contract",
		PayloadHash:	routerHash("cross-zone"),
		GasClass:	RouterGasClassLow,
		DomainRouteKey:	"financial.to.contract",
		StakingPower:	1,
		Call: VMCall{
			Runtime:	RuntimeAVM,
			Action:		ActionExternalCall,
			GasLimit:	3,
			Entrypoint:	avm.EntryReceiveExternal,
		},
	}}, policy)
	require.NoError(t, err)
	require.Len(t, plan.Dispatches, 1)
	require.Equal(t, RouterLaneAsync, plan.Dispatches[0].Lane)
	require.Equal(t, RouterDispatchModeCrossZone, plan.Dispatches[0].DispatchMode)
	require.Equal(t, RouterReceiptDeferred, plan.Dispatches[0].ReceiptPolicy)
	require.Contains(t, plan.Dispatches[0].QueueID, "cross_zone_async")
	require.Equal(t, "escrow", plan.Dispatches[0].ExecutionTarget)
}

func TestExecutionRouterRejectsInvalidFilterBudgetAndConflicts(t *testing.T) {
	contract := routerRuntime(t, routerZone(zonestypes.ZoneIDContract, zonestypes.ZoneKindContract, zonestypes.VMPolicyAVM), []string{"contract.execute"})
	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	policy := DefaultRuntimePolicy()
	valid := ExecutionRouterMessage{
		Sequence:	1,
		MsgType:	"contract.execute",
		Source:		"wallet",
		Destination:	"contract",
		PayloadHash:	routerHash("payload"),
		StakingPower:	1,
		Call: VMCall{
			Runtime:	RuntimeAVM,
			Action:		ActionExternalCall,
			GasLimit:	10,
			Entrypoint:	avm.EntryReceiveExternal,
		},
	}

	disallowed := valid
	disallowed.MsgType = "contract.migrate"
	_, err := BuildExecutionRouterPlan(1, binding, []zonestypes.ZoneRuntimeState{contract}, []ExecutionRouterMessage{disallowed}, policy)
	require.ErrorContains(t, err, "not allowed")
	require.Equal(t, uint64(0), contract.Budget.GasUsed)

	overBudget := contract
	overBudget.Budget = zonestypes.ZoneExecutionBudget{MaxGas: 9, MaxMessages: 1}
	overBudget.ProofRoot = zonestypes.ComputeZoneRuntimeProofRoot(overBudget)
	_, err = BuildExecutionRouterPlan(1, binding, []zonestypes.ZoneRuntimeState{overBudget}, []ExecutionRouterMessage{valid}, policy)
	require.ErrorContains(t, err, "gas used exceeds")

	conflictingA := valid
	conflictingA.BlockSTMKey = "vm/CONTRACT_ZONE/router/a"
	conflictingB := valid
	conflictingB.Sequence = 2
	conflictingB.BlockSTMKey = "vm/CONTRACT_ZONE/router/a/storage"
	_, err = BuildExecutionRouterPlan(1, binding, []zonestypes.ZoneRuntimeState{contract}, []ExecutionRouterMessage{conflictingA, conflictingB}, policy)
	require.ErrorContains(t, err, "BlockSTM conflict")

	unclassified := valid
	unclassified.MsgType = "unknown.execute"
	unclassified.Call.Runtime = ""
	_, err = BuildExecutionRouterPlan(1, binding, []zonestypes.ZoneRuntimeState{contract}, []ExecutionRouterMessage{unclassified}, policy)
	require.ErrorContains(t, err, "cannot classify")
}

func mustClassifyRouterZone(t *testing.T, msgType string) zonestypes.ZoneID {
	t.Helper()
	zoneID, err := ClassifyExecutionZone(ExecutionRouterMessage{MsgType: msgType})
	require.NoError(t, err)
	return zoneID
}

func routerRuntime(t *testing.T, zone zonestypes.Zone, allowed []string) zonestypes.ZoneRuntimeState {
	t.Helper()
	runtime, err := zonestypes.NewZoneRuntimeState(
		zone,
		routerHash(string(zone.ID)+"-state"),
		nil,
		zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
		zonestypes.DefaultZoneGasPolicy(),
		zonestypes.ZoneMessageFilter{AllowedMessageTypes: allowed},
	)
	require.NoError(t, err)
	return runtime
}

func routerZone(id zonestypes.ZoneID, kind zonestypes.ZoneKind, vm zonestypes.VMPolicy) zonestypes.Zone {
	return zonestypes.Zone{
		ID:			id,
		Kind:			kind,
		VMPolicy:		vm,
		FeePolicy:		zonestypes.FeePolicyNaet,
		GenesisStateHash:	routerHash(string(id)),
		StateTransitionID:	"avm-router",
		UpgradePolicy:		zonestypes.UpgradePolicyGovernance,
		DataAvailabilityPolicy:	zonestypes.DataAvailabilityCoreCommitment,
		AuditStatus:		zonestypes.AuditStatusInternalReview,
		ActivationHeight:	1,
	}
}

func routerOutputsByZone(outputs []ExecutionRouterZoneOutput) map[zonestypes.ZoneID]ExecutionRouterZoneOutput {
	out := make(map[zonestypes.ZoneID]ExecutionRouterZoneOutput, len(outputs))
	for _, output := range outputs {
		out[output.ZoneID] = output
	}
	return out
}

func routerHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
