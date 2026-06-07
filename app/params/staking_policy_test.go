package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultStakingDelegationPolicyMatchesAetraModel(t *testing.T) {
	policy := DefaultStakingDelegationPolicy()

	require.NoError(t, policy.Validate())
	require.Equal(t, BaseDenom, policy.Denom)
	require.Equal(t, int64(10_000*BaseUnitsPerDisplay), policy.MinSelfBondNaet)
	require.Equal(t, int64(50_000*BaseUnitsPerDisplay), policy.MinValidatorBondNaet)
	require.Greater(t, policy.MinValidatorBondNaet, policy.MinSelfBondNaet)
	require.Equal(t, StakingCommissionFloorBps, policy.MinCommissionBps)
	require.Equal(t, StakingCommissionCeilingBps, policy.MaxCommissionBps)
	require.Equal(t, StakingMaxDailyCommissionBps, policy.MaxDailyCommissionBps)
	require.True(t, policy.DelegationEnabled)
	require.True(t, policy.RedelegationEnabled)
	require.True(t, policy.NominationPoolsEnabled)
	require.True(t, policy.SlashingInherited)
	require.True(t, policy.RequireValidatorMetadata)
}

func TestStakingPolicyUnbondingWindowIsFourteenToTwentyOneDays(t *testing.T) {
	policy := DefaultStakingDelegationPolicy()

	require.Equal(t, uint64(201_600), policy.UnbondingMinBlocks)
	require.Equal(t, uint64(302_400), policy.UnbondingMaxBlocks)
	require.Equal(t, uint64(201_600), policy.UnbondingDefaultBlocks)
	require.NoError(t, ValidateStakingUnbondingBlocks(policy.UnbondingMinBlocks))
	require.NoError(t, ValidateStakingUnbondingBlocks(policy.UnbondingMaxBlocks))
	require.ErrorContains(t, ValidateStakingUnbondingBlocks(policy.UnbondingMinBlocks-1), "14-21 days")
	require.ErrorContains(t, ValidateStakingUnbondingBlocks(policy.UnbondingMaxBlocks+1), "14-21 days")
}

func TestStakingPolicyRequiresValidatorBondAboveSelfBond(t *testing.T) {
	policy := DefaultStakingDelegationPolicy()

	require.NoError(t, policy.ValidateValidatorBond(policy.MinSelfBondNaet, policy.MinValidatorBondNaet))
	require.ErrorContains(t, policy.ValidateValidatorBond(policy.MinSelfBondNaet-1, policy.MinValidatorBondNaet), "self-bond")
	require.ErrorContains(t, policy.ValidateValidatorBond(policy.MinSelfBondNaet, policy.MinValidatorBondNaet-1), "validator bond")
	require.ErrorContains(t, policy.ValidateValidatorBond(policy.MinValidatorBondNaet+1, policy.MinValidatorBondNaet), "lower than self-bond")
}

func TestStakingPolicyRejectsUnsafeConfiguration(t *testing.T) {
	policy := DefaultStakingDelegationPolicy()
	policy.Denom = "uatom"
	require.ErrorContains(t, policy.Validate(), BaseDenom)

	policy = DefaultStakingDelegationPolicy()
	policy.MinValidatorBondNaet = policy.MinSelfBondNaet
	require.ErrorContains(t, policy.Validate(), "higher than minimum self-bond")

	policy = DefaultStakingDelegationPolicy()
	policy.UnbondingDefaultBlocks = 1_000
	require.ErrorContains(t, policy.Validate(), "14-21 days")

	policy = DefaultStakingDelegationPolicy()
	policy.SlashingInherited = false
	require.ErrorContains(t, policy.Validate(), "slashing risk")

	policy = DefaultStakingDelegationPolicy()
	policy.RequireValidatorMetadata = false
	require.ErrorContains(t, policy.Validate(), "metadata")
}

func TestDefaultAntiCartelPolicyMatchesAetraControls(t *testing.T) {
	policy := DefaultAntiCartelPolicy()

	require.NoError(t, policy.Validate())
	require.Equal(t, int64(300), policy.CommissionFloorBps)
	require.Equal(t, int64(2_000), policy.CommissionCeilingBps)
	require.Equal(t, int64(100), policy.MaxDailyCommissionChangeBps)
	require.True(t, policy.ValidatorIdentityRegistry)
	require.False(t, policy.MandatoryKYC)
	require.True(t, policy.ValidatorMetadataTransparency)
	require.True(t, policy.PublicConcentrationMetrics)
	require.True(t, policy.SelfBondRatioVisibility)
	require.True(t, policy.ObjectiveCorrelationWarnings)
	require.True(t, policy.EconomicSignalsInsteadOfHalting)
}

func TestAntiCartelPolicyRejectsCentralizedAdmissionAndOpaqueMetrics(t *testing.T) {
	policy := DefaultAntiCartelPolicy()
	policy.MandatoryKYC = true
	require.ErrorContains(t, policy.Validate(), "mandatory KYC")

	policy = DefaultAntiCartelPolicy()
	policy.PublicConcentrationMetrics = false
	require.ErrorContains(t, policy.Validate(), "transparency metrics")

	policy = DefaultAntiCartelPolicy()
	policy.ObjectiveCorrelationWarnings = false
	require.ErrorContains(t, policy.Validate(), "objective evidence")

	policy = DefaultAntiCartelPolicy()
	policy.EconomicSignalsInsteadOfHalting = false
	require.ErrorContains(t, policy.Validate(), "economic signals")
}
