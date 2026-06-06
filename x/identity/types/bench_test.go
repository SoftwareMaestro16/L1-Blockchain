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

func BenchmarkIdentityStoreV2DirectResolutionReadPath(b *testing.B) {
	names := benchmarkIdentityNames(benchmarkIdentityDomainCount)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessSet, err := IdentityStoreV2SpecDirectResolutionReadAccessSet(names[i%len(names)], false)
		if err != nil {
			b.Fatal(err)
		}
		if len(accessSet.Reads) != 2 {
			b.Fatal("direct resolution read path is not compact")
		}
	}
}

func BenchmarkIdentityStoreV2RecursiveResolutionReadPath(b *testing.B) {
	names := make([]string, benchmarkIdentityDomainCount)
	for i := range names {
		names[i] = fmt.Sprintf("svc.api.%s", benchmarkIdentityName(i))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		accessSet, err := IdentityStoreV2SpecRecursiveResolutionReadAccessSet(names[i%len(names)], false)
		if err != nil {
			b.Fatal(err)
		}
		if len(accessSet.Reads) != 4 {
			b.Fatal("recursive resolution read path should read one domain per label plus final resolver")
		}
	}
}

func BenchmarkIdentityProofQuery(b *testing.B) {
	state := benchmarkIdentityState(b, benchmarkIdentityDomainCount)
	query := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: state, Height: benchmarkIdentityResolveHeight, DefaultTTL: 30})
	names := benchmarkIdentityNames(benchmarkIdentityDomainCount)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := query.QueryResolutionProof(names[i%len(names)])
		if resp.Code != IdentityQueryOK || resp.Proof == nil {
			b.Fatalf("proof query failed: %s", resp.Error)
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

func benchmarkIdentityNames(count int) []string {
	names := make([]string, count)
	for i := range names {
		names[i] = benchmarkIdentityName(i)
	}
	return names
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
