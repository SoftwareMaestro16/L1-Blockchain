package params

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
)

const (
	EconomicSecurityAlertReserveLow		= "security_reserve_below_minimum"
	EconomicSecurityAlertConcentration	= "validator_concentration_above_threshold"
	EconomicSecurityAlertTopNConcentration	= "top_n_concentration_above_threshold"
	EconomicSecurityAlertFeeInstability	= "fee_market_instability"
	EconomicSecurityAlertStateGrowth	= "state_growth_abnormal"
	EconomicSecurityAlertStakeMovement	= "stake_movement_abnormal"
	EconomicSecurityAlertEvidenceDuplicate	= "duplicate_evidence_rejected"
	EconomicSecurityAlertCircuitBreaker	= "economic_circuit_breaker_active"

	EconomicSecurityEventInvariantViolation	= "economic_security_invariant_violation"
	EconomicSecurityEventReserveFunding	= "economic_security_reserve_funding_request"
	EconomicSecurityEventPenaltyRouting	= "economic_security_penalty_routing"
	EconomicSecurityEventCircuitBreaker	= "economic_security_circuit_breaker"
	EconomicSecurityEventAudit		= "economic_security_audit"

	EconomicSecuritySeverityInfo		= "info"
	EconomicSecuritySeverityWarning		= "warning"
	EconomicSecuritySeverityCritical	= "critical"

	DefaultSecurityReserveMinimumNaet	= int64(1_000_000)
	DefaultSecurityReserveTargetNaet	= int64(5_000_000)
	DefaultSecurityReserveMaxRequestNaet	= int64(10_000_000)
	DefaultSecuritySlashingBurnRatioBps	= int64(2_000)
	DefaultSecuritySlashingTreasuryBps	= int64(3_000)
	DefaultSecurityReporterRewardCapNaet	= int64(1_000)
	DefaultSecurityGovernanceVersion	= "economic-security/v1"
)

type EconomicSecurityParams struct {
	MaxValidatorConcentrationBps	int64
	MaxTopNConcentrationBps		int64
	MaxStateGrowthBytes		int64
	MaxStakeMovementBps		int64
	SecurityReserveMinimumNaet	sdkmath.Int
	SecurityReserveTargetNaet	sdkmath.Int
	MaxReserveFundingRequestNaet	sdkmath.Int
	SlashingBurnRatioBps		int64
	SlashingTreasuryRatioBps	int64
	ReporterRewardCapNaet		sdkmath.Int
	GovernanceThresholdVersion	string
	CircuitBreakerParams		EconomicCircuitBreakerParams
}

type EconomicSecurityEpochInput struct {
	EpochID				uint64
	BlockHeight			uint64
	SlashingPenaltyNaet		sdkmath.Int
	EvidenceAccepted		bool
	EvidenceDuplicate		bool
	RequestedReporterRewardNaet	sdkmath.Int
	ValidatorConcentrationBps	int64
	TopNConcentrationBps		int64
	BlockLoadBps			int64
	FeeSpikeBps			int64
	ControllerDriftBps		int64
	FailedTxRateBps			int64
	BurnToMintBps			int64
	StateGrowthBytes		int64
	StakeInflowBps			int64
	StakeOutflowBps			int64
	SecurityReserveBalanceNaet	sdkmath.Int
	SecurityReserveInflowNaet	sdkmath.Int
	SecurityReserveOutflowNaet	sdkmath.Int
	GovernanceThresholdVersion	string
}

type EconomicSecurityAlert struct {
	Type		string
	Severity	string
	ObservedBps	int64
	ThresholdBps	int64
	ObservedBytes	int64
	ThresholdBytes	int64
	AmountNaet	sdkmath.Int
	Message		string
}

type PenaltyRoutingDecision struct {
	PenaltyNaet		sdkmath.Int
	BurnNaet		sdkmath.Int
	TreasuryNaet		sdkmath.Int
	ReporterRewardNaet	sdkmath.Int
	ValidatorPoolNaet	sdkmath.Int
	EvidenceAccepted	bool
	EvidenceDuplicate	bool
	Routed			bool
}

type SecurityReserveAccounting struct {
	StartingBalanceNaet	sdkmath.Int
	InflowNaet		sdkmath.Int
	OutflowNaet		sdkmath.Int
	EndingBalanceNaet	sdkmath.Int
	MinimumBalanceNaet	sdkmath.Int
	TargetBalanceNaet	sdkmath.Int
	FundingRequestNaet	sdkmath.Int
	Consistent		bool
}

type EconomicSecurityIncidentEvent struct {
	Type		string
	EpochID		uint64
	BlockHeight	uint64
	Severity	string
	Reason		string
	AmountNaet	sdkmath.Int
	Reconciled	bool
}

type EconomicSecurityAuditLog struct {
	EpochID		uint64
	Action		string
	Deterministic	bool
	Details		string
}

type EconomicSecurityGovernanceReport struct {
	ThresholdVersion	string
	Summary			string
	ActiveRestrictions	[]string
	ReserveFundingRequest	sdkmath.Int
	AlertTypes		[]string
}

type EconomicSecurityEpochReport struct {
	EpochID			uint64
	Alerts			[]EconomicSecurityAlert
	CircuitBreaker		EconomicCircuitBreakerOutput
	PenaltyRouting		PenaltyRoutingDecision
	ReserveAccounting	SecurityReserveAccounting
	InvariantEvents		[]EconomicSecurityIncidentEvent
	GovernanceReport	EconomicSecurityGovernanceReport
	AuditLogs		[]EconomicSecurityAuditLog
	Passed			bool
	Failed			[]string
}

func DefaultEconomicSecurityParams() EconomicSecurityParams {
	return EconomicSecurityParams{
		MaxValidatorConcentrationBps:	MaxTopValidatorConcentrationBps,
		MaxTopNConcentrationBps:	DefaultTopNConcentrationThresholdBps,
		MaxStateGrowthBytes:		DefaultStateGrowthSurchargeThresholdBytes,
		MaxStakeMovementBps:		DefaultStakeMovementThresholdBps,
		SecurityReserveMinimumNaet:	sdkmath.NewInt(DefaultSecurityReserveMinimumNaet),
		SecurityReserveTargetNaet:	sdkmath.NewInt(DefaultSecurityReserveTargetNaet),
		MaxReserveFundingRequestNaet:	sdkmath.NewInt(DefaultSecurityReserveMaxRequestNaet),
		SlashingBurnRatioBps:		DefaultSecuritySlashingBurnRatioBps,
		SlashingTreasuryRatioBps:	DefaultSecuritySlashingTreasuryBps,
		ReporterRewardCapNaet:		sdkmath.NewInt(DefaultSecurityReporterRewardCapNaet),
		GovernanceThresholdVersion:	DefaultSecurityGovernanceVersion,
		CircuitBreakerParams:		DefaultEconomicCircuitBreakerParams(),
	}
}

func EvaluateEconomicSecurityEpoch(input EconomicSecurityEpochInput, params EconomicSecurityParams) (EconomicSecurityEpochReport, error) {
	params = params.withDefaults()
	if err := params.Validate(); err != nil {
		return EconomicSecurityEpochReport{}, err
	}
	if err := input.Validate(params); err != nil {
		return EconomicSecurityEpochReport{}, err
	}

	routing, err := routeSecurityPenalty(input, params)
	if err != nil {
		return EconomicSecurityEpochReport{}, err
	}
	reserve := accountSecurityReserve(input, params)
	breaker, err := EvaluateEconomicCircuitBreaker(EconomicCircuitBreakerInput{
		BlockLoadBps:		input.BlockLoadBps,
		FeeSpikeBps:		input.FeeSpikeBps,
		ControllerDriftBps:	input.ControllerDriftBps,
		FailedTxRateBps:	input.FailedTxRateBps,
		BurnToMintBps:		input.BurnToMintBps,
	}, params.CircuitBreakerParams)
	if err != nil {
		return EconomicSecurityEpochReport{}, err
	}

	alerts := economicSecurityAlerts(input, params, reserve, breaker)
	failed := economicSecurityInvariantFailures(input, params, routing, reserve, breaker)
	events := economicSecurityEvents(input, alerts, reserve, routing, breaker, failed)
	auditLogs := economicSecurityAuditLogs(input, routing, reserve, breaker, len(failed) == 0)
	governance := economicSecurityGovernanceReport(input, params, alerts, reserve, breaker)

	return EconomicSecurityEpochReport{
		EpochID:		input.EpochID,
		Alerts:			alerts,
		CircuitBreaker:		breaker,
		PenaltyRouting:		routing,
		ReserveAccounting:	reserve,
		InvariantEvents:	events,
		GovernanceReport:	governance,
		AuditLogs:		auditLogs,
		Passed:			len(failed) == 0,
		Failed:			failed,
	}, nil
}

func (p EconomicSecurityParams) Validate() error {
	if err := validateBps("max_validator_concentration_bps", p.MaxValidatorConcentrationBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("max_top_n_concentration_bps", p.MaxTopNConcentrationBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.MaxStateGrowthBytes < 0 {
		return fmt.Errorf("max_state_growth_bytes must not be negative")
	}
	if err := validateBps("max_stake_movement_bps", p.MaxStakeMovementBps, 0, BasisPoints); err != nil {
		return err
	}
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "security_reserve_minimum_naet", value: p.SecurityReserveMinimumNaet},
		{name: "security_reserve_target_naet", value: p.SecurityReserveTargetNaet},
		{name: "max_reserve_funding_request_naet", value: p.MaxReserveFundingRequestNaet},
		{name: "reporter_reward_cap_naet", value: p.ReporterRewardCapNaet},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	if normalizeInt(p.SecurityReserveTargetNaet).LT(normalizeInt(p.SecurityReserveMinimumNaet)) {
		return fmt.Errorf("security_reserve_target_naet must be >= minimum")
	}
	if err := validateBps("slashing_burn_ratio_bps", p.SlashingBurnRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("slashing_treasury_ratio_bps", p.SlashingTreasuryRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if p.SlashingBurnRatioBps+p.SlashingTreasuryRatioBps > BasisPoints {
		return fmt.Errorf("slashing burn and treasury ratios exceed 100%%")
	}
	if p.GovernanceThresholdVersion == "" {
		return fmt.Errorf("governance_threshold_version is required")
	}
	return p.CircuitBreakerParams.Validate()
}

func (p EconomicSecurityParams) withDefaults() EconomicSecurityParams {
	defaults := DefaultEconomicSecurityParams()
	if p.MaxValidatorConcentrationBps == 0 {
		p.MaxValidatorConcentrationBps = defaults.MaxValidatorConcentrationBps
	}
	if p.MaxTopNConcentrationBps == 0 {
		p.MaxTopNConcentrationBps = defaults.MaxTopNConcentrationBps
	}
	if p.MaxStateGrowthBytes == 0 {
		p.MaxStateGrowthBytes = defaults.MaxStateGrowthBytes
	}
	if p.MaxStakeMovementBps == 0 {
		p.MaxStakeMovementBps = defaults.MaxStakeMovementBps
	}
	if p.SecurityReserveMinimumNaet.IsNil() {
		p.SecurityReserveMinimumNaet = defaults.SecurityReserveMinimumNaet
	}
	if p.SecurityReserveTargetNaet.IsNil() {
		p.SecurityReserveTargetNaet = defaults.SecurityReserveTargetNaet
	}
	if p.MaxReserveFundingRequestNaet.IsNil() {
		p.MaxReserveFundingRequestNaet = defaults.MaxReserveFundingRequestNaet
	}
	if p.SlashingBurnRatioBps == 0 {
		p.SlashingBurnRatioBps = defaults.SlashingBurnRatioBps
	}
	if p.SlashingTreasuryRatioBps == 0 {
		p.SlashingTreasuryRatioBps = defaults.SlashingTreasuryRatioBps
	}
	if p.ReporterRewardCapNaet.IsNil() {
		p.ReporterRewardCapNaet = defaults.ReporterRewardCapNaet
	}
	if p.GovernanceThresholdVersion == "" {
		p.GovernanceThresholdVersion = defaults.GovernanceThresholdVersion
	}
	if p.CircuitBreakerParams == (EconomicCircuitBreakerParams{}) {
		p.CircuitBreakerParams = defaults.CircuitBreakerParams
	}
	return p
}

func (input EconomicSecurityEpochInput) Validate(params EconomicSecurityParams) error {
	if input.EpochID == 0 {
		return fmt.Errorf("epoch_id must be positive")
	}
	for _, field := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "slashing_penalty_naet", value: input.SlashingPenaltyNaet},
		{name: "requested_reporter_reward_naet", value: input.RequestedReporterRewardNaet},
		{name: "security_reserve_balance_naet", value: input.SecurityReserveBalanceNaet},
		{name: "security_reserve_inflow_naet", value: input.SecurityReserveInflowNaet},
		{name: "security_reserve_outflow_naet", value: input.SecurityReserveOutflowNaet},
	} {
		if normalizeInt(field.value).IsNegative() {
			return fmt.Errorf("%s must not be negative", field.name)
		}
	}
	for _, field := range []struct {
		name	string
		value	int64
	}{
		{name: "validator_concentration_bps", value: input.ValidatorConcentrationBps},
		{name: "top_n_concentration_bps", value: input.TopNConcentrationBps},
		{name: "block_load_bps", value: input.BlockLoadBps},
		{name: "failed_tx_rate_bps", value: input.FailedTxRateBps},
		{name: "burn_to_mint_bps", value: input.BurnToMintBps},
		{name: "stake_inflow_bps", value: input.StakeInflowBps},
		{name: "stake_outflow_bps", value: input.StakeOutflowBps},
	} {
		if err := validateBps(field.name, field.value, 0, DefaultMaxLoadMultiplierBps); err != nil {
			return err
		}
	}
	if input.FeeSpikeBps < 0 {
		return fmt.Errorf("fee_spike_bps must not be negative")
	}
	if input.ControllerDriftBps < 0 {
		return fmt.Errorf("controller_drift_bps must not be negative")
	}
	if input.StateGrowthBytes < 0 {
		return fmt.Errorf("state_growth_bytes must not be negative")
	}
	if input.GovernanceThresholdVersion != "" && input.GovernanceThresholdVersion != params.GovernanceThresholdVersion {
		return fmt.Errorf("governance_threshold_version mismatch")
	}
	return nil
}

func routeSecurityPenalty(input EconomicSecurityEpochInput, params EconomicSecurityParams) (PenaltyRoutingDecision, error) {
	penalty := normalizeInt(input.SlashingPenaltyNaet)
	if !penalty.IsPositive() {
		return PenaltyRoutingDecision{
			PenaltyNaet:		sdkmath.ZeroInt(),
			EvidenceAccepted:	input.EvidenceAccepted,
			EvidenceDuplicate:	input.EvidenceDuplicate,
		}, nil
	}
	reporterReward := sdkmath.ZeroInt()
	if input.EvidenceAccepted && !input.EvidenceDuplicate {
		reporterReward = minInt(normalizeInt(input.RequestedReporterRewardNaet), normalizeInt(params.ReporterRewardCapNaet))
		reporterReward = minInt(reporterReward, penalty)
	}
	reporterRatio := int64(0)
	if reporterReward.IsPositive() {
		reporterRatio = reporterReward.MulRaw(BasisPoints).Quo(penalty).Int64()
	}
	flow, err := ComputeSlashingEconomyFlow(SlashingEconomyFlowInput{
		PenaltyNaet:		penalty,
		BurnRatioBps:		params.SlashingBurnRatioBps,
		TreasuryRatioBps:	params.SlashingTreasuryRatioBps,
		ReporterRewardBps:	reporterRatio,
	})
	if err != nil {
		return PenaltyRoutingDecision{}, err
	}
	return PenaltyRoutingDecision{
		PenaltyNaet:		flow.PenaltyNaet,
		BurnNaet:		flow.BurnNaet,
		TreasuryNaet:		flow.TreasuryNaet,
		ReporterRewardNaet:	flow.ReporterRewardNaet,
		ValidatorPoolNaet:	flow.ValidatorPoolNaet,
		EvidenceAccepted:	input.EvidenceAccepted,
		EvidenceDuplicate:	input.EvidenceDuplicate,
		Routed:			true,
	}, nil
}

func accountSecurityReserve(input EconomicSecurityEpochInput, params EconomicSecurityParams) SecurityReserveAccounting {
	starting := normalizeInt(input.SecurityReserveBalanceNaet)
	inflow := normalizeInt(input.SecurityReserveInflowNaet)
	outflow := normalizeInt(input.SecurityReserveOutflowNaet)
	ending := starting.Add(inflow).Sub(outflow)
	minimum := normalizeInt(params.SecurityReserveMinimumNaet)
	target := normalizeInt(params.SecurityReserveTargetNaet)
	request := sdkmath.ZeroInt()
	if ending.LT(minimum) {
		request = target.Sub(ending)
		if request.IsNegative() {
			request = sdkmath.ZeroInt()
		}
		request = minInt(request, normalizeInt(params.MaxReserveFundingRequestNaet))
	}
	return SecurityReserveAccounting{
		StartingBalanceNaet:	starting,
		InflowNaet:		inflow,
		OutflowNaet:		outflow,
		EndingBalanceNaet:	ending,
		MinimumBalanceNaet:	minimum,
		TargetBalanceNaet:	target,
		FundingRequestNaet:	request,
		Consistent:		!ending.IsNegative(),
	}
}

func economicSecurityAlerts(input EconomicSecurityEpochInput, params EconomicSecurityParams, reserve SecurityReserveAccounting, breaker EconomicCircuitBreakerOutput) []EconomicSecurityAlert {
	alerts := make([]EconomicSecurityAlert, 0)
	if reserve.EndingBalanceNaet.LT(reserve.MinimumBalanceNaet) {
		alerts = append(alerts, securityAmountAlert(EconomicSecurityAlertReserveLow, EconomicSecuritySeverityWarning, reserve.EndingBalanceNaet, "security reserve below governed minimum"))
	}
	if input.ValidatorConcentrationBps > params.MaxValidatorConcentrationBps {
		alerts = append(alerts, securityBpsAlert(EconomicSecurityAlertConcentration, EconomicSecuritySeverityWarning, input.ValidatorConcentrationBps, params.MaxValidatorConcentrationBps, "validator concentration above threshold"))
	}
	if input.TopNConcentrationBps > params.MaxTopNConcentrationBps {
		alerts = append(alerts, securityBpsAlert(EconomicSecurityAlertTopNConcentration, EconomicSecuritySeverityCritical, input.TopNConcentrationBps, params.MaxTopNConcentrationBps, "top-n voting power concentration above threshold"))
	}
	if input.FeeSpikeBps > params.CircuitBreakerParams.MaxFeeSpikeBps || input.ControllerDriftBps > params.CircuitBreakerParams.MaxControllerDriftBps || input.FailedTxRateBps > params.CircuitBreakerParams.MaxFailedTxRateBps {
		alerts = append(alerts, securityBpsAlert(EconomicSecurityAlertFeeInstability, EconomicSecuritySeverityWarning, maxInt64(input.FeeSpikeBps, maxInt64(input.ControllerDriftBps, input.FailedTxRateBps)), maxInt64(params.CircuitBreakerParams.MaxFeeSpikeBps, maxInt64(params.CircuitBreakerParams.MaxControllerDriftBps, params.CircuitBreakerParams.MaxFailedTxRateBps)), "fee market or controller instability detected"))
	}
	if input.StateGrowthBytes > params.MaxStateGrowthBytes {
		alerts = append(alerts, EconomicSecurityAlert{Type: EconomicSecurityAlertStateGrowth, Severity: EconomicSecuritySeverityWarning, ObservedBytes: input.StateGrowthBytes, ThresholdBytes: params.MaxStateGrowthBytes, Message: "state growth above governed threshold"})
	}
	if maxInt64(input.StakeInflowBps, input.StakeOutflowBps) > params.MaxStakeMovementBps {
		alerts = append(alerts, securityBpsAlert(EconomicSecurityAlertStakeMovement, EconomicSecuritySeverityWarning, maxInt64(input.StakeInflowBps, input.StakeOutflowBps), params.MaxStakeMovementBps, "abnormal stake movement detected"))
	}
	if input.EvidenceDuplicate {
		alerts = append(alerts, EconomicSecurityAlert{Type: EconomicSecurityAlertEvidenceDuplicate, Severity: EconomicSecuritySeverityInfo, Message: "duplicate evidence rejected without reporter reward"})
	}
	if breaker.Active {
		alerts = append(alerts, EconomicSecurityAlert{Type: EconomicSecurityAlertCircuitBreaker, Severity: EconomicSecuritySeverityCritical, Message: "deterministic economic circuit breaker active"})
	}
	sort.SliceStable(alerts, func(i, j int) bool {
		return alerts[i].Type < alerts[j].Type
	})
	return alerts
}

func economicSecurityInvariantFailures(input EconomicSecurityEpochInput, params EconomicSecurityParams, routing PenaltyRoutingDecision, reserve SecurityReserveAccounting, breaker EconomicCircuitBreakerOutput) []string {
	failed := make([]string, 0)
	if routing.Routed {
		total := normalizeInt(routing.BurnNaet).
			Add(normalizeInt(routing.TreasuryNaet)).
			Add(normalizeInt(routing.ReporterRewardNaet)).
			Add(normalizeInt(routing.ValidatorPoolNaet))
		if !total.Equal(normalizeInt(routing.PenaltyNaet)) {
			failed = append(failed, "penalty_routing_mismatch")
		}
	}
	if input.EvidenceDuplicate && normalizeInt(routing.ReporterRewardNaet).IsPositive() {
		failed = append(failed, "duplicate_evidence_reporter_reward_paid")
	}
	if !reserve.Consistent {
		failed = append(failed, "security_reserve_accounting_negative")
	}
	if breaker.Active && breaker.CooldownBlocks == 0 {
		failed = append(failed, "circuit_breaker_cooldown_missing")
	}
	if params.GovernanceThresholdVersion == "" {
		failed = append(failed, "governance_threshold_version_missing")
	}
	return uniqueStrings(failed)
}

func economicSecurityEvents(input EconomicSecurityEpochInput, alerts []EconomicSecurityAlert, reserve SecurityReserveAccounting, routing PenaltyRoutingDecision, breaker EconomicCircuitBreakerOutput, failed []string) []EconomicSecurityIncidentEvent {
	events := make([]EconomicSecurityIncidentEvent, 0, len(alerts)+len(failed)+3)
	for _, alert := range alerts {
		events = append(events, EconomicSecurityIncidentEvent{Type: alert.Type, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Severity: alert.Severity, Reason: alert.Message, AmountNaet: alert.AmountNaet, Reconciled: true})
	}
	if routing.Routed {
		events = append(events, EconomicSecurityIncidentEvent{Type: EconomicSecurityEventPenaltyRouting, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Severity: EconomicSecuritySeverityInfo, Reason: "slashing penalty routed", AmountNaet: routing.PenaltyNaet, Reconciled: true})
	}
	if reserve.FundingRequestNaet.IsPositive() {
		events = append(events, EconomicSecurityIncidentEvent{Type: EconomicSecurityEventReserveFunding, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Severity: EconomicSecuritySeverityWarning, Reason: "security reserve funding requested", AmountNaet: reserve.FundingRequestNaet, Reconciled: reserve.Consistent})
	}
	if breaker.Active {
		events = append(events, EconomicSecurityIncidentEvent{Type: EconomicSecurityEventCircuitBreaker, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Severity: EconomicSecuritySeverityCritical, Reason: "circuit breaker active", Reconciled: true})
	}
	for _, failure := range failed {
		events = append(events, EconomicSecurityIncidentEvent{Type: EconomicSecurityEventInvariantViolation, EpochID: input.EpochID, BlockHeight: input.BlockHeight, Severity: EconomicSecuritySeverityCritical, Reason: failure, Reconciled: false})
	}
	return events
}

func economicSecurityAuditLogs(input EconomicSecurityEpochInput, routing PenaltyRoutingDecision, reserve SecurityReserveAccounting, breaker EconomicCircuitBreakerOutput, passed bool) []EconomicSecurityAuditLog {
	return []EconomicSecurityAuditLog{
		{EpochID: input.EpochID, Action: "penalty_routing", Deterministic: true, Details: fmt.Sprintf("penalty=%s routed=%t", normalizeInt(routing.PenaltyNaet).String(), routing.Routed)},
		{EpochID: input.EpochID, Action: "reserve_accounting", Deterministic: true, Details: fmt.Sprintf("ending=%s request=%s consistent=%t", reserve.EndingBalanceNaet.String(), reserve.FundingRequestNaet.String(), reserve.Consistent)},
		{EpochID: input.EpochID, Action: "circuit_breaker", Deterministic: true, Details: fmt.Sprintf("active=%t reasons=%d", breaker.Active, len(breaker.Reasons))},
		{EpochID: input.EpochID, Action: "epoch_invariants", Deterministic: true, Details: fmt.Sprintf("passed=%t", passed)},
	}
}

func economicSecurityGovernanceReport(input EconomicSecurityEpochInput, params EconomicSecurityParams, alerts []EconomicSecurityAlert, reserve SecurityReserveAccounting, breaker EconomicCircuitBreakerOutput) EconomicSecurityGovernanceReport {
	alertTypes := make([]string, 0, len(alerts))
	restrictions := make([]string, 0)
	for _, alert := range alerts {
		alertTypes = append(alertTypes, alert.Type)
	}
	if breaker.Active {
		restrictions = append(restrictions, "economic_circuit_breaker_cooldown")
	}
	if reserve.FundingRequestNaet.IsPositive() {
		restrictions = append(restrictions, "security_reserve_refill_required")
	}
	sort.Strings(alertTypes)
	sort.Strings(restrictions)
	return EconomicSecurityGovernanceReport{
		ThresholdVersion:	params.GovernanceThresholdVersion,
		Summary:		fmt.Sprintf("epoch=%d alerts=%d circuit_breaker=%t reserve_request=%s", input.EpochID, len(alerts), breaker.Active, reserve.FundingRequestNaet.String()),
		ActiveRestrictions:	restrictions,
		ReserveFundingRequest:	reserve.FundingRequestNaet,
		AlertTypes:		alertTypes,
	}
}

func securityBpsAlert(alertType, severity string, observed, threshold int64, message string) EconomicSecurityAlert {
	return EconomicSecurityAlert{Type: alertType, Severity: severity, ObservedBps: observed, ThresholdBps: threshold, Message: message}
}

func securityAmountAlert(alertType, severity string, amount sdkmath.Int, message string) EconomicSecurityAlert {
	return EconomicSecurityAlert{Type: alertType, Severity: severity, AmountNaet: normalizeInt(amount), Message: message}
}
