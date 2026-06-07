package aw

import "encoding/binary"

type byteEncoder struct {
	out []byte
}

func newByteEncoder() *byteEncoder {
	return &byteEncoder{}
}

func (e *byteEncoder) writeBytes(bz []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(bz)))
	e.out = append(e.out, length[:]...)
	e.out = append(e.out, bz...)
}

func (e *byteEncoder) writeString(value string) {
	e.writeBytes([]byte(value))
}

func (e *byteEncoder) writeUint64(value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	e.out = append(e.out, bz[:]...)
}

func (e *byteEncoder) writeInt64(value int64) {
	e.writeUint64(uint64(value))
}

func (e *byteEncoder) writeBool(value bool) {
	if value {
		e.out = append(e.out, 1)
		return
	}
	e.out = append(e.out, 0)
}

func (e *byteEncoder) bytes() []byte {
	return append([]byte(nil), e.out...)
}
