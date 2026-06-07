package params

import (
	"fmt"
	"sort"
)

const (
	AetraThreatModelModuleName = "aetra-threat-model"

	AetraThreatValidatorCartel = "several_validators_coordinate_censorship_or_governance_capture"

	AetraThreatControlValidatorSetTarget             = "100_300_validator_target"
	AetraThreatControlValidatorPowerCap              = "validator_power_cap"
	AetraThreatControlTopNMonitoring                 = "top_n_monitoring"
	AetraThreatControlCommissionFloor                = "commission_floor"
	AetraThreatControlIdentityTransparency           = "identity_transparency"
	AetraThreatControlGovernanceParticipationMetrics = "governance_participation_metrics"
	AetraThreatControlDelegationWarnings             = "delegation_warnings"

	AetraThreatSimulationTop10Concentration         = "top_10_concentration_simulation"
	AetraThreatSimulationSplitIdentityValidator     = "split_identity_validator_simulation"
	AetraThreatSimulationDelegationOverflow         = "delegation_overflow_simulation"
	AetraThreatSimulationGovernanceCaptureThreshold = "governance_capture_threshold_analysis"
)

type AetraValidatorCartelThreatEvidence struct {
	ModuleName string

	Threats     []string
	Controls    []string
	Simulations []string

	UsesObjectiveChainData      bool
	UsesEconomicSignals         bool
	AvoidsMandatoryValidatorKYC bool
	DoesNotHaltStakingOnWarning bool
}

type AetraValidatorCartelThreatReport struct {
	ModuleName string
	Required   int
	Passed     int
	Failed     []string
	Ready      bool
}

func DefaultAetraValidatorCartelThreatEvidence() AetraValidatorCartelThreatEvidence {
	return AetraValidatorCartelThreatEvidence{
		ModuleName: AetraThreatModelModuleName,
		Threats: []string{
			AetraThreatValidatorCartel,
		},
		Controls:    requiredAetraValidatorCartelControls(),
		Simulations: requiredAetraValidatorCartelSimulations(),

		UsesObjectiveChainData:      true,
		UsesEconomicSignals:         true,
		AvoidsMandatoryValidatorKYC: true,
		DoesNotHaltStakingOnWarning: true,
	}
}

func ValidateAetraValidatorCartelThreat(evidence AetraValidatorCartelThreatEvidence) error {
	report := BuildAetraValidatorCartelThreatReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra validator cartel threat model failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraValidatorCartelThreatReport(evidence AetraValidatorCartelThreatEvidence) AetraValidatorCartelThreatReport {
	failed := validateAetraThreatModelModuleName(evidence.ModuleName)
	passedThreats, failedThreats := validateAetraThreatModelCatalog("threats", evidence.Threats, []string{AetraThreatValidatorCartel})
	passedControls, failedControls := validateAetraThreatModelCatalog("controls", evidence.Controls, requiredAetraValidatorCartelControls())
	passedSimulations, failedSimulations := validateAetraThreatModelCatalog("simulations", evidence.Simulations, requiredAetraValidatorCartelSimulations())
	failed = append(failed, failedThreats...)
	failed = append(failed, failedControls...)
	failed = append(failed, failedSimulations...)

	passed := passedThreats + passedControls + passedSimulations
	for _, check := range []requirementCheck{
		{"uses_objective_chain_data", evidence.UsesObjectiveChainData},
		{"uses_economic_signals", evidence.UsesEconomicSignals},
		{"avoids_mandatory_validator_kyc", evidence.AvoidsMandatoryValidatorKYC},
		{"does_not_halt_staking_on_warning", evidence.DoesNotHaltStakingOnWarning},
	} {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}

	sort.Strings(failed)
	return AetraValidatorCartelThreatReport{
		ModuleName: evidence.ModuleName,
		Required:   16,
		Passed:     passed,
		Failed:     failed,
		Ready:      len(failed) == 0,
	}
}

func requiredAetraValidatorCartelControls() []string {
	return []string{
		AetraThreatControlValidatorSetTarget,
		AetraThreatControlValidatorPowerCap,
		AetraThreatControlTopNMonitoring,
		AetraThreatControlCommissionFloor,
		AetraThreatControlIdentityTransparency,
		AetraThreatControlGovernanceParticipationMetrics,
		AetraThreatControlDelegationWarnings,
	}
}

func requiredAetraValidatorCartelSimulations() []string {
	return []string{
		AetraThreatSimulationTop10Concentration,
		AetraThreatSimulationSplitIdentityValidator,
		AetraThreatSimulationDelegationOverflow,
		AetraThreatSimulationGovernanceCaptureThreshold,
	}
}

func validateAetraThreatModelModuleName(moduleName string) []string {
	if moduleName == "" {
		return []string{"module_name_required"}
	}
	if moduleName != AetraThreatModelModuleName {
		return []string{"module_name_must_be_" + AetraThreatModelModuleName}
	}
	return nil
}

func validateAetraThreatModelCatalog(group string, actual []string, required []string) (int, []string) {
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
