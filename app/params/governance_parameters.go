package params

import (
	"fmt"
	"sort"
)

const (
	GovernanceParamValidatorSetSize     = "validator_set_size"
	GovernanceParamValidatorPowerCap    = "validator_power_cap_bps"
	GovernanceParamCommissionFloor      = "commission_floor_bps"
	GovernanceParamCommissionMax        = "commission_max_bps"
	GovernanceParamCommissionMaxChange  = "commission_max_change_bps"
	GovernanceParamInflationMin         = "inflation_min_bps"
	GovernanceParamInflationMax         = "inflation_max_bps"
	GovernanceParamTargetBondedRatio    = "target_bonded_ratio_bps"
	GovernanceParamFeeBurnShare         = "fee_burn_share_bps"
	GovernanceParamFeeRewardShare       = "fee_reward_share_bps"
	GovernanceParamFeeTreasuryShare     = "fee_treasury_share_bps"
	GovernanceParamDoubleSignSlash      = "double_sign_slash_bps"
	GovernanceParamDowntimeSlash        = "downtime_slash_bps"
	GovernanceParamDowntimeWindow       = "downtime_window_blocks"
	GovernanceParamCosmWasmUploadPolicy = "cosmwasm_upload_policy"
	GovernanceParamTreasurySpendPolicy  = "treasury_spend_policy"
	GovernanceParamValidatorScorePolicy = "validator_score_policy"
	GovernanceParamValidatorSetGrowth   = "validator_set_growth_schedule"
	GovernanceParamBlockGasLimit        = "block_gas_limit"
	GovernanceParamBlockMaxBytes        = "block_max_bytes"

	GovernanceValueTypeInteger = "integer"
	GovernanceValueTypeEnum    = "enum"

	CosmWasmUploadDisabled       = "disabled"
	CosmWasmUploadGovernanceOnly = "governance_only"
	CosmWasmUploadPermissioned   = "permissioned"

	TreasurySpendDisabled       = "disabled"
	TreasurySpendGovernanceOnly = "governance_only"
	TreasurySpendBudgetCapped   = "budget_capped"

	ValidatorScorePolicyInformationalOnly = "informational_only"
	ValidatorScorePolicyObjectiveRewards  = "objective_reward_modifier"

	ValidatorSetGrowthPaused  = "paused"
	ValidatorSetGrowthGradual = "gradual"
	ValidatorSetGrowthMature  = "mature_cap"

	GovernanceDefaultVotingPeriodBlocks  = uint64(10_000)
	GovernanceCriticalVotingPeriodBlocks = uint64(20_000)
	GovernanceDefaultQuorumBps           = int64(4_000)
	GovernanceCriticalQuorumBps          = int64(5_000)
	GovernanceValidatorPowerCapMinBps    = int64(200)
	GovernanceValidatorPowerCapMaxBps    = int64(300)
	GovernanceAuthorityGovModule         = "gov"
	GovernanceDefaultBlockGasLimit       = int64(20_000_000)
	GovernanceMaxBlockGasLimit           = int64(80_000_000)
	GovernanceDefaultBlockMaxBytes       = int64(1_048_576)
	GovernanceMaxBlockMaxBytes           = int64(4_194_304)
)

type GovernanceParameterSpec struct {
	Key               string
	Category          string
	ValueType         string
	DefaultInt        int64
	DefaultString     string
	MinInt            int64
	MaxInt            int64
	AllowedValues     []string
	Authority         string
	ApplyEpochDelay   bool
	EventType         string
	InvalidUpdateTest bool
	Critical          bool
	GenesisRequired   bool
	ExecutionBounded  bool
	EmitsEvents       bool
}

type GovernanceParamValue struct {
	Key         string
	IntValue    int64
	StringValue string
}

type GovernanceParamChange struct {
	Value              GovernanceParamValue
	VotingPeriodBlocks uint64
	QuorumBps          int64
	ProposalExecution  bool
	EmitsEvent         bool
}

type GovernanceParameterSafetyReport struct {
	Specs             []GovernanceParameterSpec
	AllBounded        bool
	AllGenesisChecked bool
	AllEmitEvents     bool
	CriticalProtected bool
	Failed            []string
}

func DefaultGovernanceParameterSpecs() []GovernanceParameterSpec {
	return []GovernanceParameterSpec{
		governanceIntegerSpec(GovernanceParamValidatorSetSize, "staking_policy", AetraValidatorSetGenesisMin, AetraValidatorSetMin, AetraValidatorSetMax, true),
		governanceIntegerSpec(GovernanceParamValidatorPowerCap, "staking_policy", GovernanceValidatorPowerCapMaxBps, GovernanceValidatorPowerCapMinBps, GovernanceValidatorPowerCapMaxBps, true),
		governanceIntegerSpec(GovernanceParamCommissionFloor, "staking_policy", StakingCommissionFloorBps, StakingCommissionFloorBps, 500, false),
		governanceIntegerSpec(GovernanceParamCommissionMax, "staking_policy", StakingCommissionCeilingBps, 1_500, 2_000, false),
		governanceIntegerSpec(GovernanceParamCommissionMaxChange, "staking_policy", StakingMaxDailyCommissionBps, 50, StakingMaxDailyCommissionBps, false),
		governanceIntegerSpec(GovernanceParamInflationMin, "economics", MinInflationBps, 150, 200, true),
		governanceIntegerSpec(GovernanceParamInflationMax, "economics", MaxInflationBps, 500, 600, true),
		governanceIntegerSpec(GovernanceParamTargetBondedRatio, "economics", AetraTargetBondedRatioDefaultBps, AetraTargetBondedRatioMinBps, AetraTargetBondedRatioMaxBps, true),
		governanceIntegerSpec(GovernanceParamFeeBurnShare, "economics", AetraFeeBurnShareMinBps, AetraFeeBurnShareMinBps, AetraFeeBurnShareMaxBps, true),
		governanceIntegerSpec(GovernanceParamFeeRewardShare, "economics", AetraFeeRewardShareMinBps, AetraFeeRewardShareMinBps, AetraFeeRewardShareMaxBps, true),
		governanceIntegerSpec(GovernanceParamFeeTreasuryShare, "economics", AetraFeeTreasuryShareMinBps, AetraFeeTreasuryShareMinBps, AetraFeeTreasuryShareMaxBps, true),
		governanceEnumSpec(GovernanceParamValidatorScorePolicy, "validator_score", ValidatorScorePolicyInformationalOnly, true, ValidatorScorePolicyInformationalOnly, ValidatorScorePolicyObjectiveRewards),
		governanceIntegerSpec(GovernanceParamDoubleSignSlash, "slashing", DoubleSignSlashDefaultBps, DoubleSignSlashMinBps, DoubleSignSlashMaxBps, true),
		governanceIntegerSpec(GovernanceParamDowntimeSlash, "slashing", DowntimeFirstSlashDefaultBps, DowntimeFirstSlashMinBps, DowntimeChronicSlashMaxBps, true),
		governanceIntegerSpec(GovernanceParamDowntimeWindow, "slashing", int64(HeightUnbondingEvidenceWindowBlocks), 1_000, 100_000, true),
		governanceEnumSpec(GovernanceParamCosmWasmUploadPolicy, "vm", CosmWasmUploadGovernanceOnly, true, CosmWasmUploadDisabled, CosmWasmUploadGovernanceOnly, CosmWasmUploadPermissioned),
		governanceEnumSpec(GovernanceParamTreasurySpendPolicy, "treasury", TreasurySpendGovernanceOnly, true, TreasurySpendDisabled, TreasurySpendGovernanceOnly, TreasurySpendBudgetCapped),
		governanceEnumSpec(GovernanceParamValidatorSetGrowth, "validator_set_growth", ValidatorSetGrowthGradual, true, ValidatorSetGrowthPaused, ValidatorSetGrowthGradual, ValidatorSetGrowthMature),
		governanceIntegerSpec(GovernanceParamBlockGasLimit, "consensus", GovernanceDefaultBlockGasLimit, 1_000_000, GovernanceMaxBlockGasLimit, true),
		governanceIntegerSpec(GovernanceParamBlockMaxBytes, "consensus", GovernanceDefaultBlockMaxBytes, 220_200, GovernanceMaxBlockMaxBytes, true),
	}
}

func DefaultGovernanceGenesisParams() []GovernanceParamValue {
	return []GovernanceParamValue{
		{Key: GovernanceParamValidatorSetSize, IntValue: AetraValidatorSetGenesisMin},
		{Key: GovernanceParamValidatorPowerCap, IntValue: GovernanceValidatorPowerCapMaxBps},
		{Key: GovernanceParamCommissionFloor, IntValue: StakingCommissionFloorBps},
		{Key: GovernanceParamCommissionMax, IntValue: StakingCommissionCeilingBps},
		{Key: GovernanceParamCommissionMaxChange, IntValue: StakingMaxDailyCommissionBps},
		{Key: GovernanceParamInflationMin, IntValue: MinInflationBps},
		{Key: GovernanceParamInflationMax, IntValue: MaxInflationBps},
		{Key: GovernanceParamTargetBondedRatio, IntValue: AetraTargetBondedRatioDefaultBps},
		{Key: GovernanceParamFeeBurnShare, IntValue: AetraFeeBurnShareMinBps},
		{Key: GovernanceParamFeeRewardShare, IntValue: AetraFeeRewardShareMinBps},
		{Key: GovernanceParamFeeTreasuryShare, IntValue: AetraFeeTreasuryShareMinBps},
		{Key: GovernanceParamValidatorScorePolicy, StringValue: ValidatorScorePolicyInformationalOnly},
		{Key: GovernanceParamDoubleSignSlash, IntValue: DoubleSignSlashDefaultBps},
		{Key: GovernanceParamDowntimeSlash, IntValue: DowntimeFirstSlashDefaultBps},
		{Key: GovernanceParamDowntimeWindow, IntValue: int64(HeightUnbondingEvidenceWindowBlocks)},
		{Key: GovernanceParamCosmWasmUploadPolicy, StringValue: CosmWasmUploadGovernanceOnly},
		{Key: GovernanceParamTreasurySpendPolicy, StringValue: TreasurySpendGovernanceOnly},
		{Key: GovernanceParamValidatorSetGrowth, StringValue: ValidatorSetGrowthGradual},
		{Key: GovernanceParamBlockGasLimit, IntValue: GovernanceDefaultBlockGasLimit},
		{Key: GovernanceParamBlockMaxBytes, IntValue: GovernanceDefaultBlockMaxBytes},
	}
}

func ValidateGovernanceParameterSpecs(specs []GovernanceParameterSpec) error {
	report := BuildGovernanceParameterSafetyReport(specs)
	if len(report.Failed) > 0 {
		return fmt.Errorf("governance parameter safety failed: %v", report.Failed)
	}
	return nil
}

func BuildGovernanceParameterSafetyReport(specs []GovernanceParameterSpec) GovernanceParameterSafetyReport {
	if specs == nil {
		specs = DefaultGovernanceParameterSpecs()
	}
	specs = normalizeGovernanceSpecs(specs)
	required := requiredGovernanceParameterKeys()
	seen := map[string]GovernanceParameterSpec{}
	failed := make([]string, 0)
	allBounded := true
	allGenesisChecked := true
	allEmitEvents := true
	criticalProtected := true

	for _, spec := range specs {
		if spec.Key == "" {
			failed = append(failed, "governance_param_key_required")
			continue
		}
		if _, duplicate := seen[spec.Key]; duplicate {
			failed = append(failed, spec.Key+":duplicate")
		}
		seen[spec.Key] = spec
		if !required[spec.Key] {
			failed = append(failed, spec.Key+":unknown")
		}
		if !spec.ExecutionBounded {
			allBounded = false
			failed = append(failed, spec.Key+":proposal_execution_not_bounded")
		}
		if !spec.GenesisRequired {
			allGenesisChecked = false
			failed = append(failed, spec.Key+":genesis_validation_missing")
		}
		if !spec.EmitsEvents {
			allEmitEvents = false
			failed = append(failed, spec.Key+":event_missing")
		}
		if spec.Authority == "" {
			failed = append(failed, spec.Key+":authority_missing")
		}
		if spec.EventType == "" {
			allEmitEvents = false
			failed = append(failed, spec.Key+":event_type_missing")
		}
		if !spec.InvalidUpdateTest {
			failed = append(failed, spec.Key+":invalid_update_test_missing")
		}
		if spec.Critical && !spec.ApplyEpochDelay {
			criticalProtected = false
			failed = append(failed, spec.Key+":critical_param_must_apply_at_epoch_boundary")
		}
		if err := spec.ValidateValueSpec(); err != nil {
			allBounded = false
			failed = append(failed, spec.Key+":"+err.Error())
		}
		if spec.Critical && (!spec.ExecutionBounded || !spec.GenesisRequired || !spec.EmitsEvents) {
			criticalProtected = false
		}
	}
	for key := range required {
		if _, ok := seen[key]; !ok {
			failed = append(failed, key+":missing")
		}
	}
	sort.Strings(failed)
	return GovernanceParameterSafetyReport{
		Specs:             specs,
		AllBounded:        allBounded,
		AllGenesisChecked: allGenesisChecked,
		AllEmitEvents:     allEmitEvents,
		CriticalProtected: criticalProtected,
		Failed:            failed,
	}
}

func ValidateGovernanceParamChange(specs []GovernanceParameterSpec, change GovernanceParamChange) error {
	spec, err := findGovernanceSpec(specs, change.Value.Key)
	if err != nil {
		return err
	}
	if !change.ProposalExecution {
		return fmt.Errorf("governance parameter change must be validated at proposal execution")
	}
	if !change.EmitsEvent {
		return fmt.Errorf("governance parameter change must emit events")
	}
	if err := spec.ValidateValue(change.Value); err != nil {
		return err
	}
	if spec.Critical {
		if change.VotingPeriodBlocks < GovernanceCriticalVotingPeriodBlocks {
			return fmt.Errorf("critical governance parameter changes require longer voting period")
		}
		if change.QuorumBps < GovernanceCriticalQuorumBps {
			return fmt.Errorf("critical governance parameter changes require higher quorum")
		}
	}
	return nil
}

func ValidateGovernanceGenesisParams(specs []GovernanceParameterSpec, values []GovernanceParamValue) error {
	if err := ValidateGovernanceParameterSpecs(specs); err != nil {
		return err
	}
	if specs == nil {
		specs = DefaultGovernanceParameterSpecs()
	}
	seen := map[string]GovernanceParamValue{}
	for _, value := range values {
		if _, duplicate := seen[value.Key]; duplicate {
			return fmt.Errorf("duplicate genesis governance parameter %q", value.Key)
		}
		seen[value.Key] = value
		if _, err := findGovernanceSpec(specs, value.Key); err != nil {
			return err
		}
	}
	for _, spec := range specs {
		if !spec.GenesisRequired {
			continue
		}
		value, ok := seen[spec.Key]
		if !ok {
			return fmt.Errorf("genesis governance parameter %q is required", spec.Key)
		}
		if err := spec.ValidateValue(value); err != nil {
			return err
		}
	}
	return nil
}

func (s GovernanceParameterSpec) ValidateValueSpec() error {
	switch s.ValueType {
	case GovernanceValueTypeInteger:
		if s.MinInt > s.MaxInt {
			return fmt.Errorf("integer bounds are invalid")
		}
		if s.DefaultInt < s.MinInt || s.DefaultInt > s.MaxInt {
			return fmt.Errorf("default integer value is outside bounds")
		}
	case GovernanceValueTypeEnum:
		if len(s.AllowedValues) == 0 {
			return fmt.Errorf("enum allowed values are required")
		}
		if s.DefaultString == "" {
			return fmt.Errorf("default enum value is required")
		}
		found := false
		for _, allowed := range s.AllowedValues {
			if s.DefaultString == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default enum value must be allowed")
		}
	default:
		return fmt.Errorf("value type is invalid")
	}
	return nil
}

func (s GovernanceParameterSpec) ValidateValue(value GovernanceParamValue) error {
	if value.Key != s.Key {
		return fmt.Errorf("governance parameter key mismatch")
	}
	switch s.ValueType {
	case GovernanceValueTypeInteger:
		if value.IntValue < s.MinInt || value.IntValue > s.MaxInt {
			return fmt.Errorf("%s must stay within %d-%d", s.Key, s.MinInt, s.MaxInt)
		}
	case GovernanceValueTypeEnum:
		for _, allowed := range s.AllowedValues {
			if value.StringValue == allowed {
				return nil
			}
		}
		return fmt.Errorf("%s must use an allowed policy value", s.Key)
	default:
		return fmt.Errorf("value type is invalid")
	}
	return nil
}

func governanceIntegerSpec(key, category string, defaultValue, minValue, maxValue int64, critical bool) GovernanceParameterSpec {
	return GovernanceParameterSpec{
		Key:               key,
		Category:          category,
		ValueType:         GovernanceValueTypeInteger,
		DefaultInt:        defaultValue,
		MinInt:            minValue,
		MaxInt:            maxValue,
		Authority:         GovernanceAuthorityGovModule,
		ApplyEpochDelay:   critical,
		EventType:         "governance.param_changed",
		InvalidUpdateTest: true,
		Critical:          critical,
		GenesisRequired:   true,
		ExecutionBounded:  true,
		EmitsEvents:       true,
	}
}

func governanceEnumSpec(key, category string, defaultValue string, critical bool, allowedValues ...string) GovernanceParameterSpec {
	values := append([]string{}, allowedValues...)
	sort.Strings(values)
	return GovernanceParameterSpec{
		Key:               key,
		Category:          category,
		ValueType:         GovernanceValueTypeEnum,
		DefaultString:     defaultValue,
		AllowedValues:     values,
		Authority:         GovernanceAuthorityGovModule,
		ApplyEpochDelay:   critical,
		EventType:         "governance.param_changed",
		InvalidUpdateTest: true,
		Critical:          critical,
		GenesisRequired:   true,
		ExecutionBounded:  true,
		EmitsEvents:       true,
	}
}

func findGovernanceSpec(specs []GovernanceParameterSpec, key string) (GovernanceParameterSpec, error) {
	if specs == nil {
		specs = DefaultGovernanceParameterSpecs()
	}
	for _, spec := range specs {
		if spec.Key == key {
			return spec, nil
		}
	}
	return GovernanceParameterSpec{}, fmt.Errorf("unknown governance parameter %q", key)
}

func normalizeGovernanceSpecs(specs []GovernanceParameterSpec) []GovernanceParameterSpec {
	out := append([]GovernanceParameterSpec{}, specs...)
	for i := range out {
		out[i].AllowedValues = append([]string{}, out[i].AllowedValues...)
		sort.Strings(out[i].AllowedValues)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func requiredGovernanceParameterKeys() map[string]bool {
	return map[string]bool{
		GovernanceParamValidatorSetSize:     true,
		GovernanceParamValidatorPowerCap:    true,
		GovernanceParamCommissionFloor:      true,
		GovernanceParamCommissionMax:        true,
		GovernanceParamCommissionMaxChange:  true,
		GovernanceParamInflationMin:         true,
		GovernanceParamInflationMax:         true,
		GovernanceParamTargetBondedRatio:    true,
		GovernanceParamFeeBurnShare:         true,
		GovernanceParamFeeRewardShare:       true,
		GovernanceParamFeeTreasuryShare:     true,
		GovernanceParamValidatorScorePolicy: true,
		GovernanceParamDoubleSignSlash:      true,
		GovernanceParamDowntimeSlash:        true,
		GovernanceParamDowntimeWindow:       true,
		GovernanceParamCosmWasmUploadPolicy: true,
		GovernanceParamTreasurySpendPolicy:  true,
		GovernanceParamValidatorSetGrowth:   true,
		GovernanceParamBlockGasLimit:        true,
		GovernanceParamBlockMaxBytes:        true,
	}
}
