package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobalStatePrefixModelCoversSectionElevenOne(t *testing.T) {
	require.NoError(t, ValidateGlobalStatePrefixModel())

	model, err := DefaultGlobalStatePrefixModel()
	require.NoError(t, err)
	require.NoError(t, model.Validate())
	require.Len(t, model.Entries, 17)
	require.NotEmpty(t, model.Root)

	byPrefix := map[string]GlobalStatePrefixDescriptor{}
	for _, entry := range model.Entries {
		require.NoError(t, entry.Validate())
		byPrefix[string(entry.Namespace)+"|"+entry.Prefix] = entry
	}

	require.Equal(t, "Core root", byPrefix["Core|core/*"].ProofScope)
	require.Contains(t, byPrefix["Core|core/zones/*"].Purpose, "Zone descriptors")
	require.Equal(t, "Global zone root", byPrefix["Core|core/zone_roots/*"].ProofScope)
	require.Equal(t, "Global message root", byPrefix["Core|core/message_roots/*"].ProofScope)
	require.Equal(t, "Proof registry root", byPrefix["Core|core/proof_roots/*"].ProofScope)

	require.Equal(t, "Zone state root", byPrefix["Zone|zone/{zone_id}/params"].ProofScope)
	require.Equal(t, "Shard layout root", byPrefix["Zone|zone/{zone_id}/shards/*"].ProofScope)
	require.Equal(t, "Zone state root", byPrefix["Zone|zone/{zone_id}/state/*"].ProofScope)
	require.Equal(t, "Zone inbox root", byPrefix["Zone|zone/{zone_id}/inbox/*"].ProofScope)
	require.Equal(t, "Zone outbox root", byPrefix["Zone|zone/{zone_id}/outbox/*"].ProofScope)
	require.Equal(t, "Zone receipt root", byPrefix["Zone|zone/{zone_id}/receipts/*"].ProofScope)
	require.Equal(t, "Zone event root", byPrefix["Zone|zone/{zone_id}/events/*"].ProofScope)

	require.Equal(t, "Shard state root", byPrefix["Shard|zone/{zone_id}/shard/{shard_id}/state/*"].ProofScope)
	require.Equal(t, "Shard inbox root", byPrefix["Shard|zone/{zone_id}/shard/{shard_id}/inbox/*"].ProofScope)
	require.Equal(t, "Shard outbox root", byPrefix["Shard|zone/{zone_id}/shard/{shard_id}/outbox/*"].ProofScope)
	require.Equal(t, "Shard receipt root", byPrefix["Shard|zone/{zone_id}/shard/{shard_id}/receipts/*"].ProofScope)
	require.Equal(t, "Shard metrics root", byPrefix["Shard|zone/{zone_id}/shard/{shard_id}/metrics/*"].ProofScope)
}

func TestGlobalStatePrefixModelRootIsCanonicalAcrossInputOrder(t *testing.T) {
	model, err := DefaultGlobalStatePrefixModel()
	require.NoError(t, err)

	reordered := append([]GlobalStatePrefixDescriptor(nil), GlobalStatePrefixDescriptors()...)
	slices.Reverse(reordered)
	reorderedModel, err := BuildGlobalStatePrefixModel(reordered)
	require.NoError(t, err)
	require.Equal(t, model.Root, reorderedModel.Root)
	require.Equal(t, model.Entries, reorderedModel.Entries)
}

func TestMaterializeStatePrefixBindsZoneAndShardPlaceholders(t *testing.T) {
	zonePrefix, err := MaterializeStatePrefix("zone/{zone_id}/inbox/*", ZoneIDFinancial, "")
	require.NoError(t, err)
	require.Equal(t, "zone/FINANCIAL_ZONE/inbox/", zonePrefix)

	shardPrefix, err := MaterializeStatePrefix("zone/{zone_id}/shard/{shard_id}/metrics/*", ZoneIDContract, "0007")
	require.NoError(t, err)
	require.Equal(t, "zone/CONTRACT_ZONE/shard/0007/metrics/", shardPrefix)

	_, err = MaterializeStatePrefix("zone/{zone_id}/state/*", "bad-zone", "")
	require.ErrorContains(t, err, "zone id")

	_, err = MaterializeStatePrefix("zone/{zone_id}/shard/{shard_id}/state/*", ZoneIDIdentity, "")
	require.ErrorContains(t, err, "shard id")
}

func TestGlobalStatePrefixModelRejectsMalformedOwnershipAndTampering(t *testing.T) {
	duplicate, err := BuildGlobalStatePrefixModel([]GlobalStatePrefixDescriptor{
		GlobalStatePrefixDescriptors()[0],
		GlobalStatePrefixDescriptors()[0],
	})
	require.ErrorContains(t, err, "duplicate state prefix")
	require.Empty(t, duplicate.Root)

	_, err = BuildGlobalStatePrefixDescriptor(GlobalStatePrefixDescriptor{
		Namespace:	StatePrefixNamespaceCore,
		Prefix:		"zone/{zone_id}/state/*",
		Purpose:	"wrong owner",
		ProofScope:	"Core root",
	})
	require.ErrorContains(t, err, "core state prefix must start with core/")

	_, err = BuildGlobalStatePrefixDescriptor(GlobalStatePrefixDescriptor{
		Namespace:	StatePrefixNamespaceZone,
		Prefix:		"zone/{zone_id}/shard/{shard_id}/state/*",
		Purpose:	"wrong owner",
		ProofScope:	"Zone state root",
	})
	require.ErrorContains(t, err, "cannot contain shard placeholder")

	_, err = BuildGlobalStatePrefixDescriptor(GlobalStatePrefixDescriptor{
		Namespace:	StatePrefixNamespaceShard,
		Prefix:		"zone/{zone_id}/state/*",
		Purpose:	"wrong owner",
		ProofScope:	"Shard state root",
	})
	require.ErrorContains(t, err, "shard state prefix must start")

	_, err = BuildGlobalStatePrefixDescriptor(GlobalStatePrefixDescriptor{
		Namespace:	StatePrefixNamespaceCore,
		Prefix:		"core/*/bad",
		Purpose:	"bad wildcard",
		ProofScope:	"Core root",
	})
	require.ErrorContains(t, err, "wildcard must be terminal")

	tampered := GlobalStatePrefixDescriptors()[0]
	tampered.Purpose = strings.ReplaceAll(tampered.Purpose, "global", "local")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}
