package types

import (
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/app/wasmconfig"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

const (
	RuntimeAVM	= "avm"
	RuntimeCosmWasm	= "cosmwasm"

	ActionDeploy		= "deploy"
	ActionExternalCall	= "external_call"
	ActionInternalCall	= "internal_call"
	ActionBouncedCall	= "bounced_call"
	ActionQuery		= "query"
)

type RuntimePolicy struct {
	AVMParams	avm.Params
	CosmWasmPolicy	wasmconfig.Policy
	AVMEnabled	bool
	CosmWasmEnabled	bool
}

type VMCall struct {
	Runtime		string
	Action		string
	CodeBytes	uint64
	GasLimit	uint64
	QueryBytes	uint64
	QueryDepth	uint32
	Entrypoint	avm.Entrypoint
}

func DefaultRuntimePolicy() RuntimePolicy {
	return RuntimePolicy{
		AVMParams:		avm.DefaultParams(),
		CosmWasmPolicy:		wasmconfig.DefaultPolicy(),
		AVMEnabled:		true,
		CosmWasmEnabled:	false,
	}
}

func ValidateRuntimePolicy(policy RuntimePolicy) error {
	if policy.AVMEnabled {
		if err := policy.AVMParams.Validate(); err != nil {
			return err
		}
	}
	if policy.CosmWasmEnabled {
		cw := policy.CosmWasmPolicy
		cw.Enabled = true
		if err := cw.Validate(); err != nil {
			return err
		}
	}
	if !policy.AVMEnabled && !policy.CosmWasmEnabled {
		return errors.New("at least one VM runtime must be enabled for VM routing")
	}
	return nil
}

func ValidateVMCall(call VMCall, policy RuntimePolicy) error {
	if err := ValidateRuntimePolicy(policy); err != nil {
		return err
	}
	if !IsVMAction(call.Action) {
		return fmt.Errorf("invalid VM action %q", call.Action)
	}
	switch call.Runtime {
	case RuntimeAVM:
		return validateAVMCall(call, policy)
	case RuntimeCosmWasm:
		return validateCosmWasmCall(call, policy)
	default:
		return fmt.Errorf("invalid VM runtime %q", call.Runtime)
	}
}

func AVMEntrypointForAction(action string, bounced bool) (avm.Entrypoint, error) {
	switch action {
	case ActionDeploy:
		return avm.EntryDeploy, nil
	case ActionExternalCall:
		return avm.EntryReceiveExternal, nil
	case ActionInternalCall:
		if bounced {
			return avm.EntryReceiveBounced, nil
		}
		return avm.EntryReceiveInternal, nil
	case ActionBouncedCall:
		return avm.EntryReceiveBounced, nil
	case ActionQuery:
		return avm.EntryQuery, nil
	default:
		return 0, fmt.Errorf("invalid VM action %q", action)
	}
}

func IsVMAction(action string) bool {
	switch action {
	case ActionDeploy, ActionExternalCall, ActionInternalCall, ActionBouncedCall, ActionQuery:
		return true
	default:
		return false
	}
}

func validateAVMCall(call VMCall, policy RuntimePolicy) error {
	if !policy.AVMEnabled {
		return errors.New("AVM runtime is disabled")
	}
	if call.GasLimit == 0 {
		return errors.New("AVM call gas limit must be positive")
	}
	if call.CodeBytes > uint64(policy.AVMParams.MaxCodeBytes) {
		return fmt.Errorf("AVM code bytes must be <= %d", policy.AVMParams.MaxCodeBytes)
	}
	entry, err := AVMEntrypointForAction(call.Action, call.Action == ActionBouncedCall)
	if err != nil {
		return err
	}
	if call.Entrypoint != 0 && call.Entrypoint != entry {
		return errors.New("AVM entrypoint does not match action")
	}
	return nil
}

func validateCosmWasmCall(call VMCall, policy RuntimePolicy) error {
	if !policy.CosmWasmEnabled {
		return errors.New("CosmWasm runtime is disabled")
	}
	cw := policy.CosmWasmPolicy
	cw.Enabled = true
	switch call.Action {
	case ActionDeploy:
		return wasmconfig.ValidateContractCodeSize(call.CodeBytes, false, cw)
	case ActionQuery:
		return wasmconfig.EnforceQueryLimit(call.GasLimit, call.QueryBytes, call.QueryDepth, cw)
	default:
		return wasmconfig.EnforceExecuteGasLimit(call.GasLimit, cw)
	}
}
