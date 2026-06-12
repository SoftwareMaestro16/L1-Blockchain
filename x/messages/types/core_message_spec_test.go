package types

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoreMessageTypeSpecCoversSectionTwelveOne(t *testing.T) {
	require.NoError(t, ValidateCoreMessageTypeSpec())

	spec, err := DefaultCoreMessageTypeSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Len(t, spec.Types, 12)
	require.NotEmpty(t, spec.Root)

	byType := map[CoreMessageType]CoreMessageTypeDescriptor{}
	for _, desc := range spec.Types {
		require.NoError(t, desc.Validate())
		byType[desc.MessageType] = desc
	}

	require.Equal(t, "Financial Zone", byType[CoreMsgZoneTransfer].PrimaryZone)
	require.Contains(t, byType[CoreMsgZoneTransfer].Purpose, "escrow")
	require.Equal(t, "Source and destination zones", byType[CoreMsgCrossZoneCall].PrimaryZone)
	require.Equal(t, "Owning zone", byType[CoreMsgShardCall].PrimaryZone)
	require.Equal(t, "Contract Zone", byType[CoreMsgContractCall].PrimaryZone)
	require.Equal(t, "Identity Zone", byType[CoreMsgIdentityLookup].PrimaryZone)
	require.Equal(t, "Requester zone", byType[CoreMsgIdentityLookupResult].PrimaryZone)
	require.Equal(t, "Financial Zone", byType[CoreMsgPaymentRoute].PrimaryZone)
	require.Equal(t, "Financial Zone", byType[CoreMsgPaymentSettle].PrimaryZone)
	require.Equal(t, "Destination promise owner", byType[CoreMsgPromiseResolve].PrimaryZone)
	require.Equal(t, "Destination promise owner", byType[CoreMsgPromiseTimeout].PrimaryZone)
	require.Equal(t, "Proof verifier target", byType[CoreMsgProofSubmit].PrimaryZone)
	require.Equal(t, "Source zone", byType[CoreMsgBounce].PrimaryZone)
}

func TestCoreMessageTypeSpecRootIsCanonicalAcrossInputOrder(t *testing.T) {
	spec, err := DefaultCoreMessageTypeSpec()
	require.NoError(t, err)

	reordered := append([]CoreMessageTypeDescriptor(nil), CoreMessageTypeDescriptors()...)
	slices.Reverse(reordered)
	reorderedSpec, err := BuildCoreMessageTypeSpec(reordered)
	require.NoError(t, err)
	require.Equal(t, spec.Root, reorderedSpec.Root)
	require.Equal(t, spec.Types, reorderedSpec.Types)
}

func TestCoreMessageTypeSpecRejectsMalformedDescriptors(t *testing.T) {
	duplicate, err := BuildCoreMessageTypeSpec([]CoreMessageTypeDescriptor{
		CoreMessageTypeDescriptors()[0],
		CoreMessageTypeDescriptors()[0],
	})
	require.ErrorContains(t, err, "duplicate core message type")
	require.Empty(t, duplicate.Root)

	_, err = BuildCoreMessageTypeDescriptor(CoreMessageTypeDescriptor{
		MessageType:	CoreMessageType("MsgUnknown"),
		Purpose:	"unknown",
		PrimaryZone:	"Core",
	})
	require.ErrorContains(t, err, "unknown core message type")

	tampered := CoreMessageTypeDescriptors()[0]
	tampered.Purpose = strings.ReplaceAll(tampered.Purpose, "Transfer", "Ignore")
	require.ErrorContains(t, tampered.Validate(), "descriptor hash mismatch")
}
