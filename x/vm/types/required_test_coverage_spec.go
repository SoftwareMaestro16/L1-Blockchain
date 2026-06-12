package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	AVMTestCoverageCategoryUnit		AVMRequiredTestCoverageCategory	= "unit"
	AVMTestCoverageCategoryIntegration	AVMRequiredTestCoverageCategory	= "integration"
	AVMTestCoverageCategoryInvariant	AVMRequiredTestCoverageCategory	= "invariant"
	AVMTestCoverageCategoryFuzz		AVMRequiredTestCoverageCategory	= "fuzz"
	AVMTestCoverageCategoryPerformance	AVMRequiredTestCoverageCategory	= "performance"

	AVMUnitCoverageMessageIDDerivation		AVMRequiredTestCoverageCase	= "message_id_derivation"
	AVMUnitCoverageSenderNonceValidation		AVMRequiredTestCoverageCase	= "sender_nonce_validation"
	AVMUnitCoverageQueueSortKeyOrdering		AVMRequiredTestCoverageCase	= "queue_sort_key_ordering"
	AVMUnitCoverageDelayHeightEligibility		AVMRequiredTestCoverageCase	= "delay_height_eligibility"
	AVMUnitCoverageExpiryHandling			AVMRequiredTestCoverageCase	= "expiry_handling"
	AVMUnitCoverageRetryPolicyHandling		AVMRequiredTestCoverageCase	= "retry_policy_handling"
	AVMUnitCoverageBounceMessageConstruction	AVMRequiredTestCoverageCase	= "bounce_message_construction"
	AVMUnitCoverageDeadLetterTransition		AVMRequiredTestCoverageCase	= "dead_letter_transition"
	AVMUnitCoverageGasPolicyCalculation		AVMRequiredTestCoverageCase	= "gas_policy_calculation"
	AVMUnitCoverageInterfaceHashCalculation		AVMRequiredTestCoverageCase	= "interface_hash_calculation"
	AVMUnitCoverageReceiptHashCalculation		AVMRequiredTestCoverageCase	= "receipt_hash_calculation"

	AVMIntegrationCoverageSyncRouterExecution	AVMRequiredTestCoverageCase	= "sync_module_execution_through_avm_router"
	AVMIntegrationCoverageAsyncFutureBlockExecution	AVMRequiredTestCoverageCase	= "async_message_submitted_and_executed_in_future_block"
	AVMIntegrationCoverageDelayedMessageExecution	AVMRequiredTestCoverageCase	= "delayed_message_execution"
	AVMIntegrationCoverageFailedMessageBounce	AVMRequiredTestCoverageCase	= "failed_message_bounce"
	AVMIntegrationCoverageRetryExhaustionDeadLetter	AVMRequiredTestCoverageCase	= "retry_exhaustion_to_dead_letter_queue"
	AVMIntegrationCoverageCrossZoneExecution	AVMRequiredTestCoverageCase	= "cross_zone_message_execution"
	AVMIntegrationCoverageActorMailboxExecution	AVMRequiredTestCoverageCase	= "actor_mailbox_execution"
	AVMIntegrationCoverageContinuationResume	AVMRequiredTestCoverageCase	= "continuation_resume"
	AVMIntegrationCoverageContractEmitsAsyncMessage	AVMRequiredTestCoverageCase	= "contract_emits_async_message"
	AVMIntegrationCoverageInterfaceDescriptorQuery	AVMRequiredTestCoverageCase	= "interface_descriptor_queried_by_client"

	AVMInvariantCoverageExecutedMessageOneReceipt		AVMRequiredTestCoverageCase	= "every_executed_message_has_one_receipt"
	AVMInvariantCoverageQueuedMessageStoredRecord		AVMRequiredTestCoverageCase	= "every_queued_message_has_stored_message_record"
	AVMInvariantCoverageConsumedMessageNoReplay		AVMRequiredTestCoverageCase	= "consumed_message_cannot_be_replayed"
	AVMInvariantCoverageExpiredMessageCannotExecute		AVMRequiredTestCoverageCase	= "expired_message_cannot_execute"
	AVMInvariantCoverageBounceCannotOverRefund		AVMRequiredTestCoverageCase	= "bounce_cannot_over_refund_value"
	AVMInvariantCoverageZoneRootQueueContinuationRoots	AVMRequiredTestCoverageCase	= "zone_root_includes_queue_and_continuation_roots"
	AVMInvariantCoverageActorMailboxOrderDeterministic	AVMRequiredTestCoverageCase	= "actor_mailbox_order_is_deterministic"
	AVMInvariantCoverageActorStateIsolationEnforced		AVMRequiredTestCoverageCase	= "actor_state_isolation_is_enforced"
	AVMInvariantCoverageContractStoragePrefixIsolated	AVMRequiredTestCoverageCase	= "contract_storage_key_prefix_is_isolated"

	AVMFuzzCoverageMalformedAsyncMessages		AVMRequiredTestCoverageCase	= "malformed_async_messages"
	AVMFuzzCoverageRandomNonceOrdering		AVMRequiredTestCoverageCase	= "random_nonce_ordering"
	AVMFuzzCoverageQueuePriorityEdgeCases		AVMRequiredTestCoverageCase	= "queue_priority_edge_cases"
	AVMFuzzCoverageRetryExpiryBoundaries		AVMRequiredTestCoverageCase	= "retry_and_expiry_boundary_conditions"
	AVMFuzzCoverageBouncePayloadLimits		AVMRequiredTestCoverageCase	= "bounce_payload_limits"
	AVMFuzzCoverageActorHandlerFailures		AVMRequiredTestCoverageCase	= "actor_handler_failures"
	AVMFuzzCoverageContinuationStatePayloads	AVMRequiredTestCoverageCase	= "continuation_state_payloads"
	AVMFuzzCoverageContractStorageKeys		AVMRequiredTestCoverageCase	= "contract_storage_keys"
	AVMFuzzCoverageInterfaceSchemaPayloads		AVMRequiredTestCoverageCase	= "interface_schema_payloads"

	AVMPerformanceCoverageQueueInsertThroughput		AVMRequiredTestCoverageCase	= "queue_insert_throughput"
	AVMPerformanceCoverageQueueDrainThroughput		AVMRequiredTestCoverageCase	= "queue_drain_throughput"
	AVMPerformanceCoverageAsyncExecutionThroughput		AVMRequiredTestCoverageCase	= "async_message_execution_throughput"
	AVMPerformanceCoverageActorMailboxThroughput		AVMRequiredTestCoverageCase	= "actor_mailbox_throughput"
	AVMPerformanceCoverageContinuationResumeThroughput	AVMRequiredTestCoverageCase	= "continuation_resume_throughput"
	AVMPerformanceCoverageCrossZoneThroughput		AVMRequiredTestCoverageCase	= "cross_zone_message_throughput"
	AVMPerformanceCoverageReceiptProofGenerationLatency	AVMRequiredTestCoverageCase	= "receipt_proof_generation_latency"
	AVMPerformanceCoverageAVMRootGenerationLatency		AVMRequiredTestCoverageCase	= "avm_root_generation_latency"
	AVMPerformanceCoverageBlockSTMConflictRateByWorkload	AVMRequiredTestCoverageCase	= "blockstm_conflict_rate_by_workload"

	MaxAVMRequiredTestCoverageCategories	= 8
	MaxAVMRequiredTestCoverageCases		= 64
	MaxAVMRequiredTestCoverageNameBytes	= 128
)

type AVMRequiredTestCoverageCategory string
type AVMRequiredTestCoverageCase string

type AVMRequiredTestCoverageGroup struct {
	Category	AVMRequiredTestCoverageCategory
	Cases		[]AVMRequiredTestCoverageCase
	GroupHash	string
}

type AVMRequiredTestCoverageSpec struct {
	SpecName	string
	Groups		[]AVMRequiredTestCoverageGroup
	SpecHash	string
}

func DefaultAVMRequiredTestCoverageSpec() (AVMRequiredTestCoverageSpec, error) {
	groups := []AVMRequiredTestCoverageGroup{
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
		{
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
		},
	}
	for i := range groups {
		group, err := NewAVMRequiredTestCoverageGroup(groups[i])
		if err != nil {
			return AVMRequiredTestCoverageSpec{}, err
		}
		groups[i] = group
	}
	return NewAVMRequiredTestCoverageSpec(AVMRequiredTestCoverageSpec{
		SpecName:	"AVM required test coverage",
		Groups:		groups,
	})
}

func NewAVMRequiredTestCoverageGroup(group AVMRequiredTestCoverageGroup) (AVMRequiredTestCoverageGroup, error) {
	group = canonicalAVMRequiredTestCoverageGroup(group)
	group.GroupHash = ComputeAVMRequiredTestCoverageGroupHash(group)
	return group, group.Validate()
}

func (g AVMRequiredTestCoverageGroup) Validate() error {
	g = canonicalAVMRequiredTestCoverageGroup(g)
	if !IsAVMRequiredTestCoverageCategory(g.Category) {
		return fmt.Errorf("invalid AVM required test coverage category %q", g.Category)
	}
	if err := validateAVMRequiredTestCoverageCases(g.Category, g.Cases); err != nil {
		return err
	}
	if g.GroupHash == "" {
		return errors.New("AVM required test coverage group hash is required")
	}
	if err := validateAVMComparisonHash("AVM required test coverage group hash", g.GroupHash); err != nil {
		return err
	}
	if g.GroupHash != ComputeAVMRequiredTestCoverageGroupHash(g) {
		return errors.New("AVM required test coverage group hash mismatch")
	}
	return nil
}

func NewAVMRequiredTestCoverageSpec(spec AVMRequiredTestCoverageSpec) (AVMRequiredTestCoverageSpec, error) {
	spec = canonicalAVMRequiredTestCoverageSpec(spec)
	spec.SpecHash = ComputeAVMRequiredTestCoverageSpecHash(spec)
	return spec, spec.Validate()
}

func (s AVMRequiredTestCoverageSpec) Validate() error {
	s = canonicalAVMRequiredTestCoverageSpec(s)
	if err := validateAVMRequiredTestCoverageName("AVM required test coverage spec name", s.SpecName); err != nil {
		return err
	}
	required := AllAVMRequiredTestCoverageCategories()
	if len(s.Groups) != len(required) || len(s.Groups) > MaxAVMRequiredTestCoverageCategories {
		return errors.New("AVM required test coverage spec must contain every section 20 coverage category")
	}
	seen := make(map[AVMRequiredTestCoverageCategory]struct{}, len(s.Groups))
	for i, group := range s.Groups {
		if err := group.Validate(); err != nil {
			return err
		}
		if _, found := seen[group.Category]; found {
			return fmt.Errorf("duplicate AVM required test coverage category %q", group.Category)
		}
		seen[group.Category] = struct{}{}
		if i > 0 && s.Groups[i-1].Category >= group.Category {
			return errors.New("AVM required test coverage groups must be sorted canonically")
		}
	}
	for _, category := range required {
		if _, found := seen[category]; !found {
			return fmt.Errorf("missing AVM required test coverage category %q", category)
		}
	}
	if s.SpecHash == "" {
		return errors.New("AVM required test coverage spec hash is required")
	}
	if err := validateAVMComparisonHash("AVM required test coverage spec hash", s.SpecHash); err != nil {
		return err
	}
	if s.SpecHash != ComputeAVMRequiredTestCoverageSpecHash(s) {
		return errors.New("AVM required test coverage spec hash mismatch")
	}
	return nil
}

func AllAVMRequiredTestCoverageCategories() []AVMRequiredTestCoverageCategory {
	categories := []AVMRequiredTestCoverageCategory{AVMTestCoverageCategoryFuzz, AVMTestCoverageCategoryIntegration, AVMTestCoverageCategoryInvariant, AVMTestCoverageCategoryPerformance, AVMTestCoverageCategoryUnit}
	sort.Slice(categories, func(i, j int) bool { return categories[i] < categories[j] })
	return categories
}

func IsAVMRequiredTestCoverageCategory(category AVMRequiredTestCoverageCategory) bool {
	switch category {
	case AVMTestCoverageCategoryUnit, AVMTestCoverageCategoryIntegration, AVMTestCoverageCategoryInvariant, AVMTestCoverageCategoryFuzz, AVMTestCoverageCategoryPerformance:
		return true
	default:
		return false
	}
}

func ComputeAVMRequiredTestCoverageGroupHash(group AVMRequiredTestCoverageGroup) string {
	group = canonicalAVMRequiredTestCoverageGroup(group)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-required-test-coverage-group-v1")
	writeEnginePart(h, string(group.Category))
	writeEngineUint64(h, uint64(len(group.Cases)))
	for _, coverageCase := range group.Cases {
		writeEnginePart(h, string(coverageCase))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMRequiredTestCoverageSpecHash(spec AVMRequiredTestCoverageSpec) string {
	spec = canonicalAVMRequiredTestCoverageSpec(spec)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-required-test-coverage-spec-v1")
	writeEnginePart(h, spec.SpecName)
	writeEngineUint64(h, uint64(len(spec.Groups)))
	for _, group := range spec.Groups {
		writeEnginePart(h, group.GroupHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMRequiredTestCoverageGroup(group AVMRequiredTestCoverageGroup) AVMRequiredTestCoverageGroup {
	group.Category = AVMRequiredTestCoverageCategory(strings.TrimSpace(string(group.Category)))
	group.Cases = sortedAVMRequiredTestCoverageCases(group.Cases)
	group.GroupHash = strings.TrimSpace(group.GroupHash)
	return group
}

func canonicalAVMRequiredTestCoverageSpec(spec AVMRequiredTestCoverageSpec) AVMRequiredTestCoverageSpec {
	spec.SpecName = strings.TrimSpace(spec.SpecName)
	spec.Groups = append([]AVMRequiredTestCoverageGroup(nil), spec.Groups...)
	for i := range spec.Groups {
		spec.Groups[i] = canonicalAVMRequiredTestCoverageGroup(spec.Groups[i])
	}
	sort.SliceStable(spec.Groups, func(i, j int) bool {
		return spec.Groups[i].Category < spec.Groups[j].Category
	})
	spec.SpecHash = strings.TrimSpace(spec.SpecHash)
	return spec
}

func validateAVMRequiredTestCoverageCases(category AVMRequiredTestCoverageCategory, values []AVMRequiredTestCoverageCase) error {
	required := requiredAVMRequiredTestCoverageCases(category)
	if len(values) != len(required) || len(values) > MaxAVMRequiredTestCoverageCases {
		return fmt.Errorf("AVM required test coverage expected %d case entries", len(required))
	}
	seen := make(map[AVMRequiredTestCoverageCase]struct{}, len(values))
	previous := ""
	for _, value := range values {
		if !isAVMRequiredTestCoverageCase(value) {
			return fmt.Errorf("AVM required test coverage unknown case %q", value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return errors.New("AVM required test coverage case entries must be sorted canonically")
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("AVM required test coverage duplicate case %s", value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("AVM required test coverage missing case %s", value)
		}
	}
	return nil
}

func requiredAVMRequiredTestCoverageCases(category AVMRequiredTestCoverageCategory) []AVMRequiredTestCoverageCase {
	switch category {
	case AVMTestCoverageCategoryUnit:
		return []AVMRequiredTestCoverageCase{
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
		}
	case AVMTestCoverageCategoryIntegration:
		return []AVMRequiredTestCoverageCase{
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
		}
	case AVMTestCoverageCategoryInvariant:
		return []AVMRequiredTestCoverageCase{
			AVMInvariantCoverageExecutedMessageOneReceipt,
			AVMInvariantCoverageQueuedMessageStoredRecord,
			AVMInvariantCoverageConsumedMessageNoReplay,
			AVMInvariantCoverageExpiredMessageCannotExecute,
			AVMInvariantCoverageBounceCannotOverRefund,
			AVMInvariantCoverageZoneRootQueueContinuationRoots,
			AVMInvariantCoverageActorMailboxOrderDeterministic,
			AVMInvariantCoverageActorStateIsolationEnforced,
			AVMInvariantCoverageContractStoragePrefixIsolated,
		}
	case AVMTestCoverageCategoryFuzz:
		return []AVMRequiredTestCoverageCase{
			AVMFuzzCoverageMalformedAsyncMessages,
			AVMFuzzCoverageRandomNonceOrdering,
			AVMFuzzCoverageQueuePriorityEdgeCases,
			AVMFuzzCoverageRetryExpiryBoundaries,
			AVMFuzzCoverageBouncePayloadLimits,
			AVMFuzzCoverageActorHandlerFailures,
			AVMFuzzCoverageContinuationStatePayloads,
			AVMFuzzCoverageContractStorageKeys,
			AVMFuzzCoverageInterfaceSchemaPayloads,
		}
	case AVMTestCoverageCategoryPerformance:
		return []AVMRequiredTestCoverageCase{
			AVMPerformanceCoverageQueueInsertThroughput,
			AVMPerformanceCoverageQueueDrainThroughput,
			AVMPerformanceCoverageAsyncExecutionThroughput,
			AVMPerformanceCoverageActorMailboxThroughput,
			AVMPerformanceCoverageContinuationResumeThroughput,
			AVMPerformanceCoverageCrossZoneThroughput,
			AVMPerformanceCoverageReceiptProofGenerationLatency,
			AVMPerformanceCoverageAVMRootGenerationLatency,
			AVMPerformanceCoverageBlockSTMConflictRateByWorkload,
		}
	default:
		return nil
	}
}

func isAVMRequiredTestCoverageCase(value AVMRequiredTestCoverageCase) bool {
	for _, category := range AllAVMRequiredTestCoverageCategories() {
		for _, required := range requiredAVMRequiredTestCoverageCases(category) {
			if value == required {
				return true
			}
		}
	}
	return false
}

func validateAVMRequiredTestCoverageName(fieldName, value string) error {
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxAVMRequiredTestCoverageNameBytes {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMRequiredTestCoverageNameBytes)
	}
	for _, r := range value {
		if r < 0x20 || r == '|' {
			return fmt.Errorf("%s contains invalid character", fieldName)
		}
	}
	return nil
}

func sortedAVMRequiredTestCoverageCases(values []AVMRequiredTestCoverageCase) []AVMRequiredTestCoverageCase {
	out := append([]AVMRequiredTestCoverageCase(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
