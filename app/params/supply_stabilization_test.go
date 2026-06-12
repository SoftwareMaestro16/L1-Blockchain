package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestSupplyStabilizationMovesTowardLowerNetIssuanceWhenActivitySupportsIt(t *testing.T) {
	params := DefaultSupplyStabilizationParams()
	report, err := GenerateSupplyStabilizationReport(SupplyStabilizationInput{
		CurrentSupplyNaet:			sdkmath.NewInt(1_000_000),
		RecentAnnualGrossMintedNaet:		sdkmath.NewInt(30_000),
		RecentAnnualBurnedNaet:			sdkmath.NewInt(5_000),
		RecentAnnualFeeRevenueNaet:		sdkmath.NewInt(20_000),
		RecentAnnualValidatorRewardsNaet:	sdkmath.NewInt(25_000),
		BondedStakeRatioBps:			DefaultTargetStakeBps,
		ActiveValidatorCount:			DefaultActiveValidatorTarget,
		SlashingRateBps:			10,
		ReserveCoverageBps:			BasisPoints,
		ProjectionYears:			3,
		ConsensusRewardAccountingPreserved:	true,
	}, params)
	require.NoError(t, err)
	require.True(t, report.Passed)
	require.Equal(t, SupplyPolicyLowerNetIssuance, report.PolicyDirection)
	require.Equal(t, int64(250), report.CurrentNetIssuanceBps)
	require.Equal(t, int64(200), report.TargetNetIssuanceBps)
	require.Len(t, report.LowerIssuanceConditions, 5)
	for _, condition := range report.LowerIssuanceConditions {
		require.True(t, condition.Met, condition.Name)
	}
	require.Len(t, report.ProjectionYears, 3)
	require.Equal(t, sdkmath.NewInt(1_000_000), report.ProjectionYears[0].StartingSupplyNaet)
	require.Equal(t, sdkmath.NewInt(20_000), report.ProjectionYears[0].ProjectedNetIssuanceNaet)
	require.Equal(t, sdkmath.NewInt(25_000), report.ProjectionYears[0].ProjectedGrossMintedNaet)
	require.Equal(t, sdkmath.NewInt(1_020_000), report.ProjectionYears[0].ProjectedEndingSupplyNaet)
	require.Contains(t, report.GovernanceSummary, "policy=lower_net_issuance")
}

func TestSupplyStabilizationTemporarilyRaisesIssuanceUnderSecurityStress(t *testing.T) {
	params := DefaultSupplyStabilizationParams()
	report, err := GenerateSupplyStabilizationReport(SupplyStabilizationInput{
		CurrentSupplyNaet:			sdkmath.NewInt(1_000_000),
		RecentAnnualGrossMintedNaet:		sdkmath.NewInt(15_000),
		RecentAnnualBurnedNaet:			sdkmath.NewInt(5_000),
		RecentAnnualFeeRevenueNaet:		sdkmath.NewInt(1_000),
		RecentAnnualValidatorRewardsNaet:	sdkmath.NewInt(14_000),
		BondedStakeRatioBps:			4_900,
		ActiveValidatorCount:			50,
		SlashingRateBps:			200,
		ReserveCoverageBps:			5_000,
		ValidatorAttritionBps:			300,
		SecurityRiskBps:			400,
		ProjectionYears:			1,
		ConsensusRewardAccountingPreserved:	true,
	}, params)
	require.NoError(t, err)
	require.True(t, report.Passed)
	require.Equal(t, SupplyPolicyHigherNetIssuance, report.PolicyDirection)
	require.Equal(t, int64(100), report.CurrentNetIssuanceBps)
	require.Equal(t, int64(175), report.TargetNetIssuanceBps)
	require.Len(t, report.ProjectionYears, 1)
	require.Equal(t, sdkmath.NewInt(17_500), report.ProjectionYears[0].ProjectedNetIssuanceNaet)
	require.Equal(t, sdkmath.NewInt(22_500), report.ProjectionYears[0].ProjectedGrossMintedNaet)

	metHigher := map[string]bool{}
	for _, condition := range report.HigherIssuanceConditions {
		metHigher[condition.Name] = condition.Met
	}
	require.True(t, metHigher["low_bonded_stake"])
	require.True(t, metHigher["validator_attrition"])
	require.True(t, metHigher["low_fee_revenue"])
	require.True(t, metHigher["elevated_security_risk"])
}

func TestSupplyProjectionStressScenariosForOneThreeAndFiveYears(t *testing.T) {
	params := DefaultSupplyStabilizationParams()
	for _, years := range []uint32{1, 3, 5} {
		report, err := GenerateSupplyStabilizationReport(SupplyStabilizationInput{
			CurrentSupplyNaet:			sdkmath.NewInt(2_000_000),
			RecentAnnualGrossMintedNaet:		sdkmath.NewInt(60_000),
			RecentAnnualBurnedNaet:			sdkmath.NewInt(20_000),
			RecentAnnualFeeRevenueNaet:		sdkmath.NewInt(50_000),
			RecentAnnualValidatorRewardsNaet:	sdkmath.NewInt(45_000),
			BondedStakeRatioBps:			DefaultTargetStakeBps,
			ActiveValidatorCount:			DefaultActiveValidatorTarget + 10,
			SlashingRateBps:			0,
			ReserveCoverageBps:			15_000,
			ProjectionYears:			years,
			ConsensusRewardAccountingPreserved:	true,
		}, params)
		require.NoError(t, err)
		require.True(t, report.Passed)
		require.Len(t, report.ProjectionYears, int(years))
		require.Equal(t, years, report.ProjectionYears[len(report.ProjectionYears)-1].Year)
		for _, projection := range report.ProjectionYears {
			require.True(t, projection.ProjectedEndingSupplyNaet.GTE(projection.StartingSupplyNaet))
			require.Equal(t, projection.StartingSupplyNaet.Add(projection.ProjectedNetIssuanceNaet), projection.ProjectedEndingSupplyNaet)
			require.Equal(t, projection.ProjectedGrossMintedNaet.Sub(projection.ProjectedBurnedNaet), projection.ProjectedNetIssuanceNaet)
		}
	}
}

func TestSupplyStabilizationFlagsConsensusRewardAccountingBypass(t *testing.T) {
	report, err := GenerateSupplyStabilizationReport(SupplyStabilizationInput{
		CurrentSupplyNaet:			sdkmath.NewInt(1_000_000),
		RecentAnnualGrossMintedNaet:		sdkmath.NewInt(30_000),
		RecentAnnualBurnedNaet:			sdkmath.NewInt(5_000),
		RecentAnnualFeeRevenueNaet:		sdkmath.NewInt(20_000),
		RecentAnnualValidatorRewardsNaet:	sdkmath.NewInt(25_000),
		BondedStakeRatioBps:			DefaultTargetStakeBps,
		ActiveValidatorCount:			DefaultActiveValidatorTarget,
		SlashingRateBps:			10,
		ReserveCoverageBps:			BasisPoints,
		ProjectionYears:			1,
		ConsensusRewardAccountingPreserved:	false,
	}, DefaultSupplyStabilizationParams())
	require.NoError(t, err)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "consensus_reward_accounting_not_preserved")
	require.False(t, report.ConsensusRewardAccountingPreserved)
}

func TestSupplyStabilizationRejectsManualOffChainMissingInputs(t *testing.T) {
	_, err := GenerateSupplyStabilizationReport(SupplyStabilizationInput{
		RecentAnnualGrossMintedNaet:		sdkmath.NewInt(30_000),
		RecentAnnualBurnedNaet:			sdkmath.NewInt(5_000),
		RecentAnnualFeeRevenueNaet:		sdkmath.NewInt(20_000),
		RecentAnnualValidatorRewardsNaet:	sdkmath.NewInt(25_000),
		BondedStakeRatioBps:			DefaultTargetStakeBps,
		ActiveValidatorCount:			DefaultActiveValidatorTarget,
		ProjectionYears:			1,
		ConsensusRewardAccountingPreserved:	true,
	}, DefaultSupplyStabilizationParams())
	require.ErrorContains(t, err, "current_supply_naet must be positive")
}
