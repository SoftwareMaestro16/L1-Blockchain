package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuctionModuleBreakdownV2CoversSection134(t *testing.T) {
	breakdown, err := DefaultAuctionModuleBreakdownV2()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.NotEmpty(t, breakdown.BreakdownHash)
	require.ElementsMatch(t, requiredAuctionModuleStateObjectsV2(), breakdown.StateObjects)
	require.ElementsMatch(t, requiredAuctionModuleMessagesV2(), breakdown.Messages)
	require.ElementsMatch(t, requiredAuctionModuleQueriesV2(), breakdown.Queries)
	require.ElementsMatch(t, requiredAuctionModuleIntegrationPointsV2(), breakdown.IntegrationPoints)
	require.Contains(t, breakdown.Messages, AuctionModuleMsgCancelExpiredAuction)
	require.Contains(t, breakdown.Messages, AuctionModuleMsgClaimAuctionRefund)
	require.Contains(t, breakdown.BackingPrimitives, "FinalizeSealedAuctionFairV2")
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecAuctionsPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecAuctionsByNamePrefix)
}

func TestAuctionModuleV2DeterministicFinalizeRefundsAndFailures(t *testing.T) {
	state := EmptyIdentityState(DefaultIdentityParams())
	state, auction, err := StartSealedAuction(state, "market.aet", 20)
	require.NoError(t, err)
	params := DefaultIdentityAuctionFairnessParamsV2("aetra-local-1")
	machine, err := DescribeIdentityAuctionStateMachineV2(auction, params)
	require.NoError(t, err)
	require.Equal(t, auction.RevealStartHeight, machine.CommitEndHeight)

	leftCommit, err := ComputeAuctionCommitment(auction.Name, addr(1), 100, "left")
	require.NoError(t, err)
	rightCommit, err := ComputeAuctionCommitment(auction.Name, addr(2), 100, "right")
	require.NoError(t, err)
	hiddenCommit, err := ComputeAuctionCommitment(auction.Name, addr(3), 250, "hidden")
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(1), leftCommit, auction.CommitStartHeight)
	require.NoError(t, err)
	_, _, err = CommitAuctionBid(state, auction.Name, addr(9), leftCommit, auction.CommitStartHeight+1)
	require.ErrorContains(t, err, "duplicate commitment")
	state, _, err = CommitAuctionBid(state, auction.Name, addr(2), rightCommit, auction.CommitStartHeight+1)
	require.NoError(t, err)
	state, _, err = CommitAuctionBid(state, auction.Name, addr(3), hiddenCommit, auction.CommitStartHeight+2)
	require.NoError(t, err)

	_, _, err = RevealAuctionBid(state, auction.Name, addr(1), 100, "left", auction.RevealStartHeight-1)
	require.ErrorContains(t, err, "not in reveal phase")
	state, _, err = RevealAuctionBid(state, auction.Name, addr(2), 100, "right", auction.RevealStartHeight)
	require.NoError(t, err)
	state, _, err = RevealAuctionBid(state, auction.Name, addr(1), 100, "left", auction.RevealStartHeight+1)
	require.NoError(t, err)
	_, _, err = FinalizeAuctionModuleV2(state, auction.Name, addr(7), auction.RevealEndHeight-1, params)
	require.ErrorContains(t, err, "reveal phase is not over")

	finalState, result, err := FinalizeAuctionModuleV2(state, auction.Name, addr(7), auction.RevealEndHeight, params)
	require.NoError(t, err)
	require.Equal(t, AuctionPhaseFinalized, finalState.Auctions[0].Phase)
	require.Equal(t, addr(2), result.Winner)
	require.Len(t, result.LosingBidRefunds, 1)
	require.Len(t, result.UnrevealedForfeits, 1)
	require.NoError(t, ValidateAuctionModuleDeterministicWinnerV2(result.Auction, params))
	require.NoError(t, ValidateAuctionModuleRefundAccountingV2(result, params))

	tampered := result
	tampered.Auction.Winner = addr(9)
	require.ErrorContains(t, ValidateAuctionModuleDeterministicWinnerV2(tampered.Auction, params), string(AuctionModuleFailureNonDeterministicTie))

	v2Commit, err := ComputeAuctionCommitmentV2(auction.Name, addr(1), 100, "left", params.ChainID, params.ModuleVersion)
	require.NoError(t, err)
	require.True(t, AuctionCommitmentMatchesChainDomainV2(auction.Name, addr(1), 100, "left", params.ChainID, params.ModuleVersion, v2Commit))
	require.False(t, AuctionCommitmentMatchesChainDomainV2(auction.Name, addr(1), 100, "left", "other-chain", params.ModuleVersion, v2Commit))
}

func TestAuctionModuleBreakdownV2RejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultAuctionModuleBreakdownV2()
	require.NoError(t, err)
	missing := breakdown
	missing.Queries = missing.Queries[:len(missing.Queries)-1]
	_, err = NewAuctionModuleBreakdownV2(missing)
	require.ErrorContains(t, err, "query entries")
}

func TestProofVerificationModuleBreakdownV2CoversSection135(t *testing.T) {
	breakdown, err := DefaultProofVerificationModuleBreakdownV2()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.NotEmpty(t, breakdown.BreakdownHash)
	require.ElementsMatch(t, requiredProofModuleStateObjectsV2(), breakdown.StateObjects)
	require.ElementsMatch(t, requiredProofModuleMessagesV2(), breakdown.Messages)
	require.ElementsMatch(t, requiredProofModuleQueriesV2(), breakdown.Queries)
	require.ElementsMatch(t, requiredProofModuleIntegrationPointsV2(), breakdown.IntegrationPoints)
	require.Contains(t, breakdown.Messages, ProofModuleMsgInvalidateResolutionCache)
	require.Contains(t, breakdown.BackingPrimitives, "VerifyIdentityResolutionProofLightClientV2")
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecResolutionCachePrefix)

	schema := BuildProofModuleSchemaDescriptorV2()
	require.NoError(t, ValidateProofModuleSchemaDescriptorV2(schema))
	require.Equal(t, IdentityProofSchemaVersionV2, schema.SchemaVersion)
	require.Equal(t, IdentityResolutionProofFormatV2FieldOrder, schema.ResolutionFields)
	require.Equal(t, RecursiveResolutionProofV2FieldOrder, schema.RecursiveFields)
}

func TestProofVerificationModuleV2AssemblyCacheHeightAndLightClientFailures(t *testing.T) {
	state := resolverModuleState(t)
	rootProof, err := BuildIdentityProof(state, mustIdentityDomainStoreKeyV2("alice.aet"))
	require.NoError(t, err)
	appHash := rootProof.RootHash
	proof, err := BuildProofModuleResolutionProofV2(state, "aetra-local-1", appHash, "alice.aet", 14, 30)
	require.NoError(t, err)
	require.NoError(t, ValidateProofModuleResolutionProofV2(proof))

	badVersion := proof
	badVersion.RecordVersion++
	badVersion.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(badVersion)
	require.ErrorContains(t, ValidateProofModuleResolutionProofV2(badVersion), string(ProofModuleFailureInconsistentRecordVersions))

	target, err := VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
		ExpectedChainID:	"aetra-local-1",
		RequestedName:		"alice.aet",
		TrustedHeader:		IdentityTrustedHeaderV2{ChainID: "aetra-local-1", Height: 14, AppHash: appHash, Trusted: true},
		Proof:			proof,
		TargetType:		IdentityResolutionTargetPrimary,
		CurrentHeight:		14,
		AllowRenewalWindow:	true,
		NormalizationVersion:	NameNormalizationVersionV2,
	})
	require.NoError(t, err)
	require.Equal(t, addr(2), target.Address)

	_, err = BuildProofModuleNonExistenceProofV2(state, "aetra-local-1", appHash, "Bad..Name", 14, 30)
	require.ErrorContains(t, err, string(ProofModuleFailureMalformedNameNonExistence))
	absence, err := BuildProofModuleNonExistenceProofV2(state, "aetra-local-1", appHash, "missing.aet", 14, 30)
	require.NoError(t, err)
	require.NotNil(t, absence.NonExistenceProofOptional)

	cache, err := NewResolutionCacheRecordV2("alice.aet", identityHash("path"), identityHash("record"), 100, proof.RecordVersion, 1, 1)
	require.NoError(t, err)
	require.NoError(t, ValidateProofModuleCacheUseV2(cache, ResolutionCacheUseContextV2{
		Height:		14,
		SourceVersion:	proof.RecordVersion,
		ParentEpoch:	1,
		ChildEpoch:	1,
		LightClient:	true,
		ProofVerified:	true,
	}))
	invalidated := InvalidateResolutionCacheRecordV2ForDomainMutation(cache, proof.RecordVersion+1, 2, 2)
	require.ErrorContains(t, ValidateProofModuleCacheUseV2(invalidated, ResolutionCacheUseContextV2{
		Height:		14,
		SourceVersion:	proof.RecordVersion + 1,
		ParentEpoch:	2,
		ChildEpoch:	2,
		LightClient:	true,
		ProofVerified:	true,
	}), string(ProofModuleFailureStaleCacheAfterInvalidation))

	require.NoError(t, ValidateProofModuleHeightAvailableV2(14, 10, 20))
	require.ErrorContains(t, ValidateProofModuleHeightAvailableV2(9, 10, 20), string(ProofModuleFailureProofHeightPruned))

	recursive, err := BuildRecursiveResolutionProofV2(state, "aetra-local-1", "alice.aet", "alice.aet", 14, 30, nil)
	require.NoError(t, err)
	require.NoError(t, ValidateProofModuleRecursiveProofV2(recursive))
	brokenRecursive := recursive
	brokenRecursive.PathLabels = []string{"api", "alice"}
	brokenRecursive.PathDomainRecords = nil
	brokenRecursive.PathDelegationRecords = nil
	brokenRecursive.ProofCommitmentHash = ComputeRecursiveResolutionProofCommitmentHashV2(brokenRecursive)
	require.ErrorContains(t, ValidateProofModuleRecursiveProofV2(brokenRecursive), string(ProofModuleFailureMissingDelegationConstraint))
}

func TestProofVerificationModuleBreakdownV2RejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultProofVerificationModuleBreakdownV2()
	require.NoError(t, err)
	missing := breakdown
	missing.StateObjects = missing.StateObjects[:len(missing.StateObjects)-1]
	_, err = NewProofVerificationModuleBreakdownV2(missing)
	require.ErrorContains(t, err, "state object entries")

	schema := BuildProofModuleSchemaDescriptorV2()
	schema.ResolutionFields[0], schema.ResolutionFields[1] = schema.ResolutionFields[1], schema.ResolutionFields[0]
	require.ErrorContains(t, ValidateProofModuleSchemaDescriptorV2(schema), "field order")
}

func TestProofModuleReverseResolutionProofV2(t *testing.T) {
	state := resolverModuleState(t)
	var err error
	state, _, err = SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 14)
	require.NoError(t, err)
	state = state.Export()
	require.NoError(t, state.Validate())
	proof, err := BuildProofModuleReverseResolutionProofV2(state, "aetra-local-1", identityHash("app"), "alice.aet", 15, 30, addr(2))
	require.NoError(t, err)
	require.NotNil(t, proof.ReverseRecordOptional)
	require.Equal(t, IdentityProofQueryResolveReverse, proof.QueryType)
	require.True(t, strings.HasPrefix(proof.ReverseRecordProofOptional.Key, IdentityStoreV2SpecReversePrefix+"/"))
}
