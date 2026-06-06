package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetBaseCommission{}, "l1/dynamiccommission/MsgSetBaseCommission", nil)
	cdc.RegisterConcrete(&MsgUpdateCommissionParams{}, "l1/dynamiccommission/MsgUpdateCommissionParams", nil)
	cdc.RegisterConcrete(&MsgRecomputeEffectiveCommission{}, "l1/dynamiccommission/MsgRecomputeEffectiveCommission", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSetBaseCommission{},
		&MsgUpdateCommissionParams{},
		&MsgRecomputeEffectiveCommission{},
	)
}
