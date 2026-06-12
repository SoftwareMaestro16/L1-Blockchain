package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	IdentityObservabilityLookupVolumeV2	= "identity_lookup_volume"
	IdentityABCIEventMalformedRejectedV2	= "identity_malformed_rejected"
	IdentityABCIEventProposalGroupV2	= "identity_proposal_group"
	IdentityABCIEventDomainExpiredV2	= "identity_domain_expired"
	IdentityABCIEventCacheInvalidatedV2	= "identity_cache_invalidated"
	IdentityABCIEventVoteTelemetryV2	= "identity_vote_telemetry"
	IdentityProposalGroupGlobalV2		= "global"
	IdentityProposalGroupNameHashUnknownV2	= "unknown"
	IdentityProposalGroupingOrderNameHashV2	= "name_hash"
)

type IdentityLookupObservabilityEventV2 struct {
	Type		string
	NameHash	string
	QueryType	string
	Height		uint64
	Count		uint64
	ConsensusFree	bool
}

type IdentityReadOnlyQueryAuditV2 struct {
	BeforeRoot		string
	AfterRoot		string
	MutationDetected	bool
	ConsensusWrites		[]string
	Observability		[]IdentityLookupObservabilityEventV2
}

type IdentityABCIPlusPrecheckV2 struct {
	Accepted	bool
	MessageName	string
	Error		string
	Event		*IdentityABCIEventV2
}

type IdentityProposalGroupV2 struct {
	GroupKey	string
	Indexes		[]uint32
	Names		[]string
	NameHashes	[]string
}

type IdentityProposalGroupingV2 struct {
	Order	string
	Groups	[]IdentityProposalGroupV2
	Events	[]IdentityABCIEventV2
}

type IdentityABCIEventV2 struct {
	Type		string
	Height		uint64
	Name		string
	NameHash	string
	Message		string
	Attributes	[]string
}

type IdentityFinalizeRequestV2 struct {
	State		IdentityState
	Height		uint64
	ExpiryLimit	uint32
	CacheRecords	[]ResolutionCacheRecordV2
}

type IdentityFinalizeResponseV2 struct {
	State			IdentityState
	ExpiredDomains		[]Domain
	InvalidatedCaches	[]ResolutionCacheRecordV2
	Events			[]IdentityABCIEventV2
}

type IdentityVoteExtensionTelemetryV2 struct {
	Enabled		bool
	Height		uint64
	LookupCount	uint64
	Event		*IdentityABCIEventV2
}

func AuditIdentityReadOnlyQueryV2(state IdentityState, height uint64, queryType string, name string, fn func(IdentityQueryServiceV2) IdentityQueryResponseV2) (IdentityQueryResponseV2, IdentityReadOnlyQueryAuditV2, error) {
	before := state.Export()
	beforeRoot, err := IdentityStateRoot(before)
	if err != nil {
		return IdentityQueryResponseV2{}, IdentityReadOnlyQueryAuditV2{}, err
	}
	service := NewIdentityQueryServiceV2(IdentityQueryContextV2{State: before, Height: height, DefaultTTL: 30})
	response := fn(service)
	afterRoot, err := IdentityStateRoot(before)
	if err != nil {
		return IdentityQueryResponseV2{}, IdentityReadOnlyQueryAuditV2{}, err
	}
	event, err := NewIdentityLookupObservabilityEventV2(name, queryType, height, 1)
	if err != nil {
		return IdentityQueryResponseV2{}, IdentityReadOnlyQueryAuditV2{}, err
	}
	audit := IdentityReadOnlyQueryAuditV2{
		BeforeRoot:		beforeRoot,
		AfterRoot:		afterRoot,
		MutationDetected:	beforeRoot != afterRoot,
		ConsensusWrites:	nil,
		Observability:		[]IdentityLookupObservabilityEventV2{event},
	}
	if audit.MutationDetected {
		return response, audit, errors.New("identity read-only query mutated state")
	}
	return response, audit, nil
}

func NewIdentityLookupObservabilityEventV2(name string, queryType string, height uint64, count uint64) (IdentityLookupObservabilityEventV2, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return IdentityLookupObservabilityEventV2{}, err
	}
	if queryType == "" {
		return IdentityLookupObservabilityEventV2{}, errors.New("identity lookup observability query_type is required")
	}
	if height == 0 {
		return IdentityLookupObservabilityEventV2{}, errors.New("identity lookup observability height is required")
	}
	if count == 0 {
		return IdentityLookupObservabilityEventV2{}, errors.New("identity lookup observability count is required")
	}
	return IdentityLookupObservabilityEventV2{
		Type:		IdentityObservabilityLookupVolumeV2,
		NameHash:	nameHash,
		QueryType:	queryType,
		Height:		height,
		Count:		count,
		ConsensusFree:	true,
	}, nil
}

func PrecheckIdentityABCIPlusTxV2(msg IdentityMsgV2, height uint64) IdentityABCIPlusPrecheckV2 {
	name := ""
	if msg != nil {
		name = msg.IdentityMessageName()
	}
	if height == 0 {
		return IdentityABCIPlusPrecheckV2{Accepted: false, MessageName: name, Error: "identity ABCI++ precheck height is required", Event: &IdentityABCIEventV2{Type: IdentityABCIEventMalformedRejectedV2, Message: name}}
	}
	if err := ValidateIdentityMsgV2(msg); err != nil {
		return IdentityABCIPlusPrecheckV2{
			Accepted:	false,
			MessageName:	name,
			Error:		err.Error(),
			Event:		&IdentityABCIEventV2{Type: IdentityABCIEventMalformedRejectedV2, Height: height, Message: name, Attributes: []string{err.Error()}},
		}
	}
	return IdentityABCIPlusPrecheckV2{Accepted: true, MessageName: name}
}

func GroupIdentityProposalUpdatesV2(msgs []IdentityMsgV2, height uint64) (IdentityProposalGroupingV2, error) {
	if height == 0 {
		return IdentityProposalGroupingV2{}, errors.New("identity proposal grouping height is required")
	}
	groups := map[string]*IdentityProposalGroupV2{}
	for i, msg := range msgs {
		precheck := PrecheckIdentityABCIPlusTxV2(msg, height)
		if !precheck.Accepted {
			return IdentityProposalGroupingV2{}, errors.New(precheck.Error)
		}
		hashes, names := identityMsgNameHashesV2(msg)
		key := IdentityProposalGroupGlobalV2
		if len(hashes) > 0 {
			key = hashes[0]
		}
		group, found := groups[key]
		if !found {
			group = &IdentityProposalGroupV2{GroupKey: key}
			groups[key] = group
		}
		group.Indexes = append(group.Indexes, uint32(i))
		group.Names = append(group.Names, names...)
		group.NameHashes = append(group.NameHashes, hashes...)
	}
	out := IdentityProposalGroupingV2{Order: IdentityProposalGroupingOrderNameHashV2}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		group := *groups[key]
		group.NameHashes = sortedUniqueStrings(group.NameHashes)
		group.Names = sortedUniqueStrings(group.Names)
		out.Groups = append(out.Groups, group)
		out.Events = append(out.Events, IdentityABCIEventV2{Type: IdentityABCIEventProposalGroupV2, Height: height, NameHash: group.GroupKey, Attributes: []string{fmt.Sprintf("txs=%d", len(group.Indexes))}})
	}
	return out, nil
}

func FinalizeIdentityABCIPlusV2(request IdentityFinalizeRequestV2) (IdentityFinalizeResponseV2, error) {
	if request.Height == 0 {
		return IdentityFinalizeResponseV2{}, errors.New("identity finalize height is required")
	}
	if request.ExpiryLimit == 0 {
		return IdentityFinalizeResponseV2{}, errors.New("identity finalize expiry limit is required")
	}
	state := request.State.Export()
	if err := state.Validate(); err != nil {
		return IdentityFinalizeResponseV2{}, err
	}
	expired := make([]Domain, 0)
	for _, domain := range state.Domains {
		if domain.ExpiryHeight <= request.Height {
			expired = append(expired, cloneDomain(domain))
		}
	}
	sort.SliceStable(expired, func(i, j int) bool {
		if expired[i].ExpiryHeight != expired[j].ExpiryHeight {
			return expired[i].ExpiryHeight < expired[j].ExpiryHeight
		}
		return expired[i].Name < expired[j].Name
	})
	if len(expired) > int(request.ExpiryLimit) {
		expired = expired[:request.ExpiryLimit]
	}
	response := IdentityFinalizeResponseV2{State: state, ExpiredDomains: expired}
	invalidated := append([]ResolutionCacheRecordV2(nil), request.CacheRecords...)
	for _, domain := range expired {
		nameHash, err := DomainRecordV2NameHash(domain.Name)
		if err != nil {
			return IdentityFinalizeResponseV2{}, err
		}
		event := IdentityCacheInvalidationEventV2{
			Trigger:	IdentityCacheInvalidDomainExpiryV2,
			NameHash:	nameHash,
			RecordVersion:	domain.UpdatedHeight + 1,
			Height:		request.Height,
			ParentEpoch:	domain.UpdatedHeight + 1,
			ChildEpoch:	domain.UpdatedHeight + 1,
		}
		next, err := InvalidateIdentityResolutionCachesV2(invalidated, event)
		if err != nil {
			return IdentityFinalizeResponseV2{}, err
		}
		invalidated = next
		response.Events = append(response.Events,
			IdentityABCIEventV2{Type: IdentityABCIEventDomainExpiredV2, Height: request.Height, Name: domain.Name, NameHash: nameHash},
			IdentityABCIEventV2{Type: IdentityABCIEventCacheInvalidatedV2, Height: request.Height, Name: domain.Name, NameHash: nameHash, Attributes: []string{string(IdentityCacheInvalidDomainExpiryV2)}},
		)
	}
	response.InvalidatedCaches = invalidated
	return response, nil
}

func BuildIdentityVoteExtensionTelemetryV2(enabled bool, height uint64, lookupCount uint64) (IdentityVoteExtensionTelemetryV2, error) {
	if !enabled {
		return IdentityVoteExtensionTelemetryV2{Enabled: false, Height: height}, nil
	}
	if height == 0 {
		return IdentityVoteExtensionTelemetryV2{}, errors.New("identity vote extension telemetry height is required")
	}
	event := IdentityABCIEventV2{Type: IdentityABCIEventVoteTelemetryV2, Height: height, Attributes: []string{fmt.Sprintf("lookup_count=%d", lookupCount)}}
	return IdentityVoteExtensionTelemetryV2{Enabled: true, Height: height, LookupCount: lookupCount, Event: &event}, nil
}

func identityMsgNameHashesV2(msg IdentityMsgV2) ([]string, []string) {
	switch m := msg.(type) {
	case MsgCommitRegistrationV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgRevealRegistrationV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgRegisterDirectV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgRenewDomainV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgTransferDomainV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgSetResolverV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgUpdateResolverRecordV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgCreateSubdomainV2:
		return normalizeIdentityMsgNameHashPairV2(m.ParentName, m.ParentNameHash)
	case MsgRevokeDelegationV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgExpireDomainV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgBatchUpdateResolversV2:
		hashes := make([]string, 0, len(m.Updates))
		names := make([]string, 0, len(m.Updates))
		for _, update := range m.Updates {
			h, n := normalizeIdentityMsgNameHashPairV2(update.Name, update.NameHash)
			hashes = append(hashes, h...)
			names = append(names, n...)
		}
		return sortedUniqueStrings(hashes), sortedUniqueStrings(names)
	case MsgBatchRenewDomainsV2:
		hashes := make([]string, 0, len(m.Renewals))
		names := make([]string, 0, len(m.Renewals))
		for _, renewal := range m.Renewals {
			h, n := normalizeIdentityMsgNameHashPairV2(renewal.Name, renewal.NameHash)
			hashes = append(hashes, h...)
			names = append(names, n...)
		}
		return sortedUniqueStrings(hashes), sortedUniqueStrings(names)
	case MsgDelegateSubdomainV2:
		return []string{m.Delegation.NameHash}, nil
	case MsgSetReverseRecordV2:
		return []string{m.Record.NameHash}, []string{m.Record.Name}
	case MsgVerifyReverseRecordV2:
		return []string{m.Record.NameHash}, []string{m.Record.Name}
	case MsgStartAuctionV2:
		return normalizeIdentityMsgNameHashPairV2(m.Name, m.NameHash)
	case MsgCommitBidV2:
		return []string{m.NameHash}, nil
	case MsgRevealBidV2:
		return []string{m.NameHash}, nil
	case MsgFinalizeAuctionV2:
		return []string{m.NameHash}, nil
	case MsgInvalidateResolutionCacheV2:
		return []string{m.NameHash}, nil
	default:
		return []string{IdentityProposalGroupNameHashUnknownV2}, nil
	}
}

func normalizeIdentityMsgNameHashPairV2(name string, nameHash string) ([]string, []string) {
	hash, err := validateIdentityTxNameOrHashV2(name, nameHash)
	if err != nil {
		return []string{IdentityProposalGroupNameHashUnknownV2}, nil
	}
	names := []string(nil)
	if name != "" {
		if normalized, err := NormalizeAETDomain(name); err == nil {
			names = []string{normalized}
		}
	}
	return []string{hash}, names
}
