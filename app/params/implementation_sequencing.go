package params

import (
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	EconomicSequencePhase0	= "phase_0_measurement_accounting"
	EconomicSequencePhase1	= "phase_1_production_safety"
	EconomicSequencePhase2	= "phase_2_validator_delegation_incentives"
	EconomicSequencePhase3	= "phase_3_fee_storage_execution_optimization"
	EconomicSequencePhase4	= "phase_4_adaptive_controllers_long_term_stabilization"

	SequencingStatusReady	= "ready"
	SequencingStatusBlocked	= "blocked"

	SequencingTaskNetIssuanceAccounting			= "net_issuance_accounting"
	SequencingTaskCumulativeBurnAccounting			= "cumulative_burn_accounting"
	SequencingTaskFeeBucketAccounting			= "fee_bucket_accounting"
	SequencingTaskValidatorConcentrationMetrics		= "validator_concentration_metrics"
	SequencingTaskStateGrowthTelemetry			= "state_growth_telemetry"
	SequencingTaskValidatorRewardPerVotingPower		= "validator_reward_per_voting_power_telemetry"
	SequencingTaskEpochEconomicReportGeneration		= "epoch_economic_report_generation"
	SequencingTaskBurnControllerFeeDistributionWiring	= "burn_controller_fee_distribution_wiring"
	SequencingTaskDeflationGuardEnforcement			= "deflation_guard_enforcement"
	SequencingTaskBurnCaps					= "burn_caps"
	SequencingTaskFeeControllerBounds			= "fee_controller_bounds"
	SequencingTaskMempoolExecutionFeeValidationAlignment	= "mempool_execution_fee_validation_alignment"
	SequencingTaskSlashingFundRouting			= "slashing_fund_routing"
	SequencingTaskFundMovementInvariantTests		= "fund_movement_invariant_tests"
	SequencingTaskValidatorScoring				= "validator_scoring"
	SequencingTaskEpochBasedSelectionProduction		= "epoch_based_selection_production"
	SequencingTaskConcentrationRewardDampening		= "concentration_reward_dampening"
	SequencingTaskValidatorRiskScoreQueries			= "validator_risk_score_queries"
	SequencingTaskCommissionChangeWarnings			= "commission_change_warnings"
	SequencingTaskRiskAdjustedYieldEstimates		= "risk_adjusted_yield_estimates"
	SequencingTaskValidatorBootstrapBand			= "validator_bootstrap_band"
	SequencingTaskResourceSpecificFeeMultipliers		= "resource_specific_fee_multipliers"
	SequencingTaskSenderLocalSpamSurcharge			= "sender_local_spam_surcharge"
	SequencingTaskStorageWriteUpdatePricing			= "storage_write_update_pricing"
	SequencingTaskStorageFootprintQueries			= "storage_footprint_queries"
	SequencingTaskDeleteRefundPolicy			= "delete_refund_policy"
	SequencingTaskDeploymentForwardingFeeEstimation		= "deployment_forwarding_fee_estimation"
	SequencingTaskStateGrowthSurcharge			= "state_growth_surcharge"
	SequencingTaskAdaptiveInflationController		= "adaptive_inflation_controller"
	SequencingTaskSupplyProjectionReports			= "supply_projection_reports"
	SequencingTaskEconomicSecurityModule			= "economic_security_module"
	SequencingTaskFeeMarketCircuitBreaker			= "fee_market_circuit_breaker"
	SequencingTaskSecurityReserveAccounting			= "security_reserve_accounting"
	SequencingTaskPreUpgradeEconomicSimulationRequirement	= "pre_upgrade_economic_simulation_requirement"
)

type EconomicSequencingTask struct {
	PhaseID		string
	Name		string
	Implemented	bool
	Observable	bool
	Queryable	bool
	InvariantTested	bool
	Reconciled	bool
	Evidence	[]string
}

type EconomicSequencingExitCriterion struct {
	Name		string
	Met		bool
	FailedReason	string
}

type EconomicSequencingPhaseReport struct {
	PhaseID		string
	Goal		string
	Status		string
	Tasks		[]EconomicSequencingTask
	ExitCriteria	[]EconomicSequencingExitCriterion
	Failed		[]string
	Passed		bool
}

type EconomicSequencingReport struct {
	Phases				[]EconomicSequencingPhaseReport
	ReadyForPhase1			bool
	ReadyForPhase2			bool
	ReadyForPhase3			bool
	ReadyForPhase4			bool
	ReadyForIncentiveChanges	bool
	ReadyForResourcePricing		bool
	ReadyForLongTermStabilization	bool
	Failed				[]string
	Passed				bool
	GovernanceSummary		string
}

type EconomicSequencingMetricInput struct {
	EpochReport			EpochEconomicReport
	BurnSupply			BurnSupplyQueryOutput
	FeeAllocation			FeeAllocationBuckets
	SecurityReport			EconomicSecurityEpochReport
	StateGrowth			StateGrowthTelemetryOutput
	FeeOptimizer			FeeMarketOptimizerOutput
	BurnFeeDistribution		BurnIntegratedFeeDistributionOutput
	BurnSlashingDistribution	BurnIntegratedSlashingDistributionOutput
	BurnFeeInvariant		BurnAccountingInvariantReport
	BurnSlashingInvariant		BurnAccountingInvariantReport
	AdaptiveInflation		AdaptiveInflationEpochReport
	ValidatorRewardTelemetry	[]ValidatorRewardTelemetrySample
}

type ValidatorRewardTelemetrySample struct {
	ValidatorID	string
	VotingPowerBps	int64
	RewardNaet	sdkmath.Int
}

type ValidatorRewardPerPowerTelemetry struct {
	ValidatorID		string
	VotingPowerBps		int64
	RewardNaet		sdkmath.Int
	RewardPerBpsNaet	sdkmath.Int
}

func DefaultPhase0SequencingTasks() []EconomicSequencingTask {
	return []EconomicSequencingTask{
		readySequencingTask(EconomicSequencePhase0, SequencingTaskNetIssuanceAccounting, true, false, []string{"adaptive_inflation_epoch_report", "epoch_economic_report"}),
		readySequencingTask(EconomicSequencePhase0, SequencingTaskCumulativeBurnAccounting, true, true, []string{"burn_supply_query", "burn_accounting_events"}),
		readySequencingTask(EconomicSequencePhase0, SequencingTaskFeeBucketAccounting, true, true, []string{"fee_market_allocation_buckets", "fee_bucket_sum_invariant"}),
		readySequencingTask(EconomicSequencePhase0, SequencingTaskValidatorConcentrationMetrics, false, true, []string{"staking_concentration_metrics", "validator_reputation_queries"}),
		readySequencingTask(EconomicSequencePhase0, SequencingTaskStateGrowthTelemetry, false, true, []string{"state_growth_telemetry", "state_growth_alerts"}),
		readySequencingTask(EconomicSequencePhase0, SequencingTaskValidatorRewardPerVotingPower, false, true, []string{"validator_reward_per_voting_power_telemetry"}),
		readySequencingTask(EconomicSequencePhase0, SequencingTaskEpochEconomicReportGeneration, true, true, []string{"epoch_economic_report", "participant_incentive_map"}),
	}
}

func DefaultPhase1SequencingTasks() []EconomicSequencingTask {
	return []EconomicSequencingTask{
		readySequencingTask(EconomicSequencePhase1, SequencingTaskBurnControllerFeeDistributionWiring, true, true, []string{"integrated_fee_distribution_burn_path"}),
		readySequencingTask(EconomicSequencePhase1, SequencingTaskDeflationGuardEnforcement, true, true, []string{"deflation_guard_status", "security_reward_floor_priority"}),
		readySequencingTask(EconomicSequencePhase1, SequencingTaskBurnCaps, true, true, []string{"epoch_burn_cap", "burn_cap_remainder"}),
		readySequencingTask(EconomicSequencePhase1, SequencingTaskFeeControllerBounds, true, true, []string{"bounded_base_fee_delta", "fee_market_simulation"}),
		readySequencingTask(EconomicSequencePhase1, SequencingTaskMempoolExecutionFeeValidationAlignment, true, true, []string{"mempool_execution_aligned_fee_validation"}),
		readySequencingTask(EconomicSequencePhase1, SequencingTaskSlashingFundRouting, true, true, []string{"slashing_burn_treasury_reporter_routing"}),
		readySequencingTask(EconomicSequencePhase1, SequencingTaskFundMovementInvariantTests, true, true, []string{"burn_accounting_invariants", "fee_bucket_invariants", "security_invariants"}),
	}
}

func DefaultPhase2SequencingTasks() []EconomicSequencingTask {
	return []EconomicSequencingTask{
		readySequencingTask(EconomicSequencePhase2, SequencingTaskValidatorScoring, true, true, []string{"validator_eligibility_score", "score_component_events"}),
		readySequencingTask(EconomicSequencePhase2, SequencingTaskEpochBasedSelectionProduction, true, true, []string{"epoch_selection_recommendation", "bounded_validator_churn"}),
		readySequencingTask(EconomicSequencePhase2, SequencingTaskConcentrationRewardDampening, true, true, []string{"concentration_soft_cap", "reward_adjustment_bounds"}),
		readySequencingTask(EconomicSequencePhase2, SequencingTaskValidatorRiskScoreQueries, false, true, []string{"validator_reputation_score", "risk_score_explanation"}),
		readySequencingTask(EconomicSequencePhase2, SequencingTaskCommissionChangeWarnings, false, true, []string{"commission_stability_score", "capture_risk_warning"}),
		readySequencingTask(EconomicSequencePhase2, SequencingTaskRiskAdjustedYieldEstimates, false, true, []string{"estimated_net_yield", "risk_adjusted_yield_formula"}),
		readySequencingTask(EconomicSequencePhase2, SequencingTaskValidatorBootstrapBand, true, true, []string{"bootstrap_bonus_bps", "automatic_expiry_conditions"}),
	}
}

func DefaultPhase3SequencingTasks() []EconomicSequencingTask {
	return []EconomicSequencingTask{
		readySequencingTask(EconomicSequencePhase3, SequencingTaskResourceSpecificFeeMultipliers, true, true, []string{"compute_storage_deployment_forwarding_multipliers"}),
		readySequencingTask(EconomicSequencePhase3, SequencingTaskSenderLocalSpamSurcharge, true, true, []string{"sender_failed_tx_surcharge", "mempool_execution_alignment"}),
		readySequencingTask(EconomicSequencePhase3, SequencingTaskStorageWriteUpdatePricing, true, true, []string{"state_write_fee", "state_update_fee"}),
		readySequencingTask(EconomicSequencePhase3, SequencingTaskStorageFootprintQueries, false, true, []string{"storage_footprint_query", "rent_status_query"}),
		readySequencingTask(EconomicSequencePhase3, SequencingTaskDeleteRefundPolicy, true, true, []string{"delete_refund_cap", "refund_decay"}),
		readySequencingTask(EconomicSequencePhase3, SequencingTaskDeploymentForwardingFeeEstimation, true, true, []string{"deployment_fee_estimate", "async_forwarding_fee_estimate"}),
		readySequencingTask(EconomicSequencePhase3, SequencingTaskStateGrowthSurcharge, true, true, []string{"state_growth_surcharge_bps", "state_growth_telemetry"}),
	}
}

func DefaultPhase4SequencingTasks() []EconomicSequencingTask {
	return []EconomicSequencingTask{
		readySequencingTask(EconomicSequencePhase4, SequencingTaskAdaptiveInflationController, true, true, []string{"adaptive_inflation_bounds", "net_issuance_accounting"}),
		readySequencingTask(EconomicSequencePhase4, SequencingTaskSupplyProjectionReports, true, true, []string{"supply_projection_report", "target_net_issuance_range"}),
		readySequencingTask(EconomicSequencePhase4, SequencingTaskEconomicSecurityModule, true, true, []string{"economic_security_epoch_report", "economic_invariant_events"}),
		readySequencingTask(EconomicSequencePhase4, SequencingTaskFeeMarketCircuitBreaker, true, true, []string{"fee_market_circuit_breaker", "controller_instability_thresholds"}),
		readySequencingTask(EconomicSequencePhase4, SequencingTaskSecurityReserveAccounting, true, true, []string{"security_reserve_accounting", "reserve_funding_requests"}),
		readySequencingTask(EconomicSequencePhase4, SequencingTaskPreUpgradeEconomicSimulationRequirement, true, true, []string{"governance_parameter_impact_report", "pre_upgrade_simulation_required"}),
	}
}

func BuildImplementationSequencingReport(phase0Tasks, phase1Tasks []EconomicSequencingTask) EconomicSequencingReport {
	return buildImplementationSequencingReport(phase0Tasks, phase1Tasks, nil, nil, nil, false)
}

func BuildFullImplementationSequencingReport(phase0Tasks, phase1Tasks, phase2Tasks, phase3Tasks []EconomicSequencingTask, phase4Tasks ...[]EconomicSequencingTask) EconomicSequencingReport {
	var phase4 []EconomicSequencingTask
	if len(phase4Tasks) > 0 {
		phase4 = phase4Tasks[0]
	}
	return buildImplementationSequencingReport(phase0Tasks, phase1Tasks, phase2Tasks, phase3Tasks, phase4, true)
}

func buildImplementationSequencingReport(phase0Tasks, phase1Tasks, phase2Tasks, phase3Tasks, phase4Tasks []EconomicSequencingTask, includeLaterPhases bool) EconomicSequencingReport {
	if phase0Tasks == nil {
		phase0Tasks = DefaultPhase0SequencingTasks()
	}
	if phase1Tasks == nil {
		phase1Tasks = DefaultPhase1SequencingTasks()
	}
	if phase2Tasks == nil {
		phase2Tasks = DefaultPhase2SequencingTasks()
	}
	if phase3Tasks == nil {
		phase3Tasks = DefaultPhase3SequencingTasks()
	}
	if phase4Tasks == nil {
		phase4Tasks = DefaultPhase4SequencingTasks()
	}

	phase0 := buildSequencingPhaseReport(
		EconomicSequencePhase0,
		"make existing economics observable before changing incentives",
		phase0Tasks,
		phase0ExpectedTasks(),
	)
	phase1 := buildSequencingPhaseReport(
		EconomicSequencePhase1,
		"close incomplete production paths that affect supply, fees, and penalties",
		phase1Tasks,
		phase1ExpectedTasks(),
	)
	phase2 := buildSequencingPhaseReport(
		EconomicSequencePhase2,
		"improve security and decentralization of staking",
		phase2Tasks,
		phase2ExpectedTasks(),
	)
	phase3 := buildSequencingPhaseReport(
		EconomicSequencePhase3,
		"price resource usage more accurately",
		phase3Tasks,
		phase3ExpectedTasks(),
	)
	phase4 := buildSequencingPhaseReport(
		EconomicSequencePhase4,
		"couple issuance, burns, security budget, and activity into a stable economic loop",
		phase4Tasks,
		phase4ExpectedTasks(),
	)

	readyForPhase1 := phase0.Passed
	readyForPhase2 := phase0.Passed && phase1.Passed
	readyForPhase3 := readyForPhase2 && phase2.Passed
	readyForPhase4 := readyForPhase3 && phase3.Passed
	readyForIncentiveChanges := readyForPhase2
	readyForResourcePricing := false
	readyForLongTermStabilization := false
	phases := []EconomicSequencingPhaseReport{phase0, phase1}
	failed := append([]string{}, phase0.Failed...)
	failed = append(failed, phase1.Failed...)
	if includeLaterPhases {
		readyForIncentiveChanges = readyForPhase3
		readyForResourcePricing = readyForPhase3 && phase3.Passed
		readyForLongTermStabilization = readyForPhase4 && phase4.Passed
		phases = append(phases, phase2, phase3, phase4)
		failed = append(failed, phase2.Failed...)
		failed = append(failed, phase3.Failed...)
		failed = append(failed, phase4.Failed...)
	}
	sort.Strings(failed)

	status := SequencingStatusReady
	if len(failed) > 0 {
		status = SequencingStatusBlocked
	}

	return EconomicSequencingReport{
		Phases:				phases,
		ReadyForPhase1:			readyForPhase1,
		ReadyForPhase2:			readyForPhase2,
		ReadyForPhase3:			readyForPhase3,
		ReadyForPhase4:			readyForPhase4,
		ReadyForIncentiveChanges:	readyForIncentiveChanges,
		ReadyForResourcePricing:	readyForResourcePricing,
		ReadyForLongTermStabilization:	readyForLongTermStabilization,
		Failed:				failed,
		Passed:				len(failed) == 0,
		GovernanceSummary:		sequencingGovernanceSummary(status, phases),
	}
}

func ValidateEconomicSequencingMetrics(input EconomicSequencingMetricInput) EconomicSequencingExitCriterion {
	failed := make([]string, 0)
	if !input.EpochReport.Reconciled {
		failed = append(failed, "epoch_economic_report_not_reconciled")
	}
	if normalizeInt(input.BurnSupply.CumulativeBurnedNaet).IsNegative() || normalizeInt(input.BurnSupply.RecentBurnedNaet).IsNegative() {
		failed = append(failed, "burn_supply_negative")
	}
	if !input.FeeAllocation.SumsExactly {
		failed = append(failed, "fee_buckets_do_not_sum")
	}
	if !input.FeeOptimizer.Validation.MempoolExecutionAligned {
		failed = append(failed, "mempool_execution_fee_validation_mismatch")
	}
	if input.StateGrowth.EpochID == 0 || input.StateGrowth.BlockHeight == 0 {
		failed = append(failed, "state_growth_not_queryable")
	}
	if !input.AdaptiveInflation.Reconciled {
		failed = append(failed, "net_issuance_not_reconciled")
	}
	if !input.BurnFeeInvariant.Passed {
		failed = append(failed, "burn_fee_distribution_invariant_failed")
	}
	if !input.BurnSlashingInvariant.Passed {
		failed = append(failed, "burn_slashing_distribution_invariant_failed")
	}
	if len(input.SecurityReport.Failed) > 0 {
		failed = append(failed, "security_invariant_violations")
	}
	if _, err := ComputeValidatorRewardPerVotingPowerTelemetry(input.ValidatorRewardTelemetry); err != nil {
		failed = append(failed, "validator_reward_telemetry_invalid")
	}

	return EconomicSequencingExitCriterion{
		Name:		"phase_0_phase_1_accounting_reconciliation",
		Met:		len(failed) == 0,
		FailedReason:	strings.Join(failed, ","),
	}
}

func ComputeValidatorRewardPerVotingPowerTelemetry(samples []ValidatorRewardTelemetrySample) ([]ValidatorRewardPerPowerTelemetry, error) {
	out := make([]ValidatorRewardPerPowerTelemetry, 0, len(samples))
	seen := make(map[string]struct{}, len(samples))
	for i, sample := range samples {
		if sample.ValidatorID == "" {
			return nil, fmt.Errorf("validator_reward_telemetry[%d] validator_id is required", i)
		}
		if _, ok := seen[sample.ValidatorID]; ok {
			return nil, fmt.Errorf("validator_reward_telemetry[%d] duplicate validator_id %s", i, sample.ValidatorID)
		}
		seen[sample.ValidatorID] = struct{}{}
		if err := validateBps(fmt.Sprintf("validator_reward_telemetry[%d].voting_power_bps", i), sample.VotingPowerBps, 0, BasisPoints); err != nil {
			return nil, err
		}
		reward := normalizeInt(sample.RewardNaet)
		if reward.IsNegative() {
			return nil, fmt.Errorf("validator_reward_telemetry[%d].reward_naet must not be negative", i)
		}
		rewardPerBps := sdkmath.ZeroInt()
		if sample.VotingPowerBps > 0 {
			rewardPerBps = reward.QuoRaw(sample.VotingPowerBps)
		}
		out = append(out, ValidatorRewardPerPowerTelemetry{
			ValidatorID:		sample.ValidatorID,
			VotingPowerBps:		sample.VotingPowerBps,
			RewardNaet:		reward,
			RewardPerBpsNaet:	rewardPerBps,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ValidatorID < out[j].ValidatorID
	})
	return out, nil
}

func readySequencingTask(phaseID, name string, reconciled bool, queryable bool, evidence []string) EconomicSequencingTask {
	return EconomicSequencingTask{
		PhaseID:		phaseID,
		Name:			name,
		Implemented:		true,
		Observable:		true,
		Queryable:		queryable,
		InvariantTested:	true,
		Reconciled:		reconciled,
		Evidence:		append([]string{}, evidence...),
	}
}

func buildSequencingPhaseReport(phaseID, goal string, tasks []EconomicSequencingTask, expected []string) EconomicSequencingPhaseReport {
	failed := validateSequencingTaskSet(phaseID, tasks, expected)
	taskByName := make(map[string]EconomicSequencingTask, len(tasks))
	for _, task := range tasks {
		taskByName[task.Name] = task
		if !sequencingTaskReady(task) {
			failed = append(failed, phaseID+":"+task.Name+"_not_ready")
		}
	}

	criteria := []EconomicSequencingExitCriterion{
		sequencingCriterion(phaseID+":all_tasks_implemented_observable_tested", allSequencingTasksReady(tasks), "one_or_more_tasks_not_ready"),
		sequencingCriterion(phaseID+":required_tasks_present", sequencingExpectedTasksPresent(taskByName, expected), "required_task_missing"),
	}
	switch phaseID {
	case EconomicSequencePhase0:
		criteria = append(criteria,
			sequencingCriterion(phaseID+":accounting_reconciles_per_epoch", sequencingTasksReconciled(taskByName, []string{SequencingTaskNetIssuanceAccounting, SequencingTaskCumulativeBurnAccounting, SequencingTaskFeeBucketAccounting, SequencingTaskEpochEconomicReportGeneration}), "accounting_task_not_reconciled"),
			sequencingCriterion(phaseID+":concentration_and_state_growth_queryable", sequencingTasksQueryable(taskByName, []string{SequencingTaskValidatorConcentrationMetrics, SequencingTaskStateGrowthTelemetry, SequencingTaskValidatorRewardPerVotingPower}), "measurement_task_not_queryable"),
		)
	case EconomicSequencePhase1:
		criteria = append(criteria,
			sequencingCriterion(phaseID+":fund_paths_connected", sequencingTasksReconciled(taskByName, []string{SequencingTaskBurnControllerFeeDistributionWiring, SequencingTaskSlashingFundRouting, SequencingTaskFundMovementInvariantTests}), "fund_path_not_reconciled"),
			sequencingCriterion(phaseID+":controllers_bounded_and_aligned", sequencingTasksReconciled(taskByName, []string{SequencingTaskDeflationGuardEnforcement, SequencingTaskBurnCaps, SequencingTaskFeeControllerBounds, SequencingTaskMempoolExecutionFeeValidationAlignment}), "controller_or_fee_validation_not_reconciled"),
		)
	case EconomicSequencePhase2:
		criteria = append(criteria,
			sequencingCriterion(phaseID+":delegator_risk_yield_data_queryable", sequencingTasksQueryable(taskByName, []string{SequencingTaskValidatorRiskScoreQueries, SequencingTaskCommissionChangeWarnings, SequencingTaskRiskAdjustedYieldEstimates}), "delegator_query_surface_not_ready"),
			sequencingCriterion(phaseID+":concentration_incentives_active_bounded", sequencingTasksReconciled(taskByName, []string{SequencingTaskValidatorScoring, SequencingTaskConcentrationRewardDampening, SequencingTaskValidatorBootstrapBand}), "validator_incentive_not_bounded"),
			sequencingCriterion(phaseID+":active_set_transitions_deterministic_tested", sequencingTasksReconciled(taskByName, []string{SequencingTaskEpochBasedSelectionProduction}), "epoch_selection_not_deterministic"),
		)
	case EconomicSequencePhase3:
		criteria = append(criteria,
			sequencingCriterion(phaseID+":persistent_state_and_spam_directly_priced", sequencingTasksReconciled(taskByName, []string{SequencingTaskSenderLocalSpamSurcharge, SequencingTaskStorageWriteUpdatePricing, SequencingTaskDeleteRefundPolicy, SequencingTaskStateGrowthSurcharge}), "resource_abuse_not_directly_priced"),
			sequencingCriterion(phaseID+":fee_estimator_covers_transaction_deployment_async", sequencingTasksReconciled(taskByName, []string{SequencingTaskResourceSpecificFeeMultipliers, SequencingTaskDeploymentForwardingFeeEstimation}), "fee_estimator_coverage_missing"),
			sequencingCriterion(phaseID+":state_growth_bounded_by_pricing_telemetry", sequencingTasksQueryable(taskByName, []string{SequencingTaskStorageFootprintQueries}) && sequencingTasksReconciled(taskByName, []string{SequencingTaskStateGrowthSurcharge}), "state_growth_bound_not_queryable"),
		)
	case EconomicSequencePhase4:
		criteria = append(criteria,
			sequencingCriterion(phaseID+":issuance_burn_net_issuance_explicit_policy", sequencingTasksReconciled(taskByName, []string{SequencingTaskAdaptiveInflationController, SequencingTaskSupplyProjectionReports, SequencingTaskEconomicSecurityModule, SequencingTaskSecurityReserveAccounting}), "supply_policy_loop_not_explicit"),
			sequencingCriterion(phaseID+":controllers_simulation_tested_before_activation", sequencingTasksReconciled(taskByName, []string{SequencingTaskAdaptiveInflationController, SequencingTaskFeeMarketCircuitBreaker, SequencingTaskPreUpgradeEconomicSimulationRequirement}), "controller_simulation_gate_missing"),
			sequencingCriterion(phaseID+":governance_parameter_impact_reports_available", sequencingTasksQueryable(taskByName, []string{SequencingTaskSupplyProjectionReports, SequencingTaskPreUpgradeEconomicSimulationRequirement}), "governance_impact_report_not_queryable"),
		)
	}
	for _, criterion := range criteria {
		if !criterion.Met {
			failed = append(failed, criterion.Name+":"+criterion.FailedReason)
		}
	}
	sort.Strings(failed)
	status := SequencingStatusReady
	if len(failed) > 0 {
		status = SequencingStatusBlocked
	}
	return EconomicSequencingPhaseReport{
		PhaseID:	phaseID,
		Goal:		goal,
		Status:		status,
		Tasks:		append([]EconomicSequencingTask{}, tasks...),
		ExitCriteria:	criteria,
		Failed:		failed,
		Passed:		len(failed) == 0,
	}
}

func validateSequencingTaskSet(phaseID string, tasks []EconomicSequencingTask, expected []string) []string {
	failed := make([]string, 0)
	seen := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		if task.PhaseID != phaseID {
			failed = append(failed, phaseID+":"+task.Name+"_wrong_phase")
		}
		if task.Name == "" {
			failed = append(failed, phaseID+":task_name_required")
			continue
		}
		if _, ok := seen[task.Name]; ok {
			failed = append(failed, phaseID+":"+task.Name+"_duplicate")
		}
		seen[task.Name] = struct{}{}
	}
	for _, name := range expected {
		if _, ok := seen[name]; !ok {
			failed = append(failed, phaseID+":"+name+"_missing")
		}
	}
	return failed
}

func sequencingTaskReady(task EconomicSequencingTask) bool {
	if !task.Implemented || !task.Observable || !task.InvariantTested {
		return false
	}
	switch task.Name {
	case SequencingTaskValidatorConcentrationMetrics, SequencingTaskStateGrowthTelemetry, SequencingTaskValidatorRewardPerVotingPower:
		return task.Queryable
	case SequencingTaskValidatorRiskScoreQueries, SequencingTaskCommissionChangeWarnings, SequencingTaskRiskAdjustedYieldEstimates, SequencingTaskStorageFootprintQueries:
		return task.Queryable
	case SequencingTaskNetIssuanceAccounting, SequencingTaskCumulativeBurnAccounting, SequencingTaskFeeBucketAccounting, SequencingTaskEpochEconomicReportGeneration,
		SequencingTaskBurnControllerFeeDistributionWiring, SequencingTaskDeflationGuardEnforcement, SequencingTaskBurnCaps, SequencingTaskFeeControllerBounds,
		SequencingTaskMempoolExecutionFeeValidationAlignment, SequencingTaskSlashingFundRouting, SequencingTaskFundMovementInvariantTests,
		SequencingTaskValidatorScoring, SequencingTaskEpochBasedSelectionProduction, SequencingTaskConcentrationRewardDampening, SequencingTaskValidatorBootstrapBand,
		SequencingTaskResourceSpecificFeeMultipliers, SequencingTaskSenderLocalSpamSurcharge, SequencingTaskStorageWriteUpdatePricing, SequencingTaskDeleteRefundPolicy,
		SequencingTaskDeploymentForwardingFeeEstimation, SequencingTaskStateGrowthSurcharge,
		SequencingTaskAdaptiveInflationController, SequencingTaskSupplyProjectionReports, SequencingTaskEconomicSecurityModule, SequencingTaskFeeMarketCircuitBreaker,
		SequencingTaskSecurityReserveAccounting, SequencingTaskPreUpgradeEconomicSimulationRequirement:
		return task.Reconciled
	default:
		return false
	}
}

func allSequencingTasksReady(tasks []EconomicSequencingTask) bool {
	if len(tasks) == 0 {
		return false
	}
	for _, task := range tasks {
		if !sequencingTaskReady(task) {
			return false
		}
	}
	return true
}

func sequencingExpectedTasksPresent(taskByName map[string]EconomicSequencingTask, expected []string) bool {
	for _, name := range expected {
		if _, ok := taskByName[name]; !ok {
			return false
		}
	}
	return true
}

func sequencingTasksReconciled(taskByName map[string]EconomicSequencingTask, names []string) bool {
	for _, name := range names {
		task, ok := taskByName[name]
		if !ok || !task.Reconciled {
			return false
		}
	}
	return true
}

func sequencingTasksQueryable(taskByName map[string]EconomicSequencingTask, names []string) bool {
	for _, name := range names {
		task, ok := taskByName[name]
		if !ok || !task.Queryable {
			return false
		}
	}
	return true
}

func sequencingCriterion(name string, met bool, reason string) EconomicSequencingExitCriterion {
	if met {
		return EconomicSequencingExitCriterion{Name: name, Met: true}
	}
	return EconomicSequencingExitCriterion{Name: name, Met: false, FailedReason: reason}
}

func phase0ExpectedTasks() []string {
	return []string{
		SequencingTaskNetIssuanceAccounting,
		SequencingTaskCumulativeBurnAccounting,
		SequencingTaskFeeBucketAccounting,
		SequencingTaskValidatorConcentrationMetrics,
		SequencingTaskStateGrowthTelemetry,
		SequencingTaskValidatorRewardPerVotingPower,
		SequencingTaskEpochEconomicReportGeneration,
	}
}

func phase1ExpectedTasks() []string {
	return []string{
		SequencingTaskBurnControllerFeeDistributionWiring,
		SequencingTaskDeflationGuardEnforcement,
		SequencingTaskBurnCaps,
		SequencingTaskFeeControllerBounds,
		SequencingTaskMempoolExecutionFeeValidationAlignment,
		SequencingTaskSlashingFundRouting,
		SequencingTaskFundMovementInvariantTests,
	}
}

func phase2ExpectedTasks() []string {
	return []string{
		SequencingTaskValidatorScoring,
		SequencingTaskEpochBasedSelectionProduction,
		SequencingTaskConcentrationRewardDampening,
		SequencingTaskValidatorRiskScoreQueries,
		SequencingTaskCommissionChangeWarnings,
		SequencingTaskRiskAdjustedYieldEstimates,
		SequencingTaskValidatorBootstrapBand,
	}
}

func phase3ExpectedTasks() []string {
	return []string{
		SequencingTaskResourceSpecificFeeMultipliers,
		SequencingTaskSenderLocalSpamSurcharge,
		SequencingTaskStorageWriteUpdatePricing,
		SequencingTaskStorageFootprintQueries,
		SequencingTaskDeleteRefundPolicy,
		SequencingTaskDeploymentForwardingFeeEstimation,
		SequencingTaskStateGrowthSurcharge,
	}
}

func phase4ExpectedTasks() []string {
	return []string{
		SequencingTaskAdaptiveInflationController,
		SequencingTaskSupplyProjectionReports,
		SequencingTaskEconomicSecurityModule,
		SequencingTaskFeeMarketCircuitBreaker,
		SequencingTaskSecurityReserveAccounting,
		SequencingTaskPreUpgradeEconomicSimulationRequirement,
	}
}

func sequencingGovernanceSummary(status string, phases []EconomicSequencingPhaseReport) string {
	parts := make([]string, 0, len(phases)+1)
	parts = append(parts, status+":"+strings.TrimPrefix(EconomicSequencePhase0, "phase_"))
	for i, phase := range phases {
		parts = append(parts, fmt.Sprintf("phase%d=%s", i, phase.Status))
	}
	return strings.Join(parts, " ")
}
