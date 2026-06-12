package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestPhase33PersistentExportImportAfterRewardsAndClaimIdempotency(t *testing.T) {
	ctx := context.Background()
	user := aePoolAddress(t, "91")
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	source.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, source.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &source, "phase33-reward-persist")
	_, err := source.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)
	summary, err := source.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:		source.ExportGenesis().Params.Authority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		Height:			3,
		RewardRateBps:		1_000,
		EmissionsAllocated:	types.DefaultMinPoolDeposit,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:		rawPoolAddress("92"),
			PoolAllocatedStake:	types.DefaultMinPoolDeposit,
			PerformanceBps:		types.MaxBasisPoints,
		}},
	})
	require.NoError(t, err)
	require.NotZero(t, summary.RewardIndexAfter)

	firstClaim, err := source.ClaimPoolRewardsWithReceipt(types.MsgClaimPoolRewards{PoolID: pool.PoolID, OwnerAddress: user, Height: 4})
	require.NoError(t, err)
	require.NotZero(t, firstClaim.Amount)
	secondClaim, err := source.ClaimPoolRewardsWithReceipt(types.MsgClaimPoolRewards{PoolID: pool.PoolID, OwnerAddress: user, Height: 5})
	require.NoError(t, err)
	require.Zero(t, secondClaim.Amount)

	exported, err := source.ExportGenesisState(ctx)
	require.NoError(t, err)
	imported := NewPersistentKeeper(kvtest.NewStoreService())
	imported.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	roundTrip, err := imported.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, exported.State.Pools[0].RewardIndex, roundTrip.State.Pools[0].RewardIndex)
	require.Equal(t, exported.State.PoolShares[0].PendingRewards, roundTrip.State.PoolShares[0].PendingRewards)
	require.Equal(t, exported.State.RewardClaims, roundTrip.State.RewardClaims)
}

func TestPhase33ReputationCannotIncreaseWithoutStakeTimeExposure(t *testing.T) {
	user := aePoolAddress(t, "93")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "phase33-reputation")
	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	claim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{
		PoolID:		pool.PoolID,
		OwnerAddress:	user,
		Height:		2,
	})
	require.NoError(t, err)
	require.Zero(t, claim.ReputationDelta)
	require.Zero(t, claim.ReputationScore)
}
