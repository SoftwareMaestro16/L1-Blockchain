package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutingIntegrationModuleBreakdownV2CoversSection136(t *testing.T) {
	breakdown, err := DefaultRoutingIntegrationModuleBreakdownV2()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())

	require.Equal(t, RoutingIntegrationModulePathV2, breakdown.ModulePath)
	require.ElementsMatch(t, requiredRoutingIntegrationStateObjectsV2(), breakdown.StateObjects)
	require.ElementsMatch(t, requiredRoutingIntegrationMessagesV2(), breakdown.Messages)
	require.ElementsMatch(t, requiredRoutingIntegrationQueriesV2(), breakdown.Queries)
	require.ElementsMatch(t, requiredRoutingIntegrationIntegrationPointsV2(), breakdown.IntegrationPoints)

	require.Contains(t, breakdown.StateObjects, RoutingIntegrationStateRoutingPolicy)
	require.Contains(t, breakdown.StateObjects, RoutingIntegrationStateIdentityTransactionMapping)
	require.Contains(t, breakdown.StateObjects, RoutingIntegrationStateContractInvocationMapping)
	require.Contains(t, breakdown.StateObjects, RoutingIntegrationStateServiceMapping)
	require.Contains(t, breakdown.StateObjects, RoutingIntegrationStateInterfaceMapping)
	require.Contains(t, breakdown.StateObjects, RoutingIntegrationStateExecutionHintPolicy)

	require.Contains(t, breakdown.Messages, RoutingIntegrationMsgSetRoutingPolicy)
	require.Contains(t, breakdown.Messages, RoutingIntegrationMsgUpdateExecutionHints)
	require.Contains(t, breakdown.Messages, RoutingIntegrationMsgRegisterInterfaceDescriptor)
	require.Contains(t, breakdown.Messages, RoutingIntegrationMsgRegisterServiceEndpoint)
	require.Contains(t, breakdown.Messages, RoutingIntegrationMsgClearRoutingMetadata)

	require.Contains(t, breakdown.Queries, RoutingIntegrationQueryTransactionMapping)
	require.Contains(t, breakdown.Queries, RoutingIntegrationQueryContractInvocationMapping)
	require.Contains(t, breakdown.Queries, RoutingIntegrationQueryServiceMapping)
	require.Contains(t, breakdown.Queries, RoutingIntegrationQueryInterfaceMapping)
	require.Contains(t, breakdown.Queries, RoutingIntegrationQueryExecutionHints)
	require.Contains(t, breakdown.Queries, RoutingIntegrationQueryResolvedExecutionTarget)

	require.Contains(t, breakdown.BackingPrimitives, "BuildIdentitySendByNameV2")
	require.Contains(t, breakdown.BackingPrimitives, "BuildIdentityInvokeByNameV2")
	require.Contains(t, breakdown.BackingPrimitives, "BuildIdentityServiceDiscoveryV2")
	require.Contains(t, breakdown.BackingPrimitives, "BuildIdentityInterfaceSchemaMappingV2")
	require.Contains(t, breakdown.BackingPrimitives, "ResolveNamedExecutionTarget")
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecResolversPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecInterfaceMetadataPrefix)
}

func TestRoutingIntegrationModuleMappingsAndResolvedTargetsV2(t *testing.T) {
	state := routingIntegrationState(t)
	interfaceHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)

	tx, err := BuildRoutingIntegrationTransactionMappingV2(IdentitySendByNameRequestV2{
		Name:			"alice.aet",
		State:			state,
		Height:			14,
		RecordTTL:		30,
		CurrentHeight:		15,
		IncludeAuditMemo:	true,
	})
	require.NoError(t, err)
	require.Equal(t, "alice.aet", tx.Name)
	require.Equal(t, addr(2), tx.Address)
	require.True(t, strings.HasPrefix(tx.AuditMemo, IdentityAuditMemoPrefixV2+";"))
	require.NoError(t, ValidateIdentityTransactionMappingV2(tx))

	contract, err := BuildRoutingIntegrationContractInvocationMappingV2(IdentityInvokeByNameRequestV2{
		Name:			"alice.aet",
		TargetID:		ResolverKeyContract,
		InterfaceID:		"aw5",
		ExpectedInterfaceHash:	interfaceHash,
		Method:			"swap",
		PayloadHash:		identityHash("payload"),
		State:			state,
		Height:			14,
		RecordTTL:		30,
		CurrentHeight:		15,
	})
	require.NoError(t, err)
	require.Equal(t, ResolverKeyContract, contract.TargetID)
	require.Equal(t, addr(3), contract.ContractAddress)
	require.Equal(t, "swap", contract.Entrypoint)
	require.Equal(t, interfaceHash, contract.InterfaceHash)
	require.True(t, contract.RequiresConfirm)
	require.True(t, contract.SimulationNeeded)
	require.NoError(t, ValidateContractInvocationMappingV2(contract))

	service, err := BuildRoutingIntegrationServiceMappingV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		SupportedServiceTypes:	[]string{IdentityServiceTypeGenericV1},
		State:			state,
		Height:			14,
		RecordTTL:		30,
		CurrentHeight:		15,
	})
	require.NoError(t, err)
	require.Equal(t, "rpc", service.ServiceID)
	require.Equal(t, "https://rpc.aet", service.Endpoint.Endpoint)
	require.Empty(t, service.FallbackEndpoints)
	require.False(t, service.EndpointAvailabilityConsensusGuaranteed)
	require.NoError(t, ValidateRoutingIntegrationServiceEndpointAvailabilityV2(service, true))
	require.NoError(t, ValidateServiceMappingV2(service))

	proof := routingIntegrationInlineInterfaceProof(t, state)
	inlineHash := proof.ResolverRecord.InterfaceDescriptors[0].SchemaHash
	iface, err := BuildRoutingIntegrationInterfaceMappingV2(IdentityInterfaceSchemaRequestV2{
		Name:			"alice.aet",
		InterfaceID:		"aw5",
		ExpectedSchemaHash:	inlineHash,
		WalletPolicy:		DefaultIdentityInterfaceWalletPolicyV2(),
		CurrentHeight:		15,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.Equal(t, "aw5", iface.InterfaceID)
	require.Equal(t, inlineHash, iface.SchemaHash)
	require.True(t, iface.SchemaHashVerified)
	require.True(t, iface.UserConfirmationRequired)
	require.True(t, iface.ExecutionTargetImmutable)
	require.NoError(t, ValidateInterfaceMappingV2(iface))

	helper, err := BuildRoutingIntegrationWalletSDKHelperV2(&tx, &contract, &service, &iface, true)
	require.NoError(t, err)
	require.True(t, helper.ProofRequired)
	require.True(t, helper.AdvisoryOnly)
	require.True(t, helper.UserConfirmation)
	require.NotEmpty(t, helper.HelperHash)

	resolvedTx, err := QueryRoutingIntegrationResolvedExecutionTargetV2(RoutingIntegrationResolvedExecutionTargetRequestV2{
		Name:		"alice.aet",
		TargetType:	IdentityResolutionTargetPrimary,
		State:		state,
		Height:		14,
		RecordTTL:	30,
		CurrentHeight:	15,
	})
	require.NoError(t, err)
	require.Equal(t, addr(2), resolvedTx.Address)
	require.Equal(t, IdentityResolutionTargetPrimary, resolvedTx.TargetType)
	require.NotEmpty(t, resolvedTx.ResultHash)

	resolvedContract, err := QueryRoutingIntegrationResolvedExecutionTargetV2(RoutingIntegrationResolvedExecutionTargetRequestV2{
		Name:		"alice.aet",
		TargetType:	IdentityResolutionTargetContract,
		TargetKey:	ResolverKeyContract,
		InterfaceID:	"aw5",
		ExpectedHash:	interfaceHash,
		Method:		"swap",
		PayloadHash:	identityHash("payload"),
		State:		state,
		Height:		14,
		RecordTTL:	30,
		CurrentHeight:	15,
	})
	require.NoError(t, err)
	require.Equal(t, addr(3), resolvedContract.Address)
	require.Equal(t, ResolverKeyContract, resolvedContract.TargetKey)
	require.Equal(t, "swap", resolvedContract.Route.Entrypoint)

	resolvedService, err := QueryRoutingIntegrationResolvedExecutionTargetV2(RoutingIntegrationResolvedExecutionTargetRequestV2{
		Name:		"alice.aet",
		TargetType:	IdentityResolutionTargetService,
		TargetKey:	"rpc",
		State:		state,
		Height:		14,
		RecordTTL:	30,
		CurrentHeight:	15,
	})
	require.NoError(t, err)
	require.Equal(t, "https://rpc.aet", resolvedService.Endpoint)
	require.Equal(t, "rpc", resolvedService.TargetKey)
}

func TestRoutingIntegrationModuleFailureGuardsV2(t *testing.T) {
	state := routingIntegrationState(t)
	require.NoError(t, ValidateRoutingIntegrationMetadataFreshnessV2(1, 1))
	require.ErrorContains(t, ValidateRoutingIntegrationMetadataFreshnessV2(1, 2), string(RoutingIntegrationFailureStaleMetadata))

	interfaceHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)
	require.ErrorContains(t, ValidateRoutingIntegrationInterfaceTargetV2(ResolverKeyContract, InterfaceDescriptorV2{
		InterfaceID:			"aw5",
		SchemaHash:			interfaceHash,
		Version:			"v1",
		RenderPolicy:			IdentityRenderPolicyConfirmV2,
		ContractTargetIDOptional:	"other",
	}), string(RoutingIntegrationFailureWrongInterfaceTarget))

	service, err := BuildRoutingIntegrationServiceMappingV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		State:			state,
		Height:			14,
		RecordTTL:		30,
		CurrentHeight:		15,
	})
	require.NoError(t, err)
	require.ErrorContains(t, ValidateRoutingIntegrationServiceEndpointAvailabilityV2(service, false), string(RoutingIntegrationFailureServiceEndpointUnavailable))

	record, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 14, 30)
	require.NoError(t, err)
	record.ExecutionHints = []ExecutionHintV2{{
		Key:				"hint.default_gas_limit",
		Value:				"1000",
		DefaultGasLimitHint:		1_000,
		PreferredFeeMode:		"priority",
		AsyncAllowed:			true,
		RequiresInterfaceConfirmation:	true,
		SimulationRequired:		true,
	}}
	require.ErrorContains(t, ValidateRoutingIntegrationExecutionHintsV2(record, ExecutionHintPolicyV2{
		MaxGasLimitHint:	MaxExecutionGasLimitHintV2,
		AllowedFeeModes:	[]string{"standard"},
		AllowAsyncExecution:	false,
	}), string(RoutingIntegrationFailureExecutionHintConflict))

	require.ErrorContains(t, ValidateRoutingIntegrationAdvisoryAuthorizationV2(RoutingIntegrationWalletSDKHelperV2{
		AdvisoryOnly:		true,
		UserConfirmation:	true,
	}, true), string(RoutingIntegrationFailureAdvisoryAsAuthorization))
	require.ErrorContains(t, ValidateRoutingIntegrationAdvisoryAuthorizationV2(RoutingIntegrationWalletSDKHelperV2{
		AdvisoryOnly:		true,
		UserConfirmation:	false,
	}, false), "requires user confirmation")
}

func TestRoutingIntegrationModuleBreakdownV2RejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultRoutingIntegrationModuleBreakdownV2()
	require.NoError(t, err)

	missingMessage := breakdown
	missingMessage.Messages = missingMessage.Messages[:len(missingMessage.Messages)-1]
	_, err = NewRoutingIntegrationModuleBreakdownV2(missingMessage)
	require.ErrorContains(t, err, "message entries")

	duplicateQuery := breakdown
	duplicateQuery.Queries[0] = duplicateQuery.Queries[1]
	_, err = NewRoutingIntegrationModuleBreakdownV2(duplicateQuery)
	require.ErrorContains(t, err, "duplicate query")
}

func routingIntegrationState(t *testing.T) IdentityState {
	t.Helper()

	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	serviceKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	backupServiceKey, err := ResolverMetadataServiceKey("rpc-backup")
	require.NoError(t, err)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	metadata := mustMetadataV2(t, []ResolverMetadataEntry{
		{Key: ResolverMetadataRouteEntrypoint, Value: "swap"},
		{Key: "route.id", Value: "swap-route"},
		{Key: "route.target_type", Value: string(IdentityResolutionTargetContract)},
		{Key: "route.preferred_target", Value: ResolverKeyContract},
		{Key: serviceKey, Value: "https://rpc.aet"},
		{Key: backupServiceKey, Value: "https://backup.aet"},
		{Key: interfaceKey, Value: "wallet-v1"},
		{Key: "hint.requires_interface_confirmation", Value: "true"},
		{Key: "hint.simulation_required", Value: "true"},
	})
	next, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(3),
		Metadata:	metadata,
	}, 12)
	require.NoError(t, err)
	return next
}

func routingIntegrationInlineInterfaceProof(t *testing.T, state IdentityState) IdentityResolutionProofFormatV2 {
	t.Helper()

	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)
	inlineSchema := `{"type":"wallet","version":"v1"}`
	inlineHash, err := InterfaceDescriptorHashV2(inlineSchema)
	require.NoError(t, err)
	require.NotNil(t, proof.ResolverRecord)
	proof.ResolverRecord.InterfaceDescriptors = []InterfaceDescriptorV2{{
		InterfaceID:			"aw5",
		SchemaHash:			inlineHash,
		SchemaInlineOptional:		inlineSchema,
		Version:			"v1",
		RenderPolicy:			IdentityRenderPolicyConfirmV2,
		ContractTargetIDOptional:	ResolverKeyContract,
	}}
	proof.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(proof)
	require.NoError(t, ValidateIdentityResolutionProofFormatV2(proof))
	return proof
}
