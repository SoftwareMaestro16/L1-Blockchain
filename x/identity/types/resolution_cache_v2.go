package types

import (
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	IdentityCacheLayerFullNodeDirectV2	IdentityCacheLayerV2			= "full_node_direct"
	IdentityCacheLayerFullNodeRecursiveV2	IdentityCacheLayerV2			= "full_node_recursive"
	IdentityCacheLayerWalletVerifiedV2	IdentityCacheLayerV2			= "wallet_verified_proof"
	IdentityCacheLayerServiceEndpointV2	IdentityCacheLayerV2			= "service_endpoint_ttl"
	IdentityCacheLayerReverseVerifiedV2	IdentityCacheLayerV2			= "reverse_verified"
	IdentityCacheInvalidDomainTransferV2	IdentityCacheInvalidationTriggerV2	= "domain_transfer"
	IdentityCacheInvalidResolverUpdateV2	IdentityCacheInvalidationTriggerV2	= "resolver_update"
	IdentityCacheInvalidNFTBindingUpdateV2	IdentityCacheInvalidationTriggerV2	= "nft_binding_update"
	IdentityCacheInvalidDomainExpiryV2	IdentityCacheInvalidationTriggerV2	= "domain_expiry"
	IdentityCacheInvalidRenewalEpochV2	IdentityCacheInvalidationTriggerV2	= "renewal_lifecycle_epoch"
	IdentityCacheInvalidDelegationUpdateV2	IdentityCacheInvalidationTriggerV2	= "delegation_update"
	IdentityCacheInvalidZonePolicyUpdateV2	IdentityCacheInvalidationTriggerV2	= "zone_policy_update"
	IdentityCacheInvalidReverseUpdateV2	IdentityCacheInvalidationTriggerV2	= "reverse_record_update"
)

type IdentityCacheLayerV2 string

type IdentityCacheInvalidationTriggerV2 string

type IdentityResolutionCacheKeyV2 struct {
	Layer		IdentityCacheLayerV2
	NameHash	string
	RecordVersion	uint64
	Height		uint64
	PathHash	string
	TargetKey	string
}

type IdentityVerifiedCacheMetadataV2 struct {
	Key			IdentityResolutionCacheKeyV2
	ProofHeight		uint64
	TrustedHeaderChainID	string
	TrustedHeaderHeight	uint64
	TrustedHeaderAppHash	string
	FreshUntilHeight	uint64
	ResolverTTL		uint64
	DomainExpiryHeight	uint64
	Verified		bool
	LightClient		bool
}

type IdentityCacheInvalidationEventV2 struct {
	Trigger			IdentityCacheInvalidationTriggerV2
	NameHash		string
	RecordVersion		uint64
	Height			uint64
	ParentEpoch		uint64
	ChildEpoch		uint64
	ReverseAddressHex	string
}

type ResolutionCacheRecordV2 struct {
	NameHash		string
	ResolutionPathHash	string
	ResolvedRecordHash	string
	ValidUntilHeight	uint64
	SourceVersion		uint64
	ParentEpoch		uint64
	ChildEpoch		uint64
}

type ResolutionCacheUseContextV2 struct {
	Height		uint64
	SourceVersion	uint64
	ParentEpoch	uint64
	ChildEpoch	uint64
	LightClient	bool
	ProofVerified	bool
}

func NewResolutionCacheRecordV2(name string, resolutionPathHash string, resolvedRecordHash string, validUntilHeight uint64, sourceVersion uint64, parentEpoch uint64, childEpoch uint64) (ResolutionCacheRecordV2, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return ResolutionCacheRecordV2{}, err
	}
	record := ResolutionCacheRecordV2{
		NameHash:		nameHash,
		ResolutionPathHash:	resolutionPathHash,
		ResolvedRecordHash:	resolvedRecordHash,
		ValidUntilHeight:	validUntilHeight,
		SourceVersion:		sourceVersion,
		ParentEpoch:		parentEpoch,
		ChildEpoch:		childEpoch,
	}
	return record, ValidateResolutionCacheRecordV2(record)
}

func ValidateResolutionCacheRecordV2(record ResolutionCacheRecordV2) error {
	if err := validateHexHash("identity v2 cache name hash", record.NameHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 cache resolution path hash", record.ResolutionPathHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 cache resolved record hash", record.ResolvedRecordHash); err != nil {
		return err
	}
	if record.ValidUntilHeight == 0 {
		return errors.New("identity v2 cache valid_until_height is required")
	}
	if record.SourceVersion == 0 {
		return errors.New("identity v2 cache source_version is required")
	}
	return nil
}

func ValidateResolutionCacheRecordV2Use(record ResolutionCacheRecordV2, ctx ResolutionCacheUseContextV2) error {
	if err := ValidateResolutionCacheRecordV2(record); err != nil {
		return err
	}
	if ctx.Height == 0 {
		return errors.New("identity v2 cache use height is required")
	}
	if ctx.Height > record.ValidUntilHeight {
		return errors.New("identity v2 cache record is expired")
	}
	if ctx.SourceVersion != record.SourceVersion {
		return errors.New("identity v2 cache source version changed")
	}
	if ctx.ParentEpoch != record.ParentEpoch {
		return errors.New("identity v2 cache parent epoch changed")
	}
	if ctx.ChildEpoch != record.ChildEpoch {
		return errors.New("identity v2 cache child epoch changed")
	}
	if ctx.LightClient && !ctx.ProofVerified {
		return errors.New("identity v2 cache cannot bypass light-client proof verification")
	}
	return nil
}

func NewIdentityResolutionCacheKeyV2(layer IdentityCacheLayerV2, name string, recordVersion uint64, height uint64, pathHash string, targetKey string) (IdentityResolutionCacheKeyV2, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return IdentityResolutionCacheKeyV2{}, err
	}
	key := IdentityResolutionCacheKeyV2{
		Layer:		layer,
		NameHash:	nameHash,
		RecordVersion:	recordVersion,
		Height:		height,
		PathHash:	pathHash,
		TargetKey:	targetKey,
	}
	return key, ValidateIdentityResolutionCacheKeyV2(key)
}

func ValidateIdentityResolutionCacheKeyV2(key IdentityResolutionCacheKeyV2) error {
	if err := validateIdentityCacheLayerV2(key.Layer); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 cache key name hash", key.NameHash); err != nil {
		return err
	}
	if key.RecordVersion == 0 {
		return errors.New("identity v2 cache key record_version is required")
	}
	if key.Height == 0 {
		return errors.New("identity v2 cache key height is required")
	}
	if key.PathHash != "" {
		if err := validateHexHash("identity v2 cache key path hash", key.PathHash); err != nil {
			return err
		}
	}
	if key.TargetKey != "" {
		if err := ValidateResolverMetadataKey(key.TargetKey); err != nil {
			return err
		}
	}
	return nil
}

func FormatIdentityResolutionCacheKeyV2(key IdentityResolutionCacheKeyV2) (string, error) {
	if err := ValidateIdentityResolutionCacheKeyV2(key); err != nil {
		return "", err
	}
	pathHash := key.PathHash
	if pathHash == "" {
		pathHash = "direct"
	}
	targetKey := key.TargetKey
	if targetKey == "" {
		targetKey = "default"
	}
	return fmt.Sprintf("%s/%s/%020d/%020d/%s/%s", key.Layer, key.NameHash, key.RecordVersion, key.Height, pathHash, targetKey), nil
}

func NewIdentityVerifiedCacheMetadataV2(key IdentityResolutionCacheKeyV2, proofHeight uint64, trustedHeader IdentityTrustedHeaderV2, resolverTTL uint64, domainExpiryHeight uint64, freshnessThreshold uint64, lightClient bool) (IdentityVerifiedCacheMetadataV2, error) {
	if freshnessThreshold == 0 {
		return IdentityVerifiedCacheMetadataV2{}, errors.New("identity v2 verified cache freshness threshold is required")
	}
	freshUntil := proofHeight + freshnessThreshold
	if resolverTTL != 0 && proofHeight+resolverTTL < freshUntil {
		freshUntil = proofHeight + resolverTTL
	}
	if domainExpiryHeight != 0 && domainExpiryHeight < freshUntil {
		freshUntil = domainExpiryHeight
	}
	metadata := IdentityVerifiedCacheMetadataV2{
		Key:			key,
		ProofHeight:		proofHeight,
		TrustedHeaderChainID:	trustedHeader.ChainID,
		TrustedHeaderHeight:	trustedHeader.Height,
		TrustedHeaderAppHash:	trustedHeader.AppHash,
		FreshUntilHeight:	freshUntil,
		ResolverTTL:		resolverTTL,
		DomainExpiryHeight:	domainExpiryHeight,
		Verified:		trustedHeader.Trusted,
		LightClient:		lightClient,
	}
	return metadata, ValidateIdentityVerifiedCacheMetadataV2(metadata)
}

func ValidateIdentityVerifiedCacheMetadataV2(metadata IdentityVerifiedCacheMetadataV2) error {
	if err := ValidateIdentityResolutionCacheKeyV2(metadata.Key); err != nil {
		return err
	}
	if metadata.ProofHeight == 0 {
		return errors.New("identity v2 verified cache proof_height is required")
	}
	if metadata.FreshUntilHeight < metadata.ProofHeight {
		return errors.New("identity v2 verified cache fresh_until_height must be at or after proof_height")
	}
	if metadata.ResolverTTL == 0 {
		return errors.New("identity v2 verified cache resolver_ttl is required")
	}
	if metadata.FreshUntilHeight > metadata.ProofHeight+metadata.ResolverTTL {
		return errors.New("identity v2 verified cache ttl exceeds resolver ttl")
	}
	if metadata.DomainExpiryHeight != 0 && metadata.FreshUntilHeight > metadata.DomainExpiryHeight {
		return errors.New("identity v2 verified cache ttl exceeds domain expiry")
	}
	if metadata.LightClient {
		if !metadata.Verified {
			return errors.New("identity v2 light-client cache must be verified")
		}
		if metadata.TrustedHeaderChainID == "" || metadata.TrustedHeaderHeight == 0 || metadata.TrustedHeaderAppHash == "" {
			return errors.New("identity v2 light-client cache requires trusted header reference")
		}
		if metadata.TrustedHeaderHeight != metadata.ProofHeight {
			return errors.New("identity v2 light-client cache trusted header height must match proof height")
		}
		if err := validateHexHash("identity v2 light-client trusted app hash", metadata.TrustedHeaderAppHash); err != nil {
			return err
		}
	}
	return nil
}

func ValidateIdentityVerifiedCacheFreshnessV2(metadata IdentityVerifiedCacheMetadataV2, currentHeight uint64, freshnessThreshold uint64, executionCritical bool) error {
	if err := ValidateIdentityVerifiedCacheMetadataV2(metadata); err != nil {
		return err
	}
	if currentHeight == 0 {
		return errors.New("identity v2 cache freshness current height is required")
	}
	if currentHeight > metadata.FreshUntilHeight {
		return errors.New("identity v2 verified cache is stale")
	}
	if executionCritical && freshnessThreshold != 0 && currentHeight > metadata.ProofHeight+freshnessThreshold {
		return errors.New("identity v2 execution-critical cache proof freshness threshold exceeded")
	}
	return nil
}

func InvalidateIdentityResolutionCachesV2(records []ResolutionCacheRecordV2, event IdentityCacheInvalidationEventV2) ([]ResolutionCacheRecordV2, error) {
	if err := ValidateIdentityCacheInvalidationEventV2(event); err != nil {
		return nil, err
	}
	out := make([]ResolutionCacheRecordV2, 0, len(records))
	for _, record := range records {
		next := record
		if record.NameHash == event.NameHash {
			next = InvalidateResolutionCacheRecordV2ForDomainMutation(record, event.RecordVersion, event.ParentEpoch, event.ChildEpoch)
		}
		out = append(out, next)
	}
	return out, nil
}

func ValidateIdentityCacheInvalidationEventV2(event IdentityCacheInvalidationEventV2) error {
	if err := validateIdentityCacheInvalidationTriggerV2(event.Trigger); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 cache invalidation name hash", event.NameHash); err != nil {
		return err
	}
	if event.RecordVersion == 0 {
		return errors.New("identity v2 cache invalidation record_version is required")
	}
	if event.Height == 0 {
		return errors.New("identity v2 cache invalidation height is required")
	}
	if event.ReverseAddressHex != "" {
		if _, err := hex.DecodeString(event.ReverseAddressHex); err != nil {
			return errors.New("identity v2 cache invalidation reverse address must be hex")
		}
	}
	return nil
}

func InvalidateResolutionCacheRecordV2ForDomainMutation(record ResolutionCacheRecordV2, sourceVersion uint64, parentEpoch uint64, childEpoch uint64) ResolutionCacheRecordV2 {
	next := record
	next.SourceVersion = sourceVersion
	next.ParentEpoch = parentEpoch
	next.ChildEpoch = childEpoch
	next.ValidUntilHeight = 0
	return next
}

func validateIdentityCacheLayerV2(layer IdentityCacheLayerV2) error {
	switch layer {
	case IdentityCacheLayerFullNodeDirectV2, IdentityCacheLayerFullNodeRecursiveV2, IdentityCacheLayerWalletVerifiedV2, IdentityCacheLayerServiceEndpointV2, IdentityCacheLayerReverseVerifiedV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 cache layer %q", layer)
	}
}

func validateIdentityCacheInvalidationTriggerV2(trigger IdentityCacheInvalidationTriggerV2) error {
	switch trigger {
	case IdentityCacheInvalidDomainTransferV2,
		IdentityCacheInvalidResolverUpdateV2,
		IdentityCacheInvalidNFTBindingUpdateV2,
		IdentityCacheInvalidDomainExpiryV2,
		IdentityCacheInvalidRenewalEpochV2,
		IdentityCacheInvalidDelegationUpdateV2,
		IdentityCacheInvalidZonePolicyUpdateV2,
		IdentityCacheInvalidReverseUpdateV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 cache invalidation trigger %q", trigger)
	}
}

func ComputeResolutionPathHashV2(candidates []string) (string, error) {
	if len(candidates) == 0 {
		return "", errors.New("identity v2 cache resolution path is required")
	}
	normalized := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		name, err := NormalizeAETDomain(candidate)
		if err != nil {
			return "", err
		}
		normalized = append(normalized, name)
	}
	return identityHash(append([]string{"identity-v2-resolution-path", fmt.Sprintf("%020d", len(normalized))}, normalized...)...), nil
}

func ComputeResolvedRecordHashV2(record UnifiedResolutionRecordV2) (string, error) {
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return "", err
	}
	parts := []string{
		"identity-v2-resolved-record",
		record.NameHash,
		hex.EncodeToString(record.Owner),
		hex.EncodeToString(record.PrimaryAddress),
		fmt.Sprintf("%020d", record.RecordVersion),
		fmt.Sprintf("%020d", record.RecordTTL),
		fmt.Sprintf("%020d", record.UpdatedAtHeight),
		fmt.Sprintf("%020d", record.MaxPayloadBytes),
		fmt.Sprintf("%020d", record.SchemaVersion),
		record.RoutingMetadata.ZoneID,
		record.RoutingMetadata.ShardID,
		record.RoutingMetadata.VM,
		record.RoutingMetadata.Entrypoint,
		record.RoutingMetadata.RouteID,
		record.RoutingMetadata.TargetType,
		record.RoutingMetadata.PreferredTarget,
		fmt.Sprintf("%020d", len(record.RoutingMetadata.FallbackTargets)),
	}
	parts = append(parts, record.RoutingMetadata.FallbackTargets...)
	parts = append(parts,
		record.RoutingMetadata.ChainContext,
		record.RoutingMetadata.FeeHint,
		fmt.Sprintf("%020d", record.RoutingMetadata.TimeoutHint),
		record.RoutingMetadata.MemoPolicy,
		fmt.Sprintf("%020d", len(record.RoutingMetadata.CapabilityRequirements)),
	)
	parts = append(parts, record.RoutingMetadata.CapabilityRequirements...)
	for _, target := range record.ContractTargets {
		parts = append(parts,
			"contract",
			target.Key,
			hex.EncodeToString(target.Address),
			target.CodeID,
			target.TargetID,
			hex.EncodeToString(target.ContractAddress),
			target.Entrypoint,
			target.InterfaceHash,
			target.RequiredFundsPolicy,
			fmt.Sprintf("%020d", target.GasHint),
			fmt.Sprintf("%t", target.Enabled),
			fmt.Sprintf("%020d", target.UpdatedAtHeight),
		)
	}
	for _, endpoint := range record.ServiceEndpoints {
		parts = append(parts,
			"service",
			endpoint.Key,
			endpoint.Endpoint,
			endpoint.ServiceID,
			endpoint.ServiceType,
			endpoint.Transport,
			endpoint.AuthPolicy,
			endpoint.HealthPathOptional,
			fmt.Sprintf("%020d", endpoint.Priority),
			fmt.Sprintf("%020d", endpoint.Weight),
			fmt.Sprintf("%020d", endpoint.TTL),
			endpoint.SchemaHashOptional,
		)
	}
	for _, descriptor := range record.InterfaceDescriptors {
		parts = append(parts,
			"interface",
			descriptor.InterfaceID,
			descriptor.Descriptor,
			descriptor.SchemaHash,
			descriptor.SchemaURIOptional,
			descriptor.SchemaInlineOptional,
			descriptor.Version,
			descriptor.RenderPolicy,
			fmt.Sprintf("%020d", len(descriptor.PermissionsRequired)),
		)
		parts = append(parts, descriptor.PermissionsRequired...)
		parts = append(parts, descriptor.ContractTargetIDOptional, descriptor.ServiceIDOptional)
	}
	for _, hint := range record.ExecutionHints {
		parts = append(parts,
			"hint",
			hint.Key,
			hint.Value,
			fmt.Sprintf("%020d", hint.DefaultGasLimitHint),
			hint.PreferredFeeMode,
			hint.MessageType,
			fmt.Sprintf("%t", hint.AsyncAllowed),
			fmt.Sprintf("%t", hint.RequiresMemo),
			fmt.Sprintf("%t", hint.RequiresInterfaceConfirmation),
			fmt.Sprintf("%t", hint.SimulationRequired),
		)
	}
	return identityHash(parts...), nil
}
