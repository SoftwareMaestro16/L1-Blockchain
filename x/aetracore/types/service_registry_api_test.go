package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceRegistryMessagesValidateAllRegistryOperations(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	provider := testFogProvider("provider.compute.low", "3", 80)
	receipt := testRegistryAPIReceipt(t, service)
	authHash := ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10)
	newOwner := "4:0000000000000000000000000000000000000000000000000000000000000002"

	messages := []ServiceRegistryMessage{}
	register, err := NewMsgRegisterService(service.Owner, service, authHash)
	require.NoError(t, err)
	messages = append(messages, register)

	update, err := NewMsgUpdateService(service.Owner, service, 1)
	require.NoError(t, err)
	messages = append(messages, update)

	renew, err := NewMsgRenewService(service.Owner, service.ServiceID, 200, 1)
	require.NoError(t, err)
	messages = append(messages, renew)

	disable, err := NewMsgDisableService(service.Owner, service.ServiceID, "owner-request", 1)
	require.NoError(t, err)
	messages = append(messages, disable)

	transfer, err := NewMsgTransferService(service.Owner, service.ServiceID, newOwner, 1)
	require.NoError(t, err)
	messages = append(messages, transfer)

	bind, err := NewMsgBindServiceIdentity(service.Owner, service.ServiceID, "identity.aet", 1)
	require.NoError(t, err)
	messages = append(messages, bind)

	unbind, err := NewMsgUnbindServiceIdentity(service.Owner, service.ServiceID, "identity.aet", 1)
	require.NoError(t, err)
	messages = append(messages, unbind)

	registerProvider, err := NewMsgRegisterProvider(service.Owner, "fog-compute", provider)
	require.NoError(t, err)
	messages = append(messages, registerProvider)

	updateProvider, err := NewMsgUpdateProvider(service.Owner, "fog-compute", provider)
	require.NoError(t, err)
	messages = append(messages, updateProvider)

	stake, err := NewMsgStakeProviderCollateral(service.Owner, "fog-compute", provider.ProviderID, NativeFeePolicyID, "50", 20)
	require.NoError(t, err)
	messages = append(messages, stake)

	unstake, err := NewMsgUnstakeProviderCollateral(service.Owner, "fog-compute", provider.ProviderID, NativeFeePolicyID, "10", 21)
	require.NoError(t, err)
	messages = append(messages, unstake)

	anchor, err := NewMsgAnchorServiceReceipt(service.Owner, receipt, testHash("receipt/anchor"))
	require.NoError(t, err)
	messages = append(messages, anchor)

	dispute, err := NewMsgSubmitServiceDispute(service.Owner, service.ServiceID, receipt.CallID, provider.ProviderID, testHash("dispute/evidence"), "bad-result", 22)
	require.NoError(t, err)
	messages = append(messages, dispute)

	for _, msg := range messages {
		require.NoError(t, ValidateServiceRegistryMessage(msg), msg.ServiceRegistryMessageName())
		require.NotEmpty(t, msg.ServiceRegistrySigner())
		require.Equal(t, ComputeServiceRegistryMessageHash(msg), registryMessageHashForTest(msg))
	}
}

func TestServiceRegistryMessagesRejectInvalidAuthorityVersionCollateralAndDispute(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	authHash := ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10)

	mismatched := testOffChainService("indexer-feed", ZoneIDApplication)
	mismatched.Owner = "4:0000000000000000000000000000000000000000000000000000000000000002"
	_, err := NewMsgRegisterService(DefaultAuthority, mismatched, authHash)
	require.ErrorContains(t, err, "authority must match descriptor owner")

	_, err = NewMsgUpdateService(service.Owner, service, 2)
	require.ErrorContains(t, err, "expected version")

	_, err = NewMsgTransferService(service.Owner, service.ServiceID, service.Owner, 1)
	require.ErrorContains(t, err, "new owner")

	_, err = NewMsgStakeProviderCollateral(service.Owner, "fog-compute", "provider.compute.low", NativeFeePolicyID, "0", 20)
	require.ErrorContains(t, err, "positive")

	_, err = NewMsgSubmitServiceDispute(service.Owner, service.ServiceID, testHash("call"), "provider.compute.low", "", "bad-result", 22)
	require.ErrorContains(t, err, "evidence")
}

func TestServiceRegistryQueriesResolveAllRegistryIndexes(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	indexer := testOffChainService("indexer-feed", ZoneIDApplication)
	provider := testFogProvider("provider.compute.low", "3", 80)
	providerRecord, err := NewProviderRecord("fog-compute", provider)
	require.NoError(t, err)
	reputation, err := NewReputationRecord(provider.ProviderID, provider.ReputationScore, 10, 1, 30)
	require.NoError(t, err)
	receipt := testRegistryAPIReceipt(t, service)
	anchor, err := NewServiceAnchorFromDescriptor(service)
	require.NoError(t, err)

	state, err := NewServiceRegistryState(
		[]ServiceDescriptor{indexer, service},
		[]ServiceAnchor{anchor},
		nil,
		[]ProviderRecord{providerRecord},
		[]ReputationRecord{reputation},
		[]ServiceReceipt{receipt},
		51,
	)
	require.NoError(t, err)
	params := TestnetParams()

	byID, err := state.QueryService(QueryService{ServiceID: service.ServiceID, IncludeAnchor: true, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, byID.Found)
	require.Equal(t, service.ServiceID, byID.Descriptor.ServiceID)
	require.Equal(t, anchor.AnchorHash, byID.Anchor.AnchorHash)
	require.NoError(t, byID.Proof.Validate())

	byName, err := state.QueryServiceByName(QueryServiceByName{ServiceName: service.Discovery.ServiceName, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, byName.Found)
	require.Equal(t, service.ServiceID, byName.Descriptor.ServiceID)

	byOwner, err := state.QueryServicesByOwner(QueryServicesByOwner{Owner: service.Owner, Pagination: QueryPagination{Limit: 1}}, params)
	require.NoError(t, err)
	require.Equal(t, uint64(2), byOwner.Total)
	require.Len(t, byOwner.Services, 1)

	byIdentity, err := state.QueryServicesByIdentity(QueryServicesByIdentity{IdentityName: service.Discovery.IdentityName}, params)
	require.NoError(t, err)
	require.Equal(t, uint64(1), byIdentity.Total)
	require.Equal(t, service.ServiceID, byIdentity.Services[0].ServiceID)

	providers, err := state.QueryProvidersByService(QueryProvidersByService{ServiceID: "fog-compute"}, params)
	require.NoError(t, err)
	require.Equal(t, uint64(1), providers.Total)
	require.Equal(t, provider.ProviderID, providers.Providers[0].Provider.ProviderID)

	iface, err := state.QueryServiceInterface(QueryServiceInterface{InterfaceHash: service.Interface.InterfaceHash})
	require.NoError(t, err)
	require.True(t, iface.Found)
	require.Equal(t, service.Interface.InterfaceHash, iface.Interface.InterfaceHash)

	payment, err := state.QueryServicePaymentModel(QueryServicePaymentModel{ServiceID: service.ServiceID})
	require.NoError(t, err)
	require.True(t, payment.Found)
	require.Equal(t, registryPaymentModel(service), payment.PaymentModel)

	verification, err := state.QueryServiceVerificationModel(QueryServiceVerificationModel{ServiceID: service.ServiceID})
	require.NoError(t, err)
	require.True(t, verification.Found)
	require.Equal(t, service.Verification.Model, verification.VerificationModel)

	receiptResponse, err := state.QueryServiceReceipt(QueryServiceReceipt{ServiceID: receipt.ServiceID, CallID: receipt.CallID})
	require.NoError(t, err)
	require.True(t, receiptResponse.Found)
	require.Equal(t, receipt.ReceiptHash, receiptResponse.Receipt.ReceiptHash)

	proof, err := state.QueryServiceProof(QueryServiceProof{ServiceID: service.ServiceID})
	require.NoError(t, err)
	require.True(t, proof.Found)
	require.NoError(t, proof.Proof.Validate())

	paramsResponse, err := QueryServiceParamsResponseFor(params)
	require.NoError(t, err)
	require.Equal(t, params.Authority, paramsResponse.Params.Authority)
}

func TestServiceRegistryQueriesValidateInputsAndPagination(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	state, err := NewServiceRegistryState([]ServiceDescriptor{service}, nil, nil, nil, nil, nil, 10)
	require.NoError(t, err)
	params := TestnetParams()

	_, err = state.QueryService(QueryService{})
	require.ErrorContains(t, err, "service id")

	_, err = state.QueryServiceInterface(QueryServiceInterface{InterfaceHash: "bad"})
	require.ErrorContains(t, err, "hash")

	_, err = state.QueryServiceReceipt(QueryServiceReceipt{ServiceID: service.ServiceID, CallID: "bad"})
	require.ErrorContains(t, err, "call id")

	result, err := state.QueryServicesByOwner(QueryServicesByOwner{
		Owner:		service.Owner,
		Pagination:	QueryPagination{Offset: 100, Limit: MaxQueryLimit + 1},
	}, params)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.Total)
	require.Empty(t, result.Services)
}

func testRegistryAPIReceipt(t *testing.T, service ServiceDescriptor) ServiceReceipt {
	t.Helper()
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
	return receipt
}

func registryMessageHashForTest(msg ServiceRegistryMessage) string {
	switch m := msg.(type) {
	case MsgRegisterService:
		return m.MessageHash
	case MsgUpdateService:
		return m.MessageHash
	case MsgRenewService:
		return m.MessageHash
	case MsgDisableService:
		return m.MessageHash
	case MsgTransferService:
		return m.MessageHash
	case MsgBindServiceIdentity:
		return m.MessageHash
	case MsgUnbindServiceIdentity:
		return m.MessageHash
	case MsgRegisterProvider:
		return m.MessageHash
	case MsgUpdateProvider:
		return m.MessageHash
	case MsgStakeProviderCollateral:
		return m.MessageHash
	case MsgUnstakeProviderCollateral:
		return m.MessageHash
	case MsgAnchorServiceReceipt:
		return m.MessageHash
	case MsgSubmitServiceDispute:
		return m.MessageHash
	default:
		return ""
	}
}
