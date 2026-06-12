package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceRegistryStateKeysMatchSpecification(t *testing.T) {
	interfaceHash := testHash("interface")
	callID := testHash("call")

	key, err := ServiceDescriptorStateKey("identity-resolver")
	require.NoError(t, err)
	require.Equal(t, "services/descriptors/identity-resolver", key)

	key, err = ServiceAnchorStateKey("identity-resolver")
	require.NoError(t, err)
	require.Equal(t, "services/anchors/identity-resolver", key)

	key, err = ServiceInterfaceStateKey(interfaceHash)
	require.NoError(t, err)
	require.Equal(t, "services/interfaces/"+interfaceHash, key)

	key, err = ServiceOwnerStateKey(DefaultAuthority, "identity-resolver")
	require.NoError(t, err)
	require.Equal(t, "services/owners/"+DefaultAuthority+"/identity-resolver", key)

	key, err = ServiceNameStateKey("identity.aet")
	require.NoError(t, err)
	require.Equal(t, "services/names/identity.aet", key)

	key, err = IdentityServiceBindingStateKey("identity.aet", "identity-resolver")
	require.NoError(t, err)
	require.Equal(t, "services/identity_bindings/identity.aet/identity-resolver", key)

	key, err = ServiceProviderStateKey("fog-compute", "provider.compute.low")
	require.NoError(t, err)
	require.Equal(t, "services/providers/fog-compute/provider.compute.low", key)

	key, err = ServiceExpiryStateKey(7, "identity-resolver")
	require.NoError(t, err)
	require.Equal(t, "services/expiry/00000000000000000007/identity-resolver", key)

	key, err = ServiceReputationStateKey("provider.compute.low")
	require.NoError(t, err)
	require.Equal(t, "services/reputation/provider.compute.low", key)

	key, err = ServiceReceiptStateKey("identity-resolver", callID)
	require.NoError(t, err)
	require.Equal(t, "services/receipts/identity-resolver/"+callID, key)
}

func TestServiceRegistryStateIndexesDescriptorAnchorOwnerNameIdentityAndExpiry(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	anchor, err := NewServiceAnchorFromDescriptor(service)
	require.NoError(t, err)

	state, err := NewServiceRegistryState([]ServiceDescriptor{service}, []ServiceAnchor{anchor}, nil, nil, nil, nil, 10)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	require.Equal(t, ComputeServiceRegistryStateRoot(state), state.StateRoot)
	require.Len(t, state.Descriptors, 1)
	require.Len(t, state.Anchors, 1)
	require.Len(t, state.Interfaces, 1)
	require.Len(t, state.OwnerIndex, 1)
	require.Len(t, state.NameIndex, 1)
	require.Len(t, state.IdentityBindings, 1)
	require.Len(t, state.ExpiryIndex, 1)

	expectedKeys := map[string]string{}
	for _, entry := range state.Entries {
		expectedKeys[entry.Key] = entry.Value
		require.NoError(t, entry.Validate())
	}
	descriptorKey, err := ServiceDescriptorStateKey(service.ServiceID)
	require.NoError(t, err)
	require.Equal(t, ComputeServiceDescriptorHash(service), expectedKeys[descriptorKey])

	anchorKey, err := ServiceAnchorStateKey(service.ServiceID)
	require.NoError(t, err)
	require.Equal(t, anchor.AnchorHash, expectedKeys[anchorKey])

	interfaceKey, err := ServiceInterfaceStateKey(service.Interface.InterfaceHash)
	require.NoError(t, err)
	require.Equal(t, service.Interface.InterfaceHash, expectedKeys[interfaceKey])

	ownerKey, err := ServiceOwnerStateKey(service.Owner, service.ServiceID)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, expectedKeys[ownerKey])

	nameKey, err := ServiceNameStateKey(service.Discovery.ServiceName)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, expectedKeys[nameKey])

	identityKey, err := IdentityServiceBindingStateKey(service.Discovery.IdentityName, service.ServiceID)
	require.NoError(t, err)
	require.Equal(t, state.IdentityBindings[0].BindingHash, expectedKeys[identityKey])

	expiryKey, err := ServiceExpiryStateKey(service.ExpiryHeight, service.ServiceID)
	require.NoError(t, err)
	require.Equal(t, service.ServiceID, expectedKeys[expiryKey])
}

func TestServiceRegistryStateIndexesProvidersReputationAndReceipts(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	provider := testFogProvider("provider.compute.low", "3", 80)
	providerRecord, err := NewProviderRecord("fog-compute", provider)
	require.NoError(t, err)
	reputation, err := NewReputationRecord(provider.ProviderID, provider.ReputationScore, 10, 1, 30)
	require.NoError(t, err)

	ctx := ServiceConsensusContext{ChainID: "aetra-test", Height: 50}
	call := servicePipelineCall(ctx, service, "resolve", ServiceCallKindOnChain, 1, "identity/state", "1")
	receipt, err := NewServiceCallReceipt(call, ServiceExecutionOutcome{
		CallID:		call.CallID,
		Status:		ServiceCallStatusExecuted,
		ResponseHash:	testHash("identity/response"),
		PaymentStatus:	ServicePaymentStatusSettled,
		GasUsed:	100,
		ExecutedHeight:	50,
		AnchoredHeight:	50,
	})
	require.NoError(t, err)

	state, err := NewServiceRegistryState([]ServiceDescriptor{service}, nil, nil, []ProviderRecord{providerRecord}, []ReputationRecord{reputation}, []ServiceReceipt{receipt}, 51)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	require.Len(t, state.Providers, 1)
	require.Len(t, state.Reputations, 1)
	require.Len(t, state.Receipts, 1)

	entries := map[string]ServiceRegistryStateEntry{}
	for _, entry := range state.Entries {
		entries[entry.Key] = entry
	}
	providerKey, err := ServiceProviderStateKey(providerRecord.ServiceID, provider.ProviderID)
	require.NoError(t, err)
	require.Equal(t, ServiceRegistryStateProvider, entries[providerKey].EntryType)
	require.Equal(t, providerRecord.RecordHash, entries[providerKey].Value)

	reputationKey, err := ServiceReputationStateKey(provider.ProviderID)
	require.NoError(t, err)
	require.Equal(t, ServiceRegistryStateReputation, entries[reputationKey].EntryType)
	require.Equal(t, reputation.RecordHash, entries[reputationKey].Value)

	receiptKey, err := ServiceReceiptStateKey(receipt.ServiceID, receipt.CallID)
	require.NoError(t, err)
	require.Equal(t, ServiceRegistryStateReceipt, entries[receiptKey].EntryType)
	require.Equal(t, receipt.ReceiptHash, entries[receiptKey].Value)
}

func TestServiceRegistryStateRejectsDuplicateServiceNameAndExpiredBinding(t *testing.T) {
	first := testService("identity-resolver", ZoneIDIdentity)
	second := testService("identity-alias", ZoneIDIdentity)
	second.Discovery.ServiceName = first.Discovery.ServiceName

	_, err := NewServiceRegistryState([]ServiceDescriptor{first, second}, nil, nil, nil, nil, nil, 10)
	require.ErrorContains(t, err, "key collision")

	binding := IdentityServiceBinding{
		IdentityName:	"identity.aet",
		ServiceID:	"identity-resolver",
		Owner:		DefaultAuthority,
		DescriptorHash:	ComputeServiceDescriptorHash(first),
		CreatedHeight:	10,
		ExpiryHeight:	10,
	}
	binding.BindingHash = ComputeIdentityServiceBindingHash(binding)
	require.ErrorContains(t, binding.Validate(), "expiry")
}
