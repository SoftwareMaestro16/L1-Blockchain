package types

import (
	"errors"
	"fmt"
	"math"

	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
)

const (
	WalletActionNormalSpend		= "normal_spend"
	WalletActionContractExecute	= "contract_execute"
	WalletActionStakingDeposit	= "staking_deposit"
	WalletActionReadQuery		= "read_query"
	WalletActionProofQuery		= "proof_query"
	WalletActionTopUp		= "top_up"
	WalletActionPayStorageDebt	= "pay_storage_debt"
	WalletActionUnfreeze		= "unfreeze"
)

type WalletStorageUsage struct {
	CodeBytes	uint64
	DataBytes	uint64
	IndexBytes	uint64
}

type WalletRentInput struct {
	Account			Account
	Balance			uint64
	Storage			WalletStorageUsage
	GasFee			uint64
	CurrentHeight		uint64
	Params			storagerenttypes.StorageRentParams
	FreezeDebtThreshold	uint64
}

type WalletRentResult struct {
	Account			Account
	Balance			uint64
	StorageBytes		uint64
	StorageRentDelta	uint64
	EffectiveFee		uint64
	Frozen			bool
	DebtPaid		bool
}

type AccountCloseObligations struct {
	Balance			uint64
	StorageDebt		uint64
	Stake			uint64
	PoolShares		uint64
	Unbonding		uint64
	PendingRewards		uint64
	Domains			uint64
	OwnershipObligations	uint64
	RequiredReputation	uint64
}

func WalletPersistentStorageSize(usage WalletStorageUsage) (uint64, error) {
	total := usage.CodeBytes
	if total > math.MaxUint64-usage.DataBytes {
		return 0, errors.New("native wallet storage size overflow")
	}
	return total + usage.DataBytes, nil
}

func CollectWalletStorageRent(input WalletRentInput) (WalletRentResult, error) {
	account := cloneAccount(input.Account)
	result := WalletRentResult{Account: account, Balance: input.Balance}
	if account.Status == AccountStatusInactive || account.Status == AccountStatusClosed || account.Status == AccountStatusArchived {
		return result, nil
	}
	if input.CurrentHeight == 0 || input.CurrentHeight < account.LastStorageChargeHeight {
		return WalletRentResult{}, errors.New("native wallet storage rent height must be monotonic")
	}
	if err := input.Params.Validate(); err != nil {
		return WalletRentResult{}, err
	}
	size, err := WalletPersistentStorageSize(input.Storage)
	if err != nil {
		return WalletRentResult{}, err
	}
	result.StorageBytes = size
	if size == 0 {
		result.EffectiveFee = input.GasFee
		return result, nil
	}
	lastHeight := account.LastStorageChargeHeight
	if lastHeight == 0 {
		lastHeight = account.CreatedHeight
	}
	if input.CurrentHeight < lastHeight {
		return WalletRentResult{}, errors.New("native wallet storage rent height must be monotonic")
	}
	delta, err := storagerenttypes.RentForBlocks(size, input.Params, input.CurrentHeight-lastHeight)
	if err != nil {
		return WalletRentResult{}, err
	}
	effectiveFee, err := walletEffectiveFee(input.GasFee, delta, account.StorageRentDebt)
	if err != nil {
		return WalletRentResult{}, err
	}
	result.StorageRentDelta = delta
	result.EffectiveFee = effectiveFee
	account.LastStorageChargeHeight = input.CurrentHeight
	if input.Balance >= effectiveFee {
		result.Balance = input.Balance - effectiveFee
		account.StorageRentDebt = 0
		result.DebtPaid = true
		result.Account = account
		return result, ValidateAccountInvariant(account)
	}
	if account.StorageRentDebt > math.MaxUint64-delta {
		return WalletRentResult{}, errors.New("native wallet storage rent debt overflow")
	}
	account.StorageRentDebt += delta
	if account.StorageRentDebt > input.FreezeDebtThreshold || input.Balance < effectiveFee {
		account.Status = AccountStatusFrozen
		result.Frozen = true
	}
	result.Account = account
	return result, ValidateAccountInvariant(account)
}

func CanWalletPerformAction(account Account, action string) bool {
	switch account.Status {
	case AccountStatusActive, AccountStatusRecovered:
		return true
	case AccountStatusFrozen:
		switch action {
		case WalletActionReadQuery, WalletActionProofQuery, WalletActionTopUp, WalletActionPayStorageDebt, WalletActionUnfreeze:
			return true
		default:
			return false
		}
	case AccountStatusArchived:
		return action == WalletActionReadQuery || action == WalletActionProofQuery
	default:
		return false
	}
}

func ApplyWalletTopUp(balance, amount uint64) (uint64, error) {
	if balance > math.MaxUint64-amount {
		return 0, errors.New("native wallet top-up balance overflow")
	}
	return balance + amount, nil
}

func walletEffectiveFee(gasFee, rentDelta, unpaidDebt uint64) (uint64, error) {
	if gasFee > math.MaxUint64-rentDelta {
		return 0, errors.New("native wallet effective fee overflow")
	}
	total := gasFee + rentDelta
	if total > math.MaxUint64-unpaidDebt {
		return 0, errors.New("native wallet effective fee overflow")
	}
	return total + unpaidDebt, nil
}

func CanArchiveAccount(account Account, obligations AccountCloseObligations) error {
	if account.Status == AccountStatusClosed {
		return errors.New("closed account cannot be archived")
	}
	if obligations.HasAny() {
		return fmt.Errorf("native account %s cannot archive with outstanding obligations", account.AddressUser)
	}
	return nil
}

func CanCloseAccount(account Account, obligations AccountCloseObligations) error {
	if account.Status == AccountStatusInactive {
		return errors.New("inactive account is virtual and has no close flow")
	}
	if obligations.HasAny() {
		return fmt.Errorf("native account %s cannot close with outstanding obligations", account.AddressUser)
	}
	return nil
}

func (o AccountCloseObligations) HasAny() bool {
	return o.Balance != 0 ||
		o.StorageDebt != 0 ||
		o.Stake != 0 ||
		o.PoolShares != 0 ||
		o.Unbonding != 0 ||
		o.PendingRewards != 0 ||
		o.Domains != 0 ||
		o.OwnershipObligations != 0 ||
		o.RequiredReputation != 0
}
