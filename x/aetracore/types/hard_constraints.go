package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type HardConstraintID string

const (
	HardConstraintNoNondeterministicConsensusLogic	HardConstraintID	= "no-nondeterministic-consensus-logic"
	HardConstraintNoExternalNetworkStateTransition	HardConstraintID	= "no-external-network-calls-state-transitions"
	HardConstraintCrossZoneInteractionsMessageOnly	HardConstraintID	= "cross-zone-interactions-message-based"
	HardConstraintModulesExposeMsgQueryKeeper	HardConstraintID	= "modules-expose-msgserver-queryserver-keeper"
	HardConstraintStateExportableReproducible	HardConstraintID	= "state-exportable-reproducible"
	HardConstraintCriticalOutputsCommitmentBacked	HardConstraintID	= "critical-outputs-commitment-backed"
	HardConstraintNoDirectCrossZoneWrites		HardConstraintID	= "no-direct-writes-across-zones"
	HardConstraintBoundedQueueDraining		HardConstraintID	= "no-unbounded-queue-draining"
	HardConstraintBoundedStorageIteration		HardConstraintID	= "no-unbounded-storage-iteration"
	HardConstraintGasAccountedProofVerification	HardConstraintID	= "gas-accounted-proof-verification"
)

type HardConstraint struct {
	ID			HardConstraintID
	EnforcementTarget	string
	Modules			[]CosmosSDKModuleName
	ConsensusCritical	bool
	StateTransitionGuard	bool
	CommitmentBacked	bool
	Evidence		[]string
	ConstraintHash		string
}

type HardConstraintManifest struct {
	Constraints	[]HardConstraint
	ManifestHash	string
}

func DefaultHardConstraintManifest() (HardConstraintManifest, error) {
	allModules := RequiredCosmosSDKModules()
	constraints := []HardConstraint{
		hardConstraint(HardConstraintNoNondeterministicConsensusLogic, "consensus state transitions", allModules, "FinalizeBlock uses consensus context only", "unordered inputs are canonically sorted before execution"),
		hardConstraint(HardConstraintNoExternalNetworkStateTransition, "state transition execution", allModules, "external network calls are outside consensus", "service storage and routing observations enter state only through committed messages or proofs"),
		hardConstraint(HardConstraintCrossZoneInteractionsMessageOnly, "cross-zone module interactions", allModules, "cross-zone mutations route through x/messages", "direct cross-zone keeper writes are rejected"),
		hardConstraint(HardConstraintModulesExposeMsgQueryKeeper, "required Cosmos SDK module surface", allModules, "each required module exposes MsgServer QueryServer and Keeper", "module manifest validation rejects incomplete surfaces"),
		hardConstraint(HardConstraintStateExportableReproducible, "genesis export import and replay", allModules, "module state is exportable and importable", "root metadata round trip reproduces committed state"),
		hardConstraint(HardConstraintCriticalOutputsCommitmentBacked, "critical subsystem outputs", allModules, "roots receipts services identities storage messages payments contracts and routing outputs are commitment backed", "tampered root contribution changes manifest hash"),
		hardConstraint(HardConstraintNoDirectCrossZoneWrites, "zone keeper writes", allModules, "keepers write only their own store key", "cross-zone state changes require message receipts"),
		hardConstraint(HardConstraintBoundedQueueDraining, "message queue draining", []CosmosSDKModuleName{CosmosModuleMessages, CosmosModuleRouting, CosmosModuleZones}, "queue draining uses configured per-block limits", "sender queues preserve deterministic backlog after limit is reached"),
		hardConstraint(HardConstraintBoundedStorageIteration, "storage iteration", []CosmosSDKModuleName{CosmosModuleStorage}, "storage scans are bounded by prefix and limit parameters", "proof generation avoids unbounded object iteration"),
		hardConstraint(HardConstraintGasAccountedProofVerification, "consensus proof verification", []CosmosSDKModuleName{CosmosModuleContracts, CosmosModuleMessages, CosmosModuleStorage}, "proof verification charges gas in consensus execution", "ungas-metered proof verification is rejected"),
	}
	return NewHardConstraintManifest(constraints)
}

func NewHardConstraintManifest(constraints []HardConstraint) (HardConstraintManifest, error) {
	manifest := HardConstraintManifest{Constraints: normalizeHardConstraints(constraints)}
	if err := manifest.ValidateFormat(); err != nil {
		return HardConstraintManifest{}, err
	}
	for i := range manifest.Constraints {
		manifest.Constraints[i].ConstraintHash = ComputeHardConstraintHash(manifest.Constraints[i])
	}
	manifest.ManifestHash = ComputeHardConstraintManifestHash(manifest)
	return manifest, manifest.Validate()
}

func RequiredHardConstraintIDs() []HardConstraintID {
	return []HardConstraintID{
		HardConstraintNoNondeterministicConsensusLogic,
		HardConstraintNoExternalNetworkStateTransition,
		HardConstraintCrossZoneInteractionsMessageOnly,
		HardConstraintModulesExposeMsgQueryKeeper,
		HardConstraintStateExportableReproducible,
		HardConstraintCriticalOutputsCommitmentBacked,
		HardConstraintNoDirectCrossZoneWrites,
		HardConstraintBoundedQueueDraining,
		HardConstraintBoundedStorageIteration,
		HardConstraintGasAccountedProofVerification,
	}
}

func (manifest HardConstraintManifest) ValidateFormat() error {
	manifest.Constraints = normalizeHardConstraints(manifest.Constraints)
	required := requiredHardConstraintIDStrings()
	if len(manifest.Constraints) != len(required) {
		return fmt.Errorf("aetracore hard constraint manifest must include %d required constraints", len(required))
	}
	seen := make(map[HardConstraintID]struct{}, len(manifest.Constraints))
	var previous HardConstraintID
	for i, constraint := range manifest.Constraints {
		if err := constraint.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[constraint.ID]; found {
			return fmt.Errorf("duplicate aetracore hard constraint %s", constraint.ID)
		}
		seen[constraint.ID] = struct{}{}
		if i > 0 && previous >= constraint.ID {
			return errors.New("aetracore hard constraints must be sorted canonically")
		}
		previous = constraint.ID
	}
	for _, id := range required {
		if _, found := seen[HardConstraintID(id)]; !found {
			return fmt.Errorf("missing aetracore hard constraint %s", id)
		}
	}
	if manifest.ManifestHash != "" {
		return ValidateHash("aetracore hard constraint manifest hash", manifest.ManifestHash)
	}
	return nil
}

func (manifest HardConstraintManifest) Validate() error {
	manifest.Constraints = normalizeHardConstraints(manifest.Constraints)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, constraint := range manifest.Constraints {
		if err := constraint.Validate(); err != nil {
			return err
		}
	}
	if manifest.ManifestHash != ComputeHardConstraintManifestHash(manifest) {
		return errors.New("aetracore hard constraint manifest hash mismatch")
	}
	return nil
}

func (constraint HardConstraint) ValidateFormat() error {
	constraint = normalizeHardConstraint(constraint)
	if !IsRequiredHardConstraintID(constraint.ID) {
		return fmt.Errorf("unknown aetracore hard constraint %q", constraint.ID)
	}
	if err := validateRoadmapText("aetracore hard constraint enforcement target", constraint.EnforcementTarget); err != nil {
		return err
	}
	if len(constraint.Modules) == 0 {
		return fmt.Errorf("aetracore hard constraint %s must name at least one module", constraint.ID)
	}
	seenModules := make(map[CosmosSDKModuleName]struct{}, len(constraint.Modules))
	var previousModule CosmosSDKModuleName
	for i, moduleName := range constraint.Modules {
		if !IsRequiredCosmosSDKModule(moduleName) {
			return fmt.Errorf("aetracore hard constraint %s references unknown module %s", constraint.ID, moduleName)
		}
		if _, found := seenModules[moduleName]; found {
			return fmt.Errorf("duplicate aetracore hard constraint module %s", moduleName)
		}
		seenModules[moduleName] = struct{}{}
		if i > 0 && previousModule >= moduleName {
			return fmt.Errorf("aetracore hard constraint %s modules must be sorted canonically", constraint.ID)
		}
		previousModule = moduleName
	}
	if !constraint.ConsensusCritical {
		return fmt.Errorf("aetracore hard constraint %s must be consensus critical", constraint.ID)
	}
	if !constraint.StateTransitionGuard {
		return fmt.Errorf("aetracore hard constraint %s must guard state transitions", constraint.ID)
	}
	if !constraint.CommitmentBacked {
		return fmt.Errorf("aetracore hard constraint %s must be commitment backed", constraint.ID)
	}
	if err := validateHardConstraintEvidence(constraint.ID, constraint.Evidence); err != nil {
		return err
	}
	if constraint.ConstraintHash != "" {
		return ValidateHash("aetracore hard constraint hash", constraint.ConstraintHash)
	}
	return nil
}

func (constraint HardConstraint) Validate() error {
	constraint = normalizeHardConstraint(constraint)
	if err := constraint.ValidateFormat(); err != nil {
		return err
	}
	if constraint.ConstraintHash != ComputeHardConstraintHash(constraint) {
		return errors.New("aetracore hard constraint hash mismatch")
	}
	return nil
}

func HardConstraintByID(manifest HardConstraintManifest, id HardConstraintID) (HardConstraint, bool) {
	for _, constraint := range manifest.Constraints {
		if constraint.ID == id {
			return constraint, true
		}
	}
	return HardConstraint{}, false
}

func IsRequiredHardConstraintID(id HardConstraintID) bool {
	for _, required := range RequiredHardConstraintIDs() {
		if required == id {
			return true
		}
	}
	return false
}

func ComputeHardConstraintHash(constraint HardConstraint) string {
	constraint = normalizeHardConstraint(constraint)
	return hashRoot("aetra-aek-hard-constraint-v1", func(w byteWriter) {
		writePart(w, string(constraint.ID))
		writePart(w, constraint.EnforcementTarget)
		writeUint64(w, uint64(len(constraint.Modules)))
		for _, moduleName := range constraint.Modules {
			writePart(w, string(moduleName))
		}
		writePart(w, fmt.Sprint(constraint.ConsensusCritical))
		writePart(w, fmt.Sprint(constraint.StateTransitionGuard))
		writePart(w, fmt.Sprint(constraint.CommitmentBacked))
		writeStringParts(w, constraint.Evidence)
	})
}

func ComputeHardConstraintManifestHash(manifest HardConstraintManifest) string {
	manifest.Constraints = normalizeHardConstraints(manifest.Constraints)
	return hashRoot("aetra-aek-hard-constraint-manifest-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(manifest.Constraints)))
		for _, constraint := range manifest.Constraints {
			writePart(w, constraint.ConstraintHash)
		}
	})
}

func hardConstraint(id HardConstraintID, target string, modules []CosmosSDKModuleName, evidence ...string) HardConstraint {
	return HardConstraint{
		ID:			id,
		EnforcementTarget:	target,
		Modules:		modules,
		ConsensusCritical:	true,
		StateTransitionGuard:	true,
		CommitmentBacked:	true,
		Evidence:		evidence,
	}
}

func normalizeHardConstraint(constraint HardConstraint) HardConstraint {
	constraint.ID = HardConstraintID(strings.TrimSpace(string(constraint.ID)))
	constraint.EnforcementTarget = strings.TrimSpace(constraint.EnforcementTarget)
	constraint.Modules = append([]CosmosSDKModuleName(nil), constraint.Modules...)
	for i, moduleName := range constraint.Modules {
		constraint.Modules[i] = CosmosSDKModuleName(strings.TrimSpace(string(moduleName)))
	}
	sort.SliceStable(constraint.Modules, func(i, j int) bool {
		return constraint.Modules[i] < constraint.Modules[j]
	})
	constraint.Evidence = normalizeRoadmapStringSet(constraint.Evidence)
	constraint.ConstraintHash = strings.ToLower(strings.TrimSpace(constraint.ConstraintHash))
	return constraint
}

func normalizeHardConstraints(constraints []HardConstraint) []HardConstraint {
	out := make([]HardConstraint, len(constraints))
	for i, constraint := range constraints {
		constraint = normalizeHardConstraint(constraint)
		if constraint.ConstraintHash == "" {
			constraint.ConstraintHash = ComputeHardConstraintHash(constraint)
		}
		out[i] = constraint
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func validateHardConstraintEvidence(id HardConstraintID, evidence []string) error {
	if len(evidence) == 0 {
		return fmt.Errorf("aetracore hard constraint %s evidence is required", id)
	}
	seen := make(map[string]struct{}, len(evidence))
	var previous string
	for i, item := range evidence {
		if err := validateRoadmapText("aetracore hard constraint evidence", item); err != nil {
			return err
		}
		if _, found := seen[item]; found {
			return fmt.Errorf("duplicate aetracore hard constraint evidence %s", item)
		}
		seen[item] = struct{}{}
		if i > 0 && previous >= item {
			return fmt.Errorf("aetracore hard constraint %s evidence must be sorted canonically", id)
		}
		previous = item
	}
	return nil
}

func requiredHardConstraintIDStrings() []string {
	ids := RequiredHardConstraintIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}
