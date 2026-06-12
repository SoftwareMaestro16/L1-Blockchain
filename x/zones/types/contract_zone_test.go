package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContractZoneBoundaryAndSpecStateKeys(t *testing.T) {
	boundary := DefaultContractZoneBoundary()
	require.NoError(t, boundary.Validate())
	require.Contains(t, boundary.Messages, ContractMessageStoreCode)
	require.Contains(t, boundary.Messages, ContractMessageInstantiate)
	require.Contains(t, boundary.Messages, ContractMessageExecute)
	require.Contains(t, boundary.Messages, ContractMessageMigrate)
	require.Contains(t, boundary.Messages, ContractMessageCallback)
	require.Contains(t, boundary.Messages, ContractMessageProofVerify)
	require.Contains(t, boundary.Messages, ContractMessageAsyncCall)
	require.Contains(t, boundary.ProofKinds, ContractProofCode)
	require.Contains(t, boundary.ProofKinds, ContractProofContract)
	require.Contains(t, boundary.ProofKinds, ContractProofState)
	require.Contains(t, boundary.ProofKinds, ContractProofABI)
	require.Contains(t, boundary.ProofKinds, ContractProofReceipt)
	require.Contains(t, boundary.ProofKinds, ContractProofGasTable)
	require.Contains(t, boundary.ProofKinds, ContractProofEvent)
	require.Contains(t, boundary.ProofKinds, ContractProofInbox)
	require.Contains(t, boundary.ProofKinds, ContractProofOutbox)

	codeKey, err := ContractCodeKey("code-1")
	require.NoError(t, err)
	require.Equal(t, "contract/code/code-1", codeKey)

	instanceKey, err := ContractInstanceKey("contract-1")
	require.NoError(t, err)
	require.Equal(t, "contract/instance/contract-1", instanceKey)

	storageKey, err := ContractStorageKey("contract-1", "balance")
	require.NoError(t, err)
	require.Equal(t, "contract/storage/contract-1/balance", storageKey)

	abiKey, err := ContractABIKey("code-1")
	require.NoError(t, err)
	require.Equal(t, "contract/abi/code-1", abiKey)

	inboxKey, err := ContractInboxKey("contract-1", "msg-1")
	require.NoError(t, err)
	require.Equal(t, "contract/inbox/contract-1/msg-1", inboxKey)

	receiptKey, err := ContractReceiptKey("contract-1", "receipt-1")
	require.NoError(t, err)
	require.Equal(t, "contract/receipts/contract-1/receipt-1", receiptKey)

	outboxKey, err := ContractOutboxKey("contract-1", "msg-1")
	require.NoError(t, err)
	require.Equal(t, "contract/outbox/contract-1/msg-1", outboxKey)

	eventKey, err := ContractEventKey("contract-1", "event-1")
	require.NoError(t, err)
	require.Equal(t, "contract/events/contract-1/event-1", eventKey)
}

func TestContractBytecodeAndCosmWasmBoundaryCommitments(t *testing.T) {
	iface, err := NewContractBytecodeInterface(ContractBytecodeInterface{
		Runtime:		ContractRuntimeAVM,
		InstructionSet:		"avm-v1",
		BytecodeHash:		hash("bytecode"),
		ABIHash:		hash("abi"),
		DeterminismHash:	hash("determinism"),
		MaxCodeBytes:		1 << 20,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeContractBytecodeInterfaceHash(iface), iface.InterfaceHash)
	require.NoError(t, iface.Validate())

	adapter, err := NewContractCosmWasmAdapterDescriptor(ContractCosmWasmAdapterDescriptor{
		AdapterID:	"cw-adapter",
		Version:	"v1",
		PolicyHash:	hash("adapter-policy"),
		CapabilityRoot:	hash("adapter-capability"),
	})
	require.NoError(t, err)
	require.Equal(t, ComputeContractCosmWasmAdapterHash(adapter), adapter.DescriptorHash)
	require.NoError(t, adapter.Validate())

	rootA := ComputeContractBytecodeInterfaceRoot([]ContractBytecodeInterface{iface})
	rootB := ComputeContractBytecodeInterfaceRoot([]ContractBytecodeInterface{iface})
	require.Equal(t, rootA, rootB)
	require.NotEmpty(t, ComputeContractCosmWasmAdapterRoot([]ContractCosmWasmAdapterDescriptor{adapter}))
}

func TestContractStorageInboxAndReceiptsAreDeterministic(t *testing.T) {
	storage, err := UpsertContractStorage(nil, contractStorage("contract-b", "slot-2", "value-2"))
	require.NoError(t, err)
	storage, err = UpsertContractStorage(storage, contractStorage("contract-a", "slot-1", "value-1"))
	require.NoError(t, err)
	storage, err = UpsertContractStorage(storage, contractStorage("contract-b", "slot-2", "value-2b"))
	require.NoError(t, err)
	require.Equal(t, "contract-a", storage[0].ContractAddr)
	require.Len(t, storage, 2)

	inbox, err := EnqueueContractInbox(nil, contractInbox("contract-1", "msg-2", 2))
	require.NoError(t, err)
	inbox, err = EnqueueContractInbox(inbox, contractInbox("contract-1", "msg-1", 1))
	require.NoError(t, err)
	require.Equal(t, "msg-1", inbox[0].MsgID)
	_, err = EnqueueContractInbox(inbox, contractInbox("contract-1", "msg-dup", 1))
	require.ErrorContains(t, err, "sequence already exists")

	receipt, err := NewContractExecutionReceipt(contractReceipt("contract-1", "receipt-1", "msg-1", 1))
	require.NoError(t, err)
	zoneReceipt, err := receipt.ZoneReceipt()
	require.NoError(t, err)
	require.Equal(t, ZoneIDContract, zoneReceipt.ZoneID)
	require.Equal(t, ZoneReceiptStatusSuccess, zoneReceipt.Status)

	rootA := ComputeContractStorageRoot(storage)
	rootB := ComputeContractStorageRoot([]ContractStorageEntry{storage[1], storage[0]})
	require.Equal(t, rootA, rootB)
	require.Equal(t, ComputeContractInboxRoot(inbox), ComputeContractInboxRoot([]ContractInboxMessage{inbox[1], inbox[0]}))
}

func TestContractAVMExecutionShardingAsyncCallsAndProofRoots(t *testing.T) {
	table, err := DefaultAVMGasTable(1)
	require.NoError(t, err)
	keeper, err := NewAVMExecutionKeeper("avm-2", table, 10_000, 16)
	require.NoError(t, err)
	require.NoError(t, keeper.ValidateHash())

	instance := ContractInstance{
		ContractAddr:	"contract-1",
		CodeID:		"code-1",
		Runtime:	ContractRuntimeAVM,
		Admin:		"alice",
		StorageRoot:	hash("storage"),
		CreatedHeight:	77,
		UpdatedHeight:	77,
	}
	msg := contractInbox("contract-1", "msg-1", 1)
	receipt, err := keeper.Execute(instance, msg, []AVMInstruction{
		{Opcode: "PUSH", OperandHash: hash("one")},
		{Opcode: "ADD", OperandHash: hash("two")},
		{Opcode: "RET", OperandHash: hash("ret")},
	}, 77, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(3), receipt.GasUsed)
	require.Equal(t, ZoneReceiptStatusSuccess, receipt.Status)

	instanceRoute, err := RouteContractInstanceShard("contract-1", 8, 2)
	require.NoError(t, err)
	storageRoute, err := RouteContractStorageShard("contract-1", "balance", 8, 2)
	require.NoError(t, err)
	codeRoute, err := RouteContractCodeShard("code-1", 8, 2)
	require.NoError(t, err)
	require.Equal(t, ContractRouteAddress, instanceRoute.RoutingMode)
	require.Equal(t, ContractRouteStorageKey, storageRoute.RoutingMode)
	require.Equal(t, ContractRouteCodeRegistry, codeRoute.RoutingMode)

	call, outboxMsg, err := NewContractAsyncCall(ContractAsyncCall{
		Source:		"workflow-1",
		SourceZone:	ZoneIDApplication,
		TargetContract:	"contract-1",
		PayloadHash:	hash("call-payload"),
		GasLimit:	1_000,
		RetryNonce:	2,
		ExpiryHeight:	100,
		CreatedHeight:	77,
	})
	require.NoError(t, err)
	require.NoError(t, call.ValidateHash())
	require.Equal(t, ContractMessageAsyncCall, outboxMsg.MessageKind)
	outbox, err := EnqueueContractOutbox(nil, outboxMsg)
	require.NoError(t, err)
	require.Len(t, outbox, 1)

	event := ContractEvent{ContractAddr: "contract-1", EventID: "event-1", Topic: "transfer", PayloadHash: hash("event"), Height: 77, Sequence: 1}
	event.EventHash = ComputeContractEventHash(event)
	require.NoError(t, event.ValidateHash())

	roots := ContractZoneRoots{
		Height:			77,
		CodeRoot:		hash("code-root"),
		InstanceRoot:		hash("instance-root"),
		StorageRoot:		hash("storage-root"),
		ABIRoot:		hash("abi-root"),
		InboxRoot:		ComputeContractInboxRoot([]ContractInboxMessage{msg}),
		OutboxRoot:		ComputeContractOutboxRoot(outbox),
		ReceiptRoot:		ComputeContractReceiptRoot([]ContractExecutionReceipt{receipt}),
		GasTableRoot:		ComputeAVMGasTableRoot(table),
		EventRoot:		ComputeContractEventRoot([]ContractEvent{event}),
		BytecodeInterfaceRoot:	hash("iface-root"),
		CosmWasmAdapterRoot:	hash("adapter-root"),
		ExecutionRoot:		ComputeContractExecutionRoot([]ContractExecutionReceipt{receipt}),
		ProofRoot:		hash("proof-root"),
	}
	exports, err := BuildContractProofRootExports(77, roots)
	require.NoError(t, err)
	require.Len(t, exports, 8)
	require.True(t, hasContractProofRoot(exports, ContractProofRootGasTable, roots.GasTableRoot))
	require.True(t, hasContractProofRoot(exports, ContractProofRootOutbox, roots.OutboxRoot))
}

func TestContractZoneStateRootAndProofRequest(t *testing.T) {
	iface, err := NewContractBytecodeInterface(ContractBytecodeInterface{
		Runtime:		ContractRuntimeAVM,
		InstructionSet:		"avm-v1",
		BytecodeHash:		hash("bytecode"),
		ABIHash:		hash("abi"),
		DeterminismHash:	hash("determinism"),
		MaxCodeBytes:		1 << 20,
	})
	require.NoError(t, err)
	adapter, err := NewContractCosmWasmAdapterDescriptor(ContractCosmWasmAdapterDescriptor{
		AdapterID:	"cw-adapter",
		Version:	"v1",
		PolicyHash:	hash("adapter-policy"),
		CapabilityRoot:	hash("adapter-capability"),
	})
	require.NoError(t, err)
	receipt, err := NewContractExecutionReceipt(contractReceipt("contract-1", "receipt-1", "msg-1", 1))
	require.NoError(t, err)
	gasTable, err := DefaultAVMGasTable(1)
	require.NoError(t, err)
	event := ContractEvent{ContractAddr: "contract-1", EventID: "event-1", Topic: "execute", PayloadHash: hash("event"), Height: 77, Sequence: 1}
	event.EventHash = ComputeContractEventHash(event)

	state := ContractZoneState{
		Height:	77,
		Codes: []ContractCode{{
			CodeID:		"code-1",
			Runtime:	ContractRuntimeAVM,
			BytecodeHash:	hash("bytecode"),
			BytecodeSize:	4096,
			ABIHash:	hash("abi"),
			InterfaceHash:	iface.InterfaceHash,
			Uploader:	"alice",
			UploadedHeight:	77,
		}},
		Instances: []ContractInstance{{
			ContractAddr:	"contract-1",
			CodeID:		"code-1",
			Runtime:	ContractRuntimeAVM,
			Admin:		"alice",
			StorageRoot:	hash("storage"),
			CreatedHeight:	77,
			UpdatedHeight:	77,
		}},
		Storage:		[]ContractStorageEntry{contractStorage("contract-1", "slot-1", "value-1")},
		ABIs:			[]ContractABIDescriptor{{CodeID: "code-1", ABIHash: hash("abi"), InterfaceHash: iface.InterfaceHash, ExportedMethods: []string{"execute", "query"}, RegisteredHeight: 77}},
		Inbox:			[]ContractInboxMessage{contractInbox("contract-1", "msg-1", 1)},
		Outbox:			[]ContractInboxMessage{contractInbox("contract-1", "msg-2", 2)},
		Receipts:		[]ContractExecutionReceipt{receipt},
		BytecodeInterfaces:	[]ContractBytecodeInterface{iface},
		CosmWasmAdapters:	[]ContractCosmWasmAdapterDescriptor{adapter},
		GasTable:		gasTable,
		Events:			[]ContractEvent{event},
	}
	require.NoError(t, state.Validate())

	root, err := BuildContractZoneRootFromState(77, state, hash("proofs"))
	require.NoError(t, err)
	require.Equal(t, ZoneIDContract, root.ZoneID)
	require.Equal(t, ComputeContractZoneStateRoot(state), root.ZoneStateRoot)

	req, err := ContractProofRequest(ContractProofState, "contract-1/slot-1", 77, root.RootHash, 4)
	require.NoError(t, err)
	require.Equal(t, "QueryContractState/contract-1/slot-1", req.Key)
}

func hasContractProofRoot(exports []ContractProofRootExport, rootType ContractProofRootType, rootHash string) bool {
	for _, export := range exports {
		if export.RootType == rootType && export.RootHash == rootHash {
			return true
		}
	}
	return false
}

func contractStorage(contractAddr, key, value string) ContractStorageEntry {
	return ContractStorageEntry{
		ContractAddr:	contractAddr,
		Key:		key,
		ValueHash:	hash(value),
	}
}

func contractInbox(contractAddr, msgID string, sequence uint64) ContractInboxMessage {
	return ContractInboxMessage{
		ContractAddr:	contractAddr,
		MsgID:		msgID,
		MessageKind:	ContractMessageExecute,
		Source:		"application-zone",
		PayloadHash:	hash(msgID),
		GasLimit:	1_000,
		Sequence:	sequence,
		ReceivedHeight:	77,
	}
}

func contractReceipt(contractAddr, receiptID, msgID string, sequence uint64) ContractExecutionReceipt {
	return ContractExecutionReceipt{
		ContractAddr:		contractAddr,
		ReceiptID:		receiptID,
		MsgID:			msgID,
		MessageKind:		ContractMessageExecute,
		Status:			ZoneReceiptStatusSuccess,
		GasUsed:		500,
		OutputHash:		hash(receiptID + "-output"),
		StorageRoot:		hash(receiptID + "-storage"),
		EmittedMessages:	1,
		Height:			77,
		Sequence:		sequence,
	}
}
