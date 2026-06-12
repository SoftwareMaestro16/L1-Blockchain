package keeper

import (
	"context"
	"errors"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/actor-registry/types"
	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	RegistryParams	types.ActorRegistryParams
	State		types.ActorRegistryState
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
		RegistryParams:	types.DefaultActorRegistryParams(),
		State:		types.EmptyActorRegistryState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("actor registry prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.RegistryParams)
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

func (k *Keeper) RegisterCodeHash(code types.CodeRecord) error {
	code = code.Normalize()
	if err := code.Validate(); err != nil {
		return err
	}
	if k.hasCodeHash(code.CodeHash) {
		return errors.New("actor registry code hash already registered")
	}
	next := cloneGenesis(k.genesis)
	next.State.CodeStore = append(next.State.CodeStore, code)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k *Keeper) RegisterActor(msg types.MsgRegisterActor) (types.ActorRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ActorRecord{}, err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ActorRecord{}, err
	}
	if !k.hasCodeHash(msg.CodeHash) {
		return types.ActorRecord{}, errors.New("actor registry code hash is missing from AVM code store")
	}
	actorID := types.DeriveActorID(msg.Owner, msg.CodeHash, msg.Salt)
	if msg.ActorID != "" && msg.ActorID != actorID {
		return types.ActorRecord{}, errors.New("actor id derivation mismatch")
	}
	address := types.DeriveContractAddress(actorID)
	if msg.ContractAddress != "" && msg.ContractAddress != address {
		return types.ActorRecord{}, errors.New("actor contract address derivation mismatch")
	}
	if _, found := k.actorByID(actorID); found {
		return types.ActorRecord{}, errors.New("actor already registered")
	}
	if _, found := k.actorByAddress(address); found {
		return types.ActorRecord{}, errors.New("actor contract address already registered")
	}
	actor := types.ActorRecord{
		ActorID:		actorID,
		ContractAddress:	address,
		Owner:			msg.Owner,
		CodeHash:		msg.CodeHash,
		StorageRoot:		rootOrDefault(msg.StorageRoot, "storage"),
		MailboxRoot:		rootOrDefault(msg.MailboxRoot, "mailbox"),
		Balance:		msg.Balance,
		LogicalTime:		1,
		Status:			types.ActorStatusActive,
		RentStatus:		types.RentStatusCurrent,
		LastActiveHeight:	msg.Height,
		Capabilities:		msg.Capabilities,
	}
	if actor.LastActiveHeight == 0 {
		return types.ActorRecord{}, errors.New("actor registration height must be positive")
	}
	if err := actor.Validate(k.genesis.RegistryParams); err != nil {
		return types.ActorRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Actors = append(next.State.Actors, actor)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ActorRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ActorRecord{}, err
	}
	return actor.Normalize(), nil
}

func (k *Keeper) UpdateActorCode(msg types.MsgUpdateActorCode) (types.ActorRecord, error) {
	return k.updateActor(msg.Authority, msg.ActorID, msg.Height, msg.LogicalTime, func(actor types.ActorRecord, logicalTime uint64) (types.ActorRecord, error) {
		if actor.Status == types.ActorStatusDeleted {
			return types.ActorRecord{}, errors.New("deleted actor cannot update code")
		}
		if !k.hasCodeHash(msg.CodeHash) {
			return types.ActorRecord{}, errors.New("actor registry code hash is missing from AVM code store")
		}
		actor.CodeHash = msg.CodeHash
		actor.LogicalTime = logicalTime
		return actor, nil
	})
}

func (k *Keeper) FreezeActor(msg types.MsgFreezeActor) (types.ActorRecord, error) {
	return k.updateActor(msg.Authority, msg.ActorID, msg.Height, 0, func(actor types.ActorRecord, logicalTime uint64) (types.ActorRecord, error) {
		if actor.Status != types.ActorStatusActive {
			return types.ActorRecord{}, errors.New("only active actor can be frozen")
		}
		actor.Status = types.ActorStatusFrozen
		actor.LogicalTime = logicalTime
		return actor, nil
	})
}

func (k *Keeper) UnfreezeActor(msg types.MsgUnfreezeActor) (types.ActorRecord, error) {
	return k.updateActor(msg.Authority, msg.ActorID, msg.Height, 0, func(actor types.ActorRecord, logicalTime uint64) (types.ActorRecord, error) {
		if actor.Status != types.ActorStatusFrozen {
			return types.ActorRecord{}, errors.New("only frozen actor can be unfrozen")
		}
		actor.Status = types.ActorStatusActive
		actor.LogicalTime = logicalTime
		return actor, nil
	})
}

func (k *Keeper) DeleteActor(msg types.MsgDeleteActor) (types.ActorRecord, error) {
	return k.updateActor(msg.Authority, msg.ActorID, msg.Height, 0, func(actor types.ActorRecord, logicalTime uint64) (types.ActorRecord, error) {
		if actor.Status == types.ActorStatusDeleted {
			return types.ActorRecord{}, errors.New("actor already deleted")
		}
		actor.Status = types.ActorStatusDeleted
		actor.LogicalTime = logicalTime
		return actor, nil
	})
}

func (k *Keeper) MigrateActor(msg types.MsgMigrateActor) (types.ActorRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ActorRecord{}, err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ActorRecord{}, err
	}
	if msg.Height == 0 {
		return types.ActorRecord{}, errors.New("actor update height must be positive")
	}
	index, actor, found := k.actorIndex(msg.ActorID)
	if !found {
		return types.ActorRecord{}, errors.New("actor not found")
	}
	if actor.Status == types.ActorStatusDeleted {
		return types.ActorRecord{}, errors.New("deleted actor cannot migrate")
	}
	if !k.hasCodeHash(msg.NewCodeHash) {
		return types.ActorRecord{}, errors.New("actor registry code hash is missing from AVM code store")
	}
	logicalTime, err := types.NextLogicalTime(actor.LogicalTime, msg.LogicalTime)
	if err != nil {
		return types.ActorRecord{}, err
	}
	previousID := actor.ActorID
	previousCode := actor.CodeHash
	newActorID := msg.NewActorID
	if newActorID == "" {
		newActorID = types.DeriveActorID(actor.Owner, msg.NewCodeHash, actor.ActorID)
	}
	newAddress := msg.NewAddress
	if newAddress == "" {
		newAddress = types.DeriveContractAddress(newActorID)
	}
	if existing, found := k.actorByID(newActorID); found && existing.ActorID != actor.ActorID {
		return types.ActorRecord{}, errors.New("migrated actor id already exists")
	}
	actor.MigratedFrom = previousID
	actor.MigratedTo = newActorID
	actor.ActorID = newActorID
	actor.ContractAddress = newAddress
	actor.CodeHash = msg.NewCodeHash
	actor.StorageRoot = rootOrDefault(msg.NewStorageRoot, "storage")
	actor.MailboxRoot = rootOrDefault(msg.NewMailboxRoot, "mailbox")
	actor.Status = types.ActorStatusMigrated
	actor.LogicalTime = logicalTime
	actor.LastActiveHeight = msg.Height
	if err := actor.Validate(k.genesis.RegistryParams); err != nil {
		return types.ActorRecord{}, err
	}
	migration := types.ActorMigrationRecord{
		ActorID:	actor.ActorID,
		FromCodeHash:	previousCode,
		ToCodeHash:	msg.NewCodeHash,
		Height:		msg.Height,
		LogicalTime:	logicalTime,
	}
	if err := migration.Validate(); err != nil {
		return types.ActorRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Actors[index] = actor.Normalize()
	next.State.Migrations = append(next.State.Migrations, migration)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ActorRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ActorRecord{}, err
	}
	return actor.Normalize(), nil
}

func (k Keeper) Actor(actorID string) (types.ActorRecord, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.ActorRecord{}, false, err
	}
	actor, found := k.actorByID(actorID)
	return actor, found, nil
}

func (k Keeper) ActorsByOwner(owner string) ([]types.ActorRecord, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.ActorRecord, 0)
	for _, actor := range k.genesis.State.Export().Actors {
		if actor.Owner == owner {
			out = append(out, actor)
		}
	}
	types.SortActors(out)
	return out, nil
}

func (k Keeper) ActorsByCodeHash(codeHash string) ([]types.ActorRecord, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.ActorRecord, 0)
	for _, actor := range k.genesis.State.Export().Actors {
		if actor.CodeHash == codeHash {
			out = append(out, actor)
		}
	}
	types.SortActors(out)
	return out, nil
}

func (k Keeper) ActorStatus(actorID string) (string, bool, error) {
	actor, found, err := k.Actor(actorID)
	return actor.Status, found, err
}

func (k Keeper) ActorMailbox(actorID string) (string, bool, error) {
	actor, found, err := k.Actor(actorID)
	return actor.MailboxRoot, found, err
}

func (k Keeper) ActorStorageRoot(actorID string) (string, bool, error) {
	actor, found, err := k.Actor(actorID)
	return actor.StorageRoot, found, err
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

func (k *Keeper) updateActor(authority, actorID string, height, requestedLogicalTime uint64, mutate func(types.ActorRecord, uint64) (types.ActorRecord, error)) (types.ActorRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.ActorRecord{}, err
	}
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return types.ActorRecord{}, err
	}
	if height == 0 {
		return types.ActorRecord{}, errors.New("actor update height must be positive")
	}
	index, actor, found := k.actorIndex(actorID)
	if !found {
		return types.ActorRecord{}, errors.New("actor not found")
	}
	logicalTime, err := types.NextLogicalTime(actor.LogicalTime, requestedLogicalTime)
	if err != nil {
		return types.ActorRecord{}, err
	}
	nextActor, err := mutate(actor, logicalTime)
	if err != nil {
		return types.ActorRecord{}, err
	}
	nextActor.LastActiveHeight = height
	if err := nextActor.Validate(k.genesis.RegistryParams); err != nil {
		return types.ActorRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Actors[index] = nextActor.Normalize()
	next.State.Migrations = k.genesis.State.Migrations
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ActorRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ActorRecord{}, err
	}
	return nextActor.Normalize(), nil
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

func (k Keeper) actorIndex(actorID string) (int, types.ActorRecord, bool) {
	for i, actor := range k.genesis.State.Export().Actors {
		if actor.ActorID == actorID {
			return i, actor, true
		}
	}
	return -1, types.ActorRecord{}, false
}

func (k Keeper) actorByID(actorID string) (types.ActorRecord, bool) {
	_, actor, found := k.actorIndex(actorID)
	return actor, found
}

func (k Keeper) actorByAddress(address string) (types.ActorRecord, bool) {
	for _, actor := range k.genesis.State.Export().Actors {
		if actor.ContractAddress == address {
			return actor, true
		}
	}
	return types.ActorRecord{}, false
}

func (k Keeper) hasCodeHash(codeHash string) bool {
	for _, code := range k.genesis.State.Export().CodeStore {
		if code.CodeHash == codeHash {
			return true
		}
	}
	return false
}

func rootOrDefault(root, seed string) string {
	if root != "" {
		return root
	}
	return types.DefaultRoot(seed)
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
