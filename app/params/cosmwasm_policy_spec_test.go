package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraCosmWasmLaunchPolicyCoversSection281(t *testing.T) {
	evidence := DefaultAetraCosmWasmLaunchPolicyEvidence()

	report := BuildAetraCosmWasmLaunchPolicyReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraCosmWasmPolicyModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 6, report.Required)
	require.NoError(t, ValidateAetraCosmWasmLaunchPolicy(evidence))
}

func TestAetraCosmWasmLaunchPolicyRejectsMissingPhaseAndAVMBoundary(t *testing.T) {
	evidence := DefaultAetraCosmWasmLaunchPolicyEvidence()
	evidence.AVMRemainsPrimaryRuntime = false
	evidence.OptionalGatedCompatibilityLayer = false
	evidence.MainnetAfterSecurityReview = false
	evidence.PhasePolicies = removeCosmWasmPhase(evidence.PhasePolicies, AetraCosmWasmLaunchPhaseEarlyTestnet)

	report := BuildAetraCosmWasmLaunchPolicyReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraCosmWasmRoleAVMPrimaryRuntime)
	require.Contains(t, report.Failed, AetraCosmWasmRoleOptionalCompatibility)
	require.Contains(t, report.Failed, AetraCosmWasmMainnetAfterSecurityReview)
	require.Contains(t, report.Failed, "launch_policy."+AetraCosmWasmLaunchPhaseEarlyTestnet+":missing")
	require.Error(t, ValidateAetraCosmWasmLaunchPolicy(evidence))
}

func TestAetraCosmWasmLaunchPolicyRejectsWrongDuplicateAndUnexpectedPhase(t *testing.T) {
	evidence := DefaultAetraCosmWasmLaunchPolicyEvidence()
	evidence.ModuleName = "x/wasm"
	evidence.PhasePolicies[1].UploadPolicy = AetraCosmWasmUploadPermissionedOrGovernanceGated
	evidence.PhasePolicies = append(evidence.PhasePolicies,
		AetraCosmWasmLaunchPhasePolicy{Phase: AetraCosmWasmLaunchPhaseMainnet, UploadPolicy: AetraCosmWasmMainnetAfterSecurityReview},
		AetraCosmWasmLaunchPhasePolicy{Phase: "private_mainnet", UploadPolicy: AetraCosmWasmUploadPermissionedOrGovernanceGated},
	)

	report := BuildAetraCosmWasmLaunchPolicyReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraCosmWasmPolicyModuleName)
	require.Contains(t, report.Failed, "launch_policy."+AetraCosmWasmLaunchPhaseLaterTestnet+":wrong_policy")
	require.Contains(t, report.Failed, "launch_policy."+AetraCosmWasmLaunchPhaseMainnet+":duplicate")
	require.Contains(t, report.Failed, "launch_policy.private_mainnet:unexpected")
}

func TestDefaultAetraCosmWasmGasStorageCoversSection282(t *testing.T) {
	evidence := DefaultAetraCosmWasmGasStorageEvidence()

	report := BuildAetraCosmWasmGasStorageReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 11, report.Required)
	for _, control := range []string{
		AetraCosmWasmGasStorageMaxWasmCodeSize,
		AetraCosmWasmGasStorageMaxInstantiateGas,
		AetraCosmWasmGasStorageMaxExecuteGasPerTx,
		AetraCosmWasmGasStorageMaxQueryGas,
		AetraCosmWasmGasStorageStorageRentOrPricing,
		AetraCosmWasmGasStorageContractUploadFee,
		AetraCosmWasmGasStorageMigrationAuthorityRules,
		AetraCosmWasmGasStoragePinnedCodePolicyIfUsed,
	} {
		require.Contains(t, evidence.RequiredControls, control)
	}
	require.NoError(t, ValidateAetraCosmWasmGasStorage(evidence))
}

func TestAetraCosmWasmGasStorageRejectsMissingControlsAndSafetyGates(t *testing.T) {
	evidence := DefaultAetraCosmWasmGasStorageEvidence()
	evidence.RequiredControls = removeString(evidence.RequiredControls,
		AetraCosmWasmGasStorageMaxInstantiateGas,
		AetraCosmWasmGasStorageStorageRentOrPricing,
		AetraCosmWasmGasStorageContractUploadFee,
	)
	evidence.GovernanceConfigurableWithBounds = false
	evidence.DeterministicIntegerAccounting = false
	evidence.SecurityReviewAndBenchmarks = false

	report := BuildAetraCosmWasmGasStorageReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "gas_storage."+AetraCosmWasmGasStorageMaxInstantiateGas+":missing")
	require.Contains(t, report.Failed, "gas_storage."+AetraCosmWasmGasStorageStorageRentOrPricing+":missing")
	require.Contains(t, report.Failed, "gas_storage."+AetraCosmWasmGasStorageContractUploadFee+":missing")
	require.Contains(t, report.Failed, AetraCosmWasmGasStorageGovernanceConfigurable)
	require.Contains(t, report.Failed, AetraCosmWasmGasStorageDeterministicAccounting)
	require.Contains(t, report.Failed, AetraCosmWasmGasStorageSecurityAndBenchmarkGates)
	require.Error(t, ValidateAetraCosmWasmGasStorage(evidence))
}

func TestDefaultAetraCosmWasmTestsCoverImplementationGate(t *testing.T) {
	evidence := DefaultAetraCosmWasmTestEvidence()

	report := BuildAetraCosmWasmTestReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 7, report.Required)
	for _, requiredTest := range []string{
		AetraCosmWasmTestLaunchPolicyCovered,
		AetraCosmWasmTestGasStorageCovered,
		AetraCosmWasmTestUploadFeeCovered,
		AetraCosmWasmTestStoragePricingCovered,
		AetraCosmWasmTestMigrationAuthority,
		AetraCosmWasmTestPinnedCodePolicy,
		AetraCosmWasmTestAVMCompatibilityBoundary,
	} {
		require.Contains(t, evidence.RequiredTests, requiredTest)
	}
	require.NoError(t, ValidateAetraCosmWasmTests(evidence))
}

func TestAetraCosmWasmTestsRejectMissingDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraCosmWasmTestEvidence()
	evidence.ModuleName = ""
	evidence.RequiredTests = removeString(evidence.RequiredTests, AetraCosmWasmTestStoragePricingCovered)
	evidence.RequiredTests = append(evidence.RequiredTests, AetraCosmWasmTestLaunchPolicyCovered, "manual_untracked_smoke")

	report := BuildAetraCosmWasmTestReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, "tests."+AetraCosmWasmTestStoragePricingCovered+":missing")
	require.Contains(t, report.Failed, "tests."+AetraCosmWasmTestLaunchPolicyCovered+":duplicate")
	require.Contains(t, report.Failed, "tests.manual_untracked_smoke:unexpected")
	require.Error(t, ValidateAetraCosmWasmTests(evidence))
}

func TestDefaultAetraCosmWasmContractSecurityTestsCoverSection283(t *testing.T) {
	evidence := DefaultAetraCosmWasmContractSecurityTestEvidence()

	report := BuildAetraCosmWasmContractSecurityTestReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 9, report.Required)
	for _, requiredTest := range []string{
		AetraCosmWasmSecurityTestInfiniteLoopGasLimit,
		AetraCosmWasmSecurityTestLargeStorageWriteBounded,
		AetraCosmWasmSecurityTestFailedContractStateSafe,
		AetraCosmWasmSecurityTestReservedModuleFundsDenied,
		AetraCosmWasmSecurityTestMigrationAuthorization,
		AetraCosmWasmSecurityTestReplySubmessageDeterminism,
		AetraCosmWasmSecurityTestStableEvents,
		AetraCosmWasmSecurityTestExportImportContracts,
		AetraCosmWasmSecurityTestQueryNoStateMutation,
	} {
		require.Contains(t, evidence.RequiredTests, requiredTest)
	}
	require.NoError(t, ValidateAetraCosmWasmContractSecurityTests(evidence))
}

func TestAetraCosmWasmContractSecurityTestsRejectMissingDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraCosmWasmContractSecurityTestEvidence()
	evidence.ModuleName = "x/wasm"
	evidence.RequiredTests = removeString(evidence.RequiredTests,
		AetraCosmWasmSecurityTestReservedModuleFundsDenied,
		AetraCosmWasmSecurityTestQueryNoStateMutation,
	)
	evidence.RequiredTests = append(evidence.RequiredTests,
		AetraCosmWasmSecurityTestInfiniteLoopGasLimit,
		"manual_contract_audit_only",
	)

	report := BuildAetraCosmWasmContractSecurityTestReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraCosmWasmPolicyModuleName)
	require.Contains(t, report.Failed, "contract_security_tests."+AetraCosmWasmSecurityTestReservedModuleFundsDenied+":missing")
	require.Contains(t, report.Failed, "contract_security_tests."+AetraCosmWasmSecurityTestQueryNoStateMutation+":missing")
	require.Contains(t, report.Failed, "contract_security_tests."+AetraCosmWasmSecurityTestInfiniteLoopGasLimit+":duplicate")
	require.Contains(t, report.Failed, "contract_security_tests.manual_contract_audit_only:unexpected")
	require.Error(t, ValidateAetraCosmWasmContractSecurityTests(evidence))
}

func removeCosmWasmPhase(values []AetraCosmWasmLaunchPhasePolicy, target string) []AetraCosmWasmLaunchPhasePolicy {
	out := make([]AetraCosmWasmLaunchPhasePolicy, 0, len(values))
	for _, value := range values {
		if value.Phase != target {
			out = append(out, value)
		}
	}
	return out
}
