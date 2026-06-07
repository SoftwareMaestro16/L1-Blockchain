package params

import (
	"fmt"
	"sort"
)

const (
	SecurityRequirementDeterministicStateTransitions  = "deterministic_state_transitions"
	SecurityRequirementNoExternalConsensusCalls       = "no_non_deterministic_external_calls"
	SecurityRequirementNoWallClockStateTransitions    = "no_wall_clock_dependency_except_block_time"
	SecurityRequirementNoFloatingPointAccounting      = "no_floating_point_accounting"
	SecurityRequirementNoUnorderedMapStateEffects     = "no_unordered_map_iteration_affecting_state"
	SecurityRequirementDeterministicSerialization     = "deterministic_serialization"
	SecurityRequirementExportImportEqualityTests      = "export_import_equality_tests"
	SecurityRequirementAppHashStabilityTests          = "app_hash_stability_tests"
	SecurityRequirementNoUnboundedMint                = "no_unbounded_mint"
	SecurityRequirementNoUnauthorizedModuleMintBurn   = "no_unauthorized_module_account_mint_burn"
	SecurityRequirementSupplyInvariants               = "supply_invariants"
	SecurityRequirementFeeSplitInvariants             = "fee_split_invariants"
	SecurityRequirementDelegationShareInvariants      = "delegation_share_invariants"
	SecurityRequirementRewardDistributionInvariants   = "reward_distribution_invariants"
	SecurityRequirementSlashingNoStakeUnderflow       = "slashing_cannot_underflow_stake"
	SecurityRequirementJailedValidatorRewardExclusion = "jailed_validators_no_active_rewards"
	SecurityRequirementNoConsensusPanicOnInvalidInput = "no_consensus_panic_on_invalid_input"
	SecurityRequirementInvariantTestsRequired         = "invariant_tests_required"
)

type ConsensusSafetyRequirements struct {
	DeterministicStateTransitions            bool
	NoNonDeterministicExternalCalls          bool
	NoWallClockDependencyExceptBlockTime     bool
	NoFloatingPointAccounting                bool
	NoUnorderedMapIterationAffectingState    bool
	DeterministicSerialization               bool
	ExportImportEqualityTests                bool
	AppHashStabilityTests                    bool
	NoConsensusPanicOnInvalidInput           bool
	ConsensusProvidedBlockTimeOnly           bool
	CrossArchitectureDeterminismReview       bool
	DeterminismStaticGateRequired            bool
	ConsensusPathExternalFilesystemForbidden bool
}

type EconomicSafetyRequirements struct {
	NoUnboundedMint                        bool
	NoUnauthorizedModuleAccountMintBurn    bool
	SupplyInvariants                       bool
	FeeSplitInvariants                     bool
	DelegationShareInvariants              bool
	RewardDistributionInvariants           bool
	SlashingCannotUnderflowStake           bool
	JailedValidatorsCannotReceiveRewards   bool
	ModuleAccountAuthorityChecked          bool
	BurnAndMintEventsAfterBankSuccess      bool
	FixedPointRewardMath                   bool
	NegativeBalanceForbidden               bool
	DistributionSkipsInactiveOrJailedPower bool
	EconomicInvariantTestsRequired         bool
	ExportImportSupplyStabilityTests       bool
}

type ValidatorRewardEligibility struct {
	ValidatorAddress string
	ActiveSet        bool
	Jailed           bool
	Tombstoned       bool
	RewardNaet       int64
}

type SecurityRequirementsReport struct {
	ConsensusRequired int
	ConsensusPassed   int
	EconomicRequired  int
	EconomicPassed    int
	Failed            []string
	Passed            bool
}

func DefaultConsensusSafetyRequirements() ConsensusSafetyRequirements {
	return ConsensusSafetyRequirements{
		DeterministicStateTransitions:            true,
		NoNonDeterministicExternalCalls:          true,
		NoWallClockDependencyExceptBlockTime:     true,
		NoFloatingPointAccounting:                true,
		NoUnorderedMapIterationAffectingState:    true,
		DeterministicSerialization:               true,
		ExportImportEqualityTests:                true,
		AppHashStabilityTests:                    true,
		NoConsensusPanicOnInvalidInput:           true,
		ConsensusProvidedBlockTimeOnly:           true,
		CrossArchitectureDeterminismReview:       true,
		DeterminismStaticGateRequired:            true,
		ConsensusPathExternalFilesystemForbidden: true,
	}
}

func DefaultEconomicSafetyRequirements() EconomicSafetyRequirements {
	return EconomicSafetyRequirements{
		NoUnboundedMint:                        true,
		NoUnauthorizedModuleAccountMintBurn:    true,
		SupplyInvariants:                       true,
		FeeSplitInvariants:                     true,
		DelegationShareInvariants:              true,
		RewardDistributionInvariants:           true,
		SlashingCannotUnderflowStake:           true,
		JailedValidatorsCannotReceiveRewards:   true,
		ModuleAccountAuthorityChecked:          true,
		BurnAndMintEventsAfterBankSuccess:      true,
		FixedPointRewardMath:                   true,
		NegativeBalanceForbidden:               true,
		DistributionSkipsInactiveOrJailedPower: true,
		EconomicInvariantTestsRequired:         true,
		ExportImportSupplyStabilityTests:       true,
	}
}

func ValidateSecurityRequirements(consensus ConsensusSafetyRequirements, economic EconomicSafetyRequirements) error {
	report := BuildSecurityRequirementsReport(consensus, economic)
	if !report.Passed {
		return fmt.Errorf("security requirements failed: %v", report.Failed)
	}
	return nil
}

func BuildSecurityRequirementsReport(consensus ConsensusSafetyRequirements, economic EconomicSafetyRequirements) SecurityRequirementsReport {
	consensusChecks := []requirementCheck{
		{SecurityRequirementDeterministicStateTransitions, consensus.DeterministicStateTransitions},
		{SecurityRequirementNoExternalConsensusCalls, consensus.NoNonDeterministicExternalCalls},
		{SecurityRequirementNoWallClockStateTransitions, consensus.NoWallClockDependencyExceptBlockTime && consensus.ConsensusProvidedBlockTimeOnly},
		{SecurityRequirementNoFloatingPointAccounting, consensus.NoFloatingPointAccounting},
		{SecurityRequirementNoUnorderedMapStateEffects, consensus.NoUnorderedMapIterationAffectingState},
		{SecurityRequirementDeterministicSerialization, consensus.DeterministicSerialization},
		{SecurityRequirementExportImportEqualityTests, consensus.ExportImportEqualityTests},
		{SecurityRequirementAppHashStabilityTests, consensus.AppHashStabilityTests},
		{SecurityRequirementNoConsensusPanicOnInvalidInput, consensus.NoConsensusPanicOnInvalidInput},
		{"cross_architecture_determinism_review", consensus.CrossArchitectureDeterminismReview},
		{"determinism_static_gate_required", consensus.DeterminismStaticGateRequired},
		{"consensus_path_external_filesystem_forbidden", consensus.ConsensusPathExternalFilesystemForbidden},
	}
	economicChecks := []requirementCheck{
		{SecurityRequirementNoUnboundedMint, economic.NoUnboundedMint},
		{SecurityRequirementNoUnauthorizedModuleMintBurn, economic.NoUnauthorizedModuleAccountMintBurn && economic.ModuleAccountAuthorityChecked},
		{SecurityRequirementSupplyInvariants, economic.SupplyInvariants && economic.ExportImportSupplyStabilityTests},
		{SecurityRequirementFeeSplitInvariants, economic.FeeSplitInvariants},
		{SecurityRequirementDelegationShareInvariants, economic.DelegationShareInvariants},
		{SecurityRequirementRewardDistributionInvariants, economic.RewardDistributionInvariants && economic.FixedPointRewardMath},
		{SecurityRequirementSlashingNoStakeUnderflow, economic.SlashingCannotUnderflowStake && economic.NegativeBalanceForbidden},
		{SecurityRequirementJailedValidatorRewardExclusion, economic.JailedValidatorsCannotReceiveRewards && economic.DistributionSkipsInactiveOrJailedPower},
		{"burn_and_mint_events_after_bank_success", economic.BurnAndMintEventsAfterBankSuccess},
		{SecurityRequirementInvariantTestsRequired, economic.EconomicInvariantTestsRequired},
	}
	failed := make([]string, 0)
	consensusPassed := 0
	economicPassed := 0
	for _, check := range consensusChecks {
		if check.Passed {
			consensusPassed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	for _, check := range economicChecks {
		if check.Passed {
			economicPassed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	sort.Strings(failed)
	return SecurityRequirementsReport{
		ConsensusRequired: len(consensusChecks),
		ConsensusPassed:   consensusPassed,
		EconomicRequired:  len(economicChecks),
		EconomicPassed:    economicPassed,
		Failed:            failed,
		Passed:            len(failed) == 0,
	}
}

func ValidateSlashingDoesNotUnderflowStake(stakeNaet, slashNaet int64) error {
	if stakeNaet < 0 || slashNaet < 0 {
		return fmt.Errorf("stake and slash amounts must be non-negative")
	}
	if slashNaet > stakeNaet {
		return fmt.Errorf("slashing cannot underflow stake")
	}
	return nil
}

func ValidateActiveValidatorRewardEligibility(eligibility ValidatorRewardEligibility) error {
	if eligibility.ValidatorAddress == "" {
		return fmt.Errorf("validator address is required")
	}
	if eligibility.RewardNaet < 0 {
		return fmt.Errorf("active validator reward cannot be negative")
	}
	if eligibility.RewardNaet == 0 {
		return nil
	}
	if eligibility.Jailed {
		return fmt.Errorf("jailed validators cannot receive active validator rewards")
	}
	if eligibility.Tombstoned {
		return fmt.Errorf("tombstoned validators cannot receive active validator rewards")
	}
	if !eligibility.ActiveSet {
		return fmt.Errorf("inactive validators cannot receive active validator rewards")
	}
	return nil
}

type requirementCheck struct {
	ID     string
	Passed bool
}
