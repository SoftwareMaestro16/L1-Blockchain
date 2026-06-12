package types

import (
	"encoding/binary"
	"fmt"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

// Reader is a read cursor over Chunk data and its references.
type Reader struct {
	chunk		*chunk.Chunk
	data		[]byte
	offset		int	// bit offset
	refIndex	int
}

func NewReader(c *chunk.Chunk) *Reader {
	return &Reader{
		chunk:	c,
		data:	c.Data(),
	}
}

func (r *Reader) ReadBool() (bool, error) {
	b, err := r.ReadBits(1)
	if err != nil {
		return false, err
	}
	return (b[0] & 0x80) != 0, nil
}

func (r *Reader) ReadUint8() (uint8, error) {
	b, err := r.ReadBits(8)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (r *Reader) ReadUint32() (uint32, error) {
	b, err := r.ReadBits(32)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (r *Reader) ReadUint64() (uint64, error) {
	b, err := r.ReadBits(64)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}

func (r *Reader) ReadRef() (*chunk.Chunk, error) {
	refs := r.chunk.Refs()
	if r.refIndex >= len(refs) {
		return nil, fmt.Errorf("no more refs in chunk")
	}
	ref := refs[r.refIndex]
	r.refIndex++
	return ref, nil
}

func (r *Reader) ReadBits(n int) ([]byte, error) {
	if r.offset+n > int(r.chunk.BitCount()) {
		return nil, fmt.Errorf("read out of bounds: offset %d + bits %d > %d", r.offset, n, r.chunk.BitCount())
	}

	if r.offset%8 == 0 && n%8 == 0 {
		start := r.offset / 8
		end := start + (n / 8)
		r.offset += n
		return append([]byte(nil), r.data[start:end]...), nil
	}

	return nil, fmt.Errorf("unaligned bit reading not yet implemented")
}

// Writer is an immutable Chunk constructor with typed methods.
type Writer struct {
	builder		*chunk.Builder
	data		[]byte
	bitCount	int
	refs		[]*chunk.Chunk
}

func NewWriter() *Writer {
	return &Writer{
		builder: chunk.NewBuilder(),
	}
}

func (w *Writer) WriteBool(v bool) error {
	if v {
		return w.WriteBits([]byte{0x80}, 1)
	}
	return w.WriteBits([]byte{0x00}, 1)
}

func (w *Writer) WriteUint8(v uint8) error {
	return w.WriteBits([]byte{v}, 8)
}

func (w *Writer) WriteUint32(v uint32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return w.WriteBits(b, 32)
}

func (w *Writer) WriteUint64(v uint64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return w.WriteBits(b, 64)
}

func (w *Writer) AddRef(c *chunk.Chunk) error {
	w.refs = append(w.refs, c)
	return nil
}

func (w *Writer) WriteBits(data []byte, n int) error {

	if w.bitCount%8 == 0 && n%8 == 0 {
		w.data = append(w.data, data...)
		w.bitCount += n
		return nil
	}
	return fmt.Errorf("unaligned bit writing not yet implemented")
}

func (w *Writer) Build() (*chunk.Chunk, error) {
	b := chunk.NewBuilder().
		SetData(w.data, uint16(w.bitCount))
	for _, ref := range w.refs {
		b.AddRef(ref)
	}
	return b.Build()
}
