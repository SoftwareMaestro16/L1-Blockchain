package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestBurnIntegratedFeeDistributionGuardsDeflationAndPreservesRewards(t *testing.T) {
	params := DefaultBurnMechanicsParams()
	params.EpochBurnCapNaet = sdkmath.NewInt(200)
	params.NetIssuanceFloorNaet = sdkmath.NewInt(100)
	params.SecurityRewardFloorNaet = sdkmath.NewInt(400)
	params.FeeSpikeThresholdBps = 15_000

	out, err := ComputeBurnIntegratedFeeDistribution(BurnIntegratedFeeDistributionInput{
		EpochID:			7,
		BlockHeight:			100,
		CollectedFeesNaet:		sdkmath.NewInt(1_000),
		BurnRatioBps:			5_000,
		CommunityPoolRatioBps:		1_000,
		StateMaintenanceReserveBps:	1_000,
		GrossMintedNaet:		sdkmath.NewInt(300),
		CumulativeBurnedNaet:		sdkmath.NewInt(1_000),
		FeeSpikeBps:			25_000,
		BondedStakeRatioBps:		DefaultTargetStakeBps,
		Params:				params,
	})
	require.NoError(t, err)
	require.Zero(t, out.BurnNaet.Int64())
	require.Equal(t, sdkmath.NewInt(400), out.ValidatorRewardNaet)
	require.Equal(t, sdkmath.NewInt(100), out.CommunityPoolNaet)
	require.Equal(t, sdkmath.NewInt(100), out.StateMaintenanceReserveNaet)
	require.Equal(t, sdkmath.NewInt(400), out.DeflationReserveNaet)
	require.True(t, out.DeflationGuard.Active)
	require.ElementsMatch(t, []string{
		"security_reward_floor_priority",
		"burn_cap_applied",
		"fee_spike_diverted_to_reserve",
	}, out.DeflationGuard.Reasons)

	invariants := ValidateBurnFeeDistributionInvariants(out, params)
	require.True(t, invariants.Passed)
	require.Empty(t, invariants.Failed)
	require.Empty(t, out.Events)
}

func TestBurnIntegratedFeeDistributionEnforcesNetIssuanceFloor(t *testing.T) {
	params := DefaultBurnMechanicsParams()
	params.NetIssuanceFloorNaet = sdkmath.NewInt(50)

	out, err := ComputeBurnIntegratedFeeDistribution(BurnIntegratedFeeDistributionInput{
		EpochID:		8,
		BlockHeight:		120,
		CollectedFeesNaet:	sdkmath.NewInt(1_000),
		BurnRatioBps:		5_000,
		GrossMintedNaet:	sdkmath.NewInt(100),
		CumulativeBurnedNaet:	sdkmath.NewInt(2_000),
		BondedStakeRatioBps:	DefaultTargetStakeBps,
		Params:			params,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(50), out.BurnNaet)
	require.Equal(t, sdkmath.NewInt(450), out.DeflationReserveNaet)
	require.Equal(t, sdkmath.NewInt(50), out.DeflationGuard.NetIssuanceAfterBurnNaet)
	require.Contains(t, out.DeflationGuard.Reasons, "net_issuance_floor_guard")
	require.Len(t, out.Events, 1)
	require.True(t, out.Events[0].RemovedFromSpendableSupply)
	require.Equal(t, sdkmath.NewInt(2_050), out.CumulativeBurnedNaet)

	invariants := ValidateBurnFeeDistributionInvariants(out, params)
	require.True(t, invariants.Passed)
}

func TestBurnFloorDisabledByDefault(t *testing.T) {
	params := DefaultBurnMechanicsParams()
	require.True(t, params.BurnFloorNaet.IsZero())

	out, err := ComputeBurnIntegratedFeeDistribution(BurnIntegratedFeeDistributionInput{
		EpochID:		9,
		BlockHeight:		140,
		CollectedFeesNaet:	sdkmath.NewInt(1_000),
		BurnRatioBps:		0,
		GrossMintedNaet:	sdkmath.NewInt(1_000),
		BondedStakeRatioBps:	DefaultTargetStakeBps,
		Params:			params,
	})
	require.NoError(t, err)
	require.True(t, out.BurnNaet.IsZero())
	require.True(t, out.CumulativeBurnedNaet.IsZero())
	require.Empty(t, out.Events)
}

func TestBurnIntegratedSlashingDistributionAppliesBurnCapWithoutMisrouting(t *testing.T) {
	params := DefaultBurnMechanicsParams()
	params.EpochBurnCapNaet = sdkmath.NewInt(200)
	out, err := ComputeBurnIntegratedSlashingDistribution(BurnIntegratedSlashingDistributionInput{
		EpochID:		10,
		BlockHeight:		160,
		PenaltyNaet:		sdkmath.NewInt(1_000),
		BurnRatioBps:		5_000,
		TreasuryRatioBps:	1_000,
		ReporterRewardBps:	500,
		GrossMintedNaet:	sdkmath.NewInt(1_000),
		CumulativeBurnedNaet:	sdkmath.NewInt(3_000),
		Params:			params,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(200), out.BurnNaet)
	require.Equal(t, sdkmath.NewInt(300), out.DeflationReserveNaet)
	require.Equal(t, sdkmath.NewInt(100), out.TreasuryNaet)
	require.Equal(t, sdkmath.NewInt(50), out.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(350), out.ValidatorPoolNaet)
	require.Equal(t, sdkmath.NewInt(3_200), out.CumulativeBurnedNaet)
	require.Contains(t, out.DeflationGuard.Reasons, "burn_cap_applied")
	require.Len(t, out.Events, 1)
	require.True(t, out.Events[0].RemovedFromSpendableSupply)

	invariants := ValidateBurnSlashingDistributionInvariants(out, params)
	require.True(t, invariants.Passed)
	require.Empty(t, invariants.Failed)
}

func TestBurnSupplyQueryReportsCumulativeAndRecentRate(t *testing.T) {
	query, err := QueryBurnSupply(BurnSupplyQueryInput{
		CumulativeBurnedNaet:	sdkmath.NewInt(5_000),
		CurrentBlockHeight:	200,
		RecentWindowBlocks:	20,
		Events: []BurnAccountingEvent{
			{BlockHeight: 170, BurnedNaet: sdkmath.NewInt(1_000)},
			{BlockHeight: 185, BurnedNaet: sdkmath.NewInt(600)},
			{BlockHeight: 199, BurnedNaet: sdkmath.NewInt(400)},
		},
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(5_000), query.CumulativeBurnedNaet)
	require.Equal(t, sdkmath.NewInt(1_000), query.RecentBurnedNaet)
	require.Equal(t, sdkmath.NewInt(50), query.RecentBurnRateNaetPerBlock)
	require.Equal(t, 3, query.EventCount)
}

func TestDeflationGuardCanBeExplicitlyRelaxedByGovernance(t *testing.T) {
	params := DefaultBurnMechanicsParams()
	params.NetIssuanceFloorNaet = sdkmath.NewInt(50)
	params.GovernanceAllowsBelowFloor = true

	out, err := ComputeBurnIntegratedFeeDistribution(BurnIntegratedFeeDistributionInput{
		EpochID:		11,
		BlockHeight:		220,
		CollectedFeesNaet:	sdkmath.NewInt(1_000),
		BurnRatioBps:		5_000,
		GrossMintedNaet:	sdkmath.NewInt(100),
		BondedStakeRatioBps:	DefaultTargetStakeBps,
		Params:			params,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(500), out.BurnNaet)
	require.Equal(t, sdkmath.NewInt(-400), out.DeflationGuard.NetIssuanceAfterBurnNaet)
	require.NotContains(t, out.DeflationGuard.Reasons, "net_issuance_floor_guard")
}
