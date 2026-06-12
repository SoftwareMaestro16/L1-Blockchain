package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityProofSchemaVersionV2	uint64	= 1

	IdentityResolutionProofCommitmentDomainV2	= "identity-v2-resolution-proof-commitment"
	RecursiveResolutionProofCommitmentDomainV2	= "identity-v2-recursive-resolution-proof-commitment"
)

type IdentityProofQueryTypeV2 string

const (
	IdentityProofQueryResolvePrimary	IdentityProofQueryTypeV2	= "resolve_primary"
	IdentityProofQueryResolveRecord		IdentityProofQueryTypeV2	= "resolve_record"
	IdentityProofQueryResolveReverse	IdentityProofQueryTypeV2	= "resolve_reverse"
	IdentityProofQueryDomainExists		IdentityProofQueryTypeV2	= "domain_exists"
	IdentityProofQueryDomainAbsent		IdentityProofQueryTypeV2	= "domain_absent"
)

var IdentityResolutionProofFormatV2FieldOrder = []string{
	"proof_version",
	"chain_id",
	"height",
	"app_hash",
	"name",
	"name_hash",
	"query_type",
	"normalized_name_proof",
	"domain_record",
	"domain_record_proof",
	"nft_binding",
	"nft_binding_proof",
	"resolver_record",
	"resolver_record_proof",
	"reverse_record_optional",
	"reverse_record_proof_optional",
	"delegation_chain",
	"delegation_chain_proofs",
	"subdomain_path",
	"subdomain_path_proofs",
	"non_existence_proof_optional",
	"record_version",
	"proof_commitment_hash",
}

var RecursiveResolutionProofV2FieldOrder = []string{
	"proof_version",
	"chain_id",
	"height",
	"root_name",
	"target_name",
	"path_labels",
	"path_hashes",
	"path_domain_records",
	"path_resolver_records",
	"path_delegation_records",
	"path_proofs",
	"final_resolution_record",
	"final_record_proof",
	"cache_record_optional",
	"cache_record_proof_optional",
}

type IdentityResolutionProofFormatV2 struct {
	ProofVersion			uint64
	ChainID				string
	Height				uint64
	AppHash				string
	Name				string
	NameHash			string
	QueryType			IdentityProofQueryTypeV2
	NormalizedNameProof		string
	DomainRecord			*DomainRecordV2
	DomainRecordProof		*IdentityInclusionProof
	NFTBinding			*DomainNFTBinding
	NFTBindingProof			*IdentityInclusionProof
	ResolverRecord			*UnifiedResolutionRecordV2
	ResolverRecordProof		*IdentityInclusionProof
	ReverseRecordOptional		*ReverseResolutionRecordV2
	ReverseRecordProofOptional	*IdentityInclusionProof
	DelegationChain			[]DelegationRecordV2
	DelegationChainProofs		[]IdentityInclusionProof
	SubdomainPath			[]SubdomainRecord
	SubdomainPathProofs		[]IdentityInclusionProof
	NonExistenceProofOptional	*IdentityAbsenceProof
	RecordVersion			uint64
	ProofCommitmentHash		string
}

type RecursiveResolutionProofV2 struct {
	ProofVersion			uint64
	ChainID				string
	Height				uint64
	RootName			string
	TargetName			string
	PathLabels			[]string
	PathHashes			[]string
	PathDomainRecords		[]DomainRecordV2
	PathResolverRecords		[]UnifiedResolutionRecordV2
	PathDelegationRecords		[]DelegationRecordV2
	PathProofs			[]IdentityInclusionProof
	FinalResolutionRecord		UnifiedResolutionRecordV2
	FinalRecordProof		IdentityInclusionProof
	CacheRecordOptional		*ResolutionCacheRecordV2
	CacheRecordProofOptional	*IdentityInclusionProof
	ProofCommitmentHash		string
}

func BuildIdentityResolutionProofFormatV2(state IdentityState, chainID string, appHash string, name string, queryType IdentityProofQueryTypeV2, height uint64, ttl uint64, reverseAddress sdk.AccAddress) (IdentityResolutionProofFormatV2, error) {
	if err := validateProofFormatHeaderV2(chainID, appHash, height); err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}
	if ttl == 0 {
		return IdentityResolutionProofFormatV2{}, errors.New("identity v2 proof format ttl is required")
	}
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}
	if err := validateIdentityProofQueryTypeV2(queryType); err != nil {
		return IdentityResolutionProofFormatV2{}, err
	}

	out := IdentityResolutionProofFormatV2{
		ProofVersion:		IdentityProofSchemaVersionV2,
		ChainID:		chainID,
		Height:			height,
		AppHash:		appHash,
		Name:			normalized,
		NameHash:		nameHash,
		QueryType:		queryType,
		NormalizedNameProof:	identityHash("identity-v2-normalized-name-proof", normalized, normalized, nameHash, fmt.Sprintf("%020d", NameNormalizationVersionV2)),
		RecordVersion:		1,
	}

	if domain, found := findDomain(state, normalized); found {
		record, err := domainRecordV2ForProof(state, domain, height)
		if err != nil {
			return IdentityResolutionProofFormatV2{}, err
		}
		domainProof, err := BuildIdentityProof(state, mustIdentityDomainStoreKeyV2(domain.Name))
		if err != nil {
			return IdentityResolutionProofFormatV2{}, err
		}
		out.DomainRecord = &record
		out.DomainRecordProof = &domainProof
	} else {
		absence, err := BuildIdentityAbsenceProof(state, mustIdentityDomainStoreKeyV2(normalized))
		if err != nil {
			return IdentityResolutionProofFormatV2{}, err
		}
		out.NonExistenceProofOptional = &absence
	}

	if resolutionProof, err := BuildIdentityResolutionProof(state, normalized, height); err == nil {
		resolverRecord, err := BuildUnifiedResolutionRecordV2(state, resolutionProof.ResolverDomain, height, ttl)
		if err != nil {
			return IdentityResolutionProofFormatV2{}, err
		}
		resolverRecord.RecordVersion = ResolverRecordVersionV2(resolutionProof.Resolver)
		nftBinding, err := nftBindingForProof(resolutionProof.AuthorityDomain, resolutionProof.AuthorityNFT, height)
		if err != nil {
			return IdentityResolutionProofFormatV2{}, err
		}
		out.NFTBinding = &nftBinding
		out.NFTBindingProof = &resolutionProof.NFTProof
		out.ResolverRecord = &resolverRecord
		out.ResolverRecordProof = &resolutionProof.ResolverProof
		out.RecordVersion = resolverRecord.RecordVersion
		out.SubdomainPath, out.SubdomainPathProofs, err = buildSubdomainDelegationProofsV2(state, normalized)
		if err != nil {
			return IdentityResolutionProofFormatV2{}, err
		}
		if len(reverseAddress) > 0 {
			reverseRecord, reverseProof, err := reverseProofForFormatV2(state, reverseAddress, height, []string{ResolverKeyWallet})
			if err != nil {
				return IdentityResolutionProofFormatV2{}, err
			}
			out.ReverseRecordOptional = &reverseRecord
			out.ReverseRecordProofOptional = &reverseProof
		}
	} else if queryType != IdentityProofQueryDomainAbsent && queryType != IdentityProofQueryDomainExists {
		return IdentityResolutionProofFormatV2{}, err
	}

	out.ProofCommitmentHash = ComputeIdentityResolutionProofCommitmentHashV2(out)
	return out, ValidateIdentityResolutionProofFormatV2(out)
}

func ValidateIdentityResolutionProofFormatV2(proof IdentityResolutionProofFormatV2) error {
	if proof.ProofVersion != IdentityProofSchemaVersionV2 {
		return fmt.Errorf("unsupported identity v2 resolution proof schema version %d", proof.ProofVersion)
	}
	if err := validateProofFormatHeaderV2(proof.ChainID, proof.AppHash, proof.Height); err != nil {
		return err
	}
	if proof.Name == "" || proof.NameHash == "" {
		return errors.New("identity v2 resolution proof name and name_hash are required")
	}
	normalized, err := NormalizeAETDomain(proof.Name)
	if err != nil {
		return err
	}
	if proof.Name != normalized {
		return errors.New("identity v2 resolution proof name must be normalized")
	}
	expectedNameHash, err := DomainRecordV2NameHash(proof.Name)
	if err != nil {
		return err
	}
	if proof.NameHash != expectedNameHash {
		return errors.New("identity v2 resolution proof name_hash mismatch")
	}
	if err := validateIdentityProofQueryTypeV2(proof.QueryType); err != nil {
		return err
	}
	if proof.NormalizedNameProof != identityHash("identity-v2-normalized-name-proof", proof.Name, proof.Name, proof.NameHash, fmt.Sprintf("%020d", NameNormalizationVersionV2)) {
		return errors.New("identity v2 normalized_name_proof mismatch")
	}
	if proof.DomainRecord != nil && proof.DomainRecordProof == nil {
		return errors.New("identity v2 domain_record_proof is required")
	}
	if proof.DomainRecordProof != nil {
		if err := VerifyIdentityProof(*proof.DomainRecordProof); err != nil {
			return err
		}
	}
	if proof.NFTBinding != nil && proof.NFTBindingProof == nil {
		return errors.New("identity v2 nft_binding_proof is required")
	}
	if proof.NFTBindingProof != nil {
		if err := VerifyIdentityProof(*proof.NFTBindingProof); err != nil {
			return err
		}
	}
	if proof.ResolverRecord != nil && proof.ResolverRecordProof == nil {
		return errors.New("identity v2 resolver_record_proof is required")
	}
	if proof.ResolverRecord != nil {
		if err := ValidateUnifiedResolutionRecordV2(*proof.ResolverRecord); err != nil {
			return err
		}
	}
	if proof.ResolverRecordProof != nil {
		if err := VerifyIdentityProof(*proof.ResolverRecordProof); err != nil {
			return err
		}
	}
	if proof.ReverseRecordOptional != nil && proof.ReverseRecordProofOptional == nil {
		return errors.New("identity v2 reverse_record_proof_optional is required")
	}
	if proof.ReverseRecordProofOptional != nil {
		if err := VerifyIdentityProof(*proof.ReverseRecordProofOptional); err != nil {
			return err
		}
	}
	if len(proof.DelegationChain) != len(proof.DelegationChainProofs) {
		return errors.New("identity v2 delegation chain proof count mismatch")
	}
	if len(proof.SubdomainPath) != len(proof.SubdomainPathProofs) {
		return errors.New("identity v2 subdomain path proof count mismatch")
	}
	for _, subdomainProof := range proof.SubdomainPathProofs {
		if err := VerifyIdentityProof(subdomainProof); err != nil {
			return err
		}
	}
	if proof.NonExistenceProofOptional != nil {
		if err := VerifyIdentityAbsenceProof(*proof.NonExistenceProofOptional); err != nil {
			return err
		}
	}
	if proof.RecordVersion == 0 {
		return errors.New("identity v2 resolution proof record_version is required")
	}
	if proof.ProofCommitmentHash == "" || proof.ProofCommitmentHash != ComputeIdentityResolutionProofCommitmentHashV2(proof) {
		return errors.New("identity v2 resolution proof commitment hash mismatch")
	}
	return nil
}

func EncodeIdentityResolutionProofFormatV2(proof IdentityResolutionProofFormatV2) ([]byte, error) {
	encoder := newIdentityProofBinaryEncoderV2()
	encoder.writeString("IdentityResolutionProof")
	encoder.writeStringSlice(IdentityResolutionProofFormatV2FieldOrder)
	encoder.writeUint64(proof.ProofVersion)
	encoder.writeString(proof.ChainID)
	encoder.writeUint64(proof.Height)
	encoder.writeString(proof.AppHash)
	encoder.writeString(proof.Name)
	encoder.writeString(proof.NameHash)
	encoder.writeString(string(proof.QueryType))
	encoder.writeString(proof.NormalizedNameProof)
	encoder.writeDomainRecordV2Ptr(proof.DomainRecord)
	encoder.writeInclusionProofPtr(proof.DomainRecordProof)
	encoder.writeNFTBindingPtr(proof.NFTBinding)
	encoder.writeInclusionProofPtr(proof.NFTBindingProof)
	encoder.writeUnifiedResolutionRecordPtr(proof.ResolverRecord)
	encoder.writeInclusionProofPtr(proof.ResolverRecordProof)
	encoder.writeReverseRecordPtr(proof.ReverseRecordOptional)
	encoder.writeInclusionProofPtr(proof.ReverseRecordProofOptional)
	encoder.writeDelegations(proof.DelegationChain)
	encoder.writeInclusionProofs(proof.DelegationChainProofs)
	encoder.writeSubdomains(proof.SubdomainPath)
	encoder.writeInclusionProofs(proof.SubdomainPathProofs)
	encoder.writeAbsenceProofPtr(proof.NonExistenceProofOptional)
	encoder.writeUint64(proof.RecordVersion)
	encoder.writeString(proof.ProofCommitmentHash)
	return encoder.bytes(), nil
}

func ComputeIdentityResolutionProofCommitmentHashV2(proof IdentityResolutionProofFormatV2) string {
	proof.ProofCommitmentHash = ""
	encoded, _ := EncodeIdentityResolutionProofFormatV2(proof)
	sum := sha256.Sum256(append([]byte(IdentityResolutionProofCommitmentDomainV2), encoded...))
	return hex.EncodeToString(sum[:])
}

func BuildRecursiveResolutionProofV2(state IdentityState, chainID string, rootName string, targetName string, height uint64, ttl uint64, cache *ResolutionCacheRecordV2) (RecursiveResolutionProofV2, error) {
	stateRoot, err := IdentityStateRoot(state)
	if err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	if err := validateProofFormatHeaderV2(chainID, stateRoot, height); err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	if ttl == 0 {
		return RecursiveResolutionProofV2{}, errors.New("identity v2 recursive proof ttl is required")
	}
	rootResult, err := NormalizeAETDomainVersioned(rootName, NameNormalizationVersionV2)
	if err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	targetResult, err := NormalizeAETDomainVersioned(targetName, NameNormalizationVersionV2)
	if err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	path, err := CanonicalResolutionPathV2(targetResult.NormalizedName)
	if err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	if len(path.Path) == 0 || path.Path[0] != rootResult.NormalizedName {
		return RecursiveResolutionProofV2{}, errors.New("identity v2 recursive proof target is not under root_name")
	}
	out := RecursiveResolutionProofV2{
		ProofVersion:	IdentityProofSchemaVersionV2,
		ChainID:	chainID,
		Height:		height,
		RootName:	rootResult.NormalizedName,
		TargetName:	targetResult.NormalizedName,
		PathLabels:	append([]string(nil), path.Labels...),
	}
	for i, candidate := range path.Path {
		out.PathHashes = append(out.PathHashes, path.PathHashes[i])
		if domain, found := findDomain(state, candidate); found {
			record, err := domainRecordV2ForProof(state, domain, height)
			if err != nil {
				return RecursiveResolutionProofV2{}, err
			}
			out.PathDomainRecords = append(out.PathDomainRecords, record)
			proof, err := BuildIdentityProof(state, mustIdentityDomainStoreKeyV2(candidate))
			if err != nil {
				return RecursiveResolutionProofV2{}, err
			}
			out.PathProofs = append(out.PathProofs, proof)
		}
		if resolver, found := findResolver(state, candidate); found {
			record, err := BuildUnifiedResolutionRecordV2(state, candidate, height, ttl)
			if err != nil {
				return RecursiveResolutionProofV2{}, err
			}
			record.RecordVersion = ResolverRecordVersionV2(resolver)
			out.PathResolverRecords = append(out.PathResolverRecords, record)
			proof, err := BuildIdentityProof(state, mustIdentityResolverStoreKeyV2(candidate))
			if err != nil {
				return RecursiveResolutionProofV2{}, err
			}
			out.PathProofs = append(out.PathProofs, proof)
		}
	}
	finalProof, err := BuildIdentityResolutionProof(state, targetResult.NormalizedName, height)
	if err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	finalRecord, err := BuildUnifiedResolutionRecordV2(state, finalProof.ResolverDomain, height, ttl)
	if err != nil {
		return RecursiveResolutionProofV2{}, err
	}
	finalRecord.RecordVersion = ResolverRecordVersionV2(finalProof.Resolver)
	out.FinalResolutionRecord = finalRecord
	out.FinalRecordProof = finalProof.ResolverProof
	if cache != nil {
		copied := *cache
		out.CacheRecordOptional = &copied
	}
	out.ProofCommitmentHash = ComputeRecursiveResolutionProofCommitmentHashV2(out)
	return out, ValidateRecursiveResolutionProofV2(out)
}

func ValidateRecursiveResolutionProofV2(proof RecursiveResolutionProofV2) error {
	if proof.ProofVersion != IdentityProofSchemaVersionV2 {
		return fmt.Errorf("unsupported identity v2 recursive proof schema version %d", proof.ProofVersion)
	}
	if err := validateProofFormatHeaderV2(proof.ChainID, identityHash("recursive-app-hash-ok", proof.FinalRecordProof.RootHash), proof.Height); err != nil {
		if proof.FinalRecordProof.RootHash == "" {
			return err
		}
	}
	if proof.RootName == "" || proof.TargetName == "" {
		return errors.New("identity v2 recursive proof root_name and target_name are required")
	}
	if _, err := NormalizeAETDomain(proof.RootName); err != nil {
		return err
	}
	if _, err := NormalizeAETDomain(proof.TargetName); err != nil {
		return err
	}
	if len(proof.PathLabels) == 0 || len(proof.PathHashes) == 0 {
		return errors.New("identity v2 recursive proof path labels and hashes are required")
	}
	for _, pathHash := range proof.PathHashes {
		if err := validateHexHash("identity v2 recursive proof path hash", pathHash); err != nil {
			return err
		}
	}
	for _, record := range proof.PathDomainRecords {
		if err := ValidateDomainRecordV2(record, DomainRecordV2ValidationContext{CurrentHeight: proof.Height, NFTOwner: record.Owner}); err != nil {
			return err
		}
	}
	for _, record := range proof.PathResolverRecords {
		if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
			return err
		}
	}
	for _, delegation := range proof.PathDelegationRecords {
		if err := ValidateDelegationRecordV2(delegation); err != nil {
			return err
		}
	}
	for _, inclusion := range proof.PathProofs {
		if err := VerifyIdentityProof(inclusion); err != nil {
			return err
		}
	}
	if err := ValidateUnifiedResolutionRecordV2(proof.FinalResolutionRecord); err != nil {
		return err
	}
	if err := VerifyIdentityProof(proof.FinalRecordProof); err != nil {
		return err
	}
	if proof.CacheRecordOptional != nil {
		if err := ValidateResolutionCacheRecordV2(*proof.CacheRecordOptional); err != nil {
			return err
		}
	}
	if proof.CacheRecordProofOptional != nil {
		if err := VerifyIdentityProof(*proof.CacheRecordProofOptional); err != nil {
			return err
		}
	}
	if proof.ProofCommitmentHash == "" || proof.ProofCommitmentHash != ComputeRecursiveResolutionProofCommitmentHashV2(proof) {
		return errors.New("identity v2 recursive proof commitment hash mismatch")
	}
	return nil
}

func EncodeRecursiveResolutionProofV2(proof RecursiveResolutionProofV2) ([]byte, error) {
	encoder := newIdentityProofBinaryEncoderV2()
	encoder.writeString("RecursiveResolutionProof")
	encoder.writeStringSlice(RecursiveResolutionProofV2FieldOrder)
	encoder.writeUint64(proof.ProofVersion)
	encoder.writeString(proof.ChainID)
	encoder.writeUint64(proof.Height)
	encoder.writeString(proof.RootName)
	encoder.writeString(proof.TargetName)
	encoder.writeStringSlice(proof.PathLabels)
	encoder.writeStringSlice(proof.PathHashes)
	encoder.writeDomainRecordV2s(proof.PathDomainRecords)
	encoder.writeUnifiedResolutionRecords(proof.PathResolverRecords)
	encoder.writeDelegations(proof.PathDelegationRecords)
	encoder.writeInclusionProofs(proof.PathProofs)
	encoder.writeUnifiedResolutionRecord(proof.FinalResolutionRecord)
	encoder.writeInclusionProof(proof.FinalRecordProof)
	encoder.writeResolutionCachePtr(proof.CacheRecordOptional)
	encoder.writeInclusionProofPtr(proof.CacheRecordProofOptional)
	return encoder.bytes(), nil
}

func ComputeRecursiveResolutionProofCommitmentHashV2(proof RecursiveResolutionProofV2) string {
	proof.ProofCommitmentHash = ""
	encoded, _ := EncodeRecursiveResolutionProofV2(proof)
	sum := sha256.Sum256(append([]byte(RecursiveResolutionProofCommitmentDomainV2), encoded...))
	return hex.EncodeToString(sum[:])
}

type identityProofBinaryEncoderV2 struct {
	buf bytes.Buffer
}

func newIdentityProofBinaryEncoderV2() *identityProofBinaryEncoderV2 {
	return &identityProofBinaryEncoderV2{}
}

func (e *identityProofBinaryEncoderV2) bytes() []byte {
	return append([]byte(nil), e.buf.Bytes()...)
}

func (e *identityProofBinaryEncoderV2) writeBool(value bool) {
	if value {
		e.buf.WriteByte(1)
		return
	}
	e.buf.WriteByte(0)
}

func (e *identityProofBinaryEncoderV2) writeUint64(value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	e.buf.Write(out[:])
}

func (e *identityProofBinaryEncoderV2) writeInt(value int) {
	e.writeUint64(uint64(value))
}

func (e *identityProofBinaryEncoderV2) writeString(value string) {
	e.writeUint64(uint64(len(value)))
	e.buf.WriteString(value)
}

func (e *identityProofBinaryEncoderV2) writeBytes(value []byte) {
	e.writeUint64(uint64(len(value)))
	e.buf.Write(value)
}

func (e *identityProofBinaryEncoderV2) writeStringSlice(values []string) {
	e.writeUint64(uint64(len(values)))
	for _, value := range values {
		e.writeString(value)
	}
}

func (e *identityProofBinaryEncoderV2) writeAddress(value sdk.AccAddress) {
	e.writeBytes(value)
}

func (e *identityProofBinaryEncoderV2) writeInclusionProofPtr(proof *IdentityInclusionProof) {
	e.writeBool(proof != nil)
	if proof != nil {
		e.writeInclusionProof(*proof)
	}
}

func (e *identityProofBinaryEncoderV2) writeInclusionProofs(proofs []IdentityInclusionProof) {
	e.writeUint64(uint64(len(proofs)))
	for _, proof := range proofs {
		e.writeInclusionProof(proof)
	}
}

func (e *identityProofBinaryEncoderV2) writeInclusionProof(proof IdentityInclusionProof) {
	e.writeString(proof.RootHash)
	e.writeString(proof.Key)
	e.writeString(proof.ValueHash)
	e.writeString(proof.LeafHash)
	e.writeInt(proof.Index)
	e.writeInt(proof.LeafCount)
	e.writeUint64(uint64(len(proof.Steps)))
	for _, step := range proof.Steps {
		e.writeString(step.Hash)
		e.writeBool(step.SiblingOnLeft)
	}
}

func (e *identityProofBinaryEncoderV2) writeAbsenceProofPtr(proof *IdentityAbsenceProof) {
	e.writeBool(proof != nil)
	if proof == nil {
		return
	}
	e.writeString(proof.RootHash)
	e.writeString(proof.Key)
	e.writeInt(proof.LeafCount)
	e.writeInclusionProofPtr(proof.Previous)
	e.writeInclusionProofPtr(proof.Next)
}

func (e *identityProofBinaryEncoderV2) writeDomainRecordV2Ptr(record *DomainRecordV2) {
	e.writeBool(record != nil)
	if record != nil {
		e.writeDomainRecordV2(*record)
	}
}

func (e *identityProofBinaryEncoderV2) writeDomainRecordV2s(records []DomainRecordV2) {
	e.writeUint64(uint64(len(records)))
	for _, record := range records {
		e.writeDomainRecordV2(record)
	}
}

func (e *identityProofBinaryEncoderV2) writeDomainRecordV2(record DomainRecordV2) {
	e.writeString(record.Name)
	e.writeString(record.NameHash)
	e.writeString(record.NormalizedName)
	e.writeString(record.ParentNameHash)
	e.writeString(record.TLD)
	e.writeAddress(record.Owner)
	e.writeAddress(record.Resolver)
	e.writeUint64(record.ExpiryHeight)
	e.writeUint64(uint64(record.ExpiryTime))
	e.writeUint64(record.RenewalStartHeight)
	e.writeString(record.NFTClassID)
	e.writeString(record.NFTItemID)
	e.writeString(string(record.Status))
	e.writeUint64(record.LifecycleEpoch)
	e.writeUint64(record.CreatedAtHeight)
	e.writeUint64(record.UpdatedAtHeight)
	e.writeUint64(record.Version)
	e.writeUint64(record.Flags)
}

func (e *identityProofBinaryEncoderV2) writeNFTBindingPtr(binding *DomainNFTBinding) {
	e.writeBool(binding != nil)
	if binding == nil {
		return
	}
	e.writeString(binding.NameHash)
	e.writeString(binding.NFTClassID)
	e.writeString(binding.NFTItemID)
	e.writeAddress(binding.Owner)
	e.writeUint64(binding.LastVerifiedHeight)
	e.writeUint64(binding.BindingVersion)
}

func (e *identityProofBinaryEncoderV2) writeUnifiedResolutionRecordPtr(record *UnifiedResolutionRecordV2) {
	e.writeBool(record != nil)
	if record != nil {
		e.writeUnifiedResolutionRecord(*record)
	}
}

func (e *identityProofBinaryEncoderV2) writeUnifiedResolutionRecords(records []UnifiedResolutionRecordV2) {
	e.writeUint64(uint64(len(records)))
	for _, record := range records {
		e.writeUnifiedResolutionRecord(record)
	}
}

func (e *identityProofBinaryEncoderV2) writeUnifiedResolutionRecord(record UnifiedResolutionRecordV2) {
	e.writeString(record.NameHash)
	e.writeAddress(record.Owner)
	e.writeAddress(record.PrimaryAddress)
	e.writeUint64(uint64(len(record.ContractTargets)))
	for _, target := range record.ContractTargets {
		e.writeString(target.Key)
		e.writeAddress(target.Address)
		e.writeString(target.CodeID)
		e.writeString(target.TargetID)
		e.writeAddress(target.ContractAddress)
		e.writeString(target.Entrypoint)
		e.writeString(target.InterfaceHash)
		e.writeString(target.RequiredFundsPolicy)
		e.writeUint64(target.GasHint)
		e.writeBool(target.Enabled)
		e.writeUint64(target.UpdatedAtHeight)
	}
	e.writeUint64(uint64(len(record.ServiceEndpoints)))
	for _, endpoint := range record.ServiceEndpoints {
		e.writeString(endpoint.Key)
		e.writeString(endpoint.Endpoint)
		e.writeString(endpoint.ServiceID)
		e.writeString(endpoint.ServiceType)
		e.writeString(endpoint.Transport)
		e.writeString(endpoint.AuthPolicy)
		e.writeString(endpoint.HealthPathOptional)
		e.writeUint64(uint64(endpoint.Priority))
		e.writeUint64(uint64(endpoint.Weight))
		e.writeUint64(endpoint.TTL)
		e.writeString(endpoint.SchemaHashOptional)
	}
	e.writeUint64(uint64(len(record.InterfaceDescriptors)))
	for _, descriptor := range record.InterfaceDescriptors {
		e.writeString(descriptor.InterfaceID)
		e.writeString(descriptor.Descriptor)
		e.writeString(descriptor.SchemaHash)
		e.writeString(descriptor.SchemaURIOptional)
		e.writeString(descriptor.SchemaInlineOptional)
		e.writeString(descriptor.Version)
		e.writeString(descriptor.RenderPolicy)
		e.writeStringSlice(descriptor.PermissionsRequired)
		e.writeString(descriptor.ContractTargetIDOptional)
		e.writeString(descriptor.ServiceIDOptional)
	}
	e.writeString(record.RoutingMetadata.ZoneID)
	e.writeString(record.RoutingMetadata.ShardID)
	e.writeString(record.RoutingMetadata.VM)
	e.writeString(record.RoutingMetadata.Entrypoint)
	e.writeString(record.RoutingMetadata.RouteID)
	e.writeString(record.RoutingMetadata.TargetType)
	e.writeString(record.RoutingMetadata.PreferredTarget)
	e.writeStringSlice(record.RoutingMetadata.FallbackTargets)
	e.writeString(record.RoutingMetadata.ChainContext)
	e.writeString(record.RoutingMetadata.FeeHint)
	e.writeUint64(record.RoutingMetadata.TimeoutHint)
	e.writeString(record.RoutingMetadata.MemoPolicy)
	e.writeStringSlice(record.RoutingMetadata.CapabilityRequirements)
	e.writeUint64(uint64(len(record.ExecutionHints)))
	for _, hint := range record.ExecutionHints {
		e.writeString(hint.Key)
		e.writeString(hint.Value)
		e.writeUint64(hint.DefaultGasLimitHint)
		e.writeString(hint.PreferredFeeMode)
		e.writeString(hint.MessageType)
		e.writeBool(hint.AsyncAllowed)
		e.writeBool(hint.RequiresMemo)
		e.writeBool(hint.RequiresInterfaceConfirmation)
		e.writeBool(hint.SimulationRequired)
	}
	e.writeUint64(record.RecordVersion)
	e.writeUint64(record.RecordTTL)
	e.writeUint64(record.UpdatedAtHeight)
	e.writeUint64(record.MaxPayloadBytes)
	e.writeUint64(record.SchemaVersion)
	e.writeBytes(record.OwnerSignatureOptional)
}

func (e *identityProofBinaryEncoderV2) writeReverseRecordPtr(record *ReverseResolutionRecordV2) {
	e.writeBool(record != nil)
	if record == nil {
		return
	}
	e.writeAddress(record.Address)
	e.writeString(record.NameHash)
	e.writeString(record.Name)
	e.writeBool(record.Verified)
	e.writeUint64(record.UpdatedAtHeight)
	e.writeUint64(record.ExpiryHeight)
}

func (e *identityProofBinaryEncoderV2) writeDelegations(records []DelegationRecordV2) {
	ordered := append([]DelegationRecordV2(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].NameHash != ordered[j].NameHash {
			return ordered[i].NameHash < ordered[j].NameHash
		}
		leftDelegate := hex.EncodeToString(ordered[i].Delegate)
		rightDelegate := hex.EncodeToString(ordered[j].Delegate)
		if leftDelegate != rightDelegate {
			return leftDelegate < rightDelegate
		}
		return ordered[i].Scope < ordered[j].Scope
	})
	e.writeUint64(uint64(len(ordered)))
	for _, record := range ordered {
		e.writeString(record.NameHash)
		e.writeAddress(record.Delegate)
		e.writeString(string(record.Scope))
		e.writeUint64(uint64(record.ScopeBits))
		e.writeStringSlice(record.Permissions)
		e.writeUint64(record.ExpiresAtHeight)
		e.writeUint64(uint64(record.SubtreeLimit))
		e.writeString(record.RecordPrefixLimit)
		e.writeUint64(record.CreatedAtHeight)
		e.writeUint64(record.TimeLockedUntilHeight)
		e.writeUint64(effectiveDelegationVersionV2(record))
		e.writeBool(record.CanTransferParent)
	}
}

func (e *identityProofBinaryEncoderV2) writeSubdomains(records []SubdomainRecord) {
	ordered := append([]SubdomainRecord(nil), records...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Name < ordered[j].Name })
	e.writeUint64(uint64(len(ordered)))
	for _, record := range ordered {
		e.writeString(record.ParentName)
		e.writeString(record.Name)
		e.writeAddress(record.Owner)
		e.writeBool(record.ParentControlsRecord)
		e.writeUint64(record.CreatedHeight)
		e.writeString(string(record.DelegationType))
		e.writeBool(record.Detached)
		e.writeBool(record.Ephemeral)
		e.writeUint64(record.ExpiryHeight)
		e.writeUint64(record.TimeLockedUntilHeight)
		e.writeBool(record.ParentAuthorized)
	}
}

func (e *identityProofBinaryEncoderV2) writeResolutionCachePtr(record *ResolutionCacheRecordV2) {
	e.writeBool(record != nil)
	if record == nil {
		return
	}
	e.writeString(record.NameHash)
	e.writeString(record.ResolutionPathHash)
	e.writeString(record.ResolvedRecordHash)
	e.writeUint64(record.ValidUntilHeight)
	e.writeUint64(record.SourceVersion)
	e.writeUint64(record.ParentEpoch)
	e.writeUint64(record.ChildEpoch)
}

func validateProofFormatHeaderV2(chainID string, appHash string, height uint64) error {
	if strings.TrimSpace(chainID) == "" {
		return errors.New("identity v2 proof chain_id is required")
	}
	if strings.TrimSpace(chainID) != chainID {
		return errors.New("identity v2 proof chain_id must not have surrounding whitespace")
	}
	if height == 0 {
		return errors.New("identity v2 proof height is required")
	}
	return validateHexHash("identity v2 proof app_hash", appHash)
}

func validateIdentityProofQueryTypeV2(queryType IdentityProofQueryTypeV2) error {
	switch queryType {
	case IdentityProofQueryResolvePrimary,
		IdentityProofQueryResolveRecord,
		IdentityProofQueryResolveReverse,
		IdentityProofQueryDomainExists,
		IdentityProofQueryDomainAbsent:
		return nil
	default:
		return fmt.Errorf("unknown identity v2 proof query type %q", queryType)
	}
}

func domainRecordV2ForProof(state IdentityState, domain Domain, height uint64) (DomainRecordV2, error) {
	status, err := DomainLifecycle(state, domain.Name, height)
	if err != nil {
		return DomainRecordV2{}, err
	}
	recordStatus := DomainRecordV2Active
	switch status {
	case DomainLifecycleAvailable:
		recordStatus = DomainRecordV2Available
	case DomainLifecycleCommitted:
		recordStatus = DomainRecordV2Committed
	case DomainLifecycleActive:
		recordStatus = DomainRecordV2Active
	case DomainLifecycleRenewalWindow:
		recordStatus = DomainRecordV2RenewalWindow
	case DomainLifecycleExpired:
		recordStatus = DomainRecordV2Expired
	}
	return NewDomainRecordV2FromDomain(domain, recordStatus, int64(domain.ExpiryHeight), height)
}

func nftBindingForProof(domain Domain, nft DomainNFT, height uint64) (DomainNFTBinding, error) {
	nameHash, err := DomainRecordV2NameHash(domain.Name)
	if err != nil {
		return DomainNFTBinding{}, err
	}
	lastVerifiedHeight := nft.TransferHeight
	if lastVerifiedHeight == 0 {
		lastVerifiedHeight = nft.MintHeight
	}
	if lastVerifiedHeight == 0 {
		lastVerifiedHeight = height
	}
	binding := DomainNFTBinding{
		NameHash:		nameHash,
		NFTClassID:		DomainNFTClassID,
		NFTItemID:		nft.ID,
		Owner:			cloneSpecAddress(nft.Owner),
		LastVerifiedHeight:	lastVerifiedHeight,
		BindingVersion:		1,
	}
	return binding, ValidateDomainNFTBinding(binding, DomainNFTBindingContext{RegistryOwner: domain.Owner, NFTModuleOwner: nft.Owner, CurrentHeight: height})
}

func reverseProofForFormatV2(state IdentityState, address sdk.AccAddress, height uint64, authorizedAliasKeys []string) (ReverseResolutionRecordV2, IdentityInclusionProof, error) {
	key, err := IdentityReverseStoreKey(address)
	if err != nil {
		return ReverseResolutionRecordV2{}, IdentityInclusionProof{}, err
	}
	proof, err := BuildIdentityProof(state, key)
	if err != nil {
		return ReverseResolutionRecordV2{}, IdentityInclusionProof{}, err
	}
	for _, reverse := range state.ReverseRecords {
		if !bytes.Equal(reverse.Address, address) {
			continue
		}
		record, err := reverseRecordV2FromLegacy(state, reverse, true)
		if err != nil {
			return ReverseResolutionRecordV2{}, IdentityInclusionProof{}, err
		}
		if err := ValidateReverseResolutionRecordV2(state, record, height, authorizedAliasKeys); err != nil {
			return ReverseResolutionRecordV2{}, IdentityInclusionProof{}, err
		}
		return record, proof, nil
	}
	return ReverseResolutionRecordV2{}, IdentityInclusionProof{}, errors.New("identity v2 reverse proof record not found")
}

func mustIdentityDomainStoreKeyV2(name string) string {
	key, err := IdentityDomainStoreKey(name)
	if err != nil {
		panic(err)
	}
	return key
}

func mustIdentityResolverStoreKeyV2(name string) string {
	key, err := IdentityResolverStoreKey(name)
	if err != nil {
		panic(err)
	}
	return key
}
