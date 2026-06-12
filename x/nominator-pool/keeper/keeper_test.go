package keeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestNominatorPoolUnbondingUsesStakingPolicyWindow(t *testing.T) {
	gs := DefaultGenesis()

	require.Equal(t, types.DefaultUnbondingBlocks, gs.Params.UnbondingBlocks)
	require.NoError(t, gs.Validate())

	gs.Params.UnbondingBlocks = appparams.StakingUnbondingMinBlocks - 1
	require.ErrorContains(t, gs.Validate(), "14-21 days")
}

func TestDefaultStakingParamsUseArchitectureGenesisValues(t *testing.T) {
	params := DefaultGenesis().Params

	require.Equal(t, appparams.BaseDenom, params.BaseDenom)
	require.Equal(t, appparams.DisplayDenom, params.DisplayDenom)
	require.Equal(t, appparams.DisplayDenomExponent, params.DisplayExponent)
	require.Equal(t, uint64(1_000_000)*types.DefaultAETBaseUnits, params.MinValidatorStake)
	require.Equal(t, uint64(1_000_000)*types.DefaultAETBaseUnits, params.SoloValidatorMinSelfStake)
	require.Equal(t, uint64(400_000)*types.DefaultAETBaseUnits, params.PoolBackedValidatorMinSelfStake)
	require.Equal(t, uint64(600_000)*types.DefaultAETBaseUnits, params.PoolBackedValidatorMaxNominatorStake)
	require.Equal(t, uint32(4_000), params.ValidatorSelfStakeMinRatioBps)
	require.Equal(t, uint32(6_000), params.ValidatorNominatorStakeMaxRatioBps)
	require.Equal(t, uint64(10)*types.DefaultAETBaseUnits, params.MinPoolDeposit)
	require.False(t, params.DirectUserValidatorDelegationEnabled)
	require.Equal(t, uint32(100), params.GovernanceMinValidatorCount)
	require.Equal(t, uint32(128), params.TargetValidatorCount)
	require.Equal(t, uint32(300), params.MaxValidatorCount)
	require.Equal(t, uint32(300), params.GovernanceMaxValidatorCount)
	require.Equal(t, uint32(300), params.ValidatorPowerCapBps)
	require.Equal(t, uint32(500), params.ValidatorCommissionFloorBps)
	require.Equal(t, uint32(1_000), params.DefaultValidatorCommissionBps)
	require.Equal(t, uint32(2_000), params.ValidatorCommissionCeilingBps)
	require.Equal(t, uint32(100), params.ValidatorCommissionMaxDailyChangeBps)
	require.Equal(t, uint64(18*24*60*60/appparams.StakingUnbondingBlockTimeSeconds), params.UnbondingBlocks)
	require.Equal(t, uint64(24*60*60/appparams.StakingUnbondingBlockTimeSeconds), params.RewardEpochDurationBlocks)
	require.Equal(t, uint32(5_000), params.BurnFeeShareBps)
	require.Equal(t, uint32(3_500), params.RewardFeeShareBps)
	require.Equal(t, uint32(1_500), params.TreasuryFeeShareBps)
	require.Equal(t, uint64(3_000_000), params.MinTxFeeBaseUnits)
	require.Equal(t, uint64(1), params.StorageRentRatePerByteSecond)
	require.Equal(t, uint64(365), params.SystemStorageReserveMinRunwayDays)
	require.Equal(t, uint64(180), params.SystemStorageReserveWarningRunwayDays)
	require.Equal(t, uint64(90), params.SystemStorageReserveCriticalRunwayDays)
	require.NoError(t, params.Validate())
}

func TestValidatorFundingPolicyRejectsInsufficientStake(t *testing.T) {
	params := DefaultGenesis().Params

	require.ErrorContains(t, params.ValidateValidatorFunding(types.ValidatorFunding{
		Mode:		types.ValidatorFundingSolo,
		SelfStake:	params.MinValidatorStake - 1,
	}), "minimum validator stake")
	require.ErrorContains(t, params.ValidateValidatorFunding(types.ValidatorFunding{
		Mode:		types.ValidatorFundingPoolBacked,
		SelfStake:	params.PoolBackedValidatorMinSelfStake - 1,
		NominatorStake:	params.PoolBackedValidatorMaxNominatorStake + 1,
	}), "self-stake")
	require.ErrorContains(t, params.ValidateValidatorFunding(types.ValidatorFunding{
		Mode:		types.ValidatorFundingPoolBacked,
		SelfStake:	params.PoolBackedValidatorMinSelfStake,
		NominatorStake:	params.PoolBackedValidatorMaxNominatorStake + 1,
	}), "nominator stake")

	ratioParams := params
	ratioParams.PoolBackedValidatorMaxNominatorStake = params.MinValidatorStake
	require.ErrorContains(t, ratioParams.ValidateValidatorFunding(types.ValidatorFunding{
		Mode:		types.ValidatorFundingPoolBacked,
		SelfStake:	params.PoolBackedValidatorMinSelfStake,
		NominatorStake:	params.PoolBackedValidatorMaxNominatorStake + types.DefaultAETBaseUnits,
	}), "self-stake ratio")
	require.NoError(t, params.ValidateValidatorFunding(types.ValidatorFunding{
		Mode:		types.ValidatorFundingPoolBacked,
		SelfStake:	params.PoolBackedValidatorMinSelfStake,
		NominatorStake:	params.PoolBackedValidatorMaxNominatorStake,
	}))
}

func TestValidatorCountCommissionAndPowerCapParams(t *testing.T) {
	params := DefaultGenesis().Params

	require.NoError(t, params.ValidateActiveValidatorCount(100, false))
	require.ErrorContains(t, params.ValidateActiveValidatorCount(99, false), "below governance minimum")
	require.NoError(t, params.ValidateActiveValidatorCount(99, true))
	require.ErrorContains(t, params.ValidateActiveValidatorCount(301, false), "exceeds configured maximum")
	cap150, err := params.PowerCapBpsForValidatorCount(150)
	require.NoError(t, err)
	cap250, err := params.PowerCapBpsForValidatorCount(250)
	require.NoError(t, err)
	cap251, err := params.PowerCapBpsForValidatorCount(251)
	require.NoError(t, err)
	require.Equal(t, uint32(300), cap150)
	require.Equal(t, uint32(250), cap250)
	require.Equal(t, uint32(200), cap251)

	require.NoError(t, params.ValidateCommission(1_000, 950, 50))
	require.ErrorContains(t, params.ValidateCommission(499, 500, 1), "below configured floor")
	require.ErrorContains(t, params.ValidateCommission(2_001, 2_000, 1), "above configured ceiling")
	require.ErrorContains(t, params.ValidateCommission(1_200, 1_000, 200), "daily change")
}

func TestAllocationWeightsAreDeterministicAndPolicyBounded(t *testing.T) {
	params := DefaultGenesis().Params
	candidates := []types.ValidatorPolicyCandidate{
		{
			ValidatorAddress:	aePoolAddress(t, "44"),
			ReputationScore:	8_000,
			UptimeBps:		9_000,
			CommissionBps:		1_000,
			StakeEfficiencyBps:	7_500,
			SlashingRiskBps:	250,
			NetworkLoadBps:		1_000,
			CurrentAllocationBps:	params.MaxPoolValidatorAllocationBps,
		},
		{
			ValidatorAddress:	aePoolAddress(t, "22"),
			ReputationScore:	6_000,
			UptimeBps:		9_500,
			CommissionBps:		500,
			StakeEfficiencyBps:	8_000,
			SlashingRiskBps:	100,
			NetworkLoadBps:		2_000,
		},
		{
			ValidatorAddress:	aePoolAddress(t, "33"),
			ReputationScore:	9_000,
			UptimeBps:		8_000,
			CommissionBps:		1_500,
			StakeEfficiencyBps:	7_000,
			SlashingRiskBps:	500,
			NetworkLoadBps:		1_000,
			Jailed:			true,
		},
	}
	weights, err := params.AllocationWeights(candidates)
	require.NoError(t, err)
	require.Equal(t, aePoolAddress(t, "22"), weights[0].ValidatorAddress)
	require.Equal(t, aePoolAddress(t, "33"), weights[1].ValidatorAddress)
	require.Equal(t, aePoolAddress(t, "44"), weights[2].ValidatorAddress)
	require.Equal(t, uint32(10_000), weights[0].WeightBps)
	require.Zero(t, weights[1].WeightBps)
	require.Zero(t, weights[2].WeightBps)

	candidates[1].UptimeBps = 1_000
	changed, err := params.AllocationWeights(candidates)
	require.NoError(t, err)
	require.Less(t, changed[0].Score, weights[0].Score)
}

func TestGovernanceParamsUpdateAuthorizationAndRoundTrip(t *testing.T) {
	source := NewKeeper()
	next := source.ExportGenesis().Params
	next.TargetValidatorCount = 150
	_, err := source.UpdateParams(types.MsgUpdateParams{
		Authority:	aePoolAddress(t, "22"),
		Params:		next,
		Height:		2,
	})
	require.ErrorContains(t, err, "governance authority")

	updated, err := source.UpdateParams(types.MsgUpdateParams{
		Authority:	prototype.DefaultAuthority,
		Params:		next,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(150), updated.TargetValidatorCount)

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
}

func TestDepositMintsShares(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")

	share, err := k.DepositToPool(types.MsgDepositToPool{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		Delegator:	rawPoolAddress("22"),
		Amount:		1_000,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1_000), share.Shares)
	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(1_000), stored.TotalShares)
	require.Equal(t, uint64(1_000), stored.TotalBondedStake)
}

func TestOfficialLiquidStakingSmallDepositMintsDeterministicShares(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	user := aePoolAddress(t, "22")

	share, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, types.DefaultMinPoolDeposit, share.Shares)

	secondShare, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	user,
		Amount:		90 * types.DefaultAETBaseUnits,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(100)*types.DefaultAETBaseUnits, secondShare.Shares)

	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(100)*types.DefaultAETBaseUnits, stored.TotalBondedStake)
	require.Equal(t, uint64(100)*types.DefaultAETBaseUnits, stored.TotalShares)
	require.Equal(t, rawPoolAddress("22"), stored.DelegatorShares[0].Delegator)
	require.Equal(t, uint64(100)*types.DefaultAETBaseUnits, stored.DelegatorShares[0].Shares)
}

func TestPoolShareMintingUsesOverflowSafeBaseUnitMath(t *testing.T) {
	pool := types.NominatorPool{
		TotalShares:		types.DefaultMinPoolDeposit,
		TotalBondedStake:	types.DefaultMinPoolDeposit,
	}
	shares, err := types.SharesForDepositChecked(pool, 90*types.DefaultAETBaseUnits)
	require.NoError(t, err)
	require.Equal(t, uint64(90)*types.DefaultAETBaseUnits, shares)
}

func TestOfficialLiquidStakingDepositRejectsValidatorAddress(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		UserAddress:		aePoolAddress(t, "22"),
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit,
		Height:			2,
	})
	require.ErrorContains(t, err, "must not include a validator address")
}

func TestOfficialLiquidStakingDepositBelowMinimumRejected(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	aePoolAddress(t, "22"),
		Amount:		types.DefaultMinPoolDeposit - 1,
		Height:		2,
	})
	require.ErrorContains(t, err, "below configured minimum")
}

func TestOfficialLiquidStakingDepositRequiresAEUserAddress(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	rawPoolAddress("22"),
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.ErrorContains(t, err, "must use AE user-facing address format")
}

func TestOfficialLiquidStakingReceiptExportImportRoundTrip(t *testing.T) {
	source := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &source, "official-pool")
	user := aePoolAddress(t, "22")
	share, err := source.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))

	roundTripShare, found := target.PoolDelegator(pool.PoolID, rawPoolAddress("22"))
	require.True(t, found)
	require.Equal(t, share, roundTripShare)
	require.Equal(t, exported, target.ExportGenesis())
}

func TestDirectUserDelegationDisabledByDefault(t *testing.T) {
	k := NewKeeper()

	err := k.DelegateUserToValidator(types.MsgDelegateToValidator{
		Authority:		prototype.DefaultAuthority,
		UserAddress:		aePoolAddress(t, "22"),
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			100,
		Height:			2,
	})
	require.ErrorContains(t, err, "direct user delegation to validators is disabled")
}

func TestOfficialContractCanInjectPooledStake(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	aePoolAddress(t, "22"),
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	updated, err := k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit,
		Height:			3,
	})
	require.NoError(t, err)
	require.Equal(t, []types.PoolAllocation{{
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit,
		Height:			3,
	}}, updated.Allocations)
}

func TestUnauthorizedContractCannotInjectPooledStake(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	aePoolAddress(t, "22"),
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	_, err = k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser:	aePoolAddress(t, "77"),
		PoolID:			pool.PoolID,
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit,
		Height:			3,
	})
	require.ErrorContains(t, err, "requires official liquid staking contract")
}

func TestPooledStakeInjectionCannotExceedPoolAccounting(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	aePoolAddress(t, "22"),
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	_, err = k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit + 1,
		Height:			3,
	})
	require.ErrorContains(t, err, "exceeds unallocated pool stake")
}

func TestFrozenLimitedOfficialPoolRejectsDeposits(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	gs := k.ExportGenesis()
	gs.State.Pools[0].Status = types.PoolStatusFrozenLimited
	require.NoError(t, k.InitGenesis(gs))

	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	aePoolAddress(t, "22"),
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.ErrorContains(t, err, "must be active for deposits")
}

func TestFrozenLimitedOfficialPoolRejectsPooledStakeInjection(t *testing.T) {
	k := NewKeeper()
	pool := createOfficialLiquidStakingPool(t, &k, "official-pool")
	_, err := k.DepositToOfficialLiquidStaking(types.MsgDepositToOfficialLiquidStaking{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		UserAddress:	aePoolAddress(t, "22"),
		Amount:		types.DefaultMinPoolDeposit,
		Height:		2,
	})
	require.NoError(t, err)

	gs := k.ExportGenesis()
	gs.State.Pools[0].Status = types.PoolStatusFrozenLimited
	gs.State.LiquidStakingPools[0].Status = types.PoolStatusFrozenLimited
	require.NoError(t, k.InitGenesis(gs))

	_, err = k.InjectPooledStake(types.MsgInjectPooledStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		ValidatorAddress:	aePoolAddress(t, "33"),
		Amount:			types.DefaultMinPoolDeposit,
		Height:			3,
	})
	require.ErrorContains(t, err, "must be active for stake injection")
}

func TestWithdrawalBurnsShares(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 1_000, 2)

	withdrawal, err := k.RequestPoolWithdrawal(types.MsgRequestPoolWithdrawal{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		WithdrawalID:	"withdraw-1",
		Delegator:	rawPoolAddress("22"),
		Shares:		400,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(400), withdrawal.Amount)
	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(600), stored.TotalShares)
	require.Equal(t, uint64(600), stored.TotalBondedStake)
	require.Len(t, stored.UnbondingQueue, 1)
}

func TestRewardsDistributeProportionally(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	a := rawPoolAddress("22")
	b := rawPoolAddress("33")
	deposit(t, &k, pool.PoolID, a, 1_000, 2)
	deposit(t, &k, pool.PoolID, b, 3_000, 3)

	_, err := k.ApplyPoolReward(pool.PoolID, 400)
	require.NoError(t, err)
	rewardA, found := k.PoolRewards(pool.PoolID, a)
	require.True(t, found)
	rewardB, found := k.PoolRewards(pool.PoolID, b)
	require.True(t, found)
	require.Equal(t, uint64(99), rewardA)
	require.Equal(t, uint64(297), rewardB)

	claimed, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{Authority: prototype.DefaultAuthority, PoolID: pool.PoolID, Delegator: a, Height: 4})
	require.NoError(t, err)
	require.Equal(t, uint64(99), claimed)
}

func TestSyncPoolRewardsAfterEpochProgressionAndIllustrativeEconomics(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-rewards")
	user := rawPoolAddress("22")
	deposit(t, &k, pool.PoolID, user, 3_000_000, 2)

	summary, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		RewardRateBps:		1_440,
		EmissionsAllocated:	864_000,
		Height:			3,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:		rawPoolAddress("12"),
			PoolAllocatedStake:	3_000_000,
			ValidatorSelfStake:	3_000_000,
			PerformanceBps:		types.MaxBasisPoints,
			CommissionBps:		1_000,
			InfrastructureCost:	20_000,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(432_000), summary.GrossPoolRewards)
	require.Equal(t, uint64(43_200), summary.ValidatorCommission)
	require.Equal(t, uint64(3_888), summary.PoolProtocolFee)
	require.Equal(t, uint64(384_912), summary.PoolUserRewards)
	require.Equal(t, uint64(432_000), summary.ValidatorSelfStakeRewards)
	require.Equal(t, uint64(475_200), summary.ValidatorGrossIncome)
	require.Equal(t, int64(455_200), summary.ValidatorNetIncome)
	require.Greater(t, summary.ValidatorGrossIncome, summary.PoolUserRewards)

	claimed, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{Authority: prototype.DefaultAuthority, PoolID: pool.PoolID, Delegator: user, Height: 4})
	require.NoError(t, err)
	require.Equal(t, uint64(384_912), claimed)

	claimedAgain, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{Authority: prototype.DefaultAuthority, PoolID: pool.PoolID, Delegator: user, Height: 5})
	require.NoError(t, err)
	require.Zero(t, claimedAgain)
}

func TestSyncPoolRewardsProportionalSharesAndDeterministicRounding(t *testing.T) {
	k := NewKeeper()
	pool := createPoolWithCommission(t, &k, "pool-proportional", 0)
	a := rawPoolAddress("22")
	b := rawPoolAddress("33")
	deposit(t, &k, pool.PoolID, a, 1_000, 2)
	deposit(t, &k, pool.PoolID, b, 3_000, 3)

	require.Equal(t, uint64(333_333_333), types.RewardDelta(1, 3))
	require.Equal(t, uint64(0), types.IndexedRewardAmount(types.RewardDelta(1, 3), 3))

	summary, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		RewardRateBps:		1_000,
		EmissionsAllocated:	400,
		Height:			4,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:		rawPoolAddress("12"),
			PoolAllocatedStake:	4_000,
			PerformanceBps:		types.MaxBasisPoints,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(400), summary.PoolUserRewards)

	rewardA, found := k.PoolRewards(pool.PoolID, a)
	require.True(t, found)
	rewardB, found := k.PoolRewards(pool.PoolID, b)
	require.True(t, found)
	require.Equal(t, uint64(100), rewardA)
	require.Equal(t, uint64(300), rewardB)
}

func TestValidatorPerformanceCommissionPoolFeeAndJailAreDeterministic(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-performance")
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 2_000, 2)

	summary, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		RewardRateBps:		1_000,
		EmissionsAllocated:	2_000,
		Height:			3,
		Allocations: []types.ValidatorRewardAllocation{
			{
				Validator:		rawPoolAddress("12"),
				PoolAllocatedStake:	1_000,
				PerformanceBps:		types.MaxBasisPoints,
				CommissionBps:		1_000,
			},
			{
				Validator:			rawPoolAddress("13"),
				PoolAllocatedStake:		1_000,
				PerformanceBps:			5_000,
				CommissionBps:			1_000,
				Jailed:				true,
				OperatorPerformanceBonusBps:	100,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(100), summary.GrossPoolRewards)
	require.Equal(t, uint64(10), summary.ValidatorCommission)
	require.Equal(t, uint64(0), summary.OperatorPerformanceBonus)
	require.Equal(t, uint64(90), summary.PoolUserRewards)

	allocations, found := k.PoolAllocations(types.QueryPoolAllocationsRequest{PoolID: pool.PoolID})
	require.True(t, found)
	require.Len(t, allocations.Allocations, 2)
	require.Equal(t, uint64(100), allocations.Allocations[0].GrossPoolRewards)
	require.Zero(t, allocations.Allocations[1].GrossPoolRewards)
	require.Zero(t, allocations.Allocations[1].OperatorPerformanceBonus)
}

func TestSyncPoolRewardsRejectsEmissionFeeCapExceeded(t *testing.T) {
	k := NewKeeper()
	pool := createPoolWithCommission(t, &k, "pool-cap", 0)
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 1_000, 2)

	_, err := k.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		RewardRateBps:		1_000,
		EmissionsAllocated:	99,
		Height:			3,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:		rawPoolAddress("12"),
			PoolAllocatedStake:	1_000,
			PerformanceBps:		types.MaxBasisPoints,
		}},
	})
	require.ErrorContains(t, err, "exceed emissions")
}

func TestExportImportPreservesSyncedRewardStateAndPendingRewards(t *testing.T) {
	source := NewKeeper()
	pool := createPoolWithCommission(t, &source, "pool-export", 0)
	user := rawPoolAddress("22")
	deposit(t, &source, pool.PoolID, user, 3, 2)
	_, err := source.SyncPoolRewards(types.MsgSyncPoolRewards{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		Epoch:			1,
		RewardRateBps:		3_334,
		EmissionsAllocated:	1,
		Height:			3,
		Allocations: []types.ValidatorRewardAllocation{{
			Validator:		rawPoolAddress("12"),
			PoolAllocatedStake:	3,
			PerformanceBps:		types.MaxBasisPoints,
		}},
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	rewards, found := target.PoolRewards(pool.PoolID, user)
	require.True(t, found)
	require.Equal(t, uint64(0), rewards)
	stored, found := target.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, uint64(1), stored.RewardRemainder)
	require.Len(t, stored.ValidatorAllocations, 1)
}

func TestClaimPoolRewardsTouchesBoundedKeysWithLargeUserSet(t *testing.T) {
	k := NewKeeper()
	const users = 1_000_000
	shares := make([]types.DelegatorShare, users)
	for i := range shares {
		shares[i] = types.DelegatorShare{
			Delegator:		rawPoolAddressFromInt(i + 1),
			Shares:			1,
			RewardIndexCheckpoint:	0,
		}
	}
	k.genesis.State.Pools = []types.NominatorPool{{
		PoolID:			"pool-large",
		PoolOperator:		rawPoolAddress("11"),
		ValidatorTarget:	rawPoolAddress("12"),
		TotalShares:		users,
		TotalBondedStake:	users,
		DelegatorShares:	shares,
		RewardIndex:		types.IndexScale,
		PoolCommissionBps:	100,
		Status:			types.PoolStatusActive,
	}}
	k.rebuildIndexes()
	k.ResetOperationCounters()

	claimed, err := k.ClaimPoolRewards(types.MsgClaimPoolRewards{
		Authority:	prototype.DefaultAuthority,
		PoolID:		"pool-large",
		Delegator:	rawPoolAddressFromInt(users),
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), claimed)
	counters := k.OperationCounters()
	require.Equal(t, uint64(1), counters.PoolLookups)
	require.Equal(t, uint64(1), counters.DelegatorLookups)
	require.Equal(t, uint64(1), counters.DelegatorRewardUpdates)
}

func TestStakingRewardsCompatibilityIsInternalOnly(t *testing.T) {
	k := NewKeeper()
	_, err := k.ClaimStakingRewards(types.MsgClaimStakingRewards{
		Authority:	prototype.DefaultAuthority,
		Delegator:	rawPoolAddress("22"),
		Validator:	rawPoolAddress("12"),
		Height:		1,
	})
	require.ErrorContains(t, err, "internal migration only")

	amount, err := k.ClaimStakingRewards(types.MsgClaimStakingRewards{
		Authority:		prototype.DefaultAuthority,
		Delegator:		rawPoolAddress("22"),
		Validator:		rawPoolAddress("12"),
		Height:			1,
		InternalMigration:	true,
	})
	require.NoError(t, err)
	require.Zero(t, amount)

	_, err = k.StakingRewards(types.QueryStakingRewardsRequest{Delegator: rawPoolAddress("22"), Validator: rawPoolAddress("12")})
	require.ErrorContains(t, err, "internal migration only")
}

func TestStakingProofQueryIsBoundedAndReturnsMetadata(t *testing.T) {
	k := NewKeeper()
	k.ResetOperationCounters()
	account := proofUserAddress("55")
	appHash := "app-root-ref"
	poolRoot := "nominator-pool-root-ref"
	reputationRoot := "reputation-root-ref"
	cases := []struct {
		name		string
		req		types.StakingProofRequest
		storeKey	string
		stateKey	string
		rootHash	string
	}{
		{
			name:		"deposit",
			req:		types.StakingProofRequest{Kind: types.StakingProofDeposit, Height: 42, PoolID: "pool-a", Account: account, AppHash: appHash, RootHash: poolRoot},
			storeKey:	types.StoreKey,
			stateKey:	types.PoolDepositProofStateKey("pool-a", account),
			rootHash:	poolRoot,
		},
		{
			name:		"share",
			req:		types.StakingProofRequest{Kind: types.StakingProofShare, Height: 42, PoolID: "pool-a", Account: account, AppHash: appHash, RootHash: poolRoot},
			storeKey:	types.StoreKey,
			stateKey:	types.PoolShareProofStateKey("pool-a", account),
			rootHash:	poolRoot,
		},
		{
			name:		"allocation",
			req:		types.StakingProofRequest{Kind: types.StakingProofAllocation, Height: 42, PoolID: "pool-a", Epoch: 7, AppHash: appHash, RootHash: poolRoot},
			storeKey:	types.StoreKey,
			stateKey:	types.PoolAllocationProofStateKey("pool-a", 7),
			rootHash:	poolRoot,
		},
		{
			name:		"reward",
			req:		types.StakingProofRequest{Kind: types.StakingProofReward, Height: 42, PoolID: "pool-a", Account: account, AppHash: appHash, RootHash: poolRoot},
			storeKey:	types.StoreKey,
			stateKey:	types.PoolRewardProofStateKey("pool-a", account),
			rootHash:	poolRoot,
		},
		{
			name:		"reputation",
			req:		types.StakingProofRequest{Kind: types.StakingProofReputation, Height: 42, Account: account, AppHash: appHash, RootHash: reputationRoot},
			storeKey:	reputationtypes.StoreKey,
			stateKey:	types.StakeReputationProofStateKey(account),
			rootHash:	reputationRoot,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			metadata, err := k.StakingProof(tc.req)
			require.NoError(t, err)
			require.Equal(t, tc.storeKey, metadata.StoreKey)
			require.Equal(t, tc.stateKey, metadata.StateKey)
			require.Equal(t, uint64(42), metadata.Height)
			require.Equal(t, appHash, metadata.AppHash)
			require.Equal(t, tc.rootHash, metadata.RootHash)
			require.Len(t, metadata.ProofPath, 2)
			require.Equal(t, tc.storeKey, metadata.ProofPath[1].StoreKey)
			require.Equal(t, tc.stateKey, metadata.ProofPath[1].StateKey)
			require.True(t, metadata.BoundedLookup)
		})
	}

	counters := k.OperationCounters()
	require.Equal(t, uint64(len(cases)), counters.ProofQueries)
	require.Zero(t, counters.PoolLookups)
	require.Zero(t, counters.DelegatorLookups)
	require.Zero(t, counters.DelegatorRewardUpdates)
}

func TestSlashAppliesProportionally(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	a := rawPoolAddress("22")
	b := rawPoolAddress("33")
	deposit(t, &k, pool.PoolID, a, 1_000, 2)
	deposit(t, &k, pool.PoolID, b, 3_000, 3)

	slashed, err := k.ApplyPoolSlash(pool.PoolID, 800)
	require.NoError(t, err)
	require.Equal(t, uint64(3_200), slashed.TotalBondedStake)

	shareA, found := k.PoolDelegator(pool.PoolID, a)
	require.True(t, found)
	shareB, found := k.PoolDelegator(pool.PoolID, b)
	require.True(t, found)
	require.Equal(t, uint64(800), types.ShareValue(slashed, shareA.Shares))
	require.Equal(t, uint64(2_400), types.ShareValue(slashed, shareB.Shares))
}

func TestPoolCannotWithdrawMoreThanTotalStake(t *testing.T) {
	k := NewKeeper()
	pool := createPool(t, &k, "pool-a")
	deposit(t, &k, pool.PoolID, rawPoolAddress("22"), 1_000, 2)

	_, err := k.RequestPoolWithdrawal(types.MsgRequestPoolWithdrawal{
		Authority:	prototype.DefaultAuthority,
		PoolID:		pool.PoolID,
		WithdrawalID:	"withdraw-too-much",
		Delegator:	rawPoolAddress("22"),
		Shares:		1_001,
		Height:		3,
	})
	require.ErrorContains(t, err, "more than total stake")
}

func TestExportImportPreservesRewardIndex(t *testing.T) {
	source := NewKeeper()
	pool := createPool(t, &source, "pool-a")
	deposit(t, &source, pool.PoolID, rawPoolAddress("22"), 1_000, 2)
	_, err := source.ApplyPoolReward(pool.PoolID, 123)
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
	stored, found := target.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, exported.State.Pools[0].RewardIndex, stored.RewardIndex)
}

func TestPoolValidatorChangeDelayEnforced(t *testing.T) {
	k := NewKeeper()
	gs := k.ExportGenesis()
	gs.Params.ValidatorChangeDelay = 10
	require.NoError(t, k.InitGenesis(gs))
	pool := createPool(t, &k, "pool-a")
	nextValidator := rawPoolAddress("44")

	pending, err := k.ChangePoolValidator(types.MsgChangePoolValidator{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		PoolOperator:		pool.PoolOperator,
		ValidatorTarget:	nextValidator,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
		Height:			5,
	})
	require.NoError(t, err)
	require.Equal(t, pool.ValidatorTarget, pending.ValidatorTarget)
	require.Equal(t, nextValidator, pending.PendingValidatorTarget)
	require.Equal(t, uint64(15), pending.ValidatorChangeHeight)

	stillPending, err := k.ChangePoolValidator(types.MsgChangePoolValidator{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		PoolOperator:		pool.PoolOperator,
		ValidatorTarget:	nextValidator,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
		Height:			14,
	})
	require.NoError(t, err)
	require.NotEqual(t, nextValidator, stillPending.ValidatorTarget)

	finalized, err := k.ChangePoolValidator(types.MsgChangePoolValidator{
		Authority:		prototype.DefaultAuthority,
		PoolID:			pool.PoolID,
		PoolOperator:		pool.PoolOperator,
		ValidatorTarget:	nextValidator,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
		Height:			15,
	})
	require.NoError(t, err)
	require.Equal(t, nextValidator, finalized.ValidatorTarget)
	require.Empty(t, finalized.PendingValidatorTarget)
}

func TestPoolCannotDelegateToJailedValidator(t *testing.T) {
	k := NewKeeper()
	_, err := k.CreateNominatorPool(types.MsgCreateNominatorPool{
		Authority:		prototype.DefaultAuthority,
		PoolID:			"pool-jailed",
		PoolOperator:		rawPoolAddress("11"),
		ValidatorTarget:	rawPoolAddress("12"),
		PoolCommissionBps:	100,
		Height:			1,
		ValidatorStatus:	validatorregistrytypes.StatusJailed,
	})
	require.ErrorContains(t, err, "jailed validator")
}

func createPool(t *testing.T, k *Keeper, poolID string) types.NominatorPool {
	t.Helper()
	return createPoolWithCommission(t, k, poolID, 100)
}

func createPoolWithCommission(t *testing.T, k *Keeper, poolID string, commissionBps uint32) types.NominatorPool {
	t.Helper()
	pool, err := k.CreateNominatorPool(types.MsgCreateNominatorPool{
		Authority:		prototype.DefaultAuthority,
		PoolID:			poolID,
		PoolOperator:		rawPoolAddress("11"),
		ValidatorTarget:	rawPoolAddress("12"),
		PoolCommissionBps:	commissionBps,
		Height:			1,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	})
	require.NoError(t, err)
	return pool
}

func createOfficialLiquidStakingPool(t *testing.T, k *Keeper, poolID string) types.NominatorPool {
	t.Helper()
	contractRaw := rawPoolAddress("66")
	pool, err := k.CreateOfficialLiquidStakingPool(types.MsgCreateOfficialLiquidStakingPool{
		Authority:		prototype.DefaultAuthority,
		PoolID:			poolID,
		ContractAddressUser:	aeFromRaw(t, contractRaw),
		ContractAddressRaw:	contractRaw,
		PoolOperator:		rawPoolAddress("11"),
		PoolCommissionBps:	100,
		Height:			1,
	})
	require.NoError(t, err)
	return pool
}

func deposit(t *testing.T, k *Keeper, poolID string, delegator string, amount uint64, height uint64) {
	t.Helper()
	_, err := k.DepositToPool(types.MsgDepositToPool{
		Authority:	prototype.DefaultAuthority,
		PoolID:		poolID,
		Delegator:	delegator,
		Amount:		amount,
		Height:		height,
	})
	require.NoError(t, err)
}

func rawPoolAddress(hexByte string) string {
	return "4:000000000000000000000000" + fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte)
}

func rawPoolAddressFromInt(value int) string {
	return fmt.Sprintf("4:%064x", value)
}

func proofUserAddress(hexByte string) string {
	bz, err := addressing.Parse(rawPoolAddress(hexByte))
	if err != nil {
		panic(err)
	}
	text, err := addressing.FormatUserFriendly(bz)
	if err != nil {
		panic(err)
	}
	return text
}

func aePoolAddress(t *testing.T, hexByte string) string {
	t.Helper()
	return aeFromRaw(t, rawPoolAddress(hexByte))
}

func aeFromRaw(t *testing.T, raw string) string {
	t.Helper()
	bz, err := addressing.Parse(raw)
	require.NoError(t, err)
	user, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return user
}
