package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestValidatorIncomeModelIsDeterministicAndBounded(t *testing.T) {
	income, err := ComputeValidatorIncome(ValidatorIncomeInput{
		TotalMintRewards: sdkmath.NewInt(1_000_000),
		TotalFeeRewards:  sdkmath.NewInt(100_000),
		ValidatorPower:   sdkmath.NewInt(20),
		TotalPower:       sdkmath.NewInt(100),
		CommissionBps:    1_000,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2_000), income.RewardWeightBps)
	require.Equal(t, sdkmath.NewInt(200_000), income.MintRewardShare)
	require.Equal(t, sdkmath.NewInt(20_000), income.FeeRewardShare)
	require.Equal(t, sdkmath.NewInt(22_000), income.ValidatorCommission)
	require.Equal(t, sdkmath.NewInt(242_000), income.ValidatorIncome)
	require.Equal(t, sdkmath.NewInt(198_000), income.DelegatorIncome)
}

func TestValidatorIncomeRejectsUnsafeCommissionAndPower(t *testing.T) {
	_, err := ComputeValidatorIncome(ValidatorIncomeInput{
		TotalMintRewards: sdkmath.NewInt(1),
		TotalFeeRewards:  sdkmath.NewInt(1),
		ValidatorPower:   sdkmath.NewInt(101),
		TotalPower:       sdkmath.NewInt(100),
		CommissionBps:    1_000,
	})
	require.ErrorContains(t, err, "validator power must be <= total power")

	require.Error(t, ValidateCommissionBounds(MinCommissionBps-1, 0))
	require.Error(t, ValidateCommissionBounds(MaxCommissionBps+1, 0))
	require.Error(t, ValidateCommissionBounds(MinCommissionBps, MaxDailyCommissionChangeBps+1))
	require.NoError(t, ValidateCommissionBounds(MinCommissionBps, MaxDailyCommissionChangeBps))
}

func TestBalanceControllerRaisesAndLowersInflationWithStaking(t *testing.T) {
	lowStake, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       4_000,
		BlockLoadBps:        DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Greater(t, lowStake.InflationBps, DefaultTargetInflationBps)

	highStake, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       9_000,
		BlockLoadBps:        DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Less(t, highStake.InflationBps, DefaultTargetInflationBps)
}

func TestBalanceControllerCouplesInflationToNetworkActivity(t *testing.T) {
	atTarget, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       DefaultTargetStakeBps,
		BlockLoadBps:        DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Equal(t, DefaultTargetInflationBps, atTarget.InflationBps)
	require.Zero(t, atTarget.ActivityInflationDeltaBps)

	active, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       DefaultTargetStakeBps,
		BlockLoadBps:        BasisPoints,
	})
	require.NoError(t, err)
	require.Equal(t, -DefaultActivityCouplingBps, active.ActivityInflationDeltaBps)
	require.Equal(t, DefaultTargetInflationBps-DefaultActivityCouplingBps, active.InflationBps)
	require.True(t, active.Congested)
	require.True(t, active.RateLimited)
}

func TestBalanceControllerClampsInflationAndBurn(t *testing.T) {
	minOut, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: MinInflationBps,
		StakeRatioBps:       BasisPoints,
		BlockLoadBps:        0,
	})
	require.NoError(t, err)
	require.Equal(t, MinInflationBps, minOut.InflationBps)

	maxOut, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: MaxInflationBps,
		StakeRatioBps:       0,
		BlockLoadBps:        BasisPoints,
		AnnualMint:          sdkmath.NewInt(100),
		AnnualBurn:          sdkmath.NewInt(200),
		AsyncQueueDepth:     10,
		FailedTxRateBps:     1_500,
	})
	require.NoError(t, err)
	require.Equal(t, MaxInflationBps, maxOut.InflationBps)
	require.Equal(t, int64(3_500), maxOut.BurnRatioBps)
	require.Equal(t, int64(5_500), maxOut.ValidatorFeeRatioBps)
	require.True(t, maxOut.Congested)
	require.True(t, maxOut.DeflationGuardActive)
	require.True(t, maxOut.QueueLimited)
	require.True(t, maxOut.RateLimited)
}

func TestBalanceControllerDeflationGuardIsParameterized(t *testing.T) {
	params := DefaultBalanceControllerParams()
	params.DeflationGuardBurnToMintBps = 11_000
	params.DeflationGuardStepBps = 1_000

	out, err := BalanceControllerWithParams(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       DefaultTargetStakeBps,
		BlockLoadBps:        DefaultTargetLoadBps,
		AnnualMint:          sdkmath.NewInt(100),
		AnnualBurn:          sdkmath.NewInt(111),
	}, params)
	require.NoError(t, err)
	require.True(t, out.DeflationGuardActive)
	require.Equal(t, NormalBurnRatioBps-1_000, out.BurnRatioBps)
}

func TestBalanceControllerIncludesProtocolActivityInDeflationGuard(t *testing.T) {
	out, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       DefaultTargetStakeBps,
		BlockLoadBps:        DefaultTargetLoadBps,
		AnnualMint:          sdkmath.NewInt(100),
		AnnualBurn:          sdkmath.NewInt(100),
		Activity: ProtocolEconomicActivity{
			TxFeeNaet:             sdkmath.NewInt(10),
			AVMStorageFeeNaet:     sdkmath.NewInt(10),
			AVMForwardingFeeNaet:  sdkmath.NewInt(3),
			AVMDeploymentCostNaet: sdkmath.NewInt(3),
		},
	})
	require.NoError(t, err)
	require.True(t, out.DeflationGuardActive)
	require.Equal(t, NormalBurnRatioBps-DeflationGuardStepBps, out.BurnRatioBps)
}

func TestProtocolEconomicFlowPreservesCharges(t *testing.T) {
	flow, err := ComputeProtocolEconomicFlow(ProtocolEconomicFlowInput{
		Activity: ProtocolEconomicActivity{
			TxFeeNaet:             sdkmath.NewInt(100),
			AVMStorageFeeNaet:     sdkmath.NewInt(7),
			AVMForwardingFeeNaet:  sdkmath.NewInt(3),
			AVMDeploymentCostNaet: sdkmath.NewInt(1_000),
		},
		BurnRatioBps:     2_500,
		TreasuryRatioBps: 1_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_110), flow.TotalChargesNaet)
	require.Equal(t, sdkmath.NewInt(277), flow.BurnNaet)
	require.Equal(t, sdkmath.NewInt(111), flow.TreasuryNaet)
	require.Equal(t, sdkmath.NewInt(722), flow.ValidatorRewardsNaet)
	require.Equal(t, flow.TotalChargesNaet, flow.BurnNaet.Add(flow.TreasuryNaet).Add(flow.ValidatorRewardsNaet))
}

func TestEvaluateOptimalEconomicStateAcceptsHealthyControlLoop(t *testing.T) {
	state, err := EvaluateOptimalEconomicState(OptimalEconomicStateInput{
		StakeRatioBps:                  DefaultTargetStakeBps,
		InflationBps:                   DefaultTargetInflationBps,
		ValidatorRewardCoverageBps:     MinValidatorRewardCoverageBps,
		DelegatorRiskSignalCoverageBps: MinDelegatorRiskSignalCoverageBps,
		ActiveValidatorCount:           128,
		MinActiveValidatorCount:        64,
		TopValidatorStakeBps:           MaxTopValidatorConcentrationBps,
		BlockLoadBps:                   DefaultTargetLoadBps,
		FeeResponseBps:                 MinFeeResponseBps,
		SpamCostMultiplierBps:          MinSpamCostMultiplierBps,
		StorageCostCoverageBps:         MinStorageCostCoverageBps,
		BurnToMintBps:                  DeflationGuardBurnToMintBps,
		SlashingPenaltyCoverageBps:     MinSlashingPenaltyCoverageBps,
		TreasuryFundingCoverageBps:     MinTreasuryFundingCoverageBps,
	})
	require.NoError(t, err)
	require.True(t, state.Optimal)
	require.Empty(t, state.FailedConditions)
}

func TestEvaluateOptimalEconomicStateReportsUnhealthyConditions(t *testing.T) {
	state, err := EvaluateOptimalEconomicState(OptimalEconomicStateInput{
		StakeRatioBps:                  DefaultTargetStakeBps - DefaultStakeTargetToleranceBps - 1,
		InflationBps:                   DefaultTargetInflationBps,
		ValidatorRewardCoverageBps:     MinValidatorRewardCoverageBps - 1,
		DelegatorRiskSignalCoverageBps: MinDelegatorRiskSignalCoverageBps - 1,
		ActiveValidatorCount:           10,
		MinActiveValidatorCount:        64,
		TopValidatorStakeBps:           MaxTopValidatorConcentrationBps + 1,
		BlockLoadBps:                   HighCongestionLoadBps,
		FeeResponseBps:                 MinFeeResponseBps - 1,
		SpamCostMultiplierBps:          MinSpamCostMultiplierBps - 1,
		StorageCostCoverageBps:         MinStorageCostCoverageBps - 1,
		BurnToMintBps:                  DeflationGuardBurnToMintBps + 1,
		SlashingPenaltyCoverageBps:     MinSlashingPenaltyCoverageBps - 1,
		TreasuryFundingCoverageBps:     MinTreasuryFundingCoverageBps - 1,
	})
	require.NoError(t, err)
	require.False(t, state.Optimal)
	require.ElementsMatch(t, []string{
		"stake_ratio_outside_target_band",
		"validator_rewards_below_operating_cost",
		"delegator_risk_signals_incomplete",
		"active_validator_set_too_small",
		"validator_stake_too_concentrated",
		"fee_response_outside_predictable_bounds",
		"spam_cost_not_escalating",
		"storage_cost_not_accountable",
		"burn_pressure_exceeds_deflation_guard",
		"slashing_penalties_under_security_damage",
		"treasury_funding_below_maintenance_need",
	}, state.FailedConditions)
}

func TestEvaluateOptimalEconomicStateRejectsInvalidBounds(t *testing.T) {
	_, err := EvaluateOptimalEconomicState(OptimalEconomicStateInput{
		StakeRatioBps:              DefaultTargetStakeBps,
		StakeTargetToleranceBps:    -1,
		InflationBps:               DefaultTargetInflationBps,
		ValidatorRewardCoverageBps: MinValidatorRewardCoverageBps,
		ActiveValidatorCount:       1,
		TopValidatorStakeBps:       0,
		FeeResponseBps:             MinFeeResponseBps,
		SpamCostMultiplierBps:      MinSpamCostMultiplierBps,
		StorageCostCoverageBps:     MinStorageCostCoverageBps,
		SlashingPenaltyCoverageBps: MinSlashingPenaltyCoverageBps,
		TreasuryFundingCoverageBps: MinTreasuryFundingCoverageBps,
	})
	require.ErrorContains(t, err, "stake_target_tolerance_bps")
}

func TestBalanceControllerRejectsInvalidInputs(t *testing.T) {
	_, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps: MaxInflationBps + 1,
		StakeRatioBps:       0,
		BlockLoadBps:        0,
	})
	require.ErrorContains(t, err, "current_inflation_bps")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       -1,
		BlockLoadBps:        0,
	})
	require.ErrorContains(t, err, "stake_ratio_bps")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       0,
		BlockLoadBps:        0,
		AnnualMint:          sdkmath.NewInt(-1),
	})
	require.ErrorContains(t, err, "annual mint and burn")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       0,
		BlockLoadBps:        0,
		Activity: ProtocolEconomicActivity{
			AVMStorageFeeNaet: sdkmath.NewInt(-1),
		},
	})
	require.ErrorContains(t, err, "avm_storage_fee_naet")

	params := DefaultBalanceControllerParams()
	params.DeflationGuardBurnToMintBps = BasisPoints - 1
	_, err = BalanceControllerWithParams(BalanceControllerInput{
		CurrentInflationBps: DefaultTargetInflationBps,
		StakeRatioBps:       0,
		BlockLoadBps:        0,
	}, params)
	require.ErrorContains(t, err, "deflation_guard_burn_to_mint_bps")
}
