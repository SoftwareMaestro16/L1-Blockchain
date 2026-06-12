package types

import (
	"fmt"
	"unicode/utf8"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

// Codec defines how a value of type T is encoded and decoded.
type Codec interface {
	Encode(w *Writer, v Value) error
	Decode(r *Reader) (Value, error)
	GasCost() uint64
	MaxEncodedSize() uint32
	SchemaDescriptor() string
}

type Uint32Codec struct{}

func (c Uint32Codec) Encode(w *Writer, v Value) error {
	val, ok := v.Payload.(uint32)
	if !ok {
		return fmt.Errorf("invalid payload for uint32: %T", v.Payload)
	}
	return w.WriteUint32(val)
}

func (c Uint32Codec) Decode(r *Reader) (Value, error) {
	val, err := r.ReadUint32()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: TypeUint32, Payload: val}, nil
}

func (c Uint32Codec) GasCost() uint64		{ return 10 }
func (c Uint32Codec) MaxEncodedSize() uint32	{ return 4 }
func (c Uint32Codec) SchemaDescriptor() string	{ return "uint32" }

type StringCodec struct {
	MaxLength uint32
}

func (c StringCodec) Encode(w *Writer, v Value) error {
	s, ok := v.Payload.(string)
	if !ok {
		return fmt.Errorf("invalid payload for string: %T", v.Payload)
	}
	if uint32(len(s)) > c.MaxLength {
		return fmt.Errorf("string length %d exceeds max %d", len(s), c.MaxLength)
	}
	if !utf8.ValidString(s) {
		return fmt.Errorf("invalid UTF-8 string")
	}

	if err := w.WriteUint32(uint32(len(s))); err != nil {
		return err
	}
	return w.WriteBits([]byte(s), len(s)*8)
}

func (c StringCodec) Decode(r *Reader) (Value, error) {
	length, err := r.ReadUint32()
	if err != nil {
		return Value{}, err
	}
	if length > c.MaxLength {
		return Value{}, fmt.Errorf("decoded string length %d exceeds max %d", length, c.MaxLength)
	}

	bits, err := r.ReadBits(int(length) * 8)
	if err != nil {
		return Value{}, err
	}
	s := string(bits)
	if !utf8.ValidString(s) {
		return Value{}, fmt.Errorf("decoded invalid UTF-8 string")
	}

	return Value{
		Type:		&StringType{BaseType: BaseType{kind: KindString}, MaxLength: c.MaxLength},
		Payload:	s,
	}, nil
}

func (c StringCodec) GasCost() uint64		{ return 20 }
func (c StringCodec) MaxEncodedSize() uint32	{ return 4 + c.MaxLength }
func (c StringCodec) SchemaDescriptor() string	{ return fmt.Sprintf("string(%d)", c.MaxLength) }

// ChunkCodec handles lazy decoding of chunks.
type ChunkCodec struct{}

func (c ChunkCodec) Encode(w *Writer, v Value) error {
	ch, ok := v.Payload.(*chunk.Chunk)
	if !ok {
		return fmt.Errorf("invalid payload for chunk: %T", v.Payload)
	}
	return w.AddRef(ch)
}

func (c ChunkCodec) Decode(r *Reader) (Value, error) {
	ch, err := r.ReadRef()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: TypeChunk, Payload: ch}, nil
}

func (c ChunkCodec) GasCost() uint64		{ return 50 }
func (c ChunkCodec) MaxEncodedSize() uint32	{ return 32 }
func (c ChunkCodec) SchemaDescriptor() string	{ return "chunk" }
