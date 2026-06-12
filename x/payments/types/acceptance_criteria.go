package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type PaymentAcceptanceCriterionID string
type PaymentAcceptanceDomain string
type PaymentAcceptanceStatus string

const (
	PaymentAcceptanceSettlement	PaymentAcceptanceDomain	= "settlement"
	PaymentAcceptanceFraud		PaymentAcceptanceDomain	= "fraud"
	PaymentAcceptanceConditional	PaymentAcceptanceDomain	= "conditional"
	PaymentAcceptanceVirtual	PaymentAcceptanceDomain	= "virtual"
	PaymentAcceptanceExecution	PaymentAcceptanceDomain	= "execution"
	PaymentAcceptanceRecovery	PaymentAcceptanceDomain	= "recovery"
	PaymentAcceptanceEconomics	PaymentAcceptanceDomain	= "economics"
	PaymentAcceptanceSecurity	PaymentAcceptanceDomain	= "security"
	PaymentAcceptanceObservability	PaymentAcceptanceDomain	= "observability"

	PaymentAcceptanceStatusSatisfied	PaymentAcceptanceStatus	= "satisfied"
)

type PaymentAcceptanceCriterion struct {
	CriterionID	PaymentAcceptanceCriterionID
	Domain		PaymentAcceptanceDomain
	Description	string
	Status		PaymentAcceptanceStatus
	Evidence	[]string
	TestNames	[]string
	ItemHash	string
}

type PaymentAcceptanceReport struct {
	Criteria	[]PaymentAcceptanceCriterion
	CriterionCount	uint64
	SatisfiedCount	uint64
	DomainCount	uint64
	ReportHash	string
}

func BuildPaymentAcceptanceReport() PaymentAcceptanceReport {
	criteria := []PaymentAcceptanceCriterion{
		acceptanceCriterion("accept_bidirectional_unidirectional_trustless_settlement", PaymentAcceptanceSettlement, "Bidirectional and unidirectional channels can settle trustlessly.", []string{"ChannelTypeBidirectional", "ChannelTypeUnidirectional", "FinalizeSettlementWithRequest", "ReceiverClose"}, []string{"TestPaymentChannelCloseDisputeFraudAndSettlement", "TestUnidirectionalStreamingClaimAndPayerReclaim"}),
		acceptanceCriterion("accept_any_participant_unilateral_close", PaymentAcceptanceSettlement, "Any participant can close unilaterally.", []string{"SubmitCloseWithRequest", "MsgUnilateralClose", "CloseReasonUnilateral"}, []string{"TestUnilateralCloseRequestStoresReasonAndDetachedSignatures"}),
		acceptanceCriterion("accept_stale_close_dispute_newer_state", PaymentAcceptanceSettlement, "Stale closes can be disputed with newer signed states.", []string{"DisputeChannel", "MsgDisputeClose", "FraudProofTypeStaleClose"}, []string{"TestPaymentChannelCloseDisputeFraudAndSettlement"}),
		acceptanceCriterion("accept_same_nonce_conflict_penalizable", PaymentAcceptanceFraud, "Same-nonce conflicting signatures are slashable or penalizable.", []string{"FraudProofTypeDoubleSign", "SubmitFraudProofWithPolicy", "BuildPenaltyRouteAccounting"}, []string{"TestPaymentChannelCloseDisputeFraudAndSettlement", "TestFraudProofInvalidBalanceRoutesPenaltyRemainder"}),
		acceptanceCriterion("accept_conditional_hash_time_settlement", PaymentAcceptanceConditional, "Conditional payments support hash-lock and time-lock settlement.", []string{"VerifyPromisePreimage", "RevealPromisePreimage", "ExpireConditionalPromises", "ValidatePromiseTimeoutOrdering"}, []string{"TestRequiredPaymentTestCoverageMatrixCoversUnitAndIntegrationSpecs", "TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria"}),
		acceptanceCriterion("accept_multihop_atomic_settlement", PaymentAcceptanceConditional, "Multi-hop payments preserve atomic settlement semantics.", []string{"BatchSettleLinkedPromises", "ConditionLinkageProof", "ConditionSettlementModePreimage"}, []string{"TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria"}),
		acceptanceCriterion("accept_virtual_parent_reservations", PaymentAcceptanceVirtual, "Virtual channels can be backed by parent-channel reservations.", []string{"VirtualParentReserve", "BuildVirtualActivationProof", "ValidateVirtualActivationProof", "VirtualReserveRelease"}, []string{"TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria"}),
		acceptanceCriterion("accept_store_v2_compact_queryable", PaymentAcceptanceExecution, "Store v2 state layout is compact and queryable.", []string{"BuildStoreV2Layout", "StoreV2Layout.Validate", "QueryStoreV2ParticipantChannels", "QueryPendingFinalizations"}, []string{"TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria"}),
		acceptanceCriterion("accept_blockstm_independent_settlements_parallel", PaymentAcceptanceExecution, "Independent settlement transactions are parallelizable under BlockSTM.", []string{"AccessPlanForSettlementOperation", "ProfileBlockSTMConflicts", "PaymentChannelMessagesConflictProfile"}, []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads"}),
		acceptanceCriterion("accept_adaptivesync_consensus_state_recovery", PaymentAcceptanceRecovery, "AdaptiveSync recovery restores all consensus-critical payment state.", []string{"BuildAdaptiveSyncSnapshot", "RecoverAdaptiveSyncSafety", "AdaptiveSyncWatcherReplayEvent"}, []string{"TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria"}),
		acceptanceCriterion("accept_fees_cover_storage_and_disputes", PaymentAcceptanceEconomics, "Fees cover channel storage and dispute verification costs.", []string{"DefaultPaymentFeeSchedule", "RequiredPaymentFee", "EstimateChannelStorageFootprint", "MeterFraudProofVerification"}, []string{"TestRequiredPaymentTestCoverageMatrixCoversUnitAndIntegrationSpecs"}),
		acceptanceCriterion("accept_fraud_proofs_deterministic_bounded_tested", PaymentAcceptanceFraud, "Fraud proofs are deterministic, bounded, and test-covered.", []string{"ComputeCanonicalFraudEvidenceHash", "MeterFraudProofVerification", "ValidateFraudProofSubmission"}, []string{"TestPaymentRoadmapPhase0ThroughPhase6VectorsAndExitCriteria", "TestRequiredPaymentTestCoverageMatrixCoversUnitAndIntegrationSpecs"}),
		acceptanceCriterion("accept_replay_protection_full_domain", PaymentAcceptanceSecurity, "Replay protection covers chain, channel, nonce, epoch, and finalization domains.", []string{"ValidateSignatureEnvelope", "ClosedChannelTombstone", "RejectEarlyTombstonePruning", "SignaturePreimageHash"}, []string{"TestRequiredPaymentTestCoverageMatrixCoversUnitAndIntegrationSpecs"}),
		acceptanceCriterion("accept_observability_liquidity_settlement_dispute_fee_perf", PaymentAcceptanceObservability, "Observability covers liquidity, settlement, disputes, fees, and performance.", []string{"BuildPaymentObservabilityMetrics", "EvaluatePaymentObservabilityAlerts", "BuildPaymentObservabilityReports", "PaymentReportWeeklyPerformance"}, []string{"TestPaymentObservabilityMetricsCoverOperationalSignals"}),
	}
	report := PaymentAcceptanceReport{Criteria: normalizePaymentAcceptanceCriteria(criteria)}
	domains := map[PaymentAcceptanceDomain]struct{}{}
	for _, criterion := range report.Criteria {
		report.CriterionCount++
		if criterion.Status == PaymentAcceptanceStatusSatisfied {
			report.SatisfiedCount++
		}
		domains[criterion.Domain] = struct{}{}
	}
	report.DomainCount = uint64(len(domains))
	report.ReportHash = ComputePaymentAcceptanceReportHash(report)
	return report
}

func ValidatePaymentAcceptanceReport(report PaymentAcceptanceReport) error {
	report.Criteria = normalizePaymentAcceptanceCriteria(report.Criteria)
	required := paymentAcceptanceCriterionIDs()
	seen := make(map[PaymentAcceptanceCriterionID]struct{}, len(required))
	domains := map[PaymentAcceptanceDomain]struct{}{}
	var criterionCount, satisfiedCount uint64
	for _, criterion := range report.Criteria {
		criterion = criterion.Normalize()
		if !isPaymentAcceptanceCriterionID(criterion.CriterionID) {
			return fmt.Errorf("unknown payments acceptance criterion %q", criterion.CriterionID)
		}
		if _, duplicate := seen[criterion.CriterionID]; duplicate {
			return fmt.Errorf("duplicate payments acceptance criterion %q", criterion.CriterionID)
		}
		seen[criterion.CriterionID] = struct{}{}
		if criterion.Description == "" || len(criterion.Evidence) == 0 || len(criterion.TestNames) == 0 {
			return fmt.Errorf("payments acceptance criterion %q lacks description, evidence, or tests", criterion.CriterionID)
		}
		if criterion.Status != PaymentAcceptanceStatusSatisfied {
			return fmt.Errorf("payments acceptance criterion %q is not satisfied", criterion.CriterionID)
		}
		if !isPaymentAcceptanceDomain(criterion.Domain) {
			return fmt.Errorf("unknown payments acceptance domain %q", criterion.Domain)
		}
		if err := ValidateHash("payments acceptance criterion hash", criterion.ItemHash); err != nil {
			return err
		}
		if expected := ComputePaymentAcceptanceCriterionHash(criterion); criterion.ItemHash != expected {
			return fmt.Errorf("payments acceptance criterion %q hash mismatch", criterion.CriterionID)
		}
		criterionCount++
		satisfiedCount++
		domains[criterion.Domain] = struct{}{}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("missing payments acceptance criterion %q", id)
		}
	}
	if criterionCount != 14 || satisfiedCount != 14 {
		return errors.New("payments acceptance criteria count must match section 20")
	}
	if len(domains) != 9 {
		return errors.New("payments acceptance criteria must cover all hardening domains")
	}
	if report.CriterionCount != criterionCount || report.SatisfiedCount != satisfiedCount || report.DomainCount != uint64(len(domains)) {
		return errors.New("payments acceptance criteria counters are invalid")
	}
	if err := ValidateHash("payments acceptance report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputePaymentAcceptanceReportHash(report); report.ReportHash != expected {
		return errors.New("payments acceptance report hash mismatch")
	}
	return nil
}

func ComputePaymentAcceptanceReportHash(report PaymentAcceptanceReport) string {
	report.Criteria = normalizePaymentAcceptanceCriteria(report.Criteria)
	parts := []string{
		"payments-acceptance-criteria-v1",
		fmt.Sprintf("%020d", report.CriterionCount),
		fmt.Sprintf("%020d", report.SatisfiedCount),
		fmt.Sprintf("%020d", report.DomainCount),
	}
	for _, criterion := range report.Criteria {
		criterion = criterion.Normalize()
		parts = append(parts, string(criterion.CriterionID), string(criterion.Domain), string(criterion.Status), criterion.Description, criterion.ItemHash)
		parts = append(parts, criterion.Evidence...)
		parts = append(parts, criterion.TestNames...)
	}
	return HashParts(parts...)
}

func ComputePaymentAcceptanceCriterionHash(criterion PaymentAcceptanceCriterion) string {
	criterion = criterion.Normalize()
	parts := []string{"payments-acceptance-criterion", string(criterion.CriterionID), string(criterion.Domain), string(criterion.Status), criterion.Description}
	parts = append(parts, criterion.Evidence...)
	parts = append(parts, criterion.TestNames...)
	return HashParts(parts...)
}

func (criterion PaymentAcceptanceCriterion) Normalize() PaymentAcceptanceCriterion {
	criterion.Description = strings.TrimSpace(criterion.Description)
	criterion.Evidence = normalizePaymentAcceptanceStrings(criterion.Evidence)
	criterion.TestNames = normalizePaymentAcceptanceStrings(criterion.TestNames)
	criterion.ItemHash = normalizeOptionalHash(criterion.ItemHash)
	return criterion
}

func acceptanceCriterion(id PaymentAcceptanceCriterionID, domain PaymentAcceptanceDomain, description string, evidence, tests []string) PaymentAcceptanceCriterion {
	criterion := PaymentAcceptanceCriterion{
		CriterionID:	id,
		Domain:		domain,
		Description:	description,
		Status:		PaymentAcceptanceStatusSatisfied,
		Evidence:	evidence,
		TestNames:	tests,
	}
	criterion.ItemHash = ComputePaymentAcceptanceCriterionHash(criterion)
	return criterion.Normalize()
}

func normalizePaymentAcceptanceCriteria(criteria []PaymentAcceptanceCriterion) []PaymentAcceptanceCriterion {
	out := make([]PaymentAcceptanceCriterion, 0, len(criteria))
	for _, criterion := range criteria {
		out = append(out, criterion.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CriterionID < out[j].CriterionID })
	return out
}

func normalizePaymentAcceptanceStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func paymentAcceptanceCriterionIDs() []PaymentAcceptanceCriterionID {
	return []PaymentAcceptanceCriterionID{
		"accept_bidirectional_unidirectional_trustless_settlement",
		"accept_any_participant_unilateral_close",
		"accept_stale_close_dispute_newer_state",
		"accept_same_nonce_conflict_penalizable",
		"accept_conditional_hash_time_settlement",
		"accept_multihop_atomic_settlement",
		"accept_virtual_parent_reservations",
		"accept_store_v2_compact_queryable",
		"accept_blockstm_independent_settlements_parallel",
		"accept_adaptivesync_consensus_state_recovery",
		"accept_fees_cover_storage_and_disputes",
		"accept_fraud_proofs_deterministic_bounded_tested",
		"accept_replay_protection_full_domain",
		"accept_observability_liquidity_settlement_dispute_fee_perf",
	}
}

func isPaymentAcceptanceCriterionID(id PaymentAcceptanceCriterionID) bool {
	for _, required := range paymentAcceptanceCriterionIDs() {
		if id == required {
			return true
		}
	}
	return false
}

func isPaymentAcceptanceDomain(domain PaymentAcceptanceDomain) bool {
	switch domain {
	case PaymentAcceptanceSettlement,
		PaymentAcceptanceFraud,
		PaymentAcceptanceConditional,
		PaymentAcceptanceVirtual,
		PaymentAcceptanceExecution,
		PaymentAcceptanceRecovery,
		PaymentAcceptanceEconomics,
		PaymentAcceptanceSecurity,
		PaymentAcceptanceObservability:
		return true
	default:
		return false
	}
}
