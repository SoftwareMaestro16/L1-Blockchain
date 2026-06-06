package types

import (
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

func copyIdentityRequiredUnitTestCoverageMapV2(in map[IdentityUnitTestCoverageAreaV2][]string) map[IdentityUnitTestCoverageAreaV2][]string {
	out := make(map[IdentityUnitTestCoverageAreaV2][]string, len(in))
	for area, tests := range in {
		out[area] = append([]string(nil), tests...)
	}
	return out
}
