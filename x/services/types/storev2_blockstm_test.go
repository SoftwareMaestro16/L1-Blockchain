package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceStoreV2LayoutCoversRequiredPrefixesAndPerformanceRules(t *testing.T) {
	layout, err := DefaultServiceStoreV2Layout()
	require.NoError(t, err)
	require.NoError(t, layout.Validate())
	require.NotEmpty(t, layout.LayoutHash)

	descriptor := requireStoreV2Entry(t, layout, ServiceStoreV2DescriptorPrefix)
	require.Equal(t, ServiceStoreV2LookupPrimaryRead, descriptor.LookupKind)
	require.Equal(t, "service_id", descriptor.PartitionKey)
	require.Equal(t, uint32(1), descriptor.PrimaryReadBound)

	iface := requireStoreV2Entry(t, layout, ServiceStoreV2InterfacePrefix)
	require.Equal(t, ServiceStoreV2LookupPrimaryRead, iface.LookupKind)
	require.Equal(t, "interface_hash", iface.PartitionKey)
	require.Equal(t, uint32(1), iface.PrimaryReadBound)
	require.Equal(t, ServiceStoreV2LargeValueHashCommittedOnNeed, iface.LargeValuePolicy)

	for _, prefix := range []string{ServiceStoreV2OwnerIndexPrefix, ServiceStoreV2IdentityIndexPrefix, ServiceStoreV2ProviderIndexPrefix, ServiceStoreV2MethodIndexPrefix} {
		entry := requireStoreV2Entry(t, layout, prefix)
		require.True(t, entry.PrefixQueryable, prefix)
		require.Equal(t, ServiceStoreV2LookupPrefixQuery, entry.LookupKind)
	}

	receipt := requireStoreV2Entry(t, layout, ServiceStoreV2ReceiptPrefix)
	require.True(t, receipt.HeightIndexed)
	receiptHeight := requireStoreV2Entry(t, layout, ServiceStoreV2ReceiptHeightPrefix)
	require.True(t, receiptHeight.HeightIndexed)
	require.True(t, receiptHeight.PrefixQueryable)
}

func TestServiceStoreV2LayoutRejectsMissingRequiredPrefix(t *testing.T) {
	layout, err := DefaultServiceStoreV2Layout()
	require.NoError(t, err)
	layout.Entries = removeStoreV2Entry(layout.Entries, ServiceStoreV2PaymentPrefix)
	layout.LayoutHash = ComputeServiceStoreV2LayoutHash(layout)
	require.ErrorContains(t, layout.Validate(), ServiceStoreV2PaymentPrefix)
}

func TestServiceStoreV2LayoutRejectsSlowPrimaryLookupAndUnindexedReceipts(t *testing.T) {
	layout, err := DefaultServiceStoreV2Layout()
	require.NoError(t, err)
	for i := range layout.Entries {
		if layout.Entries[i].Prefix == ServiceStoreV2DescriptorPrefix {
			layout.Entries[i].PrimaryReadBound = 2
			layout.Entries[i].EntryHash = ComputeServiceStoreV2EntryHash(layout.Entries[i])
		}
	}
	layout.LayoutHash = ComputeServiceStoreV2LayoutHash(layout)
	require.ErrorContains(t, layout.Validate(), "one primary read")

	layout, err = DefaultServiceStoreV2Layout()
	require.NoError(t, err)
	for i := range layout.Entries {
		if layout.Entries[i].Prefix == ServiceStoreV2ReceiptHeightPrefix {
			layout.Entries[i].HeightIndexed = false
			layout.Entries[i].EntryHash = ComputeServiceStoreV2EntryHash(layout.Entries[i])
		}
	}
	layout.LayoutHash = ComputeServiceStoreV2LayoutHash(layout)
	require.ErrorContains(t, layout.Validate(), "receipt height index")
}

func TestDefaultServiceBlockSTMStrategyCoversPartitionRules(t *testing.T) {
	strategy, err := DefaultServiceBlockSTMStrategy()
	require.NoError(t, err)
	require.NoError(t, strategy.Validate())
	require.NotEmpty(t, strategy.StrategyHash)

	for _, family := range []string{"descriptor", "receipt", "provider", "payment", "service_local_state", "interface"} {
		rule, found := strategy.ruleByFamily(family)
		require.True(t, found, family)
		require.NoError(t, rule.Validate())
	}

	payment, found := strategy.ruleByFamily("payment")
	require.True(t, found)
	require.Equal(t, []string{"escrow_id", "stream_id"}, payment.PartitionBy)
}

func TestServiceBlockSTMExpectedVersionRequiredForDescriptorUpdates(t *testing.T) {
	_, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateService, map[string]string{"service_id": "svc.dex"}, 0)
	require.ErrorContains(t, err, "expected version")

	op, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateService, map[string]string{"service_id": "svc.dex"}, 7)
	require.NoError(t, err)
	require.Equal(t, uint64(7), op.ExpectedDescriptorVersion)
	require.Contains(t, op.WritePartitions, "descriptor/svc.dex")
}

func TestServiceBlockSTMParallelSafeOperationsUseDisjointPartitions(t *testing.T) {
	registerLeft, err := NewServiceBlockSTMOperation(ServiceBlockSTMRegisterService, map[string]string{"service_id": "svc.dex"}, 0)
	require.NoError(t, err)
	registerRight, err := NewServiceBlockSTMOperation(ServiceBlockSTMRegisterService, map[string]string{"service_id": "svc.storage"}, 0)
	require.NoError(t, err)
	require.False(t, ServiceBlockSTMOperationsConflict(registerLeft, registerRight))

	receiptLeft, err := NewServiceBlockSTMOperation(ServiceBlockSTMAnchorReceipt, map[string]string{"call_id": "call.alpha"}, 0)
	require.NoError(t, err)
	receiptRight, err := NewServiceBlockSTMOperation(ServiceBlockSTMAnchorReceipt, map[string]string{"call_id": "call.beta"}, 0)
	require.NoError(t, err)
	require.False(t, ServiceBlockSTMOperationsConflict(receiptLeft, receiptRight))

	providerLeft, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateProvider, map[string]string{"provider_id": "provider.a"}, 0)
	require.NoError(t, err)
	providerRight, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateProvider, map[string]string{"provider_id": "provider.b"}, 0)
	require.NoError(t, err)
	require.False(t, ServiceBlockSTMOperationsConflict(providerLeft, providerRight))

	paymentLeft, err := NewServiceBlockSTMOperation(ServiceBlockSTMSettlePayment, map[string]string{"escrow_id": "escrow.a"}, 0)
	require.NoError(t, err)
	paymentRight, err := NewServiceBlockSTMOperation(ServiceBlockSTMSettlePayment, map[string]string{"stream_id": "stream.b"}, 0)
	require.NoError(t, err)
	require.False(t, ServiceBlockSTMOperationsConflict(paymentLeft, paymentRight))
}

func TestServiceBlockSTMDetectsConflictProneOperations(t *testing.T) {
	updateLeft, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateService, map[string]string{"service_id": "svc.dex"}, 3)
	require.NoError(t, err)
	updateRight, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateService, map[string]string{"service_id": "svc.dex"}, 3)
	require.NoError(t, err)
	require.True(t, ServiceBlockSTMOperationsConflict(updateLeft, updateRight))

	callLeft, err := NewServiceBlockSTMOperation(ServiceBlockSTMExecuteOnChainCall, map[string]string{"call_id": "call.a", "service_id": "svc.dex", "state_key": "pool/atom"}, 0)
	require.NoError(t, err)
	callRight, err := NewServiceBlockSTMOperation(ServiceBlockSTMExecuteOnChainCall, map[string]string{"call_id": "call.b", "service_id": "svc.dex", "state_key": "pool/atom"}, 0)
	require.NoError(t, err)
	require.True(t, ServiceBlockSTMOperationsConflict(callLeft, callRight))

	providerUpdate, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateProvider, map[string]string{"provider_id": "provider.a"}, 0)
	require.NoError(t, err)
	providerSlash, err := NewServiceBlockSTMOperation(ServiceBlockSTMSlashProvider, map[string]string{"provider_id": "provider.a"}, 0)
	require.NoError(t, err)
	require.True(t, ServiceBlockSTMOperationsConflict(providerUpdate, providerSlash))

	settle, err := NewServiceBlockSTMOperation(ServiceBlockSTMSettlePayment, map[string]string{"escrow_id": "escrow.a"}, 0)
	require.NoError(t, err)
	dispute, err := NewServiceBlockSTMOperation(ServiceBlockSTMDisputePayment, map[string]string{"call_id": "call.a", "escrow_id": "escrow.a"}, 0)
	require.NoError(t, err)
	require.True(t, ServiceBlockSTMOperationsConflict(settle, dispute))

	interfaceUpdate, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateInterface, map[string]string{"interface_hash": "iface.v2"}, 0)
	require.NoError(t, err)
	descriptorUpdate, err := NewServiceBlockSTMOperation(ServiceBlockSTMUpdateDescriptorIface, map[string]string{"service_id": "svc.dex", "interface_hash": "iface.v2"}, 0)
	require.NoError(t, err)
	require.True(t, ServiceBlockSTMOperationsConflict(interfaceUpdate, descriptorUpdate))
}

func TestServiceBlockSTMHighVolumeServiceCallsConflictOnlyOnSharedStateKeys(t *testing.T) {
	ops := make([]ServiceBlockSTMOperation, 0, 64)
	for i := 0; i < 64; i++ {
		op, err := NewServiceBlockSTMOperation(ServiceBlockSTMExecuteOnChainCall, map[string]string{
			"call_id":	fmt.Sprintf("call.%02d", i),
			"service_id":	"svc.dex",
			"state_key":	fmt.Sprintf("pool/%02d", i),
		}, 0)
		require.NoError(t, err)
		ops = append(ops, op)
	}
	for i := range ops {
		for j := i + 1; j < len(ops); j++ {
			require.False(t, ServiceBlockSTMOperationsConflict(ops[i], ops[j]), "ops %d and %d should be parallel-safe", i, j)
		}
	}

	shared, err := NewServiceBlockSTMOperation(ServiceBlockSTMExecuteOnChainCall, map[string]string{
		"call_id":	"call.shared",
		"service_id":	"svc.dex",
		"state_key":	"pool/07",
	}, 0)
	require.NoError(t, err)
	require.True(t, ServiceBlockSTMOperationsConflict(ops[7], shared))
}

func requireStoreV2Entry(t *testing.T, layout ServiceStoreV2Layout, prefix string) ServiceStoreV2Entry {
	t.Helper()
	entry, found := serviceStoreV2EntryByPrefix(layout.Entries, prefix)
	require.True(t, found, prefix)
	require.NoError(t, entry.Validate())
	return entry
}

func removeStoreV2Entry(entries []ServiceStoreV2Entry, prefix string) []ServiceStoreV2Entry {
	out := make([]ServiceStoreV2Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Prefix != prefix {
			out = append(out, entry)
		}
	}
	return out
}
