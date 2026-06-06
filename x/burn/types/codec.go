package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgBurnProtocolCoins{}, "l1/burn/MsgBurnProtocolCoins", nil)
	cdc.RegisterConcrete(&MsgBurnUserCoins{}, "l1/burn/MsgBurnUserCoins", nil)
	cdc.RegisterConcrete(&MsgUpdateBurnParams{}, "l1/burn/MsgUpdateBurnParams", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgBurnProtocolCoins{},
		&MsgBurnUserCoins{},
		&MsgUpdateBurnParams{},
	)
}
