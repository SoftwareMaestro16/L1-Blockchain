package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/treasury/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) SubmitTreasurySpend(ctx context.Context, msg *types.MsgSubmitTreasurySpend) (*types.MsgSubmitTreasurySpendResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidSpend.Wrap("empty request")
	}
	spend, err := m.Keeper.SubmitSpend(ctx, msg.Proposer, msg.Recipient, msg.Amount, msg.Bucket, msg.Epoch, msg.VestingStartEpoch, msg.VestingEndEpoch, msg.Metadata)
	if err != nil {
		return nil, err
	}
	emitSpendEvent(ctx, types.EventTypeSubmitSpend, "", msg.Proposer, spend)
	return &types.MsgSubmitTreasurySpendResponse{Spend: spend}, nil
}

func (m msgServer) ApproveTreasurySpend(ctx context.Context, msg *types.MsgApproveTreasurySpend) (*types.MsgApproveTreasurySpendResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidSpend.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	spend, err := m.Keeper.ApproveSpend(ctx, msg.SpendId, msg.Metadata)
	if err != nil {
		return nil, err
	}
	emitSpendEvent(ctx, types.EventTypeApproveSpend, msg.Authority, "", spend)
	return &types.MsgApproveTreasurySpendResponse{Spend: spend}, nil
}

func (m msgServer) RejectTreasurySpend(ctx context.Context, msg *types.MsgRejectTreasurySpend) (*types.MsgRejectTreasurySpendResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidSpend.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	spend, err := m.Keeper.RejectSpend(ctx, msg.SpendId, msg.Metadata)
	if err != nil {
		return nil, err
	}
	emitSpendEvent(ctx, types.EventTypeRejectSpend, msg.Authority, "", spend)
	return &types.MsgRejectTreasurySpendResponse{Spend: spend}, nil
}

func (m msgServer) ExecuteTreasurySpend(ctx context.Context, msg *types.MsgExecuteTreasurySpend) (*types.MsgExecuteTreasurySpendResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidSpend.Wrap("empty request")
	}
	if err := m.requireAuthority(msg.Authority); err != nil {
		return nil, err
	}
	spend, err := m.Keeper.ExecuteSpend(ctx, msg.SpendId, msg.Epoch)
	if err != nil {
		return nil, err
	}
	emitSpendEvent(ctx, types.EventTypeExecuteSpend, msg.Authority, "", spend)
	return &types.MsgExecuteTreasurySpendResponse{Spend: spend}, nil
}

func (m msgServer) CancelTreasurySpend(ctx context.Context, msg *types.MsgCancelTreasurySpend) (*types.MsgCancelTreasurySpendResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidSpend.Wrap("empty request")
	}
	if err := aetraaddress.ValidateAuthorityAddress("actor", msg.Actor); err != nil {
		return nil, types.ErrUnauthorized.Wrap(err.Error())
	}
	spend, err := m.Keeper.CancelSpend(ctx, msg.SpendId, msg.Actor, msg.Metadata)
	if err != nil {
		return nil, err
	}
	emitSpendEvent(ctx, types.EventTypeCancelSpend, "", msg.Actor, spend)
	return &types.MsgCancelTreasurySpendResponse{Spend: spend}, nil
}

func (m msgServer) UpdateTreasuryParams(ctx context.Context, msg *types.MsgUpdateTreasuryParams) (*types.MsgUpdateTreasuryParamsResponse, error) {
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
	return &types.MsgUpdateTreasuryParamsResponse{}, nil
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

func emitSpendEvent(ctx context.Context, eventType, authority, actor string, spend types.TreasurySpend) {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeKeySpendID, fmt.Sprintf("%d", spend.Id)),
		sdk.NewAttribute(types.AttributeKeyRecipient, spend.Recipient),
		sdk.NewAttribute(types.AttributeKeyAmount, spend.Amount.String()),
		sdk.NewAttribute(types.AttributeKeyBucket, spend.Bucket),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", spend.Epoch)),
	}
	if authority != "" {
		attrs = append(attrs, sdk.NewAttribute(types.AttributeKeyAuthority, authority))
	}
	if actor != "" {
		attrs = append(attrs, sdk.NewAttribute(types.AttributeKeyActor, actor))
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(eventType, attrs...))
}
