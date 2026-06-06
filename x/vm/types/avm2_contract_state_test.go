package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestAVM2ContractStateKeysAndLayout(t *testing.T) {
	require.Equal(t, "contract/code/00000000000000000007", AVM2ContractCodeStateKey(7))
	require.Equal(t, "contract/instance/contract-a", AVM2ContractInstanceStateKey("contract-a"))
	require.Equal(t, "contract/storage/contract-a/balance/main", AVM2ContractStorageStateKey("contract-a", "balance/main"))
	require.Equal(t, "contract/abi/00000000000000000007/00000000000000000002", AVM2ContractABIStateKey(7, 2))
	require.Equal(t, "contract/events/00000000000000000010/contract-a/event-1", AVM2ContractEventStateKey(10, "contract-a", "event-1"))
	require.Equal(t, "contract/message_nonce/contract-a", AVM2ContractMessageNonceStateKey("contract-a"))
	require.Equal(t, "code_id/00000000000000000007", AVM2ContractCodeShardKey(7))
	require.Equal(t, "contract-a/balance", AVM2ContractStorageShardKey("contract-a", "balance/main"))

	layout, err := DefaultAVM2ContractStateLayout()
	require.NoError(t, err)
	require.Len(t, layout.Entries, 6)
	require.Equal(t, ComputeAVM2ContractStateLayoutRoot(layout), layout.LayoutRoot)

	mutated := layout
	mutated.Entries[0].ShardKey = "wrong"
	mutated.Entries[0].EntryHash = ComputeAVM2ContractStateEntryHash(mutated.Entries[0])
	mutated.LayoutRoot = ComputeAVM2ContractStateLayoutRoot(mutated)
	require.NotEqual(t, layout.LayoutRoot, mutated.LayoutRoot)
}

func TestAVM2CodeAndContractRecordsValidateConsensusFields(t *testing.T) {
	code := mustAVM2CodeRecord(t, 7, true)
	contract := mustAVM2ContractRecord(t, code.CodeID, "contract-a")

	require.NoError(t, ValidateAVM2ContractInstantiation(code, contract))
	require.Equal(t, ComputeAVM2CodeRecordHash(code), code.RecordHash)
	require.Equal(t, ComputeAVM2ContractRecordHash(contract), contract.RecordHash)

	disabled := code
	disabled.Enabled = false
	disabled.RecordHash = ComputeAVM2CodeRecordHash(disabled)
	require.ErrorContains(t, ValidateAVM2ContractInstantiation(disabled, contract), "disabled code")

	negative := contract
	negative.BalanceNAET = sdkmath.NewInt(-1)
	negative.RecordHash = ComputeAVM2ContractRecordHash(negative)
	require.ErrorContains(t, negative.Validate(), "cannot be negative")

	staleHeight := contract
	staleHeight.UpdatedAtHeight = staleHeight.CreatedAtHeight - 1
	staleHeight.RecordHash = ComputeAVM2ContractRecordHash(staleHeight)
	require.ErrorContains(t, staleHeight.Validate(), "heights")
}

func TestAVM2ContractStateIndexRootsAreCanonical(t *testing.T) {
	codeA := mustAVM2CodeRecord(t, 1, true)
	codeB := mustAVM2CodeRecord(t, 2, true)
	contractA := mustAVM2ContractRecord(t, codeA.CodeID, "contract-a")
	contractB := mustAVM2ContractRecord(t, codeB.CodeID, "contract-b")
	storage := mustAVM2StorageValue(t, "contract-a", "balance/main")
	abi := mustAVM2ABIRecord(t, codeA.CodeID)
	event := mustAVM2EventRecord(t, "contract-a")
	nonce := mustAVM2NonceRecord(t, "contract-a", 9)

	index, err := NewAVM2ContractStateIndex(AVM2ContractStateIndex{
		Codes:         []AVM2CodeRecord{codeB, codeA},
		Contracts:     []AVM2ContractRecord{contractB, contractA},
		Storage:       []AVM2ContractStorageValue{storage},
		ABIs:          []AVM2ContractABIDescriptorRecord{abi},
		Events:        []AVM2ContractEventRecord{event},
		MessageNonces: []AVM2ContractMessageNonceRecord{nonce},
	})
	require.NoError(t, err)
	require.NoError(t, index.Validate())
	require.Equal(t, ComputeAVM2ContractStateIndexRoot(index), index.StateRoot)

	reordered, err := NewAVM2ContractStateIndex(AVM2ContractStateIndex{
		Contracts:     []AVM2ContractRecord{contractA, contractB},
		Codes:         []AVM2CodeRecord{codeA, codeB},
		MessageNonces: []AVM2ContractMessageNonceRecord{nonce},
		Events:        []AVM2ContractEventRecord{event},
		ABIs:          []AVM2ContractABIDescriptorRecord{abi},
		Storage:       []AVM2ContractStorageValue{storage},
	})
	require.NoError(t, err)
	require.Equal(t, index.StateRoot, reordered.StateRoot)

	missingCode := index
	missingCode.Contracts[0].CodeID = 99
	missingCode.Contracts[0].RecordHash = ComputeAVM2ContractRecordHash(missingCode.Contracts[0])
	missingCode.StateRoot = ComputeAVM2ContractStateIndexRoot(missingCode)
	require.ErrorContains(t, missingCode.Validate(), "missing code")
}

func TestAVM2ContractStateChildRecordsValidateKeys(t *testing.T) {
	storage := mustAVM2StorageValue(t, "contract-a", "balance/main")
	require.Equal(t, "contract-a/balance", storage.ShardKey)

	badStorage := storage
	badStorage.ShardKey = "contract-a/other"
	badStorage.RecordHash = ComputeAVM2ContractStorageValueHash(badStorage)
	require.ErrorContains(t, badStorage.Validate(), "shard key")

	abi := mustAVM2ABIRecord(t, 4)
	require.Equal(t, AVM2ContractABIStateKey(4, 1), abi.Key)
	require.Equal(t, AVM2ContractCodeShardKey(4), abi.ShardKey)

	event := mustAVM2EventRecord(t, "contract-a")
	require.Equal(t, AVM2ContractEventStateKey(event.Event.Height, "contract-a", event.Event.EventID), event.Key)

	nonce := mustAVM2NonceRecord(t, "contract-a", 10)
	require.Equal(t, AVM2ContractMessageNonceStateKey("contract-a"), nonce.Key)
	require.Equal(t, uint64(10), nonce.Nonce)
}

func mustAVM2CodeRecord(t *testing.T, codeID uint64, enabled bool) AVM2CodeRecord {
	t.Helper()
	abi := mustAVM2ABIForCode(t, codeID)
	record, err := NewAVM2CodeRecord(AVM2CodeRecord{
		CodeID:                codeID,
		CodeHash:              ComputeAVM2BytesHash([]byte(AVM2ContractCodeStateKey(codeID))),
		VMVersion:             AVM2VMVersion,
		InstructionSetVersion: AVM2DefaultInstructionSet,
		ABIHash:               abi.InterfaceHash,
		Deployer:              "deployer-a",
		CreatedAtHeight:       10,
		CodeBytesRef:          "store:v2/code/" + AVM2ContractCodeShardKey(codeID),
		MeteringProfile:       AVM2MeteringProfileDefault,
		Enabled:               enabled,
	})
	require.NoError(t, err)
	return record
}

func mustAVM2ContractRecord(t *testing.T, codeID uint64, contractAddr string) AVM2ContractRecord {
	t.Helper()
	record, err := NewAVM2ContractRecord(AVM2ContractRecord{
		ContractAddr:    contractAddr,
		CodeID:          codeID,
		Creator:         "creator-a",
		AdminOptional:   "admin-a",
		StorageRoot:     ComputeAVM2BytesHash([]byte("storage-root-" + contractAddr)),
		BalanceNAET:     sdkmath.NewInt(100),
		CreatedAtHeight: 11,
		UpdatedAtHeight: 12,
		InstanceVersion: 1,
		ShardID:         3,
	})
	require.NoError(t, err)
	return record
}

func mustAVM2StorageValue(t *testing.T, contractAddr, key string) AVM2ContractStorageValue {
	t.Helper()
	value, err := NewAVM2ContractStorageValue(AVM2ContractStorageValue{
		ContractAddr: contractAddr,
		StorageKey:   key,
		ValueHash:    ComputeAVM2BytesHash([]byte("value-" + key)),
		ValueBytes:   32,
	})
	require.NoError(t, err)
	return value
}

func mustAVM2ABIRecord(t *testing.T, codeID uint64) AVM2ContractABIDescriptorRecord {
	t.Helper()
	record, err := NewAVM2ContractABIDescriptorRecord(AVM2ContractABIDescriptorRecord{
		ABI: mustAVM2ABIForCode(t, codeID),
	})
	require.NoError(t, err)
	return record
}

func mustAVM2ABIForCode(t *testing.T, codeID uint64) AVM2ABIDescriptor {
	t.Helper()
	abi := AVM2ABIDescriptor{
		ABIVersion:    1,
		CodeID:        codeID,
		Methods:       []string{"execute", "query"},
		Events:        []string{"updated"},
		Errors:        []string{"not_authorized"},
		RequiredFunds: []string{"naet"},
		GasHints:      []string{"execute/100"},
	}
	abi.InterfaceHash = ComputeAVM2ABIInterfaceHash(abi)
	require.NoError(t, abi.Validate(DefaultAVM2Limits()))
	return abi
}

func mustAVM2EventRecord(t *testing.T, contractAddr string) AVM2ContractEventRecord {
	t.Helper()
	event := AVM2Event{
		Height:          12,
		ContractAddress: contractAddr,
		EventID:         "event-1",
		Name:            "updated",
		PayloadHash:     ComputeAVM2BytesHash([]byte("payload")),
	}
	event.EventHash = ComputeAVM2EventHash(event)
	record, err := NewAVM2ContractEventRecord(AVM2ContractEventRecord{Event: event})
	require.NoError(t, err)
	return record
}

func mustAVM2NonceRecord(t *testing.T, contractAddr string, nonce uint64) AVM2ContractMessageNonceRecord {
	t.Helper()
	record, err := NewAVM2ContractMessageNonceRecord(AVM2ContractMessageNonceRecord{
		ContractAddr: contractAddr,
		Nonce:        nonce,
	})
	require.NoError(t, err)
	return record
}
