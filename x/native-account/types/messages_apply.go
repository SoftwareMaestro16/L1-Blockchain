package types

import (
	"errors"
	"fmt"
)

func ApplyMsgUpdateAuthPolicy(account Account, msg MsgUpdateAuthPolicy) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("auth policy update account address mismatch")
	}
	if msg.CurrentHeight < account.AuthPolicy.Timelock.AuthPolicyUpdateEndHeight {
		return Account{}, errors.New("auth policy update timelock has not expired")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationAuthPolicyUpdate,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	nextPolicy := msg.NewAuthPolicy.Normalize()
	if err := nextPolicy.Validate(); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.AuthPolicy = nextPolicy
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgRotateKey(account Account, msg MsgRotateKey) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("key rotation account address mismatch")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationAuthPolicyUpdate,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	if containsSecretLikeText(msg.NewKey.ID) || containsSecretLikeText(msg.NewKey.PublicKey) {
		return Account{}, errors.New("native account rotated key must not contain private keys or seed phrases")
	}
	next := cloneAccount(account)
	next.AuthPolicy = next.AuthPolicy.Normalize()
	replaced := false
	for idx, key := range next.AuthPolicy.Keys {
		if key.ID == msg.OldKeyID {
			next.AuthPolicy.Keys[idx] = msg.NewKey.Normalize()
			replaced = true
			break
		}
	}
	if !replaced {
		for idx, key := range next.PubKeys {
			if key == msg.OldKeyID {
				next.PubKeys[idx] = msg.NewKey.PublicKey
				replaced = true
				break
			}
		}
	}
	if !replaced {
		return Account{}, errors.New("native account key to rotate not found")
	}
	next = normalizeAccount(next)
	next.AuthPolicy = next.AuthPolicy.Normalize()
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgRecoverAccount(account Account, msg MsgRecoverAccount) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("recovery account address mismatch")
	}
	if err := AuthorizeRecoveryPolicy(account, msg); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Status = AccountStatusRecovered
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgFreezeAccount(account Account, msg MsgFreezeAccount) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("freeze account address mismatch")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationFreezeAccount,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Status = AccountStatusFrozen
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgUnfreezeAccount(account Account, msg MsgUnfreezeAccount) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("unfreeze account address mismatch")
	}
	if account.Status != AccountStatusFrozen {
		return Account{}, errors.New("only frozen accounts can be unfrozen")
	}
	if !msg.StorageDebtPaid || account.StorageRentDebt != 0 || msg.OtherFreezeReason {
		return Account{}, errors.New("account cannot unfreeze until storage debt is paid and freeze reasons are cleared")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationUnfreezeAccount,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Status = AccountStatusActive
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgPayStorageDebt(account Account, msg MsgPayStorageDebt) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("storage debt payment account address mismatch")
	}
	if account.Status == AccountStatusInactive || account.Status == AccountStatusClosed || account.Status == AccountStatusArchived {
		return Account{}, fmt.Errorf("%s account cannot pay storage debt", account.Status)
	}
	if msg.Amount == 0 {
		return Account{}, errors.New("storage debt payment amount must be positive")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationPayStorageDebt,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	if msg.Amount >= next.StorageRentDebt {
		next.StorageRentDebt = 0
	} else {
		next.StorageRentDebt -= msg.Amount
	}
	if msg.CurrentHeight > next.LastStorageChargeHeight {
		next.LastStorageChargeHeight = msg.CurrentHeight
	}
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgUpdateAccountMetadata(account Account, msg MsgUpdateAccountMetadata) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("metadata update account address mismatch")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationMetadataUpdate,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	if err := msg.Metadata.Validate(); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Metadata = msg.Metadata
	return next, ValidateAccountInvariant(next)
}

func ApplyMsgUpdateAccountParams(account Account, msg MsgUpdateAccountParams) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("account params update address mismatch")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:	account.AddressUser,
		Sequence:	account.Sequence,
		Signers:	msg.Signers,
		Operation:	AuthOperationParamsUpdate,
		CurrentHeight:	msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.FeatureFlags = cloneStrings(msg.FeatureFlags)
	next = normalizeAccount(next)
	return next, ValidateAccountInvariant(next)
}
