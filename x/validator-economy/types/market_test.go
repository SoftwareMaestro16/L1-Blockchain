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
		Validator:		"val-risk",
		SelfBondNaet:		sdkmath.NewInt(1_000),
		Delegations:		delegations,
		SlashFractionBps:	1_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(100), proportional.SelfBondSlashedNaet)
	require.Equal(t, sdkmath.NewInt(400), proportional.DelegatorSlashes[0].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(600), proportional.DelegatorSlashes[1].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(1_100), proportional.TotalSlashedNaet)

	firstLoss, err := PropagateSlash(SlashPropagationInput{
		Validator:		"val-risk",
		SelfBondNaet:		sdkmath.NewInt(1_000),
		Delegations:		delegations,
		SlashFractionBps:	1_000,
		SelfBondFirstLoss:	true,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), firstLoss.SelfBondSlashedNaet)
	require.Equal(t, sdkmath.NewInt(40), firstLoss.DelegatorSlashes[0].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(60), firstLoss.DelegatorSlashes[1].SlashedNaet)
	require.Equal(t, sdkmath.NewInt(1_100), firstLoss.TotalSlashedNaet)

	accounting, err := BuildFirstLossSelfBondAccounting(SlashPropagationInput{
		Validator:		"val-risk",
		SelfBondNaet:		sdkmath.NewInt(1_000),
		Delegations:		delegations,
		SlashFractionBps:	1_000,
		SelfBondFirstLoss:	true,
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
				Validator:	"val-slash",
				Severity:	severity,
				TotalStakeNaet:	sdkmath.NewInt(1_000_000),
				Evidence: SlashingEvidence{
					EvidenceID:	"evidence-1",
					ReporterID:	"reporter-1",
					Accepted:	true,
				},
				CurrentEpoch:	10,
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
			Validator:	"val-slash",
			Severity:	SlashSeverityEquivocation,
			TotalStakeNaet:	sdkmath.NewInt(1_000_000),
			Evidence:	evidence,
			CurrentEpoch:	7,
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
		Validator:	"val-slash",
		Severity:	SlashSeverityEquivocation,
		TotalStakeNaet:	sdkmath.NewInt(1_000_000),
		Evidence:	SlashingEvidence{EvidenceID: "accepted", ReporterID: "reporter-1", Accepted: true},
		CurrentEpoch:	7,
		Params:		params,
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
		TopN:				2,
		MaxValidatorShareBps:		4_000,
		MaxTopNShareBps:		8_000,
		MaxDelegatorConcentrationBps:	6_000,
		MinSelfDelegationRatioBps:	1_500,
		MinSelfDelegationNaet:		sdkmath.NewInt(2_000),
		MaxCommissionWeightedBps:	700,
		RewardDampeningSafetyFloorBps:	7_500,
		StakeMovementIncentiveBps:	250,
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

func TestDelegatorValidatorProfileExposesRiskAdjustedYieldDisclosureAndPolicies(t *testing.T) {
	params := testParams()
	params.MaxVotingPowerBps = 6_000
	candidates := []postypes.Candidate{
		marketCandidate("val-safe", 2_000, 3_000, 500),
		marketCandidate("val-peer", 2_500, 2_500, 500),
	}
	delegations := []DelegationRecord{
		marketDelegation("del-a", "val-safe", 1_000, ""),
		marketDelegation("del-b", "val-safe", 2_000, ""),
		marketDelegation("del-peer", "val-peer", 2_500, ""),
	}
	score := testRecord(5, "val-safe", 5_000)
	score.PerformanceFactor = 9_600
	score.UptimeFactor = 9_800
	score.ReliabilityIndex = 9_700
	slash := ValidatorSlashHistoryRecord{
		EpochID:		3,
		Height:			30,
		Validator:		"val-safe",
		Misbehavior:		postypes.MisbehaviorDowntime,
		SlashFractionBps:	100,
		SelfBondSlashedNaet:	sdkmath.NewInt(20),
		DelegatorSlashedNaet:	sdkmath.NewInt(80),
		TotalSlashedNaet:	sdkmath.NewInt(100),
	}
	state, err := NewValidatorMarketState(params, candidates, delegations, []ValidatorScoreRecord{score}, []ValidatorSlashHistoryRecord{slash}, nil)
	require.NoError(t, err)
	decParams := DefaultDecentralizationParams(params)
	decParams.MaxTopNShareBps = postypes.BasisPoints
	decParams.MaxDelegatorConcentrationBps = 7_000

	profile, found, err := state.QueryDelegatorValidatorProfile("del-a", "val-safe", sdkmath.NewInt(1_000), sdkmath.NewInt(500), decParams, []string{"val-safe", "val-peer"})
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, profile.AdvisoryOnly)
	require.Equal(t, uint32(1_300), profile.Risk.RiskScoreBps)
	require.Equal(t, uint32(1_000), profile.RiskComponents.SlashHistoryRiskBps)
	require.Equal(t, uint32(300), profile.RiskComponents.ReliabilityRiskBps)
	require.Equal(t, uint32(1_800), profile.RiskComponents.TotalRiskScoreBps)

	require.True(t, profile.YieldEstimate.UsesDistributionInputs)
	require.Equal(t, sdkmath.NewInt(500), profile.YieldEstimate.RewardInputNaet)
	require.Equal(t, sdkmath.NewInt(480), profile.YieldEstimate.AdjustedRewardInputNaet)
	require.Equal(t, sdkmath.NewInt(91), profile.YieldEstimate.EstimatedRewardNaet)
	require.Equal(t, uint32(1_000), profile.YieldEstimate.GrossYieldBps)
	require.Equal(t, uint32(910), profile.YieldEstimate.NetYieldBps)
	require.Equal(t, uint32(500), profile.YieldEstimate.CommissionBps)
	require.Equal(t, uint32(9_600), profile.YieldEstimate.PerformanceAdjustmentBps)
	require.Zero(t, profile.YieldEstimate.ConcentrationAdjustmentBps)

	require.Equal(t, uint32(500), profile.Disclosure.CommissionBps)
	require.Equal(t, uint32(1_500), profile.Disclosure.MaxCommissionChangeBps)
	require.Equal(t, uint32(9_800), profile.Disclosure.UptimeBps)
	require.Equal(t, uint32(1), profile.Disclosure.SlashHistoryCount)
	require.Equal(t, sdkmath.NewInt(2_000), profile.Disclosure.SelfDelegationNaet)
	require.Equal(t, ConcentrationStatusNormal, profile.Disclosure.ConcentrationStatus)

	evaluations := policyEvaluationMap(profile.PolicyEvaluations)
	require.True(t, evaluations[DelegationPolicyLowRisk].Matches)
	require.False(t, evaluations[DelegationPolicyHighAvailability].Matches)
	require.Contains(t, evaluations[DelegationPolicyHighAvailability].Reasons, "uptime_below_policy_minimum")
	require.True(t, evaluations[DelegationPolicyLowRisk].AdvisoryOnly)
}

func TestRedelegationRewardPreviewIsAdvisoryAndDoesNotMoveStake(t *testing.T) {
	params := testParams()
	params.MaxVotingPowerBps = postypes.BasisPoints
	candidates := []postypes.Candidate{
		marketCandidate("val-from", 1_000, 4_000, 1_000),
		marketCandidate("val-to", 2_000, 1_000, 0),
	}
	delegations := []DelegationRecord{
		marketDelegation("del-a", "val-from", 1_000, ""),
		marketDelegation("del-other", "val-from", 3_000, ""),
		marketDelegation("del-target", "val-to", 1_000, ""),
	}
	scores := []ValidatorScoreRecord{
		testRecord(7, "val-from", 5_000),
		testRecord(7, "val-to", 3_000),
	}
	state, err := NewValidatorMarketState(params, candidates, delegations, scores, nil, nil)
	require.NoError(t, err)
	beforeFrom := state.totalDelegatedAtValidator("val-from")
	beforeTo := state.totalDelegatedAtValidator("val-to")
	decParams := DefaultDecentralizationParams(params)
	decParams.MaxTopNShareBps = postypes.BasisPoints

	preview, found, err := state.QueryRedelegationRewardPreview("del-a", "val-from", "val-to", sdkmath.NewInt(1_000), sdkmath.NewInt(500), decParams, []string{"val-from", "val-to"})
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, preview.AdvisoryOnly)
	require.False(t, preview.StakeMovementExecuted)
	require.Equal(t, sdkmath.NewInt(90), preview.CurrentEstimate.EstimatedRewardNaet)
	require.Equal(t, sdkmath.NewInt(125), preview.TargetEstimate.EstimatedRewardNaet)
	require.Equal(t, sdkmath.NewInt(35), preview.RewardDeltaNaet)
	require.Equal(t, int32(350), preview.NetYieldDeltaBps)
	require.Equal(t, beforeFrom, state.totalDelegatedAtValidator("val-from"))
	require.Equal(t, beforeTo, state.totalDelegatedAtValidator("val-to"))
}

func TestValidatorCaptureSignalsEmitMachineReadableEventsAndAdvisoryAlerts(t *testing.T) {
	metadata, changed, err := TrackValidatorMetadataChange(ValidatorMetadataChangeInput{
		EpochID:	10,
		Height:		100,
		Validator:	"val-capture",
		Previous: ValidatorMetadataSnapshot{
			OperatorID:	"operator-a",
			ConsensusKeyID:	"key-a",
			Moniker:	"validator-a",
			PayoutAddress:	"payout-a",
		},
		Current: ValidatorMetadataSnapshot{
			OperatorID:	"operator-b",
			ConsensusKeyID:	"key-b",
			Moniker:	"validator-b",
			PayoutAddress:	"payout-b",
		},
		CooldownEpochs:	2,
	})
	require.NoError(t, err)
	require.True(t, changed)
	require.True(t, metadata.Material)
	require.Equal(t, uint64(12), metadata.CooldownUntilEpoch)
	require.Equal(t, ValidatorEventMetadataChange, metadata.Event.Type)
	require.Equal(t, []string{"consensus_key_id", "moniker", "operator_id", "payout_address"}, metadata.ChangedFields)

	commission, changed, err := TrackValidatorCommissionChange(CommissionChangeInput{
		EpochID:			10,
		Height:				101,
		Validator:			"val-capture",
		PreviousCommissionBps:		500,
		NewCommissionBps:		1_400,
		MaxIncreaseBpsPerInterval:	300,
		WarningPeriodEpochs:		2,
	})
	require.NoError(t, err)
	require.True(t, changed)
	require.True(t, commission.RiskFlag)
	require.Equal(t, uint64(12), commission.EffectiveEpoch)
	require.Equal(t, ValidatorEventCommissionChange, commission.Event.Type)

	report, err := EvaluateValidatorCaptureRisk(CaptureRiskInput{
		Validator:		"val-capture",
		CurrentEpoch:		11,
		Height:			110,
		PreviousCandidate:	marketCandidate("val-capture", 5_000, 1_000, 500),
		CurrentCandidate:	marketCandidate("val-capture", 3_000, 6_000, 1_400),
		MetadataChanges:	[]ValidatorMetadataChangeRecord{metadata},
		CommissionHistory: []ValidatorCommissionRecord{
			{EpochID: 9, Height: 90, Validator: "val-capture", CommissionBps: 500},
			{EpochID: 11, Height: 110, Validator: "val-capture", CommissionBps: 1_400},
		},
		SlashHistory: []ValidatorSlashHistoryRecord{
			{EpochID: 10, Height: 100, Validator: "val-capture", Misbehavior: postypes.MisbehaviorDowntime, SlashFractionBps: 100, SelfBondSlashedNaet: sdkmath.NewInt(10), DelegatorSlashedNaet: sdkmath.NewInt(10), TotalSlashedNaet: sdkmath.NewInt(20)},
		},
		Params: CaptureRiskParams{
			MaterialChangeCooldownEpochs:		2,
			CommissionChangeIntervalEpochs:		4,
			MaxCommissionIncreaseBpsPerInterval:	300,
			SuddenDelegationInflowBps:		3_000,
			SelfDelegationWithdrawalBps:		2_500,
			RecentSlashWindowEpochs:		4,
			HighRiskIndicatorThreshold:		2,
		},
	})
	require.NoError(t, err)
	require.True(t, report.HighRisk)
	require.True(t, report.AdvisoryOnly)
	require.False(t, report.SlashableBehavior)
	require.Len(t, report.Indicators, 5)
	require.Len(t, report.AlertEvents, 5)
	require.Contains(t, captureIndicatorNames(report.Indicators), CaptureRiskSuddenDelegationInflow)
	require.Contains(t, captureIndicatorNames(report.Indicators), CaptureRiskRapidCommissionChange)
	require.Contains(t, captureIndicatorNames(report.Indicators), CaptureRiskRecentSlash)
	require.Contains(t, captureIndicatorNames(report.Indicators), CaptureRiskSelfDelegationWithdrawal)
	require.Contains(t, captureIndicatorNames(report.Indicators), CaptureRiskOperatorMetadataInconsistency)
	require.True(t, report.AlertEvents[0].AdvisoryOnly)
}

func TestRiskAdjustedYieldProjectionRewardBandsAndVarianceAreQueryable(t *testing.T) {
	params := testParams()
	candidate := marketCandidate("val-yield", 1_000, 1_000, 1_000)
	delegation := marketDelegation("del-a", "val-yield", 1_000, "")
	score := testRecord(6, "val-yield", 2_000)
	score.PerformanceFactor = 9_000
	score.UptimeFactor = 9_000
	score.ReliabilityIndex = 9_000
	slash := ValidatorSlashHistoryRecord{
		EpochID:		4,
		Height:			40,
		Validator:		"val-yield",
		Misbehavior:		postypes.MisbehaviorDowntime,
		SlashFractionBps:	100,
		SelfBondSlashedNaet:	sdkmath.NewInt(10),
		DelegatorSlashedNaet:	sdkmath.NewInt(10),
		TotalSlashedNaet:	sdkmath.NewInt(20),
	}
	state, err := NewValidatorMarketState(params, []postypes.Candidate{candidate}, []DelegationRecord{delegation}, []ValidatorScoreRecord{score}, []ValidatorSlashHistoryRecord{slash}, nil)
	require.NoError(t, err)
	decParams := DefaultDecentralizationParams(params)
	decParams.MaxValidatorShareBps = postypes.BasisPoints

	projection, found, err := state.QueryRiskAdjustedYieldProjection(RiskAdjustedYieldInput{
		Delegator:			"del-a",
		Validator:			"val-yield",
		AmountNaet:			sdkmath.NewInt(1_000),
		AnnualRewardsNaet:		sdkmath.NewInt(200),
		UnbondingLiquidityCostBps:	200,
		Decentralization:		decParams,
		ActiveValidatorIDs:		[]string{"val-yield"},
	})
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, projection.ReproducibleFromQueries)
	require.Equal(t, uint32(1_000), projection.GrossRewardRateBps)
	require.Equal(t, uint32(1_000), projection.CommissionBps)
	require.Equal(t, uint32(9_000), projection.HistoricalUptimeBps)
	require.Equal(t, uint32(1_500), projection.SlashProbabilityProxyBps)
	require.Equal(t, uint32(606), projection.RiskAdjustedYieldBps)
	require.Equal(t, sdkmath.NewInt(60), projection.ExpectedRewardNaet)
	require.Equal(t, sdkmath.NewInt(49), projection.LowRewardNaet)
	require.Equal(t, sdkmath.NewInt(70), projection.HighRewardNaet)
	require.Equal(t, uint32(1_700), projection.UncertaintyBps)

	variance, found, err := ComputeValidatorRewardVariance("val-yield", []ValidatorRewardObservation{
		{EpochID: 1, Validator: "val-yield", RewardPerStakeBps: 800},
		{EpochID: 2, Validator: "val-yield", RewardPerStakeBps: 1_000},
		{EpochID: 3, Validator: "val-yield", RewardPerStakeBps: 1_200},
	})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint32(3), variance.ObservationCount)
	require.Equal(t, uint32(1_000), variance.MeanRewardBps)
	require.Equal(t, uint32(133), variance.VarianceBps)
}

func TestDelegationSimulatorHandlesCommissionSlashAndDampeningInputs(t *testing.T) {
	params := testParams()
	params.MaxVotingPowerBps = postypes.BasisPoints
	candidates := []postypes.Candidate{
		marketCandidate("val-from", 2_000, 1_000, 500),
		marketCandidate("val-to", 1_000, 7_000, 500),
	}
	delegations := []DelegationRecord{
		marketDelegation("del-a", "val-from", 1_000, ""),
		marketDelegation("del-target", "val-to", 7_000, ""),
	}
	scores := []ValidatorScoreRecord{
		testRecord(8, "val-from", 3_000),
		testRecord(8, "val-to", 8_000),
	}
	state, err := NewValidatorMarketState(params, candidates, delegations, scores, nil, nil)
	require.NoError(t, err)
	decParams := DefaultDecentralizationParams(params)
	decParams.MaxValidatorShareBps = 4_000
	decParams.MaxTopNShareBps = postypes.BasisPoints

	result, found, err := state.SimulateDelegation(DelegationSimulationInput{
		Delegator:		"del-a",
		FromValidator:		"val-from",
		ToValidator:		"val-to",
		AmountNaet:		sdkmath.NewInt(1_000),
		AnnualRewardsNaet:	sdkmath.NewInt(500),
		CurrentEpoch:		9,
		Height:			90,
		CommissionOverrides: []ValidatorCommissionRecord{
			{EpochID: 9, Height: 90, Validator: "val-to", CommissionBps: 1_500},
		},
		SlashEvents: []ValidatorSlashHistoryRecord{
			{EpochID: 9, Height: 90, Validator: "val-to", Misbehavior: postypes.MisbehaviorDowntime, SlashFractionBps: 100, SelfBondSlashedNaet: sdkmath.NewInt(10), DelegatorSlashedNaet: sdkmath.NewInt(10), TotalSlashedNaet: sdkmath.NewInt(20)},
		},
		Decentralization:		decParams,
		ActiveValidatorIDs:		[]string{"val-from", "val-to"},
		UnbondingLiquidityCostBps:	100,
	})
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, result.AdvisoryOnly)
	require.Equal(t, uint32(1_500), result.TargetProjection.CommissionBps)
	require.True(t, result.TargetProjection.SlashProbabilityProxyBps > 0)
	require.True(t, result.TargetProjection.ConcentrationAdjustmentBps > 0)
	require.True(t, result.RedelegationCost.EstimatedCostNaet.IsPositive())
	require.Equal(t, result.TargetProjection.ExpectedRewardNaet.Sub(result.CurrentProjection.ExpectedRewardNaet), result.ProjectedRewardDeltaNaet)
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
		EpochID:		3,
		Height:			33,
		Validator:		"val-risk",
		Misbehavior:		postypes.MisbehaviorDoubleSign,
		SlashFractionBps:	1_000,
		SelfBondSlashedNaet:	sdkmath.NewInt(200),
		DelegatorSlashedNaet:	sdkmath.NewInt(800),
		TotalSlashedNaet:	sdkmath.NewInt(1_000),
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
		EpochID:		8,
		Height:			80,
		Validator:		"val-old",
		Misbehavior:		postypes.MisbehaviorDowntime,
		SlashFractionBps:	500,
		SelfBondSlashedNaet:	sdkmath.NewInt(50),
		DelegatorSlashedNaet:	sdkmath.NewInt(50),
		TotalSlashedNaet:	sdkmath.NewInt(100),
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
		Delegator:		delegator,
		Validator:		validator,
		Amount:			sdkmath.NewInt(amount),
		ActivationEpoch:	11,
		RiskAppetite:		RiskAppetiteBalanced,
		CommissionTolerance:	1_000,
		LockDurationPreference:	LockDurationEpoch,
		RewardStrategy:		RewardStrategyLiquid,
		RiskTrancheOptional:	tranche,
		CreatedHeight:		10,
		UpdatedHeight:		10,
	}
}

func marketCandidate(id string, selfStake int64, delegatedStake int64, commissionBps uint32) postypes.Candidate {
	nominations := []postypes.Nomination(nil)
	if delegatedStake > 0 {
		nominations = []postypes.Nomination{{NominatorID: "market-delegators", StakeNaet: sdkmath.NewInt(delegatedStake)}}
	}
	return postypes.Candidate{
		ValidatorID:		id,
		SelfStakeNaet:		sdkmath.NewInt(selfStake),
		DelegatedStakeNaet:	sdkmath.NewInt(delegatedStake),
		PerformanceScoreBps:	postypes.BasisPoints,
		UptimeFactorBps:	postypes.BasisPoints,
		CommissionBps:		commissionBps,
		Nominations:		nominations,
	}
}

func policyEvaluationMap(evaluations []DelegationPolicyEvaluation) map[string]DelegationPolicyEvaluation {
	out := make(map[string]DelegationPolicyEvaluation, len(evaluations))
	for _, evaluation := range evaluations {
		out[evaluation.PolicyName] = evaluation
	}
	return out
}

func captureIndicatorNames(indicators []CaptureRiskIndicator) []string {
	out := make([]string, len(indicators))
	for i, indicator := range indicators {
		out[i] = indicator.Name
	}
	return out
}
