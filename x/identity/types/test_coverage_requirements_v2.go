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
