package params

import (
	"fmt"
	"sort"
	"strings"
)

const (
	OpenDesignDecisionRewardDampeningScope	= "validator_reward_dampening_scope"
	OpenDesignDecisionStateRentActivation	= "state_rent_activation_scope"
	OpenDesignDecisionDeleteRefundPayment	= "storage_delete_refund_payment"
	OpenDesignDecisionBootstrapFunding	= "validator_bootstrap_funding_source"
	OpenDesignDecisionRiskYieldLocation	= "risk_adjusted_yield_query_location"
	OpenDesignDecisionFeeBucketWeights	= "fee_bucket_weight_control"
	OpenDesignDecisionSecurityReserveSpend	= "security_reserve_spending_authority"

	OpenDesignDecisionStatusOpen	= "open"
)

type EconomicOpenDesignDecision struct {
	ID				string
	Question			string
	Status				string
	Options				[]string
	LinkedGovernanceParams		[]string
	RequiredBeforeActivation	string
	Queryable			bool
	GovernanceRequired		bool
	ConsensusImpact			bool
	SimulationRequired		bool
	ImplementationMustNotPick	bool
	Resolution			string
}

type EconomicOpenDesignDecisionReport struct {
	Decisions	[]EconomicOpenDesignDecision
	Required	int
	Covered		int
	CoverageBps	int64
	Passed		bool
	Failed		[]string
	Summary		string
}

func DefaultEconomicOpenDesignDecisions() []EconomicOpenDesignDecision {
	return []EconomicOpenDesignDecision{
		openDesignDecision(
			OpenDesignDecisionRewardDampeningScope,
			"whether validator reward dampening affects validator commission, total validator-delegator rewards, or both",
			[]string{"validator_commission_only", "total_validator_delegator_rewards", "commission_and_total_rewards"},
			"activate_concentration_reward_dampening",
			[]string{GovernanceParamRewardDampeningCurve, GovernanceParamConcentrationSoftCap},
			true,
		),
		openDesignDecision(
			OpenDesignDecisionStateRentActivation,
			"whether state rent applies immediately to all state or only to state created after activation",
			[]string{"all_existing_and_new_state", "new_state_only", "grandfather_existing_state_until_migration_epoch"},
			"activate_state_rent",
			[]string{GovernanceParamRentRate, GovernanceParamRentGracePeriod},
			true,
		),
		openDesignDecision(
			OpenDesignDecisionDeleteRefundPayment,
			"whether storage delete refunds are paid immediately or credited against future storage fees",
			[]string{"immediate_refund", "future_storage_fee_credit", "hybrid_capped_immediate_credit"},
			"activate_storage_delete_refunds",
			[]string{GovernanceParamDeleteRefundCap, GovernanceParamDeleteRefundDecay},
			true,
		),
		openDesignDecision(
			OpenDesignDecisionBootstrapFunding,
			"whether the validator bootstrap band is funded by inflation, fee allocation, or reward redistribution",
			[]string{"inflation_funded", "fee_allocation_funded", "reward_redistribution_funded"},
			"activate_validator_bootstrap_band",
			[]string{GovernanceParamBootstrapEligibility, GovernanceParamValidatorRewardFloor, GovernanceParamFeeAllocationBucketWeights},
			true,
		),
		openDesignDecision(
			OpenDesignDecisionRiskYieldLocation,
			"whether risk-adjusted yield estimates are chain-native queries or maintained by indexers using chain-native data",
			[]string{"chain_native_query", "indexer_derived_from_chain_data", "chain_inputs_with_indexer_projection"},
			"activate_delegator_risk_adjusted_yield",
			[]string{GovernanceParamValidatorScoreWeights, GovernanceParamMaximumValidatorCommission, GovernanceParamConcentrationSoftCap},
			false,
		),
		openDesignDecision(
			OpenDesignDecisionFeeBucketWeights,
			"whether fee bucket weights are static governance parameters or controller-adjusted within bounds",
			[]string{"static_governance_parameters", "controller_adjusted_within_bounds", "static_base_with_activity_adjustment_caps"},
			"activate_bucketed_fee_distribution",
			[]string{GovernanceParamFeeAllocationBucketWeights, GovernanceParamFeeBurnAllocation, GovernanceParamStateMaintenanceReserveAllocation, GovernanceParamSecurityReserveAllocation},
			true,
		),
		openDesignDecision(
			OpenDesignDecisionSecurityReserveSpend,
			"whether security reserve spending requires governance approval, automatic trigger conditions, or both",
			[]string{"governance_approval_only", "automatic_trigger_conditions", "automatic_triggers_with_governance_ratification"},
			"activate_security_reserve_spending",
			[]string{GovernanceParamSecurityReserveAllocation, GovernanceParamCircuitBreakerThresholds},
			true,
		),
	}
}

func BuildEconomicOpenDesignDecisionReport(decisions []EconomicOpenDesignDecision) EconomicOpenDesignDecisionReport {
	if decisions == nil {
		decisions = DefaultEconomicOpenDesignDecisions()
	}
	out, failed, required, covered := evaluateEconomicOpenDesignDecisions(decisions)
	sort.Strings(failed)
	coverage := coverageBps(covered, required)
	return EconomicOpenDesignDecisionReport{
		Decisions:	out,
		Required:	required,
		Covered:	covered,
		CoverageBps:	coverage,
		Passed:		len(failed) == 0 && coverage == BasisPoints,
		Failed:		failed,
		Summary:	fmt.Sprintf("open_design_decisions=%d/%d coverage_bps=%d", covered, required, coverage),
	}
}

func openDesignDecision(id, question string, options []string, requiredBefore string, linkedParams []string, consensusImpact bool) EconomicOpenDesignDecision {
	return EconomicOpenDesignDecision{
		ID:				id,
		Question:			question,
		Status:				OpenDesignDecisionStatusOpen,
		Options:			append([]string{}, options...),
		LinkedGovernanceParams:		append([]string{}, linkedParams...),
		RequiredBeforeActivation:	requiredBefore,
		Queryable:			true,
		GovernanceRequired:		true,
		ConsensusImpact:		consensusImpact,
		SimulationRequired:		true,
		ImplementationMustNotPick:	true,
	}
}

func evaluateEconomicOpenDesignDecisions(decisions []EconomicOpenDesignDecision) ([]EconomicOpenDesignDecision, []string, int, int) {
	out := append([]EconomicOpenDesignDecision{}, decisions...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	expected := requiredEconomicOpenDesignDecisionIDs()
	governanceParams := knownEconomicGovernanceParameterIDs()
	failed := make([]string, 0)
	seen := make(map[string]EconomicOpenDesignDecision, len(out))
	for _, decision := range out {
		if decision.ID == "" {
			failed = append(failed, "open_design_decision_id_required")
			continue
		}
		if _, ok := expected[decision.ID]; !ok {
			failed = append(failed, decision.ID+":unknown_open_design_decision")
		}
		if _, ok := seen[decision.ID]; ok {
			failed = append(failed, decision.ID+":duplicate_open_design_decision")
		}
		seen[decision.ID] = decision
		failed = append(failed, validateEconomicOpenDesignDecision(decision, governanceParams)...)
	}
	covered := 0
	for id := range expected {
		decision, ok := seen[id]
		if !ok {
			failed = append(failed, id+":missing_required_open_design_decision")
			continue
		}
		if economicOpenDesignDecisionCovered(decision, governanceParams) {
			covered++
		}
	}
	return out, failed, len(expected), covered
}

func validateEconomicOpenDesignDecision(decision EconomicOpenDesignDecision, governanceParams map[string]struct{}) []string {
	failed := make([]string, 0)
	if strings.TrimSpace(decision.Question) == "" {
		failed = append(failed, decision.ID+":question_missing")
	}
	if decision.Status != OpenDesignDecisionStatusOpen {
		failed = append(failed, decision.ID+":status_must_remain_open")
	}
	if strings.TrimSpace(decision.Resolution) != "" {
		failed = append(failed, decision.ID+":resolution_must_be_empty_until_governance_decision")
	}
	if len(decision.Options) < 2 {
		failed = append(failed, decision.ID+":options_missing")
	}
	if strings.TrimSpace(decision.RequiredBeforeActivation) == "" {
		failed = append(failed, decision.ID+":activation_gate_missing")
	}
	if len(decision.LinkedGovernanceParams) == 0 {
		failed = append(failed, decision.ID+":governance_param_link_missing")
	}
	if !decision.Queryable {
		failed = append(failed, decision.ID+":not_queryable")
	}
	if !decision.GovernanceRequired {
		failed = append(failed, decision.ID+":governance_required_missing")
	}
	if !decision.SimulationRequired {
		failed = append(failed, decision.ID+":simulation_required_missing")
	}
	if !decision.ImplementationMustNotPick {
		failed = append(failed, decision.ID+":implementation_must_not_pick_missing")
	}
	for i, option := range decision.Options {
		if strings.TrimSpace(option) == "" {
			failed = append(failed, fmt.Sprintf("%s:option_%d_blank", decision.ID, i))
		}
	}
	for _, param := range decision.LinkedGovernanceParams {
		if _, ok := governanceParams[param]; !ok {
			failed = append(failed, decision.ID+":unknown_linked_governance_param:"+param)
		}
	}
	return failed
}

func economicOpenDesignDecisionCovered(decision EconomicOpenDesignDecision, governanceParams map[string]struct{}) bool {
	return len(validateEconomicOpenDesignDecision(decision, governanceParams)) == 0
}

func requiredEconomicOpenDesignDecisionIDs() map[string]struct{} {
	return map[string]struct{}{
		OpenDesignDecisionRewardDampeningScope:	{},
		OpenDesignDecisionStateRentActivation:	{},
		OpenDesignDecisionDeleteRefundPayment:	{},
		OpenDesignDecisionBootstrapFunding:	{},
		OpenDesignDecisionRiskYieldLocation:	{},
		OpenDesignDecisionFeeBucketWeights:	{},
		OpenDesignDecisionSecurityReserveSpend:	{},
	}
}

func knownEconomicGovernanceParameterIDs() map[string]struct{} {
	out := make(map[string]struct{})
	for _, ids := range requiredEconomicGovernanceParameterIDs() {
		for _, id := range ids {
			out[id] = struct{}{}
		}
	}
	return out
}
