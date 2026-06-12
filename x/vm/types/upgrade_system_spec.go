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
	AVMUpgradeComponentVMInterpreter		AVMUpgradeComponentKind	= "vm_interpreter_version"
	AVMUpgradeComponentSchedulerRules		AVMUpgradeComponentKind	= "scheduler_rules"
	AVMUpgradeComponentGasModel			AVMUpgradeComponentKind	= "gas_model"
	AVMUpgradeComponentZoneConfiguration		AVMUpgradeComponentKind	= "zone_configuration"
	AVMUpgradeComponentBackendAdapters		AVMUpgradeComponentKind	= "backend_adapters"
	AVMUpgradeComponentInterfaceSchemaVersion	AVMUpgradeComponentKind	= "interface_schema_version"
	AVMUpgradeComponentRetryPolicies		AVMUpgradeComponentKind	= "retry_policies"
	AVMUpgradeComponentQueueLimits			AVMUpgradeComponentKind	= "queue_limits"

	MaxAVMUpgradeComponents		= 32
	MaxAVMUpgradeSchedulerRules	= 32
	MaxAVMUpgradeVersionedItems	= 4096
	MaxAVMUpgradeTokenLength	= 128
	MaxAVMUpgradeVersionLength	= 64
	MaxAVMUpgradeProposalIDBytes	= 128
)

type AVMUpgradeComponentKind string

type AVMUpgradeComponent struct {
	Kind			AVMUpgradeComponentKind
	PreviousVersion		string
	NextVersion		string
	ActivationHeight	uint64
	ComponentHash		string
}

type AVMSchedulerRuleVersion struct {
	RuleSetID		string
	Version			string
	EffectiveFromHeight	uint64
	EffectiveUntilHeight	uint64
	RuleHash		string
}

type AVMGasTableActivation struct {
	ActivationHeight	uint64
	PolicyVersion		string
	Policy			AVMGasPolicy
	Schedule		AVMGasSchedule
	TableHash		string
}

type AVMContinuationRuntimeVersion struct {
	ContinuationID	string
	ActorID		string
	RuntimeVersion	string
	StoredHeight	uint64
	VersionHash	string
}

type AVMContractCodeVMVersion struct {
	CodeID		uint64
	BackendKind	AVMContractBackendKind
	CodeHash	string
	VMVersion	string
	VersionHash	string
}

type AVMVersionedMessageExecutionPolicy struct {
	MessageID		string
	CreatedHeight		uint64
	ExecutionHeight		uint64
	ActivationHeight	uint64
	RuntimeVersion		string
	SchedulerVersion	string
	GasPolicyVersion	string
	PolicyHash		string
}

type AVMUpgradeManifest struct {
	UpgradeID		string
	GovernanceProposalID	string
	StagedHeight		uint64
	Components		[]AVMUpgradeComponent
	SchedulerRules		[]AVMSchedulerRuleVersion
	GasTables		[]AVMGasTableActivation
	Continuations		[]AVMContinuationRuntimeVersion
	ContractCodes		[]AVMContractCodeVMVersion
	ManifestHash		string
}

func NewAVMUpgradeComponent(component AVMUpgradeComponent) (AVMUpgradeComponent, error) {
	component = canonicalAVMUpgradeComponent(component)
	component.ComponentHash = ComputeAVMUpgradeComponentHash(component)
	return component, component.Validate()
}

func (c AVMUpgradeComponent) Validate() error {
	c = canonicalAVMUpgradeComponent(c)
	if !IsAVMUpgradeComponentKind(c.Kind) {
		return fmt.Errorf("invalid AVM upgrade component kind %q", c.Kind)
	}
	if err := validateRouterOptionalToken("AVM upgrade previous version", c.PreviousVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM upgrade next version", c.NextVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if c.PreviousVersion == c.NextVersion {
		return errors.New("AVM upgrade component must change version")
	}
	if c.ActivationHeight == 0 {
		return errors.New("AVM upgrade component activation height must be staged")
	}
	if c.ComponentHash == "" {
		return errors.New("AVM upgrade component hash is required")
	}
	if err := zonestypes.ValidateHash("AVM upgrade component hash", c.ComponentHash); err != nil {
		return err
	}
	if c.ComponentHash != ComputeAVMUpgradeComponentHash(c) {
		return errors.New("AVM upgrade component hash mismatch")
	}
	return nil
}

func NewAVMSchedulerRuleVersion(rule AVMSchedulerRuleVersion) (AVMSchedulerRuleVersion, error) {
	rule = canonicalAVMSchedulerRuleVersion(rule)
	rule.RuleHash = ComputeAVMSchedulerRuleVersionHash(rule)
	return rule, rule.Validate()
}

func (r AVMSchedulerRuleVersion) Validate() error {
	r = canonicalAVMSchedulerRuleVersion(r)
	if err := validateEngineToken("AVM scheduler rule set id", r.RuleSetID, MaxAVMUpgradeTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM scheduler rule version", r.Version, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if r.EffectiveFromHeight == 0 {
		return errors.New("AVM scheduler rule effective height must be positive")
	}
	if r.EffectiveUntilHeight != 0 && r.EffectiveUntilHeight < r.EffectiveFromHeight {
		return errors.New("AVM scheduler rule effective range is inverted")
	}
	if r.RuleHash == "" {
		return errors.New("AVM scheduler rule hash is required")
	}
	if err := zonestypes.ValidateHash("AVM scheduler rule hash", r.RuleHash); err != nil {
		return err
	}
	if r.RuleHash != ComputeAVMSchedulerRuleVersionHash(r) {
		return errors.New("AVM scheduler rule hash mismatch")
	}
	return nil
}

func NewAVMGasTableActivation(table AVMGasTableActivation) (AVMGasTableActivation, error) {
	table = canonicalAVMGasTableActivation(table)
	if table.Policy.PolicyHash == "" {
		table.Policy.PolicyHash = ComputeAVMGasPolicyHash(table.Policy)
	}
	if table.Schedule.ScheduleHash == "" {
		table.Schedule.ScheduleHash = ComputeAVMGasScheduleHash(table.Schedule)
	}
	table.TableHash = ComputeAVMGasTableActivationHash(table)
	return table, table.Validate()
}

func (t AVMGasTableActivation) Validate() error {
	t = canonicalAVMGasTableActivation(t)
	if t.ActivationHeight == 0 {
		return errors.New("AVM gas table activation height must be positive")
	}
	if err := validateEngineToken("AVM gas table policy version", t.PolicyVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if err := t.Policy.Validate(); err != nil {
		return err
	}
	if err := t.Schedule.Validate(); err != nil {
		return err
	}
	if t.TableHash == "" {
		return errors.New("AVM gas table activation hash is required")
	}
	if err := zonestypes.ValidateHash("AVM gas table activation hash", t.TableHash); err != nil {
		return err
	}
	if t.TableHash != ComputeAVMGasTableActivationHash(t) {
		return errors.New("AVM gas table activation hash mismatch")
	}
	return nil
}

func NewAVMContinuationRuntimeVersion(version AVMContinuationRuntimeVersion) (AVMContinuationRuntimeVersion, error) {
	version = canonicalAVMContinuationRuntimeVersion(version)
	version.VersionHash = ComputeAVMContinuationRuntimeVersionHash(version)
	return version, version.Validate()
}

func (v AVMContinuationRuntimeVersion) Validate() error {
	v = canonicalAVMContinuationRuntimeVersion(v)
	if err := validateEngineToken("AVM continuation id", v.ContinuationID, MaxContinuationTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM continuation actor id", v.ActorID, MaxActorRuntimeTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM continuation runtime version", v.RuntimeVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if v.StoredHeight == 0 {
		return errors.New("AVM continuation runtime version stored height must be positive")
	}
	if v.VersionHash == "" {
		return errors.New("AVM continuation runtime version hash is required")
	}
	if err := zonestypes.ValidateHash("AVM continuation runtime version hash", v.VersionHash); err != nil {
		return err
	}
	if v.VersionHash != ComputeAVMContinuationRuntimeVersionHash(v) {
		return errors.New("AVM continuation runtime version hash mismatch")
	}
	return nil
}

func NewAVMContractCodeVMVersion(version AVMContractCodeVMVersion) (AVMContractCodeVMVersion, error) {
	version = canonicalAVMContractCodeVMVersion(version)
	version.VersionHash = ComputeAVMContractCodeVMVersionHash(version)
	return version, version.Validate()
}

func (v AVMContractCodeVMVersion) Validate() error {
	v = canonicalAVMContractCodeVMVersion(v)
	if v.CodeID == 0 {
		return errors.New("AVM contract code VM version requires code id")
	}
	if !IsAVMContractBackendKind(v.BackendKind) {
		return fmt.Errorf("invalid AVM contract code backend kind %q", v.BackendKind)
	}
	if err := zonestypes.ValidateHash("AVM contract code hash", v.CodeHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM contract code VM version", v.VMVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if v.VersionHash == "" {
		return errors.New("AVM contract code VM version hash is required")
	}
	if err := zonestypes.ValidateHash("AVM contract code VM version hash", v.VersionHash); err != nil {
		return err
	}
	if v.VersionHash != ComputeAVMContractCodeVMVersionHash(v) {
		return errors.New("AVM contract code VM version hash mismatch")
	}
	return nil
}

func NewAVMVersionedMessageExecutionPolicy(msg AVMAsyncMessage, activationHeight, executionHeight uint64, preRuntime, postRuntime, preScheduler, postScheduler, preGas, postGas string) (AVMVersionedMessageExecutionPolicy, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMVersionedMessageExecutionPolicy{}, err
	}
	if activationHeight == 0 || executionHeight == 0 {
		return AVMVersionedMessageExecutionPolicy{}, errors.New("AVM versioned message policy requires activation and execution height")
	}
	policy := AVMVersionedMessageExecutionPolicy{
		MessageID:		msg.ID,
		CreatedHeight:		msg.CreatedHeight,
		ExecutionHeight:	executionHeight,
		ActivationHeight:	activationHeight,
		RuntimeVersion:		strings.TrimSpace(postRuntime),
		SchedulerVersion:	strings.TrimSpace(postScheduler),
		GasPolicyVersion:	strings.TrimSpace(postGas),
	}
	if msg.CreatedHeight < activationHeight {
		policy.RuntimeVersion = strings.TrimSpace(preRuntime)
		policy.SchedulerVersion = strings.TrimSpace(preScheduler)
		policy.GasPolicyVersion = strings.TrimSpace(preGas)
	}
	policy.PolicyHash = ComputeAVMVersionedMessageExecutionPolicyHash(policy)
	return policy, policy.Validate()
}

func (p AVMVersionedMessageExecutionPolicy) Validate() error {
	p = canonicalAVMVersionedMessageExecutionPolicy(p)
	if err := zonestypes.ValidateHash("AVM versioned message id", p.MessageID); err != nil {
		return err
	}
	if p.CreatedHeight == 0 || p.ExecutionHeight == 0 || p.ActivationHeight == 0 {
		return errors.New("AVM versioned message policy heights must be positive")
	}
	if err := validateEngineToken("AVM versioned message runtime version", p.RuntimeVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM versioned message scheduler version", p.SchedulerVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM versioned message gas policy version", p.GasPolicyVersion, MaxAVMUpgradeVersionLength); err != nil {
		return err
	}
	if p.PolicyHash == "" {
		return errors.New("AVM versioned message policy hash is required")
	}
	if err := zonestypes.ValidateHash("AVM versioned message policy hash", p.PolicyHash); err != nil {
		return err
	}
	if p.PolicyHash != ComputeAVMVersionedMessageExecutionPolicyHash(p) {
		return errors.New("AVM versioned message policy hash mismatch")
	}
	return nil
}

func NewAVMUpgradeManifest(manifest AVMUpgradeManifest) (AVMUpgradeManifest, error) {
	manifest = canonicalAVMUpgradeManifest(manifest)
	manifest.ManifestHash = ComputeAVMUpgradeManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (m AVMUpgradeManifest) Validate() error {
	m = canonicalAVMUpgradeManifest(m)
	if err := validateEngineToken("AVM upgrade id", m.UpgradeID, MaxAVMUpgradeTokenLength); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM governance proposal id", m.GovernanceProposalID, MaxAVMUpgradeProposalIDBytes); err != nil {
		return err
	}
	if m.GovernanceProposalID == "" {
		return errors.New("AVM upgrade manifest requires governance proposal id")
	}
	if m.StagedHeight == 0 {
		return errors.New("AVM upgrade manifest staged height must be positive")
	}
	if len(m.Components) == 0 || len(m.Components) > MaxAVMUpgradeComponents {
		return fmt.Errorf("AVM upgrade components must be 1..%d", MaxAVMUpgradeComponents)
	}
	if err := validateAVMUpgradeComponents(m.Components, m.StagedHeight); err != nil {
		return err
	}
	if err := validateAVMSchedulerRuleVersions(m.SchedulerRules, m.StagedHeight); err != nil {
		return err
	}
	if err := validateAVMGasTableActivations(m.GasTables, m.StagedHeight); err != nil {
		return err
	}
	if err := validateAVMContinuationRuntimeVersions(m.Continuations); err != nil {
		return err
	}
	if err := validateAVMContractCodeVMVersions(m.ContractCodes); err != nil {
		return err
	}
	if m.ManifestHash == "" {
		return errors.New("AVM upgrade manifest hash is required")
	}
	if err := zonestypes.ValidateHash("AVM upgrade manifest hash", m.ManifestHash); err != nil {
		return err
	}
	if m.ManifestHash != ComputeAVMUpgradeManifestHash(m) {
		return errors.New("AVM upgrade manifest hash mismatch")
	}
	return nil
}

func AllAVMUpgradeComponentKinds() []AVMUpgradeComponentKind {
	return []AVMUpgradeComponentKind{
		AVMUpgradeComponentBackendAdapters,
		AVMUpgradeComponentGasModel,
		AVMUpgradeComponentInterfaceSchemaVersion,
		AVMUpgradeComponentQueueLimits,
		AVMUpgradeComponentRetryPolicies,
		AVMUpgradeComponentSchedulerRules,
		AVMUpgradeComponentVMInterpreter,
		AVMUpgradeComponentZoneConfiguration,
	}
}

func IsAVMUpgradeComponentKind(kind AVMUpgradeComponentKind) bool {
	switch kind {
	case AVMUpgradeComponentVMInterpreter,
		AVMUpgradeComponentSchedulerRules,
		AVMUpgradeComponentGasModel,
		AVMUpgradeComponentZoneConfiguration,
		AVMUpgradeComponentBackendAdapters,
		AVMUpgradeComponentInterfaceSchemaVersion,
		AVMUpgradeComponentRetryPolicies,
		AVMUpgradeComponentQueueLimits:
		return true
	default:
		return false
	}
}

func ComputeAVMUpgradeComponentHash(component AVMUpgradeComponent) string {
	component = canonicalAVMUpgradeComponent(component)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-upgrade-component-v1")
	writeEnginePart(h, string(component.Kind))
	writeEnginePart(h, component.PreviousVersion)
	writeEnginePart(h, component.NextVersion)
	writeEngineUint64(h, component.ActivationHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMSchedulerRuleVersionHash(rule AVMSchedulerRuleVersion) string {
	rule = canonicalAVMSchedulerRuleVersion(rule)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-scheduler-rule-version-v1")
	writeEnginePart(h, rule.RuleSetID)
	writeEnginePart(h, rule.Version)
	writeEngineUint64(h, rule.EffectiveFromHeight)
	writeEngineUint64(h, rule.EffectiveUntilHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMGasTableActivationHash(table AVMGasTableActivation) string {
	table = canonicalAVMGasTableActivation(table)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-gas-table-activation-v1")
	writeEngineUint64(h, table.ActivationHeight)
	writeEnginePart(h, table.PolicyVersion)
	writeEnginePart(h, table.Policy.PolicyHash)
	writeEnginePart(h, table.Schedule.ScheduleHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContinuationRuntimeVersionHash(version AVMContinuationRuntimeVersion) string {
	version = canonicalAVMContinuationRuntimeVersion(version)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-continuation-runtime-version-v1")
	writeEnginePart(h, version.ContinuationID)
	writeEnginePart(h, version.ActorID)
	writeEnginePart(h, version.RuntimeVersion)
	writeEngineUint64(h, version.StoredHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMContractCodeVMVersionHash(version AVMContractCodeVMVersion) string {
	version = canonicalAVMContractCodeVMVersion(version)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-contract-code-vm-version-v1")
	writeEngineUint64(h, version.CodeID)
	writeEnginePart(h, string(version.BackendKind))
	writeEnginePart(h, version.CodeHash)
	writeEnginePart(h, version.VMVersion)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMVersionedMessageExecutionPolicyHash(policy AVMVersionedMessageExecutionPolicy) string {
	policy = canonicalAVMVersionedMessageExecutionPolicy(policy)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-versioned-message-policy-v1")
	writeEnginePart(h, policy.MessageID)
	writeEngineUint64(h, policy.CreatedHeight)
	writeEngineUint64(h, policy.ExecutionHeight)
	writeEngineUint64(h, policy.ActivationHeight)
	writeEnginePart(h, policy.RuntimeVersion)
	writeEnginePart(h, policy.SchedulerVersion)
	writeEnginePart(h, policy.GasPolicyVersion)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMUpgradeManifestHash(manifest AVMUpgradeManifest) string {
	manifest = canonicalAVMUpgradeManifest(manifest)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-upgrade-manifest-v1")
	writeEnginePart(h, manifest.UpgradeID)
	writeEnginePart(h, manifest.GovernanceProposalID)
	writeEngineUint64(h, manifest.StagedHeight)
	writeEngineUint64(h, uint64(len(manifest.Components)))
	for _, component := range manifest.Components {
		writeEnginePart(h, component.ComponentHash)
	}
	writeEngineUint64(h, uint64(len(manifest.SchedulerRules)))
	for _, rule := range manifest.SchedulerRules {
		writeEnginePart(h, rule.RuleHash)
	}
	writeEngineUint64(h, uint64(len(manifest.GasTables)))
	for _, table := range manifest.GasTables {
		writeEnginePart(h, table.TableHash)
	}
	writeEngineUint64(h, uint64(len(manifest.Continuations)))
	for _, continuation := range manifest.Continuations {
		writeEnginePart(h, continuation.VersionHash)
	}
	writeEngineUint64(h, uint64(len(manifest.ContractCodes)))
	for _, code := range manifest.ContractCodes {
		writeEnginePart(h, code.VersionHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMUpgradeComponent(component AVMUpgradeComponent) AVMUpgradeComponent {
	component.PreviousVersion = strings.TrimSpace(component.PreviousVersion)
	component.NextVersion = strings.TrimSpace(component.NextVersion)
	component.ComponentHash = strings.TrimSpace(component.ComponentHash)
	return component
}

func canonicalAVMSchedulerRuleVersion(rule AVMSchedulerRuleVersion) AVMSchedulerRuleVersion {
	rule.RuleSetID = strings.TrimSpace(rule.RuleSetID)
	rule.Version = strings.TrimSpace(rule.Version)
	rule.RuleHash = strings.TrimSpace(rule.RuleHash)
	return rule
}

func canonicalAVMGasTableActivation(table AVMGasTableActivation) AVMGasTableActivation {
	table.PolicyVersion = strings.TrimSpace(table.PolicyVersion)
	table.Policy = canonicalAVMGasPolicy(table.Policy)
	table.Schedule = canonicalAVMGasSchedule(table.Schedule)
	table.TableHash = strings.TrimSpace(table.TableHash)
	return table
}

func canonicalAVMContinuationRuntimeVersion(version AVMContinuationRuntimeVersion) AVMContinuationRuntimeVersion {
	version.ContinuationID = strings.TrimSpace(version.ContinuationID)
	version.ActorID = strings.TrimSpace(version.ActorID)
	version.RuntimeVersion = strings.TrimSpace(version.RuntimeVersion)
	version.VersionHash = strings.TrimSpace(version.VersionHash)
	return version
}

func canonicalAVMContractCodeVMVersion(version AVMContractCodeVMVersion) AVMContractCodeVMVersion {
	version.CodeHash = strings.TrimSpace(version.CodeHash)
	version.VMVersion = strings.TrimSpace(version.VMVersion)
	version.VersionHash = strings.TrimSpace(version.VersionHash)
	return version
}

func canonicalAVMVersionedMessageExecutionPolicy(policy AVMVersionedMessageExecutionPolicy) AVMVersionedMessageExecutionPolicy {
	policy.MessageID = strings.TrimSpace(policy.MessageID)
	policy.RuntimeVersion = strings.TrimSpace(policy.RuntimeVersion)
	policy.SchedulerVersion = strings.TrimSpace(policy.SchedulerVersion)
	policy.GasPolicyVersion = strings.TrimSpace(policy.GasPolicyVersion)
	policy.PolicyHash = strings.TrimSpace(policy.PolicyHash)
	return policy
}

func canonicalAVMUpgradeManifest(manifest AVMUpgradeManifest) AVMUpgradeManifest {
	manifest.UpgradeID = strings.TrimSpace(manifest.UpgradeID)
	manifest.GovernanceProposalID = strings.TrimSpace(manifest.GovernanceProposalID)
	manifest.Components = append([]AVMUpgradeComponent(nil), manifest.Components...)
	for i := range manifest.Components {
		manifest.Components[i] = canonicalAVMUpgradeComponent(manifest.Components[i])
	}
	sort.SliceStable(manifest.Components, func(i, j int) bool { return manifest.Components[i].Kind < manifest.Components[j].Kind })
	manifest.SchedulerRules = append([]AVMSchedulerRuleVersion(nil), manifest.SchedulerRules...)
	for i := range manifest.SchedulerRules {
		manifest.SchedulerRules[i] = canonicalAVMSchedulerRuleVersion(manifest.SchedulerRules[i])
	}
	sort.SliceStable(manifest.SchedulerRules, func(i, j int) bool {
		if manifest.SchedulerRules[i].RuleSetID != manifest.SchedulerRules[j].RuleSetID {
			return manifest.SchedulerRules[i].RuleSetID < manifest.SchedulerRules[j].RuleSetID
		}
		return manifest.SchedulerRules[i].EffectiveFromHeight < manifest.SchedulerRules[j].EffectiveFromHeight
	})
	manifest.GasTables = append([]AVMGasTableActivation(nil), manifest.GasTables...)
	for i := range manifest.GasTables {
		manifest.GasTables[i] = canonicalAVMGasTableActivation(manifest.GasTables[i])
	}
	sort.SliceStable(manifest.GasTables, func(i, j int) bool {
		if manifest.GasTables[i].ActivationHeight != manifest.GasTables[j].ActivationHeight {
			return manifest.GasTables[i].ActivationHeight < manifest.GasTables[j].ActivationHeight
		}
		return manifest.GasTables[i].PolicyVersion < manifest.GasTables[j].PolicyVersion
	})
	manifest.Continuations = append([]AVMContinuationRuntimeVersion(nil), manifest.Continuations...)
	for i := range manifest.Continuations {
		manifest.Continuations[i] = canonicalAVMContinuationRuntimeVersion(manifest.Continuations[i])
	}
	sort.SliceStable(manifest.Continuations, func(i, j int) bool {
		return manifest.Continuations[i].ContinuationID < manifest.Continuations[j].ContinuationID
	})
	manifest.ContractCodes = append([]AVMContractCodeVMVersion(nil), manifest.ContractCodes...)
	for i := range manifest.ContractCodes {
		manifest.ContractCodes[i] = canonicalAVMContractCodeVMVersion(manifest.ContractCodes[i])
	}
	sort.SliceStable(manifest.ContractCodes, func(i, j int) bool { return manifest.ContractCodes[i].CodeID < manifest.ContractCodes[j].CodeID })
	manifest.ManifestHash = strings.TrimSpace(manifest.ManifestHash)
	return manifest
}

func validateAVMUpgradeComponents(components []AVMUpgradeComponent, stagedHeight uint64) error {
	seen := make(map[AVMUpgradeComponentKind]struct{}, len(components))
	for i, component := range components {
		if err := component.Validate(); err != nil {
			return err
		}
		if component.ActivationHeight < stagedHeight {
			return errors.New("AVM upgrade component activation height must be staged before activation")
		}
		if _, found := seen[component.Kind]; found {
			return fmt.Errorf("duplicate AVM upgrade component kind %q", component.Kind)
		}
		seen[component.Kind] = struct{}{}
		if i > 0 && components[i-1].Kind >= component.Kind {
			return errors.New("AVM upgrade components must be sorted canonically")
		}
	}
	return nil
}

func validateAVMSchedulerRuleVersions(rules []AVMSchedulerRuleVersion, stagedHeight uint64) error {
	if len(rules) > MaxAVMUpgradeSchedulerRules {
		return fmt.Errorf("AVM scheduler rule versions must be <= %d", MaxAVMUpgradeSchedulerRules)
	}
	for i, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if rule.EffectiveFromHeight < stagedHeight {
			return errors.New("AVM scheduler rule activation height must be staged")
		}
		if i > 0 {
			prev := rules[i-1]
			if prev.RuleSetID > rule.RuleSetID || (prev.RuleSetID == rule.RuleSetID && prev.EffectiveFromHeight >= rule.EffectiveFromHeight) {
				return errors.New("AVM scheduler rules must be sorted canonically")
			}
			if prev.RuleSetID == rule.RuleSetID {
				prevUntil := prev.EffectiveUntilHeight
				if prevUntil == 0 || prevUntil >= rule.EffectiveFromHeight {
					return errors.New("old and new AVM scheduler rules must not overlap in one block")
				}
			}
		}
	}
	return nil
}

func validateAVMGasTableActivations(tables []AVMGasTableActivation, stagedHeight uint64) error {
	if len(tables) > MaxAVMUpgradeComponents {
		return fmt.Errorf("AVM gas table activations must be <= %d", MaxAVMUpgradeComponents)
	}
	seen := make(map[string]struct{}, len(tables))
	for i, table := range tables {
		if err := table.Validate(); err != nil {
			return err
		}
		if table.ActivationHeight < stagedHeight {
			return errors.New("AVM gas table changes apply only at staged activation height")
		}
		key := fmt.Sprintf("%020d/%s", table.ActivationHeight, table.PolicyVersion)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate AVM gas table activation %q", key)
		}
		seen[key] = struct{}{}
		if i > 0 {
			prevKey := fmt.Sprintf("%020d/%s", tables[i-1].ActivationHeight, tables[i-1].PolicyVersion)
			if prevKey >= key {
				return errors.New("AVM gas table activations must be sorted canonically")
			}
		}
	}
	return nil
}

func validateAVMContinuationRuntimeVersions(versions []AVMContinuationRuntimeVersion) error {
	if len(versions) > MaxAVMUpgradeVersionedItems {
		return fmt.Errorf("AVM continuation runtime versions must be <= %d", MaxAVMUpgradeVersionedItems)
	}
	seen := make(map[string]struct{}, len(versions))
	for i, version := range versions {
		if err := version.Validate(); err != nil {
			return err
		}
		if _, found := seen[version.ContinuationID]; found {
			return fmt.Errorf("duplicate AVM continuation runtime version %q", version.ContinuationID)
		}
		seen[version.ContinuationID] = struct{}{}
		if i > 0 && versions[i-1].ContinuationID >= version.ContinuationID {
			return errors.New("AVM continuation runtime versions must be sorted canonically")
		}
	}
	return nil
}

func validateAVMContractCodeVMVersions(versions []AVMContractCodeVMVersion) error {
	if len(versions) > MaxAVMUpgradeVersionedItems {
		return fmt.Errorf("AVM contract code VM versions must be <= %d", MaxAVMUpgradeVersionedItems)
	}
	seen := make(map[uint64]struct{}, len(versions))
	for i, version := range versions {
		if err := version.Validate(); err != nil {
			return err
		}
		if _, found := seen[version.CodeID]; found {
			return fmt.Errorf("duplicate AVM contract code VM version %d", version.CodeID)
		}
		seen[version.CodeID] = struct{}{}
		if i > 0 && versions[i-1].CodeID >= version.CodeID {
			return errors.New("AVM contract code VM versions must be sorted canonically")
		}
	}
	return nil
}
