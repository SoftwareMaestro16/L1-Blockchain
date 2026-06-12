package types

import "fmt"

const (
	DefaultIdentityLightClientFreshnessThresholdV2	= uint64(20)

	IdentityLightClientProofRequestResolutionV2		IdentityLightClientProofRequestKindV2	= "QueryResolutionProof"
	IdentityLightClientProofRequestRecursiveResolutionV2	IdentityLightClientProofRequestKindV2	= "QueryRecursiveResolutionProof"

	IdentityLightClientCheckPassedV2	IdentityLightClientCheckStatusV2	= "passed"
	IdentityLightClientCheckFailedV2	IdentityLightClientCheckStatusV2	= "failed"
	IdentityLightClientCheckSkippedV2	IdentityLightClientCheckStatusV2	= "skipped"

	IdentityLightClientCheckHeaderTrustV2		IdentityLightClientCheckNameV2	= "header_trust"
	IdentityLightClientCheckChainIDV2		IdentityLightClientCheckNameV2	= "chain_id_match"
	IdentityLightClientCheckProofHeightV2		IdentityLightClientCheckNameV2	= "proof_height_match"
	IdentityLightClientCheckNameNormalizationV2	IdentityLightClientCheckNameV2	= "name_normalization_match"
	IdentityLightClientCheckNameHashV2		IdentityLightClientCheckNameV2	= "name_hash_match"
	IdentityLightClientCheckDomainProofV2		IdentityLightClientCheckNameV2	= "domain_record_proof"
	IdentityLightClientCheckDomainLifecycleV2	IdentityLightClientCheckNameV2	= "domain_lifecycle_validity"
	IdentityLightClientCheckNFTBindingV2		IdentityLightClientCheckNameV2	= "nft_binding_proof"
	IdentityLightClientCheckOwnershipV2		IdentityLightClientCheckNameV2	= "ownership_consistency"
	IdentityLightClientCheckResolverProofV2		IdentityLightClientCheckNameV2	= "resolver_proof"
	IdentityLightClientCheckTargetExistsV2		IdentityLightClientCheckNameV2	= "requested_target_field_existence"
	IdentityLightClientCheckTTLExpiryV2		IdentityLightClientCheckNameV2	= "record_ttl_and_expiry"
	IdentityLightClientCheckDelegationProofV2	IdentityLightClientCheckNameV2	= "delegation_proof_for_subdomains"
	IdentityLightClientCheckReverseConsistencyV2	IdentityLightClientCheckNameV2	= "reverse_forward_consistency_proof"
)

type IdentityLightClientProofRequestKindV2 string

type IdentityLightClientCheckStatusV2 string

type IdentityLightClientCheckNameV2 string

type IdentityLightClientResolutionCheckV2 struct {
	Name	IdentityLightClientCheckNameV2
	Status	IdentityLightClientCheckStatusV2
	Code	IdentityLightClientFailureCodeV2
	Message	string
}

type IdentityLightweightResolutionRequestV2 struct {
	ExpectedChainID			string
	RequestedName			string
	NormalizationVersion		uint64
	TrustedHeader			IdentityTrustedHeaderV2
	Proof				IdentityResolutionProofFormatV2
	RecursiveProof			*RecursiveResolutionProofV2
	TargetType			IdentityResolutionTargetTypeV2
	TargetKey			string
	AllowRenewalWindow		bool
	CurrentHeight			uint64
	AuthorizedAliasKeys		[]string
	RequireReverseResolution	bool
	FreshnessThreshold		uint64
}

type IdentityLightweightResolutionResultV2 struct {
	Verified	bool
	ProofRequest	IdentityLightClientProofRequestKindV2
	NormalizedName	string
	NameHash	string
	Target		IdentityLightClientVerifiedTargetV2
	CacheMetadata	*IdentityVerifiedCacheMetadataV2
	Checks		[]IdentityLightClientResolutionCheckV2
	FailureCode	IdentityLightClientFailureCodeV2
	Error		string
}

func ResolveIdentityLightweightV2(request IdentityLightweightResolutionRequestV2) (IdentityLightweightResolutionResultV2, error) {
	checks := newLightClientResolutionChecksV2(request)
	result := IdentityLightweightResolutionResultV2{
		ProofRequest:	lightClientProofRequestTypeV2(request),
		Checks:		checks,
	}
	normalized, err := lightClientPreflightResolutionRequestV2(request, &result)
	if err != nil {
		return result, err
	}
	result.NormalizedName = normalized.NormalizedName
	result.NameHash = normalized.NameHash

	target, err := VerifyIdentityResolutionProofLightClientV2(IdentityLightClientVerificationRequestV2{
		ExpectedChainID:		request.ExpectedChainID,
		RequestedName:			request.RequestedName,
		NormalizationVersion:		normalized.Version,
		TrustedHeader:			request.TrustedHeader,
		Proof:				request.Proof,
		RecursiveProof:			request.RecursiveProof,
		TargetType:			request.TargetType,
		TargetKey:			request.TargetKey,
		AllowRenewalWindow:		request.AllowRenewalWindow,
		CurrentHeight:			request.CurrentHeight,
		AuthorizedAliasKeys:		append([]string(nil), request.AuthorizedAliasKeys...),
		RequireReverseResolution:	request.RequireReverseResolution,
	})
	if err != nil {
		failLightClientResultV2(&result, checkForLightClientFailureV2(err), err)
		return result, err
	}
	result.Target = target
	metadata, err := buildLightClientVerifiedCacheMetadataV2(request, target)
	if err != nil {
		failLightClientResultV2(&result, IdentityLightClientCheckTTLExpiryV2, err)
		return result, err
	}
	result.CacheMetadata = &metadata
	markLightClientChecksSuccessV2(&result, request)
	result.Verified = true
	return result, nil
}

func lightClientPreflightResolutionRequestV2(request IdentityLightweightResolutionRequestV2, result *IdentityLightweightResolutionResultV2) (NameNormalizationResultV2, error) {
	version := request.NormalizationVersion
	if version == 0 {
		version = NameNormalizationVersionV2
	}
	if err := ValidateNameNormalizationVersionV2(version); err != nil {
		wrapped := lightClientFailV2(IdentityLightClientErrUnsupportedNormalizationVersion, "unsupported name normalization version", err)
		failLightClientResultV2(result, IdentityLightClientCheckNameNormalizationV2, wrapped)
		return NameNormalizationResultV2{}, wrapped
	}
	normalized, err := NormalizeAETDomainVersioned(request.RequestedName, version)
	if err != nil {
		wrapped := lightClientFailV2(IdentityLightClientErrInvalidName, "requested name failed deterministic normalization", err)
		failLightClientResultV2(result, IdentityLightClientCheckNameNormalizationV2, wrapped)
		return NameNormalizationResultV2{}, wrapped
	}
	if !request.TrustedHeader.Trusted || request.TrustedHeader.Height == 0 || request.TrustedHeader.AppHash == "" {
		err := lightClientFailV2(IdentityLightClientErrProofHeightUntrusted, "trusted header for proof height is required", nil)
		failLightClientResultV2(result, IdentityLightClientCheckHeaderTrustV2, err)
		return NameNormalizationResultV2{}, err
	}
	if request.TrustedHeader.Height != request.Proof.Height || request.TrustedHeader.AppHash != request.Proof.AppHash {
		err := lightClientFailV2(IdentityLightClientErrProofHeightUntrusted, "trusted header does not match proof height or app_hash", nil)
		failLightClientResultV2(result, IdentityLightClientCheckProofHeightV2, err)
		return NameNormalizationResultV2{}, err
	}
	if request.ExpectedChainID == "" || request.TrustedHeader.ChainID != request.ExpectedChainID || request.Proof.ChainID != request.ExpectedChainID {
		err := lightClientFailV2(IdentityLightClientErrProofInvalid, "proof chain_id does not match expected chain", nil)
		failLightClientResultV2(result, IdentityLightClientCheckChainIDV2, err)
		return NameNormalizationResultV2{}, err
	}
	if request.Proof.Name != normalized.NormalizedName {
		err := lightClientFailV2(IdentityLightClientErrInvalidName, "requested name does not match proof name", nil)
		failLightClientResultV2(result, IdentityLightClientCheckNameNormalizationV2, err)
		return NameNormalizationResultV2{}, err
	}
	if request.Proof.NameHash != normalized.NameHash {
		err := lightClientFailV2(IdentityLightClientErrInvalidName, "requested name_hash does not match proof name_hash", nil)
		failLightClientResultV2(result, IdentityLightClientCheckNameHashV2, err)
		return NameNormalizationResultV2{}, err
	}
	return normalized, nil
}

func buildLightClientVerifiedCacheMetadataV2(request IdentityLightweightResolutionRequestV2, target IdentityLightClientVerifiedTargetV2) (IdentityVerifiedCacheMetadataV2, error) {
	if request.Proof.ResolverRecord == nil || request.Proof.DomainRecord == nil {
		return IdentityVerifiedCacheMetadataV2{}, fmt.Errorf("identity light-client cache requires resolver and domain records")
	}
	pathHash := ""
	if request.RecursiveProof != nil {
		pathHash = request.RecursiveProof.ProofCommitmentHash
	}
	targetKey := target.TargetKey
	if targetKey == "" {
		targetKey = request.TargetKey
	}
	key, err := NewIdentityResolutionCacheKeyV2(IdentityCacheLayerWalletVerifiedV2, target.Name, target.RecordVersion, target.ProofHeight, pathHash, targetKey)
	if err != nil {
		return IdentityVerifiedCacheMetadataV2{}, err
	}
	freshness := request.FreshnessThreshold
	if freshness == 0 {
		freshness = request.Proof.ResolverRecord.RecordTTL
	}
	if freshness == 0 {
		freshness = DefaultIdentityLightClientFreshnessThresholdV2
	}
	return NewIdentityVerifiedCacheMetadataV2(key, target.ProofHeight, request.TrustedHeader, request.Proof.ResolverRecord.RecordTTL, request.Proof.DomainRecord.ExpiryHeight, freshness, true)
}

func newLightClientResolutionChecksV2(request IdentityLightweightResolutionRequestV2) []IdentityLightClientResolutionCheckV2 {
	names := []IdentityLightClientCheckNameV2{
		IdentityLightClientCheckHeaderTrustV2,
		IdentityLightClientCheckChainIDV2,
		IdentityLightClientCheckProofHeightV2,
		IdentityLightClientCheckNameNormalizationV2,
		IdentityLightClientCheckNameHashV2,
		IdentityLightClientCheckDomainProofV2,
		IdentityLightClientCheckDomainLifecycleV2,
		IdentityLightClientCheckNFTBindingV2,
		IdentityLightClientCheckOwnershipV2,
		IdentityLightClientCheckResolverProofV2,
		IdentityLightClientCheckTargetExistsV2,
		IdentityLightClientCheckTTLExpiryV2,
		IdentityLightClientCheckDelegationProofV2,
		IdentityLightClientCheckReverseConsistencyV2,
	}
	checks := make([]IdentityLightClientResolutionCheckV2, 0, len(names))
	for _, name := range names {
		status := IdentityLightClientCheckSkippedV2
		if lightClientCheckApplicableV2(name, request) {
			status = IdentityLightClientCheckFailedV2
		}
		checks = append(checks, IdentityLightClientResolutionCheckV2{Name: name, Status: status})
	}
	return checks
}

func markLightClientChecksSuccessV2(result *IdentityLightweightResolutionResultV2, request IdentityLightweightResolutionRequestV2) {
	for i := range result.Checks {
		if lightClientCheckApplicableV2(result.Checks[i].Name, request) {
			result.Checks[i].Status = IdentityLightClientCheckPassedV2
			result.Checks[i].Code = ""
			result.Checks[i].Message = ""
		} else {
			result.Checks[i].Status = IdentityLightClientCheckSkippedV2
		}
	}
}

func failLightClientResultV2(result *IdentityLightweightResolutionResultV2, check IdentityLightClientCheckNameV2, err error) {
	code, _ := IdentityLightClientFailureCodeFromErrorV2(err)
	result.FailureCode = code
	if err != nil {
		result.Error = err.Error()
	}
	for i := range result.Checks {
		if result.Checks[i].Name != check {
			continue
		}
		result.Checks[i].Status = IdentityLightClientCheckFailedV2
		result.Checks[i].Code = code
		if err != nil {
			result.Checks[i].Message = err.Error()
		}
		return
	}
}

func checkForLightClientFailureV2(err error) IdentityLightClientCheckNameV2 {
	code, ok := IdentityLightClientFailureCodeFromErrorV2(err)
	if !ok {
		return IdentityLightClientCheckResolverProofV2
	}
	switch code {
	case IdentityLightClientErrInvalidName, IdentityLightClientErrUnsupportedNormalizationVersion:
		return IdentityLightClientCheckNameNormalizationV2
	case IdentityLightClientErrProofHeightUntrusted:
		return IdentityLightClientCheckProofHeightV2
	case IdentityLightClientErrDomainNotFound:
		return IdentityLightClientCheckDomainProofV2
	case IdentityLightClientErrDomainExpired, IdentityLightClientErrDomainNotActive:
		return IdentityLightClientCheckDomainLifecycleV2
	case IdentityLightClientErrNFTBindingMismatch:
		return IdentityLightClientCheckNFTBindingV2
	case IdentityLightClientErrResolverNotFound, IdentityLightClientErrResolverUnauthorized:
		return IdentityLightClientCheckResolverProofV2
	case IdentityLightClientErrTargetNotFound:
		return IdentityLightClientCheckTargetExistsV2
	case IdentityLightClientErrDelegationMissing, IdentityLightClientErrDelegationExpired:
		return IdentityLightClientCheckDelegationProofV2
	case IdentityLightClientErrReverseNotVerified:
		return IdentityLightClientCheckReverseConsistencyV2
	case IdentityLightClientErrRecordStale:
		return IdentityLightClientCheckTTLExpiryV2
	default:
		return IdentityLightClientCheckResolverProofV2
	}
}

func lightClientCheckApplicableV2(check IdentityLightClientCheckNameV2, request IdentityLightweightResolutionRequestV2) bool {
	switch check {
	case IdentityLightClientCheckDelegationProofV2:
		return request.RecursiveProof != nil || lightClientNameHasSubdomainLabelsV2(request.Proof.Name)
	case IdentityLightClientCheckReverseConsistencyV2:
		return request.RequireReverseResolution || request.Proof.QueryType == IdentityProofQueryResolveReverse || request.Proof.ReverseRecordOptional != nil
	default:
		return true
	}
}

func lightClientProofRequestTypeV2(request IdentityLightweightResolutionRequestV2) IdentityLightClientProofRequestKindV2 {
	if request.RecursiveProof != nil {
		return IdentityLightClientProofRequestRecursiveResolutionV2
	}
	return IdentityLightClientProofRequestResolutionV2
}

func lightClientNameHasSubdomainLabelsV2(name string) bool {
	result, err := NormalizeAETDomainVersioned(name, NameNormalizationVersionV2)
	if err != nil {
		return false
	}
	return len(result.Labels) > 1
}
