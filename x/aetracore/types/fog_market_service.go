package types

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

type FogServiceCategory string
type FogProviderStatus string
type FogPricingUnit string
type FogSelectionStrategy string
type FogDisputeStatus string

const (
	FogCategoryCompute	FogServiceCategory	= "COMPUTE_PROVIDER"
	FogCategoryStorage	FogServiceCategory	= "STORAGE_PROVIDER"
	FogCategoryRouting	FogServiceCategory	= "ROUTING_PROVIDER"
	FogCategoryExecution	FogServiceCategory	= "EXECUTION_PROVIDER"
	FogCategoryIndexing	FogServiceCategory	= "INDEXING_PROVIDER"
	FogCategoryAvailability	FogServiceCategory	= "AVAILABILITY_PROVIDER"

	FogProviderActive	FogProviderStatus	= "ACTIVE"
	FogProviderSuspended	FogProviderStatus	= "SUSPENDED"
	FogProviderSlashed	FogProviderStatus	= "SLASHED"
	FogProviderExpired	FogProviderStatus	= "EXPIRED"

	FogPricingPerRequest		FogPricingUnit	= "REQUEST"
	FogPricingPerByte		FogPricingUnit	= "BYTE"
	FogPricingPerComputeUnit	FogPricingUnit	= "COMPUTE_UNIT"
	FogPricingPerStorageGiB		FogPricingUnit	= "STORAGE_GIB"
	FogPricingPerRoute		FogPricingUnit	= "ROUTE"

	FogSelectionDeterministicScore	FogSelectionStrategy	= "DETERMINISTIC_SCORE"
	FogSelectionLowestPrice		FogSelectionStrategy	= "LOWEST_PRICE"
	FogSelectionReputationWeighted	FogSelectionStrategy	= "REPUTATION_WEIGHTED"

	FogDisputeOpen		FogDisputeStatus	= "OPEN"
	FogDisputeRejected	FogDisputeStatus	= "REJECTED"
	FogDisputeProven	FogDisputeStatus	= "PROVEN"
)

type FogProviderRegistry struct {
	ServiceID		string
	ProviderPoolID		string
	CollateralDenom		string
	MinCollateralAmount	string
	Providers		[]FogProviderRecord
	ReputationEvents	[]FogReputationEvent
	Disputes		[]FogProviderDispute
	Slashes			[]FogProviderSlash
	RegistryHash		string
}

type FogProviderRecord struct {
	ProviderID		string
	IdentityKey		string
	Category		FogServiceCategory
	Pricing			FogProviderPricing
	ReputationScore		uint64
	CollateralDenom		string
	CollateralAmount	string
	StakeAmount		string
	AvailabilityCommitment	FogAvailabilityCommitment
	SupportedInterfaces	[]string
	Status			FogProviderStatus
	RegisteredHeight	uint64
	UpdatedHeight		uint64
	ExpiryHeight		uint64
	ProviderHash		string
}

type FogProviderPricing struct {
	Denom		string
	Amount		string
	MaxAmount	string
	Unit		FogPricingUnit
	ModelHash	string
}

type FogAvailabilityCommitment struct {
	CommitmentHash	string
	EndpointHash	string
	WindowStart	uint64
	WindowEnd	uint64
	UptimeTargetBps	uint32
	RenewalNonce	uint64
	SignatureHash	string
}

type FogReputationEvent struct {
	EventID		string
	ProviderID	string
	Height		uint64
	Successes	uint64
	Failures	uint64
	ScoreDelta	int64
	Reason		string
	EventHash	string
}

type FogProviderSelectionPolicy struct {
	Category		FogServiceCategory
	RequiredInterface	string
	Strategy		FogSelectionStrategy
	MinReputation		uint64
	MaxPriceAmount		string
	Limit			uint32
	SelectionNonce		string
}

type FogProviderSelection struct {
	Policy		FogProviderSelectionPolicy
	ProviderIDs	[]string
	SelectionHash	string
}

type FogProviderDispute struct {
	DisputeID	string
	ProviderID	string
	ServiceID	string
	Challenger	string
	FaultClass	MixedServiceFaultClass
	EvidenceHash	string
	OpenedHeight	uint64
	ResolveByHeight	uint64
	Status		FogDisputeStatus
	DisputeHash	string
}

type FogProviderSlash struct {
	SlashID		string
	DisputeID	string
	ProviderID	string
	ServiceID	string
	FaultClass	MixedServiceFaultClass
	PenaltyDenom	string
	PenaltyAmount	string
	Recipient	string
	SlashedHeight	uint64
	ReputationDelta	int64
	SlashHash	string
}

func NewFogProviderRegistry(descriptor ServiceDescriptor) (FogProviderRegistry, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return FogProviderRegistry{}, err
	}
	if descriptor.ServiceType != ServiceTypeFogMarket {
		return FogProviderRegistry{}, errors.New("aetracore fog registry requires fog market service descriptor")
	}
	registry := FogProviderRegistry{
		ServiceID:		descriptor.ServiceID,
		ProviderPoolID:		descriptor.Execution.ProviderPoolID,
		CollateralDenom:	descriptor.Verification.ProviderCollateralDenom,
		MinCollateralAmount:	descriptor.Verification.ProviderCollateralAmount,
		Providers:		[]FogProviderRecord{},
		ReputationEvents:	[]FogReputationEvent{},
		Disputes:		[]FogProviderDispute{},
		Slashes:		[]FogProviderSlash{},
	}
	registry.RegistryHash = ComputeFogProviderRegistryHash(registry)
	return registry, registry.Validate()
}

func RegisterFogProvider(registry FogProviderRegistry, provider FogProviderRecord) (FogProviderRegistry, FogProviderRecord, error) {
	if err := registry.Validate(); err != nil {
		return FogProviderRegistry{}, FogProviderRecord{}, err
	}
	provider = CanonicalFogProviderRecord(provider)
	if provider.CollateralDenom == "" {
		provider.CollateralDenom = registry.CollateralDenom
	}
	if provider.StakeAmount == "" {
		provider.StakeAmount = provider.CollateralAmount
	}
	provider.ProviderHash = ComputeFogProviderHash(provider)
	if err := provider.ValidateForRegistry(registry); err != nil {
		return FogProviderRegistry{}, FogProviderRecord{}, err
	}
	if _, found := registry.ProviderByID(provider.ProviderID); found {
		return FogProviderRegistry{}, FogProviderRecord{}, fmt.Errorf("aetracore fog provider %s already registered", provider.ProviderID)
	}
	next := registry.clone()
	next.Providers = append(next.Providers, provider)
	sortFogProviders(next.Providers)
	next.RegistryHash = ComputeFogProviderRegistryHash(next)
	return next, provider, next.Validate()
}

func RenewFogProvider(registry FogProviderRegistry, providerID string, commitment FogAvailabilityCommitment, expiryHeight uint64, updatedHeight uint64) (FogProviderRegistry, FogProviderRecord, error) {
	if err := registry.Validate(); err != nil {
		return FogProviderRegistry{}, FogProviderRecord{}, err
	}
	provider, index, found := registry.providerByIDWithIndex(providerID)
	if !found {
		return FogProviderRegistry{}, FogProviderRecord{}, fmt.Errorf("aetracore fog provider %s not found", providerID)
	}
	if updatedHeight == 0 || updatedHeight < provider.UpdatedHeight {
		return FogProviderRegistry{}, FogProviderRecord{}, errors.New("aetracore fog provider renewal height is invalid")
	}
	if expiryHeight <= provider.ExpiryHeight {
		return FogProviderRegistry{}, FogProviderRecord{}, errors.New("aetracore fog provider renewal must extend expiry")
	}
	commitment = CanonicalFogAvailabilityCommitment(commitment)
	commitment.CommitmentHash = ComputeFogAvailabilityCommitmentHash(commitment)
	provider.AvailabilityCommitment = commitment
	provider.UpdatedHeight = updatedHeight
	provider.ExpiryHeight = expiryHeight
	provider.Status = FogProviderActive
	provider.ProviderHash = ComputeFogProviderHash(provider)
	if err := provider.ValidateForRegistry(registry); err != nil {
		return FogProviderRegistry{}, FogProviderRecord{}, err
	}
	next := registry.clone()
	next.Providers[index] = provider
	sortFogProviders(next.Providers)
	next.RegistryHash = ComputeFogProviderRegistryHash(next)
	return next, provider, next.Validate()
}

func UpdateFogProviderReputation(registry FogProviderRegistry, event FogReputationEvent) (FogProviderRegistry, FogProviderRecord, error) {
	if err := registry.Validate(); err != nil {
		return FogProviderRegistry{}, FogProviderRecord{}, err
	}
	provider, index, found := registry.providerByIDWithIndex(event.ProviderID)
	if !found {
		return FogProviderRegistry{}, FogProviderRecord{}, fmt.Errorf("aetracore fog provider %s not found", event.ProviderID)
	}
	event = CanonicalFogReputationEvent(event)
	event.EventID = ComputeFogReputationEventID(event)
	event.EventHash = ComputeFogReputationEventHash(event)
	if err := event.Validate(); err != nil {
		return FogProviderRegistry{}, FogProviderRecord{}, err
	}
	provider.ReputationScore = applyFogReputationDelta(provider.ReputationScore, event.ScoreDelta)
	provider.UpdatedHeight = event.Height
	provider.ProviderHash = ComputeFogProviderHash(provider)
	next := registry.clone()
	next.Providers[index] = provider
	next.ReputationEvents = append(next.ReputationEvents, event)
	sortFogProviders(next.Providers)
	sortFogReputationEvents(next.ReputationEvents)
	next.RegistryHash = ComputeFogProviderRegistryHash(next)
	return next, provider, next.Validate()
}

func SelectFogProviders(registry FogProviderRegistry, policy FogProviderSelectionPolicy, currentHeight uint64) (FogProviderSelection, error) {
	if err := registry.Validate(); err != nil {
		return FogProviderSelection{}, err
	}
	policy = CanonicalFogProviderSelectionPolicy(policy)
	if err := policy.Validate(); err != nil {
		return FogProviderSelection{}, err
	}
	if currentHeight == 0 {
		return FogProviderSelection{}, errors.New("aetracore fog provider selection height must be positive")
	}
	candidates := make([]FogProviderRecord, 0, len(registry.Providers))
	for _, provider := range registry.Providers {
		if !provider.Selectable(policy, currentHeight) {
			continue
		}
		candidates = append(candidates, provider)
	}
	sortFogCandidates(candidates, policy)
	limit := int(policy.Limit)
	if limit == 0 || limit > len(candidates) {
		limit = len(candidates)
	}
	ids := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		ids = append(ids, candidates[i].ProviderID)
	}
	selection := FogProviderSelection{Policy: policy, ProviderIDs: ids}
	selection.SelectionHash = ComputeFogProviderSelectionHash(selection)
	return selection, selection.Validate()
}

func OpenFogProviderDispute(registry FogProviderRegistry, dispute FogProviderDispute) (FogProviderRegistry, FogProviderDispute, error) {
	if err := registry.Validate(); err != nil {
		return FogProviderRegistry{}, FogProviderDispute{}, err
	}
	if _, found := registry.ProviderByID(dispute.ProviderID); !found {
		return FogProviderRegistry{}, FogProviderDispute{}, fmt.Errorf("aetracore fog provider %s not found", dispute.ProviderID)
	}
	dispute = CanonicalFogProviderDispute(dispute)
	if dispute.ServiceID == "" {
		dispute.ServiceID = registry.ServiceID
	}
	dispute.Status = FogDisputeOpen
	dispute.DisputeID = ComputeFogProviderDisputeID(dispute)
	dispute.DisputeHash = ComputeFogProviderDisputeHash(dispute)
	if err := dispute.ValidateForRegistry(registry); err != nil {
		return FogProviderRegistry{}, FogProviderDispute{}, err
	}
	if _, found := registry.DisputeByID(dispute.DisputeID); found {
		return FogProviderRegistry{}, FogProviderDispute{}, fmt.Errorf("aetracore fog dispute %s already exists", dispute.DisputeID)
	}
	next := registry.clone()
	next.Disputes = append(next.Disputes, dispute)
	sortFogDisputes(next.Disputes)
	next.RegistryHash = ComputeFogProviderRegistryHash(next)
	return next, dispute, next.Validate()
}

func SlashFogProvider(registry FogProviderRegistry, disputeID string, proofAccepted bool, slashedHeight uint64) (FogProviderRegistry, FogProviderSlash, error) {
	if err := registry.Validate(); err != nil {
		return FogProviderRegistry{}, FogProviderSlash{}, err
	}
	dispute, disputeIndex, found := registry.disputeByIDWithIndex(disputeID)
	if !found {
		return FogProviderRegistry{}, FogProviderSlash{}, fmt.Errorf("aetracore fog dispute %s not found", disputeID)
	}
	if dispute.Status != FogDisputeOpen {
		return FogProviderRegistry{}, FogProviderSlash{}, errors.New("aetracore fog dispute is not open")
	}
	if slashedHeight == 0 || slashedHeight > dispute.ResolveByHeight {
		return FogProviderRegistry{}, FogProviderSlash{}, errors.New("aetracore fog slashing is outside dispute window")
	}
	provider, providerIndex, found := registry.providerByIDWithIndex(dispute.ProviderID)
	if !found {
		return FogProviderRegistry{}, FogProviderSlash{}, errors.New("aetracore fog dispute provider is not registered")
	}
	next := registry.clone()
	if !proofAccepted {
		next.Disputes[disputeIndex].Status = FogDisputeRejected
		next.Disputes[disputeIndex].DisputeHash = ComputeFogProviderDisputeHash(next.Disputes[disputeIndex])
		next.RegistryHash = ComputeFogProviderRegistryHash(next)
		return next, FogProviderSlash{}, next.Validate()
	}
	penalty := fogPenaltyAmount(registry.MinCollateralAmount, dispute.FaultClass)
	slash := FogProviderSlash{
		DisputeID:		dispute.DisputeID,
		ProviderID:		provider.ProviderID,
		ServiceID:		registry.ServiceID,
		FaultClass:		dispute.FaultClass,
		PenaltyDenom:		registry.CollateralDenom,
		PenaltyAmount:		penalty,
		Recipient:		dispute.Challenger,
		SlashedHeight:		slashedHeight,
		ReputationDelta:	fogSlashReputationDelta(dispute.FaultClass),
	}
	slash.SlashID = ComputeFogProviderSlashID(slash)
	slash.SlashHash = ComputeFogProviderSlashHash(slash)
	if err := slash.ValidateForRegistry(registry); err != nil {
		return FogProviderRegistry{}, FogProviderSlash{}, err
	}
	if err := mixedAmountAtLeast("aetracore fog provider collateral", provider.CollateralAmount, slash.PenaltyAmount); err != nil {
		return FogProviderRegistry{}, FogProviderSlash{}, err
	}
	provider.ReputationScore = applyFogReputationDelta(provider.ReputationScore, slash.ReputationDelta)
	provider.Status = FogProviderSlashed
	provider.UpdatedHeight = slashedHeight
	provider.ProviderHash = ComputeFogProviderHash(provider)
	next.Providers[providerIndex] = provider
	next.Disputes[disputeIndex].Status = FogDisputeProven
	next.Disputes[disputeIndex].DisputeHash = ComputeFogProviderDisputeHash(next.Disputes[disputeIndex])
	next.Slashes = append(next.Slashes, slash)
	sortFogProviders(next.Providers)
	sortFogDisputes(next.Disputes)
	sortFogSlashes(next.Slashes)
	next.RegistryHash = ComputeFogProviderRegistryHash(next)
	return next, slash, next.Validate()
}

func CanonicalFogProviderRecord(provider FogProviderRecord) FogProviderRecord {
	provider.ProviderID = strings.TrimSpace(provider.ProviderID)
	provider.IdentityKey = strings.TrimSpace(provider.IdentityKey)
	provider.CollateralDenom = strings.TrimSpace(provider.CollateralDenom)
	provider.CollateralAmount = strings.TrimSpace(provider.CollateralAmount)
	provider.StakeAmount = strings.TrimSpace(provider.StakeAmount)
	provider.SupportedInterfaces = append([]string(nil), provider.SupportedInterfaces...)
	sort.Strings(provider.SupportedInterfaces)
	provider.Pricing = CanonicalFogProviderPricing(provider.Pricing)
	provider.AvailabilityCommitment = CanonicalFogAvailabilityCommitment(provider.AvailabilityCommitment)
	return provider
}

func (provider FogProviderRecord) ValidateForRegistry(registry FogProviderRegistry) error {
	if err := provider.Validate(); err != nil {
		return err
	}
	if provider.CollateralDenom != registry.CollateralDenom {
		return errors.New("aetracore fog provider collateral denom mismatch")
	}
	return mixedAmountAtLeast("aetracore fog provider collateral", provider.CollateralAmount, registry.MinCollateralAmount)
}

func (provider FogProviderRecord) Validate() error {
	if err := validatePolicyID("aetracore fog provider id", provider.ProviderID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog provider identity key", provider.IdentityKey); err != nil {
		return err
	}
	if !IsFogServiceCategory(provider.Category) {
		return fmt.Errorf("unknown aetracore fog service category %q", provider.Category)
	}
	if err := provider.Pricing.Validate(); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog provider collateral denom", provider.CollateralDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore fog provider collateral amount", provider.CollateralAmount); err != nil {
		return err
	}
	if err := validateAmountString("aetracore fog provider stake amount", provider.StakeAmount); err != nil {
		return err
	}
	if err := provider.AvailabilityCommitment.Validate(); err != nil {
		return err
	}
	if err := validateSortedStringSet("aetracore fog provider supported interface", provider.SupportedInterfaces); err != nil {
		return err
	}
	if len(provider.SupportedInterfaces) == 0 {
		return errors.New("aetracore fog provider requires supported interfaces")
	}
	if !IsFogProviderStatus(provider.Status) {
		return fmt.Errorf("unknown aetracore fog provider status %q", provider.Status)
	}
	if provider.RegisteredHeight == 0 || provider.UpdatedHeight < provider.RegisteredHeight {
		return errors.New("aetracore fog provider heights are invalid")
	}
	if provider.ExpiryHeight <= provider.UpdatedHeight {
		return errors.New("aetracore fog provider expiry must exceed updated height")
	}
	if err := ValidateHash("aetracore fog provider hash", provider.ProviderHash); err != nil {
		return err
	}
	if expected := ComputeFogProviderHash(provider); provider.ProviderHash != expected {
		return fmt.Errorf("aetracore fog provider hash mismatch: expected %s", expected)
	}
	return nil
}

func CanonicalFogProviderPricing(pricing FogProviderPricing) FogProviderPricing {
	pricing.Denom = strings.TrimSpace(pricing.Denom)
	pricing.Amount = strings.TrimSpace(pricing.Amount)
	pricing.MaxAmount = strings.TrimSpace(pricing.MaxAmount)
	pricing.ModelHash = strings.ToLower(strings.TrimSpace(pricing.ModelHash))
	return pricing
}

func (pricing FogProviderPricing) Validate() error {
	if err := validatePolicyID("aetracore fog provider pricing denom", pricing.Denom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore fog provider pricing amount", pricing.Amount); err != nil {
		return err
	}
	if pricing.MaxAmount != "" {
		if err := validateAmountString("aetracore fog provider pricing max amount", pricing.MaxAmount); err != nil {
			return err
		}
		if err := mixedAmountAtLeast("aetracore fog provider pricing max amount", pricing.MaxAmount, pricing.Amount); err != nil {
			return err
		}
	}
	if !IsFogPricingUnit(pricing.Unit) {
		return fmt.Errorf("unknown aetracore fog pricing unit %q", pricing.Unit)
	}
	return validateOptionalHash("aetracore fog provider pricing model hash", pricing.ModelHash)
}

func CanonicalFogAvailabilityCommitment(commitment FogAvailabilityCommitment) FogAvailabilityCommitment {
	commitment.CommitmentHash = strings.ToLower(strings.TrimSpace(commitment.CommitmentHash))
	commitment.EndpointHash = strings.ToLower(strings.TrimSpace(commitment.EndpointHash))
	commitment.SignatureHash = strings.ToLower(strings.TrimSpace(commitment.SignatureHash))
	return commitment
}

func (commitment FogAvailabilityCommitment) Validate() error {
	if err := ValidateHash("aetracore fog availability commitment hash", commitment.CommitmentHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore fog availability endpoint hash", commitment.EndpointHash); err != nil {
		return err
	}
	if commitment.WindowStart == 0 || commitment.WindowEnd <= commitment.WindowStart {
		return errors.New("aetracore fog availability window is invalid")
	}
	if commitment.UptimeTargetBps == 0 || commitment.UptimeTargetBps > 10_000 {
		return errors.New("aetracore fog availability target must be 1..10000 bps")
	}
	if commitment.RenewalNonce == 0 {
		return errors.New("aetracore fog availability renewal nonce must be positive")
	}
	if err := ValidateHash("aetracore fog availability signature hash", commitment.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeFogAvailabilityCommitmentHash(commitment); commitment.CommitmentHash != expected {
		return fmt.Errorf("aetracore fog availability commitment hash mismatch: expected %s", expected)
	}
	return nil
}

func (registry FogProviderRegistry) Validate() error {
	if err := validatePolicyID("aetracore fog registry service id", registry.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog registry provider pool id", registry.ProviderPoolID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog registry collateral denom", registry.CollateralDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore fog registry min collateral", registry.MinCollateralAmount); err != nil {
		return err
	}
	if err := validateFogProviders(registry); err != nil {
		return err
	}
	if err := validateFogReputationEvents(registry.ReputationEvents); err != nil {
		return err
	}
	if err := validateFogDisputes(registry); err != nil {
		return err
	}
	if err := validateFogSlashes(registry); err != nil {
		return err
	}
	if err := ValidateHash("aetracore fog registry hash", registry.RegistryHash); err != nil {
		return err
	}
	if expected := ComputeFogProviderRegistryHash(registry); registry.RegistryHash != expected {
		return fmt.Errorf("aetracore fog registry hash mismatch: expected %s", expected)
	}
	return nil
}

func (registry FogProviderRegistry) ProviderByID(providerID string) (FogProviderRecord, bool) {
	provider, _, found := registry.providerByIDWithIndex(providerID)
	return provider, found
}

func (registry FogProviderRegistry) providerByIDWithIndex(providerID string) (FogProviderRecord, int, bool) {
	for i, provider := range registry.Providers {
		if provider.ProviderID == providerID {
			return provider, i, true
		}
	}
	return FogProviderRecord{}, 0, false
}

func (registry FogProviderRegistry) DisputeByID(disputeID string) (FogProviderDispute, bool) {
	dispute, _, found := registry.disputeByIDWithIndex(disputeID)
	return dispute, found
}

func (registry FogProviderRegistry) disputeByIDWithIndex(disputeID string) (FogProviderDispute, int, bool) {
	for i, dispute := range registry.Disputes {
		if dispute.DisputeID == disputeID {
			return dispute, i, true
		}
	}
	return FogProviderDispute{}, 0, false
}

func (registry FogProviderRegistry) clone() FogProviderRegistry {
	registry.Providers = append([]FogProviderRecord(nil), registry.Providers...)
	registry.ReputationEvents = append([]FogReputationEvent(nil), registry.ReputationEvents...)
	registry.Disputes = append([]FogProviderDispute(nil), registry.Disputes...)
	registry.Slashes = append([]FogProviderSlash(nil), registry.Slashes...)
	return registry
}

func (provider FogProviderRecord) Selectable(policy FogProviderSelectionPolicy, currentHeight uint64) bool {
	if provider.Status != FogProviderActive || provider.ExpiryHeight < currentHeight {
		return false
	}
	if provider.Category != policy.Category {
		return false
	}
	if provider.ReputationScore < policy.MinReputation {
		return false
	}
	if policy.RequiredInterface != "" && !fogProviderSupportsInterface(provider.SupportedInterfaces, policy.RequiredInterface) {
		return false
	}
	if policy.MaxPriceAmount != "" {
		if err := mixedAmountAtMost("aetracore fog provider price", provider.Pricing.Amount, policy.MaxPriceAmount); err != nil {
			return false
		}
	}
	return true
}

func CanonicalFogProviderSelectionPolicy(policy FogProviderSelectionPolicy) FogProviderSelectionPolicy {
	policy.RequiredInterface = strings.TrimSpace(policy.RequiredInterface)
	policy.MaxPriceAmount = strings.TrimSpace(policy.MaxPriceAmount)
	policy.SelectionNonce = strings.TrimSpace(policy.SelectionNonce)
	return policy
}

func (policy FogProviderSelectionPolicy) Validate() error {
	if !IsFogServiceCategory(policy.Category) {
		return fmt.Errorf("unknown aetracore fog selection category %q", policy.Category)
	}
	if policy.RequiredInterface != "" {
		if err := validatePolicyID("aetracore fog selection interface", policy.RequiredInterface); err != nil {
			return err
		}
	}
	if !IsFogSelectionStrategy(policy.Strategy) {
		return fmt.Errorf("unknown aetracore fog selection strategy %q", policy.Strategy)
	}
	if policy.MaxPriceAmount != "" {
		if err := validateAmountString("aetracore fog selection max price", policy.MaxPriceAmount); err != nil {
			return err
		}
	}
	if policy.Limit == 0 {
		return errors.New("aetracore fog selection limit must be positive")
	}
	return validatePolicyID("aetracore fog selection nonce", policy.SelectionNonce)
}

func (selection FogProviderSelection) Validate() error {
	if err := selection.Policy.Validate(); err != nil {
		return err
	}
	if len(selection.ProviderIDs) == 0 {
		return errors.New("aetracore fog selection requires providers")
	}
	if err := validateFogSelectionProviders(selection.ProviderIDs); err != nil {
		return err
	}
	if err := ValidateHash("aetracore fog selection hash", selection.SelectionHash); err != nil {
		return err
	}
	if expected := ComputeFogProviderSelectionHash(selection); selection.SelectionHash != expected {
		return fmt.Errorf("aetracore fog selection hash mismatch: expected %s", expected)
	}
	return nil
}

func CanonicalFogReputationEvent(event FogReputationEvent) FogReputationEvent {
	event.ProviderID = strings.TrimSpace(event.ProviderID)
	event.Reason = strings.TrimSpace(event.Reason)
	return event
}

func (event FogReputationEvent) Validate() error {
	if err := ValidateHash("aetracore fog reputation event id", event.EventID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog reputation provider id", event.ProviderID); err != nil {
		return err
	}
	if event.Height == 0 {
		return errors.New("aetracore fog reputation height must be positive")
	}
	if event.Successes == 0 && event.Failures == 0 && event.ScoreDelta == 0 {
		return errors.New("aetracore fog reputation event must change score or counters")
	}
	if err := validatePolicyID("aetracore fog reputation reason", event.Reason); err != nil {
		return err
	}
	if err := ValidateHash("aetracore fog reputation event hash", event.EventHash); err != nil {
		return err
	}
	if expected := ComputeFogReputationEventHash(event); event.EventHash != expected {
		return fmt.Errorf("aetracore fog reputation event hash mismatch: expected %s", expected)
	}
	return nil
}

func CanonicalFogProviderDispute(dispute FogProviderDispute) FogProviderDispute {
	dispute.ProviderID = strings.TrimSpace(dispute.ProviderID)
	dispute.ServiceID = strings.TrimSpace(dispute.ServiceID)
	dispute.Challenger = strings.TrimSpace(dispute.Challenger)
	dispute.EvidenceHash = strings.ToLower(strings.TrimSpace(dispute.EvidenceHash))
	return dispute
}

func (dispute FogProviderDispute) ValidateForRegistry(registry FogProviderRegistry) error {
	if err := dispute.Validate(); err != nil {
		return err
	}
	if dispute.ServiceID != registry.ServiceID {
		return errors.New("aetracore fog dispute service mismatch")
	}
	return nil
}

func (dispute FogProviderDispute) Validate() error {
	if err := ValidateHash("aetracore fog dispute id", dispute.DisputeID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog dispute provider id", dispute.ProviderID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog dispute service id", dispute.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog dispute challenger", dispute.Challenger); err != nil {
		return err
	}
	if !IsMixedFaultClass(dispute.FaultClass) {
		return fmt.Errorf("unknown aetracore fog dispute fault class %q", dispute.FaultClass)
	}
	if err := ValidateHash("aetracore fog dispute evidence hash", dispute.EvidenceHash); err != nil {
		return err
	}
	if dispute.OpenedHeight == 0 || dispute.ResolveByHeight <= dispute.OpenedHeight {
		return errors.New("aetracore fog dispute window is invalid")
	}
	if !IsFogDisputeStatus(dispute.Status) {
		return fmt.Errorf("unknown aetracore fog dispute status %q", dispute.Status)
	}
	if err := ValidateHash("aetracore fog dispute hash", dispute.DisputeHash); err != nil {
		return err
	}
	if expected := ComputeFogProviderDisputeHash(dispute); dispute.DisputeHash != expected {
		return fmt.Errorf("aetracore fog dispute hash mismatch: expected %s", expected)
	}
	return nil
}

func (slash FogProviderSlash) ValidateForRegistry(registry FogProviderRegistry) error {
	if err := slash.Validate(); err != nil {
		return err
	}
	if slash.ServiceID != registry.ServiceID {
		return errors.New("aetracore fog slash service mismatch")
	}
	if slash.PenaltyDenom != registry.CollateralDenom {
		return errors.New("aetracore fog slash collateral denom mismatch")
	}
	return nil
}

func (slash FogProviderSlash) Validate() error {
	if err := ValidateHash("aetracore fog slash id", slash.SlashID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore fog slash dispute id", slash.DisputeID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog slash provider id", slash.ProviderID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog slash service id", slash.ServiceID); err != nil {
		return err
	}
	if !IsMixedFaultClass(slash.FaultClass) {
		return fmt.Errorf("unknown aetracore fog slash fault class %q", slash.FaultClass)
	}
	if err := validatePolicyID("aetracore fog slash penalty denom", slash.PenaltyDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore fog slash penalty amount", slash.PenaltyAmount); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore fog slash recipient", slash.Recipient); err != nil {
		return err
	}
	if slash.SlashedHeight == 0 {
		return errors.New("aetracore fog slash height must be positive")
	}
	if slash.ReputationDelta >= 0 {
		return errors.New("aetracore fog slash reputation delta must be negative")
	}
	if err := ValidateHash("aetracore fog slash hash", slash.SlashHash); err != nil {
		return err
	}
	if expected := ComputeFogProviderSlashHash(slash); slash.SlashHash != expected {
		return fmt.Errorf("aetracore fog slash hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeFogProviderRegistryHash(registry FogProviderRegistry) string {
	parts := []string{
		"aetra-aek-fog-provider-registry-v1",
		registry.ServiceID,
		registry.ProviderPoolID,
		registry.CollateralDenom,
		registry.MinCollateralAmount,
		fmt.Sprint(len(registry.Providers)),
	}
	for _, provider := range registry.Providers {
		parts = append(parts, provider.ProviderHash)
	}
	parts = append(parts, fmt.Sprint(len(registry.ReputationEvents)))
	for _, event := range registry.ReputationEvents {
		parts = append(parts, event.EventHash)
	}
	parts = append(parts, fmt.Sprint(len(registry.Disputes)))
	for _, dispute := range registry.Disputes {
		parts = append(parts, dispute.DisputeHash)
	}
	parts = append(parts, fmt.Sprint(len(registry.Slashes)))
	for _, slash := range registry.Slashes {
		parts = append(parts, slash.SlashHash)
	}
	return hashParts(parts...)
}

func ComputeFogProviderHash(provider FogProviderRecord) string {
	parts := []string{
		"aetra-aek-fog-provider-v1",
		provider.ProviderID,
		provider.IdentityKey,
		string(provider.Category),
		ComputeFogProviderPricingHash(provider.Pricing),
		fmt.Sprint(provider.ReputationScore),
		provider.CollateralDenom,
		provider.CollateralAmount,
		provider.StakeAmount,
		provider.AvailabilityCommitment.CommitmentHash,
		string(provider.Status),
		fmt.Sprint(provider.RegisteredHeight),
		fmt.Sprint(provider.UpdatedHeight),
		fmt.Sprint(provider.ExpiryHeight),
	}
	parts = appendStringSliceParts(parts, "interfaces", provider.SupportedInterfaces)
	return hashParts(parts...)
}

func ComputeFogProviderPricingHash(pricing FogProviderPricing) string {
	return hashParts(
		"aetra-aek-fog-provider-pricing-v1",
		pricing.Denom,
		pricing.Amount,
		pricing.MaxAmount,
		string(pricing.Unit),
		pricing.ModelHash,
	)
}

func ComputeFogAvailabilityCommitmentHash(commitment FogAvailabilityCommitment) string {
	return hashParts(
		"aetra-aek-fog-availability-v1",
		commitment.EndpointHash,
		fmt.Sprint(commitment.WindowStart),
		fmt.Sprint(commitment.WindowEnd),
		fmt.Sprint(commitment.UptimeTargetBps),
		fmt.Sprint(commitment.RenewalNonce),
		commitment.SignatureHash,
	)
}

func ComputeFogReputationEventID(event FogReputationEvent) string {
	return hashParts("aetra-aek-fog-reputation-id-v1", event.ProviderID, fmt.Sprint(event.Height), event.Reason)
}

func ComputeFogReputationEventHash(event FogReputationEvent) string {
	return hashParts(
		"aetra-aek-fog-reputation-v1",
		event.EventID,
		event.ProviderID,
		fmt.Sprint(event.Height),
		fmt.Sprint(event.Successes),
		fmt.Sprint(event.Failures),
		fmt.Sprint(event.ScoreDelta),
		event.Reason,
	)
}

func ComputeFogProviderSelectionHash(selection FogProviderSelection) string {
	parts := []string{
		"aetra-aek-fog-selection-v1",
		string(selection.Policy.Category),
		selection.Policy.RequiredInterface,
		string(selection.Policy.Strategy),
		fmt.Sprint(selection.Policy.MinReputation),
		selection.Policy.MaxPriceAmount,
		fmt.Sprint(selection.Policy.Limit),
		selection.Policy.SelectionNonce,
	}
	parts = appendStringSliceParts(parts, "providers", selection.ProviderIDs)
	return hashParts(parts...)
}

func ComputeFogProviderDisputeID(dispute FogProviderDispute) string {
	return hashParts("aetra-aek-fog-dispute-id-v1", dispute.ProviderID, dispute.Challenger, dispute.EvidenceHash)
}

func ComputeFogProviderDisputeHash(dispute FogProviderDispute) string {
	return hashParts(
		"aetra-aek-fog-dispute-v1",
		dispute.DisputeID,
		dispute.ProviderID,
		dispute.ServiceID,
		dispute.Challenger,
		string(dispute.FaultClass),
		dispute.EvidenceHash,
		fmt.Sprint(dispute.OpenedHeight),
		fmt.Sprint(dispute.ResolveByHeight),
		string(dispute.Status),
	)
}

func ComputeFogProviderSlashID(slash FogProviderSlash) string {
	return hashParts("aetra-aek-fog-slash-id-v1", slash.DisputeID, slash.ProviderID, string(slash.FaultClass))
}

func ComputeFogProviderSlashHash(slash FogProviderSlash) string {
	return hashParts(
		"aetra-aek-fog-slash-v1",
		slash.SlashID,
		slash.DisputeID,
		slash.ProviderID,
		slash.ServiceID,
		string(slash.FaultClass),
		slash.PenaltyDenom,
		slash.PenaltyAmount,
		slash.Recipient,
		fmt.Sprint(slash.SlashedHeight),
		fmt.Sprint(slash.ReputationDelta),
	)
}

func IsFogServiceCategory(category FogServiceCategory) bool {
	switch category {
	case FogCategoryCompute, FogCategoryStorage, FogCategoryRouting, FogCategoryExecution, FogCategoryIndexing, FogCategoryAvailability:
		return true
	default:
		return false
	}
}

func IsFogProviderStatus(status FogProviderStatus) bool {
	switch status {
	case FogProviderActive, FogProviderSuspended, FogProviderSlashed, FogProviderExpired:
		return true
	default:
		return false
	}
}

func IsFogPricingUnit(unit FogPricingUnit) bool {
	switch unit {
	case FogPricingPerRequest, FogPricingPerByte, FogPricingPerComputeUnit, FogPricingPerStorageGiB, FogPricingPerRoute:
		return true
	default:
		return false
	}
}

func IsFogSelectionStrategy(strategy FogSelectionStrategy) bool {
	switch strategy {
	case FogSelectionDeterministicScore, FogSelectionLowestPrice, FogSelectionReputationWeighted:
		return true
	default:
		return false
	}
}

func IsFogDisputeStatus(status FogDisputeStatus) bool {
	switch status {
	case FogDisputeOpen, FogDisputeRejected, FogDisputeProven:
		return true
	default:
		return false
	}
}

func validateFogProviders(registry FogProviderRegistry) error {
	var previous string
	seen := make(map[string]struct{}, len(registry.Providers))
	for _, provider := range registry.Providers {
		if err := provider.ValidateForRegistry(registry); err != nil {
			return err
		}
		if _, found := seen[provider.ProviderID]; found {
			return fmt.Errorf("duplicate aetracore fog provider %s", provider.ProviderID)
		}
		seen[provider.ProviderID] = struct{}{}
		if previous != "" && previous >= provider.ProviderID {
			return errors.New("aetracore fog providers must be sorted canonically")
		}
		previous = provider.ProviderID
	}
	return nil
}

func validateFogReputationEvents(events []FogReputationEvent) error {
	var previous string
	seen := make(map[string]struct{}, len(events))
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := seen[event.EventID]; found {
			return fmt.Errorf("duplicate aetracore fog reputation event %s", event.EventID)
		}
		seen[event.EventID] = struct{}{}
		if previous != "" && previous >= event.EventID {
			return errors.New("aetracore fog reputation events must be sorted canonically")
		}
		previous = event.EventID
	}
	return nil
}

func validateFogDisputes(registry FogProviderRegistry) error {
	var previous string
	seen := make(map[string]struct{}, len(registry.Disputes))
	for _, dispute := range registry.Disputes {
		if err := dispute.ValidateForRegistry(registry); err != nil {
			return err
		}
		if _, found := seen[dispute.DisputeID]; found {
			return fmt.Errorf("duplicate aetracore fog dispute %s", dispute.DisputeID)
		}
		seen[dispute.DisputeID] = struct{}{}
		if previous != "" && previous >= dispute.DisputeID {
			return errors.New("aetracore fog disputes must be sorted canonically")
		}
		previous = dispute.DisputeID
	}
	return nil
}

func validateFogSlashes(registry FogProviderRegistry) error {
	var previous string
	seen := make(map[string]struct{}, len(registry.Slashes))
	for _, slash := range registry.Slashes {
		if err := slash.ValidateForRegistry(registry); err != nil {
			return err
		}
		if _, found := seen[slash.SlashID]; found {
			return fmt.Errorf("duplicate aetracore fog slash %s", slash.SlashID)
		}
		seen[slash.SlashID] = struct{}{}
		if previous != "" && previous >= slash.SlashID {
			return errors.New("aetracore fog slashes must be sorted canonically")
		}
		previous = slash.SlashID
	}
	return nil
}

func validateFogSelectionProviders(providerIDs []string) error {
	seen := make(map[string]struct{}, len(providerIDs))
	for _, providerID := range providerIDs {
		if err := validatePolicyID("aetracore fog selection provider", providerID); err != nil {
			return err
		}
		if _, found := seen[providerID]; found {
			return fmt.Errorf("duplicate aetracore fog selection provider %s", providerID)
		}
		seen[providerID] = struct{}{}
	}
	return nil
}

func fogProviderSupportsInterface(interfaces []string, required string) bool {
	for _, supported := range interfaces {
		if supported == required {
			return true
		}
	}
	return false
}

func sortFogProviders(providers []FogProviderRecord) {
	sort.SliceStable(providers, func(i, j int) bool { return providers[i].ProviderID < providers[j].ProviderID })
}

func sortFogReputationEvents(events []FogReputationEvent) {
	sort.SliceStable(events, func(i, j int) bool { return events[i].EventID < events[j].EventID })
}

func sortFogDisputes(disputes []FogProviderDispute) {
	sort.SliceStable(disputes, func(i, j int) bool { return disputes[i].DisputeID < disputes[j].DisputeID })
}

func sortFogSlashes(slashes []FogProviderSlash) {
	sort.SliceStable(slashes, func(i, j int) bool { return slashes[i].SlashID < slashes[j].SlashID })
}

func sortFogCandidates(candidates []FogProviderRecord, policy FogProviderSelectionPolicy) {
	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		switch policy.Strategy {
		case FogSelectionLowestPrice:
			if cmp, err := mixedCompareAmount(left.Pricing.Amount, right.Pricing.Amount); err == nil && cmp != 0 {
				return cmp < 0
			}
			if left.ReputationScore != right.ReputationScore {
				return left.ReputationScore > right.ReputationScore
			}
		case FogSelectionReputationWeighted:
			leftScore := left.ReputationScore * (left.AvailabilityScore() + 1)
			rightScore := right.ReputationScore * (right.AvailabilityScore() + 1)
			if leftScore != rightScore {
				return leftScore > rightScore
			}
		default:
			if left.ReputationScore != right.ReputationScore {
				return left.ReputationScore > right.ReputationScore
			}
		}
		return left.ProviderID < right.ProviderID
	})
}

func (provider FogProviderRecord) AvailabilityScore() uint64 {
	return uint64(provider.AvailabilityCommitment.UptimeTargetBps)
}

func applyFogReputationDelta(score uint64, delta int64) uint64 {
	if delta < 0 {
		sub := uint64(-delta)
		if sub > score {
			return 0
		}
		return score - sub
	}
	return score + uint64(delta)
}

func fogPenaltyAmount(minCollateral string, faultClass MixedServiceFaultClass) string {
	divisor := "4"
	switch faultClass {
	case MixedFaultLow:
		divisor = "10"
	case MixedFaultMedium:
		divisor = "4"
	case MixedFaultHigh:
		divisor = "2"
	case MixedFaultCritical:
		return minCollateral
	}
	value := decimalDivideCeil(minCollateral, divisor)
	if value == "0" {
		return "1"
	}
	return value
}

func fogSlashReputationDelta(faultClass MixedServiceFaultClass) int64 {
	switch faultClass {
	case MixedFaultLow:
		return -10
	case MixedFaultMedium:
		return -25
	case MixedFaultHigh:
		return -50
	case MixedFaultCritical:
		return -100
	default:
		return -25
	}
}

func decimalDivideCeil(value string, divisor string) string {
	left, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return "0"
	}
	right, ok := new(big.Int).SetString(divisor, 10)
	if !ok || right.Sign() == 0 {
		return "0"
	}
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(left, right, remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.String()
}
