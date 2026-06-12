package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentityOperationKind string

const (
	IdentityOperationRegister	IdentityOperationKind	= "register"
	IdentityOperationRenew		IdentityOperationKind	= "renew"
	IdentityOperationTransfer	IdentityOperationKind	= "transfer"
	IdentityOperationResolverUpdate	IdentityOperationKind	= "resolver_update"
)

type IdentityFlowStep string

const (
	IdentityFlowStepDeterministicValidation	IdentityFlowStep	= "deterministic_validation"
	IdentityFlowStepRegistryNFTCheck	IdentityFlowStep	= "registry_state_nft_ownership_check"
	IdentityFlowStepResolverStateUpdate	IdentityFlowStep	= "resolver_state_update"
	IdentityFlowStepStoreV2Writes		IdentityFlowStep	= "proof_indexed_store_v2_writes"
	IdentityFlowStepEventsProofsRouting	IdentityFlowStep	= "events_queryable_proofs_execution_routing_hooks"
)

type IdentityDataFlowRequest struct {
	Operation	IdentityOperationKind
	Domain		string
	Actor		sdk.AccAddress
	NewOwner	sdk.AccAddress
	ResolverPatch	*ResolverPatch
}

type IdentityDataFlow struct {
	Operation	IdentityOperationKind
	Domain		string
	Actor		sdk.AccAddress
	Steps		[]IdentityFlowStep
	AccessSet	IdentityAccessSet
	StoreWrites	[]string
	EventTypes	[]string
	ProofRoot	string
	RoutingHooks	[]IdentityV2Component
}

type IdentityBoundaryStatus string

const (
	IdentityBoundaryTrusted		IdentityBoundaryStatus	= "trusted"
	IdentityBoundaryAdvisory	IdentityBoundaryStatus	= "advisory"
	IdentityBoundaryRejected	IdentityBoundaryStatus	= "rejected"
)

type IdentityTrustBoundaryReport struct {
	RegistryNFTAuthority	IdentityBoundaryStatus
	ResolverAuthority	IdentityBoundaryStatus
	ServiceMetadata		IdentityBoundaryStatus
	InterfaceDescriptor	IdentityBoundaryStatus
	LightClientProof	IdentityBoundaryStatus
	CacheFreshness		IdentityBoundaryStatus
}

func PlanIdentityDataFlow(state IdentityState, request IdentityDataFlowRequest, height uint64) (IdentityDataFlow, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return IdentityDataFlow{}, err
	}
	normalized, err := NormalizeAETDomain(request.Domain)
	if err != nil {
		return IdentityDataFlow{}, err
	}
	if err := validateSpecAddress("identity flow actor", request.Actor); err != nil {
		return IdentityDataFlow{}, err
	}
	flow := IdentityDataFlow{
		Operation:	request.Operation,
		Domain:		normalized,
		Actor:		cloneSpecAddress(request.Actor),
		Steps: []IdentityFlowStep{
			IdentityFlowStepDeterministicValidation,
			IdentityFlowStepRegistryNFTCheck,
			IdentityFlowStepStoreV2Writes,
			IdentityFlowStepEventsProofsRouting,
		},
	}
	switch request.Operation {
	case IdentityOperationRegister:
		if !IsDomainAvailable(state, normalized, height) {
			return IdentityDataFlow{}, errors.New("identity flow register domain is not available")
		}
		accessSet, writes, err := registerFlowAccessSet(normalized, request.Actor)
		if err != nil {
			return IdentityDataFlow{}, err
		}
		flow.AccessSet = accessSet
		flow.StoreWrites = writes
		flow.EventTypes = []string{"identity_domain_registered"}
	case IdentityOperationRenew:
		domain, nft, err := ValidateRegistryNFTAuthority(state, normalized, height)
		if err != nil {
			return IdentityDataFlow{}, err
		}
		if !addressesEqual(request.Actor, domain.Owner) {
			return IdentityDataFlow{}, errors.New("identity flow renew requires owner")
		}
		accessSet, writes, err := domainNFTFlowAccessSet(domain, nft)
		if err != nil {
			return IdentityDataFlow{}, err
		}
		flow.AccessSet = accessSet
		flow.StoreWrites = writes
		flow.EventTypes = []string{"identity_domain_renewed"}
	case IdentityOperationTransfer:
		domain, nft, err := ValidateRegistryNFTAuthority(state, normalized, height)
		if err != nil {
			return IdentityDataFlow{}, err
		}
		if !addressesEqual(request.Actor, domain.Owner) {
			return IdentityDataFlow{}, errors.New("identity flow transfer requires owner")
		}
		if err := validateSpecAddress("identity flow new owner", request.NewOwner); err != nil {
			return IdentityDataFlow{}, err
		}
		accessSet, writes, err := transferFlowAccessSet(domain, nft, request.NewOwner)
		if err != nil {
			return IdentityDataFlow{}, err
		}
		flow.AccessSet = accessSet
		flow.StoreWrites = writes
		flow.EventTypes = []string{"identity_domain_transferred", "identity_domain_nft_transferred"}
	case IdentityOperationResolverUpdate:
		if request.ResolverPatch == nil {
			return IdentityDataFlow{}, errors.New("identity flow resolver patch is required")
		}
		if _, err := ValidateResolverControlBoundary(state, normalized, request.Actor, height); err != nil {
			return IdentityDataFlow{}, err
		}
		accessSet, err := IdentityResolverPatchAccessSet(normalized)
		if err != nil {
			return IdentityDataFlow{}, err
		}
		flow.Steps = insertResolverFlowStep(flow.Steps)
		flow.AccessSet = accessSet
		flow.StoreWrites = append([]string(nil), accessSet.Writes...)
		flow.EventTypes = []string{"identity_resolver_updated"}
		flow.RoutingHooks = []IdentityV2Component{IdentityV2RoutingIntegration}
	default:
		return IdentityDataFlow{}, fmt.Errorf("unsupported identity flow operation %q", request.Operation)
	}
	root, err := IdentityStateRoot(state)
	if err != nil {
		return IdentityDataFlow{}, err
	}
	flow.ProofRoot = root
	flow.StoreWrites = sortedUniqueStrings(flow.StoreWrites)
	return flow, nil
}

func ValidateRegistryNFTAuthority(state IdentityState, name string, height uint64) (Domain, DomainNFT, error) {
	domain, err := requireActiveDomain(state, name, height)
	if err != nil {
		return Domain{}, DomainNFT{}, err
	}
	nft, found := findDomainNFTByID(state, domain.NFTID)
	if !found {
		return Domain{}, DomainNFT{}, errors.New("identity registry nft not found")
	}
	if nft.Domain != domain.Name || nft.ID != domain.NFTID || !addressesEqual(nft.Owner, domain.Owner) {
		return Domain{}, DomainNFT{}, errors.New("identity registry and nft ownership mismatch")
	}
	return domain, nft, nil
}

func ValidateResolverControlBoundary(state IdentityState, resolverDomain string, actor sdk.AccAddress, height uint64) (Domain, error) {
	normalized, err := NormalizeAETDomain(resolverDomain)
	if err != nil {
		return Domain{}, err
	}
	authority, err := requireResolverAuthorityDomain(state, normalized, height)
	if err != nil {
		return Domain{}, err
	}
	if _, _, err := ValidateRegistryNFTAuthority(state, authority.Name, height); err != nil {
		return Domain{}, err
	}
	if err := canControlResolver(state, authority, actor); err != nil {
		return Domain{}, err
	}
	return authority, nil
}

func ValidateResolutionProofBoundary(proof IdentityResolutionProof, trustedRoot string, height uint64) (IdentityResolution, error) {
	if trustedRoot == "" {
		return IdentityResolution{}, errors.New("identity trusted proof root is required")
	}
	if proof.StateRoot != trustedRoot {
		return IdentityResolution{}, errors.New("identity resolution proof does not match trusted root")
	}
	return VerifyIdentityResolutionProof(proof, height)
}

func ValidateResolutionCacheFreshness(cachedHeight uint64, currentHeight uint64, maxAge uint64) error {
	if cachedHeight == 0 {
		return errors.New("identity cached resolution height is required")
	}
	if currentHeight < cachedHeight {
		return errors.New("identity cached resolution height is from the future")
	}
	if currentHeight-cachedHeight > maxAge {
		return errors.New("identity cached resolution is stale")
	}
	return nil
}

func EvaluateIdentityTrustBoundaries(state IdentityState, name string, actor sdk.AccAddress, height uint64, proof *IdentityResolutionProof, cachedHeight uint64, maxCacheAge uint64) IdentityTrustBoundaryReport {
	report := IdentityTrustBoundaryReport{
		ServiceMetadata:	IdentityBoundaryAdvisory,
		InterfaceDescriptor:	IdentityBoundaryAdvisory,
		LightClientProof:	IdentityBoundaryAdvisory,
		CacheFreshness:		IdentityBoundaryAdvisory,
	}
	if _, _, err := ValidateRegistryNFTAuthority(state, name, height); err != nil {
		report.RegistryNFTAuthority = IdentityBoundaryRejected
	} else {
		report.RegistryNFTAuthority = IdentityBoundaryTrusted
	}
	if _, err := ValidateResolverControlBoundary(state, name, actor, height); err != nil {
		report.ResolverAuthority = IdentityBoundaryRejected
	} else {
		report.ResolverAuthority = IdentityBoundaryTrusted
	}
	if proof != nil {
		root, err := IdentityStateRoot(state)
		if err == nil {
			if _, err := ValidateResolutionProofBoundary(*proof, root, height); err != nil {
				report.LightClientProof = IdentityBoundaryRejected
			} else {
				report.LightClientProof = IdentityBoundaryTrusted
			}
		}
	}
	if cachedHeight != 0 {
		if err := ValidateResolutionCacheFreshness(cachedHeight, height, maxCacheAge); err != nil {
			report.CacheFreshness = IdentityBoundaryRejected
		} else {
			report.CacheFreshness = IdentityBoundaryTrusted
		}
	}
	return report
}

func registerFlowAccessSet(domain string, owner sdk.AccAddress) (IdentityAccessSet, []string, error) {
	domainKey, err := IdentityDomainStoreKey(domain)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	nftID, err := DomainNFTID(domain)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	nftKey, err := IdentityNFTStoreKey(nftID)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	ownerKey, err := IdentityOwnerDomainIndexKey(owner, domain)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	writes := []string{domainKey, nftKey, ownerKey}
	return newIdentityAccessSet([]string{domainKey}, writes), writes, nil
}

func domainNFTFlowAccessSet(domain Domain, nft DomainNFT) (IdentityAccessSet, []string, error) {
	domainKey, err := IdentityDomainStoreKey(domain.Name)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	nftKey, err := IdentityNFTStoreKey(nft.ID)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	writes := []string{domainKey}
	return newIdentityAccessSet([]string{domainKey, nftKey}, writes), writes, nil
}

func transferFlowAccessSet(domain Domain, nft DomainNFT, newOwner sdk.AccAddress) (IdentityAccessSet, []string, error) {
	domainKey, err := IdentityDomainStoreKey(domain.Name)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	nftKey, err := IdentityNFTStoreKey(nft.ID)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	oldOwnerKey, err := IdentityOwnerDomainIndexKey(domain.Owner, domain.Name)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	newOwnerKey, err := IdentityOwnerDomainIndexKey(newOwner, domain.Name)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	writes := []string{domainKey, nftKey, oldOwnerKey, newOwnerKey}
	return newIdentityAccessSet([]string{domainKey, nftKey}, writes), writes, nil
}

func insertResolverFlowStep(steps []IdentityFlowStep) []IdentityFlowStep {
	out := make([]IdentityFlowStep, 0, len(steps)+1)
	for _, step := range steps {
		if step == IdentityFlowStepStoreV2Writes {
			out = append(out, IdentityFlowStepResolverStateUpdate)
		}
		out = append(out, step)
	}
	return out
}

func addressesEqual(left sdk.AccAddress, right sdk.AccAddress) bool {
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
