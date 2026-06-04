package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

func DefaultParams() Params {
	return Params{
		AllowedFeeDenoms:      []string{BondDenom},
		ValidatorRewardsRatio: "0.98",
		CommunityPoolRatio:    "0.02",
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams()}
}

func (p Params) Validate() error {
	if len(p.AllowedFeeDenoms) != 1 || p.AllowedFeeDenoms[0] != BondDenom {
		return fmt.Errorf("v1 only allows fee denom %s", BondDenom)
	}
	validatorRatio, err := validateRatio("validator_rewards_ratio", p.ValidatorRewardsRatio)
	if err != nil {
		return err
	}
	communityRatio, err := validateRatio("community_pool_ratio", p.CommunityPoolRatio)
	if err != nil {
		return err
	}
	if !validatorRatio.Add(communityRatio).Equal(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("fee split ratios must sum to 1")
	}
	return nil
}

func validateRatio(name, value string) (sdkmath.LegacyDec, error) {
	if value == "" {
		return sdkmath.LegacyDec{}, fmt.Errorf("%s must be set", name)
	}
	ratio, err := sdkmath.LegacyNewDecFromStr(value)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("invalid %s: %w", name, err)
	}
	if ratio.IsNegative() || ratio.GT(sdkmath.LegacyOneDec()) {
		return sdkmath.LegacyDec{}, fmt.Errorf("%s must be between 0 and 1", name)
	}
	return ratio, nil
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
