package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

type MsgActivateAccount struct {
	AddressUser	string			`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw	string			`protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	PublicKeyType	string			`protobuf:"bytes,3,opt,name=public_key_type,json=publicKeyType,proto3" json:"public_key_type,omitempty"`
	PublicKeyHex	string			`protobuf:"bytes,4,opt,name=public_key_hex,json=publicKeyHex,proto3" json:"public_key_hex,omitempty"`
	FeePaid		uint64			`protobuf:"varint,5,opt,name=fee_paid,json=feePaid,proto3" json:"fee_paid,omitempty"`
	PublicKey	cryptotypes.PubKey	`json:"-"`
}

func (m MsgActivateAccount) ValidateBasic() error {
	if err := addressing.ValidateNewUserAccountAddress("activation address", m.AddressUser); err != nil {
		return err
	}
	pubKey, err := m.EffectivePublicKey()
	if err != nil {
		return err
	}
	pair, err := addressing.DeriveAccountAddress(pubKey)
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

func (m MsgActivateAccount) EffectivePublicKey() (cryptotypes.PubKey, error) {
	if m.PublicKey != nil {
		return m.PublicKey, nil
	}
	publicKeyType := strings.TrimSpace(m.PublicKeyType)
	publicKeyHex := strings.TrimSpace(m.PublicKeyHex)
	if publicKeyType == "" || publicKeyHex == "" {
		return nil, fmt.Errorf("activation public key is required")
	}
	raw, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("activation public key hex is invalid: %w", err)
	}
	switch publicKeyType {
	case "secp256k1", "secp256k1.PubKey", "cosmos.crypto.secp256k1.PubKey", "/cosmos.crypto.secp256k1.PubKey":
		return &secp256k1.PubKey{Key: raw}, nil
	default:
		return nil, fmt.Errorf("unsupported activation public key type %q", publicKeyType)
	}
}

func ActivationAddressPair(pubKey cryptotypes.PubKey) (addressing.AddressPair, error) {
	return addressing.DeriveAccountAddress(pubKey)
}

const (
	ActivationInitialSequence	= uint64(0)
	EventTypeAccountActivated	= "AccountActivated"
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
	Type		string	`json:"type"`
	AddressUser	string	`json:"address_user"`
	AddressRaw	string	`json:"address_raw"`
	AccountNumber	uint64	`json:"account_number"`
	Sequence	uint64	`json:"sequence"`
	PubKeyHash	string	`json:"pubkey_hash"`
	Height		uint64	`json:"height"`
	FeePaid		uint64	`json:"fee_paid"`
}

type AccountActivationResult struct {
	Account	Account
	Event	AccountActivatedEvent
}

type AccountActivationStore interface {
	AccountByUser(userAddress string) (Account, bool, error)
	SetAccount(account Account) error
	NextAccountNumber() uint64
}

type AccountActivationService struct {
	store		AccountActivationStore
	feePolicy	ActivationFeePolicy
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
	pubKey, err := msg.EffectivePublicKey()
	if err != nil {
		return AccountActivationResult{}, err
	}
	pair, err := ActivationAddressPair(pubKey)
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
		Version:			CurrentAccountVersion,
		AddressUser:			pair.User,
		AddressRaw:			pair.Raw,
		PubKeys:			[]string{PublicKeyText(pubKey)},
		AccountNumber:			s.store.NextAccountNumber(),
		Sequence:			ActivationInitialSequence,
		Status:				AccountStatusActive,
		AuthPolicy:			AuthPolicy{Version: 1, Mode: AuthModeSingleKey},
		FeatureFlags:			features,
		CreatedHeight:			createdHeight,
		LastActiveHeight:		createdHeight,
		LastStorageChargeHeight:	createdHeight,
	}
	if err := ValidateAccountInvariant(account); err != nil {
		return AccountActivationResult{}, err
	}
	if err := s.store.SetAccount(account); err != nil {
		return AccountActivationResult{}, err
	}
	return AccountActivationResult{
		Account:	account,
		Event: AccountActivatedEvent{
			Type:		EventTypeAccountActivated,
			AddressUser:	account.AddressUser,
			AddressRaw:	account.AddressRaw,
			AccountNumber:	account.AccountNumber,
			Sequence:	account.Sequence,
			PubKeyHash:	PublicKeyHash(pubKey),
			Height:		createdHeight,
			FeePaid:	msg.FeePaid,
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
