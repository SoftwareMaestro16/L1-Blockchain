package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NonGoalID string

const (
	NonGoalNoMessagingSocialApplicationLayer	NonGoalID	= "no-messaging-or-social-application-layer"
	NonGoalNoUIAssumptions				NonGoalID	= "no-ui-assumptions"
	NonGoalNoCentralizedServiceDependency		NonGoalID	= "no-centralized-service-dependency"
	NonGoalNoExternalAPIReliance			NonGoalID	= "no-external-api-reliance"
	NonGoalNoCanonicalOffchainResultWithoutProof	NonGoalID	= "no-offchain-service-result-canonical-without-proof-or-receipt"
	NonGoalNoDirectSynchronousCrossZoneFunctionCall	NonGoalID	= "no-direct-synchronous-cross-zone-function-calls"
)

type NonGoal struct {
	ID			NonGoalID
	Boundary		string
	Modules			[]CosmosSDKModuleName
	Forbidden		bool
	ConsensusBoundary	bool
	CommitmentRequired	bool
	Rationale		[]string
	NonGoalHash		string
}

type NonGoalManifest struct {
	NonGoals	[]NonGoal
	ManifestHash	string
}

func DefaultNonGoalManifest() (NonGoalManifest, error) {
	allModules := RequiredCosmosSDKModules()
	nonGoals := []NonGoal{
		nonGoal(NonGoalNoMessagingSocialApplicationLayer, "application layer scope", allModules, false, "core modules expose routing and receipts only", "social messaging products remain outside Aether Core scope"),
		nonGoal(NonGoalNoUIAssumptions, "client interface scope", []CosmosSDKModuleName{CosmosModuleAetraCore, CosmosModuleServices, CosmosModuleZones}, false, "protocol state does not depend on client layout or UI flow", "queries expose data without presentation assumptions"),
		nonGoal(NonGoalNoCentralizedServiceDependency, "service discovery dependency model", []CosmosSDKModuleName{CosmosModuleRouting, CosmosModuleServices}, false, "service discovery uses committed descriptors and indexes", "centralized service availability cannot gate consensus"),
		nonGoal(NonGoalNoExternalAPIReliance, "consensus data source boundary", allModules, false, "external APIs cannot influence state transitions", "off-chain observations become state only through committed messages proofs or receipts"),
		nonGoal(NonGoalNoCanonicalOffchainResultWithoutProof, "off-chain service result authority", []CosmosSDKModuleName{CosmosModuleMessages, CosmosModuleServices, CosmosModuleStorage}, true, "off-chain service result is advisory until committed", "canonical result requires committed proof or receipt"),
		nonGoal(NonGoalNoDirectSynchronousCrossZoneFunctionCall, "cross-zone execution boundary", allModules, false, "cross-zone execution is asynchronous and message based", "synchronous cross-zone keeper calls cannot mutate state"),
	}
	return NewNonGoalManifest(nonGoals)
}

func NewNonGoalManifest(nonGoals []NonGoal) (NonGoalManifest, error) {
	manifest := NonGoalManifest{NonGoals: normalizeNonGoals(nonGoals)}
	if err := manifest.ValidateFormat(); err != nil {
		return NonGoalManifest{}, err
	}
	for i := range manifest.NonGoals {
		manifest.NonGoals[i].NonGoalHash = ComputeNonGoalHash(manifest.NonGoals[i])
	}
	manifest.ManifestHash = ComputeNonGoalManifestHash(manifest)
	return manifest, manifest.Validate()
}

func RequiredNonGoalIDs() []NonGoalID {
	return []NonGoalID{
		NonGoalNoMessagingSocialApplicationLayer,
		NonGoalNoUIAssumptions,
		NonGoalNoCentralizedServiceDependency,
		NonGoalNoExternalAPIReliance,
		NonGoalNoCanonicalOffchainResultWithoutProof,
		NonGoalNoDirectSynchronousCrossZoneFunctionCall,
	}
}

func (manifest NonGoalManifest) ValidateFormat() error {
	manifest.NonGoals = normalizeNonGoals(manifest.NonGoals)
	required := requiredNonGoalIDStrings()
	if len(manifest.NonGoals) != len(required) {
		return fmt.Errorf("aetracore non-goal manifest must include %d required non-goals", len(required))
	}
	seen := make(map[NonGoalID]struct{}, len(manifest.NonGoals))
	var previous NonGoalID
	for i, nonGoal := range manifest.NonGoals {
		if err := nonGoal.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[nonGoal.ID]; found {
			return fmt.Errorf("duplicate aetracore non-goal %s", nonGoal.ID)
		}
		seen[nonGoal.ID] = struct{}{}
		if i > 0 && previous >= nonGoal.ID {
			return errors.New("aetracore non-goals must be sorted canonically")
		}
		previous = nonGoal.ID
	}
	for _, id := range required {
		if _, found := seen[NonGoalID(id)]; !found {
			return fmt.Errorf("missing aetracore non-goal %s", id)
		}
	}
	if manifest.ManifestHash != "" {
		return ValidateHash("aetracore non-goal manifest hash", manifest.ManifestHash)
	}
	return nil
}

func (manifest NonGoalManifest) Validate() error {
	manifest.NonGoals = normalizeNonGoals(manifest.NonGoals)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, nonGoal := range manifest.NonGoals {
		if err := nonGoal.Validate(); err != nil {
			return err
		}
	}
	if manifest.ManifestHash != ComputeNonGoalManifestHash(manifest) {
		return errors.New("aetracore non-goal manifest hash mismatch")
	}
	return nil
}

func (nonGoal NonGoal) ValidateFormat() error {
	nonGoal = normalizeNonGoal(nonGoal)
	if !IsRequiredNonGoalID(nonGoal.ID) {
		return fmt.Errorf("unknown aetracore non-goal %q", nonGoal.ID)
	}
	if err := validateRoadmapText("aetracore non-goal boundary", nonGoal.Boundary); err != nil {
		return err
	}
	if len(nonGoal.Modules) == 0 {
		return fmt.Errorf("aetracore non-goal %s must name at least one module", nonGoal.ID)
	}
	seenModules := make(map[CosmosSDKModuleName]struct{}, len(nonGoal.Modules))
	var previousModule CosmosSDKModuleName
	for i, moduleName := range nonGoal.Modules {
		if !IsRequiredCosmosSDKModule(moduleName) {
			return fmt.Errorf("aetracore non-goal %s references unknown module %s", nonGoal.ID, moduleName)
		}
		if _, found := seenModules[moduleName]; found {
			return fmt.Errorf("duplicate aetracore non-goal module %s", moduleName)
		}
		seenModules[moduleName] = struct{}{}
		if i > 0 && previousModule >= moduleName {
			return fmt.Errorf("aetracore non-goal %s modules must be sorted canonically", nonGoal.ID)
		}
		previousModule = moduleName
	}
	if !nonGoal.Forbidden {
		return fmt.Errorf("aetracore non-goal %s must be forbidden", nonGoal.ID)
	}
	if !nonGoal.ConsensusBoundary {
		return fmt.Errorf("aetracore non-goal %s must define consensus boundary", nonGoal.ID)
	}
	if nonGoal.ID == NonGoalNoCanonicalOffchainResultWithoutProof && !nonGoal.CommitmentRequired {
		return fmt.Errorf("aetracore non-goal %s must require committed proof or receipt", nonGoal.ID)
	}
	if err := validateNonGoalRationale(nonGoal.ID, nonGoal.Rationale); err != nil {
		return err
	}
	if nonGoal.NonGoalHash != "" {
		return ValidateHash("aetracore non-goal hash", nonGoal.NonGoalHash)
	}
	return nil
}

func (nonGoal NonGoal) Validate() error {
	nonGoal = normalizeNonGoal(nonGoal)
	if err := nonGoal.ValidateFormat(); err != nil {
		return err
	}
	if nonGoal.NonGoalHash != ComputeNonGoalHash(nonGoal) {
		return errors.New("aetracore non-goal hash mismatch")
	}
	return nil
}

func NonGoalByID(manifest NonGoalManifest, id NonGoalID) (NonGoal, bool) {
	for _, nonGoal := range manifest.NonGoals {
		if nonGoal.ID == id {
			return nonGoal, true
		}
	}
	return NonGoal{}, false
}

func IsRequiredNonGoalID(id NonGoalID) bool {
	for _, required := range RequiredNonGoalIDs() {
		if required == id {
			return true
		}
	}
	return false
}

func ComputeNonGoalHash(nonGoal NonGoal) string {
	nonGoal = normalizeNonGoal(nonGoal)
	return hashRoot("aetra-aek-non-goal-v1", func(w byteWriter) {
		writePart(w, string(nonGoal.ID))
		writePart(w, nonGoal.Boundary)
		writeUint64(w, uint64(len(nonGoal.Modules)))
		for _, moduleName := range nonGoal.Modules {
			writePart(w, string(moduleName))
		}
		writePart(w, fmt.Sprint(nonGoal.Forbidden))
		writePart(w, fmt.Sprint(nonGoal.ConsensusBoundary))
		writePart(w, fmt.Sprint(nonGoal.CommitmentRequired))
		writeStringParts(w, nonGoal.Rationale)
	})
}

func ComputeNonGoalManifestHash(manifest NonGoalManifest) string {
	manifest.NonGoals = normalizeNonGoals(manifest.NonGoals)
	return hashRoot("aetra-aek-non-goal-manifest-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(manifest.NonGoals)))
		for _, nonGoal := range manifest.NonGoals {
			writePart(w, nonGoal.NonGoalHash)
		}
	})
}

func nonGoal(id NonGoalID, boundary string, modules []CosmosSDKModuleName, commitmentRequired bool, rationale ...string) NonGoal {
	return NonGoal{
		ID:			id,
		Boundary:		boundary,
		Modules:		modules,
		Forbidden:		true,
		ConsensusBoundary:	true,
		CommitmentRequired:	commitmentRequired,
		Rationale:		rationale,
	}
}

func normalizeNonGoal(nonGoal NonGoal) NonGoal {
	nonGoal.ID = NonGoalID(strings.TrimSpace(string(nonGoal.ID)))
	nonGoal.Boundary = strings.TrimSpace(nonGoal.Boundary)
	nonGoal.Modules = append([]CosmosSDKModuleName(nil), nonGoal.Modules...)
	for i, moduleName := range nonGoal.Modules {
		nonGoal.Modules[i] = CosmosSDKModuleName(strings.TrimSpace(string(moduleName)))
	}
	sort.SliceStable(nonGoal.Modules, func(i, j int) bool {
		return nonGoal.Modules[i] < nonGoal.Modules[j]
	})
	nonGoal.Rationale = normalizeRoadmapStringSet(nonGoal.Rationale)
	nonGoal.NonGoalHash = strings.ToLower(strings.TrimSpace(nonGoal.NonGoalHash))
	return nonGoal
}

func normalizeNonGoals(nonGoals []NonGoal) []NonGoal {
	out := make([]NonGoal, len(nonGoals))
	for i, nonGoal := range nonGoals {
		nonGoal = normalizeNonGoal(nonGoal)
		if nonGoal.NonGoalHash == "" {
			nonGoal.NonGoalHash = ComputeNonGoalHash(nonGoal)
		}
		out[i] = nonGoal
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func validateNonGoalRationale(id NonGoalID, rationale []string) error {
	if len(rationale) == 0 {
		return fmt.Errorf("aetracore non-goal %s rationale is required", id)
	}
	seen := make(map[string]struct{}, len(rationale))
	var previous string
	for i, item := range rationale {
		if err := validateRoadmapText("aetracore non-goal rationale", item); err != nil {
			return err
		}
		if _, found := seen[item]; found {
			return fmt.Errorf("duplicate aetracore non-goal rationale %s", item)
		}
		seen[item] = struct{}{}
		if i > 0 && previous >= item {
			return fmt.Errorf("aetracore non-goal %s rationale must be sorted canonically", id)
		}
		previous = item
	}
	return nil
}

func requiredNonGoalIDStrings() []string {
	ids := RequiredNonGoalIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}
