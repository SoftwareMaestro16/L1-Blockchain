package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

type MsgActivateAccount struct {
	AddressUser string
	AddressRaw  string
	PublicKey   cryptotypes.PubKey
	FeePaid     uint64
}

func (m MsgActivateAccount) ValidateBasic() error {
	if err := addressing.ValidateUserAddress("activation address", m.AddressUser); err != nil {
		return err
	}
	pair, err := addressing.DeriveAccountAddress(m.PublicKey)
	if err != nil {
		return err
	}
	if pair.User != m.AddressUser {
		return fmt.Errorf("activation address must equal derived public key address")
	}
	if m.AddressRaw != "" {
		if err := ValidateRawAddress("activation raw address", m.AddressRaw); err != nil {
			return err
		}
		if pair.Raw != m.AddressRaw {
			return fmt.Errorf("activation raw address must equal derived public key raw address")
		}
	}
	return nil
}

func ActivationAddressPair(pubKey cryptotypes.PubKey) (addressing.AddressPair, error) {
	return addressing.DeriveAccountAddress(pubKey)
}

const (
	ActivationInitialSequence = uint64(0)
	EventTypeAccountActivated = "AccountActivated"
)

type ActivationFeePolicy struct {
	MinActivationFee uint64
}

func (p ActivationFeePolicy) ValidateActivationFee(feePaid uint64) error {
	if feePaid < p.MinActivationFee {
		return fmt.Errorf("activation fee %d below minimum %d", feePaid, p.MinActivationFee)
	}
	return nil
}

type AccountActivatedEvent struct {
	Type          string `json:"type"`
	AddressUser   string `json:"address_user"`
	AddressRaw    string `json:"address_raw"`
	AccountNumber uint64 `json:"account_number"`
	Sequence      uint64 `json:"sequence"`
	PubKeyHash    string `json:"pubkey_hash"`
	Height        uint64 `json:"height"`
	FeePaid       uint64 `json:"fee_paid"`
}

type AccountActivationResult struct {
	Account Account
	Event   AccountActivatedEvent
}

type AccountActivationStore interface {
	AccountByUser(userAddress string) (Account, bool, error)
	SetAccount(account Account) error
	NextAccountNumber() uint64
}

type AccountActivationService struct {
	store     AccountActivationStore
	feePolicy ActivationFeePolicy
}

func NewAccountActivationService(store AccountActivationStore, feePolicy ActivationFeePolicy) (AccountActivationService, error) {
	if store == nil {
		return AccountActivationService{}, fmt.Errorf("native account activation store is required")
	}
	return AccountActivationService{store: store, feePolicy: feePolicy}, nil
}

func (s AccountActivationService) ActivateAccount(msg MsgActivateAccount, createdHeight uint64) (AccountActivationResult, error) {
	if createdHeight == 0 {
		return AccountActivationResult{}, fmt.Errorf("activation height must be positive")
	}
	if err := msg.ValidateBasic(); err != nil {
		return AccountActivationResult{}, err
	}
	if err := s.feePolicy.ValidateActivationFee(msg.FeePaid); err != nil {
		return AccountActivationResult{}, err
	}
	pair, err := ActivationAddressPair(msg.PublicKey)
	if err != nil {
		return AccountActivationResult{}, err
	}
	if _, found, err := s.store.AccountByUser(pair.User); err != nil {
		return AccountActivationResult{}, err
	} else if found {
		return AccountActivationResult{}, fmt.Errorf("native account %s already active", pair.User)
	}
	features, err := DefaultFeatureFlags(CurrentAccountVersion)
	if err != nil {
		return AccountActivationResult{}, err
	}
	account := Account{
		Version:                 CurrentAccountVersion,
		AddressUser:             pair.User,
		AddressRaw:              pair.Raw,
		PubKeys:                 []string{PublicKeyText(msg.PublicKey)},
		AccountNumber:           s.store.NextAccountNumber(),
		Sequence:                ActivationInitialSequence,
		Status:                  AccountStatusActive,
		AuthPolicy:              AuthPolicy{Version: 1, Mode: AuthModeSingleKey},
		FeatureFlags:            features,
		CreatedHeight:           createdHeight,
		LastActiveHeight:        createdHeight,
		LastStorageChargeHeight: createdHeight,
	}
	if err := ValidateAccountInvariant(account); err != nil {
		return AccountActivationResult{}, err
	}
	if err := s.store.SetAccount(account); err != nil {
		return AccountActivationResult{}, err
	}
	return AccountActivationResult{
		Account: account,
		Event: AccountActivatedEvent{
			Type:          EventTypeAccountActivated,
			AddressUser:   account.AddressUser,
			AddressRaw:    account.AddressRaw,
			AccountNumber: account.AccountNumber,
			Sequence:      account.Sequence,
			PubKeyHash:    PublicKeyHash(msg.PublicKey),
			Height:        createdHeight,
			FeePaid:       msg.FeePaid,
		},
	}, nil
}

func PublicKeyText(pubKey cryptotypes.PubKey) string {
	if pubKey == nil {
		return ""
	}
	return pubKey.Type() + ":" + hex.EncodeToString(pubKey.Bytes())
}

func PublicKeyHash(pubKey cryptotypes.PubKey) string {
	if pubKey == nil {
		return ""
	}
	sum := sha256.Sum256(pubKey.Bytes())
	return hex.EncodeToString(sum[:])
}
