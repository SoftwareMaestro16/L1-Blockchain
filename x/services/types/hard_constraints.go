package types

import (
	"errors"
	"fmt"
	"sort"
)

type ServiceHardConstraintID string
type ServiceHardConstraintCategory string
type ServiceHardConstraintStage string

const (
	ServiceConstraintNoCentralizedBackendAssumption		ServiceHardConstraintID	= "no_centralized_service_backend_assumptions"
	ServiceConstraintNoMessagingApplicationDependency	ServiceHardConstraintID	= "no_messaging_application_dependency"
	ServiceConstraintNoMonolithicExecutionEngine		ServiceHardConstraintID	= "no_monolithic_service_execution_engine"
	ServiceConstraintNoManualABIForRegisteredServices	ServiceHardConstraintID	= "no_manual_abi_integrations_for_registered_services"
	ServiceConstraintNoExternalAPIInConsensusExecution	ServiceHardConstraintID	= "no_external_api_reliance_in_consensus_execution"
	ServiceConstraintNoNondeterministicStateTransition	ServiceHardConstraintID	= "no_nondeterministic_state_transitions"
	ServiceConstraintNoUnboundedRegistryScans		ServiceHardConstraintID	= "no_unbounded_registry_scans"
	ServiceConstraintNoUnmeteredProofVerification		ServiceHardConstraintID	= "no_unmetered_proof_verification"
	ServiceConstraintNoUnverifiedOffChainCanonicalState	ServiceHardConstraintID	= "no_unverified_offchain_result_canonicalization"

	ServiceConstraintCategoryArchitecture	ServiceHardConstraintCategory	= "architecture"
	ServiceConstraintCategoryIntegration	ServiceHardConstraintCategory	= "integration"
	ServiceConstraintCategoryConsensus	ServiceHardConstraintCategory	= "consensus"
	ServiceConstraintCategoryPerformance	ServiceHardConstraintCategory	= "performance"
	ServiceConstraintCategorySecurity	ServiceHardConstraintCategory	= "security"

	ServiceConstraintStageDesign		ServiceHardConstraintStage	= "design"
	ServiceConstraintStageClientSDK		ServiceHardConstraintStage	= "client_sdk"
	ServiceConstraintStageAnteHandler	ServiceHardConstraintStage	= "ante_handler"
	ServiceConstraintStageProcessProposal	ServiceHardConstraintStage	= "process_proposal"
	ServiceConstraintStageFinalizeBlock	ServiceHardConstraintStage	= "finalize_block"
	ServiceConstraintStageKeeper		ServiceHardConstraintStage	= "keeper"
	ServiceConstraintStageQuery		ServiceHardConstraintStage	= "query"
	ServiceConstraintStageDispute		ServiceHardConstraintStage	= "dispute"
)

type ServiceHardConstraint struct {
	ConstraintID		ServiceHardConstraintID
	Category		ServiceHardConstraintCategory
	ConsensusCritical	bool
	Stages			[]ServiceHardConstraintStage
	RequiredControls	[]string
	ConstraintHash		string
}

type ServiceHardConstraintsManifest struct {
	Constraints	[]ServiceHardConstraint
	ManifestHash	string
}

type ServiceExecutionHardConstraintPolicy struct {
	CentralizedBackendRequired		bool
	MessagingApplicationRequired		bool
	MonolithicExecutionEngine		bool
	ManualABIRequiredForRegisteredService	bool
	ExternalAPIRequiredInConsensus		bool
	NondeterministicStateTransition		bool
	UnboundedRegistryScan			bool
	UnmeteredProofVerification		bool
	OffChainResultCanonical			bool
	OffChainResultHasProof			bool
	OffChainResultHasSignature		bool
	OffChainResultHasChallengeWindow	bool
	OffChainResultHasExplicitTrustModel	bool
}

func DefaultServiceHardConstraintsManifest() (ServiceHardConstraintsManifest, error) {
	return NewServiceHardConstraintsManifest([]ServiceHardConstraint{
		newServiceHardConstraint(
			ServiceConstraintNoCentralizedBackendAssumption,
			ServiceConstraintCategoryArchitecture,
			false,
			[]ServiceHardConstraintStage{ServiceConstraintStageDesign, ServiceConstraintStageClientSDK, ServiceConstraintStageQuery},
			[]string{"distributed_discovery", "signed_advertisements", "registry_anchor"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoMessagingApplicationDependency,
			ServiceConstraintCategoryIntegration,
			false,
			[]ServiceHardConstraintStage{ServiceConstraintStageDesign, ServiceConstraintStageClientSDK},
			[]string{"transport_agnostic_calls", "service_network_abstraction"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoMonolithicExecutionEngine,
			ServiceConstraintCategoryArchitecture,
			false,
			[]ServiceHardConstraintStage{ServiceConstraintStageDesign, ServiceConstraintStageKeeper},
			[]string{"module_separated_stf", "service_target_routing"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoManualABIForRegisteredServices,
			ServiceConstraintCategoryIntegration,
			false,
			[]ServiceHardConstraintStage{ServiceConstraintStageClientSDK, ServiceConstraintStageQuery},
			[]string{"interface_hash_verification", "method_schema_registry"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoExternalAPIInConsensusExecution,
			ServiceConstraintCategoryConsensus,
			true,
			[]ServiceHardConstraintStage{ServiceConstraintStageAnteHandler, ServiceConstraintStageProcessProposal, ServiceConstraintStageFinalizeBlock, ServiceConstraintStageKeeper},
			[]string{"deterministic_inputs_only", "receipt_anchor", "stf_network_ban"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoNondeterministicStateTransition,
			ServiceConstraintCategoryConsensus,
			true,
			[]ServiceHardConstraintStage{ServiceConstraintStageAnteHandler, ServiceConstraintStageProcessProposal, ServiceConstraintStageFinalizeBlock, ServiceConstraintStageKeeper},
			[]string{"consensus_context_height", "canonical_encoding", "deterministic_ordering"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoUnboundedRegistryScans,
			ServiceConstraintCategoryPerformance,
			true,
			[]ServiceHardConstraintStage{ServiceConstraintStageKeeper, ServiceConstraintStageQuery, ServiceConstraintStageFinalizeBlock},
			[]string{"primary_service_id_lookup", "prefix_bounded_iteration", "expiry_batch_limit"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoUnmeteredProofVerification,
			ServiceConstraintCategorySecurity,
			true,
			[]ServiceHardConstraintStage{ServiceConstraintStageAnteHandler, ServiceConstraintStageFinalizeBlock, ServiceConstraintStageDispute},
			[]string{"gas_metered_proofs", "proof_size_limit", "verification_budget"},
		),
		newServiceHardConstraint(
			ServiceConstraintNoUnverifiedOffChainCanonicalState,
			ServiceConstraintCategorySecurity,
			true,
			[]ServiceHardConstraintStage{ServiceConstraintStageAnteHandler, ServiceConstraintStageFinalizeBlock, ServiceConstraintStageDispute},
			[]string{"challenge_window", "explicit_trust_model", "result_signature", "verification_proof"},
		),
	})
}

func NewServiceHardConstraintsManifest(constraints []ServiceHardConstraint) (ServiceHardConstraintsManifest, error) {
	manifest := ServiceHardConstraintsManifest{
		Constraints: canonicalServiceHardConstraints(constraints),
	}
	if err := manifest.ValidateFormat(); err != nil {
		return ServiceHardConstraintsManifest{}, err
	}
	for i := range manifest.Constraints {
		manifest.Constraints[i].ConstraintHash = ComputeServiceHardConstraintHash(manifest.Constraints[i])
	}
	manifest.ManifestHash = ComputeServiceHardConstraintsManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (manifest ServiceHardConstraintsManifest) ValidateFormat() error {
	manifest.Constraints = canonicalServiceHardConstraints(manifest.Constraints)
	if len(manifest.Constraints) != len(requiredServiceHardConstraintIDs()) {
		return fmt.Errorf("services hard constraints manifest must include %d constraints", len(requiredServiceHardConstraintIDs()))
	}
	seen := map[ServiceHardConstraintID]struct{}{}
	for _, constraint := range manifest.Constraints {
		if err := constraint.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[constraint.ConstraintID]; found {
			return fmt.Errorf("duplicate services hard constraint %q", constraint.ConstraintID)
		}
		seen[constraint.ConstraintID] = struct{}{}
	}
	for _, constraintID := range requiredServiceHardConstraintIDs() {
		if _, found := seen[constraintID]; !found {
			return fmt.Errorf("missing services hard constraint %q", constraintID)
		}
	}
	return nil
}

func (manifest ServiceHardConstraintsManifest) Validate() error {
	manifest.Constraints = canonicalServiceHardConstraints(manifest.Constraints)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, constraint := range manifest.Constraints {
		if constraint.ConstraintHash == "" {
			return fmt.Errorf("services hard constraint %q hash is required", constraint.ConstraintID)
		}
		if expected := ComputeServiceHardConstraintHash(constraint); constraint.ConstraintHash != expected {
			return fmt.Errorf("services hard constraint %q hash mismatch: expected %s", constraint.ConstraintID, expected)
		}
	}
	if manifest.ManifestHash == "" {
		return errors.New("services hard constraints manifest hash is required")
	}
	if expected := ComputeServiceHardConstraintsManifestHash(manifest); manifest.ManifestHash != expected {
		return fmt.Errorf("services hard constraints manifest hash mismatch: expected %s", expected)
	}
	return nil
}

func (constraint ServiceHardConstraint) ValidateFormat() error {
	constraint = canonicalServiceHardConstraint(constraint)
	if !IsServiceHardConstraintID(constraint.ConstraintID) {
		return fmt.Errorf("unknown services hard constraint %q", constraint.ConstraintID)
	}
	if !IsServiceHardConstraintCategory(constraint.Category) {
		return fmt.Errorf("unknown services hard constraint category %q", constraint.Category)
	}
	if len(constraint.Stages) == 0 {
		return fmt.Errorf("services hard constraint %q must declare enforcement stages", constraint.ConstraintID)
	}
	for _, stage := range constraint.Stages {
		if !IsServiceHardConstraintStage(stage) {
			return fmt.Errorf("unknown services hard constraint stage %q", stage)
		}
	}
	if err := validateSortedServiceHardConstraintStages("services hard constraint stage", constraint.Stages); err != nil {
		return err
	}
	if err := validateSortedConstraintTokens("services hard constraint control", constraint.RequiredControls); err != nil {
		return err
	}
	return nil
}

func (policy ServiceExecutionHardConstraintPolicy) ValidateAgainstHardConstraints() error {
	switch {
	case policy.CentralizedBackendRequired:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoCentralizedBackendAssumption)
	case policy.MessagingApplicationRequired:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoMessagingApplicationDependency)
	case policy.MonolithicExecutionEngine:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoMonolithicExecutionEngine)
	case policy.ManualABIRequiredForRegisteredService:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoManualABIForRegisteredServices)
	case policy.ExternalAPIRequiredInConsensus:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoExternalAPIInConsensusExecution)
	case policy.NondeterministicStateTransition:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoNondeterministicStateTransition)
	case policy.UnboundedRegistryScan:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoUnboundedRegistryScans)
	case policy.UnmeteredProofVerification:
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoUnmeteredProofVerification)
	case policy.OffChainResultCanonical && !policy.HasOffChainCanonicalizationGate():
		return fmt.Errorf("violates services hard constraint %q", ServiceConstraintNoUnverifiedOffChainCanonicalState)
	default:
		return nil
	}
}

func (policy ServiceExecutionHardConstraintPolicy) HasOffChainCanonicalizationGate() bool {
	return policy.OffChainResultHasProof ||
		policy.OffChainResultHasSignature ||
		policy.OffChainResultHasChallengeWindow ||
		policy.OffChainResultHasExplicitTrustModel
}

func ComputeServiceHardConstraintHash(constraint ServiceHardConstraint) string {
	constraint = canonicalServiceHardConstraint(constraint)
	parts := []string{
		"aetra-services-hard-constraint-v1",
		string(constraint.ConstraintID),
		string(constraint.Category),
		fmt.Sprint(constraint.ConsensusCritical),
		"stages",
		fmt.Sprint(len(constraint.Stages)),
	}
	for _, stage := range constraint.Stages {
		parts = append(parts, string(stage))
	}
	parts = append(parts, "controls", fmt.Sprint(len(constraint.RequiredControls)))
	parts = append(parts, constraint.RequiredControls...)
	return servicesHashParts(parts...)
}

func ComputeServiceHardConstraintsManifestHash(manifest ServiceHardConstraintsManifest) string {
	manifest.Constraints = canonicalServiceHardConstraints(manifest.Constraints)
	parts := []string{
		"aetra-services-hard-constraints-manifest-v1",
		fmt.Sprint(len(manifest.Constraints)),
	}
	for _, constraint := range manifest.Constraints {
		parts = append(parts, string(constraint.ConstraintID), ComputeServiceHardConstraintHash(constraint))
	}
	return servicesHashParts(parts...)
}

func IsServiceHardConstraintID(constraintID ServiceHardConstraintID) bool {
	for _, required := range requiredServiceHardConstraintIDs() {
		if constraintID == required {
			return true
		}
	}
	return false
}

func IsServiceHardConstraintCategory(category ServiceHardConstraintCategory) bool {
	switch category {
	case ServiceConstraintCategoryArchitecture, ServiceConstraintCategoryIntegration, ServiceConstraintCategoryConsensus,
		ServiceConstraintCategoryPerformance, ServiceConstraintCategorySecurity:
		return true
	default:
		return false
	}
}

func IsServiceHardConstraintStage(stage ServiceHardConstraintStage) bool {
	switch stage {
	case ServiceConstraintStageDesign, ServiceConstraintStageClientSDK, ServiceConstraintStageAnteHandler,
		ServiceConstraintStageProcessProposal, ServiceConstraintStageFinalizeBlock, ServiceConstraintStageKeeper,
		ServiceConstraintStageQuery, ServiceConstraintStageDispute:
		return true
	default:
		return false
	}
}

func newServiceHardConstraint(constraintID ServiceHardConstraintID, category ServiceHardConstraintCategory, consensusCritical bool, stages []ServiceHardConstraintStage, controls []string) ServiceHardConstraint {
	return ServiceHardConstraint{
		ConstraintID:		constraintID,
		Category:		category,
		ConsensusCritical:	consensusCritical,
		Stages:			stages,
		RequiredControls:	controls,
	}
}

func canonicalServiceHardConstraint(constraint ServiceHardConstraint) ServiceHardConstraint {
	constraint.Stages = append([]ServiceHardConstraintStage(nil), constraint.Stages...)
	constraint.RequiredControls = append([]string(nil), constraint.RequiredControls...)
	sort.SliceStable(constraint.Stages, func(i, j int) bool {
		return constraint.Stages[i] < constraint.Stages[j]
	})
	sort.Strings(constraint.RequiredControls)
	return constraint
}

func canonicalServiceHardConstraints(constraints []ServiceHardConstraint) []ServiceHardConstraint {
	canonical := append([]ServiceHardConstraint(nil), constraints...)
	for i := range canonical {
		canonical[i] = canonicalServiceHardConstraint(canonical[i])
	}
	sort.SliceStable(canonical, func(i, j int) bool {
		return canonical[i].ConstraintID < canonical[j].ConstraintID
	})
	return canonical
}

func validateSortedServiceHardConstraintStages(fieldName string, stages []ServiceHardConstraintStage) error {
	previous := ServiceHardConstraintStage("")
	for _, stage := range stages {
		if previous != "" && previous >= stage {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
		previous = stage
	}
	return nil
}

func validateSortedConstraintTokens(fieldName string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s list must not be empty", fieldName)
	}
	previous := ""
	for _, value := range values {
		if err := validateInterfaceToken(fieldName, value); err != nil {
			return err
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("%s list must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func requiredServiceHardConstraintIDs() []ServiceHardConstraintID {
	return []ServiceHardConstraintID{
		ServiceConstraintNoCentralizedBackendAssumption,
		ServiceConstraintNoExternalAPIInConsensusExecution,
		ServiceConstraintNoManualABIForRegisteredServices,
		ServiceConstraintNoMessagingApplicationDependency,
		ServiceConstraintNoMonolithicExecutionEngine,
		ServiceConstraintNoNondeterministicStateTransition,
		ServiceConstraintNoUnboundedRegistryScans,
		ServiceConstraintNoUnmeteredProofVerification,
		ServiceConstraintNoUnverifiedOffChainCanonicalState,
	}
}
