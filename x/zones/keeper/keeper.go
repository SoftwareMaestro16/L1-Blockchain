package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	prototype.Params
	State	zonestypes.ZoneRegistryState
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
		State:		zonestypes.EmptyState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("zones prototype unsupported genesis version")
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

func (k *Keeper) RegisterZone(zone zonestypes.Zone) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := zonestypes.RegisterZone(k.genesis.State, zone)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) ActivateZone(id zonestypes.ZoneID, height uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := zonestypes.ActivateZone(k.genesis.State, id, height)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) AppendCommitment(commitment zonestypes.ZoneCommitment) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := zonestypes.AppendCommitment(k.genesis.State, commitment)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k Keeper) Zones(req *prototype.PageRequest) ([]zonestypes.Zone, prototype.PageResponse, error) {
	zones := k.genesis.State.Export().Zones
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(zones))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]zonestypes.Zone, end-start)
	copy(out, zones[start:end])
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
