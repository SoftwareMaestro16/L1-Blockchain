package params

import (
	"errors"
	"fmt"
	"strings"
)

const (
	ChainIDMaxLength = 64
)

// ValidateAetraChainID enforces the launch/testnet chain-id naming policy.
// IDs are lower-case ASCII tokens with dash-separated segments and must start
// with "aetra-". Examples: aetra-local-1, aetra-testnet-1, aetra-mainnet-1.
func ValidateAetraChainID(chainID string) error {
	chainID = strings.TrimSpace(chainID)
	if chainID == "" {
		return errors.New("chain-id is required")
	}
	if len(chainID) > ChainIDMaxLength {
		return fmt.Errorf("chain-id must not exceed %d bytes", ChainIDMaxLength)
	}
	if chainID != strings.ToLower(chainID) {
		return errors.New("chain-id may contain only lower-case letters, digits, and dashes")
	}
	if !strings.HasPrefix(chainID, "aetra-") {
		return errors.New("chain-id must start with aetra-")
	}
	if strings.Contains(chainID, "--") || strings.HasSuffix(chainID, "-") {
		return errors.New("chain-id must use non-empty dash-separated segments")
	}
	for _, ch := range chainID {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			continue
		}
		return errors.New("chain-id may contain only lower-case letters, digits, and dashes")
	}
	return nil
}

func ValidateAetraTestnetChainID(chainID string) error {
	if err := ValidateAetraChainID(chainID); err != nil {
		return err
	}
	if !strings.Contains(chainID, "-testnet-") && !strings.Contains(chainID, "-local-") && !strings.Contains(chainID, "-preflight-") {
		return errors.New("testnet chain-id must include -testnet-, -local-, or -preflight-")
	}
	return nil
}
