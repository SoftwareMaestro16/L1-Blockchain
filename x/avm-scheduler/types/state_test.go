package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencyGraphRejectsParallelConflict(t *testing.T) {
	params := DefaultAVMSchedulerParams()
	graph := AVMDependencyGraph{
		BatchID:	"bad",
		Nodes: []AVMDependencyNode{
			{TaskID: "a", ContractAddress: "contract-a", Mailbox: "m-a", WriteSet: []string{"shared"}},
			{TaskID: "b", ContractAddress: "contract-b", Mailbox: "m-b", ReadSet: []string{"shared"}},
		},
		ParallelGroups:	[]AVMExecutionGroup{{GroupIndex: 0, TaskIDs: []string{"a", "b"}}},
	}
	graph.GraphHash = ComputeDependencyGraphHash(graph)

	require.ErrorContains(t, graph.Validate(params), "cannot execute in parallel")
}

func TestMailboxSchedulingSerializesSameMailbox(t *testing.T) {
	params := DefaultAVMSchedulerParams()
	batch := AVMExecutionBatch{
		BatchID:		"mailbox",
		SubmittedHeight:	1,
		Tasks: []AVMExecutionTask{
			testTask("a", "contract-a", "same", nil, []string{"state/a"}),
			testTask("b", "contract-b", "same", nil, []string{"state/b"}),
		},
	}
	graph, err := BuildDependencyGraph(batch, params, false)
	require.NoError(t, err)
	require.Len(t, graph.ParallelGroups, 2)
	require.Equal(t, []string{"a"}, graph.ParallelGroups[0].TaskIDs)
	require.Equal(t, []string{"b"}, graph.ParallelGroups[1].TaskIDs)
}

func TestParallelismLimitBoundsExecutionGroups(t *testing.T) {
	params := DefaultAVMSchedulerParams()
	params.MaxParallelism = 1
	batch := AVMExecutionBatch{
		BatchID:		"bounded",
		SubmittedHeight:	1,
		Tasks: []AVMExecutionTask{
			testTask("a", "contract-a", "m-a", nil, []string{"state/a"}),
			testTask("b", "contract-b", "m-b", nil, []string{"state/b"}),
		},
	}
	graph, err := BuildDependencyGraph(batch, params, false)
	require.NoError(t, err)
	require.Len(t, graph.ParallelGroups, 2)
}

func testTask(id, contract, mailbox string, reads []string, writes []string) AVMExecutionTask {
	stateWrites := make([]AVMStateWrite, 0, len(writes))
	for _, key := range writes {
		stateWrites = append(stateWrites, AVMStateWrite{Key: key, Value: id + "-value"})
	}
	return AVMExecutionTask{
		TaskID:			id,
		ContractAddress:	contract,
		Mailbox:		mailbox,
		ReadSet:		reads,
		WriteSet:		writes,
		StateWrites:		stateWrites,
		GasLimit:		1,
	}
}
