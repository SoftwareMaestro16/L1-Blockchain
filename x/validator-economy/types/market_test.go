package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestPropagateSlashAppliesProportionalAndFirstLossRules(t *testing.T) {
	delegations := []DelegationRecord{
		marketDelegation("del-a", "val-risk", 4_000, RiskTrancheSenior),
		marketDelegation("del-b", "val-risk", 6_000, RiskTrancheJunior),
	}

	proportional, err := PropagateSlash(SlashPropagationInput{
		Validator:        "val-risk",
		SelfBondNaet:     sdkmath.NewInt(1_000),
		Delegations:      delegations,
		SlashFractionBps: 1_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(100), proportional.SelfBondSlashedNaet)
	require.Equal(t, sdkmath.NewInt(400), proportional.DelegatorSlashes[0].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(600), proportional.DelegatorSlashes[1].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(1_100), proportional.TotalSlashedNaet)

	firstLoss, err := PropagateSlash(SlashPropagationInput{
		Validator:         "val-risk",
		SelfBondNaet:      sdkmath.NewInt(1_000),
		Delegations:       delegations,
		SlashFractionBps:  1_000,
		SelfBondFirstLoss: true,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), firstLoss.SelfBondSlashedNaet)
	require.Equal(t, sdkmath.NewInt(40), firstLoss.DelegatorSlashes[0].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(60), firstLoss.DelegatorSlashes[1].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(1_100), firstLoss.TotalSlashedNaet)

	accounting, err := BuildFirstLossSelfBondAccounting(SlashPropagationInput{
		Validator:         "val-risk",
		SelfBondNaet:      sdkmath.NewInt(1_000),
		Delegations:       delegations,
		SlashFractionBps:  1_000,
		SelfBondFirstLoss: true,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_100), accounting.TargetSlashNaet)
	require.Equal(t, sdkmath.NewInt(1_000), accounting.SelfBondAbsorbedNaet)
	require.Equal(t, sdkmath.NewInt(100), accounting.DelegatorResidualSlashNaet)
	require.True(t, accounting.FirstLossApplied)
}

func TestRouteSlashingCoversSeverityClassesAndFundRoutingInvariant(t *testing.T) {
	severities := []SlashSeverityClass{
		SlashSeverityMinorDowntime,
		SlashSeverityMajorDowntime,
		SlashSeverityRepeatedDowntime,
		SlashSeverityEquivocation,
		SlashSeverityEvidenceManipulation,
		SlashSeverityKeyCompromiseResponseFailure,
	}
	for _, severity := range severities {
		t.Run(string(severity), func(t *testing.T) {
			result, err := RouteSlashing(SlashingRoutingInput{
				Validator:      "val-slash",
				Severity:       severity,
				TotalStakeNaet: sdkmath.NewInt(1_000_000),
				Evidence: SlashingEvidence{
					EvidenceID: "evidence-1",
					ReporterID: "reporter-1",
					Accepted:   true,
				},
				CurrentEpoch: 10,
			})
			require.NoError(t, err)
			require.True(t, result.PenaltyNaet.IsPositive())
			require.True(t, result.ReporterPaid)
			require.Equal(t, result.PenaltyNaet, result.BurnNaet.Add(result.TreasuryNaet).Add(result.ReporterRewardNaet).Add(result.ValidatorResidualNaet))
			require.Equal(t, result.PenaltyNaet, result.Event.PenaltyNaet)
			require.Equal(t, result.BurnNaet, result.Event.BurnNaet)
			require.Equal(t, result.TreasuryNaet, result.Event.TreasuryNaet)
			require.Equal(t, result.ReporterRewardNaet, result.Event.ReporterRewardNaet)
		})
	}
}

func TestRouteSlashingPaysReporterOnlyForAcceptedNonDuplicateEvidence(t *testing.T) {
	for _, evidence := range []SlashingEvidence{
		{EvidenceID: "rejected", ReporterID: "reporter-1", Accepted: false},
		{EvidenceID: "duplicate", ReporterID: "reporter-1", Accepted: true, Duplicate: true},
		{EvidenceID: "missing-reporter", Accepted: true},
	} {
		result, err := RouteSlashing(SlashingRoutingInput{
			Validator:      "val-slash",
			Severity:       SlashSeverityEquivocation,
			TotalStakeNaet: sdkmath.NewInt(1_000_000),
			Evidence:       evidence,
			CurrentEpoch:   7,
		})
		require.NoError(t, err)
		require.False(t, result.ReporterPaid)
		require.Equal(t, sdkmath.ZeroInt(), result.ReporterRewardNaet)
		require.Equal(t, result.PenaltyNaet, result.BurnNaet.Add(result.TreasuryNaet).Add(result.ValidatorResidualNaet))
	}

	params := DefaultSlashingSeverityParams()
	params[3].TreasuryBps = 3_000
	params[3].ReporterRewardBps = 2_000
	params[3].ReporterRewardCapBps = 500
	capped, err := RouteSlashing(SlashingRoutingInput{
		Validator:      "val-slash",
		Severity:       SlashSeverityEquivocation,
		TotalStakeNaet: sdkmath.NewInt(1_000_000),
		Evidence:       SlashingEvidence{EvidenceID: "accepted", ReporterID: "reporter-1", Accepted: true},
		CurrentEpoch:   7,
		Params:         params,
	})
	require.NoError(t, err)
	require.Equal(t, mulIntBps(capped.PenaltyNaet, 500), capped.ReporterRewardNaet)
}

func TestRepeatOffenseMultiplierIsDeterministicBoundedAndDecays(t *testing.T) {
	left, err := ComputeRepeatOffenseMultiplier(10, []uint64{9, 6, 5, 2}, 4, 2_500, 15_000)
	require.NoError(t, err)
	right, err := ComputeRepeatOffenseMultiplier(10, []uint64{2, 5, 6, 9}, 4, 2_500, 15_000)
	require.NoError(t, err)
	require.Equal(t, left, right)
	require.Equal(t, uint32(15_000), left)

	decayed, err := ComputeRepeatOffenseMultiplier(10, []uint64{1, 2, 3}, 4, 2_500, 30_000)
	require.NoError(t, err)
	require.Equal(t, uint32(10_000), decayed)
}

func TestConcentrationReportExposesWarningsDampeningAndStakeMovementIncentives(t *testing.T) {
	params := testParams()
	candidates := []postypes.Candidate{
		marketCandidate("val-a", 1_000, 9_000, 1_000),
		marketCandidate("val-b", 1_000, 2_000, 500),
		marketCandidate("val-c", 1_000, 1_000, 500),
	}
	delegations := []DelegationRecord{
		marketDelegation("del-whale", "val-a", 8_000, ""),
		marketDelegation("del-small", "val-a", 1_000, ""),
		marketDelegation("del-b", "val-b", 2_000, ""),
		marketDelegation("del-c", "val-c", 1_000, ""),
	}
	state, err := NewValidatorMarketState(params, candidates, delegations, nil, nil, nil)
	require.NoError(t, err)

	report, err := state.QueryConcentrationReport(DecentralizationParams{
		TopN:                          2,
		MaxValidatorShareBps:          4_000,
		MaxTopNShareBps:               8_000,
		MaxDelegatorConcentrationBps:  6_000,
		MinSelfDelegationRatioBps:     1_500,
		MinSelfDelegationNaet:         sdkmath.NewInt(2_000),
		MaxCommissionWeightedBps:      700,
		RewardDampeningSafetyFloorBps: 7_500,
		StakeMovementIncentiveBps:     250,
	}, []string{"val-a", "val-b", "val-c"})
	require.NoError(t, err)
	require.Len(t, report.Metrics, 3)
	require.Equal(t, uint32(8_666), report.TopNShareBps)
	require.Contains(t, report.ActiveSetWarnings, ConcentrationWarningTopNShare)
	require.Contains(t, report.ActiveSetWarnings, ConcentrationWarningCommissionByPower)

	metric := report.Metrics[0]
	require.Equal(t, "val-a", metric.Validator)
	require.Equal(t, uint32(6_666), metric.VotingPowerShareBps)
	require.Equal(t, uint32(8_888), metric.DelegatorConcentrationBps)
	require.Equal(t, uint32(1_000), metric.SelfDelegationRatioBps)
	require.Contains(t, metric.Warnings, ConcentrationWarningValidatorShare)
	require.Contains(t, metric.Warnings, ConcentrationWarningDelegatorShare)
	require.Contains(t, metric.Warnings, ConcentrationWarningSelfDelegation)
	require.Contains(t, metric.Warnings, ConcentrationWarningRewardDampeningActive)
	require.Equal(t, uint32(250), metric.StakeMovementIncentiveBps)
	require.LessOrEqual(t, metric.RewardDampeningBps, uint32(2_500))
}

func TestDelegationMarketQueriesExposeRiskYieldSaturationAndHistory(t *testing.T) {
	params := testParams()
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = 50_000
	params.StakeSaturationNaet = sdkmath.NewInt(100_000)
	candidate := marketCandidate("val-risk", 1_000, 10_000, 700)
	delegations := []DelegationRecord{
		marketDelegation("del-a", "val-risk", 4_000, RiskTrancheSenior),
		marketDelegation("del-b", "val-risk", 6_000, RiskTrancheJunior),
	}
	score := testRecord(4, "val-risk", 8_000)
	score.ReliabilityIndex = 8_000
	slash := ValidatorSlashHistoryRecord{
		EpochID:              3,
		Height:               33,
		Validator:            "val-risk",
		Misbehavior:          postypes.MisbehaviorDoubleSign,
		SlashFractionBps:     1_000,
		SelfBondSlashedNaet:  sdkmath.NewInt(200),
		DelegatorSlashedNaet: sdkmath.NewInt(800),
		TotalSlashedNaet:     sdkmath.NewInt(1_000),
	}
	commissions := []ValidatorCommissionRecord{
		{EpochID: 1, Height: 10, Validator: "val-risk", CommissionBps: 500},
		{EpochID: 4, Height: 40, Validator: "val-risk", CommissionBps: 700},
	}
	state, err := NewValidatorMarketState(params, []postypes.Candidate{candidate}, delegations, []ValidatorScoreRecord{score}, []ValidatorSlashHistoryRecord{slash}, commissions)
	require.NoError(t, err)

	risk, found := state.QueryValidatorRisk("val-risk")
	require.True(t, found)
	require.Equal(t, uint32(1), risk.SlashEventCount)
	require.Equal(t, sdkmath.NewInt(1_000), risk.TotalSlashedNaet)
	require.Equal(t, uint32(8_000), risk.LatestReliabilityBps)
	require.Equal(t, uint32(3_000), risk.RiskScoreBps)
	require.True(t, risk.DelegatorRiskInherited)

	yield, found, err := state.QueryValidatorEffectiveYield("val-risk", sdkmath.NewInt(1_100))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(11_000), yield.RawStakeNaet)
	require.Equal(t, uint32(1_000), yield.GrossYieldBps)
	require.Equal(t, uint32(930), yield.NetYieldBps)
	require.Equal(t, uint32(700), yield.CommissionBps)
	require.True(t, yield.SaturationDampeningBps < postypes.BasisPoints)

	saturation, found, err := state.QueryValidatorSaturation("val-risk")
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, saturation.Saturated)
	require.Equal(t, sdkmath.NewInt(5_000), saturation.SaturationCapNaet)

	exposure, found, err := state.QueryDelegationRiskExposure("del-a", "val-risk", 1_000, true)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(40), exposure.ProjectedSlashNaet)
	require.Equal(t, sdkmath.NewInt(360), exposure.FirstLossProtectedNaet)
	require.Equal(t, sdkmath.NewInt(320), exposure.HistoricalSlashNaet)
	require.True(t, exposure.AdvisoryRiskProfile)
	require.Len(t, exposure.SlashEventsInherited, 1)

	activation, found := state.QueryDelegationActivationEpoch("del-a", "val-risk")
	require.True(t, found)
	require.Equal(t, uint64(11), activation)

	commissionStatus, found, err := state.QueryDelegationCommissionStatus("del-a", "val-risk", 44, true)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, DelegationStatusActive, commissionStatus.Status)
	require.False(t, commissionStatus.CommissionExceeded)
	require.Nil(t, commissionStatus.Alert)

	lockEligibility, found, err := state.QueryDelegationLockEligibility("del-a", "val-risk", params.EvidenceWindowEpochs)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, LockDurationEpoch, lockEligibility.LockDurationPreference)
	require.False(t, lockEligibility.EligibleForRewardMultiplier)
	require.Equal(t, params.UnbondingSeconds, lockEligibility.EffectiveUnbondingSeconds)

	require.Equal(t, commissions, state.QueryValidatorCommissionHistory("val-risk"))
	require.Equal(t, []ValidatorSlashHistoryRecord{slash}, state.QueryValidatorSlashHistory("val-risk"))
	require.Equal(t, []ValidatorScoreRecord{score}, state.QueryValidatorPerformanceHistory("val-risk"))
}

func TestDelegationMarketCommissionQueryEmitsAlertWithoutMovingStake(t *testing.T) {
	params := testParams()
	candidate := marketCandidate("val-commission", 1_000, 1_000, 500)
	delegation := marketDelegation("del-a", "val-commission", 1_000, "")
	delegation.CommissionTolerance = 600
	state, err := NewValidatorMarketState(params, []postypes.Candidate{candidate}, []DelegationRecord{delegation}, nil, nil, []ValidatorCommissionRecord{
		{EpochID: 5, Height: 50, Validator: "val-commission", CommissionBps: 700},
	})
	require.NoError(t, err)

	status, found, err := state.QueryDelegationCommissionStatus("del-a", "val-commission", 55, true)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, status.CommissionExceeded)
	require.Equal(t, DelegationStatusCommissionExceeded, status.Status)
	require.NotNil(t, status.Alert)
	require.True(t, status.Alert.RedelegationAdvisory)
	require.Equal(t, sdkmath.NewInt(1_000), state.totalDelegatedAtValidator("val-commission"))
}

func TestDelegationRiskExposureSurvivesRedelegationRecords(t *testing.T) {
	params := testParams()
	oldRecord := marketDelegation("del-a", "val-old", 1_000, "")
	newRecord := marketDelegation("del-a", "val-new", 1_000, "")
	oldCandidate := marketCandidate("val-old", 1_000, 1_000, 500)
	newCandidate := marketCandidate("val-new", 1_000, 1_000, 500)
	slash := ValidatorSlashHistoryRecord{
		EpochID:              8,
		Height:               80,
		Validator:            "val-old",
		Misbehavior:          postypes.MisbehaviorDowntime,
		SlashFractionBps:     500,
		SelfBondSlashedNaet:  sdkmath.NewInt(50),
		DelegatorSlashedNaet: sdkmath.NewInt(50),
		TotalSlashedNaet:     sdkmath.NewInt(100),
	}
	state, err := NewValidatorMarketState(params, []postypes.Candidate{oldCandidate, newCandidate}, []DelegationRecord{oldRecord, newRecord}, nil, []ValidatorSlashHistoryRecord{slash}, nil)
	require.NoError(t, err)

	oldExposure, found, err := state.QueryDelegationRiskExposure("del-a", "val-old", 500, false)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdkmath.NewInt(50), oldExposure.HistoricalSlashNaet)

	newExposure, found, err := state.QueryDelegationRiskExposure("del-a", "val-new", 500, false)
	require.NoError(t, err)
	require.True(t, found)
	require.Empty(t, newExposure.SlashEventsInherited)
	require.Equal(t, sdkmath.ZeroInt(), newExposure.HistoricalSlashNaet)
}

func marketDelegation(delegator string, validator string, amount int64, tranche string) DelegationRecord {
	return DelegationRecord{
		Delegator:              delegator,
		Validator:              validator,
		Amount:                 sdkmath.NewInt(amount),
		ActivationEpoch:        11,
		RiskAppetite:           RiskAppetiteBalanced,
		CommissionTolerance:    1_000,
		LockDurationPreference: LockDurationEpoch,
		RewardStrategy:         RewardStrategyLiquid,
		RiskTrancheOptional:    tranche,
		CreatedHeight:          10,
		UpdatedHeight:          10,
	}
}

func marketCandidate(id string, selfStake int64, delegatedStake int64, commissionBps uint32) postypes.Candidate {
	nominations := []postypes.Nomination(nil)
	if delegatedStake > 0 {
		nominations = []postypes.Nomination{{NominatorID: "market-delegators", StakeNaet: sdkmath.NewInt(delegatedStake)}}
	}
	return postypes.Candidate{
		ValidatorID:         id,
		SelfStakeNaet:       sdkmath.NewInt(selfStake),
		DelegatedStakeNaet:  sdkmath.NewInt(delegatedStake),
		PerformanceScoreBps: postypes.BasisPoints,
		UptimeFactorBps:     postypes.BasisPoints,
		CommissionBps:       commissionBps,
		Nominations:         nominations,
	}
}
