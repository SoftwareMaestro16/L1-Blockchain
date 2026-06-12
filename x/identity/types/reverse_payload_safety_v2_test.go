package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReverseResolutionSafetyInvalidationHooksV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, domain.Name, addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), domain.Name, 13)
	require.NoError(t, err)

	next, invalidated, err := InvalidateReverseRecordsForResolverUpdateV2(state, domain.Name, 14, true, nil)
	require.NoError(t, err)
	require.Len(t, invalidated, 1)
	require.Empty(t, next.ReverseRecords)

	state, _, err = SetIdentityReverse(state, addr(2), addr(2), domain.Name, 14)
	require.NoError(t, err)
	next, invalidated, err = InvalidateReverseRecordsForDomainTransferV2(state, domain.Name, 15)
	require.NoError(t, err)
	require.Len(t, invalidated, 1)
	require.Empty(t, next.ReverseRecords)

	state, _, err = SetIdentityReverse(state, addr(2), addr(2), domain.Name, 15)
	require.NoError(t, err)
	for i := range state.Domains {
		if state.Domains[i].Name == domain.Name {
			state.Domains[i].ExpiryHeight = 20
		}
	}
	state = state.Export()
	next, invalidated, err = InvalidateReverseRecordsForDomainExpiryV2(state, 20)
	require.NoError(t, err)
	require.Len(t, invalidated, 1)
	require.Empty(t, next.ReverseRecords)
}

func TestVerifiedReverseProofFormatAndWalletDisplayV2(t *testing.T) {
	state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, domain.Name, addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), domain.Name, 13)
	require.NoError(t, err)
	appHash, err := IdentityStateRoot(state)
	require.NoError(t, err)

	proof, err := BuildVerifiedReverseResolutionProofV2(state, addr(2), "aetra-local-1", appHash, 14, 30, nil)
	require.NoError(t, err)
	require.Equal(t, IdentityReverseProofFormatVersionV2, proof.ProofVersion)
	require.True(t, proof.Record.Verified)
	require.NotEmpty(t, proof.ProofHash)
	require.NoError(t, ValidateVerifiedReverseResolutionProofV2(proof, IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		domain.Name,
		TrustedHeader:		trustedHeaderForProofV2(proof.Proof),
		CurrentHeight:		14,
		AllowRenewalWindow:	true,
		NormalizationVersion:	NameNormalizationVersionV2,
	}))

	display := BuildIdentityReverseWalletDisplayStateV2(state, addr(2), 14, nil)
	require.Equal(t, IdentityReverseWalletDisplayVerifiedV2, display.State)
	require.True(t, display.DisplayAsCanonical)

	claim, err := NewReverseResolutionRecordV2(addr(3), domain.Name, false, 14, domain.ExpiryHeight)
	require.NoError(t, err)
	display = BuildIdentityReverseWalletDisplayFromRecordV2(state, claim, 15, nil)
	require.Equal(t, IdentityReverseWalletDisplaySeparatedV2, display.State)
	require.True(t, display.DisplaySeparately)
	require.False(t, display.DisplayAsCanonical)
}

func TestResolverPayloadSafetyFeesLimitsAndMalformedMetadataV2(t *testing.T) {
	record := safePayloadRecordV2(t)
	params := DefaultIdentityResolverPayloadSafetyParamsV2()
	quote, err := QuoteIdentityResolverPayloadUpdateFeeV2(record, params.FreeUpdatesPerWindow, params)
	require.NoError(t, err)
	require.Equal(t, uint32(DomainDistributionDenominatorBps), quote.ChurnMultiplierBps)
	require.True(t, quote.TotalUpdateFee.GT(quote.BaseUpdateFee))
	require.Contains(t, quote.Formula, "payload_bytes")

	churned, err := QuoteIdentityResolverPayloadUpdateFeeV2(record, params.FreeUpdatesPerWindow+2, params)
	require.NoError(t, err)
	require.Greater(t, churned.ChurnMultiplierBps, quote.ChurnMultiplierBps)
	require.True(t, churned.TotalUpdateFee.GT(quote.TotalUpdateFee))

	tooLarge := record
	tooLarge.MaxPayloadBytes = EstimateUnifiedResolverPayloadBytesV2(record) - 1
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(tooLarge), "payload size")

	spam := record
	for i := 0; i <= MaxUnifiedServiceEndpoints; i++ {
		spam.ServiceEndpoints = append(spam.ServiceEndpoints, ServiceEndpointV2{
			ServiceID:	"svc" + string(rune('a'+i)),
			Endpoint:	"https://svc" + string(rune('a'+i)) + ".aet",
			ServiceType:	"service.v1",
			Transport:	"https",
			AuthPolicy:	"none",
			Priority:	uint32(i),
			Weight:		1,
			TTL:		10,
		})
	}
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(spam), "service endpoints")

	malformed := record
	malformed.InterfaceDescriptors[0].SchemaInlineOptional = strings.Repeat("x", MaxInterfaceInlineSchemaBytesV2+1)
	malformed.InterfaceDescriptors[0].SchemaHash = identityHash("not-prefixed")
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(malformed), "sha256")
}

func FuzzUnifiedResolverPayloadSafetyV2(f *testing.F) {
	f.Add("https://rpc.aet", "rpc", uint64(10), uint32(1))
	f.Add("ipfs://schema", "gateway", uint64(30), uint32(8))
	f.Fuzz(func(t *testing.T, endpoint string, serviceID string, ttl uint64, updates uint32) {
		if serviceID == "" {
			serviceID = "svc"
		}
		if ttl == 0 || ttl > 1000 {
			ttl = 10
		}
		record := safePayloadRecordV2(t)
		record.RecordTTL = ttl
		record.ServiceEndpoints[0].ServiceID = serviceID
		record.ServiceEndpoints[0].Key = serviceID
		record.ServiceEndpoints[0].Endpoint = endpoint
		record.ServiceEndpoints[0].TTL = ttl
		if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
			t.Skip()
		}
		quote, err := QuoteIdentityResolverPayloadUpdateFeeV2(record, updates, DefaultIdentityResolverPayloadSafetyParamsV2())
		require.NoError(t, err)
		require.Greater(t, quote.PayloadBytes, uint64(0))
	})
}

func safePayloadRecordV2(t *testing.T) UnifiedResolutionRecordV2 {
	t.Helper()
	inline := `{"type":"wallet","version":"v1"}`
	hash, err := InterfaceDescriptorHashV2(inline)
	require.NoError(t, err)
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		PrimaryAddress:	addr(2),
		ServiceEndpoints: []ServiceEndpointV2{{
			ServiceID:	"rpc",
			Endpoint:	"https://rpc.aet",
			ServiceType:	"service.v1",
			Transport:	"https",
			AuthPolicy:	"none",
			Priority:	1,
			Weight:		1,
			TTL:		10,
		}},
		InterfaceDescriptors: []InterfaceDescriptorV2{{
			InterfaceID:		"wallet",
			SchemaHash:		hash,
			SchemaInlineOptional:	inline,
			Version:		"v1",
			RenderPolicy:		IdentityRenderPolicyConfirmV2,
		}},
		RecordVersion:		1,
		RecordTTL:		10,
		UpdatedAtHeight:	10,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	sortUnifiedResolutionRecordV2(&record)
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
	return record
}
