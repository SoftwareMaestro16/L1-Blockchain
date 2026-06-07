package keeper

import (
	"context"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/stake-concentration/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ValidatorConcentration(ctx context.Context, req *types.QueryValidatorConcentrationRequest) (*types.QueryValidatorConcentrationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := aetraaddress.ValidateUserAddress("operator_address", req.OperatorAddress); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	validator, found, err := k.GetValidatorConcentration(ctx, req.OperatorAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrNotFound.Error())
	}
	return &types.QueryValidatorConcentrationResponse{Concentration: validator}, nil
}

func (k Keeper) NetworkConcentration(ctx context.Context, req *types.QueryNetworkConcentrationRequest) (*types.QueryNetworkConcentrationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	network, err := k.GetNetworkConcentration(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryNetworkConcentrationResponse{Network: network}, nil
}

func (k Keeper) ConcentrationParams(ctx context.Context, req *types.QueryConcentrationParamsRequest) (*types.QueryConcentrationParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryConcentrationParamsResponse{Params: params}, nil
}
