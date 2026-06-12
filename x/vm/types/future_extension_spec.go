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
	AVMFutureExtensionSpeculativeExecution		AVMFutureExtensionArea	= "speculative_execution_layer"
	AVMFutureExtensionParallelActorScheduling	AVMFutureExtensionArea	= "parallel_actor_scheduling"
	AVMFutureExtensionZKExecutionAttestation	AVMFutureExtensionArea	= "zero_knowledge_execution_attestation"
	AVMFutureExtensionDistributedScheduler		AVMFutureExtensionArea	= "distributed_async_scheduler"
	AVMFutureExtensionCrossChainBridge		AVMFutureExtensionArea	= "cross_chain_message_bridge_layer"
	AVMFutureExtensionActorStateRent		AVMFutureExtensionArea	= "actor_state_rent"
	AVMFutureExtensionInterfacePackageRegistry	AVMFutureExtensionArea	= "interface_package_registry"
	AVMFutureExtensionFormalVMVerification		AVMFutureExtensionArea	= "formal_vm_verification_test_suite"
	AVMFutureExtensionReplayDebugger		AVMFutureExtensionArea	= "deterministic_replay_debugger"

	AVMFutureExtensionStatusPlanned		AVMFutureExtensionStatus	= "planned"
	AVMFutureExtensionStatusResearch	AVMFutureExtensionStatus	= "research"
	AVMFutureExtensionStatusExperimental	AVMFutureExtensionStatus	= "experimental"

	MaxAVMFutureExtensions		= 32
	MaxAVMFutureExtensionText	= 160
	MaxAVMFuturePrerequisites	= 16
	MaxAVMFuturePrerequisiteName	= 96
)

type AVMFutureExtensionArea string
type AVMFutureExtensionStatus string

type AVMFutureExtensionDescriptor struct {
	Area			AVMFutureExtensionArea
	Name			string
	Description		string
	Prerequisites		[]string
	ConsensusAffecting	bool
	RequiresGovernedUpgrade	bool
	Status			AVMFutureExtensionStatus
	DescriptorHash		string
}

type AVMFutureExtensionRegistry struct {
	RegistryName	string
	Extensions	[]AVMFutureExtensionDescriptor
	RegistryHash	string
}

func DefaultAVMFutureExtensionRegistry() (AVMFutureExtensionRegistry, error) {
	descriptors := []AVMFutureExtensionDescriptor{
		{
			Area:				AVMFutureExtensionSpeculativeExecution,
			Name:				"Speculative execution layer",
			Description:			"Pre-execute eligible deterministic workloads while preserving commit-time validation.",
			Prerequisites:			[]string{"blockstm_conflict_model", "deterministic_replay_tests"},
			ConsensusAffecting:		true,
			RequiresGovernedUpgrade:	true,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionParallelActorScheduling,
			Name:				"Parallel actor scheduling",
			Description:			"Schedule independent actor mailboxes concurrently with actor-local conflict isolation.",
			Prerequisites:			[]string{"actor_mailbox_roots", "blockstm_conflict_model"},
			ConsensusAffecting:		true,
			RequiresGovernedUpgrade:	true,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionZKExecutionAttestation,
			Name:				"Zero-knowledge execution attestation",
			Description:			"Attach proof-oriented execution attestations to AVM receipts and roots.",
			Prerequisites:			[]string{"receipt_roots", "proof_gas_metering"},
			ConsensusAffecting:		true,
			RequiresGovernedUpgrade:	true,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionDistributedScheduler,
			Name:				"Distributed async scheduler",
			Description:			"Coordinate async scheduling across deterministic scheduler shards.",
			Prerequisites:			[]string{"queue_roots", "scheduler_safety_checks"},
			ConsensusAffecting:		true,
			RequiresGovernedUpgrade:	true,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionCrossChainBridge,
			Name:				"Cross-chain message bridge layer",
			Description:			"Bridge AVM async messages across chains with proof-bound ingress and egress.",
			Prerequisites:			[]string{"cross_zone_proofs", "message_replay_protection"},
			ConsensusAffecting:		true,
			RequiresGovernedUpgrade:	true,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionActorStateRent,
			Name:				"Actor state rent",
			Description:			"Meter long-lived actor state with deterministic rent accounting.",
			Prerequisites:			[]string{"actor_state_prefixes", "storage_byte_metering"},
			ConsensusAffecting:		true,
			RequiresGovernedUpgrade:	true,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionInterfacePackageRegistry,
			Name:				"Interface package registry",
			Description:			"Publish interface descriptors as versioned packages for SDKs, CLIs, and wallets.",
			Prerequisites:			[]string{"interface_hash_verification", "sdk_binding_metadata"},
			ConsensusAffecting:		false,
			RequiresGovernedUpgrade:	false,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionFormalVMVerification,
			Name:				"Formal VM verification test suite",
			Description:			"Add formal and replay-focused VM verification cases for consensus-critical behavior.",
			Prerequisites:			[]string{"determinism_tests", "state_invariant_tests"},
			ConsensusAffecting:		false,
			RequiresGovernedUpgrade:	false,
			Status:				AVMFutureExtensionStatusPlanned,
		},
		{
			Area:				AVMFutureExtensionReplayDebugger,
			Name:				"Deterministic replay debugger",
			Description:			"Inspect deterministic execution traces without changing consensus outputs.",
			Prerequisites:			[]string{"execution_receipts", "replay_export_import"},
			ConsensusAffecting:		false,
			RequiresGovernedUpgrade:	false,
			Status:				AVMFutureExtensionStatusPlanned,
		},
	}
	for i := range descriptors {
		descriptor, err := NewAVMFutureExtensionDescriptor(descriptors[i])
		if err != nil {
			return AVMFutureExtensionRegistry{}, err
		}
		descriptors[i] = descriptor
	}
	return NewAVMFutureExtensionRegistry(AVMFutureExtensionRegistry{
		RegistryName:	"AVM future extensions",
		Extensions:	descriptors,
	})
}

func NewAVMFutureExtensionDescriptor(descriptor AVMFutureExtensionDescriptor) (AVMFutureExtensionDescriptor, error) {
	descriptor = canonicalAVMFutureExtensionDescriptor(descriptor)
	descriptor.DescriptorHash = ComputeAVMFutureExtensionDescriptorHash(descriptor)
	return descriptor, descriptor.Validate()
}

func (d AVMFutureExtensionDescriptor) Validate() error {
	d = canonicalAVMFutureExtensionDescriptor(d)
	if !IsAVMFutureExtensionArea(d.Area) {
		return fmt.Errorf("invalid AVM future extension area %q", d.Area)
	}
	if err := validateAVMFutureExtensionText("AVM future extension name", d.Name); err != nil {
		return err
	}
	if err := validateAVMFutureExtensionText("AVM future extension description", d.Description); err != nil {
		return err
	}
	if len(d.Prerequisites) == 0 || len(d.Prerequisites) > MaxAVMFuturePrerequisites {
		return fmt.Errorf("AVM future extension prerequisites must be 1..%d", MaxAVMFuturePrerequisites)
	}
	if err := validateEngineTokens("AVM future extension prerequisite", d.Prerequisites, MaxAVMFuturePrerequisiteName); err != nil {
		return err
	}
	if d.ConsensusAffecting && !d.RequiresGovernedUpgrade {
		return errors.New("AVM consensus-affecting future extension requires governed upgrade")
	}
	if !IsAVMFutureExtensionStatus(d.Status) {
		return fmt.Errorf("invalid AVM future extension status %q", d.Status)
	}
	if d.DescriptorHash == "" {
		return errors.New("AVM future extension descriptor hash is required")
	}
	if err := validateAVMComparisonHash("AVM future extension descriptor hash", d.DescriptorHash); err != nil {
		return err
	}
	if d.DescriptorHash != ComputeAVMFutureExtensionDescriptorHash(d) {
		return errors.New("AVM future extension descriptor hash mismatch")
	}
	return nil
}

func NewAVMFutureExtensionRegistry(registry AVMFutureExtensionRegistry) (AVMFutureExtensionRegistry, error) {
	registry = canonicalAVMFutureExtensionRegistry(registry)
	registry.RegistryHash = ComputeAVMFutureExtensionRegistryHash(registry)
	return registry, registry.Validate()
}

func (r AVMFutureExtensionRegistry) Validate() error {
	r = canonicalAVMFutureExtensionRegistry(r)
	if err := validateAVMFutureExtensionText("AVM future extension registry name", r.RegistryName); err != nil {
		return err
	}
	required := AllAVMFutureExtensionAreas()
	if len(r.Extensions) != len(required) || len(r.Extensions) > MaxAVMFutureExtensions {
		return errors.New("AVM future extension registry must contain every section 17 extension area")
	}
	seen := make(map[AVMFutureExtensionArea]struct{}, len(r.Extensions))
	for i, descriptor := range r.Extensions {
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if _, found := seen[descriptor.Area]; found {
			return fmt.Errorf("duplicate AVM future extension area %q", descriptor.Area)
		}
		seen[descriptor.Area] = struct{}{}
		if i > 0 && r.Extensions[i-1].Area >= descriptor.Area {
			return errors.New("AVM future extensions must be sorted canonically")
		}
	}
	for _, area := range required {
		if _, found := seen[area]; !found {
			return fmt.Errorf("missing AVM future extension area %q", area)
		}
	}
	if r.RegistryHash == "" {
		return errors.New("AVM future extension registry hash is required")
	}
	if err := validateAVMComparisonHash("AVM future extension registry hash", r.RegistryHash); err != nil {
		return err
	}
	if r.RegistryHash != ComputeAVMFutureExtensionRegistryHash(r) {
		return errors.New("AVM future extension registry hash mismatch")
	}
	return nil
}

func RenderAVMFutureExtensionsMarkdown(registry AVMFutureExtensionRegistry) (string, error) {
	registry = canonicalAVMFutureExtensionRegistry(registry)
	if err := registry.Validate(); err != nil {
		return "", err
	}
	lines := make([]string, 0, len(registry.Extensions))
	for _, descriptor := range sortAVMFutureExtensionsByDocumentOrder(registry.Extensions) {
		lines = append(lines, "- "+descriptor.Name+".")
	}
	return strings.Join(lines, "\n"), nil
}

func AllAVMFutureExtensionAreas() []AVMFutureExtensionArea {
	areas := []AVMFutureExtensionArea{
		AVMFutureExtensionActorStateRent,
		AVMFutureExtensionCrossChainBridge,
		AVMFutureExtensionDistributedScheduler,
		AVMFutureExtensionFormalVMVerification,
		AVMFutureExtensionInterfacePackageRegistry,
		AVMFutureExtensionParallelActorScheduling,
		AVMFutureExtensionReplayDebugger,
		AVMFutureExtensionSpeculativeExecution,
		AVMFutureExtensionZKExecutionAttestation,
	}
	sort.Slice(areas, func(i, j int) bool { return areas[i] < areas[j] })
	return areas
}

func IsAVMFutureExtensionArea(area AVMFutureExtensionArea) bool {
	switch area {
	case AVMFutureExtensionSpeculativeExecution,
		AVMFutureExtensionParallelActorScheduling,
		AVMFutureExtensionZKExecutionAttestation,
		AVMFutureExtensionDistributedScheduler,
		AVMFutureExtensionCrossChainBridge,
		AVMFutureExtensionActorStateRent,
		AVMFutureExtensionInterfacePackageRegistry,
		AVMFutureExtensionFormalVMVerification,
		AVMFutureExtensionReplayDebugger:
		return true
	default:
		return false
	}
}

func IsAVMFutureExtensionStatus(status AVMFutureExtensionStatus) bool {
	switch status {
	case AVMFutureExtensionStatusPlanned,
		AVMFutureExtensionStatusResearch,
		AVMFutureExtensionStatusExperimental:
		return true
	default:
		return false
	}
}

func ComputeAVMFutureExtensionDescriptorHash(descriptor AVMFutureExtensionDescriptor) string {
	descriptor = canonicalAVMFutureExtensionDescriptor(descriptor)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-future-extension-descriptor-v1")
	writeEnginePart(h, string(descriptor.Area))
	writeEnginePart(h, descriptor.Name)
	writeEnginePart(h, descriptor.Description)
	writeEngineUint64(h, uint64(len(descriptor.Prerequisites)))
	for _, prerequisite := range descriptor.Prerequisites {
		writeEnginePart(h, prerequisite)
	}
	writeEngineBool(h, descriptor.ConsensusAffecting)
	writeEngineBool(h, descriptor.RequiresGovernedUpgrade)
	writeEnginePart(h, string(descriptor.Status))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMFutureExtensionRegistryHash(registry AVMFutureExtensionRegistry) string {
	registry = canonicalAVMFutureExtensionRegistry(registry)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-future-extension-registry-v1")
	writeEnginePart(h, registry.RegistryName)
	writeEngineUint64(h, uint64(len(registry.Extensions)))
	for _, descriptor := range registry.Extensions {
		writeEnginePart(h, descriptor.DescriptorHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMFutureExtensionDescriptor(descriptor AVMFutureExtensionDescriptor) AVMFutureExtensionDescriptor {
	descriptor.Area = AVMFutureExtensionArea(strings.TrimSpace(string(descriptor.Area)))
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.Description = strings.TrimSpace(descriptor.Description)
	descriptor.Prerequisites = cloneSortedStrings(descriptor.Prerequisites)
	descriptor.Status = AVMFutureExtensionStatus(strings.TrimSpace(string(descriptor.Status)))
	descriptor.DescriptorHash = strings.TrimSpace(descriptor.DescriptorHash)
	return descriptor
}

func canonicalAVMFutureExtensionRegistry(registry AVMFutureExtensionRegistry) AVMFutureExtensionRegistry {
	registry.RegistryName = strings.TrimSpace(registry.RegistryName)
	registry.Extensions = append([]AVMFutureExtensionDescriptor(nil), registry.Extensions...)
	for i := range registry.Extensions {
		registry.Extensions[i] = canonicalAVMFutureExtensionDescriptor(registry.Extensions[i])
	}
	sort.SliceStable(registry.Extensions, func(i, j int) bool {
		return registry.Extensions[i].Area < registry.Extensions[j].Area
	})
	registry.RegistryHash = strings.TrimSpace(registry.RegistryHash)
	return registry
}

func sortAVMFutureExtensionsByDocumentOrder(extensions []AVMFutureExtensionDescriptor) []AVMFutureExtensionDescriptor {
	out := append([]AVMFutureExtensionDescriptor(nil), extensions...)
	order := map[AVMFutureExtensionArea]int{
		AVMFutureExtensionSpeculativeExecution:		0,
		AVMFutureExtensionParallelActorScheduling:	1,
		AVMFutureExtensionZKExecutionAttestation:	2,
		AVMFutureExtensionDistributedScheduler:		3,
		AVMFutureExtensionCrossChainBridge:		4,
		AVMFutureExtensionActorStateRent:		5,
		AVMFutureExtensionInterfacePackageRegistry:	6,
		AVMFutureExtensionFormalVMVerification:		7,
		AVMFutureExtensionReplayDebugger:		8,
	}
	sort.SliceStable(out, func(i, j int) bool {
		return order[out[i].Area] < order[out[j].Area]
	})
	return out
}

func validateAVMFutureExtensionText(fieldName, value string) error {
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", fieldName)
	}
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > MaxAVMFutureExtensionText {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxAVMFutureExtensionText)
	}
	for _, r := range value {
		if r < 0x20 || r == '|' {
			return fmt.Errorf("%s contains invalid character", fieldName)
		}
	}
	return nil
}
