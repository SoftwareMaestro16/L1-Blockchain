package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type CosmosSDKModuleName string

const (
	CosmosModuleAetherCore CosmosSDKModuleName = "aethercore"
	CosmosModuleZones      CosmosSDKModuleName = "zones"
	CosmosModuleMessages   CosmosSDKModuleName = "messages"
	CosmosModuleServices   CosmosSDKModuleName = "services"
	CosmosModuleStorage    CosmosSDKModuleName = "storage"
	CosmosModuleIdentity   CosmosSDKModuleName = "identity"
	CosmosModuleRouting    CosmosSDKModuleName = "routing"
	CosmosModulePayments   CosmosSDKModuleName = "payments"
	CosmosModuleContracts  CosmosSDKModuleName = "contracts"
)

type CosmosModuleSurface struct {
	ModuleName       CosmosSDKModuleName
	ModulePath       string
	MsgServer        bool
	QueryServer      bool
	Keeper           bool
	Params           bool
	GenesisExport    bool
	GenesisImport    bool
	Invariants       bool
	Events           []string
	TypedErrors      []string
	RootContribution RootContribution
	SurfaceHash      string
}

type CosmosModuleRequirementManifest struct {
	Modules      []CosmosModuleSurface
	ManifestHash string
}

func DefaultCosmosModuleRequirementManifest() (CosmosModuleRequirementManifest, error) {
	modules := []CosmosModuleSurface{
		cosmosModuleSurface(CosmosModuleAetherCore, "x/aethercore", RootType("aethercore"), []string{"aethercore.params_updated", "aethercore.global_root_committed"}, []string{"ErrInvalidGenesis", "ErrUnauthorized", "ErrRootMismatch"}),
		cosmosModuleSurface(CosmosModuleZones, "x/zones", RootType("zones"), []string{"zones.zone_registered", "zones.commitment_appended"}, []string{"ErrZoneNotFound", "ErrDuplicateZone", "ErrCommitmentMismatch"}),
		cosmosModuleSurface(CosmosModuleMessages, "x/messages", RootType("message"), []string{"messages.queued", "messages.receipted"}, []string{"ErrInvalidEnvelope", "ErrReplay", "ErrExpired"}),
		cosmosModuleSurface(CosmosModuleServices, "x/services", RootType("services"), []string{"services.registered", "services.disabled"}, []string{"ErrServiceNotFound", "ErrUnauthorized", "ErrInterfaceMismatch"}),
		cosmosModuleSurface(CosmosModuleStorage, "x/storage", RootType("storage"), []string{"storage.object_registered", "storage.receipt_submitted"}, []string{"ErrObjectNotFound", "ErrInvalidProof", "ErrPolicyViolation"}),
		cosmosModuleSurface(CosmosModuleIdentity, "x/identity", RootType("identity"), []string{"identity.registered", "identity.resolver_updated"}, []string{"ErrIdentityNotFound", "ErrUnauthorized", "ErrExpired"}),
		cosmosModuleSurface(CosmosModuleRouting, "x/routing", RootType("routing"), []string{"routing.table_updated", "routing.route_selected"}, []string{"ErrRouteNotFound", "ErrInvalidEpoch", "ErrRootMismatch"}),
		cosmosModuleSurface(CosmosModulePayments, "x/payments", RootType("payments"), []string{"payments.channel_opened", "payments.settled"}, []string{"ErrPaymentNotFound", "ErrDisputeWindow", "ErrInsufficientFunds"}),
		cosmosModuleSurface(CosmosModuleContracts, "x/contracts", RootType("contracts"), []string{"contracts.code_stored", "contracts.executed"}, []string{"ErrContractNotFound", "ErrInvalidBytecode", "ErrExecutionFailed"}),
	}
	return NewCosmosModuleRequirementManifest(modules)
}

func NewCosmosModuleRequirementManifest(modules []CosmosModuleSurface) (CosmosModuleRequirementManifest, error) {
	manifest := CosmosModuleRequirementManifest{Modules: normalizeCosmosModuleSurfaces(modules)}
	if err := manifest.ValidateFormat(); err != nil {
		return CosmosModuleRequirementManifest{}, err
	}
	for i := range manifest.Modules {
		manifest.Modules[i].SurfaceHash = ComputeCosmosModuleSurfaceHash(manifest.Modules[i])
	}
	manifest.ManifestHash = ComputeCosmosModuleRequirementManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (surface CosmosModuleSurface) ValidateFormat() error {
	surface = normalizeCosmosModuleSurface(surface)
	if !IsRequiredCosmosSDKModule(surface.ModuleName) {
		return fmt.Errorf("unknown required Cosmos SDK module %q", surface.ModuleName)
	}
	if err := validateToken("aethercore Cosmos module path", surface.ModulePath, MaxScopeLength); err != nil {
		return err
	}
	if !strings.HasPrefix(surface.ModulePath, "x/") {
		return errors.New("aethercore Cosmos module path must be under x/")
	}
	if !surface.MsgServer || !surface.QueryServer || !surface.Keeper || !surface.Params || !surface.GenesisExport || !surface.GenesisImport || !surface.Invariants {
		return fmt.Errorf("aethercore Cosmos module %s is missing required module surface", surface.ModuleName)
	}
	if err := validateSortedRequirementTokens("aethercore Cosmos module event", surface.Events); err != nil {
		return err
	}
	if err := validateSortedRequirementTokens("aethercore Cosmos module typed error", surface.TypedErrors); err != nil {
		return err
	}
	if err := surface.RootContribution.Validate(); err != nil {
		return err
	}
	if surface.SurfaceHash != "" {
		if err := ValidateHash("aethercore Cosmos module surface hash", surface.SurfaceHash); err != nil {
			return err
		}
	}
	return nil
}

func (surface CosmosModuleSurface) Validate() error {
	surface = normalizeCosmosModuleSurface(surface)
	if err := surface.ValidateFormat(); err != nil {
		return err
	}
	if surface.SurfaceHash != ComputeCosmosModuleSurfaceHash(surface) {
		return errors.New("aethercore Cosmos module surface hash mismatch")
	}
	return nil
}

func (manifest CosmosModuleRequirementManifest) ValidateFormat() error {
	manifest.Modules = normalizeCosmosModuleSurfaces(manifest.Modules)
	required := RequiredCosmosSDKModules()
	if len(manifest.Modules) != len(required) {
		return fmt.Errorf("aethercore Cosmos module manifest must include %d required modules", len(required))
	}
	seen := make(map[CosmosSDKModuleName]struct{}, len(manifest.Modules))
	var previous CosmosSDKModuleName
	for i, surface := range manifest.Modules {
		if err := surface.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[surface.ModuleName]; found {
			return fmt.Errorf("duplicate aethercore Cosmos module %s", surface.ModuleName)
		}
		seen[surface.ModuleName] = struct{}{}
		if i > 0 && previous >= surface.ModuleName {
			return errors.New("aethercore Cosmos modules must be sorted canonically")
		}
		previous = surface.ModuleName
	}
	for _, moduleName := range required {
		if _, found := seen[moduleName]; !found {
			return fmt.Errorf("missing required Cosmos SDK module %s", moduleName)
		}
	}
	if manifest.ManifestHash != "" {
		if err := ValidateHash("aethercore Cosmos module manifest hash", manifest.ManifestHash); err != nil {
			return err
		}
	}
	return nil
}

func (manifest CosmosModuleRequirementManifest) Validate() error {
	manifest.Modules = normalizeCosmosModuleSurfaces(manifest.Modules)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, surface := range manifest.Modules {
		if err := surface.Validate(); err != nil {
			return err
		}
	}
	if manifest.ManifestHash != ComputeCosmosModuleRequirementManifestHash(manifest) {
		return errors.New("aethercore Cosmos module manifest hash mismatch")
	}
	return nil
}

func RequiredCosmosSDKModules() []CosmosSDKModuleName {
	return []CosmosSDKModuleName{
		CosmosModuleAetherCore,
		CosmosModuleContracts,
		CosmosModuleIdentity,
		CosmosModuleMessages,
		CosmosModulePayments,
		CosmosModuleRouting,
		CosmosModuleServices,
		CosmosModuleStorage,
		CosmosModuleZones,
	}
}

func IsRequiredCosmosSDKModule(moduleName CosmosSDKModuleName) bool {
	for _, required := range RequiredCosmosSDKModules() {
		if required == moduleName {
			return true
		}
	}
	return false
}

func ComputeCosmosModuleSurfaceHash(surface CosmosModuleSurface) string {
	surface = normalizeCosmosModuleSurface(surface)
	return hashRoot("aetheris-aek-cosmos-module-surface-v1", func(w byteWriter) {
		writePart(w, string(surface.ModuleName))
		writePart(w, surface.ModulePath)
		writePart(w, fmt.Sprint(surface.MsgServer))
		writePart(w, fmt.Sprint(surface.QueryServer))
		writePart(w, fmt.Sprint(surface.Keeper))
		writePart(w, fmt.Sprint(surface.Params))
		writePart(w, fmt.Sprint(surface.GenesisExport))
		writePart(w, fmt.Sprint(surface.GenesisImport))
		writePart(w, fmt.Sprint(surface.Invariants))
		for _, event := range surface.Events {
			writePart(w, event)
		}
		for _, typedError := range surface.TypedErrors {
			writePart(w, typedError)
		}
		writePart(w, surface.RootContribution.ContributionHash)
	})
}

func ComputeCosmosModuleRequirementManifestHash(manifest CosmosModuleRequirementManifest) string {
	manifest.Modules = normalizeCosmosModuleSurfaces(manifest.Modules)
	return hashRoot("aetheris-aek-cosmos-module-requirements-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(manifest.Modules)))
		for _, surface := range manifest.Modules {
			writePart(w, surface.SurfaceHash)
		}
	})
}

func cosmosModuleSurface(moduleName CosmosSDKModuleName, path string, rootType RootType, events []string, typedErrors []string) CosmosModuleSurface {
	root, _ := NewRootContribution(rootType, string(moduleName), DeterministicEmptyRootCommitment(rootType, string(moduleName)))
	return CosmosModuleSurface{
		ModuleName:       moduleName,
		ModulePath:       path,
		MsgServer:        true,
		QueryServer:      true,
		Keeper:           true,
		Params:           true,
		GenesisExport:    true,
		GenesisImport:    true,
		Invariants:       true,
		Events:           events,
		TypedErrors:      typedErrors,
		RootContribution: root,
	}
}

func normalizeCosmosModuleSurface(surface CosmosModuleSurface) CosmosModuleSurface {
	surface.ModuleName = CosmosSDKModuleName(strings.TrimSpace(string(surface.ModuleName)))
	surface.ModulePath = strings.TrimSpace(surface.ModulePath)
	surface.Events = append([]string(nil), surface.Events...)
	surface.TypedErrors = append([]string(nil), surface.TypedErrors...)
	sort.Strings(surface.Events)
	sort.Strings(surface.TypedErrors)
	surface.RootContribution = normalizeRootContribution(surface.RootContribution)
	if surface.RootContribution.ContributionHash == "" && surface.RootContribution.RootHash != "" {
		surface.RootContribution.ContributionHash = ComputeRootContributionHash(surface.RootContribution)
	}
	surface.SurfaceHash = strings.ToLower(strings.TrimSpace(surface.SurfaceHash))
	return surface
}

func normalizeCosmosModuleSurfaces(surfaces []CosmosModuleSurface) []CosmosModuleSurface {
	out := make([]CosmosModuleSurface, len(surfaces))
	for i, surface := range surfaces {
		surface = normalizeCosmosModuleSurface(surface)
		if surface.SurfaceHash == "" {
			surface.SurfaceHash = ComputeCosmosModuleSurfaceHash(surface)
		}
		out[i] = surface
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ModuleName < out[j].ModuleName
	})
	return out
}

func validateSortedRequirementTokens(field string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	seen := make(map[string]struct{}, len(values))
	var previous string
	for i, value := range values {
		if err := validateToken(field, value, MaxScopeLength); err != nil {
			return err
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s %s", field, value)
		}
		seen[value] = struct{}{}
		if i > 0 && previous >= value {
			return fmt.Errorf("%s must be sorted canonically", field)
		}
		previous = value
	}
	return nil
}
