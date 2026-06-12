package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	prototype.Params
	State	meshtypes.MeshState
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
		Params:		prototype.DefaultParams(),
		State:		meshtypes.EmptyState(meshtypes.DefaultParams()),
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

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
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
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
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

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
