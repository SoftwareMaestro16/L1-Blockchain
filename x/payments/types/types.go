package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	NativeDenom              = "naet"
	CurrentStateVersion      = uint32(1)
	DefaultDisputePeriod     = uint64(16)
	DefaultOpeningFee        = "1"
	MinCloseDelay            = uint64(1)
	MaxCloseDelay            = uint64(10_000)
	MinChallengePeriod       = uint64(1)
	MaxChallengePeriod       = uint64(20_000)
	MaxParticipants          = 8
	MaxConditionsPerState    = 128
	MaxParentChannels        = 16
	MaxSettlementBatchOps    = 256
	MaxRoutingHops           = 16
	MaxTokenLength           = 128
	MaxSettlementFeeFraction = int64(10_000)
)

type ChannelType string

const (
	ChannelTypeBidirectional  ChannelType = "BIDIRECTIONAL"
	ChannelTypeUnidirectional ChannelType = "UNIDIRECTIONAL"
	ChannelTypeAsync          ChannelType = "ASYNC"
)

type ChannelStatus string

const (
	ChannelStatusOpen         ChannelStatus = "OPEN"
	ChannelStatusPendingClose ChannelStatus = "PENDING_CLOSE"
	ChannelStatusSettled      ChannelStatus = "SETTLED"
)

type ConditionType string

const (
	ConditionTypeHashLock ConditionType = "HASH_LOCK"
	ConditionTypeTimeLock ConditionType = "TIME_LOCK"
)

type FraudProofType string

const (
	FraudProofTypeDoubleSign FraudProofType = "DOUBLE_SIGN"
	FraudProofTypeStaleClose FraudProofType = "STALE_CLOSE"
)

type BatchOperationType string

const (
	BatchOperationOpen    BatchOperationType = "OPEN"
	BatchOperationClose   BatchOperationType = "CLOSE"
	BatchOperationDispute BatchOperationType = "DISPUTE"
	BatchOperationSettle  BatchOperationType = "SETTLE"
)

type VirtualChannelStatus string

const (
	VirtualChannelStatusOpen    VirtualChannelStatus = "OPEN"
	VirtualChannelStatusSettled VirtualChannelStatus = "SETTLED"
)

type Balance struct {
	Participant string
	Amount      string
}

type ConditionalPayment struct {
	ConditionID   string
	ConditionType ConditionType
	Payer         string
	Payee         string
	Amount        string
	HashLock      string
	TimeoutHeight uint64
	NonceStart    uint64
	NonceEnd      uint64
}

type StateSignature struct {
	Signer        string
	StateHash     string
	SignatureHash string
}

type ClaimSignature struct {
	Signer        string
	ClaimHash     string
	SignatureHash string
}

type DeltaSignature struct {
	Signer        string
	DeltaHash     string
	SignatureHash string
}

type ChannelOpenRequest struct {
	ChainID                      string
	ChannelID                    string
	Participants                 []string
	InitialBalances              []Balance
	ChannelType                  ChannelType
	Collateral                   string
	CloseDelay                   uint64
	ChallengePeriod              uint64
	FeePolicyID                  string
	OpeningFeeDenom              string
	OpeningFeePaid               string
	RoutingAdvertised            bool
	ConditionalPaymentsSupported bool
	OpenHeight                   uint64
	ExpirationHeight             uint64
	ExpirationTimestamp          int64
}

type ChannelState struct {
	ChainID               string
	ChannelID             string
	ChannelType           ChannelType
	Denom                 string
	Version               uint32
	ParticipantA          string
	ParticipantB          string
	BalanceA              string
	BalanceB              string
	ReserveA              string
	ReserveB              string
	Epoch                 uint64
	Nonce                 uint64
	PendingConditionsRoot string
	Balances              []Balance
	Conditions            []ConditionalPayment
	PreviousStateHash     string
	StateHash             string
	TimeoutHeight         uint64
	TimeoutTimestamp      int64
	CloseDelay            uint64
	FeePolicyID           string
	CheckpointNonce       uint64
	CheckpointBalances    []Balance
	AsyncUpdateRoot       string
	AcceptedUpdateRoot    string
	SendWindow            uint64
	ReceiveWindow         uint64
	MaxUnackedAmount      string
	ExpiryHeight          uint64
	Signatures            []StateSignature
}

type AsyncPaymentDelta struct {
	UpdateID     string
	ChannelID    string
	From         string
	To           string
	Direction    string
	Amount       string
	NonceStart   uint64
	NonceEnd     uint64
	ExpiryHeight uint64
	DeltaHash    string
	Signature    DeltaSignature
}

type AsyncDeltaDisputeProof struct {
	ProofID         string
	ChannelID       string
	CheckpointState ChannelState
	Deltas          []AsyncPaymentDelta
	EvidenceHash    string
}

type UnidirectionalClaim struct {
	ChannelID           string
	Payer               string
	Receiver            string
	LockedAmount        string
	ClaimedAmount       string
	Nonce               uint64
	ExpirationHeight    uint64
	ExpirationTimestamp int64
	StateHash           string
	PayerSignature      ClaimSignature
	ReceiverAckOptional ClaimSignature
}

type StreamingPaymentFrame struct {
	ChannelID           string
	StreamID            string
	Payer               string
	Receiver            string
	PreviousClaimed     string
	RatePerBlock        string
	StartHeight         uint64
	CurrentHeight       uint64
	Nonce               uint64
	ExpirationHeight    uint64
	ExpirationTimestamp int64
}

type PendingClose struct {
	Submitter          string
	SubmittedHeight    uint64
	SettleAfterHeight  uint64
	SettlementFeeDenom string
	SettlementFee      string
	State              ChannelState
	FraudProofs        []FraudProof
	Penalties          []Penalty
}

type FraudProof struct {
	ProofID         string
	ProofType       FraudProofType
	SubmittedBy     string
	OffendingSigner string
	StateA          ChannelState
	StateB          ChannelState
	PenaltyDenom    string
	PenaltyAmount   string
	EvidenceHash    string
}

type Penalty struct {
	Offender  string
	Recipient string
	Denom     string
	Amount    string
}

type ChannelRecord struct {
	ChainID             string
	ChannelID           string
	ChannelType         ChannelType
	Participants        []string
	RequiredSigners     []string
	Payer               string
	Receiver            string
	ReceiverAckRequired bool
	Denom               string
	Collateral          string
	OpenHeight          uint64
	CloseDelay          uint64
	DisputePeriod       uint64
	ExpirationHeight    uint64
	ExpirationTimestamp int64
	OpeningFeeDenom     string
	OpeningFeePaid      string
	RoutingAdvertised   bool
	ConditionalPayments bool
	CustodyDenom        string
	CustodyAmount       string
	Status              ChannelStatus
	OpeningStateHash    string
	FinalizedNonce      uint64
	LatestState         ChannelState
	LatestClaim         UnidirectionalClaim
	PendingClose        PendingClose
}

type SettlementRecord struct {
	ChannelID          string
	StateHash          string
	Nonce              uint64
	FinalBalances      []Balance
	SettlementFeeDenom string
	SettlementFee      string
	Penalties          []Penalty
	SettledHeight      uint64
	SettlementHash     string
}

type CustodyLock struct {
	ChannelID string
	Denom     string
	Amount    string
}

type PaymentEventAttribute struct {
	Key   string
	Value string
}

type PaymentEvent struct {
	EventID    string
	EventType  string
	ChannelID  string
	Height     uint64
	Attributes []PaymentEventAttribute
}

type ChannelEdge struct {
	ChannelID     string
	From          string
	To            string
	Capacity      string
	FeeDenom      string
	FeeAmount     string
	ExpiresHeight uint64
	Active        bool
}

type VirtualChannel struct {
	VirtualChannelID string
	ParentChannelIDs []string
	Endpoints        []string
	Capacity         string
	ExpiresHeight    uint64
	Status           VirtualChannelStatus
	AnchorCommitment string
}

type SettlementOperation struct {
	OperationID   string
	OperationType BatchOperationType
	ChannelID     string
	Nonce         uint64
	StateHash     string
}

type SettlementBatch struct {
	BatchID    string
	Operations []SettlementOperation
	RootHash   string
}

func BuildState(state ChannelState) (ChannelState, error) {
	state = state.Normalize()
	if err := validateUnsignedStateShape(state); err != nil {
		return ChannelState{}, err
	}
	state.StateHash = ComputeStateHash(state)
	return state, nil
}

func SignatureForState(state ChannelState, signer string) (StateSignature, error) {
	if state.StateHash == "" {
		var err error
		state, err = BuildState(state)
		if err != nil {
			return StateSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments state signer", signer); err != nil {
		return StateSignature{}, err
	}
	return StateSignature{
		Signer:        signer,
		StateHash:     state.StateHash,
		SignatureHash: ComputeSignatureHash(signer, state.StateHash),
	}, nil
}

func BuildChannelFromOpenRequest(req ChannelOpenRequest) (ChannelRecord, error) {
	req = req.Normalize()
	if err := req.Validate(); err != nil {
		return ChannelRecord{}, err
	}
	channel := ChannelRecord{
		ChainID:             req.ChainID,
		ChannelID:           req.ChannelID,
		ChannelType:         req.ChannelType,
		Participants:        req.Participants,
		Denom:               NativeDenom,
		Collateral:          req.Collateral,
		OpenHeight:          req.OpenHeight,
		CloseDelay:          req.CloseDelay,
		DisputePeriod:       req.ChallengePeriod,
		ExpirationHeight:    req.ExpirationHeight,
		ExpirationTimestamp: req.ExpirationTimestamp,
		OpeningFeeDenom:     req.OpeningFeeDenom,
		OpeningFeePaid:      req.OpeningFeePaid,
		RoutingAdvertised:   req.RoutingAdvertised,
		ConditionalPayments: req.ConditionalPaymentsSupported,
		CustodyDenom:        NativeDenom,
		CustodyAmount:       req.Collateral,
		Status:              ChannelStatusOpen,
	}
	if req.ChannelType == ChannelTypeUnidirectional && len(req.Participants) == 2 {
		channel.Payer = req.Participants[0]
		channel.Receiver = req.Participants[1]
	}
	state, err := BuildState(openingStateForRequest(req, channel))
	if err != nil {
		return ChannelRecord{}, err
	}
	for _, signer := range channel.Normalize().Participants {
		sig, err := SignatureForState(state, signer)
		if err != nil {
			return ChannelRecord{}, err
		}
		state.Signatures = append(state.Signatures, sig)
	}
	channel.LatestState = state.Normalize()
	channel.OpeningStateHash = channel.LatestState.StateHash
	if err := channel.Validate(); err != nil {
		return ChannelRecord{}, err
	}
	return channel.Normalize(), nil
}

func (r ChannelOpenRequest) Normalize() ChannelOpenRequest {
	r.ChainID = strings.TrimSpace(r.ChainID)
	r.ChannelID = normalizeOptionalHash(r.ChannelID)
	r.Participants = normalizeAddressSet(r.Participants)
	r.InitialBalances = normalizeBalances(r.InitialBalances)
	r.Collateral = strings.TrimSpace(r.Collateral)
	r.FeePolicyID = strings.TrimSpace(r.FeePolicyID)
	if r.FeePolicyID == "" {
		r.FeePolicyID = NativeDenom
	}
	r.OpeningFeeDenom = normalizeAssetDenom(r.OpeningFeeDenom)
	r.OpeningFeePaid = strings.TrimSpace(r.OpeningFeePaid)
	if r.ChannelID == "" {
		parts := append([]string{"open", r.ChainID, string(r.ChannelType), r.Collateral}, r.Participants...)
		r.ChannelID = HashParts(parts...)
	}
	return r
}

func (r ChannelOpenRequest) Validate() error {
	req := r.Normalize()
	if strings.TrimSpace(req.ChainID) == "" {
		return errors.New("payments open chain id is required")
	}
	if err := ValidateHash("payments open channel id", req.ChannelID); err != nil {
		return err
	}
	if !IsChannelType(req.ChannelType) {
		return fmt.Errorf("unknown payments open channel type %q", req.ChannelType)
	}
	if err := validateAddressSet("payments open participant", req.Participants, 2, MaxParticipants); err != nil {
		return err
	}
	if err := validateBalances(req.InitialBalances); err != nil {
		return err
	}
	if err := validateInitialBalances(req.InitialBalances, req.Participants, req.Collateral); err != nil {
		return err
	}
	if err := validatePositiveInt("payments open collateral", req.Collateral); err != nil {
		return err
	}
	if err := validateCloseDelay(req.CloseDelay); err != nil {
		return err
	}
	if err := validateChallengePeriod(req.ChallengePeriod); err != nil {
		return err
	}
	if req.FeePolicyID != NativeDenom {
		return fmt.Errorf("payments open fee policy must be %s", NativeDenom)
	}
	if req.OpeningFeeDenom != NativeDenom {
		return fmt.Errorf("payments opening fee denom must be %s", NativeDenom)
	}
	if err := validateOpeningFeePaid(req.OpeningFeePaid); err != nil {
		return err
	}
	if req.OpenHeight == 0 {
		return errors.New("payments open height must be positive")
	}
	if req.ExpirationTimestamp < 0 {
		return errors.New("payments open expiration timestamp must be non-negative")
	}
	if req.ChannelType == ChannelTypeUnidirectional && req.ExpirationHeight == 0 {
		return errors.New("payments unidirectional open expiration height must be positive")
	}
	if req.ChannelType == ChannelTypeAsync && req.ExpirationHeight == 0 {
		return errors.New("payments async open expiry height must be positive")
	}
	return nil
}

func BuildUnidirectionalClaim(claim UnidirectionalClaim) (UnidirectionalClaim, error) {
	claim = claim.Normalize()
	if err := validateUnsignedUnidirectionalClaim(claim); err != nil {
		return UnidirectionalClaim{}, err
	}
	claim.StateHash = ComputeUnidirectionalClaimHash(claim)
	return claim, nil
}

func SignatureForClaim(claim UnidirectionalClaim, signer string) (ClaimSignature, error) {
	if claim.StateHash == "" {
		var err error
		claim, err = BuildUnidirectionalClaim(claim)
		if err != nil {
			return ClaimSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments claim signer", signer); err != nil {
		return ClaimSignature{}, err
	}
	return ClaimSignature{
		Signer:        signer,
		ClaimHash:     claim.StateHash,
		SignatureHash: ComputeClaimSignatureHash(signer, claim.StateHash),
	}, nil
}

func BuildAsyncDelta(delta AsyncPaymentDelta) (AsyncPaymentDelta, error) {
	delta = delta.Normalize()
	if err := validateUnsignedAsyncDelta(delta); err != nil {
		return AsyncPaymentDelta{}, err
	}
	delta.DeltaHash = ComputeAsyncDeltaHash(delta)
	return delta, nil
}

func SignatureForAsyncDelta(delta AsyncPaymentDelta, signer string) (DeltaSignature, error) {
	if delta.DeltaHash == "" {
		var err error
		delta, err = BuildAsyncDelta(delta)
		if err != nil {
			return DeltaSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments async delta signer", signer); err != nil {
		return DeltaSignature{}, err
	}
	return DeltaSignature{
		Signer:        signer,
		DeltaHash:     delta.DeltaHash,
		SignatureHash: ComputeDeltaSignatureHash(signer, delta.DeltaHash),
	}, nil
}

func AsyncDeltaDirection(from, to string) string {
	return strings.TrimSpace(from) + "->" + strings.TrimSpace(to)
}

func BuildAsyncCheckpointState(channel ChannelRecord, deltas []AsyncPaymentDelta, checkpointNonce, currentHeight uint64) (ChannelState, error) {
	channel = channel.Normalize()
	if channel.ChannelType != ChannelTypeAsync {
		return ChannelState{}, errors.New("payments async checkpoint requires async channel")
	}
	if checkpointNonce == 0 {
		return ChannelState{}, errors.New("payments async checkpoint nonce must be positive")
	}
	base := channel.LatestState.Normalize()
	if base.StateHash == "" {
		return ChannelState{}, errors.New("payments async checkpoint requires latest state")
	}
	if checkpointNonce <= base.CheckpointNonce {
		return ChannelState{}, errors.New("payments async checkpoint nonce must increase")
	}
	if currentHeight == 0 {
		return ChannelState{}, errors.New("payments async checkpoint height must be positive")
	}
	if currentHeight > base.ExpiryHeight {
		return ChannelState{}, errors.New("payments async checkpoint is expired")
	}
	normalizedDeltas := normalizeAsyncDeltas(deltas)
	if err := validateAsyncDeltasForCheckpoint(channel, base, normalizedDeltas, checkpointNonce, currentHeight); err != nil {
		return ChannelState{}, err
	}
	nextBalances, err := applyAsyncDeltas(base.Balances, normalizedDeltas)
	if err != nil {
		return ChannelState{}, err
	}
	state, err := BuildState(ChannelState{
		ChainID:            channel.ChainID,
		ChannelID:          channel.ChannelID,
		ChannelType:        ChannelTypeAsync,
		Denom:              channel.Denom,
		Version:            CurrentStateVersion,
		Epoch:              base.Epoch,
		Nonce:              checkpointNonce,
		Balances:           nextBalances,
		CheckpointNonce:    checkpointNonce,
		CheckpointBalances: nextBalances,
		AsyncUpdateRoot:    ComputeAsyncDeltaRoot(normalizedDeltas),
		AcceptedUpdateRoot: ComputeAsyncDeltaRoot(normalizedDeltas),
		SendWindow:         base.SendWindow,
		ReceiveWindow:      base.ReceiveWindow,
		MaxUnackedAmount:   base.MaxUnackedAmount,
		ExpiryHeight:       base.ExpiryHeight,
		TimeoutHeight:      base.TimeoutHeight,
		TimeoutTimestamp:   base.TimeoutTimestamp,
		CloseDelay:         base.CloseDelay,
		FeePolicyID:        NativeDenom,
	})
	if err != nil {
		return ChannelState{}, err
	}
	return state, nil
}

func StreamingClaimForChannel(channel ChannelRecord, frame StreamingPaymentFrame) (UnidirectionalClaim, error) {
	channel = channel.Normalize()
	frame = frame.Normalize()
	if channel.ChannelType != ChannelTypeUnidirectional {
		return UnidirectionalClaim{}, errors.New("payments streaming claim requires unidirectional channel")
	}
	if frame.ChannelID != channel.ChannelID {
		return UnidirectionalClaim{}, errors.New("payments streaming claim channel mismatch")
	}
	if frame.Payer != channel.Payer || frame.Receiver != channel.Receiver {
		return UnidirectionalClaim{}, errors.New("payments streaming claim parties mismatch")
	}
	if frame.CurrentHeight < frame.StartHeight {
		return UnidirectionalClaim{}, errors.New("payments streaming current height must be >= start height")
	}
	previous, err := parseNonNegativeInt("payments streaming previous claimed", frame.PreviousClaimed)
	if err != nil {
		return UnidirectionalClaim{}, err
	}
	rate, err := parseNonNegativeInt("payments streaming rate", frame.RatePerBlock)
	if err != nil {
		return UnidirectionalClaim{}, err
	}
	if frame.CurrentHeight-frame.StartHeight > uint64(^uint(0)>>1) {
		return UnidirectionalClaim{}, errors.New("payments streaming elapsed height is too large")
	}
	elapsed := sdkmath.NewInt(int64(frame.CurrentHeight - frame.StartHeight))
	claimed := previous.Add(rate.Mul(elapsed))
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return UnidirectionalClaim{}, err
	}
	if claimed.GT(collateral) {
		claimed = collateral
	}
	return BuildUnidirectionalClaim(UnidirectionalClaim{
		ChannelID:           channel.ChannelID,
		Payer:               channel.Payer,
		Receiver:            channel.Receiver,
		LockedAmount:        channel.Collateral,
		ClaimedAmount:       claimed.String(),
		Nonce:               frame.Nonce,
		ExpirationHeight:    frame.ExpirationHeight,
		ExpirationTimestamp: frame.ExpirationTimestamp,
	})
}

func (s ChannelState) Normalize() ChannelState {
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.Denom = strings.TrimSpace(s.Denom)
	if s.Version == 0 {
		s.Version = CurrentStateVersion
	}
	s.ParticipantA = strings.TrimSpace(s.ParticipantA)
	s.ParticipantB = strings.TrimSpace(s.ParticipantB)
	s.BalanceA = strings.TrimSpace(s.BalanceA)
	s.BalanceB = strings.TrimSpace(s.BalanceB)
	s.ReserveA = strings.TrimSpace(s.ReserveA)
	s.ReserveB = strings.TrimSpace(s.ReserveB)
	s.PreviousStateHash = normalizeOptionalHash(s.PreviousStateHash)
	s.StateHash = normalizeOptionalHash(s.StateHash)
	s.Balances = normalizeBalances(s.Balances)
	s.Conditions = normalizeConditions(s.Conditions)
	s.PendingConditionsRoot = normalizeOptionalHash(s.PendingConditionsRoot)
	if s.PendingConditionsRoot == "" {
		s.PendingConditionsRoot = ComputeConditionsRoot(s.Conditions)
	}
	if s.FeePolicyID == "" {
		s.FeePolicyID = NativeDenom
	}
	s.CheckpointBalances = normalizeBalances(s.CheckpointBalances)
	s.AsyncUpdateRoot = normalizeOptionalHash(s.AsyncUpdateRoot)
	s.AcceptedUpdateRoot = normalizeOptionalHash(s.AcceptedUpdateRoot)
	s.MaxUnackedAmount = strings.TrimSpace(s.MaxUnackedAmount)
	if s.ChannelType == ChannelTypeAsync {
		if s.CheckpointNonce == 0 {
			s.CheckpointNonce = s.Nonce
		}
		if len(s.CheckpointBalances) == 0 && len(s.Balances) > 0 {
			s.CheckpointBalances = normalizeBalances(s.Balances)
		}
		if len(s.Balances) == 0 && len(s.CheckpointBalances) > 0 {
			s.Balances = normalizeBalances(s.CheckpointBalances)
		}
		if s.AsyncUpdateRoot == "" {
			s.AsyncUpdateRoot = ComputeAsyncDeltaRoot(nil)
		}
		if s.AcceptedUpdateRoot == "" {
			s.AcceptedUpdateRoot = ComputeAsyncDeltaRoot(nil)
		}
	}
	if s.ChannelType == ChannelTypeBidirectional && len(s.Balances) == 2 {
		if s.ParticipantA == "" {
			s.ParticipantA = s.Balances[0].Participant
		}
		if s.ParticipantB == "" {
			s.ParticipantB = s.Balances[1].Participant
		}
		if s.BalanceA == "" {
			s.BalanceA = s.Balances[0].Amount
		}
		if s.BalanceB == "" {
			s.BalanceB = s.Balances[1].Amount
		}
	}
	if s.ReserveA == "" {
		s.ReserveA = "0"
	}
	if s.ReserveB == "" {
		s.ReserveB = "0"
	}
	s.Signatures = normalizeSignatures(s.Signatures)
	return s
}

func (s ClaimSignature) Normalize() ClaimSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ClaimHash = normalizeHash(s.ClaimHash)
	s.SignatureHash = normalizeHash(s.SignatureHash)
	return s
}

func (s ClaimSignature) Validate(expectedClaimHash string) error {
	s = s.Normalize()
	if err := addressing.ValidateUserAddress("payments claim signature signer", s.Signer); err != nil {
		return err
	}
	if s.ClaimHash != expectedClaimHash {
		return errors.New("payments claim signature hash mismatch")
	}
	if err := ValidateHash("payments claim signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeClaimSignatureHash(s.Signer, s.ClaimHash); s.SignatureHash != expected {
		return errors.New("payments claim signature value mismatch")
	}
	return nil
}

func (s DeltaSignature) Normalize() DeltaSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.DeltaHash = normalizeHash(s.DeltaHash)
	s.SignatureHash = normalizeHash(s.SignatureHash)
	return s
}

func (s DeltaSignature) Validate(expectedDeltaHash string) error {
	s = s.Normalize()
	if err := addressing.ValidateUserAddress("payments async delta signature signer", s.Signer); err != nil {
		return err
	}
	if s.DeltaHash != expectedDeltaHash {
		return errors.New("payments async delta signature hash mismatch")
	}
	if err := ValidateHash("payments async delta signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeDeltaSignatureHash(s.Signer, s.DeltaHash); s.SignatureHash != expected {
		return errors.New("payments async delta signature value mismatch")
	}
	return nil
}

func (d AsyncPaymentDelta) Normalize() AsyncPaymentDelta {
	d.UpdateID = normalizeHash(d.UpdateID)
	d.ChannelID = normalizeHash(d.ChannelID)
	d.From = strings.TrimSpace(d.From)
	d.To = strings.TrimSpace(d.To)
	d.Direction = strings.TrimSpace(d.Direction)
	if d.Direction == "" && d.From != "" && d.To != "" {
		d.Direction = AsyncDeltaDirection(d.From, d.To)
	}
	d.Amount = strings.TrimSpace(d.Amount)
	d.DeltaHash = normalizeOptionalHash(d.DeltaHash)
	d.Signature = d.Signature.Normalize()
	return d
}

func (d AsyncPaymentDelta) ValidateForChannel(channel ChannelRecord, currentHeight uint64) error {
	channel = channel.Normalize()
	if err := channel.ValidateCore(); err != nil {
		return err
	}
	if channel.ChannelType != ChannelTypeAsync {
		return errors.New("payments async delta requires async channel")
	}
	delta := d.Normalize()
	if err := validateUnsignedAsyncDelta(delta); err != nil {
		return err
	}
	if delta.ChannelID != channel.ChannelID {
		return errors.New("payments async delta channel mismatch")
	}
	if !containsString(channel.Participants, delta.From) || !containsString(channel.Participants, delta.To) {
		return errors.New("payments async delta parties must be channel participants")
	}
	if delta.From == delta.To {
		return errors.New("payments async delta parties must differ")
	}
	if delta.Direction != AsyncDeltaDirection(delta.From, delta.To) {
		return errors.New("payments async delta direction mismatch")
	}
	if currentHeight > delta.ExpiryHeight {
		return errors.New("payments async delta is expired")
	}
	if delta.DeltaHash == "" {
		return errors.New("payments async delta hash is required")
	}
	if expected := ComputeAsyncDeltaHash(delta); delta.DeltaHash != expected {
		return errors.New("payments async delta hash mismatch")
	}
	if err := delta.Signature.Validate(delta.DeltaHash); err != nil {
		return err
	}
	if delta.Signature.Signer != delta.From {
		return errors.New("payments async delta signer must be sender")
	}
	return nil
}

func (p AsyncDeltaDisputeProof) Normalize() AsyncDeltaDisputeProof {
	p.ProofID = normalizeHash(p.ProofID)
	p.ChannelID = normalizeHash(p.ChannelID)
	p.CheckpointState = p.CheckpointState.Normalize()
	p.Deltas = normalizeAsyncDeltas(p.Deltas)
	p.EvidenceHash = normalizeHash(p.EvidenceHash)
	return p
}

func (p AsyncDeltaDisputeProof) ValidateForChannel(channel ChannelRecord, currentHeight uint64) error {
	proof := p.Normalize()
	if err := ValidateHash("payments async dispute proof id", proof.ProofID); err != nil {
		return err
	}
	if proof.ChannelID != channel.Normalize().ChannelID {
		return errors.New("payments async dispute proof channel mismatch")
	}
	if err := ValidateHash("payments async dispute evidence hash", proof.EvidenceHash); err != nil {
		return err
	}
	if err := proof.CheckpointState.ValidateForChannel(channel, false); err != nil {
		return err
	}
	reconstructed, err := BuildAsyncCheckpointState(channel, proof.Deltas, proof.CheckpointState.CheckpointNonce, currentHeight)
	if err != nil {
		return err
	}
	if reconstructed.StateHash != proof.CheckpointState.StateHash {
		return errors.New("payments async dispute proof does not reconstruct checkpoint")
	}
	if proof.EvidenceHash != HashParts("async-dispute", proof.CheckpointState.StateHash, ComputeAsyncDeltaRoot(proof.Deltas)) {
		return errors.New("payments async dispute evidence hash mismatch")
	}
	return nil
}

func (c UnidirectionalClaim) Normalize() UnidirectionalClaim {
	c.ChannelID = normalizeHash(c.ChannelID)
	c.Payer = strings.TrimSpace(c.Payer)
	c.Receiver = strings.TrimSpace(c.Receiver)
	c.LockedAmount = strings.TrimSpace(c.LockedAmount)
	c.ClaimedAmount = strings.TrimSpace(c.ClaimedAmount)
	c.StateHash = normalizeOptionalHash(c.StateHash)
	c.PayerSignature = c.PayerSignature.Normalize()
	c.ReceiverAckOptional = c.ReceiverAckOptional.Normalize()
	return c
}

func (c UnidirectionalClaim) IsZero() bool {
	c = c.Normalize()
	return c.ChannelID == "" && c.StateHash == ""
}

func (c UnidirectionalClaim) ValidateForChannel(channel ChannelRecord) error {
	channel = channel.Normalize()
	if err := channel.ValidateCore(); err != nil {
		return err
	}
	if channel.ChannelType != ChannelTypeUnidirectional {
		return errors.New("payments claim requires unidirectional channel")
	}
	claim := c.Normalize()
	if err := validateUnsignedUnidirectionalClaim(claim); err != nil {
		return err
	}
	if claim.ChannelID != channel.ChannelID {
		return errors.New("payments claim channel mismatch")
	}
	if claim.Payer != channel.Payer || claim.Receiver != channel.Receiver {
		return errors.New("payments claim parties mismatch")
	}
	locked, err := parsePositiveInt("payments claim locked amount", claim.LockedAmount)
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	if !locked.Equal(collateral) {
		return errors.New("payments claim locked amount mismatch")
	}
	claimed, err := parseNonNegativeInt("payments claimed amount", claim.ClaimedAmount)
	if err != nil {
		return err
	}
	if claimed.GT(collateral) {
		return errors.New("payments claimed amount exceeds locked collateral")
	}
	if claim.StateHash == "" {
		return errors.New("payments claim state hash is required")
	}
	if expected := ComputeUnidirectionalClaimHash(claim); claim.StateHash != expected {
		return errors.New("payments claim state hash mismatch")
	}
	if err := claim.PayerSignature.Validate(claim.StateHash); err != nil {
		return err
	}
	if claim.PayerSignature.Signer != channel.Payer {
		return errors.New("payments claim payer signature is required")
	}
	if claim.ReceiverAckOptional.SignatureHash == "" {
		if channel.ReceiverAckRequired {
			return errors.New("payments receiver acknowledgement is required")
		}
		return nil
	}
	if err := claim.ReceiverAckOptional.Validate(claim.StateHash); err != nil {
		return err
	}
	if claim.ReceiverAckOptional.Signer != channel.Receiver {
		return errors.New("payments receiver acknowledgement signer mismatch")
	}
	return nil
}

func (f StreamingPaymentFrame) Normalize() StreamingPaymentFrame {
	f.ChannelID = normalizeHash(f.ChannelID)
	f.StreamID = normalizeHash(f.StreamID)
	f.Payer = strings.TrimSpace(f.Payer)
	f.Receiver = strings.TrimSpace(f.Receiver)
	f.PreviousClaimed = strings.TrimSpace(f.PreviousClaimed)
	f.RatePerBlock = strings.TrimSpace(f.RatePerBlock)
	return f
}

func (s ChannelState) ValidateForChannel(channel ChannelRecord, requireAllParticipants bool) error {
	channel = channel.Normalize()
	if err := channel.ValidateCore(); err != nil {
		return err
	}
	state := s.Normalize()
	if err := validateUnsignedStateShape(state); err != nil {
		return err
	}
	if state.ChainID != channel.ChainID {
		return errors.New("payments channel state chain id mismatch")
	}
	if state.ChannelID != channel.ChannelID {
		return errors.New("payments channel state id mismatch")
	}
	if state.ChannelType != channel.ChannelType {
		return errors.New("payments channel state type mismatch")
	}
	if state.Denom != channel.Denom {
		return errors.New("payments channel state denom mismatch")
	}
	if state.StateHash == "" {
		return errors.New("payments channel state hash is required")
	}
	if expected := ComputeStateHash(state); state.StateHash != expected {
		return errors.New("payments channel state hash mismatch")
	}
	if state.Nonce > 1 && channel.ChannelType != ChannelTypeAsync && state.PreviousStateHash == "" {
		return errors.New("payments channel state previous hash is required")
	}
	if err := validateStateParticipants(state, channel); err != nil {
		return err
	}
	if err := validateCollateralConservation(state, channel); err != nil {
		return err
	}
	required := channel.RequiredSigners
	if requireAllParticipants {
		required = channel.Participants
	}
	return validateSignatureQuorum(state.Signatures, required, state.StateHash)
}

func (s StateSignature) Normalize() StateSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.StateHash = normalizeHash(s.StateHash)
	s.SignatureHash = normalizeHash(s.SignatureHash)
	return s
}

func (s StateSignature) Validate(expectedStateHash string) error {
	s = s.Normalize()
	if err := addressing.ValidateUserAddress("payments signature signer", s.Signer); err != nil {
		return err
	}
	if s.StateHash != expectedStateHash {
		return errors.New("payments signature state hash mismatch")
	}
	if err := ValidateHash("payments signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeSignatureHash(s.Signer, s.StateHash); s.SignatureHash != expected {
		return errors.New("payments signature hash mismatch")
	}
	return nil
}

func (c ChannelRecord) Normalize() ChannelRecord {
	c.ChainID = strings.TrimSpace(c.ChainID)
	c.ChannelID = normalizeHash(c.ChannelID)
	c.Denom = strings.TrimSpace(c.Denom)
	c.Collateral = strings.TrimSpace(c.Collateral)
	c.OpeningStateHash = normalizeOptionalHash(c.OpeningStateHash)
	c.OpeningFeeDenom = normalizeAssetDenom(c.OpeningFeeDenom)
	c.OpeningFeePaid = strings.TrimSpace(c.OpeningFeePaid)
	c.CustodyDenom = normalizeAssetDenom(c.CustodyDenom)
	c.CustodyAmount = strings.TrimSpace(c.CustodyAmount)
	c.Participants = normalizeAddressSet(c.Participants)
	c.RequiredSigners = normalizeAddressSet(c.RequiredSigners)
	c.Payer = strings.TrimSpace(c.Payer)
	c.Receiver = strings.TrimSpace(c.Receiver)
	if c.ChannelType == ChannelTypeUnidirectional && len(c.Participants) == 2 {
		if c.Payer == "" {
			c.Payer = c.Participants[0]
		}
		if c.Receiver == "" {
			c.Receiver = c.Participants[1]
		}
	}
	if len(c.RequiredSigners) == 0 {
		c.RequiredSigners = append([]string(nil), c.Participants...)
	}
	if c.DisputePeriod == 0 {
		c.DisputePeriod = DefaultDisputePeriod
	}
	if c.CloseDelay == 0 && c.LatestState.CloseDelay != 0 {
		c.CloseDelay = c.LatestState.CloseDelay
	}
	if c.CustodyAmount == "" {
		c.CustodyAmount = c.Collateral
	}
	if c.Status == "" {
		c.Status = ChannelStatusOpen
	}
	c.LatestState = c.LatestState.Normalize()
	c.LatestClaim = c.LatestClaim.Normalize()
	c.PendingClose = c.PendingClose.Normalize()
	return c
}

func (c ChannelRecord) ValidateCore() error {
	if strings.TrimSpace(c.ChainID) == "" {
		return errors.New("payments chain id is required")
	}
	if len(c.ChainID) > MaxTokenLength {
		return fmt.Errorf("payments chain id must be <= %d bytes", MaxTokenLength)
	}
	if err := ValidateHash("payments channel id", c.ChannelID); err != nil {
		return err
	}
	if !IsChannelType(c.ChannelType) {
		return fmt.Errorf("unknown payments channel type %q", c.ChannelType)
	}
	if c.Denom != NativeDenom {
		return fmt.Errorf("payments channel collateral denom must be %s", NativeDenom)
	}
	if err := validatePositiveInt("payments channel collateral", c.Collateral); err != nil {
		return err
	}
	if c.OpenHeight == 0 {
		return errors.New("payments channel open height must be positive")
	}
	if err := validateCloseDelay(c.CloseDelay); err != nil {
		return err
	}
	if c.DisputePeriod == 0 {
		return errors.New("payments channel dispute period must be positive")
	}
	if err := validateChallengePeriod(c.DisputePeriod); err != nil {
		return err
	}
	if c.OpeningFeeDenom != NativeDenom {
		return fmt.Errorf("payments opening fee denom must be %s", NativeDenom)
	}
	if err := validateOpeningFeePaid(c.OpeningFeePaid); err != nil {
		return err
	}
	if c.CustodyDenom != NativeDenom {
		return fmt.Errorf("payments custody denom must be %s", NativeDenom)
	}
	if c.CustodyAmount != c.Collateral {
		return errors.New("payments custody amount must match channel collateral")
	}
	if !IsChannelStatus(c.Status) {
		return fmt.Errorf("unknown payments channel status %q", c.Status)
	}
	if err := validateAddressSet("payments channel participant", c.Participants, 2, MaxParticipants); err != nil {
		return err
	}
	if err := validateAddressSet("payments channel required signer", c.RequiredSigners, 1, MaxParticipants); err != nil {
		return err
	}
	if c.ChannelType == ChannelTypeUnidirectional {
		if err := validateUnidirectionalChannelCore(c); err != nil {
			return err
		}
	}
	for _, signer := range c.RequiredSigners {
		if !containsString(c.Participants, signer) {
			return errors.New("payments required signer must be a channel participant")
		}
	}
	if c.OpeningStateHash != "" {
		if err := ValidateHash("payments opening state hash", c.OpeningStateHash); err != nil {
			return err
		}
	}
	return nil
}

func (c ChannelRecord) Validate() error {
	channel := c.Normalize()
	if err := channel.ValidateCore(); err != nil {
		return err
	}
	if channel.LatestState.StateHash == "" {
		return errors.New("payments channel latest state is required")
	}
	if err := channel.LatestState.ValidateForChannel(channel, false); err != nil {
		return err
	}
	if channel.LatestState.CloseDelay != channel.CloseDelay {
		return errors.New("payments opening state close delay mismatch")
	}
	if channel.LatestState.FeePolicyID != NativeDenom {
		return fmt.Errorf("payments opening state fee policy must be %s", NativeDenom)
	}
	if channel.OpeningStateHash == "" {
		return errors.New("payments opening state hash is required")
	}
	if channel.ChannelType == ChannelTypeUnidirectional {
		if err := validateUnidirectionalOpeningState(channel); err != nil {
			return err
		}
		if !channel.LatestClaim.IsZero() {
			if err := channel.LatestClaim.ValidateForChannel(channel); err != nil {
				return err
			}
		}
	}
	if channel.FinalizedNonce > channel.LatestState.Nonce {
		if channel.ChannelType != ChannelTypeUnidirectional || channel.LatestClaim.IsZero() || channel.FinalizedNonce > channel.LatestClaim.Nonce {
			return errors.New("payments finalized nonce cannot exceed latest state nonce")
		}
	}
	switch channel.Status {
	case ChannelStatusOpen:
		if channel.PendingClose.State.StateHash != "" {
			return errors.New("payments open channel must not have pending close")
		}
	case ChannelStatusPendingClose:
		if err := channel.PendingClose.ValidateForChannel(channel); err != nil {
			return err
		}
	case ChannelStatusSettled:
		if channel.PendingClose.State.StateHash != "" {
			return errors.New("payments settled channel must not have pending close")
		}
	}
	return nil
}

func (p PendingClose) Normalize() PendingClose {
	p.Submitter = strings.TrimSpace(p.Submitter)
	p.SettlementFeeDenom = normalizeAssetDenom(p.SettlementFeeDenom)
	p.SettlementFee = strings.TrimSpace(p.SettlementFee)
	p.State = p.State.Normalize()
	p.FraudProofs = normalizeFraudProofs(p.FraudProofs)
	p.Penalties = normalizePenalties(p.Penalties)
	return p
}

func (p PendingClose) ValidateForChannel(channel ChannelRecord) error {
	p = p.Normalize()
	if err := addressing.ValidateUserAddress("payments pending close submitter", p.Submitter); err != nil {
		return err
	}
	if !containsString(channel.Participants, p.Submitter) {
		return errors.New("payments pending close submitter must be participant")
	}
	if p.SubmittedHeight == 0 {
		return errors.New("payments pending close submitted height must be positive")
	}
	if p.SettleAfterHeight <= p.SubmittedHeight {
		return errors.New("payments pending close settlement height must exceed submitted height")
	}
	if p.SettlementFeeDenom != NativeDenom {
		return fmt.Errorf("payments settlement fee denom must be %s", NativeDenom)
	}
	if err := validateNonNegativeInt("payments settlement fee", p.SettlementFee); err != nil {
		return err
	}
	if err := p.State.ValidateForChannel(channel, false); err != nil {
		return err
	}
	for _, proof := range p.FraudProofs {
		if err := proof.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	for _, penalty := range p.Penalties {
		if err := penalty.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	return nil
}

func (f FraudProof) Normalize() FraudProof {
	f.ProofID = normalizeHash(f.ProofID)
	f.SubmittedBy = strings.TrimSpace(f.SubmittedBy)
	f.OffendingSigner = strings.TrimSpace(f.OffendingSigner)
	f.PenaltyDenom = normalizeAssetDenom(f.PenaltyDenom)
	f.PenaltyAmount = strings.TrimSpace(f.PenaltyAmount)
	f.EvidenceHash = normalizeHash(f.EvidenceHash)
	f.StateA = f.StateA.Normalize()
	f.StateB = f.StateB.Normalize()
	return f
}

func (f FraudProof) ValidateForChannel(channel ChannelRecord) error {
	proof := f.Normalize()
	if err := ValidateHash("payments fraud proof id", proof.ProofID); err != nil {
		return err
	}
	if !IsFraudProofType(proof.ProofType) {
		return fmt.Errorf("unknown payments fraud proof type %q", proof.ProofType)
	}
	if err := addressing.ValidateUserAddress("payments fraud submitter", proof.SubmittedBy); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments fraud offender", proof.OffendingSigner); err != nil {
		return err
	}
	if !containsString(channel.Participants, proof.SubmittedBy) || !containsString(channel.Participants, proof.OffendingSigner) {
		return errors.New("payments fraud parties must be channel participants")
	}
	if proof.PenaltyDenom != NativeDenom {
		return fmt.Errorf("payments fraud penalty denom must be %s", NativeDenom)
	}
	if err := validatePositiveInt("payments fraud penalty", proof.PenaltyAmount); err != nil {
		return err
	}
	if err := ValidateHash("payments fraud evidence hash", proof.EvidenceHash); err != nil {
		return err
	}
	switch proof.ProofType {
	case FraudProofTypeDoubleSign:
		if err := proof.StateA.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if err := proof.StateB.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if proof.StateA.ChannelID != proof.StateB.ChannelID || proof.StateA.Epoch != proof.StateB.Epoch || proof.StateA.Nonce != proof.StateB.Nonce {
			return errors.New("payments double-sign proof states must share channel, epoch, and nonce")
		}
		if proof.StateA.StateHash == proof.StateB.StateHash {
			return errors.New("payments double-sign proof requires conflicting state hashes")
		}
		if !stateSignedBy(proof.StateA, proof.OffendingSigner) || !stateSignedBy(proof.StateB, proof.OffendingSigner) {
			return errors.New("payments double-sign proof requires offender signature on both states")
		}
	case FraudProofTypeStaleClose:
		if err := proof.StateA.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if err := proof.StateB.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if proof.StateB.Nonce <= proof.StateA.Nonce {
			return errors.New("payments stale-close proof requires newer state")
		}
	default:
		return fmt.Errorf("unknown payments fraud proof type %q", proof.ProofType)
	}
	return nil
}

func (p Penalty) Normalize() Penalty {
	p.Offender = strings.TrimSpace(p.Offender)
	p.Recipient = strings.TrimSpace(p.Recipient)
	p.Denom = normalizeAssetDenom(p.Denom)
	p.Amount = strings.TrimSpace(p.Amount)
	return p
}

func (p Penalty) ValidateForChannel(channel ChannelRecord) error {
	p = p.Normalize()
	if err := addressing.ValidateUserAddress("payments penalty offender", p.Offender); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments penalty recipient", p.Recipient); err != nil {
		return err
	}
	if !containsString(channel.Participants, p.Offender) || !containsString(channel.Participants, p.Recipient) {
		return errors.New("payments penalty parties must be channel participants")
	}
	if p.Offender == p.Recipient {
		return errors.New("payments penalty parties must differ")
	}
	if p.Denom != NativeDenom {
		return fmt.Errorf("payments penalty denom must be %s", NativeDenom)
	}
	return validatePositiveInt("payments penalty amount", p.Amount)
}

func (s SettlementRecord) Normalize() SettlementRecord {
	s.ChannelID = normalizeHash(s.ChannelID)
	s.StateHash = normalizeHash(s.StateHash)
	s.SettlementFeeDenom = normalizeAssetDenom(s.SettlementFeeDenom)
	s.SettlementFee = strings.TrimSpace(s.SettlementFee)
	s.SettlementHash = normalizeOptionalHash(s.SettlementHash)
	s.FinalBalances = normalizeBalances(s.FinalBalances)
	s.Penalties = normalizePenalties(s.Penalties)
	return s
}

func (s SettlementRecord) ValidateForChannel(channel ChannelRecord) error {
	settlement := s.Normalize()
	if settlement.ChannelID != channel.ChannelID {
		return errors.New("payments settlement channel mismatch")
	}
	if err := ValidateHash("payments settlement state hash", settlement.StateHash); err != nil {
		return err
	}
	if settlement.Nonce == 0 {
		return errors.New("payments settlement nonce must be positive")
	}
	if settlement.SettledHeight == 0 {
		return errors.New("payments settlement height must be positive")
	}
	if settlement.SettlementFeeDenom != NativeDenom {
		return fmt.Errorf("payments settlement fee denom must be %s", NativeDenom)
	}
	if err := validateNonNegativeInt("payments settlement fee", settlement.SettlementFee); err != nil {
		return err
	}
	for _, balance := range settlement.FinalBalances {
		if !containsString(channel.Participants, balance.Participant) {
			return errors.New("payments settlement balance participant must be in channel")
		}
	}
	for _, penalty := range settlement.Penalties {
		if err := penalty.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	finalTotal, err := sumBalances(settlement.FinalBalances)
	if err != nil {
		return err
	}
	fee, err := parseNonNegativeInt("payments settlement fee", settlement.SettlementFee)
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	if !finalTotal.Add(fee).Equal(collateral) {
		return errors.New("payments settlement must conserve collateral minus fee")
	}
	if settlement.SettlementHash == "" {
		return errors.New("payments settlement hash is required")
	}
	if expected := ComputeSettlementHash(settlement); settlement.SettlementHash != expected {
		return errors.New("payments settlement hash mismatch")
	}
	return nil
}

func (c CustodyLock) Normalize() CustodyLock {
	c.ChannelID = normalizeHash(c.ChannelID)
	c.Denom = normalizeAssetDenom(c.Denom)
	c.Amount = strings.TrimSpace(c.Amount)
	return c
}

func (c CustodyLock) ValidateForChannel(channel ChannelRecord) error {
	lock := c.Normalize()
	if lock.ChannelID != channel.Normalize().ChannelID {
		return errors.New("payments custody channel mismatch")
	}
	if lock.Denom != NativeDenom {
		return fmt.Errorf("payments custody denom must be %s", NativeDenom)
	}
	locked, err := parsePositiveInt("payments custody amount", lock.Amount)
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	if !locked.Equal(collateral) {
		return errors.New("payments custody amount must match channel collateral")
	}
	return nil
}

func (e PaymentEventAttribute) Normalize() PaymentEventAttribute {
	e.Key = strings.TrimSpace(e.Key)
	e.Value = strings.TrimSpace(e.Value)
	return e
}

func (e PaymentEvent) Normalize() PaymentEvent {
	e.EventID = normalizeHash(e.EventID)
	e.EventType = strings.TrimSpace(e.EventType)
	e.ChannelID = normalizeHash(e.ChannelID)
	e.Attributes = normalizePaymentEventAttributes(e.Attributes)
	return e
}

func (e PaymentEvent) Validate() error {
	event := e.Normalize()
	if err := ValidateHash("payments event id", event.EventID); err != nil {
		return err
	}
	if event.EventType == "" {
		return errors.New("payments event type is required")
	}
	if err := ValidateHash("payments event channel id", event.ChannelID); err != nil {
		return err
	}
	if event.Height == 0 {
		return errors.New("payments event height must be positive")
	}
	seen := make(map[string]struct{}, len(event.Attributes))
	for _, attr := range event.Attributes {
		if attr.Key == "" {
			return errors.New("payments event attribute key is required")
		}
		if _, found := seen[attr.Key]; found {
			return errors.New("payments duplicate event attribute")
		}
		seen[attr.Key] = struct{}{}
	}
	return nil
}

func ChannelOpenEvent(channel ChannelRecord) PaymentEvent {
	channel = channel.Normalize()
	event := PaymentEvent{
		EventID:   HashParts("channel-open", channel.ChannelID, channel.OpeningStateHash),
		EventType: "channel-open",
		ChannelID: channel.ChannelID,
		Height:    channel.OpenHeight,
		Attributes: []PaymentEventAttribute{
			{Key: "channel_type", Value: string(channel.ChannelType)},
			{Key: "collateral", Value: channel.Collateral},
			{Key: "denom", Value: channel.Denom},
			{Key: "opening_fee", Value: channel.OpeningFeePaid},
			{Key: "routing_advertised", Value: fmt.Sprintf("%t", channel.RoutingAdvertised)},
			{Key: "conditional_payments", Value: fmt.Sprintf("%t", channel.ConditionalPayments)},
		},
	}
	return event.Normalize()
}

func (e ChannelEdge) Normalize() ChannelEdge {
	e.ChannelID = normalizeHash(e.ChannelID)
	e.From = strings.TrimSpace(e.From)
	e.To = strings.TrimSpace(e.To)
	e.Capacity = strings.TrimSpace(e.Capacity)
	e.FeeDenom = normalizeAssetDenom(e.FeeDenom)
	e.FeeAmount = strings.TrimSpace(e.FeeAmount)
	return e
}

func (e ChannelEdge) Validate() error {
	e = e.Normalize()
	if err := ValidateHash("payments routing channel id", e.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments routing from", e.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments routing to", e.To); err != nil {
		return err
	}
	if e.From == e.To {
		return errors.New("payments routing edge endpoints must differ")
	}
	if err := validatePositiveInt("payments routing capacity", e.Capacity); err != nil {
		return err
	}
	if e.FeeDenom != NativeDenom {
		return fmt.Errorf("payments routing fee denom must be %s", NativeDenom)
	}
	return validateNonNegativeInt("payments routing fee", e.FeeAmount)
}

func (v VirtualChannel) Normalize() VirtualChannel {
	v.VirtualChannelID = normalizeHash(v.VirtualChannelID)
	for i := range v.ParentChannelIDs {
		v.ParentChannelIDs[i] = normalizeHash(v.ParentChannelIDs[i])
	}
	v.Endpoints = normalizeAddressSet(v.Endpoints)
	v.Capacity = strings.TrimSpace(v.Capacity)
	v.AnchorCommitment = normalizeOptionalHash(v.AnchorCommitment)
	if v.Status == "" {
		v.Status = VirtualChannelStatusOpen
	}
	return v
}

func (v VirtualChannel) Validate() error {
	vc := v.Normalize()
	if err := ValidateHash("payments virtual channel id", vc.VirtualChannelID); err != nil {
		return err
	}
	if len(vc.ParentChannelIDs) == 0 || len(vc.ParentChannelIDs) > MaxParentChannels {
		return fmt.Errorf("payments virtual parent channels must be between 1 and %d", MaxParentChannels)
	}
	seen := make(map[string]struct{}, len(vc.ParentChannelIDs))
	for _, id := range vc.ParentChannelIDs {
		if err := ValidateHash("payments virtual parent channel id", id); err != nil {
			return err
		}
		if _, found := seen[id]; found {
			return errors.New("payments virtual parent channels must be unique")
		}
		seen[id] = struct{}{}
	}
	if err := validateAddressSet("payments virtual endpoint", vc.Endpoints, 2, 2); err != nil {
		return err
	}
	if err := validatePositiveInt("payments virtual capacity", vc.Capacity); err != nil {
		return err
	}
	if vc.ExpiresHeight == 0 {
		return errors.New("payments virtual channel expiry height must be positive")
	}
	if !IsVirtualChannelStatus(vc.Status) {
		return fmt.Errorf("unknown payments virtual channel status %q", vc.Status)
	}
	if vc.AnchorCommitment == "" {
		return errors.New("payments virtual channel anchor is required")
	}
	if expected := ComputeVirtualChannelAnchor(vc); vc.AnchorCommitment != expected {
		return errors.New("payments virtual channel anchor mismatch")
	}
	return nil
}

func (op SettlementOperation) Normalize() SettlementOperation {
	op.OperationID = normalizeHash(op.OperationID)
	op.ChannelID = normalizeHash(op.ChannelID)
	op.StateHash = normalizeHash(op.StateHash)
	return op
}

func (op SettlementOperation) Validate() error {
	op = op.Normalize()
	if err := ValidateHash("payments settlement operation id", op.OperationID); err != nil {
		return err
	}
	if !IsBatchOperationType(op.OperationType) {
		return fmt.Errorf("unknown payments batch operation type %q", op.OperationType)
	}
	if err := ValidateHash("payments settlement operation channel id", op.ChannelID); err != nil {
		return err
	}
	if op.Nonce == 0 {
		return errors.New("payments settlement operation nonce must be positive")
	}
	return ValidateHash("payments settlement operation state hash", op.StateHash)
}

func (b SettlementBatch) Normalize() SettlementBatch {
	b.BatchID = normalizeHash(b.BatchID)
	b.RootHash = normalizeOptionalHash(b.RootHash)
	b.Operations = SortSettlementOperations(b.Operations)
	return b
}

func (b SettlementBatch) Validate() error {
	batch := b.Normalize()
	if err := ValidateHash("payments settlement batch id", batch.BatchID); err != nil {
		return err
	}
	if len(batch.Operations) == 0 || len(batch.Operations) > MaxSettlementBatchOps {
		return fmt.Errorf("payments settlement batch operations must be between 1 and %d", MaxSettlementBatchOps)
	}
	seenOps := make(map[string]struct{}, len(batch.Operations))
	seenChannels := make(map[string]struct{}, len(batch.Operations))
	for i, op := range batch.Operations {
		if err := op.Validate(); err != nil {
			return err
		}
		if _, found := seenOps[op.OperationID]; found {
			return errors.New("payments duplicate settlement batch operation")
		}
		seenOps[op.OperationID] = struct{}{}
		if _, found := seenChannels[op.ChannelID]; found {
			return errors.New("payments settlement batch must contain independent channels")
		}
		seenChannels[op.ChannelID] = struct{}{}
		if i > 0 && compareSettlementOperations(batch.Operations[i-1], op) >= 0 {
			return errors.New("payments settlement batch operations must be sorted canonically")
		}
	}
	if batch.RootHash == "" {
		return errors.New("payments settlement batch root is required")
	}
	if expected := ComputeBatchRoot(batch.Operations); batch.RootHash != expected {
		return errors.New("payments settlement batch root mismatch")
	}
	return nil
}

func NewSettlementBatch(batchID string, operations []SettlementOperation) (SettlementBatch, error) {
	batch := SettlementBatch{
		BatchID:    normalizeHash(batchID),
		Operations: SortSettlementOperations(operations),
	}
	batch.RootHash = ComputeBatchRoot(batch.Operations)
	if err := batch.Validate(); err != nil {
		return SettlementBatch{}, err
	}
	return batch, nil
}

func SortSettlementOperations(operations []SettlementOperation) []SettlementOperation {
	out := make([]SettlementOperation, len(operations))
	for i, op := range operations {
		out[i] = op.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareSettlementOperations(out[i], out[j]) < 0
	})
	return out
}

func IsChannelType(value ChannelType) bool {
	switch value {
	case ChannelTypeBidirectional, ChannelTypeUnidirectional, ChannelTypeAsync:
		return true
	default:
		return false
	}
}

func IsChannelStatus(value ChannelStatus) bool {
	switch value {
	case ChannelStatusOpen, ChannelStatusPendingClose, ChannelStatusSettled:
		return true
	default:
		return false
	}
}

func IsFraudProofType(value FraudProofType) bool {
	switch value {
	case FraudProofTypeDoubleSign, FraudProofTypeStaleClose:
		return true
	default:
		return false
	}
}

func IsBatchOperationType(value BatchOperationType) bool {
	switch value {
	case BatchOperationOpen, BatchOperationClose, BatchOperationDispute, BatchOperationSettle:
		return true
	default:
		return false
	}
}

func IsVirtualChannelStatus(value VirtualChannelStatus) bool {
	switch value {
	case VirtualChannelStatusOpen, VirtualChannelStatusSettled:
		return true
	default:
		return false
	}
}

func validateUnsignedStateShape(state ChannelState) error {
	if strings.TrimSpace(state.ChainID) == "" {
		return errors.New("payments channel state chain id is required")
	}
	if len(state.ChainID) > MaxTokenLength {
		return fmt.Errorf("payments channel state chain id must be <= %d bytes", MaxTokenLength)
	}
	if err := ValidateHash("payments channel state channel id", state.ChannelID); err != nil {
		return err
	}
	if !IsChannelType(state.ChannelType) {
		return fmt.Errorf("unknown payments channel state type %q", state.ChannelType)
	}
	if state.Denom != NativeDenom {
		return fmt.Errorf("payments channel state denom must be %s", NativeDenom)
	}
	if state.Version != CurrentStateVersion {
		return fmt.Errorf("payments channel state version must be %d", CurrentStateVersion)
	}
	if state.Epoch == 0 {
		return errors.New("payments channel state epoch must be positive")
	}
	if state.Nonce == 0 {
		return errors.New("payments channel state nonce must be positive")
	}
	if state.PreviousStateHash != "" {
		if err := ValidateHash("payments channel state previous hash", state.PreviousStateHash); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments pending conditions root", state.PendingConditionsRoot); err != nil {
		return err
	}
	if expected := ComputeConditionsRoot(state.Conditions); state.PendingConditionsRoot != expected {
		return errors.New("payments pending conditions root mismatch")
	}
	if state.TimeoutTimestamp < 0 {
		return errors.New("payments channel state timeout timestamp must be non-negative")
	}
	if state.FeePolicyID != NativeDenom {
		return fmt.Errorf("payments channel state fee policy must be %s", NativeDenom)
	}
	if err := validateBalances(state.Balances); err != nil {
		return err
	}
	return validateConditions(state.Conditions)
}

func openingStateForRequest(req ChannelOpenRequest, channel ChannelRecord) ChannelState {
	state := ChannelState{
		ChainID:       req.ChainID,
		ChannelID:     req.ChannelID,
		ChannelType:   req.ChannelType,
		Denom:         NativeDenom,
		Version:       CurrentStateVersion,
		Epoch:         1,
		Nonce:         1,
		Balances:      req.InitialBalances,
		TimeoutHeight: req.OpenHeight + req.ChallengePeriod,
		CloseDelay:    req.CloseDelay,
		FeePolicyID:   req.FeePolicyID,
	}
	if req.ChannelType == ChannelTypeUnidirectional {
		state.TimeoutHeight = req.ExpirationHeight
		state.TimeoutTimestamp = req.ExpirationTimestamp
	}
	if req.ChannelType == ChannelTypeAsync {
		state.CheckpointNonce = 1
		state.CheckpointBalances = req.InitialBalances
		state.AsyncUpdateRoot = ComputeAsyncDeltaRoot(nil)
		state.AcceptedUpdateRoot = ComputeAsyncDeltaRoot(nil)
		state.SendWindow = req.CloseDelay
		state.ReceiveWindow = req.ChallengePeriod
		state.MaxUnackedAmount = req.Collateral
		state.ExpiryHeight = req.ExpirationHeight
		state.TimeoutHeight = req.ExpirationHeight
	}
	if channel.ChannelType == ChannelTypeBidirectional && len(req.InitialBalances) == 2 {
		state.ParticipantA = channel.Normalize().Participants[0]
		state.ParticipantB = channel.Normalize().Participants[1]
	}
	return state
}

func validateInitialBalances(balances []Balance, participants []string, collateralText string) error {
	if len(balances) != len(participants) {
		return errors.New("payments initial balances must include every participant")
	}
	for _, balance := range normalizeBalances(balances) {
		if !containsString(participants, balance.Participant) {
			return errors.New("payments initial balance participant must be in channel")
		}
	}
	total, err := sumBalances(balances)
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments open collateral", collateralText)
	if err != nil {
		return err
	}
	if !total.Equal(collateral) {
		return errors.New("payments initial balances must sum to collateral")
	}
	return nil
}

func validateCloseDelay(closeDelay uint64) error {
	if closeDelay < MinCloseDelay || closeDelay > MaxCloseDelay {
		return fmt.Errorf("payments close delay must be between %d and %d", MinCloseDelay, MaxCloseDelay)
	}
	return nil
}

func validateChallengePeriod(period uint64) error {
	if period < MinChallengePeriod || period > MaxChallengePeriod {
		return fmt.Errorf("payments challenge period must be between %d and %d", MinChallengePeriod, MaxChallengePeriod)
	}
	return nil
}

func validateOpeningFeePaid(feePaid string) error {
	paid, err := parseNonNegativeInt("payments opening fee paid", feePaid)
	if err != nil {
		return err
	}
	required, err := parsePositiveInt("payments opening fee required", DefaultOpeningFee)
	if err != nil {
		return err
	}
	if paid.LT(required) {
		return errors.New("payments opening fee is not paid")
	}
	return nil
}

func validateUnsignedUnidirectionalClaim(claim UnidirectionalClaim) error {
	if err := ValidateHash("payments claim channel id", claim.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments claim payer", claim.Payer); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments claim receiver", claim.Receiver); err != nil {
		return err
	}
	if claim.Payer == claim.Receiver {
		return errors.New("payments claim parties must differ")
	}
	if err := validatePositiveInt("payments claim locked amount", claim.LockedAmount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments claim claimed amount", claim.ClaimedAmount); err != nil {
		return err
	}
	if claim.Nonce == 0 {
		return errors.New("payments claim nonce must be positive")
	}
	if claim.ExpirationHeight == 0 {
		return errors.New("payments claim expiration height must be positive")
	}
	if claim.ExpirationTimestamp < 0 {
		return errors.New("payments claim expiration timestamp must be non-negative")
	}
	return nil
}

func validateUnsignedAsyncDelta(delta AsyncPaymentDelta) error {
	if err := ValidateHash("payments async delta update id", delta.UpdateID); err != nil {
		return err
	}
	if err := ValidateHash("payments async delta channel id", delta.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments async delta from", delta.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments async delta to", delta.To); err != nil {
		return err
	}
	if delta.From == delta.To {
		return errors.New("payments async delta parties must differ")
	}
	if delta.Direction != AsyncDeltaDirection(delta.From, delta.To) {
		return errors.New("payments async delta direction mismatch")
	}
	if err := validatePositiveInt("payments async delta amount", delta.Amount); err != nil {
		return err
	}
	if delta.NonceStart == 0 || delta.NonceEnd < delta.NonceStart {
		return errors.New("payments async delta nonce range is invalid")
	}
	if delta.ExpiryHeight == 0 {
		return errors.New("payments async delta expiry height must be positive")
	}
	return nil
}

func validateUnidirectionalChannelCore(channel ChannelRecord) error {
	if len(channel.Participants) != 2 {
		return errors.New("payments unidirectional channel requires exactly two participants")
	}
	if err := addressing.ValidateUserAddress("payments unidirectional payer", channel.Payer); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments unidirectional receiver", channel.Receiver); err != nil {
		return err
	}
	if channel.Payer == channel.Receiver {
		return errors.New("payments unidirectional parties must differ")
	}
	if !containsString(channel.Participants, channel.Payer) || !containsString(channel.Participants, channel.Receiver) {
		return errors.New("payments unidirectional parties must be channel participants")
	}
	if channel.ExpirationHeight == 0 {
		return errors.New("payments unidirectional expiration height must be positive")
	}
	if channel.ExpirationTimestamp < 0 {
		return errors.New("payments unidirectional expiration timestamp must be non-negative")
	}
	return nil
}

func validateUnidirectionalOpeningState(channel ChannelRecord) error {
	if channel.LatestState.TimeoutHeight != channel.ExpirationHeight {
		return errors.New("payments unidirectional opening state expiration height mismatch")
	}
	if channel.LatestState.TimeoutTimestamp != channel.ExpirationTimestamp {
		return errors.New("payments unidirectional opening state expiration timestamp mismatch")
	}
	balanceByParticipant := map[string]string{}
	for _, balance := range channel.LatestState.Balances {
		balanceByParticipant[balance.Participant] = balance.Amount
	}
	payerBalance, err := parseNonNegativeInt("payments unidirectional payer opening balance", balanceByParticipant[channel.Payer])
	if err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	if !payerBalance.Equal(collateral) {
		return errors.New("payments unidirectional payer must lock full collateral on open")
	}
	receiverBalance, err := parseNonNegativeInt("payments unidirectional receiver opening balance", balanceByParticipant[channel.Receiver])
	if err != nil {
		return err
	}
	if !receiverBalance.IsZero() {
		return errors.New("payments unidirectional receiver opening balance must be zero")
	}
	return nil
}

func validateUnidirectionalClaimProgress(previous, next UnidirectionalClaim) error {
	if previous.IsZero() {
		return nil
	}
	previous = previous.Normalize()
	next = next.Normalize()
	if next.Nonce <= previous.Nonce {
		return errors.New("payments claim nonce must strictly increase")
	}
	previousClaimed, err := parseNonNegativeInt("payments previous claimed amount", previous.ClaimedAmount)
	if err != nil {
		return err
	}
	nextClaimed, err := parseNonNegativeInt("payments claimed amount", next.ClaimedAmount)
	if err != nil {
		return err
	}
	if nextClaimed.LT(previousClaimed) {
		return errors.New("payments claimed amount must not decrease")
	}
	return nil
}

func validateStateParticipants(state ChannelState, channel ChannelRecord) error {
	if channel.ChannelType == ChannelTypeBidirectional {
		if err := validateBidirectionalProjection(state, channel); err != nil {
			return err
		}
	}
	if channel.ChannelType == ChannelTypeAsync {
		if err := validateAsyncStateProjection(state, channel); err != nil {
			return err
		}
	}
	for _, balance := range state.Balances {
		if !containsString(channel.Participants, balance.Participant) {
			return errors.New("payments balance participant must be in channel")
		}
	}
	for _, condition := range state.Conditions {
		if !containsString(channel.Participants, condition.Payer) || !containsString(channel.Participants, condition.Payee) {
			return errors.New("payments condition parties must be in channel")
		}
	}
	return nil
}

func validateAsyncStateProjection(state ChannelState, channel ChannelRecord) error {
	if state.CheckpointNonce == 0 {
		return errors.New("payments async checkpoint nonce must be positive")
	}
	if state.CheckpointNonce != state.Nonce {
		return errors.New("payments async checkpoint nonce must match state nonce")
	}
	if err := validateBalances(state.CheckpointBalances); err != nil {
		return err
	}
	if !sameBalances(state.Balances, state.CheckpointBalances) {
		return errors.New("payments async checkpoint balances mismatch")
	}
	if err := ValidateHash("payments async update root", state.AsyncUpdateRoot); err != nil {
		return err
	}
	if err := ValidateHash("payments async accepted update root", state.AcceptedUpdateRoot); err != nil {
		return err
	}
	if state.SendWindow == 0 {
		return errors.New("payments async send window must be positive")
	}
	if state.ReceiveWindow == 0 {
		return errors.New("payments async receive window must be positive")
	}
	if err := validatePositiveInt("payments async max unacked amount", state.MaxUnackedAmount); err != nil {
		return err
	}
	if state.ExpiryHeight == 0 {
		return errors.New("payments async expiry height must be positive")
	}
	if channel.LatestState.StateHash != "" && channel.LatestState.ChannelType == ChannelTypeAsync {
		previous := channel.LatestState.Normalize()
		if previous.SendWindow != 0 && state.SendWindow != previous.SendWindow {
			return errors.New("payments async send window cannot change inside checkpoint")
		}
		if previous.ReceiveWindow != 0 && state.ReceiveWindow != previous.ReceiveWindow {
			return errors.New("payments async receive window cannot change inside checkpoint")
		}
		if previous.MaxUnackedAmount != "" && state.MaxUnackedAmount != previous.MaxUnackedAmount {
			return errors.New("payments async max unacked amount cannot change inside checkpoint")
		}
	}
	return nil
}

func validateBidirectionalProjection(state ChannelState, channel ChannelRecord) error {
	if len(channel.Participants) != 2 {
		return errors.New("payments bidirectional channel requires exactly two participants")
	}
	if len(state.Balances) != 2 {
		return errors.New("payments bidirectional state requires exactly two balances")
	}
	if state.ParticipantA == "" || state.ParticipantB == "" {
		return errors.New("payments bidirectional state participants are required")
	}
	if state.ParticipantA == state.ParticipantB {
		return errors.New("payments bidirectional state participants must differ")
	}
	if state.ParticipantA != channel.Participants[0] || state.ParticipantB != channel.Participants[1] {
		return errors.New("payments bidirectional state participants must match canonical channel order")
	}
	if state.TimeoutHeight == 0 {
		return errors.New("payments bidirectional state timeout height must be positive")
	}
	if state.CloseDelay == 0 {
		return errors.New("payments bidirectional state close delay must be positive")
	}
	balanceByParticipant := map[string]string{}
	for _, balance := range state.Balances {
		balanceByParticipant[balance.Participant] = balance.Amount
	}
	if balanceByParticipant[state.ParticipantA] != state.BalanceA || balanceByParticipant[state.ParticipantB] != state.BalanceB {
		return errors.New("payments bidirectional state balance projection mismatch")
	}
	if err := validateNonNegativeInt("payments bidirectional reserve a", state.ReserveA); err != nil {
		return err
	}
	return validateNonNegativeInt("payments bidirectional reserve b", state.ReserveB)
}

func validateCollateralConservation(state ChannelState, channel ChannelRecord) error {
	collateral, err := parsePositiveInt("payments channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	if channel.ChannelType == ChannelTypeBidirectional {
		balanceA, err := parseNonNegativeInt("payments bidirectional balance a", state.BalanceA)
		if err != nil {
			return err
		}
		balanceB, err := parseNonNegativeInt("payments bidirectional balance b", state.BalanceB)
		if err != nil {
			return err
		}
		reserveA, err := parseNonNegativeInt("payments bidirectional reserve a", state.ReserveA)
		if err != nil {
			return err
		}
		reserveB, err := parseNonNegativeInt("payments bidirectional reserve b", state.ReserveB)
		if err != nil {
			return err
		}
		total := balanceA.Add(balanceB).Add(reserveA).Add(reserveB)
		if !total.Equal(collateral) {
			return errors.New("payments channel state must conserve collateral")
		}
		return nil
	}
	total, err := sumBalances(state.Balances)
	if err != nil {
		return err
	}
	conditionTotal, err := sumConditions(state.Conditions)
	if err != nil {
		return err
	}
	if !total.Add(conditionTotal).Equal(collateral) {
		return errors.New("payments channel state must conserve collateral")
	}
	return nil
}

func validateSignatureQuorum(signatures []StateSignature, required []string, stateHash string) error {
	if err := validateAddressSet("payments required signer", required, 1, MaxParticipants); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(signatures))
	for i, sig := range signatures {
		sig = sig.Normalize()
		if err := sig.Validate(stateHash); err != nil {
			return err
		}
		if _, found := seen[sig.Signer]; found {
			return errors.New("payments duplicate state signature")
		}
		seen[sig.Signer] = struct{}{}
		if i > 0 && signatures[i-1].Normalize().Signer >= sig.Signer {
			return errors.New("payments state signatures must be sorted canonically")
		}
	}
	for _, signer := range required {
		if _, found := seen[signer]; !found {
			return errors.New("payments state signatures do not satisfy channel quorum")
		}
	}
	return nil
}

func validateBalances(balances []Balance) error {
	if len(balances) == 0 || len(balances) > MaxParticipants {
		return fmt.Errorf("payments balances must be between 1 and %d", MaxParticipants)
	}
	var previous string
	seen := make(map[string]struct{}, len(balances))
	for i, balance := range balances {
		if err := addressing.ValidateUserAddress("payments balance participant", balance.Participant); err != nil {
			return err
		}
		if err := validateNonNegativeInt("payments balance amount", balance.Amount); err != nil {
			return err
		}
		if _, found := seen[balance.Participant]; found {
			return errors.New("payments duplicate balance participant")
		}
		seen[balance.Participant] = struct{}{}
		if i > 0 && previous >= balance.Participant {
			return errors.New("payments balances must be sorted canonically")
		}
		previous = balance.Participant
	}
	return nil
}

func validateConditions(conditions []ConditionalPayment) error {
	if len(conditions) > MaxConditionsPerState {
		return fmt.Errorf("payments conditions must be <= %d", MaxConditionsPerState)
	}
	var previous string
	seen := make(map[string]struct{}, len(conditions))
	for i, condition := range conditions {
		if err := condition.Validate(); err != nil {
			return err
		}
		if _, found := seen[condition.ConditionID]; found {
			return errors.New("payments duplicate condition id")
		}
		seen[condition.ConditionID] = struct{}{}
		if i > 0 && previous >= condition.ConditionID {
			return errors.New("payments conditions must be sorted canonically")
		}
		previous = condition.ConditionID
	}
	return nil
}

func (c ConditionalPayment) Normalize() ConditionalPayment {
	c.ConditionID = normalizeHash(c.ConditionID)
	c.Payer = strings.TrimSpace(c.Payer)
	c.Payee = strings.TrimSpace(c.Payee)
	c.Amount = strings.TrimSpace(c.Amount)
	c.HashLock = normalizeOptionalHash(c.HashLock)
	return c
}

func (c ConditionalPayment) Validate() error {
	condition := c.Normalize()
	if err := ValidateHash("payments condition id", condition.ConditionID); err != nil {
		return err
	}
	if !IsConditionType(condition.ConditionType) {
		return fmt.Errorf("unknown payments condition type %q", condition.ConditionType)
	}
	if err := addressing.ValidateUserAddress("payments condition payer", condition.Payer); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments condition payee", condition.Payee); err != nil {
		return err
	}
	if condition.Payer == condition.Payee {
		return errors.New("payments condition parties must differ")
	}
	if err := validatePositiveInt("payments condition amount", condition.Amount); err != nil {
		return err
	}
	if condition.TimeoutHeight == 0 {
		return errors.New("payments condition timeout height must be positive")
	}
	if condition.NonceStart == 0 || condition.NonceEnd < condition.NonceStart {
		return errors.New("payments condition nonce range is invalid")
	}
	if condition.ConditionType == ConditionTypeHashLock {
		return ValidateHash("payments condition hash lock", condition.HashLock)
	}
	if condition.HashLock != "" {
		return errors.New("payments time-lock condition must not include hash lock")
	}
	return nil
}

func IsConditionType(value ConditionType) bool {
	switch value {
	case ConditionTypeHashLock, ConditionTypeTimeLock:
		return true
	default:
		return false
	}
}

func parsePositiveInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(strings.TrimSpace(value))
	if !ok || !out.IsPositive() {
		return sdkmath.Int{}, fmt.Errorf("%s must be a positive integer", field)
	}
	return out, nil
}

func parseNonNegativeInt(field, value string) (sdkmath.Int, error) {
	out, ok := sdkmath.NewIntFromString(strings.TrimSpace(value))
	if !ok || out.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("%s must be a non-negative integer", field)
	}
	return out, nil
}

func validatePositiveInt(field, value string) error {
	_, err := parsePositiveInt(field, value)
	return err
}

func validateNonNegativeInt(field, value string) error {
	_, err := parseNonNegativeInt(field, value)
	return err
}

func sumBalances(balances []Balance) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, balance := range balances {
		amount, err := parseNonNegativeInt("payments balance amount", balance.Amount)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(amount)
	}
	return total, nil
}

func sumConditions(conditions []ConditionalPayment) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, condition := range conditions {
		amount, err := parsePositiveInt("payments condition amount", condition.Amount)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(amount)
	}
	return total, nil
}

func normalizeBalances(balances []Balance) []Balance {
	out := make([]Balance, len(balances))
	for i, balance := range balances {
		out[i] = Balance{
			Participant: strings.TrimSpace(balance.Participant),
			Amount:      strings.TrimSpace(balance.Amount),
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Participant < out[j].Participant
	})
	return out
}

func normalizeConditions(conditions []ConditionalPayment) []ConditionalPayment {
	out := make([]ConditionalPayment, len(conditions))
	for i, condition := range conditions {
		out[i] = condition.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ConditionID < out[j].ConditionID
	})
	return out
}

func normalizeSignatures(signatures []StateSignature) []StateSignature {
	out := make([]StateSignature, len(signatures))
	for i, sig := range signatures {
		out[i] = sig.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Signer < out[j].Signer
	})
	return out
}

func normalizePaymentEventAttributes(attrs []PaymentEventAttribute) []PaymentEventAttribute {
	out := make([]PaymentEventAttribute, len(attrs))
	for i, attr := range attrs {
		out[i] = attr.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

func normalizeAsyncDeltas(deltas []AsyncPaymentDelta) []AsyncPaymentDelta {
	out := make([]AsyncPaymentDelta, len(deltas))
	for i, delta := range deltas {
		out[i] = delta.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdateID != out[j].UpdateID {
			return out[i].UpdateID < out[j].UpdateID
		}
		return out[i].DeltaHash < out[j].DeltaHash
	})
	return out
}

func validateAsyncDeltasForCheckpoint(channel ChannelRecord, base ChannelState, deltas []AsyncPaymentDelta, checkpointNonce, currentHeight uint64) error {
	if len(deltas) == 0 {
		return errors.New("payments async checkpoint requires signed deltas")
	}
	maxExposure, err := parsePositiveInt("payments async max unacked amount", base.MaxUnackedAmount)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(deltas))
	exposureBySender := make(map[string]sdkmath.Int, len(channel.Participants))
	for _, delta := range normalizeAsyncDeltas(deltas) {
		if _, found := seen[delta.UpdateID]; found {
			return errors.New("payments duplicate async delta update id")
		}
		seen[delta.UpdateID] = struct{}{}
		if err := delta.ValidateForChannel(channel, currentHeight); err != nil {
			return err
		}
		if delta.NonceEnd-delta.NonceStart+1 > base.SendWindow {
			return errors.New("payments async delta exceeds send window")
		}
		if delta.NonceEnd > checkpointNonce {
			return errors.New("payments async delta nonce exceeds checkpoint")
		}
		if checkpointNonce-delta.NonceEnd > base.ReceiveWindow {
			return errors.New("payments async delta is outside receive window")
		}
		amount, err := parsePositiveInt("payments async delta amount", delta.Amount)
		if err != nil {
			return err
		}
		currentExposure, found := exposureBySender[delta.From]
		if !found {
			currentExposure = sdkmath.ZeroInt()
		}
		exposureBySender[delta.From] = currentExposure.Add(amount)
		if exposureBySender[delta.From].GT(maxExposure) {
			return errors.New("payments async max unacked exposure exceeded")
		}
	}
	return nil
}

func applyAsyncDeltas(baseBalances []Balance, deltas []AsyncPaymentDelta) ([]Balance, error) {
	amounts := make(map[string]sdkmath.Int, len(baseBalances))
	for _, balance := range normalizeBalances(baseBalances) {
		amount, err := parseNonNegativeInt("payments async base balance", balance.Amount)
		if err != nil {
			return nil, err
		}
		amounts[balance.Participant] = amount
	}
	for _, delta := range normalizeAsyncDeltas(deltas) {
		amount, err := parsePositiveInt("payments async delta amount", delta.Amount)
		if err != nil {
			return nil, err
		}
		fromBalance, found := amounts[delta.From]
		if !found {
			return nil, errors.New("payments async delta sender has no balance")
		}
		if fromBalance.LT(amount) {
			return nil, errors.New("payments async delta exceeds sender balance")
		}
		if _, found := amounts[delta.To]; !found {
			return nil, errors.New("payments async delta receiver has no balance")
		}
		amounts[delta.From] = fromBalance.Sub(amount)
		amounts[delta.To] = amounts[delta.To].Add(amount)
	}
	out := make([]Balance, 0, len(amounts))
	for participant, amount := range amounts {
		out = append(out, Balance{Participant: participant, Amount: amount.String()})
	}
	return normalizeBalances(out), nil
}

func sameBalances(left, right []Balance) bool {
	left = normalizeBalances(left)
	right = normalizeBalances(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func normalizeFraudProofs(proofs []FraudProof) []FraudProof {
	out := make([]FraudProof, len(proofs))
	for i, proof := range proofs {
		out[i] = proof.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ProofID < out[j].ProofID
	})
	return out
}

func normalizePenalties(penalties []Penalty) []Penalty {
	out := make([]Penalty, len(penalties))
	for i, penalty := range penalties {
		out[i] = penalty.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Offender != out[j].Offender {
			return out[i].Offender < out[j].Offender
		}
		if out[i].Recipient != out[j].Recipient {
			return out[i].Recipient < out[j].Recipient
		}
		return out[i].Amount < out[j].Amount
	})
	return out
}

func normalizeAddressSet(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sortStrings(out)
	return out
}

func normalizeAddress(value string) string {
	return strings.TrimSpace(value)
}

func validateAddressSet(field string, values []string, min, max int) error {
	if len(values) < min || len(values) > max {
		return fmt.Errorf("%s count must be between %d and %d", field, min, max)
	}
	seen := make(map[string]struct{}, len(values))
	var previous string
	for i, value := range values {
		if err := addressing.ValidateUserAddress(field, value); err != nil {
			return err
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s", field)
		}
		seen[value] = struct{}{}
		if i > 0 && previous >= value {
			return fmt.Errorf("%s set must be sorted canonically", field)
		}
		previous = value
	}
	return nil
}

func normalizeHash(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeOptionalHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return normalizeHash(value)
}

func normalizeAssetDenom(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return NativeDenom
	}
	return value
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func stateSignedBy(state ChannelState, signer string) bool {
	for _, sig := range state.Signatures {
		if sig.Normalize().Signer == signer {
			return true
		}
	}
	return false
}

func compareSettlementOperations(left, right SettlementOperation) int {
	if left.ChannelID != right.ChannelID {
		return compareString(left.ChannelID, right.ChannelID)
	}
	if left.OperationType != right.OperationType {
		return compareString(string(left.OperationType), string(right.OperationType))
	}
	if left.Nonce < right.Nonce {
		return -1
	}
	if left.Nonce > right.Nonce {
		return 1
	}
	return compareString(left.OperationID, right.OperationID)
}
