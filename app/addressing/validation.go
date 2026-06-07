package addressing

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ParseUserAddress(field, text string) (sdk.AccAddress, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty address string is not allowed")
	}
	if !strings.HasPrefix(text, UserFriendlyPrefix) {
		return nil, fmt.Errorf("%s must use AE user-facing address format", field)
	}
	addr, err := ParseAccAddress(text)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", field, err)
	}
	if IsZeroAccAddress(addr) {
		return nil, fmt.Errorf("%s must not be zero address", field)
	}
	return addr, nil
}

func ValidateUserAddress(field, text string) error {
	_, err := ParseUserAddress(field, text)
	return err
}

func RejectZeroAddress(field string, bz []byte) error {
	if IsZero(bz) {
		return fmt.Errorf("%s must not be zero address", field)
	}
	return nil
}

func ParseAuthorityAddress(field, text string) (sdk.AccAddress, error) {
	addr, err := ParseAccAddress(text)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", field, err)
	}
	if IsZeroAccAddress(addr) {
		return nil, fmt.Errorf("%s must not be zero address", field)
	}
	return addr, nil
}

func ValidateAuthorityAddress(field, text string) error {
	_, err := ParseAuthorityAddress(field, text)
	return err
}

func ParseContractAddress(field, text string) (sdk.AccAddress, error) {
	return ParseUserAddress(field, text)
}

func ValidateContractAddress(field, text string) error {
	_, err := ParseContractAddress(field, text)
	return err
}

func ParseOptionalAdminAddress(field, text string) (sdk.AccAddress, bool, error) {
	if strings.TrimSpace(text) == "" {
		return nil, false, nil
	}
	addr, err := ParseUserAddress(field, text)
	if err != nil {
		return nil, false, err
	}
	return addr, true, nil
}

func ValidateOptionalAdminAddress(field, text string) error {
	_, _, err := ParseOptionalAdminAddress(field, text)
	return err
}

func ValidateNoZeroFactoryDenomAdmin(field, denom string) error {
	if !strings.HasPrefix(denom, "factory/") {
		return nil
	}
	parts := strings.SplitN(denom, "/", 3)
	if len(parts) < 2 || parts[1] == "" {
		return nil
	}
	return ValidateUserAddress(field+" factory admin", parts[1])
}
