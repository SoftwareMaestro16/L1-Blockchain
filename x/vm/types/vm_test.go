package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

const governance = "4:0000000000000000000000000000000000000000000000000000000000000001"

func TestValidateAVMCallAndEntrypoints(t *testing.T) {
	policy := DefaultRuntimePolicy()
	require.NoError(t, ValidateVMCall(VMCall{Runtime: RuntimeAVM, Action: ActionDeploy, GasLimit: 1, Entrypoint: avm.EntryDeploy}, policy))
	entry, err := AVMEntrypointForAction(ActionInternalCall, false)
	require.NoError(t, err)
	require.Equal(t, avm.EntryReceiveInternal, entry)
	entry, err = AVMEntrypointForAction(ActionInternalCall, true)
	require.NoError(t, err)
	require.Equal(t, avm.EntryReceiveBounced, entry)

	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: RuntimeAVM, Action: ActionDeploy, GasLimit: 0}, policy), "gas limit")
	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: RuntimeAVM, Action: ActionDeploy, GasLimit: 1, Entrypoint: avm.EntryQuery}, policy), "entrypoint")
}

func TestCosmWasmDisabledByDefaultAndGatedWhenEnabled(t *testing.T) {
	policy := DefaultRuntimePolicy()
	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: RuntimeCosmWasm, Action: ActionDeploy, CodeBytes: 1, GasLimit: 1}, policy), "disabled")

	policy.CosmWasmEnabled = true
	policy.CosmWasmPolicy = wasmconfig.DefaultPolicy()
	policy.CosmWasmPolicy.Enabled = true
	policy.CosmWasmPolicy.GovernanceAuthority = governance
	require.NoError(t, ValidateVMCall(VMCall{Runtime: RuntimeCosmWasm, Action: ActionDeploy, CodeBytes: 1, GasLimit: 1}, policy))
	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: RuntimeCosmWasm, Action: ActionDeploy, CodeBytes: wasmconfig.DefaultMaxContractSizeBytes + 1, GasLimit: 1}, policy), "code size")
	require.NoError(t, ValidateVMCall(VMCall{Runtime: RuntimeCosmWasm, Action: ActionQuery, GasLimit: wasmconfig.DefaultSmartQueryGasLimit, QueryBytes: 1, QueryDepth: 1}, policy))
}

func TestRuntimePolicyRequiresEnabledRuntimeAndValidAction(t *testing.T) {
	policy := DefaultRuntimePolicy()
	policy.AVMEnabled = false
	require.ErrorContains(t, ValidateRuntimePolicy(policy), "at least one")
	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: RuntimeAVM, Action: "bad", GasLimit: 1}, DefaultRuntimePolicy()), "invalid VM action")
	require.ErrorContains(t, ValidateVMCall(VMCall{Runtime: "bad", Action: ActionDeploy, GasLimit: 1}, DefaultRuntimePolicy()), "invalid VM runtime")
}
