package app

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

func TestNativeAccountRegisteredAsFullAppModule(t *testing.T) {
	app, genesis := setup(true, 5)

	require.Contains(t, app.ModuleManager.Modules, nativeaccounttypes.ModuleName)
	require.Contains(t, app.keys, nativeaccounttypes.StoreKey)
	require.Contains(t, genesis, nativeaccounttypes.ModuleName)
	require.Contains(t, aetherCoreInitGenesisOrder(), nativeaccounttypes.ModuleName)
	require.Contains(t, aetherCoreExportGenesisOrder(), nativeaccounttypes.ModuleName)
	require.NotContains(t, aetherCoreBeginBlockerOrder(), nativeaccounttypes.ModuleName)
	require.NotContains(t, aetherCoreEndBlockerOrder(), nativeaccounttypes.ModuleName)
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

	route := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgActivateAccount{})
	require.NotNil(t, route, "MsgActivateAccount must be registered as an on-chain Msg service route")

	var gs nativeaccounttypes.GenesisState
	require.NoError(t, json.Unmarshal(genesis[nativeaccounttypes.ModuleName], &gs))
	require.NoError(t, gs.Validate())
}

func TestNativeAccountMsgRouteActivatesAndPersistsAccount(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(21)
	pubKey := nativeAccountModuleTestPubKey()
	pair, err := nativeaccounttypes.ActivationAddressPair(pubKey)
	require.NoError(t, err)
	msg := nativeAccountActivationMsg(pubKey, pair.User, pair.Raw)

	route := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgActivateAccount{})
	require.NotNil(t, route)
	_, err = route(ctx, &msg)
	require.NoError(t, err)

	account, found, err := app.NativeAccountKeeper.AccountByUser(ctx, pair.User)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, pair.User, account.AddressUser)
	require.Equal(t, pair.Raw, account.AddressRaw)
	require.Equal(t, uint64(1), account.AccountNumber)
	require.Equal(t, nativeaccounttypes.ActivationInitialSequence, account.Sequence)
	require.Equal(t, nativeaccounttypes.CurrentAccountVersion, account.Version)

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Len(t, exported.Accounts, 1)
	exportedJSON, err := json.Marshal(exported)
	require.NoError(t, err)
	require.NotContains(t, string(exportedJSON), "private key")
	require.NotContains(t, string(exportedJSON), "seed phrase")
}

func TestNativeAccountMsgRouteRejectsMismatchWithoutPartialState(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(22)
	pubKey := nativeAccountModuleTestPubKey()
	otherPair, err := nativeaccounttypes.ActivationAddressPair(&secp256k1.PubKey{Key: []byte{
		0x03, 0x7d, 0x5c, 0xf0, 0x48, 0x9f, 0x77, 0x6a, 0x10, 0x48, 0x0b, 0x42, 0xeb, 0xe4, 0x4d, 0xdb,
		0x4f, 0x12, 0x8a, 0x9e, 0x1d, 0x77, 0x5f, 0x86, 0x99, 0x79, 0x2e, 0x4b, 0xac, 0x20, 0x0f, 0x4e, 0x25,
	}})
	require.NoError(t, err)
	msg := nativeAccountActivationMsg(pubKey, otherPair.User, otherPair.Raw)

	route := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgActivateAccount{})
	require.NotNil(t, route)
	_, err = route(ctx, &msg)
	require.ErrorContains(t, err, "must equal derived public key address")

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Empty(t, exported.Accounts)
}

func TestNativeAccountGenesisImportRejectsMalformedBeforePartialWrite(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(30)
	account := nativeAccountGenesisFixture(t)
	duplicate := account

	err := app.NativeAccountKeeper.InitGenesis(ctx, nativeaccounttypes.GenesisState{
		Version: nativeaccounttypes.DefaultGenesis().Version,
		Accounts: []nativeaccounttypes.Account{
			account,
			duplicate,
		},
	})
	require.ErrorContains(t, err, "duplicate native account user address")

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Empty(t, exported.Accounts)
}

func nativeAccountActivationMsg(pubKey *secp256k1.PubKey, userAddress, rawAddress string) nativeaccounttypes.MsgActivateAccount {
	return nativeaccounttypes.MsgActivateAccount{
		AddressUser:   userAddress,
		AddressRaw:    rawAddress,
		PublicKeyType: "secp256k1",
		PublicKeyHex:  hex.EncodeToString(pubKey.Bytes()),
		FeePaid:       0,
	}
}

func nativeAccountGenesisFixture(t *testing.T) nativeaccounttypes.Account {
	t.Helper()
	pubKey := nativeAccountModuleTestPubKey()
	pair, err := nativeaccounttypes.ActivationAddressPair(pubKey)
	require.NoError(t, err)
	features, err := nativeaccounttypes.DefaultFeatureFlags(nativeaccounttypes.CurrentAccountVersion)
	require.NoError(t, err)
	return nativeaccounttypes.Account{
		Version:       nativeaccounttypes.CurrentAccountVersion,
		AddressUser:   pair.User,
		AddressRaw:    pair.Raw,
		PubKeys:       []string{nativeaccounttypes.PublicKeyText(pubKey)},
		AccountNumber: 1,
		Sequence:      nativeaccounttypes.ActivationInitialSequence,
		Status:        nativeaccounttypes.AccountStatusActive,
		AuthPolicy: nativeaccounttypes.AuthPolicy{
			Version: 1,
			Mode:    nativeaccounttypes.AuthModeSingleKey,
		},
		FeatureFlags:            features,
		CreatedHeight:           1,
		LastActiveHeight:        1,
		LastStorageChargeHeight: 1,
	}
}

func nativeAccountModuleTestPubKey() *secp256k1.PubKey {
	return &secp256k1.PubKey{Key: []byte{
		0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b,
		0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98,
	}}
}
