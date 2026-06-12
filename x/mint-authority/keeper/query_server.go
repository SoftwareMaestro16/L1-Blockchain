package keeper

import (
	"context"

	"github.com/sovereign-l1/l1/x/mint-authority/types"
	mintauthoritypb "github.com/sovereign-l1/l1/x/mint-authority/types/mintauthoritypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ mintauthoritypb.QueryServer = Keeper{}

func (k Keeper) MintAuthority(ctx context.Context, _ *mintauthoritypb.QueryMintAuthorityRequest) (*mintauthoritypb.QueryMintAuthorityResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	params, registration, callers := types.QueryMintAuthority(state)
	out, err := mustJSON(struct {
		Params		types.MintAuthorityParams
		Registration	types.SystemAccountRegistration
		Callers		[]types.AllowedCaller
	}{params, registration, callers})
	return &mintauthoritypb.QueryMintAuthorityResponse{AuthorityJson: out}, err
}

func (k Keeper) MintedByEpoch(ctx context.Context, req *mintauthoritypb.QueryMintedByEpochRequest) (*mintauthoritypb.QueryMintedByEpochResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryMintedByEpoch(state, types.QueryMintedByEpochRequest{Epoch: req.Epoch, Denom: req.Denom}))
	return &mintauthoritypb.QueryMintedByEpochResponse{MintedJson: out}, err
}

func (k Keeper) MintedLifetime(ctx context.Context, req *mintauthoritypb.QueryMintedLifetimeRequest) (*mintauthoritypb.QueryMintedLifetimeResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryMintedLifetime(state, req.Denom))
	return &mintauthoritypb.QueryMintedLifetimeResponse{MintedJson: out}, err
}

func (k Keeper) MintCaps(ctx context.Context, _ *mintauthoritypb.QueryMintCapsRequest) (*mintauthoritypb.QueryMintCapsResponse, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out, err := mustJSON(types.QueryMintCaps(state))
	return &mintauthoritypb.QueryMintCapsResponse{CapsJson: out}, err
}
