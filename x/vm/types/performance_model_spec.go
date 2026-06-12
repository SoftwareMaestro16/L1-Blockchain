package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMPerformanceParallelZoneExecution		AVMPerformanceTarget	= "parallel_zone_execution"
	AVMPerformancePipelinedAsyncQueues		AVMPerformanceTarget	= "pipelined_async_queue_processing"
	AVMPerformanceBatchedMessageExecution		AVMPerformanceTarget	= "batched_message_execution"
	AVMPerformanceMinimizedStoreV2Writes		AVMPerformanceTarget	= "minimized_store_v2_writes"
	AVMPerformanceLazyStateLoading			AVMPerformanceTarget	= "lazy_state_loading"
	AVMPerformanceBoundedReceiptGeneration		AVMPerformanceTarget	= "bounded_receipt_generation"
	AVMPerformanceActorLocalConflictIsolation	AVMPerformanceTarget	= "actor_local_conflict_isolation"

	AVMBlockSTMConflictDifferentZones	AVMBlockSTMConflictKeyKind	= "zone"
	AVMBlockSTMConflictActorMailbox		AVMBlockSTMConflictKeyKind	= "actor_mailbox"
	AVMBlockSTMConflictContractStorage	AVMBlockSTMConflictKeyKind	= "contract_storage"
	AVMBlockSTMConflictZoneQueueHead	AVMBlockSTMConflictKeyKind	= "zone_queue_head"
	AVMBlockSTMConflictSenderNonce		AVMBlockSTMConflictKeyKind	= "sender_nonce"
	AVMBlockSTMConflictPaymentEscrow	AVMBlockSTMConflictKeyKind	= "payment_escrow"
	AVMBlockSTMConflictContinuation		AVMBlockSTMConflictKeyKind	= "continuation"
	AVMBlockSTMConflictServiceCall		AVMBlockSTMConflictKeyKind	= "service_call"

	MaxAVMBlockSTMWorkloads	= 4096
)

type AVMPerformanceTarget string
type AVMBlockSTMConflictKeyKind string

type AVMPerformanceModel struct {
	Targets		[]AVMPerformanceTarget
	ModelHash	string
}

type AVMBlockSTMWorkload struct {
	WorkloadID			string
	ZoneID				zonestypes.ZoneID
	ActorIDOptional			string
	ContractAddressOptional		string
	QueueIDOptional			string
	ContinuationIDOptional		string
	ServiceCallOptional		string
	SenderNonceScopeOptional	string
	PaymentEscrowOptional		string
	StorageKeyOptional		string
	ExpectedVersion			uint64
	GasEstimate			uint64
	WritesState			bool
	UsesGlobalCounter		bool
	ConflictKeyKind			AVMBlockSTMConflictKeyKind
	ConflictKey			string
	WorkloadHash			string
}

type AVMBlockSTMPartition struct {
	PartitionID	string
	ZoneID		zonestypes.ZoneID
	ActorID		string
	QueueID		string
	WorkloadIDs	[]string
	PartitionHash	string
}

type AVMZoneExecutionAccumulator struct {
	ZoneID		zonestypes.ZoneID
	MessageCount	uint32
	StoreWriteCount	uint32
	ReceiptCount	uint32
	GasUsed		uint64
	AccumulatorHash	string
}

type AVMBlockSTMExecutionPlan struct {
	Workloads	[]AVMBlockSTMWorkload
	Partitions	[]AVMBlockSTMPartition
	Accumulators	[]AVMZoneExecutionAccumulator
	PlanHash	string
}

func DefaultAVMPerformanceModel() (AVMPerformanceModel, error) {
	model := AVMPerformanceModel{Targets: AllAVMPerformanceTargets()}
	model.ModelHash = ComputeAVMPerformanceModelHash(model)
	return model, model.Validate()
}

func (m AVMPerformanceModel) Validate() error {
	m = canonicalAVMPerformanceModel(m)
	required := AllAVMPerformanceTargets()
	if len(m.Targets) != len(required) {
		return errors.New("AVM performance model must declare every target property")
	}
	for i, target := range m.Targets {
		if !IsAVMPerformanceTarget(target) {
			return fmt.Errorf("invalid AVM performance target %q", target)
		}
		if i > 0 && m.Targets[i-1] >= target {
			return errors.New("AVM performance targets must be sorted canonically")
		}
	}
	if m.ModelHash == "" {
		return errors.New("AVM performance model hash is required")
	}
	if err := zonestypes.ValidateHash("AVM performance model hash", m.ModelHash); err != nil {
		return err
	}
	if m.ModelHash != ComputeAVMPerformanceModelHash(m) {
		return errors.New("AVM performance model hash mismatch")
	}
	return nil
}

func NewAVMBlockSTMWorkload(workload AVMBlockSTMWorkload) (AVMBlockSTMWorkload, error) {
	workload = canonicalAVMBlockSTMWorkload(workload)
	if workload.ConflictKey == "" {
		workload.ConflictKey = AVMBlockSTMConflictKey(workload)
	}
	workload.WorkloadHash = ComputeAVMBlockSTMWorkloadHash(workload)
	return workload, workload.Validate()
}

func (w AVMBlockSTMWorkload) Validate() error {
	w = canonicalAVMBlockSTMWorkload(w)
	if err := validateEngineToken("AVM BlockSTM workload id", w.WorkloadID, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(w.ZoneID); err != nil {
		return err
	}
	if w.ActorIDOptional != "" {
		if err := validateEngineToken("AVM BlockSTM actor id", w.ActorIDOptional, MaxActorRuntimeTokenLength); err != nil {
			return err
		}
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM BlockSTM contract address", value: w.ContractAddressOptional},
		{name: "AVM BlockSTM queue id", value: w.QueueIDOptional},
		{name: "AVM BlockSTM continuation id", value: w.ContinuationIDOptional},
		{name: "AVM BlockSTM service call", value: w.ServiceCallOptional},
		{name: "AVM BlockSTM sender nonce scope", value: w.SenderNonceScopeOptional},
		{name: "AVM BlockSTM payment escrow", value: w.PaymentEscrowOptional},
		{name: "AVM BlockSTM storage key", value: w.StorageKeyOptional},
	} {
		if err := validateRouterOptionalToken(item.name, item.value, MaxAVMStateKeySegmentLength); err != nil {
			return err
		}
	}
	if !IsAVMBlockSTMConflictKeyKind(w.ConflictKeyKind) {
		return fmt.Errorf("invalid AVM BlockSTM conflict key kind %q", w.ConflictKeyKind)
	}
	if w.ConflictKey != AVMBlockSTMConflictKey(w) {
		return errors.New("AVM BlockSTM conflict key mismatch")
	}
	if w.WritesState && w.ExpectedVersion == 0 {
		return errors.New("AVM BlockSTM state updates require expected versions")
	}
	if w.GasEstimate == 0 {
		return errors.New("AVM BlockSTM workload gas estimate must be positive")
	}
	if w.UsesGlobalCounter {
		return errors.New("AVM BlockSTM hot paths must avoid global counters")
	}
	if w.WorkloadHash == "" {
		return errors.New("AVM BlockSTM workload hash is required")
	}
	if err := zonestypes.ValidateHash("AVM BlockSTM workload hash", w.WorkloadHash); err != nil {
		return err
	}
	if w.WorkloadHash != ComputeAVMBlockSTMWorkloadHash(w) {
		return errors.New("AVM BlockSTM workload hash mismatch")
	}
	return nil
}

func NewAVMBlockSTMExecutionPlan(workloads []AVMBlockSTMWorkload) (AVMBlockSTMExecutionPlan, error) {
	if len(workloads) > MaxAVMBlockSTMWorkloads {
		return AVMBlockSTMExecutionPlan{}, fmt.Errorf("AVM BlockSTM workloads must be <= %d", MaxAVMBlockSTMWorkloads)
	}
	plan := AVMBlockSTMExecutionPlan{Workloads: append([]AVMBlockSTMWorkload(nil), workloads...)}
	plan = canonicalAVMBlockSTMExecutionPlan(plan)
	if err := validateAVMBlockSTMWorkloadSet(plan.Workloads); err != nil {
		return AVMBlockSTMExecutionPlan{}, err
	}
	plan.Partitions = PartitionAVMBlockSTMWorkloads(plan.Workloads)
	plan.Accumulators = BuildAVMZoneExecutionAccumulators(plan.Workloads)
	plan = canonicalAVMBlockSTMExecutionPlan(plan)
	plan.PlanHash = ComputeAVMBlockSTMExecutionPlanHash(plan)
	return plan, plan.Validate()
}

func (p AVMBlockSTMExecutionPlan) Validate() error {
	p = canonicalAVMBlockSTMExecutionPlan(p)
	if len(p.Workloads) > MaxAVMBlockSTMWorkloads {
		return fmt.Errorf("AVM BlockSTM workloads must be <= %d", MaxAVMBlockSTMWorkloads)
	}
	if err := validateAVMBlockSTMWorkloadSet(p.Workloads); err != nil {
		return err
	}
	for i, partition := range p.Partitions {
		if err := partition.Validate(); err != nil {
			return err
		}
		if i > 0 && p.Partitions[i-1].PartitionID >= partition.PartitionID {
			return errors.New("AVM BlockSTM partitions must be sorted canonically")
		}
	}
	for i, accumulator := range p.Accumulators {
		if err := accumulator.Validate(); err != nil {
			return err
		}
		if i > 0 && p.Accumulators[i-1].ZoneID >= accumulator.ZoneID {
			return errors.New("AVM per-zone accumulators must be sorted canonically")
		}
	}
	if p.PlanHash == "" {
		return errors.New("AVM BlockSTM execution plan hash is required")
	}
	if err := zonestypes.ValidateHash("AVM BlockSTM execution plan hash", p.PlanHash); err != nil {
		return err
	}
	if p.PlanHash != ComputeAVMBlockSTMExecutionPlanHash(p) {
		return errors.New("AVM BlockSTM execution plan hash mismatch")
	}
	return nil
}

func (p AVMBlockSTMPartition) Validate() error {
	p = canonicalAVMBlockSTMPartition(p)
	if err := validateEngineToken("AVM BlockSTM partition id", p.PartitionID, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(p.ZoneID); err != nil {
		return err
	}
	if err := validateEngineTokens("AVM BlockSTM partition workload id", p.WorkloadIDs, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if p.PartitionHash == "" {
		return errors.New("AVM BlockSTM partition hash is required")
	}
	if err := zonestypes.ValidateHash("AVM BlockSTM partition hash", p.PartitionHash); err != nil {
		return err
	}
	if p.PartitionHash != ComputeAVMBlockSTMPartitionHash(p) {
		return errors.New("AVM BlockSTM partition hash mismatch")
	}
	return nil
}

func (a AVMZoneExecutionAccumulator) Validate() error {
	a = canonicalAVMZoneExecutionAccumulator(a)
	if err := zonestypes.ValidateZoneID(a.ZoneID); err != nil {
		return err
	}
	if a.MessageCount == 0 {
		return errors.New("AVM zone accumulator message count must be positive")
	}
	if a.ReceiptCount > a.MessageCount {
		return errors.New("AVM bounded receipt generation cannot exceed message count")
	}
	if a.AccumulatorHash == "" {
		return errors.New("AVM zone accumulator hash is required")
	}
	if err := zonestypes.ValidateHash("AVM zone accumulator hash", a.AccumulatorHash); err != nil {
		return err
	}
	if a.AccumulatorHash != ComputeAVMZoneExecutionAccumulatorHash(a) {
		return errors.New("AVM zone accumulator hash mismatch")
	}
	return nil
}

func PartitionAVMBlockSTMWorkloads(workloads []AVMBlockSTMWorkload) []AVMBlockSTMPartition {
	grouped := map[string][]string{}
	zoneByPartition := map[string]zonestypes.ZoneID{}
	actorByPartition := map[string]string{}
	queueByPartition := map[string]string{}
	for _, workload := range workloads {
		workload = canonicalAVMBlockSTMWorkload(workload)
		partitionID := AVMBlockSTMPartitionID(workload)
		grouped[partitionID] = append(grouped[partitionID], workload.WorkloadID)
		zoneByPartition[partitionID] = workload.ZoneID
		actorByPartition[partitionID] = workload.ActorIDOptional
		queueByPartition[partitionID] = workload.QueueIDOptional
	}
	partitions := make([]AVMBlockSTMPartition, 0, len(grouped))
	for partitionID, workloadIDs := range grouped {
		sort.Strings(workloadIDs)
		partition := AVMBlockSTMPartition{
			PartitionID:	partitionID,
			ZoneID:		zoneByPartition[partitionID],
			ActorID:	actorByPartition[partitionID],
			QueueID:	queueByPartition[partitionID],
			WorkloadIDs:	workloadIDs,
		}
		partition.PartitionHash = ComputeAVMBlockSTMPartitionHash(partition)
		partitions = append(partitions, partition)
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].PartitionID < partitions[j].PartitionID
	})
	return partitions
}

func BuildAVMZoneExecutionAccumulators(workloads []AVMBlockSTMWorkload) []AVMZoneExecutionAccumulator {
	byZone := map[zonestypes.ZoneID]AVMZoneExecutionAccumulator{}
	for _, workload := range workloads {
		acc := byZone[workload.ZoneID]
		acc.ZoneID = workload.ZoneID
		acc.MessageCount++
		acc.ReceiptCount++
		acc.GasUsed += workload.GasEstimate
		if workload.WritesState {
			acc.StoreWriteCount++
		}
		byZone[workload.ZoneID] = acc
	}
	accumulators := make([]AVMZoneExecutionAccumulator, 0, len(byZone))
	for _, acc := range byZone {
		acc.AccumulatorHash = ComputeAVMZoneExecutionAccumulatorHash(acc)
		accumulators = append(accumulators, acc)
	}
	sort.Slice(accumulators, func(i, j int) bool {
		return accumulators[i].ZoneID < accumulators[j].ZoneID
	})
	return accumulators
}

func AVMBlockSTMPartitionID(workload AVMBlockSTMWorkload) string {
	workload = canonicalAVMBlockSTMWorkload(workload)
	actor := workload.ActorIDOptional
	if actor == "" {
		actor = "_"
	}
	queue := workload.QueueIDOptional
	if queue == "" {
		queue = "_"
	}
	return fmt.Sprintf("%s/%s/%s", workload.ZoneID, actor, queue)
}

func AVMBlockSTMConflictKey(workload AVMBlockSTMWorkload) string {
	workload = canonicalAVMBlockSTMWorkload(workload)
	switch workload.ConflictKeyKind {
	case AVMBlockSTMConflictActorMailbox:
		return fmt.Sprintf("actor-mailbox/%s/%s", workload.ZoneID, workload.ActorIDOptional)
	case AVMBlockSTMConflictContractStorage:
		return fmt.Sprintf("contract-storage/%s/%s/%s", workload.ZoneID, workload.ContractAddressOptional, workload.StorageKeyOptional)
	case AVMBlockSTMConflictZoneQueueHead:
		return fmt.Sprintf("zone-queue-head/%s/%s", workload.ZoneID, workload.QueueIDOptional)
	case AVMBlockSTMConflictSenderNonce:
		return fmt.Sprintf("sender-nonce/%s", workload.SenderNonceScopeOptional)
	case AVMBlockSTMConflictPaymentEscrow:
		return fmt.Sprintf("payment-escrow/%s/%s", workload.ZoneID, workload.PaymentEscrowOptional)
	case AVMBlockSTMConflictContinuation:
		return fmt.Sprintf("continuation/%s/%s", workload.ZoneID, workload.ContinuationIDOptional)
	case AVMBlockSTMConflictServiceCall:
		return fmt.Sprintf("service-call/%s/%s", workload.ZoneID, workload.ServiceCallOptional)
	default:
		return fmt.Sprintf("zone/%s/%s", workload.ZoneID, workload.WorkloadID)
	}
}

func AllAVMPerformanceTargets() []AVMPerformanceTarget {
	items := []AVMPerformanceTarget{
		AVMPerformanceParallelZoneExecution,
		AVMPerformancePipelinedAsyncQueues,
		AVMPerformanceBatchedMessageExecution,
		AVMPerformanceMinimizedStoreV2Writes,
		AVMPerformanceLazyStateLoading,
		AVMPerformanceBoundedReceiptGeneration,
		AVMPerformanceActorLocalConflictIsolation,
	}
	sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })
	return items
}

func IsAVMPerformanceTarget(target AVMPerformanceTarget) bool {
	switch target {
	case AVMPerformanceParallelZoneExecution,
		AVMPerformancePipelinedAsyncQueues,
		AVMPerformanceBatchedMessageExecution,
		AVMPerformanceMinimizedStoreV2Writes,
		AVMPerformanceLazyStateLoading,
		AVMPerformanceBoundedReceiptGeneration,
		AVMPerformanceActorLocalConflictIsolation:
		return true
	default:
		return false
	}
}

func IsAVMBlockSTMConflictKeyKind(kind AVMBlockSTMConflictKeyKind) bool {
	switch kind {
	case AVMBlockSTMConflictDifferentZones,
		AVMBlockSTMConflictActorMailbox,
		AVMBlockSTMConflictContractStorage,
		AVMBlockSTMConflictZoneQueueHead,
		AVMBlockSTMConflictSenderNonce,
		AVMBlockSTMConflictPaymentEscrow,
		AVMBlockSTMConflictContinuation,
		AVMBlockSTMConflictServiceCall:
		return true
	default:
		return false
	}
}

func ComputeAVMPerformanceModelHash(model AVMPerformanceModel) string {
	model = canonicalAVMPerformanceModel(model)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-performance-model-v1")
	writeEngineUint64(h, uint64(len(model.Targets)))
	for _, target := range model.Targets {
		writeEnginePart(h, string(target))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMBlockSTMWorkloadHash(workload AVMBlockSTMWorkload) string {
	workload = canonicalAVMBlockSTMWorkload(workload)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-blockstm-workload-v1")
	writeEnginePart(h, workload.WorkloadID)
	writeEnginePart(h, string(workload.ZoneID))
	writeEnginePart(h, workload.ActorIDOptional)
	writeEnginePart(h, workload.ContractAddressOptional)
	writeEnginePart(h, workload.QueueIDOptional)
	writeEnginePart(h, workload.ContinuationIDOptional)
	writeEnginePart(h, workload.ServiceCallOptional)
	writeEnginePart(h, workload.SenderNonceScopeOptional)
	writeEnginePart(h, workload.PaymentEscrowOptional)
	writeEnginePart(h, workload.StorageKeyOptional)
	writeEngineUint64(h, workload.ExpectedVersion)
	writeEngineUint64(h, workload.GasEstimate)
	writeEngineBool(h, workload.WritesState)
	writeEngineBool(h, workload.UsesGlobalCounter)
	writeEnginePart(h, string(workload.ConflictKeyKind))
	writeEnginePart(h, workload.ConflictKey)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMBlockSTMPartitionHash(partition AVMBlockSTMPartition) string {
	partition = canonicalAVMBlockSTMPartition(partition)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-blockstm-partition-v1")
	writeEnginePart(h, partition.PartitionID)
	writeEnginePart(h, string(partition.ZoneID))
	writeEnginePart(h, partition.ActorID)
	writeEnginePart(h, partition.QueueID)
	writeEngineUint64(h, uint64(len(partition.WorkloadIDs)))
	for _, workloadID := range partition.WorkloadIDs {
		writeEnginePart(h, workloadID)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneExecutionAccumulatorHash(acc AVMZoneExecutionAccumulator) string {
	acc = canonicalAVMZoneExecutionAccumulator(acc)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-execution-accumulator-v1")
	writeEnginePart(h, string(acc.ZoneID))
	writeEngineUint64(h, uint64(acc.MessageCount))
	writeEngineUint64(h, uint64(acc.StoreWriteCount))
	writeEngineUint64(h, uint64(acc.ReceiptCount))
	writeEngineUint64(h, acc.GasUsed)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMBlockSTMExecutionPlanHash(plan AVMBlockSTMExecutionPlan) string {
	plan = canonicalAVMBlockSTMExecutionPlan(plan)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-blockstm-execution-plan-v1")
	writeEngineUint64(h, uint64(len(plan.Workloads)))
	for _, workload := range plan.Workloads {
		writeEnginePart(h, workload.WorkloadHash)
	}
	writeEngineUint64(h, uint64(len(plan.Partitions)))
	for _, partition := range plan.Partitions {
		writeEnginePart(h, partition.PartitionHash)
	}
	writeEngineUint64(h, uint64(len(plan.Accumulators)))
	for _, accumulator := range plan.Accumulators {
		writeEnginePart(h, accumulator.AccumulatorHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func validateAVMBlockSTMWorkloadSet(workloads []AVMBlockSTMWorkload) error {
	seenIDs := map[string]struct{}{}
	conflictKeys := map[string]string{}
	var previous string
	for _, workload := range workloads {
		if err := workload.Validate(); err != nil {
			return err
		}
		if _, found := seenIDs[workload.WorkloadID]; found {
			return fmt.Errorf("duplicate AVM BlockSTM workload %q", workload.WorkloadID)
		}
		if previous != "" && previous >= workload.WorkloadID {
			return errors.New("AVM BlockSTM workloads must be sorted canonically")
		}
		if existing, found := conflictKeys[workload.ConflictKey]; found {
			return fmt.Errorf("AVM BlockSTM conflict key %q overlaps workloads %q and %q", workload.ConflictKey, existing, workload.WorkloadID)
		}
		conflictKeys[workload.ConflictKey] = workload.WorkloadID
		seenIDs[workload.WorkloadID] = struct{}{}
		previous = workload.WorkloadID
	}
	return nil
}

func canonicalAVMPerformanceModel(model AVMPerformanceModel) AVMPerformanceModel {
	model.Targets = append([]AVMPerformanceTarget(nil), model.Targets...)
	sort.Slice(model.Targets, func(i, j int) bool { return model.Targets[i] < model.Targets[j] })
	model.ModelHash = strings.TrimSpace(model.ModelHash)
	return model
}

func canonicalAVMBlockSTMWorkload(workload AVMBlockSTMWorkload) AVMBlockSTMWorkload {
	workload.WorkloadID = strings.TrimSpace(workload.WorkloadID)
	workload.ActorIDOptional = strings.TrimSpace(workload.ActorIDOptional)
	workload.ContractAddressOptional = strings.TrimSpace(workload.ContractAddressOptional)
	workload.QueueIDOptional = strings.TrimSpace(workload.QueueIDOptional)
	workload.ContinuationIDOptional = strings.TrimSpace(workload.ContinuationIDOptional)
	workload.ServiceCallOptional = strings.TrimSpace(workload.ServiceCallOptional)
	workload.SenderNonceScopeOptional = strings.TrimSpace(workload.SenderNonceScopeOptional)
	workload.PaymentEscrowOptional = strings.TrimSpace(workload.PaymentEscrowOptional)
	workload.StorageKeyOptional = strings.TrimSpace(workload.StorageKeyOptional)
	workload.ConflictKey = strings.TrimSpace(workload.ConflictKey)
	workload.WorkloadHash = strings.TrimSpace(workload.WorkloadHash)
	return workload
}

func canonicalAVMBlockSTMPartition(partition AVMBlockSTMPartition) AVMBlockSTMPartition {
	partition.PartitionID = strings.TrimSpace(partition.PartitionID)
	partition.ActorID = strings.TrimSpace(partition.ActorID)
	partition.QueueID = strings.TrimSpace(partition.QueueID)
	partition.WorkloadIDs = append([]string(nil), partition.WorkloadIDs...)
	for i := range partition.WorkloadIDs {
		partition.WorkloadIDs[i] = strings.TrimSpace(partition.WorkloadIDs[i])
	}
	sort.Strings(partition.WorkloadIDs)
	partition.PartitionHash = strings.TrimSpace(partition.PartitionHash)
	return partition
}

func canonicalAVMZoneExecutionAccumulator(acc AVMZoneExecutionAccumulator) AVMZoneExecutionAccumulator {
	acc.AccumulatorHash = strings.TrimSpace(acc.AccumulatorHash)
	return acc
}

func canonicalAVMBlockSTMExecutionPlan(plan AVMBlockSTMExecutionPlan) AVMBlockSTMExecutionPlan {
	plan.Workloads = append([]AVMBlockSTMWorkload(nil), plan.Workloads...)
	for i := range plan.Workloads {
		plan.Workloads[i] = canonicalAVMBlockSTMWorkload(plan.Workloads[i])
	}
	sort.Slice(plan.Workloads, func(i, j int) bool {
		return plan.Workloads[i].WorkloadID < plan.Workloads[j].WorkloadID
	})
	plan.Partitions = append([]AVMBlockSTMPartition(nil), plan.Partitions...)
	for i := range plan.Partitions {
		plan.Partitions[i] = canonicalAVMBlockSTMPartition(plan.Partitions[i])
	}
	sort.Slice(plan.Partitions, func(i, j int) bool {
		return plan.Partitions[i].PartitionID < plan.Partitions[j].PartitionID
	})
	plan.Accumulators = append([]AVMZoneExecutionAccumulator(nil), plan.Accumulators...)
	for i := range plan.Accumulators {
		plan.Accumulators[i] = canonicalAVMZoneExecutionAccumulator(plan.Accumulators[i])
	}
	sort.Slice(plan.Accumulators, func(i, j int) bool {
		return plan.Accumulators[i].ZoneID < plan.Accumulators[j].ZoneID
	})
	plan.PlanHash = strings.TrimSpace(plan.PlanHash)
	return plan
}
