package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityUnitTestCoverageAreaV2 string

const (
	IdentityUnitTestNameNormalizationV2          IdentityUnitTestCoverageAreaV2 = "name_normalization"
	IdentityUnitTestNameHashGenerationV2         IdentityUnitTestCoverageAreaV2 = "name_hash_generation"
	IdentityUnitTestCommitmentHashGenerationV2   IdentityUnitTestCoverageAreaV2 = "commitment_hash_generation"
	IdentityUnitTestCommitmentRevealValidationV2 IdentityUnitTestCoverageAreaV2 = "commitment_reveal_validation"
	IdentityUnitTestDomainLifecycleTransitionsV2 IdentityUnitTestCoverageAreaV2 = "domain_lifecycle_transitions"
	IdentityUnitTestNFTBindingChecksV2           IdentityUnitTestCoverageAreaV2 = "nft_binding_checks"
	IdentityUnitTestResolverFieldValidationV2    IdentityUnitTestCoverageAreaV2 = "resolver_field_validation"
	IdentityUnitTestReverseForwardConsistencyV2  IdentityUnitTestCoverageAreaV2 = "reverse_forward_consistency_validation"
	IdentityUnitTestDelegationScopeChecksV2      IdentityUnitTestCoverageAreaV2 = "delegation_scope_checks"
	IdentityUnitTestZonePolicyValidationV2       IdentityUnitTestCoverageAreaV2 = "zone_policy_validation"
	IdentityUnitTestPricingFunctionV2            IdentityUnitTestCoverageAreaV2 = "pricing_function"
	IdentityUnitTestAuctionWinnerSelectionV2     IdentityUnitTestCoverageAreaV2 = "auction_winner_selection"
	IdentityUnitTestProofEncodingV2              IdentityUnitTestCoverageAreaV2 = "proof_encoding"
)

type IdentityRequiredUnitTestCoverageV2 struct {
	RequiredAreas []IdentityUnitTestCoverageAreaV2
	ExistingTests map[IdentityUnitTestCoverageAreaV2][]string
	CoverageHash  string
}

func DefaultIdentityRequiredUnitTestCoverageV2() IdentityRequiredUnitTestCoverageV2 {
	coverage := IdentityRequiredUnitTestCoverageV2{
		RequiredAreas: IdentityRequiredUnitTestCoverageAreasV2(),
		ExistingTests: map[IdentityUnitTestCoverageAreaV2][]string{
			IdentityUnitTestNameNormalizationV2: {
				"x/identity/types/validation_v2_test.go:TestNameNormalizationV2MigrationAndTxVersionRejection",
				"x/identity/types/validation_v2_test.go:TestNameNormalizationV2ValidAndInvalidVectors",
			},
			IdentityUnitTestNameHashGenerationV2: {
				"x/identity/types/domain_v2_test.go:TestDomainRecordV2ParentHashUsesImmediateParent",
				"x/identity/types/domain_v2_test.go:TestDomainRecordV2RejectsHashAndParentMismatch",
			},
			IdentityUnitTestCommitmentHashGenerationV2: {
				"x/identity/types/commitment_v2_test.go:TestDomainCommitmentV2BindsNameCommitterSaltChainVersionAndIntent",
				"x/identity/types/spam_fairness_v2_test.go:TestIdentityAuctionCommitmentV2BindsChainDomain",
			},
			IdentityUnitTestCommitmentRevealValidationV2: {
				"x/identity/types/commitment_v2_test.go:TestDomainCommitmentV2ExpiresAfterRevealWindowAndRejectsReplay",
				"x/identity/types/commitment_v2_test.go:TestDomainCommitmentV2RevealWindowBoundaries",
				"x/identity/types/commitment_v2_test.go:TestRegistrationCommitRevealCreatesUsedTombstoneAndRejectsReplay",
			},
			IdentityUnitTestDomainLifecycleTransitionsV2: {
				"x/identity/types/lifecycle_state_machine_v2_test.go:TestDomainLifecycleStateMachineV2AuctionAlternative",
				"x/identity/types/lifecycle_state_machine_v2_test.go:TestDomainLifecycleStateMachineV2RegistrationRenewalGraceAndRelease",
			},
			IdentityUnitTestNFTBindingChecksV2: {
				"x/identity/types/nft_binding_test.go:TestDomainNFTBindingRequiresRegistryAndNFTModuleOwnerAgreement",
				"x/identity/types/ownership_consistency_v2_test.go:TestBrokenNFTBindingBlocksResolverUntilRepairV2",
			},
			IdentityUnitTestResolverFieldValidationV2: {
				"x/identity/types/resolution_v2_test.go:TestUnifiedResolutionRecordV2ResolverValidationLimitsAndFormats",
				"x/identity/types/reverse_payload_safety_v2_test.go:TestResolverPayloadSafetyFeesLimitsAndMalformedMetadataV2",
			},
			IdentityUnitTestReverseForwardConsistencyV2: {
				"x/identity/types/resolution_v2_test.go:TestReverseResolutionRecordV2VerifiedPrimaryAndAlias",
				"x/identity/types/resolver_spoofing_v2_test.go:TestResolverSpoofingPreventionReverseRecordsRequireForwardConsistencyV2",
			},
			IdentityUnitTestDelegationScopeChecksV2: {
				"x/identity/types/delegation_auction_v2_test.go:TestDelegationRecordV2ValidatesScopePermissionsAndUse",
				"x/identity/types/partial_delegation_zone_policy_v2_test.go:TestPartialDelegationV2ScopesVersionAndPrefixBoundLabels",
			},
			IdentityUnitTestZonePolicyValidationV2: {
				"x/identity/types/partial_delegation_zone_policy_v2_test.go:TestZonePolicyV2InheritanceLimitsCacheInvalidationAndRecursiveProof",
				"x/identity/types/resolver_subdomain_module_breakdown_v2_test.go:TestSubdomainModuleValidationV2DelegationZoneIndexAndPath",
			},
			IdentityUnitTestPricingFunctionV2: {
				"x/identity/types/cost_models_v2_test.go:TestIdentityResolverUpdateCostModelV2StorageDeltaChurnAndInlineFees",
				"x/identity/types/economics_v2_test.go:TestIdentityEconomicsV2PricingBoundariesAndQueries",
			},
			IdentityUnitTestAuctionWinnerSelectionV2: {
				"x/identity/types/keepers_v2_test.go:TestAuctionKeeperV2DeterministicWinnerAndFeeSplit",
				"x/identity/types/spam_fairness_v2_test.go:TestIdentityAuctionFairnessV2StateMachineTieForfeitAndFinalization",
			},
			IdentityUnitTestProofEncodingV2: {
				"x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment",
				"x/identity/types/proof_format_v2_test.go:TestRecursiveResolutionProofV2EncodingCommitmentAndCache",
			},
		},
	}
	coverage.CoverageHash = ComputeIdentityRequiredUnitTestCoverageHashV2(coverage)
	return coverage
}

func IdentityRequiredUnitTestCoverageAreasV2() []IdentityUnitTestCoverageAreaV2 {
	return []IdentityUnitTestCoverageAreaV2{
		IdentityUnitTestAuctionWinnerSelectionV2,
		IdentityUnitTestCommitmentHashGenerationV2,
		IdentityUnitTestCommitmentRevealValidationV2,
		IdentityUnitTestDelegationScopeChecksV2,
		IdentityUnitTestDomainLifecycleTransitionsV2,
		IdentityUnitTestNameHashGenerationV2,
		IdentityUnitTestNameNormalizationV2,
		IdentityUnitTestNFTBindingChecksV2,
		IdentityUnitTestPricingFunctionV2,
		IdentityUnitTestProofEncodingV2,
		IdentityUnitTestResolverFieldValidationV2,
		IdentityUnitTestReverseForwardConsistencyV2,
		IdentityUnitTestZonePolicyValidationV2,
	}
}

func ValidateIdentityRequiredUnitTestCoverageV2(coverage IdentityRequiredUnitTestCoverageV2) error {
	required := IdentityRequiredUnitTestCoverageAreasV2()
	if len(coverage.RequiredAreas) != len(required) {
		return fmt.Errorf("identity v2 unit test coverage must define %d required areas", len(required))
	}
	if !identityCoverageAreasEqualV2(coverage.RequiredAreas, required) {
		return errors.New("identity v2 unit test coverage required areas mismatch")
	}
	for _, area := range required {
		tests := append([]string(nil), coverage.ExistingTests[area]...)
		if len(tests) == 0 {
			return fmt.Errorf("identity v2 unit test coverage missing tests for %s", area)
		}
		sorted := append([]string(nil), tests...)
		sort.Strings(sorted)
		for i, test := range sorted {
			if test == "" || !strings.Contains(test, "_test.go:Test") {
				return fmt.Errorf("identity v2 unit test coverage invalid test reference %q", test)
			}
			if test != tests[i] {
				return fmt.Errorf("identity v2 unit test coverage tests for %s must be sorted", area)
			}
			if i > 0 && sorted[i-1] == test {
				return fmt.Errorf("duplicate identity v2 unit test coverage reference %q", test)
			}
		}
	}
	for area := range coverage.ExistingTests {
		if !identityCoverageAreaKnownV2(area, required) {
			return fmt.Errorf("identity v2 unit test coverage unknown area %s", area)
		}
	}
	if coverage.CoverageHash == "" || coverage.CoverageHash != ComputeIdentityRequiredUnitTestCoverageHashV2(coverage) {
		return errors.New("identity v2 unit test coverage hash mismatch")
	}
	return nil
}

func ComputeIdentityRequiredUnitTestCoverageHashV2(coverage IdentityRequiredUnitTestCoverageV2) string {
	parts := []string{"identity-required-unit-test-coverage-v2"}
	areas := append([]IdentityUnitTestCoverageAreaV2(nil), coverage.RequiredAreas...)
	sort.Slice(areas, func(i, j int) bool { return areas[i] < areas[j] })
	for _, area := range areas {
		parts = append(parts, string(area))
		parts = append(parts, sortedBreakdownStringsV2(coverage.ExistingTests[area])...)
	}
	return identityHash(parts...)
}

func identityCoverageAreasEqualV2(got []IdentityUnitTestCoverageAreaV2, want []IdentityUnitTestCoverageAreaV2) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func identityCoverageAreaKnownV2(area IdentityUnitTestCoverageAreaV2, required []IdentityUnitTestCoverageAreaV2) bool {
	for _, requiredArea := range required {
		if area == requiredArea {
			return true
		}
	}
	return false
}

type IdentityIntegrationTestCoverageAreaV2 string

const (
	IdentityIntegrationRegisterMintNFTV2                 IdentityIntegrationTestCoverageAreaV2 = "register_domain_and_mint_nft"
	IdentityIntegrationTransferNFTAtomicV2               IdentityIntegrationTestCoverageAreaV2 = "transfer_domain_and_update_nft_ownership_atomically"
	IdentityIntegrationResolverOwnerUpdateV2             IdentityIntegrationTestCoverageAreaV2 = "update_resolver_as_owner"
	IdentityIntegrationRejectUnauthorizedResolverV2      IdentityIntegrationTestCoverageAreaV2 = "reject_resolver_update_by_unauthorized_account"
	IdentityIntegrationCreateDelegatedSubdomainV2        IdentityIntegrationTestCoverageAreaV2 = "create_delegated_subdomain"
	IdentityIntegrationRevokeDelegationV2                IdentityIntegrationTestCoverageAreaV2 = "revoke_delegation_and_reject_further_delegate_updates"
	IdentityIntegrationRenewBeforeExpiryV2               IdentityIntegrationTestCoverageAreaV2 = "renew_domain_before_expiry"
	IdentityIntegrationExpireReleaseV2                   IdentityIntegrationTestCoverageAreaV2 = "expire_and_release_domain"
	IdentityIntegrationVerifyReverseV2                   IdentityIntegrationTestCoverageAreaV2 = "verify_reverse_resolution"
	IdentityIntegrationInvalidateReverseResolverUpdateV2 IdentityIntegrationTestCoverageAreaV2 = "invalidate_reverse_on_resolver_update"
	IdentityIntegrationCommitRevealAuctionV2             IdentityIntegrationTestCoverageAreaV2 = "run_commit_reveal_auction"
	IdentityIntegrationBatchDisjointResolversV2          IdentityIntegrationTestCoverageAreaV2 = "batch_update_disjoint_resolvers"
	IdentityIntegrationRecursiveProofV2                  IdentityIntegrationTestCoverageAreaV2 = "generate_and_verify_recursive_proof"
)

type IdentityRequiredIntegrationTestCoverageV2 struct {
	RequiredAreas []IdentityIntegrationTestCoverageAreaV2
	ExistingTests map[IdentityIntegrationTestCoverageAreaV2][]string
	CoverageHash  string
}

func DefaultIdentityRequiredIntegrationTestCoverageV2() IdentityRequiredIntegrationTestCoverageV2 {
	coverage := IdentityRequiredIntegrationTestCoverageV2{
		RequiredAreas: IdentityRequiredIntegrationTestCoverageAreasV2(),
		ExistingTests: map[IdentityIntegrationTestCoverageAreaV2][]string{
			IdentityIntegrationRegisterMintNFTV2: {
				"x/identity/types/keepers_v2_test.go:TestIdentityKeeperV2LifecycleAndNFTInvariant",
				"x/identity/types/spec_test.go:TestIdentitySpecRegisterAETDomain",
			},
			IdentityIntegrationTransferNFTAtomicV2: {
				"x/identity/types/nft_binding_test.go:TestDomainNFTBindingTransferUpdatesRegistryAndBindingAtomically",
				"x/identity/types/ownership_consistency_v2_test.go:TestIdentityNFTTransferHooksUpdateOrRejectAtomicallyV2",
			},
			IdentityIntegrationResolverOwnerUpdateV2: {
				"x/identity/types/resolver_test.go:TestApplyResolverUpdateSetAndChange",
				"x/identity/types/spec_test.go:TestIdentitySpecResolverUpdateRequiresOwner",
			},
			IdentityIntegrationRejectUnauthorizedResolverV2: {
				"x/identity/types/resolver_spoofing_v2_test.go:TestResolverSpoofingPreventionUnauthorizedUpdatesV2",
				"x/identity/types/resolver_test.go:TestApplyResolverUpdateRejectsZeroAndUnauthorized",
			},
			IdentityIntegrationCreateDelegatedSubdomainV2: {
				"x/identity/types/hierarchy_v2_test.go:TestSubdomainCreationV2DelegateZoneAndEphemeralValidation",
				"x/identity/types/partial_delegation_zone_policy_v2_test.go:TestPartialDelegationV2ScopesVersionAndPrefixBoundLabels",
			},
			IdentityIntegrationRevokeDelegationV2: {
				"x/identity/types/hierarchy_v2_test.go:TestDelegationScopeBitsAndTimeLockedRevocationV2",
				"x/identity/types/partial_delegation_zone_policy_v2_test.go:TestPartialDelegationV2RejectsEscalationAndParentTransferWithoutGrant",
			},
			IdentityIntegrationRenewBeforeExpiryV2: {
				"x/identity/types/lifecycle_test.go:TestRenewDomainExtendsExpiry",
				"x/identity/types/spec_test.go:TestIdentitySpecRenewalPreservesOwnership",
			},
			IdentityIntegrationExpireReleaseV2: {
				"x/identity/types/abci_lifecycle_v2_test.go:TestIdentityABCIPlusFinalizeBoundedExpiryAndCacheEventsV2",
				"x/identity/types/anti_squatting_v2_test.go:TestIdentityAntiSquattingExpiredDomainRecoveryAndReleaseV2",
			},
			IdentityIntegrationVerifyReverseV2: {
				"x/identity/types/resolution_v2_test.go:TestReverseResolutionVerificationTransactionV2ChecksVersionAndForward",
				"x/identity/types/resolver_test.go:TestReverseResolution",
			},
			IdentityIntegrationInvalidateReverseResolverUpdateV2: {
				"x/identity/types/resolution_v2_test.go:TestReverseResolutionRecordV2RequiresActiveDomainExpiryAndInvalidates",
				"x/identity/types/reverse_payload_safety_v2_test.go:TestReverseResolutionSafetyInvalidationHooksV2",
			},
			IdentityIntegrationCommitRevealAuctionV2: {
				"x/identity/types/delegation_auction_v2_test.go:TestAuctionRecordV2BuildsFromSealedAuctionLifecycle",
				"x/identity/types/spam_fairness_v2_test.go:TestIdentityAuctionFairnessV2StateMachineTieForfeitAndFinalization",
			},
			IdentityIntegrationBatchDisjointResolversV2: {
				"x/identity/types/batch_resolver_cache_v2_test.go:TestBatchResolverUpdatesV2AtomicAndPartialResults",
				"x/identity/types/blockstm_v2_test.go:TestIdentityBlockSTMBatchResolverUpdatesUseDisjointNameHashesV2",
			},
			IdentityIntegrationRecursiveProofV2: {
				"x/identity/types/light_client_verifier_v2_test.go:TestIdentityResolutionProofLightClientV2VerifiesRecursivePath",
				"x/identity/types/resolution_path_v2_test.go:TestBuildRecursiveResolutionProofV2UsesCanonicalRootToTargetPath",
			},
		},
	}
	coverage.CoverageHash = ComputeIdentityRequiredIntegrationTestCoverageHashV2(coverage)
	return coverage
}

func IdentityRequiredIntegrationTestCoverageAreasV2() []IdentityIntegrationTestCoverageAreaV2 {
	return []IdentityIntegrationTestCoverageAreaV2{
		IdentityIntegrationBatchDisjointResolversV2,
		IdentityIntegrationCommitRevealAuctionV2,
		IdentityIntegrationCreateDelegatedSubdomainV2,
		IdentityIntegrationExpireReleaseV2,
		IdentityIntegrationInvalidateReverseResolverUpdateV2,
		IdentityIntegrationRecursiveProofV2,
		IdentityIntegrationRegisterMintNFTV2,
		IdentityIntegrationRejectUnauthorizedResolverV2,
		IdentityIntegrationRenewBeforeExpiryV2,
		IdentityIntegrationResolverOwnerUpdateV2,
		IdentityIntegrationRevokeDelegationV2,
		IdentityIntegrationTransferNFTAtomicV2,
		IdentityIntegrationVerifyReverseV2,
	}
}

func ValidateIdentityRequiredIntegrationTestCoverageV2(coverage IdentityRequiredIntegrationTestCoverageV2) error {
	required := IdentityRequiredIntegrationTestCoverageAreasV2()
	if len(coverage.RequiredAreas) != len(required) {
		return fmt.Errorf("identity v2 integration test coverage must define %d required areas", len(required))
	}
	if !identityIntegrationCoverageAreasEqualV2(coverage.RequiredAreas, required) {
		return errors.New("identity v2 integration test coverage required areas mismatch")
	}
	if err := validateIdentityIntegrationCoverageReferencesV2(required, coverage.ExistingTests); err != nil {
		return err
	}
	if coverage.CoverageHash == "" || coverage.CoverageHash != ComputeIdentityRequiredIntegrationTestCoverageHashV2(coverage) {
		return errors.New("identity v2 integration test coverage hash mismatch")
	}
	return nil
}

func ComputeIdentityRequiredIntegrationTestCoverageHashV2(coverage IdentityRequiredIntegrationTestCoverageV2) string {
	parts := []string{"identity-required-integration-test-coverage-v2"}
	areas := append([]IdentityIntegrationTestCoverageAreaV2(nil), coverage.RequiredAreas...)
	sort.Slice(areas, func(i, j int) bool { return areas[i] < areas[j] })
	for _, area := range areas {
		parts = append(parts, string(area))
		parts = append(parts, sortedBreakdownStringsV2(coverage.ExistingTests[area])...)
	}
	return identityHash(parts...)
}

type IdentityInvariantTestCoverageAreaV2 string

const (
	IdentityInvariantRegistryNFTOwnerV2       IdentityInvariantTestCoverageAreaV2 = "active_registry_owner_equals_nft_owner"
	IdentityInvariantResolverActiveDomainV2   IdentityInvariantTestCoverageAreaV2 = "resolver_record_requires_active_domain"
	IdentityInvariantVerifiedReverseForwardV2 IdentityInvariantTestCoverageAreaV2 = "verified_reverse_record_has_matching_forward_resolution"
	IdentityInvariantChildExpiryV2            IdentityInvariantTestCoverageAreaV2 = "child_domain_expiry_does_not_exceed_parent_unless_detached"
	IdentityInvariantDelegationExpiryV2       IdentityInvariantTestCoverageAreaV2 = "delegation_expiry_does_not_exceed_authorized_domain_unless_detached"
	IdentityInvariantAuctionActiveDomainV2    IdentityInvariantTestCoverageAreaV2 = "auction_cannot_activate_already_active_domain"
	IdentityInvariantResolverByteLimitV2      IdentityInvariantTestCoverageAreaV2 = "resolver_record_byte_size_never_exceeds_parameter_limit"
	IdentityInvariantExpiryIndexV2            IdentityInvariantTestCoverageAreaV2 = "expiry_index_contains_every_expiring_domain"
	IdentityInvariantOwnerIndexV2             IdentityInvariantTestCoverageAreaV2 = "owner_index_matches_domain_ownership"
	IdentityInvariantCacheSourceVersionV2     IdentityInvariantTestCoverageAreaV2 = "cache_record_invalidates_on_source_version_change"
)

type IdentityRequiredInvariantTestCoverageV2 struct {
	RequiredAreas []IdentityInvariantTestCoverageAreaV2
	ExistingTests map[IdentityInvariantTestCoverageAreaV2][]string
	CoverageHash  string
}

func DefaultIdentityRequiredInvariantTestCoverageV2() IdentityRequiredInvariantTestCoverageV2 {
	coverage := IdentityRequiredInvariantTestCoverageV2{
		RequiredAreas: IdentityRequiredInvariantTestCoverageAreasV2(),
		ExistingTests: map[IdentityInvariantTestCoverageAreaV2][]string{
			IdentityInvariantRegistryNFTOwnerV2: {
				"x/identity/types/domain_v2_test.go:TestDomainRecordV2EnforcesNFTOwnerAndExpiry",
				"x/identity/types/ownership_consistency_v2_test.go:TestIdentityConsistencyAuditDetectsInvariantsV2",
			},
			IdentityInvariantResolverActiveDomainV2: {
				"x/identity/types/ownership_consistency_v2_test.go:TestIdentityConsistencyAuditDetectsInvariantsV2",
				"x/identity/types/resolver_spoofing_v2_test.go:TestResolverSpoofingPreventionResolverOwnerMustMatchActiveDomainV2",
			},
			IdentityInvariantVerifiedReverseForwardV2: {
				"x/identity/types/proof_v2_test.go:TestIdentityLightClientProofV2RejectsReverseMismatch",
				"x/identity/types/resolution_v2_test.go:TestReverseResolutionRecordV2VerifiedPrimaryAndAlias",
			},
			IdentityInvariantChildExpiryV2: {
				"x/identity/types/cost_models_v2_test.go:TestIdentitySubdomainCreationCostModelV2ExpiryConstraints",
				"x/identity/types/hierarchy_v2_test.go:TestSubdomainCreationV2DetachedLifecycleAndExpiryRules",
			},
			IdentityInvariantDelegationExpiryV2: {
				"x/identity/types/delegation_auction_v2_test.go:TestDelegationRecordV2RejectsInvalidCanonicalState",
				"x/identity/types/ownership_consistency_v2_test.go:TestIdentityConsistencyAuditDetectsInvariantsV2",
			},
			IdentityInvariantAuctionActiveDomainV2: {
				"x/identity/types/delegation_auction_v2_test.go:TestAuctionRecordV2RejectsInvalidFinalizedState",
				"x/identity/types/ownership_consistency_v2_test.go:TestIdentityConsistencyAuditDetectsInvariantsV2",
			},
			IdentityInvariantResolverByteLimitV2: {
				"x/identity/types/resolution_v2_test.go:TestUnifiedResolutionRecordV2ResolverValidationLimitsAndFormats",
				"x/identity/types/reverse_payload_safety_v2_test.go:TestResolverPayloadSafetyFeesLimitsAndMalformedMetadataV2",
			},
			IdentityInvariantExpiryIndexV2: {
				"x/identity/types/identity_core_module_breakdown_v2_test.go:TestIdentityCoreInvariantReportV2CatchesCoreFailureModes",
				"x/identity/types/storev2_spec_test.go:TestIdentityStoreV2SpecPrimaryKeyLayout",
			},
			IdentityInvariantOwnerIndexV2: {
				"x/identity/types/identity_core_module_breakdown_v2_test.go:TestIdentityCoreInvariantReportV2CatchesCoreFailureModes",
				"x/identity/types/storev2_spec_test.go:TestIdentityStoreV2SpecPrimaryKeyLayout",
			},
			IdentityInvariantCacheSourceVersionV2: {
				"x/identity/types/keepers_v2_test.go:TestKeeperInvariantRejectsStaleResolutionCache",
				"x/identity/types/resolution_cache_v2_test.go:TestResolutionCacheRecordV2InvalidatesOnDomainMutation",
			},
		},
	}
	coverage.CoverageHash = ComputeIdentityRequiredInvariantTestCoverageHashV2(coverage)
	return coverage
}

func IdentityRequiredInvariantTestCoverageAreasV2() []IdentityInvariantTestCoverageAreaV2 {
	return []IdentityInvariantTestCoverageAreaV2{
		IdentityInvariantAuctionActiveDomainV2,
		IdentityInvariantCacheSourceVersionV2,
		IdentityInvariantChildExpiryV2,
		IdentityInvariantDelegationExpiryV2,
		IdentityInvariantExpiryIndexV2,
		IdentityInvariantOwnerIndexV2,
		IdentityInvariantRegistryNFTOwnerV2,
		IdentityInvariantResolverActiveDomainV2,
		IdentityInvariantResolverByteLimitV2,
		IdentityInvariantVerifiedReverseForwardV2,
	}
}

func ValidateIdentityRequiredInvariantTestCoverageV2(coverage IdentityRequiredInvariantTestCoverageV2) error {
	required := IdentityRequiredInvariantTestCoverageAreasV2()
	if len(coverage.RequiredAreas) != len(required) {
		return fmt.Errorf("identity v2 invariant test coverage must define %d required areas", len(required))
	}
	if !identityInvariantCoverageAreasEqualV2(coverage.RequiredAreas, required) {
		return errors.New("identity v2 invariant test coverage required areas mismatch")
	}
	if err := validateIdentityInvariantCoverageReferencesV2(required, coverage.ExistingTests); err != nil {
		return err
	}
	if coverage.CoverageHash == "" || coverage.CoverageHash != ComputeIdentityRequiredInvariantTestCoverageHashV2(coverage) {
		return errors.New("identity v2 invariant test coverage hash mismatch")
	}
	return nil
}

func ComputeIdentityRequiredInvariantTestCoverageHashV2(coverage IdentityRequiredInvariantTestCoverageV2) string {
	parts := []string{"identity-required-invariant-test-coverage-v2"}
	areas := append([]IdentityInvariantTestCoverageAreaV2(nil), coverage.RequiredAreas...)
	sort.Slice(areas, func(i, j int) bool { return areas[i] < areas[j] })
	for _, area := range areas {
		parts = append(parts, string(area))
		parts = append(parts, sortedBreakdownStringsV2(coverage.ExistingTests[area])...)
	}
	return identityHash(parts...)
}

func identityIntegrationCoverageAreasEqualV2(got []IdentityIntegrationTestCoverageAreaV2, want []IdentityIntegrationTestCoverageAreaV2) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func identityInvariantCoverageAreasEqualV2(got []IdentityInvariantTestCoverageAreaV2, want []IdentityInvariantTestCoverageAreaV2) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func validateIdentityIntegrationCoverageReferencesV2(required []IdentityIntegrationTestCoverageAreaV2, existing map[IdentityIntegrationTestCoverageAreaV2][]string) error {
	known := map[IdentityIntegrationTestCoverageAreaV2]bool{}
	for _, area := range required {
		known[area] = true
		if err := validateIdentityCoverageTestReferencesV2("integration", string(area), existing[area]); err != nil {
			return err
		}
	}
	for area := range existing {
		if !known[area] {
			return fmt.Errorf("identity v2 integration test coverage unknown area %s", area)
		}
	}
	return nil
}

func validateIdentityInvariantCoverageReferencesV2(required []IdentityInvariantTestCoverageAreaV2, existing map[IdentityInvariantTestCoverageAreaV2][]string) error {
	known := map[IdentityInvariantTestCoverageAreaV2]bool{}
	for _, area := range required {
		known[area] = true
		if err := validateIdentityCoverageTestReferencesV2("invariant", string(area), existing[area]); err != nil {
			return err
		}
	}
	for area := range existing {
		if !known[area] {
			return fmt.Errorf("identity v2 invariant test coverage unknown area %s", area)
		}
	}
	return nil
}

func validateIdentityCoverageTestReferencesV2(kind string, area string, tests []string) error {
	if len(tests) == 0 {
		return fmt.Errorf("identity v2 %s test coverage missing tests for %s", kind, area)
	}
	sorted := append([]string(nil), tests...)
	sort.Strings(sorted)
	for i, test := range sorted {
		if test == "" || !strings.Contains(test, "_test.go:Test") {
			return fmt.Errorf("identity v2 %s test coverage invalid test reference %q", kind, test)
		}
		if test != tests[i] {
			return fmt.Errorf("identity v2 %s test coverage tests for %s must be sorted", kind, area)
		}
		if i > 0 && sorted[i-1] == test {
			return fmt.Errorf("duplicate identity v2 %s test coverage reference %q", kind, test)
		}
	}
	return nil
}
