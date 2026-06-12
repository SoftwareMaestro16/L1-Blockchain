package keeper

import (
	"context"
	"encoding/json"
	"errors"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/epoch/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

var genesisKey = []byte{0x01}

const DefaultEpochHeightSpan = uint64(100)

type GenesisState struct {
	Version	uint64
	State	types.EpochState
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
		State: types.EpochState{
			Params:			postypes.DefaultParams(),
			EpochHeightSpan:	DefaultEpochHeightSpan,
		},
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("epoch unsupported genesis version")
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

func (k Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if k.storeService == nil {
		return errors.New("epoch persistent store is not configured")
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

func (k *Keeper) Bootstrap(params postypes.Params, epochID uint64, startHeight uint64, startUnixSeconds uint64, epochHeightSpan uint64, validators []postypes.ScoredValidator) (postypes.EpochRecord, error) {
	if epochHeightSpan == 0 {
		return postypes.EpochRecord{}, errors.New("epoch height span must be positive")
	}
	if startUnixSeconds == 0 {
		return postypes.EpochRecord{}, errors.New("epoch start time must be positive")
	}
	record, err := postypes.NewEpochRecord(params, epochID, startHeight, startHeight+epochHeightSpan-1, postypes.EpochPhaseDelegation, "", validators)
	if err != nil {
		return postypes.EpochRecord{}, err
	}
	next := k.genesis
	next.State = types.EpochState{
		Params:				params,
		Current:			record,
		CurrentStartUnixSeconds:	startUnixSeconds,
		EpochHeightSpan:		epochHeightSpan,
	}
	if err := next.Validate(); err != nil {
		return postypes.EpochRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) BeginEpoch(height uint64, unixSeconds uint64) (types.HookRecord, error) {
	state := k.genesis.State
	if err := requireCurrent(state); err != nil {
		return types.HookRecord{}, err
	}
	if height != state.Current.StartHeight {
		return types.HookRecord{}, errors.New("epoch begin hook must run at epoch start height")
	}
	hook := types.HookRecord{
		Event:			types.HookEventEpochBegin,
		EpochID:		state.Current.EpochID,
		Height:			height,
		UnixSeconds:		unixSeconds,
		ToPhase:		postypes.EpochPhaseDelegation,
		ValidatorSetHash:	state.Current.ValidatorSetHash,
	}
	if err := hook.Validate(); err != nil {
		return types.HookRecord{}, err
	}
	k.genesis.State.HookLog = append(k.genesis.State.HookLog, hook)
	return hook, nil
}

func (k *Keeper) TransitionPhase(height uint64, unixSeconds uint64, to postypes.EpochPhase) (types.HookRecord, error) {
	state := k.genesis.State
	if err := requireCurrent(state); err != nil {
		return types.HookRecord{}, err
	}
	if err := postypes.ValidateEpochPhaseTransition(state.Current.Phase, to); err != nil {
		return types.HookRecord{}, err
	}
	next := state.Current
	next.Phase = to
	if err := next.Validate(); err != nil {
		return types.HookRecord{}, err
	}
	hook := types.HookRecord{
		Event:			types.HookEventPhaseTransition,
		EpochID:		next.EpochID,
		Height:			height,
		UnixSeconds:		unixSeconds,
		FromPhase:		state.Current.Phase,
		ToPhase:		to,
		ValidatorSetHash:	next.ValidatorSetHash,
	}
	if err := hook.Validate(); err != nil {
		return types.HookRecord{}, err
	}
	k.genesis.State.Current = next
	k.genesis.State.HookLog = append(k.genesis.State.HookLog, hook)
	return hook, nil
}

func (k *Keeper) SyncPhaseByTime(height uint64, unixSeconds uint64) ([]types.HookRecord, error) {
	state := k.genesis.State
	if err := requireCurrent(state); err != nil {
		return nil, err
	}
	target, err := postypes.EpochPhaseAt(state.Params, state.CurrentStartUnixSeconds, unixSeconds)
	if err != nil {
		return nil, err
	}
	if target == postypes.EpochPhaseClosed {
		target = postypes.EpochPhaseSettlement
	}
	hooks := make([]types.HookRecord, 0)
	for k.genesis.State.Current.Phase != target {
		next, _, err := postypes.NextEpochPhase(k.genesis.State.Current.Phase)
		if err != nil {
			return nil, err
		}
		if next == postypes.EpochPhaseClosed {
			break
		}
		hook, err := k.TransitionPhase(height, unixSeconds, next)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, hook)
	}
	return hooks, nil
}

func (k *Keeper) EndEpoch(height uint64, unixSeconds uint64, roots postypes.EpochSettlementRoots, nextValidators []postypes.ScoredValidator) (postypes.EpochRecord, error) {
	state := k.genesis.State
	if err := requireCurrent(state); err != nil {
		return postypes.EpochRecord{}, err
	}
	if height != state.Current.EndHeight {
		return postypes.EpochRecord{}, errors.New("epoch end hook must run at epoch end height")
	}
	closed, err := postypes.CloseEpochRecord(state.Current, roots.PerformanceRoot, roots.RewardRoot, roots.SlashRoot)
	if err != nil {
		return postypes.EpochRecord{}, err
	}
	nextRecord, err := postypes.NewEpochRecord(
		state.Params,
		closed.EpochID+1,
		closed.EndHeight+1,
		closed.EndHeight+state.EpochHeightSpan,
		postypes.EpochPhaseDelegation,
		closed.Seed,
		nextValidators,
	)
	if err != nil {
		return postypes.EpochRecord{}, err
	}
	if err := postypes.ValidateConsecutiveEpochs(closed, nextRecord); err != nil {
		return postypes.EpochRecord{}, err
	}
	endHook := types.HookRecord{
		Event:			types.HookEventEpochEnd,
		EpochID:		closed.EpochID,
		Height:			height,
		UnixSeconds:		unixSeconds,
		FromPhase:		postypes.EpochPhaseSettlement,
		ToPhase:		postypes.EpochPhaseClosed,
		ValidatorSetHash:	closed.ValidatorSetHash,
	}
	beginHook := types.HookRecord{
		Event:			types.HookEventEpochBegin,
		EpochID:		nextRecord.EpochID,
		Height:			nextRecord.StartHeight,
		UnixSeconds:		unixSeconds + 1,
		ToPhase:		postypes.EpochPhaseDelegation,
		ValidatorSetHash:	nextRecord.ValidatorSetHash,
	}
	if err := endHook.Validate(); err != nil {
		return postypes.EpochRecord{}, err
	}
	if err := beginHook.Validate(); err != nil {
		return postypes.EpochRecord{}, err
	}
	k.genesis.State.History = append(k.genesis.State.History, closed)
	k.genesis.State.Current = nextRecord
	k.genesis.State.CurrentStartUnixSeconds = unixSeconds + 1
	k.genesis.State.HookLog = append(k.genesis.State.HookLog, endHook, beginHook)
	return closed, nil
}

func (k Keeper) CurrentEpoch() (postypes.EpochRecord, error) {
	if err := requireCurrent(k.genesis.State); err != nil {
		return postypes.EpochRecord{}, err
	}
	return k.genesis.State.Current, nil
}

func (k Keeper) Epoch(epochID uint64) (postypes.EpochRecord, bool, error) {
	if err := k.genesis.State.Validate(); err != nil {
		return postypes.EpochRecord{}, false, err
	}
	if k.genesis.State.Current.Seed != "" && k.genesis.State.Current.EpochID == epochID {
		return k.genesis.State.Current, true, nil
	}
	for _, record := range k.genesis.State.History {
		if record.EpochID == epochID {
			return record, true, nil
		}
	}
	return postypes.EpochRecord{}, false, nil
}

func (k Keeper) HistoricalEpochs(req *prototype.PageRequest) ([]postypes.EpochRecord, prototype.PageResponse, error) {
	state := k.genesis.State
	if err := state.Validate(); err != nil {
		return nil, prototype.PageResponse{}, err
	}
	start, end, res, err := prototype.NormalizePage(req, prototype.TestnetParams(), len(state.History))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]postypes.EpochRecord, end-start)
	copy(out, state.History[start:end])
	return out, res, nil
}

func (k Keeper) HookLog() []types.HookRecord {
	hooks := k.genesis.State.HookLog
	out := make([]types.HookRecord, len(hooks))
	copy(out, hooks)
	return out
}

func (k Keeper) ExportState() types.EpochState {
	return types.CloneState(k.genesis.State)
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

func requireCurrent(state types.EpochState) error {
	if err := state.Validate(); err != nil {
		return err
	}
	if state.Current.Seed == "" {
		return errors.New("current epoch is not initialized")
	}
	return nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = types.CloneState(gs.State)
	return gs
}
