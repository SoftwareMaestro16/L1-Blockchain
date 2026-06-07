package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

type AccountQueryStore interface {
	AccountByUser(userAddress string) (Account, bool, error)
	AccountByRaw(rawAddress string) (Account, bool, error)
	AccountByReputation(reputationID string) (Account, bool, error)
	AccountsAfter(cursor string, limit uint64) ([]Account, bool, error)
}

type AccountQueryService struct {
	store  AccountQueryStore
	params prototype.Params
}

type AccountsPage struct {
	Accounts   []Account
	NextCursor string
	Limit      uint64
}

func NewAccountQueryService(store AccountQueryStore, params prototype.Params) (AccountQueryService, error) {
	if store == nil {
		return AccountQueryService{}, errors.New("native account query store is required")
	}
	if err := params.Validate(); err != nil {
		return AccountQueryService{}, err
	}
	return AccountQueryService{store: store, params: params}, nil
}

func (q AccountQueryService) Account(userAddress string) (Account, bool, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return Account{}, false, err
	}
	return q.store.AccountByUser(pair.User)
}

func (q AccountQueryService) AccountByRawAddress(rawAddress string) (Account, bool, error) {
	pair, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, rawAddress)
	if err != nil {
		return Account{}, false, err
	}
	return q.store.AccountByRaw(pair.Raw)
}

func (q AccountQueryService) AccountReputation(reputationID string) (Account, bool, error) {
	reputationID = strings.TrimSpace(reputationID)
	if _, err := AccountByReputationKey(reputationID); err != nil {
		return Account{}, false, err
	}
	return q.store.AccountByReputation(reputationID)
}

func (q AccountQueryService) Accounts(cursor string, limit uint64) (AccountsPage, error) {
	limit, err := normalizeAccountQueryLimit(limit, q.params)
	if err != nil {
		return AccountsPage{}, err
	}
	cursor = strings.TrimSpace(cursor)
	if cursor != "" {
		pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, cursor)
		if err != nil {
			return AccountsPage{}, err
		}
		cursor = pair.User
	}
	accounts, hasMore, err := q.store.AccountsAfter(cursor, limit)
	if err != nil {
		return AccountsPage{}, err
	}
	accounts = SortAccounts(accounts)
	page := AccountsPage{
		Accounts: accounts,
		Limit:    limit,
	}
	if hasMore && len(accounts) > 0 {
		page.NextCursor = accounts[len(accounts)-1].AddressUser
	}
	return page, nil
}

type IndexedAccountStore struct {
	byUser       map[string]Account
	byRaw        map[string]string
	byNumber     map[uint64]string
	byReputation map[string]string
	sortedUsers  []string
}

func NewIndexedAccountStore(accounts ...Account) (*IndexedAccountStore, error) {
	store := &IndexedAccountStore{}
	for _, account := range accounts {
		if err := store.SetAccount(account); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *IndexedAccountStore) SetAccount(account Account) error {
	if s == nil {
		return errors.New("native account indexed store is required")
	}
	if err := ValidateAccountInvariant(account); err != nil {
		return err
	}
	account = normalizeAccount(account)
	if s.byUser == nil {
		s.byUser = map[string]Account{}
		s.byRaw = map[string]string{}
		s.byNumber = map[uint64]string{}
		s.byReputation = map[string]string{}
	}
	if existing, found := s.byRaw[account.AddressRaw]; found && existing != account.AddressUser {
		return fmt.Errorf("duplicate native account raw address %s", account.AddressRaw)
	}
	if existing, found := s.byNumber[account.AccountNumber]; found && existing != account.AddressUser {
		return fmt.Errorf("duplicate native account number %d", account.AccountNumber)
	}
	reputationID := strings.TrimSpace(account.ReputationID)
	if reputationID != "" {
		if existing, found := s.byReputation[reputationID]; found && existing != account.AddressUser {
			return fmt.Errorf("duplicate native account reputation id %s", reputationID)
		}
	}
	s.byUser[account.AddressUser] = account
	s.byRaw[account.AddressRaw] = account.AddressUser
	s.byNumber[account.AccountNumber] = account.AddressUser
	if reputationID != "" {
		s.byReputation[reputationID] = account.AddressUser
	}
	s.rebuildSortedUsers()
	return nil
}

func (s *IndexedAccountStore) AccountByUser(userAddress string) (Account, bool, error) {
	key, err := AccountByUserKey(userAddress)
	if err != nil {
		return Account{}, false, err
	}
	user := strings.TrimPrefix(key, AccountByUserPrefix)
	account, found := s.byUser[user]
	return cloneAccount(account), found, nil
}

func (s *IndexedAccountStore) AccountByRaw(rawAddress string) (Account, bool, error) {
	key, err := AccountByRawKey(rawAddress)
	if err != nil {
		return Account{}, false, err
	}
	raw := strings.TrimPrefix(key, AccountByRawPrefix)
	user, found := s.byRaw[raw]
	if !found {
		return Account{}, false, nil
	}
	account, found := s.byUser[user]
	return cloneAccount(account), found, nil
}

func (s *IndexedAccountStore) AccountByReputation(reputationID string) (Account, bool, error) {
	if _, err := AccountByReputationKey(reputationID); err != nil {
		return Account{}, false, err
	}
	user, found := s.byReputation[strings.TrimSpace(reputationID)]
	if !found {
		return Account{}, false, nil
	}
	account, found := s.byUser[user]
	return cloneAccount(account), found, nil
}

func (s *IndexedAccountStore) AccountsAfter(cursor string, limit uint64) ([]Account, bool, error) {
	if limit == 0 || limit > prototype.MaxQueryLimit {
		return nil, false, fmt.Errorf("native account query limit must be between 1 and %d", prototype.MaxQueryLimit)
	}
	cursor = strings.TrimSpace(cursor)
	if cursor != "" {
		pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, cursor)
		if err != nil {
			return nil, false, err
		}
		cursor = pair.User
	}
	start := 0
	for start < len(s.sortedUsers) && s.sortedUsers[start] <= cursor {
		start++
	}
	if start >= len(s.sortedUsers) {
		return nil, false, nil
	}
	end := start + int(limit)
	if end > len(s.sortedUsers) {
		end = len(s.sortedUsers)
	}
	accounts := make([]Account, 0, end-start)
	for _, user := range s.sortedUsers[start:end] {
		accounts = append(accounts, cloneAccount(s.byUser[user]))
	}
	return accounts, end < len(s.sortedUsers), nil
}

func (s *IndexedAccountStore) NextAccountNumber() uint64 {
	if s == nil || len(s.byNumber) == 0 {
		return 1
	}
	var max uint64
	for accountNumber := range s.byNumber {
		if accountNumber > max {
			max = accountNumber
		}
	}
	return max + 1
}

func (s *IndexedAccountStore) rebuildSortedUsers() {
	s.sortedUsers = s.sortedUsers[:0]
	for user := range s.byUser {
		s.sortedUsers = append(s.sortedUsers, user)
	}
	sort.Strings(s.sortedUsers)
}

func normalizeAccountQueryLimit(limit uint64, params prototype.Params) (uint64, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if limit == 0 {
		limit = params.DefaultQueryLimit
	}
	if limit == 0 || limit > params.MaxQueryLimit || limit > prototype.MaxQueryLimit {
		return 0, fmt.Errorf("native account query limit must be between 1 and %d", params.MaxQueryLimit)
	}
	return limit, nil
}
