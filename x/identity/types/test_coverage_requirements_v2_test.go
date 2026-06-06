package types

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityRequiredUnitTestCoverageV2CoversSection171(t *testing.T) {
	coverage := DefaultIdentityRequiredUnitTestCoverageV2()
	require.NoError(t, ValidateIdentityRequiredUnitTestCoverageV2(coverage))
	require.Len(t, coverage.RequiredAreas, 13)
	require.NotEmpty(t, coverage.CoverageHash)

	for _, area := range []IdentityUnitTestCoverageAreaV2{
		IdentityUnitTestNameNormalizationV2,
		IdentityUnitTestNameHashGenerationV2,
		IdentityUnitTestCommitmentHashGenerationV2,
		IdentityUnitTestCommitmentRevealValidationV2,
		IdentityUnitTestDomainLifecycleTransitionsV2,
		IdentityUnitTestNFTBindingChecksV2,
		IdentityUnitTestResolverFieldValidationV2,
		IdentityUnitTestReverseForwardConsistencyV2,
		IdentityUnitTestDelegationScopeChecksV2,
		IdentityUnitTestZonePolicyValidationV2,
		IdentityUnitTestPricingFunctionV2,
		IdentityUnitTestAuctionWinnerSelectionV2,
		IdentityUnitTestProofEncodingV2,
	} {
		require.Contains(t, coverage.RequiredAreas, area)
		require.NotEmpty(t, coverage.ExistingTests[area])
	}

	require.Contains(t, coverage.ExistingTests[IdentityUnitTestNameNormalizationV2], "x/identity/types/validation_v2_test.go:TestNameNormalizationV2ValidAndInvalidVectors")
	require.Contains(t, coverage.ExistingTests[IdentityUnitTestProofEncodingV2], "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment")
	requireCoverageReferencesExistV2(t, identityUnitCoverageReferencesV2(coverage))
}

func TestIdentityRequiredUnitTestCoverageV2RejectsMissingAreaBadReferenceAndHashMismatch(t *testing.T) {
	coverage := DefaultIdentityRequiredUnitTestCoverageV2()

	missingArea := coverage
	missingArea.RequiredAreas = missingArea.RequiredAreas[:len(missingArea.RequiredAreas)-1]
	missingArea.CoverageHash = ComputeIdentityRequiredUnitTestCoverageHashV2(missingArea)
	require.ErrorContains(t, ValidateIdentityRequiredUnitTestCoverageV2(missingArea), "required areas")

	badReference := coverage
	badReference.ExistingTests = copyIdentityRequiredUnitTestCoverageMapV2(coverage.ExistingTests)
	badReference.ExistingTests[IdentityUnitTestNameNormalizationV2] = []string{"x/identity/types/validation_v2_test.go:not_a_test"}
	badReference.CoverageHash = ComputeIdentityRequiredUnitTestCoverageHashV2(badReference)
	require.ErrorContains(t, ValidateIdentityRequiredUnitTestCoverageV2(badReference), "invalid test reference")

	tampered := coverage
	tampered.ExistingTests = copyIdentityRequiredUnitTestCoverageMapV2(coverage.ExistingTests)
	tampered.ExistingTests[IdentityUnitTestProofEncodingV2][0] = "x/identity/types/proof_format_v2_test.go:TestAProofEncodingTamper"
	require.ErrorContains(t, ValidateIdentityRequiredUnitTestCoverageV2(tampered), "hash mismatch")
}

func TestIdentityRequiredIntegrationTestCoverageV2CoversSection172(t *testing.T) {
	coverage := DefaultIdentityRequiredIntegrationTestCoverageV2()
	require.NoError(t, ValidateIdentityRequiredIntegrationTestCoverageV2(coverage))
	require.Len(t, coverage.RequiredAreas, 13)
	require.NotEmpty(t, coverage.CoverageHash)

	for _, area := range []IdentityIntegrationTestCoverageAreaV2{
		IdentityIntegrationRegisterMintNFTV2,
		IdentityIntegrationTransferNFTAtomicV2,
		IdentityIntegrationResolverOwnerUpdateV2,
		IdentityIntegrationRejectUnauthorizedResolverV2,
		IdentityIntegrationCreateDelegatedSubdomainV2,
		IdentityIntegrationRevokeDelegationV2,
		IdentityIntegrationRenewBeforeExpiryV2,
		IdentityIntegrationExpireReleaseV2,
		IdentityIntegrationVerifyReverseV2,
		IdentityIntegrationInvalidateReverseResolverUpdateV2,
		IdentityIntegrationCommitRevealAuctionV2,
		IdentityIntegrationBatchDisjointResolversV2,
		IdentityIntegrationRecursiveProofV2,
	} {
		require.Contains(t, coverage.RequiredAreas, area)
		require.NotEmpty(t, coverage.ExistingTests[area])
	}

	require.Contains(t, coverage.ExistingTests[IdentityIntegrationRegisterMintNFTV2], "x/identity/types/spec_test.go:TestIdentitySpecRegisterAETDomain")
	require.Contains(t, coverage.ExistingTests[IdentityIntegrationRecursiveProofV2], "x/identity/types/light_client_verifier_v2_test.go:TestIdentityResolutionProofLightClientV2VerifiesRecursivePath")
	requireCoverageReferencesExistV2(t, identityIntegrationCoverageReferencesV2(coverage))
}

func TestIdentityRequiredInvariantTestCoverageV2CoversSection173(t *testing.T) {
	coverage := DefaultIdentityRequiredInvariantTestCoverageV2()
	require.NoError(t, ValidateIdentityRequiredInvariantTestCoverageV2(coverage))
	require.Len(t, coverage.RequiredAreas, 10)
	require.NotEmpty(t, coverage.CoverageHash)

	for _, area := range []IdentityInvariantTestCoverageAreaV2{
		IdentityInvariantRegistryNFTOwnerV2,
		IdentityInvariantResolverActiveDomainV2,
		IdentityInvariantVerifiedReverseForwardV2,
		IdentityInvariantChildExpiryV2,
		IdentityInvariantDelegationExpiryV2,
		IdentityInvariantAuctionActiveDomainV2,
		IdentityInvariantResolverByteLimitV2,
		IdentityInvariantExpiryIndexV2,
		IdentityInvariantOwnerIndexV2,
		IdentityInvariantCacheSourceVersionV2,
	} {
		require.Contains(t, coverage.RequiredAreas, area)
		require.NotEmpty(t, coverage.ExistingTests[area])
	}

	require.Contains(t, coverage.ExistingTests[IdentityInvariantRegistryNFTOwnerV2], "x/identity/types/domain_v2_test.go:TestDomainRecordV2EnforcesNFTOwnerAndExpiry")
	require.Contains(t, coverage.ExistingTests[IdentityInvariantCacheSourceVersionV2], "x/identity/types/keepers_v2_test.go:TestKeeperInvariantRejectsStaleResolutionCache")
	requireCoverageReferencesExistV2(t, identityInvariantCoverageReferencesV2(coverage))
}

func TestIdentityRequiredIntegrationAndInvariantCoverageV2RejectBadReferenceAndHashMismatch(t *testing.T) {
	integration := DefaultIdentityRequiredIntegrationTestCoverageV2()
	badIntegration := integration
	badIntegration.ExistingTests = copyIdentityRequiredIntegrationTestCoverageMapV2(integration.ExistingTests)
	badIntegration.ExistingTests[IdentityIntegrationRegisterMintNFTV2] = []string{"x/identity/types/spec_test.go:not_a_test"}
	badIntegration.CoverageHash = ComputeIdentityRequiredIntegrationTestCoverageHashV2(badIntegration)
	require.ErrorContains(t, ValidateIdentityRequiredIntegrationTestCoverageV2(badIntegration), "invalid test reference")

	tamperedIntegration := integration
	tamperedIntegration.ExistingTests = copyIdentityRequiredIntegrationTestCoverageMapV2(integration.ExistingTests)
	tamperedIntegration.ExistingTests[IdentityIntegrationRecursiveProofV2][0] = "x/identity/types/light_client_verifier_v2_test.go:TestARecursiveProofTamper"
	require.ErrorContains(t, ValidateIdentityRequiredIntegrationTestCoverageV2(tamperedIntegration), "hash mismatch")

	invariant := DefaultIdentityRequiredInvariantTestCoverageV2()
	missingInvariant := invariant
	missingInvariant.RequiredAreas = missingInvariant.RequiredAreas[:len(missingInvariant.RequiredAreas)-1]
	missingInvariant.CoverageHash = ComputeIdentityRequiredInvariantTestCoverageHashV2(missingInvariant)
	require.ErrorContains(t, ValidateIdentityRequiredInvariantTestCoverageV2(missingInvariant), "required areas")

	tamperedInvariant := invariant
	tamperedInvariant.ExistingTests = copyIdentityRequiredInvariantTestCoverageMapV2(invariant.ExistingTests)
	tamperedInvariant.ExistingTests[IdentityInvariantCacheSourceVersionV2][0] = "x/identity/types/keepers_v2_test.go:TestACacheSourceVersionTamper"
	require.ErrorContains(t, ValidateIdentityRequiredInvariantTestCoverageV2(tamperedInvariant), "hash mismatch")
}

func TestIdentityRequiredFuzzTestCoverageV2CoversSection174(t *testing.T) {
	coverage := DefaultIdentityRequiredFuzzTestCoverageV2()
	require.NoError(t, ValidateIdentityRequiredFuzzTestCoverageV2(coverage))
	require.Len(t, coverage.RequiredAreas, 11)
	require.NotEmpty(t, coverage.CoverageHash)

	for _, area := range []IdentityFuzzTestCoverageAreaV2{
		IdentityFuzzMalformedNamesV2,
		IdentityFuzzBoundaryLengthNamesV2,
		IdentityFuzzSpoofingPatternCandidatesV2,
		IdentityFuzzCommitmentPreimagesV2,
		IdentityFuzzAuctionBidRevealOrderingV2,
		IdentityFuzzResolverPayloadsV2,
		IdentityFuzzInterfaceDescriptorSchemasV2,
		IdentityFuzzDelegationPermissionCombosV2,
		IdentityFuzzRecursiveProofPathsV2,
		IdentityFuzzReverseResolutionMismatchesV2,
		IdentityFuzzBatchUpdateOrderingV2,
	} {
		require.Contains(t, coverage.RequiredAreas, area)
		require.NotEmpty(t, coverage.ExistingTests[area])
	}

	require.Contains(t, coverage.ExistingTests[IdentityFuzzMalformedNamesV2], "x/identity/types/fuzz_dns_v2_test.go:FuzzIdentityMalformedNamesV2")
	require.Contains(t, coverage.ExistingTests[IdentityFuzzCommitmentPreimagesV2], "x/identity/types/commitment_v2_test.go:FuzzDomainCommitmentV2RevealReplayProtection")
	requireCoverageReferencesExistV2(t, identityFuzzCoverageReferencesV2(coverage))
}

func TestIdentityRequiredPerformanceTestCoverageV2CoversSection175(t *testing.T) {
	coverage := DefaultIdentityRequiredPerformanceTestCoverageV2()
	require.NoError(t, ValidateIdentityRequiredPerformanceTestCoverageV2(coverage))
	require.Len(t, coverage.RequiredAreas, 10)
	require.NotEmpty(t, coverage.CoverageHash)

	for _, area := range []IdentityPerformanceTestCoverageAreaV2{
		IdentityPerformanceDirectResolutionReadLatencyV2,
		IdentityPerformanceRecursiveReadLatencyByDepthV2,
		IdentityPerformanceResolverUpdateWriteLatencyV2,
		IdentityPerformanceBatchResolverUpdatesPerBlockV2,
		IdentityPerformanceBatchRenewalsPerBlockV2,
		IdentityPerformanceDomainRegistrationsPerBlockV2,
		IdentityPerformanceBlockSTMMixedConflictRateV2,
		IdentityPerformanceStoreV2ProofGenerationLatencyV2,
		IdentityPerformanceAdaptiveSyncLargeRecoveryTimeV2,
		IdentityPerformanceExportImportIdentityStateTimeV2,
	} {
		require.Contains(t, coverage.RequiredAreas, area)
		require.NotEmpty(t, coverage.ExistingTests[area])
	}

	require.Contains(t, coverage.ExistingTests[IdentityPerformanceDirectResolutionReadLatencyV2], "x/identity/types/bench_test.go:BenchmarkIdentityStoreV2DirectResolutionReadPath")
	require.Contains(t, coverage.ExistingTests[IdentityPerformanceAdaptiveSyncLargeRecoveryTimeV2], "x/identity/types/bench_test.go:BenchmarkIdentityAdaptiveSyncRecoveryLargeState")
	requireCoverageReferencesExistV2(t, identityPerformanceCoverageReferencesV2(coverage))
}

func TestIdentityRequiredFuzzAndPerformanceCoverageV2RejectBadReferenceAndHashMismatch(t *testing.T) {
	fuzzCoverage := DefaultIdentityRequiredFuzzTestCoverageV2()
	badFuzz := fuzzCoverage
	badFuzz.ExistingTests = copyIdentityRequiredFuzzTestCoverageMapV2(fuzzCoverage.ExistingTests)
	badFuzz.ExistingTests[IdentityFuzzMalformedNamesV2] = []string{"x/identity/types/fuzz_dns_v2_test.go:not_a_fuzz"}
	badFuzz.CoverageHash = ComputeIdentityRequiredFuzzTestCoverageHashV2(badFuzz)
	require.ErrorContains(t, ValidateIdentityRequiredFuzzTestCoverageV2(badFuzz), "invalid test reference")

	tamperedFuzz := fuzzCoverage
	tamperedFuzz.ExistingTests = copyIdentityRequiredFuzzTestCoverageMapV2(fuzzCoverage.ExistingTests)
	tamperedFuzz.ExistingTests[IdentityFuzzBatchUpdateOrderingV2][0] = "x/identity/types/fuzz_dns_v2_test.go:FuzzABatchTamper"
	require.ErrorContains(t, ValidateIdentityRequiredFuzzTestCoverageV2(tamperedFuzz), "hash mismatch")

	performanceCoverage := DefaultIdentityRequiredPerformanceTestCoverageV2()
	missingPerformance := performanceCoverage
	missingPerformance.RequiredAreas = missingPerformance.RequiredAreas[:len(missingPerformance.RequiredAreas)-1]
	missingPerformance.CoverageHash = ComputeIdentityRequiredPerformanceTestCoverageHashV2(missingPerformance)
	require.ErrorContains(t, ValidateIdentityRequiredPerformanceTestCoverageV2(missingPerformance), "required areas")

	tamperedPerformance := performanceCoverage
	tamperedPerformance.ExistingTests = copyIdentityRequiredPerformanceTestCoverageMapV2(performanceCoverage.ExistingTests)
	tamperedPerformance.ExistingTests[IdentityPerformanceDirectResolutionReadLatencyV2][0] = "x/identity/types/bench_test.go:BenchmarkAReadLatencyTamper"
	require.ErrorContains(t, ValidateIdentityRequiredPerformanceTestCoverageV2(tamperedPerformance), "hash mismatch")
}

func copyIdentityRequiredUnitTestCoverageMapV2(in map[IdentityUnitTestCoverageAreaV2][]string) map[IdentityUnitTestCoverageAreaV2][]string {
	out := make(map[IdentityUnitTestCoverageAreaV2][]string, len(in))
	for area, tests := range in {
		out[area] = append([]string(nil), tests...)
	}
	return out
}

func copyIdentityRequiredFuzzTestCoverageMapV2(in map[IdentityFuzzTestCoverageAreaV2][]string) map[IdentityFuzzTestCoverageAreaV2][]string {
	out := make(map[IdentityFuzzTestCoverageAreaV2][]string, len(in))
	for area, tests := range in {
		out[area] = append([]string(nil), tests...)
	}
	return out
}

func copyIdentityRequiredPerformanceTestCoverageMapV2(in map[IdentityPerformanceTestCoverageAreaV2][]string) map[IdentityPerformanceTestCoverageAreaV2][]string {
	out := make(map[IdentityPerformanceTestCoverageAreaV2][]string, len(in))
	for area, tests := range in {
		out[area] = append([]string(nil), tests...)
	}
	return out
}

func copyIdentityRequiredIntegrationTestCoverageMapV2(in map[IdentityIntegrationTestCoverageAreaV2][]string) map[IdentityIntegrationTestCoverageAreaV2][]string {
	out := make(map[IdentityIntegrationTestCoverageAreaV2][]string, len(in))
	for area, tests := range in {
		out[area] = append([]string(nil), tests...)
	}
	return out
}

func copyIdentityRequiredInvariantTestCoverageMapV2(in map[IdentityInvariantTestCoverageAreaV2][]string) map[IdentityInvariantTestCoverageAreaV2][]string {
	out := make(map[IdentityInvariantTestCoverageAreaV2][]string, len(in))
	for area, tests := range in {
		out[area] = append([]string(nil), tests...)
	}
	return out
}

func identityUnitCoverageReferencesV2(coverage IdentityRequiredUnitTestCoverageV2) []string {
	refs := make([]string, 0)
	for _, area := range coverage.RequiredAreas {
		refs = append(refs, coverage.ExistingTests[area]...)
	}
	return refs
}

func identityIntegrationCoverageReferencesV2(coverage IdentityRequiredIntegrationTestCoverageV2) []string {
	refs := make([]string, 0)
	for _, area := range coverage.RequiredAreas {
		refs = append(refs, coverage.ExistingTests[area]...)
	}
	return refs
}

func identityInvariantCoverageReferencesV2(coverage IdentityRequiredInvariantTestCoverageV2) []string {
	refs := make([]string, 0)
	for _, area := range coverage.RequiredAreas {
		refs = append(refs, coverage.ExistingTests[area]...)
	}
	return refs
}

func identityFuzzCoverageReferencesV2(coverage IdentityRequiredFuzzTestCoverageV2) []string {
	refs := make([]string, 0)
	for _, area := range coverage.RequiredAreas {
		refs = append(refs, coverage.ExistingTests[area]...)
	}
	return refs
}

func identityPerformanceCoverageReferencesV2(coverage IdentityRequiredPerformanceTestCoverageV2) []string {
	refs := make([]string, 0)
	for _, area := range coverage.RequiredAreas {
		refs = append(refs, coverage.ExistingTests[area]...)
	}
	return refs
}

func requireCoverageReferencesExistV2(t *testing.T, refs []string) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
	for _, ref := range refs {
		parts := strings.Split(ref, ":")
		require.Len(t, parts, 2, ref)
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(parts[0])))
		require.NoError(t, err, ref)
		fileContent := string(content)
		require.True(t, strings.Contains(fileContent, "func "+parts[1]+"(") || strings.Contains(fileContent, ") "+parts[1]+"("), ref)
	}
}
