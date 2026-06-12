package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestPhase32PersistentPoolMutationExportImportQuerySameState(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	user := aePoolAddress(t, "71")
	source := NewPersistentKeeper(service)
	source.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, source.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &source, "phase32-persist")

	receipt, err := source.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	legacy, err := service.RawStore().Get(genesisKey)
	require.NoError(t, err)
	require.Empty(t, legacy)
	poolRecord, err := service.RawStore().Get(types.PoolKey(pool.PoolID))
	require.NoError(t, err)
	require.NotEmpty(t, poolRecord)
	shareRecord, err := service.RawStore().Get(types.PoolShareKey(pool.PoolID, user))
	require.NoError(t, err)
	require.NotEmpty(t, shareRecord)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, receipt.Shares, exported.State.PoolShares[0].Shares)
	require.Equal(t, receipt.Amount, exported.State.LiquidStakingPools[0].TotalDeposited)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	imported.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	roundTrip, err := imported.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)

	query, found := imported.PoolShare(types.QueryPoolShareRequest{PoolID: pool.PoolID, Delegator: rawPoolAddress("71")})
	require.True(t, found)
	require.Equal(t, receipt.Shares, query.Share.Shares)
}

func TestPhase32ExportOrderDeterministicAndPaginationBounded(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	k := NewPersistentKeeper(service)
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	for _, id := range []string{"pool-c", "pool-a", "pool-b"} {
		_, err := k.CreateNominatorPool(types.MsgCreateNominatorPool{
			Authority:		prototype.DefaultAuthority,
			PoolID:			id,
			PoolOperator:		rawPoolAddress("11"),
			ValidatorTarget:	rawPoolAddress("12"),
			PoolCommissionBps:	100,
			Height:			1,
			ValidatorStatus:	"active",
		})
		require.NoError(t, err)
	}

	exported, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{"pool-a", "pool-b", "pool-c"}, []string{
		exported.State.Pools[0].PoolID,
		exported.State.Pools[1].PoolID,
		exported.State.Pools[2].PoolID,
	})

	query := NewQueryServerImpl(&k)
	page, err := query.NominatorPools(ctx, &types.QueryNominatorPoolsRequest{Limit: 2})
	require.NoError(t, err)
	require.Len(t, page.Pools, 2)
	require.Equal(t, uint64(2), page.NextOffset)
	require.Equal(t, uint64(3), page.Total)
	require.Equal(t, "pool-a", page.Pools[0].PoolID)
	require.Equal(t, "pool-b", page.Pools[1].PoolID)

	next, err := query.NominatorPools(ctx, &types.QueryNominatorPoolsRequest{Offset: page.NextOffset, Limit: 2})
	require.NoError(t, err)
	require.Len(t, next.Pools, 1)
	require.Zero(t, next.NextOffset)
	require.Equal(t, "pool-c", next.Pools[0].PoolID)
}

func TestPhase32StorageRentDebtPreservedInKVExportImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	k := NewPersistentKeeper(service)
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &k, "phase32-rent")

	gs := k.ExportGenesis()
	gs.State.LiquidStakingPools[0].StorageRentDebt = 123
	gs.State.LiquidStakingPools[0].StorageRentReserve = 456
	require.NoError(t, k.InitGenesisState(ctx, gs))

	exported, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, pool.PoolID, exported.State.LiquidStakingPools[0].PoolID)
	require.Equal(t, uint64(123), exported.State.LiquidStakingPools[0].StorageRentDebt)
	require.Equal(t, uint64(456), exported.State.LiquidStakingPools[0].StorageRentReserve)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	roundTrip, err := imported.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, exported.State.LiquidStakingPools[0], roundTrip.State.LiquidStakingPools[0])
}

func TestPhase32OfficialPoolRentAccrualHookChargesReserveOnMutation(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	user := aePoolAddress(t, "7a")
	k := NewPersistentKeeper(service)
	k.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &k, "phase32-rent-hook")

	gs := k.ExportGenesis()
	gs.State.LiquidStakingPools[0].LastStorageChargeHeight = 2
	gs.State.LiquidStakingPools[0].StorageRentReserve = 1_000_000
	gs.State.LiquidStakingPools[0].StorageRentDebt = 0
	require.NoError(t, k.InitGenesisState(ctx, gs))

	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		10,
	})
	require.NoError(t, err)

	exported, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	liquid := exported.State.LiquidStakingPools[0]
	require.Equal(t, uint64(10), liquid.LastStorageChargeHeight)
	require.Less(t, liquid.StorageRentReserve, uint64(1_000_000))
	require.Zero(t, liquid.StorageRentDebt)
}

func TestPhase32OfficialPoolRentAccruesOnClaimAndEpochRebalance(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	user := aePoolAddress(t, "7b")
	validator := aePoolAddress(t, "2a")
	k := NewPersistentKeeper(service)
	k.accountStatusReader = accountStatusFixture{user: accountStatusActive, validator: accountStatusActive}
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &k, "phase32-rent-claim-epoch")

	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)
	_, err = k.ApplyPoolReward(pool.PoolID, 1_000)
	require.NoError(t, err)
	_, err = k.RegisterValidator(types.MsgRegisterValidator{
		SignerAddress:		validator,
		ValidatorAddress:	validator,
		SelfStake:		types.DefaultMinValidatorStake,
		CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
		Height:			3,
	})
	require.NoError(t, err)

	gs := k.ExportGenesis()
	gs.State.LiquidStakingPools[0].LastStorageChargeHeight = 2
	gs.State.LiquidStakingPools[0].StorageRentReserve = 5_000_000
	gs.State.LiquidStakingPools[0].StorageRentDebt = 0
	require.NoError(t, k.InitGenesisState(ctx, gs))

	_, err = k.ClaimPoolRewardsWithReceipt(types.MsgClaimPoolRewards{
		PoolID:		pool.PoolID,
		OwnerAddress:	user,
		Height:		10,
	})
	require.NoError(t, err)

	afterClaim, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	liquidAfterClaim := afterClaim.State.LiquidStakingPools[0]
	require.Equal(t, uint64(10), liquidAfterClaim.LastStorageChargeHeight)
	require.Less(t, liquidAfterClaim.StorageRentReserve, uint64(5_000_000))

	_, err = k.RebalancePoolAllocations(types.MsgRebalancePoolAllocations{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		Epoch:			1,
		Height:			14,
		Candidates: []types.ValidatorPolicyCandidate{{
			ValidatorAddress:	validator,
			ReputationScore:	8_000,
			UptimeBps:		9_500,
			CommissionBps:		1_000,
			StakeEfficiencyBps:	8_000,
			SlashingRiskBps:	100,
			NetworkLoadBps:		1_000,
		}},
	})
	require.NoError(t, err)

	afterSync, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	liquidAfterSync := afterSync.State.LiquidStakingPools[0]
	require.Equal(t, uint64(14), liquidAfterSync.LastStorageChargeHeight)
	require.Less(t, liquidAfterSync.StorageRentReserve, liquidAfterClaim.StorageRentReserve)
}

func TestPhase32OfficialPoolRentDebtRecoveryAndExportImportRoundTrip(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	user := aePoolAddress(t, "7c")
	k := NewPersistentKeeper(service)
	k.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &k, "phase32-rent-recovery")

	gs := k.ExportGenesis()
	gs.State.LiquidStakingPools[0].LastStorageChargeHeight = 2
	gs.State.LiquidStakingPools[0].StorageRentReserve = 0
	gs.State.LiquidStakingPools[0].StorageRentDebt = 0
	require.NoError(t, k.InitGenesisState(ctx, gs))

	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		10,
	})
	require.NoError(t, err)

	withDebt, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	liquidWithDebt := withDebt.State.LiquidStakingPools[0]
	require.Greater(t, liquidWithDebt.StorageRentDebt, uint64(0))
	require.Equal(t, types.PoolStatusFrozenLimited, liquidWithDebt.Status)

	_, err = k.TopUpPoolReserve(types.MsgTopUpPoolReserve{
		PoolID:		pool.PoolID,
		PayerAddress:	user,
		Amount:		liquidWithDebt.StorageRentDebt,
		Height:		10,
	})
	require.NoError(t, err)

	recovered, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	liquidRecovered := recovered.State.LiquidStakingPools[0]
	require.Zero(t, liquidRecovered.StorageRentDebt)
	require.Equal(t, types.PoolStatusActive, liquidRecovered.Status)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	imported.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, imported.InitGenesisState(ctx, recovered))
	roundTrip, err := imported.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, liquidRecovered, roundTrip.State.LiquidStakingPools[0])
}

func TestPhase32PoolAllocationKVUpdateTouchesOnlyChangedAllocationKey(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	v1 := aePoolAddress(t, "81")
	v2 := aePoolAddress(t, "82")
	k := NewPersistentKeeper(service)
	k.accountStatusReader = accountStatusFixture{v1: accountStatusActive, v2: accountStatusActive}
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &k, "phase32-alloc")
	for _, validator := range []string{v1, v2} {
		_, err := k.RegisterValidator(types.MsgRegisterValidator{
			SignerAddress:		validator,
			ValidatorAddress:	validator,
			SelfStake:		types.DefaultMinValidatorStake,
			CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
			Height:			1,
		})
		require.NoError(t, err)
	}
	require.NoError(t, k.upsertPoolValidatorAllocation(pool.PoolID, v1, 100, 2))
	require.NoError(t, k.upsertPoolValidatorAllocation(pool.PoolID, v2, 200, 2))

	service.RawStore().ResetWriteCounts()
	require.NoError(t, k.upsertPoolValidatorAllocation(pool.PoolID, v2, 300, 3))

	require.Zero(t, service.RawStore().SetCount(types.PoolAllocationKey(pool.PoolID, v1)))
	require.Equal(t, uint64(1), service.RawStore().SetCount(types.PoolAllocationKey(pool.PoolID, v2)))
	require.Zero(t, service.RawStore().DeleteCount(types.PoolAllocationKey(pool.PoolID, v1)))
}
