package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityUnitTestCoverageAreaV2 string

const (
	IdentityUnitTestNameNormalizationV2		IdentityUnitTestCoverageAreaV2	= "name_normalization"
	IdentityUnitTestNameHashGenerationV2		IdentityUnitTestCoverageAreaV2	= "name_hash_generation"
	IdentityUnitTestCommitmentHashGenerationV2	IdentityUnitTestCoverageAreaV2	= "commitment_hash_generation"
	IdentityUnitTestCommitmentRevealValidationV2	IdentityUnitTestCoverageAreaV2	= "commitment_reveal_validation"
	IdentityUnitTestDomainLifecycleTransitionsV2	IdentityUnitTestCoverageAreaV2	= "domain_lifecycle_transitions"
	IdentityUnitTestNFTBindingChecksV2		IdentityUnitTestCoverageAreaV2	= "nft_binding_checks"
	IdentityUnitTestResolverFieldValidationV2	IdentityUnitTestCoverageAreaV2	= "resolver_field_validation"
	IdentityUnitTestReverseForwardConsistencyV2	IdentityUnitTestCoverageAreaV2	= "reverse_forward_consistency_validation"
	IdentityUnitTestDelegationScopeChecksV2		IdentityUnitTestCoverageAreaV2	= "delegation_scope_checks"
	IdentityUnitTestZonePolicyValidationV2		IdentityUnitTestCoverageAreaV2	= "zone_policy_validation"
	IdentityUnitTestPricingFunctionV2		IdentityUnitTestCoverageAreaV2	= "pricing_function"
	IdentityUnitTestAuctionWinnerSelectionV2	IdentityUnitTestCoverageAreaV2	= "auction_winner_selection"
	IdentityUnitTestProofEncodingV2			IdentityUnitTestCoverageAreaV2	= "proof_encoding"
)

type IdentityRequiredUnitTestCoverageV2 struct {
	RequiredAreas	[]IdentityUnitTestCoverageAreaV2
	ExistingTests	map[IdentityUnitTestCoverageAreaV2][]string
	CoverageHash	string
}

func DefaultIdentityRequiredUnitTestCoverageV2() IdentityRequiredUnitTestCoverageV2 {
	coverage := IdentityRequiredUnitTestCoverageV2{
		RequiredAreas:	IdentityRequiredUnitTestCoverageAreasV2(),
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
	IdentityIntegrationRegisterMintNFTV2			IdentityIntegrationTestCoverageAreaV2	= "register_domain_and_mint_nft"
	IdentityIntegrationTransferNFTAtomicV2			IdentityIntegrationTestCoverageAreaV2	= "transfer_domain_and_update_nft_ownership_atomically"
	IdentityIntegrationResolverOwnerUpdateV2		IdentityIntegrationTestCoverageAreaV2	= "update_resolver_as_owner"
	IdentityIntegrationRejectUnauthorizedResolverV2		IdentityIntegrationTestCoverageAreaV2	= "reject_resolver_update_by_unauthorized_account"
	IdentityIntegrationCreateDelegatedSubdomainV2		IdentityIntegrationTestCoverageAreaV2	= "create_delegated_subdomain"
	IdentityIntegrationRevokeDelegationV2			IdentityIntegrationTestCoverageAreaV2	= "revoke_delegation_and_reject_further_delegate_updates"
	IdentityIntegrationRenewBeforeExpiryV2			IdentityIntegrationTestCoverageAreaV2	= "renew_domain_before_expiry"
	IdentityIntegrationExpireReleaseV2			IdentityIntegrationTestCoverageAreaV2	= "expire_and_release_domain"
	IdentityIntegrationVerifyReverseV2			IdentityIntegrationTestCoverageAreaV2	= "verify_reverse_resolution"
	IdentityIntegrationInvalidateReverseResolverUpdateV2	IdentityIntegrationTestCoverageAreaV2	= "invalidate_reverse_on_resolver_update"
	IdentityIntegrationCommitRevealAuctionV2		IdentityIntegrationTestCoverageAreaV2	= "run_commit_reveal_auction"
	IdentityIntegrationBatchDisjointResolversV2		IdentityIntegrationTestCoverageAreaV2	= "batch_update_disjoint_resolvers"
	IdentityIntegrationRecursiveProofV2			IdentityIntegrationTestCoverageAreaV2	= "generate_and_verify_recursive_proof"
)

type IdentityRequiredIntegrationTestCoverageV2 struct {
	RequiredAreas	[]IdentityIntegrationTestCoverageAreaV2
	ExistingTests	map[IdentityIntegrationTestCoverageAreaV2][]string
	CoverageHash	string
}

func DefaultIdentityRequiredIntegrationTestCoverageV2() IdentityRequiredIntegrationTestCoverageV2 {
	coverage := IdentityRequiredIntegrationTestCoverageV2{
		RequiredAreas:	IdentityRequiredIntegrationTestCoverageAreasV2(),
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
	IdentityInvariantRegistryNFTOwnerV2		IdentityInvariantTestCoverageAreaV2	= "active_registry_owner_equals_nft_owner"
	IdentityInvariantResolverActiveDomainV2		IdentityInvariantTestCoverageAreaV2	= "resolver_record_requires_active_domain"
	IdentityInvariantVerifiedReverseForwardV2	IdentityInvariantTestCoverageAreaV2	= "verified_reverse_record_has_matching_forward_resolution"
	IdentityInvariantChildExpiryV2			IdentityInvariantTestCoverageAreaV2	= "child_domain_expiry_does_not_exceed_parent_unless_detached"
	IdentityInvariantDelegationExpiryV2		IdentityInvariantTestCoverageAreaV2	= "delegation_expiry_does_not_exceed_authorized_domain_unless_detached"
	IdentityInvariantAuctionActiveDomainV2		IdentityInvariantTestCoverageAreaV2	= "auction_cannot_activate_already_active_domain"
	IdentityInvariantResolverByteLimitV2		IdentityInvariantTestCoverageAreaV2	= "resolver_record_byte_size_never_exceeds_parameter_limit"
	IdentityInvariantExpiryIndexV2			IdentityInvariantTestCoverageAreaV2	= "expiry_index_contains_every_expiring_domain"
	IdentityInvariantOwnerIndexV2			IdentityInvariantTestCoverageAreaV2	= "owner_index_matches_domain_ownership"
	IdentityInvariantCacheSourceVersionV2		IdentityInvariantTestCoverageAreaV2	= "cache_record_invalidates_on_source_version_change"
)

type IdentityRequiredInvariantTestCoverageV2 struct {
	RequiredAreas	[]IdentityInvariantTestCoverageAreaV2
	ExistingTests	map[IdentityInvariantTestCoverageAreaV2][]string
	CoverageHash	string
}

func DefaultIdentityRequiredInvariantTestCoverageV2() IdentityRequiredInvariantTestCoverageV2 {
	coverage := IdentityRequiredInvariantTestCoverageV2{
		RequiredAreas:	IdentityRequiredInvariantTestCoverageAreasV2(),
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
		if test == "" || !identityCoverageReferenceHasSupportedFunctionV2(test) {
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

func identityCoverageReferenceHasSupportedFunctionV2(ref string) bool {
	return strings.Contains(ref, "_test.go:Test") ||
		strings.Contains(ref, "_test.go:Fuzz") ||
		strings.Contains(ref, "_test.go:Benchmark")
}

type IdentityFuzzTestCoverageAreaV2 string

const (
	IdentityFuzzMalformedNamesV2			IdentityFuzzTestCoverageAreaV2	= "malformed_names"
	IdentityFuzzBoundaryLengthNamesV2		IdentityFuzzTestCoverageAreaV2	= "boundary_length_names"
	IdentityFuzzSpoofingPatternCandidatesV2		IdentityFuzzTestCoverageAreaV2	= "spoofing_pattern_candidates"
	IdentityFuzzCommitmentPreimagesV2		IdentityFuzzTestCoverageAreaV2	= "commitment_preimages"
	IdentityFuzzAuctionBidRevealOrderingV2		IdentityFuzzTestCoverageAreaV2	= "auction_bid_reveal_ordering"
	IdentityFuzzResolverPayloadsV2			IdentityFuzzTestCoverageAreaV2	= "resolver_payloads"
	IdentityFuzzInterfaceDescriptorSchemasV2	IdentityFuzzTestCoverageAreaV2	= "interface_descriptor_schemas"
	IdentityFuzzDelegationPermissionCombosV2	IdentityFuzzTestCoverageAreaV2	= "delegation_permission_combinations"
	IdentityFuzzRecursiveProofPathsV2		IdentityFuzzTestCoverageAreaV2	= "recursive_proof_paths"
	IdentityFuzzReverseResolutionMismatchesV2	IdentityFuzzTestCoverageAreaV2	= "reverse_resolution_mismatches"
	IdentityFuzzBatchUpdateOrderingV2		IdentityFuzzTestCoverageAreaV2	= "batch_update_ordering"
)

type IdentityRequiredFuzzTestCoverageV2 struct {
	RequiredAreas	[]IdentityFuzzTestCoverageAreaV2
	ExistingTests	map[IdentityFuzzTestCoverageAreaV2][]string
	CoverageHash	string
}

func DefaultIdentityRequiredFuzzTestCoverageV2() IdentityRequiredFuzzTestCoverageV2 {
	coverage := IdentityRequiredFuzzTestCoverageV2{
		RequiredAreas:	IdentityRequiredFuzzTestCoverageAreasV2(),
		ExistingTests: map[IdentityFuzzTestCoverageAreaV2][]string{
			IdentityFuzzMalformedNamesV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzIdentityMalformedNamesV2",
			},
			IdentityFuzzBoundaryLengthNamesV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzIdentityBoundaryLengthNamesV2",
			},
			IdentityFuzzSpoofingPatternCandidatesV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzIdentitySpoofingPatternCandidatesV2",
			},
			IdentityFuzzCommitmentPreimagesV2: {
				"x/identity/types/commitment_v2_test.go:FuzzDomainCommitmentV2RevealReplayProtection",
			},
			IdentityFuzzAuctionBidRevealOrderingV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzAuctionBidRevealOrderingV2",
			},
			IdentityFuzzResolverPayloadsV2: {
				"x/identity/types/reverse_payload_safety_v2_test.go:FuzzUnifiedResolverPayloadSafetyV2",
			},
			IdentityFuzzInterfaceDescriptorSchemasV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzInterfaceDescriptorSchemasV2",
			},
			IdentityFuzzDelegationPermissionCombosV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzDelegationPermissionCombinationsV2",
			},
			IdentityFuzzRecursiveProofPathsV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzRecursiveProofPathsV2",
			},
			IdentityFuzzReverseResolutionMismatchesV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzReverseResolutionMismatchesV2",
			},
			IdentityFuzzBatchUpdateOrderingV2: {
				"x/identity/types/fuzz_dns_v2_test.go:FuzzBatchUpdateOrderingV2",
			},
		},
	}
	coverage.CoverageHash = ComputeIdentityRequiredFuzzTestCoverageHashV2(coverage)
	return coverage
}

func IdentityRequiredFuzzTestCoverageAreasV2() []IdentityFuzzTestCoverageAreaV2 {
	return []IdentityFuzzTestCoverageAreaV2{
		IdentityFuzzAuctionBidRevealOrderingV2,
		IdentityFuzzBatchUpdateOrderingV2,
		IdentityFuzzBoundaryLengthNamesV2,
		IdentityFuzzCommitmentPreimagesV2,
		IdentityFuzzDelegationPermissionCombosV2,
		IdentityFuzzInterfaceDescriptorSchemasV2,
		IdentityFuzzMalformedNamesV2,
		IdentityFuzzRecursiveProofPathsV2,
		IdentityFuzzResolverPayloadsV2,
		IdentityFuzzReverseResolutionMismatchesV2,
		IdentityFuzzSpoofingPatternCandidatesV2,
	}
}

func ValidateIdentityRequiredFuzzTestCoverageV2(coverage IdentityRequiredFuzzTestCoverageV2) error {
	required := IdentityRequiredFuzzTestCoverageAreasV2()
	if len(coverage.RequiredAreas) != len(required) {
		return fmt.Errorf("identity v2 fuzz test coverage must define %d required areas", len(required))
	}
	if !identityFuzzCoverageAreasEqualV2(coverage.RequiredAreas, required) {
		return errors.New("identity v2 fuzz test coverage required areas mismatch")
	}
	if err := validateIdentityFuzzCoverageReferencesV2(required, coverage.ExistingTests); err != nil {
		return err
	}
	if coverage.CoverageHash == "" || coverage.CoverageHash != ComputeIdentityRequiredFuzzTestCoverageHashV2(coverage) {
		return errors.New("identity v2 fuzz test coverage hash mismatch")
	}
	return nil
}

func ComputeIdentityRequiredFuzzTestCoverageHashV2(coverage IdentityRequiredFuzzTestCoverageV2) string {
	parts := []string{"identity-required-fuzz-test-coverage-v2"}
	areas := append([]IdentityFuzzTestCoverageAreaV2(nil), coverage.RequiredAreas...)
	sort.Slice(areas, func(i, j int) bool { return areas[i] < areas[j] })
	for _, area := range areas {
		parts = append(parts, string(area))
		parts = append(parts, sortedBreakdownStringsV2(coverage.ExistingTests[area])...)
	}
	return identityHash(parts...)
}

type IdentityPerformanceTestCoverageAreaV2 string

const (
	IdentityPerformanceDirectResolutionReadLatencyV2	IdentityPerformanceTestCoverageAreaV2	= "direct_resolution_read_latency"
	IdentityPerformanceRecursiveReadLatencyByDepthV2	IdentityPerformanceTestCoverageAreaV2	= "recursive_resolution_read_latency_by_depth"
	IdentityPerformanceResolverUpdateWriteLatencyV2		IdentityPerformanceTestCoverageAreaV2	= "resolver_update_write_latency"
	IdentityPerformanceBatchResolverUpdatesPerBlockV2	IdentityPerformanceTestCoverageAreaV2	= "batch_resolver_updates_per_block"
	IdentityPerformanceBatchRenewalsPerBlockV2		IdentityPerformanceTestCoverageAreaV2	= "batch_renewals_per_block"
	IdentityPerformanceDomainRegistrationsPerBlockV2	IdentityPerformanceTestCoverageAreaV2	= "domain_registrations_per_block"
	IdentityPerformanceBlockSTMMixedConflictRateV2		IdentityPerformanceTestCoverageAreaV2	= "blockstm_conflict_rate_under_mixed_identity_workload"
	IdentityPerformanceStoreV2ProofGenerationLatencyV2	IdentityPerformanceTestCoverageAreaV2	= "store_v2_proof_generation_latency"
	IdentityPerformanceAdaptiveSyncLargeRecoveryTimeV2	IdentityPerformanceTestCoverageAreaV2	= "adaptive_sync_recovery_time_with_large_identity_state"
	IdentityPerformanceExportImportIdentityStateTimeV2	IdentityPerformanceTestCoverageAreaV2	= "export_import_time_for_identity_state"
)

type IdentityRequiredPerformanceTestCoverageV2 struct {
	RequiredAreas	[]IdentityPerformanceTestCoverageAreaV2
	ExistingTests	map[IdentityPerformanceTestCoverageAreaV2][]string
	CoverageHash	string
}

func DefaultIdentityRequiredPerformanceTestCoverageV2() IdentityRequiredPerformanceTestCoverageV2 {
	coverage := IdentityRequiredPerformanceTestCoverageV2{
		RequiredAreas:	IdentityRequiredPerformanceTestCoverageAreasV2(),
		ExistingTests: map[IdentityPerformanceTestCoverageAreaV2][]string{
			IdentityPerformanceDirectResolutionReadLatencyV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityStoreV2DirectResolutionReadPath",
			},
			IdentityPerformanceRecursiveReadLatencyByDepthV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityStoreV2RecursiveResolutionReadPath",
			},
			IdentityPerformanceResolverUpdateWriteLatencyV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityResolverUpdateWritePath",
			},
			IdentityPerformanceBatchResolverUpdatesPerBlockV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityBlockSTMBatchResolverUpdates",
			},
			IdentityPerformanceBatchRenewalsPerBlockV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityBlockSTMBatchRenewalsPerBlock",
			},
			IdentityPerformanceDomainRegistrationsPerBlockV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityRegistrationsPerBlock",
			},
			IdentityPerformanceBlockSTMMixedConflictRateV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityBlockSTMMixedConflictClassification",
			},
			IdentityPerformanceStoreV2ProofGenerationLatencyV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityProofQuery",
			},
			IdentityPerformanceAdaptiveSyncLargeRecoveryTimeV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityAdaptiveSyncRecoveryLargeState",
			},
			IdentityPerformanceExportImportIdentityStateTimeV2: {
				"x/identity/types/bench_test.go:BenchmarkIdentityExportImportLargeState",
			},
		},
	}
	coverage.CoverageHash = ComputeIdentityRequiredPerformanceTestCoverageHashV2(coverage)
	return coverage
}

func IdentityRequiredPerformanceTestCoverageAreasV2() []IdentityPerformanceTestCoverageAreaV2 {
	return []IdentityPerformanceTestCoverageAreaV2{
		IdentityPerformanceAdaptiveSyncLargeRecoveryTimeV2,
		IdentityPerformanceBatchRenewalsPerBlockV2,
		IdentityPerformanceBatchResolverUpdatesPerBlockV2,
		IdentityPerformanceBlockSTMMixedConflictRateV2,
		IdentityPerformanceDirectResolutionReadLatencyV2,
		IdentityPerformanceDomainRegistrationsPerBlockV2,
		IdentityPerformanceExportImportIdentityStateTimeV2,
		IdentityPerformanceRecursiveReadLatencyByDepthV2,
		IdentityPerformanceResolverUpdateWriteLatencyV2,
		IdentityPerformanceStoreV2ProofGenerationLatencyV2,
	}
}

func ValidateIdentityRequiredPerformanceTestCoverageV2(coverage IdentityRequiredPerformanceTestCoverageV2) error {
	required := IdentityRequiredPerformanceTestCoverageAreasV2()
	if len(coverage.RequiredAreas) != len(required) {
		return fmt.Errorf("identity v2 performance test coverage must define %d required areas", len(required))
	}
	if !identityPerformanceCoverageAreasEqualV2(coverage.RequiredAreas, required) {
		return errors.New("identity v2 performance test coverage required areas mismatch")
	}
	if err := validateIdentityPerformanceCoverageReferencesV2(required, coverage.ExistingTests); err != nil {
		return err
	}
	if coverage.CoverageHash == "" || coverage.CoverageHash != ComputeIdentityRequiredPerformanceTestCoverageHashV2(coverage) {
		return errors.New("identity v2 performance test coverage hash mismatch")
	}
	return nil
}

func ComputeIdentityRequiredPerformanceTestCoverageHashV2(coverage IdentityRequiredPerformanceTestCoverageV2) string {
	parts := []string{"identity-required-performance-test-coverage-v2"}
	areas := append([]IdentityPerformanceTestCoverageAreaV2(nil), coverage.RequiredAreas...)
	sort.Slice(areas, func(i, j int) bool { return areas[i] < areas[j] })
	for _, area := range areas {
		parts = append(parts, string(area))
		parts = append(parts, sortedBreakdownStringsV2(coverage.ExistingTests[area])...)
	}
	return identityHash(parts...)
}

func identityFuzzCoverageAreasEqualV2(got []IdentityFuzzTestCoverageAreaV2, want []IdentityFuzzTestCoverageAreaV2) bool {
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

func identityPerformanceCoverageAreasEqualV2(got []IdentityPerformanceTestCoverageAreaV2, want []IdentityPerformanceTestCoverageAreaV2) bool {
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

func validateIdentityFuzzCoverageReferencesV2(required []IdentityFuzzTestCoverageAreaV2, existing map[IdentityFuzzTestCoverageAreaV2][]string) error {
	known := map[IdentityFuzzTestCoverageAreaV2]bool{}
	for _, area := range required {
		known[area] = true
		if err := validateIdentityCoverageTestReferencesV2("fuzz", string(area), existing[area]); err != nil {
			return err
		}
	}
	for area := range existing {
		if !known[area] {
			return fmt.Errorf("identity v2 fuzz test coverage unknown area %s", area)
		}
	}
	return nil
}

func validateIdentityPerformanceCoverageReferencesV2(required []IdentityPerformanceTestCoverageAreaV2, existing map[IdentityPerformanceTestCoverageAreaV2][]string) error {
	known := map[IdentityPerformanceTestCoverageAreaV2]bool{}
	for _, area := range required {
		known[area] = true
		if err := validateIdentityCoverageTestReferencesV2("performance", string(area), existing[area]); err != nil {
			return err
		}
	}
	for area := range existing {
		if !known[area] {
			return fmt.Errorf("identity v2 performance test coverage unknown area %s", area)
		}
	}
	return nil
}
