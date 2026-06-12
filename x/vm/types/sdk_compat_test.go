package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetravm/avm"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMSDKBindingRequiresFinalizeBlockKeeperMsgServerStoreV2AndFinality(t *testing.T) {
	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	require.NoError(t, binding.Validate())
	require.Contains(t, binding.Lifecycle, LifecycleStage(LifecycleFinalizeBlock))
	require.Contains(t, binding.StateAccess, StateAccessKind(StateAccessStoreV2))
	require.Contains(t, binding.StateAccess, StateAccessKind(StateAccessKVStore))
	require.NotEmpty(t, binding.KeeperName)
	require.NotEmpty(t, binding.MsgServerName)
	require.True(t, binding.StakingFinalityRequired)
	require.True(t, binding.CometBFTFinalityRequired)

	missingFinalize := binding
	missingFinalize.Lifecycle = []LifecycleStage{LifecycleCommit}
	require.ErrorContains(t, missingFinalize.Validate(), "FinalizeBlock")

	missingStoreV2 := binding
	missingStoreV2.StateAccess = []StateAccessKind{StateAccessKVStore}
	require.ErrorContains(t, missingStoreV2.Validate(), "Store v2")

	missingFinality := binding
	missingFinality.StakingFinalityRequired = false
	require.ErrorContains(t, missingFinality.Validate(), "staking")
}

func TestAVMKVStoreLayoutIsZoneScopedAndKVCompatible(t *testing.T) {
	layout := AVMKVStoreLayout(zonestypes.ZoneIDContract)
	require.Len(t, layout, 5)
	for _, entry := range layout {
		require.Equal(t, DefaultAVMStoreKey, entry.StoreKey)
		require.Contains(t, entry.Prefix, string(zonestypes.ZoneIDContract))
		require.Contains(t, entry.Prefix, "vm/")
		require.NotEmpty(t, entry.Purpose)
	}

	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	binding.KVLayout[0].Prefix = binding.KVLayout[1].Prefix
	require.ErrorContains(t, binding.Validate(), "duplicate KV layout prefix")
}

func TestFinalizeBlockPlanSortsDispatchesAndRejectsBlockSTMConflictOverlap(t *testing.T) {
	policy := DefaultRuntimePolicy()
	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	dispatches := []SDKDispatch{
		testSDKDispatch(zonestypes.ZoneIDContract, "/aetra.avm.v1.MsgExecute", "vm/CONTRACT_ZONE/contract/b"),
		testSDKDispatch(zonestypes.ZoneIDApplication, "/l1.aetravm.async.v1.MsgDeliver", "vm/APPLICATION_ZONE/queue/a"),
	}

	plan, err := BuildFinalizeBlockPlan(100, binding, dispatches, policy)
	require.NoError(t, err)
	require.NoError(t, plan.Validate(policy))
	require.Equal(t, zonestypes.ZoneIDApplication, plan.Dispatches[0].ZoneID)
	require.Equal(t, uint64(100), plan.Dispatches[0].ExecutionHeight)

	conflicting := []SDKDispatch{
		testSDKDispatch(zonestypes.ZoneIDContract, "/aetra.avm.v1.MsgExecute", "vm/CONTRACT_ZONE/contract/a"),
		testSDKDispatch(zonestypes.ZoneIDContract, "/aetra.avm.v1.MsgExecute", "vm/CONTRACT_ZONE/contract/a/storage"),
	}
	_, err = BuildFinalizeBlockPlan(100, binding, conflicting, policy)
	require.ErrorContains(t, err, "BlockSTM conflict")
}

func TestFinalizeBlockPlanRejectsInvalidLifecycleHeightAndStakingPower(t *testing.T) {
	policy := DefaultRuntimePolicy()
	binding := DefaultAVMSDKBinding(zonestypes.ZoneIDContract)
	dispatch := testSDKDispatch(zonestypes.ZoneIDContract, "/aetra.avm.v1.MsgExecute", "vm/CONTRACT_ZONE/contract/a")

	_, err := BuildFinalizeBlockPlan(0, binding, []SDKDispatch{dispatch}, policy)
	require.ErrorContains(t, err, "height")

	dispatch.StakingPower = 0
	_, err = BuildFinalizeBlockPlan(1, binding, []SDKDispatch{dispatch}, policy)
	require.ErrorContains(t, err, "staking voting power")
}

func testSDKDispatch(zoneID zonestypes.ZoneID, msgType string, blockSTMKey string) SDKDispatch {
	return SDKDispatch{
		ZoneID:		zoneID,
		MsgType:	msgType,
		KVPrefix:	ContractZoneKVPrefix(zoneID),
		BlockSTMKey:	blockSTMKey,
		StakingPower:	1,
		Call: VMCall{
			Runtime:	RuntimeAVM,
			Action:		ActionExternalCall,
			GasLimit:	1,
			Entrypoint:	avm.EntryReceiveExternal,
		},
	}
}
