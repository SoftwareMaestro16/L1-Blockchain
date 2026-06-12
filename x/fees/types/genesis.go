package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	MinDefaultFeeAmount	= "1"
	FeeCollectorModuleName	= "fee_collector"
	DistributionModuleName	= "distribution"
	ProtocolPoolModuleName	= "protocolpool"
	ValidatorRewardsTarget	= "distribution/validator_rewards"
	CommunityPoolTarget	= "protocolpool/community_pool"
)

func DefaultParams() Params {
	return Params{
		AllowedFeeDenoms:		[]string{BondDenom},
		ValidatorRewardsRatio:		"0.98",
		CommunityPoolRatio:		"0.02",
		MinFeeAmount:			MinDefaultFeeAmount,
		FeeCollectorModule:		FeeCollectorModuleName,
		ValidatorRewardsTarget:		ValidatorRewardsTarget,
		CommunityPoolTarget:		CommunityPoolTarget,
		BaseFeeAmount:			DefaultBaseFeeAmount,
		MaxFeeAmount:			DefaultMaxFeeAmount,
		TargetBlockUtilizationBps:	DefaultTargetUtilizationBps,
		CongestionThresholdBps:		DefaultCongestionBps,
		MaxTxGas:			DefaultMaxTxGas,
		MaxBlockGas:			DefaultMaxBlockGas,
		MaxBlockTxs:			DefaultMaxBlockTxs,
		MaxSenderTxsPerBlock:		DefaultSenderTxsPerBlock,
		StakeTxAllowanceStepAmount:	DefaultStakeAllowanceStep,
		MaxSenderTxsPerBlockWithStake:	DefaultStakeSenderTxsPerBlock,
		FeePriorityWeightBps:		DefaultFeePriorityWeightBps,
		StakePriorityWeightBps:		DefaultStakePriorityWeightBps,
	}
}

func DefaultProtocolFeeState() ProtocolFeeState {
	return ProtocolFeeState{
		TotalCollected:		sdk.NewCoins(),
		ValidatorRewards:	sdk.NewCoins(),
		CommunityPool:		sdk.NewCoins(),
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams(), ProtocolFeeState: DefaultProtocolFeeState()}
}

func (p Params) Validate() error {
	if err := appparams.ValidateNativeFeeDenomsV1(p.AllowedFeeDenoms, MaxAllowedFeeDenomsV1); err != nil {
		return err
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
	if _, err := validateMinFeeAmount(p.MinFeeAmount); err != nil {
		return err
	}
	if _, err := validateBaseFeeAmount(p.BaseFeeAmount, p.MinFeeAmount); err != nil {
		return err
	}
	if _, err := validateMaxFeeAmount(p.MaxFeeAmount, p.BaseFeeAmount); err != nil {
		return err
	}
	if err := validateUtilizationParams(p); err != nil {
		return err
	}
	if _, err := validatePositiveAmount("stake_tx_allowance_step_amount", p.StakeTxAllowanceStepAmount, MaxFeeAmountV1); err != nil {
		return err
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
	if params.BaseFeeAmount == "" {
		params.BaseFeeAmount = DefaultBaseFeeAmount
	}
	if params.MaxFeeAmount == "" {
		params.MaxFeeAmount = DefaultMaxFeeAmount
	}
	if params.TargetBlockUtilizationBps == 0 {
		params.TargetBlockUtilizationBps = DefaultTargetUtilizationBps
	}
	if params.CongestionThresholdBps == 0 {
		params.CongestionThresholdBps = DefaultCongestionBps
	}
	if params.MaxTxGas == 0 {
		params.MaxTxGas = DefaultMaxTxGas
	}
	if params.MaxBlockGas == 0 {
		params.MaxBlockGas = DefaultMaxBlockGas
	}
	if params.MaxBlockTxs == 0 {
		params.MaxBlockTxs = DefaultMaxBlockTxs
	}
	if params.MaxSenderTxsPerBlock == 0 {
		params.MaxSenderTxsPerBlock = DefaultSenderTxsPerBlock
	}
	if params.StakeTxAllowanceStepAmount == "" {
		params.StakeTxAllowanceStepAmount = DefaultStakeAllowanceStep
	}
	if params.MaxSenderTxsPerBlockWithStake == 0 {
		params.MaxSenderTxsPerBlockWithStake = DefaultStakeSenderTxsPerBlock
	}
	if params.FeePriorityWeightBps == 0 {
		params.FeePriorityWeightBps = DefaultFeePriorityWeightBps
	}
	if params.StakePriorityWeightBps == 0 {
		params.StakePriorityWeightBps = DefaultStakePriorityWeightBps
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

func validateMinFeeAmount(value string) (sdkmath.Int, error) {
	return validatePositiveAmount("min_fee_amount", value, MaxMinFeeAmountV1)
}

func validateBaseFeeAmount(baseValue, minValue string) (sdkmath.Int, error) {
	baseFee, err := validatePositiveAmount("base_fee_amount", baseValue, MaxFeeAmountV1)
	if err != nil {
		return sdkmath.Int{}, err
	}
	minFee, err := validateMinFeeAmount(minValue)
	if err != nil {
		return sdkmath.Int{}, err
	}
	if baseFee.LT(minFee) {
		return sdkmath.Int{}, fmt.Errorf("base_fee_amount must be >= min_fee_amount")
	}
	return baseFee, nil
}

func validateMaxFeeAmount(maxValue, baseValue string) (sdkmath.Int, error) {
	maxFee, err := validatePositiveAmount("max_fee_amount", maxValue, MaxFeeAmountV1)
	if err != nil {
		return sdkmath.Int{}, err
	}
	baseFee, err := validatePositiveAmount("base_fee_amount", baseValue, MaxFeeAmountV1)
	if err != nil {
		return sdkmath.Int{}, err
	}
	if maxFee.LT(baseFee) {
		return sdkmath.Int{}, fmt.Errorf("max_fee_amount must be >= base_fee_amount")
	}
	return maxFee, nil
}

func validatePositiveAmount(name, value, maxValue string) (sdkmath.Int, error) {
	amount, ok := sdkmath.NewIntFromString(value)
	if !ok || !amount.IsPositive() {
		return sdkmath.Int{}, fmt.Errorf("%s must be a positive integer", name)
	}
	maxAmount, ok := sdkmath.NewIntFromString(maxValue)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid max bound for %s", name)
	}
	if amount.GT(maxAmount) {
		return sdkmath.Int{}, fmt.Errorf("%s must be <= %s", name, maxValue)
	}
	return amount, nil
}

func validateUtilizationParams(params Params) error {
	if params.TargetBlockUtilizationBps == 0 || params.TargetBlockUtilizationBps >= uint32(BasisPoints) {
		return fmt.Errorf("target_block_utilization_bps must be between 1 and 9999")
	}
	if params.CongestionThresholdBps <= params.TargetBlockUtilizationBps || params.CongestionThresholdBps > uint32(BasisPoints) {
		return fmt.Errorf("congestion_threshold_bps must be greater than target and <= 10000")
	}
	if params.MaxTxGas == 0 || params.MaxBlockGas == 0 || params.MaxTxGas > params.MaxBlockGas {
		return fmt.Errorf("gas limits must be positive and max_tx_gas must be <= max_block_gas")
	}
	if params.MaxBlockTxs == 0 {
		return fmt.Errorf("max_block_txs must be positive")
	}
	if params.MaxSenderTxsPerBlock == 0 {
		return fmt.Errorf("max_sender_txs_per_block must be positive")
	}
	if params.MaxSenderTxsPerBlockWithStake < params.MaxSenderTxsPerBlock {
		return fmt.Errorf("max_sender_txs_per_block_with_stake must be >= max_sender_txs_per_block")
	}
	if uint64(params.FeePriorityWeightBps)+uint64(params.StakePriorityWeightBps) != BasisPoints {
		return fmt.Errorf("priority weights must sum to 10000 bps")
	}
	return nil
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
	return validateMinFeeAmount(p.MinFeeAmount)
}

func (p Params) BaseFeeInt() (sdkmath.Int, error) {
	return validateBaseFeeAmount(p.BaseFeeAmount, p.MinFeeAmount)
}

func (p Params) MaxFeeInt() (sdkmath.Int, error) {
	return validateMaxFeeAmount(p.MaxFeeAmount, p.BaseFeeAmount)
}

func (p Params) StakeTxAllowanceStepInt() (sdkmath.Int, error) {
	return validatePositiveAmount("stake_tx_allowance_step_amount", p.StakeTxAllowanceStepAmount, MaxFeeAmountV1)
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
		maxFee, err := params.MaxFeeInt()
		if err != nil {
			return err
		}
		if fees.AmountOf(BondDenom).GT(maxFee) {
			return ErrInvalidFee.Wrapf("fee must not exceed hard cap %s%s", maxFee.String(), BondDenom)
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
