package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/fees/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidParams.Wrap("empty request")
	}
	if err := aetraaddress.ValidateAuthorityAddress("authority", msg.Authority); err != nil {
		return nil, types.ErrUnauthorized.Wrap(err.Error())
	}
	if msg.Authority != m.Authority() {
		return nil, types.ErrUnauthorized.Wrap("invalid authority")
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeUpdateParams,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		sdk.NewAttribute(types.AttributeKeyAllowedFeeDenom, msg.Params.AllowedFeeDenoms[0]),
		sdk.NewAttribute(types.AttributeKeyValidatorRewardsRatio, msg.Params.ValidatorRewardsRatio),
		sdk.NewAttribute(types.AttributeKeyCommunityPoolRatio, msg.Params.CommunityPoolRatio),
	))
	return &types.MsgUpdateParamsResponse{}, nil
}
