package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/treasury/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) TreasuryBalance(ctx context.Context, req *types.QueryTreasuryBalanceRequest) (*types.QueryTreasuryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := k.SyncIncomingFunds(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	allocations, err := k.GetAllocations(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	addr := k.accountKeeper.GetModuleAddress(params.TreasuryModule)
	if addr == nil {
		return nil, status.Error(codes.Internal, "treasury module account not configured")
	}
	return &types.QueryTreasuryBalanceResponse{
		ModuleAccount:		aetraaddress.FormatAccAddress(addr),
		BankBalance:		k.bankKeeper.GetAllBalances(ctx, addr),
		AccountingBalance:	allocations.AccountingBalance(),
	}, nil
}

func (k Keeper) TreasuryAllocations(ctx context.Context, req *types.QueryTreasuryAllocationsRequest) (*types.QueryTreasuryAllocationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := k.SyncIncomingFunds(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	allocations, err := k.GetAllocations(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTreasuryAllocationsResponse{Allocations: allocations}, nil
}

func (k Keeper) TreasurySpend(ctx context.Context, req *types.QueryTreasurySpendRequest) (*types.QueryTreasurySpendResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	spend, found, err := k.GetSpend(ctx, req.SpendId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrNotFound.Error())
	}
	return &types.QueryTreasurySpendResponse{Spend: spend}, nil
}

func (k Keeper) TreasurySpends(ctx context.Context, req *types.QueryTreasurySpendsRequest) (*types.QueryTreasurySpendsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	spends, err := k.GetAllSpends(ctx, req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTreasurySpendsResponse{Spends: spends}, nil
}

func (k Keeper) TreasuryParams(ctx context.Context, req *types.QueryTreasuryParamsRequest) (*types.QueryTreasuryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryTreasuryParamsResponse{Params: params}, nil
}
