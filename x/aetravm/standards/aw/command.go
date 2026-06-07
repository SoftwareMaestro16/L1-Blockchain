package aw

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

func ValidateProtocolFees(fees sdk.Coins) error {
	return feestypes.ValidateFeeCoins(feestypes.DefaultParams(), fees, true)
}

func (s *State) VerifyExternalCommand(cmd ExternalCommand, now int64) error {
	if err := s.Wallet.Validate(); err != nil {
		return err
	}
	if !s.Wallet.SignatureAllowed {
		return errors.New("external signatures are disabled")
	}
	if !cmd.WalletAddress.Equals(s.Wallet.Address) {
		return errors.New("wrong wallet address")
	}
	if cmd.WalletID != s.Wallet.WalletID {
		return errors.New("wrong wallet_id")
	}
	if cmd.Seqno != s.Wallet.Seqno {
		return errors.New("replayed or out-of-order seqno")
	}
	if now > cmd.ValidUntil {
		return errors.New("signed command expired")
	}
	if len(cmd.Signature) != SignatureLength {
		return fmt.Errorf("signature must be %d bytes", SignatureLength)
	}
	signingBytes, err := cmd.SigningBytes()
	if err != nil {
		return err
	}
	if !ed25519.Verify(s.Wallet.PublicKey, signingBytes, cmd.Signature) {
		return errors.New("invalid signature")
	}
	return nil
}

func (cmd ExternalCommand) SigningBytes() ([]byte, error) {
	if err := cmd.validateShape(); err != nil {
		return nil, err
	}
	encoder := newByteEncoder()
	encoder.writeString(signingBytesDomain)
	encoder.writeBytes(cmd.WalletAddress)
	encoder.writeUint64(cmd.WalletID)
	encoder.writeUint64(cmd.Seqno)
	encoder.writeInt64(cmd.ValidUntil)
	encoder.writeString(string(cmd.Kind))
	encoder.writeBool(cmd.SignatureAllowed)
	encoder.writeBytes(cmd.Extension)
	encoder.writeUint64(uint64(len(cmd.Messages)))
	for _, msg := range cmd.Messages {
		encoder.writeBytes(msg.To)
		encoder.writeString(msg.Amount.String())
		encoder.writeBytes(msg.Payload)
	}
	return encoder.bytes(), nil
}

func (cmd ExternalCommand) validateShape() error {
	switch cmd.Kind {
	case CommandSend:
		return ValidateOutboundMessages(cmd.Messages, true)
	case CommandUpdateSignatureAllowed:
		if len(cmd.Messages) != 0 {
			return errors.New("signature update command must not include outbound messages")
		}
		return nil
	case CommandInstallExtension, CommandRemoveExtension:
		if len(cmd.Messages) != 0 {
			return errors.New("extension admin command must not include outbound messages")
		}
		return aetraaddress.RejectZeroAddress("wallet extension", cmd.Extension)
	default:
		return errors.New("unknown wallet command kind")
	}
}

func ValidateOutboundMessages(messages []OutboundMessage, requireNonEmpty bool) error {
	if requireNonEmpty && len(messages) == 0 {
		return errors.New("multi-send message batch must be non-empty")
	}
	if len(messages) > MaxMultiSendCount {
		return fmt.Errorf("multi-send message batch must be <= %d", MaxMultiSendCount)
	}
	for _, msg := range messages {
		if err := aetraaddress.RejectZeroAddress("message recipient", msg.To); err != nil {
			return err
		}
		if !msg.Amount.IsValid() {
			return errors.New("message amount must be valid")
		}
		if len(msg.Payload) > MaxPayloadBytes {
			return fmt.Errorf("message payload length must be <= %d", MaxPayloadBytes)
		}
	}
	return nil
}
