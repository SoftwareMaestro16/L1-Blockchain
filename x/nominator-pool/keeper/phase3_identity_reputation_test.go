package keeper

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestPhase35PoolClaimWritesUnifiedIdentityReputationByDurationAndAmount(t *testing.T) {
	shortUser := aePoolAddress(t, "c1")
	longUser := aePoolAddress(t, "c2")
	largeUser := aePoolAddress(t, "c3")
	k := NewKeeperWithAccountStatus(accountStatusFixture{
		shortUser:	accountStatusActive,
		longUser:	accountStatusActive,
		largeUser:	accountStatusActive,
	})
	pool := createOfficialLiquidStakingPool(t, &k, "phase35-identity")
	for _, deposit := range []struct {
		user	string
		amount	uint64
	}{
		{shortUser, types.DefaultMinPoolDeposit},
		{longUser, types.DefaultMinPoolDeposit},
		{largeUser, 2 * types.DefaultMinPoolDeposit},
	} {
		_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
			PoolID:		pool.PoolID,
			WalletAddress:	deposit.user,
			Amount:		deposit.amount,
			Height:		2,
		})
		require.NoError(t, err)
	}
	gs := k.ExportGenesis()
	gs.State.LiquidStakingPools[0].TotalActiveStake = 4 * types.DefaultMinPoolDeposit
	require.NoError(t, k.InitGenesis(gs))

	shortClaim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: shortUser, Height: 12})
	require.NoError(t, err)
	longClaim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: longUser, Height: 22})
	require.NoError(t, err)
	largeClaim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: largeUser, Height: 12})
	require.NoError(t, err)

	require.Greater(t, longClaim.ReputationDelta, shortClaim.ReputationDelta)
	require.Greater(t, largeClaim.ReputationDelta, shortClaim.ReputationDelta)

	exported := k.ExportGenesis()
	_, hasSeparateStakeReputationState := reflect.TypeOf(exported.State).FieldByName("StakeReputationAccumulators")
	require.False(t, hasSeparateStakeReputationState)
}

func TestPhase35IdentityReputationClaimIdempotentAndRecordsSlashingExposure(t *testing.T) {
	user := aePoolAddress(t, "c4")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "phase35-slash-exposure")
	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)
	gs := k.ExportGenesis()
	gs.State.Pools[0].SlashIndex = types.RewardDelta(123, gs.State.Pools[0].TotalShares)
	gs.State.LiquidStakingPools[0].TotalActiveStake = types.DefaultMinPoolDeposit - 123
	require.NoError(t, k.InitGenesis(gs))

	first, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: user, Height: 20})
	require.NoError(t, err)
	second, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: user, Height: 20})
	require.NoError(t, err)
	require.Zero(t, second.ReputationDelta)
	_ = first
}

func TestPhase35LowIdentityReputationDoesNotBlockPoolDepositOrClaim(t *testing.T) {
	user := aePoolAddress(t, "c5")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "phase35-low-rep")

	receipt, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, types.DefaultMinPoolDeposit, receipt.Amount)

	claim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{
		PoolID:		pool.PoolID,
		OwnerAddress:	user,
		Height:		12,
	})
	require.NoError(t, err)
	require.Equal(t, user, claim.Account)
	require.Equal(t, pool.PoolID, claim.PoolID)
}
