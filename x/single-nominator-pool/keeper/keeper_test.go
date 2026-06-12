package keeper

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/single-nominator-pool/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestSingleNominatorPoolUnbondingUsesStakingPolicyWindow(t *testing.T) {
	gs := DefaultGenesis()

	require.Equal(t, appparams.StakingUnbondingDefaultBlocks, gs.Params.UnbondingBlocks)
	require.NoError(t, gs.Validate())

	gs.Params.UnbondingBlocks = appparams.StakingUnbondingMaxBlocks + 1
	require.ErrorContains(t, gs.Validate(), "14-21 days")
}

func TestCreatePool(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))

	require.Equal(t, rawSingleAddress("10"), pool.PoolAddress)
	require.Equal(t, rawSingleAddress("11"), pool.Owner)
	require.Equal(t, rawSingleAddress("12"), pool.Validator)
	require.Equal(t, types.StatusActive, pool.Status)
}

func TestDepositAndDelegate(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))

	deposited, err := k.DepositSingleNominator(types.MsgDepositSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		1_000,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1_000), deposited.BondedStake)
	require.Equal(t, pool.Validator, deposited.Validator)
}

func TestDepositOverflowRejected(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))
	depositSingle(t, &k, pool, math.MaxUint64, 2)

	_, err := k.DepositSingleNominator(types.MsgDepositSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		1,
		Height:		3,
	})
	require.ErrorContains(t, err, "overflow bonded stake")
}

func TestOwnerOnlyWithdrawal(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))
	depositSingle(t, &k, pool, 1_000, 2)

	_, err := k.WithdrawSingleNominator(types.MsgWithdrawSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		rawSingleAddress("33"),
		Amount:		100,
		Height:		3,
	})
	require.ErrorContains(t, err, "only single nominator owner")

	withdrawal, err := k.WithdrawSingleNominator(types.MsgWithdrawSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		400,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, types.WithdrawalStatusPending, withdrawal.Status)
	require.Equal(t, uint64(400), withdrawal.Amount)
	stored, found := k.SingleNominatorPool(pool.PoolAddress)
	require.True(t, found)
	require.Equal(t, uint64(600), stored.BondedStake)
}

func TestEmergencyLockBehavior(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))
	depositSingle(t, &k, pool, 1_000, 2)

	locked, err := k.EmergencyLockSingleNominator(types.MsgEmergencyLockSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Locked:		true,
		Height:		3,
	})
	require.NoError(t, err)
	require.True(t, locked.EmergencyLock)

	_, err = k.WithdrawSingleNominator(types.MsgWithdrawSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		100,
		Height:		4,
	})
	require.ErrorContains(t, err, "emergency lock blocks withdrawals")

	slashed, err := k.ApplySingleNominatorSlash(pool.PoolAddress, 250)
	require.NoError(t, err)
	require.Equal(t, uint64(750), slashed.BondedStake)

	_, err = k.EmergencyLockSingleNominator(types.MsgEmergencyLockSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Locked:		false,
		Height:		5,
	})
	require.NoError(t, err)
	withdrawal, err := k.WithdrawSingleNominator(types.MsgWithdrawSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		100,
		Height:		6,
	})
	require.NoError(t, err)
	require.Equal(t, types.WithdrawalStatusPending, withdrawal.Status)
}

func TestSlashBehavior(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))
	depositSingle(t, &k, pool, 1_000, 2)

	rewarded, err := k.ApplySingleNominatorReward(pool.PoolAddress, 125)
	require.NoError(t, err)
	require.Equal(t, uint64(125), rewarded.RewardBalance)
	rewards, found := k.SingleNominatorRewards(pool.PoolAddress)
	require.True(t, found)
	require.Equal(t, uint64(125), rewards)
	claimed, err := k.ClaimSingleNominatorRewards(types.MsgClaimSingleNominatorRewards{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(125), claimed)

	slashed, err := k.ApplySingleNominatorSlash(pool.PoolAddress, 300)
	require.NoError(t, err)
	require.Equal(t, uint64(700), slashed.BondedStake)
	slashed, err = k.ApplySingleNominatorSlash(pool.PoolAddress, 1_000)
	require.NoError(t, err)
	require.Equal(t, uint64(0), slashed.BondedStake)
}

func TestRewardOverflowRejected(t *testing.T) {
	k := NewKeeper()
	pool := createSinglePool(t, &k, rawSingleAddress("10"))
	_, err := k.ApplySingleNominatorReward(pool.PoolAddress, math.MaxUint64)
	require.NoError(t, err)

	_, err = k.ApplySingleNominatorReward(pool.PoolAddress, 1)
	require.ErrorContains(t, err, "overflow balance")
}

func TestExportImportDuringPendingWithdrawal(t *testing.T) {
	source := NewKeeper()
	pool := createSinglePool(t, &source, rawSingleAddress("10"))
	depositSingle(t, &source, pool, 1_000, 2)
	withdrawal, err := source.WithdrawSingleNominator(types.MsgWithdrawSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		400,
		Height:		3,
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
	stored, found := target.SingleNominatorPool(pool.PoolAddress)
	require.True(t, found)
	require.Equal(t, withdrawal, stored.PendingWithdrawal)
}

func TestCannotDelegateToJailedValidator(t *testing.T) {
	k := NewKeeper()
	_, err := k.CreateSingleNominatorPool(types.MsgCreateSingleNominatorPool{
		Authority:		prototype.DefaultAuthority,
		PoolAddress:		rawSingleAddress("10"),
		Owner:			rawSingleAddress("11"),
		Validator:		rawSingleAddress("12"),
		ValidatorStatus:	validatorregistrytypes.StatusJailed,
		Height:			1,
	})
	require.ErrorContains(t, err, "jailed validator")

	pool := createSinglePool(t, &k, rawSingleAddress("20"))
	_, err = k.ChangeSingleNominatorValidator(types.MsgChangeSingleNominatorValidator{
		Authority:		prototype.DefaultAuthority,
		PoolAddress:		pool.PoolAddress,
		Owner:			pool.Owner,
		Validator:		rawSingleAddress("44"),
		ValidatorStatus:	validatorregistrytypes.StatusJailed,
		Height:			2,
	})
	require.ErrorContains(t, err, "jailed validator")
}

func createSinglePool(t *testing.T, k *Keeper, poolAddress string) types.SingleNominatorPool {
	t.Helper()
	pool, err := k.CreateSingleNominatorPool(types.MsgCreateSingleNominatorPool{
		Authority:		prototype.DefaultAuthority,
		PoolAddress:		poolAddress,
		Owner:			rawSingleAddress("11"),
		Validator:		rawSingleAddress("12"),
		ValidatorStatus:	validatorregistrytypes.StatusActive,
		Height:			1,
	})
	require.NoError(t, err)
	return pool
}

func depositSingle(t *testing.T, k *Keeper, pool types.SingleNominatorPool, amount uint64, height uint64) {
	t.Helper()
	_, err := k.DepositSingleNominator(types.MsgDepositSingleNominator{
		Authority:	prototype.DefaultAuthority,
		PoolAddress:	pool.PoolAddress,
		Owner:		pool.Owner,
		Amount:		amount,
		Height:		height,
	})
	require.NoError(t, err)
}

func rawSingleAddress(hexByte string) string {
	return "4:000000000000000000000000" + fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte)
}
