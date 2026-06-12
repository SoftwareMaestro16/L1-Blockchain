package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const CoreMessageSpecVersion = uint64(1)

type CoreMessageType string

const (
	CoreMsgZoneTransfer		CoreMessageType	= "MsgZoneTransfer"
	CoreMsgCrossZoneCall		CoreMessageType	= "MsgCrossZoneCall"
	CoreMsgShardCall		CoreMessageType	= "MsgShardCall"
	CoreMsgContractCall		CoreMessageType	= "MsgContractCall"
	CoreMsgIdentityLookup		CoreMessageType	= "MsgIdentityLookup"
	CoreMsgIdentityLookupResult	CoreMessageType	= "MsgIdentityLookupResult"
	CoreMsgPaymentRoute		CoreMessageType	= "MsgPaymentRoute"
	CoreMsgPaymentSettle		CoreMessageType	= "MsgPaymentSettle"
	CoreMsgPromiseResolve		CoreMessageType	= "MsgPromiseResolve"
	CoreMsgPromiseTimeout		CoreMessageType	= "MsgPromiseTimeout"
	CoreMsgProofSubmit		CoreMessageType	= "MsgProofSubmit"
	CoreMsgBounce			CoreMessageType	= "MsgBounce"
)

type CoreMessageTypeDescriptor struct {
	MessageType	CoreMessageType
	Purpose		string
	PrimaryZone	string
	DescriptorHash	string
}

type CoreMessageTypeSpec struct {
	Version	uint64
	Types	[]CoreMessageTypeDescriptor
	Root	string
}

func CoreMessageTypeDescriptors() []CoreMessageTypeDescriptor {
	return []CoreMessageTypeDescriptor{
		coreMessageType(CoreMsgZoneTransfer, "Transfer value between zones through escrow and message receipts.", "Financial Zone"),
		coreMessageType(CoreMsgCrossZoneCall, "Invoke a module, service, or contract in another zone.", "Source and destination zones"),
		coreMessageType(CoreMsgShardCall, "Route a call between shards inside one zone.", "Owning zone"),
		coreMessageType(CoreMsgContractCall, "Execute an AVM contract method in Contract Zone.", "Contract Zone"),
		coreMessageType(CoreMsgIdentityLookup, "Request .aet name, reverse, resolver, or ownership lookup.", "Identity Zone"),
		coreMessageType(CoreMsgIdentityLookupResult, "Return identity lookup result with optional proof.", "Requester zone"),
		coreMessageType(CoreMsgPaymentRoute, "Reserve or execute payment route across channels, shards, or zones.", "Financial Zone"),
		coreMessageType(CoreMsgPaymentSettle, "Settle, refund, expire, or dispute a payment.", "Financial Zone"),
		coreMessageType(CoreMsgPromiseResolve, "Resolve asynchronous contract or service promise.", "Destination promise owner"),
		coreMessageType(CoreMsgPromiseTimeout, "Timeout unresolved promise and emit deterministic receipt.", "Destination promise owner"),
		coreMessageType(CoreMsgProofSubmit, "Submit proof-backed state, message, identity, contract, or payment evidence.", "Proof verifier target"),
		coreMessageType(CoreMsgBounce, "Return value, fee, or failure payload after failed delivery.", "Source zone"),
	}
}

func BuildCoreMessageTypeSpec(types []CoreMessageTypeDescriptor) (CoreMessageTypeSpec, error) {
	spec := CoreMessageTypeSpec{
		Version:	CoreMessageSpecVersion,
		Types:		normalizeCoreMessageTypeDescriptors(types),
	}
	if err := spec.ValidateFormat(); err != nil {
		return CoreMessageTypeSpec{}, err
	}
	spec.Root = ComputeCoreMessageTypeSpecRoot(spec.Types)
	return spec, spec.Validate()
}

func DefaultCoreMessageTypeSpec() (CoreMessageTypeSpec, error) {
	return BuildCoreMessageTypeSpec(CoreMessageTypeDescriptors())
}

func (s CoreMessageTypeSpec) Normalize() CoreMessageTypeSpec {
	if s.Version == 0 {
		s.Version = CoreMessageSpecVersion
	}
	s.Types = normalizeCoreMessageTypeDescriptors(s.Types)
	s.Root = strings.ToLower(strings.TrimSpace(s.Root))
	return s
}

func (s CoreMessageTypeSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != CoreMessageSpecVersion {
		return fmt.Errorf("core message spec version must be %d", CoreMessageSpecVersion)
	}
	if len(s.Types) == 0 {
		return errors.New("core message spec requires message types")
	}
	seen := make(map[CoreMessageType]struct{}, len(s.Types))
	var previous CoreMessageType
	for i, desc := range s.Types {
		if err := desc.Validate(); err != nil {
			return err
		}
		if _, found := seen[desc.MessageType]; found {
			return fmt.Errorf("duplicate core message type %s", desc.MessageType)
		}
		seen[desc.MessageType] = struct{}{}
		if i > 0 && previous >= desc.MessageType {
			return errors.New("core message types must be sorted canonically")
		}
		previous = desc.MessageType
	}
	if s.Root != "" {
		if err := zonestypes.ValidateHash("core message spec root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s CoreMessageTypeSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("core message spec root is required")
	}
	expected := ComputeCoreMessageTypeSpecRoot(s.Types)
	if s.Root != expected {
		return fmt.Errorf("core message spec root mismatch: expected %s", expected)
	}
	return nil
}

func BuildCoreMessageTypeDescriptor(desc CoreMessageTypeDescriptor) (CoreMessageTypeDescriptor, error) {
	desc = desc.Normalize()
	if desc.DescriptorHash != "" {
		return CoreMessageTypeDescriptor{}, errors.New("core message descriptor hash must be empty before construction")
	}
	if err := desc.ValidateFormat(); err != nil {
		return CoreMessageTypeDescriptor{}, err
	}
	desc.DescriptorHash = ComputeCoreMessageTypeDescriptorHash(desc)
	return desc, desc.Validate()
}

func (d CoreMessageTypeDescriptor) Normalize() CoreMessageTypeDescriptor {
	d.Purpose = strings.Join(strings.Fields(strings.TrimSpace(d.Purpose)), " ")
	d.PrimaryZone = strings.Join(strings.Fields(strings.TrimSpace(d.PrimaryZone)), " ")
	d.DescriptorHash = strings.ToLower(strings.TrimSpace(d.DescriptorHash))
	return d
}

func (d CoreMessageTypeDescriptor) ValidateFormat() error {
	d = d.Normalize()
	if !IsCoreMessageType(d.MessageType) {
		return fmt.Errorf("unknown core message type %q", d.MessageType)
	}
	if d.Purpose == "" {
		return errors.New("core message purpose is required")
	}
	if d.PrimaryZone == "" {
		return errors.New("core message primary zone is required")
	}
	if d.DescriptorHash != "" {
		if err := zonestypes.ValidateHash("core message descriptor hash", d.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (d CoreMessageTypeDescriptor) Validate() error {
	d = d.Normalize()
	if err := d.ValidateFormat(); err != nil {
		return err
	}
	if d.DescriptorHash == "" {
		return errors.New("core message descriptor hash is required")
	}
	expected := ComputeCoreMessageTypeDescriptorHash(d)
	if d.DescriptorHash != expected {
		return fmt.Errorf("core message descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func IsCoreMessageType(messageType CoreMessageType) bool {
	switch messageType {
	case CoreMsgZoneTransfer,
		CoreMsgCrossZoneCall,
		CoreMsgShardCall,
		CoreMsgContractCall,
		CoreMsgIdentityLookup,
		CoreMsgIdentityLookupResult,
		CoreMsgPaymentRoute,
		CoreMsgPaymentSettle,
		CoreMsgPromiseResolve,
		CoreMsgPromiseTimeout,
		CoreMsgProofSubmit,
		CoreMsgBounce:
		return true
	default:
		return false
	}
}

func ComputeCoreMessageTypeDescriptorHash(desc CoreMessageTypeDescriptor) string {
	desc = desc.Normalize()
	return hashParts("aetra-core-message-type-descriptor-v1", string(desc.MessageType), desc.Purpose, desc.PrimaryZone)
}

func ComputeCoreMessageTypeSpecRoot(types []CoreMessageTypeDescriptor) string {
	ordered := normalizeCoreMessageTypeDescriptors(types)
	parts := []string{"aetra-core-message-type-spec-v1", fmt.Sprintf("%020d", CoreMessageSpecVersion)}
	for _, desc := range ordered {
		parts = append(parts, string(desc.MessageType), desc.DescriptorHash)
	}
	return hashParts(parts...)
}

func ValidateCoreMessageTypeSpec() error {
	spec, err := DefaultCoreMessageTypeSpec()
	if err != nil {
		return err
	}
	required := []CoreMessageType{
		CoreMsgZoneTransfer,
		CoreMsgCrossZoneCall,
		CoreMsgShardCall,
		CoreMsgContractCall,
		CoreMsgIdentityLookup,
		CoreMsgIdentityLookupResult,
		CoreMsgPaymentRoute,
		CoreMsgPaymentSettle,
		CoreMsgPromiseResolve,
		CoreMsgPromiseTimeout,
		CoreMsgProofSubmit,
		CoreMsgBounce,
	}
	seen := make(map[CoreMessageType]struct{}, len(spec.Types))
	for _, desc := range spec.Types {
		seen[desc.MessageType] = struct{}{}
	}
	for _, messageType := range required {
		if _, found := seen[messageType]; !found {
			return fmt.Errorf("core message spec missing %s", messageType)
		}
	}
	return nil
}

func coreMessageType(messageType CoreMessageType, purpose string, primaryZone string) CoreMessageTypeDescriptor {
	desc, err := BuildCoreMessageTypeDescriptor(CoreMessageTypeDescriptor{
		MessageType:	messageType,
		Purpose:	purpose,
		PrimaryZone:	primaryZone,
	})
	if err != nil {
		panic(err)
	}
	return desc
}

func normalizeCoreMessageTypeDescriptors(values []CoreMessageTypeDescriptor) []CoreMessageTypeDescriptor {
	out := make([]CoreMessageTypeDescriptor, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeCoreMessageTypeDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].MessageType < out[j].MessageType
	})
	return out
}
