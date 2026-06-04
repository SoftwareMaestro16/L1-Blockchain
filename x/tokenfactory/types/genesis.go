package types

import (
	"fmt"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultMinSubdenomLength = uint32(3)
	DefaultMaxSubdenomLength = uint32(64)
)

var subdenomRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9/:._-]*$`)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Denoms: []DenomAuthorityMetadata{}, Params: DefaultParams()}
}

func DefaultParams() Params {
	return Params{
		MinSubdenomLength:    DefaultMinSubdenomLength,
		MaxSubdenomLength:    DefaultMaxSubdenomLength,
		DenomCreationEnabled: true,
		MintingEnabled:       true,
		BurningEnabled:       true,
	}
}

func NormalizeParams(params Params) Params {
	if isZeroParams(params) {
		return DefaultParams()
	}
	if params.MinSubdenomLength == 0 {
		params.MinSubdenomLength = DefaultMinSubdenomLength
	}
	if params.MaxSubdenomLength == 0 {
		params.MaxSubdenomLength = DefaultMaxSubdenomLength
	}
	return params
}

func isZeroParams(params Params) bool {
	return params.MinSubdenomLength == 0 &&
		params.MaxSubdenomLength == 0 &&
		!params.DenomCreationEnabled &&
		!params.MintingEnabled &&
		!params.BurningEnabled
}

func (p Params) Validate() error {
	if p.MinSubdenomLength == 0 {
		return fmt.Errorf("min_subdenom_length must be positive")
	}
	if p.MaxSubdenomLength < p.MinSubdenomLength {
		return fmt.Errorf("max_subdenom_length must be >= min_subdenom_length")
	}
	if p.MaxSubdenomLength > DefaultMaxSubdenomLength {
		return fmt.Errorf("max_subdenom_length must be <= %d", DefaultMaxSubdenomLength)
	}
	return nil
}

func ValidateSubdenom(subdenom string, params Params) error {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return err
	}
	if len(subdenom) < int(params.MinSubdenomLength) || len(subdenom) > int(params.MaxSubdenomLength) {
		return fmt.Errorf("subdenom length must be between %d and %d", params.MinSubdenomLength, params.MaxSubdenomLength)
	}
	if !subdenomRe.MatchString(subdenom) || strings.Contains(subdenom, "//") {
		return fmt.Errorf("subdenom must start with a letter and contain only letters, numbers, /, :, ., _, or -")
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, denom := range gs.Denoms {
		if denom.Denom == "" {
			return fmt.Errorf("empty denom")
		}
		if err := sdk.ValidateDenom(denom.Denom); err != nil {
			return fmt.Errorf("invalid denom %s: %w", denom.Denom, err)
		}
		if !strings.HasPrefix(denom.Denom, FactoryDenomPrefix+"/") {
			return fmt.Errorf("denom %s must use %s prefix", denom.Denom, FactoryDenomPrefix)
		}
		if denom.Admin == "" {
			return fmt.Errorf("empty admin for denom %s", denom.Denom)
		}
		if _, err := sdk.AccAddressFromBech32(denom.Admin); err != nil {
			return fmt.Errorf("invalid admin for denom %s: %w", denom.Denom, err)
		}
		if _, ok := seen[denom.Denom]; ok {
			return fmt.Errorf("duplicate denom %s", denom.Denom)
		}
		seen[denom.Denom] = struct{}{}
	}
	return nil
}
