package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestUnifiedMessageLifecycleCommitsRouteAndFinality(t *testing.T) {
	params := testMessageParams()
	msg := testMessage(t, params, 31, 41, []byte("contract-call"))
	object, err := NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:		MessageCapabilityContractCall,
		TraceID:		"trace-contract-31",
		ExecutionMode:		UnifiedExecutionPromise,
		OrderingClass:		UnifiedOrderingStrictTraceOrder,
		SourceShardID:		"financial-1",
		DestinationShardID:	"contract-7",
	}, params)
	require.NoError(t, err)
	require.Equal(t, MessageLifecycleCreated, object.LifecycleStage)
	require.NoError(t, zonestypes.ValidateHash("object hash", object.ObjectHash))

	routed, err := CommitUnifiedMessageRoute(object, UnifiedMessageRoute{
		SourceShardID:		"financial-1",
		DestinationShardID:	"contract-7",
		ModuleRoute:		"contract/execute",
		CommittedHeight:	50,
		FinalityDelay:		2,
	}, params)
	require.NoError(t, err)
	require.Equal(t, MessageLifecycleRouted, routed.LifecycleStage)
	require.Equal(t, uint64(52), routed.Route.DeliveryEligibleFrom)
	require.NoError(t, routed.Route.ValidateHash())
	require.NoError(t, routed.ValidateHash(params))
}

func TestUnifiedMessageLifecycleRootIsCanonical(t *testing.T) {
	params := testMessageParams()
	msg := testMessage(t, params, 32, 42, []byte("lifecycle"))
	object, err := NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:		MessageCapabilityCrossZone,
		TraceID:		"trace-cross-zone-32",
		ExecutionMode:		UnifiedExecutionAsync,
		OrderingClass:		UnifiedOrderingSenderOrdered,
		SourceShardID:		"financial-1",
		DestinationShardID:	"contract-7",
	}, params)
	require.NoError(t, err)
	object, err = CommitUnifiedMessageRoute(object, UnifiedMessageRoute{
		SourceShardID:		"financial-1",
		DestinationShardID:	"contract-7",
		CommittedHeight:	60,
		FinalityDelay:		1,
	}, params)
	require.NoError(t, err)

	outbox, err := BuildMessageLifecycleRecord(object, MessageLifecycleOutbox, 60, MessageReceipt{}, params)
	require.NoError(t, err)
	core, err := BuildMessageLifecycleRecord(object, MessageLifecycleCore, 61, MessageReceipt{}, params)
	require.NoError(t, err)
	inbox, err := BuildMessageLifecycleRecord(object, MessageLifecycleInbox, 62, MessageReceipt{}, params)
	require.NoError(t, err)
	executed, err := BuildMessageLifecycleRecord(object, MessageLifecycleExecuted, 63, MessageReceipt{}, params)
	require.NoError(t, err)

	receipt, err := NewMessageReceipt(ReceiptFromMessage(msg, MessageStatusExecuted, 99, sdkmath.NewInt(3), ComputePayloadHash([]byte("return")), nil, 63))
	require.NoError(t, err)
	receiptRecord, err := BuildMessageLifecycleRecord(object, MessageLifecycleReceipt, 63, receipt, params)
	require.NoError(t, err)

	rootA, err := ComputeMessageLifecycleRoot([]MessageLifecycleRecord{receiptRecord, inbox, outbox, executed, core})
	require.NoError(t, err)
	rootB, err := ComputeMessageLifecycleRoot([]MessageLifecycleRecord{outbox, core, inbox, executed, receiptRecord})
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)
	require.NoError(t, receiptRecord.Validate())
}

func TestUnifiedMessageCapabilitiesRequireExplicitMetadata(t *testing.T) {
	params := testMessageParams()
	msg := testMessage(t, params, 33, 43, []byte("module"))

	_, err := NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:	MessageCapabilityModule,
		TraceID:	"trace-module-33",
		ExecutionMode:	UnifiedExecutionAsync,
		OrderingClass:	UnifiedOrderingObjectOrdered,
	}, params)
	require.ErrorContains(t, err, "module route")

	module, err := NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:	MessageCapabilityModule,
		TraceID:	"trace-module-33",
		ExecutionMode:	UnifiedExecutionAsync,
		OrderingClass:	UnifiedOrderingObjectOrdered,
		ModuleRoute:	"identity/resolve",
	}, params)
	require.NoError(t, err)
	require.Equal(t, "identity/resolve", module.Route.ModuleRoute)

	_, err = NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:	MessageCapabilityProofRead,
		TraceID:	"trace-proof-33",
		ExecutionMode:	UnifiedExecutionDeferred,
		OrderingClass:	UnifiedOrderingReceiverOrdered,
	}, params)
	require.ErrorContains(t, err, "proof")

	proofRead, err := NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:	MessageCapabilityProofRead,
		TraceID:	"trace-proof-33",
		ExecutionMode:	UnifiedExecutionDeferred,
		OrderingClass:	UnifiedOrderingReceiverOrdered,
		StateProofHash:	EmptyHash(),
	}, params)
	require.NoError(t, err)
	require.Equal(t, EmptyHash(), proofRead.StateProofHash)

	_, err = NewUnifiedMessageObject(msg, UnifiedMessageMetadata{
		Capability:	MessageCapabilityCrossShard,
		TraceID:	"trace-shard-33",
		ExecutionMode:	UnifiedExecutionAsync,
		OrderingClass:	UnifiedOrderingSenderOrdered,
		SourceShardID:	"financial-1",
	}, params)
	require.ErrorContains(t, err, "shards")
}
