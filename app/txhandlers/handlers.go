package txhandlers

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
)

func NewAnteHandler(
	txConfig client.TxConfig,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.BaseKeeper,
	feeGrantKeeper feegrantkeeper.Keeper,
) sdk.AnteHandler {
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:		accountKeeper,
			BankKeeper:		bankKeeper,
			SignModeHandler:	txConfig.SignModeHandler(),
			FeegrantKeeper:		feeGrantKeeper,
			SigGasConsumer:		ante.DefaultSigVerificationGasConsumer,
			SigVerifyOptions: []ante.SigVerificationDecoratorOption{
				ante.WithUnorderedTxGasCost(ante.DefaultUnorderedTxGasCost),
				ante.WithMaxUnorderedTxTimeoutDuration(ante.DefaultMaxTimeoutDuration),
			},
		},
	)
	if err != nil {
		panic(err)
	}

	return anteHandler
}

func NewPostHandler() sdk.PostHandler {
	postHandler, err := posthandler.NewPostHandler(posthandler.HandlerOptions{})
	if err != nil {
		panic(err)
	}
	return postHandler
}
