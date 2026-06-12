package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityWalletFallbackStrategyV2OrderedAndNoSilentDowngrade(t *testing.T) {
	cache := walletUXCacheMetadata(t, "alice.aet", 7, 20, 30, 100, 10)
	err := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrRecordStale, Message: "stale proof"}

	decision := EvaluateIdentityWalletFallbackStrategyV2(err, &cache, IdentityWalletFallbackPolicyV2{
		CurrentHeight:		24,
		FreshnessThreshold:	10,
		AllowVerifiedCache:	true,
		UserAllowsCache:	true,
		DirectAddressConfirmed:	false,
	})
	require.Equal(t, IdentityWalletFailureKindStaleProofV2, decision.FailureKind)
	require.Equal(t, IdentityLightClientErrRecordStale, decision.FailureCode)
	require.Equal(t, []IdentityWalletFallbackActionV2{
		IdentityWalletFallbackFreshProofOtherNodeV2,
		IdentityWalletFallbackNewerTrustedHeightV2,
		IdentityWalletFallbackUseVerifiedCacheV2,
		IdentityWalletFallbackExplicitAddressV2,
	}, decision.OrderedActions)
	require.True(t, decision.UseVerifiedCache)
	require.True(t, decision.NoSilentDowngrade)
	require.True(t, decision.RequiresExplicitAddressEntry)
	require.False(t, decision.DirectAddressFallbackAllowed)

	expiredCache := walletUXCacheMetadata(t, "alice.aet", 7, 20, 5, 100, 5)
	decision = EvaluateIdentityWalletFallbackStrategyV2(err, &expiredCache, IdentityWalletFallbackPolicyV2{
		CurrentHeight:		40,
		FreshnessThreshold:	5,
		AllowVerifiedCache:	true,
		UserAllowsCache:	true,
	})
	require.False(t, decision.UseVerifiedCache)
	require.Contains(t, decision.CacheRejectedReason, "stale")
}

func TestIdentityWalletFallbackStrategyV2ClassifiesProofFailures(t *testing.T) {
	cases := []struct {
		code	IdentityLightClientFailureCodeV2
		kind	IdentityWalletProofFailureKindV2
	}{
		{IdentityLightClientErrRecordStale, IdentityWalletFailureKindStaleProofV2},
		{IdentityLightClientErrDomainNotFound, IdentityWalletFailureKindMissingProofV2},
		{IdentityLightClientErrProofHeightUntrusted, IdentityWalletFailureKindUntrustedProofV2},
		{IdentityLightClientErrProofInvalid, IdentityWalletFailureKindInvalidProofV2},
	}
	for _, tc := range cases {
		err := IdentityLightClientVerificationErrorV2{Code: tc.code, Message: string(tc.kind)}
		decision := EvaluateIdentityWalletFallbackStrategyV2(err, nil, IdentityWalletFallbackPolicyV2{})
		require.Equal(t, tc.kind, decision.FailureKind)
		require.Equal(t, []IdentityWalletFallbackActionV2{
			IdentityWalletFallbackFreshProofOtherNodeV2,
			IdentityWalletFallbackNewerTrustedHeightV2,
			IdentityWalletFallbackExplicitAddressV2,
		}, decision.OrderedActions)
	}
}

func TestIdentityWalletResolutionUXV2StatesAndSendOverrideAudit(t *testing.T) {
	target := &IdentityLightClientVerifiedTargetV2{
		Name:			"alice.aet",
		NameHash:		identityHash("alice"),
		TargetType:		IdentityResolutionTargetPrimary,
		Address:		addr(2),
		RecordVersion:		7,
		FreshUntilHeight:	50,
		ProofHeight:		20,
	}
	cache := walletUXCacheMetadata(t, "alice.aet", 7, 20, 30, 100, 10)
	current, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:		"alice.aet",
		Target:		target,
		CacheMetadata:	&cache,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationSendByNameV2,
			CurrentHeight:		24,
			FreshnessThreshold:	10,
		},
	})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateVerifiedCurrentV2, current.State)
	require.True(t, current.CanSendByName)
	require.False(t, current.ExplicitOverrideRequired)
	require.NotNil(t, current.CacheMetadata)
	require.Contains(t, current.CacheMetadata.AdvancedDetails, "proof_height=20")

	stale, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:		"alice.aet",
		Target:		target,
		CacheMetadata:	&cache,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationSendByNameV2,
			CurrentHeight:		40,
			FreshnessThreshold:	10,
		},
	})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateVerifiedStaleV2, stale.State)
	require.True(t, stale.StaleProofWarning)
	require.False(t, stale.CanSendByName)
	require.True(t, stale.ExplicitOverrideRequired)

	proofErr := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrProofInvalid, Message: "bad proof"}
	override, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:			"alice.aet",
		ProofError:		proofErr,
		CacheMetadata:		&cache,
		DirectAddress:		addr(9),
		OverrideReason:		"user entered address after proof failure",
		OverrideCreatedHeight:	42,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationSendByNameV2,
			CurrentHeight:		42,
			DirectAddressConfirmed:	true,
		},
	})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateProofFailedV2, override.State)
	require.True(t, override.CanSendByName)
	require.False(t, override.ExplicitOverrideRequired)
	require.NotNil(t, override.OverrideAudit)
	require.Equal(t, IdentityWalletStateProofFailedV2, override.OverrideAudit.OriginalState)
	require.Equal(t, IdentityLightClientErrProofInvalid, override.OverrideAudit.FailureCode)
	require.NotEmpty(t, override.OverrideAudit.AuditHash)
}

func TestIdentityWalletResolutionUXV2DisplayStatesAndInvokeRules(t *testing.T) {
	target := &IdentityLightClientVerifiedTargetV2{Name: "alice.aet", RecordVersion: 7, ProofHeight: 20, FreshUntilHeight: 50}

	ownErr := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrNFTBindingMismatch, Message: "nft mismatch"}
	ownership, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{Name: "alice.aet", ProofError: ownErr})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateOwnershipMismatchV2, ownership.State)

	resolverErr := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrResolverNotFound, Message: "missing resolver"}
	resolver, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{Name: "alice.aet", ProofError: resolverErr})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateResolverMissingV2, resolver.State)

	expiredErr := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrDomainExpired, Message: "expired"}
	expired, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{Name: "alice.aet", ProofError: expiredErr})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateExpiredV2, expired.State)

	invoke, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:	"alice.aet",
		Target:	target,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationInvokeByNameV2,
			CurrentHeight:		21,
			FreshnessThreshold:	10,
			ContractTargetVerified:	true,
			InterfaceProofRequired:	true,
			InterfaceProofVerified:	true,
		},
	})
	require.NoError(t, err)
	require.Equal(t, IdentityWalletStateVerifiedCurrentV2, invoke.State)
	require.True(t, invoke.CanInvokeByName)
	require.False(t, invoke.AutoGeneratedUIDisabled)

	interfaceFailed, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:	"alice.aet",
		Target:	target,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationInvokeByNameV2,
			CurrentHeight:		21,
			FreshnessThreshold:	10,
			ContractTargetVerified:	true,
			InterfaceProofRequired:	true,
			InterfaceProofVerified:	false,
		},
	})
	require.NoError(t, err)
	require.False(t, interfaceFailed.CanInvokeByName)
	require.True(t, interfaceFailed.AutoGeneratedUIDisabled)

	reverse, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:	"alice.aet",
		Target:	target,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationReverseLookupV2,
			CurrentHeight:		21,
			FreshnessThreshold:	10,
			ReverseLookup:		true,
		},
	})
	require.NoError(t, err)
	require.True(t, reverse.ReverseCanonical)
}

func TestIdentityWalletProofCacheMetadataV2FormatAndServiceDetails(t *testing.T) {
	cache := walletUXCacheMetadata(t, "alice.aet", 9, 30, 25, 80, 15)
	meta, err := BuildIdentityWalletProofCacheMetadataV2(cache, 15)
	require.NoError(t, err)
	require.Equal(t, uint64(30), meta.ProofHeight)
	require.Equal(t, uint64(9), meta.RecordVersion)
	require.Equal(t, uint64(25), meta.ResolverTTL)
	require.Equal(t, uint64(45), meta.FreshUntilHeight)
	require.Contains(t, meta.FormattedKey, string(IdentityCacheLayerWalletVerifiedV2))
	require.Contains(t, meta.AdvancedDetails, "ttl=25")

	ux, err := BuildIdentityWalletResolutionUXV2(IdentityWalletResolutionUXRequestV2{
		Name:		"alice.aet",
		Target:		&IdentityLightClientVerifiedTargetV2{Name: "alice.aet", RecordVersion: 9, ProofHeight: 30, FreshUntilHeight: 45},
		CacheMetadata:	&cache,
		Policy: IdentityWalletUXPolicyV2{
			Operation:		IdentityWalletOperationServiceLookupV2,
			CurrentHeight:		34,
			FreshnessThreshold:	15,
			ServiceEndpoint:	true,
		},
	})
	require.NoError(t, err)
	require.Contains(t, ux.ServiceAdvancedDetails, "proof_height=30")
	require.Contains(t, ux.ServiceAdvancedDetails, "ttl=25")
}

func walletUXCacheMetadata(t *testing.T, name string, version uint64, proofHeight uint64, ttl uint64, domainExpiry uint64, freshness uint64) IdentityVerifiedCacheMetadataV2 {
	t.Helper()
	key, err := NewIdentityResolutionCacheKeyV2(IdentityCacheLayerWalletVerifiedV2, name, version, proofHeight, "", ResolverKeyPrimary)
	require.NoError(t, err)
	header := IdentityTrustedHeaderV2{ChainID: "aetheris-local-1", Height: proofHeight, AppHash: identityHash("wallet-ux-app", name), Trusted: true}
	meta, err := NewIdentityVerifiedCacheMetadataV2(key, proofHeight, header, ttl, domainExpiry, freshness, true)
	require.NoError(t, err)
	return meta
}
