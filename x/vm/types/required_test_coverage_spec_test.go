package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMRequiredTestCoverageSpecMatchesSection20(t *testing.T) {
	spec, err := DefaultAVMRequiredTestCoverageSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Equal(t, "AVM required test coverage", spec.SpecName)
	require.Equal(t, ComputeAVMRequiredTestCoverageSpecHash(spec), spec.SpecHash)
	require.Len(t, spec.Groups, 5)

	byCategory := map[AVMRequiredTestCoverageCategory]AVMRequiredTestCoverageGroup{}
	for _, group := range spec.Groups {
		require.NoError(t, group.Validate())
		byCategory[group.Category] = group
	}
	require.Len(t, byCategory[AVMTestCoverageCategoryUnit].Cases, 11)
	require.Len(t, byCategory[AVMTestCoverageCategoryIntegration].Cases, 10)
	require.Len(t, byCategory[AVMTestCoverageCategoryInvariant].Cases, 9)
	require.Len(t, byCategory[AVMTestCoverageCategoryFuzz].Cases, 9)
	require.Len(t, byCategory[AVMTestCoverageCategoryPerformance].Cases, 9)
}

func TestAVMRequiredUnitCoverageMatchesSection201(t *testing.T) {
	group, err := NewAVMRequiredTestCoverageGroup(AVMRequiredTestCoverageGroup{
		Category:	AVMTestCoverageCategoryUnit,
		Cases: []AVMRequiredTestCoverageCase{
			AVMUnitCoverageMessageIDDerivation,
			AVMUnitCoverageSenderNonceValidation,
			AVMUnitCoverageQueueSortKeyOrdering,
			AVMUnitCoverageDelayHeightEligibility,
			AVMUnitCoverageExpiryHandling,
			AVMUnitCoverageRetryPolicyHandling,
			AVMUnitCoverageBounceMessageConstruction,
			AVMUnitCoverageDeadLetterTransition,
			AVMUnitCoverageGasPolicyCalculation,
			AVMUnitCoverageInterfaceHashCalculation,
			AVMUnitCoverageReceiptHashCalculation,
		},
	})
	require.NoError(t, err)
	require.NoError(t, group.Validate())
	require.Equal(t, ComputeAVMRequiredTestCoverageGroupHash(group), group.GroupHash)
	require.ElementsMatch(t, []AVMRequiredTestCoverageCase{
		AVMUnitCoverageMessageIDDerivation,
		AVMUnitCoverageSenderNonceValidation,
		AVMUnitCoverageQueueSortKeyOrdering,
		AVMUnitCoverageDelayHeightEligibility,
		AVMUnitCoverageExpiryHandling,
		AVMUnitCoverageRetryPolicyHandling,
		AVMUnitCoverageBounceMessageConstruction,
		AVMUnitCoverageDeadLetterTransition,
		AVMUnitCoverageGasPolicyCalculation,
		AVMUnitCoverageInterfaceHashCalculation,
		AVMUnitCoverageReceiptHashCalculation,
	}, group.Cases)
}

func TestAVMRequiredIntegrationCoverageMatchesSection202(t *testing.T) {
	group, err := NewAVMRequiredTestCoverageGroup(AVMRequiredTestCoverageGroup{
		Category:	AVMTestCoverageCategoryIntegration,
		Cases: []AVMRequiredTestCoverageCase{
			AVMIntegrationCoverageSyncRouterExecution,
			AVMIntegrationCoverageAsyncFutureBlockExecution,
			AVMIntegrationCoverageDelayedMessageExecution,
			AVMIntegrationCoverageFailedMessageBounce,
			AVMIntegrationCoverageRetryExhaustionDeadLetter,
			AVMIntegrationCoverageCrossZoneExecution,
			AVMIntegrationCoverageActorMailboxExecution,
			AVMIntegrationCoverageContinuationResume,
			AVMIntegrationCoverageContractEmitsAsyncMessage,
			AVMIntegrationCoverageInterfaceDescriptorQuery,
		},
	})
	require.NoError(t, err)
	require.NoError(t, group.Validate())
	require.Equal(t, ComputeAVMRequiredTestCoverageGroupHash(group), group.GroupHash)
	require.ElementsMatch(t, []AVMRequiredTestCoverageCase{
		AVMIntegrationCoverageSyncRouterExecution,
		AVMIntegrationCoverageAsyncFutureBlockExecution,
		AVMIntegrationCoverageDelayedMessageExecution,
		AVMIntegrationCoverageFailedMessageBounce,
		AVMIntegrationCoverageRetryExhaustionDeadLetter,
		AVMIntegrationCoverageCrossZoneExecution,
		AVMIntegrationCoverageActorMailboxExecution,
		AVMIntegrationCoverageContinuationResume,
		AVMIntegrationCoverageContractEmitsAsyncMessage,
		AVMIntegrationCoverageInterfaceDescriptorQuery,
	}, group.Cases)
}

func TestAVMRequiredInvariantCoverageMatchesSection203(t *testing.T) {
	group, err := NewAVMRequiredTestCoverageGroup(AVMRequiredTestCoverageGroup{
		Category:	AVMTestCoverageCategoryInvariant,
		Cases: []AVMRequiredTestCoverageCase{
			AVMInvariantCoverageExecutedMessageOneReceipt,
			AVMInvariantCoverageQueuedMessageStoredRecord,
			AVMInvariantCoverageConsumedMessageNoReplay,
			AVMInvariantCoverageExpiredMessageCannotExecute,
			AVMInvariantCoverageBounceCannotOverRefund,
			AVMInvariantCoverageZoneRootQueueContinuationRoots,
			AVMInvariantCoverageActorMailboxOrderDeterministic,
			AVMInvariantCoverageActorStateIsolationEnforced,
			AVMInvariantCoverageContractStoragePrefixIsolated,
		},
	})
	require.NoError(t, err)
	require.NoError(t, group.Validate())
	require.Equal(t, ComputeAVMRequiredTestCoverageGroupHash(group), group.GroupHash)
	require.ElementsMatch(t, []AVMRequiredTestCoverageCase{
		AVMInvariantCoverageExecutedMessageOneReceipt,
		AVMInvariantCoverageQueuedMessageStoredRecord,
		AVMInvariantCoverageConsumedMessageNoReplay,
		AVMInvariantCoverageExpiredMessageCannotExecute,
		AVMInvariantCoverageBounceCannotOverRefund,
		AVMInvariantCoverageZoneRootQueueContinuationRoots,
		AVMInvariantCoverageActorMailboxOrderDeterministic,
		AVMInvariantCoverageActorStateIsolationEnforced,
		AVMInvariantCoverageContractStoragePrefixIsolated,
	}, group.Cases)
}

func TestAVMRequiredFuzzCoverageMatchesSection204(t *testing.T) {
	group, err := NewAVMRequiredTestCoverageGroup(AVMRequiredTestCoverageGroup{
		Category:	AVMTestCoverageCategoryFuzz,
		Cases: []AVMRequiredTestCoverageCase{
			AVMFuzzCoverageMalformedAsyncMessages,
			AVMFuzzCoverageRandomNonceOrdering,
			AVMFuzzCoverageQueuePriorityEdgeCases,
			AVMFuzzCoverageRetryExpiryBoundaries,
			AVMFuzzCoverageBouncePayloadLimits,
			AVMFuzzCoverageActorHandlerFailures,
			AVMFuzzCoverageContinuationStatePayloads,
			AVMFuzzCoverageContractStorageKeys,
			AVMFuzzCoverageInterfaceSchemaPayloads,
		},
	})
	require.NoError(t, err)
	require.NoError(t, group.Validate())
	require.Equal(t, ComputeAVMRequiredTestCoverageGroupHash(group), group.GroupHash)
	require.ElementsMatch(t, []AVMRequiredTestCoverageCase{
		AVMFuzzCoverageMalformedAsyncMessages,
		AVMFuzzCoverageRandomNonceOrdering,
		AVMFuzzCoverageQueuePriorityEdgeCases,
		AVMFuzzCoverageRetryExpiryBoundaries,
		AVMFuzzCoverageBouncePayloadLimits,
		AVMFuzzCoverageActorHandlerFailures,
		AVMFuzzCoverageContinuationStatePayloads,
		AVMFuzzCoverageContractStorageKeys,
		AVMFuzzCoverageInterfaceSchemaPayloads,
	}, group.Cases)
}

func TestAVMRequiredPerformanceCoverageMatchesSection205(t *testing.T) {
	group, err := NewAVMRequiredTestCoverageGroup(AVMRequiredTestCoverageGroup{
		Category:	AVMTestCoverageCategoryPerformance,
		Cases: []AVMRequiredTestCoverageCase{
			AVMPerformanceCoverageQueueInsertThroughput,
			AVMPerformanceCoverageQueueDrainThroughput,
			AVMPerformanceCoverageAsyncExecutionThroughput,
			AVMPerformanceCoverageActorMailboxThroughput,
			AVMPerformanceCoverageContinuationResumeThroughput,
			AVMPerformanceCoverageCrossZoneThroughput,
			AVMPerformanceCoverageReceiptProofGenerationLatency,
			AVMPerformanceCoverageAVMRootGenerationLatency,
			AVMPerformanceCoverageBlockSTMConflictRateByWorkload,
		},
	})
	require.NoError(t, err)
	require.NoError(t, group.Validate())
	require.Equal(t, ComputeAVMRequiredTestCoverageGroupHash(group), group.GroupHash)
	require.ElementsMatch(t, []AVMRequiredTestCoverageCase{
		AVMPerformanceCoverageQueueInsertThroughput,
		AVMPerformanceCoverageQueueDrainThroughput,
		AVMPerformanceCoverageAsyncExecutionThroughput,
		AVMPerformanceCoverageActorMailboxThroughput,
		AVMPerformanceCoverageContinuationResumeThroughput,
		AVMPerformanceCoverageCrossZoneThroughput,
		AVMPerformanceCoverageReceiptProofGenerationLatency,
		AVMPerformanceCoverageAVMRootGenerationLatency,
		AVMPerformanceCoverageBlockSTMConflictRateByWorkload,
	}, group.Cases)
}

func TestAVMRequiredCoverageRejectsMissingDuplicateCrossCategoryAndHashMismatch(t *testing.T) {
	spec, err := DefaultAVMRequiredTestCoverageSpec()
	require.NoError(t, err)

	missingCategory := spec
	missingCategory.Groups = missingCategory.Groups[:1]
	missingCategory.SpecHash = ComputeAVMRequiredTestCoverageSpecHash(missingCategory)
	require.ErrorContains(t, missingCategory.Validate(), "every section 20 coverage category")

	mutated := spec
	mutated.Groups[0].Cases[0] = AVMUnitCoverageMessageIDDerivation
	mutated.Groups[0].GroupHash = ComputeAVMRequiredTestCoverageGroupHash(mutated.Groups[0])
	mutated.SpecHash = ComputeAVMRequiredTestCoverageSpecHash(mutated)
	require.ErrorContains(t, mutated.Validate(), "case")

	spec, err = DefaultAVMRequiredTestCoverageSpec()
	require.NoError(t, err)
	invariant := coverageGroupForTest(t, spec, AVMTestCoverageCategoryInvariant)
	invariant.Cases[0] = AVMFuzzCoverageMalformedAsyncMessages
	invariant.GroupHash = ComputeAVMRequiredTestCoverageGroupHash(invariant)
	require.ErrorContains(t, invariant.Validate(), "case")

	spec, err = DefaultAVMRequiredTestCoverageSpec()
	require.NoError(t, err)
	duplicate := spec
	duplicate.Groups[1] = spec.Groups[0]
	duplicate.SpecHash = ComputeAVMRequiredTestCoverageSpecHash(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	spec, err = DefaultAVMRequiredTestCoverageSpec()
	require.NoError(t, err)
	hashMismatch := spec
	hashMismatch.Groups[0].GroupHash = "0000000000000000000000000000000000000000000000000000000000000000"
	require.ErrorContains(t, hashMismatch.Validate(), "group hash mismatch")
}

func TestAVMRequiredCoverageRejectsInvalidNames(t *testing.T) {
	group, err := NewAVMRequiredTestCoverageGroup(AVMRequiredTestCoverageGroup{
		Category:	AVMTestCoverageCategoryUnit,
		Cases: []AVMRequiredTestCoverageCase{
			AVMUnitCoverageMessageIDDerivation,
		},
	})
	require.ErrorContains(t, err, "case entries")
	require.NotEmpty(t, group.GroupHash)

	_, err = NewAVMRequiredTestCoverageSpec(AVMRequiredTestCoverageSpec{
		SpecName:	"AVM|required test coverage",
		Groups:		[]AVMRequiredTestCoverageGroup{},
	})
	require.ErrorContains(t, err, "invalid character")
}

func coverageGroupForTest(t *testing.T, spec AVMRequiredTestCoverageSpec, category AVMRequiredTestCoverageCategory) AVMRequiredTestCoverageGroup {
	t.Helper()
	for _, group := range spec.Groups {
		if group.Category == category {
			return group
		}
	}
	t.Fatalf("missing coverage group %s", category)
	return AVMRequiredTestCoverageGroup{}
}
