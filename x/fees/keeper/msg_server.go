package keeper

import (
	"context"

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
	if msg.Authority != m.Authority() {
		return nil, types.ErrUnauthorized.Wrap("invalid authority")
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
