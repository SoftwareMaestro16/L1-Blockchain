package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/dynamic-commission/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) SetBaseCommission(ctx context.Context, msg *types.MsgSetBaseCommission) (*types.MsgSetBaseCommissionResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidCommission.Wrap("empty request")
	}
	if err := aetraaddress.ValidateUserAddress("validator address", msg.ValidatorAddress); err != nil {
		return nil, types.ErrUnauthorized.Wrap(err.Error())
	}
	commission, err := m.Keeper.SetBaseCommission(ctx, msg.ValidatorAddress, msg.BaseCommissionBps, msg.Height)
	if err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSetBaseCommission,
		sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
		sdk.NewAttribute(types.AttributeKeyBaseCommissionBps, fmt.Sprintf("%d", commission.BaseCommissionBps)),
		sdk.NewAttribute(types.AttributeKeyEffectiveCommissionBps, fmt.Sprintf("%d", commission.EffectiveCommissionBps)),
	))
	return &types.MsgSetBaseCommissionResponse{Commission: commission}, nil
}

func (m msgServer) UpdateCommissionParams(ctx context.Context, msg *types.MsgUpdateCommissionParams) (*types.MsgUpdateCommissionParamsResponse, error) {
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
		types.EventTypeUpdateCommissionParams,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
	))
	return &types.MsgUpdateCommissionParamsResponse{}, nil
}

func (m msgServer) RecomputeEffectiveCommission(ctx context.Context, msg *types.MsgRecomputeEffectiveCommission) (*types.MsgRecomputeEffectiveCommissionResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidCommission.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	if err := aetraaddress.ValidateUserAddress("validator address", msg.ValidatorAddress); err != nil {
		return nil, types.ErrInvalidCommission.Wrap(err.Error())
	}
	commission, err := m.Keeper.RecomputeEffectiveCommission(ctx, msg.ValidatorAddress, msg.PerformanceScoreBps, msg.ReputationScoreBps, msg.Jailed, msg.Height)
	if err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRecomputeEffective,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
		sdk.NewAttribute(types.AttributeKeyEffectiveCommissionBps, fmt.Sprintf("%d", commission.EffectiveCommissionBps)),
	))
	return &types.MsgRecomputeEffectiveCommissionResponse{Commission: commission}, nil
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
