package keeperconfig

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/tx/signing"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func NewTxConfig(appCodec codec.Codec, bankKeeper bankkeeper.BaseKeeper) client.TxConfig {
	enabledSignModes := append(authtx.DefaultSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)
	txConfig, err := authtx.NewTxConfigWithOptions(
		appCodec,
		authtx.ConfigOptions{
			EnabledSignModes:	enabledSignModes,
			SigningOptions: &signing.Options{
				AddressCodec:		aetraaddress.Codec{},
				ValidatorAddressCodec:	aetraaddress.Codec{},
			},
			TextualCoinMetadataQueryFn:	txmodule.NewBankKeeperCoinMetadataQueryFn(bankKeeper),
		},
	)
	if err != nil {
		panic(err)
	}
	return txConfig
}
