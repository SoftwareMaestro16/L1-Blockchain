package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var avmForbiddenBytecodeTokens = [][]byte{
	[]byte("time.now"),
	[]byte("wall_clock"),
	[]byte("random"),
	[]byte("filesystem"),
	[]byte("network"),
	[]byte("float"),
}

func ValidateAVMBytecode(params Params, bytecode []byte) error {
	if len(bytecode) == 0 {
		return errors.New(ErrInvalidBytecode + ": bytecode is required")
	}
	if uint64(len(bytecode)) > params.MaxCodeBytes {
		return errors.New(ErrInvalidBytecode + ": code size out of bounds")
	}
	if !bytes.HasPrefix(bytecode, []byte("AVM1")) {
		return errors.New(ErrInvalidBytecode + ": unsupported AVM bytecode header")
	}
	lower := []byte(strings.ToLower(string(bytecode)))
	for _, token := range avmForbiddenBytecodeTokens {
		if bytes.Contains(lower, token) {
			return errors.New(ErrInvalidBytecode + ": forbidden nondeterministic host capability")
		}
	}
	return nil
}

func CanonicalCodeHash(bytecode []byte) string {
	sum := sha256.Sum256(append([]byte("aetra-avm-code-v1/"), bytecode...))
	return hex.EncodeToString(sum[:])
}
