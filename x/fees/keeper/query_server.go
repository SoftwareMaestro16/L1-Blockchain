package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sovereign-l1/l1/x/fees/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) Accounting(ctx context.Context, req *types.QueryAccountingRequest) (*types.QueryAccountingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	state, err := k.GetProtocolFeeState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryAccountingResponse{ProtocolFeeState: state}, nil
}

func (k Keeper) ModuleBalances(ctx context.Context, req *types.QueryModuleBalancesRequest) (*types.QueryModuleBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	balances, err := k.GetModuleBalances(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryModuleBalancesResponse{Balances: balances}, nil
}

func (k Keeper) NetworkLoad(ctx context.Context, req *types.QueryNetworkLoadRequest) (*types.QueryNetworkLoadResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	blockGasConsumed := uint64(0)
	if sdkCtx.BlockGasMeter() != nil {
		blockGasConsumed = sdkCtx.BlockGasMeter().GasConsumed()
	}
	utilization := types.BlockUtilizationBps(blockGasConsumed, 0, params.MaxBlockGas)
	return &types.QueryNetworkLoadResponse{
		BlockGasConsumed:	blockGasConsumed,
		MaxBlockGas:		params.MaxBlockGas,
		UtilizationBps:		utilization,
		Congested:		utilization >= params.CongestionThresholdBps,
	}, nil
}

func (k Keeper) EstimateFee(ctx context.Context, req *types.QueryEstimateFeeRequest) (*types.QueryEstimateFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockGasConsumed := uint64(0)
	if sdkCtx.BlockGasMeter() != nil {
		blockGasConsumed = sdkCtx.BlockGasMeter().GasConsumed()
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	quote, err := types.QuoteFee(params, req.GasLimit, blockGasConsumed)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &types.QueryEstimateFeeResponse{
		RequiredFee:	quote.RequiredFee.String(),
		BaseFee:	quote.BaseFee.String(),
		MaxFee:		quote.MaxFee.String(),
		UtilizationBps:	quote.UtilizationBps,
		Congested:	quote.Congested,
		AtHardCap:	quote.AtHardCap,
	}, nil
}
