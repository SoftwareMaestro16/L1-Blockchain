package types

import (
	"errors"
	"fmt"
	"sort"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type CanonicalServiceEndpointType string
type CanonicalServiceStatus string

const (
	CanonicalEndpointApplication	CanonicalServiceEndpointType	= "application"
	CanonicalEndpointAPI		CanonicalServiceEndpointType	= "api"
	CanonicalEndpointOffChain	CanonicalServiceEndpointType	= "off_chain_compute"
	CanonicalEndpointHybrid		CanonicalServiceEndpointType	= "hybrid"
	CanonicalEndpointZoneAware	CanonicalServiceEndpointType	= "zone_aware"

	CanonicalServiceStatusActive	CanonicalServiceStatus	= "active"
	CanonicalServiceStatusDisabled	CanonicalServiceStatus	= "disabled"
	CanonicalServiceStatusExpired	CanonicalServiceStatus	= "expired"
)

type CanonicalServiceDescriptor struct {
	ServiceID		string
	EndpointType		CanonicalServiceEndpointType
	InterfaceHash		string
	SupportedMethods	[]string
	AuthModel		string
	StateDependency		string
	Owner			string
	ZoneID			string
	Version			uint64
	EndpointURIHash		string
	MetadataHash		string
	TTLHeight		uint64
	Status			CanonicalServiceStatus
	Capabilities		[]string
	DescriptorHash		string
}

func NewCanonicalServiceDescriptor(descriptor CanonicalServiceDescriptor) (CanonicalServiceDescriptor, error) {
	if descriptor.DescriptorHash != "" {
		return CanonicalServiceDescriptor{}, errors.New("canonical service descriptor hash must be empty before construction")
	}
	descriptor = CanonicalizeServiceDescriptor(descriptor)
	if err := descriptor.ValidateFormat(); err != nil {
		return CanonicalServiceDescriptor{}, err
	}
	descriptor.DescriptorHash = ComputeCanonicalServiceDescriptorHash(descriptor)
	return descriptor, descriptor.Validate()
}

func CanonicalizeServiceDescriptor(descriptor CanonicalServiceDescriptor) CanonicalServiceDescriptor {
	descriptor.SupportedMethods = normalizeDescriptorTokens(descriptor.SupportedMethods)
	descriptor.Capabilities = normalizeDescriptorTokens(descriptor.Capabilities)
	return descriptor
}

func ProjectDistributedServiceDescriptor(record DistributedServiceRecord, endpoint DistributedServiceEndpoint, iface DistributedInterfaceDescriptor, methods []string, authModel string, stateDependency string, capabilities []string) (CanonicalServiceDescriptor, error) {
	if err := record.Validate(); err != nil {
		return CanonicalServiceDescriptor{}, err
	}
	if endpoint.ServiceID != record.ServiceID {
		return CanonicalServiceDescriptor{}, errors.New("canonical descriptor endpoint service mismatch")
	}
	if err := endpoint.Validate(); err != nil {
		return CanonicalServiceDescriptor{}, err
	}
	if err := iface.Validate(); err != nil {
		return CanonicalServiceDescriptor{}, err
	}
	if endpoint.InterfaceHash != record.InterfaceHash || iface.InterfaceHash != record.InterfaceHash {
		return CanonicalServiceDescriptor{}, errors.New("canonical descriptor interface mismatch")
	}
	status := CanonicalServiceStatusActive
	if !record.Discoverable {
		status = CanonicalServiceStatusDisabled
	}
	return NewCanonicalServiceDescriptor(CanonicalServiceDescriptor{
		ServiceID:		record.ServiceID,
		EndpointType:		CanonicalServiceEndpointType(endpoint.Kind),
		InterfaceHash:		record.InterfaceHash,
		SupportedMethods:	methods,
		AuthModel:		authModel,
		StateDependency:	stateDependency,
		Owner:			record.Owner,
		ZoneID:			record.ZoneID,
		Version:		iface.Version,
		EndpointURIHash:	endpoint.CommitmentHash,
		MetadataHash:		record.MetadataHash,
		TTLHeight:		record.ExpiryHeight,
		Status:			status,
		Capabilities:		capabilities,
	})
}

func (descriptor CanonicalServiceDescriptor) ValidateFormat() error {
	if _, err := DistributedServiceRecordKey(descriptor.ServiceID); err != nil {
		return err
	}
	if !IsCanonicalEndpointType(descriptor.EndpointType) {
		return fmt.Errorf("unknown canonical service endpoint type %q", descriptor.EndpointType)
	}
	if err := coretypes.ValidateHash("canonical service interface hash", descriptor.InterfaceHash); err != nil {
		return err
	}
	if err := validateDescriptorTokenList("canonical service supported method", descriptor.SupportedMethods); err != nil {
		return err
	}
	if err := validateInterfaceToken("canonical service auth model", descriptor.AuthModel); err != nil {
		return err
	}
	if err := validateInterfaceToken("canonical service state dependency", descriptor.StateDependency); err != nil {
		return err
	}
	if err := validateInterfaceToken("canonical service owner", descriptor.Owner); err != nil {
		return err
	}
	if err := validateInterfaceToken("canonical service zone id", descriptor.ZoneID); err != nil {
		return err
	}
	if descriptor.Version == 0 {
		return errors.New("canonical service version must be positive")
	}
	if err := coretypes.ValidateHash("canonical service endpoint uri hash", descriptor.EndpointURIHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("canonical service metadata hash", descriptor.MetadataHash); err != nil {
		return err
	}
	if descriptor.TTLHeight == 0 {
		return errors.New("canonical service ttl height must be positive")
	}
	if !IsCanonicalServiceStatus(descriptor.Status) {
		return fmt.Errorf("unknown canonical service status %q", descriptor.Status)
	}
	if err := validateDescriptorTokenList("canonical service capability", descriptor.Capabilities); err != nil {
		return err
	}
	if descriptor.DescriptorHash != "" {
		return coretypes.ValidateHash("canonical service descriptor hash", descriptor.DescriptorHash)
	}
	return nil
}

func (descriptor CanonicalServiceDescriptor) Validate() error {
	descriptor = CanonicalizeServiceDescriptor(descriptor)
	if err := descriptor.ValidateFormat(); err != nil {
		return err
	}
	if descriptor.DescriptorHash == "" {
		return errors.New("canonical service descriptor hash is required")
	}
	if descriptor.DescriptorHash != ComputeCanonicalServiceDescriptorHash(descriptor) {
		return errors.New("canonical service descriptor hash mismatch")
	}
	return nil
}

func ComputeCanonicalServiceDescriptorHash(descriptor CanonicalServiceDescriptor) string {
	descriptor = CanonicalizeServiceDescriptor(descriptor)
	parts := []string{
		"aetra-services-canonical-descriptor-v1",
		descriptor.ServiceID,
		string(descriptor.EndpointType),
		descriptor.InterfaceHash,
		descriptor.AuthModel,
		descriptor.StateDependency,
		descriptor.Owner,
		descriptor.ZoneID,
		fmt.Sprint(descriptor.Version),
		descriptor.EndpointURIHash,
		descriptor.MetadataHash,
		fmt.Sprint(descriptor.TTLHeight),
		string(descriptor.Status),
		"methods",
		fmt.Sprint(len(descriptor.SupportedMethods)),
	}
	parts = append(parts, descriptor.SupportedMethods...)
	parts = append(parts, "capabilities", fmt.Sprint(len(descriptor.Capabilities)))
	parts = append(parts, descriptor.Capabilities...)
	return servicesHashParts(parts...)
}

func IsCanonicalEndpointType(endpointType CanonicalServiceEndpointType) bool {
	switch endpointType {
	case CanonicalEndpointApplication, CanonicalEndpointAPI, CanonicalEndpointOffChain, CanonicalEndpointHybrid, CanonicalEndpointZoneAware:
		return true
	default:
		return false
	}
}

func IsCanonicalServiceStatus(status CanonicalServiceStatus) bool {
	switch status {
	case CanonicalServiceStatusActive, CanonicalServiceStatusDisabled, CanonicalServiceStatusExpired:
		return true
	default:
		return false
	}
}

func validateDescriptorTokenList(fieldName string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s list must not be empty", fieldName)
	}
	previous := ""
	for _, value := range values {
		if err := validateInterfaceToken(fieldName, value); err != nil {
			return err
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("%s list must be sorted canonically", fieldName)
		}
		previous = value
	}
	return nil
}

func normalizeDescriptorTokens(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	unique := make([]string, 0, len(out))
	for _, value := range out {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}
