package launchreadiness

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLaunchEvidenceReportRequiresProductionEvidence(t *testing.T) {
	report := completeLaunchEvidenceReport()
	require.NoError(t, ValidateLaunchEvidenceReport(report))

	for _, kind := range requiredLaunchEvidenceKinds() {
		filtered := report
		filtered.Evidence = nil
		for _, item := range report.Evidence {
			if item.Kind != kind {
				filtered.Evidence = append(filtered.Evidence, item)
			}
		}
		err := ValidateLaunchEvidenceReport(filtered)
		require.ErrorContains(t, err, "missing launch evidence "+kind)
	}
}

func TestLaunchEvidenceReportRejectsUnsafeOrIncompleteEvidence(t *testing.T) {
	tests := []struct {
		name	string
		mutate	func(*LaunchEvidenceReport)
		wantErr	string
	}{
		{
			name:	"too few validators",
			mutate: func(report *LaunchEvidenceReport) {
				item := findEvidence(report, LaunchEvidenceLocalnet100Validators)
				item.ValidatorCount = 99
			},
			wantErr:	"100-128 validator localnet",
		},
		{
			name:	"stress finality above target",
			mutate: func(report *LaunchEvidenceReport) {
				item := findEvidence(report, LaunchEvidenceStressFinality)
				item.StressFinalityP95Seconds = 91
			},
			wantErr:	"stress finality exceeds target",
		},
		{
			name:	"restart does not advance",
			mutate: func(report *LaunchEvidenceReport) {
				item := findEvidence(report, LaunchEvidenceRestartPersistence)
				item.RestartHeightAfter = item.RestartHeightBefore
			},
			wantErr:	"height advances after restart",
		},
		{
			name:	"state sync before snapshot",
			mutate: func(report *LaunchEvidenceReport) {
				item := findEvidence(report, LaunchEvidenceSnapshotStateSync)
				item.StateSyncHeight = item.SnapshotHeight - 1
			},
			wantErr:	"snapshot/state sync recovery",
		},
		{
			name:	"avm growth over limit",
			mutate: func(report *LaunchEvidenceReport) {
				item := findEvidence(report, LaunchEvidenceAVMStateGrowth)
				item.AVMStateGrowthBytes = item.AVMStateGrowthLimitBytes + 1
			},
			wantErr:	"AVM state growth exceeds",
		},
		{
			name:	"private material not redacted",
			mutate: func(report *LaunchEvidenceReport) {
				report.Evidence[0].PrivateMaterialRedacted = false
			},
			wantErr:	"redact private material",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := completeLaunchEvidenceReport()
			tt.mutate(&report)
			require.ErrorContains(t, ValidateLaunchEvidenceReport(report), tt.wantErr)
		})
	}
}

func TestPatternTriageReportClassifiesKnownSensitiveMatches(t *testing.T) {
	report := PatternTriageReport{
		DeterministicSensitiveMatches:	DeterministicSensitiveMatchCount,
		PanicMustMatches:		PanicMustMatchCount,
		Findings: []PatternTriageFinding{
			{
				ID:		"gw-must-patterns",
				Pattern:	"runtime.MustPattern",
				PatternKind:	"panic_must",
				MatchCount:	220,
				File:		"x/*/types/query.pb.gw.go",
				Category:	"generated_gateway_must",
				Generated:	true,
				Action:		"accept_generated",
			},
			{
				ID:		"module-init-export-panic",
				Pattern:	"panic(err)",
				PatternKind:	"panic_must",
				MatchCount:	64,
				File:		"x/*/module.go",
				Category:	"sdk_module_init_export_panic",
				Action:		"document_startup_invariant",
			},
			{
				ID:		"consensus-sensitive-review",
				Pattern:	"range map / time / rand / panic",
				PatternKind:	"deterministic_sensitive",
				MatchCount:	47,
				File:		"app,x",
				Category:	"consensus_path_review_required",
				ConsensusPath:	true,
				Action:		"requires_owner_review",
			},
		},
	}
	require.NoError(t, ValidatePatternTriageReport(report))
}

func TestPatternTriageReportRejectsUnclassifiedOrUnsafeConsensusPatterns(t *testing.T) {
	report := PatternTriageReport{
		DeterministicSensitiveMatches:	DeterministicSensitiveMatchCount - 1,
		PanicMustMatches:		PanicMustMatchCount,
		Findings:			[]PatternTriageFinding{validConsensusFinding()},
	}
	require.ErrorContains(t, ValidatePatternTriageReport(report), "deterministic-sensitive match count changed")

	report = PatternTriageReport{
		DeterministicSensitiveMatches:	DeterministicSensitiveMatchCount,
		PanicMustMatches:		PanicMustMatchCount,
	}
	require.ErrorContains(t, ValidatePatternTriageReport(report), "findings are required")

	report = PatternTriageReport{
		DeterministicSensitiveMatches:	DeterministicSensitiveMatchCount,
		PanicMustMatches:		PanicMustMatchCount,
		Findings: []PatternTriageFinding{{
			ID:		"unsafe-consensus",
			Pattern:	"panic",
			PatternKind:	"panic_must",
			MatchCount:	PanicMustMatchCount,
			File:		"app/lifecycle/block.go",
			Category:	"operator_startup_invariant",
			ConsensusPath:	true,
			Action:		"document_startup_invariant",
		}},
	}
	require.ErrorContains(t, ValidatePatternTriageReport(report), "consensus-path pattern requires")

	report = PatternTriageReport{
		DeterministicSensitiveMatches:	DeterministicSensitiveMatchCount,
		PanicMustMatches:		PanicMustMatchCount,
		Findings:			[]PatternTriageFinding{validConsensusFinding()},
	}
	require.ErrorContains(t, ValidatePatternTriageReport(report), "panic/Must triage coverage mismatch")
}

func completeLaunchEvidenceReport() LaunchEvidenceReport {
	base := LaunchEvidenceItem{
		ArtifactPath:			"reports/launch/localnet.json",
		ValidatorCount:			100,
		ObservedBlocks:			2_000,
		NormalFinalityP95Seconds:	12,
		StressFinalityP95Seconds:	70,
		WorstFinalitySeconds:		110,
		RestartHeightBefore:		500,
		RestartHeightAfter:		550,
		SnapshotHeight:			1_000,
		StateSyncHeight:		1_005,
		AVMStateGrowthBytes:		64 * 1024 * 1024,
		AVMStateGrowthLimitBytes:	128 * 1024 * 1024,
		DeterministicReplayHash:	strings.Repeat("a", 64),
		OperatorSummaryIncluded:	true,
		PrivateMaterialRedacted:	true,
	}
	kinds := requiredLaunchEvidenceKinds()
	report := LaunchEvidenceReport{Evidence: make([]LaunchEvidenceItem, 0, len(kinds))}
	for _, kind := range kinds {
		item := base
		item.Kind = kind
		item.ArtifactPath = "reports/launch/" + kind + ".json"
		report.Evidence = append(report.Evidence, item)
	}
	return report
}

func findEvidence(report *LaunchEvidenceReport, kind string) *LaunchEvidenceItem {
	for i := range report.Evidence {
		if report.Evidence[i].Kind == kind {
			return &report.Evidence[i]
		}
	}
	panic("missing fixture evidence")
}

func validConsensusFinding() PatternTriageFinding {
	return PatternTriageFinding{
		ID:		"consensus-sensitive-review",
		Pattern:	"range map",
		PatternKind:	"deterministic_sensitive",
		MatchCount:	DeterministicSensitiveMatchCount,
		File:		"x/example/keeper.go",
		Category:	"consensus_path_review_required",
		ConsensusPath:	true,
		Action:		"requires_owner_review",
	}
}
