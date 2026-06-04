package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/dex/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Pool(ctx context.Context, req *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	pool, found, err := k.GetPool(ctx, req.PoolId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "pool not found")
	}
	return &types.QueryPoolResponse{Pool: pool}, nil
}

func (k Keeper) Pools(ctx context.Context, _ *types.QueryPoolsRequest) (*types.QueryPoolsResponse, error) {
	pools, err := k.GetAllPools(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryPoolsResponse{Pools: pools}, nil
}
