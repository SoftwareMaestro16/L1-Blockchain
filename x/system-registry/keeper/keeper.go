package keeper

import (
	"context"
	"errors"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/system-registry/types"
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
	runtimeCtx	context.Context
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		params,
		State:		types.DefaultState().Normalize(params),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("system registry unsupported genesis version")
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

func (k *Keeper) RegisterSystemEntity(msg types.MsgRegisterSystemEntity) (types.SystemEntity, types.SystemEntityEvent, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	entity := msg.Entity.Normalize(k.genesis.Params)
	if _, found := k.genesis.State.Entity(entity.ModuleName); found {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry entity already registered")
	}
	if err := entity.Validate(k.genesis.Params); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Entities = types.UpsertEntity(next.State.Entities, entity)
	if err := next.Validate(); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	return entity, event(types.EventTypeRegistered, entity, 0), nil
}

func (k *Keeper) UpdateSystemEntity(msg types.MsgUpdateSystemEntity) (types.SystemEntity, types.SystemEntityEvent, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	existing, found := k.genesis.State.Entity(msg.Entity.ModuleName)
	if !found {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry entity not found")
	}
	updated := msg.Entity.Normalize(k.genesis.Params)
	if existing.Required {
		updated.Required = true
	}
	if existing.Required && updated.Status != types.StatusActive {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry required module cannot be made inactive")
	}
	if existing.Required && updated.ModuleAccountAddress != existing.ModuleAccountAddress {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry required module account cannot be changed")
	}
	if existing.Required && (updated.RawAddress != existing.RawAddress || updated.UserFriendlyAddress != existing.UserFriendlyAddress) {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry required address mapping cannot be changed without constitutional upgrade")
	}
	if err := updated.Validate(k.genesis.Params); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Entities = types.UpsertEntity(next.State.Entities, updated)
	if err := next.Validate(); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	return updated, event(types.EventTypeUpdated, updated, 0), nil
}

func (k *Keeper) PauseSystemEntity(msg types.MsgPauseSystemEntity) (types.SystemEntity, types.SystemEntityEvent, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	entity, found := k.genesis.State.Entity(msg.ModuleName)
	if !found {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry entity not found")
	}
	if entity.Required {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry required module cannot be paused")
	}
	if entity.Status == types.StatusDeprecated {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry deprecated module cannot be paused")
	}
	entity.Status = types.StatusPaused
	entity.PrivilegedCallsAllowedWhilePaused = msg.AllowPrivilegedCallsWhilePaused
	return k.transition(entity, types.EventTypePaused, msg.Height)
}

func (k *Keeper) ResumeSystemEntity(msg types.MsgResumeSystemEntity) (types.SystemEntity, types.SystemEntityEvent, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	entity, found := k.genesis.State.Entity(msg.ModuleName)
	if !found {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry entity not found")
	}
	if entity.Status != types.StatusPaused {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry module is not paused")
	}
	entity.Status = types.StatusActive
	entity.PrivilegedCallsAllowedWhilePaused = false
	return k.transition(entity, types.EventTypeResumed, msg.Height)
}

func (k *Keeper) DeprecateSystemEntity(msg types.MsgDeprecateSystemEntity) (types.SystemEntity, types.SystemEntityEvent, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	entity, found := k.genesis.State.Entity(msg.ModuleName)
	if !found {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry entity not found")
	}
	if entity.Required {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry required module cannot be deprecated")
	}
	entity.Status = types.StatusDeprecated
	entity.PrivilegedCallsAllowedWhilePaused = false
	return k.transition(entity, types.EventTypeDeprecated, msg.Height)
}

func (k Keeper) SystemEntity(moduleName string) (types.SystemEntity, bool, error) {
	if err := k.genesis.Validate(); err != nil {
		return types.SystemEntity{}, false, err
	}
	entity, found := k.genesis.State.Entity(moduleName)
	return entity, found, nil
}

func (k Keeper) SystemEntities() ([]types.SystemEntity, error) {
	if err := k.genesis.Validate(); err != nil {
		return nil, err
	}
	return types.SortEntities(k.genesis.State.Entities), nil
}

func (k Keeper) ModuleAccount(moduleName string) (string, bool, error) {
	entity, found, err := k.SystemEntity(moduleName)
	if err != nil || !found {
		return "", found, err
	}
	return entity.ModuleAccountAddress, true, nil
}

func (k Keeper) Capabilities(moduleName string) ([]string, bool, error) {
	entity, found, err := k.SystemEntity(moduleName)
	if err != nil || !found {
		return nil, found, err
	}
	out := make([]string, len(entity.Capabilities))
	copy(out, entity.Capabilities)
	return out, true, nil
}

func (k Keeper) DependencyGraph() ([]types.DependencyEdge, error) {
	if err := k.genesis.Validate(); err != nil {
		return nil, err
	}
	return k.genesis.State.DependencyGraph(), nil
}

func (k Keeper) CanReceivePrivilegedCall(moduleName string) (bool, error) {
	entity, found, err := k.SystemEntity(moduleName)
	if err != nil {
		return false, err
	}
	if !found {
		return false, errors.New("system registry entity not found")
	}
	return types.PrivilegedCallAllowed(entity), nil
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

func (k *Keeper) transition(entity types.SystemEntity, eventType string, height uint64) (types.SystemEntity, types.SystemEntityEvent, error) {
	if height == 0 {
		return types.SystemEntity{}, types.SystemEntityEvent{}, errors.New("system registry event height must be positive")
	}
	entity = entity.Normalize(k.genesis.Params)
	if err := entity.Validate(k.genesis.Params); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Entities = types.UpsertEntity(next.State.Entities, entity)
	if err := next.Validate(); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.SystemEntity{}, types.SystemEntityEvent{}, err
	}
	return entity, event(eventType, entity, height), nil
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

func event(eventType string, entity types.SystemEntity, height uint64) types.SystemEntityEvent {
	return types.SystemEntityEvent{
		Type:		eventType,
		ModuleName:	entity.ModuleName,
		Status:		entity.Status,
		Height:		height,
	}
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.Params.RequiredModules = append([]string(nil), gs.Params.RequiredModules...)
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}
