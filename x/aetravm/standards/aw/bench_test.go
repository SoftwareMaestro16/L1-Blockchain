package aw

import "testing"

func BenchmarkWalletSignedSend(b *testing.B) {
	state, _, privateKey := newTestState(b)
	recipient := testAddr(3)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := signedSend(b, state, privateKey, recipient, 1)
		if err := state.ApplyExternalCommand(cmd, testNow, naetFee()); err != nil {
			b.Fatal(err)
		}
	}
}
