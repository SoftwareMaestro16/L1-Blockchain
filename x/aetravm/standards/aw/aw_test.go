package aw

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const testNow = int64(1_800_000_000)

func TestDeployWalletQueryAndUpdateSignatureAllowed(t *testing.T) {
	state, _, privateKey := newTestState(t)
	require.Equal(t, DefaultWalletSeqno, state.Wallet.Seqno)
	require.True(t, state.Wallet.SignatureAllowed)
	require.Equal(t, uint64(42), state.QueryWalletState().WalletID)

	cmd := signCommand(t, privateKey, ExternalCommand{
		WalletAddress:    state.Wallet.Address,
		WalletID:         state.Wallet.WalletID,
		Seqno:            state.Wallet.Seqno,
		ValidUntil:       testNow + 60,
		Kind:             CommandUpdateSignatureAllowed,
		SignatureAllowed: false,
	})
	require.NoError(t, state.ApplyExternalCommand(cmd, testNow, naetFee()))
	require.False(t, state.Wallet.SignatureAllowed)
	require.Equal(t, uint64(1), state.Wallet.Seqno)

	next := signedSend(t, state, privateKey, testAddr(3), 1)
	require.ErrorContains(t, state.ApplyExternalCommand(next, testNow, naetFee()), "external signatures are disabled")
	require.Len(t, state.SentMessages, 0)
}

func TestSignedSendAndReplayedSeqnoRejected(t *testing.T) {
	state, _, privateKey := newTestState(t)
	cmd := signedSend(t, state, privateKey, testAddr(3), 7)

	require.NoError(t, state.ApplyExternalCommand(cmd, testNow, naetFee()))
	require.Equal(t, uint64(1), state.Wallet.Seqno)
	require.Len(t, state.SentMessages, 1)

	require.ErrorContains(t, state.ApplyExternalCommand(cmd, testNow, naetFee()), "seqno")
	require.Equal(t, uint64(1), state.Wallet.Seqno)
	require.Len(t, state.SentMessages, 1)
}

func TestWrongWalletIDExpiredAndInvalidSignatureRejectedBeforeMutation(t *testing.T) {
	state, _, privateKey := newTestState(t)

	wrongWallet := signedSend(t, state, privateKey, testAddr(3), 1)
	wrongWallet.WalletID++
	wrongWallet = signCommand(t, privateKey, wrongWallet)
	require.ErrorContains(t, state.ApplyExternalCommand(wrongWallet, testNow, naetFee()), "wrong wallet_id")
	require.Equal(t, uint64(0), state.Wallet.Seqno)

	expired := signedSend(t, state, privateKey, testAddr(3), 1)
	expired.ValidUntil = testNow - 1
	expired = signCommand(t, privateKey, expired)
	require.ErrorContains(t, state.ApplyExternalCommand(expired, testNow, naetFee()), "expired")
	require.Equal(t, uint64(0), state.Wallet.Seqno)

	invalid := signedSend(t, state, privateKey, testAddr(3), 1)
	invalid.Signature[0] ^= 0xff
	require.ErrorContains(t, state.ApplyExternalCommand(invalid, testNow, naetFee()), "invalid signature")
	require.Equal(t, uint64(0), state.Wallet.Seqno)
	require.Len(t, state.SentMessages, 0)
}

func TestExtensionInstallRemoveAndAuthorizedSend(t *testing.T) {
	state, _, privateKey := newTestState(t)
	extension := testAddr(9)

	install := signCommand(t, privateKey, ExternalCommand{
		WalletAddress: state.Wallet.Address,
		WalletID:      state.Wallet.WalletID,
		Seqno:         state.Wallet.Seqno,
		ValidUntil:    testNow + 60,
		Kind:          CommandInstallExtension,
		Extension:     extension,
	})
	require.NoError(t, state.ApplyExternalCommand(install, testNow, naetFee()))
	require.Contains(t, state.Wallet.Extensions, string(extension))

	extensionSend := InternalExtensionCommand{
		Extension: extension,
		Messages: []OutboundMessage{{
			To:     testAddr(3),
			Amount: sdk.NewCoins(sdk.NewInt64Coin("naet", 1)),
		}},
	}
	require.NoError(t, state.ApplyExtensionCommand(extensionSend))
	require.Len(t, state.SentMessages, 1)

	remove := signCommand(t, privateKey, ExternalCommand{
		WalletAddress: state.Wallet.Address,
		WalletID:      state.Wallet.WalletID,
		Seqno:         state.Wallet.Seqno,
		ValidUntil:    testNow + 60,
		Kind:          CommandRemoveExtension,
		Extension:     extension,
	})
	require.NoError(t, state.ApplyExternalCommand(remove, testNow, naetFee()))
	require.NotContains(t, state.Wallet.Extensions, string(extension))
	require.ErrorContains(t, state.ApplyExtensionCommand(extensionSend), "unauthorized wallet extension")
}

func TestUnauthorizedExtensionRejected(t *testing.T) {
	state, _, _ := newTestState(t)
	err := state.ApplyExtensionCommand(InternalExtensionCommand{
		Extension: testAddr(9),
		Messages: []OutboundMessage{{
			To:     testAddr(3),
			Amount: sdk.NewCoins(sdk.NewInt64Coin("naet", 1)),
		}},
	})
	require.ErrorContains(t, err, "unauthorized wallet extension")
	require.Len(t, state.SentMessages, 0)
}

func TestMultiSendBounded(t *testing.T) {
	state, _, privateKey := newTestState(t)
	messages := make([]OutboundMessage, MaxMultiSendCount)
	for i := range messages {
		messages[i] = OutboundMessage{
			To:     testAddr(byte(i + 10)),
			Amount: sdk.NewCoins(sdk.NewInt64Coin("naet", 1)),
		}
	}
	cmd := signCommand(t, privateKey, ExternalCommand{
		WalletAddress: state.Wallet.Address,
		WalletID:      state.Wallet.WalletID,
		Seqno:         state.Wallet.Seqno,
		ValidUntil:    testNow + 60,
		Kind:          CommandSend,
		Messages:      messages,
	})
	require.NoError(t, state.ApplyExternalCommand(cmd, testNow, naetFee()))
	require.Len(t, state.SentMessages, MaxMultiSendCount)

	state, _, privateKey = newTestState(t)
	tooMany := append(messages, OutboundMessage{To: testAddr(99)})
	bad := ExternalCommand{
		WalletAddress: state.Wallet.Address,
		WalletID:      state.Wallet.WalletID,
		Seqno:         state.Wallet.Seqno,
		ValidUntil:    testNow + 60,
		Kind:          CommandSend,
		Messages:      tooMany,
		Signature:     bytes.Repeat([]byte{1}, SignatureLength),
	}
	require.ErrorContains(t, state.ApplyExternalCommand(bad, testNow, naetFee()), "multi-send message batch")
	require.Len(t, state.SentMessages, 0)
}

func TestRelayerFlowPaysNaet(t *testing.T) {
	state, _, privateKey := newTestState(t)
	cmd := signedSend(t, state, privateKey, testAddr(3), 1)
	require.NoError(t, state.ApplyRelayedCommand(RelayedCommand{
		Relayer: testAddr(4),
		Fees:    naetFee(),
		Command: cmd,
	}, testNow))
	require.Equal(t, uint64(1), state.Wallet.Seqno)
	require.Len(t, state.SentMessages, 1)

	state, _, privateKey = newTestState(t)
	cmd = signedSend(t, state, privateKey, testAddr(3), 1)
	err := state.ApplyRelayedCommand(RelayedCommand{
		Relayer: testAddr(4),
		Fees:    sdk.NewCoins(sdk.NewInt64Coin("testtoken", 1)),
		Command: cmd,
	}, testNow)
	require.Error(t, err)
	require.Equal(t, uint64(0), state.Wallet.Seqno)
	require.Len(t, state.SentMessages, 0)
}

func TestRecoveryPolicyValidatedAndCloned(t *testing.T) {
	state, _, _ := newTestState(t)
	state.Wallet.RecoveryPolicy = RecoveryPolicy{
		Enabled:      true,
		Authority:    testAddr(11),
		DelaySeconds: 60,
	}
	require.NoError(t, state.Wallet.Validate())
	queried := state.QueryWalletState()
	queried.RecoveryPolicy.Authority[0] = 0xff
	require.NotEqual(t, queried.RecoveryPolicy.Authority, state.Wallet.RecoveryPolicy.Authority)

	state.Wallet.RecoveryPolicy.DelaySeconds = MaxRecoveryDelaySec + 1
	require.ErrorContains(t, state.Wallet.Validate(), "recovery delay")

	state.Wallet.RecoveryPolicy = RecoveryPolicy{Enabled: false, Authority: testAddr(12)}
	require.ErrorContains(t, state.Wallet.Validate(), "disabled recovery policy")
}

func newTestState(t testing.TB) (*State, ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	state, err := NewState(WalletState{
		Address:          testAddr(1),
		SignatureAllowed: true,
		Seqno:            DefaultWalletSeqno,
		WalletID:         42,
		PublicKey:        publicKey,
		Owner:            testAddr(2),
		Extensions:       make(map[string]ExtensionState),
	})
	require.NoError(t, err)
	return state, publicKey, privateKey
}

func signedSend(t testing.TB, state *State, privateKey ed25519.PrivateKey, recipient sdk.AccAddress, amount int64) ExternalCommand {
	t.Helper()
	return signCommand(t, privateKey, ExternalCommand{
		WalletAddress: state.Wallet.Address,
		WalletID:      state.Wallet.WalletID,
		Seqno:         state.Wallet.Seqno,
		ValidUntil:    testNow + 60,
		Kind:          CommandSend,
		Messages: []OutboundMessage{{
			To:     recipient,
			Amount: sdk.NewCoins(sdk.NewInt64Coin("naet", amount)),
		}},
	})
}

func signCommand(t testing.TB, privateKey ed25519.PrivateKey, cmd ExternalCommand) ExternalCommand {
	t.Helper()
	cmd.Signature = nil
	signingBytes, err := cmd.SigningBytes()
	require.NoError(t, err)
	cmd.Signature = ed25519.Sign(privateKey, signingBytes)
	return cmd
}

func naetFee() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("naet", 1))
}

func testAddr(fill byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{fill}, 20))
}
