package params

import (
	"fmt"
	"sort"
)

const (
	AetraAcceptanceCategoryBaseNode			= "base_node"
	AetraAcceptanceCategoryStaking			= "staking"
	AetraAcceptanceCategoryAntiCentralization	= "anti_centralization"
	AetraAcceptanceCategorySlashing			= "slashing"
	AetraAcceptanceCategoryEconomics		= "economics"
	AetraAcceptanceCategoryAVM			= "avm"
	AetraAcceptanceCategoryGovernance		= "governance"
	AetraAcceptanceCategoryObservability		= "observability"
)

const (
	AetraAcceptanceBaseNodeBootSingle		= "boot_single_node"
	AetraAcceptanceBaseNodeBootMultiValidator	= "boot_multi_validator_localnet"
	AetraAcceptanceBaseNodeRestart			= "restart"
	AetraAcceptanceBaseNodeExportImport		= "export_import"
	AetraAcceptanceBaseNodeStateSyncSnapshotRestore	= "state_sync_or_snapshot_restore"
)

const (
	AetraAcceptanceStakingCreateValidator		= "create_validator"
	AetraAcceptanceStakingDelegate			= "delegate"
	AetraAcceptanceStakingRedelegate		= "redelegate"
	AetraAcceptanceStakingUnbond			= "unbond"
	AetraAcceptanceStakingWithdrawRewards		= "withdraw_rewards"
	AetraAcceptanceStakingValidatorCommissionUpdate	= "validator_commission_update"
)

const (
	AetraAcceptanceAntiCentralizationValidatorReachesCap	= "validator_reaches_cap"
	AetraAcceptanceAntiCentralizationValidatorExceedsCap	= "validator_exceeds_cap"
	AetraAcceptanceAntiCentralizationRewardPenalty		= "excess_stake_reward_penalty_applied"
	AetraAcceptanceAntiCentralizationTopNQuery		= "top_n_concentration_query_works"
	AetraAcceptanceAntiCentralizationCommissionFloor	= "commission_floor_enforced"
)

const (
	AetraAcceptanceSlashingDowntimeTracked		= "downtime_tracked"
	AetraAcceptanceSlashingDowntimeJail		= "downtime_jail"
	AetraAcceptanceSlashingDoubleSignEvidence	= "double_sign_evidence_path_where_feasible"
	AetraAcceptanceSlashingTombstoneBehavior	= "tombstone_behavior"
	AetraAcceptanceSlashingDelegatorAccounting	= "delegator_slash_accounting"
)

const (
	AetraAcceptanceEconomicsInflationUpdate		= "inflation_update"
	AetraAcceptanceEconomicsFeeBurn			= "fee_burn"
	AetraAcceptanceEconomicsTreasuryAllocation	= "treasury_allocation"
	AetraAcceptanceEconomicsRewardsAllocation	= "rewards_allocation"
	AetraAcceptanceEconomicsAPRQuery		= "apr_query"
	AetraAcceptanceEconomicsSupplyInvariant		= "supply_invariant"
)

const (
	AetraAcceptanceAVMUploadCode			= "upload_code"
	AetraAcceptanceAVMInstantiate			= "instantiate"
	AetraAcceptanceAVMExecute			= "execute"
	AetraAcceptanceAVMQuery				= "query"
	AetraAcceptanceAVMMigrateIfEnabled		= "migrate_if_enabled"
	AetraAcceptanceAVMGasExhaustionContained	= "gas_exhaustion_contained"
)

const (
	AetraAcceptanceGovernanceValidParamProposal		= "valid_param_proposal"
	AetraAcceptanceGovernanceInvalidParamProposal		= "invalid_param_proposal"
	AetraAcceptanceGovernanceTreasuryProposal		= "treasury_proposal"
	AetraAcceptanceGovernanceDelayedCriticalActivation	= "delayed_critical_param_activation"
)

const (
	AetraAcceptanceObservabilityPrometheusMetrics	= "prometheus_metrics"
	AetraAcceptanceObservabilityCLIQueries		= "cli_queries"
	AetraAcceptanceObservabilityGRPCQueries		= "grpc_queries"
	AetraAcceptanceObservabilityEventsIndexable	= "events_indexable"
)

type AetraAcceptanceCategoryEvidence struct {
	Category	string
	Scenarios	[]string
}

type AetraAcceptanceMatrixReport struct {
	Categories	[]AetraAcceptanceCategoryEvidence
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraAcceptanceMatrixEvidence() []AetraAcceptanceCategoryEvidence {
	return []AetraAcceptanceCategoryEvidence{
		{Category: AetraAcceptanceCategoryBaseNode, Scenarios: RequiredAetraAcceptanceBaseNodeScenarios()},
		{Category: AetraAcceptanceCategoryStaking, Scenarios: RequiredAetraAcceptanceStakingScenarios()},
		{Category: AetraAcceptanceCategoryAntiCentralization, Scenarios: RequiredAetraAcceptanceAntiCentralizationScenarios()},
		{Category: AetraAcceptanceCategorySlashing, Scenarios: RequiredAetraAcceptanceSlashingScenarios()},
		{Category: AetraAcceptanceCategoryEconomics, Scenarios: RequiredAetraAcceptanceEconomicsScenarios()},
		{Category: AetraAcceptanceCategoryAVM, Scenarios: RequiredAetraAcceptanceAVMScenarios()},
		{Category: AetraAcceptanceCategoryGovernance, Scenarios: RequiredAetraAcceptanceGovernanceScenarios()},
		{Category: AetraAcceptanceCategoryObservability, Scenarios: RequiredAetraAcceptanceObservabilityScenarios()},
	}
}

func ValidateAetraAcceptanceMatrix(evidence []AetraAcceptanceCategoryEvidence) error {
	report := BuildAetraAcceptanceMatrixReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra acceptance matrix failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraAcceptanceMatrixReport(evidence []AetraAcceptanceCategoryEvidence) AetraAcceptanceMatrixReport {
	if evidence == nil {
		evidence = DefaultAetraAcceptanceMatrixEvidence()
	}
	evidence = normalizeAcceptanceCategories(evidence)
	requiredCategories := requiredAcceptanceCategories()
	seen := map[string]AetraAcceptanceCategoryEvidence{}
	failed := make([]string, 0)
	required := 0
	passed := 0

	for _, category := range evidence {
		if category.Category == "" {
			failed = append(failed, "category_required")
			continue
		}
		if _, duplicate := seen[category.Category]; duplicate {
			failed = append(failed, category.Category+":duplicate_category")
		}
		seen[category.Category] = category
		requiredScenarios, known := requiredCategories[category.Category]
		if !known {
			failed = append(failed, category.Category+":unknown_category")
			continue
		}
		required += len(requiredScenarios)
		passedCategory, failedCategory := validateAcceptanceScenarioCatalog(category.Category, category.Scenarios, requiredScenarios)
		passed += passedCategory
		failed = append(failed, failedCategory...)
	}
	for category := range requiredCategories {
		if _, ok := seen[category]; !ok {
			failed = append(failed, category+":missing_category")
			required += len(requiredCategories[category])
		}
	}

	sort.Strings(failed)
	return AetraAcceptanceMatrixReport{
		Categories:	evidence,
		Required:	required,
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func RequiredAetraAcceptanceBaseNodeScenarios() []string {
	return []string{
		AetraAcceptanceBaseNodeBootSingle,
		AetraAcceptanceBaseNodeBootMultiValidator,
		AetraAcceptanceBaseNodeRestart,
		AetraAcceptanceBaseNodeExportImport,
		AetraAcceptanceBaseNodeStateSyncSnapshotRestore,
	}
}

func RequiredAetraAcceptanceStakingScenarios() []string {
	return []string{
		AetraAcceptanceStakingCreateValidator,
		AetraAcceptanceStakingDelegate,
		AetraAcceptanceStakingRedelegate,
		AetraAcceptanceStakingUnbond,
		AetraAcceptanceStakingWithdrawRewards,
		AetraAcceptanceStakingValidatorCommissionUpdate,
	}
}

func RequiredAetraAcceptanceAntiCentralizationScenarios() []string {
	return []string{
		AetraAcceptanceAntiCentralizationValidatorReachesCap,
		AetraAcceptanceAntiCentralizationValidatorExceedsCap,
		AetraAcceptanceAntiCentralizationRewardPenalty,
		AetraAcceptanceAntiCentralizationTopNQuery,
		AetraAcceptanceAntiCentralizationCommissionFloor,
	}
}

func RequiredAetraAcceptanceSlashingScenarios() []string {
	return []string{
		AetraAcceptanceSlashingDowntimeTracked,
		AetraAcceptanceSlashingDowntimeJail,
		AetraAcceptanceSlashingDoubleSignEvidence,
		AetraAcceptanceSlashingTombstoneBehavior,
		AetraAcceptanceSlashingDelegatorAccounting,
	}
}

func RequiredAetraAcceptanceEconomicsScenarios() []string {
	return []string{
		AetraAcceptanceEconomicsInflationUpdate,
		AetraAcceptanceEconomicsFeeBurn,
		AetraAcceptanceEconomicsTreasuryAllocation,
		AetraAcceptanceEconomicsRewardsAllocation,
		AetraAcceptanceEconomicsAPRQuery,
		AetraAcceptanceEconomicsSupplyInvariant,
	}
}

func RequiredAetraAcceptanceAVMScenarios() []string {
	return []string{
		AetraAcceptanceAVMUploadCode,
		AetraAcceptanceAVMInstantiate,
		AetraAcceptanceAVMExecute,
		AetraAcceptanceAVMQuery,
		AetraAcceptanceAVMMigrateIfEnabled,
		AetraAcceptanceAVMGasExhaustionContained,
	}
}

func RequiredAetraAcceptanceGovernanceScenarios() []string {
	return []string{
		AetraAcceptanceGovernanceValidParamProposal,
		AetraAcceptanceGovernanceInvalidParamProposal,
		AetraAcceptanceGovernanceTreasuryProposal,
		AetraAcceptanceGovernanceDelayedCriticalActivation,
	}
}

func RequiredAetraAcceptanceObservabilityScenarios() []string {
	return []string{
		AetraAcceptanceObservabilityPrometheusMetrics,
		AetraAcceptanceObservabilityCLIQueries,
		AetraAcceptanceObservabilityGRPCQueries,
		AetraAcceptanceObservabilityEventsIndexable,
	}
}

func requiredAcceptanceCategories() map[string][]string {
	return map[string][]string{
		AetraAcceptanceCategoryBaseNode:		RequiredAetraAcceptanceBaseNodeScenarios(),
		AetraAcceptanceCategoryStaking:			RequiredAetraAcceptanceStakingScenarios(),
		AetraAcceptanceCategoryAntiCentralization:	RequiredAetraAcceptanceAntiCentralizationScenarios(),
		AetraAcceptanceCategorySlashing:		RequiredAetraAcceptanceSlashingScenarios(),
		AetraAcceptanceCategoryEconomics:		RequiredAetraAcceptanceEconomicsScenarios(),
		AetraAcceptanceCategoryAVM:			RequiredAetraAcceptanceAVMScenarios(),
		AetraAcceptanceCategoryGovernance:		RequiredAetraAcceptanceGovernanceScenarios(),
		AetraAcceptanceCategoryObservability:		RequiredAetraAcceptanceObservabilityScenarios(),
	}
}

func validateAcceptanceScenarioCatalog(category string, actual []string, required []string) (int, []string) {
	requiredSet := map[string]bool{}
	actualCounts := map[string]int{}
	for _, item := range required {
		requiredSet[item] = true
	}
	for _, item := range actual {
		actualCounts[item]++
	}

	failed := make([]string, 0)
	passed := 0
	for _, item := range required {
		switch actualCounts[item] {
		case 0:
			failed = append(failed, category+"."+item+":missing")
		case 1:
			passed++
		default:
			failed = append(failed, category+"."+item+":duplicate")
		}
	}
	for item := range actualCounts {
		if !requiredSet[item] {
			failed = append(failed, category+"."+item+":unexpected")
		}
	}
	return passed, failed
}

func normalizeAcceptanceCategories(categories []AetraAcceptanceCategoryEvidence) []AetraAcceptanceCategoryEvidence {
	out := append([]AetraAcceptanceCategoryEvidence{}, categories...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Category < out[j].Category })
	return out
}
