package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActiveShardRequiresValidatorCoverage(t *testing.T) {
	state := EmptyShardingCoordinatorState()
	state.Shards = append(state.Shards, testShard("shard-a", ShardStatusActive))

	require.ErrorContains(t, state.Validate(DefaultShardingCoordinatorParams()), "insufficient validator coverage")
}

func TestValidatorAssignmentLimitInvariant(t *testing.T) {
	params := DefaultShardingCoordinatorParams()
	params.MaxShardAssignmentsPerValidator = 1
	state := EmptyShardingCoordinatorState()
	state.Shards = append(state.Shards, testShard("shard-a", ShardStatusPaused), testShard("shard-b", ShardStatusPaused))
	state.Assignments = append(state.Assignments,
		ShardValidatorAssignment{ShardID: "shard-a", Validators: []string{"val1"}, AssignedHeight: 2, AssignmentEpoch: 1},
		ShardValidatorAssignment{ShardID: "shard-b", Validators: []string{"val1"}, AssignedHeight: 2, AssignmentEpoch: 1},
	)

	require.ErrorContains(t, state.Validate(params), "assignment limit")
}

func TestCrossShardRouteRequiredForActivePair(t *testing.T) {
	state := EmptyShardingCoordinatorState()
	state.Shards = append(state.Shards, testShard("shard-a", ShardStatusActive), testShard("shard-b", ShardStatusActive))
	state.Assignments = append(state.Assignments,
		ShardValidatorAssignment{ShardID: "shard-a", Validators: []string{"val1", "val2"}, AssignedHeight: 2, AssignmentEpoch: 1},
		ShardValidatorAssignment{ShardID: "shard-b", Validators: []string{"val3", "val4"}, AssignedHeight: 2, AssignmentEpoch: 1},
	)

	require.ErrorContains(t, state.Validate(DefaultShardingCoordinatorParams()), "route required")
}

func TestExportSortsAssignmentsAndProposalsDeterministically(t *testing.T) {
	state := EmptyShardingCoordinatorState()
	state.Shards = append(state.Shards, testShard("shard-b", ShardStatusPaused), testShard("shard-a", ShardStatusPaused))
	state.Assignments = append(state.Assignments,
		ShardValidatorAssignment{ShardID: "shard-b", Validators: []string{"val2", "val1"}, AssignedHeight: 2, AssignmentEpoch: 1},
		ShardValidatorAssignment{ShardID: "shard-a", Validators: []string{"val4", "val3"}, AssignedHeight: 2, AssignmentEpoch: 1},
	)
	state.RebalanceProposals = append(state.RebalanceProposals,
		RebalanceProposal{ProposalID: "p2", SourceShardID: "shard-b", TargetShardID: "shard-a", ValidatorMoves: []ValidatorMove{{ValidatorID: "val2", FromShardID: "shard-b", ToShardID: "shard-a", Sequence: 2}}, ProposedHeight: 5},
		RebalanceProposal{ProposalID: "p1", SourceShardID: "shard-b", TargetShardID: "shard-a", ValidatorMoves: []ValidatorMove{{ValidatorID: "val1", FromShardID: "shard-b", ToShardID: "shard-a", Sequence: 1}}, ProposedHeight: 4},
	)

	exported := state.Export()
	require.Equal(t, "shard-a", exported.Shards[0].ShardID)
	require.Equal(t, []string{"val1", "val2"}, exported.Assignments[1].Validators)
	require.Equal(t, "p1", exported.RebalanceProposals[0].ProposalID)
}

func TestStateRootHexValidated(t *testing.T) {
	root := ShardStateRootReference{ShardID: "shard-a", Height: 1, RootHex: strings.Repeat("z", 64)}

	require.ErrorContains(t, root.Validate(), "hex")
}

func testShard(id, status string) Shard {
	return Shard{
		ShardID:		id,
		Status:			status,
		SecurityLevel:		ShardSecurityStandard,
		RequiredValidatorCount:	2,
		CrossShardRoutingParams: CrossShardRoutingParams{
			AllowInbound:		true,
			AllowOutbound:		true,
			MaxMessageBytes:	1024,
			MaxTimeoutBlocks:	100,
			DefaultRouteLimit:	10,
		},
		RegisteredHeight:	1,
		UpdatedHeight:		1,
	}
}
