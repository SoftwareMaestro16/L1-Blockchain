package txhandlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signing "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

type mockNativeAccountReader struct {
	accounts map[string]string
}

func (m mockNativeAccountReader) AccountStatus(_ context.Context, userAddress string) (string, bool, error) {
	status, found := m.accounts[userAddress]
	if !found {
		return nativeaccounttypes.AccountStatusInactive, false, nil
	}
	return status, found, nil
}

type sigTx struct {
	msgs	[]sdk.Msg
	signers	[][]byte
}

func (tx sigTx) GetMsgs() []sdk.Msg					{ return tx.msgs }
func (tx sigTx) GetMsgsV2() ([]protov2.Message, error)			{ return nil, nil }
func (tx sigTx) GetSigners() ([][]byte, error)				{ return tx.signers, nil }
func (tx sigTx) GetPubKeys() ([]cryptotypes.PubKey, error)		{ return nil, nil }
func (tx sigTx) GetSignaturesV2() ([]signing.SignatureV2, error)	{ return nil, nil }

func newSigTx(addr sdk.AccAddress, msg sdk.Msg) sigTx {
	return sigTx{msgs: []sdk.Msg{msg}, signers: [][]byte{addr}}
}

func TestStorageRentDecoratorAllowsActiveAccount(t *testing.T) {
	addr := testAccAddress(t)
	aeAddr := aetraaddress.FormatAccAddress(addr)
	reader := mockNativeAccountReader{accounts: map[string]string{aeAddr: nativeaccounttypes.AccountStatusActive}}

	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := StorageRentDecorator(reader, next)(sdk.Context{}, newSigTx(addr, bankSendMsg(t, addr)), false)
	require.NoError(t, err)
	require.True(t, called)
}

func TestStorageRentDecoratorRejectsFrozenAccount(t *testing.T) {
	addr := testAccAddress(t)
	aeAddr := aetraaddress.FormatAccAddress(addr)
	reader := mockNativeAccountReader{accounts: map[string]string{aeAddr: nativeaccounttypes.AccountStatusFrozen}}

	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := StorageRentDecorator(reader, next)(sdk.Context{}, newSigTx(addr, bankSendMsg(t, addr)), false)
	require.ErrorContains(t, err, "frozen native account")
	require.ErrorContains(t, err, "only storage debt payment, unfreeze, and recovery are allowed")
	require.False(t, called)
}

func TestStorageRentDecoratorSkipsNoNativeAccount(t *testing.T) {
	addr := testAccAddress(t)
	reader := mockNativeAccountReader{accounts: map[string]string{}}

	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := StorageRentDecorator(reader, next)(sdk.Context{}, newSigTx(addr, bankSendMsg(t, addr)), false)
	require.NoError(t, err)
	require.True(t, called)
}

func TestStorageRentDecoratorSkipsRecoveredAccount(t *testing.T) {
	addr := testAccAddress(t)
	aeAddr := aetraaddress.FormatAccAddress(addr)
	reader := mockNativeAccountReader{accounts: map[string]string{aeAddr: nativeaccounttypes.AccountStatusRecovered}}

	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := StorageRentDecorator(reader, next)(sdk.Context{}, newSigTx(addr, bankSendMsg(t, addr)), false)
	require.NoError(t, err)
	require.True(t, called)
}

func TestStorageRentDecoratorSkipsSimulation(t *testing.T) {
	addr := testAccAddress(t)
	aeAddr := aetraaddress.FormatAccAddress(addr)
	reader := mockNativeAccountReader{accounts: map[string]string{aeAddr: nativeaccounttypes.AccountStatusFrozen}}

	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := StorageRentDecorator(reader, next)(sdk.Context{}, newSigTx(addr, bankSendMsg(t, addr)), true)
	require.NoError(t, err)
	require.True(t, called)
}

func TestStorageRentDecoratorAllowsNonNativeMsg(t *testing.T) {
	addr := testAccAddress(t)
	reader := mockNativeAccountReader{accounts: map[string]string{}}

	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	_, err := StorageRentDecorator(reader, next)(sdk.Context{}, newSigTx(addr, bankSendMsg(t, addr)), false)
	require.NoError(t, err)
	require.True(t, called)
}

func testAccAddress(t *testing.T) sdk.AccAddress {
	t.Helper()
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = byte(i + 1)
	}
	return sdk.AccAddress(bz)
}

func bankSendMsg(t *testing.T, from sdk.AccAddress) *banktypes.MsgSend {
	t.Helper()
	aeAddr := aetraaddress.FormatAccAddress(from)
	return &banktypes.MsgSend{
		FromAddress:	aeAddr,
		ToAddress:	aetraaddress.FormatAccAddress(make([]byte, 20)),
		Amount:		sdk.NewCoins(sdk.NewInt64Coin("naet", 100)),
	}
}
