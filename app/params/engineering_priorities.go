package params

import (
	"fmt"
	"sort"
)

const (
	AetraEngineeringPriorityP0	= "P0"
	AetraEngineeringPriorityP1	= "P1"
	AetraEngineeringPriorityP2	= "P2"
	AetraEngineeringPriorityP3	= "P3"
)

const (
	AetraEngineeringP0ConsensusSafety	= "consensus_safety"
	AetraEngineeringP0DeterministicState	= "deterministic_state"
	AetraEngineeringP0StakingCorrectness	= "staking_correctness"
	AetraEngineeringP0SlashingCorrectness	= "slashing_correctness"
	AetraEngineeringP0SupplyInvariants	= "supply_invariants"
	AetraEngineeringP0ExportImport		= "export_import"
)

const (
	AetraEngineeringP1ValidatorPowerCap	= "validator_power_cap"
	AetraEngineeringP1FeeBurnEconomics	= "fee_burn_economics"
	AetraEngineeringP1ValidatorScore	= "validator_score"
	AetraEngineeringP1NominationPoolSafety	= "nomination_pool_safety"
	AetraEngineeringP1GovernanceBounds	= "governance_bounds"
)

const (
	AetraEngineeringP2AVMHardening		= "avm_production_hardening"
	AetraEngineeringP2Observability		= "observability"
	AetraEngineeringP2Dashboards		= "dashboards"
	AetraEngineeringP2LoadTests		= "load_tests"
	AetraEngineeringP2PublicTestnetDocs	= "public_testnet_docs"
)

const (
	AetraEngineeringP3AntiCartelAnalytics		= "advanced_anti_cartel_analytics"
	AetraEngineeringP3AVMLanguageResearch		= "avm_language_research"
	AetraEngineeringP3MEVPolicy			= "mev_policy"
	AetraEngineeringP3EncryptedMempoolResearch	= "encrypted_mempool_research"
	AetraEngineeringP3HigherValidatorCapExperiments	= "higher_validator_cap_experiments"
)

type AetraEngineeringPriorityEvidence struct {
	Priority	string
	Items		[]string
	Stable		bool
}

type AetraEngineeringPrioritiesReport struct {
	Priorities	[]AetraEngineeringPriorityEvidence
	Required	int
	Passed		int
	Failed		[]string
	P3Allowed	bool
	Ready		bool
}

func DefaultAetraEngineeringPrioritiesEvidence() []AetraEngineeringPriorityEvidence {
	return []AetraEngineeringPriorityEvidence{
		{Priority: AetraEngineeringPriorityP0, Items: RequiredAetraEngineeringP0Items(), Stable: true},
		{Priority: AetraEngineeringPriorityP1, Items: RequiredAetraEngineeringP1Items(), Stable: true},
		{Priority: AetraEngineeringPriorityP2, Items: RequiredAetraEngineeringP2Items(), Stable: false},
		{Priority: AetraEngineeringPriorityP3, Items: RequiredAetraEngineeringP3Items(), Stable: false},
	}
}

func ValidateAetraEngineeringPriorities(evidence []AetraEngineeringPriorityEvidence) error {
	report := BuildAetraEngineeringPrioritiesReport(evidence)
	if !report.Ready {
		return fmt.Errorf("aetra engineering priorities failed: %v", report.Failed)
	}
	return nil
}

func BuildAetraEngineeringPrioritiesReport(evidence []AetraEngineeringPriorityEvidence) AetraEngineeringPrioritiesReport {
	if evidence == nil {
		evidence = DefaultAetraEngineeringPrioritiesEvidence()
	}
	evidence = normalizeEngineeringPriorities(evidence)
	requiredPriorities := requiredEngineeringPriorities()
	seen := map[string]AetraEngineeringPriorityEvidence{}
	failed := make([]string, 0)
	required := 0
	passed := 0
	p0Stable := false
	p1Stable := false
	p3Present := false

	for _, priority := range evidence {
		if priority.Priority == "" {
			failed = append(failed, "priority_required")
			continue
		}
		if _, duplicate := seen[priority.Priority]; duplicate {
			failed = append(failed, priority.Priority+":duplicate_priority")
		}
		seen[priority.Priority] = priority
		items, known := requiredPriorities[priority.Priority]
		if !known {
			failed = append(failed, priority.Priority+":unknown_priority")
			continue
		}
		required += len(items)
		passedItems, failedItems := validateEngineeringPriorityCatalog(priority.Priority, priority.Items, items)
		passed += passedItems
		failed = append(failed, failedItems...)
		switch priority.Priority {
		case AetraEngineeringPriorityP0:
			p0Stable = priority.Stable
		case AetraEngineeringPriorityP1:
			p1Stable = priority.Stable
		case AetraEngineeringPriorityP3:
			p3Present = true
		}
	}
	for priority := range requiredPriorities {
		if _, ok := seen[priority]; !ok {
			failed = append(failed, priority+":missing_priority")
			required += len(requiredPriorities[priority])
		}
	}

	p3Allowed := p0Stable && p1Stable
	if p3Present && !p3Allowed {
		failed = append(failed, "p3_requires_p0_and_p1_stable")
	}

	sort.Strings(failed)
	return AetraEngineeringPrioritiesReport{
		Priorities:	evidence,
		Required:	required,
		Passed:		passed,
		Failed:		failed,
		P3Allowed:	p3Allowed,
		Ready:		len(failed) == 0,
	}
}

func RequiredAetraEngineeringP0Items() []string {
	return []string{
		AetraEngineeringP0ConsensusSafety,
		AetraEngineeringP0DeterministicState,
		AetraEngineeringP0StakingCorrectness,
		AetraEngineeringP0SlashingCorrectness,
		AetraEngineeringP0SupplyInvariants,
		AetraEngineeringP0ExportImport,
	}
}

func RequiredAetraEngineeringP1Items() []string {
	return []string{
		AetraEngineeringP1ValidatorPowerCap,
		AetraEngineeringP1FeeBurnEconomics,
		AetraEngineeringP1ValidatorScore,
		AetraEngineeringP1NominationPoolSafety,
		AetraEngineeringP1GovernanceBounds,
	}
}

func RequiredAetraEngineeringP2Items() []string {
	return []string{
		AetraEngineeringP2AVMHardening,
		AetraEngineeringP2Observability,
		AetraEngineeringP2Dashboards,
		AetraEngineeringP2LoadTests,
		AetraEngineeringP2PublicTestnetDocs,
	}
}

func RequiredAetraEngineeringP3Items() []string {
	return []string{
		AetraEngineeringP3AntiCartelAnalytics,
		AetraEngineeringP3AVMLanguageResearch,
		AetraEngineeringP3MEVPolicy,
		AetraEngineeringP3EncryptedMempoolResearch,
		AetraEngineeringP3HigherValidatorCapExperiments,
	}
}

func requiredEngineeringPriorities() map[string][]string {
	return map[string][]string{
		AetraEngineeringPriorityP0:	RequiredAetraEngineeringP0Items(),
		AetraEngineeringPriorityP1:	RequiredAetraEngineeringP1Items(),
		AetraEngineeringPriorityP2:	RequiredAetraEngineeringP2Items(),
		AetraEngineeringPriorityP3:	RequiredAetraEngineeringP3Items(),
	}
}

func validateEngineeringPriorityCatalog(priority string, actual []string, required []string) (int, []string) {
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
			failed = append(failed, priority+"."+item+":missing")
		case 1:
			passed++
		default:
			failed = append(failed, priority+"."+item+":duplicate")
		}
	}
	for item := range actualCounts {
		if !requiredSet[item] {
			failed = append(failed, priority+"."+item+":unexpected")
		}
	}
	return passed, failed
}

func normalizeEngineeringPriorities(priorities []AetraEngineeringPriorityEvidence) []AetraEngineeringPriorityEvidence {
	out := append([]AetraEngineeringPriorityEvidence{}, priorities...)
	sort.SliceStable(out, func(i, j int) bool {
		return engineeringPriorityRank(out[i].Priority) < engineeringPriorityRank(out[j].Priority)
	})
	return out
}

func engineeringPriorityRank(priority string) int {
	switch priority {
	case AetraEngineeringPriorityP0:
		return 0
	case AetraEngineeringPriorityP1:
		return 1
	case AetraEngineeringPriorityP2:
		return 2
	case AetraEngineeringPriorityP3:
		return 3
	default:
		return 99
	}
}
