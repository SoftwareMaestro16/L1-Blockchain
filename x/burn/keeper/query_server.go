package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/burn/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) BurnedByDenom(ctx context.Context, req *types.QueryBurnedByDenomRequest) (*types.QueryBurnedByDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom != "" {
		entry, found, err := k.GetBurnedDenomEntry(ctx, req.Denom)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return &types.QueryBurnedByDenomResponse{BurnedByDenom: []types.BurnedByDenomEntry{}}, nil
		}
		return &types.QueryBurnedByDenomResponse{BurnedByDenom: []types.BurnedByDenomEntry{entry}}, nil
	}
	entries, err := k.GetAllBurnedByDenom(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryBurnedByDenomResponse{BurnedByDenom: entries}, nil
}

func (k Keeper) BurnedByEpoch(ctx context.Context, req *types.QueryBurnedByEpochRequest) (*types.QueryBurnedByEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Epoch != 0 {
		entry, found, err := k.GetBurnedEpochEntry(ctx, req.Epoch)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return &types.QueryBurnedByEpochResponse{BurnedByEpoch: []types.BurnedByEpochEntry{}}, nil
		}
		return &types.QueryBurnedByEpochResponse{BurnedByEpoch: []types.BurnedByEpochEntry{entry}}, nil
	}
	entries, err := k.GetAllBurnedByEpoch(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryBurnedByEpochResponse{BurnedByEpoch: entries}, nil
}

func (k Keeper) BurnParams(ctx context.Context, req *types.QueryBurnParamsRequest) (*types.QueryBurnParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryBurnParamsResponse{Params: params}, nil
}
