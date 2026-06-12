package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityCoreModuleBreakdownV2CoversSection131(t *testing.T) {
	breakdown, err := DefaultIdentityCoreModuleBreakdownV2()
	require.NoError(t, err)
	require.NoError(t, breakdown.Validate())
	require.NotEmpty(t, breakdown.BreakdownHash)

	require.ElementsMatch(t, requiredIdentityCoreStateObjectsV2(), breakdown.StateObjects)
	require.ElementsMatch(t, requiredIdentityCoreMessagesV2(), breakdown.Messages)
	require.ElementsMatch(t, requiredIdentityCoreQueriesV2(), breakdown.Queries)
	require.ElementsMatch(t, requiredIdentityCoreIntegrationPointsV2(), breakdown.IntegrationPoints)
	require.ElementsMatch(t, requiredIdentityCoreMessagesV2(), breakdown.Keeper.MessagesHandled)
	require.ElementsMatch(t, requiredIdentityCoreQueriesV2(), breakdown.Keeper.QueriesServed)
	require.Contains(t, breakdown.Keeper.NFTHooks, "ApplyIdentityNFTTransferHookStateV2")
	require.Contains(t, breakdown.Keeper.NFTHooks, "ApplyIdentityRegistryTransferHookStateV2")
	require.Contains(t, breakdown.Keeper.NFTHooks, "RepairDomainNFTBindingInternalFailureV2")
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecDomainsPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecDomainNamesPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecCommitmentsPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecNFTBindingsPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecNFTBindingsByNamePrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecExpiryIndexPrefix)
	require.Contains(t, breakdown.StoreKeys, IdentityStoreV2SpecOwnerIndexPrefix)

	examples, err := IdentityCoreStoreV2KeyExamplesV2("alice.aet", addr(1), identityHash("registration-commitment"))
	require.NoError(t, err)
	require.Len(t, examples, 7)
	for _, key := range examples {
		require.True(t, strings.HasPrefix(key, IdentityStoreV2Prefix+"/"), key)
	}
}

func TestIdentityCoreModuleBreakdownV2RejectsMissingSurface(t *testing.T) {
	breakdown, err := DefaultIdentityCoreModuleBreakdownV2()
	require.NoError(t, err)

	missingMessage := breakdown
	missingMessage.Messages = missingMessage.Messages[:len(missingMessage.Messages)-1]
	_, err = NewIdentityCoreModuleBreakdownV2(missingMessage)
	require.ErrorContains(t, err, "message entries")

	missingQuery := breakdown
	missingQuery.Queries = missingQuery.Queries[:len(missingQuery.Queries)-1]
	_, err = NewIdentityCoreModuleBreakdownV2(missingQuery)
	require.ErrorContains(t, err, "query entries")

	missingState := breakdown
	missingState.StateObjects = missingState.StateObjects[:len(missingState.StateObjects)-1]
	_, err = NewIdentityCoreModuleBreakdownV2(missingState)
	require.ErrorContains(t, err, "state object entries")
}

func TestIdentityCoreInvariantReportV2CatchesCoreFailureModes(t *testing.T) {
	state := validIdentityCoreState(t)
	report, err := ValidateIdentityCoreModuleInvariantsV2(state, 20)
	require.NoError(t, err)
	require.True(t, report.Valid, report.Issues)
	require.NotEmpty(t, report.ReportHash)
	require.Len(t, report.DerivedIndexes.ExpiryIndex, 1)
	require.Len(t, report.DerivedIndexes.OwnerIndex, 1)

	missingExpiry := report.DerivedIndexes
	missingExpiry.ExpiryIndex = nil
	indexReport := ValidateIdentityCoreStoreV2IndexesV2(state, missingExpiry)
	require.False(t, indexReport.Valid)
	require.Contains(t, strings.Join(indexReport.Issues, "\n"), string(IdentityCoreFailureMissingExpiryIndex))

	brokenBinding := state.Clone()
	brokenBinding.DomainNFTs[0].Owner = addr(9)
	bindingReport, err := ValidateIdentityCoreModuleInvariantsV2(brokenBinding, 20)
	require.NoError(t, err)
	require.False(t, bindingReport.Valid)
	require.Contains(t, strings.Join(bindingReport.Issues, "\n"), string(IdentityCoreFailureNFTOwnerMismatch))

	replayedCommitment := state.Clone()
	replayedCommitment.UsedCommitments = append(replayedCommitment.UsedCommitments, replayedCommitment.UsedCommitments[0])
	replayReport, err := ValidateIdentityCoreModuleInvariantsV2(replayedCommitment, 20)
	require.NoError(t, err)
	require.False(t, replayReport.Valid)
	require.Contains(t, strings.Join(replayReport.Issues, "\n"), string(IdentityCoreFailureCommitmentReplay))

	resolverRace := state.Clone()
	resolverRace.Domains[0].Owner = addr(7)
	resolverRace.DomainNFTs[0].Owner = addr(7)
	resolverRace.PendingResolverUpdates = []ResolverUpdateIntent{{Domain: "alice.aet", Actor: addr(1), Nonce: 1}}
	raceReport, err := ValidateIdentityCoreModuleInvariantsV2(resolverRace, 20)
	require.NoError(t, err)
	require.False(t, raceReport.Valid)
	require.Contains(t, strings.Join(raceReport.Issues, "\n"), string(IdentityCoreFailureTransferResolverRace))
}

func TestIdentityCoreLifecycleKeeperContractV2HooksAndTransitions(t *testing.T) {
	state := validIdentityCoreState(t)
	state.PendingResolverUpdates = []ResolverUpdateIntent{{Domain: "alice.aet", Actor: addr(1), Nonce: 1}}
	next, transferred, err := ApplyIdentityNFTTransferHookStateV2(state, state.DomainNFTs[0].ID, addr(7), 21, IdentityNFTTransferHookUpdateRegistryV2)
	require.NoError(t, err)
	require.Equal(t, addr(7), transferred.Owner)
	require.Empty(t, next.PendingResolverUpdates)
	require.NoError(t, next.Validate())

	record, err := NewDomainRecordV2FromDomain(transferred, DomainRecordV2Active, 0, 21)
	require.NoError(t, err)
	_, err = ValidateIdentityCoreLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventRevealRegistration,
		Height:	22,
	})
	require.ErrorContains(t, err, string(IdentityCoreFailureInvalidLifecycleChange))
}

func validIdentityCoreState(t *testing.T) IdentityState {
	t.Helper()
	state := EmptyIdentityState(DefaultIdentityParams())
	commitment, err := ComputeRegistrationCommitment("alice.aet", addr(1), "salt")
	require.NoError(t, err)
	state, err = CommitDomainRegistration(state, "alice.aet", addr(1), commitment, 10)
	require.NoError(t, err)
	state, _, err = RevealRegisterDomain(state, "alice.aet", addr(1), "salt", 11)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	return state
}
