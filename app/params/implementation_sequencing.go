package params

import (
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	EconomicSequencePhase0 = "phase_0_measurement_accounting"
	EconomicSequencePhase1 = "phase_1_production_safety"

	SequencingStatusReady   = "ready"
	SequencingStatusBlocked = "blocked"

	SequencingTaskNetIssuanceAccounting                  = "net_issuance_accounting"
	SequencingTaskCumulativeBurnAccounting               = "cumulative_burn_accounting"
	SequencingTaskFeeBucketAccounting                    = "fee_bucket_accounting"
	SequencingTaskValidatorConcentrationMetrics          = "validator_concentration_metrics"
	SequencingTaskStateGrowthTelemetry                   = "state_growth_telemetry"
	SequencingTaskValidatorRewardPerVotingPower          = "validator_reward_per_voting_power_telemetry"
	SequencingTaskEpochEconomicReportGeneration          = "epoch_economic_report_generation"
	SequencingTaskBurnControllerFeeDistributionWiring    = "burn_controller_fee_distribution_wiring"
	SequencingTaskDeflationGuardEnforcement              = "deflation_guard_enforcement"
	SequencingTaskBurnCaps                               = "burn_caps"
	SequencingTaskFeeControllerBounds                    = "fee_controller_bounds"
	SequencingTaskMempoolExecutionFeeValidationAlignment = "mempool_execution_fee_validation_alignment"
	SequencingTaskSlashingFundRouting                    = "slashing_fund_routing"
	SequencingTaskFundMovementInvariantTests             = "fund_movement_invariant_tests"
)

type EconomicSequencingTask struct {
	PhaseID         string
	Name            string
	Implemented     bool
	Observable      bool
	Queryable       bool
	InvariantTested bool
	Reconciled      bool
	Evidence        []string
}

type EconomicSequencingExitCriterion struct {
	Name         string
	Met          bool
	FailedReason string
}

type EconomicSequencingPhaseReport struct {
	PhaseID      string
	Goal         string
	Status       string
	Tasks        []EconomicSequencingTask
	ExitCriteria []EconomicSequencingExitCriterion
	Failed       []string
	Passed       bool
}

type EconomicSequencingReport struct {
	Phases                   []EconomicSequencingPhaseReport
	ReadyForPhase1           bool
	ReadyForIncentiveChanges bool
	Failed                   []string
	Passed                   bool
	GovernanceSummary        string
}

type EconomicSequencingMetricInput struct {
	EpochReport              EpochEconomicReport
	BurnSupply               BurnSupplyQueryOutput
	FeeAllocation            FeeAllocationBuckets
	SecurityReport           EconomicSecurityEpochReport
	StateGrowth              StateGrowthTelemetryOutput
	FeeOptimizer             FeeMarketOptimizerOutput
	BurnFeeDistribution      BurnIntegratedFeeDistributionOutput
	BurnSlashingDistribution BurnIntegratedSlashingDistributionOutput
	BurnFeeInvariant         BurnAccountingInvariantReport
	BurnSlashingInvariant    BurnAccountingInvariantReport
	AdaptiveInflation        AdaptiveInflationEpochReport
	ValidatorRewardTelemetry []ValidatorRewardTelemetrySample
}

type ValidatorRewardTelemetrySample struct {
	ValidatorID    string
	VotingPowerBps int64
	RewardNaet     sdkmath.Int
}

type ValidatorRewardPerPowerTelemetry struct {
	ValidatorID      string
	VotingPowerBps   int64
	RewardNaet       sdkmath.Int
	RewardPerBpsNaet sdkmath.Int
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

func BuildImplementationSequencingReport(phase0Tasks, phase1Tasks []EconomicSequencingTask) EconomicSequencingReport {
	if phase0Tasks == nil {
		phase0Tasks = DefaultPhase0SequencingTasks()
	}
	if phase1Tasks == nil {
		phase1Tasks = DefaultPhase1SequencingTasks()
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

	readyForPhase1 := phase0.Passed
	readyForIncentiveChanges := phase0.Passed && phase1.Passed
	failed := append([]string{}, phase0.Failed...)
	failed = append(failed, phase1.Failed...)
	sort.Strings(failed)

	status := SequencingStatusReady
	if !readyForIncentiveChanges {
		status = SequencingStatusBlocked
	}

	return EconomicSequencingReport{
		Phases:                   []EconomicSequencingPhaseReport{phase0, phase1},
		ReadyForPhase1:           readyForPhase1,
		ReadyForIncentiveChanges: readyForIncentiveChanges,
		Failed:                   failed,
		Passed:                   len(failed) == 0,
		GovernanceSummary:        fmt.Sprintf("%s:%s phase0=%s phase1=%s", status, strings.TrimPrefix(EconomicSequencePhase0, "phase_"), phase0.Status, phase1.Status),
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
		Name:         "phase_0_phase_1_accounting_reconciliation",
		Met:          len(failed) == 0,
		FailedReason: strings.Join(failed, ","),
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
			ValidatorID:      sample.ValidatorID,
			VotingPowerBps:   sample.VotingPowerBps,
			RewardNaet:       reward,
			RewardPerBpsNaet: rewardPerBps,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ValidatorID < out[j].ValidatorID
	})
	return out, nil
}

func readySequencingTask(phaseID, name string, reconciled bool, queryable bool, evidence []string) EconomicSequencingTask {
	return EconomicSequencingTask{
		PhaseID:         phaseID,
		Name:            name,
		Implemented:     true,
		Observable:      true,
		Queryable:       queryable,
		InvariantTested: true,
		Reconciled:      reconciled,
		Evidence:        append([]string{}, evidence...),
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
	if phaseID == EconomicSequencePhase0 {
		criteria = append(criteria,
			sequencingCriterion(phaseID+":accounting_reconciles_per_epoch", sequencingTasksReconciled(taskByName, []string{SequencingTaskNetIssuanceAccounting, SequencingTaskCumulativeBurnAccounting, SequencingTaskFeeBucketAccounting, SequencingTaskEpochEconomicReportGeneration}), "accounting_task_not_reconciled"),
			sequencingCriterion(phaseID+":concentration_and_state_growth_queryable", sequencingTasksQueryable(taskByName, []string{SequencingTaskValidatorConcentrationMetrics, SequencingTaskStateGrowthTelemetry, SequencingTaskValidatorRewardPerVotingPower}), "measurement_task_not_queryable"),
		)
	} else {
		criteria = append(criteria,
			sequencingCriterion(phaseID+":fund_paths_connected", sequencingTasksReconciled(taskByName, []string{SequencingTaskBurnControllerFeeDistributionWiring, SequencingTaskSlashingFundRouting, SequencingTaskFundMovementInvariantTests}), "fund_path_not_reconciled"),
			sequencingCriterion(phaseID+":controllers_bounded_and_aligned", sequencingTasksReconciled(taskByName, []string{SequencingTaskDeflationGuardEnforcement, SequencingTaskBurnCaps, SequencingTaskFeeControllerBounds, SequencingTaskMempoolExecutionFeeValidationAlignment}), "controller_or_fee_validation_not_reconciled"),
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
		PhaseID:      phaseID,
		Goal:         goal,
		Status:       status,
		Tasks:        append([]EconomicSequencingTask{}, tasks...),
		ExitCriteria: criteria,
		Failed:       failed,
		Passed:       len(failed) == 0,
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
	case SequencingTaskNetIssuanceAccounting, SequencingTaskCumulativeBurnAccounting, SequencingTaskFeeBucketAccounting, SequencingTaskEpochEconomicReportGeneration,
		SequencingTaskBurnControllerFeeDistributionWiring, SequencingTaskDeflationGuardEnforcement, SequencingTaskBurnCaps, SequencingTaskFeeControllerBounds,
		SequencingTaskMempoolExecutionFeeValidationAlignment, SequencingTaskSlashingFundRouting, SequencingTaskFundMovementInvariantTests:
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
