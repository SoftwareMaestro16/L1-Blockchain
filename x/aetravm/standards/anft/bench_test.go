package anft

import "testing"

func BenchmarkNFTMint(b *testing.B) {
	admin := testAddr(1)
	owner := testAddr(2)
	state := newTestState(b, admin)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := state.MintNFT(admin, owner, testMetadata("Bench NFT")); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNFTTransfer(b *testing.B) {
	admin := testAddr(1)
	alice := testAddr(2)
	bob := testAddr(3)
	state := newTestState(b, admin)
	item, err := state.MintNFT(admin, alice, testMetadata("Transfer Bench NFT"))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		from, to := alice, bob
		if i%2 == 1 {
			from, to = bob, alice
		}
		if err := state.TransferNFT(from, item.Address, to); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSBTProofAndRevoke(b *testing.B) {
	admin := testAddr(1)
	owner := testAddr(2)
	authority := testAddr(3)
	state := newTestState(b, admin)
	item, err := state.MintSBT(admin, owner, authority, testMetadata("Proof Bench SBT"))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := state.ProveSBTOwnership(item.Address, owner); err != nil {
			b.Fatal(err)
		}
		if err := state.RevokeSBT(authority, item.Address, int64(i+1), "bench"); err != nil {
			b.Fatal(err)
		}
	}
}
