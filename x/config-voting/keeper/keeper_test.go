package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	configvotingtypes "github.com/sovereign-l1/l1/x/config-voting/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const authority = prototype.DefaultAuthority

func setupKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	gs.VotingParams.QuorumBps = 5000
	gs.VotingParams.ThresholdBps = 5000
	gs.VotingParams.ExecutionDelay = 5
	gs.VotingParams.VotingPeriod = 10
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func proposal(id string, snapshot []configvotingtypes.VotingPowerSnapshotEntry) configvotingtypes.ConfigProposal {
	return configvotingtypes.ConfigProposal{
		ProposalID:		id,
		Title:			"critical config",
		ConfigKey:		configtypes.KeyConsensusMaxBlockGas,
		ConfigValue:		"1000000",
		Operation:		configtypes.OperationSet,
		SnapshotHeight:		1,
		SubmitHeight:		2,
		VotingPowerSnapshot:	snapshot,
	}
}

func defaultSnapshot() []configvotingtypes.VotingPowerSnapshotEntry {
	return []configvotingtypes.VotingPowerSnapshotEntry{
		{Voter: "val1", Power: 60},
		{Voter: "val2", Power: 40},
	}
}

func submit(t *testing.T, k *Keeper, id string, snapshot []configvotingtypes.VotingPowerSnapshotEntry) configvotingtypes.ConfigProposal {
	t.Helper()
	p, err := k.SubmitConfigProposal(configvotingtypes.MsgSubmitConfigProposal{Authority: authority, Proposal: proposal(id, snapshot)})
	require.NoError(t, err)
	return p
}

func executeMsg(id string, height uint64) configvotingtypes.MsgExecuteConfigProposal {
	return configvotingtypes.MsgExecuteConfigProposal{
		Authority:	authority,
		ProposalID:	id,
		Height:		height,
		ConfigParams:	configtypes.DefaultParams(),
		ConfigState:	configtypes.ConfigState{},
	}
}

func TestDefaultGenesisDisabled(t *testing.T) {
	gs := DefaultGenesis()
	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
}

func TestSubmitVoteExecute(t *testing.T) {
	k := setupKeeper(t)
	p := submit(t, &k, "p1", defaultSnapshot())
	_, err := k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val1", ProposalID: "p1", Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	executed, err := k.ExecuteConfigProposal(executeMsg("p1", p.EarliestExecutionHeight))
	require.NoError(t, err)
	require.Equal(t, configvotingtypes.ProposalStatusExecuted, executed.Status)
	require.Equal(t, p.EarliestExecutionHeight, executed.ExecutionConstitutionValidatedAt)
}

func TestQuorumFailure(t *testing.T) {
	k := setupKeeper(t)
	p := submit(t, &k, "p1", defaultSnapshot())
	_, err := k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val2", ProposalID: "p1", Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	rejected, err := k.ExecuteConfigProposal(executeMsg("p1", p.EarliestExecutionHeight))
	require.NoError(t, err)
	require.Equal(t, configvotingtypes.ProposalStatusRejected, rejected.Status)
}

func TestThresholdFailure(t *testing.T) {
	k := setupKeeper(t)
	p := submit(t, &k, "p1", defaultSnapshot())
	_, err := k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val1", ProposalID: "p1", Option: configvotingtypes.VoteOptionNo, Height: 3})
	require.NoError(t, err)
	_, err = k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val2", ProposalID: "p1", Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	rejected, err := k.ExecuteConfigProposal(executeMsg("p1", p.EarliestExecutionHeight))
	require.NoError(t, err)
	require.Equal(t, configvotingtypes.ProposalStatusRejected, rejected.Status)
}

func TestVetoFlow(t *testing.T) {
	k := setupKeeper(t)
	submit(t, &k, "p1", defaultSnapshot())

	vetoed, err := k.VetoConfigProposal(configvotingtypes.MsgVetoConfigProposal{Authority: authority, ProposalID: "p1", Reason: "unsafe", Height: 3})
	require.NoError(t, err)
	require.Equal(t, configvotingtypes.ProposalStatusVetoed, vetoed.Status)
	_, err = k.ExecuteConfigProposal(executeMsg("p1", vetoed.EarliestExecutionHeight))
	require.ErrorContains(t, err, "terminal")
}

func TestExecutionDelay(t *testing.T) {
	k := setupKeeper(t)
	p := submit(t, &k, "p1", defaultSnapshot())
	_, err := k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val1", ProposalID: "p1", Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	_, err = k.ExecuteConfigProposal(executeMsg("p1", p.EarliestExecutionHeight-1))
	require.ErrorContains(t, err, "delay")
}

func TestProposalCannotExecuteIfItViolatesConstitution(t *testing.T) {
	k := setupKeeper(t)
	unsafe := proposal("p1", defaultSnapshot())
	unsafe.ConfigKey = configtypes.KeyStorageRentPerByteEpoch
	unsafe.ConfigValue = "0"
	_, err := k.SubmitConfigProposal(configvotingtypes.MsgSubmitConfigProposal{Authority: authority, Proposal: unsafe})
	require.NoError(t, err)
	_, err = k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val1", ProposalID: "p1", Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	msg := executeMsg("p1", 17)
	msg.ConfigState = configtypes.ConfigState{Entries: configtypes.SortedEntries([]configtypes.ConfigEntry{
		{Key: configtypes.KeyStorageContractStateActive, Value: "true", Owner: authority, Version: 1, UpdatedHeight: 1},
	})}
	_, err = k.ExecuteConfigProposal(msg)
	require.ErrorContains(t, err, "constitutional")
}

func TestVotingPowerSnapshotPreservedAcrossValidatorSetChange(t *testing.T) {
	k := setupKeeper(t)
	p := submit(t, &k, "p1", []configvotingtypes.VotingPowerSnapshotEntry{{Voter: "val1", Power: 60}, {Voter: "val2", Power: 40}})
	_, err := k.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val1", ProposalID: "p1", Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	submit(t, &k, "p2", []configvotingtypes.VotingPowerSnapshotEntry{{Voter: "val1", Power: 10}, {Voter: "val3", Power: 90}})
	executed, err := k.ExecuteConfigProposal(executeMsg("p1", p.EarliestExecutionHeight))
	require.NoError(t, err)
	require.Equal(t, configvotingtypes.ProposalStatusExecuted, executed.Status)

	votes, err := k.ConfigVotes("p1")
	require.NoError(t, err)
	require.Equal(t, uint64(60), votes[0].Power)
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, source.InitGenesisState(ctx, gs))

	p := submit(t, &source, "persistent-p1", defaultSnapshot())
	_, err := source.VoteConfigProposal(configvotingtypes.MsgVoteConfigProposal{Voter: "val1", ProposalID: p.ProposalID, Option: configvotingtypes.VoteOptionYes, Height: 3})
	require.NoError(t, err)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Proposals, 1)
	require.Len(t, exported.State.Votes, 1)
	require.Equal(t, p.ProposalID, exported.State.Votes[0].ProposalID)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	votes, err := imported.ConfigVotes(p.ProposalID)
	require.NoError(t, err)
	require.Len(t, votes, 1)
	require.Equal(t, uint64(60), votes[0].Power)
}
