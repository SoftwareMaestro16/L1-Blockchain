package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestValidatorIncomeModelIsDeterministicAndBounded(t *testing.T) {
	income, err := ComputeValidatorIncome(ValidatorIncomeInput{
		TotalMintRewards:	sdkmath.NewInt(1_000_000),
		TotalFeeRewards:	sdkmath.NewInt(100_000),
		ValidatorPower:		sdkmath.NewInt(20),
		TotalPower:		sdkmath.NewInt(100),
		CommissionBps:		1_000,
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
		TotalMintRewards:	sdkmath.NewInt(1),
		TotalFeeRewards:	sdkmath.NewInt(1),
		ValidatorPower:		sdkmath.NewInt(101),
		TotalPower:		sdkmath.NewInt(100),
		CommissionBps:		1_000,
	})
	require.ErrorContains(t, err, "validator power must be <= total power")

	require.Error(t, ValidateCommissionBounds(MinCommissionBps-1, 0))
	require.Error(t, ValidateCommissionBounds(MaxCommissionBps+1, 0))
	require.Error(t, ValidateCommissionBounds(MinCommissionBps, MaxDailyCommissionChangeBps+1))
	require.NoError(t, ValidateCommissionBounds(MinCommissionBps, MaxDailyCommissionChangeBps))
}

func TestBalanceControllerRaisesAndLowersInflationWithStaking(t *testing.T) {
	lowStake, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		4_000,
		BlockLoadBps:		DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Greater(t, lowStake.InflationBps, DefaultTargetInflationBps)

	highStake, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		9_000,
		BlockLoadBps:		DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Less(t, highStake.InflationBps, DefaultTargetInflationBps)
}

func TestBalanceControllerCouplesInflationToNetworkActivity(t *testing.T) {
	atTarget, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		DefaultTargetStakeBps,
		BlockLoadBps:		DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	require.Equal(t, DefaultTargetInflationBps, atTarget.InflationBps)
	require.Zero(t, atTarget.ActivityInflationDeltaBps)

	active, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		DefaultTargetStakeBps,
		BlockLoadBps:		BasisPoints,
	})
	require.NoError(t, err)
	require.Equal(t, -DefaultActivityCouplingBps, active.ActivityInflationDeltaBps)
	require.Equal(t, DefaultTargetInflationBps-DefaultActivityCouplingBps, active.InflationBps)
	require.True(t, active.Congested)
	require.True(t, active.RateLimited)
}

func TestBalanceControllerClampsInflationAndBurn(t *testing.T) {
	minOut, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	MinInflationBps,
		StakeRatioBps:		BasisPoints,
		BlockLoadBps:		0,
	})
	require.NoError(t, err)
	require.Equal(t, MinInflationBps, minOut.InflationBps)

	maxOut, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	MaxInflationBps,
		StakeRatioBps:		0,
		BlockLoadBps:		BasisPoints,
		AnnualMint:		sdkmath.NewInt(100),
		AnnualBurn:		sdkmath.NewInt(200),
		AsyncQueueDepth:	10,
		FailedTxRateBps:	1_500,
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
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		DefaultTargetStakeBps,
		BlockLoadBps:		DefaultTargetLoadBps,
		AnnualMint:		sdkmath.NewInt(100),
		AnnualBurn:		sdkmath.NewInt(111),
	}, params)
	require.NoError(t, err)
	require.True(t, out.DeflationGuardActive)
	require.Equal(t, NormalBurnRatioBps-1_000, out.BurnRatioBps)
}

func TestBalanceControllerIncludesProtocolActivityInDeflationGuard(t *testing.T) {
	out, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		DefaultTargetStakeBps,
		BlockLoadBps:		DefaultTargetLoadBps,
		AnnualMint:		sdkmath.NewInt(100),
		AnnualBurn:		sdkmath.NewInt(100),
		Activity: ProtocolEconomicActivity{
			TxFeeNaet:		sdkmath.NewInt(10),
			AVMStorageFeeNaet:	sdkmath.NewInt(10),
			AVMForwardingFeeNaet:	sdkmath.NewInt(3),
			AVMDeploymentCostNaet:	sdkmath.NewInt(3),
		},
	})
	require.NoError(t, err)
	require.True(t, out.DeflationGuardActive)
	require.Equal(t, NormalBurnRatioBps-DeflationGuardStepBps, out.BurnRatioBps)
}

func TestProtocolEconomicFlowPreservesCharges(t *testing.T) {
	flow, err := ComputeProtocolEconomicFlow(ProtocolEconomicFlowInput{
		Activity: ProtocolEconomicActivity{
			TxFeeNaet:		sdkmath.NewInt(100),
			AVMStorageFeeNaet:	sdkmath.NewInt(7),
			AVMForwardingFeeNaet:	sdkmath.NewInt(3),
			AVMDeploymentCostNaet:	sdkmath.NewInt(1_000),
		},
		BurnRatioBps:		2_500,
		TreasuryRatioBps:	1_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_110), flow.TotalChargesNaet)
	require.Equal(t, sdkmath.NewInt(277), flow.BurnNaet)
	require.Equal(t, sdkmath.NewInt(111), flow.TreasuryNaet)
	require.Equal(t, sdkmath.NewInt(722), flow.ValidatorRewardsNaet)
	require.Equal(t, flow.TotalChargesNaet, flow.BurnNaet.Add(flow.TreasuryNaet).Add(flow.ValidatorRewardsNaet))
}

func TestEvaluateEconomicInvariantsAcceptsBoundedAETEconomy(t *testing.T) {
	control, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		DefaultTargetStakeBps,
		BlockLoadBps:		DefaultTargetLoadBps,
	})
	require.NoError(t, err)
	flow, err := ComputeProtocolEconomicFlow(ProtocolEconomicFlowInput{
		Activity:		ProtocolEconomicActivity{TxFeeNaet: sdkmath.NewInt(100)},
		BurnRatioBps:		control.BurnRatioBps,
		TreasuryRatioBps:	TreasuryFeeRatioBps,
	})
	require.NoError(t, err)

	report, err := EvaluateEconomicInvariants(EconomicInvariantInput{
		StakingDenom:			BaseDenom,
		FeeDenom:			BaseDenom,
		RewardDenom:			BaseDenom,
		SlashingDenom:			BaseDenom,
		ExecutionChargeDenom:		BaseDenom,
		CirculatingSupply:		sdkmath.NewInt(1_000_000),
		AnnualMint:			sdkmath.NewInt(30_000),
		AnnualBurn:			sdkmath.NewInt(10_000),
		ControllerOutput:		control,
		FeeFlow:			flow,
		MaxBlockFeeNaet:		sdkmath.NewInt(1_000),
		BlockFeeNaet:			sdkmath.NewInt(100),
		ValidatorRewardsDeterministic:	true,
		FeeComputationDeterministic:	true,
		SlashingDeterministic:		true,
		SlashingAuditable:		true,
		SlashingRewardSafe:		true,
		ControllerParamsExposed:	true,
		ControllerStateExposed:		true,
		ControllerEventsExposed:	true,
		StorageFeePerByteNaet:		sdkmath.NewInt(2),
		LongLivedStorageBytes:		10,
		StorageRetentionPeriods:	2,
		TransientExecutionChargeNaet:	sdkmath.NewInt(10),
	})
	require.NoError(t, err)
	require.True(t, report.Passed)
	require.Empty(t, report.FailedInvariants)
}

func TestEvaluateEconomicInvariantsReportsSpecViolations(t *testing.T) {
	control, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		DefaultTargetStakeBps,
		BlockLoadBps:		DefaultTargetLoadBps,
	})
	require.NoError(t, err)

	report, err := EvaluateEconomicInvariants(EconomicInvariantInput{
		StakingDenom:		"uatom",
		FeeDenom:		"uusdc",
		RewardDenom:		"ureward",
		SlashingDenom:		"uslash",
		ExecutionChargeDenom:	"uexec",
		CirculatingSupply:	sdkmath.NewInt(1_000),
		AnnualMint:		sdkmath.NewInt(200),
		AnnualBurn:		sdkmath.NewInt(500),
		MaxNetIssuanceBps:	100,
		MaxNetBurnBps:		1_000,
		ControllerOutput:	control,
		FeeFlow: ProtocolEconomicFlowOutput{
			TotalChargesNaet:	sdkmath.NewInt(10),
			BurnNaet:		sdkmath.NewInt(3),
			TreasuryNaet:		sdkmath.NewInt(1),
			ValidatorRewardsNaet:	sdkmath.NewInt(5),
		},
		MaxBlockFeeNaet:		sdkmath.NewInt(10),
		BlockFeeNaet:			sdkmath.NewInt(11),
		StorageFeePerByteNaet:		sdkmath.NewInt(1),
		LongLivedStorageBytes:		1,
		StorageRetentionPeriods:	1,
		TransientExecutionChargeNaet:	sdkmath.NewInt(1),
	})
	require.NoError(t, err)
	require.False(t, report.Passed)
	require.ElementsMatch(t, []string{
		"staking_not_aet_primary_asset",
		"fees_not_aet_primary_asset",
		"rewards_not_aet_primary_asset",
		"slashing_not_aet_primary_asset",
		"execution_charges_not_aet_primary_asset",
		"net_burn_outside_bounds",
		"validator_rewards_not_deterministic",
		"fee_computation_not_deterministic",
		"block_fee_exceeds_bound",
		"slashing_invariant_not_satisfied",
		"adaptive_controller_not_observable",
		"economic_flow_not_conservative",
		"storage_pricing_not_above_transient_execution",
	}, report.FailedInvariants)
}

func TestEvaluateEconomicInvariantsRejectsInvalidControllerBounds(t *testing.T) {
	params := DefaultBalanceControllerParams()
	params.MaxBurnRatioBps = BasisPoints
	control := BalanceControllerOutput{
		InflationBps:		DefaultTargetInflationBps,
		BurnRatioBps:		NormalBurnRatioBps,
		ValidatorFeeRatioBps:	BasisPoints - NormalBurnRatioBps - TreasuryFeeRatioBps,
	}
	_, err := EvaluateEconomicInvariants(EconomicInvariantInput{
		StakingDenom:		BaseDenom,
		FeeDenom:		BaseDenom,
		RewardDenom:		BaseDenom,
		SlashingDenom:		BaseDenom,
		ExecutionChargeDenom:	BaseDenom,
		ControllerParams:	params,
		ControllerOutput:	control,
	})
	require.ErrorContains(t, err, "max burn and treasury ratios exceed")
}

func TestEvaluateEconomicWeaknessControlsReportsProductionReadiness(t *testing.T) {
	ready := EvaluateEconomicWeaknessControls(EconomicWeaknessControlInput{
		BurnControllerWired:			true,
		InflationUsesNetworkActivity:		true,
		DeflationGuardEnforced:			true,
		SlashingFlowIntegrated:			true,
		EpochValidatorSelectionProduction:	true,
		AVMFeesInGlobalFeeMarket:		true,
		StateRentOrCleanupIncentive:		true,
		ValidatorReputationInDelegation:	true,
		StakeConcentrationDampening:		true,
		EconomicCircuitBreakerEnabled:		true,
	})
	require.True(t, ready.ProductionReady)
	require.Empty(t, ready.MissingControls)

	missing := EvaluateEconomicWeaknessControls(EconomicWeaknessControlInput{})
	require.False(t, missing.ProductionReady)
	require.ElementsMatch(t, []string{
		"burn_controller_not_wired_to_fee_reward_flow",
		"inflation_controller_not_activity_coupled",
		"deflation_guard_not_enforced",
		"slashing_flow_not_integrated",
		"epoch_validator_selection_not_productionized",
		"avm_fees_not_in_global_market",
		"state_rent_or_cleanup_missing",
		"validator_reputation_not_in_delegation",
		"stake_concentration_dampening_missing",
		"economic_circuit_breaker_missing",
	}, missing.MissingControls)
}

func TestEvaluateInflationRisksFlagsSectionRisks(t *testing.T) {
	report, err := EvaluateInflationRisks(InflationRiskInput{
		CirculatingSupply:		sdkmath.NewInt(1_000_000),
		AnnualMint:			sdkmath.NewInt(80_000),
		AnnualBurn:			sdkmath.NewInt(10_000),
		ValidatorRewardPoolNaet:	sdkmath.NewInt(90),
		ValidatorOperatingCostNaet:	sdkmath.NewInt(100),
		CurrentInflationBps:		DefaultTargetInflationBps,
		StakeRatioBps:			DefaultTargetStakeBps,
		TopValidatorStakeBps:		MaxTopValidatorConcentrationBps + 1,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps - 1,
		ActivitySamplesBps:		[]int64{500, 2_000},
		MaxNetIssuanceBps:		500,
		ActivityNoiseToleranceBps:	1_000,
	})
	require.NoError(t, err)
	require.False(t, report.Stable)
	require.Equal(t, int64(700), report.NetIssuanceBps)
	require.Equal(t, int64(9_000), report.RewardCoverageBps)
	require.Equal(t, int64(1_500), report.ActivityVolatilityBps)
	require.ElementsMatch(t, []string{
		"net_issuance_target_missing",
		"security_overpaid_during_low_activity",
		"validator_security_underpaid",
		"stake_target_risk_not_priced",
		"inflation_activity_signal_noisy",
		"burn_not_integrated_with_issuance",
	}, report.Risks)
}

func TestEvaluateInflationRisksAcceptsBoundedIntegratedModel(t *testing.T) {
	report, err := EvaluateInflationRisks(InflationRiskInput{
		CirculatingSupply:		sdkmath.NewInt(1_000_000),
		AnnualMint:			sdkmath.NewInt(30_000),
		AnnualBurn:			sdkmath.NewInt(10_000),
		ValidatorRewardPoolNaet:	sdkmath.NewInt(100),
		ValidatorOperatingCostNaet:	sdkmath.NewInt(100),
		CurrentInflationBps:		DefaultTargetInflationBps,
		StakeRatioBps:			DefaultTargetStakeBps,
		TopValidatorStakeBps:		MaxTopValidatorConcentrationBps,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps,
		ActivitySamplesBps:		[]int64{DefaultTargetLoadBps - 100, DefaultTargetLoadBps + 100},
		BurnIntegratedWithIssuance:	true,
		NetIssuanceTargetConfigured:	true,
		MaxNetIssuanceBps:		MaxInflationBps,
	})
	require.NoError(t, err)
	require.True(t, report.Stable)
	require.Empty(t, report.Risks)
	require.Equal(t, int64(200), report.ActivityVolatilityBps)
}

func TestEconomicCircuitBreakerFlagsAbnormalActivity(t *testing.T) {
	out, err := EvaluateEconomicCircuitBreaker(EconomicCircuitBreakerInput{
		BlockLoadBps:		HighCongestionLoadBps,
		FeeSpikeBps:		DefaultCircuitBreakerFeeSpikeBps + 1,
		ControllerDriftBps:	DefaultCircuitBreakerControllerDriftBps + 1,
		FailedTxRateBps:	DefaultCircuitBreakerFailedTxRateBps + 1,
		BurnToMintBps:		DeflationGuardBurnToMintBps + 1,
	}, EconomicCircuitBreakerParams{})
	require.NoError(t, err)
	require.True(t, out.Active)
	require.Equal(t, uint64(1), out.CooldownBlocks)
	require.ElementsMatch(t, []string{
		"block_load_abnormal",
		"fee_spike_abnormal",
		"controller_drift_abnormal",
		"failed_tx_rate_abnormal",
		"burn_pressure_abnormal",
	}, out.Reasons)
}

func TestSlashingEconomyFlowRoutesPenaltyDeterministically(t *testing.T) {
	flow, err := ComputeSlashingEconomyFlow(SlashingEconomyFlowInput{
		PenaltyNaet:		sdkmath.NewInt(1_000),
		BurnRatioBps:		2_500,
		TreasuryRatioBps:	1_000,
		ReporterRewardBps:	500,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(250), flow.BurnNaet)
	require.Equal(t, sdkmath.NewInt(100), flow.TreasuryNaet)
	require.Equal(t, sdkmath.NewInt(50), flow.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(600), flow.ValidatorPoolNaet)
	require.Equal(t, flow.PenaltyNaet, flow.BurnNaet.Add(flow.TreasuryNaet).Add(flow.ReporterRewardNaet).Add(flow.ValidatorPoolNaet))
}

func TestStateRentChargesLongLivedStateAndFundsCleanup(t *testing.T) {
	rent, err := ComputeStateRent(StateRentInput{
		StorageBytes:		100,
		RetentionPeriods:	3,
		FeePerByteNaet:		sdkmath.NewInt(2),
		CleanupEligible:	true,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(600), rent.RentNaet)
	require.Equal(t, sdkmath.NewInt(120), rent.CleanupRewardNaet)
	require.Equal(t, sdkmath.NewInt(480), rent.BurnableRentNaet)
}

func TestEvaluateValidatorIncentivesPricesConcentrationAndReliability(t *testing.T) {
	report, err := EvaluateValidatorIncentives(ValidatorIncentiveInput{
		ValidatorStakeBps:		4_000,
		CommissionBps:			500,
		UptimeBps:			9_700,
		MissedBlockRateBps:		100,
		RepeatedFaultCount:		2,
		ReporterRewardBps:		500,
		ReporterRewardsWired:		true,
		SelectionEconomyLinked:		true,
		PerformanceRiskSignals:		true,
		StakeConcentrationDampening:	true,
	})
	require.NoError(t, err)
	require.False(t, report.Healthy)
	require.Equal(t, int64(299), report.ConcentrationPenaltyBps)
	require.Equal(t, int64(750), report.ReliabilityPenaltyBps)
	require.Equal(t, int64(8_951), report.RewardMultiplierBps)
	require.ElementsMatch(t, []string{
		"stake_concentration_reward_dampened",
		"soft_reliability_failure_priced",
	}, report.Findings)
}

func TestEvaluateValidatorIncentivesSupportsNewValidatorBootstrap(t *testing.T) {
	report, err := EvaluateValidatorIncentives(ValidatorIncentiveInput{
		ValidatorStakeBps:		50,
		CommissionBps:			500,
		UptimeBps:			BasisPoints,
		ReporterRewardBps:		500,
		ReporterRewardsWired:		true,
		SelectionEconomyLinked:		true,
		PerformanceRiskSignals:		true,
		StakeConcentrationDampening:	true,
		BootstrapSupportEnabled:	true,
		NewValidator:			true,
	})
	require.NoError(t, err)
	require.True(t, report.Healthy)
	require.Equal(t, DefaultValidatorBootstrapBonusBps, report.BootstrapBonusBps)
	require.Equal(t, BasisPoints+DefaultValidatorBootstrapBonusBps, report.RewardMultiplierBps)
	require.Empty(t, report.Findings)
}

func TestEvaluateValidatorIncentivesReportsMissingEconomicWiring(t *testing.T) {
	report, err := EvaluateValidatorIncentives(ValidatorIncentiveInput{
		ValidatorStakeBps:	4_000,
		CommissionBps:		MinCommissionBps,
		UptimeBps:		BasisPoints,
		NewValidator:		true,
	})
	require.NoError(t, err)
	require.False(t, report.Healthy)
	require.ElementsMatch(t, []string{
		"stake_concentration_not_dampened",
		"validator_behavior_not_visible_to_delegators",
		"reporter_rewards_not_wired",
		"validator_selection_not_linked_to_economics",
		"commission_below_sustainable_floor",
		"new_validator_bootstrap_disadvantage",
	}, report.Findings)
}

func TestEvaluateStakingCentralizationAcceptsDiverseDelegation(t *testing.T) {
	report, err := EvaluateStakingCentralization(StakingCentralizationInput{
		ValidatorStakeBps:		[]int64{1_600, 1_400, 1_200, 1_000, 900, 800},
		CommissionBps:			[]int64{500, 700},
		SelfDelegationRequirementBps:	500,
		RedelegationLagBlocks:		100,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps,
		GovernanceVotingPowerBps:	2_000,
	})
	require.NoError(t, err)
	require.True(t, report.Healthy)
	require.Equal(t, int64(1_600), report.TopValidatorStakeBps)
	require.Equal(t, int64(6_100), report.TopValidatorsStakeBps)
	require.Equal(t, int64(841), report.VotingPowerHHIBps)
	require.Empty(t, report.Risks)
}

func TestEvaluateStakingCentralizationReportsDelegationAndGovernanceRisks(t *testing.T) {
	report, err := EvaluateStakingCentralization(StakingCentralizationInput{
		ValidatorStakeBps:		[]int64{4_000, 2_000, 1_000, 900, 800},
		CommissionBps:			[]int64{MinCommissionBps, 400},
		SelfDelegationRequirementBps:	DefaultMaxSelfDelegationBps + 1,
		RedelegationLagBlocks:		DefaultMaxRedelegationLagBlocks + 1,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps - 1,
		GovernanceVotingPowerBps:	DefaultGovernanceCaptureThresholdBps + 1,
	})
	require.NoError(t, err)
	require.False(t, report.Healthy)
	require.Equal(t, int64(4_000), report.TopValidatorStakeBps)
	require.Equal(t, int64(8_700), report.TopValidatorsStakeBps)
	require.Equal(t, int64(2_245), report.VotingPowerHHIBps)
	require.ElementsMatch(t, []string{
		"delegation_concentrated_in_top_validator",
		"delegation_concentrated_in_visible_validators",
		"commission_race_to_unsustainable_pricing",
		"self_delegation_requirement_reduces_operator_diversity",
		"redelegation_lags_validator_risk",
		"delegator_risk_information_incomplete",
		"voting_power_governance_capture_risk",
	}, report.Risks)
}

func TestEvaluateStakingCentralizationReportsWeakSelfDelegation(t *testing.T) {
	report, err := EvaluateStakingCentralization(StakingCentralizationInput{
		TopValidatorStakeBps:		1_000,
		TopValidatorsStakeBps:		3_000,
		CommissionBps:			[]int64{500},
		SelfDelegationRequirementBps:	DefaultMinSelfDelegationBps - 1,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps,
		GovernanceVotingPowerBps:	1_000,
	})
	require.NoError(t, err)
	require.False(t, report.Healthy)
	require.ElementsMatch(t, []string{"self_delegation_requirement_too_low"}, report.Risks)
}

func TestEvaluateOptimalEconomicStateAcceptsHealthyControlLoop(t *testing.T) {
	state, err := EvaluateOptimalEconomicState(OptimalEconomicStateInput{
		StakeRatioBps:			DefaultTargetStakeBps,
		InflationBps:			DefaultTargetInflationBps,
		ValidatorRewardCoverageBps:	MinValidatorRewardCoverageBps,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps,
		ActiveValidatorCount:		128,
		MinActiveValidatorCount:	64,
		TopValidatorStakeBps:		MaxTopValidatorConcentrationBps,
		BlockLoadBps:			DefaultTargetLoadBps,
		FeeResponseBps:			MinFeeResponseBps,
		SpamCostMultiplierBps:		MinSpamCostMultiplierBps,
		StorageCostCoverageBps:		MinStorageCostCoverageBps,
		BurnToMintBps:			DeflationGuardBurnToMintBps,
		SlashingPenaltyCoverageBps:	MinSlashingPenaltyCoverageBps,
		TreasuryFundingCoverageBps:	MinTreasuryFundingCoverageBps,
	})
	require.NoError(t, err)
	require.True(t, state.Optimal)
	require.Empty(t, state.FailedConditions)
}

func TestEvaluateOptimalEconomicStateReportsUnhealthyConditions(t *testing.T) {
	state, err := EvaluateOptimalEconomicState(OptimalEconomicStateInput{
		StakeRatioBps:			DefaultTargetStakeBps - DefaultStakeTargetToleranceBps - 1,
		InflationBps:			DefaultTargetInflationBps,
		ValidatorRewardCoverageBps:	MinValidatorRewardCoverageBps - 1,
		DelegatorRiskSignalCoverageBps:	MinDelegatorRiskSignalCoverageBps - 1,
		ActiveValidatorCount:		10,
		MinActiveValidatorCount:	64,
		TopValidatorStakeBps:		MaxTopValidatorConcentrationBps + 1,
		BlockLoadBps:			HighCongestionLoadBps,
		FeeResponseBps:			MinFeeResponseBps - 1,
		SpamCostMultiplierBps:		MinSpamCostMultiplierBps - 1,
		StorageCostCoverageBps:		MinStorageCostCoverageBps - 1,
		BurnToMintBps:			DeflationGuardBurnToMintBps + 1,
		SlashingPenaltyCoverageBps:	MinSlashingPenaltyCoverageBps - 1,
		TreasuryFundingCoverageBps:	MinTreasuryFundingCoverageBps - 1,
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
		StakeRatioBps:			DefaultTargetStakeBps,
		StakeTargetToleranceBps:	-1,
		InflationBps:			DefaultTargetInflationBps,
		ValidatorRewardCoverageBps:	MinValidatorRewardCoverageBps,
		ActiveValidatorCount:		1,
		TopValidatorStakeBps:		0,
		FeeResponseBps:			MinFeeResponseBps,
		SpamCostMultiplierBps:		MinSpamCostMultiplierBps,
		StorageCostCoverageBps:		MinStorageCostCoverageBps,
		SlashingPenaltyCoverageBps:	MinSlashingPenaltyCoverageBps,
		TreasuryFundingCoverageBps:	MinTreasuryFundingCoverageBps,
	})
	require.ErrorContains(t, err, "stake_target_tolerance_bps")
}

func TestBalanceControllerRejectsInvalidInputs(t *testing.T) {
	_, err := BalanceController(BalanceControllerInput{
		CurrentInflationBps:	MaxInflationBps + 1,
		StakeRatioBps:		0,
		BlockLoadBps:		0,
	})
	require.ErrorContains(t, err, "current_inflation_bps")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		-1,
		BlockLoadBps:		0,
	})
	require.ErrorContains(t, err, "stake_ratio_bps")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		0,
		BlockLoadBps:		0,
		AnnualMint:		sdkmath.NewInt(-1),
	})
	require.ErrorContains(t, err, "annual mint and burn")

	_, err = BalanceController(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		0,
		BlockLoadBps:		0,
		Activity: ProtocolEconomicActivity{
			AVMStorageFeeNaet: sdkmath.NewInt(-1),
		},
	})
	require.ErrorContains(t, err, "avm_storage_fee_naet")

	params := DefaultBalanceControllerParams()
	params.DeflationGuardBurnToMintBps = BasisPoints - 1
	_, err = BalanceControllerWithParams(BalanceControllerInput{
		CurrentInflationBps:	DefaultTargetInflationBps,
		StakeRatioBps:		0,
		BlockLoadBps:		0,
	}, params)
	require.ErrorContains(t, err, "deflation_guard_burn_to_mint_bps")
}
