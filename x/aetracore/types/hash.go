package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	HashHexLength	= 64
	EmptyRootHash	= "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type byteWriter interface {
	Write([]byte) (int, error)
}

func ValidateHash(fieldName, value string) error {
	if len(value) != HashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	return nil
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		writePart(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashRoot(domain string, write func(byteWriter)) string {
	h := sha256.New()
	writePart(h, domain)
	write(h)
	return hex.EncodeToString(h.Sum(nil))
}

func writePart(w byteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func writeUint64(w byteWriter, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}
