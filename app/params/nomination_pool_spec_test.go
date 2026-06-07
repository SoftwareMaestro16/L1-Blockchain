package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraNominationPoolModelCoversSection261(t *testing.T) {
	evidence := DefaultAetraNominationPoolModelEvidence()

	report := BuildAetraNominationPoolModelReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraNominationPoolModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 18, report.Required)
	require.NoError(t, ValidateAetraNominationPoolModel(evidence))
}

func TestAetraNominationPoolModelRejectsMissingRequiredFieldsAndRisks(t *testing.T) {
	evidence := DefaultAetraNominationPoolModelEvidence()
	evidence.ModuleName = ""
	evidence.PoolFields = removeNominationPoolString(evidence.PoolFields, AetraNominationPoolFieldCreatedHeight)
	evidence.PoolDelegationFields = removeNominationPoolString(evidence.PoolDelegationFields, AetraNominationPoolFieldPrincipalEstimate)
	evidence.AccessibilityRiskAcknowledged = false
	evidence.AccountingRiskAcknowledged = false
	evidence.CentralizationRiskAcknowledged = false
	evidence.CurrentImplementationMapped = false

	report := BuildAetraNominationPoolModelReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, AetraNominationPoolStatePool+"."+AetraNominationPoolFieldCreatedHeight+":missing")
	require.Contains(t, report.Failed, AetraNominationPoolStatePoolDelegation+"."+AetraNominationPoolFieldPrincipalEstimate+":missing")
	require.Contains(t, report.Failed, AetraNominationPoolRiskAccessibility)
	require.Contains(t, report.Failed, AetraNominationPoolRiskAccounting)
	require.Contains(t, report.Failed, AetraNominationPoolRiskCentralization)
	require.Contains(t, report.Failed, AetraNominationPoolImplementationMap)
	require.Error(t, ValidateAetraNominationPoolModel(evidence))
}

func TestAetraNominationPoolModelRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraNominationPoolModelEvidence()
	evidence.ModuleName = "x/nomination-pool"
	evidence.PoolFields = append(evidence.PoolFields, AetraNominationPoolFieldPoolID, "OperatorKycStatus")
	evidence.PoolDelegationFields = append(evidence.PoolDelegationFields, AetraNominationPoolFieldShares, "LocalUiEstimate")

	report := BuildAetraNominationPoolModelReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	require.Contains(t, report.Failed, AetraNominationPoolStatePool+"."+AetraNominationPoolFieldPoolID+":duplicate")
	require.Contains(t, report.Failed, AetraNominationPoolStatePool+".OperatorKycStatus:unexpected")
	require.Contains(t, report.Failed, AetraNominationPoolStatePoolDelegation+"."+AetraNominationPoolFieldShares+":duplicate")
	require.Contains(t, report.Failed, AetraNominationPoolStatePoolDelegation+".LocalUiEstimate:unexpected")
	require.Error(t, ValidateAetraNominationPoolModel(evidence))
}

func TestDefaultAetraNominationPoolRequirementsCoverSection262(t *testing.T) {
	evidence := DefaultAetraNominationPoolRequirementsEvidence()

	report := BuildAetraNominationPoolRequirementsReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraNominationPoolModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.NoError(t, ValidateAetraNominationPoolRequirements(evidence))
}

func TestAetraNominationPoolRequirementsRejectMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraNominationPoolRequirementsEvidence()
	evidence.ModuleName = ""
	evidence.UsersDepositNativeStakingDenom = false
	evidence.MintsSharesDeterministically = false
	evidence.DelegatesToValidator = false
	evidence.DistributesRewardsProRata = false
	evidence.CommissionBounded = false
	evidence.WithdrawalFollowsUnbondingPeriod = false
	evidence.SlashingReducesShareValue = false
	evidence.OperatorCannotWithdrawPrincipal = false
	evidence.CannotBypassValidatorPowerCap = false
	evidence.ExposesRiskWarnings = false

	report := BuildAetraNominationPoolRequirementsReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, AetraNominationPoolRequirementNativeStakingDenom)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementDeterministicShareMint)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementDelegatesToValidator)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementProRataRewards)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementCommissionBounded)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementWithdrawalUnbonding)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementSlashingReducesShare)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementOperatorNoPrincipalTheft)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementCannotBypassPowerCap)
	require.Contains(t, report.Failed, AetraNominationPoolRequirementRiskWarnings)
	require.Error(t, ValidateAetraNominationPoolRequirements(evidence))
}

func TestDefaultAetraNominationPoolTestingCoversSection263(t *testing.T) {
	evidence := DefaultAetraNominationPoolTestingEvidence()

	report := BuildAetraNominationPoolTestingReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraNominationPoolModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 12, report.Required)
	require.NoError(t, ValidateAetraNominationPoolTesting(evidence))
}

func TestAetraNominationPoolTestingRejectsMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraNominationPoolTestingEvidence()
	evidence.ModuleName = "x/pool"
	evidence.FirstDepositSharePrice = false
	evidence.SubsequentDepositSharePrice = false
	evidence.RewardDistribution = false
	evidence.CommissionDeduction = false
	evidence.PartialWithdrawal = false
	evidence.FullWithdrawal = false
	evidence.SlashingPoolValidator = false
	evidence.JailedValidator = false
	evidence.RedelegationIfAllowed = false
	evidence.PoolOperatorAbuseAttempt = false
	evidence.ExportImportActiveUnbonding = false
	evidence.RoundingDustHandling = false

	report := BuildAetraNominationPoolTestingReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	require.Contains(t, report.Failed, AetraNominationPoolTestFirstDepositSharePrice)
	require.Contains(t, report.Failed, AetraNominationPoolTestSubsequentDepositSharePrice)
	require.Contains(t, report.Failed, AetraNominationPoolTestRewardDistribution)
	require.Contains(t, report.Failed, AetraNominationPoolTestCommissionDeduction)
	require.Contains(t, report.Failed, AetraNominationPoolTestPartialWithdrawal)
	require.Contains(t, report.Failed, AetraNominationPoolTestFullWithdrawal)
	require.Contains(t, report.Failed, AetraNominationPoolTestSlashingPoolValidator)
	require.Contains(t, report.Failed, AetraNominationPoolTestJailedValidator)
	require.Contains(t, report.Failed, AetraNominationPoolTestRedelegationIfAllowed)
	require.Contains(t, report.Failed, AetraNominationPoolTestOperatorAbuseAttempt)
	require.Contains(t, report.Failed, AetraNominationPoolTestExportImportActiveUnbonding)
	require.Contains(t, report.Failed, AetraNominationPoolTestRoundingDustHandling)
	require.Error(t, ValidateAetraNominationPoolTesting(evidence))
}

func removeNominationPoolString(values []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}

	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if !targetSet[value] {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
