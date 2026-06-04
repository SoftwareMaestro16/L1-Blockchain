package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	if _, found, err := m.GetDenom(ctx, denom); err != nil {
		return nil, err
	} else if found {
		return nil, types.ErrDenomExists.Wrap(denom)
	}
	meta := types.DenomAuthorityMetadata{Denom: denom, Admin: creator.String()}
	if err := m.SetDenom(ctx, meta); err != nil {
		return nil, err
	}
	m.bankKeeper.SetDenomMetaData(ctx, BankMetadata(denom))
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
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	meta, found, err := m.GetDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Amount.Denom)
	}
	if meta.Admin != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only denom admin can mint")
	}
	to, err := sdk.AccAddressFromBech32(msg.MintToAddress)
	if err != nil {
		return nil, err
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return nil, types.ErrInvalidDenom.Wrap("mint amount must be positive")
	}
	coins := sdk.NewCoins(msg.Amount)
	if err := m.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, to, coins); err != nil {
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
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	meta, found, err := m.GetDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Amount.Denom)
	}
	if meta.Admin != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only denom admin can burn")
	}
	from, err := sdk.AccAddressFromBech32(msg.BurnFromAddress)
	if err != nil {
		return nil, err
	}
	if !from.Equals(sender) {
		return nil, types.ErrUnauthorized.Wrap("burn_from_address must match sender")
	}
	if !msg.Amount.IsValid() || !msg.Amount.IsPositive() {
		return nil, types.ErrInvalidDenom.Wrap("burn amount must be positive")
	}
	coins := sdk.NewCoins(msg.Amount)
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleName, coins); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
		return nil, err
	}
	return &types.MsgBurnResponse{}, nil
}

func (m msgServer) ChangeAdmin(ctx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	meta, found, err := m.GetDenom(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrDenomMissing.Wrap(msg.Denom)
	}
	if meta.Admin != sender.String() {
		return nil, types.ErrUnauthorized.Wrap("only denom admin can change admin")
	}
	newAdmin, err := sdk.AccAddressFromBech32(msg.NewAdmin)
	if err != nil {
		return nil, err
	}
	meta.Admin = newAdmin.String()
	if err := m.SetDenom(ctx, meta); err != nil {
		return nil, err
	}
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
