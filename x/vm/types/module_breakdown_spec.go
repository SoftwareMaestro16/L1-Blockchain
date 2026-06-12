package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	AVMModulePathActors		AVMCosmosModulePath	= "x/actors"
	AVMModulePathAVM		AVMCosmosModulePath	= "x/avm"
	AVMModulePathAVMContracts	AVMCosmosModulePath	= "x/avmcontracts"
	AVMModulePathAVMInterfaces	AVMCosmosModulePath	= "x/avminterfaces"
	AVMModulePathAsync		AVMCosmosModulePath	= "x/async"
	AVMModulePathContinuations	AVMCosmosModulePath	= "x/continuations"

	AVMModuleStateAVMParams		AVMModuleStateObject	= "AVMParams"
	AVMModuleStateRouteDescriptor	AVMModuleStateObject	= "RouteDescriptor"
	AVMModuleStateAVMRoot		AVMModuleStateObject	= "AVMRoot"
	AVMModuleStateExecutionReceipt	AVMModuleStateObject	= "ExecutionReceipt"
	AVMModuleStateRuntimeVersion	AVMModuleStateObject	= "RuntimeVersion"

	AVMModuleStateAsyncMessage	AVMModuleStateObject	= "AsyncMessage"
	AVMModuleStateZoneQueue		AVMModuleStateObject	= "ZoneQueue"
	AVMModuleStateRetryRecord	AVMModuleStateObject	= "RetryRecord"
	AVMModuleStateDeadLetter	AVMModuleStateObject	= "DeadLetterRecord"
	AVMModuleStateReplayTombstone	AVMModuleStateObject	= "ReplayTombstone"

	AVMModuleStateActorRecord	AVMModuleStateObject	= "ActorRecord"
	AVMModuleStateActorMailbox	AVMModuleStateObject	= "ActorMailbox"
	AVMModuleStateActorState	AVMModuleStateObject	= "ActorState"
	AVMModuleStateActorPermission	AVMModuleStateObject	= "ActorPermission"

	AVMModuleStateContinuation		AVMModuleStateObject	= "Continuation"
	AVMModuleStateContinuationQueue		AVMModuleStateObject	= "ContinuationQueue"
	AVMModuleStateContinuationReceipt	AVMModuleStateObject	= "ContinuationReceipt"

	AVMModuleStateCodeRecord	AVMModuleStateObject	= "CodeRecord"
	AVMModuleStateContractRecord	AVMModuleStateObject	= "ContractRecord"
	AVMModuleStateStorageValue	AVMModuleStateObject	= "StorageValue"
	AVMModuleStateBackendConfig	AVMModuleStateObject	= "BackendConfig"

	AVMModuleStateInterfaceDescriptor	AVMModuleStateObject	= "InterfaceDescriptor"
	AVMModuleStateMethodDescriptor		AVMModuleStateObject	= "MethodDescriptor"
	AVMModuleStateEventDescriptor		AVMModuleStateObject	= "EventDescriptor"
	AVMModuleStateAsyncHandlerDescriptor	AVMModuleStateObject	= "AsyncHandlerDescriptor"

	AVMModuleMsgSubmitAVMMessage		AVMModuleMessageName	= "MsgSubmitAVMMessage"
	AVMModuleMsgRegisterRoute		AVMModuleMessageName	= "MsgRegisterRoute"
	AVMModuleMsgUpdateAVMParams		AVMModuleMessageName	= "MsgUpdateAVMParams"
	AVMModuleMsgScheduleRuntimeUpgrade	AVMModuleMessageName	= "MsgScheduleRuntimeUpgrade"

	AVMModuleMsgSubmitAsyncMessage	AVMModuleMessageName	= "MsgSubmitAsyncMessage"
	AVMModuleMsgCancelAsyncMessage	AVMModuleMessageName	= "MsgCancelAsyncMessage"
	AVMModuleMsgRetryAsyncMessage	AVMModuleMessageName	= "MsgRetryAsyncMessage"
	AVMModuleMsgExpireAsyncMessage	AVMModuleMessageName	= "MsgExpireAsyncMessage"

	AVMModuleMsgCreateActor		AVMModuleMessageName	= "MsgCreateActor"
	AVMModuleMsgSendActorMessage	AVMModuleMessageName	= "MsgSendActorMessage"
	AVMModuleMsgUpdateActor		AVMModuleMessageName	= "MsgUpdateActor"
	AVMModuleMsgPauseActor		AVMModuleMessageName	= "MsgPauseActor"

	AVMModuleMsgCreateContinuation	AVMModuleMessageName	= "MsgCreateContinuation"
	AVMModuleMsgResumeContinuation	AVMModuleMessageName	= "MsgResumeContinuation"
	AVMModuleMsgCancelContinuation	AVMModuleMessageName	= "MsgCancelContinuation"
	AVMModuleMsgExpireContinuation	AVMModuleMessageName	= "MsgExpireContinuation"

	AVMModuleMsgStoreCode		AVMModuleMessageName	= "MsgStoreCode"
	AVMModuleMsgInstantiateContract	AVMModuleMessageName	= "MsgInstantiateContract"
	AVMModuleMsgExecuteContract	AVMModuleMessageName	= "MsgExecuteContract"
	AVMModuleMsgMigrateContract	AVMModuleMessageName	= "MsgMigrateContract"

	AVMModuleMsgRegisterInterface	AVMModuleMessageName	= "MsgRegisterInterface"
	AVMModuleMsgUpdateInterface	AVMModuleMessageName	= "MsgUpdateInterface"
	AVMModuleMsgDeprecateInterface	AVMModuleMessageName	= "MsgDeprecateInterface"

	AVMModuleQueryAVMParams		AVMModuleQueryName	= "QueryAVMParams"
	AVMModuleQueryAVMRoot		AVMModuleQueryName	= "QueryAVMRoot"
	AVMModuleQueryRoute		AVMModuleQueryName	= "QueryRoute"
	AVMModuleQueryExecutionReceipt	AVMModuleQueryName	= "QueryExecutionReceipt"
	AVMModuleQueryRuntimeVersion	AVMModuleQueryName	= "QueryRuntimeVersion"

	AVMModuleQueryAsyncMessage	AVMModuleQueryName	= "QueryAsyncMessage"
	AVMModuleQueryZoneQueue		AVMModuleQueryName	= "QueryZoneQueue"
	AVMModuleQueryDeadLetter	AVMModuleQueryName	= "QueryDeadLetter"
	AVMModuleQueryReplayTombstone	AVMModuleQueryName	= "QueryReplayTombstone"

	AVMModuleQueryActor		AVMModuleQueryName	= "QueryActor"
	AVMModuleQueryActorMailbox	AVMModuleQueryName	= "QueryActorMailbox"
	AVMModuleQueryActorState	AVMModuleQueryName	= "QueryActorState"

	AVMModuleQueryContinuation		AVMModuleQueryName	= "QueryContinuation"
	AVMModuleQueryContinuationsByActor	AVMModuleQueryName	= "QueryContinuationsByActor"
	AVMModuleQueryContinuationReceipt	AVMModuleQueryName	= "QueryContinuationReceipt"

	AVMModuleQueryCode		AVMModuleQueryName	= "QueryCode"
	AVMModuleQueryContract		AVMModuleQueryName	= "QueryContract"
	AVMModuleQueryContractStorage	AVMModuleQueryName	= "QueryContractStorage"
	AVMModuleQueryContractProof	AVMModuleQueryName	= "QueryContractProof"

	AVMModuleQueryInterface		AVMModuleQueryName	= "QueryInterface"
	AVMModuleQueryMethod		AVMModuleQueryName	= "QueryMethod"
	AVMModuleQueryInterfaceByTarget	AVMModuleQueryName	= "QueryInterfaceByTarget"

	MaxAVMModulePurposeItems	= 16
	MaxAVMModuleSurfaceItems	= 32
	MaxAVMModulePurposeText		= 128
)

type AVMCosmosModulePath string
type AVMModuleStateObject string
type AVMModuleMessageName string
type AVMModuleQueryName string

type AVMCosmosModuleBreakdown struct {
	ModulePath	AVMCosmosModulePath
	Purpose		[]string
	StateObjects	[]AVMModuleStateObject
	Messages	[]AVMModuleMessageName
	Queries		[]AVMModuleQueryName
	BreakdownHash	string
}

type AVMCosmosModuleRegistry struct {
	Modules		[]AVMCosmosModuleBreakdown
	RegistryHash	string
}

func DefaultAVMCosmosModuleRegistry() (AVMCosmosModuleRegistry, error) {
	avm, err := DefaultXAVMModuleBreakdown()
	if err != nil {
		return AVMCosmosModuleRegistry{}, err
	}
	async, err := DefaultXAsyncModuleBreakdown()
	if err != nil {
		return AVMCosmosModuleRegistry{}, err
	}
	actors, err := DefaultXActorsModuleBreakdown()
	if err != nil {
		return AVMCosmosModuleRegistry{}, err
	}
	contracts, err := DefaultXAVMContractsModuleBreakdown()
	if err != nil {
		return AVMCosmosModuleRegistry{}, err
	}
	interfaces, err := DefaultXAVMInterfacesModuleBreakdown()
	if err != nil {
		return AVMCosmosModuleRegistry{}, err
	}
	continuations, err := DefaultXContinuationsModuleBreakdown()
	if err != nil {
		return AVMCosmosModuleRegistry{}, err
	}
	return NewAVMCosmosModuleRegistry(AVMCosmosModuleRegistry{
		Modules: []AVMCosmosModuleBreakdown{avm, async, actors, contracts, interfaces, continuations},
	})
}

func DefaultXAVMModuleBreakdown() (AVMCosmosModuleBreakdown, error) {
	return NewAVMCosmosModuleBreakdown(AVMCosmosModuleBreakdown{
		ModulePath:	AVMModulePathAVM,
		Purpose: []string{
			"execution_receipts",
			"roots",
			"routing",
			"runtime_parameters",
			"runtime_versions",
		},
		StateObjects: []AVMModuleStateObject{
			AVMModuleStateAVMParams,
			AVMModuleStateRouteDescriptor,
			AVMModuleStateAVMRoot,
			AVMModuleStateExecutionReceipt,
			AVMModuleStateRuntimeVersion,
		},
		Messages: []AVMModuleMessageName{
			AVMModuleMsgSubmitAVMMessage,
			AVMModuleMsgRegisterRoute,
			AVMModuleMsgUpdateAVMParams,
			AVMModuleMsgScheduleRuntimeUpgrade,
		},
		Queries: []AVMModuleQueryName{
			AVMModuleQueryAVMParams,
			AVMModuleQueryAVMRoot,
			AVMModuleQueryRoute,
			AVMModuleQueryExecutionReceipt,
			AVMModuleQueryRuntimeVersion,
		},
	})
}

func DefaultXAsyncModuleBreakdown() (AVMCosmosModuleBreakdown, error) {
	return NewAVMCosmosModuleBreakdown(AVMCosmosModuleBreakdown{
		ModulePath:	AVMModulePathAsync,
		Purpose: []string{
			"async_message_queues",
			"dead_letter_queue",
			"delayed_queue",
			"retry_queue",
			"replay_tombstones",
		},
		StateObjects: []AVMModuleStateObject{
			AVMModuleStateAsyncMessage,
			AVMModuleStateZoneQueue,
			AVMModuleStateRetryRecord,
			AVMModuleStateDeadLetter,
			AVMModuleStateReplayTombstone,
		},
		Messages: []AVMModuleMessageName{
			AVMModuleMsgSubmitAsyncMessage,
			AVMModuleMsgCancelAsyncMessage,
			AVMModuleMsgRetryAsyncMessage,
			AVMModuleMsgExpireAsyncMessage,
		},
		Queries: []AVMModuleQueryName{
			AVMModuleQueryAsyncMessage,
			AVMModuleQueryZoneQueue,
			AVMModuleQueryDeadLetter,
			AVMModuleQueryReplayTombstone,
		},
	})
}

func DefaultXActorsModuleBreakdown() (AVMCosmosModuleBreakdown, error) {
	return NewAVMCosmosModuleBreakdown(AVMCosmosModuleBreakdown{
		ModulePath:	AVMModulePathActors,
		Purpose: []string{
			"actor_records",
			"actor_state",
			"continuation_integration",
			"mailboxes",
			"permissions",
		},
		StateObjects: []AVMModuleStateObject{
			AVMModuleStateActorRecord,
			AVMModuleStateActorMailbox,
			AVMModuleStateActorState,
			AVMModuleStateActorPermission,
		},
		Messages: []AVMModuleMessageName{
			AVMModuleMsgCreateActor,
			AVMModuleMsgSendActorMessage,
			AVMModuleMsgUpdateActor,
			AVMModuleMsgPauseActor,
		},
		Queries: []AVMModuleQueryName{
			AVMModuleQueryActor,
			AVMModuleQueryActorMailbox,
			AVMModuleQueryActorState,
		},
	})
}

func DefaultXContinuationsModuleBreakdown() (AVMCosmosModuleBreakdown, error) {
	return NewAVMCosmosModuleBreakdown(AVMCosmosModuleBreakdown{
		ModulePath:	AVMModulePathContinuations,
		Purpose: []string{
			"async_workflow_state",
			"continuation_queues",
			"continuation_receipts",
			"expiry",
			"resume",
		},
		StateObjects: []AVMModuleStateObject{
			AVMModuleStateContinuation,
			AVMModuleStateContinuationQueue,
			AVMModuleStateContinuationReceipt,
		},
		Messages: []AVMModuleMessageName{
			AVMModuleMsgCreateContinuation,
			AVMModuleMsgResumeContinuation,
			AVMModuleMsgCancelContinuation,
			AVMModuleMsgExpireContinuation,
		},
		Queries: []AVMModuleQueryName{
			AVMModuleQueryContinuation,
			AVMModuleQueryContinuationsByActor,
			AVMModuleQueryContinuationReceipt,
		},
	})
}

func DefaultXAVMContractsModuleBreakdown() (AVMCosmosModuleBreakdown, error) {
	return NewAVMCosmosModuleBreakdown(AVMCosmosModuleBreakdown{
		ModulePath:	AVMModulePathAVMContracts,
		Purpose: []string{
			"backend_adapters",
			"contract_code",
			"contract_instances",
			"contract_storage",
		},
		StateObjects: []AVMModuleStateObject{
			AVMModuleStateCodeRecord,
			AVMModuleStateContractRecord,
			AVMModuleStateStorageValue,
			AVMModuleStateBackendConfig,
		},
		Messages: []AVMModuleMessageName{
			AVMModuleMsgStoreCode,
			AVMModuleMsgInstantiateContract,
			AVMModuleMsgExecuteContract,
			AVMModuleMsgMigrateContract,
		},
		Queries: []AVMModuleQueryName{
			AVMModuleQueryCode,
			AVMModuleQueryContract,
			AVMModuleQueryContractStorage,
			AVMModuleQueryContractProof,
		},
	})
}

func DefaultXAVMInterfacesModuleBreakdown() (AVMCosmosModuleBreakdown, error) {
	return NewAVMCosmosModuleBreakdown(AVMCosmosModuleBreakdown{
		ModulePath:	AVMModulePathAVMInterfaces,
		Purpose: []string{
			"actor_interface_schemas",
			"contract_interface_schemas",
			"interface_schema_registry",
			"service_interface_schemas",
		},
		StateObjects: []AVMModuleStateObject{
			AVMModuleStateInterfaceDescriptor,
			AVMModuleStateMethodDescriptor,
			AVMModuleStateEventDescriptor,
			AVMModuleStateAsyncHandlerDescriptor,
		},
		Messages: []AVMModuleMessageName{
			AVMModuleMsgRegisterInterface,
			AVMModuleMsgUpdateInterface,
			AVMModuleMsgDeprecateInterface,
		},
		Queries: []AVMModuleQueryName{
			AVMModuleQueryInterface,
			AVMModuleQueryMethod,
			AVMModuleQueryInterfaceByTarget,
		},
	})
}

func NewAVMCosmosModuleBreakdown(breakdown AVMCosmosModuleBreakdown) (AVMCosmosModuleBreakdown, error) {
	breakdown = canonicalAVMCosmosModuleBreakdown(breakdown)
	breakdown.BreakdownHash = ComputeAVMCosmosModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func (b AVMCosmosModuleBreakdown) Validate() error {
	b = canonicalAVMCosmosModuleBreakdown(b)
	if !IsAVMCosmosModulePath(b.ModulePath) {
		return fmt.Errorf("unknown AVM Cosmos module path %q", b.ModulePath)
	}
	if err := validateAVMModulePurpose(b.Purpose); err != nil {
		return err
	}
	if err := validateAVMModuleStateObjects(b.ModulePath, b.StateObjects); err != nil {
		return err
	}
	if err := validateAVMModuleMessages(b.ModulePath, b.Messages); err != nil {
		return err
	}
	if err := validateAVMModuleQueries(b.ModulePath, b.Queries); err != nil {
		return err
	}
	if b.BreakdownHash == "" {
		return errors.New("AVM module breakdown hash is required")
	}
	if err := validateAVMComparisonHash("AVM module breakdown hash", b.BreakdownHash); err != nil {
		return err
	}
	if b.BreakdownHash != ComputeAVMCosmosModuleBreakdownHash(b) {
		return errors.New("AVM module breakdown hash mismatch")
	}
	return nil
}

func NewAVMCosmosModuleRegistry(registry AVMCosmosModuleRegistry) (AVMCosmosModuleRegistry, error) {
	registry = canonicalAVMCosmosModuleRegistry(registry)
	registry.RegistryHash = ComputeAVMCosmosModuleRegistryHash(registry)
	return registry, registry.Validate()
}

func (r AVMCosmosModuleRegistry) Validate() error {
	r = canonicalAVMCosmosModuleRegistry(r)
	required := AllAVMCosmosModulePaths()
	if len(r.Modules) != len(required) {
		return errors.New("AVM module registry must contain x/actors, x/avm, x/async, x/avmcontracts, x/avminterfaces, and x/continuations")
	}
	seen := make(map[AVMCosmosModulePath]struct{}, len(r.Modules))
	for i, module := range r.Modules {
		if err := module.Validate(); err != nil {
			return err
		}
		if _, found := seen[module.ModulePath]; found {
			return fmt.Errorf("duplicate AVM Cosmos module path %q", module.ModulePath)
		}
		seen[module.ModulePath] = struct{}{}
		if i > 0 && r.Modules[i-1].ModulePath >= module.ModulePath {
			return errors.New("AVM module registry must be sorted canonically")
		}
	}
	for _, path := range required {
		if _, found := seen[path]; !found {
			return fmt.Errorf("missing AVM Cosmos module path %q", path)
		}
	}
	if r.RegistryHash == "" {
		return errors.New("AVM module registry hash is required")
	}
	if err := validateAVMComparisonHash("AVM module registry hash", r.RegistryHash); err != nil {
		return err
	}
	if r.RegistryHash != ComputeAVMCosmosModuleRegistryHash(r) {
		return errors.New("AVM module registry hash mismatch")
	}
	return nil
}

func AllAVMCosmosModulePaths() []AVMCosmosModulePath {
	return []AVMCosmosModulePath{AVMModulePathActors, AVMModulePathAVM, AVMModulePathAsync, AVMModulePathAVMContracts, AVMModulePathAVMInterfaces, AVMModulePathContinuations}
}

func IsAVMCosmosModulePath(path AVMCosmosModulePath) bool {
	switch path {
	case AVMModulePathActors, AVMModulePathAVM, AVMModulePathAsync, AVMModulePathAVMContracts, AVMModulePathAVMInterfaces, AVMModulePathContinuations:
		return true
	default:
		return false
	}
}

func ComputeAVMCosmosModuleBreakdownHash(breakdown AVMCosmosModuleBreakdown) string {
	breakdown = canonicalAVMCosmosModuleBreakdown(breakdown)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cosmos-module-breakdown-v1")
	writeEnginePart(h, string(breakdown.ModulePath))
	writeEngineUint64(h, uint64(len(breakdown.Purpose)))
	for _, purpose := range breakdown.Purpose {
		writeEnginePart(h, purpose)
	}
	writeEngineUint64(h, uint64(len(breakdown.StateObjects)))
	for _, object := range breakdown.StateObjects {
		writeEnginePart(h, string(object))
	}
	writeEngineUint64(h, uint64(len(breakdown.Messages)))
	for _, message := range breakdown.Messages {
		writeEnginePart(h, string(message))
	}
	writeEngineUint64(h, uint64(len(breakdown.Queries)))
	for _, query := range breakdown.Queries {
		writeEnginePart(h, string(query))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCosmosModuleRegistryHash(registry AVMCosmosModuleRegistry) string {
	registry = canonicalAVMCosmosModuleRegistry(registry)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cosmos-module-registry-v1")
	writeEngineUint64(h, uint64(len(registry.Modules)))
	for _, module := range registry.Modules {
		writeEnginePart(h, module.BreakdownHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMCosmosModuleBreakdown(breakdown AVMCosmosModuleBreakdown) AVMCosmosModuleBreakdown {
	breakdown.ModulePath = AVMCosmosModulePath(strings.TrimSpace(string(breakdown.ModulePath)))
	breakdown.Purpose = cloneSortedStrings(breakdown.Purpose)
	breakdown.StateObjects = sortedAVMModuleStateObjects(breakdown.StateObjects)
	breakdown.Messages = sortedAVMModuleMessages(breakdown.Messages)
	breakdown.Queries = sortedAVMModuleQueries(breakdown.Queries)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func canonicalAVMCosmosModuleRegistry(registry AVMCosmosModuleRegistry) AVMCosmosModuleRegistry {
	registry.Modules = append([]AVMCosmosModuleBreakdown(nil), registry.Modules...)
	for i := range registry.Modules {
		registry.Modules[i] = canonicalAVMCosmosModuleBreakdown(registry.Modules[i])
	}
	sort.SliceStable(registry.Modules, func(i, j int) bool {
		return registry.Modules[i].ModulePath < registry.Modules[j].ModulePath
	})
	registry.RegistryHash = strings.TrimSpace(registry.RegistryHash)
	return registry
}

func validateAVMModulePurpose(purpose []string) error {
	if len(purpose) == 0 || len(purpose) > MaxAVMModulePurposeItems {
		return fmt.Errorf("AVM module purpose entries must be 1..%d", MaxAVMModulePurposeItems)
	}
	for _, item := range purpose {
		if err := validateEngineToken("AVM module purpose", item, MaxAVMModulePurposeText); err != nil {
			return err
		}
	}
	return validateEngineTokens("AVM module purpose", purpose, MaxAVMModulePurposeText)
}

func validateAVMModuleStateObjects(module AVMCosmosModulePath, values []AVMModuleStateObject) error {
	return validateAVMModuleEnumSet("state object", values, requiredAVMModuleStateObjects(module), isAVMModuleStateObject)
}

func validateAVMModuleMessages(module AVMCosmosModulePath, values []AVMModuleMessageName) error {
	return validateAVMModuleEnumSet("message", values, requiredAVMModuleMessages(module), isAVMModuleMessageName)
}

func validateAVMModuleQueries(module AVMCosmosModulePath, values []AVMModuleQueryName) error {
	return validateAVMModuleEnumSet("query", values, requiredAVMModuleQueries(module), isAVMModuleQueryName)
}

func requiredAVMModuleStateObjects(module AVMCosmosModulePath) []AVMModuleStateObject {
	switch module {
	case AVMModulePathActors:
		return []AVMModuleStateObject{AVMModuleStateActorRecord, AVMModuleStateActorMailbox, AVMModuleStateActorState, AVMModuleStateActorPermission}
	case AVMModulePathAVM:
		return []AVMModuleStateObject{AVMModuleStateAVMParams, AVMModuleStateRouteDescriptor, AVMModuleStateAVMRoot, AVMModuleStateExecutionReceipt, AVMModuleStateRuntimeVersion}
	case AVMModulePathAsync:
		return []AVMModuleStateObject{AVMModuleStateAsyncMessage, AVMModuleStateZoneQueue, AVMModuleStateRetryRecord, AVMModuleStateDeadLetter, AVMModuleStateReplayTombstone}
	case AVMModulePathAVMContracts:
		return []AVMModuleStateObject{AVMModuleStateCodeRecord, AVMModuleStateContractRecord, AVMModuleStateStorageValue, AVMModuleStateBackendConfig}
	case AVMModulePathAVMInterfaces:
		return []AVMModuleStateObject{AVMModuleStateInterfaceDescriptor, AVMModuleStateMethodDescriptor, AVMModuleStateEventDescriptor, AVMModuleStateAsyncHandlerDescriptor}
	case AVMModulePathContinuations:
		return []AVMModuleStateObject{AVMModuleStateContinuation, AVMModuleStateContinuationQueue, AVMModuleStateContinuationReceipt}
	default:
		return nil
	}
}

func requiredAVMModuleMessages(module AVMCosmosModulePath) []AVMModuleMessageName {
	switch module {
	case AVMModulePathActors:
		return []AVMModuleMessageName{AVMModuleMsgCreateActor, AVMModuleMsgSendActorMessage, AVMModuleMsgUpdateActor, AVMModuleMsgPauseActor}
	case AVMModulePathAVM:
		return []AVMModuleMessageName{AVMModuleMsgSubmitAVMMessage, AVMModuleMsgRegisterRoute, AVMModuleMsgUpdateAVMParams, AVMModuleMsgScheduleRuntimeUpgrade}
	case AVMModulePathAsync:
		return []AVMModuleMessageName{AVMModuleMsgSubmitAsyncMessage, AVMModuleMsgCancelAsyncMessage, AVMModuleMsgRetryAsyncMessage, AVMModuleMsgExpireAsyncMessage}
	case AVMModulePathAVMContracts:
		return []AVMModuleMessageName{AVMModuleMsgStoreCode, AVMModuleMsgInstantiateContract, AVMModuleMsgExecuteContract, AVMModuleMsgMigrateContract}
	case AVMModulePathAVMInterfaces:
		return []AVMModuleMessageName{AVMModuleMsgRegisterInterface, AVMModuleMsgUpdateInterface, AVMModuleMsgDeprecateInterface}
	case AVMModulePathContinuations:
		return []AVMModuleMessageName{AVMModuleMsgCreateContinuation, AVMModuleMsgResumeContinuation, AVMModuleMsgCancelContinuation, AVMModuleMsgExpireContinuation}
	default:
		return nil
	}
}

func requiredAVMModuleQueries(module AVMCosmosModulePath) []AVMModuleQueryName {
	switch module {
	case AVMModulePathActors:
		return []AVMModuleQueryName{AVMModuleQueryActor, AVMModuleQueryActorMailbox, AVMModuleQueryActorState}
	case AVMModulePathAVM:
		return []AVMModuleQueryName{AVMModuleQueryAVMParams, AVMModuleQueryAVMRoot, AVMModuleQueryRoute, AVMModuleQueryExecutionReceipt, AVMModuleQueryRuntimeVersion}
	case AVMModulePathAsync:
		return []AVMModuleQueryName{AVMModuleQueryAsyncMessage, AVMModuleQueryZoneQueue, AVMModuleQueryDeadLetter, AVMModuleQueryReplayTombstone}
	case AVMModulePathAVMContracts:
		return []AVMModuleQueryName{AVMModuleQueryCode, AVMModuleQueryContract, AVMModuleQueryContractStorage, AVMModuleQueryContractProof}
	case AVMModulePathAVMInterfaces:
		return []AVMModuleQueryName{AVMModuleQueryInterface, AVMModuleQueryMethod, AVMModuleQueryInterfaceByTarget}
	case AVMModulePathContinuations:
		return []AVMModuleQueryName{AVMModuleQueryContinuation, AVMModuleQueryContinuationsByActor, AVMModuleQueryContinuationReceipt}
	default:
		return nil
	}
}

func isAVMModuleStateObject(value AVMModuleStateObject) bool {
	for _, module := range AllAVMCosmosModulePaths() {
		for _, required := range requiredAVMModuleStateObjects(module) {
			if value == required {
				return true
			}
		}
	}
	return false
}

func isAVMModuleMessageName(value AVMModuleMessageName) bool {
	for _, module := range AllAVMCosmosModulePaths() {
		for _, required := range requiredAVMModuleMessages(module) {
			if value == required {
				return true
			}
		}
	}
	return false
}

func isAVMModuleQueryName(value AVMModuleQueryName) bool {
	for _, module := range AllAVMCosmosModulePaths() {
		for _, required := range requiredAVMModuleQueries(module) {
			if value == required {
				return true
			}
		}
	}
	return false
}

func validateAVMModuleEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) || len(values) > MaxAVMModuleSurfaceItems {
		return fmt.Errorf("AVM module expected %d %s entries", len(required), label)
	}
	seen := make(map[T]struct{}, len(values))
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("AVM module unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("AVM module %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("AVM module duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("AVM module missing %s %s", label, value)
		}
	}
	return nil
}

func sortedAVMModuleStateObjects(values []AVMModuleStateObject) []AVMModuleStateObject {
	out := append([]AVMModuleStateObject(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedAVMModuleMessages(values []AVMModuleMessageName) []AVMModuleMessageName {
	out := append([]AVMModuleMessageName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedAVMModuleQueries(values []AVMModuleQueryName) []AVMModuleQueryName {
	out := append([]AVMModuleQueryName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
