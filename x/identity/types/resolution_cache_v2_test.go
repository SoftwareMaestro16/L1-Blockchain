package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolutionCacheRecordV2ValidatesFreshnessEpochsAndLightClientProof(t *testing.T) {
	pathHash, err := ComputeResolutionPathHashV2([]string{"api.alice.aet", "alice.aet"})
	require.NoError(t, err)
	unified := validUnifiedResolutionRecordV2(t, "alice.aet")
	recordHash, err := ComputeResolvedRecordHashV2(unified)
	require.NoError(t, err)

	record, err := NewResolutionCacheRecordV2("api.alice.aet", pathHash, recordHash, 100, 7, 2, 3)
	require.NoError(t, err)
	require.NoError(t, ValidateResolutionCacheRecordV2Use(record, ResolutionCacheUseContextV2{
		Height:		50,
		SourceVersion:	7,
		ParentEpoch:	2,
		ChildEpoch:	3,
	}))
	require.ErrorContains(t, ValidateResolutionCacheRecordV2Use(record, ResolutionCacheUseContextV2{
		Height:		50,
		SourceVersion:	7,
		ParentEpoch:	2,
		ChildEpoch:	3,
		LightClient:	true,
	}), "proof verification")
	require.NoError(t, ValidateResolutionCacheRecordV2Use(record, ResolutionCacheUseContextV2{
		Height:		50,
		SourceVersion:	7,
		ParentEpoch:	2,
		ChildEpoch:	3,
		LightClient:	true,
		ProofVerified:	true,
	}))
	require.ErrorContains(t, ValidateResolutionCacheRecordV2Use(record, ResolutionCacheUseContextV2{
		Height:		101,
		SourceVersion:	7,
		ParentEpoch:	2,
		ChildEpoch:	3,
	}), "expired")
	require.ErrorContains(t, ValidateResolutionCacheRecordV2Use(record, ResolutionCacheUseContextV2{
		Height:		50,
		SourceVersion:	8,
		ParentEpoch:	2,
		ChildEpoch:	3,
	}), "source version changed")
	require.ErrorContains(t, ValidateResolutionCacheRecordV2Use(record, ResolutionCacheUseContextV2{
		Height:		50,
		SourceVersion:	7,
		ParentEpoch:	9,
		ChildEpoch:	3,
	}), "parent epoch changed")
}

func TestResolutionCacheRecordV2InvalidatesOnDomainMutation(t *testing.T) {
	pathHash := identityHash("path")
	recordHash := identityHash("record")
	record, err := NewResolutionCacheRecordV2("alice.aet", pathHash, recordHash, 100, 7, 2, 3)
	require.NoError(t, err)

	invalidated := InvalidateResolutionCacheRecordV2ForDomainMutation(record, 8, 4, 5)
	require.Equal(t, uint64(0), invalidated.ValidUntilHeight)
	require.Equal(t, uint64(8), invalidated.SourceVersion)
	require.Equal(t, uint64(4), invalidated.ParentEpoch)
	require.Equal(t, uint64(5), invalidated.ChildEpoch)
	require.ErrorContains(t, ValidateResolutionCacheRecordV2(invalidated), "valid_until_height")
}

func validUnifiedResolutionRecordV2(t *testing.T, name string) UnifiedResolutionRecordV2 {
	t.Helper()
	nameHash, err := DomainRecordV2NameHash(name)
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:		nameHash,
		Owner:			addr(1),
		PrimaryAddress:		addr(2),
		RecordVersion:		1,
		RecordTTL:		30,
		UpdatedAtHeight:	12,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
		ContractTargets:	[]ContractTargetV2{NewContractTargetV2(ResolverKeyContract, addr(3), 12)},
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
	return record
}
