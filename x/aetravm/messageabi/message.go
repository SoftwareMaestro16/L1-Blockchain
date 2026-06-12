package messageabi

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	Magic				= "AVMM"
	Version			uint16	= 1
	MaxAddressTextBytes		= 128
	DefaultMaxBodyBytes		= 64 * 1024
	DefaultMaxMetadata		= 1024
	DefaultMaxStateInit		= 64 * 1024
	DefaultMaxSignature		= 4096
)

type Kind uint8

const (
	KindExternal	Kind	= 1
	KindInternal	Kind	= 2
	KindBounced	Kind	= 3
	KindSystem	Kind	= 4
)

type Params struct {
	MaxBodyBytes		uint32
	MaxMetadataBytes	uint32
	MaxStateInitBytes	uint32
	MaxSignatureBytes	uint32
}

type AddressPair struct {
	User	string	`json:"user"`
	Raw	string	`json:"raw"`
}

type Message struct {
	Kind		Kind		`json:"kind"`
	Opcode		uint64		`json:"opcode"`
	QueryID		uint64		`json:"query_id"`
	Sender		AddressPair	`json:"sender"`
	Destination	AddressPair	`json:"destination"`
	ValueNAET	uint64		`json:"value_naet"`
	Bounce		bool		`json:"bounce"`
	Bounced		bool		`json:"bounced"`
	DeadlineBlock	uint64		`json:"deadline_block"`
	GasLimit	uint64		`json:"gas_limit"`
	Body		[]byte		`json:"body"`
	StateInit	[]byte		`json:"state_init,omitempty"`
	Metadata	[]byte		`json:"metadata,omitempty"`
	Signature	[]byte		`json:"signature,omitempty"`
}

type DebugMessage struct {
	Magic		string		`json:"magic"`
	Version		uint16		`json:"version"`
	MessageID	string		`json:"message_id"`
	Kind		string		`json:"kind"`
	Opcode		uint64		`json:"opcode"`
	QueryID		uint64		`json:"query_id"`
	Sender		AddressPair	`json:"sender"`
	Destination	AddressPair	`json:"destination"`
	ValueNAET	uint64		`json:"value_naet"`
	Bounce		bool		`json:"bounce"`
	Bounced		bool		`json:"bounced"`
	DeadlineBlock	uint64		`json:"deadline_block"`
	GasLimit	uint64		`json:"gas_limit"`
	BodyHex		string		`json:"body_hex"`
	StateInitHex	string		`json:"state_init_hex,omitempty"`
	MetadataHex	string		`json:"metadata_hex,omitempty"`
	SignatureHex	string		`json:"signature_hex,omitempty"`
}

func DefaultParams() Params {
	return Params{
		MaxBodyBytes:		DefaultMaxBodyBytes,
		MaxMetadataBytes:	DefaultMaxMetadata,
		MaxStateInitBytes:	DefaultMaxStateInit,
		MaxSignatureBytes:	DefaultMaxSignature,
	}
}

func (p Params) Normalize() Params {
	if p.MaxBodyBytes == 0 {
		p.MaxBodyBytes = DefaultMaxBodyBytes
	}
	if p.MaxMetadataBytes == 0 {
		p.MaxMetadataBytes = DefaultMaxMetadata
	}
	if p.MaxStateInitBytes == 0 {
		p.MaxStateInitBytes = DefaultMaxStateInit
	}
	if p.MaxSignatureBytes == 0 {
		p.MaxSignatureBytes = DefaultMaxSignature
	}
	return p
}

func (p Params) Validate() error {
	p = p.Normalize()
	if p.MaxBodyBytes == 0 {
		return errors.New("AVM message ABI max body bytes must be positive")
	}
	if p.MaxMetadataBytes == 0 {
		return errors.New("AVM message ABI max metadata bytes must be positive")
	}
	if p.MaxStateInitBytes == 0 {
		return errors.New("AVM message ABI max state init bytes must be positive")
	}
	if p.MaxSignatureBytes == 0 {
		return errors.New("AVM message ABI max signature bytes must be positive")
	}
	return nil
}

func Encode(msg Message, params Params) ([]byte, error) {
	return CanonicalBytes(msg, params)
}

func Decode(encoded []byte, params Params, currentHeight uint64) (Message, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return Message{}, err
	}
	decoder := byteDecoder{data: encoded}
	if string(decoder.readFixed(len(Magic))) != Magic {
		return Message{}, errors.New("AVM message ABI invalid magic")
	}
	version := decoder.readU16()
	if version != Version {
		return Message{}, fmt.Errorf("AVM message ABI unsupported version %d", version)
	}
	msg := Message{
		Kind:		Kind(decoder.readU8()),
		Opcode:		decoder.readU64(),
		QueryID:	decoder.readU64(),
		Sender:		AddressPair{User: decoder.readString(MaxAddressTextBytes), Raw: decoder.readString(MaxAddressTextBytes)},
		Destination:	AddressPair{User: decoder.readString(MaxAddressTextBytes), Raw: decoder.readString(MaxAddressTextBytes)},
		ValueNAET:	decoder.readU64(),
		Bounce:		decoder.readBool(),
		Bounced:	decoder.readBool(),
		DeadlineBlock:	decoder.readU64(),
		GasLimit:	decoder.readU64(),
		Body:		decoder.readBytes(params.MaxBodyBytes),
		StateInit:	decoder.readBytes(params.MaxStateInitBytes),
		Metadata:	decoder.readBytes(params.MaxMetadataBytes),
		Signature:	decoder.readBytes(params.MaxSignatureBytes),
	}
	if decoder.err != nil {
		return Message{}, decoder.err
	}
	if decoder.pos != len(encoded) {
		return Message{}, errors.New("AVM message ABI trailing bytes")
	}
	if err := msg.Validate(params, currentHeight); err != nil {
		return Message{}, err
	}
	return msg.Clone(), nil
}

func CanonicalBytes(msg Message, params Params) ([]byte, error) {
	params = params.Normalize()
	if err := msg.Validate(params, 0); err != nil {
		return nil, err
	}
	msg = msg.Clone()
	var buf bytes.Buffer
	buf.WriteString(Magic)
	writeU16(&buf, Version)
	writeU8(&buf, uint8(msg.Kind))
	writeU64(&buf, msg.Opcode)
	writeU64(&buf, msg.QueryID)
	writeString(&buf, msg.Sender.User)
	writeString(&buf, msg.Sender.Raw)
	writeString(&buf, msg.Destination.User)
	writeString(&buf, msg.Destination.Raw)
	writeU64(&buf, msg.ValueNAET)
	writeBool(&buf, msg.Bounce)
	writeBool(&buf, msg.Bounced)
	writeU64(&buf, msg.DeadlineBlock)
	writeU64(&buf, msg.GasLimit)
	writeBytes(&buf, msg.Body)
	writeBytes(&buf, msg.StateInit)
	writeBytes(&buf, msg.Metadata)
	writeBytes(&buf, msg.Signature)
	return buf.Bytes(), nil
}

func MessageID(msg Message, params Params) ([32]byte, error) {
	encoded, err := CanonicalBytes(msg, params)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(encoded), nil
}

func MessageIDHex(msg Message, params Params) (string, error) {
	id, err := MessageID(msg, params)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id[:]), nil
}

func CanonicalHashHex(msg Message, params Params) (string, error) {
	return MessageIDHex(msg, params)
}

func DebugJSON(msg Message, params Params) ([]byte, error) {
	if err := msg.Validate(params, 0); err != nil {
		return nil, err
	}
	id, err := MessageIDHex(msg, params)
	if err != nil {
		return nil, err
	}
	debug := DebugMessage{
		Magic:		Magic,
		Version:	Version,
		MessageID:	id,
		Kind:		msg.Kind.String(),
		Opcode:		msg.Opcode,
		QueryID:	msg.QueryID,
		Sender:		msg.Sender,
		Destination:	msg.Destination,
		ValueNAET:	msg.ValueNAET,
		Bounce:		msg.Bounce,
		Bounced:	msg.Bounced,
		DeadlineBlock:	msg.DeadlineBlock,
		GasLimit:	msg.GasLimit,
		BodyHex:	hex.EncodeToString(msg.Body),
		StateInitHex:	hex.EncodeToString(msg.StateInit),
		MetadataHex:	hex.EncodeToString(msg.Metadata),
		SignatureHex:	hex.EncodeToString(msg.Signature),
	}
	return json.Marshal(debug)
}

func (m Message) Validate(params Params, currentHeight uint64) error {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if !m.Kind.Valid() {
		return fmt.Errorf("AVM message ABI invalid kind %d", m.Kind)
	}
	if err := validateAddressPair("sender", m.Sender); err != nil {
		return err
	}
	if err := validateAddressPair("destination", m.Destination); err != nil {
		return err
	}
	if m.GasLimit == 0 {
		return errors.New("AVM message ABI gas limit must be positive")
	}
	if len(m.Body) > int(params.MaxBodyBytes) {
		return fmt.Errorf("AVM message ABI body exceeds maximum size %d", params.MaxBodyBytes)
	}
	if len(m.Metadata) > int(params.MaxMetadataBytes) {
		return fmt.Errorf("AVM message ABI metadata exceeds maximum size %d", params.MaxMetadataBytes)
	}
	if len(m.StateInit) > int(params.MaxStateInitBytes) {
		return fmt.Errorf("AVM message ABI state init exceeds maximum size %d", params.MaxStateInitBytes)
	}
	if len(m.Signature) > int(params.MaxSignatureBytes) {
		return fmt.Errorf("AVM message ABI signature exceeds maximum size %d", params.MaxSignatureBytes)
	}
	if currentHeight != 0 && m.DeadlineBlock != 0 && currentHeight > m.DeadlineBlock {
		return errors.New("AVM message ABI message expired")
	}
	if m.Bounce && m.Bounced {
		return errors.New("AVM message ABI invalid bounce combination")
	}
	if m.Kind == KindBounced {
		if !m.Bounced {
			return errors.New("AVM message ABI bounced kind requires bounced flag")
		}
		if m.Bounce {
			return errors.New("AVM message ABI bounced message cannot request bounce")
		}
	} else if m.Bounced {
		return errors.New("AVM message ABI bounced flag requires bounced kind")
	}
	return nil
}

func (m Message) Clone() Message {
	out := m
	out.Sender.User = strings.TrimSpace(out.Sender.User)
	out.Sender.Raw = strings.TrimSpace(out.Sender.Raw)
	out.Destination.User = strings.TrimSpace(out.Destination.User)
	out.Destination.Raw = strings.TrimSpace(out.Destination.Raw)
	out.Body = append([]byte(nil), m.Body...)
	out.StateInit = append([]byte(nil), m.StateInit...)
	out.Metadata = append([]byte(nil), m.Metadata...)
	out.Signature = append([]byte(nil), m.Signature...)
	return out
}

func (k Kind) Valid() bool {
	switch k {
	case KindExternal, KindInternal, KindBounced, KindSystem:
		return true
	default:
		return false
	}
}

func (k Kind) String() string {
	switch k {
	case KindExternal:
		return "external"
	case KindInternal:
		return "internal"
	case KindBounced:
		return "bounced"
	case KindSystem:
		return "system"
	default:
		return fmt.Sprintf("unknown:%d", k)
	}
}

func AddressPairFromUser(user string) (AddressPair, error) {
	user = strings.TrimSpace(user)
	bz, err := addressing.Parse(user)
	if err != nil {
		return AddressPair{}, err
	}
	formatted, err := addressing.FormatUserFriendly(bz)
	if err != nil {
		return AddressPair{}, err
	}
	return AddressPair{User: formatted, Raw: addressing.Format(bz)}, nil
}

func validateAddressPair(field string, pair AddressPair) error {
	pair.User = strings.TrimSpace(pair.User)
	pair.Raw = strings.TrimSpace(pair.Raw)
	if !strings.HasPrefix(pair.User, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("AVM message ABI %s user address must use AE format", field)
	}
	userBytes, err := addressing.Parse(pair.User)
	if err != nil {
		return fmt.Errorf("AVM message ABI invalid %s user address: %w", field, err)
	}
	formattedUser, err := addressing.FormatUserFriendly(userBytes)
	if err != nil {
		return fmt.Errorf("AVM message ABI invalid %s user address: %w", field, err)
	}
	if formattedUser != pair.User {
		return fmt.Errorf("AVM message ABI %s user address is not canonical", field)
	}
	if !strings.HasPrefix(pair.Raw, addressing.RawPrefix) {
		return fmt.Errorf("AVM message ABI %s raw address must use 4: format", field)
	}
	rawBytes, err := addressing.Parse(pair.Raw)
	if err != nil {
		return fmt.Errorf("AVM message ABI invalid %s raw address: %w", field, err)
	}
	if addressing.Format(rawBytes) != pair.Raw {
		return fmt.Errorf("AVM message ABI %s raw address is not canonical", field)
	}
	if !bytes.Equal(addressing.FromRawPayload(mustRawPayload(userBytes)), addressing.FromRawPayload(mustRawPayload(rawBytes))) {
		return fmt.Errorf("AVM message ABI %s AE/raw address pair mismatch", field)
	}
	if addressing.IsZero(userBytes) || addressing.IsZero(rawBytes) {
		return fmt.Errorf("AVM message ABI %s address must not be zero", field)
	}
	return nil
}

func mustRawPayload(bz []byte) []byte {
	raw, err := addressing.ToRawPayload(bz)
	if err != nil {
		return nil
	}
	return raw
}

type byteDecoder struct {
	data	[]byte
	pos	int
	err	error
}

func (d *byteDecoder) readFixed(n int) []byte {
	if d.err != nil {
		return nil
	}
	if n < 0 || d.pos+n > len(d.data) {
		d.err = errors.New("AVM message ABI truncated data")
		return nil
	}
	out := d.data[d.pos : d.pos+n]
	d.pos += n
	return out
}

func (d *byteDecoder) readU8() uint8 {
	bz := d.readFixed(1)
	if d.err != nil {
		return 0
	}
	return bz[0]
}

func (d *byteDecoder) readBool() bool {
	value := d.readU8()
	if d.err != nil {
		return false
	}
	switch value {
	case 0:
		return false
	case 1:
		return true
	default:
		d.err = errors.New("AVM message ABI invalid bool")
		return false
	}
}

func (d *byteDecoder) readU16() uint16 {
	bz := d.readFixed(2)
	if d.err != nil {
		return 0
	}
	return binary.BigEndian.Uint16(bz)
}

func (d *byteDecoder) readU32() uint32 {
	bz := d.readFixed(4)
	if d.err != nil {
		return 0
	}
	return binary.BigEndian.Uint32(bz)
}

func (d *byteDecoder) readU64() uint64 {
	bz := d.readFixed(8)
	if d.err != nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (d *byteDecoder) readString(max uint32) string {
	return string(d.readBytes(max))
}

func (d *byteDecoder) readBytes(max uint32) []byte {
	size := d.readU32()
	if d.err != nil {
		return nil
	}
	if size > max {
		d.err = fmt.Errorf("AVM message ABI field exceeds maximum size %d", max)
		return nil
	}
	return append([]byte(nil), d.readFixed(int(size))...)
}

func writeU8(buf *bytes.Buffer, value uint8) {
	buf.WriteByte(value)
}

func writeBool(buf *bytes.Buffer, value bool) {
	if value {
		buf.WriteByte(1)
		return
	}
	buf.WriteByte(0)
}

func writeU16(buf *bytes.Buffer, value uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], value)
	buf.Write(b[:])
}

func writeU32(buf *bytes.Buffer, value uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], value)
	buf.Write(b[:])
}

func writeU64(buf *bytes.Buffer, value uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], value)
	buf.Write(b[:])
}

func writeString(buf *bytes.Buffer, value string) {
	writeBytes(buf, []byte(value))
}

func writeBytes(buf *bytes.Buffer, value []byte) {
	if len(value) > int(^uint32(0)) {
		panic("AVM message ABI field too large")
	}
	writeU32(buf, uint32(len(value)))
	buf.Write(value)
}
