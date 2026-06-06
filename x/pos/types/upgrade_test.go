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

	closed, err := CloseEpochRecord(left, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseClosed, closed.Phase)
	require.Equal(t, SettlementStatusFinalized, closed.SettlementStatus)
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
