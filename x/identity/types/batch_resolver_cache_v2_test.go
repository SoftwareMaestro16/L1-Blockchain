package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchResolverUpdatesV2AtomicAndPartialResults(t *testing.T) {
	owner := addr(11)
	state, _ := registerSpecDomain(t, "alice", owner, "salt-a", 10)
	state, _ = registerSpecDomainInState(t, state, "bob", owner, "salt-b", 10)
	state, _ = registerSpecDomainInState(t, state, "carol", addr(99), "salt-c", 10)

	msg := MsgBatchUpdateResolversV2{
		Auth:	txAuth(IdentitySignerScopeBatchAdmin, 1),
		Updates: []ResolverBatchUpdateV2{
			{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Primary: addr(21)}, ExpectedRecordVersion: 1, RecordTTL: 30},
			{Name: "carol.aet", NameHash: mustDomainHashV2(t, "carol.aet"), Patch: ResolverPatch{Primary: addr(22)}, ExpectedRecordVersion: 1, RecordTTL: 30},
			{Name: "bob.aet", NameHash: mustDomainHashV2(t, "bob.aet"), Patch: ResolverPatch{Primary: addr(23)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		},
	}
	msg.Auth.Signer = owner
	atomicNext, atomicResp, err := ExecuteBatchResolverUpdatesV2(state, msg, IdentityBatchResolverUpdateOptionsV2{
		Mode:		IdentityBatchFailureAtomicV2,
		Height:		20,
		GasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		GasLimit:	MinIdentityBatchResolverUpdateGasV2 * 3,
	})
	require.ErrorContains(t, err, "requires owner")
	require.Equal(t, state.Export(), atomicNext)
	require.Equal(t, uint32(1), atomicResp.Successes)
	require.Equal(t, uint32(1), atomicResp.Failures)

	partialNext, partialResp, err := ExecuteBatchResolverUpdatesV2(state, msg, IdentityBatchResolverUpdateOptionsV2{
		Mode:		IdentityBatchFailurePartialV2,
		Height:		20,
		GasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		GasLimit:	MinIdentityBatchResolverUpdateGasV2 * 3,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(2), partialResp.Successes)
	require.Equal(t, uint32(1), partialResp.Failures)
	require.Equal(t, "input_index", partialResp.ResultOrder)
	require.Equal(t, IdentityBatchUpdateSuccessV2, partialResp.Results[0].Status)
	require.Equal(t, IdentityBatchUpdateUnauthorizedV2, partialResp.Results[1].Status)
	require.Equal(t, IdentityBatchUpdateSuccessV2, partialResp.Results[2].Status)
	require.NoError(t, ValidateBatchResolverUpdateResponseV2(partialResp))
	_, found := findResolver(partialNext, "alice.aet")
	require.True(t, found)
	_, found = findResolver(partialNext, "bob.aet")
	require.True(t, found)
	_, found = findResolver(partialNext, "carol.aet")
	require.False(t, found)
}

func TestBatchResolverUpdatesV2VersionGasAndLimits(t *testing.T) {
	owner := addr(11)
	state, _ := registerSpecDomain(t, "alice", owner, "salt-a", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", owner, ResolverPatch{Primary: addr(20)}, 20)
	require.NoError(t, err)
	msg := MsgBatchUpdateResolversV2{
		Auth:	txAuth(IdentitySignerScopeBatchAdmin, 1),
		Updates: []ResolverBatchUpdateV2{
			{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Primary: addr(21)}, ExpectedRecordVersion: 2, RecordTTL: 30},
			{Name: "alice.aet", NameHash: mustDomainHashV2(t, "alice.aet"), Patch: ResolverPatch{Contract: addr(22)}, ExpectedRecordVersion: 1, RecordTTL: 30},
		},
	}
	msg.Auth.Signer = owner
	require.ErrorContains(t, msg.ValidateBasic(), "duplicate domain")

	msg.Updates = msg.Updates[:1]
	next, resp, err := ExecuteBatchResolverUpdatesV2(state, msg, IdentityBatchResolverUpdateOptionsV2{
		Mode:		IdentityBatchFailurePartialV2,
		Height:		21,
		GasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		GasLimit:	MinIdentityBatchResolverUpdateGasV2,
	})
	require.NoError(t, err)
	require.Equal(t, state.Export(), next)
	require.Equal(t, IdentityBatchUpdateVersionErrorV2, resp.Results[0].Status)
	require.NoError(t, ValidateBatchResolverUpdateResponseV2(resp))

	msg.Updates[0].ExpectedRecordVersion = 1
	next, resp, err = ExecuteBatchResolverUpdatesV2(state, msg, IdentityBatchResolverUpdateOptionsV2{
		Mode:		IdentityBatchFailurePartialV2,
		Height:		21,
		GasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		GasLimit:	MinIdentityBatchResolverUpdateGasV2 - 1,
	})
	require.NoError(t, err)
	require.Equal(t, state.Export(), next)
	require.Equal(t, IdentityBatchUpdateGasErrorV2, resp.Results[0].Status)

	oversized := MsgBatchUpdateResolversV2{Auth: txAuth(IdentitySignerScopeBatchAdmin, 1)}
	for i := 0; i <= MaxIdentityTxBatchResolverUpdatesV2; i++ {
		name := benchmarkIdentityName(i)
		oversized.Updates = append(oversized.Updates, ResolverBatchUpdateV2{Name: name, NameHash: mustDomainHashV2(t, name), Patch: ResolverPatch{Primary: addr(byte(i + 1))}, ExpectedRecordVersion: 1, RecordTTL: 30})
	}
	require.ErrorContains(t, oversized.ValidateBasic(), "must not exceed")
}

func TestIdentityResolutionCacheV2KeyMetadataFreshnessAndInvalidation(t *testing.T) {
	key, err := NewIdentityResolutionCacheKeyV2(IdentityCacheLayerWalletVerifiedV2, "alice.aet", 7, 20, identityHash("path"), ResolverKeyPrimary)
	require.NoError(t, err)
	formatted, err := FormatIdentityResolutionCacheKeyV2(key)
	require.NoError(t, err)
	require.Contains(t, formatted, key.NameHash)
	require.Contains(t, formatted, "00000000000000000007")

	header := IdentityTrustedHeaderV2{ChainID: "aetra-local-1", Height: 20, AppHash: identityHash("app"), Trusted: true}
	metadata, err := NewIdentityVerifiedCacheMetadataV2(key, 20, header, 10, 25, 20, true)
	require.NoError(t, err)
	require.Equal(t, uint64(25), metadata.FreshUntilHeight)
	require.NoError(t, ValidateIdentityVerifiedCacheFreshnessV2(metadata, 24, 6, true))
	require.ErrorContains(t, ValidateIdentityVerifiedCacheFreshnessV2(metadata, 27, 6, true), "stale")
	require.ErrorContains(t, ValidateIdentityVerifiedCacheFreshnessV2(metadata, 27, 20, false), "stale")

	badHeader := header
	badHeader.Trusted = false
	_, err = NewIdentityVerifiedCacheMetadataV2(key, 20, badHeader, 10, 25, 20, true)
	require.ErrorContains(t, err, "light-client cache must be verified")

	cache, err := NewResolutionCacheRecordV2("alice.aet", identityHash("path"), identityHash("record"), 100, 7, 2, 3)
	require.NoError(t, err)
	other, err := NewResolutionCacheRecordV2("bob.aet", identityHash("path-bob"), identityHash("record-bob"), 100, 7, 2, 3)
	require.NoError(t, err)
	for _, trigger := range []IdentityCacheInvalidationTriggerV2{
		IdentityCacheInvalidDomainTransferV2,
		IdentityCacheInvalidResolverUpdateV2,
		IdentityCacheInvalidNFTBindingUpdateV2,
		IdentityCacheInvalidDomainExpiryV2,
		IdentityCacheInvalidRenewalEpochV2,
		IdentityCacheInvalidDelegationUpdateV2,
		IdentityCacheInvalidZonePolicyUpdateV2,
		IdentityCacheInvalidReverseUpdateV2,
	} {
		invalidated, err := InvalidateIdentityResolutionCachesV2([]ResolutionCacheRecordV2{cache, other}, IdentityCacheInvalidationEventV2{
			Trigger:	trigger,
			NameHash:	cache.NameHash,
			RecordVersion:	8,
			Height:		30,
			ParentEpoch:	4,
			ChildEpoch:	5,
		})
		require.NoError(t, err, trigger)
		require.Equal(t, uint64(0), invalidated[0].ValidUntilHeight)
		require.Equal(t, uint64(100), invalidated[1].ValidUntilHeight)
	}
}

func registerSpecDomainInState(t *testing.T, state IdentityState, label string, owner []byte, salt string, height uint64) (IdentityState, Domain) {
	t.Helper()
	commitment, err := ComputeRegistrationCommitment(label+".aet", owner, salt)
	require.NoError(t, err)
	next, err := CommitDomainRegistration(state, label+".aet", owner, commitment, height)
	require.NoError(t, err)
	next, domain, err := RevealRegisterDomain(next, label+".aet", owner, salt, height+1)
	require.NoError(t, err)
	return next, domain
}
