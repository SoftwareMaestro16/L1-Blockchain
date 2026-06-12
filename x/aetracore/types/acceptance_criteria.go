package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type AcceptanceCriterionID string

const (
	AcceptanceAEKCoordinatesDefaultZone		AcceptanceCriterionID	= "aek-coordinates-default-zone"
	AcceptanceFourCanonicalZonesSpecified		AcceptanceCriterionID	= "four-canonical-zones-specified"
	AcceptanceCrossZoneMessagingSemantics		AcceptanceCriterionID	= "cross-zone-messaging-semantics"
	AcceptanceServiceRegistryProofLookup		AcceptanceCriterionID	= "service-registry-proof-backed-lookup"
	AcceptanceIdentityResolverBindings		AcceptanceCriterionID	= "identity-resolver-bindings"
	AcceptanceStorageCommitmentsChunkProofs		AcceptanceCriterionID	= "storage-commitments-chunk-proofs"
	AcceptanceRoutingCommittedDeterministicTables	AcceptanceCriterionID	= "routing-committed-deterministic-tables"
	AcceptancePaymentFinancialConditionalSettlement	AcceptanceCriterionID	= "payment-financial-conditional-settlement"
	AcceptanceContractMessageDrivenIsolation	AcceptanceCriterionID	= "contract-message-driven-isolation"
	AcceptanceGlobalRootExposesAllRoots		AcceptanceCriterionID	= "global-root-exposes-all-roots"
	AcceptanceModulesExportImportInvariantsTests	AcceptanceCriterionID	= "modules-export-import-invariants-tests-typed-queries"
)

type AcceptanceCriterion struct {
	ID		AcceptanceCriterionID
	PhaseID		ImplementationRoadmapPhaseID
	Modules		[]CosmosSDKModuleName
	RootTypes	[]RootType
	PlanningReady	bool
	Evidence	[]string
	CriterionHash	string
}

type AcceptanceCriteriaManifest struct {
	Criteria	[]AcceptanceCriterion
	ManifestHash	string
}

func DefaultAcceptanceCriteriaManifest() (AcceptanceCriteriaManifest, error) {
	allModules := RequiredCosmosSDKModules()
	criteria := []AcceptanceCriterion{
		acceptanceCriterion(
			AcceptanceAEKCoordinatesDefaultZone,
			RoadmapPhaseKernelRootModel,
			[]CosmosSDKModuleName{CosmosModuleAetraCore, CosmosModuleZones},
			[]RootType{RootType("zones")},
			"AEK can register and coordinate a default execution zone",
			"default zone root contributes to global root metadata",
		),
		acceptanceCriterion(
			AcceptanceFourCanonicalZonesSpecified,
			RoadmapPhaseCanonicalZones,
			[]CosmosSDKModuleName{CosmosModuleContracts, CosmosModuleIdentity, CosmosModulePayments, CosmosModuleZones},
			[]RootType{RootType("application"), RootType("contracts"), RootType("financial"), RootType("identity")},
			"application contract financial and identity zones define state messages queries and roots",
			"canonical zone surfaces include keeper MsgServer QueryServer and message queue",
		),
		acceptanceCriterion(
			AcceptanceCrossZoneMessagingSemantics,
			RoadmapPhaseCrossZoneMessages,
			[]CosmosSDKModuleName{CosmosModuleMessages},
			[]RootType{MessageProofRootType, ReceiptProofRootType},
			"messages preserve FIFO order per sender",
			"replay protection bounce handling and receipts are mandatory",
		),
		acceptanceCriterion(
			AcceptanceServiceRegistryProofLookup,
			RoadmapPhaseServiceStorageRouting,
			[]CosmosSDKModuleName{CosmosModuleServices},
			[]RootType{RootType("services")},
			"service registry supports deterministic lookup indexes",
			"descriptor verification is proof backed",
		),
		acceptanceCriterion(
			AcceptanceIdentityResolverBindings,
			RoadmapPhaseIdentityPaymentIntegration,
			[]CosmosSDKModuleName{CosmosModuleIdentity},
			[]RootType{RootType("identity")},
			"identity zone supports account zone service contract and composite resolver outputs",
			"cross-zone identity bindings are explicit and proofable",
		),
		acceptanceCriterion(
			AcceptanceStorageCommitmentsChunkProofs,
			RoadmapPhaseServiceStorageRouting,
			[]CosmosSDKModuleName{CosmosModuleStorage},
			[]RootType{RootType("storage")},
			"storage objects are content addressed and commitment backed",
			"chunk inclusion proofs verify against object roots",
		),
		acceptanceCriterion(
			AcceptanceRoutingCommittedDeterministicTables,
			RoadmapPhaseServiceStorageRouting,
			[]CosmosSDKModuleName{CosmosModuleRouting},
			[]RootType{RoutingTableRootType},
			"routing uses committed route table epochs",
			"route selection is deterministic and proof queryable",
		),
		acceptanceCriterion(
			AcceptancePaymentFinancialConditionalSettlement,
			RoadmapPhaseIdentityPaymentIntegration,
			[]CosmosSDKModuleName{CosmosModuleMessages, CosmosModulePayments},
			[]RootType{RootType("payments"), ReceiptProofRootType},
			"payment layer supports conditional transfers",
			"final settlement is routed through the Financial Zone",
		),
		acceptanceCriterion(
			AcceptanceContractMessageDrivenIsolation,
			RoadmapPhaseVMRuntime,
			[]CosmosSDKModuleName{CosmosModuleContracts, CosmosModuleMessages},
			[]RootType{RootType("contracts"), MessageProofRootType, ReceiptProofRootType},
			"contract execution emits cross-zone messages instead of direct writes",
			"contract zone storage remains isolated and proof verifiable",
		),
		acceptanceCriterion(
			AcceptanceGlobalRootExposesAllRoots,
			RoadmapPhaseKernelRootModel,
			[]CosmosSDKModuleName{CosmosModuleAetraCore},
			[]RootType{RootType("contracts"), RootType("identity"), MessageProofRootType, RootType("payments"), ReceiptProofRootType, RoutingTableRootType, RootType("services"), RootType("storage"), RootType("zones")},
			"global root exposes zones services identity storage messages receipts routing payments and contracts",
			"unified root set is commitment backed and queryable",
		),
		acceptanceCriterion(
			AcceptanceModulesExportImportInvariantsTests,
			RoadmapPhasePerformanceHardening,
			allModules,
			[]RootType{RootType("aetracore"), RootType("contracts"), RootType("identity"), MessageProofRootType, RootType("payments"), RoutingTableRootType, RootType("services"), RootType("storage"), RootType("zones")},
			"all modules have export import invariants and tests",
			"all modules expose typed query interfaces",
		),
	}
	return NewAcceptanceCriteriaManifest(criteria)
}

func NewAcceptanceCriteriaManifest(criteria []AcceptanceCriterion) (AcceptanceCriteriaManifest, error) {
	manifest := AcceptanceCriteriaManifest{Criteria: normalizeAcceptanceCriteria(criteria)}
	if err := manifest.ValidateFormat(); err != nil {
		return AcceptanceCriteriaManifest{}, err
	}
	for i := range manifest.Criteria {
		manifest.Criteria[i].CriterionHash = ComputeAcceptanceCriterionHash(manifest.Criteria[i])
	}
	manifest.ManifestHash = ComputeAcceptanceCriteriaManifestHash(manifest)
	return manifest, manifest.Validate()
}

func RequiredAcceptanceCriterionIDs() []AcceptanceCriterionID {
	return []AcceptanceCriterionID{
		AcceptanceAEKCoordinatesDefaultZone,
		AcceptanceFourCanonicalZonesSpecified,
		AcceptanceCrossZoneMessagingSemantics,
		AcceptanceServiceRegistryProofLookup,
		AcceptanceIdentityResolverBindings,
		AcceptanceStorageCommitmentsChunkProofs,
		AcceptanceRoutingCommittedDeterministicTables,
		AcceptancePaymentFinancialConditionalSettlement,
		AcceptanceContractMessageDrivenIsolation,
		AcceptanceGlobalRootExposesAllRoots,
		AcceptanceModulesExportImportInvariantsTests,
	}
}

func (manifest AcceptanceCriteriaManifest) ValidateFormat() error {
	manifest.Criteria = normalizeAcceptanceCriteria(manifest.Criteria)
	required := requiredAcceptanceCriterionIDStrings()
	if len(manifest.Criteria) != len(required) {
		return fmt.Errorf("aetracore acceptance criteria manifest must include %d required criteria", len(required))
	}
	seen := make(map[AcceptanceCriterionID]struct{}, len(manifest.Criteria))
	var previous AcceptanceCriterionID
	for i, criterion := range manifest.Criteria {
		if err := criterion.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[criterion.ID]; found {
			return fmt.Errorf("duplicate aetracore acceptance criterion %s", criterion.ID)
		}
		seen[criterion.ID] = struct{}{}
		if i > 0 && previous >= criterion.ID {
			return errors.New("aetracore acceptance criteria must be sorted canonically")
		}
		previous = criterion.ID
	}
	for _, id := range required {
		if _, found := seen[AcceptanceCriterionID(id)]; !found {
			return fmt.Errorf("missing aetracore acceptance criterion %s", id)
		}
	}
	if manifest.ManifestHash != "" {
		return ValidateHash("aetracore acceptance criteria manifest hash", manifest.ManifestHash)
	}
	return nil
}

func (manifest AcceptanceCriteriaManifest) Validate() error {
	manifest.Criteria = normalizeAcceptanceCriteria(manifest.Criteria)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, criterion := range manifest.Criteria {
		if err := criterion.Validate(); err != nil {
			return err
		}
	}
	if manifest.ManifestHash != ComputeAcceptanceCriteriaManifestHash(manifest) {
		return errors.New("aetracore acceptance criteria manifest hash mismatch")
	}
	return nil
}

func (criterion AcceptanceCriterion) ValidateFormat() error {
	criterion = normalizeAcceptanceCriterion(criterion)
	if !IsRequiredAcceptanceCriterionID(criterion.ID) {
		return fmt.Errorf("unknown aetracore acceptance criterion %q", criterion.ID)
	}
	if !IsImplementationRoadmapPhaseID(criterion.PhaseID) {
		return fmt.Errorf("aetracore acceptance criterion %s references unknown roadmap phase %s", criterion.ID, criterion.PhaseID)
	}
	if len(criterion.Modules) == 0 {
		return fmt.Errorf("aetracore acceptance criterion %s must name at least one module", criterion.ID)
	}
	var previousModule CosmosSDKModuleName
	seenModules := make(map[CosmosSDKModuleName]struct{}, len(criterion.Modules))
	for i, moduleName := range criterion.Modules {
		if !IsRequiredCosmosSDKModule(moduleName) {
			return fmt.Errorf("aetracore acceptance criterion %s references unknown module %s", criterion.ID, moduleName)
		}
		if _, found := seenModules[moduleName]; found {
			return fmt.Errorf("duplicate aetracore acceptance criterion module %s", moduleName)
		}
		seenModules[moduleName] = struct{}{}
		if i > 0 && previousModule >= moduleName {
			return fmt.Errorf("aetracore acceptance criterion %s modules must be sorted canonically", criterion.ID)
		}
		previousModule = moduleName
	}
	if len(criterion.RootTypes) == 0 {
		return fmt.Errorf("aetracore acceptance criterion %s must name at least one root type", criterion.ID)
	}
	var previousRoot RootType
	seenRoots := make(map[RootType]struct{}, len(criterion.RootTypes))
	for i, rootType := range criterion.RootTypes {
		if err := validatePolicyID("aetracore acceptance criterion root type", string(rootType)); err != nil {
			return err
		}
		if _, found := seenRoots[rootType]; found {
			return fmt.Errorf("duplicate aetracore acceptance criterion root type %s", rootType)
		}
		seenRoots[rootType] = struct{}{}
		if i > 0 && previousRoot >= rootType {
			return fmt.Errorf("aetracore acceptance criterion %s root types must be sorted canonically", criterion.ID)
		}
		previousRoot = rootType
	}
	if !criterion.PlanningReady {
		return fmt.Errorf("aetracore acceptance criterion %s must be planning ready", criterion.ID)
	}
	if err := validateAcceptanceEvidence(criterion.ID, criterion.Evidence); err != nil {
		return err
	}
	if criterion.CriterionHash != "" {
		return ValidateHash("aetracore acceptance criterion hash", criterion.CriterionHash)
	}
	return nil
}

func (criterion AcceptanceCriterion) Validate() error {
	criterion = normalizeAcceptanceCriterion(criterion)
	if err := criterion.ValidateFormat(); err != nil {
		return err
	}
	if criterion.CriterionHash != ComputeAcceptanceCriterionHash(criterion) {
		return errors.New("aetracore acceptance criterion hash mismatch")
	}
	return nil
}

func AcceptanceCriterionByID(manifest AcceptanceCriteriaManifest, id AcceptanceCriterionID) (AcceptanceCriterion, bool) {
	for _, criterion := range manifest.Criteria {
		if criterion.ID == id {
			return criterion, true
		}
	}
	return AcceptanceCriterion{}, false
}

func IsRequiredAcceptanceCriterionID(id AcceptanceCriterionID) bool {
	for _, required := range RequiredAcceptanceCriterionIDs() {
		if required == id {
			return true
		}
	}
	return false
}

func ComputeAcceptanceCriterionHash(criterion AcceptanceCriterion) string {
	criterion = normalizeAcceptanceCriterion(criterion)
	return hashRoot("aetra-aek-acceptance-criterion-v1", func(w byteWriter) {
		writePart(w, string(criterion.ID))
		writePart(w, string(criterion.PhaseID))
		writeUint64(w, uint64(len(criterion.Modules)))
		for _, moduleName := range criterion.Modules {
			writePart(w, string(moduleName))
		}
		writeUint64(w, uint64(len(criterion.RootTypes)))
		for _, rootType := range criterion.RootTypes {
			writePart(w, string(rootType))
		}
		writePart(w, fmt.Sprint(criterion.PlanningReady))
		writeStringParts(w, criterion.Evidence)
	})
}

func ComputeAcceptanceCriteriaManifestHash(manifest AcceptanceCriteriaManifest) string {
	manifest.Criteria = normalizeAcceptanceCriteria(manifest.Criteria)
	return hashRoot("aetra-aek-acceptance-criteria-manifest-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(manifest.Criteria)))
		for _, criterion := range manifest.Criteria {
			writePart(w, criterion.CriterionHash)
		}
	})
}

func acceptanceCriterion(id AcceptanceCriterionID, phaseID ImplementationRoadmapPhaseID, modules []CosmosSDKModuleName, rootTypes []RootType, evidence ...string) AcceptanceCriterion {
	return AcceptanceCriterion{
		ID:		id,
		PhaseID:	phaseID,
		Modules:	modules,
		RootTypes:	rootTypes,
		PlanningReady:	true,
		Evidence:	evidence,
	}
}

func normalizeAcceptanceCriterion(criterion AcceptanceCriterion) AcceptanceCriterion {
	criterion.ID = AcceptanceCriterionID(strings.TrimSpace(string(criterion.ID)))
	criterion.PhaseID = ImplementationRoadmapPhaseID(strings.TrimSpace(string(criterion.PhaseID)))
	criterion.Modules = append([]CosmosSDKModuleName(nil), criterion.Modules...)
	for i, moduleName := range criterion.Modules {
		criterion.Modules[i] = CosmosSDKModuleName(strings.TrimSpace(string(moduleName)))
	}
	sort.SliceStable(criterion.Modules, func(i, j int) bool {
		return criterion.Modules[i] < criterion.Modules[j]
	})
	criterion.RootTypes = append([]RootType(nil), criterion.RootTypes...)
	for i, rootType := range criterion.RootTypes {
		criterion.RootTypes[i] = RootType(strings.TrimSpace(string(rootType)))
	}
	sort.SliceStable(criterion.RootTypes, func(i, j int) bool {
		return criterion.RootTypes[i] < criterion.RootTypes[j]
	})
	criterion.Evidence = normalizeRoadmapStringSet(criterion.Evidence)
	criterion.CriterionHash = strings.ToLower(strings.TrimSpace(criterion.CriterionHash))
	return criterion
}

func normalizeAcceptanceCriteria(criteria []AcceptanceCriterion) []AcceptanceCriterion {
	out := make([]AcceptanceCriterion, len(criteria))
	for i, criterion := range criteria {
		criterion = normalizeAcceptanceCriterion(criterion)
		if criterion.CriterionHash == "" {
			criterion.CriterionHash = ComputeAcceptanceCriterionHash(criterion)
		}
		out[i] = criterion
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func validateAcceptanceEvidence(id AcceptanceCriterionID, evidence []string) error {
	if len(evidence) == 0 {
		return fmt.Errorf("aetracore acceptance criterion %s evidence is required", id)
	}
	seen := make(map[string]struct{}, len(evidence))
	var previous string
	for i, item := range evidence {
		if err := validateRoadmapText("aetracore acceptance criterion evidence", item); err != nil {
			return err
		}
		if _, found := seen[item]; found {
			return fmt.Errorf("duplicate aetracore acceptance criterion evidence %s", item)
		}
		seen[item] = struct{}{}
		if i > 0 && previous >= item {
			return fmt.Errorf("aetracore acceptance criterion %s evidence must be sorted canonically", id)
		}
		previous = item
	}
	return nil
}

func requiredAcceptanceCriterionIDStrings() []string {
	ids := RequiredAcceptanceCriterionIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}
