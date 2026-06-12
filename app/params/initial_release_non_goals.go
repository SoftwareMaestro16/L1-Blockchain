package params

import (
	"fmt"
	"sort"
)

const (
	InitialReleaseNonGoalPoH			= "poh"
	InitialReleaseNonGoalSolanaLevelTPS		= "solana_level_tps"
	InitialReleaseNonGoalOneSecondBlocks		= "one_second_blocks"
	InitialReleaseNonGoalMandatoryKYC		= "mandatory_kyc_validator_admission"
	InitialReleaseNonGoalEVMAtGenesis		= "evm_at_genesis_unless_separately_approved"
	InitialReleaseNonGoalSubjectiveSlashing		= "subjective_slashing"
	InitialReleaseNonGoalUnlimitedValidatorSet	= "unlimited_validator_set"
	InitialReleaseNonGoalUnboundedContractExecution	= "unbounded_contract_execution"
	InitialReleaseNonGoalHighInflationAPRMarketing	= "high_inflation_apr_marketing"
)

type InitialReleaseScopePolicy struct {
	AttemptsPoH				bool
	AttemptsSolanaLevelTPS			bool
	AllowsOneSecondBlocks			bool
	RequiresMandatoryKYCValidators		bool
	EnablesEVMAtGenesisWithoutApproval	bool
	AllowsSubjectiveSlashing		bool
	AllowsUnlimitedValidatorSet		bool
	AllowsUnboundedContractExecution	bool
	UsesHighInflationAPRMarketing		bool
}

type InitialReleaseNonGoalsReport struct {
	NonGoalsChecked	int
	Violations	[]string
	Allowed		bool
}

func DefaultInitialReleaseScopePolicy() InitialReleaseScopePolicy {
	return InitialReleaseScopePolicy{}
}

func ValidateInitialReleaseScope(policy InitialReleaseScopePolicy) error {
	report := BuildInitialReleaseNonGoalsReport(policy)
	if !report.Allowed {
		return fmt.Errorf("initial release scope violates non-goals: %v", report.Violations)
	}
	return nil
}

func BuildInitialReleaseNonGoalsReport(policy InitialReleaseScopePolicy) InitialReleaseNonGoalsReport {
	checks := []requirementCheck{
		{InitialReleaseNonGoalPoH, !policy.AttemptsPoH},
		{InitialReleaseNonGoalSolanaLevelTPS, !policy.AttemptsSolanaLevelTPS},
		{InitialReleaseNonGoalOneSecondBlocks, !policy.AllowsOneSecondBlocks},
		{InitialReleaseNonGoalMandatoryKYC, !policy.RequiresMandatoryKYCValidators},
		{InitialReleaseNonGoalEVMAtGenesis, !policy.EnablesEVMAtGenesisWithoutApproval},
		{InitialReleaseNonGoalSubjectiveSlashing, !policy.AllowsSubjectiveSlashing},
		{InitialReleaseNonGoalUnlimitedValidatorSet, !policy.AllowsUnlimitedValidatorSet},
		{InitialReleaseNonGoalUnboundedContractExecution, !policy.AllowsUnboundedContractExecution},
		{InitialReleaseNonGoalHighInflationAPRMarketing, !policy.UsesHighInflationAPRMarketing},
	}

	violations := make([]string, 0)
	for _, check := range checks {
		if !check.Passed {
			violations = append(violations, check.ID)
		}
	}
	sort.Strings(violations)
	return InitialReleaseNonGoalsReport{
		NonGoalsChecked:	len(checks),
		Violations:		violations,
		Allowed:		len(violations) == 0,
	}
}
