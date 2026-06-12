package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMStatePrefixParams			= "avm/params"
	AVMStatePrefixZones			= "avm/zones"
	AVMStatePrefixRouterRoutes		= "avm/router/routes"
	AVMStatePrefixAsyncMessages		= "avm/async/messages"
	AVMStatePrefixAsyncQueues		= "avm/async/queues"
	AVMStatePrefixAsyncRetry		= "avm/async/retry"
	AVMStatePrefixAsyncDead			= "avm/async/dead"
	AVMStatePrefixActors			= "avm/actors"
	AVMStatePrefixActorMailbox		= "avm/actors/mailbox"
	AVMStatePrefixContinuations		= "avm/continuations"
	AVMStatePrefixContractCode		= "avm/contracts/code"
	AVMStatePrefixContractInstances		= "avm/contracts/instances"
	AVMStatePrefixContractStorage		= "avm/contracts/storage"
	AVMStatePrefixInterfaces		= "avm/interfaces"
	AVMStatePrefixReceipts			= "avm/receipts"
	AVMStatePrefixRoots			= "avm/roots"
	AVMStatePrefixZoneStateRoots		= "avm/roots/state"
	AVMStatePrefixZoneMessageRoots		= "avm/roots/messages"
	AVMStatePrefixZoneContinuationRoot	= "avm/roots/continuations"

	AVMStateValueParams		= "AVMParams"
	AVMStateValueZoneRuntimeConfig	= "ZoneRuntimeConfig"
	AVMStateValueRouteDescriptor	= "RouteDescriptor"
	AVMStateValueAsyncMessage	= "AsyncMessage"
	AVMStateValueMessageID		= "message_id"
	AVMStateValueRetryRecord	= "RetryRecord"
	AVMStateValueDeadLetterRecord	= "DeadLetterRecord"
	AVMStateValueActorRecord	= "ActorRecord"
	AVMStateValueContinuation	= "Continuation"
	AVMStateValueCodeRecord		= "CodeRecord"
	AVMStateValueContractRecord	= "ContractRecord"
	AVMStateValueStorageValue	= "StorageValue"
	AVMStateValueInterface		= "InterfaceDescriptor"
	AVMStateValueReceipt		= "ExecutionReceipt"
	AVMStateValueRoot		= "AVMRoot"

	MaxAVMStateKeySegmentLength	= 128
	MaxAVMMessageBytes		= 1024 * 1024
)

type AVMStatePrefixDescriptor struct {
	Prefix		string
	KeyTemplate	string
	ValueType	string
}

type AVMStateModel struct {
	Prefixes	[]AVMStatePrefixDescriptor
	Root		string
}

type AVMZoneRuntimeConfig struct {
	ZoneID			zonestypes.ZoneID
	Enabled			bool
	ExecutionBudgetPerBlock	zonestypes.ZoneExecutionBudget
	AsyncBudgetPerBlock	zonestypes.ZoneExecutionBudget
	MaxQueueDepth		uint32
	MaxMessageBytes		uint32
	GasPolicyID		string
	RetryPolicyID		string
	AllowedMessageTypes	[]string
	StateRootPrefix		string
	MessageRootPrefix	string
	ContinuationRootPrefix	string
	ConfigRoot		string
}

func DefaultAVMStateModel() AVMStateModel {
	model := AVMStateModel{Prefixes: []AVMStatePrefixDescriptor{
		{Prefix: AVMStatePrefixParams, KeyTemplate: "avm/params", ValueType: AVMStateValueParams},
		{Prefix: AVMStatePrefixZones, KeyTemplate: "avm/zones/{zone_id}", ValueType: AVMStateValueZoneRuntimeConfig},
		{Prefix: AVMStatePrefixRouterRoutes, KeyTemplate: "avm/router/routes/{route_key}", ValueType: AVMStateValueRouteDescriptor},
		{Prefix: AVMStatePrefixAsyncMessages, KeyTemplate: "avm/async/messages/{message_id}", ValueType: AVMStateValueAsyncMessage},
		{Prefix: AVMStatePrefixAsyncQueues, KeyTemplate: "avm/async/queues/{zone_id}/{queue_id}/{sort_key}", ValueType: AVMStateValueMessageID},
		{Prefix: AVMStatePrefixAsyncRetry, KeyTemplate: "avm/async/retry/{zone_id}/{height}/{message_id}", ValueType: AVMStateValueRetryRecord},
		{Prefix: AVMStatePrefixAsyncDead, KeyTemplate: "avm/async/dead/{zone_id}/{message_id}", ValueType: AVMStateValueDeadLetterRecord},
		{Prefix: AVMStatePrefixActors, KeyTemplate: "avm/actors/{actor_id}", ValueType: AVMStateValueActorRecord},
		{Prefix: AVMStatePrefixActorMailbox, KeyTemplate: "avm/actors/mailbox/{actor_id}/{sort_key}", ValueType: AVMStateValueMessageID},
		{Prefix: AVMStatePrefixContinuations, KeyTemplate: "avm/continuations/{continuation_id}", ValueType: AVMStateValueContinuation},
		{Prefix: AVMStatePrefixContractCode, KeyTemplate: "avm/contracts/code/{code_id}", ValueType: AVMStateValueCodeRecord},
		{Prefix: AVMStatePrefixContractInstances, KeyTemplate: "avm/contracts/instances/{contract_addr}", ValueType: AVMStateValueContractRecord},
		{Prefix: AVMStatePrefixContractStorage, KeyTemplate: "avm/contracts/storage/{contract_addr}/{key}", ValueType: AVMStateValueStorageValue},
		{Prefix: AVMStatePrefixInterfaces, KeyTemplate: "avm/interfaces/{interface_hash}", ValueType: AVMStateValueInterface},
		{Prefix: AVMStatePrefixReceipts, KeyTemplate: "avm/receipts/{receipt_id}", ValueType: AVMStateValueReceipt},
		{Prefix: AVMStatePrefixRoots, KeyTemplate: "avm/roots/{height}", ValueType: AVMStateValueRoot},
	}}
	model = canonicalAVMStateModel(model)
	model.Root = ComputeAVMStateModelRoot(model)
	return model
}

func DefaultAVMZoneRuntimeConfig(zoneID zonestypes.ZoneID) (AVMZoneRuntimeConfig, error) {
	config := AVMZoneRuntimeConfig{
		ZoneID:				zoneID,
		Enabled:			true,
		ExecutionBudgetPerBlock:	zonestypes.DefaultZoneExecutionBudget(),
		AsyncBudgetPerBlock: zonestypes.ZoneExecutionBudget{
			MaxGas:		500_000,
			MaxMessages:	500,
		},
		MaxQueueDepth:		10_000,
		MaxMessageBytes:	64 * 1024,
		GasPolicyID:		"default-gas",
		RetryPolicyID:		"default-retry",
		AllowedMessageTypes:	[]string{"*"},
		StateRootPrefix:	AVMZoneStateRootPrefix(zoneID),
		MessageRootPrefix:	AVMZoneMessageRootPrefix(zoneID),
		ContinuationRootPrefix:	AVMZoneContinuationRootPrefix(zoneID),
	}
	config = canonicalAVMZoneRuntimeConfig(config)
	config.ConfigRoot = ComputeAVMZoneRuntimeConfigRoot(config)
	return config, config.Validate()
}

func (m AVMStateModel) Validate() error {
	m = canonicalAVMStateModel(m)
	if len(m.Prefixes) == 0 {
		return errors.New("AVM state model must declare prefixes")
	}
	seen := make(map[string]struct{}, len(m.Prefixes))
	byPrefix := make(map[string]AVMStatePrefixDescriptor, len(m.Prefixes))
	for i, descriptor := range m.Prefixes {
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if _, found := seen[descriptor.Prefix]; found {
			return fmt.Errorf("duplicate AVM state prefix %q", descriptor.Prefix)
		}
		seen[descriptor.Prefix] = struct{}{}
		byPrefix[descriptor.Prefix] = descriptor
		if i > 0 && m.Prefixes[i-1].Prefix >= descriptor.Prefix {
			return errors.New("AVM state prefixes must be sorted canonically")
		}
	}
	if err := validateRequiredAVMStatePrefixes(byPrefix); err != nil {
		return err
	}
	if m.Root == "" {
		return errors.New("AVM state model root is required")
	}
	if m.Root != ComputeAVMStateModelRoot(m) {
		return errors.New("AVM state model root mismatch")
	}
	return nil
}

func (d AVMStatePrefixDescriptor) Validate() error {
	if err := validateAVMStatePrefix("AVM state prefix", d.Prefix); err != nil {
		return err
	}
	if err := validateAVMStatePrefix("AVM state key template", d.KeyTemplate); err != nil {
		return err
	}
	if !strings.HasPrefix(d.KeyTemplate, d.Prefix) {
		return errors.New("AVM state key template must start with its prefix")
	}
	return validateEngineToken("AVM state value type", d.ValueType, MaxAVMStateKeySegmentLength)
}

func (c AVMZoneRuntimeConfig) Validate() error {
	c = canonicalAVMZoneRuntimeConfig(c)
	if err := zonestypes.ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if err := c.ExecutionBudgetPerBlock.Validate(); err != nil {
		return fmt.Errorf("execution budget per block: %w", err)
	}
	if err := c.AsyncBudgetPerBlock.Validate(); err != nil {
		return fmt.Errorf("async budget per block: %w", err)
	}
	if c.MaxQueueDepth == 0 {
		return errors.New("zone runtime max queue depth must be positive")
	}
	if c.MaxMessageBytes == 0 || c.MaxMessageBytes > MaxAVMMessageBytes {
		return fmt.Errorf("zone runtime max message bytes must be 1..%d", MaxAVMMessageBytes)
	}
	if err := validateEngineToken("zone runtime gas policy id", c.GasPolicyID, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := validateEngineToken("zone runtime retry policy id", c.RetryPolicyID, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := (zonestypes.ZoneMessageFilter{AllowedMessageTypes: c.AllowedMessageTypes}).Validate(); err != nil {
		return err
	}
	if err := validateExactAVMRootPrefix("zone runtime state root prefix", c.StateRootPrefix, AVMZoneStateRootPrefix(c.ZoneID)); err != nil {
		return err
	}
	if err := validateExactAVMRootPrefix("zone runtime message root prefix", c.MessageRootPrefix, AVMZoneMessageRootPrefix(c.ZoneID)); err != nil {
		return err
	}
	if err := validateExactAVMRootPrefix("zone runtime continuation root prefix", c.ContinuationRootPrefix, AVMZoneContinuationRootPrefix(c.ZoneID)); err != nil {
		return err
	}
	if c.ConfigRoot == "" {
		return errors.New("zone runtime config root is required")
	}
	if c.ConfigRoot != ComputeAVMZoneRuntimeConfigRoot(c) {
		return errors.New("zone runtime config root mismatch")
	}
	return nil
}

func AVMParamsKey() string {
	return AVMStatePrefixParams
}

func AVMZoneRuntimeConfigKey(zoneID zonestypes.ZoneID) string {
	return AVMStatePrefixZones + "/" + string(zoneID)
}

func AVMRouteDescriptorKey(routeKey string) string {
	return AVMStatePrefixRouterRoutes + "/" + strings.TrimSpace(routeKey)
}

func AVMAsyncMessageKey(messageID string) string {
	return AVMStatePrefixAsyncMessages + "/" + strings.TrimSpace(messageID)
}

func AVMAsyncQueueKey(zoneID zonestypes.ZoneID, queueID, sortKey string) string {
	return AVMStatePrefixAsyncQueues + "/" + string(zoneID) + "/" + strings.TrimSpace(queueID) + "/" + strings.TrimSpace(sortKey)
}

func AVMAsyncRetryKey(zoneID zonestypes.ZoneID, height uint64, messageID string) string {
	return fmt.Sprintf("%s/%s/%020d/%s", AVMStatePrefixAsyncRetry, zoneID, height, strings.TrimSpace(messageID))
}

func AVMAsyncDeadLetterKey(zoneID zonestypes.ZoneID, messageID string) string {
	return AVMStatePrefixAsyncDead + "/" + string(zoneID) + "/" + strings.TrimSpace(messageID)
}

func AVMActorRecordKey(actorID string) string {
	return AVMStatePrefixActors + "/" + strings.TrimSpace(actorID)
}

func AVMActorMailboxKey(actorID, sortKey string) string {
	return AVMStatePrefixActorMailbox + "/" + strings.TrimSpace(actorID) + "/" + strings.TrimSpace(sortKey)
}

func AVMContinuationKey(continuationID string) string {
	return AVMStatePrefixContinuations + "/" + strings.TrimSpace(continuationID)
}

func AVMContractCodeKey(codeID uint64) string {
	return fmt.Sprintf("%s/%020d", AVMStatePrefixContractCode, codeID)
}

func AVMContractInstanceKey(contractAddr string) string {
	return AVMStatePrefixContractInstances + "/" + strings.TrimSpace(contractAddr)
}

func AVMContractStorageKey(contractAddr, key string) string {
	return AVMStatePrefixContractStorage + "/" + strings.TrimSpace(contractAddr) + "/" + strings.TrimSpace(key)
}

func AVMInterfaceDescriptorKey(interfaceHash string) string {
	return AVMStatePrefixInterfaces + "/" + strings.TrimSpace(interfaceHash)
}

func AVMReceiptKey(receiptID string) string {
	return AVMStatePrefixReceipts + "/" + strings.TrimSpace(receiptID)
}

func AVMRootKey(height uint64) string {
	return fmt.Sprintf("%s/%020d", AVMStatePrefixRoots, height)
}

func AVMZoneStateRootPrefix(zoneID zonestypes.ZoneID) string {
	return AVMStatePrefixZoneStateRoots + "/" + string(zoneID) + "/"
}

func AVMZoneMessageRootPrefix(zoneID zonestypes.ZoneID) string {
	return AVMStatePrefixZoneMessageRoots + "/" + string(zoneID) + "/"
}

func AVMZoneContinuationRootPrefix(zoneID zonestypes.ZoneID) string {
	return AVMStatePrefixZoneContinuationRoot + "/" + string(zoneID) + "/"
}

func ComputeAVMStateModelRoot(model AVMStateModel) string {
	model = canonicalAVMStateModel(model)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-state-model-v1")
	writeEngineUint64(h, uint64(len(model.Prefixes)))
	for _, descriptor := range model.Prefixes {
		writeEnginePart(h, descriptor.Prefix)
		writeEnginePart(h, descriptor.KeyTemplate)
		writeEnginePart(h, descriptor.ValueType)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneRuntimeConfigRoot(config AVMZoneRuntimeConfig) string {
	config = canonicalAVMZoneRuntimeConfig(config)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-runtime-config-v1")
	writeEnginePart(h, string(config.ZoneID))
	writeEngineBool(h, config.Enabled)
	writeEngineUint64(h, config.ExecutionBudgetPerBlock.MaxGas)
	writeEngineUint64(h, config.ExecutionBudgetPerBlock.GasUsed)
	writeEngineUint64(h, uint64(config.ExecutionBudgetPerBlock.MaxMessages))
	writeEngineUint64(h, uint64(config.ExecutionBudgetPerBlock.MessagesUsed))
	writeEngineUint64(h, config.AsyncBudgetPerBlock.MaxGas)
	writeEngineUint64(h, config.AsyncBudgetPerBlock.GasUsed)
	writeEngineUint64(h, uint64(config.AsyncBudgetPerBlock.MaxMessages))
	writeEngineUint64(h, uint64(config.AsyncBudgetPerBlock.MessagesUsed))
	writeEngineUint64(h, uint64(config.MaxQueueDepth))
	writeEngineUint64(h, uint64(config.MaxMessageBytes))
	writeEnginePart(h, config.GasPolicyID)
	writeEnginePart(h, config.RetryPolicyID)
	writeEngineUint64(h, uint64(len(config.AllowedMessageTypes)))
	for _, messageType := range config.AllowedMessageTypes {
		writeEnginePart(h, messageType)
	}
	writeEnginePart(h, config.StateRootPrefix)
	writeEnginePart(h, config.MessageRootPrefix)
	writeEnginePart(h, config.ContinuationRootPrefix)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMStateModel(model AVMStateModel) AVMStateModel {
	out := AVMStateModel{
		Prefixes:	append([]AVMStatePrefixDescriptor(nil), model.Prefixes...),
		Root:		strings.TrimSpace(model.Root),
	}
	for i := range out.Prefixes {
		out.Prefixes[i].Prefix = strings.TrimSpace(out.Prefixes[i].Prefix)
		out.Prefixes[i].KeyTemplate = strings.TrimSpace(out.Prefixes[i].KeyTemplate)
		out.Prefixes[i].ValueType = strings.TrimSpace(out.Prefixes[i].ValueType)
	}
	sort.SliceStable(out.Prefixes, func(i, j int) bool {
		return out.Prefixes[i].Prefix < out.Prefixes[j].Prefix
	})
	return out
}

func canonicalAVMZoneRuntimeConfig(config AVMZoneRuntimeConfig) AVMZoneRuntimeConfig {
	config.GasPolicyID = strings.TrimSpace(config.GasPolicyID)
	config.RetryPolicyID = strings.TrimSpace(config.RetryPolicyID)
	config.StateRootPrefix = strings.TrimSpace(config.StateRootPrefix)
	config.MessageRootPrefix = strings.TrimSpace(config.MessageRootPrefix)
	config.ContinuationRootPrefix = strings.TrimSpace(config.ContinuationRootPrefix)
	config.ConfigRoot = strings.TrimSpace(config.ConfigRoot)
	config.AllowedMessageTypes = append([]string(nil), config.AllowedMessageTypes...)
	for i, messageType := range config.AllowedMessageTypes {
		config.AllowedMessageTypes[i] = strings.TrimSpace(messageType)
	}
	sort.Strings(config.AllowedMessageTypes)
	return config
}

func validateRequiredAVMStatePrefixes(byPrefix map[string]AVMStatePrefixDescriptor) error {
	required := map[string]AVMStatePrefixDescriptor{}
	for _, descriptor := range DefaultAVMStateModel().Prefixes {
		required[descriptor.Prefix] = descriptor
	}
	for prefix, expected := range required {
		actual, found := byPrefix[prefix]
		if !found {
			return fmt.Errorf("AVM state model missing prefix %q", prefix)
		}
		if actual.KeyTemplate != expected.KeyTemplate || actual.ValueType != expected.ValueType {
			return fmt.Errorf("AVM state prefix %q descriptor mismatch", prefix)
		}
	}
	return nil
}

func validateAVMStatePrefix(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if !strings.HasPrefix(value, "avm/") {
		return fmt.Errorf("%s must start with avm/", fieldName)
	}
	if strings.Contains(value, "//") {
		return fmt.Errorf("%s must not contain empty path segments", fieldName)
	}
	for _, part := range strings.Split(value, "/") {
		if part == "" {
			return fmt.Errorf("%s must not contain empty path segments", fieldName)
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			if err := validateEngineToken(fieldName+" placeholder", name, MaxAVMStateKeySegmentLength); err != nil {
				return err
			}
			continue
		}
		if err := validateEngineToken(fieldName+" segment", part, MaxAVMStateKeySegmentLength); err != nil {
			return err
		}
	}
	return nil
}

func validateExactAVMRootPrefix(fieldName, actual, expected string) error {
	if err := validateAVMStatePrefix(fieldName, strings.TrimSuffix(actual, "/")); err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s must be %q", fieldName, expected)
	}
	return nil
}
