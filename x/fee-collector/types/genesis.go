package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	FeeSplitBurnMinBps		uint32	= 3_000
	FeeSplitBurnMaxBps		uint32	= 6_000
	FeeSplitValidatorsMinBps	uint32	= 2_000
	FeeSplitValidatorsMaxBps	uint32	= 4_000
	FeeSplitTreasuryMinBps		uint32	= 1_000
	FeeSplitTreasuryMaxBps		uint32	= 2_000
	FeeSplitProtectionBps		uint32	= 0
)

func DefaultParams() Params {
	return Params{
		BaseDenom:		BaseDenom,
		TreasuryBps:		1_500,
		ProtectionBps:		FeeSplitProtectionBps,
		ValidatorsBps:		3_500,
		BurnBps:		5_000,
		CollectorModule:	CollectorModuleName,
		TreasuryModule:		TreasuryModuleName,
		ProtectionModule:	ProtectionModuleName,
		ValidatorsModule:	authtypes.FeeCollectorName,
	}
}

func DefaultFeeBalances() FeeBalances {
	return FeeBalances{
		GasFees:		sdk.NewCoins(),
		ForwardingFees:		sdk.NewCoins(),
		ProtocolFees:		sdk.NewCoins(),
		TotalCollected:		sdk.NewCoins(),
		TotalDistributed:	sdk.NewCoins(),
		TotalBurned:		sdk.NewCoins(),
	}
}

func DefaultPendingDistribution() PendingDistribution {
	return PendingDistribution{
		Treasury:	sdk.NewCoins(),
		Protection:	sdk.NewCoins(),
		Validators:	sdk.NewCoins(),
		Burn:		sdk.NewCoins(),
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:			DefaultParams(),
		Balances:		DefaultFeeBalances(),
		PendingDistribution:	DefaultPendingDistribution(),
		FeeHistory:		[]FeeHistoryEntry{},
	}
}

func NormalizeParams(params Params) Params {
	if params.BaseDenom == "" {
		params.BaseDenom = BaseDenom
	}
	if params.CollectorModule == "" {
		params.CollectorModule = CollectorModuleName
	}
	if params.TreasuryModule == "" {
		params.TreasuryModule = TreasuryModuleName
	}
	if params.ProtectionModule == "" {
		params.ProtectionModule = ProtectionModuleName
	}
	if params.ValidatorsModule == "" {
		params.ValidatorsModule = authtypes.FeeCollectorName
	}
	return params
}

func (p Params) Validate() error {
	if p.BaseDenom != BaseDenom {
		return fmt.Errorf("base_denom must be %s", BaseDenom)
	}
	if p.CollectorModule != CollectorModuleName {
		return fmt.Errorf("collector_module must be %s", CollectorModuleName)
	}
	if p.TreasuryModule != TreasuryModuleName {
		return fmt.Errorf("treasury_module must be %s", TreasuryModuleName)
	}
	if p.ProtectionModule != ProtectionModuleName {
		return fmt.Errorf("protection_module must be %s", ProtectionModuleName)
	}
	if p.ValidatorsModule != authtypes.FeeCollectorName {
		return fmt.Errorf("validators_module must be %s", authtypes.FeeCollectorName)
	}
	if err := validateUint32Bps("burn_bps", p.BurnBps, FeeSplitBurnMinBps, FeeSplitBurnMaxBps); err != nil {
		return err
	}
	if err := validateUint32Bps("validators_bps", p.ValidatorsBps, FeeSplitValidatorsMinBps, FeeSplitValidatorsMaxBps); err != nil {
		return err
	}
	if err := validateUint32Bps("treasury_bps", p.TreasuryBps, FeeSplitTreasuryMinBps, FeeSplitTreasuryMaxBps); err != nil {
		return err
	}
	if p.ProtectionBps != FeeSplitProtectionBps {
		return fmt.Errorf("protection_bps must be %d for Aetra fee split v1", FeeSplitProtectionBps)
	}
	total := uint64(p.TreasuryBps) + uint64(p.ProtectionBps) + uint64(p.ValidatorsBps) + uint64(p.BurnBps)
	if total != uint64(BasisPoints) {
		return fmt.Errorf("distribution proportions must sum to %d bps", BasisPoints)
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := gs.Balances.Validate(params.BaseDenom); err != nil {
		return err
	}
	if err := gs.PendingDistribution.Validate(params.BaseDenom); err != nil {
		return err
	}
	if !gs.Balances.AccountingBalance().Equal(gs.PendingDistribution.Total()) {
		return fmt.Errorf("accounting balance %s must equal pending distribution %s", gs.Balances.AccountingBalance(), gs.PendingDistribution.Total())
	}
	seen := map[uint64]struct{}{}
	for _, entry := range gs.FeeHistory {
		if _, ok := seen[entry.Epoch]; ok {
			return fmt.Errorf("duplicate fee history epoch %d", entry.Epoch)
		}
		seen[entry.Epoch] = struct{}{}
		if err := entry.Validate(params.BaseDenom); err != nil {
			return err
		}
	}
	return nil
}

func (b FeeBalances) Validate(baseDenom string) error {
	for name, coins := range map[string]sdk.Coins{
		"gas_fees":		b.GasFees,
		"forwarding_fees":	b.ForwardingFees,
		"protocol_fees":	b.ProtocolFees,
		"total_collected":	b.TotalCollected,
		"total_distributed":	b.TotalDistributed,
		"total_burned":		b.TotalBurned,
	} {
		if err := validateBaseCoins(name, baseDenom, coins); err != nil {
			return err
		}
	}
	if !b.TotalCollected.Equal(b.AccountingBalance().Add(b.TotalDistributed...).Add(b.TotalBurned...)) {
		return fmt.Errorf("collected fees cannot disappear")
	}
	return nil
}

func (b FeeBalances) AccountingBalance() sdk.Coins {
	return b.GasFees.Add(b.ForwardingFees...).Add(b.ProtocolFees...)
}

func (p PendingDistribution) Validate(baseDenom string) error {
	for name, coins := range map[string]sdk.Coins{
		"treasury":	p.Treasury,
		"protection":	p.Protection,
		"validators":	p.Validators,
		"burn":		p.Burn,
	} {
		if err := validateBaseCoins(name, baseDenom, coins); err != nil {
			return err
		}
	}
	return nil
}

func (p PendingDistribution) Total() sdk.Coins {
	return p.Treasury.Add(p.Protection...).Add(p.Validators...).Add(p.Burn...)
}

func (e FeeHistoryEntry) Validate(baseDenom string) error {
	for name, coins := range map[string]sdk.Coins{
		"collected":		e.Collected,
		"treasury":		e.Treasury,
		"protection":		e.Protection,
		"validators":		e.Validators,
		"burn":			e.Burn,
		"rounding_remainder":	e.RoundingRemainder,
	} {
		if err := validateBaseCoins(name, baseDenom, coins); err != nil {
			return err
		}
	}
	out := e.Treasury.Add(e.Protection...).Add(e.Validators...).Add(e.Burn...)
	if !e.Collected.Equal(out) {
		return fmt.Errorf("fee history epoch %d creates or loses coins: %s != %s", e.Epoch, e.Collected, out)
	}
	return nil
}

func ValidateFeeCoins(baseDenom string, fees sdk.Coins) error {
	if fees.Empty() {
		return ErrInvalidFee.Wrap("fees must be positive")
	}
	if !fees.IsValid() {
		return ErrInvalidFee.Wrapf("invalid fee coins: %s", fees)
	}
	for _, fee := range fees {
		if fee.IsNil() || !fee.IsPositive() {
			return ErrInvalidFee.Wrapf("fee coin must be positive: %s", fee)
		}
		if fee.Denom != baseDenom {
			return ErrInvalidFee.Wrapf("fee denom %s not accepted; use %s", fee.Denom, baseDenom)
		}
	}
	return nil
}

func SplitFees(params Params, fees sdk.Coins) (PendingDistribution, sdk.Coins, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return PendingDistribution{}, nil, ErrInvalidParams.Wrap(err.Error())
	}
	if err := ValidateFeeCoins(params.BaseDenom, fees); err != nil {
		return PendingDistribution{}, nil, err
	}

	out := DefaultPendingDistribution()
	remainder := sdk.NewCoins()
	for _, fee := range fees {
		treasuryAmount := bpsAmount(fee.Amount, params.TreasuryBps)
		protectionAmount := bpsAmount(fee.Amount, params.ProtectionBps)
		validatorsAmount := bpsAmount(fee.Amount, params.ValidatorsBps)
		burnAmount := bpsAmount(fee.Amount, params.BurnBps)
		allocated := treasuryAmount.Add(protectionAmount).Add(validatorsAmount).Add(burnAmount)
		rounding := fee.Amount.Sub(allocated)
		if rounding.IsPositive() {
			treasuryAmount = treasuryAmount.Add(rounding)
			remainder = remainder.Add(sdk.NewCoin(fee.Denom, rounding))
		}
		out.Treasury = addPositive(out.Treasury, fee.Denom, treasuryAmount)
		out.Protection = addPositive(out.Protection, fee.Denom, protectionAmount)
		out.Validators = addPositive(out.Validators, fee.Denom, validatorsAmount)
		out.Burn = addPositive(out.Burn, fee.Denom, burnAmount)
	}
	return out, remainder, nil
}

func validateBaseCoins(name, baseDenom string, coins sdk.Coins) error {
	if !coins.IsValid() {
		return fmt.Errorf("%s coins must be valid", name)
	}
	for _, coin := range coins {
		if coin.IsNil() || coin.IsNegative() {
			return fmt.Errorf("%s coin must be non-negative: %s", name, coin)
		}
		if coin.Denom != baseDenom {
			return fmt.Errorf("%s only supports denom %s", name, baseDenom)
		}
	}
	return nil
}

func validateUint32Bps(name string, value, min, max uint32) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d bps", name, min, max)
	}
	return nil
}

func bpsAmount(amount sdkmath.Int, bps uint32) sdkmath.Int {
	return amount.MulRaw(int64(bps)).QuoRaw(int64(BasisPoints))
}

func addPositive(coins sdk.Coins, denom string, amount sdkmath.Int) sdk.Coins {
	if amount.IsPositive() {
		return coins.Add(sdk.NewCoin(denom, amount))
	}
	return coins
}
