package params

import (
	"fmt"
	"sort"
	"strings"
)

const (
	EconomicBacklogPriorityHigh	= "high"
	EconomicBacklogPriorityMedium	= "medium"
	EconomicBacklogPriorityLower	= "lower"

	EconomicBacklogAddEconomicsLocalExclude		= "add_economics_md_to_local_git_exclude"
	EconomicBacklogEpochEconomicReportDataModel	= "implement_epoch_economic_report_data_model"
	EconomicBacklogNetIssuanceAccounting		= "add_net_issuance_accounting"
	EconomicBacklogBurnAccountingQueries		= "add_burn_accounting_and_queries"
	EconomicBacklogBurnFeeDistribution		= "wire_burn_allocation_into_fee_distribution"
	EconomicBacklogDeflationGuard			= "enforce_deflation_guard"
	EconomicBacklogFeeAllocationInvariantTests	= "add_fee_allocation_invariant_tests"
	EconomicBacklogSlashingRouteInvariantTests	= "add_slashing_route_invariant_tests"
	EconomicBacklogValidatorConcentrationQueries	= "add_validator_concentration_queries"
	EconomicBacklogStateGrowthTelemetry		= "add_state_growth_telemetry"
	EconomicBacklogValidatorRiskScoreQuery		= "implement_validator_risk_score_query"
	EconomicBacklogCommissionChangeWarningEvent	= "implement_commission_change_warning_event"
	EconomicBacklogDynamicBaseFeeSimulationTests	= "add_dynamic_base_fee_simulation_tests"
	EconomicBacklogSenderLocalSpamSurchargeDesign	= "add_sender_local_spam_surcharge_design"
	EconomicBacklogStoragePricingSpecification	= "add_storage_pricing_specification"
	EconomicBacklogSupplyProjectionCommandOrQuery	= "add_supply_projection_command_or_query"
	EconomicBacklogGovernanceImpactReport		= "add_governance_parameter_impact_report"
	EconomicBacklogFullStateRentLifecycle		= "implement_full_state_rent_lifecycle"
	EconomicBacklogAdaptiveInflationController	= "implement_adaptive_inflation_controller"
	EconomicBacklogSecurityReserveModule		= "implement_security_reserve_module"
	EconomicBacklogFeeMarketCircuitBreaker		= "implement_fee_market_circuit_breaker"
	EconomicBacklogDelegationSimulator		= "implement_delegation_simulator"
	EconomicBacklogValidatorBootstrapBand		= "implement_validator_bootstrap_band"

	EconomicNonGoalSecondStakingAsset			= "do_not_introduce_second_staking_asset"
	EconomicNonGoalExternalValidatorRewardAssets		= "do_not_use_external_assets_for_validator_rewards"
	EconomicNonGoalOffChainFeeAccounting			= "do_not_use_off_chain_data_for_fee_accounting"
	EconomicNonGoalNondeterministicConsensusReputation	= "do_not_use_nondeterministic_reputation_in_consensus"
	EconomicNonGoalUnverifiableValidatorIdentity		= "do_not_rely_on_unverifiable_validator_identity"
	EconomicNonGoalDiscretionarySlashing			= "do_not_reduce_slashing_determinism"
	EconomicNonGoalBurnOverSecurityBudget			= "do_not_prioritize_burn_over_security_budget"
	EconomicNonGoalUnboundedControllers			= "do_not_allow_unbounded_untested_controllers"
)

type EconomicEngineeringBacklogItem struct {
	ID			string
	Priority		string
	Description		string
	Evidence		[]string
	RequiresTests		bool
	RequiresTelemetry	bool
	RequiresQuery		bool
	LocalOnly		bool
	Tracked			bool
}

type EconomicEngineeringBacklogReport struct {
	Items			[]EconomicEngineeringBacklogItem
	RequiredHigh		int
	RequiredMedium		int
	RequiredLower		int
	CoveredHigh		int
	CoveredMedium		int
	CoveredLower		int
	HighCoverageBps		int64
	MediumCoverageBps	int64
	LowerCoverageBps	int64
	Passed			bool
	Failed			[]string
	Summary			string
}

type EconomicNonGoal struct {
	ID			string
	Statement		string
	Enforcement		[]string
	ConsensusCritical	bool
	DeterminismGuard	bool
	SecurityBudgetGuard	bool
	Tracked			bool
}

type EconomicNonGoalReport struct {
	NonGoals	[]EconomicNonGoal
	Required	int
	Covered		int
	CoverageBps	int64
	Passed		bool
	Failed		[]string
	Summary		string
}

func DefaultEconomicEngineeringBacklog() []EconomicEngineeringBacklogItem {
	return []EconomicEngineeringBacklogItem{
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogAddEconomicsLocalExclude, "add /ECONOMICS.md to local git exclude", []string{".git/info/exclude:/ECONOMICS.md"}, false, false, false, true),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogEpochEconomicReportDataModel, "implement epoch economic report data model", []string{"EpochEconomicReport", "BuildEpochEconomicReport"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogNetIssuanceAccounting, "add net issuance accounting", []string{"NetIssuanceReport", "AdaptiveInflationEpochReport"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogBurnAccountingQueries, "add burn accounting and queries", []string{"BurnSupplyQueryOutput", "BuildBurnSupplyQuery"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogBurnFeeDistribution, "wire burn allocation into fee distribution", []string{"BurnIntegratedFeeDistributionOutput", "RouteFeeDistributionWithBurn"}, true, true, false, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogDeflationGuard, "enforce deflation guard", []string{"DeflationGuardStatus", "ApplyDeflationGuard"}, true, true, false, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogFeeAllocationInvariantTests, "add fee allocation invariant tests", []string{"fee_allocation_buckets_sum_exactly"}, true, false, false, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogSlashingRouteInvariantTests, "add slashing route invariant tests", []string{"slashing_route_sums_to_penalty"}, true, false, false, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogValidatorConcentrationQueries, "add validator concentration queries", []string{"ValidatorConcentrationReport", "ValidatorReputationReport"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityHigh, EconomicBacklogStateGrowthTelemetry, "add state growth telemetry", []string{"StateGrowthTelemetryOutput", "StateGrowthTelemetryEvent"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogValidatorRiskScoreQuery, "implement validator risk score query", []string{"ValidatorReputationReport.RiskScoreBps"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogCommissionChangeWarningEvent, "implement commission-change warning event", []string{"commission_stability_score", "capture_risk_warning"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogDynamicBaseFeeSimulationTests, "add dynamic base fee simulation tests", []string{"FeeMarketSimulationReport", "low_steady_burst_spam_load"}, true, false, false, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogSenderLocalSpamSurchargeDesign, "add sender-local spam surcharge design", []string{"AntiSpamAdmissionDecision", "sender_local_surcharge"}, true, true, false, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogStoragePricingSpecification, "add storage pricing specification for first write update delete and refund", []string{"StorageFeeQuote", "DeleteRefundPolicy"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogSupplyProjectionCommandOrQuery, "add supply projection command or query", []string{"SupplyProjectionReport", "ProjectSupply"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityMedium, EconomicBacklogGovernanceImpactReport, "add governance parameter impact report", []string{"GovernanceParameterImpactReport"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityLower, EconomicBacklogFullStateRentLifecycle, "implement full state rent lifecycle", []string{"StateRentStatus", "RentGracePeriod"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityLower, EconomicBacklogAdaptiveInflationController, "implement adaptive inflation controller", []string{"AdaptiveInflationEpochReport"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityLower, EconomicBacklogSecurityReserveModule, "implement security reserve module", []string{"EconomicSecurityEpochReport", "SecurityReserveAllocation"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityLower, EconomicBacklogFeeMarketCircuitBreaker, "implement fee market circuit breaker", []string{"EconomicCircuitBreakerParams", "CircuitBreakerStatus"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityLower, EconomicBacklogDelegationSimulator, "implement delegation simulator", []string{"risk_adjusted_yield_estimates", "redelegation_preview"}, true, true, true, false),
		backlogItem(EconomicBacklogPriorityLower, EconomicBacklogValidatorBootstrapBand, "implement validator bootstrap band", []string{"bootstrap_bonus_bps", "automatic_expiry_conditions"}, true, true, true, false),
	}
}

func DefaultEconomicNonGoals() []EconomicNonGoal {
	return []EconomicNonGoal{
		nonGoal(EconomicNonGoalSecondStakingAsset, "do not introduce a second staking asset", []string{"native_denom_naet", "staking_asset_invariant"}, true, true, false),
		nonGoal(EconomicNonGoalExternalValidatorRewardAssets, "do not make external assets part of validator rewards", []string{"reward_denom_naet_only", "distribution_denom_invariant"}, true, true, false),
		nonGoal(EconomicNonGoalOffChainFeeAccounting, "do not make fee accounting depend on off-chain data", []string{"fee_accounting_consensus_inputs_only", "mempool_execution_fee_alignment"}, true, true, false),
		nonGoal(EconomicNonGoalNondeterministicConsensusReputation, "do not use non-deterministic reputation inputs in consensus-critical calculations", []string{"consensus_safe_score_components", "advisory_reputation_guard"}, true, true, false),
		nonGoal(EconomicNonGoalUnverifiableValidatorIdentity, "do not rely on unverifiable validator identity assumptions", []string{"economic_constraints_not_identity_rules", "sybil_simulation"}, true, true, false),
		nonGoal(EconomicNonGoalDiscretionarySlashing, "do not reduce slashing determinism for discretionary penalty handling", []string{"deterministic_slashing_severity", "slashing_route_invariant"}, true, true, false),
		nonGoal(EconomicNonGoalBurnOverSecurityBudget, "do not make burn priority higher than validator security budget", []string{"security_reward_floor_priority", "deflation_guard"}, true, true, true),
		nonGoal(EconomicNonGoalUnboundedControllers, "do not allow economic controllers to operate without bounds telemetry and tests", []string{"controller_bounds", "telemetry_events", "invariant_tests"}, true, true, true),
	}
}

func BuildEconomicEngineeringBacklogReport(items []EconomicEngineeringBacklogItem) EconomicEngineeringBacklogReport {
	if items == nil {
		items = DefaultEconomicEngineeringBacklog()
	}
	out, failed, required, covered := evaluateEconomicEngineeringBacklog(items)
	sort.Strings(failed)
	highCoverage := coverageBps(covered[EconomicBacklogPriorityHigh], required[EconomicBacklogPriorityHigh])
	mediumCoverage := coverageBps(covered[EconomicBacklogPriorityMedium], required[EconomicBacklogPriorityMedium])
	lowerCoverage := coverageBps(covered[EconomicBacklogPriorityLower], required[EconomicBacklogPriorityLower])
	return EconomicEngineeringBacklogReport{
		Items:			out,
		RequiredHigh:		required[EconomicBacklogPriorityHigh],
		RequiredMedium:		required[EconomicBacklogPriorityMedium],
		RequiredLower:		required[EconomicBacklogPriorityLower],
		CoveredHigh:		covered[EconomicBacklogPriorityHigh],
		CoveredMedium:		covered[EconomicBacklogPriorityMedium],
		CoveredLower:		covered[EconomicBacklogPriorityLower],
		HighCoverageBps:	highCoverage,
		MediumCoverageBps:	mediumCoverage,
		LowerCoverageBps:	lowerCoverage,
		Passed:			len(failed) == 0 && highCoverage == BasisPoints && mediumCoverage == BasisPoints && lowerCoverage == BasisPoints,
		Failed:			failed,
		Summary:		fmt.Sprintf("backlog_high=%d/%d backlog_medium=%d/%d backlog_lower=%d/%d coverage_bps=%d/%d/%d", covered[EconomicBacklogPriorityHigh], required[EconomicBacklogPriorityHigh], covered[EconomicBacklogPriorityMedium], required[EconomicBacklogPriorityMedium], covered[EconomicBacklogPriorityLower], required[EconomicBacklogPriorityLower], highCoverage, mediumCoverage, lowerCoverage),
	}
}

func BuildEconomicNonGoalReport(nonGoals []EconomicNonGoal) EconomicNonGoalReport {
	if nonGoals == nil {
		nonGoals = DefaultEconomicNonGoals()
	}
	out, failed, required, covered := evaluateEconomicNonGoals(nonGoals)
	sort.Strings(failed)
	coverage := coverageBps(covered, required)
	return EconomicNonGoalReport{
		NonGoals:	out,
		Required:	required,
		Covered:	covered,
		CoverageBps:	coverage,
		Passed:		len(failed) == 0 && coverage == BasisPoints,
		Failed:		failed,
		Summary:	fmt.Sprintf("economic_non_goals=%d/%d coverage_bps=%d", covered, required, coverage),
	}
}

func backlogItem(priority, id, description string, evidence []string, requiresTests, requiresTelemetry, requiresQuery, localOnly bool) EconomicEngineeringBacklogItem {
	return EconomicEngineeringBacklogItem{
		ID:			id,
		Priority:		priority,
		Description:		description,
		Evidence:		append([]string{}, evidence...),
		RequiresTests:		requiresTests,
		RequiresTelemetry:	requiresTelemetry,
		RequiresQuery:		requiresQuery,
		LocalOnly:		localOnly,
		Tracked:		true,
	}
}

func nonGoal(id, statement string, enforcement []string, consensusCritical, determinismGuard, securityBudgetGuard bool) EconomicNonGoal {
	return EconomicNonGoal{
		ID:			id,
		Statement:		statement,
		Enforcement:		append([]string{}, enforcement...),
		ConsensusCritical:	consensusCritical,
		DeterminismGuard:	determinismGuard,
		SecurityBudgetGuard:	securityBudgetGuard,
		Tracked:		true,
	}
}

func evaluateEconomicEngineeringBacklog(items []EconomicEngineeringBacklogItem) ([]EconomicEngineeringBacklogItem, []string, map[string]int, map[string]int) {
	out := append([]EconomicEngineeringBacklogItem{}, items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].ID < out[j].ID
		}
		return backlogPriorityRank(out[i].Priority) < backlogPriorityRank(out[j].Priority)
	})
	expected := requiredEconomicEngineeringBacklogIDs()
	required := map[string]int{}
	covered := map[string]int{}
	failed := make([]string, 0)
	seen := make(map[string]EconomicEngineeringBacklogItem, len(out))
	for priority, ids := range expected {
		required[priority] = len(ids)
	}
	for _, item := range out {
		if item.ID == "" {
			failed = append(failed, "economic_backlog_item_id_required")
			continue
		}
		if _, ok := expected[item.Priority]; !ok {
			failed = append(failed, item.ID+":unknown_backlog_priority")
		}
		if _, ok := seen[item.ID]; ok {
			failed = append(failed, item.ID+":duplicate_backlog_item")
		}
		seen[item.ID] = item
		failed = append(failed, validateEconomicEngineeringBacklogItem(item)...)
	}
	for priority, ids := range expected {
		for _, id := range ids {
			item, ok := seen[id]
			if !ok {
				failed = append(failed, id+":missing_required_backlog_item")
				continue
			}
			if item.Priority != priority {
				failed = append(failed, id+":wrong_backlog_priority")
				continue
			}
			if economicEngineeringBacklogItemCovered(item) {
				covered[priority]++
			}
		}
	}
	return out, failed, required, covered
}

func evaluateEconomicNonGoals(nonGoals []EconomicNonGoal) ([]EconomicNonGoal, []string, int, int) {
	out := append([]EconomicNonGoal{}, nonGoals...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	expected := requiredEconomicNonGoalIDs()
	failed := make([]string, 0)
	seen := make(map[string]EconomicNonGoal, len(out))
	for _, nonGoal := range out {
		if nonGoal.ID == "" {
			failed = append(failed, "economic_non_goal_id_required")
			continue
		}
		if _, ok := expected[nonGoal.ID]; !ok {
			failed = append(failed, nonGoal.ID+":unknown_non_goal")
		}
		if _, ok := seen[nonGoal.ID]; ok {
			failed = append(failed, nonGoal.ID+":duplicate_non_goal")
		}
		seen[nonGoal.ID] = nonGoal
		failed = append(failed, validateEconomicNonGoal(nonGoal)...)
	}
	covered := 0
	for id := range expected {
		nonGoal, ok := seen[id]
		if !ok {
			failed = append(failed, id+":missing_required_non_goal")
			continue
		}
		if economicNonGoalCovered(nonGoal) {
			covered++
		}
	}
	return out, failed, len(expected), covered
}

func validateEconomicEngineeringBacklogItem(item EconomicEngineeringBacklogItem) []string {
	failed := make([]string, 0)
	if strings.TrimSpace(item.Description) == "" {
		failed = append(failed, item.ID+":description_missing")
	}
	if !item.Tracked {
		failed = append(failed, item.ID+":not_tracked")
	}
	if len(item.Evidence) == 0 {
		failed = append(failed, item.ID+":evidence_missing")
	}
	for i, evidence := range item.Evidence {
		if strings.TrimSpace(evidence) == "" {
			failed = append(failed, fmt.Sprintf("%s:evidence_%d_blank", item.ID, i))
		}
	}
	if item.Priority == EconomicBacklogPriorityHigh && !item.LocalOnly {
		if !item.RequiresTests {
			failed = append(failed, item.ID+":high_priority_requires_tests")
		}
		if item.ID != EconomicBacklogFeeAllocationInvariantTests && item.ID != EconomicBacklogSlashingRouteInvariantTests && !item.RequiresTelemetry && !item.RequiresQuery {
			failed = append(failed, item.ID+":high_priority_requires_telemetry_or_query")
		}
	}
	return failed
}

func validateEconomicNonGoal(nonGoal EconomicNonGoal) []string {
	failed := make([]string, 0)
	if strings.TrimSpace(nonGoal.Statement) == "" {
		failed = append(failed, nonGoal.ID+":statement_missing")
	}
	if !nonGoal.Tracked {
		failed = append(failed, nonGoal.ID+":not_tracked")
	}
	if len(nonGoal.Enforcement) == 0 {
		failed = append(failed, nonGoal.ID+":enforcement_missing")
	}
	if strings.Contains(nonGoal.ID, "nondeterministic") && !nonGoal.DeterminismGuard {
		failed = append(failed, nonGoal.ID+":determinism_guard_missing")
	}
	if strings.Contains(nonGoal.ID, "burn") && !nonGoal.SecurityBudgetGuard {
		failed = append(failed, nonGoal.ID+":security_budget_guard_missing")
	}
	for i, enforcement := range nonGoal.Enforcement {
		if strings.TrimSpace(enforcement) == "" {
			failed = append(failed, fmt.Sprintf("%s:enforcement_%d_blank", nonGoal.ID, i))
		}
	}
	return failed
}

func economicEngineeringBacklogItemCovered(item EconomicEngineeringBacklogItem) bool {
	return len(validateEconomicEngineeringBacklogItem(item)) == 0
}

func economicNonGoalCovered(nonGoal EconomicNonGoal) bool {
	return len(validateEconomicNonGoal(nonGoal)) == 0
}

func backlogPriorityRank(priority string) int {
	switch priority {
	case EconomicBacklogPriorityHigh:
		return 0
	case EconomicBacklogPriorityMedium:
		return 1
	case EconomicBacklogPriorityLower:
		return 2
	default:
		return 3
	}
}

func requiredEconomicEngineeringBacklogIDs() map[string][]string {
	return map[string][]string{
		EconomicBacklogPriorityHigh: {
			EconomicBacklogAddEconomicsLocalExclude,
			EconomicBacklogEpochEconomicReportDataModel,
			EconomicBacklogNetIssuanceAccounting,
			EconomicBacklogBurnAccountingQueries,
			EconomicBacklogBurnFeeDistribution,
			EconomicBacklogDeflationGuard,
			EconomicBacklogFeeAllocationInvariantTests,
			EconomicBacklogSlashingRouteInvariantTests,
			EconomicBacklogValidatorConcentrationQueries,
			EconomicBacklogStateGrowthTelemetry,
		},
		EconomicBacklogPriorityMedium: {
			EconomicBacklogValidatorRiskScoreQuery,
			EconomicBacklogCommissionChangeWarningEvent,
			EconomicBacklogDynamicBaseFeeSimulationTests,
			EconomicBacklogSenderLocalSpamSurchargeDesign,
			EconomicBacklogStoragePricingSpecification,
			EconomicBacklogSupplyProjectionCommandOrQuery,
			EconomicBacklogGovernanceImpactReport,
		},
		EconomicBacklogPriorityLower: {
			EconomicBacklogFullStateRentLifecycle,
			EconomicBacklogAdaptiveInflationController,
			EconomicBacklogSecurityReserveModule,
			EconomicBacklogFeeMarketCircuitBreaker,
			EconomicBacklogDelegationSimulator,
			EconomicBacklogValidatorBootstrapBand,
		},
	}
}

func requiredEconomicNonGoalIDs() map[string]struct{} {
	return map[string]struct{}{
		EconomicNonGoalSecondStakingAsset:			{},
		EconomicNonGoalExternalValidatorRewardAssets:		{},
		EconomicNonGoalOffChainFeeAccounting:			{},
		EconomicNonGoalNondeterministicConsensusReputation:	{},
		EconomicNonGoalUnverifiableValidatorIdentity:		{},
		EconomicNonGoalDiscretionarySlashing:			{},
		EconomicNonGoalBurnOverSecurityBudget:			{},
		EconomicNonGoalUnboundedControllers:			{},
	}
}
