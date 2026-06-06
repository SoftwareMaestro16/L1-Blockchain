package types

import (
	"fmt"
	"math/bits"
	"sort"

	"github.com/sovereign-l1/l1/app/addressing"
)

func DefaultParams() Params {
	return Params{
		MaxVotingPowerBps:          3_334,
		SoftVotingPowerBps:         2_500,
		MaxRewardReductionBps:      3_000,
		WarningThresholdBps:        2_500,
		DelegationRejectionEnabled: true,
	}
}

func DefaultNetworkConcentration() NetworkConcentration {
	return NetworkConcentration{Validators: []ValidatorConcentration{}, Warnings: []string{}}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{Params: DefaultParams(), Network: DefaultNetworkConcentration()}
}

func NormalizeParams(params Params) Params {
	if params.MaxVotingPowerBps == 0 {
		params.MaxVotingPowerBps = DefaultParams().MaxVotingPowerBps
	}
	if params.SoftVotingPowerBps == 0 {
		params.SoftVotingPowerBps = DefaultParams().SoftVotingPowerBps
	}
	if params.WarningThresholdBps == 0 {
		params.WarningThresholdBps = params.SoftVotingPowerBps
	}
	if params.MaxRewardReductionBps == 0 {
		params.MaxRewardReductionBps = DefaultParams().MaxRewardReductionBps
	}
	return params
}

func (p Params) Validate() error {
	if p.MaxVotingPowerBps == 0 || p.MaxVotingPowerBps > BasisPoints {
		return fmt.Errorf("max_voting_power_bps must be between 1 and %d", BasisPoints)
	}
	if p.SoftVotingPowerBps == 0 || p.SoftVotingPowerBps > p.MaxVotingPowerBps {
		return fmt.Errorf("soft_voting_power_bps must be between 1 and max_voting_power_bps")
	}
	if p.WarningThresholdBps == 0 || p.WarningThresholdBps > p.MaxVotingPowerBps {
		return fmt.Errorf("warning_threshold_bps must be between 1 and max_voting_power_bps")
	}
	if p.MaxRewardReductionBps > BasisPoints {
		return fmt.Errorf("max_reward_reduction_bps cannot exceed %d", BasisPoints)
	}
	return nil
}

func (v ValidatorPower) Validate() error {
	if err := addressing.ValidateUserAddress("operator_address", v.OperatorAddress); err != nil {
		return err
	}
	return nil
}

func (v ValidatorConcentration) Validate(params Params) error {
	if err := addressing.ValidateUserAddress("operator_address", v.OperatorAddress); err != nil {
		return err
	}
	if v.RawVotingPowerBps > BasisPoints {
		return fmt.Errorf("raw_voting_power_bps cannot exceed %d", BasisPoints)
	}
	if v.EffectiveVotingPowerBps > params.MaxVotingPowerBps {
		return fmt.Errorf("effective voting power exceeds hard cap")
	}
	if v.RewardModifierBps > BasisPoints {
		return fmt.Errorf("reward_modifier_bps cannot exceed %d", BasisPoints)
	}
	if v.AboveHardCap && v.DelegationAllowed && params.DelegationRejectionEnabled {
		return fmt.Errorf("validator above hard cap cannot accept delegation")
	}
	return nil
}

func (n NetworkConcentration) Validate(params Params) error {
	previous := ""
	seen := map[string]struct{}{}
	var top uint32
	for i, validator := range n.Validators {
		if err := validator.Validate(params); err != nil {
			return err
		}
		if _, ok := seen[validator.OperatorAddress]; ok {
			return fmt.Errorf("duplicate validator %s", validator.OperatorAddress)
		}
		seen[validator.OperatorAddress] = struct{}{}
		if i > 0 && previous >= validator.OperatorAddress {
			return fmt.Errorf("validators must be sorted canonically")
		}
		previous = validator.OperatorAddress
		if validator.EffectiveVotingPowerBps > top {
			top = validator.EffectiveVotingPowerBps
		}
	}
	if n.MaxValidatorPowerBps != top {
		return fmt.Errorf("max_validator_power_bps does not match validators")
	}
	if n.MaxValidatorPowerBps > params.MaxVotingPowerBps {
		return fmt.Errorf("network max validator power exceeds hard cap")
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	return gs.Network.Validate(params)
}

func ComputeNetworkConcentration(params Params, epoch uint64, validatorSet []ValidatorPower, height int64) (NetworkConcentration, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return NetworkConcentration{}, ErrInvalidParams.Wrap(err.Error())
	}
	canonical := append([]ValidatorPower(nil), validatorSet...)
	sort.Slice(canonical, func(i, j int) bool { return canonical[i].OperatorAddress < canonical[j].OperatorAddress })
	seen := map[string]struct{}{}
	total := uint64(0)
	for _, validator := range canonical {
		if err := validator.Validate(); err != nil {
			return NetworkConcentration{}, ErrInvalidConcentration.Wrap(err.Error())
		}
		if _, ok := seen[validator.OperatorAddress]; ok {
			return NetworkConcentration{}, ErrInvalidConcentration.Wrapf("duplicate validator %s", validator.OperatorAddress)
		}
		seen[validator.OperatorAddress] = struct{}{}
		if validator.VotingPower > ^uint64(0)-total {
			return NetworkConcentration{}, ErrInvalidConcentration.Wrap("total voting power overflows uint64")
		}
		total += validator.VotingPower
	}
	if len(canonical) > 0 && total == 0 {
		return NetworkConcentration{}, ErrInvalidConcentration.Wrap("total voting power must be positive")
	}
	out := NetworkConcentration{
		Epoch:            epoch,
		TotalVotingPower: total,
		Validators:       []ValidatorConcentration{},
		Warnings:         []string{},
		RecomputedHeight: height,
	}
	rawBps := make([]uint32, 0, len(canonical))
	for _, validator := range canonical {
		raw := powerBps(validator.VotingPower, total)
		rawBps = append(rawBps, raw)
		effective := raw
		if effective > params.MaxVotingPowerBps {
			effective = params.MaxVotingPowerBps
		}
		modifier := rewardModifier(params, raw)
		warning := ""
		if raw >= params.WarningThresholdBps {
			warning = "concentration_warning"
		}
		if raw > params.MaxVotingPowerBps {
			warning = "hard_cap_exceeded"
		}
		metric := ValidatorConcentration{
			OperatorAddress:         validator.OperatorAddress,
			VotingPower:             validator.VotingPower,
			RawVotingPowerBps:       raw,
			EffectiveVotingPowerBps: effective,
			AboveSoftCap:            raw > params.SoftVotingPowerBps,
			AboveHardCap:            raw > params.MaxVotingPowerBps,
			DelegationAllowed:       !params.DelegationRejectionEnabled || raw < params.MaxVotingPowerBps,
			RewardModifierBps:       modifier,
			Warning:                 warning,
		}
		if warning != "" {
			out.Warnings = append(out.Warnings, validator.OperatorAddress+":"+warning)
		}
		out.Validators = append(out.Validators, metric)
		if effective > out.MaxValidatorPowerBps {
			out.MaxValidatorPowerBps = effective
		}
	}
	out.TopThreePowerBps = topNBps(rawBps, 3)
	if err := out.Validate(params); err != nil {
		return NetworkConcentration{}, ErrInvalidConcentration.Wrap(err.Error())
	}
	return out, nil
}

func SortValidatorConcentrations(in []ValidatorConcentration) []ValidatorConcentration {
	out := append([]ValidatorConcentration(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].OperatorAddress < out[j].OperatorAddress })
	return out
}

func powerBps(power, total uint64) uint32 {
	if power == 0 || total == 0 {
		return 0
	}
	hi, lo := bits.Mul64(power, uint64(BasisPoints))
	q, _ := bits.Div64(hi, lo, total)
	return uint32(q)
}

func rewardModifier(params Params, rawBps uint32) uint32 {
	if rawBps <= params.SoftVotingPowerBps || params.MaxRewardReductionBps == 0 {
		return BasisPoints
	}
	denom := BasisPoints - params.SoftVotingPowerBps
	if denom == 0 {
		return BasisPoints - params.MaxRewardReductionBps
	}
	reduction := (rawBps - params.SoftVotingPowerBps) * params.MaxRewardReductionBps / denom
	if reduction > params.MaxRewardReductionBps {
		reduction = params.MaxRewardReductionBps
	}
	return BasisPoints - reduction
}

func topNBps(values []uint32, n int) uint32 {
	sorted := append([]uint32(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] > sorted[j] })
	var out uint32
	for i := 0; i < len(sorted) && i < n; i++ {
		out += sorted[i]
	}
	if out > BasisPoints {
		return BasisPoints
	}
	return out
}
