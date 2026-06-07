package avm

import "testing"

func FuzzDecodeModuleRejectsMalformedWithoutPanic(f *testing.F) {
	valid, err := EncodeModule(counterModule())
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add([]byte("bad"))
	f.Add([]byte{0x00, 0x01, 0x02})

	f.Fuzz(func(t *testing.T, bz []byte) {
		module, err := DecodeModule(bz)
		if err != nil {
			return
		}
		verifier, err := NewVerifier(DefaultParams())
		if err != nil {
			t.Fatal(err)
		}
		_ = verifier.Verify(module)
	})
}
