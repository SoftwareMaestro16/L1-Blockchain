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
	AVMUpgradeCompatibilityNone		AVMUpgradeCompatibilityMode	= "none"
	AVMUpgradeCompatibilityVersionedPolicy	AVMUpgradeCompatibilityMode	= "versioned_policy"
	AVMUpgradeCompatibilityLegacyRuntime	AVMUpgradeCompatibilityMode	= "legacy_runtime"

	AVMUpgradeStatusScheduled	AVMUpgradeStatus	= "scheduled"
	AVMUpgradeStatusActive		AVMUpgradeStatus	= "active"
	AVMUpgradeStatusCompleted	AVMUpgradeStatus	= "completed"
	AVMUpgradeStatusFailed		AVMUpgradeStatus	= "failed"

	AVMUpgradeMigrationQueue	AVMUpgradeMigrationKind	= "queue"
	AVMUpgradeMigrationContinuation	AVMUpgradeMigrationKind	= "continuation"

	MaxAVMUpgradeStates		= 4096
	MaxAVMUpgradeMigrationHandlers	= 256
)

type AVMUpgradeCompatibilityMode string
type AVMUpgradeStatus string
type AVMUpgradeMigrationKind string

type AVMRuntimeVersionSet struct {
	VMInterpreterVersion	string
	SchedulerVersion	string
	GasPolicyVersion	string
	ZoneConfigVersion	string
	BackendAdapterVersion	string
	InterfaceSchemaVersion	string
	RetryPolicyVersion	string
	QueueLimitVersion	string
	RuntimeVersionSetHash	string
}

type AVMScheduledUpgradeState struct {
	UpgradeID		string
	Component		AVMUpgradeComponentKind
	FromVersion		string
	ToVersion		string
	ActivationHeight	uint64
	MigrationRequired	bool
	CompatibilityMode	AVMUpgradeCompatibilityMode
	Status			AVMUpgradeStatus
	StateHash		string
}

type AVMUpgradeMigrationHandler struct {
	UpgradeID	string
	Kind		AVMUpgradeMigrationKind
	FromVersion	string
	ToVersion	string
	HandlerID	string
	SourceRoot	string
	TargetRoot	string
	MigratedCount	uint32
	Bounded		bool
	HandlerHash	string
}

type AVMVersionedGasTable struct {
	Tables		[]AVMGasTableActivation
	TableRoot	string
}

type AVMUpgradeStateRegistry struct {
	RuntimeVersions		AVMRuntimeVersionSet
	States			[]AVMScheduledUpgradeState
	MigrationHandlers	[]AVMUpgradeMigrationHandler
	GasTable		AVMVersionedGasTable
	RegistryHash		string
}

func NewAVMRuntimeVersionSet(versions AVMRuntimeVersionSet) (AVMRuntimeVersionSet, error) {
	versions = canonicalAVMRuntimeVersionSet(versions)
	versions.RuntimeVersionSetHash = ComputeAVMRuntimeVersionSetHash(versions)
	return versions, versions.Validate()
}

func (v AVMRuntimeVersionSet) Validate() error {
	v = canonicalAVMRuntimeVersionSet(v)
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM runtime VM interpreter version", value: v.VMInterpreterVersion},
		{name: "AVM runtime scheduler version", value: v.SchedulerVersion},
		{name: "AVM runtime gas policy version", value: v.GasPolicyVersion},
		{name: "AVM runtime zone config version", value: v.ZoneConfigVersion},
		{name: "AVM runtime backend adapter version", value: v.BackendAdapterVersion},
		{name: "AVM runtime interface schema version", value: v.InterfaceSchemaVersion},
		{name: "AVM runtime retry policy version", value: v.RetryPolicyVersion},
		{name: "AVM runtime queue limit version", value: v.QueueLimitVersion},
	} {
		if err := validateEngineToken(item.name, item.value, MaxAVMUpgradeVersionLength); err != nil {
			return err
		}
	}
	if v.RuntimeVersionSetHash == "" {
		return errors.New("AVM runtime version set hash is required")
	}
	if err := zonestypes.ValidateHash("AVM runtime version set hash", v.RuntimeVersionSetHash); err != nil {
		return err
	}
	if v.RuntimeVersionSetHash != ComputeAVMRuntimeVersionSetHash(v) {
		return errors.New("AVM runtime version set hash mismatch")
	}
	return nil
}

func NewAVMScheduledUpgradeState(state AVMScheduledUpgradeState) (AVMScheduledUpgradeState, error) {
	state = canonicalAVMScheduledUpgradeState(state)
	state.StateHash = ComputeAVMScheduledUpgradeStateHash(state)
	return state, state.Validate()
}

func (s AVMScheduledUpgradeState) Validate() error {
	s = canonicalAVMScheduledUpgradeState(s)
	if err := validateEngineToken("AVM scheduled upgrade id", s.UpgradeID, MaxAVMUpgradeTokenLength); err != nil {
		return err
	}
	if !IsAVMUpgradeComponentKind(s.Component) {
		return fmt.Errorf("invalid AVM scheduled upgrade component %q", s.Component)
	}
	if err := validateRouterOptionalToken("AVM scheduled upgrade from version", s.FromVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM scheduled upgrade to version", s.ToVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if s.FromVersion == s.ToVersion {
		return errors.New("AVM scheduled upgrade must change version")
	}
	if s.ActivationHeight == 0 {
		return errors.New("AVM scheduled upgrade activation height must be positive")
	}
	if !IsAVMUpgradeCompatibilityMode(s.CompatibilityMode) {
		return fmt.Errorf("invalid AVM upgrade compatibility mode %q", s.CompatibilityMode)
	}
	if !IsAVMUpgradeStatus(s.Status) {
		return fmt.Errorf("invalid AVM upgrade status %q", s.Status)
	}
	if s.Status == AVMUpgradeStatusActive && s.ActivationHeight == 0 {
		return errors.New("AVM active upgrade requires activation height")
	}
	if s.MigrationRequired && s.CompatibilityMode == AVMUpgradeCompatibilityNone {
		return errors.New("AVM migration-required upgrade must declare compatibility mode")
	}
	if s.StateHash == "" {
		return errors.New("AVM scheduled upgrade state hash is required")
	}
	if err := zonestypes.ValidateHash("AVM scheduled upgrade state hash", s.StateHash); err != nil {
		return err
	}
	if s.StateHash != ComputeAVMScheduledUpgradeStateHash(s) {
		return errors.New("AVM scheduled upgrade state hash mismatch")
	}
	return nil
}

func NewAVMUpgradeMigrationHandler(handler AVMUpgradeMigrationHandler) (AVMUpgradeMigrationHandler, error) {
	handler = canonicalAVMUpgradeMigrationHandler(handler)
	handler.HandlerHash = ComputeAVMUpgradeMigrationHandlerHash(handler)
	return handler, handler.Validate()
}

func (h AVMUpgradeMigrationHandler) Validate() error {
	h = canonicalAVMUpgradeMigrationHandler(h)
	if err := validateEngineToken("AVM migration upgrade id", h.UpgradeID, MaxAVMUpgradeTokenLength); err != nil {
		return err
	}
	if !IsAVMUpgradeMigrationKind(h.Kind) {
		return fmt.Errorf("invalid AVM upgrade migration kind %q", h.Kind)
	}
	if err := validateRouterOptionalToken("AVM migration from version", h.FromVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM migration to version", h.ToVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if h.FromVersion == h.ToVersion {
		return errors.New("AVM migration handler must change version")
	}
	if err := validateEngineToken("AVM migration handler id", h.HandlerID, MaxAVMUpgradeTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM migration source root", h.SourceRoot); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM migration target root", h.TargetRoot); err != nil {
		return err
	}
	if h.MigratedCount == 0 {
		return errors.New("AVM migration handler must migrate at least one item")
	}
	if !h.Bounded {
		return errors.New("AVM migration handler must be bounded")
	}
	if h.HandlerHash == "" {
		return errors.New("AVM migration handler hash is required")
	}
	if err := zonestypes.ValidateHash("AVM migration handler hash", h.HandlerHash); err != nil {
		return err
	}
	if h.HandlerHash != ComputeAVMUpgradeMigrationHandlerHash(h) {
		return errors.New("AVM migration handler hash mismatch")
	}
	return nil
}

func NewAVMVersionedGasTable(tables []AVMGasTableActivation) (AVMVersionedGasTable, error) {
	table := AVMVersionedGasTable{Tables: append([]AVMGasTableActivation(nil), tables...)}
	table = canonicalAVMVersionedGasTable(table)
	table.TableRoot = ComputeAVMVersionedGasTableRoot(table)
	return table, table.Validate()
}

func (t AVMVersionedGasTable) Validate() error {
	t = canonicalAVMVersionedGasTable(t)
	if len(t.Tables) == 0 {
		return errors.New("AVM versioned gas table requires at least one activation")
	}
	if err := validateAVMGasTableActivations(t.Tables, 1); err != nil {
		return err
	}
	if t.TableRoot == "" {
		return errors.New("AVM versioned gas table root is required")
	}
	if err := zonestypes.ValidateHash("AVM versioned gas table root", t.TableRoot); err != nil {
		return err
	}
	if t.TableRoot != ComputeAVMVersionedGasTableRoot(t) {
		return errors.New("AVM versioned gas table root mismatch")
	}
	return nil
}

func (t AVMVersionedGasTable) ActiveAt(height uint64) (AVMGasTableActivation, error) {
	t = canonicalAVMVersionedGasTable(t)
	if err := t.Validate(); err != nil {
		return AVMGasTableActivation{}, err
	}
	if height == 0 {
		return AVMGasTableActivation{}, errors.New("AVM gas table selection height must be positive")
	}
	var selected AVMGasTableActivation
	for _, table := range t.Tables {
		if table.ActivationHeight <= height {
			selected = table
		}
	}
	if selected.ActivationHeight == 0 {
		return AVMGasTableActivation{}, errors.New("no AVM gas table active at height")
	}
	return selected, nil
}

func NewAVMUpgradeStateRegistry(registry AVMUpgradeStateRegistry) (AVMUpgradeStateRegistry, error) {
	registry = canonicalAVMUpgradeStateRegistry(registry)
	registry.RegistryHash = ComputeAVMUpgradeStateRegistryHash(registry)
	return registry, registry.Validate()
}

func (r AVMUpgradeStateRegistry) Validate() error {
	r = canonicalAVMUpgradeStateRegistry(r)
	if err := r.RuntimeVersions.Validate(); err != nil {
		return err
	}
	if err := validateAVMScheduledUpgradeStates(r.States); err != nil {
		return err
	}
	if err := validateAVMUpgradeMigrationHandlers(r.MigrationHandlers); err != nil {
		return err
	}
	if err := r.GasTable.Validate(); err != nil {
		return err
	}
	if err := ValidateAVMUpgradeRollbackPrevention(r.States); err != nil {
		return err
	}
	if r.RegistryHash == "" {
		return errors.New("AVM upgrade state registry hash is required")
	}
	if err := zonestypes.ValidateHash("AVM upgrade state registry hash", r.RegistryHash); err != nil {
		return err
	}
	if r.RegistryHash != ComputeAVMUpgradeStateRegistryHash(r) {
		return errors.New("AVM upgrade state registry hash mismatch")
	}
	return nil
}

func BuildAVMQueueUpgradeMigration(upgradeID, fromVersion, toVersion string, queue AVMZoneQueue) (AVMUpgradeMigrationHandler, error) {
	queue = canonicalAVMZoneQueue(queue)
	queue.QueueRoot = ComputeAVMZoneQueueRoot(queue)
	if err := queue.Validate(); err != nil {
		return AVMUpgradeMigrationHandler{}, err
	}
	count := len(queue.PriorityQueue) + len(queue.DelayedQueue) + len(queue.RetryQueue) + len(queue.FailedQueue)
	return NewAVMUpgradeMigrationHandler(AVMUpgradeMigrationHandler{
		UpgradeID:	upgradeID,
		Kind:		AVMUpgradeMigrationQueue,
		FromVersion:	fromVersion,
		ToVersion:	toVersion,
		HandlerID:	"queue-migration",
		SourceRoot:	queue.QueueRoot,
		TargetRoot:	ComputeAVMUpgradeMigratedRoot(queue.QueueRoot, toVersion, uint32(count)),
		MigratedCount:	uint32(count),
		Bounded:	true,
	})
}

func BuildAVMContinuationUpgradeMigration(upgradeID, fromVersion, toVersion string, continuations []ContinuationRecord) (AVMUpgradeMigrationHandler, error) {
	if len(continuations) == 0 {
		return AVMUpgradeMigrationHandler{}, errors.New("AVM continuation migration requires records")
	}
	sourceRoot := ComputeAVMUpgradeContinuationSetRoot(continuations)
	return NewAVMUpgradeMigrationHandler(AVMUpgradeMigrationHandler{
		UpgradeID:	upgradeID,
		Kind:		AVMUpgradeMigrationContinuation,
		FromVersion:	fromVersion,
		ToVersion:	toVersion,
		HandlerID:	"continuation-migration",
		SourceRoot:	sourceRoot,
		TargetRoot:	ComputeAVMUpgradeMigratedRoot(sourceRoot, toVersion, uint32(len(continuations))),
		MigratedCount:	uint32(len(continuations)),
		Bounded:	true,
	})
}

func BuildAVMPendingMessageCompatibilityPolicies(messages []AVMAsyncMessage, activationHeight, executionHeight uint64, preRuntime, postRuntime, preScheduler, postScheduler, preGas, postGas string) ([]AVMVersionedMessageExecutionPolicy, error) {
	policies := make([]AVMVersionedMessageExecutionPolicy, 0, len(messages))
	for _, msg := range messages {
		policy, err := NewAVMVersionedMessageExecutionPolicy(msg, activationHeight, executionHeight, preRuntime, postRuntime, preScheduler, postScheduler, preGas, postGas)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	sort.SliceStable(policies, func(i, j int) bool { return policies[i].MessageID < policies[j].MessageID })
	return policies, validateAVMVersionedMessagePolicies(policies)
}

func ValidateAVMUpgradeRollbackPrevention(states []AVMScheduledUpgradeState) error {
	latestActivated := make(map[AVMUpgradeComponentKind]AVMScheduledUpgradeState)
	for _, state := range states {
		state = canonicalAVMScheduledUpgradeState(state)
		if state.Status != AVMUpgradeStatusActive && state.Status != AVMUpgradeStatusCompleted {
			continue
		}
		previous, found := latestActivated[state.Component]
		if found {
			if state.ActivationHeight <= previous.ActivationHeight {
				return errors.New("AVM activated upgrades cannot roll back activation height")
			}
			if state.ToVersion == previous.FromVersion {
				return errors.New("AVM activated upgrades cannot roll back to previous version")
			}
		}
		latestActivated[state.Component] = state
	}
	return nil
}

func IsAVMUpgradeCompatibilityMode(mode AVMUpgradeCompatibilityMode) bool {
	switch mode {
	case AVMUpgradeCompatibilityNone, AVMUpgradeCompatibilityVersionedPolicy, AVMUpgradeCompatibilityLegacyRuntime:
		return true
	default:
		return false
	}
}

func IsAVMUpgradeStatus(status AVMUpgradeStatus) bool {
	switch status {
	case AVMUpgradeStatusScheduled, AVMUpgradeStatusActive, AVMUpgradeStatusCompleted, AVMUpgradeStatusFailed:
		return true
	default:
		return false
	}
}

func IsAVMUpgradeMigrationKind(kind AVMUpgradeMigrationKind) bool {
	switch kind {
	case AVMUpgradeMigrationQueue, AVMUpgradeMigrationContinuation:
		return true
	default:
		return false
	}
}

func ComputeAVMRuntimeVersionSetHash(versions AVMRuntimeVersionSet) string {
	versions = canonicalAVMRuntimeVersionSet(versions)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-runtime-version-set-v1")
	writeEnginePart(h, versions.VMInterpreterVersion)
	writeEnginePart(h, versions.SchedulerVersion)
	writeEnginePart(h, versions.GasPolicyVersion)
	writeEnginePart(h, versions.ZoneConfigVersion)
	writeEnginePart(h, versions.BackendAdapterVersion)
	writeEnginePart(h, versions.InterfaceSchemaVersion)
	writeEnginePart(h, versions.RetryPolicyVersion)
	writeEnginePart(h, versions.QueueLimitVersion)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMScheduledUpgradeStateHash(state AVMScheduledUpgradeState) string {
	state = canonicalAVMScheduledUpgradeState(state)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-scheduled-upgrade-state-v1")
	writeEnginePart(h, state.UpgradeID)
	writeEnginePart(h, string(state.Component))
	writeEnginePart(h, state.FromVersion)
	writeEnginePart(h, state.ToVersion)
	writeEngineUint64(h, state.ActivationHeight)
	writeEngineBool(h, state.MigrationRequired)
	writeEnginePart(h, string(state.CompatibilityMode))
	writeEnginePart(h, string(state.Status))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMUpgradeMigrationHandlerHash(handler AVMUpgradeMigrationHandler) string {
	handler = canonicalAVMUpgradeMigrationHandler(handler)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-upgrade-migration-handler-v1")
	writeEnginePart(h, handler.UpgradeID)
	writeEnginePart(h, string(handler.Kind))
	writeEnginePart(h, handler.FromVersion)
	writeEnginePart(h, handler.ToVersion)
	writeEnginePart(h, handler.HandlerID)
	writeEnginePart(h, handler.SourceRoot)
	writeEnginePart(h, handler.TargetRoot)
	writeEngineUint64(h, uint64(handler.MigratedCount))
	writeEngineBool(h, handler.Bounded)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMVersionedGasTableRoot(table AVMVersionedGasTable) string {
	table = canonicalAVMVersionedGasTable(table)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-versioned-gas-table-v1")
	writeEngineUint64(h, uint64(len(table.Tables)))
	for _, activation := range table.Tables {
		writeEnginePart(h, activation.TableHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMUpgradeStateRegistryHash(registry AVMUpgradeStateRegistry) string {
	registry = canonicalAVMUpgradeStateRegistry(registry)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-upgrade-state-registry-v1")
	writeEnginePart(h, registry.RuntimeVersions.RuntimeVersionSetHash)
	writeEngineUint64(h, uint64(len(registry.States)))
	for _, state := range registry.States {
		writeEnginePart(h, state.StateHash)
	}
	writeEngineUint64(h, uint64(len(registry.MigrationHandlers)))
	for _, handler := range registry.MigrationHandlers {
		writeEnginePart(h, handler.HandlerHash)
	}
	writeEnginePart(h, registry.GasTable.TableRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMUpgradeMigratedRoot(sourceRoot, toVersion string, migratedCount uint32) string {
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-upgrade-migrated-root-v1")
	writeEnginePart(h, strings.TrimSpace(sourceRoot))
	writeEnginePart(h, strings.TrimSpace(toVersion))
	writeEngineUint64(h, uint64(migratedCount))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMUpgradeContinuationSetRoot(continuations []ContinuationRecord) string {
	records := append([]ContinuationRecord(nil), continuations...)
	sort.SliceStable(records, func(i, j int) bool { return compareContinuationRecords(records[i], records[j]) < 0 })
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-upgrade-continuation-set-v1")
	writeEngineUint64(h, uint64(len(records)))
	for _, continuation := range records {
		writeContinuationRecord(h, continuation)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMRuntimeVersionSet(versions AVMRuntimeVersionSet) AVMRuntimeVersionSet {
	versions.VMInterpreterVersion = strings.TrimSpace(versions.VMInterpreterVersion)
	versions.SchedulerVersion = strings.TrimSpace(versions.SchedulerVersion)
	versions.GasPolicyVersion = strings.TrimSpace(versions.GasPolicyVersion)
	versions.ZoneConfigVersion = strings.TrimSpace(versions.ZoneConfigVersion)
	versions.BackendAdapterVersion = strings.TrimSpace(versions.BackendAdapterVersion)
	versions.InterfaceSchemaVersion = strings.TrimSpace(versions.InterfaceSchemaVersion)
	versions.RetryPolicyVersion = strings.TrimSpace(versions.RetryPolicyVersion)
	versions.QueueLimitVersion = strings.TrimSpace(versions.QueueLimitVersion)
	versions.RuntimeVersionSetHash = strings.TrimSpace(versions.RuntimeVersionSetHash)
	return versions
}

func canonicalAVMScheduledUpgradeState(state AVMScheduledUpgradeState) AVMScheduledUpgradeState {
	state.UpgradeID = strings.TrimSpace(state.UpgradeID)
	state.FromVersion = strings.TrimSpace(state.FromVersion)
	state.ToVersion = strings.TrimSpace(state.ToVersion)
	state.StateHash = strings.TrimSpace(state.StateHash)
	return state
}

func canonicalAVMUpgradeMigrationHandler(handler AVMUpgradeMigrationHandler) AVMUpgradeMigrationHandler {
	handler.UpgradeID = strings.TrimSpace(handler.UpgradeID)
	handler.FromVersion = strings.TrimSpace(handler.FromVersion)
	handler.ToVersion = strings.TrimSpace(handler.ToVersion)
	handler.HandlerID = strings.TrimSpace(handler.HandlerID)
	handler.SourceRoot = strings.TrimSpace(handler.SourceRoot)
	handler.TargetRoot = strings.TrimSpace(handler.TargetRoot)
	handler.HandlerHash = strings.TrimSpace(handler.HandlerHash)
	return handler
}

func canonicalAVMVersionedGasTable(table AVMVersionedGasTable) AVMVersionedGasTable {
	table.Tables = append([]AVMGasTableActivation(nil), table.Tables...)
	for i := range table.Tables {
		table.Tables[i] = canonicalAVMGasTableActivation(table.Tables[i])
	}
	sort.SliceStable(table.Tables, func(i, j int) bool {
		if table.Tables[i].ActivationHeight != table.Tables[j].ActivationHeight {
			return table.Tables[i].ActivationHeight < table.Tables[j].ActivationHeight
		}
		return table.Tables[i].PolicyVersion < table.Tables[j].PolicyVersion
	})
	table.TableRoot = strings.TrimSpace(table.TableRoot)
	return table
}

func canonicalAVMUpgradeStateRegistry(registry AVMUpgradeStateRegistry) AVMUpgradeStateRegistry {
	registry.RuntimeVersions = canonicalAVMRuntimeVersionSet(registry.RuntimeVersions)
	registry.States = append([]AVMScheduledUpgradeState(nil), registry.States...)
	for i := range registry.States {
		registry.States[i] = canonicalAVMScheduledUpgradeState(registry.States[i])
	}
	sort.SliceStable(registry.States, func(i, j int) bool {
		if registry.States[i].Component != registry.States[j].Component {
			return registry.States[i].Component < registry.States[j].Component
		}
		if registry.States[i].ActivationHeight != registry.States[j].ActivationHeight {
			return registry.States[i].ActivationHeight < registry.States[j].ActivationHeight
		}
		return registry.States[i].UpgradeID < registry.States[j].UpgradeID
	})
	registry.MigrationHandlers = append([]AVMUpgradeMigrationHandler(nil), registry.MigrationHandlers...)
	for i := range registry.MigrationHandlers {
		registry.MigrationHandlers[i] = canonicalAVMUpgradeMigrationHandler(registry.MigrationHandlers[i])
	}
	sort.SliceStable(registry.MigrationHandlers, func(i, j int) bool {
		if registry.MigrationHandlers[i].Kind != registry.MigrationHandlers[j].Kind {
			return registry.MigrationHandlers[i].Kind < registry.MigrationHandlers[j].Kind
		}
		return registry.MigrationHandlers[i].HandlerID < registry.MigrationHandlers[j].HandlerID
	})
	registry.GasTable = canonicalAVMVersionedGasTable(registry.GasTable)
	registry.RegistryHash = strings.TrimSpace(registry.RegistryHash)
	return registry
}

func validateAVMScheduledUpgradeStates(states []AVMScheduledUpgradeState) error {
	if len(states) == 0 || len(states) > MaxAVMUpgradeStates {
		return fmt.Errorf("AVM scheduled upgrade states must be 1..%d", MaxAVMUpgradeStates)
	}
	seen := make(map[string]struct{}, len(states))
	for i, state := range states {
		if err := state.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s/%020d", state.Component, state.UpgradeID, state.ActivationHeight)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate AVM scheduled upgrade state %q", key)
		}
		seen[key] = struct{}{}
		if i > 0 {
			prev := states[i-1]
			prevKey := fmt.Sprintf("%s/%020d/%s", prev.Component, prev.ActivationHeight, prev.UpgradeID)
			thisKey := fmt.Sprintf("%s/%020d/%s", state.Component, state.ActivationHeight, state.UpgradeID)
			if prevKey >= thisKey {
				return errors.New("AVM scheduled upgrade states must be sorted canonically")
			}
		}
	}
	return nil
}

func validateAVMUpgradeMigrationHandlers(handlers []AVMUpgradeMigrationHandler) error {
	if len(handlers) > MaxAVMUpgradeMigrationHandlers {
		return fmt.Errorf("AVM migration handlers must be <= %d", MaxAVMUpgradeMigrationHandlers)
	}
	seen := make(map[string]struct{}, len(handlers))
	for i, handler := range handlers {
		if err := handler.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s", handler.Kind, handler.HandlerID)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate AVM migration handler %q", key)
		}
		seen[key] = struct{}{}
		if i > 0 {
			prevKey := fmt.Sprintf("%s/%s", handlers[i-1].Kind, handlers[i-1].HandlerID)
			if prevKey >= key {
				return errors.New("AVM migration handlers must be sorted canonically")
			}
		}
	}
	return nil
}

func validateAVMVersionedMessagePolicies(policies []AVMVersionedMessageExecutionPolicy) error {
	seen := make(map[string]struct{}, len(policies))
	for i, policy := range policies {
		if err := policy.Validate(); err != nil {
			return err
		}
		if _, found := seen[policy.MessageID]; found {
			return fmt.Errorf("duplicate AVM versioned message policy %q", policy.MessageID)
		}
		seen[policy.MessageID] = struct{}{}
		if i > 0 && policies[i-1].MessageID >= policy.MessageID {
			return errors.New("AVM versioned message policies must be sorted canonically")
		}
	}
	return nil
}
