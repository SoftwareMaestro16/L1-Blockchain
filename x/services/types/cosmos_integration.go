package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ServiceIntegrationModuleKind string
type ServiceKeeperBoundaryName string

const (
	ServiceIntegrationModuleRequired	ServiceIntegrationModuleKind	= "required"
	ServiceIntegrationModuleOptional	ServiceIntegrationModuleKind	= "optional"

	ServicesKeeperBoundary	ServiceKeeperBoundaryName	= "ServicesKeeper"
	InterfaceKeeperBoundary	ServiceKeeperBoundaryName	= "InterfaceKeeper"
	CallKeeperBoundary	ServiceKeeperBoundaryName	= "CallKeeper"
	PaymentKeeperBoundary	ServiceKeeperBoundaryName	= "PaymentKeeper"
	ProviderKeeperBoundary	ServiceKeeperBoundaryName	= "ProviderKeeper"
	ReceiptKeeperBoundary	ServiceKeeperBoundaryName	= "ReceiptKeeper"

	ServiceModuleServices		= "x/services"
	ServiceModuleInterface		= "x/serviceinterface"
	ServiceModuleCalls		= "x/servicecalls"
	ServiceModulePayments		= "x/servicepayments"
	ServiceModuleProviders		= "x/serviceproviders"
	ServiceModuleReceipts		= "x/servicereceipts"
	ServiceModuleIdentity		= "x/identity"
	ServiceModuleStorage		= "x/storage"
	ServiceModuleRouting		= "x/routing"
	ServiceModuleExternalPayment	= "x/payments"
	ServiceModuleContracts		= "x/contracts"
	ServiceModuleAVM		= "x/avm"
)

type ServiceIntegrationModule struct {
	ModulePath	string
	Kind		ServiceIntegrationModuleKind
	Purpose		string
	StoreKey	string
	ModuleHash	string
}

type ServiceKeeperBoundary struct {
	KeeperName		ServiceKeeperBoundaryName
	ModulePath		string
	Responsibilities	[]string
	StorePrefixes		[]string
	BoundaryHash		string
}

type ServiceKeeperIntegrationRules struct {
	StoreV2IsolatedPrefixes		bool
	CrossZoneOperationMode		string
	PaymentSettlementIntegration	string
	IdentityAuthorizationModule	string
	ContractExecutionIntegration	string
	RulesHash			string
}

type CosmosSDKServiceIntegrationManifest struct {
	RequiredModules		[]ServiceIntegrationModule
	OptionalModules		[]ServiceIntegrationModule
	KeeperBoundaries	[]ServiceKeeperBoundary
	Rules			ServiceKeeperIntegrationRules
	ManifestHash		string
}

func DefaultCosmosSDKServiceIntegrationManifest() (CosmosSDKServiceIntegrationManifest, error) {
	required := []ServiceIntegrationModule{
		newServiceIntegrationModule(ServiceModuleServices, ServiceIntegrationModuleRequired, "descriptors_anchors_lifecycle", StoreKey),
		newServiceIntegrationModule(ServiceModuleInterface, ServiceIntegrationModuleRequired, "schemas_methods_interface_proofs", StoreKey),
		newServiceIntegrationModule(ServiceModuleCalls, ServiceIntegrationModuleRequired, "call_envelopes_nonces_idempotency_receipts", StoreKey),
		newServiceIntegrationModule(ServiceModulePayments, ServiceIntegrationModuleRequired, "payment_models_escrow_stream_metering", StoreKey),
		newServiceIntegrationModule(ServiceModuleProviders, ServiceIntegrationModuleRequired, "provider_registry_stake_collateral_reputation", StoreKey),
		newServiceIntegrationModule(ServiceModuleReceipts, ServiceIntegrationModuleRequired, "receipt_roots_tombstones_proof_queries", StoreKey),
	}
	optional := []ServiceIntegrationModule{
		newServiceIntegrationModule(ServiceModuleIdentity, ServiceIntegrationModuleOptional, "aet_service_binding_authorization", "identity"),
		newServiceIntegrationModule(ServiceModuleStorage, ServiceIntegrationModuleOptional, "hybrid_storage_commitments", "storage"),
		newServiceIntegrationModule(ServiceModuleRouting, ServiceIntegrationModuleOptional, "service_network_routing", "routing"),
		newServiceIntegrationModule(ServiceModuleExternalPayment, ServiceIntegrationModuleOptional, "streaming_escrow_settlement", "payments"),
		newServiceIntegrationModule(ServiceModuleContracts, ServiceIntegrationModuleOptional, "contract_backed_services", "contracts"),
		newServiceIntegrationModule(ServiceModuleAVM, ServiceIntegrationModuleOptional, "avm_backed_services", "avm"),
	}
	boundaries := []ServiceKeeperBoundary{
		newServiceKeeperBoundary(ServicesKeeperBoundary, ServiceModuleServices, []string{"descriptors", "anchors", "lifecycle"}, []string{
			ServiceStorePrefix + "descriptors",
			ServiceStorePrefix + "anchors",
			ServiceStorePrefix + "owners",
			ServiceStorePrefix + "names",
			ServiceStorePrefix + "identity_bindings",
		}),
		newServiceKeeperBoundary(InterfaceKeeperBoundary, ServiceModuleInterface, []string{"schemas", "methods", "interface_proofs"}, []string{
			ServiceStorePrefix + "interfaces",
			ServiceStorePrefix + "interface_schemas",
			ServiceStorePrefix + "interface_proofs",
		}),
		newServiceKeeperBoundary(CallKeeperBoundary, ServiceModuleCalls, []string{"call_envelopes", "nonces", "idempotency", "receipts"}, []string{
			ServiceStorePrefix + "calls/envelopes",
			ServiceStorePrefix + "calls/nonces",
			ServiceStorePrefix + "calls/idempotency",
			ServiceStorePrefix + "calls/retries",
		}),
		newServiceKeeperBoundary(PaymentKeeperBoundary, ServiceModulePayments, []string{"payment_models", "escrow", "stream", "metering"}, []string{
			ServicePaymentModelPrefix,
			ServicePaymentEscrowPrefix,
			ServicePaymentStreamPrefix,
			ServicePaymentMeterPrefix,
			ServicePaymentSettlementPrefix,
		}),
		newServiceKeeperBoundary(ProviderKeeperBoundary, ServiceModuleProviders, []string{"provider_registry", "stake", "collateral", "reputation"}, []string{
			ServiceStorePrefix + "providers",
			ServiceStorePrefix + "providers/collateral",
			ServiceStorePrefix + "providers/stake",
			ServiceStorePrefix + "reputation",
		}),
		newServiceKeeperBoundary(ReceiptKeeperBoundary, ServiceModuleReceipts, []string{"receipt_roots", "tombstones", "proof_queries"}, []string{
			ServiceStorePrefix + "receipts",
			ServiceStorePrefix + "receipts/roots",
			ServiceStorePrefix + "receipts/tombstones",
			ServiceStorePrefix + "receipts/proofs",
		}),
	}
	rules := ServiceKeeperIntegrationRules{
		StoreV2IsolatedPrefixes:	true,
		CrossZoneOperationMode:		"messages",
		PaymentSettlementIntegration:	"bank_or_financial_zone",
		IdentityAuthorizationModule:	ServiceModuleIdentity,
		ContractExecutionIntegration:	"contract_module_interface",
	}
	return NewCosmosSDKServiceIntegrationManifest(required, optional, boundaries, rules)
}

func NewCosmosSDKServiceIntegrationManifest(required []ServiceIntegrationModule, optional []ServiceIntegrationModule, boundaries []ServiceKeeperBoundary, rules ServiceKeeperIntegrationRules) (CosmosSDKServiceIntegrationManifest, error) {
	manifest := CosmosSDKServiceIntegrationManifest{
		RequiredModules:	normalizeServiceIntegrationModules(required),
		OptionalModules:	normalizeServiceIntegrationModules(optional),
		KeeperBoundaries:	normalizeServiceKeeperBoundaries(boundaries),
		Rules:			canonicalServiceKeeperIntegrationRules(rules),
	}
	if manifest.Rules.RulesHash == "" {
		manifest.Rules.RulesHash = ComputeServiceKeeperIntegrationRulesHash(manifest.Rules)
	}
	if err := manifest.ValidateFormat(); err != nil {
		return CosmosSDKServiceIntegrationManifest{}, err
	}
	manifest.ManifestHash = ComputeCosmosSDKServiceIntegrationManifestHash(manifest)
	return manifest, manifest.Validate()
}

func (manifest CosmosSDKServiceIntegrationManifest) ValidateFormat() error {
	if err := validateServiceIntegrationModules(manifest.RequiredModules, ServiceIntegrationModuleRequired, requiredServiceIntegrationModules()); err != nil {
		return err
	}
	if err := validateServiceIntegrationModules(manifest.OptionalModules, ServiceIntegrationModuleOptional, optionalServiceIntegrationModules()); err != nil {
		return err
	}
	if err := validateServiceKeeperBoundaries(manifest.KeeperBoundaries); err != nil {
		return err
	}
	return manifest.Rules.Validate()
}

func (manifest CosmosSDKServiceIntegrationManifest) Validate() error {
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	if err := validateKeeperModulesCovered(manifest.RequiredModules, manifest.KeeperBoundaries); err != nil {
		return err
	}
	if err := validateKeeperStorePrefixIsolation(manifest.KeeperBoundaries); err != nil {
		return err
	}
	if err := validateOptionalIntegrationRules(manifest.OptionalModules, manifest.Rules); err != nil {
		return err
	}
	if err := validateInterfaceToken("services cosmos integration manifest hash", manifest.ManifestHash); err != nil {
		return err
	}
	if expected := ComputeCosmosSDKServiceIntegrationManifestHash(manifest); manifest.ManifestHash != expected {
		return fmt.Errorf("services cosmos integration manifest hash mismatch: expected %s", expected)
	}
	return nil
}

func (module ServiceIntegrationModule) Validate(expectedKind ServiceIntegrationModuleKind) error {
	if err := validateInterfaceToken("services cosmos integration module path", module.ModulePath); err != nil {
		return err
	}
	if !IsServiceIntegrationModuleKind(module.Kind) {
		return fmt.Errorf("services cosmos integration unknown module kind %q", module.Kind)
	}
	if expectedKind != "" && module.Kind != expectedKind {
		return fmt.Errorf("services cosmos integration module %s kind mismatch", module.ModulePath)
	}
	if err := validateInterfaceToken("services cosmos integration module purpose", module.Purpose); err != nil {
		return err
	}
	if err := validateInterfaceToken("services cosmos integration module store key", module.StoreKey); err != nil {
		return err
	}
	if err := validateInterfaceToken("services cosmos integration module hash", module.ModuleHash); err != nil {
		return err
	}
	if expected := ComputeServiceIntegrationModuleHash(module); module.ModuleHash != expected {
		return fmt.Errorf("services cosmos integration module hash mismatch: expected %s", expected)
	}
	return nil
}

func (boundary ServiceKeeperBoundary) Validate() error {
	if !IsServiceKeeperBoundaryName(boundary.KeeperName) {
		return fmt.Errorf("services cosmos integration unknown keeper %q", boundary.KeeperName)
	}
	if err := validateInterfaceToken("services cosmos integration keeper module", boundary.ModulePath); err != nil {
		return err
	}
	if len(boundary.Responsibilities) == 0 {
		return errors.New("services cosmos integration keeper responsibilities are required")
	}
	if err := validateSortedTokens("services cosmos integration keeper responsibility", boundary.Responsibilities); err != nil {
		return err
	}
	if len(boundary.StorePrefixes) == 0 {
		return errors.New("services cosmos integration keeper store prefixes are required")
	}
	previous := ""
	for _, prefix := range boundary.StorePrefixes {
		if !IsServiceStoreKey(prefix + "/_") {
			return fmt.Errorf("services cosmos integration keeper prefix %s must use services Store v2 prefix", prefix)
		}
		if previous != "" && previous >= prefix {
			return errors.New("services cosmos integration keeper prefixes must be sorted canonically")
		}
		previous = prefix
	}
	if err := validateInterfaceToken("services cosmos integration keeper hash", boundary.BoundaryHash); err != nil {
		return err
	}
	if expected := ComputeServiceKeeperBoundaryHash(boundary); boundary.BoundaryHash != expected {
		return fmt.Errorf("services cosmos integration keeper hash mismatch: expected %s", expected)
	}
	return nil
}

func (rules ServiceKeeperIntegrationRules) Validate() error {
	if !rules.StoreV2IsolatedPrefixes {
		return errors.New("services cosmos integration requires isolated Store v2 prefixes")
	}
	if rules.CrossZoneOperationMode != "messages" {
		return errors.New("services cosmos integration cross-zone operations must use messages")
	}
	if rules.PaymentSettlementIntegration != "bank_or_financial_zone" {
		return errors.New("services cosmos integration payment settlement must use bank or Financial Zone integration")
	}
	if rules.IdentityAuthorizationModule != ServiceModuleIdentity {
		return errors.New("services cosmos integration identity binding must authorize through x/identity")
	}
	if rules.ContractExecutionIntegration != "contract_module_interface" {
		return errors.New("services cosmos integration contract execution must use contract module interface")
	}
	if err := validateInterfaceToken("services cosmos integration rules hash", rules.RulesHash); err != nil {
		return err
	}
	if expected := ComputeServiceKeeperIntegrationRulesHash(rules); rules.RulesHash != expected {
		return fmt.Errorf("services cosmos integration rules hash mismatch: expected %s", expected)
	}
	return nil
}

func IsServiceIntegrationModuleKind(kind ServiceIntegrationModuleKind) bool {
	switch kind {
	case ServiceIntegrationModuleRequired, ServiceIntegrationModuleOptional:
		return true
	default:
		return false
	}
}

func IsServiceKeeperBoundaryName(name ServiceKeeperBoundaryName) bool {
	switch name {
	case ServicesKeeperBoundary, InterfaceKeeperBoundary, CallKeeperBoundary, PaymentKeeperBoundary, ProviderKeeperBoundary, ReceiptKeeperBoundary:
		return true
	default:
		return false
	}
}

func ComputeServiceIntegrationModuleHash(module ServiceIntegrationModule) string {
	return servicesHashParts("aetra-services-cosmos-module-v1", module.ModulePath, string(module.Kind), module.Purpose, module.StoreKey)
}

func ComputeServiceKeeperBoundaryHash(boundary ServiceKeeperBoundary) string {
	return servicesHashParts(
		"aetra-services-cosmos-keeper-boundary-v1",
		string(boundary.KeeperName),
		boundary.ModulePath,
		strings.Join(boundary.Responsibilities, ","),
		strings.Join(boundary.StorePrefixes, ","),
	)
}

func ComputeServiceKeeperIntegrationRulesHash(rules ServiceKeeperIntegrationRules) string {
	return servicesHashParts(
		"aetra-services-cosmos-integration-rules-v1",
		fmt.Sprint(rules.StoreV2IsolatedPrefixes),
		rules.CrossZoneOperationMode,
		rules.PaymentSettlementIntegration,
		rules.IdentityAuthorizationModule,
		rules.ContractExecutionIntegration,
	)
}

func ComputeCosmosSDKServiceIntegrationManifestHash(manifest CosmosSDKServiceIntegrationManifest) string {
	required := normalizeServiceIntegrationModules(manifest.RequiredModules)
	optional := normalizeServiceIntegrationModules(manifest.OptionalModules)
	boundaries := normalizeServiceKeeperBoundaries(manifest.KeeperBoundaries)
	parts := []string{
		"aetra-services-cosmos-integration-manifest-v1",
		fmt.Sprint(len(required)),
		fmt.Sprint(len(optional)),
		fmt.Sprint(len(boundaries)),
		manifest.Rules.RulesHash,
	}
	for _, module := range required {
		parts = append(parts, module.ModuleHash)
	}
	for _, module := range optional {
		parts = append(parts, module.ModuleHash)
	}
	for _, boundary := range boundaries {
		parts = append(parts, boundary.BoundaryHash)
	}
	return servicesHashParts(parts...)
}

func newServiceIntegrationModule(modulePath string, kind ServiceIntegrationModuleKind, purpose, storeKey string) ServiceIntegrationModule {
	module := ServiceIntegrationModule{ModulePath: modulePath, Kind: kind, Purpose: purpose, StoreKey: storeKey}
	module.ModuleHash = ComputeServiceIntegrationModuleHash(module)
	return module
}

func newServiceKeeperBoundary(name ServiceKeeperBoundaryName, modulePath string, responsibilities []string, prefixes []string) ServiceKeeperBoundary {
	boundary := ServiceKeeperBoundary{
		KeeperName:		name,
		ModulePath:		modulePath,
		Responsibilities:	sortedStrings(responsibilities),
		StorePrefixes:		sortedStrings(prefixes),
	}
	boundary.BoundaryHash = ComputeServiceKeeperBoundaryHash(boundary)
	return boundary
}

func canonicalServiceKeeperIntegrationRules(rules ServiceKeeperIntegrationRules) ServiceKeeperIntegrationRules {
	rules.CrossZoneOperationMode = strings.TrimSpace(rules.CrossZoneOperationMode)
	rules.PaymentSettlementIntegration = strings.TrimSpace(rules.PaymentSettlementIntegration)
	rules.IdentityAuthorizationModule = strings.TrimSpace(rules.IdentityAuthorizationModule)
	rules.ContractExecutionIntegration = strings.TrimSpace(rules.ContractExecutionIntegration)
	rules.RulesHash = strings.TrimSpace(rules.RulesHash)
	return rules
}

func normalizeServiceIntegrationModules(modules []ServiceIntegrationModule) []ServiceIntegrationModule {
	out := append([]ServiceIntegrationModule(nil), modules...)
	for i := range out {
		out[i].ModulePath = strings.TrimSpace(out[i].ModulePath)
		out[i].Purpose = strings.TrimSpace(out[i].Purpose)
		out[i].StoreKey = strings.TrimSpace(out[i].StoreKey)
		out[i].ModuleHash = strings.TrimSpace(out[i].ModuleHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ModulePath < out[j].ModulePath })
	return out
}

func normalizeServiceKeeperBoundaries(boundaries []ServiceKeeperBoundary) []ServiceKeeperBoundary {
	out := append([]ServiceKeeperBoundary(nil), boundaries...)
	for i := range out {
		out[i].ModulePath = strings.TrimSpace(out[i].ModulePath)
		out[i].Responsibilities = sortedStrings(out[i].Responsibilities)
		out[i].StorePrefixes = sortedStrings(out[i].StorePrefixes)
		out[i].BoundaryHash = strings.TrimSpace(out[i].BoundaryHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].KeeperName < out[j].KeeperName })
	return out
}

func validateServiceIntegrationModules(modules []ServiceIntegrationModule, kind ServiceIntegrationModuleKind, required []string) error {
	if len(modules) != len(required) {
		return fmt.Errorf("services cosmos integration expected %d %s modules", len(required), kind)
	}
	requiredSet := map[string]struct{}{}
	for _, modulePath := range required {
		requiredSet[modulePath] = struct{}{}
	}
	seen := map[string]struct{}{}
	for _, module := range modules {
		if err := module.Validate(kind); err != nil {
			return err
		}
		if _, ok := requiredSet[module.ModulePath]; !ok {
			return fmt.Errorf("services cosmos integration unexpected %s module %s", kind, module.ModulePath)
		}
		if _, found := seen[module.ModulePath]; found {
			return fmt.Errorf("services cosmos integration duplicate module %s", module.ModulePath)
		}
		seen[module.ModulePath] = struct{}{}
	}
	return nil
}

func validateServiceKeeperBoundaries(boundaries []ServiceKeeperBoundary) error {
	if len(boundaries) != 6 {
		return errors.New("services cosmos integration requires six keeper boundaries")
	}
	seen := map[ServiceKeeperBoundaryName]struct{}{}
	for _, boundary := range boundaries {
		if err := boundary.Validate(); err != nil {
			return err
		}
		if _, found := seen[boundary.KeeperName]; found {
			return fmt.Errorf("services cosmos integration duplicate keeper %s", boundary.KeeperName)
		}
		seen[boundary.KeeperName] = struct{}{}
	}
	for _, expected := range []ServiceKeeperBoundaryName{ServicesKeeperBoundary, InterfaceKeeperBoundary, CallKeeperBoundary, PaymentKeeperBoundary, ProviderKeeperBoundary, ReceiptKeeperBoundary} {
		if _, found := seen[expected]; !found {
			return fmt.Errorf("services cosmos integration missing keeper %s", expected)
		}
	}
	return nil
}

func validateKeeperModulesCovered(required []ServiceIntegrationModule, boundaries []ServiceKeeperBoundary) error {
	requiredSet := map[string]struct{}{}
	for _, module := range required {
		requiredSet[module.ModulePath] = struct{}{}
	}
	for _, boundary := range boundaries {
		if _, found := requiredSet[boundary.ModulePath]; !found {
			return fmt.Errorf("services cosmos integration keeper %s references non-required module %s", boundary.KeeperName, boundary.ModulePath)
		}
	}
	return nil
}

func validateKeeperStorePrefixIsolation(boundaries []ServiceKeeperBoundary) error {
	owner := map[string]ServiceKeeperBoundaryName{}
	for _, boundary := range boundaries {
		for _, prefix := range boundary.StorePrefixes {
			if existing, found := owner[prefix]; found && existing != boundary.KeeperName {
				return fmt.Errorf("services cosmos integration prefix %s shared by %s and %s", prefix, existing, boundary.KeeperName)
			}
			for existingPrefix, existingKeeper := range owner {
				if existingKeeper == boundary.KeeperName {
					continue
				}
				if strings.HasPrefix(prefix+"/", existingPrefix+"/") || strings.HasPrefix(existingPrefix+"/", prefix+"/") {
					return fmt.Errorf("services cosmos integration prefix overlap between %s and %s", boundary.KeeperName, existingKeeper)
				}
			}
			owner[prefix] = boundary.KeeperName
		}
	}
	return nil
}

func validateOptionalIntegrationRules(optional []ServiceIntegrationModule, rules ServiceKeeperIntegrationRules) error {
	optionalSet := map[string]struct{}{}
	for _, module := range optional {
		optionalSet[module.ModulePath] = struct{}{}
	}
	if _, found := optionalSet[rules.IdentityAuthorizationModule]; !found {
		return errors.New("services cosmos integration identity authorization module must be declared optional")
	}
	if _, found := optionalSet[ServiceModuleExternalPayment]; !found {
		return errors.New("services cosmos integration payments module must be declared optional")
	}
	if _, contracts := optionalSet[ServiceModuleContracts]; !contracts {
		if _, avm := optionalSet[ServiceModuleAVM]; !avm {
			return errors.New("services cosmos integration requires contracts or avm optional integration")
		}
	}
	return nil
}

func requiredServiceIntegrationModules() []string {
	return []string{ServiceModuleServices, ServiceModuleInterface, ServiceModuleCalls, ServiceModulePayments, ServiceModuleProviders, ServiceModuleReceipts}
}

func optionalServiceIntegrationModules() []string {
	return []string{ServiceModuleIdentity, ServiceModuleStorage, ServiceModuleRouting, ServiceModuleExternalPayment, ServiceModuleContracts, ServiceModuleAVM}
}
