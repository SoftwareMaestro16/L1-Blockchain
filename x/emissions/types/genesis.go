package types

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/sovereign-l1/l1/app/params"
)

func DefaultDistributionWeights() DistributionWeights {
	return DistributionWeights{
		ValidatorRewardBps:	7_000,
		TreasuryBps:		1_000,
		ProtectionBps:		1_000,
		BurnBps:		500,
		EcosystemBps:		500,
	}
}

func DefaultParams() Params {
	return Params{
		BaseDenom:			BaseDenom,
		CurrentInflationBps:		uint32(appparams.DefaultTargetInflationBps),
		TargetStakingRatioBps:		uint32(appparams.DefaultTargetStakeBps),
		MinAnnualInflationBps:		uint32(appparams.MinInflationBps),
		MaxAnnualInflationBps:		uint32(appparams.MaxInflationBps),
		ConstitutionalMaxInflationBps:	uint32(appparams.MaxInflationBps),
		ResponsivenessBps:		uint32(appparams.DefaultResponsivenessBps),
		AnnualReferenceSupply:		sdk.NewInt64Coin(BaseDenom, 365_000_000_000),
		EpochsPerYear:			365,
		DistributionWeights:		DefaultDistributionWeights(),
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:			DefaultParams(),
		EpochHistory:		[]EmissionEpoch{},
		TotalMintedAccounting:	sdk.NewInt64Coin(BaseDenom, 0),
	}
}

func NormalizeParams(params Params) Params {
	if params.BaseDenom == "" {
		params.BaseDenom = BaseDenom
	}
	if params.CurrentInflationBps == 0 {
		params.CurrentInflationBps = uint32(appparams.DefaultTargetInflationBps)
	}
	if params.TargetStakingRatioBps == 0 {
		params.TargetStakingRatioBps = uint32(appparams.DefaultTargetStakeBps)
	}
	if params.MinAnnualInflationBps == 0 {
		params.MinAnnualInflationBps = uint32(appparams.MinInflationBps)
	}
	if params.MaxAnnualInflationBps == 0 {
		params.MaxAnnualInflationBps = uint32(appparams.MaxInflationBps)
	}
	if params.ConstitutionalMaxInflationBps == 0 {
		params.ConstitutionalMaxInflationBps = params.MaxAnnualInflationBps
	}
	if params.ResponsivenessBps == 0 {
		params.ResponsivenessBps = uint32(appparams.DefaultResponsivenessBps)
	}
	if params.AnnualReferenceSupply.Denom == "" && params.AnnualReferenceSupply.Amount.IsNil() {
		params.AnnualReferenceSupply = sdk.NewInt64Coin(params.BaseDenom, 365_000_000_000)
	}
	if params.EpochsPerYear == 0 {
		params.EpochsPerYear = 365
	}
	if params.DistributionWeights == (DistributionWeights{}) {
		params.DistributionWeights = DefaultDistributionWeights()
	}
	return params
}

func (p Params) Validate() error {
	if p.BaseDenom != BaseDenom {
		return fmt.Errorf("base_denom must be %s", BaseDenom)
	}
	if p.TargetStakingRatioBps > BasisPoints {
		return fmt.Errorf("target_staking_ratio_bps cannot exceed %d", BasisPoints)
	}
	if p.CurrentInflationBps > p.ConstitutionalMaxInflationBps {
		return fmt.Errorf("current inflation cannot exceed constitutional maximum")
	}
	if p.MinAnnualInflationBps > p.MaxAnnualInflationBps {
		return fmt.Errorf("min annual inflation cannot exceed max")
	}
	if p.MaxAnnualInflationBps > p.ConstitutionalMaxInflationBps {
		return fmt.Errorf("max annual inflation cannot exceed constitutional maximum")
	}
	if p.ResponsivenessBps > BasisPoints {
		return fmt.Errorf("responsiveness_bps cannot exceed %d", BasisPoints)
	}
	if p.EpochsPerYear == 0 {
		return fmt.Errorf("epochs_per_year must be positive")
	}
	if err := validateCoin(p.BaseDenom, p.AnnualReferenceSupply, true); err != nil {
		return fmt.Errorf("annual_reference_supply: %w", err)
	}
	return p.DistributionWeights.Validate()
}

func (w DistributionWeights) Validate() error {
	total := uint64(w.ValidatorRewardBps) + uint64(w.TreasuryBps) + uint64(w.ProtectionBps) + uint64(w.BurnBps) + uint64(w.EcosystemBps)
	if total != uint64(BasisPoints) {
		return fmt.Errorf("distribution weights must sum to %d bps", BasisPoints)
	}
	return nil
}

func (e EmissionEpoch) Validate(params Params) error {
	if e.Epoch == 0 {
		return fmt.Errorf("epoch must be positive")
	}
	if e.StakingRatioBps > BasisPoints {
		return fmt.Errorf("staking_ratio_bps cannot exceed %d", BasisPoints)
	}
	if e.InflationBps > params.ConstitutionalMaxInflationBps {
		return fmt.Errorf("inflation cannot exceed constitutional maximum")
	}
	for name, coin := range map[string]sdk.Coin{
		"emission_amount":	e.EmissionAmount,
		"validator_reward":	e.ValidatorReward,
		"treasury":		e.Treasury,
		"protection_fund":	e.ProtectionFund,
		"burn":			e.Burn,
		"ecosystem":		e.Ecosystem,
		"rounding_remainder":	e.RoundingRemainder,
	} {
		if err := validateCoin(params.BaseDenom, coin, false); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	sum := e.ValidatorReward.Amount.Add(e.Treasury.Amount).Add(e.ProtectionFund.Amount).Add(e.Burn.Amount).Add(e.Ecosystem.Amount).Add(e.RoundingRemainder.Amount)
	if !sum.Equal(e.EmissionAmount.Amount) {
		return fmt.Errorf("emission amount %s does not match distribution accounting %s", e.EmissionAmount, sum.String())
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validateCoin(params.BaseDenom, gs.TotalMintedAccounting, false); err != nil {
		return fmt.Errorf("total_minted_accounting: %w", err)
	}
	seen := map[uint64]struct{}{}
	total := sdkmath.ZeroInt()
	for _, epoch := range gs.EpochHistory {
		if _, ok := seen[epoch.Epoch]; ok {
			return fmt.Errorf("duplicate emission epoch %d", epoch.Epoch)
		}
		seen[epoch.Epoch] = struct{}{}
		if err := epoch.Validate(params); err != nil {
			return err
		}
		total = total.Add(epoch.EmissionAmount.Amount)
	}
	if !total.Equal(gs.TotalMintedAccounting.Amount) {
		return fmt.Errorf("minted accounting %s does not match epoch total %s", gs.TotalMintedAccounting.Amount, total)
	}
	return nil
}

func SortEmissionEpochs(in []EmissionEpoch) []EmissionEpoch {
	out := append([]EmissionEpoch(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out
}

func ComputeInflationBps(params Params, stakingRatioBps uint32) uint32 {
	current := int64(params.CurrentInflationBps)
	delta := int64(params.TargetStakingRatioBps) - int64(stakingRatioBps)
	adjustment := delta * int64(params.ResponsivenessBps) / int64(BasisPoints)
	next := current + adjustment
	if next < int64(params.MinAnnualInflationBps) {
		next = int64(params.MinAnnualInflationBps)
	}
	if next > int64(params.MaxAnnualInflationBps) {
		next = int64(params.MaxAnnualInflationBps)
	}
	if next > int64(params.ConstitutionalMaxInflationBps) {
		next = int64(params.ConstitutionalMaxInflationBps)
	}
	return uint32(next)
}

func ComputeEpochEmission(params Params, epoch, stakingRatioBps uint64, height int64) (EmissionEpoch, error) {
	if stakingRatioBps > uint64(BasisPoints) {
		return EmissionEpoch{}, ErrInvalidEpoch.Wrap("staking_ratio_bps cannot exceed basis points")
	}
	inflationBps := ComputeInflationBps(params, uint32(stakingRatioBps))
	annual := params.AnnualReferenceSupply.Amount.MulRaw(int64(inflationBps)).QuoRaw(int64(BasisPoints))
	amount := annual.QuoRaw(int64(params.EpochsPerYear))
	emission := sdk.NewCoin(params.BaseDenom, amount)
	epochRecord := EmissionEpoch{
		Epoch:			epoch,
		StakingRatioBps:	uint32(stakingRatioBps),
		InflationBps:		inflationBps,
		EmissionAmount:		emission,
		ValidatorReward:	sdk.NewCoin(params.BaseDenom, bpsAmount(amount, params.DistributionWeights.ValidatorRewardBps)),
		Treasury:		sdk.NewCoin(params.BaseDenom, bpsAmount(amount, params.DistributionWeights.TreasuryBps)),
		ProtectionFund:		sdk.NewCoin(params.BaseDenom, bpsAmount(amount, params.DistributionWeights.ProtectionBps)),
		Burn:			sdk.NewCoin(params.BaseDenom, bpsAmount(amount, params.DistributionWeights.BurnBps)),
		Ecosystem:		sdk.NewCoin(params.BaseDenom, bpsAmount(amount, params.DistributionWeights.EcosystemBps)),
		FinalizedHeight:	height,
	}
	distributed := epochRecord.ValidatorReward.Amount.Add(epochRecord.Treasury.Amount).Add(epochRecord.ProtectionFund.Amount).Add(epochRecord.Burn.Amount).Add(epochRecord.Ecosystem.Amount)
	epochRecord.RoundingRemainder = sdk.NewCoin(params.BaseDenom, amount.Sub(distributed))
	if err := epochRecord.Validate(params); err != nil {
		return EmissionEpoch{}, ErrInvalidEpoch.Wrap(err.Error())
	}
	return epochRecord, nil
}

func validateCoin(denom string, coin sdk.Coin, allowPositive bool) error {
	if coin.Denom != denom {
		return fmt.Errorf("denom must be %s", denom)
	}
	if coin.Amount.IsNil() || coin.Amount.IsNegative() {
		return fmt.Errorf("amount cannot be negative")
	}
	if allowPositive && coin.Amount.IsZero() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

func bpsAmount(amount sdkmath.Int, bps uint32) sdkmath.Int {
	if amount.IsZero() || bps == 0 {
		return sdkmath.ZeroInt()
	}
	return amount.MulRaw(int64(bps)).QuoRaw(int64(BasisPoints))
}
