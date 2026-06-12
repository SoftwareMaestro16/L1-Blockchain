package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxIDLength		= 64
	MaxNameLength		= 128
	MaxScopeLength		= 128
	MaxCapabilityLength	= 64
)

func validateZoneID(id string) error {
	text := strings.TrimSpace(id)
	if text != id || text == "" {
		return errors.New("aetracore zone id is required and must not have surrounding whitespace")
	}
	if len(text) > MaxIDLength {
		return fmt.Errorf("aetracore zone id must be <= %d bytes", MaxIDLength)
	}
	for i, r := range text {
		if i == 0 && (r < 'A' || r > 'Z') {
			return errors.New("aetracore zone id must start with A-Z")
		}
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			continue
		}
		return errors.New("aetracore zone id must contain only A-Z, 0-9, or underscore")
	}
	return nil
}

func validateToken(fieldName, value string, maxLen int) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxLen {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxLen)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func validateCapabilitiesForField(fieldName string, capabilities []string) error {
	var previous string
	seen := make(map[string]struct{}, len(capabilities))
	for i, capability := range capabilities {
		if err := validateToken(fieldName, capability, MaxCapabilityLength); err != nil {
			return err
		}
		if _, found := seen[capability]; found {
			return fmt.Errorf("duplicate %s %s", fieldName, capability)
		}
		seen[capability] = struct{}{}
		if i > 0 && previous >= capability {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
		previous = capability
	}
	return nil
}

func validateModuleName(value string) error {
	return validateToken("aetracore zone module name", value, MaxScopeLength)
}

func validatePolicyID(fieldName, value string) error {
	return validateToken(fieldName, value, MaxScopeLength)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
