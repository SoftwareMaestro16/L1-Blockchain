package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	prototype.Params
	State	networkingtypes.NetworkingState
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
		State:		networkingtypes.EmptyState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("networking prototype unsupported genesis version")
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

func (k *Keeper) RegisterNodeRecord(record networkingtypes.NodeRecord, networkSalt []byte, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := networkingtypes.RegisterNodeRecord(k.genesis.State, record, networkSalt, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) ApplyIdentityTransition(transition networkingtypes.IdentityTransitionRecord, newRecord networkingtypes.NodeRecord, networkSalt []byte, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := networkingtypes.ApplyIdentityTransition(k.genesis.State, transition, newRecord, networkSalt, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) OpenSession(session networkingtypes.SessionChannel, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := networkingtypes.OpenSession(k.genesis.State, session, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) RegisterRoleCommitment(commitment networkingtypes.RoleCommitment, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := networkingtypes.RegisterRoleCommitment(k.genesis.State, commitment, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) RegisterOverlayDescriptor(desc networkingtypes.OverlayDescriptor, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := networkingtypes.RegisterOverlayDescriptor(k.genesis.State, desc, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) PruneExpired(currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := networkingtypes.PruneExpired(k.genesis.State, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k Keeper) NodeRecords(req *prototype.PageRequest) ([]networkingtypes.NodeRecord, prototype.PageResponse, error) {
	records := k.genesis.State.Export().NodeRecords
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(records))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]networkingtypes.NodeRecord, end-start)
	copy(out, records[start:end])
	return out, res, nil
}

func (k Keeper) Sessions(req *prototype.PageRequest) ([]networkingtypes.SessionChannel, prototype.PageResponse, error) {
	sessions := k.genesis.State.Export().Sessions
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(sessions))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]networkingtypes.SessionChannel, end-start)
	copy(out, sessions[start:end])
	return out, res, nil
}

func (k Keeper) OverlayDescriptors(req *prototype.PageRequest) ([]networkingtypes.OverlayDescriptor, prototype.PageResponse, error) {
	descriptors := k.genesis.State.Export().OverlayDescriptors
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(descriptors))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]networkingtypes.OverlayDescriptor, end-start)
	copy(out, descriptors[start:end])
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
