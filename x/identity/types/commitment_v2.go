package types

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultDomainCommitmentVersion = uint64(1)
	DomainRegistrationIntent       = "register"
)

type DomainCommitmentV2 struct {
	CommitmentHash    string
	Committer         sdk.AccAddress
	CreatedAtHeight   uint64
	ExpiresAtHeight   uint64
	Deposit           sdkmath.Int
	CommitmentVersion uint64
	SaltHashOptional  string
}

type DomainCommitmentV2Context struct {
	CurrentHeight            uint64
	RevealWindowBlocks       uint64
	ChainID                  string
	ModuleVersion            uint64
	RegistrationIntent       string
	RevealedCommitmentHashes []string
	RefundValidRevealDeposit bool
	RefundCleanExpiryDeposit bool
}

type DomainCommitmentRefundReason string

const (
	DomainCommitmentRefundValidReveal DomainCommitmentRefundReason = "valid_reveal"
	DomainCommitmentRefundCleanExpiry DomainCommitmentRefundReason = "clean_expiry"
)

func NewDomainCommitmentV2(name string, committer sdk.AccAddress, salt string, deposit sdkmath.Int, createdAtHeight uint64, ctx DomainCommitmentV2Context, includeSaltHash bool) (DomainCommitmentV2, error) {
	if err := validateDomainCommitmentV2Context(ctx); err != nil {
		return DomainCommitmentV2{}, err
	}
	commitmentHash, err := ComputeDomainCommitmentV2Hash(name, committer, salt, ctx.ChainID, ctx.ModuleVersion, ctx.RegistrationIntent)
	if err != nil {
		return DomainCommitmentV2{}, err
	}
	if createdAtHeight == 0 {
		return DomainCommitmentV2{}, errors.New("identity v2 commitment created_at_height is required")
	}
	commitment := DomainCommitmentV2{
		CommitmentHash:    commitmentHash,
		Committer:         cloneSpecAddress(committer),
		CreatedAtHeight:   createdAtHeight,
		ExpiresAtHeight:   createdAtHeight + ctx.RevealWindowBlocks,
		Deposit:           deposit,
		CommitmentVersion: ctx.ModuleVersion,
	}
	if includeSaltHash {
		commitment.SaltHashOptional = ComputeDomainCommitmentV2SaltHash(salt)
	}
	return commitment, ValidateDomainCommitmentV2(commitment, ctx)
}

func ValidateDomainCommitmentV2(commitment DomainCommitmentV2, ctx DomainCommitmentV2Context) error {
	if err := validateDomainCommitmentV2Context(ctx); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 commitment hash", commitment.CommitmentHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 commitment committer", commitment.Committer); err != nil {
		return err
	}
	if commitment.CreatedAtHeight == 0 {
		return errors.New("identity v2 commitment created_at_height is required")
	}
	if commitment.ExpiresAtHeight <= commitment.CreatedAtHeight {
		return errors.New("identity v2 commitment expires_at_height must be after created_at_height")
	}
	if ctx.RevealWindowBlocks != 0 && commitment.ExpiresAtHeight != commitment.CreatedAtHeight+ctx.RevealWindowBlocks {
		return errors.New("identity v2 commitment expiry must match configured reveal window")
	}
	if commitment.Deposit.IsNegative() {
		return errors.New("identity v2 commitment deposit must not be negative")
	}
	if commitment.CommitmentVersion != ctx.ModuleVersion {
		return errors.New("identity v2 commitment version mismatch")
	}
	if commitment.SaltHashOptional != "" {
		if err := validateHexHash("identity v2 commitment salt hash", commitment.SaltHashOptional); err != nil {
			return err
		}
	}
	if stringInSet(commitment.CommitmentHash, ctx.RevealedCommitmentHashes) {
		return errors.New("identity v2 commitment hash already revealed")
	}
	return nil
}

func ValidateDomainCommitmentV2Reveal(commitment DomainCommitmentV2, name string, salt string, ctx DomainCommitmentV2Context) error {
	if err := ValidateDomainCommitmentV2(commitment, ctx); err != nil {
		return err
	}
	if ctx.CurrentHeight == 0 {
		return errors.New("identity v2 reveal height is required")
	}
	if ctx.CurrentHeight > commitment.ExpiresAtHeight {
		return errors.New("identity v2 commitment reveal window expired")
	}
	expected, err := ComputeDomainCommitmentV2Hash(name, commitment.Committer, salt, ctx.ChainID, commitment.CommitmentVersion, ctx.RegistrationIntent)
	if err != nil {
		return err
	}
	if commitment.CommitmentHash != expected {
		return errors.New("identity v2 commitment reveal hash mismatch")
	}
	if commitment.SaltHashOptional != "" && commitment.SaltHashOptional != ComputeDomainCommitmentV2SaltHash(salt) {
		return errors.New("identity v2 commitment salt hash mismatch")
	}
	return nil
}

func DomainCommitmentV2Refundable(commitment DomainCommitmentV2, reason DomainCommitmentRefundReason, ctx DomainCommitmentV2Context) (bool, error) {
	if err := validateDomainCommitmentV2Context(ctx); err != nil {
		return false, err
	}
	if err := validateHexHash("identity v2 commitment hash", commitment.CommitmentHash); err != nil {
		return false, err
	}
	if stringInSet(commitment.CommitmentHash, ctx.RevealedCommitmentHashes) {
		return false, errors.New("identity v2 revealed commitment cannot be refunded again")
	}
	switch reason {
	case DomainCommitmentRefundValidReveal:
		if ctx.CurrentHeight == 0 || ctx.CurrentHeight > commitment.ExpiresAtHeight {
			return false, errors.New("identity v2 valid reveal refund requires an unexpired commitment")
		}
		return ctx.RefundValidRevealDeposit, nil
	case DomainCommitmentRefundCleanExpiry:
		if ctx.CurrentHeight <= commitment.ExpiresAtHeight {
			return false, errors.New("identity v2 clean expiry refund requires expired commitment")
		}
		return ctx.RefundCleanExpiryDeposit, nil
	default:
		return false, fmt.Errorf("unsupported identity v2 commitment refund reason %q", reason)
	}
}

func ComputeDomainCommitmentV2Hash(name string, committer sdk.AccAddress, salt string, chainID string, moduleVersion uint64, registrationIntent string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 commitment committer", committer); err != nil {
		return "", err
	}
	if strings.TrimSpace(salt) == "" {
		return "", errors.New("identity v2 commitment salt is required")
	}
	if strings.TrimSpace(chainID) == "" {
		return "", errors.New("identity v2 commitment chain id is required")
	}
	if moduleVersion == 0 {
		return "", errors.New("identity v2 commitment module version is required")
	}
	if strings.TrimSpace(registrationIntent) == "" {
		return "", errors.New("identity v2 commitment registration intent is required")
	}
	return identityHash(
		"identity-v2-domain-commitment",
		normalized,
		string(committer),
		salt,
		chainID,
		fmt.Sprintf("%020d", moduleVersion),
		registrationIntent,
	), nil
}

func ComputeDomainCommitmentV2SaltHash(salt string) string {
	return identityHash("identity-v2-domain-commitment-salt", salt)
}

func validateDomainCommitmentV2Context(ctx DomainCommitmentV2Context) error {
	if ctx.RevealWindowBlocks == 0 {
		return errors.New("identity v2 commitment reveal window is required")
	}
	if strings.TrimSpace(ctx.ChainID) == "" {
		return errors.New("identity v2 commitment chain id is required")
	}
	if ctx.ModuleVersion == 0 {
		return errors.New("identity v2 commitment module version is required")
	}
	if strings.TrimSpace(ctx.RegistrationIntent) == "" {
		return errors.New("identity v2 commitment registration intent is required")
	}
	for _, hash := range ctx.RevealedCommitmentHashes {
		if err := validateHexHash("identity v2 revealed commitment hash", hash); err != nil {
			return err
		}
	}
	return nil
}

func stringInSet(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
