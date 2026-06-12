package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/sharding-coordinator/types"
)

const authority = prototype.DefaultAuthority

func setupKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func shard(id string) types.Shard {
	return types.Shard{
		ShardID:		id,
		Status:			types.ShardStatusPaused,
		SecurityLevel:		types.ShardSecurityStandard,
		RequiredValidatorCount:	2,
		CrossShardRoutingParams: types.CrossShardRoutingParams{
			AllowInbound:		true,
			AllowOutbound:		true,
			MaxMessageBytes:	1024,
			MaxTimeoutBlocks:	100,
			DefaultRouteLimit:	10,
		},
		RegisteredHeight:	1,
	}
}

func registerShard(t *testing.T, k *Keeper, id string) {
	t.Helper()
	_, err := k.RegisterShard(types.MsgRegisterShard{Authority: authority, Shard: shard(id)})
	require.NoError(t, err)
}

func assign(t *testing.T, k *Keeper, shardID string, validators ...string) {
	t.Helper()
	_, err := k.AssignValidatorsToShard(types.MsgAssignValidatorsToShard{
		Authority:	authority,
		Assignment: types.ShardValidatorAssignment{
			ShardID:		shardID,
			Validators:		validators,
			AssignedHeight:		2,
			AssignmentEpoch:	1,
		},
	})
	require.NoError(t, err)
}

func TestDefaultGenesisDisabled(t *testing.T) {
	gs := DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
}

func TestRegisterShard(t *testing.T) {
	k := setupKeeper(t)
	registered, err := k.RegisterShard(types.MsgRegisterShard{Authority: authority, Shard: shard("shard-a")})
	require.NoError(t, err)
	require.Equal(t, "shard-a", registered.ShardID)
	require.Equal(t, types.ShardStatusPaused, registered.Status)
}

func TestAssignValidators(t *testing.T) {
	k := setupKeeper(t)
	registerShard(t, &k, "shard-a")
	assign(t, &k, "shard-a", "val2", "val1")

	validators, found, err := k.ShardValidators("shard-a")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []string{"val1", "val2"}, validators)
}

func TestSubmitShardLoad(t *testing.T) {
	k := setupKeeper(t)
	registerShard(t, &k, "shard-a")

	load, err := k.SubmitShardLoad(types.MsgSubmitShardLoad{
		Reporter:	"oracle",
		Load: types.ShardLoadMetric{
			ShardID:		"shard-a",
			TransactionsPerBlock:	100,
			GasPerBlock:		1_000,
			StateBytes:		10_000,
			PendingMessages:	5,
			ReportedHeight:		3,
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(3), load.ReportedHeight)
}

func TestRebalanceProposalLifecycle(t *testing.T) {
	k := setupKeeper(t)
	registerShard(t, &k, "shard-a")
	registerShard(t, &k, "shard-b")
	assign(t, &k, "shard-a", "val1", "val2", "val3")
	assign(t, &k, "shard-b", "val4", "val5")
	_, err := k.UpdateShardStatus(types.MsgUpdateShardStatus{Authority: authority, ShardID: "shard-a", Status: types.ShardStatusActive, Height: 3})
	require.NoError(t, err)
	_, err = k.UpdateShardStatus(types.MsgUpdateShardStatus{Authority: authority, ShardID: "shard-b", Status: types.ShardStatusActive, Height: 4})
	require.NoError(t, err)

	proposal, err := k.ProposeShardRebalance(types.MsgProposeShardRebalance{
		Authority:	authority,
		Proposal: types.RebalanceProposal{
			ProposalID:	"rebalance-1",
			SourceShardID:	"shard-a",
			TargetShardID:	"shard-b",
			ValidatorMoves:	[]types.ValidatorMove{{ValidatorID: "val3", FromShardID: "shard-a", ToShardID: "shard-b", Sequence: 1}},
			Reason:		"load skew",
			ProposedHeight:	5,
		},
	})
	require.NoError(t, err)
	require.Equal(t, types.RebalancePending, proposal.Status)

	executed, err := k.ExecuteShardRebalance(types.MsgExecuteShardRebalance{Authority: authority, ProposalID: "rebalance-1", Height: 6})
	require.NoError(t, err)
	require.Equal(t, types.RebalanceExecuted, executed.Status)
	source, found, err := k.ShardValidators("shard-a")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []string{"val1", "val2"}, source)
	target, found, err := k.ShardValidators("shard-b")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []string{"val3", "val4", "val5"}, target)
	_, err = k.ExecuteShardRebalance(types.MsgExecuteShardRebalance{Authority: authority, ProposalID: "rebalance-1", Height: 7})
	require.ErrorContains(t, err, "twice")
}

func TestInsufficientValidatorCoverageRejected(t *testing.T) {
	k := setupKeeper(t)
	registerShard(t, &k, "shard-a")
	assign(t, &k, "shard-a", "val1")

	_, err := k.UpdateShardStatus(types.MsgUpdateShardStatus{Authority: authority, ShardID: "shard-a", Status: types.ShardStatusActive, Height: 3})
	require.ErrorContains(t, err, "insufficient validator coverage")
}

func TestExportImportPreservesAssignments(t *testing.T) {
	k := setupKeeper(t)
	registerShard(t, &k, "shard-b")
	registerShard(t, &k, "shard-a")
	assign(t, &k, "shard-b", "val4", "val3")
	assign(t, &k, "shard-a", "val2", "val1")

	exported := k.ExportGenesis()
	var imported Keeper
	require.NoError(t, imported.InitGenesis(exported))
	shards, err := imported.Shards()
	require.NoError(t, err)
	require.Equal(t, []string{"shard-a", "shard-b"}, []string{shards[0].ShardID, shards[1].ShardID})
	validators, found, err := imported.ShardValidators("shard-a")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []string{"val1", "val2"}, validators)
}
