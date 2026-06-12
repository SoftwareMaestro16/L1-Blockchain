package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestBuildDelegationRecordCapturesCapitalLayerAttributes(t *testing.T) {
	params := testParams()
	record, err := BuildDelegationRecord(params, 20, 100, "del-a", "val-a", sdkmath.NewInt(1_000), DelegationPreferences{
		RiskAppetite:		RiskAppetiteAggressive,
		CommissionTolerance:	700,
		LockDurationPreference:	LockDurationLongTerm,
		RewardStrategy:		RewardStrategyAutoRedelegate,
		RiskTrancheOptional:	RiskTrancheFirstLoss,
	})
	require.NoError(t, err)
	require.Equal(t, DelegationRecord{
		Delegator:		"del-a",
		Validator:		"val-a",
		Amount:			sdkmath.NewInt(1_000),
		ActivationEpoch:	21,
		Status:			DelegationStatusActive,
		RiskAppetite:		RiskAppetiteAggressive,
		CommissionTolerance:	700,
		LockDurationPreference:	LockDurationLongTerm,
		RewardStrategy:		RewardStrategyAutoRedelegate,
		RiskTrancheOptional:	RiskTrancheFirstLoss,
		CreatedHeight:		100,
		UpdatedHeight:		100,
	}, record)
}

func TestBuildDelegationRecordAppliesSafeDefaults(t *testing.T) {
	params := testParams()
	record, err := BuildDelegationRecord(params, 9, 50, "del-default", "val-default", sdkmath.NewInt(250), DelegationPreferences{
		CommissionTolerance: 500,
	})
	require.NoError(t, err)
	require.Equal(t, RiskAppetiteBalanced, record.RiskAppetite)
	require.Equal(t, DelegationStatusActive, record.Status)
	require.Equal(t, LockDurationEpoch, record.LockDurationPreference)
	require.Equal(t, RewardStrategyLiquid, record.RewardStrategy)
	require.Empty(t, record.RiskTrancheOptional)
}

func TestBuildDelegationRecordFromIntentUsesActivationDelayAndTolerance(t *testing.T) {
	params := testParams()
	intent := postypes.DelegationIntent{
		NominatorID:		"del-intent",
		ValidatorID:		"val-intent",
		StakeNaet:		sdkmath.NewInt(900),
		RequestedEpoch:		30,
		MaxCommissionBps:	600,
		MinPerformanceScoreBps:	9_000,
	}

	record, err := BuildDelegationRecordFromIntent(params, intent, 222, DelegationPreferences{
		RiskAppetite:		RiskAppetiteConservative,
		LockDurationPreference:	LockDurationFlexible,
		RewardStrategy:		RewardStrategyCompound,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(31), record.ActivationEpoch)
	require.Equal(t, uint32(600), record.CommissionTolerance)
	require.Equal(t, sdkmath.NewInt(900), record.Amount)

	_, err = BuildDelegationRecordFromIntent(params, intent, 222, DelegationPreferences{
		RiskAppetite:		RiskAppetiteConservative,
		CommissionTolerance:	500,
		LockDurationPreference:	LockDurationFlexible,
		RewardStrategy:		RewardStrategyCompound,
	})
	require.ErrorContains(t, err, "below intent tolerance")
}

func TestDelegationRecordValidationRejectsUnsafePreferences(t *testing.T) {
	params := testParams()
	valid, err := BuildDelegationRecord(params, 1, 1, "del-a", "val-a", sdkmath.NewInt(1), DelegationPreferences{
		CommissionTolerance: 1,
	})
	require.NoError(t, err)

	invalid := valid
	invalid.Amount = sdkmath.ZeroInt()
	require.ErrorContains(t, invalid.Validate(params), "amount")

	invalid = valid
	invalid.Status = "moved"
	require.ErrorContains(t, invalid.Validate(params), "delegation status")

	invalid = valid
	invalid.RiskAppetite = "reckless"
	require.ErrorContains(t, invalid.Validate(params), "risk appetite")

	invalid = valid
	invalid.CommissionTolerance = params.MaxCommissionBps + 1
	require.ErrorContains(t, invalid.Validate(params), "commission tolerance")

	invalid = valid
	invalid.LockDurationPreference = "forever"
	require.ErrorContains(t, invalid.Validate(params), "lock duration")

	invalid = valid
	invalid.RewardStrategy = "unknown"
	require.ErrorContains(t, invalid.Validate(params), "reward strategy")

	invalid = valid
	invalid.RiskTrancheOptional = "unsafe"
	require.ErrorContains(t, invalid.Validate(params), "risk tranche")

	invalid = valid
	invalid.UpdatedHeight = invalid.CreatedHeight - 1
	require.ErrorContains(t, invalid.Validate(params), "updated height")
}

func TestDelegationCapitalStateQueriesByDelegatorAndValidator(t *testing.T) {
	params := testParams()
	a, err := BuildDelegationRecord(params, 4, 10, "del-a", "val-b", sdkmath.NewInt(100), DelegationPreferences{CommissionTolerance: 500})
	require.NoError(t, err)
	b, err := BuildDelegationRecord(params, 3, 11, "del-a", "val-a", sdkmath.NewInt(200), DelegationPreferences{CommissionTolerance: 500})
	require.NoError(t, err)
	c, err := BuildDelegationRecord(params, 3, 12, "del-b", "val-a", sdkmath.NewInt(300), DelegationPreferences{CommissionTolerance: 500})
	require.NoError(t, err)

	state, err := NewDelegationCapitalState(params, []DelegationRecord{a, b, c})
	require.NoError(t, err)
	require.Equal(t, []DelegationRecord{b, a}, state.RecordsForDelegator("del-a"))
	require.Equal(t, []DelegationRecord{b, c}, state.RecordsForValidator("val-a"))
	require.Equal(t, sdkmath.NewInt(500), state.TotalDelegatedToValidator("val-a"))

	_, err = NewDelegationCapitalState(params, []DelegationRecord{b, b})
	require.ErrorContains(t, err, "duplicate delegation record")
}

func TestCommissionToleranceMarksExceededAndEmitsAdvisoryAlert(t *testing.T) {
	params := testParams()
	record, err := BuildDelegationRecord(params, 5, 100, "del-a", "val-a", sdkmath.NewInt(500), DelegationPreferences{
		CommissionTolerance: 600,
	})
	require.NoError(t, err)

	updated, alert, err := CheckCommissionTolerance(params, record, 700, 120, true)
	require.NoError(t, err)
	require.Equal(t, DelegationStatusCommissionExceeded, updated.Status)
	require.Equal(t, uint64(120), updated.UpdatedHeight)
	require.NotNil(t, alert)
	require.Equal(t, DelegationStatusActive, alert.PreviousStatus)
	require.Equal(t, DelegationStatusCommissionExceeded, alert.NewStatus)
	require.Equal(t, uint32(600), alert.CommissionToleranceBps)
	require.Equal(t, uint32(700), alert.CurrentCommissionBps)
	require.True(t, alert.RedelegationAdvisory)

	recovered, alert, err := CheckCommissionTolerance(params, updated, 500, 130, true)
	require.NoError(t, err)
	require.Nil(t, alert)
	require.Equal(t, DelegationStatusActive, recovered.Status)
}

func TestLockDurationPreferencePreservesUnbondingAndRequiresExtendedSlashWindow(t *testing.T) {
	params := testParams()
	record, err := BuildDelegationRecord(params, 5, 100, "del-long", "val-a", sdkmath.NewInt(500), DelegationPreferences{
		CommissionTolerance:	500,
		LockDurationPreference:	LockDurationLongTerm,
	})
	require.NoError(t, err)

	withoutWindow, err := EvaluateLockDurationPreference(params, record, params.EvidenceWindowEpochs)
	require.NoError(t, err)
	require.Equal(t, params.UnbondingSeconds*2, withoutWindow.EffectiveUnbondingSeconds)
	require.False(t, withoutWindow.EligibleForRewardMultiplier)
	require.Equal(t, uint32(postypes.BasisPoints), withoutWindow.RewardMultiplierBps)
	require.True(t, withoutWindow.RedelegationKeepsRiskHistory)

	withWindow, err := EvaluateLockDurationPreference(params, record, params.EvidenceWindowEpochs*2)
	require.NoError(t, err)
	require.True(t, withWindow.EligibleForRewardMultiplier)
	require.Equal(t, uint32(11_000), withWindow.RewardMultiplierBps)
	require.GreaterOrEqual(t, withWindow.EffectiveUnbondingSeconds, postypes.MinUnbondingSeconds)
}
