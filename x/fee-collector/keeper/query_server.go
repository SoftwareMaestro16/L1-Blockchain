package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/fee-collector/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) FeeCollector(ctx context.Context, req *types.QueryFeeCollectorRequest) (*types.QueryFeeCollectorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	balances, err := k.GetFeeBalances(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pending, err := k.GetPendingDistribution(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeeCollectorResponse{
		ModuleName:		types.CollectorModuleName,
		ModuleAccount:		k.ModuleAccountAddress(),
		Balances:		balances,
		PendingDistribution:	pending,
	}, nil
}

func (k Keeper) FeeBalances(ctx context.Context, req *types.QueryFeeBalancesRequest) (*types.QueryFeeBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	balances, err := k.GetFeeBalances(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeeBalancesResponse{Balances: balances}, nil
}

func (k Keeper) FeeDistribution(ctx context.Context, req *types.QueryFeeDistributionRequest) (*types.QueryFeeDistributionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pending, err := k.GetPendingDistribution(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeeDistributionResponse{Params: params, PendingDistribution: pending}, nil
}

func (k Keeper) FeeHistory(ctx context.Context, req *types.QueryFeeHistoryRequest) (*types.QueryFeeHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Epoch != 0 {
		entry, found, err := k.GetFeeHistory(ctx, req.Epoch)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return &types.QueryFeeHistoryResponse{History: []types.FeeHistoryEntry{}}, nil
		}
		return &types.QueryFeeHistoryResponse{History: []types.FeeHistoryEntry{entry}}, nil
	}
	history, err := k.GetAllFeeHistory(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryFeeHistoryResponse{History: history}, nil
}
