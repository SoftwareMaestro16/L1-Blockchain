package types

import (
	"errors"
	"fmt"
	"sort"
)

type CoreImplementationPriority string
type CoreImplementationTaskID string

const (
	CoreTaskPriorityP0	CoreImplementationPriority	= "P0"
	CoreTaskPriorityP1	CoreImplementationPriority	= "P1"
	CoreTaskPriorityP2	CoreImplementationPriority	= "P2"

	CoreTaskZoneRegistry			CoreImplementationTaskID	= "zone-registry"
	CoreTaskZoneDescriptorUpgrade		CoreImplementationTaskID	= "zone-descriptor-upgrade-metadata"
	CoreTaskZoneCommitmentAggregation	CoreImplementationTaskID	= "zone-commitment-aggregation"
	CoreTaskGlobalMessageRoot		CoreImplementationTaskID	= "global-message-root-construction"
	CoreTaskProposalGrouping		CoreImplementationTaskID	= "proposal-grouping-by-zone-and-shard"
	CoreTaskInboundMessageScheduler		CoreImplementationTaskID	= "deterministic-inbound-message-scheduler"
	CoreTaskProofRootRegistry		CoreImplementationTaskID	= "proof-root-registry"
	CoreTaskRootConsistencyInvariants	CoreImplementationTaskID	= "root-consistency-invariants"
	CoreTaskBlockReplayTests		CoreImplementationTaskID	= "block-replay-tests"
	CoreTaskKeeperIntegration		CoreImplementationTaskID	= "keeper-integration"
	CoreTaskOperationalExportImport		CoreImplementationTaskID	= "operational-export-import"
)

type CoreImplementationTaskSpec struct {
	Priority		CoreImplementationPriority
	TaskID			CoreImplementationTaskID
	Task			string
	Target			string
	AcceptanceCriteria	[]string
	TaskHash		string
}

type CoreImplementationEvidence struct {
	ZoneRegistry			bool
	ZoneDescriptorMetadata		bool
	ZoneCommitmentAggregation	bool
	GlobalMessageRoot		bool
	ProposalGrouping		bool
	InboundMessageScheduler		bool
	ProofRootRegistry		bool
	RootConsistencyInvariants	bool
	BlockReplayTests		bool
	KeeperIntegration		bool
	OperationalExportImport		bool
}

type CoreImplementationReadiness struct {
	TaskCount		uint64
	ReadyTaskIDs		[]CoreImplementationTaskID
	MissingTaskIDs		[]CoreImplementationTaskID
	RequiredP0Missing	[]CoreImplementationTaskID
	Ready			bool
	ReadinessHash		string
}

func DefaultCoreImplementationTasks() ([]CoreImplementationTaskSpec, error) {
	tasks := []CoreImplementationTaskSpec{
		coreTask(CoreTaskPriorityP0, CoreTaskZoneRegistry, "Zone registry", "x/aetracore keeper and types", []string{"register-zone-descriptors", "update-zone-descriptors", "disable-zone-descriptors", "export-zone-descriptors", "validate-canonical-ordering", "enforce-version-gates"}),
		coreTask(CoreTaskPriorityP0, CoreTaskZoneDescriptorUpgrade, "Zone descriptor and upgrade metadata", "ZoneDescriptor validation", []string{"reject-duplicate-zones", "reject-disabled-scheduling-targets", "reject-invalid-state-machine-versions", "reject-non-native-fee-policy"}),
		coreTask(CoreTaskPriorityP0, CoreTaskZoneCommitmentAggregation, "Zone commitment aggregation", "ZoneCommitment and GlobalStateRoot", []string{"commit-zone-state-root", "commit-inbox-root", "commit-outbox-root", "commit-receipt-root", "commit-event-root", "commit-shard-root", "commit-params-hash", "commit-execution-summary-hash"}),
		coreTask(CoreTaskPriorityP0, CoreTaskGlobalMessageRoot, "Global message root construction", "GlobalMessageRoot", []string{"commit-inbox-root-by-height", "commit-outbox-root-by-height", "expose-message-proof-root", "support-cross-zone-delivery-eligibility"}),
		coreTask(CoreTaskPriorityP0, CoreTaskProposalGrouping, "Proposal grouping by zone and shard", "ProposalSchedule", []string{"sort-by-zone", "sort-by-shard", "sort-by-priority", "sort-by-admission-height", "sort-by-transaction-hash", "sort-by-message-index"}),
		coreTask(CoreTaskPriorityP0, CoreTaskInboundMessageScheduler, "Deterministic inbound message scheduler", "kernel message pipeline", []string{"deliver-by-committed-priority", "reject-missing-messages", "reject-duplicate-messages", "reject-expired-messages", "reject-reordered-messages"}),
		coreTask(CoreTaskPriorityP0, CoreTaskProofRootRegistry, "Proof root registry", "ProofRoot and root snapshots", []string{"commit-state-roots", "commit-account-roots", "commit-message-roots", "commit-zone-roots", "commit-identity-roots", "commit-resolver-roots", "commit-payment-roots", "commit-vm-roots", "commit-routing-roots", "commit-shard-layout-roots"}),
		coreTask(CoreTaskPriorityP1, CoreTaskRootConsistencyInvariants, "Root consistency invariants", "x/aetracore invariants", []string{"assert-zone-roots", "assert-message-roots", "assert-receipt-roots", "assert-routing-table-roots", "assert-shard-layout-roots", "assert-global-root"}),
		coreTask(CoreTaskPriorityP1, CoreTaskBlockReplayTests, "Block replay tests", "x/aetracore tests", []string{"rebuild-state-from-different-input-orders", "require-identical-zone-commitments", "require-identical-proof-roots", "require-identical-global-app-hash"}),
		coreTask(CoreTaskPriorityP1, CoreTaskKeeperIntegration, "Keeper integration", "x/aetracore/keeper", []string{"persist-descriptors", "persist-layouts", "persist-routing-tables", "persist-commitments", "persist-root-snapshots", "persist-finality-records"}),
		coreTask(CoreTaskPriorityP2, CoreTaskOperationalExportImport, "Operational export/import", "state export manifest", []string{"export-committed-roots", "export-descriptors", "export-layouts", "export-routing-epochs", "export-manifests", "support-reproducible-bootstrap"}),
	}
	for i := range tasks {
		tasks[i] = canonicalCoreImplementationTask(tasks[i])
		tasks[i].TaskHash = ComputeCoreImplementationTaskHash(tasks[i])
		if err := tasks[i].ValidateHash(); err != nil {
			return nil, err
		}
	}
	return tasks, validateCoreImplementationTasks(tasks)
}

func AssessCoreImplementationReadiness(tasks []CoreImplementationTaskSpec, evidence CoreImplementationEvidence) (CoreImplementationReadiness, error) {
	if err := validateCoreImplementationTasks(tasks); err != nil {
		return CoreImplementationReadiness{}, err
	}
	ready := make([]CoreImplementationTaskID, 0, len(tasks))
	missing := make([]CoreImplementationTaskID, 0)
	requiredP0Missing := make([]CoreImplementationTaskID, 0)
	for _, task := range tasks {
		ok := evidence.TaskReady(task.TaskID)
		if ok {
			ready = append(ready, task.TaskID)
			continue
		}
		missing = append(missing, task.TaskID)
		if task.Priority == CoreTaskPriorityP0 {
			requiredP0Missing = append(requiredP0Missing, task.TaskID)
		}
	}
	sortCoreTaskIDs(ready)
	sortCoreTaskIDs(missing)
	sortCoreTaskIDs(requiredP0Missing)
	result := CoreImplementationReadiness{
		TaskCount:		uint64(len(tasks)),
		ReadyTaskIDs:		ready,
		MissingTaskIDs:		missing,
		RequiredP0Missing:	requiredP0Missing,
		Ready:			len(requiredP0Missing) == 0,
	}
	result.ReadinessHash = ComputeCoreImplementationReadinessHash(result)
	return result, nil
}

func DeriveCoreImplementationEvidence(state CoreState, pipeline CoreExecutionPipelineSpec) CoreImplementationEvidence {
	state = state.Export()
	evidence := CoreImplementationEvidence{}
	evidence.ZoneRegistry = len(state.ZoneDescriptors) > 0
	evidence.ZoneDescriptorMetadata = hasValidZoneDescriptorMetadata(state)
	evidence.ZoneCommitmentAggregation = len(state.ZoneCommitments) > 0 && len(state.GlobalRoots) > 0
	evidence.GlobalMessageRoot = hasCommittedMessageRoot(state)
	evidence.ProposalGrouping = pipelineHasPhaseSignal(pipeline, KernelPhasePrepareProposal, "group-transactions-by-zone-and-shard")
	evidence.InboundMessageScheduler = pipelineHasPhaseSignal(pipeline, KernelPhasePrepareProposal, "include-pending-inbound-messages-by-priority") &&
		pipelineHasPhaseSignal(pipeline, KernelPhaseProcessProposal, "wrong-message-delivery-order")
	evidence.ProofRootRegistry = hasProofRootTypes(state, []RootType{MessageProofRootType, ReceiptProofRootType, ShardLayoutRootType})
	evidence.RootConsistencyInvariants = ValidateRootAggregationInvariants(state) == nil
	evidence.BlockReplayTests = len(state.GlobalRoots) > 0 && len(state.RootSnapshots) > 0
	evidence.KeeperIntegration = false
	evidence.OperationalExportImport = len(state.ExportManifests) > 0
	return evidence
}

func (t CoreImplementationTaskSpec) ValidateHash() error {
	t = canonicalCoreImplementationTask(t)
	if err := t.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeCoreImplementationTaskHash(t)
	if t.TaskHash != expected {
		return fmt.Errorf("aetracore implementation task hash mismatch: expected %s", expected)
	}
	return nil
}

func (t CoreImplementationTaskSpec) ValidateFormat() error {
	if !isCoreImplementationPriority(t.Priority) {
		return fmt.Errorf("unknown aetracore implementation task priority %q", t.Priority)
	}
	if err := validatePolicyID("aetracore implementation task id", string(t.TaskID)); err != nil {
		return err
	}
	if err := validateTopologyLabel("aetracore implementation task name", t.Task); err != nil {
		return err
	}
	if err := validateTopologyLabel("aetracore implementation task target", t.Target); err != nil {
		return err
	}
	if len(t.AcceptanceCriteria) == 0 {
		return errors.New("aetracore implementation task requires acceptance criteria")
	}
	if err := validateCapabilitiesForField("aetracore implementation acceptance criterion", t.AcceptanceCriteria); err != nil {
		return err
	}
	if t.TaskHash != "" {
		return ValidateHash("aetracore implementation task hash", t.TaskHash)
	}
	return nil
}

func (e CoreImplementationEvidence) TaskReady(taskID CoreImplementationTaskID) bool {
	switch taskID {
	case CoreTaskZoneRegistry:
		return e.ZoneRegistry
	case CoreTaskZoneDescriptorUpgrade:
		return e.ZoneDescriptorMetadata
	case CoreTaskZoneCommitmentAggregation:
		return e.ZoneCommitmentAggregation
	case CoreTaskGlobalMessageRoot:
		return e.GlobalMessageRoot
	case CoreTaskProposalGrouping:
		return e.ProposalGrouping
	case CoreTaskInboundMessageScheduler:
		return e.InboundMessageScheduler
	case CoreTaskProofRootRegistry:
		return e.ProofRootRegistry
	case CoreTaskRootConsistencyInvariants:
		return e.RootConsistencyInvariants
	case CoreTaskBlockReplayTests:
		return e.BlockReplayTests
	case CoreTaskKeeperIntegration:
		return e.KeeperIntegration
	case CoreTaskOperationalExportImport:
		return e.OperationalExportImport
	default:
		return false
	}
}

func ComputeCoreImplementationTaskHash(task CoreImplementationTaskSpec) string {
	task = canonicalCoreImplementationTask(task)
	parts := []string{"aetra-aek-core-implementation-task-v1", string(task.Priority), string(task.TaskID), task.Task, task.Target}
	parts = appendStringSliceParts(parts, "acceptance", task.AcceptanceCriteria)
	return hashParts(parts...)
}

func ComputeCoreImplementationReadinessHash(result CoreImplementationReadiness) string {
	ready := append([]CoreImplementationTaskID(nil), result.ReadyTaskIDs...)
	missing := append([]CoreImplementationTaskID(nil), result.MissingTaskIDs...)
	p0 := append([]CoreImplementationTaskID(nil), result.RequiredP0Missing...)
	sortCoreTaskIDs(ready)
	sortCoreTaskIDs(missing)
	sortCoreTaskIDs(p0)
	parts := []string{"aetra-aek-core-implementation-readiness-v1", fmt.Sprint(result.TaskCount), fmt.Sprint(result.Ready)}
	appendTaskIDs := func(label string, ids []CoreImplementationTaskID) {
		parts = append(parts, label, fmt.Sprint(len(ids)))
		for _, id := range ids {
			parts = append(parts, string(id))
		}
	}
	appendTaskIDs("ready", ready)
	appendTaskIDs("missing", missing)
	appendTaskIDs("p0-missing", p0)
	return hashParts(parts...)
}

func coreTask(priority CoreImplementationPriority, taskID CoreImplementationTaskID, task string, target string, acceptance []string) CoreImplementationTaskSpec {
	return CoreImplementationTaskSpec{Priority: priority, TaskID: taskID, Task: task, Target: target, AcceptanceCriteria: acceptance}
}

func canonicalCoreImplementationTask(task CoreImplementationTaskSpec) CoreImplementationTaskSpec {
	task.AcceptanceCriteria = append([]string(nil), task.AcceptanceCriteria...)
	sort.Strings(task.AcceptanceCriteria)
	return task
}

func validateCoreImplementationTasks(tasks []CoreImplementationTaskSpec) error {
	if len(tasks) == 0 {
		return errors.New("aetracore implementation tasks are required")
	}
	seen := make(map[CoreImplementationTaskID]struct{}, len(tasks))
	for _, task := range tasks {
		if err := task.ValidateHash(); err != nil {
			return err
		}
		if _, found := seen[task.TaskID]; found {
			return fmt.Errorf("duplicate aetracore implementation task %s", task.TaskID)
		}
		seen[task.TaskID] = struct{}{}
	}
	for _, taskID := range []CoreImplementationTaskID{
		CoreTaskZoneRegistry,
		CoreTaskZoneDescriptorUpgrade,
		CoreTaskZoneCommitmentAggregation,
		CoreTaskGlobalMessageRoot,
		CoreTaskProposalGrouping,
		CoreTaskInboundMessageScheduler,
		CoreTaskProofRootRegistry,
		CoreTaskRootConsistencyInvariants,
		CoreTaskBlockReplayTests,
		CoreTaskKeeperIntegration,
		CoreTaskOperationalExportImport,
	} {
		if _, found := seen[taskID]; !found {
			return fmt.Errorf("aetracore implementation task %s is missing", taskID)
		}
	}
	return nil
}

func isCoreImplementationPriority(priority CoreImplementationPriority) bool {
	switch priority {
	case CoreTaskPriorityP0, CoreTaskPriorityP1, CoreTaskPriorityP2:
		return true
	default:
		return false
	}
}

func sortCoreTaskIDs(ids []CoreImplementationTaskID) {
	sort.SliceStable(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
}

func hasValidZoneDescriptorMetadata(state CoreState) bool {
	for _, descriptor := range state.ZoneDescriptors {
		descriptor = CanonicalZoneDescriptor(descriptor)
		if !descriptor.Enabled || descriptor.StateMachineVersion == 0 || descriptor.FeePolicyID != NativeFeePolicyID || descriptor.ShardLayoutEpoch == 0 {
			return false
		}
	}
	return len(state.ZoneDescriptors) > 0
}

func hasCommittedMessageRoot(state CoreState) bool {
	for _, snapshot := range state.RootSnapshots {
		if snapshot.Finality.GlobalMessageRoot != "" && snapshot.GlobalMessageRoot.RootHash == snapshot.Finality.GlobalMessageRoot {
			return true
		}
	}
	return false
}

func hasProofRootTypes(state CoreState, required []RootType) bool {
	if len(state.RootSnapshots) == 0 {
		return false
	}
	seen := make(map[RootType]struct{})
	for _, snapshot := range state.RootSnapshots {
		if snapshot.GlobalMessageRoot.RootType != "" {
			seen[snapshot.GlobalMessageRoot.RootType] = struct{}{}
		}
		if snapshot.ExecutionReceiptRoot.RootType != "" {
			seen[snapshot.ExecutionReceiptRoot.RootType] = struct{}{}
		}
		for _, root := range snapshot.ProofRoots {
			seen[root.RootType] = struct{}{}
		}
	}
	for _, rootType := range required {
		if _, found := seen[rootType]; !found {
			return false
		}
	}
	return true
}

func pipelineHasPhaseSignal(pipeline CoreExecutionPipelineSpec, phase KernelABCIPhase, signal string) bool {
	if err := pipeline.ValidateHash(); err != nil {
		return false
	}
	for _, spec := range pipeline.Phases {
		if spec.Phase != phase {
			continue
		}
		return containsString(spec.DeterministicWork, signal) || containsString(spec.RejectionChecks, signal) || containsString(spec.CommittedOutput, signal)
	}
	return false
}
