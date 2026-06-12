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
		Primary:	addr(2),
		Contract:	addr(3),
		Records: map[string]sdk.AccAddress{
			ResolverKeyWallet: addr(4),
		},
		Metadata:	metadata,
	}, 12)
	require.NoError(t, err)

	record, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 13, 30)
	require.NoError(t, err)
	expectedHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	require.Equal(t, expectedHash, record.NameHash)
	require.Equal(t, addr(1), record.Owner)
	require.Equal(t, addr(2), record.PrimaryAddress)
	require.Equal(t, []ContractTargetV2{
		NewContractTargetV2(ResolverKeyContract, addr(3), 13),
		NewContractTargetV2(ResolverKeyWallet, addr(4), 13),
	}, record.ContractTargets)
	require.Equal(t, []ServiceEndpointV2{{
		Key:		"rpc",
		Endpoint:	"https://rpc.aet",
		ServiceID:	"rpc",
		ServiceType:	"service.v1",
		Transport:	"https",
		AuthPolicy:	"none",
		Priority:	100,
		Weight:		1,
		TTL:		30,
	}}, record.ServiceEndpoints)
	expectedDescriptorHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)
	require.Equal(t, []InterfaceDescriptorV2{{
		InterfaceID:	"aw5",
		Descriptor:	expectedDescriptorHash,
		SchemaHash:	expectedDescriptorHash,
		Version:	"v1",
		RenderPolicy:	"wallet_confirm",
	}}, record.InterfaceDescriptors)
	require.Equal(t, RoutingMetadataV2{ZoneID: "CONTRACT_ZONE", ShardID: "0:1", VM: "AVM", Entrypoint: "swap"}, record.RoutingMetadata)
	require.Equal(t, []ExecutionHintV2{{Key: "hint.priority", Value: "fast"}}, record.ExecutionHints)
	require.Equal(t, uint64(1), record.RecordVersion)
	require.Equal(t, uint64(30), record.RecordTTL)
	require.Equal(t, uint64(13), record.UpdatedAtHeight)
	require.Equal(t, uint64(MaxUnifiedPayloadBytesV2), record.MaxPayloadBytes)
	require.Equal(t, UnifiedResolutionSchemaVersionV2, record.SchemaVersion)
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
}

func TestUnifiedResolutionRecordV2BuildsRoutingAndExecutionHintsFromMetadata(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	metadata, err := EncodeResolverMetadata([]ResolverMetadataEntry{
		{Key: "route.id", Value: "swap-route"},
		{Key: "route.target_type", Value: string(IdentityResolutionTargetContract)},
		{Key: "route.preferred_target", Value: "swap"},
		{Key: "route.fallback_targets", Value: "swap-backup"},
		{Key: "route.fee_hint", Value: "standard"},
		{Key: "hint.default_gas_limit", Value: "250000"},
		{Key: "hint.requires_interface_confirmation", Value: "true"},
	})
	require.NoError(t, err)

	state, _, err = PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
		Metadata:	metadata,
	}, 12)
	require.NoError(t, err)

	record, err := BuildUnifiedResolutionRecordV2(state, "alice.aet", 13, 30)
	require.NoError(t, err)
	require.Equal(t, RoutingMetadataV2{
		RouteID:		"swap-route",
		TargetType:		string(IdentityResolutionTargetContract),
		PreferredTarget:	"swap",
		FallbackTargets:	[]string{"swap-backup"},
		FeeHint:		"standard",
	}, record.RoutingMetadata)
	require.Equal(t, []ExecutionHintV2{
		{Key: "hint.default_gas_limit", Value: "250000", DefaultGasLimitHint: 250_000},
		{Key: "hint.requires_interface_confirmation", Value: "true", RequiresInterfaceConfirmation: true},
	}, record.ExecutionHints)
}

func TestUnifiedResolutionRecordV2RejectsNonCanonicalAndTTL(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:		nameHash,
		Owner:			addr(1),
		PrimaryAddress:		addr(2),
		RecordVersion:		1,
		RecordTTL:		10,
		UpdatedAtHeight:	1,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
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
		NameHash:		nameHash,
		Owner:			addr(1),
		PrimaryAddress:		addr(2),
		ContractTargets:	[]ContractTargetV2{{Key: ResolverKeyContract, CodeID: "avm:swap-v1"}},
		ServiceEndpoints: []ServiceEndpointV2{{
			ServiceID:	"rpc",
			ServiceType:	"rpc.v1",
			Endpoint:	"https://rpc.aet",
			Transport:	"https",
			AuthPolicy:	"none",
			Weight:		1,
			TTL:		10,
		}},
		InterfaceDescriptors: []InterfaceDescriptorV2{{
			InterfaceID:	"aw5",
			SchemaHash:	descriptorHash,
			Version:	"v1",
			RenderPolicy:	"wallet_confirm",
		}},
		RoutingMetadata:	RoutingMetadataV2{ZoneID: "zone-a", ShardID: "shard-1", VM: "avm", Entrypoint: "swap"},
		ExecutionHints:		[]ExecutionHintV2{{Key: "hint.priority", Value: "fast"}},
		RecordVersion:		7,
		RecordTTL:		10,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	record.ServiceEndpoints[0].Endpoint = "ftp://rpc.aet"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "scheme")
	record.ServiceEndpoints[0].Endpoint = "https://rpc.aet"

	record.InterfaceDescriptors[0].SchemaHash = "wallet-v1"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "sha256:<64 hex>")
	record.InterfaceDescriptors[0].SchemaHash = descriptorHash

	record.ExecutionHints[0].Key = "exec.required"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "advisory")
	record.ExecutionHints[0].Key = "hint.priority"

	record.RoutingMetadata.ZoneID = strings.Repeat("z", MaxUnifiedRoutingMetadataBytes+1)
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(record), "routing metadata")
	require.NoError(t, ValidateResolverRecordVersionForUpdateV2(7, 7))
	require.ErrorContains(t, ValidateResolverRecordVersionForUpdateV2(8, 7), "version conflict")
}

func TestUnifiedResolutionRecordV2ContractTargetSchemaValidation(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	descriptorHash, err := InterfaceDescriptorHashV2("wallet-v1")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:		nameHash,
		Owner:			addr(1),
		PrimaryAddress:		addr(2),
		InterfaceDescriptors:	[]InterfaceDescriptorV2{{InterfaceID: "aw5", SchemaHash: descriptorHash, Version: "v1", RenderPolicy: "wallet_confirm"}},
		ContractTargets: []ContractTargetV2{{
			TargetID:		"swap",
			ContractAddress:	addr(3),
			Entrypoint:		"execute_swap",
			InterfaceHash:		descriptorHash,
			RequiredFundsPolicy:	"optional",
			GasHint:		500_000,
			Enabled:		true,
			UpdatedAtHeight:	12,
		}},
		RecordVersion:		7,
		RecordTTL:		10,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	duplicate := record
	duplicate.ContractTargets = append([]ContractTargetV2(nil), record.ContractTargets...)
	duplicate.ContractTargets = append(duplicate.ContractTargets, record.ContractTargets[0])
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(duplicate), "duplicate")

	badEntrypoint := record
	badEntrypoint.ContractTargets = append([]ContractTargetV2(nil), record.ContractTargets...)
	badEntrypoint.ContractTargets[0].Entrypoint = "bad entry"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badEntrypoint), "entrypoint")

	badInterface := record
	badInterface.ContractTargets = append([]ContractTargetV2(nil), record.ContractTargets...)
	badInterface.ContractTargets[0].InterfaceHash = InterfaceDescriptorHashPrefixV2 + strings.Repeat("a", 64)
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badInterface), "interface_hash")

	badGas := record
	badGas.ContractTargets = append([]ContractTargetV2(nil), record.ContractTargets...)
	badGas.ContractTargets[0].GasHint = MaxContractGasHintV2 + 1
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badGas), "gas_hint")
}

func TestUnifiedResolutionRecordV2RoutingMetadataValidation(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		PrimaryAddress:	addr(2),
		RoutingMetadata: RoutingMetadataV2{
			RouteID:		"swap-route",
			TargetType:		string(IdentityResolutionTargetContract),
			PreferredTarget:	"swap",
			FallbackTargets:	[]string{"swap-backup", "swap-cold"},
			ChainContext:		"aetra-main",
			FeeHint:		"standard",
			TimeoutHint:		30,
			MemoPolicy:		"optional",
			CapabilityRequirements:	[]string{"cap.contract.invoke", "cap.fee.pay"},
		},
		RecordVersion:		7,
		RecordTTL:		10,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
	require.True(t, routingMetadataHasTargetV2(record.RoutingMetadata))

	badFallback := record
	badFallback.RoutingMetadata.FallbackTargets = []string{"swap-cold", "swap-backup"}
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badFallback), "sorted")

	badFee := record
	badFee.RoutingMetadata.FeeHint = "override-fee-module"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badFee), "fee hints")

	badTarget := record
	badTarget.RoutingMetadata.TargetType = "validator"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badTarget), "target_type")

	badTimeout := record
	badTimeout.RoutingMetadata.TimeoutHint = MaxRouteTimeoutHintV2 + 1
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badTimeout), "timeout_hint")
}

func TestUnifiedResolutionRecordV2ExecutionHintValidation(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		PrimaryAddress:	addr(2),
		ExecutionHints: []ExecutionHintV2{{
			Key:				"hint.invoke",
			DefaultGasLimitHint:		250_000,
			PreferredFeeMode:		"standard",
			MessageType:			"cosmos.msg.v1",
			AsyncAllowed:			true,
			RequiresMemo:			true,
			RequiresInterfaceConfirmation:	true,
			SimulationRequired:		true,
		}},
		RecordVersion:		7,
		RecordTTL:		10,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	badGas := record
	badGas.ExecutionHints = append([]ExecutionHintV2(nil), record.ExecutionHints...)
	badGas.ExecutionHints[0].DefaultGasLimitHint = MaxExecutionGasLimitHintV2 + 1
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badGas), "default_gas_limit_hint")

	badFee := record
	badFee.ExecutionHints = append([]ExecutionHintV2(nil), record.ExecutionHints...)
	badFee.ExecutionHints[0].PreferredFeeMode = "bypass-ante"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badFee), "ante-handler")

	badKey := record
	badKey.ExecutionHints = append([]ExecutionHintV2(nil), record.ExecutionHints...)
	badKey.ExecutionHints[0].Key = "exec.invoke"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badKey), "hint.*")
}

func TestUnifiedResolutionRecordV2ServiceEndpointSchemaValidation(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	schemaHash, err := InterfaceDescriptorHashV2("rpc-service-v1")
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		PrimaryAddress:	addr(2),
		ServiceEndpoints: []ServiceEndpointV2{{
			ServiceID:		"rpc",
			ServiceType:		"rpc.v1",
			Endpoint:		"https://rpc.aet",
			Transport:		"https",
			AuthPolicy:		"none",
			HealthPathOptional:	"/healthz",
			Priority:		100,
			Weight:			1,
			TTL:			10,
			SchemaHashOptional:	schemaHash,
		}},
		RecordVersion:		7,
		RecordTTL:		10,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	badType := record
	badType.ServiceEndpoints = append([]ServiceEndpointV2(nil), record.ServiceEndpoints...)
	badType.ServiceEndpoints[0].ServiceType = "rpc"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badType), "versioned")

	badTTL := record
	badTTL.ServiceEndpoints = append([]ServiceEndpointV2(nil), record.ServiceEndpoints...)
	badTTL.ServiceEndpoints[0].TTL = 11
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badTTL), "ttl")

	badScheme := record
	badScheme.ServiceEndpoints = append([]ServiceEndpointV2(nil), record.ServiceEndpoints...)
	badScheme.ServiceEndpoints[0].Endpoint = "ftp://rpc.aet"
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badScheme), "scheme")
}

func TestUnifiedResolutionRecordV2InterfaceDescriptorSchemaValidation(t *testing.T) {
	nameHash, err := DomainRecordV2NameHash("alice.aet")
	require.NoError(t, err)
	inline := `{"type":"wallet","version":"v1"}`
	schemaHash, err := InterfaceDescriptorHashV2(inline)
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:	nameHash,
		Owner:		addr(1),
		PrimaryAddress:	addr(2),
		InterfaceDescriptors: []InterfaceDescriptorV2{{
			InterfaceID:			"aw5",
			SchemaHash:			schemaHash,
			SchemaURIOptional:		"https://schema.aet/wallet-v1.json",
			SchemaInlineOptional:		inline,
			Version:			"v1",
			RenderPolicy:			"wallet_confirm",
			PermissionsRequired:		[]string{"read.balance"},
			ContractTargetIDOptional:	"swap",
			ServiceIDOptional:		"rpc",
		}},
		RecordVersion:		7,
		RecordTTL:		10,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))

	badInline := record
	badInline.InterfaceDescriptors = append([]InterfaceDescriptorV2(nil), record.InterfaceDescriptors...)
	badInline.InterfaceDescriptors[0].SchemaInlineOptional = `{"type":"different"}`
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badInline), "inline schema")

	badPermission := record
	badPermission.InterfaceDescriptors = append([]InterfaceDescriptorV2(nil), record.InterfaceDescriptors...)
	badPermission.InterfaceDescriptors[0].PermissionsRequired = []string{"execute"}
	require.ErrorContains(t, ValidateUnifiedResolutionRecordV2(badPermission), "cannot grant execution")
}

func TestReverseResolutionRecordV2VerifiedPrimaryAndAlias(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	state, _, err := PatchIdentityResolver(state, "alice.aet", addr(1), ResolverPatch{
		Primary:	addr(2),
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
