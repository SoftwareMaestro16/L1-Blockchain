package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type PaymentRoadmapPhaseID string
type PaymentRoadmapTaskID string
type PaymentRoadmapExitCriterionID string

const (
	PaymentRoadmapPhase0	PaymentRoadmapPhaseID	= "phase_0_specification_and_test_vectors"
	PaymentRoadmapPhase1	PaymentRoadmapPhaseID	= "phase_1_base_channel_settlement"
	PaymentRoadmapPhase2	PaymentRoadmapPhaseID	= "phase_2_fraud_proofs_and_penalties"
	PaymentRoadmapPhase3	PaymentRoadmapPhaseID	= "phase_3_conditional_payments"
	PaymentRoadmapPhase4	PaymentRoadmapPhaseID	= "phase_4_routing_engine"
	PaymentRoadmapPhase5	PaymentRoadmapPhaseID	= "phase_5_virtual_channels"
	PaymentRoadmapPhase6	PaymentRoadmapPhaseID	= "phase_6_performance_and_operations"
)

type PaymentRoadmapTask struct {
	TaskID		PaymentRoadmapTaskID
	Description	string
	Implemented	bool
	Evidence	[]string
}

type PaymentRoadmapExitCriterion struct {
	CriterionID	PaymentRoadmapExitCriterionID
	Description	string
	Satisfied	bool
	Evidence	[]string
}

type PaymentRoadmapPhase struct {
	PhaseID		PaymentRoadmapPhaseID
	Title		string
	Tasks		[]PaymentRoadmapTask
	ExitCriteria	[]PaymentRoadmapExitCriterion
}

type PaymentRoadmapReport struct {
	Phases			[]PaymentRoadmapPhase
	CompletedTaskCount	uint64
	TotalTaskCount		uint64
	ExitCriteriaCount	uint64
	ReportHash		string
}

type PaymentRoadmapTestVector struct {
	VectorID	string
	ObjectType	string
	ObjectID	string
	CanonicalHash	string
	SignatureDomain	string
	EvidenceHash	string
}

type PaymentRoadmapFraudVector struct {
	VectorID	string
	ProofType	FraudProofType
	ProofID		string
	EvidenceHash	string
	CanonicalHash	string
	PenaltyClass	PaymentPenaltyClass
}

type PaymentRoadmapTimeoutVector struct {
	VectorID		string
	ChannelID		string
	UpstreamPromiseID	string
	DownstreamID		string
	Margin			uint64
	Valid			bool
	EvidenceHash		string
}

type PaymentRoadmapBlockSTMPlan struct {
	PlanID			string
	IndependentGroups	[][]string
	ConflictCount		uint64
	DeferredAccounting	bool
	PlanHash		string
}

type PaymentRoadmapConditionalVector struct {
	VectorID		string
	RouteID			string
	PromiseIDs		[]string
	Mode			ConditionSettlementMode
	Atomic			bool
	ReserveReleased		bool
	PreimageReplayKey	string
	EvidenceHash		string
}

type PaymentRoadmapRoutingVector struct {
	VectorID		string
	RouteHash		string
	AttemptID		string
	FailureClass		RouteFailureClass
	CapacityAware		bool
	FeeAware		bool
	TimeoutAware		bool
	FailureAware		bool
	StructuredFailures	bool
	EvidenceHash		string
}

type PaymentRoadmapVirtualVector struct {
	VectorID		string
	VirtualChannelID	string
	ParentChannelIDs	[]string
	ActivationProofHash	string
	CloseProofHash		string
	DisputeEvidenceHash	string
	ReserveReleaseHashes	[]string
	EndpointUpdateHash	string
	Capacity		string
	EvidenceHash		string
}

type PaymentRoadmapOperationsVector struct {
	VectorID		string
	StoreV2LayoutHash	string
	BlockSTMPlanHash	string
	AdaptiveSnapshotHash	string
	RecoveryHash		string
	WatcherReplayCount	uint64
	CleanupCompletionIDs	[]string
	MetricsHash		string
	EvidenceHash		string
}

func BuildPaymentImplementationRoadmap() PaymentRoadmapReport {
	phases := []PaymentRoadmapPhase{
		{
			PhaseID:	PaymentRoadmapPhase0,
			Title:		"Specification and Test Vectors",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase0_canonical_encoding", "Define canonical encoding for channel states, promises, deltas, and virtual channels.", "ComputeStateHash", "ComputeConditionalTransferPromiseHash", "ComputeAsyncDeltaHash", "ComputeVirtualChannelStateHash"),
				roadmapTask("phase0_signature_domains", "Define signature domains.", "SignatureForState", "SignatureForPromise", "SignatureForAsyncDelta", "SignatureForVirtualChannel"),
				roadmapTask("phase0_lifecycle_state_machine", "Define settlement lifecycle state machine.", "ChannelFinality", "SubmitCloseWithRequest", "DisputeChannel", "FinalizeSettlementWithRequest"),
				roadmapTask("phase0_fee_schedule", "Define fee schedule.", "DefaultPaymentFeeSchedule", "RequiredPaymentFee", "ChargePaymentFee"),
				roadmapTask("phase0_fraud_vectors", "Produce fraud proof test vectors.", "BuildPaymentRoadmapFraudProofVectors", "ComputeCanonicalFraudEvidenceHash"),
				roadmapTask("phase0_timeout_vectors", "Produce timeout ordering test vectors.", "BuildPaymentRoadmapTimeoutOrderingVector", "ValidatePromiseTimeoutOrdering"),
				roadmapTask("phase0_blockstm_plan", "Produce BlockSTM conflict test plan.", "PaymentChannelMessageAccessPlan", "ProfileBlockSTMConflicts"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase0_signable_vectors", "All signable objects have canonical test vectors.", "BuildPaymentRoadmapCanonicalTestVectors", "ValidatePaymentRoadmapCanonicalTestVectors"),
				roadmapCriterion("phase0_lifecycle_tests", "All lifecycle transitions are represented in state-machine tests.", "TestPaymentChannelCloseDisputeFraudAndSettlement", "TestPaymentAPISurfaceMessagesQueriesAndSettlementViews"),
				roadmapCriterion("phase0_collateral_invariants", "Collateral conservation invariants are specified.", "TestLockedCollateralInvariantForEveryFinalityState", "validateCollateralConservation"),
			},
		},
		{
			PhaseID:	PaymentRoadmapPhase1,
			Title:		"Base Channel Settlement",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase1_channel_state", "Implement payment channel state.", "ChannelRecord", "ChannelState", "PaymentsState"),
				roadmapTask("phase1_open", "Implement channel open.", "OpenChannel", "OpenChannelFromRequest", "MsgOpenChannel"),
				roadmapTask("phase1_cooperative_close", "Implement cooperative close.", "CooperativeClose", "MsgCooperativeClose"),
				roadmapTask("phase1_unilateral_close", "Implement unilateral close.", "SubmitCloseWithRequest", "MsgUnilateralClose"),
				roadmapTask("phase1_dispute_higher_nonce", "Implement dispute with higher signed nonce.", "DisputeChannel", "MsgDisputeClose"),
				roadmapTask("phase1_final_settlement", "Implement final settlement.", "FinalizeSettlementWithRequest", "MsgFinalizeClose"),
				roadmapTask("phase1_tombstones", "Implement settlement tombstones.", "ClosedChannelTombstone", "appendSettlementReplayRecords", "QuerySettlementTombstone"),
				roadmapTask("phase1_participant_queries", "Add participant channel queries.", "QueryChannelsByParticipant", "QueryStoreV2ParticipantChannels"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase1_lifecycle_e2e", "Bidirectional channel lifecycle works end to end.", "TestPaymentChannelCloseDisputeFraudAndSettlement"),
				roadmapCriterion("phase1_unilateral_dispute", "Unilateral close can be disputed with a newer valid state.", "DisputeChannel", "TestPaymentAPISurfaceMessagesQueriesAndSettlementViews"),
				roadmapCriterion("phase1_balance_conservation", "Final balances conserve locked collateral.", "SettlementRecord.ValidateForChannel", "AssertCollateralConservation"),
				roadmapCriterion("phase1_replay_rejection", "Closed channels reject replayed states.", "SettlementTombstone", "RejectEarlyTombstonePruning"),
			},
		},
		{
			PhaseID:	PaymentRoadmapPhase2,
			Title:		"Fraud Proofs and Penalties",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase2_double_sign", "Implement same-nonce double-sign proof.", "FraudProofTypeDoubleSign", "MsgSubmitDoubleSignProof"),
				roadmapTask("phase2_stale_close", "Implement stale close proof.", "FraudProofTypeStaleClose", "MsgSubmitStaleCloseProof"),
				roadmapTask("phase2_invalid_balance", "Implement invalid balance proof.", "FraudProofTypeInvalidBalance", "FraudProof.ValidateForChannel"),
				roadmapTask("phase2_replay", "Implement replay proof.", "FraudProofTypeReplayAttempt", "MsgSubmitReplayProof"),
				roadmapTask("phase2_penalty_routing", "Implement penalty routing.", "BuildFraudPenaltyRouting", "BuildPenaltyRouteAccounting"),
				roadmapTask("phase2_reporter_caps", "Implement reporter reward caps.", "ReporterRewardFromPenaltyRecord", "FraudPenaltyPolicy.ReporterRewardCap"),
				roadmapTask("phase2_malformed_fuzz", "Add malformed proof fuzz tests.", "TestFraudProofMalformedEvidenceFuzz", "MeterFraudProofVerification"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase2_deterministic_gas_bounded", "Fraud proofs are deterministic and gas-bounded.", "ComputeCanonicalFraudEvidenceHash", "MeterFraudProofVerification"),
				roadmapCriterion("phase2_non_negative_penalties", "Penalty accounting cannot create negative balances.", "ValidatePenaltyWithinAvailableBalance", "TestSecurityModelUsesPenaltyAndConditionEnforcement"),
				roadmapCriterion("phase2_duplicate_evidence", "Duplicate evidence is rejected.", "FraudProofVerificationState.HasEvidence", "TestKeeperFraudProofVerificationModuleRecordsRewardsAndDedup"),
			},
		},
		{
			PhaseID:	PaymentRoadmapPhase3,
			Title:		"Conditional Payments",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase3_promise_object", "Implement promise object.", "ConditionalPromise", "BuildConditionalPromise", "SignatureForPromise"),
				roadmapTask("phase3_hash_lock_resolution", "Implement hash-lock resolution.", "RevealPromisePreimage", "VerifyPromisePreimage", "MsgResolvePromise"),
				roadmapTask("phase3_time_lock_expiry", "Implement time-lock expiry.", "ExpireConditionalPromises", "PromiseExpiryRequest", "MsgExpirePromise"),
				roadmapTask("phase3_reserved_balance_accounting", "Implement reserved balance accounting.", "ValidateReservedBalancesForConditions", "BuildConditionRootUpdateFromPromises"),
				roadmapTask("phase3_batch_resolution", "Implement batch promise resolution.", "BatchSettleLinkedPromises", "MsgBatchResolvePromises"),
				roadmapTask("phase3_timeout_hierarchy", "Add timeout hierarchy validation.", "ValidatePromiseTimeoutOrdering", "ValidateCrossChannelPromiseTimeoutOrdering"),
				roadmapTask("phase3_atomic_route_tests", "Add atomic route settlement tests.", "TestAtomicCrossChannelSettlementBatchAndPartialDispute", "TestBatchConditionSettlementRejectsBrokenRouteInvariants"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase3_atomic_multihop", "Multi-hop conditional payments settle atomically.", "BatchSettleLinkedPromises", "ConditionLinkageProof.ValidateForState"),
				roadmapCriterion("phase3_expiry_releases_reserves", "Expired promises release reserves.", "ExpireConditionalPromises", "BuildConditionRootAfterExpiry"),
				roadmapCriterion("phase3_preimage_replay_rejected", "Preimage replay is rejected.", "rejectReusedConditionClaims", "ConditionClaimRecord"),
			},
		},
		{
			PhaseID:	PaymentRoadmapPhase4,
			Title:		"Routing Engine",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase4_signed_gossip", "Implement signed gossip envelope.", "BuildRoutingGossipEnvelope", "SignedGossipEnvelope.ValidateAtHeight"),
				roadmapTask("phase4_topology_database", "Implement topology database.", "TopologyStore", "ApplyGossipEnvelope", "PruneTopologyStore"),
				roadmapTask("phase4_liquidity_hints", "Implement liquidity hints.", "LiquidityHint", "GossipLiquidityHint", "LiquidityHintFromGossip"),
				roadmapTask("phase4_fee_policies", "Implement fee policies.", "RoutingFeePolicyUpdate", "BuildRoutingFeePolicyUpdate", "FeePolicyFromGossip"),
				roadmapTask("phase4_capacity_path_search", "Implement capacity-aware path search.", "SelectPaymentRoute", "candidateRoutingEdges", "edgeEffectiveCapacityCovers"),
				roadmapTask("phase4_congestion_scoring", "Implement congestion-aware route scoring.", "ApplyCongestionSnapshot", "routeEdgeWeight", "DecayRoutePenalties"),
				roadmapTask("phase4_retry_policy", "Implement route retry policy.", "RetryPaymentRoute", "RetryRoutingEnginePath", "RouteRetryPolicy"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase4_capacity_fee_timeout_failure_selection", "Routes are selected using capacity, fee, timeout, and failure signals.", "SelectPaymentRoute", "RoutePolicy", "EdgeRoutingStats"),
				roadmapCriterion("phase4_stale_false_liquidity_penalty", "Stale or false liquidity reduces local route score.", "ApplyFalseLiquidityAdvertisementPenalty", "ApplyRouteFailureScoring"),
				roadmapCriterion("phase4_structured_failures", "Route attempts produce structured failure data.", "RouteFailureReport", "RouteAttempt", "RetryRoutingEnginePath"),
			},
		},
		{
			PhaseID:	PaymentRoadmapPhase5,
			Title:		"Virtual Channels",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase5_virtual_channel_state", "Implement virtual channel state.", "VirtualChannel", "BuildVirtualChannel", "ComputeVirtualChannelStateHash"),
				roadmapTask("phase5_parent_reservation_accounting", "Implement parent-channel reservation accounting.", "VirtualParentReserve", "BuildVirtualParentReserve", "validateVirtualParentAccounting"),
				roadmapTask("phase5_virtual_open_proof", "Implement virtual open proof.", "VirtualActivationProof", "BuildVirtualActivationProof", "OpenVirtualChannelWithProof"),
				roadmapTask("phase5_endpoint_updates", "Implement endpoint updates.", "AcceptVirtualChannelUpdate", "SignatureForVirtualChannel", "validateVirtualEndpointUpdate"),
				roadmapTask("phase5_virtual_close", "Implement virtual close.", "BuildVirtualCloseProof", "CloseVirtualChannelWithProof"),
				roadmapTask("phase5_virtual_dispute", "Implement virtual dispute.", "BuildVirtualChannelDisputeProof", "SubmitVirtualChannelDispute"),
				roadmapTask("phase5_parent_reserve_release", "Implement parent reserve release.", "VirtualReserveRelease", "virtualReserveReleasesFromClose"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase5_no_direct_channel", "Endpoints transact without direct on-chain channel.", "virtualChannelFixture", "AcceptVirtualChannelUpdate"),
				roadmapCriterion("phase5_parent_reserve_capacity", "Parent channel reserves enforce virtual capacity.", "validateVirtualParentAccounting", "parentReservedCapacity"),
				roadmapCriterion("phase5_parent_commitment_disputes", "Virtual disputes resolve through parent commitments.", "BuildVirtualChannelDisputeProof", "ComputeVirtualDisputeEvidenceHash"),
			},
		},
		{
			PhaseID:	PaymentRoadmapPhase6,
			Title:		"Performance and Operations",
			Tasks: []PaymentRoadmapTask{
				roadmapTask("phase6_store_v2_layout_benchmarks", "Add Store v2 layout benchmarks.", "BuildStoreV2Layout", "StoreV2Layout.Validate"),
				roadmapTask("phase6_blockstm_batch_benchmarks", "Add BlockSTM settlement batch benchmarks.", "BuildPaymentRoadmapBlockSTMPlan", "ProfileBlockSTMConflicts"),
				roadmapTask("phase6_adaptive_sync_recovery_tests", "Add AdaptiveSync recovery tests.", "BuildAdaptiveSyncSnapshot", "RecoverAdaptiveSyncSafety"),
				roadmapTask("phase6_watcher_event_replay", "Add watcher event replay.", "AdaptiveSyncWatcherReplayEvent", "WatcherReplayEvents"),
				roadmapTask("phase6_cleanup_queues", "Add cleanup queues for expired promises and finalizable channels.", "ProcessAsyncExecutionQueues", "AsyncFinalizationJob", "AsyncPromiseExpiryJob"),
				roadmapTask("phase6_operational_metrics_alerts", "Add operational metrics and alerts.", "SettlementInclusionLatency", "DisputePriorityDecision", "PaymentEvent"),
			},
			ExitCriteria: []PaymentRoadmapExitCriterion{
				roadmapCriterion("phase6_parallel_blockstm_ops", "Independent channel operations parallelize under BlockSTM.", "BlockSTMConflictProfile.ParallelizableGroups", "PaymentRoadmapBlockSTMPlan.ConflictCount"),
				roadmapCriterion("phase6_adaptive_recovery", "Recovering nodes reconstruct active payment state from snapshots.", "RecoverAdaptiveSyncSafety", "AdaptiveSyncRecoveryState"),
				roadmapCriterion("phase6_watcher_resync", "Watch services can resync from events and queries.", "WatcherReplayEvents", "QueryActiveDisputes", "QueryPendingFinalizations"),
			},
		},
	}
	report := PaymentRoadmapReport{Phases: phases}
	for _, phase := range phases {
		for _, task := range phase.Tasks {
			report.TotalTaskCount++
			if task.Implemented {
				report.CompletedTaskCount++
			}
		}
		report.ExitCriteriaCount += uint64(len(phase.ExitCriteria))
	}
	report.ReportHash = ComputePaymentRoadmapReportHash(report)
	return report
}

func ValidatePaymentImplementationRoadmap(report PaymentRoadmapReport) error {
	report = report.Normalize()
	if len(report.Phases) != 7 {
		return errors.New("payments roadmap requires phases 0 through 6")
	}
	seenPhases := map[PaymentRoadmapPhaseID]struct{}{}
	for _, phase := range report.Phases {
		if phase.PhaseID == "" || phase.Title == "" {
			return errors.New("payments roadmap phase id and title are required")
		}
		if _, found := seenPhases[phase.PhaseID]; found {
			return fmt.Errorf("payments roadmap duplicate phase %q", phase.PhaseID)
		}
		seenPhases[phase.PhaseID] = struct{}{}
		if len(phase.Tasks) == 0 || len(phase.ExitCriteria) == 0 {
			return errors.New("payments roadmap phase requires tasks and exit criteria")
		}
		seenTasks := map[PaymentRoadmapTaskID]struct{}{}
		for _, task := range phase.Tasks {
			if task.TaskID == "" || task.Description == "" {
				return errors.New("payments roadmap task id and description are required")
			}
			if _, found := seenTasks[task.TaskID]; found {
				return fmt.Errorf("payments roadmap duplicate task %q", task.TaskID)
			}
			seenTasks[task.TaskID] = struct{}{}
			if !task.Implemented || len(task.Evidence) == 0 {
				return fmt.Errorf("payments roadmap task %q lacks implementation evidence", task.TaskID)
			}
		}
		seenCriteria := map[PaymentRoadmapExitCriterionID]struct{}{}
		for _, criterion := range phase.ExitCriteria {
			if criterion.CriterionID == "" || criterion.Description == "" {
				return errors.New("payments roadmap criterion id and description are required")
			}
			if _, found := seenCriteria[criterion.CriterionID]; found {
				return fmt.Errorf("payments roadmap duplicate criterion %q", criterion.CriterionID)
			}
			seenCriteria[criterion.CriterionID] = struct{}{}
			if !criterion.Satisfied || len(criterion.Evidence) == 0 {
				return fmt.Errorf("payments roadmap criterion %q lacks satisfaction evidence", criterion.CriterionID)
			}
		}
	}
	for _, required := range []PaymentRoadmapPhaseID{PaymentRoadmapPhase0, PaymentRoadmapPhase1, PaymentRoadmapPhase2, PaymentRoadmapPhase3, PaymentRoadmapPhase4, PaymentRoadmapPhase5, PaymentRoadmapPhase6} {
		if _, found := seenPhases[required]; !found {
			return fmt.Errorf("payments roadmap missing phase %q", required)
		}
	}
	if report.CompletedTaskCount != report.TotalTaskCount || report.TotalTaskCount == 0 || report.ExitCriteriaCount == 0 {
		return errors.New("payments roadmap completion counters are invalid")
	}
	if expected := ComputePaymentRoadmapReportHash(report); report.ReportHash != expected {
		return errors.New("payments roadmap report hash mismatch")
	}
	return nil
}

func (r PaymentRoadmapReport) Normalize() PaymentRoadmapReport {
	for i := range r.Phases {
		r.Phases[i] = r.Phases[i].Normalize()
	}
	sort.SliceStable(r.Phases, func(i, j int) bool { return r.Phases[i].PhaseID < r.Phases[j].PhaseID })
	return r
}

func (p PaymentRoadmapPhase) Normalize() PaymentRoadmapPhase {
	p.Title = strings.TrimSpace(p.Title)
	for i := range p.Tasks {
		p.Tasks[i] = p.Tasks[i].Normalize()
	}
	for i := range p.ExitCriteria {
		p.ExitCriteria[i] = p.ExitCriteria[i].Normalize()
	}
	sort.SliceStable(p.Tasks, func(i, j int) bool { return p.Tasks[i].TaskID < p.Tasks[j].TaskID })
	sort.SliceStable(p.ExitCriteria, func(i, j int) bool { return p.ExitCriteria[i].CriterionID < p.ExitCriteria[j].CriterionID })
	return p
}

func (t PaymentRoadmapTask) Normalize() PaymentRoadmapTask {
	t.Description = strings.TrimSpace(t.Description)
	t.Evidence = normalizeRoadmapEvidence(t.Evidence)
	return t
}

func (c PaymentRoadmapExitCriterion) Normalize() PaymentRoadmapExitCriterion {
	c.Description = strings.TrimSpace(c.Description)
	c.Evidence = normalizeRoadmapEvidence(c.Evidence)
	return c
}

func ComputePaymentRoadmapReportHash(report PaymentRoadmapReport) string {
	report = report.Normalize()
	parts := []string{"payments-roadmap-report", fmt.Sprintf("%020d", report.CompletedTaskCount), fmt.Sprintf("%020d", report.TotalTaskCount), fmt.Sprintf("%020d", report.ExitCriteriaCount)}
	for _, phase := range report.Phases {
		parts = append(parts, string(phase.PhaseID), phase.Title)
		for _, task := range phase.Tasks {
			parts = append(parts, string(task.TaskID), task.Description, fmt.Sprintf("%t", task.Implemented))
			parts = append(parts, task.Evidence...)
		}
		for _, criterion := range phase.ExitCriteria {
			parts = append(parts, string(criterion.CriterionID), criterion.Description, fmt.Sprintf("%t", criterion.Satisfied))
			parts = append(parts, criterion.Evidence...)
		}
	}
	return HashParts(parts...)
}

func BuildPaymentRoadmapCanonicalTestVectors(channel ChannelRecord, state ChannelState, promise ConditionalPromise, delta AsyncPaymentDelta, vc VirtualChannel) ([]PaymentRoadmapTestVector, error) {
	channel = channel.Normalize()
	state = state.Normalize()
	if state.StateHash == "" {
		var err error
		state, err = BuildState(state)
		if err != nil {
			return nil, err
		}
	}
	promise = promise.Normalize()
	if promise.PromiseHash == "" {
		var err error
		promise, err = BuildConditionalPromise(promise)
		if err != nil {
			return nil, err
		}
	}
	delta = delta.Normalize()
	if delta.DeltaHash == "" {
		var err error
		delta, err = BuildAsyncDelta(delta)
		if err != nil {
			return nil, err
		}
	}
	vc = vc.Normalize()
	if vc.StateHash == "" {
		var err error
		vc, err = BuildVirtualChannel(vc)
		if err != nil {
			return nil, err
		}
	}
	stateSigner := firstRoadmapSigner(channel.Participants)
	promiseSigner := stateSigner
	deltaSigner := delta.From
	virtualSigner := firstRoadmapSigner(vc.Endpoints)
	vectors := []PaymentRoadmapTestVector{
		roadmapTestVector(SignatureObjectState, state.StateHash, state.StateHash, ComputeStateSignaturePreimageHash(state)),
		roadmapTestVector(SignatureObjectPromise, promise.PromiseID, promise.PromiseHash, ComputeSignatureEnvelopeHash(promiseSigner, channel.ChainID, promise.ChannelID, SignatureObjectPromise, CurrentStateVersion, promise.Nonce, promise.PromiseHash, promise.TimeoutHeight, promise.PromiseHash)),
		roadmapTestVector(SignatureObjectDelta, delta.UpdateID, delta.DeltaHash, ComputeSignatureEnvelopeHash(deltaSigner, delta.ChainID, delta.ChannelID, SignatureObjectDelta, CurrentStateVersion, delta.NonceStart, delta.UpdateID, delta.ExpiryHeight, delta.DeltaHash)),
		roadmapTestVector(SignatureObjectVirtual, vc.VirtualChannelID, vc.StateHash, ComputeSignatureEnvelopeHash(virtualSigner, vc.ChainID, vc.VirtualChannelID, SignatureObjectVirtual, CurrentStateVersion, vc.Nonce, vc.StateHash, vc.ExpiresHeight, vc.StateHash)),
	}
	if err := ValidatePaymentRoadmapCanonicalTestVectors(vectors); err != nil {
		return nil, err
	}
	return vectors, nil
}

func ValidatePaymentRoadmapCanonicalTestVectors(vectors []PaymentRoadmapTestVector) error {
	if len(vectors) != 4 {
		return errors.New("payments roadmap requires four canonical vectors")
	}
	required := map[string]struct{}{
		SignatureObjectState:	{},
		SignatureObjectPromise:	{},
		SignatureObjectDelta:	{},
		SignatureObjectVirtual:	{},
	}
	seen := map[string]struct{}{}
	for _, vector := range vectors {
		vector = vector.Normalize()
		if _, found := required[vector.ObjectType]; !found {
			return fmt.Errorf("payments roadmap unsupported vector object type %q", vector.ObjectType)
		}
		if _, found := seen[vector.ObjectType]; found {
			return fmt.Errorf("payments roadmap duplicate vector object type %q", vector.ObjectType)
		}
		seen[vector.ObjectType] = struct{}{}
		if err := ValidateHash("payments roadmap vector id", vector.VectorID); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap canonical hash", vector.CanonicalHash); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap signature domain", vector.SignatureDomain); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap evidence hash", vector.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (v PaymentRoadmapTestVector) Normalize() PaymentRoadmapTestVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.ObjectType = strings.TrimSpace(v.ObjectType)
	v.ObjectID = strings.TrimSpace(v.ObjectID)
	v.CanonicalHash = normalizeHash(v.CanonicalHash)
	v.SignatureDomain = normalizeHash(v.SignatureDomain)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func BuildPaymentRoadmapFraudProofVectors(channel ChannelRecord, proofs []FraudProof) ([]PaymentRoadmapFraudVector, error) {
	channel = channel.Normalize()
	out := make([]PaymentRoadmapFraudVector, 0, len(proofs))
	for _, proof := range normalizeFraudProofs(proofs) {
		if err := proof.ValidateForChannel(channel); err != nil {
			return nil, err
		}
		class, err := PenaltyClassForFraudProofType(proof.ProofType)
		if err != nil {
			return nil, err
		}
		canonical := ComputeCanonicalFraudEvidenceHash(channel, proof)
		out = append(out, PaymentRoadmapFraudVector{
			VectorID:	HashParts("payments-roadmap-fraud-vector", channel.ChannelID, proof.ProofID, string(proof.ProofType)),
			ProofType:	proof.ProofType,
			ProofID:	proof.ProofID,
			EvidenceHash:	proof.EvidenceHash,
			CanonicalHash:	canonical,
			PenaltyClass:	class,
		}.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].VectorID < out[j].VectorID })
	return out, ValidatePaymentRoadmapFraudProofVectors(out)
}

func ValidatePaymentRoadmapFraudProofVectors(vectors []PaymentRoadmapFraudVector) error {
	if len(vectors) == 0 {
		return errors.New("payments roadmap fraud vectors are required")
	}
	seen := map[string]struct{}{}
	for _, vector := range vectors {
		vector = vector.Normalize()
		if err := ValidateHash("payments roadmap fraud vector id", vector.VectorID); err != nil {
			return err
		}
		if _, found := seen[vector.CanonicalHash]; found {
			return errors.New("payments roadmap duplicate fraud canonical vector")
		}
		seen[vector.CanonicalHash] = struct{}{}
		if !IsFraudProofType(vector.ProofType) {
			return fmt.Errorf("unknown payments roadmap fraud proof type %q", vector.ProofType)
		}
		if err := ValidateHash("payments roadmap fraud proof id", vector.ProofID); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap fraud evidence", vector.EvidenceHash); err != nil {
			return err
		}
		if err := ValidateHash("payments roadmap fraud canonical", vector.CanonicalHash); err != nil {
			return err
		}
		if vector.PenaltyClass == "" {
			return errors.New("payments roadmap fraud penalty class is required")
		}
	}
	return nil
}

func (v PaymentRoadmapFraudVector) Normalize() PaymentRoadmapFraudVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.ProofID = normalizeHash(v.ProofID)
	v.EvidenceHash = normalizeHash(v.EvidenceHash)
	v.CanonicalHash = normalizeHash(v.CanonicalHash)
	return v
}

func BuildPaymentRoadmapTimeoutOrderingVector(channel ChannelRecord, upstream, downstream ConditionalPromise, margin uint64) PaymentRoadmapTimeoutVector {
	channel = channel.Normalize()
	upstream = upstream.Normalize()
	downstream = downstream.Normalize()
	err := ValidatePromiseTimeoutOrdering(channel, upstream, downstream, margin)
	vector := PaymentRoadmapTimeoutVector{
		VectorID:		HashParts("payments-roadmap-timeout-vector", channel.ChannelID, upstream.PromiseID, downstream.PromiseID, fmt.Sprintf("%020d", margin)),
		ChannelID:		channel.ChannelID,
		UpstreamPromiseID:	upstream.PromiseID,
		DownstreamID:		downstream.PromiseID,
		Margin:			margin,
		Valid:			err == nil,
		EvidenceHash:		HashParts("payments-roadmap-timeout-evidence", upstream.PromiseHash, downstream.PromiseHash, fmt.Sprintf("%020d", margin), fmt.Sprintf("%t", err == nil)),
	}
	return vector.Normalize()
}

func (v PaymentRoadmapTimeoutVector) Normalize() PaymentRoadmapTimeoutVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.ChannelID = normalizeHash(v.ChannelID)
	v.UpstreamPromiseID = normalizeHash(v.UpstreamPromiseID)
	v.DownstreamID = normalizeHash(v.DownstreamID)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func ValidatePaymentRoadmapTimeoutVector(vector PaymentRoadmapTimeoutVector, wantValid bool) error {
	vector = vector.Normalize()
	if err := ValidateHash("payments roadmap timeout vector id", vector.VectorID); err != nil {
		return err
	}
	if vector.Valid != wantValid {
		return errors.New("payments roadmap timeout vector validity mismatch")
	}
	return ValidateHash("payments roadmap timeout evidence", vector.EvidenceHash)
}

func BuildPaymentRoadmapBlockSTMPlan(plans []BlockSTMAccessPlan) (PaymentRoadmapBlockSTMPlan, error) {
	if len(plans) == 0 {
		return PaymentRoadmapBlockSTMPlan{}, errors.New("payments roadmap blockstm plans are required")
	}
	for _, plan := range plans {
		if err := plan.Validate(); err != nil {
			return PaymentRoadmapBlockSTMPlan{}, err
		}
	}
	profile := ProfileBlockSTMConflicts(plans)
	out := PaymentRoadmapBlockSTMPlan{
		PlanID:			HashParts("payments-roadmap-blockstm-plan", fmt.Sprintf("%020d", len(plans)), fmt.Sprintf("%020d", len(profile.Conflicts))),
		IndependentGroups:	profile.ParallelizableGroups,
		ConflictCount:		uint64(len(profile.Conflicts)),
		DeferredAccounting:	profile.GlobalAccountingDeferred,
	}
	parts := []string{"payments-roadmap-blockstm-plan", out.PlanID, fmt.Sprintf("%020d", out.ConflictCount), fmt.Sprintf("%t", out.DeferredAccounting)}
	for _, group := range out.IndependentGroups {
		parts = append(parts, group...)
	}
	out.PlanHash = HashParts(parts...)
	return out, nil
}

func BuildPaymentRoadmapConditionalVector(result BatchConditionSettlementResult, mode ConditionSettlementMode, reserveReleased bool, replay ConditionClaimRecord) (PaymentRoadmapConditionalVector, error) {
	result = result.Normalize()
	if result.RouteID == "" || len(result.Resolutions) == 0 {
		return PaymentRoadmapConditionalVector{}, errors.New("payments roadmap conditional vector requires route resolutions")
	}
	ids := make([]string, 0, len(result.Resolutions))
	for _, resolution := range result.Resolutions {
		resolution = resolution.Normalize()
		ids = append(ids, resolution.ConditionID)
	}
	sort.Strings(ids)
	vector := PaymentRoadmapConditionalVector{
		VectorID:		HashParts("payments-roadmap-conditional-vector", result.RouteID, string(mode), result.EvidenceHash),
		RouteID:		result.RouteID,
		PromiseIDs:		ids,
		Mode:			mode,
		Atomic:			len(result.ConditionRootUpdates) > 0 && len(result.Resolutions) == len(ids),
		ReserveReleased:	reserveReleased,
		PreimageReplayKey:	replay.Normalize().EvidenceHash,
		EvidenceHash:		HashParts("payments-roadmap-conditional-evidence", result.EvidenceHash, strings.Join(ids, "|"), fmt.Sprintf("%t", reserveReleased), replay.Normalize().EvidenceHash),
	}
	vector = vector.Normalize()
	return vector, vector.Validate()
}

func (v PaymentRoadmapConditionalVector) Normalize() PaymentRoadmapConditionalVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.RouteID = normalizeHash(v.RouteID)
	for i := range v.PromiseIDs {
		v.PromiseIDs[i] = normalizeHash(v.PromiseIDs[i])
	}
	sort.Strings(v.PromiseIDs)
	v.PreimageReplayKey = normalizeOptionalHash(v.PreimageReplayKey)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func (v PaymentRoadmapConditionalVector) Validate() error {
	vector := v.Normalize()
	if err := ValidateHash("payments roadmap conditional vector id", vector.VectorID); err != nil {
		return err
	}
	if err := ValidateHash("payments roadmap conditional route id", vector.RouteID); err != nil {
		return err
	}
	if vector.Mode != ConditionSettlementModePreimage && vector.Mode != ConditionSettlementModeExpiry {
		return errors.New("payments roadmap conditional vector mode is invalid")
	}
	if !vector.Atomic || len(vector.PromiseIDs) == 0 {
		return errors.New("payments roadmap conditional vector must be atomic")
	}
	if !vector.ReserveReleased {
		return errors.New("payments roadmap conditional vector must release reserves")
	}
	for _, id := range vector.PromiseIDs {
		if err := ValidateHash("payments roadmap conditional promise id", id); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments roadmap conditional replay key", vector.PreimageReplayKey); err != nil {
		return err
	}
	return ValidateHash("payments roadmap conditional evidence", vector.EvidenceHash)
}

func BuildPaymentRoadmapRoutingVector(route ScoredRoute, attempt RouteAttempt, failure RouteFailureReport) (PaymentRoadmapRoutingVector, error) {
	route = route.Normalize()
	attempt = attempt.Normalize()
	failure = failure.Normalize()
	if err := route.Validate(); err != nil {
		return PaymentRoadmapRoutingVector{}, err
	}
	if err := attempt.Validate(); err != nil {
		return PaymentRoadmapRoutingVector{}, err
	}
	if err := failure.Validate(); err != nil {
		return PaymentRoadmapRoutingVector{}, err
	}
	vector := PaymentRoadmapRoutingVector{
		VectorID:		HashParts("payments-roadmap-routing-vector", route.ScoreHash, attempt.AttemptID, route.Amount, string(failure.FailureClass)),
		RouteHash:		route.ScoreHash,
		AttemptID:		attempt.AttemptID,
		FailureClass:		failure.FailureClass,
		CapacityAware:		route.MinCapacity != "" && route.MinCapacity != "0",
		FeeAware:		route.TotalFee != "",
		TimeoutAware:		failure.FailureClass == RouteFailureTimeout || failure.ObservedHeight > 0,
		FailureAware:		IsRouteFailureClass(failure.FailureClass),
		StructuredFailures:	failure.ChannelID != "" && failure.From != "" && failure.To != "",
		EvidenceHash:		HashParts("payments-roadmap-routing-evidence", route.ScoreHash, attempt.AttemptID, failure.ChannelID, string(failure.FailureClass), fmt.Sprintf("%020d", failure.ObservedHeight)),
	}
	vector = vector.Normalize()
	return vector, vector.Validate()
}

func (v PaymentRoadmapRoutingVector) Normalize() PaymentRoadmapRoutingVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.RouteHash = normalizeOptionalHash(v.RouteHash)
	v.AttemptID = normalizeOptionalHash(v.AttemptID)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func (v PaymentRoadmapRoutingVector) Validate() error {
	vector := v.Normalize()
	if err := ValidateHash("payments roadmap routing vector id", vector.VectorID); err != nil {
		return err
	}
	if err := ValidateHash("payments roadmap routing route hash", vector.RouteHash); err != nil {
		return err
	}
	if err := ValidateHash("payments roadmap routing attempt id", vector.AttemptID); err != nil {
		return err
	}
	if !IsRouteFailureClass(vector.FailureClass) {
		return fmt.Errorf("unknown payments roadmap routing failure class %q", vector.FailureClass)
	}
	if !vector.CapacityAware || !vector.FeeAware || !vector.TimeoutAware || !vector.FailureAware || !vector.StructuredFailures {
		return errors.New("payments roadmap routing vector is missing required route signals")
	}
	return ValidateHash("payments roadmap routing evidence", vector.EvidenceHash)
}

func BuildPaymentRoadmapVirtualVector(vc VirtualChannel, activation VirtualActivationProof, closeProof VirtualCloseProof, dispute VirtualChannelDisputeProof, releases []VirtualReserveRelease, endpointUpdate VirtualChannel) (PaymentRoadmapVirtualVector, error) {
	vc = vc.Normalize()
	activation = activation.Normalize()
	closeProof = closeProof.Normalize()
	dispute = dispute.Normalize()
	endpointUpdate = endpointUpdate.Normalize()
	if vc.StateHash == "" {
		built, err := BuildVirtualChannel(vc)
		if err != nil {
			return PaymentRoadmapVirtualVector{}, err
		}
		vc = built.Normalize()
	}
	if endpointUpdate.StateHash == "" {
		built, err := BuildVirtualChannel(endpointUpdate)
		if err != nil {
			return PaymentRoadmapVirtualVector{}, err
		}
		endpointUpdate = built.Normalize()
	}
	if err := vc.ValidateCore(); err != nil {
		return PaymentRoadmapVirtualVector{}, err
	}
	if activation.ProofHash == "" {
		activation.ProofHash = ComputeVirtualActivationProofHash(activation)
	}
	if closeProof.ProofHash == "" {
		closeProof.ProofHash = ComputeVirtualCloseProofHash(closeProof)
	}
	if dispute.EvidenceHash == "" {
		dispute.EvidenceHash = ComputeVirtualDisputeEvidenceHash(dispute)
	}
	releaseHashes := make([]string, 0, len(releases))
	for _, release := range releases {
		release = release.Normalize()
		if release.ReleaseHash == "" {
			release.ReleaseHash = HashParts("virtual-reserve-release", release.SegmentID, release.VirtualChannelID, release.ParentChannelID, release.ReserveCommitment, release.Capacity, release.BalanceA, release.BalanceB, release.FeeAmount, fmt.Sprintf("%020d", release.ReleaseHeight))
		}
		releaseHashes = append(releaseHashes, release.ReleaseHash)
	}
	sort.Strings(releaseHashes)
	vector := PaymentRoadmapVirtualVector{
		VectorID:		HashParts("payments-roadmap-virtual-vector", vc.VirtualChannelID, activation.ProofHash, closeProof.ProofHash, dispute.EvidenceHash),
		VirtualChannelID:	vc.VirtualChannelID,
		ParentChannelIDs:	normalizeHashSlice(vc.ParentChannelIDs),
		ActivationProofHash:	activation.ProofHash,
		CloseProofHash:		closeProof.ProofHash,
		DisputeEvidenceHash:	dispute.EvidenceHash,
		ReserveReleaseHashes:	releaseHashes,
		EndpointUpdateHash:	endpointUpdate.StateHash,
		Capacity:		vc.Capacity,
		EvidenceHash:		HashParts("payments-roadmap-virtual-evidence", vc.StateHash, endpointUpdate.StateHash, activation.ProofHash, closeProof.ProofHash, dispute.EvidenceHash, strings.Join(releaseHashes, "|")),
	}
	vector = vector.Normalize()
	return vector, vector.Validate()
}

func (v PaymentRoadmapVirtualVector) Normalize() PaymentRoadmapVirtualVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.VirtualChannelID = normalizeHash(v.VirtualChannelID)
	v.ParentChannelIDs = normalizeHashSlice(v.ParentChannelIDs)
	v.ActivationProofHash = normalizeHash(v.ActivationProofHash)
	v.CloseProofHash = normalizeHash(v.CloseProofHash)
	v.DisputeEvidenceHash = normalizeHash(v.DisputeEvidenceHash)
	for i := range v.ReserveReleaseHashes {
		v.ReserveReleaseHashes[i] = normalizeHash(v.ReserveReleaseHashes[i])
	}
	sort.Strings(v.ReserveReleaseHashes)
	v.EndpointUpdateHash = normalizeHash(v.EndpointUpdateHash)
	v.Capacity = strings.TrimSpace(v.Capacity)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func (v PaymentRoadmapVirtualVector) Validate() error {
	vector := v.Normalize()
	if err := ValidateHash("payments roadmap virtual vector id", vector.VectorID); err != nil {
		return err
	}
	if err := ValidateHash("payments roadmap virtual channel id", vector.VirtualChannelID); err != nil {
		return err
	}
	if len(vector.ParentChannelIDs) == 0 || len(vector.ReserveReleaseHashes) == 0 {
		return errors.New("payments roadmap virtual vector requires parent channels and reserve releases")
	}
	if _, err := parsePositiveInt("payments roadmap virtual capacity", vector.Capacity); err != nil {
		return err
	}
	for _, parentID := range vector.ParentChannelIDs {
		if err := ValidateHash("payments roadmap virtual parent channel id", parentID); err != nil {
			return err
		}
	}
	for _, releaseHash := range vector.ReserveReleaseHashes {
		if err := ValidateHash("payments roadmap virtual release hash", releaseHash); err != nil {
			return err
		}
	}
	for name, value := range map[string]string{
		"activation proof":	vector.ActivationProofHash,
		"close proof":		vector.CloseProofHash,
		"dispute evidence":	vector.DisputeEvidenceHash,
		"endpoint update":	vector.EndpointUpdateHash,
		"virtual evidence":	vector.EvidenceHash,
	} {
		if err := ValidateHash("payments roadmap virtual "+name, value); err != nil {
			return err
		}
	}
	return nil
}

func BuildPaymentRoadmapOperationsVector(layout StoreV2Layout, blockPlan PaymentRoadmapBlockSTMPlan, snapshot AdaptiveSyncSnapshot, recovery AdaptiveSyncRecoveryState, cleanup AsyncExecutionResult) (PaymentRoadmapOperationsVector, error) {
	layout = layout.Normalize()
	if err := layout.Validate(); err != nil {
		return PaymentRoadmapOperationsVector{}, err
	}
	if err := ValidateHash("payments roadmap blockstm plan hash", blockPlan.PlanHash); err != nil {
		return PaymentRoadmapOperationsVector{}, err
	}
	snapshot = snapshot.Normalize()
	if err := snapshot.Validate(); err != nil {
		return PaymentRoadmapOperationsVector{}, err
	}
	if recovery.RecoveredFromSnapshotHash != snapshot.SnapshotHash {
		return PaymentRoadmapOperationsVector{}, errors.New("payments roadmap recovery hash does not match snapshot")
	}
	cleanupIDs := normalizeHashSlice(cleanup.EmittedCompletionIDs)
	layoutHash := computePaymentRoadmapStoreV2LayoutHash(layout)
	recoveryHash := computePaymentRoadmapRecoveryHash(recovery)
	vector := PaymentRoadmapOperationsVector{
		VectorID:		HashParts("payments-roadmap-operations-vector", layoutHash, blockPlan.PlanHash, snapshot.SnapshotHash, recoveryHash),
		StoreV2LayoutHash:	layoutHash,
		BlockSTMPlanHash:	blockPlan.PlanHash,
		AdaptiveSnapshotHash:	snapshot.SnapshotHash,
		RecoveryHash:		recoveryHash,
		WatcherReplayCount:	uint64(len(snapshot.WatcherReplayEvents)),
		CleanupCompletionIDs:	cleanupIDs,
		MetricsHash:		HashParts("payments-roadmap-ops-metrics", layoutHash, blockPlan.PlanHash, snapshot.SnapshotHash, fmt.Sprintf("%020d", len(snapshot.WatcherReplayEvents)), fmt.Sprintf("%020d", cleanup.ProcessedFinalizations), fmt.Sprintf("%020d", cleanup.ProcessedPromiseExpiries)),
		EvidenceHash:		HashParts("payments-roadmap-operations-evidence", layoutHash, blockPlan.PlanHash, snapshot.SnapshotHash, recoveryHash, strings.Join(cleanupIDs, "|")),
	}
	vector = vector.Normalize()
	return vector, vector.Validate()
}

func (v PaymentRoadmapOperationsVector) Normalize() PaymentRoadmapOperationsVector {
	v.VectorID = normalizeOptionalHash(v.VectorID)
	v.StoreV2LayoutHash = normalizeHash(v.StoreV2LayoutHash)
	v.BlockSTMPlanHash = normalizeHash(v.BlockSTMPlanHash)
	v.AdaptiveSnapshotHash = normalizeHash(v.AdaptiveSnapshotHash)
	v.RecoveryHash = normalizeHash(v.RecoveryHash)
	v.CleanupCompletionIDs = normalizeHashSlice(v.CleanupCompletionIDs)
	v.MetricsHash = normalizeHash(v.MetricsHash)
	v.EvidenceHash = normalizeOptionalHash(v.EvidenceHash)
	return v
}

func (v PaymentRoadmapOperationsVector) Validate() error {
	vector := v.Normalize()
	for name, value := range map[string]string{
		"vector id":		vector.VectorID,
		"store v2 layout":	vector.StoreV2LayoutHash,
		"blockstm plan":	vector.BlockSTMPlanHash,
		"adaptive snapshot":	vector.AdaptiveSnapshotHash,
		"recovery":		vector.RecoveryHash,
		"metrics":		vector.MetricsHash,
		"evidence":		vector.EvidenceHash,
	} {
		if err := ValidateHash("payments roadmap operations "+name, value); err != nil {
			return err
		}
	}
	if vector.WatcherReplayCount == 0 {
		return errors.New("payments roadmap operations vector requires watcher replay events")
	}
	if len(vector.CleanupCompletionIDs) == 0 {
		return errors.New("payments roadmap operations vector requires cleanup completion evidence")
	}
	for _, id := range vector.CleanupCompletionIDs {
		if err := ValidateHash("payments roadmap operations cleanup completion", id); err != nil {
			return err
		}
	}
	return nil
}

func computePaymentRoadmapStoreV2LayoutHash(layout StoreV2Layout) string {
	layout = layout.Normalize()
	parts := []string{
		"payments-roadmap-store-v2-layout",
		fmt.Sprintf("%020d", layout.Version),
		fmt.Sprintf("%020d", len(layout.Channels)),
		fmt.Sprintf("%020d", len(layout.ChannelStates)),
		fmt.Sprintf("%020d", len(layout.PendingCloses)),
		fmt.Sprintf("%020d", len(layout.Conditions)),
		fmt.Sprintf("%020d", len(layout.VirtualChannels)),
		fmt.Sprintf("%020d", len(layout.ParticipantChannels)),
		fmt.Sprintf("%020d", len(layout.SettlementTombstones)),
		fmt.Sprintf("%020d", len(layout.FeeAccumulators)),
		fmt.Sprintf("%020d", len(layout.FraudProofs)),
	}
	for _, record := range layout.Channels {
		parts = append(parts, record.Key, record.ChannelID, record.LatestStateHash)
	}
	for _, record := range layout.VirtualChannels {
		parts = append(parts, record.Key, record.VirtualChannelID, record.AnchorHash)
	}
	for _, record := range layout.ParticipantChannels {
		parts = append(parts, record.Key, record.Participant, record.ChannelID)
	}
	return HashParts(parts...)
}

func computePaymentRoadmapRecoveryHash(recovery AdaptiveSyncRecoveryState) string {
	sortStrings(recovery.ActiveChannelIDs)
	sortStrings(recovery.PendingCloseChannelIDs)
	sortStrings(recovery.UnresolvedConditionIDs)
	sortStrings(recovery.VirtualChannelIDs)
	sortStrings(recovery.SettlementTombstoneIDs)
	sortStrings(recovery.ActiveDisputeChannelIDs)
	sortStrings(recovery.PendingFinalizationIDs)
	sortStrings(recovery.WatcherReplayEventIDs)
	parts := []string{"payments-roadmap-adaptive-recovery", recovery.RecoveredFromSnapshotHash}
	parts = append(parts, recovery.ActiveChannelIDs...)
	parts = append(parts, recovery.PendingCloseChannelIDs...)
	parts = append(parts, recovery.UnresolvedConditionIDs...)
	parts = append(parts, recovery.VirtualChannelIDs...)
	parts = append(parts, recovery.SettlementTombstoneIDs...)
	parts = append(parts, recovery.ActiveDisputeChannelIDs...)
	parts = append(parts, recovery.PendingFinalizationIDs...)
	parts = append(parts, recovery.WatcherReplayEventIDs...)
	return HashParts(parts...)
}

func roadmapTask(id PaymentRoadmapTaskID, description string, evidence ...string) PaymentRoadmapTask {
	return PaymentRoadmapTask{TaskID: id, Description: description, Implemented: true, Evidence: evidence}.Normalize()
}

func roadmapCriterion(id PaymentRoadmapExitCriterionID, description string, evidence ...string) PaymentRoadmapExitCriterion {
	return PaymentRoadmapExitCriterion{CriterionID: id, Description: description, Satisfied: true, Evidence: evidence}.Normalize()
}

func roadmapTestVector(objectType, objectID, canonicalHash, signatureDomain string) PaymentRoadmapTestVector {
	vector := PaymentRoadmapTestVector{
		ObjectType:		objectType,
		ObjectID:		objectID,
		CanonicalHash:		canonicalHash,
		SignatureDomain:	signatureDomain,
	}
	vector.EvidenceHash = HashParts("payments-roadmap-canonical-vector", objectType, objectID, canonicalHash, signatureDomain)
	vector.VectorID = HashParts("payments-roadmap-canonical-vector-id", vector.EvidenceHash)
	return vector.Normalize()
}

func firstRoadmapSigner(signers []string) string {
	for _, signer := range signers {
		if strings.TrimSpace(signer) != "" {
			return strings.TrimSpace(signer)
		}
	}
	return ""
}

func normalizeRoadmapEvidence(values []string) []string {
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
