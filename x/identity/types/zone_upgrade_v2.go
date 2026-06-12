package types

import (
	"errors"
	"fmt"
	"sort"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type IdentityZoneCapability string

const (
	IdentityCapabilityCommitRevealRegistration	IdentityZoneCapability	= "commit_reveal_registration"
	IdentityCapabilityNFTOwnershipBinding		IdentityZoneCapability	= "nft_ownership_binding"
	IdentityCapabilityResolverRecords		IdentityZoneCapability	= "resolver_records"
	IdentityCapabilityReverseLookup			IdentityZoneCapability	= "reverse_lookup"
	IdentityCapabilityAuctions			IdentityZoneCapability	= "auctions"
	IdentityCapabilityDeterministicResolverVMHook	IdentityZoneCapability	= "deterministic_resolver_vm_hooks"
	IdentityCapabilityMultiRecordResolutionGraph	IdentityZoneCapability	= "multi_record_resolution_graph"
	IdentityCapabilityCrossZoneIdentityBinding	IdentityZoneCapability	= "cross_zone_identity_binding"
	IdentityCapabilityLightClientProofVerification	IdentityZoneCapability	= "light_client_proof_verification"
)

type IdentityZoneUpgradeManifest struct {
	ZoneID			string
	ZoneType		string
	StorePrefix		string
	ArchitectureHash	string
	ZoneDescriptorHash	string
	StateMachineHash	string
	CapabilityRoot		string
	RequiredCapabilities	[]IdentityZoneCapability
	MessageRoot		string
	ProofQueryRoot		string
	StateRoot		string
	UpgradeHash		string
}

type IdentityZoneUpgradeArtifacts struct {
	State		IdentityState
	Hooks		[]IdentityResolverVMHook
	Graphs		[]IdentityResolutionGraph
	Bindings	[]IdentityCrossZoneBinding
	Proofs		[]IdentityZoneProofIndexEntry
	Height		uint64
}

func DefaultIdentityZoneCapabilities() []IdentityZoneCapability {
	return []IdentityZoneCapability{
		IdentityCapabilityAuctions,
		IdentityCapabilityCommitRevealRegistration,
		IdentityCapabilityCrossZoneIdentityBinding,
		IdentityCapabilityDeterministicResolverVMHook,
		IdentityCapabilityLightClientProofVerification,
		IdentityCapabilityMultiRecordResolutionGraph,
		IdentityCapabilityNFTOwnershipBinding,
		IdentityCapabilityResolverRecords,
		IdentityCapabilityReverseLookup,
	}
}

func DefaultAetraCoreIdentityZoneDescriptor() coretypes.ZoneDescriptor {
	return coretypes.ZoneDescriptor{
		ZoneID:			coretypes.ZoneIDIdentity,
		ZoneName:		"identity",
		ZoneType:		coretypes.ZoneTypeIdentity,
		ModuleName:		"identity",
		Enabled:		true,
		StateMachineVersion:	1,
		StateVersion:		1,
		KeeperScope:		"identity.keeper",
		MsgServerScope:		"identity.msg",
		QueryServerScope:	"identity.query",
		MempoolPolicyID:	coretypes.DefaultMempoolPolicy,
		FeePolicyID:		coretypes.NativeFeePolicyID,
		GasPolicyID:		coretypes.DefaultGasPolicy,
		MessagePolicyID:	coretypes.DefaultMessagePolicy,
		RootPrefix:		IdentityStoreV2Prefix,
		ShardLayoutEpoch:	1,
		MaxShards:		64,
		MessageCapabilities:	[]string{"identity-async-lookup", "identity-receipts", "cross-zone-binding"},
		ProofCapabilities:	[]string{"domain", "resolver", "reverse", "nft-binding", "auction", "identity-root"},
		Capabilities:		identityCapabilitiesAsStrings(DefaultIdentityZoneCapabilities()),
	}
}

func BuildIdentityZoneUpgradeManifest(aekDescriptor coretypes.ZoneDescriptor, zoneDescriptor IdentityZoneStateMachineDescriptor, arch IdentityV2Architecture, roots IdentityZoneRoots, capabilities []IdentityZoneCapability) (IdentityZoneUpgradeManifest, error) {
	capabilities = normalizeIdentityZoneCapabilities(capabilities)
	if len(capabilities) == 0 {
		capabilities = DefaultIdentityZoneCapabilities()
	}
	if err := ValidateIdentityZoneAEKDescriptor(aekDescriptor); err != nil {
		return IdentityZoneUpgradeManifest{}, err
	}
	if err := zoneDescriptor.Validate(); err != nil {
		return IdentityZoneUpgradeManifest{}, err
	}
	if err := ValidateIdentityV2Architecture(arch); err != nil {
		return IdentityZoneUpgradeManifest{}, err
	}
	if err := roots.Validate(); err != nil {
		return IdentityZoneUpgradeManifest{}, err
	}
	architectureHash, err := IdentityV2ArchitectureHash(arch)
	if err != nil {
		return IdentityZoneUpgradeManifest{}, err
	}
	manifest := IdentityZoneUpgradeManifest{
		ZoneID:			string(aekDescriptor.ZoneID),
		ZoneType:		string(aekDescriptor.ZoneType),
		StorePrefix:		zoneDescriptor.StorePrefix,
		ArchitectureHash:	architectureHash,
		ZoneDescriptorHash:	coretypes.ComputeZoneDescriptorHash(aekDescriptor),
		StateMachineHash:	ComputeIdentityZoneStateMachineDescriptorHash(zoneDescriptor),
		CapabilityRoot:		ComputeIdentityZoneCapabilityRoot(capabilities),
		RequiredCapabilities:	capabilities,
		MessageRoot:		ComputeIdentityZoneMessageRoot(zoneDescriptor.MessageHandlers),
		ProofQueryRoot:		ComputeIdentityZoneProofQueryRoot(zoneDescriptor.ProofQueries),
		StateRoot:		roots.StateRoot,
	}
	manifest.UpgradeHash = ComputeIdentityZoneUpgradeManifestHash(manifest)
	return manifest, manifest.Validate()
}

func BuildDefaultIdentityZoneUpgradeManifest(artifacts IdentityZoneUpgradeArtifacts) (IdentityZoneUpgradeManifest, IdentityZoneRoots, error) {
	if err := ValidateIdentityZoneUpgradeArtifacts(artifacts); err != nil {
		return IdentityZoneUpgradeManifest{}, IdentityZoneRoots{}, err
	}
	roots, err := BuildIdentityZoneRoots(artifacts.Height, artifacts.State, artifacts.Hooks, artifacts.Graphs, artifacts.Bindings, artifacts.Proofs)
	if err != nil {
		return IdentityZoneUpgradeManifest{}, IdentityZoneRoots{}, err
	}
	manifest, err := BuildIdentityZoneUpgradeManifest(
		DefaultAetraCoreIdentityZoneDescriptor(),
		DefaultIdentityZoneStateMachineDescriptor(),
		DefaultIdentityV2Architecture(),
		roots,
		DefaultIdentityZoneCapabilities(),
	)
	return manifest, roots, err
}

func ValidateIdentityZoneAEKDescriptor(descriptor coretypes.ZoneDescriptor) error {
	descriptor = coretypes.CanonicalZoneDescriptor(descriptor)
	if err := descriptor.Validate(coretypes.TestnetParams()); err != nil {
		return err
	}
	if descriptor.ZoneID != coretypes.ZoneIDIdentity {
		return errors.New("identity upgrade AEK descriptor must use IDENTITY_ZONE")
	}
	if descriptor.ZoneType != coretypes.ZoneTypeIdentity {
		return errors.New("identity upgrade AEK descriptor must use identity zone type")
	}
	if !descriptor.Enabled {
		return errors.New("identity upgrade AEK descriptor must be enabled")
	}
	if descriptor.ModuleName != "identity" || descriptor.RootPrefix != IdentityStoreV2Prefix {
		return errors.New("identity upgrade AEK descriptor must route to identity module and root prefix")
	}
	for _, capability := range identityCapabilitiesAsStrings(DefaultIdentityZoneCapabilities()) {
		if !stringInSet(capability, descriptor.Capabilities) {
			return fmt.Errorf("identity upgrade AEK descriptor missing capability %q", capability)
		}
	}
	return nil
}

func ValidateIdentityZoneUpgradeArtifacts(artifacts IdentityZoneUpgradeArtifacts) error {
	if artifacts.Height == 0 {
		return errors.New("identity zone upgrade height must be positive")
	}
	if err := artifacts.State.Validate(); err != nil {
		return err
	}
	if artifacts.State.Params.CommitTTLBlocks == 0 {
		return errors.New("identity zone upgrade requires commit/reveal registration params")
	}
	for _, domain := range artifacts.State.Domains {
		if err := CheckIdentityNFTBinding(artifacts.State, domain.Name, artifacts.Height); err != nil {
			return err
		}
	}
	for _, hook := range artifacts.Hooks {
		if err := hook.Validate(); err != nil {
			return err
		}
	}
	for _, graph := range artifacts.Graphs {
		if err := graph.Validate(); err != nil {
			return err
		}
	}
	for _, binding := range artifacts.Bindings {
		if err := binding.Validate(); err != nil {
			return err
		}
	}
	for _, proof := range artifacts.Proofs {
		if err := proof.Validate(); err != nil {
			return err
		}
	}
	if len(artifacts.Hooks) == 0 {
		return errors.New("identity zone upgrade requires deterministic resolver VM hooks")
	}
	if len(artifacts.Graphs) == 0 {
		return errors.New("identity zone upgrade requires a multi-record resolution graph")
	}
	if len(artifacts.Bindings) == 0 {
		return errors.New("identity zone upgrade requires cross-zone identity binding support")
	}
	if len(artifacts.Proofs) == 0 {
		return errors.New("identity zone upgrade requires light-client proof index entries")
	}
	return nil
}

func (manifest IdentityZoneUpgradeManifest) Validate() error {
	if manifest.ZoneID != IdentityZoneID {
		return errors.New("identity zone upgrade manifest must target IDENTITY_ZONE")
	}
	if manifest.ZoneType != string(coretypes.ZoneTypeIdentity) {
		return errors.New("identity zone upgrade manifest must target identity zone type")
	}
	if manifest.StorePrefix != IdentityStoreV2Prefix {
		return errors.New("identity zone upgrade manifest store prefix mismatch")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "identity zone upgrade architecture hash", value: manifest.ArchitectureHash},
		{name: "identity zone upgrade descriptor hash", value: manifest.ZoneDescriptorHash},
		{name: "identity zone upgrade state machine hash", value: manifest.StateMachineHash},
		{name: "identity zone upgrade capability root", value: manifest.CapabilityRoot},
		{name: "identity zone upgrade message root", value: manifest.MessageRoot},
		{name: "identity zone upgrade proof query root", value: manifest.ProofQueryRoot},
		{name: "identity zone upgrade state root", value: manifest.StateRoot},
		{name: "identity zone upgrade hash", value: manifest.UpgradeHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if err := ValidateIdentityZoneCapabilities(manifest.RequiredCapabilities); err != nil {
		return err
	}
	if manifest.CapabilityRoot != ComputeIdentityZoneCapabilityRoot(manifest.RequiredCapabilities) {
		return errors.New("identity zone upgrade capability root mismatch")
	}
	if manifest.UpgradeHash != ComputeIdentityZoneUpgradeManifestHash(manifest) {
		return errors.New("identity zone upgrade manifest hash mismatch")
	}
	return nil
}

func ValidateIdentityZoneCapabilities(capabilities []IdentityZoneCapability) error {
	ordered := normalizeIdentityZoneCapabilities(capabilities)
	if len(ordered) != len(DefaultIdentityZoneCapabilities()) {
		return errors.New("identity zone upgrade requires the complete capability set")
	}
	for i, required := range DefaultIdentityZoneCapabilities() {
		if ordered[i] != required {
			return fmt.Errorf("identity zone upgrade missing capability %q", required)
		}
	}
	for i, capability := range capabilities {
		if !IsIdentityZoneCapability(capability) {
			return fmt.Errorf("unknown identity zone capability %q", capability)
		}
		if i > 0 && capabilities[i-1] >= capability {
			return errors.New("identity zone capabilities must be sorted canonically")
		}
	}
	return nil
}

func IsIdentityZoneCapability(capability IdentityZoneCapability) bool {
	switch capability {
	case IdentityCapabilityCommitRevealRegistration,
		IdentityCapabilityNFTOwnershipBinding,
		IdentityCapabilityResolverRecords,
		IdentityCapabilityReverseLookup,
		IdentityCapabilityAuctions,
		IdentityCapabilityDeterministicResolverVMHook,
		IdentityCapabilityMultiRecordResolutionGraph,
		IdentityCapabilityCrossZoneIdentityBinding,
		IdentityCapabilityLightClientProofVerification:
		return true
	default:
		return false
	}
}

func ComputeIdentityZoneStateMachineDescriptorHash(descriptor IdentityZoneStateMachineDescriptor) string {
	messages := append([]IdentityZoneMessageKind(nil), descriptor.MessageHandlers...)
	proofs := append([]IdentityZoneProofKind(nil), descriptor.ProofQueries...)
	sort.SliceStable(messages, func(i, j int) bool { return messages[i] < messages[j] })
	sort.SliceStable(proofs, func(i, j int) bool { return proofs[i] < proofs[j] })
	parts := []string{"identity-zone-state-machine-descriptor-v1", descriptor.ZoneID, descriptor.StorePrefix, descriptor.ShardStrategy}
	for _, message := range messages {
		parts = append(parts, "msg", string(message))
	}
	for _, proof := range proofs {
		parts = append(parts, "proof", string(proof))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneCapabilityRoot(capabilities []IdentityZoneCapability) string {
	ordered := normalizeIdentityZoneCapabilities(capabilities)
	parts := []string{"identity-zone-capability-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, capability := range ordered {
		parts = append(parts, string(capability))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneMessageRoot(messages []IdentityZoneMessageKind) string {
	ordered := append([]IdentityZoneMessageKind(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	parts := []string{"identity-zone-message-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, message := range ordered {
		parts = append(parts, string(message))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneProofQueryRoot(proofs []IdentityZoneProofKind) string {
	ordered := append([]IdentityZoneProofKind(nil), proofs...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	parts := []string{"identity-zone-proof-query-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, proof := range ordered {
		parts = append(parts, string(proof))
	}
	return identityHash(parts...)
}

func ComputeIdentityZoneUpgradeManifestHash(manifest IdentityZoneUpgradeManifest) string {
	return identityHash(
		"identity-zone-upgrade-manifest-v1",
		manifest.ZoneID,
		manifest.ZoneType,
		manifest.StorePrefix,
		manifest.ArchitectureHash,
		manifest.ZoneDescriptorHash,
		manifest.StateMachineHash,
		manifest.CapabilityRoot,
		manifest.MessageRoot,
		manifest.ProofQueryRoot,
		manifest.StateRoot,
	)
}

func normalizeIdentityZoneCapabilities(capabilities []IdentityZoneCapability) []IdentityZoneCapability {
	out := append([]IdentityZoneCapability(nil), capabilities...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	unique := make([]IdentityZoneCapability, 0, len(out))
	for _, capability := range out {
		if len(unique) == 0 || unique[len(unique)-1] != capability {
			unique = append(unique, capability)
		}
	}
	return unique
}

func identityCapabilitiesAsStrings(capabilities []IdentityZoneCapability) []string {
	ordered := normalizeIdentityZoneCapabilities(capabilities)
	out := make([]string, 0, len(ordered))
	for _, capability := range ordered {
		out = append(out, string(capability))
	}
	return out
}
