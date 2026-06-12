package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type SubdomainDelegationTypeV2 string

const (
	SubdomainDelegationOwnerControlledV2	SubdomainDelegationTypeV2	= "owner_controlled"
	SubdomainDelegationDelegateControlledV2	SubdomainDelegationTypeV2	= "delegate_controlled"
	SubdomainDelegationZoneManagedV2	SubdomainDelegationTypeV2	= "zone_managed"
	SubdomainDelegationDetachedPaidV2	SubdomainDelegationTypeV2	= "detached_paid"
	SubdomainDelegationEphemeralServiceV2	SubdomainDelegationTypeV2	= "ephemeral_service"

	IdentityPathCommitmentVersionV2	uint64	= 1
	ZonePolicyVersionV2		uint64	= 1
	RecursivePolicyProofVersionV2	uint64	= 1

	ZoneSubdomainCreationOpenV2		= "open"
	ZoneSubdomainCreationOwnerOnlyV2	= "owner_only"
	ZoneSubdomainCreationDelegatedV2	= "delegated"
	ZoneSubdomainCreationClosedV2		= "closed"
	ZoneResolverUpdateOwnerOnlyV2		= "owner_only"
	ZoneResolverUpdateDelegatedV2		= "delegated"
	ZoneResolverUpdateClosedV2		= "closed"
	ZoneInterfacePolicyHashRequiredV2	= "hash_required"
	ZoneInterfacePolicyWalletPolicyV2	= "wallet_policy"
	ZoneInterfacePolicyClosedV2		= "closed"
	ZoneRoutingPolicyExplicitTargetsV2	= "explicit_targets"
	ZoneRoutingPolicyWalletPolicyV2		= "wallet_policy"
	ZoneRoutingPolicyClosedV2		= "closed"
	ZonePolicyWildcardV2			= "*"
)

type SubdomainCreationPolicyV2 struct {
	ParentName		string
	Label			string
	Actor			[]byte
	ChildOwner		[]byte
	Height			uint64
	ChildExpiryHeight	uint64
	DelegationType		SubdomainDelegationTypeV2
	ParentControlsRecord	bool
	DetachedPaid		bool
	IndependentPayment	bool
	ParentAuthorization	bool
	Ephemeral		bool
	TimeLockedUntilHeight	uint64
	Delegation		*DelegationRecordV2
}

type IdentityPathCommitmentV2 struct {
	CommitmentVersion	uint64
	RootName		string
	TargetName		string
	PathLabels		[]string
	PathHashes		[]string
	PathHash		string
	SourceVersion		uint64
	ParentEpoch		uint64
	ChildEpoch		uint64
	CommitmentHash		string
}

type OptimizedRecursiveResolutionProofRequestV2 struct {
	State		IdentityState
	ChainID		string
	RootName	string
	TargetName	string
	Height		uint64
	TTL		uint64
	Cache		*ResolutionCacheRecordV2
	SourceVersion	uint64
	ParentEpoch	uint64
	ChildEpoch	uint64
	LightClient	bool
	ProofVerified	bool
}

type ZonePolicyV2 struct {
	PolicyVersion		uint64
	NameHash		string
	AllowedRecordTypes	[]string
	AllowedServiceTypes	[]string
	SubdomainCreationPolicy	string
	ResolverUpdatePolicy	string
	InterfacePolicy		string
	RoutingPolicy		string
	MaxChildDepth		uint8
	MaxChildRecords		uint32
	LifecycleEpoch		uint64
	UpdatedAtHeight		uint64
	ParentPolicyHash	string
	OverrideParent		bool
	PolicyHash		string
}

type RecursivePolicyProofV2 struct {
	ProofVersion	uint64
	RootName	string
	TargetName	string
	PathCommitment	IdentityPathCommitmentV2
	ZonePolicies	[]ZonePolicyV2
	ProofHash	string
}

func ValidateSubdomainCreationV2(state IdentityState, policy SubdomainCreationPolicyV2) (string, error) {
	state = normalizeIdentityStateParams(state)
	if policy.Height == 0 {
		return "", errors.New("identity v2 subdomain creation height is required")
	}
	parent, err := requireActiveDomain(state, policy.ParentName, policy.Height)
	if err != nil {
		return "", err
	}
	if err := validateDomainLabel(policy.Label); err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 subdomain actor", policy.Actor); err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 subdomain child owner", policy.ChildOwner); err != nil {
		return "", err
	}
	delegationType := policy.DelegationType
	if delegationType == "" {
		delegationType = SubdomainDelegationOwnerControlledV2
	}
	if err := validateSubdomainDelegationTypeV2(delegationType); err != nil {
		return "", err
	}
	childName, err := NormalizeAETDomain(policy.Label + "." + parent.Name)
	if err != nil {
		return "", err
	}
	if !IsDomainAvailable(state, childName, policy.Height) {
		return "", errors.New("identity v2 subdomain already exists")
	}
	if err := validateSubdomainAuthorizationForTypeV2(parent, policy, delegationType); err != nil {
		return "", err
	}
	childExpiry := policy.ChildExpiryHeight
	if childExpiry == 0 {
		childExpiry = parent.ExpiryHeight
	}
	if childExpiry <= policy.Height {
		return "", errors.New("identity v2 subdomain expiry must be after creation height")
	}
	if childExpiry > parent.ExpiryHeight && !policy.DetachedPaid {
		return "", errors.New("identity v2 child expiry cannot exceed parent expiry unless detached mode is enabled")
	}
	if policy.DetachedPaid {
		if delegationType != SubdomainDelegationDetachedPaidV2 {
			return "", errors.New("identity v2 detached mode requires detached_paid delegation type")
		}
		if !policy.IndependentPayment || !policy.ParentAuthorization {
			return "", errors.New("identity v2 detached subdomain requires independent payment and explicit parent authorization")
		}
	}
	if delegationType == SubdomainDelegationEphemeralServiceV2 && !policy.Ephemeral {
		return "", errors.New("identity v2 ephemeral service subdomain must be marked ephemeral")
	}
	if policy.TimeLockedUntilHeight != 0 && policy.TimeLockedUntilHeight >= childExpiry {
		return "", errors.New("identity v2 subdomain time lock must end before child expiry")
	}
	return childName, nil
}

func IssueSubdomainV2(state IdentityState, policy SubdomainCreationPolicyV2) (IdentityState, SubdomainRecord, error) {
	childName, err := ValidateSubdomainCreationV2(state, policy)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	parent, err := requireActiveDomain(state, policy.ParentName, policy.Height)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	childExpiry := policy.ChildExpiryHeight
	if childExpiry == 0 {
		childExpiry = parent.ExpiryHeight
	}
	delegationType := policy.DelegationType
	if delegationType == "" {
		delegationType = SubdomainDelegationOwnerControlledV2
	}
	nftID, err := DomainNFTID(childName)
	if err != nil {
		return IdentityState{}, SubdomainRecord{}, err
	}
	domain := Domain{Name: childName, Owner: cloneSpecAddress(policy.ChildOwner), NFTID: nftID, RegisteredHeight: policy.Height, ExpiryHeight: childExpiry, UpdatedHeight: policy.Height, ParentName: parent.Name, ParentControlsRecord: policy.ParentControlsRecord}
	nft := DomainNFT{ID: nftID, Domain: childName, Owner: cloneSpecAddress(policy.ChildOwner), MintHeight: policy.Height}
	record := SubdomainRecord{
		ParentName:		parent.Name,
		Name:			childName,
		Owner:			cloneSpecAddress(policy.ChildOwner),
		ParentControlsRecord:	policy.ParentControlsRecord,
		CreatedHeight:		policy.Height,
		DelegationType:		delegationType,
		Detached:		policy.DetachedPaid,
		Ephemeral:		policy.Ephemeral,
		ExpiryHeight:		childExpiry,
		TimeLockedUntilHeight:	policy.TimeLockedUntilHeight,
		ParentAuthorized:	policy.ParentAuthorization || bytes.Equal(policy.Actor, parent.Owner),
	}
	next := state.Clone()
	next.Domains = upsertDomain(next.Domains, domain)
	next.DomainNFTs = upsertDomainNFT(next.DomainNFTs, nft)
	next.Subdomains = append(next.Subdomains, record)
	sortIdentityState(&next)
	return next, record, next.Validate()
}

func RevokeDelegationV2(records []DelegationRecordV2, name string, delegate []byte, scope DelegationScopeV2, actor []byte, owner []byte, height uint64) ([]DelegationRecordV2, bool, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return nil, false, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return nil, false, err
	}
	if err := validateSpecAddress("identity v2 delegation revocation actor", actor); err != nil {
		return nil, false, err
	}
	if err := validateSpecAddress("identity v2 delegation revocation owner", owner); err != nil {
		return nil, false, err
	}
	if !bytes.Equal(actor, owner) {
		return nil, false, errors.New("identity v2 delegation revocation requires parent owner")
	}
	out := make([]DelegationRecordV2, 0, len(records))
	revoked := false
	for _, record := range records {
		if record.NameHash == nameHash && bytes.Equal(record.Delegate, delegate) && record.Scope == scope {
			if record.TimeLockedUntilHeight != 0 && height < record.TimeLockedUntilHeight {
				return nil, false, errors.New("identity v2 delegation is time-locked and cannot be revoked yet")
			}
			revoked = true
			continue
		}
		out = append(out, cloneDelegationRecordV2(record))
	}
	return out, revoked, nil
}

func BuildIdentityPathCommitmentV2(path DeterministicResolutionPathV2, sourceVersion uint64, parentEpoch uint64, childEpoch uint64) (IdentityPathCommitmentV2, error) {
	if len(path.Path) == 0 || len(path.PathHashes) == 0 || len(path.Path) != len(path.PathHashes) {
		return IdentityPathCommitmentV2{}, errors.New("identity v2 path commitment requires canonical path and hashes")
	}
	pathHash, err := ComputeResolutionPathHashV2(path.Path)
	if err != nil {
		return IdentityPathCommitmentV2{}, err
	}
	commitment := IdentityPathCommitmentV2{
		CommitmentVersion:	IdentityPathCommitmentVersionV2,
		RootName:		path.Path[0],
		TargetName:		path.TargetName,
		PathLabels:		append([]string(nil), path.Labels...),
		PathHashes:		append([]string(nil), path.PathHashes...),
		PathHash:		pathHash,
		SourceVersion:		sourceVersion,
		ParentEpoch:		parentEpoch,
		ChildEpoch:		childEpoch,
	}
	commitment.CommitmentHash = ComputeIdentityPathCommitmentHashV2(commitment)
	return commitment, ValidateIdentityPathCommitmentV2(commitment)
}

func ValidateIdentityPathCommitmentV2(commitment IdentityPathCommitmentV2) error {
	if commitment.CommitmentVersion != IdentityPathCommitmentVersionV2 {
		return fmt.Errorf("unsupported identity v2 path commitment version %d", commitment.CommitmentVersion)
	}
	if _, err := NormalizeAETDomain(commitment.RootName); err != nil {
		return err
	}
	if _, err := NormalizeAETDomain(commitment.TargetName); err != nil {
		return err
	}
	if len(commitment.PathLabels) == 0 || len(commitment.PathHashes) == 0 {
		return errors.New("identity v2 path commitment labels and hashes are required")
	}
	for _, hash := range commitment.PathHashes {
		if err := validateHexHash("identity v2 path commitment path hash", hash); err != nil {
			return err
		}
	}
	if err := validateHexHash("identity v2 path commitment path_hash", commitment.PathHash); err != nil {
		return err
	}
	if commitment.SourceVersion == 0 {
		return errors.New("identity v2 path commitment source_version is required")
	}
	if commitment.CommitmentHash == "" || commitment.CommitmentHash != ComputeIdentityPathCommitmentHashV2(commitment) {
		return errors.New("identity v2 path commitment hash mismatch")
	}
	return nil
}

func ComputeIdentityPathCommitmentHashV2(commitment IdentityPathCommitmentV2) string {
	parts := []string{
		"identity-v2-path-commitment",
		fmt.Sprintf("%020d", commitment.CommitmentVersion),
		commitment.RootName,
		commitment.TargetName,
		fmt.Sprintf("%020d", len(commitment.PathLabels)),
	}
	parts = append(parts, commitment.PathLabels...)
	parts = append(parts, fmt.Sprintf("%020d", len(commitment.PathHashes)))
	parts = append(parts, commitment.PathHashes...)
	parts = append(parts,
		commitment.PathHash,
		fmt.Sprintf("%020d", commitment.SourceVersion),
		fmt.Sprintf("%020d", commitment.ParentEpoch),
		fmt.Sprintf("%020d", commitment.ChildEpoch),
	)
	return identityHash(parts...)
}

func InvalidateResolutionCacheRecordV2ForParentEpochChange(record ResolutionCacheRecordV2, parentEpoch uint64) ResolutionCacheRecordV2 {
	next := record
	next.ParentEpoch = parentEpoch
	next.ValidUntilHeight = 0
	return next
}

func BuildOptimizedRecursiveResolutionProofV2(request OptimizedRecursiveResolutionProofRequestV2) (RecursiveResolutionProofV2, IdentityPathCommitmentV2, error) {
	path, err := CanonicalResolutionPathV2(request.TargetName)
	if err != nil {
		return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
	}
	commitment, err := BuildIdentityPathCommitmentV2(path, request.SourceVersion, request.ParentEpoch, request.ChildEpoch)
	if err != nil {
		return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
	}
	if request.Cache != nil {
		if err := ValidateResolutionCacheRecordV2Use(*request.Cache, ResolutionCacheUseContextV2{
			Height:		request.Height,
			SourceVersion:	request.SourceVersion,
			ParentEpoch:	request.ParentEpoch,
			ChildEpoch:	request.ChildEpoch,
			LightClient:	request.LightClient,
			ProofVerified:	request.ProofVerified,
		}); err != nil {
			return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
		}
		if request.Cache.ResolutionPathHash != commitment.PathHash {
			return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, errors.New("identity v2 optimized recursive proof cache path commitment mismatch")
		}
	}
	proof, err := BuildRecursiveResolutionProofV2(request.State, request.ChainID, request.RootName, request.TargetName, request.Height, request.TTL, request.Cache)
	if err != nil {
		return RecursiveResolutionProofV2{}, IdentityPathCommitmentV2{}, err
	}
	return proof, commitment, nil
}

func NewZonePolicyV2(name string, policy ZonePolicyV2) (ZonePolicyV2, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return ZonePolicyV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return ZonePolicyV2{}, err
	}
	policy.PolicyVersion = ZonePolicyVersionV2
	policy.NameHash = nameHash
	policy.AllowedRecordTypes = sortStringSet(policy.AllowedRecordTypes)
	policy.AllowedServiceTypes = sortStringSet(policy.AllowedServiceTypes)
	if policy.LifecycleEpoch == 0 {
		policy.LifecycleEpoch = 1
	}
	policy.PolicyHash = ""
	policy.PolicyHash = ComputeZonePolicyHashV2(policy)
	return policy, ValidateZonePolicyV2(policy)
}

func ValidateZonePolicyV2(policy ZonePolicyV2) error {
	if policy.PolicyVersion != ZonePolicyVersionV2 {
		return fmt.Errorf("unsupported identity v2 zone policy version %d", policy.PolicyVersion)
	}
	if err := validateHexHash("identity v2 zone policy name hash", policy.NameHash); err != nil {
		return err
	}
	if err := validateZonePolicyKeySetV2("identity v2 zone allowed_record_types", policy.AllowedRecordTypes); err != nil {
		return err
	}
	if err := validateZonePolicyKeySetV2("identity v2 zone allowed_service_types", policy.AllowedServiceTypes); err != nil {
		return err
	}
	if err := validateZoneSubdomainCreationPolicyV2(policy.SubdomainCreationPolicy); err != nil {
		return err
	}
	if err := validateZoneResolverUpdatePolicyV2(policy.ResolverUpdatePolicy); err != nil {
		return err
	}
	if err := validateZoneInterfacePolicyV2(policy.InterfacePolicy); err != nil {
		return err
	}
	if err := validateZoneRoutingPolicyV2(policy.RoutingPolicy); err != nil {
		return err
	}
	if policy.MaxChildDepth > MaxResolverLabels {
		return fmt.Errorf("identity v2 zone max_child_depth must not exceed %d", MaxResolverLabels)
	}
	if policy.LifecycleEpoch == 0 {
		return errors.New("identity v2 zone lifecycle_epoch is required")
	}
	if policy.UpdatedAtHeight == 0 {
		return errors.New("identity v2 zone updated_at_height is required")
	}
	if policy.ParentPolicyHash != "" {
		if err := validateHexHash("identity v2 zone parent policy hash", policy.ParentPolicyHash); err != nil {
			return err
		}
	}
	if policy.PolicyHash == "" || policy.PolicyHash != ComputeZonePolicyHashV2(policy) {
		return errors.New("identity v2 zone policy hash mismatch")
	}
	return nil
}

func ComputeZonePolicyHashV2(policy ZonePolicyV2) string {
	parts := []string{
		"identity-v2-zone-policy",
		fmt.Sprintf("%020d", policy.PolicyVersion),
		policy.NameHash,
		fmt.Sprintf("%020d", len(policy.AllowedRecordTypes)),
	}
	parts = append(parts, policy.AllowedRecordTypes...)
	parts = append(parts, fmt.Sprintf("%020d", len(policy.AllowedServiceTypes)))
	parts = append(parts, policy.AllowedServiceTypes...)
	parts = append(parts,
		policy.SubdomainCreationPolicy,
		policy.ResolverUpdatePolicy,
		policy.InterfacePolicy,
		policy.RoutingPolicy,
		fmt.Sprintf("%020d", policy.MaxChildDepth),
		fmt.Sprintf("%020d", policy.MaxChildRecords),
		fmt.Sprintf("%020d", policy.LifecycleEpoch),
		fmt.Sprintf("%020d", policy.UpdatedAtHeight),
		policy.ParentPolicyHash,
		fmt.Sprintf("%t", policy.OverrideParent),
	)
	return identityHash(parts...)
}

func ResolveZonePolicyForChildV2(parent ZonePolicyV2, child *ZonePolicyV2, subdomain SubdomainRecord) (ZonePolicyV2, error) {
	if err := ValidateZonePolicyV2(parent); err != nil {
		return ZonePolicyV2{}, err
	}
	if child == nil {
		return cloneZonePolicyV2(parent), nil
	}
	if err := ValidateZonePolicyV2(*child); err != nil {
		return ZonePolicyV2{}, err
	}
	if child.OverrideParent || subdomain.Detached {
		return cloneZonePolicyV2(*child), nil
	}
	return cloneZonePolicyV2(parent), nil
}

func ValidateZonePolicyForSubdomainV2(policy ZonePolicyV2, subdomain SubdomainRecord, childDepth uint8, childRecordCount uint32, recordType string, serviceType string) error {
	if err := ValidateZonePolicyV2(policy); err != nil {
		return err
	}
	if err := validateZonePolicySubdomainRecordV2(subdomain); err != nil {
		return err
	}
	if policy.MaxChildDepth != 0 && childDepth > policy.MaxChildDepth {
		return errors.New("identity v2 zone max_child_depth exceeded")
	}
	if policy.MaxChildRecords != 0 && childRecordCount > policy.MaxChildRecords {
		return errors.New("identity v2 zone max_child_records exceeded")
	}
	if recordType != "" && !zonePolicyAllowsValueV2(policy.AllowedRecordTypes, recordType) {
		return fmt.Errorf("identity v2 zone disallows record type %q", recordType)
	}
	if serviceType != "" && !zonePolicyAllowsValueV2(policy.AllowedServiceTypes, serviceType) {
		return fmt.Errorf("identity v2 zone disallows service type %q", serviceType)
	}
	return nil
}

func ApplyZonePolicyChangeV2(policy ZonePolicyV2, height uint64, dependentCaches []ResolutionCacheRecordV2) (ZonePolicyV2, []ResolutionCacheRecordV2, error) {
	if err := ValidateZonePolicyV2(policy); err != nil {
		return ZonePolicyV2{}, nil, err
	}
	if height == 0 {
		return ZonePolicyV2{}, nil, errors.New("identity v2 zone policy update height is required")
	}
	next := cloneZonePolicyV2(policy)
	next.LifecycleEpoch++
	next.UpdatedAtHeight = height
	next.PolicyHash = ""
	next.PolicyHash = ComputeZonePolicyHashV2(next)
	invalidated := make([]ResolutionCacheRecordV2, 0, len(dependentCaches))
	for _, cache := range dependentCaches {
		invalidated = append(invalidated, InvalidateResolutionCacheRecordV2ForParentEpochChange(cache, next.LifecycleEpoch))
	}
	return next, invalidated, ValidateZonePolicyV2(next)
}

func BuildRecursivePolicyProofV2(rootName string, targetName string, commitment IdentityPathCommitmentV2, policies []ZonePolicyV2) (RecursivePolicyProofV2, error) {
	root, err := NormalizeAETDomain(rootName)
	if err != nil {
		return RecursivePolicyProofV2{}, err
	}
	target, err := NormalizeAETDomain(targetName)
	if err != nil {
		return RecursivePolicyProofV2{}, err
	}
	if commitment.RootName != root || commitment.TargetName != target {
		return RecursivePolicyProofV2{}, errors.New("identity v2 recursive policy proof path commitment mismatch")
	}
	if err := ValidateIdentityPathCommitmentV2(commitment); err != nil {
		return RecursivePolicyProofV2{}, err
	}
	out := RecursivePolicyProofV2{
		ProofVersion:	RecursivePolicyProofVersionV2,
		RootName:	root,
		TargetName:	target,
		PathCommitment:	commitment,
		ZonePolicies:	cloneZonePoliciesV2(policies),
	}
	out.ProofHash = ComputeRecursivePolicyProofHashV2(out)
	return out, ValidateRecursivePolicyProofV2(out)
}

func ValidateRecursivePolicyProofV2(proof RecursivePolicyProofV2) error {
	if proof.ProofVersion != RecursivePolicyProofVersionV2 {
		return fmt.Errorf("unsupported identity v2 recursive policy proof version %d", proof.ProofVersion)
	}
	if _, err := NormalizeAETDomain(proof.RootName); err != nil {
		return err
	}
	if _, err := NormalizeAETDomain(proof.TargetName); err != nil {
		return err
	}
	if err := ValidateIdentityPathCommitmentV2(proof.PathCommitment); err != nil {
		return err
	}
	if len(proof.ZonePolicies) == 0 {
		return errors.New("identity v2 recursive policy proof requires zone policies")
	}
	for _, policy := range proof.ZonePolicies {
		if err := ValidateZonePolicyV2(policy); err != nil {
			return err
		}
	}
	if proof.ProofHash == "" || proof.ProofHash != ComputeRecursivePolicyProofHashV2(proof) {
		return errors.New("identity v2 recursive policy proof hash mismatch")
	}
	return nil
}

func ComputeRecursivePolicyProofHashV2(proof RecursivePolicyProofV2) string {
	parts := []string{
		"identity-v2-recursive-policy-proof",
		fmt.Sprintf("%020d", proof.ProofVersion),
		proof.RootName,
		proof.TargetName,
		proof.PathCommitment.CommitmentHash,
		fmt.Sprintf("%020d", len(proof.ZonePolicies)),
	}
	for _, policy := range proof.ZonePolicies {
		parts = append(parts, ComputeZonePolicyHashV2(policy))
	}
	return identityHash(parts...)
}

func validateSubdomainDelegationTypeV2(value SubdomainDelegationTypeV2) error {
	switch value {
	case SubdomainDelegationOwnerControlledV2,
		SubdomainDelegationDelegateControlledV2,
		SubdomainDelegationZoneManagedV2,
		SubdomainDelegationDetachedPaidV2,
		SubdomainDelegationEphemeralServiceV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 subdomain delegation type %q", value)
	}
}

func validateSubdomainAuthorizationForTypeV2(parent Domain, policy SubdomainCreationPolicyV2, delegationType SubdomainDelegationTypeV2) error {
	if bytes.Equal(policy.Actor, parent.Owner) {
		return nil
	}
	if policy.Delegation == nil {
		return errors.New("identity v2 subdomain creation requires parent owner or scoped delegate")
	}
	parentRecord, err := NewDomainRecordV2FromDomain(parent, DomainRecordV2Active, 0, policy.Height)
	if err != nil {
		return err
	}
	switch delegationType {
	case SubdomainDelegationZoneManagedV2:
		if !bytes.Equal(policy.Actor, policy.Delegation.Delegate) {
			return errors.New("identity v2 zone-managed subdomain delegate mismatch")
		}
		if policy.Delegation.NameHash != parentRecord.NameHash {
			return errors.New("identity v2 zone-managed subdomain delegation name_hash mismatch")
		}
		return ValidateDelegationRecordV2Use(*policy.Delegation, DelegationScopeZoneAdmin, "create", policy.Label, 1, policy.Height)
	default:
		return ValidateSubdomainCreationAuthorizationV2(parentRecord, policy.Actor, policy.Delegation, policy.Label, 1, policy.Height)
	}
}

func validateZonePolicyKeySetV2(field string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s are required", field)
	}
	seen := map[string]struct{}{}
	for i, value := range values {
		if value != ZonePolicyWildcardV2 {
			if err := ValidateResolverMetadataKey(value); err != nil {
				return err
			}
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s value %q", field, value)
		}
		seen[value] = struct{}{}
		if i > 0 && values[i-1] >= value {
			return fmt.Errorf("%s must be sorted canonically", field)
		}
	}
	return nil
}

func validateZoneSubdomainCreationPolicyV2(value string) error {
	switch value {
	case ZoneSubdomainCreationOpenV2, ZoneSubdomainCreationOwnerOnlyV2, ZoneSubdomainCreationDelegatedV2, ZoneSubdomainCreationClosedV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 zone subdomain_creation_policy %q", value)
	}
}

func validateZoneResolverUpdatePolicyV2(value string) error {
	switch value {
	case ZoneResolverUpdateOwnerOnlyV2, ZoneResolverUpdateDelegatedV2, ZoneResolverUpdateClosedV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 zone resolver_update_policy %q", value)
	}
}

func validateZoneInterfacePolicyV2(value string) error {
	switch value {
	case ZoneInterfacePolicyHashRequiredV2, ZoneInterfacePolicyWalletPolicyV2, ZoneInterfacePolicyClosedV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 zone interface_policy %q", value)
	}
}

func validateZoneRoutingPolicyV2(value string) error {
	switch value {
	case ZoneRoutingPolicyExplicitTargetsV2, ZoneRoutingPolicyWalletPolicyV2, ZoneRoutingPolicyClosedV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 zone routing_policy %q", value)
	}
}

func zonePolicyAllowsValueV2(allowed []string, value string) bool {
	if err := ValidateResolverMetadataKey(value); err != nil {
		return false
	}
	for _, candidate := range allowed {
		if candidate == ZonePolicyWildcardV2 || candidate == value {
			return true
		}
	}
	return false
}

func validateZonePolicySubdomainRecordV2(record SubdomainRecord) error {
	parent, err := NormalizeAETDomain(record.ParentName)
	if err != nil {
		return err
	}
	name, err := NormalizeAETDomain(record.Name)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(name, "."+parent) {
		return errors.New("identity v2 zone subdomain must be under parent")
	}
	if err := validateSpecAddress("identity v2 zone subdomain owner", record.Owner); err != nil {
		return err
	}
	if record.CreatedHeight == 0 {
		return errors.New("identity v2 zone subdomain created_height is required")
	}
	if record.DelegationType != "" {
		if err := validateSubdomainDelegationTypeV2(record.DelegationType); err != nil {
			return err
		}
	}
	return nil
}

func cloneZonePolicyV2(policy ZonePolicyV2) ZonePolicyV2 {
	policy.AllowedRecordTypes = append([]string(nil), policy.AllowedRecordTypes...)
	policy.AllowedServiceTypes = append([]string(nil), policy.AllowedServiceTypes...)
	return policy
}

func cloneZonePoliciesV2(policies []ZonePolicyV2) []ZonePolicyV2 {
	out := make([]ZonePolicyV2, 0, len(policies))
	for _, policy := range policies {
		out = append(out, cloneZonePolicyV2(policy))
	}
	return out
}
