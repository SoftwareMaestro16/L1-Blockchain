package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultMinimumChannelCollateral			= "1"
	DefaultMaximumChannelCollateral			= "1000000000000000000"
	DefaultMaximumPromiseAmountRatioBps		= uint32(10_000)
	DefaultMaximumPromiseLifetime			= uint64(10_000)
	DefaultExpiredPromiseCleanupLimitPerBlock	= uint64(256)
	DefaultMaxVirtualChannelsPerParentChannel	= uint64(128)
	DefaultMaximumVirtualChannelDepth		= uint64(MaxRoutingHops)
	DefaultMinimumParentTimeoutMargin		= uint64(DefaultTimeoutMargin)
	DefaultVirtualChannelReservationExpiry		= uint64(256)
	DefaultMultiSegmentVirtualMaxSegments		= uint64(MaxParentChannels)
	DefaultStaleClosePenalty			= "10"
	DefaultSameNonceDoubleSignPenalty		= "20"
	DefaultInvalidConditionPenalty			= "8"
	DefaultReplayAttemptPenalty			= "8"
	DefaultInvalidFraudProofDeposit			= "1"
	DefaultReporterRewardPercentageBps		= uint32(1_000)
	DefaultReporterRewardCap			= "5"
	DefaultPenaltyBurnAllocationBps			= uint32(2_500)
	DefaultSecurityReserveAllocationBps		= uint32(2_500)
	DefaultMaxTopologyUpdatesPerPeerWindow		= uint32(16)
	DefaultRouteFailureScoreDecay			= uint64(DefaultGossipTTL)
	DefaultCongestionPenaltyDecay			= uint64(DefaultGossipTTL)
	DefaultCapacityProbeRateLimit			= uint32(16)
	DefaultFinalizationQueueWorkLimit		= uint64(256)
	DefaultChannelOpenCongestionMultiplierBps	= uint32(20_000)
	DefaultDisputeCongestionMultiplierBps		= uint32(25_000)
	DefaultStorePruningHorizon			= uint64(DefaultReplayHorizon)
)

type PaymentGovernanceParams struct {
	Channel		PaymentChannelGovernanceParams
	Conditional	PaymentConditionalGovernanceParams
	Virtual		PaymentVirtualChannelGovernanceParams
	FraudPenalty	PaymentFraudPenaltyGovernanceParams
	Routing		PaymentRoutingGovernanceParams
	Execution	PaymentExecutionGovernanceParams
	ParamsHash	string
}

type PaymentChannelGovernanceParams struct {
	MinimumChannelCollateral	string
	MaximumChannelCollateral	string
	MinimumChallengePeriod		uint64
	MaximumChallengePeriod		uint64
	DefaultChallengePeriod		uint64
	MinimumCloseDelay		uint64
	MaximumCloseDelay		uint64
	ChannelOpenBaseFee		string
	ChannelStorageFeePerByte	string
	ChannelTombstoneRetention	uint64
}

type PaymentConditionalGovernanceParams struct {
	MaximumActivePromisesPerChannel		uint64
	MaximumPromiseAmountRatioBps		uint32
	MinimumTimeoutMargin			uint64
	MaximumPromiseLifetime			uint64
	BatchResolutionMaximumSize		uint64
	PromiseStorageFee			string
	ExpiredPromiseCleanupLimitPerBlock	uint64
}

type PaymentVirtualChannelGovernanceParams struct {
	MaximumVirtualChannelsPerParentChannel	uint64
	MaximumVirtualChannelDepth		uint64
	MinimumParentTimeoutMargin		uint64
	VirtualChannelAnchorFee			string
	VirtualChannelReservationExpiry		uint64
	MultiSegmentVirtualChannelMaxSegments	uint64
}

type PaymentFraudPenaltyGovernanceParams struct {
	StaleClosePenalty			string
	SameNonceDoubleSignPenalty		string
	InvalidConditionPenalty			string
	ReplayAttemptPenalty			string
	InvalidFraudProofDeposit		string
	ReporterRewardPercentageBps		uint32
	ReporterRewardCap			string
	PenaltyBurnAllocationBps		uint32
	SecurityReserveAllocationBps		uint32
	CounterpartyCompensationPriority	bool
}

type PaymentRoutingGovernanceParams struct {
	RoutingAdvertisementDeposit		string
	GossipMessageExpiry			uint64
	LiquidityHintExpiry			uint64
	MaximumTopologyUpdatesPerPeerWindow	uint32
	RouteFailureScoreDecay			uint64
	CongestionPenaltyDecay			uint64
	CapacityProbeRateLimit			uint32
	CapacityProbeWindow			uint64
}

type PaymentExecutionGovernanceParams struct {
	SettlementBatchMaximumSize		uint64
	FinalizationQueueWorkLimitPerBlock	uint64
	ExpiredPromiseCleanupWorkLimitPerBlock	uint64
	ChannelOpenCongestionFeeMultiplierBps	uint32
	DisputeCongestionFeeMultiplierBps	uint32
	StorePruningHorizon			uint64
}

func DefaultPaymentGovernanceParams() PaymentGovernanceParams {
	feeSchedule := DefaultPaymentFeeSchedule().Normalize()
	params := PaymentGovernanceParams{
		Channel: PaymentChannelGovernanceParams{
			MinimumChannelCollateral:	DefaultMinimumChannelCollateral,
			MaximumChannelCollateral:	DefaultMaximumChannelCollateral,
			MinimumChallengePeriod:		MinChallengePeriod,
			MaximumChallengePeriod:		MaxChallengePeriod,
			DefaultChallengePeriod:		DefaultDisputePeriod,
			MinimumCloseDelay:		MinCloseDelay,
			MaximumCloseDelay:		MaxCloseDelay,
			ChannelOpenBaseFee:		feeSchedule.ChannelOpenFee,
			ChannelStorageFeePerByte:	feeSchedule.StorageByteFee,
			ChannelTombstoneRetention:	DefaultReplayHorizon,
		},
		Conditional: PaymentConditionalGovernanceParams{
			MaximumActivePromisesPerChannel:	MaxConditionsPerState,
			MaximumPromiseAmountRatioBps:		DefaultMaximumPromiseAmountRatioBps,
			MinimumTimeoutMargin:			DefaultTimeoutMargin,
			MaximumPromiseLifetime:			DefaultMaximumPromiseLifetime,
			BatchResolutionMaximumSize:		MaxSettlementBatchOps,
			PromiseStorageFee:			feeSchedule.ConditionalPromiseSettlementFee,
			ExpiredPromiseCleanupLimitPerBlock:	DefaultExpiredPromiseCleanupLimitPerBlock,
		},
		Virtual: PaymentVirtualChannelGovernanceParams{
			MaximumVirtualChannelsPerParentChannel:	DefaultMaxVirtualChannelsPerParentChannel,
			MaximumVirtualChannelDepth:		DefaultMaximumVirtualChannelDepth,
			MinimumParentTimeoutMargin:		DefaultMinimumParentTimeoutMargin,
			VirtualChannelAnchorFee:		feeSchedule.VirtualChannelAnchorFee,
			VirtualChannelReservationExpiry:	DefaultVirtualChannelReservationExpiry,
			MultiSegmentVirtualChannelMaxSegments:	DefaultMultiSegmentVirtualMaxSegments,
		},
		FraudPenalty: PaymentFraudPenaltyGovernanceParams{
			StaleClosePenalty:			DefaultStaleClosePenalty,
			SameNonceDoubleSignPenalty:		DefaultSameNonceDoubleSignPenalty,
			InvalidConditionPenalty:		DefaultInvalidConditionPenalty,
			ReplayAttemptPenalty:			DefaultReplayAttemptPenalty,
			InvalidFraudProofDeposit:		DefaultInvalidFraudProofDeposit,
			ReporterRewardPercentageBps:		DefaultReporterRewardPercentageBps,
			ReporterRewardCap:			DefaultReporterRewardCap,
			PenaltyBurnAllocationBps:		DefaultPenaltyBurnAllocationBps,
			SecurityReserveAllocationBps:		DefaultSecurityReserveAllocationBps,
			CounterpartyCompensationPriority:	true,
		},
		Routing: PaymentRoutingGovernanceParams{
			RoutingAdvertisementDeposit:		feeSchedule.RoutingAdvertisementDeposit,
			GossipMessageExpiry:			DefaultGossipTTL,
			LiquidityHintExpiry:			DefaultGossipTTL,
			MaximumTopologyUpdatesPerPeerWindow:	DefaultMaxTopologyUpdatesPerPeerWindow,
			RouteFailureScoreDecay:			DefaultRouteFailureScoreDecay,
			CongestionPenaltyDecay:			DefaultCongestionPenaltyDecay,
			CapacityProbeRateLimit:			DefaultCapacityProbeRateLimit,
			CapacityProbeWindow:			DefaultGossipTTL,
		},
		Execution: PaymentExecutionGovernanceParams{
			SettlementBatchMaximumSize:		MaxSettlementBatchOps,
			FinalizationQueueWorkLimitPerBlock:	DefaultFinalizationQueueWorkLimit,
			ExpiredPromiseCleanupWorkLimitPerBlock:	DefaultExpiredPromiseCleanupLimitPerBlock,
			ChannelOpenCongestionFeeMultiplierBps:	DefaultChannelOpenCongestionMultiplierBps,
			DisputeCongestionFeeMultiplierBps:	DefaultDisputeCongestionMultiplierBps,
			StorePruningHorizon:			DefaultStorePruningHorizon,
		},
	}
	params = params.Normalize()
	params.ParamsHash = ComputePaymentGovernanceParamsHash(params)
	return params
}

func ComputePaymentGovernanceParamsHash(params PaymentGovernanceParams) string {
	params = params.Normalize()
	return HashParts(
		"payments-governance-params-v1",
		params.Channel.MinimumChannelCollateral,
		params.Channel.MaximumChannelCollateral,
		fmt.Sprintf("%020d", params.Channel.MinimumChallengePeriod),
		fmt.Sprintf("%020d", params.Channel.MaximumChallengePeriod),
		fmt.Sprintf("%020d", params.Channel.DefaultChallengePeriod),
		fmt.Sprintf("%020d", params.Channel.MinimumCloseDelay),
		fmt.Sprintf("%020d", params.Channel.MaximumCloseDelay),
		params.Channel.ChannelOpenBaseFee,
		params.Channel.ChannelStorageFeePerByte,
		fmt.Sprintf("%020d", params.Channel.ChannelTombstoneRetention),
		fmt.Sprintf("%020d", params.Conditional.MaximumActivePromisesPerChannel),
		fmt.Sprintf("%010d", params.Conditional.MaximumPromiseAmountRatioBps),
		fmt.Sprintf("%020d", params.Conditional.MinimumTimeoutMargin),
		fmt.Sprintf("%020d", params.Conditional.MaximumPromiseLifetime),
		fmt.Sprintf("%020d", params.Conditional.BatchResolutionMaximumSize),
		params.Conditional.PromiseStorageFee,
		fmt.Sprintf("%020d", params.Conditional.ExpiredPromiseCleanupLimitPerBlock),
		fmt.Sprintf("%020d", params.Virtual.MaximumVirtualChannelsPerParentChannel),
		fmt.Sprintf("%020d", params.Virtual.MaximumVirtualChannelDepth),
		fmt.Sprintf("%020d", params.Virtual.MinimumParentTimeoutMargin),
		params.Virtual.VirtualChannelAnchorFee,
		fmt.Sprintf("%020d", params.Virtual.VirtualChannelReservationExpiry),
		fmt.Sprintf("%020d", params.Virtual.MultiSegmentVirtualChannelMaxSegments),
		params.FraudPenalty.StaleClosePenalty,
		params.FraudPenalty.SameNonceDoubleSignPenalty,
		params.FraudPenalty.InvalidConditionPenalty,
		params.FraudPenalty.ReplayAttemptPenalty,
		params.FraudPenalty.InvalidFraudProofDeposit,
		fmt.Sprintf("%010d", params.FraudPenalty.ReporterRewardPercentageBps),
		params.FraudPenalty.ReporterRewardCap,
		fmt.Sprintf("%010d", params.FraudPenalty.PenaltyBurnAllocationBps),
		fmt.Sprintf("%010d", params.FraudPenalty.SecurityReserveAllocationBps),
		fmt.Sprintf("%t", params.FraudPenalty.CounterpartyCompensationPriority),
		params.Routing.RoutingAdvertisementDeposit,
		fmt.Sprintf("%020d", params.Routing.GossipMessageExpiry),
		fmt.Sprintf("%020d", params.Routing.LiquidityHintExpiry),
		fmt.Sprintf("%010d", params.Routing.MaximumTopologyUpdatesPerPeerWindow),
		fmt.Sprintf("%020d", params.Routing.RouteFailureScoreDecay),
		fmt.Sprintf("%020d", params.Routing.CongestionPenaltyDecay),
		fmt.Sprintf("%010d", params.Routing.CapacityProbeRateLimit),
		fmt.Sprintf("%020d", params.Routing.CapacityProbeWindow),
		fmt.Sprintf("%020d", params.Execution.SettlementBatchMaximumSize),
		fmt.Sprintf("%020d", params.Execution.FinalizationQueueWorkLimitPerBlock),
		fmt.Sprintf("%020d", params.Execution.ExpiredPromiseCleanupWorkLimitPerBlock),
		fmt.Sprintf("%010d", params.Execution.ChannelOpenCongestionFeeMultiplierBps),
		fmt.Sprintf("%010d", params.Execution.DisputeCongestionFeeMultiplierBps),
		fmt.Sprintf("%020d", params.Execution.StorePruningHorizon),
	)
}

func BuildGovernedPaymentFeeSchedule(params PaymentGovernanceParams) (PaymentFeeSchedule, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return PaymentFeeSchedule{}, err
	}
	schedule := DefaultPaymentFeeSchedule().Normalize()
	schedule.ChannelOpenFee = params.Channel.ChannelOpenBaseFee
	schedule.OpenFeeMin = params.Channel.ChannelOpenBaseFee
	schedule.StorageByteFee = params.Channel.ChannelStorageFeePerByte
	schedule.ConditionalPromiseSettlementFee = params.Conditional.PromiseStorageFee
	schedule.VirtualChannelAnchorFee = params.Virtual.VirtualChannelAnchorFee
	schedule.RoutingAdvertisementDeposit = params.Routing.RoutingAdvertisementDeposit
	return schedule.Normalize(), schedule.Validate()
}

func ValidateChannelOpenRequestWithGovernance(req ChannelOpenRequest, params PaymentGovernanceParams) error {
	req = req.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	if err := params.Channel.ValidateOpenCollateral(req.Collateral); err != nil {
		return err
	}
	if req.ChallengePeriod < params.Channel.MinimumChallengePeriod || req.ChallengePeriod > params.Channel.MaximumChallengePeriod {
		return fmt.Errorf("payments open challenge period must be between governance bounds %d and %d", params.Channel.MinimumChallengePeriod, params.Channel.MaximumChallengePeriod)
	}
	if req.CloseDelay < params.Channel.MinimumCloseDelay || req.CloseDelay > params.Channel.MaximumCloseDelay {
		return fmt.Errorf("payments open close delay must be between governance bounds %d and %d", params.Channel.MinimumCloseDelay, params.Channel.MaximumCloseDelay)
	}
	paid, err := parseNonNegativeInt("payments opening fee paid", req.OpeningFeePaid)
	if err != nil {
		return err
	}
	required, err := parseNonNegativeInt("payments governance channel open base fee", params.Channel.ChannelOpenBaseFee)
	if err != nil {
		return err
	}
	if paid.LT(required) {
		return errors.New("payments opening fee paid is below governance base fee")
	}
	return nil
}

func ValidateConditionalPromisesForChannelWithGovernance(channel ChannelRecord, promises []ConditionalPromise, settledClaims []ConditionClaimRecord, params PaymentGovernanceParams) error {
	channel = channel.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if uint64(len(promises)) > params.Conditional.MaximumActivePromisesPerChannel {
		return fmt.Errorf("payments active promises exceed governance maximum %d", params.Conditional.MaximumActivePromisesPerChannel)
	}
	if err := ValidateConditionalPromisesForChannel(channel, promises, settledClaims); err != nil {
		return err
	}
	collateral, err := parsePositiveInt("payments governance channel collateral", channel.Collateral)
	if err != nil {
		return err
	}
	maxPromiseAmount := collateral.MulRaw(int64(params.Conditional.MaximumPromiseAmountRatioBps)).QuoRaw(10_000)
	for _, promise := range normalizeConditionalPromises(promises) {
		if err := params.Conditional.ValidatePromiseWindow(channel, promise); err != nil {
			return err
		}
		amount, err := parsePositiveInt("payments governance promise amount", promise.Amount)
		if err != nil {
			return err
		}
		fee, err := parseNonNegativeInt("payments governance promise fee", promise.Fee)
		if err != nil {
			return err
		}
		if amount.Add(fee).GT(maxPromiseAmount) {
			return errors.New("payments promise amount exceeds governance channel ratio")
		}
	}
	return nil
}

func ValidateBatchConditionSettlementWithGovernance(req BatchConditionSettlementRequest, state PaymentsState, settledClaims []ConditionClaimRecord, params PaymentGovernanceParams) error {
	req = req.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if uint64(len(req.LinkageProof.Promises)) > params.Conditional.BatchResolutionMaximumSize {
		return fmt.Errorf("payments batch promise resolution exceeds governance maximum %d", params.Conditional.BatchResolutionMaximumSize)
	}
	if err := req.ValidateForState(state, settledClaims); err != nil {
		return err
	}
	return nil
}

func SettlementTombstoneExpiryHeight(closedHeight uint64, params PaymentGovernanceParams) (uint64, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if closedHeight == 0 {
		return 0, errors.New("payments tombstone closed height must be positive")
	}
	if closedHeight+params.Channel.ChannelTombstoneRetention < closedHeight {
		return 0, errors.New("payments tombstone retention overflows height")
	}
	return closedHeight + params.Channel.ChannelTombstoneRetention, nil
}

func ExpiredPromiseCleanupLimit(params PaymentGovernanceParams) (uint64, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return 0, err
	}
	return params.Conditional.ExpiredPromiseCleanupLimitPerBlock, nil
}

func (p PaymentGovernanceParams) Normalize() PaymentGovernanceParams {
	defaults := DefaultPaymentGovernanceParamsNoHash()
	p.Channel = p.Channel.Normalize(defaults.Channel)
	p.Conditional = p.Conditional.Normalize(defaults.Conditional)
	p.Virtual = p.Virtual.Normalize(defaults.Virtual)
	p.FraudPenalty = p.FraudPenalty.Normalize(defaults.FraudPenalty)
	p.Routing = p.Routing.Normalize(defaults.Routing)
	p.Execution = p.Execution.Normalize(defaults.Execution)
	p.ParamsHash = normalizeOptionalHash(p.ParamsHash)
	return p
}

func (p PaymentGovernanceParams) WithHash() PaymentGovernanceParams {
	p = p.Normalize()
	p.ParamsHash = ComputePaymentGovernanceParamsHash(p)
	return p.Normalize()
}

func (p PaymentGovernanceParams) Validate() error {
	params := p.Normalize()
	if err := params.Channel.Validate(); err != nil {
		return err
	}
	if err := params.Conditional.Validate(); err != nil {
		return err
	}
	if err := params.Virtual.Validate(); err != nil {
		return err
	}
	if err := params.FraudPenalty.Validate(); err != nil {
		return err
	}
	if err := params.Routing.Validate(); err != nil {
		return err
	}
	if err := params.Execution.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("payments governance params hash", params.ParamsHash); err != nil {
		return err
	}
	if expected := ComputePaymentGovernanceParamsHash(params); params.ParamsHash != expected {
		return errors.New("payments governance params hash mismatch")
	}
	return nil
}

func DefaultPaymentGovernanceParamsNoHash() PaymentGovernanceParams {
	feeSchedule := DefaultPaymentFeeSchedule().Normalize()
	return PaymentGovernanceParams{
		Channel: PaymentChannelGovernanceParams{
			MinimumChannelCollateral:	DefaultMinimumChannelCollateral,
			MaximumChannelCollateral:	DefaultMaximumChannelCollateral,
			MinimumChallengePeriod:		MinChallengePeriod,
			MaximumChallengePeriod:		MaxChallengePeriod,
			DefaultChallengePeriod:		DefaultDisputePeriod,
			MinimumCloseDelay:		MinCloseDelay,
			MaximumCloseDelay:		MaxCloseDelay,
			ChannelOpenBaseFee:		feeSchedule.ChannelOpenFee,
			ChannelStorageFeePerByte:	feeSchedule.StorageByteFee,
			ChannelTombstoneRetention:	DefaultReplayHorizon,
		},
		Conditional: PaymentConditionalGovernanceParams{
			MaximumActivePromisesPerChannel:	MaxConditionsPerState,
			MaximumPromiseAmountRatioBps:		DefaultMaximumPromiseAmountRatioBps,
			MinimumTimeoutMargin:			DefaultTimeoutMargin,
			MaximumPromiseLifetime:			DefaultMaximumPromiseLifetime,
			BatchResolutionMaximumSize:		MaxSettlementBatchOps,
			PromiseStorageFee:			feeSchedule.ConditionalPromiseSettlementFee,
			ExpiredPromiseCleanupLimitPerBlock:	DefaultExpiredPromiseCleanupLimitPerBlock,
		},
		Virtual: PaymentVirtualChannelGovernanceParams{
			MaximumVirtualChannelsPerParentChannel:	DefaultMaxVirtualChannelsPerParentChannel,
			MaximumVirtualChannelDepth:		DefaultMaximumVirtualChannelDepth,
			MinimumParentTimeoutMargin:		DefaultMinimumParentTimeoutMargin,
			VirtualChannelAnchorFee:		feeSchedule.VirtualChannelAnchorFee,
			VirtualChannelReservationExpiry:	DefaultVirtualChannelReservationExpiry,
			MultiSegmentVirtualChannelMaxSegments:	DefaultMultiSegmentVirtualMaxSegments,
		},
		FraudPenalty: PaymentFraudPenaltyGovernanceParams{
			StaleClosePenalty:			DefaultStaleClosePenalty,
			SameNonceDoubleSignPenalty:		DefaultSameNonceDoubleSignPenalty,
			InvalidConditionPenalty:		DefaultInvalidConditionPenalty,
			ReplayAttemptPenalty:			DefaultReplayAttemptPenalty,
			InvalidFraudProofDeposit:		DefaultInvalidFraudProofDeposit,
			ReporterRewardPercentageBps:		DefaultReporterRewardPercentageBps,
			ReporterRewardCap:			DefaultReporterRewardCap,
			PenaltyBurnAllocationBps:		DefaultPenaltyBurnAllocationBps,
			SecurityReserveAllocationBps:		DefaultSecurityReserveAllocationBps,
			CounterpartyCompensationPriority:	true,
		},
		Routing: PaymentRoutingGovernanceParams{
			RoutingAdvertisementDeposit:		feeSchedule.RoutingAdvertisementDeposit,
			GossipMessageExpiry:			DefaultGossipTTL,
			LiquidityHintExpiry:			DefaultGossipTTL,
			MaximumTopologyUpdatesPerPeerWindow:	DefaultMaxTopologyUpdatesPerPeerWindow,
			RouteFailureScoreDecay:			DefaultRouteFailureScoreDecay,
			CongestionPenaltyDecay:			DefaultCongestionPenaltyDecay,
			CapacityProbeRateLimit:			DefaultCapacityProbeRateLimit,
			CapacityProbeWindow:			DefaultGossipTTL,
		},
		Execution: PaymentExecutionGovernanceParams{
			SettlementBatchMaximumSize:		MaxSettlementBatchOps,
			FinalizationQueueWorkLimitPerBlock:	DefaultFinalizationQueueWorkLimit,
			ExpiredPromiseCleanupWorkLimitPerBlock:	DefaultExpiredPromiseCleanupLimitPerBlock,
			ChannelOpenCongestionFeeMultiplierBps:	DefaultChannelOpenCongestionMultiplierBps,
			DisputeCongestionFeeMultiplierBps:	DefaultDisputeCongestionMultiplierBps,
			StorePruningHorizon:			DefaultStorePruningHorizon,
		},
	}
}

func (p PaymentChannelGovernanceParams) Normalize(defaults PaymentChannelGovernanceParams) PaymentChannelGovernanceParams {
	p.MinimumChannelCollateral = normalizeAmountOrDefault(p.MinimumChannelCollateral, defaults.MinimumChannelCollateral)
	p.MaximumChannelCollateral = normalizeAmountOrDefault(p.MaximumChannelCollateral, defaults.MaximumChannelCollateral)
	if p.MinimumChallengePeriod == 0 {
		p.MinimumChallengePeriod = defaults.MinimumChallengePeriod
	}
	if p.MaximumChallengePeriod == 0 {
		p.MaximumChallengePeriod = defaults.MaximumChallengePeriod
	}
	if p.DefaultChallengePeriod == 0 {
		p.DefaultChallengePeriod = defaults.DefaultChallengePeriod
	}
	if p.MinimumCloseDelay == 0 {
		p.MinimumCloseDelay = defaults.MinimumCloseDelay
	}
	if p.MaximumCloseDelay == 0 {
		p.MaximumCloseDelay = defaults.MaximumCloseDelay
	}
	p.ChannelOpenBaseFee = normalizeAmountOrDefault(p.ChannelOpenBaseFee, defaults.ChannelOpenBaseFee)
	p.ChannelStorageFeePerByte = normalizeAmountOrDefault(p.ChannelStorageFeePerByte, defaults.ChannelStorageFeePerByte)
	if p.ChannelTombstoneRetention == 0 {
		p.ChannelTombstoneRetention = defaults.ChannelTombstoneRetention
	}
	return p
}

func (p PaymentChannelGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Channel)
	minCollateral, err := parsePositiveInt("payments governance minimum channel collateral", params.MinimumChannelCollateral)
	if err != nil {
		return err
	}
	maxCollateral, err := parsePositiveInt("payments governance maximum channel collateral", params.MaximumChannelCollateral)
	if err != nil {
		return err
	}
	if maxCollateral.LT(minCollateral) {
		return errors.New("payments governance maximum channel collateral must be >= minimum")
	}
	if params.MinimumChallengePeriod < MinChallengePeriod || params.MaximumChallengePeriod > MaxChallengePeriod || params.MinimumChallengePeriod > params.MaximumChallengePeriod {
		return fmt.Errorf("payments governance challenge period bounds must be within %d and %d", MinChallengePeriod, MaxChallengePeriod)
	}
	if params.DefaultChallengePeriod < params.MinimumChallengePeriod || params.DefaultChallengePeriod > params.MaximumChallengePeriod {
		return errors.New("payments governance default challenge period must fit bounds")
	}
	if params.MinimumCloseDelay < MinCloseDelay || params.MaximumCloseDelay > MaxCloseDelay || params.MinimumCloseDelay > params.MaximumCloseDelay {
		return fmt.Errorf("payments governance close delay bounds must be within %d and %d", MinCloseDelay, MaxCloseDelay)
	}
	if err := validatePositiveInt("payments governance channel open base fee", params.ChannelOpenBaseFee); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments governance channel storage fee per byte", params.ChannelStorageFeePerByte); err != nil {
		return err
	}
	if params.ChannelTombstoneRetention == 0 {
		return errors.New("payments governance tombstone retention must be positive")
	}
	return nil
}

func (p PaymentChannelGovernanceParams) ValidateOpenCollateral(collateralText string) error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Channel)
	collateral, err := parsePositiveInt("payments governance open collateral", collateralText)
	if err != nil {
		return err
	}
	minCollateral, err := parsePositiveInt("payments governance minimum channel collateral", params.MinimumChannelCollateral)
	if err != nil {
		return err
	}
	maxCollateral, err := parsePositiveInt("payments governance maximum channel collateral", params.MaximumChannelCollateral)
	if err != nil {
		return err
	}
	if collateral.LT(minCollateral) {
		return errors.New("payments open collateral below governance minimum")
	}
	if collateral.GT(maxCollateral) {
		return errors.New("payments open collateral above governance maximum")
	}
	return nil
}

func (p PaymentConditionalGovernanceParams) Normalize(defaults PaymentConditionalGovernanceParams) PaymentConditionalGovernanceParams {
	if p.MaximumActivePromisesPerChannel == 0 {
		p.MaximumActivePromisesPerChannel = defaults.MaximumActivePromisesPerChannel
	}
	if p.MaximumPromiseAmountRatioBps == 0 {
		p.MaximumPromiseAmountRatioBps = defaults.MaximumPromiseAmountRatioBps
	}
	if p.MinimumTimeoutMargin == 0 {
		p.MinimumTimeoutMargin = defaults.MinimumTimeoutMargin
	}
	if p.MaximumPromiseLifetime == 0 {
		p.MaximumPromiseLifetime = defaults.MaximumPromiseLifetime
	}
	if p.BatchResolutionMaximumSize == 0 {
		p.BatchResolutionMaximumSize = defaults.BatchResolutionMaximumSize
	}
	p.PromiseStorageFee = normalizeAmountOrDefault(p.PromiseStorageFee, defaults.PromiseStorageFee)
	if p.ExpiredPromiseCleanupLimitPerBlock == 0 {
		p.ExpiredPromiseCleanupLimitPerBlock = defaults.ExpiredPromiseCleanupLimitPerBlock
	}
	return p
}

func (p PaymentConditionalGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Conditional)
	if params.MaximumActivePromisesPerChannel == 0 || params.MaximumActivePromisesPerChannel > MaxConditionsPerState {
		return fmt.Errorf("payments governance active promises per channel must be between 1 and %d", MaxConditionsPerState)
	}
	if params.MaximumPromiseAmountRatioBps == 0 || params.MaximumPromiseAmountRatioBps > 10_000 {
		return errors.New("payments governance promise amount ratio must be between 1 and 10000 bps")
	}
	if params.MinimumTimeoutMargin == 0 {
		return errors.New("payments governance minimum timeout margin must be positive")
	}
	if params.MaximumPromiseLifetime < params.MinimumTimeoutMargin {
		return errors.New("payments governance maximum promise lifetime must cover timeout margin")
	}
	if params.BatchResolutionMaximumSize == 0 || params.BatchResolutionMaximumSize > MaxSettlementBatchOps {
		return fmt.Errorf("payments governance batch resolution maximum size must be between 1 and %d", MaxSettlementBatchOps)
	}
	if err := validateNonNegativeInt("payments governance promise storage fee", params.PromiseStorageFee); err != nil {
		return err
	}
	if params.ExpiredPromiseCleanupLimitPerBlock == 0 {
		return errors.New("payments governance expired promise cleanup limit must be positive")
	}
	return nil
}

func (p PaymentConditionalGovernanceParams) ValidatePromiseWindow(channel ChannelRecord, promise ConditionalPromise) error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Conditional)
	channel = channel.Normalize()
	promise = promise.Normalize()
	if promise.TimeoutHeight <= channel.OpenHeight {
		return errors.New("payments governance promise timeout must be after channel open")
	}
	lifetime := promise.TimeoutHeight - channel.OpenHeight
	if lifetime > params.MaximumPromiseLifetime {
		return errors.New("payments promise lifetime exceeds governance maximum")
	}
	maxHeight := channel.LatestState.Normalize().TimeoutHeight
	if maxHeight == 0 {
		maxHeight = channel.OpenHeight + channel.CloseDelay + channel.DisputePeriod
	}
	if promise.TimeoutHeight+params.MinimumTimeoutMargin < promise.TimeoutHeight || promise.TimeoutHeight+params.MinimumTimeoutMargin > maxHeight {
		return errors.New("payments promise timeout does not leave governance timeout margin")
	}
	return nil
}

func ValidateVirtualChannelWithGovernance(state PaymentsState, vc VirtualChannel, params PaymentGovernanceParams) error {
	state = state.Export()
	params = params.Normalize()
	vc = vc.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if err := state.Validate(); err != nil {
		return err
	}
	if err := vc.ValidateCore(); err != nil {
		return err
	}
	if uint64(len(vc.ParentChannelIDs)) > params.Virtual.MultiSegmentVirtualChannelMaxSegments {
		return fmt.Errorf("payments virtual parent segments exceed governance maximum %d", params.Virtual.MultiSegmentVirtualChannelMaxSegments)
	}
	if VirtualChannelDepth(vc) > params.Virtual.MaximumVirtualChannelDepth {
		return fmt.Errorf("payments virtual channel depth exceeds governance maximum %d", params.Virtual.MaximumVirtualChannelDepth)
	}
	anchorFeePaid, err := parseNonNegativeInt("payments virtual anchor fee paid", vc.AnchorFeePaid)
	if err != nil {
		return err
	}
	requiredAnchorFee, err := parseNonNegativeInt("payments governance virtual anchor fee", params.Virtual.VirtualChannelAnchorFee)
	if err != nil {
		return err
	}
	if anchorFeePaid.LT(requiredAnchorFee) {
		return errors.New("payments virtual channel anchor fee below governance minimum")
	}
	parentCounts := activeVirtualChannelsByParent(state)
	for _, parentID := range vc.ParentChannelIDs {
		if parentCounts[parentID] >= params.Virtual.MaximumVirtualChannelsPerParentChannel {
			return fmt.Errorf("payments parent channel %s exceeds governance virtual channel limit", parentID)
		}
		parent, found := state.ChannelByID(parentID)
		if !found {
			return errors.New("payments virtual parent channel not found")
		}
		if err := params.Virtual.ValidateParentTimeoutMargin(parent, vc); err != nil {
			return err
		}
	}
	return nil
}

func ValidateVirtualActivationProofWithGovernance(state PaymentsState, proof VirtualActivationProof, params PaymentGovernanceParams) error {
	proof = proof.Normalize()
	if err := ValidateVirtualChannelWithGovernance(state, proof.VirtualChannel, params); err != nil {
		return err
	}
	if uint64(len(proof.ParentReserves)) > params.Normalize().Virtual.MultiSegmentVirtualChannelMaxSegments {
		return fmt.Errorf("payments virtual reserves exceed governance maximum %d", params.Normalize().Virtual.MultiSegmentVirtualChannelMaxSegments)
	}
	for _, reserve := range proof.ParentReserves {
		reserve = reserve.Normalize()
		if reserve.Signature.ExpirationHeight == 0 {
			return errors.New("payments virtual reservation expiration height is required")
		}
		if reserve.Signature.ExpirationHeight > proof.VirtualChannel.ExpiresHeight {
			return errors.New("payments virtual reservation expiry must not exceed virtual channel expiry")
		}
		if proof.RouteTimeoutHeight > 0 && reserve.Signature.ExpirationHeight+params.Normalize().Virtual.MinimumParentTimeoutMargin > proof.RouteTimeoutHeight {
			return errors.New("payments virtual reservation expiry does not leave parent timeout margin")
		}
	}
	return nil
}

func VirtualChannelReservationExpiryHeight(currentHeight uint64, params PaymentGovernanceParams) (uint64, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if currentHeight == 0 {
		return 0, errors.New("payments virtual reservation current height must be positive")
	}
	if currentHeight+params.Virtual.VirtualChannelReservationExpiry < currentHeight {
		return 0, errors.New("payments virtual reservation expiry overflows height")
	}
	return currentHeight + params.Virtual.VirtualChannelReservationExpiry, nil
}

func VirtualChannelDepth(vc VirtualChannel) uint64 {
	vc = vc.Normalize()
	depth := uint64(len(vc.Intermediaries) + 1)
	if segmentDepth := uint64(len(vc.ParentChannelIDs)); segmentDepth > depth {
		depth = segmentDepth
	}
	return depth
}

func BuildGovernedPenaltyMatrix(params PaymentGovernanceParams) ([]PenaltyMatrixEntry, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return nil, err
	}
	fraud := params.FraudPenalty
	matrix := []PenaltyMatrixEntry{
		fraud.matrixEntry(PenaltyClassInvalidClose, FraudProofTypeInvalidClose, PenaltySourceChannelBalance, fraud.StaleClosePenalty),
		fraud.matrixEntry(PenaltyClassInvalidClose, FraudProofTypeInvalidBalance, PenaltySourceChannelBalance, fraud.StaleClosePenalty),
		fraud.matrixEntry(PenaltyClassStaleClose, FraudProofTypeStaleClose, PenaltySourceChannelBalance, fraud.StaleClosePenalty),
		fraud.matrixEntry(PenaltyClassDoubleSign, FraudProofTypeDoubleSign, PenaltySourceParticipantBond, fraud.SameNonceDoubleSignPenalty),
		fraud.matrixEntry(PenaltyClassInvalidCondition, FraudProofTypeInvalidCondition, PenaltySourceChannelBalance, fraud.InvalidConditionPenalty),
		fraud.matrixEntry(PenaltyClassReplayAttempt, FraudProofTypeReplayAttempt, PenaltySourceChannelBalance, fraud.ReplayAttemptPenalty),
		fraud.matrixEntry(PenaltyClassAsyncOverexposure, FraudProofTypeAsyncOverexposure, PenaltySourceParticipantBond, fraud.SameNonceDoubleSignPenalty),
		{
			Class:				PenaltyClassInvalidFraudProof,
			Source:				PenaltySourceFraudProofDeposit,
			BasePenalty:			fraud.InvalidFraudProofDeposit,
			InvalidProofVerifierCost:	fraud.InvalidFraudProofDeposit,
			Bounded:			true,
		},
	}
	for _, entry := range matrix {
		if err := entry.Validate(); err != nil {
			return nil, err
		}
	}
	return normalizePenaltyMatrix(matrix), nil
}

func BuildGovernedFraudPenaltyPolicy(params PaymentGovernanceParams) (FraudPenaltyPolicy, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return FraudPenaltyPolicy{}, err
	}
	counterpartyBps := uint32(0)
	if params.FraudPenalty.CounterpartyCompensationPriority {
		counterpartyBps = MaxPenaltyRouteBps
	}
	policy := FraudPenaltyPolicy{
		ReporterRewardCap:		params.FraudPenalty.ReporterRewardCap,
		CounterpartyRewardBps:		counterpartyBps,
		BurnShareBps:			params.FraudPenalty.PenaltyBurnAllocationBps,
		SecurityReserveShareBps:	params.FraudPenalty.SecurityReserveAllocationBps,
		CommunityPoolShareBps:		MaxPenaltyRouteBps - params.FraudPenalty.PenaltyBurnAllocationBps - params.FraudPenalty.SecurityReserveAllocationBps,
		SecurityReserveHook:		params.FraudPenalty.SecurityReserveAllocationBps > 0,
	}.Normalize()
	return policy, policy.Validate()
}

func GovernedReporterRewardAmount(penaltyAmount string, params PaymentGovernanceParams) (string, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return "", err
	}
	available, err := parseNonNegativeInt("payments governance reporter reward penalty", penaltyAmount)
	if err != nil {
		return "", err
	}
	reward, err := cappedPenaltyPortion(available, params.FraudPenalty.ReporterRewardCap, params.FraudPenalty.ReporterRewardPercentageBps)
	if err != nil {
		return "", err
	}
	return reward.String(), nil
}

func (p PaymentVirtualChannelGovernanceParams) Normalize(defaults PaymentVirtualChannelGovernanceParams) PaymentVirtualChannelGovernanceParams {
	if p.MaximumVirtualChannelsPerParentChannel == 0 {
		p.MaximumVirtualChannelsPerParentChannel = defaults.MaximumVirtualChannelsPerParentChannel
	}
	if p.MaximumVirtualChannelDepth == 0 {
		p.MaximumVirtualChannelDepth = defaults.MaximumVirtualChannelDepth
	}
	if p.MinimumParentTimeoutMargin == 0 {
		p.MinimumParentTimeoutMargin = defaults.MinimumParentTimeoutMargin
	}
	p.VirtualChannelAnchorFee = normalizeAmountOrDefault(p.VirtualChannelAnchorFee, defaults.VirtualChannelAnchorFee)
	if p.VirtualChannelReservationExpiry == 0 {
		p.VirtualChannelReservationExpiry = defaults.VirtualChannelReservationExpiry
	}
	if p.MultiSegmentVirtualChannelMaxSegments == 0 {
		p.MultiSegmentVirtualChannelMaxSegments = defaults.MultiSegmentVirtualChannelMaxSegments
	}
	return p
}

func (p PaymentVirtualChannelGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Virtual)
	if params.MaximumVirtualChannelsPerParentChannel == 0 {
		return errors.New("payments governance maximum virtual channels per parent must be positive")
	}
	if params.MaximumVirtualChannelDepth == 0 || params.MaximumVirtualChannelDepth > MaxRoutingHops {
		return fmt.Errorf("payments governance virtual channel depth must be between 1 and %d", MaxRoutingHops)
	}
	if params.MinimumParentTimeoutMargin == 0 {
		return errors.New("payments governance parent timeout margin must be positive")
	}
	if err := validateNonNegativeInt("payments governance virtual anchor fee", params.VirtualChannelAnchorFee); err != nil {
		return err
	}
	if params.VirtualChannelReservationExpiry == 0 {
		return errors.New("payments governance virtual reservation expiry must be positive")
	}
	if params.MultiSegmentVirtualChannelMaxSegments == 0 || params.MultiSegmentVirtualChannelMaxSegments > MaxParentChannels {
		return fmt.Errorf("payments governance virtual max segments must be between 1 and %d", MaxParentChannels)
	}
	return nil
}

func (p PaymentVirtualChannelGovernanceParams) ValidateParentTimeoutMargin(parent ChannelRecord, vc VirtualChannel) error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Virtual)
	parent = parent.Normalize()
	vc = vc.Normalize()
	maxHeight := parent.LatestState.Normalize().TimeoutHeight
	if maxHeight == 0 {
		maxHeight = parent.OpenHeight + parent.CloseDelay + parent.DisputePeriod
	}
	if vc.ExpiresHeight+params.MinimumParentTimeoutMargin < vc.ExpiresHeight || vc.ExpiresHeight+params.MinimumParentTimeoutMargin > maxHeight {
		return errors.New("payments virtual expiry does not leave governance parent timeout margin")
	}
	return nil
}

func (p PaymentFraudPenaltyGovernanceParams) Normalize(defaults PaymentFraudPenaltyGovernanceParams) PaymentFraudPenaltyGovernanceParams {
	p.StaleClosePenalty = normalizeAmountOrDefault(p.StaleClosePenalty, defaults.StaleClosePenalty)
	p.SameNonceDoubleSignPenalty = normalizeAmountOrDefault(p.SameNonceDoubleSignPenalty, defaults.SameNonceDoubleSignPenalty)
	p.InvalidConditionPenalty = normalizeAmountOrDefault(p.InvalidConditionPenalty, defaults.InvalidConditionPenalty)
	p.ReplayAttemptPenalty = normalizeAmountOrDefault(p.ReplayAttemptPenalty, defaults.ReplayAttemptPenalty)
	p.InvalidFraudProofDeposit = normalizeAmountOrDefault(p.InvalidFraudProofDeposit, defaults.InvalidFraudProofDeposit)
	if p.ReporterRewardPercentageBps == 0 {
		p.ReporterRewardPercentageBps = defaults.ReporterRewardPercentageBps
	}
	p.ReporterRewardCap = normalizeAmountOrDefault(p.ReporterRewardCap, defaults.ReporterRewardCap)
	if p.PenaltyBurnAllocationBps == 0 && defaults.PenaltyBurnAllocationBps > 0 {
		p.PenaltyBurnAllocationBps = defaults.PenaltyBurnAllocationBps
	}
	if p.SecurityReserveAllocationBps == 0 && defaults.SecurityReserveAllocationBps > 0 {
		p.SecurityReserveAllocationBps = defaults.SecurityReserveAllocationBps
	}
	return p
}

func (p PaymentFraudPenaltyGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().FraudPenalty)
	for _, item := range []struct {
		name	string
		amount	string
	}{
		{"payments governance stale close penalty", params.StaleClosePenalty},
		{"payments governance double sign penalty", params.SameNonceDoubleSignPenalty},
		{"payments governance invalid condition penalty", params.InvalidConditionPenalty},
		{"payments governance replay attempt penalty", params.ReplayAttemptPenalty},
		{"payments governance invalid fraud proof deposit", params.InvalidFraudProofDeposit},
	} {
		if err := validatePositiveInt(item.name, item.amount); err != nil {
			return err
		}
	}
	if params.ReporterRewardPercentageBps == 0 || params.ReporterRewardPercentageBps > MaxPenaltyRouteBps {
		return errors.New("payments governance reporter reward percentage must be between 1 and 10000 bps")
	}
	if err := validateNonNegativeInt("payments governance reporter reward cap", params.ReporterRewardCap); err != nil {
		return err
	}
	if params.PenaltyBurnAllocationBps+params.SecurityReserveAllocationBps > MaxPenaltyRouteBps {
		return errors.New("payments governance penalty burn and security reserve allocations exceed 10000 bps")
	}
	return nil
}

func (p PaymentFraudPenaltyGovernanceParams) matrixEntry(class PaymentPenaltyClass, proofType FraudProofType, source PenaltySource, penalty string) PenaltyMatrixEntry {
	reporterCap, _ := governedReporterRewardCapFromPenalty(penalty, p)
	counterpartyComp := "0"
	if p.CounterpartyCompensationPriority {
		counterpartyComp = penalty
	}
	return PenaltyMatrixEntry{
		Class:				class,
		ProofType:			proofType,
		Source:				source,
		BasePenalty:			penalty,
		ReporterRewardCap:		reporterCap,
		CounterpartyCompensation:	counterpartyComp,
		BurnShareBps:			p.PenaltyBurnAllocationBps,
		SecurityReserveShareBps:	p.SecurityReserveAllocationBps,
		CommunityPoolShareBps:		MaxPenaltyRouteBps - p.PenaltyBurnAllocationBps - p.SecurityReserveAllocationBps,
		Bounded:			true,
	}
}

func activeVirtualChannelsByParent(state PaymentsState) map[string]uint64 {
	counts := map[string]uint64{}
	for _, vc := range state.VirtualChannels {
		vc = vc.Normalize()
		if vc.Status != VirtualChannelStatusOpen {
			continue
		}
		for _, parentID := range vc.ParentChannelIDs {
			counts[parentID]++
		}
	}
	return counts
}

func governedReporterRewardCapFromPenalty(penaltyAmount string, params PaymentFraudPenaltyGovernanceParams) (string, error) {
	penalty, err := parseNonNegativeInt("payments governance penalty amount", penaltyAmount)
	if err != nil {
		return "", err
	}
	reward := penalty.MulRaw(int64(params.ReporterRewardPercentageBps)).QuoRaw(int64(MaxPenaltyRouteBps))
	capAmount, err := parseNonNegativeInt("payments governance reporter reward cap", params.ReporterRewardCap)
	if err != nil {
		return "", err
	}
	if reward.GT(capAmount) {
		reward = capAmount
	}
	return reward.String(), nil
}

func BuildGovernedRoutePolicy(params PaymentGovernanceParams) (RoutePolicy, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return RoutePolicy{}, err
	}
	policy := DefaultRoutePolicy().Normalize()
	policy.StaleLiquidityAfter = params.Routing.LiquidityHintExpiry
	policy.DecayHalfLife = params.Routing.CongestionPenaltyDecay
	return policy.Normalize(), policy.Validate()
}

func BuildGovernedGossipRateLimitPolicy(params PaymentGovernanceParams) (GossipRateLimitPolicy, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return GossipRateLimitPolicy{}, err
	}
	policy := DefaultGossipRateLimitPolicy().Normalize()
	policy.WindowBlocks = params.Routing.GossipMessageExpiry
	policy.MaxTopologyUpdates = params.Routing.MaximumTopologyUpdatesPerPeerWindow
	return policy.Normalize(), policy.Validate()
}

func BuildGovernedRouteFailureScoringPolicy(params PaymentGovernanceParams) (RouteFailureScoringPolicy, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return RouteFailureScoringPolicy{}, err
	}
	policy := DefaultRouteFailureScoringPolicy().Normalize()
	return policy, policy.Validate()
}

func ValidateGossipMessageExpiryWithGovernance(message GossipMessage, params PaymentGovernanceParams) error {
	message = message.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if err := message.ValidateBasic(); err != nil {
		return err
	}
	expiry := params.Routing.GossipMessageExpiry
	if message.MessageType == GossipLiquidityHint {
		expiry = params.Routing.LiquidityHintExpiry
	}
	if message.ValidUntilHeight < message.ValidAfterHeight {
		return errors.New("payments gossip validity window is inverted")
	}
	if message.ValidUntilHeight-message.ValidAfterHeight > expiry {
		return errors.New("payments gossip message expiry exceeds governance maximum")
	}
	return nil
}

func ValidateCapacityProbeRateLimitWithGovernance(existing []CapacityProbeRequest, req CapacityProbeRequest, params PaymentGovernanceParams) error {
	req = req.Normalize()
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	window := params.Routing.CapacityProbeWindow
	windowStart := uint64(1)
	if req.CurrentHeight > window {
		windowStart = req.CurrentHeight - window + 1
	}
	count := uint32(0)
	for _, probe := range existing {
		probe = probe.Normalize()
		if probe.From != req.From || probe.CurrentHeight < windowStart || probe.CurrentHeight > req.CurrentHeight {
			continue
		}
		count++
	}
	if count >= params.Routing.CapacityProbeRateLimit {
		return errors.New("payments capacity probe rate limit exceeded")
	}
	return nil
}

func DecayRouteFailureScoreWithGovernance(score RouteFailureScore, currentHeight uint64, params PaymentGovernanceParams) (RouteFailureScore, error) {
	score.NodeID = strings.TrimSpace(score.NodeID)
	score.ChannelID = normalizeHash(score.ChannelID)
	score.ScoreHash = normalizeOptionalHash(score.ScoreHash)
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return RouteFailureScore{}, err
	}
	if score.NodeID == "" {
		return RouteFailureScore{}, errors.New("payments route failure score node id is required")
	}
	if err := ValidateHash("payments route failure score channel", score.ChannelID); err != nil {
		return RouteFailureScore{}, err
	}
	if !IsRouteFailureClass(score.FailureClass) {
		return RouteFailureScore{}, fmt.Errorf("unknown payments route failure class %q", score.FailureClass)
	}
	if score.FailureCount == 0 || score.ObservedHeight == 0 {
		return RouteFailureScore{}, errors.New("payments route failure score count and height must be positive")
	}
	if currentHeight < score.ObservedHeight {
		return RouteFailureScore{}, errors.New("payments route failure decay height before observation")
	}
	halfLife := params.Routing.RouteFailureScoreDecay
	periods := (currentHeight - score.ObservedHeight) / halfLife
	decayed := score.ScoreDelta
	for i := uint64(0); i < periods && decayed != 0; i++ {
		decayed /= 2
	}
	score.ScoreDelta = decayed
	score.ObservedHeight = currentHeight
	score.ScoreHash = score.Hash()
	return score, nil
}

func NewSettlementBatchWithGovernance(batchID string, operations []SettlementOperation, params PaymentGovernanceParams) (SettlementBatch, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return SettlementBatch{}, err
	}
	if uint64(len(operations)) > params.Execution.SettlementBatchMaximumSize {
		return SettlementBatch{}, fmt.Errorf("payments settlement batch exceeds governance maximum %d", params.Execution.SettlementBatchMaximumSize)
	}
	return NewSettlementBatch(batchID, operations)
}

func ProcessAsyncExecutionQueuesWithGovernance(state PaymentsState, currentHeight uint64, params PaymentGovernanceParams) (PaymentsState, AsyncExecutionResult, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return PaymentsState{}, AsyncExecutionResult{}, err
	}
	return ProcessAsyncExecutionQueues(state, currentHeight, params.Execution.FinalizationQueueWorkLimitPerBlock, params.Execution.ExpiredPromiseCleanupWorkLimitPerBlock)
}

func ApplyGovernedExecutionFeeMultipliers(state PaymentsState, currentHeight, channelOpenCongestionBps, disputeCongestionBps uint64, params PaymentGovernanceParams) (PaymentsState, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return PaymentsState{}, err
	}
	if currentHeight == 0 {
		return PaymentsState{}, errors.New("payments governed fee multiplier height must be positive")
	}
	if channelOpenCongestionBps > 10_000 || disputeCongestionBps > 10_000 {
		return PaymentsState{}, errors.New("payments governed fee multiplier congestion bps exceeds 10000")
	}
	next := state.Export()
	var err error
	if channelOpenCongestionBps > 0 {
		next, err = SetPaymentFeeMultiplier(next, PaymentFeeMultiplier{
			FeeClass:	PaymentFeeClassChannelOpen,
			MultiplierBps:	params.Execution.ChannelOpenCongestionFeeMultiplierBps,
			CongestionBps:	uint32(channelOpenCongestionBps),
			UpdatedHeight:	currentHeight,
		})
		if err != nil {
			return PaymentsState{}, err
		}
	}
	if disputeCongestionBps > 0 {
		next, err = SetPaymentFeeMultiplier(next, PaymentFeeMultiplier{
			FeeClass:	PaymentFeeClassDispute,
			MultiplierBps:	params.Execution.DisputeCongestionFeeMultiplierBps,
			CongestionBps:	uint32(disputeCongestionBps),
			UpdatedHeight:	currentHeight,
		})
		if err != nil {
			return PaymentsState{}, err
		}
	}
	return next, next.Validate()
}

func StorePruneAfterHeightWithGovernance(recordHeight uint64, params PaymentGovernanceParams) (uint64, error) {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if recordHeight == 0 {
		return 0, errors.New("payments store prune record height must be positive")
	}
	if recordHeight+params.Execution.StorePruningHorizon < recordHeight {
		return 0, errors.New("payments store pruning horizon overflows height")
	}
	return recordHeight + params.Execution.StorePruningHorizon, nil
}

func (p PaymentRoutingGovernanceParams) Normalize(defaults PaymentRoutingGovernanceParams) PaymentRoutingGovernanceParams {
	p.RoutingAdvertisementDeposit = normalizeAmountOrDefault(p.RoutingAdvertisementDeposit, defaults.RoutingAdvertisementDeposit)
	if p.GossipMessageExpiry == 0 {
		p.GossipMessageExpiry = defaults.GossipMessageExpiry
	}
	if p.LiquidityHintExpiry == 0 {
		p.LiquidityHintExpiry = defaults.LiquidityHintExpiry
	}
	if p.MaximumTopologyUpdatesPerPeerWindow == 0 {
		p.MaximumTopologyUpdatesPerPeerWindow = defaults.MaximumTopologyUpdatesPerPeerWindow
	}
	if p.RouteFailureScoreDecay == 0 {
		p.RouteFailureScoreDecay = defaults.RouteFailureScoreDecay
	}
	if p.CongestionPenaltyDecay == 0 {
		p.CongestionPenaltyDecay = defaults.CongestionPenaltyDecay
	}
	if p.CapacityProbeRateLimit == 0 {
		p.CapacityProbeRateLimit = defaults.CapacityProbeRateLimit
	}
	if p.CapacityProbeWindow == 0 {
		p.CapacityProbeWindow = defaults.CapacityProbeWindow
	}
	return p
}

func (p PaymentRoutingGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Routing)
	if err := validateNonNegativeInt("payments governance routing advertisement deposit", params.RoutingAdvertisementDeposit); err != nil {
		return err
	}
	if params.GossipMessageExpiry == 0 || params.LiquidityHintExpiry == 0 {
		return errors.New("payments governance gossip expiries must be positive")
	}
	if params.MaximumTopologyUpdatesPerPeerWindow == 0 {
		return errors.New("payments governance topology update limit must be positive")
	}
	if params.RouteFailureScoreDecay == 0 || params.CongestionPenaltyDecay == 0 {
		return errors.New("payments governance routing decay windows must be positive")
	}
	if params.CapacityProbeRateLimit == 0 || params.CapacityProbeWindow == 0 {
		return errors.New("payments governance capacity probe rate limits must be positive")
	}
	return nil
}

func (p PaymentExecutionGovernanceParams) Normalize(defaults PaymentExecutionGovernanceParams) PaymentExecutionGovernanceParams {
	if p.SettlementBatchMaximumSize == 0 {
		p.SettlementBatchMaximumSize = defaults.SettlementBatchMaximumSize
	}
	if p.FinalizationQueueWorkLimitPerBlock == 0 {
		p.FinalizationQueueWorkLimitPerBlock = defaults.FinalizationQueueWorkLimitPerBlock
	}
	if p.ExpiredPromiseCleanupWorkLimitPerBlock == 0 {
		p.ExpiredPromiseCleanupWorkLimitPerBlock = defaults.ExpiredPromiseCleanupWorkLimitPerBlock
	}
	if p.ChannelOpenCongestionFeeMultiplierBps == 0 {
		p.ChannelOpenCongestionFeeMultiplierBps = defaults.ChannelOpenCongestionFeeMultiplierBps
	}
	if p.DisputeCongestionFeeMultiplierBps == 0 {
		p.DisputeCongestionFeeMultiplierBps = defaults.DisputeCongestionFeeMultiplierBps
	}
	if p.StorePruningHorizon == 0 {
		p.StorePruningHorizon = defaults.StorePruningHorizon
	}
	return p
}

func (p PaymentExecutionGovernanceParams) Validate() error {
	params := p.Normalize(DefaultPaymentGovernanceParamsNoHash().Execution)
	if params.SettlementBatchMaximumSize == 0 || params.SettlementBatchMaximumSize > MaxSettlementBatchOps {
		return fmt.Errorf("payments governance settlement batch max must be between 1 and %d", MaxSettlementBatchOps)
	}
	if params.FinalizationQueueWorkLimitPerBlock == 0 || params.ExpiredPromiseCleanupWorkLimitPerBlock == 0 {
		return errors.New("payments governance queue work limits must be positive")
	}
	defaultSchedule := DefaultPaymentFeeSchedule().Normalize()
	if params.ChannelOpenCongestionFeeMultiplierBps < defaultSchedule.BaseMultiplierBps || params.ChannelOpenCongestionFeeMultiplierBps > defaultSchedule.MaxMultiplierBps {
		return errors.New("payments governance channel-open fee multiplier outside schedule bounds")
	}
	if params.DisputeCongestionFeeMultiplierBps < defaultSchedule.BaseMultiplierBps || params.DisputeCongestionFeeMultiplierBps > defaultSchedule.MaxMultiplierBps {
		return errors.New("payments governance dispute fee multiplier outside schedule bounds")
	}
	if params.StorePruningHorizon == 0 {
		return errors.New("payments governance store pruning horizon must be positive")
	}
	return nil
}

func normalizeAmountOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return strings.TrimSpace(fallback)
	}
	return value
}
