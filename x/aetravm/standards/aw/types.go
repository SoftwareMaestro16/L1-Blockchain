package aw

import (
	"crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	StandardName        = "AW-5"
	MaxExtensions       = 32
	MaxMultiSendCount   = 16
	MaxPayloadBytes     = 4096
	MaxRecoveryDelaySec = int64(30 * 24 * 60 * 60)
	PublicKeyLength     = ed25519.PublicKeySize
	SignatureLength     = ed25519.SignatureSize
	signingBytesDomain  = "aetra/AW-5/external-command/v1"
	DefaultWalletSeqno  = uint64(0)
	DefaultWalletIDBase = uint64(1)
)

type CommandKind string

const (
	CommandSend                   CommandKind = "send"
	CommandUpdateSignatureAllowed CommandKind = "update_signature_allowed"
	CommandInstallExtension       CommandKind = "install_extension"
	CommandRemoveExtension        CommandKind = "remove_extension"
)

type ExtensionState struct {
	Address   sdk.AccAddress
	Installed bool
}

type RecoveryPolicy struct {
	Enabled      bool
	Authority    sdk.AccAddress
	DelaySeconds int64
}

type WalletState struct {
	Address               sdk.AccAddress
	SignatureAllowed      bool
	Seqno                 uint64
	WalletID              uint64
	PublicKey             ed25519.PublicKey
	Owner                 sdk.AccAddress
	Extensions            map[string]ExtensionState
	RecoveryAuthority     sdk.AccAddress
	RecoveryPolicy        RecoveryPolicy
	SubscriptionsEnabled  bool
	StandingPaymentPolicy string
}

type OutboundMessage struct {
	To      sdk.AccAddress
	Amount  sdk.Coins
	Payload []byte
}

type ExternalCommand struct {
	WalletAddress    sdk.AccAddress
	WalletID         uint64
	Seqno            uint64
	ValidUntil       int64
	Kind             CommandKind
	Messages         []OutboundMessage
	SignatureAllowed bool
	Extension        sdk.AccAddress
	Signature        []byte
}

type InternalExtensionCommand struct {
	Extension sdk.AccAddress
	Messages  []OutboundMessage
}

type RelayedCommand struct {
	Relayer sdk.AccAddress
	Fees    sdk.Coins
	Command ExternalCommand
}

type State struct {
	Wallet       WalletState
	SentMessages []OutboundMessage
}
