package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSequentialPlanIsDeterministic(t *testing.T) {
	plan, err := PlanSequential([]Task{
		task("b", "", 1, 2, 0, nil, nil, []string{"b"}),
		task("a", "", 1, 1, 0, nil, nil, []string{"a"}),
	})
	require.NoError(t, err)
	require.Equal(t, ModeSequential, plan.Mode)
	require.Len(t, plan.Batches, 2)
	require.Equal(t, "a", plan.Batches[0].Tasks[0].ID)
	require.Equal(t, "b", plan.Batches[1].Tasks[0].ID)
}

func TestConflictDetectionAndOptimisticBatching(t *testing.T) {
	independentA := task("a", "", 1, 1, 0, nil, []string{"x"}, []string{"y"})
	independentB := task("b", "", 1, 2, 0, nil, []string{"m"}, []string{"n"})
	conflict := task("c", "", 1, 3, 0, nil, []string{"y"}, []string{"z"})

	require.False(t, Conflicts(independentA, independentB))
	require.True(t, Conflicts(independentA, conflict))

	plan, err := PlanOptimistic([]Task{conflict, independentB, independentA})
	require.NoError(t, err)
	require.Equal(t, ModeOptimisticParallel, plan.Mode)
	require.Len(t, plan.Batches, 2)
	require.Len(t, plan.Batches[0].Tasks, 2)
	require.Len(t, plan.Batches[1].Tasks, 1)
}

func TestSchedulerValidation(t *testing.T) {
	_, err := PlanSequential([]Task{{ID: "empty"}})
	require.ErrorContains(t, err, "read or write")
	_, err = PlanSequential([]Task{
		task("a", "", 1, 0, 0, nil, nil, []string{"x"}),
		task("a", "", 1, 1, 0, nil, nil, []string{"y"}),
	})
	require.ErrorContains(t, err, "duplicate")
}

func TestDAGPlanRespectsDependencies(t *testing.T) {
	a := task("a", "", 1, 1, 0, nil, []string{"x"}, []string{"y"})
	b := task("b", "", 1, 2, 0, []string{"a"}, []string{"y"}, []string{"z"})
	c := task("c", "", 1, 3, 0, []string{"a"}, nil, []string{"w"})

	plan, err := BuildDAGPlan([]Task{b, c, a})
	require.NoError(t, err)
	require.Equal(t, ModeDAG, plan.Mode)
	require.Equal(t, StatusReady, plan.Status)
	require.GreaterOrEqual(t, len(plan.Batches), 2)
	require.Contains(t, plan.Batches[0].Tasks[0].ID, "a")

	ordered := make(map[string]int)
	for i, batch := range plan.Batches {
		for _, task := range batch.Tasks {
			ordered[task.ID] = i
		}
	}
	for _, task := range plan.Batches[0].Tasks {
		require.Equal(t, "a", task.ID)
	}
	require.Less(t, ordered["a"], ordered["b"])
	require.Less(t, ordered["a"], ordered["c"])
}

func TestDAGPlanCycleDetection(t *testing.T) {
	a := task("a", "", 1, 1, 0, []string{"b"}, []string{"x"}, []string{"y"})
	b := task("b", "", 1, 2, 0, []string{"a"}, []string{"y"}, []string{"z"})

	_, err := BuildDAGPlan([]Task{a, b})
	require.ErrorContains(t, err, "dependency cycle")
}

func TestMailboxPlanGroupsByActor(t *testing.T) {
	alice1 := task("t1", "alice", 1, 1, 0, nil, []string{"x"}, []string{"y"})
	bob1 := task("t2", "bob", 1, 2, 0, nil, []string{"m"}, []string{"n"})
	alice2 := task("t3", "alice", 1, 3, 0, nil, []string{"y"}, []string{"z"})
	alice3 := task("t4", "alice", 1, 4, 0, nil, []string{"q"}, []string{"r"})

	plans, err := BuildMailboxPlan([]Task{alice1, bob1, alice2, alice3})
	require.NoError(t, err)
	require.Len(t, plans, 2)

	for _, mp := range plans {
		require.True(t, mp.Actor == "alice" || mp.Actor == "bob")
		if mp.Actor == "alice" {
			require.Len(t, mp.Tasks, 3)
		} else {
			require.Len(t, mp.Tasks, 1)
		}
	}
}

func TestReplayHashIsDeterministic(t *testing.T) {
	tasks := []Task{
		task("a", "", 1, 1, 0, nil, nil, []string{"x"}),
		task("b", "", 1, 2, 0, nil, []string{"x"}, []string{"y"}),
	}
	plan1, err := BuildDAGPlan(tasks)
	require.NoError(t, err)
	plan2, err := BuildDAGPlan(tasks)
	require.NoError(t, err)
	require.Equal(t, plan1.ReplayHash, plan2.ReplayHash)
}

func TestDAGPlanWithConflicts(t *testing.T) {
	a := task("a", "", 1, 1, 0, nil, nil, []string{"x"})
	b := task("b", "", 1, 2, 0, nil, nil, []string{"x"})

	plan, err := BuildDAGPlan([]Task{b, a})
	require.NoError(t, err)
	require.Equal(t, ModeDAG, plan.Mode)
	require.Len(t, plan.Batches, 2)
	require.Equal(t, StatusConflictSerialized, plan.Status)
}

func task(id, actor string, height uint64, tx uint32, msg uint32, deps, reads, writes []string) Task {
	return Task{
		ID:	id, Actor: actor, TxHeight: height, TxIndex: tx,
		MessageIndex:	msg, Dependencies: deps, Reads: reads, Writes: writes,
	}
}
