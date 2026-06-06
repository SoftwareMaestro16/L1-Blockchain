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
	CanonicalEncodingVersion = byte(1)
	CurrentAppVersion        = uint32(1)
	CurrentStateVersion      = uint32(1)
	SignatureSchemeEd25519   = "ed25519-aetheris-v1"
	SignatureObjectState     = "channel_state"
	SignatureObjectClaim     = "unidirectional_claim"
	SignatureObjectDelta     = "async_delta"
	SignatureObjectPromise   = "conditional_promise"
	DefaultDisputePeriod     = uint64(16)
	DefaultOpeningFee        = "1"
	MaxDisputeExtensions     = uint32(2)
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
	MaxPenaltyRouteBps       = uint32(10_000)
	DefaultTimeoutMargin     = uint64(16)
	DefaultReplayHorizon     = uint64(100_000)
	SignerIsolationProcess   = "process"
	SignerIsolationHardware  = "hardware"
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

type ChannelFinality string

const (
	ChannelFinalityOpen                       ChannelFinality = "OPEN"
	ChannelFinalityPendingClose               ChannelFinality = "PENDING_CLOSE"
	ChannelFinalityInDispute                  ChannelFinality = "IN_DISPUTE"
	ChannelFinalityPendingConditionResolution ChannelFinality = "PENDING_CONDITION_RESOLUTION"
	ChannelFinalityFinalizable                ChannelFinality = "FINALIZABLE"
	ChannelFinalitySettled                    ChannelFinality = "SETTLED"
	ChannelFinalityPenalized                  ChannelFinality = "PENALIZED"
	ChannelFinalityExpired                    ChannelFinality = "EXPIRED"
)

func CanonicalStateRequiredFields() []string {
	return []string{
		"chain_id",
		"app_version",
		"module_name",
		"channel_id",
		"channel_type",
		"participant_set_hash",
		"balances",
		"reserves",
		"pending_condition_amounts",
		"accrued_fees",
		"nonce",
		"epoch",
		"previous_state_hash",
		"timeout_height",
		"timeout_timestamp",
		"challenge_period",
		"condition_root",
		"condition_count",
		"required_signer_bitmap",
		"signature_scheme",
		"signature_preimage_hash",
	}
}

type ConditionType string

const (
	ConditionTypeHashLock ConditionType = "HASH_LOCK"
	ConditionTypeTimeLock ConditionType = "TIME_LOCK"
)

type FraudProofType string

const (
	FraudProofTypeDoubleSign        FraudProofType = "DOUBLE_SIGN"
	FraudProofTypeStaleClose        FraudProofType = "STALE_CLOSE"
	FraudProofTypeInvalidClose      FraudProofType = "INVALID_CLOSE"
	FraudProofTypeInvalidBalance    FraudProofType = "INVALID_BALANCE"
	FraudProofTypeInvalidCondition  FraudProofType = "INVALID_CONDITION"
	FraudProofTypeReplayAttempt     FraudProofType = "REPLAY_ATTEMPT"
	FraudProofTypeAsyncOverexposure FraudProofType = "ASYNC_OVEREXPOSURE"
)

type BatchOperationType string

const (
	BatchOperationOpen    BatchOperationType = "OPEN"
	BatchOperationClose   BatchOperationType = "CLOSE"
	BatchOperationDispute BatchOperationType = "DISPUTE"
	BatchOperationSettle  BatchOperationType = "SETTLE"
)

type CloseReason string

const (
	CloseReasonUnilateral  CloseReason = "UNILATERAL"
	CloseReasonCooperative CloseReason = "COOPERATIVE"
	CloseReasonTimeout     CloseReason = "TIMEOUT"
	CloseReasonFraud       CloseReason = "FRAUD"
)

type SettlementArbitrationOperation string

const (
	SettlementArbitrationOpen                SettlementArbitrationOperation = "OPEN"
	SettlementArbitrationCollateralCustody   SettlementArbitrationOperation = "COLLATERAL_CUSTODY"
	SettlementArbitrationCooperativeClose    SettlementArbitrationOperation = "COOPERATIVE_CLOSE"
	SettlementArbitrationUnilateralClose     SettlementArbitrationOperation = "UNILATERAL_CLOSE"
	SettlementArbitrationDispute             SettlementArbitrationOperation = "DISPUTE"
	SettlementArbitrationFraudProof          SettlementArbitrationOperation = "FRAUD_PROOF"
	SettlementArbitrationConditionResolution SettlementArbitrationOperation = "CONDITION_RESOLUTION"
	SettlementArbitrationPenaltyRouting      SettlementArbitrationOperation = "PENALTY_ROUTING"
	SettlementArbitrationFinalSettlement     SettlementArbitrationOperation = "FINAL_SETTLEMENT"
	SettlementArbitrationReplayProtection    SettlementArbitrationOperation = "REPLAY_PROTECTION"
)

type VirtualChannelStatus string

const (
	VirtualChannelStatusOpen    VirtualChannelStatus = "OPEN"
	VirtualChannelStatusSettled VirtualChannelStatus = "SETTLED"
)

type PenaltyRoute string

const (
	PenaltyRouteReporter        PenaltyRoute = "REPORTER"
	PenaltyRouteBurn            PenaltyRoute = "BURN"
	PenaltyRouteSecurityReserve PenaltyRoute = "SECURITY_RESERVE"
	PenaltyRouteCommunityPool   PenaltyRoute = "COMMUNITY_POOL"
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

type ConditionalPromise struct {
	PromiseID                 string
	ChannelID                 string
	Source                    string
	Destination               string
	Amount                    string
	Fee                       string
	HashLock                  string
	TimeoutHeight             uint64
	TimeoutTimestamp          int64
	ConditionType             ConditionType
	RouteIDOptional           string
	PreviousPromiseIDOptional string
	NextPromiseIDOptional     string
	Nonce                     uint64
	PromiseHash               string
	Signature                 PromiseSignature
}

type PromiseSignature struct {
	Signer           string
	ChainID          string
	ChannelID        string
	ObjectType       string
	Version          uint32
	Nonce            uint64
	ObjectID         string
	ExpirationHeight uint64
	CommitmentHash   string
	PromiseHash      string
	SignatureHash    string
}

type StateSignature struct {
	Signer           string
	ChainID          string
	ChannelID        string
	ObjectType       string
	Version          uint32
	Nonce            uint64
	ObjectID         string
	ExpirationHeight uint64
	CommitmentHash   string
	StateHash        string
	SignatureHash    string
}

type SignedNonceRecord struct {
	Signer        string
	ChainID       string
	ChannelID     string
	Epoch         uint64
	Nonce         uint64
	StateHash     string
	WALHash       string
	Released      bool
	IsolationMode string
}

type SignerPersistence struct {
	Records       []SignedNonceRecord
	IsolationMode string
}

func (r SignedNonceRecord) Normalize() SignedNonceRecord {
	r.Signer = strings.TrimSpace(r.Signer)
	r.ChainID = strings.TrimSpace(r.ChainID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.StateHash = normalizeHash(r.StateHash)
	r.WALHash = normalizeOptionalHash(r.WALHash)
	r.IsolationMode = strings.TrimSpace(r.IsolationMode)
	if r.IsolationMode == "" {
		r.IsolationMode = SignerIsolationProcess
	}
	return r
}

func (p SignerPersistence) Normalize() SignerPersistence {
	p.IsolationMode = strings.TrimSpace(p.IsolationMode)
	if p.IsolationMode == "" {
		p.IsolationMode = SignerIsolationProcess
	}
	p.Records = normalizeSignedNonceRecords(p.Records)
	return p
}

func (p SignerPersistence) HighestSignedNonce(signer, chainID, channelID string, epoch uint64) uint64 {
	p = p.Normalize()
	signer = strings.TrimSpace(signer)
	chainID = strings.TrimSpace(chainID)
	channelID = normalizeHash(channelID)
	var highest uint64
	for _, record := range p.Records {
		if record.Signer == signer && record.ChainID == chainID && record.ChannelID == channelID && record.Epoch == epoch && record.Nonce > highest {
			highest = record.Nonce
		}
	}
	return highest
}

func (p SignerPersistence) SignState(state ChannelState, signer string) (SignerPersistence, StateSignature, error) {
	p = p.Normalize()
	records, sig, err := SignStateWithWriteAhead(p.Records, state, signer, p.IsolationMode)
	if err != nil {
		return p, StateSignature{}, err
	}
	p.Records = records
	return p.Normalize(), sig, nil
}

type ClaimSignature struct {
	Signer           string
	ChainID          string
	ChannelID        string
	ObjectType       string
	Version          uint32
	Nonce            uint64
	ObjectID         string
	ExpirationHeight uint64
	CommitmentHash   string
	ClaimHash        string
	SignatureHash    string
}

type DeltaSignature struct {
	Signer           string
	ChainID          string
	ChannelID        string
	ObjectType       string
	Version          uint32
	Nonce            uint64
	ObjectID         string
	ExpirationHeight uint64
	CommitmentHash   string
	DeltaHash        string
	SignatureHash    string
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

type ChannelUpdateRequest struct {
	ChannelID            string
	State                ChannelState
	ConditionCommitments []ConditionalPayment
	AsyncDeltas          []AsyncPaymentDelta
	RegisterCheckpoint   bool
	Submitter            string
	CurrentHeight        uint64
}

type ChannelUpdateResult struct {
	ChannelID            string
	StateHash            string
	Nonce                uint64
	ValidatedOffChain    bool
	CheckpointRegistered bool
	Liquidity            []Balance
}

type ChannelCloseRequest struct {
	ChannelID     string
	ClosingState  ChannelState
	Signatures    []StateSignature
	CloseReason   CloseReason
	Submitter     string
	CurrentHeight uint64
	SettlementFee string
}

type ConditionResolution struct {
	ConditionID  string
	Resolver     string
	Recipient    string
	Amount       string
	Expired      bool
	EvidenceHash string
}

type ClosedChannelTombstone struct {
	ChainID        string
	ChannelID      string
	FinalizedNonce uint64
	StateHash      string
	ClosedHeight   uint64
	ExpiresHeight  uint64
}

type ConditionClaimRecord struct {
	ChainID        string
	ChannelID      string
	ConditionID    string
	EvidenceHash   string
	PreimageHash   string
	ResolvedHeight uint64
	ExpiresHeight  uint64
}

type PreimageRevealRequest struct {
	ChannelID     string
	Promises      []ConditionalPromise
	Preimage      string
	Revealer      string
	CurrentHeight uint64
}

type PromiseExpiryRequest struct {
	ChannelID     string
	Promises      []ConditionalPromise
	Resolver      string
	CurrentHeight uint64
}

type ConditionRootUpdate struct {
	ChannelID      string
	Nonce          uint64
	ConditionRoot  string
	ConditionCount uint32
	Conditions     []ConditionalPayment
}

type ChannelDisputeRequest struct {
	ChannelID             string
	ClosingStateReference string
	NewerState            ChannelState
	FraudProof            FraudProof
	ConditionProofs       []ConditionResolution
	Submitter             string
	CurrentHeight         uint64
}

type WatchDisputeSubmission struct {
	WatchService          string
	Delegator             string
	ChannelID             string
	ClosingStateReference string
	NewerState            ChannelState
	CurrentHeight         uint64
	EvidenceHash          string
}

type FinalSettlementRequest struct {
	ChannelID           string
	ResolvedConditions  []ConditionResolution
	CurrentHeight       uint64
	FeeAccountingState  string
	RoutingFeeClaimHash string
}

type SettlementArbitrationInput struct {
	Operation         SettlementArbitrationOperation
	ChannelID         string
	SignedState       ChannelState
	Claim             UnidirectionalClaim
	FraudProof        FraudProof
	ConditionProofs   []ConditionResolution
	RouteHints        []ChannelEdge
	GossipStateHash   string
	ExternalLiquidity []Balance
	UnsignedBalances  []Balance
	OffchainIntent    string
	CurrentHeight     uint64
}

type StateHashDebug struct {
	ChannelID                string
	Status                   ChannelStatus
	LatestNonce              uint64
	LatestStateHash          string
	ComputedLatestStateHash  string
	PendingNonce             uint64
	PendingStateHash         string
	ComputedPendingStateHash string
	FinalizedNonce           uint64
	DisputedNonce            uint64
}

type ChannelState struct {
	ChainID               string
	AppVersion            uint32
	ModuleName            string
	RequiredFields        []string
	ChannelID             string
	ChannelType           ChannelType
	ParticipantSetHash    string
	Denom                 string
	Version               uint32
	ParticipantA          string
	ParticipantB          string
	BalanceA              string
	BalanceB              string
	ReserveA              string
	ReserveB              string
	AccruedFees           string
	Epoch                 uint64
	Nonce                 uint64
	PendingConditionsRoot string
	ConditionRoot         string
	ConditionCount        uint32
	Balances              []Balance
	Conditions            []ConditionalPayment
	PreviousStateHash     string
	StateHash             string
	TimeoutHeight         uint64
	TimeoutTimestamp      int64
	ChallengePeriod       uint64
	CloseDelay            uint64
	FeePolicyID           string
	RequiredSignerBitmap  string
	SignatureScheme       string
	SignaturePreimageHash string
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
	ChainID      string
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
	ChainID             string
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
	DisputeCount       uint32
	CloseReason        CloseReason
	SettlementFeeDenom string
	SettlementFee      string
	State              ChannelState
	FraudProofs        []FraudProof
	ConditionProofs    []ConditionResolution
	Penalties          []Penalty
	PenaltyAllocations []PenaltyAllocation
}

type FraudProof struct {
	ProofID         string
	ProofType       FraudProofType
	SubmittedBy     string
	OffendingSigner string
	StateA          ChannelState
	StateB          ChannelState
	AsyncProof      AsyncDeltaDisputeProof
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

type PenaltyAllocation struct {
	Offender string
	Route    PenaltyRoute
	Denom    string
	Amount   string
}

type FraudPenaltyPolicy struct {
	ReporterRewardCap       string
	BurnShareBps            uint32
	SecurityReserveShareBps uint32
	CommunityPoolShareBps   uint32
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
	Finality            ChannelFinality
	OpeningStateHash    string
	FinalizedNonce      uint64
	DisputedNonce       uint64
	LatestState         ChannelState
	LatestClaim         UnidirectionalClaim
	PendingClose        PendingClose
}

type SettlementRecord struct {
	ChainID            string
	ChannelID          string
	StateHash          string
	Nonce              uint64
	FinalBalances      []Balance
	SettlementFeeDenom string
	SettlementFee      string
	Penalties          []Penalty
	PenaltyAllocations []PenaltyAllocation
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
	ChainID          string
	Nonce            uint64
	ParentChannelIDs []string
	Endpoints        []string
	Capacity         string
	ExpiresHeight    uint64
	Status           VirtualChannelStatus
	AnchorCommitment string
	StateHash        string
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
	state.SignaturePreimageHash = ComputeStateSignaturePreimageHash(state)
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
		Signer:           signer,
		ChainID:          state.ChainID,
		ChannelID:        state.ChannelID,
		ObjectType:       SignatureObjectState,
		Version:          state.Version,
		Nonce:            state.Nonce,
		ObjectID:         state.StateHash,
		ExpirationHeight: state.TimeoutHeight,
		CommitmentHash:   state.StateHash,
		StateHash:        state.StateHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			state.ChainID,
			state.ChannelID,
			SignatureObjectState,
			state.Version,
			state.Nonce,
			state.StateHash,
			state.TimeoutHeight,
			state.StateHash,
		),
	}, nil
}

func SignStateWithWriteAhead(records []SignedNonceRecord, state ChannelState, signer, isolationMode string) ([]SignedNonceRecord, StateSignature, error) {
	state = state.Normalize()
	if state.StateHash == "" {
		var err error
		state, err = BuildState(state)
		if err != nil {
			return nil, StateSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments signer wal signer", signer); err != nil {
		return nil, StateSignature{}, err
	}
	isolationMode = strings.TrimSpace(isolationMode)
	if isolationMode == "" {
		isolationMode = SignerIsolationProcess
	}
	if isolationMode != SignerIsolationProcess && isolationMode != SignerIsolationHardware {
		return nil, StateSignature{}, errors.New("payments signer isolation mode is unsupported")
	}
	normalized := normalizeSignedNonceRecords(records)
	var highest uint64
	for _, record := range normalized {
		if record.Signer == signer && record.ChainID == state.ChainID && record.ChannelID == state.ChannelID && record.Epoch == state.Epoch && record.Nonce > highest {
			highest = record.Nonce
		}
	}
	if highest > 0 && state.Nonce < highest {
		return nil, StateSignature{}, errors.New("payments signer refuses nonce below highest signed nonce")
	}
	for i, record := range normalized {
		if record.Signer != signer || record.ChainID != state.ChainID || record.ChannelID != state.ChannelID || record.Epoch != state.Epoch || record.Nonce != state.Nonce {
			continue
		}
		if record.StateHash != state.StateHash {
			return nil, StateSignature{}, errors.New("payments signer refuses same nonce replacement")
		}
		if record.Released {
			sig, err := SignatureForState(state, signer)
			return normalized, sig, err
		}
		normalized[i].Released = true
		sig, err := SignatureForState(state, signer)
		return normalized, sig, err
	}
	record := SignedNonceRecord{
		Signer:        signer,
		ChainID:       state.ChainID,
		ChannelID:     state.ChannelID,
		Epoch:         state.Epoch,
		Nonce:         state.Nonce,
		StateHash:     state.StateHash,
		IsolationMode: isolationMode,
	}
	record.WALHash = ComputeSignedNonceWALHash(record)
	normalized = append(normalized, record)
	normalized = normalizeSignedNonceRecords(normalized)
	for i := range normalized {
		if normalized[i].WALHash == record.WALHash {
			normalized[i].Released = true
			break
		}
	}
	sig, err := SignatureForState(state, signer)
	if err != nil {
		return nil, StateSignature{}, err
	}
	return normalized, sig, nil
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

func (r ChannelUpdateRequest) Normalize() ChannelUpdateRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.State = r.State.Normalize()
	r.ConditionCommitments = normalizeConditions(r.ConditionCommitments)
	r.AsyncDeltas = normalizeAsyncDeltas(r.AsyncDeltas)
	r.Submitter = strings.TrimSpace(r.Submitter)
	return r
}

func (r ChannelCloseRequest) Normalize() ChannelCloseRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ClosingState = r.ClosingState.Normalize()
	r.Signatures = normalizeStateSignatures(r.Signatures)
	r.Submitter = strings.TrimSpace(r.Submitter)
	r.SettlementFee = strings.TrimSpace(r.SettlementFee)
	if r.CloseReason == "" {
		r.CloseReason = CloseReasonUnilateral
	}
	return r
}

func (r ChannelCloseRequest) ClosingStateWithSignatures() ChannelState {
	req := r.Normalize()
	state := req.ClosingState
	if len(req.Signatures) > 0 {
		state.Signatures = req.Signatures
	}
	return state.Normalize()
}

func (r ChannelCloseRequest) ValidateForChannel(channel ChannelRecord) error {
	req := r.Normalize()
	channel = channel.Normalize()
	if req.ChannelID != channel.ChannelID {
		return errors.New("payments close request channel mismatch")
	}
	if err := validateCloseReason(req.CloseReason); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments close submitter", req.Submitter); err != nil {
		return err
	}
	if !containsString(channel.Participants, req.Submitter) {
		return errors.New("payments close submitter must be participant")
	}
	if req.CurrentHeight == 0 {
		return errors.New("payments close height must be positive")
	}
	if err := validateNonNegativeInt("payments settlement fee", req.SettlementFee); err != nil {
		return err
	}
	return req.ClosingStateWithSignatures().ValidateForChannel(channel, false)
}

func (r ChannelDisputeRequest) Normalize() ChannelDisputeRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ClosingStateReference = normalizeHash(r.ClosingStateReference)
	r.NewerState = r.NewerState.Normalize()
	r.FraudProof = r.FraudProof.Normalize()
	r.ConditionProofs = normalizeConditionResolutions(r.ConditionProofs)
	r.Submitter = strings.TrimSpace(r.Submitter)
	return r
}

func (w WatchDisputeSubmission) Normalize() WatchDisputeSubmission {
	w.WatchService = strings.TrimSpace(w.WatchService)
	w.Delegator = strings.TrimSpace(w.Delegator)
	w.ChannelID = normalizeHash(w.ChannelID)
	w.ClosingStateReference = normalizeHash(w.ClosingStateReference)
	w.NewerState = w.NewerState.Normalize()
	w.EvidenceHash = normalizeOptionalHash(w.EvidenceHash)
	return w
}

func (w WatchDisputeSubmission) ValidateForChannel(channel ChannelRecord) error {
	w = w.Normalize()
	channel = channel.Normalize()
	if err := addressing.ValidateUserAddress("payments watch service", w.WatchService); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments watch delegator", w.Delegator); err != nil {
		return err
	}
	if !containsString(channel.Participants, w.Delegator) {
		return errors.New("payments watch delegator must be channel participant")
	}
	if w.ChannelID != channel.ChannelID {
		return errors.New("payments watch dispute channel mismatch")
	}
	if err := ValidateHash("payments watch dispute closing reference", w.ClosingStateReference); err != nil {
		return err
	}
	if w.CurrentHeight == 0 {
		return errors.New("payments watch dispute height must be positive")
	}
	if w.EvidenceHash != "" {
		if err := ValidateHash("payments watch dispute evidence hash", w.EvidenceHash); err != nil {
			return err
		}
	}
	return w.NewerState.ValidateForChannel(channel, false)
}

func (r FinalSettlementRequest) Normalize() FinalSettlementRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ResolvedConditions = normalizeConditionResolutions(r.ResolvedConditions)
	r.FeeAccountingState = strings.TrimSpace(r.FeeAccountingState)
	r.RoutingFeeClaimHash = normalizeOptionalHash(r.RoutingFeeClaimHash)
	return r
}

func BuildConditionalPromise(promise ConditionalPromise) (ConditionalPromise, error) {
	promise = promise.Normalize()
	if err := promise.ValidateBasic(); err != nil {
		return ConditionalPromise{}, err
	}
	promise.PromiseHash = ComputeConditionalTransferPromiseHash(promise)
	return promise, nil
}

func SignatureForPromise(channel ChannelRecord, promise ConditionalPromise, signer string) (PromiseSignature, error) {
	channel = channel.Normalize()
	if promise.PromiseHash == "" {
		var err error
		promise, err = BuildConditionalPromise(promise)
		if err != nil {
			return PromiseSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments promise signer", signer); err != nil {
		return PromiseSignature{}, err
	}
	return PromiseSignature{
		Signer:           signer,
		ChainID:          channel.ChainID,
		ChannelID:        promise.ChannelID,
		ObjectType:       SignatureObjectPromise,
		Version:          CurrentStateVersion,
		Nonce:            promise.Nonce,
		ObjectID:         promise.PromiseHash,
		ExpirationHeight: promise.TimeoutHeight,
		CommitmentHash:   promise.PromiseHash,
		PromiseHash:      promise.PromiseHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			channel.ChainID,
			promise.ChannelID,
			SignatureObjectPromise,
			CurrentStateVersion,
			promise.Nonce,
			promise.PromiseHash,
			promise.TimeoutHeight,
			promise.PromiseHash,
		),
	}, nil
}

func (p ConditionalPromise) Normalize() ConditionalPromise {
	p.PromiseID = normalizeHash(p.PromiseID)
	p.ChannelID = normalizeHash(p.ChannelID)
	p.Source = strings.TrimSpace(p.Source)
	p.Destination = strings.TrimSpace(p.Destination)
	p.Amount = strings.TrimSpace(p.Amount)
	p.Fee = strings.TrimSpace(p.Fee)
	p.HashLock = normalizeOptionalHash(p.HashLock)
	p.RouteIDOptional = normalizeOptionalHash(p.RouteIDOptional)
	p.PreviousPromiseIDOptional = normalizeOptionalHash(p.PreviousPromiseIDOptional)
	p.NextPromiseIDOptional = normalizeOptionalHash(p.NextPromiseIDOptional)
	p.PromiseHash = normalizeOptionalHash(p.PromiseHash)
	p.Signature = p.Signature.Normalize()
	return p
}

func (p ConditionalPromise) ValidateBasic() error {
	promise := p.Normalize()
	if err := ValidateHash("payments promise id", promise.PromiseID); err != nil {
		return err
	}
	if err := ValidateHash("payments promise channel id", promise.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments promise source", promise.Source); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments promise destination", promise.Destination); err != nil {
		return err
	}
	if promise.Source == promise.Destination {
		return errors.New("payments promise parties must differ")
	}
	if err := validatePositiveInt("payments promise amount", promise.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments promise fee", promise.Fee); err != nil {
		return err
	}
	if promise.TimeoutHeight == 0 {
		return errors.New("payments promise timeout height must be positive")
	}
	if promise.TimeoutTimestamp < 0 {
		return errors.New("payments promise timeout timestamp must be non-negative")
	}
	if promise.Nonce == 0 {
		return errors.New("payments promise nonce must be positive")
	}
	if !IsConditionType(promise.ConditionType) {
		return fmt.Errorf("unknown payments promise condition type %q", promise.ConditionType)
	}
	if promise.ConditionType == ConditionTypeHashLock {
		if err := ValidateHash("payments promise hash lock", promise.HashLock); err != nil {
			return err
		}
	} else if promise.HashLock != "" {
		return errors.New("payments time-lock promise must not include hash lock")
	}
	return nil
}

func (p ConditionalPromise) ValidateForChannel(channel ChannelRecord) error {
	promise := p.Normalize()
	channel = channel.Normalize()
	if err := promise.ValidateBasic(); err != nil {
		return err
	}
	if promise.ChannelID != channel.ChannelID {
		return errors.New("payments promise channel mismatch")
	}
	if !containsString(channel.Participants, promise.Source) || !containsString(channel.Participants, promise.Destination) {
		return errors.New("payments promise parties must be channel participants")
	}
	if !channel.ConditionalPayments {
		return errors.New("payments channel does not support conditional promises")
	}
	if err := validatePromiseTimeoutWindow(channel, promise); err != nil {
		return err
	}
	if promise.PromiseHash == "" {
		return errors.New("payments promise hash is required")
	}
	if expected := ComputeConditionalTransferPromiseHash(promise); promise.PromiseHash != expected {
		return errors.New("payments promise hash mismatch")
	}
	if err := promise.Signature.Validate(promise.PromiseHash); err != nil {
		return err
	}
	return validatePromiseSignatureEnvelope(channel, promise.Signature, promise)
}

func (p ConditionalPromise) ToConditionalPayment() ConditionalPayment {
	promise := p.Normalize()
	return ConditionalPayment{
		ConditionID:   promise.PromiseID,
		ConditionType: promise.ConditionType,
		Payer:         promise.Source,
		Payee:         promise.Destination,
		Amount:        promise.Amount,
		HashLock:      promise.HashLock,
		TimeoutHeight: promise.TimeoutHeight,
		NonceStart:    promise.Nonce,
		NonceEnd:      promise.Nonce,
	}.Normalize()
}

func (s PromiseSignature) Normalize() PromiseSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = normalizeOptionalHash(s.ObjectID)
	s.CommitmentHash = normalizeOptionalHash(s.CommitmentHash)
	s.PromiseHash = normalizeOptionalHash(s.PromiseHash)
	s.SignatureHash = normalizeOptionalHash(s.SignatureHash)
	return s
}

func (s PromiseSignature) Validate(expectedPromiseHash string) error {
	s = s.Normalize()
	if err := addressing.ValidateUserAddress("payments promise signature signer", s.Signer); err != nil {
		return err
	}
	if s.PromiseHash != expectedPromiseHash {
		return errors.New("payments promise signature hash mismatch")
	}
	if s.ObjectType != SignatureObjectPromise {
		return errors.New("payments promise signature object type mismatch")
	}
	if s.ObjectID != s.PromiseHash {
		return errors.New("payments promise signature object id mismatch")
	}
	if s.CommitmentHash != s.PromiseHash {
		return errors.New("payments promise signature commitment mismatch")
	}
	if err := ValidateHash("payments promise signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeSignatureEnvelopeHash(s.Signer, s.ChainID, s.ChannelID, s.ObjectType, s.Version, s.Nonce, s.ObjectID, s.ExpirationHeight, s.CommitmentHash); s.SignatureHash != expected {
		return errors.New("payments promise signature hash mismatch")
	}
	return nil
}

func (i SettlementArbitrationInput) Normalize() SettlementArbitrationInput {
	i.ChannelID = normalizeHash(i.ChannelID)
	i.SignedState = i.SignedState.Normalize()
	i.Claim = i.Claim.Normalize()
	i.FraudProof = i.FraudProof.Normalize()
	i.ConditionProofs = normalizeConditionResolutions(i.ConditionProofs)
	for index := range i.RouteHints {
		i.RouteHints[index] = i.RouteHints[index].Normalize()
	}
	i.GossipStateHash = normalizeOptionalHash(i.GossipStateHash)
	i.ExternalLiquidity = normalizeBalances(i.ExternalLiquidity)
	i.UnsignedBalances = normalizeBalances(i.UnsignedBalances)
	i.OffchainIntent = strings.TrimSpace(i.OffchainIntent)
	return i
}

func (i SettlementArbitrationInput) ValidateForChannel(channel ChannelRecord) error {
	input := i.Normalize()
	channel = channel.Normalize()
	if err := channel.ValidateCore(); err != nil {
		return err
	}
	if input.ChannelID != channel.ChannelID {
		return errors.New("payments settlement arbitration channel mismatch")
	}
	if !IsSettlementArbitrationOperation(input.Operation) {
		return fmt.Errorf("unknown payments settlement arbitration operation %q", input.Operation)
	}
	if len(input.RouteHints) > 0 {
		return errors.New("payments settlement contract must not select payment routes")
	}
	if input.GossipStateHash != "" {
		return errors.New("payments settlement contract must not trust gossip state")
	}
	if len(input.ExternalLiquidity) > 0 {
		return errors.New("payments settlement contract must not depend on external liquidity reports")
	}
	if len(input.UnsignedBalances) > 0 {
		return errors.New("payments settlement contract must not accept unsigned balance updates")
	}
	if input.OffchainIntent != "" {
		return errors.New("payments settlement contract must not infer participant intent from unsigned off-chain messages")
	}
	if input.CurrentHeight == 0 && operationRequiresHeight(input.Operation) {
		return errors.New("payments settlement arbitration height must be positive")
	}
	if input.Operation == SettlementArbitrationCollateralCustody {
		return validateSettlementCustody(channel)
	}
	if input.Operation == SettlementArbitrationFraudProof {
		return input.FraudProof.ValidateForChannel(channel)
	}
	if input.Operation == SettlementArbitrationReplayProtection {
		return validateSettlementReplayProtection(channel, input.SignedState)
	}
	if input.Operation == SettlementArbitrationPenaltyRouting {
		if input.FraudProof.ProofID == "" {
			return errors.New("payments settlement penalty routing requires accepted fraud proof")
		}
		return input.FraudProof.ValidateForChannel(channel)
	}
	if input.Operation == SettlementArbitrationConditionResolution {
		if err := validateSettlementSignedState(channel, input.SignedState, false); err != nil {
			return err
		}
		return validateConditionResolutionsForState(input.SignedState, channel, input.ConditionProofs, true)
	}
	if input.Operation == SettlementArbitrationOpen {
		if err := validateSettlementCustody(channel); err != nil {
			return err
		}
		return validateSettlementSignedState(channel, channel.LatestState, true)
	}
	if input.Operation == SettlementArbitrationUnilateralClose && !input.Claim.IsZero() {
		if err := input.Claim.ValidateForChannel(channel); err != nil {
			return err
		}
		return validateSettlementClaimReplayProtection(channel, input.Claim)
	}
	requireAll := input.Operation == SettlementArbitrationCooperativeClose
	if err := validateSettlementSignedState(channel, input.SignedState, requireAll); err != nil {
		return err
	}
	return validateSettlementReplayProtection(channel, input.SignedState)
}

func IsSettlementArbitrationOperation(operation SettlementArbitrationOperation) bool {
	switch operation {
	case SettlementArbitrationOpen,
		SettlementArbitrationCollateralCustody,
		SettlementArbitrationCooperativeClose,
		SettlementArbitrationUnilateralClose,
		SettlementArbitrationDispute,
		SettlementArbitrationFraudProof,
		SettlementArbitrationConditionResolution,
		SettlementArbitrationPenaltyRouting,
		SettlementArbitrationFinalSettlement,
		SettlementArbitrationReplayProtection:
		return true
	default:
		return false
	}
}

func operationRequiresHeight(operation SettlementArbitrationOperation) bool {
	switch operation {
	case SettlementArbitrationUnilateralClose,
		SettlementArbitrationDispute,
		SettlementArbitrationFraudProof,
		SettlementArbitrationFinalSettlement,
		SettlementArbitrationReplayProtection:
		return true
	default:
		return false
	}
}

func validateSettlementCustody(channel ChannelRecord) error {
	if channel.Denom != NativeDenom || channel.CustodyDenom != NativeDenom {
		return fmt.Errorf("payments settlement custody must use %s", NativeDenom)
	}
	if channel.Collateral == "" || channel.CustodyAmount == "" {
		return errors.New("payments settlement custody amount is required")
	}
	if channel.CustodyAmount != channel.Collateral {
		return errors.New("payments settlement custody must equal locked collateral")
	}
	return validatePositiveInt("payments settlement custody amount", channel.CustodyAmount)
}

func validateSettlementSignedState(channel ChannelRecord, signedState ChannelState, requireAllParticipants bool) error {
	state := signedState.Normalize()
	if state.StateHash == "" {
		return errors.New("payments settlement arbitration signed state is required")
	}
	return state.ValidateForChannel(channel, requireAllParticipants)
}

func validateSettlementReplayProtection(channel ChannelRecord, signedState ChannelState) error {
	state := signedState.Normalize()
	if state.StateHash == "" {
		return errors.New("payments settlement replay protection signed state is required")
	}
	if state.Nonce < channel.FinalizedNonce {
		return errors.New("payments settlement replay state nonce is below finalized nonce")
	}
	if channel.Status == ChannelStatusSettled && state.Nonce <= channel.FinalizedNonce {
		return errors.New("payments settlement replay state targets closed channel")
	}
	return state.ValidateForChannel(channel, false)
}

func validateSettlementClaimReplayProtection(channel ChannelRecord, claim UnidirectionalClaim) error {
	claim = claim.Normalize()
	if claim.StateHash == "" {
		return errors.New("payments settlement replay protection signed claim is required")
	}
	if claim.Nonce < channel.FinalizedNonce {
		return errors.New("payments settlement replay claim nonce is below finalized nonce")
	}
	if channel.Status == ChannelStatusSettled && claim.Nonce <= channel.FinalizedNonce {
		return errors.New("payments settlement replay claim targets closed channel")
	}
	return claim.ValidateForChannel(channel)
}

func (r ConditionResolution) Normalize() ConditionResolution {
	r.ConditionID = normalizeHash(r.ConditionID)
	r.Resolver = strings.TrimSpace(r.Resolver)
	r.Recipient = strings.TrimSpace(r.Recipient)
	r.Amount = strings.TrimSpace(r.Amount)
	r.EvidenceHash = normalizeHash(r.EvidenceHash)
	return r
}

func (r ConditionResolution) ValidateForCondition(condition ConditionalPayment, channel ChannelRecord) error {
	resolution := r.Normalize()
	condition = condition.Normalize()
	if resolution.ConditionID != condition.ConditionID {
		return errors.New("payments condition resolution id mismatch")
	}
	if err := addressing.ValidateUserAddress("payments condition resolver", resolution.Resolver); err != nil {
		return err
	}
	if !containsString(channel.Participants, resolution.Resolver) {
		return errors.New("payments condition resolver must be participant")
	}
	if err := addressing.ValidateUserAddress("payments condition resolution recipient", resolution.Recipient); err != nil {
		return err
	}
	if resolution.Recipient != condition.Payer && resolution.Recipient != condition.Payee {
		return errors.New("payments condition resolution recipient must be condition party")
	}
	if resolution.Expired && resolution.Recipient != condition.Payer {
		return errors.New("payments expired condition must return to payer")
	}
	if !resolution.Expired && resolution.Recipient != condition.Payee {
		return errors.New("payments resolved condition must pay payee")
	}
	amount, err := parsePositiveInt("payments condition resolution amount", resolution.Amount)
	if err != nil {
		return err
	}
	conditionAmount, err := parsePositiveInt("payments condition amount", condition.Amount)
	if err != nil {
		return err
	}
	if !amount.Equal(conditionAmount) {
		return errors.New("payments condition resolution amount mismatch")
	}
	return ValidateHash("payments condition resolution evidence hash", resolution.EvidenceHash)
}

func ValidateOffchainUpdate(channel ChannelRecord, req ChannelUpdateRequest) (ChannelUpdateResult, error) {
	channel = channel.Normalize()
	if err := channel.ValidateCore(); err != nil {
		return ChannelUpdateResult{}, err
	}
	if channel.Status != ChannelStatusOpen {
		return ChannelUpdateResult{}, errors.New("payments update requires open channel")
	}
	req = req.Normalize()
	if req.ChannelID != channel.ChannelID {
		return ChannelUpdateResult{}, errors.New("payments update channel mismatch")
	}
	if req.CurrentHeight == 0 {
		return ChannelUpdateResult{}, errors.New("payments update height must be positive")
	}
	if req.Submitter != "" && !containsString(channel.Participants, req.Submitter) {
		return ChannelUpdateResult{}, errors.New("payments update submitter must be participant")
	}
	if req.State.Nonce <= channel.LatestState.Nonce {
		return ChannelUpdateResult{}, errors.New("payments update nonce must increase")
	}
	if err := ValidatePreviousHashContinuity(channel, req.State); err != nil {
		return ChannelUpdateResult{}, err
	}
	if len(req.ConditionCommitments) > 0 {
		if !channel.ConditionalPayments {
			return ChannelUpdateResult{}, errors.New("payments channel does not support conditional payments")
		}
		if err := validateConditions(req.ConditionCommitments); err != nil {
			return ChannelUpdateResult{}, err
		}
		if req.State.PendingConditionsRoot != ComputeConditionsRoot(req.ConditionCommitments) {
			return ChannelUpdateResult{}, errors.New("payments update condition commitment root mismatch")
		}
	}
	if err := req.State.ValidateForChannel(channel, false); err != nil {
		return ChannelUpdateResult{}, err
	}
	if err := validateUpdateExposure(req.State); err != nil {
		return ChannelUpdateResult{}, err
	}
	if len(req.AsyncDeltas) > 0 {
		reconstructed, err := BuildAsyncCheckpointState(channel, req.AsyncDeltas, req.State.CheckpointNonce, req.CurrentHeight)
		if err != nil {
			return ChannelUpdateResult{}, err
		}
		if reconstructed.StateHash != req.State.StateHash {
			return ChannelUpdateResult{}, errors.New("payments async update checkpoint mismatch")
		}
	}
	return ChannelUpdateResult{
		ChannelID:         channel.ChannelID,
		StateHash:         req.State.StateHash,
		Nonce:             req.State.Nonce,
		ValidatedOffChain: true,
		Liquidity:         req.State.Balances,
	}, nil
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
		Signer:           signer,
		ChainID:          claim.ChainID,
		ChannelID:        claim.ChannelID,
		ObjectType:       SignatureObjectClaim,
		Version:          CurrentStateVersion,
		Nonce:            claim.Nonce,
		ObjectID:         claim.StateHash,
		ExpirationHeight: claim.ExpirationHeight,
		CommitmentHash:   claim.StateHash,
		ClaimHash:        claim.StateHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			claim.ChainID,
			claim.ChannelID,
			SignatureObjectClaim,
			CurrentStateVersion,
			claim.Nonce,
			claim.StateHash,
			claim.ExpirationHeight,
			claim.StateHash,
		),
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
		Signer:           signer,
		ChainID:          delta.ChainID,
		ChannelID:        delta.ChannelID,
		ObjectType:       SignatureObjectDelta,
		Version:          CurrentStateVersion,
		Nonce:            delta.NonceStart,
		ObjectID:         delta.UpdateID,
		ExpirationHeight: delta.ExpiryHeight,
		CommitmentHash:   delta.DeltaHash,
		DeltaHash:        delta.DeltaHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			delta.ChainID,
			delta.ChannelID,
			SignatureObjectDelta,
			CurrentStateVersion,
			delta.NonceStart,
			delta.UpdateID,
			delta.ExpiryHeight,
			delta.DeltaHash,
		),
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
		ChainID:              channel.ChainID,
		AppVersion:           CurrentAppVersion,
		ModuleName:           ModuleName,
		ChannelID:            channel.ChannelID,
		ChannelType:          ChannelTypeAsync,
		ParticipantSetHash:   ComputeParticipantSetHash(channel.Participants),
		Denom:                channel.Denom,
		Version:              CurrentStateVersion,
		Epoch:                base.Epoch,
		Nonce:                checkpointNonce,
		Balances:             nextBalances,
		CheckpointNonce:      checkpointNonce,
		CheckpointBalances:   nextBalances,
		AsyncUpdateRoot:      ComputeAsyncDeltaRootForChannel(channel, normalizedDeltas),
		AcceptedUpdateRoot:   ComputeAsyncDeltaRootForChannel(channel, normalizedDeltas),
		SendWindow:           base.SendWindow,
		ReceiveWindow:        base.ReceiveWindow,
		MaxUnackedAmount:     base.MaxUnackedAmount,
		ExpiryHeight:         base.ExpiryHeight,
		TimeoutHeight:        base.TimeoutHeight,
		TimeoutTimestamp:     base.TimeoutTimestamp,
		ChallengePeriod:      base.ChallengePeriod,
		CloseDelay:           base.CloseDelay,
		FeePolicyID:          NativeDenom,
		RequiredSignerBitmap: ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners),
		SignatureScheme:      SignatureSchemeEd25519,
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
		ChainID:             channel.ChainID,
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
	if s.AppVersion == 0 {
		s.AppVersion = CurrentAppVersion
	}
	s.ModuleName = strings.TrimSpace(s.ModuleName)
	if s.ModuleName == "" {
		s.ModuleName = ModuleName
	}
	s.RequiredFields = normalizeRequiredFields(s.RequiredFields)
	if len(s.RequiredFields) == 0 {
		s.RequiredFields = CanonicalStateRequiredFields()
	}
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ParticipantSetHash = normalizeOptionalHash(s.ParticipantSetHash)
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
	s.AccruedFees = strings.TrimSpace(s.AccruedFees)
	s.PreviousStateHash = normalizeOptionalHash(s.PreviousStateHash)
	s.StateHash = normalizeOptionalHash(s.StateHash)
	s.Balances = normalizeBalances(s.Balances)
	s.Conditions = normalizeConditions(s.Conditions)
	s.PendingConditionsRoot = normalizeOptionalHash(s.PendingConditionsRoot)
	s.ConditionRoot = normalizeOptionalHash(s.ConditionRoot)
	if s.ConditionRoot == "" && s.PendingConditionsRoot != "" {
		s.ConditionRoot = s.PendingConditionsRoot
	}
	if s.ConditionRoot == "" {
		s.ConditionRoot = ComputeConditionsRoot(s.Conditions)
	}
	if s.PendingConditionsRoot == "" {
		s.PendingConditionsRoot = s.ConditionRoot
	}
	if s.ConditionCount == 0 && len(s.Conditions) > 0 {
		s.ConditionCount = uint32(len(s.Conditions))
	}
	if s.ChallengePeriod == 0 {
		s.ChallengePeriod = s.CloseDelay
	}
	if s.FeePolicyID == "" {
		s.FeePolicyID = NativeDenom
	}
	s.RequiredSignerBitmap = strings.TrimSpace(s.RequiredSignerBitmap)
	s.SignatureScheme = strings.TrimSpace(s.SignatureScheme)
	if s.SignatureScheme == "" {
		s.SignatureScheme = SignatureSchemeEd25519
	}
	s.SignaturePreimageHash = normalizeOptionalHash(s.SignaturePreimageHash)
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
	if s.AccruedFees == "" {
		s.AccruedFees = "0"
	}
	if s.ParticipantSetHash == "" {
		s.ParticipantSetHash = ComputeParticipantSetHash(participantsFromBalances(s.Balances))
	}
	if s.RequiredSignerBitmap == "" {
		participants := participantsFromBalances(s.Balances)
		s.RequiredSignerBitmap = ComputeRequiredSignerBitmap(participants, participants)
	}
	if s.SignaturePreimageHash == "" {
		s.SignaturePreimageHash = ComputeStateSignaturePreimageHash(s)
	}
	s.Signatures = normalizeSignatures(s.Signatures)
	return s
}

func (s ClaimSignature) Normalize() ClaimSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = strings.TrimSpace(s.ObjectID)
	s.CommitmentHash = normalizeHash(s.CommitmentHash)
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
	if s.ObjectType != SignatureObjectClaim {
		return errors.New("payments claim signature object type mismatch")
	}
	if s.Version != CurrentStateVersion {
		return errors.New("payments claim signature version mismatch")
	}
	if s.ObjectID != s.ClaimHash {
		return errors.New("payments claim signature object id mismatch")
	}
	if s.CommitmentHash != s.ClaimHash {
		return errors.New("payments claim signature commitment mismatch")
	}
	if err := ValidateHash("payments claim signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeSignatureEnvelopeHash(s.Signer, s.ChainID, s.ChannelID, s.ObjectType, s.Version, s.Nonce, s.ObjectID, s.ExpirationHeight, s.CommitmentHash); s.SignatureHash != expected {
		return errors.New("payments claim signature value mismatch")
	}
	return nil
}

func (t ClosedChannelTombstone) Normalize() ClosedChannelTombstone {
	t.ChainID = strings.TrimSpace(t.ChainID)
	t.ChannelID = normalizeHash(t.ChannelID)
	t.StateHash = normalizeHash(t.StateHash)
	return t
}

func (t ClosedChannelTombstone) Validate() error {
	t = t.Normalize()
	if strings.TrimSpace(t.ChainID) == "" {
		return errors.New("payments tombstone chain id is required")
	}
	if err := ValidateHash("payments tombstone channel id", t.ChannelID); err != nil {
		return err
	}
	if t.FinalizedNonce == 0 {
		return errors.New("payments tombstone finalized nonce must be positive")
	}
	if err := ValidateHash("payments tombstone state hash", t.StateHash); err != nil {
		return err
	}
	if t.ClosedHeight == 0 {
		return errors.New("payments tombstone closed height must be positive")
	}
	if t.ExpiresHeight <= t.ClosedHeight {
		return errors.New("payments tombstone replay horizon must exceed close height")
	}
	return nil
}

func (r ConditionClaimRecord) Normalize() ConditionClaimRecord {
	r.ChainID = strings.TrimSpace(r.ChainID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ConditionID = normalizeHash(r.ConditionID)
	r.EvidenceHash = normalizeHash(r.EvidenceHash)
	r.PreimageHash = normalizeOptionalHash(r.PreimageHash)
	return r
}

func (r ConditionClaimRecord) Validate() error {
	r = r.Normalize()
	if strings.TrimSpace(r.ChainID) == "" {
		return errors.New("payments condition claim chain id is required")
	}
	if err := ValidateHash("payments condition claim channel id", r.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments condition claim condition id", r.ConditionID); err != nil {
		return err
	}
	if err := ValidateHash("payments condition claim evidence hash", r.EvidenceHash); err != nil {
		return err
	}
	if r.PreimageHash != "" {
		if err := ValidateHash("payments condition claim preimage hash", r.PreimageHash); err != nil {
			return err
		}
	}
	if r.ResolvedHeight == 0 {
		return errors.New("payments condition claim resolved height must be positive")
	}
	if r.ExpiresHeight <= r.ResolvedHeight {
		return errors.New("payments condition claim replay horizon must exceed resolution height")
	}
	return nil
}

func (r PreimageRevealRequest) Normalize() PreimageRevealRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Promises = normalizeConditionalPromises(r.Promises)
	r.Preimage = strings.TrimSpace(r.Preimage)
	r.Revealer = strings.TrimSpace(r.Revealer)
	return r
}

func (r PreimageRevealRequest) ValidateForChannel(channel ChannelRecord, settledClaims []ConditionClaimRecord) error {
	req := r.Normalize()
	channel = channel.Normalize()
	if req.ChannelID != channel.ChannelID {
		return errors.New("payments preimage reveal channel mismatch")
	}
	if err := addressing.ValidateUserAddress("payments preimage revealer", req.Revealer); err != nil {
		return err
	}
	if !containsString(channel.Participants, req.Revealer) {
		return errors.New("payments preimage revealer must be channel participant")
	}
	if req.CurrentHeight == 0 {
		return errors.New("payments preimage reveal height must be positive")
	}
	if req.Preimage == "" {
		return errors.New("payments preimage reveal preimage is required")
	}
	if len(req.Promises) == 0 {
		return errors.New("payments preimage reveal requires promises")
	}
	var hashLock string
	seen := make(map[string]struct{}, len(req.Promises))
	for _, promise := range req.Promises {
		if err := promise.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[promise.PromiseID]; found {
			return errors.New("payments duplicate promise id")
		}
		seen[promise.PromiseID] = struct{}{}
		if promise.ConditionType != ConditionTypeHashLock {
			return errors.New("payments preimage reveal requires hash-lock promises")
		}
		if req.CurrentHeight > promise.TimeoutHeight {
			return errors.New("payments preimage reveal promise has timed out")
		}
		if err := VerifyPromisePreimage(promise, req.Preimage); err != nil {
			return err
		}
		if promiseWasSettled(channel, promise.PromiseID, settledClaims) {
			return errors.New("payments promise has already been settled")
		}
		if hashLock == "" {
			hashLock = promise.HashLock
		} else if promise.HashLock != hashLock {
			return errors.New("payments linked promises must use compatible hash locks")
		}
	}
	return nil
}

func (r PromiseExpiryRequest) Normalize() PromiseExpiryRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Promises = normalizeConditionalPromises(r.Promises)
	r.Resolver = strings.TrimSpace(r.Resolver)
	return r
}

func (r PromiseExpiryRequest) ValidateForChannel(channel ChannelRecord, settledClaims []ConditionClaimRecord) error {
	req := r.Normalize()
	channel = channel.Normalize()
	if req.ChannelID != channel.ChannelID {
		return errors.New("payments promise expiry channel mismatch")
	}
	if err := addressing.ValidateUserAddress("payments promise expiry resolver", req.Resolver); err != nil {
		return err
	}
	if !containsString(channel.Participants, req.Resolver) {
		return errors.New("payments promise expiry resolver must be channel participant")
	}
	if req.CurrentHeight == 0 {
		return errors.New("payments promise expiry height must be positive")
	}
	if len(req.Promises) == 0 {
		return errors.New("payments promise expiry requires promises")
	}
	seen := make(map[string]struct{}, len(req.Promises))
	for _, promise := range req.Promises {
		if err := promise.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[promise.PromiseID]; found {
			return errors.New("payments duplicate promise id")
		}
		seen[promise.PromiseID] = struct{}{}
		if req.CurrentHeight <= promise.TimeoutHeight {
			return errors.New("payments promise has not expired")
		}
		if promiseWasSettled(channel, promise.PromiseID, settledClaims) {
			return errors.New("payments promise has already been settled")
		}
	}
	return nil
}

func (s DeltaSignature) Normalize() DeltaSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = strings.TrimSpace(s.ObjectID)
	s.CommitmentHash = normalizeHash(s.CommitmentHash)
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
	if s.ObjectType != SignatureObjectDelta {
		return errors.New("payments async delta signature object type mismatch")
	}
	if s.Version != CurrentStateVersion {
		return errors.New("payments async delta signature version mismatch")
	}
	if s.ObjectID == "" {
		return errors.New("payments async delta signature object id is required")
	}
	if s.CommitmentHash != s.DeltaHash {
		return errors.New("payments async delta signature commitment mismatch")
	}
	if err := ValidateHash("payments async delta signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeSignatureEnvelopeHash(s.Signer, s.ChainID, s.ChannelID, s.ObjectType, s.Version, s.Nonce, s.ObjectID, s.ExpirationHeight, s.CommitmentHash); s.SignatureHash != expected {
		return errors.New("payments async delta signature value mismatch")
	}
	return nil
}

func (d AsyncPaymentDelta) Normalize() AsyncPaymentDelta {
	d.UpdateID = normalizeHash(d.UpdateID)
	d.ChainID = strings.TrimSpace(d.ChainID)
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
	if delta.ChainID != channel.ChainID {
		return errors.New("payments async delta chain id mismatch")
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
	if err := validateDeltaSignatureEnvelope(delta.Signature, delta); err != nil {
		return err
	}
	if currentHeight > delta.Signature.ExpirationHeight {
		return errors.New("payments async delta signature is expired")
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
	if proof.EvidenceHash != HashParts("async-dispute", proof.CheckpointState.StateHash, ComputeAsyncDeltaRootForChannel(channel, proof.Deltas)) {
		return errors.New("payments async dispute evidence hash mismatch")
	}
	return nil
}

func (c UnidirectionalClaim) Normalize() UnidirectionalClaim {
	c.ChainID = strings.TrimSpace(c.ChainID)
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
	if claim.ChainID != channel.ChainID {
		return errors.New("payments claim chain id mismatch")
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
	if err := validateClaimSignatureEnvelope(claim.PayerSignature, claim); err != nil {
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
	if err := validateClaimSignatureEnvelope(claim.ReceiverAckOptional, claim); err != nil {
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
	if expected := ComputeParticipantSetHash(channel.Participants); state.ParticipantSetHash != expected {
		return errors.New("payments channel state participant set hash mismatch")
	}
	if state.Denom != channel.Denom {
		return errors.New("payments channel state denom mismatch")
	}
	if state.ChallengePeriod != channel.DisputePeriod {
		return errors.New("payments channel state challenge period mismatch")
	}
	if expected := ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners); state.RequiredSignerBitmap != expected {
		return errors.New("payments channel state required signer bitmap mismatch")
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
	return validateSignatureQuorum(state.Signatures, required, state)
}

func ValidatePreviousHashContinuity(channel ChannelRecord, nextState ChannelState) error {
	channel = channel.Normalize()
	nextState = nextState.Normalize()
	if channel.ChannelType == ChannelTypeAsync {
		return nil
	}
	if nextState.Nonce <= 1 {
		return nil
	}
	if nextState.PreviousStateHash != channel.LatestState.StateHash {
		return errors.New("payments channel state previous hash must match latest state")
	}
	return nil
}

func (s StateSignature) Normalize() StateSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = strings.TrimSpace(s.ObjectID)
	s.CommitmentHash = normalizeHash(s.CommitmentHash)
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
	if s.ObjectType != SignatureObjectState {
		return errors.New("payments signature object type mismatch")
	}
	if s.ObjectID != s.StateHash {
		return errors.New("payments signature object id mismatch")
	}
	if s.CommitmentHash != s.StateHash {
		return errors.New("payments signature commitment mismatch")
	}
	if err := ValidateHash("payments signature hash", s.SignatureHash); err != nil {
		return err
	}
	if expected := ComputeSignatureEnvelopeHash(s.Signer, s.ChainID, s.ChannelID, s.ObjectType, s.Version, s.Nonce, s.ObjectID, s.ExpirationHeight, s.CommitmentHash); s.SignatureHash != expected {
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
	if c.Finality == "" {
		c.Finality = DerivedChannelFinality(c)
	}
	return c
}

func DerivedChannelFinality(channel ChannelRecord) ChannelFinality {
	channel.Status = ChannelStatus(strings.TrimSpace(string(channel.Status)))
	channel.PendingClose = channel.PendingClose.Normalize()
	switch channel.Status {
	case ChannelStatusSettled:
		if len(channel.PendingClose.Penalties) > 0 || len(channel.PendingClose.PenaltyAllocations) > 0 {
			return ChannelFinalityPenalized
		}
		return ChannelFinalitySettled
	case ChannelStatusPendingClose:
		if len(channel.PendingClose.Penalties) > 0 || len(channel.PendingClose.PenaltyAllocations) > 0 {
			return ChannelFinalityPenalized
		}
		if len(channel.PendingClose.ConditionProofs) > 0 || len(channel.PendingClose.State.Conditions) > 0 {
			return ChannelFinalityPendingConditionResolution
		}
		if channel.PendingClose.DisputeCount > 0 {
			return ChannelFinalityInDispute
		}
		if channel.PendingClose.CloseReason == CloseReasonTimeout {
			return ChannelFinalityExpired
		}
		return ChannelFinalityPendingClose
	default:
		return ChannelFinalityOpen
	}
}

func FinalityAfterPendingClose(channel ChannelRecord, currentHeight uint64) ChannelFinality {
	channel = channel.Normalize()
	if channel.Status != ChannelStatusPendingClose {
		return channel.Finality
	}
	if channel.Finality == ChannelFinalityPenalized || channel.Finality == ChannelFinalityExpired {
		return channel.Finality
	}
	if currentHeight >= channel.PendingClose.SettleAfterHeight {
		return ChannelFinalityFinalizable
	}
	return channel.Finality
}

func PendingFinalizationHeightForChannel(channel ChannelRecord) (uint64, bool) {
	channel = channel.Normalize()
	if channel.Status != ChannelStatusPendingClose || channel.PendingClose.SettleAfterHeight == 0 {
		return 0, false
	}
	return channel.PendingClose.SettleAfterHeight, true
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
	if !IsChannelFinality(c.Finality) {
		return fmt.Errorf("unknown payments channel finality %q", c.Finality)
	}
	if err := validateChannelFinalityForStatus(c); err != nil {
		return err
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
		if channel.DisputedNonce < channel.PendingClose.State.Nonce {
			return errors.New("payments disputed nonce cannot be below pending close nonce")
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
	if p.CloseReason == "" {
		p.CloseReason = CloseReasonUnilateral
	}
	p.SettlementFeeDenom = normalizeAssetDenom(p.SettlementFeeDenom)
	p.SettlementFee = strings.TrimSpace(p.SettlementFee)
	p.State = p.State.Normalize()
	p.FraudProofs = normalizeFraudProofs(p.FraudProofs)
	p.ConditionProofs = normalizeConditionResolutions(p.ConditionProofs)
	p.Penalties = normalizePenalties(p.Penalties)
	p.PenaltyAllocations = normalizePenaltyAllocations(p.PenaltyAllocations)
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
	if err := validateCloseReason(p.CloseReason); err != nil {
		return err
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
	if err := validateConditionResolutionsForState(p.State, channel, p.ConditionProofs, false); err != nil {
		return err
	}
	for _, penalty := range p.Penalties {
		if err := penalty.ValidateForChannel(channel); err != nil {
			return err
		}
	}
	for _, allocation := range p.PenaltyAllocations {
		if err := allocation.ValidateForChannel(channel); err != nil {
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
	f.AsyncProof = f.AsyncProof.Normalize()
	return f
}

func (f FraudProof) ChannelID() string {
	proof := f.Normalize()
	if proof.StateA.ChannelID != "" {
		return proof.StateA.ChannelID
	}
	return proof.StateB.ChannelID
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
	case FraudProofTypeInvalidClose:
		if err := proof.StateA.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if proof.StateA.StateHash == channel.LatestState.StateHash {
			return errors.New("payments invalid-close proof requires non-latest close state")
		}
	case FraudProofTypeInvalidBalance:
		if err := validateSignedStateDomainForFraud(channel, proof.StateA); err != nil {
			return err
		}
		if err := validateCollateralConservation(proof.StateA, channel); err == nil {
			return errors.New("payments invalid-balance proof requires collateral conservation failure")
		}
	case FraudProofTypeInvalidCondition:
		if err := proof.StateA.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if len(proof.StateA.Conditions) == 0 {
			return errors.New("payments invalid-condition proof requires conditions")
		}
	case FraudProofTypeReplayAttempt:
		if proof.StateA.ChainID != channel.ChainID || proof.StateA.ChannelID != channel.ChannelID {
			if err := validateSignedStateShapeForFraud(channel, proof.StateA); err != nil {
				return err
			}
			return nil
		}
		if err := proof.StateA.ValidateForChannel(channel, false); err != nil {
			return err
		}
		if proof.StateA.Nonce > channel.FinalizedNonce {
			return errors.New("payments replay proof requires finalized or older nonce")
		}
	case FraudProofTypeAsyncOverexposure:
		if err := validateAsyncOverexposureProof(channel, proof.AsyncProof); err != nil {
			return err
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

func (p PenaltyAllocation) Normalize() PenaltyAllocation {
	p.Offender = strings.TrimSpace(p.Offender)
	p.Denom = normalizeAssetDenom(p.Denom)
	p.Amount = strings.TrimSpace(p.Amount)
	return p
}

func (p PenaltyAllocation) ValidateForChannel(channel ChannelRecord) error {
	allocation := p.Normalize()
	if err := addressing.ValidateUserAddress("payments penalty allocation offender", allocation.Offender); err != nil {
		return err
	}
	if !containsString(channel.Participants, allocation.Offender) {
		return errors.New("payments penalty allocation offender must be channel participant")
	}
	if !IsPenaltyRoute(allocation.Route) || allocation.Route == PenaltyRouteReporter {
		return errors.New("payments penalty allocation route must be burn, security reserve, or community pool")
	}
	if allocation.Denom != NativeDenom {
		return fmt.Errorf("payments penalty allocation denom must be %s", NativeDenom)
	}
	return validatePositiveInt("payments penalty allocation amount", allocation.Amount)
}

func (p FraudPenaltyPolicy) Normalize() FraudPenaltyPolicy {
	p.ReporterRewardCap = strings.TrimSpace(p.ReporterRewardCap)
	return p
}

func (p FraudPenaltyPolicy) Validate() error {
	p = p.Normalize()
	if p.ReporterRewardCap != "" {
		if err := validateNonNegativeInt("payments reporter reward cap", p.ReporterRewardCap); err != nil {
			return err
		}
	}
	total := p.BurnShareBps + p.SecurityReserveShareBps + p.CommunityPoolShareBps
	if total > MaxPenaltyRouteBps {
		return errors.New("payments penalty route shares exceed 10000 bps")
	}
	return nil
}

func BuildFraudPenaltyRouting(channel ChannelRecord, proof FraudProof, policy FraudPenaltyPolicy) ([]Penalty, []PenaltyAllocation, error) {
	channel = channel.Normalize()
	proof = proof.Normalize()
	policy = policy.Normalize()
	if err := proof.ValidateForChannel(channel); err != nil {
		return nil, nil, err
	}
	if err := policy.Validate(); err != nil {
		return nil, nil, err
	}
	penaltyAmount, err := parsePositiveInt("payments fraud penalty", proof.PenaltyAmount)
	if err != nil {
		return nil, nil, err
	}
	reporterReward := penaltyAmount
	if policy.ReporterRewardCap != "" {
		capAmount, err := parseNonNegativeInt("payments reporter reward cap", policy.ReporterRewardCap)
		if err != nil {
			return nil, nil, err
		}
		if reporterReward.GT(capAmount) {
			reporterReward = capAmount
		}
	}
	penalties := []Penalty{}
	if reporterReward.IsPositive() {
		penalty := Penalty{Offender: proof.OffendingSigner, Recipient: proof.SubmittedBy, Denom: NativeDenom, Amount: reporterReward.String()}.Normalize()
		if err := penalty.ValidateForChannel(channel); err != nil {
			return nil, nil, err
		}
		penalties = append(penalties, penalty)
	}
	remaining := penaltyAmount.Sub(reporterReward)
	allocations, err := splitPenaltyRemainder(proof.OffendingSigner, remaining, policy)
	if err != nil {
		return nil, nil, err
	}
	for _, allocation := range allocations {
		if err := allocation.ValidateForChannel(channel); err != nil {
			return nil, nil, err
		}
	}
	return normalizePenalties(penalties), normalizePenaltyAllocations(allocations), nil
}

func (s SettlementRecord) Normalize() SettlementRecord {
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.StateHash = normalizeHash(s.StateHash)
	s.SettlementFeeDenom = normalizeAssetDenom(s.SettlementFeeDenom)
	s.SettlementFee = strings.TrimSpace(s.SettlementFee)
	s.SettlementHash = normalizeOptionalHash(s.SettlementHash)
	s.FinalBalances = normalizeBalances(s.FinalBalances)
	s.Penalties = normalizePenalties(s.Penalties)
	s.PenaltyAllocations = normalizePenaltyAllocations(s.PenaltyAllocations)
	return s
}

func (s SettlementRecord) ValidateForChannel(channel ChannelRecord) error {
	settlement := s.Normalize()
	channel = channel.Normalize()
	if settlement.ChainID != channel.ChainID {
		return errors.New("payments settlement chain id mismatch")
	}
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
	for _, allocation := range settlement.PenaltyAllocations {
		if err := allocation.ValidateForChannel(channel); err != nil {
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
	allocationTotal, err := sumPenaltyAllocations(settlement.PenaltyAllocations)
	if err != nil {
		return err
	}
	if !finalTotal.Add(fee).Add(allocationTotal).Equal(collateral) {
		return errors.New("payments settlement must conserve collateral minus fee and routed penalties")
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

func ChannelDisputeEvent(channel ChannelRecord, submitter string, height uint64) PaymentEvent {
	channel = channel.Normalize()
	event := PaymentEvent{
		EventID:   HashParts("channel-dispute", channel.ChannelID, channel.PendingClose.State.StateHash, fmt.Sprintf("%d", height)),
		EventType: "channel-dispute",
		ChannelID: channel.ChannelID,
		Height:    height,
		Attributes: []PaymentEventAttribute{
			{Key: "submitter", Value: strings.TrimSpace(submitter)},
			{Key: "state_hash", Value: channel.PendingClose.State.StateHash},
			{Key: "nonce", Value: fmt.Sprintf("%d", channel.PendingClose.State.Nonce)},
			{Key: "settle_after_height", Value: fmt.Sprintf("%d", channel.PendingClose.SettleAfterHeight)},
		},
	}
	return event.Normalize()
}

func ChannelFinalityTransitionEvent(channel ChannelRecord, previous, next ChannelFinality, height uint64) PaymentEvent {
	channel = channel.Normalize()
	attrs := []PaymentEventAttribute{
		{Key: "from_finality", Value: string(previous)},
		{Key: "to_finality", Value: string(next)},
		{Key: "status", Value: string(channel.Status)},
	}
	if pendingHeight, ok := PendingFinalizationHeightForChannel(channel); ok {
		attrs = append(attrs, PaymentEventAttribute{Key: "pending_finalization_height", Value: fmt.Sprintf("%d", pendingHeight)})
	}
	event := PaymentEvent{
		EventID:    HashParts("channel-finality", channel.ChannelID, string(previous), string(next), fmt.Sprintf("%d", height)),
		EventType:  "channel-finality-transition",
		ChannelID:  channel.ChannelID,
		Height:     height,
		Attributes: attrs,
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
	v.ChainID = strings.TrimSpace(v.ChainID)
	for i := range v.ParentChannelIDs {
		v.ParentChannelIDs[i] = normalizeHash(v.ParentChannelIDs[i])
	}
	v.Endpoints = normalizeAddressSet(v.Endpoints)
	v.Capacity = strings.TrimSpace(v.Capacity)
	if v.Nonce == 0 {
		v.Nonce = 1
	}
	v.AnchorCommitment = normalizeOptionalHash(v.AnchorCommitment)
	v.StateHash = normalizeOptionalHash(v.StateHash)
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
	if strings.TrimSpace(vc.ChainID) == "" {
		return errors.New("payments virtual channel chain id is required")
	}
	if vc.Nonce == 0 {
		return errors.New("payments virtual channel nonce must be positive")
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
	if err := ValidateHash("payments virtual channel state hash", vc.StateHash); err != nil {
		return err
	}
	if expected := ComputeVirtualChannelStateHash(vc); vc.StateHash != expected {
		return errors.New("payments virtual channel state hash mismatch")
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

func IsChannelFinality(value ChannelFinality) bool {
	switch value {
	case ChannelFinalityOpen,
		ChannelFinalityPendingClose,
		ChannelFinalityInDispute,
		ChannelFinalityPendingConditionResolution,
		ChannelFinalityFinalizable,
		ChannelFinalitySettled,
		ChannelFinalityPenalized,
		ChannelFinalityExpired:
		return true
	default:
		return false
	}
}

func IsFraudProofType(value FraudProofType) bool {
	switch value {
	case FraudProofTypeDoubleSign,
		FraudProofTypeStaleClose,
		FraudProofTypeInvalidClose,
		FraudProofTypeInvalidBalance,
		FraudProofTypeInvalidCondition,
		FraudProofTypeReplayAttempt,
		FraudProofTypeAsyncOverexposure:
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

func IsCloseReason(value CloseReason) bool {
	switch value {
	case CloseReasonUnilateral, CloseReasonCooperative, CloseReasonTimeout, CloseReasonFraud:
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

func IsPenaltyRoute(value PenaltyRoute) bool {
	switch value {
	case PenaltyRouteReporter, PenaltyRouteBurn, PenaltyRouteSecurityReserve, PenaltyRouteCommunityPool:
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
	if state.AppVersion != CurrentAppVersion {
		return fmt.Errorf("payments channel state app version must be %d", CurrentAppVersion)
	}
	if state.ModuleName != ModuleName {
		return fmt.Errorf("payments channel state module name must be %s", ModuleName)
	}
	if err := validateRequiredFields(state.RequiredFields); err != nil {
		return err
	}
	if err := ValidateHash("payments channel state channel id", state.ChannelID); err != nil {
		return err
	}
	if !IsChannelType(state.ChannelType) {
		return fmt.Errorf("unknown payments channel state type %q", state.ChannelType)
	}
	if err := ValidateHash("payments channel state participant set hash", state.ParticipantSetHash); err != nil {
		return err
	}
	if state.Denom != NativeDenom {
		return fmt.Errorf("payments channel state denom must be %s", NativeDenom)
	}
	if state.Version != CurrentStateVersion {
		return fmt.Errorf("payments channel state version must be %d", CurrentStateVersion)
	}
	if err := validateNonNegativeInt("payments channel state accrued fees", state.AccruedFees); err != nil {
		return err
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
	if err := ValidateHash("payments condition root", state.ConditionRoot); err != nil {
		return err
	}
	if state.ConditionRoot != state.PendingConditionsRoot {
		return errors.New("payments condition root must match pending conditions root")
	}
	if expected := ComputeConditionsRoot(state.Conditions); state.ConditionRoot != expected {
		return errors.New("payments pending conditions root mismatch")
	}
	if state.ConditionCount != uint32(len(state.Conditions)) {
		return errors.New("payments condition count mismatch")
	}
	if state.TimeoutTimestamp < 0 {
		return errors.New("payments channel state timeout timestamp must be non-negative")
	}
	if state.ChallengePeriod == 0 {
		return errors.New("payments channel state challenge period must be positive")
	}
	if err := validateChallengePeriod(state.ChallengePeriod); err != nil {
		return err
	}
	if state.FeePolicyID != NativeDenom {
		return fmt.Errorf("payments channel state fee policy must be %s", NativeDenom)
	}
	if err := validateRequiredSignerBitmap(state.RequiredSignerBitmap); err != nil {
		return err
	}
	if state.SignatureScheme != SignatureSchemeEd25519 {
		return fmt.Errorf("payments channel state signature scheme must be %s", SignatureSchemeEd25519)
	}
	if err := ValidateHash("payments channel state signature preimage hash", state.SignaturePreimageHash); err != nil {
		return err
	}
	if expected := ComputeStateSignaturePreimageHash(state); state.SignaturePreimageHash != expected {
		return errors.New("payments channel state signature preimage hash mismatch")
	}
	if err := validateBalances(state.Balances); err != nil {
		return err
	}
	return validateConditions(state.Conditions)
}

func openingStateForRequest(req ChannelOpenRequest, channel ChannelRecord) ChannelState {
	channel = channel.Normalize()
	state := ChannelState{
		ChainID:              req.ChainID,
		AppVersion:           CurrentAppVersion,
		ModuleName:           ModuleName,
		ChannelID:            req.ChannelID,
		ChannelType:          req.ChannelType,
		ParticipantSetHash:   ComputeParticipantSetHash(channel.Participants),
		Denom:                NativeDenom,
		Version:              CurrentStateVersion,
		Epoch:                1,
		Nonce:                1,
		Balances:             req.InitialBalances,
		TimeoutHeight:        req.OpenHeight + req.ChallengePeriod,
		ChallengePeriod:      req.ChallengePeriod,
		CloseDelay:           req.CloseDelay,
		FeePolicyID:          req.FeePolicyID,
		RequiredSignerBitmap: ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners),
		SignatureScheme:      SignatureSchemeEd25519,
	}
	if req.ChannelType == ChannelTypeUnidirectional {
		state.TimeoutHeight = req.ExpirationHeight
		state.TimeoutTimestamp = req.ExpirationTimestamp
	}
	if req.ChannelType == ChannelTypeAsync {
		state.CheckpointNonce = 1
		state.CheckpointBalances = req.InitialBalances
		state.AsyncUpdateRoot = ComputeAsyncDeltaRootForChannel(channel, nil)
		state.AcceptedUpdateRoot = ComputeAsyncDeltaRootForChannel(channel, nil)
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

func validateUpdateExposure(state ChannelState) error {
	state = state.Normalize()
	if state.ChannelType != ChannelTypeBidirectional || len(state.Conditions) == 0 {
		return nil
	}
	conditionTotal, err := sumConditions(state.Conditions)
	if err != nil {
		return err
	}
	reserveA, err := parseNonNegativeInt("payments update reserve a", state.ReserveA)
	if err != nil {
		return err
	}
	reserveB, err := parseNonNegativeInt("payments update reserve b", state.ReserveB)
	if err != nil {
		return err
	}
	if conditionTotal.GT(reserveA.Add(reserveB)) {
		return errors.New("payments update conditions exceed reserve limits")
	}
	return nil
}

func validateConditionResolutionsForState(state ChannelState, channel ChannelRecord, resolutions []ConditionResolution, requireAll bool) error {
	state = state.Normalize()
	if len(state.Conditions) == 0 {
		if len(resolutions) > 0 {
			return errors.New("payments condition proofs supplied for state without conditions")
		}
		return nil
	}
	conditionByID := make(map[string]ConditionalPayment, len(state.Conditions))
	for _, condition := range state.Conditions {
		condition = condition.Normalize()
		conditionByID[condition.ConditionID] = condition
	}
	seen := make(map[string]struct{}, len(resolutions))
	for _, resolution := range normalizeConditionResolutions(resolutions) {
		condition, found := conditionByID[resolution.ConditionID]
		if !found {
			return errors.New("payments condition proof references unknown condition")
		}
		if _, found := seen[resolution.ConditionID]; found {
			return errors.New("payments duplicate condition proof")
		}
		if err := resolution.ValidateForCondition(condition, channel); err != nil {
			return err
		}
		seen[resolution.ConditionID] = struct{}{}
	}
	if requireAll && len(seen) != len(conditionByID) {
		return errors.New("payments all conditions must be resolved or expired")
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
	if strings.TrimSpace(claim.ChainID) == "" {
		return errors.New("payments claim chain id is required")
	}
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
	if strings.TrimSpace(delta.ChainID) == "" {
		return errors.New("payments async delta chain id is required")
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

func validateCloseReason(reason CloseReason) error {
	if !IsCloseReason(reason) {
		return fmt.Errorf("unknown payments close reason %q", reason)
	}
	return nil
}

func validateChannelFinalityForStatus(channel ChannelRecord) error {
	switch channel.Status {
	case ChannelStatusOpen:
		if channel.Finality != ChannelFinalityOpen && channel.Finality != ChannelFinalityExpired {
			return errors.New("payments open channel finality must be open or expired")
		}
	case ChannelStatusPendingClose:
		switch channel.Finality {
		case ChannelFinalityPendingClose,
			ChannelFinalityInDispute,
			ChannelFinalityPendingConditionResolution,
			ChannelFinalityFinalizable,
			ChannelFinalityPenalized,
			ChannelFinalityExpired:
			return nil
		default:
			return errors.New("payments pending close channel has invalid finality")
		}
	case ChannelStatusSettled:
		if channel.Finality != ChannelFinalitySettled && channel.Finality != ChannelFinalityPenalized {
			return errors.New("payments settled channel finality must be settled or penalized")
		}
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

func validateSignatureQuorum(signatures []StateSignature, required []string, state ChannelState) error {
	if err := validateAddressSet("payments required signer", required, 1, MaxParticipants); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(signatures))
	for i, sig := range signatures {
		sig = sig.Normalize()
		if sig.ChainID != state.ChainID {
			return errors.New("payments signature chain id mismatch")
		}
		if sig.ChannelID != state.ChannelID {
			return errors.New("payments signature channel id mismatch")
		}
		if sig.Version != state.Version {
			return errors.New("payments signature version mismatch")
		}
		if sig.Nonce != state.Nonce {
			return errors.New("payments signature nonce mismatch")
		}
		if sig.ExpirationHeight != state.TimeoutHeight {
			return errors.New("payments signature expiration height mismatch")
		}
		if sig.CommitmentHash != state.StateHash {
			return errors.New("payments signature commitment mismatch")
		}
		if sig.ObjectID != state.StateHash {
			return errors.New("payments signature object id mismatch")
		}
		if err := sig.Validate(state.StateHash); err != nil {
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

func validateSignedStateShapeForFraud(channel ChannelRecord, state ChannelState) error {
	channel = channel.Normalize()
	state = state.Normalize()
	if err := validateUnsignedStateShape(state); err != nil {
		return err
	}
	if state.StateHash == "" {
		return errors.New("payments fraud state hash is required")
	}
	if expected := ComputeStateHash(state); state.StateHash != expected {
		return errors.New("payments fraud state hash mismatch")
	}
	return validateSignatureQuorum(state.Signatures, channel.RequiredSigners, state)
}

func validateSignedStateDomainForFraud(channel ChannelRecord, state ChannelState) error {
	channel = channel.Normalize()
	state = state.Normalize()
	if err := validateSignedStateShapeForFraud(channel, state); err != nil {
		return err
	}
	if state.ChainID != channel.ChainID {
		return errors.New("payments fraud state chain id mismatch")
	}
	if state.ChannelID != channel.ChannelID {
		return errors.New("payments fraud state channel mismatch")
	}
	if state.ChannelType != channel.ChannelType {
		return errors.New("payments fraud state type mismatch")
	}
	if expected := ComputeParticipantSetHash(channel.Participants); state.ParticipantSetHash != expected {
		return errors.New("payments fraud state participant set hash mismatch")
	}
	if state.Denom != channel.Denom {
		return errors.New("payments fraud state denom mismatch")
	}
	if state.ChallengePeriod != channel.DisputePeriod {
		return errors.New("payments fraud state challenge period mismatch")
	}
	if expected := ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners); state.RequiredSignerBitmap != expected {
		return errors.New("payments fraud state required signer bitmap mismatch")
	}
	return validateStateParticipants(state, channel)
}

func validateAsyncOverexposureProof(channel ChannelRecord, asyncProof AsyncDeltaDisputeProof) error {
	channel = channel.Normalize()
	proof := asyncProof.Normalize()
	if channel.ChannelType != ChannelTypeAsync {
		return errors.New("payments async overexposure proof requires async channel")
	}
	if err := ValidateHash("payments async overexposure proof id", proof.ProofID); err != nil {
		return err
	}
	if proof.ChannelID != channel.ChannelID {
		return errors.New("payments async overexposure proof channel mismatch")
	}
	if err := ValidateHash("payments async overexposure evidence hash", proof.EvidenceHash); err != nil {
		return err
	}
	if err := proof.CheckpointState.ValidateForChannel(channel, false); err != nil {
		return err
	}
	if len(proof.Deltas) == 0 {
		return errors.New("payments async overexposure proof requires deltas")
	}
	maxExposure, err := parsePositiveInt("payments async max unacked amount", proof.CheckpointState.MaxUnackedAmount)
	if err != nil {
		return err
	}
	currentHeight := channel.OpenHeight
	if channel.PendingClose.SubmittedHeight != 0 {
		currentHeight = channel.PendingClose.SubmittedHeight
	}
	exposureBySender := make(map[string]sdkmath.Int, len(channel.Participants))
	overexposed := false
	for _, delta := range normalizeAsyncDeltas(proof.Deltas) {
		if err := delta.ValidateForChannel(channel, currentHeight); err != nil {
			return err
		}
		amount, err := parsePositiveInt("payments async delta amount", delta.Amount)
		if err != nil {
			return err
		}
		exposureBySender[delta.From] = exposureBySender[delta.From].Add(amount)
		if exposureBySender[delta.From].GT(maxExposure) {
			overexposed = true
		}
	}
	if !overexposed {
		return errors.New("payments async overexposure proof requires exposure above max")
	}
	if proof.EvidenceHash != HashParts("async-overexposure", proof.CheckpointState.StateHash, ComputeAsyncDeltaRootForChannel(channel, proof.Deltas)) {
		return errors.New("payments async overexposure evidence hash mismatch")
	}
	return nil
}

func validateClaimSignatureEnvelope(sig ClaimSignature, claim UnidirectionalClaim) error {
	sig = sig.Normalize()
	claim = claim.Normalize()
	if sig.ChainID != claim.ChainID {
		return errors.New("payments claim signature chain id mismatch")
	}
	if sig.ChannelID != claim.ChannelID {
		return errors.New("payments claim signature channel id mismatch")
	}
	if sig.Nonce != claim.Nonce {
		return errors.New("payments claim signature nonce mismatch")
	}
	if sig.ExpirationHeight != claim.ExpirationHeight {
		return errors.New("payments claim signature expiration height mismatch")
	}
	if sig.ObjectID != claim.StateHash {
		return errors.New("payments claim signature object id mismatch")
	}
	if sig.CommitmentHash != claim.StateHash {
		return errors.New("payments claim signature commitment mismatch")
	}
	return nil
}

func validateDeltaSignatureEnvelope(sig DeltaSignature, delta AsyncPaymentDelta) error {
	sig = sig.Normalize()
	delta = delta.Normalize()
	if sig.ChainID != delta.ChainID {
		return errors.New("payments async delta signature chain id mismatch")
	}
	if sig.ChannelID != delta.ChannelID {
		return errors.New("payments async delta signature channel id mismatch")
	}
	if sig.Nonce != delta.NonceStart {
		return errors.New("payments async delta signature nonce mismatch")
	}
	if sig.ExpirationHeight != delta.ExpiryHeight {
		return errors.New("payments async delta signature expiration height mismatch")
	}
	if sig.ObjectID != delta.UpdateID {
		return errors.New("payments async delta signature object id mismatch")
	}
	if sig.CommitmentHash != delta.DeltaHash {
		return errors.New("payments async delta signature commitment mismatch")
	}
	return nil
}

func validateRequiredSignerBitmap(bitmap string) error {
	if bitmap == "" {
		return errors.New("payments required signer bitmap is required")
	}
	if len(bitmap) > MaxParticipants {
		return fmt.Errorf("payments required signer bitmap must be <= %d bits", MaxParticipants)
	}
	hasRequired := false
	for _, bit := range bitmap {
		if bit != '0' && bit != '1' {
			return errors.New("payments required signer bitmap must contain only 0 or 1")
		}
		if bit == '1' {
			hasRequired = true
		}
	}
	if !hasRequired {
		return errors.New("payments required signer bitmap must require at least one signer")
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

func ValidateConditionalPromisesForChannel(channel ChannelRecord, promises []ConditionalPromise, settledClaims []ConditionClaimRecord) error {
	channel = channel.Normalize()
	if len(promises) > MaxConditionsPerState {
		return fmt.Errorf("payments promises must be <= %d", MaxConditionsPerState)
	}
	seen := make(map[string]struct{}, len(promises))
	reservedBySource := make(map[string]sdkmath.Int, len(channel.Participants))
	for _, promise := range normalizeConditionalPromises(promises) {
		if err := promise.ValidateForChannel(channel); err != nil {
			return err
		}
		if _, found := seen[promise.PromiseID]; found {
			return errors.New("payments duplicate promise id")
		}
		seen[promise.PromiseID] = struct{}{}
		if promiseWasSettled(channel, promise.PromiseID, settledClaims) {
			return errors.New("payments promise has already been settled")
		}
		amount, err := parsePositiveInt("payments promise amount", promise.Amount)
		if err != nil {
			return err
		}
		fee, err := parseNonNegativeInt("payments promise fee", promise.Fee)
		if err != nil {
			return err
		}
		current, found := reservedBySource[promise.Source]
		if !found {
			current = sdkmath.ZeroInt()
		}
		reservedBySource[promise.Source] = current.Add(amount).Add(fee)
	}
	for source, reserved := range reservedBySource {
		available, err := availablePromiseReserve(channel, source)
		if err != nil {
			return err
		}
		if reserved.GT(available) {
			return errors.New("payments promises exceed available reserve")
		}
	}
	return nil
}

func VerifyPromisePreimage(promise ConditionalPromise, preimage string) error {
	promise = promise.Normalize()
	preimage = strings.TrimSpace(preimage)
	if promise.ConditionType != ConditionTypeHashLock {
		return errors.New("payments preimage verification requires hash-lock promise")
	}
	if preimage == "" {
		return errors.New("payments preimage is required")
	}
	if HashParts(preimage) != promise.HashLock {
		return errors.New("payments preimage does not satisfy hash lock")
	}
	return nil
}

func ValidatePromiseTimeoutOrdering(channel ChannelRecord, upstream, downstream ConditionalPromise, margin uint64) error {
	channel = channel.Normalize()
	upstream = upstream.Normalize()
	downstream = downstream.Normalize()
	if margin == 0 {
		margin = DefaultTimeoutMargin
	}
	minMargin := channel.CloseDelay + channel.DisputePeriod
	if margin < minMargin {
		return errors.New("payments timeout margin must cover dispute and settlement latency")
	}
	if upstream.ChannelID != channel.ChannelID || downstream.ChannelID != channel.ChannelID {
		return errors.New("payments timeout ordering channel mismatch")
	}
	if upstream.HashLock != downstream.HashLock {
		return errors.New("payments timeout ordering requires compatible hash locks")
	}
	if downstream.TimeoutHeight+margin < downstream.TimeoutHeight || downstream.TimeoutHeight+margin > upstream.TimeoutHeight {
		return errors.New("payments downstream timeout must expire before upstream by margin")
	}
	if upstream.PreviousPromiseIDOptional != "" && upstream.PreviousPromiseIDOptional != downstream.PromiseID {
		return errors.New("payments upstream previous promise link mismatch")
	}
	if downstream.NextPromiseIDOptional != "" && downstream.NextPromiseIDOptional != upstream.PromiseID {
		return errors.New("payments downstream next promise link mismatch")
	}
	return nil
}

func ValidatePromiseTimeoutChain(channel ChannelRecord, promises []ConditionalPromise, margin uint64) error {
	byID := make(map[string]ConditionalPromise, len(promises))
	for _, promise := range normalizeConditionalPromises(promises) {
		byID[promise.PromiseID] = promise
	}
	for _, downstream := range byID {
		if downstream.NextPromiseIDOptional == "" {
			continue
		}
		upstream, found := byID[downstream.NextPromiseIDOptional]
		if !found {
			return errors.New("payments timeout chain references unknown upstream promise")
		}
		if err := ValidatePromiseTimeoutOrdering(channel, upstream, downstream, margin); err != nil {
			return err
		}
	}
	return nil
}

func BuildConditionRootUpdateFromPromises(channel ChannelRecord, base ChannelState, promises []ConditionalPromise, settledClaims []ConditionClaimRecord) (ChannelState, ConditionRootUpdate, error) {
	channel = channel.Normalize()
	base = base.Normalize()
	if err := base.ValidateForChannel(channel, false); err != nil {
		return ChannelState{}, ConditionRootUpdate{}, err
	}
	if err := ValidateConditionalPromisesForChannel(channel, promises, settledClaims); err != nil {
		return ChannelState{}, ConditionRootUpdate{}, err
	}
	conditions := make([]ConditionalPayment, 0, len(promises))
	for _, promise := range normalizeConditionalPromises(promises) {
		conditions = append(conditions, promise.ToConditionalPayment())
	}
	if err := validateConditions(conditions); err != nil {
		return ChannelState{}, ConditionRootUpdate{}, err
	}
	next := base
	next.Conditions = conditions
	next.ConditionRoot = ComputeConditionsRoot(conditions)
	next.PendingConditionsRoot = next.ConditionRoot
	next.ConditionCount = uint32(len(conditions))
	next.SignaturePreimageHash = ComputeStateSignaturePreimageHash(next)
	next.StateHash = ComputeStateHash(next)
	return next.Normalize(), ConditionRootUpdate{
		ChannelID:      channel.ChannelID,
		Nonce:          next.Nonce,
		ConditionRoot:  next.ConditionRoot,
		ConditionCount: next.ConditionCount,
		Conditions:     next.Conditions,
	}, nil
}

func BuildConditionRootAfterExpiry(base ChannelState, expired []ConditionalPromise) (ChannelState, ConditionRootUpdate, error) {
	base = base.Normalize()
	expiredByID := make(map[string]struct{}, len(expired))
	for _, promise := range normalizeConditionalPromises(expired) {
		expiredByID[promise.PromiseID] = struct{}{}
	}
	conditions := make([]ConditionalPayment, 0, len(base.Conditions))
	for _, condition := range normalizeConditions(base.Conditions) {
		if _, found := expiredByID[condition.ConditionID]; found {
			continue
		}
		conditions = append(conditions, condition)
	}
	if len(conditions) == len(base.Conditions) {
		return ChannelState{}, ConditionRootUpdate{}, errors.New("payments expiry did not remove any condition")
	}
	next := base
	next.Conditions = conditions
	next.ConditionRoot = ComputeConditionsRoot(conditions)
	next.PendingConditionsRoot = next.ConditionRoot
	next.ConditionCount = uint32(len(conditions))
	next.SignaturePreimageHash = ComputeStateSignaturePreimageHash(next)
	next.StateHash = ComputeStateHash(next)
	return next.Normalize(), ConditionRootUpdate{
		ChannelID:      next.ChannelID,
		Nonce:          next.Nonce,
		ConditionRoot:  next.ConditionRoot,
		ConditionCount: next.ConditionCount,
		Conditions:     next.Conditions,
	}, nil
}

func validatePromiseTimeoutWindow(channel ChannelRecord, promise ConditionalPromise) error {
	if promise.TimeoutHeight <= channel.OpenHeight {
		return errors.New("payments promise timeout must be after channel open height")
	}
	maxHeight := channel.LatestState.TimeoutHeight
	if maxHeight == 0 {
		maxHeight = channel.OpenHeight + channel.CloseDelay + channel.DisputePeriod
	}
	if promise.TimeoutHeight > maxHeight {
		return errors.New("payments promise timeout exceeds channel timeout")
	}
	if promise.TimeoutHeight+channel.DisputePeriod < promise.TimeoutHeight || promise.TimeoutHeight+channel.DisputePeriod > maxHeight {
		return errors.New("payments promise timeout must fit dispute window")
	}
	return nil
}

func validatePromiseSignatureEnvelope(channel ChannelRecord, sig PromiseSignature, promise ConditionalPromise) error {
	sig = sig.Normalize()
	promise = promise.Normalize()
	if sig.ChainID != channel.ChainID {
		return errors.New("payments promise signature chain id mismatch")
	}
	if sig.ChannelID != promise.ChannelID {
		return errors.New("payments promise signature channel id mismatch")
	}
	if sig.Version != CurrentStateVersion {
		return errors.New("payments promise signature version mismatch")
	}
	if sig.Nonce != promise.Nonce {
		return errors.New("payments promise signature nonce mismatch")
	}
	if sig.ExpirationHeight != promise.TimeoutHeight {
		return errors.New("payments promise signature expiration height mismatch")
	}
	if sig.Signer != promise.Source {
		return errors.New("payments promise signature signer must be source")
	}
	return nil
}

func availablePromiseReserve(channel ChannelRecord, source string) (sdkmath.Int, error) {
	state := channel.Normalize().LatestState.Normalize()
	if channel.ChannelType == ChannelTypeBidirectional {
		if source == state.ParticipantA {
			return parseNonNegativeInt("payments promise reserve a", state.ReserveA)
		}
		if source == state.ParticipantB {
			return parseNonNegativeInt("payments promise reserve b", state.ReserveB)
		}
	}
	return sdkmath.Int{}, errors.New("payments promise source reserve not found")
}

func promiseWasSettled(channel ChannelRecord, promiseID string, settledClaims []ConditionClaimRecord) bool {
	channel = channel.Normalize()
	promiseID = normalizeHash(promiseID)
	for _, claim := range settledClaims {
		claim = claim.Normalize()
		if claim.ChannelID == channel.ChannelID && claim.ConditionID == promiseID {
			return true
		}
	}
	return false
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

func sumPenaltyAllocations(allocations []PenaltyAllocation) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, allocation := range normalizePenaltyAllocations(allocations) {
		amount, err := parsePositiveInt("payments penalty allocation amount", allocation.Amount)
		if err != nil {
			return sdkmath.Int{}, err
		}
		total = total.Add(amount)
	}
	return total, nil
}

func splitPenaltyRemainder(offender string, remaining sdkmath.Int, policy FraudPenaltyPolicy) ([]PenaltyAllocation, error) {
	if !remaining.IsPositive() {
		return nil, nil
	}
	shares := []struct {
		route PenaltyRoute
		bps   uint32
	}{
		{route: PenaltyRouteBurn, bps: policy.BurnShareBps},
		{route: PenaltyRouteSecurityReserve, bps: policy.SecurityReserveShareBps},
		{route: PenaltyRouteCommunityPool, bps: policy.CommunityPoolShareBps},
	}
	allocated := sdkmath.ZeroInt()
	amountByRoute := map[PenaltyRoute]sdkmath.Int{}
	for _, share := range shares {
		if share.bps == 0 {
			continue
		}
		amount := remaining.MulRaw(int64(share.bps)).QuoRaw(int64(MaxPenaltyRouteBps))
		if !amount.IsPositive() {
			continue
		}
		allocated = allocated.Add(amount)
		current, found := amountByRoute[share.route]
		if !found {
			current = sdkmath.ZeroInt()
		}
		amountByRoute[share.route] = current.Add(amount)
	}
	remainder := remaining.Sub(allocated)
	if remainder.IsPositive() {
		route := PenaltyRouteCommunityPool
		if policy.BurnShareBps == 0 && policy.SecurityReserveShareBps == 0 && policy.CommunityPoolShareBps == 0 {
			route = PenaltyRouteCommunityPool
		}
		current, found := amountByRoute[route]
		if !found {
			current = sdkmath.ZeroInt()
		}
		amountByRoute[route] = current.Add(remainder)
	}
	allocations := make([]PenaltyAllocation, 0, len(amountByRoute))
	for route, amount := range amountByRoute {
		if amount.IsPositive() {
			allocations = append(allocations, PenaltyAllocation{Offender: offender, Route: route, Denom: NativeDenom, Amount: amount.String()})
		}
	}
	return normalizePenaltyAllocations(allocations), nil
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

func participantsFromBalances(balances []Balance) []string {
	participants := make([]string, 0, len(balances))
	for _, balance := range normalizeBalances(balances) {
		if balance.Participant != "" {
			participants = append(participants, balance.Participant)
		}
	}
	return normalizeAddressSet(participants)
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

func normalizeConditionalPromises(promises []ConditionalPromise) []ConditionalPromise {
	out := make([]ConditionalPromise, len(promises))
	for i, promise := range promises {
		out[i] = promise.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].PromiseID < out[j].PromiseID
	})
	return out
}

func normalizeConditionResolutions(resolutions []ConditionResolution) []ConditionResolution {
	out := make([]ConditionResolution, len(resolutions))
	for i, resolution := range resolutions {
		out[i] = resolution.Normalize()
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

func normalizeSignedNonceRecords(records []SignedNonceRecord) []SignedNonceRecord {
	out := make([]SignedNonceRecord, len(records))
	for i, record := range records {
		out[i] = record.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Signer != out[j].Signer {
			return out[i].Signer < out[j].Signer
		}
		if out[i].ChainID != out[j].ChainID {
			return out[i].ChainID < out[j].ChainID
		}
		if out[i].ChannelID != out[j].ChannelID {
			return out[i].ChannelID < out[j].ChannelID
		}
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		return out[i].Nonce < out[j].Nonce
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
	seenNonce := make(map[string]struct{}, len(deltas))
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
		for nonce := delta.NonceStart; nonce <= delta.NonceEnd; nonce++ {
			key := fmt.Sprintf("%s/%d", delta.From, nonce)
			if _, found := seenNonce[key]; found {
				return errors.New("payments duplicate async delta nonce")
			}
			seenNonce[key] = struct{}{}
			if nonce == ^uint64(0) {
				break
			}
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

func normalizePenaltyAllocations(allocations []PenaltyAllocation) []PenaltyAllocation {
	out := make([]PenaltyAllocation, len(allocations))
	for i, allocation := range allocations {
		out[i] = allocation.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Offender != out[j].Offender {
			return out[i].Offender < out[j].Offender
		}
		if out[i].Route != out[j].Route {
			return out[i].Route < out[j].Route
		}
		return out[i].Amount < out[j].Amount
	})
	return out
}

func normalizeStateSignatures(signatures []StateSignature) []StateSignature {
	out := make([]StateSignature, len(signatures))
	for i, signature := range signatures {
		out[i] = signature.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Signer < out[j].Signer
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

func normalizeRequiredFields(values []string) []string {
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

func validateRequiredFields(fields []string) error {
	fields = normalizeRequiredFields(fields)
	known := normalizeRequiredFields(CanonicalStateRequiredFields())
	knownSet := make(map[string]struct{}, len(known))
	for _, field := range known {
		knownSet[field] = struct{}{}
	}
	for _, field := range fields {
		if _, found := knownSet[field]; !found {
			return fmt.Errorf("payments channel state unknown required field %q", field)
		}
	}
	if len(fields) != len(known) {
		return errors.New("payments channel state required fields mismatch")
	}
	for i, field := range fields {
		if field != known[i] {
			return fmt.Errorf("payments channel state unknown required field %q", field)
		}
	}
	return nil
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
