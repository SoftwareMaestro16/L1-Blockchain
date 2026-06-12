package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAVMContractStateKeysAndLayout(t *testing.T) {
	require.Equal(t, "contract/code/00000000000000000007", AVMContractCodeStateKey(7))
	require.Equal(t, "contract/instance/contract-a", AVMContractInstanceStateKey("contract-a"))
	require.Equal(t, "contract/storage/contract-a/balance/main", AVMContractStorageStateKey("contract-a", "balance/main"))
	require.Equal(t, "contract/abi/00000000000000000007/00000000000000000002", AVMContractABIStateKey(7, 2))
	require.Equal(t, "contract/events/00000000000000000010/contract-a/event-1", AVMContractEventStateKey(10, "contract-a", "event-1"))
	require.Equal(t, "contract/message_nonce/contract-a", AVMContractMessageNonceStateKey("contract-a"))
	require.Equal(t, "code_id/00000000000000000007", AVMContractCodeShardKey(7))
	require.Equal(t, "contract-a/balance", AVMContractStorageShardKey("contract-a", "balance/main"))

	layout, err := DefaultAVMContractStateLayout()
	require.NoError(t, err)
	require.Len(t, layout.Entries, 6)
	require.Equal(t, ComputeAVMContractStateLayoutRoot(layout), layout.LayoutRoot)

	mutated := layout
	mutated.Entries[0].ShardKey = "wrong"
	mutated.Entries[0].EntryHash = ComputeAVMContractStateEntryHash(mutated.Entries[0])
	mutated.LayoutRoot = ComputeAVMContractStateLayoutRoot(mutated)
	require.NotEqual(t, layout.LayoutRoot, mutated.LayoutRoot)
}

func TestAVMCodeAndContractRecordsValidateConsensusFields(t *testing.T) {
	code := mustAVMCodeRecord(t, 7, true)
	contract := mustAVMContractRecord(t, code.CodeID, "contract-a")

	require.NoError(t, ValidateAVMContractInstantiation(code, contract))
	require.Equal(t, ComputeAVMCodeRecordHash(code), code.RecordHash)
	require.Equal(t, ComputeAVMContractRecordHash(contract), contract.RecordHash)

	disabled := code
	disabled.Enabled = false
	disabled.RecordHash = ComputeAVMCodeRecordHash(disabled)
	require.ErrorContains(t, ValidateAVMContractInstantiation(disabled, contract), "disabled code")

	negative := contract
	negative.BalanceNAET = sdkmath.NewInt(-1)
	negative.RecordHash = ComputeAVMContractRecordHash(negative)
	require.ErrorContains(t, negative.Validate(), "cannot be negative")

	staleHeight := contract
	staleHeight.UpdatedAtHeight = staleHeight.CreatedAtHeight - 1
	staleHeight.RecordHash = ComputeAVMContractRecordHash(staleHeight)
	require.ErrorContains(t, staleHeight.Validate(), "heights")
}

func TestAVMContractStateIndexRootsAreCanonical(t *testing.T) {
	codeA := mustAVMCodeRecord(t, 1, true)
	codeB := mustAVMCodeRecord(t, 2, true)
	contractA := mustAVMContractRecord(t, codeA.CodeID, "contract-a")
	contractB := mustAVMContractRecord(t, codeB.CodeID, "contract-b")
	storage := mustAVMStorageValue(t, "contract-a", "balance/main")
	abi := mustAVMABIRecord(t, codeA.CodeID)
	event := mustAVMEventRecord(t, "contract-a")
	nonce := mustAVMNonceRecord(t, "contract-a", 9)

	index, err := NewAVMContractStateIndex(AVMContractStateIndex{
		Codes:		[]AVMCodeRecord{codeB, codeA},
		Contracts:	[]AVMContractRecord{contractB, contractA},
		Storage:	[]AVMContractStorageValue{storage},
		ABIs:		[]AVMContractABIDescriptorRecord{abi},
		Events:		[]AVMContractEventRecord{event},
		MessageNonces:	[]AVMContractMessageNonceRecord{nonce},
	})
	require.NoError(t, err)
	require.NoError(t, index.Validate())
	require.Equal(t, ComputeAVMContractStateIndexRoot(index), index.StateRoot)

	reordered, err := NewAVMContractStateIndex(AVMContractStateIndex{
		Contracts:	[]AVMContractRecord{contractA, contractB},
		Codes:		[]AVMCodeRecord{codeA, codeB},
		MessageNonces:	[]AVMContractMessageNonceRecord{nonce},
		Events:		[]AVMContractEventRecord{event},
		ABIs:		[]AVMContractABIDescriptorRecord{abi},
		Storage:	[]AVMContractStorageValue{storage},
	})
	require.NoError(t, err)
	require.Equal(t, index.StateRoot, reordered.StateRoot)

	missingCode := index
	missingCode.Contracts[0].CodeID = 99
	missingCode.Contracts[0].RecordHash = ComputeAVMContractRecordHash(missingCode.Contracts[0])
	missingCode.StateRoot = ComputeAVMContractStateIndexRoot(missingCode)
	require.ErrorContains(t, missingCode.Validate(), "missing code")
}

func TestAVMContractStateChildRecordsValidateKeys(t *testing.T) {
	storage := mustAVMStorageValue(t, "contract-a", "balance/main")
	require.Equal(t, "contract-a/balance", storage.ShardKey)

	badStorage := storage
	badStorage.ShardKey = "contract-a/other"
	badStorage.RecordHash = ComputeAVMContractStorageValueHash(badStorage)
	require.ErrorContains(t, badStorage.Validate(), "shard key")

	abi := mustAVMABIRecord(t, 4)
	require.Equal(t, AVMContractABIStateKey(4, 1), abi.Key)
	require.Equal(t, AVMContractCodeShardKey(4), abi.ShardKey)

	event := mustAVMEventRecord(t, "contract-a")
	require.Equal(t, AVMContractEventStateKey(event.Event.Height, "contract-a", event.Event.EventID), event.Key)

	nonce := mustAVMNonceRecord(t, "contract-a", 10)
	require.Equal(t, AVMContractMessageNonceStateKey("contract-a"), nonce.Key)
	require.Equal(t, uint64(10), nonce.Nonce)
}

func mustAVMCodeRecord(t *testing.T, codeID uint64, enabled bool) AVMCodeRecord {
	t.Helper()
	abi := mustAVMABIForCode(t, codeID)
	record, err := NewAVMCodeRecord(AVMCodeRecord{
		CodeID:			codeID,
		CodeHash:		ComputeAVMBytesHash([]byte(AVMContractCodeStateKey(codeID))),
		VMVersion:		AVMVMVersion,
		InstructionSetVersion:	AVMDefaultInstructionSet,
		ABIHash:		abi.InterfaceHash,
		Deployer:		"deployer-a",
		CreatedAtHeight:	10,
		CodeBytesRef:		"store:v2/code/" + AVMContractCodeShardKey(codeID),
		MeteringProfile:	AVMMeteringProfileDefault,
		Enabled:		enabled,
	})
	require.NoError(t, err)
	return record
}

func mustAVMContractRecord(t *testing.T, codeID uint64, contractAddr string) AVMContractRecord {
	t.Helper()
	record, err := NewAVMContractRecord(AVMContractRecord{
		ContractAddr:		contractAddr,
		CodeID:			codeID,
		Creator:		"creator-a",
		AdminOptional:		"admin-a",
		StorageRoot:		ComputeAVMBytesHash([]byte("storage-root-" + contractAddr)),
		BalanceNAET:		sdkmath.NewInt(100),
		CreatedAtHeight:	11,
		UpdatedAtHeight:	12,
		InstanceVersion:	1,
		ShardID:		3,
	})
	require.NoError(t, err)
	return record
}

func mustAVMStorageValue(t *testing.T, contractAddr, key string) AVMContractStorageValue {
	t.Helper()
	value, err := NewAVMContractStorageValue(AVMContractStorageValue{
		ContractAddr:	contractAddr,
		StorageKey:	key,
		ValueHash:	ComputeAVMBytesHash([]byte("value-" + key)),
		ValueBytes:	32,
	})
	require.NoError(t, err)
	return value
}

func mustAVMABIRecord(t *testing.T, codeID uint64) AVMContractABIDescriptorRecord {
	t.Helper()
	record, err := NewAVMContractABIDescriptorRecord(AVMContractABIDescriptorRecord{
		ABI: mustAVMABIForCode(t, codeID),
	})
	require.NoError(t, err)
	return record
}

func mustAVMABIForCode(t *testing.T, codeID uint64) AVMABIDescriptor {
	t.Helper()
	abi := AVMABIDescriptor{
		ABIVersion:	1,
		CodeID:		codeID,
		Methods:	[]string{"execute", "query"},
		Events:		[]string{"updated"},
		Errors:		[]string{"not_authorized"},
		RequiredFunds:	[]string{"naet"},
		GasHints:	[]string{"execute/100"},
	}
	abi.InterfaceHash = ComputeAVMABIInterfaceHash(abi)
	require.NoError(t, abi.Validate(DefaultAVMLimits()))
	return abi
}

func mustAVMEventRecord(t *testing.T, contractAddr string) AVMContractEventRecord {
	t.Helper()
	event := AVMEvent{
		Height:			12,
		ContractAddress:	contractAddr,
		EventID:		"event-1",
		Name:			"updated",
		PayloadHash:		ComputeAVMBytesHash([]byte("payload")),
	}
	event.EventHash = ComputeAVMEventHash(event)
	record, err := NewAVMContractEventRecord(AVMContractEventRecord{Event: event})
	require.NoError(t, err)
	return record
}

func mustAVMNonceRecord(t *testing.T, contractAddr string, nonce uint64) AVMContractMessageNonceRecord {
	t.Helper()
	record, err := NewAVMContractMessageNonceRecord(AVMContractMessageNonceRecord{
		ContractAddr:	contractAddr,
		Nonce:		nonce,
	})
	require.NoError(t, err)
	return record
}
