package types

import (
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	"github.com/stretchr/testify/require"
)

func TestUint32Codec(t *testing.T) {
	codec := Uint32Codec{}
	val := Value{Type: TypeUint32, Payload: uint32(0x12345678)}

	w := NewWriter()
	err := codec.Encode(w, val)
	require.NoError(t, err)

	c, err := w.Build()
	require.NoError(t, err)
	require.Equal(t, uint16(32), c.BitCount())

	r := NewReader(c)
	decoded, err := codec.Decode(r)
	require.NoError(t, err)
	require.Equal(t, val.Payload, decoded.Payload)
}

func TestStringCodec(t *testing.T) {
	codec := StringCodec{MaxLength: 100}
	val := Value{Type: &StringType{MaxLength: 100}, Payload: "Hello AVM"}

	w := NewWriter()
	err := codec.Encode(w, val)
	require.NoError(t, err)

	c, err := w.Build()
	require.NoError(t, err)

	r := NewReader(c)
	decoded, err := codec.Decode(r)
	require.NoError(t, err)
	require.Equal(t, val.Payload, decoded.Payload)

	invalidVal := Value{Type: &StringType{MaxLength: 100}, Payload: string([]byte{0xff, 0xfe, 0xfd})}
	err = codec.Encode(w, invalidVal)
	require.Error(t, err)
}

func TestChunkCodec(t *testing.T) {
	codec := ChunkCodec{}
	leaf, _ := chunk.NewBuilder().SetData([]byte{1, 2, 3}, 24).Build()
	val := Value{Type: TypeChunk, Payload: leaf}

	w := NewWriter()
	err := codec.Encode(w, val)
	require.NoError(t, err)

	c, err := w.Build()
	require.NoError(t, err)
	require.Equal(t, 1, len(c.Refs()))

	r := NewReader(c)
	decoded, err := codec.Decode(r)
	require.NoError(t, err)

	decodedChunk := decoded.Payload.(*chunk.Chunk)
	require.Equal(t, leaf.Hash(), decodedChunk.Hash())
}

func TestTypeLattice(t *testing.T) {
	require.True(t, TypeUint32.IsAssignableFrom(TypeUint32))
	require.False(t, TypeUint32.IsAssignableFrom(TypeInt32))

	require.True(t, TypeCoins.IsAssignableFrom(TypeUint128))
}
