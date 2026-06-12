package types

import (
	"errors"
	"fmt"
	"strings"
)

type IdentityResolutionTargetTypeV2 string

const (
	IdentityResolutionTargetPrimary		IdentityResolutionTargetTypeV2	= "primary"
	IdentityResolutionTargetContract	IdentityResolutionTargetTypeV2	= "contract"
	IdentityResolutionTargetService		IdentityResolutionTargetTypeV2	= "service"
	IdentityResolutionTargetInterface	IdentityResolutionTargetTypeV2	= "interface"
	IdentityResolutionTargetRoute		IdentityResolutionTargetTypeV2	= "route"
	IdentityResolutionTargetRecord		IdentityResolutionTargetTypeV2	= "record"
)

type DeterministicResolutionPathV2 struct {
	TargetName	string
	Labels		[]string
	Path		[]string
	PathHashes	[]string
}

type DeterministicResolutionStepV2 struct {
	Name		string
	NameHash	string
	DomainRecord	*DomainRecordV2
	ResolverRecord	*UnifiedResolutionRecordV2
	Delegation	*DelegationRecordV2
	Exists		bool
	Delegated	bool
}

type DeterministicResolutionPathValidationV2 struct {
	TargetType	IdentityResolutionTargetTypeV2
	TargetKey	string
	Height		uint64
	RecordTTL	uint64
	MaxRecordAge	uint64
	Delegations	[]DelegationRecordV2
}

type DeterministicResolutionPathResultV2 struct {
	Path		DeterministicResolutionPathV2
	Steps		[]DeterministicResolutionStepV2
	Resolution	IdentityResolution
}

func CanonicalResolutionPathV2(name string) (DeterministicResolutionPathV2, error) {
	normalized, err := NormalizeAETDomainVersioned(name, NameNormalizationVersionV2)
	if err != nil {
		return DeterministicResolutionPathV2{}, err
	}
	labels := normalized.Labels
	path := make([]string, 0, len(labels))
	pathLabels := make([]string, 0, len(labels))
	pathHashes := make([]string, 0, len(labels))
	for i := len(labels) - 1; i >= 0; i-- {
		pathLabels = append(pathLabels, labels[i])
		candidate := strings.Join(labels[i:], ".") + DomainTLD
		nameHash, err := DomainRecordV2NameHash(candidate)
		if err != nil {
			return DeterministicResolutionPathV2{}, err
		}
		path = append(path, candidate)
		pathHashes = append(pathHashes, nameHash)
	}
	return DeterministicResolutionPathV2{
		TargetName:	normalized.NormalizedName,
		Labels:		pathLabels,
		Path:		path,
		PathHashes:	pathHashes,
	}, nil
}

func VerifyDeterministicResolutionPathV2(state IdentityState, name string, opts DeterministicResolutionPathValidationV2) (DeterministicResolutionPathResultV2, error) {
	state = normalizeIdentityStateParams(state)
	if err := state.Validate(); err != nil {
		return DeterministicResolutionPathResultV2{}, err
	}
	if opts.Height == 0 {
		return DeterministicResolutionPathResultV2{}, errors.New("identity v2 deterministic resolution height is required")
	}
	if opts.TargetType == "" {
		return DeterministicResolutionPathResultV2{}, errors.New("identity v2 deterministic resolution target type is required")
	}
	path, err := CanonicalResolutionPathV2(name)
	if err != nil {
		return DeterministicResolutionPathResultV2{}, err
	}
	steps := make([]DeterministicResolutionStepV2, 0, len(path.Path))
	for i, candidate := range path.Path {
		step := DeterministicResolutionStepV2{Name: candidate, NameHash: path.PathHashes[i]}
		if domain, found := findDomain(state, candidate); found {
			if domain.ExpiryHeight <= opts.Height {
				return DeterministicResolutionPathResultV2{}, fmt.Errorf("identity v2 deterministic path domain %q is expired", candidate)
			}
			if err := verifyDeterministicPathParentV2(path, i, domain, state, opts); err != nil {
				return DeterministicResolutionPathResultV2{}, err
			}
			record, err := NewDomainRecordV2FromDomain(domain, DomainRecordV2Active, 0, opts.Height)
			if err != nil {
				return DeterministicResolutionPathResultV2{}, err
			}
			step.DomainRecord = &record
			step.Exists = true
		} else {
			delegation, err := deterministicPathDelegationV2(path, i, opts)
			if err != nil {
				return DeterministicResolutionPathResultV2{}, err
			}
			if delegation == nil {
				if i == 0 {
					return DeterministicResolutionPathResultV2{}, fmt.Errorf("identity v2 deterministic path root %q is not active", candidate)
				}
				return DeterministicResolutionPathResultV2{}, fmt.Errorf("identity v2 deterministic path %q requires parent domain or delegation", candidate)
			}
			step.Delegation = delegation
			step.Delegated = true
		}
		if resolver, found := findResolver(state, candidate); found {
			record, err := BuildUnifiedResolutionRecordV2(state, candidate, opts.Height, deterministicPathTTLV2(opts))
			if err != nil {
				return DeterministicResolutionPathResultV2{}, err
			}
			record.RecordVersion = ResolverRecordVersionV2(resolver)
			step.ResolverRecord = &record
		}
		steps = append(steps, step)
	}
	resolution, err := ResolveIdentityRecordRecursive(state, path.TargetName, opts.Height)
	if err != nil {
		return DeterministicResolutionPathResultV2{}, err
	}
	if err := verifyDeterministicResolverFreshnessV2(resolution.Record, opts); err != nil {
		return DeterministicResolutionPathResultV2{}, err
	}
	if err := verifyDeterministicTargetExistsV2(state, path.TargetName, resolution, opts); err != nil {
		return DeterministicResolutionPathResultV2{}, err
	}
	return DeterministicResolutionPathResultV2{Path: path, Steps: steps, Resolution: resolution}, nil
}

func verifyDeterministicPathParentV2(path DeterministicResolutionPathV2, index int, domain Domain, state IdentityState, opts DeterministicResolutionPathValidationV2) error {
	if index == 0 {
		if domain.ParentName != "" {
			return fmt.Errorf("identity v2 deterministic path root %q must not have a parent", domain.Name)
		}
		return nil
	}
	expectedParent := path.Path[index-1]
	parentHash, err := DomainRecordV2ParentNameHash(domain.Name)
	if err != nil {
		return err
	}
	if parentHash != path.PathHashes[index-1] {
		return fmt.Errorf("identity v2 deterministic path parent hash mismatch for %q", domain.Name)
	}
	if domain.ParentName != "" {
		parent, err := NormalizeAETDomainVersioned(domain.ParentName, NameNormalizationVersionV2)
		if err != nil {
			return err
		}
		if parent.NormalizedName != expectedParent {
			return fmt.Errorf("identity v2 deterministic path parent mismatch for %q", domain.Name)
		}
		return nil
	}
	if parent, found := findDomain(state, expectedParent); found && addressesEqual(parent.Owner, domain.Owner) {
		return nil
	}
	delegation, err := deterministicPathDelegationV2(path, index, opts)
	if err != nil {
		return err
	}
	if delegation != nil {
		return nil
	}
	return fmt.Errorf("identity v2 deterministic path child %q requires matching owner or delegation", domain.Name)
}

func deterministicPathDelegationV2(path DeterministicResolutionPathV2, index int, opts DeterministicResolutionPathValidationV2) (*DelegationRecordV2, error) {
	if index == 0 {
		return nil, nil
	}
	for ancestor := 0; ancestor < index; ancestor++ {
		depth := uint8(index - ancestor)
		recordKey := strings.Join(path.Labels[ancestor+1:index+1], ".")
		for _, delegation := range opts.Delegations {
			if delegation.NameHash != path.PathHashes[ancestor] {
				continue
			}
			if err := ValidateDelegationRecordV2Use(delegation, DelegationScopeSubdomainCreate, "create", recordKey, depth, opts.Height); err == nil {
				copied := delegation
				return &copied, nil
			}
			if err := ValidateDelegationRecordV2Use(delegation, DelegationScopeZoneAdmin, "resolve", recordKey, depth, opts.Height); err == nil {
				copied := delegation
				return &copied, nil
			}
		}
	}
	return nil, nil
}

func verifyDeterministicResolverFreshnessV2(record ResolverRecord, opts DeterministicResolutionPathValidationV2) error {
	if record.UpdatedAtUnix <= 0 {
		return errors.New("identity v2 deterministic resolver updated_at_height is required")
	}
	updated := uint64(record.UpdatedAtUnix)
	if updated > opts.Height {
		return errors.New("identity v2 deterministic resolver update height is in the future")
	}
	if opts.MaxRecordAge > 0 && opts.Height-updated > opts.MaxRecordAge {
		return errors.New("identity v2 deterministic resolver record is stale")
	}
	return nil
}

func verifyDeterministicTargetExistsV2(state IdentityState, name string, resolution IdentityResolution, opts DeterministicResolutionPathValidationV2) error {
	switch opts.TargetType {
	case IdentityResolutionTargetPrimary:
		if len(resolution.Record.Primary) == 0 {
			return errors.New("identity v2 deterministic primary target is not resolved")
		}
	case IdentityResolutionTargetContract:
		if len(resolution.Record.Contract) == 0 {
			return errors.New("identity v2 deterministic contract target is not resolved")
		}
	case IdentityResolutionTargetRecord:
		if opts.TargetKey == "" {
			return errors.New("identity v2 deterministic record target key is required")
		}
		if opts.TargetKey == ResolverKeyPrimary {
			return verifyDeterministicTargetExistsV2(state, name, resolution, DeterministicResolutionPathValidationV2{TargetType: IdentityResolutionTargetPrimary})
		}
		if opts.TargetKey == ResolverKeyContract {
			return verifyDeterministicTargetExistsV2(state, name, resolution, DeterministicResolutionPathValidationV2{TargetType: IdentityResolutionTargetContract})
		}
		if len(resolution.Record.Records[opts.TargetKey]) == 0 {
			return fmt.Errorf("identity v2 deterministic record target %q is not resolved", opts.TargetKey)
		}
	case IdentityResolutionTargetService, IdentityResolutionTargetInterface, IdentityResolutionTargetRoute:
		record, err := BuildUnifiedResolutionRecordV2(state, name, opts.Height, deterministicPathTTLV2(opts))
		if err != nil {
			return err
		}
		return verifyUnifiedTargetExistsV2(record, opts)
	default:
		return fmt.Errorf("unsupported identity v2 deterministic target type %q", opts.TargetType)
	}
	return nil
}

func verifyUnifiedTargetExistsV2(record UnifiedResolutionRecordV2, opts DeterministicResolutionPathValidationV2) error {
	switch opts.TargetType {
	case IdentityResolutionTargetService:
		if opts.TargetKey == "" {
			return errors.New("identity v2 deterministic service target key is required")
		}
		for _, endpoint := range record.ServiceEndpoints {
			if serviceEndpointIDV2(endpoint) == opts.TargetKey {
				return nil
			}
		}
		return fmt.Errorf("identity v2 deterministic service target %q is not resolved", opts.TargetKey)
	case IdentityResolutionTargetInterface:
		if opts.TargetKey == "" {
			return errors.New("identity v2 deterministic interface target key is required")
		}
		for _, descriptor := range record.InterfaceDescriptors {
			if descriptor.InterfaceID == opts.TargetKey {
				return nil
			}
		}
		return fmt.Errorf("identity v2 deterministic interface target %q is not resolved", opts.TargetKey)
	case IdentityResolutionTargetRoute:
		if !routingMetadataHasTargetV2(record.RoutingMetadata) {
			return errors.New("identity v2 deterministic route target is not resolved")
		}
	default:
		return fmt.Errorf("unsupported identity v2 deterministic unified target type %q", opts.TargetType)
	}
	return nil
}

func deterministicPathTTLV2(opts DeterministicResolutionPathValidationV2) uint64 {
	if opts.RecordTTL != 0 {
		return opts.RecordTTL
	}
	return 1
}
