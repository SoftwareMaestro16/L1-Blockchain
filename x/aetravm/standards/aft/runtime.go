package aft

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
	OpcodeMint        = uint32(0x4146_0001)
	OpcodeTransfer    = uint32(0x4146_0002)
	OpcodeBurn        = uint32(0x4146_0003)
	OpcodeChangeAdmin = uint32(0x4146_0004)
	OpcodeRenounce    = uint32(0x4146_0005)
	OpcodeMetadata    = uint32(0x4146_0006)
)

type MintMessage struct {
	Caller    sdk.AccAddress
	Recipient sdk.AccAddress
	Amount    sdkmath.Int
}

type TransferMessage struct {
	Owner     sdk.AccAddress
	Recipient sdk.AccAddress
	Amount    sdkmath.Int
}

type BurnMessage struct {
	Owner  sdk.AccAddress
	Amount sdkmath.Int
}

func EncodeMintMessage(msg MintMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Caller)
	encoder.writeAddress(msg.Recipient)
	encoder.writeInt(msg.Amount)
	return encoder.bytes()
}

func EncodeTransferMessage(msg TransferMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Owner)
	encoder.writeAddress(msg.Recipient)
	encoder.writeInt(msg.Amount)
	return encoder.bytes()
}

func EncodeBurnMessage(msg BurnMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Owner)
	encoder.writeInt(msg.Amount)
	return encoder.bytes()
}

func EncodeChangeAdminMessage(caller, nextAdmin sdk.AccAddress) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(caller)
	encoder.writeAddress(nextAdmin)
	return encoder.bytes()
}

func EncodeChangeMetadataMessage(caller sdk.AccAddress, metadata TokenMetadata) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(caller)
	encoder.writeMetadata(metadata)
	return encoder.bytes()
}

func EncodeRenounceMessage(caller sdk.AccAddress) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(caller)
	return encoder.bytes()
}

func (s *State) AsyncHandler() async.Handler {
	return func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
		if err := ValidateOperationFees(sdk.NewCoins(msg.ForwardFee)); err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		if msg.Bounced {
			transfer, err := decodeTransferMessage(msg.Body)
			if err != nil {
				return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
			}
			if err := s.BounceTransfer(transfer.Owner, transfer.Amount, msg.QueryID); err != nil {
				return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
			}
			return async.ExecutionResult{NewState: s.asyncStateBytes(), StorageWrites: 1, ResultCode: async.ResultOK}
		}

		var outgoing []async.MessageEnvelope
		var err error
		switch msg.Opcode {
		case OpcodeMint:
			err = s.applyMint(msg.Body)
		case OpcodeTransfer:
			outgoing, err = s.applyTransfer(msg)
		case OpcodeBurn:
			err = s.applyBurn(msg.Body, msg.QueryID)
		case OpcodeChangeAdmin:
			err = s.applyChangeAdmin(msg.Body)
		case OpcodeMetadata:
			err = s.applyChangeMetadata(msg.Body)
		case OpcodeRenounce:
			err = s.applyRenounce(msg.Body)
		default:
			err = errors.New("unknown AFT-44 opcode")
		}
		if err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		return async.ExecutionResult{
			NewState:      s.asyncStateBytes(),
			Outgoing:      outgoing,
			GasUsed:       10_000 + uint64(len(outgoing))*1_000,
			StorageWrites: 1,
			ResultCode:    async.ResultOK,
		}
	}
}

func (s *State) applyMint(body []byte) error {
	msg, err := decodeMintMessage(body)
	if err != nil {
		return err
	}
	return s.Mint(msg.Caller, msg.Recipient, msg.Amount)
}

func (s *State) applyTransfer(envelope async.MessageEnvelope) ([]async.MessageEnvelope, error) {
	msg, err := decodeTransferMessage(envelope.Body)
	if err != nil {
		return nil, err
	}
	if err := s.Transfer(msg.Owner, msg.Recipient, msg.Amount, envelope.QueryID); err != nil {
		return nil, err
	}
	wallet, err := s.WalletAddress(msg.Recipient)
	if err != nil {
		return nil, err
	}
	return []async.MessageEnvelope{{
		Destination: wallet,
		Value:       sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
		Opcode:      OpcodeTransfer,
		QueryID:     envelope.QueryID,
		Body:        append([]byte(nil), envelope.Body...),
		Bounce:      true,
		GasLimit:    envelope.GasLimit,
		ForwardFee:  sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
	}}, nil
}

func (s *State) applyBurn(body []byte, queryID uint64) error {
	msg, err := decodeBurnMessage(body)
	if err != nil {
		return err
	}
	return s.Burn(msg.Owner, msg.Amount, queryID)
}

func (s *State) applyChangeAdmin(body []byte) error {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return err
	}
	nextAdmin, err := decoder.readAddress()
	if err != nil {
		return err
	}
	if err := decoder.done(); err != nil {
		return err
	}
	return s.ChangeAdmin(caller, nextAdmin)
}

func (s *State) applyChangeMetadata(body []byte) error {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return err
	}
	metadata, err := decoder.readMetadata()
	if err != nil {
		return err
	}
	if err := decoder.done(); err != nil {
		return err
	}
	return s.ChangeMetadata(caller, metadata)
}

func (s *State) applyRenounce(body []byte) error {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return err
	}
	if err := decoder.done(); err != nil {
		return err
	}
	return s.RenounceAdmin(caller)
}

func decodeMintMessage(body []byte) (MintMessage, error) {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return MintMessage{}, err
	}
	recipient, err := decoder.readAddress()
	if err != nil {
		return MintMessage{}, err
	}
	amount, err := decoder.readInt()
	if err != nil {
		return MintMessage{}, err
	}
	return MintMessage{Caller: caller, Recipient: recipient, Amount: amount}, decoder.done()
}

func decodeTransferMessage(body []byte) (TransferMessage, error) {
	decoder := newRuntimeDecoder(body)
	owner, err := decoder.readAddress()
	if err != nil {
		return TransferMessage{}, err
	}
	recipient, err := decoder.readAddress()
	if err != nil {
		return TransferMessage{}, err
	}
	amount, err := decoder.readInt()
	if err != nil {
		return TransferMessage{}, err
	}
	return TransferMessage{Owner: owner, Recipient: recipient, Amount: amount}, decoder.done()
}

func decodeBurnMessage(body []byte) (BurnMessage, error) {
	decoder := newRuntimeDecoder(body)
	owner, err := decoder.readAddress()
	if err != nil {
		return BurnMessage{}, err
	}
	amount, err := decoder.readInt()
	if err != nil {
		return BurnMessage{}, err
	}
	return BurnMessage{Owner: owner, Amount: amount}, decoder.done()
}

func (s *State) asyncStateBytes() []byte {
	encoder := newRuntimeEncoder()
	encoder.writeString(StandardName)
	encoder.writeInt(s.Master.TotalSupply)
	encoder.writeUint64(uint64(len(s.Wallets)))
	return encoder.bytes()
}

type runtimeEncoder struct {
	out []byte
}

func newRuntimeEncoder() *runtimeEncoder {
	return &runtimeEncoder{}
}

func (e *runtimeEncoder) writeAddress(address sdk.AccAddress) {
	e.writeBytes(address)
}

func (e *runtimeEncoder) writeInt(value sdkmath.Int) {
	e.writeString(value.String())
}

func (e *runtimeEncoder) writeString(value string) {
	e.writeBytes([]byte(value))
}

func (e *runtimeEncoder) writeMetadata(metadata TokenMetadata) {
	e.writeString(metadata.Name)
	e.writeString(metadata.Symbol)
	e.writeUint64(uint64(metadata.Decimals))
	e.writeString(metadata.ContentRef)
	e.writeString(metadata.DisplayName)
}

func (e *runtimeEncoder) writeBytes(bz []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(bz)))
	e.out = append(e.out, length[:]...)
	e.out = append(e.out, bz...)
}

func (e *runtimeEncoder) writeUint64(value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	e.out = append(e.out, bz[:]...)
}

func (e *runtimeEncoder) bytes() []byte {
	return append([]byte(nil), e.out...)
}

type runtimeDecoder struct {
	reader *bytes.Reader
}

func newRuntimeDecoder(bz []byte) *runtimeDecoder {
	return &runtimeDecoder{reader: bytes.NewReader(bz)}
}

func (d *runtimeDecoder) readAddress() (sdk.AccAddress, error) {
	bz, err := d.readBytes()
	return sdk.AccAddress(bz), err
}

func (d *runtimeDecoder) readInt() (sdkmath.Int, error) {
	value, err := d.readString()
	if err != nil {
		return sdkmath.Int{}, err
	}
	amount, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid integer amount: %s", value)
	}
	return amount, nil
}

func (d *runtimeDecoder) readString() (string, error) {
	bz, err := d.readBytes()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (d *runtimeDecoder) readUint64() (uint64, error) {
	var bz [8]byte
	if _, err := io.ReadFull(d.reader, bz[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bz[:]), nil
}

func (d *runtimeDecoder) readMetadata() (TokenMetadata, error) {
	name, err := d.readString()
	if err != nil {
		return TokenMetadata{}, err
	}
	symbol, err := d.readString()
	if err != nil {
		return TokenMetadata{}, err
	}
	decimals, err := d.readUint64()
	if err != nil {
		return TokenMetadata{}, err
	}
	contentRef, err := d.readString()
	if err != nil {
		return TokenMetadata{}, err
	}
	displayName, err := d.readString()
	if err != nil {
		return TokenMetadata{}, err
	}
	return TokenMetadata{Name: name, Symbol: symbol, Decimals: uint32(decimals), ContentRef: contentRef, DisplayName: displayName}, nil
}

func (d *runtimeDecoder) readBytes() ([]byte, error) {
	var length [4]byte
	if _, err := io.ReadFull(d.reader, length[:]); err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(length[:])
	if size > async.DefaultParams().MaxBodySize {
		return nil, errors.New("encoded field length exceeds max body size")
	}
	out := make([]byte, size)
	if _, err := io.ReadFull(d.reader, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *runtimeDecoder) done() error {
	if d.reader.Len() != 0 {
		return errors.New("encoded AFT-44 message has trailing data")
	}
	return nil
}
