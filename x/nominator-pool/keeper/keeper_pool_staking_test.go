package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

type accountStatusFixture map[string]string

func (f accountStatusFixture) AccountStatus(address string) (string, bool) {
	status, found := f[address]
	return status, found
}

func TestPoolDepositMintsReceiptAndKeepsRawInternal(t *testing.T) {
	user := aePoolAddress(t, "22")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "official-staking")

	receipt, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, user, receipt.OwnerAddress)
	require.Equal(t, pool.ContractAddressUser, receipt.PoolContractAddressUser)
	require.Equal(t, types.DefaultMinPoolDeposit, receipt.Shares)
	require.Equal(t, types.DefaultParams().PoolReceiptDenomOrCodeID, receipt.ReceiptToken)
	require.Equal(t, rawPoolAddress("22"), receipt.InternalMetadata.OwnerRaw)
	require.Equal(t, pool.ContractAddressRaw, receipt.InternalMetadata.PoolContractAddressRaw)
	require.Equal(t, []string{
		string(types.PoolKey(pool.PoolID)),
		string(types.PoolShareKey(pool.PoolID, user)),
	}, receipt.InternalMetadata.TouchedKeys)

	exported := k.ExportGenesis()
	require.Len(t, exported.State.PoolShares, 1)
	require.Equal(t, user, exported.State.PoolShares[0].Owner)
	require.Equal(t, types.DefaultMinPoolDeposit, exported.State.PoolShares[0].Shares)
	require.Len(t, exported.State.LiquidStakingPools, 1)
	require.Equal(t, pool.ContractAddressUser, exported.State.LiquidStakingPools[0].ContractAddressUser)
	require.Equal(t, pool.ContractAddressRaw, exported.State.LiquidStakingPools[0].ContractAddressRaw)

	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:			pool.PoolID,
		WalletAddress:		user,
		ReservedRouting:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit,
		Height:			3,
	})
	require.ErrorContains(t, err, "must not include a routing field")
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	user := aePoolAddress(t, "52")
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	source.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, source.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &source, "official-persistent")

	receipt, err := source.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Pools, 1)
	require.Len(t, exported.State.PoolShares, 1)
	require.Equal(t, receipt.Shares, exported.State.PoolShares[0].Shares)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	imported.accountStatusReader = accountStatusFixture{user: accountStatusActive}
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	share, found := imported.PoolShare(types.QueryPoolShareRequest{PoolID: pool.PoolID, Delegator: rawPoolAddress("52")})
	require.True(t, found)
	require.Equal(t, receipt.Shares, share.Share.Shares)
}

func TestPoolDepositRejectsInactiveFrozenLowAndFrozenLimitedPool(t *testing.T) {
	active := aePoolAddress(t, "21")
	inactive := aePoolAddress(t, "22")
	frozen := aePoolAddress(t, "23")
	k := NewKeeperWithAccountStatus(accountStatusFixture{
		active:		accountStatusActive,
		inactive:	accountStatusInactive,
		frozen:		accountStatusFrozen,
	})
	pool := createOfficialLiquidStakingPool(t, &k, "official-rejects")

	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: inactive, Amount: types.DefaultMinPoolDeposit, Height: 2})
	require.ErrorContains(t, err, "requires active wallet")

	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: frozen, Amount: types.DefaultMinPoolDeposit, Height: 2})
	require.ErrorContains(t, err, "frozen wallet")

	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: active, Amount: types.DefaultMinPoolDeposit - 1, Height: 2})
	require.ErrorContains(t, err, "below configured minimum")

	gs := k.ExportGenesis()
	gs.State.Pools[0].Status = types.PoolStatusFrozenLimited
	gs.State.LiquidStakingPools[0].Status = types.PoolStatusFrozenLimited
	require.NoError(t, k.InitGenesis(gs))
	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: active, Amount: types.DefaultMinPoolDeposit, Height: 3})
	require.ErrorContains(t, err, "must be active for deposits")

	_, err = k.TopUpPoolReserve(types.MsgTopUpPoolReserve{PoolID: pool.PoolID, PayerAddress: active, Amount: 0, Height: 3})
	require.ErrorContains(t, err, "amount and height must be positive")

	_, err = k.TopUpPoolReserve(types.MsgTopUpPoolReserve{PoolID: pool.PoolID, PayerAddress: inactive, Amount: 1, Height: 3})
	require.ErrorContains(t, err, "requires active wallet")

	_, err = k.TopUpPoolReserve(types.MsgTopUpPoolReserve{PoolID: pool.PoolID, PayerAddress: frozen, Amount: 1, Height: 3})
	require.ErrorContains(t, err, "frozen wallet")
}

func TestFrozenLimitedPoolAllowsTopUpClaimUnbondAndMaturedWithdrawals(t *testing.T) {
	user := aePoolAddress(t, "24")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "official-exits")
	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: user, Amount: 2 * types.DefaultMinPoolDeposit, Height: 2})
	require.NoError(t, err)
	_, err = k.ApplyPoolReward(pool.PoolID, 100)
	require.NoError(t, err)

	gs := k.ExportGenesis()
	gs.State.Pools[0].Status = types.PoolStatusFrozenLimited
	gs.State.LiquidStakingPools[0].Status = types.PoolStatusFrozenLimited
	gs.State.LiquidStakingPools[0].StorageRentDebt = 123
	require.NoError(t, k.InitGenesis(gs))

	topUp, err := k.TopUpPoolReserve(types.MsgTopUpPoolReserve{
		PoolID:		pool.PoolID,
		PayerAddress:	user,
		Amount:		50,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(50), topUp.StorageDebtPaid)
	require.Equal(t, []string{
		string(types.PoolKey(pool.PoolID)),
		string(types.PoolStorageDebtKey(pool.PoolID)),
	}, topUp.InternalMetadata.TouchedKeys)
	exportedAfterTopUp := k.ExportGenesis()
	require.Greater(t, exportedAfterTopUp.State.LiquidStakingPools[0].StorageRentDebt, uint64(0))
	require.Equal(t, types.PoolStatusFrozenLimited, exportedAfterTopUp.State.LiquidStakingPools[0].Status)

	claim, err := k.ClaimPoolRewardsWithReceipt(types.MsgClaimPoolRewards{PoolID: pool.PoolID, OwnerAddress: user, Height: 4})
	require.NoError(t, err)
	require.NotZero(t, claim.Amount)

	unbond, err := k.RequestPoolUnbond(types.MsgRequestPoolUnbond{
		PoolID:		pool.PoolID,
		OwnerAddress:	user,
		RequestID:	"unbond-1",
		Shares:		types.DefaultMinPoolDeposit,
		Height:		5,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(5)+k.ExportGenesis().Params.UnbondingBlocks, unbond.CompleteHeight)

	_, err = k.WithdrawPoolStake(types.MsgWithdrawPoolStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		OwnerAddress:		user,
		RequestID:		"unbond-1",
		Height:			unbond.CompleteHeight - 1,
	})
	require.ErrorContains(t, err, "before unbonding period")

	withdrawal, err := k.WithdrawPoolStake(types.MsgWithdrawPoolStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		OwnerAddress:		user,
		RequestID:		"unbond-1",
		Height:			unbond.CompleteHeight,
	})
	require.NoError(t, err)
	require.Equal(t, unbond.Amount, withdrawal.Amount)
	require.Contains(t, withdrawal.InternalMetadata.TouchedKeys, string(types.PoolUnbondingKey(pool.PoolID, user, "unbond-1")))
}

func TestValidatorRegistrationUpdateAndDuplicate(t *testing.T) {
	validator := aePoolAddress(t, "31")
	k := NewKeeperWithAccountStatus(accountStatusFixture{validator: accountStatusActive})

	_, err := k.RegisterValidator(types.MsgRegisterValidator{
		SignerAddress:		validator,
		ValidatorAddress:	validator,
		SelfStake:		types.DefaultMinValidatorStake - 1,
		CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
		Height:			1,
	})
	require.ErrorContains(t, err, "minimum validator stake")

	receipt, err := k.RegisterValidator(types.MsgRegisterValidator{
		SignerAddress:		validator,
		ValidatorAddress:	validator,
		SelfStake:		types.DefaultMinValidatorStake,
		CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, []string{string(types.ValidatorKey(validator))}, receipt.TouchedKeys)

	_, err = k.RegisterValidator(types.MsgRegisterValidator{
		SignerAddress:		validator,
		ValidatorAddress:	validator,
		SelfStake:		types.DefaultMinValidatorStake,
		CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
		Height:			3,
	})
	require.ErrorContains(t, err, "already registered")

	updated, err := k.UpdateValidator(types.MsgUpdateValidator{
		SignerAddress:		validator,
		ValidatorAddress:	validator,
		PerformanceScore:	9_500,
		CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps + 1,
		AllocationLimitBps:	types.MaxBasisPoints,
		Status:			types.StateValidatorStatusActive,
		Height:			4,
	})
	require.NoError(t, err)
	require.Equal(t, validator, updated.Validator)
}

func TestInjectAndRebalanceAllocationsAreDeterministicAndBounded(t *testing.T) {
	user := aePoolAddress(t, "40")
	v1 := aePoolAddress(t, "41")
	v2 := aePoolAddress(t, "42")
	v3 := aePoolAddress(t, "43")
	k := NewKeeperWithAccountStatus(accountStatusFixture{
		user:	accountStatusActive,
		v1:	accountStatusActive,
		v2:	accountStatusActive,
		v3:	accountStatusActive,
	})
	pool := createOfficialLiquidStakingPool(t, &k, "official-alloc")
	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: user, Amount: 100 * types.DefaultAETBaseUnits, Height: 2})
	require.NoError(t, err)
	for _, validator := range []string{v1, v2, v3} {
		_, err := k.RegisterValidator(types.MsgRegisterValidator{
			SignerAddress:		validator,
			ValidatorAddress:	validator,
			SelfStake:		types.DefaultMinValidatorStake,
			CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
			Height:			3,
		})
		require.NoError(t, err)
	}

	injected, err := k.InjectPoolStake(types.MsgInjectPoolStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		Height:			4,
		Allocations: []types.PoolAllocation{
			{ValidatorAddress: v2, Amount: 40 * types.DefaultAETBaseUnits, Height: 4},
			{ValidatorAddress: v1, Amount: 60 * types.DefaultAETBaseUnits, Height: 4},
		},
	})
	require.NoError(t, err)
	require.Len(t, injected.Allocations, 2)
	require.Equal(t, []string{
		string(types.PoolKey(pool.PoolID)),
		string(types.PoolAllocationKey(pool.PoolID, v1)),
		string(types.PoolAllocationKey(pool.PoolID, v2)),
	}, injected.InternalMetadata.TouchedKeys)

	beforeShare, found := k.PoolDelegator(pool.PoolID, rawPoolAddress("40"))
	require.True(t, found)
	rebalanced, err := k.RebalancePoolAllocations(types.MsgRebalancePoolAllocations{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		Epoch:			1,
		Height:			5,
		Candidates: []types.ValidatorPolicyCandidate{
			{ValidatorAddress: v3, ReputationScore: 6_000, UptimeBps: 8_000, CommissionBps: 1_000, StakeEfficiencyBps: 7_000, SlashingRiskBps: 200, NetworkLoadBps: 2_000},
			{ValidatorAddress: v1, ReputationScore: 9_000, UptimeBps: 9_500, CommissionBps: 500, StakeEfficiencyBps: 9_000, SlashingRiskBps: 100, NetworkLoadBps: 1_000},
			{ValidatorAddress: v2, ReputationScore: 7_000, UptimeBps: 9_000, CommissionBps: 1_500, StakeEfficiencyBps: 8_000, SlashingRiskBps: 300, NetworkLoadBps: 2_500},
		},
	})
	require.NoError(t, err)
	require.Len(t, rebalanced.Allocations, 3)
	require.Equal(t, v1, rebalanced.Allocations[0].Validator)
	require.Equal(t, v2, rebalanced.Allocations[1].Validator)
	require.Equal(t, v3, rebalanced.Allocations[2].Validator)
	afterShare, found := k.PoolDelegator(pool.PoolID, rawPoolAddress("40"))
	require.True(t, found)
	require.Equal(t, beforeShare.Shares, afterShare.Shares)
	require.Equal(t, []string{
		string(types.PoolKey(pool.PoolID)),
		string(types.PoolAllocationKey(pool.PoolID, v1)),
		string(types.PoolAllocationKey(pool.PoolID, v2)),
		string(types.PoolAllocationKey(pool.PoolID, v3)),
	}, rebalanced.InternalMetadata.TouchedKeys)
}

func TestStakeReputationClaimTouchesOnlyShareKey(t *testing.T) {
	user := aePoolAddress(t, "50")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "official-reputation")
	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: user, Amount: types.DefaultMinPoolDeposit, Height: 2})
	require.NoError(t, err)
	gs := k.ExportGenesis()
	gs.State.LiquidStakingPools[0].TotalActiveStake = types.DefaultMinPoolDeposit
	require.NoError(t, k.InitGenesis(gs))

	claim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: user, Height: 12})
	require.NoError(t, err)
	require.NotZero(t, claim.ReputationDelta)
	require.Equal(t, []string{
		string(types.PoolShareKey(pool.PoolID, user)),
	}, claim.InternalMetadata.TouchedKeys)

	exported := k.ExportGenesis()
	imported := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	require.NoError(t, imported.InitGenesis(exported))
}

func TestStakeReputationNoActiveExposureNoIncrease(t *testing.T) {
	user := aePoolAddress(t, "51")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "official-no-exposure")
	_, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{PoolID: pool.PoolID, WalletAddress: user, Amount: types.DefaultMinPoolDeposit, Height: 2})
	require.NoError(t, err)

	claim, err := k.ClaimStakeReputation(types.MsgClaimStakeReputation{PoolID: pool.PoolID, OwnerAddress: user, Height: 12})
	require.NoError(t, err)
	require.Zero(t, claim.ReputationDelta)
	require.Equal(t, []string{string(types.PoolShareKey(pool.PoolID, user))}, claim.InternalMetadata.TouchedKeys)
}

func TestUpdateStakingParamsAlias(t *testing.T) {
	k := NewKeeper()
	next := k.ExportGenesis().Params
	next.TargetValidatorCount = 129
	updated, err := k.UpdateStakingParams(types.MsgUpdateStakingParams{Authority: prototype.DefaultAuthority, Params: next, Height: 2})
	require.NoError(t, err)
	require.Equal(t, uint32(129), updated.TargetValidatorCount)
}
