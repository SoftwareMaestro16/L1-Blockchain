package types

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultDomainCommitmentVersion	= uint64(1)
	DefaultIdentityModuleName	= "x/identity"
	DefaultRegistrationClass	= "direct"
	DomainRegistrationIntent	= "register"
)

type DomainCommitmentV2 struct {
	CommitmentHash		string
	Committer		sdk.AccAddress
	CreatedAtHeight		uint64
	ExpiresAtHeight		uint64
	Deposit			sdkmath.Int
	CommitmentVersion	uint64
	SaltHashOptional	string
	ChainID			string
	ModuleName		string
	RegistrationClass	string
	MaxPrice		sdkmath.Int
}

type DomainCommitmentV2Context struct {
	CurrentHeight			uint64
	RevealWindowBlocks		uint64
	ChainID				string
	ModuleName			string
	ModuleVersion			uint64
	RegistrationIntent		string
	RegistrationClass		string
	MaxPrice			sdkmath.Int
	RevealedCommitmentHashes	[]string
	RefundValidRevealDeposit	bool
	RefundCleanExpiryDeposit	bool
}

type DomainCommitmentV2Preimage struct {
	ChainID			string
	ModuleName		string
	ModuleVersion		uint64
	NormalizedName		string
	Committer		sdk.AccAddress
	Salt			string
	RegistrationClass	string
	MaxPrice		sdkmath.Int
	ExpiryHeight		uint64
}

type DomainCommitmentRefundReason string

const (
	DomainCommitmentRefundValidReveal	DomainCommitmentRefundReason	= "valid_reveal"
	DomainCommitmentRefundCleanExpiry	DomainCommitmentRefundReason	= "clean_expiry"
)

func NewDomainCommitmentV2(name string, committer sdk.AccAddress, salt string, deposit sdkmath.Int, createdAtHeight uint64, ctx DomainCommitmentV2Context, includeSaltHash bool) (DomainCommitmentV2, error) {
	ctx = normalizeDomainCommitmentV2Context(ctx)
	if err := validateDomainCommitmentV2Context(ctx); err != nil {
		return DomainCommitmentV2{}, err
	}
	if createdAtHeight == 0 {
		return DomainCommitmentV2{}, errors.New("identity v2 commitment created_at_height is required")
	}
	expiresAtHeight := createdAtHeight + ctx.RevealWindowBlocks
	commitmentHash, err := ComputeDomainCommitmentV2PreimageHash(DomainCommitmentV2Preimage{
		ChainID:		ctx.ChainID,
		ModuleName:		ctx.ModuleName,
		ModuleVersion:		ctx.ModuleVersion,
		NormalizedName:		name,
		Committer:		committer,
		Salt:			salt,
		RegistrationClass:	ctx.RegistrationClass,
		MaxPrice:		ctx.MaxPrice,
		ExpiryHeight:		expiresAtHeight,
	})
	if err != nil {
		return DomainCommitmentV2{}, err
	}
	commitment := DomainCommitmentV2{
		CommitmentHash:		commitmentHash,
		Committer:		cloneSpecAddress(committer),
		CreatedAtHeight:	createdAtHeight,
		ExpiresAtHeight:	expiresAtHeight,
		Deposit:		deposit,
		CommitmentVersion:	ctx.ModuleVersion,
		ChainID:		ctx.ChainID,
		ModuleName:		ctx.ModuleName,
		RegistrationClass:	ctx.RegistrationClass,
		MaxPrice:		ctx.MaxPrice,
	}
	if includeSaltHash {
		commitment.SaltHashOptional = ComputeDomainCommitmentV2SaltHash(salt)
	}
	return commitment, ValidateDomainCommitmentV2(commitment, ctx)
}

func ValidateDomainCommitmentV2(commitment DomainCommitmentV2, ctx DomainCommitmentV2Context) error {
	ctx = normalizeDomainCommitmentV2Context(ctx)
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
	if strings.TrimSpace(commitment.ChainID) == "" {
		return errors.New("identity v2 commitment chain id is required")
	}
	if strings.TrimSpace(commitment.ModuleName) == "" {
		return errors.New("identity v2 commitment module name is required")
	}
	if strings.TrimSpace(commitment.RegistrationClass) == "" {
		return errors.New("identity v2 commitment registration class is required")
	}
	if commitment.MaxPrice.IsNil() || commitment.MaxPrice.IsNegative() {
		return errors.New("identity v2 commitment max price must not be negative")
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
	ctx = normalizeDomainCommitmentV2Context(ctx)
	if err := ValidateDomainCommitmentV2(commitment, ctx); err != nil {
		return err
	}
	if ctx.CurrentHeight == 0 {
		return errors.New("identity v2 reveal height is required")
	}
	if ctx.CurrentHeight < commitment.CreatedAtHeight {
		return errors.New("identity v2 reveal height is before commitment")
	}
	if ctx.CurrentHeight > commitment.ExpiresAtHeight {
		return errors.New("identity v2 commitment reveal window expired")
	}
	expected, err := ComputeDomainCommitmentV2PreimageHash(DomainCommitmentV2Preimage{
		ChainID:		ctx.ChainID,
		ModuleName:		ctx.ModuleName,
		ModuleVersion:		commitment.CommitmentVersion,
		NormalizedName:		name,
		Committer:		commitment.Committer,
		Salt:			salt,
		RegistrationClass:	ctx.RegistrationClass,
		MaxPrice:		ctx.MaxPrice,
		ExpiryHeight:		commitment.ExpiresAtHeight,
	})
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
	ctx = normalizeDomainCommitmentV2Context(ctx)
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
	return ComputeDomainCommitmentV2PreimageHash(DomainCommitmentV2Preimage{
		ChainID:		chainID,
		ModuleName:		DefaultIdentityModuleName,
		ModuleVersion:		moduleVersion,
		NormalizedName:		name,
		Committer:		committer,
		Salt:			salt,
		RegistrationClass:	registrationIntent,
		MaxPrice:		sdkmath.ZeroInt(),
		ExpiryHeight:		0,
	})
}

func ComputeDomainCommitmentV2PreimageHash(preimage DomainCommitmentV2Preimage) (string, error) {
	normalized, err := NormalizeAETDomain(preimage.NormalizedName)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity v2 commitment committer", preimage.Committer); err != nil {
		return "", err
	}
	if strings.TrimSpace(preimage.Salt) == "" {
		return "", errors.New("identity v2 commitment salt is required")
	}
	if strings.TrimSpace(preimage.ChainID) == "" {
		return "", errors.New("identity v2 commitment chain id is required")
	}
	if strings.TrimSpace(preimage.ModuleName) == "" {
		return "", errors.New("identity v2 commitment module name is required")
	}
	if preimage.ModuleVersion == 0 {
		return "", errors.New("identity v2 commitment module version is required")
	}
	if strings.TrimSpace(preimage.RegistrationClass) == "" {
		return "", errors.New("identity v2 commitment registration class is required")
	}
	if preimage.MaxPrice.IsNil() || preimage.MaxPrice.IsNegative() {
		return "", errors.New("identity v2 commitment max price must not be negative")
	}
	return identityHash(
		"identity-v2-domain-commitment",
		strings.TrimSpace(preimage.ChainID),
		strings.TrimSpace(preimage.ModuleName),
		fmt.Sprintf("%020d", preimage.ModuleVersion),
		normalized,
		string(preimage.Committer),
		preimage.Salt,
		strings.TrimSpace(preimage.RegistrationClass),
		preimage.MaxPrice.String(),
		fmt.Sprintf("%020d", preimage.ExpiryHeight),
	), nil
}

func ComputeDomainCommitmentV2SaltHash(salt string) string {
	return identityHash("identity-v2-domain-commitment-salt", salt)
}

func NewUsedDomainCommitment(commit DomainCommit, revealedHeight uint64) UsedDomainCommitment {
	return UsedDomainCommitment{
		CommitmentHash:		commit.CommitmentHash,
		Name:			commit.Name,
		Owner:			cloneSpecAddress(commit.Owner),
		RevealedHeight:		revealedHeight,
		ExpiresHeight:		commit.ExpiresHeight,
		ModuleName:		DefaultIdentityModuleName,
		ModuleVersion:		DefaultDomainCommitmentVersion,
		RegistrationClass:	DefaultRegistrationClass,
		MaxPrice:		"0",
	}
}

func validateDomainCommitmentV2Context(ctx DomainCommitmentV2Context) error {
	ctx = normalizeDomainCommitmentV2Context(ctx)
	if ctx.RevealWindowBlocks == 0 {
		return errors.New("identity v2 commitment reveal window is required")
	}
	if strings.TrimSpace(ctx.ChainID) == "" {
		return errors.New("identity v2 commitment chain id is required")
	}
	if ctx.ModuleVersion == 0 {
		return errors.New("identity v2 commitment module version is required")
	}
	if strings.TrimSpace(ctx.ModuleName) == "" {
		return errors.New("identity v2 commitment module name is required")
	}
	if strings.TrimSpace(ctx.RegistrationIntent) == "" {
		return errors.New("identity v2 commitment registration intent is required")
	}
	if strings.TrimSpace(ctx.RegistrationClass) == "" {
		return errors.New("identity v2 commitment registration class is required")
	}
	if ctx.MaxPrice.IsNil() || ctx.MaxPrice.IsNegative() {
		return errors.New("identity v2 commitment max price must not be negative")
	}
	for _, hash := range ctx.RevealedCommitmentHashes {
		if err := validateHexHash("identity v2 revealed commitment hash", hash); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDomainCommitmentV2Context(ctx DomainCommitmentV2Context) DomainCommitmentV2Context {
	ctx.ChainID = strings.TrimSpace(ctx.ChainID)
	ctx.ModuleName = strings.TrimSpace(ctx.ModuleName)
	if ctx.ModuleName == "" {
		ctx.ModuleName = DefaultIdentityModuleName
	}
	if ctx.RegistrationIntent == "" {
		ctx.RegistrationIntent = DomainRegistrationIntent
	}
	ctx.RegistrationClass = strings.TrimSpace(ctx.RegistrationClass)
	if ctx.RegistrationClass == "" {
		ctx.RegistrationClass = strings.TrimSpace(ctx.RegistrationIntent)
	}
	if ctx.RegistrationClass == "" {
		ctx.RegistrationClass = DefaultRegistrationClass
	}
	if ctx.MaxPrice.IsNil() {
		ctx.MaxPrice = sdkmath.ZeroInt()
	}
	return ctx
}

func stringInSet(value string, values []string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
