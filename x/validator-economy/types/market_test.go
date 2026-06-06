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
	require.Equal(t, commissions, state.QueryValidatorCommissionHistory("val-risk"))
	require.Equal(t, []ValidatorSlashHistoryRecord{slash}, state.QueryValidatorSlashHistory("val-risk"))
	require.Equal(t, []ValidatorScoreRecord{score}, state.QueryValidatorPerformanceHistory("val-risk"))
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
