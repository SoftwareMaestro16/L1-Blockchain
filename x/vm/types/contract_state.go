package types

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

func EmptyContractZoneState(policy ContractZonePolicy) ContractZoneState {
	return ContractZoneState{Policy: policy, NextCodeID: 1, NextReceiptSeq: 1}
}

func UploadAVMCode(state ContractZoneState, actor sdk.AccAddress, module avm.Module, height uint64) (ContractZoneState, ContractCode, error) {
	state = normalizeContractState(state)
	if err := state.Validate(); err != nil {
		return ContractZoneState{}, ContractCode{}, err
	}
	if err := CanUploadContract(actor, state.Policy); err != nil {
		return ContractZoneState{}, ContractCode{}, err
	}
	if height == 0 {
		return ContractZoneState{}, ContractCode{}, errors.New("contract upload height must be positive")
	}
	verifier, err := avm.NewVerifier(state.Policy.Runtime.AVMParams)
	if err != nil {
		return ContractZoneState{}, ContractCode{}, err
	}
	if err := verifier.Verify(module); err != nil {
		return ContractZoneState{}, ContractCode{}, err
	}
	encoded, err := avm.EncodeModule(module)
	if err != nil {
		return ContractZoneState{}, ContractCode{}, err
	}
	if uint64(len(encoded)) > state.Policy.Limits.MaxCodeSizeBytes {
		return ContractZoneState{}, ContractCode{}, fmt.Errorf("contract code size must be <= %d bytes", state.Policy.Limits.MaxCodeSizeBytes)
	}
	codeHash, err := avm.CodeHash(module)
	if err != nil {
		return ContractZoneState{}, ContractCode{}, err
	}
	code := ContractCode{
		CodeID:		state.NextCodeID,
		Runtime:	RuntimeAVM,
		CodeHash:	codeHash[:],
		Owner:		cloneAddress(actor),
		AVMModule:	cloneAVMModule(module),
		UploadedAt:	height,
		EncodedBytes:	uint64(len(encoded)),
	}
	next := state.Clone()
	next.Codes = append(next.Codes, code)
	next.NextCodeID++
	sortContractCodes(next.Codes)
	return next, code, next.Validate()
}

func InstantiateAVMContract(state ContractZoneState, actor sdk.AccAddress, codeID uint64, admin sdk.AccAddress, salt []byte, initial avm.Storage, height uint64, gasLimit uint64) (ContractZoneState, ContractInstance, ContractReceipt, error) {
	state = normalizeContractState(state)
	if err := state.Validate(); err != nil {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, err
	}
	code, found := findCode(state, codeID)
	if !found {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, errors.New("contract code not found")
	}
	if err := CanInstantiateContract(actor, code, state.Policy); err != nil {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, err
	}
	if err := validateContractAddress("contract admin", admin); err != nil {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, err
	}
	if height == 0 {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, errors.New("contract instantiate height must be positive")
	}
	if gasLimit == 0 || gasLimit > state.Policy.GasModel.DeployGas {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, fmt.Errorf("contract deploy gas must be in 1..%d", state.Policy.GasModel.DeployGas)
	}
	address := deriveContractAddress(codeID, actor, salt)
	if _, exists := findContract(state, address); exists {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, errors.New("contract address already exists")
	}
	namespace := ContractNamespace(address)
	storage := StorageFromAVM(namespace, initial)
	if err := ValidateStorageEntries(storage, state.Policy.Limits); err != nil {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, err
	}
	instance := ContractInstance{
		Address:	address,
		CodeID:		codeID,
		Runtime:	RuntimeAVM,
		Owner:		cloneAddress(actor),
		Admin:		cloneAddress(admin),
		Storage:	storage,
		CreatedHeight:	height,
		UpdatedHeight:	height,
	}
	exec, receipt, err := runAVMCall(state.Policy, code, instance, avm.EntryDeploy, gasLimit, height, nil, nil)
	if err != nil {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, err
	}
	if exec.ResultCode != async.ResultOK {
		return ContractZoneState{}, ContractInstance{}, receipt, errors.New("contract deploy execution failed")
	}
	instance.Storage = StorageFromAVM(namespace, exec.State)
	if err := ValidateStorageEntries(instance.Storage, state.Policy.Limits); err != nil {
		return ContractZoneState{}, ContractInstance{}, ContractReceipt{}, err
	}
	next := state.Clone()
	receipt.Sequence = next.NextReceiptSeq
	next.NextReceiptSeq++
	next.Contracts = append(next.Contracts, instance)
	next.Receipts = append(next.Receipts, receipt)
	sortContracts(next.Contracts)
	sortReceipts(next.Receipts)
	return next, instance, receipt, next.Validate()
}

func ExecuteAVMContract(state ContractZoneState, call ContractCall, height uint64) (ContractZoneState, ContractReceipt, error) {
	state = normalizeContractState(state)
	if err := state.Validate(); err != nil {
		return ContractZoneState{}, ContractReceipt{}, err
	}
	if err := validateContractAddress("contract execute actor", call.Actor); err != nil {
		return ContractZoneState{}, ContractReceipt{}, err
	}
	if call.Entrypoint == avm.EntryQuery || call.Entrypoint == avm.EntryDeploy || call.Entrypoint == avm.EntryMigrate {
		return ContractZoneState{}, ContractReceipt{}, errors.New("contract execute entrypoint is invalid")
	}
	return applyAVMCall(state, call, height, state.Policy.GasModel.ExecuteGas, true)
}

func QueryAVMContract(state ContractZoneState, call ContractCall, height uint64) (ContractReceipt, error) {
	state = normalizeContractState(state)
	if err := state.Validate(); err != nil {
		return ContractReceipt{}, err
	}
	if call.QueryDepth == 0 || call.QueryDepth > state.Policy.Limits.MaxQueryDepth {
		return ContractReceipt{}, fmt.Errorf("contract query depth must be in 1..%d", state.Policy.Limits.MaxQueryDepth)
	}
	if call.QueryResponseBytes > state.Policy.Limits.MaxQueryResponseBytes {
		return ContractReceipt{}, fmt.Errorf("contract query response bytes must be <= %d", state.Policy.Limits.MaxQueryResponseBytes)
	}
	call.Entrypoint = avm.EntryQuery
	_, receipt, err := runReadOnlyAVMCall(state, call, height, state.Policy.GasModel.QueryGas)
	if err != nil {
		return ContractReceipt{}, err
	}
	receipt.QueryResponse = make([]byte, call.QueryResponseBytes)
	return receipt, nil
}

func MigrateAVMContract(state ContractZoneState, call ContractCall, newCodeID uint64, height uint64) (ContractZoneState, ContractReceipt, error) {
	state = normalizeContractState(state)
	if err := state.Validate(); err != nil {
		return ContractZoneState{}, ContractReceipt{}, err
	}
	contract, found := findContract(state, call.Contract)
	if !found {
		return ContractZoneState{}, ContractReceipt{}, errors.New("contract not found")
	}
	if err := CanMigrateContract(call.Actor, contract, state.Policy); err != nil {
		return ContractZoneState{}, ContractReceipt{}, err
	}
	newCode, found := findCode(state, newCodeID)
	if !found {
		return ContractZoneState{}, ContractReceipt{}, errors.New("new contract code not found")
	}
	if newCode.Runtime != RuntimeAVM {
		return ContractZoneState{}, ContractReceipt{}, errors.New("new contract code must be AVM")
	}
	call.Entrypoint = avm.EntryMigrate
	next, receipt, err := applyAVMCall(state, call, height, state.Policy.GasModel.ExecuteGas, true)
	if err != nil {
		return ContractZoneState{}, ContractReceipt{}, err
	}
	contract, _ = findContract(next, call.Contract)
	contract.CodeID = newCodeID
	contract.UpdatedHeight = height
	next.Contracts = upsertContract(next.Contracts, contract)
	sortContracts(next.Contracts)
	return next, receipt, next.Validate()
}

func ImportContractZoneState(state ContractZoneState) (ContractZoneState, error) {
	state = normalizeContractState(state)
	if err := state.Validate(); err != nil {
		return ContractZoneState{}, err
	}
	return state.Export(), nil
}

func (s ContractZoneState) Export() ContractZoneState {
	out := s.Clone()
	sortContractCodes(out.Codes)
	sortContracts(out.Contracts)
	sortQueuedMessages(out.QueuedMessages)
	sortReceipts(out.Receipts)
	return out
}

func (s ContractZoneState) Clone() ContractZoneState {
	return ContractZoneState{
		Policy:			clonePolicy(s.Policy),
		Codes:			cloneCodes(s.Codes),
		Contracts:		cloneContracts(s.Contracts),
		QueuedMessages:		cloneQueuedMessages(s.QueuedMessages),
		Receipts:		cloneReceipts(s.Receipts),
		NextCodeID:		s.NextCodeID,
		NextReceiptSeq:		s.NextReceiptSeq,
		BlockMessageCount:	s.BlockMessageCount,
	}
}

func applyAVMCall(state ContractZoneState, call ContractCall, height uint64, actionLimit uint64, mutate bool) (ContractZoneState, ContractReceipt, error) {
	exec, receipt, err := runReadOnlyAVMCall(state, call, height, actionLimit)
	if err != nil {
		return ContractZoneState{}, ContractReceipt{}, err
	}
	next := state.Clone()
	receipt.Sequence = next.NextReceiptSeq
	next.NextReceiptSeq++
	if exec.ResultCode == async.ResultOK && mutate {
		contract, _ := findContract(next, call.Contract)
		contract.Storage = StorageFromAVM(ContractNamespace(contract.Address), exec.State)
		contract.UpdatedHeight = height
		next.Contracts = upsertContract(next.Contracts, contract)
		next.QueuedMessages = append(next.QueuedMessages, exec.Outgoing...)
		next.BlockMessageCount += uint32(len(exec.Outgoing))
	}
	next.Receipts = append(next.Receipts, receipt)
	sortContracts(next.Contracts)
	sortQueuedMessages(next.QueuedMessages)
	sortReceipts(next.Receipts)
	return next, receipt, next.Validate()
}

func runReadOnlyAVMCall(state ContractZoneState, call ContractCall, height uint64, actionLimit uint64) (avm.Execution, ContractReceipt, error) {
	if call.GasLimit == 0 || call.GasLimit > actionLimit {
		return avm.Execution{}, ContractReceipt{}, fmt.Errorf("contract gas must be in 1..%d", actionLimit)
	}
	if height == 0 {
		return avm.Execution{}, ContractReceipt{}, errors.New("contract call height must be positive")
	}
	contract, found := findContract(state, call.Contract)
	if !found {
		return avm.Execution{}, ContractReceipt{}, errors.New("contract not found")
	}
	code, found := findCode(state, contract.CodeID)
	if !found {
		return avm.Execution{}, ContractReceipt{}, errors.New("contract code not found")
	}
	return runAVMCall(state.Policy, code, contract, call.Entrypoint, call.GasLimit, height, call.Body, call.EmitDestination)
}

func runAVMCall(policy ContractZonePolicy, code ContractCode, contract ContractInstance, entry avm.Entrypoint, gasLimit uint64, height uint64, body []byte, emitDestination sdk.AccAddress) (avm.Execution, ContractReceipt, error) {
	runner, err := avm.NewRunner(policy.Runtime.AVMParams)
	if err != nil {
		return avm.Execution{}, ContractReceipt{}, err
	}
	exec, err := runner.Run(code.AVMModule, StorageToAVM(contract.Storage), avm.RuntimeContext{
		Entry:			entry,
		ContractAddress:	contract.Address,
		Message: async.MessageEnvelope{
			Source:		contract.Owner,
			Destination:	contract.Address,
			Value:		sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
			Body:		append([]byte(nil), body...),
			GasLimit:	gasLimit,
		},
		BlockHeight:		height,
		GasLimit:		gasLimit,
		EmitDestination:	emitDestination,
	})
	if err != nil {
		return avm.Execution{}, ContractReceipt{}, err
	}
	for i := range exec.Outgoing {
		if len(exec.Outgoing[i].Source) == 0 {
			exec.Outgoing[i].Source = cloneAddress(contract.Address)
		}
		if exec.Outgoing[i].CreatedLogicalTime == 0 {
			exec.Outgoing[i].CreatedLogicalTime = height
		}
		if exec.Outgoing[i].GasLimit == 0 {
			exec.Outgoing[i].GasLimit = gasLimit
		}
	}
	totalGas, err := totalContractGas(exec, policy.GasModel)
	if err != nil {
		return avm.Execution{}, ContractReceipt{}, err
	}
	if totalGas > gasLimit {
		exec.ResultCode = async.ResultLimitExceeded
	}
	if len(exec.Outgoing) > int(policy.Limits.MaxEmittedMessages) {
		return avm.Execution{}, ContractReceipt{}, fmt.Errorf("contract emitted messages must be <= %d", policy.Limits.MaxEmittedMessages)
	}
	if uint32(len(exec.Outgoing)) > policy.Limits.MaxMessagesPerBlock {
		return avm.Execution{}, ContractReceipt{}, fmt.Errorf("contract messages per block must be <= %d", policy.Limits.MaxMessagesPerBlock)
	}
	if exec.ResultCode == async.ResultOK {
		storage := StorageFromAVM(ContractNamespace(contract.Address), exec.State)
		if err := ValidateStorageEntries(storage, policy.Limits); err != nil {
			return avm.Execution{}, ContractReceipt{}, err
		}
	}
	receipt := ContractReceipt{
		Contract:		cloneAddress(contract.Address),
		Entrypoint:		entry,
		ResultCode:		exec.ResultCode,
		GasUsed:		totalGas,
		StorageWrites:		exec.StorageWrites,
		EmittedMessages:	uint32(len(exec.Outgoing)),
		ReturnValue:		exec.ReturnValue,
	}
	return exec, receipt, nil
}

func totalContractGas(exec avm.Execution, model GasModel) (uint64, error) {
	total := exec.GasUsed
	var err error
	total, err = safeAddMul(total, uint64(exec.StorageWrites), model.StorageWriteGas)
	if err != nil {
		return 0, err
	}
	return safeAddMul(total, uint64(len(exec.Outgoing)), model.MessageForwardingGas)
}

func safeAddMul(base uint64, count uint64, price uint64) (uint64, error) {
	if count != 0 && price > (^uint64(0)-base)/count {
		return 0, errors.New("contract gas overflow")
	}
	return base + count*price, nil
}

func deriveContractAddress(codeID uint64, actor sdk.AccAddress, salt []byte) sdk.AccAddress {
	h := sha256.New()
	h.Write([]byte("aetra-contract-address-v1"))
	var id [8]byte
	binary.BigEndian.PutUint64(id[:], codeID)
	h.Write(id[:])
	h.Write(actor)
	h.Write(salt)
	sum := h.Sum(nil)
	return sdk.AccAddress(append([]byte(nil), sum[len(sum)-20:]...))
}
