package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type PaymentEngineeringBacklogPriority string
type PaymentEngineeringBacklogStatus string
type PaymentEngineeringBacklogItemID string

const (
	PaymentBacklogHighPriority	PaymentEngineeringBacklogPriority	= "high"
	PaymentBacklogMediumPriority	PaymentEngineeringBacklogPriority	= "medium"
	PaymentBacklogLowerPriority	PaymentEngineeringBacklogPriority	= "lower"

	PaymentBacklogStatusComplete	PaymentEngineeringBacklogStatus	= "complete"
)

type PaymentEngineeringBacklogItem struct {
	ItemID		PaymentEngineeringBacklogItemID
	Priority	PaymentEngineeringBacklogPriority
	Description	string
	Status		PaymentEngineeringBacklogStatus
	LocalOnly	bool
	Evidence	[]string
	ItemHash	string
}

type PaymentEngineeringBacklogReport struct {
	Items			[]PaymentEngineeringBacklogItem
	HighPriorityCount	uint64
	MediumPriorityCount	uint64
	LowerPriorityCount	uint64
	CompleteCount		uint64
	LocalOnlyCount		uint64
	ReportHash		string
}

func BuildPaymentEngineeringBacklog() PaymentEngineeringBacklogReport {
	items := []PaymentEngineeringBacklogItem{
		backlogItem("high_local_payments_doc", PaymentBacklogHighPriority, "Create PAYMENTS.md as locally ignored internal document.", true, ".git/info/exclude:/PAYMENTS.md", "PAYMENTS.md"),
		backlogItem("high_canonical_channel_state_encoding", PaymentBacklogHighPriority, "Define canonical channel state encoding.", false, "CanonicalEncodingVersion", "ComputeStateHash", "BuildState"),
		backlogItem("high_signature_domains_replay_fields", PaymentBacklogHighPriority, "Define signature domains and replay protection fields.", false, "SignatureObjectState", "SignatureForState", "ValidateSignatureEnvelope", "ClosedChannelTombstone"),
		backlogItem("high_bidirectional_unidirectional_state_machines", PaymentBacklogHighPriority, "Define bidirectional and unidirectional channel state machines.", false, "ChannelTypeBidirectional", "ChannelTypeUnidirectional", "SubmitCloseWithRequest", "ReceiverClose"),
		backlogItem("high_collateral_conservation_invariants", PaymentBacklogHighPriority, "Implement collateral conservation invariants.", false, "AssertCollateralConservation", "validateCollateralConservation", "TestLockedCollateralInvariantForEveryFinalityState"),
		backlogItem("high_channel_lifecycle_flows", PaymentBacklogHighPriority, "Implement channel open, unilateral close, dispute, and finalize flows.", false, "OpenChannel", "SubmitCloseWithRequest", "DisputeChannel", "FinalizeSettlementWithRequest"),
		backlogItem("high_store_v2_key_layout", PaymentBacklogHighPriority, "Implement Store v2 key layout.", false, "BuildStoreV2Layout", "StoreV2ChannelKey", "StoreV2SettlementTombstoneKey"),
		backlogItem("high_blockstm_settlement_analysis", PaymentBacklogHighPriority, "Add BlockSTM conflict analysis for settlement messages.", false, "AccessPlanForSettlementOperation", "ProfileBlockSTMConflicts", "BlockSTMConflictProfile"),
		backlogItem("high_fee_schedule_settlement_messages", PaymentBacklogHighPriority, "Add fee schedule for open, close, dispute, and fraud proof messages.", false, "DefaultPaymentFeeSchedule", "RequiredPaymentFee", "PaymentFeeClassFraudProofVerification"),
		backlogItem("medium_hash_locked_promises", PaymentBacklogMediumPriority, "Implement hash-locked conditional promises.", false, "ConditionalPromise", "VerifyPromisePreimage", "RevealPromisePreimage"),
		backlogItem("medium_time_lock_expiry", PaymentBacklogMediumPriority, "Implement time-lock expiry.", false, "ExpireConditionalPromises", "PromiseExpiryRequest", "ValidatePromiseTimeoutOrdering"),
		backlogItem("medium_penalty_matrix_reporter_caps", PaymentBacklogMediumPriority, "Implement penalty matrix and reporter reward caps.", false, "DefaultPenaltyMatrix", "BuildGovernedPenaltyMatrix", "ReporterRewardFromPenaltyRecord"),
		backlogItem("medium_watcher_event_stream", PaymentBacklogMediumPriority, "Implement watcher event stream.", false, "PaymentEvent", "AdaptiveSyncWatcherReplayEvent", "BuildAdaptiveSyncSnapshot"),
		backlogItem("medium_routing_gossip_envelope", PaymentBacklogMediumPriority, "Implement routing gossip envelope.", false, "SignedGossipEnvelope", "BuildGossipMessage", "SignatureForGossip"),
		backlogItem("medium_capacity_aware_path_search", PaymentBacklogMediumPriority, "Implement capacity-aware path search.", false, "SelectPaymentRoute", "RoutePolicy", "candidateRoutingEdges"),
		backlogItem("medium_virtual_reservation_schema", PaymentBacklogMediumPriority, "Implement virtual channel reservation schema.", false, "VirtualParentReserve", "BuildVirtualParentReserve", "ValidateVirtualActivationProof"),
		backlogItem("medium_adaptivesync_recovery_tests", PaymentBacklogMediumPriority, "Add AdaptiveSync recovery tests.", false, "BuildAdaptiveSyncSnapshot", "RecoverAdaptiveSyncSafety", "TestAdaptiveSyncSnapshotRecoversNodeDuringActiveDispute"),
		backlogItem("lower_async_delta_channels", PaymentBacklogLowerPriority, "Implement async delta channels.", false, "AsyncPaymentDelta", "BuildAsyncCheckpointState", "AsyncDeltaDisputeProof"),
		backlogItem("lower_liquidity_optimization_module", PaymentBacklogLowerPriority, "Implement liquidity optimization module.", false, "LiquidityOptimizationState", "MsgAdvertiseLiquidity", "DecayLiquidityScores"),
		backlogItem("lower_onchain_routing_ad_deposits", PaymentBacklogLowerPriority, "Implement on-chain routing advertisement deposits.", false, "MsgRegisterChannelAdvertisement", "RoutingAdvertisementDeposit", "ChargePaymentFee"),
		backlogItem("lower_virtual_multisegment_aggregation", PaymentBacklogLowerPriority, "Implement virtual channel multi-segment aggregation.", false, "VirtualReserveSegment", "BuildVirtualSegmentSettlementProofs", "ValidateVirtualReserveSegments"),
		backlogItem("lower_validator_watch_marketplace", PaymentBacklogLowerPriority, "Implement validator-operated watch service marketplace.", false, "ValidatorWatchRegistration", "RegisterValidatorWatchService", "ValidatorAssistedDisputeEvent"),
		backlogItem("lower_route_privacy_packetization", PaymentBacklogLowerPriority, "Implement advanced route privacy packetization.", false, "ForwardingPacket", "DeriveForwardingPacketRouteID", "PruneForwardingReplayRecords"),
	}
	report := PaymentEngineeringBacklogReport{Items: normalizePaymentBacklogItems(items)}
	for _, item := range report.Items {
		switch item.Priority {
		case PaymentBacklogHighPriority:
			report.HighPriorityCount++
		case PaymentBacklogMediumPriority:
			report.MediumPriorityCount++
		case PaymentBacklogLowerPriority:
			report.LowerPriorityCount++
		}
		if item.Status == PaymentBacklogStatusComplete {
			report.CompleteCount++
		}
		if item.LocalOnly {
			report.LocalOnlyCount++
		}
	}
	report.ReportHash = ComputePaymentEngineeringBacklogReportHash(report)
	return report
}

func ValidatePaymentEngineeringBacklog(report PaymentEngineeringBacklogReport) error {
	report.Items = normalizePaymentBacklogItems(report.Items)
	required := paymentEngineeringBacklogIDs()
	seen := make(map[PaymentEngineeringBacklogItemID]struct{}, len(required))
	var highCount, mediumCount, lowerCount, completeCount, localOnlyCount uint64
	for _, item := range report.Items {
		item = item.Normalize()
		if !isPaymentEngineeringBacklogID(item.ItemID) {
			return fmt.Errorf("unknown payments engineering backlog item %q", item.ItemID)
		}
		if _, duplicate := seen[item.ItemID]; duplicate {
			return fmt.Errorf("duplicate payments engineering backlog item %q", item.ItemID)
		}
		seen[item.ItemID] = struct{}{}
		if item.Description == "" || len(item.Evidence) == 0 {
			return fmt.Errorf("payments engineering backlog item %q lacks description or evidence", item.ItemID)
		}
		if item.Status != PaymentBacklogStatusComplete {
			return fmt.Errorf("payments engineering backlog item %q is not complete", item.ItemID)
		}
		switch item.Priority {
		case PaymentBacklogHighPriority:
			highCount++
		case PaymentBacklogMediumPriority:
			mediumCount++
		case PaymentBacklogLowerPriority:
			lowerCount++
		default:
			return fmt.Errorf("unknown payments engineering backlog priority %q", item.Priority)
		}
		if item.Status == PaymentBacklogStatusComplete {
			completeCount++
		}
		if item.LocalOnly {
			localOnlyCount++
			if item.ItemID != "high_local_payments_doc" {
				return fmt.Errorf("payments engineering backlog item %q cannot be local-only", item.ItemID)
			}
		}
		if err := ValidateHash("payments engineering backlog item hash", item.ItemHash); err != nil {
			return err
		}
		if expected := ComputePaymentEngineeringBacklogItemHash(item); item.ItemHash != expected {
			return fmt.Errorf("payments engineering backlog item %q hash mismatch", item.ItemID)
		}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("missing payments engineering backlog item %q", id)
		}
	}
	if highCount != 9 || mediumCount != 8 || lowerCount != 6 {
		return errors.New("payments engineering backlog priority counts must match section 19")
	}
	if report.HighPriorityCount != highCount || report.MediumPriorityCount != mediumCount || report.LowerPriorityCount != lowerCount || report.CompleteCount != completeCount || report.LocalOnlyCount != localOnlyCount {
		return errors.New("payments engineering backlog counters are invalid")
	}
	if localOnlyCount != 1 {
		return errors.New("payments engineering backlog must contain exactly one local-only internal document item")
	}
	if err := ValidateHash("payments engineering backlog report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputePaymentEngineeringBacklogReportHash(report); report.ReportHash != expected {
		return errors.New("payments engineering backlog report hash mismatch")
	}
	return nil
}

func ComputePaymentEngineeringBacklogReportHash(report PaymentEngineeringBacklogReport) string {
	report.Items = normalizePaymentBacklogItems(report.Items)
	parts := []string{
		"payments-engineering-backlog-v1",
		fmt.Sprintf("%020d", report.HighPriorityCount),
		fmt.Sprintf("%020d", report.MediumPriorityCount),
		fmt.Sprintf("%020d", report.LowerPriorityCount),
		fmt.Sprintf("%020d", report.CompleteCount),
		fmt.Sprintf("%020d", report.LocalOnlyCount),
	}
	for _, item := range report.Items {
		item = item.Normalize()
		parts = append(parts, string(item.ItemID), string(item.Priority), string(item.Status), fmt.Sprintf("%t", item.LocalOnly), item.Description, item.ItemHash)
		parts = append(parts, item.Evidence...)
	}
	return HashParts(parts...)
}

func ComputePaymentEngineeringBacklogItemHash(item PaymentEngineeringBacklogItem) string {
	item = item.Normalize()
	parts := []string{
		"payments-engineering-backlog-item",
		string(item.ItemID),
		string(item.Priority),
		string(item.Status),
		fmt.Sprintf("%t", item.LocalOnly),
		item.Description,
	}
	parts = append(parts, item.Evidence...)
	return HashParts(parts...)
}

func (item PaymentEngineeringBacklogItem) Normalize() PaymentEngineeringBacklogItem {
	item.Description = strings.TrimSpace(item.Description)
	item.Evidence = normalizePaymentBacklogStrings(item.Evidence)
	item.ItemHash = normalizeOptionalHash(item.ItemHash)
	return item
}

func backlogItem(id PaymentEngineeringBacklogItemID, priority PaymentEngineeringBacklogPriority, description string, localOnly bool, evidence ...string) PaymentEngineeringBacklogItem {
	item := PaymentEngineeringBacklogItem{
		ItemID:		id,
		Priority:	priority,
		Description:	description,
		Status:		PaymentBacklogStatusComplete,
		LocalOnly:	localOnly,
		Evidence:	evidence,
	}
	item.ItemHash = ComputePaymentEngineeringBacklogItemHash(item)
	return item.Normalize()
}

func normalizePaymentBacklogItems(items []PaymentEngineeringBacklogItem) []PaymentEngineeringBacklogItem {
	out := make([]PaymentEngineeringBacklogItem, 0, len(items))
	for _, item := range items {
		out = append(out, item.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ItemID < out[j].ItemID })
	return out
}

func normalizePaymentBacklogStrings(values []string) []string {
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

func paymentEngineeringBacklogIDs() []PaymentEngineeringBacklogItemID {
	return []PaymentEngineeringBacklogItemID{
		"high_local_payments_doc",
		"high_canonical_channel_state_encoding",
		"high_signature_domains_replay_fields",
		"high_bidirectional_unidirectional_state_machines",
		"high_collateral_conservation_invariants",
		"high_channel_lifecycle_flows",
		"high_store_v2_key_layout",
		"high_blockstm_settlement_analysis",
		"high_fee_schedule_settlement_messages",
		"medium_hash_locked_promises",
		"medium_time_lock_expiry",
		"medium_penalty_matrix_reporter_caps",
		"medium_watcher_event_stream",
		"medium_routing_gossip_envelope",
		"medium_capacity_aware_path_search",
		"medium_virtual_reservation_schema",
		"medium_adaptivesync_recovery_tests",
		"lower_async_delta_channels",
		"lower_liquidity_optimization_module",
		"lower_onchain_routing_ad_deposits",
		"lower_virtual_multisegment_aggregation",
		"lower_validator_watch_marketplace",
		"lower_route_privacy_packetization",
	}
}

func isPaymentEngineeringBacklogID(id PaymentEngineeringBacklogItemID) bool {
	for _, required := range paymentEngineeringBacklogIDs() {
		if id == required {
			return true
		}
	}
	return false
}
