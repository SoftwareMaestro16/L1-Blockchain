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
	DefaultBroadcastTreeFanout   = uint32(4)
	DefaultBroadcastGossipFanout = uint32(8)
	MaxBroadcastFanout           = uint32(256)
	MaxBroadcastTTL              = uint32(128)
)

type BroadcastPayloadType string

const (
	BroadcastPayloadConsensus BroadcastPayloadType = "consensus"
	BroadcastPayloadBlock     BroadcastPayloadType = "block"
	BroadcastPayloadExecution BroadcastPayloadType = "execution"
	BroadcastPayloadService   BroadcastPayloadType = "service"
	BroadcastPayloadDiscovery BroadcastPayloadType = "discovery"
	BroadcastPayloadData      BroadcastPayloadType = "data"
	BroadcastPayloadRouting   BroadcastPayloadType = "routing"
	BroadcastPayloadStateSync BroadcastPayloadType = "state_sync"
)

type BroadcastFanoutPolicy struct {
	TreeFanout   uint32
	GossipFanout uint32
	OverlayBound bool
}

type BroadcastMessage struct {
	BroadcastID  string
	OriginNode   string
	OverlayID    string
	PayloadHash  string
	PayloadType  BroadcastPayloadType
	Height       uint64
	TTL          uint32
	Priority     uint32
	FanoutPolicy BroadcastFanoutPolicy
	Signature    []byte
}

type BroadcastDeduper struct {
	SeenKeys []string
}

type BroadcastPlan struct {
	BroadcastID   string
	OverlayID     string
	DedupKey      string
	TreeTargets   []string
	GossipTargets []string
	FallbackUsed  bool
	Priority      uint32
	TTLRemaining  uint32
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
		msg.PayloadHash,
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
	return HashParts("broadcast-dedupe", msg.BroadcastID, msg.PayloadHash)
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
	if err := msg.ValidateBasic(0); err != nil {
		return BroadcastDeduper{}, false, err
	}
	key := ComputeBroadcastDedupKey(msg)
	next := BroadcastDeduper{SeenKeys: append([]string(nil), d.SeenKeys...)}
	sortStrings(next.SeenKeys)
	if hasString(next.SeenKeys, key) {
		return next, false, nil
	}
	next.SeenKeys = append(next.SeenKeys, key)
	sortStrings(next.SeenKeys)
	return next, true, nil
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
		BroadcastID:   msg.BroadcastID,
		OverlayID:     msg.OverlayID,
		DedupKey:      ComputeBroadcastDedupKey(msg),
		TreeTargets:   treeTargets,
		GossipTargets: gossipTargets,
		FallbackUsed:  fallbackUsed || len(graph.Edges) == 0,
		Priority:      msg.Priority,
		TTLRemaining:  msg.TTL - 1,
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
