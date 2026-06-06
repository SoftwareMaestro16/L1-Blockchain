package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMStateModelDeclaresDesignStorePrefixes(t *testing.T) {
	model := DefaultAVMStateModel()
	require.NoError(t, model.Validate())
	require.Equal(t, ComputeAVMStateModelRoot(model), model.Root)

	byPrefix := make(map[string]AVMStatePrefixDescriptor, len(model.Prefixes))
	for _, descriptor := range model.Prefixes {
		byPrefix[descriptor.Prefix] = descriptor
	}
	require.Equal(t, AVMStateValueParams, byPrefix[AVMStatePrefixParams].ValueType)
	require.Equal(t, "avm/zones/{zone_id}", byPrefix[AVMStatePrefixZones].KeyTemplate)
	require.Equal(t, AVMStateValueRouteDescriptor, byPrefix[AVMStatePrefixRouterRoutes].ValueType)
	require.Equal(t, AVMStateValueAsyncMessage, byPrefix[AVMStatePrefixAsyncMessages].ValueType)
	require.Equal(t, AVMStateValueMessageID, byPrefix[AVMStatePrefixAsyncQueues].ValueType)
	require.Equal(t, AVMStateValueRetryRecord, byPrefix[AVMStatePrefixAsyncRetry].ValueType)
	require.Equal(t, AVMStateValueDeadLetterRecord, byPrefix[AVMStatePrefixAsyncDead].ValueType)
	require.Equal(t, AVMStateValueActorRecord, byPrefix[AVMStatePrefixActors].ValueType)
	require.Equal(t, AVMStateValueMessageID, byPrefix[AVMStatePrefixActorMailbox].ValueType)
	require.Equal(t, AVMStateValueContinuation, byPrefix[AVMStatePrefixContinuations].ValueType)
	require.Equal(t, AVMStateValueCodeRecord, byPrefix[AVMStatePrefixContractCode].ValueType)
	require.Equal(t, AVMStateValueContractRecord, byPrefix[AVMStatePrefixContractInstances].ValueType)
	require.Equal(t, AVMStateValueStorageValue, byPrefix[AVMStatePrefixContractStorage].ValueType)
	require.Equal(t, AVMStateValueInterface, byPrefix[AVMStatePrefixInterfaces].ValueType)
	require.Equal(t, AVMStateValueReceipt, byPrefix[AVMStatePrefixReceipts].ValueType)
	require.Equal(t, AVMStateValueRoot, byPrefix[AVMStatePrefixRoots].ValueType)

	shuffled := AVMStateModel{Prefixes: append([]AVMStatePrefixDescriptor(nil), model.Prefixes...)}
	shuffled.Prefixes[0], shuffled.Prefixes[len(shuffled.Prefixes)-1] = shuffled.Prefixes[len(shuffled.Prefixes)-1], shuffled.Prefixes[0]
	shuffled.Root = ComputeAVMStateModelRoot(shuffled)
	require.Equal(t, model.Root, shuffled.Root)
	require.NoError(t, shuffled.Validate())
}

func TestAVMStateKeysMatchDesignTemplates(t *testing.T) {
	zoneID := zonestypes.ZoneIDContract
	require.Equal(t, "avm/params", AVMParamsKey())
	require.Equal(t, "avm/zones/CONTRACT_ZONE", AVMZoneRuntimeConfigKey(zoneID))
	require.Equal(t, "avm/router/routes/contract.execute", AVMRouteDescriptorKey("contract.execute"))
	require.Equal(t, "avm/async/messages/msg-1", AVMAsyncMessageKey("msg-1"))
	require.Equal(t, "avm/async/queues/CONTRACT_ZONE/default/0001", AVMAsyncQueueKey(zoneID, "default", "0001"))
	require.Equal(t, "avm/async/retry/CONTRACT_ZONE/00000000000000000123/msg-1", AVMAsyncRetryKey(zoneID, 123, "msg-1"))
	require.Equal(t, "avm/async/dead/CONTRACT_ZONE/msg-1", AVMAsyncDeadLetterKey(zoneID, "msg-1"))
	require.Equal(t, "avm/actors/actor-1", AVMActorRecordKey("actor-1"))
	require.Equal(t, "avm/actors/mailbox/actor-1/0001", AVMActorMailboxKey("actor-1", "0001"))
	require.Equal(t, "avm/continuations/cont-1", AVMContinuationKey("cont-1"))
	require.Equal(t, "avm/contracts/code/00000000000000000007", AVMContractCodeKey(7))
	require.Equal(t, "avm/contracts/instances/contract-1", AVMContractInstanceKey("contract-1"))
	require.Equal(t, "avm/contracts/storage/contract-1/balance", AVMContractStorageKey("contract-1", "balance"))
	require.Equal(t, "avm/interfaces/iface-hash", AVMInterfaceDescriptorKey("iface-hash"))
	require.Equal(t, "avm/receipts/receipt-1", AVMReceiptKey("receipt-1"))
	require.Equal(t, "avm/roots/00000000000000000077", AVMRootKey(77))
}

func TestAVMStateModelRejectsMissingDuplicateAndMalformedPrefixes(t *testing.T) {
	model := DefaultAVMStateModel()
	missing := model
	missing.Prefixes = append([]AVMStatePrefixDescriptor(nil), model.Prefixes[1:]...)
	missing.Root = ComputeAVMStateModelRoot(missing)
	require.ErrorContains(t, missing.Validate(), "missing prefix")

	duplicate := model
	duplicate.Prefixes = append([]AVMStatePrefixDescriptor(nil), model.Prefixes...)
	duplicate.Prefixes = append(duplicate.Prefixes, duplicate.Prefixes[0])
	duplicate.Root = ComputeAVMStateModelRoot(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	malformed := model
	malformed.Prefixes = append([]AVMStatePrefixDescriptor(nil), model.Prefixes...)
	malformed.Prefixes[0].Prefix = "bad/params"
	malformed.Root = ComputeAVMStateModelRoot(malformed)
	require.ErrorContains(t, malformed.Validate(), "avm/")
}

func TestAVMZoneRuntimeConfigModelsBudgetsFiltersAndRootPrefixes(t *testing.T) {
	config, err := DefaultAVMZoneRuntimeConfig(zonestypes.ZoneIDApplication)
	require.NoError(t, err)
	require.NoError(t, config.Validate())
	require.True(t, config.Enabled)
	require.Equal(t, "avm/zones/APPLICATION_ZONE", AVMZoneRuntimeConfigKey(zonestypes.ZoneIDApplication))
	require.Equal(t, "avm/roots/state/APPLICATION_ZONE/", config.StateRootPrefix)
	require.Equal(t, "avm/roots/messages/APPLICATION_ZONE/", config.MessageRootPrefix)
	require.Equal(t, "avm/roots/continuations/APPLICATION_ZONE/", config.ContinuationRootPrefix)
	require.Equal(t, ComputeAVMZoneRuntimeConfigRoot(config), config.ConfigRoot)

	mutated := config
	mutated.MaxQueueDepth++
	require.NotEqual(t, config.ConfigRoot, ComputeAVMZoneRuntimeConfigRoot(mutated))
}

func TestAVMZoneRuntimeConfigRejectsInvalidBudgetsFiltersAndRootDrift(t *testing.T) {
	config, err := DefaultAVMZoneRuntimeConfig(zonestypes.ZoneIDContract)
	require.NoError(t, err)

	zeroAsync := config
	zeroAsync.AsyncBudgetPerBlock.MaxGas = 0
	zeroAsync.ConfigRoot = ComputeAVMZoneRuntimeConfigRoot(zeroAsync)
	require.ErrorContains(t, zeroAsync.Validate(), "async budget")

	badFilter := config
	badFilter.AllowedMessageTypes = []string{"contract.execute", "contract.execute"}
	badFilter.ConfigRoot = ComputeAVMZoneRuntimeConfigRoot(badFilter)
	require.ErrorContains(t, badFilter.Validate(), "duplicate")

	badPrefix := config
	badPrefix.StateRootPrefix = AVMZoneStateRootPrefix(zonestypes.ZoneIDFinancial)
	badPrefix.ConfigRoot = ComputeAVMZoneRuntimeConfigRoot(badPrefix)
	require.ErrorContains(t, badPrefix.Validate(), "state root prefix")

	oversizedMessage := config
	oversizedMessage.MaxMessageBytes = MaxAVMMessageBytes + 1
	oversizedMessage.ConfigRoot = ComputeAVMZoneRuntimeConfigRoot(oversizedMessage)
	require.ErrorContains(t, oversizedMessage.Validate(), "max message bytes")

	rootMismatch := config
	rootMismatch.ConfigRoot = engineHash("wrong-config-root")
	require.ErrorContains(t, rootMismatch.Validate(), "root mismatch")
}
