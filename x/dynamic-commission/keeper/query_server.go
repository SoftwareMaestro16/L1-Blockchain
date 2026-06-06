package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/dynamic-commission/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ValidatorCommission(ctx context.Context, req *types.QueryValidatorCommissionRequest) (*types.QueryValidatorCommissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	commission, found, err := k.GetValidatorCommission(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrNotFound.Error())
	}
	return &types.QueryValidatorCommissionResponse{Commission: commission}, nil
}

func (k Keeper) CommissionHistory(ctx context.Context, req *types.QueryCommissionHistoryRequest) (*types.QueryCommissionHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	history, err := k.GetCommissionHistory(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryCommissionHistoryResponse{History: history}, nil
}

func (k Keeper) CommissionParams(ctx context.Context, req *types.QueryCommissionParamsRequest) (*types.QueryCommissionParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryCommissionParamsResponse{Params: params}, nil
}
