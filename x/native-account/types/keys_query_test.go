package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	goldenUserAddress	= "AEAAAQAAAAAAAAAAAAAAAHUedugZkZbUVJQcRdGzoyPxQzvW"
	goldenRawAddress	= "4:000000000000000000000000751e76e8199196d454941c45d1b3a323f1433bd6"
)

func TestAccountStoreKeysAreDeterministicGoldenValues(t *testing.T) {
	byUser, err := AccountByUserKey(goldenUserAddress)
	require.NoError(t, err)
	require.Equal(t, "account/by_user/AEAAAQAAAAAAAAAAAAAAAHUedugZkZbUVJQcRdGzoyPxQzvW", byUser)

	byRaw, err := AccountByRawKey(goldenRawAddress)
	require.NoError(t, err)
	require.Equal(t, "account/by_raw/4:000000000000000000000000751e76e8199196d454941c45d1b3a323f1433bd6", byRaw)

	require.Equal(t, "account/number/00000000000000000042", AccountByNumberKey(42))

	byReputation, err := AccountByReputationKey("rep-0001")
	require.NoError(t, err)
	require.Equal(t, "account/reputation/rep-0001", byReputation)

	storage, err := AccountStorageKey(goldenUserAddress)
	require.NoError(t, err)
	require.Equal(t, "account/storage/AEAAAQAAAAAAAAAAAAAAAHUedugZkZbUVJQcRdGzoyPxQzvW", storage)
}

func TestAccountStoreKeyDescriptorsDocumentRequiredPrefixes(t *testing.T) {
	descriptors := DefaultAccountStoreKeyDescriptors()

	require.Equal(t, []AccountStoreKeyDescriptor{
		{Name: "account/by_user", Prefix: AccountByUserPrefix},
		{Name: "account/by_raw", Prefix: AccountByRawPrefix},
		{Name: "account/number", Prefix: AccountByNumberPrefix},
		{Name: "account/reputation", Prefix: AccountByReputationPrefix},
		{Name: "account/storage", Prefix: AccountStoragePrefix},
	}, descriptors)
}

func TestAccountQueryFindsActiveAccountByAEAddress(t *testing.T) {
	account := completeActiveAccount(t, 0x91, 1001, 1)
	store, err := NewIndexedAccountStore(account)
	require.NoError(t, err)
	metered := &meteredAccountQueryStore{AccountQueryStore: store}
	query := newTestAccountQueryService(t, metered)

	found, ok, err := query.Account(account.AddressUser)

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, account, found)
	require.Equal(t, 1, metered.byUser)
	require.Zero(t, metered.scans)
}

func TestAccountByRawAddressQueryReturnsSameAccount(t *testing.T) {
	account := completeActiveAccount(t, 0x92, 1002, 2)
	store, err := NewIndexedAccountStore(account)
	require.NoError(t, err)
	query := newTestAccountQueryService(t, store)

	byUser, ok, err := query.Account(account.AddressUser)
	require.NoError(t, err)
	require.True(t, ok)
	byRaw, ok, err := query.AccountByRawAddress(account.AddressRaw)
	require.NoError(t, err)
	require.True(t, ok)

	require.Equal(t, byUser, byRaw)
}

func TestAccountReputationQueryUsesSingleDeterministicLookup(t *testing.T) {
	account := completeActiveAccount(t, 0x93, 1003, 3)
	account.ReputationID = "rep-query-1"
	store, err := NewIndexedAccountStore(account)
	require.NoError(t, err)
	metered := &meteredAccountQueryStore{AccountQueryStore: store}
	query := newTestAccountQueryService(t, metered)

	found, ok, err := query.AccountReputation(account.ReputationID)

	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, account, found)
	require.Equal(t, 1, metered.byReputation)
	require.Zero(t, metered.scans)
}

func TestAccountsQueryIsPaginatedAndDeterministic(t *testing.T) {
	accounts := []Account{
		completeActiveAccount(t, 0x96, 1006, 6),
		completeActiveAccount(t, 0x94, 1004, 4),
		completeActiveAccount(t, 0x95, 1005, 5),
	}
	store, err := NewIndexedAccountStore(accounts...)
	require.NoError(t, err)
	metered := &meteredAccountQueryStore{AccountQueryStore: store}
	query := newTestAccountQueryService(t, metered)
	expected := SortAccounts(accounts)

	first, err := query.Accounts("", 2)
	require.NoError(t, err)
	require.Equal(t, expected[:2], first.Accounts)
	require.Equal(t, expected[1].AddressUser, first.NextCursor)
	require.Equal(t, uint64(2), first.Limit)

	second, err := query.Accounts(first.NextCursor, 2)
	require.NoError(t, err)
	require.Equal(t, expected[2:], second.Accounts)
	require.Empty(t, second.NextCursor)

	repeated, err := query.Accounts("", 2)
	require.NoError(t, err)
	require.Equal(t, first, repeated)
	require.Equal(t, 3, metered.scans)
}

func TestAccountQueriesRejectUnsupportedAddressFormatsSafely(t *testing.T) {
	account := completeActiveAccount(t, 0x97, 1007, 7)
	store, err := NewIndexedAccountStore(account)
	require.NoError(t, err)
	query := newTestAccountQueryService(t, store)

	_, _, err = query.Account(account.AddressRaw)
	require.Error(t, err)

	_, _, err = query.Account("cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a")
	require.Error(t, err)

	_, _, err = query.AccountByRawAddress(account.AddressUser)
	require.Error(t, err)

	_, _, err = query.AccountByRawAddress("1:0000000000000000000000000000000000000000")
	require.Error(t, err)

	_, err = query.Accounts("", prototype.MaxQueryLimit+1)
	require.ErrorContains(t, err, "query limit")

	_, _, err = query.AccountReputation("")
	require.ErrorContains(t, err, "reputation id is required")

	_, _, err = query.AccountReputation("seed_phrase:bad")
	require.ErrorContains(t, err, "seed phrases")
}

type meteredAccountQueryStore struct {
	AccountQueryStore
	byUser		int
	byRaw		int
	byReputation	int
	scans		int
}

func (s *meteredAccountQueryStore) AccountByUser(userAddress string) (Account, bool, error) {
	s.byUser++
	return s.AccountQueryStore.AccountByUser(userAddress)
}

func (s *meteredAccountQueryStore) AccountByRaw(rawAddress string) (Account, bool, error) {
	s.byRaw++
	return s.AccountQueryStore.AccountByRaw(rawAddress)
}

func (s *meteredAccountQueryStore) AccountByReputation(reputationID string) (Account, bool, error) {
	s.byReputation++
	return s.AccountQueryStore.AccountByReputation(reputationID)
}

func (s *meteredAccountQueryStore) AccountsAfter(cursor string, limit uint64) ([]Account, bool, error) {
	s.scans++
	return s.AccountQueryStore.AccountsAfter(cursor, limit)
}

func newTestAccountQueryService(t *testing.T, store AccountQueryStore) AccountQueryService {
	t.Helper()
	query, err := NewAccountQueryService(store, prototype.DefaultParams())
	require.NoError(t, err)
	return query
}
