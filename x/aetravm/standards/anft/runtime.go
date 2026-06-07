package anft

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

const (
	OpcodeMintNFT   = uint32(0x4e46_0001)
	OpcodeTransfer  = uint32(0x4e46_0002)
	OpcodeMintSBT   = uint32(0x5342_0001)
	OpcodeRevokeSBT = uint32(0x5342_0002)
	OpcodeProveSBT  = uint32(0x5342_0003)
)

type MintNFTMessage struct {
	Caller   sdk.AccAddress
	Owner    sdk.AccAddress
	Metadata Metadata
}

type MintSBTMessage struct {
	Caller    sdk.AccAddress
	Owner     sdk.AccAddress
	Authority sdk.AccAddress
	Metadata  Metadata
}

type TransferMessage struct {
	Caller      sdk.AccAddress
	ItemAddress sdk.AccAddress
	NewOwner    sdk.AccAddress
}

type RevokeSBTMessage struct {
	Caller      sdk.AccAddress
	ItemAddress sdk.AccAddress
	RevokedAt   int64
	Reason      string
}

func EncodeMintNFTMessage(msg MintNFTMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Caller)
	encoder.writeAddress(msg.Owner)
	encoder.writeMetadata(msg.Metadata)
	return encoder.bytes()
}

func EncodeMintSBTMessage(msg MintSBTMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Caller)
	encoder.writeAddress(msg.Owner)
	encoder.writeAddress(msg.Authority)
	encoder.writeMetadata(msg.Metadata)
	return encoder.bytes()
}

func EncodeTransferMessage(msg TransferMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Caller)
	encoder.writeAddress(msg.ItemAddress)
	encoder.writeAddress(msg.NewOwner)
	return encoder.bytes()
}

func EncodeRevokeSBTMessage(msg RevokeSBTMessage) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(msg.Caller)
	encoder.writeAddress(msg.ItemAddress)
	encoder.writeInt64(msg.RevokedAt)
	encoder.writeString(msg.Reason)
	return encoder.bytes()
}

func EncodeProveSBTMessage(itemAddress, owner sdk.AccAddress) []byte {
	encoder := newRuntimeEncoder()
	encoder.writeAddress(itemAddress)
	encoder.writeAddress(owner)
	return encoder.bytes()
}

func (s *State) AsyncHandler() async.Handler {
	return func(contract async.ContractAccount, msg async.MessageEnvelope) async.ExecutionResult {
		if err := feestypes.ValidateFeeCoins(feestypes.DefaultParams(), sdk.NewCoins(msg.ForwardFee), true); err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		if msg.Bounced {
			return async.ExecutionResult{NewState: s.asyncStateBytes(), ResultCode: async.ResultOK}
		}
		var err error
		switch msg.Opcode {
		case OpcodeMintNFT:
			err = s.applyMintNFT(msg.Body)
		case OpcodeMintSBT:
			err = s.applyMintSBT(msg.Body)
		case OpcodeTransfer:
			err = s.applyTransfer(msg.Body)
		case OpcodeRevokeSBT:
			err = s.applyRevokeSBT(msg.Body)
		case OpcodeProveSBT:
			err = s.applyProveSBT(msg.Body)
		default:
			err = errors.New("unknown ANFT-66/ASBT-67 opcode")
		}
		if err != nil {
			return async.ExecutionResult{NewState: contract.State, ResultCode: async.ResultExecutionFailed, Error: err.Error()}
		}
		return async.ExecutionResult{
			NewState:      s.asyncStateBytes(),
			GasUsed:       10_000,
			StorageWrites: 1,
			ResultCode:    async.ResultOK,
		}
	}
}

func (s *State) applyMintNFT(body []byte) error {
	msg, err := decodeMintNFTMessage(body)
	if err != nil {
		return err
	}
	_, err = s.MintNFT(msg.Caller, msg.Owner, msg.Metadata)
	return err
}

func (s *State) applyMintSBT(body []byte) error {
	msg, err := decodeMintSBTMessage(body)
	if err != nil {
		return err
	}
	_, err = s.MintSBT(msg.Caller, msg.Owner, msg.Authority, msg.Metadata)
	return err
}

func (s *State) applyTransfer(body []byte) error {
	msg, err := decodeTransferMessage(body)
	if err != nil {
		return err
	}
	return s.TransferNFT(msg.Caller, msg.ItemAddress, msg.NewOwner)
}

func (s *State) applyRevokeSBT(body []byte) error {
	msg, err := decodeRevokeSBTMessage(body)
	if err != nil {
		return err
	}
	return s.RevokeSBT(msg.Caller, msg.ItemAddress, msg.RevokedAt, msg.Reason)
}

func (s *State) applyProveSBT(body []byte) error {
	decoder := newRuntimeDecoder(body)
	itemAddress, err := decoder.readAddress()
	if err != nil {
		return err
	}
	owner, err := decoder.readAddress()
	if err != nil {
		return err
	}
	if err := decoder.done(); err != nil {
		return err
	}
	_, err = s.ProveSBTOwnership(itemAddress, owner)
	return err
}

func decodeMintNFTMessage(body []byte) (MintNFTMessage, error) {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return MintNFTMessage{}, err
	}
	owner, err := decoder.readAddress()
	if err != nil {
		return MintNFTMessage{}, err
	}
	metadata, err := decoder.readMetadata()
	if err != nil {
		return MintNFTMessage{}, err
	}
	return MintNFTMessage{Caller: caller, Owner: owner, Metadata: metadata}, decoder.done()
}

func decodeMintSBTMessage(body []byte) (MintSBTMessage, error) {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return MintSBTMessage{}, err
	}
	owner, err := decoder.readAddress()
	if err != nil {
		return MintSBTMessage{}, err
	}
	authority, err := decoder.readAddress()
	if err != nil {
		return MintSBTMessage{}, err
	}
	metadata, err := decoder.readMetadata()
	if err != nil {
		return MintSBTMessage{}, err
	}
	return MintSBTMessage{Caller: caller, Owner: owner, Authority: authority, Metadata: metadata}, decoder.done()
}

func decodeTransferMessage(body []byte) (TransferMessage, error) {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return TransferMessage{}, err
	}
	itemAddress, err := decoder.readAddress()
	if err != nil {
		return TransferMessage{}, err
	}
	newOwner, err := decoder.readAddress()
	if err != nil {
		return TransferMessage{}, err
	}
	return TransferMessage{Caller: caller, ItemAddress: itemAddress, NewOwner: newOwner}, decoder.done()
}

func decodeRevokeSBTMessage(body []byte) (RevokeSBTMessage, error) {
	decoder := newRuntimeDecoder(body)
	caller, err := decoder.readAddress()
	if err != nil {
		return RevokeSBTMessage{}, err
	}
	itemAddress, err := decoder.readAddress()
	if err != nil {
		return RevokeSBTMessage{}, err
	}
	revokedAt, err := decoder.readInt64()
	if err != nil {
		return RevokeSBTMessage{}, err
	}
	reason, err := decoder.readString()
	if err != nil {
		return RevokeSBTMessage{}, err
	}
	return RevokeSBTMessage{Caller: caller, ItemAddress: itemAddress, RevokedAt: revokedAt, Reason: reason}, decoder.done()
}

func (s *State) asyncStateBytes() []byte {
	encoder := newRuntimeEncoder()
	encoder.writeString(NFTStandardName)
	encoder.writeString(SBTStandardName)
	encoder.writeUint64(s.Collection.NextItemIndex)
	encoder.writeUint64(uint64(len(s.Items)))
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

func (e *runtimeEncoder) writeMetadata(metadata Metadata) {
	e.writeString(metadata.Name)
	e.writeString(metadata.Symbol)
	e.writeString(metadata.ContentRef)
}

func (e *runtimeEncoder) writeString(value string) {
	e.writeBytes([]byte(value))
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

func (e *runtimeEncoder) writeInt64(value int64) {
	e.writeUint64(uint64(value))
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

func (d *runtimeDecoder) readMetadata() (Metadata, error) {
	name, err := d.readString()
	if err != nil {
		return Metadata{}, err
	}
	symbol, err := d.readString()
	if err != nil {
		return Metadata{}, err
	}
	contentRef, err := d.readString()
	if err != nil {
		return Metadata{}, err
	}
	return Metadata{Name: name, Symbol: symbol, ContentRef: contentRef}, nil
}

func (d *runtimeDecoder) readString() (string, error) {
	bz, err := d.readBytes()
	if err != nil {
		return "", err
	}
	return string(bz), nil
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

func (d *runtimeDecoder) readUint64() (uint64, error) {
	var bz [8]byte
	if _, err := io.ReadFull(d.reader, bz[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(bz[:]), nil
}

func (d *runtimeDecoder) readInt64() (int64, error) {
	value, err := d.readUint64()
	return int64(value), err
}

func (d *runtimeDecoder) done() error {
	if d.reader.Len() != 0 {
		return fmt.Errorf("encoded ANFT-66/ASBT-67 message has trailing data")
	}
	return nil
}
