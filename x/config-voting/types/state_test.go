package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	configtypes "github.com/sovereign-l1/l1/x/config/types"
)

func TestTallyDeterministicWithSortedVotes(t *testing.T) {
	proposal := proposalFixture("p1")
	votes := []ConfigVote{
		{ProposalID: "p1", Voter: "val2", Option: VoteOptionNo, Power: 40, Height: 3},
		{ProposalID: "p1", Voter: "val1", Option: VoteOptionYes, Power: 60, Height: 3},
	}

	tally := Tally(proposal, votes)

	require.Equal(t, uint64(60), tally.YesPower)
	require.Equal(t, uint64(40), tally.NoPower)
	require.True(t, tally.HasQuorum(DefaultConfigVotingParams()))
	require.True(t, tally.HasThreshold(DefaultConfigVotingParams()))
}

func TestProposalSnapshotMustMatchTotalPower(t *testing.T) {
	proposal := proposalFixture("p1")
	proposal.TotalVotingPower = 10

	require.ErrorContains(t, proposal.Validate(DefaultConfigVotingParams()), "total voting power")
}

func TestDeleteProposalCannotCarryValue(t *testing.T) {
	proposal := proposalFixture("p1")
	proposal.Operation = configtypes.OperationDelete
	proposal.ConfigValue = "unexpected"

	require.ErrorContains(t, proposal.Validate(DefaultConfigVotingParams()), "delete")
}

func proposalFixture(id string) ConfigProposal {
	params := DefaultConfigVotingParams()
	return ConfigProposal{
		ProposalID:			id,
		Title:				"raise block gas",
		ConfigKey:			configtypes.KeyConsensusMaxBlockGas,
		ConfigValue:			"1000000",
		Operation:			configtypes.OperationSet,
		SubmittedBy:			"authority",
		Status:				ProposalStatusVoting,
		SnapshotHeight:			1,
		SubmitHeight:			2,
		VotingEndHeight:		2 + params.VotingPeriod,
		EarliestExecutionHeight:	2 + params.VotingPeriod + params.ExecutionDelay,
		VotingPowerSnapshot: []VotingPowerSnapshotEntry{
			{Voter: "val1", Power: 60},
			{Voter: "val2", Power: 40},
		},
		TotalVotingPower:	100,
	}
}
