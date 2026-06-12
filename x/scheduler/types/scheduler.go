package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
)

const (
	ModeSequential			= "sequential"
	ModeOptimisticParallel		= "optimistic_parallel"
	ModeDAG				= "dag"
	ModeMailbox			= "mailbox"
	StatusReady			= "ready"
	StatusConflictSequential	= "conflict_fallback_sequential"
	StatusConflictSerialized	= "conflict_serialized"
)

type Task struct {
	ID		string
	Actor		string
	TxHeight	uint64
	TxIndex		uint32
	MessageIndex	uint32
	Dependencies	[]string
	Reads		[]string
	Writes		[]string
	Payload		[]byte
}

type Plan struct {
	Mode		string
	Batches		[]Batch
	Status		string
	ReplayHash	[]byte
}

type Batch struct {
	Tasks []Task
}

type MailboxPlan struct {
	Actor	string
	Tasks	[]Task
}

func PlanSequential(tasks []Task) (Plan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return Plan{}, err
	}
	ordered := cloneTasks(tasks)
	sort.SliceStable(ordered, func(i, j int) bool {
		return taskLess(ordered[i], ordered[j])
	})
	batches := make([]Batch, len(ordered))
	for i, task := range ordered {
		batches[i] = Batch{Tasks: []Task{task}}
	}
	return Plan{Mode: ModeSequential, Batches: batches, Status: StatusReady}, nil
}

func PlanOptimistic(tasks []Task) (Plan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return Plan{}, err
	}
	ordered := cloneTasks(tasks)
	sort.SliceStable(ordered, func(i, j int) bool {
		return taskLess(ordered[i], ordered[j])
	})
	batches := make([]Batch, 0)
	for _, task := range ordered {
		placed := false
		for i := range batches {
			if !ConflictsWithBatch(task, batches[i]) {
				batches[i].Tasks = append(batches[i].Tasks, task)
				placed = true
				break
			}
		}
		if !placed {
			batches = append(batches, Batch{Tasks: []Task{task}})
		}
	}
	status := StatusReady
	if len(batches) == len(ordered) && HasConflicts(ordered) {
		status = StatusConflictSequential
	}
	return Plan{Mode: ModeOptimisticParallel, Batches: batches, Status: status}, nil
}

func BuildDAGPlan(tasks []Task) (Plan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return Plan{}, err
	}
	remaining := taskMap(tasks)
	done := make(map[string]struct{}, len(tasks))
	batches := make([]Batch, 0)
	serializedConflicts := false

	for len(remaining) > 0 {
		ready := readyTasks(remaining, done)
		if len(ready) == 0 {
			return Plan{}, errors.New("scheduler dependency cycle detected")
		}
		sortTasks(ready)
		batch := Batch{}
		for _, task := range ready {
			if ConflictsWithBatch(task, batch) {
				serializedConflicts = true
				continue
			}
			batch.Tasks = append(batch.Tasks, task.Clone())
		}
		if len(batch.Tasks) == 0 {
			task := ready[0]
			batch.Tasks = append(batch.Tasks, task.Clone())
			serializedConflicts = true
		}
		for _, task := range batch.Tasks {
			done[task.ID] = struct{}{}
			delete(remaining, task.ID)
		}
		batches = append(batches, batch)
	}

	status := StatusReady
	if serializedConflicts {
		status = StatusConflictSerialized
	}
	return Plan{Mode: ModeDAG, Batches: batches, Status: status, ReplayHash: ReplayHash(batches)}, nil
}

func BuildMailboxPlan(tasks []Task) ([]MailboxPlan, error) {
	if err := ValidateTasks(tasks); err != nil {
		return nil, err
	}
	byActor := make(map[string][]Task)
	for _, task := range tasks {
		byActor[task.Actor] = append(byActor[task.Actor], task.Clone())
	}
	actors := make([]string, 0, len(byActor))
	for actor := range byActor {
		actors = append(actors, actor)
		sortTasks(byActor[actor])
	}
	sort.Strings(actors)
	out := make([]MailboxPlan, len(actors))
	for i, actor := range actors {
		out[i] = MailboxPlan{Actor: actor, Tasks: cloneTasks(byActor[actor])}
	}
	return out, nil
}

func HasConflicts(tasks []Task) bool {
	for i := range tasks {
		for j := i + 1; j < len(tasks); j++ {
			if Conflicts(tasks[i], tasks[j]) {
				return true
			}
		}
	}
	return false
}

func ConflictsWithBatch(task Task, batch Batch) bool {
	for _, other := range batch.Tasks {
		if Conflicts(task, other) {
			return true
		}
	}
	return false
}

func Conflicts(a, b Task) bool {
	aWrites := set(a.Writes)
	bWrites := set(b.Writes)
	for key := range aWrites {
		if _, ok := bWrites[key]; ok {
			return true
		}
	}
	for _, key := range a.Writes {
		if contains(b.Reads, key) {
			return true
		}
	}
	for _, key := range b.Writes {
		if contains(a.Reads, key) {
			return true
		}
	}
	return false
}

func ValidateTasks(tasks []Task) error {
	seen := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		if task.ID == "" {
			return errors.New("scheduler task id is required")
		}
		if _, ok := seen[task.ID]; ok {
			return fmt.Errorf("duplicate scheduler task id %q", task.ID)
		}
		seen[task.ID] = struct{}{}
		if len(task.Writes) == 0 && len(task.Reads) == 0 {
			return fmt.Errorf("scheduler task %q must declare read or write set", task.ID)
		}
	}
	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if _, ok := seen[dep]; !ok {
				return fmt.Errorf("scheduler task %q dependency %q not found", task.ID, dep)
			}
		}
	}
	return nil
}

func ReplayHash(batches []Batch) []byte {
	h := sha256.New()
	for _, batch := range batches {
		h.Write([]byte("batch"))
		for _, task := range batch.Tasks {
			h.Write([]byte(task.ID))
			h.Write([]byte{0})
		}
	}
	return h.Sum(nil)
}

func (t Task) Clone() Task {
	out := t
	out.Dependencies = append([]string(nil), t.Dependencies...)
	out.Reads = append([]string(nil), t.Reads...)
	out.Writes = append([]string(nil), t.Writes...)
	out.Payload = append([]byte(nil), t.Payload...)
	return out
}

func taskLess(a, b Task) bool {
	if a.TxHeight != b.TxHeight {
		return a.TxHeight < b.TxHeight
	}
	if a.TxIndex != b.TxIndex {
		return a.TxIndex < b.TxIndex
	}
	if a.MessageIndex != b.MessageIndex {
		return a.MessageIndex < b.MessageIndex
	}
	if a.Actor != b.Actor {
		return a.Actor < b.Actor
	}
	return a.ID < b.ID
}

func sortTasks(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		return taskLess(tasks[i], tasks[j])
	})
}

func taskMap(tasks []Task) map[string]Task {
	out := make(map[string]Task, len(tasks))
	for _, task := range tasks {
		out[task.ID] = task.Clone()
	}
	return out
}

func readyTasks(remaining map[string]Task, done map[string]struct{}) []Task {
	ready := make([]Task, 0)
	for _, task := range remaining {
		blocked := false
		for _, dep := range task.Dependencies {
			if _, ok := done[dep]; !ok {
				blocked = true
				break
			}
		}
		if !blocked {
			ready = append(ready, task.Clone())
		}
	}
	return ready
}

func set(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
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

func cloneTasks(tasks []Task) []Task {
	out := make([]Task, len(tasks))
	for i, task := range tasks {
		out[i] = task.Clone()
	}
	return out
}
