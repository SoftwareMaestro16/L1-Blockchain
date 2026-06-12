package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type NativePaymentGoal string

const (
	NativePaymentGoalFastTransfers		NativePaymentGoal	= "fast_offchain_style_transfers"
	NativePaymentGoalChannelSettlement	NativePaymentGoal	= "channel_like_settlement"
	NativePaymentGoalConditionalPayments	NativePaymentGoal	= "conditional_payments"
	NativePaymentGoalMultiZoneRouting	NativePaymentGoal	= "multi_zone_multi_shard_routing"
	NativePaymentGoalRouteFeeOptimization	NativePaymentGoal	= "route_fee_optimization"
	NativePaymentGoalTrustlessFallback	NativePaymentGoal	= "trustless_fallback_settlement"
)

type NativePaymentObject string

const (
	NativePaymentObjectChannel		NativePaymentObject	= "PaymentChannel"
	NativePaymentObjectVirtualChannel	NativePaymentObject	= "VirtualPaymentChannel"
	NativePaymentObjectCondition		NativePaymentObject	= "ConditionalPayment"
	NativePaymentObjectRoute		NativePaymentObject	= "PaymentRoute"
	NativePaymentObjectReservation		NativePaymentObject	= "LiquidityReservation"
	NativePaymentObjectSettlementProof	NativePaymentObject	= "SettlementProof"
	NativePaymentObjectReceipt		NativePaymentObject	= "PaymentReceipt"
)

type PaymentCommittedRoot string

const (
	PaymentCommittedRootChannel		PaymentCommittedRoot	= "channel_root"
	PaymentCommittedRootVirtualChannel	PaymentCommittedRoot	= "virtual_channel_root"
	PaymentCommittedRootCondition		PaymentCommittedRoot	= "condition_root"
	PaymentCommittedRootRoute		PaymentCommittedRoot	= "route_root"
	PaymentCommittedRootReservation		PaymentCommittedRoot	= "reservation_root"
	PaymentCommittedRootSettlementProof	PaymentCommittedRoot	= "settlement_proof_root"
	PaymentCommittedRootReceipt		PaymentCommittedRoot	= "payment_receipt_root"
)

type PaymentReceiptStatus string

const (
	PaymentReceiptExecuted	PaymentReceiptStatus	= "EXECUTED"
	PaymentReceiptFailed	PaymentReceiptStatus	= "FAILED"
	PaymentReceiptRefunded	PaymentReceiptStatus	= "REFUNDED"
	PaymentReceiptExpired	PaymentReceiptStatus	= "EXPIRED"
	PaymentReceiptSettled	PaymentReceiptStatus	= "SETTLED"
)

type SettlementProofType string

const (
	SettlementProofLatestState		SettlementProofType	= "LATEST_STATE"
	SettlementProofFraud			SettlementProofType	= "FRAUD_PROOF"
	SettlementProofCloseAuthorization	SettlementProofType	= "CLOSE_AUTHORIZATION"
	SettlementProofFallbackSettlement	SettlementProofType	= "FALLBACK_SETTLEMENT"
)

type NativePaymentGoalDescriptor struct {
	Goal		NativePaymentGoal
	Requirement	string
}

type NativePaymentAbstractionDescriptor struct {
	Object		NativePaymentObject
	Purpose		string
	CommittedRoot	PaymentCommittedRoot
}

type PaymentChannel struct {
	ChannelID		string
	ChainID			string
	ZoneID			string
	ShardID			uint32
	Participants		[]string
	Denom			string
	Collateral		string
	LatestStateHash		string
	LatestNonce		uint64
	ChallengePeriod		uint64
	SettlementStatus	ChannelStatus
	ChannelRoot		string
}

type VirtualPaymentChannel struct {
	VirtualChannelID	string
	ChainID			string
	ZoneID			string
	ShardID			uint32
	ParentRouteID		string
	ParentChannelIDs	[]string
	Endpoints		[]string
	Intermediaries		[]string
	Capacity		string
	BalanceA		string
	BalanceB		string
	RoutingFeeAmount	string
	ExpiresHeight		uint64
	Status			VirtualChannelStatus
	StateHash		string
	VirtualRoot		string
}

type PaymentRoute struct {
	RouteID		string
	Payer		string
	Payee		string
	Amount		string
	MaxFee		string
	Hops		[]PaymentRouteHop
	Capacity	string
	ExpiryHeight	uint64
	ConditionRoot	string
	RouteCommitment	string
	RouteRoot	string
}

type LiquidityReservation struct {
	ReservationID	string
	RouteID		string
	ChannelID	string
	Participant	string
	Amount		string
	FeeAmount	string
	ReservedHeight	uint64
	ExpiryHeight	uint64
	ReservationRoot	string
}

type SettlementProof struct {
	ProofID			string
	ProofType		SettlementProofType
	ChannelID		string
	LatestStateHash		string
	FraudProofHashOptional	string
	CloseAuthHashOptional	string
	FallbackRootOptional	string
	SubmittedBy		string
	Height			uint64
	ProofRoot		string
}

type PaymentReceipt struct {
	PaymentID	string
	RouteID		string
	ChannelID	string
	Status		PaymentReceiptStatus
	Amount		string
	FeeAmount	string
	ValueReturned	string
	Height		uint64
	ReceiptHash	string
}

type NativePaymentRoutingSnapshot struct {
	Height			uint64
	Channels		[]PaymentChannel
	VirtualChannels		[]VirtualPaymentChannel
	Conditions		[]ConditionalPayment
	Routes			[]PaymentRoute
	Reservations		[]LiquidityReservation
	SettlementProofs	[]SettlementProof
	Receipts		[]PaymentReceipt
	ChannelRoot		string
	VirtualChannelRoot	string
	ConditionRoot		string
	RouteRoot		string
	ReservationRoot		string
	SettlementProofRoot	string
	ReceiptRoot		string
	StateRoot		string
}

func NativePaymentGoals() []NativePaymentGoalDescriptor {
	return []NativePaymentGoalDescriptor{
		{Goal: NativePaymentGoalFastTransfers, Requirement: "signed channel state or local execution can advance value transfer before final settlement"},
		{Goal: NativePaymentGoalChannelSettlement, Requirement: "collateral, latest state, challenge windows, and close status are committed and proofable"},
		{Goal: NativePaymentGoalConditionalPayments, Requirement: "hash locks, time locks, promise conditions, and chained conditions resolve deterministically"},
		{Goal: NativePaymentGoalMultiZoneRouting, Requirement: "routes use committed route state and message receipts across zones and shards"},
		{Goal: NativePaymentGoalRouteFeeOptimization, Requirement: "fees are selected from committed liquidity, capacity, and congestion metrics"},
		{Goal: NativePaymentGoalTrustlessFallback, Requirement: "disputes and final settlement can fall back to Aether Core or Financial Zone proofs"},
	}
}

func NativePaymentAbstractions() []NativePaymentAbstractionDescriptor {
	return []NativePaymentAbstractionDescriptor{
		{Object: NativePaymentObjectChannel, Purpose: "locks collateral and tracks participant balances, latest state, challenge period, and settlement status", CommittedRoot: PaymentCommittedRootChannel},
		{Object: NativePaymentObjectVirtualChannel, Purpose: "routes payment capacity through one or more underlying channels", CommittedRoot: PaymentCommittedRootVirtualChannel},
		{Object: NativePaymentObjectCondition, Purpose: "represents hash-lock, time-lock, promise, or chained settlement condition", CommittedRoot: PaymentCommittedRootCondition},
		{Object: NativePaymentObjectRoute, Purpose: "describes hop sequence, fees, capacity, expiry, and route commitment", CommittedRoot: PaymentCommittedRootRoute},
		{Object: NativePaymentObjectReservation, Purpose: "temporarily reserves capacity for a route or condition", CommittedRoot: PaymentCommittedRootReservation},
		{Object: NativePaymentObjectSettlementProof, Purpose: "proves latest state, fraud proof, close authorization, or fallback settlement", CommittedRoot: PaymentCommittedRootSettlementProof},
		{Object: NativePaymentObjectReceipt, Purpose: "records payment execution, failure, refund, expiry, or final settlement", CommittedRoot: PaymentCommittedRootReceipt},
	}
}

func ValidateNativePaymentArchitectureDescriptors() error {
	seenGoals := map[NativePaymentGoal]struct{}{}
	for _, goal := range NativePaymentGoals() {
		if goal.Goal == "" || strings.TrimSpace(goal.Requirement) == "" {
			return errors.New("payments native goal descriptor is incomplete")
		}
		if _, found := seenGoals[goal.Goal]; found {
			return fmt.Errorf("payments duplicate native goal %s", goal.Goal)
		}
		seenGoals[goal.Goal] = struct{}{}
	}
	seenObjects := map[NativePaymentObject]struct{}{}
	seenRoots := map[PaymentCommittedRoot]struct{}{}
	for _, abstraction := range NativePaymentAbstractions() {
		if abstraction.Object == "" || abstraction.CommittedRoot == "" || strings.TrimSpace(abstraction.Purpose) == "" {
			return errors.New("payments native abstraction descriptor is incomplete")
		}
		if _, found := seenObjects[abstraction.Object]; found {
			return fmt.Errorf("payments duplicate native abstraction %s", abstraction.Object)
		}
		if _, found := seenRoots[abstraction.CommittedRoot]; found {
			return fmt.Errorf("payments duplicate committed root %s", abstraction.CommittedRoot)
		}
		seenObjects[abstraction.Object] = struct{}{}
		seenRoots[abstraction.CommittedRoot] = struct{}{}
	}
	return nil
}

func NewPaymentChannelFromRecord(channel ChannelRecord, zoneID string, shardID uint32) (PaymentChannel, error) {
	channel = channel.Normalize()
	if err := channel.Validate(); err != nil {
		return PaymentChannel{}, err
	}
	return BuildPaymentChannel(PaymentChannel{
		ChannelID:		channel.ChannelID,
		ChainID:		channel.ChainID,
		ZoneID:			zoneID,
		ShardID:		shardID,
		Participants:		channel.Participants,
		Denom:			channel.Denom,
		Collateral:		channel.Collateral,
		LatestStateHash:	channel.LatestState.StateHash,
		LatestNonce:		channel.LatestState.Nonce,
		ChallengePeriod:	channel.DisputePeriod,
		SettlementStatus:	channel.Status,
	})
}

func BuildPaymentChannel(channel PaymentChannel) (PaymentChannel, error) {
	channel = channel.Normalize()
	if channel.ChannelRoot != "" {
		return PaymentChannel{}, errors.New("payments channel root must be empty before construction")
	}
	if err := channel.ValidateFormat(); err != nil {
		return PaymentChannel{}, err
	}
	channel.ChannelRoot = ComputePaymentChannelRoot(channel)
	return channel, channel.Validate()
}

func (channel PaymentChannel) Normalize() PaymentChannel {
	channel.ChannelID = normalizeHash(channel.ChannelID)
	channel.ChainID = strings.TrimSpace(channel.ChainID)
	channel.ZoneID = strings.TrimSpace(channel.ZoneID)
	channel.Participants = normalizeAddressSet(channel.Participants)
	channel.Denom = normalizeAssetDenom(channel.Denom)
	channel.Collateral = strings.TrimSpace(channel.Collateral)
	channel.LatestStateHash = normalizeHash(channel.LatestStateHash)
	channel.ChannelRoot = normalizeOptionalHash(channel.ChannelRoot)
	return channel
}

func (channel PaymentChannel) ValidateFormat() error {
	channel = channel.Normalize()
	if err := ValidateHash("payments native channel id", channel.ChannelID); err != nil {
		return err
	}
	if strings.TrimSpace(channel.ChainID) == "" {
		return errors.New("payments native channel chain id is required")
	}
	if err := validatePaymentRoutingToken("payments native channel zone id", channel.ZoneID); err != nil {
		return err
	}
	if err := validateAddressSet("payments native channel participant", channel.Participants, 2, MaxParticipants); err != nil {
		return err
	}
	if channel.Denom != NativeDenom {
		return fmt.Errorf("payments native channel denom must be %s", NativeDenom)
	}
	if err := validatePositiveInt("payments native channel collateral", channel.Collateral); err != nil {
		return err
	}
	if err := ValidateHash("payments native channel latest state", channel.LatestStateHash); err != nil {
		return err
	}
	if channel.LatestNonce == 0 {
		return errors.New("payments native channel latest nonce must be positive")
	}
	if channel.ChallengePeriod == 0 {
		return errors.New("payments native channel challenge period must be positive")
	}
	if !IsChannelStatus(channel.SettlementStatus) {
		return fmt.Errorf("unknown payments native channel settlement status %q", channel.SettlementStatus)
	}
	if channel.ChannelRoot != "" {
		return ValidateHash("payments native channel root", channel.ChannelRoot)
	}
	return nil
}

func (channel PaymentChannel) Validate() error {
	channel = channel.Normalize()
	if err := channel.ValidateFormat(); err != nil {
		return err
	}
	if channel.ChannelRoot == "" {
		return errors.New("payments native channel root is required")
	}
	if expected := ComputePaymentChannelRoot(channel); channel.ChannelRoot != expected {
		return fmt.Errorf("payments native channel root mismatch: expected %s", expected)
	}
	return nil
}

func NewVirtualPaymentChannelFromRecord(channel VirtualChannel, zoneID string, shardID uint32) (VirtualPaymentChannel, error) {
	channel = channel.Normalize()
	if err := channel.Validate(); err != nil {
		return VirtualPaymentChannel{}, err
	}
	return BuildVirtualPaymentChannel(VirtualPaymentChannel{
		VirtualChannelID:	channel.VirtualChannelID,
		ChainID:		channel.ChainID,
		ZoneID:			zoneID,
		ShardID:		shardID,
		ParentRouteID:		channel.ParentRouteID,
		ParentChannelIDs:	channel.ParentChannelIDs,
		Endpoints:		channel.Endpoints,
		Intermediaries:		channel.Intermediaries,
		Capacity:		channel.Capacity,
		BalanceA:		channel.BalanceA,
		BalanceB:		channel.BalanceB,
		RoutingFeeAmount:	channel.RoutingFeeAmount,
		ExpiresHeight:		channel.ExpiresHeight,
		Status:			channel.Status,
		StateHash:		channel.StateHash,
	})
}

func BuildVirtualPaymentChannel(channel VirtualPaymentChannel) (VirtualPaymentChannel, error) {
	channel = channel.Normalize()
	if channel.VirtualRoot != "" {
		return VirtualPaymentChannel{}, errors.New("payments virtual payment channel root must be empty before construction")
	}
	if err := channel.ValidateFormat(); err != nil {
		return VirtualPaymentChannel{}, err
	}
	channel.VirtualRoot = ComputeVirtualPaymentChannelRoot(channel)
	return channel, channel.Validate()
}

func (channel VirtualPaymentChannel) Normalize() VirtualPaymentChannel {
	channel.VirtualChannelID = normalizeHash(channel.VirtualChannelID)
	channel.ChainID = strings.TrimSpace(channel.ChainID)
	channel.ZoneID = strings.TrimSpace(channel.ZoneID)
	channel.ParentRouteID = normalizeOptionalHash(channel.ParentRouteID)
	channel.ParentChannelIDs = normalizeHashSlice(channel.ParentChannelIDs)
	channel.Endpoints = normalizeAddressSet(channel.Endpoints)
	channel.Intermediaries = normalizeAddressSet(channel.Intermediaries)
	channel.Capacity = strings.TrimSpace(channel.Capacity)
	channel.BalanceA = strings.TrimSpace(channel.BalanceA)
	channel.BalanceB = strings.TrimSpace(channel.BalanceB)
	channel.RoutingFeeAmount = strings.TrimSpace(channel.RoutingFeeAmount)
	channel.StateHash = normalizeHash(channel.StateHash)
	channel.VirtualRoot = normalizeOptionalHash(channel.VirtualRoot)
	return channel
}

func (channel VirtualPaymentChannel) ValidateFormat() error {
	channel = channel.Normalize()
	if err := ValidateHash("payments virtual payment channel id", channel.VirtualChannelID); err != nil {
		return err
	}
	if strings.TrimSpace(channel.ChainID) == "" {
		return errors.New("payments virtual payment channel chain id is required")
	}
	if err := validatePaymentRoutingToken("payments virtual payment channel zone id", channel.ZoneID); err != nil {
		return err
	}
	if len(channel.ParentChannelIDs) == 0 || len(channel.ParentChannelIDs) > MaxParentChannels {
		return fmt.Errorf("payments virtual payment channel parent count must be between 1 and %d", MaxParentChannels)
	}
	for _, parentID := range channel.ParentChannelIDs {
		if err := ValidateHash("payments virtual payment parent channel id", parentID); err != nil {
			return err
		}
	}
	if err := validateAddressSet("payments virtual payment endpoint", channel.Endpoints, 2, 2); err != nil {
		return err
	}
	if len(channel.Intermediaries) > MaxParticipants {
		return fmt.Errorf("payments virtual payment intermediaries must be <= %d", MaxParticipants)
	}
	if err := validatePositiveInt("payments virtual payment capacity", channel.Capacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual payment balance a", channel.BalanceA); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual payment balance b", channel.BalanceB); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual payment fee", channel.RoutingFeeAmount); err != nil {
		return err
	}
	if err := ValidateHash("payments virtual payment state hash", channel.StateHash); err != nil {
		return err
	}
	if channel.ExpiresHeight == 0 {
		return errors.New("payments virtual payment expiry height must be positive")
	}
	if !IsVirtualChannelStatus(channel.Status) {
		return fmt.Errorf("unknown payments virtual payment status %q", channel.Status)
	}
	if channel.VirtualRoot != "" {
		return ValidateHash("payments virtual payment root", channel.VirtualRoot)
	}
	return nil
}

func (channel VirtualPaymentChannel) Validate() error {
	channel = channel.Normalize()
	if err := channel.ValidateFormat(); err != nil {
		return err
	}
	if channel.VirtualRoot == "" {
		return errors.New("payments virtual payment root is required")
	}
	if expected := ComputeVirtualPaymentChannelRoot(channel); channel.VirtualRoot != expected {
		return fmt.Errorf("payments virtual payment root mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentRoute(route PaymentRoute) (PaymentRoute, error) {
	route = route.Normalize()
	if route.RouteRoot != "" {
		return PaymentRoute{}, errors.New("payments native route root must be empty before construction")
	}
	if err := route.ValidateFormat(); err != nil {
		return PaymentRoute{}, err
	}
	route.RouteRoot = ComputeNativePaymentRouteRoot(route)
	return route, route.Validate()
}

func NewPaymentRouteFromMsg(msg MsgPaymentRoute, capacity string) (PaymentRoute, error) {
	msg = msg.Normalize()
	if err := msg.ValidateBasic(); err != nil {
		return PaymentRoute{}, err
	}
	return BuildPaymentRoute(PaymentRoute{
		RouteID:		msg.RouteID,
		Payer:			msg.Payer,
		Payee:			msg.Payee,
		Amount:			msg.Amount,
		MaxFee:			msg.MaxFee,
		Hops:			msg.Hops,
		Capacity:		capacity,
		ExpiryHeight:		msg.ExpiryHeight,
		ConditionRoot:		msg.ConditionRoot,
		RouteCommitment:	ComputePaymentRouteCommitmentHash(msg),
	})
}

func (route PaymentRoute) Normalize() PaymentRoute {
	route.RouteID = normalizeHash(route.RouteID)
	route.Payer = strings.TrimSpace(route.Payer)
	route.Payee = strings.TrimSpace(route.Payee)
	route.Amount = strings.TrimSpace(route.Amount)
	route.MaxFee = strings.TrimSpace(route.MaxFee)
	for i := range route.Hops {
		route.Hops[i] = route.Hops[i].Normalize()
	}
	route.Capacity = strings.TrimSpace(route.Capacity)
	route.ConditionRoot = normalizeHash(route.ConditionRoot)
	route.RouteCommitment = normalizeHash(route.RouteCommitment)
	route.RouteRoot = normalizeOptionalHash(route.RouteRoot)
	return route
}

func (route PaymentRoute) ValidateFormat() error {
	route = route.Normalize()
	if err := ValidateHash("payments native route id", route.RouteID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments native route payer", route.Payer); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments native route payee", route.Payee); err != nil {
		return err
	}
	if route.Payer == route.Payee {
		return errors.New("payments native route endpoints must differ")
	}
	amount, err := parsePositiveInt("payments native route amount", route.Amount)
	if err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments native route max fee", route.MaxFee); err != nil {
		return err
	}
	capacity, err := parsePositiveInt("payments native route capacity", route.Capacity)
	if err != nil {
		return err
	}
	if capacity.LT(amount) {
		return errors.New("payments native route capacity must cover amount")
	}
	if len(route.Hops) == 0 || len(route.Hops) > MaxRoutingHops {
		return fmt.Errorf("payments native route hops must be between 1 and %d", MaxRoutingHops)
	}
	msg := MsgPaymentRoute{
		RouteID:	route.RouteID,
		Payer:		route.Payer,
		Payee:		route.Payee,
		Amount:		route.Amount,
		MaxFee:		route.MaxFee,
		Hops:		route.Hops,
		ConditionRoot:	route.ConditionRoot,
		ExpiryHeight:	route.ExpiryHeight,
		SettlementMode:	ConditionSettlementModePreimage,
	}
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if err := ValidateHash("payments native route commitment", route.RouteCommitment); err != nil {
		return err
	}
	if expected := ComputePaymentRouteCommitmentHash(msg); route.RouteCommitment != expected {
		return errors.New("payments native route commitment mismatch")
	}
	if route.RouteRoot != "" {
		return ValidateHash("payments native route root", route.RouteRoot)
	}
	return nil
}

func (route PaymentRoute) Validate() error {
	route = route.Normalize()
	if err := route.ValidateFormat(); err != nil {
		return err
	}
	if route.RouteRoot == "" {
		return errors.New("payments native route root is required")
	}
	if expected := ComputeNativePaymentRouteRoot(route); route.RouteRoot != expected {
		return fmt.Errorf("payments native route root mismatch: expected %s", expected)
	}
	return nil
}

func BuildLiquidityReservation(reservation LiquidityReservation) (LiquidityReservation, error) {
	reservation = reservation.Normalize()
	if reservation.ReservationRoot != "" {
		return LiquidityReservation{}, errors.New("payments liquidity reservation root must be empty before construction")
	}
	if err := reservation.ValidateFormat(); err != nil {
		return LiquidityReservation{}, err
	}
	reservation.ReservationRoot = ComputeLiquidityReservationRoot(reservation)
	return reservation, reservation.Validate()
}

func (reservation LiquidityReservation) Normalize() LiquidityReservation {
	reservation.ReservationID = normalizeHash(reservation.ReservationID)
	reservation.RouteID = normalizeHash(reservation.RouteID)
	reservation.ChannelID = normalizeHash(reservation.ChannelID)
	reservation.Participant = strings.TrimSpace(reservation.Participant)
	reservation.Amount = strings.TrimSpace(reservation.Amount)
	reservation.FeeAmount = strings.TrimSpace(reservation.FeeAmount)
	reservation.ReservationRoot = normalizeOptionalHash(reservation.ReservationRoot)
	return reservation
}

func (reservation LiquidityReservation) ValidateFormat() error {
	reservation = reservation.Normalize()
	if err := ValidateHash("payments liquidity reservation id", reservation.ReservationID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity reservation route id", reservation.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity reservation channel id", reservation.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity reservation participant", reservation.Participant); err != nil {
		return err
	}
	if err := validatePositiveInt("payments liquidity reservation amount", reservation.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments liquidity reservation fee", reservation.FeeAmount); err != nil {
		return err
	}
	if reservation.ReservedHeight == 0 || reservation.ExpiryHeight <= reservation.ReservedHeight {
		return errors.New("payments liquidity reservation height range is invalid")
	}
	if reservation.ReservationRoot != "" {
		return ValidateHash("payments liquidity reservation root", reservation.ReservationRoot)
	}
	return nil
}

func (reservation LiquidityReservation) Validate() error {
	reservation = reservation.Normalize()
	if err := reservation.ValidateFormat(); err != nil {
		return err
	}
	if reservation.ReservationRoot == "" {
		return errors.New("payments liquidity reservation root is required")
	}
	if expected := ComputeLiquidityReservationRoot(reservation); reservation.ReservationRoot != expected {
		return fmt.Errorf("payments liquidity reservation root mismatch: expected %s", expected)
	}
	return nil
}

func BuildSettlementProof(proof SettlementProof) (SettlementProof, error) {
	proof = proof.Normalize()
	if proof.ProofRoot != "" {
		return SettlementProof{}, errors.New("payments settlement proof root must be empty before construction")
	}
	if err := proof.ValidateFormat(); err != nil {
		return SettlementProof{}, err
	}
	proof.ProofRoot = ComputeSettlementProofRoot(proof)
	return proof, proof.Validate()
}

func (proof SettlementProof) Normalize() SettlementProof {
	proof.ProofID = normalizeHash(proof.ProofID)
	proof.ChannelID = normalizeHash(proof.ChannelID)
	proof.LatestStateHash = normalizeHash(proof.LatestStateHash)
	proof.FraudProofHashOptional = normalizeOptionalHash(proof.FraudProofHashOptional)
	proof.CloseAuthHashOptional = normalizeOptionalHash(proof.CloseAuthHashOptional)
	proof.FallbackRootOptional = normalizeOptionalHash(proof.FallbackRootOptional)
	proof.SubmittedBy = strings.TrimSpace(proof.SubmittedBy)
	proof.ProofRoot = normalizeOptionalHash(proof.ProofRoot)
	return proof
}

func (proof SettlementProof) ValidateFormat() error {
	proof = proof.Normalize()
	if err := ValidateHash("payments settlement proof id", proof.ProofID); err != nil {
		return err
	}
	if !IsSettlementProofType(proof.ProofType) {
		return fmt.Errorf("unknown payments settlement proof type %q", proof.ProofType)
	}
	if err := ValidateHash("payments settlement proof channel id", proof.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments settlement proof latest state hash", proof.LatestStateHash); err != nil {
		return err
	}
	if proof.FraudProofHashOptional != "" {
		if err := ValidateHash("payments settlement fraud proof hash", proof.FraudProofHashOptional); err != nil {
			return err
		}
	}
	if proof.CloseAuthHashOptional != "" {
		if err := ValidateHash("payments settlement close auth hash", proof.CloseAuthHashOptional); err != nil {
			return err
		}
	}
	if proof.FallbackRootOptional != "" {
		if err := ValidateHash("payments settlement fallback root", proof.FallbackRootOptional); err != nil {
			return err
		}
	}
	if err := addressing.ValidateUserAddress("payments settlement proof submitter", proof.SubmittedBy); err != nil {
		return err
	}
	if proof.Height == 0 {
		return errors.New("payments settlement proof height must be positive")
	}
	if proof.ProofType == SettlementProofFraud && proof.FraudProofHashOptional == "" {
		return errors.New("payments fraud settlement proof requires fraud proof hash")
	}
	if proof.ProofType == SettlementProofCloseAuthorization && proof.CloseAuthHashOptional == "" {
		return errors.New("payments close authorization proof requires auth hash")
	}
	if proof.ProofType == SettlementProofFallbackSettlement && proof.FallbackRootOptional == "" {
		return errors.New("payments fallback settlement proof requires fallback root")
	}
	if proof.ProofRoot != "" {
		return ValidateHash("payments settlement proof root", proof.ProofRoot)
	}
	return nil
}

func (proof SettlementProof) Validate() error {
	proof = proof.Normalize()
	if err := proof.ValidateFormat(); err != nil {
		return err
	}
	if proof.ProofRoot == "" {
		return errors.New("payments settlement proof root is required")
	}
	if expected := ComputeSettlementProofRoot(proof); proof.ProofRoot != expected {
		return fmt.Errorf("payments settlement proof root mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentReceipt(receipt PaymentReceipt) (PaymentReceipt, error) {
	receipt = receipt.Normalize()
	if receipt.ReceiptHash != "" {
		return PaymentReceipt{}, errors.New("payments native receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return PaymentReceipt{}, err
	}
	receipt.ReceiptHash = ComputeNativePaymentReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (receipt PaymentReceipt) Normalize() PaymentReceipt {
	receipt.PaymentID = normalizeHash(receipt.PaymentID)
	receipt.RouteID = normalizeOptionalHash(receipt.RouteID)
	receipt.ChannelID = normalizeOptionalHash(receipt.ChannelID)
	receipt.Amount = strings.TrimSpace(receipt.Amount)
	receipt.FeeAmount = strings.TrimSpace(receipt.FeeAmount)
	receipt.ValueReturned = strings.TrimSpace(receipt.ValueReturned)
	receipt.ReceiptHash = normalizeOptionalHash(receipt.ReceiptHash)
	return receipt
}

func (receipt PaymentReceipt) ValidateFormat() error {
	receipt = receipt.Normalize()
	if err := ValidateHash("payments native receipt payment id", receipt.PaymentID); err != nil {
		return err
	}
	if receipt.RouteID == "" && receipt.ChannelID == "" {
		return errors.New("payments native receipt requires route id or channel id")
	}
	if receipt.RouteID != "" {
		if err := ValidateHash("payments native receipt route id", receipt.RouteID); err != nil {
			return err
		}
	}
	if receipt.ChannelID != "" {
		if err := ValidateHash("payments native receipt channel id", receipt.ChannelID); err != nil {
			return err
		}
	}
	if !IsPaymentReceiptStatus(receipt.Status) {
		return fmt.Errorf("unknown payments native receipt status %q", receipt.Status)
	}
	if err := validatePositiveInt("payments native receipt amount", receipt.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments native receipt fee", receipt.FeeAmount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments native receipt value returned", receipt.ValueReturned); err != nil {
		return err
	}
	if receipt.Height == 0 {
		return errors.New("payments native receipt height must be positive")
	}
	if receipt.ReceiptHash != "" {
		return ValidateHash("payments native receipt hash", receipt.ReceiptHash)
	}
	return nil
}

func (receipt PaymentReceipt) Validate() error {
	receipt = receipt.Normalize()
	if err := receipt.ValidateFormat(); err != nil {
		return err
	}
	if receipt.ReceiptHash == "" {
		return errors.New("payments native receipt hash is required")
	}
	if expected := ComputeNativePaymentReceiptHash(receipt); receipt.ReceiptHash != expected {
		return fmt.Errorf("payments native receipt hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildNativePaymentRoutingSnapshot(snapshot NativePaymentRoutingSnapshot) (NativePaymentRoutingSnapshot, error) {
	snapshot = snapshot.Normalize()
	if snapshot.StateRoot != "" {
		return NativePaymentRoutingSnapshot{}, errors.New("payments native snapshot root must be empty before construction")
	}
	if snapshot.Height == 0 {
		return NativePaymentRoutingSnapshot{}, errors.New("payments native snapshot height must be positive")
	}
	var err error
	snapshot.Channels, err = buildPaymentChannels(snapshot.Channels)
	if err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	snapshot.VirtualChannels, err = buildVirtualPaymentChannels(snapshot.VirtualChannels)
	if err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	if err := validateConditions(snapshot.Conditions); err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	snapshot.Routes, err = buildNativePaymentRoutes(snapshot.Routes)
	if err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	snapshot.Reservations, err = buildLiquidityReservations(snapshot.Reservations)
	if err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	snapshot.SettlementProofs, err = buildSettlementProofs(snapshot.SettlementProofs)
	if err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	snapshot.Receipts, err = buildPaymentReceipts(snapshot.Receipts)
	if err != nil {
		return NativePaymentRoutingSnapshot{}, err
	}
	snapshot.ChannelRoot = ComputePaymentChannelSetRoot(snapshot.Channels)
	snapshot.VirtualChannelRoot = ComputeVirtualPaymentChannelSetRoot(snapshot.VirtualChannels)
	snapshot.ConditionRoot = ComputeConditionsRoot(snapshot.Conditions)
	snapshot.RouteRoot = ComputeNativePaymentRouteSetRoot(snapshot.Routes)
	snapshot.ReservationRoot = ComputeLiquidityReservationSetRoot(snapshot.Reservations)
	snapshot.SettlementProofRoot = ComputeSettlementProofSetRoot(snapshot.SettlementProofs)
	snapshot.ReceiptRoot = ComputeNativePaymentReceiptSetRoot(snapshot.Receipts)
	snapshot.StateRoot = ComputeNativePaymentRoutingSnapshotRoot(snapshot)
	return snapshot, snapshot.Validate()
}

func (snapshot NativePaymentRoutingSnapshot) Normalize() NativePaymentRoutingSnapshot {
	snapshot.Channels = normalizePaymentChannels(snapshot.Channels)
	snapshot.VirtualChannels = normalizeVirtualPaymentChannels(snapshot.VirtualChannels)
	snapshot.Conditions = normalizeConditions(snapshot.Conditions)
	snapshot.Routes = normalizeNativePaymentRoutes(snapshot.Routes)
	snapshot.Reservations = normalizeLiquidityReservations(snapshot.Reservations)
	snapshot.SettlementProofs = normalizeSettlementProofs(snapshot.SettlementProofs)
	snapshot.Receipts = normalizePaymentReceipts(snapshot.Receipts)
	snapshot.ChannelRoot = normalizeOptionalHash(snapshot.ChannelRoot)
	snapshot.VirtualChannelRoot = normalizeOptionalHash(snapshot.VirtualChannelRoot)
	snapshot.ConditionRoot = normalizeOptionalHash(snapshot.ConditionRoot)
	snapshot.RouteRoot = normalizeOptionalHash(snapshot.RouteRoot)
	snapshot.ReservationRoot = normalizeOptionalHash(snapshot.ReservationRoot)
	snapshot.SettlementProofRoot = normalizeOptionalHash(snapshot.SettlementProofRoot)
	snapshot.ReceiptRoot = normalizeOptionalHash(snapshot.ReceiptRoot)
	snapshot.StateRoot = normalizeOptionalHash(snapshot.StateRoot)
	return snapshot
}

func (snapshot NativePaymentRoutingSnapshot) Validate() error {
	snapshot = snapshot.Normalize()
	if snapshot.Height == 0 {
		return errors.New("payments native snapshot height must be positive")
	}
	if err := validatePaymentChannels(snapshot.Channels); err != nil {
		return err
	}
	if err := validateVirtualPaymentChannels(snapshot.VirtualChannels); err != nil {
		return err
	}
	if err := validateConditions(snapshot.Conditions); err != nil {
		return err
	}
	if err := validateNativePaymentRoutes(snapshot.Routes); err != nil {
		return err
	}
	if err := validateLiquidityReservations(snapshot.Reservations); err != nil {
		return err
	}
	if err := validateSettlementProofs(snapshot.SettlementProofs); err != nil {
		return err
	}
	if err := validatePaymentReceipts(snapshot.Receipts); err != nil {
		return err
	}
	expectedRoots := map[string][2]string{
		"payments native channel root":			{snapshot.ChannelRoot, ComputePaymentChannelSetRoot(snapshot.Channels)},
		"payments native virtual channel root":		{snapshot.VirtualChannelRoot, ComputeVirtualPaymentChannelSetRoot(snapshot.VirtualChannels)},
		"payments native condition root":		{snapshot.ConditionRoot, ComputeConditionsRoot(snapshot.Conditions)},
		"payments native route root":			{snapshot.RouteRoot, ComputeNativePaymentRouteSetRoot(snapshot.Routes)},
		"payments native reservation root":		{snapshot.ReservationRoot, ComputeLiquidityReservationSetRoot(snapshot.Reservations)},
		"payments native settlement proof root":	{snapshot.SettlementProofRoot, ComputeSettlementProofSetRoot(snapshot.SettlementProofs)},
		"payments native receipt root":			{snapshot.ReceiptRoot, ComputeNativePaymentReceiptSetRoot(snapshot.Receipts)},
	}
	for field, roots := range expectedRoots {
		if roots[0] == "" {
			return fmt.Errorf("%s is required", field)
		}
		if roots[0] != roots[1] {
			return fmt.Errorf("%s mismatch: expected %s", field, roots[1])
		}
	}
	if snapshot.StateRoot == "" {
		return errors.New("payments native snapshot root is required")
	}
	if expected := ComputeNativePaymentRoutingSnapshotRoot(snapshot); snapshot.StateRoot != expected {
		return fmt.Errorf("payments native snapshot root mismatch: expected %s", expected)
	}
	return nil
}

func RootForNativePaymentObject(snapshot NativePaymentRoutingSnapshot, object NativePaymentObject) (string, bool) {
	snapshot = snapshot.Normalize()
	switch object {
	case NativePaymentObjectChannel:
		return snapshot.ChannelRoot, snapshot.ChannelRoot != ""
	case NativePaymentObjectVirtualChannel:
		return snapshot.VirtualChannelRoot, snapshot.VirtualChannelRoot != ""
	case NativePaymentObjectCondition:
		return snapshot.ConditionRoot, snapshot.ConditionRoot != ""
	case NativePaymentObjectRoute:
		return snapshot.RouteRoot, snapshot.RouteRoot != ""
	case NativePaymentObjectReservation:
		return snapshot.ReservationRoot, snapshot.ReservationRoot != ""
	case NativePaymentObjectSettlementProof:
		return snapshot.SettlementProofRoot, snapshot.SettlementProofRoot != ""
	case NativePaymentObjectReceipt:
		return snapshot.ReceiptRoot, snapshot.ReceiptRoot != ""
	default:
		return "", false
	}
}

func ComputePaymentChannelRoot(channel PaymentChannel) string {
	channel = channel.Normalize()
	return HashParts(
		"aetra-native-payment-channel-v1",
		channel.ChannelID,
		channel.ChainID,
		channel.ZoneID,
		fmt.Sprintf("%020d", uint64(channel.ShardID)),
		strings.Join(channel.Participants, "|"),
		channel.Denom,
		channel.Collateral,
		channel.LatestStateHash,
		fmt.Sprintf("%020d", channel.LatestNonce),
		fmt.Sprintf("%020d", channel.ChallengePeriod),
		string(channel.SettlementStatus),
	)
}

func ComputeVirtualPaymentChannelRoot(channel VirtualPaymentChannel) string {
	channel = channel.Normalize()
	parts := []string{
		"aetra-native-virtual-payment-channel-v1",
		channel.VirtualChannelID,
		channel.ChainID,
		channel.ZoneID,
		fmt.Sprintf("%020d", uint64(channel.ShardID)),
		channel.ParentRouteID,
		strings.Join(channel.Endpoints, "|"),
		strings.Join(channel.Intermediaries, "|"),
		channel.Capacity,
		channel.BalanceA,
		channel.BalanceB,
		channel.RoutingFeeAmount,
		fmt.Sprintf("%020d", channel.ExpiresHeight),
		string(channel.Status),
		channel.StateHash,
	}
	parts = append(parts, channel.ParentChannelIDs...)
	return HashParts(parts...)
}

func ComputeNativePaymentRouteRoot(route PaymentRoute) string {
	route = route.Normalize()
	parts := []string{
		"aetra-native-payment-route-v1",
		route.RouteID,
		route.Payer,
		route.Payee,
		route.Amount,
		route.MaxFee,
		route.Capacity,
		fmt.Sprintf("%020d", route.ExpiryHeight),
		route.ConditionRoot,
		route.RouteCommitment,
	}
	for _, hop := range route.Hops {
		parts = append(parts, hop.ChannelID, hop.From, hop.To, hop.FeeAmount, fmt.Sprintf("%020d", hop.TimeoutHeight))
	}
	return HashParts(parts...)
}

func ComputeLiquidityReservationRoot(reservation LiquidityReservation) string {
	reservation = reservation.Normalize()
	return HashParts(
		"aetra-native-liquidity-reservation-v1",
		reservation.ReservationID,
		reservation.RouteID,
		reservation.ChannelID,
		reservation.Participant,
		reservation.Amount,
		reservation.FeeAmount,
		fmt.Sprintf("%020d", reservation.ReservedHeight),
		fmt.Sprintf("%020d", reservation.ExpiryHeight),
	)
}

func ComputeSettlementProofRoot(proof SettlementProof) string {
	proof = proof.Normalize()
	return HashParts(
		"aetra-native-settlement-proof-v1",
		proof.ProofID,
		string(proof.ProofType),
		proof.ChannelID,
		proof.LatestStateHash,
		proof.FraudProofHashOptional,
		proof.CloseAuthHashOptional,
		proof.FallbackRootOptional,
		proof.SubmittedBy,
		fmt.Sprintf("%020d", proof.Height),
	)
}

func ComputeNativePaymentReceiptHash(receipt PaymentReceipt) string {
	receipt = receipt.Normalize()
	return HashParts(
		"aetra-native-payment-receipt-v1",
		receipt.PaymentID,
		receipt.RouteID,
		receipt.ChannelID,
		string(receipt.Status),
		receipt.Amount,
		receipt.FeeAmount,
		receipt.ValueReturned,
		fmt.Sprintf("%020d", receipt.Height),
	)
}

func ComputePaymentChannelSetRoot(channels []PaymentChannel) string {
	parts := []string{"aetra-native-payment-channel-root-v1"}
	for _, channel := range normalizePaymentChannels(channels) {
		parts = append(parts, channel.ChannelRoot)
	}
	return HashParts(parts...)
}

func ComputeVirtualPaymentChannelSetRoot(channels []VirtualPaymentChannel) string {
	parts := []string{"aetra-native-virtual-payment-channel-root-v1"}
	for _, channel := range normalizeVirtualPaymentChannels(channels) {
		parts = append(parts, channel.VirtualRoot)
	}
	return HashParts(parts...)
}

func ComputeNativePaymentRouteSetRoot(routes []PaymentRoute) string {
	parts := []string{"aetra-native-payment-route-root-v1"}
	for _, route := range normalizeNativePaymentRoutes(routes) {
		parts = append(parts, route.RouteRoot)
	}
	return HashParts(parts...)
}

func ComputeLiquidityReservationSetRoot(reservations []LiquidityReservation) string {
	parts := []string{"aetra-native-liquidity-reservation-root-v1"}
	for _, reservation := range normalizeLiquidityReservations(reservations) {
		parts = append(parts, reservation.ReservationRoot)
	}
	return HashParts(parts...)
}

func ComputeSettlementProofSetRoot(proofs []SettlementProof) string {
	parts := []string{"aetra-native-settlement-proof-root-v1"}
	for _, proof := range normalizeSettlementProofs(proofs) {
		parts = append(parts, proof.ProofRoot)
	}
	return HashParts(parts...)
}

func ComputeNativePaymentReceiptSetRoot(receipts []PaymentReceipt) string {
	parts := []string{"aetra-native-payment-receipt-root-v1"}
	for _, receipt := range normalizePaymentReceipts(receipts) {
		parts = append(parts, receipt.ReceiptHash)
	}
	return HashParts(parts...)
}

func ComputeNativePaymentRoutingSnapshotRoot(snapshot NativePaymentRoutingSnapshot) string {
	snapshot = snapshot.Normalize()
	return HashParts(
		"aetra-native-payment-routing-snapshot-v1",
		fmt.Sprintf("%020d", snapshot.Height),
		snapshot.ChannelRoot,
		snapshot.VirtualChannelRoot,
		snapshot.ConditionRoot,
		snapshot.RouteRoot,
		snapshot.ReservationRoot,
		snapshot.SettlementProofRoot,
		snapshot.ReceiptRoot,
	)
}

func IsPaymentReceiptStatus(status PaymentReceiptStatus) bool {
	switch status {
	case PaymentReceiptExecuted, PaymentReceiptFailed, PaymentReceiptRefunded, PaymentReceiptExpired, PaymentReceiptSettled:
		return true
	default:
		return false
	}
}

func IsSettlementProofType(proofType SettlementProofType) bool {
	switch proofType {
	case SettlementProofLatestState, SettlementProofFraud, SettlementProofCloseAuthorization, SettlementProofFallbackSettlement:
		return true
	default:
		return false
	}
}

func validatePaymentRoutingToken(field, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > MaxTokenLength {
		return fmt.Errorf("%s must be <= %d bytes", field, MaxTokenLength)
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' {
			continue
		}
		return fmt.Errorf("%s contains unsupported character %q", field, r)
	}
	return nil
}

func buildPaymentChannels(channels []PaymentChannel) ([]PaymentChannel, error) {
	out := make([]PaymentChannel, len(channels))
	for i, channel := range channels {
		channel = channel.Normalize()
		if channel.ChannelRoot == "" {
			built, err := BuildPaymentChannel(channel)
			if err != nil {
				return nil, err
			}
			out[i] = built
			continue
		}
		if err := channel.Validate(); err != nil {
			return nil, err
		}
		out[i] = channel
	}
	return normalizePaymentChannels(out), nil
}

func buildVirtualPaymentChannels(channels []VirtualPaymentChannel) ([]VirtualPaymentChannel, error) {
	out := make([]VirtualPaymentChannel, len(channels))
	for i, channel := range channels {
		channel = channel.Normalize()
		if channel.VirtualRoot == "" {
			built, err := BuildVirtualPaymentChannel(channel)
			if err != nil {
				return nil, err
			}
			out[i] = built
			continue
		}
		if err := channel.Validate(); err != nil {
			return nil, err
		}
		out[i] = channel
	}
	return normalizeVirtualPaymentChannels(out), nil
}

func buildNativePaymentRoutes(routes []PaymentRoute) ([]PaymentRoute, error) {
	out := make([]PaymentRoute, len(routes))
	for i, route := range routes {
		route = route.Normalize()
		if route.RouteRoot == "" {
			built, err := BuildPaymentRoute(route)
			if err != nil {
				return nil, err
			}
			out[i] = built
			continue
		}
		if err := route.Validate(); err != nil {
			return nil, err
		}
		out[i] = route
	}
	return normalizeNativePaymentRoutes(out), nil
}

func buildLiquidityReservations(reservations []LiquidityReservation) ([]LiquidityReservation, error) {
	out := make([]LiquidityReservation, len(reservations))
	for i, reservation := range reservations {
		reservation = reservation.Normalize()
		if reservation.ReservationRoot == "" {
			built, err := BuildLiquidityReservation(reservation)
			if err != nil {
				return nil, err
			}
			out[i] = built
			continue
		}
		if err := reservation.Validate(); err != nil {
			return nil, err
		}
		out[i] = reservation
	}
	return normalizeLiquidityReservations(out), nil
}

func buildSettlementProofs(proofs []SettlementProof) ([]SettlementProof, error) {
	out := make([]SettlementProof, len(proofs))
	for i, proof := range proofs {
		proof = proof.Normalize()
		if proof.ProofRoot == "" {
			built, err := BuildSettlementProof(proof)
			if err != nil {
				return nil, err
			}
			out[i] = built
			continue
		}
		if err := proof.Validate(); err != nil {
			return nil, err
		}
		out[i] = proof
	}
	return normalizeSettlementProofs(out), nil
}

func buildPaymentReceipts(receipts []PaymentReceipt) ([]PaymentReceipt, error) {
	out := make([]PaymentReceipt, len(receipts))
	for i, receipt := range receipts {
		receipt = receipt.Normalize()
		if receipt.ReceiptHash == "" {
			built, err := BuildPaymentReceipt(receipt)
			if err != nil {
				return nil, err
			}
			out[i] = built
			continue
		}
		if err := receipt.Validate(); err != nil {
			return nil, err
		}
		out[i] = receipt
	}
	return normalizePaymentReceipts(out), nil
}

func validatePaymentChannels(channels []PaymentChannel) error {
	seen := map[string]struct{}{}
	for _, channel := range channels {
		if err := channel.Validate(); err != nil {
			return err
		}
		if _, found := seen[channel.Normalize().ChannelID]; found {
			return errors.New("payments duplicate native channel")
		}
		seen[channel.Normalize().ChannelID] = struct{}{}
	}
	return nil
}

func validateVirtualPaymentChannels(channels []VirtualPaymentChannel) error {
	seen := map[string]struct{}{}
	for _, channel := range channels {
		if err := channel.Validate(); err != nil {
			return err
		}
		if _, found := seen[channel.Normalize().VirtualChannelID]; found {
			return errors.New("payments duplicate virtual payment channel")
		}
		seen[channel.Normalize().VirtualChannelID] = struct{}{}
	}
	return nil
}

func validateNativePaymentRoutes(routes []PaymentRoute) error {
	seen := map[string]struct{}{}
	for _, route := range routes {
		if err := route.Validate(); err != nil {
			return err
		}
		if _, found := seen[route.Normalize().RouteID]; found {
			return errors.New("payments duplicate native route")
		}
		seen[route.Normalize().RouteID] = struct{}{}
	}
	return nil
}

func validateLiquidityReservations(reservations []LiquidityReservation) error {
	seen := map[string]struct{}{}
	for _, reservation := range reservations {
		if err := reservation.Validate(); err != nil {
			return err
		}
		if _, found := seen[reservation.Normalize().ReservationID]; found {
			return errors.New("payments duplicate liquidity reservation")
		}
		seen[reservation.Normalize().ReservationID] = struct{}{}
	}
	return nil
}

func validateSettlementProofs(proofs []SettlementProof) error {
	seen := map[string]struct{}{}
	for _, proof := range proofs {
		if err := proof.Validate(); err != nil {
			return err
		}
		if _, found := seen[proof.Normalize().ProofID]; found {
			return errors.New("payments duplicate settlement proof")
		}
		seen[proof.Normalize().ProofID] = struct{}{}
	}
	return nil
}

func validatePaymentReceipts(receipts []PaymentReceipt) error {
	seen := map[string]struct{}{}
	for _, receipt := range receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seen[receipt.Normalize().PaymentID]; found {
			return errors.New("payments duplicate native receipt")
		}
		seen[receipt.Normalize().PaymentID] = struct{}{}
	}
	return nil
}

func normalizePaymentChannels(channels []PaymentChannel) []PaymentChannel {
	out := make([]PaymentChannel, len(channels))
	for i, channel := range channels {
		out[i] = channel.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeVirtualPaymentChannels(channels []VirtualPaymentChannel) []VirtualPaymentChannel {
	out := make([]VirtualPaymentChannel, len(channels))
	for i, channel := range channels {
		out[i] = channel.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].VirtualChannelID < out[j].VirtualChannelID })
	return out
}

func normalizeNativePaymentRoutes(routes []PaymentRoute) []PaymentRoute {
	out := make([]PaymentRoute, len(routes))
	for i, route := range routes {
		out[i] = route.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].RouteID < out[j].RouteID })
	return out
}

func normalizeLiquidityReservations(reservations []LiquidityReservation) []LiquidityReservation {
	out := make([]LiquidityReservation, len(reservations))
	for i, reservation := range reservations {
		out[i] = reservation.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReservationID < out[j].ReservationID })
	return out
}

func normalizeSettlementProofs(proofs []SettlementProof) []SettlementProof {
	out := make([]SettlementProof, len(proofs))
	for i, proof := range proofs {
		out[i] = proof.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ProofID < out[j].ProofID })
	return out
}

func normalizePaymentReceipts(receipts []PaymentReceipt) []PaymentReceipt {
	out := make([]PaymentReceipt, len(receipts))
	for i, receipt := range receipts {
		out[i] = receipt.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PaymentID < out[j].PaymentID })
	return out
}
