package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConflictReductionSpecCoversSectionThirteenTwo(t *testing.T) {
	require.NoError(t, ValidateConflictReductionCoverage())

	spec, err := DefaultConflictReductionSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Rules, 6)
	require.NotEmpty(t, spec.Root)

	byID := map[ConflictReductionRuleID]ConflictReductionRule{}
	for _, rule := range spec.Rules {
		require.NoError(t, rule.Validate())
		byID[rule.RuleID] = rule
	}

	require.Contains(t, byID[ConflictRuleAvoidGlobalCounters].Enforcement, "global")
	require.Contains(t, byID[ConflictRuleShardFeeAccumulator].Rule, "per-shard fee")
	require.Contains(t, byID[ConflictRuleShardMessageQueues].Evidence, "BlockSTMMessageBatch")
	require.Contains(t, byID[ConflictRuleVersionedObjects].Evidence, "ConflictKey")
	require.Contains(t, byID[ConflictRuleObjectLocalLocks].Rule, "object-local locks")
	require.Contains(t, byID[ConflictRuleAsyncRemoteWrites].Enforcement, "ViaMessage")
}

func TestConflictReductionSpecRootIsCanonicalAndRejectsTamper(t *testing.T) {
	spec, err := DefaultConflictReductionSpec()
	require.NoError(t, err)

	reordered := append([]ConflictReductionRule(nil), ConflictReductionRules()...)
	slices.Reverse(reordered)
	reorderedSpec, err := BuildConflictReductionSpec(reordered)
	require.NoError(t, err)
	require.Equal(t, spec.Root, reorderedSpec.Root)
	require.Equal(t, spec.Rules, reorderedSpec.Rules)

	_, err = BuildConflictReductionSpec([]ConflictReductionRule{ConflictReductionRules()[0], ConflictReductionRules()[0]})
	require.ErrorContains(t, err, "duplicate")

	tampered := ConflictReductionRules()[0]
	tampered.Enforcement = strings.ReplaceAll(tampered.Enforcement, "global", "local")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}

func TestConflictReductionRulesBindToBlockSTMPlanValidation(t *testing.T) {
	schedule, err := BuildProposalSchedule(30, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "0", "hot", 1, 30, 0),
	}, TestnetParams())
	require.NoError(t, err)

	_, err = BuildBlockSTMZonePerformancePlan(schedule, []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDFinancial, "0", ZoneIDContract, "0", "contract/storage/remote", BlockSTMAccessWrite, false),
	}, nil)
	require.ErrorContains(t, err, "cross-zone writes")

	coreSchedule, err := BuildProposalSchedule(31, []ProposalItem{
		testProposalItem(ZoneIDAetraCore, "0", "global-hot", 1, 31, 0),
	}, TestnetParams())
	require.NoError(t, err)
	_, err = BuildBlockSTMZonePerformancePlan(coreSchedule, []BlockSTMStateAccess{
		blockSTMAccess(ZoneIDAetraCore, "0", ZoneIDAetraCore, "0", "core/global-lock", BlockSTMAccessWrite, false),
	}, nil)
	require.ErrorContains(t, err, "global write lock")
}

func TestStoreV2UsageSpecCoversSectionThirteenThree(t *testing.T) {
	require.NoError(t, ValidateStoreV2UsageCoverage())

	spec, err := DefaultStoreV2UsageSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Layouts, 5)
	require.Len(t, spec.Benchmarks, 8)
	require.NotEmpty(t, spec.Root)

	layoutKinds := map[StoreV2LayoutKind]StoreV2LayoutDescriptor{}
	for _, layout := range spec.Layouts {
		require.NoError(t, layout.Validate())
		layoutKinds[layout.Kind] = layout
	}
	require.Contains(t, layoutKinds[StoreV2LayoutObjectStore].StateClasses, "accounts")
	require.Contains(t, layoutKinds[StoreV2LayoutKVHybrid].StateClasses, "contract-storage")
	require.Contains(t, layoutKinds[StoreV2LayoutPrefixProof].ProofSurface, "shard prefix proof")
	require.Contains(t, layoutKinds[StoreV2LayoutCompactRoot].ProofSurface, "global zone root")
	require.True(t, layoutKinds[StoreV2LayoutBoundedRangeScan].BoundedScan)

	benchmarks := map[StoreV2BenchmarkID]StoreV2BenchmarkRequirement{}
	for _, benchmark := range spec.Benchmarks {
		require.NoError(t, benchmark.Validate())
		benchmarks[benchmark.BenchmarkID] = benchmark
	}
	require.Contains(t, benchmarks[StoreV2BenchmarkDirectBalanceRead].Operation, "balance")
	require.Contains(t, benchmarks[StoreV2BenchmarkDirectIdentityResolve].Operation, "identity")
	require.Contains(t, benchmarks[StoreV2BenchmarkRecursiveIdentity].Operation, "recursive")
	require.Contains(t, benchmarks[StoreV2BenchmarkContractStorageRW].Operation, "contract storage")
	require.Contains(t, benchmarks[StoreV2BenchmarkMessageEnqueueDequeue].Operation, "enqueue")
	require.Contains(t, benchmarks[StoreV2BenchmarkPaymentChannelSettle].Operation, "payment channel")
	require.Contains(t, benchmarks[StoreV2BenchmarkDEXPoolUpdate].Operation, "DEX pool")
	require.Contains(t, benchmarks[StoreV2BenchmarkProofGeneration].Operation, "proof")
}

func TestStoreV2UsageSpecRootIsCanonicalAndRejectsBadReferences(t *testing.T) {
	spec, err := DefaultStoreV2UsageSpec()
	require.NoError(t, err)

	layouts := append([]StoreV2LayoutDescriptor(nil), StoreV2LayoutDescriptors()...)
	benchmarks := append([]StoreV2BenchmarkRequirement(nil), StoreV2BenchmarkRequirements()...)
	slices.Reverse(layouts)
	slices.Reverse(benchmarks)
	reordered, err := BuildStoreV2UsageSpec(layouts, benchmarks)
	require.NoError(t, err)
	require.Equal(t, spec.Root, reordered.Root)
	require.Equal(t, spec.Layouts, reordered.Layouts)
	require.Equal(t, spec.Benchmarks, reordered.Benchmarks)

	_, err = BuildStoreV2UsageSpec([]StoreV2LayoutDescriptor{StoreV2LayoutDescriptors()[0], StoreV2LayoutDescriptors()[0]}, StoreV2BenchmarkRequirements())
	require.ErrorContains(t, err, "duplicate")

	badBenchmark := StoreV2BenchmarkRequirements()[0]
	badBenchmark.LayoutID = "missing-layout"
	badBenchmark.DescriptorHash = ComputeStoreV2BenchmarkDescriptorHash(badBenchmark)
	_, err = BuildStoreV2UsageSpec(StoreV2LayoutDescriptors(), []StoreV2BenchmarkRequirement{badBenchmark})
	require.ErrorContains(t, err, "missing layout")

	tampered := StoreV2LayoutDescriptors()[0]
	tampered.ProofSurface = "uncommitted proof"
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}
