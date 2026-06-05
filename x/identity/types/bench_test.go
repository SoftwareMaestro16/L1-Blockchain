package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const benchmarkIdentityDomainCount = 1000

func BenchmarkIdentityLookup(b *testing.B) {
	state := benchmarkIdentityState(b, benchmarkIdentityDomainCount)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := i % benchmarkIdentityDomainCount
		resolved, err := ResolveIdentityAddress(state, benchmarkIdentityName(index), benchmarkIdentityResolveHeight)
		if err != nil {
			b.Fatal(err)
		}
		if len(resolved) == 0 {
			b.Fatal("identity lookup returned empty address")
		}
	}
}

func BenchmarkIdentityExportImportLargeState(b *testing.B) {
	state := benchmarkIdentityState(b, benchmarkIdentityDomainCount)
	exported := state.Export()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imported, err := ImportIdentityState(exported)
		if err != nil {
			b.Fatal(err)
		}
		if len(imported.Domains) != benchmarkIdentityDomainCount {
			b.Fatal("identity domain count changed")
		}
	}
}

const (
	benchmarkIdentityRevealHeight  = uint64(11)
	benchmarkIdentityResolveHeight = uint64(12)
)

func benchmarkIdentityState(b *testing.B, count int) IdentityState {
	b.Helper()

	state := EmptyIdentityState(DefaultIdentityParams())
	for i := 0; i < count; i++ {
		name := benchmarkIdentityName(i)
		owner := benchmarkIdentityAddress(i + 1)
		nftID, err := DomainNFTID(name)
		if err != nil {
			b.Fatal(err)
		}
		state.Domains = append(state.Domains, Domain{
			Name:             name,
			Owner:            owner,
			NFTID:            nftID,
			RegisteredHeight: benchmarkIdentityRevealHeight,
			ExpiryHeight:     benchmarkIdentityRevealHeight + state.Params.RegistrationPeriodBlocks,
			UpdatedHeight:    benchmarkIdentityResolveHeight,
		})
		state.DomainNFTs = append(state.DomainNFTs, DomainNFT{
			ID:         nftID,
			Domain:     name,
			Owner:      owner,
			MintHeight: benchmarkIdentityRevealHeight,
		})
		state.Resolvers = append(state.Resolvers, ResolverRecord{
			Domain:  name,
			Owner:   owner,
			Primary: benchmarkIdentityAddress(i + 10_000),
			Records: map[string]sdk.AccAddress{
				ResolverKeyWallet: benchmarkIdentityAddress(i + 20_000),
			},
			UpdatedAtUnix: int64(benchmarkIdentityResolveHeight),
		})
	}
	state = state.Export()
	if err := state.Validate(); err != nil {
		b.Fatal(err)
	}
	return state
}

func benchmarkIdentityName(index int) string {
	return fmt.Sprintf("bench%04d.aet", index)
}

func benchmarkIdentityAddress(seed int) sdk.AccAddress {
	out := make([]byte, 20)
	for i := range out {
		out[i] = byte((seed + i + 1) % 251)
	}
	if out[19] == 0 {
		out[19] = 1
	}
	return sdk.AccAddress(out)
}
