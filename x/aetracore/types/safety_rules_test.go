package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafetyRulesSpecCoversSectionsFourteenOneAndTwo(t *testing.T) {
	require.NoError(t, ValidateSafetyRulesCoverage())

	spec, err := DefaultSafetyRulesSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Determinism, 6)
	require.Len(t, spec.RoutingSafety, 5)
	require.NotEmpty(t, spec.Root)

	determinism := map[DeterminismRuleID]DeterminismRule{}
	for _, rule := range spec.Determinism {
		require.NoError(t, rule.Validate())
		determinism[rule.RuleID] = rule
	}
	require.Contains(t, determinism[DeterminismNoExternalAPIs].Rule, "external API")
	require.Contains(t, determinism[DeterminismNoLocalClock].Enforcement, "block height/time")
	require.Contains(t, determinism[DeterminismNoRandomShard].Evidence, "ShardLayout")
	require.Contains(t, determinism[DeterminismNoMapIteration].Enforcement, "sort canonical")
	require.Contains(t, determinism[DeterminismNoMempoolOnlyOutput].Rule, "mempool-only")
	require.Contains(t, determinism[DeterminismNoFloatingPoint].Enforcement, "integers")

	routing := map[RoutingSafetyRuleID]RoutingSafetyRule{}
	for _, rule := range spec.RoutingSafety {
		require.NoError(t, rule.Validate())
		routing[rule.RuleID] = rule
	}
	require.Contains(t, routing[RoutingSafetyCommittedTable].Evidence, "RoutingTableCommitment")
	require.Contains(t, routing[RoutingSafetyCommittedMetrics].Enforcement, "committed heights")
	require.Contains(t, routing[RoutingSafetyDeterministicTieBreak].Enforcement, "total cost")
	require.Contains(t, routing[RoutingSafetyRouteFailureReceipt].Rule, "receipt")
	require.Contains(t, routing[RoutingSafetyBounceValueCap].Rule, "extra value")
}

func TestSafetyRulesSpecRootIsCanonicalAndRejectsTamper(t *testing.T) {
	spec, err := DefaultSafetyRulesSpec()
	require.NoError(t, err)

	determinism := append([]DeterminismRule(nil), DeterminismRules()...)
	routing := append([]RoutingSafetyRule(nil), RoutingSafetyRules()...)
	slices.Reverse(determinism)
	slices.Reverse(routing)
	reordered, err := BuildSafetyRulesSpec(determinism, routing)
	require.NoError(t, err)
	require.Equal(t, spec.Root, reordered.Root)
	require.Equal(t, spec.Determinism, reordered.Determinism)
	require.Equal(t, spec.RoutingSafety, reordered.RoutingSafety)

	_, err = BuildSafetyRulesSpec([]DeterminismRule{DeterminismRules()[0], DeterminismRules()[0]}, RoutingSafetyRules())
	require.ErrorContains(t, err, "duplicate")

	tampered := DeterminismRules()[0]
	tampered.Enforcement = strings.ReplaceAll(tampered.Enforcement, "network", "random")
	require.ErrorContains(t, tampered.Validate(), "hash mismatch")
}

func TestConsensusDeterminismEvidenceRejectsAllForbiddenInputs(t *testing.T) {
	evidence := ConsensusDeterminismEvidence{}
	evidence.EvidenceHash = ComputeConsensusDeterminismEvidenceHash(evidence)
	require.NoError(t, ValidateConsensusDeterminismEvidence(evidence))

	cases := []struct {
		name	string
		mutate	func(*ConsensusDeterminismEvidence)
		err	string
	}{
		{"external API calls", func(e *ConsensusDeterminismEvidence) { e.ExternalAPICalls = true }, "external API"},
		{"local clock", func(e *ConsensusDeterminismEvidence) { e.LocalClockUsage = true }, "local clock"},
		{"random shard", func(e *ConsensusDeterminismEvidence) { e.RandomShardPlacement = true }, "random shard"},
		{"map iteration", func(e *ConsensusDeterminismEvidence) { e.MapIterationOutput = true }, "map iteration"},
		{"mempool only", func(e *ConsensusDeterminismEvidence) { e.MempoolOnlyResult = true }, "mempool-only"},
		{"floating point", func(e *ConsensusDeterminismEvidence) { e.FloatingPointMath = true }, "floating point"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := ConsensusDeterminismEvidence{}
			tc.mutate(&bad)
			bad.EvidenceHash = ComputeConsensusDeterminismEvidenceHash(bad)
			require.ErrorContains(t, ValidateConsensusDeterminismEvidence(bad), tc.err)
		})
	}

	tampered := evidence
	tampered.FloatingPointMath = true
	require.ErrorContains(t, ValidateConsensusDeterminismEvidence(tampered), "floating point")
}

func TestRoutingSafetyEvidenceRequiresCommittedRootsReceiptsAndBounceCap(t *testing.T) {
	evidence := RoutingSafetyEvidence{
		RoutingTableHash:		hashParts("routing-table"),
		RoutingMetricsRoot:		hashParts("routing-metrics"),
		DeterministicTieBreakHash:	hashParts("tie-break"),
		FailureReceiptRoot:		hashParts("failure-receipt"),
		BounceValueCapHash:		hashParts("bounce-cap"),
		FailedRouteValueNAET:		100,
		ReceiptedValueNAET:		100,
		OriginalBounceValueNAET:	25,
		BounceValueNAET:		25,
	}
	evidence.EvidenceHash = ComputeRoutingSafetyEvidenceHash(evidence)
	require.NoError(t, ValidateRoutingSafetyEvidence(evidence))

	missingReceipt := evidence
	missingReceipt.ReceiptedValueNAET = 0
	missingReceipt.EvidenceHash = ComputeRoutingSafetyEvidenceHash(missingReceipt)
	require.ErrorContains(t, ValidateRoutingSafetyEvidence(missingReceipt), "requires receipt")

	overReceipt := evidence
	overReceipt.ReceiptedValueNAET = 101
	overReceipt.EvidenceHash = ComputeRoutingSafetyEvidenceHash(overReceipt)
	require.ErrorContains(t, ValidateRoutingSafetyEvidence(overReceipt), "exceeds failed route")

	overBounce := evidence
	overBounce.BounceValueNAET = 26
	overBounce.EvidenceHash = ComputeRoutingSafetyEvidenceHash(overBounce)
	require.ErrorContains(t, ValidateRoutingSafetyEvidence(overBounce), "extra value")

	tampered := evidence
	tampered.RoutingMetricsRoot = hashParts("different-metrics")
	require.ErrorContains(t, ValidateRoutingSafetyEvidence(tampered), "hash mismatch")
}
