package chunk

import (
	"errors"
	"fmt"

	"lukechampine.com/blake3"
)

const (
	maxTrieDepth	= 86	// ceil(256/3)
	gasBaseLookup	= 100
	gasBaseInsert	= 200
	gasBaseDelete	= 150
	gasBaseProof	= 500
	gasPerDepthStep	= 10
	gasPerChunkByte	= 1
)

var (
	ErrMaxDepthExceeded	= fmt.Errorf("chunk map: max trie depth %d exceeded", maxTrieDepth)
	ErrInvalidMap		= errors.New("chunk map: invalid structure")
	ErrEmptyBranch		= errors.New("chunk map: non-canonical empty branch")
)

// GasCost holds the gas cost estimates for Map operations.
type GasCost struct {
	Lookup	uint64
	Insert	uint64
	Delete	uint64
	Proof	uint64
}

// GasCostFor returns gas cost estimates based on trie depth.
// Complexity: O(depth) for all operations.
func GasCostFor(depth int) GasCost {
	d := uint64(depth)
	return GasCost{
		Lookup:	gasBaseLookup + d*gasPerDepthStep,
		Insert:	gasBaseInsert + d*gasPerDepthStep,
		Delete:	gasBaseDelete + d*gasPerDepthStep,
		Proof:	gasBaseProof + d*gasPerDepthStep*gasPerChunkByte,
	}
}

// Entry represents a single key-value pair in iteration output.
type Entry struct {
	Key	[]byte
	Value	*Chunk
}

// Map represents a persistent, immutable Merkle Trie with 8-fanout.
// The root Chunk is a cryptographic commitment to the entire map state.
// Every modification (Put, Delete, Compress) produces a NEW Map — no in-place mutation.
//
// Security invariants:
//  1. Each key maps to exactly one leaf Chunk. No duplicate paths.
//  2. Internal nodes are a deterministic function of their children.
//  3. Empty branches are canonicalized (nil represents "no branch").
//  4. Root Hash = BLAKE3(canonical(root)) — any modification propagates to root deterministically.
//  5. Max depth = ceil(256/3) = 86. Deeper trees are rejected.
//  6. Updates in disjoint top-level buckets (different first 3 bits) have no write conflicts.
type Map struct {
	root	*Chunk
	version	uint64
}

// NewEmptyMap returns a Map with a canonical empty root.
func NewEmptyMap() *Map {
	return &Map{root: nil, version: 0}
}

// NewMap wraps an existing root Chunk into a Map.
func NewMap(root *Chunk) *Map {
	return &Map{root: root, version: 0}
}

// IsEmpty returns true if the map has no entries.
func (m *Map) IsEmpty() bool {
	return m.root == nil
}

// Root returns the root Chunk.
func (m *Map) Root() *Chunk {
	return m.root
}

// Version returns the monotonic version counter incremented on every mutation.
func (m *Map) Version() uint64 {
	return m.version
}

// RootHash returns the cryptographic commitment for the map state.
// root_hash = BLAKE3(canonical(root_chunk_encoding))
func (m *Map) RootHash() [HashSize]byte {
	if m.root == nil {
		return blake3.Sum256(nil)
	}
	var h [HashSize]byte
	copy(h[:], m.root.Hash())
	return h
}

// Validate checks structural invariants:
//  1. Depth does not exceed maxTrieDepth
//  2. Empty branches use nil (not placeholder Chunks)
func (m *Map) Validate() error {
	if m.root == nil {
		return nil
	}
	return m.validateRecursive(m.root, 0)
}

func (m *Map) validateRecursive(node *Chunk, depth int) error {
	if node == nil {
		return nil
	}
	if depth > maxTrieDepth {
		return fmt.Errorf("%w at depth %d", ErrMaxDepthExceeded, depth)
	}

	for i := 0; i < MaxRefs; i++ {
		child := node.RefAt(i)
		if child != nil {
			if child.TypeTag() == TypeNormal && child.BitCount() == 0 {
				hasChildren := false
				for j := 0; j < MaxRefs; j++ {
					if child.RefAt(j) != nil {
						hasChildren = true
						break
					}
				}
				if !hasChildren {
					return fmt.Errorf("%w: empty non-canonical branch at depth %d ref %d", ErrEmptyBranch, depth, i)
				}
			}
			if err := m.validateRecursive(child, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// getDepth returns the depth of the trie at a given node.
func getDepth(node *Chunk) int {
	if node == nil {
		return 0
	}
	maxChild := 0
	for i := 0; i < MaxRefs; i++ {
		child := node.RefAt(i)
		if child != nil {
			d := getDepth(child)
			if d > maxChild {
				maxChild = d
			}
		}
	}
	return maxChild + 1
}

// Get retrieves the value Chunk for the given key.
func (m *Map) Get(key []byte) (*Chunk, error) {
	if m.root == nil {
		return nil, nil
	}
	hash := HashKey(key)
	return m.getRecursive(m.root, hash[:], 0)
}

func (m *Map) getRecursive(node *Chunk, hash []byte, bitOffset int) (*Chunk, error) {
	if node == nil {
		return nil, nil
	}
	if bitOffset >= 256 {
		return node, nil
	}

	index := getIndex(hash, bitOffset)
	child := node.RefAt(index)
	if child == nil {
		return nil, nil
	}

	return m.getRecursive(child, hash, bitOffset+3)
}

// Put inserts or updates a key-value pair, returning a new Map with the updated root.
func (m *Map) Put(key []byte, value *Chunk) (*Map, error) {
	hash := HashKey(key)
	newRoot, err := m.putRecursive(m.root, hash[:], 0, value, 0, key)
	if err != nil {
		return nil, err
	}
	return &Map{root: newRoot, version: m.version + 1}, nil
}

func (m *Map) putRecursive(node *Chunk, hash []byte, bitOffset int, value *Chunk, depth int, fullKey []byte) (*Chunk, error) {
	if depth > maxTrieDepth {
		return nil, ErrMaxDepthExceeded
	}
	if bitOffset >= 256 {
		return value, nil
	}

	index := getIndex(hash, bitOffset)

	builder := NewBuilder().SetTypeTag(TypeNormal)
	var childToUpdate *Chunk
	if node != nil {
		builder.SetData(node.Data(), node.BitCount())
		for i := 0; i < MaxRefs; i++ {
			builder.SetRef(i, node.RefAt(i))
		}
		childToUpdate = node.RefAt(index)
	}

	newChild, err := m.putRecursive(childToUpdate, hash, bitOffset+3, value, depth+1, fullKey)
	if err != nil {
		return nil, err
	}
	builder.SetRef(index, newChild)

	return builder.Build()
}

// Delete removes a key, returning a new Map.
func (m *Map) Delete(key []byte) (*Map, error) {
	if m.root == nil {
		return &Map{root: nil, version: m.version}, nil
	}
	hash := HashKey(key)
	newRoot, err := m.deleteRecursive(m.root, hash[:], 0)
	if err != nil {
		return nil, err
	}
	return &Map{root: newRoot, version: m.version + 1}, nil
}

func (m *Map) deleteRecursive(node *Chunk, hash []byte, bitOffset int) (*Chunk, error) {
	if node == nil {
		return nil, nil
	}
	if bitOffset >= 256 {
		return nil, nil
	}

	index := getIndex(hash, bitOffset)
	child := node.RefAt(index)
	if child == nil {
		return node, nil
	}

	newChild, err := m.deleteRecursive(child, hash, bitOffset+3)
	if err != nil {
		return nil, err
	}

	hasOtherChildren := false
	for i := 0; i < MaxRefs; i++ {
		if i == index {
			if newChild != nil {
				hasOtherChildren = true
			}
		} else {
			if node.RefAt(i) != nil {
				hasOtherChildren = true
			}
		}
	}

	if !hasOtherChildren && node.BitCount() == 0 {
		return nil, nil
	}

	builder := NewBuilder().SetTypeTag(TypeNormal).SetData(node.Data(), node.BitCount())
	for i := 0; i < MaxRefs; i++ {
		if i == index {
			builder.SetRef(i, newChild)
		} else {
			builder.SetRef(i, node.RefAt(i))
		}
	}
	return builder.Build()
}

// Prove returns a pruned Chunk tree serving as a Merkle inclusion proof for the key.
// The proof root hash MUST equal the map's RootHash().
func (m *Map) Prove(key []byte) (*Chunk, error) {
	if m.root == nil {
		return nil, nil
	}
	hash := HashKey(key)
	return m.proveRecursive(m.root, hash[:], 0, false)
}

// ProveAbsence returns a pruned Chunk tree proving that a key does NOT exist.
func (m *Map) ProveAbsence(key []byte) (*Chunk, error) {
	if m.root == nil {
		return NewEmptyChunk(), nil
	}
	hash := HashKey(key)
	return m.proveRecursive(m.root, hash[:], 0, true)
}

func (m *Map) proveRecursive(node *Chunk, hash []byte, bitOffset int, absence bool) (*Chunk, error) {
	if node == nil {
		return nil, nil
	}
	if node.TypeTag() == TypePruned {
		return node, nil
	}
	if bitOffset >= 256 {
		if absence {
			return nil, nil
		}
		return node, nil
	}

	index := getIndex(hash, bitOffset)
	builder := NewBuilder().
		SetTypeTag(node.TypeTag()).
		SetData(node.Data(), node.BitCount())

	for i := 0; i < MaxRefs; i++ {
		child := node.RefAt(i)
		if child == nil {
			continue
		}
		if i == index {
			provedChild, err := m.proveRecursive(child, hash, bitOffset+3, absence)
			if err != nil {
				return nil, err
			}
			if provedChild != nil {
				builder.SetRef(i, provedChild)
			}
		} else {
			pruned, _ := NewPrunedChunk(child.Level(), [][]byte{child.HashLayer(0), child.HashLayer(1)})
			builder.SetRef(i, pruned)
		}
	}

	return builder.Build()
}

// Compress removes empty sub-branches from the trie, reducing unnecessary depth.
func (m *Map) Compress() *Map {
	if m.root == nil {
		return &Map{root: nil, version: m.version}
	}
	newRoot := m.compressRecursive(m.root)
	return &Map{root: newRoot, version: m.version + 1}
}

func (m *Map) compressRecursive(node *Chunk) *Chunk {
	if node == nil {
		return nil
	}

	children := make([]*Chunk, MaxRefs)
	childCount := 0
	for i := 0; i < MaxRefs; i++ {
		child := m.compressRecursive(node.RefAt(i))
		children[i] = child
		if child != nil {
			childCount++
		}
	}

	if childCount == 0 && node.BitCount() == 0 {
		return nil
	}

	builder := NewBuilder().SetTypeTag(TypeNormal).SetData(node.Data(), node.BitCount())
	for i := 0; i < MaxRefs; i++ {
		builder.SetRef(i, children[i])
	}
	result, _ := builder.Build()
	return result
}

// Iterate returns all entries with their Chunk values.
// Keys are not recoverable from the trie — iteration returns key hashes.
func (m *Map) Iterate() []Entry {
	if m.root == nil {
		return nil
	}
	var entries []Entry
	m.collectEntries(m.root, &entries)
	return entries
}

func (m *Map) collectEntries(node *Chunk, entries *[]Entry) {
	if node == nil {
		return
	}
	isLeaf := true
	for i := 0; i < MaxRefs; i++ {
		if node.RefAt(i) != nil {
			isLeaf = false
			m.collectEntries(node.RefAt(i), entries)
		}
	}
	if isLeaf && node.BitCount() > 0 {
		*entries = append(*entries, Entry{Key: nil, Value: node})
	}
}

// NewEmptyChunk returns a canonical empty chunk with type System.
func NewEmptyChunk() *Chunk {
	builder := NewBuilder().SetTypeTag(TypeSystem)
	chunk, _ := builder.Build()
	return chunk
}

// HashKey computes the BLAKE3 hash of the key.
func HashKey(key []byte) [32]byte {
	return blake3.Sum256(key)
}

// getIndex extracts 3 bits from the hash at the given bit offset.
// Path encoding: each level consumes 3 bits (0-7 → fanout branch index).
func getIndex(hash []byte, bitOffset int) int {
	byteIdx := bitOffset / 8
	bitIdx := bitOffset % 8

	var val uint32
	val = uint32(hash[byteIdx]) << 16
	if byteIdx+1 < len(hash) {
		val |= uint32(hash[byteIdx+1]) << 8
	}
	if byteIdx+2 < len(hash) {
		val |= uint32(hash[byteIdx+2])
	}

	shift := 24 - bitIdx - 3
	return int((val >> shift) & 0x07)
}
