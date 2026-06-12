package types

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
	"strings"
)

type Params struct {
	Authority			string		`json:"authority"`
	PowerCapSchedule		[]PowerCapPhase	`json:"power_cap_schedule"`
	CommissionFloorBps		uint32		`json:"commission_floor_bps"`
	CommissionMaxBps		uint32		`json:"commission_max_bps"`
	MaxCommissionChangeBps		uint32		`json:"max_commission_change_bps"`
	WarningThresholdBps		uint32		`json:"warning_threshold_bps"`
	OverflowRewardMultiplierBps	uint32		`json:"overflow_reward_multiplier_bps"`
	MaxRewardReductionBps		uint32		`json:"max_reward_reduction_bps"`
}

type PowerCapPhase struct {
	MinActiveValidators	uint32	`json:"min_active_validators"`
	MaxActiveValidators	uint32	`json:"max_active_validators"`
	PowerCapBps		uint32	`json:"power_cap_bps"`
}

type ValidatorStake struct {
	OperatorAddress		string	`json:"operator_address"`
	RawStake		uint64	`json:"raw_stake"`
	CommissionBps		uint32	`json:"commission_bps"`
	PreviousCommissionBps	uint32	`json:"previous_commission_bps"`
}

type ValidatorPolicy struct {
	OperatorAddress		string	`json:"operator_address"`
	RawStake		uint64	`json:"raw_stake"`
	EffectiveStake		uint64	`json:"effective_stake"`
	OverflowStake		uint64	`json:"overflow_stake"`
	RawPowerBps		uint32	`json:"raw_power_bps"`
	EffectivePowerBps	uint32	`json:"effective_power_bps"`
	PowerCapBps		uint32	`json:"power_cap_bps"`
	RewardMultiplierBps	uint32	`json:"reward_multiplier_bps"`
	DelegationWarning	string	`json:"delegation_warning"`
	CommissionAllowed	bool	`json:"commission_allowed"`
	CommissionViolation	string	`json:"commission_violation,omitempty"`
	WarningAcknowledged	bool	`json:"warning_acknowledged"`
}

type NetworkPolicy struct {
	Epoch			uint64			`json:"epoch"`
	ActiveValidators	uint32			`json:"active_validators"`
	TotalRawStake		uint64			`json:"total_raw_stake"`
	PowerCapBps		uint32			`json:"power_cap_bps"`
	Validators		[]ValidatorPolicy	`json:"validators"`
	Top10PowerBps		uint32			`json:"top_10_power_bps"`
	Top20PowerBps		uint32			`json:"top_20_power_bps"`
	Top33PowerBps		uint32			`json:"top_33_power_bps"`
	ConcentrationWarn	bool			`json:"concentration_warn"`
}

type ValidatorIdentityMetadata struct {
	OperatorAddress	string	`json:"operator_address"`
	Moniker		string	`json:"moniker"`
	Website		string	`json:"website,omitempty"`
	SecurityContact	string	`json:"security_contact,omitempty"`
	Details		string	`json:"details,omitempty"`
}

type WarningAcknowledgement struct {
	OperatorAddress	string	`json:"operator_address"`
	AcknowledgedAt	int64	`json:"acknowledged_at"`
	Warning		string	`json:"warning"`
}

type GenesisState struct {
	Params			Params				`json:"params"`
	Network			NetworkPolicy			`json:"network"`
	Identities		[]ValidatorIdentityMetadata	`json:"identities"`
	WarningAcknowledgements	[]WarningAcknowledgement	`json:"warning_acknowledgements"`
}

type MsgUpdateStakingPolicyParams struct {
	Authority	string	`json:"authority"`
	Params		Params	`json:"params"`
}

type MsgRegisterValidatorIdentity struct {
	Authority	string				`json:"authority"`
	Identity	ValidatorIdentityMetadata	`json:"identity"`
}

type MsgAcknowledgeConcentrationWarning struct {
	Authority	string	`json:"authority"`
	OperatorAddress	string	`json:"operator_address"`
	Warning		string	`json:"warning"`
	Height		int64	`json:"height"`
}

type QueryParamsRequest struct{}
type QueryParamsResponse struct{ Params Params }

type QueryValidatorEffectivePowerRequest struct{ OperatorAddress string }
type QueryValidatorEffectivePowerResponse struct {
	OperatorAddress		string
	EffectiveStake		uint64
	EffectivePowerBps	uint32
	PowerCapBps		uint32
}

type QueryValidatorStakeRequest struct{ OperatorAddress string }
type QueryValidatorStakeResponse struct {
	OperatorAddress		string
	RawStake		uint64
	EffectiveStake		uint64
	OverflowStake		uint64
	RawPowerBps		uint32
	EffectivePowerBps	uint32
}

type QueryTopNConcentrationRequest struct{ N uint32 }
type QueryTopNConcentrationResponse struct {
	N		uint32
	PowerBps	uint32
	TargetBps	uint32
	Exceeded	bool
}

type QueryValidatorRewardMultiplierRequest struct{ OperatorAddress string }
type QueryValidatorRewardMultiplierResponse struct {
	OperatorAddress		string
	RewardMultiplierBps	uint32
}

type QueryDelegationWarningStatusRequest struct{ OperatorAddress string }
type QueryDelegationWarningStatusResponse struct {
	OperatorAddress		string
	Warning			string
	WarningAcknowledged	bool
}

func DefaultParams(authority string) Params {
	return Params{
		Authority:	authority,
		PowerCapSchedule: []PowerCapPhase{
			{MinActiveValidators: 1, MaxActiveValidators: PhaseOneValidatorSetMax, PowerCapBps: PhaseOnePowerCapBps},
			{MinActiveValidators: PhaseOneValidatorSetMax + 1, MaxActiveValidators: PhaseTwoValidatorSetMax, PowerCapBps: PhaseTwoPowerCapBps},
			{MinActiveValidators: PhaseTwoValidatorSetMax + 1, MaxActiveValidators: 0, PowerCapBps: MatureSetPowerCapBps},
		},
		CommissionFloorBps:		300,
		CommissionMaxBps:		2_000,
		MaxCommissionChangeBps:		100,
		WarningThresholdBps:		PhaseTwoPowerCapBps,
		OverflowRewardMultiplierBps:	0,
		MaxRewardReductionBps:		3_000,
	}
}

func DefaultGenesisState(authority string) GenesisState {
	return GenesisState{
		Params:				DefaultParams(authority),
		Network:			NetworkPolicy{Validators: []ValidatorPolicy{}},
		Identities:			[]ValidatorIdentityMetadata{},
		WarningAcknowledgements:	[]WarningAcknowledgement{},
	}
}

func (p Params) Validate() error {
	if strings.TrimSpace(p.Authority) == "" {
		return errors.New("authority must be non-empty")
	}
	if len(p.PowerCapSchedule) == 0 {
		return errors.New("power cap schedule is required")
	}
	previousMax := uint32(0)
	for i, phase := range p.PowerCapSchedule {
		if phase.MinActiveValidators == 0 {
			return fmt.Errorf("power cap phase %d min validators must be positive", i)
		}
		if phase.MaxActiveValidators != 0 && phase.MaxActiveValidators < phase.MinActiveValidators {
			return fmt.Errorf("power cap phase %d max validators must be >= min validators", i)
		}
		if i > 0 && phase.MinActiveValidators != previousMax+1 {
			return fmt.Errorf("power cap schedule must be contiguous")
		}
		if phase.PowerCapBps == 0 || phase.PowerCapBps > BasisPoints {
			return fmt.Errorf("power cap phase %d power cap must be between 1 and %d", i, BasisPoints)
		}
		previousMax = phase.MaxActiveValidators
		if previousMax == 0 && i != len(p.PowerCapSchedule)-1 {
			return fmt.Errorf("open-ended power cap phase must be last")
		}
	}
	if p.CommissionFloorBps == 0 || p.CommissionFloorBps > p.CommissionMaxBps || p.CommissionMaxBps > BasisPoints {
		return fmt.Errorf("commission floor/max bounds are invalid")
	}
	if p.MaxCommissionChangeBps == 0 || p.MaxCommissionChangeBps > p.CommissionMaxBps {
		return fmt.Errorf("max commission change must be positive and <= max commission")
	}
	if p.WarningThresholdBps == 0 || p.WarningThresholdBps > BasisPoints {
		return fmt.Errorf("warning threshold must be between 1 and %d", BasisPoints)
	}
	if p.OverflowRewardMultiplierBps > BasisPoints {
		return fmt.Errorf("overflow reward multiplier cannot exceed %d", BasisPoints)
	}
	if p.MaxRewardReductionBps > BasisPoints {
		return fmt.Errorf("max reward reduction cannot exceed %d", BasisPoints)
	}
	return nil
}

func (v ValidatorStake) Validate(params Params) error {
	if strings.TrimSpace(v.OperatorAddress) == "" {
		return fmt.Errorf("operator address must be non-empty")
	}
	if v.RawStake == 0 {
		return fmt.Errorf("raw stake must be positive")
	}
	if v.CommissionBps < params.CommissionFloorBps || v.CommissionBps > params.CommissionMaxBps {
		return fmt.Errorf("commission outside configured floor/max")
	}
	if v.PreviousCommissionBps != 0 && absUint32Delta(v.CommissionBps, v.PreviousCommissionBps) > params.MaxCommissionChangeBps {
		return fmt.Errorf("commission change exceeds configured max")
	}
	return nil
}

func (v ValidatorIdentityMetadata) Validate() error {
	if strings.TrimSpace(v.OperatorAddress) == "" {
		return fmt.Errorf("operator address must be non-empty")
	}
	if strings.TrimSpace(v.Moniker) == "" {
		return fmt.Errorf("moniker must be non-empty")
	}
	return nil
}

func (a WarningAcknowledgement) Validate() error {
	if strings.TrimSpace(a.OperatorAddress) == "" {
		return fmt.Errorf("operator address must be non-empty")
	}
	if strings.TrimSpace(a.Warning) == "" {
		return fmt.Errorf("warning must be non-empty")
	}
	if a.AcknowledgedAt <= 0 {
		return fmt.Errorf("acknowledged_at must be positive")
	}
	return nil
}

func (n NetworkPolicy) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if n.ActiveValidators != uint32(len(n.Validators)) {
		return fmt.Errorf("active_validators must match validator count")
	}
	if n.PowerCapBps != 0 && n.PowerCapBps != EffectivePowerCapBps(params, n.ActiveValidators) {
		return fmt.Errorf("network power cap does not match schedule")
	}
	if !validatorPoliciesSorted(n.Validators) {
		return fmt.Errorf("validator policies must be sorted by operator address")
	}
	seen := map[string]struct{}{}
	total := uint64(0)
	for _, validator := range n.Validators {
		if strings.TrimSpace(validator.OperatorAddress) == "" {
			return fmt.Errorf("operator address must be non-empty")
		}
		if _, found := seen[validator.OperatorAddress]; found {
			return fmt.Errorf("duplicate validator %s", validator.OperatorAddress)
		}
		seen[validator.OperatorAddress] = struct{}{}
		if validator.EffectivePowerBps > validator.PowerCapBps {
			return fmt.Errorf("effective power exceeds validator cap")
		}
		if validator.EffectiveStake+validator.OverflowStake != validator.RawStake {
			return fmt.Errorf("validator stake accounting must conserve raw stake")
		}
		if validator.RewardMultiplierBps > BasisPoints {
			return fmt.Errorf("reward multiplier cannot exceed %d", BasisPoints)
		}
		if total > ^uint64(0)-validator.RawStake {
			return fmt.Errorf("total raw stake overflows")
		}
		total += validator.RawStake
	}
	if total != n.TotalRawStake {
		return fmt.Errorf("total raw stake does not match validators")
	}
	return nil
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.Network.Validate(gs.Params); err != nil {
		return err
	}
	identitySeen := map[string]struct{}{}
	for _, identity := range gs.Identities {
		if err := identity.Validate(); err != nil {
			return err
		}
		if _, found := identitySeen[identity.OperatorAddress]; found {
			return fmt.Errorf("duplicate validator identity %s", identity.OperatorAddress)
		}
		identitySeen[identity.OperatorAddress] = struct{}{}
	}
	ackSeen := map[string]struct{}{}
	for _, ack := range gs.WarningAcknowledgements {
		if err := ack.Validate(); err != nil {
			return err
		}
		key := ack.OperatorAddress + ":" + ack.Warning
		if _, found := ackSeen[key]; found {
			return fmt.Errorf("duplicate warning acknowledgement %s", key)
		}
		ackSeen[key] = struct{}{}
	}
	return nil
}

func EffectivePowerCapBps(params Params, activeValidators uint32) uint32 {
	for _, phase := range params.PowerCapSchedule {
		if activeValidators >= phase.MinActiveValidators && (phase.MaxActiveValidators == 0 || activeValidators <= phase.MaxActiveValidators) {
			return phase.PowerCapBps
		}
	}
	return params.PowerCapSchedule[len(params.PowerCapSchedule)-1].PowerCapBps
}

func ComputeNetworkPolicy(params Params, epoch uint64, validators []ValidatorStake, acknowledgements []WarningAcknowledgement) (NetworkPolicy, error) {
	if err := params.Validate(); err != nil {
		return NetworkPolicy{}, ErrInvalidParams.Wrap(err.Error())
	}
	if epoch == 0 {
		return NetworkPolicy{}, ErrInvalidPolicy.Wrap("epoch must be positive")
	}
	canonical := append([]ValidatorStake(nil), validators...)
	sort.Slice(canonical, func(i, j int) bool { return canonical[i].OperatorAddress < canonical[j].OperatorAddress })
	seen := map[string]struct{}{}
	total := uint64(0)
	for _, validator := range canonical {
		if err := validator.Validate(params); err != nil {
			return NetworkPolicy{}, ErrInvalidPolicy.Wrap(err.Error())
		}
		if _, found := seen[validator.OperatorAddress]; found {
			return NetworkPolicy{}, ErrInvalidPolicy.Wrapf("duplicate validator %s", validator.OperatorAddress)
		}
		seen[validator.OperatorAddress] = struct{}{}
		if total > ^uint64(0)-validator.RawStake {
			return NetworkPolicy{}, ErrInvalidPolicy.Wrap("total stake overflows")
		}
		total += validator.RawStake
	}
	if len(canonical) > 0 && total == 0 {
		return NetworkPolicy{}, ErrInvalidPolicy.Wrap("total stake must be positive")
	}
	capBps := EffectivePowerCapBps(params, uint32(len(canonical)))
	acknowledged := acknowledgementSet(acknowledgements)
	out := NetworkPolicy{
		Epoch:			epoch,
		ActiveValidators:	uint32(len(canonical)),
		TotalRawStake:		total,
		PowerCapBps:		capBps,
		Validators:		make([]ValidatorPolicy, 0, len(canonical)),
	}
	rawShares := make([]uint32, 0, len(canonical))
	for _, validator := range canonical {
		rawBps := powerBps(validator.RawStake, total)
		effectiveBps := rawBps
		if effectiveBps > capBps {
			effectiveBps = capBps
		}
		effectiveStake := bpsToPower(total, effectiveBps)
		if effectiveStake > validator.RawStake {
			effectiveStake = validator.RawStake
		}
		overflowStake := validator.RawStake - effectiveStake
		warning := DelegationWarningNone
		switch {
		case rawBps > capBps:
			warning = DelegationWarningOverloaded
		case rawBps >= params.WarningThresholdBps:
			warning = DelegationWarningNearCap
		}
		policy := ValidatorPolicy{
			OperatorAddress:	validator.OperatorAddress,
			RawStake:		validator.RawStake,
			EffectiveStake:		effectiveStake,
			OverflowStake:		overflowStake,
			RawPowerBps:		rawBps,
			EffectivePowerBps:	effectiveBps,
			PowerCapBps:		capBps,
			RewardMultiplierBps:	rewardMultiplierBps(params, rawBps, capBps),
			DelegationWarning:	warning,
			CommissionAllowed:	true,
			WarningAcknowledged:	acknowledged[validator.OperatorAddress+":"+warning],
		}
		out.Validators = append(out.Validators, policy)
		rawShares = append(rawShares, rawBps)
	}
	out.Top10PowerBps = TopNBps(rawShares, 10)
	out.Top20PowerBps = TopNBps(rawShares, 20)
	out.Top33PowerBps = TopNBps(rawShares, 33)
	out.ConcentrationWarn = out.Top10PowerBps >= Top10ConcentrationTargetBps ||
		out.Top20PowerBps >= Top20ConcentrationTargetBps ||
		out.Top33PowerBps >= Top33ConcentrationTargetBps
	if err := out.Validate(params); err != nil {
		return NetworkPolicy{}, ErrInvalidPolicy.Wrap(err.Error())
	}
	return out, nil
}

func TopNBps(values []uint32, n uint32) uint32 {
	if n == 0 {
		return 0
	}
	sorted := append([]uint32(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] > sorted[j] })
	var out uint32
	for i := 0; i < len(sorted) && uint32(i) < n; i++ {
		out += sorted[i]
		if out > BasisPoints {
			return BasisPoints
		}
	}
	return out
}

func TopNTargetBps(n uint32) uint32 {
	switch {
	case n <= 10:
		return Top10ConcentrationTargetBps
	case n <= 20:
		return Top20ConcentrationTargetBps
	default:
		return Top33ConcentrationTargetBps
	}
}

func SortIdentities(in []ValidatorIdentityMetadata) []ValidatorIdentityMetadata {
	out := append([]ValidatorIdentityMetadata(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].OperatorAddress < out[j].OperatorAddress })
	return out
}

func SortAcknowledgements(in []WarningAcknowledgement) []WarningAcknowledgement {
	out := append([]WarningAcknowledgement(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].OperatorAddress == out[j].OperatorAddress {
			return out[i].Warning < out[j].Warning
		}
		return out[i].OperatorAddress < out[j].OperatorAddress
	})
	return out
}

func powerBps(power, total uint64) uint32 {
	if power == 0 || total == 0 {
		return 0
	}
	hi, lo := bits.Mul64(power, uint64(BasisPoints))
	q, _ := bits.Div64(hi, lo, total)
	if q > uint64(BasisPoints) {
		return BasisPoints
	}
	return uint32(q)
}

func bpsToPower(total uint64, bps uint32) uint64 {
	if total == 0 || bps == 0 {
		return 0
	}
	hi, lo := bits.Mul64(total, uint64(bps))
	q, _ := bits.Div64(hi, lo, uint64(BasisPoints))
	return q
}

func rewardMultiplierBps(params Params, rawBps, capBps uint32) uint32 {
	if rawBps <= capBps {
		return BasisPoints
	}
	if capBps >= BasisPoints {
		return BasisPoints - params.MaxRewardReductionBps
	}
	over := rawBps - capBps
	denom := BasisPoints - capBps
	reduction := uint32(uint64(over) * uint64(params.MaxRewardReductionBps) / uint64(denom))
	if reduction > params.MaxRewardReductionBps {
		reduction = params.MaxRewardReductionBps
	}
	multiplier := BasisPoints - reduction
	if multiplier < params.OverflowRewardMultiplierBps {
		return params.OverflowRewardMultiplierBps
	}
	return multiplier
}

func validatorPoliciesSorted(values []ValidatorPolicy) bool {
	return sort.SliceIsSorted(values, func(i, j int) bool { return values[i].OperatorAddress < values[j].OperatorAddress })
}

func acknowledgementSet(acks []WarningAcknowledgement) map[string]bool {
	out := map[string]bool{}
	for _, ack := range acks {
		out[ack.OperatorAddress+":"+ack.Warning] = true
	}
	return out
}

func absUint32Delta(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
