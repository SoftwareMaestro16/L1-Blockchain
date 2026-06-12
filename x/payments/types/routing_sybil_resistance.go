package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type GossipRateLimitPolicy struct {
	WindowBlocks		uint64
	MaxMessagesPerNode	uint32
	MaxMessagesPerChannel	uint32
	MaxTopologyUpdates	uint32
	RejectPenalty		int64
}

type GossipRateLimitDecision struct {
	NodeID		string
	ChannelID	string
	WindowStart	uint64
	WindowEnd	uint64
	NodeMessages	uint32
	ChannelMessages	uint32
	TopologyUpdates	uint32
	Allowed		bool
	Reason		string
	PolicyHash	string
	DecisionHash	string
}

type RouteFailureScoringPolicy struct {
	CapacityPenalty		int64
	TimeoutPenalty		int64
	CongestionPenalty	int64
	LiquidityStalePenalty	int64
	NodeUnavailablePenalty	int64
	PolicyRejectedPenalty	int64
	UnknownPenalty		int64
	RepeatedFailurePenalty	int64
	MaxPenalty		int64
}

type RouteFailureScore struct {
	NodeID		string
	ChannelID	string
	FailureClass	RouteFailureClass
	FailureCount	uint32
	ScoreDelta	int64
	ObservedHeight	uint64
	ScoreHash	string
}

type RouteLiquidityProof struct {
	ChannelID		string
	Amount			string
	HighValueThreshold	string
	RequiredDeposit		string
	CurrentHeight		uint64
	Advertisement		LiquidityAdvertisement
	Reservation		SignedLiquidityReservation
	ProofHash		string
}

type TopologySpamSimulation struct {
	Accepted	uint32
	Rejected	uint32
	PenalizedNodes	[]string
	FinalReputation	[]GossipReputation
	Decisions	[]GossipRateLimitDecision
	SimulationHash	string
}

func DefaultGossipRateLimitPolicy() GossipRateLimitPolicy {
	return GossipRateLimitPolicy{
		WindowBlocks:		8,
		MaxMessagesPerNode:	8,
		MaxMessagesPerChannel:	12,
		MaxTopologyUpdates:	16,
		RejectPenalty:		InvalidGossipPenalty,
	}
}

func DefaultRouteFailureScoringPolicy() RouteFailureScoringPolicy {
	return RouteFailureScoringPolicy{
		CapacityPenalty:	20,
		TimeoutPenalty:		25,
		CongestionPenalty:	30,
		LiquidityStalePenalty:	20,
		NodeUnavailablePenalty:	40,
		PolicyRejectedPenalty:	25,
		UnknownPenalty:		15,
		RepeatedFailurePenalty:	10,
		MaxPenalty:		250,
	}
}

func CheckGossipRateLimit(store TopologyStore, envelope SignedGossipEnvelope, currentHeight uint64, policy GossipRateLimitPolicy) (GossipRateLimitDecision, error) {
	store = store.Normalize()
	envelope = envelope.Normalize()
	policy = policy.Normalize()
	if currentHeight == 0 {
		return GossipRateLimitDecision{}, errors.New("payments gossip rate limit height must be positive")
	}
	if err := policy.Validate(); err != nil {
		return GossipRateLimitDecision{}, err
	}
	message := envelope.Message.Normalize()
	nodeID := strings.TrimSpace(message.NodeID)
	if nodeID == "" {
		nodeID = strings.TrimSpace(envelope.Signature.Signer)
	}
	if err := addressingValidateOptionalNode("payments gossip rate limit node", nodeID); err != nil {
		return GossipRateLimitDecision{}, err
	}
	channelID := normalizeOptionalHash(message.ChannelID)
	windowStart := currentHeight - ((currentHeight - 1) % policy.WindowBlocks)
	windowEnd := windowStart + policy.WindowBlocks - 1
	decision := GossipRateLimitDecision{
		NodeID:		nodeID,
		ChannelID:	channelID,
		WindowStart:	windowStart,
		WindowEnd:	windowEnd,
		Allowed:	true,
		PolicyHash:	policy.Hash(),
	}
	for _, existing := range store.Messages {
		existing = existing.Normalize()
		if existing.ReceivedAt < windowStart || existing.ReceivedAt > windowEnd {
			continue
		}
		existingNode := strings.TrimSpace(existing.Message.NodeID)
		if existingNode == nodeID {
			decision.NodeMessages++
		}
		if channelID != "" && existing.Message.ChannelID == channelID {
			decision.ChannelMessages++
		}
		if isTopologyUpdate(existing.Message.MessageType) {
			decision.TopologyUpdates++
		}
	}
	if decision.NodeMessages >= policy.MaxMessagesPerNode {
		decision.Allowed = false
		decision.Reason = "node message rate limit"
	} else if channelID != "" && decision.ChannelMessages >= policy.MaxMessagesPerChannel {
		decision.Allowed = false
		decision.Reason = "channel message rate limit"
	} else if isTopologyUpdate(message.MessageType) && decision.TopologyUpdates >= policy.MaxTopologyUpdates {
		decision.Allowed = false
		decision.Reason = "topology update rate limit"
	}
	decision.DecisionHash = decision.Hash()
	return decision, nil
}

func ApplyGossipEnvelopeWithRateLimit(store TopologyStore, state PaymentsState, envelope SignedGossipEnvelope, currentHeight uint64, policy GossipRateLimitPolicy) (TopologyStore, GossipRateLimitDecision, error) {
	decision, err := CheckGossipRateLimit(store, envelope, currentHeight, policy)
	if err != nil {
		return TopologyStore{}, GossipRateLimitDecision{}, err
	}
	if !decision.Allowed {
		next := PenalizeInvalidGossip(store, decision.NodeID, currentHeight)
		if policy.Normalize().RejectPenalty != InvalidGossipPenalty {
			next.Reputation = addGossipReputation(next.Reputation, decision.NodeID, -(policy.Normalize().RejectPenalty - InvalidGossipPenalty), false, currentHeight)
		}
		return next.Normalize(), decision, fmt.Errorf("payments gossip rejected by %s", decision.Reason)
	}
	next, err := ApplyGossipEnvelope(store, state, envelope, currentHeight)
	return next, decision, err
}

func BuildRouteFailureScore(report RouteFailureReport, failureCount uint32, policy RouteFailureScoringPolicy) (RouteFailureScore, error) {
	report = report.Normalize()
	policy = policy.Normalize()
	if err := report.Validate(); err != nil {
		return RouteFailureScore{}, err
	}
	if err := policy.Validate(); err != nil {
		return RouteFailureScore{}, err
	}
	if failureCount == 0 {
		failureCount = 1
	}
	base := policy.penaltyFor(report.FailureClass)
	total := base + int64(failureCount-1)*policy.RepeatedFailurePenalty
	if total > policy.MaxPenalty {
		total = policy.MaxPenalty
	}
	score := RouteFailureScore{
		NodeID:		report.From,
		ChannelID:	report.ChannelID,
		FailureClass:	report.FailureClass,
		FailureCount:	failureCount,
		ScoreDelta:	-total,
		ObservedHeight:	report.ObservedHeight,
	}
	score.ScoreHash = score.Hash()
	return score, nil
}

func ApplyRouteFailureScoring(store TopologyStore, reports []RouteFailureReport, policy RouteFailureScoringPolicy) (TopologyStore, []RouteFailureScore, error) {
	store = store.Normalize()
	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		return TopologyStore{}, nil, err
	}
	counts := map[string]uint32{}
	next := store
	scores := make([]RouteFailureScore, 0, len(reports))
	for _, report := range reports {
		report = report.Normalize()
		key := routeFailureScoreKey(report)
		counts[key]++
		score, err := BuildRouteFailureScore(report, counts[key], policy)
		if err != nil {
			return TopologyStore{}, nil, err
		}
		next.Reputation = addGossipReputation(next.Reputation, score.NodeID, score.ScoreDelta, false, score.ObservedHeight)
		scores = append(scores, score)
	}
	sortRouteFailureScores(scores)
	return next.Normalize(), scores, next.Validate()
}

func BuildRouteLiquidityProof(proof RouteLiquidityProof) (RouteLiquidityProof, error) {
	proof = proof.Normalize()
	proof.ProofHash = ""
	proof.ProofHash = ComputeRouteLiquidityProofHash(proof)
	if err := proof.ValidateBasic(); err != nil {
		return RouteLiquidityProof{}, err
	}
	return proof.Normalize(), nil
}

func ComputeRouteLiquidityProofHash(proof RouteLiquidityProof) string {
	proof = proof.Normalize()
	return HashParts(
		"route-liquidity-proof",
		proof.ChannelID,
		proof.Amount,
		proof.HighValueThreshold,
		proof.RequiredDeposit,
		fmt.Sprintf("%020d", proof.CurrentHeight),
		proof.Advertisement.AdvertisementHash,
		proof.Reservation.CommitmentHash,
	)
}

func VerifyRouteLiquidityProof(state PaymentsState, proof RouteLiquidityProof) error {
	state = state.Export()
	proof = proof.Normalize()
	if err := proof.ValidateBasic(); err != nil {
		return err
	}
	amount, err := parsePositiveInt("payments route liquidity proof amount", proof.Amount)
	if err != nil {
		return err
	}
	threshold, err := parseNonNegativeInt("payments route liquidity high value threshold", proof.HighValueThreshold)
	if err != nil {
		return err
	}
	if amount.LT(threshold) {
		return nil
	}
	channel, found := state.ChannelByID(proof.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return errors.New("payments route liquidity proof requires open channel")
	}
	if proof.Advertisement.ChannelID != proof.ChannelID || proof.Reservation.ChannelID != proof.ChannelID {
		return errors.New("payments route liquidity proof channel mismatch")
	}
	if !containsString(channel.Participants, proof.Advertisement.Advertiser) || !containsString(channel.Participants, proof.Advertisement.Counterparty) {
		return errors.New("payments route liquidity proof advertisement parties must be channel participants")
	}
	if proof.CurrentHeight > proof.Advertisement.ValidUntilHeight || proof.CurrentHeight > proof.Reservation.ExpirationHeight {
		return errors.New("payments route liquidity proof expired")
	}
	if !proof.Advertisement.BackedByReservation {
		return errors.New("payments route liquidity proof requires backed advertisement")
	}
	if proof.Reservation.AdvertisementID != proof.Advertisement.AdvertisementID {
		return errors.New("payments route liquidity proof advertisement mismatch")
	}
	adCapacity, err := parsePositiveInt("payments route liquidity advertised capacity", proof.Advertisement.Capacity)
	if err != nil {
		return err
	}
	reserveCapacity, err := parsePositiveInt("payments route liquidity reserved capacity", proof.Reservation.Capacity)
	if err != nil {
		return err
	}
	if adCapacity.LT(amount) || reserveCapacity.LT(amount) {
		return errors.New("payments route liquidity proof capacity below amount")
	}
	return nil
}

func SimulateTopologySpam(state PaymentsState, initial TopologyStore, envelopes []SignedGossipEnvelope, currentHeight uint64, policy GossipRateLimitPolicy) (TopologyStore, TopologySpamSimulation, error) {
	if currentHeight == 0 {
		return TopologyStore{}, TopologySpamSimulation{}, errors.New("payments topology spam simulation height must be positive")
	}
	store := initial.Normalize()
	sim := TopologySpamSimulation{}
	penalized := map[string]struct{}{}
	for _, envelope := range envelopes {
		next, decision, err := ApplyGossipEnvelopeWithRateLimit(store, state, envelope, currentHeight, policy)
		sim.Decisions = append(sim.Decisions, decision)
		store = next.Normalize()
		if err != nil {
			sim.Rejected++
			if decision.NodeID != "" {
				penalized[decision.NodeID] = struct{}{}
			}
			continue
		}
		sim.Accepted++
	}
	for node := range penalized {
		sim.PenalizedNodes = append(sim.PenalizedNodes, node)
	}
	sortStrings(sim.PenalizedNodes)
	sim.FinalReputation = normalizeGossipReputation(store.Reputation)
	sim.Decisions = normalizeGossipRateLimitDecisions(sim.Decisions)
	sim.SimulationHash = sim.Hash()
	if err := store.Validate(); err != nil {
		return TopologyStore{}, TopologySpamSimulation{}, err
	}
	return store, sim, nil
}

func (p GossipRateLimitPolicy) Normalize() GossipRateLimitPolicy {
	defaults := DefaultGossipRateLimitPolicy()
	if p.WindowBlocks == 0 {
		p.WindowBlocks = defaults.WindowBlocks
	}
	if p.MaxMessagesPerNode == 0 {
		p.MaxMessagesPerNode = defaults.MaxMessagesPerNode
	}
	if p.MaxMessagesPerChannel == 0 {
		p.MaxMessagesPerChannel = defaults.MaxMessagesPerChannel
	}
	if p.MaxTopologyUpdates == 0 {
		p.MaxTopologyUpdates = defaults.MaxTopologyUpdates
	}
	if p.RejectPenalty == 0 {
		p.RejectPenalty = defaults.RejectPenalty
	}
	return p
}

func (p GossipRateLimitPolicy) Validate() error {
	p = p.Normalize()
	if p.WindowBlocks == 0 || p.MaxMessagesPerNode == 0 || p.MaxMessagesPerChannel == 0 || p.MaxTopologyUpdates == 0 {
		return errors.New("payments gossip rate limit policy values must be positive")
	}
	if p.RejectPenalty <= 0 {
		return errors.New("payments gossip rate limit penalty must be positive")
	}
	return nil
}

func (p GossipRateLimitPolicy) Hash() string {
	p = p.Normalize()
	return HashParts("gossip-rate-limit-policy", fmt.Sprintf("%d", p.WindowBlocks), fmt.Sprintf("%d", p.MaxMessagesPerNode), fmt.Sprintf("%d", p.MaxMessagesPerChannel), fmt.Sprintf("%d", p.MaxTopologyUpdates), fmt.Sprintf("%d", p.RejectPenalty))
}

func (d GossipRateLimitDecision) Hash() string {
	return HashParts(
		"gossip-rate-limit-decision",
		strings.TrimSpace(d.NodeID),
		normalizeOptionalHash(d.ChannelID),
		fmt.Sprintf("%d", d.WindowStart),
		fmt.Sprintf("%d", d.WindowEnd),
		fmt.Sprintf("%d", d.NodeMessages),
		fmt.Sprintf("%d", d.ChannelMessages),
		fmt.Sprintf("%d", d.TopologyUpdates),
		fmt.Sprintf("%t", d.Allowed),
		strings.TrimSpace(d.Reason),
		normalizeHash(d.PolicyHash),
	)
}

func (p RouteFailureScoringPolicy) Normalize() RouteFailureScoringPolicy {
	defaults := DefaultRouteFailureScoringPolicy()
	if p.CapacityPenalty == 0 {
		p.CapacityPenalty = defaults.CapacityPenalty
	}
	if p.TimeoutPenalty == 0 {
		p.TimeoutPenalty = defaults.TimeoutPenalty
	}
	if p.CongestionPenalty == 0 {
		p.CongestionPenalty = defaults.CongestionPenalty
	}
	if p.LiquidityStalePenalty == 0 {
		p.LiquidityStalePenalty = defaults.LiquidityStalePenalty
	}
	if p.NodeUnavailablePenalty == 0 {
		p.NodeUnavailablePenalty = defaults.NodeUnavailablePenalty
	}
	if p.PolicyRejectedPenalty == 0 {
		p.PolicyRejectedPenalty = defaults.PolicyRejectedPenalty
	}
	if p.UnknownPenalty == 0 {
		p.UnknownPenalty = defaults.UnknownPenalty
	}
	if p.RepeatedFailurePenalty == 0 {
		p.RepeatedFailurePenalty = defaults.RepeatedFailurePenalty
	}
	if p.MaxPenalty == 0 {
		p.MaxPenalty = defaults.MaxPenalty
	}
	return p
}

func (p RouteFailureScoringPolicy) Validate() error {
	p = p.Normalize()
	if p.CapacityPenalty <= 0 || p.TimeoutPenalty <= 0 || p.CongestionPenalty <= 0 || p.LiquidityStalePenalty <= 0 || p.NodeUnavailablePenalty <= 0 || p.PolicyRejectedPenalty <= 0 || p.UnknownPenalty <= 0 || p.RepeatedFailurePenalty <= 0 || p.MaxPenalty <= 0 {
		return errors.New("payments route failure scoring penalties must be positive")
	}
	return nil
}

func (p RouteFailureScoringPolicy) penaltyFor(class RouteFailureClass) int64 {
	switch class {
	case RouteFailureCapacity:
		return p.CapacityPenalty
	case RouteFailureTimeout:
		return p.TimeoutPenalty
	case RouteFailureCongestion:
		return p.CongestionPenalty
	case RouteFailureLiquidityStale:
		return p.LiquidityStalePenalty
	case RouteFailureNodeUnavailable:
		return p.NodeUnavailablePenalty
	case RouteFailurePolicyRejected:
		return p.PolicyRejectedPenalty
	default:
		return p.UnknownPenalty
	}
}

func (s RouteFailureScore) Hash() string {
	return HashParts("route-failure-score", strings.TrimSpace(s.NodeID), normalizeHash(s.ChannelID), string(s.FailureClass), fmt.Sprintf("%d", s.FailureCount), fmt.Sprintf("%d", s.ScoreDelta), fmt.Sprintf("%d", s.ObservedHeight))
}

func (p RouteLiquidityProof) Normalize() RouteLiquidityProof {
	p.ChannelID = normalizeHash(p.ChannelID)
	p.Amount = strings.TrimSpace(p.Amount)
	p.HighValueThreshold = strings.TrimSpace(p.HighValueThreshold)
	if p.HighValueThreshold == "" {
		p.HighValueThreshold = "0"
	}
	p.RequiredDeposit = strings.TrimSpace(p.RequiredDeposit)
	if p.RequiredDeposit == "" {
		p.RequiredDeposit = "0"
	}
	p.Advertisement = p.Advertisement.Normalize()
	p.Reservation = p.Reservation.Normalize()
	p.ProofHash = normalizeOptionalHash(p.ProofHash)
	return p
}

func (p RouteLiquidityProof) ValidateBasic() error {
	proof := p.Normalize()
	if err := ValidateHash("payments route liquidity proof channel id", proof.ChannelID); err != nil {
		return err
	}
	if err := validatePositiveInt("payments route liquidity proof amount", proof.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments route liquidity high value threshold", proof.HighValueThreshold); err != nil {
		return err
	}
	if proof.CurrentHeight == 0 {
		return errors.New("payments route liquidity proof height must be positive")
	}
	if err := proof.Advertisement.Validate(proof.RequiredDeposit); err != nil {
		return err
	}
	if err := proof.Reservation.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("payments route liquidity proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputeRouteLiquidityProofHash(proof); proof.ProofHash != expected {
		return errors.New("payments route liquidity proof hash mismatch")
	}
	return nil
}

func (s TopologySpamSimulation) Hash() string {
	parts := []string{"topology-spam-simulation", fmt.Sprintf("%d", s.Accepted), fmt.Sprintf("%d", s.Rejected)}
	parts = append(parts, s.PenalizedNodes...)
	for _, reputation := range normalizeGossipReputation(s.FinalReputation) {
		parts = append(parts, reputation.NodeID, fmt.Sprintf("%d", reputation.Score), fmt.Sprintf("%d", reputation.InvalidGossip), fmt.Sprintf("%d", reputation.LastUpdateHeight))
	}
	for _, decision := range normalizeGossipRateLimitDecisions(s.Decisions) {
		parts = append(parts, decision.DecisionHash)
	}
	return HashParts(parts...)
}

func isTopologyUpdate(messageType GossipMessageType) bool {
	switch messageType {
	case GossipChannelAnnouncement, GossipChannelUpdate, GossipLiquidityHint, GossipFeePolicyUpdate:
		return true
	default:
		return false
	}
}

func addressingValidateOptionalNode(field, nodeID string) error {
	if strings.TrimSpace(nodeID) == "" {
		return errors.New(field + " is required")
	}
	return addressing.ValidateUserAddress(field, nodeID)
}

func routeFailureScoreKey(report RouteFailureReport) string {
	report = report.Normalize()
	return strings.Join([]string{report.ChannelID, report.From, report.To, string(report.FailureClass)}, "/")
}

func sortRouteFailureScores(scores []RouteFailureScore) {
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].ObservedHeight == scores[j].ObservedHeight {
			return scores[i].ScoreHash < scores[j].ScoreHash
		}
		return scores[i].ObservedHeight < scores[j].ObservedHeight
	})
}

func normalizeGossipRateLimitDecisions(decisions []GossipRateLimitDecision) []GossipRateLimitDecision {
	out := make([]GossipRateLimitDecision, len(decisions))
	copy(out, decisions)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].WindowStart == out[j].WindowStart {
			return out[i].DecisionHash < out[j].DecisionHash
		}
		return out[i].WindowStart < out[j].WindowStart
	})
	return out
}
