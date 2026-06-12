package types

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountValidationAcceptsCompleteActiveAccount(t *testing.T) {
	account := completeActiveAccount(t, 0x81, 700, 12)

	require.NoError(t, ValidateAccountInvariant(account))
}

func TestAccountValidationRejectsSecretsInSerializedFields(t *testing.T) {
	fixtures := []struct {
		name	string
		mutate	func(*Account)
	}{
		{
			name:	"pubkey private key",
			mutate: func(account *Account) {
				account.PubKeys = []string{"private_key:do-not-store"}
			},
		},
		{
			name:	"auth seed phrase",
			mutate: func(account *Account) {
				account.AuthPolicy.Mode = "seed phrase recovery"
			},
		},
		{
			name:	"feature mnemonic",
			mutate: func(account *Account) {
				account.FeatureFlags = []string{"mnemonic"}
			},
		},
		{
			name:	"metadata seed",
			mutate: func(account *Account) {
				account.Metadata.MetadataHash = "seed_phrase:bad"
			},
		},
		{
			name:	"reputation private key",
			mutate: func(account *Account) {
				account.ReputationID = "private key reference"
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			account := completeActiveAccount(t, 0x82, 701, 13)
			fixture.mutate(&account)

			err := ValidateAccountInvariant(account)

			require.Error(t, err)
			require.Contains(t, strings.ToLower(err.Error()), "private keys or seed phrases")
		})
	}
}

func TestAccountValidationRejectsMalformedState(t *testing.T) {
	fixtures := []struct {
		name	string
		mutate	func(*Account)
		wantErr	string
	}{
		{
			name:	"empty AE address",
			mutate: func(account *Account) {
				account.AddressUser = ""
			},
			wantErr:	"AE user-facing",
		},
		{
			name:	"malformed raw address",
			mutate: func(account *Account) {
				account.AddressRaw = "4:abcdef"
			},
			wantErr:	"invalid native account raw address",
		},
		{
			name:	"mismatched address pair",
			mutate: func(account *Account) {
				_, raw := testAddressPair(t, 0x83)
				account.AddressRaw = raw
			},
			wantErr:	"must represent the same account",
		},
		{
			name:	"inactive persistent account",
			mutate: func(account *Account) {
				account.Status = AccountStatusInactive
			},
			wantErr:	"virtual only",
		},
		{
			name:	"unsupported status",
			mutate: func(account *Account) {
				account.Status = "deleted"
			},
			wantErr:	"unsupported native account status",
		},
		{
			name:	"unsupported version",
			mutate: func(account *Account) {
				account.Version = CurrentAccountVersion + 1
			},
			wantErr:	"unsupported native account version",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			account := completeActiveAccount(t, 0x84, 702, 14)
			fixture.mutate(&account)

			require.ErrorContains(t, ValidateAccountInvariant(account), fixture.wantErr)
		})
	}
}

func TestAccountExportImportPreservesStateExactly(t *testing.T) {
	accountA := completeActiveAccount(t, 0x86, 800, 21)
	accountB := completeActiveAccount(t, 0x85, 801, 22)
	accountB.Status = AccountStatusFrozen
	accountB.StorageRentDebt = 99
	source := newTestAccountStore(accountA, accountB)

	exported, err := ExportGenesis(source)
	require.NoError(t, err)
	target := newTestAccountStore()
	require.NoError(t, ImportGenesis(target, exported))
	roundTrip, err := ExportGenesis(target)
	require.NoError(t, err)

	require.Equal(t, exported, roundTrip)
	require.Less(t, roundTrip.Accounts[0].AddressUser, roundTrip.Accounts[1].AddressUser)
}

func TestAccountMetadataSizeAndFieldsAreBounded(t *testing.T) {
	account := completeActiveAccount(t, 0x87, 900, 31)
	account.Metadata.MetadataHash = strings.Repeat("m", MaxMetadataHashBytes)
	account.Metadata.DisplayNameHash = strings.Repeat("d", MaxDisplayNameHashBytes)
	account.Metadata.DomainAlias = strings.Repeat("a", MaxDomainAliasBytes)
	require.NoError(t, ValidateAccountInvariant(account))

	account.Metadata.MetadataHash = strings.Repeat("m", MaxMetadataHashBytes+1)
	require.ErrorContains(t, ValidateAccountInvariant(account), "metadata hash exceeds")

	account = completeActiveAccount(t, 0x88, 901, 32)
	account.Metadata.DisplayNameHash = strings.Repeat("d", MaxDisplayNameHashBytes+1)
	require.ErrorContains(t, ValidateAccountInvariant(account), "display name hash exceeds")

	account = completeActiveAccount(t, 0x89, 902, 33)
	account.Metadata.DomainAlias = strings.Repeat("a", MaxDomainAliasBytes+1)
	require.ErrorContains(t, ValidateAccountInvariant(account), "domain alias exceeds")

	bz, err := json.Marshal(completeActiveAccount(t, 0x8a, 903, 34))
	require.NoError(t, err)
	text := string(bz)
	require.Contains(t, text, `"metadata_hash"`)
	require.Contains(t, text, `"display_name_hash"`)
	require.Contains(t, text, `"domain_alias"`)
	require.NotContains(t, text, "profile")
	require.NotContains(t, text, "avatar")
}

func TestAccountStateDoesNotDuplicateBalancesTokensNFTsDomainsOrHistory(t *testing.T) {
	forbidden := []string{
		"Balance",
		"Balances",
		"Token",
		"Tokens",
		"NFT",
		"NFTs",
		"DomainRecords",
		"Domains",
		"Transaction",
		"Transactions",
		"History",
		"Profile",
		"Avatar",
		"PrivateKey",
		"SeedPhrase",
	}
	accountType := reflect.TypeOf(Account{})
	for i := 0; i < accountType.NumField(); i++ {
		field := accountType.Field(i)
		for _, denied := range forbidden {
			require.NotContains(t, field.Name, denied, field.Name)
		}
	}

	bz, err := json.Marshal(completeActiveAccount(t, 0x8b, 904, 35))
	require.NoError(t, err)
	lower := strings.ToLower(string(bz))
	for _, denied := range []string{
		"private_key",
		"private key",
		"seed_phrase",
		"seed phrase",
		"mnemonic",
		"token_balances",
		"nfts",
		"all_domains",
		"transaction_history",
		"off_chain_profile",
		"avatar",
	} {
		require.NotContains(t, lower, denied)
	}
}

func TestAccountJSONUsesVersionedCanonicalFieldNames(t *testing.T) {
	bz, err := json.Marshal(completeActiveAccount(t, 0x8c, 905, 36))
	require.NoError(t, err)
	text := string(bz)

	for _, field := range []string{
		`"version"`,
		`"address_user"`,
		`"address_raw"`,
		`"pubkeys"`,
		`"account_number"`,
		`"sequence"`,
		`"status"`,
		`"auth_policy"`,
		`"features"`,
		`"metadata"`,
		`"reputation_id"`,
		`"created_height"`,
		`"last_active_height"`,
		`"last_storage_charge_height"`,
		`"storage_rent_debt"`,
	} {
		require.Contains(t, text, field)
	}
	require.NotContains(t, text, "AddressUser")
	require.NotContains(t, text, "FeatureFlags")
}

func completeActiveAccount(t *testing.T, fill byte, accountNumber, sequence uint64) Account {
	t.Helper()
	account := v1Account(t, fill, accountNumber, sequence)
	account.Version = AccountVersionV2
	account.FeatureFlags = []string{
		AccountFeatureInternalMessagesV2,
		AccountFeatureMetadataV2,
		AccountFeatureRecoveryPolicyV2,
	}
	return account
}
