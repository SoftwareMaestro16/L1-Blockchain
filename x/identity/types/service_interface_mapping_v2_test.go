package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityServiceDiscoveryV2ProofAwareFreshnessAndFallback(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	rpcKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	backupKey, err := ResolverMetadataServiceKey("rpc-backup")
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata: mustMetadataV2(t, []ResolverMetadataEntry{
			{Key: rpcKey, Value: "https://rpc.aet"},
			{Key: backupKey, Value: "https://backup.aet"},
		}),
	}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)

	result, err := BuildIdentityServiceDiscoveryV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		SupportedServiceTypes:	[]string{"service.v1"},
		CurrentHeight:		20,
		FreshnessThreshold:	10,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, result.ProofVerified)
	require.Equal(t, "rpc", result.Endpoint.ServiceID)
	require.True(t, result.EndpointTTLRespected)
	require.True(t, result.TypeRegistryMatched)
	require.False(t, result.EndpointAvailabilityConsensusGuaranteed)

	_, err = BuildIdentityServiceDiscoveryV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"grpcs"},
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.ErrorContains(t, err, "failed closed")

	local, err := BuildIdentityServiceDiscoveryV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		State:			state,
		Height:			14,
		RecordTTL:		30,
		CurrentHeight:		20,
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
	})
	require.NoError(t, err)
	require.Len(t, local.FallbackEndpoints, 1)
}

func TestIdentityServiceDiscoveryV2ExternalMetadataHash(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	rpcKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata:	mustMetadataV2(t, []ResolverMetadataEntry{{Key: rpcKey, Value: "https://rpc.aet"}}),
	}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)
	metadata := `{"service":"rpc","version":"v1"}`
	hash, err := InterfaceDescriptorHashV2(metadata)
	require.NoError(t, err)
	proof.ResolverRecord.ServiceEndpoints[0].SchemaHashOptional = hash
	proof.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(proof)

	result, err := BuildIdentityServiceDiscoveryV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		ExternalMetadata:	metadata,
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, result.MetadataHashVerified)

	_, err = BuildIdentityServiceDiscoveryV2(IdentityServiceDiscoveryRequestV2{
		Name:			"alice.aet",
		ServiceID:		"rpc",
		SupportedTransports:	[]string{"https"},
		AllowedAuthPolicies:	[]string{"none"},
		ExternalMetadata:	`{"service":"rpc","version":"v2"}`,
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.ErrorContains(t, err, "hash mismatch")
}

func TestIdentityInterfaceSchemaMappingV2PolicyAndHashVerification(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata:	mustMetadataV2(t, []ResolverMetadataEntry{{Key: interfaceKey, Value: "placeholder"}}),
	}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)
	inline := `{"type":"wallet","version":"v1"}`
	hash, err := InterfaceDescriptorHashV2(inline)
	require.NoError(t, err)
	proof.ResolverRecord.InterfaceDescriptors = []InterfaceDescriptorV2{{
		InterfaceID:		"aw5",
		SchemaHash:		hash,
		SchemaInlineOptional:	inline,
		Version:		"v1",
		RenderPolicy:		"wallet_confirm",
	}}
	proof.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(proof)

	result, err := BuildIdentityInterfaceSchemaMappingV2(IdentityInterfaceSchemaRequestV2{
		Name:			"alice.aet",
		InterfaceID:		"aw5",
		ExpectedSchemaHash:	hash,
		WalletPolicy:		DefaultIdentityInterfaceWalletPolicyV2(),
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, result.ProofVerified)
	require.True(t, result.SchemaHashVerified)
	require.True(t, result.RenderPolicySupported)
	require.True(t, result.UserConfirmationRequired)
	require.True(t, result.ExecutionTargetImmutable)
	require.Equal(t, inline, result.SchemaSource)

	rejectInline := DefaultIdentityInterfaceWalletPolicyV2()
	rejectInline.AllowInlineSchemas = false
	_, err = BuildIdentityInterfaceSchemaMappingV2(IdentityInterfaceSchemaRequestV2{
		Name:			"alice.aet",
		InterfaceID:		"aw5",
		ExpectedSchemaHash:	hash,
		WalletPolicy:		rejectInline,
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.ErrorContains(t, err, "rejects inline")

	unsupportedRender := DefaultIdentityInterfaceWalletPolicyV2()
	unsupportedRender.SupportedRenderPolicies = []string{IdentityRenderPolicyReadOnlyV2}
	result, err = BuildIdentityInterfaceSchemaMappingV2(IdentityInterfaceSchemaRequestV2{
		Name:			"alice.aet",
		InterfaceID:		"aw5",
		ExpectedSchemaHash:	hash,
		WalletPolicy:		unsupportedRender,
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.False(t, result.RenderPolicySupported)

	noConfirmation := DefaultIdentityInterfaceWalletPolicyV2()
	noConfirmation.RequireUserConfirmation = false
	_, err = BuildIdentityInterfaceSchemaMappingV2(IdentityInterfaceSchemaRequestV2{
		Name:			"alice.aet",
		InterfaceID:		"aw5",
		ExpectedSchemaHash:	hash,
		WalletPolicy:		noConfirmation,
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.ErrorContains(t, err, "user confirmation")
}

func TestIdentityQueryServiceV2ResolveServiceRecordWithFallbacks(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	rpcKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	backupKey, err := ResolverMetadataServiceKey("rpc-backup")
	require.NoError(t, err)
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata: mustMetadataV2(t, []ResolverMetadataEntry{
			{Key: rpcKey, Value: "https://rpc.aet"},
			{Key: backupKey, Value: "https://backup.aet"},
		}),
	}, 12)
	require.NoError(t, err)
	service := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: state, Height: 14, DefaultTTL: 30})

	resp := service.QueryResolveServiceRecord("alice.aet", "rpc", true)
	require.Equal(t, IdentityQueryOK, resp.Code)
	require.NotNil(t, resp.Service)
	require.Len(t, resp.Services, 2)
}
