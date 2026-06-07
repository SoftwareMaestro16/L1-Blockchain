package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/emissions/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) UpdateEmissionsParams(ctx context.Context, msg *types.MsgUpdateEmissionsParams) (*types.MsgUpdateEmissionsParamsResponse, error) {
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
	return &types.MsgUpdateEmissionsParamsResponse{}, nil
}

func (m msgServer) FinalizeEmissionEpoch(ctx context.Context, msg *types.MsgFinalizeEmissionEpoch) (*types.MsgFinalizeEmissionEpochResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidEpoch.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	record, err := m.Keeper.FinalizeEmissionEpoch(ctx, msg.Epoch, msg.StakingRatioBps)
	if err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeFinalizeEpoch,
		sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", record.Epoch)),
		sdk.NewAttribute(types.AttributeKeyInflationBps, fmt.Sprintf("%d", record.InflationBps)),
		sdk.NewAttribute(types.AttributeKeyAmount, record.EmissionAmount.String()),
	))
	return &types.MsgFinalizeEmissionEpochResponse{EmissionEpoch: record}, nil
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
