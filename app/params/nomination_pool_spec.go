package params

import (
	"fmt"
	"sort"
)

const (
	AetraNominationPoolModuleName	= "x/nominator-pool"

	AetraNominationPoolPurposeAccessibilityWithAccountingAndCentralizationRisk	= "nomination_pools_improve_accessibility_but_introduce_accounting_and_centralization_risks"

	AetraNominationPoolStatePool		= "Pool"
	AetraNominationPoolStatePoolDelegation	= "PoolDelegation"

	AetraNominationPoolFieldPoolID			= "PoolId"
	AetraNominationPoolFieldOperatorAddress		= "OperatorAddress"
	AetraNominationPoolFieldValidatorAddress	= "ValidatorAddress"
	AetraNominationPoolFieldTotalBonded		= "TotalBonded"
	AetraNominationPoolFieldTotalShares		= "TotalShares"
	AetraNominationPoolFieldCommissionBps		= "CommissionBps"
	AetraNominationPoolFieldStatus			= "Status"
	AetraNominationPoolFieldCreatedHeight		= "CreatedHeight"
	AetraNominationPoolFieldUnbondingEntries	= "UnbondingEntries"

	AetraNominationPoolFieldDelegatorAddress	= "DelegatorAddress"
	AetraNominationPoolFieldDelegationPoolID	= "PoolId"
	AetraNominationPoolFieldShares			= "Shares"
	AetraNominationPoolFieldPrincipalEstimate	= "PrincipalEstimate"
	AetraNominationPoolFieldRewardsAccrued		= "RewardsAccrued"

	AetraNominationPoolRiskAccessibility	= "accessibility_for_users_without_validator_infrastructure"
	AetraNominationPoolRiskAccounting	= "deterministic_pool_share_principal_reward_accounting"
	AetraNominationPoolRiskCentralization	= "pool_operator_and_validator_concentration_risk"
	AetraNominationPoolImplementationMap	= "current_x_nominator_pool_state_mapping_required"

	AetraNominationPoolRequirementNativeStakingDenom	= "users_deposit_native_staking_denom"
	AetraNominationPoolRequirementDeterministicShareMint	= "pool_mints_shares_deterministically"
	AetraNominationPoolRequirementDelegatesToValidator	= "pool_delegates_to_validator"
	AetraNominationPoolRequirementProRataRewards		= "pool_distributes_rewards_pro_rata"
	AetraNominationPoolRequirementCommissionBounded		= "pool_commission_bounded"
	AetraNominationPoolRequirementWithdrawalUnbonding	= "pool_withdrawal_follows_unbonding_period"
	AetraNominationPoolRequirementSlashingReducesShare	= "pool_slashing_reduces_share_value"
	AetraNominationPoolRequirementOperatorNoPrincipalTheft	= "pool_operator_cannot_withdraw_user_principal"
	AetraNominationPoolRequirementCannotBypassPowerCap	= "pool_cannot_bypass_validator_power_cap"
	AetraNominationPoolRequirementRiskWarnings		= "pool_must_expose_risk_warnings"

	AetraNominationPoolTestFirstDepositSharePrice		= "first_deposit_share_price"
	AetraNominationPoolTestSubsequentDepositSharePrice	= "subsequent_deposit_share_price"
	AetraNominationPoolTestRewardDistribution		= "reward_distribution"
	AetraNominationPoolTestCommissionDeduction		= "commission_deduction"
	AetraNominationPoolTestPartialWithdrawal		= "partial_withdrawal"
	AetraNominationPoolTestFullWithdrawal			= "full_withdrawal"
	AetraNominationPoolTestSlashingPoolValidator		= "slashing_pool_validator"
	AetraNominationPoolTestJailedValidator			= "jailed_validator"
	AetraNominationPoolTestRedelegationIfAllowed		= "redelegation_if_allowed"
	AetraNominationPoolTestOperatorAbuseAttempt		= "pool_operator_abuse_attempt"
	AetraNominationPoolTestExportImportActiveUnbonding	= "export_import_with_active_unbonding_entries"
	AetraNominationPoolTestRoundingDustHandling		= "rounding_dust_handling"
)

type AetraNominationPoolModelEvidence struct {
	ModuleName	string

	PoolFields		[]string
	PoolDelegationFields	[]string

	AccessibilityRiskAcknowledged	bool
	AccountingRiskAcknowledged	bool
	CentralizationRiskAcknowledged	bool
	CurrentImplementationMapped	bool
}

type AetraNominationPoolModelReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraNominationPoolRequirementsEvidence struct {
	ModuleName	string

	UsersDepositNativeStakingDenom		bool
	MintsSharesDeterministically		bool
	DelegatesToValidator			bool
	DistributesRewardsProRata		bool
	CommissionBounded			bool
	WithdrawalFollowsUnbondingPeriod	bool
	SlashingReducesShareValue		bool
	OperatorCannotWithdrawPrincipal		bool
	CannotBypassValidatorPowerCap		bool
	ExposesRiskWarnings			bool
}

type AetraNominationPoolRequirementsReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

type AetraNominationPoolTestingEvidence struct {
	ModuleName	string

	FirstDepositSharePrice		bool
	SubsequentDepositSharePrice	bool
	RewardDistribution		bool
	CommissionDeduction		bool
	PartialWithdrawal		bool
	FullWithdrawal			bool
	SlashingPoolValidator		bool
	JailedValidator			bool
	RedelegationIfAllowed		bool
	PoolOperatorAbuseAttempt	bool
	ExportImportActiveUnbonding	bool
	RoundingDustHandling		bool
}

type AetraNominationPoolTestingReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraNominationPoolModelEvidence() AetraNominationPoolModelEvidence {
	return AetraNominationPoolModelEvidence{
		ModuleName:	AetraNominationPoolModuleName,
		PoolFields: []string{
			AetraNominationPoolFieldPoolID,
			AetraNominationPoolFieldOperatorAddress,
			AetraNominationPoolFieldValidatorAddress,
			AetraNominationPoolFieldTotalBonded,
			AetraNominationPoolFieldTotalShares,
			AetraNominationPoolFieldCommissionBps,
			AetraNominationPoolFieldStatus,
			AetraNominationPoolFieldCreatedHeight,
			AetraNominationPoolFieldUnbondingEntries,
		},
		PoolDelegationFields: []string{
			AetraNominationPoolFieldDelegatorAddress,
			AetraNominationPoolFieldDelegationPoolID,
			AetraNominationPoolFieldShares,
			AetraNominationPoolFieldPrincipalEstimate,
			AetraNominationPoolFieldRewardsAccrued,
		},
		AccessibilityRiskAcknowledged:	true,
		AccountingRiskAcknowledged:	true,
		CentralizationRiskAcknowledged:	true,
		CurrentImplementationMapped:	true,
	}
}

func ValidateAetraNominationPoolModel(evidence AetraNominationPoolModelEvidence) error {
	report := BuildAetraNominationPoolModelReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra nomination pool model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraNominationPoolModelReport(evidence AetraNominationPoolModelEvidence) AetraNominationPoolModelReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraNominationPoolModuleName {
		failed = append(failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	}

	requiredPool := requiredAetraNominationPoolFields()
	requiredDelegation := requiredAetraNominationPoolDelegationFields()
	passedPool, failedPool := validateAetraNominationPoolCatalog(AetraNominationPoolStatePool, evidence.PoolFields, requiredPool)
	passedDelegation, failedDelegation := validateAetraNominationPoolCatalog(AetraNominationPoolStatePoolDelegation, evidence.PoolDelegationFields, requiredDelegation)
	failed = append(failed, failedPool...)
	failed = append(failed, failedDelegation...)

	checks := []requirementCheck{
		{AetraNominationPoolRiskAccessibility, evidence.AccessibilityRiskAcknowledged},
		{AetraNominationPoolRiskAccounting, evidence.AccountingRiskAcknowledged},
		{AetraNominationPoolRiskCentralization, evidence.CentralizationRiskAcknowledged},
		{AetraNominationPoolImplementationMap, evidence.CurrentImplementationMapped},
	}
	passedChecks := 0
	for _, check := range checks {
		if check.Passed {
			passedChecks++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraNominationPoolModelReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(requiredPool) + len(requiredDelegation) + len(checks),
		Passed:		passedPool + passedDelegation + passedChecks,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraNominationPoolRequirementsEvidence() AetraNominationPoolRequirementsEvidence {
	return AetraNominationPoolRequirementsEvidence{
		ModuleName:	AetraNominationPoolModuleName,

		UsersDepositNativeStakingDenom:		true,
		MintsSharesDeterministically:		true,
		DelegatesToValidator:			true,
		DistributesRewardsProRata:		true,
		CommissionBounded:			true,
		WithdrawalFollowsUnbondingPeriod:	true,
		SlashingReducesShareValue:		true,
		OperatorCannotWithdrawPrincipal:	true,
		CannotBypassValidatorPowerCap:		true,
		ExposesRiskWarnings:			true,
	}
}

func ValidateAetraNominationPoolRequirements(evidence AetraNominationPoolRequirementsEvidence) error {
	report := BuildAetraNominationPoolRequirementsReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra nomination pool requirements failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraNominationPoolRequirementsReport(evidence AetraNominationPoolRequirementsEvidence) AetraNominationPoolRequirementsReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraNominationPoolModuleName {
		failed = append(failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	}
	checks := []requirementCheck{
		{AetraNominationPoolRequirementNativeStakingDenom, evidence.UsersDepositNativeStakingDenom},
		{AetraNominationPoolRequirementDeterministicShareMint, evidence.MintsSharesDeterministically},
		{AetraNominationPoolRequirementDelegatesToValidator, evidence.DelegatesToValidator},
		{AetraNominationPoolRequirementProRataRewards, evidence.DistributesRewardsProRata},
		{AetraNominationPoolRequirementCommissionBounded, evidence.CommissionBounded},
		{AetraNominationPoolRequirementWithdrawalUnbonding, evidence.WithdrawalFollowsUnbondingPeriod},
		{AetraNominationPoolRequirementSlashingReducesShare, evidence.SlashingReducesShareValue},
		{AetraNominationPoolRequirementOperatorNoPrincipalTheft, evidence.OperatorCannotWithdrawPrincipal},
		{AetraNominationPoolRequirementCannotBypassPowerCap, evidence.CannotBypassValidatorPowerCap},
		{AetraNominationPoolRequirementRiskWarnings, evidence.ExposesRiskWarnings},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	sort.Strings(failed)
	return AetraNominationPoolRequirementsReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultAetraNominationPoolTestingEvidence() AetraNominationPoolTestingEvidence {
	return AetraNominationPoolTestingEvidence{
		ModuleName:	AetraNominationPoolModuleName,

		FirstDepositSharePrice:		true,
		SubsequentDepositSharePrice:	true,
		RewardDistribution:		true,
		CommissionDeduction:		true,
		PartialWithdrawal:		true,
		FullWithdrawal:			true,
		SlashingPoolValidator:		true,
		JailedValidator:		true,
		RedelegationIfAllowed:		true,
		PoolOperatorAbuseAttempt:	true,
		ExportImportActiveUnbonding:	true,
		RoundingDustHandling:		true,
	}
}

func ValidateAetraNominationPoolTesting(evidence AetraNominationPoolTestingEvidence) error {
	report := BuildAetraNominationPoolTestingReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra nomination pool testing failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraNominationPoolTestingReport(evidence AetraNominationPoolTestingEvidence) AetraNominationPoolTestingReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraNominationPoolModuleName {
		failed = append(failed, "module_name_must_be_"+AetraNominationPoolModuleName)
	}
	checks := []requirementCheck{
		{AetraNominationPoolTestFirstDepositSharePrice, evidence.FirstDepositSharePrice},
		{AetraNominationPoolTestSubsequentDepositSharePrice, evidence.SubsequentDepositSharePrice},
		{AetraNominationPoolTestRewardDistribution, evidence.RewardDistribution},
		{AetraNominationPoolTestCommissionDeduction, evidence.CommissionDeduction},
		{AetraNominationPoolTestPartialWithdrawal, evidence.PartialWithdrawal},
		{AetraNominationPoolTestFullWithdrawal, evidence.FullWithdrawal},
		{AetraNominationPoolTestSlashingPoolValidator, evidence.SlashingPoolValidator},
		{AetraNominationPoolTestJailedValidator, evidence.JailedValidator},
		{AetraNominationPoolTestRedelegationIfAllowed, evidence.RedelegationIfAllowed},
		{AetraNominationPoolTestOperatorAbuseAttempt, evidence.PoolOperatorAbuseAttempt},
		{AetraNominationPoolTestExportImportActiveUnbonding, evidence.ExportImportActiveUnbonding},
		{AetraNominationPoolTestRoundingDustHandling, evidence.RoundingDustHandling},
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	sort.Strings(failed)
	return AetraNominationPoolTestingReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func requiredAetraNominationPoolFields() []string {
	return []string{
		AetraNominationPoolFieldPoolID,
		AetraNominationPoolFieldOperatorAddress,
		AetraNominationPoolFieldValidatorAddress,
		AetraNominationPoolFieldTotalBonded,
		AetraNominationPoolFieldTotalShares,
		AetraNominationPoolFieldCommissionBps,
		AetraNominationPoolFieldStatus,
		AetraNominationPoolFieldCreatedHeight,
		AetraNominationPoolFieldUnbondingEntries,
	}
}

func requiredAetraNominationPoolDelegationFields() []string {
	return []string{
		AetraNominationPoolFieldDelegatorAddress,
		AetraNominationPoolFieldDelegationPoolID,
		AetraNominationPoolFieldShares,
		AetraNominationPoolFieldPrincipalEstimate,
		AetraNominationPoolFieldRewardsAccrued,
	}
}

func validateAetraNominationPoolCatalog(group string, actual []string, required []string) (int, []string) {
	failed := make([]string, 0)
	requiredSet := map[string]bool{}
	for _, item := range required {
		requiredSet[item] = true
	}
	seen := map[string]bool{}
	for _, item := range actual {
		if item == "" {
			failed = append(failed, group+".item_required")
			continue
		}
		if seen[item] {
			failed = append(failed, group+"."+item+":duplicate")
			continue
		}
		seen[item] = true
		if !requiredSet[item] {
			failed = append(failed, group+"."+item+":unexpected")
		}
	}
	passed := 0
	for _, item := range required {
		if seen[item] {
			passed++
		} else {
			failed = append(failed, group+"."+item+":missing")
		}
	}
	return passed, failed
}
