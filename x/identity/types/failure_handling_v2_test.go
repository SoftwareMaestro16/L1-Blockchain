package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityLightClientFailureHandlingV2InvalidatesCacheAndRequiresConfirmation(t *testing.T) {
	cache, err := NewResolutionCacheRecordV2("alice.aet", identityHash("path"), identityHash("record"), 100, 7, 2, 3)
	require.NoError(t, err)
	proofErr := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrRecordStale, Message: "stale proof"}

	handling := HandleIdentityLightClientFailureV2(proofErr, &cache, false)
	require.Equal(t, IdentityLightClientErrRecordStale, handling.FailureCode)
	require.True(t, handling.RejectTargetUse)
	require.True(t, handling.RequestFreshProof)
	require.True(t, handling.DirectAddressFallbackRequiresConfirmation)
	require.False(t, handling.DirectAddressFallbackAllowed)
	require.True(t, handling.CacheInvalid)
	require.NotNil(t, handling.InvalidatedCache)
	require.Equal(t, uint64(0), handling.InvalidatedCache.ValidUntilHeight)

	confirmed := HandleIdentityLightClientFailureV2(proofErr, nil, true)
	require.True(t, confirmed.DirectAddressFallbackAllowed)
	require.False(t, confirmed.CacheInvalid)
}

func TestIdentityLightClientFailureHandlingV2DefaultsUnknownErrorsToProofInvalid(t *testing.T) {
	handling := HandleIdentityLightClientFailureV2(errors.New("transport truncated proof"), nil, false)
	require.Equal(t, IdentityLightClientErrProofInvalid, handling.FailureCode)
	require.True(t, handling.RejectTargetUse)
	require.True(t, handling.RequestFreshProof)
}

func TestIdentityWalletResolutionStatusV2DoesNotTrustFailedProofs(t *testing.T) {
	proofErr := IdentityLightClientVerificationErrorV2{Code: IdentityLightClientErrReverseNotVerified, Message: "reverse mismatch"}
	status := EvaluateIdentityWalletResolutionStatusV2(nil, proofErr, 30, 5)
	require.False(t, status.VerifiedStatusVisible)
	require.False(t, status.AutoFillTargetAllowed)
	require.True(t, status.RejectTargetUse)
	require.Equal(t, IdentityLightClientErrReverseNotVerified, status.FailureCode)
}

func TestIdentityWalletResolutionStatusV2WarnsOnOldProofHeight(t *testing.T) {
	target := IdentityLightClientVerifiedTargetV2{
		Name:		"alice.aet",
		ProofHeight:	10,
	}
	fresh := EvaluateIdentityWalletResolutionStatusV2(&target, nil, 13, 5)
	require.True(t, fresh.VerifiedStatusVisible)
	require.True(t, fresh.AutoFillTargetAllowed)
	require.False(t, fresh.FreshnessWarning)

	stale := EvaluateIdentityWalletResolutionStatusV2(&target, nil, 20, 5)
	require.True(t, stale.VerifiedStatusVisible)
	require.True(t, stale.AutoFillTargetAllowed)
	require.True(t, stale.FreshnessWarning)
}
