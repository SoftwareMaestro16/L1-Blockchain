package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	AWCE1Version	= uint32(1)

	BIP44Purpose		= uint32(44)
	BIP44CoinType		= uint32(118)
	BIP44Account		= uint32(0)
	BIP44Change		= uint32(0)
	BIP44AddressIndex	= uint32(0)

	BIP44FullPath	= "m/44'/118'/0'/0/0"

	SignDocTypeURL	= "/cosmos.tx.v1beta1.Tx"

	AWCE1FeatureKeyRotationV2	= "key_rotation_v2"
	AWCE1FeatureStorageRentV2	= "storage_rent_v2"
	AWCE1FeatureRecoveryPolicyV2	= "recovery_policy_v2"
	AWCE1FeatureAuthPolicyV2	= "auth_policy_v2"

	MaxAWCE1FeatureFlags		= 16
	MaxAWCE1SecretPatternBytes	= 256
)

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)private.?key`),
		regexp.MustCompile(`(?i)mnemonic`),
		regexp.MustCompile(`(?i)seed.?phrase`),
		regexp.MustCompile(`(?i)secret`),
		regexp.MustCompile(`(?i)password`),
		regexp.MustCompile(`(?i)passphrase`),
		regexp.MustCompile(`(?i)privkey`),
		regexp.MustCompile(`(?i)priv_key`),
	}
)

type AWCE1WalletProfile struct {
	Version			uint32
	SpecVersion		uint32
	AddressPair		addressing.AddressPair
	BIP44Path		string
	KeyType			string
	Features		[]string
	AccountStatus		string
	ActivationHeight	uint64
}

func NewAWCE1WalletProfile(pair addressing.AddressPair) *AWCE1WalletProfile {
	features, _ := DefaultFeatureFlags(AccountVersionV2)
	return &AWCE1WalletProfile{
		Version:		AWCE1Version,
		SpecVersion:		AWCE1Version,
		AddressPair:		pair,
		BIP44Path:		BIP44FullPath,
		KeyType:		"secp256k1",
		Features:		features,
		AccountStatus:		AccountStatusInactive,
		ActivationHeight:	0,
	}
}

func (p *AWCE1WalletProfile) Validate() error {
	if p.Version == 0 {
		return fmt.Errorf("AWCE1: version must be > 0")
	}
	if p.AddressPair.User == "" || p.AddressPair.Raw == "" {
		return fmt.Errorf("AWCE1: address pair must have both User and Raw")
	}
	if err := ValidateAddressPairConsistency(p.AddressPair); err != nil {
		return fmt.Errorf("AWCE1: address pair inconsistent: %w", err)
	}
	if err := ValidateNoSecretLikeText(p.BIP44Path); err != nil {
		return fmt.Errorf("AWCE1: BIP44 path: %w", err)
	}
	if p.KeyType != "secp256k1" {
		return fmt.Errorf("AWCE1: unsupported key type %q; only secp256k1 is supported for AWCE-1", p.KeyType)
	}
	if len(p.Features) > MaxAWCE1FeatureFlags {
		return fmt.Errorf("AWCE1: too many feature flags (%d > %d)", len(p.Features), MaxAWCE1FeatureFlags)
	}
	return nil
}

func ValidateAddressPairConsistency(pair addressing.AddressPair) error {
	if pair.User == "" {
		return fmt.Errorf("user-facing address (AE...) must not be empty")
	}
	if pair.Raw == "" {
		return fmt.Errorf("raw internal address (4:...) must not be empty")
	}
	if !strings.HasPrefix(pair.User, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("user-facing address must start with %s", addressing.UserFriendlyPrefix)
	}
	if !strings.HasPrefix(pair.Raw, addressing.RawPrefix) {
		return fmt.Errorf("raw internal address must start with %s", addressing.RawPrefix)
	}
	userBytes, err := addressing.Parse(pair.User)
	if err != nil {
		return fmt.Errorf("user address parse error: %w", err)
	}
	rawBytes, err := addressing.Parse(pair.Raw)
	if err != nil {
		return fmt.Errorf("raw address parse error: %w", err)
	}
	if len(userBytes) != len(rawBytes) {
		return fmt.Errorf("address byte lengths differ: user=%d raw=%d", len(userBytes), len(rawBytes))
	}
	for i := range userBytes {
		if userBytes[i] != rawBytes[i] {
			return fmt.Errorf("AE and 4: addresses do not map to same identity")
		}
	}
	return nil
}

func DeriveAWCE1AddressPairFromPublicKey(pubKeyHex string, keyType string) (addressing.AddressPair, error) {
	if keyType != "secp256k1" && keyType != "secp256k1.PubKey" && keyType != "cosmos.crypto.secp256k1.PubKey" {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: unsupported key type %q", keyType)
	}
	raw, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: invalid public key hex: %w", err)
	}
	if len(raw) != 33 {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: secp256k1 public key must be 33 bytes, got %d", len(raw))
	}
	hash := sha256.Sum256(raw)
	addressBytes := hash[:20]
	padded := make([]byte, 32)
	copy(padded[12:], addressBytes)
	rawAddr := addressing.Format(padded)
	userAddr, err := addressing.FormatUserFriendly(padded)
	if err != nil {
		return addressing.AddressPair{}, fmt.Errorf("AWCE1: format user address: %w", err)
	}
	return addressing.AddressPair{
		Role:	addressing.AddressRoleAccount,
		User:	userAddr,
		Raw:	rawAddr,
	}, nil
}

type KeyRotationRequest struct {
	AccountAddress	string
	NewAuthPolicy	AuthPolicy
	RotationHeight	uint64
	Justification	string
}

type KeyRotationResult struct {
	AccountAddress		string
	PreviousKeyCount	int
	NewKeyCount		int
	AddressPreserved	bool
	RotationHeight		uint64
}

func ValidateKeyRotation(account Account, request KeyRotationRequest) (*KeyRotationResult, error) {
	if err := ValidateAddressPairConsistency(addressing.AddressPair{
		Role:	addressing.AddressRoleAccount,
		User:	account.AddressUser,
		Raw:	account.AddressRaw,
	}); err != nil {
		return nil, fmt.Errorf("AWCE1 key rotation: address pair invalid: %w", err)
	}
	if err := ValidateNoSecretLikeText(request.Justification); err != nil {
		return nil, fmt.Errorf("AWCE1 key rotation: justification: %w", err)
	}
	if request.RotationHeight < account.CreatedHeight {
		return nil, fmt.Errorf("AWCE1 key rotation: rotation height %d before account creation %d",
			request.RotationHeight, account.CreatedHeight)
	}
	return &KeyRotationResult{
		AccountAddress:		account.AddressUser,
		PreviousKeyCount:	len(account.AuthPolicy.Keys),
		NewKeyCount:		len(request.NewAuthPolicy.Keys),
		AddressPreserved:	true,
		RotationHeight:		request.RotationHeight,
	}, nil
}

type AccountLifecycleTransition struct {
	From	string
	To	string
}

var ValidLifecycleTransitions = []AccountLifecycleTransition{
	{AccountStatusInactive, AccountStatusActive},
	{AccountStatusActive, AccountStatusFrozen},
	{AccountStatusActive, AccountStatusRecovered},
	{AccountStatusActive, AccountStatusArchived},
	{AccountStatusFrozen, AccountStatusActive},
	{AccountStatusFrozen, AccountStatusArchived},
	{AccountStatusRecovered, AccountStatusActive},
	{AccountStatusRecovered, AccountStatusArchived},
	{AccountStatusArchived, AccountStatusClosed},
}

func ValidateLifecycleTransition(from, to string) error {
	if from == to {
		return nil
	}
	for _, t := range ValidLifecycleTransitions {
		if t.From == from && t.To == to {
			return nil
		}
	}
	return fmt.Errorf("AWCE1: invalid lifecycle transition %s -> %s", from, to)
}

func CanActivate(status string) bool {
	return status == AccountStatusInactive
}

func CanTransact(status string) bool {
	return status == AccountStatusActive || status == AccountStatusRecovered
}

func IsTerminalStatus(status string) bool {
	return status == AccountStatusClosed
}

var (
	ErrAccountAlreadyActive		= errors.New("AWCE1: account already active — duplicate activation rejected")
	ErrAccountNotFound		= errors.New("AWCE1: account not found")
	ErrAccountInactive		= errors.New("AWCE1: account is inactive")
	ErrAccountFrozen		= errors.New("AWCE1: account is frozen")
	ErrAccountClosed		= errors.New("AWCE1: account is closed")
	ErrInvalidLifecycleState	= errors.New("AWCE1: invalid account lifecycle state")
	ErrAddressMismatch		= errors.New("AWCE1: address pair mismatch")
	ErrSecretInAccount		= errors.New("AWCE1: secret-like text found in account")
)

func ValidateActivation(account Account, pair addressing.AddressPair) error {
	if account.Status == AccountStatusActive {
		return ErrAccountAlreadyActive
	}
	if account.Status != AccountStatusInactive && account.Status != "" {
		return fmt.Errorf("AWCE1: cannot activate account in status %q", account.Status)
	}
	if account.AddressUser != "" && account.AddressUser != pair.User {
		return ErrAddressMismatch
	}
	if account.AddressRaw != "" && account.AddressRaw != pair.Raw {
		return ErrAddressMismatch
	}
	if err := ValidateAccountNoSecrets(account); err != nil {
		return ErrSecretInAccount
	}
	return nil
}

func ActivateAccount(pair addressing.AddressPair, height uint64) (*Account, error) {
	if err := ValidateAddressPairConsistency(pair); err != nil {
		return nil, fmt.Errorf("AWCE1 activation: %w", err)
	}
	features, _ := DefaultFeatureFlags(CurrentAccountVersion)
	account := &Account{
		Version:		CurrentAccountVersion,
		AddressUser:		pair.User,
		AddressRaw:		pair.Raw,
		Sequence:		ActivationInitialSequence,
		Status:			AccountStatusActive,
		AuthPolicy:		DefaultAuthPolicy(),
		FeatureFlags:		features,
		CreatedHeight:		height,
		LastActiveHeight:	height,
		StorageRentDebt:	0,
	}
	return account, nil
}

func DefaultAuthPolicy() AuthPolicy {
	return AuthPolicy{
		Version:	1,
		Mode:		"single_key",
		Threshold:	1,
	}
}

type CosmosSignDoc struct {
	AccountNumber	uint64		`json:"account_number"`
	ChainID		string		`json:"chain_id"`
	Fee		CosmosFee	`json:"fee"`
	Memo		string		`json:"memo"`
	Msgs		[]CosmosMsg	`json:"msgs"`
	Sequence	uint64		`json:"sequence"`
}

type CosmosFee struct {
	Amount	[]CosmosCoin	`json:"amount"`
	Gas	string		`json:"gas"`
}

type CosmosCoin struct {
	Denom	string	`json:"denom"`
	Amount	string	`json:"amount"`
}

type CosmosMsg struct {
	TypeURL	string	`json:"type_url"`
	Value	string	`json:"value"`
}

func NewCosmosSignDoc(accountNumber, sequence uint64, chainID string, fee CosmosFee, msgs []CosmosMsg, memo string) *CosmosSignDoc {
	return &CosmosSignDoc{
		AccountNumber:	accountNumber,
		ChainID:	chainID,
		Fee:		fee,
		Memo:		memo,
		Msgs:		msgs,
		Sequence:	sequence,
	}
}

func (d *CosmosSignDoc) SignBytes() []byte {
	return []byte(fmt.Sprintf(`{"account_number":"%d","chain_id":"%s","sequence":"%d"}`,
		d.AccountNumber, d.ChainID, d.Sequence))
}

func ValidateCosmosSignDoc(doc *CosmosSignDoc) error {
	if doc.ChainID == "" {
		return fmt.Errorf("AWCE1: SignDoc chain_id must not be empty")
	}
	if len(doc.Msgs) == 0 {
		return fmt.Errorf("AWCE1: SignDoc must have at least one message")
	}
	for i, msg := range doc.Msgs {
		if msg.TypeURL == "" {
			return fmt.Errorf("AWCE1: SignDoc msg[%d] has empty type_url", i)
		}
	}
	return nil
}

type StorageRentDebt struct {
	Account			string
	CurrentDebt		uint64
	LastChargeHeight	uint64
	AccumulatedRent		uint64
	GenesisFrozen		bool
	GenesisDebt		uint64
}

func (d StorageRentDebt) IsActiveDebt() bool {
	return d.CurrentDebt > 0
}

func (d StorageRentDebt) IsFrozen() bool {
	return d.GenesisFrozen || d.CurrentDebt > 0
}

func NewStorageRentDebt(account string) StorageRentDebt {
	return StorageRentDebt{
		Account:		account,
		CurrentDebt:		0,
		LastChargeHeight:	0,
		AccumulatedRent:	0,
		GenesisFrozen:		false,
		GenesisDebt:		0,
	}
}

func ValidateNoSecretLikeText(text string) error {
	for _, pattern := range secretPatterns {
		if pattern.MatchString(text) {
			return fmt.Errorf("AWCE1: text contains secret-like pattern matching %q", pattern.String())
		}
	}
	return nil
}

func ValidateAccountNoSecrets(account Account) error {
	for _, pk := range account.PubKeys {
		if err := ValidateNoSecretLikeText(pk); err != nil {
			return fmt.Errorf("public key field: %w", err)
		}
	}
	if err := ValidateNoSecretLikeText(account.ReputationID); err != nil {
		return fmt.Errorf("reputation ID: %w", err)
	}
	if err := ValidateNoSecretLikeText(account.Metadata.MetadataHash); err != nil {
		return fmt.Errorf("metadata hash: %w", err)
	}
	if err := ValidateNoSecretLikeText(account.Metadata.DisplayNameHash); err != nil {
		return fmt.Errorf("display name hash: %w", err)
	}
	if err := ValidateNoSecretLikeText(account.Metadata.DomainAlias); err != nil {
		return fmt.Errorf("domain alias: %w", err)
	}
	return nil
}
