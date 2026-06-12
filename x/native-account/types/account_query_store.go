package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/x/internal/prototype"
)

type AccountQueryStore interface {
	AccountByUser(userAddress string) (Account, bool, error)
	AccountByRaw(rawAddress string) (Account, bool, error)
	AccountByReputation(reputationID string) (Account, bool, error)
	AccountsAfter(cursor string, limit uint64) ([]Account, bool, error)
}

type indexedAccountStore struct {
	byUser	map[string]Account
	byRaw	map[string]Account
	byRep	map[string]Account
	sorted	[]Account
}

func NewIndexedAccountStore(accounts ...Account) (AccountQueryStore, error) {
	store := &indexedAccountStore{
		byUser:	make(map[string]Account, len(accounts)),
		byRaw:	make(map[string]Account, len(accounts)),
		byRep:	make(map[string]Account),
	}
	for _, account := range accounts {
		if err := ValidateAccountInvariant(account); err != nil {
			return nil, fmt.Errorf("invalid account %s: %w", account.AddressUser, err)
		}
		store.byUser[account.AddressUser] = account
		store.byRaw[account.AddressRaw] = account
		if account.ReputationID != "" {
			store.byRep[account.ReputationID] = account
		}
	}
	store.sorted = SortAccounts(accounts)
	return store, nil
}

func (s *indexedAccountStore) AccountByUser(userAddress string) (Account, bool, error) {
	account, ok := s.byUser[userAddress]
	if !ok {
		return Account{}, false, nil
	}
	return account, true, nil
}

func (s *indexedAccountStore) AccountByRaw(rawAddress string) (Account, bool, error) {
	account, ok := s.byRaw[rawAddress]
	if !ok {
		return Account{}, false, nil
	}
	return account, true, nil
}

func (s *indexedAccountStore) AccountByReputation(reputationID string) (Account, bool, error) {
	account, ok := s.byRep[reputationID]
	if !ok {
		return Account{}, false, nil
	}
	return account, true, nil
}

func (s *indexedAccountStore) AccountsAfter(cursor string, limit uint64) ([]Account, bool, error) {
	start := 0
	if cursor != "" {
		pos := sort.Search(len(s.sorted), func(i int) bool {
			return s.sorted[i].AddressUser > cursor
		})
		start = pos
	}
	if start >= len(s.sorted) {
		return nil, false, nil
	}
	end := start + int(limit)
	if end > len(s.sorted) {
		end = len(s.sorted)
	}
	hasMore := end < len(s.sorted)
	out := make([]Account, end-start)
	copy(out, s.sorted[start:end])
	return out, hasMore, nil
}

type AccountsResult struct {
	Accounts	[]Account
	NextCursor	string
	Limit		uint64
}

type AccountQueryService interface {
	Account(userAddress string) (Account, bool, error)
	AccountByRawAddress(rawAddress string) (Account, bool, error)
	AccountReputation(reputationID string) (Account, bool, error)
	Accounts(cursor string, limit uint64) (AccountsResult, error)
}

type accountQueryService struct {
	store	AccountQueryStore
	params	prototype.Params
}

func NewAccountQueryService(store AccountQueryStore, params prototype.Params) (AccountQueryService, error) {
	if store == nil {
		return nil, fmt.Errorf("native account query store is required")
	}
	return &accountQueryService{store: store, params: params}, nil
}

func (s *accountQueryService) Account(userAddress string) (Account, bool, error) {
	if err := ValidateUserFacingAEAddress("user_address", userAddress); err != nil {
		return Account{}, false, err
	}
	return s.store.AccountByUser(userAddress)
}

func (s *accountQueryService) AccountByRawAddress(rawAddress string) (Account, bool, error) {
	if err := ValidateRawAddress("raw_address", rawAddress); err != nil {
		return Account{}, false, err
	}
	return s.store.AccountByRaw(rawAddress)
}

func (s *accountQueryService) AccountReputation(reputationID string) (Account, bool, error) {
	reputationID = strings.TrimSpace(reputationID)
	if err := validateReputationID(reputationID); err != nil {
		return Account{}, false, err
	}
	if reputationID == "" {
		return Account{}, false, fmt.Errorf("native account reputation id is required")
	}
	return s.store.AccountByReputation(reputationID)
}

func (s *accountQueryService) Accounts(cursor string, limit uint64) (AccountsResult, error) {
	if limit == 0 || limit > s.params.MaxQueryLimit {
		return AccountsResult{}, fmt.Errorf("native account query limit must be between 1 and %d", s.params.MaxQueryLimit)
	}
	accounts, hasMore, err := s.store.AccountsAfter(cursor, limit)
	if err != nil {
		return AccountsResult{}, err
	}
	result := AccountsResult{
		Accounts:	accounts,
		Limit:		limit,
	}
	if len(accounts) > 0 {
		result.NextCursor = accounts[len(accounts)-1].AddressUser
	}
	if !hasMore {
		result.NextCursor = ""
	}
	return result, nil
}
