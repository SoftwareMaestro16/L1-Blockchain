package types

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func FuzzIdentityMalformedNamesV2(f *testing.F) {
	f.Add("")
	f.Add(".aet")
	f.Add("alice..aet")
	f.Add("Alice.aet")
	f.Add("alice.aet.")
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = NormalizeAETDomain(input)
		_ = ValidateDomainName(input)
	})
}

func FuzzIdentityBoundaryLengthNamesV2(f *testing.F) {
	f.Add(1, "a")
	f.Add(int(MaxDomainFullBytes), "b")
	f.Add(int(MaxDomainFullBytes)+1, "c")
	f.Fuzz(func(t *testing.T, length int, seed string) {
		if length < 0 {
			length = -length
		}
		if length > int(MaxDomainFullBytes)+32 {
			length = int(MaxDomainFullBytes) + 32
		}
		label := sanitizeFuzzDomainLabelV2(seed)
		nameBodyBytes := length - len(DomainTLD)
		if nameBodyBytes < 1 {
			nameBodyBytes = 1
		}
		name := strings.Repeat(label[:1], nameBodyBytes) + DomainTLD
		_, _ = NormalizeAETDomain(name)
		_ = ValidateDomainName(name)
	})
}

func FuzzIdentitySpoofingPatternCandidatesV2(f *testing.F) {
	f.Add("xn--alice.aet")
	f.Add("аlice.aet")
	f.Add("alice-.aet")
	f.Add("-alice.aet")
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = NormalizeAETDomainVersioned(input, NameNormalizationVersionV2)
	})
}

func FuzzAuctionBidRevealOrderingV2(f *testing.F) {
	f.Add(uint64(100), uint64(100), uint64(10), uint64(11), "left", "right")
	f.Add(uint64(1), uint64(2), uint64(20), uint64(20), "same-height-left", "same-height-right")
	f.Fuzz(func(t *testing.T, leftBid uint64, rightBid uint64, leftHeight uint64, rightHeight uint64, leftSalt string, rightSalt string) {
		if leftBid == 0 {
			leftBid = 1
		}
		if rightBid == 0 {
			rightBid = 1
		}
		if leftHeight == 0 {
			leftHeight = 1
		}
		if rightHeight == 0 {
			rightHeight = 1
		}
		leftCommitment, err := ComputeAuctionCommitment("alice.aet", addr(1), leftBid, leftSalt)
		if err != nil {
			t.Skip()
		}
		rightCommitment, err := ComputeAuctionCommitment("alice.aet", addr(2), rightBid, rightSalt)
		if err != nil {
			t.Skip()
		}
		reveals := []AuctionReveal{
			{Name: "alice.aet", Bidder: addr(2), Bid: rightBid, Salt: rightSalt, RevealHeight: rightHeight, CommitmentHash: rightCommitment},
			{Name: "alice.aet", Bidder: addr(1), Bid: leftBid, Salt: leftSalt, RevealHeight: leftHeight, CommitmentHash: leftCommitment},
		}
		sortAuctionReveals(reveals)
		if err := validateAuctionReveals(reveals); err != nil {
			t.Skip()
		}
		winner := chooseAuctionWinner(reveals)
		if len(winner.Bidder) == 0 {
			t.Fatal("auction winner is empty")
		}
	})
}

func FuzzInterfaceDescriptorSchemasV2(f *testing.F) {
	f.Add(`{"type":"wallet","version":"v1"}`, "wallet")
	f.Add(`{"schema":"contract"}`, "contract")
	f.Fuzz(func(t *testing.T, schema string, interfaceID string) {
		if len(schema) > MaxInterfaceInlineSchemaBytesV2 {
			schema = schema[:MaxInterfaceInlineSchemaBytesV2]
		}
		if interfaceID == "" {
			interfaceID = "iface"
		}
		hash, err := InterfaceDescriptorHashV2(schema)
		if err != nil {
			t.Skip()
		}
		record := safePayloadRecordV2(t)
		record.InterfaceDescriptors[0].InterfaceID = sanitizeFuzzRecordKeyV2(interfaceID)
		record.InterfaceDescriptors[0].SchemaInlineOptional = schema
		record.InterfaceDescriptors[0].SchemaHash = hash
		_ = ValidateUnifiedResolutionRecordV2(record)
	})
}

func FuzzDelegationPermissionCombinationsV2(f *testing.F) {
	f.Add(uint8(0), "primary", uint64(10))
	f.Add(uint8(1), "label.api", uint64(20))
	f.Add(uint8(4), "interface.wallet", uint64(30))
	f.Fuzz(func(t *testing.T, scopeIndex uint8, permission string, expiry uint64) {
		scopes := []DelegationScopeV2{
			DelegationScopeResolverUpdate,
			DelegationScopeSubdomainCreate,
			DelegationScopeSubdomainTransfer,
			DelegationScopeServiceRecordUpdate,
			DelegationScopeInterfaceRecordUpdate,
			DelegationScopeRoutingRecordUpdate,
			DelegationScopeZoneAdmin,
		}
		scope := scopes[int(scopeIndex)%len(scopes)]
		if expiry <= 10 {
			expiry = 11
		}
		permission = sanitizeFuzzDelegationPermissionV2(permission)
		record, err := NewDelegationRecordV2("alice.aet", addr(7), scope, []string{permission}, expiry, 2, permission, 10)
		if err != nil {
			t.Skip()
		}
		_ = ValidateDelegationRecordV2Use(record, scope, permission, permission, 1, 10)
	})
}

func FuzzRecursiveProofPathsV2(f *testing.F) {
	f.Add("api", uint8(1))
	f.Add("svc", uint8(2))
	f.Add("edge", uint8(3))
	f.Fuzz(func(t *testing.T, labelSeed string, depth uint8) {
		if depth == 0 {
			depth = 1
		}
		if depth > MaxDomainLabels-1 {
			depth = MaxDomainLabels - 1
		}
		label := sanitizeFuzzDomainLabelV2(labelSeed)
		labels := make([]string, 0, depth+1)
		for i := uint8(0); i < depth; i++ {
			labels = append(labels, fmt.Sprintf("%s%d", label, i))
		}
		labels = append(labels, "alice")
		path, err := CanonicalResolutionPathV2(strings.Join(labels, ".") + DomainTLD)
		if err != nil {
			t.Skip()
		}
		commitment, err := BuildIdentityPathCommitmentV2(path, 1, 1, 1)
		if err != nil {
			t.Skip()
		}
		if err := ValidateIdentityPathCommitmentV2(commitment); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzReverseResolutionMismatchesV2(f *testing.F) {
	f.Add(byte(2), true)
	f.Add(byte(9), true)
	f.Add(byte(9), false)
	f.Fuzz(func(t *testing.T, lastByte byte, verified bool) {
		state, domain := registerSpecDomain(t, "alice", addr(1), "salt", 10)
		state, _, err := PatchIdentityResolver(state, domain.Name, addr(1), ResolverPatch{Primary: addr(2)}, 12)
		if err != nil {
			t.Fatal(err)
		}
		candidate := addr(2)
		candidate[19] = lastByte
		record, err := NewReverseResolutionRecordV2(candidate, domain.Name, verified, 13, domain.ExpiryHeight)
		if err != nil {
			t.Skip()
		}
		err = ValidateReverseResolutionRecordV2(state, record, 14, nil)
		if verified && !bytes.Equal(candidate, addr(2)) && err == nil {
			t.Fatal("verified reverse mismatch was accepted")
		}
	})
}

func FuzzBatchUpdateOrderingV2(f *testing.F) {
	f.Add("alice", "bob", false)
	f.Add("alice", "alice", true)
	f.Fuzz(func(t *testing.T, leftLabel string, rightLabel string, duplicate bool) {
		left := sanitizeFuzzDomainLabelV2(leftLabel) + DomainTLD
		right := sanitizeFuzzDomainLabelV2(rightLabel) + DomainTLD
		if duplicate {
			right = left
		}
		leftHash, err := DomainRecordV2NameHash(left)
		if err != nil {
			t.Skip()
		}
		rightHash, err := DomainRecordV2NameHash(right)
		if err != nil {
			t.Skip()
		}
		msg := MsgBatchUpdateResolversV2{
			Auth:	benchmarkIdentityTxAuth(IdentitySignerScopeBatchAdmin, 1),
			Updates: []ResolverBatchUpdateV2{
				{Name: left, NameHash: leftHash, Patch: ResolverPatch{Primary: addr(3)}, ExpectedRecordVersion: 1, RecordTTL: 10},
				{Name: right, NameHash: rightHash, Patch: ResolverPatch{Primary: addr(4)}, ExpectedRecordVersion: 1, RecordTTL: 10},
			},
		}
		err = msg.ValidateBasic()
		if duplicate && err == nil {
			t.Fatal("duplicate batch update was accepted")
		}
	})
}

func sanitizeFuzzDomainLabelV2(input string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(input) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
		if builder.Len() >= MaxDomainLabelBytes {
			break
		}
	}
	out := strings.Trim(builder.String(), "-")
	if out == "" {
		return "svc"
	}
	return out
}

func sanitizeFuzzRecordKeyV2(input string) string {
	out := sanitizeFuzzDelegationPermissionV2(input)
	if out == "" {
		return "key"
	}
	return out
}

func sanitizeFuzzDelegationPermissionV2(input string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(input) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
		if builder.Len() >= MaxDelegationPermissionBytesV2 {
			break
		}
	}
	out := strings.Trim(builder.String(), ".-_")
	if out == "" {
		return DelegationPermissionCreateV2
	}
	return out
}
