package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePool{}, "l1/dex/MsgCreatePool", nil)
	cdc.RegisterConcrete(&MsgAddLiquidity{}, "l1/dex/MsgAddLiquidity", nil)
	cdc.RegisterConcrete(&MsgRemoveLiquidity{}, "l1/dex/MsgRemoveLiquidity", nil)
	cdc.RegisterConcrete(&MsgSwapExactAmountIn{}, "l1/dex/MsgSwapExactAmountIn", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "l1/dex/MsgUpdateParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreatePool{},
		&MsgAddLiquidity{},
		&MsgRemoveLiquidity{},
		&MsgSwapExactAmountIn{},
		&MsgUpdateParams{},
	)
}
