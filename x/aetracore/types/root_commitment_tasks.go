package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	RootEncodingVersion	= "aetra-aek-root-encoding-v1"
	CanonicalEmptyRootLabel	= "canonical-empty-root"
)

type RootEncodingDescriptor struct {
	Version		string
	EmptyRoot	string
	EncodingHash	string
}

type ProofRegistryEntry struct {
	ProofType	CommitmentProofType
	RootType	RootType
	RootID		string
	RootHash	string
	Enabled		bool
	RegistryHash	string
}

type ProofRegistry struct {
	Height		uint64
	Entries		[]ProofRegistryEntry
	RegistryRoot	string
}

type RootQuery struct {
	Height		uint64
	RootType	RootType
	RootID		string
}

type RootQueryResponse struct {
	QueryHash	string
	Found		bool
	Root		RootContribution
	ProofRoot	ProofRoot
}

type ProofVerificationRequest struct {
	ExpectedRoot	string
	Registry	ProofRegistry
	Proof		StateCommitmentProof
}

type ProofVerificationResult struct {
	Verified		bool
	ProofHash		string
	RegistryRoot		string
	VerifiedRoot		string
	VerificationHash	string
}

func DefaultRootEncodingDescriptor() RootEncodingDescriptor {
	descriptor := RootEncodingDescriptor{
		Version:	RootEncodingVersion,
		EmptyRoot:	CanonicalEmptyRootValue(),
	}
	descriptor.EncodingHash = ComputeRootEncodingDescriptorHash(descriptor)
	return descriptor
}

func CanonicalEmptyRootValue() string {
	return DeterministicEmptyRootCommitment(RootType(CanonicalEmptyRootLabel), "global")
}

func NewProofRegistryEntry(proofType CommitmentProofType, rootType RootType, rootID string, rootHash string, enabled bool) (ProofRegistryEntry, error) {
	entry := ProofRegistryEntry{
		ProofType:	CommitmentProofType(strings.TrimSpace(string(proofType))),
		RootType:	RootType(strings.TrimSpace(string(rootType))),
		RootID:		strings.TrimSpace(rootID),
		RootHash:	strings.ToLower(strings.TrimSpace(rootHash)),
		Enabled:	enabled,
	}
	if entry.RootHash == "" {
		entry.RootHash = DeterministicEmptyRootCommitment(entry.RootType, entry.RootID)
	}
	if err := entry.ValidateFormat(); err != nil {
		return ProofRegistryEntry{}, err
	}
	entry.RegistryHash = ComputeProofRegistryEntryHash(entry)
	return entry, entry.Validate()
}

func BuildProofRegistry(height uint64, entries []ProofRegistryEntry) (ProofRegistry, error) {
	registry := ProofRegistry{
		Height:		height,
		Entries:	normalizeProofRegistryEntries(entries),
	}
	if err := registry.ValidateFormat(); err != nil {
		return ProofRegistry{}, err
	}
	registry.RegistryRoot = ComputeProofRegistryRoot(registry)
	return registry, registry.Validate()
}

func BuildProofRegistryFromGlobalRoot(root GlobalStateRoot) (ProofRegistry, error) {
	if err := root.ValidateHash(); err != nil {
		return ProofRegistry{}, err
	}
	entries := []ProofRegistryEntry{}
	for _, contribution := range RootContributionsFromGlobalRoot(root) {
		proofType := ProofTypeForRootType(contribution.RootType)
		entry, err := NewProofRegistryEntry(proofType, contribution.RootType, contribution.ID, contribution.RootHash, true)
		if err != nil {
			return ProofRegistry{}, err
		}
		entries = append(entries, entry)
	}
	return BuildProofRegistry(root.Height, entries)
}

func QueryCommittedRoot(state CoreState, query RootQuery) (RootQueryResponse, error) {
	query = normalizeRootQuery(query)
	if err := query.Validate(); err != nil {
		return RootQueryResponse{}, err
	}
	if err := state.Validate(); err != nil {
		return RootQueryResponse{}, err
	}
	response := RootQueryResponse{QueryHash: ComputeRootQueryHash(query)}
	root, found := state.GlobalRootByHeight(query.Height)
	if !found {
		return response, nil
	}
	for _, contribution := range RootContributionsFromGlobalRoot(root) {
		if contribution.RootType == query.RootType && contribution.ID == query.RootID {
			response.Found = true
			response.Root = contribution
			response.ProofRoot = ProofRoot{
				Height:		query.Height,
				RootType:	query.RootType,
				RootHash:	contribution.RootHash,
				Source:		"aetracore.root_query",
			}
			return response, response.Validate()
		}
	}
	return response, nil
}

func QueryStateProof(state CoreState, registry ProofRegistry, proof StateCommitmentProof) (ProofVerificationResult, error) {
	if err := state.Validate(); err != nil {
		return ProofVerificationResult{}, err
	}
	if err := registry.Validate(); err != nil {
		return ProofVerificationResult{}, err
	}
	rootResponse, err := QueryCommittedRoot(state, RootQuery{Height: proof.Height, RootType: proof.RootType, RootID: "global"})
	if err != nil {
		return ProofVerificationResult{}, err
	}
	if !rootResponse.Found {
		return ProofVerificationResult{}, errors.New("aetracore proof query root not found")
	}
	return VerifyStateCommitmentProof(ProofVerificationRequest{
		ExpectedRoot:	rootResponse.Root.RootHash,
		Registry:	registry,
		Proof:		proof,
	})
}

func VerifyStateCommitmentProof(req ProofVerificationRequest) (ProofVerificationResult, error) {
	req.ExpectedRoot = strings.ToLower(strings.TrimSpace(req.ExpectedRoot))
	if err := ValidateHash("aetracore proof verification expected root", req.ExpectedRoot); err != nil {
		return ProofVerificationResult{}, err
	}
	if err := req.Registry.Validate(); err != nil {
		return ProofVerificationResult{}, err
	}
	proof := normalizeStateCommitmentProof(req.Proof)
	if err := proof.Validate(); err != nil {
		return ProofVerificationResult{}, err
	}
	if proof.RootHash != req.ExpectedRoot {
		return ProofVerificationResult{}, errors.New("aetracore proof verification root mismatch")
	}
	entry, found := req.Registry.EntryFor(proof.ProofType, proof.RootType, "global")
	if !found {
		return ProofVerificationResult{}, errors.New("aetracore proof verification registry entry not found")
	}
	if !entry.Enabled {
		return ProofVerificationResult{}, errors.New("aetracore proof verification registry entry disabled")
	}
	if entry.RootHash != proof.RootHash {
		return ProofVerificationResult{}, errors.New("aetracore proof verification registry root mismatch")
	}
	result := ProofVerificationResult{
		Verified:	true,
		ProofHash:	proof.ProofHash,
		RegistryRoot:	req.Registry.RegistryRoot,
		VerifiedRoot:	proof.RootHash,
	}
	result.VerificationHash = ComputeProofVerificationResultHash(result)
	return result, nil
}

func ValidateExportImportRootChecks(state CoreState, manifest ExportManifest) error {
	if err := state.Validate(); err != nil {
		return err
	}
	if err := manifest.ValidateHash(); err != nil {
		return err
	}
	root, found := state.GlobalRootByHeight(manifest.Height)
	if !found {
		return fmt.Errorf("aetracore export/import root check missing global root at height %d", manifest.Height)
	}
	if manifest.GlobalRoot != root.GlobalRoot ||
		manifest.ZonesRoot != root.ZonesRoot ||
		manifest.ServicesRoot != root.ServicesRoot ||
		manifest.IdentityRoot != root.IdentityRoot ||
		manifest.StorageRoot != root.StorageRoot ||
		manifest.MessageRoot != root.MessageRoot ||
		manifest.ReceiptsRoot != root.ReceiptsRoot ||
		manifest.RoutingRoot != root.RoutingRoot ||
		manifest.PaymentsRoot != root.PaymentsRoot ||
		manifest.ContractsRoot != root.ContractsRoot ||
		manifest.VMRoot != root.VMRoot {
		return errors.New("aetracore export/import root set mismatch")
	}
	return nil
}

func RootContributionsFromGlobalRoot(root GlobalStateRoot) []RootContribution {
	entries := []struct {
		rootType	RootType
		id		string
		rootHash	string
	}{
		{RootType("contracts"), "global", root.ContractsRoot},
		{RootType("identity"), "global", root.IdentityRoot},
		{RootType("message"), "global", root.MessageRoot},
		{RootType("payments"), "global", root.PaymentsRoot},
		{RootType("receipts"), "global", root.ReceiptsRoot},
		{RootType("routing"), "global", root.RoutingRoot},
		{RootType("services"), "global", root.ServicesRoot},
		{RootType("storage"), "global", root.StorageRoot},
		{RootType("zones"), "global", root.ZonesRoot},
	}
	out := make([]RootContribution, 0, len(entries))
	for _, entry := range entries {
		contribution, err := NewRootContribution(entry.rootType, entry.id, entry.rootHash)
		if err == nil {
			out = append(out, contribution)
		}
	}
	return normalizeRootContributions(out)
}

func ProofTypeForRootType(rootType RootType) CommitmentProofType {
	switch strings.TrimSpace(string(rootType)) {
	case "zones", "zone", string(ZoneCommitmentsRoot):
		return ZoneProofType
	case "services":
		return ServiceProofType
	case "identity":
		return IdentityProofType
	case "storage":
		return StorageProofType
	case "message", string(MessageProofRootType):
		return MessageProofType
	case "receipt", string(ReceiptProofRootType):
		return ReceiptProofType
	case "payments":
		return PaymentProofType
	case "contracts":
		return ContractProofType
	case "routing", string(RoutingTableRootType):
		return RoutingProofType
	default:
		return NonExistenceProofType
	}
}

func (d RootEncodingDescriptor) Validate() error {
	d.Version = strings.TrimSpace(d.Version)
	d.EmptyRoot = strings.ToLower(strings.TrimSpace(d.EmptyRoot))
	d.EncodingHash = strings.ToLower(strings.TrimSpace(d.EncodingHash))
	if d.Version != RootEncodingVersion {
		return errors.New("aetracore root encoding version mismatch")
	}
	if err := ValidateHash("aetracore root encoding empty root", d.EmptyRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore root encoding hash", d.EncodingHash); err != nil {
		return err
	}
	if d.EncodingHash != ComputeRootEncodingDescriptorHash(d) {
		return errors.New("aetracore root encoding hash mismatch")
	}
	return nil
}

func (entry ProofRegistryEntry) ValidateFormat() error {
	entry = normalizeProofRegistryEntry(entry)
	if !IsCommitmentProofType(entry.ProofType) {
		return fmt.Errorf("unknown aetracore proof registry proof type %q", entry.ProofType)
	}
	if err := validatePolicyID("aetracore proof registry root type", string(entry.RootType)); err != nil {
		return err
	}
	if err := validateToken("aetracore proof registry root id", entry.RootID, MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore proof registry root hash", entry.RootHash); err != nil {
		return err
	}
	if entry.RegistryHash != "" {
		if err := ValidateHash("aetracore proof registry entry hash", entry.RegistryHash); err != nil {
			return err
		}
	}
	return nil
}

func (entry ProofRegistryEntry) Validate() error {
	entry = normalizeProofRegistryEntry(entry)
	if err := entry.ValidateFormat(); err != nil {
		return err
	}
	if entry.RegistryHash != ComputeProofRegistryEntryHash(entry) {
		return errors.New("aetracore proof registry entry hash mismatch")
	}
	return nil
}

func (registry ProofRegistry) ValidateFormat() error {
	registry = normalizeProofRegistry(registry)
	if registry.Height == 0 {
		return errors.New("aetracore proof registry height must be positive")
	}
	seen := make(map[string]struct{}, len(registry.Entries))
	var previous string
	for i, entry := range registry.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		key := proofRegistryEntryKey(entry)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate aetracore proof registry entry %s", key)
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("aetracore proof registry entries must be ordered canonically")
		}
		previous = key
	}
	if registry.RegistryRoot != "" {
		if err := ValidateHash("aetracore proof registry root", registry.RegistryRoot); err != nil {
			return err
		}
	}
	return nil
}

func (registry ProofRegistry) Validate() error {
	registry = normalizeProofRegistry(registry)
	if err := registry.ValidateFormat(); err != nil {
		return err
	}
	if registry.RegistryRoot != ComputeProofRegistryRoot(registry) {
		return errors.New("aetracore proof registry root mismatch")
	}
	return nil
}

func (registry ProofRegistry) EntryFor(proofType CommitmentProofType, rootType RootType, rootID string) (ProofRegistryEntry, bool) {
	registry = normalizeProofRegistry(registry)
	needle := proofRegistryEntryKey(ProofRegistryEntry{
		ProofType:	CommitmentProofType(strings.TrimSpace(string(proofType))),
		RootType:	RootType(strings.TrimSpace(string(rootType))),
		RootID:		strings.TrimSpace(rootID),
	})
	for _, entry := range registry.Entries {
		if proofRegistryEntryKey(entry) == needle {
			return entry, true
		}
	}
	return ProofRegistryEntry{}, false
}

func (query RootQuery) Validate() error {
	query = normalizeRootQuery(query)
	if query.Height == 0 {
		return errors.New("aetracore root query height must be positive")
	}
	if err := validatePolicyID("aetracore root query type", string(query.RootType)); err != nil {
		return err
	}
	return validateToken("aetracore root query id", query.RootID, MaxScopeLength)
}

func (response RootQueryResponse) Validate() error {
	response.QueryHash = strings.ToLower(strings.TrimSpace(response.QueryHash))
	if err := ValidateHash("aetracore root query hash", response.QueryHash); err != nil {
		return err
	}
	if !response.Found {
		return nil
	}
	if err := response.Root.Validate(); err != nil {
		return err
	}
	return response.ProofRoot.Validate()
}

func ComputeRootEncodingDescriptorHash(d RootEncodingDescriptor) string {
	return hashRoot("aetra-aek-root-encoding-descriptor-v1", func(w byteWriter) {
		writePart(w, strings.TrimSpace(d.Version))
		writePart(w, strings.ToLower(strings.TrimSpace(d.EmptyRoot)))
	})
}

func ComputeProofRegistryEntryHash(entry ProofRegistryEntry) string {
	entry = normalizeProofRegistryEntry(entry)
	return hashRoot("aetra-aek-proof-registry-entry-v1", func(w byteWriter) {
		writePart(w, string(entry.ProofType))
		writePart(w, string(entry.RootType))
		writePart(w, entry.RootID)
		writePart(w, entry.RootHash)
		writePart(w, fmt.Sprint(entry.Enabled))
	})
}

func ComputeProofRegistryRoot(registry ProofRegistry) string {
	registry = normalizeProofRegistry(registry)
	return hashRoot("aetra-aek-proof-registry-root-v1", func(w byteWriter) {
		writeUint64(w, registry.Height)
		writeUint64(w, uint64(len(registry.Entries)))
		for _, entry := range registry.Entries {
			writePart(w, entry.RegistryHash)
		}
	})
}

func ComputeRootQueryHash(query RootQuery) string {
	query = normalizeRootQuery(query)
	return hashRoot("aetra-aek-root-query-v1", func(w byteWriter) {
		writeUint64(w, query.Height)
		writePart(w, string(query.RootType))
		writePart(w, query.RootID)
	})
}

func ComputeProofVerificationResultHash(result ProofVerificationResult) string {
	return hashRoot("aetra-aek-proof-verification-result-v1", func(w byteWriter) {
		writePart(w, fmt.Sprint(result.Verified))
		writePart(w, strings.ToLower(strings.TrimSpace(result.ProofHash)))
		writePart(w, strings.ToLower(strings.TrimSpace(result.RegistryRoot)))
		writePart(w, strings.ToLower(strings.TrimSpace(result.VerifiedRoot)))
	})
}

func normalizeProofRegistryEntry(entry ProofRegistryEntry) ProofRegistryEntry {
	entry.ProofType = CommitmentProofType(strings.TrimSpace(string(entry.ProofType)))
	entry.RootType = RootType(strings.TrimSpace(string(entry.RootType)))
	entry.RootID = strings.TrimSpace(entry.RootID)
	entry.RootHash = strings.ToLower(strings.TrimSpace(entry.RootHash))
	entry.RegistryHash = strings.ToLower(strings.TrimSpace(entry.RegistryHash))
	return entry
}

func normalizeProofRegistryEntries(entries []ProofRegistryEntry) []ProofRegistryEntry {
	out := make([]ProofRegistryEntry, len(entries))
	for i, entry := range entries {
		entry = normalizeProofRegistryEntry(entry)
		if entry.RootHash == "" {
			entry.RootHash = DeterministicEmptyRootCommitment(entry.RootType, entry.RootID)
		}
		if entry.RegistryHash == "" {
			entry.RegistryHash = ComputeProofRegistryEntryHash(entry)
		}
		out[i] = entry
	}
	sort.SliceStable(out, func(i, j int) bool {
		return proofRegistryEntryKey(out[i]) < proofRegistryEntryKey(out[j])
	})
	return out
}

func normalizeProofRegistry(registry ProofRegistry) ProofRegistry {
	registry.Entries = normalizeProofRegistryEntries(registry.Entries)
	registry.RegistryRoot = strings.ToLower(strings.TrimSpace(registry.RegistryRoot))
	return registry
}

func normalizeRootQuery(query RootQuery) RootQuery {
	query.RootType = RootType(strings.TrimSpace(string(query.RootType)))
	query.RootID = strings.TrimSpace(query.RootID)
	return query
}

func proofRegistryEntryKey(entry ProofRegistryEntry) string {
	entry = normalizeProofRegistryEntry(entry)
	return string(entry.ProofType) + "/" + string(entry.RootType) + "/" + entry.RootID
}
