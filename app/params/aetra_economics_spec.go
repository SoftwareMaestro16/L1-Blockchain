package params

import (
	"fmt"
	"sort"
)

const (
	AetraEconomicsModuleName = "x/aetra-economics"

	AetraEconomicsPurposeLowModerateInflation = "low_moderate_inflation"
	AetraEconomicsPurposeFeeBurn              = "fee_burn"
	AetraEconomicsPurposeTreasuryAllocation   = "treasury_allocation"
	AetraEconomicsPurposeRewardSmoothing      = "reward_smoothing"
	AetraEconomicsPurposeTransparentAPRModel  = "transparent_apr_model"
)

type AetraEconomicsSpecEvidence struct {
	ModuleName string

	LowModerateInflation bool
	FeeBurn              bool
	TreasuryAllocation   bool
	RewardSmoothing      bool
	TransparentAPRModel  bool
}

type AetraEconomicsSpecReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraEconomicsSpecEvidence() AetraEconomicsSpecEvidence {
	return AetraEconomicsSpecEvidence{
		ModuleName: AetraEconomicsModuleName,

		LowModerateInflation: true,
		FeeBurn:              true,
		TreasuryAllocation:   true,
		RewardSmoothing:      true,
		TransparentAPRModel:  true,
	}
}

func ValidateAetraEconomicsSpec(evidence AetraEconomicsSpecEvidence) error {
	report := BuildAetraEconomicsSpecReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra economics spec failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEconomicsSpecReport(evidence AetraEconomicsSpecEvidence) AetraEconomicsSpecReport {
	failed := make([]string, 0)
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
	} else if evidence.ModuleName != AetraEconomicsModuleName {
		failed = append(failed, "module_name_must_be_"+AetraEconomicsModuleName)
	}

	checks := []requirementCheck{
		{AetraEconomicsPurposeLowModerateInflation, evidence.LowModerateInflation},
		{AetraEconomicsPurposeFeeBurn, evidence.FeeBurn},
		{AetraEconomicsPurposeTreasuryAllocation, evidence.TreasuryAllocation},
		{AetraEconomicsPurposeRewardSmoothing, evidence.RewardSmoothing},
		{AetraEconomicsPurposeTransparentAPRModel, evidence.TransparentAPRModel},
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
	return AetraEconomicsSpecReport{
		ModuleName: evidence.ModuleName,
		Required:   len(checks),
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}
