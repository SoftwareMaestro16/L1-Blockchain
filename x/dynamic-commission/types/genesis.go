package types

import (
	"fmt"
	"sort"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

const (
	DefaultBaseCommissionBps	= uint32(1_000)
	DefaultCommissionFloorBps	= uint32(300)
	DefaultCommissionCeilingBps	= uint32(2_000)
	DefaultMaxRateChangeBps		= uint32(100)
	DefaultHighPerformanceThreshold	= uint32(9_000)
	DefaultLowPerformanceThreshold	= uint32(5_000)
	DefaultHighReputationThreshold	= uint32(8_500)
	DefaultLowReputationThreshold	= uint32(4_000)
	DefaultPerformanceBonusBps	= uint32(100)
	DefaultPerformancePenaltyBps	= uint32(100)
	DefaultReputationBonusBps	= uint32(0)
	DefaultReputationPenaltyBps	= uint32(0)
)

func DefaultParams() Params {
	return Params{
		CommissionFloorBps:		DefaultCommissionFloorBps,
		CommissionCeilingBps:		DefaultCommissionCeilingBps,
		MaxRateChangeBps:		DefaultMaxRateChangeBps,
		HighPerformanceThresholdBps:	DefaultHighPerformanceThreshold,
		LowPerformanceThresholdBps:	DefaultLowPerformanceThreshold,
		HighReputationThresholdBps:	DefaultHighReputationThreshold,
		LowReputationThresholdBps:	DefaultLowReputationThreshold,
		PerformanceBonusBps:		DefaultPerformanceBonusBps,
		PerformancePenaltyBps:		DefaultPerformancePenaltyBps,
		ReputationBonusBps:		DefaultReputationBonusBps,
		ReputationPenaltyBps:		DefaultReputationPenaltyBps,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:			DefaultParams(),
		Commissions:		[]ValidatorCommission{},
		CommissionHistory:	[]CommissionHistoryEntry{},
	}
}

func NormalizeParams(params Params) Params {
	if params.CommissionFloorBps == 0 &&
		params.CommissionCeilingBps == 0 &&
		params.MaxRateChangeBps == 0 &&
		params.HighPerformanceThresholdBps == 0 &&
		params.LowPerformanceThresholdBps == 0 &&
		params.HighReputationThresholdBps == 0 &&
		params.LowReputationThresholdBps == 0 &&
		params.PerformanceBonusBps == 0 &&
		params.PerformancePenaltyBps == 0 &&
		params.ReputationBonusBps == 0 &&
		params.ReputationPenaltyBps == 0 {
		return DefaultParams()
	}
	if params.CommissionCeilingBps == 0 {
		params.CommissionCeilingBps = DefaultCommissionCeilingBps
	}
	if params.MaxRateChangeBps == 0 {
		params.MaxRateChangeBps = DefaultMaxRateChangeBps
	}
	if params.HighPerformanceThresholdBps == 0 {
		params.HighPerformanceThresholdBps = DefaultHighPerformanceThreshold
	}
	if params.LowPerformanceThresholdBps == 0 {
		params.LowPerformanceThresholdBps = DefaultLowPerformanceThreshold
	}
	if params.HighReputationThresholdBps == 0 {
		params.HighReputationThresholdBps = DefaultHighReputationThreshold
	}
	if params.LowReputationThresholdBps == 0 {
		params.LowReputationThresholdBps = DefaultLowReputationThreshold
	}
	if params.PerformanceBonusBps == 0 {
		params.PerformanceBonusBps = DefaultPerformanceBonusBps
	}
	if params.PerformancePenaltyBps == 0 {
		params.PerformancePenaltyBps = DefaultPerformancePenaltyBps
	}
	if params.ReputationBonusBps == 0 {
		params.ReputationBonusBps = DefaultReputationBonusBps
	}
	if params.ReputationPenaltyBps == 0 {
		params.ReputationPenaltyBps = DefaultReputationPenaltyBps
	}
	return params
}

func (p Params) Validate() error {
	if p.CommissionFloorBps > p.CommissionCeilingBps {
		return fmt.Errorf("commission floor must be <= ceiling")
	}
	if p.CommissionCeilingBps > BasisPoints {
		return fmt.Errorf("commission ceiling must be <= %d bps", BasisPoints)
	}
	if p.MaxRateChangeBps == 0 || p.MaxRateChangeBps > BasisPoints {
		return fmt.Errorf("max rate change must be between 1 and %d bps", BasisPoints)
	}
	if p.LowPerformanceThresholdBps > p.HighPerformanceThresholdBps || p.HighPerformanceThresholdBps > BasisPoints {
		return fmt.Errorf("performance thresholds must be ordered and <= %d bps", BasisPoints)
	}
	if p.LowReputationThresholdBps > p.HighReputationThresholdBps || p.HighReputationThresholdBps > BasisPoints {
		return fmt.Errorf("reputation thresholds must be ordered and <= %d bps", BasisPoints)
	}
	for name, value := range map[string]uint32{
		"performance_bonus_bps":	p.PerformanceBonusBps,
		"performance_penalty_bps":	p.PerformancePenaltyBps,
		"reputation_bonus_bps":		p.ReputationBonusBps,
		"reputation_penalty_bps":	p.ReputationPenaltyBps,
	} {
		if value > BasisPoints {
			return fmt.Errorf("%s must be <= %d", name, BasisPoints)
		}
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, commission := range gs.Commissions {
		if err := commission.Validate(params); err != nil {
			return err
		}
		if _, ok := seen[commission.ValidatorAddress]; ok {
			return fmt.Errorf("duplicate validator commission %s", commission.ValidatorAddress)
		}
		seen[commission.ValidatorAddress] = struct{}{}
	}
	historyKeys := map[string]struct{}{}
	for _, entry := range gs.CommissionHistory {
		if err := entry.Validate(params); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d/%d/%d", entry.ValidatorAddress, entry.Height, entry.BaseCommissionBps, entry.EffectiveCommissionBps)
		if _, ok := historyKeys[key]; ok {
			return fmt.Errorf("duplicate commission history entry %s", key)
		}
		historyKeys[key] = struct{}{}
	}
	return nil
}

func (c ValidatorCommission) Validate(params Params) error {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := aetraaddress.ValidateUserAddress("validator address", c.ValidatorAddress); err != nil {
		return err
	}
	if c.BaseCommissionBps > BasisPoints || c.EffectiveCommissionBps > BasisPoints {
		return fmt.Errorf("commission bps must be <= %d", BasisPoints)
	}
	if c.CommissionFloorBps != params.CommissionFloorBps || c.CommissionCeilingBps != params.CommissionCeilingBps {
		return fmt.Errorf("commission bounds snapshot must match params")
	}
	if c.EffectiveCommissionBps < c.CommissionFloorBps || c.EffectiveCommissionBps > c.CommissionCeilingBps {
		return fmt.Errorf("effective commission must stay inside floor/ceiling")
	}
	if c.Jailed && c.PerformanceModifierBps > 0 {
		return fmt.Errorf("jailed validators cannot receive performance bonuses")
	}
	return nil
}

func (e CommissionHistoryEntry) Validate(params Params) error {
	params = NormalizeParams(params)
	if err := aetraaddress.ValidateUserAddress("validator address", e.ValidatorAddress); err != nil {
		return err
	}
	if e.Height == 0 {
		return fmt.Errorf("commission history height must be positive")
	}
	commission := ValidatorCommission{
		ValidatorAddress:	e.ValidatorAddress,
		BaseCommissionBps:	e.BaseCommissionBps,
		EffectiveCommissionBps:	e.EffectiveCommissionBps,
		PerformanceModifierBps:	e.PerformanceModifierBps,
		ReputationModifierBps:	e.ReputationModifierBps,
		CommissionFloorBps:	params.CommissionFloorBps,
		CommissionCeilingBps:	params.CommissionCeilingBps,
		LastUpdateHeight:	e.Height,
		Jailed:			e.Jailed,
	}
	return commission.Validate(params)
}

func DefaultCommission(validator string, params Params) ValidatorCommission {
	params = NormalizeParams(params)
	return ValidatorCommission{
		ValidatorAddress:	validator,
		BaseCommissionBps:	clamp(DefaultBaseCommissionBps, params.CommissionFloorBps, params.CommissionCeilingBps),
		EffectiveCommissionBps:	clamp(DefaultBaseCommissionBps, params.CommissionFloorBps, params.CommissionCeilingBps),
		CommissionFloorBps:	params.CommissionFloorBps,
		CommissionCeilingBps:	params.CommissionCeilingBps,
	}
}

func ComputeModifiers(params Params, performanceScoreBps, reputationScoreBps uint32, jailed bool) (int32, int32, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return 0, 0, err
	}
	if performanceScoreBps > BasisPoints || reputationScoreBps > BasisPoints {
		return 0, 0, fmt.Errorf("scores must be <= %d bps", BasisPoints)
	}
	performance := int32(0)
	switch {
	case performanceScoreBps >= params.HighPerformanceThresholdBps && !jailed:
		performance = int32(params.PerformanceBonusBps)
	case performanceScoreBps <= params.LowPerformanceThresholdBps:
		performance = -int32(params.PerformancePenaltyBps)
	}
	reputation := int32(0)
	switch {
	case reputationScoreBps >= params.HighReputationThresholdBps:
		reputation = int32(params.ReputationBonusBps)
	case reputationScoreBps <= params.LowReputationThresholdBps:
		reputation = -int32(params.ReputationPenaltyBps)
	}
	return performance, reputation, nil
}

func EffectiveCommission(params Params, base uint32, performanceModifier, reputationModifier int32) uint32 {
	params = NormalizeParams(params)
	total := int64(base) + int64(performanceModifier) + int64(reputationModifier)
	if total < int64(params.CommissionFloorBps) {
		return params.CommissionFloorBps
	}
	if total > int64(params.CommissionCeilingBps) {
		return params.CommissionCeilingBps
	}
	return uint32(total)
}

func RateLimitExceeded(previous, next, maxDelta uint32) bool {
	if previous > next {
		return previous-next > maxDelta
	}
	return next-previous > maxDelta
}

func SortCommissions(in []ValidatorCommission) []ValidatorCommission {
	out := append([]ValidatorCommission(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].ValidatorAddress < out[j].ValidatorAddress })
	return out
}

func SortHistory(in []CommissionHistoryEntry) []CommissionHistoryEntry {
	out := append([]CommissionHistoryEntry(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ValidatorAddress == out[j].ValidatorAddress {
			return out[i].Height < out[j].Height
		}
		return out[i].ValidatorAddress < out[j].ValidatorAddress
	})
	return out
}

func clamp(value, floor, ceiling uint32) uint32 {
	if value < floor {
		return floor
	}
	if value > ceiling {
		return ceiling
	}
	return value
}
