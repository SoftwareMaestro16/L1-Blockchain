package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/stake-concentration/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) UpdateConcentrationParams(ctx context.Context, msg *types.MsgUpdateConcentrationParams) (*types.MsgUpdateConcentrationParamsResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidParams.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	if err := m.Keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
	))
	return &types.MsgUpdateConcentrationParamsResponse{}, nil
}

func (m msgServer) RecomputeConcentration(ctx context.Context, msg *types.MsgRecomputeConcentration) (*types.MsgRecomputeConcentrationResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidConcentration.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	network, err := m.Keeper.RecomputeConcentration(ctx, msg.Epoch, msg.ValidatorSet)
	if err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRecompute,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", network.Epoch)),
		sdk.NewAttribute(types.AttributeKeyMaxPower, fmt.Sprintf("%d", network.MaxValidatorPowerBps)),
	))
	return &types.MsgRecomputeConcentrationResponse{Network: network}, nil
}

func (m msgServer) requireAuthority(authority string) error {
	if err := aetraaddress.ValidateAuthorityAddress("authority", authority); err != nil {
		return types.ErrUnauthorized.Wrap(err.Error())
	}
	if authority != m.Authority() {
		return types.ErrUnauthorized.Wrap("invalid authority")
	}
	return nil
}
