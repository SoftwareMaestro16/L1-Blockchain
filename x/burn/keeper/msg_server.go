package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/burn/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) BurnProtocolCoins(ctx context.Context, msg *types.MsgBurnProtocolCoins) (*types.MsgBurnProtocolCoinsResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidBurn.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	record, err := m.Keeper.BurnProtocolCoins(ctx, msg.SourceModule, msg.Amount, msg.Epoch, msg.Reason)
	if err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeBurnProtocolCoins,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		sdk.NewAttribute(types.AttributeKeyModule, msg.SourceModule),
		sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", msg.Epoch)),
		sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
	))
	return &types.MsgBurnProtocolCoinsResponse{Burn: record}, nil
}

func (m msgServer) BurnUserCoins(ctx context.Context, msg *types.MsgBurnUserCoins) (*types.MsgBurnUserCoinsResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidBurn.Wrap("empty request")
	}
	if err := aetraaddress.ValidateUserAddress("burner", msg.Burner); err != nil {
		return nil, types.ErrInvalidBurn.Wrap(err.Error())
	}
	burner, err := aetraaddress.ParseAccAddress(msg.Burner)
	if err != nil {
		return nil, types.ErrInvalidBurn.Wrap(err.Error())
	}
	record, err := m.Keeper.BurnUserCoins(ctx, burner, msg.Amount, msg.Epoch, msg.Reason)
	if err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeBurnUserCoins,
		sdk.NewAttribute(types.AttributeKeyBurner, msg.Burner),
		sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", msg.Epoch)),
		sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
	))
	return &types.MsgBurnUserCoinsResponse{Burn: record}, nil
}

func (m msgServer) UpdateBurnParams(ctx context.Context, msg *types.MsgUpdateBurnParams) (*types.MsgUpdateBurnParamsResponse, error) {
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
		types.EventTypeUpdateBurnParams,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
	))
	return &types.MsgUpdateBurnParamsResponse{}, nil
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
