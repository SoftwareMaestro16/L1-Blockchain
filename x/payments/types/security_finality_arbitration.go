package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type EconomicFinalityMode string

const (
	EconomicFinalityCooperative	EconomicFinalityMode	= "COOPERATIVE_SETTLEMENT"
	EconomicFinalityUnilateral	EconomicFinalityMode	= "UNILATERAL_SETTLEMENT"
	EconomicFinalityConditional	EconomicFinalityMode	= "CONDITIONAL_SETTLEMENT"
	EconomicFinalityVirtual		EconomicFinalityMode	= "VIRTUAL_CHANNEL_SETTLEMENT"
	EconomicFinalityPenalty		EconomicFinalityMode	= "PENALTY_SETTLEMENT"
)

type EconomicFinalityRequirement struct {
	Mode				EconomicFinalityMode
	RequiresInclusion		bool
	RequiresChallengeExpiry		bool
	RequiresProofOrTimeout		bool
	RequiresParentReserveResolve	bool
	RequiresAcceptedFraudProof	bool
	Description			string
}

type ChallengePeriodSizing struct {
	MessagePropagationDelay	uint64
	WatchServiceReaction	uint64
	CongestionBuffer	uint64
	MultiHopTimeoutMargin	uint64
}

type EconomicFinalityCheck struct {
	Mode		EconomicFinalityMode
	Passed		bool
	EvidenceHash	string
	Reason		string
}

type EconomicFinalityReport struct {
	Requirements		[]EconomicFinalityRequirement
	Checks			[]EconomicFinalityCheck
	RequiredChallengeSize	uint64
	ReportHash		string
}

type DisputePriorityPolicy struct {
	BasePriority		uint64
	NearExpiryBoost		uint64
	FraudProofBoost		uint64
	CongestionBoostBps	uint32
	NearExpiryThreshold	uint64
	MaxPriority		uint64
	RequiredFeeClass	PaymentFeeClass
	Deterministic		bool
}

type DisputePriorityRequest struct {
	Operation		SettlementArbitrationOperation
	OperationID		string
	ChannelID		string
	SubmittedHeight		uint64
	CurrentHeight		uint64
	SettleAfterHeight	uint64
	HasFraudProof		bool
	FeePaid			string
	RequiredFee		string
	EstimatedGas		uint64
	CongestionBps		uint32
}

type DisputePriorityDecision struct {
	OperationID	string
	ChannelID	string
	Operation	SettlementArbitrationOperation
	PriorityScore	uint64
	BlocksRemaining	uint64
	NearExpiry	bool
	FeeCovered	bool
	Deterministic	bool
	EstimatedGas	uint64
	PolicyHash	string
	DecisionHash	string
	DeferredReason	string
}

type DisputeInclusionStressResult struct {
	Included	[]DisputePriorityDecision
	Deferred	[]DisputePriorityDecision
	Conflicts	[]BlockSTMConflict
	TotalGas	uint64
	GasLimit	uint64
	ConflictFree	bool
	StressHash	string
}

type NearExpiryDisputeAlert struct {
	ChannelID		string
	SubmittedHeight		uint64
	SettleAfterHeight	uint64
	CurrentHeight		uint64
	BlocksRemaining		uint64
	DisputeCount		uint32
	Severity		string
	EvidenceHash		string
}

func DefaultEconomicFinalityRequirements() []EconomicFinalityRequirement {
	return []EconomicFinalityRequirement{
		{Mode: EconomicFinalityCooperative, RequiresInclusion: true, Description: "cooperative settlement finalizes after on-chain inclusion and successful execution"},
		{Mode: EconomicFinalityUnilateral, RequiresInclusion: true, RequiresChallengeExpiry: true, Description: "unilateral settlement finalizes after the challenge period expires"},
		{Mode: EconomicFinalityConditional, RequiresInclusion: true, RequiresProofOrTimeout: true, Description: "conditional settlement finalizes after valid proof or timeout resolution"},
		{Mode: EconomicFinalityVirtual, RequiresInclusion: true, RequiresParentReserveResolve: true, Description: "virtual settlement finalizes after endpoint state and parent reserve resolution"},
		{Mode: EconomicFinalityPenalty, RequiresInclusion: true, RequiresAcceptedFraudProof: true, Description: "penalty settlement finalizes with accepted fraud proof execution"},
	}
}

func ValidateEconomicFinalityRequirements(requirements []EconomicFinalityRequirement) error {
	seen := map[EconomicFinalityMode]struct{}{}
	for _, requirement := range normalizeEconomicFinalityRequirements(requirements) {
		if !IsEconomicFinalityMode(requirement.Mode) {
			return fmt.Errorf("unknown payments economic finality mode %q", requirement.Mode)
		}
		if _, duplicate := seen[requirement.Mode]; duplicate {
			return fmt.Errorf("duplicate payments economic finality mode %q", requirement.Mode)
		}
		if !requirement.RequiresInclusion {
			return fmt.Errorf("payments economic finality mode %q must require inclusion", requirement.Mode)
		}
		seen[requirement.Mode] = struct{}{}
	}
	for _, mode := range requiredEconomicFinalityModes() {
		if _, found := seen[mode]; !found {
			return fmt.Errorf("missing payments economic finality mode %q", mode)
		}
	}
	return nil
}

func BuildEconomicFinalityReport(state PaymentsState, currentHeight uint64, sizing ChallengePeriodSizing) (EconomicFinalityReport, error) {
	if currentHeight == 0 {
		return EconomicFinalityReport{}, errors.New("payments economic finality report height must be positive")
	}
	state = state.Export()
	requirements := DefaultEconomicFinalityRequirements()
	if err := ValidateEconomicFinalityRequirements(requirements); err != nil {
		return EconomicFinalityReport{}, err
	}
	requiredChallenge := sizing.TotalRequired()
	if requiredChallenge == 0 {
		return EconomicFinalityReport{}, errors.New("payments economic finality challenge sizing must be positive")
	}
	checks := []EconomicFinalityCheck{
		buildEconomicFinalityCheck(state, EconomicFinalityCooperative, economicCooperativeFinalityPasses(state)),
		buildEconomicFinalityCheck(state, EconomicFinalityUnilateral, economicUnilateralFinalityPasses(state, currentHeight, requiredChallenge)),
		buildEconomicFinalityCheck(state, EconomicFinalityConditional, economicConditionalFinalityPasses(state)),
		buildEconomicFinalityCheck(state, EconomicFinalityVirtual, economicVirtualFinalityPasses(state)),
		buildEconomicFinalityCheck(state, EconomicFinalityPenalty, economicPenaltyFinalityPasses(state)),
	}
	report := EconomicFinalityReport{
		Requirements:		normalizeEconomicFinalityRequirements(requirements),
		Checks:			normalizeEconomicFinalityChecks(checks),
		RequiredChallengeSize:	requiredChallenge,
	}
	report.ReportHash = ComputeEconomicFinalityReportHash(report)
	if err := report.Validate(); err != nil {
		return EconomicFinalityReport{}, err
	}
	return report, nil
}

func ComputeEconomicFinalityReportHash(report EconomicFinalityReport) string {
	report.Requirements = normalizeEconomicFinalityRequirements(report.Requirements)
	report.Checks = normalizeEconomicFinalityChecks(report.Checks)
	parts := []string{"payments-economic-finality-v1", fmt.Sprintf("%d", report.RequiredChallengeSize)}
	for _, requirement := range report.Requirements {
		parts = append(parts,
			string(requirement.Mode),
			fmt.Sprintf("%t", requirement.RequiresInclusion),
			fmt.Sprintf("%t", requirement.RequiresChallengeExpiry),
			fmt.Sprintf("%t", requirement.RequiresProofOrTimeout),
			fmt.Sprintf("%t", requirement.RequiresParentReserveResolve),
			fmt.Sprintf("%t", requirement.RequiresAcceptedFraudProof),
		)
	}
	for _, check := range report.Checks {
		parts = append(parts, string(check.Mode), fmt.Sprintf("%t", check.Passed), check.EvidenceHash, check.Reason)
	}
	return HashParts(parts...)
}

func (report EconomicFinalityReport) Validate() error {
	report.Requirements = normalizeEconomicFinalityRequirements(report.Requirements)
	report.Checks = normalizeEconomicFinalityChecks(report.Checks)
	if err := ValidateEconomicFinalityRequirements(report.Requirements); err != nil {
		return err
	}
	seen := map[EconomicFinalityMode]struct{}{}
	for _, check := range report.Checks {
		check = check.Normalize()
		if !IsEconomicFinalityMode(check.Mode) {
			return fmt.Errorf("unknown payments economic finality check mode %q", check.Mode)
		}
		if _, duplicate := seen[check.Mode]; duplicate {
			return fmt.Errorf("duplicate payments economic finality check %q", check.Mode)
		}
		if !check.Passed {
			return fmt.Errorf("payments economic finality check %q failed: %s", check.Mode, check.Reason)
		}
		if err := ValidateHash("payments economic finality evidence hash", check.EvidenceHash); err != nil {
			return err
		}
		seen[check.Mode] = struct{}{}
	}
	for _, mode := range requiredEconomicFinalityModes() {
		if _, found := seen[mode]; !found {
			return fmt.Errorf("missing payments economic finality check %q", mode)
		}
	}
	if err := ValidateHash("payments economic finality report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputeEconomicFinalityReportHash(report); report.ReportHash != expected {
		return errors.New("payments economic finality report hash mismatch")
	}
	return nil
}

func (s ChallengePeriodSizing) TotalRequired() uint64 {
	total := s.MessagePropagationDelay + s.WatchServiceReaction
	if total < s.MessagePropagationDelay {
		return ^uint64(0)
	}
	total += s.CongestionBuffer
	if total < s.CongestionBuffer {
		return ^uint64(0)
	}
	total += s.MultiHopTimeoutMargin
	if total < s.MultiHopTimeoutMargin {
		return ^uint64(0)
	}
	return total
}

func ValidateChallengePeriodSizing(period uint64, sizing ChallengePeriodSizing) error {
	if period == 0 {
		return errors.New("payments challenge period must be positive")
	}
	required := sizing.TotalRequired()
	if required == 0 {
		return errors.New("payments challenge period sizing components must be positive")
	}
	if period <= required {
		return fmt.Errorf("payments challenge period %d must exceed required sizing %d", period, required)
	}
	return validateChallengePeriod(period)
}

func DefaultDisputePriorityPolicy() DisputePriorityPolicy {
	return DisputePriorityPolicy{
		BasePriority:		1_000,
		NearExpiryBoost:	10_000,
		FraudProofBoost:	5_000,
		CongestionBoostBps:	2_500,
		NearExpiryThreshold:	4,
		MaxPriority:		100_000,
		RequiredFeeClass:	PaymentFeeClassDispute,
		Deterministic:		true,
	}
}

func ComputeDisputeTransactionPriority(policy DisputePriorityPolicy, req DisputePriorityRequest) (DisputePriorityDecision, error) {
	policy = policy.Normalize()
	req = req.Normalize()
	if err := policy.Validate(); err != nil {
		return DisputePriorityDecision{}, err
	}
	if err := req.Validate(); err != nil {
		return DisputePriorityDecision{}, err
	}
	paid, err := parseNonNegativeInt("payments dispute priority fee paid", req.FeePaid)
	if err != nil {
		return DisputePriorityDecision{}, err
	}
	required, err := parseNonNegativeInt("payments dispute priority required fee", req.RequiredFee)
	if err != nil {
		return DisputePriorityDecision{}, err
	}
	remaining := req.SettleAfterHeight - req.CurrentHeight
	score := policy.BasePriority
	nearExpiry := remaining <= policy.NearExpiryThreshold
	if nearExpiry {
		score += policy.NearExpiryBoost
	}
	if req.HasFraudProof || req.Operation == SettlementArbitrationFraudProof {
		score += policy.FraudProofBoost
	}
	congestionBoost := policy.BasePriority * uint64(req.CongestionBps) * uint64(policy.CongestionBoostBps) / 100_000_000
	score += congestionBoost
	if score > policy.MaxPriority {
		score = policy.MaxPriority
	}
	decision := DisputePriorityDecision{
		OperationID:		req.OperationID,
		ChannelID:		req.ChannelID,
		Operation:		req.Operation,
		PriorityScore:		score,
		BlocksRemaining:	remaining,
		NearExpiry:		nearExpiry,
		FeeCovered:		!paid.LT(required),
		Deterministic:		policy.Deterministic,
		EstimatedGas:		req.EstimatedGas,
		PolicyHash:		policy.Hash(),
	}
	if !decision.FeeCovered {
		decision.DeferredReason = "insufficient dispute fee"
	}
	decision.DecisionHash = decision.Hash()
	return decision, nil
}

func SimulateDisputeInclusionStress(policy DisputePriorityPolicy, requests []DisputePriorityRequest, gasLimit uint64) (DisputeInclusionStressResult, error) {
	if gasLimit == 0 {
		return DisputeInclusionStressResult{}, errors.New("payments dispute stress gas limit must be positive")
	}
	decisions := make([]DisputePriorityDecision, 0, len(requests))
	plans := make([]BlockSTMAccessPlan, 0, len(requests))
	for _, req := range requests {
		decision, err := ComputeDisputeTransactionPriority(policy, req)
		if err != nil {
			return DisputeInclusionStressResult{}, err
		}
		decisions = append(decisions, decision)
		opType := BatchOperationDispute
		if req.Operation == SettlementArbitrationFinalSettlement {
			opType = BatchOperationSettle
		}
		plan, err := AccessPlanForSettlementOperation(SettlementOperation{
			OperationID:	decision.OperationID,
			ChannelID:	decision.ChannelID,
			OperationType:	opType,
			Nonce:		1,
			StateHash:	HashParts("dispute-priority-state", decision.OperationID),
		}, req.CurrentHeight)
		if err != nil {
			return DisputeInclusionStressResult{}, err
		}
		plans = append(plans, plan)
	}
	sort.SliceStable(decisions, func(i, j int) bool {
		if decisions[i].PriorityScore == decisions[j].PriorityScore {
			return decisions[i].DecisionHash < decisions[j].DecisionHash
		}
		return decisions[i].PriorityScore > decisions[j].PriorityScore
	})
	result := DisputeInclusionStressResult{GasLimit: gasLimit}
	for _, decision := range decisions {
		if !decision.FeeCovered {
			result.Deferred = append(result.Deferred, decision)
			continue
		}
		if result.TotalGas+decision.EstimatedGas > gasLimit {
			decision.DeferredReason = "block gas limit"
			decision.DecisionHash = decision.Hash()
			result.Deferred = append(result.Deferred, decision)
			continue
		}
		result.TotalGas += decision.EstimatedGas
		result.Included = append(result.Included, decision)
	}
	profile := ProfileBlockSTMConflicts(plans)
	result.Conflicts = profile.Conflicts
	result.ConflictFree = profile.ConflictFree
	result.StressHash = result.Hash()
	return result, nil
}

func MonitorNearExpiryDisputes(state PaymentsState, currentHeight, threshold uint64) ([]NearExpiryDisputeAlert, error) {
	if currentHeight == 0 {
		return nil, errors.New("payments near-expiry monitor height must be positive")
	}
	state = state.Export()
	alerts := []NearExpiryDisputeAlert{}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status != ChannelStatusPendingClose {
			continue
		}
		if currentHeight > channel.PendingClose.SettleAfterHeight {
			continue
		}
		remaining := channel.PendingClose.SettleAfterHeight - currentHeight
		if remaining > threshold {
			continue
		}
		severity := "warning"
		if remaining <= 1 {
			severity = "critical"
		}
		alert := NearExpiryDisputeAlert{
			ChannelID:		channel.ChannelID,
			SubmittedHeight:	channel.PendingClose.SubmittedHeight,
			SettleAfterHeight:	channel.PendingClose.SettleAfterHeight,
			CurrentHeight:		currentHeight,
			BlocksRemaining:	remaining,
			DisputeCount:		channel.PendingClose.DisputeCount,
			Severity:		severity,
			EvidenceHash:		HashParts("near-expiry-dispute", channel.ChannelID, fmt.Sprintf("%d", currentHeight), fmt.Sprintf("%d", remaining)),
		}
		alerts = append(alerts, alert)
	}
	sort.SliceStable(alerts, func(i, j int) bool {
		if alerts[i].BlocksRemaining == alerts[j].BlocksRemaining {
			return alerts[i].ChannelID < alerts[j].ChannelID
		}
		return alerts[i].BlocksRemaining < alerts[j].BlocksRemaining
	})
	return alerts, nil
}

func (r EconomicFinalityRequirement) Normalize() EconomicFinalityRequirement {
	r.Description = strings.TrimSpace(r.Description)
	return r
}

func (c EconomicFinalityCheck) Normalize() EconomicFinalityCheck {
	c.EvidenceHash = normalizeHash(c.EvidenceHash)
	c.Reason = strings.TrimSpace(c.Reason)
	return c
}

func (p DisputePriorityPolicy) Normalize() DisputePriorityPolicy {
	defaults := DefaultDisputePriorityPolicy()
	if p.BasePriority == 0 {
		p.BasePriority = defaults.BasePriority
	}
	if p.NearExpiryBoost == 0 {
		p.NearExpiryBoost = defaults.NearExpiryBoost
	}
	if p.FraudProofBoost == 0 {
		p.FraudProofBoost = defaults.FraudProofBoost
	}
	if p.CongestionBoostBps == 0 {
		p.CongestionBoostBps = defaults.CongestionBoostBps
	}
	if p.NearExpiryThreshold == 0 {
		p.NearExpiryThreshold = defaults.NearExpiryThreshold
	}
	if p.MaxPriority == 0 {
		p.MaxPriority = defaults.MaxPriority
	}
	if p.RequiredFeeClass == "" {
		p.RequiredFeeClass = PaymentFeeClassDispute
	}
	if !p.Deterministic {
		p.Deterministic = defaults.Deterministic
	}
	return p
}

func (p DisputePriorityPolicy) Validate() error {
	p = p.Normalize()
	if p.BasePriority == 0 || p.NearExpiryBoost == 0 || p.FraudProofBoost == 0 || p.MaxPriority == 0 {
		return errors.New("payments dispute priority weights must be positive")
	}
	if p.BasePriority > p.MaxPriority {
		return errors.New("payments dispute priority base exceeds maximum")
	}
	if p.CongestionBoostBps > MaxPenaltyRouteBps {
		return errors.New("payments dispute priority congestion boost exceeds bps maximum")
	}
	if !IsPaymentFeeClass(p.RequiredFeeClass) {
		return fmt.Errorf("unknown payments dispute priority fee class %q", p.RequiredFeeClass)
	}
	if !p.Deterministic {
		return errors.New("payments dispute priority policy must be deterministic")
	}
	return nil
}

func (p DisputePriorityPolicy) Hash() string {
	p = p.Normalize()
	return HashParts(
		"dispute-priority-policy",
		fmt.Sprintf("%d", p.BasePriority),
		fmt.Sprintf("%d", p.NearExpiryBoost),
		fmt.Sprintf("%d", p.FraudProofBoost),
		fmt.Sprintf("%d", p.CongestionBoostBps),
		fmt.Sprintf("%d", p.NearExpiryThreshold),
		fmt.Sprintf("%d", p.MaxPriority),
		string(p.RequiredFeeClass),
		fmt.Sprintf("%t", p.Deterministic),
	)
}

func (r DisputePriorityRequest) Normalize() DisputePriorityRequest {
	r.OperationID = normalizeOptionalHash(r.OperationID)
	r.ChannelID = normalizeHash(r.ChannelID)
	r.FeePaid = strings.TrimSpace(r.FeePaid)
	if r.FeePaid == "" {
		r.FeePaid = "0"
	}
	r.RequiredFee = strings.TrimSpace(r.RequiredFee)
	if r.RequiredFee == "" {
		r.RequiredFee = "0"
	}
	if r.OperationID == "" && r.ChannelID != "" {
		r.OperationID = HashParts("dispute-priority-operation", r.ChannelID, string(r.Operation), fmt.Sprintf("%d", r.CurrentHeight))
	}
	if r.EstimatedGas == 0 {
		r.EstimatedGas = DefaultSettlementGasCostSchedule().DisputeGas
	}
	return r
}

func (r DisputePriorityRequest) Validate() error {
	r = r.Normalize()
	if r.Operation != SettlementArbitrationDispute && r.Operation != SettlementArbitrationFraudProof {
		return fmt.Errorf("payments dispute priority operation must be dispute or fraud proof, got %q", r.Operation)
	}
	if err := ValidateHash("payments dispute priority operation id", r.OperationID); err != nil {
		return err
	}
	if err := ValidateHash("payments dispute priority channel id", r.ChannelID); err != nil {
		return err
	}
	if r.SubmittedHeight == 0 || r.CurrentHeight == 0 || r.SettleAfterHeight == 0 {
		return errors.New("payments dispute priority heights must be positive")
	}
	if r.CurrentHeight < r.SubmittedHeight {
		return errors.New("payments dispute priority current height precedes submission")
	}
	if r.CurrentHeight > r.SettleAfterHeight {
		return errors.New("payments dispute priority window has closed")
	}
	if r.CongestionBps > MaxPenaltyRouteBps {
		return errors.New("payments dispute priority congestion exceeds bps maximum")
	}
	if r.EstimatedGas == 0 {
		return errors.New("payments dispute priority estimated gas must be positive")
	}
	if err := validateNonNegativeInt("payments dispute priority fee paid", r.FeePaid); err != nil {
		return err
	}
	return validateNonNegativeInt("payments dispute priority required fee", r.RequiredFee)
}

func (d DisputePriorityDecision) Hash() string {
	d.OperationID = normalizeHash(d.OperationID)
	d.ChannelID = normalizeHash(d.ChannelID)
	d.PolicyHash = normalizeHash(d.PolicyHash)
	return HashParts(
		"dispute-priority-decision",
		d.OperationID,
		d.ChannelID,
		string(d.Operation),
		fmt.Sprintf("%d", d.PriorityScore),
		fmt.Sprintf("%d", d.BlocksRemaining),
		fmt.Sprintf("%t", d.NearExpiry),
		fmt.Sprintf("%t", d.FeeCovered),
		fmt.Sprintf("%t", d.Deterministic),
		fmt.Sprintf("%d", d.EstimatedGas),
		d.PolicyHash,
		strings.TrimSpace(d.DeferredReason),
	)
}

func (r DisputeInclusionStressResult) Hash() string {
	parts := []string{"dispute-inclusion-stress", fmt.Sprintf("%d", r.GasLimit), fmt.Sprintf("%d", r.TotalGas), fmt.Sprintf("%t", r.ConflictFree)}
	for _, decision := range r.Included {
		parts = append(parts, decision.DecisionHash)
	}
	for _, decision := range r.Deferred {
		parts = append(parts, decision.DecisionHash, decision.DeferredReason)
	}
	for _, conflict := range r.Conflicts {
		conflict = conflict.Normalize()
		parts = append(parts, conflict.LeftOperationID, conflict.RightOperationID, conflict.Key, conflict.Reason)
	}
	return HashParts(parts...)
}

func IsEconomicFinalityMode(value EconomicFinalityMode) bool {
	for _, mode := range requiredEconomicFinalityModes() {
		if value == mode {
			return true
		}
	}
	return false
}

func buildEconomicFinalityCheck(state PaymentsState, mode EconomicFinalityMode, passed bool) EconomicFinalityCheck {
	reason := "verified"
	if !passed {
		reason = "state violates economic finality"
	}
	return EconomicFinalityCheck{
		Mode:		mode,
		Passed:		passed,
		EvidenceHash:	economicFinalityEvidenceHash(state, mode),
		Reason:		reason,
	}
}

func economicCooperativeFinalityPasses(state PaymentsState) bool {
	settlements := settlementRecordChannelSet(state.Settlements)
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status == ChannelStatusSettled {
			if _, found := settlements[channel.ChannelID]; !found {
				return false
			}
		}
	}
	return true
}

func economicUnilateralFinalityPasses(state PaymentsState, currentHeight, requiredChallenge uint64) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status != ChannelStatusPendingClose {
			continue
		}
		if channel.PendingClose.SettleAfterHeight <= channel.PendingClose.SubmittedHeight {
			return false
		}
		if requiredChallenge > 0 && channel.DisputePeriod <= requiredChallenge {
			return false
		}
		if currentHeight < channel.PendingClose.SettleAfterHeight && channel.Finality == ChannelFinalityFinalizable {
			return false
		}
	}
	return true
}

func economicConditionalFinalityPasses(state PaymentsState) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status != ChannelStatusPendingClose || len(channel.PendingClose.State.Conditions) == 0 {
			continue
		}
		if channel.Finality == ChannelFinalityFinalizable && len(channel.PendingClose.ConditionProofs) < len(channel.PendingClose.State.Conditions) {
			return false
		}
	}
	return true
}

func economicVirtualFinalityPasses(state PaymentsState) bool {
	for _, vc := range state.VirtualChannels {
		vc = vc.Normalize()
		if err := vc.ValidateCore(); err != nil {
			return false
		}
		if len(vc.ParentReserveCommitments) != len(vc.ParentChannelIDs) {
			return false
		}
	}
	return true
}

func economicPenaltyFinalityPasses(state PaymentsState) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if len(channel.PendingClose.FraudProofs) == 0 {
			continue
		}
		if len(channel.PendingClose.Penalties) == 0 && len(channel.PendingClose.PenaltyAllocations) == 0 {
			return false
		}
		if channel.Finality != ChannelFinalityPenalized {
			return false
		}
	}
	return true
}

func economicFinalityEvidenceHash(state PaymentsState, mode EconomicFinalityMode) string {
	parts := []string{"economic-finality-check", string(mode)}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		parts = append(parts, channel.ChannelID, string(channel.Status), string(channel.Finality), fmt.Sprintf("%d", channel.PendingClose.SettleAfterHeight), fmt.Sprintf("%d", len(channel.PendingClose.FraudProofs)))
	}
	for _, vc := range state.VirtualChannels {
		vc = vc.Normalize()
		parts = append(parts, vc.VirtualChannelID, string(vc.Status), vc.StateHash)
	}
	for _, settlement := range state.Settlements {
		settlement = settlement.Normalize()
		parts = append(parts, settlement.ChannelID, settlement.SettlementHash)
	}
	return HashParts(parts...)
}

func settlementRecordChannelSet(records []SettlementRecord) map[string]struct{} {
	out := make(map[string]struct{}, len(records))
	for _, record := range records {
		record = record.Normalize()
		out[record.ChannelID] = struct{}{}
	}
	return out
}

func requiredEconomicFinalityModes() []EconomicFinalityMode {
	return []EconomicFinalityMode{
		EconomicFinalityCooperative,
		EconomicFinalityUnilateral,
		EconomicFinalityConditional,
		EconomicFinalityVirtual,
		EconomicFinalityPenalty,
	}
}

func normalizeEconomicFinalityRequirements(requirements []EconomicFinalityRequirement) []EconomicFinalityRequirement {
	out := make([]EconomicFinalityRequirement, len(requirements))
	for i, requirement := range requirements {
		out[i] = requirement.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func normalizeEconomicFinalityChecks(checks []EconomicFinalityCheck) []EconomicFinalityCheck {
	out := make([]EconomicFinalityCheck, len(checks))
	for i, check := range checks {
		out[i] = check.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}
