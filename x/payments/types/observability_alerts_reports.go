package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

type PaymentObservabilityAlertType string

const (
	PaymentAlertHighPendingDisputeCount			PaymentObservabilityAlertType	= "HIGH_PENDING_DISPUTE_COUNT"
	PaymentAlertChallengeNearExpiryWithoutFinalization	PaymentObservabilityAlertType	= "CHALLENGE_NEAR_EXPIRY_WITHOUT_FINALIZATION"
	PaymentAlertFraudProofRejectionSpike			PaymentObservabilityAlertType	= "FRAUD_PROOF_REJECTION_SPIKE"
	PaymentAlertChannelOpenSpamSpike			PaymentObservabilityAlertType	= "CHANNEL_OPEN_SPAM_SPIKE"
	PaymentAlertPromiseExpiryBacklog			PaymentObservabilityAlertType	= "PROMISE_EXPIRY_BACKLOG"
	PaymentAlertSettlementQueueBacklog			PaymentObservabilityAlertType	= "SETTLEMENT_QUEUE_BACKLOG"
	PaymentAlertBlockSTMConflictRateAboveThreshold		PaymentObservabilityAlertType	= "BLOCKSTM_CONFLICT_RATE_ABOVE_THRESHOLD"
	PaymentAlertPaymentModuleStoreLatencyAboveThreshold	PaymentObservabilityAlertType	= "PAYMENT_MODULE_STORE_LATENCY_ABOVE_THRESHOLD"
	PaymentAlertWatchServiceEventReplayLag			PaymentObservabilityAlertType	= "WATCH_SERVICE_EVENT_REPLAY_LAG"
	PaymentAlertRoutingGossipSpamRateAboveThreshold		PaymentObservabilityAlertType	= "ROUTING_GOSSIP_SPAM_RATE_ABOVE_THRESHOLD"
)

type PaymentObservabilityAlertSeverity string

const (
	PaymentAlertSeverityWarning	PaymentObservabilityAlertSeverity	= "WARNING"
	PaymentAlertSeverityCritical	PaymentObservabilityAlertSeverity	= "CRITICAL"
)

type PaymentObservabilityAlertThresholds struct {
	WindowBlocks			uint64
	HighPendingDisputeCount		uint64
	NearExpiryWithoutFinalize	uint64
	FraudProofRejectionSpike	uint64
	ChannelOpenSpamSpike		uint64
	PromiseExpiryBacklog		uint64
	SettlementQueueBacklog		uint64
	BlockSTMConflictRateBps		uint64
	StoreLatencyOps			uint64
	WatchReplayLagBlocks		uint64
	RoutingGossipSpamRateBps	uint64
}

type PaymentObservabilityExternalSignals struct {
	WatcherReplayHeight		uint64
	RoutingGossipMessages		uint64
	RoutingGossipRejected		uint64
	RoutingGossipSpamRateBps	uint64
}

type PaymentObservabilityAlert struct {
	AlertID		string
	AlertType	PaymentObservabilityAlertType
	Severity	PaymentObservabilityAlertSeverity
	Height		uint64
	MetricName	string
	MetricValue	uint64
	Threshold	uint64
	Message		string
	EvidenceHash	string
}

type PaymentObservabilityReportType string

const (
	PaymentReportDailyLockedLiquidity		PaymentObservabilityReportType	= "DAILY_LOCKED_LIQUIDITY"
	PaymentReportDailySettlementVolume		PaymentObservabilityReportType	= "DAILY_SETTLEMENT_VOLUME"
	PaymentReportDailyRoutingFee			PaymentObservabilityReportType	= "DAILY_ROUTING_FEE"
	PaymentReportDailyDisputeAndFraud		PaymentObservabilityReportType	= "DAILY_DISPUTE_AND_FRAUD"
	PaymentReportDailyStateFootprint		PaymentObservabilityReportType	= "DAILY_STATE_FOOTPRINT"
	PaymentReportWeeklyChannelChurn			PaymentObservabilityReportType	= "WEEKLY_CHANNEL_CHURN"
	PaymentReportWeeklyLiquidityConcentration	PaymentObservabilityReportType	= "WEEKLY_LIQUIDITY_CONCENTRATION"
	PaymentReportWeeklyPerformance			PaymentObservabilityReportType	= "WEEKLY_PERFORMANCE"
)

type PaymentObservabilityReport struct {
	ReportID			string
	ReportType			PaymentObservabilityReportType
	PeriodStartHeight		uint64
	PeriodEndHeight			uint64
	TotalLockedNaet			string
	SettlementVolumeNaet		string
	RoutingFeesNaet			string
	DisputeCount			uint64
	FraudProofsSubmitted		uint64
	FraudProofsAccepted		uint64
	FraudProofsRejected		uint64
	StateFootprintRecords		uint64
	ChannelOpens			uint64
	ChannelSettlements		uint64
	ChannelChurnRateBps		uint64
	LargestLockedChannelNaet	string
	LiquidityConcentrationBps	uint64
	BlockSTMConflictRateBps		uint64
	StoreV2PaymentModuleLatencyOps	uint64
	DisputeInclusionLatencyBlocks	uint64
	ReportHash			string
}

func DefaultPaymentObservabilityAlertThresholds() PaymentObservabilityAlertThresholds {
	return PaymentObservabilityAlertThresholds{
		WindowBlocks:			100,
		HighPendingDisputeCount:	100,
		NearExpiryWithoutFinalize:	1,
		FraudProofRejectionSpike:	10,
		ChannelOpenSpamSpike:		1_000,
		PromiseExpiryBacklog:		500,
		SettlementQueueBacklog:		500,
		BlockSTMConflictRateBps:	2_500,
		StoreLatencyOps:		10_000,
		WatchReplayLagBlocks:		50,
		RoutingGossipSpamRateBps:	2_000,
	}
}

func EvaluatePaymentObservabilityAlerts(state PaymentsState, metrics PaymentObservabilityMetrics, thresholds PaymentObservabilityAlertThresholds, signals PaymentObservabilityExternalSignals) ([]PaymentObservabilityAlert, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return nil, err
	}
	metrics = metrics.Normalize()
	if err := metrics.Validate(); err != nil {
		return nil, err
	}
	thresholds = thresholds.Normalize()
	currentHeight := metrics.Height
	alerts := []PaymentObservabilityAlert{}
	add := func(alertType PaymentObservabilityAlertType, severity PaymentObservabilityAlertSeverity, metricName string, value, threshold uint64, message string) {
		if value <= threshold {
			return
		}
		alert := PaymentObservabilityAlert{
			AlertType:	alertType,
			Severity:	severity,
			Height:		currentHeight,
			MetricName:	metricName,
			MetricValue:	value,
			Threshold:	threshold,
			Message:	message,
		}.WithHash()
		alerts = append(alerts, alert)
	}

	add(PaymentAlertHighPendingDisputeCount, PaymentAlertSeverityCritical, "active_disputes", metrics.ActiveDisputes, thresholds.HighPendingDisputeCount, "pending dispute count above configured threshold")
	add(PaymentAlertChallengeNearExpiryWithoutFinalization, PaymentAlertSeverityCritical, "challenge_period_near_expiry_count", metrics.ChallengePeriodNearExpiryCount, thresholds.NearExpiryWithoutFinalize, "challenge period near expiry without finalization")
	add(PaymentAlertFraudProofRejectionSpike, PaymentAlertSeverityWarning, "fraud_proofs_rejected", metrics.FraudProofsRejected, thresholds.FraudProofRejectionSpike, "fraud proof rejection spike above configured threshold")
	add(PaymentAlertChannelOpenSpamSpike, PaymentAlertSeverityWarning, "recent_channel_open_fees", recentChannelOpenFees(state, currentHeight, thresholds.WindowBlocks), thresholds.ChannelOpenSpamSpike, "channel open spam spike above configured threshold")
	add(PaymentAlertPromiseExpiryBacklog, PaymentAlertSeverityWarning, "promise_expiry_queue_backlog", promiseExpiryBacklog(state), thresholds.PromiseExpiryBacklog, "promise expiry backlog above configured threshold")
	add(PaymentAlertSettlementQueueBacklog, PaymentAlertSeverityCritical, "settlement_queue_backlog", settlementQueueBacklog(state), thresholds.SettlementQueueBacklog, "settlement queue backlog above configured threshold")
	add(PaymentAlertBlockSTMConflictRateAboveThreshold, PaymentAlertSeverityWarning, "blockstm_conflict_rate_bps", metrics.BlockSTMConflictRateBps, thresholds.BlockSTMConflictRateBps, "BlockSTM conflict rate above configured threshold")
	storeLatency := maxUint64(metrics.StoreV2PaymentModuleReadLatencyOps, metrics.StoreV2PaymentModuleWriteLatencyOps)
	add(PaymentAlertPaymentModuleStoreLatencyAboveThreshold, PaymentAlertSeverityWarning, "payment_module_store_latency_ops", storeLatency, thresholds.StoreLatencyOps, "payment module store latency above configured threshold")
	if signals.WatcherReplayHeight < currentHeight {
		add(PaymentAlertWatchServiceEventReplayLag, PaymentAlertSeverityCritical, "watcher_replay_lag_blocks", currentHeight-signals.WatcherReplayHeight, thresholds.WatchReplayLagBlocks, "watch service event replay lag above configured threshold")
	}
	gossipSpamRate := signals.Normalize().RoutingGossipSpamRateBps
	add(PaymentAlertRoutingGossipSpamRateAboveThreshold, PaymentAlertSeverityWarning, "routing_gossip_spam_rate_bps", gossipSpamRate, thresholds.RoutingGossipSpamRateBps, "routing gossip spam rate above configured threshold")

	sort.SliceStable(alerts, func(i, j int) bool {
		if alerts[i].AlertType == alerts[j].AlertType {
			return alerts[i].AlertID < alerts[j].AlertID
		}
		return alerts[i].AlertType < alerts[j].AlertType
	})
	for _, alert := range alerts {
		if err := alert.Validate(); err != nil {
			return nil, err
		}
	}
	return alerts, nil
}

func BuildPaymentObservabilityReports(state PaymentsState, fraud FraudProofVerificationState, metrics PaymentObservabilityMetrics, periodStartHeight, periodEndHeight uint64) ([]PaymentObservabilityReport, error) {
	if periodStartHeight == 0 {
		return nil, errors.New("payments observability report start height must be positive")
	}
	if periodEndHeight < periodStartHeight {
		return nil, errors.New("payments observability report end height must be >= start height")
	}
	state = state.Export()
	if err := state.Validate(); err != nil {
		return nil, err
	}
	fraud = fraud.Export()
	if err := fraud.Validate(); err != nil {
		return nil, err
	}
	metrics = metrics.Normalize()
	if err := metrics.Validate(); err != nil {
		return nil, err
	}
	layout, err := BuildStoreV2Layout(state)
	if err != nil {
		return nil, err
	}
	summary, err := buildPaymentObservabilityReportSummary(state, fraud, metrics, layout, periodStartHeight, periodEndHeight)
	if err != nil {
		return nil, err
	}
	reportTypes := []PaymentObservabilityReportType{
		PaymentReportDailyLockedLiquidity,
		PaymentReportDailySettlementVolume,
		PaymentReportDailyRoutingFee,
		PaymentReportDailyDisputeAndFraud,
		PaymentReportDailyStateFootprint,
		PaymentReportWeeklyChannelChurn,
		PaymentReportWeeklyLiquidityConcentration,
		PaymentReportWeeklyPerformance,
	}
	reports := make([]PaymentObservabilityReport, 0, len(reportTypes))
	for _, reportType := range reportTypes {
		report := summary
		report.ReportType = reportType
		report.ReportID = HashParts("payments-observability-report-id", string(reportType), fmt.Sprintf("%020d", periodStartHeight), fmt.Sprintf("%020d", periodEndHeight))
		report = report.WithHash()
		if err := report.Validate(); err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}

func ComputePaymentObservabilityAlertEvidenceHash(alert PaymentObservabilityAlert) string {
	alert = alert.Normalize()
	return HashParts(
		"payments-observability-alert-v1",
		string(alert.AlertType),
		string(alert.Severity),
		fmt.Sprintf("%020d", alert.Height),
		alert.MetricName,
		fmt.Sprintf("%020d", alert.MetricValue),
		fmt.Sprintf("%020d", alert.Threshold),
		alert.Message,
	)
}

func ComputePaymentObservabilityReportHash(report PaymentObservabilityReport) string {
	report = report.Normalize()
	return HashParts(
		"payments-observability-report-v1",
		report.ReportID,
		string(report.ReportType),
		fmt.Sprintf("%020d", report.PeriodStartHeight),
		fmt.Sprintf("%020d", report.PeriodEndHeight),
		report.TotalLockedNaet,
		report.SettlementVolumeNaet,
		report.RoutingFeesNaet,
		fmt.Sprintf("%020d", report.DisputeCount),
		fmt.Sprintf("%020d", report.FraudProofsSubmitted),
		fmt.Sprintf("%020d", report.FraudProofsAccepted),
		fmt.Sprintf("%020d", report.FraudProofsRejected),
		fmt.Sprintf("%020d", report.StateFootprintRecords),
		fmt.Sprintf("%020d", report.ChannelOpens),
		fmt.Sprintf("%020d", report.ChannelSettlements),
		fmt.Sprintf("%020d", report.ChannelChurnRateBps),
		report.LargestLockedChannelNaet,
		fmt.Sprintf("%020d", report.LiquidityConcentrationBps),
		fmt.Sprintf("%020d", report.BlockSTMConflictRateBps),
		fmt.Sprintf("%020d", report.StoreV2PaymentModuleLatencyOps),
		fmt.Sprintf("%020d", report.DisputeInclusionLatencyBlocks),
	)
}

func (t PaymentObservabilityAlertThresholds) Normalize() PaymentObservabilityAlertThresholds {
	if t.WindowBlocks == 0 {
		t.WindowBlocks = 1
	}
	return t
}

func (s PaymentObservabilityExternalSignals) Normalize() PaymentObservabilityExternalSignals {
	if s.RoutingGossipSpamRateBps == 0 && s.RoutingGossipMessages > 0 {
		s.RoutingGossipSpamRateBps = s.RoutingGossipRejected * 10_000 / s.RoutingGossipMessages
	}
	return s
}

func (a PaymentObservabilityAlert) WithHash() PaymentObservabilityAlert {
	a = a.Normalize()
	a.EvidenceHash = ComputePaymentObservabilityAlertEvidenceHash(a)
	a.AlertID = HashParts("payments-observability-alert-id", string(a.AlertType), fmt.Sprintf("%020d", a.Height), a.EvidenceHash)
	return a.Normalize()
}

func (a PaymentObservabilityAlert) Normalize() PaymentObservabilityAlert {
	a.AlertID = normalizeOptionalHash(a.AlertID)
	a.MetricName = strings.TrimSpace(a.MetricName)
	a.Message = strings.TrimSpace(a.Message)
	a.EvidenceHash = normalizeOptionalHash(a.EvidenceHash)
	return a
}

func (a PaymentObservabilityAlert) Validate() error {
	alert := a.Normalize()
	if !IsPaymentObservabilityAlertType(alert.AlertType) {
		return fmt.Errorf("unknown payments observability alert type %q", alert.AlertType)
	}
	if !IsPaymentObservabilityAlertSeverity(alert.Severity) {
		return fmt.Errorf("unknown payments observability alert severity %q", alert.Severity)
	}
	if alert.Height == 0 {
		return errors.New("payments observability alert height must be positive")
	}
	if alert.MetricName == "" {
		return errors.New("payments observability alert metric name is required")
	}
	if alert.Message == "" {
		return errors.New("payments observability alert message is required")
	}
	if alert.MetricValue <= alert.Threshold {
		return errors.New("payments observability alert metric must exceed threshold")
	}
	if err := ValidateHash("payments observability alert evidence hash", alert.EvidenceHash); err != nil {
		return err
	}
	if expected := ComputePaymentObservabilityAlertEvidenceHash(alert); alert.EvidenceHash != expected {
		return errors.New("payments observability alert evidence hash mismatch")
	}
	if err := ValidateHash("payments observability alert id", alert.AlertID); err != nil {
		return err
	}
	if expected := HashParts("payments-observability-alert-id", string(alert.AlertType), fmt.Sprintf("%020d", alert.Height), alert.EvidenceHash); alert.AlertID != expected {
		return errors.New("payments observability alert id mismatch")
	}
	return nil
}

func (r PaymentObservabilityReport) WithHash() PaymentObservabilityReport {
	r = r.Normalize()
	r.ReportHash = ComputePaymentObservabilityReportHash(r)
	return r.Normalize()
}

func (r PaymentObservabilityReport) Normalize() PaymentObservabilityReport {
	r.ReportID = normalizeOptionalHash(r.ReportID)
	r.TotalLockedNaet = normalizeAmountString(r.TotalLockedNaet)
	r.SettlementVolumeNaet = normalizeAmountString(r.SettlementVolumeNaet)
	r.RoutingFeesNaet = normalizeAmountString(r.RoutingFeesNaet)
	r.LargestLockedChannelNaet = normalizeAmountString(r.LargestLockedChannelNaet)
	r.ReportHash = normalizeOptionalHash(r.ReportHash)
	return r
}

func (r PaymentObservabilityReport) Validate() error {
	report := r.Normalize()
	if !IsPaymentObservabilityReportType(report.ReportType) {
		return fmt.Errorf("unknown payments observability report type %q", report.ReportType)
	}
	if report.PeriodStartHeight == 0 {
		return errors.New("payments observability report start height must be positive")
	}
	if report.PeriodEndHeight < report.PeriodStartHeight {
		return errors.New("payments observability report end height must be >= start height")
	}
	if err := ValidateHash("payments observability report id", report.ReportID); err != nil {
		return err
	}
	for label, value := range map[string]string{
		"total locked naet":		report.TotalLockedNaet,
		"settlement volume naet":	report.SettlementVolumeNaet,
		"routing fees naet":		report.RoutingFeesNaet,
		"largest locked channel naet":	report.LargestLockedChannelNaet,
	} {
		if err := validateNonNegativeInt("payments observability report "+label, value); err != nil {
			return err
		}
	}
	if report.LiquidityConcentrationBps > 10_000 {
		return errors.New("payments observability report liquidity concentration exceeds 100%")
	}
	if err := ValidateHash("payments observability report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputePaymentObservabilityReportHash(report); report.ReportHash != expected {
		return errors.New("payments observability report hash mismatch")
	}
	return nil
}

func IsPaymentObservabilityAlertType(alertType PaymentObservabilityAlertType) bool {
	switch alertType {
	case PaymentAlertHighPendingDisputeCount,
		PaymentAlertChallengeNearExpiryWithoutFinalization,
		PaymentAlertFraudProofRejectionSpike,
		PaymentAlertChannelOpenSpamSpike,
		PaymentAlertPromiseExpiryBacklog,
		PaymentAlertSettlementQueueBacklog,
		PaymentAlertBlockSTMConflictRateAboveThreshold,
		PaymentAlertPaymentModuleStoreLatencyAboveThreshold,
		PaymentAlertWatchServiceEventReplayLag,
		PaymentAlertRoutingGossipSpamRateAboveThreshold:
		return true
	default:
		return false
	}
}

func IsPaymentObservabilityAlertSeverity(severity PaymentObservabilityAlertSeverity) bool {
	switch severity {
	case PaymentAlertSeverityWarning, PaymentAlertSeverityCritical:
		return true
	default:
		return false
	}
}

func IsPaymentObservabilityReportType(reportType PaymentObservabilityReportType) bool {
	switch reportType {
	case PaymentReportDailyLockedLiquidity,
		PaymentReportDailySettlementVolume,
		PaymentReportDailyRoutingFee,
		PaymentReportDailyDisputeAndFraud,
		PaymentReportDailyStateFootprint,
		PaymentReportWeeklyChannelChurn,
		PaymentReportWeeklyLiquidityConcentration,
		PaymentReportWeeklyPerformance:
		return true
	default:
		return false
	}
}

func buildPaymentObservabilityReportSummary(state PaymentsState, fraud FraudProofVerificationState, metrics PaymentObservabilityMetrics, layout StoreV2Layout, startHeight, endHeight uint64) (PaymentObservabilityReport, error) {
	totalLocked, err := parseNonNegativeInt("payments observability report total locked", metrics.TotalLockedNaet)
	if err != nil {
		return PaymentObservabilityReport{}, err
	}
	settlementVolume := sdkmath.ZeroInt()
	routingFees := sdkmath.ZeroInt()
	channelSettlements := uint64(0)
	channelOpens := uint64(0)
	largestLocked := sdkmath.ZeroInt()
	for _, settlement := range state.Settlements {
		settlement = settlement.Normalize()
		if !heightInWindow(settlement.SettledHeight, startHeight, endHeight) {
			continue
		}
		channelSettlements++
		for _, balance := range settlement.FinalBalances {
			amount, err := parseNonNegativeInt("payments observability report settlement balance", strings.TrimSpace(balance.Amount))
			if err != nil {
				return PaymentObservabilityReport{}, err
			}
			settlementVolume = settlementVolume.Add(amount)
		}
	}
	for _, charge := range state.FeeCharges {
		charge = charge.Normalize()
		if !heightInWindow(charge.Height, startHeight, endHeight) {
			continue
		}
		if charge.FeeClass == PaymentFeeClassChannelOpen {
			channelOpens++
		}
		if charge.FeeClass == PaymentFeeClassRoutingAdvertisement || charge.FeeClass == PaymentFeeClassConditionalPromiseSettlement {
			amount, err := parseNonNegativeInt("payments observability report routing fee", charge.Amount)
			if err != nil {
				return PaymentObservabilityReport{}, err
			}
			routingFees = routingFees.Add(amount)
		}
	}
	for _, lock := range state.CustodyLocks {
		lock = lock.Normalize()
		if lock.Denom != NativeDenom {
			continue
		}
		amount, err := parseNonNegativeInt("payments observability report locked channel", lock.Amount)
		if err != nil {
			return PaymentObservabilityReport{}, err
		}
		if amount.GT(largestLocked) {
			largestLocked = amount
		}
	}
	churnRate := uint64(0)
	if channelOpens > 0 {
		churnRate = channelSettlements * 10_000 / channelOpens
	}
	concentration := uint64(0)
	if totalLocked.IsPositive() && largestLocked.IsPositive() {
		concentration = largestLocked.MulRaw(10_000).Quo(totalLocked).Uint64()
	}
	storeLatency := maxUint64(metrics.StoreV2PaymentModuleReadLatencyOps, metrics.StoreV2PaymentModuleWriteLatencyOps)
	return PaymentObservabilityReport{
		PeriodStartHeight:		startHeight,
		PeriodEndHeight:		endHeight,
		TotalLockedNaet:		totalLocked.String(),
		SettlementVolumeNaet:		settlementVolume.String(),
		RoutingFeesNaet:		routingFees.String(),
		DisputeCount:			metrics.ActiveDisputes,
		FraudProofsSubmitted:		metrics.FraudProofsSubmitted,
		FraudProofsAccepted:		uint64(len(fraud.EvidenceRecords)),
		FraudProofsRejected:		metrics.FraudProofsRejected,
		StateFootprintRecords:		paymentStateFootprintRecords(layout),
		ChannelOpens:			channelOpens,
		ChannelSettlements:		channelSettlements,
		ChannelChurnRateBps:		churnRate,
		LargestLockedChannelNaet:	largestLocked.String(),
		LiquidityConcentrationBps:	concentration,
		BlockSTMConflictRateBps:	metrics.BlockSTMConflictRateBps,
		StoreV2PaymentModuleLatencyOps:	storeLatency,
		DisputeInclusionLatencyBlocks:	metrics.DisputeInclusionLatency,
	}, nil
}

func recentChannelOpenFees(state PaymentsState, currentHeight, windowBlocks uint64) uint64 {
	count := uint64(0)
	for _, charge := range state.FeeCharges {
		charge = charge.Normalize()
		if charge.FeeClass == PaymentFeeClassChannelOpen && heightInRecentWindow(charge.Height, currentHeight, windowBlocks) {
			count++
		}
	}
	return count
}

func promiseExpiryBacklog(state PaymentsState) uint64 {
	count := uint64(0)
	for _, job := range state.AsyncPromiseExpiryQueue {
		if !job.Normalize().Completed {
			count++
		}
	}
	return count
}

func settlementQueueBacklog(state PaymentsState) uint64 {
	count := uint64(0)
	for _, job := range state.AsyncFinalizationQueue {
		if !job.Normalize().Completed {
			count++
		}
	}
	return count
}

func paymentStateFootprintRecords(layout StoreV2Layout) uint64 {
	layout = layout.Normalize()
	return uint64(len(layout.Channels) + len(layout.ChannelStates) + len(layout.PendingCloses) + len(layout.Conditions) + len(layout.VirtualChannels) + len(layout.ParticipantChannels) + len(layout.SettlementTombstones) + len(layout.FeeAccumulators) + len(layout.FraudProofs))
}

func heightInWindow(height, startHeight, endHeight uint64) bool {
	return height >= startHeight && height <= endHeight
}

func heightInRecentWindow(height, currentHeight, windowBlocks uint64) bool {
	if height == 0 || height > currentHeight {
		return false
	}
	if windowBlocks == 0 {
		windowBlocks = 1
	}
	startHeight := uint64(1)
	if currentHeight > windowBlocks {
		startHeight = currentHeight - windowBlocks + 1
	}
	return height >= startHeight
}

func maxUint64(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}

func normalizeAmountString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0"
	}
	return value
}
