package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultZoneStoreKey	= "zones"

	ZoneStateNamespacePrefix	= "zone/state/"
	ZoneQueueNamespacePrefix	= "zone/queue/"
	ZoneProofNamespacePrefix	= "zone/proof/"
	ZoneQueryNamespacePrefix	= "zone/query/"
	ZoneKVPrefixRoot		= "zones/"
	ZonePipelinePrefix		= "zone/pipeline/"

	MaxZoneMessageTypeLength	= 96
	MaxZoneEndpointLength		= 128
	MaxZoneNamespaceLength		= 128
	MaxZoneModuleNameLength		= 96
)

type ZoneExecutionBudget struct {
	MaxGas		uint64
	GasUsed		uint64
	MaxMessages	uint32
	MessagesUsed	uint32
}

type ZoneGasPolicy struct {
	Denom			string
	MaxGasPerBlock		uint64
	MaxGasPerMessage	uint64
}

type ZoneMessageFilter struct {
	AllowedMessageTypes []string
}

type ZoneMessage struct {
	ZoneID		ZoneID
	MessageType	string
	Source		string
	Destination	string
	GasLimit	uint64
	PayloadHash	string
	Sequence	uint64
}

type ZoneRuntimeState struct {
	ZoneID			ZoneID
	StoreKey		string
	StateNamespace		string
	QueueNamespace		string
	ProofNamespace		string
	QueryNamespace		string
	KVPrefix		string
	ModuleSet		[]string
	ModuleSetRoot		string
	ExecutionPipeline	string
	StateRoot		string
	ReceiptRoot		string
	MessageRoot		string
	ExecutionResultRoot	string
	ProofRoot		string
	MessageQueue		[]ZoneMessage
	Budget			ZoneExecutionBudget
	GasPolicy		ZoneGasPolicy
	MessageFilter		ZoneMessageFilter
}

type ParallelExecutionPlan struct {
	Height	uint64
	Zones	[]ZoneRuntimeState
}

func DefaultZoneExecutionBudget() ZoneExecutionBudget {
	return ZoneExecutionBudget{
		MaxGas:		1_000_000,
		MaxMessages:	1_000,
	}
}

func DefaultZoneGasPolicy() ZoneGasPolicy {
	return ZoneGasPolicy{
		Denom:			FeePolicyNaet,
		MaxGasPerBlock:		1_000_000,
		MaxGasPerMessage:	100_000,
	}
}

func DefaultZoneMessageFilter() ZoneMessageFilter {
	return ZoneMessageFilter{AllowedMessageTypes: []string{"*"}}
}

func NewZoneRuntimeState(zone Zone, stateRoot string, queue []ZoneMessage, budget ZoneExecutionBudget, gasPolicy ZoneGasPolicy, filter ZoneMessageFilter) (ZoneRuntimeState, error) {
	if err := zone.Validate(); err != nil {
		return ZoneRuntimeState{}, err
	}
	if budget.MaxGas == 0 && budget.MaxMessages == 0 {
		budget = DefaultZoneExecutionBudget()
	}
	if gasPolicy.Denom == "" {
		gasPolicy = DefaultZoneGasPolicy()
	}
	if len(filter.AllowedMessageTypes) == 0 {
		filter = DefaultZoneMessageFilter()
	}
	runtime := ZoneRuntimeState{
		ZoneID:			zone.ID,
		StoreKey:		DefaultZoneStoreKey,
		StateNamespace:		ZoneStateNamespace(zone.ID),
		QueueNamespace:		ZoneQueueNamespace(zone.ID),
		ProofNamespace:		ZoneProofNamespace(zone.ID),
		QueryNamespace:		ZoneQueryNamespace(zone.ID),
		KVPrefix:		ZoneKVPrefix(zone.ID),
		ModuleSet:		DefaultZoneModuleSet(zone),
		ExecutionPipeline:	ZoneExecutionPipeline(zone.ID),
		StateRoot:		stateRoot,
		ReceiptRoot:		EmptyRootHash(),
		MessageRoot:		ComputeZoneMessageRoot(queue),
		ExecutionResultRoot:	EmptyRootHash(),
		MessageQueue:		cloneZoneMessages(queue),
		Budget:			budget,
		GasPolicy:		gasPolicy,
		MessageFilter:		filter,
	}
	runtime.ModuleSetRoot = ComputeZoneModuleSetRoot(runtime.ModuleSet)
	runtime.ProofRoot = ComputeZoneRuntimeProofRoot(runtime)
	return runtime, runtime.Validate()
}

func BuildParallelExecutionPlan(height uint64, zones []ZoneRuntimeState) (ParallelExecutionPlan, error) {
	if height == 0 {
		return ParallelExecutionPlan{}, errors.New("parallel execution height must be positive")
	}
	out := make([]ZoneRuntimeState, len(zones))
	for i, zone := range zones {
		if err := zone.Validate(); err != nil {
			return ParallelExecutionPlan{}, err
		}
		out[i] = zone.Clone()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ZoneID < out[j].ZoneID
	})
	plan := ParallelExecutionPlan{Height: height, Zones: out}
	return plan, plan.Validate()
}

func (p ParallelExecutionPlan) Validate() error {
	if p.Height == 0 {
		return errors.New("parallel execution height must be positive")
	}
	seenZones := make(map[ZoneID]struct{}, len(p.Zones))
	seenPrefixes := make([]string, 0, len(p.Zones))
	for i, zone := range p.Zones {
		if err := zone.Validate(); err != nil {
			return err
		}
		if _, found := seenZones[zone.ZoneID]; found {
			return fmt.Errorf("duplicate execution zone %s", zone.ZoneID)
		}
		seenZones[zone.ZoneID] = struct{}{}
		if i > 0 && p.Zones[i-1].ZoneID >= zone.ZoneID {
			return errors.New("parallel execution zones must be sorted canonically")
		}
		for _, existing := range seenPrefixes {
			if prefixesOverlap(existing, zone.KVPrefix) {
				return fmt.Errorf("zone KV prefix %q overlaps %q", zone.KVPrefix, existing)
			}
		}
		seenPrefixes = append(seenPrefixes, zone.KVPrefix)
	}
	return nil
}

func (z ZoneRuntimeState) Validate() error {
	if err := ValidateZoneID(z.ZoneID); err != nil {
		return err
	}
	if err := validatePolicyID("zone runtime store key", z.StoreKey); err != nil {
		return err
	}
	if err := validateZoneNamespace("zone state namespace", z.StateNamespace, ZoneStateNamespace(z.ZoneID)); err != nil {
		return err
	}
	if err := validateZoneNamespace("zone queue namespace", z.QueueNamespace, ZoneQueueNamespace(z.ZoneID)); err != nil {
		return err
	}
	if err := validateZoneNamespace("zone proof namespace", z.ProofNamespace, ZoneProofNamespace(z.ZoneID)); err != nil {
		return err
	}
	if err := validateZoneNamespace("zone query namespace", z.QueryNamespace, ZoneQueryNamespace(z.ZoneID)); err != nil {
		return err
	}
	if err := validateZoneNamespace("zone KV prefix", z.KVPrefix, ZoneKVPrefix(z.ZoneID)); err != nil {
		return err
	}
	if err := validateRuntimeToken("zone execution pipeline", z.ExecutionPipeline, MaxZoneNamespaceLength); err != nil {
		return err
	}
	if err := validateZoneModuleSet(z.ModuleSet); err != nil {
		return err
	}
	if err := ValidateHash("zone module set root", z.ModuleSetRoot); err != nil {
		return err
	}
	if z.ModuleSetRoot != ComputeZoneModuleSetRoot(z.ModuleSet) {
		return errors.New("zone module set root mismatch")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "zone state root", value: z.StateRoot},
		{name: "zone receipt root", value: z.ReceiptRoot},
		{name: "zone message root", value: z.MessageRoot},
		{name: "zone execution result root", value: z.ExecutionResultRoot},
		{name: "zone proof root", value: z.ProofRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if z.ProofRoot != ComputeZoneRuntimeProofRoot(z) {
		return errors.New("zone proof root mismatch")
	}
	if err := z.Budget.Validate(); err != nil {
		return err
	}
	if err := z.GasPolicy.Validate(); err != nil {
		return err
	}
	if err := z.MessageFilter.Validate(); err != nil {
		return err
	}
	if z.MessageRoot != ComputeZoneMessageRoot(z.MessageQueue) {
		return errors.New("zone message root mismatch")
	}
	return validateZoneMessageQueue(z.ZoneID, z.MessageQueue, z.Budget, z.GasPolicy, z.MessageFilter)
}

func (z ZoneRuntimeState) Clone() ZoneRuntimeState {
	z.ModuleSet = append([]string(nil), z.ModuleSet...)
	z.MessageQueue = cloneZoneMessages(z.MessageQueue)
	z.MessageFilter.AllowedMessageTypes = append([]string(nil), z.MessageFilter.AllowedMessageTypes...)
	return z
}

func (b ZoneExecutionBudget) Validate() error {
	if b.MaxGas == 0 {
		return errors.New("zone execution max gas must be positive")
	}
	if b.GasUsed > b.MaxGas {
		return errors.New("zone execution gas used exceeds max gas")
	}
	if b.MaxMessages == 0 {
		return errors.New("zone execution max messages must be positive")
	}
	if b.MessagesUsed > b.MaxMessages {
		return errors.New("zone execution messages used exceeds max messages")
	}
	return nil
}

func (b ZoneExecutionBudget) Consume(gas uint64, messages uint32) (ZoneExecutionBudget, error) {
	if gas > ^uint64(0)-b.GasUsed {
		return ZoneExecutionBudget{}, errors.New("zone execution gas overflow")
	}
	next := b
	next.GasUsed += gas
	next.MessagesUsed += messages
	if next.MessagesUsed < b.MessagesUsed {
		return ZoneExecutionBudget{}, errors.New("zone execution message count overflow")
	}
	return next, next.Validate()
}

func (g ZoneGasPolicy) Validate() error {
	if strings.TrimSpace(g.Denom) != FeePolicyNaet {
		return fmt.Errorf("zone gas policy must use %s", FeePolicyNaet)
	}
	if g.MaxGasPerBlock == 0 {
		return errors.New("zone gas max gas per block must be positive")
	}
	if g.MaxGasPerMessage == 0 {
		return errors.New("zone gas max gas per message must be positive")
	}
	if g.MaxGasPerMessage > g.MaxGasPerBlock {
		return errors.New("zone gas per-message limit must not exceed per-block limit")
	}
	return nil
}

func (f ZoneMessageFilter) Validate() error {
	if len(f.AllowedMessageTypes) == 0 {
		return errors.New("zone message filter must allow at least one message type")
	}
	seen := make(map[string]struct{}, len(f.AllowedMessageTypes))
	var previous string
	for i, item := range f.AllowedMessageTypes {
		if item == "*" {
			if len(f.AllowedMessageTypes) != 1 {
				return errors.New("wildcard zone message filter must be the only entry")
			}
			return nil
		}
		if err := validateRuntimeToken("zone message type", item, MaxZoneMessageTypeLength); err != nil {
			return err
		}
		if _, found := seen[item]; found {
			return fmt.Errorf("duplicate zone message type %q", item)
		}
		seen[item] = struct{}{}
		if i > 0 && previous >= item {
			return errors.New("zone message filter must be sorted canonically")
		}
		previous = item
	}
	return nil
}

func (f ZoneMessageFilter) Allows(messageType string) bool {
	for _, allowed := range f.AllowedMessageTypes {
		if allowed == "*" || allowed == messageType {
			return true
		}
	}
	return false
}

func (m ZoneMessage) Validate(expectedZone ZoneID) error {
	if err := ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if m.ZoneID != expectedZone {
		return fmt.Errorf("zone message belongs to %s, expected %s", m.ZoneID, expectedZone)
	}
	if err := validateRuntimeToken("zone message type", m.MessageType, MaxZoneMessageTypeLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("zone message source", m.Source, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("zone message destination", m.Destination, MaxZoneEndpointLength); err != nil {
		return err
	}
	if m.GasLimit == 0 {
		return errors.New("zone message gas limit must be positive")
	}
	if err := ValidateHash("zone message payload hash", m.PayloadHash); err != nil {
		return err
	}
	return nil
}

func ZoneStateNamespace(id ZoneID) string {
	return ZoneStateNamespacePrefix + string(id)
}

func ZoneQueueNamespace(id ZoneID) string {
	return ZoneQueueNamespacePrefix + string(id)
}

func ZoneProofNamespace(id ZoneID) string {
	return ZoneProofNamespacePrefix + string(id)
}

func ZoneQueryNamespace(id ZoneID) string {
	return ZoneQueryNamespacePrefix + string(id)
}

func ZoneKVPrefix(id ZoneID) string {
	return ZoneKVPrefixRoot + string(id) + "/"
}

func ZoneExecutionPipeline(id ZoneID) string {
	return ZonePipelinePrefix + string(id)
}

func DefaultZoneModuleSet(zone Zone) []string {
	modules := []string{
		"kind:" + string(zone.Kind),
		"transition:" + zone.StateTransitionID,
		"vm:" + string(zone.VMPolicy),
	}
	sort.Strings(modules)
	return modules
}

func EmptyRootHash() string {
	sum := sha256.Sum256(nil)
	return hex.EncodeToString(sum[:])
}

func ComputeZoneMessageRoot(messages []ZoneMessage) string {
	queue := cloneZoneMessages(messages)
	sortZoneMessages(queue)
	h := sha256.New()
	writeRuntimePart(h, "aetra-zone-message-root-v1")
	writeRuntimeUint64(h, uint64(len(queue)))
	for _, msg := range queue {
		writeRuntimePart(h, string(msg.ZoneID))
		writeRuntimePart(h, msg.MessageType)
		writeRuntimePart(h, msg.Source)
		writeRuntimePart(h, msg.Destination)
		writeRuntimeUint64(h, msg.GasLimit)
		writeRuntimePart(h, msg.PayloadHash)
		writeRuntimeUint64(h, msg.Sequence)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeZoneRuntimeProofRoot(runtime ZoneRuntimeState) string {
	h := sha256.New()
	writeRuntimePart(h, "aetra-zone-runtime-proof-v1")
	writeRuntimePart(h, string(runtime.ZoneID))
	writeRuntimePart(h, runtime.StoreKey)
	writeRuntimePart(h, runtime.StateNamespace)
	writeRuntimePart(h, runtime.QueueNamespace)
	writeRuntimePart(h, runtime.ProofNamespace)
	writeRuntimePart(h, runtime.QueryNamespace)
	writeRuntimePart(h, runtime.KVPrefix)
	writeRuntimePart(h, runtime.ModuleSetRoot)
	writeRuntimePart(h, runtime.ExecutionPipeline)
	writeRuntimePart(h, runtime.StateRoot)
	writeRuntimePart(h, runtime.ReceiptRoot)
	writeRuntimePart(h, runtime.MessageRoot)
	writeRuntimePart(h, runtime.ExecutionResultRoot)
	writeRuntimeUint64(h, runtime.Budget.MaxGas)
	writeRuntimeUint64(h, runtime.Budget.GasUsed)
	writeRuntimeUint64(h, uint64(runtime.Budget.MaxMessages))
	writeRuntimeUint64(h, uint64(runtime.Budget.MessagesUsed))
	writeRuntimePart(h, runtime.GasPolicy.Denom)
	writeRuntimeUint64(h, runtime.GasPolicy.MaxGasPerBlock)
	writeRuntimeUint64(h, runtime.GasPolicy.MaxGasPerMessage)
	filter := append([]string(nil), runtime.MessageFilter.AllowedMessageTypes...)
	sort.Strings(filter)
	for _, item := range filter {
		writeRuntimePart(h, item)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeZoneModuleSetRoot(modules []string) string {
	ordered := append([]string(nil), modules...)
	sort.Strings(ordered)
	h := sha256.New()
	writeRuntimePart(h, "aetra-zone-module-set-v1")
	writeRuntimeUint64(h, uint64(len(ordered)))
	for _, module := range ordered {
		writeRuntimePart(h, module)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func BuildCommitmentFromRuntime(runtime ZoneRuntimeState, height uint64, previousCommitment string) (ZoneCommitment, error) {
	if err := runtime.Validate(); err != nil {
		return ZoneCommitment{}, err
	}
	return NewZoneCommitment(
		runtime.ZoneID,
		height,
		runtime.StateRoot,
		runtime.ReceiptRoot,
		runtime.MessageRoot,
		runtime.ExecutionResultRoot,
		previousCommitment,
	)
}

func validateZoneMessageQueue(zoneID ZoneID, queue []ZoneMessage, budget ZoneExecutionBudget, gasPolicy ZoneGasPolicy, filter ZoneMessageFilter) error {
	var previous ZoneMessage
	for i, msg := range queue {
		if err := msg.Validate(zoneID); err != nil {
			return err
		}
		if !filter.Allows(msg.MessageType) {
			return fmt.Errorf("zone message type %q is not allowed", msg.MessageType)
		}
		if msg.GasLimit > gasPolicy.MaxGasPerMessage {
			return errors.New("zone message gas exceeds per-message policy")
		}
		if i > 0 && compareZoneMessages(previous, msg) >= 0 {
			return errors.New("zone message queue must be sorted canonically")
		}
		previous = msg
	}
	gas, err := queueGas(queue)
	if err != nil {
		return err
	}
	consumed, err := budget.Consume(gas, uint32(len(queue)))
	if err != nil {
		return err
	}
	_ = consumed
	return nil
}

func validateZoneModuleSet(modules []string) error {
	if len(modules) == 0 {
		return errors.New("zone module set must contain at least one module")
	}
	seen := make(map[string]struct{}, len(modules))
	var previous string
	for i, module := range modules {
		if err := validateRuntimeToken("zone module", module, MaxZoneModuleNameLength); err != nil {
			return err
		}
		if _, found := seen[module]; found {
			return fmt.Errorf("duplicate zone module %q", module)
		}
		seen[module] = struct{}{}
		if i > 0 && previous >= module {
			return errors.New("zone module set must be sorted canonically")
		}
		previous = module
	}
	return nil
}

func queueGas(queue []ZoneMessage) (uint64, error) {
	var total uint64
	for _, msg := range queue {
		if msg.GasLimit > ^uint64(0)-total {
			return 0, errors.New("zone message gas overflow")
		}
		total += msg.GasLimit
	}
	return total, nil
}

func cloneZoneMessages(messages []ZoneMessage) []ZoneMessage {
	out := append([]ZoneMessage(nil), messages...)
	sortZoneMessages(out)
	return out
}

func sortZoneMessages(messages []ZoneMessage) {
	sort.SliceStable(messages, func(i, j int) bool {
		return compareZoneMessages(messages[i], messages[j]) < 0
	})
}

func compareZoneMessages(left, right ZoneMessage) int {
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	if left.MessageType < right.MessageType {
		return -1
	}
	if left.MessageType > right.MessageType {
		return 1
	}
	if left.Destination < right.Destination {
		return -1
	}
	if left.Destination > right.Destination {
		return 1
	}
	return 0
}

func validateZoneNamespace(fieldName, value, expected string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > MaxZoneNamespaceLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxZoneNamespaceLength)
	}
	if value != expected {
		return fmt.Errorf("%s must be %q", fieldName, expected)
	}
	return nil
}

func validateRuntimeToken(fieldName, value string, maxLen int) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxLen)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func prefixesOverlap(left, right string) bool {
	return strings.HasPrefix(left, right) || strings.HasPrefix(right, left)
}

func writeRuntimePart(w byteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func writeRuntimeUint64(w byteWriter, value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	_, _ = w.Write(out[:])
}
