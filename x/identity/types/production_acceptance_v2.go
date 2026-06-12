package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityProductionAcceptanceCriterionIDV2 string

const (
	IdentityAcceptanceOnChainLifecycleV2		IdentityProductionAcceptanceCriterionIDV2	= "x_identity_v2_owns_domain_lifecycle_on_chain"
	IdentityAcceptanceAtomicRegistryNFTV2		IdentityProductionAcceptanceCriterionIDV2	= "registry_and_nft_ownership_are_atomically_consistent"
	IdentityAcceptanceLifecycleOperationsV2		IdentityProductionAcceptanceCriterionIDV2	= "registration_renewal_transfer_resolver_update_subdomain_delegation_supported"
	IdentityAcceptanceProofVerifiableResolutionV2	IdentityProductionAcceptanceCriterionIDV2	= "direct_and_recursive_resolution_are_proof_verifiable"
	IdentityAcceptanceLightClientVerificationV2	IdentityProductionAcceptanceCriterionIDV2	= "light_clients_verify_status_ownership_resolver_reverse_consistency"
	IdentityAcceptanceUnifiedResolverTargetsV2	IdentityProductionAcceptanceCriterionIDV2	= "unified_resolver_supports_all_target_types"
	IdentityAcceptancePreSigningExecutionV2		IdentityProductionAcceptanceCriterionIDV2	= "send_by_name_and_invoke_by_name_resolve_before_signing"
	IdentityAcceptanceStoreV2EfficientReadsV2	IdentityProductionAcceptanceCriterionIDV2	= "store_v2_supports_efficient_direct_and_recursive_reads"
	IdentityAcceptanceBlockSTMParallelismV2		IdentityProductionAcceptanceCriterionIDV2	= "blockstm_parallelizes_independent_identity_updates"
	IdentityAcceptanceBatchResolverConflictSafeV2	IdentityProductionAcceptanceCriterionIDV2	= "batched_resolver_updates_are_versioned_bounded_conflict_safe"
	IdentityAcceptanceCacheInvalidationV2		IdentityProductionAcceptanceCriterionIDV2	= "caches_invalidate_on_ownership_resolver_delegation_expiry_zone_policy_changes"
	IdentityAcceptanceParameterizedEconomicsV2	IdentityProductionAcceptanceCriterionIDV2	= "economic_security_models_are_parameterized"
	IdentityAcceptanceReverseForwardConsistencyV2	IdentityProductionAcceptanceCriterionIDV2	= "reverse_resolution_requires_forward_consistency"
	IdentityAcceptanceRequiredTestsImplementedV2	IdentityProductionAcceptanceCriterionIDV2	= "proof_lifecycle_ownership_resolver_delegation_auction_performance_tests_implemented"
)

type IdentityProductionAcceptanceCriterionV2 struct {
	ID		IdentityProductionAcceptanceCriterionIDV2
	Evidence	[]string
}

type IdentityProductionAcceptanceV2 struct {
	Criteria	[]IdentityProductionAcceptanceCriterionV2
	AcceptanceHash	string
}

func DefaultIdentityProductionAcceptanceV2() IdentityProductionAcceptanceV2 {
	acceptance := IdentityProductionAcceptanceV2{
		Criteria: []IdentityProductionAcceptanceCriterionV2{
			{ID: IdentityAcceptanceOnChainLifecycleV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/spec_state.go:RevealRegisterDomain"}},
			{ID: IdentityAcceptanceAtomicRegistryNFTV2, Evidence: []string{"x/identity/types/nft_binding.go:TransferDomainNFTBindingAtomic", "x/identity/types/validation_v2.go:TransferDomainNFTBindingWithInvariantsV2"}},
			{ID: IdentityAcceptanceLifecycleOperationsV2, Evidence: []string{"x/identity/types/lifecycle.go:RenewDomain", "x/identity/types/spec_state.go:IssueSubdomain", "x/identity/types/spec_state.go:RevealRegisterDomain", "x/identity/types/spec_state.go:SetIdentityResolver", "x/identity/types/spec_state.go:TransferDomainNFT"}},
			{ID: IdentityAcceptanceProofVerifiableResolutionV2, Evidence: []string{"x/identity/types/light_client_verifier_v2.go:VerifyIdentityResolutionProofLightClientV2", "x/identity/types/query_v2.go:QueryRecursiveResolutionProof", "x/identity/types/query_v2.go:QueryResolutionProof"}},
			{ID: IdentityAcceptanceLightClientVerificationV2, Evidence: []string{"x/identity/types/light_client_verifier_v2.go:VerifyIdentityResolutionProofLightClientV2", "x/identity/types/proof_v2_test.go:TestIdentityLightClientProofV2VerifiesAllResolutionObjectives"}},
			{ID: IdentityAcceptanceUnifiedResolverTargetsV2, Evidence: []string{"x/identity/types/resolution_v2.go:BuildUnifiedResolutionRecordV2", "x/identity/types/resolution_v2.go:ValidateUnifiedResolutionRecordV2"}},
			{ID: IdentityAcceptancePreSigningExecutionV2, Evidence: []string{"x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKBuildInvokeByNameTxV2", "x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKBuildSendByNameTxV2", "x/identity/types/execution_integration_v2.go:BuildIdentityInvokeByNameV2", "x/identity/types/execution_integration_v2.go:BuildIdentitySendByNameV2"}},
			{ID: IdentityAcceptanceStoreV2EfficientReadsV2, Evidence: []string{"x/identity/types/storev2.go:IdentityStoreV2SpecDirectResolverReadAccessSet", "x/identity/types/storev2.go:IdentityStoreV2SpecResolutionProofReadAccessSet", "x/identity/types/storev2_spec_test.go:TestIdentityStoreV2SpecPerformanceAccessPatterns"}},
			{ID: IdentityAcceptanceBlockSTMParallelismV2, Evidence: []string{"x/identity/types/blockstm_v2.go:IdentityBlockSTMAccessSetV2", "x/identity/types/blockstm_v2.go:IdentityBlockSTMConflictClassifyV2", "x/identity/types/blockstm_v2_test.go:TestIdentityBlockSTMParallelSafeMessageClassesV2"}},
			{ID: IdentityAcceptanceBatchResolverConflictSafeV2, Evidence: []string{"x/identity/types/batch_resolver_cache_v2_test.go:TestBatchResolverUpdatesV2VersionGasAndLimits", "x/identity/types/batch_resolver_v2.go:ExecuteBatchResolverUpdatesV2", "x/identity/types/batch_resolver_v2.go:ValidateBatchResolverUpdateResponseV2"}},
			{ID: IdentityAcceptanceCacheInvalidationV2, Evidence: []string{"x/identity/types/resolution_cache_v2.go:InvalidateIdentityResolutionCachesV2", "x/identity/types/resolution_cache_v2.go:ValidateIdentityCacheInvalidationEventV2", "x/identity/types/resolution_cache_v2_test.go:TestResolutionCacheRecordV2InvalidatesOnDomainMutation"}},
			{ID: IdentityAcceptanceParameterizedEconomicsV2, Evidence: []string{"x/identity/types/anti_squatting_v2.go:QuoteIdentityRegistrationPriceV2", "x/identity/types/cost_models_v2.go:QuoteIdentityResolverUpdateFeeV2", "x/identity/types/cost_models_v2.go:QuoteIdentitySubdomainCreationFeeV2", "x/identity/types/economics_v2.go:QuoteIdentityDomainPriceV2", "x/identity/types/spam_fairness_v2.go:EstimateIdentitySpamCostV2", "x/identity/types/spam_fairness_v2.go:FinalizeSealedAuctionFairV2"}},
			{ID: IdentityAcceptanceReverseForwardConsistencyV2, Evidence: []string{"x/identity/types/resolution_v2.go:ValidateReverseResolutionRecordV2", "x/identity/types/reverse_safety_v2.go:BuildVerifiedReverseResolutionProofV2"}},
			{ID: IdentityAcceptanceRequiredTestsImplementedV2, Evidence: []string{"x/identity/types/test_coverage_requirements_v2.go:DefaultIdentityRequiredFuzzTestCoverageV2", "x/identity/types/test_coverage_requirements_v2.go:DefaultIdentityRequiredIntegrationTestCoverageV2", "x/identity/types/test_coverage_requirements_v2.go:DefaultIdentityRequiredInvariantTestCoverageV2", "x/identity/types/test_coverage_requirements_v2.go:DefaultIdentityRequiredPerformanceTestCoverageV2", "x/identity/types/test_coverage_requirements_v2.go:DefaultIdentityRequiredUnitTestCoverageV2"}},
		},
	}
	acceptance.AcceptanceHash = ComputeIdentityProductionAcceptanceHashV2(acceptance)
	return acceptance
}

func ValidateIdentityProductionAcceptanceV2(acceptance IdentityProductionAcceptanceV2) error {
	required := IdentityProductionAcceptanceCriteriaV2()
	if len(acceptance.Criteria) != len(required) {
		return fmt.Errorf("identity v2 production acceptance must define %d criteria", len(required))
	}
	for i, criterionID := range required {
		if acceptance.Criteria[i].ID != criterionID {
			return fmt.Errorf("identity v2 production acceptance criterion %d must be %s", i, criterionID)
		}
		if err := validateIdentityProductionAcceptanceEvidenceV2(acceptance.Criteria[i]); err != nil {
			return err
		}
	}
	if acceptance.AcceptanceHash == "" || acceptance.AcceptanceHash != ComputeIdentityProductionAcceptanceHashV2(acceptance) {
		return errors.New("identity v2 production acceptance hash mismatch")
	}
	return nil
}

func ComputeIdentityProductionAcceptanceHashV2(acceptance IdentityProductionAcceptanceV2) string {
	parts := []string{"identity-production-acceptance-v2"}
	criteria := append([]IdentityProductionAcceptanceCriterionV2(nil), acceptance.Criteria...)
	sort.Slice(criteria, func(i, j int) bool { return criteria[i].ID < criteria[j].ID })
	for _, criterion := range criteria {
		parts = append(parts, string(criterion.ID))
		parts = append(parts, sortedBreakdownStringsV2(criterion.Evidence)...)
	}
	return identityHash(parts...)
}

func IdentityProductionAcceptanceCriteriaV2() []IdentityProductionAcceptanceCriterionIDV2 {
	return []IdentityProductionAcceptanceCriterionIDV2{
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
	}
}

func validateIdentityProductionAcceptanceEvidenceV2(criterion IdentityProductionAcceptanceCriterionV2) error {
	if len(criterion.Evidence) == 0 {
		return fmt.Errorf("identity v2 production acceptance missing evidence for %s", criterion.ID)
	}
	sorted := append([]string(nil), criterion.Evidence...)
	sort.Strings(sorted)
	for i, ref := range sorted {
		if ref == "" || !strings.Contains(ref, ".go:") {
			return fmt.Errorf("identity v2 production acceptance invalid evidence reference %q", ref)
		}
		if ref != criterion.Evidence[i] {
			return fmt.Errorf("identity v2 production acceptance evidence for %s must be sorted", criterion.ID)
		}
		if i > 0 && sorted[i-1] == ref {
			return fmt.Errorf("duplicate identity v2 production acceptance evidence reference %q", ref)
		}
	}
	return nil
}
