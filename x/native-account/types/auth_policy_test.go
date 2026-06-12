package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	authKeyPrimaryPub	= "ed25519:primary"
	authKeyDevicePub	= "ed25519:device"
	authKeyRecoveryPub	= "ed25519:recovery"
	authKeyBackupPub	= "ed25519:backup"
)

func TestSingleKeyPolicyAuthorizesNormalTx(t *testing.T) {
	account := completeActiveAccount(t, 0xc1, 600, 1)
	account.AuthPolicy = AuthPolicy{Version: 1, Mode: AuthModeSingleKey}

	next, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{account.PubKeys[0]},
		Operation:	AuthOperationTransfer,
		Amount:		10,
	})

	require.NoError(t, err)
	require.Equal(t, account.Sequence+1, next.Sequence)
}

func TestMultisigThresholdPolicyRejectsInsufficientSignatures(t *testing.T) {
	account := accountWithPolicy(t, AuthPolicy{
		Version:	1,
		Mode:		AuthModeThreshold,
		Keys:		authKeys(),
		Threshold:	2,
	})

	_, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary"},
		Operation:	AuthOperationTransfer,
	})
	require.ErrorContains(t, err, "below threshold")

	next, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary", "device"},
		Operation:	AuthOperationTransfer,
	})
	require.NoError(t, err)
	require.Equal(t, account.Sequence+1, next.Sequence)
}

func TestWeightedMultisigSumsWeightsDeterministically(t *testing.T) {
	account := accountWithPolicy(t, AuthPolicy{
		Version:	1,
		Mode:		AuthModeWeighted,
		Keys:		authKeys(),
		Threshold:	7,
		Weights: []AuthWeight{
			{KeyID: "recovery", Weight: 1},
			{KeyID: "primary", Weight: 5},
			{KeyID: "device", Weight: 3},
		},
	})

	_, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary"},
		Operation:	AuthOperationTransfer,
	})
	require.ErrorContains(t, err, "below threshold")

	result, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"device", "primary"},
		Operation:	AuthOperationTransfer,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(8), result.Weight)

	normalized := account.AuthPolicy.Normalize()
	require.Equal(t, []AuthWeight{
		{KeyID: "device", Weight: 3},
		{KeyID: "primary", Weight: 5},
		{KeyID: "recovery", Weight: 1},
	}, normalized.Weights)
}

func TestTwoDevicePolicyRequiresBothKeysForProtectedOperations(t *testing.T) {
	account := accountWithPolicy(t, AuthPolicy{
		Version:	1,
		Mode:		AuthModeTwoDevice,
		Keys:		authKeys(),
		Threshold:	2,
	})

	_, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary"},
		Operation:	AuthOperationStakingChange,
	})
	require.ErrorContains(t, err, "primary and device")

	next, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary", "device"},
		Operation:	AuthOperationStakingChange,
	})
	require.NoError(t, err)
	require.Equal(t, account.Sequence+1, next.Sequence)
}

func TestSpendingLimitAllowsSmallTransferAndRejectsLargeTransfer(t *testing.T) {
	account := accountWithPolicy(t, AuthPolicy{
		Version:	1,
		Mode:		AuthModeTwoDevice,
		Keys:		authKeys(),
		SpendingLimits: []SpendingLimit{
			{Operation: AuthOperationTransfer, MaxAmount: 100},
		},
	})

	_, err := ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary"},
		Operation:	AuthOperationTransfer,
		Amount:		100,
	})
	require.NoError(t, err)

	_, err = ApplyExternalMessage(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	[]string{"primary"},
		Operation:	AuthOperationTransfer,
		Amount:		101,
	})
	require.ErrorContains(t, err, "primary and device")
}

func TestTimelockPreventsEarlyRecoveryAndAuthChange(t *testing.T) {
	account := accountWithPolicy(t, recoveryPolicy(100))

	_, err := ApplyMsgRecoverAccount(account, MsgRecoverAccount{
		AccountUser:	account.AddressUser,
		Signers:	[]string{authKeyRecoveryPub},
		CurrentHeight:	99,
	})
	require.ErrorContains(t, err, "timelock")

	_, err = ApplyMsgUpdateAuthPolicy(account, MsgUpdateAuthPolicy{
		AccountUser:	account.AddressUser,
		NewAuthPolicy:	AuthPolicy{Version: 1, Mode: AuthModeSingleKey},
		Signers:	[]string{"primary", "device"},
		CurrentHeight:	99,
	})
	require.ErrorContains(t, err, "timelock")
}

func TestRecoveryPolicyChangesStatusAfterValidAuthorization(t *testing.T) {
	account := accountWithPolicy(t, recoveryPolicy(10))
	account.Status = AccountStatusFrozen

	recovered, err := ApplyMsgRecoverAccount(account, MsgRecoverAccount{
		AccountUser:	account.AddressUser,
		Signers:	[]string{authKeyRecoveryPub},
		CurrentHeight:	10,
	})

	require.NoError(t, err)
	require.Equal(t, AccountStatusRecovered, recovered.Status)
	require.Equal(t, account.AddressUser, recovered.AddressUser)
	require.Equal(t, account.AddressRaw, recovered.AddressRaw)
}

func TestKeyRotationPreservesAEAndRawAddresses(t *testing.T) {
	account := accountWithPolicy(t, AuthPolicy{
		Version:	1,
		Mode:		AuthModeTwoDevice,
		Keys:		authKeys(),
	})

	rotated, err := ApplyMsgRotateKey(account, MsgRotateKey{
		AccountUser:	account.AddressUser,
		OldKeyID:	"device",
		NewKey:		AuthKey{ID: "device", PublicKey: "ed25519:new-device", Role: AuthKeyRoleDevice},
		Signers:	[]string{"primary", "device"},
	})

	require.NoError(t, err)
	require.Equal(t, account.AddressUser, rotated.AddressUser)
	require.Equal(t, account.AddressRaw, rotated.AddressRaw)
	require.Equal(t, "ed25519:new-device", authKeyByID(rotated.AuthPolicy.Keys, "device").PublicKey)
}

func TestAuthPolicyUpdateRequiresAuthorization(t *testing.T) {
	account := accountWithPolicy(t, AuthPolicy{
		Version:	1,
		Mode:		AuthModeTwoDevice,
		Keys:		authKeys(),
	})
	nextPolicy := AuthPolicy{
		Version:	1,
		Mode:		AuthModeThreshold,
		Keys:		authKeys(),
		Threshold:	2,
	}

	_, err := ApplyMsgUpdateAuthPolicy(account, MsgUpdateAuthPolicy{
		AccountUser:	account.AddressUser,
		NewAuthPolicy:	nextPolicy,
		Signers:	[]string{"primary"},
	})
	require.ErrorContains(t, err, "primary and device")

	updated, err := ApplyMsgUpdateAuthPolicy(account, MsgUpdateAuthPolicy{
		AccountUser:	account.AddressUser,
		NewAuthPolicy:	nextPolicy,
		Signers:	[]string{"primary", "device"},
	})
	require.NoError(t, err)
	require.Equal(t, AuthModeThreshold, updated.AuthPolicy.Mode)
	require.Equal(t, account.AddressUser, updated.AddressUser)
	require.Equal(t, account.AddressRaw, updated.AddressRaw)
}

func TestAuthPolicySerializationRejectsPrivateSeedSMSTOTPSecrets(t *testing.T) {
	fixtures := []AuthPolicy{
		{Version: 1, Mode: AuthModeSingleKey, Keys: []AuthKey{{ID: "primary", PublicKey: "private_key:bad", Role: AuthKeyRolePrimary}}},
		{Version: 1, Mode: AuthModeSingleKey, Keys: []AuthKey{{ID: "primary", PublicKey: "seed phrase bad", Role: AuthKeyRolePrimary}}},
		{Version: 1, Mode: AuthModeSingleKey, Keys: []AuthKey{{ID: "sms_secret", PublicKey: "ed25519:ok", Role: AuthKeyRolePrimary}}},
		{Version: 1, Mode: AuthModeSingleKey, Keys: []AuthKey{{ID: "totp_secret", PublicKey: "ed25519:ok", Role: AuthKeyRolePrimary}}},
	}

	for _, policy := range fixtures {
		require.Error(t, policy.Validate())
	}

	account := accountWithPolicy(t, AuthPolicy{Version: 1, Mode: AuthModeSingleKey})
	bz, err := json.Marshal(account)
	require.NoError(t, err)
	lower := strings.ToLower(string(bz))
	require.NotContains(t, lower, "private_key")
	require.NotContains(t, lower, "seed phrase")
	require.NotContains(t, lower, "sms_secret")
	require.NotContains(t, lower, "totp_secret")
}

func accountWithPolicy(t *testing.T, policy AuthPolicy) Account {
	t.Helper()
	account := completeActiveAccount(t, 0xc2, 601, 2)
	account.AuthPolicy = policy.Normalize()
	require.NoError(t, ValidateAccountInvariant(account))
	return account
}

func authKeys() []AuthKey {
	return []AuthKey{
		{ID: "primary", PublicKey: authKeyPrimaryPub, Role: AuthKeyRolePrimary},
		{ID: "device", PublicKey: authKeyDevicePub, Role: AuthKeyRoleDevice},
		{ID: "recovery", PublicKey: authKeyRecoveryPub, Role: AuthKeyRoleRecovery},
	}
}

func authKeyByID(keys []AuthKey, id string) AuthKey {
	for _, key := range keys {
		if key.ID == id {
			return key
		}
	}
	return AuthKey{}
}

func recoveryPolicy(height uint64) AuthPolicy {
	return AuthPolicy{
		Version:	1,
		Mode:		AuthModeTwoDevice,
		Keys:		authKeys(),
		RecoveryPolicy: RecoveryPolicy{
			Keys:			[]string{authKeyRecoveryPub, authKeyBackupPub},
			Threshold:		1,
			TimelockEndHeight:	height,
		},
		Timelock: TimelockPolicy{
			AuthPolicyUpdateEndHeight:	height,
			RecoveryEndHeight:		height,
		},
	}
}
