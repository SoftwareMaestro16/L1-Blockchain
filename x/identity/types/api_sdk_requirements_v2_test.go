package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityAPIAndSDKRequirementsV2CoversSection14(t *testing.T) {
	requirements, err := DefaultIdentityAPIAndSDKRequirementsV2()
	require.NoError(t, err)
	require.NoError(t, ValidateIdentityAPIAndSDKRequirementsV2(requirements))

	require.ElementsMatch(t, requiredIdentityNodeAPIEndpointsV2(), requirements.NodeAPIEndpoints)
	require.ElementsMatch(t, requiredIdentityWalletSDKHelpersV2(), requirements.WalletSDKHelpers)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIResolvePrimary)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIResolveContractTarget)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIResolveServiceEndpoint)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIResolveInterface)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIResolveRoutingMetadata)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIResolveReverse)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIQueryDomainLifecycle)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIQueryRegistrationPrice)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIQueryRenewalPrice)
	require.Contains(t, requirements.NodeAPIEndpoints, IdentityNodeAPIQueryDelegationAuth)

	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKNormalizeName)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKValidateName)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKResolvePrimaryVerified)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKResolveContractTargetVerified)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKResolveServiceVerified)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKResolveInterfaceVerified)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKVerifyResolutionProof)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKBuildSendByNameTx)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKBuildInvokeByNameTx)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKRenderVerifiedInterface)
	require.Contains(t, requirements.WalletSDKHelpers, IdentityWalletSDKCheckReverseResolution)

	require.Len(t, requirements.IndexerSchemas, 6)
	for _, schema := range requirements.IndexerSchemas {
		require.True(t, schema.RequiresProofReference)
		require.Equal(t, IdentityNodeAPIProofPassthroughVersionV2, schema.ProofPassthroughFormat)
		require.NoError(t, ValidateIdentityIndexerEventSchemaV2(schema))
	}
}

func TestIdentityNodeAPIV2EndpointsReturnOptionalProofPassthrough(t *testing.T) {
	state := routingIntegrationState(t)
	state, _, err := SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)
	api := newIdentityNodeAPITestAPI(t, state, 15)

	primary := api.ResolvePrimaryAddress(IdentityNodeAPIRequestV2{Name: "alice.aet", IncludeProof: true})
	require.Equal(t, IdentityQueryOK, primary.QueryCode)
	require.Equal(t, addr(2), primary.Address)
	require.NotNil(t, primary.ProofPassthrough)
	require.NoError(t, ValidateIdentityProofPassthroughFormatV2(*primary.ProofPassthrough))
	verified, err := IdentityWalletSDKVerifyResolutionProofV2(IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		trustedHeaderForProofV2(*primary.ProofPassthrough.Proof),
		Proof:			*primary.ProofPassthrough.Proof,
		TargetType:		IdentityResolutionTargetPrimary,
		CurrentHeight:		15,
		AllowRenewalWindow:	true,
		NormalizationVersion:	NameNormalizationVersionV2,
	})
	require.NoError(t, err)
	require.Equal(t, addr(2), verified.Address)

	contract := api.ResolveContractTarget(IdentityNodeAPIRequestV2{Name: "alice.aet", TargetKey: ResolverKeyContract, IncludeProof: true})
	require.Equal(t, IdentityQueryOK, contract.QueryCode)
	require.Equal(t, addr(3), contract.Address)
	require.NotNil(t, contract.ContractTarget)
	require.NotNil(t, contract.ProofPassthrough)

	service := api.ResolveServiceEndpoint(IdentityNodeAPIRequestV2{Name: "alice.aet", TargetKey: "rpc", IncludeProof: true})
	require.Equal(t, IdentityQueryOK, service.QueryCode)
	require.NotNil(t, service.ServiceEndpoint)
	require.Equal(t, "https://rpc.aet", service.ServiceEndpoint.Endpoint)
	require.NotNil(t, service.ProofPassthrough)

	iface := api.ResolveInterfaceDescriptor(IdentityNodeAPIRequestV2{Name: "alice.aet", TargetKey: "aw5", IncludeProof: true})
	require.Equal(t, IdentityQueryOK, iface.QueryCode)
	require.NotNil(t, iface.InterfaceDescriptor)
	require.Equal(t, "aw5", iface.InterfaceDescriptor.InterfaceID)
	require.NotNil(t, iface.ProofPassthrough)

	route := api.ResolveRoutingMetadata(IdentityNodeAPIRequestV2{Name: "alice.aet", IncludeProof: true})
	require.Equal(t, IdentityQueryOK, route.QueryCode)
	require.NotNil(t, route.RoutingMetadata)
	require.Equal(t, "swap-route", route.RoutingMetadata.RouteID)
	require.NotNil(t, route.ProofPassthrough)

	reverse := api.ResolveReverseRecord(IdentityNodeAPIRequestV2{Address: addr(2), IncludeProof: true})
	require.Equal(t, IdentityQueryOK, reverse.QueryCode)
	require.NotNil(t, reverse.ReverseRecord)
	require.True(t, reverse.ReverseRecord.Verified)
	require.NotNil(t, reverse.ProofPassthrough)

	lifecycle := api.QueryDomainLifecycleState(IdentityNodeAPIRequestV2{Name: "alice.aet"})
	require.Equal(t, IdentityQueryOK, lifecycle.QueryCode)
	require.Equal(t, DomainLifecycleActive, lifecycle.Lifecycle)

	regPrice := api.QueryRegistrationPrice(IdentityNodeAPIRequestV2{Name: "market.aet", DurationBlocks: DefaultIdentityPricingParamsV2().BaseDurationBlocks})
	require.Equal(t, IdentityQueryOK, regPrice.QueryCode)
	require.NotNil(t, regPrice.RegistrationPrice)

	renewPrice := api.QueryRenewalPrice(IdentityNodeAPIRequestV2{Name: "market.aet", RenewalPeriods: 1})
	require.Equal(t, IdentityQueryOK, renewPrice.QueryCode)
	require.NotNil(t, renewPrice.RenewalPrice)

	delegation, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, 90, 0, ResolverKeyPrimary, 15)
	require.NoError(t, err)
	auth := api.QueryDelegationAuthorization(IdentityNodeAPIRequestV2{
		Delegation:	&delegation,
		DelegationAuth: PartialDelegationAuthorizationV2{
			Scope:				DelegationScopeResolverUpdate,
			Permission:			ResolverKeyPrimary,
			RecordKey:			ResolverKeyPrimary,
			Height:				20,
			ExpectedDelegationVersion:	delegation.DelegationVersion,
		},
	})
	require.Equal(t, IdentityQueryOK, auth.QueryCode)
	require.True(t, auth.DelegationAuthorized)
}

func TestIdentityWalletSDKV2VerifiedHelpersBuildTransactions(t *testing.T) {
	state := routingIntegrationState(t)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)
	header := trustedHeaderForProofV2(proof)
	interfaceHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)

	normalized, err := IdentityWalletSDKNormalizeNameV2("Alice.AET")
	require.NoError(t, err)
	require.Equal(t, "alice.aet", normalized)
	require.NoError(t, IdentityWalletSDKValidateNameV2("alice.aet"))

	primary, err := IdentityWalletSDKResolvePrimaryVerifiedV2(IdentitySendByNameRequestV2{
		Name:			"alice.aet",
		CurrentHeight:		15,
		IncludeAuditMemo:	true,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		header,
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, primary.ProofVerified)
	sendTx, err := IdentityWalletSDKBuildSendByNameTxV2(primary, "naet", "100")
	require.NoError(t, err)
	require.Equal(t, addr(2), sendTx.ToAddress)
	require.NoError(t, ValidateIdentityWalletSendByNameTxV2(sendTx))

	invoke, err := IdentityWalletSDKResolveContractTargetVerifiedV2(IdentityInvokeByNameRequestV2{
		Name:			"alice.aet",
		TargetID:		ResolverKeyContract,
		InterfaceID:		"aw5",
		ExpectedInterfaceHash:	interfaceHash,
		Method:			"swap",
		PayloadHash:		identityHash("payload"),
		CurrentHeight:		15,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		header,
		Proof:			&proof,
	})
	require.NoError(t, err)
	invokeTx, err := IdentityWalletSDKBuildInvokeByNameTxV2(invoke)
	require.NoError(t, err)
	require.Equal(t, addr(3), invokeTx.ContractAddress)
	require.Equal(t, "swap", invokeTx.Entrypoint)
	require.NoError(t, ValidateIdentityWalletInvokeByNameTxV2(invokeTx))

	service, err := IdentityWalletSDKResolveServiceVerifiedV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		CurrentHeight:		15,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		header,
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, service.ProofVerified)
	require.Equal(t, "https://rpc.aet", service.Endpoint.Endpoint)

	inlineProof := routingIntegrationInlineInterfaceProof(t, state)
	inlineHash := inlineProof.ResolverRecord.InterfaceDescriptors[0].SchemaHash
	rendered, err := IdentityWalletSDKRenderVerifiedInterfaceV2(IdentityInterfaceSchemaRequestV2{
		Name:			"alice.aet",
		InterfaceID:		"aw5",
		ExpectedSchemaHash:	inlineHash,
		WalletPolicy:		DefaultIdentityInterfaceWalletPolicyV2(),
		CurrentHeight:		15,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(inlineProof),
		Proof:			&inlineProof,
	})
	require.NoError(t, err)
	require.True(t, rendered.ProofVerified)
	require.True(t, rendered.SchemaHashVerified)
	require.NotEmpty(t, rendered.SchemaSource)
}

func TestIdentityIndexerV2ProofPassthroughAndReplay(t *testing.T) {
	state := routingIntegrationState(t)
	api := newIdentityNodeAPITestAPI(t, state, 15)
	primary := api.ResolvePrimaryAddress(IdentityNodeAPIRequestV2{Name: "alice.aet", IncludeProof: true})
	require.Equal(t, IdentityQueryOK, primary.QueryCode)
	require.NotNil(t, primary.ProofPassthrough)
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)

	domainEvent, err := BuildIdentityIndexerEventV2(IdentityIndexerEventDomainV2, 15, map[string]string{
		IdentityIndexerAttrName:		"alice.aet",
		IdentityIndexerAttrNameHash:		nameHash,
		IdentityIndexerAttrOwner:		hex.EncodeToString(addr(1)),
		IdentityIndexerAttrExpiryHeight:	"1010",
		IdentityIndexerAttrRecordVersion:	"1",
	}, primary.ProofPassthrough)
	require.NoError(t, err)
	resolverEvent, err := BuildIdentityIndexerEventV2(IdentityIndexerEventResolverV2, 15, map[string]string{
		IdentityIndexerAttrName:		"alice.aet",
		IdentityIndexerAttrNameHash:		nameHash,
		IdentityIndexerAttrResolver:		hex.EncodeToString(addr(1)),
		IdentityIndexerAttrRecordVersion:	"1",
	}, primary.ProofPassthrough)
	require.NoError(t, err)
	expiryEvent, err := BuildIdentityIndexerEventV2(IdentityIndexerEventExpiryV2, 16, map[string]string{
		IdentityIndexerAttrNameHash:		nameHash,
		IdentityIndexerAttrExpiryHeight:	"1010",
		IdentityIndexerAttrRecordVersion:	"1",
	}, primary.ProofPassthrough)
	require.NoError(t, err)

	replay, err := ReplayIdentityIndexerEventsV2([]IdentityIndexerEventV2{domainEvent, resolverEvent, expiryEvent}, nil)
	require.NoError(t, err)
	require.Equal(t, uint64(3), replay.EventsReplayed)
	require.Equal(t, "alice.aet", replay.DomainIndex[nameHash])
	require.Contains(t, replay.OwnerIndex[hex.EncodeToString(addr(1))], nameHash)
	require.Contains(t, replay.ResolverIndex[hex.EncodeToString(addr(1))], nameHash)
	require.Contains(t, replay.ExpiryIndex["1010"], nameHash)
	require.NotEmpty(t, replay.ProofReferences[primary.ProofPassthrough.ProofCommitmentHash])
	require.NotEmpty(t, replay.ReplayHash)

	missingProof := domainEvent
	missingProof.ProofPassthrough = nil
	missingProof.EventHash = ComputeIdentityIndexerEventHashV2(missingProof)
	_, err = ReplayIdentityIndexerEventsV2([]IdentityIndexerEventV2{missingProof}, nil)
	require.ErrorContains(t, err, "requires proof passthrough or proof reference")
}

func newIdentityNodeAPITestAPI(t *testing.T, state IdentityState, height uint64) IdentityNodeAPIV2 {
	t.Helper()
	appHash, err := IdentityStateRoot(state)
	require.NoError(t, err)
	api, err := NewIdentityNodeAPIV2(IdentityQueryContextV2{State: state, Height: height, DefaultTTL: 30}, "aetra-local-1", appHash)
	require.NoError(t, err)
	return api
}
