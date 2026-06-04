package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MinDefaultFeeAmount    = "1"
	FeeCollectorModuleName = "fee_collector"
	DistributionModuleName = "distribution"
	ProtocolPoolModuleName = "protocolpool"
	ValidatorRewardsTarget = "distribution/validator_rewards"
	CommunityPoolTarget    = "protocolpool/community_pool"
	MaxAllowedFeeDenoms    = 8
)

func DefaultParams() Params {
	return Params{
		AllowedFeeDenoms:       []string{BondDenom},
		ValidatorRewardsRatio:  "0.98",
		CommunityPoolRatio:     "0.02",
		MinFeeAmount:           MinDefaultFeeAmount,
		FeeCollectorModule:     FeeCollectorModuleName,
		ValidatorRewardsTarget: ValidatorRewardsTarget,
		CommunityPoolTarget:    CommunityPoolTarget,
	}
}

func DefaultProtocolFeeState() ProtocolFeeState {
	return ProtocolFeeState{
		TotalCollected:   sdk.NewCoins(),
		ValidatorRewards: sdk.NewCoins(),
		CommunityPool:    sdk.NewCoins(),
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams(), ProtocolFeeState: DefaultProtocolFeeState()}
}

func (p Params) Validate() error {
	if len(p.AllowedFeeDenoms) == 0 || len(p.AllowedFeeDenoms) > MaxAllowedFeeDenoms {
		return fmt.Errorf("allowed fee denom count must be between 1 and %d", MaxAllowedFeeDenoms)
	}
	seenDenoms := map[string]struct{}{}
	for _, denom := range p.AllowedFeeDenoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return fmt.Errorf("invalid fee denom %q: %w", denom, err)
		}
		if _, ok := seenDenoms[denom]; ok {
			return fmt.Errorf("duplicate fee denom %s", denom)
		}
		seenDenoms[denom] = struct{}{}
	}
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
	minFee, ok := sdkmath.NewIntFromString(p.MinFeeAmount)
	if !ok || !minFee.IsPositive() {
		return fmt.Errorf("min_fee_amount must be a positive integer")
	}
	if p.FeeCollectorModule != FeeCollectorModuleName {
		return fmt.Errorf("fee_collector_module must be %s", FeeCollectorModuleName)
	}
	if p.ValidatorRewardsTarget != ValidatorRewardsTarget {
		return fmt.Errorf("validator_rewards_target must be %s", ValidatorRewardsTarget)
	}
	if p.CommunityPoolTarget != CommunityPoolTarget {
		return fmt.Errorf("community_pool_target must be %s", CommunityPoolTarget)
	}
	return nil
}

func NormalizeParams(params Params) Params {
	if params.MinFeeAmount == "" {
		params.MinFeeAmount = MinDefaultFeeAmount
	}
	if params.FeeCollectorModule == "" {
		params.FeeCollectorModule = FeeCollectorModuleName
	}
	if params.ValidatorRewardsTarget == "" {
		params.ValidatorRewardsTarget = ValidatorRewardsTarget
	}
	if params.CommunityPoolTarget == "" {
		params.CommunityPoolTarget = CommunityPoolTarget
	}
	return params
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
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	return gs.ProtocolFeeState.Validate()
}

func (p Params) CommunityRatioDec() (sdkmath.LegacyDec, error) {
	return validateRatio("community_pool_ratio", p.CommunityPoolRatio)
}

func (p Params) MinFeeInt() (sdkmath.Int, error) {
	minFee, ok := sdkmath.NewIntFromString(p.MinFeeAmount)
	if !ok || !minFee.IsPositive() {
		return sdkmath.Int{}, fmt.Errorf("min_fee_amount must be a positive integer")
	}
	return minFee, nil
}

func ValidateFeeCoins(params Params, fees sdk.Coins, enforceMin bool) error {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return err
	}
	if !fees.IsValid() {
		return ErrInvalidFee.Wrapf("invalid fee coins: %s", fees)
	}
	if fees.Empty() {
		if enforceMin {
			return ErrInvalidFee.Wrap("fee must be positive")
		}
		return nil
	}
	for _, fee := range fees {
		if fee.IsNil() || !fee.IsPositive() {
			return ErrInvalidFee.Wrapf("fee coin must be positive: %s", fee)
		}
		allowed := false
		for _, denom := range params.AllowedFeeDenoms {
			if fee.Denom == denom {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrInvalidFee.Wrapf("fee denom %s not accepted; use %s", fee.Denom, BondDenom)
		}
	}
	if enforceMin {
		minFee, err := params.MinFeeInt()
		if err != nil {
			return err
		}
		if fees.AmountOf(BondDenom).LT(minFee) {
			return ErrInvalidFee.Wrapf("fee must be at least %s%s", minFee.String(), BondDenom)
		}
	}
	return nil
}

func SplitFees(params Params, fees sdk.Coins) (sdk.Coins, sdk.Coins, error) {
	params = NormalizeParams(params)
	if err := ValidateFeeCoins(params, fees, false); err != nil {
		return nil, nil, err
	}
	communityRatio, err := params.CommunityRatioDec()
	if err != nil {
		return nil, nil, err
	}
	validatorRewards := sdk.NewCoins()
	communityPool := sdk.NewCoins()
	for _, fee := range fees {
		communityAmount := sdkmath.LegacyNewDecFromInt(fee.Amount).Mul(communityRatio).TruncateInt()
		validatorAmount := fee.Amount.Sub(communityAmount)
		if communityAmount.IsPositive() {
			communityPool = communityPool.Add(sdk.NewCoin(fee.Denom, communityAmount))
		}
		if validatorAmount.IsPositive() {
			validatorRewards = validatorRewards.Add(sdk.NewCoin(fee.Denom, validatorAmount))
		}
	}
	return validatorRewards, communityPool, nil
}

func (s ProtocolFeeState) Validate() error {
	if !s.TotalCollected.IsValid() || !s.ValidatorRewards.IsValid() || !s.CommunityPool.IsValid() {
		return fmt.Errorf("protocol fee accounting coins must be valid")
	}
	for _, coins := range []sdk.Coins{s.TotalCollected, s.ValidatorRewards, s.CommunityPool} {
		for _, coin := range coins {
			if coin.Denom != BondDenom {
				return fmt.Errorf("protocol fee accounting only supports denom %s", BondDenom)
			}
		}
	}
	targetTotal := s.ValidatorRewards.Add(s.CommunityPool...)
	if !s.TotalCollected.Equal(targetTotal) {
		return fmt.Errorf("protocol fee accounting mismatch: total %s != targets %s", s.TotalCollected, targetTotal)
	}
	return nil
}
