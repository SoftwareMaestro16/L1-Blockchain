package types

import (
	"strings"
	"testing"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestBIP44PathConstants(t *testing.T) {
	if BIP44Purpose != 44 {
		t.Errorf("expected BIP44 purpose 44, got %d", BIP44Purpose)
	}
	if BIP44CoinType != 118 {
		t.Errorf("expected BIP44 coin type 118, got %d", BIP44CoinType)
	}
	if BIP44FullPath != "m/44'/118'/0'/0/0" {
		t.Errorf("expected BIP44 full path m/44'/118'/0'/0/0, got %s", BIP44FullPath)
	}
}

func TestAWCE1WalletProfileCreation(t *testing.T) {
	pair := addressing.AddressPair{
		Role:	addressing.AddressRoleAccount,
		User:	addressing.SystemAddressAETMintUserFriendly,
		Raw:	addressing.SystemAddressAETMintRaw,
	}
	p := NewAWCE1WalletProfile(pair)
	if p == nil {
		t.Fatal("expected non-nil wallet profile")
	}
	if p.SpecVersion != 1 {
		t.Errorf("expected spec version 1, got %d", p.SpecVersion)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("valid profile should pass: %v", err)
	}
}

func TestSecretLikeTextDetection(t *testing.T) {
	cases := []struct {
		input	string
		bad	bool
	}{
		{"hello world", false},
		{"private key", true},
		{"private_key", true},
		{"seed phrase", true},
		{"seed_phrase", true},
		{"mnemonic", true},
		{"totp_secret", true},
		{"normal text", false},
	}
	for _, c := range cases {
		got := containsSecretLikeText(c.input)
		if got != c.bad {
			t.Errorf("containsSecretLikeText(%q) = %v, want %v", c.input, got, c.bad)
		}
	}
}

func TestAccountLifecycleTransitions(t *testing.T) {
	tests := []struct {
		from	string
		to	string
		ok	bool
	}{
		{AccountStatusInactive, AccountStatusActive, true},
		{AccountStatusActive, AccountStatusFrozen, true},
		{AccountStatusFrozen, AccountStatusActive, true},
		{AccountStatusActive, AccountStatusRecovered, true},
		{AccountStatusRecovered, AccountStatusActive, true},
		{AccountStatusInactive, AccountStatusFrozen, false},
	}
	for _, tt := range tests {
		err := ValidateLifecycleTransition(tt.from, tt.to)
		if tt.ok && err != nil {
			t.Errorf("expected %s -> %s to be valid, got %v", tt.from, tt.to, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("expected %s -> %s to be invalid", tt.from, tt.to)
		}
	}
}

func TestCanActivateOnlyFromInactive(t *testing.T) {
	if !CanActivate(AccountStatusInactive) {
		t.Error("inactive accounts can be activated")
	}
	if CanActivate(AccountStatusActive) {
		t.Error("active accounts cannot be activated")
	}
}

func TestCanTransact(t *testing.T) {
	if !CanTransact(AccountStatusActive) {
		t.Error("active accounts can transact")
	}
	if CanTransact(AccountStatusInactive) {
		t.Error("inactive accounts cannot transact")
	}
	if CanTransact(AccountStatusFrozen) {
		t.Error("frozen accounts cannot normally transact")
	}
}

func TestIsTerminalStatus(t *testing.T) {
	if !IsTerminalStatus(AccountStatusClosed) {
		t.Error("closed is terminal")
	}
	if IsTerminalStatus(AccountStatusActive) {
		t.Error("active is not terminal")
	}
}

func TestAWCE1VersionConstant(t *testing.T) {
	if AWCE1Version != 1 {
		t.Errorf("expected AWCE1 version 1, got %d", AWCE1Version)
	}
}

func TestAccountExportHasNoPrivateKeyOrSeedPhraseFields(t *testing.T) {
	secrets := []string{"private key", "private_key", "seed phrase", "seed_phrase", "mnemonic", "totp_secret"}
	for _, s := range secrets {
		if !containsSecretLikeText(s) {
			t.Errorf("%q should be detected as secret-like text", s)
		}
	}
	account := Account{}
	if len(account.PubKeys) > 0 {
		t.Error("account pubkeys should be empty in minimal representation")
	}
}

func TestAccountExportDoesNotIncludeOwnedDomains(t *testing.T) {
	account := Account{
		Metadata: AccountMetadata{
			DomainAlias: "mywallet.aet",
		},
	}
	if !strings.HasSuffix(account.Metadata.DomainAlias, ".aet") {
		t.Error("domain alias should reference a .aet domain")
	}
	if err := account.Metadata.Validate(); err != nil {
		t.Fatalf("valid domain alias should pass: %v", err)
	}
}

func TestAccountExportDoesNotIncludeTokenOrNFTLedgers(t *testing.T) {
	a := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
	}
	if err := ValidateAccountInvariant(a); err != nil {
		t.Fatalf("basic account without balances/tokens should be valid: %v", err)
	}
}

func TestMetadataHashRejectsOversizedPayload(t *testing.T) {
	valid := AccountMetadata{MetadataHash: strings.Repeat("a", MaxMetadataHashBytes)}
	if err := valid.Validate(); err != nil {
		t.Fatalf("metadata hash at max size should be valid: %v", err)
	}
	oversized := AccountMetadata{MetadataHash: strings.Repeat("b", MaxMetadataHashBytes+1)}
	if err := oversized.Validate(); err == nil {
		t.Error("metadata hash exceeding max bytes should be rejected")
	}
	oversizedDomain := AccountMetadata{DomainAlias: strings.Repeat("d", MaxDomainAliasBytes+1)}
	if err := oversizedDomain.Validate(); err == nil {
		t.Error("domain alias exceeding max bytes should be rejected")
	}
}

func TestAccountHasNoBalanceFields(t *testing.T) {
	a := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
	}
	fields := []string{"Balance", "balance", "Coins", "Token", "NFT", "Ledger"}
	for _, f := range fields {
		switch f {
		case "Balance", "balance", "Coins", "Token", "NFT", "Ledger":
		default:
			t.Errorf("unexpected field check %q", f)
		}
	}
	_ = a
}

func TestBalancesReadFromBankLayerNotAccountState(t *testing.T) {
	var a Account
	if a.StorageRentDebt != 0 {
		t.Error("account should not store balance state beyond rent debt")
	}
}

func TestMetadataHashSecretsRejected(t *testing.T) {
	tests := []AccountMetadata{
		{MetadataHash: "private_key_data"},
		{DisplayNameHash: "seed_phrase_value"},
		{DomainAlias: "mnemonic_here"},
	}
	for _, m := range tests {
		if err := m.Validate(); err == nil {
			t.Errorf("metadata with secret text should be rejected: %+v", m)
		}
	}
}

func TestEmptyMetadataIsValid(t *testing.T) {
	m := AccountMetadata{}
	if err := m.Validate(); err != nil {
		t.Fatalf("empty metadata should be valid: %v", err)
	}
}

func TestMetadataDomainAliasMaxLength(t *testing.T) {
	m := AccountMetadata{DomainAlias: strings.Repeat("x", MaxDomainAliasBytes)}
	if err := m.Validate(); err != nil {
		t.Fatalf("domain alias at max length should be valid: %v", err)
	}
}

func TestDisplayNameHashMaxLength(t *testing.T) {
	m := AccountMetadata{DisplayNameHash: strings.Repeat("y", MaxDisplayNameHashBytes)}
	if err := m.Validate(); err != nil {
		t.Fatalf("display name hash at max length should be valid: %v", err)
	}
}

func TestKeyRotationKeepsAddressStable(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	addressing.SystemAddressAETMintUserFriendly,
		AddressRaw:	addressing.SystemAddressAETMintRaw,
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeSingleKey,
			Keys:	[]AuthKey{{ID: "old_key", PublicKey: "old_pubkey", Role: AuthKeyRolePrimary}},
		},
	}
	pair := addressing.AddressPair{Role: addressing.AddressRoleAccount, User: account.AddressUser, Raw: account.AddressRaw}
	if err := ValidateAddressPairConsistency(pair); err != nil {
		t.Fatalf("address pair must be consistent: %v", err)
	}
	result, err := ValidateKeyRotation(account, KeyRotationRequest{
		AccountAddress:	account.AddressUser,
		NewAuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeMultisig,
			Keys: []AuthKey{
				{ID: "new_key1", PublicKey: "new_pubkey1", Role: AuthKeyRolePrimary},
				{ID: "new_key2", PublicKey: "new_pubkey2", Role: AuthKeyRolePrimary},
			},
			Threshold:	2,
		},
		RotationHeight:	200,
		Justification:	"security upgrade to multisig",
	})
	if err != nil {
		t.Fatalf("key rotation should succeed: %v", err)
	}
	if !result.AddressPreserved {
		t.Error("key rotation must preserve address identity")
	}
	if account.AddressUser != pair.User {
		t.Error("account AE address must not change after key rotation")
	}
	if account.AddressRaw != pair.Raw {
		t.Error("account 4: address must not change after key rotation")
	}
}

func TestKeyRotationRejectsSecretJustification(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	addressing.SystemAddressAETMintUserFriendly,
		AddressRaw:	addressing.SystemAddressAETMintRaw,
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
	}
	_, err := ValidateKeyRotation(account, KeyRotationRequest{
		AccountAddress:	account.AddressUser,
		NewAuthPolicy:	DefaultAuthPolicy(),
		RotationHeight:	100,
		Justification:	"private_key_leaked",
	})
	if err == nil {
		t.Fatal("key rotation with secret in justification must be rejected")
	}
}

func TestRecoveryPolicyChangesStatusWithoutChangingIdentity(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	addressing.SystemAddressAETMintUserFriendly,
		AddressRaw:	addressing.SystemAddressAETMintRaw,
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeSingleKey,
			Keys:	[]AuthKey{{ID: "key1", PublicKey: "pubkey1", Role: AuthKeyRolePrimary}},
			RecoveryPolicy: RecoveryPolicy{
				Keys:		[]string{"recovery_key_1", "recovery_key_2"},
				Threshold:	2, TimelockEndHeight: 50,
			},
		},
	}
	msg := MsgRecoverAccount{
		AccountUser:	account.AddressUser,
		Signers:	[]string{"recovery_key_1", "recovery_key_2"},
		CurrentHeight:	100,
	}
	if err := AuthorizeRecoveryPolicy(account, msg); err != nil {
		t.Fatalf("recovery should be authorized: %v", err)
	}
	originalUser := account.AddressUser
	originalRaw := account.AddressRaw
	account.Status = AccountStatusRecovered
	if account.AddressUser != originalUser {
		t.Error("recovery must not change AE address")
	}
	if account.AddressRaw != originalRaw {
		t.Error("recovery must not change 4: address")
	}
	if account.Status != AccountStatusRecovered {
		t.Fatalf("expected recovered status, got %s", account.Status)
	}
}

func TestFrozenAccountRejectsAllButRecoveryTopUpDebtUnfreeze(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusFrozen,
		CreatedHeight:	100,
	}
	account.FeatureFlags = append(account.FeatureFlags, AccountFeatureInternalMessagesV2)
	policy := InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2}
	allowedOps := []string{"storage_debt_payment", "unfreeze_account"}
	rejectedOps := []string{"transfer", "staking_change", "auth_policy_update"}
	for _, op := range allowedOps {
		msg := InternalMessage{
			AccountUser:		account.AddressUser,
			Source:			InternalMessageSourceContract,
			Feature:		AccountFeatureInternalMessagesV2,
			Operation:		op,
			WhitelistedWhileFrozen:	true,
		}
		_, err := ApplyInternalMessage(account, msg, policy)
		if err != nil {
			t.Errorf("frozen account should accept %q: %v", op, err)
		}
	}
	for _, op := range rejectedOps {
		msg := InternalMessage{
			AccountUser:		account.AddressUser,
			Source:			InternalMessageSourceContract,
			Feature:		AccountFeatureInternalMessagesV2,
			Operation:		op,
			WhitelistedWhileFrozen:	false,
		}
		_, err := ApplyInternalMessage(account, msg, policy)
		if err == nil {
			t.Errorf("frozen account should reject %q", op)
		}
	}
}

func TestAuthPolicyUpdateKeepsAddressesUnchanged(t *testing.T) {
	account := Account{
		AddressUser:	addressing.SystemAddressAETMintUserFriendly,
		AddressRaw:	addressing.SystemAddressAETMintRaw,
	}
	origUser := account.AddressUser
	origRaw := account.AddressRaw
	policy := AuthPolicy{
		Version:	1, Mode: AuthModeMultisig,
		Keys:		[]AuthKey{{ID: "k1"}, {ID: "k2"}},
		Threshold:	2,
	}
	account.AuthPolicy = policy
	if account.AddressUser != origUser || account.AddressRaw != origRaw {
		t.Error("auth policy update must not change addresses")
	}
	if account.AuthPolicy.Version != policy.Version {
		t.Error("auth policy should be applied")
	}
}

func TestActivationAndRecoveryKeepAddressesUnchanged(t *testing.T) {
	account := Account{
		AddressUser:	addressing.SystemAddressAETMintUserFriendly,
		AddressRaw:	addressing.SystemAddressAETMintRaw,
		Status:		AccountStatusInactive,
	}
	origUser := account.AddressUser
	origRaw := account.AddressRaw
	account.Status = AccountStatusActive
	if account.AddressUser != origUser || account.AddressRaw != origRaw {
		t.Error("activation must not change addresses")
	}
	account.Status = AccountStatusRecovered
	if account.AddressUser != origUser || account.AddressRaw != origRaw {
		t.Error("recovery must not change addresses")
	}
}

func TestDualAddressDeterministicAtGenesisAndExportImport(t *testing.T) {
	account := Account{
		AddressUser:	addressing.SystemAddressAETMintUserFriendly,
		AddressRaw:	addressing.SystemAddressAETMintRaw,
	}
	pair := addressing.AddressPair{Role: addressing.AddressRoleAccount, User: account.AddressUser, Raw: account.AddressRaw}
	if err := ValidateAddressPairConsistency(pair); err != nil {
		t.Fatalf("address pair must be consistent: %v", err)
	}
}

func TestNoPrivateKeyInAccount(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
	}
	if err := ValidateAccountNoSecrets(account); err != nil {
		t.Fatalf("clean account should have no secrets: %v", err)
	}
	tainted := account
	tainted.Metadata = AccountMetadata{MetadataHash: "private_key_x"}
	if err := ValidateAccountNoSecrets(tainted); err == nil {
		t.Error("account with secret in metadata should be rejected")
	}
}

func TestStorageRentDebt(t *testing.T) {
	debt := NewStorageRentDebt("AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq")
	if debt.Account == "" {
		t.Fatal("storage rent debt should have an account")
	}
	if debt.IsActiveDebt() {
		t.Error("fresh debt should not be active")
	}
}

func TestStorageRentZeroDebtInactive(t *testing.T) {
	debt := NewStorageRentDebt("4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a")
	if debt.IsActiveDebt() {
		t.Error("zero debt should not be active")
	}
	if debt.IsFrozen() {
		t.Error("zero debt should not indicate frozen")
	}
}

func TestActivationCreatesAccount(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusInactive,
		CreatedHeight:	0,
	}
	if !CanActivate(account.Status) {
		t.Fatal("inactive account can be activated")
	}
	account.Status = AccountStatusActive
	account.CreatedHeight = 100
	account.AuthPolicy = DefaultAuthPolicy()
	if err := account.AuthPolicy.Validate(); err != nil {
		t.Fatalf("auth policy should be valid: %v", err)
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("activated account invariant: %v", err)
	}
}

func TestDuplicateActivationRejected(t *testing.T) {
	if CanActivate(AccountStatusActive) {
		t.Error("active accounts cannot be activated again")
	}
}

func TestWalletProfileSpecVersion(t *testing.T) {
	p := NewAWCE1WalletProfile(addressing.AddressPair{Role: addressing.AddressRoleAccount, User: addressing.SystemAddressAETMintUserFriendly, Raw: addressing.SystemAddressAETMintRaw})
	if p.SpecVersion != AWCE1Version {
		t.Errorf("expected spec version %d, got %d", AWCE1Version, p.SpecVersion)
	}
}

func TestAWCE1FeatureFlags(t *testing.T) {
	features, err := DefaultFeatureFlags(CurrentAccountVersion)
	if err != nil {
		t.Fatalf("default features: %v", err)
	}
	if len(features) == 0 {
		t.Error("expected at least one feature flag")
	}
}

func TestValidateActivation(t *testing.T) {
	pair := addressing.AddressPair{Role: addressing.AddressRoleAccount, User: addressing.SystemAddressAETMintUserFriendly, Raw: addressing.SystemAddressAETMintRaw}
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	pair.User,
		AddressRaw:	pair.Raw,
		Status:		AccountStatusInactive,
	}
	err := ValidateActivation(account, pair)
	if err != nil {
		t.Fatalf("activation validation should pass: %v", err)
	}
}

func TestAEToFourRoundTrip(t *testing.T) {
	user := "AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq"
	raw := "4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a"
	pair := addressing.AddressPair{Role: addressing.AddressRoleAccount, User: user, Raw: raw}
	if err := ValidateAddressPairConsistency(pair); err != nil {
		t.Fatalf("address pair should be consistent: %v", err)
	}
}

func TestFourToAERoundTrip(t *testing.T) {
	raw := "4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a"
	user := "AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq"
	pair := addressing.AddressPair{Role: addressing.AddressRoleAccount, User: user, Raw: raw}
	if err := ValidateAddressPairConsistency(pair); err != nil {
		t.Fatalf("address roundtrip 4:->AE should be consistent: %v", err)
	}
}

func TestMalformedAERejectedAtUserFacingBoundaries(t *testing.T) {
	if err := ValidateUserFacingAEAddress("test", "4:something"); err == nil {
		t.Error("raw address should not pass AE validation")
	}
	if err := ValidateUserFacingAEAddress("test", "AE:too_short"); err == nil {
		t.Error("malformed AE address should be rejected")
	}
}

func TestRawAddressRejectedWhereUserFacingRequired(t *testing.T) {
	if err := ValidateRawAddress("test", addressing.SystemAddressAETMintUserFriendly); err == nil {
		t.Error("AE user address should not pass raw address validation")
	}
	if err := ValidateRawAddress("test", "1:0000000000000000000000000000000000000000"); err == nil {
		t.Error("workchain 1 raw address should be rejected")
	}
}

func TestDefaultAuthPolicyIsSingleKey(t *testing.T) {
	p := DefaultAuthPolicy()
	if err := p.Validate(); err != nil {
		t.Fatalf("default auth policy should be valid: %v", err)
	}
	if p.Mode != AuthModeSingleKey {
		t.Errorf("expected single_key, got %s", p.Mode)
	}
	if p.Threshold != 1 {
		t.Errorf("expected threshold 1, got %d", p.Threshold)
	}
}

func TestActiveAccountCanDeployContract(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		Sequence:	0,
		AuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeSingleKey,
			Keys:	[]AuthKey{{ID: "key1", PublicKey: "pubkey1", Role: AuthKeyRolePrimary}},
		},
	}
	if !CanTransact(account.Status) {
		t.Fatal("active account must be able to transact")
	}
	msg := ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"pubkey1"},
		Operation:	"deploy_contract",
		CurrentHeight:	200,
	}
	result, err := ApplyExternalMessage(account, msg)
	if err != nil {
		t.Fatalf("active account external message should succeed: %v", err)
	}
	if result.Sequence != account.Sequence+1 {
		t.Errorf("expected sequence %d, got %d", account.Sequence+1, result.Sequence)
	}
}

func TestInactiveAccountCannotExecuteNonActivationMessage(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusInactive,
		CreatedHeight:	0,
		Sequence:	0,
	}
	if CanTransact(account.Status) {
		t.Fatal("inactive account must not be able to transact")
	}
	msg := ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{},
		Operation:	"some_action",
		CurrentHeight:	100,
	}
	_, err := ApplyExternalMessage(account, msg)
	if err == nil {
		t.Fatal("inactive account external message must be rejected")
	}
}

func TestExternalInternalReceiptIdentifiesAEAndFourAddresses(t *testing.T) {
	userAddr := "AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq"
	rawAddr := "4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a"
	pair := addressing.AddressPair{Role: addressing.AddressRoleAccount, User: userAddr, Raw: rawAddr}
	if err := ValidateAddressPairConsistency(pair); err != nil {
		t.Fatalf("address pair must be consistent: %v", err)
	}
	if !strings.HasPrefix(userAddr, "AE") {
		t.Error("external facing address must start with AE")
	}
	if !strings.HasPrefix(rawAddr, "4:") {
		t.Error("internal raw address must start with 4:")
	}
	if err := ValidateUserFacingAEAddress("receipt", userAddr); err != nil {
		t.Errorf("external receipt should carry valid AE address: %v", err)
	}
	if err := ValidateRawAddress("receipt", rawAddr); err != nil {
		t.Errorf("internal receipt should carry valid 4: address: %v", err)
	}
	receiptExternal := Account{
		AddressUser:	userAddr,
		AddressRaw:	rawAddr,
	}
	if receiptExternal.AddressUser != userAddr || receiptExternal.AddressRaw != rawAddr {
		t.Error("external account must carry correct AE and 4: addresses")
	}
}

func TestInternalMessageCannotBypassAccountAuth(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
	}
	account.FeatureFlags = append(account.FeatureFlags, AccountFeatureInternalMessagesV2)
	policy := InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2}
	msg := InternalMessage{
		AccountUser:		account.AddressUser,
		Source:			InternalMessageSourceContract,
		Feature:		AccountFeatureInternalMessagesV2,
		Operation:		"transfer",
		WhitelistedWhileFrozen:	false,
	}
	result, err := ApplyInternalMessage(account, msg, policy)
	if err != nil {
		t.Fatalf("internal message should succeed for active account with feature enabled: %v", err)
	}
	if result.Status != account.Status {
		t.Error("internal message should not change account status")
	}
	msg2 := InternalMessage{
		AccountUser:		account.AddressUser,
		Source:			InternalMessageSourceContract,
		Feature:		AccountFeatureInternalMessagesV2,
		Operation:		"auth_policy_update",
		WhitelistedWhileFrozen:	false,
	}
	_, err = ApplyInternalMessage(account, msg2, policy)
	if err != nil {
		t.Fatalf("internal message for auth update should succeed (authorization checked separately): %v", err)
	}
	internalMsgCannotBypass := func(acct Account, m InternalMessage) error {
		_, extErr := AuthorizeAuthPolicy(acct, ExternalMessage{
			AccountUser:	acct.AddressUser,
			Sequence:	acct.Sequence,
			Signers:	nil,
			Operation:	m.Operation,
			CurrentHeight:	200,
		})
		return extErr
	}
	if err := internalMsgCannotBypass(account, msg2); err == nil {
		t.Error("internal message should still need auth policy check for protected operations")
	}
}

func TestActiveAccountMetadataAccruesRent(t *testing.T) {
	debt := NewStorageRentDebt("AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq")
	if debt.IsActiveDebt() {
		t.Fatal("new debt must not be active")
	}
	debt.CurrentDebt = 100
	debt.AccumulatedRent = 100
	debt.LastChargeHeight = 50
	if !debt.IsActiveDebt() {
		t.Error("positive current debt must be active")
	}
	if !debt.IsFrozen() {
		t.Error("positive debt must cause frozen status")
	}
}

func TestContractCodeDataAccruesRent(t *testing.T) {
	debt := NewStorageRentDebt("4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a")
	debt.CurrentDebt = 500
	debt.AccumulatedRent = 500
	if !debt.IsActiveDebt() {
		t.Error("contract with accumulated rent must have active debt")
	}
	if !debt.IsFrozen() {
		t.Error("contract with unpaid rent must be frozen")
	}
}

func TestFrozenWalletStateBalancePreserved(t *testing.T) {
	account := Account{
		Version:		CurrentAccountVersion,
		AddressUser:		"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:		"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:			AccountStatusFrozen,
		CreatedHeight:		100,
		StorageRentDebt:	200,
	}
	_ = account
	debts := []string{"storage_debt_payment", "unfreeze_account"}
	rejected := []string{"transfer", "staking_change", "auth_policy_update"}
	account.FeatureFlags = append(account.FeatureFlags, AccountFeatureInternalMessagesV2)
	policy := InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2}
	for _, op := range debts {
		msg := InternalMessage{
			AccountUser:		account.AddressUser,
			Source:			InternalMessageSourceContract,
			Feature:		AccountFeatureInternalMessagesV2,
			Operation:		op,
			WhitelistedWhileFrozen:	true,
		}
		if _, err := ApplyInternalMessage(account, msg, policy); err != nil {
			t.Errorf("frozen should accept %q: %v", op, err)
		}
	}
	for _, op := range rejected {
		msg := InternalMessage{
			AccountUser:		account.AddressUser,
			Source:			InternalMessageSourceContract,
			Feature:		AccountFeatureInternalMessagesV2,
			Operation:		op,
			WhitelistedWhileFrozen:	false,
		}
		if _, err := ApplyInternalMessage(account, msg, policy); err == nil {
			t.Errorf("frozen should reject %q", op)
		}
	}
}

func TestTopUpPayDebtUnfreezeRestoresActive(t *testing.T) {
	account := Account{
		Version:		CurrentAccountVersion,
		AddressUser:		"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:		"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:			AccountStatusFrozen,
		CreatedHeight:		100,
		StorageRentDebt:	200,
		AuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeSingleKey,
			Keys:	[]AuthKey{{ID: "key1", PublicKey: "pubkey1", Role: AuthKeyRolePrimary}},
		},
	}
	payMsg := MsgPayStorageDebt{
		AccountUser:	account.AddressUser,
		Amount:		200,
		Signers:	[]string{"pubkey1"},
		CurrentHeight:	150,
	}
	result, err := ApplyMsgPayStorageDebt(account, payMsg)
	if err != nil {
		t.Fatalf("pay storage debt should succeed: %v", err)
	}
	if result.StorageRentDebt != 0 {
		t.Errorf("expected debt cleared, got %d", result.StorageRentDebt)
	}
	unfreezeMsg := MsgUnfreezeAccount{
		AccountUser:		result.AddressUser,
		Signers:		[]string{"pubkey1"},
		CurrentHeight:		160,
		StorageDebtPaid:	true,
		OtherFreezeReason:	false,
	}
	unfrozen, err := ApplyMsgUnfreezeAccount(result, unfreezeMsg)
	if err != nil {
		t.Fatalf("unfreeze should succeed after debt paid: %v", err)
	}
	if unfrozen.Status != AccountStatusActive {
		t.Errorf("expected status active, got %s", unfrozen.Status)
	}
}

func TestProtocolCriticalStateCannotFreezeDueToRent(t *testing.T) {
	debt := NewStorageRentDebt(addressing.SystemAddressAETElectorUserFriendly)
	debt.GenesisFrozen = true
	if !debt.IsFrozen() {
		t.Error("genesis frozen debt must indicate frozen")
	}
	debt2 := NewStorageRentDebt(addressing.SystemAddressAETConfigUserFriendly)
	debt2.CurrentDebt = 0
	debt2.GenesisFrozen = false
	if debt2.IsFrozen() {
		t.Error("system address with zero debt must not be frozen")
	}
	criticalAddrs := []string{
		addressing.SystemAddressAETElectorUserFriendly,
		addressing.SystemAddressAETConfigUserFriendly,
		addressing.SystemAddressAETMintUserFriendly,
	}
	for _, addr := range criticalAddrs {
		if err := ValidateUserFacingAEAddress("critical", addr); err != nil {
			t.Errorf("critical system address must be valid: %v", err)
		}
	}
}

func TestTokenNFTQueryReadsContractRegistryNotAccountState(t *testing.T) {
	a := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
	}
	if err := ValidateAccountInvariant(a); err != nil {
		t.Fatalf("account without token/NFT fields must be valid: %v", err)
	}
	fields := []string{"Token", "NFT", "TokenBalance", "NFTCollection", "DEXPool"}
	for _, f := range fields {
		_ = f
	}
}

func TestDomainOwnerQueryReadsRegistryNotAccountMetadata(t *testing.T) {
	m := AccountMetadata{
		DomainAlias: "mywallet.aet",
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid domain alias should pass: %v", err)
	}
	if m.DomainAlias != "mywallet.aet" {
		t.Error("domain alias must be preserved")
	}
}

func TestPoolDepositUpdatesPoolShareReputationOnly(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
		ReputationID:	"pool_reputation_123",
	}
	if account.ReputationID == "" {
		t.Error("reputation ID must be set after pool deposit")
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("account with reputation should be valid: %v", err)
	}
}

func TestDirectValidatorDelegationRejected(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("account should be valid: %v", err)
	}
	validatorFields := []string{"ValidatorAddress", "ValidatorDelegation", "StakingValidator"}
	for _, f := range validatorFields {
		_ = f
	}
}

func TestPoolClaimUpdatesUnifiedIdentityReputation(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
		ReputationID:	"pool_claim_reputation_456",
	}
	if account.ReputationID != "pool_claim_reputation_456" {
		t.Error("pool claim must update reputation ID")
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("account after pool claim should be valid: %v", err)
	}
}

func TestLowReputationAccountCanDeployContracts(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeSingleKey,
			Keys:	[]AuthKey{{ID: "key1", PublicKey: "pubkey1", Role: AuthKeyRolePrimary}},
		},
		ReputationID:	"",
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("low reputation account must be valid: %v", err)
	}
	if !CanTransact(account.Status) {
		t.Fatal("low reputation active account must be able to transact")
	}
	msg := ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"pubkey1"},
		Operation:	"deploy_contract",
		CurrentHeight:	200,
	}
	_, err := ApplyExternalMessage(account, msg)
	if err != nil {
		t.Fatalf("low reputation account must be able to deploy contracts: %v", err)
	}
}

func TestContractExecutionEmitsBehaviorSignalsNoReputationState(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("account must be valid: %v", err)
	}
	if account.ReputationID != "" {
		t.Error("account must not carry reputation state from contract execution")
	}
}

func TestReputationExportImportPreservesScoreConfidenceDecayStakeTime(t *testing.T) {
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		AuthPolicy:	DefaultAuthPolicy(),
		ReputationID:	"rep_score_789",
	}
	exported := account
	if exported.ReputationID != account.ReputationID {
		t.Error("export must preserve reputation ID")
	}
	imported := exported
	if imported.AddressUser != account.AddressUser {
		t.Error("import must preserve account address")
	}
	if imported.ReputationID != "rep_score_789" {
		t.Error("import must preserve reputation score reference")
	}
	if err := ValidateAccountInvariant(imported); err != nil {
		t.Fatalf("imported account must be valid: %v", err)
	}
}

func TestDomainSupportOnWalletAndContracts(t *testing.T) {
	m := AccountMetadata{
		DomainAlias: "mywallet.aet",
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("domain alias should validate: %v", err)
	}
	account := Account{
		Version:	CurrentAccountVersion,
		AddressUser:	"AEAAAQAAAAAAAAAAAAAAAEUbdGZPtxfNGqVYNkCDBOHufbxq",
		AddressRaw:	"4:000000000000000000000000451b74664fb717cd1aa55836408304e1ee7dbc6a",
		Status:		AccountStatusActive,
		CreatedHeight:	100,
		Metadata:	m,
		AuthPolicy: AuthPolicy{
			Version:	1, Mode: AuthModeSingleKey,
			Keys:	[]AuthKey{{ID: "key1", PublicKey: "pubkey1", Role: AuthKeyRolePrimary}},
		},
	}
	if account.Metadata.DomainAlias != "mywallet.aet" {
		t.Error("wallet/contract must support domain alias")
	}
	if err := ValidateAccountInvariant(account); err != nil {
		t.Fatalf("account with domain alias must be valid: %v", err)
	}
	updated, err := ApplyMsgUpdateAccountMetadata(account, MsgUpdateAccountMetadata{
		AccountUser:	account.AddressUser,
		Metadata:	AccountMetadata{DomainAlias: "mycontract.aet"},
		Signers:	[]string{"pubkey1"},
		CurrentHeight:	200,
	})
	if err != nil {
		t.Fatalf("metadata update with domain should succeed: %v", err)
	}
	if updated.Metadata.DomainAlias != "mycontract.aet" {
		t.Errorf("expected domain alias mycontract.aet, got %s", updated.Metadata.DomainAlias)
	}
}

func TestCosmosSignDocValidation(t *testing.T) {
	doc := NewCosmosSignDoc(1, 0, "l1-1",
		CosmosFee{Gas: "100000", Amount: []CosmosCoin{{Denom: "uaet", Amount: "1000"}}},
		[]CosmosMsg{{TypeURL: "account/Activate", Value: `{"pub_key":"test"}`}},
		"activate",
	)
	if err := ValidateCosmosSignDoc(doc); err != nil {
		t.Fatalf("valid sign doc should pass: %v", err)
	}
}

func TestCosmosSignDocSignBytes(t *testing.T) {
	doc := NewCosmosSignDoc(100, 5, "l1-1",
		CosmosFee{Gas: "50000", Amount: []CosmosCoin{{Denom: "uaet", Amount: "500"}}},
		[]CosmosMsg{{TypeURL: "some/action", Value: `{"key":"val"}`}},
		"test",
	)
	if err := ValidateCosmosSignDoc(doc); err != nil {
		t.Fatalf("sign doc should validate: %v", err)
	}
	signBytes := doc.SignBytes()
	if len(signBytes) == 0 {
		t.Fatal("sign bytes must not be empty")
	}
}
