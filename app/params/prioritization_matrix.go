package params

import (
	"fmt"
	"sort"
)

const (
	PriorityUrgencyLow	= "Low"
	PriorityUrgencyMedium	= "Medium"
	PriorityUrgencyHigh	= "High"

	PriorityWaveCritical	= "wave_0_critical"
	PriorityWaveHigh	= "wave_1_high"
	PriorityWaveMedium	= "wave_2_medium"
	PriorityWaveBacklog	= "wave_3_backlog"
)

type EconomicPriorityItem struct {
	Improvement			string
	SecurityImpact			int64
	DecentralizationImpact		int64
	ImplementationComplexity	int64
	Urgency				string
}

type EconomicPriorityParams struct {
	SecurityWeight			int64
	DecentralizationWeight		int64
	UrgencyWeight			int64
	ComplexityPenaltyWeight		int64
	CriticalPriorityScore		int64
	HighPriorityScore		int64
	MediumPriorityScore		int64
	RequireKnownMatrixComplete	bool
}

type RankedEconomicPriority struct {
	EconomicPriorityItem
	UrgencyScore		int64
	PriorityScore		int64
	ExecutionWave		string
	GovernanceRationale	string
}

type EconomicPriorityWaveSummary struct {
	Wave		string
	Count		int
	Items		[]string
	ScoreMin	int64
	ScoreMax	int64
}

type EconomicPriorityMatrixReport struct {
	Ranked			[]RankedEconomicPriority
	Waves			[]EconomicPriorityWaveSummary
	HighestPriority		RankedEconomicPriority
	HighUrgencyCount	int
	AverageComplexity	int64
	GovernanceSummary	string
	Passed			bool
	Failed			[]string
}

func DefaultEconomicPriorityParams() EconomicPriorityParams {
	return EconomicPriorityParams{
		SecurityWeight:			400,
		DecentralizationWeight:		300,
		UrgencyWeight:			200,
		ComplexityPenaltyWeight:	100,
		CriticalPriorityScore:		7_000,
		HighPriorityScore:		5_500,
		MediumPriorityScore:		4_000,
		RequireKnownMatrixComplete:	true,
	}
}

func DefaultEconomicPrioritizationMatrix() []EconomicPriorityItem {
	return []EconomicPriorityItem{
		{Improvement: "Wire burn controller into production fee flow", SecurityImpact: 8, DecentralizationImpact: 4, ImplementationComplexity: 6, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add net issuance accounting and supply reports", SecurityImpact: 8, DecentralizationImpact: 3, ImplementationComplexity: 4, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add deflation guard enforcement", SecurityImpact: 8, DecentralizationImpact: 3, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
		{Improvement: "Productionize epoch-based validator selection", SecurityImpact: 9, DecentralizationImpact: 8, ImplementationComplexity: 8, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add validator concentration metrics", SecurityImpact: 8, DecentralizationImpact: 9, ImplementationComplexity: 4, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add reward dampening above concentration thresholds", SecurityImpact: 8, DecentralizationImpact: 9, ImplementationComplexity: 7, Urgency: PriorityUrgencyHigh},
		{Improvement: "Extend slashing fund routing to burn, treasury, and reporters", SecurityImpact: 9, DecentralizationImpact: 5, ImplementationComplexity: 7, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add repeat-offense slashing multipliers", SecurityImpact: 8, DecentralizationImpact: 5, ImplementationComplexity: 6, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add validator risk and reputation queries", SecurityImpact: 6, DecentralizationImpact: 8, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add risk-adjusted delegation yield estimates", SecurityImpact: 6, DecentralizationImpact: 8, ImplementationComplexity: 5, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add validator metadata and commission-change warnings", SecurityImpact: 6, DecentralizationImpact: 7, ImplementationComplexity: 4, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add dynamic base fee adjustment bounds", SecurityImpact: 8, DecentralizationImpact: 4, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add congestion simulations and fee controller tests", SecurityImpact: 8, DecentralizationImpact: 3, ImplementationComplexity: 4, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add sender-local anti-spam surcharge", SecurityImpact: 8, DecentralizationImpact: 3, ImplementationComplexity: 6, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add resource-specific fee multipliers", SecurityImpact: 7, DecentralizationImpact: 3, ImplementationComplexity: 7, Urgency: PriorityUrgencyMedium},
		{Improvement: "Replace static fee split with bucketed allocation", SecurityImpact: 7, DecentralizationImpact: 5, ImplementationComplexity: 6, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add state write and update pricing", SecurityImpact: 8, DecentralizationImpact: 4, ImplementationComplexity: 6, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add storage footprint queries", SecurityImpact: 6, DecentralizationImpact: 4, ImplementationComplexity: 4, Urgency: PriorityUrgencyMedium},
		{Improvement: "Design and implement state rent", SecurityImpact: 9, DecentralizationImpact: 5, ImplementationComplexity: 9, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add state delete refund policy", SecurityImpact: 7, DecentralizationImpact: 4, ImplementationComplexity: 6, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add state growth telemetry and alerts", SecurityImpact: 7, DecentralizationImpact: 4, ImplementationComplexity: 4, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add adaptive inflation controller", SecurityImpact: 9, DecentralizationImpact: 5, ImplementationComplexity: 8, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add inflation smoothing and per-window limits", SecurityImpact: 8, DecentralizationImpact: 4, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add economic invariant tests", SecurityImpact: 9, DecentralizationImpact: 5, ImplementationComplexity: 5, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add economic attack simulations", SecurityImpact: 9, DecentralizationImpact: 6, ImplementationComplexity: 6, Urgency: PriorityUrgencyHigh},
		{Improvement: "Add security reserve accounting", SecurityImpact: 7, DecentralizationImpact: 4, ImplementationComplexity: 6, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add fee market circuit breaker", SecurityImpact: 8, DecentralizationImpact: 3, ImplementationComplexity: 7, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add delegation simulator", SecurityImpact: 5, DecentralizationImpact: 7, ImplementationComplexity: 5, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add validator bootstrap band", SecurityImpact: 6, DecentralizationImpact: 8, ImplementationComplexity: 7, Urgency: PriorityUrgencyMedium},
		{Improvement: "Add governance parameter impact reports", SecurityImpact: 7, DecentralizationImpact: 5, ImplementationComplexity: 6, Urgency: PriorityUrgencyMedium},
	}
}

func BuildEconomicPriorityMatrixReport(items []EconomicPriorityItem, params EconomicPriorityParams) (EconomicPriorityMatrixReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return EconomicPriorityMatrixReport{}, err
	}
	if len(items) == 0 {
		items = DefaultEconomicPrioritizationMatrix()
	}
	failed := validatePriorityItems(items, params)
	ranked := make([]RankedEconomicPriority, 0, len(items))
	totalComplexity := int64(0)
	highUrgency := 0
	for _, item := range items {
		urgencyScore := priorityUrgencyScore(item.Urgency)
		score := item.SecurityImpact*params.SecurityWeight +
			item.DecentralizationImpact*params.DecentralizationWeight +
			urgencyScore*params.UrgencyWeight -
			item.ImplementationComplexity*params.ComplexityPenaltyWeight
		ranked = append(ranked, RankedEconomicPriority{
			EconomicPriorityItem:	item,
			UrgencyScore:		urgencyScore,
			PriorityScore:		score,
			ExecutionWave:		priorityExecutionWave(score, params),
			GovernanceRationale:	priorityRationale(item, score),
		})
		totalComplexity += item.ImplementationComplexity
		if item.Urgency == PriorityUrgencyHigh {
			highUrgency++
		}
	}
	sortRankedEconomicPriorities(ranked)
	waves := summarizePriorityWaves(ranked)
	highest := RankedEconomicPriority{}
	if len(ranked) > 0 {
		highest = ranked[0]
	}
	averageComplexity := int64(0)
	if len(items) > 0 {
		averageComplexity = totalComplexity / int64(len(items))
	}
	return EconomicPriorityMatrixReport{
		Ranked:			ranked,
		Waves:			waves,
		HighestPriority:	highest,
		HighUrgencyCount:	highUrgency,
		AverageComplexity:	averageComplexity,
		GovernanceSummary:	fmt.Sprintf("items=%d high_urgency=%d top=%q top_score=%d waves=%d", len(items), highUrgency, highest.Improvement, highest.PriorityScore, len(waves)),
		Passed:			len(failed) == 0,
		Failed:			failed,
	}, nil
}

func (p EconomicPriorityParams) Validate() error {
	for _, field := range []struct {
		name	string
		value	int64
	}{
		{name: "security_weight", value: p.SecurityWeight},
		{name: "decentralization_weight", value: p.DecentralizationWeight},
		{name: "urgency_weight", value: p.UrgencyWeight},
		{name: "complexity_penalty_weight", value: p.ComplexityPenaltyWeight},
		{name: "critical_priority_score", value: p.CriticalPriorityScore},
		{name: "high_priority_score", value: p.HighPriorityScore},
		{name: "medium_priority_score", value: p.MediumPriorityScore},
	} {
		if field.value < 0 {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	if p.CriticalPriorityScore < p.HighPriorityScore || p.HighPriorityScore < p.MediumPriorityScore {
		return fmt.Errorf("priority score thresholds must be descending")
	}
	return nil
}

func (p EconomicPriorityParams) withDefaults() EconomicPriorityParams {
	defaults := DefaultEconomicPriorityParams()
	if p.SecurityWeight == 0 {
		p.SecurityWeight = defaults.SecurityWeight
	}
	if p.DecentralizationWeight == 0 {
		p.DecentralizationWeight = defaults.DecentralizationWeight
	}
	if p.UrgencyWeight == 0 {
		p.UrgencyWeight = defaults.UrgencyWeight
	}
	if p.ComplexityPenaltyWeight == 0 {
		p.ComplexityPenaltyWeight = defaults.ComplexityPenaltyWeight
	}
	if p.CriticalPriorityScore == 0 {
		p.CriticalPriorityScore = defaults.CriticalPriorityScore
	}
	if p.HighPriorityScore == 0 {
		p.HighPriorityScore = defaults.HighPriorityScore
	}
	if p.MediumPriorityScore == 0 {
		p.MediumPriorityScore = defaults.MediumPriorityScore
	}
	return p
}

func validatePriorityItems(items []EconomicPriorityItem, params EconomicPriorityParams) []string {
	failed := make([]string, 0)
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if item.Improvement == "" {
			failed = append(failed, "improvement_missing")
		}
		if seen[item.Improvement] {
			failed = append(failed, "duplicate_improvement:"+item.Improvement)
		}
		seen[item.Improvement] = true
		for _, field := range []struct {
			name	string
			value	int64
		}{
			{name: "security_impact", value: item.SecurityImpact},
			{name: "decentralization_impact", value: item.DecentralizationImpact},
			{name: "implementation_complexity", value: item.ImplementationComplexity},
		} {
			if field.value < 1 || field.value > 10 {
				failed = append(failed, item.Improvement+":"+field.name+"_out_of_range")
			}
		}
		if priorityUrgencyScore(item.Urgency) == 0 {
			failed = append(failed, item.Improvement+":urgency_invalid")
		}
	}
	if params.RequireKnownMatrixComplete {
		defaults := DefaultEconomicPrioritizationMatrix()
		if len(items) != len(defaults) {
			failed = append(failed, "prioritization_matrix_size_mismatch")
		}
		for _, item := range defaults {
			if !seen[item.Improvement] {
				failed = append(failed, "missing_default_improvement:"+item.Improvement)
			}
		}
	}
	return uniqueStrings(failed)
}

func sortRankedEconomicPriorities(items []RankedEconomicPriority) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].PriorityScore != items[j].PriorityScore {
			return items[i].PriorityScore > items[j].PriorityScore
		}
		if items[i].UrgencyScore != items[j].UrgencyScore {
			return items[i].UrgencyScore > items[j].UrgencyScore
		}
		if items[i].SecurityImpact != items[j].SecurityImpact {
			return items[i].SecurityImpact > items[j].SecurityImpact
		}
		if items[i].DecentralizationImpact != items[j].DecentralizationImpact {
			return items[i].DecentralizationImpact > items[j].DecentralizationImpact
		}
		if items[i].ImplementationComplexity != items[j].ImplementationComplexity {
			return items[i].ImplementationComplexity < items[j].ImplementationComplexity
		}
		return items[i].Improvement < items[j].Improvement
	})
}

func summarizePriorityWaves(items []RankedEconomicPriority) []EconomicPriorityWaveSummary {
	byWave := make(map[string][]RankedEconomicPriority)
	for _, item := range items {
		byWave[item.ExecutionWave] = append(byWave[item.ExecutionWave], item)
	}
	order := []string{PriorityWaveCritical, PriorityWaveHigh, PriorityWaveMedium, PriorityWaveBacklog}
	out := make([]EconomicPriorityWaveSummary, 0, len(byWave))
	for _, wave := range order {
		group := byWave[wave]
		if len(group) == 0 {
			continue
		}
		names := make([]string, 0, len(group))
		minScore := group[0].PriorityScore
		maxScore := group[0].PriorityScore
		for _, item := range group {
			names = append(names, item.Improvement)
			if item.PriorityScore < minScore {
				minScore = item.PriorityScore
			}
			if item.PriorityScore > maxScore {
				maxScore = item.PriorityScore
			}
		}
		out = append(out, EconomicPriorityWaveSummary{Wave: wave, Count: len(group), Items: names, ScoreMin: minScore, ScoreMax: maxScore})
	}
	return out
}

func priorityUrgencyScore(urgency string) int64 {
	switch urgency {
	case PriorityUrgencyHigh:
		return 10
	case PriorityUrgencyMedium:
		return 6
	case PriorityUrgencyLow:
		return 3
	default:
		return 0
	}
}

func priorityExecutionWave(score int64, params EconomicPriorityParams) string {
	switch {
	case score >= params.CriticalPriorityScore:
		return PriorityWaveCritical
	case score >= params.HighPriorityScore:
		return PriorityWaveHigh
	case score >= params.MediumPriorityScore:
		return PriorityWaveMedium
	default:
		return PriorityWaveBacklog
	}
}

func priorityRationale(item EconomicPriorityItem, score int64) string {
	return fmt.Sprintf("security=%d decentralization=%d complexity=%d urgency=%s priority_score=%d", item.SecurityImpact, item.DecentralizationImpact, item.ImplementationComplexity, item.Urgency, score)
}
