package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestUnifiedResolutionRecordV2BuildsFromResolverView(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	serviceKey, err := ResolverMetadataServiceKey("rpc")
	require.NoError(t, err)
	interfaceKey, err := ResolverMetadataInterfaceKey("aw5")
	require.NoError(t, err)
	metadata, err := EncodeResolverMetadata([]ResolverMetadataEntry{
		{Key: ResolverMetadataRouteZone, Value: "CONTRACT_ZONE"},
		{Key: ResolverMetadataRouteShard, Value: "0:1"},
		{Key: ResolverMetadataRouteVM, Value: "AVM"},
		{Key: ResolverMetadataRouteEntrypoint, Value: "swap"},
		{Key: serviceKey, Value: "https://rpc.aet"},
		{Key: interfaceKey, Value: "wallet-v1"},
		{Key: "hint.priority", Value: "fast"},
	})
	require.NoError(t, err)

	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:  addr(2),
		Contract: addr(3),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(4),
		},
		Metadata: metadata,
	}, 12)
	require.NoError(t, err)

	record, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 13, 30)
	require.NoError(t, err)
	expectedHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	require.Equal(t, expectedHash, record.NameHash)
	require.Equal(t, addr(2), record.PrimaryAddress)
	require.Equal(t, []ContractTargetV2{
		{Key: ResolverKeyContract, Address: addr(3)},
		{Key: ResolverKeyWallet, Address: addr(4)},
	}, record.ContractTargets)
	require.Equal(t, []ServiceEndpointV2{{Key: "rpc", Endpoint: "https://rpc.aet"}}, record.ServiceEndpoints)
	expectedDescriptorHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)
	require.Equal(t, []InterfaceDescriptorV2{{InterfaceID: "aw5", Descriptor: expectedDescriptorHash}}, record.InterfaceDescriptors)
	require.Equal(t, RoutingMetadataV2{ZoneID: "CONTRACT_ZONE", ShardID: "0:1", VM: "AVM", Entrypoint: "swap"}, record.RoutingMetadata)
	require.Equal(t, []ExecutionHintV2{{Key: "hint.priority", Value: "fast"}}, record.ExecutionHints)
	require.Equal(t, uint64(1), record.RecordVersion)
	require.Equal(t, uint64(30), record.RecordTTL)
	require.Equal(t, uint64(13), record.UpdatedAtHeight)
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
}

func TestUnifiedResolutionRecordV2RejectsNonCanonicalAndTTL(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:        nameHash,
		PrimaryAddress:  addr(2),
		RecordVersion:   1,
		RecordTTL:       10,
		UpdatedAtHeight: 1,
		ContractTargets: []ContractTargetV2{
			{Key: ResolverKeyWallet, Address: addr(4)},
			{Key: ResolverKeyContract, Address: addr(3)},
		},
	}
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "contract targets must be sorted")

	record.ContractTargets = []ContractTargetV2{{Key: ResolverKeyContract, Address: addr(3)}}
	record.RecordTTL = 0
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "ttl is required")
}

func TestUnifiedResolutionRecordV2ResolverValidationLimitsAndFormats(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	descriptorHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:             nameHash,
		PrimaryAddress:       addr(2),
		ContractTargets:      []ContractTargetV2{{Key: ResolverKeyContract, CodeID: "avm:swap-v1"}},
		ServiceEndpoints:     []ServiceEndpointV2{{Key: "rpc", Endpoint: "https://rpc.aet"}},
		InterfaceDescriptors: []InterfaceDescriptorV2{{InterfaceID: "aw5", Descriptor: descriptorHash}},
		RoutingMetadata:      RoutingMetadataV2{ZoneID: "zone-a", ShardID: "shard-1", VM: "avm", Entrypoint: "swap"},
		ExecutionHints:       []ExecutionHintV2{{Key: "hint.priority", Value: "fast"}},
		RecordVersion:        7,
		RecordTTL:            10,
		UpdatedAtHeight:      12,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	record.ServiceEndpoints[0].Endpoint = "ftp://rpc.aet"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "scheme")
	record.ServiceEndpoints[0].Endpoint = "https://rpc.aet"

	record.InterfaceDescriptors[0].Descriptor = "wallet-v1"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "sha256:<64 hex>")
	record.InterfaceDescriptors[0].Descriptor = descriptorHash

	record.ExecutionHints[0].Key = "exec.required"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "advisory")
	record.ExecutionHints[0].Key = "hint.priority"

	record.RoutingMetadata.ZoneID = strings.Repeat("z", MaxUnifiedRoutingMetadataBytes+1)
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "routing metadata")
	require.NoError(t, ValidateResolverRecordVersionForUpdateV2(7, 7))
	require.ErrorContains(t, ValidateResolverRecordVersionForUpdateV2(8, 7), "version conflict")
}

func TestReverseResolutionRecordV2VerifiedPrimaryAndAlias(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary: addr(2),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(3),
		},
	}, 12)
	require.NoError(t, err)

	primary, err := NewReverseResolutionRecordV2(addr(2), "alice.aet", true, 13, 100)
	require.NoError(t, err)
	require.NoError(t, ValidateReverseResolutionRecordV2(state, primary, 14, nil))

	alias, err := NewReverseResolutionRecordV2(addr(3), "alice.aet", true, 13, 100)
	require.NoError(t, err)
	require.NoError(t, ValidateReverseResolutionRecordV2(state, alias, 14, []string{ResolverKeyWallet}))
	require.ErrorContains(t, ValidateReverseResolutionRecordV2(state, alias, 14, nil), "forward primary or authorized alias")
}

func TestReverseResolutionRecordV2UnverifiedIsNotCanonical(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	record, err := NewReverseResolutionRecordV2(addr(2), "alice.aet", false, 13, 100)
	require.NoError(t, err)
	require.NoError(t, ValidateReverseResolutionRecordV2(state, record, 14, nil))

	_, err = CanonicalReverseResolutionName(record)
	require.ErrorContains(t, err, "unverified reverse record is not canonical")
	require.ErrorContains(t, ValidateReverseResolutionRecordV2(state, record, 100, nil), "reverse record is expired")
}

func TestReverseResolutionRecordV2RequiresActiveDomainExpiryAndInvalidates(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	state, reverse, err := SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)

	domain, found := findDomain(state, "alice.aet")
	require.True(t, found)
	verified, err := NewReverseResolutionRecordV2(addr(2), "alice.aet", true, 13, domain.ExpiryHeight+1)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateReverseResolutionRecordV2(state, verified, 14, nil), "expires after domain")

	verified.ExpiryHeight = domain.ExpiryHeight
	require.NoError(t, ValidateReverseResolutionRecordV2(state, verified, 14, nil))

	missing, err := NewReverseResolutionRecordV2(addr(2), "missing.aet", false, 13, 100)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateReverseResolutionRecordV2(state, missing, 14, nil), "existing active domain")

	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(3)}, 15)
	require.NoError(t, err)
	require.Empty(t, state.ReverseRecords)

	state.ReverseRecords = []ReverseRecord{reverse}
	next, invalidated, err := InvalidateReverseRecordsForDomainV2(state, "alice.aet", 16, nil)
	require.NoError(t, err)
	require.Len(t, invalidated, 1)
	require.Empty(t, next.ReverseRecords)
}

func TestReverseResolutionVerificationTransactionV2ChecksVersionAndForward(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{Primary: addr(2)}, 12)
	require.NoError(t, err)
	domain, found := findDomain(state, "alice.aet")
	require.True(t, found)
	reverse, err := NewReverseResolutionRecordV2(addr(2), "alice.aet", true, 13, domain.ExpiryHeight)
	require.NoError(t, err)

	msg := MsgVerifyReverseRecordV2{Auth: txAuth(IdentitySignerScopeReverseUpdate, 90), Record: reverse, ExpectedRecordVersion: 4}
	verified, err := VerifyReverseResolutionTransactionV2(state, msg, 14, 4)
	require.NoError(t, err)
	require.True(t, verified.Verified)

	_, err = VerifyReverseResolutionTransactionV2(state, msg, 14, 5)
	require.ErrorContains(t, err, "version conflict")
	msg.Record.Verified = false
	_, err = VerifyReverseResolutionTransactionV2(state, msg, 14, 4)
	require.ErrorContains(t, err, "verify requires verified record")
}
