package keeper

import (
	"errors"
	"sort"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
)

type ShardConfig struct {
	ZoneID       routingtypes.ZoneID
	ActiveShards uint32
}

type GenesisState struct {
	Version      uint64
	Params       prototype.Params
	RoutingEpoch uint64
	Shards       []ShardConfig
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
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("routing prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	seen := make(map[routingtypes.ZoneID]struct{}, len(gs.Shards))
	for i, shard := range gs.Shards {
		if _, err := routingtypes.ZoneForClass(classForZone(shard.ZoneID)); err != nil {
			return err
		}
		if shard.ActiveShards == 0 {
			return errors.New("routing prototype active shard count must be positive")
		}
		if _, found := seen[shard.ZoneID]; found {
			return errors.New("routing prototype duplicate zone shard config")
		}
		seen[shard.ZoneID] = struct{}{}
		if i > 0 && gs.Shards[i-1].ZoneID >= shard.ZoneID {
			return errors.New("routing prototype shard configs must be sorted canonically")
		}
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

func (k *Keeper) SetRoutingTable(authority string, epoch uint64, shards []ShardConfig) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	next.RoutingEpoch = epoch
	next.Shards = cloneShards(shards)
	normalizeShardConfigs(next.Shards)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) Route(input routingtypes.RouteInput) (routingtypes.RouteDecision, error) {
	if input.ActiveShards == nil {
		input.ActiveShards = k.activeShardMap()
	}
	if input.RoutingEpoch == 0 {
		input.RoutingEpoch = k.genesis.RoutingEpoch
	}
	return routingtypes.Route(input)
}

func (k Keeper) Shards(req *prototype.PageRequest) ([]ShardConfig, prototype.PageResponse, error) {
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(k.genesis.Shards))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return cloneShards(k.genesis.Shards[start:end]), res, nil
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

func (k Keeper) activeShardMap() map[routingtypes.ZoneID]uint32 {
	out := make(map[routingtypes.ZoneID]uint32, len(k.genesis.Shards))
	for _, shard := range k.genesis.Shards {
		out[shard.ZoneID] = shard.ActiveShards
	}
	return out
}

func classForZone(zone routingtypes.ZoneID) routingtypes.TxClass {
	switch zone {
	case routingtypes.ZoneFinancial:
		return routingtypes.TxClassFinancial
	case routingtypes.ZoneIdentity:
		return routingtypes.TxClassIdentity
	case routingtypes.ZoneContract:
		return routingtypes.TxClassContract
	case routingtypes.ZoneApplication:
		return routingtypes.TxClassApplication
	case routingtypes.ZoneAetherCore:
		return routingtypes.TxClassCriticalSystem
	default:
		return ""
	}
}

func normalizeShardConfigs(shards []ShardConfig) {
	sort.SliceStable(shards, func(i, j int) bool {
		return shards[i].ZoneID < shards[j].ZoneID
	})
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.Shards = cloneShards(gs.Shards)
	return gs
}

func cloneShards(shards []ShardConfig) []ShardConfig {
	out := make([]ShardConfig, len(shards))
	copy(out, shards)
	return out
}
