package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/constitution/types"
)

var _ types.MsgServer = grpcMsgServer{}
var _ types.QueryServer = grpcQueryServer{}

type grpcMsgServer struct{ keeper *Keeper }
type grpcQueryServer struct{ keeper *Keeper }

func NewGRPCMsgServer(k *Keeper) types.MsgServer	{ return grpcMsgServer{keeper: k} }
func NewGRPCQueryServer(k *Keeper) types.QueryServer	{ return grpcQueryServer{keeper: k} }

func (s grpcMsgServer) ProposeConstitutionAmendment(ctx context.Context, msg *types.MsgProposeConstitutionAmendment) (*types.MsgProposeConstitutionAmendmentResponse, error) {
	if msg == nil {
		return nil, errors.New("empty constitution amendment proposal")
	}
	amendment, err := s.keeper.ProposeConstitutionAmendment(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	return &types.MsgProposeConstitutionAmendmentResponse{Amendment: amendment}, nil
}

func (s grpcMsgServer) VoteConstitutionAmendment(ctx context.Context, msg *types.MsgVoteConstitutionAmendment) (*types.MsgVoteConstitutionAmendmentResponse, error) {
	if msg == nil {
		return nil, errors.New("empty constitution amendment vote")
	}
	amendment, err := s.keeper.VoteConstitutionAmendment(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	return &types.MsgVoteConstitutionAmendmentResponse{Amendment: amendment}, nil
}

func (s grpcMsgServer) ExecuteConstitutionAmendment(ctx context.Context, msg *types.MsgExecuteConstitutionAmendment) (*types.MsgExecuteConstitutionAmendmentResponse, error) {
	if msg == nil {
		return nil, errors.New("empty constitution amendment execution")
	}
	constitution, amendment, err := s.keeper.ExecuteConstitutionAmendment(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	return &types.MsgExecuteConstitutionAmendmentResponse{Constitution: constitution, Amendment: amendment}, nil
}

func (s grpcMsgServer) CancelConstitutionAmendment(ctx context.Context, msg *types.MsgCancelConstitutionAmendment) (*types.MsgCancelConstitutionAmendmentResponse, error) {
	if msg == nil {
		return nil, errors.New("empty constitution amendment cancellation")
	}
	amendment, err := s.keeper.CancelConstitutionAmendment(*msg, sdkHeight(ctx))
	if err != nil {
		return nil, err
	}
	return &types.MsgCancelConstitutionAmendmentResponse{Amendment: amendment}, nil
}

func (s grpcQueryServer) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: s.keeper.ExportGenesis().Params}, nil
}

func (s grpcQueryServer) Constitution(context.Context, *types.QueryConstitutionRequest) (*types.QueryConstitutionResponse, error) {
	return &types.QueryConstitutionResponse{Constitution: s.keeper.Constitution()}, nil
}

func (s grpcQueryServer) PendingAmendments(_ context.Context, req *types.QueryPendingAmendmentsRequest) (*types.QueryPendingAmendmentsResponse, error) {
	if req == nil {
		req = &types.QueryPendingAmendmentsRequest{}
	}
	amendments, err := s.keeper.PendingAmendments()
	if err != nil {
		return nil, err
	}
	page, next := paginateAmendments(amendments, req.Offset, req.Limit)
	return &types.QueryPendingAmendmentsResponse{Amendments: page, Next: next, Total: uint64(len(amendments))}, nil
}

func (s grpcQueryServer) Amendment(_ context.Context, req *types.QueryAmendmentRequest) (*types.QueryAmendmentResponse, error) {
	if req == nil {
		return nil, errors.New("empty constitution amendment query")
	}
	amendment, found, err := s.keeper.Amendment(req.ID)
	if err != nil {
		return nil, err
	}
	return &types.QueryAmendmentResponse{Amendment: amendment, Found: found}, nil
}

func (s grpcQueryServer) ProtectedLimits(context.Context, *types.QueryProtectedLimitsRequest) (*types.QueryProtectedLimitsResponse, error) {
	return &types.QueryProtectedLimitsResponse{Limits: s.keeper.ProtectedLimits()}, nil
}

func sdkHeight(ctx context.Context) uint64 {
	sdkCtx, ok := sdkContext(ctx)
	if ok {
		if sdkCtx.BlockHeight() < 0 {
			return 0
		}
		return uint64(sdkCtx.BlockHeight())
	}
	return 0
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

func paginateAmendments(amendments []types.Amendment, offset, limit uint64) ([]types.Amendment, uint64) {
	if limit == 0 || limit > 100 {
		limit = 100
	}
	if offset >= uint64(len(amendments)) {
		return []types.Amendment{}, 0
	}
	end := offset + limit
	if end > uint64(len(amendments)) {
		end = uint64(len(amendments))
	}
	next := uint64(0)
	if end < uint64(len(amendments)) {
		next = end
	}
	return amendments[int(offset):int(end)], next
}
