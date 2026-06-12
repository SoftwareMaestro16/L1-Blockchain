package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

type ValidatorMarketState struct {
	Params			postypes.Params
	Candidates		[]postypes.Candidate
	Delegations		[]DelegationRecord
	ScoreRecords		[]ValidatorScoreRecord
	SlashHistory		[]ValidatorSlashHistoryRecord
	CommissionHistory	[]ValidatorCommissionRecord
}

type ValidatorSlashHistoryRecord struct {
	EpochID			uint64
	Height			int64
	Validator		string
	Misbehavior		string
	SlashFractionBps	uint32
	SelfBondSlashedNaet	sdkmath.Int
	DelegatorSlashedNaet	sdkmath.Int
	TotalSlashedNaet	sdkmath.Int
}

type ValidatorCommissionRecord struct {
	EpochID		uint64
	Height		int64
	Validator	string
	CommissionBps	uint32
}

type ValidatorRisk struct {
	Validator		string
	SlashEventCount		uint32
	TotalSlashedNaet	sdkmath.Int
	LatestReliabilityBps	uint32
	RiskScoreBps		uint32
	DelegatorRiskInherited	bool
}

type ValidatorEffectiveYield struct {
	Validator		string
	RawStakeNaet		sdkmath.Int
	RewardWeightNaet	sdkmath.Int
	GrossYieldBps		uint32
	NetYieldBps		uint32
	CommissionBps		uint32
	SaturationDampeningBps	uint32
}

type DelegationRiskExposure struct {
	Delegator		string
	Validator		string
	Amount			sdkmath.Int
	RiskAppetite		string
	AdvisoryRiskProfile	bool
	SlashEventsInherited	[]ValidatorSlashHistoryRecord
	ProjectedSlashNaet	sdkmath.Int
	FirstLossProtectedNaet	sdkmath.Int
	HistoricalSlashNaet	sdkmath.Int
}

type DelegationCommissionStatus struct {
	Delegator		string
	Validator		string
	Status			string
	CommissionToleranceBps	uint32
	CurrentCommissionBps	uint32
	CommissionExceeded	bool
	Alert			*DelegationCommissionAlert
}

const (
	DelegationPolicyLowRisk				= "low_risk"
	DelegationPolicyHighAvailability		= "high_availability"
	DelegationPolicyDecentralizationSupporting	= "decentralization_supporting"
	DelegationPolicyMaxYieldWithinRiskBounds	= "max_yield_within_risk_bounds"

	ConcentrationStatusNormal	= "normal"
	ConcentrationStatusWarning	= "warning"

	ValidatorEventMetadataChange	= "validator_metadata_change"
	ValidatorEventCommissionChange	= "validator_commission_change"
	DelegatorAlertCaptureRisk	= "delegator_alert_capture_risk"

	CaptureRiskSuddenDelegationInflow		= "sudden_delegation_inflow"
	CaptureRiskRapidCommissionChange		= "rapid_commission_change"
	CaptureRiskRecentSlash				= "recent_slash"
	CaptureRiskSelfDelegationWithdrawal		= "self_delegation_withdrawal"
	CaptureRiskOperatorMetadataInconsistency	= "operator_metadata_inconsistency"
)

type ValidatorRiskScoreComponents struct {
	Validator		string
	SlashHistoryRiskBps	uint32
	ReliabilityRiskBps	uint32
	ConcentrationRiskBps	uint32
	CommissionRiskBps	uint32
	TotalRiskScoreBps	uint32
}

type DelegationYieldEstimate struct {
	Validator			string
	Delegator			string
	DelegationAmountNaet		sdkmath.Int
	RewardInputNaet			sdkmath.Int
	AdjustedRewardInputNaet		sdkmath.Int
	EstimatedRewardNaet		sdkmath.Int
	GrossYieldBps			uint32
	NetYieldBps			uint32
	CommissionBps			uint32
	PerformanceAdjustmentBps	uint32
	ConcentrationAdjustmentBps	uint32
	ValidatorCommissionNaet		sdkmath.Int
	UsesDistributionInputs		bool
}

type ValidatorDisclosure struct {
	Validator		string
	CommissionBps		uint32
	MaxCommissionChangeBps	uint32
	UptimeBps		uint32
	SlashHistoryCount	uint32
	TotalSlashedNaet	sdkmath.Int
	SelfDelegationNaet	sdkmath.Int
	ConcentrationStatus	string
	ConcentrationWarnings	[]string
}

type DelegationPolicyTemplate struct {
	Name				string
	RiskAppetite			string
	MaxRiskScoreBps			uint32
	MinUptimeBps			uint32
	MaxCommissionBps		uint32
	MaxValidatorShareBps		uint32
	RequireNoSlashHistory		bool
	PreferConcentrationRelief	bool
	AdvisoryOnly			bool
}

type DelegationPolicyEvaluation struct {
	PolicyName	string
	Matches		bool
	Reasons		[]string
	AdvisoryOnly	bool
}

type DelegatorValidatorProfile struct {
	Validator		string
	Risk			ValidatorRisk
	RiskComponents		ValidatorRiskScoreComponents
	YieldEstimate		DelegationYieldEstimate
	ConcentrationWarnings	[]string
	Disclosure		ValidatorDisclosure
	PolicyEvaluations	[]DelegationPolicyEvaluation
	AdvisoryOnly		bool
}

type RedelegationRewardPreview struct {
	Delegator		string
	FromValidator		string
	ToValidator		string
	AmountNaet		sdkmath.Int
	CurrentEstimate		DelegationYieldEstimate
	TargetEstimate		DelegationYieldEstimate
	RewardDeltaNaet		sdkmath.Int
	NetYieldDeltaBps	int32
	TargetRisk		ValidatorRisk
	TargetRiskComponents	ValidatorRiskScoreComponents
	TargetDisclosure	ValidatorDisclosure
	PolicyEvaluations	[]DelegationPolicyEvaluation
	AdvisoryOnly		bool
	StakeMovementExecuted	bool
}

type EventAttribute struct {
	Key	string
	Value	string
}

type ValidatorMachineEvent struct {
	Type		string
	Validator	string
	EpochID		uint64
	Height		int64
	Attributes	[]EventAttribute
}

type ValidatorMetadataSnapshot struct {
	OperatorID	string
	ConsensusKeyID	string
	Moniker		string
	Website		string
	SecurityContact	string
	PayoutAddress	string
}

type ValidatorMetadataChangeInput struct {
	EpochID		uint64
	Height		int64
	Validator	string
	Previous	ValidatorMetadataSnapshot
	Current		ValidatorMetadataSnapshot
	CooldownEpochs	uint64
}

type ValidatorMetadataChangeRecord struct {
	EpochID			uint64
	Height			int64
	Validator		string
	Previous		ValidatorMetadataSnapshot
	Current			ValidatorMetadataSnapshot
	ChangedFields		[]string
	Material		bool
	CooldownUntilEpoch	uint64
	Event			ValidatorMachineEvent
}

type CommissionChangeInput struct {
	EpochID				uint64
	Height				int64
	Validator			string
	PreviousCommissionBps		uint32
	NewCommissionBps		uint32
	MaxIncreaseBpsPerInterval	uint32
	WarningPeriodEpochs		uint64
}

type CommissionChangeRecord struct {
	EpochID			uint64
	Height			int64
	Validator		string
	PreviousCommissionBps	uint32
	NewCommissionBps	uint32
	IncreaseBps		uint32
	RiskFlag		bool
	EffectiveEpoch		uint64
	Event			ValidatorMachineEvent
}

type CaptureRiskParams struct {
	MaterialChangeCooldownEpochs		uint64
	CommissionChangeIntervalEpochs		uint64
	MaxCommissionIncreaseBpsPerInterval	uint32
	SuddenDelegationInflowBps		uint32
	SelfDelegationWithdrawalBps		uint32
	RecentSlashWindowEpochs			uint64
	HighRiskIndicatorThreshold		uint32
}

type CaptureRiskInput struct {
	Validator		string
	CurrentEpoch		uint64
	Height			int64
	PreviousCandidate	postypes.Candidate
	CurrentCandidate	postypes.Candidate
	MetadataChanges		[]ValidatorMetadataChangeRecord
	CommissionHistory	[]ValidatorCommissionRecord
	SlashHistory		[]ValidatorSlashHistoryRecord
	Params			CaptureRiskParams
}

type CaptureRiskIndicator struct {
	Name		string
	SeverityBps	uint32
	Detail		string
}

type DelegatorAlertEvent struct {
	Type		string
	Validator	string
	EpochID		uint64
	Height		int64
	Reason		string
	AdvisoryOnly	bool
}

type CaptureRiskReport struct {
	Validator		string
	RiskScoreBps		uint32
	Indicators		[]CaptureRiskIndicator
	AlertEvents		[]DelegatorAlertEvent
	HighRisk		bool
	AdvisoryOnly		bool
	SlashableBehavior	bool
}

type RiskAdjustedYieldInput struct {
	Delegator			string
	Validator			string
	AmountNaet			sdkmath.Int
	AnnualRewardsNaet		sdkmath.Int
	UnbondingLiquidityCostBps	uint32
	Decentralization		DecentralizationParams
	ActiveValidatorIDs		[]string
}

type RiskAdjustedYieldProjection struct {
	Validator			string
	Delegator			string
	GrossRewardRateBps		uint32
	CommissionBps			uint32
	HistoricalUptimeBps		uint32
	SlashProbabilityProxyBps	uint32
	ConcentrationAdjustmentBps	uint32
	UnbondingLiquidityCostBps	uint32
	RiskAdjustedYieldBps		uint32
	ExpectedRewardNaet		sdkmath.Int
	LowRewardNaet			sdkmath.Int
	HighRewardNaet			sdkmath.Int
	UncertaintyBps			uint32
	ReproducibleFromQueries		bool
}

type ValidatorRewardObservation struct {
	EpochID			uint64
	Validator		string
	RewardPerStakeBps	uint32
}

type ValidatorRewardVariance struct {
	Validator		string
	ObservationCount	uint32
	MeanRewardBps		uint32
	MinRewardBps		uint32
	MaxRewardBps		uint32
	VarianceBps		uint32
}

type RedelegationCostEstimate struct {
	AmountNaet		sdkmath.Int
	UnbondingSeconds	uint64
	LiquidityCostBps	uint32
	CommissionDeltaBps	int32
	RiskDeltaBps		int32
	EstimatedCostNaet	sdkmath.Int
}

type DelegationSimulationInput struct {
	Delegator			string
	FromValidator			string
	ToValidator			string
	AmountNaet			sdkmath.Int
	AnnualRewardsNaet		sdkmath.Int
	CurrentEpoch			uint64
	Height				int64
	CommissionOverrides		[]ValidatorCommissionRecord
	SlashEvents			[]ValidatorSlashHistoryRecord
	Decentralization		DecentralizationParams
	ActiveValidatorIDs		[]string
	UnbondingLiquidityCostBps	uint32
}

type DelegationSimulationResult struct {
	CurrentProjection		RiskAdjustedYieldProjection
	TargetProjection		RiskAdjustedYieldProjection
	RedelegationPreview		RedelegationRewardPreview
	RedelegationCost		RedelegationCostEstimate
	ProjectedRewardDeltaNaet	sdkmath.Int
	AdvisoryOnly			bool
}

type FirstLossSelfBondAccounting struct {
	Validator			string
	TargetSlashNaet			sdkmath.Int
	SelfBondAvailableNaet		sdkmath.Int
	SelfBondAbsorbedNaet		sdkmath.Int
	DelegatorResidualSlashNaet	sdkmath.Int
	FirstLossApplied		bool
}

type SlashPropagationInput struct {
	Validator		string
	SelfBondNaet		sdkmath.Int
	Delegations		[]DelegationRecord
	SlashFractionBps	uint32
	SelfBondFirstLoss	bool
	EvidenceHeight		int64
	Misbehavior		string
	EpochID			uint64
}

type SlashPropagationResult struct {
	Validator		string
	SelfBondSlashedNaet	sdkmath.Int
	DelegatorSlashes	[]DelegatorSlashExposure
	TotalDelegatorSlashed	sdkmath.Int
	TotalSlashedNaet	sdkmath.Int
}

type DelegatorSlashExposure struct {
	Delegator	string
	Validator	string
	Amount		sdkmath.Int
	SlashedNaet	sdkmath.Int
	RiskTranche	string
}

const (
	SlashSeverityMinorDowntime			SlashSeverityClass	= "minor_downtime"
	SlashSeverityMajorDowntime			SlashSeverityClass	= "major_downtime"
	SlashSeverityRepeatedDowntime			SlashSeverityClass	= "repeated_downtime"
	SlashSeverityEquivocation			SlashSeverityClass	= "equivocation"
	SlashSeverityEvidenceManipulation		SlashSeverityClass	= "evidence_manipulation"
	SlashSeverityKeyCompromiseResponseFailure	SlashSeverityClass	= "key_compromise_response_failure"
	DefaultRepeatOffenseDecayEpochs						= uint64(4)
	DefaultRepeatOffenseStepBps						= uint32(2_500)
	DefaultMaxRepeatOffenseMultiplierBps					= uint32(30_000)

	ConcentrationWarningValidatorShare		= "validator_voting_power_concentration"
	ConcentrationWarningTopNShare			= "top_n_voting_power_concentration"
	ConcentrationWarningDelegatorShare		= "delegator_concentration"
	ConcentrationWarningSelfDelegation		= "self_delegation_below_requirement"
	ConcentrationWarningCommissionByPower		= "commission_concentration_by_power"
	ConcentrationWarningRewardDampeningActive	= "reward_dampening_active"
)

type SlashSeverityClass string

type SlashingSeverityParam struct {
	Severity		SlashSeverityClass
	BasePenaltyBps		uint32
	BurnBps			uint32
	TreasuryBps		uint32
	ReporterRewardBps	uint32
	ReporterRewardCapBps	uint32
}

type SlashingEvidence struct {
	EvidenceID	string
	ReporterID	string
	Accepted	bool
	Duplicate	bool
}

type SlashingRoutingInput struct {
	Validator		string
	Severity		SlashSeverityClass
	TotalStakeNaet		sdkmath.Int
	Evidence		SlashingEvidence
	PriorOffenseEpochs	[]uint64
	CurrentEpoch		uint64
	Params			[]SlashingSeverityParam
}

type SlashingRoutingResult struct {
	Validator		string
	Severity		SlashSeverityClass
	PenaltyBps		uint32
	RepeatMultiplierBps	uint32
	PenaltyNaet		sdkmath.Int
	BurnNaet		sdkmath.Int
	TreasuryNaet		sdkmath.Int
	ReporterRewardNaet	sdkmath.Int
	ValidatorResidualNaet	sdkmath.Int
	ReporterPaid		bool
	Event			SlashingEvent
}

type SlashingEvent struct {
	Validator		string
	Severity		SlashSeverityClass
	Reason			string
	PenaltyNaet		sdkmath.Int
	BurnNaet		sdkmath.Int
	TreasuryNaet		sdkmath.Int
	ReporterRewardNaet	sdkmath.Int
	ReporterID		string
	ReporterPaid		bool
}

type DecentralizationParams struct {
	TopN				int
	MaxValidatorShareBps		uint32
	MaxTopNShareBps			uint32
	MaxDelegatorConcentrationBps	uint32
	MinSelfDelegationRatioBps	uint32
	MinSelfDelegationNaet		sdkmath.Int
	MaxCommissionWeightedBps	uint32
	RewardDampeningSafetyFloorBps	uint32
	StakeMovementIncentiveBps	uint32
}

type ValidatorConcentrationMetric struct {
	Validator			string
	VotingPowerShareBps		uint32
	DelegatorConcentrationBps	uint32
	SelfDelegationRatioBps		uint32
	CommissionBps			uint32
	RewardDampeningBps		uint32
	StakeMovementIncentiveBps	uint32
	Warnings			[]string
}

type ActiveSetConcentrationReport struct {
	Metrics			[]ValidatorConcentrationMetric
	TopNShareBps		uint32
	CommissionWeightedBps	uint32
	ActiveSetWarnings	[]string
}

func NewValidatorMarketState(params postypes.Params, candidates []postypes.Candidate, delegations []DelegationRecord, scoreRecords []ValidatorScoreRecord, slashHistory []ValidatorSlashHistoryRecord, commissionHistory []ValidatorCommissionRecord) (ValidatorMarketState, error) {
	if err := params.Validate(); err != nil {
		return ValidatorMarketState{}, err
	}
	candidateCopies := make([]postypes.Candidate, len(candidates))
	for i, candidate := range candidates {
		if err := candidate.Validate(params); err != nil {
			return ValidatorMarketState{}, err
		}
		candidateCopies[i] = candidate
	}
	delegationState, err := NewDelegationCapitalState(params, delegations)
	if err != nil {
		return ValidatorMarketState{}, err
	}
	scoreState, err := NewScoreComponentState(scoreRecords)
	if err != nil {
		return ValidatorMarketState{}, err
	}
	slashCopies := make([]ValidatorSlashHistoryRecord, len(slashHistory))
	for i, record := range slashHistory {
		record.Validator = strings.TrimSpace(record.Validator)
		if err := record.Validate(params); err != nil {
			return ValidatorMarketState{}, err
		}
		slashCopies[i] = record
	}
	sortSlashHistory(slashCopies)
	commissionCopies := make([]ValidatorCommissionRecord, len(commissionHistory))
	for i, record := range commissionHistory {
		record.Validator = strings.TrimSpace(record.Validator)
		if err := record.Validate(params); err != nil {
			return ValidatorMarketState{}, err
		}
		commissionCopies[i] = record
	}
	sortCommissionHistory(commissionCopies)
	return ValidatorMarketState{
		Params:			params,
		Candidates:		candidateCopies,
		Delegations:		delegationState.Records,
		ScoreRecords:		scoreState.Records,
		SlashHistory:		slashCopies,
		CommissionHistory:	commissionCopies,
	}, nil
}

func (r ValidatorSlashHistoryRecord) Validate(params postypes.Params) error {
	if r.EpochID == 0 {
		return errors.New("slash history epoch id is required")
	}
	if r.Height < 0 {
		return errors.New("slash history height cannot be negative")
	}
	if err := validateEconomyToken("slash history validator", r.Validator); err != nil {
		return err
	}
	if !postypes.IsSlashableMisbehavior(r.Misbehavior) {
		return fmt.Errorf("unsupported slash history misbehavior %q", r.Misbehavior)
	}
	if r.SlashFractionBps == 0 || r.SlashFractionBps > postypes.BasisPoints {
		return fmt.Errorf("slash history fraction must be within 1..%d bps", postypes.BasisPoints)
	}
	if r.SelfBondSlashedNaet.IsNegative() || r.DelegatorSlashedNaet.IsNegative() || r.TotalSlashedNaet.IsNegative() {
		return errors.New("slash history amounts cannot be negative")
	}
	if !r.SelfBondSlashedNaet.Add(r.DelegatorSlashedNaet).Equal(r.TotalSlashedNaet) {
		return errors.New("slash history total must equal self bond plus delegator slashes")
	}
	return params.Validate()
}

func (r ValidatorCommissionRecord) Validate(params postypes.Params) error {
	if r.EpochID == 0 {
		return errors.New("commission history epoch id is required")
	}
	if r.Height < 0 {
		return errors.New("commission history height cannot be negative")
	}
	if err := validateEconomyToken("commission history validator", r.Validator); err != nil {
		return err
	}
	if r.CommissionBps > params.MaxCommissionBps {
		return fmt.Errorf("commission history bps must be <= %d", params.MaxCommissionBps)
	}
	return nil
}

func DefaultSlashingSeverityParams() []SlashingSeverityParam {
	return []SlashingSeverityParam{
		{Severity: SlashSeverityMinorDowntime, BasePenaltyBps: 100, BurnBps: 2_000, TreasuryBps: 7_500, ReporterRewardBps: 500, ReporterRewardCapBps: 500},
		{Severity: SlashSeverityMajorDowntime, BasePenaltyBps: 500, BurnBps: 3_000, TreasuryBps: 6_000, ReporterRewardBps: 1_000, ReporterRewardCapBps: 1_000},
		{Severity: SlashSeverityRepeatedDowntime, BasePenaltyBps: 1_000, BurnBps: 4_000, TreasuryBps: 5_000, ReporterRewardBps: 1_000, ReporterRewardCapBps: 1_000},
		{Severity: SlashSeverityEquivocation, BasePenaltyBps: 5_000, BurnBps: 5_000, TreasuryBps: 4_000, ReporterRewardBps: 1_000, ReporterRewardCapBps: 1_500},
		{Severity: SlashSeverityEvidenceManipulation, BasePenaltyBps: 7_500, BurnBps: 6_000, TreasuryBps: 3_000, ReporterRewardBps: 1_000, ReporterRewardCapBps: 1_500},
		{Severity: SlashSeverityKeyCompromiseResponseFailure, BasePenaltyBps: 3_000, BurnBps: 4_000, TreasuryBps: 5_000, ReporterRewardBps: 1_000, ReporterRewardCapBps: 1_500},
	}
}

func (s SlashSeverityClass) Validate() error {
	switch s {
	case SlashSeverityMinorDowntime, SlashSeverityMajorDowntime, SlashSeverityRepeatedDowntime,
		SlashSeverityEquivocation, SlashSeverityEvidenceManipulation, SlashSeverityKeyCompromiseResponseFailure:
		return nil
	default:
		return fmt.Errorf("unsupported slash severity %q", s)
	}
}

func (p SlashingSeverityParam) Validate() error {
	if err := p.Severity.Validate(); err != nil {
		return err
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "base_penalty_bps", value: p.BasePenaltyBps},
		{name: "burn_bps", value: p.BurnBps},
		{name: "treasury_bps", value: p.TreasuryBps},
		{name: "reporter_reward_bps", value: p.ReporterRewardBps},
		{name: "reporter_reward_cap_bps", value: p.ReporterRewardCapBps},
	} {
		if item.value > postypes.BasisPoints {
			return fmt.Errorf("%s must be <= %d bps", item.name, postypes.BasisPoints)
		}
	}
	if p.BasePenaltyBps == 0 {
		return errors.New("base penalty bps is required")
	}
	if uint64(p.BurnBps)+uint64(p.TreasuryBps)+uint64(p.ReporterRewardBps) > uint64(postypes.BasisPoints) {
		return errors.New("slashing routing bps cannot exceed 10000")
	}
	return nil
}

func ComputeRepeatOffenseMultiplier(currentEpoch uint64, priorOffenseEpochs []uint64, decayEpochs uint64, stepBps uint32, maxBps uint32) (uint32, error) {
	if currentEpoch == 0 {
		return 0, errors.New("current epoch is required")
	}
	if decayEpochs == 0 {
		decayEpochs = DefaultRepeatOffenseDecayEpochs
	}
	if stepBps == 0 {
		stepBps = DefaultRepeatOffenseStepBps
	}
	if maxBps == 0 {
		maxBps = DefaultMaxRepeatOffenseMultiplierBps
	}
	if maxBps < postypes.BasisPoints {
		return 0, errors.New("max repeat offense multiplier cannot be below 10000 bps")
	}
	activeOffenses := uint32(0)
	for _, epoch := range priorOffenseEpochs {
		if epoch == 0 || epoch > currentEpoch {
			continue
		}
		if currentEpoch-epoch <= decayEpochs {
			activeOffenses++
		}
	}
	multiplier := uint64(postypes.BasisPoints) + uint64(activeOffenses)*uint64(stepBps)
	if multiplier > uint64(maxBps) {
		multiplier = uint64(maxBps)
	}
	return uint32(multiplier), nil
}

func RouteSlashing(input SlashingRoutingInput) (SlashingRoutingResult, error) {
	validator := strings.TrimSpace(input.Validator)
	if err := validateEconomyToken("slashing validator", validator); err != nil {
		return SlashingRoutingResult{}, err
	}
	if err := input.Severity.Validate(); err != nil {
		return SlashingRoutingResult{}, err
	}
	totalStake := normalizeEconomyInt(input.TotalStakeNaet)
	if !totalStake.IsPositive() {
		return SlashingRoutingResult{}, errors.New("slashing total stake must be positive")
	}
	params := input.Params
	if len(params) == 0 {
		params = DefaultSlashingSeverityParams()
	}
	param, found, err := findSlashingSeverityParam(params, input.Severity)
	if err != nil {
		return SlashingRoutingResult{}, err
	}
	if !found {
		return SlashingRoutingResult{}, fmt.Errorf("missing slashing params for severity %q", input.Severity)
	}
	multiplier, err := ComputeRepeatOffenseMultiplier(input.CurrentEpoch, input.PriorOffenseEpochs, DefaultRepeatOffenseDecayEpochs, DefaultRepeatOffenseStepBps, DefaultMaxRepeatOffenseMultiplierBps)
	if err != nil {
		return SlashingRoutingResult{}, err
	}
	penaltyBps := uint32(uint64(param.BasePenaltyBps) * uint64(multiplier) / uint64(postypes.BasisPoints))
	penaltyBps = minUint32(penaltyBps, postypes.BasisPoints)
	penalty := mulIntBps(totalStake, penaltyBps)
	burn := mulIntBps(penalty, param.BurnBps)
	treasury := mulIntBps(penalty, param.TreasuryBps)
	reporter := sdkmath.ZeroInt()
	reporterPaid := false
	reporterID := strings.TrimSpace(input.Evidence.ReporterID)
	if input.Evidence.Accepted && !input.Evidence.Duplicate && reporterID != "" {
		reporter = mulIntBps(penalty, param.ReporterRewardBps)
		capAmount := mulIntBps(penalty, param.ReporterRewardCapBps)
		if reporter.GT(capAmount) {
			reporter = capAmount
		}
		reporterPaid = reporter.IsPositive()
	}
	routed := burn.Add(treasury).Add(reporter)
	if routed.GT(penalty) {
		return SlashingRoutingResult{}, errors.New("slashing routing exceeds penalty amount")
	}
	residual := penalty.Sub(routed)
	event := SlashingEvent{
		Validator:		validator,
		Severity:		input.Severity,
		Reason:			string(input.Severity),
		PenaltyNaet:		penalty,
		BurnNaet:		burn,
		TreasuryNaet:		treasury,
		ReporterRewardNaet:	reporter,
		ReporterID:		reporterID,
		ReporterPaid:		reporterPaid,
	}
	result := SlashingRoutingResult{
		Validator:		validator,
		Severity:		input.Severity,
		PenaltyBps:		penaltyBps,
		RepeatMultiplierBps:	multiplier,
		PenaltyNaet:		penalty,
		BurnNaet:		burn,
		TreasuryNaet:		treasury,
		ReporterRewardNaet:	reporter,
		ValidatorResidualNaet:	residual,
		ReporterPaid:		reporterPaid,
		Event:			event,
	}
	if !result.BurnNaet.Add(result.TreasuryNaet).Add(result.ReporterRewardNaet).Add(result.ValidatorResidualNaet).Equal(result.PenaltyNaet) {
		return SlashingRoutingResult{}, errors.New("slashing routing invariant failed")
	}
	return result, nil
}

func findSlashingSeverityParam(params []SlashingSeverityParam, severity SlashSeverityClass) (SlashingSeverityParam, bool, error) {
	seen := make(map[SlashSeverityClass]struct{}, len(params))
	for _, param := range params {
		if err := param.Validate(); err != nil {
			return SlashingSeverityParam{}, false, err
		}
		if _, duplicate := seen[param.Severity]; duplicate {
			return SlashingSeverityParam{}, false, fmt.Errorf("duplicate slashing params for severity %q", param.Severity)
		}
		seen[param.Severity] = struct{}{}
		if param.Severity == severity {
			return param, true, nil
		}
	}
	return SlashingSeverityParam{}, false, nil
}

func DefaultDecentralizationParams(params postypes.Params) DecentralizationParams {
	return DecentralizationParams{
		TopN:				3,
		MaxValidatorShareBps:		params.MaxVotingPowerBps,
		MaxTopNShareBps:		6_700,
		MaxDelegatorConcentrationBps:	5_000,
		MinSelfDelegationRatioBps:	500,
		MinSelfDelegationNaet:		params.MinStakeNaet,
		MaxCommissionWeightedBps:	params.MaxCommissionBps,
		RewardDampeningSafetyFloorBps:	7_500,
		StakeMovementIncentiveBps:	250,
	}
}

func (p DecentralizationParams) Validate(posParams postypes.Params) error {
	if err := posParams.Validate(); err != nil {
		return err
	}
	if p.TopN <= 0 {
		return errors.New("top n must be positive")
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "max_validator_share_bps", value: p.MaxValidatorShareBps},
		{name: "max_top_n_share_bps", value: p.MaxTopNShareBps},
		{name: "max_delegator_concentration_bps", value: p.MaxDelegatorConcentrationBps},
		{name: "min_self_delegation_ratio_bps", value: p.MinSelfDelegationRatioBps},
		{name: "max_commission_weighted_bps", value: p.MaxCommissionWeightedBps},
		{name: "reward_dampening_safety_floor_bps", value: p.RewardDampeningSafetyFloorBps},
		{name: "stake_movement_incentive_bps", value: p.StakeMovementIncentiveBps},
	} {
		if item.value > postypes.BasisPoints {
			return fmt.Errorf("%s must be <= %d bps", item.name, postypes.BasisPoints)
		}
	}
	if p.MaxValidatorShareBps == 0 || p.MaxTopNShareBps == 0 || p.RewardDampeningSafetyFloorBps == 0 {
		return errors.New("concentration thresholds and reward safety floor are required")
	}
	if !normalizeEconomyInt(p.MinSelfDelegationNaet).IsPositive() {
		return errors.New("minimum self delegation amount must be positive")
	}
	return nil
}

func DefaultCaptureRiskParams(params postypes.Params) CaptureRiskParams {
	return CaptureRiskParams{
		MaterialChangeCooldownEpochs:		2,
		CommissionChangeIntervalEpochs:		4,
		MaxCommissionIncreaseBpsPerInterval:	500,
		SuddenDelegationInflowBps:		3_000,
		SelfDelegationWithdrawalBps:		2_500,
		RecentSlashWindowEpochs:		params.EvidenceWindowEpochs,
		HighRiskIndicatorThreshold:		2,
	}
}

func (p CaptureRiskParams) Validate(params postypes.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if p.MaterialChangeCooldownEpochs == 0 {
		return errors.New("material change cooldown epochs must be positive")
	}
	if p.CommissionChangeIntervalEpochs == 0 {
		return errors.New("commission change interval epochs must be positive")
	}
	if p.RecentSlashWindowEpochs == 0 {
		return errors.New("recent slash window epochs must be positive")
	}
	if p.MaxCommissionIncreaseBpsPerInterval > params.MaxCommissionBps {
		return fmt.Errorf("max commission increase must be <= %d bps", params.MaxCommissionBps)
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "sudden_delegation_inflow_bps", value: p.SuddenDelegationInflowBps},
		{name: "self_delegation_withdrawal_bps", value: p.SelfDelegationWithdrawalBps},
		{name: "high_risk_indicator_threshold", value: p.HighRiskIndicatorThreshold},
	} {
		if item.value == 0 {
			return fmt.Errorf("%s must be positive", item.name)
		}
		if item.value > postypes.BasisPoints {
			return fmt.Errorf("%s must be <= %d bps", item.name, postypes.BasisPoints)
		}
	}
	return nil
}

func TrackValidatorMetadataChange(input ValidatorMetadataChangeInput) (ValidatorMetadataChangeRecord, bool, error) {
	validator := strings.TrimSpace(input.Validator)
	if input.EpochID == 0 {
		return ValidatorMetadataChangeRecord{}, false, errors.New("metadata change epoch id is required")
	}
	if input.Height < 0 {
		return ValidatorMetadataChangeRecord{}, false, errors.New("metadata change height cannot be negative")
	}
	if err := validateEconomyToken("metadata change validator", validator); err != nil {
		return ValidatorMetadataChangeRecord{}, false, err
	}
	if input.CooldownEpochs == 0 {
		input.CooldownEpochs = DefaultCaptureRiskParams(postypes.DefaultParams()).MaterialChangeCooldownEpochs
	}
	changed := metadataChangedFields(input.Previous, input.Current)
	if len(changed) == 0 {
		return ValidatorMetadataChangeRecord{}, false, nil
	}
	material := metadataChangeMaterial(changed)
	cooldownUntil := input.EpochID
	if material {
		cooldownUntil += input.CooldownEpochs
	}
	attrs := []EventAttribute{
		{Key: "changed_fields", Value: strings.Join(changed, ",")},
		{Key: "material", Value: boolString(material)},
		{Key: "cooldown_until_epoch", Value: fmt.Sprintf("%d", cooldownUntil)},
	}
	record := ValidatorMetadataChangeRecord{
		EpochID:		input.EpochID,
		Height:			input.Height,
		Validator:		validator,
		Previous:		normalizeMetadataSnapshot(input.Previous),
		Current:		normalizeMetadataSnapshot(input.Current),
		ChangedFields:		changed,
		Material:		material,
		CooldownUntilEpoch:	cooldownUntil,
		Event: ValidatorMachineEvent{
			Type:		ValidatorEventMetadataChange,
			Validator:	validator,
			EpochID:	input.EpochID,
			Height:		input.Height,
			Attributes:	attrs,
		},
	}
	return record, true, nil
}

func TrackValidatorCommissionChange(input CommissionChangeInput) (CommissionChangeRecord, bool, error) {
	validator := strings.TrimSpace(input.Validator)
	if input.EpochID == 0 {
		return CommissionChangeRecord{}, false, errors.New("commission change epoch id is required")
	}
	if input.Height < 0 {
		return CommissionChangeRecord{}, false, errors.New("commission change height cannot be negative")
	}
	if err := validateEconomyToken("commission change validator", validator); err != nil {
		return CommissionChangeRecord{}, false, err
	}
	if input.PreviousCommissionBps > postypes.DefaultMaxCommissionBps || input.NewCommissionBps > postypes.DefaultMaxCommissionBps {
		return CommissionChangeRecord{}, false, errors.New("commission change cannot exceed 20%")
	}
	if input.WarningPeriodEpochs == 0 {
		input.WarningPeriodEpochs = 1
	}
	if input.PreviousCommissionBps == input.NewCommissionBps {
		return CommissionChangeRecord{}, false, nil
	}
	increase := uint32(0)
	if input.NewCommissionBps > input.PreviousCommissionBps {
		increase = input.NewCommissionBps - input.PreviousCommissionBps
	}
	riskFlag := input.MaxIncreaseBpsPerInterval > 0 && increase > input.MaxIncreaseBpsPerInterval
	record := CommissionChangeRecord{
		EpochID:		input.EpochID,
		Height:			input.Height,
		Validator:		validator,
		PreviousCommissionBps:	input.PreviousCommissionBps,
		NewCommissionBps:	input.NewCommissionBps,
		IncreaseBps:		increase,
		RiskFlag:		riskFlag,
		EffectiveEpoch:		input.EpochID + input.WarningPeriodEpochs,
		Event: ValidatorMachineEvent{
			Type:		ValidatorEventCommissionChange,
			Validator:	validator,
			EpochID:	input.EpochID,
			Height:		input.Height,
			Attributes: []EventAttribute{
				{Key: "previous_commission_bps", Value: fmt.Sprintf("%d", input.PreviousCommissionBps)},
				{Key: "new_commission_bps", Value: fmt.Sprintf("%d", input.NewCommissionBps)},
				{Key: "increase_bps", Value: fmt.Sprintf("%d", increase)},
				{Key: "effective_epoch", Value: fmt.Sprintf("%d", input.EpochID+input.WarningPeriodEpochs)},
				{Key: "risk_flag", Value: boolString(riskFlag)},
			},
		},
	}
	return record, true, nil
}

func EvaluateValidatorCaptureRisk(input CaptureRiskInput) (CaptureRiskReport, error) {
	validator := strings.TrimSpace(input.Validator)
	if err := validateEconomyToken("capture risk validator", validator); err != nil {
		return CaptureRiskReport{}, err
	}
	if input.CurrentEpoch == 0 {
		return CaptureRiskReport{}, errors.New("capture risk current epoch is required")
	}
	params := input.Params
	if params.MaterialChangeCooldownEpochs == 0 {
		params = DefaultCaptureRiskParams(postypes.DefaultParams())
	}
	if err := params.Validate(postypes.DefaultParams()); err != nil {
		return CaptureRiskReport{}, err
	}
	indicators := make([]CaptureRiskIndicator, 0, 5)
	prevDelegated := normalizeEconomyInt(input.PreviousCandidate.DelegatedStakeNaet)
	currentDelegated := normalizeEconomyInt(input.CurrentCandidate.DelegatedStakeNaet)
	if currentDelegated.GT(prevDelegated) {
		inflow := currentDelegated.Sub(prevDelegated)
		inflowBps := shareBps(inflow, maxPositiveInt(prevDelegated, sdkmath.NewInt(1)))
		if prevDelegated.IsZero() && inflow.IsPositive() {
			inflowBps = postypes.BasisPoints
		}
		if inflowBps >= params.SuddenDelegationInflowBps {
			indicators = append(indicators, CaptureRiskIndicator{Name: CaptureRiskSuddenDelegationInflow, SeverityBps: inflowBps, Detail: "delegated stake increased sharply"})
		}
	}
	prevSelf := normalizeEconomyInt(input.PreviousCandidate.SelfStakeNaet)
	currentSelf := normalizeEconomyInt(input.CurrentCandidate.SelfStakeNaet)
	if prevSelf.IsPositive() && currentSelf.LT(prevSelf) {
		withdrawalBps := shareBps(prevSelf.Sub(currentSelf), prevSelf)
		if withdrawalBps >= params.SelfDelegationWithdrawalBps {
			indicators = append(indicators, CaptureRiskIndicator{Name: CaptureRiskSelfDelegationWithdrawal, SeverityBps: withdrawalBps, Detail: "validator self-delegation decreased"})
		}
	}
	if rapid, increaseBps := rapidCommissionIncrease(input.CommissionHistory, validator, input.CurrentEpoch, params.CommissionChangeIntervalEpochs, params.MaxCommissionIncreaseBpsPerInterval); rapid {
		indicators = append(indicators, CaptureRiskIndicator{Name: CaptureRiskRapidCommissionChange, SeverityBps: increaseBps, Detail: "commission increased faster than governance interval limit"})
	}
	if recentSlash := recentSlashCount(input.SlashHistory, validator, input.CurrentEpoch, params.RecentSlashWindowEpochs); recentSlash > 0 {
		indicators = append(indicators, CaptureRiskIndicator{Name: CaptureRiskRecentSlash, SeverityBps: minBps(uint64(recentSlash) * 1_000), Detail: "validator has recent slash history"})
	}
	for _, change := range input.MetadataChanges {
		if change.Validator == validator && change.Material && input.CurrentEpoch <= change.CooldownUntilEpoch {
			indicators = append(indicators, CaptureRiskIndicator{Name: CaptureRiskOperatorMetadataInconsistency, SeverityBps: 2_500, Detail: strings.Join(change.ChangedFields, ",")})
			break
		}
	}
	sortCaptureRiskIndicators(indicators)
	score := uint64(0)
	for _, indicator := range indicators {
		score += uint64(indicator.SeverityBps)
	}
	report := CaptureRiskReport{
		Validator:		validator,
		RiskScoreBps:		minBps(score),
		Indicators:		indicators,
		HighRisk:		uint32(len(indicators)) >= params.HighRiskIndicatorThreshold,
		AdvisoryOnly:		true,
		SlashableBehavior:	false,
	}
	if report.HighRisk {
		for _, indicator := range indicators {
			report.AlertEvents = append(report.AlertEvents, DelegatorAlertEvent{
				Type:		DelegatorAlertCaptureRisk,
				Validator:	validator,
				EpochID:	input.CurrentEpoch,
				Height:		input.Height,
				Reason:		indicator.Name,
				AdvisoryOnly:	true,
			})
		}
	}
	return report, nil
}

func PropagateSlash(input SlashPropagationInput) (SlashPropagationResult, error) {
	if err := validateEconomyToken("slash validator", input.Validator); err != nil {
		return SlashPropagationResult{}, err
	}
	if input.SelfBondNaet.IsNegative() {
		return SlashPropagationResult{}, errors.New("self bond cannot be negative")
	}
	if input.SlashFractionBps == 0 || input.SlashFractionBps > postypes.BasisPoints {
		return SlashPropagationResult{}, fmt.Errorf("slash fraction must be within 1..%d bps", postypes.BasisPoints)
	}
	validatorDelegations := filterDelegationsForValidator(input.Delegations, input.Validator)
	totalDelegated := sdkmath.ZeroInt()
	for _, record := range validatorDelegations {
		totalDelegated = totalDelegated.Add(record.Amount)
	}
	totalStake := input.SelfBondNaet.Add(totalDelegated)
	targetSlash := mulIntBps(totalStake, input.SlashFractionBps)
	result := SlashPropagationResult{
		Validator:		strings.TrimSpace(input.Validator),
		SelfBondSlashedNaet:	mulIntBps(input.SelfBondNaet, input.SlashFractionBps),
		DelegatorSlashes:	make([]DelegatorSlashExposure, 0, len(validatorDelegations)),
		TotalDelegatorSlashed:	sdkmath.ZeroInt(),
		TotalSlashedNaet:	sdkmath.ZeroInt(),
	}
	if input.SelfBondFirstLoss && targetSlash.GT(result.SelfBondSlashedNaet) {
		if input.SelfBondNaet.GTE(targetSlash) {
			result.SelfBondSlashedNaet = targetSlash
		} else {
			result.SelfBondSlashedNaet = input.SelfBondNaet
		}
	}
	remaining := targetSlash.Sub(result.SelfBondSlashedNaet)
	for _, record := range validatorDelegations {
		slashed := sdkmath.ZeroInt()
		if remaining.IsPositive() && totalDelegated.IsPositive() {
			slashed = shareByStake(remaining, record.Amount, totalDelegated)
		}
		result.DelegatorSlashes = append(result.DelegatorSlashes, DelegatorSlashExposure{
			Delegator:	record.Delegator,
			Validator:	record.Validator,
			Amount:		record.Amount,
			SlashedNaet:	slashed,
			RiskTranche:	record.RiskTrancheOptional,
		})
		result.TotalDelegatorSlashed = result.TotalDelegatorSlashed.Add(slashed)
	}
	result.TotalSlashedNaet = result.SelfBondSlashedNaet.Add(result.TotalDelegatorSlashed)
	return result, nil
}

func BuildFirstLossSelfBondAccounting(input SlashPropagationInput) (FirstLossSelfBondAccounting, error) {
	propagation, err := PropagateSlash(input)
	if err != nil {
		return FirstLossSelfBondAccounting{}, err
	}
	validatorDelegations := filterDelegationsForValidator(input.Delegations, input.Validator)
	totalDelegated := sdkmath.ZeroInt()
	for _, record := range validatorDelegations {
		totalDelegated = totalDelegated.Add(record.Amount)
	}
	targetSlash := mulIntBps(input.SelfBondNaet.Add(totalDelegated), input.SlashFractionBps)
	return FirstLossSelfBondAccounting{
		Validator:			strings.TrimSpace(input.Validator),
		TargetSlashNaet:		targetSlash,
		SelfBondAvailableNaet:		input.SelfBondNaet,
		SelfBondAbsorbedNaet:		propagation.SelfBondSlashedNaet,
		DelegatorResidualSlashNaet:	propagation.TotalDelegatorSlashed,
		FirstLossApplied:		input.SelfBondFirstLoss,
	}, nil
}

func (s ValidatorMarketState) QueryValidatorRisk(validator string) (ValidatorRisk, bool) {
	validator = strings.TrimSpace(validator)
	history := s.QueryValidatorSlashHistory(validator)
	if len(history) == 0 && latestScoreRecord(s.ScoreRecords, validator).ValidatorAddress == "" {
		return ValidatorRisk{}, false
	}
	totalSlashed := sdkmath.ZeroInt()
	for _, record := range history {
		totalSlashed = totalSlashed.Add(record.TotalSlashedNaet)
	}
	latestScore := latestScoreRecord(s.ScoreRecords, validator)
	reliability := latestScore.ReliabilityIndex
	if reliability == 0 {
		reliability = postypes.BasisPoints
	}
	risk := ValidatorRisk{
		Validator:		validator,
		SlashEventCount:	uint32(len(history)),
		TotalSlashedNaet:	totalSlashed,
		LatestReliabilityBps:	reliability,
		DelegatorRiskInherited:	len(history) > 0,
	}
	risk.RiskScoreBps = minBps(uint64(len(history))*1_000 + uint64(postypes.BasisPoints-reliability))
	return risk, true
}

func (s ValidatorMarketState) QueryValidatorRiskComponents(validator string, decParams DecentralizationParams, activeValidatorIDs []string) (ValidatorRiskScoreComponents, bool, error) {
	validator = strings.TrimSpace(validator)
	risk, found := s.QueryValidatorRisk(validator)
	if !found {
		return ValidatorRiskScoreComponents{}, false, nil
	}
	commission := uint32(0)
	if candidate, candidateFound := s.findCandidate(validator); candidateFound {
		commission = latestCommissionBps(s.CommissionHistory, validator, candidate.CommissionBps)
	}
	concentrationRisk := uint32(0)
	report, err := s.QueryConcentrationReport(decParams, activeValidatorIDs)
	if err != nil {
		return ValidatorRiskScoreComponents{}, false, err
	}
	for _, metric := range report.Metrics {
		if metric.Validator == validator {
			concentrationRisk = minBps(uint64(metric.RewardDampeningBps) + uint64(metric.StakeMovementIncentiveBps))
			break
		}
	}
	components := ValidatorRiskScoreComponents{
		Validator:		validator,
		SlashHistoryRiskBps:	minBps(uint64(risk.SlashEventCount) * 1_000),
		ReliabilityRiskBps:	postypes.BasisPoints - risk.LatestReliabilityBps,
		ConcentrationRiskBps:	concentrationRisk,
		CommissionRiskBps:	commission,
	}
	components.TotalRiskScoreBps = minBps(uint64(components.SlashHistoryRiskBps) + uint64(components.ReliabilityRiskBps) + uint64(components.ConcentrationRiskBps) + uint64(components.CommissionRiskBps))
	return components, true, nil
}

func (s ValidatorMarketState) QueryValidatorEffectiveYield(validator string, annualRewardsNaet sdkmath.Int) (ValidatorEffectiveYield, bool, error) {
	if annualRewardsNaet.IsNegative() {
		return ValidatorEffectiveYield{}, false, errors.New("annual rewards cannot be negative")
	}
	candidate, found := s.findCandidate(validator)
	if !found {
		return ValidatorEffectiveYield{}, false, nil
	}
	preview, err := postypes.PreviewStakeSaturation(s.Params, candidate)
	if err != nil {
		return ValidatorEffectiveYield{}, false, err
	}
	commission := latestCommissionBps(s.CommissionHistory, validator, candidate.CommissionBps)
	gross := shareBps(annualRewardsNaet, preview.BondedStakeNaet)
	net := uint32((uint64(gross) * uint64(postypes.BasisPoints-commission)) / uint64(postypes.BasisPoints))
	return ValidatorEffectiveYield{
		Validator:		strings.TrimSpace(validator),
		RawStakeNaet:		preview.BondedStakeNaet,
		RewardWeightNaet:	preview.RewardWeightNaet,
		GrossYieldBps:		gross,
		NetYieldBps:		net,
		CommissionBps:		commission,
		SaturationDampeningBps:	shareBps(preview.RewardWeightNaet, preview.BondedStakeNaet),
	}, true, nil
}

func (s ValidatorMarketState) QueryDelegationYieldEstimate(delegator string, validator string, delegationAmountNaet sdkmath.Int, annualRewardsNaet sdkmath.Int, decParams DecentralizationParams, activeValidatorIDs []string) (DelegationYieldEstimate, bool, error) {
	delegator = strings.TrimSpace(delegator)
	validator = strings.TrimSpace(validator)
	if err := validateEconomyToken("yield estimate delegator", delegator); err != nil {
		return DelegationYieldEstimate{}, false, err
	}
	candidate, found := s.findCandidate(validator)
	if !found {
		return DelegationYieldEstimate{}, false, nil
	}
	if annualRewardsNaet.IsNegative() {
		return DelegationYieldEstimate{}, false, errors.New("annual rewards cannot be negative")
	}
	amount := normalizeEconomyInt(delegationAmountNaet)
	if !amount.IsPositive() {
		if record, recordFound := s.findDelegation(delegator, validator); recordFound {
			amount = record.Amount
		}
	}
	if !amount.IsPositive() {
		return DelegationYieldEstimate{}, false, errors.New("delegation amount must be positive")
	}
	commission := latestCommissionBps(s.CommissionHistory, validator, candidate.CommissionBps)
	performance := validatorPerformanceAdjustmentBps(candidate, latestScoreRecord(s.ScoreRecords, validator))
	concentration := uint32(0)
	report, err := s.QueryConcentrationReport(decParams, activeValidatorIDs)
	if err != nil {
		return DelegationYieldEstimate{}, false, err
	}
	for _, metric := range report.Metrics {
		if metric.Validator == validator {
			concentration = metric.RewardDampeningBps
			break
		}
	}
	adjustedRewards := annualRewardsNaet.MulRaw(int64(performance)).QuoRaw(int64(postypes.BasisPoints))
	adjustedRewards = adjustedRewards.MulRaw(int64(postypes.BasisPoints - concentration)).QuoRaw(int64(postypes.BasisPoints))
	nominations := s.rewardNominationsForValidator(validator, delegator, amount)
	distribution, err := postypes.DistributeRewards(postypes.RewardInput{
		ValidatorID:		validator,
		TotalRewardsNaet:	adjustedRewards,
		CommissionBps:		commission,
		SelfStakeNaet:		candidate.SelfStakeNaet,
		Nominations:		nominations,
	})
	if err != nil {
		return DelegationYieldEstimate{}, false, err
	}
	estimatedReward := sdkmath.ZeroInt()
	for _, reward := range distribution.NominatorRewards {
		if reward.NominatorID == delegator {
			estimatedReward = reward.RewardNaet
			break
		}
	}
	totalStake := candidate.SelfStakeNaet.Add(sumNominationStake(nominations))
	return DelegationYieldEstimate{
		Validator:			validator,
		Delegator:			delegator,
		DelegationAmountNaet:		amount,
		RewardInputNaet:		annualRewardsNaet,
		AdjustedRewardInputNaet:	adjustedRewards,
		EstimatedRewardNaet:		estimatedReward,
		GrossYieldBps:			shareBps(annualRewardsNaet, totalStake),
		NetYieldBps:			shareBps(estimatedReward, amount),
		CommissionBps:			commission,
		PerformanceAdjustmentBps:	performance,
		ConcentrationAdjustmentBps:	concentration,
		ValidatorCommissionNaet:	distribution.ValidatorCommissionNaet,
		UsesDistributionInputs:		true,
	}, true, nil
}

func (s ValidatorMarketState) QueryValidatorSaturation(validator string) (postypes.StakeSaturationPreview, bool, error) {
	candidate, found := s.findCandidate(validator)
	if !found {
		return postypes.StakeSaturationPreview{}, false, nil
	}
	preview, err := postypes.PreviewStakeSaturation(s.Params, candidate)
	if err != nil {
		return postypes.StakeSaturationPreview{}, false, err
	}
	return preview, true, nil
}

func (s ValidatorMarketState) QueryDelegationRiskExposure(delegator string, validator string, slashFractionBps uint32, selfBondFirstLoss bool) (DelegationRiskExposure, bool, error) {
	delegator = strings.TrimSpace(delegator)
	validator = strings.TrimSpace(validator)
	record, found := s.findDelegation(delegator, validator)
	if !found {
		return DelegationRiskExposure{}, false, nil
	}
	candidate, candidateFound := s.findCandidate(validator)
	if !candidateFound {
		return DelegationRiskExposure{}, false, fmt.Errorf("validator %q is not in market candidates", validator)
	}
	propagation, err := PropagateSlash(SlashPropagationInput{
		Validator:		validator,
		SelfBondNaet:		candidate.SelfStakeNaet,
		Delegations:		s.Delegations,
		SlashFractionBps:	slashFractionBps,
		SelfBondFirstLoss:	selfBondFirstLoss,
	})
	if err != nil {
		return DelegationRiskExposure{}, false, err
	}
	projected := sdkmath.ZeroInt()
	firstLossProtected := sdkmath.ZeroInt()
	proportionalWithoutFirstLoss := mulIntBps(record.Amount, slashFractionBps)
	for _, slash := range propagation.DelegatorSlashes {
		if slash.Delegator == delegator && slash.Validator == validator {
			projected = slash.SlashedNaet
			break
		}
	}
	if proportionalWithoutFirstLoss.GT(projected) {
		firstLossProtected = proportionalWithoutFirstLoss.Sub(projected)
	}
	history := s.QueryValidatorSlashHistory(validator)
	historical := sdkmath.ZeroInt()
	for _, event := range history {
		historical = historical.Add(shareByStake(event.DelegatorSlashedNaet, record.Amount, s.totalDelegatedAtValidator(validator)))
	}
	return DelegationRiskExposure{
		Delegator:		delegator,
		Validator:		validator,
		Amount:			record.Amount,
		RiskAppetite:		record.RiskAppetite,
		AdvisoryRiskProfile:	true,
		SlashEventsInherited:	history,
		ProjectedSlashNaet:	projected,
		FirstLossProtectedNaet:	firstLossProtected,
		HistoricalSlashNaet:	historical,
	}, true, nil
}

func (s ValidatorMarketState) QueryDelegationActivationEpoch(delegator string, validator string) (uint64, bool) {
	record, found := s.findDelegation(strings.TrimSpace(delegator), strings.TrimSpace(validator))
	if !found {
		return 0, false
	}
	return record.ActivationEpoch, true
}

func (s ValidatorMarketState) QueryDelegationCommissionStatus(delegator string, validator string, height uint64, emitRedelegationAlert bool) (DelegationCommissionStatus, bool, error) {
	delegator = strings.TrimSpace(delegator)
	validator = strings.TrimSpace(validator)
	record, found := s.findDelegation(delegator, validator)
	if !found {
		return DelegationCommissionStatus{}, false, nil
	}
	candidate, candidateFound := s.findCandidate(validator)
	fallback := uint32(0)
	if candidateFound {
		fallback = candidate.CommissionBps
	}
	currentCommission := latestCommissionBps(s.CommissionHistory, validator, fallback)
	updated, alert, err := CheckCommissionTolerance(s.Params, record, currentCommission, height, emitRedelegationAlert)
	if err != nil {
		return DelegationCommissionStatus{}, false, err
	}
	return DelegationCommissionStatus{
		Delegator:		delegator,
		Validator:		validator,
		Status:			updated.Status,
		CommissionToleranceBps:	record.CommissionTolerance,
		CurrentCommissionBps:	currentCommission,
		CommissionExceeded:	updated.Status == DelegationStatusCommissionExceeded,
		Alert:			alert,
	}, true, nil
}

func (s ValidatorMarketState) QueryDelegationLockEligibility(delegator string, validator string, slashableWindowEpochs uint64) (LockDurationRewardEligibility, bool, error) {
	record, found := s.findDelegation(strings.TrimSpace(delegator), strings.TrimSpace(validator))
	if !found {
		return LockDurationRewardEligibility{}, false, nil
	}
	eligibility, err := EvaluateLockDurationPreference(s.Params, record, slashableWindowEpochs)
	if err != nil {
		return LockDurationRewardEligibility{}, false, err
	}
	return eligibility, true, nil
}

func (s ValidatorMarketState) QueryDelegatorValidatorProfile(delegator string, validator string, delegationAmountNaet sdkmath.Int, annualRewardsNaet sdkmath.Int, decParams DecentralizationParams, activeValidatorIDs []string) (DelegatorValidatorProfile, bool, error) {
	risk, found := s.QueryValidatorRisk(validator)
	if !found {
		return DelegatorValidatorProfile{}, false, nil
	}
	components, found, err := s.QueryValidatorRiskComponents(validator, decParams, activeValidatorIDs)
	if err != nil || !found {
		return DelegatorValidatorProfile{}, found, err
	}
	yield, found, err := s.QueryDelegationYieldEstimate(delegator, validator, delegationAmountNaet, annualRewardsNaet, decParams, activeValidatorIDs)
	if err != nil || !found {
		return DelegatorValidatorProfile{}, found, err
	}
	disclosure, found, err := s.QueryValidatorDisclosure(validator, decParams, activeValidatorIDs)
	if err != nil || !found {
		return DelegatorValidatorProfile{}, found, err
	}
	evaluations := EvaluateDelegationPolicyTemplates(DefaultDelegationPolicyTemplates(s.Params), components, yield, disclosure)
	return DelegatorValidatorProfile{
		Validator:		strings.TrimSpace(validator),
		Risk:			risk,
		RiskComponents:		components,
		YieldEstimate:		yield,
		ConcentrationWarnings:	disclosure.ConcentrationWarnings,
		Disclosure:		disclosure,
		PolicyEvaluations:	evaluations,
		AdvisoryOnly:		true,
	}, true, nil
}

func (s ValidatorMarketState) QueryValidatorDisclosure(validator string, decParams DecentralizationParams, activeValidatorIDs []string) (ValidatorDisclosure, bool, error) {
	validator = strings.TrimSpace(validator)
	candidate, found := s.findCandidate(validator)
	if !found {
		return ValidatorDisclosure{}, false, nil
	}
	risk, _ := s.QueryValidatorRisk(validator)
	report, err := s.QueryConcentrationReport(decParams, activeValidatorIDs)
	if err != nil {
		return ValidatorDisclosure{}, false, err
	}
	warnings := []string(nil)
	status := ConcentrationStatusNormal
	for _, metric := range report.Metrics {
		if metric.Validator == validator {
			warnings = append(warnings, metric.Warnings...)
			if len(metric.Warnings) > 0 {
				status = ConcentrationStatusWarning
			}
			break
		}
	}
	commission := latestCommissionBps(s.CommissionHistory, validator, candidate.CommissionBps)
	uptime := candidate.UptimeFactorBps
	if latest := latestScoreRecord(s.ScoreRecords, validator); latest.ValidatorAddress != "" {
		uptime = latest.UptimeFactor
	}
	return ValidatorDisclosure{
		Validator:		validator,
		CommissionBps:		commission,
		MaxCommissionChangeBps:	s.Params.MaxCommissionBps - commission,
		UptimeBps:		normalizeOptionalBps(uptime),
		SlashHistoryCount:	risk.SlashEventCount,
		TotalSlashedNaet:	risk.TotalSlashedNaet,
		SelfDelegationNaet:	candidate.SelfStakeNaet,
		ConcentrationStatus:	status,
		ConcentrationWarnings:	warnings,
	}, true, nil
}

func (s ValidatorMarketState) QueryRedelegationRewardPreview(delegator string, fromValidator string, toValidator string, amount sdkmath.Int, annualRewardsNaet sdkmath.Int, decParams DecentralizationParams, activeValidatorIDs []string) (RedelegationRewardPreview, bool, error) {
	delegator = strings.TrimSpace(delegator)
	fromValidator = strings.TrimSpace(fromValidator)
	toValidator = strings.TrimSpace(toValidator)
	if err := validateEconomyToken("redelegation delegator", delegator); err != nil {
		return RedelegationRewardPreview{}, false, err
	}
	if err := validateEconomyToken("redelegation source validator", fromValidator); err != nil {
		return RedelegationRewardPreview{}, false, err
	}
	if err := validateEconomyToken("redelegation target validator", toValidator); err != nil {
		return RedelegationRewardPreview{}, false, err
	}
	if !normalizeEconomyInt(amount).IsPositive() {
		return RedelegationRewardPreview{}, false, errors.New("redelegation preview amount must be positive")
	}
	if _, found := s.findCandidate(fromValidator); !found {
		return RedelegationRewardPreview{}, false, nil
	}
	if _, found := s.findCandidate(toValidator); !found {
		return RedelegationRewardPreview{}, false, nil
	}
	current, found, err := s.QueryDelegationYieldEstimate(delegator, fromValidator, amount, annualRewardsNaet, decParams, activeValidatorIDs)
	if err != nil || !found {
		return RedelegationRewardPreview{}, found, err
	}
	target, found, err := s.QueryDelegationYieldEstimate(delegator, toValidator, amount, annualRewardsNaet, decParams, activeValidatorIDs)
	if err != nil || !found {
		return RedelegationRewardPreview{}, found, err
	}
	targetRisk, _ := s.QueryValidatorRisk(toValidator)
	targetComponents, found, err := s.QueryValidatorRiskComponents(toValidator, decParams, activeValidatorIDs)
	if err != nil || !found {
		return RedelegationRewardPreview{}, found, err
	}
	targetDisclosure, found, err := s.QueryValidatorDisclosure(toValidator, decParams, activeValidatorIDs)
	if err != nil || !found {
		return RedelegationRewardPreview{}, found, err
	}
	return RedelegationRewardPreview{
		Delegator:		delegator,
		FromValidator:		fromValidator,
		ToValidator:		toValidator,
		AmountNaet:		amount,
		CurrentEstimate:	current,
		TargetEstimate:		target,
		RewardDeltaNaet:	target.EstimatedRewardNaet.Sub(current.EstimatedRewardNaet),
		NetYieldDeltaBps:	int32(target.NetYieldBps) - int32(current.NetYieldBps),
		TargetRisk:		targetRisk,
		TargetRiskComponents:	targetComponents,
		TargetDisclosure:	targetDisclosure,
		PolicyEvaluations:	EvaluateDelegationPolicyTemplates(DefaultDelegationPolicyTemplates(s.Params), targetComponents, target, targetDisclosure),
		AdvisoryOnly:		true,
		StakeMovementExecuted:	false,
	}, true, nil
}

func (s ValidatorMarketState) QueryRiskAdjustedYieldProjection(input RiskAdjustedYieldInput) (RiskAdjustedYieldProjection, bool, error) {
	if input.UnbondingLiquidityCostBps > postypes.BasisPoints {
		return RiskAdjustedYieldProjection{}, false, fmt.Errorf("unbonding liquidity cost must be <= %d bps", postypes.BasisPoints)
	}
	yield, found, err := s.QueryDelegationYieldEstimate(input.Delegator, input.Validator, input.AmountNaet, input.AnnualRewardsNaet, input.Decentralization, input.ActiveValidatorIDs)
	if err != nil || !found {
		return RiskAdjustedYieldProjection{}, found, err
	}
	components, found, err := s.QueryValidatorRiskComponents(input.Validator, input.Decentralization, input.ActiveValidatorIDs)
	if err != nil || !found {
		return RiskAdjustedYieldProjection{}, found, err
	}
	disclosure, found, err := s.QueryValidatorDisclosure(input.Validator, input.Decentralization, input.ActiveValidatorIDs)
	if err != nil || !found {
		return RiskAdjustedYieldProjection{}, found, err
	}
	slashProxy := minBps(uint64(components.SlashHistoryRiskBps) + uint64(components.ReliabilityRiskBps)/2)
	riskYield := uint64(yield.NetYieldBps)
	riskYield = riskYield * uint64(disclosure.UptimeBps) / uint64(postypes.BasisPoints)
	riskYield = riskYield * uint64(postypes.BasisPoints-slashProxy) / uint64(postypes.BasisPoints)
	riskYield = riskYield * uint64(postypes.BasisPoints-yield.ConcentrationAdjustmentBps) / uint64(postypes.BasisPoints)
	riskYield = riskYield * uint64(postypes.BasisPoints-input.UnbondingLiquidityCostBps) / uint64(postypes.BasisPoints)
	riskAdjusted := minBps(riskYield)
	amount := normalizeEconomyInt(input.AmountNaet)
	if !amount.IsPositive() {
		amount = yield.DelegationAmountNaet
	}
	expected := mulIntBps(amount, riskAdjusted)
	uncertainty := minBps(uint64(slashProxy) + uint64(yield.ConcentrationAdjustmentBps) + uint64(input.UnbondingLiquidityCostBps))
	low := expected.MulRaw(int64(postypes.BasisPoints - uncertainty)).QuoRaw(int64(postypes.BasisPoints))
	highBand := uncertainty
	if highBand > 5_000 {
		highBand = 5_000
	}
	high := expected.MulRaw(int64(postypes.BasisPoints + highBand)).QuoRaw(int64(postypes.BasisPoints))
	return RiskAdjustedYieldProjection{
		Validator:			yield.Validator,
		Delegator:			yield.Delegator,
		GrossRewardRateBps:		yield.GrossYieldBps,
		CommissionBps:			yield.CommissionBps,
		HistoricalUptimeBps:		disclosure.UptimeBps,
		SlashProbabilityProxyBps:	slashProxy,
		ConcentrationAdjustmentBps:	yield.ConcentrationAdjustmentBps,
		UnbondingLiquidityCostBps:	input.UnbondingLiquidityCostBps,
		RiskAdjustedYieldBps:		riskAdjusted,
		ExpectedRewardNaet:		expected,
		LowRewardNaet:			low,
		HighRewardNaet:			high,
		UncertaintyBps:			uncertainty,
		ReproducibleFromQueries:	true,
	}, true, nil
}

func ComputeValidatorRewardVariance(validator string, observations []ValidatorRewardObservation) (ValidatorRewardVariance, bool, error) {
	validator = strings.TrimSpace(validator)
	if err := validateEconomyToken("reward variance validator", validator); err != nil {
		return ValidatorRewardVariance{}, false, err
	}
	values := make([]uint32, 0, len(observations))
	for _, observation := range observations {
		if observation.EpochID == 0 {
			return ValidatorRewardVariance{}, false, errors.New("reward observation epoch id is required")
		}
		if observation.RewardPerStakeBps > postypes.BasisPoints {
			return ValidatorRewardVariance{}, false, fmt.Errorf("reward observation bps must be <= %d", postypes.BasisPoints)
		}
		if strings.TrimSpace(observation.Validator) == validator {
			values = append(values, observation.RewardPerStakeBps)
		}
	}
	if len(values) == 0 {
		return ValidatorRewardVariance{}, false, nil
	}
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
	sum := uint64(0)
	for _, value := range values {
		sum += uint64(value)
	}
	mean := uint32(sum / uint64(len(values)))
	deviation := uint64(0)
	for _, value := range values {
		if value > mean {
			deviation += uint64(value - mean)
		} else {
			deviation += uint64(mean - value)
		}
	}
	return ValidatorRewardVariance{
		Validator:		validator,
		ObservationCount:	uint32(len(values)),
		MeanRewardBps:		mean,
		MinRewardBps:		values[0],
		MaxRewardBps:		values[len(values)-1],
		VarianceBps:		uint32(deviation / uint64(len(values))),
	}, true, nil
}

func (s ValidatorMarketState) SimulateDelegation(input DelegationSimulationInput) (DelegationSimulationResult, bool, error) {
	amount := normalizeEconomyInt(input.AmountNaet)
	if !amount.IsPositive() {
		return DelegationSimulationResult{}, false, errors.New("delegation simulation amount must be positive")
	}
	simulated := s
	if len(input.CommissionOverrides) > 0 {
		simulated.CommissionHistory = append([]ValidatorCommissionRecord{}, s.CommissionHistory...)
		simulated.CommissionHistory = append(simulated.CommissionHistory, input.CommissionOverrides...)
		sortCommissionHistory(simulated.CommissionHistory)
	}
	if len(input.SlashEvents) > 0 {
		simulated.SlashHistory = append([]ValidatorSlashHistoryRecord{}, s.SlashHistory...)
		simulated.SlashHistory = append(simulated.SlashHistory, input.SlashEvents...)
		sortSlashHistory(simulated.SlashHistory)
	}
	current, found, err := simulated.QueryRiskAdjustedYieldProjection(RiskAdjustedYieldInput{
		Delegator:			input.Delegator,
		Validator:			input.FromValidator,
		AmountNaet:			amount,
		AnnualRewardsNaet:		input.AnnualRewardsNaet,
		UnbondingLiquidityCostBps:	input.UnbondingLiquidityCostBps,
		Decentralization:		input.Decentralization,
		ActiveValidatorIDs:		input.ActiveValidatorIDs,
	})
	if err != nil || !found {
		return DelegationSimulationResult{}, found, err
	}
	target, found, err := simulated.QueryRiskAdjustedYieldProjection(RiskAdjustedYieldInput{
		Delegator:			input.Delegator,
		Validator:			input.ToValidator,
		AmountNaet:			amount,
		AnnualRewardsNaet:		input.AnnualRewardsNaet,
		UnbondingLiquidityCostBps:	input.UnbondingLiquidityCostBps,
		Decentralization:		input.Decentralization,
		ActiveValidatorIDs:		input.ActiveValidatorIDs,
	})
	if err != nil || !found {
		return DelegationSimulationResult{}, found, err
	}
	preview, found, err := simulated.QueryRedelegationRewardPreview(input.Delegator, input.FromValidator, input.ToValidator, amount, input.AnnualRewardsNaet, input.Decentralization, input.ActiveValidatorIDs)
	if err != nil || !found {
		return DelegationSimulationResult{}, found, err
	}
	cost := simulated.estimateRedelegationCost(amount, current, target, input.UnbondingLiquidityCostBps)
	return DelegationSimulationResult{
		CurrentProjection:		current,
		TargetProjection:		target,
		RedelegationPreview:		preview,
		RedelegationCost:		cost,
		ProjectedRewardDeltaNaet:	target.ExpectedRewardNaet.Sub(current.ExpectedRewardNaet),
		AdvisoryOnly:			true,
	}, true, nil
}

func (s ValidatorMarketState) QueryConcentrationReport(params DecentralizationParams, activeValidatorIDs []string) (ActiveSetConcentrationReport, error) {
	if params.MinSelfDelegationNaet.IsNil() {
		params = DefaultDecentralizationParams(s.Params)
	}
	if err := params.Validate(s.Params); err != nil {
		return ActiveSetConcentrationReport{}, err
	}
	activeSet := normalizeActiveSet(activeValidatorIDs)
	candidates := make([]postypes.Candidate, 0, len(s.Candidates))
	for _, candidate := range s.Candidates {
		validatorID := strings.TrimSpace(candidate.ValidatorID)
		if len(activeSet) == 0 {
			candidates = append(candidates, candidate)
			continue
		}
		if _, active := activeSet[validatorID]; active {
			candidates = append(candidates, candidate)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i].SelfStakeNaet.Add(candidates[i].DelegatedStakeNaet)
		right := candidates[j].SelfStakeNaet.Add(candidates[j].DelegatedStakeNaet)
		if !left.Equal(right) {
			return left.GT(right)
		}
		return strings.TrimSpace(candidates[i].ValidatorID) < strings.TrimSpace(candidates[j].ValidatorID)
	})
	totalStake := sdkmath.ZeroInt()
	for _, candidate := range candidates {
		totalStake = totalStake.Add(candidate.SelfStakeNaet).Add(candidate.DelegatedStakeNaet)
	}
	report := ActiveSetConcentrationReport{
		Metrics: make([]ValidatorConcentrationMetric, 0, len(candidates)),
	}
	topLimit := params.TopN
	if topLimit > len(candidates) {
		topLimit = len(candidates)
	}
	topStake := sdkmath.ZeroInt()
	commissionWeighted := uint64(0)
	for i, candidate := range candidates {
		validatorID := strings.TrimSpace(candidate.ValidatorID)
		totalForValidator := candidate.SelfStakeNaet.Add(candidate.DelegatedStakeNaet)
		share := shareBps(totalForValidator, totalStake)
		if i < topLimit {
			topStake = topStake.Add(totalForValidator)
		}
		commission := latestCommissionBps(s.CommissionHistory, validatorID, candidate.CommissionBps)
		commissionWeighted += uint64(share) * uint64(commission)
		metric := ValidatorConcentrationMetric{
			Validator:			validatorID,
			VotingPowerShareBps:		share,
			DelegatorConcentrationBps:	s.delegatorConcentrationBps(validatorID),
			SelfDelegationRatioBps:		shareBps(candidate.SelfStakeNaet, totalForValidator),
			CommissionBps:			commission,
		}
		metric.RewardDampeningBps = concentrationRewardDampeningBps(share, params.MaxValidatorShareBps, params.RewardDampeningSafetyFloorBps)
		metric.Warnings = concentrationWarnings(params, candidate, metric)
		if hasConcentrationStakeMovementWarning(metric.Warnings) {
			metric.StakeMovementIncentiveBps = params.StakeMovementIncentiveBps
		}
		report.Metrics = append(report.Metrics, metric)
	}
	report.TopNShareBps = shareBps(topStake, totalStake)
	report.CommissionWeightedBps = uint32(commissionWeighted / uint64(postypes.BasisPoints))
	if report.TopNShareBps > params.MaxTopNShareBps {
		report.ActiveSetWarnings = append(report.ActiveSetWarnings, ConcentrationWarningTopNShare)
	}
	if report.CommissionWeightedBps > params.MaxCommissionWeightedBps {
		report.ActiveSetWarnings = append(report.ActiveSetWarnings, ConcentrationWarningCommissionByPower)
	}
	sort.SliceStable(report.Metrics, func(i, j int) bool {
		if report.Metrics[i].VotingPowerShareBps != report.Metrics[j].VotingPowerShareBps {
			return report.Metrics[i].VotingPowerShareBps > report.Metrics[j].VotingPowerShareBps
		}
		return report.Metrics[i].Validator < report.Metrics[j].Validator
	})
	return report, nil
}

func (s ValidatorMarketState) QueryValidatorCommissionHistory(validator string) []ValidatorCommissionRecord {
	validator = strings.TrimSpace(validator)
	out := make([]ValidatorCommissionRecord, 0)
	for _, record := range s.CommissionHistory {
		if record.Validator == validator {
			out = append(out, record)
		}
	}
	sortCommissionHistory(out)
	return out
}

func (s ValidatorMarketState) QueryValidatorSlashHistory(validator string) []ValidatorSlashHistoryRecord {
	validator = strings.TrimSpace(validator)
	out := make([]ValidatorSlashHistoryRecord, 0)
	for _, record := range s.SlashHistory {
		if record.Validator == validator {
			out = append(out, record)
		}
	}
	sortSlashHistory(out)
	return out
}

func (s ValidatorMarketState) QueryValidatorPerformanceHistory(validator string) []ValidatorScoreRecord {
	validator = strings.TrimSpace(validator)
	out := make([]ValidatorScoreRecord, 0)
	for _, record := range s.ScoreRecords {
		if record.ValidatorAddress == validator {
			out = append(out, record)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].EpochID < out[j].EpochID
	})
	return out
}

func (s ValidatorMarketState) findCandidate(validator string) (postypes.Candidate, bool) {
	validator = strings.TrimSpace(validator)
	for _, candidate := range s.Candidates {
		if strings.TrimSpace(candidate.ValidatorID) == validator {
			return candidate, true
		}
	}
	return postypes.Candidate{}, false
}

func (s ValidatorMarketState) findDelegation(delegator string, validator string) (DelegationRecord, bool) {
	for _, record := range s.Delegations {
		if record.Delegator == delegator && record.Validator == validator {
			return record, true
		}
	}
	return DelegationRecord{}, false
}

func (s ValidatorMarketState) totalDelegatedAtValidator(validator string) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, record := range s.Delegations {
		if record.Validator == validator {
			total = total.Add(record.Amount)
		}
	}
	return total
}

func (s ValidatorMarketState) delegatorConcentrationBps(validator string) uint32 {
	total := sdkmath.ZeroInt()
	maxDelegation := sdkmath.ZeroInt()
	for _, record := range s.Delegations {
		if record.Validator != validator {
			continue
		}
		total = total.Add(record.Amount)
		if record.Amount.GT(maxDelegation) {
			maxDelegation = record.Amount
		}
	}
	return shareBps(maxDelegation, total)
}

func (s ValidatorMarketState) estimateRedelegationCost(amount sdkmath.Int, current RiskAdjustedYieldProjection, target RiskAdjustedYieldProjection, liquidityCostBps uint32) RedelegationCostEstimate {
	commissionDelta := int32(target.CommissionBps) - int32(current.CommissionBps)
	riskDelta := int32(target.SlashProbabilityProxyBps+target.ConcentrationAdjustmentBps) - int32(current.SlashProbabilityProxyBps+current.ConcentrationAdjustmentBps)
	totalCostBps := uint32(0)
	if commissionDelta > 0 {
		totalCostBps += uint32(commissionDelta)
	}
	if riskDelta > 0 {
		totalCostBps += uint32(riskDelta)
	}
	totalCostBps = minBps(uint64(totalCostBps) + uint64(liquidityCostBps))
	return RedelegationCostEstimate{
		AmountNaet:		amount,
		UnbondingSeconds:	s.Params.UnbondingSeconds,
		LiquidityCostBps:	liquidityCostBps,
		CommissionDeltaBps:	commissionDelta,
		RiskDeltaBps:		riskDelta,
		EstimatedCostNaet:	mulIntBps(amount, totalCostBps),
	}
}

func DefaultDelegationPolicyTemplates(params postypes.Params) []DelegationPolicyTemplate {
	maxConcentration := params.MaxVotingPowerBps
	if maxConcentration == 0 {
		maxConcentration = postypes.DefaultMaxVotingPowerBps
	}
	return []DelegationPolicyTemplate{
		{
			Name:			DelegationPolicyLowRisk,
			RiskAppetite:		RiskAppetiteConservative,
			MaxRiskScoreBps:	2_000,
			MinUptimeBps:		9_800,
			MaxCommissionBps:	minUint32(params.MaxCommissionBps, 1_000),
			MaxValidatorShareBps:	maxConcentration,
			RequireNoSlashHistory:	false,
			AdvisoryOnly:		true,
		},
		{
			Name:			DelegationPolicyHighAvailability,
			RiskAppetite:		RiskAppetiteBalanced,
			MaxRiskScoreBps:	3_000,
			MinUptimeBps:		9_900,
			MaxCommissionBps:	params.MaxCommissionBps,
			MaxValidatorShareBps:	postypes.BasisPoints,
			AdvisoryOnly:		true,
		},
		{
			Name:				DelegationPolicyDecentralizationSupporting,
			RiskAppetite:			RiskAppetiteBalanced,
			MaxRiskScoreBps:		5_000,
			MinUptimeBps:			params.MinUptimeBps,
			MaxCommissionBps:		params.MaxCommissionBps,
			MaxValidatorShareBps:		maxConcentration,
			PreferConcentrationRelief:	true,
			AdvisoryOnly:			true,
		},
		{
			Name:			DelegationPolicyMaxYieldWithinRiskBounds,
			RiskAppetite:		RiskAppetiteAggressive,
			MaxRiskScoreBps:	5_000,
			MinUptimeBps:		params.MinUptimeBps,
			MaxCommissionBps:	params.MaxCommissionBps,
			MaxValidatorShareBps:	postypes.BasisPoints,
			AdvisoryOnly:		true,
		},
	}
}

func EvaluateDelegationPolicyTemplates(templates []DelegationPolicyTemplate, risk ValidatorRiskScoreComponents, yield DelegationYieldEstimate, disclosure ValidatorDisclosure) []DelegationPolicyEvaluation {
	out := make([]DelegationPolicyEvaluation, 0, len(templates))
	for _, template := range templates {
		reasons := make([]string, 0, 4)
		if risk.TotalRiskScoreBps > template.MaxRiskScoreBps {
			reasons = append(reasons, "risk_score_above_policy_bound")
		}
		if disclosure.UptimeBps < template.MinUptimeBps {
			reasons = append(reasons, "uptime_below_policy_minimum")
		}
		if disclosure.CommissionBps > template.MaxCommissionBps {
			reasons = append(reasons, "commission_above_policy_bound")
		}
		if template.RequireNoSlashHistory && disclosure.SlashHistoryCount > 0 {
			reasons = append(reasons, "slash_history_not_empty")
		}
		if template.MaxValidatorShareBps < postypes.BasisPoints && hasWarning(disclosure.ConcentrationWarnings, ConcentrationWarningValidatorShare) {
			reasons = append(reasons, "validator_concentration_above_policy_bound")
		}
		if template.PreferConcentrationRelief && disclosure.ConcentrationStatus == ConcentrationStatusWarning {
			reasons = append(reasons, "target_does_not_relieve_concentration")
		}
		if yield.NetYieldBps == 0 {
			reasons = append(reasons, "net_yield_unavailable")
		}
		out = append(out, DelegationPolicyEvaluation{
			PolicyName:	template.Name,
			Matches:	len(reasons) == 0,
			Reasons:	reasons,
			AdvisoryOnly:	true,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].PolicyName < out[j].PolicyName
	})
	return out
}

func (s ValidatorMarketState) rewardNominationsForValidator(validator string, delegator string, amount sdkmath.Int) []postypes.Nomination {
	validator = strings.TrimSpace(validator)
	delegator = strings.TrimSpace(delegator)
	amount = normalizeEconomyInt(amount)
	byDelegator := make(map[string]sdkmath.Int)
	for _, record := range s.Delegations {
		if record.Validator != validator {
			continue
		}
		current := byDelegator[record.Delegator]
		if current.IsNil() {
			current = sdkmath.ZeroInt()
		}
		byDelegator[record.Delegator] = current.Add(record.Amount)
	}
	if delegator != "" && amount.IsPositive() {
		byDelegator[delegator] = amount
	}
	out := make([]postypes.Nomination, 0, len(byDelegator))
	for nominator, stake := range byDelegator {
		if stake.IsPositive() {
			out = append(out, postypes.Nomination{NominatorID: nominator, StakeNaet: stake})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].NominatorID < out[j].NominatorID
	})
	return out
}

func sumNominationStake(nominations []postypes.Nomination) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, nomination := range nominations {
		total = total.Add(nomination.StakeNaet)
	}
	return total
}

func validatorPerformanceAdjustmentBps(candidate postypes.Candidate, latest ValidatorScoreRecord) uint32 {
	performance := normalizeOptionalBps(candidate.PerformanceScoreBps)
	uptime := normalizeOptionalBps(candidate.UptimeFactorBps)
	reliability := normalizeOptionalBps(candidate.ReliabilityIndexBps)
	if latest.ValidatorAddress != "" {
		performance = normalizeOptionalBps(latest.PerformanceFactor)
		uptime = normalizeOptionalBps(latest.UptimeFactor)
		reliability = normalizeOptionalBps(latest.ReliabilityIndex)
	}
	return minUint32(performance, minUint32(uptime, reliability))
}

func hasWarning(warnings []string, target string) bool {
	for _, warning := range warnings {
		if warning == target {
			return true
		}
	}
	return false
}

func metadataChangedFields(previous ValidatorMetadataSnapshot, current ValidatorMetadataSnapshot) []string {
	previous = normalizeMetadataSnapshot(previous)
	current = normalizeMetadataSnapshot(current)
	fields := make([]string, 0, 6)
	if previous.OperatorID != current.OperatorID {
		fields = append(fields, "operator_id")
	}
	if previous.ConsensusKeyID != current.ConsensusKeyID {
		fields = append(fields, "consensus_key_id")
	}
	if previous.Moniker != current.Moniker {
		fields = append(fields, "moniker")
	}
	if previous.Website != current.Website {
		fields = append(fields, "website")
	}
	if previous.SecurityContact != current.SecurityContact {
		fields = append(fields, "security_contact")
	}
	if previous.PayoutAddress != current.PayoutAddress {
		fields = append(fields, "payout_address")
	}
	sort.Strings(fields)
	return fields
}

func normalizeMetadataSnapshot(snapshot ValidatorMetadataSnapshot) ValidatorMetadataSnapshot {
	return ValidatorMetadataSnapshot{
		OperatorID:		strings.TrimSpace(snapshot.OperatorID),
		ConsensusKeyID:		strings.TrimSpace(snapshot.ConsensusKeyID),
		Moniker:		strings.TrimSpace(snapshot.Moniker),
		Website:		strings.TrimSpace(snapshot.Website),
		SecurityContact:	strings.TrimSpace(snapshot.SecurityContact),
		PayoutAddress:		strings.TrimSpace(snapshot.PayoutAddress),
	}
}

func metadataChangeMaterial(fields []string) bool {
	for _, field := range fields {
		switch field {
		case "operator_id", "consensus_key_id", "security_contact", "payout_address":
			return true
		}
	}
	return false
}

func rapidCommissionIncrease(records []ValidatorCommissionRecord, validator string, currentEpoch uint64, intervalEpochs uint64, maxIncreaseBps uint32) (bool, uint32) {
	if currentEpoch == 0 || intervalEpochs == 0 {
		return false, 0
	}
	validator = strings.TrimSpace(validator)
	windowStart := uint64(1)
	if currentEpoch > intervalEpochs {
		windowStart = currentEpoch - intervalEpochs
	}
	minCommission := uint32(postypes.BasisPoints)
	maxCommission := uint32(0)
	found := false
	for _, record := range records {
		if record.Validator != validator || record.EpochID < windowStart || record.EpochID > currentEpoch {
			continue
		}
		found = true
		if record.CommissionBps < minCommission {
			minCommission = record.CommissionBps
		}
		if record.CommissionBps > maxCommission {
			maxCommission = record.CommissionBps
		}
	}
	if !found || maxCommission <= minCommission {
		return false, 0
	}
	increase := maxCommission - minCommission
	return increase > maxIncreaseBps, increase
}

func recentSlashCount(records []ValidatorSlashHistoryRecord, validator string, currentEpoch uint64, windowEpochs uint64) uint32 {
	if currentEpoch == 0 || windowEpochs == 0 {
		return 0
	}
	validator = strings.TrimSpace(validator)
	count := uint32(0)
	for _, record := range records {
		if record.Validator != validator || record.EpochID > currentEpoch {
			continue
		}
		if currentEpoch-record.EpochID <= windowEpochs {
			count++
		}
	}
	return count
}

func sortCaptureRiskIndicators(indicators []CaptureRiskIndicator) {
	sort.SliceStable(indicators, func(i, j int) bool {
		if indicators[i].SeverityBps != indicators[j].SeverityBps {
			return indicators[i].SeverityBps > indicators[j].SeverityBps
		}
		return indicators[i].Name < indicators[j].Name
	})
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func maxPositiveInt(a sdkmath.Int, b sdkmath.Int) sdkmath.Int {
	if a.IsPositive() && a.GTE(b) {
		return a
	}
	return b
}

func normalizeActiveSet(activeValidatorIDs []string) map[string]struct{} {
	out := make(map[string]struct{}, len(activeValidatorIDs))
	for _, validatorID := range activeValidatorIDs {
		validatorID = strings.TrimSpace(validatorID)
		if validatorID != "" {
			out[validatorID] = struct{}{}
		}
	}
	return out
}

func concentrationRewardDampeningBps(share uint32, threshold uint32, safetyFloor uint32) uint32 {
	if share <= threshold || threshold >= postypes.BasisPoints {
		return 0
	}
	maxDampening := postypes.BasisPoints - safetyFloor
	denom := postypes.BasisPoints - threshold
	if denom == 0 {
		return maxDampening
	}
	over := share - threshold
	dampening := uint32(uint64(over) * uint64(maxDampening) / uint64(denom))
	return minUint32(dampening, maxDampening)
}

func concentrationWarnings(params DecentralizationParams, candidate postypes.Candidate, metric ValidatorConcentrationMetric) []string {
	warnings := make([]string, 0, 4)
	if metric.VotingPowerShareBps > params.MaxValidatorShareBps {
		warnings = append(warnings, ConcentrationWarningValidatorShare)
	}
	if metric.DelegatorConcentrationBps > params.MaxDelegatorConcentrationBps {
		warnings = append(warnings, ConcentrationWarningDelegatorShare)
	}
	if metric.SelfDelegationRatioBps < params.MinSelfDelegationRatioBps || candidate.SelfStakeNaet.LT(params.MinSelfDelegationNaet) {
		warnings = append(warnings, ConcentrationWarningSelfDelegation)
	}
	if metric.RewardDampeningBps > 0 {
		warnings = append(warnings, ConcentrationWarningRewardDampeningActive)
	}
	return warnings
}

func hasConcentrationStakeMovementWarning(warnings []string) bool {
	for _, warning := range warnings {
		switch warning {
		case ConcentrationWarningValidatorShare, ConcentrationWarningDelegatorShare, ConcentrationWarningRewardDampeningActive:
			return true
		}
	}
	return false
}

func filterDelegationsForValidator(records []DelegationRecord, validator string) []DelegationRecord {
	validator = strings.TrimSpace(validator)
	out := make([]DelegationRecord, 0)
	for _, record := range records {
		if record.Validator == validator {
			out = append(out, record)
		}
	}
	sortDelegationRecords(out)
	return out
}

func latestScoreRecord(records []ValidatorScoreRecord, validator string) ValidatorScoreRecord {
	validator = strings.TrimSpace(validator)
	var latest ValidatorScoreRecord
	for _, record := range records {
		if record.ValidatorAddress == validator && record.EpochID >= latest.EpochID {
			latest = record
		}
	}
	return latest
}

func latestCommissionBps(records []ValidatorCommissionRecord, validator string, fallback uint32) uint32 {
	validator = strings.TrimSpace(validator)
	latestEpoch := uint64(0)
	commission := fallback
	for _, record := range records {
		if record.Validator == validator && record.EpochID >= latestEpoch {
			latestEpoch = record.EpochID
			commission = record.CommissionBps
		}
	}
	return commission
}

func sortSlashHistory(records []ValidatorSlashHistoryRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].EpochID != records[j].EpochID {
			return records[i].EpochID < records[j].EpochID
		}
		if records[i].Height != records[j].Height {
			return records[i].Height < records[j].Height
		}
		return records[i].Validator < records[j].Validator
	})
}

func sortCommissionHistory(records []ValidatorCommissionRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].EpochID != records[j].EpochID {
			return records[i].EpochID < records[j].EpochID
		}
		if records[i].Height != records[j].Height {
			return records[i].Height < records[j].Height
		}
		return records[i].Validator < records[j].Validator
	})
}

func mulIntBps(value sdkmath.Int, bps uint32) sdkmath.Int {
	return value.MulRaw(int64(bps)).QuoRaw(int64(postypes.BasisPoints))
}

func shareByStake(amount sdkmath.Int, stake sdkmath.Int, totalStake sdkmath.Int) sdkmath.Int {
	if !amount.IsPositive() || !stake.IsPositive() || !totalStake.IsPositive() {
		return sdkmath.ZeroInt()
	}
	return amount.Mul(stake).Quo(totalStake)
}

func minBps(value uint64) uint32 {
	if value > uint64(postypes.BasisPoints) {
		return postypes.BasisPoints
	}
	return uint32(value)
}
