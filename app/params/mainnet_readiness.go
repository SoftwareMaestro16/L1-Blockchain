package params

import (
	"fmt"
	"sort"
)

const (
	MainnetReadinessValidatorSetPolicy		= "validator_set_policy_implemented_and_tested"
	MainnetReadinessEffectivePowerCap		= "effective_power_cap_implemented_and_tested"
	MainnetReadinessAntiConcentrationRewards	= "anti_concentration_rewards_implemented_and_tested"
	MainnetReadinessDynamicInflation		= "dynamic_inflation_implemented_and_tested"
	MainnetReadinessFeeBurnTreasuryRewards		= "fee_burn_treasury_reward_split_implemented_and_tested"
	MainnetReadinessSlashing			= "slashing_configured_and_tested"
	MainnetReadinessAVM				= "avm_integrated_and_tested"
	MainnetReadinessExportImport			= "export_import_stable"
	MainnetReadinessStateSyncSnapshots		= "state_sync_snapshots_stable"
	MainnetReadinessPublicTestnetDuration		= "public_testnet_observed_validator_behavior"
	MainnetReadinessFinalityLoadTests		= "load_tests_demonstrate_finality_target"
	MainnetReadinessSecurityAudit			= "security_audit_completed"
	MainnetReadinessCriticalFindingsFixed		= "critical_findings_fixed"
	MainnetReadinessDocsComplete			= "docs_complete_for_validators_delegators_contract_developers"
)

type MainnetReadinessEvidence struct {
	ValidatorSetPolicyImplementedAndTested		bool
	EffectivePowerCapImplementedAndTested		bool
	AntiConcentrationRewardsImplementedAndTested	bool
	DynamicInflationImplementedAndTested		bool
	FeeBurnTreasuryRewardSplitImplementedAndTested	bool
	SlashingConfiguredAndTested			bool
	AVMIntegratedAndTested				bool
	ExportImportStable				bool
	StateSyncSnapshotsStable			bool
	PublicTestnetObservedValidatorBehavior		bool
	LoadTestsDemonstrateFinalityTarget		bool
	SecurityAuditCompleted				bool
	CriticalFindingsFixed				bool
	DocsComplete					bool
}

type MainnetReadinessReport struct {
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func ValidateMainnetReadiness(evidence MainnetReadinessEvidence) error {
	report := BuildMainnetReadinessReport(evidence)
	if !report.Ready {
		return fmt.Errorf("mainnet readiness failed: %v", report.Failed)
	}
	return nil
}

func BuildMainnetReadinessReport(evidence MainnetReadinessEvidence) MainnetReadinessReport {
	checks := []requirementCheck{
		{MainnetReadinessValidatorSetPolicy, evidence.ValidatorSetPolicyImplementedAndTested},
		{MainnetReadinessEffectivePowerCap, evidence.EffectivePowerCapImplementedAndTested},
		{MainnetReadinessAntiConcentrationRewards, evidence.AntiConcentrationRewardsImplementedAndTested},
		{MainnetReadinessDynamicInflation, evidence.DynamicInflationImplementedAndTested},
		{MainnetReadinessFeeBurnTreasuryRewards, evidence.FeeBurnTreasuryRewardSplitImplementedAndTested},
		{MainnetReadinessSlashing, evidence.SlashingConfiguredAndTested},
		{MainnetReadinessAVM, evidence.AVMIntegratedAndTested},
		{MainnetReadinessExportImport, evidence.ExportImportStable},
		{MainnetReadinessStateSyncSnapshots, evidence.StateSyncSnapshotsStable},
		{MainnetReadinessPublicTestnetDuration, evidence.PublicTestnetObservedValidatorBehavior},
		{MainnetReadinessFinalityLoadTests, evidence.LoadTestsDemonstrateFinalityTarget},
		{MainnetReadinessSecurityAudit, evidence.SecurityAuditCompleted},
		{MainnetReadinessCriticalFindingsFixed, evidence.CriticalFindingsFixed},
		{MainnetReadinessDocsComplete, evidence.DocsComplete},
	}

	failed := make([]string, 0)
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	sort.Strings(failed)
	return MainnetReadinessReport{
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}
