package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentitySendByNameV2ProofAwareMemoAndStaleWarning(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolvePrimary, 14, 30, nil)

	result, err := BuildIdentitySendByNameV2(IdentitySendByNameRequestV2{
		Name:			"alice.aet",
		CurrentHeight:		40,
		FreshnessThreshold:	5,
		IncludeAuditMemo:	true,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, result.ProofVerified)
	require.True(t, result.StaleProofWarning)
	require.Equal(t, addr(2), result.Address)
	require.Equal(t, uint64(14), result.ProofHeight)
	require.Contains(t, result.AuditMemo, "aet:v2;name=alice.aet;height=14")
	require.Contains(t, result.WalletDisplayLabel, "alice.aet ->")
}

func TestIdentityInvokeByNameV2VerifiesContractTargetAndInterface(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	inline := `{"type":"swap","version":"v1"}`
	interfaceHash, err := InterfaceDescriptorHashV2(inline)
	require.NoError(t, err)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	record := ResolverPatch{
		Primary:	addr(2),
		Records: map[string]sdk.AccAddress{
			"swap": addr(3),
		},
		Metadata: mustMetadataV2(t, []ResolverMetadataEntry{
			{Key: interfaceKey, Value: inline},
			{Key: "hint.requires_interface_confirmation", Value: "true"},
			{Key: "hint.simulation_required", Value: "true"},
		}),
	}
	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), record, 12)
	require.NoError(t, err)

	unified, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 13, 30)
	require.NoError(t, err)
	descriptor, err := VerifyIdentityInterfaceDescriptorForInvokeV2(unified, "aw5", interfaceHash)
	require.NoError(t, err)
	require.Equal(t, interfaceHash, interfaceDescriptorSchemaHashV2(*descriptor))

	result, err := BuildIdentityInvokeByNameV2(IdentityInvokeByNameRequestV2{
		Name:			"alice.aet",
		TargetID:		"swap",
		InterfaceID:		"aw5",
		ExpectedInterfaceHash:	interfaceHash,
		Method:			"execute_swap",
		State:			state,
		Height:			13,
		RecordTTL:		30,
		CurrentHeight:		20,
		FreshnessThreshold:	10,
	})
	require.NoError(t, err)
	require.Equal(t, addr(3), result.ContractAddress)
	require.True(t, result.InterfaceDescriptorVerified)
	require.True(t, result.RequiresInterfaceConfirmation)
	require.True(t, result.SimulationRequiredBeforeSigning)
	require.False(t, result.StaleInterfaceDescriptorWarning)
}

func TestIdentityInvokeByNameV2ProofAwareContractTarget(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Contract:	addr(3),
	}, 12)
	require.NoError(t, err)
	proof := buildLightClientFormatProofV2(t, state, "alice.aet", IdentityProofQueryResolveRecord, 14, 30, nil)

	result, err := BuildIdentityInvokeByNameV2(IdentityInvokeByNameRequestV2{
		Name:			"alice.aet",
		TargetID:		ResolverKeyContract,
		Method:			"execute_swap",
		CurrentHeight:		20,
		ExpectedChainID:	"aetra-local-1",
		TrustedHeader:		trustedHeaderForProofV2(proof),
		Proof:			&proof,
	})
	require.NoError(t, err)
	require.True(t, result.ProofVerified)
	require.Equal(t, addr(3), result.ContractAddress)
	require.Equal(t, "execute_swap", result.Entrypoint)
	require.Equal(t, uint64(14), result.ProofHeight)
}

func TestIdentityQueryServiceV2ResolveContractTarget(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Contract: addr(3)}, 12)
	require.NoError(t, err)
	service := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: state, Height: 13, DefaultTTL: 30})

	resp := service.QueryResolveContractTarget("alice.aet", ResolverKeyContract)
	require.Equal(t, IdentityQueryOK, resp.Code)
	require.NotNil(t, resp.ContractTarget)
	require.Equal(t, addr(3), resp.Address)
}

func mustMetadataV2(t *testing.T, entries []ResolverMetadataEntry) []byte {
	t.Helper()
	metadata, err := EncodeResolverMetadata(entries)
	require.NoError(t, err)
	return metadata
}
