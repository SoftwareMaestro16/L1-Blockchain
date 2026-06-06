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
	AVMRoadmapPhase0 AVMImplementationRoadmapPhaseID = "phase_0_specification_and_test_vectors"
	AVMRoadmapPhase1 AVMImplementationRoadmapPhaseID = "phase_1_sync_router"
	AVMRoadmapPhase2 AVMImplementationRoadmapPhaseID = "phase_2_async_engine"

	AVMRoadmapTaskCanonicalAsyncMessageEncoding AVMImplementationRoadmapTask = "define_canonical_async_message_encoding"
	AVMRoadmapTaskMessageIDDerivation           AVMImplementationRoadmapTask = "define_message_id_derivation"
	AVMRoadmapTaskDeterministicQueueSortKey     AVMImplementationRoadmapTask = "define_deterministic_queue_sort_key"
	AVMRoadmapTaskExecutionReceiptSchema        AVMImplementationRoadmapTask = "define_execution_receipt_schema"
	AVMRoadmapTaskAVMRootSchema                 AVMImplementationRoadmapTask = "define_avm_root_schema"
	AVMRoadmapTaskGasPolicySchema               AVMImplementationRoadmapTask = "define_gas_policy_schema"
	AVMRoadmapTaskInterfaceDescriptorSchema     AVMImplementationRoadmapTask = "define_interface_descriptor_schema"

	AVMRoadmapTaskRouterSkeleton       AVMImplementationRoadmapTask = "implement_avm_router_skeleton"
	AVMRoadmapTaskSyncEngineWrapper    AVMImplementationRoadmapTask = "route_msgserver_calls_through_sync_engine_wrapper"
	AVMRoadmapTaskSyncReceipts         AVMImplementationRoadmapTask = "add_execution_receipts_for_routed_sync_messages"
	AVMRoadmapTaskZoneRouteDescriptors AVMImplementationRoadmapTask = "add_zone_route_descriptors"
	AVMRoadmapTaskAVMRootPlaceholder   AVMImplementationRoadmapTask = "add_avm_root_placeholder"

	AVMRoadmapTaskAsyncMessageStore AVMImplementationRoadmapTask = "implement_async_message_store"
	AVMRoadmapTaskZoneQueues        AVMImplementationRoadmapTask = "implement_zone_queues"
	AVMRoadmapTaskDelayedQueue      AVMImplementationRoadmapTask = "implement_delayed_queue"
	AVMRoadmapTaskRetryQueue        AVMImplementationRoadmapTask = "implement_retry_queue"
	AVMRoadmapTaskDeadLetterQueue   AVMImplementationRoadmapTask = "implement_dead_letter_queue"
	AVMRoadmapTaskReplayTombstones  AVMImplementationRoadmapTask = "implement_replay_tombstones"
	AVMRoadmapTaskQueueRoots        AVMImplementationRoadmapTask = "add_queue_roots"

	AVMRoadmapExitSignableHashableObjectsHaveTestVectors AVMImplementationExitCriterion = "signable_hashable_objects_have_test_vectors"
	AVMRoadmapExitQueueOrderingTestCovered               AVMImplementationExitCriterion = "queue_ordering_test_covered"
	AVMRoadmapExitRootEncodingFixed                      AVMImplementationExitCriterion = "root_encoding_fixed"

	AVMRoadmapExitExistingCallsRepresentedAsRoutedSync AVMImplementationExitCriterion = "existing_module_calls_represented_as_avm_routed_sync_execution"
	AVMRoadmapExitDeterministicSyncReceipts            AVMImplementationExitCriterion = "sync_receipts_emitted_deterministically"

	AVMRoadmapExitAsyncScheduledLaterBlocks  AVMImplementationExitCriterion = "async_messages_scheduled_and_executed_in_later_blocks"
	AVMRoadmapExitExpiredMessagesDoNotRun    AVMImplementationExitCriterion = "expired_messages_do_not_execute"
	AVMRoadmapExitRetryDeadLetterDeterminism AVMImplementationExitCriterion = "retry_and_dead_letter_flows_are_deterministic"

	AVMRoadmapVectorAsyncMessageEncoding      AVMTestVectorTarget = "async_message_encoding"
	AVMRoadmapVectorMessageIDDerivation       AVMTestVectorTarget = "message_id_derivation"
	AVMRoadmapVectorDeterministicQueueSort    AVMTestVectorTarget = "deterministic_queue_sort_key"
	AVMRoadmapVectorExecutionReceiptSchema    AVMTestVectorTarget = "execution_receipt_schema"
	AVMRoadmapVectorAVMRootSchema             AVMTestVectorTarget = "avm_root_schema"
	AVMRoadmapVectorGasPolicySchema           AVMTestVectorTarget = "gas_policy_schema"
	AVMRoadmapVectorInterfaceDescriptorSchema AVMTestVectorTarget = "interface_descriptor_schema"

	MaxAVMRoadmapPhases       = 16
	MaxAVMRoadmapTasks        = 32
	MaxAVMRoadmapExitCriteria = 16
	MaxAVMTestVectorTargets   = 32
	MaxAVMRoadmapTextBytes    = 128
)

type AVMImplementationRoadmapPhaseID string
type AVMImplementationRoadmapTask string
type AVMImplementationExitCriterion string
type AVMTestVectorTarget string

type AVMImplementationRoadmapPhase struct {
	PhaseID           AVMImplementationRoadmapPhaseID
	Name              string
	Tasks             []AVMImplementationRoadmapTask
	ExitCriteria      []AVMImplementationExitCriterion
	TestVectorTargets []AVMTestVectorTarget
	ConsensusCritical bool
	PhaseHash         string
}

type AVMImplementationRoadmap struct {
	RoadmapName string
	Phases      []AVMImplementationRoadmapPhase
	RoadmapHash string
}

func DefaultAVMImplementationRoadmap() (AVMImplementationRoadmap, error) {
	phases := []AVMImplementationRoadmapPhase{
		{
			PhaseID: AVMRoadmapPhase0,
			Name:    "Specification and Test Vectors",
			Tasks: []AVMImplementationRoadmapTask{
				AVMRoadmapTaskCanonicalAsyncMessageEncoding,
				AVMRoadmapTaskMessageIDDerivation,
				AVMRoadmapTaskDeterministicQueueSortKey,
				AVMRoadmapTaskExecutionReceiptSchema,
				AVMRoadmapTaskAVMRootSchema,
				AVMRoadmapTaskGasPolicySchema,
				AVMRoadmapTaskInterfaceDescriptorSchema,
			},
			ExitCriteria: []AVMImplementationExitCriterion{
				AVMRoadmapExitSignableHashableObjectsHaveTestVectors,
				AVMRoadmapExitQueueOrderingTestCovered,
				AVMRoadmapExitRootEncodingFixed,
			},
			TestVectorTargets: []AVMTestVectorTarget{
				AVMRoadmapVectorAsyncMessageEncoding,
				AVMRoadmapVectorMessageIDDerivation,
				AVMRoadmapVectorDeterministicQueueSort,
				AVMRoadmapVectorExecutionReceiptSchema,
				AVMRoadmapVectorAVMRootSchema,
				AVMRoadmapVectorGasPolicySchema,
				AVMRoadmapVectorInterfaceDescriptorSchema,
			},
			ConsensusCritical: true,
		},
		{
			PhaseID: AVMRoadmapPhase1,
			Name:    "Sync Router",
			Tasks: []AVMImplementationRoadmapTask{
				AVMRoadmapTaskRouterSkeleton,
				AVMRoadmapTaskSyncEngineWrapper,
				AVMRoadmapTaskSyncReceipts,
				AVMRoadmapTaskZoneRouteDescriptors,
				AVMRoadmapTaskAVMRootPlaceholder,
			},
			ExitCriteria: []AVMImplementationExitCriterion{
				AVMRoadmapExitExistingCallsRepresentedAsRoutedSync,
				AVMRoadmapExitDeterministicSyncReceipts,
			},
			ConsensusCritical: true,
		},
		{
			PhaseID: AVMRoadmapPhase2,
			Name:    "Async Engine",
			Tasks: []AVMImplementationRoadmapTask{
				AVMRoadmapTaskAsyncMessageStore,
				AVMRoadmapTaskZoneQueues,
				AVMRoadmapTaskDelayedQueue,
				AVMRoadmapTaskRetryQueue,
				AVMRoadmapTaskDeadLetterQueue,
				AVMRoadmapTaskReplayTombstones,
				AVMRoadmapTaskQueueRoots,
			},
			ExitCriteria: []AVMImplementationExitCriterion{
				AVMRoadmapExitAsyncScheduledLaterBlocks,
				AVMRoadmapExitExpiredMessagesDoNotRun,
				AVMRoadmapExitRetryDeadLetterDeterminism,
			},
			ConsensusCritical: true,
		},
	}
	for i := range phases {
		phase, err := NewAVMImplementationRoadmapPhase(phases[i])
		if err != nil {
			return AVMImplementationRoadmap{}, err
		}
		phases[i] = phase
	}
	return NewAVMImplementationRoadmap(AVMImplementationRoadmap{
		RoadmapName: "AVM implementation roadmap",
		Phases:      phases,
	})
}

func NewAVMImplementationRoadmapPhase(phase AVMImplementationRoadmapPhase) (AVMImplementationRoadmapPhase, error) {
	phase = canonicalAVMImplementationRoadmapPhase(phase)
	phase.PhaseHash = ComputeAVMImplementationRoadmapPhaseHash(phase)
	return phase, phase.Validate()
}

func (p AVMImplementationRoadmapPhase) Validate() error {
	p = canonicalAVMImplementationRoadmapPhase(p)
	if !IsAVMImplementationRoadmapPhaseID(p.PhaseID) {
		return fmt.Errorf("invalid AVM implementation roadmap phase %q", p.PhaseID)
	}
	if err := validateAVMRoadmapText("AVM roadmap phase name", p.Name); err != nil {
		return err
	}
	if !p.ConsensusCritical {
		return errors.New("AVM roadmap phase must be consensus-critical")
	}
	if err := validateAVMRoadmapTasks(p.PhaseID, p.Tasks); err != nil {
		return err
	}
	if err := validateAVMRoadmapExitCriteria(p.PhaseID, p.ExitCriteria); err != nil {
		return err
	}
	if err := validateAVMTestVectorTargets(p.PhaseID, p.TestVectorTargets); err != nil {
		return err
	}
	if p.PhaseHash == "" {
		return errors.New("AVM roadmap phase hash is required")
	}
	if err := validateAVMComparisonHash("AVM roadmap phase hash", p.PhaseHash); err != nil {
		return err
	}
	if p.PhaseHash != ComputeAVMImplementationRoadmapPhaseHash(p) {
		return errors.New("AVM roadmap phase hash mismatch")
	}
	return nil
}

func NewAVMImplementationRoadmap(roadmap AVMImplementationRoadmap) (AVMImplementationRoadmap, error) {
	roadmap = canonicalAVMImplementationRoadmap(roadmap)
	roadmap.RoadmapHash = ComputeAVMImplementationRoadmapHash(roadmap)
	return roadmap, roadmap.Validate()
}

func (r AVMImplementationRoadmap) Validate() error {
	r = canonicalAVMImplementationRoadmap(r)
	if err := validateAVMRoadmapText("AVM roadmap name", r.RoadmapName); err != nil {
		return err
	}
	required := AllAVMImplementationRoadmapPhaseIDs()
	if len(r.Phases) != len(required) || len(r.Phases) > MaxAVMRoadmapPhases {
		return errors.New("AVM implementation roadmap must contain every section 19 phase")
	}
	seen := make(map[AVMImplementationRoadmapPhaseID]struct{}, len(r.Phases))
	for i, phase := range r.Phases {
		if err := phase.Validate(); err != nil {
			return err
		}
		if _, found := seen[phase.PhaseID]; found {
			return fmt.Errorf("duplicate AVM implementation roadmap phase %q", phase.PhaseID)
		}
		seen[phase.PhaseID] = struct{}{}
		if i > 0 && r.Phases[i-1].PhaseID >= phase.PhaseID {
			return errors.New("AVM implementation roadmap phases must be sorted canonically")
		}
	}
	for _, phaseID := range required {
		if _, found := seen[phaseID]; !found {
			return fmt.Errorf("missing AVM implementation roadmap phase %q", phaseID)
		}
	}
	if r.RoadmapHash == "" {
		return errors.New("AVM implementation roadmap hash is required")
	}
	if err := validateAVMComparisonHash("AVM implementation roadmap hash", r.RoadmapHash); err != nil {
		return err
	}
	if r.RoadmapHash != ComputeAVMImplementationRoadmapHash(r) {
		return errors.New("AVM implementation roadmap hash mismatch")
	}
	return nil
}

func AllAVMImplementationRoadmapPhaseIDs() []AVMImplementationRoadmapPhaseID {
	phases := []AVMImplementationRoadmapPhaseID{AVMRoadmapPhase0, AVMRoadmapPhase1, AVMRoadmapPhase2}
	sort.Slice(phases, func(i, j int) bool { return phases[i] < phases[j] })
	return phases
}

func IsAVMImplementationRoadmapPhaseID(phaseID AVMImplementationRoadmapPhaseID) bool {
	switch phaseID {
	case AVMRoadmapPhase0, AVMRoadmapPhase1, AVMRoadmapPhase2:
		return true
	default:
		return false
	}
}

func ComputeAVMImplementationRoadmapPhaseHash(phase AVMImplementationRoadmapPhase) string {
	phase = canonicalAVMImplementationRoadmapPhase(phase)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-implementation-roadmap-phase-v1")
	writeEnginePart(h, string(phase.PhaseID))
	writeEnginePart(h, phase.Name)
	writeEngineUint64(h, uint64(len(phase.Tasks)))
	for _, task := range phase.Tasks {
		writeEnginePart(h, string(task))
	}
	writeEngineUint64(h, uint64(len(phase.ExitCriteria)))
	for _, criterion := range phase.ExitCriteria {
		writeEnginePart(h, string(criterion))
	}
	writeEngineUint64(h, uint64(len(phase.TestVectorTargets)))
	for _, target := range phase.TestVectorTargets {
		writeEnginePart(h, string(target))
	}
	writeEngineBool(h, phase.ConsensusCritical)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMImplementationRoadmapHash(roadmap AVMImplementationRoadmap) string {
	roadmap = canonicalAVMImplementationRoadmap(roadmap)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-implementation-roadmap-v1")
	writeEnginePart(h, roadmap.RoadmapName)
	writeEngineUint64(h, uint64(len(roadmap.Phases)))
	for _, phase := range roadmap.Phases {
		writeEnginePart(h, phase.PhaseHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMImplementationRoadmapPhase(phase AVMImplementationRoadmapPhase) AVMImplementationRoadmapPhase {
	phase.PhaseID = AVMImplementationRoadmapPhaseID(strings.TrimSpace(string(phase.PhaseID)))
	phase.Name = strings.TrimSpace(phase.Name)
	phase.Tasks = sortedAVMRoadmapTasks(phase.Tasks)
	phase.ExitCriteria = sortedAVMRoadmapExitCriteria(phase.ExitCriteria)
	phase.TestVectorTargets = sortedAVMTestVectorTargets(phase.TestVectorTargets)
	phase.PhaseHash = strings.TrimSpace(phase.PhaseHash)
	return phase
}

func canonicalAVMImplementationRoadmap(roadmap AVMImplementationRoadmap) AVMImplementationRoadmap {
	roadmap.RoadmapName = strings.TrimSpace(roadmap.RoadmapName)
	roadmap.Phases = append([]AVMImplementationRoadmapPhase(nil), roadmap.Phases...)
	for i := range roadmap.Phases {
		roadmap.Phases[i] = canonicalAVMImplementationRoadmapPhase(roadmap.Phases[i])
	}
	sort.SliceStable(roadmap.Phases, func(i, j int) bool {
		return roadmap.Phases[i].PhaseID < roadmap.Phases[j].PhaseID
	})
	roadmap.RoadmapHash = strings.TrimSpace(roadmap.RoadmapHash)
	return roadmap
}

func validateAVMRoadmapTasks(phaseID AVMImplementationRoadmapPhaseID, values []AVMImplementationRoadmapTask) error {
	return validateAVMRoadmapEnumSet("task", values, requiredAVMRoadmapTasks(phaseID), MaxAVMRoadmapTasks, isAVMRoadmapTask)
}

func validateAVMRoadmapExitCriteria(phaseID AVMImplementationRoadmapPhaseID, values []AVMImplementationExitCriterion) error {
	return validateAVMRoadmapEnumSet("exit criterion", values, requiredAVMRoadmapExitCriteria(phaseID), MaxAVMRoadmapExitCriteria, isAVMRoadmapExitCriterion)
}

func validateAVMTestVectorTargets(phaseID AVMImplementationRoadmapPhaseID, values []AVMTestVectorTarget) error {
	return validateAVMRoadmapEnumSet("test vector target", values, requiredAVMTestVectorTargets(phaseID), MaxAVMTestVectorTargets, isAVMTestVectorTarget)
}

func requiredAVMRoadmapTasks(phaseID AVMImplementationRoadmapPhaseID) []AVMImplementationRoadmapTask {
	switch phaseID {
	case AVMRoadmapPhase0:
		return []AVMImplementationRoadmapTask{
			AVMRoadmapTaskCanonicalAsyncMessageEncoding,
			AVMRoadmapTaskMessageIDDerivation,
			AVMRoadmapTaskDeterministicQueueSortKey,
			AVMRoadmapTaskExecutionReceiptSchema,
			AVMRoadmapTaskAVMRootSchema,
			AVMRoadmapTaskGasPolicySchema,
			AVMRoadmapTaskInterfaceDescriptorSchema,
		}
	case AVMRoadmapPhase1:
		return []AVMImplementationRoadmapTask{
			AVMRoadmapTaskRouterSkeleton,
			AVMRoadmapTaskSyncEngineWrapper,
			AVMRoadmapTaskSyncReceipts,
			AVMRoadmapTaskZoneRouteDescriptors,
			AVMRoadmapTaskAVMRootPlaceholder,
		}
	case AVMRoadmapPhase2:
		return []AVMImplementationRoadmapTask{
			AVMRoadmapTaskAsyncMessageStore,
			AVMRoadmapTaskZoneQueues,
			AVMRoadmapTaskDelayedQueue,
			AVMRoadmapTaskRetryQueue,
			AVMRoadmapTaskDeadLetterQueue,
			AVMRoadmapTaskReplayTombstones,
			AVMRoadmapTaskQueueRoots,
		}
	default:
		return nil
	}
}

func requiredAVMRoadmapExitCriteria(phaseID AVMImplementationRoadmapPhaseID) []AVMImplementationExitCriterion {
	switch phaseID {
	case AVMRoadmapPhase0:
		return []AVMImplementationExitCriterion{
			AVMRoadmapExitSignableHashableObjectsHaveTestVectors,
			AVMRoadmapExitQueueOrderingTestCovered,
			AVMRoadmapExitRootEncodingFixed,
		}
	case AVMRoadmapPhase1:
		return []AVMImplementationExitCriterion{
			AVMRoadmapExitExistingCallsRepresentedAsRoutedSync,
			AVMRoadmapExitDeterministicSyncReceipts,
		}
	case AVMRoadmapPhase2:
		return []AVMImplementationExitCriterion{
			AVMRoadmapExitAsyncScheduledLaterBlocks,
			AVMRoadmapExitExpiredMessagesDoNotRun,
			AVMRoadmapExitRetryDeadLetterDeterminism,
		}
	default:
		return nil
	}
}

func requiredAVMTestVectorTargets(phaseID AVMImplementationRoadmapPhaseID) []AVMTestVectorTarget {
	if phaseID != AVMRoadmapPhase0 {
		return nil
	}
	return []AVMTestVectorTarget{
		AVMRoadmapVectorAsyncMessageEncoding,
		AVMRoadmapVectorMessageIDDerivation,
		AVMRoadmapVectorDeterministicQueueSort,
		AVMRoadmapVectorExecutionReceiptSchema,
		AVMRoadmapVectorAVMRootSchema,
		AVMRoadmapVectorGasPolicySchema,
		AVMRoadmapVectorInterfaceDescriptorSchema,
	}
}

func isAVMRoadmapTask(value AVMImplementationRoadmapTask) bool {
	for _, phaseID := range AllAVMImplementationRoadmapPhaseIDs() {
		for _, required := range requiredAVMRoadmapTasks(phaseID) {
			if value == required {
				return true
			}
		}
	}
	return false
}

func isAVMRoadmapExitCriterion(value AVMImplementationExitCriterion) bool {
	for _, phaseID := range AllAVMImplementationRoadmapPhaseIDs() {
		for _, required := range requiredAVMRoadmapExitCriteria(phaseID) {
			if value == required {
				return true
			}
		}
	}
	return false
}

func isAVMTestVectorTarget(value AVMTestVectorTarget) bool {
	for _, required := range requiredAVMTestVectorTargets(AVMRoadmapPhase0) {
		if value == required {
			return true
		}
	}
	return false
}

func validateAVMRoadmapEnumSet[T ~string](label string, values []T, required []T, maxEntries int, allowed func(T) bool) error {
	if len(values) != len(required) || len(values) > maxEntries {
		return fmt.Errorf("AVM roadmap expected %d %s entries", len(required), label)
	}
	seen := make(map[T]struct{}, len(values))
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("AVM roadmap unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("AVM roadmap %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("AVM roadmap duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("AVM roadmap missing %s %s", label, value)
		}
	}
	return nil
}

func validateAVMRoadmapText(fieldName, value string) error {
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxAVMRoadmapTextBytes {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMRoadmapTextBytes)
	}
	for _, r := range value {
		if r < 0x20 || r == '|' {
			return fmt.Errorf("%s contains invalid character", fieldName)
		}
	}
	return nil
}

func sortedAVMRoadmapTasks(values []AVMImplementationRoadmapTask) []AVMImplementationRoadmapTask {
	out := append([]AVMImplementationRoadmapTask(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedAVMRoadmapExitCriteria(values []AVMImplementationExitCriterion) []AVMImplementationExitCriterion {
	out := append([]AVMImplementationExitCriterion(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedAVMTestVectorTargets(values []AVMTestVectorTarget) []AVMTestVectorTarget {
	out := append([]AVMTestVectorTarget(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
