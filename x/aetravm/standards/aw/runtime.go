package aw

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

const (
	OpcodeSignedExternal = uint32(0x4157_0001)
)

func EncodeExternalCommand(cmd ExternalCommand) ([]byte, error) {
	if err := cmd.validateShape(); err != nil {
		return nil, err
	}
	encoder := newByteEncoder()
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
	encoder.writeBytes(cmd.Signature)
	return encoder.bytes(), nil
}

func DecodeExternalCommand(bz []byte) (ExternalCommand, error) {
	decoder := newByteDecoder(bz)
	wallet, err := decoder.readBytes()
	if err != nil {
		return ExternalCommand{}, err
	}
	walletID, err := decoder.readUint64()
	if err != nil {
		return ExternalCommand{}, err
	}
	seqno, err := decoder.readUint64()
	if err != nil {
		return ExternalCommand{}, err
	}
	validUntil, err := decoder.readInt64()
	if err != nil {
		return ExternalCommand{}, err
	}
	kind, err := decoder.readString()
	if err != nil {
		return ExternalCommand{}, err
	}
	signatureAllowed, err := decoder.readBool()
	if err != nil {
		return ExternalCommand{}, err
	}
	extension, err := decoder.readBytes()
	if err != nil {
		return ExternalCommand{}, err
	}
	count, err := decoder.readUint64()
	if err != nil {
		return ExternalCommand{}, err
	}
	if count > MaxMultiSendCount {
		return ExternalCommand{}, fmt.Errorf("multi-send message batch must be <= %d", MaxMultiSendCount)
	}
	messages := make([]OutboundMessage, count)
	for i := range messages {
		to, err := decoder.readBytes()
		if err != nil {
			return ExternalCommand{}, err
		}
		amountText, err := decoder.readString()
		if err != nil {
			return ExternalCommand{}, err
		}
		amount, err := sdk.ParseCoinsNormalized(amountText)
		if err != nil {
			return ExternalCommand{}, err
		}
		payload, err := decoder.readBytes()
		if err != nil {
			return ExternalCommand{}, err
		}
		messages[i] = OutboundMessage{To: sdk.AccAddress(to), Amount: amount, Payload: payload}
	}
	signature, err := decoder.readBytes()
	if err != nil {
		return ExternalCommand{}, err
	}
	if err := decoder.done(); err != nil {
		return ExternalCommand{}, err
	}
	cmd := ExternalCommand{
		WalletAddress:    sdk.AccAddress(wallet),
		WalletID:         walletID,
		Seqno:            seqno,
		ValidUntil:       validUntil,
		Kind:             CommandKind(kind),
		Messages:         messages,
		SignatureAllowed: signatureAllowed,
		Extension:        sdk.AccAddress(extension),
		Signature:        signature,
	}
	if err := cmd.validateShape(); err != nil {
		return ExternalCommand{}, err
	}
	return cmd, nil
}

func (s *State) AsyncHandler(now int64) async.Handler {
	return func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
		if msg.Bounced {
			return async.ExecutionResult{NewState: s.asyncStateBytes(), ResultCode: async.ResultOK}
		}
		if msg.Opcode != OpcodeSignedExternal {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: "unknown AW-5 opcode"}
		}
		cmd, err := DecodeExternalCommand(msg.Body)
		if err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		before := len(s.SentMessages)
		if err := s.ApplyExternalCommand(cmd, now, sdk.NewCoins(msg.ForwardFee)); err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		outgoing := s.asyncOutgoing(s.SentMessages[before:], msg)
		return async.ExecutionResult{
			NewState:      s.asyncStateBytes(),
			Outgoing:      outgoing,
			GasUsed:       10_000 + uint64(len(outgoing))*1_000,
			StorageWrites: 1,
			ResultCode:    async.ResultOK,
		}
	}
}

func (s *State) asyncOutgoing(messages []OutboundMessage, inbound async.MessageEnvelope) []async.MessageEnvelope {
	out := make([]async.MessageEnvelope, 0, len(messages))
	for _, msg := range messages {
		out = append(out, async.MessageEnvelope{
			Destination: msg.To,
			Value:       sdk.NewCoin(appparams.BaseDenom, msg.Amount.AmountOf(appparams.BaseDenom)),
			Opcode:      OpcodeSignedExternal,
			QueryID:     inbound.QueryID,
			Body:        append([]byte(nil), msg.Payload...),
			Bounce:      true,
			GasLimit:    inbound.GasLimit,
			ForwardFee:  sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
		})
	}
	return out
}

func (s *State) asyncStateBytes() []byte {
	encoder := newByteEncoder()
	encoder.writeString(StandardName)
	encoder.writeUint64(s.Wallet.Seqno)
	encoder.writeUint64(uint64(len(s.Wallet.Extensions)))
	encoder.writeUint64(uint64(len(s.SentMessages)))
	return encoder.bytes()
}

type byteDecoder struct {
	reader *bytes.Reader
}

func newByteDecoder(bz []byte) *byteDecoder {
	return &byteDecoder{reader: bytes.NewReader(bz)}
}

func (d *byteDecoder) readBytes() ([]byte, error) {
	length, err := d.readUint64()
	if err != nil {
		return nil, err
	}
	if length > MaxPayloadBytes*uint64(MaxMultiSendCount+1) {
		return nil, errors.New("encoded field length is too large")
	}
	out := make([]byte, length)
	if _, err := io.ReadFull(d.reader, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *byteDecoder) readString() (string, error) {
	bz, err := d.readBytes()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (d *byteDecoder) readUint64() (uint64, error) {
	var bz [8]byte
	if _, err := io.ReadFull(d.reader, bz[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bz[:]), nil
}

func (d *byteDecoder) readInt64() (int64, error) {
	value, err := d.readUint64()
	return int64(value), err
}

func (d *byteDecoder) readBool() (bool, error) {
	value, err := d.reader.ReadByte()
	if err != nil {
		return false, err
	}
	switch value {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errors.New("encoded bool must be 0 or 1")
	}
}

func (d *byteDecoder) done() error {
	if d.reader.Len() != 0 {
		return errors.New("encoded command has trailing data")
	}
	return nil
}

func naetValue(amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(appparams.BaseDenom, amount)
}
