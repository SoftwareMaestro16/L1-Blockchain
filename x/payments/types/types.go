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
	NativeDenom			= "naet"
	CanonicalEncodingVersion	= byte(1)
	CurrentAppVersion		= uint32(1)
	CurrentStateVersion		= uint32(1)
	SignatureSchemeEd25519		= "ed25519-aetra-v1"
	SignatureObjectState		= "channel_state"
	SignatureObjectClaim		= "unidirectional_claim"
	SignatureObjectDelta		= "async_delta"
	SignatureObjectPromise		= "conditional_promise"
	SignatureObjectGossip		= "payment_gossip"
	SignatureObjectLiquidity	= "liquidity_reservation"
	SignatureObjectRoutingFee	= "routing_fee_policy"
	SignatureObjectVirtual		= "virtual_channel"
	SignatureObjectVirtualReserve	= "virtual_reservation"
	SignatureObjectVirtualClose	= "virtual_close"
	DefaultDisputePeriod		= uint64(16)
	DefaultOpeningFee		= "1"
	MaxDisputeExtensions		= uint32(2)
	MinCloseDelay			= uint64(1)
	MaxCloseDelay			= uint64(10_000)
	MinChallengePeriod		= uint64(1)
	MaxChallengePeriod		= uint64(20_000)
	MaxParticipants			= 8
	MaxConditionsPerState		= 128
	MaxParentChannels		= 16
	MaxSettlementBatchOps		= 256
	MaxRoutingHops			= 16
	MaxTokenLength			= 128
	MaxSettlementFeeFraction	= int64(10_000)
	MaxPenaltyRouteBps		= uint32(10_000)
	DefaultGossipTTL		= uint64(512)
	InvalidGossipPenalty		= int64(25)
	DefaultTimeoutMargin		= uint64(16)
	DefaultReplayHorizon		= uint64(100_000)
	SignerIsolationProcess		= "process"
	SignerIsolationHardware		= "hardware"
)

type ChannelType string

const (
	ChannelTypeBidirectional	ChannelType	= "BIDIRECTIONAL"
	ChannelTypeUnidirectional	ChannelType	= "UNIDIRECTIONAL"
	ChannelTypeAsync		ChannelType	= "ASYNC"
)

type ChannelStatus string

const (
	ChannelStatusOpen		ChannelStatus	= "OPEN"
	ChannelStatusPendingClose	ChannelStatus	= "PENDING_CLOSE"
	ChannelStatusSettled		ChannelStatus	= "SETTLED"
)

type ChannelFinality string

const (
	ChannelFinalityOpen				ChannelFinality	= "OPEN"
	ChannelFinalityPendingClose			ChannelFinality	= "PENDING_CLOSE"
	ChannelFinalityInDispute			ChannelFinality	= "IN_DISPUTE"
	ChannelFinalityPendingConditionResolution	ChannelFinality	= "PENDING_CONDITION_RESOLUTION"
	ChannelFinalityFinalizable			ChannelFinality	= "FINALIZABLE"
	ChannelFinalitySettled				ChannelFinality	= "SETTLED"
	ChannelFinalityPenalized			ChannelFinality	= "PENALIZED"
	ChannelFinalityExpired				ChannelFinality	= "EXPIRED"
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
	ConditionTypeHashLock	ConditionType	= "HASH_LOCK"
	ConditionTypeTimeLock	ConditionType	= "TIME_LOCK"
)

type FraudProofType string

const (
	FraudProofTypeDoubleSign	FraudProofType	= "DOUBLE_SIGN"
	FraudProofTypeStaleClose	FraudProofType	= "STALE_CLOSE"
	FraudProofTypeInvalidClose	FraudProofType	= "INVALID_CLOSE"
	FraudProofTypeInvalidBalance	FraudProofType	= "INVALID_BALANCE"
	FraudProofTypeInvalidCondition	FraudProofType	= "INVALID_CONDITION"
	FraudProofTypeReplayAttempt	FraudProofType	= "REPLAY_ATTEMPT"
	FraudProofTypeAsyncOverexposure	FraudProofType	= "ASYNC_OVEREXPOSURE"
)

type BatchOperationType string

const (
	BatchOperationOpen	BatchOperationType	= "OPEN"
	BatchOperationClose	BatchOperationType	= "CLOSE"
	BatchOperationDispute	BatchOperationType	= "DISPUTE"
	BatchOperationSettle	BatchOperationType	= "SETTLE"
)

type BlockSTMTransactionClass string

const (
	BlockSTMClassOpenChannel	BlockSTMTransactionClass	= "OPEN_CHANNEL"
	BlockSTMClassUpdateCheckpoint	BlockSTMTransactionClass	= "UPDATE_CHECKPOINT"
	BlockSTMClassCloseChannel	BlockSTMTransactionClass	= "CLOSE_CHANNEL"
	BlockSTMClassDisputeChannel	BlockSTMTransactionClass	= "DISPUTE_CHANNEL"
	BlockSTMClassSettleChannel	BlockSTMTransactionClass	= "SETTLE_CHANNEL"
	BlockSTMClassResolveCondition	BlockSTMTransactionClass	= "RESOLVE_CONDITION"
	BlockSTMClassBatchConditions	BlockSTMTransactionClass	= "BATCH_CONDITIONS"
	BlockSTMClassPenaltyAccounting	BlockSTMTransactionClass	= "PENALTY_ACCOUNTING"
)

type CloseReason string

const (
	CloseReasonUnilateral	CloseReason	= "UNILATERAL"
	CloseReasonCooperative	CloseReason	= "COOPERATIVE"
	CloseReasonTimeout	CloseReason	= "TIMEOUT"
	CloseReasonFraud	CloseReason	= "FRAUD"
)

type SettlementArbitrationOperation string

const (
	SettlementArbitrationOpen			SettlementArbitrationOperation	= "OPEN"
	SettlementArbitrationCollateralCustody		SettlementArbitrationOperation	= "COLLATERAL_CUSTODY"
	SettlementArbitrationCooperativeClose		SettlementArbitrationOperation	= "COOPERATIVE_CLOSE"
	SettlementArbitrationUnilateralClose		SettlementArbitrationOperation	= "UNILATERAL_CLOSE"
	SettlementArbitrationDispute			SettlementArbitrationOperation	= "DISPUTE"
	SettlementArbitrationFraudProof			SettlementArbitrationOperation	= "FRAUD_PROOF"
	SettlementArbitrationConditionResolution	SettlementArbitrationOperation	= "CONDITION_RESOLUTION"
	SettlementArbitrationPenaltyRouting		SettlementArbitrationOperation	= "PENALTY_ROUTING"
	SettlementArbitrationFinalSettlement		SettlementArbitrationOperation	= "FINAL_SETTLEMENT"
	SettlementArbitrationReplayProtection		SettlementArbitrationOperation	= "REPLAY_PROTECTION"
)

type ConditionSettlementMode string

const (
	ConditionSettlementModePreimage	ConditionSettlementMode	= "PREIMAGE"
	ConditionSettlementModeExpiry	ConditionSettlementMode	= "EXPIRY"
)

type GossipMessageType string

const (
	GossipChannelAnnouncement	GossipMessageType	= "ChannelAnnouncement"
	GossipChannelUpdate		GossipMessageType	= "ChannelUpdate"
	GossipLiquidityHint		GossipMessageType	= "LiquidityHint"
	GossipFeePolicyUpdate		GossipMessageType	= "FeePolicyUpdate"
	GossipNodeAnnouncement		GossipMessageType	= "NodeAnnouncement"
	GossipRouteFailure		GossipMessageType	= "RouteFailure"
	GossipCapacityProbe		GossipMessageType	= "CapacityProbe"
)

type VirtualChannelStatus string

const (
	VirtualChannelStatusOpen	VirtualChannelStatus	= "OPEN"
	VirtualChannelStatusSettled	VirtualChannelStatus	= "SETTLED"
)

type VirtualCloseMode string

const (
	VirtualCloseModeCooperative		VirtualCloseMode	= "COOPERATIVE_ENDPOINT"
	VirtualCloseModeExpired			VirtualCloseMode	= "EXPIRED"
	VirtualCloseModeIntermediaryRisk	VirtualCloseMode	= "INTERMEDIARY_RISK"
	VirtualCloseModeDisputed		VirtualCloseMode	= "DISPUTED"
)

type PenaltyRoute string

const (
	PenaltyRouteReporter		PenaltyRoute	= "REPORTER"
	PenaltyRouteCounterparty	PenaltyRoute	= "COUNTERPARTY"
	PenaltyRouteBurn		PenaltyRoute	= "BURN"
	PenaltyRouteSecurityReserve	PenaltyRoute	= "SECURITY_RESERVE"
	PenaltyRouteCommunityPool	PenaltyRoute	= "COMMUNITY_POOL"
)

type PaymentPenaltyClass string

const (
	PenaltyClassInvalidClose	PaymentPenaltyClass	= "INVALID_CLOSE_SUBMISSION"
	PenaltyClassStaleClose		PaymentPenaltyClass	= "STALE_CLOSE"
	PenaltyClassDoubleSign		PaymentPenaltyClass	= "SAME_NONCE_DOUBLE_SIGN"
	PenaltyClassInvalidCondition	PaymentPenaltyClass	= "INVALID_CONDITION_CLAIM"
	PenaltyClassReplayAttempt	PaymentPenaltyClass	= "REPLAY_ATTEMPT"
	PenaltyClassAsyncOverexposure	PaymentPenaltyClass	= "ASYNC_OVEREXPOSURE"
	PenaltyClassInvalidFraudProof	PaymentPenaltyClass	= "INVALID_FRAUD_PROOF"
)

type PaymentFeeClass string

const (
	PaymentFeeClassChannelOpen			PaymentFeeClass	= "CHANNEL_OPEN"
	PaymentFeeClassChannelCheckpoint		PaymentFeeClass	= "CHANNEL_CHECKPOINT"
	PaymentFeeClassCooperativeClose			PaymentFeeClass	= "COOPERATIVE_CLOSE"
	PaymentFeeClassUnilateralClose			PaymentFeeClass	= "UNILATERAL_CLOSE"
	PaymentFeeClassDispute				PaymentFeeClass	= "DISPUTE"
	PaymentFeeClassFraudProofVerification		PaymentFeeClass	= "FRAUD_PROOF_VERIFICATION"
	PaymentFeeClassConditionalPromiseSettlement	PaymentFeeClass	= "CONDITIONAL_PROMISE_SETTLEMENT"
	PaymentFeeClassVirtualChannelAnchor		PaymentFeeClass	= "VIRTUAL_CHANNEL_ANCHOR"
	PaymentFeeClassRoutingAdvertisement		PaymentFeeClass	= "ROUTING_ADVERTISEMENT"
)

type Balance struct {
	Participant	string
	Amount		string
}

type ConditionalPayment struct {
	ConditionID	string
	ConditionType	ConditionType
	Payer		string
	Payee		string
	Amount		string
	HashLock	string
	TimeoutHeight	uint64
	NonceStart	uint64
	NonceEnd	uint64
}

type ConditionalPromise struct {
	PromiseID			string
	ChannelID			string
	Source				string
	Destination			string
	Amount				string
	Fee				string
	HashLock			string
	TimeoutHeight			uint64
	TimeoutTimestamp		int64
	ConditionType			ConditionType
	RouteIDOptional			string
	PreviousPromiseIDOptional	string
	NextPromiseIDOptional		string
	Nonce				uint64
	PromiseHash			string
	Signature			PromiseSignature
}

type PromiseSignature struct {
	Signer			string
	ChainID			string
	ChannelID		string
	ObjectType		string
	Version			uint32
	Nonce			uint64
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	PromiseHash		string
	SignatureHash		string
}

type StateSignature struct {
	Signer			string
	ChainID			string
	ChannelID		string
	ObjectType		string
	Version			uint32
	Nonce			uint64
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	StateHash		string
	SignatureHash		string
}

type SignedNonceRecord struct {
	Signer		string
	ChainID		string
	ChannelID	string
	Epoch		uint64
	Nonce		uint64
	StateHash	string
	WALHash		string
	Released	bool
	IsolationMode	string
}

type SignerPersistence struct {
	Records		[]SignedNonceRecord
	IsolationMode	string
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
	Signer			string
	ChainID			string
	ChannelID		string
	ObjectType		string
	Version			uint32
	Nonce			uint64
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	ClaimHash		string
	SignatureHash		string
}

type DeltaSignature struct {
	Signer			string
	ChainID			string
	ChannelID		string
	ObjectType		string
	Version			uint32
	Nonce			uint64
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	DeltaHash		string
	SignatureHash		string
}

type GossipSignature struct {
	Signer			string
	ChainID			string
	ObjectType		string
	Version			uint32
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	SignatureHash		string
}

type ChannelOpenRequest struct {
	ChainID				string
	ChannelID			string
	Participants			[]string
	InitialBalances			[]Balance
	ChannelType			ChannelType
	Collateral			string
	CloseDelay			uint64
	ChallengePeriod			uint64
	FeePolicyID			string
	OpeningFeeDenom			string
	OpeningFeePaid			string
	RoutingAdvertised		bool
	ConditionalPaymentsSupported	bool
	OpenHeight			uint64
	ExpirationHeight		uint64
	ExpirationTimestamp		int64
}

type ChannelUpdateRequest struct {
	ChannelID		string
	State			ChannelState
	ConditionCommitments	[]ConditionalPayment
	AsyncDeltas		[]AsyncPaymentDelta
	RegisterCheckpoint	bool
	Submitter		string
	CurrentHeight		uint64
	CheckpointFeePaid	string
}

type ChannelUpdateResult struct {
	ChannelID		string
	StateHash		string
	Nonce			uint64
	ValidatedOffChain	bool
	CheckpointRegistered	bool
	Liquidity		[]Balance
}

type GossipMessage struct {
	MessageID		string
	MessageType		GossipMessageType
	ChainID			string
	ChannelID		string
	NodeID			string
	From			string
	To			string
	Capacity		string
	Liquidity		string
	FeeDenom		string
	FeeAmount		string
	MaxFee			string
	ValidAfterHeight	uint64
	ValidUntilHeight	uint64
	ChannelCommitment	string
	FailureCode		string
	ProbeAmount		string
	ReputationDelta		int64
	Sequence		uint64
	Advisory		bool
}

type SignedGossipEnvelope struct {
	Message		GossipMessage
	MessageHash	string
	Signature	GossipSignature
	ReceivedFrom	string
	ReceivedAt	uint64
}

type GossipReputation struct {
	NodeID			string
	Score			int64
	InvalidGossip		uint64
	LastUpdateHeight	uint64
}

type TopologyStore struct {
	Messages	[]SignedGossipEnvelope
	Edges		[]ChannelEdge
	Reputation	[]GossipReputation
	LastPrunedAt	uint64
}

type EdgeRoutingStats struct {
	ChannelID		string
	From			string
	To			string
	SuccessRateBps		uint32
	LiquidityUpdatedHeight	uint64
	CongestionBps		uint32
	NodeAvailabilityBps	uint32
	FailureCount		uint32
	TimeoutMargin		uint64
	PendingConditionCount	uint32
	AvgResolutionLatency	uint64
	RetryCount		uint32
	ReservePressureBps	uint32
	NodeQueueDelay		uint64
	LastFailureHeight	uint64
	LastUpdatedHeight	uint64
}

type RoutePolicy struct {
	MaxHops			int
	RequiredTimeoutMargin	uint64
	StaleLiquidityAfter	uint64
	HopPenalty		string
	CongestionPenalty	string
	StaleLiquidityPenalty	string
	FailurePenalty		string
	TimeoutPenalty		string
	SuccessPenalty		string
	AvailabilityPenalty	string
	ReservePressurePenalty	string
	QueueDelayPenalty	string
	PendingConditionPenalty	string
	LatencyPenalty		string
	ProportionalFeeBps	uint32
	DecayHalfLife		uint64
	MaxCongestedPaymentBps	uint32
	MaxFeeAmount		string
	EnableMultiPath		bool
	MaxSplits		int
	ExcludedNodes		[]string
	ExcludedChannels	[]string
	EdgeStats		[]EdgeRoutingStats
}

type RoutingFeePolicyUpdate struct {
	PolicyID		string
	ChainID			string
	ChannelID		string
	From			string
	To			string
	FeeDenom		string
	BaseHopFee		string
	ProportionalFeeBps	uint32
	LiquidityReservationFee	string
	VirtualChannelSetupFee	string
	CongestionSurcharge	string
	FailurePenalty		string
	MaxHopFee		string
	ValidAfterHeight	uint64
	ValidUntilHeight	uint64
	Sequence		uint64
	PolicyHash		string
	Signature		RoutingFeePolicySignature
}

type RoutingFeePolicySignature struct {
	Signer			string
	ChainID			string
	ChannelID		string
	ObjectType		string
	Version			uint32
	Sequence		uint64
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	SignatureHash		string
}

type HopFeeCalculationRequest struct {
	Amount			string
	Policy			RoutingFeePolicyUpdate
	CurrentHeight		uint64
	IncludeVirtualSetup	bool
	RepeatedInvalidAttempts	uint32
}

type RoutingHopFee struct {
	Denom			string
	BaseHopFee		string
	ProportionalFee		string
	LiquidityReservationFee	string
	VirtualChannelSetupFee	string
	CongestionSurcharge	string
	FailurePenalty		string
	RepeatedInvalidAttempts	uint32
	TotalFee		string
	PolicyHash		string
}

type RouteFailureClass string

const (
	RouteFailureCapacity		RouteFailureClass	= "CAPACITY"
	RouteFailureTimeout		RouteFailureClass	= "TIMEOUT"
	RouteFailureCongestion		RouteFailureClass	= "CONGESTION"
	RouteFailureLiquidityStale	RouteFailureClass	= "LIQUIDITY_STALE"
	RouteFailureNodeUnavailable	RouteFailureClass	= "NODE_UNAVAILABLE"
	RouteFailurePolicyRejected	RouteFailureClass	= "POLICY_REJECTED"
	RouteFailureUnknown		RouteFailureClass	= "UNKNOWN"
)

type RouteFailureReport struct {
	ChannelID	string
	From		string
	To		string
	FailureClass	RouteFailureClass
	Retryable	bool
	ObservedHeight	uint64
}

type CongestionSnapshot struct {
	ChannelID			string
	From				string
	To				string
	ChannelUpdateFailureRateBps	uint32
	PendingConditionCount		uint32
	AvgResolutionLatency		uint64
	RouteRetryCount			uint32
	ReservePressureBps		uint32
	NodeQueueDelay			uint64
	LiquidityUpdatedHeight		uint64
	ObservedHeight			uint64
}

type RouteSelectionRequest struct {
	From		string
	To		string
	Amount		string
	CurrentHeight	uint64
	Policy		RoutePolicy
}

type RouteRetryPolicy struct {
	MaxAttempts		uint32
	AlternateRouteLimit	uint32
	ExcludeFailedEdges	bool
	CongestionRetryDelay	uint64
}

type RouteRetryRequest struct {
	Selection	RouteSelectionRequest
	Failures	[]RouteFailureReport
	Policy		RouteRetryPolicy
}

type RouteRetryResult struct {
	Route		ScoredRoute
	Attempts	uint32
	Retryable	bool
	Reason		string
	PolicyHash	string
}

type ScoredRoute struct {
	Edges		[]ChannelEdge
	Amount		string
	TotalFee	string
	TotalCost	string
	MinCapacity	string
	ScoreHash	string
}

type RouteSimulationResult struct {
	Route		ScoredRoute
	Attemptable	bool
	Reason		string
	TotalFee	string
}

type MultiPathRoute struct {
	Parts		[]ScoredRoute
	TotalAmount	string
	TotalFee	string
	ScoreHash	string
}

type ForwardingPacket struct {
	PacketID	string
	RouteID		string
	HopPaymentID	string
	ChannelID	string
	ForwardingNode	string
	NextNode	string
	Amount		string
	FeeAmount	string
	TimeoutHeight	uint64
	NextPacketHash	string
	PacketHash	string
}

type ForwardingPacketReplayRecord struct {
	PacketID	string
	RouteID		string
	HopPaymentID	string
	RecordedHeight	uint64
	ExpiresHeight	uint64
}

type ForwardingLogRecord struct {
	PacketID	string
	RouteID		string
	HopPaymentID	string
	ChannelID	string
	ForwardingNode	string
	NextNodeHash	string
	AmountHash	string
	RecordedHeight	uint64
}

type ChannelCloseRequest struct {
	ChannelID	string
	ClosingState	ChannelState
	Signatures	[]StateSignature
	CloseReason	CloseReason
	Submitter	string
	CurrentHeight	uint64
	SettlementFee	string
}

type ConditionResolution struct {
	ConditionID	string
	Resolver	string
	Recipient	string
	Amount		string
	Expired		bool
	EvidenceHash	string
}

type ClosedChannelTombstone struct {
	ChainID		string
	ChannelID	string
	FinalizedNonce	uint64
	StateHash	string
	ClosedHeight	uint64
	ExpiresHeight	uint64
}

type ConditionClaimRecord struct {
	ChainID		string
	ChannelID	string
	ConditionID	string
	EvidenceHash	string
	PreimageHash	string
	ResolvedHeight	uint64
	ExpiresHeight	uint64
}

type AsyncFinalizationJob struct {
	JobID		string
	ChannelID	string
	FinalizeHeight	uint64
	EnqueuedHeight	uint64
	Attempts	uint32
	LastRunHeight	uint64
	LastError	string
	Completed	bool
	CompletedHeight	uint64
	SettlementHash	string
}

type AsyncPromiseExpiryJob struct {
	JobID			string
	ChannelID		string
	PromiseID		string
	Promise			ConditionalPromise
	Resolver		string
	ExpireAfterHeight	uint64
	EnqueuedHeight		uint64
	Attempts		uint32
	LastRunHeight		uint64
	LastError		string
	Completed		bool
	CompletedHeight		uint64
	ResolutionHash		string
}

type AsyncSettlementCompletion struct {
	CompletionID	string
	JobID		string
	JobType		string
	ChannelID	string
	ObjectID	string
	ResultHash	string
	Height		uint64
}

type AsyncExecutionResult struct {
	ProcessedFinalizations		uint64
	ProcessedPromiseExpiries	uint64
	CompletedJobIDs			[]string
	FailedJobIDs			[]string
	EmittedCompletionIDs		[]string
}

type PreimageRevealRequest struct {
	ChannelID	string
	Promises	[]ConditionalPromise
	Preimage	string
	Revealer	string
	CurrentHeight	uint64
}

type PromiseExpiryRequest struct {
	ChannelID	string
	Promises	[]ConditionalPromise
	Resolver	string
	CurrentHeight	uint64
}

type ConditionLinkageProof struct {
	RouteID				string
	Promises			[]ConditionalPromise
	Sender				string
	Receiver			string
	Amount				string
	TotalFees			string
	HashLock			string
	TimeoutMargin			uint64
	PartialDispute			bool
	OffchainResolvedPromiseIDs	[]string
	EvidenceHash			string
}

type RouteFeeClaim struct {
	ChannelID	string
	PromiseID	string
	Recipient	string
	Amount		string
	EvidenceHash	string
}

type BatchConditionSettlementRequest struct {
	LinkageProof		ConditionLinkageProof
	Mode			ConditionSettlementMode
	Preimage		string
	Resolver		string
	CurrentHeight		uint64
	SettlementFeePaid	string
}

type BatchConditionSettlementResult struct {
	RouteID			string
	Resolutions		[]ConditionResolution
	FeeClaims		[]RouteFeeClaim
	ConditionRootUpdates	[]ConditionRootUpdate
	EvidenceHash		string
}

type ConditionRootUpdate struct {
	ChannelID	string
	Nonce		uint64
	ConditionRoot	string
	ConditionCount	uint32
	Conditions	[]ConditionalPayment
}

type ChannelDisputeRequest struct {
	ChannelID		string
	ClosingStateReference	string
	NewerState		ChannelState
	FraudProof		FraudProof
	ConditionProofs		[]ConditionResolution
	Submitter		string
	CurrentHeight		uint64
	DisputeFeePaid		string
}

type WatchDisputeSubmission struct {
	WatchService		string
	Delegator		string
	ChannelID		string
	ClosingStateReference	string
	NewerState		ChannelState
	CurrentHeight		uint64
	EvidenceHash		string
}

type ValidatorPaymentServiceMetadata struct {
	ValidatorAddress	string
	ServiceAddress		string
	WatchEndpoint		string
	RoutingEndpoint		string
	PublicKey		string
	MinDelegation		string
	CommissionBps		uint32
	Active			bool
	UpdatedHeight		uint64
	MetadataHash		string
}

type ValidatorWatchRegistration struct {
	ValidatorAddress	string
	ServiceAddress		string
	Delegator		string
	MinDelegation		string
	RegisteredHeight	uint64
	MetadataHash		string
}

type ValidatorAssistedDisputeSubmission struct {
	ValidatorAddress	string
	ServiceAddress		string
	Delegator		string
	ChannelID		string
	ClosingStateReference	string
	NewerState		ChannelState
	CurrentHeight		uint64
	EvidenceHash		string
}

type FinalSettlementRequest struct {
	ChannelID		string
	ResolvedConditions	[]ConditionResolution
	CurrentHeight		uint64
	FeeAccountingState	string
	RoutingFeeClaimHash	string
}

type SettlementArbitrationInput struct {
	Operation		SettlementArbitrationOperation
	ChannelID		string
	SignedState		ChannelState
	Claim			UnidirectionalClaim
	FraudProof		FraudProof
	ConditionProofs		[]ConditionResolution
	RouteHints		[]ChannelEdge
	GossipStateHash		string
	ExternalLiquidity	[]Balance
	UnsignedBalances	[]Balance
	OffchainIntent		string
	CurrentHeight		uint64
}

type StateHashDebug struct {
	ChannelID			string
	Status				ChannelStatus
	LatestNonce			uint64
	LatestStateHash			string
	ComputedLatestStateHash		string
	PendingNonce			uint64
	PendingStateHash		string
	ComputedPendingStateHash	string
	FinalizedNonce			uint64
	DisputedNonce			uint64
}

type ChannelState struct {
	ChainID			string
	AppVersion		uint32
	ModuleName		string
	RequiredFields		[]string
	ChannelID		string
	ChannelType		ChannelType
	ParticipantSetHash	string
	Denom			string
	Version			uint32
	ParticipantA		string
	ParticipantB		string
	BalanceA		string
	BalanceB		string
	ReserveA		string
	ReserveB		string
	AccruedFees		string
	Epoch			uint64
	Nonce			uint64
	PendingConditionsRoot	string
	ConditionRoot		string
	ConditionCount		uint32
	Balances		[]Balance
	Conditions		[]ConditionalPayment
	PreviousStateHash	string
	StateHash		string
	TimeoutHeight		uint64
	TimeoutTimestamp	int64
	ChallengePeriod		uint64
	CloseDelay		uint64
	FeePolicyID		string
	RequiredSignerBitmap	string
	SignatureScheme		string
	SignaturePreimageHash	string
	CheckpointNonce		uint64
	CheckpointBalances	[]Balance
	AsyncUpdateRoot		string
	AcceptedUpdateRoot	string
	SendWindow		uint64
	ReceiveWindow		uint64
	MaxUnackedAmount	string
	ExpiryHeight		uint64
	Signatures		[]StateSignature
}

type AsyncPaymentDelta struct {
	UpdateID	string
	ChainID		string
	ChannelID	string
	From		string
	To		string
	Direction	string
	Amount		string
	NonceStart	uint64
	NonceEnd	uint64
	ExpiryHeight	uint64
	DeltaHash	string
	Signature	DeltaSignature
}

type AsyncDeltaDisputeProof struct {
	ProofID		string
	ChannelID	string
	CheckpointState	ChannelState
	Deltas		[]AsyncPaymentDelta
	EvidenceHash	string
}

type UnidirectionalClaim struct {
	ChainID			string
	ChannelID		string
	Payer			string
	Receiver		string
	LockedAmount		string
	ClaimedAmount		string
	Nonce			uint64
	ExpirationHeight	uint64
	ExpirationTimestamp	int64
	StateHash		string
	PayerSignature		ClaimSignature
	ReceiverAckOptional	ClaimSignature
}

type StreamingPaymentFrame struct {
	ChannelID		string
	StreamID		string
	Payer			string
	Receiver		string
	PreviousClaimed		string
	RatePerBlock		string
	StartHeight		uint64
	CurrentHeight		uint64
	Nonce			uint64
	ExpirationHeight	uint64
	ExpirationTimestamp	int64
}

type PendingClose struct {
	Submitter		string
	SubmittedHeight		uint64
	SettleAfterHeight	uint64
	DisputeCount		uint32
	CloseReason		CloseReason
	SettlementFeeDenom	string
	SettlementFee		string
	State			ChannelState
	FraudProofs		[]FraudProof
	ConditionProofs		[]ConditionResolution
	Penalties		[]Penalty
	PenaltyAllocations	[]PenaltyAllocation
}

type FraudProof struct {
	ProofID			string
	ProofType		FraudProofType
	SubmittedBy		string
	OffendingSigner		string
	StateA			ChannelState
	StateB			ChannelState
	AsyncProof		AsyncDeltaDisputeProof
	PenaltyDenom		string
	PenaltyAmount		string
	EvidenceHash		string
	VerificationFeePaid	string
}

type Penalty struct {
	Offender	string
	Recipient	string
	Denom		string
	Amount		string
}

type PenaltyAllocation struct {
	Offender	string
	Route		PenaltyRoute
	Denom		string
	Amount		string
}

type FraudPenaltyPolicy struct {
	ReporterRewardCap	string
	CounterpartyRewardCap	string
	CounterpartyRewardBps	uint32
	BurnShareBps		uint32
	SecurityReserveShareBps	uint32
	CommunityPoolShareBps	uint32
	SecurityReserveHook	bool
}

type PenaltySource string

const (
	PenaltySourceChannelBalance			PenaltySource	= "CHANNEL_BALANCE"
	PenaltySourceParticipantBond			PenaltySource	= "PARTICIPANT_BOND"
	PenaltySourceRoutingAdvertisementDeposit	PenaltySource	= "ROUTING_ADVERTISEMENT_DEPOSIT"
	PenaltySourceFraudProofDeposit			PenaltySource	= "FRAUD_PROOF_DEPOSIT"
)

type PenaltyMatrixEntry struct {
	Class				PaymentPenaltyClass
	ProofType			FraudProofType
	Source				PenaltySource
	BasePenalty			string
	ReporterRewardCap		string
	CounterpartyCompensation	string
	BurnShareBps			uint32
	SecurityReserveShareBps		uint32
	CommunityPoolShareBps		uint32
	InvalidProofVerifierCost	string
	Bounded				bool
}

type PenaltyRouteAccounting struct {
	Class			PaymentPenaltyClass
	Source			PenaltySource
	TotalPenalty		string
	ReporterReward		string
	CounterpartyComp	string
	Allocations		[]PenaltyAllocation
	Penalties		[]Penalty
}

type InvalidFraudProofSubmissionPenalty struct {
	Submitter		string
	Denom			string
	DepositAmount		string
	VerificationCost	string
	ForfeitedAmount		string
	RefundAmount		string
}

type LiquidityAdvertisement struct {
	AdvertisementID		string
	ChannelID		string
	Advertiser		string
	Counterparty		string
	Capacity		string
	FeeDenom		string
	BaseFee			string
	ReservationFee		string
	VirtualSetupFee		string
	ReliabilityBps		uint32
	ValidUntilHeight	uint64
	DepositAmount		string
	BackedByReservation	bool
	AdvertisementHash	string
}

type LiquidityReservationSignature struct {
	Signer			string
	ChainID			string
	ChannelID		string
	ObjectType		string
	Version			uint32
	Nonce			uint64
	ObjectID		string
	ExpirationHeight	uint64
	CommitmentHash		string
	SignatureHash		string
}

type SignedLiquidityReservation struct {
	ReservationID		string
	AdvertisementID		string
	ChainID			string
	ChannelID		string
	Reserver		string
	Counterparty		string
	Capacity		string
	FeeAmount		string
	ExpirationHeight	uint64
	Nonce			uint64
	CommitmentHash		string
	Signature		LiquidityReservationSignature
}

type PaymentFeeSchedule struct {
	Denom				string
	ChannelOpenFee			string
	ChannelOpenPerParticipantFee	string
	ChannelCheckpointFee		string
	CooperativeCloseFee		string
	UnilateralCloseFee		string
	DisputeFee			string
	FraudProofVerificationFee	string
	ConditionalPromiseSettlementFee	string
	VirtualChannelAnchorFee		string
	RoutingAdvertisementFee		string
	RoutingAdvertisementDeposit	string
	ConditionalCapabilitySurcharge	string
	VirtualChannelAnchorSurcharge	string
	StorageByteFee			string
	StorageFeeEnabled		bool
	OpenFeeMin			string
	OpenFeeMax			string
	StorageRentPerBlock		string
	RenewalPeriod			uint64
	BaseMultiplierBps		uint32
	MaxMultiplierBps		uint32
}

type PaymentFeeMultiplier struct {
	FeeClass	PaymentFeeClass
	MultiplierBps	uint32
	CongestionBps	uint32
	UpdatedHeight	uint64
}

type PaymentFeeCharge struct {
	FeeID		string
	FeeClass	PaymentFeeClass
	ChannelID	string
	ObjectID	string
	Payer		string
	Denom		string
	Amount		string
	RequiredAmount	string
	StorageBytes	uint64
	MultiplierBps	uint32
	Height		uint64
	Refunded	bool
}

type PaymentFeeRefund struct {
	RefundID	string
	FeeID		string
	Recipient	string
	Denom		string
	Amount		string
	Reason		string
	Height		uint64
}

type SecurityReserveAllocationHook struct {
	HookID		string
	ChannelID	string
	ProofID		string
	Offender	string
	Denom		string
	Amount		string
	Height		uint64
	Route		PenaltyRoute
	Allocation	string
}

type SettlementInclusionLatency struct {
	RecordID	string
	OperationID	string
	ChannelID	string
	Operation	SettlementArbitrationOperation
	SubmittedHeight	uint64
	IncludedHeight	uint64
	LatencyBlocks	uint64
	SLOThreshold	uint64
	Breached	bool
}

type SettlementGasCostSchedule struct {
	OpenGas			uint64
	CooperativeCloseGas	uint64
	UnilateralCloseGas	uint64
	DisputeGas		uint64
	FraudProofGas		uint64
	ConditionResolutionGas	uint64
	PenaltyRoutingGas	uint64
	FinalSettlementGas	uint64
	ReplayProtectionGas	uint64
	PerSignatureGas		uint64
	PerConditionGas		uint64
	PerFraudProofGas	uint64
	PerPenaltyAllocationGas	uint64
	PerStateByteGas		uint64
}

type SettlementGasEstimate struct {
	Operation		SettlementArbitrationOperation
	BaseGas			uint64
	SignatureGas		uint64
	ConditionGas		uint64
	FraudProofGas		uint64
	PenaltyAllocationGas	uint64
	StateByteGas		uint64
	TotalGas		uint64
}

type ChannelOpenFeeFormula struct {
	Denom			string
	BaseFee			string
	ParticipantFee		string
	ParticipantCount	uint64
	StorageByteFee		string
	StorageBytes		uint64
	StorageFee		string
	ConditionalSurcharge	string
	VirtualAnchorSurcharge	string
	RoutingDeposit		string
	RentReserve		string
	MultiplierBps		uint32
	MinFee			string
	MaxFee			string
	TotalFee		string
}

func DefaultPaymentFeeSchedule() PaymentFeeSchedule {
	return PaymentFeeSchedule{
		Denom:					NativeDenom,
		ChannelOpenFee:				DefaultOpeningFee,
		ChannelOpenPerParticipantFee:		"0",
		ChannelCheckpointFee:			"0",
		CooperativeCloseFee:			"0",
		UnilateralCloseFee:			"0",
		DisputeFee:				"0",
		FraudProofVerificationFee:		"0",
		ConditionalPromiseSettlementFee:	"0",
		VirtualChannelAnchorFee:		"0",
		RoutingAdvertisementFee:		"0",
		RoutingAdvertisementDeposit:		"0",
		ConditionalCapabilitySurcharge:		"0",
		VirtualChannelAnchorSurcharge:		"0",
		StorageByteFee:				"0",
		OpenFeeMin:				DefaultOpeningFee,
		OpenFeeMax:				"0",
		StorageRentPerBlock:			"0",
		BaseMultiplierBps:			10_000,
		MaxMultiplierBps:			100_000,
	}
}

type ChannelRecord struct {
	ChainID			string
	ChannelID		string
	ChannelType		ChannelType
	Participants		[]string
	RequiredSigners		[]string
	Payer			string
	Receiver		string
	ReceiverAckRequired	bool
	Denom			string
	Collateral		string
	OpenHeight		uint64
	CloseDelay		uint64
	DisputePeriod		uint64
	ExpirationHeight	uint64
	ExpirationTimestamp	int64
	OpeningFeeDenom		string
	OpeningFeePaid		string
	RoutingAdvertised	bool
	ConditionalPayments	bool
	CustodyDenom		string
	CustodyAmount		string
	Status			ChannelStatus
	Finality		ChannelFinality
	OpeningStateHash	string
	FinalizedNonce		uint64
	DisputedNonce		uint64
	LatestState		ChannelState
	LatestClaim		UnidirectionalClaim
	PendingClose		PendingClose
}

type SettlementRecord struct {
	ChainID			string
	ChannelID		string
	StateHash		string
	Nonce			uint64
	FinalBalances		[]Balance
	SettlementFeeDenom	string
	SettlementFee		string
	Penalties		[]Penalty
	PenaltyAllocations	[]PenaltyAllocation
	SettledHeight		uint64
	SettlementHash		string
}

type CustodyLock struct {
	ChannelID	string
	Denom		string
	Amount		string
}

type PaymentEventAttribute struct {
	Key	string
	Value	string
}

type PaymentEvent struct {
	EventID		string
	EventType	string
	ChannelID	string
	Height		uint64
	Attributes	[]PaymentEventAttribute
}

type StoreV2ChannelRecord struct {
	Key			string
	Version			uint64
	ChannelID		string
	Channel			ChannelRecord
	LatestStateHash		string
	LatestStateNonce	uint64
	PendingCloseKey		string
	ParticipantIndexKeys	[]string
	RoutingAdvertisementKey	string
}

type StoreV2ChannelStateRecord struct {
	Key			string
	Version			uint64
	ChannelID		string
	Nonce			uint64
	StateHash		string
	FullState		ChannelState
	SubmittedOnChain	bool
	CheckpointHeight	uint64
}

type StoreV2PendingCloseRecord struct {
	Key		string
	Version		uint64
	ChannelID	string
	Close		PendingClose
}

type StoreV2ConditionRecord struct {
	Key		string
	Version		uint64
	ConditionID	string
	ChannelID	string
	Promise		ConditionalPromise
	ExpiresHeight	uint64
	Settled		bool
	ClaimEvidence	string
}

type StoreV2VirtualChannelRecord struct {
	Key			string
	Version			uint64
	VirtualChannelID	string
	Channel			VirtualChannel
	AnchorHash		string
}

type StoreV2ParticipantChannelRecord struct {
	Key		string
	Version		uint64
	Participant	string
	ChannelID	string
}

type StoreV2SettlementTombstoneRecord struct {
	Key			string
	Version			uint64
	ChannelID		string
	Tombstone		ClosedChannelTombstone
	PruneAfterHeight	uint64
}

type StoreV2FeeAccumulatorRecord struct {
	Key		string
	Version		uint64
	BlockOrEpoch	string
	Bucket		string
	Amount		string
}

type StoreV2FraudProofRecord struct {
	Key		string
	Version		uint64
	ProofID		string
	ChannelID	string
	Proof		FraudProof
}

type StoreV2Layout struct {
	Version			uint64
	Channels		[]StoreV2ChannelRecord
	ChannelStates		[]StoreV2ChannelStateRecord
	PendingCloses		[]StoreV2PendingCloseRecord
	Conditions		[]StoreV2ConditionRecord
	VirtualChannels		[]StoreV2VirtualChannelRecord
	ParticipantChannels	[]StoreV2ParticipantChannelRecord
	SettlementTombstones	[]StoreV2SettlementTombstoneRecord
	FeeAccumulators		[]StoreV2FeeAccumulatorRecord
	FraudProofs		[]StoreV2FraudProofRecord
}

type ParticipantChannelPageRequest struct {
	Address	string
	Offset	uint64
	Limit	uint64
}

type ParticipantChannelPageResponse struct {
	Entries		[]StoreV2ParticipantChannelRecord
	NextOffset	uint64
	Total		uint64
}

type AdaptiveSyncSnapshot struct {
	Key			string
	Version			uint64
	Height			uint64
	Layout			StoreV2Layout
	ActiveDisputes		[]AdaptiveSyncActiveDisputeIndex
	PendingFinalizations	[]AdaptiveSyncPendingFinalizationIndex
	WatcherReplayEvents	[]AdaptiveSyncWatcherReplayEvent
	SnapshotHash		string
	ConsensusOnly		bool
	RoutingTopologyExcluded	bool
}

type AdaptiveSyncActiveDisputeIndex struct {
	Key			string
	ChannelID		string
	PendingStateHash	string
	PendingNonce		uint64
	SubmittedHeight		uint64
	SettleAfterHeight	uint64
	DisputeCount		uint32
	Submitter		string
}

type AdaptiveSyncPendingFinalizationIndex struct {
	Key			string
	ChannelID		string
	PendingHeight		uint64
	Finality		ChannelFinality
	PendingStateHash	string
	PendingNonce		uint64
}

type AdaptiveSyncWatcherReplayEvent struct {
	Key		string
	Event		PaymentEvent
	EventHash	string
}

type AdaptiveSyncRecoveryState struct {
	ActiveChannelIDs		[]string
	PendingCloseChannelIDs		[]string
	UnresolvedConditionIDs		[]string
	VirtualChannelIDs		[]string
	SettlementTombstoneIDs		[]string
	ActiveDisputeChannelIDs		[]string
	PendingFinalizationIDs		[]string
	WatcherReplayEventIDs		[]string
	RecoveredFromSnapshotHash	string
}

type ChannelEdge struct {
	ChannelID		string
	From			string
	To			string
	Capacity		string
	FeeDenom		string
	FeeAmount		string
	AdvertisementFeePaid	string
	ExpiresHeight		uint64
	Active			bool
}

type VirtualChannel struct {
	VirtualChannelID		string
	ChainID				string
	Nonce				uint64
	ParentRouteID			string
	ParentChannelIDs		[]string
	ParentReserveCommitments	[]string
	Endpoints			[]string
	EndpointA			string
	EndpointB			string
	Intermediaries			[]string
	IntermediarySetHash		string
	Capacity			string
	BalanceA			string
	BalanceB			string
	RoutingFeeAmount		string
	AnchorFeePaid			string
	ExpiresHeight			uint64
	Status				VirtualChannelStatus
	AnchorCommitment		string
	ConditionRoot			string
	StateHash			string
	Signatures			[]StateSignature
}

type VirtualReservationSignature struct {
	Signer			string
	ChainID			string
	VirtualChannelID	string
	ParentRouteID		string
	ParentChannelID		string
	ObjectType		string
	Version			uint32
	Capacity		string
	SplitAmount		string
	FeeAmount		string
	ExpirationHeight	uint64
	CommitmentHash		string
	SignatureHash		string
}

type VirtualParentReserve struct {
	SegmentID		string
	ParentChannelID		string
	ReservedBy		string
	Capacity		string
	SplitAmount		string
	FeeAmount		string
	ReserveCommitment	string
	Signature		VirtualReservationSignature
}

type VirtualActivationProof struct {
	VirtualChannel		VirtualChannel
	ParentReserves		[]VirtualParentReserve
	RouteTimeoutHeight	uint64
	AggregatedCapacity	bool
	ProofHash		string
}

type VirtualReserveRelease struct {
	SegmentID		string
	VirtualChannelID	string
	ParentChannelID		string
	ReserveCommitment	string
	Capacity		string
	BalanceA		string
	BalanceB		string
	FeeAmount		string
	ReleaseHeight		uint64
	ReleaseHash		string
}

type VirtualCloseProof struct {
	VirtualChannelID		string
	ParentRouteID			string
	CloseMode			VirtualCloseMode
	FinalState			VirtualChannel
	ParentReserveCommitments	[]string
	SubmittedBy			string
	CloseHeight			uint64
	ReleaseHeight			uint64
	ProofHash			string
}

type VirtualChannelDisputeProof struct {
	VirtualChannelID		string
	ParentRouteID			string
	LatestState			VirtualChannel
	ParentReserveCommitments	[]string
	SubmittedBy			string
	EvidenceHash			string
}

type VirtualReserveSegment struct {
	SegmentID		string
	VirtualChannelID	string
	ParentChannelID		string
	ReserveCommitment	string
	Capacity		string
	BalanceA		string
	BalanceB		string
	FeeAmount		string
	SegmentHash		string
}

type VirtualSegmentSettlementProof struct {
	SegmentID		string
	VirtualChannelID	string
	ParentChannelID		string
	FinalStateHash		string
	ReserveCommitment	string
	BalanceA		string
	BalanceB		string
	SettlementHash		string
}

type VirtualPartialActivationFailure struct {
	VirtualChannelID	string
	FailedSegmentID		string
	Reason			string
	RefundCommitments	[]string
	FailureHash		string
}

type SettlementOperation struct {
	OperationID	string
	OperationType	BatchOperationType
	ChannelID	string
	Nonce		uint64
	StateHash	string
}

type BlockSTMAccessPlan struct {
	OperationID		string
	TxClass			BlockSTMTransactionClass
	ChannelID		string
	ConditionIDs		[]string
	ReadKeys		[]string
	WriteKeys		[]string
	AccumulatorKeys		[]string
	ConflictDomain		string
	DeterministicGroup	string
}

type BlockSTMConflict struct {
	LeftOperationID		string
	RightOperationID	string
	Key			string
	Reason			string
}

type BlockSTMConflictProfile struct {
	Plans				[]BlockSTMAccessPlan
	Conflicts			[]BlockSTMConflict
	ParallelizableGroups		[][]string
	ConflictFree			bool
	GlobalAccountingDeferred	bool
}

type PaymentBlockAccumulator struct {
	BlockHeight	uint64
	FeeAmount	string
	BurnAmount	string
	PenaltyAmount	string
	OperationCount	uint64
	AccumulatorKey	string
}

type SettlementBatch struct {
	BatchID		string
	Operations	[]SettlementOperation
	RootHash	string
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
		Signer:			signer,
		ChainID:		state.ChainID,
		ChannelID:		state.ChannelID,
		ObjectType:		SignatureObjectState,
		Version:		state.Version,
		Nonce:			state.Nonce,
		ObjectID:		state.StateHash,
		ExpirationHeight:	state.TimeoutHeight,
		CommitmentHash:		state.StateHash,
		StateHash:		state.StateHash,
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
		Signer:		signer,
		ChainID:	state.ChainID,
		ChannelID:	state.ChannelID,
		Epoch:		state.Epoch,
		Nonce:		state.Nonce,
		StateHash:	state.StateHash,
		IsolationMode:	isolationMode,
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
		ChainID:		req.ChainID,
		ChannelID:		req.ChannelID,
		ChannelType:		req.ChannelType,
		Participants:		req.Participants,
		Denom:			NativeDenom,
		Collateral:		req.Collateral,
		OpenHeight:		req.OpenHeight,
		CloseDelay:		req.CloseDelay,
		DisputePeriod:		req.ChallengePeriod,
		ExpirationHeight:	req.ExpirationHeight,
		ExpirationTimestamp:	req.ExpirationTimestamp,
		OpeningFeeDenom:	req.OpeningFeeDenom,
		OpeningFeePaid:		req.OpeningFeePaid,
		RoutingAdvertised:	req.RoutingAdvertised,
		ConditionalPayments:	req.ConditionalPaymentsSupported,
		CustodyDenom:		NativeDenom,
		CustodyAmount:		req.Collateral,
		Status:			ChannelStatusOpen,
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
	r.CheckpointFeePaid = strings.TrimSpace(r.CheckpointFeePaid)
	if r.CheckpointFeePaid == "" {
		r.CheckpointFeePaid = "0"
	}
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
	r.DisputeFeePaid = strings.TrimSpace(r.DisputeFeePaid)
	if r.DisputeFeePaid == "" {
		r.DisputeFeePaid = "0"
	}
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

func (m ValidatorPaymentServiceMetadata) Normalize() ValidatorPaymentServiceMetadata {
	m.ValidatorAddress = strings.TrimSpace(m.ValidatorAddress)
	m.ServiceAddress = strings.TrimSpace(m.ServiceAddress)
	m.WatchEndpoint = strings.TrimSpace(m.WatchEndpoint)
	m.RoutingEndpoint = strings.TrimSpace(m.RoutingEndpoint)
	m.PublicKey = strings.TrimSpace(m.PublicKey)
	m.MinDelegation = strings.TrimSpace(m.MinDelegation)
	if m.MinDelegation == "" {
		m.MinDelegation = "0"
	}
	m.MetadataHash = normalizeOptionalHash(m.MetadataHash)
	return m
}

func (m ValidatorPaymentServiceMetadata) Validate() error {
	m = m.Normalize()
	if err := addressing.ValidateUserAddress("payments validator service validator", m.ValidatorAddress); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments validator service address", m.ServiceAddress); err != nil {
		return err
	}
	if m.WatchEndpoint == "" && m.RoutingEndpoint == "" {
		return errors.New("payments validator service requires watch or routing endpoint")
	}
	if err := validateNonNegativeInt("payments validator service minimum delegation", m.MinDelegation); err != nil {
		return err
	}
	if m.CommissionBps > MaxPenaltyRouteBps {
		return errors.New("payments validator service commission exceeds 10000 bps")
	}
	if m.UpdatedHeight == 0 {
		return errors.New("payments validator service update height must be positive")
	}
	expected := ComputeValidatorPaymentServiceMetadataHash(m)
	if m.MetadataHash != "" && m.MetadataHash != expected {
		return errors.New("payments validator service metadata hash mismatch")
	}
	return nil
}

func (r ValidatorWatchRegistration) Normalize() ValidatorWatchRegistration {
	r.ValidatorAddress = strings.TrimSpace(r.ValidatorAddress)
	r.ServiceAddress = strings.TrimSpace(r.ServiceAddress)
	r.Delegator = strings.TrimSpace(r.Delegator)
	r.MinDelegation = strings.TrimSpace(r.MinDelegation)
	if r.MinDelegation == "" {
		r.MinDelegation = "0"
	}
	r.MetadataHash = normalizeOptionalHash(r.MetadataHash)
	return r
}

func (r ValidatorWatchRegistration) Validate(metadata ValidatorPaymentServiceMetadata) error {
	r = r.Normalize()
	metadata = metadata.Normalize()
	if err := metadata.Validate(); err != nil {
		return err
	}
	if !metadata.Active || metadata.WatchEndpoint == "" {
		return errors.New("payments validator watch service is not active")
	}
	if r.ValidatorAddress != metadata.ValidatorAddress || r.ServiceAddress != metadata.ServiceAddress {
		return errors.New("payments validator watch registration service mismatch")
	}
	if err := addressing.ValidateUserAddress("payments validator watch delegator", r.Delegator); err != nil {
		return err
	}
	if r.RegisteredHeight == 0 {
		return errors.New("payments validator watch registration height must be positive")
	}
	if r.RegisteredHeight < metadata.UpdatedHeight {
		return errors.New("payments validator watch registration predates metadata")
	}
	if r.MetadataHash != "" && r.MetadataHash != metadata.MetadataHash {
		return errors.New("payments validator watch registration metadata hash mismatch")
	}
	return validateNonNegativeInt("payments validator watch minimum delegation", r.MinDelegation)
}

func (s ValidatorAssistedDisputeSubmission) Normalize() ValidatorAssistedDisputeSubmission {
	s.ValidatorAddress = strings.TrimSpace(s.ValidatorAddress)
	s.ServiceAddress = strings.TrimSpace(s.ServiceAddress)
	s.Delegator = strings.TrimSpace(s.Delegator)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ClosingStateReference = normalizeHash(s.ClosingStateReference)
	s.NewerState = s.NewerState.Normalize()
	s.EvidenceHash = normalizeOptionalHash(s.EvidenceHash)
	return s
}

func (s ValidatorAssistedDisputeSubmission) ValidateForChannel(channel ChannelRecord, metadata ValidatorPaymentServiceMetadata) error {
	s = s.Normalize()
	metadata = metadata.Normalize()
	if err := metadata.Validate(); err != nil {
		return err
	}
	if !metadata.Active || metadata.WatchEndpoint == "" {
		return errors.New("payments validator watch service is not active")
	}
	if s.ValidatorAddress != metadata.ValidatorAddress {
		return errors.New("payments validator assisted dispute validator mismatch")
	}
	if s.ServiceAddress == "" {
		s.ServiceAddress = metadata.ServiceAddress
	}
	if s.ServiceAddress != metadata.ServiceAddress {
		return errors.New("payments validator assisted dispute service mismatch")
	}
	return (WatchDisputeSubmission{
		WatchService:		metadata.ServiceAddress,
		Delegator:		s.Delegator,
		ChannelID:		s.ChannelID,
		ClosingStateReference:	s.ClosingStateReference,
		NewerState:		s.NewerState,
		CurrentHeight:		s.CurrentHeight,
		EvidenceHash:		s.EvidenceHash,
	}).ValidateForChannel(channel)
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
		Signer:			signer,
		ChainID:		channel.ChainID,
		ChannelID:		promise.ChannelID,
		ObjectType:		SignatureObjectPromise,
		Version:		CurrentStateVersion,
		Nonce:			promise.Nonce,
		ObjectID:		promise.PromiseHash,
		ExpirationHeight:	promise.TimeoutHeight,
		CommitmentHash:		promise.PromiseHash,
		PromiseHash:		promise.PromiseHash,
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
		ConditionID:	promise.PromiseID,
		ConditionType:	promise.ConditionType,
		Payer:		promise.Source,
		Payee:		promise.Destination,
		Amount:		promise.Amount,
		HashLock:	promise.HashLock,
		TimeoutHeight:	promise.TimeoutHeight,
		NonceStart:	promise.Nonce,
		NonceEnd:	promise.Nonce,
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

func (s GossipSignature) Normalize() GossipSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = normalizeHash(s.ObjectID)
	s.CommitmentHash = normalizeHash(s.CommitmentHash)
	s.SignatureHash = normalizeHash(s.SignatureHash)
	return s
}

func (s GossipSignature) Validate(message GossipMessage) error {
	sig := s.Normalize()
	message = message.Normalize()
	if err := addressing.ValidateUserAddress("payments gossip signature signer", sig.Signer); err != nil {
		return err
	}
	if sig.ChainID != message.ChainID {
		return errors.New("payments gossip signature chain id mismatch")
	}
	if sig.ObjectType != SignatureObjectGossip {
		return errors.New("payments gossip signature object type mismatch")
	}
	if sig.Version != CurrentStateVersion {
		return errors.New("payments gossip signature version mismatch")
	}
	if sig.ObjectID != message.MessageID || sig.CommitmentHash != message.MessageID {
		return errors.New("payments gossip signature commitment mismatch")
	}
	if sig.ExpirationHeight != message.ValidUntilHeight {
		return errors.New("payments gossip signature expiration mismatch")
	}
	expected := ComputeSignatureEnvelopeHash(sig.Signer, sig.ChainID, "", sig.ObjectType, sig.Version, 0, sig.ObjectID, sig.ExpirationHeight, sig.CommitmentHash)
	if sig.SignatureHash != expected {
		return errors.New("payments gossip signature value mismatch")
	}
	return nil
}

func (m GossipMessage) Normalize() GossipMessage {
	m.MessageID = normalizeOptionalHash(m.MessageID)
	m.ChainID = strings.TrimSpace(m.ChainID)
	m.ChannelID = normalizeOptionalHash(m.ChannelID)
	m.NodeID = strings.TrimSpace(m.NodeID)
	m.From = strings.TrimSpace(m.From)
	m.To = strings.TrimSpace(m.To)
	m.Capacity = strings.TrimSpace(m.Capacity)
	m.Liquidity = strings.TrimSpace(m.Liquidity)
	m.FeeDenom = normalizeAssetDenom(m.FeeDenom)
	m.FeeAmount = strings.TrimSpace(m.FeeAmount)
	m.MaxFee = strings.TrimSpace(m.MaxFee)
	m.ChannelCommitment = normalizeOptionalHash(m.ChannelCommitment)
	m.FailureCode = strings.TrimSpace(m.FailureCode)
	m.ProbeAmount = strings.TrimSpace(m.ProbeAmount)
	if m.NodeID == "" {
		m.NodeID = m.From
	}
	if m.ValidUntilHeight == 0 && m.ValidAfterHeight > 0 {
		m.ValidUntilHeight = m.ValidAfterHeight + DefaultGossipTTL
	}
	return m
}

func (m GossipMessage) ValidateBasic() error {
	message := m.Normalize()
	if !IsGossipMessageType(message.MessageType) {
		return errors.New("payments gossip message type is invalid")
	}
	if strings.TrimSpace(message.ChainID) == "" {
		return errors.New("payments gossip chain id is required")
	}
	if message.ValidAfterHeight == 0 {
		return errors.New("payments gossip valid-after height must be positive")
	}
	if message.ValidUntilHeight <= message.ValidAfterHeight {
		return errors.New("payments gossip validity window must advance")
	}
	if err := addressing.ValidateUserAddress("payments gossip node", message.NodeID); err != nil {
		return err
	}
	switch message.MessageType {
	case GossipNodeAnnouncement:
		return nil
	case GossipRouteFailure:
		if message.FailureCode == "" {
			return errors.New("payments route failure code is required")
		}
		return validateGossipEdgeFields(message, false)
	case GossipCapacityProbe:
		if err := validatePositiveInt("payments capacity probe amount", message.ProbeAmount); err != nil {
			return err
		}
		return validateGossipEdgeFields(message, false)
	case GossipLiquidityHint:
		if err := validateNonNegativeInt("payments liquidity hint amount", message.Liquidity); err != nil {
			return err
		}
		return validateGossipEdgeFields(message, false)
	case GossipFeePolicyUpdate:
		if err := validateNonNegativeInt("payments fee policy max fee", message.MaxFee); err != nil {
			return err
		}
		return validateGossipEdgeFields(message, false)
	case GossipChannelAnnouncement, GossipChannelUpdate:
		return validateGossipEdgeFields(message, true)
	default:
		return errors.New("payments gossip message type is invalid")
	}
}

func (m GossipMessage) ToChannelEdge() (ChannelEdge, bool) {
	message := m.Normalize()
	switch message.MessageType {
	case GossipChannelAnnouncement, GossipChannelUpdate, GossipLiquidityHint, GossipFeePolicyUpdate:
		if message.ChannelID == "" {
			return ChannelEdge{}, false
		}
		capacity := message.Capacity
		if message.MessageType == GossipLiquidityHint && message.Liquidity != "" {
			capacity = message.Liquidity
		}
		if strings.TrimSpace(capacity) == "" {
			return ChannelEdge{}, false
		}
		edge := ChannelEdge{
			ChannelID:	message.ChannelID,
			From:		message.From,
			To:		message.To,
			Capacity:	capacity,
			FeeDenom:	message.FeeDenom,
			FeeAmount:	message.FeeAmount,
			ExpiresHeight:	message.ValidUntilHeight,
			Active:		true,
		}.Normalize()
		return edge, true
	default:
		return ChannelEdge{}, false
	}
}

func (e SignedGossipEnvelope) Normalize() SignedGossipEnvelope {
	e.Message = e.Message.Normalize()
	e.MessageHash = normalizeOptionalHash(e.MessageHash)
	e.Signature = e.Signature.Normalize()
	e.ReceivedFrom = strings.TrimSpace(e.ReceivedFrom)
	if e.ReceivedFrom == "" {
		e.ReceivedFrom = e.Signature.Signer
	}
	return e
}

func (e SignedGossipEnvelope) ValidateForState(state PaymentsState, currentHeight uint64) error {
	envelope := e.Normalize()
	if currentHeight == 0 {
		return errors.New("payments gossip validation height must be positive")
	}
	message, err := BuildGossipMessage(envelope.Message)
	if err != nil {
		return err
	}
	if envelope.MessageHash != "" && envelope.MessageHash != message.MessageID {
		return errors.New("payments gossip message hash mismatch")
	}
	envelope.Message = message
	if currentHeight < message.ValidAfterHeight {
		return errors.New("payments gossip message is not yet valid")
	}
	if currentHeight > message.ValidUntilHeight {
		return errors.New("payments gossip message is expired")
	}
	if err := envelope.Signature.Validate(message); err != nil {
		return err
	}
	if envelope.Signature.Signer != message.NodeID {
		return errors.New("payments gossip signer must match advertising node")
	}
	if message.MessageType == GossipChannelAnnouncement {
		if message.ChannelID == "" && message.ChannelCommitment == "" {
			return errors.New("payments channel announcement requires channel id or commitment")
		}
		if message.ChannelID != "" {
			channel, found := state.ChannelByID(message.ChannelID)
			if !found || channel.Status != ChannelStatusOpen {
				return errors.New("payments channel announcement requires open channel")
			}
		}
	}
	return nil
}

func (r GossipReputation) Normalize() GossipReputation {
	r.NodeID = strings.TrimSpace(r.NodeID)
	return r
}

func (r GossipReputation) Validate() error {
	r = r.Normalize()
	if err := addressing.ValidateUserAddress("payments gossip reputation node", r.NodeID); err != nil {
		return err
	}
	if r.LastUpdateHeight == 0 {
		return errors.New("payments gossip reputation update height must be positive")
	}
	return nil
}

func (s TopologyStore) Normalize() TopologyStore {
	for i, envelope := range s.Messages {
		s.Messages[i] = envelope.Normalize()
	}
	for i, edge := range s.Edges {
		s.Edges[i] = edge.Normalize()
	}
	for i, reputation := range s.Reputation {
		s.Reputation[i] = reputation.Normalize()
	}
	sortGossipEnvelopes(s.Messages)
	sortEdges(s.Edges)
	sortGossipReputation(s.Reputation)
	return s
}

func (s TopologyStore) Validate() error {
	store := s.Normalize()
	seenMessages := make(map[string]struct{}, len(store.Messages))
	for _, envelope := range store.Messages {
		id := envelope.Message.MessageID
		if id == "" {
			id = envelope.MessageHash
		}
		if err := ValidateHash("payments gossip store message id", id); err != nil {
			return err
		}
		if _, found := seenMessages[id]; found {
			return errors.New("payments duplicate gossip message")
		}
		seenMessages[id] = struct{}{}
	}
	for _, edge := range store.Edges {
		if err := edge.Validate(); err != nil {
			return err
		}
	}
	seenReputation := make(map[string]struct{}, len(store.Reputation))
	for _, reputation := range store.Reputation {
		if err := reputation.Validate(); err != nil {
			return err
		}
		if _, found := seenReputation[reputation.NodeID]; found {
			return errors.New("payments duplicate gossip reputation")
		}
		seenReputation[reputation.NodeID] = struct{}{}
	}
	return nil
}

func (s EdgeRoutingStats) Normalize() EdgeRoutingStats {
	s.ChannelID = normalizeHash(s.ChannelID)
	s.From = strings.TrimSpace(s.From)
	s.To = strings.TrimSpace(s.To)
	return s
}

func (s EdgeRoutingStats) Validate() error {
	stats := s.Normalize()
	if err := ValidateHash("payments route stats channel id", stats.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route stats from", stats.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route stats to", stats.To); err != nil {
		return err
	}
	if stats.SuccessRateBps > 10_000 || stats.CongestionBps > 10_000 || stats.NodeAvailabilityBps > 10_000 || stats.ReservePressureBps > 10_000 {
		return errors.New("payments route stats bps must be <= 10000")
	}
	return nil
}

func (r RouteFailureReport) Normalize() RouteFailureReport {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
	return r
}

func (r RouteFailureReport) Validate() error {
	report := r.Normalize()
	if err := ValidateHash("payments route failure channel id", report.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route failure from", report.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route failure to", report.To); err != nil {
		return err
	}
	if !IsRouteFailureClass(report.FailureClass) {
		return errors.New("payments route failure class is invalid")
	}
	if report.ObservedHeight == 0 {
		return errors.New("payments route failure observed height must be positive")
	}
	return nil
}

func (s CongestionSnapshot) Normalize() CongestionSnapshot {
	s.ChannelID = normalizeHash(s.ChannelID)
	s.From = strings.TrimSpace(s.From)
	s.To = strings.TrimSpace(s.To)
	return s
}

func (s CongestionSnapshot) Validate() error {
	snapshot := s.Normalize()
	if err := ValidateHash("payments congestion channel id", snapshot.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments congestion from", snapshot.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments congestion to", snapshot.To); err != nil {
		return err
	}
	if snapshot.ChannelUpdateFailureRateBps > 10_000 || snapshot.ReservePressureBps > 10_000 {
		return errors.New("payments congestion bps must be <= 10000")
	}
	if snapshot.ObservedHeight == 0 {
		return errors.New("payments congestion observed height must be positive")
	}
	return nil
}

func DefaultRoutePolicy() RoutePolicy {
	return RoutePolicy{
		MaxHops:			MaxRoutingHops,
		RequiredTimeoutMargin:		DefaultTimeoutMargin,
		StaleLiquidityAfter:		DefaultGossipTTL,
		HopPenalty:			"1",
		CongestionPenalty:		"10",
		StaleLiquidityPenalty:		"25",
		FailurePenalty:			"25",
		TimeoutPenalty:			"50",
		SuccessPenalty:			"25",
		AvailabilityPenalty:		"25",
		ReservePressurePenalty:		"25",
		QueueDelayPenalty:		"10",
		PendingConditionPenalty:	"5",
		LatencyPenalty:			"10",
		DecayHalfLife:			DefaultGossipTTL,
		MaxCongestedPaymentBps:		5_000,
		MaxSplits:			1,
	}
}

func (p RoutePolicy) Normalize() RoutePolicy {
	defaults := DefaultRoutePolicy()
	if p.MaxHops <= 0 || p.MaxHops > MaxRoutingHops {
		p.MaxHops = defaults.MaxHops
	}
	if p.RequiredTimeoutMargin == 0 {
		p.RequiredTimeoutMargin = defaults.RequiredTimeoutMargin
	}
	if p.StaleLiquidityAfter == 0 {
		p.StaleLiquidityAfter = defaults.StaleLiquidityAfter
	}
	if strings.TrimSpace(p.HopPenalty) == "" {
		p.HopPenalty = defaults.HopPenalty
	}
	if strings.TrimSpace(p.CongestionPenalty) == "" {
		p.CongestionPenalty = defaults.CongestionPenalty
	}
	if strings.TrimSpace(p.StaleLiquidityPenalty) == "" {
		p.StaleLiquidityPenalty = defaults.StaleLiquidityPenalty
	}
	if strings.TrimSpace(p.FailurePenalty) == "" {
		p.FailurePenalty = defaults.FailurePenalty
	}
	if strings.TrimSpace(p.TimeoutPenalty) == "" {
		p.TimeoutPenalty = defaults.TimeoutPenalty
	}
	if strings.TrimSpace(p.SuccessPenalty) == "" {
		p.SuccessPenalty = defaults.SuccessPenalty
	}
	if strings.TrimSpace(p.AvailabilityPenalty) == "" {
		p.AvailabilityPenalty = defaults.AvailabilityPenalty
	}
	if strings.TrimSpace(p.ReservePressurePenalty) == "" {
		p.ReservePressurePenalty = defaults.ReservePressurePenalty
	}
	if strings.TrimSpace(p.QueueDelayPenalty) == "" {
		p.QueueDelayPenalty = defaults.QueueDelayPenalty
	}
	if strings.TrimSpace(p.PendingConditionPenalty) == "" {
		p.PendingConditionPenalty = defaults.PendingConditionPenalty
	}
	if strings.TrimSpace(p.LatencyPenalty) == "" {
		p.LatencyPenalty = defaults.LatencyPenalty
	}
	if p.DecayHalfLife == 0 {
		p.DecayHalfLife = defaults.DecayHalfLife
	}
	if p.MaxCongestedPaymentBps == 0 {
		p.MaxCongestedPaymentBps = defaults.MaxCongestedPaymentBps
	}
	if strings.TrimSpace(p.MaxFeeAmount) != "" {
		p.MaxFeeAmount = strings.TrimSpace(p.MaxFeeAmount)
	}
	if p.MaxSplits <= 0 {
		p.MaxSplits = defaults.MaxSplits
	}
	if p.MaxSplits > MaxRoutingHops {
		p.MaxSplits = MaxRoutingHops
	}
	for i := range p.ExcludedNodes {
		p.ExcludedNodes[i] = strings.TrimSpace(p.ExcludedNodes[i])
	}
	sort.Strings(p.ExcludedNodes)
	for i := range p.ExcludedChannels {
		p.ExcludedChannels[i] = normalizeHash(p.ExcludedChannels[i])
	}
	sort.Strings(p.ExcludedChannels)
	for i := range p.EdgeStats {
		p.EdgeStats[i] = p.EdgeStats[i].Normalize()
	}
	sort.SliceStable(p.EdgeStats, func(i, j int) bool {
		return routeStatsKey(p.EdgeStats[i]) < routeStatsKey(p.EdgeStats[j])
	})
	return p
}

func (p RoutePolicy) Validate() error {
	policy := p.Normalize()
	for _, value := range []struct {
		field	string
		text	string
	}{
		{"payments route hop penalty", policy.HopPenalty},
		{"payments route congestion penalty", policy.CongestionPenalty},
		{"payments route stale liquidity penalty", policy.StaleLiquidityPenalty},
		{"payments route failure penalty", policy.FailurePenalty},
		{"payments route timeout penalty", policy.TimeoutPenalty},
		{"payments route success penalty", policy.SuccessPenalty},
		{"payments route availability penalty", policy.AvailabilityPenalty},
		{"payments route reserve pressure penalty", policy.ReservePressurePenalty},
		{"payments route queue delay penalty", policy.QueueDelayPenalty},
		{"payments route pending condition penalty", policy.PendingConditionPenalty},
		{"payments route latency penalty", policy.LatencyPenalty},
	} {
		if err := validateNonNegativeInt(value.field, value.text); err != nil {
			return err
		}
	}
	if policy.MaxCongestedPaymentBps > 10_000 {
		return errors.New("payments max congested payment bps must be <= 10000")
	}
	if policy.MaxFeeAmount != "" {
		if err := validateNonNegativeInt("payments route max fee", policy.MaxFeeAmount); err != nil {
			return err
		}
	}
	if policy.ProportionalFeeBps > 100_000 {
		return errors.New("payments route proportional fee bps is too high")
	}
	for _, node := range policy.ExcludedNodes {
		if err := addressing.ValidateUserAddress("payments route excluded node", node); err != nil {
			return err
		}
	}
	for _, channelID := range policy.ExcludedChannels {
		if err := ValidateHash("payments route excluded channel", channelID); err != nil {
			return err
		}
	}
	seenStats := make(map[string]struct{}, len(policy.EdgeStats))
	for _, stats := range policy.EdgeStats {
		if err := stats.Validate(); err != nil {
			return err
		}
		key := routeStatsKey(stats)
		if _, found := seenStats[key]; found {
			return errors.New("payments duplicate route stats")
		}
		seenStats[key] = struct{}{}
	}
	return nil
}

func BuildRoutingFeePolicyUpdate(update RoutingFeePolicyUpdate, signer string) (RoutingFeePolicyUpdate, error) {
	update = update.Normalize()
	if update.PolicyID == "" {
		update.PolicyID = HashParts("routing-fee-policy-id", update.ChainID, update.ChannelID, update.From, update.To, fmt.Sprintf("%020d", update.Sequence))
	}
	update.PolicyHash = ""
	update.Signature = RoutingFeePolicySignature{}
	update.PolicyHash = ComputeRoutingFeePolicyHash(update)
	signature, err := SignatureForRoutingFeePolicy(update, signer)
	if err != nil {
		return RoutingFeePolicyUpdate{}, err
	}
	update.Signature = signature
	if err := update.ValidateAtHeight(update.ValidAfterHeight); err != nil {
		return RoutingFeePolicyUpdate{}, err
	}
	return update.Normalize(), nil
}

func ComputeRoutingFeePolicyHash(update RoutingFeePolicyUpdate) string {
	update = update.Normalize()
	return HashParts(
		"routing-fee-policy",
		update.PolicyID,
		update.ChainID,
		update.ChannelID,
		update.From,
		update.To,
		update.FeeDenom,
		update.BaseHopFee,
		fmt.Sprintf("%010d", update.ProportionalFeeBps),
		update.LiquidityReservationFee,
		update.VirtualChannelSetupFee,
		update.CongestionSurcharge,
		update.FailurePenalty,
		update.MaxHopFee,
		fmt.Sprintf("%020d", update.ValidAfterHeight),
		fmt.Sprintf("%020d", update.ValidUntilHeight),
		fmt.Sprintf("%020d", update.Sequence),
	)
}

func SignatureForRoutingFeePolicy(update RoutingFeePolicyUpdate, signer string) (RoutingFeePolicySignature, error) {
	update = update.Normalize()
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments routing fee policy signer", signer); err != nil {
		return RoutingFeePolicySignature{}, err
	}
	if update.PolicyHash == "" {
		update.PolicyHash = ComputeRoutingFeePolicyHash(update)
	}
	return RoutingFeePolicySignature{
		Signer:			signer,
		ChainID:		update.ChainID,
		ChannelID:		update.ChannelID,
		ObjectType:		SignatureObjectRoutingFee,
		Version:		CurrentStateVersion,
		Sequence:		update.Sequence,
		ObjectID:		update.PolicyHash,
		ExpirationHeight:	update.ValidUntilHeight,
		CommitmentHash:		update.PolicyHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			update.ChainID,
			update.ChannelID,
			SignatureObjectRoutingFee,
			CurrentStateVersion,
			update.Sequence,
			update.PolicyHash,
			update.ValidUntilHeight,
			update.PolicyHash,
		),
	}, nil
}

func (u RoutingFeePolicyUpdate) Normalize() RoutingFeePolicyUpdate {
	u.PolicyID = normalizeOptionalHash(u.PolicyID)
	u.ChainID = strings.TrimSpace(u.ChainID)
	u.ChannelID = normalizeHash(u.ChannelID)
	u.From = strings.TrimSpace(u.From)
	u.To = strings.TrimSpace(u.To)
	u.FeeDenom = normalizeAssetDenom(u.FeeDenom)
	u.BaseHopFee = strings.TrimSpace(u.BaseHopFee)
	u.LiquidityReservationFee = strings.TrimSpace(u.LiquidityReservationFee)
	u.VirtualChannelSetupFee = strings.TrimSpace(u.VirtualChannelSetupFee)
	u.CongestionSurcharge = strings.TrimSpace(u.CongestionSurcharge)
	u.FailurePenalty = strings.TrimSpace(u.FailurePenalty)
	u.MaxHopFee = strings.TrimSpace(u.MaxHopFee)
	for _, field := range []*string{&u.BaseHopFee, &u.LiquidityReservationFee, &u.VirtualChannelSetupFee, &u.CongestionSurcharge, &u.FailurePenalty, &u.MaxHopFee} {
		if *field == "" {
			*field = "0"
		}
	}
	u.PolicyHash = normalizeOptionalHash(u.PolicyHash)
	u.Signature = u.Signature.Normalize()
	return u
}

func (u RoutingFeePolicyUpdate) ValidateAtHeight(currentHeight uint64) error {
	update := u.Normalize()
	if update.PolicyID == "" {
		return errors.New("payments routing fee policy id is required")
	}
	if err := ValidateHash("payments routing fee policy id", update.PolicyID); err != nil {
		return err
	}
	if update.ChainID == "" {
		return errors.New("payments routing fee policy chain id is required")
	}
	if err := ValidateHash("payments routing fee policy channel id", update.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments routing fee policy from", update.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments routing fee policy to", update.To); err != nil {
		return err
	}
	if update.From == update.To {
		return errors.New("payments routing fee policy endpoints must differ")
	}
	if update.FeeDenom != NativeDenom {
		return fmt.Errorf("payments routing fee policy denom must be %s", NativeDenom)
	}
	for _, value := range []struct {
		field	string
		text	string
	}{
		{"payments routing base hop fee", update.BaseHopFee},
		{"payments routing liquidity reservation fee", update.LiquidityReservationFee},
		{"payments routing virtual setup fee", update.VirtualChannelSetupFee},
		{"payments routing congestion surcharge", update.CongestionSurcharge},
		{"payments routing failure penalty", update.FailurePenalty},
		{"payments routing max hop fee", update.MaxHopFee},
	} {
		if err := validateNonNegativeInt(value.field, value.text); err != nil {
			return err
		}
	}
	if update.ProportionalFeeBps > 100_000 {
		return errors.New("payments routing proportional fee bps is too high")
	}
	if update.ValidAfterHeight == 0 || update.ValidUntilHeight == 0 || update.ValidUntilHeight < update.ValidAfterHeight {
		return errors.New("payments routing fee policy validity window is invalid")
	}
	if currentHeight != 0 && (currentHeight < update.ValidAfterHeight || currentHeight > update.ValidUntilHeight) {
		return errors.New("payments routing fee policy is outside validity window")
	}
	if update.Sequence == 0 {
		return errors.New("payments routing fee policy sequence must be positive")
	}
	expectedHash := ComputeRoutingFeePolicyHash(update)
	if update.PolicyHash != expectedHash {
		return errors.New("payments routing fee policy hash mismatch")
	}
	return update.Signature.Validate(update)
}

func (s RoutingFeePolicySignature) Normalize() RoutingFeePolicySignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = normalizeOptionalHash(s.ObjectID)
	s.CommitmentHash = normalizeOptionalHash(s.CommitmentHash)
	s.SignatureHash = normalizeOptionalHash(s.SignatureHash)
	return s
}

func (s RoutingFeePolicySignature) Validate(update RoutingFeePolicyUpdate) error {
	sig := s.Normalize()
	update = update.Normalize()
	if err := addressing.ValidateUserAddress("payments routing fee policy signature signer", sig.Signer); err != nil {
		return err
	}
	if sig.Signer != update.From {
		return errors.New("payments routing fee policy signer must be forwarding node")
	}
	if sig.ChainID != update.ChainID || sig.ChannelID != update.ChannelID {
		return errors.New("payments routing fee policy signature domain mismatch")
	}
	if sig.ObjectType != SignatureObjectRoutingFee {
		return errors.New("payments routing fee policy signature object type mismatch")
	}
	if sig.Version != CurrentStateVersion || sig.Sequence != update.Sequence {
		return errors.New("payments routing fee policy signature version or sequence mismatch")
	}
	if sig.ObjectID != update.PolicyHash || sig.CommitmentHash != update.PolicyHash {
		return errors.New("payments routing fee policy signature commitment mismatch")
	}
	if sig.ExpirationHeight != update.ValidUntilHeight {
		return errors.New("payments routing fee policy signature expiration mismatch")
	}
	if err := ValidateHash("payments routing fee policy signature hash", sig.SignatureHash); err != nil {
		return err
	}
	expected := ComputeSignatureEnvelopeHash(sig.Signer, sig.ChainID, sig.ChannelID, sig.ObjectType, sig.Version, sig.Sequence, sig.ObjectID, sig.ExpirationHeight, sig.CommitmentHash)
	if sig.SignatureHash != expected {
		return errors.New("payments routing fee policy signature value mismatch")
	}
	return nil
}

func (r RouteSelectionRequest) Normalize() RouteSelectionRequest {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
	r.Amount = strings.TrimSpace(r.Amount)
	r.Policy = r.Policy.Normalize()
	return r
}

func (r RouteSelectionRequest) Validate() error {
	req := r.Normalize()
	if err := addressing.ValidateUserAddress("payments route request from", req.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route request to", req.To); err != nil {
		return err
	}
	if req.From == req.To {
		return errors.New("payments route endpoints must differ")
	}
	if _, err := parsePositiveInt("payments route request amount", req.Amount); err != nil {
		return err
	}
	if req.CurrentHeight == 0 {
		return errors.New("payments route request height must be positive")
	}
	return req.Policy.Validate()
}

func (p RouteRetryPolicy) Normalize() RouteRetryPolicy {
	if p.MaxAttempts == 0 {
		p.MaxAttempts = 3
	}
	if p.AlternateRouteLimit == 0 {
		p.AlternateRouteLimit = p.MaxAttempts
	}
	return p
}

func (p RouteRetryPolicy) Validate() error {
	policy := p.Normalize()
	if policy.MaxAttempts == 0 || policy.MaxAttempts > 32 {
		return errors.New("payments route retry max attempts must be between 1 and 32")
	}
	if policy.AlternateRouteLimit == 0 || policy.AlternateRouteLimit > policy.MaxAttempts {
		return errors.New("payments route retry alternate limit must be within attempts")
	}
	return nil
}

func (r RouteRetryRequest) Normalize() RouteRetryRequest {
	r.Selection = r.Selection.Normalize()
	for i, failure := range r.Failures {
		r.Failures[i] = failure.Normalize()
	}
	sort.SliceStable(r.Failures, func(i, j int) bool {
		return routeFailureKey(r.Failures[i]) < routeFailureKey(r.Failures[j])
	})
	r.Policy = r.Policy.Normalize()
	return r
}

func (r RouteRetryRequest) Validate() error {
	req := r.Normalize()
	if err := req.Selection.Validate(); err != nil {
		return err
	}
	if err := req.Policy.Validate(); err != nil {
		return err
	}
	for _, failure := range req.Failures {
		if err := failure.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func IsRouteFailureClass(failureClass RouteFailureClass) bool {
	switch failureClass {
	case RouteFailureCapacity,
		RouteFailureTimeout,
		RouteFailureCongestion,
		RouteFailureLiquidityStale,
		RouteFailureNodeUnavailable,
		RouteFailurePolicyRejected,
		RouteFailureUnknown:
		return true
	default:
		return false
	}
}

func (r ScoredRoute) Normalize() ScoredRoute {
	for i, edge := range r.Edges {
		r.Edges[i] = edge.Normalize()
	}
	r.Amount = strings.TrimSpace(r.Amount)
	r.TotalFee = strings.TrimSpace(r.TotalFee)
	r.TotalCost = strings.TrimSpace(r.TotalCost)
	r.MinCapacity = strings.TrimSpace(r.MinCapacity)
	r.ScoreHash = normalizeOptionalHash(r.ScoreHash)
	return r
}

func (r ScoredRoute) Validate() error {
	route := r.Normalize()
	if len(route.Edges) == 0 {
		return errors.New("payments scored route requires edges")
	}
	if _, err := parsePositiveInt("payments scored route amount", route.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments scored route total fee", route.TotalFee); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments scored route total cost", route.TotalCost); err != nil {
		return err
	}
	if _, err := parsePositiveInt("payments scored route min capacity", route.MinCapacity); err != nil {
		return err
	}
	if route.ScoreHash != "" {
		return ValidateHash("payments scored route hash", route.ScoreHash)
	}
	return nil
}

func DeriveRouteID(seed string, nonce uint64) (string, error) {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		return "", errors.New("payments route id seed is required")
	}
	if nonce == 0 {
		return "", errors.New("payments route id nonce must be positive")
	}
	return HashParts("route-id", seed, fmt.Sprintf("%020d", nonce)), nil
}

func DeriveHopRouteID(routeID string, hopIndex int, channelID string) (string, error) {
	routeID = normalizeHash(routeID)
	channelID = normalizeHash(channelID)
	if err := ValidateHash("payments root route id", routeID); err != nil {
		return "", err
	}
	if hopIndex < 0 {
		return "", errors.New("payments hop index must be non-negative")
	}
	if err := ValidateHash("payments hop route channel id", channelID); err != nil {
		return "", err
	}
	return HashParts("hop-route-id", routeID, fmt.Sprintf("%020d", uint64(hopIndex)), channelID), nil
}

func DeriveHopPaymentID(routeID string, hopIndex int, channelID string) (string, error) {
	hopRouteID, err := DeriveHopRouteID(routeID, hopIndex, channelID)
	if err != nil {
		return "", err
	}
	return HashParts("hop-payment-id", hopRouteID), nil
}

func ComputeForwardingPacketHash(packet ForwardingPacket) string {
	packet = packet.Normalize()
	return HashParts(
		"forwarding-packet",
		packet.RouteID,
		packet.HopPaymentID,
		packet.ChannelID,
		packet.ForwardingNode,
		packet.NextNode,
		packet.Amount,
		packet.FeeAmount,
		fmt.Sprintf("%020d", packet.TimeoutHeight),
		packet.NextPacketHash,
	)
}

func (p ForwardingPacket) Normalize() ForwardingPacket {
	p.PacketID = normalizeOptionalHash(p.PacketID)
	p.RouteID = normalizeHash(p.RouteID)
	p.HopPaymentID = normalizeHash(p.HopPaymentID)
	p.ChannelID = normalizeHash(p.ChannelID)
	p.ForwardingNode = strings.TrimSpace(p.ForwardingNode)
	p.NextNode = strings.TrimSpace(p.NextNode)
	p.Amount = strings.TrimSpace(p.Amount)
	p.FeeAmount = strings.TrimSpace(p.FeeAmount)
	p.NextPacketHash = normalizeOptionalHash(p.NextPacketHash)
	p.PacketHash = normalizeOptionalHash(p.PacketHash)
	return p
}

func (p ForwardingPacket) Validate() error {
	packet := p.Normalize()
	if err := ValidateHash("payments forwarding packet id", packet.PacketID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding route id", packet.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding payment id", packet.HopPaymentID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding channel id", packet.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments forwarding node", packet.ForwardingNode); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments forwarding next node", packet.NextNode); err != nil {
		return err
	}
	if packet.ForwardingNode == packet.NextNode {
		return errors.New("payments forwarding packet nodes must differ")
	}
	if _, err := parsePositiveInt("payments forwarding amount", packet.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments forwarding fee", packet.FeeAmount); err != nil {
		return err
	}
	if packet.TimeoutHeight == 0 {
		return errors.New("payments forwarding timeout height must be positive")
	}
	if packet.NextPacketHash != "" {
		if err := ValidateHash("payments forwarding next packet hash", packet.NextPacketHash); err != nil {
			return err
		}
	}
	if packet.PacketHash != ComputeForwardingPacketHash(packet) {
		return errors.New("payments forwarding packet hash mismatch")
	}
	return nil
}

func (r ForwardingPacketReplayRecord) Normalize() ForwardingPacketReplayRecord {
	r.PacketID = normalizeHash(r.PacketID)
	r.RouteID = normalizeHash(r.RouteID)
	r.HopPaymentID = normalizeHash(r.HopPaymentID)
	return r
}

func (r ForwardingPacketReplayRecord) Validate() error {
	record := r.Normalize()
	if err := ValidateHash("payments forwarding replay packet id", record.PacketID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding replay route id", record.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding replay payment id", record.HopPaymentID); err != nil {
		return err
	}
	if record.RecordedHeight == 0 {
		return errors.New("payments forwarding replay recorded height must be positive")
	}
	if record.ExpiresHeight <= record.RecordedHeight {
		return errors.New("payments forwarding replay expiry must exceed recorded height")
	}
	return nil
}

func (r ForwardingLogRecord) Normalize() ForwardingLogRecord {
	r.PacketID = normalizeHash(r.PacketID)
	r.RouteID = normalizeHash(r.RouteID)
	r.HopPaymentID = normalizeHash(r.HopPaymentID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.ForwardingNode = strings.TrimSpace(r.ForwardingNode)
	r.NextNodeHash = normalizeHash(r.NextNodeHash)
	r.AmountHash = normalizeHash(r.AmountHash)
	return r
}

func (r ForwardingLogRecord) Validate() error {
	record := r.Normalize()
	if err := ValidateHash("payments forwarding log packet id", record.PacketID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding log route id", record.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding log payment id", record.HopPaymentID); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding log channel id", record.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments forwarding log node", record.ForwardingNode); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding log next node hash", record.NextNodeHash); err != nil {
		return err
	}
	if err := ValidateHash("payments forwarding log amount hash", record.AmountHash); err != nil {
		return err
	}
	if record.RecordedHeight == 0 {
		return errors.New("payments forwarding log height must be positive")
	}
	return nil
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
		ChannelID:		channel.ChannelID,
		StateHash:		req.State.StateHash,
		Nonce:			req.State.Nonce,
		ValidatedOffChain:	true,
		Liquidity:		req.State.Balances,
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
		Signer:			signer,
		ChainID:		claim.ChainID,
		ChannelID:		claim.ChannelID,
		ObjectType:		SignatureObjectClaim,
		Version:		CurrentStateVersion,
		Nonce:			claim.Nonce,
		ObjectID:		claim.StateHash,
		ExpirationHeight:	claim.ExpirationHeight,
		CommitmentHash:		claim.StateHash,
		ClaimHash:		claim.StateHash,
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

func BuildGossipMessage(message GossipMessage) (GossipMessage, error) {
	message = message.Normalize()
	if err := message.ValidateBasic(); err != nil {
		return GossipMessage{}, err
	}
	message.MessageID = ComputeGossipMessageHash(message)
	return message, nil
}

func SignatureForGossip(message GossipMessage, signer string) (GossipSignature, error) {
	message = message.Normalize()
	if message.MessageID == "" {
		var err error
		message, err = BuildGossipMessage(message)
		if err != nil {
			return GossipSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments gossip signer", signer); err != nil {
		return GossipSignature{}, err
	}
	return GossipSignature{
		Signer:			signer,
		ChainID:		message.ChainID,
		ObjectType:		SignatureObjectGossip,
		Version:		CurrentStateVersion,
		ObjectID:		message.MessageID,
		ExpirationHeight:	message.ValidUntilHeight,
		CommitmentHash:		message.MessageID,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			message.ChainID,
			"",
			SignatureObjectGossip,
			CurrentStateVersion,
			0,
			message.MessageID,
			message.ValidUntilHeight,
			message.MessageID,
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
		Signer:			signer,
		ChainID:		delta.ChainID,
		ChannelID:		delta.ChannelID,
		ObjectType:		SignatureObjectDelta,
		Version:		CurrentStateVersion,
		Nonce:			delta.NonceStart,
		ObjectID:		delta.UpdateID,
		ExpirationHeight:	delta.ExpiryHeight,
		CommitmentHash:		delta.DeltaHash,
		DeltaHash:		delta.DeltaHash,
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
		ChainID:		channel.ChainID,
		AppVersion:		CurrentAppVersion,
		ModuleName:		ModuleName,
		ChannelID:		channel.ChannelID,
		ChannelType:		ChannelTypeAsync,
		ParticipantSetHash:	ComputeParticipantSetHash(channel.Participants),
		Denom:			channel.Denom,
		Version:		CurrentStateVersion,
		Epoch:			base.Epoch,
		Nonce:			checkpointNonce,
		Balances:		nextBalances,
		CheckpointNonce:	checkpointNonce,
		CheckpointBalances:	nextBalances,
		AsyncUpdateRoot:	ComputeAsyncDeltaRootForChannel(channel, normalizedDeltas),
		AcceptedUpdateRoot:	ComputeAsyncDeltaRootForChannel(channel, normalizedDeltas),
		SendWindow:		base.SendWindow,
		ReceiveWindow:		base.ReceiveWindow,
		MaxUnackedAmount:	base.MaxUnackedAmount,
		ExpiryHeight:		base.ExpiryHeight,
		TimeoutHeight:		base.TimeoutHeight,
		TimeoutTimestamp:	base.TimeoutTimestamp,
		ChallengePeriod:	base.ChallengePeriod,
		CloseDelay:		base.CloseDelay,
		FeePolicyID:		NativeDenom,
		RequiredSignerBitmap:	ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners),
		SignatureScheme:	SignatureSchemeEd25519,
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
		ChainID:		channel.ChainID,
		ChannelID:		channel.ChannelID,
		Payer:			channel.Payer,
		Receiver:		channel.Receiver,
		LockedAmount:		channel.Collateral,
		ClaimedAmount:		claimed.String(),
		Nonce:			frame.Nonce,
		ExpirationHeight:	frame.ExpirationHeight,
		ExpirationTimestamp:	frame.ExpirationTimestamp,
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

func (j AsyncFinalizationJob) Normalize() AsyncFinalizationJob {
	j.JobID = normalizeOptionalHash(j.JobID)
	j.ChannelID = normalizeHash(j.ChannelID)
	j.LastError = strings.TrimSpace(j.LastError)
	j.SettlementHash = normalizeOptionalHash(j.SettlementHash)
	return j
}

func (j AsyncFinalizationJob) Validate() error {
	job := j.Normalize()
	if err := ValidateHash("payments async finalization job id", job.JobID); err != nil {
		return err
	}
	if err := ValidateHash("payments async finalization channel id", job.ChannelID); err != nil {
		return err
	}
	if job.FinalizeHeight == 0 || job.EnqueuedHeight == 0 {
		return errors.New("payments async finalization heights must be positive")
	}
	if job.Completed {
		if job.CompletedHeight == 0 {
			return errors.New("payments async finalization completed height must be positive")
		}
		if err := ValidateHash("payments async finalization settlement hash", job.SettlementHash); err != nil {
			return err
		}
	}
	return nil
}

func (j AsyncPromiseExpiryJob) Normalize() AsyncPromiseExpiryJob {
	j.JobID = normalizeOptionalHash(j.JobID)
	j.ChannelID = normalizeHash(j.ChannelID)
	j.PromiseID = normalizeHash(j.PromiseID)
	j.Promise = j.Promise.Normalize()
	j.Resolver = strings.TrimSpace(j.Resolver)
	j.LastError = strings.TrimSpace(j.LastError)
	j.ResolutionHash = normalizeOptionalHash(j.ResolutionHash)
	return j
}

func (j AsyncPromiseExpiryJob) Validate() error {
	job := j.Normalize()
	if err := ValidateHash("payments async promise expiry job id", job.JobID); err != nil {
		return err
	}
	if err := ValidateHash("payments async promise expiry channel id", job.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments async promise id", job.PromiseID); err != nil {
		return err
	}
	if job.Promise.ChannelID != job.ChannelID || job.Promise.PromiseID != job.PromiseID {
		return errors.New("payments async promise expiry job promise mismatch")
	}
	if err := addressing.ValidateUserAddress("payments async promise resolver", job.Resolver); err != nil {
		return err
	}
	if job.ExpireAfterHeight == 0 || job.EnqueuedHeight == 0 {
		return errors.New("payments async promise expiry heights must be positive")
	}
	if job.Completed {
		if job.CompletedHeight == 0 {
			return errors.New("payments async promise expiry completed height must be positive")
		}
		if err := ValidateHash("payments async promise expiry resolution hash", job.ResolutionHash); err != nil {
			return err
		}
	}
	return nil
}

func (c AsyncSettlementCompletion) Normalize() AsyncSettlementCompletion {
	c.CompletionID = normalizeOptionalHash(c.CompletionID)
	c.JobID = normalizeHash(c.JobID)
	c.JobType = strings.TrimSpace(c.JobType)
	c.ChannelID = normalizeHash(c.ChannelID)
	c.ObjectID = strings.TrimSpace(c.ObjectID)
	c.ResultHash = normalizeHash(c.ResultHash)
	return c
}

func (c AsyncSettlementCompletion) Validate() error {
	completion := c.Normalize()
	if err := ValidateHash("payments async completion id", completion.CompletionID); err != nil {
		return err
	}
	if err := ValidateHash("payments async completion job id", completion.JobID); err != nil {
		return err
	}
	if completion.JobType == "" {
		return errors.New("payments async completion job type is required")
	}
	if err := ValidateHash("payments async completion channel id", completion.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments async completion result hash", completion.ResultHash); err != nil {
		return err
	}
	if completion.Height == 0 {
		return errors.New("payments async completion height must be positive")
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

func (p ConditionLinkageProof) Normalize() ConditionLinkageProof {
	p.RouteID = normalizeHash(p.RouteID)
	p.Promises = normalizePromiseRoute(p.Promises)
	p.Sender = strings.TrimSpace(p.Sender)
	p.Receiver = strings.TrimSpace(p.Receiver)
	p.Amount = strings.TrimSpace(p.Amount)
	p.TotalFees = strings.TrimSpace(p.TotalFees)
	p.HashLock = normalizeHash(p.HashLock)
	p.EvidenceHash = normalizeOptionalHash(p.EvidenceHash)
	for i, id := range p.OffchainResolvedPromiseIDs {
		p.OffchainResolvedPromiseIDs[i] = normalizeHash(id)
	}
	sort.Strings(p.OffchainResolvedPromiseIDs)
	return p
}

func (p ConditionLinkageProof) ValidateForState(state PaymentsState, settledClaims []ConditionClaimRecord) error {
	proof := p.Normalize()
	state = state.Export()
	if err := ValidateHash("payments condition linkage route id", proof.RouteID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments condition linkage sender", proof.Sender); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments condition linkage receiver", proof.Receiver); err != nil {
		return err
	}
	if proof.Sender == proof.Receiver {
		return errors.New("payments condition linkage endpoints must differ")
	}
	if err := validatePositiveInt("payments condition linkage amount", proof.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments condition linkage total fees", proof.TotalFees); err != nil {
		return err
	}
	if err := ValidateHash("payments condition linkage hash lock", proof.HashLock); err != nil {
		return err
	}
	if proof.EvidenceHash != "" {
		if err := ValidateHash("payments condition linkage evidence hash", proof.EvidenceHash); err != nil {
			return err
		}
	}
	if len(proof.Promises) == 0 {
		return errors.New("payments condition linkage requires promises")
	}
	if len(proof.Promises) < 2 && !proof.PartialDispute {
		return errors.New("payments condition linkage requires at least two promises")
	}
	if proof.TimeoutMargin == 0 {
		proof.TimeoutMargin = DefaultTimeoutMargin
	}
	seen := make(map[string]struct{}, len(proof.Promises)+len(proof.OffchainResolvedPromiseIDs))
	for _, id := range proof.OffchainResolvedPromiseIDs {
		if err := ValidateHash("payments offchain resolved promise id", id); err != nil {
			return err
		}
		if _, found := seen[id]; found {
			return errors.New("payments duplicate offchain resolved promise id")
		}
		seen[id] = struct{}{}
	}
	if !proof.PartialDispute && len(proof.OffchainResolvedPromiseIDs) > 0 {
		return errors.New("payments offchain resolved promises require partial dispute proof")
	}
	if proof.PartialDispute && len(proof.OffchainResolvedPromiseIDs) == 0 {
		return errors.New("payments partial dispute requires offchain resolved promise ids")
	}
	channels := make([]ChannelRecord, len(proof.Promises))
	for i, promise := range proof.Promises {
		if _, found := seen[promise.PromiseID]; found {
			return errors.New("payments duplicate linked promise id")
		}
		seen[promise.PromiseID] = struct{}{}
		channel, found := state.ChannelByID(promise.ChannelID)
		if !found {
			return errors.New("payments linked promise channel not found")
		}
		channel = channel.Normalize()
		if channel.Status != ChannelStatusOpen {
			return errors.New("payments linked promise channel must be open")
		}
		if !channel.ConditionalPayments {
			return errors.New("payments linked promise channel must support conditions")
		}
		if err := promise.ValidateForChannel(channel); err != nil {
			return err
		}
		if promise.ConditionType != ConditionTypeHashLock {
			return errors.New("payments linked promise must be hash-lock")
		}
		if promise.HashLock != proof.HashLock {
			return errors.New("payments linked promises must share hash lock")
		}
		if promise.RouteIDOptional != "" && promise.RouteIDOptional != proof.RouteID {
			return errors.New("payments linked promise route id mismatch")
		}
		if promiseWasSettled(channel, promise.PromiseID, settledClaims) {
			return errors.New("payments linked promise has already been settled")
		}
		channels[i] = channel
	}
	if proof.Promises[0].Source != proof.Sender {
		return errors.New("payments linked route sender mismatch")
	}
	if proof.Promises[len(proof.Promises)-1].Destination != proof.Receiver {
		return errors.New("payments linked route receiver mismatch")
	}
	if err := validateLinkedPromiseConservation(proof.Promises, proof.Amount, proof.TotalFees); err != nil {
		return err
	}
	for i := 0; i < len(proof.Promises)-1; i++ {
		upstream := proof.Promises[i]
		downstream := proof.Promises[i+1]
		if upstream.Destination != downstream.Source {
			return errors.New("payments linked promise hop mismatch")
		}
		if upstream.NextPromiseIDOptional != "" && upstream.NextPromiseIDOptional != downstream.PromiseID {
			return errors.New("payments linked promise next id mismatch")
		}
		if downstream.PreviousPromiseIDOptional != "" && downstream.PreviousPromiseIDOptional != upstream.PromiseID {
			return errors.New("payments linked promise previous id mismatch")
		}
		if err := ValidateCrossChannelPromiseTimeoutOrdering(channels[i], channels[i+1], upstream, downstream, proof.TimeoutMargin); err != nil {
			return err
		}
	}
	return nil
}

func (c RouteFeeClaim) Normalize() RouteFeeClaim {
	c.ChannelID = normalizeHash(c.ChannelID)
	c.PromiseID = normalizeHash(c.PromiseID)
	c.Recipient = strings.TrimSpace(c.Recipient)
	c.Amount = strings.TrimSpace(c.Amount)
	c.EvidenceHash = normalizeHash(c.EvidenceHash)
	return c
}

func (c RouteFeeClaim) Validate() error {
	c = c.Normalize()
	if err := ValidateHash("payments route fee claim channel id", c.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments route fee claim promise id", c.PromiseID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route fee recipient", c.Recipient); err != nil {
		return err
	}
	if err := validatePositiveInt("payments route fee amount", c.Amount); err != nil {
		return err
	}
	return ValidateHash("payments route fee evidence hash", c.EvidenceHash)
}

func (r BatchConditionSettlementRequest) Normalize() BatchConditionSettlementRequest {
	r.LinkageProof = r.LinkageProof.Normalize()
	r.Preimage = strings.TrimSpace(r.Preimage)
	r.Resolver = strings.TrimSpace(r.Resolver)
	r.SettlementFeePaid = strings.TrimSpace(r.SettlementFeePaid)
	if r.SettlementFeePaid == "" {
		r.SettlementFeePaid = "0"
	}
	return r
}

func (r BatchConditionSettlementRequest) ValidateForState(state PaymentsState, settledClaims []ConditionClaimRecord) error {
	req := r.Normalize()
	if req.Mode != ConditionSettlementModePreimage && req.Mode != ConditionSettlementModeExpiry {
		return errors.New("payments batch condition settlement mode is invalid")
	}
	if err := addressing.ValidateUserAddress("payments batch condition resolver", req.Resolver); err != nil {
		return err
	}
	if req.CurrentHeight == 0 {
		return errors.New("payments batch condition settlement height must be positive")
	}
	if err := req.LinkageProof.ValidateForState(state, settledClaims); err != nil {
		return err
	}
	proof := req.LinkageProof.Normalize()
	if req.Mode == ConditionSettlementModePreimage && req.Resolver != proof.Receiver {
		return errors.New("payments batch preimage resolver must be route receiver")
	}
	for _, promise := range req.LinkageProof.Normalize().Promises {
		if req.Mode == ConditionSettlementModePreimage {
			if req.CurrentHeight > promise.TimeoutHeight {
				return errors.New("payments batch preimage promise has timed out")
			}
			if err := VerifyPromisePreimage(promise, req.Preimage); err != nil {
				return err
			}
			continue
		}
		if req.CurrentHeight <= promise.TimeoutHeight {
			return errors.New("payments batch expiry promise has not expired")
		}
	}
	return nil
}

func (r BatchConditionSettlementResult) Normalize() BatchConditionSettlementResult {
	r.RouteID = normalizeHash(r.RouteID)
	r.Resolutions = normalizeConditionResolutions(r.Resolutions)
	r.ConditionRootUpdates = normalizeConditionRootUpdates(r.ConditionRootUpdates)
	for i, claim := range r.FeeClaims {
		r.FeeClaims[i] = claim.Normalize()
	}
	sort.SliceStable(r.FeeClaims, func(i, j int) bool {
		if r.FeeClaims[i].ChannelID == r.FeeClaims[j].ChannelID {
			return r.FeeClaims[i].PromiseID < r.FeeClaims[j].PromiseID
		}
		return r.FeeClaims[i].ChannelID < r.FeeClaims[j].ChannelID
	})
	r.EvidenceHash = normalizeHash(r.EvidenceHash)
	return r
}

func (r BatchConditionSettlementResult) Validate() error {
	result := r.Normalize()
	if err := ValidateHash("payments batch condition result route id", result.RouteID); err != nil {
		return err
	}
	if len(result.Resolutions) == 0 {
		return errors.New("payments batch condition result requires resolutions")
	}
	for _, claim := range result.FeeClaims {
		if err := claim.Validate(); err != nil {
			return err
		}
	}
	return ValidateHash("payments batch condition result evidence hash", result.EvidenceHash)
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
	f.VerificationFeePaid = strings.TrimSpace(f.VerificationFeePaid)
	if f.VerificationFeePaid == "" {
		f.VerificationFeePaid = "0"
	}
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
	if err := validateNonNegativeInt("payments fraud verification fee", proof.VerificationFeePaid); err != nil {
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
	if !IsPenaltyRoute(allocation.Route) || allocation.Route == PenaltyRouteReporter || allocation.Route == PenaltyRouteCounterparty {
		return errors.New("payments penalty allocation route must be burn, security reserve, or community pool")
	}
	if allocation.Denom != NativeDenom {
		return fmt.Errorf("payments penalty allocation denom must be %s", NativeDenom)
	}
	return validatePositiveInt("payments penalty allocation amount", allocation.Amount)
}

func (p FraudPenaltyPolicy) Normalize() FraudPenaltyPolicy {
	p.ReporterRewardCap = strings.TrimSpace(p.ReporterRewardCap)
	p.CounterpartyRewardCap = strings.TrimSpace(p.CounterpartyRewardCap)
	return p
}

func (p FraudPenaltyPolicy) Validate() error {
	p = p.Normalize()
	if p.ReporterRewardCap != "" {
		if err := validateNonNegativeInt("payments reporter reward cap", p.ReporterRewardCap); err != nil {
			return err
		}
	}
	if p.CounterpartyRewardCap != "" {
		if err := validateNonNegativeInt("payments counterparty reward cap", p.CounterpartyRewardCap); err != nil {
			return err
		}
	}
	if p.CounterpartyRewardBps > MaxPenaltyRouteBps {
		return errors.New("payments counterparty reward bps exceeds 10000")
	}
	total := p.BurnShareBps + p.SecurityReserveShareBps + p.CommunityPoolShareBps
	if total > MaxPenaltyRouteBps {
		return errors.New("payments penalty route shares exceed 10000 bps")
	}
	return nil
}

func (e PenaltyMatrixEntry) Normalize() PenaltyMatrixEntry {
	e.BasePenalty = strings.TrimSpace(e.BasePenalty)
	e.ReporterRewardCap = strings.TrimSpace(e.ReporterRewardCap)
	e.CounterpartyCompensation = strings.TrimSpace(e.CounterpartyCompensation)
	e.InvalidProofVerifierCost = strings.TrimSpace(e.InvalidProofVerifierCost)
	if e.BasePenalty == "" {
		e.BasePenalty = "0"
	}
	if e.ReporterRewardCap == "" {
		e.ReporterRewardCap = "0"
	}
	if e.CounterpartyCompensation == "" {
		e.CounterpartyCompensation = "0"
	}
	if e.InvalidProofVerifierCost == "" {
		e.InvalidProofVerifierCost = "0"
	}
	return e
}

func (e PenaltyMatrixEntry) Validate() error {
	entry := e.Normalize()
	if !IsPaymentPenaltyClass(entry.Class) {
		return fmt.Errorf("unknown payments penalty class %q", entry.Class)
	}
	if entry.ProofType != "" && !IsFraudProofType(entry.ProofType) {
		return fmt.Errorf("unknown payments penalty matrix proof type %q", entry.ProofType)
	}
	if !IsPenaltySource(entry.Source) {
		return fmt.Errorf("unknown payments penalty source %q", entry.Source)
	}
	for _, item := range []struct {
		name	string
		amount	string
	}{
		{"payments penalty matrix base", entry.BasePenalty},
		{"payments penalty reporter cap", entry.ReporterRewardCap},
		{"payments penalty counterparty compensation", entry.CounterpartyCompensation},
		{"payments invalid proof verifier cost", entry.InvalidProofVerifierCost},
	} {
		if err := validateNonNegativeInt(item.name, item.amount); err != nil {
			return err
		}
	}
	if entry.BurnShareBps+entry.SecurityReserveShareBps+entry.CommunityPoolShareBps > MaxPenaltyRouteBps {
		return errors.New("payments penalty matrix route shares exceed 10000 bps")
	}
	return nil
}

func (s PaymentFeeSchedule) Normalize() PaymentFeeSchedule {
	defaults := DefaultPaymentFeeSchedule()
	s.Denom = normalizeAssetDenom(s.Denom)
	if s.Denom == "" {
		s.Denom = defaults.Denom
	}
	fields := []*string{
		&s.ChannelOpenFee,
		&s.ChannelOpenPerParticipantFee,
		&s.ChannelCheckpointFee,
		&s.CooperativeCloseFee,
		&s.UnilateralCloseFee,
		&s.DisputeFee,
		&s.FraudProofVerificationFee,
		&s.ConditionalPromiseSettlementFee,
		&s.VirtualChannelAnchorFee,
		&s.RoutingAdvertisementFee,
		&s.RoutingAdvertisementDeposit,
		&s.ConditionalCapabilitySurcharge,
		&s.VirtualChannelAnchorSurcharge,
		&s.StorageByteFee,
		&s.OpenFeeMin,
		&s.OpenFeeMax,
		&s.StorageRentPerBlock,
	}
	for _, field := range fields {
		*field = strings.TrimSpace(*field)
		if *field == "" {
			*field = "0"
		}
	}
	if s.ChannelOpenFee == "0" {
		s.ChannelOpenFee = defaults.ChannelOpenFee
	}
	if s.OpenFeeMin == "0" {
		s.OpenFeeMin = defaults.OpenFeeMin
	}
	if s.BaseMultiplierBps == 0 {
		s.BaseMultiplierBps = defaults.BaseMultiplierBps
	}
	if s.MaxMultiplierBps == 0 {
		s.MaxMultiplierBps = defaults.MaxMultiplierBps
	}
	return s
}

func (s PaymentFeeSchedule) Validate() error {
	s = s.Normalize()
	if s.Denom != NativeDenom {
		return fmt.Errorf("payments fee schedule denom must be %s", NativeDenom)
	}
	for _, item := range []struct {
		name	string
		amount	string
	}{
		{"payments channel open fee", s.ChannelOpenFee},
		{"payments channel open per participant fee", s.ChannelOpenPerParticipantFee},
		{"payments checkpoint fee", s.ChannelCheckpointFee},
		{"payments cooperative close fee", s.CooperativeCloseFee},
		{"payments unilateral close fee", s.UnilateralCloseFee},
		{"payments dispute fee", s.DisputeFee},
		{"payments fraud proof verification fee", s.FraudProofVerificationFee},
		{"payments conditional promise settlement fee", s.ConditionalPromiseSettlementFee},
		{"payments virtual channel anchor fee", s.VirtualChannelAnchorFee},
		{"payments routing advertisement fee", s.RoutingAdvertisementFee},
		{"payments routing advertisement deposit", s.RoutingAdvertisementDeposit},
		{"payments conditional capability surcharge", s.ConditionalCapabilitySurcharge},
		{"payments virtual channel anchor surcharge", s.VirtualChannelAnchorSurcharge},
		{"payments storage byte fee", s.StorageByteFee},
		{"payments open fee minimum", s.OpenFeeMin},
		{"payments open fee maximum", s.OpenFeeMax},
		{"payments storage rent per block", s.StorageRentPerBlock},
	} {
		if err := validateNonNegativeInt(item.name, item.amount); err != nil {
			return err
		}
	}
	if s.OpenFeeMax != "0" {
		minFee, err := parseNonNegativeInt("payments open fee minimum", s.OpenFeeMin)
		if err != nil {
			return err
		}
		maxFee, err := parseNonNegativeInt("payments open fee maximum", s.OpenFeeMax)
		if err != nil {
			return err
		}
		if maxFee.LT(minFee) {
			return errors.New("payments open fee maximum cannot be below minimum")
		}
	}
	if s.BaseMultiplierBps == 0 || s.BaseMultiplierBps > s.MaxMultiplierBps {
		return errors.New("payments fee multiplier bounds are invalid")
	}
	return nil
}

func (m PaymentFeeMultiplier) Normalize() PaymentFeeMultiplier {
	return m
}

func (m PaymentFeeMultiplier) Validate() error {
	if !IsPaymentFeeClass(m.FeeClass) {
		return fmt.Errorf("unknown payments fee class %q", m.FeeClass)
	}
	if m.MultiplierBps == 0 {
		return errors.New("payments fee multiplier must be positive")
	}
	if m.UpdatedHeight == 0 {
		return errors.New("payments fee multiplier height must be positive")
	}
	return nil
}

func (c PaymentFeeCharge) Normalize() PaymentFeeCharge {
	c.FeeID = normalizeOptionalHash(c.FeeID)
	c.ChannelID = normalizeOptionalHash(c.ChannelID)
	c.ObjectID = strings.TrimSpace(c.ObjectID)
	c.Payer = strings.TrimSpace(c.Payer)
	c.Denom = normalizeAssetDenom(c.Denom)
	c.Amount = strings.TrimSpace(c.Amount)
	c.RequiredAmount = strings.TrimSpace(c.RequiredAmount)
	return c
}

func (c PaymentFeeCharge) Validate() error {
	c = c.Normalize()
	if err := ValidateHash("payments fee id", c.FeeID); err != nil {
		return err
	}
	if !IsPaymentFeeClass(c.FeeClass) {
		return fmt.Errorf("unknown payments fee class %q", c.FeeClass)
	}
	if c.ChannelID != "" {
		if err := ValidateHash("payments fee channel id", c.ChannelID); err != nil {
			return err
		}
	}
	if err := addressing.ValidateUserAddress("payments fee payer", c.Payer); err != nil {
		return err
	}
	if c.Denom != NativeDenom {
		return fmt.Errorf("payments fee denom must be %s", NativeDenom)
	}
	if err := validateNonNegativeInt("payments fee amount", c.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments required fee amount", c.RequiredAmount); err != nil {
		return err
	}
	paid, err := parseNonNegativeInt("payments fee amount", c.Amount)
	if err != nil {
		return err
	}
	required, err := parseNonNegativeInt("payments required fee amount", c.RequiredAmount)
	if err != nil {
		return err
	}
	if paid.LT(required) {
		return errors.New("payments fee charge is below required amount")
	}
	if c.MultiplierBps == 0 {
		return errors.New("payments fee charge multiplier must be positive")
	}
	if c.Height == 0 {
		return errors.New("payments fee charge height must be positive")
	}
	return nil
}

func (r PaymentFeeRefund) Normalize() PaymentFeeRefund {
	r.RefundID = normalizeOptionalHash(r.RefundID)
	r.FeeID = normalizeOptionalHash(r.FeeID)
	r.Recipient = strings.TrimSpace(r.Recipient)
	r.Denom = normalizeAssetDenom(r.Denom)
	r.Amount = strings.TrimSpace(r.Amount)
	r.Reason = strings.TrimSpace(r.Reason)
	return r
}

func (r PaymentFeeRefund) Validate() error {
	r = r.Normalize()
	if err := ValidateHash("payments fee refund id", r.RefundID); err != nil {
		return err
	}
	if err := ValidateHash("payments refunded fee id", r.FeeID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments fee refund recipient", r.Recipient); err != nil {
		return err
	}
	if r.Denom != NativeDenom {
		return fmt.Errorf("payments fee refund denom must be %s", NativeDenom)
	}
	if err := validatePositiveInt("payments fee refund amount", r.Amount); err != nil {
		return err
	}
	if r.Reason == "" {
		return errors.New("payments fee refund reason is required")
	}
	if r.Height == 0 {
		return errors.New("payments fee refund height must be positive")
	}
	return nil
}

func DefaultSettlementGasCostSchedule() SettlementGasCostSchedule {
	return SettlementGasCostSchedule{
		OpenGas:			30_000,
		CooperativeCloseGas:		22_000,
		UnilateralCloseGas:		35_000,
		DisputeGas:			45_000,
		FraudProofGas:			60_000,
		ConditionResolutionGas:		40_000,
		PenaltyRoutingGas:		20_000,
		FinalSettlementGas:		50_000,
		ReplayProtectionGas:		10_000,
		PerSignatureGas:		2_000,
		PerConditionGas:		3_000,
		PerFraudProofGas:		8_000,
		PerPenaltyAllocationGas:	1_500,
		PerStateByteGas:		8,
	}
}

func (s SettlementGasCostSchedule) Normalize() SettlementGasCostSchedule {
	defaults := DefaultSettlementGasCostSchedule()
	if s.OpenGas == 0 {
		s.OpenGas = defaults.OpenGas
	}
	if s.CooperativeCloseGas == 0 {
		s.CooperativeCloseGas = defaults.CooperativeCloseGas
	}
	if s.UnilateralCloseGas == 0 {
		s.UnilateralCloseGas = defaults.UnilateralCloseGas
	}
	if s.DisputeGas == 0 {
		s.DisputeGas = defaults.DisputeGas
	}
	if s.FraudProofGas == 0 {
		s.FraudProofGas = defaults.FraudProofGas
	}
	if s.ConditionResolutionGas == 0 {
		s.ConditionResolutionGas = defaults.ConditionResolutionGas
	}
	if s.PenaltyRoutingGas == 0 {
		s.PenaltyRoutingGas = defaults.PenaltyRoutingGas
	}
	if s.FinalSettlementGas == 0 {
		s.FinalSettlementGas = defaults.FinalSettlementGas
	}
	if s.ReplayProtectionGas == 0 {
		s.ReplayProtectionGas = defaults.ReplayProtectionGas
	}
	if s.PerSignatureGas == 0 {
		s.PerSignatureGas = defaults.PerSignatureGas
	}
	if s.PerConditionGas == 0 {
		s.PerConditionGas = defaults.PerConditionGas
	}
	if s.PerFraudProofGas == 0 {
		s.PerFraudProofGas = defaults.PerFraudProofGas
	}
	if s.PerPenaltyAllocationGas == 0 {
		s.PerPenaltyAllocationGas = defaults.PerPenaltyAllocationGas
	}
	if s.PerStateByteGas == 0 {
		s.PerStateByteGas = defaults.PerStateByteGas
	}
	return s
}

func (s SettlementGasCostSchedule) Validate() error {
	schedule := s.Normalize()
	values := []uint64{
		schedule.OpenGas,
		schedule.CooperativeCloseGas,
		schedule.UnilateralCloseGas,
		schedule.DisputeGas,
		schedule.FraudProofGas,
		schedule.ConditionResolutionGas,
		schedule.PenaltyRoutingGas,
		schedule.FinalSettlementGas,
		schedule.ReplayProtectionGas,
		schedule.PerSignatureGas,
		schedule.PerConditionGas,
		schedule.PerFraudProofGas,
		schedule.PerPenaltyAllocationGas,
		schedule.PerStateByteGas,
	}
	for _, value := range values {
		if value == 0 {
			return errors.New("payments settlement gas cost schedule must be positive")
		}
	}
	return nil
}

func (h SecurityReserveAllocationHook) Normalize() SecurityReserveAllocationHook {
	h.HookID = normalizeOptionalHash(h.HookID)
	h.ChannelID = normalizeHash(h.ChannelID)
	h.ProofID = normalizeHash(h.ProofID)
	h.Offender = strings.TrimSpace(h.Offender)
	h.Denom = normalizeAssetDenom(h.Denom)
	h.Amount = strings.TrimSpace(h.Amount)
	h.Allocation = normalizeOptionalHash(h.Allocation)
	if h.Route == "" {
		h.Route = PenaltyRouteSecurityReserve
	}
	return h
}

func (h SecurityReserveAllocationHook) ValidateForChannel(channel ChannelRecord) error {
	hook := h.Normalize()
	channel = channel.Normalize()
	if err := ValidateHash("payments security reserve hook id", hook.HookID); err != nil {
		return err
	}
	if hook.ChannelID != channel.ChannelID {
		return errors.New("payments security reserve hook channel mismatch")
	}
	if err := ValidateHash("payments security reserve proof id", hook.ProofID); err != nil {
		return err
	}
	if !containsString(channel.Participants, hook.Offender) {
		return errors.New("payments security reserve hook offender must be channel participant")
	}
	if hook.Denom != NativeDenom {
		return fmt.Errorf("payments security reserve hook denom must be %s", NativeDenom)
	}
	if hook.Route != PenaltyRouteSecurityReserve {
		return errors.New("payments security reserve hook route mismatch")
	}
	if err := validatePositiveInt("payments security reserve hook amount", hook.Amount); err != nil {
		return err
	}
	if hook.Height == 0 {
		return errors.New("payments security reserve hook height must be positive")
	}
	if err := ValidateHash("payments security reserve allocation commitment", hook.Allocation); err != nil {
		return err
	}
	return nil
}

func (l SettlementInclusionLatency) Normalize() SettlementInclusionLatency {
	l.RecordID = normalizeOptionalHash(l.RecordID)
	l.OperationID = normalizeOptionalHash(l.OperationID)
	l.ChannelID = normalizeHash(l.ChannelID)
	if l.IncludedHeight >= l.SubmittedHeight && l.SubmittedHeight != 0 {
		l.LatencyBlocks = l.IncludedHeight - l.SubmittedHeight
	}
	if l.SLOThreshold > 0 {
		l.Breached = l.LatencyBlocks > l.SLOThreshold
	}
	return l
}

func (l SettlementInclusionLatency) Validate(channels []ChannelRecord) error {
	record := l.Normalize()
	if err := ValidateHash("payments settlement inclusion latency id", record.RecordID); err != nil {
		return err
	}
	if err := ValidateHash("payments settlement inclusion operation id", record.OperationID); err != nil {
		return err
	}
	if err := ValidateHash("payments settlement inclusion channel id", record.ChannelID); err != nil {
		return err
	}
	if !IsSettlementArbitrationOperation(record.Operation) {
		return fmt.Errorf("unknown payments settlement inclusion operation %q", record.Operation)
	}
	if record.SubmittedHeight == 0 || record.IncludedHeight == 0 || record.IncludedHeight < record.SubmittedHeight {
		return errors.New("payments settlement inclusion heights are invalid")
	}
	if record.LatencyBlocks != record.IncludedHeight-record.SubmittedHeight {
		return errors.New("payments settlement inclusion latency mismatch")
	}
	if record.SLOThreshold == 0 {
		return errors.New("payments settlement inclusion SLO threshold must be positive")
	}
	if record.Breached != (record.LatencyBlocks > record.SLOThreshold) {
		return errors.New("payments settlement inclusion breach marker mismatch")
	}
	if _, found := channelMap(channels)[record.ChannelID]; !found {
		return errors.New("payments settlement inclusion channel not found")
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

func DefaultPenaltyMatrix() []PenaltyMatrixEntry {
	return []PenaltyMatrixEntry{
		{Class: PenaltyClassInvalidClose, ProofType: FraudProofTypeInvalidClose, Source: PenaltySourceChannelBalance, BasePenalty: "10", ReporterRewardCap: "5", CounterpartyCompensation: "5", BurnShareBps: 2_500, SecurityReserveShareBps: 2_500, CommunityPoolShareBps: 5_000, Bounded: true},
		{Class: PenaltyClassInvalidClose, ProofType: FraudProofTypeInvalidBalance, Source: PenaltySourceChannelBalance, BasePenalty: "10", ReporterRewardCap: "5", CounterpartyCompensation: "5", BurnShareBps: 2_500, SecurityReserveShareBps: 2_500, CommunityPoolShareBps: 5_000, Bounded: true},
		{Class: PenaltyClassStaleClose, ProofType: FraudProofTypeStaleClose, Source: PenaltySourceChannelBalance, BasePenalty: "10", ReporterRewardCap: "5", CounterpartyCompensation: "5", BurnShareBps: 2_500, SecurityReserveShareBps: 2_500, CommunityPoolShareBps: 5_000, Bounded: true},
		{Class: PenaltyClassDoubleSign, ProofType: FraudProofTypeDoubleSign, Source: PenaltySourceParticipantBond, BasePenalty: "20", ReporterRewardCap: "8", CounterpartyCompensation: "6", BurnShareBps: 3_000, SecurityReserveShareBps: 4_000, CommunityPoolShareBps: 3_000, Bounded: true},
		{Class: PenaltyClassInvalidCondition, ProofType: FraudProofTypeInvalidCondition, Source: PenaltySourceChannelBalance, BasePenalty: "8", ReporterRewardCap: "4", CounterpartyCompensation: "4", BurnShareBps: 2_500, SecurityReserveShareBps: 2_500, CommunityPoolShareBps: 5_000, Bounded: true},
		{Class: PenaltyClassReplayAttempt, ProofType: FraudProofTypeReplayAttempt, Source: PenaltySourceChannelBalance, BasePenalty: "8", ReporterRewardCap: "4", CounterpartyCompensation: "4", BurnShareBps: 3_000, SecurityReserveShareBps: 3_000, CommunityPoolShareBps: 4_000, Bounded: true},
		{Class: PenaltyClassAsyncOverexposure, ProofType: FraudProofTypeAsyncOverexposure, Source: PenaltySourceParticipantBond, BasePenalty: "12", ReporterRewardCap: "5", CounterpartyCompensation: "5", BurnShareBps: 3_000, SecurityReserveShareBps: 3_000, CommunityPoolShareBps: 4_000, Bounded: true},
		{Class: PenaltyClassInvalidFraudProof, Source: PenaltySourceFraudProofDeposit, BasePenalty: "0", InvalidProofVerifierCost: "1", Bounded: true},
	}
}

func PenaltyClassForFraudProofType(proofType FraudProofType) (PaymentPenaltyClass, error) {
	switch proofType {
	case FraudProofTypeInvalidClose:
		return PenaltyClassInvalidClose, nil
	case FraudProofTypeStaleClose:
		return PenaltyClassStaleClose, nil
	case FraudProofTypeDoubleSign:
		return PenaltyClassDoubleSign, nil
	case FraudProofTypeInvalidCondition:
		return PenaltyClassInvalidCondition, nil
	case FraudProofTypeReplayAttempt:
		return PenaltyClassReplayAttempt, nil
	case FraudProofTypeAsyncOverexposure:
		return PenaltyClassAsyncOverexposure, nil
	case FraudProofTypeInvalidBalance:
		return PenaltyClassInvalidClose, nil
	default:
		return "", fmt.Errorf("unknown payments fraud proof type %q", proofType)
	}
}

func PenaltyMatrixEntryForProof(proofType FraudProofType, matrix []PenaltyMatrixEntry) (PenaltyMatrixEntry, error) {
	class, err := PenaltyClassForFraudProofType(proofType)
	if err != nil {
		return PenaltyMatrixEntry{}, err
	}
	for _, entry := range normalizePenaltyMatrix(matrix) {
		if entry.Class == class && (entry.ProofType == "" || entry.ProofType == proofType) {
			return entry, nil
		}
	}
	return PenaltyMatrixEntry{}, errors.New("payments penalty matrix entry not found")
}

func BuildPenaltyRouteAccounting(channel ChannelRecord, proof FraudProof, matrix []PenaltyMatrixEntry, policy FraudPenaltyPolicy) (PenaltyRouteAccounting, error) {
	channel = channel.Normalize()
	proof = proof.Normalize()
	if err := proof.ValidateForChannel(channel); err != nil {
		return PenaltyRouteAccounting{}, err
	}
	entry, err := PenaltyMatrixEntryForProof(proof.ProofType, matrix)
	if err != nil {
		return PenaltyRouteAccounting{}, err
	}
	if err := entry.Validate(); err != nil {
		return PenaltyRouteAccounting{}, err
	}
	policy = policy.Normalize()
	if policy.ReporterRewardCap == "" {
		policy.ReporterRewardCap = entry.ReporterRewardCap
	}
	if policy.CounterpartyRewardCap == "" {
		policy.CounterpartyRewardCap = entry.CounterpartyCompensation
	}
	if policy.BurnShareBps == 0 && policy.SecurityReserveShareBps == 0 && policy.CommunityPoolShareBps == 0 {
		policy.BurnShareBps = entry.BurnShareBps
		policy.SecurityReserveShareBps = entry.SecurityReserveShareBps
		policy.CommunityPoolShareBps = entry.CommunityPoolShareBps
	}
	if err := policy.Validate(); err != nil {
		return PenaltyRouteAccounting{}, err
	}
	penaltyAmount, err := parsePositiveInt("payments penalty accounting amount", proof.PenaltyAmount)
	if err != nil {
		return PenaltyRouteAccounting{}, err
	}
	if entry.BasePenalty != "" {
		basePenalty, err := parseNonNegativeInt("payments penalty matrix base", entry.BasePenalty)
		if err != nil {
			return PenaltyRouteAccounting{}, err
		}
		if penaltyAmount.LT(basePenalty) {
			return PenaltyRouteAccounting{}, errors.New("payments fraud penalty below matrix minimum")
		}
	}
	remaining := penaltyAmount
	penalties := []Penalty{}
	counterpartyComp := sdkmath.ZeroInt()
	if proof.OffendingSigner == channel.PendingClose.Submitter {
		counterparty := channelCounterparty(channel, proof.OffendingSigner)
		if counterparty != "" {
			counterpartyComp, err = cappedPenaltyPortion(remaining, policy.CounterpartyRewardCap, policy.CounterpartyRewardBps)
			if err != nil {
				return PenaltyRouteAccounting{}, err
			}
			if counterpartyComp.IsPositive() {
				penalty := Penalty{Offender: proof.OffendingSigner, Recipient: counterparty, Denom: NativeDenom, Amount: counterpartyComp.String()}.Normalize()
				if err := penalty.ValidateForChannel(channel); err != nil {
					return PenaltyRouteAccounting{}, err
				}
				penalties = append(penalties, penalty)
				remaining = remaining.Sub(counterpartyComp)
			}
		}
	}
	reporterReward := sdkmath.ZeroInt()
	if remaining.IsPositive() {
		reporterReward, err = cappedPenaltyPortion(remaining, policy.ReporterRewardCap, 0)
		if err != nil {
			return PenaltyRouteAccounting{}, err
		}
		if reporterReward.IsPositive() {
			penalty := Penalty{Offender: proof.OffendingSigner, Recipient: proof.SubmittedBy, Denom: NativeDenom, Amount: reporterReward.String()}.Normalize()
			if err := penalty.ValidateForChannel(channel); err != nil {
				return PenaltyRouteAccounting{}, err
			}
			penalties = append(penalties, penalty)
			remaining = remaining.Sub(reporterReward)
		}
	}
	allocations, err := splitPenaltyRemainder(proof.OffendingSigner, remaining, policy)
	if err != nil {
		return PenaltyRouteAccounting{}, err
	}
	return PenaltyRouteAccounting{
		Class:			entry.Class,
		Source:			entry.Source,
		TotalPenalty:		penaltyAmount.String(),
		ReporterReward:		reporterReward.String(),
		CounterpartyComp:	counterpartyComp.String(),
		Allocations:		allocations,
		Penalties:		normalizePenalties(penalties),
	}, nil
}

func ComputeInvalidFraudProofSubmissionPenalty(submitter, depositAmount, verificationCost string) (InvalidFraudProofSubmissionPenalty, error) {
	submitter = strings.TrimSpace(submitter)
	if err := addressing.ValidateUserAddress("payments invalid fraud proof submitter", submitter); err != nil {
		return InvalidFraudProofSubmissionPenalty{}, err
	}
	deposit, err := parseNonNegativeInt("payments invalid fraud proof deposit", strings.TrimSpace(depositAmount))
	if err != nil {
		return InvalidFraudProofSubmissionPenalty{}, err
	}
	cost, err := parseNonNegativeInt("payments invalid fraud proof verification cost", strings.TrimSpace(verificationCost))
	if err != nil {
		return InvalidFraudProofSubmissionPenalty{}, err
	}
	forfeited := cost
	if forfeited.GT(deposit) {
		forfeited = deposit
	}
	return InvalidFraudProofSubmissionPenalty{
		Submitter:		submitter,
		Denom:			NativeDenom,
		DepositAmount:		deposit.String(),
		VerificationCost:	cost.String(),
		ForfeitedAmount:	forfeited.String(),
		RefundAmount:		deposit.Sub(forfeited).String(),
	}, nil
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
		EventID:	HashParts("channel-open", channel.ChannelID, channel.OpeningStateHash),
		EventType:	"channel-open",
		ChannelID:	channel.ChannelID,
		Height:		channel.OpenHeight,
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
		EventID:	HashParts("channel-dispute", channel.ChannelID, channel.PendingClose.State.StateHash, fmt.Sprintf("%d", height)),
		EventType:	"channel-dispute",
		ChannelID:	channel.ChannelID,
		Height:		height,
		Attributes: []PaymentEventAttribute{
			{Key: "submitter", Value: strings.TrimSpace(submitter)},
			{Key: "state_hash", Value: channel.PendingClose.State.StateHash},
			{Key: "nonce", Value: fmt.Sprintf("%d", channel.PendingClose.State.Nonce)},
			{Key: "settle_after_height", Value: fmt.Sprintf("%d", channel.PendingClose.SettleAfterHeight)},
		},
	}
	return event.Normalize()
}

func ValidatorAssistedDisputeEvent(metadata ValidatorPaymentServiceMetadata, channel ChannelRecord, delegator string, height uint64) PaymentEvent {
	metadata = metadata.Normalize()
	channel = channel.Normalize()
	event := PaymentEvent{
		EventID:	HashParts("validator-assisted-dispute", metadata.ValidatorAddress, channel.ChannelID, delegator, fmt.Sprintf("%d", height)),
		EventType:	"validator-assisted-dispute",
		ChannelID:	channel.ChannelID,
		Height:		height,
		Attributes: []PaymentEventAttribute{
			{Key: "validator", Value: metadata.ValidatorAddress},
			{Key: "watch_service", Value: metadata.ServiceAddress},
			{Key: "delegator", Value: strings.TrimSpace(delegator)},
			{Key: "metadata_hash", Value: metadata.MetadataHash},
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
		EventID:	HashParts("channel-finality", channel.ChannelID, string(previous), string(next), fmt.Sprintf("%d", height)),
		EventType:	"channel-finality-transition",
		ChannelID:	channel.ChannelID,
		Height:		height,
		Attributes:	attrs,
	}
	return event.Normalize()
}

func AsyncSettlementCompletionEvent(completion AsyncSettlementCompletion) PaymentEvent {
	completion = completion.Normalize()
	event := PaymentEvent{
		EventID:	HashParts("async-settlement-completion-event", completion.CompletionID),
		EventType:	"async-settlement-completion",
		ChannelID:	completion.ChannelID,
		Height:		completion.Height,
		Attributes: []PaymentEventAttribute{
			{Key: "job_id", Value: completion.JobID},
			{Key: "job_type", Value: completion.JobType},
			{Key: "object_id", Value: completion.ObjectID},
			{Key: "result_hash", Value: completion.ResultHash},
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
	e.AdvertisementFeePaid = strings.TrimSpace(e.AdvertisementFeePaid)
	if e.AdvertisementFeePaid == "" {
		e.AdvertisementFeePaid = "0"
	}
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
	if err := validateNonNegativeInt("payments routing fee", e.FeeAmount); err != nil {
		return err
	}
	return validateNonNegativeInt("payments routing advertisement fee", e.AdvertisementFeePaid)
}

func BuildLiquidityAdvertisement(ad LiquidityAdvertisement, requiredDeposit string) (LiquidityAdvertisement, error) {
	ad = ad.Normalize()
	if ad.AdvertisementID == "" {
		ad.AdvertisementID = HashParts("liquidity-advertisement", ad.ChannelID, ad.Advertiser, ad.Counterparty, ad.Capacity, fmt.Sprintf("%020d", ad.ValidUntilHeight))
	}
	ad.AdvertisementHash = ""
	ad.AdvertisementHash = ComputeLiquidityAdvertisementHash(ad)
	if err := ad.Validate(requiredDeposit); err != nil {
		return LiquidityAdvertisement{}, err
	}
	return ad.Normalize(), nil
}

func ComputeLiquidityAdvertisementHash(ad LiquidityAdvertisement) string {
	ad = ad.Normalize()
	return HashParts(
		"liquidity-advertisement",
		ad.AdvertisementID,
		ad.ChannelID,
		ad.Advertiser,
		ad.Counterparty,
		ad.Capacity,
		ad.FeeDenom,
		ad.BaseFee,
		ad.ReservationFee,
		ad.VirtualSetupFee,
		fmt.Sprintf("%010d", ad.ReliabilityBps),
		fmt.Sprintf("%020d", ad.ValidUntilHeight),
		ad.DepositAmount,
		fmt.Sprintf("%t", ad.BackedByReservation),
	)
}

func (ad LiquidityAdvertisement) Normalize() LiquidityAdvertisement {
	ad.AdvertisementID = normalizeOptionalHash(ad.AdvertisementID)
	ad.ChannelID = normalizeHash(ad.ChannelID)
	ad.Advertiser = strings.TrimSpace(ad.Advertiser)
	ad.Counterparty = strings.TrimSpace(ad.Counterparty)
	ad.Capacity = strings.TrimSpace(ad.Capacity)
	ad.FeeDenom = normalizeAssetDenom(ad.FeeDenom)
	ad.BaseFee = strings.TrimSpace(ad.BaseFee)
	ad.ReservationFee = strings.TrimSpace(ad.ReservationFee)
	ad.VirtualSetupFee = strings.TrimSpace(ad.VirtualSetupFee)
	ad.DepositAmount = strings.TrimSpace(ad.DepositAmount)
	for _, field := range []*string{&ad.BaseFee, &ad.ReservationFee, &ad.VirtualSetupFee, &ad.DepositAmount} {
		if *field == "" {
			*field = "0"
		}
	}
	ad.AdvertisementHash = normalizeOptionalHash(ad.AdvertisementHash)
	return ad
}

func (ad LiquidityAdvertisement) Validate(requiredDeposit string) error {
	ad = ad.Normalize()
	if err := ValidateHash("payments liquidity advertisement id", ad.AdvertisementID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity advertisement channel id", ad.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity advertiser", ad.Advertiser); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity counterparty", ad.Counterparty); err != nil {
		return err
	}
	if ad.Advertiser == ad.Counterparty {
		return errors.New("payments liquidity advertisement parties must differ")
	}
	if err := validatePositiveInt("payments liquidity advertised capacity", ad.Capacity); err != nil {
		return err
	}
	if ad.FeeDenom != NativeDenom {
		return fmt.Errorf("payments liquidity advertisement fee denom must be %s", NativeDenom)
	}
	for _, item := range []struct {
		name	string
		amount	string
	}{
		{"payments liquidity base fee", ad.BaseFee},
		{"payments liquidity reservation fee", ad.ReservationFee},
		{"payments liquidity virtual setup fee", ad.VirtualSetupFee},
		{"payments liquidity advertisement deposit", ad.DepositAmount},
	} {
		if err := validateNonNegativeInt(item.name, item.amount); err != nil {
			return err
		}
	}
	if ad.ReliabilityBps > MaxPenaltyRouteBps {
		return errors.New("payments liquidity reliability bps exceeds 10000")
	}
	if ad.ValidUntilHeight == 0 {
		return errors.New("payments liquidity advertisement validity height must be positive")
	}
	requiredDeposit = strings.TrimSpace(requiredDeposit)
	if requiredDeposit != "" {
		required, err := parseNonNegativeInt("payments liquidity required deposit", requiredDeposit)
		if err != nil {
			return err
		}
		deposit, err := parseNonNegativeInt("payments liquidity advertisement deposit", ad.DepositAmount)
		if err != nil {
			return err
		}
		if deposit.LT(required) {
			return errors.New("payments liquidity advertisement deposit below required")
		}
	}
	if ad.AdvertisementHash == "" {
		return errors.New("payments liquidity advertisement hash is required")
	}
	if expected := ComputeLiquidityAdvertisementHash(ad); ad.AdvertisementHash != expected {
		return errors.New("payments liquidity advertisement hash mismatch")
	}
	return nil
}

func LiquidityAvailabilityScore(ad LiquidityAdvertisement, stats EdgeRoutingStats) (int64, error) {
	ad = ad.Normalize()
	if err := ad.Validate("0"); err != nil {
		return 0, err
	}
	capacity, err := parsePositiveInt("payments liquidity score capacity", ad.Capacity)
	if err != nil {
		return 0, err
	}
	score := capacity.QuoRaw(10).Int64()
	score += int64(ad.ReliabilityBps) / 100
	if ad.BackedByReservation {
		score += 100
	}
	stats = stats.Normalize()
	score += int64(stats.SuccessRateBps) / 200
	score -= int64(stats.FailureCount) * 25
	score -= int64(stats.CongestionBps) / 200
	score -= int64(stats.PendingConditionCount) * 5
	return score, nil
}

func ApplyFalseLiquidityAdvertisementPenalty(store TopologyStore, ad LiquidityAdvertisement, currentHeight uint64) (TopologyStore, string, error) {
	ad = ad.Normalize()
	if err := ad.Validate("0"); err != nil {
		return TopologyStore{}, "", err
	}
	forfeited, err := parseNonNegativeInt("payments false liquidity deposit", ad.DepositAmount)
	if err != nil {
		return TopologyStore{}, "", err
	}
	next := PenalizeInvalidGossip(store, ad.Advertiser, currentHeight)
	return next, forfeited.String(), nil
}

func BuildSignedLiquidityReservation(reservation SignedLiquidityReservation, signer string) (SignedLiquidityReservation, error) {
	reservation = reservation.Normalize()
	if reservation.ReservationID == "" {
		reservation.ReservationID = HashParts("liquidity-reservation", reservation.AdvertisementID, reservation.ChannelID, reservation.Reserver, fmt.Sprintf("%020d", reservation.Nonce))
	}
	reservation.CommitmentHash = ""
	reservation.Signature = LiquidityReservationSignature{}
	reservation.CommitmentHash = ComputeLiquidityReservationHash(reservation)
	signature, err := SignatureForLiquidityReservation(reservation, signer)
	if err != nil {
		return SignedLiquidityReservation{}, err
	}
	reservation.Signature = signature
	if err := reservation.Validate(); err != nil {
		return SignedLiquidityReservation{}, err
	}
	return reservation.Normalize(), nil
}

func ComputeLiquidityReservationHash(reservation SignedLiquidityReservation) string {
	reservation = reservation.Normalize()
	return HashParts(
		"liquidity-reservation",
		reservation.ReservationID,
		reservation.AdvertisementID,
		reservation.ChainID,
		reservation.ChannelID,
		reservation.Reserver,
		reservation.Counterparty,
		reservation.Capacity,
		reservation.FeeAmount,
		fmt.Sprintf("%020d", reservation.ExpirationHeight),
		fmt.Sprintf("%020d", reservation.Nonce),
	)
}

func SignatureForLiquidityReservation(reservation SignedLiquidityReservation, signer string) (LiquidityReservationSignature, error) {
	reservation = reservation.Normalize()
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments liquidity reservation signer", signer); err != nil {
		return LiquidityReservationSignature{}, err
	}
	if reservation.CommitmentHash == "" {
		reservation.CommitmentHash = ComputeLiquidityReservationHash(reservation)
	}
	return LiquidityReservationSignature{
		Signer:			signer,
		ChainID:		reservation.ChainID,
		ChannelID:		reservation.ChannelID,
		ObjectType:		SignatureObjectLiquidity,
		Version:		CurrentStateVersion,
		Nonce:			reservation.Nonce,
		ObjectID:		reservation.ReservationID,
		ExpirationHeight:	reservation.ExpirationHeight,
		CommitmentHash:		reservation.CommitmentHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			reservation.ChainID,
			reservation.ChannelID,
			SignatureObjectLiquidity,
			CurrentStateVersion,
			reservation.Nonce,
			reservation.ReservationID,
			reservation.ExpirationHeight,
			reservation.CommitmentHash,
		),
	}, nil
}

func (r SignedLiquidityReservation) Normalize() SignedLiquidityReservation {
	r.ReservationID = normalizeOptionalHash(r.ReservationID)
	r.AdvertisementID = normalizeHash(r.AdvertisementID)
	r.ChainID = strings.TrimSpace(r.ChainID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Reserver = strings.TrimSpace(r.Reserver)
	r.Counterparty = strings.TrimSpace(r.Counterparty)
	r.Capacity = strings.TrimSpace(r.Capacity)
	r.FeeAmount = strings.TrimSpace(r.FeeAmount)
	if r.FeeAmount == "" {
		r.FeeAmount = "0"
	}
	r.CommitmentHash = normalizeOptionalHash(r.CommitmentHash)
	r.Signature = r.Signature.Normalize()
	return r
}

func (r SignedLiquidityReservation) Validate() error {
	reservation := r.Normalize()
	if err := ValidateHash("payments liquidity reservation id", reservation.ReservationID); err != nil {
		return err
	}
	if err := ValidateHash("payments liquidity reservation advertisement id", reservation.AdvertisementID); err != nil {
		return err
	}
	if reservation.ChainID == "" {
		return errors.New("payments liquidity reservation chain id is required")
	}
	if err := ValidateHash("payments liquidity reservation channel id", reservation.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity reserver", reservation.Reserver); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments liquidity reservation counterparty", reservation.Counterparty); err != nil {
		return err
	}
	if reservation.Reserver == reservation.Counterparty {
		return errors.New("payments liquidity reservation parties must differ")
	}
	if err := validatePositiveInt("payments liquidity reservation capacity", reservation.Capacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments liquidity reservation fee", reservation.FeeAmount); err != nil {
		return err
	}
	if reservation.ExpirationHeight == 0 || reservation.Nonce == 0 {
		return errors.New("payments liquidity reservation expiration and nonce must be positive")
	}
	if reservation.CommitmentHash == "" {
		return errors.New("payments liquidity reservation commitment is required")
	}
	if expected := ComputeLiquidityReservationHash(reservation); reservation.CommitmentHash != expected {
		return errors.New("payments liquidity reservation commitment mismatch")
	}
	return reservation.Signature.Validate(reservation)
}

func (s LiquidityReservationSignature) Normalize() LiquidityReservationSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.ChannelID = normalizeHash(s.ChannelID)
	s.ObjectType = strings.TrimSpace(s.ObjectType)
	s.ObjectID = normalizeOptionalHash(s.ObjectID)
	s.CommitmentHash = normalizeOptionalHash(s.CommitmentHash)
	s.SignatureHash = normalizeOptionalHash(s.SignatureHash)
	return s
}

func (s LiquidityReservationSignature) Validate(reservation SignedLiquidityReservation) error {
	sig := s.Normalize()
	reservation = reservation.Normalize()
	if err := addressing.ValidateUserAddress("payments liquidity reservation signature signer", sig.Signer); err != nil {
		return err
	}
	if sig.Signer != reservation.Reserver {
		return errors.New("payments liquidity reservation signer mismatch")
	}
	if sig.ChainID != reservation.ChainID || sig.ChannelID != reservation.ChannelID {
		return errors.New("payments liquidity reservation signature domain mismatch")
	}
	if sig.ObjectType != SignatureObjectLiquidity || sig.Version != CurrentStateVersion {
		return errors.New("payments liquidity reservation signature object mismatch")
	}
	if sig.Nonce != reservation.Nonce || sig.ObjectID != reservation.ReservationID || sig.ExpirationHeight != reservation.ExpirationHeight || sig.CommitmentHash != reservation.CommitmentHash {
		return errors.New("payments liquidity reservation signature commitment mismatch")
	}
	if err := ValidateHash("payments liquidity reservation signature hash", sig.SignatureHash); err != nil {
		return err
	}
	expected := ComputeSignatureEnvelopeHash(sig.Signer, sig.ChainID, sig.ChannelID, sig.ObjectType, sig.Version, sig.Nonce, sig.ObjectID, sig.ExpirationHeight, sig.CommitmentHash)
	if sig.SignatureHash != expected {
		return errors.New("payments liquidity reservation signature value mismatch")
	}
	return nil
}

func BuildVirtualChannel(vc VirtualChannel) (VirtualChannel, error) {
	vc = vc.Normalize()
	if vc.ParentRouteID == "" {
		parts := append([]string{"virtual-parent-route", vc.VirtualChannelID}, vc.ParentChannelIDs...)
		vc.ParentRouteID = HashParts(parts...)
	}
	if vc.IntermediarySetHash == "" {
		vc.IntermediarySetHash = ComputeParticipantSetHash(vc.Intermediaries)
	}
	vc.AnchorCommitment = ""
	vc.StateHash = ""
	vc.AnchorCommitment = ComputeVirtualChannelAnchor(vc)
	vc.StateHash = ComputeVirtualChannelStateHash(vc)
	if err := vc.ValidateCore(); err != nil {
		return VirtualChannel{}, err
	}
	return vc.Normalize(), nil
}

func SignatureForVirtualChannel(vc VirtualChannel, signer string) (StateSignature, error) {
	vc = vc.Normalize()
	if vc.StateHash == "" {
		var err error
		vc, err = BuildVirtualChannel(vc)
		if err != nil {
			return StateSignature{}, err
		}
	}
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments virtual channel signer", signer); err != nil {
		return StateSignature{}, err
	}
	return StateSignature{
		Signer:			signer,
		ChainID:		vc.ChainID,
		ChannelID:		vc.VirtualChannelID,
		ObjectType:		SignatureObjectVirtual,
		Version:		CurrentStateVersion,
		Nonce:			vc.Nonce,
		ObjectID:		vc.StateHash,
		ExpirationHeight:	vc.ExpiresHeight,
		CommitmentHash:		vc.StateHash,
		StateHash:		vc.StateHash,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			vc.ChainID,
			vc.VirtualChannelID,
			SignatureObjectVirtual,
			CurrentStateVersion,
			vc.Nonce,
			vc.StateHash,
			vc.ExpiresHeight,
			vc.StateHash,
		),
	}, nil
}

func ValidateVirtualChannelSignature(sig StateSignature, vc VirtualChannel) error {
	sig = sig.Normalize()
	vc = vc.Normalize()
	if err := addressing.ValidateUserAddress("payments virtual channel signature signer", sig.Signer); err != nil {
		return err
	}
	if sig.ChainID != vc.ChainID {
		return errors.New("payments virtual channel signature chain id mismatch")
	}
	if sig.ChannelID != vc.VirtualChannelID {
		return errors.New("payments virtual channel signature channel id mismatch")
	}
	if sig.ObjectType != SignatureObjectVirtual {
		return errors.New("payments virtual channel signature object type mismatch")
	}
	if sig.Version != CurrentStateVersion {
		return errors.New("payments virtual channel signature version mismatch")
	}
	if sig.Nonce != vc.Nonce {
		return errors.New("payments virtual channel signature nonce mismatch")
	}
	if sig.ObjectID != vc.StateHash || sig.CommitmentHash != vc.StateHash || sig.StateHash != vc.StateHash {
		return errors.New("payments virtual channel signature commitment mismatch")
	}
	if sig.ExpirationHeight != vc.ExpiresHeight {
		return errors.New("payments virtual channel signature expiration mismatch")
	}
	if err := ValidateHash("payments virtual channel signature hash", sig.SignatureHash); err != nil {
		return err
	}
	expected := ComputeSignatureEnvelopeHash(sig.Signer, sig.ChainID, sig.ChannelID, sig.ObjectType, sig.Version, sig.Nonce, sig.ObjectID, sig.ExpirationHeight, sig.CommitmentHash)
	if sig.SignatureHash != expected {
		return errors.New("payments virtual channel signature hash mismatch")
	}
	return nil
}

func BuildVirtualParentReserve(vc VirtualChannel, reserve VirtualParentReserve, signer string) (VirtualParentReserve, error) {
	vc = vc.Normalize()
	reserve = reserve.Normalize()
	signer = strings.TrimSpace(signer)
	if reserve.ParentChannelID == "" {
		return VirtualParentReserve{}, errors.New("payments virtual reserve parent channel id is required")
	}
	if reserve.Capacity == "" {
		reserve.Capacity = vc.Capacity
	}
	if reserve.SplitAmount == "" {
		reserve.SplitAmount = reserve.Capacity
	}
	if reserve.FeeAmount == "" {
		reserve.FeeAmount = "0"
	}
	if reserve.SegmentID == "" {
		reserve.SegmentID = HashParts("virtual-reserve-segment", vc.VirtualChannelID, reserve.ParentChannelID, reserve.ReservedBy, reserve.SplitAmount)
	}
	if signer != "" {
		reserve.ReservedBy = signer
	}
	if reserve.ReservedBy == "" {
		return VirtualParentReserve{}, errors.New("payments virtual reserve signer is required")
	}
	reserve.ReserveCommitment = ComputeVirtualReserveCommitment(vc, reserve)
	sig, err := SignatureForVirtualReservation(vc, reserve, reserve.ReservedBy)
	if err != nil {
		return VirtualParentReserve{}, err
	}
	reserve.Signature = sig
	return reserve.Normalize(), nil
}

func ComputeVirtualReserveCommitment(vc VirtualChannel, reserve VirtualParentReserve) string {
	vc = vc.Normalize()
	reserve = reserve.Normalize()
	return HashParts(
		"virtual-reserve-commitment",
		vc.ChainID,
		vc.VirtualChannelID,
		vc.ParentRouteID,
		reserve.SegmentID,
		reserve.ParentChannelID,
		reserve.ReservedBy,
		reserve.Capacity,
		reserve.SplitAmount,
		reserve.FeeAmount,
		fmt.Sprintf("%020d", vc.ExpiresHeight),
	)
}

func SignatureForVirtualReservation(vc VirtualChannel, reserve VirtualParentReserve, signer string) (VirtualReservationSignature, error) {
	vc = vc.Normalize()
	reserve = reserve.Normalize()
	signer = strings.TrimSpace(signer)
	if err := addressing.ValidateUserAddress("payments virtual reservation signer", signer); err != nil {
		return VirtualReservationSignature{}, err
	}
	if reserve.ReserveCommitment == "" {
		reserve.ReservedBy = signer
		reserve.ReserveCommitment = ComputeVirtualReserveCommitment(vc, reserve)
	}
	return VirtualReservationSignature{
		Signer:			signer,
		ChainID:		vc.ChainID,
		VirtualChannelID:	vc.VirtualChannelID,
		ParentRouteID:		vc.ParentRouteID,
		ParentChannelID:	reserve.ParentChannelID,
		ObjectType:		SignatureObjectVirtualReserve,
		Version:		CurrentStateVersion,
		Capacity:		reserve.Capacity,
		SplitAmount:		reserve.SplitAmount,
		FeeAmount:		reserve.FeeAmount,
		ExpirationHeight:	vc.ExpiresHeight,
		CommitmentHash:		reserve.ReserveCommitment,
		SignatureHash: ComputeSignatureEnvelopeHash(
			signer,
			vc.ChainID,
			vc.VirtualChannelID,
			SignatureObjectVirtualReserve,
			CurrentStateVersion,
			vc.Nonce,
			reserve.ReserveCommitment,
			vc.ExpiresHeight,
			reserve.ReserveCommitment,
		),
	}, nil
}

func ValidateVirtualReservationSignature(sig VirtualReservationSignature, vc VirtualChannel, reserve VirtualParentReserve) error {
	sig = sig.Normalize()
	vc = vc.Normalize()
	reserve = reserve.Normalize()
	if err := addressing.ValidateUserAddress("payments virtual reservation signature signer", sig.Signer); err != nil {
		return err
	}
	if sig.ChainID != vc.ChainID {
		return errors.New("payments virtual reservation signature chain id mismatch")
	}
	if sig.VirtualChannelID != vc.VirtualChannelID {
		return errors.New("payments virtual reservation signature channel id mismatch")
	}
	if sig.ParentRouteID != vc.ParentRouteID {
		return errors.New("payments virtual reservation signature route id mismatch")
	}
	if sig.ParentChannelID != reserve.ParentChannelID {
		return errors.New("payments virtual reservation signature parent channel mismatch")
	}
	if sig.ObjectType != SignatureObjectVirtualReserve {
		return errors.New("payments virtual reservation signature object type mismatch")
	}
	if sig.Version != CurrentStateVersion {
		return errors.New("payments virtual reservation signature version mismatch")
	}
	if sig.Capacity != reserve.Capacity || sig.FeeAmount != reserve.FeeAmount {
		return errors.New("payments virtual reservation signature amount mismatch")
	}
	if sig.SplitAmount != reserve.SplitAmount {
		return errors.New("payments virtual reservation signature split amount mismatch")
	}
	if sig.ExpirationHeight != vc.ExpiresHeight {
		return errors.New("payments virtual reservation signature expiration mismatch")
	}
	if sig.CommitmentHash != reserve.ReserveCommitment {
		return errors.New("payments virtual reservation signature commitment mismatch")
	}
	expectedCommitment := ComputeVirtualReserveCommitment(vc, reserve)
	if reserve.ReserveCommitment != expectedCommitment {
		return errors.New("payments virtual reserve commitment mismatch")
	}
	expected := ComputeSignatureEnvelopeHash(sig.Signer, sig.ChainID, sig.VirtualChannelID, sig.ObjectType, sig.Version, vc.Nonce, sig.CommitmentHash, sig.ExpirationHeight, sig.CommitmentHash)
	if sig.SignatureHash != expected {
		return errors.New("payments virtual reservation signature hash mismatch")
	}
	return ValidateHash("payments virtual reservation signature hash", sig.SignatureHash)
}

func BuildVirtualActivationProof(vc VirtualChannel, reserves []VirtualParentReserve, routeTimeoutHeight uint64) (VirtualActivationProof, error) {
	vc = vc.Normalize()
	if vc.StateHash == "" || vc.AnchorCommitment == "" || vc.ParentRouteID == "" || vc.IntermediarySetHash == "" {
		built, err := BuildVirtualChannel(vc)
		if err != nil {
			return VirtualActivationProof{}, err
		}
		built.Signatures = vc.Signatures
		vc = built.Normalize()
	}
	proof := VirtualActivationProof{
		VirtualChannel:		vc,
		ParentReserves:		normalizeVirtualParentReserves(reserves),
		RouteTimeoutHeight:	routeTimeoutHeight,
	}
	proof.ProofHash = ComputeVirtualActivationProofHash(proof)
	return proof.Normalize(), nil
}

func ComputeVirtualActivationProofHash(proof VirtualActivationProof) string {
	proof = proof.Normalize()
	parts := []string{
		"virtual-activation-proof",
		proof.VirtualChannel.ChainID,
		proof.VirtualChannel.VirtualChannelID,
		proof.VirtualChannel.StateHash,
		proof.VirtualChannel.ParentRouteID,
		fmt.Sprintf("%020d", proof.RouteTimeoutHeight),
		fmt.Sprintf("%t", proof.AggregatedCapacity),
	}
	for _, reserve := range proof.ParentReserves {
		parts = append(parts, reserve.SegmentID, reserve.ParentChannelID, reserve.ReservedBy, reserve.SplitAmount, reserve.ReserveCommitment)
	}
	return HashParts(parts...)
}

func ValidateVirtualActivationProof(proof VirtualActivationProof) error {
	proof = proof.Normalize()
	vc := proof.VirtualChannel
	if err := ValidateVirtualChannelActivation(vc); err != nil {
		return err
	}
	if proof.RouteTimeoutHeight == 0 {
		return errors.New("payments virtual activation proof route timeout is required")
	}
	if proof.RouteTimeoutHeight <= vc.ExpiresHeight {
		return errors.New("payments virtual activation proof route timeout must exceed virtual expiry")
	}
	if !proof.AggregatedCapacity && len(proof.ParentReserves) != len(vc.ParentChannelIDs) {
		return errors.New("payments virtual activation proof requires one reserve per parent")
	}
	if proof.AggregatedCapacity {
		if err := ValidateVirtualReserveSegments(vc, VirtualReserveSegmentsFromProof(proof)); err != nil {
			return err
		}
	}
	parentSet := make(map[string]struct{}, len(vc.ParentChannelIDs))
	for _, parentID := range vc.ParentChannelIDs {
		parentSet[parentID] = struct{}{}
	}
	seenParents := make(map[string]struct{}, len(proof.ParentReserves))
	coveredParents := make(map[string]struct{}, len(proof.ParentReserves))
	for _, reserve := range proof.ParentReserves {
		if _, found := parentSet[reserve.ParentChannelID]; !found {
			return errors.New("payments virtual activation proof reserve references unknown parent")
		}
		if !proof.AggregatedCapacity {
			if _, found := seenParents[reserve.ParentChannelID]; found {
				return errors.New("payments virtual activation proof duplicate parent reserve")
			}
			seenParents[reserve.ParentChannelID] = struct{}{}
		}
		if _, found := seenParents[reserve.SegmentID]; found {
			return errors.New("payments virtual activation proof duplicate parent reserve")
		}
		seenParents[reserve.SegmentID] = struct{}{}
		coveredParents[reserve.ParentChannelID] = struct{}{}
		if !containsString(vc.Intermediaries, reserve.ReservedBy) && !containsString(vc.Endpoints, reserve.ReservedBy) {
			return errors.New("payments virtual reserve signer must be route participant")
		}
		reserved, err := parsePositiveInt("payments virtual reserve capacity", reserve.Capacity)
		if err != nil {
			return err
		}
		capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
		if err != nil {
			return err
		}
		if reserved.LT(capacity) {
			if !proof.AggregatedCapacity {
				return errors.New("payments virtual reserve capacity below virtual capacity")
			}
			split, err := parsePositiveInt("payments virtual reserve split amount", reserve.SplitAmount)
			if err != nil {
				return err
			}
			if reserved.LT(split) {
				return errors.New("payments virtual reserve capacity below split amount")
			}
		}
		if err := ValidateVirtualReservationSignature(reserve.Signature, vc, reserve); err != nil {
			return err
		}
	}
	for parentID := range parentSet {
		if _, found := coveredParents[parentID]; !found {
			return errors.New("payments virtual activation proof missing parent reserve")
		}
	}
	expected := ComputeVirtualActivationProofHash(proof)
	if proof.ProofHash != expected {
		return errors.New("payments virtual activation proof hash mismatch")
	}
	return nil
}

func BuildVirtualCloseProof(final VirtualChannel, mode VirtualCloseMode, commitments []string, submittedBy string, closeHeight uint64) (VirtualCloseProof, error) {
	final = final.Normalize()
	if final.StateHash == "" || final.AnchorCommitment == "" {
		built, err := BuildVirtualChannel(final)
		if err != nil {
			return VirtualCloseProof{}, err
		}
		built.Signatures = final.Signatures
		final = built.Normalize()
	}
	proof := VirtualCloseProof{
		VirtualChannelID:		final.VirtualChannelID,
		ParentRouteID:			final.ParentRouteID,
		CloseMode:			mode,
		FinalState:			final,
		ParentReserveCommitments:	normalizeHashSlice(commitments),
		SubmittedBy:			strings.TrimSpace(submittedBy),
		CloseHeight:			closeHeight,
		ReleaseHeight:			VirtualCloseReleaseHeight(mode, closeHeight, final.ExpiresHeight),
	}
	proof.ProofHash = ComputeVirtualCloseProofHash(proof)
	return proof.Normalize(), nil
}

func VirtualCloseReleaseHeight(mode VirtualCloseMode, closeHeight, expiresHeight uint64) uint64 {
	switch mode {
	case VirtualCloseModeCooperative, VirtualCloseModeExpired:
		return closeHeight
	case VirtualCloseModeIntermediaryRisk, VirtualCloseModeDisputed:
		return closeHeight + DefaultDisputePeriod
	default:
		return 0
	}
}

func ComputeVirtualCloseProofHash(proof VirtualCloseProof) string {
	proof = proof.Normalize()
	parts := []string{
		"virtual-close-proof",
		proof.VirtualChannelID,
		proof.ParentRouteID,
		string(proof.CloseMode),
		proof.FinalState.StateHash,
		proof.SubmittedBy,
		fmt.Sprintf("%020d", proof.CloseHeight),
		fmt.Sprintf("%020d", proof.ReleaseHeight),
	}
	parts = append(parts, proof.ParentReserveCommitments...)
	return HashParts(parts...)
}

func ValidateVirtualCloseProof(proof VirtualCloseProof, current VirtualChannel, currentHeight uint64) error {
	proof = proof.Normalize()
	current = current.Normalize()
	if currentHeight == 0 || proof.CloseHeight != currentHeight {
		return errors.New("payments virtual close proof height mismatch")
	}
	if !IsVirtualCloseMode(proof.CloseMode) {
		return fmt.Errorf("unknown payments virtual close mode %q", proof.CloseMode)
	}
	if proof.VirtualChannelID != current.VirtualChannelID || proof.FinalState.VirtualChannelID != current.VirtualChannelID {
		return errors.New("payments virtual close proof channel mismatch")
	}
	if proof.ParentRouteID != current.ParentRouteID || proof.FinalState.ParentRouteID != current.ParentRouteID {
		return errors.New("payments virtual close proof route mismatch")
	}
	if !containsString(current.Endpoints, proof.SubmittedBy) && !containsString(current.Intermediaries, proof.SubmittedBy) {
		return errors.New("payments virtual close submitter must be route participant")
	}
	if len(current.ParentReserveCommitments) > 0 && strings.Join(proof.ParentReserveCommitments, "/") != strings.Join(current.ParentReserveCommitments, "/") {
		return errors.New("payments virtual close reserve commitment mismatch")
	}
	if proof.ReleaseHeight != VirtualCloseReleaseHeight(proof.CloseMode, proof.CloseHeight, current.ExpiresHeight) {
		return errors.New("payments virtual close release height mismatch")
	}
	switch proof.CloseMode {
	case VirtualCloseModeCooperative:
		if proof.FinalState.Nonce < current.Nonce {
			return errors.New("payments virtual cooperative close state is stale")
		}
		if err := validateVirtualEndpointSignedState(current, proof.FinalState, false); err != nil {
			return err
		}
	case VirtualCloseModeExpired:
		if currentHeight < current.ExpiresHeight+DefaultDisputePeriod {
			return errors.New("payments virtual expired close before finalization")
		}
		if proof.FinalState.Nonce < current.Nonce {
			return errors.New("payments virtual expired close state is stale")
		}
		if err := validateVirtualEndpointSignedState(current, proof.FinalState, false); err != nil {
			return err
		}
	case VirtualCloseModeIntermediaryRisk:
		if !containsString(current.Intermediaries, proof.SubmittedBy) {
			return errors.New("payments virtual intermediary-risk close requires intermediary submitter")
		}
		if proof.FinalState.Nonce < current.Nonce {
			return errors.New("payments virtual intermediary-risk close state is stale")
		}
		if err := validateVirtualEndpointSignedState(current, proof.FinalState, false); err != nil {
			return err
		}
	case VirtualCloseModeDisputed:
		if proof.FinalState.Nonce < current.Nonce {
			return errors.New("payments virtual disputed close state is stale")
		}
		if err := validateVirtualEndpointSignedState(current, proof.FinalState, false); err != nil {
			return err
		}
	}
	expected := ComputeVirtualCloseProofHash(proof)
	if proof.ProofHash != expected {
		return errors.New("payments virtual close proof hash mismatch")
	}
	return nil
}

func VirtualReserveSegmentsFromProof(proof VirtualActivationProof) []VirtualReserveSegment {
	proof = proof.Normalize()
	out := make([]VirtualReserveSegment, 0, len(proof.ParentReserves))
	for _, reserve := range proof.ParentReserves {
		segment := VirtualReserveSegment{
			SegmentID:		reserve.SegmentID,
			VirtualChannelID:	proof.VirtualChannel.VirtualChannelID,
			ParentChannelID:	reserve.ParentChannelID,
			ReserveCommitment:	reserve.ReserveCommitment,
			Capacity:		reserve.SplitAmount,
			BalanceA:		reserve.SplitAmount,
			BalanceB:		"0",
			FeeAmount:		reserve.FeeAmount,
		}
		segment.SegmentHash = ComputeVirtualReserveSegmentHash(segment)
		out = append(out, segment.Normalize())
	}
	return normalizeVirtualReserveSegments(out)
}

func ComputeVirtualReserveSegmentHash(segment VirtualReserveSegment) string {
	segment = segment.Normalize()
	return HashParts("virtual-reserve-segment", segment.SegmentID, segment.VirtualChannelID, segment.ParentChannelID, segment.ReserveCommitment, segment.Capacity, segment.BalanceA, segment.BalanceB, segment.FeeAmount)
}

func ValidateVirtualReserveSegments(vc VirtualChannel, segments []VirtualReserveSegment) error {
	vc = vc.Normalize()
	segments = normalizeVirtualReserveSegments(segments)
	if len(segments) == 0 {
		return errors.New("payments virtual reserve segments are required")
	}
	total := sdkmath.ZeroInt()
	seen := make(map[string]struct{}, len(segments))
	for _, segment := range segments {
		if err := segment.ValidateForVirtualChannel(vc); err != nil {
			return err
		}
		if _, found := seen[segment.SegmentID]; found {
			return errors.New("payments virtual reserve segments must be unique")
		}
		seen[segment.SegmentID] = struct{}{}
		capacity, err := parsePositiveInt("payments virtual reserve segment capacity", segment.Capacity)
		if err != nil {
			return err
		}
		total = total.Add(capacity)
	}
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return err
	}
	if !total.Equal(capacity) {
		return errors.New("payments virtual reserve segment split amount must equal capacity")
	}
	return nil
}

func BuildVirtualSegmentSettlementProofs(vc VirtualChannel, segments []VirtualReserveSegment) ([]VirtualSegmentSettlementProof, error) {
	vc = vc.Normalize()
	if err := ValidateVirtualReserveSegments(vc, segments); err != nil {
		return nil, err
	}
	segments = normalizeVirtualReserveSegments(segments)
	out := make([]VirtualSegmentSettlementProof, 0, len(segments))
	for _, segment := range segments {
		proof := VirtualSegmentSettlementProof{
			SegmentID:		segment.SegmentID,
			VirtualChannelID:	vc.VirtualChannelID,
			ParentChannelID:	segment.ParentChannelID,
			FinalStateHash:		vc.StateHash,
			ReserveCommitment:	segment.ReserveCommitment,
			BalanceA:		segment.BalanceA,
			BalanceB:		segment.BalanceB,
		}
		proof.SettlementHash = ComputeVirtualSegmentSettlementHash(proof)
		out = append(out, proof.Normalize())
	}
	return normalizeVirtualSegmentSettlementProofs(out), nil
}

func ComputeVirtualSegmentSettlementHash(proof VirtualSegmentSettlementProof) string {
	proof = proof.Normalize()
	return HashParts("virtual-segment-settlement", proof.SegmentID, proof.VirtualChannelID, proof.ParentChannelID, proof.FinalStateHash, proof.ReserveCommitment, proof.BalanceA, proof.BalanceB)
}

func BuildVirtualPartialActivationFailure(vc VirtualChannel, failedSegmentID, reason string, refundCommitments []string) (VirtualPartialActivationFailure, error) {
	vc = vc.Normalize()
	failure := VirtualPartialActivationFailure{
		VirtualChannelID:	vc.VirtualChannelID,
		FailedSegmentID:	normalizeHash(failedSegmentID),
		Reason:			strings.TrimSpace(reason),
		RefundCommitments:	normalizeHashSlice(refundCommitments),
	}
	failure.FailureHash = ComputeVirtualPartialActivationFailureHash(failure)
	if err := failure.ValidateForVirtualChannel(vc); err != nil {
		return VirtualPartialActivationFailure{}, err
	}
	return failure.Normalize(), nil
}

func ComputeVirtualPartialActivationFailureHash(failure VirtualPartialActivationFailure) string {
	failure = failure.Normalize()
	parts := []string{"virtual-partial-activation-failure", failure.VirtualChannelID, failure.FailedSegmentID, failure.Reason}
	parts = append(parts, failure.RefundCommitments...)
	return HashParts(parts...)
}

func BuildVirtualChannelDisputeProof(latest VirtualChannel, commitments []string, submittedBy string) (VirtualChannelDisputeProof, error) {
	latest = latest.Normalize()
	if latest.StateHash == "" || latest.AnchorCommitment == "" {
		built, err := BuildVirtualChannel(latest)
		if err != nil {
			return VirtualChannelDisputeProof{}, err
		}
		built.Signatures = latest.Signatures
		latest = built.Normalize()
	}
	proof := VirtualChannelDisputeProof{
		VirtualChannelID:		latest.VirtualChannelID,
		ParentRouteID:			latest.ParentRouteID,
		LatestState:			latest,
		ParentReserveCommitments:	normalizeHashSlice(commitments),
		SubmittedBy:			strings.TrimSpace(submittedBy),
	}
	proof.EvidenceHash = ComputeVirtualDisputeEvidenceHash(proof)
	return proof.Normalize(), nil
}

func ComputeVirtualDisputeEvidenceHash(proof VirtualChannelDisputeProof) string {
	proof = proof.Normalize()
	parts := []string{
		"virtual-dispute-proof",
		proof.VirtualChannelID,
		proof.ParentRouteID,
		proof.LatestState.StateHash,
		proof.SubmittedBy,
	}
	parts = append(parts, proof.ParentReserveCommitments...)
	return HashParts(parts...)
}

func ValidateVirtualChannelDisputeProof(proof VirtualChannelDisputeProof, current VirtualChannel) error {
	proof = proof.Normalize()
	current = current.Normalize()
	if proof.VirtualChannelID != current.VirtualChannelID || proof.LatestState.VirtualChannelID != current.VirtualChannelID {
		return errors.New("payments virtual dispute proof channel mismatch")
	}
	if proof.ParentRouteID != current.ParentRouteID || proof.LatestState.ParentRouteID != current.ParentRouteID {
		return errors.New("payments virtual dispute proof route mismatch")
	}
	if !containsString(current.Endpoints, proof.SubmittedBy) && !containsString(current.Intermediaries, proof.SubmittedBy) {
		return errors.New("payments virtual dispute submitter must be route participant")
	}
	if proof.LatestState.Nonce <= current.Nonce {
		return errors.New("payments virtual dispute state nonce must be newer")
	}
	if len(proof.ParentReserveCommitments) != len(current.ParentChannelIDs) {
		return errors.New("payments virtual dispute proof requires parent reserve commitments")
	}
	if len(current.ParentReserveCommitments) > 0 && strings.Join(proof.ParentReserveCommitments, "/") != strings.Join(current.ParentReserveCommitments, "/") {
		return errors.New("payments virtual dispute proof reserve commitment mismatch")
	}
	if err := validateVirtualEndpointUpdate(current, proof.LatestState); err != nil {
		return err
	}
	expected := ComputeVirtualDisputeEvidenceHash(proof)
	if proof.EvidenceHash != expected {
		return errors.New("payments virtual dispute proof evidence hash mismatch")
	}
	return nil
}

func IsGossipMessageType(messageType GossipMessageType) bool {
	switch messageType {
	case GossipChannelAnnouncement,
		GossipChannelUpdate,
		GossipLiquidityHint,
		GossipFeePolicyUpdate,
		GossipNodeAnnouncement,
		GossipRouteFailure,
		GossipCapacityProbe:
		return true
	default:
		return false
	}
}

func ComputeGossipMessageHash(message GossipMessage) string {
	message = message.Normalize()
	parts := []string{
		"gossip-message",
		string(message.MessageType),
		message.ChainID,
		message.ChannelID,
		message.NodeID,
		message.From,
		message.To,
		message.Capacity,
		message.Liquidity,
		message.FeeDenom,
		message.FeeAmount,
		message.MaxFee,
		fmt.Sprintf("%020d", message.ValidAfterHeight),
		fmt.Sprintf("%020d", message.ValidUntilHeight),
		message.ChannelCommitment,
		message.FailureCode,
		message.ProbeAmount,
		fmt.Sprintf("%d", message.ReputationDelta),
		fmt.Sprintf("%020d", message.Sequence),
		fmt.Sprintf("%t", message.Advisory),
	}
	return HashParts(parts...)
}

func validateGossipEdgeFields(message GossipMessage, requireCapacity bool) error {
	message = message.Normalize()
	if message.ChannelID == "" && message.ChannelCommitment == "" {
		return errors.New("payments gossip edge requires channel id or commitment")
	}
	if message.ChannelID != "" {
		if err := ValidateHash("payments gossip channel id", message.ChannelID); err != nil {
			return err
		}
	}
	if err := addressing.ValidateUserAddress("payments gossip from", message.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments gossip to", message.To); err != nil {
		return err
	}
	if message.From == message.To {
		return errors.New("payments gossip endpoints must differ")
	}
	if requireCapacity {
		if err := validatePositiveInt("payments gossip capacity", message.Capacity); err != nil {
			return err
		}
	} else if strings.TrimSpace(message.Capacity) != "" {
		if err := validateNonNegativeInt("payments gossip capacity", message.Capacity); err != nil {
			return err
		}
	}
	if message.FeeDenom != NativeDenom {
		return fmt.Errorf("payments gossip fee denom must be %s", NativeDenom)
	}
	return validateNonNegativeInt("payments gossip fee", message.FeeAmount)
}

func (v VirtualChannel) Normalize() VirtualChannel {
	v.VirtualChannelID = normalizeHash(v.VirtualChannelID)
	v.ChainID = strings.TrimSpace(v.ChainID)
	v.ParentRouteID = normalizeOptionalHash(v.ParentRouteID)
	for i := range v.ParentChannelIDs {
		v.ParentChannelIDs[i] = normalizeHash(v.ParentChannelIDs[i])
	}
	v.ParentReserveCommitments = normalizeHashSlice(v.ParentReserveCommitments)
	v.Endpoints = normalizeAddressSet(v.Endpoints)
	v.EndpointA = strings.TrimSpace(v.EndpointA)
	v.EndpointB = strings.TrimSpace(v.EndpointB)
	if len(v.Endpoints) == 2 {
		if v.EndpointA == "" {
			v.EndpointA = v.Endpoints[0]
		}
		if v.EndpointB == "" {
			v.EndpointB = v.Endpoints[1]
		}
	}
	if v.EndpointA != "" && v.EndpointB != "" {
		v.Endpoints = normalizeAddressSet([]string{v.EndpointA, v.EndpointB})
		v.EndpointA = v.Endpoints[0]
		v.EndpointB = v.Endpoints[1]
	}
	v.Intermediaries = normalizeAddressSet(v.Intermediaries)
	v.IntermediarySetHash = normalizeOptionalHash(v.IntermediarySetHash)
	v.Capacity = strings.TrimSpace(v.Capacity)
	v.BalanceA = strings.TrimSpace(v.BalanceA)
	v.BalanceB = strings.TrimSpace(v.BalanceB)
	v.RoutingFeeAmount = strings.TrimSpace(v.RoutingFeeAmount)
	v.AnchorFeePaid = strings.TrimSpace(v.AnchorFeePaid)
	if v.Nonce == 0 {
		v.Nonce = 1
	}
	if v.BalanceA == "" {
		v.BalanceA = v.Capacity
	}
	if v.BalanceB == "" {
		v.BalanceB = "0"
	}
	if v.RoutingFeeAmount == "" {
		v.RoutingFeeAmount = "0"
	}
	if v.AnchorFeePaid == "" {
		v.AnchorFeePaid = "0"
	}
	v.AnchorCommitment = normalizeOptionalHash(v.AnchorCommitment)
	v.ConditionRoot = normalizeOptionalHash(v.ConditionRoot)
	v.StateHash = normalizeOptionalHash(v.StateHash)
	v.Signatures = normalizeSignatures(v.Signatures)
	if v.Status == "" {
		v.Status = VirtualChannelStatusOpen
	}
	return v
}

func (v VirtualChannel) Validate() error {
	return v.ValidateCore()
}

func (v VirtualChannel) ValidateCore() error {
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
	if vc.ParentRouteID != "" {
		if err := ValidateHash("payments virtual parent route id", vc.ParentRouteID); err != nil {
			return err
		}
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
	if len(vc.ParentReserveCommitments) > 0 {
		if len(vc.ParentReserveCommitments) != len(vc.ParentChannelIDs) {
			return errors.New("payments virtual parent reserve commitments must match parent channels")
		}
		seenCommitments := make(map[string]struct{}, len(vc.ParentReserveCommitments))
		for _, commitment := range vc.ParentReserveCommitments {
			if err := ValidateHash("payments virtual parent reserve commitment", commitment); err != nil {
				return err
			}
			if _, found := seenCommitments[commitment]; found {
				return errors.New("payments virtual parent reserve commitments must be unique")
			}
			seenCommitments[commitment] = struct{}{}
		}
	}
	if err := validateAddressSet("payments virtual endpoint", vc.Endpoints, 2, 2); err != nil {
		return err
	}
	if vc.EndpointA != vc.Endpoints[0] || vc.EndpointB != vc.Endpoints[1] {
		return errors.New("payments virtual endpoints must match endpoint fields")
	}
	if len(vc.Intermediaries) > MaxParticipants {
		return fmt.Errorf("payments virtual intermediaries must be <= %d", MaxParticipants)
	}
	if vc.IntermediarySetHash == "" {
		return errors.New("payments virtual intermediary set hash is required")
	}
	if expected := ComputeParticipantSetHash(vc.Intermediaries); vc.IntermediarySetHash != expected {
		return errors.New("payments virtual intermediary set hash mismatch")
	}
	if err := validatePositiveInt("payments virtual capacity", vc.Capacity); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual balance a", vc.BalanceA); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual balance b", vc.BalanceB); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual routing fee", vc.RoutingFeeAmount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments virtual anchor fee", vc.AnchorFeePaid); err != nil {
		return err
	}
	capacity, err := parsePositiveInt("payments virtual capacity", vc.Capacity)
	if err != nil {
		return err
	}
	balanceA, err := parseNonNegativeInt("payments virtual balance a", vc.BalanceA)
	if err != nil {
		return err
	}
	balanceB, err := parseNonNegativeInt("payments virtual balance b", vc.BalanceB)
	if err != nil {
		return err
	}
	if !balanceA.Add(balanceB).Equal(capacity) {
		return errors.New("payments virtual balances must equal capacity")
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
	if vc.ConditionRoot != "" {
		if err := ValidateHash("payments virtual condition root", vc.ConditionRoot); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments virtual channel state hash", vc.StateHash); err != nil {
		return err
	}
	if expected := ComputeVirtualChannelStateHash(vc); vc.StateHash != expected {
		return errors.New("payments virtual channel state hash mismatch")
	}
	if err := validateVirtualChannelSignatures(vc); err != nil {
		return err
	}
	return nil
}

func validateVirtualChannelSignatures(vc VirtualChannel) error {
	vc = vc.Normalize()
	if len(vc.Signatures) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(vc.Signatures))
	for _, sig := range vc.Signatures {
		if err := ValidateVirtualChannelSignature(sig, vc); err != nil {
			return err
		}
		if _, found := seen[sig.Normalize().Signer]; found {
			return errors.New("payments duplicate virtual channel signature")
		}
		seen[sig.Normalize().Signer] = struct{}{}
	}
	return nil
}

func ValidateVirtualChannelActivation(vc VirtualChannel) error {
	vc = vc.Normalize()
	if err := vc.ValidateCore(); err != nil {
		return err
	}
	required := normalizeAddressSet(append(append([]string{}, vc.Endpoints...), vc.Intermediaries...))
	if len(required) == 0 {
		return errors.New("payments virtual channel activation requires signers")
	}
	seen := make(map[string]struct{}, len(vc.Signatures))
	for _, sig := range vc.Signatures {
		sig = sig.Normalize()
		if err := ValidateVirtualChannelSignature(sig, vc); err != nil {
			return err
		}
		seen[sig.Signer] = struct{}{}
	}
	for _, signer := range required {
		if _, found := seen[signer]; !found {
			return errors.New("payments virtual channel missing required signature")
		}
	}
	return nil
}

func (s VirtualReservationSignature) Normalize() VirtualReservationSignature {
	s.Signer = strings.TrimSpace(s.Signer)
	s.ChainID = strings.TrimSpace(s.ChainID)
	s.VirtualChannelID = normalizeHash(s.VirtualChannelID)
	s.ParentRouteID = normalizeHash(s.ParentRouteID)
	s.ParentChannelID = normalizeHash(s.ParentChannelID)
	s.Capacity = strings.TrimSpace(s.Capacity)
	s.SplitAmount = strings.TrimSpace(s.SplitAmount)
	s.FeeAmount = strings.TrimSpace(s.FeeAmount)
	s.CommitmentHash = normalizeHash(s.CommitmentHash)
	s.SignatureHash = normalizeHash(s.SignatureHash)
	return s
}

func (r VirtualParentReserve) Normalize() VirtualParentReserve {
	r.SegmentID = normalizeOptionalHash(r.SegmentID)
	r.ParentChannelID = normalizeHash(r.ParentChannelID)
	r.ReservedBy = strings.TrimSpace(r.ReservedBy)
	r.Capacity = strings.TrimSpace(r.Capacity)
	if r.SplitAmount == "" {
		r.SplitAmount = r.Capacity
	}
	r.SplitAmount = strings.TrimSpace(r.SplitAmount)
	if r.FeeAmount == "" {
		r.FeeAmount = "0"
	}
	r.FeeAmount = strings.TrimSpace(r.FeeAmount)
	r.ReserveCommitment = normalizeOptionalHash(r.ReserveCommitment)
	r.Signature = r.Signature.Normalize()
	return r
}

func (p VirtualActivationProof) Normalize() VirtualActivationProof {
	p.VirtualChannel = p.VirtualChannel.Normalize()
	p.ParentReserves = normalizeVirtualParentReserves(p.ParentReserves)
	p.ProofHash = normalizeOptionalHash(p.ProofHash)
	return p
}

func (p VirtualChannelDisputeProof) Normalize() VirtualChannelDisputeProof {
	p.VirtualChannelID = normalizeHash(p.VirtualChannelID)
	p.ParentRouteID = normalizeHash(p.ParentRouteID)
	p.LatestState = p.LatestState.Normalize()
	p.ParentReserveCommitments = normalizeHashSlice(p.ParentReserveCommitments)
	p.SubmittedBy = strings.TrimSpace(p.SubmittedBy)
	p.EvidenceHash = normalizeOptionalHash(p.EvidenceHash)
	return p
}

func (r VirtualReserveRelease) Normalize() VirtualReserveRelease {
	r.SegmentID = normalizeOptionalHash(r.SegmentID)
	r.VirtualChannelID = normalizeHash(r.VirtualChannelID)
	r.ParentChannelID = normalizeHash(r.ParentChannelID)
	r.ReserveCommitment = normalizeHash(r.ReserveCommitment)
	r.Capacity = strings.TrimSpace(r.Capacity)
	r.BalanceA = strings.TrimSpace(r.BalanceA)
	r.BalanceB = strings.TrimSpace(r.BalanceB)
	r.FeeAmount = strings.TrimSpace(r.FeeAmount)
	r.ReleaseHash = normalizeOptionalHash(r.ReleaseHash)
	return r
}

func (p VirtualCloseProof) Normalize() VirtualCloseProof {
	p.VirtualChannelID = normalizeHash(p.VirtualChannelID)
	p.ParentRouteID = normalizeHash(p.ParentRouteID)
	p.FinalState = p.FinalState.Normalize()
	p.ParentReserveCommitments = normalizeHashSlice(p.ParentReserveCommitments)
	p.SubmittedBy = strings.TrimSpace(p.SubmittedBy)
	p.ProofHash = normalizeOptionalHash(p.ProofHash)
	return p
}

func (s VirtualReserveSegment) Normalize() VirtualReserveSegment {
	s.SegmentID = normalizeHash(s.SegmentID)
	s.VirtualChannelID = normalizeHash(s.VirtualChannelID)
	s.ParentChannelID = normalizeHash(s.ParentChannelID)
	s.ReserveCommitment = normalizeHash(s.ReserveCommitment)
	s.Capacity = strings.TrimSpace(s.Capacity)
	s.BalanceA = strings.TrimSpace(s.BalanceA)
	s.BalanceB = strings.TrimSpace(s.BalanceB)
	s.FeeAmount = strings.TrimSpace(s.FeeAmount)
	s.SegmentHash = normalizeOptionalHash(s.SegmentHash)
	return s
}

func (s VirtualReserveSegment) ValidateForVirtualChannel(vc VirtualChannel) error {
	s = s.Normalize()
	vc = vc.Normalize()
	if err := ValidateHash("payments virtual reserve segment id", s.SegmentID); err != nil {
		return err
	}
	if s.VirtualChannelID != vc.VirtualChannelID {
		return errors.New("payments virtual reserve segment channel mismatch")
	}
	if !containsString(vc.ParentChannelIDs, s.ParentChannelID) {
		return errors.New("payments virtual reserve segment references unknown parent")
	}
	if err := ValidateHash("payments virtual reserve segment commitment", s.ReserveCommitment); err != nil {
		return err
	}
	capacity, err := parsePositiveInt("payments virtual reserve segment capacity", s.Capacity)
	if err != nil {
		return err
	}
	balanceA, err := parseNonNegativeInt("payments virtual reserve segment balance a", s.BalanceA)
	if err != nil {
		return err
	}
	balanceB, err := parseNonNegativeInt("payments virtual reserve segment balance b", s.BalanceB)
	if err != nil {
		return err
	}
	if !balanceA.Add(balanceB).Equal(capacity) {
		return errors.New("payments virtual reserve segment balances must equal capacity")
	}
	if err := validateNonNegativeInt("payments virtual reserve segment fee", s.FeeAmount); err != nil {
		return err
	}
	if expected := ComputeVirtualReserveSegmentHash(s); s.SegmentHash != expected {
		return errors.New("payments virtual reserve segment hash mismatch")
	}
	return nil
}

func (p VirtualSegmentSettlementProof) Normalize() VirtualSegmentSettlementProof {
	p.SegmentID = normalizeHash(p.SegmentID)
	p.VirtualChannelID = normalizeHash(p.VirtualChannelID)
	p.ParentChannelID = normalizeHash(p.ParentChannelID)
	p.FinalStateHash = normalizeHash(p.FinalStateHash)
	p.ReserveCommitment = normalizeHash(p.ReserveCommitment)
	p.BalanceA = strings.TrimSpace(p.BalanceA)
	p.BalanceB = strings.TrimSpace(p.BalanceB)
	p.SettlementHash = normalizeOptionalHash(p.SettlementHash)
	return p
}

func (p VirtualSegmentSettlementProof) ValidateForSegment(segment VirtualReserveSegment, vc VirtualChannel) error {
	p = p.Normalize()
	segment = segment.Normalize()
	vc = vc.Normalize()
	if p.SegmentID != segment.SegmentID || p.VirtualChannelID != vc.VirtualChannelID || p.ParentChannelID != segment.ParentChannelID {
		return errors.New("payments virtual segment settlement proof domain mismatch")
	}
	if p.FinalStateHash != vc.StateHash || p.ReserveCommitment != segment.ReserveCommitment {
		return errors.New("payments virtual segment settlement proof commitment mismatch")
	}
	if p.BalanceA != segment.BalanceA || p.BalanceB != segment.BalanceB {
		return errors.New("payments virtual segment settlement proof balance mismatch")
	}
	if expected := ComputeVirtualSegmentSettlementHash(p); p.SettlementHash != expected {
		return errors.New("payments virtual segment settlement proof hash mismatch")
	}
	return nil
}

func (f VirtualPartialActivationFailure) Normalize() VirtualPartialActivationFailure {
	f.VirtualChannelID = normalizeHash(f.VirtualChannelID)
	f.FailedSegmentID = normalizeHash(f.FailedSegmentID)
	f.Reason = strings.TrimSpace(f.Reason)
	f.RefundCommitments = normalizeHashSlice(f.RefundCommitments)
	f.FailureHash = normalizeOptionalHash(f.FailureHash)
	return f
}

func (f VirtualPartialActivationFailure) ValidateForVirtualChannel(vc VirtualChannel) error {
	f = f.Normalize()
	vc = vc.Normalize()
	if f.VirtualChannelID != vc.VirtualChannelID {
		return errors.New("payments virtual partial activation failure channel mismatch")
	}
	if err := ValidateHash("payments virtual failed segment id", f.FailedSegmentID); err != nil {
		return err
	}
	if f.Reason == "" {
		return errors.New("payments virtual partial activation failure reason is required")
	}
	if len(f.RefundCommitments) == 0 {
		return errors.New("payments virtual partial activation failure refund commitments are required")
	}
	if expected := ComputeVirtualPartialActivationFailureHash(f); f.FailureHash != expected {
		return errors.New("payments virtual partial activation failure hash mismatch")
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
		BatchID:	normalizeHash(batchID),
		Operations:	SortSettlementOperations(operations),
	}
	batch.RootHash = ComputeBatchRoot(batch.Operations)
	if err := batch.Validate(); err != nil {
		return SettlementBatch{}, err
	}
	return batch, nil
}

func GroupSettlementOperationsByChannelKey(seed string, operations []SettlementOperation) ([]SettlementBatch, error) {
	operations = SortSettlementOperations(operations)
	if len(operations) == 0 {
		return nil, errors.New("payments settlement batch grouping requires operations")
	}
	groups := [][]SettlementOperation{}
	groupChannels := []map[string]struct{}{}
	for _, op := range operations {
		if err := op.Validate(); err != nil {
			return nil, err
		}
		placed := false
		for i := range groups {
			if _, found := groupChannels[i][op.ChannelID]; found {
				continue
			}
			groups[i] = append(groups[i], op)
			groupChannels[i][op.ChannelID] = struct{}{}
			placed = true
			break
		}
		if !placed {
			groups = append(groups, []SettlementOperation{op})
			groupChannels = append(groupChannels, map[string]struct{}{op.ChannelID: {}})
		}
	}
	out := make([]SettlementBatch, 0, len(groups))
	seed = strings.TrimSpace(seed)
	if seed == "" {
		seed = "settlement-batch-group"
	}
	for i, group := range groups {
		batchID := HashParts(seed, fmt.Sprintf("%020d", uint64(i)))
		batch, err := NewSettlementBatch(batchID, group)
		if err != nil {
			return nil, err
		}
		out = append(out, batch)
	}
	return out, nil
}

func AccessPlanForSettlementOperation(op SettlementOperation, blockHeight uint64) (BlockSTMAccessPlan, error) {
	op = op.Normalize()
	if err := op.Validate(); err != nil {
		return BlockSTMAccessPlan{}, err
	}
	txClass := blockSTMClassForBatchOperation(op.OperationType)
	plan := BlockSTMAccessPlan{
		OperationID:		op.OperationID,
		TxClass:		txClass,
		ChannelID:		op.ChannelID,
		ReadKeys:		[]string{PaymentChannelKey(op.ChannelID)},
		WriteKeys:		[]string{PaymentChannelKey(op.ChannelID)},
		AccumulatorKeys:	[]string{PaymentBlockAccumulatorKey(blockHeight)},
		ConflictDomain:		PaymentChannelKey(op.ChannelID),
		DeterministicGroup:	PaymentChannelKey(op.ChannelID),
	}
	switch op.OperationType {
	case BatchOperationOpen:
		plan.WriteKeys = append(plan.WriteKeys, PaymentCustodyKey(op.ChannelID))
	case BatchOperationClose, BatchOperationDispute:
		plan.WriteKeys = append(plan.WriteKeys, PaymentPendingCloseIndexKey(op.ChannelID))
	case BatchOperationSettle:
		plan.WriteKeys = append(plan.WriteKeys, PaymentSettlementKey(op.ChannelID), PaymentSettlementTombstoneKey(op.ChannelID), PaymentCustodyKey(op.ChannelID))
	}
	return plan.Normalize(), nil
}

func AccessPlanForConditionResolution(channelID string, conditionIDs []string, blockHeight uint64) (BlockSTMAccessPlan, error) {
	channelID = normalizeHash(channelID)
	if err := ValidateHash("payments blockstm condition channel id", channelID); err != nil {
		return BlockSTMAccessPlan{}, err
	}
	conditionIDs = normalizeHashSlice(conditionIDs)
	if len(conditionIDs) == 0 {
		return BlockSTMAccessPlan{}, errors.New("payments blockstm condition resolution requires condition ids")
	}
	readKeys := []string{PaymentChannelKey(channelID)}
	writeKeys := []string{PaymentChannelKey(channelID)}
	for _, conditionID := range conditionIDs {
		writeKeys = append(writeKeys, PaymentConditionIndexKey(channelID, conditionID))
	}
	plan := BlockSTMAccessPlan{
		OperationID:		HashParts("condition-resolution", channelID, strings.Join(conditionIDs, "/")),
		TxClass:		BlockSTMClassResolveCondition,
		ChannelID:		channelID,
		ConditionIDs:		conditionIDs,
		ReadKeys:		readKeys,
		WriteKeys:		writeKeys,
		AccumulatorKeys:	[]string{PaymentBlockAccumulatorKey(blockHeight)},
		ConflictDomain:		PaymentChannelKey(channelID),
		DeterministicGroup:	PaymentChannelKey(channelID),
	}
	return plan.Normalize(), nil
}

func ProfileBlockSTMConflicts(plans []BlockSTMAccessPlan) BlockSTMConflictProfile {
	normalized := make([]BlockSTMAccessPlan, len(plans))
	for i, plan := range plans {
		normalized[i] = plan.Normalize()
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i].OperationID < normalized[j].OperationID
	})
	conflicts := []BlockSTMConflict{}
	for i := range normalized {
		for j := i + 1; j < len(normalized); j++ {
			for _, conflict := range blockSTMPlanConflicts(normalized[i], normalized[j]) {
				conflicts = append(conflicts, conflict)
			}
		}
	}
	return BlockSTMConflictProfile{
		Plans:				normalized,
		Conflicts:			conflicts,
		ParallelizableGroups:		blockSTMParallelizableGroups(normalized),
		ConflictFree:			len(conflicts) == 0,
		GlobalAccountingDeferred:	blockSTMAccountingDeferred(normalized),
	}
}

func AccumulatePaymentBlockAccounting(acc PaymentBlockAccumulator, settlement SettlementRecord) (PaymentBlockAccumulator, error) {
	acc = acc.Normalize()
	settlement = settlement.Normalize()
	fee, err := parseNonNegativeInt("payments block accumulator fee", acc.FeeAmount)
	if err != nil {
		return PaymentBlockAccumulator{}, err
	}
	burn, err := parseNonNegativeInt("payments block accumulator burn", acc.BurnAmount)
	if err != nil {
		return PaymentBlockAccumulator{}, err
	}
	penalty, err := parseNonNegativeInt("payments block accumulator penalty", acc.PenaltyAmount)
	if err != nil {
		return PaymentBlockAccumulator{}, err
	}
	settlementFee, err := parseNonNegativeInt("payments settlement fee", settlement.SettlementFee)
	if err != nil {
		return PaymentBlockAccumulator{}, err
	}
	penaltyTotal, err := sumPenaltyAllocations(settlement.PenaltyAllocations)
	if err != nil {
		return PaymentBlockAccumulator{}, err
	}
	acc.FeeAmount = fee.Add(settlementFee).String()
	acc.PenaltyAmount = penalty.Add(penaltyTotal).String()
	acc.BurnAmount = burn.String()
	acc.OperationCount++
	return acc.Normalize(), nil
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

func (p BlockSTMAccessPlan) Normalize() BlockSTMAccessPlan {
	p.OperationID = normalizeOptionalHash(p.OperationID)
	p.ChannelID = normalizeOptionalHash(p.ChannelID)
	p.ConditionIDs = normalizeHashSlice(p.ConditionIDs)
	p.ReadKeys = normalizeStoreKeySlice(p.ReadKeys)
	p.WriteKeys = normalizeStoreKeySlice(p.WriteKeys)
	p.AccumulatorKeys = normalizeStoreKeySlice(p.AccumulatorKeys)
	p.ConflictDomain = strings.TrimSpace(p.ConflictDomain)
	p.DeterministicGroup = strings.TrimSpace(p.DeterministicGroup)
	if p.ConflictDomain == "" && p.ChannelID != "" {
		p.ConflictDomain = PaymentChannelKey(p.ChannelID)
	}
	if p.DeterministicGroup == "" {
		p.DeterministicGroup = p.ConflictDomain
	}
	return p
}

func (p BlockSTMAccessPlan) Validate() error {
	p = p.Normalize()
	if err := ValidateHash("payments blockstm operation id", p.OperationID); err != nil {
		return err
	}
	if !IsBlockSTMTransactionClass(p.TxClass) {
		return fmt.Errorf("unknown payments blockstm transaction class %q", p.TxClass)
	}
	if p.ChannelID != "" {
		if err := ValidateHash("payments blockstm channel id", p.ChannelID); err != nil {
			return err
		}
	}
	if len(p.WriteKeys) == 0 {
		return errors.New("payments blockstm access plan requires write keys")
	}
	if p.ConflictDomain == "" {
		return errors.New("payments blockstm access plan requires conflict domain")
	}
	return nil
}

func (c BlockSTMConflict) Normalize() BlockSTMConflict {
	c.LeftOperationID = normalizeOptionalHash(c.LeftOperationID)
	c.RightOperationID = normalizeOptionalHash(c.RightOperationID)
	c.Key = strings.TrimSpace(c.Key)
	c.Reason = strings.TrimSpace(c.Reason)
	return c
}

func (a PaymentBlockAccumulator) Normalize() PaymentBlockAccumulator {
	a.FeeAmount = strings.TrimSpace(a.FeeAmount)
	if a.FeeAmount == "" {
		a.FeeAmount = "0"
	}
	a.BurnAmount = strings.TrimSpace(a.BurnAmount)
	if a.BurnAmount == "" {
		a.BurnAmount = "0"
	}
	a.PenaltyAmount = strings.TrimSpace(a.PenaltyAmount)
	if a.PenaltyAmount == "" {
		a.PenaltyAmount = "0"
	}
	if a.AccumulatorKey == "" && a.BlockHeight > 0 {
		a.AccumulatorKey = PaymentBlockAccumulatorKey(a.BlockHeight)
	}
	a.AccumulatorKey = strings.TrimSpace(a.AccumulatorKey)
	return a
}

func (a PaymentBlockAccumulator) Validate() error {
	a = a.Normalize()
	if a.BlockHeight == 0 {
		return errors.New("payments block accumulator height must be positive")
	}
	if a.AccumulatorKey != PaymentBlockAccumulatorKey(a.BlockHeight) {
		return errors.New("payments block accumulator key mismatch")
	}
	if err := validateNonNegativeInt("payments block accumulator fees", a.FeeAmount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments block accumulator burns", a.BurnAmount); err != nil {
		return err
	}
	return validateNonNegativeInt("payments block accumulator penalties", a.PenaltyAmount)
}

func (r StoreV2ChannelRecord) Normalize() StoreV2ChannelRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Channel = r.Channel.Normalize()
	r.LatestStateHash = normalizeOptionalHash(r.LatestStateHash)
	r.PendingCloseKey = strings.TrimSpace(r.PendingCloseKey)
	r.ParticipantIndexKeys = normalizeStoreKeySlice(r.ParticipantIndexKeys)
	r.RoutingAdvertisementKey = strings.TrimSpace(r.RoutingAdvertisementKey)
	return r
}

func (r StoreV2ChannelRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 channel record version mismatch")
	}
	if r.Key != StoreV2ChannelKey(r.ChannelID) {
		return errors.New("payments store v2 channel key mismatch")
	}
	if r.Channel.ChannelID != r.ChannelID {
		return errors.New("payments store v2 channel id mismatch")
	}
	if err := ValidateHash("payments store v2 latest state hash", r.LatestStateHash); err != nil {
		return err
	}
	if r.LatestStateNonce == 0 {
		return errors.New("payments store v2 latest state nonce must be positive")
	}
	if len(r.Channel.LatestState.Signatures) != 0 {
		return errors.New("payments store v2 active channel record must stay compact")
	}
	for _, key := range r.ParticipantIndexKeys {
		if !strings.HasPrefix(key, paymentKey(StoreV2KeyParticipantChannelsPrefix)+"/") {
			return errors.New("payments store v2 participant index key prefix mismatch")
		}
	}
	return nil
}

func (r StoreV2ChannelStateRecord) Normalize() StoreV2ChannelStateRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.ChannelID = normalizeHash(r.ChannelID)
	r.StateHash = normalizeOptionalHash(r.StateHash)
	r.FullState = r.FullState.Normalize()
	return r
}

func (r StoreV2ChannelStateRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 channel state record version mismatch")
	}
	if r.Key != StoreV2ChannelStateKey(r.ChannelID, r.Nonce) {
		return errors.New("payments store v2 channel state key mismatch")
	}
	if r.Nonce == 0 {
		return errors.New("payments store v2 channel state nonce must be positive")
	}
	if err := ValidateHash("payments store v2 channel state hash", r.StateHash); err != nil {
		return err
	}
	if r.FullState.StateHash != "" && r.FullState.StateHash != r.StateHash {
		return errors.New("payments store v2 channel state hash mismatch")
	}
	if !r.SubmittedOnChain && len(r.FullState.Signatures) != 0 {
		return errors.New("payments store v2 off-chain checkpoint stores hash only")
	}
	return nil
}

func (r StoreV2PendingCloseRecord) Normalize() StoreV2PendingCloseRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Close = r.Close.Normalize()
	return r
}

func (r StoreV2PendingCloseRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 pending close record version mismatch")
	}
	if r.Key != StoreV2PendingCloseKey(r.ChannelID) {
		return errors.New("payments store v2 pending close key mismatch")
	}
	if r.Close.State.ChannelID != r.ChannelID {
		return errors.New("payments store v2 pending close channel mismatch")
	}
	return nil
}

func (r StoreV2ConditionRecord) Normalize() StoreV2ConditionRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.ConditionID = normalizeHash(r.ConditionID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Promise = r.Promise.Normalize()
	r.ClaimEvidence = normalizeOptionalHash(r.ClaimEvidence)
	return r
}

func (r StoreV2ConditionRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 condition record version mismatch")
	}
	if r.Key != StoreV2ConditionKey(r.ConditionID) {
		return errors.New("payments store v2 condition key mismatch")
	}
	if err := ValidateHash("payments store v2 condition id", r.ConditionID); err != nil {
		return err
	}
	return ValidateHash("payments store v2 condition channel id", r.ChannelID)
}

func (r StoreV2VirtualChannelRecord) Normalize() StoreV2VirtualChannelRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.VirtualChannelID = normalizeHash(r.VirtualChannelID)
	r.Channel = r.Channel.Normalize()
	r.AnchorHash = normalizeOptionalHash(r.AnchorHash)
	return r
}

func (r StoreV2VirtualChannelRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 virtual channel record version mismatch")
	}
	if r.Key != StoreV2VirtualChannelKey(r.VirtualChannelID) {
		return errors.New("payments store v2 virtual channel key mismatch")
	}
	if r.Channel.VirtualChannelID != r.VirtualChannelID {
		return errors.New("payments store v2 virtual channel id mismatch")
	}
	return ValidateHash("payments store v2 virtual channel anchor", r.AnchorHash)
}

func (r StoreV2ParticipantChannelRecord) Normalize() StoreV2ParticipantChannelRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.Participant = strings.TrimSpace(r.Participant)
	r.ChannelID = normalizeHash(r.ChannelID)
	return r
}

func (r StoreV2ParticipantChannelRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 participant channel version mismatch")
	}
	if err := addressing.ValidateUserAddress("payments store v2 participant", r.Participant); err != nil {
		return err
	}
	if r.Key != StoreV2ParticipantChannelKey(r.Participant, r.ChannelID) {
		return errors.New("payments store v2 participant channel key mismatch")
	}
	return nil
}

func (r StoreV2SettlementTombstoneRecord) Normalize() StoreV2SettlementTombstoneRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Tombstone = r.Tombstone.Normalize()
	if r.PruneAfterHeight == 0 {
		r.PruneAfterHeight = r.Tombstone.ExpiresHeight
	}
	return r
}

func (r StoreV2SettlementTombstoneRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 tombstone version mismatch")
	}
	if r.Key != StoreV2SettlementTombstoneKey(r.ChannelID) {
		return errors.New("payments store v2 tombstone key mismatch")
	}
	if r.Tombstone.ChannelID != r.ChannelID {
		return errors.New("payments store v2 tombstone channel mismatch")
	}
	if r.PruneAfterHeight < r.Tombstone.ClosedHeight {
		return errors.New("payments store v2 tombstone prune height before close")
	}
	return nil
}

func (r StoreV2FeeAccumulatorRecord) Normalize() StoreV2FeeAccumulatorRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.BlockOrEpoch = strings.TrimSpace(r.BlockOrEpoch)
	r.Bucket = strings.TrimSpace(r.Bucket)
	r.Amount = strings.TrimSpace(r.Amount)
	if r.Amount == "" {
		r.Amount = "0"
	}
	return r
}

func (r StoreV2FeeAccumulatorRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 fee accumulator version mismatch")
	}
	if r.Key != StoreV2FeeAccumulatorKey(r.BlockOrEpoch, r.Bucket) {
		return errors.New("payments store v2 fee accumulator key mismatch")
	}
	return validateNonNegativeInt("payments store v2 fee accumulator amount", r.Amount)
}

func (r StoreV2FraudProofRecord) Normalize() StoreV2FraudProofRecord {
	r.Key = strings.TrimSpace(r.Key)
	if r.Version == 0 {
		r.Version = StoreV2MigrationVersion
	}
	r.ProofID = normalizeHash(r.ProofID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.Proof = r.Proof.Normalize()
	return r
}

func (r StoreV2FraudProofRecord) Validate() error {
	r = r.Normalize()
	if r.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 fraud proof version mismatch")
	}
	if r.Key != StoreV2FraudProofKey(r.ProofID) {
		return errors.New("payments store v2 fraud proof key mismatch")
	}
	if r.Proof.ProofID != r.ProofID {
		return errors.New("payments store v2 fraud proof id mismatch")
	}
	return ValidateHash("payments store v2 fraud proof channel id", r.ChannelID)
}

func (l StoreV2Layout) Normalize() StoreV2Layout {
	if l.Version == 0 {
		l.Version = StoreV2MigrationVersion
	}
	for i := range l.Channels {
		l.Channels[i] = l.Channels[i].Normalize()
	}
	for i := range l.ChannelStates {
		l.ChannelStates[i] = l.ChannelStates[i].Normalize()
	}
	for i := range l.PendingCloses {
		l.PendingCloses[i] = l.PendingCloses[i].Normalize()
	}
	for i := range l.Conditions {
		l.Conditions[i] = l.Conditions[i].Normalize()
	}
	for i := range l.VirtualChannels {
		l.VirtualChannels[i] = l.VirtualChannels[i].Normalize()
	}
	for i := range l.ParticipantChannels {
		l.ParticipantChannels[i] = l.ParticipantChannels[i].Normalize()
	}
	for i := range l.SettlementTombstones {
		l.SettlementTombstones[i] = l.SettlementTombstones[i].Normalize()
	}
	for i := range l.FeeAccumulators {
		l.FeeAccumulators[i] = l.FeeAccumulators[i].Normalize()
	}
	for i := range l.FraudProofs {
		l.FraudProofs[i] = l.FraudProofs[i].Normalize()
	}
	sortStoreV2Layout(&l)
	return l
}

func (l StoreV2Layout) Validate() error {
	l = l.Normalize()
	if l.Version != StoreV2MigrationVersion {
		return errors.New("payments store v2 layout version mismatch")
	}
	seen := map[string]struct{}{}
	validateKey := func(key string) error {
		if key == "" {
			return errors.New("payments store v2 key is required")
		}
		if _, found := seen[key]; found {
			return fmt.Errorf("payments store v2 duplicate key %s", key)
		}
		seen[key] = struct{}{}
		return nil
	}
	for _, record := range l.Channels {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.ChannelStates {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.PendingCloses {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.Conditions {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.VirtualChannels {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.ParticipantChannels {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.SettlementTombstones {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.FeeAccumulators {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	for _, record := range l.FraudProofs {
		if err := validateKey(record.Key); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s AdaptiveSyncSnapshot) Normalize() AdaptiveSyncSnapshot {
	s.Key = strings.TrimSpace(s.Key)
	if s.Version == 0 {
		s.Version = StoreV2MigrationVersion
	}
	s.Layout = s.Layout.Normalize()
	for i := range s.ActiveDisputes {
		s.ActiveDisputes[i] = s.ActiveDisputes[i].Normalize()
	}
	for i := range s.PendingFinalizations {
		s.PendingFinalizations[i] = s.PendingFinalizations[i].Normalize()
	}
	for i := range s.WatcherReplayEvents {
		s.WatcherReplayEvents[i] = s.WatcherReplayEvents[i].Normalize()
	}
	sort.SliceStable(s.ActiveDisputes, func(i, j int) bool { return s.ActiveDisputes[i].Key < s.ActiveDisputes[j].Key })
	sort.SliceStable(s.PendingFinalizations, func(i, j int) bool { return s.PendingFinalizations[i].Key < s.PendingFinalizations[j].Key })
	sort.SliceStable(s.WatcherReplayEvents, func(i, j int) bool { return s.WatcherReplayEvents[i].Key < s.WatcherReplayEvents[j].Key })
	s.SnapshotHash = normalizeOptionalHash(s.SnapshotHash)
	return s
}

func (s AdaptiveSyncSnapshot) Validate() error {
	s = s.Normalize()
	if s.Version != StoreV2MigrationVersion {
		return errors.New("payments adaptive sync snapshot version mismatch")
	}
	if s.Height == 0 {
		return errors.New("payments adaptive sync snapshot height must be positive")
	}
	if s.Key != StoreV2AdaptiveSnapshotKey(s.Height) {
		return errors.New("payments adaptive sync snapshot key mismatch")
	}
	if !s.ConsensusOnly || !s.RoutingTopologyExcluded {
		return errors.New("payments adaptive sync snapshot must exclude routing topology")
	}
	if err := s.Layout.Validate(); err != nil {
		return err
	}
	pendingByChannel := map[string]StoreV2PendingCloseRecord{}
	for _, pending := range s.Layout.PendingCloses {
		pendingByChannel[pending.ChannelID] = pending
	}
	seen := map[string]struct{}{}
	for _, dispute := range s.ActiveDisputes {
		if err := dispute.Validate(); err != nil {
			return err
		}
		if _, found := pendingByChannel[dispute.ChannelID]; !found {
			return errors.New("payments adaptive sync active dispute missing pending close")
		}
		if _, found := seen[dispute.Key]; found {
			return errors.New("payments adaptive sync duplicate active dispute")
		}
		seen[dispute.Key] = struct{}{}
	}
	seen = map[string]struct{}{}
	for _, pending := range s.PendingFinalizations {
		if err := pending.Validate(); err != nil {
			return err
		}
		if _, found := pendingByChannel[pending.ChannelID]; !found {
			return errors.New("payments adaptive sync pending finalization missing pending close")
		}
		if _, found := seen[pending.Key]; found {
			return errors.New("payments adaptive sync duplicate pending finalization")
		}
		seen[pending.Key] = struct{}{}
	}
	seen = map[string]struct{}{}
	for _, event := range s.WatcherReplayEvents {
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := seen[event.Key]; found {
			return errors.New("payments adaptive sync duplicate watcher replay event")
		}
		seen[event.Key] = struct{}{}
	}
	if s.SnapshotHash == "" {
		return errors.New("payments adaptive sync snapshot hash is required")
	}
	if expected := ComputeAdaptiveSyncSnapshotHash(s); s.SnapshotHash != expected {
		return errors.New("payments adaptive sync snapshot hash mismatch")
	}
	return nil
}

func (i AdaptiveSyncActiveDisputeIndex) Normalize() AdaptiveSyncActiveDisputeIndex {
	i.Key = strings.TrimSpace(i.Key)
	i.ChannelID = normalizeHash(i.ChannelID)
	i.PendingStateHash = normalizeHash(i.PendingStateHash)
	i.Submitter = strings.TrimSpace(i.Submitter)
	return i
}

func (i AdaptiveSyncActiveDisputeIndex) Validate() error {
	i = i.Normalize()
	if i.Key != StoreV2ActiveDisputeKey(i.ChannelID) {
		return errors.New("payments adaptive sync active dispute key mismatch")
	}
	if err := ValidateHash("payments adaptive sync active dispute channel id", i.ChannelID); err != nil {
		return err
	}
	if err := ValidateHash("payments adaptive sync active dispute state hash", i.PendingStateHash); err != nil {
		return err
	}
	if i.PendingNonce == 0 || i.SubmittedHeight == 0 || i.SettleAfterHeight == 0 {
		return errors.New("payments adaptive sync active dispute heights and nonce must be positive")
	}
	if i.DisputeCount == 0 {
		return errors.New("payments adaptive sync active dispute count must be positive")
	}
	return addressing.ValidateUserAddress("payments adaptive sync active dispute submitter", i.Submitter)
}

func (i AdaptiveSyncPendingFinalizationIndex) Normalize() AdaptiveSyncPendingFinalizationIndex {
	i.Key = strings.TrimSpace(i.Key)
	i.ChannelID = normalizeHash(i.ChannelID)
	i.PendingStateHash = normalizeHash(i.PendingStateHash)
	return i
}

func (i AdaptiveSyncPendingFinalizationIndex) Validate() error {
	i = i.Normalize()
	if i.Key != StoreV2PendingFinalizationKey(i.ChannelID) {
		return errors.New("payments adaptive sync pending finalization key mismatch")
	}
	if err := ValidateHash("payments adaptive sync pending finalization channel id", i.ChannelID); err != nil {
		return err
	}
	if !IsChannelFinality(i.Finality) {
		return fmt.Errorf("unknown payments adaptive sync pending finalization finality %q", i.Finality)
	}
	if i.PendingHeight == 0 || i.PendingNonce == 0 {
		return errors.New("payments adaptive sync pending finalization height and nonce must be positive")
	}
	return ValidateHash("payments adaptive sync pending finalization state hash", i.PendingStateHash)
}

func (e AdaptiveSyncWatcherReplayEvent) Normalize() AdaptiveSyncWatcherReplayEvent {
	e.Key = strings.TrimSpace(e.Key)
	e.Event = e.Event.Normalize()
	e.EventHash = normalizeOptionalHash(e.EventHash)
	return e
}

func (e AdaptiveSyncWatcherReplayEvent) Validate() error {
	e = e.Normalize()
	if err := e.Event.Validate(); err != nil {
		return err
	}
	if e.Key != StoreV2WatcherReplayEventKey(e.Event.Height, e.Event.EventID) {
		return errors.New("payments adaptive sync watcher replay event key mismatch")
	}
	if e.EventHash != AdaptiveSyncEventHash(e.Event) {
		return errors.New("payments adaptive sync watcher replay event hash mismatch")
	}
	return nil
}

func AdaptiveSyncEventHash(event PaymentEvent) string {
	event = event.Normalize()
	parts := []string{"adaptive-sync-event", event.EventID, event.EventType, event.ChannelID, fmt.Sprintf("%020d", event.Height)}
	for _, attr := range event.Attributes {
		parts = append(parts, attr.Key, attr.Value)
	}
	return HashParts(parts...)
}

func ComputeAdaptiveSyncSnapshotHash(snapshot AdaptiveSyncSnapshot) string {
	snapshot = snapshot.Normalize()
	parts := []string{"adaptive-sync-snapshot", snapshot.Key, fmt.Sprintf("%020d", snapshot.Height), fmt.Sprintf("%t", snapshot.ConsensusOnly), fmt.Sprintf("%t", snapshot.RoutingTopologyExcluded)}
	for _, record := range snapshot.Layout.Channels {
		parts = append(parts, record.Key, record.LatestStateHash)
	}
	for _, record := range snapshot.Layout.PendingCloses {
		parts = append(parts, record.Key, record.Close.State.StateHash)
	}
	for _, record := range snapshot.Layout.Conditions {
		parts = append(parts, record.Key, fmt.Sprintf("%t", record.Settled))
	}
	for _, record := range snapshot.Layout.VirtualChannels {
		parts = append(parts, record.Key, record.AnchorHash)
	}
	for _, record := range snapshot.Layout.SettlementTombstones {
		parts = append(parts, record.Key, record.Tombstone.StateHash)
	}
	for _, dispute := range snapshot.ActiveDisputes {
		parts = append(parts, dispute.Key, dispute.PendingStateHash, fmt.Sprintf("%020d", dispute.PendingNonce))
	}
	for _, pending := range snapshot.PendingFinalizations {
		parts = append(parts, pending.Key, pending.PendingStateHash, fmt.Sprintf("%020d", pending.PendingHeight))
	}
	for _, event := range snapshot.WatcherReplayEvents {
		parts = append(parts, event.Key, event.EventHash)
	}
	return HashParts(parts...)
}

func blockSTMClassForBatchOperation(op BatchOperationType) BlockSTMTransactionClass {
	switch op {
	case BatchOperationOpen:
		return BlockSTMClassOpenChannel
	case BatchOperationClose:
		return BlockSTMClassCloseChannel
	case BatchOperationDispute:
		return BlockSTMClassDisputeChannel
	case BatchOperationSettle:
		return BlockSTMClassSettleChannel
	default:
		return ""
	}
}

func blockSTMPlanConflicts(left, right BlockSTMAccessPlan) []BlockSTMConflict {
	left = left.Normalize()
	right = right.Normalize()
	conflicts := []BlockSTMConflict{}
	leftWrites := setFromStrings(left.WriteKeys)
	rightWrites := setFromStrings(right.WriteKeys)
	for key := range leftWrites {
		if _, found := rightWrites[key]; found {
			conflicts = append(conflicts, BlockSTMConflict{
				LeftOperationID:	left.OperationID,
				RightOperationID:	right.OperationID,
				Key:			key,
				Reason:			"write/write",
			}.Normalize())
		}
	}
	for _, key := range left.WriteKeys {
		if containsString(right.ReadKeys, key) && !containsString(left.AccumulatorKeys, key) && !containsString(right.AccumulatorKeys, key) {
			conflicts = append(conflicts, BlockSTMConflict{LeftOperationID: left.OperationID, RightOperationID: right.OperationID, Key: key, Reason: "write/read"}.Normalize())
		}
	}
	for _, key := range right.WriteKeys {
		if containsString(left.ReadKeys, key) && !containsString(left.AccumulatorKeys, key) && !containsString(right.AccumulatorKeys, key) {
			conflicts = append(conflicts, BlockSTMConflict{LeftOperationID: left.OperationID, RightOperationID: right.OperationID, Key: key, Reason: "read/write"}.Normalize())
		}
	}
	return conflicts
}

func blockSTMParallelizableGroups(plans []BlockSTMAccessPlan) [][]string {
	groups := [][]string{}
	groupDomains := []map[string]struct{}{}
	for _, plan := range plans {
		plan = plan.Normalize()
		placed := false
		for i := range groups {
			if _, found := groupDomains[i][plan.ConflictDomain]; found {
				continue
			}
			groups[i] = append(groups[i], plan.OperationID)
			groupDomains[i][plan.ConflictDomain] = struct{}{}
			placed = true
			break
		}
		if !placed {
			groups = append(groups, []string{plan.OperationID})
			groupDomains = append(groupDomains, map[string]struct{}{plan.ConflictDomain: {}})
		}
	}
	return groups
}

func blockSTMAccountingDeferred(plans []BlockSTMAccessPlan) bool {
	for _, plan := range plans {
		plan = plan.Normalize()
		for _, accKey := range plan.AccumulatorKeys {
			if containsString(plan.WriteKeys, accKey) {
				return false
			}
		}
	}
	return true
}

func setFromStrings(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = struct{}{}
		}
	}
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

func IsBlockSTMTransactionClass(value BlockSTMTransactionClass) bool {
	switch value {
	case BlockSTMClassOpenChannel,
		BlockSTMClassUpdateCheckpoint,
		BlockSTMClassCloseChannel,
		BlockSTMClassDisputeChannel,
		BlockSTMClassSettleChannel,
		BlockSTMClassResolveCondition,
		BlockSTMClassBatchConditions,
		BlockSTMClassPenaltyAccounting:
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

func IsVirtualCloseMode(value VirtualCloseMode) bool {
	switch value {
	case VirtualCloseModeCooperative,
		VirtualCloseModeExpired,
		VirtualCloseModeIntermediaryRisk,
		VirtualCloseModeDisputed:
		return true
	default:
		return false
	}
}

func IsPenaltyRoute(value PenaltyRoute) bool {
	switch value {
	case PenaltyRouteReporter, PenaltyRouteCounterparty, PenaltyRouteBurn, PenaltyRouteSecurityReserve, PenaltyRouteCommunityPool:
		return true
	default:
		return false
	}
}

func IsPaymentPenaltyClass(value PaymentPenaltyClass) bool {
	switch value {
	case PenaltyClassInvalidClose,
		PenaltyClassStaleClose,
		PenaltyClassDoubleSign,
		PenaltyClassInvalidCondition,
		PenaltyClassReplayAttempt,
		PenaltyClassAsyncOverexposure,
		PenaltyClassInvalidFraudProof:
		return true
	default:
		return false
	}
}

func IsPenaltySource(value PenaltySource) bool {
	switch value {
	case PenaltySourceChannelBalance,
		PenaltySourceParticipantBond,
		PenaltySourceRoutingAdvertisementDeposit,
		PenaltySourceFraudProofDeposit:
		return true
	default:
		return false
	}
}

func IsPaymentFeeClass(value PaymentFeeClass) bool {
	switch value {
	case PaymentFeeClassChannelOpen,
		PaymentFeeClassChannelCheckpoint,
		PaymentFeeClassCooperativeClose,
		PaymentFeeClassUnilateralClose,
		PaymentFeeClassDispute,
		PaymentFeeClassFraudProofVerification,
		PaymentFeeClassConditionalPromiseSettlement,
		PaymentFeeClassVirtualChannelAnchor,
		PaymentFeeClassRoutingAdvertisement:
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
		ChainID:		req.ChainID,
		AppVersion:		CurrentAppVersion,
		ModuleName:		ModuleName,
		ChannelID:		req.ChannelID,
		ChannelType:		req.ChannelType,
		ParticipantSetHash:	ComputeParticipantSetHash(channel.Participants),
		Denom:			NativeDenom,
		Version:		CurrentStateVersion,
		Epoch:			1,
		Nonce:			1,
		Balances:		req.InitialBalances,
		TimeoutHeight:		req.OpenHeight + req.ChallengePeriod,
		ChallengePeriod:	req.ChallengePeriod,
		CloseDelay:		req.CloseDelay,
		FeePolicyID:		req.FeePolicyID,
		RequiredSignerBitmap:	ComputeRequiredSignerBitmap(channel.Participants, channel.RequiredSigners),
		SignatureScheme:	SignatureSchemeEd25519,
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

func ValidateCrossChannelPromiseTimeoutOrdering(upstreamChannel, downstreamChannel ChannelRecord, upstream, downstream ConditionalPromise, margin uint64) error {
	upstreamChannel = upstreamChannel.Normalize()
	downstreamChannel = downstreamChannel.Normalize()
	upstream = upstream.Normalize()
	downstream = downstream.Normalize()
	if margin == 0 {
		margin = DefaultTimeoutMargin
	}
	upstreamLatency := upstreamChannel.CloseDelay + upstreamChannel.DisputePeriod
	downstreamLatency := downstreamChannel.CloseDelay + downstreamChannel.DisputePeriod
	minMargin := upstreamLatency
	if downstreamLatency > minMargin {
		minMargin = downstreamLatency
	}
	if margin < minMargin {
		return errors.New("payments cross-channel timeout margin must cover dispute and settlement latency")
	}
	if upstream.ChannelID != upstreamChannel.ChannelID || downstream.ChannelID != downstreamChannel.ChannelID {
		return errors.New("payments cross-channel timeout ordering channel mismatch")
	}
	if upstream.HashLock != downstream.HashLock {
		return errors.New("payments cross-channel timeout ordering requires compatible hash locks")
	}
	if downstream.TimeoutHeight+margin < downstream.TimeoutHeight || downstream.TimeoutHeight+margin > upstream.TimeoutHeight {
		return errors.New("payments downstream timeout must expire before upstream by margin")
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

func validateLinkedPromiseConservation(promises []ConditionalPromise, amount, totalFees string) error {
	if len(promises) == 0 {
		return errors.New("payments linked promise conservation requires promises")
	}
	finalAmount, err := parsePositiveInt("payments linked promise final amount", amount)
	if err != nil {
		return err
	}
	expectedFees, err := parseNonNegativeInt("payments linked promise total fees", totalFees)
	if err != nil {
		return err
	}
	accumulatedFees := sdkmath.ZeroInt()
	for i := 1; i < len(promises); i++ {
		fee, err := parseNonNegativeInt("payments linked promise hop fee", promises[i].Fee)
		if err != nil {
			return err
		}
		accumulatedFees = accumulatedFees.Add(fee)
		incoming, err := parsePositiveInt("payments linked promise incoming amount", promises[i-1].Amount)
		if err != nil {
			return err
		}
		outgoing, err := parsePositiveInt("payments linked promise outgoing amount", promises[i].Amount)
		if err != nil {
			return err
		}
		if !incoming.Equal(outgoing.Add(fee)) {
			return errors.New("payments linked promise amount conservation failed")
		}
	}
	if !accumulatedFees.Equal(expectedFees) {
		return errors.New("payments linked promise total fee mismatch")
	}
	lastAmount, err := parsePositiveInt("payments linked promise receiver amount", promises[len(promises)-1].Amount)
	if err != nil {
		return err
	}
	if !lastAmount.Equal(finalAmount) {
		return errors.New("payments linked promise final amount mismatch")
	}
	firstAmount, err := parsePositiveInt("payments linked promise sender amount", promises[0].Amount)
	if err != nil {
		return err
	}
	if !firstAmount.Equal(finalAmount.Add(expectedFees)) {
		return errors.New("payments linked promise route total mismatch")
	}
	finalFee, err := parseNonNegativeInt("payments linked final promise fee", promises[len(promises)-1].Fee)
	if err != nil {
		return err
	}
	if len(promises) > 1 && finalFee.IsZero() && !expectedFees.IsZero() {
		return errors.New("payments linked final hop fee must pay forwarding intermediary")
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
		ChannelID:	channel.ChannelID,
		Nonce:		next.Nonce,
		ConditionRoot:	next.ConditionRoot,
		ConditionCount:	next.ConditionCount,
		Conditions:	next.Conditions,
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
		ChannelID:	next.ChannelID,
		Nonce:		next.Nonce,
		ConditionRoot:	next.ConditionRoot,
		ConditionCount:	next.ConditionCount,
		Conditions:	next.Conditions,
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
		route	PenaltyRoute
		bps	uint32
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
			Participant:	strings.TrimSpace(balance.Participant),
			Amount:		strings.TrimSpace(balance.Amount),
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

func normalizePromiseRoute(promises []ConditionalPromise) []ConditionalPromise {
	out := make([]ConditionalPromise, len(promises))
	for i, promise := range promises {
		out[i] = promise.Normalize()
	}
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

func normalizeConditionRootUpdates(updates []ConditionRootUpdate) []ConditionRootUpdate {
	out := make([]ConditionRootUpdate, len(updates))
	for i, update := range updates {
		update.ChannelID = normalizeHash(update.ChannelID)
		update.ConditionRoot = normalizeHash(update.ConditionRoot)
		update.Conditions = normalizeConditions(update.Conditions)
		out[i] = update
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ChannelID < out[j].ChannelID
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

func normalizePenaltyMatrix(matrix []PenaltyMatrixEntry) []PenaltyMatrixEntry {
	if len(matrix) == 0 {
		matrix = DefaultPenaltyMatrix()
	}
	out := make([]PenaltyMatrixEntry, len(matrix))
	for i, entry := range matrix {
		out[i] = entry.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := string(out[i].Class) + "/" + string(out[i].ProofType)
		right := string(out[j].Class) + "/" + string(out[j].ProofType)
		return left < right
	})
	return out
}

func cappedPenaltyPortion(available sdkmath.Int, capText string, bps uint32) (sdkmath.Int, error) {
	if !available.IsPositive() {
		return sdkmath.ZeroInt(), nil
	}
	portion := available
	if bps > 0 {
		if bps > MaxPenaltyRouteBps {
			return sdkmath.ZeroInt(), errors.New("payments penalty portion bps exceeds 10000")
		}
		portion = available.MulRaw(int64(bps)).QuoRaw(int64(MaxPenaltyRouteBps))
	}
	capText = strings.TrimSpace(capText)
	if capText != "" {
		capAmount, err := parseNonNegativeInt("payments penalty portion cap", capText)
		if err != nil {
			return sdkmath.ZeroInt(), err
		}
		if portion.GT(capAmount) {
			portion = capAmount
		}
	}
	if portion.GT(available) {
		portion = available
	}
	return portion, nil
}

func channelCounterparty(channel ChannelRecord, offender string) string {
	channel = channel.Normalize()
	offender = strings.TrimSpace(offender)
	for _, participant := range channel.Participants {
		if participant != offender {
			return participant
		}
	}
	return ""
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

func normalizeVirtualParentReserves(reserves []VirtualParentReserve) []VirtualParentReserve {
	out := make([]VirtualParentReserve, len(reserves))
	for i, reserve := range reserves {
		out[i] = reserve.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SegmentID != out[j].SegmentID {
			return out[i].SegmentID < out[j].SegmentID
		}
		if out[i].ParentChannelID != out[j].ParentChannelID {
			return out[i].ParentChannelID < out[j].ParentChannelID
		}
		return out[i].ReservedBy < out[j].ReservedBy
	})
	return out
}

func normalizeVirtualReserveSegments(segments []VirtualReserveSegment) []VirtualReserveSegment {
	out := make([]VirtualReserveSegment, len(segments))
	for i, segment := range segments {
		out[i] = segment.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].SegmentID < out[j].SegmentID
	})
	return out
}

func normalizeVirtualSegmentSettlementProofs(proofs []VirtualSegmentSettlementProof) []VirtualSegmentSettlementProof {
	out := make([]VirtualSegmentSettlementProof, len(proofs))
	for i, proof := range proofs {
		out[i] = proof.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].SegmentID < out[j].SegmentID
	})
	return out
}

func normalizeHashSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := normalizeHash(value)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	sortStrings(out)
	return out
}

func normalizeStoreKeySlice(values []string) []string {
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

func sortStoreV2Layout(layout *StoreV2Layout) {
	sort.SliceStable(layout.Channels, func(i, j int) bool { return layout.Channels[i].Key < layout.Channels[j].Key })
	sort.SliceStable(layout.ChannelStates, func(i, j int) bool { return layout.ChannelStates[i].Key < layout.ChannelStates[j].Key })
	sort.SliceStable(layout.PendingCloses, func(i, j int) bool { return layout.PendingCloses[i].Key < layout.PendingCloses[j].Key })
	sort.SliceStable(layout.Conditions, func(i, j int) bool { return layout.Conditions[i].Key < layout.Conditions[j].Key })
	sort.SliceStable(layout.VirtualChannels, func(i, j int) bool { return layout.VirtualChannels[i].Key < layout.VirtualChannels[j].Key })
	sort.SliceStable(layout.ParticipantChannels, func(i, j int) bool { return layout.ParticipantChannels[i].Key < layout.ParticipantChannels[j].Key })
	sort.SliceStable(layout.SettlementTombstones, func(i, j int) bool { return layout.SettlementTombstones[i].Key < layout.SettlementTombstones[j].Key })
	sort.SliceStable(layout.FeeAccumulators, func(i, j int) bool { return layout.FeeAccumulators[i].Key < layout.FeeAccumulators[j].Key })
	sort.SliceStable(layout.FraudProofs, func(i, j int) bool { return layout.FraudProofs[i].Key < layout.FraudProofs[j].Key })
}

func compactStoreV2Channel(channel ChannelRecord) ChannelRecord {
	channel = channel.Normalize()
	channel.LatestState = compactStoreV2State(channel.LatestState)
	channel.PendingClose = PendingClose{}
	return channel
}

func compactStoreV2State(state ChannelState) ChannelState {
	state = state.Normalize()
	return ChannelState{
		ChainID:	state.ChainID,
		ChannelID:	state.ChannelID,
		ChannelType:	state.ChannelType,
		Denom:		state.Denom,
		Version:	state.Version,
		Nonce:		state.Nonce,
		Epoch:		state.Epoch,
		StateHash:	state.StateHash,
	}
}

func StoreV2RoutingKeyForChannel(channel ChannelRecord) string {
	channel = channel.Normalize()
	if !channel.RoutingAdvertised {
		return ""
	}
	return PaymentRoutingAdvertisementIndexKey(channel.ChannelID)
}

func storeV2ConditionFromPayment(channel ChannelRecord, condition ConditionalPayment, settled bool) StoreV2ConditionRecord {
	channel = channel.Normalize()
	condition = condition.Normalize()
	return StoreV2ConditionRecord{
		Key:		StoreV2ConditionKey(condition.ConditionID),
		Version:	StoreV2MigrationVersion,
		ConditionID:	condition.ConditionID,
		ChannelID:	channel.ChannelID,
		ExpiresHeight:	condition.TimeoutHeight,
		Settled:	settled,
	}.Normalize()
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
