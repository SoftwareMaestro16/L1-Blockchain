package avm

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

const testContractRaw = "4:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestStorageABIWriteReadDelete(t *testing.T) {
	abi := newTestStorageABI(t, DefaultStorageABIParams())
	exec := abi.BeginExecution()

	require.NoError(t, exec.SetStorage(testContractRaw, []byte("counter"), EncodeU64(7)))
	value, found, err := exec.GetStorage(testContractRaw, []byte("counter"))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(7), DecodeU64(value))

	require.NoError(t, exec.DeleteStorage(testContractRaw, []byte("counter")))
	value, found, err = exec.GetStorage(testContractRaw, []byte("counter"))
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, value)
}

func TestStorageABIRejectsOversizedKey(t *testing.T) {
	params := DefaultStorageABIParams()
	params.MaxKeyBytes = 4
	abi := newTestStorageABI(t, params)

	err := abi.SetStorage(testContractRaw, []byte("abcde"), []byte("value"))
	require.ErrorContains(t, err, "key must be <= 4 bytes")
}

func TestStorageABIRejectsOversizedValue(t *testing.T) {
	params := DefaultStorageABIParams()
	params.MaxValueBytes = 4
	abi := newTestStorageABI(t, params)

	err := abi.SetStorage(testContractRaw, []byte("key"), []byte("abcde"))
	require.ErrorContains(t, err, "value must be <= 4 bytes")
}

func TestStorageABITotalStorageLimit(t *testing.T) {
	params := DefaultStorageABIParams()
	params.MaxKeyBytes = 4
	params.MaxValueBytes = 8
	params.MaxContractStorageBytes = 12
	abi := newTestStorageABI(t, params)

	require.NoError(t, abi.SetStorage(testContractRaw, []byte("aa"), []byte("1234")))
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("bb"), []byte("1234")))
	err := abi.SetStorage(testContractRaw, []byte("cc"), []byte("1"))
	require.ErrorContains(t, err, "contract storage exceeds 12 bytes")
}

func TestStorageABILimitsPerExecution(t *testing.T) {
	t.Run("reads", func(t *testing.T) {
		params := DefaultStorageABIParams()
		params.MaxReadsPerExecution = 1
		abi := newTestStorageABI(t, params)
		require.NoError(t, abi.SetStorage(testContractRaw, []byte("a"), []byte("1")))
		exec := abi.BeginExecution()
		_, _, err := exec.GetStorage(testContractRaw, []byte("a"))
		require.NoError(t, err)
		_, _, err = exec.GetStorage(testContractRaw, []byte("a"))
		require.ErrorContains(t, err, "read limit exceeded")
	})

	t.Run("writes", func(t *testing.T) {
		params := DefaultStorageABIParams()
		params.MaxWritesPerExecution = 1
		abi := newTestStorageABI(t, params)
		exec := abi.BeginExecution()
		require.NoError(t, exec.SetStorage(testContractRaw, []byte("a"), []byte("1")))
		err := exec.SetStorage(testContractRaw, []byte("b"), []byte("2"))
		require.ErrorContains(t, err, "write limit exceeded")
	})

	t.Run("deletes", func(t *testing.T) {
		params := DefaultStorageABIParams()
		params.MaxDeletesPerExecution = 1
		abi := newTestStorageABI(t, params)
		exec := abi.BeginExecution()
		require.NoError(t, exec.DeleteStorage(testContractRaw, []byte("a")))
		err := exec.DeleteStorage(testContractRaw, []byte("b"))
		require.ErrorContains(t, err, "delete limit exceeded")
	})
}

func TestStorageABIIterationRequiresLimit(t *testing.T) {
	abi := newTestStorageABI(t, DefaultStorageABIParams())

	_, err := abi.IterateStorage(testContractRaw, nil, 0)
	require.ErrorContains(t, err, "iteration limit is required")
}

func TestStorageABIIterationOrderDeterministic(t *testing.T) {
	abi := newTestStorageABI(t, DefaultStorageABIParams())
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("b"), []byte("2")))
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("aa"), []byte("11")))
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("a"), []byte("1")))

	first, err := abi.IterateStorage(testContractRaw, nil, 10)
	require.NoError(t, err)
	second, err := abi.IterateStorage(testContractRaw, nil, 10)
	require.NoError(t, err)

	require.Equal(t, [][]byte{[]byte("a"), []byte("aa"), []byte("b")}, storageKeys(first))
	require.Equal(t, first, second)
}

func TestStorageABIStateRootChangesOnWrite(t *testing.T) {
	abi := newTestStorageABI(t, DefaultStorageABIParams())
	before, err := abi.ContractStateRoot(testContractRaw)
	require.NoError(t, err)
	globalBefore, err := abi.GlobalStateRoot()
	require.NoError(t, err)

	require.NoError(t, abi.SetStorage(testContractRaw, []byte("counter"), EncodeU64(1)))
	after, err := abi.ContractStateRoot(testContractRaw)
	require.NoError(t, err)
	globalAfter, err := abi.GlobalStateRoot()
	require.NoError(t, err)

	require.NotEqual(t, before, after)
	require.NotEqual(t, globalBefore, globalAfter)
	require.Equal(t, after, ComputeContractStateRoot(testContractRaw, []AVMStorageEntry{
		{Key: []byte("counter"), Value: EncodeU64(1)},
	}))
}

func TestStorageABIExportImportPreservesStorageAndRoot(t *testing.T) {
	params := DefaultStorageABIParams()
	abi := newTestStorageABI(t, params)
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("b"), []byte("2")))
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("a"), []byte("1")))

	exported, err := abi.ExportState()
	require.NoError(t, err)
	require.Len(t, exported.Contracts, 1)
	require.NotEmpty(t, exported.Contracts[0].Root)
	require.NotEmpty(t, exported.Root)

	imported, err := ImportAVMStorageState(params, exported)
	require.NoError(t, err)
	importedState, err := imported.ExportState()
	require.NoError(t, err)
	require.Equal(t, exported, importedState)

	contractRoot, err := imported.ContractStateRoot(testContractRaw)
	require.NoError(t, err)
	require.Equal(t, exported.Contracts[0].Root, contractRoot)
	entries, err := imported.IterateStorage(testContractRaw, nil, 10)
	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("a"), []byte("b")}, storageKeys(entries))
}

func TestStorageABIRawContractNamespaceRequired(t *testing.T) {
	abi := newTestStorageABI(t, DefaultStorageABIParams())

	err := abi.SetStorage("AE123", []byte("key"), []byte("value"))
	require.ErrorContains(t, err, "raw 4: address")
}

func TestStorageABIStateRootDeterministicForEntryOrder(t *testing.T) {
	entries := []AVMStorageEntry{
		{Key: []byte("b"), Value: []byte("2")},
		{Key: []byte("a"), Value: []byte("1")},
	}
	reversed := []AVMStorageEntry{
		{Key: []byte("a"), Value: []byte("1")},
		{Key: []byte("b"), Value: []byte("2")},
	}

	require.Equal(t, ComputeContractStateRoot(testContractRaw, entries), ComputeContractStateRoot(testContractRaw, reversed))
}

func TestStorageABIImportRejectsDuplicateContractNamespace(t *testing.T) {
	params := DefaultStorageABIParams()
	state := AVMStorageState{
		Contracts: []ContractStorageExport{
			{Contract: testContractRaw, Entries: []AVMStorageEntry{{Key: []byte("a"), Value: []byte("1")}}},
			{Contract: testContractRaw, Entries: []AVMStorageEntry{{Key: []byte("b"), Value: []byte("2")}}},
		},
	}

	_, err := ImportAVMStorageState(params, state)
	require.ErrorContains(t, err, "duplicate AVM storage contract namespace")
}

func newTestStorageABI(t *testing.T, params StorageABIParams) *StorageABI {
	t.Helper()
	abi, err := NewStorageABI(params, NewMapKVBackend())
	require.NoError(t, err)
	return abi
}

func storageKeys(entries []AVMStorageEntry) [][]byte {
	out := make([][]byte, len(entries))
	for i, entry := range entries {
		out[i] = append([]byte(nil), entry.Key...)
	}
	return out
}

func TestStorageABIClonesValues(t *testing.T) {
	abi := newTestStorageABI(t, DefaultStorageABIParams())
	value := []byte("value")
	require.NoError(t, abi.SetStorage(testContractRaw, []byte("key"), value))
	value[0] = 'V'

	stored, found, err := abi.GetStorage(testContractRaw, []byte("key"))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte("value"), stored)
	stored[0] = 'X'

	storedAgain, found, err := abi.GetStorage(testContractRaw, []byte("key"))
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, bytes.Equal([]byte("value"), storedAgain))
}
