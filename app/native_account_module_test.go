package app

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	nativeaccountkeeper "github.com/sovereign-l1/l1/x/native-account/keeper"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

func TestNativeAccountRegisteredAsFullAppModule(t *testing.T) {
	app, genesis := setup(true, 5)
	_ = genesis

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
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUpdateAuthPolicy{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgRotateKey{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgRecoverAccount{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgFreezeAccount{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgPayStorageDebt{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUnfreezeAccount{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUpdateAccountMetadata{}))

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
		Version:	nativeaccounttypes.DefaultGenesis().Version,
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

func TestNativeAccountDuplicateActivationRejectedByRoute(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(31)
	pubKey := nativeAccountModuleTestPubKey()
	pair, err := nativeaccounttypes.ActivationAddressPair(pubKey)
	require.NoError(t, err)
	msg := nativeAccountActivationMsg(pubKey, pair.User, pair.Raw)
	route := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgActivateAccount{})

	_, err = route(ctx, &msg)
	require.NoError(t, err)
	_, err = route(ctx, &msg)
	require.ErrorContains(t, err, "already active")

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Len(t, exported.Accounts, 1)
}

func TestNativeAccountInactiveCannotSendNormalMessages(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(32)
	pubKey := nativeAccountModuleTestPubKey()
	pair, err := nativeaccounttypes.ActivationAddressPair(pubKey)
	require.NoError(t, err)
	route := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUpdateAccountMetadata{})

	_, err = route(ctx, &nativeaccounttypes.MsgUpdateAccountMetadata{
		AccountUser:	pair.User,
		Metadata:	nativeaccounttypes.AccountMetadata{MetadataHash: "hash"},
		Signers:	[]string{nativeaccounttypes.PublicKeyText(pubKey)},
		CurrentHeight:	32,
	})
	require.ErrorContains(t, err, "inactive account cannot send normal messages")

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Empty(t, exported.Accounts)
}

func TestNativeAccountFrozenWhitelistAllowsRecoveryDebtAndUnfreezeOnly(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(33)
	pubKey := nativeAccountModuleTestPubKey()
	account := nativeAccountActivateViaRoute(t, app, ctx, pubKey)
	signer := account.PubKeys[0]

	freezeRoute := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgFreezeAccount{})
	_, err := freezeRoute(ctx, &nativeaccounttypes.MsgFreezeAccount{
		AccountUser:	account.AddressUser,
		Signers:	[]string{signer},
		CurrentHeight:	33,
	})
	require.NoError(t, err)

	account, found, err := app.NativeAccountKeeper.AccountByUser(ctx, account.AddressUser)
	require.NoError(t, err)
	require.True(t, found)
	account.StorageRentDebt = 5
	account.AuthPolicy = account.AuthPolicy.Normalize()
	account.AuthPolicy.RecoveryPolicy = nativeaccounttypes.RecoveryPolicy{Keys: []string{"recovery-key"}, Threshold: 1}
	require.NoError(t, app.NativeAccountKeeper.SetAccount(ctx, account))

	metadataRoute := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUpdateAccountMetadata{})
	_, err = metadataRoute(ctx, &nativeaccounttypes.MsgUpdateAccountMetadata{
		AccountUser:	account.AddressUser,
		Metadata:	nativeaccounttypes.AccountMetadata{MetadataHash: "normal-update"},
		Signers:	[]string{signer},
		CurrentHeight:	34,
	})
	require.ErrorContains(t, err, "frozen account allows only")

	recoverRoute := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgRecoverAccount{})
	_, err = recoverRoute(ctx, &nativeaccounttypes.MsgRecoverAccount{
		AccountUser:	account.AddressUser,
		Signers:	[]string{"recovery-key"},
		CurrentHeight:	35,
	})
	require.NoError(t, err)

	account, found, err = app.NativeAccountKeeper.AccountByUser(ctx, account.AddressUser)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, nativeaccounttypes.AccountStatusRecovered, account.Status)
	account.Status = nativeaccounttypes.AccountStatusFrozen
	account.StorageRentDebt = 5
	require.NoError(t, app.NativeAccountKeeper.SetAccount(ctx, account))

	payDebtRoute := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgPayStorageDebt{})
	_, err = payDebtRoute(ctx, &nativeaccounttypes.MsgPayStorageDebt{
		AccountUser:	account.AddressUser,
		Amount:		5,
		Signers:	[]string{signer},
		CurrentHeight:	36,
	})
	require.NoError(t, err)

	unfreezeRoute := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUnfreezeAccount{})
	_, err = unfreezeRoute(ctx, &nativeaccounttypes.MsgUnfreezeAccount{
		AccountUser:		account.AddressUser,
		Signers:		[]string{signer},
		CurrentHeight:		37,
		StorageDebtPaid:	true,
	})
	require.NoError(t, err)

	account, found, err = app.NativeAccountKeeper.AccountByUser(ctx, account.AddressUser)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, nativeaccounttypes.AccountStatusActive, account.Status)
	require.Zero(t, account.StorageRentDebt)
}

func TestNativeAccountAuthPolicyRejectsSecretsAndExportIsClean(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(37)
	pubKey := nativeAccountModuleTestPubKey()
	account := nativeAccountActivateViaRoute(t, app, ctx, pubKey)
	route := app.MsgServiceRouter().Handler(&nativeaccounttypes.MsgUpdateAuthPolicy{})

	_, err := route(ctx, &nativeaccounttypes.MsgUpdateAuthPolicy{
		AccountUser:	account.AddressUser,
		NewAuthPolicy: nativeaccounttypes.AuthPolicy{
			Version:	1,
			Mode:		nativeaccounttypes.AuthModeSingleKey,
			Keys: []nativeaccounttypes.AuthKey{
				{ID: "primary", PublicKey: "seed_phrase: never store this"},
			},
		},
		Signers:	[]string{account.PubKeys[0]},
		CurrentHeight:	37,
	})
	require.ErrorContains(t, err, "must not contain private keys or seed phrases")

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	bz, err := json.Marshal(exported)
	require.NoError(t, err)
	require.NotContains(t, string(bz), "seed_phrase")
	require.NotContains(t, string(bz), "private key")
}

func TestNativeAccountVirtualQueryDoesNotExportInactiveAccount(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(38)
	pubKey := nativeAccountModuleTestPubKey()
	pair, err := nativeaccounttypes.ActivationAddressPair(pubKey)
	require.NoError(t, err)

	query := nativeaccountkeeper.NewQueryServerImpl(app.NativeAccountKeeper)
	resp, err := query.VirtualAccount(ctx, &nativeaccounttypes.QueryVirtualAccountRequest{AddressUser: pair.User})
	require.NoError(t, err)
	require.Equal(t, nativeaccounttypes.AccountStatusInactive, resp.Status)
	require.False(t, resp.Persistent)

	exported, err := app.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Empty(t, exported.Accounts)
}

func TestNativeAccountExportImportRoundTripStable(t *testing.T) {
	source := Setup(t, false)
	ctx := source.NewContext(false).WithBlockHeight(39)
	pubKey := nativeAccountModuleTestPubKey()
	account := nativeAccountActivateViaRoute(t, source, ctx, pubKey)
	account.Status = nativeaccounttypes.AccountStatusFrozen
	account.StorageRentDebt = 9
	require.NoError(t, source.NativeAccountKeeper.SetAccount(ctx, account))

	exported, err := source.NativeAccountKeeper.ExportGenesis(ctx)
	require.NoError(t, err)

	target := Setup(t, false)
	targetCtx := target.NewContext(false).WithBlockHeight(40)
	require.NoError(t, target.NativeAccountKeeper.InitGenesis(targetCtx, exported))
	roundTrip, err := target.NativeAccountKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}

func nativeAccountActivationMsg(pubKey *secp256k1.PubKey, userAddress, rawAddress string) nativeaccounttypes.MsgActivateAccount {
	return nativeaccounttypes.MsgActivateAccount{
		AddressUser:	userAddress,
		AddressRaw:	rawAddress,
		PublicKeyType:	"secp256k1",
		PublicKeyHex:	hex.EncodeToString(pubKey.Bytes()),
		FeePaid:	0,
	}
}

func nativeAccountActivateViaRoute(t *testing.T, app *L1App, ctx sdk.Context, pubKey *secp256k1.PubKey) nativeaccounttypes.Account {
	t.Helper()
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

	addr, err := addressing.ParseAccAddress(account.AddressUser)
	require.NoError(t, err)
	FundTestAddr(t, app, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)))
	return account
}

func nativeAccountGenesisFixture(t *testing.T) nativeaccounttypes.Account {
	t.Helper()
	pubKey := nativeAccountModuleTestPubKey()
	pair, err := nativeaccounttypes.ActivationAddressPair(pubKey)
	require.NoError(t, err)
	features, err := nativeaccounttypes.DefaultFeatureFlags(nativeaccounttypes.CurrentAccountVersion)
	require.NoError(t, err)
	return nativeaccounttypes.Account{
		Version:	nativeaccounttypes.CurrentAccountVersion,
		AddressUser:	pair.User,
		AddressRaw:	pair.Raw,
		PubKeys:	[]string{nativeaccounttypes.PublicKeyText(pubKey)},
		AccountNumber:	1,
		Sequence:	nativeaccounttypes.ActivationInitialSequence,
		Status:		nativeaccounttypes.AccountStatusActive,
		AuthPolicy: nativeaccounttypes.AuthPolicy{
			Version:	1,
			Mode:		nativeaccounttypes.AuthModeSingleKey,
		},
		FeatureFlags:			features,
		CreatedHeight:			1,
		LastActiveHeight:		1,
		LastStorageChargeHeight:	1,
	}
}

func nativeAccountModuleTestPubKey() *secp256k1.PubKey {
	return &secp256k1.PubKey{Key: []byte{
		0x02, 0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b,
		0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98,
	}}
}
