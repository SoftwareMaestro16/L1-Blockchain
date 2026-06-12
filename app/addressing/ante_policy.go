package addressing

import (
	"fmt"
	"reflect"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AddressPolicyRole string

const (
	AddressPolicyRoleSigner		AddressPolicyRole	= "signer"
	AddressPolicyRoleRecipient	AddressPolicyRole	= "recipient"
	AddressPolicyRoleAdmin		AddressPolicyRole	= "admin"
	AddressPolicyRoleAuthority	AddressPolicyRole	= "authority"
)

func ValidateAnteAddressPolicy(tx sdk.Tx) error {
	for i, msg := range tx.GetMsgs() {
		if err := ValidateMsgAddressPolicy(msg); err != nil {
			return fmt.Errorf("message %d: %w", i, err)
		}
	}
	return nil
}

func ValidateMsgAddressPolicy(msg sdk.Msg) error {
	if msg == nil {
		return nil
	}
	return validateAddressPolicyValue(reflect.ValueOf(msg), "")
}

func ValidateAddressForRole(role AddressPolicyRole, field, text string) error {
	switch role {
	case AddressPolicyRoleSigner:
		return ValidateUserSignerAddress(text)
	case AddressPolicyRoleRecipient:
		return ValidateUserRecipientAddress(text)
	case AddressPolicyRoleAdmin:
		return ValidateUserAdminAddress(field, text)
	case AddressPolicyRoleAuthority:
		return ValidateTxAuthorityAddress(field, text)
	default:
		return nil
	}
}

func validateAddressPolicyValue(value reflect.Value, path string) error {
	if !value.IsValid() {
		return nil
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Struct:
		typ := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}
			nextPath := field.Name
			if path != "" {
				nextPath = path + "." + field.Name
			}
			role, isAddress := roleForField(field.Name)
			if isAddress && value.Field(i).Kind() == reflect.String {
				if err := ValidateAddressForRole(role, nextPath, value.Field(i).String()); err != nil {
					return err
				}
				continue
			}
			if err := validateAddressPolicyValue(value.Field(i), nextPath); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := validateAddressPolicyValue(value.Index(i), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func roleForField(name string) (AddressPolicyRole, bool) {
	normalized := strings.ToLower(name)
	switch {
	case normalized == "authority" || strings.HasSuffix(normalized, "authority"):
		return AddressPolicyRoleAuthority, true
	case normalized == "admin" || strings.HasSuffix(normalized, "admin"):
		return AddressPolicyRoleAdmin, true
	case strings.Contains(normalized, "recipient") ||
		normalized == "toaddress" ||
		strings.HasSuffix(normalized, "toaddress") ||
		strings.Contains(normalized, "withdrawaddress"):
		return AddressPolicyRoleRecipient, true
	case strings.Contains(normalized, "signer") ||
		strings.Contains(normalized, "sender") ||
		normalized == "fromaddress" ||
		strings.HasSuffix(normalized, "fromaddress") ||
		strings.Contains(normalized, "payer"):
		return AddressPolicyRoleSigner, true
	default:
		return "", false
	}
}
