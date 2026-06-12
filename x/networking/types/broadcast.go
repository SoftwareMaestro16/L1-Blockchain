package types

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultBroadcastTreeFanout	= uint32(4)
	DefaultBroadcastGossipFanout	= uint32(8)
	MaxBroadcastFanout		= uint32(256)
	MaxBroadcastTTL			= uint32(128)
	DefaultBroadcastDedupHorizon	= uint64(128)
)

type BroadcastPayloadType string

const (
	BroadcastPayloadConsensus	BroadcastPayloadType	= "consensus"
	BroadcastPayloadBlock		BroadcastPayloadType	= "block"
	BroadcastPayloadExecution	BroadcastPayloadType	= "execution"
	BroadcastPayloadService		BroadcastPayloadType	= "service"
	BroadcastPayloadDiscovery	BroadcastPayloadType	= "discovery"
	BroadcastPayloadData		BroadcastPayloadType	= "data"
	BroadcastPayloadRouting		BroadcastPayloadType	= "routing"
	BroadcastPayloadStateSync	BroadcastPayloadType	= "state_sync"
)

type BroadcastFanoutPolicy struct {
	TreeFanout	uint32
	GossipFanout	uint32
	OverlayBound	bool
}

type BroadcastMessage struct {
	BroadcastID	string
	OriginNode	string
	OverlayID	string
	PayloadHash	string
	PayloadType	BroadcastPayloadType
	Height		uint64
	TTL		uint32
	Priority	uint32
	FanoutPolicy	BroadcastFanoutPolicy
	Signature	[]byte
}

type BroadcastDeduper struct {
	SeenKeys []string
}

type BroadcastDedupEntry struct {
	BroadcastID	string
	PayloadHash	string
	SeenHeight	uint64
}

type BroadcastFaultEvidence struct {
	BroadcastID		string
	ExpectedPayloadHash	string
	ConflictingPayloadHash	string
	PeerNodeID		string
	DetectedHeight		uint64
	EvidenceHash		string
}

type BroadcastDedupCache struct {
	Horizon	uint64
	Entries	[]BroadcastDedupEntry
	Faults	[]BroadcastFaultEvidence
}

type BroadcastPlan struct {
	BroadcastID	string
	OverlayID	string
	DedupKey	string
	TreeTargets	[]string
	GossipTargets	[]string
	FallbackUsed	bool
	Priority	uint32
	TTLRemaining	uint32
}

type BlockBroadcastHeader struct {
	BlockID				string
	Height				uint64
	ProposerNodeID			string
	HeaderHash			string
	ChunkSetRoot			string
	ProofSetRoot			string
	BlockRoot			string
	ChunkCount			uint32
	AvailabilityMetadataHash	string
}

type BlockChunkMetadata struct {
	BlockID		string
	ChunkIndex	uint32
	ChunkHash	string
	ChunkSize	uint64
}

type BlockProofSet struct {
	BlockID		string
	ProofRoot	string
	ProofHashes	[]string
}

type BlockPropagationSession struct {
	Header		BlockBroadcastHeader
	ProofSet	BlockProofSet
	ReceivedChunks	[]PayloadChunk
	VerifiedBitmap	[]bool
}

func NewBroadcastMessage(msg BroadcastMessage) (BroadcastMessage, error) {
	msg = NormalizeBroadcastMessage(msg)
	if msg.BroadcastID == "" {
		msg.BroadcastID = ComputeBroadcastID(msg)
	}
	if err := msg.ValidateBasic(0); err != nil {
		return BroadcastMessage{}, err
	}
	return msg, nil
}

func SignBroadcastMessage(msg BroadcastMessage, privateKey ed25519.PrivateKey, networkSalt []byte) (BroadcastMessage, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return BroadcastMessage{}, errors.New("networking broadcast private key must be ed25519")
	}
	if len(networkSalt) == 0 {
		return BroadcastMessage{}, errors.New("networking broadcast network salt is required")
	}
	pubKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return BroadcastMessage{}, errors.New("networking broadcast public key must be ed25519")
	}
	msg.Signature = nil
	msg = NormalizeBroadcastMessage(msg)
	if msg.OriginNode == "" {
		msg.OriginNode = ComputeNodeID(pubKey, networkSalt)
	}
	if msg.OriginNode != ComputeNodeID(pubKey, networkSalt) {
		return BroadcastMessage{}, errors.New("networking broadcast origin does not match signer")
	}
	msg.BroadcastID = ComputeBroadcastID(msg)
	payload, err := msg.SigningPayload()
	if err != nil {
		return BroadcastMessage{}, err
	}
	msg.Signature = ed25519.Sign(privateKey, payload)
	if err := VerifyBroadcastMessageSignature(msg, pubKey, networkSalt, 0); err != nil {
		return BroadcastMessage{}, err
	}
	return msg, nil
}

func NormalizeBroadcastMessage(msg BroadcastMessage) BroadcastMessage {
	msg.BroadcastID = normalizeHashText(msg.BroadcastID)
	msg.OriginNode = normalizeHashText(msg.OriginNode)
	msg.OverlayID = normalizeHashText(msg.OverlayID)
	msg.PayloadHash = normalizeHashText(msg.PayloadHash)
	msg.PayloadType = BroadcastPayloadType(strings.ToLower(strings.TrimSpace(string(msg.PayloadType))))
	msg.FanoutPolicy = NormalizeBroadcastFanoutPolicy(msg.FanoutPolicy)
	msg.Signature = cloneBytes(msg.Signature)
	return msg
}

func NormalizeBroadcastFanoutPolicy(policy BroadcastFanoutPolicy) BroadcastFanoutPolicy {
	if policy.TreeFanout == 0 {
		policy.TreeFanout = DefaultBroadcastTreeFanout
	}
	if policy.GossipFanout == 0 {
		policy.GossipFanout = DefaultBroadcastGossipFanout
	}
	return policy
}

func ComputeBroadcastID(msg BroadcastMessage) string {
	msg = NormalizeBroadcastMessage(msg)
	return HashParts(
		"broadcast-message",
		msg.OriginNode,
		msg.OverlayID,
		string(msg.PayloadType),
		fmt.Sprintf("%d", msg.Height),
		fmt.Sprintf("%d", msg.TTL),
		fmt.Sprintf("%d", msg.Priority),
		fmt.Sprintf("%d", msg.FanoutPolicy.TreeFanout),
		fmt.Sprintf("%d", msg.FanoutPolicy.GossipFanout),
		fmt.Sprintf("%t", msg.FanoutPolicy.OverlayBound),
	)
}

func ComputeBroadcastDedupKey(msg BroadcastMessage) string {
	msg = NormalizeBroadcastMessage(msg)
	return msg.BroadcastID
}

func (m BroadcastMessage) SigningPayload() ([]byte, error) {
	msg := NormalizeBroadcastMessage(m)
	msg.Signature = nil
	return json.Marshal(msg)
}

func (m BroadcastMessage) ValidateBasic(currentHeight uint64) error {
	msg := NormalizeBroadcastMessage(m)
	if err := ValidateHash("networking broadcast id", msg.BroadcastID); err != nil {
		return err
	}
	if msg.BroadcastID != ComputeBroadcastID(msg) {
		return errors.New("networking broadcast id mismatch")
	}
	if err := ValidateHash("networking broadcast origin node", msg.OriginNode); err != nil {
		return err
	}
	if err := ValidateHash("networking broadcast overlay id", msg.OverlayID); err != nil {
		return err
	}
	if err := ValidateHash("networking broadcast payload hash", msg.PayloadHash); err != nil {
		return err
	}
	if !IsBroadcastPayloadType(msg.PayloadType) {
		return fmt.Errorf("unknown networking broadcast payload type %q", msg.PayloadType)
	}
	if msg.TTL == 0 || msg.TTL > MaxBroadcastTTL {
		return fmt.Errorf("networking broadcast ttl must be between 1 and %d", MaxBroadcastTTL)
	}
	if msg.Priority > MaxRL2Priority {
		return fmt.Errorf("networking broadcast priority must be <= %d", MaxRL2Priority)
	}
	if err := msg.FanoutPolicy.Validate(); err != nil {
		return err
	}
	if msg.Height > 0 && currentHeight > msg.Height+uint64(msg.TTL) {
		return errors.New("networking broadcast message is expired")
	}
	return nil
}

func (p BroadcastFanoutPolicy) Validate() error {
	p = NormalizeBroadcastFanoutPolicy(p)
	if p.TreeFanout == 0 || p.TreeFanout > MaxBroadcastFanout {
		return fmt.Errorf("networking broadcast tree fanout must be between 1 and %d", MaxBroadcastFanout)
	}
	if p.GossipFanout == 0 || p.GossipFanout > MaxBroadcastFanout {
		return fmt.Errorf("networking broadcast gossip fanout must be between 1 and %d", MaxBroadcastFanout)
	}
	return nil
}

func VerifyBroadcastMessageSignature(msg BroadcastMessage, pubKey ed25519.PublicKey, networkSalt []byte, currentHeight uint64) error {
	msg = NormalizeBroadcastMessage(msg)
	if err := msg.ValidateBasic(currentHeight); err != nil {
		return err
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking broadcast public key must be %d bytes", ed25519.PublicKeySize)
	}
	if len(networkSalt) == 0 {
		return errors.New("networking broadcast network salt is required")
	}
	if msg.OriginNode != ComputeNodeID(pubKey, networkSalt) {
		return errors.New("networking broadcast origin does not match public key")
	}
	if len(msg.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("networking broadcast signature must be %d bytes", ed25519.SignatureSize)
	}
	payload, err := msg.SigningPayload()
	if err != nil {
		return err
	}
	if !ed25519.Verify(pubKey, payload, msg.Signature) {
		return errors.New("networking broadcast signature verification failed")
	}
	return nil
}

func (d BroadcastDeduper) Accept(msg BroadcastMessage) (BroadcastDeduper, bool, error) {
	cache := BroadcastDedupCache{Entries: make([]BroadcastDedupEntry, 0, len(d.SeenKeys))}
	for _, key := range d.SeenKeys {
		cache.Entries = append(cache.Entries, BroadcastDedupEntry{BroadcastID: normalizeHashText(key)})
	}
	nextCache, decision, err := cache.Accept(msg, "", 0)
	if err != nil {
		return BroadcastDeduper{}, false, err
	}
	keys := make([]string, 0, len(nextCache.Entries))
	for _, entry := range nextCache.Entries {
		keys = append(keys, entry.BroadcastID)
	}
	sortStrings(keys)
	return BroadcastDeduper{SeenKeys: keys}, decision.Accepted, nil
}

func PlanBroadcastForwarding(msg BroadcastMessage, desc OverlayDescriptor, graph RoutingGraph, localNodeID string, candidateNodeIDs []string, deduper BroadcastDeduper, currentHeight uint64) (BroadcastDeduper, BroadcastPlan, error) {
	msg = NormalizeBroadcastMessage(msg)
	if err := msg.ValidateBasic(currentHeight); err != nil {
		return BroadcastDeduper{}, BroadcastPlan{}, err
	}
	desc = NormalizeOverlayDescriptor(desc)
	if err := desc.ValidateBasic(); err != nil {
		return BroadcastDeduper{}, BroadcastPlan{}, err
	}
	if msg.OverlayID != desc.OverlayID {
		return BroadcastDeduper{}, BroadcastPlan{}, errors.New("networking broadcast overlay mismatch")
	}
	localNodeID = normalizeHashText(localNodeID)
	if err := ValidateHash("networking broadcast local node", localNodeID); err != nil {
		return BroadcastDeduper{}, BroadcastPlan{}, err
	}
	nextDeduper, accepted, err := deduper.Accept(msg)
	if err != nil {
		return BroadcastDeduper{}, BroadcastPlan{}, err
	}
	if !accepted {
		return nextDeduper, BroadcastPlan{BroadcastID: msg.BroadcastID, OverlayID: msg.OverlayID, DedupKey: ComputeBroadcastDedupKey(msg), Priority: msg.Priority}, nil
	}
	candidates, err := normalizeBroadcastCandidates(candidateNodeIDs, localNodeID, msg.OriginNode)
	if err != nil {
		return BroadcastDeduper{}, BroadcastPlan{}, err
	}
	graph = NormalizeRoutingGraph(graph)
	treeFanout := clampBroadcastFanout(msg.FanoutPolicy.TreeFanout, desc.Fanout)
	gossipFanout := clampBroadcastFanout(msg.FanoutPolicy.GossipFanout, desc.Fanout)
	treeTargets := selectBroadcastTreeTargets(graph, localNodeID, candidates, treeFanout)
	remaining := excludeBroadcastTargets(candidates, treeTargets)
	fallbackUsed := len(treeTargets) < int(treeFanout)
	gossipTargets := selectBroadcastGossipTargets(msg, remaining, gossipFanout)
	return nextDeduper, BroadcastPlan{
		BroadcastID:	msg.BroadcastID,
		OverlayID:	msg.OverlayID,
		DedupKey:	ComputeBroadcastDedupKey(msg),
		TreeTargets:	treeTargets,
		GossipTargets:	gossipTargets,
		FallbackUsed:	fallbackUsed || len(graph.Edges) == 0,
		Priority:	msg.Priority,
		TTLRemaining:	msg.TTL - 1,
	}, nil
}

func SortBroadcastMessages(messages []BroadcastMessage) []BroadcastMessage {
	out := make([]BroadcastMessage, len(messages))
	for i, msg := range messages {
		out[i] = NormalizeBroadcastMessage(msg)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].BroadcastID < out[j].BroadcastID
	})
	return out
}

func IsBroadcastPayloadType(payloadType BroadcastPayloadType) bool {
	switch payloadType {
	case BroadcastPayloadConsensus,
		BroadcastPayloadBlock,
		BroadcastPayloadExecution,
		BroadcastPayloadService,
		BroadcastPayloadDiscovery,
		BroadcastPayloadData,
		BroadcastPayloadRouting,
		BroadcastPayloadStateSync:
		return true
	default:
		return false
	}
}

func NewBroadcastDedupCache(horizon uint64) BroadcastDedupCache {
	if horizon == 0 {
		horizon = DefaultBroadcastDedupHorizon
	}
	return BroadcastDedupCache{Horizon: horizon}
}

type BroadcastDedupDecision struct {
	Accepted		bool
	DroppedDuplicate	bool
	FaultEvidence		BroadcastFaultEvidence
}

func (c BroadcastDedupCache) Accept(msg BroadcastMessage, peerNodeID string, currentHeight uint64) (BroadcastDedupCache, BroadcastDedupDecision, error) {
	msg = NormalizeBroadcastMessage(msg)
	if err := msg.ValidateBasic(currentHeight); err != nil {
		return BroadcastDedupCache{}, BroadcastDedupDecision{}, err
	}
	next := c.Prune(currentHeight)
	if next.Horizon == 0 {
		next.Horizon = DefaultBroadcastDedupHorizon
	}
	peerNodeID = normalizeHashText(peerNodeID)
	if peerNodeID != "" {
		if err := ValidateHash("networking broadcast peer node", peerNodeID); err != nil {
			return BroadcastDedupCache{}, BroadcastDedupDecision{}, err
		}
	}
	for _, entry := range next.Entries {
		if entry.BroadcastID != msg.BroadcastID {
			continue
		}
		if entry.PayloadHash == "" || entry.PayloadHash == msg.PayloadHash {
			return next, BroadcastDedupDecision{DroppedDuplicate: true}, nil
		}
		evidence := NewBroadcastFaultEvidence(msg.BroadcastID, entry.PayloadHash, msg.PayloadHash, peerNodeID, currentHeight)
		next.Faults = append(next.Faults, evidence)
		sortBroadcastFaults(next.Faults)
		return next, BroadcastDedupDecision{FaultEvidence: evidence}, nil
	}
	next.Entries = append(next.Entries, BroadcastDedupEntry{
		BroadcastID:	msg.BroadcastID,
		PayloadHash:	msg.PayloadHash,
		SeenHeight:	currentHeight,
	})
	sortBroadcastDedupEntries(next.Entries)
	return next, BroadcastDedupDecision{Accepted: true}, nil
}

func (c BroadcastDedupCache) Prune(currentHeight uint64) BroadcastDedupCache {
	next := BroadcastDedupCache{
		Horizon:	c.Horizon,
		Faults:		append([]BroadcastFaultEvidence(nil), c.Faults...),
	}
	if next.Horizon == 0 {
		next.Horizon = DefaultBroadcastDedupHorizon
	}
	for _, entry := range c.Entries {
		entry.BroadcastID = normalizeHashText(entry.BroadcastID)
		entry.PayloadHash = normalizeHashText(entry.PayloadHash)
		if currentHeight > 0 && entry.SeenHeight > 0 && currentHeight > entry.SeenHeight+next.Horizon {
			continue
		}
		next.Entries = append(next.Entries, entry)
	}
	sortBroadcastDedupEntries(next.Entries)
	sortBroadcastFaults(next.Faults)
	return next
}

func NewBroadcastFaultEvidence(broadcastID, expectedPayloadHash, conflictingPayloadHash, peerNodeID string, detectedHeight uint64) BroadcastFaultEvidence {
	evidence := BroadcastFaultEvidence{
		BroadcastID:		normalizeHashText(broadcastID),
		ExpectedPayloadHash:	normalizeHashText(expectedPayloadHash),
		ConflictingPayloadHash:	normalizeHashText(conflictingPayloadHash),
		PeerNodeID:		normalizeHashText(peerNodeID),
		DetectedHeight:		detectedHeight,
	}
	evidence.EvidenceHash = HashParts(
		"broadcast-fault",
		evidence.BroadcastID,
		evidence.ExpectedPayloadHash,
		evidence.ConflictingPayloadHash,
		evidence.PeerNodeID,
		fmt.Sprintf("%d", evidence.DetectedHeight),
	)
	return evidence
}

func NewBlockBroadcastHeader(header BlockBroadcastHeader) (BlockBroadcastHeader, error) {
	header = NormalizeBlockBroadcastHeader(header)
	if header.BlockID == "" {
		header.BlockID = ComputeBlockBroadcastID(header)
	}
	if err := header.Validate(0); err != nil {
		return BlockBroadcastHeader{}, err
	}
	return header, nil
}

func NormalizeBlockBroadcastHeader(header BlockBroadcastHeader) BlockBroadcastHeader {
	header.BlockID = normalizeHashText(header.BlockID)
	header.ProposerNodeID = normalizeHashText(header.ProposerNodeID)
	header.HeaderHash = normalizeHashText(header.HeaderHash)
	header.ChunkSetRoot = normalizeHashText(header.ChunkSetRoot)
	header.ProofSetRoot = normalizeHashText(header.ProofSetRoot)
	header.BlockRoot = normalizeHashText(header.BlockRoot)
	header.AvailabilityMetadataHash = normalizeHashText(header.AvailabilityMetadataHash)
	return header
}

func ComputeBlockBroadcastID(header BlockBroadcastHeader) string {
	header = NormalizeBlockBroadcastHeader(header)
	return HashParts(
		"block-broadcast",
		fmt.Sprintf("%d", header.Height),
		header.ProposerNodeID,
		header.HeaderHash,
		header.ChunkSetRoot,
		header.ProofSetRoot,
		header.BlockRoot,
		fmt.Sprintf("%d", header.ChunkCount),
		header.AvailabilityMetadataHash,
	)
}

func ComputeBlockRoot(headerHash, chunkSetRoot, proofSetRoot string) string {
	return HashParts("block-root", normalizeHashText(headerHash), normalizeHashText(chunkSetRoot), normalizeHashText(proofSetRoot))
}

func (h BlockBroadcastHeader) Validate(currentHeight uint64) error {
	header := NormalizeBlockBroadcastHeader(h)
	if err := ValidateHash("networking block broadcast id", header.BlockID); err != nil {
		return err
	}
	if header.BlockID != ComputeBlockBroadcastID(header) {
		return errors.New("networking block broadcast id mismatch")
	}
	if header.Height == 0 {
		return errors.New("networking block broadcast height must be positive")
	}
	if currentHeight > 0 && header.Height > currentHeight+1 {
		return errors.New("networking block broadcast height is too far ahead")
	}
	if err := ValidateHash("networking block proposer", header.ProposerNodeID); err != nil {
		return err
	}
	if err := ValidateHash("networking block header hash", header.HeaderHash); err != nil {
		return err
	}
	if err := ValidateHash("networking block chunk root", header.ChunkSetRoot); err != nil {
		return err
	}
	if err := ValidateHash("networking block proof root", header.ProofSetRoot); err != nil {
		return err
	}
	if err := ValidateHash("networking block root", header.BlockRoot); err != nil {
		return err
	}
	if header.BlockRoot != ComputeBlockRoot(header.HeaderHash, header.ChunkSetRoot, header.ProofSetRoot) {
		return errors.New("networking block root mismatch")
	}
	if header.ChunkCount == 0 || header.ChunkCount > MaxPayloadChunks {
		return fmt.Errorf("networking block chunk count must be between 1 and %d", MaxPayloadChunks)
	}
	if err := ValidateHash("networking block availability metadata hash", header.AvailabilityMetadataHash); err != nil {
		return err
	}
	return nil
}

func NewBlockChunkMetadata(blockID string, chunks []PayloadChunk) ([]BlockChunkMetadata, string, error) {
	if err := ValidateHash("networking block metadata id", normalizeHashText(blockID)); err != nil {
		return nil, "", err
	}
	ordered, err := orderedRL2Chunks(chunks)
	if err != nil {
		return nil, "", err
	}
	metadata := make([]BlockChunkMetadata, len(ordered))
	hashes := make([]string, len(ordered))
	for i, chunk := range ordered {
		metadata[i] = BlockChunkMetadata{
			BlockID:	normalizeHashText(blockID),
			ChunkIndex:	chunk.Index,
			ChunkHash:	chunk.ChunkHash,
			ChunkSize:	uint64(len(chunk.Bytes)),
		}
		hashes[i] = chunk.ChunkHash
	}
	root, err := computeRL2ChunkRootFromHashes(hashes)
	if err != nil {
		return nil, "", err
	}
	return metadata, root, nil
}

func (m BlockChunkMetadata) Validate(header BlockBroadcastHeader) error {
	meta := m
	meta.BlockID = normalizeHashText(meta.BlockID)
	meta.ChunkHash = normalizeHashText(meta.ChunkHash)
	header = NormalizeBlockBroadcastHeader(header)
	if err := header.Validate(0); err != nil {
		return err
	}
	if meta.BlockID != header.BlockID {
		return errors.New("networking block chunk metadata block mismatch")
	}
	if meta.ChunkIndex >= header.ChunkCount {
		return errors.New("networking block chunk metadata index out of range")
	}
	if err := ValidateHash("networking block chunk hash", meta.ChunkHash); err != nil {
		return err
	}
	if meta.ChunkSize == 0 || meta.ChunkSize > MaxChunkBytes {
		return fmt.Errorf("networking block chunk size must be between 1 and %d", MaxChunkBytes)
	}
	return nil
}

func NewBlockProofSet(blockID string, proofHashes []string) (BlockProofSet, error) {
	blockID = normalizeHashText(blockID)
	if err := ValidateHash("networking block proof set block id", blockID); err != nil {
		return BlockProofSet{}, err
	}
	proofs := normalizeHashPath(proofHashes)
	if len(proofs) == 0 {
		return BlockProofSet{}, errors.New("networking block proof set requires proofs")
	}
	for _, proofHash := range proofs {
		if err := ValidateHash("networking block proof hash", proofHash); err != nil {
			return BlockProofSet{}, err
		}
	}
	return BlockProofSet{
		BlockID:	blockID,
		ProofRoot:	HashParts(append([]string{"block-proof-set"}, proofs...)...),
		ProofHashes:	proofs,
	}, nil
}

func (p BlockProofSet) Validate(header BlockBroadcastHeader) error {
	proof := p
	proof.BlockID = normalizeHashText(proof.BlockID)
	proof.ProofRoot = normalizeHashText(proof.ProofRoot)
	proof.ProofHashes = normalizeHashPath(proof.ProofHashes)
	header = NormalizeBlockBroadcastHeader(header)
	if proof.BlockID != header.BlockID {
		return errors.New("networking block proof set block mismatch")
	}
	if err := ValidateHash("networking block proof root", proof.ProofRoot); err != nil {
		return err
	}
	expected := HashParts(append([]string{"block-proof-set"}, proof.ProofHashes...)...)
	if proof.ProofRoot != expected || proof.ProofRoot != header.ProofSetRoot {
		return errors.New("networking block proof set root mismatch")
	}
	return nil
}

func StartBlockPropagation(header BlockBroadcastHeader, proofSet BlockProofSet, proposer NodeRecord, currentHeight uint64) (BlockPropagationSession, error) {
	header = NormalizeBlockBroadcastHeader(header)
	proposer = NormalizeNodeRecord(proposer)
	if err := header.Validate(currentHeight); err != nil {
		return BlockPropagationSession{}, err
	}
	if proposer.NodeID != header.ProposerNodeID {
		return BlockPropagationSession{}, errors.New("networking block proposer context mismatch")
	}
	if !hasRole(proposer.Roles, NodeRoleValidator) {
		return BlockPropagationSession{}, errors.New("networking block proposer must be validator")
	}
	if err := proofSet.Validate(header); err != nil {
		return BlockPropagationSession{}, err
	}
	return BlockPropagationSession{
		Header:		header,
		ProofSet:	proofSet,
		VerifiedBitmap:	make([]bool, header.ChunkCount),
	}, nil
}

func AcceptBlockChunk(session BlockPropagationSession, metadata BlockChunkMetadata, chunk PayloadChunk) (BlockPropagationSession, error) {
	if err := metadata.Validate(session.Header); err != nil {
		return BlockPropagationSession{}, err
	}
	if err := chunk.Validate(); err != nil {
		return BlockPropagationSession{}, err
	}
	if chunk.Index != metadata.ChunkIndex || chunk.ChunkHash != metadata.ChunkHash || uint64(len(chunk.Bytes)) != metadata.ChunkSize {
		return BlockPropagationSession{}, errors.New("networking block chunk metadata mismatch")
	}
	if session.VerifiedBitmap == nil {
		session.VerifiedBitmap = make([]bool, session.Header.ChunkCount)
	}
	if !session.VerifiedBitmap[chunk.Index] {
		session.ReceivedChunks = append(session.ReceivedChunks, chunk)
		session.VerifiedBitmap[chunk.Index] = true
	}
	sort.SliceStable(session.ReceivedChunks, func(i, j int) bool {
		return session.ReceivedChunks[i].Index < session.ReceivedChunks[j].Index
	})
	return session, nil
}

func ReconstructBlock(session BlockPropagationSession, metadata []BlockChunkMetadata) ([]byte, error) {
	if err := session.Header.Validate(0); err != nil {
		return nil, err
	}
	if err := session.ProofSet.Validate(session.Header); err != nil {
		return nil, err
	}
	if uint32(len(session.ReceivedChunks)) != session.Header.ChunkCount {
		return nil, errors.New("networking block reconstruction requires all chunks")
	}
	if uint32(len(metadata)) != session.Header.ChunkCount {
		return nil, errors.New("networking block reconstruction metadata count mismatch")
	}
	orderedMeta := append([]BlockChunkMetadata(nil), metadata...)
	sort.SliceStable(orderedMeta, func(i, j int) bool {
		return orderedMeta[i].ChunkIndex < orderedMeta[j].ChunkIndex
	})
	chunkHashes := make([]string, len(orderedMeta))
	for i, meta := range orderedMeta {
		if meta.ChunkIndex != uint32(i) {
			return nil, errors.New("networking block chunk metadata sequence gap")
		}
		if err := meta.Validate(session.Header); err != nil {
			return nil, err
		}
		chunkHashes[i] = normalizeHashText(meta.ChunkHash)
	}
	root, err := computeRL2ChunkRootFromHashes(chunkHashes)
	if err != nil {
		return nil, err
	}
	if root != session.Header.ChunkSetRoot {
		return nil, errors.New("networking block chunk root mismatch")
	}
	return ReassemblePayload(session.ReceivedChunks)
}

func clampBroadcastFanout(requested, overlayFanout uint32) uint32 {
	if requested == 0 {
		requested = DefaultBroadcastGossipFanout
	}
	if overlayFanout > 0 && requested > overlayFanout {
		return overlayFanout
	}
	return requested
}

func normalizeBroadcastCandidates(candidateNodeIDs []string, localNodeID, originNodeID string) ([]string, error) {
	out := make([]string, 0, len(candidateNodeIDs))
	seen := make(map[string]struct{}, len(candidateNodeIDs))
	for _, id := range candidateNodeIDs {
		id = normalizeHashText(id)
		if id == "" || id == localNodeID || id == originNodeID {
			continue
		}
		if err := ValidateHash("networking broadcast candidate node", id); err != nil {
			return nil, err
		}
		if _, found := seen[id]; found {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sortStrings(out)
	return out, nil
}

func selectBroadcastTreeTargets(graph RoutingGraph, localNodeID string, candidates []string, fanout uint32) []string {
	if fanout == 0 || len(candidates) == 0 {
		return nil
	}
	candidateSet := make(map[string]struct{}, len(candidates))
	for _, id := range candidates {
		candidateSet[id] = struct{}{}
	}
	edges := make([]RoutingEdge, 0)
	for _, edge := range graph.Edges {
		if edge.FromNodeID != localNodeID {
			continue
		}
		if _, found := candidateSet[edge.ToNodeID]; !found {
			continue
		}
		edges = append(edges, edge)
	}
	sort.SliceStable(edges, func(i, j int) bool {
		if edges[i].Priority != edges[j].Priority {
			return edges[i].Priority < edges[j].Priority
		}
		if edges[i].LatencyMillis != edges[j].LatencyMillis {
			return edges[i].LatencyMillis < edges[j].LatencyMillis
		}
		if edges[i].Weight != edges[j].Weight {
			return edges[i].Weight > edges[j].Weight
		}
		return edges[i].ToNodeID < edges[j].ToNodeID
	})
	targets := make([]string, 0, fanout)
	for _, edge := range edges {
		targets = append(targets, edge.ToNodeID)
		if len(targets) == int(fanout) {
			break
		}
	}
	return targets
}

func selectBroadcastGossipTargets(msg BroadcastMessage, candidates []string, fanout uint32) []string {
	if fanout == 0 || len(candidates) == 0 {
		return nil
	}
	ordered := sortByHashSeed(candidates, HashParts("broadcast-gossip", msg.BroadcastID, msg.PayloadHash))
	if len(ordered) > int(fanout) {
		return ordered[:fanout]
	}
	return ordered
}

func excludeBroadcastTargets(candidates, selected []string) []string {
	selectedSet := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		selectedSet[id] = struct{}{}
	}
	out := make([]string, 0, len(candidates))
	for _, id := range candidates {
		if _, found := selectedSet[id]; found {
			continue
		}
		out = append(out, id)
	}
	return out
}

func sortBroadcastDedupEntries(entries []BroadcastDedupEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].BroadcastID < entries[j].BroadcastID
	})
}

func sortBroadcastFaults(faults []BroadcastFaultEvidence) {
	sort.SliceStable(faults, func(i, j int) bool {
		return faults[i].EvidenceHash < faults[j].EvidenceHash
	})
}
