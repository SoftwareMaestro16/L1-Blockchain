package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const identityParamsStoreKey = IdentityStoreV2Prefix + "/params"

type IdentityProofLeaf struct {
	Key		string
	ValueHash	string
	LeafHash	string
}

type IdentityProofStep struct {
	Hash		string
	SiblingOnLeft	bool
}

type IdentityInclusionProof struct {
	RootHash	string
	Key		string
	ValueHash	string
	LeafHash	string
	Index		int
	LeafCount	int
	Steps		[]IdentityProofStep
}

type IdentityAbsenceProof struct {
	RootHash	string
	Key		string
	LeafCount	int
	Previous	*IdentityInclusionProof
	Next		*IdentityInclusionProof
}

type IdentityResolutionCandidateProof struct {
	Domain		string
	DomainProof	*IdentityInclusionProof
	DomainAbsence	*IdentityAbsenceProof
	ResolverProof	*IdentityInclusionProof
	ResolverAbsence	*IdentityAbsenceProof
}

type IdentityResolutionProof struct {
	StateRoot	string
	QueryDomain	string
	ResolverDomain	string
	AuthorityDomain	Domain
	AuthorityNFT	DomainNFT
	Resolver	ResolverRecord
	Candidates	[]IdentityResolutionCandidateProof
	DomainProof	IdentityInclusionProof
	NFTProof	IdentityInclusionProof
	ResolverProof	IdentityInclusionProof
}

func IdentityStateRoot(state IdentityState) (string, error) {
	leaves, err := IdentityStateLeaves(state)
	if err != nil {
		return "", err
	}
	return identityMerkleRoot(leaves), nil
}

func IdentityStateLeaves(state IdentityState) ([]IdentityProofLeaf, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return nil, err
	}
	leaves := make([]IdentityProofLeaf, 0, 1+len(state.Domains)+len(state.DomainNFTs)+len(state.Commits)+len(state.UsedCommitments)+len(state.Resolvers)+len(state.ReverseRecords)+len(state.Subdomains)+len(state.Auctions)+len(state.PendingResolverUpdates))
	leaves = append(leaves, identityLeaf(identityParamsStoreKey, identityHash(
		"identity-params",
		fmt.Sprintf("%020d", state.Params.RegistrationPeriodBlocks),
		fmt.Sprintf("%020d", state.Params.RenewalWindowBlocks),
		fmt.Sprintf("%020d", state.Params.CommitTTLBlocks),
		fmt.Sprintf("%020d", state.Params.AuctionCommitBlocks),
		fmt.Sprintf("%020d", state.Params.AuctionRevealBlocks),
	)))
	for _, domain := range state.Domains {
		leaf, err := identityDomainLeaf(domain)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, nft := range state.DomainNFTs {
		leaf, err := identityNFTLeaf(nft)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, commit := range state.Commits {
		leaf, err := identityCommitLeaf(commit)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, commitment := range state.UsedCommitments {
		leaf, err := identityUsedCommitmentLeaf(commitment)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, resolver := range state.Resolvers {
		leaf, err := identityResolverLeaf(resolver)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, reverse := range state.ReverseRecords {
		leaf, err := identityReverseLeaf(reverse)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, subdomain := range state.Subdomains {
		leaf, err := identitySubdomainLeaf(subdomain)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, auction := range state.Auctions {
		leaf, err := identityAuctionLeaf(auction)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	for _, intent := range state.PendingResolverUpdates {
		leaf, err := identityPendingResolverLeaf(intent)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	sortIdentityProofLeaves(leaves)
	for i := 1; i < len(leaves); i++ {
		if leaves[i-1].Key == leaves[i].Key {
			return nil, fmt.Errorf("duplicate identity proof leaf key %q", leaves[i].Key)
		}
	}
	return leaves, nil
}

func BuildIdentityProof(state IdentityState, key string) (IdentityInclusionProof, error) {
	leaves, err := IdentityStateLeaves(state)
	if err != nil {
		return IdentityInclusionProof{}, err
	}
	index := sort.Search(len(leaves), func(i int) bool { return leaves[i].Key >= key })
	if index >= len(leaves) || leaves[index].Key != key {
		return IdentityInclusionProof{}, fmt.Errorf("identity proof key %q not found", key)
	}
	return buildIdentityProofFromLeaves(leaves, index), nil
}

func BuildIdentityAbsenceProof(state IdentityState, key string) (IdentityAbsenceProof, error) {
	leaves, err := IdentityStateLeaves(state)
	if err != nil {
		return IdentityAbsenceProof{}, err
	}
	index := sort.Search(len(leaves), func(i int) bool { return leaves[i].Key >= key })
	if index < len(leaves) && leaves[index].Key == key {
		return IdentityAbsenceProof{}, fmt.Errorf("identity proof key %q exists", key)
	}
	proof := IdentityAbsenceProof{RootHash: identityMerkleRoot(leaves), Key: key, LeafCount: len(leaves)}
	if index > 0 {
		previous := buildIdentityProofFromLeaves(leaves, index-1)
		proof.Previous = &previous
	}
	if index < len(leaves) {
		next := buildIdentityProofFromLeaves(leaves, index)
		proof.Next = &next
	}
	return proof, nil
}

func VerifyIdentityProof(proof IdentityInclusionProof) error {
	if proof.Key == "" {
		return errors.New("identity proof key is required")
	}
	if proof.LeafCount <= 0 {
		return errors.New("identity proof leaf count must be positive")
	}
	if proof.Index < 0 || proof.Index >= proof.LeafCount {
		return errors.New("identity proof index out of range")
	}
	if err := validateHexHash("identity proof root", proof.RootHash); err != nil {
		return err
	}
	if err := validateHexHash("identity proof value hash", proof.ValueHash); err != nil {
		return err
	}
	expectedLeaf := identityLeaf(proof.Key, proof.ValueHash)
	if proof.LeafHash != expectedLeaf.LeafHash {
		return errors.New("identity proof leaf hash mismatch")
	}
	hash := proof.LeafHash
	for _, step := range proof.Steps {
		if err := validateHexHash("identity proof sibling", step.Hash); err != nil {
			return err
		}
		if step.SiblingOnLeft {
			hash = identityNodeHash(step.Hash, hash)
		} else {
			hash = identityNodeHash(hash, step.Hash)
		}
	}
	if hash != proof.RootHash {
		return errors.New("identity proof root mismatch")
	}
	return nil
}

func VerifyIdentityAbsenceProof(proof IdentityAbsenceProof) error {
	if proof.Key == "" {
		return errors.New("identity absence proof key is required")
	}
	if proof.LeafCount <= 0 {
		return errors.New("identity absence proof leaf count must be positive")
	}
	if err := validateHexHash("identity absence proof root", proof.RootHash); err != nil {
		return err
	}
	if proof.Previous == nil && proof.Next == nil {
		return errors.New("identity absence proof requires a neighbor")
	}
	if proof.Previous != nil {
		if err := VerifyIdentityProof(*proof.Previous); err != nil {
			return err
		}
		if proof.Previous.RootHash != proof.RootHash || proof.Previous.LeafCount != proof.LeafCount {
			return errors.New("identity absence previous proof root mismatch")
		}
		if proof.Previous.Key >= proof.Key {
			return errors.New("identity absence previous key is not before target")
		}
	}
	if proof.Next != nil {
		if err := VerifyIdentityProof(*proof.Next); err != nil {
			return err
		}
		if proof.Next.RootHash != proof.RootHash || proof.Next.LeafCount != proof.LeafCount {
			return errors.New("identity absence next proof root mismatch")
		}
		if proof.Next.Key <= proof.Key {
			return errors.New("identity absence next key is not after target")
		}
	}
	if proof.Previous != nil && proof.Next != nil && proof.Previous.Index+1 != proof.Next.Index {
		return errors.New("identity absence neighbor proofs are not adjacent")
	}
	return nil
}

func BuildIdentityResolutionProof(state IdentityState, name string, height uint64) (IdentityResolutionProof, error) {
	resolution, err := ResolveIdentityRecordRecursive(state, name, height)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	stateRoot, err := IdentityStateRoot(state)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	domainKey, err := IdentityDomainStoreKey(resolution.AuthorityDomain.Name)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	resolverKey, err := IdentityResolverStoreKey(resolution.ResolverDomain)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	nftKey, err := IdentityNFTStoreKey(resolution.AuthorityDomain.NFTID)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	nft, found := findDomainNFTByID(state, resolution.AuthorityDomain.NFTID)
	if !found {
		return IdentityResolutionProof{}, errors.New("identity resolution proof nft not found")
	}
	domainProof, err := BuildIdentityProof(state, domainKey)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	resolverProof, err := BuildIdentityProof(state, resolverKey)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	nftProof, err := BuildIdentityProof(state, nftKey)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	candidates, err := resolverDomainCandidates(resolution.QueryDomain)
	if err != nil {
		return IdentityResolutionProof{}, err
	}
	proof := IdentityResolutionProof{
		StateRoot:		stateRoot,
		QueryDomain:		resolution.QueryDomain,
		ResolverDomain:		resolution.ResolverDomain,
		AuthorityDomain:	resolution.AuthorityDomain,
		AuthorityNFT:		nft,
		Resolver:		resolution.Record,
		DomainProof:		domainProof,
		NFTProof:		nftProof,
		ResolverProof:		resolverProof,
	}
	for _, candidate := range candidates {
		candidateProof, err := buildResolutionCandidateProof(state, candidate)
		if err != nil {
			return IdentityResolutionProof{}, err
		}
		proof.Candidates = append(proof.Candidates, candidateProof)
		if candidate == resolution.ResolverDomain {
			break
		}
	}
	return proof, nil
}

func VerifyIdentityResolutionProof(proof IdentityResolutionProof, height uint64) (IdentityResolution, error) {
	if proof.StateRoot == "" {
		return IdentityResolution{}, errors.New("identity resolution proof root is required")
	}
	for _, inclusion := range []IdentityInclusionProof{proof.DomainProof, proof.NFTProof, proof.ResolverProof} {
		if err := VerifyIdentityProof(inclusion); err != nil {
			return IdentityResolution{}, err
		}
		if inclusion.RootHash != proof.StateRoot {
			return IdentityResolution{}, errors.New("identity resolution proof root mismatch")
		}
	}
	expectedDomainLeaf, err := identityDomainLeaf(proof.AuthorityDomain)
	if err != nil {
		return IdentityResolution{}, err
	}
	if proof.DomainProof.Key != expectedDomainLeaf.Key || proof.DomainProof.ValueHash != expectedDomainLeaf.ValueHash {
		return IdentityResolution{}, errors.New("identity resolution domain proof value mismatch")
	}
	expectedNFTLeaf, err := identityNFTLeaf(proof.AuthorityNFT)
	if err != nil {
		return IdentityResolution{}, err
	}
	if proof.NFTProof.Key != expectedNFTLeaf.Key || proof.NFTProof.ValueHash != expectedNFTLeaf.ValueHash {
		return IdentityResolution{}, errors.New("identity resolution nft proof value mismatch")
	}
	expectedResolverLeaf, err := identityResolverLeaf(proof.Resolver)
	if err != nil {
		return IdentityResolution{}, err
	}
	if proof.ResolverProof.Key != expectedResolverLeaf.Key || proof.ResolverProof.ValueHash != expectedResolverLeaf.ValueHash {
		return IdentityResolution{}, errors.New("identity resolution resolver proof value mismatch")
	}
	if proof.AuthorityNFT.ID != proof.AuthorityDomain.NFTID || proof.AuthorityNFT.Domain != proof.AuthorityDomain.Name || !strings.EqualFold(hex.EncodeToString(proof.AuthorityNFT.Owner), hex.EncodeToString(proof.AuthorityDomain.Owner)) {
		return IdentityResolution{}, errors.New("identity resolution nft ownership mismatch")
	}
	if proof.Resolver.Domain != proof.ResolverDomain {
		return IdentityResolution{}, errors.New("identity resolution resolver domain mismatch")
	}
	if err := ValidateResolverRecordForDomain(proof.Resolver, authorityDomainRecord(proof.AuthorityDomain), int64(height)); err != nil {
		return IdentityResolution{}, err
	}
	if err := verifyResolutionCandidateProofs(proof); err != nil {
		return IdentityResolution{}, err
	}
	depth, err := resolutionDepth(proof.QueryDomain, proof.ResolverDomain)
	if err != nil {
		return IdentityResolution{}, err
	}
	return IdentityResolution{
		QueryDomain:		proof.QueryDomain,
		ResolverDomain:		proof.ResolverDomain,
		AuthorityDomain:	cloneDomain(proof.AuthorityDomain),
		Record:			cloneResolver(proof.Resolver),
		Depth:			depth,
	}, nil
}

func buildResolutionCandidateProof(state IdentityState, domain string) (IdentityResolutionCandidateProof, error) {
	domainKey, err := IdentityDomainStoreKey(domain)
	if err != nil {
		return IdentityResolutionCandidateProof{}, err
	}
	resolverKey, err := IdentityResolverStoreKey(domain)
	if err != nil {
		return IdentityResolutionCandidateProof{}, err
	}
	out := IdentityResolutionCandidateProof{Domain: domain}
	if proof, err := BuildIdentityProof(state, domainKey); err == nil {
		out.DomainProof = &proof
	} else {
		absence, absenceErr := BuildIdentityAbsenceProof(state, domainKey)
		if absenceErr != nil {
			return IdentityResolutionCandidateProof{}, absenceErr
		}
		out.DomainAbsence = &absence
	}
	if proof, err := BuildIdentityProof(state, resolverKey); err == nil {
		out.ResolverProof = &proof
	} else {
		absence, absenceErr := BuildIdentityAbsenceProof(state, resolverKey)
		if absenceErr != nil {
			return IdentityResolutionCandidateProof{}, absenceErr
		}
		out.ResolverAbsence = &absence
	}
	return out, nil
}

func verifyResolutionCandidateProofs(proof IdentityResolutionProof) error {
	if len(proof.Candidates) == 0 {
		return errors.New("identity resolution proof candidates are required")
	}
	for i, candidate := range proof.Candidates {
		if candidate.Domain == "" {
			return errors.New("identity resolution proof candidate domain is required")
		}
		expectedDomainKey, err := IdentityDomainStoreKey(candidate.Domain)
		if err != nil {
			return err
		}
		expectedResolverKey, err := IdentityResolverStoreKey(candidate.Domain)
		if err != nil {
			return err
		}
		if candidate.DomainProof != nil {
			if err := VerifyIdentityProof(*candidate.DomainProof); err != nil {
				return err
			}
			if candidate.DomainProof.Key != expectedDomainKey {
				return errors.New("identity resolution candidate domain proof key mismatch")
			}
			if candidate.DomainProof.RootHash != proof.StateRoot {
				return errors.New("identity resolution candidate domain proof root mismatch")
			}
		}
		if candidate.DomainAbsence != nil {
			if candidate.DomainAbsence.Key != expectedDomainKey {
				return errors.New("identity resolution candidate domain absence key mismatch")
			}
			if err := VerifyIdentityAbsenceProof(*candidate.DomainAbsence); err != nil {
				return err
			}
			if candidate.DomainAbsence.RootHash != proof.StateRoot {
				return errors.New("identity resolution candidate domain absence root mismatch")
			}
		}
		if candidate.ResolverProof != nil {
			if err := VerifyIdentityProof(*candidate.ResolverProof); err != nil {
				return err
			}
			if candidate.ResolverProof.Key != expectedResolverKey {
				return errors.New("identity resolution candidate resolver proof key mismatch")
			}
			if candidate.ResolverProof.RootHash != proof.StateRoot {
				return errors.New("identity resolution candidate resolver proof root mismatch")
			}
		}
		if candidate.ResolverAbsence != nil {
			if candidate.ResolverAbsence.Key != expectedResolverKey {
				return errors.New("identity resolution candidate resolver absence key mismatch")
			}
			if err := VerifyIdentityAbsenceProof(*candidate.ResolverAbsence); err != nil {
				return err
			}
			if candidate.ResolverAbsence.RootHash != proof.StateRoot {
				return errors.New("identity resolution candidate resolver absence root mismatch")
			}
		}
		if candidate.Domain == proof.ResolverDomain {
			if candidate.ResolverProof == nil {
				return errors.New("identity resolution resolver candidate proof is missing")
			}
			if i != len(proof.Candidates)-1 {
				return errors.New("identity resolution candidates must stop at resolver")
			}
			return nil
		}
		if candidate.DomainProof != nil || candidate.ResolverProof != nil {
			return errors.New("identity resolution proof crossed a closer authority or resolver")
		}
	}
	return errors.New("identity resolution resolver candidate not proven")
}

func buildIdentityProofFromLeaves(leaves []IdentityProofLeaf, index int) IdentityInclusionProof {
	proof := IdentityInclusionProof{
		RootHash:	identityMerkleRoot(leaves),
		Key:		leaves[index].Key,
		ValueHash:	leaves[index].ValueHash,
		LeafHash:	leaves[index].LeafHash,
		Index:		index,
		LeafCount:	len(leaves),
	}
	hashes := make([]string, len(leaves))
	for i, leaf := range leaves {
		hashes[i] = leaf.LeafHash
	}
	position := index
	for len(hashes) > 1 {
		sibling := position ^ 1
		if sibling >= len(hashes) {
			sibling = position
		}
		proof.Steps = append(proof.Steps, IdentityProofStep{Hash: hashes[sibling], SiblingOnLeft: sibling < position})
		next := make([]string, 0, (len(hashes)+1)/2)
		for i := 0; i < len(hashes); i += 2 {
			right := i + 1
			if right >= len(hashes) {
				right = i
			}
			next = append(next, identityNodeHash(hashes[i], hashes[right]))
		}
		hashes = next
		position /= 2
	}
	return proof
}

func identityMerkleRoot(leaves []IdentityProofLeaf) string {
	if len(leaves) == 0 {
		return identityHash("identity-root-empty")
	}
	hashes := make([]string, len(leaves))
	for i, leaf := range leaves {
		hashes[i] = leaf.LeafHash
	}
	for len(hashes) > 1 {
		next := make([]string, 0, (len(hashes)+1)/2)
		for i := 0; i < len(hashes); i += 2 {
			right := i + 1
			if right >= len(hashes) {
				right = i
			}
			next = append(next, identityNodeHash(hashes[i], hashes[right]))
		}
		hashes = next
	}
	return hashes[0]
}

func identityLeaf(key string, valueHash string) IdentityProofLeaf {
	return IdentityProofLeaf{Key: key, ValueHash: valueHash, LeafHash: identityHash("identity-leaf", key, valueHash)}
}

func identityNodeHash(left string, right string) string {
	return identityHash("identity-node", left, right)
}

func sortIdentityProofLeaves(leaves []IdentityProofLeaf) {
	sort.SliceStable(leaves, func(i, j int) bool { return leaves[i].Key < leaves[j].Key })
}

func identityDomainLeaf(domain Domain) (IdentityProofLeaf, error) {
	key, err := IdentityDomainStoreKey(domain.Name)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	return identityLeaf(key, identityHash(
		"domain",
		domain.Name,
		hex.EncodeToString(domain.Owner),
		domain.NFTID,
		fmt.Sprintf("%020d", domain.RegisteredHeight),
		fmt.Sprintf("%020d", domain.ExpiryHeight),
		fmt.Sprintf("%020d", domain.UpdatedHeight),
		domain.ParentName,
		fmt.Sprintf("%t", domain.ParentControlsRecord),
	)), nil
}

func identityNFTLeaf(nft DomainNFT) (IdentityProofLeaf, error) {
	key, err := IdentityNFTStoreKey(nft.ID)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	return identityLeaf(key, identityHash(
		"nft",
		nft.ID,
		nft.Domain,
		hex.EncodeToString(nft.Owner),
		fmt.Sprintf("%020d", nft.MintHeight),
		fmt.Sprintf("%020d", nft.TransferHeight),
	)), nil
}

func identityCommitLeaf(commit DomainCommit) (IdentityProofLeaf, error) {
	key, err := IdentityCommitStoreKey(commit.Name, commit.Owner, commit.CommitmentHash)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	return identityLeaf(key, identityHash(
		"commit",
		commit.Name,
		hex.EncodeToString(commit.Owner),
		commit.CommitmentHash,
		fmt.Sprintf("%020d", commit.CommitHeight),
		fmt.Sprintf("%020d", commit.ExpiresHeight),
	)), nil
}

func identityUsedCommitmentLeaf(commitment UsedDomainCommitment) (IdentityProofLeaf, error) {
	key := IdentityStoreV2Prefix + "/used_commitments/" + commitment.CommitmentHash
	return identityLeaf(key, identityHash(
		"used-commitment",
		commitment.CommitmentHash,
		commitment.Name,
		hex.EncodeToString(commitment.Owner),
		fmt.Sprintf("%020d", commitment.RevealedHeight),
		fmt.Sprintf("%020d", commitment.ExpiresHeight),
		commitment.ChainID,
		commitment.ModuleName,
		fmt.Sprintf("%020d", commitment.ModuleVersion),
		commitment.RegistrationClass,
		commitment.MaxPrice,
	)), nil
}

func identityResolverLeaf(resolver ResolverRecord) (IdentityProofLeaf, error) {
	key, err := IdentityResolverStoreKey(resolver.Domain)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	recordParts := make([]string, 0, len(resolver.Records)*2)
	for _, recordKey := range sortedResolverKeys(resolver.Records) {
		recordParts = append(recordParts, recordKey, hex.EncodeToString(resolver.Records[recordKey]))
	}
	parts := []string{
		"resolver",
		resolver.Domain,
		hex.EncodeToString(resolver.Owner),
		hex.EncodeToString(resolver.Primary),
		hex.EncodeToString(resolver.Contract),
		resolver.ZoneEndpoint,
		hex.EncodeToString(resolver.Metadata),
		fmt.Sprintf("%020d", resolver.UpdatedAtUnix),
	}
	parts = append(parts, recordParts...)
	return identityLeaf(key, identityHash(parts...)), nil
}

func identityReverseLeaf(reverse ReverseRecord) (IdentityProofLeaf, error) {
	key, err := IdentityReverseStoreKey(reverse.Address)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	return identityLeaf(key, identityHash(
		"reverse",
		hex.EncodeToString(reverse.Address),
		reverse.Domain,
		fmt.Sprintf("%020d", reverse.UpdatedAtUnix),
	)), nil
}

func identitySubdomainLeaf(record SubdomainRecord) (IdentityProofLeaf, error) {
	key, err := IdentitySubdomainIndexKey(record.ParentName, record.Name)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	return identityLeaf(key, identityHash(
		"subdomain",
		record.ParentName,
		record.Name,
		hex.EncodeToString(record.Owner),
		fmt.Sprintf("%t", record.ParentControlsRecord),
		fmt.Sprintf("%020d", record.CreatedHeight),
		string(record.DelegationType),
		fmt.Sprintf("%t", record.Detached),
		fmt.Sprintf("%t", record.Ephemeral),
		fmt.Sprintf("%020d", record.ExpiryHeight),
		fmt.Sprintf("%020d", record.TimeLockedUntilHeight),
		fmt.Sprintf("%t", record.ParentAuthorized),
	)), nil
}

func identityAuctionLeaf(auction Auction) (IdentityProofLeaf, error) {
	key, err := IdentityAuctionStoreKey(auction.Name)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	parts := []string{
		"auction",
		auction.Name,
		fmt.Sprintf("%020d", auction.CommitStartHeight),
		fmt.Sprintf("%020d", auction.RevealStartHeight),
		fmt.Sprintf("%020d", auction.RevealEndHeight),
		string(auction.Phase),
		hex.EncodeToString(auction.Winner),
		fmt.Sprintf("%020d", auction.WinningBid),
		auction.WinningCommitment,
	}
	for _, commitment := range auction.Commitments {
		parts = append(parts, "commitment", commitment.Name, hex.EncodeToString(commitment.Bidder), commitment.CommitmentHash, fmt.Sprintf("%020d", commitment.CommitHeight))
	}
	for _, reveal := range auction.Reveals {
		parts = append(parts, "reveal", reveal.Name, hex.EncodeToString(reveal.Bidder), fmt.Sprintf("%020d", reveal.Bid), reveal.Salt, fmt.Sprintf("%020d", reveal.RevealHeight), reveal.CommitmentHash)
	}
	for _, refund := range auction.Refunds {
		parts = append(parts, "refund", refund.ReceiptID, refund.Name, hex.EncodeToString(refund.Bidder), fmt.Sprintf("%020d", refund.Amount), refund.CommitmentHash, refund.Reason)
	}
	return identityLeaf(key, identityHash(parts...)), nil
}

func identityPendingResolverLeaf(intent ResolverUpdateIntent) (IdentityProofLeaf, error) {
	key, err := IdentityPendingResolverStoreKey(intent.Domain, intent.Actor, intent.Nonce)
	if err != nil {
		return IdentityProofLeaf{}, err
	}
	return identityLeaf(key, identityHash(
		"pending-resolver",
		intent.Domain,
		hex.EncodeToString(intent.Actor),
		fmt.Sprintf("%020d", intent.Nonce),
	)), nil
}

func findDomainNFTByID(state IdentityState, id string) (DomainNFT, bool) {
	for _, nft := range state.DomainNFTs {
		if nft.ID == id {
			return cloneDomainNFT(nft), true
		}
	}
	return DomainNFT{}, false
}

func resolutionDepth(queryDomain string, resolverDomain string) (uint8, error) {
	candidates, err := resolverDomainCandidates(queryDomain)
	if err != nil {
		return 0, err
	}
	for i, candidate := range candidates {
		if candidate == resolverDomain {
			return uint8(i), nil
		}
	}
	return 0, errors.New("identity resolver domain is not in query hierarchy")
}
