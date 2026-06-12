package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestPhase31UserDepositToPoolWithoutValidatorAtTenAET(t *testing.T) {
	user := aePoolAddress(t, "61")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "phase31-official")

	receipt, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		10 * types.DefaultAETBaseUnits,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, pool.PoolID, receipt.PoolID)
	require.Equal(t, user, receipt.OwnerAddress)
	require.Equal(t, 10*types.DefaultAETBaseUnits, receipt.Amount)
	require.Equal(t, 10*types.DefaultAETBaseUnits, receipt.Shares)

	stored, found := k.NominatorPool(pool.PoolID)
	require.True(t, found)
	require.Equal(t, 10*types.DefaultAETBaseUnits, stored.TotalBondedStake)
	require.Equal(t, 10*types.DefaultAETBaseUnits, stored.TotalShares)
	require.Empty(t, stored.ValidatorTarget)
	require.Empty(t, stored.PendingValidatorTarget)
}

func TestPhase31DepositCanTargetOfficialContractAndRejectsHiddenValidatorTargets(t *testing.T) {
	user := aePoolAddress(t, "62")
	validator := aePoolAddress(t, "63")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "phase31-contract")

	receipt, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		OfficialContract:	pool.ContractAddressUser,
		WalletAddress:		user,
		Amount:			types.DefaultMinPoolDeposit,
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, pool.PoolID, receipt.PoolID)
	require.Equal(t, pool.ContractAddressUser, receipt.PoolContractAddressUser)

	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		validator,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit,
		Height:		3,
	})
	require.ErrorContains(t, err, "pool id must not be an address")

	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:			"wrong-pool",
		OfficialContract:	pool.ContractAddressUser,
		WalletAddress:		user,
		Amount:			types.DefaultMinPoolDeposit,
		Height:			3,
	})
	require.ErrorContains(t, err, "pool id does not match official contract")

	_, err = k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		types.DefaultMinPoolDeposit - 1,
		Height:		3,
	})
	require.ErrorContains(t, err, "below configured minimum")
}

func TestPhase31PoolSharesDeterministicAndExportImportPreservesTotals(t *testing.T) {
	firstReceipt, firstExport := phase31DepositRun(t)
	secondReceipt, secondExport := phase31DepositRun(t)
	require.Equal(t, firstReceipt.Shares, secondReceipt.Shares)
	require.Equal(t, firstExport.State.Pools[0].TotalShares, secondExport.State.Pools[0].TotalShares)
	require.Equal(t, firstExport.State.Pools[0].TotalBondedStake, secondExport.State.Pools[0].TotalBondedStake)

	imported := NewKeeperWithAccountStatus(accountStatusFixture{firstReceipt.OwnerAddress: accountStatusActive})
	require.NoError(t, imported.InitGenesis(firstExport))
	roundTrip := imported.ExportGenesis()
	require.Equal(t, firstExport.State.Pools[0].TotalShares, roundTrip.State.Pools[0].TotalShares)
	require.Equal(t, firstExport.State.Pools[0].TotalBondedStake, roundTrip.State.Pools[0].TotalBondedStake)
	require.Equal(t, firstExport.State.PoolShares[0].Shares, roundTrip.State.PoolShares[0].Shares)
	require.Equal(t, firstExport.State.LiquidStakingPools[0].TotalShares, roundTrip.State.LiquidStakingPools[0].TotalShares)
}

func TestPhase31DirectUserDelegationRejectedBeforeNominatorPoolMutation(t *testing.T) {
	user := aePoolAddress(t, "66")
	validator := aePoolAddress(t, "67")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive, validator: accountStatusActive})
	createOfficialLiquidStakingPool(t, &k, "phase31-no-direct")
	before := k.ExportGenesis()

	err := k.DelegateUserToValidator(types.MsgDelegateToValidator{
		Authority:		prototype.DefaultAuthority,
		UserAddress:		user,
		ValidatorAddress:	validator,
		Amount:			types.DefaultMinPoolDeposit,
		Height:			2,
	})
	require.ErrorContains(t, err, "direct user delegation to validators is disabled")
	require.Equal(t, before, k.ExportGenesis())
}

func phase31DepositRun(t *testing.T) (types.StakingPoolDepositReceipt, GenesisState) {
	t.Helper()
	user := aePoolAddress(t, "65")
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, "phase31-deterministic")
	receipt, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		10 * types.DefaultAETBaseUnits,
		Height:		2,
	})
	require.NoError(t, err)
	exported := k.ExportGenesis()
	require.NoError(t, exported.Validate())
	return receipt, exported
}
