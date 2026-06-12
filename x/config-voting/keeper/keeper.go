package keeper

import (
	"context"
	"errors"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/config-voting/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	VotingParams	types.ConfigVotingParams
	State		types.ConfigVotingState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
	runtimeCtx	context.Context
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		prototype.DefaultParams(),
		VotingParams:	types.DefaultConfigVotingParams(),
		State:		types.EmptyConfigVotingState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("config voting prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.VotingParams)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, k.genesis)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	gs, _, err := prefixgenesis.Load(ctx, k.storeService, genesisKey, DefaultGenesis())
	if err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) SubmitConfigProposal(msg types.MsgSubmitConfigProposal) (types.ConfigProposal, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ConfigProposal{}, err
	}
	proposal := msg.Proposal.Normalize(k.genesis.VotingParams)
	proposal.SubmittedBy = msg.Authority
	proposal.Status = types.ProposalStatusVoting
	proposal.ExecutedHeight = 0
	proposal.ExecutionConstitutionValidatedAt = 0
	if _, _, found := proposalIndex(k.genesis.State.Proposals, proposal.ProposalID); found {
		return types.ConfigProposal{}, errors.New("config proposal already exists")
	}
	next := cloneGenesis(k.genesis)
	next.State.Proposals = append(next.State.Proposals, proposal)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ConfigProposal{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigProposal{}, err
	}
	return proposal, nil
}

func (k *Keeper) VoteConfigProposal(msg types.MsgVoteConfigProposal) (types.ConfigVote, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ConfigVote{}, err
	}
	idx, proposal, found := proposalIndex(k.genesis.State.Proposals, msg.ProposalID)
	if !found {
		return types.ConfigVote{}, errors.New("config proposal not found")
	}
	if proposal.Status != types.ProposalStatusVoting {
		return types.ConfigVote{}, errors.New("config proposal is not voting")
	}
	if msg.Height < proposal.SubmitHeight || msg.Height > proposal.VotingEndHeight {
		return types.ConfigVote{}, errors.New("config vote height is outside voting period")
	}
	power, found := proposal.SnapshotPower(msg.Voter)
	if !found {
		return types.ConfigVote{}, errors.New("config voter is not in proposal snapshot")
	}
	vote := types.ConfigVote{
		ProposalID:	proposal.ProposalID,
		Voter:		msg.Voter,
		Option:		msg.Option,
		Power:		power,
		Height:		msg.Height,
	}.Normalize()
	if err := vote.Validate(); err != nil {
		return types.ConfigVote{}, err
	}
	next := cloneGenesis(k.genesis)
	if voteIdx, _, found := voteIndex(next.State.Votes, vote.ProposalID, vote.Voter); found {
		next.State.Votes[voteIdx] = vote
	} else {
		next.State.Votes = append(next.State.Votes, vote)
	}
	next.State.Proposals[idx] = proposal
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ConfigVote{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigVote{}, err
	}
	return vote, nil
}

func (k *Keeper) ExecuteConfigProposal(msg types.MsgExecuteConfigProposal) (types.ConfigProposal, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ConfigProposal{}, err
	}
	if msg.Height == 0 {
		return types.ConfigProposal{}, errors.New("config proposal execution height must be positive")
	}
	idx, proposal, found := proposalIndex(k.genesis.State.Proposals, msg.ProposalID)
	if !found {
		return types.ConfigProposal{}, errors.New("config proposal not found")
	}
	if proposal.Status == types.ProposalStatusExecuted {
		return types.ConfigProposal{}, errors.New("config proposal already executed")
	}
	if proposal.Status == types.ProposalStatusVetoed || proposal.Status == types.ProposalStatusRejected {
		return types.ConfigProposal{}, errors.New("config proposal cannot execute from terminal failed status")
	}
	tally := types.Tally(proposal, k.genesis.State.Votes)
	nextStatus := types.ProposalStatusPassed
	if tally.HasVeto(k.genesis.VotingParams) {
		nextStatus = types.ProposalStatusVetoed
	} else if !tally.HasQuorum(k.genesis.VotingParams) || !tally.HasThreshold(k.genesis.VotingParams) {
		nextStatus = types.ProposalStatusRejected
	}
	if nextStatus != types.ProposalStatusPassed {
		proposal.Status = nextStatus
		next := cloneGenesis(k.genesis)
		next.State.Proposals[idx] = proposal.Normalize(next.VotingParams)
		next.State = next.State.Export()
		if err := next.Validate(); err != nil {
			return types.ConfigProposal{}, err
		}
		if err := k.saveGenesis(next); err != nil {
			return types.ConfigProposal{}, err
		}
		return proposal, nil
	}
	if msg.Height < proposal.EarliestExecutionHeight {
		return types.ConfigProposal{}, errors.New("config proposal cannot execute before execution delay")
	}
	if err := validateAgainstConstitution(msg, proposal); err != nil {
		return types.ConfigProposal{}, err
	}
	proposal.Status = types.ProposalStatusExecuted
	proposal.ExecutedHeight = msg.Height
	proposal.ExecutionConstitutionValidatedAt = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Proposals[idx] = proposal.Normalize(next.VotingParams)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ConfigProposal{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigProposal{}, err
	}
	return proposal, nil
}

func (k *Keeper) VetoConfigProposal(msg types.MsgVetoConfigProposal) (types.ConfigProposal, error) {
	if err := k.requireVetoAuthority(msg.Authority); err != nil {
		return types.ConfigProposal{}, err
	}
	if msg.Height == 0 {
		return types.ConfigProposal{}, errors.New("config proposal veto height must be positive")
	}
	idx, proposal, found := proposalIndex(k.genesis.State.Proposals, msg.ProposalID)
	if !found {
		return types.ConfigProposal{}, errors.New("config proposal not found")
	}
	if proposal.Status == types.ProposalStatusExecuted {
		return types.ConfigProposal{}, errors.New("executed config proposal cannot be vetoed")
	}
	proposal.Status = types.ProposalStatusVetoed
	next := cloneGenesis(k.genesis)
	next.State.Proposals[idx] = proposal.Normalize(next.VotingParams)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ConfigProposal{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigProposal{}, err
	}
	return proposal, nil
}

func (k *Keeper) saveGenesis(next GenesisState) error {
	next = cloneGenesis(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return prefixgenesis.Save(k.runtimeCtx, k.storeService, genesisKey, next)
}

func (k Keeper) ConfigProposal(proposalID string) (types.ConfigProposal, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.ConfigProposal{}, false, err
	}
	_, proposal, found := proposalIndex(k.genesis.State.Proposals, proposalID)
	return proposal, found, nil
}

func (k Keeper) ConfigProposals() ([]types.ConfigProposal, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	return k.genesis.State.Export().Proposals, nil
}

func (k Keeper) ConfigVotes(proposalID string) ([]types.ConfigVote, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.ConfigVote, 0)
	for _, vote := range k.genesis.State.Export().Votes {
		if proposalID == "" || vote.ProposalID == proposalID {
			out = append(out, vote)
		}
	}
	types.SortVotes(out)
	return out, nil
}

func (k Keeper) ConfigVotingParams() (types.ConfigVotingParams, error) {
	if err := k.genesis.VotingParams.Validate(); err != nil {
		return types.ConfigVotingParams{}, err
	}
	return k.genesis.VotingParams, nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k Keeper) requireAuthority(authority string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return k.genesis.Params.Authorize(authority)
}

func (k Keeper) requireVetoAuthority(authority string) error {
	if err := k.requireAuthority(authority); err != nil {
		return err
	}
	allowed := k.genesis.VotingParams.VetoAuthorities
	if len(allowed) == 0 {
		return nil
	}
	for _, candidate := range allowed {
		if candidate == authority {
			return nil
		}
	}
	return errors.New("config veto requires veto authority")
}

func validateAgainstConstitution(msg types.MsgExecuteConfigProposal, proposal types.ConfigProposal) error {
	if err := msg.ConfigParams.Validate(); err != nil {
		return err
	}
	if err := msg.ConfigState.Validate(msg.ConfigParams); err != nil {
		return err
	}
	return configChangeAgainstState(msg, proposal)
}

func configChangeAgainstState(msg types.MsgExecuteConfigProposal, proposal types.ConfigProposal) error {
	change := proposal.ToConfigChange(msg.Authority)
	return configtypes.ValidateChangeAgainstState(msg.ConfigParams, msg.ConfigState, change)
}

func proposalIndex(proposals []types.ConfigProposal, proposalID string) (int, types.ConfigProposal, bool) {
	for i, proposal := range proposals {
		if proposal.ProposalID == proposalID {
			return i, proposal, true
		}
	}
	return -1, types.ConfigProposal{}, false
}

func voteIndex(votes []types.ConfigVote, proposalID, voter string) (int, types.ConfigVote, bool) {
	for i, vote := range votes {
		if vote.ProposalID == proposalID && vote.Voter == voter {
			return i, vote, true
		}
	}
	return -1, types.ConfigVote{}, false
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
