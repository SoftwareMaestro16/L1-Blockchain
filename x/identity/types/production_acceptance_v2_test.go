package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityProductionAcceptanceV2CoversSection19(t *testing.T) {
	acceptance := DefaultIdentityProductionAcceptanceV2()
	require.NoError(t, ValidateIdentityProductionAcceptanceV2(acceptance))
	require.Len(t, acceptance.Criteria, 14)
	require.NotEmpty(t, acceptance.AcceptanceHash)

	requireAcceptanceCriteriaV2(t, acceptance, []IdentityProductionAcceptanceCriterionIDV2{
		IdentityAcceptanceOnChainLifecycleV2,
		IdentityAcceptanceAtomicRegistryNFTV2,
		IdentityAcceptanceLifecycleOperationsV2,
		IdentityAcceptanceProofVerifiableResolutionV2,
		IdentityAcceptanceLightClientVerificationV2,
		IdentityAcceptanceUnifiedResolverTargetsV2,
		IdentityAcceptancePreSigningExecutionV2,
		IdentityAcceptanceStoreV2EfficientReadsV2,
		IdentityAcceptanceBlockSTMParallelismV2,
		IdentityAcceptanceBatchResolverConflictSafeV2,
		IdentityAcceptanceCacheInvalidationV2,
		IdentityAcceptanceParameterizedEconomicsV2,
		IdentityAcceptanceReverseForwardConsistencyV2,
		IdentityAcceptanceRequiredTestsImplementedV2,
	})

	require.Contains(t, acceptanceEvidenceForCriterionV2(t, acceptance, IdentityAcceptanceProofVerifiableResolutionV2), "x/identity/types/query_v2.go:QueryResolutionProof")
	require.Contains(t, acceptanceEvidenceForCriterionV2(t, acceptance, IdentityAcceptanceRequiredTestsImplementedV2), "x/identity/types/test_coverage_requirements_v2.go:DefaultIdentityRequiredPerformanceTestCoverageV2")
	requireCoverageReferencesExistV2(t, identityProductionAcceptanceEvidenceReferencesV2(acceptance))
}

func TestIdentityProductionAcceptanceV2RejectsMissingCriterionBadEvidenceAndHashMismatch(t *testing.T) {
	acceptance := DefaultIdentityProductionAcceptanceV2()

	missing := acceptance
	missing.Criteria = cloneIdentityProductionAcceptanceCriteriaV2(acceptance.Criteria)
	missing.Criteria = missing.Criteria[:len(missing.Criteria)-1]
	missing.AcceptanceHash = ComputeIdentityProductionAcceptanceHashV2(missing)
	require.ErrorContains(t, ValidateIdentityProductionAcceptanceV2(missing), "must define")

	badEvidence := acceptance
	badEvidence.Criteria = cloneIdentityProductionAcceptanceCriteriaV2(acceptance.Criteria)
	badEvidence.Criteria[0].Evidence = []string{"not-a-go-reference"}
	badEvidence.AcceptanceHash = ComputeIdentityProductionAcceptanceHashV2(badEvidence)
	require.ErrorContains(t, ValidateIdentityProductionAcceptanceV2(badEvidence), "invalid evidence")

	unsorted := acceptance
	unsorted.Criteria = cloneIdentityProductionAcceptanceCriteriaV2(acceptance.Criteria)
	unsorted.Criteria[2].Evidence[0], unsorted.Criteria[2].Evidence[1] = unsorted.Criteria[2].Evidence[1], unsorted.Criteria[2].Evidence[0]
	unsorted.AcceptanceHash = ComputeIdentityProductionAcceptanceHashV2(unsorted)
	require.ErrorContains(t, ValidateIdentityProductionAcceptanceV2(unsorted), "must be sorted")

	tampered := acceptance
	tampered.Criteria = cloneIdentityProductionAcceptanceCriteriaV2(acceptance.Criteria)
	tampered.Criteria[4].Evidence[0] = "x/identity/types/light_client_verifier_v2.go:VerifyIdentityResolutionProofLightClientV2Tampered"
	require.ErrorContains(t, ValidateIdentityProductionAcceptanceV2(tampered), "hash mismatch")
}

func requireAcceptanceCriteriaV2(t *testing.T, acceptance IdentityProductionAcceptanceV2, want []IdentityProductionAcceptanceCriterionIDV2) {
	t.Helper()
	require.Len(t, acceptance.Criteria, len(want))
	for i, expected := range want {
		require.Equal(t, expected, acceptance.Criteria[i].ID)
		require.NotEmpty(t, acceptance.Criteria[i].Evidence)
	}
}

func acceptanceEvidenceForCriterionV2(t *testing.T, acceptance IdentityProductionAcceptanceV2, id IdentityProductionAcceptanceCriterionIDV2) []string {
	t.Helper()
	for _, criterion := range acceptance.Criteria {
		if criterion.ID == id {
			return criterion.Evidence
		}
	}
	require.FailNowf(t, "missing criterion", "%s", id)
	return nil
}

func identityProductionAcceptanceEvidenceReferencesV2(acceptance IdentityProductionAcceptanceV2) []string {
	refs := make([]string, 0)
	for _, criterion := range acceptance.Criteria {
		refs = append(refs, criterion.Evidence...)
	}
	return refs
}

func cloneIdentityProductionAcceptanceCriteriaV2(in []IdentityProductionAcceptanceCriterionV2) []IdentityProductionAcceptanceCriterionV2 {
	out := append([]IdentityProductionAcceptanceCriterionV2(nil), in...)
	for i := range out {
		out[i].Evidence = append([]string(nil), out[i].Evidence...)
	}
	return out
}
