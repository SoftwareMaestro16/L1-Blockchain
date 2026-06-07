package aft

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func BenchmarkTokenMasterMint(b *testing.B) {
	admin := testAddr(1)
	holder := testAddr(2)
	state := newTestState(b, admin)
	amount := sdkmath.NewInt(1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := state.Mint(admin, holder, amount); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenWalletTransfer(b *testing.B) {
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	state := newTestState(b, admin)
	if err := state.Mint(admin, alice, sdkmath.NewInt(1_000_000_000)); err != nil {
		b.Fatal(err)
	}
	amount := sdkmath.NewInt(1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := state.Transfer(alice, bob, amount, uint64(i+1)); err != nil {
			b.Fatal(err)
		}
	}
}
