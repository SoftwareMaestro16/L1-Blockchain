package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MaxUnifiedContractTargets      = 16
	MaxUnifiedServiceEndpoints     = 16
	MaxUnifiedInterfaceDescriptors = 16
	MaxUnifiedExecutionHints       = 16
	MaxUnifiedRecordKeyBytes       = 48
	MaxUnifiedRecordValueBytes     = 128
	MaxUnifiedEndpointBytes        = 128
	MaxUnifiedRoutingMetadataBytes = 256
	MaxUnifiedOwnerSignatureBytes  = 128
	MaxContractCodeIDBytesV2       = 64

	InterfaceDescriptorHashPrefixV2 = "sha256:"
)

type ContractTargetV2 struct {
	Key     string
	Address sdk.AccAddress
	CodeID  string
}

type ServiceEndpointV2 struct {
	Key      string
	Endpoint string
}

type InterfaceDescriptorV2 struct {
	InterfaceID string
	Descriptor  string
}

type RoutingMetadataV2 struct {
	ZoneID     string
	ShardID    string
	VM         string
	Entrypoint string
}

type ExecutionHintV2 struct {
	Key   string
	Value string
}

type UnifiedResolutionRecordV2 struct {
	NameHash               string
	PrimaryAddress         sdk.AccAddress
	ContractTargets        []ContractTargetV2
	ServiceEndpoints       []ServiceEndpointV2
	InterfaceDescriptors   []InterfaceDescriptorV2
	RoutingMetadata        RoutingMetadataV2
	ExecutionHints         []ExecutionHintV2
	RecordVersion          uint64
	RecordTTL              uint64
	UpdatedAtHeight        uint64
	OwnerSignatureOptional []byte
}

type ReverseResolutionRecordV2 struct {
	Address         sdk.AccAddress
	NameHash        string
	Name            string
	Verified        bool
	UpdatedAtHeight uint64
	ExpiryHeight    uint64
}

func BuildUnifiedResolutionRecordV2(state IdentityState, name string, height uint64, ttl uint64) (UnifiedResolutionRecordV2, error) {
	view, err := BuildUnifiedResolverView(state, name, height)
	if err != nil {
		return UnifiedResolutionRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(view.QueryDomain)
	if err != nil {
		return UnifiedResolutionRecordV2{}, err
	}
	record := UnifiedResolutionRecordV2{
		NameHash:        nameHash,
		PrimaryAddress:  cloneSpecAddress(view.Primary),
		RoutingMetadata: routeV2FromExecutionRoute(view.Route),
		RecordVersion:   1,
		RecordTTL:       ttl,
		UpdatedAtHeight: height,
	}
	if len(view.Contract) > 0 {
		record.ContractTargets = append(record.ContractTargets, ContractTargetV2{Key: ResolverKeyContract, Address: cloneSpecAddress(view.Contract)})
	}
	for _, addressRecord := range view.Records {
		record.ContractTargets = append(record.ContractTargets, ContractTargetV2{Key: addressRecord.Key, Address: cloneSpecAddress(addressRecord.Address)})
	}
	for _, entry := range view.Metadata {
		switch {
		case strings.HasPrefix(entry.Key, ResolverMetadataServicePrefix):
			record.ServiceEndpoints = append(record.ServiceEndpoints, ServiceEndpointV2{
				Key:      strings.TrimPrefix(entry.Key, ResolverMetadataServicePrefix),
				Endpoint: entry.Value,
			})
		case strings.HasPrefix(entry.Key, ResolverMetadataInterfacePrefix):
			descriptorHash, err := InterfaceDescriptorHashV2(entry.Value)
			if err != nil {
				return UnifiedResolutionRecordV2{}, err
			}
			record.InterfaceDescriptors = append(record.InterfaceDescriptors, InterfaceDescriptorV2{
				InterfaceID: strings.TrimPrefix(entry.Key, ResolverMetadataInterfacePrefix),
				Descriptor:  descriptorHash,
			})
		case isResolverRouteMetadataKey(entry.Key):
			continue
		default:
			record.ExecutionHints = append(record.ExecutionHints, ExecutionHintV2{Key: entry.Key, Value: entry.Value})
		}
	}
	sortUnifiedResolutionRecordV2(&record)
	return record, ValidateUnifiedResolutionRecordV2(record)
}

func ValidateUnifiedResolutionRecordV2(record UnifiedResolutionRecordV2) error {
	if err := validateHexHash("identity v2 unified resolution name hash", record.NameHash); err != nil {
		return err
	}
	if len(record.PrimaryAddress) > 0 {
		if err := validateSpecAddress("identity v2 unified primary address", record.PrimaryAddress); err != nil {
			return err
		}
	}
	if len(record.ContractTargets) > MaxUnifiedContractTargets {
		return fmt.Errorf("identity v2 contract targets must not exceed %d", MaxUnifiedContractTargets)
	}
	if len(record.ServiceEndpoints) > MaxUnifiedServiceEndpoints {
		return fmt.Errorf("identity v2 service endpoints must not exceed %d", MaxUnifiedServiceEndpoints)
	}
	if len(record.InterfaceDescriptors) > MaxUnifiedInterfaceDescriptors {
		return fmt.Errorf("identity v2 interface descriptors must not exceed %d", MaxUnifiedInterfaceDescriptors)
	}
	if len(record.ExecutionHints) > MaxUnifiedExecutionHints {
		return fmt.Errorf("identity v2 execution hints must not exceed %d", MaxUnifiedExecutionHints)
	}
	if err := validateContractTargetsV2(record.ContractTargets); err != nil {
		return err
	}
	if err := validateServiceEndpointsV2(record.ServiceEndpoints); err != nil {
		return err
	}
	if err := validateInterfaceDescriptorsV2(record.InterfaceDescriptors); err != nil {
		return err
	}
	if err := validateRoutingMetadataV2(record.RoutingMetadata); err != nil {
		return err
	}
	if err := validateExecutionHintsV2(record.ExecutionHints); err != nil {
		return err
	}
	if record.RecordVersion == 0 {
		return errors.New("identity v2 unified record version is required")
	}
	if record.RecordTTL == 0 {
		return errors.New("identity v2 unified record ttl is required")
	}
	if record.UpdatedAtHeight == 0 {
		return errors.New("identity v2 unified updated_at_height is required")
	}
	if len(record.OwnerSignatureOptional) > MaxUnifiedOwnerSignatureBytes {
		return fmt.Errorf("identity v2 owner signature must not exceed %d bytes", MaxUnifiedOwnerSignatureBytes)
	}
	return nil
}

func ValidateResolverRecordVersionForUpdateV2(currentVersion uint64, expectedVersion uint64) error {
	if expectedVersion == 0 {
		return errors.New("identity v2 expected record version is required")
	}
	if currentVersion == 0 {
		return errors.New("identity v2 current record version is required")
	}
	if currentVersion != expectedVersion {
		return fmt.Errorf("identity v2 resolver record version conflict: expected %d got %d", expectedVersion, currentVersion)
	}
	return nil
}

func NewReverseResolutionRecordV2(address sdk.AccAddress, name string, verified bool, updatedAtHeight uint64, expiryHeight uint64) (ReverseResolutionRecordV2, error) {
	if err := validateSpecAddress("identity v2 reverse address", address); err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	record := ReverseResolutionRecordV2{
		Address:         cloneSpecAddress(address),
		NameHash:        nameHash,
		Name:            normalized,
		Verified:        verified,
		UpdatedAtHeight: updatedAtHeight,
		ExpiryHeight:    expiryHeight,
	}
	return record, ValidateReverseResolutionRecordV2Format(record)
}

func ValidateReverseResolutionRecordV2Format(record ReverseResolutionRecordV2) error {
	if err := validateSpecAddress("identity v2 reverse address", record.Address); err != nil {
		return err
	}
	normalized, err := NormalizeAETDomain(record.Name)
	if err != nil {
		return err
	}
	if record.Name != normalized {
		return errors.New("identity v2 reverse name must be normalized")
	}
	expectedNameHash, err := DomainRecordV2NameHash(record.Name)
	if err != nil {
		return err
	}
	if record.NameHash != expectedNameHash {
		return errors.New("identity v2 reverse name_hash mismatch")
	}
	if record.UpdatedAtHeight == 0 {
		return errors.New("identity v2 reverse updated_at_height is required")
	}
	if record.ExpiryHeight <= record.UpdatedAtHeight {
		return errors.New("identity v2 reverse expiry_height must be after updated_at_height")
	}
	return nil
}

func ValidateReverseResolutionRecordV2(state IdentityState, record ReverseResolutionRecordV2, height uint64, authorizedAliasKeys []string) error {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return err
	}
	if record.ExpiryHeight <= height {
		return errors.New("identity v2 reverse record is expired")
	}
	domain, err := requireActiveDomain(state, record.Name, height)
	if err != nil {
		return fmt.Errorf("identity v2 reverse record requires existing active domain: %w", err)
	}
	if record.ExpiryHeight > domain.ExpiryHeight {
		return errors.New("identity v2 reverse record expires after domain")
	}
	if !record.Verified {
		return nil
	}
	resolution, err := ResolveIdentityRecordRecursive(state, record.Name, height)
	if err != nil {
		return err
	}
	if forwardResolutionContainsAddress(resolution.Record, record.Address, authorizedAliasKeys) {
		return nil
	}
	return errors.New("identity v2 verified reverse record requires forward primary or authorized alias")
}

func VerifyReverseResolutionTransactionV2(state IdentityState, msg MsgVerifyReverseRecordV2, height uint64, currentVersion uint64) (ReverseResolutionRecordV2, error) {
	if err := msg.ValidateBasic(); err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	if err := ValidateResolverRecordVersionForUpdateV2(currentVersion, msg.ExpectedRecordVersion); err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	record := msg.Record
	record.Verified = true
	if err := ValidateReverseResolutionRecordV2(state, record, height, msg.AuthorizedAliasKeys); err != nil {
		return ReverseResolutionRecordV2{}, err
	}
	return record, nil
}

func InvalidateReverseRecordsForDomainV2(state IdentityState, name string, height uint64, authorizedAliasKeys []string) (IdentityState, []ReverseRecord, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityState{}, nil, err
	}
	next := state.Clone()
	kept := make([]ReverseRecord, 0, len(state.ReverseRecords))
	invalidated := make([]ReverseRecord, 0)
	for _, reverse := range state.ReverseRecords {
		if reverse.Domain != normalized {
			kept = append(kept, reverse)
			continue
		}
		v2, err := reverseRecordV2FromLegacy(state, reverse, true)
		if err == nil {
			err = ValidateReverseResolutionRecordV2(state, v2, height, authorizedAliasKeys)
		}
		if err != nil {
			invalidated = append(invalidated, reverse)
			continue
		}
		kept = append(kept, reverse)
	}
	next.ReverseRecords = kept
	sortIdentityState(&next)
	if err := next.Validate(); err != nil {
		return IdentityState{}, nil, err
	}
	return next, invalidated, nil
}

func CanonicalReverseResolutionName(record ReverseResolutionRecordV2) (string, error) {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return "", err
	}
	if !record.Verified {
		return "", errors.New("identity v2 unverified reverse record is not canonical")
	}
	return record.Name, nil
}

func forwardResolutionContainsAddress(record ResolverRecord, address sdk.AccAddress, authorizedAliasKeys []string) bool {
	if addressesEqual(record.Primary, address) {
		return true
	}
	allowed := stringSet(authorizedAliasKeys)
	for _, key := range sortedResolverKeys(record.Records) {
		if _, found := allowed[key]; !found {
			continue
		}
		if addressesEqual(record.Records[key], address) {
			return true
		}
	}
	return false
}

func sortUnifiedResolutionRecordV2(record *UnifiedResolutionRecordV2) {
	sort.SliceStable(record.ContractTargets, func(i, j int) bool { return record.ContractTargets[i].Key < record.ContractTargets[j].Key })
	sort.SliceStable(record.ServiceEndpoints, func(i, j int) bool { return record.ServiceEndpoints[i].Key < record.ServiceEndpoints[j].Key })
	sort.SliceStable(record.InterfaceDescriptors, func(i, j int) bool {
		return record.InterfaceDescriptors[i].InterfaceID < record.InterfaceDescriptors[j].InterfaceID
	})
	sort.SliceStable(record.ExecutionHints, func(i, j int) bool { return record.ExecutionHints[i].Key < record.ExecutionHints[j].Key })
}

func routeV2FromExecutionRoute(route IdentityExecutionRoute) RoutingMetadataV2 {
	return RoutingMetadataV2{
		ZoneID:     route.ZoneID,
		ShardID:    route.ShardID,
		VM:         route.VM,
		Entrypoint: route.Entrypoint,
	}
}

func isResolverRouteMetadataKey(key string) bool {
	switch key {
	case ResolverMetadataRouteZone, ResolverMetadataRouteShard, ResolverMetadataRouteVM, ResolverMetadataRouteEntrypoint:
		return true
	default:
		return false
	}
}

func validateContractTargetsV2(targets []ContractTargetV2) error {
	seen := map[string]struct{}{}
	for i, target := range targets {
		if err := validateUnifiedRecordKey("identity v2 contract target key", target.Key); err != nil {
			return err
		}
		if len(target.Address) == 0 && target.CodeID == "" {
			return errors.New("identity v2 contract target requires address or code_id")
		}
		if len(target.Address) > 0 && target.CodeID != "" {
			return errors.New("identity v2 contract target must use address or code_id, not both")
		}
		if len(target.Address) > 0 {
			if err := validateSpecAddress("identity v2 contract target", target.Address); err != nil {
				return err
			}
		}
		if target.CodeID != "" {
			if err := validateContractCodeIDV2(target.CodeID); err != nil {
				return err
			}
		}
		if _, found := seen[target.Key]; found {
			return fmt.Errorf("duplicate identity v2 contract target %q", target.Key)
		}
		seen[target.Key] = struct{}{}
		if i > 0 && targets[i-1].Key >= target.Key {
			return errors.New("identity v2 contract targets must be sorted canonically")
		}
	}
	return nil
}

func validateServiceEndpointsV2(endpoints []ServiceEndpointV2) error {
	seen := map[string]struct{}{}
	for i, endpoint := range endpoints {
		if err := validateUnifiedRecordKey("identity v2 service endpoint key", endpoint.Key); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 service endpoint", endpoint.Endpoint, MaxUnifiedEndpointBytes); err != nil {
			return err
		}
		if err := validateServiceEndpointSchemeV2(endpoint.Endpoint); err != nil {
			return err
		}
		if _, found := seen[endpoint.Key]; found {
			return fmt.Errorf("duplicate identity v2 service endpoint %q", endpoint.Key)
		}
		seen[endpoint.Key] = struct{}{}
		if i > 0 && endpoints[i-1].Key >= endpoint.Key {
			return errors.New("identity v2 service endpoints must be sorted canonically")
		}
	}
	return nil
}

func validateInterfaceDescriptorsV2(descriptors []InterfaceDescriptorV2) error {
	seen := map[string]struct{}{}
	for i, descriptor := range descriptors {
		if err := validateUnifiedRecordKey("identity v2 interface id", descriptor.InterfaceID); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 interface descriptor", descriptor.Descriptor, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
		if err := ValidateInterfaceDescriptorHashFormatV2(descriptor.Descriptor); err != nil {
			return err
		}
		if _, found := seen[descriptor.InterfaceID]; found {
			return fmt.Errorf("duplicate identity v2 interface descriptor %q", descriptor.InterfaceID)
		}
		seen[descriptor.InterfaceID] = struct{}{}
		if i > 0 && descriptors[i-1].InterfaceID >= descriptor.InterfaceID {
			return errors.New("identity v2 interface descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateRoutingMetadataV2(route RoutingMetadataV2) error {
	totalBytes := len(route.ZoneID) + len(route.ShardID) + len(route.VM) + len(route.Entrypoint)
	if totalBytes > MaxUnifiedRoutingMetadataBytes {
		return fmt.Errorf("identity v2 routing metadata must not exceed %d bytes", MaxUnifiedRoutingMetadataBytes)
	}
	for field, value := range map[string]string{
		"zone_id":    route.ZoneID,
		"shard_id":   route.ShardID,
		"vm":         route.VM,
		"entrypoint": route.Entrypoint,
	} {
		if value == "" {
			continue
		}
		if err := validateUnifiedRecordValue("identity v2 routing "+field, value, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
	}
	return nil
}

func validateExecutionHintsV2(hints []ExecutionHintV2) error {
	seen := map[string]struct{}{}
	for i, hint := range hints {
		if err := validateUnifiedRecordKey("identity v2 execution hint key", hint.Key); err != nil {
			return err
		}
		if !strings.HasPrefix(hint.Key, "hint.") {
			return errors.New("identity v2 execution hints are advisory and must use hint.* keys")
		}
		if err := validateUnifiedRecordValue("identity v2 execution hint", hint.Value, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
		if _, found := seen[hint.Key]; found {
			return fmt.Errorf("duplicate identity v2 execution hint %q", hint.Key)
		}
		seen[hint.Key] = struct{}{}
		if i > 0 && hints[i-1].Key >= hint.Key {
			return errors.New("identity v2 execution hints must be sorted canonically")
		}
	}
	return nil
}

func InterfaceDescriptorHashV2(descriptor string) (string, error) {
	if strings.HasPrefix(descriptor, InterfaceDescriptorHashPrefixV2) {
		normalized := strings.ToLower(descriptor)
		if err := ValidateInterfaceDescriptorHashFormatV2(normalized); err != nil {
			return "", err
		}
		return normalized, nil
	}
	if err := validateUnifiedRecordValue("identity v2 interface descriptor source", descriptor, MaxUnifiedRecordValueBytes); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(descriptor))
	return InterfaceDescriptorHashPrefixV2 + hex.EncodeToString(sum[:]), nil
}

func ValidateInterfaceDescriptorHashFormatV2(hash string) error {
	if !strings.HasPrefix(hash, InterfaceDescriptorHashPrefixV2) {
		return errors.New("identity v2 interface descriptor must use sha256:<64 hex> schema hash")
	}
	raw := strings.TrimPrefix(hash, InterfaceDescriptorHashPrefixV2)
	if len(raw) != 64 {
		return errors.New("identity v2 interface descriptor hash must contain 64 hex characters")
	}
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return fmt.Errorf("identity v2 interface descriptor hash contains unsupported character %q", c)
	}
	return nil
}

func validateServiceEndpointSchemeV2(endpoint string) error {
	scheme, rest, found := strings.Cut(endpoint, "://")
	if !found || rest == "" {
		return errors.New("identity v2 service endpoint must include an allowed scheme")
	}
	switch scheme {
	case "https", "grpcs", "wss", "aetheris", "ipfs":
		return nil
	default:
		return fmt.Errorf("identity v2 service endpoint scheme %q is not allowed", scheme)
	}
}

func validateContractCodeIDV2(codeID string) error {
	if codeID == "" {
		return errors.New("identity v2 contract code_id is required")
	}
	if len(codeID) > MaxContractCodeIDBytesV2 {
		return fmt.Errorf("identity v2 contract code_id must not exceed %d bytes", MaxContractCodeIDBytesV2)
	}
	for i := 0; i < len(codeID); i++ {
		c := codeID[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == ':' {
			continue
		}
		return fmt.Errorf("identity v2 contract code_id contains unsupported character %q", c)
	}
	return nil
}

func reverseRecordV2FromLegacy(state IdentityState, reverse ReverseRecord, verified bool) (ReverseResolutionRecordV2, error) {
	domain, found := findDomain(state, reverse.Domain)
	if !found {
		return ReverseResolutionRecordV2{}, errors.New("identity v2 reverse legacy domain not found")
	}
	if reverse.UpdatedAtUnix <= 0 {
		return ReverseResolutionRecordV2{}, errors.New("identity v2 reverse legacy updated_at is required")
	}
	return NewReverseResolutionRecordV2(reverse.Address, reverse.Domain, verified, uint64(reverse.UpdatedAtUnix), domain.ExpiryHeight)
}

func validateUnifiedRecordKey(field string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > MaxUnifiedRecordKeyBytes {
		return fmt.Errorf("%s must not exceed %d bytes", field, MaxUnifiedRecordKeyBytes)
	}
	return ValidateResolverMetadataKey(value)
}

func validateUnifiedRecordValue(field string, value string, maxBytes int) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have surrounding whitespace", field)
	}
	if len(value) > maxBytes {
		return fmt.Errorf("%s must not exceed %d bytes", field, maxBytes)
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c < 0x21 || c > 0x7e {
			return fmt.Errorf("%s contains unsupported character %q", field, c)
		}
	}
	return nil
}
