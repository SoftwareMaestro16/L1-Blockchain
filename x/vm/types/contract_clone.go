package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

func clonePolicy(policy ContractZonePolicy) ContractZonePolicy {
	policy.GovernanceAuthority = cloneAddress(policy.GovernanceAuthority)
	policy.UploadAllowlist = cloneAddressList(policy.UploadAllowlist)
	policy.Runtime.AVMParams.GasSchedule = cloneGasSchedule(policy.Runtime.AVMParams.GasSchedule)
	return policy
}

func cloneCode(code ContractCode) ContractCode {
	code.CodeHash = append([]byte(nil), code.CodeHash...)
	code.Owner = cloneAddress(code.Owner)
	code.AVMModule = cloneAVMModule(code.AVMModule)
	return code
}

func cloneCodes(codes []ContractCode) []ContractCode {
	out := make([]ContractCode, len(codes))
	for i, code := range codes {
		out[i] = cloneCode(code)
	}
	return out
}

func cloneContract(contract ContractInstance) ContractInstance {
	contract.Address = cloneAddress(contract.Address)
	contract.Owner = cloneAddress(contract.Owner)
	contract.Admin = cloneAddress(contract.Admin)
	contract.Storage = cloneStorageEntries(contract.Storage)
	return contract
}

func cloneContracts(contracts []ContractInstance) []ContractInstance {
	out := make([]ContractInstance, len(contracts))
	for i, contract := range contracts {
		out[i] = cloneContract(contract)
	}
	return out
}

func cloneQueuedMessage(msg async.MessageEnvelope) async.MessageEnvelope {
	msg.Source = cloneAddress(msg.Source)
	msg.Destination = cloneAddress(msg.Destination)
	msg.Body = append([]byte(nil), msg.Body...)
	return msg
}

func cloneQueuedMessages(messages []async.MessageEnvelope) []async.MessageEnvelope {
	out := make([]async.MessageEnvelope, len(messages))
	for i, msg := range messages {
		out[i] = cloneQueuedMessage(msg)
	}
	return out
}

func cloneReceipt(receipt ContractReceipt) ContractReceipt {
	receipt.Contract = cloneAddress(receipt.Contract)
	receipt.QueryResponse = append([]byte(nil), receipt.QueryResponse...)
	return receipt
}

func cloneReceipts(receipts []ContractReceipt) []ContractReceipt {
	out := make([]ContractReceipt, len(receipts))
	for i, receipt := range receipts {
		out[i] = cloneReceipt(receipt)
	}
	return out
}

func cloneAVMModule(module avm.Module) avm.Module {
	module.Imports = append([]avm.HostFunction(nil), module.Imports...)
	module.Exports = cloneExports(module.Exports)
	module.Code = cloneInstructions(module.Code)
	return module
}

func cloneExports(exports map[avm.Entrypoint]uint32) map[avm.Entrypoint]uint32 {
	out := make(map[avm.Entrypoint]uint32, len(exports))
	for key, value := range exports {
		out[key] = value
	}
	return out
}

func cloneInstructions(instructions []avm.Instruction) []avm.Instruction {
	out := make([]avm.Instruction, len(instructions))
	for i, ins := range instructions {
		ins.Data = append([]byte(nil), ins.Data...)
		out[i] = ins
	}
	return out
}

func cloneGasSchedule(schedule map[avm.Opcode]uint64) map[avm.Opcode]uint64 {
	out := make(map[avm.Opcode]uint64, len(schedule))
	for key, value := range schedule {
		out[key] = value
	}
	return out
}

func cloneAddress(addr sdk.AccAddress) sdk.AccAddress {
	if len(addr) == 0 {
		return nil
	}
	return append(sdk.AccAddress(nil), addr...)
}

func cloneAddressList(addresses []sdk.AccAddress) []sdk.AccAddress {
	out := make([]sdk.AccAddress, len(addresses))
	for i, addr := range addresses {
		out[i] = cloneAddress(addr)
	}
	return out
}
