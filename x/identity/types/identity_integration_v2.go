package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type IdentityAccessPathV2 string

const (
	IdentityAccessPathCrossZoneMessage	IdentityAccessPathV2	= "verified_cross_zone_message"
	IdentityAccessPathProofQuery		IdentityAccessPathV2	= "proof_query"
	IdentityAccessPathVerifiedCache		IdentityAccessPathV2	= "cached_verified_resolver_record"
	IdentityAccessPathPreSigning		IdentityAccessPathV2	= "pre_signing_client_side_resolution"
)

type IdentityAccessPathDescriptorV2 struct {
	Path			IdentityAccessPathV2
	UseCase			string
	ConsensusRequirement	string
}

type IdentityLookupExecutionPlanV2 struct {
	Request			MsgResolveIdentity
	IdentityZoneID		string
	ReadOnly		bool
	MutatesIdentityState	bool
	ReplyToZoneID		string
	RequiresProof		bool
	ExecuteByHeight		uint64
	RequestReceiptRoot	string
	PlanHash		string
}

type IdentityAsyncResolutionEnvelopeV2 struct {
	Request			MsgResolveIdentity
	Result			MsgIdentityResolutionResult
	SourceZoneID		string
	DestinationZoneID	string
	ReplyTo			string
	RequestReceiptRoot	string
	ResultReceiptRoot	string
	EnvelopeHash		string
}

type IdentityVerifiedResolverCacheEntryV2 struct {
	Name			string
	NameHash		string
	TargetType		IdentityLookupTargetType
	ResolvedValueHash	string
	ResolverRecordVersion	uint64
	ProofHeight		uint64
	ExpiryHeight		uint64
	ProofHash		string
	TrustedAppHash		string
	InvalidationTriggers	[]IdentityCacheInvalidationTriggerV2
	EntryHash		string
}

type IdentityCacheUseV2 struct {
	Height			uint64
	ResolverRecordVersion	uint64
	ProofVerified		bool
	InvalidationTrigger	IdentityCacheInvalidationTriggerV2
}

type IdentityPreSigningResolutionBindingV2 struct {
	Name			string
	NameHash		string
	TargetType		IdentityLookupTargetType
	ResolvedValueHash	string
	TxPayloadHash		string
	BoundPayloadHash	string
}

type IdentityContractResolutionUseV2 struct {
	Result			MsgIdentityResolutionResult
	CurrentHeight		uint64
	ProofVerified		bool
	MessageOriginZoneID	string
	ReceiptRoot		string
	ReceiptHash		string
}

func IdentityIntegrationAccessPathsV2() []IdentityAccessPathDescriptorV2 {
	return []IdentityAccessPathDescriptorV2{
		{Path: IdentityAccessPathCrossZoneMessage, UseCase: "runtime resolution from another zone or contract", ConsensusRequirement: "request and reply messages are committed with receipts"},
		{Path: IdentityAccessPathProofQuery, UseCase: "light clients, contracts, and modules verifying identity state", ConsensusRequirement: "proof verifies against Identity Zone roots and resolver proof roots"},
		{Path: IdentityAccessPathVerifiedCache, UseCase: "repeated reads within another zone", ConsensusRequirement: "cache entries include proof height, expiry, and invalidation rules"},
		{Path: IdentityAccessPathPreSigning, UseCase: "wallet send-by-name or invoke-by-name UX", ConsensusRequirement: "client binds resolved value into transaction or message payload"},
	}
}

func NewIdentityLookupExecutionPlanV2(request MsgResolveIdentity, currentHeight uint64, requestReceiptRoot string) (IdentityLookupExecutionPlanV2, error) {
	if err := request.Validate(); err != nil {
		return IdentityLookupExecutionPlanV2{}, err
	}
	if currentHeight == 0 {
		return IdentityLookupExecutionPlanV2{}, errors.New("identity lookup execution current height must be positive")
	}
	if currentHeight > request.ExpiryHeight {
		return IdentityLookupExecutionPlanV2{}, errors.New("identity lookup execution request expired")
	}
	if requestReceiptRoot != "" {
		if err := validateHexHash("identity lookup request receipt root", requestReceiptRoot); err != nil {
			return IdentityLookupExecutionPlanV2{}, err
		}
	}
	plan := IdentityLookupExecutionPlanV2{
		Request:		request,
		IdentityZoneID:		IdentityZoneID,
		ReadOnly:		true,
		ReplyToZoneID:		request.SourceZoneID,
		RequiresProof:		request.ProofRequired,
		ExecuteByHeight:	request.ExpiryHeight,
		RequestReceiptRoot:	requestReceiptRoot,
	}
	plan.PlanHash = ComputeIdentityLookupExecutionPlanHashV2(plan)
	return plan, plan.Validate()
}

func NewIdentityAsyncResolutionEnvelopeV2(request MsgResolveIdentity, result MsgIdentityResolutionResult, requestReceiptRoot string, resultReceiptRoot string) (IdentityAsyncResolutionEnvelopeV2, error) {
	if err := request.Validate(); err != nil {
		return IdentityAsyncResolutionEnvelopeV2{}, err
	}
	if err := result.Validate(); err != nil {
		return IdentityAsyncResolutionEnvelopeV2{}, err
	}
	if request.RequestID != result.RequestID {
		return IdentityAsyncResolutionEnvelopeV2{}, errors.New("identity async resolution request id mismatch")
	}
	if request.ProofRequired && result.Status == IdentityResolutionStatusResolved && result.ProofHashOptional == "" {
		return IdentityAsyncResolutionEnvelopeV2{}, errors.New("identity async resolution proof is required by request")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"identity async resolution request receipt root", requestReceiptRoot},
		{"identity async resolution result receipt root", resultReceiptRoot},
	} {
		if item.value != "" {
			if err := validateHexHash(item.name, item.value); err != nil {
				return IdentityAsyncResolutionEnvelopeV2{}, err
			}
		}
	}
	envelope := IdentityAsyncResolutionEnvelopeV2{
		Request:		request,
		Result:			result,
		SourceZoneID:		IdentityZoneID,
		DestinationZoneID:	request.SourceZoneID,
		ReplyTo:		request.ReplyTo,
		RequestReceiptRoot:	requestReceiptRoot,
		ResultReceiptRoot:	resultReceiptRoot,
	}
	envelope.EnvelopeHash = ComputeIdentityAsyncResolutionEnvelopeHashV2(envelope)
	return envelope, envelope.Validate()
}

func NewIdentityVerifiedResolverCacheEntryV2(result MsgIdentityResolutionResult, proofHeight uint64, trustedAppHash string, invalidations []IdentityCacheInvalidationTriggerV2) (IdentityVerifiedResolverCacheEntryV2, error) {
	if err := result.Validate(); err != nil {
		return IdentityVerifiedResolverCacheEntryV2{}, err
	}
	if result.Status != IdentityResolutionStatusResolved {
		return IdentityVerifiedResolverCacheEntryV2{}, errors.New("identity verified cache requires resolved result")
	}
	if proofHeight == 0 {
		return IdentityVerifiedResolverCacheEntryV2{}, errors.New("identity verified cache proof height must be positive")
	}
	if proofHeight > result.ExpiryHeight {
		return IdentityVerifiedResolverCacheEntryV2{}, errors.New("identity verified cache proof height exceeds result expiry")
	}
	if err := validateHexHash("identity verified cache trusted app hash", trustedAppHash); err != nil {
		return IdentityVerifiedResolverCacheEntryV2{}, err
	}
	if result.ProofHashOptional == "" {
		return IdentityVerifiedResolverCacheEntryV2{}, errors.New("identity verified cache requires proof hash")
	}
	nameHash, err := DomainRecordV2NameHash(result.Name)
	if err != nil {
		return IdentityVerifiedResolverCacheEntryV2{}, err
	}
	entry := IdentityVerifiedResolverCacheEntryV2{
		Name:			mustNormalizeIdentityName(result.Name),
		NameHash:		nameHash,
		TargetType:		result.TargetType,
		ResolvedValueHash:	identityHash("identity-v2-resolved-value", result.ResolvedValue),
		ResolverRecordVersion:	result.ResolverRecordVersion,
		ProofHeight:		proofHeight,
		ExpiryHeight:		result.ExpiryHeight,
		ProofHash:		result.ProofHashOptional,
		TrustedAppHash:		trustedAppHash,
		InvalidationTriggers:	normalizeIdentityInvalidationTriggersV2(invalidations),
	}
	entry.EntryHash = ComputeIdentityVerifiedResolverCacheEntryHashV2(entry)
	return entry, entry.Validate()
}

func NewIdentityPreSigningResolutionBindingV2(name string, targetType IdentityLookupTargetType, resolvedValue []byte, txPayloadHash string) (IdentityPreSigningResolutionBindingV2, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityPreSigningResolutionBindingV2{}, err
	}
	if !IsIdentityLookupTargetType(targetType) {
		return IdentityPreSigningResolutionBindingV2{}, fmt.Errorf("unknown identity lookup target type %q", targetType)
	}
	if len(resolvedValue) == 0 {
		return IdentityPreSigningResolutionBindingV2{}, errors.New("identity pre-signing binding resolved value is required")
	}
	if err := validateHexHash("identity pre-signing transaction payload hash", txPayloadHash); err != nil {
		return IdentityPreSigningResolutionBindingV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return IdentityPreSigningResolutionBindingV2{}, err
	}
	binding := IdentityPreSigningResolutionBindingV2{
		Name:			normalized,
		NameHash:		nameHash,
		TargetType:		targetType,
		ResolvedValueHash:	identityHash("identity-v2-pre-signing-resolved-value", hex.EncodeToString(resolvedValue)),
		TxPayloadHash:		txPayloadHash,
	}
	binding.BoundPayloadHash = ComputeIdentityPreSigningResolutionBindingHashV2(binding)
	return binding, binding.Validate()
}

func ValidateIdentityContractResolutionUseV2(use IdentityContractResolutionUseV2) error {
	if err := use.Result.Validate(); err != nil {
		return err
	}
	if use.CurrentHeight == 0 {
		return errors.New("identity contract resolution use height must be positive")
	}
	if use.CurrentHeight > use.Result.ExpiryHeight {
		return errors.New("identity contract resolution result expired")
	}
	if use.ProofVerified {
		return nil
	}
	if use.MessageOriginZoneID != IdentityZoneID {
		return errors.New("identity contract resolution requires verified proof or Identity Zone origin")
	}
	if err := validateHexHash("identity contract resolution receipt root", use.ReceiptRoot); err != nil {
		return err
	}
	if err := validateHexHash("identity contract resolution receipt hash", use.ReceiptHash); err != nil {
		return err
	}
	return nil
}

func (p IdentityLookupExecutionPlanV2) Validate() error {
	if err := p.Request.Validate(); err != nil {
		return err
	}
	if p.IdentityZoneID != IdentityZoneID {
		return errors.New("identity lookup plan must execute in Identity Zone")
	}
	if !p.ReadOnly || p.MutatesIdentityState {
		return errors.New("identity lookup plan must be read-only")
	}
	if p.ReplyToZoneID != p.Request.SourceZoneID {
		return errors.New("identity lookup plan reply zone mismatch")
	}
	if p.ExecuteByHeight != p.Request.ExpiryHeight {
		return errors.New("identity lookup plan expiry mismatch")
	}
	if p.RequestReceiptRoot != "" {
		if err := validateHexHash("identity lookup plan request receipt root", p.RequestReceiptRoot); err != nil {
			return err
		}
	}
	if err := validateHexHash("identity lookup plan hash", p.PlanHash); err != nil {
		return err
	}
	if p.PlanHash != ComputeIdentityLookupExecutionPlanHashV2(p) {
		return errors.New("identity lookup plan hash mismatch")
	}
	return nil
}

func (e IdentityAsyncResolutionEnvelopeV2) Validate() error {
	if err := e.Request.Validate(); err != nil {
		return err
	}
	if err := e.Result.Validate(); err != nil {
		return err
	}
	if e.SourceZoneID != IdentityZoneID {
		return errors.New("identity async resolution source must be Identity Zone")
	}
	if e.DestinationZoneID != e.Request.SourceZoneID || e.ReplyTo != e.Request.ReplyTo {
		return errors.New("identity async resolution destination mismatch")
	}
	if e.Request.RequestID != e.Result.RequestID {
		return errors.New("identity async resolution request id mismatch")
	}
	if e.RequestReceiptRoot != "" {
		if err := validateHexHash("identity async request receipt root", e.RequestReceiptRoot); err != nil {
			return err
		}
	}
	if e.ResultReceiptRoot != "" {
		if err := validateHexHash("identity async result receipt root", e.ResultReceiptRoot); err != nil {
			return err
		}
	}
	if err := validateHexHash("identity async resolution envelope hash", e.EnvelopeHash); err != nil {
		return err
	}
	if e.EnvelopeHash != ComputeIdentityAsyncResolutionEnvelopeHashV2(e) {
		return errors.New("identity async resolution envelope hash mismatch")
	}
	return nil
}

func (e IdentityVerifiedResolverCacheEntryV2) Validate() error {
	normalized, err := NormalizeAETDomain(e.Name)
	if err != nil {
		return err
	}
	if e.Name != normalized {
		return errors.New("identity verified cache name must be normalized")
	}
	if err := validateHexHash("identity verified cache name hash", e.NameHash); err != nil {
		return err
	}
	expectedNameHash, err := DomainRecordV2NameHash(e.Name)
	if err != nil {
		return err
	}
	if e.NameHash != expectedNameHash {
		return errors.New("identity verified cache name hash mismatch")
	}
	if !IsIdentityLookupTargetType(e.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", e.TargetType)
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"identity verified cache resolved value hash", e.ResolvedValueHash},
		{"identity verified cache proof hash", e.ProofHash},
		{"identity verified cache trusted app hash", e.TrustedAppHash},
		{"identity verified cache entry hash", e.EntryHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if e.ResolverRecordVersion == 0 || e.ProofHeight == 0 || e.ExpiryHeight == 0 {
		return errors.New("identity verified cache version, proof height, and expiry are required")
	}
	if e.ProofHeight > e.ExpiryHeight {
		return errors.New("identity verified cache proof height exceeds expiry")
	}
	for _, trigger := range e.InvalidationTriggers {
		if !isIdentityCacheInvalidationTriggerV2(trigger) {
			return fmt.Errorf("unknown identity cache invalidation trigger %q", trigger)
		}
	}
	if e.EntryHash != ComputeIdentityVerifiedResolverCacheEntryHashV2(e) {
		return errors.New("identity verified cache entry hash mismatch")
	}
	return nil
}

func ValidateIdentityVerifiedResolverCacheUseV2(entry IdentityVerifiedResolverCacheEntryV2, use IdentityCacheUseV2) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	if use.Height == 0 {
		return errors.New("identity verified cache use height must be positive")
	}
	if use.Height > entry.ExpiryHeight {
		return errors.New("identity verified cache entry expired")
	}
	if use.ResolverRecordVersion != entry.ResolverRecordVersion {
		return errors.New("identity verified cache resolver version changed")
	}
	if !use.ProofVerified {
		return errors.New("identity verified cache use requires verified proof")
	}
	if use.InvalidationTrigger != "" {
		for _, trigger := range entry.InvalidationTriggers {
			if trigger == use.InvalidationTrigger {
				return errors.New("identity verified cache invalidated")
			}
		}
	}
	return nil
}

func (b IdentityPreSigningResolutionBindingV2) Validate() error {
	normalized, err := NormalizeAETDomain(b.Name)
	if err != nil {
		return err
	}
	if b.Name != normalized {
		return errors.New("identity pre-signing binding name must be normalized")
	}
	if !IsIdentityLookupTargetType(b.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", b.TargetType)
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"identity pre-signing name hash", b.NameHash},
		{"identity pre-signing resolved value hash", b.ResolvedValueHash},
		{"identity pre-signing tx payload hash", b.TxPayloadHash},
		{"identity pre-signing bound payload hash", b.BoundPayloadHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if expected, err := DomainRecordV2NameHash(b.Name); err != nil {
		return err
	} else if b.NameHash != expected {
		return errors.New("identity pre-signing name hash mismatch")
	}
	if b.BoundPayloadHash != ComputeIdentityPreSigningResolutionBindingHashV2(b) {
		return errors.New("identity pre-signing bound payload hash mismatch")
	}
	return nil
}

func ComputeIdentityLookupExecutionPlanHashV2(plan IdentityLookupExecutionPlanV2) string {
	return identityHash(
		"identity-v2-lookup-execution-plan",
		plan.Request.MessageHash,
		plan.IdentityZoneID,
		fmt.Sprintf("%t", plan.ReadOnly),
		fmt.Sprintf("%t", plan.MutatesIdentityState),
		plan.ReplyToZoneID,
		fmt.Sprintf("%t", plan.RequiresProof),
		fmt.Sprintf("%020d", plan.ExecuteByHeight),
		plan.RequestReceiptRoot,
	)
}

func ComputeIdentityAsyncResolutionEnvelopeHashV2(envelope IdentityAsyncResolutionEnvelopeV2) string {
	return identityHash(
		"identity-v2-async-resolution-envelope",
		envelope.Request.MessageHash,
		envelope.Result.ResultHash,
		envelope.SourceZoneID,
		envelope.DestinationZoneID,
		envelope.ReplyTo,
		envelope.RequestReceiptRoot,
		envelope.ResultReceiptRoot,
	)
}

func ComputeIdentityVerifiedResolverCacheEntryHashV2(entry IdentityVerifiedResolverCacheEntryV2) string {
	parts := []string{
		"identity-v2-verified-resolver-cache",
		entry.Name,
		entry.NameHash,
		string(entry.TargetType),
		entry.ResolvedValueHash,
		fmt.Sprintf("%020d", entry.ResolverRecordVersion),
		fmt.Sprintf("%020d", entry.ProofHeight),
		fmt.Sprintf("%020d", entry.ExpiryHeight),
		entry.ProofHash,
		entry.TrustedAppHash,
	}
	for _, trigger := range normalizeIdentityInvalidationTriggersV2(entry.InvalidationTriggers) {
		parts = append(parts, string(trigger))
	}
	return identityHash(parts...)
}

func ComputeIdentityPreSigningResolutionBindingHashV2(binding IdentityPreSigningResolutionBindingV2) string {
	return identityHash(
		"identity-v2-pre-signing-resolution-binding",
		binding.Name,
		binding.NameHash,
		string(binding.TargetType),
		binding.ResolvedValueHash,
		binding.TxPayloadHash,
	)
}

func normalizeIdentityInvalidationTriggersV2(triggers []IdentityCacheInvalidationTriggerV2) []IdentityCacheInvalidationTriggerV2 {
	out := append([]IdentityCacheInvalidationTriggerV2(nil), triggers...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	unique := make([]IdentityCacheInvalidationTriggerV2, 0, len(out))
	var previous IdentityCacheInvalidationTriggerV2
	for i, trigger := range out {
		if i > 0 && trigger == previous {
			continue
		}
		unique = append(unique, trigger)
		previous = trigger
	}
	return unique
}

func isIdentityCacheInvalidationTriggerV2(trigger IdentityCacheInvalidationTriggerV2) bool {
	switch trigger {
	case IdentityCacheInvalidDomainTransferV2, IdentityCacheInvalidResolverUpdateV2, IdentityCacheInvalidNFTBindingUpdateV2,
		IdentityCacheInvalidDomainExpiryV2, IdentityCacheInvalidRenewalEpochV2, IdentityCacheInvalidDelegationUpdateV2,
		IdentityCacheInvalidZonePolicyUpdateV2, IdentityCacheInvalidReverseUpdateV2:
		return true
	default:
		return false
	}
}

func mustNormalizeIdentityName(name string) string {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(name))
	}
	return normalized
}

func IdentityProofHashFromUniversalEnvelopeV2(proof coretypes.UniversalProofEnvelope) (string, error) {
	if err := proof.ValidateFormat(); err != nil {
		return "", err
	}
	if proof.ZoneID != "" && proof.ZoneID != coretypes.ZoneIDIdentity {
		return "", errors.New("identity universal proof must be scoped to Identity Zone")
	}
	return proof.ProofHash, nil
}
