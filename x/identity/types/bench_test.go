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

func BenchmarkIdentityBlockSTMBatchResolverUpdates(b *testing.B) {
	msg := benchmarkBatchResolverUpdateMsg(b, MaxIdentityTxBatchResolverUpdatesV2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set, hashes, err := IdentityBlockSTMBatchResolverAccessSetV2(msg)
		if err != nil {
			b.Fatal(err)
		}
		if len(hashes) != len(msg.Updates) || len(set.Writes) != len(msg.Updates) {
			b.Fatal("batch resolver update access set is not disjoint")
		}
	}
}

func BenchmarkIdentityResolverUpdateWritePath(b *testing.B) {
	state := benchmarkIdentityState(b, benchmarkIdentityDomainCount)
	name := benchmarkIdentityName(0)
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		b.Fatal(err)
	}
	resolver, found := findResolver(state, name)
	if !found {
		b.Fatal("benchmark resolver missing")
	}
	msg := MsgBatchUpdateResolversV2{
		Auth:	benchmarkIdentityTxAuth(IdentitySignerScopeBatchAdmin, 1),
		Updates: []ResolverBatchUpdateV2{{
			Name:			name,
			NameHash:		nameHash,
			Patch:			ResolverPatch{Primary: benchmarkIdentityAddress(60_000)},
			ExpectedRecordVersion:	ResolverRecordVersionV2(resolver),
			RecordTTL:		30,
		}},
	}
	msg.Auth.Signer = benchmarkIdentityAddress(1)
	options := IdentityBatchResolverUpdateOptionsV2{
		Mode:		IdentityBatchFailureAtomicV2,
		Height:		benchmarkIdentityResolveHeight + 1,
		GasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		GasLimit:	MinIdentityBatchResolverUpdateGasV2,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		next, response, err := ExecuteBatchResolverUpdatesV2(state, msg, options)
		if err != nil {
			b.Fatal(err)
		}
		if response.Successes != 1 || len(next.Resolvers) != len(state.Resolvers) {
			b.Fatal("resolver update write path failed")
		}
	}
}

func BenchmarkIdentityBlockSTMBatchRenewalsPerBlock(b *testing.B) {
	msg := benchmarkBatchRenewDomainsMsg(b, MaxIdentityTxBatchRenewDomainsV2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plan, err := IdentityBlockSTMAccessSetV2(msg, benchmarkIdentityResolveHeight)
		if err != nil {
			b.Fatal(err)
		}
		if len(plan.NameHashes) != len(msg.Renewals) || len(plan.AccessSet.Writes) != len(msg.Renewals) {
			b.Fatal("batch renewal access set is not disjoint")
		}
	}
}

func BenchmarkIdentityBlockSTMMixedConflictClassification(b *testing.B) {
	aliceHash, err := DomainRecordV2NameHash("alice.aet")
	if err != nil {
		b.Fatal(err)
	}
	bobHash, err := DomainRecordV2NameHash("bob.aet")
	if err != nil {
		b.Fatal(err)
	}
	plans := []IdentityBlockSTMPlanV2{
		benchmarkBlockSTMPlan(b, MsgUpdateResolverRecordV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeResolverUpdate, 1), Name: "alice.aet", NameHash: aliceHash, Patch: ResolverPatch{Primary: benchmarkIdentityAddress(1)}, ExpectedRecordVersion: 1, RecordTTL: 10}),
		benchmarkBlockSTMPlan(b, MsgTransferDomainV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeOwner, 2), Name: "alice.aet", NameHash: aliceHash, NewOwner: benchmarkIdentityAddress(2), ExpectedRecordVersion: 1}),
		benchmarkBlockSTMPlan(b, MsgRenewDomainV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeOwner, 3), Name: "bob.aet", NameHash: bobHash, ExpectedRecordVersion: 1}),
		benchmarkBlockSTMPlan(b, MsgRevealBidV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeAuctionBidder, 4), AuctionID: identityHash("auction"), NameHash: aliceHash, Bid: 100, Salt: "salt", CommitmentHash: identityHash("commit")}),
		benchmarkBlockSTMPlan(b, MsgFinalizeAuctionV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeAuctionAdmin, 5), AuctionID: identityHash("auction"), NameHash: aliceHash, ExpectedAuctionVersion: 1}),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		class := IdentityBlockSTMConflictClassifyV2(plans[i%len(plans)], plans[(i+1)%len(plans)])
		if class == "" {
			b.Fatal("empty BlockSTM conflict class")
		}
	}
}

func BenchmarkIdentityRegistrationsPerBlock(b *testing.B) {
	names := benchmarkIdentityNames(MaxIdentityTxBatchRenewDomainsV2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := names[i%len(names)]
		nameHash, err := DomainRecordV2NameHash(name)
		if err != nil {
			b.Fatal(err)
		}
		msg := MsgRegisterDirectV2{
			Auth:			benchmarkIdentityTxAuth(IdentitySignerScopeRegistration, uint64(i+1)),
			Name:			name,
			NameHash:		nameHash,
			Owner:			benchmarkIdentityAddress(i + 50_000),
			ExpectedRecordVersion:	1,
		}
		plan, err := IdentityBlockSTMAccessSetV2(msg, benchmarkIdentityResolveHeight)
		if err != nil {
			b.Fatal(err)
		}
		if len(plan.AccessSet.Writes) == 0 || len(plan.NameHashes) != 1 {
			b.Fatal("registration access set is empty")
		}
	}
}

func BenchmarkIdentityAdaptiveSyncRecoveryLargeState(b *testing.B) {
	state := benchmarkIdentityState(b, benchmarkIdentityDomainCount)
	snapshot, err := BuildIdentityAdaptiveSyncSnapshotV2(state, nil, benchmarkIdentityResolveHeight)
	if err != nil {
		b.Fatal(err)
	}
	probes := []string{benchmarkIdentityName(0), benchmarkIdentityName(benchmarkIdentityDomainCount - 1)}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		restored, err := RestoreIdentityAdaptiveSyncSnapshotV2(snapshot, probes)
		if err != nil {
			b.Fatal(err)
		}
		if len(restored.State.Domains) != benchmarkIdentityDomainCount {
			b.Fatal("adaptive sync recovery changed domain count")
		}
	}
}

func BenchmarkIdentityResolverUpdateSpamGasCostModelV2(b *testing.B) {
	params := DefaultIdentitySpamCostParamsV2()
	request := IdentitySpamCostRequestV2{
		ResolverUpdateCount:		MaxIdentityTxBatchResolverUpdatesV2,
		ResolverPayloadBytes:		MaxUnifiedPayloadBytesV2,
		BatchResolverGasPerUpdate:	MinIdentityBatchResolverUpdateGasV2,
		ServiceEndpointCount:		MaxUnifiedServiceEndpoints,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quote, err := EstimateIdentitySpamCostV2(request, params)
		if err != nil {
			b.Fatal(err)
		}
		if quote.ResolverUpdateGas == 0 || quote.ResolverPayloadCost.IsZero() {
			b.Fatal("resolver update spam gas cost model returned empty cost")
		}
	}
}

const (
	benchmarkIdentityRevealHeight	= uint64(11)
	benchmarkIdentityResolveHeight	= uint64(12)
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
			Name:			name,
			Owner:			owner,
			NFTID:			nftID,
			RegisteredHeight:	benchmarkIdentityRevealHeight,
			ExpiryHeight:		benchmarkIdentityRevealHeight + state.Params.RegistrationPeriodBlocks,
			UpdatedHeight:		benchmarkIdentityResolveHeight,
		})
		state.DomainNFTs = append(state.DomainNFTs, DomainNFT{
			ID:		nftID,
			Domain:		name,
			Owner:		owner,
			MintHeight:	benchmarkIdentityRevealHeight,
		})
		state.Resolvers = append(state.Resolvers, ResolverRecord{
			Domain:		name,
			Owner:		owner,
			Primary:	benchmarkIdentityAddress(i + 10_000),
			Records: map[string]sdk.AccAddress{
				ResolverKeyWallet: benchmarkIdentityAddress(i + 20_000),
			},
			UpdatedAtUnix:	int64(benchmarkIdentityResolveHeight),
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

func benchmarkBatchResolverUpdateMsg(b *testing.B, count int) MsgBatchUpdateResolversV2 {
	b.Helper()
	updates := make([]ResolverBatchUpdateV2, count)
	for i := range updates {
		name := benchmarkIdentityName(i)
		nameHash, err := DomainRecordV2NameHash(name)
		if err != nil {
			b.Fatal(err)
		}
		updates[i] = ResolverBatchUpdateV2{
			Name:			name,
			NameHash:		nameHash,
			Patch:			ResolverPatch{Primary: benchmarkIdentityAddress(i + 30_000)},
			ExpectedRecordVersion:	1,
			RecordTTL:		30,
		}
	}
	return MsgBatchUpdateResolversV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeBatchAdmin, 1), Updates: updates}
}

func benchmarkBatchRenewDomainsMsg(b *testing.B, count int) MsgBatchRenewDomainsV2 {
	b.Helper()
	renewals := make([]RenewDomainBatchItemV2, count)
	for i := range renewals {
		name := benchmarkIdentityName(i)
		nameHash, err := DomainRecordV2NameHash(name)
		if err != nil {
			b.Fatal(err)
		}
		renewals[i] = RenewDomainBatchItemV2{
			Name:			name,
			NameHash:		nameHash,
			ExpectedRecordVersion:	1,
		}
	}
	return MsgBatchRenewDomainsV2{Auth: benchmarkIdentityTxAuth(IdentitySignerScopeBatchAdmin, 2), Renewals: renewals}
}

func benchmarkBlockSTMPlan(b *testing.B, msg IdentityMsgV2) IdentityBlockSTMPlanV2 {
	b.Helper()
	plan, err := IdentityBlockSTMAccessSetV2(msg, benchmarkIdentityResolveHeight)
	if err != nil {
		b.Fatal(err)
	}
	return plan
}

func benchmarkIdentityTxAuth(scope IdentitySignerScopeV2, nonce uint64) IdentityTxAuthV2 {
	return IdentityTxAuthV2{
		ChainID:			"aetra-local-1",
		Signer:				benchmarkIdentityAddress(int(nonce + 40_000)),
		Scope:				scope,
		NameNormalizationVersion:	NameNormalizationVersionV2,
		Nonce:				nonce,
		Fee:				1,
		StorageCost:			1,
	}
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
