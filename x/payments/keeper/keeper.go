package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version uint64
	Params  prototype.Params
	State   paymentstypes.PaymentsState
}

type Keeper struct {
	genesis      GenesisState
	storeService corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version: prototype.CurrentGenesisVersion,
		Params:  prototype.DefaultParams(),
		State:   paymentstypes.EmptyState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("payments prototype unsupported genesis version")
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

func (k *Keeper) OpenChannel(channel paymentstypes.ChannelRecord) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.OpenChannel(k.genesis.State, channel)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) RegisterRoutingEdge(edge paymentstypes.ChannelEdge) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.RegisterRoutingEdge(k.genesis.State, edge)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) AcceptSignedState(channelID string, nextState paymentstypes.ChannelState, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.AcceptSignedState(k.genesis.State, channelID, nextState, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) AcceptAsyncCheckpoint(channelID string, checkpoint paymentstypes.ChannelState, deltas []paymentstypes.AsyncPaymentDelta, submitter string, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.AcceptAsyncCheckpoint(k.genesis.State, channelID, checkpoint, deltas, submitter, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) RegisterUpdateCheckpoint(req paymentstypes.ChannelUpdateRequest) (paymentstypes.ChannelUpdateResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ChannelUpdateResult{}, err
	}
	next, result, err := paymentstypes.RegisterUpdateCheckpoint(k.genesis.State, req)
	if err != nil {
		return paymentstypes.ChannelUpdateResult{}, err
	}
	k.genesis.State = next
	return result, nil
}

func (k *Keeper) RevealPromisePreimage(req paymentstypes.PreimageRevealRequest) ([]paymentstypes.ConditionResolution, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	next, resolutions, err := paymentstypes.RevealPromisePreimage(k.genesis.State, req)
	if err != nil {
		return nil, err
	}
	k.genesis.State = next
	return resolutions, nil
}

func (k *Keeper) ExpireConditionalPromises(req paymentstypes.PromiseExpiryRequest) ([]paymentstypes.ConditionResolution, paymentstypes.ConditionRootUpdate, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, paymentstypes.ConditionRootUpdate{}, err
	}
	next, resolutions, update, err := paymentstypes.ExpireConditionalPromises(k.genesis.State, req)
	if err != nil {
		return nil, paymentstypes.ConditionRootUpdate{}, err
	}
	k.genesis.State = next
	return resolutions, update, nil
}

func (k *Keeper) BatchSettleLinkedPromises(req paymentstypes.BatchConditionSettlementRequest) (paymentstypes.BatchConditionSettlementResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.BatchConditionSettlementResult{}, err
	}
	next, result, err := paymentstypes.BatchSettleLinkedPromises(k.genesis.State, req)
	if err != nil {
		return paymentstypes.BatchConditionSettlementResult{}, err
	}
	k.genesis.State = next
	return result, nil
}

func (k *Keeper) SubmitClose(channelID string, closingState paymentstypes.ChannelState, submitter string, currentHeight uint64, settlementFee string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.SubmitClose(k.genesis.State, channelID, closingState, submitter, currentHeight, settlementFee)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) SubmitCloseWithRequest(req paymentstypes.ChannelCloseRequest) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.SubmitCloseWithRequest(k.genesis.State, req)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) ForcedClose(channelID string, submitter string, currentHeight uint64, settlementFee string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.ForcedClose(k.genesis.State, channelID, submitter, currentHeight, settlementFee)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) CooperativeClose(channelID string, closingState paymentstypes.ChannelState, submitter string, currentHeight uint64, settlementFee string) (paymentstypes.SettlementRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	next, settlement, err := paymentstypes.CooperativeClose(k.genesis.State, channelID, closingState, submitter, currentHeight, settlementFee)
	if err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	k.genesis.State = next
	return settlement, nil
}

func (k *Keeper) ReceiverClose(channelID string, claim paymentstypes.UnidirectionalClaim, receiver string, currentHeight uint64, settlementFee string) (paymentstypes.SettlementRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	next, settlement, err := paymentstypes.ReceiverClose(k.genesis.State, channelID, claim, receiver, currentHeight, settlementFee)
	if err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	k.genesis.State = next
	return settlement, nil
}

func (k *Keeper) PayerReclaim(channelID string, payer string, currentHeight uint64, settlementFee string) (paymentstypes.SettlementRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	next, settlement, err := paymentstypes.PayerReclaim(k.genesis.State, channelID, payer, currentHeight, settlementFee)
	if err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	k.genesis.State = next
	return settlement, nil
}

func (k *Keeper) DisputeClose(channelID string, newerState paymentstypes.ChannelState, submitter string, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.DisputeClose(k.genesis.State, channelID, newerState, submitter, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) DisputeChannel(req paymentstypes.ChannelDisputeRequest) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.DisputeChannel(k.genesis.State, req)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) SubmitWatchDispute(submission paymentstypes.WatchDisputeSubmission) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.SubmitWatchDispute(k.genesis.State, submission)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) SubmitFraudProof(channelID string, proof paymentstypes.FraudProof, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.SubmitFraudProof(k.genesis.State, channelID, proof, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) SubmitFraudProofWithPolicy(channelID string, proof paymentstypes.FraudProof, currentHeight uint64, policy paymentstypes.FraudPenaltyPolicy) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.SubmitFraudProofWithPolicy(k.genesis.State, channelID, proof, currentHeight, policy)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) FraudClose(channelID string, currentHeight uint64) (paymentstypes.SettlementRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	next, settlement, err := paymentstypes.FraudClose(k.genesis.State, channelID, currentHeight)
	if err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	k.genesis.State = next
	return settlement, nil
}

func (k *Keeper) FinalizeSettlement(channelID string, currentHeight uint64) (paymentstypes.SettlementRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	next, settlement, err := paymentstypes.FinalizeSettlement(k.genesis.State, channelID, currentHeight)
	if err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	k.genesis.State = next
	return settlement, nil
}

func (k *Keeper) FinalizeSettlementWithRequest(req paymentstypes.FinalSettlementRequest) (paymentstypes.SettlementRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	next, settlement, err := paymentstypes.FinalizeSettlementWithRequest(k.genesis.State, req)
	if err != nil {
		return paymentstypes.SettlementRecord{}, err
	}
	k.genesis.State = next
	return settlement, nil
}

func (k *Keeper) OpenVirtualChannel(vc paymentstypes.VirtualChannel) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.OpenVirtualChannel(k.genesis.State, vc)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) OpenVirtualChannelWithProof(proof paymentstypes.VirtualActivationProof) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.OpenVirtualChannelWithProof(k.genesis.State, proof)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) AcceptVirtualChannelUpdate(vc paymentstypes.VirtualChannel, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.AcceptVirtualChannelUpdate(k.genesis.State, vc, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) SubmitVirtualChannelDispute(proof paymentstypes.VirtualChannelDisputeProof, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.SubmitVirtualChannelDispute(k.genesis.State, proof, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k *Keeper) CloseVirtualChannel(virtualChannelID string, currentHeight uint64) (paymentstypes.VirtualChannel, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.VirtualChannel{}, err
	}
	next, closed, err := paymentstypes.CloseVirtualChannel(k.genesis.State, virtualChannelID, currentHeight)
	if err != nil {
		return paymentstypes.VirtualChannel{}, err
	}
	k.genesis.State = next
	return closed, nil
}

func (k *Keeper) AddSettlementBatch(batch paymentstypes.SettlementBatch) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.AddSettlementBatch(k.genesis.State, batch)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k Keeper) QueryStateHash(channelID string) (paymentstypes.StateHashDebug, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.StateHashDebug{}, err
	}
	return k.genesis.State.StateHashDebug(channelID)
}

func (k Keeper) QueryPendingFinalizationHeight(channelID string) (uint64, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return 0, false, err
	}
	return k.genesis.State.PendingFinalizationHeight(channelID)
}

func (k *Keeper) AdvanceChannelFinality(channelID string, currentHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	next, err := paymentstypes.AdvanceChannelFinality(k.genesis.State, channelID, currentHeight)
	if err != nil {
		return err
	}
	k.genesis.State = next
	return nil
}

func (k Keeper) Channels(req *prototype.PageRequest) ([]paymentstypes.ChannelRecord, prototype.PageResponse, error) {
	channels := k.genesis.State.Export().Channels
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(channels))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]paymentstypes.ChannelRecord, end-start)
	copy(out, channels[start:end])
	return out, res, nil
}

func (k Keeper) Settlements(req *prototype.PageRequest) ([]paymentstypes.SettlementRecord, prototype.PageResponse, error) {
	settlements := k.genesis.State.Export().Settlements
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(settlements))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]paymentstypes.SettlementRecord, end-start)
	copy(out, settlements[start:end])
	return out, res, nil
}

func (k Keeper) RoutePayment(from, to, amount string, currentHeight uint64, maxHops int) ([]paymentstypes.ChannelEdge, error) {
	return paymentstypes.RoutePayment(k.genesis.State, from, to, amount, currentHeight, maxHops)
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
