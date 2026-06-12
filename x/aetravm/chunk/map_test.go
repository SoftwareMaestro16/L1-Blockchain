package chunk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkMapBasic(t *testing.T) {
	m := NewMap(nil)

	key1 := []byte("alice_balance")
	val1, _ := NewBuilder().SetData([]byte{0x00, 0x64}, 16).Build()

	key2 := []byte("bob_balance")
	val2, _ := NewBuilder().SetData([]byte{0x00, 0xC8}, 16).Build()

	m1, err := m.Put(key1, val1)
	require.NoError(t, err)
	require.NotNil(t, m1.Root())

	m2, err := m1.Put(key2, val2)
	require.NoError(t, err)

	res1, err := m2.Get(key1)
	require.NoError(t, err)
	require.Equal(t, val1.Hash(), res1.Hash())

	res2, err := m2.Get(key2)
	require.NoError(t, err)
	require.Equal(t, val2.Hash(), res2.Hash())

	val1Updated, _ := NewBuilder().SetData([]byte{0x00, 0x96}, 16).Build()
	m3, err := m2.Put(key1, val1Updated)
	require.NoError(t, err)

	res1Updated, _ := m3.Get(key1)
	require.Equal(t, val1Updated.Hash(), res1Updated.Hash())

	m4, err := m3.Delete(key1)
	require.NoError(t, err)
	res1Deleted, _ := m4.Get(key1)
	require.Nil(t, res1Deleted)

	res2StillThere, _ := m4.Get(key2)
	require.NotNil(t, res2StillThere)
	require.Equal(t, val2.Hash(), res2StillThere.Hash())
}

func TestChunkMapPersistence(t *testing.T) {
	m := NewMap(nil)
	key := []byte("persistent")
	val, _ := NewBuilder().SetData([]byte{1}, 8).Build()

	m1, _ := m.Put(key, val)
	root1 := m1.Root().Hash()

	m2, _ := m1.Put(key, val)
	root2 := m2.Root().Hash()
	require.Equal(t, root1, root2, "same key+value must produce identical root hash")

	m3, _ := m1.Put([]byte("other"), val)
	root3 := m3.Root().Hash()
	require.NotEqual(t, root1, root3, "different key must produce different root hash")
}

func TestChunkMapProof(t *testing.T) {
	m := NewMap(nil)
	key1 := []byte("key1")
	val1, _ := NewBuilder().SetData([]byte{1}, 8).Build()
	key2 := []byte("key2")
	val2, _ := NewBuilder().SetData([]byte{2}, 8).Build()

	m, _ = m.Put(key1, val1)
	m, _ = m.Put(key2, val2)

	proof, err := m.Prove(key1)
	require.NoError(t, err)
	require.Equal(t, m.Root().Hash(), proof.Hash(), "proof root hash must match actual root hash")

	pm := NewMap(proof)
	res1, _ := pm.Get(key1)
	require.NotNil(t, res1)
	require.Equal(t, val1.Hash(), res1.Hash())

	res2, _ := pm.Get(key2)
	if res2 != nil {
		require.Equal(t, TypePruned, res2.TypeTag(), "key2 should be pruned in key1's proof")
	}
}

func TestChunkMapValidate(t *testing.T) {
	m := NewEmptyMap()

	require.NoError(t, m.Validate())

	key1 := []byte("k1")
	val, _ := NewBuilder().SetData([]byte{1}, 8).Build()
	m, _ = m.Put(key1, val)
	require.NoError(t, m.Validate())

	key2 := []byte("k2")
	m, _ = m.Put(key2, val)
	require.NoError(t, m.Validate())
}

func TestChunkMapValidateEmptyBranch(t *testing.T) {
	m := NewEmptyMap()
	m, _ = m.Put([]byte("a"), mustBuild(t, []byte{1}))
	m, _ = m.Put([]byte("b"), mustBuild(t, []byte{2}))
	m, _ = m.Put([]byte("c"), mustBuild(t, []byte{3}))

	require.NoError(t, m.Validate())

	m, _ = m.Delete([]byte("a"))
	m, _ = m.Delete([]byte("b"))
	m, _ = m.Delete([]byte("c"))
	require.NoError(t, m.Validate())
}

func TestChunkMapMaxDepth(t *testing.T) {
	m := NewEmptyMap()
	val, _ := NewBuilder().SetData([]byte{1}, 8).Build()

	for i := 0; i < 50; i++ {
		key := []byte{byte(i)}
		var err error
		m, err = m.Put(key, val)
		require.NoError(t, err, "put at key %d failed", i)
	}

	require.NoError(t, m.Validate())
}

func TestChunkMapCompress(t *testing.T) {
	m := NewEmptyMap()
	val, _ := NewBuilder().SetData([]byte{0x42}, 8).Build()

	m, _ = m.Put([]byte{0x01}, val)

	beforeDepth := getDepth(m.Root())
	require.Greater(t, beforeDepth, 0)

	compressed := m.Compress()

	require.NoError(t, compressed.Validate())

	afterDepth := getDepth(compressed.Root())
	require.LessOrEqual(t, afterDepth, beforeDepth)

	res, err := compressed.Get([]byte{0x01})
	require.NoError(t, err)
	require.NotNil(t, res, "key must be retrievable after compress")
	require.Equal(t, val.Hash(), res.Hash())
}

func TestChunkMapCompressPreservesLookup(t *testing.T) {
	m := NewEmptyMap()
	keys := [][]byte{[]byte("alpha"), []byte("beta"), []byte("gamma")}
	vals := make([]*Chunk, 3)

	for i, k := range keys {
		v, _ := NewBuilder().SetData([]byte{byte(i + 1)}, 8).Build()
		vals[i] = v
		m, _ = m.Put(k, v)
	}

	compressed := m.Compress()

	for i, k := range keys {
		res, err := compressed.Get(k)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, vals[i].Hash(), res.Hash(), "key %s mismatch after compress", k)
	}
}

func TestChunkMapProveAbsence(t *testing.T) {
	m := NewEmptyMap()
	val, _ := NewBuilder().SetData([]byte{1}, 8).Build()

	m, _ = m.Put([]byte("present"), val)

	proof, err := m.ProveAbsence([]byte("absent"))
	require.NoError(t, err)
	require.NotNil(t, proof, "absence proof should exist even for empty target")

	require.Equal(t, m.Root().Hash(), proof.Hash())
}

func TestChunkMapProveAbsenceEmpty(t *testing.T) {
	m := NewEmptyMap()

	proof, err := m.ProveAbsence([]byte("any"))
	require.NoError(t, err)
	require.NotNil(t, proof)
}

func TestChunkMapVersioning(t *testing.T) {
	m := NewEmptyMap()
	require.Equal(t, uint64(0), m.Version())

	val, _ := NewBuilder().SetData([]byte{1}, 8).Build()

	m, _ = m.Put([]byte("k1"), val)
	require.Equal(t, uint64(1), m.Version())

	m, _ = m.Put([]byte("k2"), val)
	require.Equal(t, uint64(2), m.Version())

	m, _ = m.Delete([]byte("k1"))
	require.Equal(t, uint64(3), m.Version())

	m = m.Compress()
	require.Equal(t, uint64(4), m.Version())
}

func TestChunkMapRootHash(t *testing.T) {
	m := NewEmptyMap()

	emptyHash := m.RootHash()
	require.NotEqual(t, [HashSize]byte{}, emptyHash)

	val, _ := NewBuilder().SetData([]byte{0x01}, 8).Build()
	m, _ = m.Put([]byte("key"), val)

	nonEmptyHash := m.RootHash()
	require.NotEqual(t, emptyHash, nonEmptyHash)

	m2 := NewEmptyMap()
	m2, _ = m2.Put([]byte("key"), val)
	require.Equal(t, nonEmptyHash, m2.RootHash())
}

func TestChunkMapGasModel(t *testing.T) {

	gcShallow := GasCostFor(1)
	gcDeep := GasCostFor(50)

	require.Greater(t, gcDeep.Lookup, gcShallow.Lookup)
	require.Greater(t, gcDeep.Insert, gcShallow.Insert)
	require.Greater(t, gcDeep.Delete, gcShallow.Delete)
	require.Greater(t, gcDeep.Proof, gcShallow.Proof)

	gc0 := GasCostFor(0)
	require.Greater(t, gc0.Lookup, uint64(0))
	require.Greater(t, gc0.Insert, uint64(0))
	require.Greater(t, gc0.Delete, uint64(0))
	require.Greater(t, gc0.Proof, uint64(0))
}

func TestChunkMapIterate(t *testing.T) {
	m := NewEmptyMap()
	v1, _ := NewBuilder().SetData([]byte{1}, 8).Build()
	v2, _ := NewBuilder().SetData([]byte{2}, 8).Build()
	v3, _ := NewBuilder().SetData([]byte{3}, 8).Build()

	m, _ = m.Put([]byte("ccc"), v3)
	m, _ = m.Put([]byte("aaa"), v1)
	m, _ = m.Put([]byte("bbb"), v2)

	entries := m.Iterate()
	require.Equal(t, 3, len(entries))

	vals := make(map[[HashSize]byte]bool)
	for _, e := range entries {
		var h [HashSize]byte
		copy(h[:], e.Value.Hash())
		vals[h] = true
	}
	var h1, h2, h3 [HashSize]byte
	copy(h1[:], v1.Hash())
	copy(h2[:], v2.Hash())
	copy(h3[:], v3.Hash())
	require.True(t, vals[h1], "v1 missing")
	require.True(t, vals[h2], "v2 missing")
	require.True(t, vals[h3], "v3 missing")
}

func TestChunkMapIterateEmpty(t *testing.T) {
	m := NewEmptyMap()
	entries := m.Iterate()
	require.Nil(t, entries)
}

func TestChunkMapEmptyCanonical(t *testing.T) {
	m := NewEmptyMap()
	require.True(t, m.IsEmpty())
	require.Nil(t, m.Root())
	require.NoError(t, m.Validate())
}

func TestChunkMapCollisionBucket(t *testing.T) {

	m := NewEmptyMap()
	v1, _ := NewBuilder().SetData([]byte{0x0A}, 8).Build()
	v2, _ := NewBuilder().SetData([]byte{0x0B}, 8).Build()

	keys := make([][]byte, 20)
	for i := 0; i < 20; i++ {
		keys[i] = []byte{byte(i), 0x00, 0x00, 0x00}
	}
	m, _ = m.Put(keys[0], v1)
	m, _ = m.Put(keys[1], v2)

	r1, err := m.Get(keys[0])
	require.NoError(t, err)
	require.NotNil(t, r1)
	require.Equal(t, v1.Hash(), r1.Hash())

	r2, err := m.Get(keys[1])
	require.NoError(t, err)
	require.NotNil(t, r2)
	require.Equal(t, v2.Hash(), r2.Hash())
}

func TestChunkMapManyKeysDeterministic(t *testing.T) {
	m := NewEmptyMap()
	val, _ := NewBuilder().SetData([]byte{0xFF}, 8).Build()

	for i := 0; i < 100; i++ {
		key := []byte{byte(i), byte(i >> 8)}
		var err error
		m, err = m.Put(key, val)
		require.NoError(t, err)
	}

	m2 := NewEmptyMap()
	for i := 0; i < 100; i++ {
		key := []byte{byte(i), byte(i >> 8)}
		m2, _ = m2.Put(key, val)
	}

	require.Equal(t, m.RootHash(), m2.RootHash())

	require.NoError(t, m.Validate())
}

func TestChunkMapIsEmpty(t *testing.T) {
	require.True(t, NewEmptyMap().IsEmpty())
	require.False(t, NewMap(mustBuild(t, []byte{1})).IsEmpty())
}

func TestChunkMapParallelBucketsIndependent(t *testing.T) {
	m := NewEmptyMap()
	val, _ := NewBuilder().SetData([]byte{0x01}, 8).Build()

	m, _ = m.Put([]byte{0x00}, val)

	m2 := NewEmptyMap()
	m2, _ = m2.Put([]byte{0x20}, val)

	require.NoError(t, m.Validate())
	require.NoError(t, m2.Validate())
	require.NotEqual(t, m.RootHash(), m2.RootHash())
}

func mustBuild(t *testing.T, data []byte) *Chunk {
	t.Helper()
	c, err := NewBuilder().SetData(data, uint16(len(data)*8)).Build()
	require.NoError(t, err)
	return c
}
