package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	VirtualAccountStatusInactive	= "inactive"
	VirtualAccountStatusActive	= "active"
	VirtualAccountStatusFrozen	= "frozen"
	VirtualAccountStatusRecovered	= "recovered"
	VirtualAccountStatusArchived	= "archived"
	VirtualAccountStatusClosed	= "closed"

	AccountMessageActivate	= "activate_account"
	AccountMessageNormal	= "normal_account_message"

	PersistentWriteReasonActivation			= "activation"
	PersistentWriteReasonControlledMigration	= "controlled_migration"
)

type VirtualAccountView struct {
	AddressUser		string
	AddressRaw		string
	Status			string
	Persistent		bool
	StorageRentActive	bool
}

type VirtualAccountBook struct {
	accounts map[string]VirtualAccountView
}

func NewVirtualAccountBook(accounts ...VirtualAccountView) (*VirtualAccountBook, error) {
	book := &VirtualAccountBook{accounts: map[string]VirtualAccountView{}}
	for _, account := range accounts {
		if err := book.PutPersistentAccount(account, PersistentWriteReasonControlledMigration); err != nil {
			return nil, err
		}
	}
	return book, nil
}

func (b *VirtualAccountBook) QueryAccount(userAddress string) (VirtualAccountView, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return VirtualAccountView{}, err
	}
	if b != nil && b.accounts != nil {
		if account, found := b.accounts[pair.User]; found {
			return account, nil
		}
	}
	return VirtualAccountView{
		AddressUser:		pair.User,
		AddressRaw:		pair.Raw,
		Status:			VirtualAccountStatusInactive,
		Persistent:		false,
		StorageRentActive:	false,
	}, nil
}

func (b *VirtualAccountBook) PutPersistentAccount(account VirtualAccountView, reason string) error {
	if b == nil {
		return errors.New("native account book is required")
	}
	if reason != PersistentWriteReasonActivation && reason != PersistentWriteReasonControlledMigration {
		return fmt.Errorf("persistent native account writes require activation or controlled migration")
	}
	account, err := NormalizePersistentAccountView(account)
	if err != nil {
		return err
	}
	if b.accounts == nil {
		b.accounts = map[string]VirtualAccountView{}
	}
	b.accounts[account.AddressUser] = account
	return nil
}

func (b *VirtualAccountBook) ExportGenesisAccounts() []VirtualAccountView {
	if b == nil || len(b.accounts) == 0 {
		return nil
	}
	accounts := make([]VirtualAccountView, 0, len(b.accounts))
	for _, account := range b.accounts {
		if account.Persistent {
			accounts = append(accounts, account)
		}
	}
	sort.SliceStable(accounts, func(i, j int) bool {
		return accounts[i].AddressUser < accounts[j].AddressUser
	})
	return accounts
}

func NormalizePersistentAccountView(account VirtualAccountView) (VirtualAccountView, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, account.AddressUser)
	if err != nil {
		return VirtualAccountView{}, err
	}
	if strings.TrimSpace(account.AddressRaw) == "" {
		account.AddressRaw = pair.Raw
	}
	rawPair, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, account.AddressRaw)
	if err != nil {
		return VirtualAccountView{}, err
	}
	if rawPair.User != pair.User {
		return VirtualAccountView{}, errors.New("persistent native account AE and raw addresses must match")
	}
	if account.Status == VirtualAccountStatusInactive {
		return VirtualAccountView{}, errors.New("inactive account is virtual only and must not be persisted")
	}
	if !isPersistentVirtualStatus(account.Status) {
		return VirtualAccountView{}, fmt.Errorf("unsupported persistent native account status %q", account.Status)
	}
	account.AddressUser = pair.User
	account.AddressRaw = pair.Raw
	account.Persistent = true
	account.StorageRentActive = StorageRentAccruesForAccount(account)
	return account, nil
}

func StorageRentAccruesForAccount(account VirtualAccountView) bool {
	return account.Persistent && account.Status != VirtualAccountStatusInactive && account.Status != VirtualAccountStatusClosed
}

func ValidateAccountMessage(account VirtualAccountView, messageKind string) error {
	if strings.TrimSpace(messageKind) == "" {
		return errors.New("account message kind is required")
	}
	if account.Status == VirtualAccountStatusInactive && messageKind != AccountMessageActivate {
		return errors.New("inactive account can only send MsgActivateAccount")
	}
	if account.Status == VirtualAccountStatusFrozen && messageKind == AccountMessageNormal {
		return errors.New("frozen account cannot send normal account messages")
	}
	return nil
}

func isPersistentVirtualStatus(status string) bool {
	return status == VirtualAccountStatusActive ||
		status == VirtualAccountStatusFrozen ||
		status == VirtualAccountStatusRecovered ||
		status == VirtualAccountStatusArchived ||
		status == VirtualAccountStatusClosed
}
