package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type RequiredTestCoverageKind string
type RequiredTestCoverageID string

const (
	RequiredTestCoverageUnit	RequiredTestCoverageKind	= "UNIT"
	RequiredTestCoverageIntegration	RequiredTestCoverageKind	= "INTEGRATION"
	RequiredTestCoverageInvariant	RequiredTestCoverageKind	= "INVARIANT"
	RequiredTestCoverageFuzz	RequiredTestCoverageKind	= "FUZZ"
	RequiredTestCoveragePerformance	RequiredTestCoverageKind	= "PERFORMANCE"
)

type RequiredTestCoverageEntry struct {
	CoverageID	RequiredTestCoverageID
	Kind		RequiredTestCoverageKind
	Description	string
	TestNames	[]string
	Evidence	[]string
	EvidenceHash	string
}

type RequiredTestCoverageReport struct {
	UnitCount		uint64
	IntegrationCount	uint64
	InvariantCount		uint64
	FuzzCount		uint64
	PerformanceCount	uint64
	Entries			[]RequiredTestCoverageEntry
	ReportHash		string
}

func BuildRequiredTestCoverageReport() RequiredTestCoverageReport {
	entries := []RequiredTestCoverageEntry{
		requiredUnitCoverage("unit_channel_id_generation", "Channel ID generation.", []string{"TestRequiredPaymentTestCoverageMatrixCoversUnitAndIntegrationSpecs"}, "HashParts", "signedChannel"),
		requiredUnitCoverage("unit_state_hash_encoding", "State hash encoding.", []string{"TestStateHashEncodingVersionAndDomainSeparation", "TestCanonicalChannelStateIncludesAllStateDomains"}, "ComputeStateHash", "BuildState"),
		requiredUnitCoverage("unit_signature_domain_validation", "Signature domain validation.", []string{"TestSignatureEnvelopeRejectsReplayAndWrongCommitment", "TestClaimAndDeltaSignatureEnvelopeValidation"}, "ComputeSignatureEnvelopeHash", "ValidateSignatureEnvelope"),
		requiredUnitCoverage("unit_balance_conservation", "Balance conservation.", []string{"TestPaymentStateRejectsNonNaetAndCollateralMismatch", "TestLockedCollateralInvariantForEveryFinalityState"}, "validateCollateralConservation", "AssertCollateralConservation"),
		requiredUnitCoverage("unit_nonce_monotonicity", "Nonce monotonicity.", []string{"TestBidirectionalCloseAndUpdateRules", "TestRollbackVectorsRejectNonceAndPreviousHashRollback"}, "AcceptSignedState", "ValidatePreviousHashContinuity"),
		requiredUnitCoverage("unit_cooperative_close", "Cooperative close.", []string{"TestPaymentChannelCloseDisputeFraudAndSettlement", "TestPaymentAPISurfaceMessagesQueriesAndSettlementViews"}, "CooperativeClose", "MsgCooperativeClose"),
		requiredUnitCoverage("unit_unilateral_close", "Unilateral close.", []string{"TestUnilateralCloseRequestStoresReasonAndDetachedSignatures", "TestPaymentChannelCloseDisputeFraudAndSettlement"}, "SubmitCloseWithRequest", "MsgUnilateralClose"),
		requiredUnitCoverage("unit_dispute_supersession", "Dispute supersession.", []string{"TestDisputeRequestEmitsEventAndAppliesOptionalFraudProof", "TestWatchServiceSubmitsStaleCloseDispute"}, "DisputeChannel", "MsgDisputeClose"),
		requiredUnitCoverage("unit_final_settlement", "Final settlement.", []string{"TestFinalSettlementRequiresResolvedConditionsAndUnlocksCustody", "TestPaymentChannelCloseDisputeFraudAndSettlement"}, "FinalizeSettlementWithRequest", "MsgFinalizeClose"),
		requiredUnitCoverage("unit_tombstone_replay_rejection", "Tombstone replay rejection.", []string{"TestPaymentChannelCloseDisputeFraudAndSettlement", "TestStoreV2PrunesExpiredTombstonesAndConditions"}, "SettlementTombstone", "RejectEarlyTombstonePruning"),
		requiredUnitCoverage("unit_hash_lock_proof_validation", "Hash-lock proof validation.", []string{"TestHashLockedPreimageRevealResolvesLinkedPromisesAndTracksPreimage", "TestConditionalPaymentsModuleMessagesRootsClaimsAndDisputes"}, "VerifyPromisePreimage", "RevealPromisePreimage"),
		requiredUnitCoverage("unit_time_lock_expiry", "Time-lock expiry.", []string{"TestTimeoutOrderingAndExpiryResolutionReleaseConditionRoot", "TestAsyncExecutionExpiredPromiseQueueIsBoundedAndRetriable"}, "ExpireConditionalPromises", "ValidatePromiseTimeoutOrdering"),
		requiredUnitCoverage("unit_penalty_routing", "Penalty routing.", []string{"TestPenaltyMatrixCoversFraudProofCategoriesAndBoundsBalances", "TestFraudProofInvalidBalanceRoutesPenaltyRemainder"}, "BuildFraudPenaltyRouting", "BuildPenaltyRouteAccounting"),
		requiredUnitCoverage("unit_fee_calculation", "Fee calculation.", []string{"TestPaymentFeeScheduleChargesStorageAndDynamicMultiplier", "TestChannelOpenFeeFormulaComponentsAndBounds"}, "RequiredPaymentFee", "ChargePaymentFee"),
		requiredIntegrationCoverage("integration_bidirectional_lifecycle", "Bidirectional channel open, update, close, settle.", []string{"TestPaymentChannelCloseDisputeFraudAndSettlement", "TestChannelUpdateLifecycleValidatesOffchainAndRegistersCheckpoint"}, "OpenChannel", "AcceptSignedState", "CooperativeClose"),
		requiredIntegrationCoverage("integration_unidirectional_streaming", "Unidirectional streaming claim and reclaim.", []string{"TestUnidirectionalReceiverCloseUsesSinglePayerSignature", "TestUnidirectionalAcknowledgementModeAndPayerReclaim", "TestUnidirectionalStreamingPaymentHelperFormat"}, "StreamingClaimForChannel", "ReceiverClose", "PayerReclaim"),
		requiredIntegrationCoverage("integration_async_checkpoint_dispute", "Async delta checkpoint and dispute.", []string{"TestAsyncUpdateBatchCanRegisterCheckpoint", "TestAsyncCheckpointAggregationExposureExpiryAndProof"}, "BuildAsyncCheckpointState", "AsyncDeltaDisputeProof"),
		requiredIntegrationCoverage("integration_multihop_conditional_payment", "Multi-hop conditional payment.", []string{"TestBatchConditionSettlementAtomicallyResolvesChainedPromises", "TestBatchConditionSettlementRejectsBrokenRouteInvariants"}, "BatchSettleLinkedPromises", "ConditionLinkageProof"),
		requiredIntegrationCoverage("integration_virtual_channel_lifecycle", "Virtual channel open, update, close.", []string{"TestVirtualChannelOpeningRequiresReservationProofAndRouteTimeout", "TestVirtualChannelEndpointUpdatesAndDisputeProof", "TestVirtualChannelCloseProofModesAndTimeoutHierarchy"}, "OpenVirtualChannelWithProof", "AcceptVirtualChannelUpdate", "CloseVirtualChannelWithProof"),
		requiredIntegrationCoverage("integration_parent_dispute_with_virtual_active", "Parent channel dispute while virtual channel is active.", []string{"TestParentChannelDisputeWhileVirtualChannelIsActive"}, "SubmitClose", "DisputeClose", "VirtualChannel"),
		requiredIntegrationCoverage("integration_fraud_reward", "Fraud proof with reporter reward.", []string{"TestFraudProofVerificationFeeRefundsWhenAccepted", "TestFraudProofVerificationModuleDedupGasPenaltyAndRewardClaim"}, "SubmitFraudProofWithPolicy", "ReporterRewardFromPenaltyRecord"),
		requiredIntegrationCoverage("integration_fee_congestion_during_dispute", "Fee congestion during dispute.", []string{"TestDisputePriorityPolicyNearExpiryFraudAndStressInclusion", "TestPaymentFeeScheduleChargesStorageAndDynamicMultiplier"}, "ComputeDisputeTransactionPriority", "DynamicFeeMultiplier"),
		requiredIntegrationCoverage("integration_store_snapshot_pending_close", "Store snapshot recovery during pending close.", []string{"TestAdaptiveSyncSnapshotRecoversNodeDuringActiveDispute", "TestKeeperAdaptiveSyncSnapshotRecovery"}, "BuildAdaptiveSyncSnapshot", "RecoverAdaptiveSyncSafety"),
		requiredInvariantCoverage("invariant_locked_collateral_active_balances_reserves_penalties", "Locked collateral equals active balances plus reserves plus pending penalties.", []string{"TestLockedCollateralInvariantForEveryFinalityState", "TestDisputeAndPenaltyFinalityTransitionsRetainCollateralUntilSettlement"}, "ValidateLockedCollateralForFinality", "validateCollateralConservation"),
		requiredInvariantCoverage("invariant_settlement_never_overpays_locked_collateral", "Settlement never pays more than locked collateral.", []string{"TestFinalSettlementRequiresResolvedConditionsAndUnlocksCustody", "TestPaymentChannelCloseDisputeFraudAndSettlement"}, "SettlementRecord.ValidateForChannel", "applySettlementAdjustments"),
		requiredInvariantCoverage("invariant_promise_reserve_within_channel_balance", "Promise reserve cannot exceed channel balance.", []string{"TestConditionalPromiseObjectSignatureReserveAndReplayRules", "TestKeeperConditionalPaymentsModuleMessagesAndInvariants"}, "ValidateReservedBalancesForConditions", "BuildConditionRootUpdateFromPromises"),
		requiredInvariantCoverage("invariant_expired_promises_not_resolvable", "Expired promises cannot be resolved.", []string{"TestTimeoutOrderingAndExpiryResolutionReleaseConditionRoot", "FuzzPaymentRequiredFuzzVectors"}, "RevealPromisePreimage", "ConditionalPromise.TimeoutHeight"),
		requiredInvariantCoverage("invariant_preimage_single_settlement_per_promise", "Same preimage cannot settle the same promise twice.", []string{"TestHashLockedPreimageRevealResolvesLinkedPromisesAndTracksPreimage", "TestSettlementRejectsReusedConditionAndPreimageClaims"}, "ConditionClaimRecord", "rejectReusedConditionClaims"),
		requiredInvariantCoverage("invariant_finalized_channel_id_cannot_reopen", "Finalized channel cannot be reopened with same ID.", []string{"TestPaymentChannelCloseDisputeFraudAndSettlement", "TestPaymentChannelModuleMessagesDispatchAnteAndInvariants"}, "OpenChannel", "SettlementTombstone"),
		requiredInvariantCoverage("invariant_penalties_non_negative_balances", "Penalties cannot produce negative balances.", []string{"TestPenaltyMatrixCoversFraudProofCategoriesAndBoundsBalances", "TestFraudProofInvalidBalanceRoutesPenaltyRemainder"}, "ValidatePenaltyWithinAvailableBalance", "applySettlementAdjustments"),
		requiredInvariantCoverage("invariant_fee_buckets_sum_collected_fees", "Fee buckets sum to collected fees.", []string{"TestPaymentBlockAccumulatorAggregatesAfterSettlementHotPath", "TestPaymentChannelModuleMessagesDispatchAnteAndInvariants"}, "AccumulatePaymentBlockAccounting", "ChannelFeeAccumulatorFromBlock"),
		requiredInvariantCoverage("invariant_same_channel_writes_conflict", "Same-channel writes conflict deterministically.", []string{"TestBlockSTMConflictProfileDetectsSameChannelConflicts", "TestPaymentChannelModuleBlockSTMProfilesMessageConflicts"}, "ProfileBlockSTMConflicts", "PaymentChannelMessageAccessPlan"),
		requiredInvariantCoverage("invariant_distinct_channel_settlements_hot_keys", "Distinct-channel settlements do not share hot write keys.", []string{"TestBlockSTMAccessPlanUsesPerChannelKeysAndDefersAccounting", "TestSettlementBatchRequiresIndependentChannels"}, "AccessPlanForSettlementOperation", "PaymentBlockAccumulatorKey"),
		requiredFuzzCoverage("fuzz_malformed_signed_states", "Malformed signed states.", []string{"FuzzPaymentRequiredFuzzVectors", "TestSignatureEnvelopeRejectsReplayAndWrongCommitment"}, "ChannelState.ValidateForChannel", "ComputeStateSignaturePreimageHash"),
		requiredFuzzCoverage("fuzz_random_nonce_ordering", "Random nonce ordering.", []string{"FuzzPaymentRequiredFuzzVectors", "TestRollbackVectorsRejectNonceAndPreviousHashRollback"}, "AcceptSignedState", "ValidatePreviousHashContinuity"),
		requiredFuzzCoverage("fuzz_conflicting_same_nonce_states", "Conflicting same-nonce states.", []string{"FuzzPaymentRequiredFuzzVectors", "FuzzCanonicalFraudEvidenceHashMalformedInputs"}, "FraudProofTypeDoubleSign", "ComputeCanonicalFraudEvidenceHash"),
		requiredFuzzCoverage("fuzz_invalid_promise_links", "Invalid promise links.", []string{"FuzzPaymentRequiredFuzzVectors", "TestBatchConditionSettlementRejectsBrokenRouteInvariants"}, "ConditionLinkageProof", "BatchSettleLinkedPromises"),
		requiredFuzzCoverage("fuzz_timeout_boundary_conditions", "Timeout boundary conditions.", []string{"FuzzPaymentRequiredFuzzVectors", "TestTimeoutOrderingAndExpiryResolutionReleaseConditionRoot"}, "ValidatePromiseTimeoutOrdering", "RevealPromisePreimage"),
		requiredFuzzCoverage("fuzz_batch_settlement_ordering", "Batch settlement ordering.", []string{"FuzzPaymentRequiredFuzzVectors", "TestSettlementBatchGroupingByChannelKey"}, "NewSettlementBatch", "ComputeBatchRoot"),
		requiredFuzzCoverage("fuzz_fraud_proof_duplicate_encodings", "Fraud proof duplicate encodings.", []string{"FuzzPaymentRequiredFuzzVectors", "TestFraudProofVerificationModuleDedupGasPenaltyAndRewardClaim"}, "ComputeCanonicalFraudEvidenceHash", "FraudProofVerificationState.HasEvidence"),
		requiredFuzzCoverage("fuzz_route_failure_classifications", "Route failure classifications.", []string{"FuzzPaymentRequiredFuzzVectors", "TestRouteFailureScoringReducesLocalRoutingScore"}, "ClassifyRouteFailure", "BuildRouteFailureScore"),
		requiredFuzzCoverage("fuzz_async_delta_aggregation", "Async delta aggregation.", []string{"FuzzPaymentRequiredFuzzVectors", "TestAsyncCheckpointAggregationExposureExpiryAndProof"}, "BuildAsyncCheckpointState", "ComputeAsyncDeltaRootForChannel"),
		requiredPerformanceCoverage("performance_channel_opens_per_block", "Channel opens per block.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestPaymentChannelModuleMessagesDispatchAnteAndInvariants"}, "MsgOpenChannel", "BlockSTMClassOpenChannel"),
		requiredPerformanceCoverage("performance_cooperative_closes_per_block", "Cooperative closes per block.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestPaymentChannelCloseDisputeFraudAndSettlement"}, "CooperativeClose", "PaymentFeeClassCooperativeClose"),
		requiredPerformanceCoverage("performance_unilateral_closes_per_block", "Unilateral closes per block.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestUnilateralCloseRequestStoresReasonAndDetachedSignatures"}, "SubmitCloseWithRequest", "BlockSTMClassCloseChannel"),
		requiredPerformanceCoverage("performance_disputes_per_block", "Disputes per block.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestDisputeRequestEmitsEventAndAppliesOptionalFraudProof"}, "DisputeChannel", "BlockSTMClassDisputeChannel"),
		requiredPerformanceCoverage("performance_promise_resolutions_per_block", "Promise resolutions per block.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestBatchConditionSettlementAtomicallyResolvesChainedPromises"}, "RevealPromisePreimage", "AccessPlanForConditionResolution"),
		requiredPerformanceCoverage("performance_virtual_channel_disputes_per_block", "Virtual channel disputes per block.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestVirtualChannelEndpointUpdatesAndDisputeProof"}, "BuildVirtualChannelDisputeProof", "SubmitVirtualChannelDispute"),
		requiredPerformanceCoverage("performance_blockstm_conflict_rate_mix", "BlockSTM conflict rate by transaction mix.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestPaymentChannelModuleBlockSTMProfilesMessageConflicts"}, "ProfileBlockSTMConflicts", "BlockSTMConflictProfile"),
		requiredPerformanceCoverage("performance_store_v2_index_latency", "Store v2 read/write latency for channel indexes.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestStoreV2ParticipantIndexPagination"}, "BuildStoreV2Layout", "QueryStoreV2ParticipantChannels"),
		requiredPerformanceCoverage("performance_snapshot_recovery_active_disputes", "Snapshot recovery time with active disputes.", []string{"TestPaymentPerformanceCoverageProfilesPerBlockWorkloads", "TestAdaptiveSyncSnapshotRecoversNodeDuringActiveDispute"}, "BuildAdaptiveSyncSnapshot", "RecoverAdaptiveSyncSafety"),
	}
	report := RequiredTestCoverageReport{Entries: normalizeRequiredTestCoverageEntries(entries)}
	for _, entry := range report.Entries {
		switch entry.Kind {
		case RequiredTestCoverageUnit:
			report.UnitCount++
		case RequiredTestCoverageIntegration:
			report.IntegrationCount++
		case RequiredTestCoverageInvariant:
			report.InvariantCount++
		case RequiredTestCoverageFuzz:
			report.FuzzCount++
		case RequiredTestCoveragePerformance:
			report.PerformanceCount++
		}
	}
	report.ReportHash = ComputeRequiredTestCoverageReportHash(report)
	return report
}

func ValidateRequiredTestCoverageReport(report RequiredTestCoverageReport) error {
	report.Entries = normalizeRequiredTestCoverageEntries(report.Entries)
	required := requiredTestCoverageIDs()
	seen := make(map[RequiredTestCoverageID]struct{}, len(required))
	unitCount := uint64(0)
	integrationCount := uint64(0)
	invariantCount := uint64(0)
	fuzzCount := uint64(0)
	performanceCount := uint64(0)
	for _, entry := range report.Entries {
		entry = entry.Normalize()
		if !isRequiredTestCoverageID(entry.CoverageID) {
			return fmt.Errorf("unknown payments required test coverage %q", entry.CoverageID)
		}
		if _, duplicate := seen[entry.CoverageID]; duplicate {
			return fmt.Errorf("duplicate payments required test coverage %q", entry.CoverageID)
		}
		seen[entry.CoverageID] = struct{}{}
		if entry.Kind != RequiredTestCoverageUnit && entry.Kind != RequiredTestCoverageIntegration && entry.Kind != RequiredTestCoverageInvariant && entry.Kind != RequiredTestCoverageFuzz && entry.Kind != RequiredTestCoveragePerformance {
			return fmt.Errorf("unknown payments required test coverage kind %q", entry.Kind)
		}
		switch entry.Kind {
		case RequiredTestCoverageUnit:
			unitCount++
		case RequiredTestCoverageIntegration:
			integrationCount++
		case RequiredTestCoverageInvariant:
			invariantCount++
		case RequiredTestCoverageFuzz:
			fuzzCount++
		case RequiredTestCoveragePerformance:
			performanceCount++
		}
		if entry.Description == "" || len(entry.TestNames) == 0 || len(entry.Evidence) == 0 {
			return fmt.Errorf("payments required test coverage %q lacks description, test names, or evidence", entry.CoverageID)
		}
		if err := ValidateHash("payments required test coverage evidence hash", entry.EvidenceHash); err != nil {
			return err
		}
		if expected := ComputeRequiredTestCoverageEntryHash(entry); entry.EvidenceHash != expected {
			return fmt.Errorf("payments required test coverage %q evidence hash mismatch", entry.CoverageID)
		}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("missing payments required test coverage %q", id)
		}
	}
	if unitCount != 14 || integrationCount != 9 || invariantCount != 10 || fuzzCount != 9 || performanceCount != 9 {
		return errors.New("payments required test coverage counts must match section 16.1 through 16.5")
	}
	if report.UnitCount != unitCount || report.IntegrationCount != integrationCount || report.InvariantCount != invariantCount || report.FuzzCount != fuzzCount || report.PerformanceCount != performanceCount {
		return errors.New("payments required test coverage counters are invalid")
	}
	if err := ValidateHash("payments required test coverage report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputeRequiredTestCoverageReportHash(report); report.ReportHash != expected {
		return errors.New("payments required test coverage report hash mismatch")
	}
	return nil
}

func ComputeRequiredTestCoverageReportHash(report RequiredTestCoverageReport) string {
	report.Entries = normalizeRequiredTestCoverageEntries(report.Entries)
	parts := []string{"payments-required-test-coverage-v1", fmt.Sprintf("%020d", report.UnitCount), fmt.Sprintf("%020d", report.IntegrationCount), fmt.Sprintf("%020d", report.InvariantCount), fmt.Sprintf("%020d", report.FuzzCount), fmt.Sprintf("%020d", report.PerformanceCount)}
	for _, entry := range report.Entries {
		entry = entry.Normalize()
		parts = append(parts, string(entry.CoverageID), string(entry.Kind), entry.Description, entry.EvidenceHash)
		parts = append(parts, entry.TestNames...)
		parts = append(parts, entry.Evidence...)
	}
	return HashParts(parts...)
}

func ComputeRequiredTestCoverageEntryHash(entry RequiredTestCoverageEntry) string {
	entry = entry.Normalize()
	parts := []string{"payments-required-test-coverage-entry", string(entry.CoverageID), string(entry.Kind), entry.Description}
	parts = append(parts, entry.TestNames...)
	parts = append(parts, entry.Evidence...)
	return HashParts(parts...)
}

func (entry RequiredTestCoverageEntry) Normalize() RequiredTestCoverageEntry {
	entry.Description = strings.TrimSpace(entry.Description)
	entry.TestNames = normalizeRequiredTestCoverageStrings(entry.TestNames)
	entry.Evidence = normalizeRequiredTestCoverageStrings(entry.Evidence)
	entry.EvidenceHash = normalizeOptionalHash(entry.EvidenceHash)
	return entry
}

func requiredUnitCoverage(id RequiredTestCoverageID, description string, tests []string, evidence ...string) RequiredTestCoverageEntry {
	return requiredTestCoverage(id, RequiredTestCoverageUnit, description, tests, evidence...)
}

func requiredIntegrationCoverage(id RequiredTestCoverageID, description string, tests []string, evidence ...string) RequiredTestCoverageEntry {
	return requiredTestCoverage(id, RequiredTestCoverageIntegration, description, tests, evidence...)
}

func requiredInvariantCoverage(id RequiredTestCoverageID, description string, tests []string, evidence ...string) RequiredTestCoverageEntry {
	return requiredTestCoverage(id, RequiredTestCoverageInvariant, description, tests, evidence...)
}

func requiredFuzzCoverage(id RequiredTestCoverageID, description string, tests []string, evidence ...string) RequiredTestCoverageEntry {
	return requiredTestCoverage(id, RequiredTestCoverageFuzz, description, tests, evidence...)
}

func requiredPerformanceCoverage(id RequiredTestCoverageID, description string, tests []string, evidence ...string) RequiredTestCoverageEntry {
	return requiredTestCoverage(id, RequiredTestCoveragePerformance, description, tests, evidence...)
}

func requiredTestCoverage(id RequiredTestCoverageID, kind RequiredTestCoverageKind, description string, tests []string, evidence ...string) RequiredTestCoverageEntry {
	entry := RequiredTestCoverageEntry{
		CoverageID:	id,
		Kind:		kind,
		Description:	description,
		TestNames:	tests,
		Evidence:	evidence,
	}
	entry.EvidenceHash = ComputeRequiredTestCoverageEntryHash(entry)
	return entry.Normalize()
}

func normalizeRequiredTestCoverageEntries(entries []RequiredTestCoverageEntry) []RequiredTestCoverageEntry {
	out := make([]RequiredTestCoverageEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CoverageID < out[j].CoverageID })
	return out
}

func normalizeRequiredTestCoverageStrings(values []string) []string {
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

func requiredTestCoverageIDs() []RequiredTestCoverageID {
	return []RequiredTestCoverageID{
		"unit_channel_id_generation",
		"unit_state_hash_encoding",
		"unit_signature_domain_validation",
		"unit_balance_conservation",
		"unit_nonce_monotonicity",
		"unit_cooperative_close",
		"unit_unilateral_close",
		"unit_dispute_supersession",
		"unit_final_settlement",
		"unit_tombstone_replay_rejection",
		"unit_hash_lock_proof_validation",
		"unit_time_lock_expiry",
		"unit_penalty_routing",
		"unit_fee_calculation",
		"integration_bidirectional_lifecycle",
		"integration_unidirectional_streaming",
		"integration_async_checkpoint_dispute",
		"integration_multihop_conditional_payment",
		"integration_virtual_channel_lifecycle",
		"integration_parent_dispute_with_virtual_active",
		"integration_fraud_reward",
		"integration_fee_congestion_during_dispute",
		"integration_store_snapshot_pending_close",
		"invariant_locked_collateral_active_balances_reserves_penalties",
		"invariant_settlement_never_overpays_locked_collateral",
		"invariant_promise_reserve_within_channel_balance",
		"invariant_expired_promises_not_resolvable",
		"invariant_preimage_single_settlement_per_promise",
		"invariant_finalized_channel_id_cannot_reopen",
		"invariant_penalties_non_negative_balances",
		"invariant_fee_buckets_sum_collected_fees",
		"invariant_same_channel_writes_conflict",
		"invariant_distinct_channel_settlements_hot_keys",
		"fuzz_malformed_signed_states",
		"fuzz_random_nonce_ordering",
		"fuzz_conflicting_same_nonce_states",
		"fuzz_invalid_promise_links",
		"fuzz_timeout_boundary_conditions",
		"fuzz_batch_settlement_ordering",
		"fuzz_fraud_proof_duplicate_encodings",
		"fuzz_route_failure_classifications",
		"fuzz_async_delta_aggregation",
		"performance_channel_opens_per_block",
		"performance_cooperative_closes_per_block",
		"performance_unilateral_closes_per_block",
		"performance_disputes_per_block",
		"performance_promise_resolutions_per_block",
		"performance_virtual_channel_disputes_per_block",
		"performance_blockstm_conflict_rate_mix",
		"performance_store_v2_index_latency",
		"performance_snapshot_recovery_active_disputes",
	}
}

func isRequiredTestCoverageID(id RequiredTestCoverageID) bool {
	for _, required := range requiredTestCoverageIDs() {
		if id == required {
			return true
		}
	}
	return false
}
