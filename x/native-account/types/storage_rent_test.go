package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
)

func TestWalletStorageRentAccruesLazilyAndEffectiveFeeIsDeterministic(t *testing.T) {
	account := completeActiveAccount(t, 0xd1, 1000, 3)
	account.StorageRentDebt = 5
	params := walletRentParams(2)

	result, err := CollectWalletStorageRent(WalletRentInput{
		Account:	account,
		Balance:	200,
		Storage:	WalletStorageUsage{CodeBytes: 10, DataBytes: 5},
		GasFee:		10,
		CurrentHeight:	15,
		Params:		params,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(15), result.StorageBytes)
	require.Equal(t, uint64(90), result.StorageRentDelta)
	require.Equal(t, uint64(105), result.EffectiveFee)
	require.Equal(t, uint64(95), result.Balance)
	require.Zero(t, result.Account.StorageRentDebt)
	require.Equal(t, uint64(15), result.Account.LastStorageChargeHeight)
	require.Equal(t, AccountStatusActive, result.Account.Status)
	require.True(t, result.DebtPaid)
}

func TestWalletStorageRentDoesNotApplyToInactiveEmptyClosedOrTransientState(t *testing.T) {
	params := walletRentParams(10)
	inactive := completeActiveAccount(t, 0xd2, 1001, 0)
	inactive.Status = AccountStatusInactive
	inactive.StorageRentDebt = 0

	result, err := CollectWalletStorageRent(WalletRentInput{
		Account:	inactive,
		Balance:	0,
		Storage:	WalletStorageUsage{CodeBytes: 100, DataBytes: 100},
		GasFee:		0,
		CurrentHeight:	100,
		Params:		params,
	})
	require.NoError(t, err)
	require.Zero(t, result.StorageRentDelta)
	require.Zero(t, result.Account.StorageRentDebt)

	empty := completeActiveAccount(t, 0xd3, 1002, 0)
	empty.StorageRentDebt = 0
	result, err = CollectWalletStorageRent(WalletRentInput{
		Account:	empty,
		Balance:	1,
		Storage:	WalletStorageUsage{},
		GasFee:		1,
		CurrentHeight:	100,
		Params:		params,
	})
	require.NoError(t, err)
	require.Zero(t, result.StorageRentDelta)
	require.Equal(t, uint64(1), result.EffectiveFee)
	require.Equal(t, AccountStatusActive, result.Account.Status)

	closed := completeActiveAccount(t, 0xd4, 1003, 0)
	closed.Status = AccountStatusClosed
	closed.StorageRentDebt = 7
	result, err = CollectWalletStorageRent(WalletRentInput{
		Account:	closed,
		Balance:	0,
		Storage:	WalletStorageUsage{DataBytes: 100},
		CurrentHeight:	100,
		Params:		params,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(7), result.Account.StorageRentDebt)
	require.Zero(t, result.StorageRentDelta)
}

func TestWalletInsufficientBalanceAccumulatesDebtFreezesAndPreservesState(t *testing.T) {
	account := completeActiveAccount(t, 0xd5, 1004, 9)
	account.StorageRentDebt = 0
	originalPolicy := account.AuthPolicy
	params := walletRentParams(1)

	result, err := CollectWalletStorageRent(WalletRentInput{
		Account:		account,
		Balance:		0,
		Storage:		WalletStorageUsage{CodeBytes: 3, DataBytes: 7, IndexBytes: 5},
		GasFee:			2,
		CurrentHeight:		14,
		Params:			params,
		FreezeDebtThreshold:	1,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(10), result.StorageBytes)
	require.Equal(t, uint64(20), result.StorageRentDelta)
	require.Equal(t, uint64(22), result.EffectiveFee)
	require.Equal(t, uint64(20), result.Account.StorageRentDebt)
	require.Equal(t, AccountStatusFrozen, result.Account.Status)
	require.True(t, result.Frozen)
	require.Equal(t, uint64(0), result.Balance)
	require.Equal(t, account.AddressUser, result.Account.AddressUser)
	require.Equal(t, account.AddressRaw, result.Account.AddressRaw)
	require.Equal(t, account.AccountNumber, result.Account.AccountNumber)
	require.Equal(t, account.Sequence, result.Account.Sequence)
	require.Equal(t, originalPolicy, result.Account.AuthPolicy)
	require.Equal(t, account.ReputationID, result.Account.ReputationID)
}

func TestFrozenWalletAllowsTopUpDebtPaymentAndUnfreezeButBlocksNormalActions(t *testing.T) {
	account := completeActiveAccount(t, 0xd6, 1005, 4)
	account.Status = AccountStatusFrozen
	account.StorageRentDebt = 5
	account.LastStorageChargeHeight = 20

	require.False(t, CanWalletPerformAction(account, WalletActionNormalSpend))
	require.False(t, CanWalletPerformAction(account, WalletActionStakingDeposit))
	require.False(t, CanWalletPerformAction(account, WalletActionContractExecute))
	require.True(t, CanWalletPerformAction(account, WalletActionReadQuery))
	require.True(t, CanWalletPerformAction(account, WalletActionProofQuery))
	require.True(t, CanWalletPerformAction(account, WalletActionTopUp))
	require.True(t, CanWalletPerformAction(account, WalletActionPayStorageDebt))
	require.True(t, CanWalletPerformAction(account, WalletActionUnfreeze))

	balance, err := ApplyWalletTopUp(0, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(10), balance)

	_, err = ApplyMsgUnfreezeAccount(account, MsgUnfreezeAccount{
		AccountUser:		account.AddressUser,
		Signers:		[]string{account.PubKeys[0]},
		CurrentHeight:		20,
		StorageDebtPaid:	true,
	})
	require.ErrorContains(t, err, "storage debt is paid")

	paid, err := ApplyMsgPayStorageDebt(account, MsgPayStorageDebt{
		AccountUser:	account.AddressUser,
		Amount:		5,
		Signers:	[]string{account.PubKeys[0]},
		CurrentHeight:	21,
	})
	require.NoError(t, err)
	require.Zero(t, paid.StorageRentDebt)
	require.Equal(t, AccountStatusFrozen, paid.Status)
	require.Equal(t, account.AddressUser, paid.AddressUser)
	require.Equal(t, account.AddressRaw, paid.AddressRaw)
	require.Equal(t, account.Sequence, paid.Sequence)

	unfrozen, err := ApplyMsgUnfreezeAccount(paid, MsgUnfreezeAccount{
		AccountUser:		paid.AddressUser,
		Signers:		[]string{paid.PubKeys[0]},
		CurrentHeight:		22,
		StorageDebtPaid:	true,
	})
	require.NoError(t, err)
	require.Equal(t, AccountStatusActive, unfrozen.Status)
	require.Zero(t, unfrozen.StorageRentDebt)
	require.Equal(t, paid.AddressUser, unfrozen.AddressUser)
	require.Equal(t, paid.AddressRaw, unfrozen.AddressRaw)
	require.Equal(t, paid.AccountNumber, unfrozen.AccountNumber)
	require.Equal(t, paid.Sequence, unfrozen.Sequence)
	require.Equal(t, paid.AuthPolicy, unfrozen.AuthPolicy)
}

func TestWalletCloseAndArchiveRequireZeroObligations(t *testing.T) {
	account := completeActiveAccount(t, 0xd7, 1006, 1)
	obligations := AccountCloseObligations{
		Balance:		1,
		StorageDebt:		2,
		Stake:			3,
		PoolShares:		4,
		Unbonding:		5,
		PendingRewards:		6,
		Domains:		7,
		OwnershipObligations:	8,
		RequiredReputation:	9,
	}

	require.ErrorContains(t, CanCloseAccount(account, obligations), "outstanding obligations")
	require.ErrorContains(t, CanArchiveAccount(account, obligations), "outstanding obligations")
	require.NoError(t, CanCloseAccount(account, AccountCloseObligations{}))
	require.NoError(t, CanArchiveAccount(account, AccountCloseObligations{}))
}

func TestWalletStorageRentExportImportPreservesDebtAndChargeHeight(t *testing.T) {
	account := completeActiveAccount(t, 0xd8, 1007, 2)
	account.Status = AccountStatusFrozen
	account.StorageRentDebt = 123
	account.LastStorageChargeHeight = 77
	source := newTestAccountStore(account)

	exported, err := ExportGenesis(source)
	require.NoError(t, err)
	target := newTestAccountStore()
	require.NoError(t, ImportGenesis(target, exported))
	roundTrip, err := ExportGenesis(target)
	require.NoError(t, err)

	require.Equal(t, exported, roundTrip)
	require.Equal(t, uint64(123), roundTrip.Accounts[0].StorageRentDebt)
	require.Equal(t, uint64(77), roundTrip.Accounts[0].LastStorageChargeHeight)
}

func walletRentParams(rate uint64) storagerenttypes.StorageRentParams {
	params := storagerenttypes.DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteBlock = rate
	return params
}
