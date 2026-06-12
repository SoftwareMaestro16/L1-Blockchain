package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExternalMessageFromActiveAccountSucceedsWithValidAuth(t *testing.T) {
	account := completeActiveAccount(t, 0xb1, 500, 7)

	next, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{account.PubKeys[0]},
		Operation:	"send",
	})

	require.NoError(t, err)
	require.Equal(t, account.Sequence+1, next.Sequence)
	require.Equal(t, account.AddressUser, next.AddressUser)
	require.Equal(t, account.AddressRaw, next.AddressRaw)
}

func TestExternalMessageFromInactiveOrFrozenAccountRejected(t *testing.T) {
	inactive := completeActiveAccount(t, 0xb2, 501, 0)
	inactive.Status = AccountStatusInactive
	_, err := ApplyExternalMessage(inactive, ExternalMessage{
		AccountUser:	inactive.AddressUser,
		Sequence:	inactive.Sequence,
		Signers:	[]string{inactive.PubKeys[0]},
	})
	require.ErrorContains(t, err, "inactive account cannot send external messages")

	frozen := completeActiveAccount(t, 0xb3, 502, 0)
	frozen.Status = AccountStatusFrozen
	_, err = ApplyExternalMessage(frozen, ExternalMessage{
		AccountUser:	frozen.AddressUser,
		Sequence:	frozen.Sequence,
		Signers:	[]string{frozen.PubKeys[0]},
	})
	require.ErrorContains(t, err, "frozen account cannot send external messages")
}

func TestExternalMessageRejectsInvalidAuthAndSequence(t *testing.T) {
	account := completeActiveAccount(t, 0xb4, 503, 9)

	_, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence + 1,
		Signers:	[]string{account.PubKeys[0]},
	})
	require.ErrorContains(t, err, "does not match account sequence")

	_, err = ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"ed25519:not-authorized"},
	})
	require.ErrorContains(t, err, "missing authorized")
}

func TestInternalMessageAcceptedByEnabledFeatureRule(t *testing.T) {
	account := completeActiveAccount(t, 0xb5, 504, 12)
	policy := InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2}

	next, err := ApplyInternalMessage(account, InternalMessage{
		AccountUser:	account.AddressUser,
		Source:		InternalMessageSourceContract,
		Feature:	AccountFeatureInternalMessagesV2,
		Operation:	"contract_callback",
	}, policy)

	require.NoError(t, err)
	require.Equal(t, account, next)
}

func TestInternalMessageRejectedWhenFeatureDisabled(t *testing.T) {
	account := completeActiveAccount(t, 0xb6, 505, 13)
	account.FeatureFlags = []string{AccountFeatureMetadataV2, AccountFeatureRecoveryPolicyV2}
	policy := InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2}

	_, err := ApplyInternalMessage(account, InternalMessage{
		AccountUser:	account.AddressUser,
		Source:		InternalMessageSourceModule,
		Feature:	AccountFeatureInternalMessagesV2,
		Operation:	"module_notice",
	}, policy)

	require.ErrorContains(t, err, "feature disabled")
}

func TestInternalMessageRulesMigrationPreservesExistingAccountAddresses(t *testing.T) {
	v1 := v1Account(t, 0xb7, 506, 14)
	migrated, err := MigrateAccountV1ToV2(v1)
	require.NoError(t, err)

	next, err := ApplyInternalMessage(migrated, InternalMessage{
		AccountUser:	migrated.AddressUser,
		Source:		InternalMessageSourceSystem,
		Feature:	AccountFeatureInternalMessagesV2,
		Operation:	"system_notice",
	}, InternalMessagePolicy{Version: 2, EnabledFeature: AccountFeatureInternalMessagesV2})

	require.NoError(t, err)
	require.Equal(t, v1.AddressUser, next.AddressUser)
	require.Equal(t, v1.AddressRaw, next.AddressRaw)
	require.Equal(t, v1.Sequence, next.Sequence)
}

func TestInternalMessageHandlingDoesNotIncrementUserSequence(t *testing.T) {
	account := completeActiveAccount(t, 0xb8, 507, 15)

	next, err := ApplyInternalMessage(account, InternalMessage{
		AccountUser:	account.AddressUser,
		Source:		InternalMessageSourceModule,
		Feature:	AccountFeatureInternalMessagesV2,
		Operation:	"module_update",
	}, InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2})

	require.NoError(t, err)
	require.Equal(t, account.Sequence, next.Sequence)
}

func TestInternalMessagesRespectFrozenRestrictionsUnlessWhitelisted(t *testing.T) {
	account := completeActiveAccount(t, 0xb9, 508, 16)
	account.Status = AccountStatusFrozen
	policy := InternalMessagePolicy{Version: 1, EnabledFeature: AccountFeatureInternalMessagesV2}

	_, err := ApplyInternalMessage(account, InternalMessage{
		AccountUser:	account.AddressUser,
		Source:		InternalMessageSourceContract,
		Feature:	AccountFeatureInternalMessagesV2,
		Operation:	"contract_callback",
	}, policy)
	require.ErrorContains(t, err, "explicit whitelist")

	next, err := ApplyInternalMessage(account, InternalMessage{
		AccountUser:		account.AddressUser,
		Source:			InternalMessageSourceContract,
		Feature:		AccountFeatureInternalMessagesV2,
		Operation:		"storage_debt_payment",
		WhitelistedWhileFrozen:	true,
	}, policy)
	require.NoError(t, err)
	require.Equal(t, account.Sequence, next.Sequence)
	require.Equal(t, AccountStatusFrozen, next.Status)
}
