package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type RoutingEngineMessageType string

const (
	RoutingEngineMsgNodeAnnouncement	RoutingEngineMessageType	= "GossipNodeAnnouncement"
	RoutingEngineMsgChannelAnnouncement	RoutingEngineMessageType	= "GossipChannelAnnouncement"
	RoutingEngineMsgChannelUpdate		RoutingEngineMessageType	= "GossipChannelUpdate"
	RoutingEngineMsgLiquidityHint		RoutingEngineMessageType	= "GossipLiquidityHint"
	RoutingEngineMsgFeePolicyUpdate		RoutingEngineMessageType	= "GossipFeePolicyUpdate"
	RoutingEngineMsgRouteFailure		RoutingEngineMessageType	= "GossipRouteFailure"
	RoutingEngineMsgCapacityProbeReq	RoutingEngineMessageType	= "CapacityProbeRequest"
	RoutingEngineMsgCapacityProbeResp	RoutingEngineMessageType	= "CapacityProbeResponse"
)

type RoutingNode struct {
	NodeID		string
	AdvertisedAt	uint64
	LastSeenHeight	uint64
	Active		bool
	LocalScore	int64
	AnnouncementID	string
}

type LiquidityHint struct {
	HintID		string
	ChannelID	string
	From		string
	To		string
	Liquidity	string
	ObservedAt	uint64
	ExpiresAt	uint64
	Advisory	bool
	MessageHash	string
}

type FeePolicy struct {
	PolicyID	string
	ChannelID	string
	From		string
	To		string
	BaseFee		string
	MaxFee		string
	ValidAfter	uint64
	ValidUntil	uint64
	MessageHash	string
}

type RouteAttempt struct {
	AttemptID	string
	From		string
	To		string
	Amount		string
	CurrentHeight	uint64
	Route		ScoredRoute
	RetryCount	uint32
	Success		bool
	FailureClass	RouteFailureClass
}

type RouteFailure struct {
	FailureID	string
	Report		RouteFailureReport
	Score		RouteFailureScore
}

type LocalPeerScore struct {
	NodeID			string
	Score			int64
	InvalidGossip		uint64
	LastUpdateHeight	uint64
}

type RoutingEngineState struct {
	Topology	TopologyStore
	Nodes		[]RoutingNode
	ChannelEdges	[]ChannelEdge
	LiquidityHints	[]LiquidityHint
	FeePolicies	[]FeePolicy
	RouteAttempts	[]RouteAttempt
	RouteFailures	[]RouteFailure
	LocalPeerScores	[]LocalPeerScore
	Policy		RoutePolicy
	RateLimit	GossipRateLimitPolicy
	FailureScoring	RouteFailureScoringPolicy
}

type RoutingEngineMessage interface {
	RoutingEngineType() RoutingEngineMessageType
	Envelope() SignedGossipEnvelope
	ValidateBasic() error
}

type MsgGossipNodeAnnouncement struct{ Gossip SignedGossipEnvelope }
type MsgGossipChannelAnnouncement struct{ Gossip SignedGossipEnvelope }
type MsgGossipChannelUpdate struct{ Gossip SignedGossipEnvelope }
type MsgGossipLiquidityHint struct{ Gossip SignedGossipEnvelope }
type MsgGossipFeePolicyUpdate struct{ Gossip SignedGossipEnvelope }
type MsgGossipRouteFailure struct{ Gossip SignedGossipEnvelope }

type CapacityProbeRequest struct {
	ProbeID			string
	From			string
	To			string
	Amount			string
	CurrentHeight		uint64
	MaxHops			int
	Policy			RoutePolicy
	BlindedRouteHint	string
	RequestHash		string
}

type CapacityProbeResponse struct {
	ProbeID		string
	Responder	string
	Available	bool
	MinCapacity	string
	FailureReason	string
	RouteHash	string
	ResponseHash	string
}

func BuildRoutingGossipEnvelope(message GossipMessage, signer string, receivedAt uint64) (SignedGossipEnvelope, error) {
	if receivedAt == 0 {
		return SignedGossipEnvelope{}, errors.New("payments routing gossip received height must be positive")
	}
	built, err := BuildGossipMessage(message)
	if err != nil {
		return SignedGossipEnvelope{}, err
	}
	sig, err := SignatureForGossip(built, signer)
	if err != nil {
		return SignedGossipEnvelope{}, err
	}
	envelope := SignedGossipEnvelope{
		Message:	built,
		MessageHash:	built.MessageID,
		Signature:	sig,
		ReceivedFrom:	signer,
		ReceivedAt:	receivedAt,
	}.Normalize()
	return envelope, nil
}

func SnapshotRoutingEngineState(store TopologyStore, policy RoutePolicy, rateLimit GossipRateLimitPolicy, failurePolicy RouteFailureScoringPolicy) (RoutingEngineState, error) {
	store = store.Normalize()
	if err := store.Validate(); err != nil {
		return RoutingEngineState{}, err
	}
	state := RoutingEngineState{
		Topology:	store,
		ChannelEdges:	append([]ChannelEdge(nil), store.Edges...),
		Policy:		policy.Normalize(),
		RateLimit:	rateLimit.Normalize(),
		FailureScoring:	failurePolicy.Normalize(),
	}
	for _, envelope := range store.Messages {
		message := envelope.Normalize().Message
		switch message.MessageType {
		case GossipNodeAnnouncement:
			state.Nodes = upsertRoutingNode(state.Nodes, RoutingNode{
				NodeID:		message.NodeID,
				AdvertisedAt:	message.ValidAfterHeight,
				LastSeenHeight:	envelope.ReceivedAt,
				Active:		envelope.ReceivedAt <= message.ValidUntilHeight,
				AnnouncementID:	message.MessageID,
			})
		case GossipLiquidityHint:
			state.LiquidityHints = append(state.LiquidityHints, LiquidityHintFromGossip(message))
		case GossipFeePolicyUpdate:
			state.FeePolicies = append(state.FeePolicies, FeePolicyFromGossip(message))
		}
	}
	for _, reputation := range store.Reputation {
		reputation = reputation.Normalize()
		state.LocalPeerScores = append(state.LocalPeerScores, LocalPeerScore{
			NodeID:			reputation.NodeID,
			Score:			reputation.Score,
			InvalidGossip:		reputation.InvalidGossip,
			LastUpdateHeight:	reputation.LastUpdateHeight,
		}.Normalize())
	}
	return state.Normalize(), state.Validate()
}

func ApplyRoutingEngineMessage(engine RoutingEngineState, chain PaymentsState, msg RoutingEngineMessage, currentHeight uint64) (RoutingEngineState, GossipRateLimitDecision, error) {
	engine = engine.Normalize()
	if currentHeight == 0 {
		return RoutingEngineState{}, GossipRateLimitDecision{}, errors.New("payments routing engine height must be positive")
	}
	if msg == nil {
		return RoutingEngineState{}, GossipRateLimitDecision{}, errors.New("payments routing engine message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return RoutingEngineState{}, GossipRateLimitDecision{}, err
	}
	nextStore, decision, err := ApplyGossipEnvelopeWithRateLimit(engine.Topology, chain, msg.Envelope(), currentHeight, engine.RateLimit)
	if err != nil {
		return RoutingEngineState{}, decision, err
	}
	next, err := SnapshotRoutingEngineState(nextStore, engine.Policy, engine.RateLimit, engine.FailureScoring)
	if err != nil {
		return RoutingEngineState{}, GossipRateLimitDecision{}, err
	}
	return next, decision, nil
}

func PruneRoutingEngineTopology(engine RoutingEngineState, currentHeight uint64) (RoutingEngineState, error) {
	engine = engine.Normalize()
	pruned, err := PruneTopologyStore(engine.Topology, currentHeight)
	if err != nil {
		return RoutingEngineState{}, err
	}
	return SnapshotRoutingEngineState(pruned, engine.Policy, engine.RateLimit, engine.FailureScoring)
}

func SelectRoutingEnginePath(engine RoutingEngineState, chain PaymentsState, req RouteSelectionRequest) (RoutingEngineState, ScoredRoute, error) {
	engine = engine.Normalize()
	req = req.Normalize()
	if req.Policy.MaxHops == 0 {
		req.Policy = engine.Policy
	}
	route, err := SelectPaymentRoute(chain, engine.Topology, req)
	if err != nil {
		return RoutingEngineState{}, ScoredRoute{}, err
	}
	attempt := RouteAttempt{
		AttemptID:	HashParts("route-attempt", req.From, req.To, req.Amount, route.ScoreHash, fmt.Sprintf("%020d", req.CurrentHeight)),
		From:		req.From,
		To:		req.To,
		Amount:		req.Amount,
		CurrentHeight:	req.CurrentHeight,
		Route:		route,
		Success:	true,
	}.Normalize()
	engine.RouteAttempts = append(engine.RouteAttempts, attempt)
	return engine.Normalize(), route, nil
}

func RetryRoutingEnginePath(engine RoutingEngineState, chain PaymentsState, req RouteRetryRequest) (RoutingEngineState, RouteRetryResult, error) {
	engine = engine.Normalize()
	req = req.Normalize()
	if req.Selection.Policy.MaxHops == 0 {
		req.Selection.Policy = engine.Policy
	}
	if req.Policy.MaxAttempts == 0 {
		req.Policy = RouteRetryPolicy{MaxAttempts: 3, AlternateRouteLimit: 2, ExcludeFailedEdges: true}.Normalize()
	}
	result, err := RetryPaymentRoute(chain, engine.Topology, req)
	for _, report := range req.Failures {
		engine.RouteAttempts = append(engine.RouteAttempts, RouteAttempt{
			AttemptID:	HashParts("route-attempt-failure", report.ChannelID, report.From, report.To, string(report.FailureClass), fmt.Sprintf("%020d", report.ObservedHeight)),
			From:		report.From,
			To:		report.To,
			Amount:		req.Selection.Amount,
			CurrentHeight:	report.ObservedHeight,
			RetryCount:	uint32(len(req.Failures)),
			Success:	false,
			FailureClass:	report.FailureClass,
		}.Normalize())
	}
	if err != nil {
		return engine.Normalize(), result, err
	}
	engine.RouteAttempts = append(engine.RouteAttempts, RouteAttempt{
		AttemptID:	HashParts("route-attempt-retry", req.Selection.From, req.Selection.To, req.Selection.Amount, result.Route.ScoreHash, fmt.Sprintf("%020d", req.Selection.CurrentHeight)),
		From:		req.Selection.From,
		To:		req.Selection.To,
		Amount:		req.Selection.Amount,
		CurrentHeight:	req.Selection.CurrentHeight,
		Route:		result.Route,
		RetryCount:	result.Attempts,
		Success:	true,
	}.Normalize())
	return engine.Normalize(), result, nil
}

func ApplyRoutingEngineFailures(engine RoutingEngineState, reports []RouteFailureReport) (RoutingEngineState, []RouteFailureScore, error) {
	engine = engine.Normalize()
	store, scores, err := ApplyRouteFailureScoring(engine.Topology, reports, engine.FailureScoring)
	if err != nil {
		return RoutingEngineState{}, nil, err
	}
	engine.Topology = store
	for _, score := range scores {
		engine.RouteFailures = append(engine.RouteFailures, RouteFailure{
			FailureID:	HashParts("route-failure", score.ChannelID, score.NodeID, string(score.FailureClass), fmt.Sprintf("%020d", score.ObservedHeight)),
			Report: RouteFailureReport{
				ChannelID:	score.ChannelID,
				From:		score.NodeID,
				To:		score.NodeID,
				FailureClass:	score.FailureClass,
				Retryable:	true,
				ObservedHeight:	score.ObservedHeight,
			},
			Score:	score,
		}.Normalize())
	}
	next, err := SnapshotRoutingEngineState(engine.Topology, engine.Policy, engine.RateLimit, engine.FailureScoring)
	if err != nil {
		return RoutingEngineState{}, nil, err
	}
	next.RouteAttempts = engine.RouteAttempts
	next.RouteFailures = normalizeRouteFailures(engine.RouteFailures)
	return next.Normalize(), scores, nil
}

func HandleCapacityProbe(engine RoutingEngineState, chain PaymentsState, req CapacityProbeRequest, responder string) (CapacityProbeResponse, error) {
	engine = engine.Normalize()
	req = req.Normalize()
	if err := req.Validate(); err != nil {
		return CapacityProbeResponse{}, err
	}
	if err := addressing.ValidateUserAddress("payments capacity probe responder", responder); err != nil {
		return CapacityProbeResponse{}, err
	}
	selection := RouteSelectionRequest{
		From:		req.From,
		To:		req.To,
		Amount:		req.Amount,
		CurrentHeight:	req.CurrentHeight,
		Policy:		req.Policy,
	}.Normalize()
	if selection.Policy.MaxHops == 0 {
		selection.Policy = engine.Policy
	}
	if req.MaxHops > 0 {
		selection.Policy.MaxHops = req.MaxHops
	}
	route, err := SelectPaymentRoute(chain, engine.Topology, selection)
	if err != nil {
		return BuildCapacityProbeResponse(req, responder, ScoredRoute{}, false, "", err.Error())
	}
	return BuildCapacityProbeResponse(req, responder, route, true, route.MinCapacity, "")
}

func BuildCapacityProbeResponse(req CapacityProbeRequest, responder string, route ScoredRoute, available bool, minCapacity, failureReason string) (CapacityProbeResponse, error) {
	req = req.Normalize()
	responder = strings.TrimSpace(responder)
	if err := req.Validate(); err != nil {
		return CapacityProbeResponse{}, err
	}
	if err := addressing.ValidateUserAddress("payments capacity probe responder", responder); err != nil {
		return CapacityProbeResponse{}, err
	}
	response := CapacityProbeResponse{
		ProbeID:	req.ProbeID,
		Responder:	responder,
		Available:	available,
		MinCapacity:	strings.TrimSpace(minCapacity),
		FailureReason:	strings.TrimSpace(failureReason),
	}.Normalize()
	if route.ScoreHash != "" {
		response.RouteHash = route.Normalize().ScoreHash
	}
	response.ResponseHash = ComputeCapacityProbeResponseHash(response)
	return response.Normalize(), response.Validate()
}

func LiquidityHintFromGossip(message GossipMessage) LiquidityHint {
	message = message.Normalize()
	liquidity := message.Liquidity
	if liquidity == "" {
		liquidity = message.Capacity
	}
	return LiquidityHint{
		HintID:		HashParts("liquidity-hint", message.MessageID),
		ChannelID:	message.ChannelID,
		From:		message.From,
		To:		message.To,
		Liquidity:	liquidity,
		ObservedAt:	message.ValidAfterHeight,
		ExpiresAt:	message.ValidUntilHeight,
		Advisory:	true,
		MessageHash:	message.MessageID,
	}.Normalize()
}

func FeePolicyFromGossip(message GossipMessage) FeePolicy {
	message = message.Normalize()
	return FeePolicy{
		PolicyID:	HashParts("fee-policy-gossip", message.MessageID),
		ChannelID:	message.ChannelID,
		From:		message.From,
		To:		message.To,
		BaseFee:	message.FeeAmount,
		MaxFee:		message.MaxFee,
		ValidAfter:	message.ValidAfterHeight,
		ValidUntil:	message.ValidUntilHeight,
		MessageHash:	message.MessageID,
	}.Normalize()
}

func (m MsgGossipNodeAnnouncement) RoutingEngineType() RoutingEngineMessageType {
	return RoutingEngineMsgNodeAnnouncement
}
func (m MsgGossipNodeAnnouncement) Envelope() SignedGossipEnvelope	{ return m.Gossip.Normalize() }
func (m MsgGossipNodeAnnouncement) ValidateBasic() error {
	return validateRoutingGossipMessage(m.Gossip, GossipNodeAnnouncement)
}

func (m MsgGossipChannelAnnouncement) RoutingEngineType() RoutingEngineMessageType {
	return RoutingEngineMsgChannelAnnouncement
}
func (m MsgGossipChannelAnnouncement) Envelope() SignedGossipEnvelope	{ return m.Gossip.Normalize() }
func (m MsgGossipChannelAnnouncement) ValidateBasic() error {
	return validateRoutingGossipMessage(m.Gossip, GossipChannelAnnouncement)
}

func (m MsgGossipChannelUpdate) RoutingEngineType() RoutingEngineMessageType {
	return RoutingEngineMsgChannelUpdate
}
func (m MsgGossipChannelUpdate) Envelope() SignedGossipEnvelope	{ return m.Gossip.Normalize() }
func (m MsgGossipChannelUpdate) ValidateBasic() error {
	return validateRoutingGossipMessage(m.Gossip, GossipChannelUpdate)
}

func (m MsgGossipLiquidityHint) RoutingEngineType() RoutingEngineMessageType {
	return RoutingEngineMsgLiquidityHint
}
func (m MsgGossipLiquidityHint) Envelope() SignedGossipEnvelope	{ return m.Gossip.Normalize() }
func (m MsgGossipLiquidityHint) ValidateBasic() error {
	return validateRoutingGossipMessage(m.Gossip, GossipLiquidityHint)
}

func (m MsgGossipFeePolicyUpdate) RoutingEngineType() RoutingEngineMessageType {
	return RoutingEngineMsgFeePolicyUpdate
}
func (m MsgGossipFeePolicyUpdate) Envelope() SignedGossipEnvelope	{ return m.Gossip.Normalize() }
func (m MsgGossipFeePolicyUpdate) ValidateBasic() error {
	return validateRoutingGossipMessage(m.Gossip, GossipFeePolicyUpdate)
}

func (m MsgGossipRouteFailure) RoutingEngineType() RoutingEngineMessageType {
	return RoutingEngineMsgRouteFailure
}
func (m MsgGossipRouteFailure) Envelope() SignedGossipEnvelope	{ return m.Gossip.Normalize() }
func (m MsgGossipRouteFailure) ValidateBasic() error {
	return validateRoutingGossipMessage(m.Gossip, GossipRouteFailure)
}

func (s RoutingEngineState) Normalize() RoutingEngineState {
	s.Topology = s.Topology.Normalize()
	s.Nodes = normalizeRoutingNodes(s.Nodes)
	s.ChannelEdges = normalizeRoutingEngineEdges(s.ChannelEdges)
	s.LiquidityHints = normalizeLiquidityHints(s.LiquidityHints)
	s.FeePolicies = normalizeFeePolicies(s.FeePolicies)
	s.RouteAttempts = normalizeRouteAttempts(s.RouteAttempts)
	s.RouteFailures = normalizeRouteFailures(s.RouteFailures)
	s.LocalPeerScores = normalizeLocalPeerScores(s.LocalPeerScores)
	s.Policy = s.Policy.Normalize()
	s.RateLimit = s.RateLimit.Normalize()
	s.FailureScoring = s.FailureScoring.Normalize()
	return s
}

func (s RoutingEngineState) Validate() error {
	state := s.Normalize()
	if err := state.Topology.Validate(); err != nil {
		return err
	}
	if err := state.Policy.Validate(); err != nil {
		return err
	}
	if err := state.RateLimit.Validate(); err != nil {
		return err
	}
	if err := state.FailureScoring.Validate(); err != nil {
		return err
	}
	for _, node := range state.Nodes {
		if err := node.Validate(); err != nil {
			return err
		}
	}
	for _, hint := range state.LiquidityHints {
		if err := hint.Validate(); err != nil {
			return err
		}
	}
	for _, policy := range state.FeePolicies {
		if err := policy.Validate(); err != nil {
			return err
		}
	}
	for _, attempt := range state.RouteAttempts {
		if err := attempt.Validate(); err != nil {
			return err
		}
	}
	for _, failure := range state.RouteFailures {
		if err := failure.Validate(); err != nil {
			return err
		}
	}
	for _, score := range state.LocalPeerScores {
		if err := score.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (n RoutingNode) Normalize() RoutingNode {
	n.NodeID = strings.TrimSpace(n.NodeID)
	n.AnnouncementID = normalizeOptionalHash(n.AnnouncementID)
	return n
}

func (n RoutingNode) Validate() error {
	node := n.Normalize()
	if err := addressing.ValidateUserAddress("payments routing node id", node.NodeID); err != nil {
		return err
	}
	if node.LastSeenHeight == 0 {
		return errors.New("payments routing node last seen height must be positive")
	}
	if node.AnnouncementID != "" {
		return ValidateHash("payments routing node announcement id", node.AnnouncementID)
	}
	return nil
}

func (h LiquidityHint) Normalize() LiquidityHint {
	h.HintID = normalizeOptionalHash(h.HintID)
	h.ChannelID = normalizeHash(h.ChannelID)
	h.From = strings.TrimSpace(h.From)
	h.To = strings.TrimSpace(h.To)
	h.Liquidity = strings.TrimSpace(h.Liquidity)
	h.MessageHash = normalizeOptionalHash(h.MessageHash)
	return h
}

func (h LiquidityHint) Validate() error {
	hint := h.Normalize()
	if err := ValidateHash("payments liquidity hint id", hint.HintID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity hint channel id", hint.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity hint from", hint.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity hint to", hint.To); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments liquidity hint amount", hint.Liquidity); err != nil {
		return err
	}
	if hint.ObservedAt == 0 || hint.ExpiresAt <= hint.ObservedAt {
		return errors.New("payments liquidity hint validity window must advance")
	}
	return ValidateHash("payments liquidity hint message hash", hint.MessageHash)
}

func (p FeePolicy) Normalize() FeePolicy {
	p.PolicyID = normalizeOptionalHash(p.PolicyID)
	p.ChannelID = normalizeHash(p.ChannelID)
	p.From = strings.TrimSpace(p.From)
	p.To = strings.TrimSpace(p.To)
	p.BaseFee = strings.TrimSpace(p.BaseFee)
	if p.BaseFee == "" {
		p.BaseFee = "0"
	}
	p.MaxFee = strings.TrimSpace(p.MaxFee)
	if p.MaxFee == "" {
		p.MaxFee = "0"
	}
	p.MessageHash = normalizeOptionalHash(p.MessageHash)
	return p
}

func (p FeePolicy) Validate() error {
	policy := p.Normalize()
	if err := ValidateHash("payments fee policy id", policy.PolicyID); err != nil {
		return err
	}
	if err := ValidateHash("payments fee policy channel id", policy.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments fee policy from", policy.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments fee policy to", policy.To); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments fee policy base fee", policy.BaseFee); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments fee policy max fee", policy.MaxFee); err != nil {
		return err
	}
	if policy.ValidAfter == 0 || policy.ValidUntil <= policy.ValidAfter {
		return errors.New("payments fee policy validity window must advance")
	}
	return ValidateHash("payments fee policy message hash", policy.MessageHash)
}

func (a RouteAttempt) Normalize() RouteAttempt {
	a.AttemptID = normalizeOptionalHash(a.AttemptID)
	a.From = strings.TrimSpace(a.From)
	a.To = strings.TrimSpace(a.To)
	a.Amount = strings.TrimSpace(a.Amount)
	a.Route = a.Route.Normalize()
	return a
}

func (a RouteAttempt) Validate() error {
	attempt := a.Normalize()
	if err := ValidateHash("payments route attempt id", attempt.AttemptID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route attempt from", attempt.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route attempt to", attempt.To); err != nil {
		return err
	}
	if err := validatePositiveInt("payments route attempt amount", attempt.Amount); err != nil {
		return err
	}
	if attempt.CurrentHeight == 0 {
		return errors.New("payments route attempt height must be positive")
	}
	if attempt.Success {
		return attempt.Route.Validate()
	}
	if !IsRouteFailureClass(attempt.FailureClass) {
		return fmt.Errorf("unknown payments route attempt failure class %q", attempt.FailureClass)
	}
	return nil
}

func (f RouteFailure) Normalize() RouteFailure {
	f.FailureID = normalizeOptionalHash(f.FailureID)
	f.Report = f.Report.Normalize()
	f.Score.NodeID = strings.TrimSpace(f.Score.NodeID)
	f.Score.ChannelID = normalizeHash(f.Score.ChannelID)
	f.Score.ScoreHash = normalizeOptionalHash(f.Score.ScoreHash)
	return f
}

func (f RouteFailure) Validate() error {
	failure := f.Normalize()
	if err := ValidateHash("payments route failure id", failure.FailureID); err != nil {
		return err
	}
	if err := failure.Report.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route failure score node", failure.Score.NodeID); err != nil {
		return err
	}
	if err := ValidateHash("payments route failure score channel", failure.Score.ChannelID); err != nil {
		return err
	}
	if !IsRouteFailureClass(failure.Score.FailureClass) {
		return fmt.Errorf("unknown payments route failure score class %q", failure.Score.FailureClass)
	}
	if failure.Score.FailureCount == 0 || failure.Score.ObservedHeight == 0 {
		return errors.New("payments route failure score count and height must be positive")
	}
	return ValidateHash("payments route failure score hash", failure.Score.ScoreHash)
}

func (s LocalPeerScore) Normalize() LocalPeerScore {
	s.NodeID = strings.TrimSpace(s.NodeID)
	return s
}

func (s LocalPeerScore) Validate() error {
	score := s.Normalize()
	if err := addressing.ValidateUserAddress("payments local peer score node", score.NodeID); err != nil {
		return err
	}
	if score.LastUpdateHeight == 0 {
		return errors.New("payments local peer score height must be positive")
	}
	return nil
}

func (r CapacityProbeRequest) Normalize() CapacityProbeRequest {
	r.ProbeID = normalizeOptionalHash(r.ProbeID)
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
	r.Amount = strings.TrimSpace(r.Amount)
	r.Policy = r.Policy.Normalize()
	r.BlindedRouteHint = strings.TrimSpace(r.BlindedRouteHint)
	r.RequestHash = normalizeOptionalHash(r.RequestHash)
	if r.ProbeID == "" {
		r.ProbeID = HashParts("capacity-probe", r.From, r.To, r.Amount, fmt.Sprintf("%020d", r.CurrentHeight), r.BlindedRouteHint)
	}
	return r
}

func (r CapacityProbeRequest) Validate() error {
	req := r.Normalize()
	if err := ValidateHash("payments capacity probe id", req.ProbeID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments capacity probe from", req.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments capacity probe to", req.To); err != nil {
		return err
	}
	if req.From == req.To {
		return errors.New("payments capacity probe endpoints must differ")
	}
	if err := validatePositiveInt("payments capacity probe amount", req.Amount); err != nil {
		return err
	}
	if req.CurrentHeight == 0 {
		return errors.New("payments capacity probe height must be positive")
	}
	if req.MaxHops < 0 {
		return errors.New("payments capacity probe max hops must be non-negative")
	}
	if req.RequestHash != "" && req.RequestHash != ComputeCapacityProbeRequestHash(req) {
		return errors.New("payments capacity probe request hash mismatch")
	}
	return nil
}

func (r CapacityProbeResponse) Normalize() CapacityProbeResponse {
	r.ProbeID = normalizeHash(r.ProbeID)
	r.Responder = strings.TrimSpace(r.Responder)
	r.MinCapacity = strings.TrimSpace(r.MinCapacity)
	if r.MinCapacity == "" {
		r.MinCapacity = "0"
	}
	r.FailureReason = strings.TrimSpace(r.FailureReason)
	r.RouteHash = normalizeOptionalHash(r.RouteHash)
	r.ResponseHash = normalizeOptionalHash(r.ResponseHash)
	return r
}

func (r CapacityProbeResponse) Validate() error {
	resp := r.Normalize()
	if err := ValidateHash("payments capacity probe response id", resp.ProbeID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments capacity probe response responder", resp.Responder); err != nil {
		return err
	}
	if resp.Available {
		if err := validatePositiveInt("payments capacity probe response capacity", resp.MinCapacity); err != nil {
			return err
		}
		if err := ValidateHash("payments capacity probe response route hash", resp.RouteHash); err != nil {
			return err
		}
	} else if resp.FailureReason == "" {
		return errors.New("payments unavailable capacity probe requires failure reason")
	}
	if resp.ResponseHash == "" {
		return errors.New("payments capacity probe response hash is required")
	}
	if resp.ResponseHash != ComputeCapacityProbeResponseHash(resp) {
		return errors.New("payments capacity probe response hash mismatch")
	}
	return nil
}

func ComputeCapacityProbeRequestHash(req CapacityProbeRequest) string {
	req = req.Normalize()
	return HashParts("capacity-probe-request", req.ProbeID, req.From, req.To, req.Amount, fmt.Sprintf("%020d", req.CurrentHeight), fmt.Sprintf("%d", req.MaxHops), req.BlindedRouteHint)
}

func ComputeCapacityProbeResponseHash(resp CapacityProbeResponse) string {
	resp = resp.Normalize()
	return HashParts("capacity-probe-response", resp.ProbeID, resp.Responder, fmt.Sprintf("%t", resp.Available), resp.MinCapacity, resp.RouteHash, resp.FailureReason)
}

func validateRoutingGossipMessage(envelope SignedGossipEnvelope, expected GossipMessageType) error {
	envelope = envelope.Normalize()
	message := envelope.Message.Normalize()
	if message.MessageType != expected {
		return fmt.Errorf("payments routing gossip expected %s", expected)
	}
	if err := message.ValidateBasic(); err != nil {
		return err
	}
	return envelope.Signature.Validate(message)
}

func upsertRoutingNode(nodes []RoutingNode, next RoutingNode) []RoutingNode {
	next = next.Normalize()
	out := make([]RoutingNode, 0, len(nodes)+1)
	replaced := false
	for _, node := range nodes {
		node = node.Normalize()
		if node.NodeID == next.NodeID {
			node.LastSeenHeight = next.LastSeenHeight
			node.Active = next.Active
			node.AnnouncementID = next.AnnouncementID
			out = append(out, node)
			replaced = true
			continue
		}
		out = append(out, node)
	}
	if !replaced {
		out = append(out, next)
	}
	return normalizeRoutingNodes(out)
}

func normalizeRoutingNodes(nodes []RoutingNode) []RoutingNode {
	out := make([]RoutingNode, len(nodes))
	for i, node := range nodes {
		out[i] = node.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}

func normalizeRoutingEngineEdges(edges []ChannelEdge) []ChannelEdge {
	out := make([]ChannelEdge, len(edges))
	for i, edge := range edges {
		out[i] = edge.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return edgeKey(out[i]) < edgeKey(out[j]) })
	return out
}

func normalizeLiquidityHints(hints []LiquidityHint) []LiquidityHint {
	out := make([]LiquidityHint, len(hints))
	for i, hint := range hints {
		out[i] = hint.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].HintID < out[j].HintID })
	return out
}

func normalizeFeePolicies(policies []FeePolicy) []FeePolicy {
	out := make([]FeePolicy, len(policies))
	for i, policy := range policies {
		out[i] = policy.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PolicyID < out[j].PolicyID })
	return out
}

func normalizeRouteAttempts(attempts []RouteAttempt) []RouteAttempt {
	out := make([]RouteAttempt, len(attempts))
	for i, attempt := range attempts {
		out[i] = attempt.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].AttemptID < out[j].AttemptID })
	return out
}

func normalizeRouteFailures(failures []RouteFailure) []RouteFailure {
	out := make([]RouteFailure, len(failures))
	for i, failure := range failures {
		out[i] = failure.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].FailureID < out[j].FailureID })
	return out
}

func normalizeLocalPeerScores(scores []LocalPeerScore) []LocalPeerScore {
	out := make([]LocalPeerScore, len(scores))
	for i, score := range scores {
		out[i] = score.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].NodeID < out[j].NodeID })
	return out
}
