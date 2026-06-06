package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateConcentrationParams{}, "l1/stakeconcentration/MsgUpdateConcentrationParams", nil)
	cdc.RegisterConcrete(&MsgRecomputeConcentration{}, "l1/stakeconcentration/MsgRecomputeConcentration", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateConcentrationParams{},
		&MsgRecomputeConcentration{},
	)
}
