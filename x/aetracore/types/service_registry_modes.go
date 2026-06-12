package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type OnChainServiceRegistryState struct {
	Descriptor		ServiceDescriptor
	DescriptorHash		string
	InterfaceDescriptorHash	string
	PaymentModel		string
	VerificationModel	ServiceVerificationModel
	OwnerAuthorizationHash	string
	ExpiryHeight		uint64
	StateHash		string
}

type HybridServiceRegistryAnchor struct {
	ServiceID		string
	Owner			string
	DescriptorHash		string
	InterfaceHash		string
	ProviderRoot		string
	ExpiryHeight		uint64
	VerificationModel	ServiceVerificationModel
	AnchorHash		string
}

type MeshServiceRegistryAdvertisement struct {
	ServiceID		string
	Owner			string
	ProviderID		string
	Endpoint		string
	DescriptorHash		string
	InterfaceHash		string
	ProviderRoot		string
	GossipTopic		string
	IndexerNodeID		string
	SignatureHash		string
	AnchorProof		ServiceRegistryProof
	AdvertisedHeight	uint64
	ExpiryHeight		uint64
	AdvertisementHash	string
}

type MeshServiceRegistryCacheRecord struct {
	Advertisement	MeshServiceRegistryAdvertisement
	AdvisoryOnly	bool
	LocalReputation	uint64
	CachedHeight	uint64
	ExpiryHeight	uint64
	CacheHash	string
}

type ServiceRegistryModeState struct {
	Mode		ServiceRegistryMode
	OnChainStates	[]OnChainServiceRegistryState
	HybridAnchors	[]HybridServiceRegistryAnchor
	MeshRecords	[]MeshServiceRegistryCacheRecord
	StateRoot	string
	UpdatedHeight	uint64
}

func NewOnChainServiceRegistryState(descriptor ServiceDescriptor, ownerAuthorizationHash string) (OnChainServiceRegistryState, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return OnChainServiceRegistryState{}, err
	}
	if err := ValidateHash("aetracore on-chain registry owner authorization", ownerAuthorizationHash); err != nil {
		return OnChainServiceRegistryState{}, err
	}
	state := OnChainServiceRegistryState{
		Descriptor:			descriptor,
		DescriptorHash:			ComputeServiceDescriptorHash(descriptor),
		InterfaceDescriptorHash:	descriptor.Interface.InterfaceHash,
		PaymentModel:			registryPaymentModel(descriptor),
		VerificationModel:		descriptor.Verification.Model,
		OwnerAuthorizationHash:		strings.ToLower(strings.TrimSpace(ownerAuthorizationHash)),
		ExpiryHeight:			descriptor.ExpiryHeight,
	}
	state.StateHash = ComputeOnChainServiceRegistryStateHash(state)
	return state, state.Validate()
}

func NewHybridServiceRegistryAnchor(descriptor ServiceDescriptor) (HybridServiceRegistryAnchor, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return HybridServiceRegistryAnchor{}, err
	}
	providerRoot := registryProviderSet(descriptor)
	if providerRoot == "" {
		return HybridServiceRegistryAnchor{}, errors.New("aetracore hybrid registry anchor requires provider root")
	}
	anchor := HybridServiceRegistryAnchor{
		ServiceID:		descriptor.ServiceID,
		Owner:			descriptor.Owner,
		DescriptorHash:		ComputeServiceDescriptorHash(descriptor),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		ProviderRoot:		providerRoot,
		ExpiryHeight:		descriptor.ExpiryHeight,
		VerificationModel:	descriptor.Verification.Model,
	}
	anchor.AnchorHash = ComputeHybridServiceRegistryAnchorHash(anchor)
	return anchor, anchor.Validate()
}

func BuildServiceRegistryModeState(mode ServiceRegistryMode, descriptors []ServiceDescriptor, ownerAuthorizations map[string]string, height uint64) (ServiceRegistryModeState, error) {
	if height == 0 {
		return ServiceRegistryModeState{}, errors.New("aetracore service registry mode state height must be positive")
	}
	state := ServiceRegistryModeState{
		Mode:		mode,
		UpdatedHeight:	height,
	}
	switch mode {
	case ServiceRegistryOnChain:
		state.OnChainStates = make([]OnChainServiceRegistryState, 0, len(descriptors))
		for _, descriptor := range descriptors {
			descriptor = CanonicalServiceDescriptor(descriptor)
			authHash := ownerAuthorizations[descriptor.ServiceID]
			if authHash == "" {
				return ServiceRegistryModeState{}, fmt.Errorf("aetracore on-chain registry service %s requires owner authorization", descriptor.ServiceID)
			}
			onChainState, err := NewOnChainServiceRegistryState(descriptor, authHash)
			if err != nil {
				return ServiceRegistryModeState{}, err
			}
			state.OnChainStates = append(state.OnChainStates, onChainState)
		}
		sortOnChainServiceRegistryStates(state.OnChainStates)
	case ServiceRegistryHybrid:
		state.HybridAnchors = make([]HybridServiceRegistryAnchor, 0, len(descriptors))
		for _, descriptor := range descriptors {
			anchor, err := NewHybridServiceRegistryAnchor(descriptor)
			if err != nil {
				return ServiceRegistryModeState{}, err
			}
			state.HybridAnchors = append(state.HybridAnchors, anchor)
		}
		sortHybridServiceRegistryAnchors(state.HybridAnchors)
	case ServiceRegistryMesh:
		return ServiceRegistryModeState{}, errors.New("aetracore distributed registry mesh requires signed advertisements")
	default:
		return ServiceRegistryModeState{}, fmt.Errorf("aetracore service registry mode state does not implement %q", mode)
	}
	state.StateRoot = ComputeServiceRegistryModeStateRoot(state)
	return state, state.Validate()
}

func NewMeshServiceRegistryAdvertisement(descriptor ServiceDescriptor, providerID, endpoint, gossipTopic, signatureHash string, advertisedHeight uint64) (MeshServiceRegistryAdvertisement, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return MeshServiceRegistryAdvertisement{}, err
	}
	if advertisedHeight == 0 {
		return MeshServiceRegistryAdvertisement{}, errors.New("aetracore mesh registry advertisement height must be positive")
	}
	if err := ValidateHash("aetracore mesh registry advertisement signature", signatureHash); err != nil {
		return MeshServiceRegistryAdvertisement{}, err
	}
	providerRoot := registryProviderSet(descriptor)
	advertisement := MeshServiceRegistryAdvertisement{
		ServiceID:		descriptor.ServiceID,
		Owner:			descriptor.Owner,
		ProviderID:		strings.TrimSpace(providerID),
		Endpoint:		strings.TrimSpace(endpoint),
		DescriptorHash:		ComputeServiceDescriptorHash(descriptor),
		InterfaceHash:		descriptor.Interface.InterfaceHash,
		ProviderRoot:		providerRoot,
		GossipTopic:		strings.TrimSpace(gossipTopic),
		SignatureHash:		strings.ToLower(strings.TrimSpace(signatureHash)),
		AdvertisedHeight:	advertisedHeight,
		ExpiryHeight:		descriptor.ExpiryHeight,
	}
	advertisement.AdvertisementHash = ComputeMeshServiceRegistryAdvertisementHash(advertisement)
	return advertisement, advertisement.Validate()
}

func AttachMeshServiceRegistryAnchorProof(advertisement MeshServiceRegistryAdvertisement, proof ServiceRegistryProof) (MeshServiceRegistryAdvertisement, error) {
	advertisement = CanonicalMeshServiceRegistryAdvertisement(advertisement)
	if err := advertisement.Validate(); err != nil {
		return MeshServiceRegistryAdvertisement{}, err
	}
	advertisement.AnchorProof = proof
	advertisement.AdvertisementHash = ComputeMeshServiceRegistryAdvertisementHash(advertisement)
	return advertisement, advertisement.Validate()
}

func NewMeshServiceRegistryCacheRecord(advertisement MeshServiceRegistryAdvertisement, localReputation, cachedHeight uint64) (MeshServiceRegistryCacheRecord, error) {
	advertisement = CanonicalMeshServiceRegistryAdvertisement(advertisement)
	if err := advertisement.Validate(); err != nil {
		return MeshServiceRegistryCacheRecord{}, err
	}
	if cachedHeight == 0 {
		return MeshServiceRegistryCacheRecord{}, errors.New("aetracore mesh registry cache height must be positive")
	}
	if advertisement.ExpiryHeight <= cachedHeight {
		return MeshServiceRegistryCacheRecord{}, errors.New("aetracore mesh registry cached record is expired")
	}
	record := MeshServiceRegistryCacheRecord{
		Advertisement:		advertisement,
		AdvisoryOnly:		true,
		LocalReputation:	localReputation,
		CachedHeight:		cachedHeight,
		ExpiryHeight:		advertisement.ExpiryHeight,
	}
	record.CacheHash = ComputeMeshServiceRegistryCacheRecordHash(record)
	return record, record.Validate()
}

func BuildDistributedRegistryMeshState(records []MeshServiceRegistryCacheRecord, height uint64) (ServiceRegistryModeState, error) {
	if height == 0 {
		return ServiceRegistryModeState{}, errors.New("aetracore distributed registry mesh height must be positive")
	}
	state := ServiceRegistryModeState{
		Mode:		ServiceRegistryMesh,
		MeshRecords:	append([]MeshServiceRegistryCacheRecord(nil), records...),
		UpdatedHeight:	height,
	}
	sortMeshServiceRegistryCacheRecords(state.MeshRecords)
	state.StateRoot = ComputeServiceRegistryModeStateRoot(state)
	return state, state.Validate()
}

func (state OnChainServiceRegistryState) Validate() error {
	state.Descriptor = CanonicalServiceDescriptor(state.Descriptor)
	if err := state.Descriptor.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore on-chain registry descriptor hash", state.DescriptorHash); err != nil {
		return err
	}
	if expected := ComputeServiceDescriptorHash(state.Descriptor); state.DescriptorHash != expected {
		return fmt.Errorf("aetracore on-chain registry descriptor hash mismatch: expected %s", expected)
	}
	if err := ValidateHash("aetracore on-chain registry interface descriptor hash", state.InterfaceDescriptorHash); err != nil {
		return err
	}
	if state.InterfaceDescriptorHash != state.Descriptor.Interface.InterfaceHash {
		return errors.New("aetracore on-chain registry interface descriptor hash mismatch")
	}
	if err := validatePolicyID("aetracore on-chain registry payment model", state.PaymentModel); err != nil {
		return err
	}
	if state.PaymentModel != registryPaymentModel(state.Descriptor) {
		return errors.New("aetracore on-chain registry payment model mismatch")
	}
	if !IsServiceVerificationModel(state.VerificationModel) {
		return fmt.Errorf("unknown aetracore on-chain registry verification model %q", state.VerificationModel)
	}
	if state.VerificationModel != state.Descriptor.Verification.Model {
		return errors.New("aetracore on-chain registry verification model mismatch")
	}
	if err := ValidateHash("aetracore on-chain registry owner authorization", state.OwnerAuthorizationHash); err != nil {
		return err
	}
	if state.ExpiryHeight != state.Descriptor.ExpiryHeight {
		return errors.New("aetracore on-chain registry expiry mismatch")
	}
	if err := ValidateHash("aetracore on-chain registry state hash", state.StateHash); err != nil {
		return err
	}
	if expected := ComputeOnChainServiceRegistryStateHash(state); state.StateHash != expected {
		return fmt.Errorf("aetracore on-chain registry state hash mismatch: expected %s", expected)
	}
	return nil
}

func (anchor HybridServiceRegistryAnchor) Validate() error {
	anchor = CanonicalHybridServiceRegistryAnchor(anchor)
	if err := validatePolicyID("aetracore hybrid registry service id", anchor.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore hybrid registry owner", anchor.Owner); err != nil {
		return err
	}
	if err := ValidateHash("aetracore hybrid registry descriptor hash", anchor.DescriptorHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore hybrid registry interface hash", anchor.InterfaceHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore hybrid registry provider root", anchor.ProviderRoot); err != nil {
		return err
	}
	if !IsServiceVerificationModel(anchor.VerificationModel) {
		return fmt.Errorf("unknown aetracore hybrid registry verification model %q", anchor.VerificationModel)
	}
	if err := ValidateHash("aetracore hybrid registry anchor hash", anchor.AnchorHash); err != nil {
		return err
	}
	if expected := ComputeHybridServiceRegistryAnchorHash(anchor); anchor.AnchorHash != expected {
		return fmt.Errorf("aetracore hybrid registry anchor hash mismatch: expected %s", expected)
	}
	return nil
}

func (state ServiceRegistryModeState) Validate() error {
	if state.UpdatedHeight == 0 {
		return errors.New("aetracore service registry mode state updated height must be positive")
	}
	switch state.Mode {
	case ServiceRegistryOnChain:
		if len(state.HybridAnchors) != 0 {
			return errors.New("aetracore on-chain registry mode state cannot contain hybrid anchors")
		}
		if len(state.MeshRecords) != 0 {
			return errors.New("aetracore on-chain registry mode state cannot contain mesh records")
		}
		if err := validateOnChainServiceRegistryStates(state.OnChainStates); err != nil {
			return err
		}
	case ServiceRegistryHybrid:
		if len(state.OnChainStates) != 0 {
			return errors.New("aetracore hybrid registry mode state cannot contain full on-chain descriptors")
		}
		if len(state.MeshRecords) != 0 {
			return errors.New("aetracore hybrid registry mode state cannot contain mesh records")
		}
		if err := validateHybridServiceRegistryAnchors(state.HybridAnchors); err != nil {
			return err
		}
	case ServiceRegistryMesh:
		if len(state.OnChainStates) != 0 || len(state.HybridAnchors) != 0 {
			return errors.New("aetracore distributed registry mesh state cannot contain consensus registry records")
		}
		if err := validateMeshServiceRegistryCacheRecords(state.MeshRecords); err != nil {
			return err
		}
	default:
		return fmt.Errorf("aetracore service registry mode state does not implement %q", state.Mode)
	}
	if err := ValidateHash("aetracore service registry mode state root", state.StateRoot); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryModeStateRoot(state); state.StateRoot != expected {
		return fmt.Errorf("aetracore service registry mode state root mismatch: expected %s", expected)
	}
	return nil
}

func (state ServiceRegistryModeState) OnChainDescriptorByID(serviceID string) (ServiceDescriptor, ServiceRegistryProof, bool) {
	if state.Mode != ServiceRegistryOnChain {
		return ServiceDescriptor{}, ServiceRegistryProof{}, false
	}
	for _, onChainState := range state.OnChainStates {
		if onChainState.Descriptor.ServiceID == serviceID {
			proof := ServiceRegistryProof{
				ServiceID:	serviceID,
				RegistryMode:	state.Mode,
				RegistryRoot:	state.StateRoot,
				RecordHash:	onChainState.StateHash,
				DescriptorHash:	onChainState.DescriptorHash,
				InterfaceHash:	onChainState.InterfaceDescriptorHash,
				ProofHeight:	state.UpdatedHeight,
			}
			proof.ProofHash = ComputeServiceRegistryProofHash(proof)
			return onChainState.Descriptor, proof, true
		}
	}
	return ServiceDescriptor{}, ServiceRegistryProof{}, false
}

func (state ServiceRegistryModeState) HybridAnchorByID(serviceID string) (HybridServiceRegistryAnchor, ServiceRegistryProof, bool) {
	if state.Mode != ServiceRegistryHybrid {
		return HybridServiceRegistryAnchor{}, ServiceRegistryProof{}, false
	}
	for _, anchor := range state.HybridAnchors {
		if anchor.ServiceID == serviceID {
			proof := ServiceRegistryProof{
				ServiceID:	serviceID,
				RegistryMode:	state.Mode,
				RegistryRoot:	state.StateRoot,
				RecordHash:	anchor.AnchorHash,
				DescriptorHash:	anchor.DescriptorHash,
				InterfaceHash:	anchor.InterfaceHash,
				ProofHeight:	state.UpdatedHeight,
			}
			proof.ProofHash = ComputeServiceRegistryProofHash(proof)
			return anchor, proof, true
		}
	}
	return HybridServiceRegistryAnchor{}, ServiceRegistryProof{}, false
}

func (state ServiceRegistryModeState) MeshLookup(serviceID string) (MeshServiceRegistryCacheRecord, ServiceRegistryProof, bool) {
	if state.Mode != ServiceRegistryMesh {
		return MeshServiceRegistryCacheRecord{}, ServiceRegistryProof{}, false
	}
	for _, record := range state.MeshRecords {
		if record.Advertisement.ServiceID == serviceID {
			proof := ServiceRegistryProof{
				ServiceID:	serviceID,
				RegistryMode:	state.Mode,
				RegistryRoot:	state.StateRoot,
				RecordHash:	record.CacheHash,
				DescriptorHash:	record.Advertisement.DescriptorHash,
				InterfaceHash:	record.Advertisement.InterfaceHash,
				ProofHeight:	state.UpdatedHeight,
			}
			proof.ProofHash = ComputeServiceRegistryProofHash(proof)
			return record, proof, true
		}
	}
	return MeshServiceRegistryCacheRecord{}, ServiceRegistryProof{}, false
}

func CanonicalHybridServiceRegistryAnchor(anchor HybridServiceRegistryAnchor) HybridServiceRegistryAnchor {
	anchor.ServiceID = strings.TrimSpace(anchor.ServiceID)
	anchor.Owner = strings.TrimSpace(anchor.Owner)
	anchor.DescriptorHash = strings.ToLower(strings.TrimSpace(anchor.DescriptorHash))
	anchor.InterfaceHash = strings.ToLower(strings.TrimSpace(anchor.InterfaceHash))
	anchor.ProviderRoot = strings.ToLower(strings.TrimSpace(anchor.ProviderRoot))
	anchor.AnchorHash = strings.ToLower(strings.TrimSpace(anchor.AnchorHash))
	if anchor.AnchorHash == "" {
		anchor.AnchorHash = ComputeHybridServiceRegistryAnchorHash(anchor)
	}
	return anchor
}

func CanonicalMeshServiceRegistryAdvertisement(advertisement MeshServiceRegistryAdvertisement) MeshServiceRegistryAdvertisement {
	advertisement.ServiceID = strings.TrimSpace(advertisement.ServiceID)
	advertisement.Owner = strings.TrimSpace(advertisement.Owner)
	advertisement.ProviderID = strings.TrimSpace(advertisement.ProviderID)
	advertisement.Endpoint = strings.TrimSpace(advertisement.Endpoint)
	advertisement.DescriptorHash = strings.ToLower(strings.TrimSpace(advertisement.DescriptorHash))
	advertisement.InterfaceHash = strings.ToLower(strings.TrimSpace(advertisement.InterfaceHash))
	advertisement.ProviderRoot = strings.ToLower(strings.TrimSpace(advertisement.ProviderRoot))
	advertisement.GossipTopic = strings.TrimSpace(advertisement.GossipTopic)
	advertisement.IndexerNodeID = strings.TrimSpace(advertisement.IndexerNodeID)
	advertisement.SignatureHash = strings.ToLower(strings.TrimSpace(advertisement.SignatureHash))
	advertisement.AdvertisementHash = strings.ToLower(strings.TrimSpace(advertisement.AdvertisementHash))
	if advertisement.AdvertisementHash == "" {
		advertisement.AdvertisementHash = ComputeMeshServiceRegistryAdvertisementHash(advertisement)
	}
	return advertisement
}

func CanonicalMeshServiceRegistryCacheRecord(record MeshServiceRegistryCacheRecord) MeshServiceRegistryCacheRecord {
	record.Advertisement = CanonicalMeshServiceRegistryAdvertisement(record.Advertisement)
	record.CacheHash = strings.ToLower(strings.TrimSpace(record.CacheHash))
	if record.CacheHash == "" {
		record.CacheHash = ComputeMeshServiceRegistryCacheRecordHash(record)
	}
	return record
}

func (advertisement MeshServiceRegistryAdvertisement) HasAnchorProof() bool {
	return strings.TrimSpace(advertisement.AnchorProof.ProofHash) != ""
}

func (advertisement MeshServiceRegistryAdvertisement) Validate() error {
	advertisement = CanonicalMeshServiceRegistryAdvertisement(advertisement)
	if err := validatePolicyID("aetracore mesh registry service id", advertisement.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore mesh registry owner", advertisement.Owner); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mesh registry provider id", advertisement.ProviderID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mesh registry endpoint", advertisement.Endpoint); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mesh registry descriptor hash", advertisement.DescriptorHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mesh registry interface hash", advertisement.InterfaceHash); err != nil {
		return err
	}
	if advertisement.ProviderRoot != "" {
		if err := ValidateHash("aetracore mesh registry provider root", advertisement.ProviderRoot); err != nil {
			return err
		}
	}
	if err := validatePolicyID("aetracore mesh registry gossip topic", advertisement.GossipTopic); err != nil {
		return err
	}
	if advertisement.IndexerNodeID != "" {
		if err := validatePolicyID("aetracore mesh registry indexer node id", advertisement.IndexerNodeID); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore mesh registry advertisement signature", advertisement.SignatureHash); err != nil {
		return err
	}
	if advertisement.AdvertisedHeight == 0 {
		return errors.New("aetracore mesh registry advertisement height must be positive")
	}
	if advertisement.ExpiryHeight == 0 || advertisement.ExpiryHeight <= advertisement.AdvertisedHeight {
		return errors.New("aetracore mesh registry advertisement expiry must be after advertised height")
	}
	if advertisement.HasAnchorProof() {
		if err := advertisement.AnchorProof.Validate(); err != nil {
			return err
		}
		if advertisement.AnchorProof.RegistryMode == ServiceRegistryMesh {
			return errors.New("aetracore mesh registry anchor proof must reference on-chain or hybrid registry state")
		}
		if advertisement.AnchorProof.ServiceID != advertisement.ServiceID ||
			advertisement.AnchorProof.DescriptorHash != advertisement.DescriptorHash ||
			advertisement.AnchorProof.InterfaceHash != advertisement.InterfaceHash {
			return errors.New("aetracore mesh registry anchor proof does not match advertisement")
		}
	}
	if err := ValidateHash("aetracore mesh registry advertisement hash", advertisement.AdvertisementHash); err != nil {
		return err
	}
	if expected := ComputeMeshServiceRegistryAdvertisementHash(advertisement); advertisement.AdvertisementHash != expected {
		return fmt.Errorf("aetracore mesh registry advertisement hash mismatch: expected %s", expected)
	}
	return nil
}

func (record MeshServiceRegistryCacheRecord) Validate() error {
	record = CanonicalMeshServiceRegistryCacheRecord(record)
	if err := record.Advertisement.Validate(); err != nil {
		return err
	}
	if !record.AdvisoryOnly {
		return errors.New("aetracore mesh registry cache records must remain advisory")
	}
	if record.CachedHeight == 0 {
		return errors.New("aetracore mesh registry cache height must be positive")
	}
	if record.ExpiryHeight == 0 {
		return errors.New("aetracore mesh registry cached records require expiry height")
	}
	if record.ExpiryHeight != record.Advertisement.ExpiryHeight {
		return errors.New("aetracore mesh registry cache expiry must match advertisement")
	}
	if record.ExpiryHeight <= record.CachedHeight {
		return errors.New("aetracore mesh registry cached record is expired")
	}
	if err := ValidateHash("aetracore mesh registry cache hash", record.CacheHash); err != nil {
		return err
	}
	if expected := ComputeMeshServiceRegistryCacheRecordHash(record); record.CacheHash != expected {
		return fmt.Errorf("aetracore mesh registry cache hash mismatch: expected %s", expected)
	}
	return nil
}

func VerifyMeshServiceRegistryAdvertisementDescriptor(advertisement MeshServiceRegistryAdvertisement, descriptor ServiceDescriptor) error {
	advertisement = CanonicalMeshServiceRegistryAdvertisement(advertisement)
	if err := advertisement.Validate(); err != nil {
		return err
	}
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if advertisement.ServiceID != descriptor.ServiceID || advertisement.Owner != descriptor.Owner {
		return errors.New("aetracore mesh registry advertisement identity mismatch")
	}
	if advertisement.DescriptorHash != ComputeServiceDescriptorHash(descriptor) {
		return errors.New("aetracore mesh registry advertisement descriptor hash mismatch")
	}
	if advertisement.InterfaceHash != descriptor.Interface.InterfaceHash {
		return errors.New("aetracore mesh registry advertisement interface hash mismatch")
	}
	return nil
}

func ComputeServiceOwnerAuthorizationHash(serviceID, owner string, height uint64) string {
	return hashParts(
		"aetra-aek-service-owner-authorization-v1",
		strings.TrimSpace(serviceID),
		strings.TrimSpace(owner),
		fmt.Sprint(height),
	)
}

func ComputeMeshServiceRegistryAdvertisementSignatureHash(serviceID, owner, providerID, descriptorHash string, advertisedHeight uint64) string {
	return hashParts(
		"aetra-aek-mesh-service-registry-advertisement-signature-v1",
		strings.TrimSpace(serviceID),
		strings.TrimSpace(owner),
		strings.TrimSpace(providerID),
		strings.ToLower(strings.TrimSpace(descriptorHash)),
		fmt.Sprint(advertisedHeight),
	)
}

func ComputeOnChainServiceRegistryStateHash(state OnChainServiceRegistryState) string {
	return hashParts(
		"aetra-aek-on-chain-service-registry-state-v1",
		state.Descriptor.ServiceID,
		state.DescriptorHash,
		state.InterfaceDescriptorHash,
		state.PaymentModel,
		string(state.VerificationModel),
		state.OwnerAuthorizationHash,
		fmt.Sprint(state.ExpiryHeight),
	)
}

func ComputeHybridServiceRegistryAnchorHash(anchor HybridServiceRegistryAnchor) string {
	anchor.AnchorHash = ""
	return hashParts(
		"aetra-aek-hybrid-service-registry-anchor-v1",
		anchor.ServiceID,
		anchor.Owner,
		anchor.DescriptorHash,
		anchor.InterfaceHash,
		anchor.ProviderRoot,
		fmt.Sprint(anchor.ExpiryHeight),
		string(anchor.VerificationModel),
	)
}

func ComputeMeshServiceRegistryAdvertisementHash(advertisement MeshServiceRegistryAdvertisement) string {
	advertisement.AdvertisementHash = ""
	return hashParts(
		"aetra-aek-mesh-service-registry-advertisement-v1",
		advertisement.ServiceID,
		advertisement.Owner,
		advertisement.ProviderID,
		advertisement.Endpoint,
		advertisement.DescriptorHash,
		advertisement.InterfaceHash,
		advertisement.ProviderRoot,
		advertisement.GossipTopic,
		advertisement.IndexerNodeID,
		advertisement.SignatureHash,
		advertisement.AnchorProof.ProofHash,
		fmt.Sprint(advertisement.AdvertisedHeight),
		fmt.Sprint(advertisement.ExpiryHeight),
	)
}

func ComputeMeshServiceRegistryCacheRecordHash(record MeshServiceRegistryCacheRecord) string {
	record.CacheHash = ""
	return hashParts(
		"aetra-aek-mesh-service-registry-cache-record-v1",
		record.Advertisement.AdvertisementHash,
		fmt.Sprint(record.AdvisoryOnly),
		fmt.Sprint(record.LocalReputation),
		fmt.Sprint(record.CachedHeight),
		fmt.Sprint(record.ExpiryHeight),
	)
}

func ComputeServiceRegistryModeStateRoot(state ServiceRegistryModeState) string {
	onChainStates := append([]OnChainServiceRegistryState(nil), state.OnChainStates...)
	hybridAnchors := append([]HybridServiceRegistryAnchor(nil), state.HybridAnchors...)
	meshRecords := append([]MeshServiceRegistryCacheRecord(nil), state.MeshRecords...)
	sortOnChainServiceRegistryStates(onChainStates)
	sortHybridServiceRegistryAnchors(hybridAnchors)
	sortMeshServiceRegistryCacheRecords(meshRecords)
	parts := []string{
		"aetra-aek-service-registry-mode-state-root-v1",
		string(state.Mode),
		fmt.Sprint(state.UpdatedHeight),
		fmt.Sprint(len(onChainStates)),
		fmt.Sprint(len(hybridAnchors)),
	}
	for _, onChainState := range onChainStates {
		parts = append(parts, onChainState.StateHash)
	}
	for _, anchor := range hybridAnchors {
		parts = append(parts, anchor.AnchorHash)
	}
	if state.Mode == ServiceRegistryMesh || len(meshRecords) != 0 {
		parts = append(parts, fmt.Sprint(len(meshRecords)))
		for _, record := range meshRecords {
			parts = append(parts, record.CacheHash)
		}
	}
	return hashParts(parts...)
}

func validateOnChainServiceRegistryStates(states []OnChainServiceRegistryState) error {
	var previous string
	seen := make(map[string]struct{}, len(states))
	for _, state := range states {
		if err := state.Validate(); err != nil {
			return err
		}
		serviceID := state.Descriptor.ServiceID
		if _, found := seen[serviceID]; found {
			return fmt.Errorf("duplicate aetracore on-chain registry state %s", serviceID)
		}
		seen[serviceID] = struct{}{}
		if previous != "" && previous >= serviceID {
			return errors.New("aetracore on-chain registry states must be sorted canonically")
		}
		previous = serviceID
	}
	return nil
}

func validateHybridServiceRegistryAnchors(anchors []HybridServiceRegistryAnchor) error {
	var previous string
	seen := make(map[string]struct{}, len(anchors))
	for _, anchor := range anchors {
		anchor = CanonicalHybridServiceRegistryAnchor(anchor)
		if err := anchor.Validate(); err != nil {
			return err
		}
		if _, found := seen[anchor.ServiceID]; found {
			return fmt.Errorf("duplicate aetracore hybrid registry anchor %s", anchor.ServiceID)
		}
		seen[anchor.ServiceID] = struct{}{}
		if previous != "" && previous >= anchor.ServiceID {
			return errors.New("aetracore hybrid registry anchors must be sorted canonically")
		}
		previous = anchor.ServiceID
	}
	return nil
}

func validateMeshServiceRegistryCacheRecords(records []MeshServiceRegistryCacheRecord) error {
	var previous string
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		record = CanonicalMeshServiceRegistryCacheRecord(record)
		if err := record.Validate(); err != nil {
			return err
		}
		key := record.Advertisement.ServiceID + "/" + record.Advertisement.ProviderID
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate aetracore mesh registry cache record %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("aetracore mesh registry cache records must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func sortOnChainServiceRegistryStates(states []OnChainServiceRegistryState) {
	sort.SliceStable(states, func(i, j int) bool {
		return states[i].Descriptor.ServiceID < states[j].Descriptor.ServiceID
	})
}

func sortHybridServiceRegistryAnchors(anchors []HybridServiceRegistryAnchor) {
	sort.SliceStable(anchors, func(i, j int) bool { return anchors[i].ServiceID < anchors[j].ServiceID })
}

func sortMeshServiceRegistryCacheRecords(records []MeshServiceRegistryCacheRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i].Advertisement.ServiceID + "/" + records[i].Advertisement.ProviderID
		right := records[j].Advertisement.ServiceID + "/" + records[j].Advertisement.ProviderID
		return left < right
	})
}
