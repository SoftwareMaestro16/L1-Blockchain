package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestLazyMigrationPreservesExistingAccountAndTouchesSingleKey(t *testing.T) {
	store := newTestAccountStore(v1Account(t, 0x11, 42, 7))

	migrated, found, err := LazyMigrateAccount(store, store.accounts[0].AddressUser)

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, AccountVersionV2, migrated.Version)
	require.Equal(t, store.accounts[0].AddressUser, migrated.AddressUser)
	require.Equal(t, store.accounts[0].AddressRaw, migrated.AddressRaw)
	require.Equal(t, uint64(42), migrated.AccountNumber)
	require.Equal(t, uint64(7), migrated.Sequence)
	require.Equal(t, uint64(5), migrated.StorageRentDebt)
	require.Equal(t, 1, store.gets)
	require.Equal(t, 1, store.sets)
	require.Equal(t, 0, store.scans)
}

func TestUnsupportedVersionRejectedSafely(t *testing.T) {
	account := v1Account(t, 0x12, 43, 8)
	account.Version = CurrentAccountVersion + 1

	_, err := MigrateAccountIfNeeded(account)
	require.ErrorContains(t, err, "unsupported native account version")
	require.ErrorContains(t, ValidateAccountInvariant(account), "unsupported native account version")
}

func TestMigrateAccountV1ToV2DeterministicGolden(t *testing.T) {
	account := v1Account(t, 0x13, 44, 9)

	first, err := MigrateAccountV1ToV2(account)
	require.NoError(t, err)
	second, err := MigrateAccountV1ToV2(account)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Equal(t, AccountVersionV2, first.Version)
	require.Equal(t, []string{
		AccountFeatureInternalMessagesV2,
		AccountFeatureMetadataV2,
		AccountFeatureRecoveryPolicyV2,
	}, first.FeatureFlags)

	bz, err := json.Marshal(first)
	require.NoError(t, err)
	hash := sha256.Sum256(bz)
	require.Equal(t, "e811078d0945c08baf595a4fc7fea61d23fd1b628cd1408c976e8de2ccbde561", hex.EncodeToString(hash[:]))
}

func TestAddressAndSequenceSemanticsUnchangedAcrossMigration(t *testing.T) {
	account := v1Account(t, 0x14, 45, 123)

	migrated, err := MigrateAccountIfNeeded(account)

	require.NoError(t, err)
	require.Equal(t, account.AddressUser, migrated.AddressUser)
	require.Equal(t, account.AddressRaw, migrated.AddressRaw)
	require.Equal(t, account.AccountNumber, migrated.AccountNumber)
	require.Equal(t, account.Sequence, migrated.Sequence)
}

func TestFeatureDefaultsAreDeterministic(t *testing.T) {
	first, err := DefaultFeatureFlags(AccountVersionV2)
	require.NoError(t, err)
	second, err := DefaultFeatureFlags(AccountVersionV2)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Equal(t, []string{
		AccountFeatureInternalMessagesV2,
		AccountFeatureMetadataV2,
		AccountFeatureRecoveryPolicyV2,
	}, first)
	require.ErrorContains(t, validateFeatureFlags(AccountVersionV2, []string{
		AccountFeatureRecoveryPolicyV2,
		AccountFeatureMetadataV2,
	}), "sorted and unique")
}

func TestExportImportHandlesMixedAccountVersionsWithoutForcedMigration(t *testing.T) {
	v1 := v1Account(t, 0x21, 100, 1)
	v2, err := MigrateAccountV1ToV2(v1Account(t, 0x22, 101, 2))
	require.NoError(t, err)
	source := newTestAccountStore(v2, v1)

	exported, err := ExportGenesis(source)
	require.NoError(t, err)
	require.Equal(t, []uint64{AccountVersionV1, AccountVersionV2}, []uint64{exported.Accounts[0].Version, exported.Accounts[1].Version})

	target := newTestAccountStore()
	require.NoError(t, ImportGenesis(target, exported))
	roundTrip, err := ExportGenesis(target)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}

func TestImportGenesisRejectsDuplicateAndMalformedStateBeforeWrites(t *testing.T) {
	account := v1Account(t, 0x23, 200, 3)
	duplicate := v1Account(t, 0x24, 200, 4)
	store := newTestAccountStore()

	err := ImportGenesis(store, GenesisState{Version: 1, Accounts: []Account{account, duplicate}})

	require.ErrorContains(t, err, "duplicate native account number")
	require.Equal(t, 0, store.sets)
}

func TestBatchedMigrationResumesSafelyWithoutSkipOrDuplicate(t *testing.T) {
	store := newTestAccountStore(
		v1Account(t, 0x31, 1, 1),
		v1Account(t, 0x32, 2, 2),
		v1Account(t, 0x33, 3, 3),
		v1Account(t, 0x34, 4, 4),
		v1Account(t, 0x35, 5, 5),
	)
	processed := map[string]int{}
	cursor := ""
	totalMigrated := uint64(0)

	for {
		result, err := RunBatchedMigrationJob(store, cursor, 2)
		require.NoError(t, err)
		for _, user := range result.ProcessedUsers {
			processed[user]++
		}
		totalMigrated += result.Migrated
		cursor = result.NextCursor
		if result.Complete {
			break
		}
	}

	require.Equal(t, uint64(5), totalMigrated)
	require.Len(t, processed, 5)
	for _, count := range processed {
		require.Equal(t, 1, count)
	}
	for _, account := range store.accounts {
		require.Equal(t, AccountVersionV2, account.Version)
	}
	require.Equal(t, 3, store.scans)
}

func TestNativeAccountUpgradePlanRegistersLazyAndBatchHandlers(t *testing.T) {
	plan := NativeAccountVersionUpgradePlan()

	require.NoError(t, ValidateNativeAccountVersionUpgradePlan(plan))
	require.Equal(t, nativeAccountUpgradeModuleName, plan.ModuleName)
	require.True(t, plan.LazyMigrationEnabled)
	require.True(t, plan.BatchedMigrationEnabled)
	require.False(t, plan.RequiresFullBlockScan)
	require.Contains(t, plan.RegisteredHandlers, "MigrateAccountIfNeeded")
	require.Contains(t, plan.RegisteredHandlers, "RunBatchedMigrationJob")
}

func TestAccountInvariantRejectsAddressAndSecretSecurityCases(t *testing.T) {
	account := v1Account(t, 0x41, 300, 6)
	account.AddressUser = account.AddressRaw
	require.ErrorContains(t, ValidateAccountInvariant(account), "AE user-facing")

	account = v1Account(t, 0x42, 301, 7)
	account.AddressRaw = "4:abcdef"
	require.ErrorContains(t, ValidateAccountInvariant(account), "invalid native account raw address")

	account = v1Account(t, 0x43, 302, 8)
	account.PubKeys = []string{"private_key:do-not-store"}
	require.ErrorContains(t, ValidateAccountInvariant(account), "private keys or seed phrases")
}

type testAccountStore struct {
	accounts	[]Account
	gets		int
	sets		int
	scans		int
}

func newTestAccountStore(accounts ...Account) *testAccountStore {
	store := &testAccountStore{}
	for _, account := range accounts {
		if err := store.SetAccount(account); err != nil {
			panic(err)
		}
	}
	store.sets = 0
	return store
}

func (s *testAccountStore) AccountByUser(userAddress string) (Account, bool, error) {
	s.gets++
	for _, account := range s.accounts {
		if account.AddressUser == userAddress {
			return account, true, nil
		}
	}
	return Account{}, false, nil
}

func (s *testAccountStore) SetAccount(account Account) error {
	if err := ValidateAccountInvariant(account); err != nil {
		return err
	}
	s.sets++
	account = normalizeAccount(account)
	for idx, existing := range s.accounts {
		if existing.AddressUser == account.AddressUser {
			s.accounts[idx] = account
			s.accounts = SortAccounts(s.accounts)
			return nil
		}
	}
	s.accounts = append(s.accounts, account)
	s.accounts = SortAccounts(s.accounts)
	return nil
}

func (s *testAccountStore) AccountsAfter(cursor string, limit uint64) ([]Account, bool, error) {
	s.scans++
	accounts := SortAccounts(s.accounts)
	start := 0
	if cursor != "" {
		start = len(accounts)
		for idx, account := range accounts {
			if account.AddressUser > cursor {
				start = idx
				break
			}
		}
	}
	if start >= len(accounts) {
		return nil, false, nil
	}
	end := len(accounts)
	if limit > 0 && start+int(limit) < end {
		end = start + int(limit)
	}
	return accounts[start:end], end < len(accounts), nil
}

func v1Account(t *testing.T, fill byte, accountNumber, sequence uint64) Account {
	t.Helper()
	user, raw := testAddressPair(t, fill)
	return Account{
		Version:	AccountVersionV1,
		AddressUser:	user,
		AddressRaw:	raw,
		PubKeys:	[]string{fmt.Sprintf("ed25519:%064x", fill)},
		AccountNumber:	accountNumber,
		Sequence:	sequence,
		Status:		AccountStatusActive,
		AuthPolicy: AuthPolicy{
			Version:	1,
			Mode:		"single_key",
		},
		Metadata: AccountMetadata{
			MetadataHash:		fmt.Sprintf("meta-%02x", fill),
			DisplayNameHash:	fmt.Sprintf("display-%02x", fill),
			DomainAlias:		fmt.Sprintf("account-%02x.aet", fill),
			CreatedHeight:		10,
		},
		ReputationID:			fmt.Sprintf("rep-%02x", fill),
		CreatedHeight:			10,
		LastActiveHeight:		11,
		LastStorageChargeHeight:	12,
		StorageRentDebt:		5,
	}
}

func testAddressPair(t *testing.T, fill byte) (string, string) {
	t.Helper()
	addr := sdk.AccAddress(bytes20(fill))
	return addressing.FormatAccAddress(addr), addressing.Format(addr)
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
