package keeper

import (
	"context"
	"encoding/json"
	"errors"

	corestore "cosmossdk.io/core/store"

	configtypes "github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/constitution/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.State
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
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
		Params:		types.DefaultParams(),
		State: types.State{
			Constitution:		types.DefaultConstitution().Normalize(),
			PendingAmendments:	[]types.Amendment{},
		},
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("constitution unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if k.storeService == nil {
		return errors.New("constitution persistent store is not configured")
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) ProposeConstitutionAmendment(msg types.MsgProposeConstitutionAmendment, height uint64) (types.Amendment, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.Amendment{}, err
	}
	if height == 0 {
		return types.Amendment{}, errors.New("constitution amendment height must be positive")
	}
	if uint32(len(k.genesis.State.PendingAmendments)+1) > k.genesis.Params.MaxPendingAmendments {
		return types.Amendment{}, errors.New("constitution pending amendments limit reached")
	}
	amendment := msg.Amendment.Normalize(k.genesis.Params, msg.Authority, height)
	if _, _, found := types.FindAmendment(k.genesis.State.PendingAmendments, amendment.ID); found {
		return types.Amendment{}, errors.New("constitution amendment already exists")
	}
	if err := amendment.Validate(k.genesis.Params); err != nil {
		return types.Amendment{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.PendingAmendments = types.UpsertAmendment(next.State.PendingAmendments, amendment)
	if err := next.Validate(); err != nil {
		return types.Amendment{}, err
	}
	k.genesis = next
	return amendment, nil
}

func (k *Keeper) VoteConstitutionAmendment(msg types.MsgVoteConstitutionAmendment, height uint64) (types.Amendment, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.Amendment{}, err
	}
	if msg.VotingPowerBps == 0 || msg.VotingPowerBps > types.MaxBasisPoints {
		return types.Amendment{}, errors.New("constitution vote power must be positive and bounded")
	}
	return k.transitionAmendment(msg.AmendmentID, height, func(amendment types.Amendment) (types.Amendment, error) {
		if amendment.Status != types.AmendmentStatusPending && amendment.Status != types.AmendmentStatusApproved {
			return types.Amendment{}, errors.New("constitution amendment cannot be voted in current status")
		}
		switch msg.Support {
		case types.VoteSupportYes:
			amendment.YesVotingPowerBps = boundedAddBps(amendment.YesVotingPowerBps, msg.VotingPowerBps)
			if amendment.YesVotingPowerBps >= k.genesis.Params.MinQuorumBps && amendment.YesVotingPowerBps > amendment.NoVotingPowerBps {
				amendment.Status = types.AmendmentStatusApproved
				amendment.Approver = msg.Authority
			}
		case types.VoteSupportNo:
			amendment.NoVotingPowerBps = boundedAddBps(amendment.NoVotingPowerBps, msg.VotingPowerBps)
		default:
			return types.Amendment{}, errors.New("constitution vote support is invalid")
		}
		return amendment, nil
	})
}

func (k *Keeper) ExecuteConstitutionAmendment(msg types.MsgExecuteConstitutionAmendment, height uint64) (types.Constitution, types.Amendment, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.Constitution{}, types.Amendment{}, err
	}
	idx, amendment, found := types.FindAmendment(k.genesis.State.PendingAmendments, msg.AmendmentID)
	if !found {
		return types.Constitution{}, types.Amendment{}, errors.New("constitution amendment not found")
	}
	if amendment.Status != types.AmendmentStatusApproved {
		return types.Constitution{}, types.Amendment{}, errors.New("constitution amendment must be approved before execution")
	}
	if height < amendment.ExecutableHeight {
		return types.Constitution{}, types.Amendment{}, errors.New("constitution amendment delay has not elapsed")
	}
	if amendment.YesVotingPowerBps < k.genesis.Params.MinQuorumBps || amendment.YesVotingPowerBps <= amendment.NoVotingPowerBps {
		return types.Constitution{}, types.Amendment{}, errors.New("constitution amendment quorum not satisfied")
	}
	next := cloneGenesis(k.genesis)
	amendment.Status = types.AmendmentStatusExecuted
	amendment.Executor = msg.Authority
	amendment.UpdatedHeight = height
	next.State.Constitution = amendment.Proposed.Normalize()
	next.State.PendingAmendments[idx] = amendment
	next.State.PendingAmendments = types.SortedAmendments(next.State.PendingAmendments)
	if err := next.Validate(); err != nil {
		return types.Constitution{}, types.Amendment{}, err
	}
	k.genesis = next
	return next.State.Constitution, amendment, nil
}

func (k *Keeper) CancelConstitutionAmendment(msg types.MsgCancelConstitutionAmendment, height uint64) (types.Amendment, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.Amendment{}, err
	}
	return k.transitionAmendment(msg.AmendmentID, height, func(amendment types.Amendment) (types.Amendment, error) {
		if amendment.Status != types.AmendmentStatusPending && amendment.Status != types.AmendmentStatusApproved {
			return types.Amendment{}, errors.New("constitution amendment cannot be cancelled in current status")
		}
		amendment.Status = types.AmendmentStatusCancelled
		amendment.Canceller = msg.Authority
		amendment.Reason = msg.Reason
		return amendment, nil
	})
}

func (k *Keeper) ActivateEmergencyPause(authority string, currentHeight uint64, durationBlocks uint64) error {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	if currentHeight == 0 {
		return errors.New("constitution emergency pause height must be positive")
	}
	if durationBlocks == 0 || durationBlocks > k.genesis.State.Constitution.EmergencyPauseMaxBlocks || durationBlocks > k.genesis.Params.EmergencyPauseMaxBlocks {
		return errors.New("constitution emergency pause duration exceeds limit")
	}
	next := cloneGenesis(k.genesis)
	next.State.Constitution.EmergencyPauseUntilHeight = currentHeight + durationBlocks
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k *Keeper) ExpireEmergencyPause(currentHeight uint64) bool {
	if k.genesis.State.Constitution.EmergencyPauseUntilHeight == 0 || currentHeight <= k.genesis.State.Constitution.EmergencyPauseUntilHeight {
		return false
	}
	k.genesis.State.Constitution.EmergencyPauseUntilHeight = 0
	return true
}

func (k Keeper) ValidateOrdinaryConfigChange(change configtypes.ConfigChange) error {
	return types.ValidateOrdinaryConfigChange(k.genesis.State.Constitution, change)
}

func (k Keeper) Constitution() types.Constitution {
	return k.genesis.State.Constitution.Normalize()
}

func (k Keeper) PendingAmendments() ([]types.Amendment, error) {
	if err := k.genesis.State.Validate(k.genesis.Params); err != nil {
		return nil, err
	}
	out := make([]types.Amendment, 0, len(k.genesis.State.PendingAmendments))
	for _, amendment := range types.SortedAmendments(k.genesis.State.PendingAmendments) {
		if amendment.Status == types.AmendmentStatusPending || amendment.Status == types.AmendmentStatusApproved {
			out = append(out, amendment)
		}
	}
	return out, nil
}

func (k Keeper) Amendment(id string) (types.Amendment, bool, error) {
	if err := k.genesis.State.Validate(k.genesis.Params); err != nil {
		return types.Amendment{}, false, err
	}
	_, amendment, found := types.FindAmendment(k.genesis.State.PendingAmendments, id)
	return amendment, found, nil
}

func (k Keeper) ProtectedLimits() types.ProtectedLimits {
	return k.genesis.State.Constitution.ProtectedLimits()
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

func (k *Keeper) transitionAmendment(id string, height uint64, mutate func(types.Amendment) (types.Amendment, error)) (types.Amendment, error) {
	if height == 0 {
		return types.Amendment{}, errors.New("constitution amendment height must be positive")
	}
	idx, amendment, found := types.FindAmendment(k.genesis.State.PendingAmendments, id)
	if !found {
		return types.Amendment{}, errors.New("constitution amendment not found")
	}
	updated, err := mutate(amendment)
	if err != nil {
		return types.Amendment{}, err
	}
	updated.UpdatedHeight = height
	if err := updated.Validate(k.genesis.Params); err != nil {
		return types.Amendment{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.PendingAmendments[idx] = updated
	next.State.PendingAmendments = types.SortedAmendments(next.State.PendingAmendments)
	if err := next.Validate(); err != nil {
		return types.Amendment{}, err
	}
	k.genesis = next
	return updated, nil
}

func boundedAddBps(left, right uint32) uint32 {
	sum := uint64(left) + uint64(right)
	if sum > uint64(types.MaxBasisPoints) {
		return types.MaxBasisPoints
	}
	return uint32(sum)
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State.Constitution = gs.State.Constitution.Normalize()
	gs.State.PendingAmendments = types.SortedAmendments(gs.State.PendingAmendments)
	return gs
}
