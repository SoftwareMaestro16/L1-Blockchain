package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgActivateAccount{}, "l1/nativeaccount/MsgActivateAccount", nil)
	cdc.RegisterConcrete(&MsgUpdateAuthPolicy{}, "l1/nativeaccount/MsgUpdateAuthPolicy", nil)
	cdc.RegisterConcrete(&MsgRotateKey{}, "l1/nativeaccount/MsgRotateKey", nil)
	cdc.RegisterConcrete(&MsgRecoverAccount{}, "l1/nativeaccount/MsgRecoverAccount", nil)
	cdc.RegisterConcrete(&MsgFreezeAccount{}, "l1/nativeaccount/MsgFreezeAccount", nil)
	cdc.RegisterConcrete(&MsgPayStorageDebt{}, "l1/nativeaccount/MsgPayStorageDebt", nil)
	cdc.RegisterConcrete(&MsgUnfreezeAccount{}, "l1/nativeaccount/MsgUnfreezeAccount", nil)
	cdc.RegisterConcrete(&MsgUpdateAccountMetadata{}, "l1/nativeaccount/MsgUpdateAccountMetadata", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgActivateAccount{},
		&MsgUpdateAuthPolicy{},
		&MsgRotateKey{},
		&MsgRecoverAccount{},
		&MsgFreezeAccount{},
		&MsgPayStorageDebt{},
		&MsgUnfreezeAccount{},
		&MsgUpdateAccountMetadata{},
	)
	registry.RegisterImplementations(
		(*txtypes.MsgResponse)(nil),
		&MsgActivateAccountResponse{},
		&MsgUpdateAuthPolicyResponse{},
		&MsgRotateKeyResponse{},
		&MsgRecoverAccountResponse{},
		&MsgFreezeAccountResponse{},
		&MsgPayStorageDebtResponse{},
		&MsgUnfreezeAccountResponse{},
		&MsgUpdateAccountMetadataResponse{},
	)
}
