package keeper

import (
	"context"
	"errors"
	"fmt"
	"sort"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.ConfigState
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
		Params:		types.DefaultParams(),
		State:		types.ConfigState{},
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("config unsupported genesis version")
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
		return errors.New("config persistent store is not configured")
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, k.genesis)
}

func (k Keeper) saveGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if k.storeService == nil {
		return errors.New("config persistent store is not configured")
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, cloneGenesis(gs))
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
	gs.State.Entries = types.SortedEntries(gs.State.Entries)
	gs.State.PendingChanges = types.SortedChanges(gs.State.PendingChanges)
	return cloneGenesis(gs), nil
}

func (k *Keeper) UpdateParams(authority string, params types.Params) error {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	next.Params = params
	if err := next.Validate(); err != nil {
		return err
	}
	return k.saveGenesis(next)
}

func (k Keeper) UpdateParamsState(ctx context.Context, authority string, params types.Params) error {
	gs, err := k.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	if err := gs.Params.Authorize(authority); err != nil {
		return err
	}
	gs.Params = params
	return k.saveGenesisState(ctx, gs)
}

func (k *Keeper) UpsertEntry(authority string, key string, value string, height int64) (types.ConfigEntry, error) {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return types.ConfigEntry{}, err
	}
	if height < 0 {
		return types.ConfigEntry{}, errors.New("config update height must be non-negative")
	}
	entries := types.SortedEntries(k.genesis.State.Entries)
	idx := sort.Search(len(entries), func(i int) bool {
		return entries[i].Key >= key
	})
	version := uint64(1)
	if idx < len(entries) && entries[idx].Key == key {
		version = entries[idx].Version + 1
	} else if uint32(len(entries)+1) > k.genesis.Params.MaxEntries {
		return types.ConfigEntry{}, errors.New("config entries limit reached")
	}
	entry := types.ConfigEntry{
		Key:		key,
		Value:		value,
		Owner:		authority,
		Version:	version,
		UpdatedHeight:	height,
	}
	if err := entry.Validate(k.genesis.Params); err != nil {
		return types.ConfigEntry{}, err
	}
	if idx < len(entries) && entries[idx].Key == key {
		entries[idx] = entry
	} else {
		entries = append(entries, types.ConfigEntry{})
		copy(entries[idx+1:], entries[idx:])
		entries[idx] = entry
	}
	next := cloneGenesis(k.genesis)
	next.State.Entries = entries
	if err := next.Validate(); err != nil {
		return types.ConfigEntry{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigEntry{}, err
	}
	return entry, nil
}

func (k Keeper) UpsertEntryState(ctx context.Context, authority string, key string, value string, height int64) (types.ConfigEntry, error) {
	gs, err := k.ExportGenesisState(ctx)
	if err != nil {
		return types.ConfigEntry{}, err
	}
	memory := NewKeeper()
	memory.genesis = gs
	entry, err := memory.UpsertEntry(authority, key, value, height)
	if err != nil {
		return types.ConfigEntry{}, err
	}
	return entry, k.saveGenesisState(ctx, memory.genesis)
}

func (k *Keeper) SubmitConfigChange(msg types.MsgSubmitConfigChange, height int64) (types.ConfigChange, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ConfigChange{}, err
	}
	if height < 0 {
		return types.ConfigChange{}, errors.New("config change height must be non-negative")
	}
	if uint32(len(k.genesis.State.PendingChanges)+1) > k.genesis.Params.MaxPendingChanges {
		return types.ConfigChange{}, errors.New("config pending changes limit reached")
	}
	change := normalizeSubmittedChange(k.genesis.Params, msg.Change, msg.Authority, height)
	if _, _, found := types.FindChange(k.genesis.State.PendingChanges, change.ID); found {
		return types.ConfigChange{}, errors.New("config change already exists")
	}
	if err := types.ValidateChangeAgainstState(k.genesis.Params, k.genesis.State, change); err != nil {
		return types.ConfigChange{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.PendingChanges = types.UpsertChange(next.State.PendingChanges, change)
	if err := next.Validate(); err != nil {
		return types.ConfigChange{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigChange{}, err
	}
	return change, nil
}

func (k *Keeper) ApproveConfigChange(msg types.MsgApproveConfigChange, height int64) (types.ConfigChange, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ConfigChange{}, err
	}
	change, err := k.transitionChange(msg.ChangeID, height, func(change types.ConfigChange) (types.ConfigChange, error) {
		if change.Status != types.ChangeStatusPending {
			return types.ConfigChange{}, errors.New("config change must be pending to approve")
		}
		change.Status = types.ChangeStatusApproved
		change.ApprovedBy = msg.Authority
		return change, nil
	})
	return change, err
}

func (k *Keeper) RejectConfigChange(msg types.MsgRejectConfigChange, height int64) (types.ConfigChange, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ConfigChange{}, err
	}
	change, err := k.transitionChange(msg.ChangeID, height, func(change types.ConfigChange) (types.ConfigChange, error) {
		if change.Status != types.ChangeStatusPending && change.Status != types.ChangeStatusApproved {
			return types.ConfigChange{}, errors.New("config change must be pending or approved to reject")
		}
		change.Status = types.ChangeStatusRejected
		change.RejectedBy = msg.Authority
		change.Reason = msg.Reason
		return change, nil
	})
	return change, err
}

func (k *Keeper) CancelConfigChange(msg types.MsgCancelConfigChange, height int64) (types.ConfigChange, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ConfigChange{}, err
	}
	change, err := k.transitionChange(msg.ChangeID, height, func(change types.ConfigChange) (types.ConfigChange, error) {
		if change.Status != types.ChangeStatusPending && change.Status != types.ChangeStatusApproved {
			return types.ConfigChange{}, errors.New("config change must be pending or approved to cancel")
		}
		change.Status = types.ChangeStatusCancelled
		change.CancelledBy = msg.Authority
		change.Reason = msg.Reason
		return change, nil
	})
	return change, err
}

func (k *Keeper) ExecuteConfigChange(msg types.MsgExecuteConfigChange, height int64) (types.ConfigEntry, types.ConfigChange, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ConfigEntry{}, types.ConfigChange{}, err
	}
	if height < 0 {
		return types.ConfigEntry{}, types.ConfigChange{}, errors.New("config change height must be non-negative")
	}
	idx, change, found := types.FindChange(k.genesis.State.PendingChanges, msg.ChangeID)
	if !found {
		return types.ConfigEntry{}, types.ConfigChange{}, errors.New("config change not found")
	}
	if change.Status != types.ChangeStatusApproved {
		return types.ConfigEntry{}, types.ConfigChange{}, errors.New("config change must be approved before execution")
	}
	if change.ActivationHeight != 0 && height < change.ActivationHeight {
		return types.ConfigEntry{}, types.ConfigChange{}, errors.New("config change activation height has not elapsed")
	}
	if err := types.ValidateChangeAgainstState(k.genesis.Params, k.genesis.State, change); err != nil {
		return types.ConfigEntry{}, types.ConfigChange{}, err
	}
	next := cloneGenesis(k.genesis)
	entry, err := applyConfigChange(next.Params, &next.State, change, msg.Authority, height)
	if err != nil {
		return types.ConfigEntry{}, types.ConfigChange{}, err
	}
	change.Status = types.ChangeStatusExecuted
	change.ExecutedBy = msg.Authority
	change.UpdatedHeight = height
	next.State.PendingChanges[idx] = change
	next.State.PendingChanges = types.SortedChanges(next.State.PendingChanges)
	if err := next.Validate(); err != nil {
		return types.ConfigEntry{}, types.ConfigChange{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigEntry{}, types.ConfigChange{}, err
	}
	return entry, change, nil
}

func (k Keeper) Entry(key string) (types.ConfigEntry, bool, error) {
	state := k.genesis.State
	if err := state.Validate(k.genesis.Params); err != nil {
		return types.ConfigEntry{}, false, err
	}
	idx := sort.Search(len(state.Entries), func(i int) bool {
		return state.Entries[i].Key >= key
	})
	if idx >= len(state.Entries) || state.Entries[idx].Key != key {
		return types.ConfigEntry{}, false, nil
	}
	return state.Entries[idx], true, nil
}

func (k Keeper) Entries() ([]types.ConfigEntry, error) {
	state := k.genesis.State
	if err := state.Validate(k.genesis.Params); err != nil {
		return nil, err
	}
	return types.SortedEntries(state.Entries), nil
}

func (k Keeper) ConfigValue(key string) (string, bool, error) {
	entry, found, err := k.Entry(key)
	if err != nil {
		return "", false, err
	}
	return entry.Value, found, nil
}

func (k Keeper) PendingConfigChanges() ([]types.ConfigChange, error) {
	state := k.genesis.State
	if err := state.Validate(k.genesis.Params); err != nil {
		return nil, err
	}
	out := make([]types.ConfigChange, 0, len(state.PendingChanges))
	for _, change := range types.SortedChanges(state.PendingChanges) {
		if change.Status == types.ChangeStatusPending || change.Status == types.ChangeStatusApproved {
			out = append(out, change)
		}
	}
	return out, nil
}

func (k Keeper) EffectiveEntries(height int64) ([]types.ConfigEntry, error) {
	state := types.CloneState(k.genesis.State)
	if err := state.Validate(k.genesis.Params); err != nil {
		return nil, err
	}
	for _, change := range types.SortedChanges(state.PendingChanges) {
		if change.Status != types.ChangeStatusApproved {
			continue
		}
		if change.ActivationHeight != 0 && height < change.ActivationHeight {
			continue
		}
		if err := types.ValidateChangeAgainstState(k.genesis.Params, state, change); err != nil {
			return nil, err
		}
		if _, err := applyConfigChange(k.genesis.Params, &state, change, change.ApprovedBy, height); err != nil {
			return nil, err
		}
	}
	return types.SortedEntries(state.Entries), nil
}

func (k Keeper) ConfigChange(id string) (types.ConfigChange, bool, error) {
	state := k.genesis.State
	if err := state.Validate(k.genesis.Params); err != nil {
		return types.ConfigChange{}, false, err
	}
	_, change, found := types.FindChange(state.PendingChanges, id)
	return change, found, nil
}

func (k Keeper) Authority() string {
	return k.genesis.Params.Authority
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	gs := m.keeper.ExportGenesis()
	return gs.Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = types.CloneState(gs.State)
	gs.State.Entries = types.SortedEntries(gs.State.Entries)
	gs.State.PendingChanges = types.SortedChanges(gs.State.PendingChanges)
	return gs
}

func normalizeSubmittedChange(params types.Params, change types.ConfigChange, authority string, height int64) types.ConfigChange {
	change.ID = stringsTrim(change.ID)
	change.Key = stringsTrim(change.Key)
	change.Operation = stringsTrim(change.Operation)
	if change.Operation == "" {
		change.Operation = types.OperationSet
	}
	change.Status = types.ChangeStatusPending
	change.SubmittedBy = authority
	change.ApprovedBy = ""
	change.RejectedBy = ""
	change.CancelledBy = ""
	change.ExecutedBy = ""
	change.CreatedHeight = height
	change.UpdatedHeight = height
	change.Critical = types.IsCriticalConfigKey(change.Key)
	change.ActivationHeight, change.ActivationEpoch = types.ActivationHeight(params, change.Key, height)
	return change
}

func (k *Keeper) transitionChange(id string, height int64, mutate func(types.ConfigChange) (types.ConfigChange, error)) (types.ConfigChange, error) {
	if height < 0 {
		return types.ConfigChange{}, errors.New("config change height must be non-negative")
	}
	idx, change, found := types.FindChange(k.genesis.State.PendingChanges, id)
	if !found {
		return types.ConfigChange{}, errors.New("config change not found")
	}
	updated, err := mutate(change)
	if err != nil {
		return types.ConfigChange{}, err
	}
	updated.UpdatedHeight = height
	if err := updated.Validate(k.genesis.Params); err != nil {
		return types.ConfigChange{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.PendingChanges[idx] = updated
	next.State.PendingChanges = types.SortedChanges(next.State.PendingChanges)
	if err := next.Validate(); err != nil {
		return types.ConfigChange{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ConfigChange{}, err
	}
	return updated, nil
}

func applyConfigChange(params types.Params, state *types.ConfigState, change types.ConfigChange, authority string, height int64) (types.ConfigEntry, error) {
	entries := types.SortedEntries(state.Entries)
	idx := sort.Search(len(entries), func(i int) bool {
		return entries[i].Key >= change.Key
	})
	switch change.Operation {
	case types.OperationSet:
		version := uint64(1)
		if idx < len(entries) && entries[idx].Key == change.Key {
			version = entries[idx].Version + 1
		} else if uint32(len(entries)+1) > params.MaxEntries {
			return types.ConfigEntry{}, errors.New("config entries limit reached")
		}
		entry := types.ConfigEntry{
			Key:		change.Key,
			Value:		change.Value,
			Owner:		authority,
			Version:	version,
			UpdatedHeight:	height,
		}
		if err := entry.Validate(params); err != nil {
			return types.ConfigEntry{}, err
		}
		if idx < len(entries) && entries[idx].Key == change.Key {
			entries[idx] = entry
		} else {
			entries = append(entries, types.ConfigEntry{})
			copy(entries[idx+1:], entries[idx:])
			entries[idx] = entry
		}
		state.Entries = entries
		return entry, nil
	case types.OperationDelete:
		if idx >= len(entries) || entries[idx].Key != change.Key {
			return types.ConfigEntry{}, fmt.Errorf("config entry %s not found", change.Key)
		}
		entry := entries[idx]
		entries = append(entries[:idx], entries[idx+1:]...)
		state.Entries = entries
		return entry, nil
	default:
		return types.ConfigEntry{}, fmt.Errorf("unsupported config operation %s", change.Operation)
	}
}

func stringsTrim(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\t' || value[0] == '\r' || value[0] == '\n') {
		value = value[1:]
	}
	for len(value) > 0 {
		last := value[len(value)-1]
		if last != ' ' && last != '\t' && last != '\r' && last != '\n' {
			break
		}
		value = value[:len(value)-1]
	}
	return value
}
