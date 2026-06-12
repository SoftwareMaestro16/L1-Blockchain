package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIdentityProofRequirementsAndIntegrationTasksCoverSectionsEightFourFive(t *testing.T) {
	require.Len(t, IdentityProofRequirementsV2(), 8)
	seenRequirements := make(map[IdentityProofRequirementV2]struct{})
	for _, item := range IdentityProofRequirementsV2() {
		require.NotEmpty(t, item.VerifiedCondition)
		seenRequirements[item.Requirement] = struct{}{}
	}
	for _, requirement := range []IdentityProofRequirementV2{
		IdentityProofRequirementDomainOwnership,
		IdentityProofRequirementNFTBinding,
		IdentityProofRequirementDomainStatus,
		IdentityProofRequirementExpiry,
		IdentityProofRequirementResolverRecord,
		IdentityProofRequirementReverseLookup,
		IdentityProofRequirementDelegationGrants,
		IdentityProofRequirementAuctionFinality,
	} {
		_, found := seenRequirements[requirement]
		require.True(t, found, "missing requirement %s", requirement)
	}

	require.Len(t, IdentityIntegrationTasksV2(), 7)
	seenTasks := make(map[IdentityIntegrationTaskIDV2]struct{})
	for _, task := range IdentityIntegrationTasksV2() {
		require.NotEmpty(t, task.Target)
		require.NotEmpty(t, task.AcceptanceCriteria)
		seenTasks[task.TaskID] = struct{}{}
	}
	for _, task := range []IdentityIntegrationTaskIDV2{
		IdentityTaskIsolatedZone,
		IdentityTaskCrossZoneLookupMessages,
		IdentityTaskResolverProofAPIs,
		IdentityTaskVMResolverAdapter,
		IdentityTaskReverseLookupProof,
		IdentityTaskCacheInvalidationMessages,
		IdentityTaskWalletSDKHelpers,
	} {
		_, found := seenTasks[task]
		require.True(t, found, "missing task %s", task)
	}
}

func TestNativeResolverRecordInterfaceCommitsCanonicalFields(t *testing.T) {
	domain := testResolverVMDomain(t)
	resolver := testResolverVMRecord(domain)
	native, err := NewNativeResolverRecordV2(domain, resolver, IdentityLookupTargetAccount, []byte("awallet1"), "identity/resolvers/alice")
	require.NoError(t, err)

	require.Equal(t, domain.Name, native.Name)
	require.Equal(t, domain.NameHash, native.NameHash)
	require.Equal(t, ResolverRecordVersionV2(resolver), native.ResolverRecordVersion)
	require.Equal(t, domain.ExpiryHeight, native.ExpiryHeight)
	require.NoError(t, native.Validate())

	bad := native
	bad.TargetValue = nil
	bad.RecordHash = ComputeNativeResolverRecordHashV2(bad)
	require.ErrorContains(t, bad.Validate(), "target value")
}

func TestVMResolverAdapterCommitsBoundedContractOutput(t *testing.T) {
	domain := testResolverVMDomain(t)
	resolver := testResolverVMRecord(domain)
	native, ctx := testResolverVMNativeAndContext(t, domain, resolver)
	output, err := NewVMResolverContractOutputV2(VMResolverContractOutputV2{
		NameHash:		native.NameHash,
		TargetType:		native.TargetType,
		ResolvedValue:		[]byte("dynamic-wallet"),
		ResolverRecordVersion:	native.ResolverRecordVersion,
		GasUsed:		500,
		MemoryUsedBytes:	1024,
		ProofChecks:		2,
		RecursionDepth:		1,
		Status:			IdentityResolutionStatusResolved,
	})
	require.NoError(t, err)

	result, err := EvaluateVMResolverContractOutputV2(native, ctx, output, identityHash("resolver-root"), VMResolverFallbackPolicyV2{})
	require.NoError(t, err)
	require.False(t, result.UsedFallback)
	require.Equal(t, IdentityResolutionStatusResolved, result.Status)
	require.Equal(t, output.OutputHash, result.Proof.OutputHash)
	require.Equal(t, ctx.CodeID, result.Proof.CodeID)
	require.NoError(t, result.Validate())

	otherCode := result.Proof
	otherCode.CodeID++
	otherCode.ProofHash = ComputeVMResolverContractProofHashV2(otherCode)
	require.NotEqual(t, result.Proof.ProofHash, otherCode.ProofHash)
}

func TestVMResolverAdapterRejectsOwnershipOverrideAndLimitDrift(t *testing.T) {
	domain := testResolverVMDomain(t)
	resolver := testResolverVMRecord(domain)
	native, ctx := testResolverVMNativeAndContext(t, domain, resolver)

	override, err := NewVMResolverContractOutputV2(VMResolverContractOutputV2{
		NameHash:		native.NameHash,
		TargetType:		native.TargetType,
		ResolvedValue:		[]byte("dynamic-wallet"),
		ResolverRecordVersion:	native.ResolverRecordVersion,
		GasUsed:		500,
		MemoryUsedBytes:	1024,
		ProofChecks:		2,
		RecursionDepth:		1,
		OwnerOverride:		resolverVMAddr(9),
		Status:			IdentityResolutionStatusResolved,
	})
	require.NoError(t, err)
	_, err = EvaluateVMResolverContractOutputV2(native, ctx, override, identityHash("resolver-root"), VMResolverFallbackPolicyV2{})
	require.ErrorContains(t, err, "cannot override native owner")

	overGas := override
	overGas.OwnerOverride = nil
	overGas.GasUsed = ctx.Limits.GasLimit + 1
	overGas.OutputHash = ComputeVMResolverContractOutputHashV2(overGas)
	_, err = EvaluateVMResolverContractOutputV2(native, ctx, overGas, identityHash("resolver-root"), VMResolverFallbackPolicyV2{})
	require.ErrorContains(t, err, "gas limit")
}

func TestVMResolverAdapterFallsBackToNativeRecordDeterministically(t *testing.T) {
	domain := testResolverVMDomain(t)
	resolver := testResolverVMRecord(domain)
	native, ctx := testResolverVMNativeAndContext(t, domain, resolver)
	failed, err := NewVMResolverContractOutputV2(VMResolverContractOutputV2{
		NameHash:		native.NameHash,
		TargetType:		native.TargetType,
		ResolverRecordVersion:	native.ResolverRecordVersion,
		GasUsed:		700,
		MemoryUsedBytes:	2048,
		ProofChecks:		1,
		RecursionDepth:		1,
		Status:			IdentityResolutionStatusFailed,
	})
	require.NoError(t, err)

	_, err = EvaluateVMResolverContractOutputV2(native, ctx, failed, identityHash("resolver-root"), VMResolverFallbackPolicyV2{})
	require.ErrorContains(t, err, "fallback is disabled")

	result, err := EvaluateVMResolverContractOutputV2(native, ctx, failed, identityHash("resolver-root"), VMResolverFallbackPolicyV2{AllowNativeFallback: true, FallbackStatus: IdentityResolutionStatusFailed})
	require.NoError(t, err)
	require.True(t, result.UsedFallback)
	require.Equal(t, IdentityResolutionStatusFailed, result.Status)
	require.Equal(t, native.TargetValue, result.Output.ResolvedValue)
	require.Equal(t, IdentityResolutionStatusResolved, result.Output.Status)
	require.NoError(t, result.Validate())
}

func testResolverVMNativeAndContext(t *testing.T, domain DomainRecordV2, resolver ResolverRecord) (NativeResolverRecordV2, VMResolverContractContextV2) {
	t.Helper()
	native, err := NewNativeResolverRecordV2(domain, resolver, IdentityLookupTargetAccount, []byte("awallet1"), "identity/resolvers/alice")
	require.NoError(t, err)
	ctx, err := NewVMResolverContractContextV2(native, domain, 7, "contract:resolver", 20, DefaultVMResolverExecutionLimitsV2())
	require.NoError(t, err)
	return native, ctx
}

func testResolverVMDomain(t *testing.T) DomainRecordV2 {
	t.Helper()
	name := "alice.aet"
	nameHash, err := DomainRecordV2NameHash(name)
	require.NoError(t, err)
	parentHash, err := DomainRecordV2ParentNameHash(name)
	require.NoError(t, err)
	return DomainRecordV2{
		Name:			name,
		NameHash:		nameHash,
		NormalizedName:		name,
		ParentNameHash:		parentHash,
		TLD:			DomainTLD,
		Owner:			resolverVMAddr(1),
		ExpiryHeight:		100,
		NFTClassID:		DomainNFTClassID,
		NFTItemID:		"domain:alice",
		Status:			DomainRecordV2Active,
		LifecycleEpoch:		10,
		CreatedAtHeight:	10,
		UpdatedAtHeight:	12,
		Version:		1,
	}
}

func testResolverVMRecord(domain DomainRecordV2) ResolverRecord {
	return ResolverRecord{
		Domain:		domain.Name,
		Owner:		cloneSpecAddress(domain.Owner),
		Primary:	resolverVMAddr(2),
		UpdatedAtUnix:	7,
	}
}

func resolverVMAddr(id byte) sdk.AccAddress {
	return sdk.AccAddress{0xaa, id}
}
