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
	AVMAcceptanceSyncExecutionWrapsMsgServer		AVMAcceptanceCriterion	= "sync_execution_wraps_existing_cosmos_sdk_msgserver_calls"
	AVMAcceptanceAsyncLifecycleComplete			AVMAcceptanceCriterion	= "async_messages_can_be_scheduled_executed_retried_bounced_expired_and_dead_lettered"
	AVMAcceptanceQueueOrderingDeterministicTested		AVMAcceptanceCriterion	= "queue_ordering_is_deterministic_and_test_covered"
	AVMAcceptanceCrossZoneQueueReceiptRoots			AVMAcceptanceCriterion	= "cross_zone_messages_use_queue_and_receipt_roots"
	AVMAcceptanceActorMailboxIsolation			AVMAcceptanceCriterion	= "actor_runtime_supports_isolated_mailbox_execution"
	AVMAcceptanceContinuationsPauseResume			AVMAcceptanceCriterion	= "continuations_support_pause_and_resume_across_blocks"
	AVMAcceptanceContractBackendsExplicitInterface		AVMAcceptanceCriterion	= "contract_backends_are_isolated_behind_explicit_runtime_interface"
	AVMAcceptanceGasModelComplete				AVMAcceptanceCriterion	= "gas_model_covers_execution_storage_scheduling_routing_proof_verification_and_continuation_storage"
	AVMAcceptanceInterfaceRegistryComplete			AVMAcceptanceCriterion	= "interface_registry_supports_methods_events_async_handlers_and_get_methods"
	AVMAcceptanceRootCommitmentsComplete			AVMAcceptanceCriterion	= "root_commitments_include_router_message_actor_contract_continuation_interface_and_receipt_roots"
	AVMAcceptanceReplayProtection				AVMAcceptanceCriterion	= "replay_protection_prevents_duplicate_message_execution"
	AVMAcceptanceStoreV2PrefixProofQueryable		AVMAcceptanceCriterion	= "store_v2_layout_is_prefix_isolated_and_proof_queryable"
	AVMAcceptanceBlockSTMConflictStrategyBenchmarked	AVMAcceptanceCriterion	= "blockstm_conflict_strategy_is_defined_and_benchmarked"

	MaxAVMAcceptanceCriteria	= 64
	MaxAVMAcceptanceNameBytes	= 192
)

type AVMAcceptanceCriterion string

type AVMAcceptanceCriteriaSpec struct {
	Criteria	[]AVMAcceptanceCriterion
	SpecHash	string
}

func DefaultAVMAcceptanceCriteriaSpec() (AVMAcceptanceCriteriaSpec, error) {
	spec := AVMAcceptanceCriteriaSpec{Criteria: AllAVMAcceptanceCriteria()}
	spec.SpecHash = ComputeAVMAcceptanceCriteriaSpecHash(spec)
	return spec, spec.Validate()
}

func (s AVMAcceptanceCriteriaSpec) Validate() error {
	s = canonicalAVMAcceptanceCriteriaSpec(s)
	if err := validateAVMAcceptanceCriteria(s.Criteria); err != nil {
		return err
	}
	if s.SpecHash == "" {
		return errors.New("AVM acceptance criteria spec hash is required")
	}
	if err := validateAVMComparisonHash("AVM acceptance criteria spec hash", s.SpecHash); err != nil {
		return err
	}
	if s.SpecHash != ComputeAVMAcceptanceCriteriaSpecHash(s) {
		return errors.New("AVM acceptance criteria spec hash mismatch")
	}
	return nil
}

func AllAVMAcceptanceCriteria() []AVMAcceptanceCriterion {
	criteria := []AVMAcceptanceCriterion{
		AVMAcceptanceSyncExecutionWrapsMsgServer,
		AVMAcceptanceAsyncLifecycleComplete,
		AVMAcceptanceQueueOrderingDeterministicTested,
		AVMAcceptanceCrossZoneQueueReceiptRoots,
		AVMAcceptanceActorMailboxIsolation,
		AVMAcceptanceContinuationsPauseResume,
		AVMAcceptanceContractBackendsExplicitInterface,
		AVMAcceptanceGasModelComplete,
		AVMAcceptanceInterfaceRegistryComplete,
		AVMAcceptanceRootCommitmentsComplete,
		AVMAcceptanceReplayProtection,
		AVMAcceptanceStoreV2PrefixProofQueryable,
		AVMAcceptanceBlockSTMConflictStrategyBenchmarked,
	}
	sort.Slice(criteria, func(i, j int) bool { return criteria[i] < criteria[j] })
	return criteria
}

func IsAVMAcceptanceCriterion(criterion AVMAcceptanceCriterion) bool {
	for _, required := range AllAVMAcceptanceCriteria() {
		if criterion == required {
			return true
		}
	}
	return false
}

func ComputeAVMAcceptanceCriteriaSpecHash(spec AVMAcceptanceCriteriaSpec) string {
	spec = canonicalAVMAcceptanceCriteriaSpec(spec)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-acceptance-criteria-spec-v1")
	writeEngineUint64(h, uint64(len(spec.Criteria)))
	for _, criterion := range spec.Criteria {
		writeEnginePart(h, string(criterion))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMAcceptanceCriteriaSpec(spec AVMAcceptanceCriteriaSpec) AVMAcceptanceCriteriaSpec {
	spec.Criteria = append([]AVMAcceptanceCriterion(nil), spec.Criteria...)
	for i := range spec.Criteria {
		spec.Criteria[i] = AVMAcceptanceCriterion(strings.TrimSpace(string(spec.Criteria[i])))
	}
	sort.SliceStable(spec.Criteria, func(i, j int) bool {
		return spec.Criteria[i] < spec.Criteria[j]
	})
	spec.SpecHash = strings.TrimSpace(spec.SpecHash)
	return spec
}

func validateAVMAcceptanceCriteria(criteria []AVMAcceptanceCriterion) error {
	required := AllAVMAcceptanceCriteria()
	if len(criteria) != len(required) || len(criteria) > MaxAVMAcceptanceCriteria {
		return errors.New("AVM acceptance criteria spec must contain every section 22 criterion")
	}
	seen := make(map[AVMAcceptanceCriterion]struct{}, len(criteria))
	previous := ""
	for _, criterion := range criteria {
		if err := validateAVMAcceptanceName("AVM acceptance criterion", string(criterion)); err != nil {
			return err
		}
		if !IsAVMAcceptanceCriterion(criterion) {
			return fmt.Errorf("invalid AVM acceptance criterion %q", criterion)
		}
		if _, found := seen[criterion]; found {
			return fmt.Errorf("duplicate AVM acceptance criterion %q", criterion)
		}
		current := string(criterion)
		if previous != "" && previous >= current {
			return errors.New("AVM acceptance criteria must be sorted canonically")
		}
		previous = current
		seen[criterion] = struct{}{}
	}
	for _, criterion := range required {
		if _, found := seen[criterion]; !found {
			return fmt.Errorf("AVM acceptance criteria spec missing criterion %s", criterion)
		}
	}
	return nil
}

func validateAVMAcceptanceName(fieldName, value string) error {
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxAVMAcceptanceNameBytes {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMAcceptanceNameBytes)
	}
	for _, r := range value {
		if r < 0x20 || r == '|' {
			return fmt.Errorf("%s contains invalid character", fieldName)
		}
	}
	return nil
}
