package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestBuildValidatorScoreRecordCapturesDeterministicComponents(t *testing.T) {
	params := testParams()
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 20_000
	params.StakeSaturationNaet = sdkmath.NewInt(10_000)
	candidate := testCandidate("val-record", 5_000)
	candidate.PerformanceScoreBps = 9_000
	candidate.UptimeFactorBps = 9_500
	candidate.LatencyFactorBps = 8_000
	candidate.ReliabilityIndexBps = 7_000

	record, err := BuildValidatorScoreRecord(12, params, candidate)
	require.NoError(t, err)
	require.Equal(t, uint64(12), record.EpochID)
	require.Equal(t, "val-record", record.ValidatorAddress)
	require.Equal(t, sdkmath.NewInt(5_000), record.RawStake)
	require.Equal(t, sdkmath.NewInt(2_000), record.EffectiveStake)
	require.Equal(t, sdkmath.NewInt(2_000), record.StakeWeight)
	require.Equal(t, uint32(9_000), record.PerformanceFactor)
	require.Equal(t, uint32(9_500), record.UptimeFactor)
	require.Equal(t, uint32(8_000), record.LatencyFactor)
	require.Equal(t, uint32(7_000), record.ReliabilityIndex)
	require.Equal(t, sdkmath.NewInt(957), record.ValidatorScore)
	require.Equal(t, SaturationStatusSaturated, record.SaturationStatus)
	require.Equal(t, DefaultScoreVersion, record.ScoreVersion)
}

func TestScoreComponentStateQueriesHistoricalRecords(t *testing.T) {
	params := testParams()
	a, err := BuildValidatorScoreRecord(3, params, testCandidate("val-b", 2_000))
	require.NoError(t, err)
	b, err := BuildValidatorScoreRecord(2, params, testCandidate("val-a", 1_000))
	require.NoError(t, err)

	state, err := NewScoreComponentState([]ValidatorScoreRecord{a, b})
	require.NoError(t, err)
	found, ok := state.GetScoreRecord(2, "val-a")
	require.True(t, ok)
	require.Equal(t, b, found)
	require.Equal(t, []ValidatorScoreRecord{b}, state.RecordsForEpoch(2))

	_, err = NewScoreComponentState([]ValidatorScoreRecord{b, b})
	require.ErrorContains(t, err, "duplicate score record")
}

func TestElectionRankingOrdersByScoreAndReportsRejectedCandidates(t *testing.T) {
	params := testParams()
	candidates := []postypes.Candidate{
		testCandidate("val-low", 1_000),
		testCandidate("val-high", 3_000),
		testCandidate("val-jailed", 9_000),
	}
	candidates[2].Jailed = true

	ranking, err := BuildElectionRanking(5, params, candidates, 2)
	require.NoError(t, err)
	require.Equal(t, []string{"val-high", "val-low"}, recordIDs(ranking.Records))
	require.Len(t, ranking.Rejected, 1)
	require.Equal(t, "val-jailed", ranking.Rejected[0].ValidatorAddress)
	require.Contains(t, ranking.Rejected[0].Reason, "jailed")
}

func TestValidatorSetTransitionLimitDefersExcessNewValidators(t *testing.T) {
	params := testParams()
	params.MaxValidatorSetChangeRateBps = 100
	previous := make([]string, 75)
	for i := range previous {
		previous[i] = fmt.Sprintf("old-%03d", i)
	}
	records := []ValidatorScoreRecord{
		testRecord(7, "new-a", 5_000),
		testRecord(7, "new-b", 4_000),
		testRecord(7, "new-c", 3_000),
	}
	ranking := ElectionRanking{EpochID: 7, Records: records, RequestedValidatorCount: 3}

	limited, err := ApplyValidatorSetTransitionLimit(params, previous, ranking)
	require.NoError(t, err)
	require.Equal(t, uint32(1), limited.MaxValidatorSetChanges)
	require.True(t, limited.TransitionLimited)
	require.Equal(t, []string{"new-a"}, recordIDs(limited.Records))
}

func TestScoreSimulationFlagsCentralizationAfterSaturation(t *testing.T) {
	params := testParams()
	params.MaxVotingPowerBps = 3_000
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 2_0000
	candidates := []postypes.Candidate{
		testCandidate("val-whale", 9_000),
		testCandidate("val-a", 1_000),
		testCandidate("val-b", 1_000),
		testCandidate("val-c", 1_000),
	}

	result, err := SimulateScores(ScoreSimulationInput{
		EpochID:	8,
		Params:		params,
		Candidates:	candidates,
		TargetActive:	4,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(7_500), result.MaxRawStakeShareBps)
	require.Equal(t, uint32(4_000), result.MaxEffectiveShareBps)
	require.True(t, result.CentralizationWarning)
	require.Equal(t, SaturationStatusSaturated, result.Ranking.Records[0].SaturationStatus)
}

func TestStakeSplittingImprovesEffectiveWeightOnlyThroughDistribution(t *testing.T) {
	params := testParams()
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 30_000
	single, err := SimulateScores(ScoreSimulationInput{
		EpochID:	9,
		Params:		params,
		Candidates:	[]postypes.Candidate{testCandidate("val-whale", 9_000)},
		TargetActive:	1,
	})
	require.NoError(t, err)
	split, err := SimulateScores(ScoreSimulationInput{
		EpochID:	9,
		Params:		params,
		Candidates: []postypes.Candidate{
			testCandidate("val-a", 3_000),
			testCandidate("val-b", 3_000),
			testCandidate("val-c", 3_000),
		},
		TargetActive:	3,
	})
	require.NoError(t, err)

	require.Equal(t, sdkmath.NewInt(9_000), single.TotalRawStakeNaet)
	require.Equal(t, sdkmath.NewInt(3_000), single.TotalEffectiveNaet)
	require.Equal(t, sdkmath.NewInt(9_000), split.TotalRawStakeNaet)
	require.Equal(t, sdkmath.NewInt(9_000), split.TotalEffectiveNaet)
	require.Equal(t, []string{"val-a", "val-b", "val-c"}, split.ActiveValidatorIDs)
}

func TestValidatorEligibilityScoreIsDeterministicAndAuditable(t *testing.T) {
	params := testParams()
	gov := DefaultValidatorEconomyGovernanceParams(params)
	input := ValidatorSelectionInput{
		EpochID:			11,
		ValidatorAddress:		"val-score",
		BondedStakeNaet:		sdkmath.NewInt(50_000),
		SelfDelegationNaet:		params.MinStakeNaet,
		UptimeBps:			9_900,
		MissedBlockRateBps:		100,
		SlashHistoryCount:		1,
		CommissionBps:			500,
		StakeConcentrationBps:		1_000,
		MetadataCompletenessBps:	postypes.BasisPoints,
	}

	left, err := ComputeValidatorEligibilityScore(params, gov, input)
	require.NoError(t, err)
	right, err := ComputeValidatorEligibilityScore(params, gov, input)
	require.NoError(t, err)
	require.Equal(t, left, right)
	require.True(t, left.Eligible)
	require.Equal(t, uint32(8_755), left.ScoreBps)
	require.Equal(t, uint32(5_000), left.BondedStakeComponentBps)
	require.Equal(t, uint32(postypes.BasisPoints), left.SelfDelegationComponentBps)
	require.Equal(t, uint32(8_000), left.SlashHistoryComponentBps)
	require.Empty(t, left.Reasons)
}

func TestEpochSelectionEventsExposeScoreComponents(t *testing.T) {
	params := testParams()
	candidates := []postypes.Candidate{
		qualifiedCandidate("val-b", 2_000),
		qualifiedCandidate("val-a", 3_000),
	}
	ranking, err := BuildElectionRanking(12, params, candidates, 1)
	require.NoError(t, err)

	left, err := BuildEpochSelectionEvents(12, params, DefaultValidatorEconomyGovernanceParams(params), ranking, candidates)
	require.NoError(t, err)
	right, err := BuildEpochSelectionEvents(12, params, DefaultValidatorEconomyGovernanceParams(params), ranking, candidates)
	require.NoError(t, err)
	require.Equal(t, left, right)
	require.Equal(t, []string{"val-a", "val-b"}, selectionEventIDs(left))
	require.True(t, left[0].Selected)
	require.Equal(t, DefaultScoreVersion, left[0].ScoreVersion)
	require.Equal(t, uint32(postypes.BasisPoints), left[0].Score.MetadataComponentBps)
	require.NotZero(t, left[0].Score.ScoreBps)
}

func TestValidatorChurnSimulationsCoverNormalAdversarialAndLowParticipation(t *testing.T) {
	params := testParams()
	params.MaxValidatorSetChangeRateBps = 100
	previous := make([]string, 75)
	candidates := make([]postypes.Candidate, 0, 76)
	for i := range previous {
		id := fmt.Sprintf("old-%03d", i)
		previous[i] = id
		candidates = append(candidates, qualifiedCandidate(id, 1_000))
	}
	candidates = append(candidates, qualifiedCandidate("new-a", 5_000))
	normal, err := SimulateValidatorChurn(ValidatorChurnSimulationInput{
		Scenario:	ChurnScenarioNormal,
		ScoreInput: ScoreSimulationInput{
			EpochID:	13,
			Params:		params,
			Candidates:	candidates,
			PreviousActive:	previous,
			TargetActive:	75,
		},
	})
	require.NoError(t, err)
	require.True(t, normal.Passed)
	require.Equal(t, uint32(1), normal.NewValidatorCount)
	require.False(t, normal.TransitionLimited)

	adversarialCandidates := append([]postypes.Candidate{qualifiedCandidate("whale", 100_000)}, candidates[:74]...)
	adversarial, err := SimulateValidatorChurn(ValidatorChurnSimulationInput{
		Scenario:	ChurnScenarioAdversarial,
		ScoreInput: ScoreSimulationInput{
			EpochID:	14,
			Params:		params,
			Candidates:	adversarialCandidates,
			TargetActive:	75,
		},
	})
	require.NoError(t, err)
	require.False(t, adversarial.Passed)
	require.Contains(t, adversarial.Warnings, "adversarial_concentration_warning")

	low, err := SimulateValidatorChurn(ValidatorChurnSimulationInput{
		Scenario:	ChurnScenarioLowParticipation,
		ScoreInput: ScoreSimulationInput{
			EpochID:	15,
			Params:		params,
			Candidates:	candidates[:10],
			TargetActive:	75,
		},
	})
	require.NoError(t, err)
	require.False(t, low.Passed)
	require.Contains(t, low.Warnings, "low_participation_active_set_below_minimum")
}

func TestValidatorRewardAdjustmentDampensConcentrationAndReliability(t *testing.T) {
	gov := DefaultValidatorEconomyGovernanceParams(postypes.DefaultParams())
	gov.MinSelfDelegationNaet = sdkmath.NewInt(100)
	gov.ConcentrationSoftCapBps = 3_000
	record := testRecord(16, "val-risk", 6_000)
	record.PerformanceFactor = 9_300
	record.UptimeFactor = 9_500
	record.ReliabilityIndex = 9_700

	adjusted, err := ComputeValidatorRewardAdjustment(ValidatorRewardAdjustmentInput{
		Record:			record,
		TotalActiveStakeNaet:	sdkmath.NewInt(10_000),
		OperatingCostNaet:	sdkmath.NewInt(100),
		ExpectedRewardNaet:	sdkmath.NewInt(120),
		ValidatorAgeEpochs:	10,
		BootstrapQualified:	false,
		Governance:		gov,
	})
	require.NoError(t, err)
	require.False(t, adjusted.FullRewardEligible)
	require.Equal(t, uint32(200), adjusted.ReliabilityAdjustmentBps)
	require.Equal(t, uint32(1_071), adjusted.ConcentrationDampeningBps)
	require.Equal(t, uint32(8_729), adjusted.RewardMultiplierBps)
	require.True(t, adjusted.AdjustedRewardWeightNaet.LT(adjusted.BaseRewardWeightNaet))
	require.Equal(t, int32(2_000), adjusted.ProfitabilityMarginBps)
	require.True(t, adjusted.VisibleBeforeDelegation)
}

func TestValidatorRewardAdjustmentCannotImproveByReducingUptime(t *testing.T) {
	gov := DefaultValidatorEconomyGovernanceParams(postypes.DefaultParams())
	gov.MinSelfDelegationNaet = sdkmath.NewInt(100)
	healthy := testRecord(17, "val-reliable", 1_000)
	weaker := healthy
	weaker.UptimeFactor = 9_000

	healthyReward, err := ComputeValidatorRewardAdjustment(ValidatorRewardAdjustmentInput{
		Record:			healthy,
		TotalActiveStakeNaet:	sdkmath.NewInt(10_000),
		ExpectedRewardNaet:	sdkmath.NewInt(100),
		Governance:		gov,
	})
	require.NoError(t, err)
	weakerReward, err := ComputeValidatorRewardAdjustment(ValidatorRewardAdjustmentInput{
		Record:			weaker,
		TotalActiveStakeNaet:	sdkmath.NewInt(10_000),
		ExpectedRewardNaet:	sdkmath.NewInt(100),
		Governance:		gov,
	})
	require.NoError(t, err)
	require.Greater(t, healthyReward.RewardMultiplierBps, weakerReward.RewardMultiplierBps)
}

func TestValidatorBootstrapBonusExpiresAutomatically(t *testing.T) {
	gov := DefaultValidatorEconomyGovernanceParams(postypes.DefaultParams())
	gov.MinSelfDelegationNaet = sdkmath.NewInt(100)
	record := testRecord(18, "val-new", 100)

	active, err := ComputeValidatorRewardAdjustment(ValidatorRewardAdjustmentInput{
		Record:			record,
		TotalActiveStakeNaet:	sdkmath.NewInt(10_000),
		ExpectedRewardNaet:	sdkmath.NewInt(10),
		ValidatorAgeEpochs:	1,
		BootstrapQualified:	true,
		Governance:		gov,
	})
	require.NoError(t, err)
	expired, err := ComputeValidatorRewardAdjustment(ValidatorRewardAdjustmentInput{
		Record:			record,
		TotalActiveStakeNaet:	sdkmath.NewInt(10_000),
		ExpectedRewardNaet:	sdkmath.NewInt(10),
		ValidatorAgeEpochs:	gov.BootstrapMaxEpochs,
		BootstrapQualified:	true,
		Governance:		gov,
	})
	require.NoError(t, err)
	require.Equal(t, gov.BootstrapBonusBps, active.BootstrapBonusBps)
	require.False(t, active.BootstrapExpired)
	require.Zero(t, expired.BootstrapBonusBps)
	require.True(t, expired.BootstrapExpired)
	require.Greater(t, active.RewardMultiplierBps, expired.RewardMultiplierBps)
}

func testParams() postypes.Params {
	params := postypes.DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationNaet = sdkmath.NewInt(100_000)
	return params
}

func testCandidate(id string, stake int64) postypes.Candidate {
	return postypes.Candidate{
		ValidatorID:		id,
		SelfStakeNaet:		sdkmath.NewInt(stake),
		DelegatedStakeNaet:	sdkmath.ZeroInt(),
		PerformanceScoreBps:	postypes.BasisPoints,
		UptimeFactorBps:	postypes.BasisPoints,
		CommissionBps:		500,
	}
}

func qualifiedCandidate(id string, stake int64) postypes.Candidate {
	candidate := testCandidate(id, stake)
	candidate.Roles = []postypes.ValidatorRole{postypes.ValidatorRoleBlockProducer}
	candidate.Capacity = postypes.ValidatorCapacity{
		MaxTaskGroups:		1,
		SupportedWorkloads:	[]postypes.WorkloadType{postypes.WorkloadTypeGlobalConsensus},
	}
	return candidate
}

func testRecord(epochID uint64, id string, score int64) ValidatorScoreRecord {
	return ValidatorScoreRecord{
		EpochID:		epochID,
		ValidatorAddress:	id,
		RawStake:		sdkmath.NewInt(score),
		EffectiveStake:		sdkmath.NewInt(score),
		StakeWeight:		sdkmath.NewInt(score),
		PerformanceFactor:	postypes.BasisPoints,
		UptimeFactor:		postypes.BasisPoints,
		LatencyFactor:		postypes.BasisPoints,
		ReliabilityIndex:	postypes.BasisPoints,
		ValidatorScore:		sdkmath.NewInt(score),
		SaturationStatus:	SaturationStatusNone,
		ScoreVersion:		DefaultScoreVersion,
	}
}

func recordIDs(records []ValidatorScoreRecord) []string {
	ids := make([]string, len(records))
	for i, record := range records {
		ids[i] = record.ValidatorAddress
	}
	return ids
}

func selectionEventIDs(events []EpochSelectionEvent) []string {
	ids := make([]string, len(events))
	for i, event := range events {
		ids[i] = event.ValidatorAddress
	}
	return ids
}
