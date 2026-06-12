package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type NamedExecutionKind string

const (
	NamedExecutionSend	NamedExecutionKind	= "send"
	NamedExecutionInvoke	NamedExecutionKind	= "invoke"
)

type ResolverAddressRecord struct {
	Key	string
	Address	sdk.AccAddress
}

type IdentityExecutionRoute struct {
	ZoneID		string
	ShardID		string
	VM		string
	Entrypoint	string
}

type UnifiedResolverView struct {
	QueryDomain	string
	ResolverDomain	string
	AuthorityDomain	string
	AuthorityOwner	sdk.AccAddress
	Primary		sdk.AccAddress
	Contract	sdk.AccAddress
	ZoneEndpoint	string
	Records		[]ResolverAddressRecord
	Metadata	[]ResolverMetadataEntry
	Route		IdentityExecutionRoute
}

type NamedExecutionRequest struct {
	Kind		NamedExecutionKind
	Name		string
	RecordKey	string
	InterfaceID	string
	Method		string
	PayloadHash	string
}

type NamedExecutionTarget struct {
	Kind		NamedExecutionKind
	QueryDomain	string
	ResolverDomain	string
	AuthorityDomain	string
	Address		sdk.AccAddress
	Contract	sdk.AccAddress
	RecordKey	string
	InterfaceID	string
	Method		string
	PayloadHash	string
	Route		IdentityExecutionRoute
}

func BuildUnifiedResolverView(state IdentityState, name string, height uint64) (UnifiedResolverView, error) {
	resolution, err := ResolveIdentityRecordRecursive(state, name, height)
	if err != nil {
		return UnifiedResolverView{}, err
	}
	metadata, err := DecodeResolverMetadata(resolution.Record.Metadata)
	if err != nil {
		return UnifiedResolverView{}, err
	}
	view := UnifiedResolverView{
		QueryDomain:		resolution.QueryDomain,
		ResolverDomain:		resolution.ResolverDomain,
		AuthorityDomain:	resolution.AuthorityDomain.Name,
		AuthorityOwner:		cloneSpecAddress(resolution.AuthorityDomain.Owner),
		Primary:		cloneSpecAddress(resolution.Record.Primary),
		Contract:		cloneSpecAddress(resolution.Record.Contract),
		ZoneEndpoint:		resolution.Record.ZoneEndpoint,
		Metadata:		append([]ResolverMetadataEntry(nil), metadata...),
	}
	for _, key := range sortedResolverKeys(resolution.Record.Records) {
		view.Records = append(view.Records, ResolverAddressRecord{Key: key, Address: cloneSpecAddress(resolution.Record.Records[key])})
	}
	view.Route = routeFromMetadata(metadata)
	return view, nil
}

func ResolveNamedExecutionTarget(state IdentityState, request NamedExecutionRequest, height uint64) (NamedExecutionTarget, error) {
	if err := ValidateNamedExecutionRequest(request); err != nil {
		return NamedExecutionTarget{}, err
	}
	view, err := BuildUnifiedResolverView(state, request.Name, height)
	if err != nil {
		return NamedExecutionTarget{}, err
	}
	target := NamedExecutionTarget{
		Kind:			request.Kind,
		QueryDomain:		view.QueryDomain,
		ResolverDomain:		view.ResolverDomain,
		AuthorityDomain:	view.AuthorityDomain,
		RecordKey:		request.RecordKey,
		InterfaceID:		request.InterfaceID,
		Method:			request.Method,
		PayloadHash:		request.PayloadHash,
		Route:			view.Route,
	}
	switch request.Kind {
	case NamedExecutionSend:
		address, err := addressForResolverKey(view, request.RecordKey)
		if err != nil {
			return NamedExecutionTarget{}, err
		}
		target.Address = address
	case NamedExecutionInvoke:
		if len(view.Contract) == 0 {
			return NamedExecutionTarget{}, errors.New("identity invoke target contract is not resolved")
		}
		if request.InterfaceID != "" {
			interfaceKey, err := ResolverMetadataInterfaceKey(request.InterfaceID)
			if err != nil {
				return NamedExecutionTarget{}, err
			}
			if _, found, err := resolverMetadataEntryValue(view.Metadata, interfaceKey); err != nil {
				return NamedExecutionTarget{}, err
			} else if !found {
				return NamedExecutionTarget{}, fmt.Errorf("identity resolver interface %q is not advertised", request.InterfaceID)
			}
		}
		target.Address = cloneSpecAddress(view.Contract)
		target.Contract = cloneSpecAddress(view.Contract)
		if target.Route.Entrypoint == "" {
			target.Route.Entrypoint = request.Method
		}
	default:
		return NamedExecutionTarget{}, fmt.Errorf("unsupported named execution kind %q", request.Kind)
	}
	if len(target.Address) == 0 {
		return NamedExecutionTarget{}, errors.New("identity named execution target is not resolved")
	}
	return target, nil
}

func ValidateNamedExecutionRequest(request NamedExecutionRequest) error {
	if _, err := NormalizeAETDomain(request.Name); err != nil {
		return err
	}
	switch request.Kind {
	case NamedExecutionSend:
	case NamedExecutionInvoke:
		if request.Method == "" {
			return errors.New("identity invoke method is required")
		}
	default:
		return fmt.Errorf("unsupported named execution kind %q", request.Kind)
	}
	if request.RecordKey != "" {
		if err := ValidateResolverGrantKey(request.RecordKey); err != nil {
			return err
		}
	}
	if request.InterfaceID != "" {
		if _, err := ResolverMetadataInterfaceKey(request.InterfaceID); err != nil {
			return err
		}
	}
	if request.Method != "" {
		if err := ValidateResolverMetadataKey("method." + request.Method); err != nil {
			return err
		}
	}
	if request.PayloadHash != "" {
		if err := validateHexHash("identity execution payload hash", request.PayloadHash); err != nil {
			return err
		}
	}
	return nil
}

func addressForResolverKey(view UnifiedResolverView, key string) (sdk.AccAddress, error) {
	if key == "" {
		key = ResolverKeyPrimary
	}
	switch key {
	case ResolverKeyPrimary:
		return cloneSpecAddress(view.Primary), nil
	case ResolverKeyContract:
		return cloneSpecAddress(view.Contract), nil
	default:
		for _, record := range view.Records {
			if record.Key == key {
				return cloneSpecAddress(record.Address), nil
			}
		}
		return nil, fmt.Errorf("identity resolver key %q is not resolved", key)
	}
}

func routeFromMetadata(entries []ResolverMetadataEntry) IdentityExecutionRoute {
	values := make(map[string]string, len(entries))
	for _, entry := range entries {
		values[entry.Key] = entry.Value
	}
	return IdentityExecutionRoute{
		ZoneID:		values[ResolverMetadataRouteZone],
		ShardID:	values[ResolverMetadataRouteShard],
		VM:		values[ResolverMetadataRouteVM],
		Entrypoint:	values[ResolverMetadataRouteEntrypoint],
	}
}

func resolverMetadataEntryValue(entries []ResolverMetadataEntry, key string) (string, bool, error) {
	if err := ValidateResolverMetadataKey(key); err != nil {
		return "", false, err
	}
	for _, entry := range entries {
		if entry.Key == key {
			return entry.Value, true, nil
		}
	}
	return "", false, nil
}

func SortResolverAddressRecords(records []ResolverAddressRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].Key < records[j].Key })
}
