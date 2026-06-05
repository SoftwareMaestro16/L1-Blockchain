package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	txutil "github.com/sovereign-l1/l1/x/internal/tx"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) CreateDenom(ctx context.Context, msg *types.MsgCreateDenom) (*types.MsgCreateDenomResponse, error) {
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.DenomCreationEnabled {
		return nil, types.ErrOperationDisabled.Wrap("denom creation is disabled")
	}
	denom, err := m.FullDenom(ctx, msg.Creator, msg.Subdenom)
	if err != nil {
		return nil, err
	}
	creator, err := orbitaladdress.ParseAccAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	if _, found, err := m.GetDenom(ctx, denom); err != nil {
		return nil, err
	} else if found {
		return nil, types.ErrDenomExists.Wrap(denom)
	}
	creatorText := orbitaladdress.FormatAccAddress(creator)
	meta := types.DenomAuthorityMetadata{Denom: denom, Admin: creatorText}
	if err := m.SetDenom(ctx, meta); err != nil {
		return nil, err
	}
	m.bankKeeper.SetDenomMetaData(ctx, BankMetadata(denom))
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCreateDenom,
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyCreator, creatorText),
		sdk.NewAttribute(types.AttributeKeyAdmin, creatorText),
	))
	return &types.MsgCreateDenomResponse{NewTokenDenom: denom}, nil
}

func (m msgServer) Mint(ctx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.MintingEnabled {
		return nil, types.ErrOperationDisabled.Wrap("minting is disabled")
	}
	sender, err := orbitaladdress.ParseAccAddress(msg.Sender)
	if err != nil {
		return nil, err
	}
	senderText := orbitaladdress.FormatAccAddress(sender)
	meta, found, err := m.GetDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Amount.Denom)
	}
	if meta.Admin != senderText {
		return nil, types.ErrUnauthorized.Wrap("only denom admin can mint")
	}
	to, err := orbitaladdress.ParseAccAddress(msg.MintToAddress)
	if err != nil {
		return nil, err
	}
	toText := orbitaladdress.FormatAccAddress(to)
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return nil, types.ErrInvalidDenom.Wrap("mint amount must be positive")
	}
	coins := sdk.NewCoins(msg.Amount)
	if err := txutil.AtomicStateChange(ctx, func(cacheCtx context.Context) error {
		if err := m.bankKeeper.MintCoins(cacheCtx, types.ModuleName, coins); err != nil {
			return err
		}
		if err := m.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, to, coins); err != nil {
			return err
		}
		sdk.UnwrapSDKContext(cacheCtx).EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyDenom, msg.Amount.Denom),
			sdk.NewAttribute(types.AttributeKeySender, senderText),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyMintToAddress, toText),
		))
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.MsgMintResponse{}, nil
}

func (m msgServer) Burn(ctx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.BurningEnabled {
		return nil, types.ErrOperationDisabled.Wrap("burning is disabled")
	}
	sender, err := orbitaladdress.ParseAccAddress(msg.Sender)
	if err != nil {
		return nil, err
	}
	senderText := orbitaladdress.FormatAccAddress(sender)
	meta, found, err := m.GetDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Amount.Denom)
	}
	if meta.Admin != senderText {
		return nil, types.ErrUnauthorized.Wrap("only denom admin can burn")
	}
	from, err := orbitaladdress.ParseAccAddress(msg.BurnFromAddress)
	if err != nil {
		return nil, err
	}
	fromText := orbitaladdress.FormatAccAddress(from)
	if !from.Equals(sender) {
		return nil, types.ErrUnauthorized.Wrap("burn_from_address must match sender")
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return nil, types.ErrInvalidDenom.Wrap("burn amount must be positive")
	}
	coins := sdk.NewCoins(msg.Amount)
	if err := txutil.AtomicStateChange(ctx, func(cacheCtx context.Context) error {
		if err := m.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, from, types.ModuleName, coins); err != nil {
			return err
		}
		if err := m.bankKeeper.BurnCoins(cacheCtx, types.ModuleName, coins); err != nil {
			return err
		}
		sdk.UnwrapSDKContext(cacheCtx).EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeBurn,
			sdk.NewAttribute(types.AttributeKeyDenom, msg.Amount.Denom),
			sdk.NewAttribute(types.AttributeKeySender, senderText),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyBurnFromAddress, fromText),
		))
		return nil
	}); err != nil {
		return nil, err
	}
	return &types.MsgBurnResponse{}, nil
}

func (m msgServer) ChangeAdmin(ctx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	sender, err := orbitaladdress.ParseAccAddress(msg.Sender)
	if err != nil {
		return nil, err
	}
	senderText := orbitaladdress.FormatAccAddress(sender)
	meta, found, err := m.GetDenom(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Denom)
	}
	if meta.Admin != senderText {
		return nil, types.ErrUnauthorized.Wrap("only denom admin can change admin")
	}
	newAdmin, err := orbitaladdress.ParseAccAddress(msg.NewAdmin)
	if err != nil {
		return nil, err
	}
	newAdminText := orbitaladdress.FormatAccAddress(newAdmin)
	meta.Admin = newAdminText
	if err := m.SetDenom(ctx, meta); err != nil {
		return nil, err
	}
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeChangeAdmin,
		sdk.NewAttribute(types.AttributeKeyDenom, msg.Denom),
		sdk.NewAttribute(types.AttributeKeySender, senderText),
		sdk.NewAttribute(types.AttributeKeyNewAdmin, newAdminText),
	))
	return &types.MsgChangeAdminResponse{}, nil
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
