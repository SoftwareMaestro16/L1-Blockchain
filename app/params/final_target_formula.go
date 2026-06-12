package params

import (
	"fmt"
	"sort"
)

const (
	FinalTargetConsensusCometBFTBFTPoS	= "cometbft_bft_pos"
	FinalTargetCosmosSDK			= "cosmos_sdk"
	FinalTargetAVMOnlyGenesis		= "avm_only_genesis"
	FinalTargetValidatorSetRange		= "100_300_active_validators_over_time"
	FinalTargetBlockTimeRange		= "5_8_second_block_time"
	FinalTargetWorstFinality		= "120_second_worst_acceptable_finality"
	FinalTargetObjectiveSlashing		= "strict_objective_slashing"
	FinalTargetValidatorPowerCap		= "validator_effective_power_cap"
	FinalTargetAntiConcentrationRewards	= "anti_concentration_rewards"
	FinalTargetDynamicLowModerateInflation	= "dynamic_low_moderate_inflation"
	FinalTargetFeeBurn			= "fee_burn"
	FinalTargetProtocolTreasury		= "protocol_treasury"
	FinalTargetMandatoryFeatureTests	= "mandatory_tests_for_every_feature"
	FinalTargetTrustProductDecision		= "trust_over_speed_or_short_term_apr"

	FinalTargetMinActiveValidators	= 100
	FinalTargetMaxActiveValidators	= 300
	FinalTargetMinBlockTimeSeconds	= 5
	FinalTargetMaxBlockTimeSeconds	= 8
	FinalTargetWorstFinalitySeconds	= 120
)

type AetraFinalTargetFormula struct {
	CometBFTBFTPoS			bool
	CosmosSDK			bool
	AVMOnlyGenesis			bool
	MinActiveValidators		int
	MaxActiveValidators		int
	MinBlockTimeSeconds		int
	MaxBlockTimeSeconds		int
	WorstAcceptableFinalitySeconds	int
	StrictObjectiveSlashing		bool
	ValidatorEffectivePowerCap	bool
	AntiConcentrationRewards	bool
	DynamicLowModerateInflation	bool
	FeeBurn				bool
	ProtocolTreasury		bool
	MandatoryTestsForEveryFeature	bool
	TrustOverSpeedOrShortTermAPR	bool
}

type FinalTargetFormulaReport struct {
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultAetraFinalTargetFormula() AetraFinalTargetFormula {
	return AetraFinalTargetFormula{
		CometBFTBFTPoS:			true,
		CosmosSDK:			true,
		AVMOnlyGenesis:			true,
		MinActiveValidators:		FinalTargetMinActiveValidators,
		MaxActiveValidators:		FinalTargetMaxActiveValidators,
		MinBlockTimeSeconds:		FinalTargetMinBlockTimeSeconds,
		MaxBlockTimeSeconds:		FinalTargetMaxBlockTimeSeconds,
		WorstAcceptableFinalitySeconds:	FinalTargetWorstFinalitySeconds,
		StrictObjectiveSlashing:	true,
		ValidatorEffectivePowerCap:	true,
		AntiConcentrationRewards:	true,
		DynamicLowModerateInflation:	true,
		FeeBurn:			true,
		ProtocolTreasury:		true,
		MandatoryTestsForEveryFeature:	true,
		TrustOverSpeedOrShortTermAPR:	true,
	}
}

func ValidateAetraFinalTargetFormula(formula AetraFinalTargetFormula) error {
	report := BuildFinalTargetFormulaReport(formula)
	if !report.Ready {
		return fmt.Errorf("final target formula failed: %v", report.Failed)
	}
	return nil
}

func BuildFinalTargetFormulaReport(formula AetraFinalTargetFormula) FinalTargetFormulaReport {
	checks := []requirementCheck{
		{FinalTargetConsensusCometBFTBFTPoS, formula.CometBFTBFTPoS},
		{FinalTargetCosmosSDK, formula.CosmosSDK},
		{FinalTargetAVMOnlyGenesis, formula.AVMOnlyGenesis},
		{FinalTargetValidatorSetRange, formula.MinActiveValidators == FinalTargetMinActiveValidators && formula.MaxActiveValidators == FinalTargetMaxActiveValidators},
		{FinalTargetBlockTimeRange, formula.MinBlockTimeSeconds == FinalTargetMinBlockTimeSeconds && formula.MaxBlockTimeSeconds == FinalTargetMaxBlockTimeSeconds},
		{FinalTargetWorstFinality, formula.WorstAcceptableFinalitySeconds <= FinalTargetWorstFinalitySeconds && formula.WorstAcceptableFinalitySeconds > 0},
		{FinalTargetObjectiveSlashing, formula.StrictObjectiveSlashing},
		{FinalTargetValidatorPowerCap, formula.ValidatorEffectivePowerCap},
		{FinalTargetAntiConcentrationRewards, formula.AntiConcentrationRewards},
		{FinalTargetDynamicLowModerateInflation, formula.DynamicLowModerateInflation},
		{FinalTargetFeeBurn, formula.FeeBurn},
		{FinalTargetProtocolTreasury, formula.ProtocolTreasury},
		{FinalTargetMandatoryFeatureTests, formula.MandatoryTestsForEveryFeature},
		{FinalTargetTrustProductDecision, formula.TrustOverSpeedOrShortTermAPR},
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
	return FinalTargetFormulaReport{
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}
