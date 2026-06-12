package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	policykeeper "github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	"github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
)

func TestPolicyCanConsumeSDKStakingKeeperValidators(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)

	records := make([]types.ValidatorStake, 0, len(validators))
	for _, validator := range validators {
		tokens := validator.GetTokens()
		require.True(t, tokens.IsPositive())
		records = append(records, types.ValidatorStake{
			OperatorAddress:	validator.OperatorAddress,
			RawStake:		tokens.Uint64(),
			CommissionBps:		types.DefaultParams(authority).CommissionFloorBps,
		})
	}

	k := policykeeper.NewKeeper(authority)
	network, err := k.RecomputePolicy(1, records)
	require.NoError(t, err)
	require.Equal(t, uint32(len(validators)), network.ActiveValidators)
	require.Equal(t, records[0].RawStake, network.TotalRawStake)
	require.NoError(t, network.Validate(k.Params()))
}
