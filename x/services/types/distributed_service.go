package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const (
	DistributedServiceRecordPrefix		= ServiceStorePrefix + "distributed/records"
	DistributedServiceEndpointPrefix	= ServiceStorePrefix + "distributed/endpoints"
	DistributedServiceInterfacePrefix	= ServiceStorePrefix + "distributed/interfaces"
	DistributedServiceProofPrefix		= ServiceStorePrefix + "distributed/proofs"
)

type DistributedServiceKind string
type DistributedEndpointKind string
type DistributedExecutionAuthority string
type DistributedCommitmentKind string

const (
	DistributedServiceApplication	DistributedServiceKind	= "application"
	DistributedServiceAPI		DistributedServiceKind	= "api"
	DistributedServiceOffChain	DistributedServiceKind	= "off_chain_compute"
	DistributedServiceHybrid	DistributedServiceKind	= "hybrid"

	DistributedEndpointApplication	DistributedEndpointKind	= "application"
	DistributedEndpointAPI		DistributedEndpointKind	= "api"
	DistributedEndpointCompute	DistributedEndpointKind	= "off_chain_compute"
	DistributedEndpointHybrid	DistributedEndpointKind	= "hybrid"
	DistributedEndpointZoneAware	DistributedEndpointKind	= "zone_aware"

	DistributedAuthorityOnChain		DistributedExecutionAuthority	= "on_chain"
	DistributedAuthorityMessageCommit	DistributedExecutionAuthority	= "message_commit"
	DistributedAuthorityProofCommit		DistributedExecutionAuthority	= "proof_commit"
	DistributedAuthorityAdvisory		DistributedExecutionAuthority	= "advisory"

	DistributedCommitmentMessage	DistributedCommitmentKind	= "message"
	DistributedCommitmentProof	DistributedCommitmentKind	= "proof"
)

type DistributedServiceRecord struct {
	ServiceID	string
	ServiceName	string
	Kind		DistributedServiceKind
	Owner		string
	ZoneID		string
	InterfaceHash	string
	EndpointRoot	string
	DescriptorHash	string
	MetadataHash	string
	CreatedHeight	uint64
	UpdatedHeight	uint64
	ExpiryHeight	uint64
	Discoverable	bool
	RecordHash	string
}

type DistributedServiceEndpoint struct {
	ServiceID	string
	EndpointID	string
	Kind		DistributedEndpointKind
	ZoneID		string
	Target		string
	InterfaceHash	string
	Priority	uint32
	Weight		uint32
	MetadataHash	string
	CommitmentHash	string
}

type DistributedInterfaceDescriptor struct {
	InterfaceHash	string
	InterfaceName	string
	Version		uint64
	SchemaHash	string
	MethodRoot	string
	EventRoot	string
	ErrorRoot	string
	DescriptorHash	string
}

type DistributedExecutionCommitment struct {
	ServiceID	string
	EndpointID	string
	Kind		DistributedCommitmentKind
	MessageID	string
	ProofHash	string
	ResultHash	string
	CommittedHeight	uint64
	CommitmentHash	string
}

type DistributedServiceDiscoveryState struct {
	Records		[]DistributedServiceRecord
	Endpoints	[]DistributedServiceEndpoint
	Interfaces	[]DistributedInterfaceDescriptor
	Commitments	[]DistributedExecutionCommitment
	Height		uint64
	StateRoot	string
}

type DistributedServiceDiscoveryRoots struct {
	RecordRoot	string
	EndpointRoot	string
	InterfaceRoot	string
	CommitmentRoot	string
	StateRoot	string
}

func NewDistributedServiceRecord(record DistributedServiceRecord) (DistributedServiceRecord, error) {
	if record.RecordHash != "" {
		return DistributedServiceRecord{}, errors.New("distributed service record hash must be empty before construction")
	}
	if err := record.ValidateFormat(); err != nil {
		return DistributedServiceRecord{}, err
	}
	record.RecordHash = ComputeDistributedServiceRecordHash(record)
	return record, record.Validate()
}

func NewDistributedServiceEndpoint(endpoint DistributedServiceEndpoint) (DistributedServiceEndpoint, error) {
	if endpoint.CommitmentHash != "" {
		return DistributedServiceEndpoint{}, errors.New("distributed service endpoint commitment hash must be empty before construction")
	}
	if err := endpoint.ValidateFormat(); err != nil {
		return DistributedServiceEndpoint{}, err
	}
	endpoint.CommitmentHash = ComputeDistributedServiceEndpointHash(endpoint)
	return endpoint, endpoint.Validate()
}

func NewDistributedInterfaceDescriptor(descriptor DistributedInterfaceDescriptor) (DistributedInterfaceDescriptor, error) {
	if descriptor.DescriptorHash != "" {
		return DistributedInterfaceDescriptor{}, errors.New("distributed interface descriptor hash must be empty before construction")
	}
	if err := descriptor.ValidateFormat(); err != nil {
		return DistributedInterfaceDescriptor{}, err
	}
	descriptor.DescriptorHash = ComputeDistributedInterfaceDescriptorHash(descriptor)
	return descriptor, descriptor.Validate()
}

func NewDistributedExecutionCommitment(commitment DistributedExecutionCommitment) (DistributedExecutionCommitment, error) {
	if commitment.CommitmentHash != "" {
		return DistributedExecutionCommitment{}, errors.New("distributed execution commitment hash must be empty before construction")
	}
	if err := commitment.ValidateFormat(); err != nil {
		return DistributedExecutionCommitment{}, err
	}
	commitment.CommitmentHash = ComputeDistributedExecutionCommitmentHash(commitment)
	return commitment, commitment.Validate()
}

func DistributedServiceRecordKey(serviceID string) (string, error) {
	if err := validateInterfaceToken("distributed service id", serviceID); err != nil {
		return "", err
	}
	return DistributedServiceRecordPrefix + "/" + serviceID, nil
}

func DistributedServiceEndpointKey(serviceID, endpointID string) (string, error) {
	if err := validateInterfaceToken("distributed endpoint service id", serviceID); err != nil {
		return "", err
	}
	if err := validateInterfaceToken("distributed endpoint id", endpointID); err != nil {
		return "", err
	}
	return DistributedServiceEndpointPrefix + "/" + serviceID + "/" + endpointID, nil
}

func DistributedServiceInterfaceKey(interfaceHash string) (string, error) {
	if err := coretypes.ValidateHash("distributed interface hash", interfaceHash); err != nil {
		return "", err
	}
	return DistributedServiceInterfacePrefix + "/" + interfaceHash, nil
}

func DistributedServiceProofKey(serviceID, commitmentHash string) (string, error) {
	if err := validateInterfaceToken("distributed proof service id", serviceID); err != nil {
		return "", err
	}
	if err := coretypes.ValidateHash("distributed proof commitment hash", commitmentHash); err != nil {
		return "", err
	}
	return DistributedServiceProofPrefix + "/" + serviceID + "/" + commitmentHash, nil
}

func BuildDistributedDiscoveryState(records []DistributedServiceRecord, endpoints []DistributedServiceEndpoint, interfaces []DistributedInterfaceDescriptor, commitments []DistributedExecutionCommitment, height uint64) (DistributedServiceDiscoveryState, error) {
	state := DistributedServiceDiscoveryState{
		Records:	normalizeDistributedServiceRecords(records),
		Endpoints:	normalizeDistributedServiceEndpoints(endpoints),
		Interfaces:	normalizeDistributedInterfaces(interfaces),
		Commitments:	normalizeDistributedExecutionCommitments(commitments),
		Height:		height,
	}
	if err := state.ValidateFormat(); err != nil {
		return DistributedServiceDiscoveryState{}, err
	}
	state.StateRoot = ComputeDistributedServiceDiscoveryStateRoot(state)
	return state, state.Validate()
}

func DiscoverDistributedServices(state DistributedServiceDiscoveryState, kind DistributedServiceKind, zoneID string, interfaceHash string, height uint64) ([]DistributedServiceRecord, error) {
	if err := state.Validate(); err != nil {
		return nil, err
	}
	if height == 0 {
		return nil, errors.New("distributed discovery height must be positive")
	}
	if kind != "" && !IsDistributedServiceKind(kind) {
		return nil, fmt.Errorf("unknown distributed service kind %q", kind)
	}
	if zoneID != "" {
		if err := validateInterfaceToken("distributed discovery zone id", zoneID); err != nil {
			return nil, err
		}
	}
	if interfaceHash != "" {
		if err := coretypes.ValidateHash("distributed discovery interface hash", interfaceHash); err != nil {
			return nil, err
		}
	}
	out := make([]DistributedServiceRecord, 0)
	for _, record := range state.Records {
		if !record.Discoverable || record.ExpiryHeight < height {
			continue
		}
		if kind != "" && record.Kind != kind {
			continue
		}
		if zoneID != "" && record.ZoneID != zoneID {
			continue
		}
		if interfaceHash != "" && record.InterfaceHash != interfaceHash {
			continue
		}
		out = append(out, record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ServiceName != out[j].ServiceName {
			return out[i].ServiceName < out[j].ServiceName
		}
		return out[i].ServiceID < out[j].ServiceID
	})
	return out, nil
}

func ValidateOffConsensusAuthority(authority DistributedExecutionAuthority, commitments []DistributedExecutionCommitment) error {
	switch authority {
	case DistributedAuthorityOnChain, DistributedAuthorityMessageCommit, DistributedAuthorityProofCommit, DistributedAuthorityAdvisory:
	default:
		return fmt.Errorf("unknown distributed execution authority %q", authority)
	}
	if authority == DistributedAuthorityOnChain || authority == DistributedAuthorityAdvisory {
		return nil
	}
	if len(commitments) == 0 {
		return errors.New("off-consensus service execution requires message or proof commitment")
	}
	for _, commitment := range commitments {
		if err := commitment.Validate(); err != nil {
			return err
		}
		if authority == DistributedAuthorityMessageCommit && commitment.Kind != DistributedCommitmentMessage {
			return errors.New("message authority requires message commitment")
		}
		if authority == DistributedAuthorityProofCommit && commitment.Kind != DistributedCommitmentProof {
			return errors.New("proof authority requires proof commitment")
		}
	}
	return nil
}

func (record DistributedServiceRecord) ValidateFormat() error {
	if _, err := DistributedServiceRecordKey(record.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("distributed service name", record.ServiceName); err != nil {
		return err
	}
	if !IsDistributedServiceKind(record.Kind) {
		return fmt.Errorf("unknown distributed service kind %q", record.Kind)
	}
	if err := validateInterfaceToken("distributed service owner", record.Owner); err != nil {
		return err
	}
	if err := validateInterfaceToken("distributed service zone id", record.ZoneID); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "distributed service interface hash", value: record.InterfaceHash},
		{name: "distributed service endpoint root", value: record.EndpointRoot},
		{name: "distributed service descriptor hash", value: record.DescriptorHash},
		{name: "distributed service metadata hash", value: record.MetadataHash},
	} {
		if err := coretypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if record.CreatedHeight == 0 || record.UpdatedHeight == 0 || record.ExpiryHeight == 0 {
		return errors.New("distributed service heights must be positive")
	}
	if record.UpdatedHeight < record.CreatedHeight {
		return errors.New("distributed service updated height must not precede created height")
	}
	if record.ExpiryHeight < record.UpdatedHeight {
		return errors.New("distributed service expiry must not precede updated height")
	}
	if record.RecordHash != "" {
		return coretypes.ValidateHash("distributed service record hash", record.RecordHash)
	}
	return nil
}

func (record DistributedServiceRecord) Validate() error {
	if err := record.ValidateFormat(); err != nil {
		return err
	}
	if record.RecordHash == "" {
		return errors.New("distributed service record hash is required")
	}
	if record.RecordHash != ComputeDistributedServiceRecordHash(record) {
		return errors.New("distributed service record hash mismatch")
	}
	return nil
}

func (endpoint DistributedServiceEndpoint) ValidateFormat() error {
	if _, err := DistributedServiceEndpointKey(endpoint.ServiceID, endpoint.EndpointID); err != nil {
		return err
	}
	if !IsDistributedEndpointKind(endpoint.Kind) {
		return fmt.Errorf("unknown distributed endpoint kind %q", endpoint.Kind)
	}
	if err := validateInterfaceToken("distributed endpoint zone id", endpoint.ZoneID); err != nil {
		return err
	}
	if err := validateDistributedEndpointTarget(endpoint.Kind, endpoint.Target); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("distributed endpoint interface hash", endpoint.InterfaceHash); err != nil {
		return err
	}
	if endpoint.Weight == 0 {
		return errors.New("distributed endpoint weight must be positive")
	}
	if endpoint.MetadataHash != "" {
		if err := coretypes.ValidateHash("distributed endpoint metadata hash", endpoint.MetadataHash); err != nil {
			return err
		}
	}
	if endpoint.CommitmentHash != "" {
		return coretypes.ValidateHash("distributed endpoint commitment hash", endpoint.CommitmentHash)
	}
	return nil
}

func (endpoint DistributedServiceEndpoint) Validate() error {
	if err := endpoint.ValidateFormat(); err != nil {
		return err
	}
	if endpoint.CommitmentHash == "" {
		return errors.New("distributed endpoint commitment hash is required")
	}
	if endpoint.CommitmentHash != ComputeDistributedServiceEndpointHash(endpoint) {
		return errors.New("distributed endpoint commitment hash mismatch")
	}
	return nil
}

func (descriptor DistributedInterfaceDescriptor) ValidateFormat() error {
	if _, err := DistributedServiceInterfaceKey(descriptor.InterfaceHash); err != nil {
		return err
	}
	if err := validateInterfaceToken("distributed interface name", descriptor.InterfaceName); err != nil {
		return err
	}
	if descriptor.Version == 0 {
		return errors.New("distributed interface version must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "distributed interface schema hash", value: descriptor.SchemaHash},
		{name: "distributed interface method root", value: descriptor.MethodRoot},
		{name: "distributed interface event root", value: descriptor.EventRoot},
		{name: "distributed interface error root", value: descriptor.ErrorRoot},
	} {
		if err := coretypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if descriptor.DescriptorHash != "" {
		return coretypes.ValidateHash("distributed interface descriptor hash", descriptor.DescriptorHash)
	}
	return nil
}

func (descriptor DistributedInterfaceDescriptor) Validate() error {
	if err := descriptor.ValidateFormat(); err != nil {
		return err
	}
	if descriptor.DescriptorHash == "" {
		return errors.New("distributed interface descriptor hash is required")
	}
	if descriptor.DescriptorHash != ComputeDistributedInterfaceDescriptorHash(descriptor) {
		return errors.New("distributed interface descriptor hash mismatch")
	}
	return nil
}

func (commitment DistributedExecutionCommitment) ValidateFormat() error {
	if err := validateInterfaceToken("distributed commitment service id", commitment.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("distributed commitment endpoint id", commitment.EndpointID); err != nil {
		return err
	}
	if !IsDistributedCommitmentKind(commitment.Kind) {
		return fmt.Errorf("unknown distributed commitment kind %q", commitment.Kind)
	}
	if commitment.Kind == DistributedCommitmentMessage {
		if err := coretypes.ValidateHash("distributed commitment message id", commitment.MessageID); err != nil {
			return err
		}
	}
	if commitment.Kind == DistributedCommitmentProof {
		if err := coretypes.ValidateHash("distributed commitment proof hash", commitment.ProofHash); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("distributed commitment result hash", commitment.ResultHash); err != nil {
		return err
	}
	if commitment.CommittedHeight == 0 {
		return errors.New("distributed commitment height must be positive")
	}
	if commitment.CommitmentHash != "" {
		return coretypes.ValidateHash("distributed execution commitment hash", commitment.CommitmentHash)
	}
	return nil
}

func (commitment DistributedExecutionCommitment) Validate() error {
	if err := commitment.ValidateFormat(); err != nil {
		return err
	}
	if commitment.CommitmentHash == "" {
		return errors.New("distributed execution commitment hash is required")
	}
	if commitment.CommitmentHash != ComputeDistributedExecutionCommitmentHash(commitment) {
		return errors.New("distributed execution commitment hash mismatch")
	}
	return nil
}

func (state DistributedServiceDiscoveryState) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("distributed discovery state height must be positive")
	}
	if err := validateDistributedServiceRecords(state.Records); err != nil {
		return err
	}
	if err := validateDistributedEndpoints(state.Endpoints, state.Records); err != nil {
		return err
	}
	if err := validateDistributedInterfaces(state.Interfaces, state.Records); err != nil {
		return err
	}
	for _, commitment := range state.Commitments {
		if err := commitment.Validate(); err != nil {
			return err
		}
	}
	if state.StateRoot != "" {
		return coretypes.ValidateHash("distributed discovery state root", state.StateRoot)
	}
	return nil
}

func (state DistributedServiceDiscoveryState) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRoot == "" {
		return errors.New("distributed discovery state root is required")
	}
	if state.StateRoot != ComputeDistributedServiceDiscoveryStateRoot(state) {
		return errors.New("distributed discovery state root mismatch")
	}
	return nil
}

func ComputeDistributedServiceRecordHash(record DistributedServiceRecord) string {
	return servicesHashParts("aetra-services-distributed-record-v1", record.ServiceID, record.ServiceName, string(record.Kind), record.Owner, record.ZoneID, record.InterfaceHash, record.EndpointRoot, record.DescriptorHash, record.MetadataHash, fmt.Sprint(record.CreatedHeight), fmt.Sprint(record.UpdatedHeight), fmt.Sprint(record.ExpiryHeight), fmt.Sprint(record.Discoverable))
}

func ComputeDistributedServiceEndpointHash(endpoint DistributedServiceEndpoint) string {
	return servicesHashParts("aetra-services-distributed-endpoint-v1", endpoint.ServiceID, endpoint.EndpointID, string(endpoint.Kind), endpoint.ZoneID, endpoint.Target, endpoint.InterfaceHash, fmt.Sprint(endpoint.Priority), fmt.Sprint(endpoint.Weight), endpoint.MetadataHash)
}

func ComputeDistributedInterfaceDescriptorHash(descriptor DistributedInterfaceDescriptor) string {
	return servicesHashParts("aetra-services-distributed-interface-v1", descriptor.InterfaceHash, descriptor.InterfaceName, fmt.Sprint(descriptor.Version), descriptor.SchemaHash, descriptor.MethodRoot, descriptor.EventRoot, descriptor.ErrorRoot)
}

func ComputeDistributedExecutionCommitmentHash(commitment DistributedExecutionCommitment) string {
	return servicesHashParts("aetra-services-distributed-execution-commitment-v1", commitment.ServiceID, commitment.EndpointID, string(commitment.Kind), commitment.MessageID, commitment.ProofHash, commitment.ResultHash, fmt.Sprint(commitment.CommittedHeight))
}

func ComputeDistributedEndpointRoot(endpoints []DistributedServiceEndpoint) string {
	ordered := normalizeDistributedServiceEndpoints(endpoints)
	parts := []string{"aetra-services-distributed-endpoint-root-v1", fmt.Sprint(len(ordered))}
	for _, endpoint := range ordered {
		parts = append(parts, endpoint.CommitmentHash)
	}
	return servicesHashParts(parts...)
}

func ComputeDistributedRecordRoot(records []DistributedServiceRecord) string {
	ordered := normalizeDistributedServiceRecords(records)
	parts := []string{"aetra-services-distributed-record-root-v1", fmt.Sprint(len(ordered))}
	for _, record := range ordered {
		parts = append(parts, record.RecordHash)
	}
	return servicesHashParts(parts...)
}

func ComputeDistributedInterfaceRoot(interfaces []DistributedInterfaceDescriptor) string {
	ordered := normalizeDistributedInterfaces(interfaces)
	parts := []string{"aetra-services-distributed-interface-root-v1", fmt.Sprint(len(ordered))}
	for _, descriptor := range ordered {
		parts = append(parts, descriptor.DescriptorHash)
	}
	return servicesHashParts(parts...)
}

func ComputeDistributedCommitmentRoot(commitments []DistributedExecutionCommitment) string {
	ordered := normalizeDistributedExecutionCommitments(commitments)
	parts := []string{"aetra-services-distributed-commitment-root-v1", fmt.Sprint(len(ordered))}
	for _, commitment := range ordered {
		parts = append(parts, commitment.CommitmentHash)
	}
	return servicesHashParts(parts...)
}

func ComputeDistributedServiceDiscoveryRoots(state DistributedServiceDiscoveryState) DistributedServiceDiscoveryRoots {
	roots := DistributedServiceDiscoveryRoots{
		RecordRoot:	ComputeDistributedRecordRoot(state.Records),
		EndpointRoot:	ComputeDistributedEndpointRoot(state.Endpoints),
		InterfaceRoot:	ComputeDistributedInterfaceRoot(state.Interfaces),
		CommitmentRoot:	ComputeDistributedCommitmentRoot(state.Commitments),
	}
	roots.StateRoot = servicesHashParts("aetra-services-distributed-state-root-v1", roots.RecordRoot, roots.EndpointRoot, roots.InterfaceRoot, roots.CommitmentRoot, fmt.Sprint(state.Height))
	return roots
}

func ComputeDistributedServiceDiscoveryStateRoot(state DistributedServiceDiscoveryState) string {
	return ComputeDistributedServiceDiscoveryRoots(state).StateRoot
}

func IsDistributedServiceKind(kind DistributedServiceKind) bool {
	switch kind {
	case DistributedServiceApplication, DistributedServiceAPI, DistributedServiceOffChain, DistributedServiceHybrid:
		return true
	default:
		return false
	}
}

func IsDistributedEndpointKind(kind DistributedEndpointKind) bool {
	switch kind {
	case DistributedEndpointApplication, DistributedEndpointAPI, DistributedEndpointCompute, DistributedEndpointHybrid, DistributedEndpointZoneAware:
		return true
	default:
		return false
	}
}

func IsDistributedCommitmentKind(kind DistributedCommitmentKind) bool {
	switch kind {
	case DistributedCommitmentMessage, DistributedCommitmentProof:
		return true
	default:
		return false
	}
}

func normalizeDistributedServiceRecords(records []DistributedServiceRecord) []DistributedServiceRecord {
	out := append([]DistributedServiceRecord(nil), records...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ServiceID < out[j].ServiceID })
	return out
}

func normalizeDistributedServiceEndpoints(endpoints []DistributedServiceEndpoint) []DistributedServiceEndpoint {
	out := append([]DistributedServiceEndpoint(nil), endpoints...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ServiceID != out[j].ServiceID {
			return out[i].ServiceID < out[j].ServiceID
		}
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		return out[i].EndpointID < out[j].EndpointID
	})
	return out
}

func normalizeDistributedInterfaces(interfaces []DistributedInterfaceDescriptor) []DistributedInterfaceDescriptor {
	out := append([]DistributedInterfaceDescriptor(nil), interfaces...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].InterfaceHash < out[j].InterfaceHash })
	return out
}

func normalizeDistributedExecutionCommitments(commitments []DistributedExecutionCommitment) []DistributedExecutionCommitment {
	out := append([]DistributedExecutionCommitment(nil), commitments...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ServiceID != out[j].ServiceID {
			return out[i].ServiceID < out[j].ServiceID
		}
		if out[i].CommittedHeight != out[j].CommittedHeight {
			return out[i].CommittedHeight < out[j].CommittedHeight
		}
		return out[i].CommitmentHash < out[j].CommitmentHash
	})
	return out
}

func validateDistributedServiceRecords(records []DistributedServiceRecord) error {
	previous := ""
	seen := map[string]struct{}{}
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if _, found := seen[record.ServiceID]; found {
			return fmt.Errorf("duplicate distributed service %s", record.ServiceID)
		}
		seen[record.ServiceID] = struct{}{}
		if previous != "" && previous >= record.ServiceID {
			return errors.New("distributed service records must be sorted canonically")
		}
		previous = record.ServiceID
	}
	return nil
}

func validateDistributedEndpoints(endpoints []DistributedServiceEndpoint, records []DistributedServiceRecord) error {
	services := map[string]DistributedServiceRecord{}
	for _, record := range records {
		services[record.ServiceID] = record
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, endpoint := range endpoints {
		if err := endpoint.Validate(); err != nil {
			return err
		}
		key, _ := DistributedServiceEndpointKey(endpoint.ServiceID, endpoint.EndpointID)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate distributed endpoint %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("distributed endpoints must be sorted canonically")
		}
		previous = key
		record, found := services[endpoint.ServiceID]
		if !found {
			return fmt.Errorf("distributed endpoint references missing service %s", endpoint.ServiceID)
		}
		if endpoint.InterfaceHash != record.InterfaceHash {
			return fmt.Errorf("distributed endpoint interface mismatch %s", endpoint.EndpointID)
		}
	}
	return nil
}

func validateDistributedInterfaces(interfaces []DistributedInterfaceDescriptor, records []DistributedServiceRecord) error {
	available := map[string]struct{}{}
	previous := ""
	for _, descriptor := range interfaces {
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= descriptor.InterfaceHash {
			return errors.New("distributed interfaces must be sorted canonically")
		}
		previous = descriptor.InterfaceHash
		available[descriptor.InterfaceHash] = struct{}{}
	}
	for _, record := range records {
		if _, found := available[record.InterfaceHash]; !found {
			return fmt.Errorf("distributed service %s missing interface descriptor", record.ServiceID)
		}
	}
	return nil
}

func validateDistributedEndpointTarget(kind DistributedEndpointKind, target string) error {
	if strings.TrimSpace(target) != target || target == "" {
		return errors.New("distributed endpoint target is required")
	}
	if len(target) > 256 {
		return errors.New("distributed endpoint target must be <= 256 bytes")
	}
	if kind == DistributedEndpointAPI {
		if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "aether://") {
			return errors.New("distributed API endpoint must use https:// or aether://")
		}
		return nil
	}
	return validateInterfaceToken("distributed endpoint target", target)
}
