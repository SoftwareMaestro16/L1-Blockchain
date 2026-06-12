package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type CosmosSDKModuleName string

const (
	CosmosModuleAetraCore	CosmosSDKModuleName	= "aetracore"
	CosmosModuleZones	CosmosSDKModuleName	= "zones"
	CosmosModuleMessages	CosmosSDKModuleName	= "messages"
	CosmosModuleServices	CosmosSDKModuleName	= "services"
	CosmosModuleStorage	CosmosSDKModuleName	= "storage"
	CosmosModuleIdentity	CosmosSDKModuleName	= "identity"
	CosmosModuleRouting	CosmosSDKModuleName	= "routing"
	CosmosModulePayments	CosmosSDKModuleName	= "payments"
	CosmosModuleContracts	CosmosSDKModuleName	= "contracts"
)

type CosmosModuleSurface struct {
	ModuleName		CosmosSDKModuleName
	ModulePath		string
	MsgServer		bool
	QueryServer		bool
	Keeper			bool
	Params			bool
	GenesisExport		bool
	GenesisImport		bool
	Invariants		bool
	KeeperIsolation		KeeperIsolationPolicy
	IBCBoundary		IBCReadyBoundary
	ABCICompatibility	ABCICompatibilityPolicy
	Events			[]string
	TypedErrors		[]string
	RootContribution	RootContribution
	SurfaceHash		string
}

type KeeperIsolationPolicy struct {
	StoreKey				string
	ReadableStoreKeys			[]string
	WritableStoreKeys			[]string
	GrantedCapabilities			[]string
	CrossZoneWritesProhibited		bool
	CrossZoneMessagesModule			CosmosSDKModuleName
	DirectCallsLimitedToLocalHelpers	bool
	SharedStateReadOnlyOrProofBacked	bool
}

type IBCReadyBoundary struct {
	StateExportable			bool
	ReceiptsProofVerifiable		bool
	CanonicalBoundaryMessages	bool
	TimeoutRulesExplicit		bool
	ReplayRulesExplicit		bool
	DeterministicChannelRouting	bool
	BoundaryMessageEncoding		string
	TimeoutPolicyID			string
	ReplayPolicyID			string
}

type ABCICompatibilityPolicy struct {
	ProposalOptimizationValidityNeutral	bool
	PrecheckDeterministic			bool
	FinalizeBlockAuthoritative		bool
	EndBlockCleanupBounded			bool
	RootAggregationAfterExecution		bool
	CleanupLimitPolicyID			string
	RootAggregationPhase			KernelABCIPhase
	PrecheckInputPolicyID			string
}

type CosmosModuleRequirementManifest struct {
	Modules		[]CosmosModuleSurface
	ManifestHash	string
}

func DefaultCosmosModuleRequirementManifest() (CosmosModuleRequirementManifest, error) {
	modules := []CosmosModuleSurface{
		cosmosModuleSurface(CosmosModuleAetraCore, "x/aetracore", RootType("aetracore"), []string{"aetracore.params_updated", "aetracore.global_root_committed"}, []string{"ErrInvalidGenesis", "ErrUnauthorized", "ErrRootMismatch"}),
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
	if err := validateToken("aetracore Cosmos module path", surface.ModulePath, MaxScopeLength); err != nil {
		return err
	}
	if !strings.HasPrefix(surface.ModulePath, "x/") {
		return errors.New("aetracore Cosmos module path must be under x/")
	}
	if !surface.MsgServer || !surface.QueryServer || !surface.Keeper || !surface.Params || !surface.GenesisExport || !surface.GenesisImport || !surface.Invariants {
		return fmt.Errorf("aetracore Cosmos module %s is missing required module surface", surface.ModuleName)
	}
	if err := surface.KeeperIsolation.Validate(surface.ModuleName); err != nil {
		return err
	}
	if err := surface.IBCBoundary.Validate(surface.ModuleName); err != nil {
		return err
	}
	if err := surface.ABCICompatibility.Validate(surface.ModuleName); err != nil {
		return err
	}
	if err := validateSortedRequirementTokens("aetracore Cosmos module event", surface.Events); err != nil {
		return err
	}
	if err := validateSortedRequirementTokens("aetracore Cosmos module typed error", surface.TypedErrors); err != nil {
		return err
	}
	if err := surface.RootContribution.Validate(); err != nil {
		return err
	}
	if surface.SurfaceHash != "" {
		if err := ValidateHash("aetracore Cosmos module surface hash", surface.SurfaceHash); err != nil {
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
		return errors.New("aetracore Cosmos module surface hash mismatch")
	}
	return nil
}

func (manifest CosmosModuleRequirementManifest) ValidateFormat() error {
	manifest.Modules = normalizeCosmosModuleSurfaces(manifest.Modules)
	required := RequiredCosmosSDKModules()
	if len(manifest.Modules) != len(required) {
		return fmt.Errorf("aetracore Cosmos module manifest must include %d required modules", len(required))
	}
	seen := make(map[CosmosSDKModuleName]struct{}, len(manifest.Modules))
	var previous CosmosSDKModuleName
	for i, surface := range manifest.Modules {
		if err := surface.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[surface.ModuleName]; found {
			return fmt.Errorf("duplicate aetracore Cosmos module %s", surface.ModuleName)
		}
		seen[surface.ModuleName] = struct{}{}
		if i > 0 && previous >= surface.ModuleName {
			return errors.New("aetracore Cosmos modules must be sorted canonically")
		}
		previous = surface.ModuleName
	}
	for _, moduleName := range required {
		if _, found := seen[moduleName]; !found {
			return fmt.Errorf("missing required Cosmos SDK module %s", moduleName)
		}
	}
	if manifest.ManifestHash != "" {
		if err := ValidateHash("aetracore Cosmos module manifest hash", manifest.ManifestHash); err != nil {
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
		return errors.New("aetracore Cosmos module manifest hash mismatch")
	}
	return nil
}

func RequiredCosmosSDKModules() []CosmosSDKModuleName {
	return []CosmosSDKModuleName{
		CosmosModuleAetraCore,
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
	return hashRoot("aetra-aek-cosmos-module-surface-v1", func(w byteWriter) {
		writePart(w, string(surface.ModuleName))
		writePart(w, surface.ModulePath)
		writePart(w, fmt.Sprint(surface.MsgServer))
		writePart(w, fmt.Sprint(surface.QueryServer))
		writePart(w, fmt.Sprint(surface.Keeper))
		writePart(w, fmt.Sprint(surface.Params))
		writePart(w, fmt.Sprint(surface.GenesisExport))
		writePart(w, fmt.Sprint(surface.GenesisImport))
		writePart(w, fmt.Sprint(surface.Invariants))
		writePart(w, surface.KeeperIsolation.StoreKey)
		for _, storeKey := range surface.KeeperIsolation.ReadableStoreKeys {
			writePart(w, storeKey)
		}
		for _, storeKey := range surface.KeeperIsolation.WritableStoreKeys {
			writePart(w, storeKey)
		}
		for _, capability := range surface.KeeperIsolation.GrantedCapabilities {
			writePart(w, capability)
		}
		writePart(w, fmt.Sprint(surface.KeeperIsolation.CrossZoneWritesProhibited))
		writePart(w, string(surface.KeeperIsolation.CrossZoneMessagesModule))
		writePart(w, fmt.Sprint(surface.KeeperIsolation.DirectCallsLimitedToLocalHelpers))
		writePart(w, fmt.Sprint(surface.KeeperIsolation.SharedStateReadOnlyOrProofBacked))
		writePart(w, fmt.Sprint(surface.IBCBoundary.StateExportable))
		writePart(w, fmt.Sprint(surface.IBCBoundary.ReceiptsProofVerifiable))
		writePart(w, fmt.Sprint(surface.IBCBoundary.CanonicalBoundaryMessages))
		writePart(w, fmt.Sprint(surface.IBCBoundary.TimeoutRulesExplicit))
		writePart(w, fmt.Sprint(surface.IBCBoundary.ReplayRulesExplicit))
		writePart(w, fmt.Sprint(surface.IBCBoundary.DeterministicChannelRouting))
		writePart(w, surface.IBCBoundary.BoundaryMessageEncoding)
		writePart(w, surface.IBCBoundary.TimeoutPolicyID)
		writePart(w, surface.IBCBoundary.ReplayPolicyID)
		writePart(w, fmt.Sprint(surface.ABCICompatibility.ProposalOptimizationValidityNeutral))
		writePart(w, fmt.Sprint(surface.ABCICompatibility.PrecheckDeterministic))
		writePart(w, fmt.Sprint(surface.ABCICompatibility.FinalizeBlockAuthoritative))
		writePart(w, fmt.Sprint(surface.ABCICompatibility.EndBlockCleanupBounded))
		writePart(w, fmt.Sprint(surface.ABCICompatibility.RootAggregationAfterExecution))
		writePart(w, surface.ABCICompatibility.CleanupLimitPolicyID)
		writePart(w, string(surface.ABCICompatibility.RootAggregationPhase))
		writePart(w, surface.ABCICompatibility.PrecheckInputPolicyID)
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
	return hashRoot("aetra-aek-cosmos-module-requirements-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(manifest.Modules)))
		for _, surface := range manifest.Modules {
			writePart(w, surface.SurfaceHash)
		}
	})
}

func cosmosModuleSurface(moduleName CosmosSDKModuleName, path string, rootType RootType, events []string, typedErrors []string) CosmosModuleSurface {
	root, _ := NewRootContribution(rootType, string(moduleName), DeterministicEmptyRootCommitment(rootType, string(moduleName)))
	return CosmosModuleSurface{
		ModuleName:		moduleName,
		ModulePath:		path,
		MsgServer:		true,
		QueryServer:		true,
		Keeper:			true,
		Params:			true,
		GenesisExport:		true,
		GenesisImport:		true,
		Invariants:		true,
		KeeperIsolation:	DefaultKeeperIsolationPolicy(moduleName, string(moduleName)),
		IBCBoundary:		DefaultIBCReadyBoundary(),
		ABCICompatibility:	DefaultABCICompatibilityPolicy(),
		Events:			events,
		TypedErrors:		typedErrors,
		RootContribution:	root,
	}
}

func normalizeCosmosModuleSurface(surface CosmosModuleSurface) CosmosModuleSurface {
	surface.ModuleName = CosmosSDKModuleName(strings.TrimSpace(string(surface.ModuleName)))
	surface.ModulePath = strings.TrimSpace(surface.ModulePath)
	surface.Events = append([]string(nil), surface.Events...)
	surface.TypedErrors = append([]string(nil), surface.TypedErrors...)
	surface.KeeperIsolation = normalizeKeeperIsolationPolicy(surface.KeeperIsolation)
	surface.IBCBoundary = normalizeIBCReadyBoundary(surface.IBCBoundary)
	surface.ABCICompatibility = normalizeABCICompatibilityPolicy(surface.ABCICompatibility)
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

func DefaultKeeperIsolationPolicy(moduleName CosmosSDKModuleName, storeKey string) KeeperIsolationPolicy {
	storeKey = strings.TrimSpace(storeKey)
	return KeeperIsolationPolicy{
		StoreKey:				storeKey,
		ReadableStoreKeys:			[]string{storeKey},
		WritableStoreKeys:			[]string{storeKey},
		CrossZoneWritesProhibited:		true,
		CrossZoneMessagesModule:		CosmosModuleMessages,
		DirectCallsLimitedToLocalHelpers:	true,
		SharedStateReadOnlyOrProofBacked:	true,
	}
}

func (policy KeeperIsolationPolicy) Validate(moduleName CosmosSDKModuleName) error {
	policy = normalizeKeeperIsolationPolicy(policy)
	if err := validateToken("aetracore keeper isolation store key", policy.StoreKey, MaxScopeLength); err != nil {
		return err
	}
	if !IsRequiredCosmosSDKModule(moduleName) {
		return fmt.Errorf("unknown keeper isolation module %q", moduleName)
	}
	if len(policy.WritableStoreKeys) != 1 || policy.WritableStoreKeys[0] != policy.StoreKey {
		return fmt.Errorf("aetracore keeper isolation for %s must write only its own store key", moduleName)
	}
	if !containsString(policy.ReadableStoreKeys, policy.StoreKey) {
		return fmt.Errorf("aetracore keeper isolation for %s must read its own store key", moduleName)
	}
	for _, storeKey := range policy.ReadableStoreKeys {
		if err := validateToken("aetracore keeper readable store key", storeKey, MaxScopeLength); err != nil {
			return err
		}
		if storeKey != policy.StoreKey && !containsString(policy.GrantedCapabilities, "read:"+storeKey) {
			return fmt.Errorf("aetracore keeper isolation for %s reads %s without explicit capability", moduleName, storeKey)
		}
	}
	for _, storeKey := range policy.WritableStoreKeys {
		if err := validateToken("aetracore keeper writable store key", storeKey, MaxScopeLength); err != nil {
			return err
		}
	}
	if err := validateCapabilitiesForField("aetracore keeper capability", policy.GrantedCapabilities); err != nil {
		return err
	}
	if !policy.CrossZoneWritesProhibited {
		return fmt.Errorf("aetracore keeper isolation for %s must prohibit cross-zone writes", moduleName)
	}
	if policy.CrossZoneMessagesModule != CosmosModuleMessages {
		return fmt.Errorf("aetracore keeper isolation for %s must route cross-zone interactions through x/messages", moduleName)
	}
	if !policy.DirectCallsLimitedToLocalHelpers {
		return fmt.Errorf("aetracore keeper isolation for %s must limit direct calls to same-zone local helpers", moduleName)
	}
	if !policy.SharedStateReadOnlyOrProofBacked {
		return fmt.Errorf("aetracore keeper isolation for %s must keep shared state read-only or proof-backed", moduleName)
	}
	return nil
}

func DefaultIBCReadyBoundary() IBCReadyBoundary {
	return IBCReadyBoundary{
		StateExportable:		true,
		ReceiptsProofVerifiable:	true,
		CanonicalBoundaryMessages:	true,
		TimeoutRulesExplicit:		true,
		ReplayRulesExplicit:		true,
		DeterministicChannelRouting:	true,
		BoundaryMessageEncoding:	"aetra.canonical.binary.v1",
		TimeoutPolicyID:		"explicit-height-deadline-v1",
		ReplayPolicyID:			"nonce-and-tombstone-v1",
	}
}

func (boundary IBCReadyBoundary) Validate(moduleName CosmosSDKModuleName) error {
	boundary = normalizeIBCReadyBoundary(boundary)
	if !IsRequiredCosmosSDKModule(moduleName) {
		return fmt.Errorf("unknown IBC boundary module %q", moduleName)
	}
	if !boundary.StateExportable {
		return fmt.Errorf("aetracore IBC boundary for %s must export module state", moduleName)
	}
	if !boundary.ReceiptsProofVerifiable {
		return fmt.Errorf("aetracore IBC boundary for %s must make receipts proof-verifiable", moduleName)
	}
	if !boundary.CanonicalBoundaryMessages {
		return fmt.Errorf("aetracore IBC boundary for %s must use canonical packet-like messages", moduleName)
	}
	if !boundary.TimeoutRulesExplicit {
		return fmt.Errorf("aetracore IBC boundary for %s must have explicit timeout rules", moduleName)
	}
	if !boundary.ReplayRulesExplicit {
		return fmt.Errorf("aetracore IBC boundary for %s must have explicit replay rules", moduleName)
	}
	if !boundary.DeterministicChannelRouting {
		return fmt.Errorf("aetracore IBC boundary for %s must not depend on nondeterministic node routing state", moduleName)
	}
	if err := validateToken("aetracore IBC boundary encoding", boundary.BoundaryMessageEncoding, MaxScopeLength); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore IBC timeout policy id", boundary.TimeoutPolicyID); err != nil {
		return err
	}
	return validatePolicyID("aetracore IBC replay policy id", boundary.ReplayPolicyID)
}

func DefaultABCICompatibilityPolicy() ABCICompatibilityPolicy {
	return ABCICompatibilityPolicy{
		ProposalOptimizationValidityNeutral:	true,
		PrecheckDeterministic:			true,
		FinalizeBlockAuthoritative:		true,
		EndBlockCleanupBounded:			true,
		RootAggregationAfterExecution:		true,
		CleanupLimitPolicyID:			"bounded-endblock-cleanup-v1",
		RootAggregationPhase:			KernelPhaseFinalizeBlock,
		PrecheckInputPolicyID:			"consensus-context-only-v1",
	}
}

func (policy ABCICompatibilityPolicy) Validate(moduleName CosmosSDKModuleName) error {
	policy = normalizeABCICompatibilityPolicy(policy)
	if !IsRequiredCosmosSDKModule(moduleName) {
		return fmt.Errorf("unknown ABCI compatibility module %q", moduleName)
	}
	if !policy.ProposalOptimizationValidityNeutral {
		return fmt.Errorf("aetracore ABCI compatibility for %s must keep proposal optimization validity-neutral", moduleName)
	}
	if !policy.PrecheckDeterministic {
		return fmt.Errorf("aetracore ABCI compatibility for %s must keep precheck deterministic", moduleName)
	}
	if !policy.FinalizeBlockAuthoritative {
		return fmt.Errorf("aetracore ABCI compatibility for %s must keep FinalizeBlock authoritative", moduleName)
	}
	if !policy.EndBlockCleanupBounded {
		return fmt.Errorf("aetracore ABCI compatibility for %s must bound end-block cleanup", moduleName)
	}
	if !policy.RootAggregationAfterExecution {
		return fmt.Errorf("aetracore ABCI compatibility for %s must aggregate roots after deterministic execution", moduleName)
	}
	if policy.RootAggregationPhase != KernelPhaseFinalizeBlock {
		return fmt.Errorf("aetracore ABCI compatibility for %s must aggregate roots in FinalizeBlock", moduleName)
	}
	if err := validatePolicyID("aetracore ABCI cleanup limit policy id", policy.CleanupLimitPolicyID); err != nil {
		return err
	}
	return validatePolicyID("aetracore ABCI precheck input policy id", policy.PrecheckInputPolicyID)
}

func normalizeKeeperIsolationPolicy(policy KeeperIsolationPolicy) KeeperIsolationPolicy {
	policy.StoreKey = strings.TrimSpace(policy.StoreKey)
	policy.ReadableStoreKeys = append([]string(nil), policy.ReadableStoreKeys...)
	policy.WritableStoreKeys = append([]string(nil), policy.WritableStoreKeys...)
	policy.GrantedCapabilities = append([]string(nil), policy.GrantedCapabilities...)
	sort.Strings(policy.ReadableStoreKeys)
	sort.Strings(policy.WritableStoreKeys)
	sort.Strings(policy.GrantedCapabilities)
	policy.CrossZoneMessagesModule = CosmosSDKModuleName(strings.TrimSpace(string(policy.CrossZoneMessagesModule)))
	return policy
}

func normalizeIBCReadyBoundary(boundary IBCReadyBoundary) IBCReadyBoundary {
	boundary.BoundaryMessageEncoding = strings.TrimSpace(boundary.BoundaryMessageEncoding)
	boundary.TimeoutPolicyID = strings.TrimSpace(boundary.TimeoutPolicyID)
	boundary.ReplayPolicyID = strings.TrimSpace(boundary.ReplayPolicyID)
	return boundary
}

func normalizeABCICompatibilityPolicy(policy ABCICompatibilityPolicy) ABCICompatibilityPolicy {
	policy.CleanupLimitPolicyID = strings.TrimSpace(policy.CleanupLimitPolicyID)
	policy.RootAggregationPhase = KernelABCIPhase(strings.TrimSpace(string(policy.RootAggregationPhase)))
	policy.PrecheckInputPolicyID = strings.TrimSpace(policy.PrecheckInputPolicyID)
	return policy
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
