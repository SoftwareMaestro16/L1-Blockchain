package keeper

import (
	"context"
	"errors"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/config/types"
)

const (
	eventTypeConfigChangeSubmitted	= "config_change_submitted"
	eventTypeConfigChangeApproved	= "config_change_approved"
	eventTypeConfigChangeRejected	= "config_change_rejected"
	eventTypeConfigChangeExecuted	= "config_change_executed"
	eventTypeConfigChangeCancelled	= "config_change_cancelled"
)

var _ types.MsgServer = grpcMsgServer{}
var _ types.QueryServer = grpcQueryServer{}

type grpcMsgServer struct{ keeper *Keeper }
type grpcQueryServer struct{ keeper *Keeper }

func NewGRPCMsgServer(k *Keeper) types.MsgServer	{ return grpcMsgServer{keeper: k} }
func NewGRPCQueryServer(k *Keeper) types.QueryServer	{ return grpcQueryServer{keeper: k} }

func (s grpcMsgServer) SubmitConfigChange(ctx context.Context, msg *types.MsgSubmitConfigChange) (*types.MsgSubmitConfigChangeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty config change submission")
	}
	height := sdkHeight(ctx)
	change, err := s.keeper.SubmitConfigChange(*msg, height)
	if err != nil {
		return nil, err
	}
	emitConfigEvent(ctx, eventTypeConfigChangeSubmitted, change)
	return &types.MsgSubmitConfigChangeResponse{Change: change}, nil
}

func (s grpcMsgServer) ApproveConfigChange(ctx context.Context, msg *types.MsgApproveConfigChange) (*types.MsgApproveConfigChangeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty config change approval")
	}
	change, err := s.keeper.ApproveConfigChange(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	emitConfigEvent(ctx, eventTypeConfigChangeApproved, change)
	return &types.MsgApproveConfigChangeResponse{Change: change}, nil
}

func (s grpcMsgServer) RejectConfigChange(ctx context.Context, msg *types.MsgRejectConfigChange) (*types.MsgRejectConfigChangeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty config change rejection")
	}
	change, err := s.keeper.RejectConfigChange(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	emitConfigEvent(ctx, eventTypeConfigChangeRejected, change)
	return &types.MsgRejectConfigChangeResponse{Change: change}, nil
}

func (s grpcMsgServer) ExecuteConfigChange(ctx context.Context, msg *types.MsgExecuteConfigChange) (*types.MsgExecuteConfigChangeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty config change execution")
	}
	entry, change, err := s.keeper.ExecuteConfigChange(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	emitConfigEvent(ctx, eventTypeConfigChangeExecuted, change)
	return &types.MsgExecuteConfigChangeResponse{Entry: entry, Change: change}, nil
}

func (s grpcMsgServer) CancelConfigChange(ctx context.Context, msg *types.MsgCancelConfigChange) (*types.MsgCancelConfigChangeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty config change cancellation")
	}
	change, err := s.keeper.CancelConfigChange(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	emitConfigEvent(ctx, eventTypeConfigChangeCancelled, change)
	return &types.MsgCancelConfigChangeResponse{Change: change}, nil
}

func (s grpcQueryServer) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: s.keeper.ExportGenesis().Params}, nil
}

func (s grpcQueryServer) Entries(_ context.Context, req *types.QueryEntriesRequest) (*types.QueryEntriesResponse, error) {
	if req == nil {
		req = &types.QueryEntriesRequest{}
	}
	entries, err := s.keeper.Entries()
	if err != nil {
		return nil, err
	}
	page, next := paginateEntries(entries, req.Offset, req.Limit)
	return &types.QueryEntriesResponse{Entries: page, Next: next, Total: uint64(len(entries))}, nil
}

func (s grpcQueryServer) Entry(_ context.Context, req *types.QueryEntryRequest) (*types.QueryEntryResponse, error) {
	if req == nil {
		return nil, errors.New("empty config entry query")
	}
	entry, found, err := s.keeper.Entry(req.Key)
	if err != nil {
		return nil, err
	}
	return &types.QueryEntryResponse{Entry: entry, Found: found}, nil
}

func (s grpcQueryServer) PendingChanges(_ context.Context, req *types.QueryPendingChangesRequest) (*types.QueryPendingChangesResponse, error) {
	if req == nil {
		req = &types.QueryPendingChangesRequest{}
	}
	changes, err := s.keeper.PendingConfigChanges()
	if err != nil {
		return nil, err
	}
	page, next := paginateChanges(changes, req.Offset, req.Limit)
	return &types.QueryPendingChangesResponse{Changes: page, Next: next, Total: uint64(len(changes))}, nil
}

func (s grpcQueryServer) Change(_ context.Context, req *types.QueryChangeRequest) (*types.QueryChangeResponse, error) {
	if req == nil {
		return nil, errors.New("empty config change query")
	}
	change, found, err := s.keeper.ConfigChange(req.ID)
	if err != nil {
		return nil, err
	}
	return &types.QueryChangeResponse{Change: change, Found: found}, nil
}

func (s grpcQueryServer) EffectiveParams(ctx context.Context, req *types.QueryEffectiveParamsRequest) (*types.QueryEffectiveParamsResponse, error) {
	if req == nil {
		req = &types.QueryEffectiveParamsRequest{}
	}
	height := req.Height
	if height == 0 {
		height = sdkHeight(ctx)
	}
	entries, err := s.keeper.EffectiveEntries(height)
	if err != nil {
		return nil, err
	}
	page, next := paginateEntries(entries, req.Offset, req.Limit)
	return &types.QueryEffectiveParamsResponse{Entries: page, Next: next, Total: uint64(len(entries))}, nil
}

func sdkHeight(ctx context.Context) int64 {
	sdkCtx, ok := sdkContext(ctx)
	if ok {
		return sdkCtx.BlockHeight()
	}
	return 0
}

func emitConfigEvent(ctx context.Context, eventType string, change types.ConfigChange) {
	sdkCtx, ok := sdkContext(ctx)
	if !ok || sdkCtx.EventManager() == nil {
		return
	}
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		eventType,
		sdk.NewAttribute("id", change.ID),
		sdk.NewAttribute("key", change.Key),
		sdk.NewAttribute("status", change.Status),
		sdk.NewAttribute("activation_height", int64String(change.ActivationHeight)),
		sdk.NewAttribute("activation_epoch", uint64String(change.ActivationEpoch)),
	))
}

func sdkContext(ctx context.Context) (out sdk.Context, ok bool) {
	if sdkCtx, direct := ctx.(sdk.Context); direct {
		return sdkCtx, true
	}
	if ctx == nil {
		return sdk.Context{}, false
	}
	defer func() {
		if recover() != nil {
			out = sdk.Context{}
			ok = false
		}
	}()
	return sdk.UnwrapSDKContext(ctx), true
}

func paginateEntries(entries []types.ConfigEntry, offset, limit uint64) ([]types.ConfigEntry, uint64) {
	if limit == 0 || limit > 100 {
		limit = 100
	}
	if offset >= uint64(len(entries)) {
		return []types.ConfigEntry{}, 0
	}
	end := offset + limit
	if end > uint64(len(entries)) {
		end = uint64(len(entries))
	}
	next := uint64(0)
	if end < uint64(len(entries)) {
		next = end
	}
	return entries[int(offset):int(end)], next
}

func paginateChanges(changes []types.ConfigChange, offset, limit uint64) ([]types.ConfigChange, uint64) {
	if limit == 0 || limit > 100 {
		limit = 100
	}
	if offset >= uint64(len(changes)) {
		return []types.ConfigChange{}, 0
	}
	end := offset + limit
	if end > uint64(len(changes)) {
		end = uint64(len(changes))
	}
	next := uint64(0)
	if end < uint64(len(changes)) {
		next = end
	}
	return changes[int(offset):int(end)], next
}

func int64String(value int64) string	{ return strconv.FormatInt(value, 10) }
func uint64String(value uint64) string	{ return strconv.FormatUint(value, 10) }
