package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type CommitmentProofType string

const (
	ZoneProofType		CommitmentProofType	= "zone"
	ServiceProofType	CommitmentProofType	= "service"
	IdentityProofType	CommitmentProofType	= "identity"
	StorageProofType	CommitmentProofType	= "storage"
	MessageProofType	CommitmentProofType	= "message"
	ReceiptProofType	CommitmentProofType	= "receipt"
	PaymentProofType	CommitmentProofType	= "payment"
	ContractProofType	CommitmentProofType	= "contract"
	RoutingProofType	CommitmentProofType	= "routing"
	NonExistenceProofType	CommitmentProofType	= "non_existence"
)

type RootContribution struct {
	RootType		RootType
	ID			string
	RootHash		string
	ContributionHash	string
}

type ZoneRootAggregation struct {
	Height		uint64
	ZoneID		ZoneID
	ModuleRoots	[]RootContribution
	ZoneRoot	string
}

type AEKRootAggregation struct {
	Height		uint64
	ZoneRoots	[]RootContribution
	GlobalRoots	[]RootContribution
	AggregateRoot	string
}

type StateCommitmentProof struct {
	ProofType	CommitmentProofType
	Height		uint64
	RootType	RootType
	SubjectID	string
	KeyHash		string
	ValueHash	string
	RootHash	string
	Path		[]string
	Exists		bool
	ProofHash	string
}

type ZoneProof = StateCommitmentProof
type ServiceProof = StateCommitmentProof
type IdentityProof = StateCommitmentProof
type StorageProof = StateCommitmentProof
type MessageProof = StateCommitmentProof
type ReceiptProof = StateCommitmentProof
type PaymentProof = StateCommitmentProof
type ContractProof = StateCommitmentProof
type RoutingProof = StateCommitmentProof
type NonExistenceProof = StateCommitmentProof

func NewRootContribution(rootType RootType, id string, rootHash string) (RootContribution, error) {
	contribution := RootContribution{
		RootType:	RootType(strings.TrimSpace(string(rootType))),
		ID:		strings.TrimSpace(id),
		RootHash:	strings.ToLower(strings.TrimSpace(rootHash)),
	}
	if contribution.RootHash == "" {
		contribution.RootHash = DeterministicEmptyRootCommitment(contribution.RootType, contribution.ID)
	}
	if err := contribution.ValidateFormat(); err != nil {
		return RootContribution{}, err
	}
	contribution.ContributionHash = ComputeRootContributionHash(contribution)
	return contribution, nil
}

func BuildZoneRootAggregation(height uint64, zoneID ZoneID, moduleRoots []RootContribution) (ZoneRootAggregation, error) {
	aggregation := ZoneRootAggregation{
		Height:		height,
		ZoneID:		zoneID,
		ModuleRoots:	normalizeRootContributions(moduleRoots),
	}
	if err := aggregation.ValidateFormat(); err != nil {
		return ZoneRootAggregation{}, err
	}
	aggregation.ZoneRoot = ComputeZoneRootAggregationHash(aggregation)
	return aggregation, aggregation.Validate()
}

func BuildAEKRootAggregation(height uint64, zoneRoots []RootContribution, globalRoots []RootContribution) (AEKRootAggregation, error) {
	aggregation := AEKRootAggregation{
		Height:		height,
		ZoneRoots:	normalizeRootContributions(zoneRoots),
		GlobalRoots:	normalizeRootContributions(globalRoots),
	}
	if err := aggregation.ValidateFormat(); err != nil {
		return AEKRootAggregation{}, err
	}
	aggregation.AggregateRoot = ComputeAEKRootAggregationHash(aggregation)
	return aggregation, aggregation.Validate()
}

func NewStateCommitmentProof(proofType CommitmentProofType, height uint64, rootType RootType, subjectID string, keyHash string, valueHash string, rootHash string, path []string, exists bool) (StateCommitmentProof, error) {
	proof := StateCommitmentProof{
		ProofType:	CommitmentProofType(strings.TrimSpace(string(proofType))),
		Height:		height,
		RootType:	RootType(strings.TrimSpace(string(rootType))),
		SubjectID:	strings.TrimSpace(subjectID),
		KeyHash:	strings.ToLower(strings.TrimSpace(keyHash)),
		ValueHash:	strings.ToLower(strings.TrimSpace(valueHash)),
		RootHash:	strings.ToLower(strings.TrimSpace(rootHash)),
		Path:		normalizeProofPath(path),
		Exists:		exists,
	}
	if !proof.Exists && proof.ValueHash == "" {
		proof.ValueHash = DeterministicEmptyRootCommitment(proof.RootType, proof.SubjectID)
	}
	if err := proof.ValidateFormat(); err != nil {
		return StateCommitmentProof{}, err
	}
	proof.ProofHash = ComputeStateCommitmentProofHash(proof)
	return proof, proof.Validate()
}

func NewNonExistenceProof(height uint64, rootType RootType, subjectID string, keyHash string, rootHash string, path []string) (NonExistenceProof, error) {
	return NewStateCommitmentProof(NonExistenceProofType, height, rootType, subjectID, keyHash, "", rootHash, path, false)
}

func (c RootContribution) ValidateFormat() error {
	c = normalizeRootContribution(c)
	if err := validatePolicyID("aetracore root contribution type", string(c.RootType)); err != nil {
		return err
	}
	if err := validateToken("aetracore root contribution id", c.ID, MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore root contribution hash", c.RootHash); err != nil {
		return err
	}
	if c.ContributionHash != "" {
		if err := ValidateHash("aetracore root contribution commitment", c.ContributionHash); err != nil {
			return err
		}
	}
	return nil
}

func (c RootContribution) Validate() error {
	c = normalizeRootContribution(c)
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	if c.ContributionHash != ComputeRootContributionHash(c) {
		return errors.New("aetracore root contribution commitment mismatch")
	}
	return nil
}

func (a ZoneRootAggregation) ValidateFormat() error {
	a = normalizeZoneRootAggregation(a)
	if a.Height == 0 {
		return errors.New("aetracore zone root aggregation height must be positive")
	}
	if err := ValidateZoneID(a.ZoneID); err != nil {
		return err
	}
	return validateRootContributionSet("aetracore zone module roots", a.ModuleRoots)
}

func (a ZoneRootAggregation) Validate() error {
	a = normalizeZoneRootAggregation(a)
	if err := a.ValidateFormat(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore zone aggregate root", a.ZoneRoot); err != nil {
		return err
	}
	if a.ZoneRoot != ComputeZoneRootAggregationHash(a) {
		return errors.New("aetracore zone aggregate root mismatch")
	}
	return nil
}

func (a AEKRootAggregation) ValidateFormat() error {
	a = normalizeAEKRootAggregation(a)
	if a.Height == 0 {
		return errors.New("aetracore AEK root aggregation height must be positive")
	}
	if err := validateRootContributionSet("aetracore AEK zone roots", a.ZoneRoots); err != nil {
		return err
	}
	return validateRootContributionSet("aetracore AEK global roots", a.GlobalRoots)
}

func (a AEKRootAggregation) Validate() error {
	a = normalizeAEKRootAggregation(a)
	if err := a.ValidateFormat(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore AEK aggregate root", a.AggregateRoot); err != nil {
		return err
	}
	if a.AggregateRoot != ComputeAEKRootAggregationHash(a) {
		return errors.New("aetracore AEK aggregate root mismatch")
	}
	return nil
}

func (p StateCommitmentProof) ValidateFormat() error {
	p = normalizeStateCommitmentProof(p)
	if !IsCommitmentProofType(p.ProofType) {
		return fmt.Errorf("unknown aetracore proof type %q", p.ProofType)
	}
	if p.Height == 0 {
		return errors.New("aetracore proof height must be positive")
	}
	if err := validatePolicyID("aetracore proof root type", string(p.RootType)); err != nil {
		return err
	}
	if err := validateToken("aetracore proof subject id", p.SubjectID, MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore proof key hash", p.KeyHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore proof value hash", p.ValueHash); err != nil {
		return err
	}
	if err := ValidateHash("aetracore proof root hash", p.RootHash); err != nil {
		return err
	}
	for _, item := range p.Path {
		if err := ValidateHash("aetracore proof path item", item); err != nil {
			return err
		}
	}
	if p.ProofType == NonExistenceProofType && p.Exists {
		return errors.New("aetracore non-existence proof must not claim existence")
	}
	if p.ProofType != NonExistenceProofType && !p.Exists {
		return errors.New("aetracore existence proof type must claim existence")
	}
	if p.ProofHash != "" {
		if err := ValidateHash("aetracore proof hash", p.ProofHash); err != nil {
			return err
		}
	}
	return nil
}

func (p StateCommitmentProof) Validate() error {
	p = normalizeStateCommitmentProof(p)
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	if p.ProofHash != ComputeStateCommitmentProofHash(p) {
		return errors.New("aetracore proof hash mismatch")
	}
	return nil
}

func ComputeRootContributionHash(c RootContribution) string {
	c = normalizeRootContribution(c)
	return hashRoot("aetra-aek-root-contribution-v1", func(w byteWriter) {
		writePart(w, string(c.RootType))
		writePart(w, c.ID)
		writePart(w, c.RootHash)
	})
}

func ComputeZoneRootAggregationHash(a ZoneRootAggregation) string {
	a = normalizeZoneRootAggregation(a)
	return hashRoot("aetra-aek-zone-root-aggregation-v1", func(w byteWriter) {
		writeUint64(w, a.Height)
		writePart(w, string(a.ZoneID))
		writeUint64(w, uint64(len(a.ModuleRoots)))
		for _, contribution := range a.ModuleRoots {
			writePart(w, contribution.ContributionHash)
		}
	})
}

func ComputeAEKRootAggregationHash(a AEKRootAggregation) string {
	a = normalizeAEKRootAggregation(a)
	return hashRoot("aetra-aek-root-aggregation-v1", func(w byteWriter) {
		writeUint64(w, a.Height)
		writeUint64(w, uint64(len(a.ZoneRoots)))
		for _, contribution := range a.ZoneRoots {
			writePart(w, contribution.ContributionHash)
		}
		writeUint64(w, uint64(len(a.GlobalRoots)))
		for _, contribution := range a.GlobalRoots {
			writePart(w, contribution.ContributionHash)
		}
	})
}

func ComputeStateCommitmentProofHash(p StateCommitmentProof) string {
	p = normalizeStateCommitmentProof(p)
	return hashRoot("aetra-aek-state-commitment-proof-v1", func(w byteWriter) {
		writePart(w, string(p.ProofType))
		writeUint64(w, p.Height)
		writePart(w, string(p.RootType))
		writePart(w, p.SubjectID)
		writePart(w, p.KeyHash)
		writePart(w, p.ValueHash)
		writePart(w, p.RootHash)
		writeUint64(w, uint64(len(p.Path)))
		for _, item := range p.Path {
			writePart(w, item)
		}
		if p.Exists {
			writePart(w, "exists")
		} else {
			writePart(w, "absent")
		}
	})
}

func DeterministicEmptyRootCommitment(rootType RootType, id string) string {
	return hashRoot("aetra-aek-empty-root-commitment-v1", func(w byteWriter) {
		writePart(w, strings.TrimSpace(string(rootType)))
		writePart(w, strings.TrimSpace(id))
	})
}

func IsCommitmentProofType(proofType CommitmentProofType) bool {
	switch proofType {
	case ZoneProofType, ServiceProofType, IdentityProofType, StorageProofType, MessageProofType, ReceiptProofType, PaymentProofType, ContractProofType, RoutingProofType, NonExistenceProofType:
		return true
	default:
		return false
	}
}

func normalizeRootContribution(c RootContribution) RootContribution {
	c.RootType = RootType(strings.TrimSpace(string(c.RootType)))
	c.ID = strings.TrimSpace(c.ID)
	c.RootHash = strings.ToLower(strings.TrimSpace(c.RootHash))
	c.ContributionHash = strings.ToLower(strings.TrimSpace(c.ContributionHash))
	return c
}

func normalizeRootContributions(contributions []RootContribution) []RootContribution {
	out := make([]RootContribution, len(contributions))
	for i, contribution := range contributions {
		contribution = normalizeRootContribution(contribution)
		if contribution.RootHash == "" {
			contribution.RootHash = DeterministicEmptyRootCommitment(contribution.RootType, contribution.ID)
		}
		if contribution.ContributionHash == "" {
			contribution.ContributionHash = ComputeRootContributionHash(contribution)
		}
		out[i] = contribution
	}
	sort.SliceStable(out, func(i, j int) bool {
		return rootContributionKey(out[i]) < rootContributionKey(out[j])
	})
	return out
}

func normalizeZoneRootAggregation(a ZoneRootAggregation) ZoneRootAggregation {
	a.ModuleRoots = normalizeRootContributions(a.ModuleRoots)
	a.ZoneRoot = strings.ToLower(strings.TrimSpace(a.ZoneRoot))
	return a
}

func normalizeAEKRootAggregation(a AEKRootAggregation) AEKRootAggregation {
	a.ZoneRoots = normalizeRootContributions(a.ZoneRoots)
	a.GlobalRoots = normalizeRootContributions(a.GlobalRoots)
	a.AggregateRoot = strings.ToLower(strings.TrimSpace(a.AggregateRoot))
	return a
}

func normalizeStateCommitmentProof(p StateCommitmentProof) StateCommitmentProof {
	p.ProofType = CommitmentProofType(strings.TrimSpace(string(p.ProofType)))
	p.RootType = RootType(strings.TrimSpace(string(p.RootType)))
	p.SubjectID = strings.TrimSpace(p.SubjectID)
	p.KeyHash = strings.ToLower(strings.TrimSpace(p.KeyHash))
	p.ValueHash = strings.ToLower(strings.TrimSpace(p.ValueHash))
	p.RootHash = strings.ToLower(strings.TrimSpace(p.RootHash))
	p.Path = normalizeProofPath(p.Path)
	p.ProofHash = strings.ToLower(strings.TrimSpace(p.ProofHash))
	return p
}

func normalizeProofPath(path []string) []string {
	out := make([]string, len(path))
	for i, item := range path {
		out[i] = strings.ToLower(strings.TrimSpace(item))
	}
	return out
}

func validateRootContributionSet(field string, contributions []RootContribution) error {
	ordered := normalizeRootContributions(contributions)
	seen := make(map[string]struct{}, len(ordered))
	var previous string
	for i, contribution := range ordered {
		if err := contribution.Validate(); err != nil {
			return err
		}
		key := rootContributionKey(contribution)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate %s %s", field, key)
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return fmt.Errorf("%s must be ordered lexicographically by root type and id", field)
		}
		previous = key
	}
	return nil
}

func rootContributionKey(c RootContribution) string {
	c = normalizeRootContribution(c)
	return string(c.RootType) + "/" + c.ID
}
