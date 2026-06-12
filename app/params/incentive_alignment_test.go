package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestParticipantIncentiveMapAccountsRewardsPenaltiesAndCosts(t *testing.T) {
	report, err := BuildParticipantIncentiveMap(ParticipantIncentiveMapInput{
		ValidatorRewardsNaet:	sdkmath.NewInt(1_000),
		DelegatorRewardsNaet:	sdkmath.NewInt(2_000),
		UserFeesPaidNaet:	sdkmath.NewInt(500),
		ExecutionFeesNaet:	sdkmath.NewInt(100),
		StorageFeesNaet:	sdkmath.NewInt(200),
		SpamSurchargeNaet:	sdkmath.NewInt(50),
		ValidatorSlashedNaet:	sdkmath.NewInt(100),
		DelegatorSlashedNaet:	sdkmath.NewInt(20),
		ReserveFundingNaet:	sdkmath.NewInt(300),
		BurnedNaet:		sdkmath.NewInt(75),
	})
	require.NoError(t, err)
	require.True(t, report.Passed)
	require.Len(t, report.Entries, 4)
	require.Equal(t, sdkmath.NewInt(3_000), report.TotalRewardsNaet)
	require.Equal(t, sdkmath.NewInt(120), report.TotalPenaltiesNaet)
	require.Equal(t, sdkmath.NewInt(850), report.TotalFeesPaidNaet)
	require.Equal(t, sdkmath.NewInt(300), report.TotalReserveContributionNaet)

	entries := map[string]ParticipantIncentiveEntry{}
	for _, entry := range report.Entries {
		entries[entry.ParticipantClass] = entry
		require.NotEmpty(t, entry.Explanation)
	}
	require.Equal(t, sdkmath.NewInt(900), entries[ParticipantClassValidator].NetPositionNaet)
	require.Equal(t, sdkmath.NewInt(1_980), entries[ParticipantClassDelegator].NetPositionNaet)
	require.Equal(t, sdkmath.NewInt(-850), entries[ParticipantClassUser].NetPositionNaet)
	require.Equal(t, sdkmath.NewInt(300), entries[ParticipantClassProtocolReserve].NetPositionNaet)
}

func TestEpochEconomicReportReconcilesSupplyAccounting(t *testing.T) {
	report, err := GenerateEpochEconomicReport(baseIncentiveEpochInput())
	require.NoError(t, err)
	require.True(t, report.Reconciled)
	require.Empty(t, report.Failed)
	require.Equal(t, sdkmath.NewInt(7_000), report.NetIssuanceNaet)
	require.Equal(t, sdkmath.NewInt(10_000), report.RewardsDistributedNaet)
	require.Equal(t, sdkmath.NewInt(600), report.NetReserveChangeNaet)
	require.Equal(t, int64(1_200), report.StateGrowthBytes)
	require.Equal(t, int64(3_200), report.ValidatorConcentrationBps)
	require.Contains(t, report.GovernanceSummary, "epoch=9")
	require.Equal(t, report.ExpectedEndingSupplyNaet, report.EndingSupplyNaet)
}

func TestEpochEconomicReportDetectsSupplyMismatch(t *testing.T) {
	input := baseIncentiveEpochInput()
	input.EndingSupplyNaet = sdkmath.NewInt(1_006_999)

	report, err := GenerateEpochEconomicReport(input)
	require.NoError(t, err)
	require.False(t, report.Reconciled)
	require.Contains(t, report.Failed, "supply_accounting_mismatch")
}

func TestGovernanceParameterImpactRequiresPreUpgradeSimulation(t *testing.T) {
	report, err := GenerateGovernanceParameterImpactReport(GovernanceParameterImpactInput{
		ParameterName:				ParameterChangeInflation,
		CurrentValueBps:			DefaultTargetInflationBps,
		ProposedValueBps:			250,
		CurrentEpochReport:			baseIncentiveEpochInput(),
		RequirePreUpgradeSimulation:		true,
		ConsensusRewardAccountingPreserved:	true,
	})
	require.NoError(t, err)
	require.False(t, report.ActivationAllowed)
	require.False(t, report.PreUpgradeSimulationIncluded)
	require.Contains(t, report.Failed, "pre_upgrade_simulation_required")
}

func TestGovernanceParameterImpactProjectsBeforeActivation(t *testing.T) {
	report, err := GenerateGovernanceParameterImpactReport(GovernanceParameterImpactInput{
		ParameterName:				ParameterChangeInflation,
		CurrentValueBps:			DefaultTargetInflationBps,
		ProposedValueBps:			250,
		CurrentEpochReport:			baseIncentiveEpochInput(),
		ProjectedEpochs:			3,
		RequirePreUpgradeSimulation:		true,
		ConsensusRewardAccountingPreserved:	true,
	})
	require.NoError(t, err)
	require.True(t, report.ActivationAllowed)
	require.True(t, report.PreUpgradeSimulationIncluded)
	require.Len(t, report.ProjectedReports, 3)
	require.Len(t, report.DashboardRows, 4)
	require.Equal(t, int64(-50), report.DeltaBps)

	first := report.ProjectedReports[0]
	require.Equal(t, sdkmath.NewInt(25_000), first.GrossIssuedNaet)
	require.Equal(t, sdkmath.NewInt(22_000), first.NetIssuanceNaet)
	require.Equal(t, sdkmath.NewInt(1_022_000), first.EndingSupplyNaet)
	require.Equal(t, sdkmath.NewInt(25_000), first.RewardsDistributedNaet)
	require.True(t, report.ProjectedSupplyDeltaNaet.IsPositive())
	require.True(t, report.ProjectedRewardDeltaNaet.IsPositive())
}

func baseIncentiveEpochInput() EpochEconomicReportInput {
	return EpochEconomicReportInput{
		EpochID:			9,
		StartingSupplyNaet:		sdkmath.NewInt(1_000_000),
		EndingSupplyNaet:		sdkmath.NewInt(1_007_000),
		GrossIssuedNaet:		sdkmath.NewInt(10_000),
		BurnedNaet:			sdkmath.NewInt(3_000),
		FeesCollectedNaet:		sdkmath.NewInt(5_000),
		ValidatorRewardsNaet:		sdkmath.NewInt(6_000),
		DelegatorRewardsNaet:		sdkmath.NewInt(4_000),
		SlashedNaet:			sdkmath.NewInt(1_000),
		ReserveInflowNaet:		sdkmath.NewInt(800),
		ReserveOutflowNaet:		sdkmath.NewInt(200),
		StateGrowthBytes:		1_200,
		ValidatorConcentrationBps:	3_200,
		ParticipantInput: ParticipantIncentiveMapInput{
			ValidatorRewardsNaet:	sdkmath.NewInt(6_000),
			DelegatorRewardsNaet:	sdkmath.NewInt(4_000),
			UserFeesPaidNaet:	sdkmath.NewInt(5_000),
			ValidatorSlashedNaet:	sdkmath.NewInt(800),
			DelegatorSlashedNaet:	sdkmath.NewInt(200),
			ReserveFundingNaet:	sdkmath.NewInt(600),
			BurnedNaet:		sdkmath.NewInt(3_000),
		},
	}
}
