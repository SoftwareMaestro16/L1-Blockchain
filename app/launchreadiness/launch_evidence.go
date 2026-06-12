package launchreadiness

import (
	"errors"
	"fmt"
	"sort"
)

const (
	LaunchEvidenceLocalnet100Validators	= "localnet_100_validators"
	LaunchEvidenceStressFinality		= "stress_finality"
	LaunchEvidenceRestartPersistence	= "restart_persistence"
	LaunchEvidenceSnapshotStateSync		= "snapshot_state_sync"
	LaunchEvidenceAVMStateGrowth		= "avm_state_growth"

	DeterministicSensitiveMatchCount	= 47
	PanicMustMatchCount			= 284
)

type LaunchEvidenceReport struct {
	Evidence []LaunchEvidenceItem `json:"evidence"`
}

type LaunchEvidenceItem struct {
	Kind				string	`json:"kind"`
	ArtifactPath			string	`json:"artifact_path"`
	ValidatorCount			uint32	`json:"validator_count"`
	ObservedBlocks			uint64	`json:"observed_blocks"`
	NormalFinalityP95Seconds	uint32	`json:"normal_finality_p95_seconds"`
	StressFinalityP95Seconds	uint32	`json:"stress_finality_p95_seconds"`
	WorstFinalitySeconds		uint32	`json:"worst_finality_seconds"`
	RestartHeightBefore		uint64	`json:"restart_height_before"`
	RestartHeightAfter		uint64	`json:"restart_height_after"`
	SnapshotHeight			uint64	`json:"snapshot_height"`
	StateSyncHeight			uint64	`json:"state_sync_height"`
	AVMStateGrowthBytes		uint64	`json:"avm_state_growth_bytes"`
	AVMStateGrowthLimitBytes	uint64	`json:"avm_state_growth_limit_bytes"`
	DeterministicReplayHash		string	`json:"deterministic_replay_hash"`
	OperatorSummaryIncluded		bool	`json:"operator_summary_included"`
	PrivateMaterialRedacted		bool	`json:"private_material_redacted"`
}

type PatternTriageReport struct {
	DeterministicSensitiveMatches	uint32			`json:"deterministic_sensitive_matches"`
	PanicMustMatches		uint32			`json:"panic_must_matches"`
	Findings			[]PatternTriageFinding	`json:"findings"`
}

type PatternTriageFinding struct {
	ID		string	`json:"id"`
	Pattern		string	`json:"pattern"`
	PatternKind	string	`json:"pattern_kind"`
	MatchCount	uint32	`json:"match_count"`
	File		string	`json:"file"`
	Category	string	`json:"category"`
	ConsensusPath	bool	`json:"consensus_path"`
	Generated	bool	`json:"generated"`
	TestOnly	bool	`json:"test_only"`
	Action		string	`json:"action"`
}

func ValidateLaunchEvidenceReport(report LaunchEvidenceReport) error {
	byKind := make(map[string]LaunchEvidenceItem, len(report.Evidence))
	for _, item := range report.Evidence {
		if item.Kind == "" {
			return errors.New("launch evidence kind is required")
		}
		if _, found := byKind[item.Kind]; found {
			return fmt.Errorf("duplicate launch evidence %s", item.Kind)
		}
		if err := item.validateCommon(); err != nil {
			return err
		}
		byKind[item.Kind] = item
	}
	for _, kind := range requiredLaunchEvidenceKinds() {
		item, found := byKind[kind]
		if !found {
			return fmt.Errorf("missing launch evidence %s", kind)
		}
		if err := item.validateKind(); err != nil {
			return err
		}
	}
	return nil
}

func ValidatePatternTriageReport(report PatternTriageReport) error {
	if report.DeterministicSensitiveMatches != DeterministicSensitiveMatchCount {
		return fmt.Errorf("deterministic-sensitive match count changed: got %d want %d", report.DeterministicSensitiveMatches, DeterministicSensitiveMatchCount)
	}
	if report.PanicMustMatches != PanicMustMatchCount {
		return fmt.Errorf("panic/Must match count changed: got %d want %d", report.PanicMustMatches, PanicMustMatchCount)
	}
	if len(report.Findings) == 0 {
		return errors.New("pattern triage findings are required")
	}
	seen := map[string]struct{}{}
	var deterministicCovered uint32
	var panicCovered uint32
	for _, finding := range report.Findings {
		if err := finding.Validate(); err != nil {
			return err
		}
		if _, found := seen[finding.ID]; found {
			return fmt.Errorf("duplicate pattern triage finding %s", finding.ID)
		}
		seen[finding.ID] = struct{}{}
		switch finding.PatternKind {
		case "deterministic_sensitive":
			deterministicCovered += finding.MatchCount
		case "panic_must":
			panicCovered += finding.MatchCount
		case "both":
			deterministicCovered += finding.MatchCount
			panicCovered += finding.MatchCount
		}
	}
	if deterministicCovered != report.DeterministicSensitiveMatches {
		return fmt.Errorf("deterministic-sensitive triage coverage mismatch: got %d want %d", deterministicCovered, report.DeterministicSensitiveMatches)
	}
	if panicCovered != report.PanicMustMatches {
		return fmt.Errorf("panic/Must triage coverage mismatch: got %d want %d", panicCovered, report.PanicMustMatches)
	}
	return nil
}

func (f PatternTriageFinding) Validate() error {
	if f.ID == "" || f.Pattern == "" || f.File == "" {
		return errors.New("pattern triage finding requires id, pattern, and file")
	}
	if f.MatchCount == 0 {
		return errors.New("pattern triage finding requires positive match count")
	}
	switch f.PatternKind {
	case "deterministic_sensitive", "panic_must", "both":
	default:
		return fmt.Errorf("unsupported pattern triage kind %s", f.PatternKind)
	}
	switch f.Category {
	case "generated_gateway_must", "sdk_module_init_export_panic", "test_only_must", "consensus_path_review_required", "determinism_guarded", "operator_startup_invariant":
	default:
		return fmt.Errorf("unsupported pattern triage category %s", f.Category)
	}
	switch f.Action {
	case "accept_generated", "document_startup_invariant", "test_only", "replace_before_mainnet", "requires_owner_review", "covered_by_determinism_test":
	default:
		return fmt.Errorf("unsupported pattern triage action %s", f.Action)
	}
	if f.ConsensusPath && (f.Generated || f.TestOnly) {
		return errors.New("consensus-path pattern cannot be classified as generated or test-only")
	}
	if f.ConsensusPath && f.Action != "replace_before_mainnet" && f.Action != "requires_owner_review" && f.Action != "covered_by_determinism_test" {
		return errors.New("consensus-path pattern requires mainnet review, replacement, or determinism coverage")
	}
	return nil
}

func (item LaunchEvidenceItem) validateCommon() error {
	if item.ArtifactPath == "" {
		return fmt.Errorf("launch evidence %s requires artifact path", item.Kind)
	}
	if item.ObservedBlocks == 0 {
		return fmt.Errorf("launch evidence %s requires observed blocks", item.Kind)
	}
	if item.DeterministicReplayHash == "" {
		return fmt.Errorf("launch evidence %s requires deterministic replay hash", item.Kind)
	}
	if !item.OperatorSummaryIncluded {
		return fmt.Errorf("launch evidence %s requires operator summary", item.Kind)
	}
	if !item.PrivateMaterialRedacted {
		return fmt.Errorf("launch evidence %s must redact private material", item.Kind)
	}
	return nil
}

func (item LaunchEvidenceItem) validateKind() error {
	switch item.Kind {
	case LaunchEvidenceLocalnet100Validators:
		if item.ValidatorCount < 100 || item.ValidatorCount > 128 {
			return errors.New("launch evidence requires stable 100-128 validator localnet")
		}
	case LaunchEvidenceStressFinality:
		if item.StressFinalityP95Seconds == 0 || item.StressFinalityP95Seconds > 90 || item.WorstFinalitySeconds > 120 {
			return errors.New("launch evidence stress finality exceeds target")
		}
	case LaunchEvidenceRestartPersistence:
		if item.RestartHeightBefore == 0 || item.RestartHeightAfter <= item.RestartHeightBefore {
			return errors.New("launch evidence restart must prove height advances after restart")
		}
	case LaunchEvidenceSnapshotStateSync:
		if item.SnapshotHeight == 0 || item.StateSyncHeight < item.SnapshotHeight {
			return errors.New("launch evidence snapshot/state sync recovery is incomplete")
		}
	case LaunchEvidenceAVMStateGrowth:
		if item.AVMStateGrowthBytes == 0 || item.AVMStateGrowthLimitBytes == 0 || item.AVMStateGrowthBytes > item.AVMStateGrowthLimitBytes {
			return errors.New("launch evidence AVM state growth exceeds configured limit")
		}
	default:
		return fmt.Errorf("unsupported launch evidence %s", item.Kind)
	}
	return nil
}

func requiredLaunchEvidenceKinds() []string {
	kinds := []string{
		LaunchEvidenceLocalnet100Validators,
		LaunchEvidenceStressFinality,
		LaunchEvidenceRestartPersistence,
		LaunchEvidenceSnapshotStateSync,
		LaunchEvidenceAVMStateGrowth,
	}
	sort.Strings(kinds)
	return kinds
}
