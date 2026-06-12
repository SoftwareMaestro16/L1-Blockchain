package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrossZoneIdentityBindingV2KeysMatchSpec(t *testing.T) {
	key, err := CrossZoneIdentityBindingV2Key("domain/alice", "APPLICATION_ZONE", IdentityBindingTargetService, "apps/app/alice")
	require.NoError(t, err)
	require.Equal(t, "identity/cross_zone/bindings/domain/alice/APPLICATION_ZONE/service/apps/app/alice", key)

	eventKey, err := CrossZoneBindingInvalidationV2Key(42, "event/alice")
	require.NoError(t, err)
	require.Equal(t, "identity/cross_zone/invalidations/00000000000000000042/event/alice", eventKey)
}

func TestCrossZoneIdentityBindingV2BindUpdateRevokeEmitsInvalidations(t *testing.T) {
	graph := testCrossZoneBindingGraph(t)
	binding := testCrossZoneIdentityBinding(t, graph, 100, 1, true)
	confirmation := testCrossZoneProofConfirmation(t, 12)
	state, err := BuildCrossZoneIdentityBindingStateV2(nil, nil, nil, 11)
	require.NoError(t, err)

	msg := MsgBindCrossZoneIdentityV2{
		Authority:	addr(1),
		Binding:	binding,
		Confirmation:	confirmation,
		Height:		12,
	}
	msg.MessageHash = ComputeMsgBindCrossZoneIdentityV2Hash(msg)
	state, event, err := ApplyCrossZoneIdentityBindV2(state, graph, msg)
	require.NoError(t, err)
	require.Equal(t, IdentityBindingInvalidatedCreated, event.Reason)
	require.Len(t, state.Bindings, 1)
	require.Len(t, state.Confirmations, 1)
	require.Len(t, state.Invalidations, 1)
	require.NoError(t, ValidateCrossZoneIdentityBindingForRoutingV2(binding, confirmation, 12))

	updated := testCrossZoneIdentityBinding(t, graph, 120, 2, false)
	messageConfirmation := testCrossZoneMessageConfirmation(t, 13)
	update := MsgUpdateCrossZoneIdentityBindingV2{
		Authority:	addr(1),
		Binding:	updated,
		Confirmation:	messageConfirmation,
		Height:		13,
	}
	update.MessageHash = ComputeMsgUpdateCrossZoneIdentityBindingV2Hash(update)
	state, event, err = ApplyCrossZoneIdentityBindingUpdateV2(state, graph, update)
	require.NoError(t, err)
	require.Equal(t, IdentityBindingInvalidatedUpdated, event.Reason)
	require.Equal(t, uint64(2), state.Bindings[0].BindingVersion)
	require.Len(t, state.Invalidations, 2)

	revoke := MsgRevokeCrossZoneIdentityBindingV2{
		Authority:	addr(1),
		IdentityID:	updated.IdentityID,
		TargetZone:	updated.TargetZone,
		TargetType:	updated.TargetType,
		TargetKey:	updated.TargetKey,
		Height:		14,
	}
	revoke.MessageHash = ComputeMsgRevokeCrossZoneIdentityBindingV2Hash(revoke)
	state, event, err = ApplyCrossZoneIdentityBindingRevokeV2(state, graph, revoke)
	require.NoError(t, err)
	require.Equal(t, IdentityBindingInvalidatedRevoked, event.Reason)
	require.Empty(t, state.Bindings)
	require.Len(t, state.Invalidations, 3)
	require.NoError(t, state.Validate())
}

func TestCrossZoneIdentityBindingV2RejectsUnauthorizedOwner(t *testing.T) {
	graph := testCrossZoneBindingGraph(t)
	binding := testCrossZoneIdentityBinding(t, graph, 100, 1, true)
	msg := MsgBindCrossZoneIdentityV2{
		Authority:	addr(9),
		Binding:	binding,
		Confirmation:	testCrossZoneProofConfirmation(t, 12),
		Height:		12,
	}
	_, _, err := ApplyCrossZoneIdentityBindV2(testEmptyCrossZoneBindingState(t), graph, msg)
	require.ErrorContains(t, err, "identity owner")
}

func TestCrossZoneIdentityBindingV2RequiresProofOrMessageConfirmation(t *testing.T) {
	graph := testCrossZoneBindingGraph(t)
	binding := testCrossZoneIdentityBinding(t, graph, 100, 1, true)
	msg := MsgBindCrossZoneIdentityV2{
		Authority:	addr(1),
		Binding:	binding,
		Confirmation:	testCrossZoneMessageConfirmation(t, 12),
		Height:		12,
	}
	_, _, err := ApplyCrossZoneIdentityBindV2(testEmptyCrossZoneBindingState(t), graph, msg)
	require.ErrorContains(t, err, "proof-verifiable")

	binding.ProofRequired = false
	binding.BindingHash = ""
	binding, err = NewCrossZoneIdentityBindingV2(binding)
	require.NoError(t, err)
	require.NoError(t, ValidateCrossZoneBindingConfirmationForBindingV2(binding, testCrossZoneMessageConfirmation(t, 12)))
}

func TestCrossZoneIdentityBindingV2ExpiredCannotRoute(t *testing.T) {
	graph := testCrossZoneBindingGraph(t)
	binding := testCrossZoneIdentityBinding(t, graph, 20, 1, false)
	confirmation := testCrossZoneMessageConfirmation(t, 12)
	require.NoError(t, ValidateCrossZoneIdentityBindingForRoutingV2(binding, confirmation, 19))
	err := ValidateCrossZoneIdentityBindingForRoutingV2(binding, confirmation, 20)
	require.ErrorContains(t, err, "expired")
}

func TestCrossZoneIdentityBindingV2StateRootDeterministic(t *testing.T) {
	graph := testCrossZoneBindingGraph(t)
	first := testCrossZoneIdentityBinding(t, graph, 100, 1, false)
	second := testCrossZoneIdentityBindingWithTarget(t, graph, "contract/alice", IdentityBindingTargetContract, 100, 1, false)
	firstConfirmation := testCrossZoneMessageConfirmation(t, 12)
	secondConfirmation := testCrossZoneProofConfirmation(t, 12)
	firstEvent, err := NewCrossZoneBindingInvalidationEventV2(CrossZoneBindingInvalidationEventV2{
		BindingHash:	first.BindingHash,
		IdentityID:	first.IdentityID,
		TargetZone:	first.TargetZone,
		TargetType:	first.TargetType,
		TargetKey:	first.TargetKey,
		Reason:		IdentityBindingInvalidatedCreated,
		Height:		12,
	})
	require.NoError(t, err)
	secondEvent, err := NewCrossZoneBindingInvalidationEventV2(CrossZoneBindingInvalidationEventV2{
		BindingHash:	second.BindingHash,
		IdentityID:	second.IdentityID,
		TargetZone:	second.TargetZone,
		TargetType:	second.TargetType,
		TargetKey:	second.TargetKey,
		Reason:		IdentityBindingInvalidatedCreated,
		Height:		12,
	})
	require.NoError(t, err)

	left, err := BuildCrossZoneIdentityBindingStateV2(
		[]CrossZoneIdentityBindingV2{second, first},
		[]CrossZoneBindingConfirmationV2{secondConfirmation, firstConfirmation},
		[]CrossZoneBindingInvalidationEventV2{secondEvent, firstEvent},
		12,
	)
	require.NoError(t, err)
	right, err := BuildCrossZoneIdentityBindingStateV2(
		[]CrossZoneIdentityBindingV2{first, second},
		[]CrossZoneBindingConfirmationV2{firstConfirmation, secondConfirmation},
		[]CrossZoneBindingInvalidationEventV2{firstEvent, secondEvent},
		12,
	)
	require.NoError(t, err)
	require.Equal(t, left.RootHash, right.RootHash)
}

func testCrossZoneBindingGraph(t *testing.T) IdentityGraphStateV2 {
	t.Helper()
	record := testIdentityGraphUnifiedRecord(t)
	outputs, err := BuildIdentityResolverOutputsFromUnifiedRecordV2(record, 12)
	require.NoError(t, err)
	graph, err := BuildIdentityGraphFromResolverOutputsV2(record.NameHash, record.Owner, outputs, 12)
	require.NoError(t, err)
	return graph
}

func testCrossZoneIdentityBinding(t *testing.T, graph IdentityGraphStateV2, expiresHeight uint64, version uint64, proofRequired bool) CrossZoneIdentityBindingV2 {
	t.Helper()
	return testCrossZoneIdentityBindingWithTarget(t, graph, "apps/app/alice", IdentityBindingTargetService, expiresHeight, version, proofRequired)
}

func testCrossZoneIdentityBindingWithTarget(t *testing.T, graph IdentityGraphStateV2, targetKey string, targetType IdentityCrossZoneBindingTargetType, expiresHeight uint64, version uint64, proofRequired bool) CrossZoneIdentityBindingV2 {
	t.Helper()
	identityID := ""
	for _, node := range graph.Nodes {
		if node.NodeType == IdentityGraphNodeDomain {
			identityID = node.IdentityID
			break
		}
	}
	require.NotEmpty(t, identityID)
	binding, err := NewCrossZoneIdentityBindingV2(CrossZoneIdentityBindingV2{
		IdentityID:	identityID,
		TargetZone:	"APPLICATION_ZONE",
		TargetType:	targetType,
		TargetKey:	targetKey,
		ProofRequired:	proofRequired,
		ExpiresHeight:	expiresHeight,
		BindingVersion:	version,
	})
	require.NoError(t, err)
	return binding
}

func testCrossZoneProofConfirmation(t *testing.T, height uint64) CrossZoneBindingConfirmationV2 {
	t.Helper()
	confirmation, err := NewCrossZoneBindingConfirmationV2(CrossZoneBindingConfirmationV2{
		ConfirmationType:	IdentityBindingConfirmationProof,
		ProofRoot:		identityHash("proof-root"),
		ProofHash:		identityHash("proof-hash"),
		ConfirmedHeight:	height,
	})
	require.NoError(t, err)
	return confirmation
}

func testCrossZoneMessageConfirmation(t *testing.T, height uint64) CrossZoneBindingConfirmationV2 {
	t.Helper()
	confirmation, err := NewCrossZoneBindingConfirmationV2(CrossZoneBindingConfirmationV2{
		ConfirmationType:	IdentityBindingConfirmationMessage,
		MessageID:		identityHash("message-id"),
		ReceiptHash:		identityHash("receipt-hash"),
		ConfirmedHeight:	height,
	})
	require.NoError(t, err)
	return confirmation
}

func testEmptyCrossZoneBindingState(t *testing.T) CrossZoneIdentityBindingStateV2 {
	t.Helper()
	state, err := BuildCrossZoneIdentityBindingStateV2(nil, nil, nil, 11)
	require.NoError(t, err)
	return state
}
