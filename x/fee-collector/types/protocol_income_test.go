package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestProtocolIncomePolicyValidation(t *testing.T) {
	policy := DefaultProtocolIncomePolicy()
	require.NoError(t, policy.Validate())

	policy.Buckets[0].Weight++
	require.ErrorContains(t, policy.Validate(), "sum")
}

func TestZeroWeightRequiresExplicitAllowance(t *testing.T) {
	policy := DefaultProtocolIncomePolicy()
	policy.Buckets[0].Weight = 0
	policy.Buckets[1].Weight += 3_800

	require.ErrorContains(t, policy.Validate(), "zero weight")
}

func TestConstitutionalMinimumEnforced(t *testing.T) {
	policy := DefaultProtocolIncomePolicy()
	for i := range policy.Buckets {
		if policy.Buckets[i].Bucket == BucketDelegatorProtection {
			policy.Buckets[i].Weight = policy.Buckets[i].ConstitutionMin - 1
		}
		if policy.Buckets[i].Bucket == BucketTreasury {
			policy.Buckets[i].Weight++
		}
	}

	require.ErrorContains(t, policy.Validate(), "constitutional minimum")
}

func TestSplitProtocolIncomeDeterministicRounding(t *testing.T) {
	policy := ProtocolIncomePolicy{
		Scale:	BasisPoints,
		Buckets: []ProtocolIncomeBucketRule{
			{Bucket: BucketValidatorRewards, ModuleAccount: "validators", Weight: 3_333},
			{Bucket: BucketTreasury, ModuleAccount: "treasury", Weight: 3_333},
			{Bucket: BucketBurn, ModuleAccount: "burn", Weight: 3_334, Burn: true},
		},
	}

	allocations, remainder, err := SplitProtocolIncome(policy, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 10)))
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 1)), remainder)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 3)), allocations[0].Amount)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 3)), allocations[1].Amount)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(BaseDenom, 4)), allocations[2].Amount)
}
