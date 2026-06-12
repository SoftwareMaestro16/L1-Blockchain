package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MaxUnifiedContractTargets	= 16
	MaxUnifiedServiceEndpoints	= 16
	MaxUnifiedInterfaceDescriptors	= 16
	MaxUnifiedExecutionHints	= 16
	MaxUnifiedRecordKeyBytes	= 48
	MaxUnifiedRecordValueBytes	= 128
	MaxUnifiedEndpointBytes		= 128
	MaxUnifiedRoutingMetadataBytes	= 256
	MaxUnifiedOwnerSignatureBytes	= 128
	MaxUnifiedRouteListEntries	= 16
	MaxContractCodeIDBytesV2	= 64
	MaxContractEntrypointBytesV2	= 64
	MaxRequiredFundsPolicyBytesV2	= 64
	MaxUnifiedPayloadBytesV2	= 64 * 1024
	MaxContractGasHintV2		= 100_000_000
	MaxRouteIDBytesV2		= 48
	MaxRouteTargetTypeBytesV2	= 24
	MaxRouteTargetBytesV2		= 64
	MaxRouteChainContextBytesV2	= 64
	MaxRouteFeeHintBytesV2		= 64
	MaxRouteMemoPolicyBytesV2	= 48
	MaxRouteCapabilityBytesV2	= 48
	MaxRouteTimeoutHintV2		= 86_400
	MaxServiceTypeBytesV2		= 48
	MaxServiceTransportBytesV2	= 24
	MaxServiceAuthPolicyBytesV2	= 48
	MaxServiceHealthPathBytesV2	= 96
	MaxServicePriorityV2		= 1_000
	MaxServiceWeightV2		= 1_000_000
	MaxInterfaceSchemaURIBytesV2	= 128
	MaxInterfaceInlineSchemaBytesV2	= 2 * 1024
	MaxInterfaceVersionBytesV2	= 32
	MaxInterfaceRenderPolicyBytesV2	= 48
	MaxInterfacePermissionBytesV2	= 48
	MaxInterfacePermissionsV2	= 16
	MaxExecutionFeeModeBytesV2	= 32
	MaxExecutionMessageTypeBytesV2	= 64
	MaxExecutionGasLimitHintV2	= 100_000_000

	UnifiedResolutionSchemaVersionV2	uint64	= 1

	InterfaceDescriptorHashPrefixV2	= "sha256:"
)

type ContractTargetV2 struct {
	Key			string
	Address			sdk.AccAddress
	CodeID			string
	TargetID		string
	ContractAddress		sdk.AccAddress
	Entrypoint		string
	InterfaceHash		string
	RequiredFundsPolicy	string
	GasHint			uint64
	Enabled			bool
	UpdatedAtHeight		uint64
}

type ServiceEndpointV2 struct {
	Key			string
	Endpoint		string
	ServiceID		string
	ServiceType		string
	Transport		string
	AuthPolicy		string
	HealthPathOptional	string
	Priority		uint32
	Weight			uint32
	TTL			uint64
	SchemaHashOptional	string
}

type InterfaceDescriptorV2 struct {
	InterfaceID			string
	Descriptor			string
	SchemaHash			string
	SchemaURIOptional		string
	SchemaInlineOptional		string
	Version				string
	RenderPolicy			string
	PermissionsRequired		[]string
	ContractTargetIDOptional	string
	ServiceIDOptional		string
}

type RoutingMetadataV2 struct {
	ZoneID			string
	ShardID			string
	VM			string
	Entrypoint		string
	RouteID			string
	TargetType		string
	PreferredTarget		string
	FallbackTargets		[]string
	ChainContext		string
	FeeHint			string
	TimeoutHint		uint64
	MemoPolicy		string
	CapabilityRequirements	[]string
}

type ExecutionHintV2 struct {
	Key				string
	Value				string
	DefaultGasLimitHint		uint64
	PreferredFeeMode		string
	MessageType			string
	AsyncAllowed			bool
	RequiresMemo			bool
	RequiresInterfaceConfirmation	bool
	SimulationRequired		bool
}

type UnifiedResolutionRecordV2 struct {
	NameHash		string
	Owner			sdk.AccAddress
	PrimaryAddress		sdk.AccAddress
	ContractTargets		[]ContractTargetV2
	ServiceEndpoints	[]ServiceEndpointV2
	InterfaceDescriptors	[]InterfaceDescriptorV2
	RoutingMetadata		RoutingMetadataV2
	ExecutionHints		[]ExecutionHintV2
	RecordVersion		uint64
	RecordTTL		uint64
	UpdatedAtHeight		uint64
	MaxPayloadBytes		uint64
	SchemaVersion		uint64
	OwnerSignatureOptional	[]byte
}

type ReverseResolutionRecordV2 struct {
	Address		sdk.AccAddress
	NameHash	string
	Name		string
	Verified	bool
	UpdatedAtHeight	uint64
	ExpiryHeight	uint64
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
		NameHash:		nameHash,
		Owner:			cloneSpecAddress(view.AuthorityOwner),
		PrimaryAddress:		cloneSpecAddress(view.Primary),
		RoutingMetadata:	routeV2FromExecutionRoute(view.Route),
		RecordVersion:		1,
		RecordTTL:		ttl,
		UpdatedAtHeight:	height,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	if len(view.Contract) > 0 {
		record.ContractTargets = append(record.ContractTargets, NewContractTargetV2(ResolverKeyContract, view.Contract, height))
	}
	for _, addressRecord := range view.Records {
		record.ContractTargets = append(record.ContractTargets, NewContractTargetV2(addressRecord.Key, addressRecord.Address, height))
	}
	for _, entry := range view.Metadata {
		switch {
		case strings.HasPrefix(entry.Key, ResolverMetadataServicePrefix):
			serviceID := strings.TrimPrefix(entry.Key, ResolverMetadataServicePrefix)
			transport, err := serviceEndpointTransportV2(entry.Value)
			if err != nil {
				return UnifiedResolutionRecordV2{}, err
			}
			record.ServiceEndpoints = append(record.ServiceEndpoints, ServiceEndpointV2{
				Key:		serviceID,
				Endpoint:	entry.Value,
				ServiceID:	serviceID,
				ServiceType:	"service.v1",
				Transport:	transport,
				AuthPolicy:	"none",
				Priority:	100,
				Weight:		1,
				TTL:		ttl,
			})
		case strings.HasPrefix(entry.Key, ResolverMetadataInterfacePrefix):
			interfaceID := strings.TrimPrefix(entry.Key, ResolverMetadataInterfacePrefix)
			descriptorHash, err := InterfaceDescriptorHashV2(entry.Value)
			if err != nil {
				return UnifiedResolutionRecordV2{}, err
			}
			record.InterfaceDescriptors = append(record.InterfaceDescriptors, InterfaceDescriptorV2{
				InterfaceID:	interfaceID,
				Descriptor:	descriptorHash,
				SchemaHash:	descriptorHash,
				Version:	"v1",
				RenderPolicy:	"wallet_confirm",
			})
		case isResolverRouteMetadataKey(entry.Key):
			if err := applyRoutingMetadataEntryV2(&record.RoutingMetadata, entry); err != nil {
				return UnifiedResolutionRecordV2{}, err
			}
		default:
			hint, err := executionHintFromResolverMetadataV2(entry)
			if err != nil {
				return UnifiedResolutionRecordV2{}, err
			}
			record.ExecutionHints = append(record.ExecutionHints, hint)
		}
	}
	sortUnifiedResolutionRecordV2(&record)
	return record, ValidateUnifiedResolutionRecordV2(record)
}

func ValidateUnifiedResolutionRecordV2(record UnifiedResolutionRecordV2) error {
	if err := validateHexHash("identity v2 unified resolution name hash", record.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 unified owner", record.Owner); err != nil {
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
	if err := validateContractTargetsV2(record.ContractTargets, record.InterfaceDescriptors); err != nil {
		return err
	}
	if err := validateServiceEndpointsV2(record.ServiceEndpoints, record.RecordTTL); err != nil {
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
	if record.MaxPayloadBytes == 0 {
		return errors.New("identity v2 unified max_payload_bytes is required")
	}
	if record.MaxPayloadBytes > MaxUnifiedPayloadBytesV2 {
		return fmt.Errorf("identity v2 unified max_payload_bytes must not exceed %d", MaxUnifiedPayloadBytesV2)
	}
	payloadBytes := EstimateUnifiedResolverPayloadBytesV2(record)
	if payloadBytes > record.MaxPayloadBytes {
		return fmt.Errorf("identity v2 unified resolver payload size %d exceeds max_payload_bytes %d", payloadBytes, record.MaxPayloadBytes)
	}
	if record.SchemaVersion != UnifiedResolutionSchemaVersionV2 {
		return fmt.Errorf("unsupported identity v2 unified schema_version %d", record.SchemaVersion)
	}
	if len(record.OwnerSignatureOptional) > MaxUnifiedOwnerSignatureBytes {
		return fmt.Errorf("identity v2 owner signature must not exceed %d bytes", MaxUnifiedOwnerSignatureBytes)
	}
	return nil
}

func NewContractTargetV2(targetID string, contractAddress sdk.AccAddress, updatedAtHeight uint64) ContractTargetV2 {
	return ContractTargetV2{
		Key:			targetID,
		Address:		cloneSpecAddress(contractAddress),
		TargetID:		targetID,
		ContractAddress:	cloneSpecAddress(contractAddress),
		Enabled:		true,
		UpdatedAtHeight:	updatedAtHeight,
	}
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
		Address:		cloneSpecAddress(address),
		NameHash:		nameHash,
		Name:			normalized,
		Verified:		verified,
		UpdatedAtHeight:	updatedAtHeight,
		ExpiryHeight:		expiryHeight,
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
	sort.SliceStable(record.ContractTargets, func(i, j int) bool {
		return contractTargetIDV2(record.ContractTargets[i]) < contractTargetIDV2(record.ContractTargets[j])
	})
	sort.SliceStable(record.ServiceEndpoints, func(i, j int) bool {
		return serviceEndpointIDV2(record.ServiceEndpoints[i]) < serviceEndpointIDV2(record.ServiceEndpoints[j])
	})
	sort.SliceStable(record.InterfaceDescriptors, func(i, j int) bool {
		return record.InterfaceDescriptors[i].InterfaceID < record.InterfaceDescriptors[j].InterfaceID
	})
	sort.SliceStable(record.ExecutionHints, func(i, j int) bool { return record.ExecutionHints[i].Key < record.ExecutionHints[j].Key })
}

func routeV2FromExecutionRoute(route IdentityExecutionRoute) RoutingMetadataV2 {
	return RoutingMetadataV2{
		ZoneID:		route.ZoneID,
		ShardID:	route.ShardID,
		VM:		route.VM,
		Entrypoint:	route.Entrypoint,
	}
}

func isResolverRouteMetadataKey(key string) bool {
	switch key {
	case ResolverMetadataRouteZone, ResolverMetadataRouteShard, ResolverMetadataRouteVM, ResolverMetadataRouteEntrypoint,
		"route.id", "route.target_type", "route.preferred_target", "route.fallback_targets", "route.chain_context",
		"route.fee_hint", "route.timeout_hint", "route.memo_policy", "route.capability_requirements":
		return true
	default:
		return false
	}
}

func applyRoutingMetadataEntryV2(route *RoutingMetadataV2, entry ResolverMetadataEntry) error {
	switch entry.Key {
	case ResolverMetadataRouteZone:
		route.ZoneID = entry.Value
	case ResolverMetadataRouteShard:
		route.ShardID = entry.Value
	case ResolverMetadataRouteVM:
		route.VM = entry.Value
	case ResolverMetadataRouteEntrypoint:
		route.Entrypoint = entry.Value
	case "route.id":
		route.RouteID = entry.Value
	case "route.target_type":
		route.TargetType = entry.Value
	case "route.preferred_target":
		route.PreferredTarget = entry.Value
	case "route.fallback_targets":
		route.FallbackTargets = splitCanonicalCSVV2(entry.Value)
	case "route.chain_context":
		route.ChainContext = entry.Value
	case "route.fee_hint":
		route.FeeHint = entry.Value
	case "route.timeout_hint":
		timeout, err := strconv.ParseUint(entry.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("identity v2 routing timeout_hint must be numeric: %w", err)
		}
		route.TimeoutHint = timeout
	case "route.memo_policy":
		route.MemoPolicy = entry.Value
	case "route.capability_requirements":
		route.CapabilityRequirements = splitCanonicalCSVV2(entry.Value)
	}
	return nil
}

func validateContractTargetsV2(targets []ContractTargetV2, descriptors []InterfaceDescriptorV2) error {
	seen := map[string]struct{}{}
	descriptorHashes := map[string]struct{}{}
	for _, descriptor := range descriptors {
		if hash := interfaceDescriptorSchemaHashV2(descriptor); hash != "" {
			descriptorHashes[hash] = struct{}{}
		}
	}
	for i, target := range targets {
		targetID := contractTargetIDV2(target)
		if err := validateUnifiedRecordKey("identity v2 contract target_id", targetID); err != nil {
			return err
		}
		if target.Key != "" && target.TargetID != "" && target.Key != target.TargetID {
			return errors.New("identity v2 contract target key and target_id must match when both are set")
		}
		address := contractTargetAddressV2(target)
		if len(address) == 0 && target.CodeID == "" {
			return errors.New("identity v2 contract target requires contract_address or code_id")
		}
		if len(address) > 0 {
			if err := validateSpecAddress("identity v2 contract target contract_address", address); err != nil {
				return err
			}
		}
		if len(target.Address) > 0 && len(target.ContractAddress) > 0 && !addressesEqual(target.Address, target.ContractAddress) {
			return errors.New("identity v2 contract target address and contract_address must match when both are set")
		}
		if target.CodeID != "" {
			if err := validateContractCodeIDV2(target.CodeID); err != nil {
				return err
			}
		}
		if target.Entrypoint != "" {
			if err := validateContractEntrypointV2(target.Entrypoint); err != nil {
				return err
			}
		}
		if target.InterfaceHash != "" {
			if err := ValidateInterfaceDescriptorHashFormatV2(target.InterfaceHash); err != nil {
				return err
			}
			if len(descriptorHashes) > 0 {
				if _, found := descriptorHashes[target.InterfaceHash]; !found {
					return errors.New("identity v2 contract target interface_hash must match an interface descriptor")
				}
			}
		}
		if target.RequiredFundsPolicy != "" {
			if err := validateUnifiedRecordValue("identity v2 contract target required_funds_policy", target.RequiredFundsPolicy, MaxRequiredFundsPolicyBytesV2); err != nil {
				return err
			}
		}
		if target.GasHint > MaxContractGasHintV2 {
			return fmt.Errorf("identity v2 contract target gas_hint is advisory and must not exceed %d", MaxContractGasHintV2)
		}
		if target.UpdatedAtHeight == 0 && (target.TargetID != "" || len(target.ContractAddress) > 0 || target.Entrypoint != "" || target.InterfaceHash != "" || target.RequiredFundsPolicy != "" || target.GasHint != 0) {
			return errors.New("identity v2 contract target updated_at_height is required")
		}
		if _, found := seen[targetID]; found {
			return fmt.Errorf("duplicate identity v2 contract target %q", targetID)
		}
		seen[targetID] = struct{}{}
		if i > 0 && contractTargetIDV2(targets[i-1]) >= targetID {
			return errors.New("identity v2 contract targets must be sorted canonically")
		}
	}
	return nil
}

func contractTargetIDV2(target ContractTargetV2) string {
	if target.TargetID != "" {
		return target.TargetID
	}
	return target.Key
}

func contractTargetAddressV2(target ContractTargetV2) sdk.AccAddress {
	if len(target.ContractAddress) > 0 {
		return target.ContractAddress
	}
	return target.Address
}

func contractTargetEnabledV2(target ContractTargetV2) bool {
	if target.TargetID == "" &&
		len(target.ContractAddress) == 0 &&
		target.Entrypoint == "" &&
		target.InterfaceHash == "" &&
		target.RequiredFundsPolicy == "" &&
		target.GasHint == 0 &&
		target.UpdatedAtHeight == 0 {
		return true
	}
	return target.Enabled
}

func validateServiceEndpointsV2(endpoints []ServiceEndpointV2, recordTTL uint64) error {
	seen := map[string]struct{}{}
	for i, endpoint := range endpoints {
		serviceID := serviceEndpointIDV2(endpoint)
		if err := validateUnifiedRecordKey("identity v2 service_id", serviceID); err != nil {
			return err
		}
		if endpoint.Key != "" && endpoint.ServiceID != "" && endpoint.Key != endpoint.ServiceID {
			return errors.New("identity v2 service endpoint key and service_id must match when both are set")
		}
		if err := validateUnifiedRecordValue("identity v2 service endpoint", endpoint.Endpoint, MaxUnifiedEndpointBytes); err != nil {
			return err
		}
		if err := validateServiceEndpointSchemeV2(endpoint.Endpoint); err != nil {
			return err
		}
		if err := validateVersionedServiceTypeV2(endpoint.ServiceType); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 service transport", endpoint.Transport, MaxServiceTransportBytesV2); err != nil {
			return err
		}
		if endpoint.AuthPolicy != "" {
			if err := validateUnifiedRecordValue("identity v2 service auth_policy", endpoint.AuthPolicy, MaxServiceAuthPolicyBytesV2); err != nil {
				return err
			}
		}
		if endpoint.HealthPathOptional != "" {
			if len(endpoint.HealthPathOptional) > MaxServiceHealthPathBytesV2 {
				return fmt.Errorf("identity v2 service health_path_optional must not exceed %d bytes", MaxServiceHealthPathBytesV2)
			}
			if !strings.HasPrefix(endpoint.HealthPathOptional, "/") {
				return errors.New("identity v2 service health_path_optional must start with /")
			}
			if err := validateASCIIValueV2("identity v2 service health_path_optional", endpoint.HealthPathOptional); err != nil {
				return err
			}
		}
		if endpoint.Priority > MaxServicePriorityV2 {
			return fmt.Errorf("identity v2 service priority must not exceed %d", MaxServicePriorityV2)
		}
		if endpoint.Weight == 0 || endpoint.Weight > MaxServiceWeightV2 {
			return fmt.Errorf("identity v2 service weight must be between 1 and %d", MaxServiceWeightV2)
		}
		if endpoint.TTL == 0 {
			return errors.New("identity v2 service ttl is required")
		}
		if recordTTL > 0 && endpoint.TTL > recordTTL {
			return errors.New("identity v2 service ttl must not exceed resolver record ttl")
		}
		if endpoint.SchemaHashOptional != "" {
			if err := ValidateInterfaceDescriptorHashFormatV2(endpoint.SchemaHashOptional); err != nil {
				return err
			}
		}
		if _, found := seen[serviceID]; found {
			return fmt.Errorf("duplicate identity v2 service endpoint %q", serviceID)
		}
		seen[serviceID] = struct{}{}
		if i > 0 && serviceEndpointIDV2(endpoints[i-1]) >= serviceID {
			return errors.New("identity v2 service endpoints must be sorted canonically")
		}
	}
	return nil
}

func validateContractEntrypointV2(entrypoint string) error {
	if entrypoint == "" {
		return nil
	}
	if len(entrypoint) > MaxContractEntrypointBytesV2 {
		return fmt.Errorf("identity v2 contract target entrypoint must not exceed %d bytes", MaxContractEntrypointBytesV2)
	}
	for i := 0; i < len(entrypoint); i++ {
		c := entrypoint[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' || c == ':' {
			continue
		}
		return fmt.Errorf("identity v2 contract target entrypoint contains unsupported character %q", c)
	}
	return nil
}

func validateInterfaceDescriptorsV2(descriptors []InterfaceDescriptorV2) error {
	seen := map[string]struct{}{}
	for i, descriptor := range descriptors {
		if err := validateUnifiedRecordKey("identity v2 interface id", descriptor.InterfaceID); err != nil {
			return err
		}
		schemaHash := interfaceDescriptorSchemaHashV2(descriptor)
		if err := validateUnifiedRecordValue("identity v2 interface schema_hash", schemaHash, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
		if err := ValidateInterfaceDescriptorHashFormatV2(schemaHash); err != nil {
			return err
		}
		if descriptor.Descriptor != "" && descriptor.SchemaHash != "" && descriptor.Descriptor != descriptor.SchemaHash {
			return errors.New("identity v2 interface descriptor and schema_hash must match when both are set")
		}
		if descriptor.SchemaURIOptional != "" {
			if err := validateInterfaceSchemaURIV2(descriptor.SchemaURIOptional); err != nil {
				return err
			}
		}
		if descriptor.SchemaInlineOptional != "" {
			if len(descriptor.SchemaInlineOptional) > MaxInterfaceInlineSchemaBytesV2 {
				return fmt.Errorf("identity v2 interface inline schema must not exceed %d bytes", MaxInterfaceInlineSchemaBytesV2)
			}
			if err := validateASCIIValueV2("identity v2 interface inline schema", descriptor.SchemaInlineOptional); err != nil {
				return err
			}
			inlineHash, err := InterfaceDescriptorHashV2(descriptor.SchemaInlineOptional)
			if err != nil {
				return err
			}
			if inlineHash != schemaHash {
				return errors.New("identity v2 interface inline schema hash mismatch")
			}
		}
		if err := validateVersionedInterfaceVersionV2(descriptor.Version); err != nil {
			return err
		}
		if err := validateUnifiedRecordValue("identity v2 interface render_policy", descriptor.RenderPolicy, MaxInterfaceRenderPolicyBytesV2); err != nil {
			return err
		}
		if err := validateInterfacePermissionsRequiredV2(descriptor.PermissionsRequired); err != nil {
			return err
		}
		if descriptor.ContractTargetIDOptional != "" {
			if err := validateUnifiedRecordKey("identity v2 interface contract_target_id_optional", descriptor.ContractTargetIDOptional); err != nil {
				return err
			}
		}
		if descriptor.ServiceIDOptional != "" {
			if err := validateUnifiedRecordKey("identity v2 interface service_id_optional", descriptor.ServiceIDOptional); err != nil {
				return err
			}
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

func serviceEndpointIDV2(endpoint ServiceEndpointV2) string {
	if endpoint.ServiceID != "" {
		return endpoint.ServiceID
	}
	return endpoint.Key
}

func serviceEndpointTransportV2(endpoint string) (string, error) {
	scheme, _, found := strings.Cut(endpoint, "://")
	if !found || scheme == "" {
		return "", errors.New("identity v2 service endpoint must include an allowed scheme")
	}
	if err := validateServiceEndpointSchemeV2(endpoint); err != nil {
		return "", err
	}
	return scheme, nil
}

func validateVersionedServiceTypeV2(serviceType string) error {
	if err := validateUnifiedRecordValue("identity v2 service_type", serviceType, MaxServiceTypeBytesV2); err != nil {
		return err
	}
	if strings.Contains(serviceType, ".v") || strings.Contains(serviceType, ":v") || strings.Contains(serviceType, "-v") {
		return nil
	}
	return errors.New("identity v2 service_type must be versioned")
}

func interfaceDescriptorSchemaHashV2(descriptor InterfaceDescriptorV2) string {
	if descriptor.SchemaHash != "" {
		return descriptor.SchemaHash
	}
	return descriptor.Descriptor
}

func validateInterfaceSchemaURIV2(uri string) error {
	if len(uri) > MaxInterfaceSchemaURIBytesV2 {
		return fmt.Errorf("identity v2 interface schema_uri_optional must not exceed %d bytes", MaxInterfaceSchemaURIBytesV2)
	}
	if err := validateASCIIValueV2("identity v2 interface schema_uri_optional", uri); err != nil {
		return err
	}
	scheme, rest, found := strings.Cut(uri, "://")
	if !found || rest == "" {
		return errors.New("identity v2 interface schema_uri_optional must include an allowed scheme")
	}
	switch scheme {
	case "https", "ipfs", "ar", "aetra":
		return nil
	default:
		return fmt.Errorf("identity v2 interface schema_uri_optional scheme %q is not allowed", scheme)
	}
}

func validateVersionedInterfaceVersionV2(version string) error {
	if err := validateUnifiedRecordValue("identity v2 interface version", version, MaxInterfaceVersionBytesV2); err != nil {
		return err
	}
	if strings.HasPrefix(version, "v") || strings.Contains(version, ".v") {
		return nil
	}
	return errors.New("identity v2 interface version must be versioned")
}

func validateInterfacePermissionsRequiredV2(permissions []string) error {
	if len(permissions) > MaxInterfacePermissionsV2 {
		return fmt.Errorf("identity v2 interface permissions_required must not exceed %d entries", MaxInterfacePermissionsV2)
	}
	seen := map[string]struct{}{}
	for i, permission := range permissions {
		if err := validateUnifiedRecordValue("identity v2 interface permission", permission, MaxInterfacePermissionBytesV2); err != nil {
			return err
		}
		switch {
		case permission == "execute", permission == "sign", strings.HasPrefix(permission, "execute."), strings.HasPrefix(permission, "grant."):
			return errors.New("identity v2 interface descriptors cannot grant execution permission")
		}
		if _, found := seen[permission]; found {
			return fmt.Errorf("duplicate identity v2 interface permission %q", permission)
		}
		seen[permission] = struct{}{}
		if i > 0 && permissions[i-1] >= permission {
			return errors.New("identity v2 interface permissions_required must be sorted canonically")
		}
	}
	return nil
}

func validateASCIIValueV2(field string, value string) error {
	for i := 0; i < len(value); i++ {
		if value[i] > 0x7f {
			return fmt.Errorf("%s must be ASCII", field)
		}
	}
	return nil
}

func validateRoutingMetadataV2(route RoutingMetadataV2) error {
	totalBytes := len(route.ZoneID) + len(route.ShardID) + len(route.VM) + len(route.Entrypoint) +
		len(route.RouteID) + len(route.TargetType) + len(route.PreferredTarget) + len(route.ChainContext) +
		len(route.FeeHint) + len(route.MemoPolicy)
	for _, target := range route.FallbackTargets {
		totalBytes += len(target)
	}
	for _, capability := range route.CapabilityRequirements {
		totalBytes += len(capability)
	}
	if totalBytes > MaxUnifiedRoutingMetadataBytes {
		return fmt.Errorf("identity v2 routing metadata must not exceed %d bytes", MaxUnifiedRoutingMetadataBytes)
	}
	for field, value := range map[string]string{
		"zone_id":	route.ZoneID,
		"shard_id":	route.ShardID,
		"vm":		route.VM,
		"entrypoint":	route.Entrypoint,
	} {
		if value == "" {
			continue
		}
		if err := validateUnifiedRecordValue("identity v2 routing "+field, value, MaxUnifiedRecordValueBytes); err != nil {
			return err
		}
	}
	newRouteFieldsSet := route.RouteID != "" || route.TargetType != "" || route.PreferredTarget != "" ||
		len(route.FallbackTargets) > 0 || route.ChainContext != "" || route.FeeHint != "" ||
		route.TimeoutHint != 0 || route.MemoPolicy != "" || len(route.CapabilityRequirements) > 0
	if !newRouteFieldsSet {
		return nil
	}
	if err := validateUnifiedRecordKey("identity v2 routing route_id", route.RouteID); err != nil {
		return err
	}
	if err := validateRoutingTargetTypeV2(route.TargetType); err != nil {
		return err
	}
	if err := validateUnifiedRecordValue("identity v2 routing preferred_target", route.PreferredTarget, MaxRouteTargetBytesV2); err != nil {
		return err
	}
	if err := validateSortedRoutingListV2("identity v2 routing fallback target", route.FallbackTargets, MaxRouteTargetBytesV2); err != nil {
		return err
	}
	if route.ChainContext != "" {
		if err := validateUnifiedRecordValue("identity v2 routing chain_context", route.ChainContext, MaxRouteChainContextBytesV2); err != nil {
			return err
		}
	}
	if route.FeeHint != "" {
		if err := validateUnifiedRecordValue("identity v2 routing fee_hint", route.FeeHint, MaxRouteFeeHintBytesV2); err != nil {
			return err
		}
		if containsBypassSemanticsV2(route.FeeHint) {
			return errors.New("identity v2 routing fee hints cannot override fee module requirements")
		}
	}
	if route.TimeoutHint > MaxRouteTimeoutHintV2 {
		return fmt.Errorf("identity v2 routing timeout_hint must not exceed %d seconds", MaxRouteTimeoutHintV2)
	}
	if route.MemoPolicy != "" {
		if err := validateMemoPolicyV2(route.MemoPolicy); err != nil {
			return err
		}
	}
	if err := validateSortedRoutingListV2("identity v2 routing capability requirement", route.CapabilityRequirements, MaxRouteCapabilityBytesV2); err != nil {
		return err
	}
	return nil
}

func validateExecutionHintsV2(hints []ExecutionHintV2) error {
	seen := map[string]struct{}{}
	for i, hint := range hints {
		hintID := executionHintIDV2(hint)
		if err := validateUnifiedRecordKey("identity v2 execution hint key", hintID); err != nil {
			return err
		}
		if !strings.HasPrefix(hintID, "hint.") {
			return errors.New("identity v2 execution hints are advisory and must use hint.* keys")
		}
		if hint.Key != "" && hint.Value != "" {
			if err := validateUnifiedRecordValue("identity v2 execution hint", hint.Value, MaxUnifiedRecordValueBytes); err != nil {
				return err
			}
			if containsBypassSemanticsV2(hint.Value) {
				return errors.New("identity v2 execution hints must not bypass ante-handler checks")
			}
		}
		if hint.DefaultGasLimitHint > MaxExecutionGasLimitHintV2 {
			return fmt.Errorf("identity v2 execution default_gas_limit_hint must not exceed %d", MaxExecutionGasLimitHintV2)
		}
		if hint.PreferredFeeMode != "" {
			if err := validateUnifiedRecordValue("identity v2 execution preferred_fee_mode", hint.PreferredFeeMode, MaxExecutionFeeModeBytesV2); err != nil {
				return err
			}
			if containsBypassSemanticsV2(hint.PreferredFeeMode) {
				return errors.New("identity v2 execution hints must not bypass ante-handler checks")
			}
		}
		if hint.MessageType != "" {
			if err := validateUnifiedRecordValue("identity v2 execution message_type", hint.MessageType, MaxExecutionMessageTypeBytesV2); err != nil {
				return err
			}
		}
		if _, found := seen[hintID]; found {
			return fmt.Errorf("duplicate identity v2 execution hint %q", hintID)
		}
		seen[hintID] = struct{}{}
		if i > 0 && executionHintIDV2(hints[i-1]) >= hintID {
			return errors.New("identity v2 execution hints must be sorted canonically")
		}
	}
	return nil
}

func routingMetadataHasTargetV2(route RoutingMetadataV2) bool {
	return route.ZoneID != "" || route.ShardID != "" || route.VM != "" || route.Entrypoint != "" ||
		route.RouteID != "" || route.TargetType != "" || route.PreferredTarget != "" || len(route.FallbackTargets) > 0
}

func validateRoutingTargetTypeV2(targetType string) error {
	if err := validateUnifiedRecordValue("identity v2 routing target_type", targetType, MaxRouteTargetTypeBytesV2); err != nil {
		return err
	}
	switch IdentityResolutionTargetTypeV2(targetType) {
	case IdentityResolutionTargetPrimary, IdentityResolutionTargetContract, IdentityResolutionTargetService,
		IdentityResolutionTargetInterface, IdentityResolutionTargetRoute, IdentityResolutionTargetRecord:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 routing target_type %q", targetType)
	}
}

func validateSortedRoutingListV2(field string, values []string, maxBytes int) error {
	if len(values) > MaxUnifiedRouteListEntries {
		return fmt.Errorf("%s list must not exceed %d entries", field, MaxUnifiedRouteListEntries)
	}
	seen := map[string]struct{}{}
	for i, value := range values {
		if err := validateUnifiedRecordValue(field, value, maxBytes); err != nil {
			return err
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s %q", field, value)
		}
		seen[value] = struct{}{}
		if i > 0 && values[i-1] >= value {
			return fmt.Errorf("%s list must be sorted canonically", field)
		}
	}
	return nil
}

func validateMemoPolicyV2(policy string) error {
	if err := validateUnifiedRecordValue("identity v2 routing memo_policy", policy, MaxRouteMemoPolicyBytesV2); err != nil {
		return err
	}
	switch policy {
	case "none", "optional", "required", "forbidden":
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 routing memo_policy %q", policy)
	}
}

func executionHintIDV2(hint ExecutionHintV2) string {
	if hint.Key != "" {
		return hint.Key
	}
	if hint.MessageType != "" {
		return "hint.message." + hint.MessageType
	}
	return "hint.default"
}

func executionHintFromResolverMetadataV2(entry ResolverMetadataEntry) (ExecutionHintV2, error) {
	hint := ExecutionHintV2{Key: entry.Key, Value: entry.Value}
	switch entry.Key {
	case "hint.default_gas_limit":
		value, err := strconv.ParseUint(entry.Value, 10, 64)
		if err != nil {
			return ExecutionHintV2{}, fmt.Errorf("identity v2 execution default_gas_limit_hint must be numeric: %w", err)
		}
		hint.DefaultGasLimitHint = value
	case "hint.preferred_fee_mode":
		hint.PreferredFeeMode = entry.Value
	case "hint.message_type":
		hint.MessageType = entry.Value
	case "hint.async_allowed":
		value, err := strconv.ParseBool(entry.Value)
		if err != nil {
			return ExecutionHintV2{}, fmt.Errorf("identity v2 execution async_allowed must be boolean: %w", err)
		}
		hint.AsyncAllowed = value
	case "hint.requires_memo":
		value, err := strconv.ParseBool(entry.Value)
		if err != nil {
			return ExecutionHintV2{}, fmt.Errorf("identity v2 execution requires_memo must be boolean: %w", err)
		}
		hint.RequiresMemo = value
	case "hint.requires_interface_confirmation":
		value, err := strconv.ParseBool(entry.Value)
		if err != nil {
			return ExecutionHintV2{}, fmt.Errorf("identity v2 execution requires_interface_confirmation must be boolean: %w", err)
		}
		hint.RequiresInterfaceConfirmation = value
	case "hint.simulation_required":
		value, err := strconv.ParseBool(entry.Value)
		if err != nil {
			return ExecutionHintV2{}, fmt.Errorf("identity v2 execution simulation_required must be boolean: %w", err)
		}
		hint.SimulationRequired = value
	}
	return hint, nil
}

func splitCanonicalCSVV2(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, part)
	}
	return out
}

func containsBypassSemanticsV2(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "override") || strings.Contains(lower, "bypass") ||
		strings.Contains(lower, "no-fee") || strings.Contains(lower, "zero-fee") || lower == "free"
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
	case "https", "grpcs", "wss", "aetra", "ipfs":
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
