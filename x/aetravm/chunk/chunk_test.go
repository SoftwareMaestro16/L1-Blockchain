package chunk

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkBuilder(t *testing.T) {

	leaf, err := NewBuilder().
		SetTypeTag(TypeNormal).
		SetData([]byte{0xDE, 0xAD, 0xBE, 0xEF}, 32).
		Build()
	require.NoError(t, err)
	require.Equal(t, uint8(0), leaf.Level())
	require.Equal(t, uint16(32), leaf.BitCount())
	require.Equal(t, 2, len(leaf.hashes))

	h0 := hex.EncodeToString(leaf.Hash())
	h1 := hex.EncodeToString(leaf.HashLayer(1))
	require.Equal(t, h0, h1, "leaf H0 and H1 must be equal")

	parent, err := NewBuilder().
		SetTypeTag(TypeNormal).
		SetRef(0, leaf).
		SetData([]byte{0x01}, 8).
		Build()
	require.NoError(t, err)
	require.Equal(t, uint8(1), parent.Level())
	require.Equal(t, 1, len(parent.Refs()))
	require.Equal(t, leaf.Hash(), parent.RefAt(0).Hash())

	ph0 := hex.EncodeToString(parent.Hash())
	ph1 := hex.EncodeToString(parent.HashLayer(1))
	require.Equal(t, ph0, ph1, "parent H0 and H1 must be equal if children H0 and H1 are equal")
}

func TestChunkLimits(t *testing.T) {

	_, err := NewBuilder().SetData(make([]byte, 300), 2400).Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds limit 2048")

	builder := NewBuilder()
	leaf, _ := builder.SetData([]byte{0}, 8).Build()
	builder = NewBuilder()
	for i := 0; i < 8; i++ {
		builder.AddRef(leaf)
	}

	builder.AddRef(leaf)
	require.Equal(t, uint8(0xff), builder.refBitmap)
}

func TestChunkImmutabilityAndDeterminism(t *testing.T) {
	data := []byte{0xAA, 0xBB}
	c1, err := NewBuilder().SetData(data, 16).Build()
	require.NoError(t, err)

	c2, err := NewBuilder().SetData(data, 16).Build()
	require.NoError(t, err)

	require.Equal(t, c1.Hash(), c2.Hash(), "identical chunks must have identical hashes")
}

func TestChunkLevel(t *testing.T) {
	c0, _ := NewBuilder().SetData([]byte{0}, 8).Build()
	require.Equal(t, uint8(0), c0.Level())

	c1, _ := NewBuilder().SetRef(0, c0).Build()
	require.Equal(t, uint8(1), c1.Level())

	c2, _ := NewBuilder().SetRef(4, c1).Build()
	require.Equal(t, uint8(2), c2.Level())

	c3, _ := NewBuilder().SetRef(1, c0).SetRef(7, c2).Build()
	require.Equal(t, uint8(3), c3.Level())
}

func TestChunkExportImport(t *testing.T) {
	leaf, _ := NewBuilder().SetData([]byte{0xDE, 0xAD}, 16).Build()
	root, _ := NewBuilder().SetRef(0, leaf).SetData([]byte{0x01}, 8).Build()

	originalHash := root.Hash()

	exportedHashes := [][]byte{root.HashLayer(0), root.HashLayer(1)}
	exportedData := root.Data()
	exportedBitCount := root.BitCount()

	childPruned, _ := NewPrunedChunk(leaf.Level(), [][]byte{leaf.HashLayer(0), leaf.HashLayer(1)})

	rebuilt, err := NewBuilder().
		SetData(exportedData, exportedBitCount).
		SetRef(0, childPruned).
		Build()
	require.NoError(t, err)

	require.Equal(t, originalHash, rebuilt.Hash())
	require.Equal(t, exportedHashes[0], rebuilt.Hash())
}
