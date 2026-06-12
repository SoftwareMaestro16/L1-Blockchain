package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMNonceKeeperAndReplayTombstoneStore(t *testing.T) {
	state := testAVMReplayState(t, "alice", 1)
	keeper, err := NewAVMNonceKeeper(AVMNonceKeeper{})
	require.NoError(t, err)
	keeper, err = UpsertAVMNonceKeeperState(keeper, state)
	require.NoError(t, err)
	require.NoError(t, keeper.Validate())
	require.Equal(t, ComputeAVMNonceKeeperRoot(keeper), keeper.KeeperRoot)
	require.Len(t, keeper.States, 1)

	msg := testAVMSecurityMessage(t, "alice", 2, 10)
	nextState, consumed, err := ConsumeAVMReplayMessage(state, msg, 12)
	require.NoError(t, err)
	expiredMsg := testAVMSecurityMessage(t, "alice", 3, 10)
	nextState, expired, err := ExpireAVMReplayMessage(nextState, expiredMsg, expiredMsg.ExpiryHeight+1)
	require.NoError(t, err)

	store, err := NewAVMReplayTombstoneStore(AVMReplayTombstoneStore{
		ConsumedTombstones:	[]AVMAsyncReplayTombstone{consumed},
		ExpiredNonces:		[]AVMExpiredNonceTombstone{expired},
	})
	require.NoError(t, err)
	require.NoError(t, store.Validate())
	require.Equal(t, ComputeAVMReplayTombstoneStoreRoot(store), store.StoreRoot)

	keeper, err = UpsertAVMNonceKeeperState(keeper, nextState)
	require.NoError(t, err)
	require.Equal(t, uint64(3), keeper.States[0].LastNonce)
}

func TestAVMStateIsolationCapabilityChecks(t *testing.T) {
	capability, err := NewAVMZoneAccessCapability(AVMZoneAccessCapability{
		SourceZone:		zonestypes.ZoneIDApplication,
		ActorIDOptional:	"actor-a",
		ContractAddress:	"contract-a",
	})
	require.NoError(t, err)

	zoneWrite := testAVMStateAccessRequest(t, AVMStateAccessRequest{
		SourceZone:	zonestypes.ZoneIDApplication,
		TargetZone:	zonestypes.ZoneIDApplication,
		Mode:		AVMStateAccessWrite,
		Target:		AVMStateAccessTargetZone,
		StateKey:	AVMZoneRuntimeConfigKey(zonestypes.ZoneIDApplication),
	})
	require.NoError(t, ValidateAVMStateIsolationAccess(capability, zoneWrite))

	crossZoneWrite := zoneWrite
	crossZoneWrite.TargetZone = zonestypes.ZoneIDContract
	crossZoneWrite.StateKey = AVMZoneRuntimeConfigKey(zonestypes.ZoneIDContract)
	crossZoneWrite, err = NewAVMStateAccessRequest(crossZoneWrite)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateAVMStateIsolationAccess(capability, crossZoneWrite), "cross-zone writes")

	actorWrite := testAVMStateAccessRequest(t, AVMStateAccessRequest{
		SourceZone:		zonestypes.ZoneIDApplication,
		TargetZone:		zonestypes.ZoneIDApplication,
		Mode:			AVMStateAccessWrite,
		Target:			AVMStateAccessTargetActor,
		ActorIDOptional:	"actor-a",
		StateKey:		ActorStateKeyPrefix("actor-a") + "balance",
	})
	require.NoError(t, ValidateAVMStateIsolationAccess(capability, actorWrite))

	crossActorRead := testAVMStateAccessRequest(t, AVMStateAccessRequest{
		SourceZone:		zonestypes.ZoneIDApplication,
		TargetZone:		zonestypes.ZoneIDApplication,
		Mode:			AVMStateAccessRead,
		Target:			AVMStateAccessTargetActor,
		ActorIDOptional:	"actor-b",
		StateKey:		ActorStateKeyPrefix("actor-b") + "balance",
	})
	require.ErrorContains(t, ValidateAVMStateIsolationAccess(capability, crossActorRead), "read-only proof")

	crossActorRead.ProofHash = engineHash("actor-b-proof")
	crossActorRead.ReadOnlyProof = true
	crossActorRead, err = NewAVMStateAccessRequest(crossActorRead)
	require.NoError(t, err)
	require.NoError(t, ValidateAVMStateIsolationAccess(capability, crossActorRead))

	contractWrite := testAVMStateAccessRequest(t, AVMStateAccessRequest{
		SourceZone:		zonestypes.ZoneIDApplication,
		TargetZone:		zonestypes.ZoneIDApplication,
		Mode:			AVMStateAccessWrite,
		Target:			AVMStateAccessTargetContract,
		ContractAddress:	"contract-a",
		StateKey:		AVMContractStorageKey("contract-a", "balance"),
	})
	require.NoError(t, ValidateAVMStateIsolationAccess(capability, contractWrite))

	wrongContract := contractWrite
	wrongContract.ContractAddress = "contract-b"
	wrongContract.StateKey = AVMContractStorageKey("contract-b", "balance")
	wrongContract, err = NewAVMStateAccessRequest(wrongContract)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateAVMStateIsolationAccess(capability, wrongContract), "contract address")
}

func TestAVMSchedulerSafetyDelayedExpiredAndRetryBounds(t *testing.T) {
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	ready := testAVMSchedulerMessage(t, "alice", 1, 10, 0, 30, 5, 30)
	delayed := testAVMSchedulerMessage(t, "bob", 1, 10, 8, 40, 5, 30)
	expired := testAVMSchedulerMessage(t, "carol", 1, 10, 0, 12, 5, 30)
	for _, msg := range []AVMAsyncMessage{ready, delayed, expired} {
		queue, _, err = AdmitAVMZoneQueueMessage(queue, msg, 11, 10)
		require.NoError(t, err)
	}

	check, err := NewAVMSchedulerSafetyCheck(
		queue,
		[]AVMAsyncMessage{ready, delayed, expired},
		15,
		zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
		[]AVMSchedulerRetryBound{{MessageID: ready.ID, Attempt: 1, MaxAttempts: 3}},
		false,
	)
	require.NoError(t, err)
	require.NoError(t, check.Validate())
	require.Len(t, check.ReadyMessages, 1)
	require.Equal(t, ready.ID, check.ReadyMessages[0].ID)
	require.Len(t, check.ExpiredMessages, 1)
	require.Equal(t, expired.ID, check.ExpiredMessages[0].ID)
	require.Contains(t, check.RejectedEarlyMessageIDs, delayed.ID)
	require.Contains(t, check.RejectedExpiredExecutionIDs, expired.ID)

	badRetry := check
	badRetry.RetryBounds = []AVMSchedulerRetryBound{{MessageID: ready.ID, Attempt: 4, MaxAttempts: 3}}
	badRetry.SchedulerCheckHash = ComputeAVMSchedulerSafetyCheckHash(badRetry)
	require.ErrorContains(t, badRetry.Validate(), "retry count is bounded")

	early := check
	early.ReadyMessages = append(early.ReadyMessages, delayed)
	early.SchedulerCheckHash = ComputeAVMSchedulerSafetyCheckHash(early)
	require.ErrorContains(t, early.Validate(), "cannot execute early")
}

func TestAVMSchedulerSafetyPriorityCannotBypassSenderNonceOrdering(t *testing.T) {
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	nonceOne := testAVMSchedulerMessage(t, "same-sender", 1, 10, 0, 40, 1, 20)
	nonceTwoHighPriority := testAVMSchedulerMessage(t, "same-sender", 2, 10, 0, 40, 200, 20)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, nonceOne, 10, 10)
	require.NoError(t, err)
	queue, _, err = AdmitAVMZoneQueueMessage(queue, nonceTwoHighPriority, 10, 10)
	require.NoError(t, err)

	_, err = NewAVMSchedulerSafetyCheck(
		queue,
		[]AVMAsyncMessage{nonceOne, nonceTwoHighPriority},
		11,
		zonestypes.ZoneExecutionBudget{MaxGas: 100, MaxMessages: 10},
		nil,
		true,
	)
	require.ErrorContains(t, err, "nonce ordering")
}

func TestAVMSchedulerSafetyMaliciousQueueLoadIsBounded(t *testing.T) {
	queue, err := NewAVMZoneQueue(AVMZoneQueue{ZoneID: zonestypes.ZoneIDContract})
	require.NoError(t, err)
	messages := make([]AVMAsyncMessage, 0, 25)
	for i := 0; i < 25; i++ {
		msg := testAVMSchedulerMessage(t, fmt.Sprintf("sender-%02d", i), uint64(i+1), 10, 0, 40, uint8(i%10), 30)
		messages = append(messages, msg)
		queue, _, err = AdmitAVMZoneQueueMessage(queue, msg, 10, 100)
		require.NoError(t, err)
	}
	check, err := NewAVMSchedulerSafetyCheck(
		queue,
		messages,
		11,
		zonestypes.ZoneExecutionBudget{MaxGas: 90, MaxMessages: 3},
		nil,
		false,
	)
	require.NoError(t, err)
	require.NoError(t, check.Validate())
	require.Len(t, check.ReadyMessages, 3)
	require.Equal(t, uint64(90), check.Budget.GasUsed)
	require.Equal(t, uint32(3), check.Budget.MessagesUsed)
}

func testAVMStateAccessRequest(t *testing.T, request AVMStateAccessRequest) AVMStateAccessRequest {
	t.Helper()
	built, err := NewAVMStateAccessRequest(request)
	require.NoError(t, err)
	return built
}

func testAVMSchedulerMessage(t *testing.T, sender string, nonce, createdHeight, delayHeight, expiryHeight uint64, priority uint8, gasLimit uint64) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage(sender, zonestypes.ZoneIDApplication, "contract", zonestypes.ZoneIDContract, nonce, createdHeight)
	msg.DelayHeight = delayHeight
	msg.ExpiryHeight = expiryHeight
	msg.RetryPolicy = DefaultAVMRetryPolicy(expiryHeight)
	msg.Priority = priority
	msg.GasLimit = gasLimit
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}
