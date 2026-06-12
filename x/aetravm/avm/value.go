package avm

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"unicode/utf8"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	"lukechampine.com/blake3"
)

// ValueTag defines the runtime type tag for AVM values.
// Every Value MUST carry a deterministic type tag.
// No untyped values are allowed in runtime.
type ValueTag uint8

const (
	TagNull	ValueTag	= iota
	TagBool
	TagInt8
	TagInt16
	TagInt32
	TagInt64
	TagInt128
	TagInt256
	TagUint8
	TagUint16
	TagUint32
	TagUint64
	TagUint128
	TagUint256
	TagCoins
	TagTimestamp
	TagAddress
	TagHash
	TagBytes
	TagString
	TagTuple
	TagChunkRef
	TagReaderCursor
	TagWriterHandle
	TagExecFrameRef
)

func (t ValueTag) String() string {
	switch t {
	case TagNull:
		return "null"
	case TagBool:
		return "bool"
	case TagInt8:
		return "int8"
	case TagInt16:
		return "int16"
	case TagInt32:
		return "int32"
	case TagInt64:
		return "int64"
	case TagInt128:
		return "int128"
	case TagInt256:
		return "int256"
	case TagUint8:
		return "uint8"
	case TagUint16:
		return "uint16"
	case TagUint32:
		return "uint32"
	case TagUint64:
		return "uint64"
	case TagUint128:
		return "uint128"
	case TagUint256:
		return "uint256"
	case TagCoins:
		return "coins"
	case TagTimestamp:
		return "timestamp"
	case TagAddress:
		return "address"
	case TagHash:
		return "hash"
	case TagBytes:
		return "bytes"
	case TagString:
		return "string"
	case TagTuple:
		return "tuple"
	case TagChunkRef:
		return "chunk_ref"
	case TagReaderCursor:
		return "reader_cursor"
	case TagWriterHandle:
		return "writer_handle"
	case TagExecFrameRef:
		return "exec_frame_ref"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// ValueBitWidth returns the bit width for integer types.
func ValueBitWidth(tag ValueTag) (int, bool) {
	switch tag {
	case TagInt8, TagUint8:
		return 8, true
	case TagInt16, TagUint16:
		return 16, true
	case TagInt32, TagUint32:
		return 32, true
	case TagInt64, TagUint64:
		return 64, true
	case TagInt128, TagUint128, TagCoins:
		return 128, true
	case TagInt256, TagUint256:
		return 256, true
	default:
		return 0, false
	}
}

// IsSigned returns true for signed integer types.
func IsSigned(tag ValueTag) bool {
	switch tag {
	case TagInt8, TagInt16, TagInt32, TagInt64, TagInt128, TagInt256:
		return true
	default:
		return false
	}
}

// IsInteger returns true for any integer type.
func IsInteger(tag ValueTag) bool {
	_, ok := ValueBitWidth(tag)
	return ok
}

// RuntimeValue is the AVM runtime tagged union value.
// All runtime values MUST be one of these.
// No untyped values, no implicit casts, no runtime reflection.
type RuntimeValue struct {
	Tag		ValueTag
	boolVal		bool
	intVal		*big.Int
	coinsVal	[16]byte
	addrVal		string
	hashVal		[32]byte
	bytesVal	[]byte
	strVal		string
	tupleVal	[]RuntimeValue
	chunkRef	*chunk.Chunk
	readerOff	uint32
	writerPtr	*ValueWriter
	frameRef	*KernelExecutionFrame
}

func ValueNull() RuntimeValue {
	return RuntimeValue{Tag: TagNull}
}

func ValueBool(v bool) RuntimeValue {
	return RuntimeValue{Tag: TagBool, boolVal: v}
}

func ValueInt8(v int8) RuntimeValue {
	return RuntimeValue{Tag: TagInt8, intVal: big.NewInt(int64(v))}
}

func ValueInt16(v int16) RuntimeValue {
	return RuntimeValue{Tag: TagInt16, intVal: big.NewInt(int64(v))}
}

func ValueInt32(v int32) RuntimeValue {
	return RuntimeValue{Tag: TagInt32, intVal: big.NewInt(int64(v))}
}

func ValueInt64(v int64) RuntimeValue {
	return RuntimeValue{Tag: TagInt64, intVal: big.NewInt(v)}
}

func ValueUint8(v uint8) RuntimeValue {
	return RuntimeValue{Tag: TagUint8, intVal: new(big.Int).SetUint64(uint64(v))}
}

func ValueUint16(v uint16) RuntimeValue {
	return RuntimeValue{Tag: TagUint16, intVal: new(big.Int).SetUint64(uint64(v))}
}

func ValueUint32(v uint32) RuntimeValue {
	return RuntimeValue{Tag: TagUint32, intVal: new(big.Int).SetUint64(uint64(v))}
}

func ValueUint64(v uint64) RuntimeValue {
	return RuntimeValue{Tag: TagUint64, intVal: new(big.Int).SetUint64(v)}
}

func ValueBigUint128(v *big.Int) RuntimeValue {
	b := make([]byte, 16)
	v.FillBytes(b)
	val := [16]byte{}
	copy(val[:], b)
	return RuntimeValue{Tag: TagUint128, intVal: new(big.Int).Set(v), coinsVal: val}
}

func ValueBigInt256(v *big.Int) RuntimeValue {
	return RuntimeValue{Tag: TagInt256, intVal: new(big.Int).Set(v)}
}

func ValueCoins(v *big.Int) RuntimeValue {
	b := make([]byte, 16)
	v.FillBytes(b)
	val := [16]byte{}
	copy(val[:], b)
	return RuntimeValue{Tag: TagCoins, intVal: new(big.Int).Set(v), coinsVal: val}
}

func ValueTimestamp(v uint64) RuntimeValue {
	return RuntimeValue{Tag: TagTimestamp, intVal: new(big.Int).SetUint64(v)}
}

func ValueAddress(addr string) RuntimeValue {
	return RuntimeValue{Tag: TagAddress, addrVal: addr}
}

func ValueHash(h [32]byte) RuntimeValue {
	return RuntimeValue{Tag: TagHash, hashVal: h}
}

func ValueHashFromBytes(h []byte) RuntimeValue {
	var hash [32]byte
	copy(hash[:], h)
	return RuntimeValue{Tag: TagHash, hashVal: hash}
}

func ValueBytes(b []byte) RuntimeValue {
	cp := make([]byte, len(b))
	copy(cp, b)
	return RuntimeValue{Tag: TagBytes, bytesVal: cp}
}

func ValueString(s string) RuntimeValue {
	return RuntimeValue{Tag: TagString, strVal: s}
}

func ValueTuple(elements []RuntimeValue) RuntimeValue {
	cp := make([]RuntimeValue, len(elements))
	copy(cp, elements)
	return RuntimeValue{Tag: TagTuple, tupleVal: cp}
}

func ValueEmptyTuple() RuntimeValue {
	return RuntimeValue{Tag: TagTuple, tupleVal: []RuntimeValue{}}
}

func ValueChunkRef(c *chunk.Chunk) RuntimeValue {
	return RuntimeValue{Tag: TagChunkRef, chunkRef: c}
}

func ValueReaderCursor(c *chunk.Chunk, offset uint32) RuntimeValue {
	return RuntimeValue{Tag: TagReaderCursor, chunkRef: c, readerOff: offset}
}

func ValueWriterHandle(w *ValueWriter) RuntimeValue {
	return RuntimeValue{Tag: TagWriterHandle, writerPtr: w}
}

func ValueExecFrameRef(f *KernelExecutionFrame) RuntimeValue {
	return RuntimeValue{Tag: TagExecFrameRef, frameRef: f}
}

func (v RuntimeValue) AsBool() (bool, error) {
	if v.Tag != TagBool {
		return false, typeError(TagBool, v.Tag)
	}
	return v.boolVal, nil
}

func (v RuntimeValue) AsInt64() (int64, error) {
	if !IsSignedInteger(v.Tag) {
		return 0, typeError(TagInt64, v.Tag)
	}
	if v.intVal == nil {
		return 0, fmt.Errorf("AVM: nil int value")
	}
	if !v.intVal.IsInt64() {
		return 0, fmt.Errorf("AVM: int value overflows int64")
	}
	return v.intVal.Int64(), nil
}

func (v RuntimeValue) AsUint64() (uint64, error) {
	if !IsUnsignedInteger(v.Tag) && v.Tag != TagTimestamp {
		return 0, typeError(TagUint64, v.Tag)
	}
	if v.intVal == nil {
		return 0, fmt.Errorf("AVM: nil uint value")
	}
	if !v.intVal.IsUint64() {
		return 0, fmt.Errorf("AVM: uint value overflows uint64")
	}
	return v.intVal.Uint64(), nil
}

func (v RuntimeValue) AsBigInt() (*big.Int, error) {
	if !IsInteger(v.Tag) && v.Tag != TagCoins {
		return nil, typeError(TagInt256, v.Tag)
	}
	if v.intVal == nil {
		return nil, fmt.Errorf("AVM: nil big int value")
	}
	return new(big.Int).Set(v.intVal), nil
}

func (v RuntimeValue) AsAddress() (string, error) {
	if v.Tag != TagAddress {
		return "", typeError(TagAddress, v.Tag)
	}
	return v.addrVal, nil
}

func (v RuntimeValue) AsHash() ([32]byte, error) {
	if v.Tag != TagHash {
		return [32]byte{}, typeError(TagHash, v.Tag)
	}
	return v.hashVal, nil
}

func (v RuntimeValue) AsBytes() ([]byte, error) {
	if v.Tag != TagBytes && v.Tag != TagString {
		return nil, typeError(TagBytes, v.Tag)
	}
	if v.Tag == TagBytes {
		cp := make([]byte, len(v.bytesVal))
		copy(cp, v.bytesVal)
		return cp, nil
	}
	return []byte(v.strVal), nil
}

func (v RuntimeValue) AsString() (string, error) {
	if v.Tag != TagString {
		return "", typeError(TagString, v.Tag)
	}
	return v.strVal, nil
}

func (v RuntimeValue) AsTuple() ([]RuntimeValue, error) {
	if v.Tag != TagTuple {
		return nil, typeError(TagTuple, v.Tag)
	}
	return v.tupleVal, nil
}

func (v RuntimeValue) AsChunkRef() (*chunk.Chunk, error) {
	if v.Tag != TagChunkRef {
		return nil, typeError(TagChunkRef, v.Tag)
	}
	return v.chunkRef, nil
}

func (v RuntimeValue) AsReaderCursor() (*chunk.Chunk, uint32, error) {
	if v.Tag != TagReaderCursor {
		return nil, 0, typeError(TagReaderCursor, v.Tag)
	}
	return v.chunkRef, v.readerOff, nil
}

func (v RuntimeValue) AsWriterHandle() (*ValueWriter, error) {
	if v.Tag != TagWriterHandle {
		return nil, typeError(TagWriterHandle, v.Tag)
	}
	return v.writerPtr, nil
}

func (v RuntimeValue) AsExecFrameRef() (*KernelExecutionFrame, error) {
	if v.Tag != TagExecFrameRef {
		return nil, typeError(TagExecFrameRef, v.Tag)
	}
	return v.frameRef, nil
}

func (v RuntimeValue) IsNull() bool	{ return v.Tag == TagNull }
func (v RuntimeValue) IsBool() bool	{ return v.Tag == TagBool }
func (v RuntimeValue) IsInt() bool	{ return IsSignedInteger(v.Tag) }
func (v RuntimeValue) IsUint() bool	{ return IsUnsignedInteger(v.Tag) }
func (v RuntimeValue) IsCoins() bool	{ return v.Tag == TagCoins }

func IsSignedInteger(tag ValueTag) bool {
	return tag == TagInt8 || tag == TagInt16 || tag == TagInt32 || tag == TagInt64 || tag == TagInt128 || tag == TagInt256
}

func IsUnsignedInteger(tag ValueTag) bool {
	return tag == TagUint8 || tag == TagUint16 || tag == TagUint32 || tag == TagUint64 || tag == TagUint128 || tag == TagUint256
}

// CanonicalEncode encodes a RuntimeValue into deterministic bytes.
func CanonicalEncode(v RuntimeValue) ([]byte, error) {
	buf := []byte{byte(v.Tag)}

	switch v.Tag {
	case TagNull:

	case TagBool:
		if v.boolVal {
			buf = append(buf, 0x01)
		} else {
			buf = append(buf, 0x00)
		}
	case TagInt8, TagUint8:
		buf = append(buf, encodeIntBytes(v.intVal, 1)...)
	case TagInt16, TagUint16:
		buf = append(buf, encodeIntBytes(v.intVal, 2)...)
	case TagInt32, TagUint32:
		buf = append(buf, encodeIntBytes(v.intVal, 4)...)
	case TagInt64, TagUint64:
		buf = append(buf, encodeIntBytes(v.intVal, 8)...)
	case TagInt128, TagUint128, TagCoins:
		buf = append(buf, encodeIntBytes(v.intVal, 16)...)
	case TagInt256, TagUint256:
		buf = append(buf, encodeIntBytes(v.intVal, 32)...)
	case TagTimestamp:
		buf = append(buf, encodeIntBytes(v.intVal, 8)...)
	case TagAddress:
		addrBytes := []byte(v.addrVal)
		buf = append(buf, encodeLengthPrefix(uint32(len(addrBytes)))...)
		buf = append(buf, addrBytes...)
	case TagHash:
		buf = append(buf, v.hashVal[:]...)
	case TagBytes:
		buf = append(buf, encodeLengthPrefix(uint32(len(v.bytesVal)))...)
		buf = append(buf, v.bytesVal...)
	case TagString:
		if !utf8.ValidString(v.strVal) {
			return nil, fmt.Errorf("AVM: invalid UTF-8 string")
		}
		strBytes := []byte(v.strVal)
		buf = append(buf, encodeLengthPrefix(uint32(len(strBytes)))...)
		buf = append(buf, strBytes...)
	case TagTuple:
		buf = append(buf, encodeLengthPrefix(uint32(len(v.tupleVal)))...)
		for i, elem := range v.tupleVal {
			encoded, err := CanonicalEncode(elem)
			if err != nil {
				return nil, fmt.Errorf("AVM: tuple element %d: %w", i, err)
			}
			buf = append(buf, encoded...)
		}
	case TagChunkRef:
		if v.chunkRef == nil {
			buf = append(buf, make([]byte, 32)...)
		} else {
			buf = append(buf, v.chunkRef.Hash()...)
		}
	case TagReaderCursor:
		if v.chunkRef == nil {
			buf = append(buf, make([]byte, 32)...)
		} else {
			buf = append(buf, v.chunkRef.Hash()...)
		}
		buf = append(buf, encodeIntBytes(big.NewInt(int64(v.readerOff)), 4)...)
	case TagWriterHandle:
		if v.writerPtr == nil {
			buf = append(buf, 0x00)
		} else {
			buf = append(buf, 0x01)

			buf = append(buf, make([]byte, 32)...)
		}
	case TagExecFrameRef:
		buf = append(buf, 0x00)
	default:
		return nil, fmt.Errorf("AVM: unknown value tag %d", v.Tag)
	}

	return buf, nil
}

// CanonicalDecode decodes a RuntimeValue from deterministic bytes.
func CanonicalDecode(data []byte) (RuntimeValue, int, error) {
	if len(data) < 1 {
		return RuntimeValue{}, 0, fmt.Errorf("AVM: empty data for canonical decode")
	}

	tag := ValueTag(data[0])
	offset := 1

	switch tag {
	case TagNull:
		return ValueNull(), 1, nil
	case TagBool:
		if len(data) < offset+1 {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated bool")
		}
		return ValueBool(data[offset] != 0), 2, nil
	case TagInt8, TagUint8:
		width := 1
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated int8/uint8")
		}
		v := new(big.Int).SetBytes(data[offset : offset+width])
		if IsSigned(tag) && data[offset]&0x80 != 0 {
			v.SetInt64(-1)
			v.Lsh(v, uint(width*8))
			v.Or(v, new(big.Int).SetBytes(data[offset:offset+width]))
		}
		offset += width
		return RuntimeValue{Tag: tag, intVal: v}, offset, nil
	case TagInt16, TagUint16:
		width := 2
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated int16/uint16")
		}
		v := new(big.Int).SetBytes(data[offset : offset+width])
		offset += width
		return RuntimeValue{Tag: tag, intVal: v}, offset, nil
	case TagInt32, TagUint32:
		width := 4
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated int32/uint32")
		}
		v := new(big.Int).SetBytes(data[offset : offset+width])
		offset += width
		return RuntimeValue{Tag: tag, intVal: v}, offset, nil
	case TagInt64, TagUint64, TagTimestamp:
		width := 8
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated int64/uint64/timestamp")
		}
		v := new(big.Int).SetBytes(data[offset : offset+width])
		offset += width
		return RuntimeValue{Tag: tag, intVal: v}, offset, nil
	case TagInt128, TagUint128, TagCoins:
		width := 16
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated int128/uint128/coins")
		}
		v := new(big.Int).SetBytes(data[offset : offset+width])
		offset += width
		return RuntimeValue{Tag: tag, intVal: v}, offset, nil
	case TagInt256, TagUint256:
		width := 32
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated int256/uint256")
		}
		v := new(big.Int).SetBytes(data[offset : offset+width])
		offset += width
		return RuntimeValue{Tag: tag, intVal: v}, offset, nil
	case TagAddress:
		length, n, err := decodeLengthPrefix(data[offset:])
		if err != nil {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: address length: %w", err)
		}
		offset += n
		if uint32(len(data)-offset) < length {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated address")
		}
		addr := string(data[offset : offset+int(length)])
		offset += int(length)
		return ValueAddress(addr), offset, nil
	case TagHash:
		width := 32
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated hash")
		}
		var h [32]byte
		copy(h[:], data[offset:offset+width])
		offset += width
		return ValueHash(h), offset, nil
	case TagBytes:
		length, n, err := decodeLengthPrefix(data[offset:])
		if err != nil {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: bytes length: %w", err)
		}
		offset += n
		if uint32(len(data)-offset) < length {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated bytes")
		}
		b := make([]byte, length)
		copy(b, data[offset:offset+int(length)])
		offset += int(length)
		return ValueBytes(b), offset, nil
	case TagString:
		length, n, err := decodeLengthPrefix(data[offset:])
		if err != nil {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: string length: %w", err)
		}
		offset += n
		if uint32(len(data)-offset) < length {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated string")
		}
		s := string(data[offset : offset+int(length)])
		if !utf8.ValidString(s) {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: invalid UTF-8 in canonical string")
		}
		offset += int(length)
		return ValueString(s), offset, nil
	case TagTuple:
		count, n, err := decodeLengthPrefix(data[offset:])
		if err != nil {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: tuple count: %w", err)
		}
		offset += n
		elements := make([]RuntimeValue, count)
		for i := uint32(0); i < count; i++ {
			elem, consumed, err := CanonicalDecode(data[offset:])
			if err != nil {
				return RuntimeValue{}, 0, fmt.Errorf("AVM: tuple element %d: %w", i, err)
			}
			elements[i] = elem
			offset += consumed
		}
		return ValueTuple(elements), offset, nil
	case TagChunkRef:
		width := 32
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated chunk ref")
		}

		return RuntimeValue{Tag: TagChunkRef, chunkRef: nil}, offset + width, nil
	case TagReaderCursor:
		width := 32 + 4
		if len(data) < offset+width {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated reader cursor")
		}
		off := binary.BigEndian.Uint32(data[offset+32:])
		return RuntimeValue{Tag: TagReaderCursor, chunkRef: nil, readerOff: off}, offset + width, nil
	case TagWriterHandle:
		if len(data) < offset+1 {
			return RuntimeValue{}, 0, fmt.Errorf("AVM: truncated writer handle")
		}

		offset++
		return RuntimeValue{Tag: TagWriterHandle, writerPtr: nil}, offset, nil
	case TagExecFrameRef:
		return RuntimeValue{Tag: TagExecFrameRef}, 1, nil
	default:
		return RuntimeValue{}, 0, fmt.Errorf("AVM: unknown tag %d in canonical decode", tag)
	}
}

type CastExitCode uint8

const (
	CastOK			CastExitCode	= 0
	CastTypeMismatch	CastExitCode	= 1
	CastOverflow		CastExitCode	= 2
	CastTruncation		CastExitCode	= 3
	CastInvalidUTF8		CastExitCode	= 4
	CastNullToValue		CastExitCode	= 5
)

// ExplicitCast performs a type cast between integer widths.
// All casts MUST be explicit. Invalid cast → deterministic exit code.
func ExplicitCast(v RuntimeValue, targetTag ValueTag) (RuntimeValue, CastExitCode) {
	if v.Tag == targetTag {
		return v, CastOK
	}

	if v.Tag == TagNull {
		return RuntimeValue{}, CastNullToValue
	}

	if IsInteger(v.Tag) && IsInteger(targetTag) {
		srcWidth, _ := ValueBitWidth(v.Tag)
		dstWidth, _ := ValueBitWidth(targetTag)

		if dstWidth < srcWidth {
			if !IsSigned(targetTag) && IsSigned(v.Tag) {
				return RuntimeValue{}, CastOverflow
			}
		}

		return RuntimeValue{Tag: targetTag, intVal: new(big.Int).Set(v.intVal)}, CastOK
	}

	if v.Tag == TagCoins && IsInteger(targetTag) {
		return RuntimeValue{Tag: targetTag, intVal: new(big.Int).Set(v.intVal)}, CastOK
	}

	if v.Tag == TagTimestamp && targetTag == TagUint64 {
		return RuntimeValue{Tag: targetTag, intVal: new(big.Int).Set(v.intVal)}, CastOK
	}

	return RuntimeValue{}, CastTypeMismatch
}

var maxEncodedSize = map[ValueTag]uint32{
	TagNull:		1,
	TagBool:		2,
	TagInt8:		2,
	TagInt16:		3,
	TagInt32:		5,
	TagInt64:		9,
	TagInt128:		17,
	TagInt256:		33,
	TagUint8:		2,
	TagUint16:		3,
	TagUint32:		5,
	TagUint64:		9,
	TagUint128:		17,
	TagUint256:		33,
	TagCoins:		17,
	TagTimestamp:		9,
	TagAddress:		4 + MaxAddressLength,
	TagHash:		33,
	TagBytes:		4 + MaxBytesLength,
	TagString:		4 + MaxStringLength,
	TagTuple:		5 + MaxTupleElements*33,
	TagChunkRef:		33,
	TagReaderCursor:	37,
	TagWriterHandle:	34,
	TagExecFrameRef:	2,
}

var gasCostEncode = map[ValueTag]uint64{
	TagNull:		1,
	TagBool:		1,
	TagInt8:		2,
	TagInt16:		2,
	TagInt32:		3,
	TagInt64:		3,
	TagInt128:		5,
	TagInt256:		8,
	TagUint8:		2,
	TagUint16:		2,
	TagUint32:		3,
	TagUint64:		3,
	TagUint128:		5,
	TagUint256:		8,
	TagCoins:		5,
	TagTimestamp:		3,
	TagAddress:		10,
	TagHash:		10,
	TagBytes:		5 + 1,
	TagString:		5 + 1,
	TagTuple:		10,
	TagChunkRef:		15,
	TagReaderCursor:	20,
	TagWriterHandle:	20,
	TagExecFrameRef:	5,
}

var gasCostDecode = map[ValueTag]uint64{
	TagNull:		1,
	TagBool:		1,
	TagInt8:		2,
	TagInt16:		2,
	TagInt32:		3,
	TagInt64:		3,
	TagInt128:		5,
	TagInt256:		8,
	TagUint8:		2,
	TagUint16:		2,
	TagUint32:		3,
	TagUint64:		3,
	TagUint128:		5,
	TagUint256:		8,
	TagCoins:		5,
	TagTimestamp:		3,
	TagAddress:		10,
	TagHash:		10,
	TagBytes:		5 + 1,
	TagString:		5 + 1,
	TagTuple:		10,
	TagChunkRef:		15,
	TagReaderCursor:	20,
	TagWriterHandle:	20,
	TagExecFrameRef:	5,
}

// Size bounds for variable-length types
const (
	MaxAddressLength	uint32	= 128
	MaxBytesLength		uint32	= 65536
	MaxStringLength		uint32	= 65536
	MaxTupleElements	uint32	= 256
)

// MaxEncodedSize returns the maximum encoded size for a value tag.
func MaxEncodedSizeForTag(tag ValueTag) uint32 {
	if s, ok := maxEncodedSize[tag]; ok {
		return s
	}
	return 0
}

// GasCostEncode returns the gas cost to encode a value of the given tag.
func GasCostEncode(tag ValueTag) uint64 {
	if c, ok := gasCostEncode[tag]; ok {
		return c
	}
	return 100
}

// GasCostDecode returns the gas cost to decode a value of the given tag.
func GasCostDecode(tag ValueTag) uint64 {
	if c, ok := gasCostDecode[tag]; ok {
		return c
	}
	return 100
}

// CanonicalHash returns the BLAKE3 hash of the canonical encoding of a value.
func CanonicalHash(v RuntimeValue) ([32]byte, error) {
	encoded, err := CanonicalEncode(v)
	if err != nil {
		return [32]byte{}, fmt.Errorf("AVM: canonical hash: %w", err)
	}
	return blake3.Sum256(encoded), nil
}

// OptionNone returns the null value representing Option.None.
func OptionNone() RuntimeValue {
	return ValueNull()
}

// OptionSome wraps a value representing Option.Some.
func OptionSome(v RuntimeValue) RuntimeValue {
	return v
}

// IsOptionNone checks if a value is Option.None (null).
func IsOptionNone(v RuntimeValue) bool {
	return v.Tag == TagNull
}

type ValueWriter struct {
	builder	*chunk.Builder
	data	[]byte
	refs	[]*chunk.Chunk
}

func NewValueWriter() *ValueWriter {
	return &ValueWriter{
		builder: chunk.NewBuilder(),
	}
}

func (w *ValueWriter) Builder() *chunk.Builder {
	return w.builder
}

func (w *ValueWriter) WriteValue(v RuntimeValue) error {
	encoded, err := CanonicalEncode(v)
	if err != nil {
		return err
	}
	w.data = append(w.data, encoded...)
	return nil
}

func (w *ValueWriter) Build() (*chunk.Chunk, error) {
	w.builder.SetData(w.data, uint16(len(w.data)*8))
	w.builder.SetTypeTag(chunk.TypeNormal)
	for _, ref := range w.refs {
		w.builder.AddRef(ref)
	}
	return w.builder.Build()
}

func (w *ValueWriter) AddRef(c *chunk.Chunk) {
	w.refs = append(w.refs, c)
}

func ValidateDeterminism(a, b RuntimeValue) error {
	encA, err := CanonicalEncode(a)
	if err != nil {
		return fmt.Errorf("AVM: determinism check A: %w", err)
	}
	encB, err := CanonicalEncode(b)
	if err != nil {
		return fmt.Errorf("AVM: determinism check B: %w", err)
	}
	if a.Tag != b.Tag {
		return fmt.Errorf("AVM: determinism violation: different tags %v vs %v", a.Tag, b.Tag)
	}
	if len(encA) != len(encB) {
		return fmt.Errorf("AVM: determinism violation: different encoded lengths %d vs %d", len(encA), len(encB))
	}
	for i := range encA {
		if encA[i] != encB[i] {
			return fmt.Errorf("AVM: determinism violation: byte %d differs (0x%02x vs 0x%02x)", i, encA[i], encB[i])
		}
	}
	return nil
}

func typeError(expected, got ValueTag) error {
	return fmt.Errorf("AVM type error: expected %s, got %s → EXIT_TYPE_ERROR", expected, got)
}

func encodeIntBytes(v *big.Int, width int) []byte {
	if v == nil {
		return make([]byte, width)
	}
	b := make([]byte, width)
	v.FillBytes(b)
	return b
}

func encodeLengthPrefix(length uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, length)
	return b
}

func decodeLengthPrefix(data []byte) (uint32, int, error) {
	if len(data) < 4 {
		return 0, 0, fmt.Errorf("AVM: truncated length prefix")
	}
	length := binary.BigEndian.Uint32(data[:4])
	return length, 4, nil
}
