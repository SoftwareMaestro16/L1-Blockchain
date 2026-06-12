package types

import (
	"errors"
	"fmt"
	"sort"
)

type ServiceImplementationAcceptanceID string
type ServiceImplementationAcceptanceCategory string
type ServiceImplementationAcceptanceEvidence string

const (
	ServiceAcceptanceFirstClassRegistryObjects	ServiceImplementationAcceptanceID	= "services_first_class_registry_objects"
	ServiceAcceptanceAllServiceTypesDefined		ServiceImplementationAcceptanceID	= "all_service_types_defined"
	ServiceAcceptanceFormalInterfaceBinding		ServiceImplementationAcceptanceID	= "formal_interface_hash_binding"
	ServiceAcceptanceUnifiedCallEnvelope		ServiceImplementationAcceptanceID	= "unified_call_envelope"
	ServiceAcceptanceDiscoveryResolution		ServiceImplementationAcceptanceID	= "discovery_resolution_sources"
	ServiceAcceptancePaymentSettlementModes		ServiceImplementationAcceptanceID	= "payment_settlement_modes"
	ServiceAcceptanceStorageDeclarations		ServiceImplementationAcceptanceID	= "storage_declaration_models"
	ServiceAcceptanceMixedServiceDisputes		ServiceImplementationAcceptanceID	= "mixed_service_challenge_or_fallback"
	ServiceAcceptanceProviderRules			ServiceImplementationAcceptanceID	= "provider_collateral_and_reputation_rules"
	ServiceAcceptanceCosmosModuleSurface		ServiceImplementationAcceptanceID	= "cosmos_module_surface"
	ServiceAcceptanceStoreV2BlockSTMStrategy	ServiceImplementationAcceptanceID	= "storev2_blockstm_strategy"
	ServiceAcceptanceSDKExecutionFlow		ServiceImplementationAcceptanceID	= "sdk_execution_flow"

	ServiceAcceptanceCategoryRegistry	ServiceImplementationAcceptanceCategory	= "registry"
	ServiceAcceptanceCategoryInterface	ServiceImplementationAcceptanceCategory	= "interface"
	ServiceAcceptanceCategoryCall		ServiceImplementationAcceptanceCategory	= "call"
	ServiceAcceptanceCategoryDiscovery	ServiceImplementationAcceptanceCategory	= "discovery"
	ServiceAcceptanceCategoryPayment	ServiceImplementationAcceptanceCategory	= "payment"
	ServiceAcceptanceCategoryStorage	ServiceImplementationAcceptanceCategory	= "storage"
	ServiceAcceptanceCategoryMixed		ServiceImplementationAcceptanceCategory	= "mixed"
	ServiceAcceptanceCategoryProvider	ServiceImplementationAcceptanceCategory	= "provider"
	ServiceAcceptanceCategoryCosmos		ServiceImplementationAcceptanceCategory	= "cosmos"
	ServiceAcceptanceCategoryPerformance	ServiceImplementationAcceptanceCategory	= "performance"
	ServiceAcceptanceCategorySDK		ServiceImplementationAcceptanceCategory	= "sdk"

	ServiceAcceptanceEvidenceType		ServiceImplementationAcceptanceEvidence	= "type_contract"
	ServiceAcceptanceEvidenceValidation	ServiceImplementationAcceptanceEvidence	= "validation"
	ServiceAcceptanceEvidenceHash		ServiceImplementationAcceptanceEvidence	= "hash_commitment"
	ServiceAcceptanceEvidenceTest		ServiceImplementationAcceptanceEvidence	= "test_coverage"
	ServiceAcceptanceEvidenceIntegration	ServiceImplementationAcceptanceEvidence	= "integration_contract"
)

type ServiceImplementationAcceptanceCriterion struct {
	CriterionID		ServiceImplementationAcceptanceID
	Category		ServiceImplementationAcceptanceCategory
	RequiredEvidence	[]ServiceImplementationAcceptanceEvidence
	RequiredObjects		[]string
	CriterionHash		string
}

type ServiceImplementationAcceptanceManifest struct {
	Criteria	[]ServiceImplementationAcceptanceCriterion
	ManifestHash	string
}

type ServiceImplementationPlanningGate struct {
	SatisfiedCriteria []ServiceImplementationAcceptanceID
}

func DefaultServiceImplementationAcceptanceManifest() (ServiceImplementationAcceptanceManifest, error) {
	return NewServiceImplementationAcceptanceManifest([]ServiceImplementationAcceptanceCriterion{
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceFirstClassRegistryObjects,
			ServiceAcceptanceCategoryRegistry,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceHash, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"ServiceDescriptor", "ServiceAnchor", "ServiceRegistryState", "ServiceRegistryProof"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceAllServiceTypesDefined,
			ServiceAcceptanceCategoryRegistry,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"on_chain", "off_chain", "mixed", "fog_market"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceFormalInterfaceBinding,
			ServiceAcceptanceCategoryInterface,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceHash, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"ServiceInterface", "ServiceMethod", "interface_hash", "method_id"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceUnifiedCallEnvelope,
			ServiceAcceptanceCategoryCall,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceHash, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"ServiceCall", "PaymentEnvelope", "payload", "proof_requirement", "timeout", "signature", "idempotency_key"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceDiscoveryResolution,
			ServiceAcceptanceCategoryDiscovery,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"registry", "signed_cache", "identity_binding", "mesh_record"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptancePaymentSettlementModes,
			ServiceAcceptanceCategoryPayment,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"on_chain", "streaming", "prepaid", "metered", "escrow"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceStorageDeclarations,
			ServiceAcceptanceCategoryStorage,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest},
			[]string{"ephemeral", "persistent_on_chain", "distributed_off_chain", "hybrid"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceMixedServiceDisputes,
			ServiceAcceptanceCategoryMixed,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest, ServiceAcceptanceEvidenceIntegration},
			[]string{"challenge_window", "fallback_rule", "dispute_message", "settlement_state"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceProviderRules,
			ServiceAcceptanceCategoryProvider,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceTest, ServiceAcceptanceEvidenceIntegration},
			[]string{"ProviderRecord", "ProviderCollateral", "ReputationRecord", "AvailabilityCommitment"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceCosmosModuleSurface,
			ServiceAcceptanceCategoryCosmos,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceIntegration, ServiceAcceptanceEvidenceTest},
			[]string{"MsgServer", "QueryServer", "Keeper", "Genesis", "Invariants", "Events", "TypedErrors", "RootCommitments"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceStoreV2BlockSTMStrategy,
			ServiceAcceptanceCategoryPerformance,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceIntegration, ServiceAcceptanceEvidenceTest},
			[]string{"StoreV2Layout", "BlockSTMConflictStrategy", "PrimaryLookup", "PartitionedWrites"},
		),
		newServiceImplementationAcceptanceCriterion(
			ServiceAcceptanceSDKExecutionFlow,
			ServiceAcceptanceCategorySDK,
			[]ServiceImplementationAcceptanceEvidence{ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceIntegration, ServiceAcceptanceEvidenceTest},
			[]string{"ResolveService", "VerifyInterface", "BuildCall", "AttachPayment", "ExecuteCall", "VerifyReceipt"},
		),
	})
}

func NewServiceImplementationAcceptanceManifest(criteria []ServiceImplementationAcceptanceCriterion) (ServiceImplementationAcceptanceManifest, error) {
	manifest := ServiceImplementationAcceptanceManifest{
		Criteria: canonicalServiceImplementationAcceptanceCriteria(criteria),
	}
	if err := manifest.ValidateFormat(); err != nil {
		return ServiceImplementationAcceptanceManifest{}, err
	}
	for i := range manifest.Criteria {
		manifest.Criteria[i].CriterionHash = ComputeServiceImplementationAcceptanceCriterionHash(manifest.Criteria[i])
	}
	manifest.ManifestHash = ComputeServiceImplementationAcceptanceManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (manifest ServiceImplementationAcceptanceManifest) ValidateFormat() error {
	manifest.Criteria = canonicalServiceImplementationAcceptanceCriteria(manifest.Criteria)
	if len(manifest.Criteria) != len(requiredServiceImplementationAcceptanceIDs()) {
		return fmt.Errorf("services implementation acceptance manifest must include %d criteria", len(requiredServiceImplementationAcceptanceIDs()))
	}
	seen := map[ServiceImplementationAcceptanceID]struct{}{}
	for _, criterion := range manifest.Criteria {
		if err := criterion.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[criterion.CriterionID]; found {
			return fmt.Errorf("duplicate services implementation acceptance criterion %q", criterion.CriterionID)
		}
		seen[criterion.CriterionID] = struct{}{}
	}
	for _, criterionID := range requiredServiceImplementationAcceptanceIDs() {
		if _, found := seen[criterionID]; !found {
			return fmt.Errorf("missing services implementation acceptance criterion %q", criterionID)
		}
	}
	return nil
}

func (manifest ServiceImplementationAcceptanceManifest) Validate() error {
	manifest.Criteria = canonicalServiceImplementationAcceptanceCriteria(manifest.Criteria)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, criterion := range manifest.Criteria {
		if criterion.CriterionHash == "" {
			return fmt.Errorf("services implementation acceptance criterion %q hash is required", criterion.CriterionID)
		}
		if expected := ComputeServiceImplementationAcceptanceCriterionHash(criterion); criterion.CriterionHash != expected {
			return fmt.Errorf("services implementation acceptance criterion %q hash mismatch: expected %s", criterion.CriterionID, expected)
		}
	}
	if manifest.ManifestHash == "" {
		return errors.New("services implementation acceptance manifest hash is required")
	}
	if expected := ComputeServiceImplementationAcceptanceManifestHash(manifest); manifest.ManifestHash != expected {
		return fmt.Errorf("services implementation acceptance manifest hash mismatch: expected %s", expected)
	}
	return nil
}

func (criterion ServiceImplementationAcceptanceCriterion) ValidateFormat() error {
	criterion = canonicalServiceImplementationAcceptanceCriterion(criterion)
	if !IsServiceImplementationAcceptanceID(criterion.CriterionID) {
		return fmt.Errorf("unknown services implementation acceptance criterion %q", criterion.CriterionID)
	}
	if !IsServiceImplementationAcceptanceCategory(criterion.Category) {
		return fmt.Errorf("unknown services implementation acceptance category %q", criterion.Category)
	}
	if err := validateSortedServiceImplementationAcceptanceEvidence("services implementation acceptance evidence", criterion.RequiredEvidence); err != nil {
		return err
	}
	if err := validateSortedConstraintTokens("services implementation acceptance object", criterion.RequiredObjects); err != nil {
		return err
	}
	return nil
}

func (gate ServiceImplementationPlanningGate) ValidateReady(manifest ServiceImplementationAcceptanceManifest) error {
	if err := manifest.Validate(); err != nil {
		return err
	}
	satisfied := map[ServiceImplementationAcceptanceID]struct{}{}
	for _, criterionID := range gate.SatisfiedCriteria {
		if !IsServiceImplementationAcceptanceID(criterionID) {
			return fmt.Errorf("unknown services implementation acceptance gate criterion %q", criterionID)
		}
		if _, found := satisfied[criterionID]; found {
			return fmt.Errorf("duplicate services implementation acceptance gate criterion %q", criterionID)
		}
		satisfied[criterionID] = struct{}{}
	}
	for _, criterion := range manifest.Criteria {
		if _, found := satisfied[criterion.CriterionID]; !found {
			return fmt.Errorf("services implementation planning gate missing criterion %q", criterion.CriterionID)
		}
	}
	return nil
}

func ReadyServiceImplementationPlanningGate() ServiceImplementationPlanningGate {
	return ServiceImplementationPlanningGate{
		SatisfiedCriteria: requiredServiceImplementationAcceptanceIDs(),
	}
}

func ComputeServiceImplementationAcceptanceCriterionHash(criterion ServiceImplementationAcceptanceCriterion) string {
	criterion = canonicalServiceImplementationAcceptanceCriterion(criterion)
	parts := []string{
		"aetra-services-implementation-acceptance-criterion-v1",
		string(criterion.CriterionID),
		string(criterion.Category),
		"evidence",
		fmt.Sprint(len(criterion.RequiredEvidence)),
	}
	for _, evidence := range criterion.RequiredEvidence {
		parts = append(parts, string(evidence))
	}
	parts = append(parts, "objects", fmt.Sprint(len(criterion.RequiredObjects)))
	parts = append(parts, criterion.RequiredObjects...)
	return servicesHashParts(parts...)
}

func ComputeServiceImplementationAcceptanceManifestHash(manifest ServiceImplementationAcceptanceManifest) string {
	manifest.Criteria = canonicalServiceImplementationAcceptanceCriteria(manifest.Criteria)
	parts := []string{
		"aetra-services-implementation-acceptance-manifest-v1",
		fmt.Sprint(len(manifest.Criteria)),
	}
	for _, criterion := range manifest.Criteria {
		parts = append(parts, string(criterion.CriterionID), ComputeServiceImplementationAcceptanceCriterionHash(criterion))
	}
	return servicesHashParts(parts...)
}

func IsServiceImplementationAcceptanceID(criterionID ServiceImplementationAcceptanceID) bool {
	for _, required := range requiredServiceImplementationAcceptanceIDs() {
		if criterionID == required {
			return true
		}
	}
	return false
}

func IsServiceImplementationAcceptanceCategory(category ServiceImplementationAcceptanceCategory) bool {
	switch category {
	case ServiceAcceptanceCategoryRegistry, ServiceAcceptanceCategoryInterface, ServiceAcceptanceCategoryCall,
		ServiceAcceptanceCategoryDiscovery, ServiceAcceptanceCategoryPayment, ServiceAcceptanceCategoryStorage,
		ServiceAcceptanceCategoryMixed, ServiceAcceptanceCategoryProvider, ServiceAcceptanceCategoryCosmos,
		ServiceAcceptanceCategoryPerformance, ServiceAcceptanceCategorySDK:
		return true
	default:
		return false
	}
}

func IsServiceImplementationAcceptanceEvidence(evidence ServiceImplementationAcceptanceEvidence) bool {
	switch evidence {
	case ServiceAcceptanceEvidenceType, ServiceAcceptanceEvidenceValidation, ServiceAcceptanceEvidenceHash,
		ServiceAcceptanceEvidenceTest, ServiceAcceptanceEvidenceIntegration:
		return true
	default:
		return false
	}
}

func newServiceImplementationAcceptanceCriterion(criterionID ServiceImplementationAcceptanceID, category ServiceImplementationAcceptanceCategory, evidence []ServiceImplementationAcceptanceEvidence, objects []string) ServiceImplementationAcceptanceCriterion {
	return ServiceImplementationAcceptanceCriterion{
		CriterionID:		criterionID,
		Category:		category,
		RequiredEvidence:	evidence,
		RequiredObjects:	objects,
	}
}

func canonicalServiceImplementationAcceptanceCriterion(criterion ServiceImplementationAcceptanceCriterion) ServiceImplementationAcceptanceCriterion {
	criterion.RequiredEvidence = append([]ServiceImplementationAcceptanceEvidence(nil), criterion.RequiredEvidence...)
	criterion.RequiredObjects = append([]string(nil), criterion.RequiredObjects...)
	sort.SliceStable(criterion.RequiredEvidence, func(i, j int) bool {
		return criterion.RequiredEvidence[i] < criterion.RequiredEvidence[j]
	})
	sort.Strings(criterion.RequiredObjects)
	return criterion
}

func canonicalServiceImplementationAcceptanceCriteria(criteria []ServiceImplementationAcceptanceCriterion) []ServiceImplementationAcceptanceCriterion {
	canonical := append([]ServiceImplementationAcceptanceCriterion(nil), criteria...)
	for i := range canonical {
		canonical[i] = canonicalServiceImplementationAcceptanceCriterion(canonical[i])
	}
	sort.SliceStable(canonical, func(i, j int) bool {
		return canonical[i].CriterionID < canonical[j].CriterionID
	})
	return canonical
}

func validateSortedServiceImplementationAcceptanceEvidence(fieldName string, values []ServiceImplementationAcceptanceEvidence) error {
	if len(values) == 0 {
		return fmt.Errorf("%s list must not be empty", fieldName)
	}
	previous := ServiceImplementationAcceptanceEvidence("")
	for _, value := range values {
		if !IsServiceImplementationAcceptanceEvidence(value) {
			return fmt.Errorf("unknown %s %q", fieldName, value)
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("%s list must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func requiredServiceImplementationAcceptanceIDs() []ServiceImplementationAcceptanceID {
	return []ServiceImplementationAcceptanceID{
		ServiceAcceptanceAllServiceTypesDefined,
		ServiceAcceptanceCosmosModuleSurface,
		ServiceAcceptanceDiscoveryResolution,
		ServiceAcceptanceFirstClassRegistryObjects,
		ServiceAcceptanceFormalInterfaceBinding,
		ServiceAcceptanceMixedServiceDisputes,
		ServiceAcceptancePaymentSettlementModes,
		ServiceAcceptanceProviderRules,
		ServiceAcceptanceSDKExecutionFlow,
		ServiceAcceptanceStorageDeclarations,
		ServiceAcceptanceStoreV2BlockSTMStrategy,
		ServiceAcceptanceUnifiedCallEnvelope,
	}
}
