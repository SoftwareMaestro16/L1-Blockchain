package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	LoadParams	loadtypes.Params
	EMA		loadtypes.EMAState
	History		[]loadtypes.Result
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
		LoadParams:	loadtypes.DefaultParams(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("load prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.LoadParams.Validate(); err != nil {
		return err
	}
	if err := gs.EMA.Validate(); err != nil {
		return err
	}
	var previous uint64
	for i, result := range gs.History {
		if err := result.EMA.Validate(); err != nil {
			return err
		}
		if result.LoadScoreBps > loadtypes.BasisPoints || result.RawLoadScoreBps > loadtypes.BasisPoints {
			return errors.New("load prototype history score out of bounds")
		}
		if i > 0 && result.EMA.WindowHeight <= previous {
			return errors.New("load prototype history must be sorted by window height")
		}
		previous = result.EMA.WindowHeight
	}
	return nil
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

func (k *Keeper) ApplyMetrics(metrics loadtypes.Metrics) (loadtypes.Result, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return loadtypes.Result{}, err
	}
	result, err := loadtypes.ComputeLoadScore(k.genesis.LoadParams, k.genesis.EMA, metrics)
	if err != nil {
		return loadtypes.Result{}, err
	}
	k.genesis.EMA = result.EMA
	k.genesis.History = append(k.genesis.History, result)
	return result, nil
}

func (k Keeper) History(req *prototype.PageRequest) ([]loadtypes.Result, prototype.PageResponse, error) {
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(k.genesis.History))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return cloneHistory(k.genesis.History[start:end]), res, nil
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
	gs.History = cloneHistory(gs.History)
	return gs
}

func cloneHistory(history []loadtypes.Result) []loadtypes.Result {
	out := make([]loadtypes.Result, len(history))
	copy(out, history)
	return out
}
