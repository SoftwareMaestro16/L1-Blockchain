package keeper

import (
	"errors"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
)

type GenesisState struct {
	Version uint64
	Params  prototype.Params
	State   meshtypes.MeshState
}

type Keeper struct {
	genesis GenesisState
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version: prototype.CurrentGenesisVersion,
		Params:  prototype.DefaultParams(),
		State:   meshtypes.EmptyState(meshtypes.DefaultParams()),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("mesh prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate()
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k *Keeper) UpdateParams(authority string, params prototype.Params) error {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return err
	}
	k.genesis.Params = params
	return nil
}

func (k *Keeper) RegisterDestination(destination meshtypes.MeshDestination) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := meshtypes.RegisterDestination(k.genesis.State, destination)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) AddFinalizedCommitment(commitment meshtypes.FinalizedCommitment) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := meshtypes.AddFinalizedCommitment(k.genesis.State, commitment)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) ApplyMessage(msg meshtypes.MeshMessage, result meshtypes.ExecutionResult, height uint64) (meshtypes.MeshReceipt, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return meshtypes.MeshReceipt{}, err
	}
	next, receipt, err := meshtypes.ApplyMessage(k.genesis.State, msg, result, height)
	if err != nil {
		return meshtypes.MeshReceipt{}, err
	}
	k.genesis.State = next
	return receipt, nil
}

func (k Keeper) Receipts(req *prototype.PageRequest) ([]meshtypes.MeshReceipt, prototype.PageResponse, error) {
	receipts := k.genesis.State.Export().Receipts
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(receipts))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]meshtypes.MeshReceipt, end-start)
	copy(out, receipts[start:end])
	return out, res, nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	gs := m.keeper.ExportGenesis()
	if err := gs.Validate(); err != nil {
		return err
	}
	return nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
