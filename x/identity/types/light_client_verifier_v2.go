package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentityLightClientFailureCodeV2 string

const (
	IdentityLightClientErrInvalidName			IdentityLightClientFailureCodeV2	= "ERR_INVALID_NAME"
	IdentityLightClientErrUnsupportedNormalizationVersion	IdentityLightClientFailureCodeV2	= "ERR_UNSUPPORTED_NORMALIZATION_VERSION"
	IdentityLightClientErrProofHeightUntrusted		IdentityLightClientFailureCodeV2	= "ERR_PROOF_HEIGHT_UNTRUSTED"
	IdentityLightClientErrDomainNotFound			IdentityLightClientFailureCodeV2	= "ERR_DOMAIN_NOT_FOUND"
	IdentityLightClientErrDomainExpired			IdentityLightClientFailureCodeV2	= "ERR_DOMAIN_EXPIRED"
	IdentityLightClientErrDomainNotActive			IdentityLightClientFailureCodeV2	= "ERR_DOMAIN_NOT_ACTIVE"
	IdentityLightClientErrNFTBindingMismatch		IdentityLightClientFailureCodeV2	= "ERR_NFT_BINDING_MISMATCH"
	IdentityLightClientErrResolverNotFound			IdentityLightClientFailureCodeV2	= "ERR_RESOLVER_NOT_FOUND"
	IdentityLightClientErrResolverUnauthorized		IdentityLightClientFailureCodeV2	= "ERR_RESOLVER_UNAUTHORIZED"
	IdentityLightClientErrTargetNotFound			IdentityLightClientFailureCodeV2	= "ERR_TARGET_NOT_FOUND"
	IdentityLightClientErrDelegationMissing			IdentityLightClientFailureCodeV2	= "ERR_DELEGATION_MISSING"
	IdentityLightClientErrDelegationExpired			IdentityLightClientFailureCodeV2	= "ERR_DELEGATION_EXPIRED"
	IdentityLightClientErrReverseNotVerified		IdentityLightClientFailureCodeV2	= "ERR_REVERSE_NOT_VERIFIED"
	IdentityLightClientErrProofInvalid			IdentityLightClientFailureCodeV2	= "ERR_PROOF_INVALID"
	IdentityLightClientErrRecordStale			IdentityLightClientFailureCodeV2	= "ERR_RECORD_STALE"
)

type IdentityLightClientVerificationErrorV2 struct {
	Code	IdentityLightClientFailureCodeV2
	Message	string
	Cause	error
}

func (e IdentityLightClientVerificationErrorV2) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e IdentityLightClientVerificationErrorV2) Unwrap() error {
	return e.Cause
}

func IdentityLightClientFailureCodeFromErrorV2(err error) (IdentityLightClientFailureCodeV2, bool) {
	var coded IdentityLightClientVerificationErrorV2
	if errors.As(err, &coded) {
		return coded.Code, true
	}
	return "", false
}

type IdentityTrustedHeaderV2 struct {
	ChainID	string
	Height	uint64
	AppHash	string
	Trusted	bool
}

type IdentityLightClientVerificationRequestV2 struct {
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
}

type IdentityLightClientVerifiedTargetV2 struct {
	Name			string
	NameHash		string
	ResolverNameHash	string
	TargetType		IdentityResolutionTargetTypeV2
	TargetKey		string
	Address			sdk.AccAddress
	Endpoint		string
	Descriptor		string
	Route			RoutingMetadataV2
	RecordVersion		uint64
	FreshUntilHeight	uint64
	ProofHeight		uint64
}

func VerifyIdentityResolutionProofLightClientV2(request IdentityLightClientVerificationRequestV2) (IdentityLightClientVerifiedTargetV2, error) {
	version := request.NormalizationVersion
	if version == 0 {
		version = NameNormalizationVersionV2
	}
	if err := ValidateNameNormalizationVersionV2(version); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrUnsupportedNormalizationVersion, "unsupported name normalization version", err)
	}
	normalized, err := NormalizeAETDomainVersioned(request.RequestedName, version)
	if err != nil {
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrInvalidName, "requested name failed deterministic normalization", err)
	}
	if err := verifyTrustedHeaderForResolutionProofV2(request); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	proof := request.Proof
	if proof.Name != normalized.NormalizedName || proof.NameHash != normalized.NameHash {
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrInvalidName, "requested name does not match proof name", nil)
	}
	if err := ValidateIdentityResolutionProofFormatV2(proof); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrProofInvalid, "resolution proof format is invalid", err)
	}
	if err := verifyLightClientDomainV2(proof, request); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	if err := verifyLightClientNFTBindingV2(proof); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	if err := verifyLightClientResolverV2(proof, request); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	if err := verifyLightClientRecursivePathV2(proof, request); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	if err := verifyLightClientReverseV2(proof, request); err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	target, err := lightClientTargetFromRecordV2(*proof.ResolverRecord, request)
	if err != nil {
		return IdentityLightClientVerifiedTargetV2{}, err
	}
	target.Name = proof.Name
	target.NameHash = proof.NameHash
	target.ResolverNameHash = proof.ResolverRecord.NameHash
	target.RecordVersion = proof.RecordVersion
	target.FreshUntilHeight = proof.ResolverRecord.UpdatedAtHeight + proof.ResolverRecord.RecordTTL
	target.ProofHeight = proof.Height
	return target, nil
}

func verifyTrustedHeaderForResolutionProofV2(request IdentityLightClientVerificationRequestV2) error {
	header := request.TrustedHeader
	proof := request.Proof
	if !header.Trusted || header.Height == 0 || header.AppHash == "" {
		return lightClientFailV2(IdentityLightClientErrProofHeightUntrusted, "trusted header for proof height is required", nil)
	}
	if header.Height != proof.Height || header.AppHash != proof.AppHash {
		return lightClientFailV2(IdentityLightClientErrProofHeightUntrusted, "trusted header does not match proof height or app_hash", nil)
	}
	if request.ExpectedChainID == "" || header.ChainID != request.ExpectedChainID || proof.ChainID != request.ExpectedChainID {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "proof chain_id does not match expected chain", nil)
	}
	return nil
}

func verifyLightClientDomainV2(proof IdentityResolutionProofFormatV2, request IdentityLightClientVerificationRequestV2) error {
	if proof.DomainRecord == nil {
		if proof.NonExistenceProofOptional != nil && proof.QueryType == IdentityProofQueryDomainAbsent {
			if proof.NonExistenceProofOptional.RootHash != proof.AppHash {
				return lightClientFailV2(IdentityLightClientErrProofInvalid, "domain non-existence proof root mismatch", nil)
			}
			return lightClientFailV2(IdentityLightClientErrDomainNotFound, "domain is absent at proof height", nil)
		}
		return lightClientFailV2(IdentityLightClientErrDomainNotFound, "domain record proof is missing", nil)
	}
	if proof.DomainRecordProof == nil {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "domain record proof is missing", nil)
	}
	if proof.DomainRecordProof.RootHash != proof.AppHash {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "domain record proof root mismatch", nil)
	}
	expectedKey, err := IdentityDomainStoreKey(proof.Name)
	if err != nil {
		return lightClientFailV2(IdentityLightClientErrInvalidName, "domain store key is invalid", err)
	}
	if proof.DomainRecordProof.Key != expectedKey {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "domain record proof key mismatch", nil)
	}
	switch proof.DomainRecord.Status {
	case DomainRecordV2Active:
	case DomainRecordV2RenewalWindow:
		if !request.AllowRenewalWindow {
			return lightClientFailV2(IdentityLightClientErrDomainNotActive, "renewal-window domain is not allowed for this verification", nil)
		}
	default:
		return lightClientFailV2(IdentityLightClientErrDomainNotActive, "domain status is not active", nil)
	}
	if proof.DomainRecord.ExpiryHeight <= proof.Height {
		return lightClientFailV2(IdentityLightClientErrDomainExpired, "domain expiry has passed at proof height", nil)
	}
	nftOwner := proof.DomainRecord.Owner
	if proof.NFTBinding != nil {
		if !addressesEqual(proof.NFTBinding.Owner, proof.DomainRecord.Owner) {
			return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "registry owner does not match nft owner", nil)
		}
		nftOwner = proof.NFTBinding.Owner
	}
	if err := ValidateDomainRecordV2(*proof.DomainRecord, DomainRecordV2ValidationContext{CurrentHeight: proof.Height, NFTOwner: nftOwner}); err != nil {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "domain record failed validation", err)
	}
	return nil
}

func verifyLightClientNFTBindingV2(proof IdentityResolutionProofFormatV2) error {
	if proof.NFTBinding == nil || proof.NFTBindingProof == nil {
		return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "nft binding proof is required", nil)
	}
	if proof.DomainRecord == nil {
		return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "domain record is required for nft binding verification", nil)
	}
	if proof.NFTBindingProof.RootHash != proof.AppHash {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "nft binding proof root mismatch", nil)
	}
	expectedKey, err := IdentityNFTStoreKey(proof.NFTBinding.NFTItemID)
	if err != nil {
		return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "nft binding key is invalid", err)
	}
	if proof.NFTBindingProof.Key != expectedKey {
		return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "nft binding proof key mismatch", nil)
	}
	if proof.DomainRecord.NameHash != proof.NFTBinding.NameHash ||
		proof.DomainRecord.NFTClassID != proof.NFTBinding.NFTClassID ||
		proof.DomainRecord.NFTItemID != proof.NFTBinding.NFTItemID {
		return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "registry domain record and nft binding identify different assets", nil)
	}
	if err := ValidateDomainNFTBinding(*proof.NFTBinding, DomainNFTBindingContext{
		RegistryOwner:	proof.DomainRecord.Owner,
		NFTModuleOwner:	proof.NFTBinding.Owner,
		CurrentHeight:	proof.Height,
	}); err != nil {
		return lightClientFailV2(IdentityLightClientErrNFTBindingMismatch, "registry owner does not match nft owner", err)
	}
	return nil
}

func verifyLightClientResolverV2(proof IdentityResolutionProofFormatV2, request IdentityLightClientVerificationRequestV2) error {
	if proof.ResolverRecord == nil || proof.ResolverRecordProof == nil {
		return lightClientFailV2(IdentityLightClientErrResolverNotFound, "resolver record proof is required", nil)
	}
	if proof.ResolverRecordProof.RootHash != proof.AppHash {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "resolver record proof root mismatch", nil)
	}
	if err := ValidateUnifiedResolutionRecordV2(*proof.ResolverRecord); err != nil {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "resolver record failed validation", err)
	}
	if proof.RecordVersion == 0 || proof.RecordVersion != proof.ResolverRecord.RecordVersion {
		return lightClientFailV2(IdentityLightClientErrRecordStale, "resolver record version mismatch", nil)
	}
	if proof.ResolverRecord.RecordTTL == 0 || proof.ResolverRecord.UpdatedAtHeight == 0 || proof.ResolverRecord.UpdatedAtHeight > proof.Height {
		return lightClientFailV2(IdentityLightClientErrRecordStale, "resolver ttl or updated_at_height is invalid", nil)
	}
	freshUntil := proof.ResolverRecord.UpdatedAtHeight + proof.ResolverRecord.RecordTTL
	if freshUntil < proof.Height {
		return lightClientFailV2(IdentityLightClientErrRecordStale, "resolver ttl expired before proof height", nil)
	}
	if request.CurrentHeight != 0 && request.CurrentHeight > freshUntil {
		return lightClientFailV2(IdentityLightClientErrRecordStale, "resolver record is stale at current height", nil)
	}
	if !lightClientResolverAuthorizedByPathV2(proof, request) {
		return lightClientFailV2(IdentityLightClientErrResolverUnauthorized, "resolver record is not controlled by a proven path domain", nil)
	}
	return nil
}

func verifyLightClientRecursivePathV2(proof IdentityResolutionProofFormatV2, request IdentityLightClientVerificationRequestV2) error {
	if request.RecursiveProof == nil {
		return nil
	}
	recursive := *request.RecursiveProof
	if err := ValidateRecursiveResolutionProofV2(recursive); err != nil {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "recursive proof is invalid", err)
	}
	if recursive.ChainID != proof.ChainID || recursive.Height != proof.Height || recursive.TargetName != proof.Name {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "recursive proof does not match resolution proof", nil)
	}
	if recursive.FinalRecordProof.RootHash != proof.AppHash {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "recursive proof root mismatch", nil)
	}
	path, err := CanonicalResolutionPathV2(proof.Name)
	if err != nil {
		return lightClientFailV2(IdentityLightClientErrInvalidName, "recursive path name is invalid", err)
	}
	if !sameStringSliceV2(path.Labels, recursive.PathLabels) || !sameStringSliceV2(path.PathHashes, recursive.PathHashes) {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "recursive path labels or hashes mismatch", nil)
	}
	provenDomains := map[string]struct{}{}
	for _, record := range recursive.PathDomainRecords {
		if record.ExpiryHeight <= proof.Height {
			return lightClientFailV2(IdentityLightClientErrDomainExpired, "recursive path component is expired", nil)
		}
		switch record.Status {
		case DomainRecordV2Active, DomainRecordV2RenewalWindow:
			provenDomains[record.NameHash] = struct{}{}
		default:
			return lightClientFailV2(IdentityLightClientErrDomainNotActive, "recursive path component is not active", nil)
		}
	}
	for _, delegation := range recursive.PathDelegationRecords {
		if delegation.ExpiresAtHeight <= proof.Height {
			return lightClientFailV2(IdentityLightClientErrDelegationExpired, "recursive delegation is expired", nil)
		}
	}
	for i, pathHash := range path.PathHashes {
		if _, found := provenDomains[pathHash]; found {
			continue
		}
		if i == 0 || !lightClientPathHasDelegationV2(path, i, recursive.PathDelegationRecords, proof.Height) {
			return lightClientFailV2(IdentityLightClientErrDelegationMissing, "recursive path component lacks domain or delegation proof", nil)
		}
	}
	return nil
}

func verifyLightClientReverseV2(proof IdentityResolutionProofFormatV2, request IdentityLightClientVerificationRequestV2) error {
	if !request.RequireReverseResolution && proof.QueryType != IdentityProofQueryResolveReverse && proof.ReverseRecordOptional == nil {
		return nil
	}
	if proof.ReverseRecordOptional == nil || proof.ReverseRecordProofOptional == nil {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse proof is required", nil)
	}
	record := *proof.ReverseRecordOptional
	if !record.Verified {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse record is not verified", nil)
	}
	if record.ExpiryHeight <= proof.Height {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse record is expired", nil)
	}
	if record.Name != proof.Name || record.NameHash != proof.NameHash {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse record does not match requested name", nil)
	}
	if proof.ReverseRecordProofOptional.RootHash != proof.AppHash {
		return lightClientFailV2(IdentityLightClientErrProofInvalid, "reverse proof root mismatch", nil)
	}
	expectedKey, err := IdentityReverseStoreKey(record.Address)
	if err != nil {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse record address is invalid", err)
	}
	if proof.ReverseRecordProofOptional.Key != expectedKey {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse proof key mismatch", nil)
	}
	if !lightClientForwardContainsAddressV2(*proof.ResolverRecord, record.Address, request.AuthorizedAliasKeys) {
		return lightClientFailV2(IdentityLightClientErrReverseNotVerified, "reverse record does not match forward resolution", nil)
	}
	return nil
}

func lightClientTargetFromRecordV2(record UnifiedResolutionRecordV2, request IdentityLightClientVerificationRequestV2) (IdentityLightClientVerifiedTargetV2, error) {
	targetType := request.TargetType
	if targetType == "" {
		targetType = IdentityResolutionTargetPrimary
	}
	target := IdentityLightClientVerifiedTargetV2{TargetType: targetType, TargetKey: request.TargetKey}
	switch targetType {
	case IdentityResolutionTargetPrimary:
		if len(record.PrimaryAddress) == 0 {
			return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "primary target is missing", nil)
		}
		target.Address = cloneSpecAddress(record.PrimaryAddress)
	case IdentityResolutionTargetContract:
		for _, contract := range record.ContractTargets {
			if !contractTargetEnabledV2(contract) {
				continue
			}
			targetID := contractTargetIDV2(contract)
			if request.TargetKey == "" || targetID == request.TargetKey {
				address := contractTargetAddressV2(contract)
				if len(address) > 0 {
					target.Address = cloneSpecAddress(address)
					target.TargetKey = targetID
					return target, nil
				}
			}
		}
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "contract target is missing", nil)
	case IdentityResolutionTargetService:
		for _, endpoint := range record.ServiceEndpoints {
			endpointID := serviceEndpointIDV2(endpoint)
			if endpointID == request.TargetKey {
				target.Endpoint = endpoint.Endpoint
				target.TargetKey = endpointID
				return target, nil
			}
		}
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "service target is missing", nil)
	case IdentityResolutionTargetInterface:
		for _, descriptor := range record.InterfaceDescriptors {
			if descriptor.InterfaceID == request.TargetKey {
				target.Descriptor = interfaceDescriptorSchemaHashV2(descriptor)
				return target, nil
			}
		}
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "interface target is missing", nil)
	case IdentityResolutionTargetRoute:
		if !routingMetadataHasTargetV2(record.RoutingMetadata) {
			return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "route target is missing", nil)
		}
		target.Route = record.RoutingMetadata
	case IdentityResolutionTargetRecord:
		if request.TargetKey == ResolverKeyPrimary {
			if len(record.PrimaryAddress) == 0 {
				return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "primary record target is missing", nil)
			}
			target.Address = cloneSpecAddress(record.PrimaryAddress)
			return target, nil
		}
		for _, contract := range record.ContractTargets {
			if !contractTargetEnabledV2(contract) {
				continue
			}
			if contractTargetIDV2(contract) == request.TargetKey {
				address := contractTargetAddressV2(contract)
				if len(address) == 0 {
					continue
				}
				target.Address = cloneSpecAddress(address)
				return target, nil
			}
		}
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "record target is missing", nil)
	default:
		return IdentityLightClientVerifiedTargetV2{}, lightClientFailV2(IdentityLightClientErrTargetNotFound, "unsupported target type", nil)
	}
	return target, nil
}

func lightClientResolverAuthorizedByPathV2(proof IdentityResolutionProofFormatV2, request IdentityLightClientVerificationRequestV2) bool {
	if proof.DomainRecord != nil && proof.ResolverRecord.NameHash == proof.DomainRecord.NameHash {
		return true
	}
	if request.RecursiveProof == nil {
		return false
	}
	for _, record := range request.RecursiveProof.PathDomainRecords {
		if record.NameHash == proof.ResolverRecord.NameHash {
			return true
		}
	}
	return false
}

func lightClientPathHasDelegationV2(path DeterministicResolutionPathV2, index int, delegations []DelegationRecordV2, height uint64) bool {
	for ancestor := 0; ancestor < index; ancestor++ {
		depth := uint8(index - ancestor)
		recordKey := stringsJoinV2(path.Labels[ancestor+1 : index+1])
		for _, delegation := range delegations {
			if delegation.NameHash != path.PathHashes[ancestor] {
				continue
			}
			if ValidateDelegationRecordV2Use(delegation, DelegationScopeSubdomainCreate, "create", recordKey, depth, height) == nil {
				return true
			}
			if ValidateDelegationRecordV2Use(delegation, DelegationScopeZoneAdmin, "resolve", recordKey, depth, height) == nil {
				return true
			}
		}
	}
	return false
}

func lightClientForwardContainsAddressV2(record UnifiedResolutionRecordV2, address []byte, authorizedAliasKeys []string) bool {
	if addressesEqual(record.PrimaryAddress, address) {
		return true
	}
	allowed := stringSet(authorizedAliasKeys)
	for _, target := range record.ContractTargets {
		if !contractTargetEnabledV2(target) {
			continue
		}
		if _, found := allowed[contractTargetIDV2(target)]; found && addressesEqual(contractTargetAddressV2(target), address) {
			return true
		}
	}
	return false
}

func sameStringSliceV2(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func stringsJoinV2(parts []string) string {
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += "."
		}
		out += part
	}
	return out
}

func lightClientFailV2(code IdentityLightClientFailureCodeV2, message string, cause error) error {
	return IdentityLightClientVerificationErrorV2{Code: code, Message: message, Cause: cause}
}
