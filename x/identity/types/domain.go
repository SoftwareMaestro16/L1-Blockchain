package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	DomainTLD		= ".aet"
	MinDomainLabelBytes	= 1
	MaxDomainLabelBytes	= 64
)

type DomainStatus string

const (
	DomainStatusAuction	DomainStatus	= "auction"
	DomainStatusActive	DomainStatus	= "active"
	DomainStatusExpired	DomainStatus	= "expired"
)

type DomainRecord struct {
	Name		string
	TLD		string
	Owner		sdk.AccAddress
	Resolver	sdk.AccAddress
	ExpiryUnix	int64
	NFTItemID	string
	Status		DomainStatus
	CreatedAtUnix	int64
	UpdatedAtUnix	int64
}

func NormalizeDomainName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", errors.New("domain name is required")
	}
	lower := strings.ToLower(trimmed)
	if strings.HasSuffix(lower, DomainTLD) {
		lower = strings.TrimSuffix(lower, DomainTLD)
	}
	if lower == "" {
		return "", errors.New("domain label is required")
	}
	return lower, nil
}

func ValidateDomainName(name string) error {
	normalized, err := NormalizeDomainName(name)
	if err != nil {
		return err
	}
	if name != normalized && name != normalized+DomainTLD {
		return fmt.Errorf("domain name must be normalized lowercase ASCII")
	}
	if len(normalized) < MinDomainLabelBytes {
		return fmt.Errorf("domain label must be at least %d byte", MinDomainLabelBytes)
	}
	if len(normalized) > MaxDomainLabelBytes {
		return fmt.Errorf("domain label must be at most %d bytes", MaxDomainLabelBytes)
	}
	return validateDomainLabel(normalized)
}

func validateDomainLabel(label string) error {
	if label == "" {
		return errors.New("domain label is required")
	}
	if len(label) > MaxDomainLabelBytes {
		return fmt.Errorf("domain label must be at most %d bytes", MaxDomainLabelBytes)
	}
	for i := 0; i < len(label); i++ {
		c := label[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			continue
		}
		return fmt.Errorf("domain label contains unsupported character %q", c)
	}
	return nil
}

func ValidateDomainRecord(record DomainRecord) error {
	if err := ValidateDomainName(record.Name); err != nil {
		return fmt.Errorf("invalid domain name: %w", err)
	}
	if record.TLD != DomainTLD {
		return fmt.Errorf("domain tld must be %q", DomainTLD)
	}
	if len(record.Owner) == 0 {
		return errors.New("domain owner is required")
	}
	if err := addressing.RejectZeroAddress("domain owner", record.Owner); err != nil {
		return err
	}
	if len(record.Resolver) > 0 {
		if err := addressing.RejectZeroAddress("domain resolver", record.Resolver); err != nil {
			return err
		}
	}
	if !IsDomainStatus(record.Status) {
		return fmt.Errorf("invalid domain status %q", record.Status)
	}
	if record.ExpiryUnix <= 0 {
		return errors.New("domain expiry must be positive")
	}
	if record.CreatedAtUnix < 0 || record.UpdatedAtUnix < 0 {
		return errors.New("domain timestamps must be non-negative")
	}
	if record.UpdatedAtUnix < record.CreatedAtUnix {
		return errors.New("domain updated_at must not be before created_at")
	}
	if record.Status == DomainStatusActive && strings.TrimSpace(record.NFTItemID) == "" {
		return errors.New("active domain must have nft item id")
	}
	return nil
}

func IsDomainStatus(status DomainStatus) bool {
	switch status {
	case DomainStatusAuction, DomainStatusActive, DomainStatusExpired:
		return true
	default:
		return false
	}
}
