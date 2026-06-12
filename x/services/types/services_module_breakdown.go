package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type XServicesStateObject string
type XServicesMessageName string
type XServicesQueryName string
type XServicesFailureMode string
type XServicesIntegrationPoint string

const (
	XServicesStateDescriptor	XServicesStateObject	= "ServiceDescriptor"
	XServicesStateAnchor		XServicesStateObject	= "ServiceAnchor"
	XServicesStateIdentityBinding	XServicesStateObject	= "IdentityServiceBinding"
	XServicesStateStatus		XServicesStateObject	= "ServiceStatus"
	XServicesStateParams		XServicesStateObject	= "ServiceParams"

	XServicesMsgRegisterService		XServicesMessageName	= "MsgRegisterService"
	XServicesMsgUpdateService		XServicesMessageName	= "MsgUpdateService"
	XServicesMsgRenewService		XServicesMessageName	= "MsgRenewService"
	XServicesMsgDisableService		XServicesMessageName	= "MsgDisableService"
	XServicesMsgTransferService		XServicesMessageName	= "MsgTransferService"
	XServicesMsgBindServiceIdentity		XServicesMessageName	= "MsgBindServiceIdentity"
	XServicesMsgUnbindServiceIdentity	XServicesMessageName	= "MsgUnbindServiceIdentity"

	XServicesQueryService			XServicesQueryName	= "QueryService"
	XServicesQueryServiceByName		XServicesQueryName	= "QueryServiceByName"
	XServicesQueryServicesByOwner		XServicesQueryName	= "QueryServicesByOwner"
	XServicesQueryServicesByIdentity	XServicesQueryName	= "QueryServicesByIdentity"
	XServicesQueryServiceProof		XServicesQueryName	= "QueryServiceProof"

	XServicesFailureDuplicateServiceID			XServicesFailureMode	= "duplicate_service_id"
	XServicesFailureUnauthorizedDescriptorUpdate		XServicesFailureMode	= "unauthorized_descriptor_update"
	XServicesFailureExpiredDescriptorUsedForCall		XServicesFailureMode	= "expired_descriptor_used_for_call"
	XServicesFailureInterfaceHashMismatch			XServicesFailureMode	= "interface_hash_mismatch"
	XServicesFailureStaleIdentityBindingAfterTransfer	XServicesFailureMode	= "identity_binding_stale_after_domain_transfer"

	XServicesIntegrationIdentity		XServicesIntegrationPoint	= "x/identity"
	XServicesIntegrationServiceInterface	XServicesIntegrationPoint	= "x/serviceinterface"
	XServicesIntegrationServiceCalls	XServicesIntegrationPoint	= "x/servicecalls"
	XServicesIntegrationStoreV2ProofQuery	XServicesIntegrationPoint	= "store_v2_proof_queries"
)

type XServicesFailureCoverage struct {
	Mode		XServicesFailureMode
	Guard		string
	StoreScope	string
}

type XServicesModuleBreakdown struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]XServicesStateObject
	Messages		[]XServicesMessageName
	Queries			[]XServicesQueryName
	FailureModes		[]XServicesFailureCoverage
	IntegrationPoints	[]XServicesIntegrationPoint
	BreakdownHash		string
}

type XServicesCallDescriptorCheck struct {
	ServiceID		string
	ExpectedInterfaceHash	string
	CurrentHeight		uint64
	DescriptorStatus	CanonicalServiceStatus
	DescriptorTTLHeight	uint64
	DescriptorHash		string
	CheckHash		string
}

type XServicesIdentityBindingFreshnessCheck struct {
	ServiceID		string
	IdentityName		string
	BoundOwner		string
	CurrentIdentityOwner	string
	BoundHeight		uint64
	CurrentHeight		uint64
	CheckHash		string
}

func DefaultXServicesModuleBreakdown() (XServicesModuleBreakdown, error) {
	breakdown := XServicesModuleBreakdown{
		ModulePath:	ServiceModuleServices,
		Purpose: []string{
			"identity_binding",
			"lifecycle",
			"registry_queries",
			"service_anchors",
			"service_descriptors",
		},
		StateObjects: []XServicesStateObject{
			XServicesStateDescriptor,
			XServicesStateAnchor,
			XServicesStateIdentityBinding,
			XServicesStateStatus,
			XServicesStateParams,
		},
		Messages: []XServicesMessageName{
			XServicesMsgRegisterService,
			XServicesMsgUpdateService,
			XServicesMsgRenewService,
			XServicesMsgDisableService,
			XServicesMsgTransferService,
			XServicesMsgBindServiceIdentity,
			XServicesMsgUnbindServiceIdentity,
		},
		Queries: []XServicesQueryName{
			XServicesQueryService,
			XServicesQueryServiceByName,
			XServicesQueryServicesByOwner,
			XServicesQueryServicesByIdentity,
			XServicesQueryServiceProof,
		},
		FailureModes: []XServicesFailureCoverage{
			newXServicesFailureCoverage(XServicesFailureDuplicateServiceID, "RegisterServiceInRegistryV2", ServiceStoreV2DescriptorPrefix),
			newXServicesFailureCoverage(XServicesFailureUnauthorizedDescriptorUpdate, "UpdateServiceInRegistryV2", ServiceStoreV2DescriptorPrefix),
			newXServicesFailureCoverage(XServicesFailureExpiredDescriptorUsedForCall, "ValidateXServicesDescriptorUsableForCall", ServiceStoreV2DescriptorPrefix),
			newXServicesFailureCoverage(XServicesFailureInterfaceHashMismatch, "ValidateXServicesDescriptorUsableForCall", ServiceStoreV2InterfacePrefix),
			newXServicesFailureCoverage(XServicesFailureStaleIdentityBindingAfterTransfer, "ValidateXServicesIdentityBindingFreshness", ServiceStoreV2IdentityIndexPrefix),
		},
		IntegrationPoints: []XServicesIntegrationPoint{
			XServicesIntegrationIdentity,
			XServicesIntegrationServiceInterface,
			XServicesIntegrationServiceCalls,
			XServicesIntegrationStoreV2ProofQuery,
		},
	}
	return NewXServicesModuleBreakdown(breakdown)
}

func NewXServicesModuleBreakdown(breakdown XServicesModuleBreakdown) (XServicesModuleBreakdown, error) {
	breakdown = canonicalXServicesModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return XServicesModuleBreakdown{}, err
	}
	breakdown.BreakdownHash = ComputeXServicesModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func ValidateXServicesDescriptorUsableForCall(descriptor CanonicalServiceDescriptor, expectedInterfaceHash string, currentHeight uint64) (XServicesCallDescriptorCheck, error) {
	if err := descriptor.Validate(); err != nil {
		return XServicesCallDescriptorCheck{}, err
	}
	if err := coretypes.ValidateHash("x/services expected interface hash", expectedInterfaceHash); err != nil {
		return XServicesCallDescriptorCheck{}, err
	}
	if currentHeight == 0 {
		return XServicesCallDescriptorCheck{}, errors.New("x/services call descriptor check height must be positive")
	}
	check := XServicesCallDescriptorCheck{
		ServiceID:		descriptor.ServiceID,
		ExpectedInterfaceHash:	expectedInterfaceHash,
		CurrentHeight:		currentHeight,
		DescriptorStatus:	descriptor.Status,
		DescriptorTTLHeight:	descriptor.TTLHeight,
		DescriptorHash:		descriptor.DescriptorHash,
	}
	check.CheckHash = ComputeXServicesCallDescriptorCheckHash(check)
	if descriptor.Status != CanonicalServiceStatusActive {
		return check, fmt.Errorf("x/services descriptor %s is not active", descriptor.ServiceID)
	}
	if currentHeight > descriptor.TTLHeight {
		return check, fmt.Errorf("x/services expired descriptor %s used for call", descriptor.ServiceID)
	}
	if descriptor.InterfaceHash != expectedInterfaceHash {
		return check, fmt.Errorf("x/services interface hash mismatch for %s", descriptor.ServiceID)
	}
	return check, check.Validate()
}

func ValidateXServicesIdentityBindingFreshness(binding ServiceIdentityBindingV2, currentIdentityOwner string, currentHeight uint64) (XServicesIdentityBindingFreshnessCheck, error) {
	if err := binding.Validate(); err != nil {
		return XServicesIdentityBindingFreshnessCheck{}, err
	}
	if err := validateInterfaceToken("x/services current identity owner", currentIdentityOwner); err != nil {
		return XServicesIdentityBindingFreshnessCheck{}, err
	}
	if currentHeight == 0 {
		return XServicesIdentityBindingFreshnessCheck{}, errors.New("x/services identity freshness height must be positive")
	}
	check := XServicesIdentityBindingFreshnessCheck{
		ServiceID:		binding.ServiceID,
		IdentityName:		binding.IdentityName,
		BoundOwner:		binding.Owner,
		CurrentIdentityOwner:	currentIdentityOwner,
		BoundHeight:		binding.BoundHeight,
		CurrentHeight:		currentHeight,
	}
	check.CheckHash = ComputeXServicesIdentityBindingFreshnessCheckHash(check)
	if currentHeight < binding.BoundHeight {
		return check, errors.New("x/services identity freshness height cannot precede binding height")
	}
	if binding.Owner != currentIdentityOwner {
		return check, fmt.Errorf("x/services identity binding %s/%s stale after domain transfer", binding.ServiceID, binding.IdentityName)
	}
	return check, check.Validate()
}

func (breakdown XServicesModuleBreakdown) ValidateFormat() error {
	if breakdown.ModulePath != ServiceModuleServices {
		return errors.New("x/services breakdown must describe x/services")
	}
	if err := validateSortedTokens("x/services purpose", breakdown.Purpose); err != nil {
		return err
	}
	if err := validateXServicesStateObjects(breakdown.StateObjects); err != nil {
		return err
	}
	if err := validateXServicesMessages(breakdown.Messages); err != nil {
		return err
	}
	if err := validateXServicesQueries(breakdown.Queries); err != nil {
		return err
	}
	if err := validateXServicesFailureCoverage(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateXServicesIntegrationPoints(breakdown.IntegrationPoints); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return coretypes.ValidateHash("x/services breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown XServicesModuleBreakdown) Validate() error {
	breakdown = canonicalXServicesModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("x/services breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeXServicesModuleBreakdownHash(breakdown) {
		return errors.New("x/services breakdown hash mismatch")
	}
	return nil
}

func (coverage XServicesFailureCoverage) Validate() error {
	if !IsXServicesFailureMode(coverage.Mode) {
		return fmt.Errorf("x/services unknown failure mode %q", coverage.Mode)
	}
	if err := validateInterfaceToken("x/services failure guard", coverage.Guard); err != nil {
		return err
	}
	if !IsServiceStoreKey(coverage.StoreScope + "/_") {
		return fmt.Errorf("x/services failure scope %s must be services store key", coverage.StoreScope)
	}
	return nil
}

func (check XServicesCallDescriptorCheck) Validate() error {
	if err := validateInterfaceToken("x/services call check service id", check.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/services call check interface hash", check.ExpectedInterfaceHash); err != nil {
		return err
	}
	if check.CurrentHeight == 0 || check.DescriptorTTLHeight == 0 {
		return errors.New("x/services call check heights must be positive")
	}
	if !IsCanonicalServiceStatus(check.DescriptorStatus) {
		return fmt.Errorf("x/services call check unknown descriptor status %q", check.DescriptorStatus)
	}
	if err := coretypes.ValidateHash("x/services call check descriptor hash", check.DescriptorHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/services call check hash", check.CheckHash); err != nil {
		return err
	}
	if check.CheckHash != ComputeXServicesCallDescriptorCheckHash(check) {
		return errors.New("x/services call check hash mismatch")
	}
	return nil
}

func (check XServicesIdentityBindingFreshnessCheck) Validate() error {
	if err := validateInterfaceToken("x/services identity freshness service id", check.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/services identity freshness identity name", check.IdentityName); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/services identity freshness bound owner", check.BoundOwner); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/services identity freshness current owner", check.CurrentIdentityOwner); err != nil {
		return err
	}
	if check.BoundHeight == 0 || check.CurrentHeight == 0 {
		return errors.New("x/services identity freshness heights must be positive")
	}
	if err := coretypes.ValidateHash("x/services identity freshness check hash", check.CheckHash); err != nil {
		return err
	}
	if check.CheckHash != ComputeXServicesIdentityBindingFreshnessCheckHash(check) {
		return errors.New("x/services identity freshness check hash mismatch")
	}
	return nil
}

func IsXServicesStateObject(object XServicesStateObject) bool {
	switch object {
	case XServicesStateDescriptor, XServicesStateAnchor, XServicesStateIdentityBinding, XServicesStateStatus, XServicesStateParams:
		return true
	default:
		return false
	}
}

func IsXServicesMessageName(message XServicesMessageName) bool {
	switch message {
	case XServicesMsgRegisterService, XServicesMsgUpdateService, XServicesMsgRenewService, XServicesMsgDisableService, XServicesMsgTransferService, XServicesMsgBindServiceIdentity, XServicesMsgUnbindServiceIdentity:
		return true
	default:
		return false
	}
}

func IsXServicesQueryName(query XServicesQueryName) bool {
	switch query {
	case XServicesQueryService, XServicesQueryServiceByName, XServicesQueryServicesByOwner, XServicesQueryServicesByIdentity, XServicesQueryServiceProof:
		return true
	default:
		return false
	}
}

func IsXServicesFailureMode(mode XServicesFailureMode) bool {
	switch mode {
	case XServicesFailureDuplicateServiceID, XServicesFailureUnauthorizedDescriptorUpdate, XServicesFailureExpiredDescriptorUsedForCall, XServicesFailureInterfaceHashMismatch, XServicesFailureStaleIdentityBindingAfterTransfer:
		return true
	default:
		return false
	}
}

func IsXServicesIntegrationPoint(point XServicesIntegrationPoint) bool {
	switch point {
	case XServicesIntegrationIdentity, XServicesIntegrationServiceInterface, XServicesIntegrationServiceCalls, XServicesIntegrationStoreV2ProofQuery:
		return true
	default:
		return false
	}
}

func ComputeXServicesModuleBreakdownHash(breakdown XServicesModuleBreakdown) string {
	breakdown = canonicalXServicesModuleBreakdown(breakdown)
	parts := []string{
		"aetra-x-services-module-breakdown-v1",
		breakdown.ModulePath,
		"purpose",
		fmt.Sprint(len(breakdown.Purpose)),
	}
	parts = append(parts, breakdown.Purpose...)
	parts = append(parts, "state", fmt.Sprint(len(breakdown.StateObjects)))
	for _, object := range breakdown.StateObjects {
		parts = append(parts, string(object))
	}
	parts = append(parts, "messages", fmt.Sprint(len(breakdown.Messages)))
	for _, message := range breakdown.Messages {
		parts = append(parts, string(message))
	}
	parts = append(parts, "queries", fmt.Sprint(len(breakdown.Queries)))
	for _, query := range breakdown.Queries {
		parts = append(parts, string(query))
	}
	parts = append(parts, "failures", fmt.Sprint(len(breakdown.FailureModes)))
	for _, coverage := range breakdown.FailureModes {
		parts = append(parts, string(coverage.Mode), coverage.Guard, coverage.StoreScope)
	}
	parts = append(parts, "integrations", fmt.Sprint(len(breakdown.IntegrationPoints)))
	for _, point := range breakdown.IntegrationPoints {
		parts = append(parts, string(point))
	}
	return servicesHashParts(parts...)
}

func ComputeXServicesCallDescriptorCheckHash(check XServicesCallDescriptorCheck) string {
	return servicesHashParts(
		"aetra-x-services-call-descriptor-check-v1",
		check.ServiceID,
		check.ExpectedInterfaceHash,
		fmt.Sprint(check.CurrentHeight),
		string(check.DescriptorStatus),
		fmt.Sprint(check.DescriptorTTLHeight),
		check.DescriptorHash,
	)
}

func ComputeXServicesIdentityBindingFreshnessCheckHash(check XServicesIdentityBindingFreshnessCheck) string {
	return servicesHashParts(
		"aetra-x-services-identity-binding-freshness-v1",
		check.ServiceID,
		check.IdentityName,
		check.BoundOwner,
		check.CurrentIdentityOwner,
		fmt.Sprint(check.BoundHeight),
		fmt.Sprint(check.CurrentHeight),
	)
}

func newXServicesFailureCoverage(mode XServicesFailureMode, guard, scope string) XServicesFailureCoverage {
	return XServicesFailureCoverage{Mode: mode, Guard: guard, StoreScope: scope}
}

func canonicalXServicesModuleBreakdown(breakdown XServicesModuleBreakdown) XServicesModuleBreakdown {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedStrings(breakdown.Purpose)
	breakdown.StateObjects = sortedXServicesStateObjects(breakdown.StateObjects)
	breakdown.Messages = sortedXServicesMessages(breakdown.Messages)
	breakdown.Queries = sortedXServicesQueries(breakdown.Queries)
	breakdown.FailureModes = sortedXServicesFailureCoverage(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedXServicesIntegrationPoints(breakdown.IntegrationPoints)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateXServicesStateObjects(objects []XServicesStateObject) error {
	required := []XServicesStateObject{XServicesStateDescriptor, XServicesStateAnchor, XServicesStateIdentityBinding, XServicesStateStatus, XServicesStateParams}
	return validateXServicesEnumSet("state object", objects, required, IsXServicesStateObject)
}

func validateXServicesMessages(messages []XServicesMessageName) error {
	required := []XServicesMessageName{XServicesMsgRegisterService, XServicesMsgUpdateService, XServicesMsgRenewService, XServicesMsgDisableService, XServicesMsgTransferService, XServicesMsgBindServiceIdentity, XServicesMsgUnbindServiceIdentity}
	return validateXServicesEnumSet("message", messages, required, IsXServicesMessageName)
}

func validateXServicesQueries(queries []XServicesQueryName) error {
	required := []XServicesQueryName{XServicesQueryService, XServicesQueryServiceByName, XServicesQueryServicesByOwner, XServicesQueryServicesByIdentity, XServicesQueryServiceProof}
	return validateXServicesEnumSet("query", queries, required, IsXServicesQueryName)
}

func validateXServicesFailureCoverage(coverage []XServicesFailureCoverage) error {
	required := []XServicesFailureMode{XServicesFailureDuplicateServiceID, XServicesFailureUnauthorizedDescriptorUpdate, XServicesFailureExpiredDescriptorUsedForCall, XServicesFailureInterfaceHashMismatch, XServicesFailureStaleIdentityBindingAfterTransfer}
	if len(coverage) != len(required) {
		return fmt.Errorf("x/services expected %d failure modes", len(required))
	}
	seen := map[XServicesFailureMode]struct{}{}
	for _, item := range coverage {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, found := seen[item.Mode]; found {
			return fmt.Errorf("x/services duplicate failure mode %s", item.Mode)
		}
		seen[item.Mode] = struct{}{}
	}
	for _, mode := range required {
		if _, found := seen[mode]; !found {
			return fmt.Errorf("x/services missing failure mode %s", mode)
		}
	}
	return nil
}

func validateXServicesIntegrationPoints(points []XServicesIntegrationPoint) error {
	required := []XServicesIntegrationPoint{XServicesIntegrationIdentity, XServicesIntegrationServiceInterface, XServicesIntegrationServiceCalls, XServicesIntegrationStoreV2ProofQuery}
	return validateXServicesEnumSet("integration", points, required, IsXServicesIntegrationPoint)
}

func validateXServicesEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("x/services expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("x/services unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("x/services %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("x/services duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/services missing %s %s", label, value)
		}
	}
	return nil
}

func sortedXServicesStateObjects(values []XServicesStateObject) []XServicesStateObject {
	out := append([]XServicesStateObject(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServicesMessages(values []XServicesMessageName) []XServicesMessageName {
	out := append([]XServicesMessageName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServicesQueries(values []XServicesQueryName) []XServicesQueryName {
	out := append([]XServicesQueryName(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedXServicesFailureCoverage(values []XServicesFailureCoverage) []XServicesFailureCoverage {
	out := append([]XServicesFailureCoverage(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.Trim(strings.TrimSpace(out[i].StoreScope), "/")
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedXServicesIntegrationPoints(values []XServicesIntegrationPoint) []XServicesIntegrationPoint {
	out := append([]XServicesIntegrationPoint(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
