package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Denoms: []DenomAuthorityMetadata{}}
}

func (gs GenesisState) Validate() error {
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
