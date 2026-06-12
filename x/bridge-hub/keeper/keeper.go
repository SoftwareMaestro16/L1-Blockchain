package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/bridge-hub/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	BridgeParams	types.BridgeHubParams
	State		types.BridgeHubState
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
		BridgeParams:	types.DefaultBridgeHubParams(),
		State:		types.EmptyBridgeHubState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("bridge hub prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.BridgeParams)
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

func (k *Keeper) RegisterBridge(msg types.MsgRegisterBridge) (types.BridgeRecord, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.BridgeRecord{}, err
	}
	bridge := msg.Bridge.Normalize(k.genesis.BridgeParams)
	if bridge.RegisteredHeight == 0 {
		return types.BridgeRecord{}, errors.New("bridge registration height must be positive")
	}
	bridge.UpdatedHeight = bridge.RegisteredHeight
	bridge.LimitWindowStart = types.CurrentWindowStart(bridge.RegisteredHeight)
	if _, _, found := bridgeIndex(k.genesis.State.Bridges, bridge.BridgeID); found {
		return types.BridgeRecord{}, errors.New("bridge already registered")
	}
	if err := bridge.Validate(k.genesis.BridgeParams); err != nil {
		return types.BridgeRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Bridges = append(next.State.Bridges, bridge)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.BridgeRecord{}, err
	}
	k.genesis = next
	return bridge, nil
}

func (k *Keeper) PauseBridge(msg types.MsgPauseBridge) (types.BridgeRecord, error) {
	return k.updateBridge(msg.Authority, msg.BridgeID, msg.Height, func(bridge types.BridgeRecord) (types.BridgeRecord, error) {
		if bridge.Paused {
			return types.BridgeRecord{}, errors.New("bridge already paused")
		}
		bridge.Paused = true
		return bridge, nil
	})
}

func (k *Keeper) ResumeBridge(msg types.MsgResumeBridge) (types.BridgeRecord, error) {
	return k.updateBridge(msg.Authority, msg.BridgeID, msg.Height, func(bridge types.BridgeRecord) (types.BridgeRecord, error) {
		if !bridge.Paused {
			return types.BridgeRecord{}, errors.New("bridge is not paused")
		}
		bridge.Paused = false
		return bridge, nil
	})
}

func (k *Keeper) RegisterAssetMapping(msg types.MsgRegisterAssetMapping) (types.AssetMapping, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.AssetMapping{}, err
	}
	mapping := msg.Mapping.Normalize()
	if _, _, found := bridgeIndex(k.genesis.State.Bridges, mapping.BridgeID); !found {
		return types.AssetMapping{}, errors.New("bridge mapping references unknown bridge")
	}
	if _, found := mappingConflict(k.genesis.State.AssetMappings, mapping); found {
		return types.AssetMapping{}, errors.New("bridge asset mapping conflict")
	}
	if err := mapping.Validate(); err != nil {
		return types.AssetMapping{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.AssetMappings = append(next.State.AssetMappings, mapping)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.AssetMapping{}, err
	}
	k.genesis = next
	return mapping, nil
}

func (k *Keeper) UpdateBridgeLimits(msg types.MsgUpdateBridgeLimits) (types.BridgeRecord, error) {
	if msg.DailyLimit == 0 {
		return types.BridgeRecord{}, errors.New("bridge daily limit must be positive")
	}
	return k.updateBridge(msg.Authority, msg.BridgeID, msg.Height, func(bridge types.BridgeRecord) (types.BridgeRecord, error) {
		bridge.DailyLimit = msg.DailyLimit
		if bridge.DailyUsed > bridge.DailyLimit {
			return types.BridgeRecord{}, errors.New("bridge current daily usage exceeds requested limit")
		}
		return bridge, nil
	})
}

func (k *Keeper) SubmitBridgeEvent(msg types.MsgSubmitBridgeEvent) (types.BridgeEvent, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return types.BridgeEvent{}, err
	}
	if msg.Submitter == "" {
		return types.BridgeEvent{}, errors.New("bridge event submitter is required")
	}
	event := msg.Event.Normalize()
	event.SubmittedBy = msg.Submitter
	event.Status = types.BridgeEventPending
	if event.SubmittedHeight == 0 {
		return types.BridgeEvent{}, errors.New("bridge event submitted height must be positive")
	}
	_, bridge, found := bridgeIndex(k.genesis.State.Bridges, event.BridgeID)
	if !found {
		return types.BridgeEvent{}, errors.New("bridge event references unknown bridge")
	}
	if event.SourceChain != bridge.SourceChain {
		return types.BridgeEvent{}, errors.New("bridge event source chain mismatch")
	}
	if event.ProofPolicy != bridge.ProofPolicy {
		return types.BridgeEvent{}, errors.New("bridge event proof policy must match registered chain policy")
	}
	if _, _, found := eventIndex(k.genesis.State.Events, event.EventID); found {
		return types.BridgeEvent{}, errors.New("bridge event already submitted")
	}
	if err := event.Validate(); err != nil {
		return types.BridgeEvent{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Events = append(next.State.Events, event)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.BridgeEvent{}, err
	}
	k.genesis = next
	return event, nil
}

func (k *Keeper) FinalizeBridgeEvent(msg types.MsgFinalizeBridgeEvent) (types.BridgeEvent, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.BridgeEvent{}, err
	}
	if msg.Height == 0 {
		return types.BridgeEvent{}, errors.New("bridge event finalization height must be positive")
	}
	eventIdx, event, found := eventIndex(k.genesis.State.Events, msg.EventID)
	if !found {
		return types.BridgeEvent{}, errors.New("bridge event not found")
	}
	if event.Status != types.BridgeEventPending || event.FinalizedHeight != 0 {
		return types.BridgeEvent{}, errors.New("bridge event cannot finalize twice")
	}
	bridgeIdx, bridge, found := bridgeIndex(k.genesis.State.Bridges, event.BridgeID)
	if !found {
		return types.BridgeEvent{}, errors.New("bridge event references unknown bridge")
	}
	if bridge.Paused {
		return types.BridgeEvent{}, errors.New("paused bridge cannot finalize events")
	}
	if bridge.RiskStatus == types.BridgeRiskCritical {
		return types.BridgeEvent{}, errors.New("critical-risk bridge cannot finalize events")
	}
	if event.ProofPolicy != bridge.ProofPolicy {
		return types.BridgeEvent{}, errors.New("bridge event proof policy must match registered chain policy")
	}
	bridge = resetWindowIfNeeded(bridge, msg.Height)
	if event.Amount > bridge.DailyLimit-bridge.DailyUsed {
		return types.BridgeEvent{}, errors.New("bridge daily limit exceeded")
	}
	bridge.DailyUsed += event.Amount
	bridge.UpdatedHeight = msg.Height
	event.Status = types.BridgeEventFinalized
	event.FinalizedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Bridges[bridgeIdx] = bridge.Normalize(next.BridgeParams)
	next.State.Events[eventIdx] = event.Normalize()
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.BridgeEvent{}, err
	}
	k.genesis = next
	return event, nil
}

func (k Keeper) Bridge(bridgeID string) (types.BridgeRecord, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.BridgeRecord{}, false, err
	}
	_, bridge, found := bridgeIndex(k.genesis.State.Bridges, bridgeID)
	return bridge, found, nil
}

func (k Keeper) Bridges() ([]types.BridgeRecord, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	return k.genesis.State.Export().Bridges, nil
}

func (k Keeper) AssetMapping(bridgeID, sourceAsset string) (types.AssetMapping, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.AssetMapping{}, false, err
	}
	for _, mapping := range k.genesis.State.Export().AssetMappings {
		if mapping.BridgeID == bridgeID && mapping.SourceAsset == sourceAsset {
			return mapping, true, nil
		}
	}
	return types.AssetMapping{}, false, nil
}

func (k Keeper) BridgeLimits(bridgeID string) (uint64, uint64, uint64, bool, error) {
	bridge, found, err := k.Bridge(bridgeID)
	return bridge.DailyLimit, bridge.DailyUsed, bridge.LimitWindowStart, found, err
}

func (k Keeper) BridgeEvents(bridgeID string) ([]types.BridgeEvent, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	out := make([]types.BridgeEvent, 0)
	for _, event := range k.genesis.State.Export().Events {
		if bridgeID == "" || event.BridgeID == bridgeID {
			out = append(out, event)
		}
	}
	types.SortEvents(out)
	return out, nil
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

func (k Keeper) requireAuthority(authority string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return k.genesis.Params.Authorize(authority)
}

func (k *Keeper) updateBridge(authority, bridgeID string, height uint64, mutate func(types.BridgeRecord) (types.BridgeRecord, error)) (types.BridgeRecord, error) {
	if err := k.requireAuthority(authority); err != nil {
		return types.BridgeRecord{}, err
	}
	if height == 0 {
		return types.BridgeRecord{}, errors.New("bridge update height must be positive")
	}
	index, bridge, found := bridgeIndex(k.genesis.State.Bridges, bridgeID)
	if !found {
		return types.BridgeRecord{}, errors.New("bridge not found")
	}
	bridge, err := mutate(bridge)
	if err != nil {
		return types.BridgeRecord{}, err
	}
	bridge.UpdatedHeight = height
	next := cloneGenesis(k.genesis)
	next.State.Bridges[index] = bridge.Normalize(next.BridgeParams)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.BridgeRecord{}, err
	}
	k.genesis = next
	return bridge.Normalize(k.genesis.BridgeParams), nil
}

func bridgeIndex(bridges []types.BridgeRecord, bridgeID string) (int, types.BridgeRecord, bool) {
	for i, bridge := range bridges {
		if bridge.BridgeID == bridgeID {
			return i, bridge, true
		}
	}
	return -1, types.BridgeRecord{}, false
}

func eventIndex(events []types.BridgeEvent, eventID string) (int, types.BridgeEvent, bool) {
	for i, event := range events {
		if event.EventID == eventID {
			return i, event, true
		}
	}
	return -1, types.BridgeEvent{}, false
}

func mappingConflict(mappings []types.AssetMapping, next types.AssetMapping) (types.AssetMapping, bool) {
	for _, mapping := range mappings {
		if mapping.BridgeID != next.BridgeID {
			continue
		}
		if mapping.SourceAsset == next.SourceAsset || mapping.TargetAsset == next.TargetAsset {
			if mapping.SourceAsset != next.SourceAsset || mapping.TargetAsset != next.TargetAsset {
				return mapping, true
			}
			return mapping, true
		}
	}
	return types.AssetMapping{}, false
}

func resetWindowIfNeeded(bridge types.BridgeRecord, height uint64) types.BridgeRecord {
	window := types.CurrentWindowStart(height)
	if bridge.LimitWindowStart != window {
		bridge.LimitWindowStart = window
		bridge.DailyUsed = 0
	}
	return bridge
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
