package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActivateAccountSuccessInitializesDeterministicState(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	store := newTestAccountStore()
	service := newTestActivationService(t, store, 100)

	result, err := service.ActivateAccount(MsgActivateAccount{
		AddressUser:	pair.User,
		AddressRaw:	pair.Raw,
		PublicKey:	pubKey,
		FeePaid:	100,
	}, 77)

	require.NoError(t, err)
	require.Equal(t, CurrentAccountVersion, result.Account.Version)
	require.Equal(t, pair.User, result.Account.AddressUser)
	require.Equal(t, pair.Raw, result.Account.AddressRaw)
	require.Equal(t, []string{PublicKeyText(pubKey)}, result.Account.PubKeys)
	require.Equal(t, uint64(1), result.Account.AccountNumber)
	require.Equal(t, ActivationInitialSequence, result.Account.Sequence)
	require.Equal(t, AccountStatusActive, result.Account.Status)
	require.Equal(t, AuthModeSingleKey, result.Account.AuthPolicy.Mode)
	require.Equal(t, uint64(77), result.Account.CreatedHeight)
	require.Equal(t, result.Account.AddressUser, result.Event.AddressUser)
	require.Equal(t, EventTypeAccountActivated, result.Event.Type)

	persisted, found, err := store.AccountByUser(pair.User)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, result.Account, persisted)
}

func TestActivateAccountDuplicateRejected(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	store := newTestAccountStore()
	service := newTestActivationService(t, store, 1)

	_, err = service.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 1}, 10)
	require.NoError(t, err)
	_, err = service.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 1}, 11)

	require.ErrorContains(t, err, "already active")
}

func TestActivateAccountRejectsAddressNotDerivedFromPubKey(t *testing.T) {
	pubKey := activationTestPubKey()
	other := completeActiveAccount(t, 0xa1, 1, 0)
	service := newTestActivationService(t, newTestAccountStore(), 1)

	_, err := service.ActivateAccount(MsgActivateAccount{AddressUser: other.AddressUser, PublicKey: pubKey, FeePaid: 1}, 10)

	require.ErrorContains(t, err, "must equal derived")
}

func TestActivateAccountRejectsMalformedAEAndRawAddress(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	service := newTestActivationService(t, newTestAccountStore(), 1)

	_, err = service.ActivateAccount(MsgActivateAccount{AddressUser: "AE-not-valid", PublicKey: pubKey, FeePaid: 1}, 10)
	require.Error(t, err)

	_, err = service.ActivateAccount(MsgActivateAccount{
		AddressUser:	pair.User,
		AddressRaw:	"4:abcdef",
		PublicKey:	pubKey,
		FeePaid:	1,
	}, 10)
	require.ErrorContains(t, err, "invalid activation raw address")

	_, err = service.ActivateAccount(MsgActivateAccount{
		AddressUser:	pair.User,
		AddressRaw:	completeActiveAccount(t, 0xa2, 2, 0).AddressRaw,
		PublicKey:	pubKey,
		FeePaid:	1,
	}, 10)
	require.ErrorContains(t, err, "raw address must equal derived")
}

func TestActivateAccountRejectsFeeUnderMinimum(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	service := newTestActivationService(t, newTestAccountStore(), 100)

	_, err = service.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 99}, 10)

	require.ErrorContains(t, err, "below minimum")
}

func TestActivateAccountNumberAssignmentDeterministic(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	existing := completeActiveAccount(t, 0xa3, 41, 7)

	firstStore := newTestAccountStore(existing)
	first := newTestActivationService(t, firstStore, 1)
	firstResult, err := first.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 1}, 20)
	require.NoError(t, err)

	secondStore := newTestAccountStore(existing)
	second := newTestActivationService(t, secondStore, 1)
	secondResult, err := second.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 1}, 20)
	require.NoError(t, err)

	require.Equal(t, uint64(42), firstResult.Account.AccountNumber)
	require.Equal(t, firstResult.Account, secondResult.Account)
	require.Equal(t, firstResult.Event, secondResult.Event)
}

func TestAccountActivatedEventGolden(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	service := newTestActivationService(t, newTestAccountStore(), 100)

	result, err := service.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 123}, 55)
	require.NoError(t, err)
	bz, err := json.Marshal(result.Event)
	require.NoError(t, err)

	require.Equal(t, `{"type":"AccountActivated","address_user":"AEAAAQAAAAAAAAAAAAAAAHUedugZkZbUVJQcRdGzoyPxQzvW","address_raw":"4:000000000000000000000000751e76e8199196d454941c45d1b3a323f1433bd6","account_number":1,"sequence":0,"pubkey_hash":"0f715baf5d4c2ed329785cef29e562f73488c8a2bb9dbc5700b361d54b9b0554","height":55,"fee_paid":123}`, string(bz))
}

func TestActivatedAccountExportImportPreservesState(t *testing.T) {
	pubKey := activationTestPubKey()
	pair, err := ActivationAddressPair(pubKey)
	require.NoError(t, err)
	source := newTestAccountStore()
	service := newTestActivationService(t, source, 1)
	result, err := service.ActivateAccount(MsgActivateAccount{AddressUser: pair.User, PublicKey: pubKey, FeePaid: 1}, 99)
	require.NoError(t, err)

	exported, err := ExportGenesis(source)
	require.NoError(t, err)
	target := newTestAccountStore()
	require.NoError(t, ImportGenesis(target, exported))
	roundTrip, found, err := target.AccountByUser(pair.User)
	require.NoError(t, err)

	require.True(t, found)
	require.Equal(t, result.Account, roundTrip)
}

func newTestActivationService(t *testing.T, store AccountActivationStore, minFee uint64) AccountActivationService {
	t.Helper()
	service, err := NewAccountActivationService(store, ActivationFeePolicy{MinActivationFee: minFee})
	require.NoError(t, err)
	return service
}

func (s *testAccountStore) NextAccountNumber() uint64 {
	var max uint64
	for _, account := range s.accounts {
		if account.AccountNumber > max {
			max = account.AccountNumber
		}
	}
	return max + 1
}
