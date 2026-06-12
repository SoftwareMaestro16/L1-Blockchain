package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceDescriptorObjectKind string
type ServiceSchemaCompatibilityMode string
type ServiceInterfaceLifecycleStatus string
type ServiceProviderReregistrationMode string

const (
	ServiceDescriptorObjectCanonicalDescriptor	ServiceDescriptorObjectKind	= "canonical_descriptor"
	ServiceDescriptorObjectDistributedRecord	ServiceDescriptorObjectKind	= "distributed_record"
	ServiceDescriptorObjectDistributedEndpoint	ServiceDescriptorObjectKind	= "distributed_endpoint"
	ServiceDescriptorObjectInterfaceDescriptor	ServiceDescriptorObjectKind	= "interface_descriptor"

	ServiceSchemaCompatibilityBackward	ServiceSchemaCompatibilityMode	= "backward_compatible"
	ServiceSchemaCompatibilityAdditive	ServiceSchemaCompatibilityMode	= "additive"
	ServiceSchemaCompatibilityBreaking	ServiceSchemaCompatibilityMode	= "breaking"

	ServiceInterfaceLifecycleActive		ServiceInterfaceLifecycleStatus	= "active"
	ServiceInterfaceLifecycleDeprecated	ServiceInterfaceLifecycleStatus	= "deprecated"
	ServiceInterfaceLifecycleRetired	ServiceInterfaceLifecycleStatus	= "retired"

	ServiceProviderReregistrationUnchanged		ServiceProviderReregistrationMode	= "unchanged"
	ServiceProviderReregistrationOwnerAuthorized	ServiceProviderReregistrationMode	= "owner_authorized"
	ServiceProviderReregistrationCollateralRefresh	ServiceProviderReregistrationMode	= "collateral_refresh"
	ServiceProviderReregistrationCapabilityRefresh	ServiceProviderReregistrationMode	= "capability_refresh"
	ServiceProviderReregistrationFullReregistration	ServiceProviderReregistrationMode	= "full_reregistration"
	ServiceUpgradeGovernanceProposalPrefix							= "gov/services/upgrade/"
)

type ServiceDescriptorVersionRecord struct {
	ObjectKind		ServiceDescriptorObjectKind
	ObjectID		string
	Version			uint64
	DescriptorHash		string
	InterfaceHash		string
	SchemaCompatibilityHash	string
	VersionHash		string
}

type ServiceSchemaCompatibilityMetadata struct {
	SchemaID		string
	PreviousHash		string
	NextHash		string
	Mode			ServiceSchemaCompatibilityMode
	AddedFields		[]string
	DeprecatedFields	[]string
	RemovedFields		[]string
	RequiredFields		[]string
	CompatibilityHash	string
}

type ServiceRegistryMigrationHandler struct {
	FromRegistryVersion		uint64
	ToRegistryVersion		uint64
	HandlerName			string
	GovernanceProposalID		string
	DescriptorMigrationHandler	string
	InterfaceMigrationHandler	string
	ProviderMigrationHandler	string
	HandlerHash			string
}

type ServiceInterfaceDeprecationMarker struct {
	InterfaceHash			string
	Version				uint64
	DeprecatedHeight		uint64
	RetirementHeight		uint64
	ReplacementInterfaceHash	string
	Reason				string
	MarkerHash			string
}

type ServiceProviderReregistrationRule struct {
	ProviderID			string
	ServiceID			string
	PreviousInterfaceHash		string
	NextInterfaceHash		string
	Mode				ServiceProviderReregistrationMode
	RequiresOwnerAuthorization	bool
	RequiresCollateralRefresh	bool
	RequiresCapabilityRefresh	bool
	EarliestHeight			uint64
	RuleHash			string
}

type ServiceRegistryUpgradePlan struct {
	FromRegistryVersion	uint64
	ToRegistryVersion	uint64
	GovernanceProposalID	string
	DescriptorVersions	[]ServiceDescriptorVersionRecord
	SchemaCompatibility	[]ServiceSchemaCompatibilityMetadata
	MigrationHandlers	[]ServiceRegistryMigrationHandler
	InterfaceDeprecations	[]ServiceInterfaceDeprecationMarker
	ProviderReregistration	[]ServiceProviderReregistrationRule
	PlanHash		string
}

type ServiceRegistryUpgradeSimulationResult struct {
	PlanHash				string
	FromRegistryVersion			uint64
	ToRegistryVersion			uint64
	DescriptorVersionsMigrated		uint32
	BackwardCompatibleSchemas		uint32
	BreakingSchemas				uint32
	MigrationHandlersApplied		uint32
	InterfaceDeprecationsApplied		uint32
	ProvidersRequiringReregistration	uint32
	GovernanceControlled			bool
	BackwardCompatible			bool
	SimulationHash				string
}

func NewServiceDescriptorVersionRecord(record ServiceDescriptorVersionRecord) (ServiceDescriptorVersionRecord, error) {
	if record.VersionHash != "" {
		return ServiceDescriptorVersionRecord{}, errors.New("services descriptor version hash must be empty before construction")
	}
	record = canonicalServiceDescriptorVersionRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return ServiceDescriptorVersionRecord{}, err
	}
	record.VersionHash = ComputeServiceDescriptorVersionRecordHash(record)
	return record, record.Validate()
}

func NewServiceSchemaCompatibilityMetadata(metadata ServiceSchemaCompatibilityMetadata) (ServiceSchemaCompatibilityMetadata, error) {
	if metadata.CompatibilityHash != "" {
		return ServiceSchemaCompatibilityMetadata{}, errors.New("services schema compatibility hash must be empty before construction")
	}
	metadata = canonicalServiceSchemaCompatibilityMetadata(metadata)
	if err := metadata.ValidateFormat(); err != nil {
		return ServiceSchemaCompatibilityMetadata{}, err
	}
	metadata.CompatibilityHash = ComputeServiceSchemaCompatibilityHash(metadata)
	return metadata, metadata.Validate()
}

func NewServiceRegistryMigrationHandler(handler ServiceRegistryMigrationHandler) (ServiceRegistryMigrationHandler, error) {
	if handler.HandlerHash != "" {
		return ServiceRegistryMigrationHandler{}, errors.New("services registry migration handler hash must be empty before construction")
	}
	handler = canonicalServiceRegistryMigrationHandler(handler)
	if err := handler.ValidateFormat(); err != nil {
		return ServiceRegistryMigrationHandler{}, err
	}
	handler.HandlerHash = ComputeServiceRegistryMigrationHandlerHash(handler)
	return handler, handler.Validate()
}

func NewServiceInterfaceDeprecationMarker(marker ServiceInterfaceDeprecationMarker) (ServiceInterfaceDeprecationMarker, error) {
	if marker.MarkerHash != "" {
		return ServiceInterfaceDeprecationMarker{}, errors.New("services interface deprecation marker hash must be empty before construction")
	}
	marker = canonicalServiceInterfaceDeprecationMarker(marker)
	if err := marker.ValidateFormat(); err != nil {
		return ServiceInterfaceDeprecationMarker{}, err
	}
	marker.MarkerHash = ComputeServiceInterfaceDeprecationMarkerHash(marker)
	return marker, marker.Validate()
}

func NewServiceProviderReregistrationRule(rule ServiceProviderReregistrationRule) (ServiceProviderReregistrationRule, error) {
	if rule.RuleHash != "" {
		return ServiceProviderReregistrationRule{}, errors.New("services provider reregistration rule hash must be empty before construction")
	}
	rule = canonicalServiceProviderReregistrationRule(rule)
	if err := rule.ValidateFormat(); err != nil {
		return ServiceProviderReregistrationRule{}, err
	}
	rule.RuleHash = ComputeServiceProviderReregistrationRuleHash(rule)
	return rule, rule.Validate()
}

func NewServiceRegistryUpgradePlan(plan ServiceRegistryUpgradePlan) (ServiceRegistryUpgradePlan, error) {
	if plan.PlanHash != "" {
		return ServiceRegistryUpgradePlan{}, errors.New("services registry upgrade plan hash must be empty before construction")
	}
	plan = canonicalServiceRegistryUpgradePlan(plan)
	if err := plan.ValidateFormat(); err != nil {
		return ServiceRegistryUpgradePlan{}, err
	}
	plan.PlanHash = ComputeServiceRegistryUpgradePlanHash(plan)
	return plan, plan.Validate()
}

func SimulateServiceRegistryUpgrade(plan ServiceRegistryUpgradePlan) (ServiceRegistryUpgradeSimulationResult, error) {
	plan = canonicalServiceRegistryUpgradePlan(plan)
	if err := plan.Validate(); err != nil {
		return ServiceRegistryUpgradeSimulationResult{}, err
	}
	result := ServiceRegistryUpgradeSimulationResult{
		PlanHash:				plan.PlanHash,
		FromRegistryVersion:			plan.FromRegistryVersion,
		ToRegistryVersion:			plan.ToRegistryVersion,
		DescriptorVersionsMigrated:		uint32(len(plan.DescriptorVersions)),
		MigrationHandlersApplied:		uint32(len(plan.MigrationHandlers)),
		InterfaceDeprecationsApplied:		uint32(len(plan.InterfaceDeprecations)),
		ProvidersRequiringReregistration:	uint32(countProviderReregistrations(plan.ProviderReregistration)),
		GovernanceControlled:			strings.HasPrefix(plan.GovernanceProposalID, ServiceUpgradeGovernanceProposalPrefix),
		BackwardCompatible:			true,
	}
	for _, metadata := range plan.SchemaCompatibility {
		if metadata.Mode == ServiceSchemaCompatibilityBreaking {
			result.BreakingSchemas++
			result.BackwardCompatible = false
			continue
		}
		result.BackwardCompatibleSchemas++
	}
	if result.ProvidersRequiringReregistration > 0 {
		result.BackwardCompatible = false
	}
	if !result.GovernanceControlled || result.MigrationHandlersApplied == 0 {
		result.BackwardCompatible = false
	}
	result.SimulationHash = ComputeServiceRegistryUpgradeSimulationHash(result)
	return result, result.Validate()
}

func (record ServiceDescriptorVersionRecord) ValidateFormat() error {
	if !IsServiceDescriptorObjectKind(record.ObjectKind) {
		return fmt.Errorf("services descriptor version unknown object kind %q", record.ObjectKind)
	}
	if err := validateInterfaceToken("services descriptor version object id", record.ObjectID); err != nil {
		return err
	}
	if record.Version == 0 {
		return errors.New("services descriptor object version must be positive")
	}
	if err := coretypes.ValidateHash("services descriptor object descriptor hash", record.DescriptorHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services descriptor object interface hash", record.InterfaceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services descriptor object schema compatibility hash", record.SchemaCompatibilityHash); err != nil {
		return err
	}
	if record.VersionHash != "" {
		return coretypes.ValidateHash("services descriptor object version hash", record.VersionHash)
	}
	return nil
}

func (record ServiceDescriptorVersionRecord) Validate() error {
	record = canonicalServiceDescriptorVersionRecord(record)
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if record.VersionHash == "" {
		return errors.New("services descriptor version hash is required")
	}
	if record.VersionHash != ComputeServiceDescriptorVersionRecordHash(record) {
		return errors.New("services descriptor version hash mismatch")
	}
	return nil
}

func (metadata ServiceSchemaCompatibilityMetadata) ValidateFormat() error {
	if err := validateInterfaceToken("services schema compatibility id", metadata.SchemaID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services schema compatibility previous hash", metadata.PreviousHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services schema compatibility next hash", metadata.NextHash); err != nil {
		return err
	}
	if metadata.PreviousHash == metadata.NextHash {
		return errors.New("services schema compatibility requires distinct schema hashes")
	}
	if !IsServiceSchemaCompatibilityMode(metadata.Mode) {
		return fmt.Errorf("services schema compatibility unknown mode %q", metadata.Mode)
	}
	if err := validateSortedTokens("services schema compatibility added field", metadata.AddedFields); err != nil {
		return err
	}
	if err := validateSortedTokens("services schema compatibility deprecated field", metadata.DeprecatedFields); err != nil {
		return err
	}
	if err := validateSortedTokens("services schema compatibility removed field", metadata.RemovedFields); err != nil {
		return err
	}
	if err := validateSortedTokens("services schema compatibility required field", metadata.RequiredFields); err != nil {
		return err
	}
	if metadata.Mode != ServiceSchemaCompatibilityBreaking && len(metadata.RemovedFields) != 0 {
		return errors.New("services backward-compatible schema metadata cannot remove fields")
	}
	if metadata.Mode == ServiceSchemaCompatibilityAdditive && len(metadata.AddedFields) == 0 {
		return errors.New("services additive schema metadata requires added fields")
	}
	if metadata.CompatibilityHash != "" {
		return coretypes.ValidateHash("services schema compatibility hash", metadata.CompatibilityHash)
	}
	return nil
}

func (metadata ServiceSchemaCompatibilityMetadata) Validate() error {
	metadata = canonicalServiceSchemaCompatibilityMetadata(metadata)
	if err := metadata.ValidateFormat(); err != nil {
		return err
	}
	if metadata.CompatibilityHash == "" {
		return errors.New("services schema compatibility hash is required")
	}
	if metadata.CompatibilityHash != ComputeServiceSchemaCompatibilityHash(metadata) {
		return errors.New("services schema compatibility hash mismatch")
	}
	return nil
}

func (handler ServiceRegistryMigrationHandler) ValidateFormat() error {
	if handler.FromRegistryVersion == 0 || handler.ToRegistryVersion == 0 {
		return errors.New("services registry migration versions must be positive")
	}
	if handler.ToRegistryVersion <= handler.FromRegistryVersion {
		return errors.New("services registry migration must advance version")
	}
	if err := validateInterfaceToken("services registry migration handler name", handler.HandlerName); err != nil {
		return err
	}
	if !strings.HasPrefix(handler.GovernanceProposalID, ServiceUpgradeGovernanceProposalPrefix) {
		return errors.New("services registry migration requires governance proposal")
	}
	if err := validateInterfaceToken("services registry migration governance proposal", handler.GovernanceProposalID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services registry descriptor migration handler", handler.DescriptorMigrationHandler); err != nil {
		return err
	}
	if err := validateInterfaceToken("services registry interface migration handler", handler.InterfaceMigrationHandler); err != nil {
		return err
	}
	if err := validateInterfaceToken("services registry provider migration handler", handler.ProviderMigrationHandler); err != nil {
		return err
	}
	if handler.HandlerHash != "" {
		return coretypes.ValidateHash("services registry migration handler hash", handler.HandlerHash)
	}
	return nil
}

func (handler ServiceRegistryMigrationHandler) Validate() error {
	handler = canonicalServiceRegistryMigrationHandler(handler)
	if err := handler.ValidateFormat(); err != nil {
		return err
	}
	if handler.HandlerHash == "" {
		return errors.New("services registry migration handler hash is required")
	}
	if handler.HandlerHash != ComputeServiceRegistryMigrationHandlerHash(handler) {
		return errors.New("services registry migration handler hash mismatch")
	}
	return nil
}

func (marker ServiceInterfaceDeprecationMarker) ValidateFormat() error {
	if err := coretypes.ValidateHash("services deprecated interface hash", marker.InterfaceHash); err != nil {
		return err
	}
	if marker.Version == 0 {
		return errors.New("services deprecated interface version must be positive")
	}
	if marker.DeprecatedHeight == 0 {
		return errors.New("services interface deprecation height must be positive")
	}
	if marker.RetirementHeight <= marker.DeprecatedHeight {
		return errors.New("services interface retirement height must be after deprecation height")
	}
	if marker.ReplacementInterfaceHash != "" {
		if err := coretypes.ValidateHash("services replacement interface hash", marker.ReplacementInterfaceHash); err != nil {
			return err
		}
	}
	if err := validateInterfaceToken("services interface deprecation reason", marker.Reason); err != nil {
		return err
	}
	if marker.MarkerHash != "" {
		return coretypes.ValidateHash("services interface deprecation marker hash", marker.MarkerHash)
	}
	return nil
}

func (marker ServiceInterfaceDeprecationMarker) Validate() error {
	marker = canonicalServiceInterfaceDeprecationMarker(marker)
	if err := marker.ValidateFormat(); err != nil {
		return err
	}
	if marker.MarkerHash == "" {
		return errors.New("services interface deprecation marker hash is required")
	}
	if marker.MarkerHash != ComputeServiceInterfaceDeprecationMarkerHash(marker) {
		return errors.New("services interface deprecation marker hash mismatch")
	}
	return nil
}

func (rule ServiceProviderReregistrationRule) ValidateFormat() error {
	if err := validateInterfaceToken("services provider reregistration provider id", rule.ProviderID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services provider reregistration service id", rule.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services provider previous interface hash", rule.PreviousInterfaceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services provider next interface hash", rule.NextInterfaceHash); err != nil {
		return err
	}
	if !IsServiceProviderReregistrationMode(rule.Mode) {
		return fmt.Errorf("services provider unknown reregistration mode %q", rule.Mode)
	}
	if rule.PreviousInterfaceHash != rule.NextInterfaceHash && rule.Mode == ServiceProviderReregistrationUnchanged {
		return errors.New("services provider interface change requires re-registration mode")
	}
	if rule.Mode == ServiceProviderReregistrationFullReregistration {
		if !rule.RequiresOwnerAuthorization || !rule.RequiresCollateralRefresh || !rule.RequiresCapabilityRefresh {
			return errors.New("services provider full re-registration requires owner, collateral, and capability refresh")
		}
	}
	if rule.Mode == ServiceProviderReregistrationOwnerAuthorized && !rule.RequiresOwnerAuthorization {
		return errors.New("services provider owner-authorized re-registration must require owner authorization")
	}
	if rule.EarliestHeight == 0 {
		return errors.New("services provider re-registration earliest height must be positive")
	}
	if rule.RuleHash != "" {
		return coretypes.ValidateHash("services provider reregistration rule hash", rule.RuleHash)
	}
	return nil
}

func (rule ServiceProviderReregistrationRule) Validate() error {
	rule = canonicalServiceProviderReregistrationRule(rule)
	if err := rule.ValidateFormat(); err != nil {
		return err
	}
	if rule.RuleHash == "" {
		return errors.New("services provider reregistration rule hash is required")
	}
	if rule.RuleHash != ComputeServiceProviderReregistrationRuleHash(rule) {
		return errors.New("services provider reregistration rule hash mismatch")
	}
	return nil
}

func (plan ServiceRegistryUpgradePlan) ValidateFormat() error {
	if plan.FromRegistryVersion == 0 || plan.ToRegistryVersion == 0 {
		return errors.New("services registry upgrade versions must be positive")
	}
	if plan.ToRegistryVersion <= plan.FromRegistryVersion {
		return errors.New("services registry upgrade must advance version")
	}
	if !strings.HasPrefix(plan.GovernanceProposalID, ServiceUpgradeGovernanceProposalPrefix) {
		return errors.New("services registry upgrade requires governance proposal")
	}
	if err := validateInterfaceToken("services registry upgrade governance proposal", plan.GovernanceProposalID); err != nil {
		return err
	}
	if len(plan.DescriptorVersions) == 0 {
		return errors.New("services registry upgrade requires descriptor version records")
	}
	for _, record := range plan.DescriptorVersions {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	if len(plan.SchemaCompatibility) == 0 {
		return errors.New("services registry upgrade requires schema compatibility metadata")
	}
	for _, metadata := range plan.SchemaCompatibility {
		if err := metadata.Validate(); err != nil {
			return err
		}
	}
	if err := validateMigrationHandlerChain(plan.FromRegistryVersion, plan.ToRegistryVersion, plan.MigrationHandlers); err != nil {
		return err
	}
	for _, marker := range plan.InterfaceDeprecations {
		if err := marker.Validate(); err != nil {
			return err
		}
	}
	for _, rule := range plan.ProviderReregistration {
		if err := rule.Validate(); err != nil {
			return err
		}
	}
	if plan.PlanHash != "" {
		return coretypes.ValidateHash("services registry upgrade plan hash", plan.PlanHash)
	}
	return nil
}

func (plan ServiceRegistryUpgradePlan) Validate() error {
	plan = canonicalServiceRegistryUpgradePlan(plan)
	if err := plan.ValidateFormat(); err != nil {
		return err
	}
	if plan.PlanHash == "" {
		return errors.New("services registry upgrade plan hash is required")
	}
	if plan.PlanHash != ComputeServiceRegistryUpgradePlanHash(plan) {
		return errors.New("services registry upgrade plan hash mismatch")
	}
	return nil
}

func (result ServiceRegistryUpgradeSimulationResult) Validate() error {
	if err := coretypes.ValidateHash("services registry upgrade simulation plan hash", result.PlanHash); err != nil {
		return err
	}
	if result.FromRegistryVersion == 0 || result.ToRegistryVersion <= result.FromRegistryVersion {
		return errors.New("services registry upgrade simulation versions are invalid")
	}
	if !result.GovernanceControlled {
		return errors.New("services registry upgrade simulation must be governance controlled")
	}
	if result.MigrationHandlersApplied == 0 {
		return errors.New("services registry upgrade simulation requires applied migration handlers")
	}
	if err := coretypes.ValidateHash("services registry upgrade simulation hash", result.SimulationHash); err != nil {
		return err
	}
	if result.SimulationHash != ComputeServiceRegistryUpgradeSimulationHash(result) {
		return errors.New("services registry upgrade simulation hash mismatch")
	}
	return nil
}

func InterfaceLifecycleStatusAt(marker ServiceInterfaceDeprecationMarker, height uint64) (ServiceInterfaceLifecycleStatus, error) {
	if err := marker.Validate(); err != nil {
		return "", err
	}
	if height < marker.DeprecatedHeight {
		return ServiceInterfaceLifecycleActive, nil
	}
	if height < marker.RetirementHeight {
		return ServiceInterfaceLifecycleDeprecated, nil
	}
	return ServiceInterfaceLifecycleRetired, nil
}

func IsServiceDescriptorObjectKind(kind ServiceDescriptorObjectKind) bool {
	switch kind {
	case ServiceDescriptorObjectCanonicalDescriptor, ServiceDescriptorObjectDistributedRecord, ServiceDescriptorObjectDistributedEndpoint, ServiceDescriptorObjectInterfaceDescriptor:
		return true
	default:
		return false
	}
}

func IsServiceSchemaCompatibilityMode(mode ServiceSchemaCompatibilityMode) bool {
	switch mode {
	case ServiceSchemaCompatibilityBackward, ServiceSchemaCompatibilityAdditive, ServiceSchemaCompatibilityBreaking:
		return true
	default:
		return false
	}
}

func IsServiceInterfaceLifecycleStatus(status ServiceInterfaceLifecycleStatus) bool {
	switch status {
	case ServiceInterfaceLifecycleActive, ServiceInterfaceLifecycleDeprecated, ServiceInterfaceLifecycleRetired:
		return true
	default:
		return false
	}
}

func IsServiceProviderReregistrationMode(mode ServiceProviderReregistrationMode) bool {
	switch mode {
	case ServiceProviderReregistrationUnchanged, ServiceProviderReregistrationOwnerAuthorized, ServiceProviderReregistrationCollateralRefresh, ServiceProviderReregistrationCapabilityRefresh, ServiceProviderReregistrationFullReregistration:
		return true
	default:
		return false
	}
}

func ComputeServiceDescriptorVersionRecordHash(record ServiceDescriptorVersionRecord) string {
	record = canonicalServiceDescriptorVersionRecord(record)
	return servicesHashParts(
		"aetra-services-descriptor-version-record-v1",
		string(record.ObjectKind),
		record.ObjectID,
		fmt.Sprint(record.Version),
		record.DescriptorHash,
		record.InterfaceHash,
		record.SchemaCompatibilityHash,
	)
}

func ComputeServiceSchemaCompatibilityHash(metadata ServiceSchemaCompatibilityMetadata) string {
	metadata = canonicalServiceSchemaCompatibilityMetadata(metadata)
	parts := []string{
		"aetra-services-schema-compatibility-v1",
		metadata.SchemaID,
		metadata.PreviousHash,
		metadata.NextHash,
		string(metadata.Mode),
		"added",
		fmt.Sprint(len(metadata.AddedFields)),
	}
	parts = append(parts, metadata.AddedFields...)
	parts = append(parts, "deprecated", fmt.Sprint(len(metadata.DeprecatedFields)))
	parts = append(parts, metadata.DeprecatedFields...)
	parts = append(parts, "removed", fmt.Sprint(len(metadata.RemovedFields)))
	parts = append(parts, metadata.RemovedFields...)
	parts = append(parts, "required", fmt.Sprint(len(metadata.RequiredFields)))
	parts = append(parts, metadata.RequiredFields...)
	return servicesHashParts(parts...)
}

func ComputeServiceRegistryMigrationHandlerHash(handler ServiceRegistryMigrationHandler) string {
	handler = canonicalServiceRegistryMigrationHandler(handler)
	return servicesHashParts(
		"aetra-services-registry-migration-handler-v1",
		fmt.Sprint(handler.FromRegistryVersion),
		fmt.Sprint(handler.ToRegistryVersion),
		handler.HandlerName,
		handler.GovernanceProposalID,
		handler.DescriptorMigrationHandler,
		handler.InterfaceMigrationHandler,
		handler.ProviderMigrationHandler,
	)
}

func ComputeServiceInterfaceDeprecationMarkerHash(marker ServiceInterfaceDeprecationMarker) string {
	marker = canonicalServiceInterfaceDeprecationMarker(marker)
	return servicesHashParts(
		"aetra-services-interface-deprecation-marker-v1",
		marker.InterfaceHash,
		fmt.Sprint(marker.Version),
		fmt.Sprint(marker.DeprecatedHeight),
		fmt.Sprint(marker.RetirementHeight),
		marker.ReplacementInterfaceHash,
		marker.Reason,
	)
}

func ComputeServiceProviderReregistrationRuleHash(rule ServiceProviderReregistrationRule) string {
	rule = canonicalServiceProviderReregistrationRule(rule)
	return servicesHashParts(
		"aetra-services-provider-reregistration-rule-v1",
		rule.ProviderID,
		rule.ServiceID,
		rule.PreviousInterfaceHash,
		rule.NextInterfaceHash,
		string(rule.Mode),
		fmt.Sprint(rule.RequiresOwnerAuthorization),
		fmt.Sprint(rule.RequiresCollateralRefresh),
		fmt.Sprint(rule.RequiresCapabilityRefresh),
		fmt.Sprint(rule.EarliestHeight),
	)
}

func ComputeServiceRegistryUpgradePlanHash(plan ServiceRegistryUpgradePlan) string {
	plan = canonicalServiceRegistryUpgradePlan(plan)
	parts := []string{
		"aetra-services-registry-upgrade-plan-v1",
		fmt.Sprint(plan.FromRegistryVersion),
		fmt.Sprint(plan.ToRegistryVersion),
		plan.GovernanceProposalID,
		"descriptors",
		fmt.Sprint(len(plan.DescriptorVersions)),
	}
	for _, record := range plan.DescriptorVersions {
		parts = append(parts, record.VersionHash)
	}
	parts = append(parts, "compatibility", fmt.Sprint(len(plan.SchemaCompatibility)))
	for _, metadata := range plan.SchemaCompatibility {
		parts = append(parts, metadata.CompatibilityHash)
	}
	parts = append(parts, "migrations", fmt.Sprint(len(plan.MigrationHandlers)))
	for _, handler := range plan.MigrationHandlers {
		parts = append(parts, handler.HandlerHash)
	}
	parts = append(parts, "deprecations", fmt.Sprint(len(plan.InterfaceDeprecations)))
	for _, marker := range plan.InterfaceDeprecations {
		parts = append(parts, marker.MarkerHash)
	}
	parts = append(parts, "providers", fmt.Sprint(len(plan.ProviderReregistration)))
	for _, rule := range plan.ProviderReregistration {
		parts = append(parts, rule.RuleHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRegistryUpgradeSimulationHash(result ServiceRegistryUpgradeSimulationResult) string {
	return servicesHashParts(
		"aetra-services-registry-upgrade-simulation-v1",
		result.PlanHash,
		fmt.Sprint(result.FromRegistryVersion),
		fmt.Sprint(result.ToRegistryVersion),
		fmt.Sprint(result.DescriptorVersionsMigrated),
		fmt.Sprint(result.BackwardCompatibleSchemas),
		fmt.Sprint(result.BreakingSchemas),
		fmt.Sprint(result.MigrationHandlersApplied),
		fmt.Sprint(result.InterfaceDeprecationsApplied),
		fmt.Sprint(result.ProvidersRequiringReregistration),
		fmt.Sprint(result.GovernanceControlled),
		fmt.Sprint(result.BackwardCompatible),
	)
}

func canonicalServiceDescriptorVersionRecord(record ServiceDescriptorVersionRecord) ServiceDescriptorVersionRecord {
	record.ObjectID = strings.TrimSpace(record.ObjectID)
	record.VersionHash = strings.TrimSpace(record.VersionHash)
	return record
}

func canonicalServiceSchemaCompatibilityMetadata(metadata ServiceSchemaCompatibilityMetadata) ServiceSchemaCompatibilityMetadata {
	metadata.SchemaID = strings.TrimSpace(metadata.SchemaID)
	metadata.AddedFields = sortedStrings(metadata.AddedFields)
	metadata.DeprecatedFields = sortedStrings(metadata.DeprecatedFields)
	metadata.RemovedFields = sortedStrings(metadata.RemovedFields)
	metadata.RequiredFields = sortedStrings(metadata.RequiredFields)
	metadata.CompatibilityHash = strings.TrimSpace(metadata.CompatibilityHash)
	return metadata
}

func canonicalServiceRegistryMigrationHandler(handler ServiceRegistryMigrationHandler) ServiceRegistryMigrationHandler {
	handler.HandlerName = strings.TrimSpace(handler.HandlerName)
	handler.GovernanceProposalID = strings.TrimSpace(handler.GovernanceProposalID)
	handler.DescriptorMigrationHandler = strings.TrimSpace(handler.DescriptorMigrationHandler)
	handler.InterfaceMigrationHandler = strings.TrimSpace(handler.InterfaceMigrationHandler)
	handler.ProviderMigrationHandler = strings.TrimSpace(handler.ProviderMigrationHandler)
	handler.HandlerHash = strings.TrimSpace(handler.HandlerHash)
	return handler
}

func canonicalServiceInterfaceDeprecationMarker(marker ServiceInterfaceDeprecationMarker) ServiceInterfaceDeprecationMarker {
	marker.Reason = strings.TrimSpace(marker.Reason)
	marker.MarkerHash = strings.TrimSpace(marker.MarkerHash)
	return marker
}

func canonicalServiceProviderReregistrationRule(rule ServiceProviderReregistrationRule) ServiceProviderReregistrationRule {
	rule.ProviderID = strings.TrimSpace(rule.ProviderID)
	rule.ServiceID = strings.TrimSpace(rule.ServiceID)
	rule.RuleHash = strings.TrimSpace(rule.RuleHash)
	return rule
}

func canonicalServiceRegistryUpgradePlan(plan ServiceRegistryUpgradePlan) ServiceRegistryUpgradePlan {
	plan.GovernanceProposalID = strings.TrimSpace(plan.GovernanceProposalID)
	plan.DescriptorVersions = append([]ServiceDescriptorVersionRecord(nil), plan.DescriptorVersions...)
	for i := range plan.DescriptorVersions {
		plan.DescriptorVersions[i] = canonicalServiceDescriptorVersionRecord(plan.DescriptorVersions[i])
	}
	sort.SliceStable(plan.DescriptorVersions, func(i, j int) bool {
		left := fmt.Sprintf("%s/%s/%020d", plan.DescriptorVersions[i].ObjectKind, plan.DescriptorVersions[i].ObjectID, plan.DescriptorVersions[i].Version)
		right := fmt.Sprintf("%s/%s/%020d", plan.DescriptorVersions[j].ObjectKind, plan.DescriptorVersions[j].ObjectID, plan.DescriptorVersions[j].Version)
		return left < right
	})
	plan.SchemaCompatibility = append([]ServiceSchemaCompatibilityMetadata(nil), plan.SchemaCompatibility...)
	for i := range plan.SchemaCompatibility {
		plan.SchemaCompatibility[i] = canonicalServiceSchemaCompatibilityMetadata(plan.SchemaCompatibility[i])
	}
	sort.SliceStable(plan.SchemaCompatibility, func(i, j int) bool {
		return plan.SchemaCompatibility[i].CompatibilityHash < plan.SchemaCompatibility[j].CompatibilityHash
	})
	plan.MigrationHandlers = append([]ServiceRegistryMigrationHandler(nil), plan.MigrationHandlers...)
	for i := range plan.MigrationHandlers {
		plan.MigrationHandlers[i] = canonicalServiceRegistryMigrationHandler(plan.MigrationHandlers[i])
	}
	sort.SliceStable(plan.MigrationHandlers, func(i, j int) bool {
		return plan.MigrationHandlers[i].FromRegistryVersion < plan.MigrationHandlers[j].FromRegistryVersion
	})
	plan.InterfaceDeprecations = append([]ServiceInterfaceDeprecationMarker(nil), plan.InterfaceDeprecations...)
	for i := range plan.InterfaceDeprecations {
		plan.InterfaceDeprecations[i] = canonicalServiceInterfaceDeprecationMarker(plan.InterfaceDeprecations[i])
	}
	sort.SliceStable(plan.InterfaceDeprecations, func(i, j int) bool {
		return plan.InterfaceDeprecations[i].InterfaceHash < plan.InterfaceDeprecations[j].InterfaceHash
	})
	plan.ProviderReregistration = append([]ServiceProviderReregistrationRule(nil), plan.ProviderReregistration...)
	for i := range plan.ProviderReregistration {
		plan.ProviderReregistration[i] = canonicalServiceProviderReregistrationRule(plan.ProviderReregistration[i])
	}
	sort.SliceStable(plan.ProviderReregistration, func(i, j int) bool {
		left := plan.ProviderReregistration[i].ServiceID + "/" + plan.ProviderReregistration[i].ProviderID
		right := plan.ProviderReregistration[j].ServiceID + "/" + plan.ProviderReregistration[j].ProviderID
		return left < right
	})
	plan.PlanHash = strings.TrimSpace(plan.PlanHash)
	return plan
}

func validateMigrationHandlerChain(fromVersion, toVersion uint64, handlers []ServiceRegistryMigrationHandler) error {
	if len(handlers) == 0 {
		return errors.New("services registry upgrade requires migration handlers")
	}
	ordered := append([]ServiceRegistryMigrationHandler(nil), handlers...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].FromRegistryVersion < ordered[j].FromRegistryVersion })
	expectedFrom := fromVersion
	for _, handler := range ordered {
		if err := handler.Validate(); err != nil {
			return err
		}
		if handler.FromRegistryVersion != expectedFrom {
			return fmt.Errorf("services registry migration chain gap at version %d", expectedFrom)
		}
		expectedFrom = handler.ToRegistryVersion
	}
	if expectedFrom != toVersion {
		return fmt.Errorf("services registry migration chain ended at version %d", expectedFrom)
	}
	return nil
}

func countProviderReregistrations(rules []ServiceProviderReregistrationRule) int {
	count := 0
	for _, rule := range rules {
		if rule.Mode != ServiceProviderReregistrationUnchanged {
			count++
		}
	}
	return count
}
