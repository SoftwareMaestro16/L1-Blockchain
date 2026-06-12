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
	BatchStatusQueued	= "queued"
	BatchStatusFinalized	= "finalized"

	ReceiptStatusSuccess	= "success"
	ReceiptStatusFailure	= "failure"

	DefaultInitialStateRoot	= "0000000000000000000000000000000000000000000000000000000000000000"
	MaxAVMSetKeyBytes	= 256
)

type AVMSchedulerParams struct {
	MaxExecutionsPerBlock	uint32
	MaxParallelism		uint32
	MaxQueueDepth		uint32
	MaxReadSetKeys		uint32
	MaxWriteSetKeys		uint32
	MaxReceipts		uint32
}

type AVMSchedulerState struct {
	ExecutionQueue		[]AVMExecutionBatch
	DependencyGraphs	[]AVMDependencyGraph
	ExecutionReceipts	[]AVMExecutionReceipt
	ConflictCounters	[]AVMConflictCounter
	LastFinalizedHeight	uint64
}

type AVMExecutionBatch struct {
	BatchID			string
	SubmittedHeight		uint64
	InitialStateRoot	string
	Status			string
	Tasks			[]AVMExecutionTask
}

type AVMExecutionTask struct {
	TaskID		string
	ContractAddress	string
	Mailbox		string
	ReadSet		[]string
	WriteSet	[]string
	StateWrites	[]AVMStateWrite
	GasLimit	uint64
}

type AVMStateWrite struct {
	Key	string
	Value	string
}

type AVMDependencyGraph struct {
	BatchID		string
	Nodes		[]AVMDependencyNode
	ParallelGroups	[]AVMExecutionGroup
	FallbackSerial	bool
	GraphHash	string
}

type AVMDependencyNode struct {
	TaskID		string
	ContractAddress	string
	Mailbox		string
	ReadSet		[]string
	WriteSet	[]string
	Dependencies	[]string
}

type AVMExecutionGroup struct {
	GroupIndex	uint32
	TaskIDs		[]string
}

type AVMExecutionReceipt struct {
	ReceiptID	string
	BatchID		string
	TaskID		string
	ContractAddress	string
	Height		uint64
	Order		uint32
	Status		string
	GasUsed		uint64
	StateRootBefore	string
	StateRootAfter	string
	Error		string
	FallbackSerial	bool
}

type AVMConflictCounter struct {
	ContractAddress		string
	ConflictCount		uint64
	LastConflictHeight	uint64
}

type MsgSubmitAVMExecutionBatch struct {
	Authority	string
	Batch		AVMExecutionBatch
}

type MsgFinalizeAVMExecutionBatch struct {
	Authority	string
	BatchID		string
	Height		uint64
	ForceSerial	bool
	FailedTaskIDs	[]string
}

type MsgUpdateAVMSchedulerParams struct {
	Authority	string
	SchedulerParams	AVMSchedulerParams
}

func DefaultAVMSchedulerParams() AVMSchedulerParams {
	return AVMSchedulerParams{
		MaxExecutionsPerBlock:	64,
		MaxParallelism:		8,
		MaxQueueDepth:		128,
		MaxReadSetKeys:		64,
		MaxWriteSetKeys:	64,
		MaxReceipts:		1024,
	}
}

func EmptyAVMSchedulerState() AVMSchedulerState {
	return AVMSchedulerState{
		ExecutionQueue:		[]AVMExecutionBatch{},
		DependencyGraphs:	[]AVMDependencyGraph{},
		ExecutionReceipts:	[]AVMExecutionReceipt{},
		ConflictCounters:	[]AVMConflictCounter{},
	}
}

func (p AVMSchedulerParams) Validate() error {
	if p.MaxExecutionsPerBlock == 0 {
		return errors.New("AVM scheduler max executions per block must be positive")
	}
	if p.MaxParallelism == 0 {
		return errors.New("AVM scheduler max parallelism must be positive")
	}
	if p.MaxQueueDepth == 0 {
		return errors.New("AVM scheduler max queue depth must be positive")
	}
	if p.MaxReadSetKeys == 0 || p.MaxWriteSetKeys == 0 {
		return errors.New("AVM scheduler read/write set limits must be positive")
	}
	if p.MaxReceipts == 0 {
		return errors.New("AVM scheduler max receipts must be positive")
	}
	return nil
}

func (s AVMSchedulerState) Export() AVMSchedulerState {
	out := AVMSchedulerState{
		ExecutionQueue:		cloneBatches(s.ExecutionQueue),
		DependencyGraphs:	cloneGraphs(s.DependencyGraphs),
		ExecutionReceipts:	cloneReceipts(s.ExecutionReceipts),
		ConflictCounters:	cloneCounters(s.ConflictCounters),
		LastFinalizedHeight:	s.LastFinalizedHeight,
	}
	SortBatches(out.ExecutionQueue)
	SortGraphs(out.DependencyGraphs)
	SortReceipts(out.ExecutionReceipts)
	SortConflictCounters(out.ConflictCounters)
	return out
}

func (s AVMSchedulerState) Validate(params AVMSchedulerParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.ExecutionQueue)) > params.MaxQueueDepth {
		return errors.New("AVM scheduler queue depth exceeds limit")
	}
	seenBatches := map[string]struct{}{}
	for _, batch := range s.ExecutionQueue {
		if err := batch.Validate(params); err != nil {
			return err
		}
		if _, found := seenBatches[batch.BatchID]; found {
			return fmt.Errorf("duplicate AVM execution batch %q", batch.BatchID)
		}
		seenBatches[batch.BatchID] = struct{}{}
	}
	seenGraphs := map[string]struct{}{}
	for _, graph := range s.DependencyGraphs {
		if err := graph.Validate(params); err != nil {
			return err
		}
		if _, found := seenGraphs[graph.BatchID]; found {
			return fmt.Errorf("duplicate AVM dependency graph %q", graph.BatchID)
		}
		seenGraphs[graph.BatchID] = struct{}{}
	}
	for _, receipt := range s.ExecutionReceipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	for _, counter := range s.ConflictCounters {
		if err := counter.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (b AVMExecutionBatch) Normalize() AVMExecutionBatch {
	out := b
	out.BatchID = strings.TrimSpace(out.BatchID)
	out.InitialStateRoot = strings.TrimSpace(out.InitialStateRoot)
	out.Status = strings.TrimSpace(out.Status)
	if out.Status == "" {
		out.Status = BatchStatusQueued
	}
	if out.InitialStateRoot == "" {
		out.InitialStateRoot = DefaultInitialStateRoot
	}
	out.Tasks = cloneTasks(out.Tasks)
	SortTasks(out.Tasks)
	return out
}

func (b AVMExecutionBatch) Validate(params AVMSchedulerParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	for _, task := range b.Tasks {
		if err := task.Validate(params); err != nil {
			return err
		}
	}
	b = b.Normalize()
	if b.BatchID == "" {
		return errors.New("AVM execution batch id is required")
	}
	if b.SubmittedHeight == 0 {
		return errors.New("AVM execution batch submitted height must be positive")
	}
	if b.Status != BatchStatusQueued && b.Status != BatchStatusFinalized {
		return errors.New("AVM execution batch status is invalid")
	}
	if err := validateHexRoot("AVM execution batch initial state root", b.InitialStateRoot); err != nil {
		return err
	}
	if len(b.Tasks) == 0 {
		return errors.New("AVM execution batch requires tasks")
	}
	if uint32(len(b.Tasks)) > params.MaxExecutionsPerBlock {
		return errors.New("AVM execution batch exceeds per-block execution limit")
	}
	seen := map[string]struct{}{}
	for _, task := range b.Tasks {
		if _, found := seen[task.TaskID]; found {
			return fmt.Errorf("duplicate AVM execution task %q", task.TaskID)
		}
		seen[task.TaskID] = struct{}{}
	}
	return nil
}

func (t AVMExecutionTask) Normalize() AVMExecutionTask {
	out := t
	out.TaskID = strings.TrimSpace(out.TaskID)
	out.ContractAddress = strings.TrimSpace(out.ContractAddress)
	out.Mailbox = strings.TrimSpace(out.Mailbox)
	out.ReadSet = normalizeSet(out.ReadSet)
	out.WriteSet = normalizeSet(out.WriteSet)
	out.StateWrites = cloneStateWrites(out.StateWrites)
	SortStateWrites(out.StateWrites)
	return out
}

func (t AVMExecutionTask) Validate(params AVMSchedulerParams) error {
	if err := validateRawSet("AVM execution read set", t.ReadSet); err != nil {
		return err
	}
	if err := validateRawSet("AVM execution write set", t.WriteSet); err != nil {
		return err
	}
	t = t.Normalize()
	if t.TaskID == "" {
		return errors.New("AVM execution task id is required")
	}
	if t.ContractAddress == "" {
		return errors.New("AVM execution contract address is required")
	}
	if t.Mailbox == "" {
		return errors.New("AVM execution mailbox is required")
	}
	if len(t.ReadSet) == 0 && len(t.WriteSet) == 0 {
		return errors.New("AVM execution task must declare read or write set")
	}
	if uint32(len(t.ReadSet)) > params.MaxReadSetKeys {
		return errors.New("AVM execution read set exceeds limit")
	}
	if uint32(len(t.WriteSet)) > params.MaxWriteSetKeys {
		return errors.New("AVM execution write set exceeds limit")
	}
	if err := validateSet("AVM execution read set", t.ReadSet); err != nil {
		return err
	}
	if err := validateSet("AVM execution write set", t.WriteSet); err != nil {
		return err
	}
	if t.GasLimit == 0 {
		return errors.New("AVM execution gas limit must be positive")
	}
	for _, write := range t.StateWrites {
		if err := write.Validate(); err != nil {
			return err
		}
		if !contains(t.WriteSet, write.Key) {
			return fmt.Errorf("AVM state write key %q must be declared in write set", write.Key)
		}
	}
	return nil
}

func (w AVMStateWrite) Validate() error {
	w.Key = strings.TrimSpace(w.Key)
	if err := validateSet("AVM state write key", []string{w.Key}); err != nil {
		return err
	}
	if strings.TrimSpace(w.Value) != w.Value || w.Value == "" {
		return errors.New("AVM state write value is required and must not have surrounding whitespace")
	}
	return nil
}

func (g AVMDependencyGraph) Normalize() AVMDependencyGraph {
	out := g
	out.BatchID = strings.TrimSpace(out.BatchID)
	out.GraphHash = strings.TrimSpace(out.GraphHash)
	out.Nodes = cloneNodes(out.Nodes)
	out.ParallelGroups = cloneGroups(out.ParallelGroups)
	SortNodes(out.Nodes)
	SortGroups(out.ParallelGroups)
	return out
}

func (g AVMDependencyGraph) Validate(params AVMSchedulerParams) error {
	g = g.Normalize()
	if g.BatchID == "" {
		return errors.New("AVM dependency graph batch id is required")
	}
	if len(g.Nodes) == 0 {
		return errors.New("AVM dependency graph requires nodes")
	}
	if len(g.ParallelGroups) == 0 {
		return errors.New("AVM dependency graph requires execution groups")
	}
	seenNodes := map[string]struct{}{}
	for _, node := range g.Nodes {
		if err := node.Validate(params); err != nil {
			return err
		}
		if _, found := seenNodes[node.TaskID]; found {
			return fmt.Errorf("duplicate AVM dependency node %q", node.TaskID)
		}
		seenNodes[node.TaskID] = struct{}{}
	}
	seenInGroups := map[string]struct{}{}
	for _, group := range g.ParallelGroups {
		if err := group.Validate(); err != nil {
			return err
		}
		if uint32(len(group.TaskIDs)) > params.MaxParallelism {
			return errors.New("AVM execution group exceeds parallelism limit")
		}
		for i, taskID := range group.TaskIDs {
			if _, found := seenNodes[taskID]; !found {
				return fmt.Errorf("AVM execution group references unknown task %q", taskID)
			}
			if _, found := seenInGroups[taskID]; found {
				return fmt.Errorf("duplicate AVM grouped task %q", taskID)
			}
			seenInGroups[taskID] = struct{}{}
			for _, otherID := range group.TaskIDs[i+1:] {
				left := nodeByID(g.Nodes, taskID)
				right := nodeByID(g.Nodes, otherID)
				if NodesConflict(left, right) {
					return errors.New("conflicting AVM write sets cannot execute in parallel")
				}
			}
		}
	}
	if len(seenInGroups) != len(g.Nodes) {
		return errors.New("AVM dependency graph must group every node")
	}
	if g.GraphHash == "" {
		return errors.New("AVM dependency graph hash is required")
	}
	if err := validateHexRoot("AVM dependency graph hash", g.GraphHash); err != nil {
		return err
	}
	if g.GraphHash != ComputeDependencyGraphHash(g) {
		return errors.New("AVM dependency graph hash mismatch")
	}
	return nil
}

func (n AVMDependencyNode) Validate(params AVMSchedulerParams) error {
	task := AVMExecutionTask{
		TaskID:			n.TaskID,
		ContractAddress:	n.ContractAddress,
		Mailbox:		n.Mailbox,
		ReadSet:		n.ReadSet,
		WriteSet:		n.WriteSet,
		GasLimit:		1,
	}
	if err := task.Validate(params); err != nil {
		return err
	}
	if err := validateSet("AVM dependency node dependencies", n.Dependencies); err != nil {
		return err
	}
	return nil
}

func (g AVMExecutionGroup) Validate() error {
	if len(g.TaskIDs) == 0 {
		return errors.New("AVM execution group requires tasks")
	}
	return validateSet("AVM execution group task ids", g.TaskIDs)
}

func (r AVMExecutionReceipt) Validate() error {
	r = r.Normalize()
	if r.ReceiptID == "" || r.BatchID == "" || r.TaskID == "" {
		return errors.New("AVM execution receipt ids are required")
	}
	if r.ContractAddress == "" {
		return errors.New("AVM execution receipt contract address is required")
	}
	if r.Height == 0 {
		return errors.New("AVM execution receipt height must be positive")
	}
	if r.Status != ReceiptStatusSuccess && r.Status != ReceiptStatusFailure {
		return errors.New("AVM execution receipt status is invalid")
	}
	if err := validateHexRoot("AVM receipt state root before", r.StateRootBefore); err != nil {
		return err
	}
	if err := validateHexRoot("AVM receipt state root after", r.StateRootAfter); err != nil {
		return err
	}
	if r.ReceiptID != ComputeReceiptID(r) {
		return errors.New("AVM execution receipt id mismatch")
	}
	return nil
}

func (r AVMExecutionReceipt) Normalize() AVMExecutionReceipt {
	r.ReceiptID = strings.TrimSpace(r.ReceiptID)
	r.BatchID = strings.TrimSpace(r.BatchID)
	r.TaskID = strings.TrimSpace(r.TaskID)
	r.ContractAddress = strings.TrimSpace(r.ContractAddress)
	r.Status = strings.TrimSpace(r.Status)
	r.StateRootBefore = strings.TrimSpace(r.StateRootBefore)
	r.StateRootAfter = strings.TrimSpace(r.StateRootAfter)
	r.Error = strings.TrimSpace(r.Error)
	return r
}

func (c AVMConflictCounter) Validate() error {
	c.ContractAddress = strings.TrimSpace(c.ContractAddress)
	if c.ContractAddress == "" {
		return errors.New("AVM conflict counter contract address is required")
	}
	if c.ConflictCount == 0 {
		return errors.New("AVM conflict counter must be positive")
	}
	if c.LastConflictHeight == 0 {
		return errors.New("AVM conflict counter height must be positive")
	}
	return nil
}

func BuildDependencyGraph(batch AVMExecutionBatch, params AVMSchedulerParams, forceSerial bool) (AVMDependencyGraph, error) {
	if err := batch.Validate(params); err != nil {
		return AVMDependencyGraph{}, err
	}
	batch = batch.Normalize()
	nodes := make([]AVMDependencyNode, len(batch.Tasks))
	for i, task := range batch.Tasks {
		deps := make([]string, 0)
		for j := 0; j < i; j++ {
			previous := batch.Tasks[j]
			if task.Mailbox == previous.Mailbox || TasksConflict(previous, task) {
				deps = append(deps, previous.TaskID)
			}
		}
		nodes[i] = AVMDependencyNode{
			TaskID:			task.TaskID,
			ContractAddress:	task.ContractAddress,
			Mailbox:		task.Mailbox,
			ReadSet:		append([]string(nil), task.ReadSet...),
			WriteSet:		append([]string(nil), task.WriteSet...),
			Dependencies:		normalizeSet(deps),
		}
	}
	graph := AVMDependencyGraph{
		BatchID:	batch.BatchID,
		Nodes:		nodes,
		FallbackSerial:	forceSerial,
	}
	if forceSerial {
		graph.ParallelGroups = serialGroups(batch.Tasks)
	} else {
		groups, fallback := buildParallelGroups(nodes, params.MaxParallelism)
		graph.ParallelGroups = groups
		graph.FallbackSerial = fallback
	}
	graph = graph.Normalize()
	graph.GraphHash = ComputeDependencyGraphHash(graph)
	return graph, graph.Validate(params)
}

func ExecutionOrder(graph AVMDependencyGraph) []string {
	graph = graph.Normalize()
	out := make([]string, 0, len(graph.Nodes))
	for _, group := range graph.ParallelGroups {
		for _, taskID := range group.TaskIDs {
			out = append(out, taskID)
		}
	}
	return out
}

func TasksConflict(a, b AVMExecutionTask) bool {
	return setConflicts(a.WriteSet, b.ReadSet, b.WriteSet) || setConflicts(b.WriteSet, a.ReadSet, a.WriteSet)
}

func NodesConflict(a, b AVMDependencyNode) bool {
	return setConflicts(a.WriteSet, b.ReadSet, b.WriteSet) || setConflicts(b.WriteSet, a.ReadSet, a.WriteSet)
}

func ComputeSerialStateRoot(batch AVMExecutionBatch, failedTaskIDs []string) (string, error) {
	params := DefaultAVMSchedulerParams()
	params.MaxExecutionsPerBlock = uint32(len(batch.Tasks))
	if params.MaxExecutionsPerBlock == 0 {
		params.MaxExecutionsPerBlock = 1
	}
	if err := batch.Validate(params); err != nil {
		return "", err
	}
	failed := stringSet(failedTaskIDs)
	root := batch.Normalize().InitialStateRoot
	for _, task := range batch.Normalize().Tasks {
		if _, isFailed := failed[task.TaskID]; isFailed {
			continue
		}
		root = ApplyTaskStateRoot(root, task)
	}
	return root, nil
}

func ApplyTaskStateRoot(root string, task AVMExecutionTask) string {
	task = task.Normalize()
	h := sha256.New()
	writePart(h, "aetra-avm-state-root-v1")
	writePart(h, root)
	writePart(h, task.TaskID)
	for _, write := range task.StateWrites {
		writePart(h, write.Key)
		writePart(h, write.Value)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeDependencyGraphHash(graph AVMDependencyGraph) string {
	graph = graph.Normalize()
	graph.GraphHash = ""
	h := sha256.New()
	writePart(h, "aetra-avm-dependency-graph-v1")
	writePart(h, graph.BatchID)
	writeBool(h, graph.FallbackSerial)
	for _, node := range graph.Nodes {
		writePart(h, node.TaskID)
		writePart(h, node.ContractAddress)
		writePart(h, node.Mailbox)
		writeStringList(h, node.ReadSet)
		writeStringList(h, node.WriteSet)
		writeStringList(h, node.Dependencies)
	}
	for _, group := range graph.ParallelGroups {
		writeUint64(h, uint64(group.GroupIndex))
		writeStringList(h, group.TaskIDs)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeReceiptID(receipt AVMExecutionReceipt) string {
	receipt = receipt.Normalize()
	receipt.ReceiptID = ""
	h := sha256.New()
	writePart(h, "aetra-avm-execution-receipt-v1")
	writePart(h, receipt.BatchID)
	writePart(h, receipt.TaskID)
	writeUint64(h, uint64(receipt.Order))
	writePart(h, receipt.Status)
	writePart(h, receipt.StateRootBefore)
	writePart(h, receipt.StateRootAfter)
	writeBool(h, receipt.FallbackSerial)
	return hex.EncodeToString(h.Sum(nil))
}

func UpsertConflictCounter(counters []AVMConflictCounter, contract string, height uint64, delta uint64) []AVMConflictCounter {
	contract = strings.TrimSpace(contract)
	out := cloneCounters(counters)
	for i := range out {
		if out[i].ContractAddress == contract {
			out[i].ConflictCount += delta
			out[i].LastConflictHeight = height
			SortConflictCounters(out)
			return out
		}
	}
	out = append(out, AVMConflictCounter{ContractAddress: contract, ConflictCount: delta, LastConflictHeight: height})
	SortConflictCounters(out)
	return out
}

func SortBatches(batches []AVMExecutionBatch) {
	sort.SliceStable(batches, func(i, j int) bool {
		if batches[i].SubmittedHeight != batches[j].SubmittedHeight {
			return batches[i].SubmittedHeight < batches[j].SubmittedHeight
		}
		return batches[i].BatchID < batches[j].BatchID
	})
}

func SortTasks(tasks []AVMExecutionTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		left := tasks[i].Normalize()
		right := tasks[j].Normalize()
		if left.Mailbox != right.Mailbox {
			return left.Mailbox < right.Mailbox
		}
		if left.ContractAddress != right.ContractAddress {
			return left.ContractAddress < right.ContractAddress
		}
		return left.TaskID < right.TaskID
	})
}

func SortStateWrites(writes []AVMStateWrite) {
	sort.SliceStable(writes, func(i, j int) bool {
		if writes[i].Key != writes[j].Key {
			return writes[i].Key < writes[j].Key
		}
		return writes[i].Value < writes[j].Value
	})
}

func SortNodes(nodes []AVMDependencyNode) {
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].TaskID < nodes[j].TaskID })
}

func SortGroups(groups []AVMExecutionGroup) {
	sort.SliceStable(groups, func(i, j int) bool { return groups[i].GroupIndex < groups[j].GroupIndex })
	for i := range groups {
		groups[i].TaskIDs = normalizeSet(groups[i].TaskIDs)
	}
}

func SortGraphs(graphs []AVMDependencyGraph) {
	sort.SliceStable(graphs, func(i, j int) bool { return graphs[i].BatchID < graphs[j].BatchID })
}

func SortReceipts(receipts []AVMExecutionReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool {
		if receipts[i].Height != receipts[j].Height {
			return receipts[i].Height < receipts[j].Height
		}
		if receipts[i].BatchID != receipts[j].BatchID {
			return receipts[i].BatchID < receipts[j].BatchID
		}
		return receipts[i].Order < receipts[j].Order
	})
}

func SortConflictCounters(counters []AVMConflictCounter) {
	sort.SliceStable(counters, func(i, j int) bool { return counters[i].ContractAddress < counters[j].ContractAddress })
}

func buildParallelGroups(nodes []AVMDependencyNode, maxParallelism uint32) ([]AVMExecutionGroup, bool) {
	remaining := map[string]AVMDependencyNode{}
	done := map[string]struct{}{}
	for _, node := range nodes {
		remaining[node.TaskID] = node
	}
	groups := make([]AVMExecutionGroup, 0)
	fallback := false
	for len(remaining) > 0 {
		ready := make([]AVMDependencyNode, 0)
		for _, node := range remaining {
			blocked := false
			for _, dep := range node.Dependencies {
				if _, found := done[dep]; !found {
					blocked = true
					break
				}
			}
			if !blocked {
				ready = append(ready, node)
			}
		}
		SortNodes(ready)
		group := AVMExecutionGroup{GroupIndex: uint32(len(groups))}
		for _, node := range ready {
			if maxParallelism > 0 && uint32(len(group.TaskIDs)) >= maxParallelism {
				continue
			}
			if conflictsWithGroup(node, group, nodes) {
				fallback = true
				continue
			}
			group.TaskIDs = append(group.TaskIDs, node.TaskID)
		}
		if len(group.TaskIDs) == 0 {
			group.TaskIDs = append(group.TaskIDs, ready[0].TaskID)
			fallback = true
		}
		for _, taskID := range group.TaskIDs {
			done[taskID] = struct{}{}
			delete(remaining, taskID)
		}
		groups = append(groups, group)
	}
	return groups, fallback
}

func serialGroups(tasks []AVMExecutionTask) []AVMExecutionGroup {
	groups := make([]AVMExecutionGroup, len(tasks))
	for i, task := range tasks {
		groups[i] = AVMExecutionGroup{GroupIndex: uint32(i), TaskIDs: []string{task.TaskID}}
	}
	return groups
}

func conflictsWithGroup(node AVMDependencyNode, group AVMExecutionGroup, nodes []AVMDependencyNode) bool {
	for _, taskID := range group.TaskIDs {
		if NodesConflict(node, nodeByID(nodes, taskID)) {
			return true
		}
	}
	return false
}

func nodeByID(nodes []AVMDependencyNode, taskID string) AVMDependencyNode {
	for _, node := range nodes {
		if node.TaskID == taskID {
			return node
		}
	}
	return AVMDependencyNode{}
}

func setConflicts(writes []string, otherReads []string, otherWrites []string) bool {
	for _, key := range writes {
		if contains(otherReads, key) || contains(otherWrites, key) {
			return true
		}
	}
	return false
}

func validateSet(name string, values []string) error {
	previous := ""
	for i, value := range values {
		if strings.TrimSpace(value) != value || value == "" {
			return fmt.Errorf("%s contains empty or non-canonical key", name)
		}
		if len(value) > MaxAVMSetKeyBytes {
			return fmt.Errorf("%s key exceeds %d bytes", name, MaxAVMSetKeyBytes)
		}
		if strings.Contains(value, "//") {
			return fmt.Errorf("%s key must not contain empty path segments", name)
		}
		if i > 0 && value <= previous {
			return fmt.Errorf("%s must be sorted and unique", name)
		}
		previous = value
	}
	return nil
}

func validateRawSet(name string, values []string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) != value || value == "" {
			return fmt.Errorf("%s contains empty or non-canonical key", name)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("%s contains duplicate key %q", name, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateHexRoot(name, value string) error {
	if len(value) != 64 {
		return fmt.Errorf("%s must be 32-byte hex", name)
	}
	_, err := hex.DecodeString(value)
	if err != nil {
		return fmt.Errorf("%s must be hex: %w", name, err)
	}
	return nil
}

func normalizeSet(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		out[strings.TrimSpace(value)] = struct{}{}
	}
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneBatches(batches []AVMExecutionBatch) []AVMExecutionBatch {
	out := make([]AVMExecutionBatch, len(batches))
	for i, batch := range batches {
		out[i] = batch.Normalize()
	}
	return out
}

func cloneTasks(tasks []AVMExecutionTask) []AVMExecutionTask {
	out := make([]AVMExecutionTask, len(tasks))
	for i, task := range tasks {
		out[i] = task.Normalize()
	}
	return out
}

func cloneStateWrites(writes []AVMStateWrite) []AVMStateWrite {
	out := append([]AVMStateWrite(nil), writes...)
	for i := range out {
		out[i].Key = strings.TrimSpace(out[i].Key)
		out[i].Value = strings.TrimSpace(out[i].Value)
	}
	return out
}

func cloneGraphs(graphs []AVMDependencyGraph) []AVMDependencyGraph {
	out := make([]AVMDependencyGraph, len(graphs))
	for i, graph := range graphs {
		out[i] = graph.Normalize()
	}
	return out
}

func cloneNodes(nodes []AVMDependencyNode) []AVMDependencyNode {
	out := make([]AVMDependencyNode, len(nodes))
	for i, node := range nodes {
		out[i] = node
		out[i].TaskID = strings.TrimSpace(node.TaskID)
		out[i].ContractAddress = strings.TrimSpace(node.ContractAddress)
		out[i].Mailbox = strings.TrimSpace(node.Mailbox)
		out[i].ReadSet = normalizeSet(node.ReadSet)
		out[i].WriteSet = normalizeSet(node.WriteSet)
		out[i].Dependencies = normalizeSet(node.Dependencies)
	}
	return out
}

func cloneGroups(groups []AVMExecutionGroup) []AVMExecutionGroup {
	out := make([]AVMExecutionGroup, len(groups))
	for i, group := range groups {
		out[i] = group
		out[i].TaskIDs = normalizeSet(group.TaskIDs)
	}
	return out
}

func cloneReceipts(receipts []AVMExecutionReceipt) []AVMExecutionReceipt {
	out := make([]AVMExecutionReceipt, len(receipts))
	for i, receipt := range receipts {
		out[i] = receipt.Normalize()
	}
	return out
}

func cloneCounters(counters []AVMConflictCounter) []AVMConflictCounter {
	out := append([]AVMConflictCounter(nil), counters...)
	for i := range out {
		out[i].ContractAddress = strings.TrimSpace(out[i].ContractAddress)
	}
	return out
}

func writePart(h hashWriter, value string) {
	h.Write([]byte(value))
	h.Write([]byte{0})
}

func writeStringList(h hashWriter, values []string) {
	values = normalizeSet(values)
	writeUint64(h, uint64(len(values)))
	for _, value := range values {
		writePart(h, value)
	}
}

func writeUint64(h hashWriter, value uint64) {
	h.Write([]byte(fmt.Sprintf("%020d", value)))
	h.Write([]byte{0})
}

func writeBool(h hashWriter, value bool) {
	if value {
		h.Write([]byte{1})
		return
	}
	h.Write([]byte{0})
}

type hashWriter interface {
	Write([]byte) (int, error)
}
