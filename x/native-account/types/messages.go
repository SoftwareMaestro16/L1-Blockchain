package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	AccountMessageKindExternal = "external"
	AccountMessageKindInternal = "internal"

	AuthModeSingleKey = "single_key"

	InternalMessageSourceModule   = "module"
	InternalMessageSourceContract = "contract"
	InternalMessageSourceSystem   = "system"
)

type ExternalMessage struct {
	AccountUser   string
	Sequence      uint64
	Signers       []string
	Operation     string
	Amount        uint64
	CurrentHeight uint64
}

type InternalMessage struct {
	AccountUser            string
	Source                 string
	Feature                string
	Operation              string
	WhitelistedWhileFrozen bool
}

type InternalMessagePolicy struct {
	Version        uint64
	EnabledFeature string
}

func ApplyExternalMessage(account Account, msg ExternalMessage) (Account, error) {
	if err := ValidateExternalMessage(account, msg); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Sequence++
	return next, nil
}

func ValidateExternalMessage(account Account, msg ExternalMessage) error {
	if account.Status == AccountStatusInactive {
		return errors.New("inactive account cannot send external messages")
	}
	if account.Status != AccountStatusActive && account.Status != AccountStatusRecovered {
		return fmt.Errorf("%s account cannot send external messages", account.Status)
	}
	if msg.AccountUser != account.AddressUser {
		return errors.New("external message account address mismatch")
	}
	if msg.Sequence != account.Sequence {
		return fmt.Errorf("external message sequence %d does not match account sequence %d", msg.Sequence, account.Sequence)
	}
	return ValidateAuthPolicyForExternalMessage(account, msg)
}

func ValidateAuthPolicyForExternalMessage(account Account, msg ExternalMessage) error {
	_, err := AuthorizeAuthPolicy(account, msg)
	return err
}

func ApplyInternalMessage(account Account, msg InternalMessage, policy InternalMessagePolicy) (Account, error) {
	if err := ValidateInternalMessage(account, msg, policy); err != nil {
		return Account{}, err
	}
	return cloneAccount(account), nil
}

func ValidateInternalMessage(account Account, msg InternalMessage, policy InternalMessagePolicy) error {
	if msg.AccountUser != account.AddressUser {
		return errors.New("internal message account address mismatch")
	}
	if account.Status == AccountStatusInactive || account.Status == AccountStatusClosed || account.Status == AccountStatusArchived {
		return fmt.Errorf("%s account cannot receive internal messages", account.Status)
	}
	if account.Status == AccountStatusFrozen && !msg.WhitelistedWhileFrozen {
		return errors.New("frozen account internal messages require explicit whitelist")
	}
	if policy.Version == 0 {
		return errors.New("internal message policy version must be positive")
	}
	if strings.TrimSpace(policy.EnabledFeature) == "" {
		return errors.New("internal message policy feature is required")
	}
	if strings.TrimSpace(msg.Feature) == "" || msg.Feature != policy.EnabledFeature {
		return errors.New("internal message feature is not enabled by policy")
	}
	if !accountHasFeature(account, msg.Feature) {
		return errors.New("internal message feature disabled on account")
	}
	switch msg.Source {
	case InternalMessageSourceModule, InternalMessageSourceContract, InternalMessageSourceSystem:
		return nil
	default:
		return fmt.Errorf("unsupported internal message source %q", msg.Source)
	}
}

func accountHasFeature(account Account, feature string) bool {
	for _, existing := range account.FeatureFlags {
		if existing == feature {
			return true
		}
	}
	return false
}

type MsgUpdateAuthPolicy struct {
	AccountUser   string     `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	NewAuthPolicy AuthPolicy `protobuf:"bytes,2,opt,name=new_auth_policy,json=newAuthPolicy,proto3" json:"new_auth_policy"`
	Signers       []string   `protobuf:"bytes,3,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight uint64     `protobuf:"varint,4,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
}

type MsgRotateKey struct {
	AccountUser   string   `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	OldKeyID      string   `protobuf:"bytes,2,opt,name=old_key_id,json=oldKeyID,proto3" json:"old_key_id,omitempty"`
	NewKey        AuthKey  `protobuf:"bytes,3,opt,name=new_key,json=newKey,proto3" json:"new_key"`
	Signers       []string `protobuf:"bytes,4,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight uint64   `protobuf:"varint,5,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
}

type MsgRecoverAccount struct {
	AccountUser   string   `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	Signers       []string `protobuf:"bytes,2,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight uint64   `protobuf:"varint,3,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
}

type MsgFreezeAccount struct {
	AccountUser   string   `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	Signers       []string `protobuf:"bytes,2,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight uint64   `protobuf:"varint,3,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
}

type MsgUnfreezeAccount struct {
	AccountUser       string   `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	Signers           []string `protobuf:"bytes,2,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight     uint64   `protobuf:"varint,3,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
	StorageDebtPaid   bool     `protobuf:"varint,4,opt,name=storage_debt_paid,json=storageDebtPaid,proto3" json:"storage_debt_paid,omitempty"`
	OtherFreezeReason bool     `protobuf:"varint,5,opt,name=other_freeze_reason,json=otherFreezeReason,proto3" json:"other_freeze_reason,omitempty"`
}

type MsgPayStorageDebt struct {
	AccountUser   string   `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	Amount        uint64   `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	Signers       []string `protobuf:"bytes,3,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight uint64   `protobuf:"varint,4,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
}

type MsgUpdateAccountMetadata struct {
	AccountUser   string          `protobuf:"bytes,1,opt,name=account_user,json=accountUser,proto3" json:"account_user,omitempty"`
	Metadata      AccountMetadata `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata"`
	Signers       []string        `protobuf:"bytes,3,rep,name=signers,proto3" json:"signers,omitempty"`
	CurrentHeight uint64          `protobuf:"varint,4,opt,name=current_height,json=currentHeight,proto3" json:"current_height,omitempty"`
}

type MsgUpdateAccountParams struct {
	AccountUser   string
	FeatureFlags  []string
	Signers       []string
	CurrentHeight uint64
}

func ApplyMsgUpdateAuthPolicy(account Account, msg MsgUpdateAuthPolicy) (Account, error) {
	if msg.AccountUser != account.AddressUser {
		return Account{}, errors.New("auth policy update account address mismatch")
	}
	if msg.CurrentHeight < account.AuthPolicy.Timelock.AuthPolicyUpdateEndHeight {
		return Account{}, errors.New("auth policy update timelock has not expired")
	}
	if _, err := AuthorizeAuthPolicy(account, ExternalMessage{
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationAuthPolicyUpdate,
		CurrentHeight: msg.CurrentHeight,
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
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationAuthPolicyUpdate,
		CurrentHeight: msg.CurrentHeight,
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
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationFreezeAccount,
		CurrentHeight: msg.CurrentHeight,
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
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationUnfreezeAccount,
		CurrentHeight: msg.CurrentHeight,
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
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationPayStorageDebt,
		CurrentHeight: msg.CurrentHeight,
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
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationMetadataUpdate,
		CurrentHeight: msg.CurrentHeight,
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
		AccountUser:   account.AddressUser,
		Sequence:      account.Sequence,
		Signers:       msg.Signers,
		Operation:     AuthOperationParamsUpdate,
		CurrentHeight: msg.CurrentHeight,
	}); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.FeatureFlags = cloneStrings(msg.FeatureFlags)
	next = normalizeAccount(next)
	return next, ValidateAccountInvariant(next)
}
