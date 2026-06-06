package types

import (
	"slices"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestLayeredPoSArchitectureMatchesTargetStack(t *testing.T) {
	architecture := DefaultLayeredPosArchitecture()

	require.NoError(t, architecture.Validate())
	require.Equal(t, []PosLayer{
		PosLayerEconomicConsensus,
		PosLayerTaskAssignment,
		PosLayerValidatorExecution,
		PosLayerStakingCapital,
		PosLayerBaseCometBFT,
	}, DefaultPosLayerOrder())
	require.Equal(t, ComputeLayeredPosArchitectureRoot(architecture.Layers), architecture.Root)

	specs := posLayerSpecsByLayer(architecture)
	require.Equal(t, []string{
		"finality",
		"proposal and vote protocol",
		"validator public key set",
		"consensus safety and liveness",
	}, specs[PosLayerBaseCometBFT].Responsibilities)
	require.Equal(t, []string{
		"validators",
		"delegators",
		"bonded stake",
		"unbonding",
		"redelegation",
		"capital risk preferences",
		"commission and delegation market metadata",
	}, specs[PosLayerStakingCapital].Responsibilities)
	require.Equal(t, []string{
		"block production",
		"state transition verification",
		"cross-domain proof verification",
		"signature production",
		"fault rejection",
	}, specs[PosLayerValidatorExecution].Responsibilities)
	require.Equal(t, []string{
		"workload grouping",
		"shard validator groups",
		"zone validator groups",
		"evidence verification subsets",
		"collator and verifier assignments",
	}, specs[PosLayerTaskAssignment].Responsibilities)
	require.Equal(t, []string{
		"validator scoring",
		"performance incentives",
		"stake saturation",
		"role-specific reward weights",
		"slashing severity",
		"reporter incentives",
		"treasury, burn, and stabilization routing",
	}, specs[PosLayerEconomicConsensus].Responsibilities)
}

func TestLayeredPoSArchitectureRejectsReorderedOrUpwardDependencies(t *testing.T) {
	architecture := DefaultLayeredPosArchitecture()
	slices.Reverse(architecture.Layers)
	architecture.Root = ComputeLayeredPosArchitectureRoot(architecture.Layers)
	require.ErrorContains(t, architecture.Validate(), "pos layer 0")

	architecture = DefaultLayeredPosArchitecture()
	architecture.Layers[3].DependsOn = append(architecture.Layers[3].DependsOn, PosLayerEconomicConsensus)
	architecture.Root = ComputeLayeredPosArchitectureRoot(architecture.Layers)
	require.ErrorContains(t, architecture.Validate(), "lower layers")
}

func TestEpochDefinitionDefaultsMatchLifecycleParameters(t *testing.T) {
	params := DefaultParams()

	require.NoError(t, params.Validate())
	require.Equal(t, uint64(43_200), params.EpochDurationSeconds)
	require.GreaterOrEqual(t, params.EpochDurationSeconds, MinEpochDurationSeconds)
	require.LessOrEqual(t, params.EpochDurationSeconds, MaxEpochDurationSeconds)
	require.Equal(t, EpochSeedSourcePreviousSeedValidatorSet, params.EffectiveEpochSeedSource())
	require.Equal(t, DefaultMaxValidatorSetChangeRateBps, params.MaxValidatorSetChangeRateBps)

	durations := params.EffectivePhaseDurations()
	require.Equal(t, params.EpochDurationSeconds, durations.TotalSeconds())
	require.Positive(t, durations.DelegationSeconds)
	require.Positive(t, durations.ElectionSeconds)
	require.Positive(t, durations.AssignmentSeconds)
	require.Positive(t, durations.ActiveValidationSeconds)
	require.Positive(t, durations.SettlementSeconds)

	params.EpochSeedSource = "unknown"
	require.ErrorContains(t, params.Validate(), "seed source")

	params = DefaultParams()
	params.MaxValidatorSetChangeRateBps = 0
	require.ErrorContains(t, params.Validate(), "validator set change rate")
}

func TestEpochLifecycleTransitionsFollowTargetOrder(t *testing.T) {
	lifecycle := DefaultEpochLifecycle()
	require.NoError(t, ValidateEpochLifecycle(lifecycle))
	require.Equal(t, []EpochLifecycleStep{
		{Phase: EpochPhaseDelegation, Name: "delegation phase", DurationKey: "delegation_phase_duration"},
		{Phase: EpochPhaseElection, Name: "validator election", DurationKey: "election_phase_duration"},
		{Phase: EpochPhaseAssignment, Name: "task group assignment", DurationKey: "assignment_phase_duration"},
		{Phase: EpochPhaseActive, Name: "active validation", DurationKey: "active_validation_duration"},
		{Phase: EpochPhaseSettlement, Name: "settlement + reward + slash finality", DurationKey: "settlement_phase_duration"},
	}, lifecycle)

	next, closes, err := NextEpochPhase(EpochPhaseDelegation)
	require.NoError(t, err)
	require.False(t, closes)
	require.Equal(t, EpochPhaseElection, next)
	require.NoError(t, ValidateEpochPhaseTransition(EpochPhaseElection, EpochPhaseAssignment))
	require.NoError(t, ValidateEpochPhaseTransition(EpochPhaseAssignment, EpochPhaseActive))
	require.NoError(t, ValidateEpochPhaseTransition(EpochPhaseActive, EpochPhaseSettlement))

	next, closes, err = NextEpochPhase(EpochPhaseSettlement)
	require.NoError(t, err)
	require.True(t, closes)
	require.Equal(t, EpochPhaseClosed, next)
	require.ErrorContains(t, ValidateEpochPhaseTransition(EpochPhaseDelegation, EpochPhaseActive), "invalid epoch phase transition")
}

func TestEpochLifecyclePhasesAndSeedAreDeterministic(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	validators := scoredCandidates(t, params, makeCandidates(3, 1_000))
	durations := params.EffectivePhaseDurations()

	phase, err := EpochPhaseAt(params, 1_000, 1_000)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseDelegation, phase)

	phase, err = EpochPhaseAt(params, 1_000, 1_000+durations.DelegationSeconds)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseElection, phase)

	phase, err = EpochPhaseAt(params, 1_000, 1_000+params.EpochDurationSeconds)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseClosed, phase)

	left, err := NewEpochRecord(params, 7, 100, 199, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	right, err := NewEpochRecord(params, 7, 100, 199, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	require.Equal(t, left.Seed, right.Seed)
	require.Equal(t, left.ValidatorSetHash, right.ValidatorSetHash)

	next, err := NewEpochRecord(params, 8, 200, 299, EpochPhaseDelegation, left.Seed, validators)
	require.NoError(t, err)
	require.NotEqual(t, left.Seed, next.Seed)

	sourceParams := params
	sourceParams.EpochSeedSource = EpochSeedSourceCometBFTBlockID
	sourceRecord, err := NewEpochRecord(sourceParams, 7, 100, 199, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	require.NotEqual(t, left.Seed, sourceRecord.Seed)

	_, err = CloseEpochRecord(left, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.ErrorContains(t, err, "settlement phase")

	settlementRecord, err := NewEpochRecord(params, 7, 100, 199, EpochPhaseSettlement, "", validators)
	require.NoError(t, err)
	closed, err := CloseEpochRecord(settlementRecord, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseClosed, closed.Phase)
	require.Equal(t, SettlementStatusFinalized, closed.SettlementStatus)
}

func TestEpochRecordStateContractMatchesDesignFieldsAndPhaseValues(t *testing.T) {
	require.Equal(t, []string{
		"epoch_id",
		"start_height",
		"end_height",
		"phase",
		"seed",
		"validator_set_hash",
		"task_group_root",
		"performance_root",
		"reward_root",
		"slash_root",
		"settlement_status",
	}, EpochRecordFieldNames())
	require.Equal(t, []EpochPhase{
		EpochPhaseDelegation,
		EpochPhaseElection,
		EpochPhaseAssignment,
		EpochPhaseActive,
		EpochPhaseSettlement,
		EpochPhaseClosed,
	}, EpochPhaseValues())

	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	record, err := NewEpochRecord(params, 2, 10, 20, EpochPhaseDelegation, "", scoredCandidates(t, params, makeCandidates(3, 1_000)))
	require.NoError(t, err)
	require.NoError(t, record.Validate())

	record.Phase = "invalid"
	require.ErrorContains(t, record.Validate(), "unsupported epoch phase")
}

func TestEpochRulesEnforceValidatorSetChangeBoundaries(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	previousSettlement, err := NewEpochRecord(params, 5, 100, 199, EpochPhaseSettlement, "", scoredCandidates(t, params, makeCandidates(3, 1_000)))
	require.NoError(t, err)
	previous, err := CloseEpochRecord(previousSettlement, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.NoError(t, err)
	next, err := NewEpochRecord(params, 6, 200, 299, EpochPhaseDelegation, previous.Seed, scoredCandidates(t, params, makeCandidates(3, 1_000)))
	require.NoError(t, err)

	require.NoError(t, ValidateConsecutiveEpochs(previous, next))
	require.NoError(t, ValidateValidatorSetChangeActivation(next, next.StartHeight))
	require.ErrorContains(t, ValidateValidatorSetChangeActivation(next, next.StartHeight+1), "epoch boundary")

	next.StartHeight = 201
	require.ErrorContains(t, ValidateConsecutiveEpochs(previous, next), "previous end height plus one")
}

func TestEpochRulesEnforceDelegationDelayAndEvidenceWindow(t *testing.T) {
	params := DefaultParams()
	params.DelegationActivationEpochs = 2
	params.EvidenceWindowEpochs = 4

	effectiveEpoch, err := DelegationEffectiveElectionEpoch(params, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(12), effectiveEpoch)

	active, err := DelegationAffectsElection(params, 10, 11)
	require.NoError(t, err)
	require.False(t, active)
	active, err = DelegationAffectsElection(params, 10, 12)
	require.NoError(t, err)
	require.True(t, active)

	within, err := EvidenceWithinSlashableWindow(params, 10, 14)
	require.NoError(t, err)
	require.True(t, within)
	within, err = EvidenceWithinSlashableWindow(params, 10, 15)
	require.NoError(t, err)
	require.False(t, within)
	_, err = EvidenceWithinSlashableWindow(params, 10, 9)
	require.ErrorContains(t, err, "before evidence epoch")
}

func TestMaxValidatorSetChangesUsesConfiguredRate(t *testing.T) {
	params := DefaultParams()
	params.MaxValidatorSetChangeRateBps = 1_000

	changes, err := MaxValidatorSetChanges(params, 75)
	require.NoError(t, err)
	require.Equal(t, uint32(8), changes)

	params.MaxValidatorSetChangeRateBps = BasisPoints
	changes, err = MaxValidatorSetChanges(params, 75)
	require.NoError(t, err)
	require.Equal(t, uint32(75), changes)

	_, err = MaxValidatorSetChanges(params, 0)
	require.ErrorContains(t, err, "active validator count")
}

func TestBuildElectionCandidatesAppliesDelayedDelegationsToStakeWeight(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidates := []Candidate{candidate("val-market", 1_000, 0)}
	intents := []DelegationIntent{{
		NominatorID:            "nom-a",
		ValidatorID:            "val-market",
		StakeNaet:              sdkmath.NewInt(2_000),
		RequestedEpoch:         10,
		MaxCommissionBps:       500,
		MinPerformanceScoreBps: 9_000,
	}}

	early, rejected, err := BuildElectionCandidates(params, 10, candidates, intents)
	require.NoError(t, err)
	require.Len(t, rejected, 1)
	require.True(t, early[0].DelegatedStakeNaet.IsZero())

	ready, rejected, err := BuildElectionCandidates(params, 11, candidates, intents)
	require.NoError(t, err)
	require.Empty(t, rejected)
	require.Equal(t, sdkmath.NewInt(2_000), ready[0].DelegatedStakeNaet)
	require.Equal(t, []Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(2_000)}}, ready[0].Nominations)

	scored, err := ScoreCandidate(params, ready[0])
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(3_000), scored.TotalStakeNaet)
	require.Equal(t, sdkmath.NewInt(3_000), scored.ScoreComponents.StakeWeightNaet)
}

func TestValidatorElectionFactorCalculatorsUseDeterministicBps(t *testing.T) {
	performance, err := ComputePerformanceFactor(PerformanceFactorInput{
		CompletedTasks:         8,
		MissedTasks:            2,
		CorrectVerifications:   9,
		IncorrectVerifications: 1,
		AvailableWindows:       4,
		CommittedWindows:       5,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(8_400), performance)

	uptime, err := ComputeUptimeFactor(UptimeFactorInput{
		SignedBlocks:             90,
		TotalBlocks:              100,
		TaskParticipations:       8,
		MissedTaskParticipations: 2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(8_700), uptime)

	latency, err := ComputeLatencyFactor(LatencyFactorInput{
		CommittedWindow: true,
		TargetMillis:    1_000,
		P95Millis:       2_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(5_000), latency)

	latency, err = ComputeLatencyFactor(LatencyFactorInput{CommittedWindow: true, AdvisoryOnly: true})
	require.NoError(t, err)
	require.Equal(t, uint32(BasisPoints), latency)

	reliability, err := ComputeReliabilityIndex(ReliabilityIndexInput{
		PriorIndexBps:    9_000,
		SlashEvents:      1,
		DowntimeEpochs:   2,
		MissedTasks:      3,
		RejectedEvidence: 2,
		RecoveryEpochs:   5,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(5_700), reliability)
}

func TestValidatorElectionFactorCalculatorsRejectUnsafeInputs(t *testing.T) {
	_, err := ComputePerformanceFactor(PerformanceFactorInput{AvailableWindows: 2, CommittedWindows: 1})
	require.ErrorContains(t, err, "available windows")

	_, err = ComputeUptimeFactor(UptimeFactorInput{SignedBlocks: 2, TotalBlocks: 1})
	require.ErrorContains(t, err, "signed blocks")

	_, err = ComputeLatencyFactor(LatencyFactorInput{CommittedWindow: false, TargetMillis: 1, P95Millis: 1})
	require.ErrorContains(t, err, "committed measurement")

	_, err = ComputeReliabilityIndex(ReliabilityIndexInput{PriorIndexBps: BasisPoints + 1})
	require.ErrorContains(t, err, "prior reliability")
}

func TestEpochLifecycleValidationRejectsReorderedOrDuplicatePhases(t *testing.T) {
	lifecycle := DefaultEpochLifecycle()
	slices.Reverse(lifecycle)
	require.ErrorContains(t, ValidateEpochLifecycle(lifecycle), "lifecycle step 0")

	lifecycle = DefaultEpochLifecycle()
	lifecycle[1] = lifecycle[0]
	require.ErrorContains(t, ValidateEpochLifecycle(lifecycle), "lifecycle step 1")
}

func TestScoreCandidateAppliesLatencyAndReliabilityFactors(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidate := candidate("val-latency", 1_000, 0)
	candidate.LatencyFactorBps = 5_000
	candidate.ReliabilityIndexBps = 8_000

	scored, err := ScoreCandidate(params, candidate)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(400), scored.Score)
}

func TestTaskAssignmentsAreSeededRoleAwareAndStable(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 2
	params.MaxTaskGroupValidators = 3

	candidates := makeCandidates(4, 1_000)
	candidates[0].Roles = []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier}
	candidates[1].Roles = []ValidatorRole{ValidatorRoleBlockProducer}
	candidates[2].Roles = []ValidatorRole{ValidatorRoleVerifier}
	candidates[3].Roles = []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier}
	validators := scoredCandidates(t, params, candidates)
	epoch, err := NewEpochRecord(params, 3, 30, 39, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)

	tasks := []WorkloadTask{{
		TaskID:             "task-a",
		ZoneID:             "CORE",
		ShardID:            "shard-0",
		WorkloadClass:      "settlement",
		RequiredValidators: 2,
		Roles:              []ValidatorRole{ValidatorRoleVerifier, ValidatorRoleBlockProducer},
	}}
	left, err := BuildTaskAssignments(params, epoch, validators, tasks)
	require.NoError(t, err)
	right, err := BuildTaskAssignments(params, epoch, validators, tasks)
	require.NoError(t, err)
	require.Equal(t, left.Root, right.Root)
	require.Len(t, left.Assignments, 2)
	require.NoError(t, left.Validate())

	candidateByID := map[string]Candidate{}
	for _, candidate := range candidates {
		candidateByID[candidate.ValidatorID] = candidate
	}
	for _, assignment := range left.Assignments {
		require.Len(t, assignment.Validators, 2)
		for _, validatorID := range assignment.Validators {
			require.True(t, ValidatorSupportsRole(candidateByID[validatorID], assignment.Role))
		}
	}

	_, err = BuildTaskAssignments(params, epoch, validators[1:], tasks)
	require.ErrorContains(t, err, "committed validator set hash")
}

func TestDelegationIntentsRespectActivationDelayAndRiskProfile(t *testing.T) {
	params := DefaultParams()
	params.DelegationActivationEpochs = 2
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidates := []Candidate{candidate("val-market", 1_000, 0)}
	candidates[0].PerformanceScoreBps = 9_000

	intents := []DelegationIntent{
		{
			NominatorID:            "nom-a",
			ValidatorID:            "val-market",
			StakeNaet:              sdkmath.NewInt(500),
			RequestedEpoch:         10,
			MaxCommissionBps:       500,
			MinPerformanceScoreBps: 8_000,
		},
	}
	activations, rejected, err := ActivateDelegationIntents(params, 11, candidates, intents)
	require.NoError(t, err)
	require.Empty(t, activations)
	require.Len(t, rejected, 1)
	require.Contains(t, rejected[0].Reason, "activation delay")

	intents = append(intents, DelegationIntent{
		NominatorID:            "nom-b",
		ValidatorID:            "val-market",
		StakeNaet:              sdkmath.NewInt(500),
		RequestedEpoch:         10,
		MaxCommissionBps:       500,
		MinPerformanceScoreBps: 9_500,
	})
	activations, rejected, err = ActivateDelegationIntents(params, 12, candidates, intents)
	require.NoError(t, err)
	require.Len(t, activations, 1)
	require.Equal(t, "val-market", activations[0].ValidatorID)
	require.Equal(t, sdkmath.NewInt(500), activations[0].TotalStake)
	require.Len(t, rejected, 1)
	require.Contains(t, rejected[0].Reason, "performance below")
}

func TestEvidenceSettlementPaysReporterFromObjectiveSlash(t *testing.T) {
	params := DefaultParams()
	params.ReporterRewardBps = 500
	settlement, err := SettleEvidenceCase(params, 12, EvidenceCase{
		EvidenceID:       "evidence-1",
		ReporterID:       "reporter-a",
		ValidatorID:      "val-a",
		Misbehavior:      MisbehaviorDoubleSign,
		SlashFractionBps: 1_000,
		EvidenceHeight:   99,
		EvidenceEpoch:    10,
		Finalized:        true,
	}, sdkmath.NewInt(1_000), []Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(1_000)}})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(200), settlement.Slash.TotalSlashedNaet)
	require.Equal(t, sdkmath.NewInt(10), settlement.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(190), settlement.BurnNaet)
	require.Len(t, settlement.SettlementHash, PosHashHexLength)

	_, err = SettleEvidenceCase(params, 12, EvidenceCase{
		EvidenceID:       "evidence-stale",
		ReporterID:       "reporter-a",
		ValidatorID:      "val-a",
		Misbehavior:      MisbehaviorDowntime,
		SlashFractionBps: 100,
		EvidenceHeight:   99,
		EvidenceEpoch:    6,
		Finalized:        true,
	}, sdkmath.NewInt(1_000), nil)
	require.ErrorContains(t, err, "outside slashable window")
}

func TestWorkloadRewardsSplitByRoleAndCompletedUnits(t *testing.T) {
	settlement, err := SettleWorkloadRewards(WorkloadRewardInput{
		EpochID:          9,
		TotalRewardsNaet: sdkmath.NewInt(1_000),
		RoleWeights: []RoleRewardWeight{
			{Role: ValidatorRoleBlockProducer, WeightBps: 5_000},
			{Role: ValidatorRoleVerifier, WeightBps: 5_000},
		},
		Outcomes: []AssignmentOutcome{
			{TaskID: "task-a", Role: ValidatorRoleBlockProducer, ValidatorID: "val-a", Completed: true, WorkUnits: 3},
			{TaskID: "task-a", Role: ValidatorRoleBlockProducer, ValidatorID: "val-b", Completed: true, WorkUnits: 1},
			{TaskID: "task-a", Role: ValidatorRoleVerifier, ValidatorID: "val-c", Completed: true, WorkUnits: 2},
			{TaskID: "task-a", Role: ValidatorRoleVerifier, ValidatorID: "val-d", Faulted: true, WorkUnits: 10},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(6), settlement.CompletedUnits)
	require.True(t, settlement.RemainderNaet.IsZero())
	require.Equal(t, []ValidatorWorkloadReward{
		{ValidatorID: "val-a", RewardNaet: sdkmath.NewInt(375), WorkUnits: 3},
		{ValidatorID: "val-b", RewardNaet: sdkmath.NewInt(125), WorkUnits: 1},
		{ValidatorID: "val-c", RewardNaet: sdkmath.NewInt(500), WorkUnits: 2},
	}, settlement.Rewards)
	require.Len(t, settlement.RewardRoot, PosHashHexLength)
}

func scoredCandidates(t *testing.T, params Params, candidates []Candidate) []ScoredValidator {
	t.Helper()
	validators := make([]ScoredValidator, len(candidates))
	for i, candidate := range candidates {
		scored, err := ScoreCandidate(params, candidate)
		require.NoError(t, err)
		validators[i] = scored
	}
	return validators
}

func posLayerSpecsByLayer(architecture LayeredPosArchitecture) map[PosLayer]PosLayerSpec {
	specs := make(map[PosLayer]PosLayerSpec, len(architecture.Layers))
	for _, layer := range architecture.Layers {
		specs[layer.Layer] = layer
	}
	return specs
}
