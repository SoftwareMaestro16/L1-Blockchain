package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Denom(ctx context.Context, req *types.QueryDenomRequest) (*types.QueryDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	meta, found, err := k.GetDenom(ctx, req.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "denom not found")
	}
	return &types.QueryDenomResponse{Metadata: meta}, nil
}

func (k Keeper) Denoms(ctx context.Context, req *types.QueryDenomsRequest) (*types.QueryDenomsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	denoms, err := k.GetAllDenoms(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryDenomsResponse{Denoms: denoms}, nil
}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}
