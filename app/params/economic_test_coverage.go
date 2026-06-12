package params

import (
	"fmt"
	"sort"
	"strings"
)

const (
	EconomicTestCoverageKindInvariant	= "invariant"
	EconomicTestCoverageKindSimulation	= "simulation"
	EconomicTestCoverageKindUpgrade		= "upgrade"

	EconomicInvariantMintedSupplyReconciles		= "minted_supply_equals_rewards_plus_module_balances"
	EconomicInvariantBurnRemovedFromSpendable	= "burned_supply_removed_from_spendable_supply"
	EconomicInvariantFeeBucketsSum			= "fee_allocation_buckets_sum_to_collected_fees"
	EconomicInvariantSlashedFundsRouteExactly	= "slashed_funds_route_by_configured_splits"
	EconomicInvariantRewardAdjustmentsBounded	= "reward_adjustment_factors_within_bounds"
	EconomicInvariantInflationBounded		= "inflation_within_configured_bounds"
	EconomicInvariantDeflationGuardNetFloor		= "deflation_guard_enforces_net_issuance_floor"
	EconomicInvariantStorageRefundCapped		= "storage_refunds_cannot_exceed_configured_maximum"
	EconomicInvariantMempoolExecutionFeeAligned	= "mempool_execution_fee_checks_cannot_diverge"

	EconomicSimulationLowActivityLowFees		= "low_activity_low_fee_revenue"
	EconomicSimulationNormalTargetStake		= "normal_activity_near_target_stake_ratio"
	EconomicSimulationHighSustainedCongestion	= "high_activity_sustained_congestion"
	EconomicSimulationHighBurnPressure		= "high_burn_pressure"
	EconomicSimulationLowBondedStakeSafety		= "low_bonded_stake_below_safety_threshold"
	EconomicSimulationValidatorConcentration	= "validator_concentration_above_soft_cap"
	EconomicSimulationStakeSplit			= "stake_split_across_multiple_validators"
	EconomicSimulationRepeatedDowntime		= "repeated_validator_downtime"
	EconomicSimulationEquivocationReporter		= "equivocation_with_reporter_reward"
	EconomicSimulationFeeSpamSingleAccount		= "fee_spam_from_one_account"
	EconomicSimulationStateBloatAttack		= "state_bloat_attack"
	EconomicSimulationDeploymentCongestion		= "deployment_congestion"
	EconomicSimulationRapidCommissionIncrease	= "rapid_commission_increase"
	EconomicSimulationSuddenDelegationInflow	= "sudden_delegation_inflow_to_one_validator"

	EconomicUpgradeParameterMigrationPreservesBalances		= "parameter_migration_preserves_existing_balances"
	EconomicUpgradeExistingDelegationsRemainValid			= "existing_delegations_remain_valid"
	EconomicUpgradeExistingValidatorsRemainQueryable		= "existing_validators_remain_queryable"
	EconomicUpgradeFeeDistributionDoesNotStrandBalances		= "fee_distribution_changes_do_not_strand_module_balances"
	EconomicUpgradeBurnActivationKeepsRewardDistribution		= "burn_controller_activation_keeps_reward_distribution"
	EconomicUpgradeStatePricingDefinedStartingState			= "state_pricing_activation_has_defined_starting_state"
	EconomicUpgradeInflationControllerStartsFromCurrentParams	= "inflation_controller_upgrade_starts_from_current_parameters"
)

type EconomicTestCoverageCase struct {
	ID		string
	Kind		string
	Description	string
	Required	bool
	Deterministic	bool
	CIEnabled	bool
	Evidence	[]string
}

type EconomicTestCoverageReport struct {
	InvariantCases		[]EconomicTestCoverageCase
	SimulationCases		[]EconomicTestCoverageCase
	UpgradeCases		[]EconomicTestCoverageCase
	RequiredInvariants	int
	RequiredSimulations	int
	RequiredUpgrades	int
	CoveredInvariants	int
	CoveredSimulations	int
	CoveredUpgrades		int
	InvariantCoverageBps	int64
	SimulationCoverageBps	int64
	UpgradeCoverageBps	int64
	Passed			bool
	Failed			[]string
	GovernanceSummary	string
}

func DefaultRequiredEconomicInvariantCoverageCases() []EconomicTestCoverageCase {
	return []EconomicTestCoverageCase{
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantMintedSupplyReconciles, "minted supply reconciles with rewards and module balances", "TestEpochEconomicReportReconcilesSupplyAccounting"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantBurnRemovedFromSpendable, "burn accounting removes burned supply from spendable supply", "TestBurnIntegratedFeeDistributionGuardsDeflationAndPreservesRewards"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantFeeBucketsSum, "fee allocation buckets sum exactly to collected fees", "TestFeeMarketAllocationSumsExactlyWithCaps"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantSlashedFundsRouteExactly, "slashed funds route exactly by configured burn treasury reporter splits", "TestBurnIntegratedSlashingDistributionAppliesBurnCapWithoutMisrouting"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantRewardAdjustmentsBounded, "reward adjustment factors remain within configured bounds", "TestStakingEnhancementInvariantRejectsRewardAdjustmentOutOfBounds"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantInflationBounded, "inflation remains within configured bounds", "TestActivityInflationControllerClampsToConfiguredBounds"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantDeflationGuardNetFloor, "deflation guard enforces the configured net issuance floor", "TestBurnIntegratedFeeDistributionEnforcesNetIssuanceFloor"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantStorageRefundCapped, "storage refunds cannot exceed configured maximum after decay", "TestDeleteRefundCannotExceedOriginalCostAfterDecayAndCap"),
		requiredCoverageCase(EconomicTestCoverageKindInvariant, EconomicInvariantMempoolExecutionFeeAligned, "mempool and execution fee checks cannot diverge", "TestFeeMarketMempoolAndExecutionValidationAligned"),
	}
}

func DefaultRequiredEconomicSimulationCoverageCases() []EconomicTestCoverageCase {
	return []EconomicTestCoverageCase{
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationLowActivityLowFees, "low activity with low fee revenue", "TestSimulateActivityInflationCoversLowNormalHighAndAdversarialActivity"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationNormalTargetStake, "normal activity near target stake ratio", "TestSimulateActivityInflationCoversLowNormalHighAndAdversarialActivity"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationHighSustainedCongestion, "high activity with sustained congestion", "TestFeeMarketSimulationCoversLowSteadyBurstAndSpamLoad"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationHighBurnPressure, "high burn pressure", "TestAdaptiveInflationStressSupplyAndSecurityBudget"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationLowBondedStakeSafety, "low bonded stake below safety threshold", "TestAdaptiveInflationRaisesUnderLowStakeLowFeesWithinWindowLimit"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationValidatorConcentration, "validator concentration above soft cap", "TestValidatorDistributionSimulationCoversNormalAdversarialAndLowParticipation"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationStakeSplit, "stake split across multiple validators", "TestStakeSplittingDoesNotBypassConcentrationPolicy"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationRepeatedDowntime, "repeated validator downtime", "TestEvaluateValidatorIncentivesPricesConcentrationAndReliability"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationEquivocationReporter, "equivocation with reporter reward", "TestEvidenceRoutingCapsReporterRewardsAndRejectsDuplicates"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationFeeSpamSingleAccount, "fee spam from one account", "TestFeeMarketSimulationCoversLowSteadyBurstAndSpamLoad"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationStateBloatAttack, "state bloat attack", "TestEconomicAttackPreventionModelsEveryAttackClass"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationDeploymentCongestion, "deployment congestion", "TestDeploymentFeeEstimateWithinTolerance"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationRapidCommissionIncrease, "rapid commission increase", "TestValidatorReputationCaptureAndConcentrationWarnings"),
		requiredCoverageCase(EconomicTestCoverageKindSimulation, EconomicSimulationSuddenDelegationInflow, "sudden delegation inflow to one validator", "TestStakeMovementMonitorFlagsAbnormalMovements"),
	}
}

func DefaultRequiredEconomicUpgradeCoverageCases() []EconomicTestCoverageCase {
	return []EconomicTestCoverageCase{
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeParameterMigrationPreservesBalances, "parameter migration preserves existing balances", "TestCustomModuleMigrationsFromV1ToCurrent", "TestMigrationPhase0ReadinessPassesBaselineHardening"),
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeExistingDelegationsRemainValid, "existing delegations remain valid after upgrade", "TestMigrationPhase0ReadinessPassesBaselineHardening", "TestKeeperIntegrationManifestCoversRequiredStoresHooksAndMigrations"),
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeExistingValidatorsRemainQueryable, "existing validators remain queryable after upgrade", "TestKeeperIntegrationManifestCoversRequiredStoresHooksAndMigrations", "TestPosMigrationStrategyCoversAllActivationPhases"),
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeFeeDistributionDoesNotStrandBalances, "fee distribution changes do not strand module balances", "TestFeeMarketAllocationSumsExactlyWithCaps", "TestFeesMigrationSucceedsOnValidState"),
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeBurnActivationKeepsRewardDistribution, "burn controller activation does not break reward distribution", "TestBurnIntegratedFeeDistributionGuardsDeflationAndPreservesRewards", "TestBurnIntegratedFeeDistributionEnforcesNetIssuanceFloor"),
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeStatePricingDefinedStartingState, "state pricing activation has a defined starting state", "TestStorageFootprintIsQueryableForAccountsAndContracts", "TestStorageRentStatusWarningAndRecoveryPath"),
		requiredCoverageCase(EconomicTestCoverageKindUpgrade, EconomicUpgradeInflationControllerStartsFromCurrentParams, "inflation controller upgrade starts from current inflation parameters", "TestActivityInflationControllerEmergencyFreezeHoldsCurrentInflation", "TestGovernanceParameterImpactRequiresPreUpgradeSimulation"),
	}
}

func BuildRequiredEconomicTestCoverageReport(invariantCases, simulationCases []EconomicTestCoverageCase, upgradeCases ...[]EconomicTestCoverageCase) EconomicTestCoverageReport {
	if invariantCases == nil {
		invariantCases = DefaultRequiredEconomicInvariantCoverageCases()
	}
	if simulationCases == nil {
		simulationCases = DefaultRequiredEconomicSimulationCoverageCases()
	}
	upgradesInput := DefaultRequiredEconomicUpgradeCoverageCases()
	if len(upgradeCases) > 0 && upgradeCases[0] != nil {
		upgradesInput = upgradeCases[0]
	}

	invariants, invariantFailed, requiredInvariants, coveredInvariants := evaluateCoverageCases(EconomicTestCoverageKindInvariant, invariantCases, requiredInvariantCoverageIDs())
	simulations, simulationFailed, requiredSimulations, coveredSimulations := evaluateCoverageCases(EconomicTestCoverageKindSimulation, simulationCases, requiredSimulationCoverageIDs())
	upgrades, upgradeFailed, requiredUpgrades, coveredUpgrades := evaluateCoverageCases(EconomicTestCoverageKindUpgrade, upgradesInput, requiredUpgradeCoverageIDs())
	failed := append(invariantFailed, simulationFailed...)
	failed = append(failed, upgradeFailed...)
	sort.Strings(failed)

	invariantCoverage := coverageBps(coveredInvariants, requiredInvariants)
	simulationCoverage := coverageBps(coveredSimulations, requiredSimulations)
	upgradeCoverage := coverageBps(coveredUpgrades, requiredUpgrades)
	return EconomicTestCoverageReport{
		InvariantCases:		invariants,
		SimulationCases:	simulations,
		UpgradeCases:		upgrades,
		RequiredInvariants:	requiredInvariants,
		RequiredSimulations:	requiredSimulations,
		RequiredUpgrades:	requiredUpgrades,
		CoveredInvariants:	coveredInvariants,
		CoveredSimulations:	coveredSimulations,
		CoveredUpgrades:	coveredUpgrades,
		InvariantCoverageBps:	invariantCoverage,
		SimulationCoverageBps:	simulationCoverage,
		UpgradeCoverageBps:	upgradeCoverage,
		Passed:			len(failed) == 0 && invariantCoverage == BasisPoints && simulationCoverage == BasisPoints && upgradeCoverage == BasisPoints,
		Failed:			failed,
		GovernanceSummary:	fmt.Sprintf("required_invariants=%d/%d required_simulations=%d/%d required_upgrades=%d/%d invariant_coverage_bps=%d simulation_coverage_bps=%d upgrade_coverage_bps=%d", coveredInvariants, requiredInvariants, coveredSimulations, requiredSimulations, coveredUpgrades, requiredUpgrades, invariantCoverage, simulationCoverage, upgradeCoverage),
	}
}

func requiredCoverageCase(kind, id, description string, evidence ...string) EconomicTestCoverageCase {
	return EconomicTestCoverageCase{
		ID:		id,
		Kind:		kind,
		Description:	description,
		Required:	true,
		Deterministic:	true,
		CIEnabled:	true,
		Evidence:	append([]string{}, evidence...),
	}
}

func evaluateCoverageCases(kind string, cases []EconomicTestCoverageCase, expectedIDs []string) ([]EconomicTestCoverageCase, []string, int, int) {
	out := append([]EconomicTestCoverageCase{}, cases...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	failed := make([]string, 0)
	seen := make(map[string]EconomicTestCoverageCase, len(out))
	for i, item := range out {
		if item.ID == "" {
			failed = append(failed, kind+":case_id_required")
			continue
		}
		if item.Kind != kind {
			failed = append(failed, item.ID+":wrong_coverage_kind")
		}
		if _, ok := seen[item.ID]; ok {
			failed = append(failed, item.ID+":duplicate_coverage_case")
		}
		seen[item.ID] = item
		if item.Required {
			if strings.TrimSpace(item.Description) == "" {
				failed = append(failed, item.ID+":description_missing")
			}
			if len(item.Evidence) == 0 {
				failed = append(failed, item.ID+":evidence_missing")
			}
			if !item.Deterministic {
				failed = append(failed, item.ID+":not_deterministic")
			}
			if !item.CIEnabled {
				failed = append(failed, item.ID+":not_ci_enabled")
			}
			for j, evidence := range item.Evidence {
				if strings.TrimSpace(evidence) == "" {
					failed = append(failed, fmt.Sprintf("%s:evidence_%d_blank", item.ID, j))
				}
			}
		}
		_ = i
	}

	required := len(expectedIDs)
	covered := 0
	for _, id := range expectedIDs {
		item, ok := seen[id]
		if !ok {
			failed = append(failed, id+":missing_required_coverage")
			continue
		}
		if item.Required && item.Kind == kind && item.Deterministic && item.CIEnabled && len(item.Evidence) > 0 {
			covered++
		}
	}
	return out, failed, required, covered
}

func coverageBps(covered, required int) int64 {
	if required == 0 {
		return BasisPoints
	}
	return int64(covered) * BasisPoints / int64(required)
}

func requiredInvariantCoverageIDs() []string {
	return []string{
		EconomicInvariantMintedSupplyReconciles,
		EconomicInvariantBurnRemovedFromSpendable,
		EconomicInvariantFeeBucketsSum,
		EconomicInvariantSlashedFundsRouteExactly,
		EconomicInvariantRewardAdjustmentsBounded,
		EconomicInvariantInflationBounded,
		EconomicInvariantDeflationGuardNetFloor,
		EconomicInvariantStorageRefundCapped,
		EconomicInvariantMempoolExecutionFeeAligned,
	}
}

func requiredSimulationCoverageIDs() []string {
	return []string{
		EconomicSimulationLowActivityLowFees,
		EconomicSimulationNormalTargetStake,
		EconomicSimulationHighSustainedCongestion,
		EconomicSimulationHighBurnPressure,
		EconomicSimulationLowBondedStakeSafety,
		EconomicSimulationValidatorConcentration,
		EconomicSimulationStakeSplit,
		EconomicSimulationRepeatedDowntime,
		EconomicSimulationEquivocationReporter,
		EconomicSimulationFeeSpamSingleAccount,
		EconomicSimulationStateBloatAttack,
		EconomicSimulationDeploymentCongestion,
		EconomicSimulationRapidCommissionIncrease,
		EconomicSimulationSuddenDelegationInflow,
	}
}

func requiredUpgradeCoverageIDs() []string {
	return []string{
		EconomicUpgradeParameterMigrationPreservesBalances,
		EconomicUpgradeExistingDelegationsRemainValid,
		EconomicUpgradeExistingValidatorsRemainQueryable,
		EconomicUpgradeFeeDistributionDoesNotStrandBalances,
		EconomicUpgradeBurnActivationKeepsRewardDistribution,
		EconomicUpgradeStatePricingDefinedStartingState,
		EconomicUpgradeInflationControllerStartsFromCurrentParams,
	}
}
