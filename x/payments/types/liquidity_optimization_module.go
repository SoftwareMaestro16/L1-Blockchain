package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type LiquidityOptimizationMessageType string

const (
	LiquidityMsgAdvertiseLiquidity		LiquidityOptimizationMessageType	= "MsgAdvertiseLiquidity"
	LiquidityMsgReserveLiquidity		LiquidityOptimizationMessageType	= "MsgReserveLiquidity"
	LiquidityMsgReleaseReservation		LiquidityOptimizationMessageType	= "MsgReleaseReservation"
	LiquidityMsgUpdateFeePolicy		LiquidityOptimizationMessageType	= "MsgUpdateFeePolicy"
	LiquidityMsgSubmitRebalanceIntent	LiquidityOptimizationMessageType	= "MsgSubmitRebalanceIntent"
	LiquidityMsgSetLiquidityLimits		LiquidityOptimizationMessageType	= "MsgSetLiquidityLimits"
)

type LiquidityPosition struct {
	PositionID		string
	ChannelID		string
	Owner			string
	Counterparty		string
	AdvertisedCapacity	string
	ReservedCapacity	string
	AvailableCapacity	string
	FeePolicyID		string
	ReliabilityBps		uint32
	UpdatedHeight		uint64
	AdvertisementHash	string
	PositionHash		string
}

type Reservation struct {
	ReservationID		string
	AdvertisementID		string
	ChannelID		string
	Reserver		string
	Counterparty		string
	Capacity		string
	FeeAmount		string
	ExpirationHeight	uint64
	CreatedHeight		uint64
	Released		bool
	ReleaseHeight		uint64
	CommitmentHash		string
}

type RebalanceIntent struct {
	IntentID		string
	ChannelID		string
	Owner			string
	TargetCapacity		string
	MaxSettlementLoad	uint32
	Priority		uint32
	SubmittedHeight		uint64
	ExpiresHeight		uint64
	IntentHash		string
}

type CapacityForecast struct {
	ForecastID		string
	ChannelID		string
	From			string
	To			string
	AvailableCapacity	string
	ReservedCapacity	string
	PendingConditionCount	uint32
	ReservePressureBps	uint32
	ForecastHeight		uint64
	ExpiresHeight		uint64
	ForecastHash		string
}

type LiquidityScore struct {
	ScoreID		string
	ChannelID	string
	From		string
	To		string
	RawScore	int64
	Score		int64
	DecayBps	uint32
	UpdatedHeight	uint64
	ExpiresHeight	uint64
	ScoreHash	string
}

type LiquidityLimits struct {
	LimitID			string
	ChannelID		string
	Participant		string
	MaxReservedCapacity	string
	MinAvailableCapacity	string
	MaxBaseFee		string
	MaxReservationFee	string
	MaxVirtualSetupFee	string
	MaxProportionalBps	uint32
	MaxRebalanceLoad	uint32
	UpdatedHeight		uint64
	LimitHash		string
}

type LiquidityFeePolicyBounds struct {
	MaxBaseFee		string
	MaxReservationFee	string
	MaxVirtualSetupFee	string
	MaxCongestionFee	string
	MaxFailurePenalty	string
	MaxHopFee		string
	MaxProportionalBps	uint32
	MinValidityWindow	uint64
	MaxValidityWindow	uint64
}

type LiquidityOptimizationState struct {
	Positions		[]LiquidityPosition
	Reservations		[]Reservation
	RebalanceIntents	[]RebalanceIntent
	FeePolicies		[]FeePolicy
	Forecasts		[]CapacityForecast
	Scores			[]LiquidityScore
	Limits			[]LiquidityLimits
}

type LiquidityOptimizationMessage interface {
	LiquidityOptimizationType() LiquidityOptimizationMessageType
	ValidateBasic() error
}

type MsgAdvertiseLiquidity struct {
	Signer		string
	Advertisement	LiquidityAdvertisement
	RequiredDeposit	string
	CurrentHeight	uint64
}

type MsgReserveLiquidity struct {
	Reservation	SignedLiquidityReservation
	CurrentHeight	uint64
}

type MsgReleaseReservation struct {
	ReservationID	string
	ChannelID	string
	Releaser	string
	CurrentHeight	uint64
	Reason		string
}

type MsgUpdateFeePolicy struct {
	Signer		string
	Policy		RoutingFeePolicyUpdate
	Bounds		LiquidityFeePolicyBounds
	CurrentHeight	uint64
}

type MsgSubmitRebalanceIntent struct {
	Signer		string
	Intent		RebalanceIntent
	CurrentHeight	uint64
}

type MsgSetLiquidityLimits struct {
	Signer		string
	Limits		LiquidityLimits
	CurrentHeight	uint64
}

func EmptyLiquidityOptimizationState() LiquidityOptimizationState {
	return LiquidityOptimizationState{}
}

func ApplyLiquidityOptimizationMessage(chain PaymentsState, state LiquidityOptimizationState, msg LiquidityOptimizationMessage) (LiquidityOptimizationState, error) {
	chain = chain.Export()
	state = state.Export()
	if msg == nil {
		return LiquidityOptimizationState{}, errors.New("payments liquidity optimization message is required")
	}
	if err := msg.ValidateBasic(); err != nil {
		return LiquidityOptimizationState{}, err
	}
	var err error
	switch m := msg.(type) {
	case MsgAdvertiseLiquidity:
		state, err = AdvertiseLiquidity(chain, state, m)
	case MsgReserveLiquidity:
		state, err = ReserveLiquidity(chain, state, m)
	case MsgReleaseReservation:
		state, err = ReleaseLiquidityReservation(state, m)
	case MsgUpdateFeePolicy:
		state, err = UpdateLiquidityFeePolicy(chain, state, m)
	case MsgSubmitRebalanceIntent:
		state, err = SubmitLiquidityRebalanceIntent(chain, state, m)
	case MsgSetLiquidityLimits:
		state, err = SetLiquidityLimits(chain, state, m)
	default:
		return LiquidityOptimizationState{}, errors.New("payments liquidity optimization message type is unsupported")
	}
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	return state.Export(), nil
}

func AdvertiseLiquidity(chain PaymentsState, state LiquidityOptimizationState, msg MsgAdvertiseLiquidity) (LiquidityOptimizationState, error) {
	chain = chain.Export()
	state = state.Export()
	msg = msg.Normalize()
	channel, found := chain.ChannelByID(msg.Advertisement.ChannelID)
	if !found {
		return LiquidityOptimizationState{}, errors.New("payments liquidity advertisement channel not found")
	}
	if channel.Status != ChannelStatusOpen {
		return LiquidityOptimizationState{}, errors.New("payments liquidity advertisement requires open channel")
	}
	if !containsString(channel.Participants, msg.Signer) || msg.Advertisement.Advertiser != msg.Signer {
		return LiquidityOptimizationState{}, errors.New("payments liquidity advertisement signer must be advertising participant")
	}
	if !containsString(channel.Participants, msg.Advertisement.Counterparty) {
		return LiquidityOptimizationState{}, errors.New("payments liquidity advertisement counterparty must be participant")
	}
	ad, err := BuildLiquidityAdvertisement(msg.Advertisement, msg.RequiredDeposit)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	if err := validateAdvertisementCapacity(chain, state, ad); err != nil {
		return LiquidityOptimizationState{}, err
	}
	reserved, err := activeReservedCapacity(state, ad.ChannelID, ad.AdvertisementID, msg.CurrentHeight)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	available, err := capacityMinus(ad.Capacity, reserved.String())
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	position := LiquidityPosition{
		PositionID:		HashParts("liquidity-position", ad.AdvertisementID),
		ChannelID:		ad.ChannelID,
		Owner:			ad.Advertiser,
		Counterparty:		ad.Counterparty,
		AdvertisedCapacity:	ad.Capacity,
		ReservedCapacity:	reserved.String(),
		AvailableCapacity:	available,
		FeePolicyID:		ad.AdvertisementID,
		ReliabilityBps:		ad.ReliabilityBps,
		UpdatedHeight:		msg.CurrentHeight,
		AdvertisementHash:	ad.AdvertisementHash,
	}
	position = position.WithHash()
	next := state.Clone()
	next.Positions = upsertLiquidityPosition(next.Positions, position)
	forecast, err := BuildCapacityForecast(chain, next, ad.ChannelID, ad.Advertiser, ad.Counterparty, msg.CurrentHeight, DefaultGossipTTL)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	next.Forecasts = upsertCapacityForecast(next.Forecasts, forecast)
	score, err := BuildLiquidityScoreForAdvertisement(ad, EdgeRoutingStats{ChannelID: ad.ChannelID, From: ad.Advertiser, To: ad.Counterparty, SuccessRateBps: ad.ReliabilityBps, LiquidityUpdatedHeight: msg.CurrentHeight}, msg.CurrentHeight, DefaultGossipTTL)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	next.Scores = upsertLiquidityScore(next.Scores, score)
	return next.Export(), nil
}

func ReserveLiquidity(chain PaymentsState, state LiquidityOptimizationState, msg MsgReserveLiquidity) (LiquidityOptimizationState, error) {
	chain = chain.Export()
	state = state.Export()
	msg = msg.Normalize()
	reservation := msg.Reservation.Normalize()
	if err := reservation.Validate(); err != nil {
		return LiquidityOptimizationState{}, err
	}
	if msg.CurrentHeight >= reservation.ExpirationHeight {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation is already expired")
	}
	channel, found := chain.ChannelByID(reservation.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation requires open channel")
	}
	position, found := state.PositionByAdvertisement(reservation.AdvertisementID)
	if !found {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation advertisement not found")
	}
	if reservation.ChannelID != position.ChannelID || reservation.Counterparty != position.Counterparty {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation domain mismatch")
	}
	limits := state.LimitsFor(position.ChannelID, position.Owner)
	if err := validateReservationCapacity(chain, state, position, reservation, limits, msg.CurrentHeight); err != nil {
		return LiquidityOptimizationState{}, err
	}
	next := state.Clone()
	next.Reservations = upsertReservation(next.Reservations, ReservationFromSigned(reservation, msg.CurrentHeight))
	next, err := RecalculateLiquidityPosition(chain, next, position.ChannelID, position.Owner, position.Counterparty, msg.CurrentHeight)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	return next.Export(), nil
}

func ReleaseLiquidityReservation(state LiquidityOptimizationState, msg MsgReleaseReservation) (LiquidityOptimizationState, error) {
	state = state.Export()
	msg = msg.Normalize()
	index, reservation, found := state.ReservationIndex(msg.ReservationID)
	if !found {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation not found")
	}
	if reservation.ChannelID != msg.ChannelID {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation channel mismatch")
	}
	if msg.Releaser != reservation.Reserver && msg.Releaser != reservation.Counterparty {
		return LiquidityOptimizationState{}, errors.New("payments liquidity reservation releaser is unauthorized")
	}
	if reservation.Released {
		return state, nil
	}
	reservation.Released = true
	reservation.ReleaseHeight = msg.CurrentHeight
	next := state.Clone()
	next.Reservations[index] = reservation.Normalize()
	return next.Export(), nil
}

func UpdateLiquidityFeePolicy(chain PaymentsState, state LiquidityOptimizationState, msg MsgUpdateFeePolicy) (LiquidityOptimizationState, error) {
	chain = chain.Export()
	state = state.Export()
	msg = msg.Normalize()
	channel, found := chain.ChannelByID(msg.Policy.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return LiquidityOptimizationState{}, errors.New("payments liquidity fee policy requires open channel")
	}
	if msg.Policy.From != msg.Signer || !containsString(channel.Participants, msg.Signer) || !containsString(channel.Participants, msg.Policy.To) {
		return LiquidityOptimizationState{}, errors.New("payments liquidity fee policy signer must be forwarding participant")
	}
	if err := ValidateLiquidityFeePolicyBounds(msg.Policy, msg.Bounds, msg.CurrentHeight); err != nil {
		return LiquidityOptimizationState{}, err
	}
	policy := FeePolicy{
		PolicyID:	msg.Policy.PolicyID,
		ChannelID:	msg.Policy.ChannelID,
		From:		msg.Policy.From,
		To:		msg.Policy.To,
		BaseFee:	msg.Policy.BaseHopFee,
		MaxFee:		msg.Policy.MaxHopFee,
		ValidAfter:	msg.Policy.ValidAfterHeight,
		ValidUntil:	msg.Policy.ValidUntilHeight,
		MessageHash:	msg.Policy.PolicyHash,
	}.Normalize()
	if err := policy.Validate(); err != nil {
		return LiquidityOptimizationState{}, err
	}
	next := state.Clone()
	next.FeePolicies = upsertFeePolicy(next.FeePolicies, policy)
	return next.Export(), nil
}

func SubmitLiquidityRebalanceIntent(chain PaymentsState, state LiquidityOptimizationState, msg MsgSubmitRebalanceIntent) (LiquidityOptimizationState, error) {
	chain = chain.Export()
	state = state.Export()
	msg = msg.Normalize()
	channel, found := chain.ChannelByID(msg.Intent.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return LiquidityOptimizationState{}, errors.New("payments rebalance intent requires open channel")
	}
	if msg.Intent.Owner != msg.Signer || !containsString(channel.Participants, msg.Signer) {
		return LiquidityOptimizationState{}, errors.New("payments rebalance intent signer must be owner")
	}
	limit := state.LimitsFor(msg.Intent.ChannelID, msg.Signer)
	if limit.MaxRebalanceLoad > 0 && msg.Intent.MaxSettlementLoad > limit.MaxRebalanceLoad {
		return LiquidityOptimizationState{}, errors.New("payments rebalance intent exceeds settlement load limit")
	}
	if msg.Intent.MaxSettlementLoad > MaxSettlementBatchOps {
		return LiquidityOptimizationState{}, errors.New("payments rebalance intent creates excessive settlement load")
	}
	next := state.Clone()
	next.RebalanceIntents = upsertRebalanceIntent(next.RebalanceIntents, msg.Intent.WithHash())
	return next.Export(), nil
}

func SetLiquidityLimits(chain PaymentsState, state LiquidityOptimizationState, msg MsgSetLiquidityLimits) (LiquidityOptimizationState, error) {
	chain = chain.Export()
	state = state.Export()
	msg = msg.Normalize()
	channel, found := chain.ChannelByID(msg.Limits.ChannelID)
	if !found || channel.Status != ChannelStatusOpen {
		return LiquidityOptimizationState{}, errors.New("payments liquidity limits require open channel")
	}
	if msg.Limits.Participant != msg.Signer || !containsString(channel.Participants, msg.Signer) {
		return LiquidityOptimizationState{}, errors.New("payments liquidity limits signer must be participant")
	}
	next := state.Clone()
	next.Limits = upsertLiquidityLimits(next.Limits, msg.Limits.WithHash())
	return next.Export(), nil
}

func ExpireLiquidityReservations(chain PaymentsState, state LiquidityOptimizationState, currentHeight uint64) (LiquidityOptimizationState, []Reservation, error) {
	if currentHeight == 0 {
		return LiquidityOptimizationState{}, nil, errors.New("payments liquidity reservation expiry height must be positive")
	}
	chain = chain.Export()
	state = state.Export()
	next := state.Clone()
	expired := []Reservation{}
	affected := map[string]Reservation{}
	for i, reservation := range next.Reservations {
		reservation = reservation.Normalize()
		if reservation.Released || reservation.ExpirationHeight > currentHeight {
			continue
		}
		reservation.Released = true
		reservation.ReleaseHeight = currentHeight
		next.Reservations[i] = reservation.Normalize()
		expired = append(expired, reservation.Normalize())
		affected[reservation.ChannelID+"/"+reservation.Reserver+"/"+reservation.Counterparty] = reservation
	}
	var err error
	for _, reservation := range affected {
		next, err = RecalculateLiquidityPosition(chain, next, reservation.ChannelID, reservation.Reserver, reservation.Counterparty, currentHeight)
		if err != nil {
			return LiquidityOptimizationState{}, nil, err
		}
	}
	return next.Export(), normalizeReservations(expired), nil
}

func RecalculateLiquidityPosition(chain PaymentsState, state LiquidityOptimizationState, channelID, owner, counterparty string, currentHeight uint64) (LiquidityOptimizationState, error) {
	state = state.Export()
	position, found := state.PositionByEndpoint(channelID, owner, counterparty)
	if !found {
		return state, nil
	}
	reserved, err := activeReservedCapacity(state, channelID, position.FeePolicyID, currentHeight)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	available, err := capacityMinus(position.AdvertisedCapacity, reserved.String())
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	position.ReservedCapacity = reserved.String()
	position.AvailableCapacity = available
	position.UpdatedHeight = currentHeight
	position = position.WithHash()
	next := state.Clone()
	next.Positions = upsertLiquidityPosition(next.Positions, position)
	forecast, err := BuildCapacityForecast(chain, next, channelID, owner, counterparty, currentHeight, DefaultGossipTTL)
	if err != nil {
		return LiquidityOptimizationState{}, err
	}
	next.Forecasts = upsertCapacityForecast(next.Forecasts, forecast)
	return next.Export(), nil
}

func BuildCapacityForecast(chain PaymentsState, state LiquidityOptimizationState, channelID, from, to string, currentHeight, ttl uint64) (CapacityForecast, error) {
	if currentHeight == 0 {
		return CapacityForecast{}, errors.New("payments capacity forecast height must be positive")
	}
	if ttl == 0 {
		ttl = DefaultGossipTTL
	}
	chain = chain.Export()
	state = state.Export()
	channel, found := chain.ChannelByID(channelID)
	if !found {
		return CapacityForecast{}, errors.New("payments capacity forecast channel not found")
	}
	if !containsString(channel.Participants, from) || !containsString(channel.Participants, to) {
		return CapacityForecast{}, errors.New("payments capacity forecast endpoints must be participants")
	}
	ownerCapacity, err := spendableChannelBalance(channel, from)
	if err != nil {
		return CapacityForecast{}, err
	}
	reserved, err := activeReservedCapacityForEndpoint(state, channelID, from, currentHeight)
	if err != nil {
		return CapacityForecast{}, err
	}
	available := ownerCapacity.Sub(reserved)
	if available.IsNegative() {
		available = sdkmath.ZeroInt()
	}
	pressure := uint32(0)
	if !ownerCapacity.IsZero() {
		pressure = uint32(reserved.MulRaw(10_000).Quo(ownerCapacity).Uint64())
		if pressure > MaxPenaltyRouteBps {
			pressure = MaxPenaltyRouteBps
		}
	}
	forecast := CapacityForecast{
		ForecastID:		HashParts("liquidity-capacity-forecast", channelID, from, to, fmt.Sprintf("%020d", currentHeight)),
		ChannelID:		channelID,
		From:			from,
		To:			to,
		AvailableCapacity:	available.String(),
		ReservedCapacity:	reserved.String(),
		PendingConditionCount:	uint32(len(channel.LatestState.Conditions)),
		ReservePressureBps:	pressure,
		ForecastHeight:		currentHeight,
		ExpiresHeight:		currentHeight + ttl,
	}
	return forecast.WithHash(), nil
}

func BuildLiquidityScoreForAdvertisement(ad LiquidityAdvertisement, stats EdgeRoutingStats, currentHeight, ttl uint64) (LiquidityScore, error) {
	if ttl == 0 {
		ttl = DefaultGossipTTL
	}
	raw, err := LiquidityAvailabilityScore(ad, stats)
	if err != nil {
		return LiquidityScore{}, err
	}
	score := LiquidityScore{
		ScoreID:	HashParts("liquidity-score", ad.ChannelID, ad.Advertiser, ad.Counterparty),
		ChannelID:	ad.ChannelID,
		From:		ad.Advertiser,
		To:		ad.Counterparty,
		RawScore:	raw,
		Score:		raw,
		DecayBps:	0,
		UpdatedHeight:	currentHeight,
		ExpiresHeight:	currentHeight + ttl,
	}
	return score.WithHash(), nil
}

func DecayLiquidityScores(state LiquidityOptimizationState, currentHeight, halfLife uint64) (LiquidityOptimizationState, error) {
	if currentHeight == 0 {
		return LiquidityOptimizationState{}, errors.New("payments liquidity score decay height must be positive")
	}
	state = state.Export()
	if halfLife == 0 {
		halfLife = DefaultGossipTTL
	}
	next := state.Clone()
	for i, score := range next.Scores {
		score = score.Normalize()
		if currentHeight <= score.UpdatedHeight {
			continue
		}
		elapsed := currentHeight - score.UpdatedHeight
		periods := elapsed / halfLife
		if periods == 0 {
			continue
		}
		decayed := score.Score
		for j := uint64(0); j < periods; j++ {
			decayed /= 2
		}
		score.Score = decayed
		score.DecayBps = uint32Min(MaxPenaltyRouteBps, score.DecayBps+uint32Min(MaxPenaltyRouteBps, uint32(periods)*5_000))
		score.UpdatedHeight = currentHeight
		next.Scores[i] = score.WithHash()
	}
	return next.Export(), nil
}

func ValidateLiquidityFeePolicyBounds(policy RoutingFeePolicyUpdate, bounds LiquidityFeePolicyBounds, currentHeight uint64) error {
	policy = policy.Normalize()
	bounds = bounds.Normalize()
	if err := policy.ValidateAtHeight(currentHeight); err != nil {
		return err
	}
	window := policy.ValidUntilHeight - policy.ValidAfterHeight
	if bounds.MinValidityWindow > 0 && window < bounds.MinValidityWindow {
		return errors.New("payments liquidity fee policy validity window below minimum")
	}
	if bounds.MaxValidityWindow > 0 && window > bounds.MaxValidityWindow {
		return errors.New("payments liquidity fee policy validity window exceeds maximum")
	}
	if bounds.MaxProportionalBps > 0 && policy.ProportionalFeeBps > bounds.MaxProportionalBps {
		return errors.New("payments liquidity fee policy proportional fee exceeds bounds")
	}
	for _, item := range []struct {
		name	string
		value	string
		bound	string
	}{
		{"payments liquidity fee policy base fee", policy.BaseHopFee, bounds.MaxBaseFee},
		{"payments liquidity fee policy reservation fee", policy.LiquidityReservationFee, bounds.MaxReservationFee},
		{"payments liquidity fee policy virtual setup fee", policy.VirtualChannelSetupFee, bounds.MaxVirtualSetupFee},
		{"payments liquidity fee policy congestion fee", policy.CongestionSurcharge, bounds.MaxCongestionFee},
		{"payments liquidity fee policy failure penalty", policy.FailurePenalty, bounds.MaxFailurePenalty},
		{"payments liquidity fee policy max hop fee", policy.MaxHopFee, bounds.MaxHopFee},
	} {
		if err := requireAmountAtMost(item.name, item.value, item.bound); err != nil {
			return err
		}
	}
	return nil
}

func ReservationFromSigned(reservation SignedLiquidityReservation, createdHeight uint64) Reservation {
	reservation = reservation.Normalize()
	return Reservation{
		ReservationID:		reservation.ReservationID,
		AdvertisementID:	reservation.AdvertisementID,
		ChannelID:		reservation.ChannelID,
		Reserver:		reservation.Reserver,
		Counterparty:		reservation.Counterparty,
		Capacity:		reservation.Capacity,
		FeeAmount:		reservation.FeeAmount,
		ExpirationHeight:	reservation.ExpirationHeight,
		CreatedHeight:		createdHeight,
		CommitmentHash:		reservation.CommitmentHash,
	}.Normalize()
}

func (s LiquidityOptimizationState) Export() LiquidityOptimizationState {
	return s.Clone().Normalize()
}

func (s LiquidityOptimizationState) Clone() LiquidityOptimizationState {
	out := LiquidityOptimizationState{
		Positions:		make([]LiquidityPosition, len(s.Positions)),
		Reservations:		make([]Reservation, len(s.Reservations)),
		RebalanceIntents:	make([]RebalanceIntent, len(s.RebalanceIntents)),
		FeePolicies:		make([]FeePolicy, len(s.FeePolicies)),
		Forecasts:		make([]CapacityForecast, len(s.Forecasts)),
		Scores:			make([]LiquidityScore, len(s.Scores)),
		Limits:			make([]LiquidityLimits, len(s.Limits)),
	}
	copy(out.Positions, s.Positions)
	copy(out.Reservations, s.Reservations)
	copy(out.RebalanceIntents, s.RebalanceIntents)
	copy(out.FeePolicies, s.FeePolicies)
	copy(out.Forecasts, s.Forecasts)
	copy(out.Scores, s.Scores)
	copy(out.Limits, s.Limits)
	return out
}

func (s LiquidityOptimizationState) Normalize() LiquidityOptimizationState {
	s.Positions = normalizeLiquidityPositions(s.Positions)
	s.Reservations = normalizeReservations(s.Reservations)
	s.RebalanceIntents = normalizeRebalanceIntents(s.RebalanceIntents)
	s.FeePolicies = normalizeFeePolicies(s.FeePolicies)
	s.Forecasts = normalizeCapacityForecasts(s.Forecasts)
	s.Scores = normalizeLiquidityScores(s.Scores)
	s.Limits = normalizeLiquidityLimits(s.Limits)
	return s
}

func (s LiquidityOptimizationState) Validate() error {
	state := s.Normalize()
	for _, position := range state.Positions {
		if err := position.Validate(); err != nil {
			return err
		}
	}
	for _, reservation := range state.Reservations {
		if err := reservation.Validate(); err != nil {
			return err
		}
	}
	for _, intent := range state.RebalanceIntents {
		if err := intent.Validate(); err != nil {
			return err
		}
	}
	for _, policy := range state.FeePolicies {
		if err := policy.Validate(); err != nil {
			return err
		}
	}
	for _, forecast := range state.Forecasts {
		if err := forecast.Validate(); err != nil {
			return err
		}
	}
	for _, score := range state.Scores {
		if err := score.Validate(); err != nil {
			return err
		}
	}
	for _, limits := range state.Limits {
		if err := limits.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s LiquidityOptimizationState) PositionByAdvertisement(advertisementID string) (LiquidityPosition, bool) {
	advertisementID = normalizeHash(advertisementID)
	for _, position := range s.Normalize().Positions {
		if position.FeePolicyID == advertisementID {
			return position, true
		}
	}
	return LiquidityPosition{}, false
}

func (s LiquidityOptimizationState) PositionByEndpoint(channelID, owner, counterparty string) (LiquidityPosition, bool) {
	channelID = normalizeHash(channelID)
	owner = strings.TrimSpace(owner)
	counterparty = strings.TrimSpace(counterparty)
	for _, position := range s.Normalize().Positions {
		if position.ChannelID == channelID && position.Owner == owner && position.Counterparty == counterparty {
			return position, true
		}
	}
	return LiquidityPosition{}, false
}

func (s LiquidityOptimizationState) ReservationIndex(reservationID string) (int, Reservation, bool) {
	reservationID = normalizeHash(reservationID)
	for i, reservation := range s.Normalize().Reservations {
		if reservation.ReservationID == reservationID {
			return i, reservation, true
		}
	}
	return -1, Reservation{}, false
}

func (s LiquidityOptimizationState) LimitsFor(channelID, participant string) LiquidityLimits {
	channelID = normalizeHash(channelID)
	participant = strings.TrimSpace(participant)
	for _, limits := range s.Normalize().Limits {
		if limits.ChannelID == channelID && limits.Participant == participant {
			return limits
		}
	}
	return LiquidityLimits{}
}

func (p LiquidityPosition) Normalize() LiquidityPosition {
	p.PositionID = normalizeOptionalHash(p.PositionID)
	p.ChannelID = normalizeHash(p.ChannelID)
	p.Owner = strings.TrimSpace(p.Owner)
	p.Counterparty = strings.TrimSpace(p.Counterparty)
	p.AdvertisedCapacity = strings.TrimSpace(p.AdvertisedCapacity)
	p.ReservedCapacity = strings.TrimSpace(p.ReservedCapacity)
	p.AvailableCapacity = strings.TrimSpace(p.AvailableCapacity)
	p.FeePolicyID = normalizeOptionalHash(p.FeePolicyID)
	p.AdvertisementHash = normalizeOptionalHash(p.AdvertisementHash)
	p.PositionHash = normalizeOptionalHash(p.PositionHash)
	for _, field := range []*string{&p.ReservedCapacity, &p.AvailableCapacity} {
		if *field == "" {
			*field = "0"
		}
	}
	return p
}

func (p LiquidityPosition) WithHash() LiquidityPosition {
	p = p.Normalize()
	p.PositionHash = HashParts("liquidity-position", p.PositionID, p.ChannelID, p.Owner, p.Counterparty, p.AdvertisedCapacity, p.ReservedCapacity, p.AvailableCapacity, p.FeePolicyID, fmt.Sprintf("%010d", p.ReliabilityBps), fmt.Sprintf("%020d", p.UpdatedHeight), p.AdvertisementHash)
	return p.Normalize()
}

func (p LiquidityPosition) Validate() error {
	position := p.Normalize()
	if err := ValidateHash("payments liquidity position id", position.PositionID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity position channel", position.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity position owner", position.Owner); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity position counterparty", position.Counterparty); err != nil {
		return err
	}
	if position.Owner == position.Counterparty {
		return errors.New("payments liquidity position endpoints must differ")
	}
	if err := validatePositiveInt("payments liquidity position advertised capacity", position.AdvertisedCapacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments liquidity position reserved capacity", position.ReservedCapacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments liquidity position available capacity", position.AvailableCapacity); err != nil {
		return err
	}
	if position.ReliabilityBps > MaxPenaltyRouteBps {
		return errors.New("payments liquidity position reliability exceeds 10000")
	}
	if position.UpdatedHeight == 0 {
		return errors.New("payments liquidity position height must be positive")
	}
	return ValidateHash("payments liquidity position hash", position.PositionHash)
}

func (r Reservation) Normalize() Reservation {
	r.ReservationID = normalizeOptionalHash(r.ReservationID)
	r.AdvertisementID = normalizeHash(r.AdvertisementID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Reserver = strings.TrimSpace(r.Reserver)
	r.Counterparty = strings.TrimSpace(r.Counterparty)
	r.Capacity = strings.TrimSpace(r.Capacity)
	r.FeeAmount = strings.TrimSpace(r.FeeAmount)
	if r.FeeAmount == "" {
		r.FeeAmount = "0"
	}
	r.CommitmentHash = normalizeOptionalHash(r.CommitmentHash)
	return r
}

func (r Reservation) Validate() error {
	reservation := r.Normalize()
	if err := ValidateHash("payments reservation id", reservation.ReservationID); err != nil {
		return err
	}
	if err := ValidateHash("payments reservation advertisement", reservation.AdvertisementID); err != nil {
		return err
	}
	if err := ValidateHash("payments reservation channel", reservation.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments reservation reserver", reservation.Reserver); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments reservation counterparty", reservation.Counterparty); err != nil {
		return err
	}
	if reservation.Reserver == reservation.Counterparty {
		return errors.New("payments reservation endpoints must differ")
	}
	if err := validatePositiveInt("payments reservation capacity", reservation.Capacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments reservation fee", reservation.FeeAmount); err != nil {
		return err
	}
	if reservation.ExpirationHeight == 0 || reservation.CreatedHeight == 0 {
		return errors.New("payments reservation heights must be positive")
	}
	if reservation.Released && reservation.ReleaseHeight == 0 {
		return errors.New("payments released reservation requires release height")
	}
	return ValidateHash("payments reservation commitment", reservation.CommitmentHash)
}

func (r RebalanceIntent) Normalize() RebalanceIntent {
	r.IntentID = normalizeOptionalHash(r.IntentID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Owner = strings.TrimSpace(r.Owner)
	r.TargetCapacity = strings.TrimSpace(r.TargetCapacity)
	r.IntentHash = normalizeOptionalHash(r.IntentHash)
	return r
}

func (r RebalanceIntent) WithHash() RebalanceIntent {
	r = r.Normalize()
	if r.IntentID == "" {
		r.IntentID = HashParts("rebalance-intent", r.ChannelID, r.Owner, r.TargetCapacity, fmt.Sprintf("%020d", r.SubmittedHeight))
	}
	r.IntentHash = HashParts("rebalance-intent", r.IntentID, r.ChannelID, r.Owner, r.TargetCapacity, fmt.Sprintf("%010d", r.MaxSettlementLoad), fmt.Sprintf("%010d", r.Priority), fmt.Sprintf("%020d", r.SubmittedHeight), fmt.Sprintf("%020d", r.ExpiresHeight))
	return r.Normalize()
}

func (r RebalanceIntent) Validate() error {
	intent := r.Normalize()
	if err := ValidateHash("payments rebalance intent id", intent.IntentID); err != nil {
		return err
	}
	if err := ValidateHash("payments rebalance intent channel", intent.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments rebalance intent owner", intent.Owner); err != nil {
		return err
	}
	if err := validatePositiveInt("payments rebalance target capacity", intent.TargetCapacity); err != nil {
		return err
	}
	if intent.MaxSettlementLoad == 0 || intent.MaxSettlementLoad > MaxSettlementBatchOps {
		return errors.New("payments rebalance settlement load is invalid")
	}
	if intent.SubmittedHeight == 0 || intent.ExpiresHeight <= intent.SubmittedHeight {
		return errors.New("payments rebalance intent validity window must advance")
	}
	return ValidateHash("payments rebalance intent hash", intent.IntentHash)
}

func (f CapacityForecast) Normalize() CapacityForecast {
	f.ForecastID = normalizeOptionalHash(f.ForecastID)
	f.ChannelID = normalizeHash(f.ChannelID)
	f.From = strings.TrimSpace(f.From)
	f.To = strings.TrimSpace(f.To)
	f.AvailableCapacity = strings.TrimSpace(f.AvailableCapacity)
	f.ReservedCapacity = strings.TrimSpace(f.ReservedCapacity)
	f.ForecastHash = normalizeOptionalHash(f.ForecastHash)
	return f
}

func (f CapacityForecast) WithHash() CapacityForecast {
	f = f.Normalize()
	f.ForecastHash = HashParts("capacity-forecast", f.ForecastID, f.ChannelID, f.From, f.To, f.AvailableCapacity, f.ReservedCapacity, fmt.Sprintf("%010d", f.PendingConditionCount), fmt.Sprintf("%010d", f.ReservePressureBps), fmt.Sprintf("%020d", f.ForecastHeight), fmt.Sprintf("%020d", f.ExpiresHeight))
	return f.Normalize()
}

func (f CapacityForecast) Validate() error {
	forecast := f.Normalize()
	if err := ValidateHash("payments capacity forecast id", forecast.ForecastID); err != nil {
		return err
	}
	if err := ValidateHash("payments capacity forecast channel", forecast.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments capacity forecast from", forecast.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments capacity forecast to", forecast.To); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments capacity forecast available", forecast.AvailableCapacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments capacity forecast reserved", forecast.ReservedCapacity); err != nil {
		return err
	}
	if forecast.ReservePressureBps > MaxPenaltyRouteBps {
		return errors.New("payments capacity forecast pressure exceeds 10000")
	}
	if forecast.ForecastHeight == 0 || forecast.ExpiresHeight <= forecast.ForecastHeight {
		return errors.New("payments capacity forecast validity window must advance")
	}
	return ValidateHash("payments capacity forecast hash", forecast.ForecastHash)
}

func (s LiquidityScore) Normalize() LiquidityScore {
	s.ScoreID = normalizeOptionalHash(s.ScoreID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.From = strings.TrimSpace(s.From)
	s.To = strings.TrimSpace(s.To)
	s.ScoreHash = normalizeOptionalHash(s.ScoreHash)
	return s
}

func (s LiquidityScore) WithHash() LiquidityScore {
	s = s.Normalize()
	s.ScoreHash = HashParts("liquidity-score", s.ScoreID, s.ChannelID, s.From, s.To, fmt.Sprintf("%020d", s.RawScore), fmt.Sprintf("%020d", s.Score), fmt.Sprintf("%010d", s.DecayBps), fmt.Sprintf("%020d", s.UpdatedHeight), fmt.Sprintf("%020d", s.ExpiresHeight))
	return s.Normalize()
}

func (s LiquidityScore) Validate() error {
	score := s.Normalize()
	if err := ValidateHash("payments liquidity score id", score.ScoreID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity score channel", score.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity score from", score.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity score to", score.To); err != nil {
		return err
	}
	if score.DecayBps > MaxPenaltyRouteBps {
		return errors.New("payments liquidity score decay exceeds 10000")
	}
	if score.UpdatedHeight == 0 || score.ExpiresHeight <= score.UpdatedHeight {
		return errors.New("payments liquidity score validity window must advance")
	}
	return ValidateHash("payments liquidity score hash", score.ScoreHash)
}

func (l LiquidityLimits) Normalize() LiquidityLimits {
	l.LimitID = normalizeOptionalHash(l.LimitID)
	l.ChannelID = normalizeHash(l.ChannelID)
	l.Participant = strings.TrimSpace(l.Participant)
	l.MaxReservedCapacity = strings.TrimSpace(l.MaxReservedCapacity)
	l.MinAvailableCapacity = strings.TrimSpace(l.MinAvailableCapacity)
	l.MaxBaseFee = strings.TrimSpace(l.MaxBaseFee)
	l.MaxReservationFee = strings.TrimSpace(l.MaxReservationFee)
	l.MaxVirtualSetupFee = strings.TrimSpace(l.MaxVirtualSetupFee)
	l.LimitHash = normalizeOptionalHash(l.LimitHash)
	for _, field := range []*string{&l.MaxReservedCapacity, &l.MinAvailableCapacity, &l.MaxBaseFee, &l.MaxReservationFee, &l.MaxVirtualSetupFee} {
		if *field == "" {
			*field = "0"
		}
	}
	return l
}

func (l LiquidityLimits) WithHash() LiquidityLimits {
	l = l.Normalize()
	if l.LimitID == "" {
		l.LimitID = HashParts("liquidity-limits", l.ChannelID, l.Participant)
	}
	l.LimitHash = HashParts("liquidity-limits", l.LimitID, l.ChannelID, l.Participant, l.MaxReservedCapacity, l.MinAvailableCapacity, l.MaxBaseFee, l.MaxReservationFee, l.MaxVirtualSetupFee, fmt.Sprintf("%010d", l.MaxProportionalBps), fmt.Sprintf("%010d", l.MaxRebalanceLoad), fmt.Sprintf("%020d", l.UpdatedHeight))
	return l.Normalize()
}

func (l LiquidityLimits) Validate() error {
	limits := l.Normalize()
	if err := ValidateHash("payments liquidity limits id", limits.LimitID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity limits channel", limits.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity limits participant", limits.Participant); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		amount	string
	}{
		{"payments liquidity max reserved", limits.MaxReservedCapacity},
		{"payments liquidity min available", limits.MinAvailableCapacity},
		{"payments liquidity max base fee", limits.MaxBaseFee},
		{"payments liquidity max reservation fee", limits.MaxReservationFee},
		{"payments liquidity max virtual setup fee", limits.MaxVirtualSetupFee},
	} {
		if err := validateNonNegativeInt(item.name, item.amount); err != nil {
			return err
		}
	}
	if limits.MaxProportionalBps > MaxPenaltyRouteBps {
		return errors.New("payments liquidity proportional fee bound exceeds 10000")
	}
	if limits.UpdatedHeight == 0 {
		return errors.New("payments liquidity limits height must be positive")
	}
	return ValidateHash("payments liquidity limits hash", limits.LimitHash)
}

func (b LiquidityFeePolicyBounds) Normalize() LiquidityFeePolicyBounds {
	b.MaxBaseFee = strings.TrimSpace(b.MaxBaseFee)
	b.MaxReservationFee = strings.TrimSpace(b.MaxReservationFee)
	b.MaxVirtualSetupFee = strings.TrimSpace(b.MaxVirtualSetupFee)
	b.MaxCongestionFee = strings.TrimSpace(b.MaxCongestionFee)
	b.MaxFailurePenalty = strings.TrimSpace(b.MaxFailurePenalty)
	b.MaxHopFee = strings.TrimSpace(b.MaxHopFee)
	return b
}

func (m MsgAdvertiseLiquidity) LiquidityOptimizationType() LiquidityOptimizationMessageType {
	return LiquidityMsgAdvertiseLiquidity
}

func (m MsgAdvertiseLiquidity) Normalize() MsgAdvertiseLiquidity {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Advertisement = m.Advertisement.Normalize()
	m.RequiredDeposit = strings.TrimSpace(m.RequiredDeposit)
	if m.RequiredDeposit == "" {
		m.RequiredDeposit = "0"
	}
	return m
}

func (m MsgAdvertiseLiquidity) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg advertise liquidity signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg advertise liquidity height must be positive")
	}
	return validateNonNegativeInt("payments msg advertise liquidity required deposit", msg.RequiredDeposit)
}

func (m MsgReserveLiquidity) LiquidityOptimizationType() LiquidityOptimizationMessageType {
	return LiquidityMsgReserveLiquidity
}

func (m MsgReserveLiquidity) Normalize() MsgReserveLiquidity {
	m.Reservation = m.Reservation.Normalize()
	return m
}

func (m MsgReserveLiquidity) ValidateBasic() error {
	if m.CurrentHeight == 0 {
		return errors.New("payments msg reserve liquidity height must be positive")
	}
	return m.Normalize().Reservation.Validate()
}

func (m MsgReleaseReservation) LiquidityOptimizationType() LiquidityOptimizationMessageType {
	return LiquidityMsgReleaseReservation
}

func (m MsgReleaseReservation) Normalize() MsgReleaseReservation {
	m.ReservationID = normalizeHash(m.ReservationID)
	m.ChannelID = normalizeHash(m.ChannelID)
	m.Releaser = strings.TrimSpace(m.Releaser)
	m.Reason = strings.TrimSpace(m.Reason)
	return m
}

func (m MsgReleaseReservation) ValidateBasic() error {
	msg := m.Normalize()
	if err := ValidateHash("payments msg release reservation id", msg.ReservationID); err != nil {
		return err
	}
	if err := ValidateHash("payments msg release reservation channel", msg.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments msg release reservation releaser", msg.Releaser); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 || msg.Reason == "" {
		return errors.New("payments msg release reservation requires height and reason")
	}
	return nil
}

func (m MsgUpdateFeePolicy) LiquidityOptimizationType() LiquidityOptimizationMessageType {
	return LiquidityMsgUpdateFeePolicy
}

func (m MsgUpdateFeePolicy) Normalize() MsgUpdateFeePolicy {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Policy = m.Policy.Normalize()
	m.Bounds = m.Bounds.Normalize()
	return m
}

func (m MsgUpdateFeePolicy) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg liquidity fee policy signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg liquidity fee policy height must be positive")
	}
	return msg.Policy.ValidateAtHeight(msg.CurrentHeight)
}

func (m MsgSubmitRebalanceIntent) LiquidityOptimizationType() LiquidityOptimizationMessageType {
	return LiquidityMsgSubmitRebalanceIntent
}

func (m MsgSubmitRebalanceIntent) Normalize() MsgSubmitRebalanceIntent {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Intent = m.Intent.Normalize()
	if m.Intent.SubmittedHeight == 0 {
		m.Intent.SubmittedHeight = m.CurrentHeight
	}
	return m
}

func (m MsgSubmitRebalanceIntent) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg rebalance signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg rebalance height must be positive")
	}
	return msg.Intent.WithHash().Validate()
}

func (m MsgSetLiquidityLimits) LiquidityOptimizationType() LiquidityOptimizationMessageType {
	return LiquidityMsgSetLiquidityLimits
}

func (m MsgSetLiquidityLimits) Normalize() MsgSetLiquidityLimits {
	m.Signer = strings.TrimSpace(m.Signer)
	m.Limits = m.Limits.Normalize()
	if m.Limits.UpdatedHeight == 0 {
		m.Limits.UpdatedHeight = m.CurrentHeight
	}
	return m
}

func (m MsgSetLiquidityLimits) ValidateBasic() error {
	msg := m.Normalize()
	if err := addressing.ValidateUserAddress("payments msg liquidity limits signer", msg.Signer); err != nil {
		return err
	}
	if msg.CurrentHeight == 0 {
		return errors.New("payments msg liquidity limits height must be positive")
	}
	return msg.Limits.WithHash().Validate()
}

func validateAdvertisementCapacity(chain PaymentsState, state LiquidityOptimizationState, ad LiquidityAdvertisement) error {
	channel, found := chain.ChannelByID(ad.ChannelID)
	if !found {
		return errors.New("payments liquidity advertisement channel not found")
	}
	available, err := spendableChannelBalance(channel, ad.Advertiser)
	if err != nil {
		return err
	}
	capacity, err := parsePositiveInt("payments liquidity advertisement capacity", ad.Capacity)
	if err != nil {
		return err
	}
	if capacity.GT(available) {
		return errors.New("payments liquidity advertisement exceeds spendable channel capacity")
	}
	limit := state.LimitsFor(ad.ChannelID, ad.Advertiser)
	if limit.MaxReservedCapacity != "" && limit.MaxReservedCapacity != "0" {
		maxReserved, err := parseNonNegativeInt("payments liquidity max reserved", limit.MaxReservedCapacity)
		if err != nil {
			return err
		}
		if capacity.GT(maxReserved) {
			return errors.New("payments liquidity advertisement exceeds participant limit")
		}
	}
	return nil
}

func validateReservationCapacity(chain PaymentsState, state LiquidityOptimizationState, position LiquidityPosition, reservation SignedLiquidityReservation, limits LiquidityLimits, currentHeight uint64) error {
	channel, found := chain.ChannelByID(position.ChannelID)
	if !found {
		return errors.New("payments liquidity reservation channel not found")
	}
	reservationAmount, err := parsePositiveInt("payments liquidity reservation capacity", reservation.Capacity)
	if err != nil {
		return err
	}
	active, err := activeReservedCapacity(state, position.ChannelID, position.FeePolicyID, currentHeight)
	if err != nil {
		return err
	}
	total := active.Add(reservationAmount)
	advertised, err := parsePositiveInt("payments liquidity advertised capacity", position.AdvertisedCapacity)
	if err != nil {
		return err
	}
	if total.GT(advertised) {
		return errors.New("payments liquidity over-reservation exceeds advertised capacity")
	}
	spendable, err := spendableChannelBalance(channel, position.Owner)
	if err != nil {
		return err
	}
	if total.GT(spendable) {
		return errors.New("payments liquidity over-reservation blocks usable channel capacity")
	}
	if limits.MaxReservedCapacity != "" && limits.MaxReservedCapacity != "0" {
		maxReserved, err := parseNonNegativeInt("payments liquidity max reserved", limits.MaxReservedCapacity)
		if err != nil {
			return err
		}
		if total.GT(maxReserved) {
			return errors.New("payments liquidity reservation exceeds participant limit")
		}
	}
	return nil
}

func spendableChannelBalance(channel ChannelRecord, participant string) (sdkmath.Int, error) {
	channel = channel.Normalize()
	participant = strings.TrimSpace(participant)
	for _, balance := range channel.LatestState.Balances {
		if balance.Participant != participant {
			continue
		}
		amount, err := parseNonNegativeInt("payments liquidity participant balance", balance.Amount)
		if err != nil {
			return sdkmath.Int{}, err
		}
		return amount, nil
	}
	return sdkmath.ZeroInt(), nil
}

func activeReservedCapacity(state LiquidityOptimizationState, channelID, advertisementID string, currentHeight uint64) (sdkmath.Int, error) {
	state = state.Normalize()
	channelID = normalizeHash(channelID)
	advertisementID = normalizeHash(advertisementID)
	total := sdkmath.ZeroInt()
	for _, reservation := range state.Reservations {
		reservation = reservation.Normalize()
		if reservation.ChannelID != channelID || reservation.AdvertisementID != advertisementID || reservation.Released || reservation.ExpirationHeight <= currentHeight {
			continue
		}
		amount, err := parsePositiveInt("payments active liquidity reservation", reservation.Capacity)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(amount)
	}
	return total, nil
}

func activeReservedCapacityForEndpoint(state LiquidityOptimizationState, channelID, owner string, currentHeight uint64) (sdkmath.Int, error) {
	state = state.Normalize()
	total := sdkmath.ZeroInt()
	for _, position := range state.Positions {
		if position.ChannelID != normalizeHash(channelID) || position.Owner != strings.TrimSpace(owner) {
			continue
		}
		reserved, err := activeReservedCapacity(state, position.ChannelID, position.FeePolicyID, currentHeight)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(reserved)
	}
	return total, nil
}

func capacityMinus(total, reserved string) (string, error) {
	totalAmount, err := parseNonNegativeInt("payments liquidity total capacity", total)
	if err != nil {
		return "", err
	}
	reservedAmount, err := parseNonNegativeInt("payments liquidity reserved capacity", reserved)
	if err != nil {
		return "", err
	}
	if reservedAmount.GT(totalAmount) {
		return "", errors.New("payments liquidity reserved capacity exceeds total capacity")
	}
	return totalAmount.Sub(reservedAmount).String(), nil
}

func requireAmountAtMost(name, value, bound string) error {
	valueAmount, err := parseNonNegativeInt(name, value)
	if err != nil {
		return err
	}
	bound = strings.TrimSpace(bound)
	if bound == "" || bound == "0" {
		return nil
	}
	boundAmount, err := parseNonNegativeInt(name+" bound", bound)
	if err != nil {
		return err
	}
	if valueAmount.GT(boundAmount) {
		return fmt.Errorf("%s exceeds bounds", name)
	}
	return nil
}

func upsertLiquidityPosition(values []LiquidityPosition, position LiquidityPosition) []LiquidityPosition {
	position = position.Normalize()
	out := append([]LiquidityPosition(nil), values...)
	for i, existing := range out {
		existing = existing.Normalize()
		if existing.PositionID == position.PositionID || (existing.ChannelID == position.ChannelID && existing.Owner == position.Owner && existing.Counterparty == position.Counterparty) {
			out[i] = position
			return normalizeLiquidityPositions(out)
		}
	}
	out = append(out, position)
	return normalizeLiquidityPositions(out)
}

func upsertReservation(values []Reservation, reservation Reservation) []Reservation {
	reservation = reservation.Normalize()
	out := append([]Reservation(nil), values...)
	for i, existing := range out {
		if existing.Normalize().ReservationID == reservation.ReservationID {
			out[i] = reservation
			return normalizeReservations(out)
		}
	}
	out = append(out, reservation)
	return normalizeReservations(out)
}

func upsertRebalanceIntent(values []RebalanceIntent, intent RebalanceIntent) []RebalanceIntent {
	intent = intent.Normalize()
	out := append([]RebalanceIntent(nil), values...)
	for i, existing := range out {
		if existing.Normalize().IntentID == intent.IntentID {
			out[i] = intent
			return normalizeRebalanceIntents(out)
		}
	}
	out = append(out, intent)
	return normalizeRebalanceIntents(out)
}

func upsertCapacityForecast(values []CapacityForecast, forecast CapacityForecast) []CapacityForecast {
	forecast = forecast.Normalize()
	out := append([]CapacityForecast(nil), values...)
	for i, existing := range out {
		existing = existing.Normalize()
		if existing.ChannelID == forecast.ChannelID && existing.From == forecast.From && existing.To == forecast.To {
			out[i] = forecast
			return normalizeCapacityForecasts(out)
		}
	}
	out = append(out, forecast)
	return normalizeCapacityForecasts(out)
}

func upsertLiquidityScore(values []LiquidityScore, score LiquidityScore) []LiquidityScore {
	score = score.Normalize()
	out := append([]LiquidityScore(nil), values...)
	for i, existing := range out {
		if existing.Normalize().ScoreID == score.ScoreID {
			out[i] = score
			return normalizeLiquidityScores(out)
		}
	}
	out = append(out, score)
	return normalizeLiquidityScores(out)
}

func upsertFeePolicy(values []FeePolicy, policy FeePolicy) []FeePolicy {
	policy = policy.Normalize()
	out := append([]FeePolicy(nil), values...)
	for i, existing := range out {
		if existing.Normalize().PolicyID == policy.PolicyID {
			out[i] = policy
			return normalizeFeePolicies(out)
		}
	}
	out = append(out, policy)
	return normalizeFeePolicies(out)
}

func upsertLiquidityLimits(values []LiquidityLimits, limits LiquidityLimits) []LiquidityLimits {
	limits = limits.Normalize()
	out := append([]LiquidityLimits(nil), values...)
	for i, existing := range out {
		existing = existing.Normalize()
		if existing.ChannelID == limits.ChannelID && existing.Participant == limits.Participant {
			out[i] = limits
			return normalizeLiquidityLimits(out)
		}
	}
	out = append(out, limits)
	return normalizeLiquidityLimits(out)
}

func normalizeLiquidityPositions(values []LiquidityPosition) []LiquidityPosition {
	out := make([]LiquidityPosition, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PositionID < out[j].PositionID })
	return out
}

func normalizeReservations(values []Reservation) []Reservation {
	out := make([]Reservation, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReservationID < out[j].ReservationID })
	return out
}

func normalizeRebalanceIntents(values []RebalanceIntent) []RebalanceIntent {
	out := make([]RebalanceIntent, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].IntentID < out[j].IntentID })
	return out
}

func normalizeCapacityForecasts(values []CapacityForecast) []CapacityForecast {
	out := make([]CapacityForecast, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ForecastID < out[j].ForecastID })
	return out
}

func normalizeLiquidityScores(values []LiquidityScore) []LiquidityScore {
	out := make([]LiquidityScore, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ScoreID < out[j].ScoreID })
	return out
}

func normalizeLiquidityLimits(values []LiquidityLimits) []LiquidityLimits {
	out := make([]LiquidityLimits, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].LimitID < out[j].LimitID })
	return out
}
