package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentityProofObjectiveV2 string

const (
	IdentityProofObjectiveDomainExistence		IdentityProofObjectiveV2	= "domain_existence"
	IdentityProofObjectiveDomainNonExistence	IdentityProofObjectiveV2	= "domain_non_existence"
	IdentityProofObjectiveDomainStatusExpiry	IdentityProofObjectiveV2	= "domain_status_expiry"
	IdentityProofObjectiveNFTBinding		IdentityProofObjectiveV2	= "nft_ownership_binding"
	IdentityProofObjectiveResolverRecord		IdentityProofObjectiveV2	= "resolver_record_contents"
	IdentityProofObjectiveReverseConsistency	IdentityProofObjectiveV2	= "reverse_resolution_consistency"
	IdentityProofObjectiveSubdomainDelegation	IdentityProofObjectiveV2	= "subdomain_delegation_chain"
	IdentityProofObjectiveRecursiveResolution	IdentityProofObjectiveV2	= "recursive_resolution_path"
	IdentityProofObjectiveVersionAndFreshness	IdentityProofObjectiveV2	= "record_version_freshness"
)

type IdentityLightClientProofRequestV2 struct {
	Name			string
	TrustedHeight		uint64
	RecordTTL		uint64
	ReverseAddress		sdk.AccAddress
	AuthorizedAliasKeys	[]string
	Objectives		[]IdentityProofObjectiveV2
}

type IdentityLightClientResolutionProofV2 struct {
	StateRoot		string
	QueryDomain		string
	TrustedHeight		uint64
	FreshUntilHeight	uint64
	RecordVersion		uint64
	DomainStatus		DomainLifecycleStatus
	DomainExpiryHeight	uint64
	Objectives		[]IdentityProofObjectiveV2
	QueryDomainProof	*IdentityInclusionProof
	QueryDomainAbsence	*IdentityAbsenceProof
	ResolutionProof		*IdentityResolutionProof
	ReverseRecord		*ReverseResolutionRecordV2
	ReverseProof		*IdentityInclusionProof
	ReverseAbsence		*IdentityAbsenceProof
	SubdomainChain		[]SubdomainRecord
	SubdomainProofs		[]IdentityInclusionProof
	AuthorizedAliasKeys	[]string
	ProofHash		string
}

func BuildIdentityLightClientResolutionProofV2(state IdentityState, request IdentityLightClientProofRequestV2) (IdentityLightClientResolutionProofV2, error) {
	if request.TrustedHeight == 0 {
		return IdentityLightClientResolutionProofV2{}, errors.New("identity v2 light-client proof trusted height is required")
	}
	if request.RecordTTL == 0 {
		return IdentityLightClientResolutionProofV2{}, errors.New("identity v2 light-client proof record ttl is required")
	}
	query, err := NormalizeAETDomain(request.Name)
	if err != nil {
		return IdentityLightClientResolutionProofV2{}, err
	}
	stateRoot, err := IdentityStateRoot(state)
	if err != nil {
		return IdentityLightClientResolutionProofV2{}, err
	}
	objectives, err := normalizeIdentityProofObjectivesV2(request.Objectives, len(request.ReverseAddress) > 0)
	if err != nil {
		return IdentityLightClientResolutionProofV2{}, err
	}

	proof := IdentityLightClientResolutionProofV2{
		StateRoot:		stateRoot,
		QueryDomain:		query,
		TrustedHeight:		request.TrustedHeight,
		FreshUntilHeight:	request.TrustedHeight + request.RecordTTL,
		Objectives:		objectives,
		AuthorizedAliasKeys:	append([]string(nil), request.AuthorizedAliasKeys...),
	}
	sort.Strings(proof.AuthorizedAliasKeys)

	queryKey, err := IdentityDomainStoreKey(query)
	if err != nil {
		return IdentityLightClientResolutionProofV2{}, err
	}
	if domainProof, err := BuildIdentityProof(state, queryKey); err == nil {
		proof.QueryDomainProof = &domainProof
	} else {
		absence, absenceErr := BuildIdentityAbsenceProof(state, queryKey)
		if absenceErr != nil {
			return IdentityLightClientResolutionProofV2{}, absenceErr
		}
		proof.QueryDomainAbsence = &absence
	}

	status, err := DomainLifecycle(state, query, request.TrustedHeight)
	if err != nil {
		return IdentityLightClientResolutionProofV2{}, err
	}
	proof.DomainStatus = status
	if domain, found := findDomain(state, query); found {
		proof.DomainExpiryHeight = domain.ExpiryHeight
	}

	if requiresResolutionProofObjectiveV2(objectives) {
		resolutionProof, err := BuildIdentityResolutionProof(state, query, request.TrustedHeight)
		if err != nil {
			return IdentityLightClientResolutionProofV2{}, err
		}
		proof.ResolutionProof = &resolutionProof
		proof.RecordVersion = ResolverRecordVersionV2(resolutionProof.Resolver)
		if proof.DomainExpiryHeight == 0 {
			proof.DomainExpiryHeight = resolutionProof.AuthorityDomain.ExpiryHeight
		}
	}
	if proof.RecordVersion == 0 {
		proof.RecordVersion = 1
	}

	if hasIdentityProofObjectiveV2(objectives, IdentityProofObjectiveReverseConsistency) {
		if len(request.ReverseAddress) == 0 {
			return IdentityLightClientResolutionProofV2{}, errors.New("identity v2 reverse proof requires address")
		}
		if err := addReverseProofV2(state, &proof, request.ReverseAddress, request.TrustedHeight); err != nil {
			return IdentityLightClientResolutionProofV2{}, err
		}
	}

	if hasIdentityProofObjectiveV2(objectives, IdentityProofObjectiveSubdomainDelegation) {
		subdomains, subdomainProofs, err := buildSubdomainDelegationProofsV2(state, query)
		if err != nil {
			return IdentityLightClientResolutionProofV2{}, err
		}
		proof.SubdomainChain = subdomains
		proof.SubdomainProofs = subdomainProofs
	}

	proof.ProofHash = ComputeIdentityLightClientResolutionProofHashV2(proof)
	return proof, nil
}

func VerifyIdentityLightClientResolutionProofV2(proof IdentityLightClientResolutionProofV2, trustedRoot string, trustedHeight uint64, currentHeight uint64) error {
	if trustedRoot == "" {
		return errors.New("identity v2 light-client trusted root is required")
	}
	if proof.StateRoot != trustedRoot {
		return errors.New("identity v2 light-client proof root mismatch")
	}
	if proof.TrustedHeight != trustedHeight {
		return errors.New("identity v2 light-client proof trusted height mismatch")
	}
	if currentHeight < trustedHeight {
		return errors.New("identity v2 light-client current height is before trusted height")
	}
	if currentHeight > proof.FreshUntilHeight {
		return errors.New("identity v2 light-client proof is stale")
	}
	objectives, err := normalizeIdentityProofObjectivesV2(proof.Objectives, proof.ReverseRecord != nil)
	if err != nil {
		return err
	}
	if !sameIdentityProofObjectivesV2(objectives, proof.Objectives) {
		return errors.New("identity v2 light-client objectives must be sorted canonically")
	}
	if proof.ProofHash == "" || proof.ProofHash != ComputeIdentityLightClientResolutionProofHashV2(proof) {
		return errors.New("identity v2 light-client proof hash mismatch")
	}
	if err := verifyQueryDomainProofV2(proof); err != nil {
		return err
	}
	if hasIdentityProofObjectiveV2(proof.Objectives, IdentityProofObjectiveDomainExistence) && proof.QueryDomainProof == nil {
		return errors.New("identity v2 domain existence objective requires inclusion proof")
	}
	if hasIdentityProofObjectiveV2(proof.Objectives, IdentityProofObjectiveDomainNonExistence) && proof.QueryDomainAbsence == nil {
		return errors.New("identity v2 domain non-existence objective requires absence proof")
	}
	if hasIdentityProofObjectiveV2(proof.Objectives, IdentityProofObjectiveDomainStatusExpiry) {
		if err := verifyDomainStatusExpiryObjectiveV2(proof, trustedHeight); err != nil {
			return err
		}
	}
	if hasIdentityProofObjectiveV2(proof.Objectives, IdentityProofObjectiveVersionAndFreshness) {
		if proof.RecordVersion == 0 {
			return errors.New("identity v2 record version objective requires record version")
		}
		if proof.FreshUntilHeight <= trustedHeight {
			return errors.New("identity v2 freshness objective requires future fresh_until height")
		}
	}
	if requiresResolutionProofObjectiveV2(proof.Objectives) {
		if proof.ResolutionProof == nil {
			return errors.New("identity v2 recursive resolution objective requires resolution proof")
		}
		resolution, err := ValidateResolutionProofBoundary(*proof.ResolutionProof, trustedRoot, trustedHeight)
		if err != nil {
			return err
		}
		if resolution.QueryDomain != proof.QueryDomain {
			return errors.New("identity v2 resolution proof query mismatch")
		}
		if proof.RecordVersion != ResolverRecordVersionV2(resolution.Record) {
			return errors.New("identity v2 resolver record version mismatch")
		}
	}
	if hasIdentityProofObjectiveV2(proof.Objectives, IdentityProofObjectiveReverseConsistency) {
		if err := verifyReverseConsistencyObjectiveV2(proof, trustedRoot, trustedHeight); err != nil {
			return err
		}
	}
	if hasIdentityProofObjectiveV2(proof.Objectives, IdentityProofObjectiveSubdomainDelegation) {
		if err := verifySubdomainDelegationObjectiveV2(proof, trustedRoot); err != nil {
			return err
		}
	}
	return nil
}

func ResolverRecordVersionV2(record ResolverRecord) uint64 {
	if record.UpdatedAtUnix > 0 {
		return uint64(record.UpdatedAtUnix)
	}
	return 1
}

func ComputeIdentityLightClientResolutionProofHashV2(proof IdentityLightClientResolutionProofV2) string {
	parts := []string{
		"identity-v2-light-client-resolution-proof",
		proof.StateRoot,
		proof.QueryDomain,
		fmt.Sprintf("%020d", proof.TrustedHeight),
		fmt.Sprintf("%020d", proof.FreshUntilHeight),
		fmt.Sprintf("%020d", proof.RecordVersion),
		string(proof.DomainStatus),
		fmt.Sprintf("%020d", proof.DomainExpiryHeight),
	}
	for _, objective := range proof.Objectives {
		parts = append(parts, "objective", string(objective))
	}
	if proof.QueryDomainProof != nil {
		parts = append(parts, "query-domain-proof", proof.QueryDomainProof.LeafHash)
	}
	if proof.QueryDomainAbsence != nil {
		parts = append(parts, "query-domain-absence", proof.QueryDomainAbsence.Key, absenceProofBoundaryHashV2(*proof.QueryDomainAbsence))
	}
	if proof.ResolutionProof != nil {
		parts = append(parts, "resolution", proof.ResolutionProof.StateRoot, proof.ResolutionProof.QueryDomain, proof.ResolutionProof.ResolverDomain, proof.ResolutionProof.ResolverProof.LeafHash)
	}
	if proof.ReverseRecord != nil {
		parts = append(parts, "reverse-record", hex.EncodeToString(proof.ReverseRecord.Address), proof.ReverseRecord.Name, fmt.Sprintf("%t", proof.ReverseRecord.Verified), fmt.Sprintf("%020d", proof.ReverseRecord.ExpiryHeight))
	}
	if proof.ReverseProof != nil {
		parts = append(parts, "reverse-proof", proof.ReverseProof.LeafHash)
	}
	if proof.ReverseAbsence != nil {
		parts = append(parts, "reverse-absence", proof.ReverseAbsence.Key, absenceProofBoundaryHashV2(*proof.ReverseAbsence))
	}
	for _, subdomain := range proof.SubdomainChain {
		parts = append(parts, "subdomain", subdomain.ParentName, subdomain.Name, hex.EncodeToString(subdomain.Owner), fmt.Sprintf("%t", subdomain.ParentControlsRecord), fmt.Sprintf("%020d", subdomain.CreatedHeight), string(subdomain.DelegationType), fmt.Sprintf("%t", subdomain.Detached), fmt.Sprintf("%t", subdomain.Ephemeral), fmt.Sprintf("%020d", subdomain.ExpiryHeight), fmt.Sprintf("%020d", subdomain.TimeLockedUntilHeight), fmt.Sprintf("%t", subdomain.ParentAuthorized))
	}
	for _, subdomainProof := range proof.SubdomainProofs {
		parts = append(parts, "subdomain-proof", subdomainProof.LeafHash)
	}
	for _, key := range proof.AuthorizedAliasKeys {
		parts = append(parts, "alias", key)
	}
	return identityHash(parts...)
}

func normalizeIdentityProofObjectivesV2(objectives []IdentityProofObjectiveV2, includeReverse bool) ([]IdentityProofObjectiveV2, error) {
	if len(objectives) == 0 {
		objectives = []IdentityProofObjectiveV2{
			IdentityProofObjectiveDomainExistence,
			IdentityProofObjectiveDomainStatusExpiry,
			IdentityProofObjectiveNFTBinding,
			IdentityProofObjectiveResolverRecord,
			IdentityProofObjectiveRecursiveResolution,
			IdentityProofObjectiveVersionAndFreshness,
		}
		if includeReverse {
			objectives = append(objectives, IdentityProofObjectiveReverseConsistency)
		}
	}
	out := append([]IdentityProofObjectiveV2(nil), objectives...)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	for i, objective := range out {
		if !IsIdentityProofObjectiveV2(objective) {
			return nil, fmt.Errorf("unknown identity v2 proof objective %q", objective)
		}
		if i > 0 && out[i-1] == objective {
			return nil, fmt.Errorf("duplicate identity v2 proof objective %q", objective)
		}
	}
	return out, nil
}

func IsIdentityProofObjectiveV2(objective IdentityProofObjectiveV2) bool {
	switch objective {
	case IdentityProofObjectiveDomainExistence,
		IdentityProofObjectiveDomainNonExistence,
		IdentityProofObjectiveDomainStatusExpiry,
		IdentityProofObjectiveNFTBinding,
		IdentityProofObjectiveResolverRecord,
		IdentityProofObjectiveReverseConsistency,
		IdentityProofObjectiveSubdomainDelegation,
		IdentityProofObjectiveRecursiveResolution,
		IdentityProofObjectiveVersionAndFreshness:
		return true
	default:
		return false
	}
}

func hasIdentityProofObjectiveV2(objectives []IdentityProofObjectiveV2, objective IdentityProofObjectiveV2) bool {
	for _, candidate := range objectives {
		if candidate == objective {
			return true
		}
	}
	return false
}

func requiresResolutionProofObjectiveV2(objectives []IdentityProofObjectiveV2) bool {
	for _, objective := range objectives {
		switch objective {
		case IdentityProofObjectiveNFTBinding,
			IdentityProofObjectiveResolverRecord,
			IdentityProofObjectiveReverseConsistency,
			IdentityProofObjectiveRecursiveResolution:
			return true
		}
	}
	return false
}

func sameIdentityProofObjectivesV2(left []IdentityProofObjectiveV2, right []IdentityProofObjectiveV2) bool {
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

func addReverseProofV2(state IdentityState, proof *IdentityLightClientResolutionProofV2, address sdk.AccAddress, height uint64) error {
	if err := validateSpecAddress("identity v2 reverse proof address", address); err != nil {
		return err
	}
	reverseKey, err := IdentityReverseStoreKey(address)
	if err != nil {
		return err
	}
	for _, reverse := range state.ReverseRecords {
		if !bytes.Equal(reverse.Address, address) {
			continue
		}
		reverseProof, err := BuildIdentityProof(state, reverseKey)
		if err != nil {
			return err
		}
		reverseRecord, err := reverseRecordV2FromLegacy(state, reverse, true)
		if err != nil {
			return err
		}
		if err := ValidateReverseResolutionRecordV2(state, reverseRecord, height, proof.AuthorizedAliasKeys); err != nil {
			return err
		}
		proof.ReverseRecord = &reverseRecord
		proof.ReverseProof = &reverseProof
		return nil
	}
	absence, err := BuildIdentityAbsenceProof(state, reverseKey)
	if err != nil {
		return err
	}
	proof.ReverseAbsence = &absence
	return nil
}

func buildSubdomainDelegationProofsV2(state IdentityState, query string) ([]SubdomainRecord, []IdentityInclusionProof, error) {
	records := make([]SubdomainRecord, 0)
	proofs := make([]IdentityInclusionProof, 0)
	for _, record := range state.Subdomains {
		if record.Name != query && !strings.HasSuffix(query, "."+record.Name) {
			continue
		}
		key, err := IdentitySubdomainIndexKey(record.ParentName, record.Name)
		if err != nil {
			return nil, nil, err
		}
		proof, err := BuildIdentityProof(state, key)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, cloneSubdomain(record))
		proofs = append(proofs, proof)
	}
	sort.SliceStable(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	sort.SliceStable(proofs, func(i, j int) bool { return proofs[i].Key < proofs[j].Key })
	return records, proofs, nil
}

func verifyQueryDomainProofV2(proof IdentityLightClientResolutionProofV2) error {
	if proof.QueryDomainProof == nil && proof.QueryDomainAbsence == nil {
		return errors.New("identity v2 light-client proof requires query domain inclusion or absence")
	}
	expectedKey, err := IdentityDomainStoreKey(proof.QueryDomain)
	if err != nil {
		return err
	}
	if proof.QueryDomainProof != nil {
		if err := VerifyIdentityProof(*proof.QueryDomainProof); err != nil {
			return err
		}
		if proof.QueryDomainProof.RootHash != proof.StateRoot || proof.QueryDomainProof.Key != expectedKey {
			return errors.New("identity v2 query domain proof mismatch")
		}
	}
	if proof.QueryDomainAbsence != nil {
		if err := VerifyIdentityAbsenceProof(*proof.QueryDomainAbsence); err != nil {
			return err
		}
		if proof.QueryDomainAbsence.RootHash != proof.StateRoot || proof.QueryDomainAbsence.Key != expectedKey {
			return errors.New("identity v2 query domain absence mismatch")
		}
	}
	return nil
}

func verifyDomainStatusExpiryObjectiveV2(proof IdentityLightClientResolutionProofV2, trustedHeight uint64) error {
	switch proof.DomainStatus {
	case DomainLifecycleActive, DomainLifecycleRenewalWindow:
		if proof.DomainExpiryHeight <= trustedHeight {
			return errors.New("identity v2 domain status proof has expired domain")
		}
	case DomainLifecycleExpired:
		if proof.DomainExpiryHeight > trustedHeight {
			return errors.New("identity v2 domain status proof expiry mismatch")
		}
	case DomainLifecycleAvailable, DomainLifecycleCommitted:
		return nil
	default:
		return fmt.Errorf("unknown identity v2 domain lifecycle status %q", proof.DomainStatus)
	}
	return nil
}

func verifyReverseConsistencyObjectiveV2(proof IdentityLightClientResolutionProofV2, trustedRoot string, trustedHeight uint64) error {
	if proof.ReverseRecord == nil {
		return errors.New("identity v2 reverse consistency objective requires reverse record")
	}
	if proof.ReverseProof == nil {
		return errors.New("identity v2 reverse consistency objective requires reverse inclusion proof")
	}
	if err := VerifyIdentityProof(*proof.ReverseProof); err != nil {
		return err
	}
	if proof.ReverseProof.RootHash != trustedRoot {
		return errors.New("identity v2 reverse proof root mismatch")
	}
	legacy := ReverseRecord{Address: cloneSpecAddress(proof.ReverseRecord.Address), Domain: proof.ReverseRecord.Name, UpdatedAtUnix: int64(proof.ReverseRecord.UpdatedAtHeight)}
	expectedLeaf, err := identityReverseLeaf(legacy)
	if err != nil {
		return err
	}
	if proof.ReverseProof.Key != expectedLeaf.Key || proof.ReverseProof.ValueHash != expectedLeaf.ValueHash {
		return errors.New("identity v2 reverse proof value mismatch")
	}
	if !proof.ReverseRecord.Verified {
		return errors.New("identity v2 reverse consistency requires verified reverse record")
	}
	if proof.ReverseRecord.ExpiryHeight <= trustedHeight {
		return errors.New("identity v2 reverse consistency proof is expired")
	}
	if proof.ResolutionProof == nil {
		return errors.New("identity v2 reverse consistency requires resolution proof")
	}
	if !forwardResolutionContainsAddress(proof.ResolutionProof.Resolver, proof.ReverseRecord.Address, proof.AuthorizedAliasKeys) {
		return errors.New("identity v2 reverse consistency requires forward primary or authorized alias")
	}
	if proof.ReverseRecord.ExpiryHeight > proof.ResolutionProof.AuthorityDomain.ExpiryHeight {
		return errors.New("identity v2 reverse consistency expires after domain")
	}
	return nil
}

func verifySubdomainDelegationObjectiveV2(proof IdentityLightClientResolutionProofV2, trustedRoot string) error {
	if len(proof.SubdomainChain) != len(proof.SubdomainProofs) {
		return errors.New("identity v2 subdomain delegation proof count mismatch")
	}
	for i, record := range proof.SubdomainChain {
		inclusion := proof.SubdomainProofs[i]
		if err := VerifyIdentityProof(inclusion); err != nil {
			return err
		}
		if inclusion.RootHash != trustedRoot {
			return errors.New("identity v2 subdomain delegation proof root mismatch")
		}
		expectedLeaf, err := identitySubdomainLeaf(record)
		if err != nil {
			return err
		}
		if inclusion.Key != expectedLeaf.Key || inclusion.ValueHash != expectedLeaf.ValueHash {
			return errors.New("identity v2 subdomain delegation proof value mismatch")
		}
	}
	return nil
}

func absenceProofBoundaryHashV2(proof IdentityAbsenceProof) string {
	parts := []string{"identity-v2-absence-boundary", proof.RootHash, proof.Key}
	if proof.Previous != nil {
		parts = append(parts, "prev", proof.Previous.Key, proof.Previous.LeafHash)
	}
	if proof.Next != nil {
		parts = append(parts, "next", proof.Next.Key, proof.Next.LeafHash)
	}
	return identityHash(parts...)
}
