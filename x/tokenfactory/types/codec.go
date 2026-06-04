package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateDenom{}, "l1/tokenfactory/MsgCreateDenom", nil)
	cdc.RegisterConcrete(&MsgMint{}, "l1/tokenfactory/MsgMint", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "l1/tokenfactory/MsgBurn", nil)
	cdc.RegisterConcrete(&MsgChangeAdmin{}, "l1/tokenfactory/MsgChangeAdmin", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "l1/tokenfactory/MsgUpdateParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateDenom{},
		&MsgMint{},
		&MsgBurn{},
		&MsgChangeAdmin{},
		&MsgUpdateParams{},
	)
}
