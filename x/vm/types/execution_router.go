package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	RouterLaneSync	RouterLane	= "sync_engine"
	RouterLaneAsync	RouterLane	= "async_engine"
	RouterLaneActor	RouterLane	= "actor_runtime"

	RouterBackendNativeModule	RouterBackend	= "native_module"
	RouterBackendAVMActor		RouterBackend	= "avm_actor"
	RouterBackendWASMAdapter	RouterBackend	= "wasm_adapter"

	RouterGasClassLow	RouterGasClass	= "low"
	RouterGasClassStandard	RouterGasClass	= "standard"
	RouterGasClassHigh	RouterGasClass	= "high"

	RouterDispatchModeDirect	RouterDispatchMode	= "direct"
	RouterDispatchModeQueued	RouterDispatchMode	= "queued"
	RouterDispatchModeCrossZone	RouterDispatchMode	= "cross_zone_async"

	RouterReceiptCommit	RouterReceiptPolicy	= "commit"
	RouterReceiptDeferred	RouterReceiptPolicy	= "deferred"
	RouterReceiptDeadLetter	RouterReceiptPolicy	= "dead_letter"

	RouterMessageFinancialPrefix	= "financial."
	RouterMessageIdentityPrefix	= "identity."
	RouterMessageApplicationPrefix	= "application."
	RouterMessageContractPrefix	= "contract."
	RouterMessageAsyncPrefix	= "async."
	RouterMessageActorPrefix	= "actor."
	RouterMessageResolverPrefix	= "resolver."
	RouterMessageSchedulerPrefix	= "scheduler."
	RouterMessageWorkflowPrefix	= "workflow."

	MaxRouterPriority	= 255
	MaxRouterRouteKeyLength	= 128
	MaxRouterTargetLength	= 128
)

type RouterLane string
type RouterBackend string
type RouterGasClass string
type RouterDispatchMode string
type RouterReceiptPolicy string

type routerByteWriter interface {
	Write([]byte) (int, error)
}

type RouterSchedulingMetadata struct {
	DeliverAtBlock		uint64
	DeadlineBlock		uint64
	RetryCount		uint32
	MaxRetries		uint32
	RetryDelayBlocks	uint64
	ContinuationToken	string
}

type RouterGasMeter struct {
	Class		RouterGasClass
	Limit		uint64
	Reserved	uint64
}

type ExecutionRouterMessage struct {
	Sequence	uint64
	MsgType		string
	SourceZoneID	zonestypes.ZoneID
	ZoneID		zonestypes.ZoneID
	TargetModule	string
	TargetActor	string
	TargetContract	string
	Source		string
	Destination	string
	PayloadHash	string
	Call		VMCall
	Actor		bool
	Backend		RouterBackend
	GasClass	RouterGasClass
	Priority	uint8
	Scheduling	RouterSchedulingMetadata
	DomainRouteKey	string
	CrossZoneWrite	bool
	BlockSTMKey	string
	StakingPower	uint64
}

type ExecutionRouterDispatch struct {
	Sequence	uint64
	ZoneID		zonestypes.ZoneID
	MsgType		string
	ExecutionTarget	string
	QueueID		string
	GasMeter	RouterGasMeter
	Lane		RouterLane
	Backend		RouterBackend
	DispatchMode	RouterDispatchMode
	ReceiptPolicy	RouterReceiptPolicy
	Priority	uint8
	Scheduling	RouterSchedulingMetadata
	DomainRouteKey	string
	Call		VMCall
	KVPrefix	string
	BlockSTMKey	string
	StakingPower	uint64
	ExecutionHeight	uint64
}

type ExecutionRouterZoneOutput struct {
	ZoneID			zonestypes.ZoneID
	StateRoot		string
	ReceiptRoot		string
	MessageRoot		string
	ExecutionResultRoot	string
	ProofRoot		string
	Budget			zonestypes.ZoneExecutionBudget
}

type ExecutionRouterPlan struct {
	Height		uint64
	SDKPlan		FinalizeBlockPlan
	Dispatches	[]ExecutionRouterDispatch
	ZoneOutputs	[]ExecutionRouterZoneOutput
	PlanRoot	string
}

func BuildExecutionRouterPlan(
	height uint64,
	binding SDKIntegrationBinding,
	zones []zonestypes.ZoneRuntimeState,
	messages []ExecutionRouterMessage,
	policy RuntimePolicy,
) (ExecutionRouterPlan, error) {
	if height == 0 {
		return ExecutionRouterPlan{}, errors.New("AVM execution router height must be positive")
	}
	if err := binding.Validate(); err != nil {
		return ExecutionRouterPlan{}, err
	}
	zoneByID, budgets, err := indexRouterZones(zones)
	if err != nil {
		return ExecutionRouterPlan{}, err
	}
	dispatches := make([]ExecutionRouterDispatch, 0, len(messages))
	sdkDispatches := make([]SDKDispatch, 0, len(messages))
	for _, msg := range messages {
		dispatch, nextBudget, err := routeExecutionMessage(height, msg, zoneByID, budgets, policy)
		if err != nil {
			return ExecutionRouterPlan{}, err
		}
		budgets[dispatch.ZoneID] = nextBudget
		dispatches = append(dispatches, dispatch)
		sdkDispatches = append(sdkDispatches, SDKDispatch{
			ZoneID:			dispatch.ZoneID,
			MsgType:		dispatch.MsgType,
			Call:			dispatch.Call,
			KVPrefix:		dispatch.KVPrefix,
			BlockSTMKey:		dispatch.BlockSTMKey,
			StakingPower:		dispatch.StakingPower,
			ExecutionHeight:	dispatch.ExecutionHeight,
		})
	}
	sdkPlan, err := BuildFinalizeBlockPlan(height, binding, sdkDispatches, policy)
	if err != nil {
		return ExecutionRouterPlan{}, err
	}
	sort.SliceStable(dispatches, func(i, j int) bool {
		return compareRouterDispatch(dispatches[i], dispatches[j]) < 0
	})
	outputs := buildRouterZoneOutputs(zones, budgets)
	plan := ExecutionRouterPlan{Height: height, SDKPlan: sdkPlan, Dispatches: dispatches, ZoneOutputs: outputs}
	plan.PlanRoot = ComputeExecutionRouterPlanRoot(plan)
	return plan, plan.Validate(policy)
}

func (p ExecutionRouterPlan) Validate(policy RuntimePolicy) error {
	if p.Height == 0 {
		return errors.New("AVM execution router plan height must be positive")
	}
	if err := p.SDKPlan.Validate(policy); err != nil {
		return err
	}
	if p.SDKPlan.Height != p.Height {
		return errors.New("AVM execution router SDK plan height drift")
	}
	if len(p.Dispatches) != len(p.SDKPlan.Dispatches) {
		return errors.New("AVM execution router SDK dispatch count mismatch")
	}
	for i, dispatch := range p.Dispatches {
		if err := dispatch.Validate(policy); err != nil {
			return err
		}
		if i > 0 && compareRouterDispatch(p.Dispatches[i-1], dispatch) >= 0 {
			return errors.New("AVM execution router dispatches must be sorted canonically")
		}
		sdkDispatch := p.SDKPlan.Dispatches[i]
		if dispatch.ZoneID != sdkDispatch.ZoneID ||
			dispatch.MsgType != sdkDispatch.MsgType ||
			dispatch.BlockSTMKey != sdkDispatch.BlockSTMKey ||
			dispatch.ExecutionHeight != sdkDispatch.ExecutionHeight {
			return errors.New("AVM execution router SDK dispatch drift")
		}
	}
	seenZones := make(map[zonestypes.ZoneID]struct{}, len(p.ZoneOutputs))
	for i, output := range p.ZoneOutputs {
		if err := output.Validate(); err != nil {
			return err
		}
		if _, found := seenZones[output.ZoneID]; found {
			return fmt.Errorf("duplicate AVM execution router zone output %s", output.ZoneID)
		}
		seenZones[output.ZoneID] = struct{}{}
		if i > 0 && p.ZoneOutputs[i-1].ZoneID >= output.ZoneID {
			return errors.New("AVM execution router zone outputs must be sorted canonically")
		}
	}
	if p.PlanRoot != ComputeExecutionRouterPlanRoot(p) {
		return errors.New("AVM execution router plan root mismatch")
	}
	return nil
}

func (d ExecutionRouterDispatch) Validate(policy RuntimePolicy) error {
	if err := zonestypes.ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if err := validateSDKToken("AVM execution router message type", d.MsgType); err != nil {
		return err
	}
	if !IsRouterLane(d.Lane) {
		return fmt.Errorf("invalid AVM execution router lane %q", d.Lane)
	}
	if !IsRouterBackend(d.Backend) {
		return fmt.Errorf("invalid AVM execution router backend %q", d.Backend)
	}
	if err := validateRouterTarget("AVM execution router execution target", d.ExecutionTarget); err != nil {
		return err
	}
	if err := validateRouterQueueID(d.QueueID, d.DispatchMode); err != nil {
		return err
	}
	if err := d.GasMeter.Validate(); err != nil {
		return err
	}
	if !IsRouterDispatchMode(d.DispatchMode) {
		return fmt.Errorf("invalid AVM execution router dispatch mode %q", d.DispatchMode)
	}
	if !IsRouterReceiptPolicy(d.ReceiptPolicy) {
		return fmt.Errorf("invalid AVM execution router receipt policy %q", d.ReceiptPolicy)
	}
	if err := d.Scheduling.Validate(); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM execution router domain route key", d.DomainRouteKey, MaxRouterRouteKeyLength); err != nil {
		return err
	}
	if d.KVPrefix != ContractZoneKVPrefix(d.ZoneID) {
		return fmt.Errorf("AVM execution router KV prefix must be %q", ContractZoneKVPrefix(d.ZoneID))
	}
	if strings.TrimSpace(d.BlockSTMKey) != d.BlockSTMKey || d.BlockSTMKey == "" {
		return errors.New("AVM execution router BlockSTM key is required")
	}
	if !strings.HasPrefix(d.BlockSTMKey, d.KVPrefix) {
		return errors.New("AVM execution router BlockSTM key must stay within zone KV prefix")
	}
	if d.ExecutionHeight == 0 {
		return errors.New("AVM execution router execution height must be positive")
	}
	if d.StakingPower == 0 {
		return errors.New("AVM execution router requires staking voting power")
	}
	return ValidateVMCall(d.Call, policy)
}

func (m ExecutionRouterMessage) ValidateEnvelope(policy RuntimePolicy) error {
	if m.Sequence == 0 {
		return errors.New("AVM execution router message sequence must be positive")
	}
	if err := validateSDKToken("AVM execution router message type", m.MsgType); err != nil {
		return err
	}
	if m.SourceZoneID != "" {
		if err := zonestypes.ValidateZoneID(m.SourceZoneID); err != nil {
			return err
		}
	}
	if m.ZoneID != "" {
		if err := zonestypes.ValidateZoneID(m.ZoneID); err != nil {
			return err
		}
	}
	if err := validateRouterTarget("AVM execution router source", m.Source); err != nil {
		return err
	}
	if err := validateRouterTarget("AVM execution router destination", m.Destination); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM execution router target module", m.TargetModule, MaxRouterTargetLength); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM execution router target actor", m.TargetActor, MaxRouterTargetLength); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM execution router target contract", m.TargetContract, MaxRouterTargetLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM execution router payload hash", m.PayloadHash); err != nil {
		return err
	}
	if !IsRouterGasClass(m.effectiveGasClass()) {
		return fmt.Errorf("invalid AVM execution router gas class %q", m.GasClass)
	}
	if err := m.Scheduling.Validate(); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM execution router domain route key", m.DomainRouteKey, MaxRouterRouteKeyLength); err != nil {
		return err
	}
	if m.Priority > MaxRouterPriority {
		return errors.New("AVM execution router priority exceeds maximum")
	}
	return nil
}

func (m ExecutionRouterMessage) effectiveGasClass() RouterGasClass {
	if m.GasClass == "" {
		return RouterGasClassStandard
	}
	return m.GasClass
}

func (g RouterGasMeter) Validate() error {
	if !IsRouterGasClass(g.Class) {
		return fmt.Errorf("invalid AVM execution router gas class %q", g.Class)
	}
	if g.Limit == 0 {
		return errors.New("AVM execution router gas meter limit must be positive")
	}
	if g.Reserved == 0 {
		return errors.New("AVM execution router gas meter reserved gas must be positive")
	}
	if g.Reserved > g.Limit {
		return errors.New("AVM execution router reserved gas exceeds limit")
	}
	return nil
}

func (s RouterSchedulingMetadata) Validate() error {
	if s.DeadlineBlock != 0 && s.DeliverAtBlock != 0 && s.DeadlineBlock < s.DeliverAtBlock {
		return errors.New("AVM execution router deadline must not be before delivery block")
	}
	if s.RetryCount > s.MaxRetries {
		return errors.New("AVM execution router retry count exceeds max retries")
	}
	if s.MaxRetries > 0 && s.RetryDelayBlocks == 0 {
		return errors.New("AVM execution router retry delay must be positive when retries are enabled")
	}
	return validateRouterOptionalToken("AVM execution router continuation token", s.ContinuationToken, MaxRouterRouteKeyLength)
}

func (o ExecutionRouterZoneOutput) Validate() error {
	if err := zonestypes.ValidateZoneID(o.ZoneID); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM execution router state root", value: o.StateRoot},
		{name: "AVM execution router receipt root", value: o.ReceiptRoot},
		{name: "AVM execution router message root", value: o.MessageRoot},
		{name: "AVM execution router execution result root", value: o.ExecutionResultRoot},
		{name: "AVM execution router proof root", value: o.ProofRoot},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	return o.Budget.Validate()
}

func ClassifyExecutionZone(msg ExecutionRouterMessage) (zonestypes.ZoneID, error) {
	if msg.ZoneID != "" {
		return msg.ZoneID, zonestypes.ValidateZoneID(msg.ZoneID)
	}
	switch {
	case strings.HasPrefix(msg.MsgType, RouterMessageFinancialPrefix):
		return zonestypes.ZoneIDFinancial, nil
	case strings.HasPrefix(msg.MsgType, RouterMessageIdentityPrefix),
		strings.HasPrefix(msg.MsgType, RouterMessageResolverPrefix):
		return zonestypes.ZoneIDIdentity, nil
	case strings.HasPrefix(msg.MsgType, RouterMessageApplicationPrefix),
		strings.HasPrefix(msg.MsgType, RouterMessageAsyncPrefix),
		strings.HasPrefix(msg.MsgType, RouterMessageActorPrefix),
		strings.HasPrefix(msg.MsgType, RouterMessageSchedulerPrefix),
		strings.HasPrefix(msg.MsgType, RouterMessageWorkflowPrefix):
		return zonestypes.ZoneIDApplication, nil
	case strings.HasPrefix(msg.MsgType, RouterMessageContractPrefix),
		msg.Call.Runtime == RuntimeAVM,
		msg.Call.Runtime == RuntimeCosmWasm,
		msg.Backend == RouterBackendAVMActor,
		msg.Backend == RouterBackendWASMAdapter:
		return zonestypes.ZoneIDContract, nil
	default:
		return "", fmt.Errorf("AVM execution router cannot classify message type %q", msg.MsgType)
	}
}

func RouterLaneForMessage(msg ExecutionRouterMessage) RouterLane {
	if IsCrossZoneWrite(msg) {
		return RouterLaneAsync
	}
	if msg.Actor || strings.HasPrefix(msg.MsgType, RouterMessageActorPrefix) {
		return RouterLaneActor
	}
	if strings.HasPrefix(msg.MsgType, RouterMessageAsyncPrefix) {
		return RouterLaneAsync
	}
	switch msg.Call.Action {
	case ActionInternalCall, ActionBouncedCall:
		return RouterLaneAsync
	default:
		return RouterLaneSync
	}
}

func RouterDispatchModeForMessage(msg ExecutionRouterMessage) RouterDispatchMode {
	if IsCrossZoneWrite(msg) {
		return RouterDispatchModeCrossZone
	}
	if RouterLaneForMessage(msg) == RouterLaneSync {
		return RouterDispatchModeDirect
	}
	return RouterDispatchModeQueued
}

func RouterReceiptPolicyForMode(mode RouterDispatchMode) RouterReceiptPolicy {
	switch mode {
	case RouterDispatchModeDirect:
		return RouterReceiptCommit
	case RouterDispatchModeQueued, RouterDispatchModeCrossZone:
		return RouterReceiptDeferred
	default:
		return RouterReceiptDeadLetter
	}
}

func IsCrossZoneWrite(msg ExecutionRouterMessage) bool {
	if msg.CrossZoneWrite {
		return true
	}
	if msg.SourceZoneID == "" || msg.ZoneID == "" || msg.SourceZoneID == msg.ZoneID {
		return false
	}
	return msg.Call.Action != ActionQuery
}

func RouterBackendForMessage(msg ExecutionRouterMessage) (RouterBackend, error) {
	if msg.Backend != "" {
		if !IsRouterBackend(msg.Backend) {
			return "", fmt.Errorf("invalid AVM execution router backend %q", msg.Backend)
		}
		return msg.Backend, nil
	}
	switch msg.Call.Runtime {
	case RuntimeAVM:
		return RouterBackendAVMActor, nil
	case RuntimeCosmWasm:
		return RouterBackendWASMAdapter, nil
	default:
		return RouterBackendNativeModule, nil
	}
}

func RouterExecutionTarget(msg ExecutionRouterMessage) string {
	for _, value := range []string{msg.TargetContract, msg.TargetActor, msg.TargetModule, msg.Destination} {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return msg.MsgType
}

func RouterQueueID(zoneID zonestypes.ZoneID, mode RouterDispatchMode, routeKey string) string {
	if mode == RouterDispatchModeDirect {
		return "direct"
	}
	key := strings.TrimSpace(routeKey)
	if key == "" {
		key = "default"
	}
	return fmt.Sprintf("%squeue/%s/%s", ContractZoneKVPrefix(zoneID), mode, key)
}

func IsRouterLane(lane RouterLane) bool {
	switch lane {
	case RouterLaneSync, RouterLaneAsync, RouterLaneActor:
		return true
	default:
		return false
	}
}

func IsRouterGasClass(gasClass RouterGasClass) bool {
	switch gasClass {
	case RouterGasClassLow, RouterGasClassStandard, RouterGasClassHigh:
		return true
	default:
		return false
	}
}

func IsRouterDispatchMode(mode RouterDispatchMode) bool {
	switch mode {
	case RouterDispatchModeDirect, RouterDispatchModeQueued, RouterDispatchModeCrossZone:
		return true
	default:
		return false
	}
}

func IsRouterReceiptPolicy(policy RouterReceiptPolicy) bool {
	switch policy {
	case RouterReceiptCommit, RouterReceiptDeferred, RouterReceiptDeadLetter:
		return true
	default:
		return false
	}
}

func IsRouterBackend(backend RouterBackend) bool {
	switch backend {
	case RouterBackendNativeModule, RouterBackendAVMActor, RouterBackendWASMAdapter:
		return true
	default:
		return false
	}
}

func ComputeExecutionRouterPlanRoot(plan ExecutionRouterPlan) string {
	h := sha256.New()
	writeRouterPart(h, "aetra-avm-router-plan-v1")
	writeRouterUint64(h, plan.Height)
	writeRouterUint64(h, uint64(len(plan.Dispatches)))
	for _, dispatch := range plan.Dispatches {
		writeRouterUint64(h, dispatch.Sequence)
		writeRouterPart(h, string(dispatch.ZoneID))
		writeRouterPart(h, dispatch.MsgType)
		writeRouterPart(h, dispatch.ExecutionTarget)
		writeRouterPart(h, dispatch.QueueID)
		writeRouterPart(h, string(dispatch.GasMeter.Class))
		writeRouterUint64(h, dispatch.GasMeter.Limit)
		writeRouterUint64(h, dispatch.GasMeter.Reserved)
		writeRouterPart(h, string(dispatch.Lane))
		writeRouterPart(h, string(dispatch.Backend))
		writeRouterPart(h, string(dispatch.DispatchMode))
		writeRouterPart(h, string(dispatch.ReceiptPolicy))
		writeRouterUint64(h, uint64(dispatch.Priority))
		writeRouterUint64(h, dispatch.Scheduling.DeliverAtBlock)
		writeRouterUint64(h, dispatch.Scheduling.DeadlineBlock)
		writeRouterUint64(h, uint64(dispatch.Scheduling.RetryCount))
		writeRouterUint64(h, uint64(dispatch.Scheduling.MaxRetries))
		writeRouterUint64(h, dispatch.Scheduling.RetryDelayBlocks)
		writeRouterPart(h, dispatch.Scheduling.ContinuationToken)
		writeRouterPart(h, dispatch.DomainRouteKey)
		writeRouterPart(h, dispatch.Call.Runtime)
		writeRouterPart(h, dispatch.Call.Action)
		writeRouterUint64(h, dispatch.Call.GasLimit)
		writeRouterPart(h, dispatch.KVPrefix)
		writeRouterPart(h, dispatch.BlockSTMKey)
		writeRouterUint64(h, dispatch.StakingPower)
		writeRouterUint64(h, dispatch.ExecutionHeight)
	}
	writeRouterUint64(h, uint64(len(plan.ZoneOutputs)))
	for _, output := range plan.ZoneOutputs {
		writeRouterPart(h, string(output.ZoneID))
		writeRouterPart(h, output.StateRoot)
		writeRouterPart(h, output.ReceiptRoot)
		writeRouterPart(h, output.MessageRoot)
		writeRouterPart(h, output.ExecutionResultRoot)
		writeRouterPart(h, output.ProofRoot)
		writeRouterUint64(h, output.Budget.MaxGas)
		writeRouterUint64(h, output.Budget.GasUsed)
		writeRouterUint64(h, uint64(output.Budget.MaxMessages))
		writeRouterUint64(h, uint64(output.Budget.MessagesUsed))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func routeExecutionMessage(
	height uint64,
	msg ExecutionRouterMessage,
	zones map[zonestypes.ZoneID]zonestypes.ZoneRuntimeState,
	budgets map[zonestypes.ZoneID]zonestypes.ZoneExecutionBudget,
	policy RuntimePolicy,
) (ExecutionRouterDispatch, zonestypes.ZoneExecutionBudget, error) {
	if err := msg.ValidateEnvelope(policy); err != nil {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, err
	}
	zoneID, err := ClassifyExecutionZone(msg)
	if err != nil {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, err
	}
	msg.ZoneID = zoneID
	zone, found := zones[zoneID]
	if !found {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, fmt.Errorf("AVM execution router zone %s is not active", zoneID)
	}
	if !zone.MessageFilter.Allows(msg.MsgType) {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, fmt.Errorf("AVM execution router message type %q is not allowed in zone %s", msg.MsgType, zoneID)
	}
	if err := ValidateVMCall(msg.Call, policy); err != nil {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, err
	}
	if msg.Call.GasLimit > zone.GasPolicy.MaxGasPerMessage {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, errors.New("AVM execution router message gas exceeds per-message policy")
	}
	nextBudget, err := budgets[zoneID].Consume(msg.Call.GasLimit, 1)
	if err != nil {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, err
	}
	backend, err := RouterBackendForMessage(msg)
	if err != nil {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, err
	}
	blockSTMKey := strings.TrimSpace(msg.BlockSTMKey)
	kvPrefix := ContractZoneKVPrefix(zoneID)
	if blockSTMKey == "" {
		blockSTMKey = fmt.Sprintf("%srouter/%020d", kvPrefix, msg.Sequence)
	}
	mode := RouterDispatchModeForMessage(msg)
	dispatch := ExecutionRouterDispatch{
		Sequence:		msg.Sequence,
		ZoneID:			zoneID,
		MsgType:		msg.MsgType,
		ExecutionTarget:	RouterExecutionTarget(msg),
		QueueID:		RouterQueueID(zoneID, mode, msg.DomainRouteKey),
		GasMeter: RouterGasMeter{
			Class:		msg.effectiveGasClass(),
			Limit:		msg.Call.GasLimit,
			Reserved:	msg.Call.GasLimit,
		},
		Lane:			RouterLaneForMessage(msg),
		Backend:		backend,
		DispatchMode:		mode,
		ReceiptPolicy:		RouterReceiptPolicyForMode(mode),
		Priority:		msg.Priority,
		Scheduling:		msg.Scheduling,
		DomainRouteKey:		strings.TrimSpace(msg.DomainRouteKey),
		Call:			msg.Call,
		KVPrefix:		kvPrefix,
		BlockSTMKey:		blockSTMKey,
		StakingPower:		msg.StakingPower,
		ExecutionHeight:	height,
	}
	if err := dispatch.Validate(policy); err != nil {
		return ExecutionRouterDispatch{}, zonestypes.ZoneExecutionBudget{}, err
	}
	return dispatch, nextBudget, nil
}

func indexRouterZones(zones []zonestypes.ZoneRuntimeState) (map[zonestypes.ZoneID]zonestypes.ZoneRuntimeState, map[zonestypes.ZoneID]zonestypes.ZoneExecutionBudget, error) {
	zoneByID := make(map[zonestypes.ZoneID]zonestypes.ZoneRuntimeState, len(zones))
	budgets := make(map[zonestypes.ZoneID]zonestypes.ZoneExecutionBudget, len(zones))
	for _, zone := range zones {
		if err := zone.Validate(); err != nil {
			return nil, nil, err
		}
		if _, found := zoneByID[zone.ZoneID]; found {
			return nil, nil, fmt.Errorf("duplicate AVM execution router zone %s", zone.ZoneID)
		}
		zoneByID[zone.ZoneID] = zone.Clone()
		budgets[zone.ZoneID] = zone.Budget
	}
	return zoneByID, budgets, nil
}

func buildRouterZoneOutputs(zones []zonestypes.ZoneRuntimeState, budgets map[zonestypes.ZoneID]zonestypes.ZoneExecutionBudget) []ExecutionRouterZoneOutput {
	outputs := make([]ExecutionRouterZoneOutput, 0, len(zones))
	for _, zone := range zones {
		outputs = append(outputs, ExecutionRouterZoneOutput{
			ZoneID:			zone.ZoneID,
			StateRoot:		zone.StateRoot,
			ReceiptRoot:		zone.ReceiptRoot,
			MessageRoot:		zone.MessageRoot,
			ExecutionResultRoot:	zone.ExecutionResultRoot,
			ProofRoot:		zone.ProofRoot,
			Budget:			budgets[zone.ZoneID],
		})
	}
	sort.SliceStable(outputs, func(i, j int) bool {
		return outputs[i].ZoneID < outputs[j].ZoneID
	})
	return outputs
}

func compareRouterDispatch(left, right ExecutionRouterDispatch) int {
	if c := compareDispatch(
		SDKDispatch{ZoneID: left.ZoneID, MsgType: left.MsgType, BlockSTMKey: left.BlockSTMKey},
		SDKDispatch{ZoneID: right.ZoneID, MsgType: right.MsgType, BlockSTMKey: right.BlockSTMKey},
	); c != 0 {
		return c
	}
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	return 0
}

func validateRouterTarget(fieldName, value string) error {
	return validateRouterOptionalToken(fieldName, value, MaxRouterTargetLength)
}

func validateRouterQueueID(queueID string, mode RouterDispatchMode) error {
	if mode == RouterDispatchModeDirect {
		if queueID != "direct" {
			return errors.New("AVM execution router direct dispatch must use direct queue id")
		}
		return nil
	}
	return validateRouterOptionalToken("AVM execution router queue id", queueID, MaxRouterTargetLength*2)
}

func validateRouterOptionalToken(fieldName, value string, maxLen int) error {
	trimmed := strings.TrimSpace(value)
	if trimmed != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if trimmed == "" {
		return nil
	}
	if len(trimmed) > maxLen {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxLen)
	}
	for _, r := range trimmed {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func writeRouterPart(w routerByteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func writeRouterUint64(w routerByteWriter, value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	_, _ = w.Write(out[:])
}
