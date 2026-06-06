package types

import (
	"errors"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/app/addressing"
)

type ZoneType string
type ServiceType string
type ServiceStatus string
type ServiceLocation string
type ServiceTrustModel string
type ServiceVerificationModel string
type ServicePaymentSettlementMode string
type ServicePricingUnit string
type ServiceStorageModel string
type ServiceMethodExecutionType string
type ServiceFailureBehavior string

const (
	ZoneTypeCore        ZoneType = "CORE"
	ZoneTypeAetherCore  ZoneType = "AETHER_CORE"
	ZoneTypeFinancial   ZoneType = "FINANCIAL"
	ZoneTypeIdentity    ZoneType = "IDENTITY"
	ZoneTypeStorage     ZoneType = "STORAGE"
	ZoneTypePayment     ZoneType = "PAYMENT"
	ZoneTypeContract    ZoneType = "CONTRACT"
	ZoneTypeApplication ZoneType = "APPLICATION"
	ZoneTypeService     ZoneType = "SERVICE"

	ServiceTypeOnChain   ServiceType = "ON_CHAIN"
	ServiceTypeOffChain  ServiceType = "OFF_CHAIN"
	ServiceTypeMixed     ServiceType = "MIXED"
	ServiceTypeFogMarket ServiceType = "FOG_MARKET"

	ServiceStatusActive     ServiceStatus = "ACTIVE"
	ServiceStatusDisabled   ServiceStatus = "DISABLED"
	ServiceStatusDeprecated ServiceStatus = "DEPRECATED"

	ServiceLocationModule          ServiceLocation = "MODULE"
	ServiceLocationContract        ServiceLocation = "CONTRACT"
	ServiceLocationApplicationZone ServiceLocation = "APPLICATION_ZONE"
	ServiceLocationExternal        ServiceLocation = "EXTERNAL_ENDPOINT"
	ServiceLocationHybrid          ServiceLocation = "HYBRID_ENDPOINT"
	ServiceLocationProviderPool    ServiceLocation = "PROVIDER_POOL"

	ServiceTrustConsensusExecuted           ServiceTrustModel = "CONSENSUS_EXECUTED"
	ServiceTrustEconomicallySecured         ServiceTrustModel = "ECONOMICALLY_SECURED"
	ServiceTrustCryptographicallyVerifiable ServiceTrustModel = "CRYPTOGRAPHICALLY_VERIFIABLE"
	ServiceTrustHybridChallengeable         ServiceTrustModel = "HYBRID_CHALLENGEABLE"
	ServiceTrustFullyTrusted                ServiceTrustModel = "FULLY_TRUSTED"

	ServiceVerificationConsensusReceipt   ServiceVerificationModel = "CONSENSUS_RECEIPT"
	ServiceVerificationSignedResult       ServiceVerificationModel = "SIGNED_RESULT"
	ServiceVerificationProofAnchored      ServiceVerificationModel = "PROOF_ANCHORED"
	ServiceVerificationChallengeWindow    ServiceVerificationModel = "CHALLENGE_WINDOW"
	ServiceVerificationEconomicCollateral ServiceVerificationModel = "ECONOMIC_COLLATERAL"
	ServiceVerificationAdvisory           ServiceVerificationModel = "ADVISORY"

	ServicePaymentOnChain   ServicePaymentSettlementMode = "ON_CHAIN"
	ServicePaymentStreaming ServicePaymentSettlementMode = "STREAMING"
	ServicePaymentPrepaid   ServicePaymentSettlementMode = "PREPAID"
	ServicePaymentMetered   ServicePaymentSettlementMode = "METERED"
	ServicePaymentEscrow    ServicePaymentSettlementMode = "ESCROW"

	ServicePricingPerCall        ServicePricingUnit = "CALL"
	ServicePricingPerByte        ServicePricingUnit = "BYTE"
	ServicePricingPerComputeUnit ServicePricingUnit = "COMPUTE_UNIT"
	ServicePricingSubscription   ServicePricingUnit = "SUBSCRIPTION"

	ServiceStorageNone                ServiceStorageModel = "NONE"
	ServiceStorageEphemeral           ServiceStorageModel = "EPHEMERAL"
	ServiceStorageOnChain             ServiceStorageModel = "ON_CHAIN"
	ServiceStorageDistributedOffChain ServiceStorageModel = "OFF_CHAIN_DISTRIBUTED"
	ServiceStorageHybridCommitment    ServiceStorageModel = "HYBRID_COMMITMENT"

	ServiceMethodSync    ServiceMethodExecutionType = "SYNC"
	ServiceMethodAsync   ServiceMethodExecutionType = "ASYNC"
	ServiceMethodEvented ServiceMethodExecutionType = "EVENTED"

	ServiceFailureRevert          ServiceFailureBehavior = "REVERT"
	ServiceFailureRetry           ServiceFailureBehavior = "RETRY"
	ServiceFailureFallbackOnChain ServiceFailureBehavior = "FALLBACK_ON_CHAIN"
	ServiceFailureChallenge       ServiceFailureBehavior = "CHALLENGE"
	ServiceFailureSlashProvider   ServiceFailureBehavior = "SLASH_PROVIDER"
	ServiceFailureRefund          ServiceFailureBehavior = "REFUND"
	ServiceFailurePartialSettle   ServiceFailureBehavior = "PARTIAL_SETTLE"
)

type ZoneDescriptor struct {
	ZoneID                ZoneID
	ZoneType              ZoneType
	ModuleName            string
	Enabled               bool
	StateMachineVersion   uint64
	MempoolPolicyID       string
	FeePolicyID           string
	ShardLayoutEpoch      uint64
	MaxShards             uint32
	MessageCapabilities   []string
	ProofCapabilities     []string
	UpgradeHeightOptional uint64
}

type ServiceDescriptor struct {
	ServiceID        string
	Owner            string
	ServiceType      ServiceType
	ZoneID           ZoneID
	InterfaceID      string
	EndpointKey      string
	Version          uint64
	AvailabilityHash string
	Enabled          bool
	Status           ServiceStatus
	ExpiryHeight     uint64
	CreatedHeight    uint64
	UpdatedHeight    uint64
	Interface        ServiceInterfaceDescriptor
	Execution        ServiceExecutionDescriptor
	Discovery        ServiceDiscoveryDescriptor
	Payment          ServicePaymentDescriptor
	Storage          ServiceStorageDescriptor
	Verification     ServiceVerificationDescriptor
}

type ServiceInterfaceDescriptor struct {
	InterfaceID    string
	InterfaceName  string
	InterfaceHash  string
	Version        uint64
	SchemaEncoding string
	Methods        []ServiceMethodDescriptor
	Events         []string
	Errors         []string
	AuthModel      string
	PaymentModel   string
	MetadataHash   string
	CreatedHeight  uint64
}

type ServiceMethodDescriptor struct {
	MethodID             string
	Name                 string
	InputSchemaHash      string
	OutputSchemaHash     string
	ExecutionType        ServiceMethodExecutionType
	RequiredPaymentModel string
	GasModel             string
	VerificationModel    ServiceVerificationModel
	TimeoutHeightDelta   uint64
	IdempotencyRequired  bool
	CallbackSupported    bool
	FailureBehavior      ServiceFailureBehavior
}

type ServiceExecutionDescriptor struct {
	Location        ServiceLocation
	Target          string
	ModuleRoute     string
	ContractAddress string
	Endpoint        string
	ProviderPoolID  string
	Mode            ExecutionMode
	Deterministic   bool
	FailureBehavior ServiceFailureBehavior
	ResultExpiry    uint64
	ChallengeWindow uint64
}

type ServiceDiscoveryDescriptor struct {
	ServiceName            string
	IdentityName           string
	ProviderRoot           string
	MetadataHash           string
	ExternalDescriptorHash string
	CacheExpiryHeight      uint64
	SignaturePolicy        string
}

type ServicePaymentDescriptor struct {
	SettlementMode ServicePaymentSettlementMode
	Denom          string
	Amount         string
	MaxAmount      string
	PricingUnit    ServicePricingUnit
	EscrowRequired bool
	EscrowID       string
	MeterID        string
	ExpiryHeight   uint64
}

type ServiceStorageDescriptor struct {
	Model           ServiceStorageModel
	StateRootType   RootType
	CommitmentHash  string
	RetentionHeight uint64
	ProofRequired   bool
}

type ServiceVerificationDescriptor struct {
	TrustModel               ServiceTrustModel
	Model                    ServiceVerificationModel
	ProofFormat              string
	RequestSigningRequired   bool
	ResponseSigningRequired  bool
	ChallengeWindow          uint64
	FallbackServiceID        string
	ProviderCollateralDenom  string
	ProviderCollateralAmount string
	FaultPolicy              ServiceFailureBehavior
}

func (d ZoneDescriptor) Validate(params AetherCoreParams) error {
	if err := ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if !IsZoneType(d.ZoneType) {
		return fmt.Errorf("unknown aethercore zone type %q", d.ZoneType)
	}
	if err := validateModuleName(d.ModuleName); err != nil {
		return err
	}
	if d.StateMachineVersion == 0 {
		return errors.New("aethercore zone state machine version must be positive")
	}
	if err := validatePolicyID("aethercore zone mempool policy", d.MempoolPolicyID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore zone fee policy", d.FeePolicyID); err != nil {
		return err
	}
	if d.FeePolicyID != NativeFeePolicyID {
		return fmt.Errorf("aethercore zone fee policy must use %s", NativeFeePolicyID)
	}
	if d.MaxShards == 0 || d.MaxShards > params.MaxShardsPerZone {
		return fmt.Errorf("aethercore zone max shards must be between 1 and %d", params.MaxShardsPerZone)
	}
	if err := validateCapabilitiesForField("aethercore zone message capabilities", d.MessageCapabilities); err != nil {
		return err
	}
	return validateCapabilitiesForField("aethercore zone proof capabilities", d.ProofCapabilities)
}

func (d ServiceDescriptor) Validate() error {
	d = CanonicalServiceDescriptor(d)
	if err := validatePolicyID("aethercore service id", d.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aethercore service owner", d.Owner); err != nil {
		return err
	}
	if !IsServiceType(d.ServiceType) {
		return fmt.Errorf("unknown aethercore service type %q", d.ServiceType)
	}
	if !IsServiceStatus(d.Status) {
		return fmt.Errorf("unknown aethercore service status %q", d.Status)
	}
	if d.Enabled && d.Status != ServiceStatusActive {
		return errors.New("aethercore enabled service must be active")
	}
	if !d.Enabled && d.Status == ServiceStatusActive {
		return errors.New("aethercore active service must be enabled")
	}
	if err := ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service interface id", d.InterfaceID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service endpoint key", d.EndpointKey); err != nil {
		return err
	}
	if d.Version == 0 {
		return errors.New("aethercore service version must be positive")
	}
	if d.CreatedHeight == 0 {
		return errors.New("aethercore service created height must be positive")
	}
	if d.UpdatedHeight < d.CreatedHeight {
		return errors.New("aethercore service updated height must not precede created height")
	}
	if d.ExpiryHeight != 0 && d.ExpiryHeight <= d.UpdatedHeight {
		return errors.New("aethercore service expiry height must exceed updated height")
	}
	if err := ValidateHash("aethercore service availability hash", d.AvailabilityHash); err != nil {
		return err
	}
	if err := d.Interface.Validate(); err != nil {
		return err
	}
	if d.Interface.InterfaceID != d.InterfaceID {
		return errors.New("aethercore service interface id mismatch")
	}
	if d.Interface.Version != d.Version {
		return errors.New("aethercore service interface version mismatch")
	}
	if err := d.Execution.Validate(); err != nil {
		return err
	}
	if err := d.Discovery.Validate(); err != nil {
		return err
	}
	if d.Discovery.CacheExpiryHeight != 0 && d.ExpiryHeight != 0 && d.Discovery.CacheExpiryHeight > d.ExpiryHeight {
		return errors.New("aethercore discovery cache expiry must not outlive service expiry")
	}
	if err := d.Payment.Validate(); err != nil {
		return err
	}
	if d.Payment.ExpiryHeight != 0 && d.ExpiryHeight != 0 && d.Payment.ExpiryHeight > d.ExpiryHeight {
		return errors.New("aethercore payment expiry must not outlive service expiry")
	}
	if err := d.Storage.Validate(); err != nil {
		return err
	}
	if err := d.Verification.Validate(); err != nil {
		return err
	}
	return d.validateRuntimeRules()
}

func IsZoneType(zoneType ZoneType) bool {
	switch zoneType {
	case ZoneTypeCore, ZoneTypeAetherCore, ZoneTypeFinancial, ZoneTypeIdentity, ZoneTypeStorage, ZoneTypePayment, ZoneTypeContract, ZoneTypeApplication, ZoneTypeService:
		return true
	default:
		return false
	}
}

func IsServiceType(serviceType ServiceType) bool {
	switch serviceType {
	case ServiceTypeOnChain, ServiceTypeOffChain, ServiceTypeMixed, ServiceTypeFogMarket:
		return true
	default:
		return false
	}
}

func IsServiceStatus(status ServiceStatus) bool {
	switch status {
	case ServiceStatusActive, ServiceStatusDisabled, ServiceStatusDeprecated:
		return true
	default:
		return false
	}
}

func IsServiceLocation(location ServiceLocation) bool {
	switch location {
	case ServiceLocationModule, ServiceLocationContract, ServiceLocationApplicationZone, ServiceLocationExternal, ServiceLocationHybrid, ServiceLocationProviderPool:
		return true
	default:
		return false
	}
}

func IsServiceTrustModel(model ServiceTrustModel) bool {
	switch model {
	case ServiceTrustConsensusExecuted, ServiceTrustEconomicallySecured, ServiceTrustCryptographicallyVerifiable, ServiceTrustHybridChallengeable, ServiceTrustFullyTrusted:
		return true
	default:
		return false
	}
}

func IsServiceVerificationModel(model ServiceVerificationModel) bool {
	switch model {
	case ServiceVerificationConsensusReceipt, ServiceVerificationSignedResult, ServiceVerificationProofAnchored,
		ServiceVerificationChallengeWindow, ServiceVerificationEconomicCollateral, ServiceVerificationAdvisory:
		return true
	default:
		return false
	}
}

func IsServicePaymentSettlementMode(mode ServicePaymentSettlementMode) bool {
	switch mode {
	case ServicePaymentOnChain, ServicePaymentStreaming, ServicePaymentPrepaid, ServicePaymentMetered, ServicePaymentEscrow:
		return true
	default:
		return false
	}
}

func IsServicePricingUnit(unit ServicePricingUnit) bool {
	switch unit {
	case ServicePricingPerCall, ServicePricingPerByte, ServicePricingPerComputeUnit, ServicePricingSubscription:
		return true
	default:
		return false
	}
}

func IsServiceStorageModel(model ServiceStorageModel) bool {
	switch model {
	case ServiceStorageNone, ServiceStorageEphemeral, ServiceStorageOnChain, ServiceStorageDistributedOffChain, ServiceStorageHybridCommitment:
		return true
	default:
		return false
	}
}

func IsServiceMethodExecutionType(executionType ServiceMethodExecutionType) bool {
	switch executionType {
	case ServiceMethodSync, ServiceMethodAsync, ServiceMethodEvented:
		return true
	default:
		return false
	}
}

func IsServiceFailureBehavior(behavior ServiceFailureBehavior) bool {
	switch behavior {
	case ServiceFailureRevert, ServiceFailureRetry, ServiceFailureFallbackOnChain, ServiceFailureChallenge,
		ServiceFailureSlashProvider, ServiceFailureRefund, ServiceFailurePartialSettle:
		return true
	default:
		return false
	}
}

func CanonicalZoneDescriptor(d ZoneDescriptor) ZoneDescriptor {
	d.MessageCapabilities = append([]string(nil), d.MessageCapabilities...)
	d.ProofCapabilities = append([]string(nil), d.ProofCapabilities...)
	sort.Strings(d.MessageCapabilities)
	sort.Strings(d.ProofCapabilities)
	return d
}

func CanonicalServiceDescriptor(d ServiceDescriptor) ServiceDescriptor {
	if d.Interface.InterfaceID == "" {
		d.Interface.InterfaceID = d.InterfaceID
	}
	if d.InterfaceID == "" {
		d.InterfaceID = d.Interface.InterfaceID
	}
	if d.Interface.Version == 0 {
		d.Interface.Version = d.Version
	}
	if d.Version == 0 {
		d.Version = d.Interface.Version
	}
	if d.Status == "" {
		if d.Enabled {
			d.Status = ServiceStatusActive
		} else {
			d.Status = ServiceStatusDisabled
		}
	}
	if d.Execution.Mode == "" {
		switch d.ServiceType {
		case ServiceTypeOnChain:
			d.Execution.Mode = ExecutionModeSync
		default:
			d.Execution.Mode = ExecutionModeAsync
		}
	}
	if d.Execution.FailureBehavior == "" {
		d.Execution.FailureBehavior = ServiceFailureRevert
	}
	d.Interface = CanonicalServiceInterfaceDescriptor(d.Interface)
	return d
}

func CanonicalServiceInterfaceDescriptor(d ServiceInterfaceDescriptor) ServiceInterfaceDescriptor {
	d.Methods = cloneServiceMethods(d.Methods)
	d.Events = append([]string(nil), d.Events...)
	d.Errors = append([]string(nil), d.Errors...)
	sortServiceMethods(d.Methods)
	sort.Strings(d.Events)
	sort.Strings(d.Errors)
	return d
}

func (d ServiceInterfaceDescriptor) Validate() error {
	d = CanonicalServiceInterfaceDescriptor(d)
	if err := validatePolicyID("aethercore service interface id", d.InterfaceID); err != nil {
		return err
	}
	if d.InterfaceName != "" {
		if err := validatePolicyID("aethercore service interface name", d.InterfaceName); err != nil {
			return err
		}
	}
	if d.Version == 0 {
		return errors.New("aethercore service interface version must be positive")
	}
	if err := validatePolicyID("aethercore service schema encoding", d.SchemaEncoding); err != nil {
		return err
	}
	if len(d.Methods) == 0 {
		return errors.New("aethercore service interface requires methods")
	}
	if err := validateServiceMethods(d.Methods); err != nil {
		return err
	}
	if err := validateCapabilitiesForField("aethercore service interface event", d.Events); err != nil {
		return err
	}
	if err := validateCapabilitiesForField("aethercore service interface error", d.Errors); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service auth model", d.AuthModel); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service payment model", d.PaymentModel); err != nil {
		return err
	}
	if err := validateOptionalHash("aethercore service interface metadata hash", d.MetadataHash); err != nil {
		return err
	}
	if d.CreatedHeight == 0 {
		return errors.New("aethercore service interface created height must be positive")
	}
	if err := ValidateHash("aethercore service interface hash", d.InterfaceHash); err != nil {
		return err
	}
	expected := ComputeServiceInterfaceHash(d)
	if d.InterfaceHash != expected {
		return fmt.Errorf("aethercore service interface hash mismatch: expected %s", expected)
	}
	return nil
}

func (d ServiceMethodDescriptor) Validate() error {
	if err := validatePolicyID("aethercore service method id", d.MethodID); err != nil {
		return err
	}
	if err := validatePolicyID("aethercore service method name", d.Name); err != nil {
		return err
	}
	if err := ValidateHash("aethercore service method input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := ValidateHash("aethercore service method output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if !IsServiceMethodExecutionType(d.ExecutionType) {
		return fmt.Errorf("unknown aethercore service method execution type %q", d.ExecutionType)
	}
	if err := validatePolicyID("aethercore service method payment model", d.RequiredPaymentModel); err != nil {
		return err
	}
	if d.GasModel != "" {
		if err := validatePolicyID("aethercore service method gas model", d.GasModel); err != nil {
			return err
		}
	}
	if !IsServiceVerificationModel(d.VerificationModel) {
		return fmt.Errorf("unknown aethercore service method verification model %q", d.VerificationModel)
	}
	if d.TimeoutHeightDelta == 0 {
		return errors.New("aethercore service method timeout must be positive")
	}
	if !IsServiceFailureBehavior(d.FailureBehavior) {
		return fmt.Errorf("unknown aethercore service method failure behavior %q", d.FailureBehavior)
	}
	return nil
}

func (d ServiceExecutionDescriptor) Validate() error {
	if !IsServiceLocation(d.Location) {
		return fmt.Errorf("unknown aethercore service location %q", d.Location)
	}
	if d.Target != "" {
		if err := validatePolicyID("aethercore service execution target", d.Target); err != nil {
			return err
		}
	}
	if d.ModuleRoute != "" {
		if err := validateModuleName(d.ModuleRoute); err != nil {
			return err
		}
	}
	if d.ContractAddress != "" {
		if err := validatePolicyID("aethercore service contract address", d.ContractAddress); err != nil {
			return err
		}
	}
	if d.Endpoint != "" {
		if err := validatePolicyID("aethercore service endpoint", d.Endpoint); err != nil {
			return err
		}
	}
	if d.ProviderPoolID != "" {
		if err := validatePolicyID("aethercore service provider pool id", d.ProviderPoolID); err != nil {
			return err
		}
	}
	if d.Mode != ExecutionModeSync && d.Mode != ExecutionModeAsync {
		return fmt.Errorf("unknown aethercore service execution mode %q", d.Mode)
	}
	if !IsServiceFailureBehavior(d.FailureBehavior) {
		return fmt.Errorf("unknown aethercore service execution failure behavior %q", d.FailureBehavior)
	}
	switch d.Location {
	case ServiceLocationModule:
		if d.ModuleRoute == "" {
			return errors.New("aethercore module service requires module route")
		}
	case ServiceLocationContract:
		if d.ContractAddress == "" {
			return errors.New("aethercore contract service requires contract address")
		}
	case ServiceLocationApplicationZone:
		if d.Target == "" {
			return errors.New("aethercore application zone service requires execution target")
		}
	case ServiceLocationExternal, ServiceLocationHybrid:
		if d.Endpoint == "" {
			return errors.New("aethercore external service requires endpoint")
		}
	case ServiceLocationProviderPool:
		if d.ProviderPoolID == "" {
			return errors.New("aethercore provider pool service requires provider pool id")
		}
	}
	return nil
}

func (d ServiceDiscoveryDescriptor) Validate() error {
	if d.ServiceName != "" {
		if err := validatePolicyID("aethercore service discovery name", d.ServiceName); err != nil {
			return err
		}
	}
	if d.IdentityName != "" {
		if err := validatePolicyID("aethercore service identity name", d.IdentityName); err != nil {
			return err
		}
	}
	if err := validateOptionalHash("aethercore service provider root", d.ProviderRoot); err != nil {
		return err
	}
	if err := validateOptionalHash("aethercore service metadata hash", d.MetadataHash); err != nil {
		return err
	}
	if err := validateOptionalHash("aethercore external service descriptor hash", d.ExternalDescriptorHash); err != nil {
		return err
	}
	if d.SignaturePolicy != "" {
		if err := validatePolicyID("aethercore service discovery signature policy", d.SignaturePolicy); err != nil {
			return err
		}
	}
	return nil
}

func (d ServicePaymentDescriptor) Validate() error {
	if !IsServicePaymentSettlementMode(d.SettlementMode) {
		return fmt.Errorf("unknown aethercore service payment settlement mode %q", d.SettlementMode)
	}
	if err := validatePolicyID("aethercore service payment denom", d.Denom); err != nil {
		return err
	}
	if err := validateAmountString("aethercore service payment amount", d.Amount); err != nil {
		return err
	}
	if d.MaxAmount != "" {
		if err := validateAmountString("aethercore service max payment amount", d.MaxAmount); err != nil {
			return err
		}
	}
	if !IsServicePricingUnit(d.PricingUnit) {
		return fmt.Errorf("unknown aethercore service pricing unit %q", d.PricingUnit)
	}
	if d.EscrowID != "" {
		if err := validatePolicyID("aethercore service escrow id", d.EscrowID); err != nil {
			return err
		}
	}
	if d.MeterID != "" {
		if err := validatePolicyID("aethercore service meter id", d.MeterID); err != nil {
			return err
		}
	}
	if d.SettlementMode == ServicePaymentEscrow && !d.EscrowRequired {
		return errors.New("aethercore escrow payment mode requires escrow")
	}
	if d.SettlementMode == ServicePaymentMetered && d.MeterID == "" {
		return errors.New("aethercore metered payment mode requires meter id")
	}
	return nil
}

func (d ServiceStorageDescriptor) Validate() error {
	if !IsServiceStorageModel(d.Model) {
		return fmt.Errorf("unknown aethercore service storage model %q", d.Model)
	}
	if d.StateRootType != "" {
		if err := validateToken("aethercore service state root type", string(d.StateRootType), MaxScopeLength); err != nil {
			return err
		}
	}
	if err := validateOptionalHash("aethercore service storage commitment hash", d.CommitmentHash); err != nil {
		return err
	}
	if d.Model == ServiceStorageOnChain && d.StateRootType == "" {
		return errors.New("aethercore on-chain service storage requires state root type")
	}
	if d.Model == ServiceStorageHybridCommitment && d.CommitmentHash == "" {
		return errors.New("aethercore hybrid service storage requires commitment hash")
	}
	return nil
}

func (d ServiceVerificationDescriptor) Validate() error {
	if !IsServiceTrustModel(d.TrustModel) {
		return fmt.Errorf("unknown aethercore service trust model %q", d.TrustModel)
	}
	if !IsServiceVerificationModel(d.Model) {
		return fmt.Errorf("unknown aethercore service verification model %q", d.Model)
	}
	if d.ProofFormat != "" {
		if err := validatePolicyID("aethercore service proof format", d.ProofFormat); err != nil {
			return err
		}
	}
	if d.FallbackServiceID != "" {
		if err := validatePolicyID("aethercore service fallback id", d.FallbackServiceID); err != nil {
			return err
		}
	}
	if d.ProviderCollateralDenom != "" {
		if err := validatePolicyID("aethercore service provider collateral denom", d.ProviderCollateralDenom); err != nil {
			return err
		}
	}
	if d.ProviderCollateralAmount != "" {
		if err := validateAmountString("aethercore service provider collateral amount", d.ProviderCollateralAmount); err != nil {
			return err
		}
	}
	if d.FaultPolicy != "" && !IsServiceFailureBehavior(d.FaultPolicy) {
		return fmt.Errorf("unknown aethercore service fault policy %q", d.FaultPolicy)
	}
	switch d.TrustModel {
	case ServiceTrustEconomicallySecured:
		if d.ProviderCollateralDenom == "" || d.ProviderCollateralAmount == "" {
			return errors.New("aethercore economically secured service requires provider collateral")
		}
	case ServiceTrustCryptographicallyVerifiable:
		if d.ProofFormat == "" {
			return errors.New("aethercore cryptographically verifiable service requires proof format")
		}
	case ServiceTrustHybridChallengeable:
		if d.ChallengeWindow == 0 {
			return errors.New("aethercore hybrid challengeable service requires challenge window")
		}
	}
	return nil
}

func (d ServiceDescriptor) validateRuntimeRules() error {
	switch d.ServiceType {
	case ServiceTypeOnChain:
		if d.Execution.Location != ServiceLocationModule && d.Execution.Location != ServiceLocationContract && d.Execution.Location != ServiceLocationApplicationZone {
			return errors.New("aethercore on-chain service must execute in module, contract, or application zone")
		}
		if !d.Execution.Deterministic {
			return errors.New("aethercore on-chain service execution must be deterministic")
		}
		if d.Verification.TrustModel != ServiceTrustConsensusExecuted || d.Verification.Model != ServiceVerificationConsensusReceipt {
			return errors.New("aethercore on-chain service requires consensus receipt verification")
		}
		for _, method := range d.Interface.Methods {
			if method.GasModel == "" {
				return errors.New("aethercore on-chain service method requires gas model")
			}
		}
	case ServiceTypeOffChain:
		if d.Execution.Location != ServiceLocationExternal && d.Execution.Location != ServiceLocationHybrid && d.Execution.Location != ServiceLocationProviderPool {
			return errors.New("aethercore off-chain service must use external, hybrid, or provider-pool location")
		}
		if d.Verification.TrustModel == ServiceTrustConsensusExecuted {
			return errors.New("aethercore off-chain service cannot use consensus-executed trust model")
		}
		if d.Verification.Model == ServiceVerificationAdvisory && !d.Verification.ResponseSigningRequired {
			return errors.New("aethercore off-chain service requires signed, proof-backed, or economically constrained results")
		}
		if d.Execution.ResultExpiry == 0 {
			return errors.New("aethercore off-chain service requires result expiry")
		}
	case ServiceTypeMixed:
		if d.Verification.ChallengeWindow == 0 && d.Execution.ChallengeWindow == 0 && d.Verification.FallbackServiceID == "" {
			return errors.New("aethercore mixed service requires challenge window or fallback service")
		}
		if d.Verification.TrustModel != ServiceTrustHybridChallengeable &&
			d.Verification.TrustModel != ServiceTrustEconomicallySecured &&
			d.Verification.TrustModel != ServiceTrustCryptographicallyVerifiable {
			return errors.New("aethercore mixed service requires challengeable, economic, or cryptographic trust model")
		}
	case ServiceTypeFogMarket:
		if d.Execution.Location != ServiceLocationProviderPool {
			return errors.New("aethercore fog market service requires provider-pool location")
		}
		if d.Discovery.ProviderRoot == "" {
			return errors.New("aethercore fog market service requires provider root")
		}
		if d.Verification.ProviderCollateralDenom == "" || d.Verification.ProviderCollateralAmount == "" {
			return errors.New("aethercore fog market service requires provider collateral")
		}
	}
	return nil
}

func ComputeServiceInterfaceHash(d ServiceInterfaceDescriptor) string {
	d = CanonicalServiceInterfaceDescriptor(d)
	parts := []string{
		"aetheris-aek-service-interface-v1",
		d.InterfaceID,
		d.InterfaceName,
		fmt.Sprint(d.Version),
		d.SchemaEncoding,
		d.AuthModel,
		d.PaymentModel,
		d.MetadataHash,
		fmt.Sprint(d.CreatedHeight),
		fmt.Sprint(len(d.Methods)),
	}
	for _, method := range d.Methods {
		parts = append(parts, ComputeServiceMethodHash(method))
	}
	parts = appendStringSliceParts(parts, "events", d.Events)
	parts = appendStringSliceParts(parts, "errors", d.Errors)
	return hashParts(parts...)
}

func ComputeServiceMethodHash(d ServiceMethodDescriptor) string {
	return hashParts(
		"aetheris-aek-service-method-v1",
		d.MethodID,
		d.Name,
		d.InputSchemaHash,
		d.OutputSchemaHash,
		string(d.ExecutionType),
		d.RequiredPaymentModel,
		d.GasModel,
		string(d.VerificationModel),
		fmt.Sprint(d.TimeoutHeightDelta),
		fmt.Sprint(d.IdempotencyRequired),
		fmt.Sprint(d.CallbackSupported),
		string(d.FailureBehavior),
	)
}

func ComputeServiceDescriptorHash(d ServiceDescriptor) string {
	d = CanonicalServiceDescriptor(d)
	return hashParts(
		"aetheris-aek-service-descriptor-v1",
		d.ServiceID,
		d.Owner,
		string(d.ServiceType),
		string(d.ZoneID),
		d.InterfaceID,
		d.EndpointKey,
		fmt.Sprint(d.Version),
		d.AvailabilityHash,
		fmt.Sprint(d.Enabled),
		string(d.Status),
		fmt.Sprint(d.ExpiryHeight),
		fmt.Sprint(d.CreatedHeight),
		fmt.Sprint(d.UpdatedHeight),
		ComputeServiceInterfaceHash(d.Interface),
		string(d.Execution.Location),
		d.Execution.Target,
		d.Execution.ModuleRoute,
		d.Execution.ContractAddress,
		d.Execution.Endpoint,
		d.Execution.ProviderPoolID,
		string(d.Execution.Mode),
		fmt.Sprint(d.Execution.Deterministic),
		string(d.Execution.FailureBehavior),
		fmt.Sprint(d.Execution.ResultExpiry),
		fmt.Sprint(d.Execution.ChallengeWindow),
		d.Discovery.ServiceName,
		d.Discovery.IdentityName,
		d.Discovery.ProviderRoot,
		d.Discovery.MetadataHash,
		d.Discovery.ExternalDescriptorHash,
		fmt.Sprint(d.Discovery.CacheExpiryHeight),
		d.Discovery.SignaturePolicy,
		string(d.Payment.SettlementMode),
		d.Payment.Denom,
		d.Payment.Amount,
		d.Payment.MaxAmount,
		string(d.Payment.PricingUnit),
		fmt.Sprint(d.Payment.EscrowRequired),
		d.Payment.EscrowID,
		d.Payment.MeterID,
		fmt.Sprint(d.Payment.ExpiryHeight),
		string(d.Storage.Model),
		string(d.Storage.StateRootType),
		d.Storage.CommitmentHash,
		fmt.Sprint(d.Storage.RetentionHeight),
		fmt.Sprint(d.Storage.ProofRequired),
		string(d.Verification.TrustModel),
		string(d.Verification.Model),
		d.Verification.ProofFormat,
		fmt.Sprint(d.Verification.RequestSigningRequired),
		fmt.Sprint(d.Verification.ResponseSigningRequired),
		fmt.Sprint(d.Verification.ChallengeWindow),
		d.Verification.FallbackServiceID,
		d.Verification.ProviderCollateralDenom,
		d.Verification.ProviderCollateralAmount,
		string(d.Verification.FaultPolicy),
	)
}

func ComputeServiceRoot(services []ServiceDescriptor) (string, error) {
	ordered := append([]ServiceDescriptor(nil), services...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].ServiceID < ordered[j].ServiceID
	})
	parts := []string{"aetheris-aek-services-root-v1", fmt.Sprint(len(ordered))}
	var previous string
	for i, service := range ordered {
		if err := service.Validate(); err != nil {
			return "", err
		}
		if i > 0 && previous >= service.ServiceID {
			return "", errors.New("aethercore services must be sorted canonically by service id")
		}
		parts = append(parts, ComputeServiceDescriptorHash(service))
		previous = service.ServiceID
	}
	return hashParts(parts...), nil
}

func cloneServiceDescriptors(descriptors []ServiceDescriptor) []ServiceDescriptor {
	out := make([]ServiceDescriptor, len(descriptors))
	for i, descriptor := range descriptors {
		out[i] = cloneServiceDescriptor(descriptor)
	}
	return out
}

func cloneServiceDescriptor(descriptor ServiceDescriptor) ServiceDescriptor {
	descriptor.Interface.Methods = cloneServiceMethods(descriptor.Interface.Methods)
	descriptor.Interface.Events = append([]string(nil), descriptor.Interface.Events...)
	descriptor.Interface.Errors = append([]string(nil), descriptor.Interface.Errors...)
	return descriptor
}

func cloneServiceMethods(methods []ServiceMethodDescriptor) []ServiceMethodDescriptor {
	out := make([]ServiceMethodDescriptor, len(methods))
	copy(out, methods)
	return out
}

func sortServiceMethods(methods []ServiceMethodDescriptor) {
	sort.SliceStable(methods, func(i, j int) bool {
		return methods[i].MethodID < methods[j].MethodID
	})
}

func validateServiceMethods(methods []ServiceMethodDescriptor) error {
	var previous string
	seenIDs := make(map[string]struct{}, len(methods))
	seenNames := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seenIDs[method.MethodID]; found {
			return fmt.Errorf("duplicate aethercore service method id %s", method.MethodID)
		}
		seenIDs[method.MethodID] = struct{}{}
		if _, found := seenNames[method.Name]; found {
			return fmt.Errorf("duplicate aethercore service method name %s", method.Name)
		}
		seenNames[method.Name] = struct{}{}
		if i > 0 && previous >= method.MethodID {
			return errors.New("aethercore service methods must be sorted canonically")
		}
		previous = method.MethodID
	}
	return nil
}

func validateOptionalHash(fieldName, value string) error {
	if value == "" {
		return nil
	}
	return ValidateHash(fieldName, value)
}

func validateAmountString(fieldName, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return fmt.Errorf("%s must contain only decimal digits", fieldName)
		}
	}
	return nil
}

func appendStringSliceParts(parts []string, label string, values []string) []string {
	parts = append(parts, label, fmt.Sprint(len(values)))
	for _, value := range values {
		parts = append(parts, value)
	}
	return parts
}
