package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	DelegationStatusActive			= "active"
	DelegationStatusCommissionExceeded	= "commission_exceeded"

	RiskAppetiteConservative	= "conservative"
	RiskAppetiteBalanced		= "balanced"
	RiskAppetiteAggressive		= "aggressive"

	LockDurationFlexible	= "flexible"
	LockDurationEpoch	= "epoch"
	LockDurationLongTerm	= "long_term"

	RewardStrategyLiquid		= "liquid"
	RewardStrategyCompound		= "compound"
	RewardStrategyAutoRedelegate	= "auto_redelegate"

	RiskTrancheSenior	= "senior"
	RiskTrancheMezzanine	= "mezzanine"
	RiskTrancheJunior	= "junior"
	RiskTrancheFirstLoss	= "first_loss"
)

type DelegationRecord struct {
	Delegator		string
	Validator		string
	Amount			sdkmath.Int
	ActivationEpoch		uint64
	Status			string
	RiskAppetite		string
	CommissionTolerance	uint32
	LockDurationPreference	string
	RewardStrategy		string
	RiskTrancheOptional	string
	CreatedHeight		uint64
	UpdatedHeight		uint64
}

type DelegationPreferences struct {
	RiskAppetite		string
	CommissionTolerance	uint32
	LockDurationPreference	string
	RewardStrategy		string
	RiskTrancheOptional	string
}

type DelegationCapitalState struct {
	Records []DelegationRecord
}

type DelegationCommissionAlert struct {
	Delegator		string
	Validator		string
	PreviousStatus		string
	NewStatus		string
	CommissionToleranceBps	uint32
	CurrentCommissionBps	uint32
	Height			uint64
	RedelegationAdvisory	bool
}

type LockDurationRewardEligibility struct {
	Delegator			string
	Validator			string
	LockDurationPreference		string
	ProtocolUnbondingSeconds	uint64
	EffectiveUnbondingSeconds	uint64
	SlashableWindowEpochs		uint64
	RequiredSlashableWindowEpochs	uint64
	RewardMultiplierBps		uint32
	EligibleForRewardMultiplier	bool
	RedelegationKeepsRiskHistory	bool
}

func BuildDelegationRecord(params postypes.Params, requestedEpoch uint64, createdHeight uint64, delegator string, validator string, amount sdkmath.Int, preferences DelegationPreferences) (DelegationRecord, error) {
	activationEpoch, err := postypes.DelegationEffectiveElectionEpoch(params, requestedEpoch)
	if err != nil {
		return DelegationRecord{}, err
	}
	record := DelegationRecord{
		Delegator:		strings.TrimSpace(delegator),
		Validator:		strings.TrimSpace(validator),
		Amount:			amount,
		ActivationEpoch:	activationEpoch,
		Status:			DelegationStatusActive,
		RiskAppetite:		normalizeDefault(preferences.RiskAppetite, RiskAppetiteBalanced),
		CommissionTolerance:	preferences.CommissionTolerance,
		LockDurationPreference:	normalizeDefault(preferences.LockDurationPreference, LockDurationEpoch),
		RewardStrategy:		normalizeDefault(preferences.RewardStrategy, RewardStrategyLiquid),
		RiskTrancheOptional:	strings.TrimSpace(preferences.RiskTrancheOptional),
		CreatedHeight:		createdHeight,
		UpdatedHeight:		createdHeight,
	}
	return record, record.Validate(params)
}

func BuildDelegationRecordFromIntent(params postypes.Params, intent postypes.DelegationIntent, createdHeight uint64, preferences DelegationPreferences) (DelegationRecord, error) {
	if preferences.CommissionTolerance == 0 {
		preferences.CommissionTolerance = intent.MaxCommissionBps
	}
	record, err := BuildDelegationRecord(params, intent.RequestedEpoch, createdHeight, intent.NominatorID, intent.ValidatorID, intent.StakeNaet, preferences)
	if err != nil {
		return DelegationRecord{}, err
	}
	if record.CommissionTolerance < intent.MaxCommissionBps {
		return DelegationRecord{}, errors.New("delegation record commission tolerance cannot be below intent tolerance")
	}
	return record, nil
}

func (r DelegationRecord) Validate(params postypes.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validateEconomyToken("delegator", r.Delegator); err != nil {
		return err
	}
	if err := validateEconomyToken("validator", r.Validator); err != nil {
		return err
	}
	if !r.Amount.IsPositive() {
		return errors.New("delegation amount must be positive")
	}
	if r.ActivationEpoch == 0 {
		return errors.New("delegation activation epoch is required")
	}
	if !isDelegationStatus(r.Status) {
		return fmt.Errorf("unsupported delegation status %q", r.Status)
	}
	if !isRiskAppetite(r.RiskAppetite) {
		return fmt.Errorf("unsupported risk appetite %q", r.RiskAppetite)
	}
	if r.CommissionTolerance > params.MaxCommissionBps {
		return fmt.Errorf("commission tolerance must be <= %d bps", params.MaxCommissionBps)
	}
	if !isLockDurationPreference(r.LockDurationPreference) {
		return fmt.Errorf("unsupported lock duration preference %q", r.LockDurationPreference)
	}
	if !isRewardStrategy(r.RewardStrategy) {
		return fmt.Errorf("unsupported reward strategy %q", r.RewardStrategy)
	}
	if r.RiskTrancheOptional != "" && !isRiskTranche(r.RiskTrancheOptional) {
		return fmt.Errorf("unsupported risk tranche %q", r.RiskTrancheOptional)
	}
	if r.UpdatedHeight < r.CreatedHeight {
		return errors.New("delegation updated height cannot be before created height")
	}
	return nil
}

func (r DelegationRecord) ToDelegationIntent(requestedEpoch uint64) postypes.DelegationIntent {
	return postypes.DelegationIntent{
		NominatorID:		strings.TrimSpace(r.Delegator),
		ValidatorID:		strings.TrimSpace(r.Validator),
		StakeNaet:		r.Amount,
		RequestedEpoch:		requestedEpoch,
		MaxCommissionBps:	r.CommissionTolerance,
	}
}

func NewDelegationCapitalState(params postypes.Params, records []DelegationRecord) (DelegationCapitalState, error) {
	out := make([]DelegationRecord, len(records))
	seen := make(map[string]struct{}, len(records))
	for i, record := range records {
		record.Delegator = strings.TrimSpace(record.Delegator)
		record.Validator = strings.TrimSpace(record.Validator)
		record.Status = normalizeDefault(record.Status, DelegationStatusActive)
		record.RiskAppetite = strings.TrimSpace(record.RiskAppetite)
		record.LockDurationPreference = strings.TrimSpace(record.LockDurationPreference)
		record.RewardStrategy = strings.TrimSpace(record.RewardStrategy)
		record.RiskTrancheOptional = strings.TrimSpace(record.RiskTrancheOptional)
		if err := record.Validate(params); err != nil {
			return DelegationCapitalState{}, err
		}
		key := delegationRecordKey(record.Delegator, record.Validator, record.ActivationEpoch)
		if _, found := seen[key]; found {
			return DelegationCapitalState{}, fmt.Errorf("duplicate delegation record %s", key)
		}
		seen[key] = struct{}{}
		out[i] = record
	}
	sortDelegationRecords(out)
	return DelegationCapitalState{Records: out}, nil
}

func (s DelegationCapitalState) RecordsForDelegator(delegator string) []DelegationRecord {
	delegator = strings.TrimSpace(delegator)
	records := make([]DelegationRecord, 0)
	for _, record := range s.Records {
		if record.Delegator == delegator {
			records = append(records, record)
		}
	}
	sortDelegationRecords(records)
	return records
}

func (s DelegationCapitalState) RecordsForValidator(validator string) []DelegationRecord {
	validator = strings.TrimSpace(validator)
	records := make([]DelegationRecord, 0)
	for _, record := range s.Records {
		if record.Validator == validator {
			records = append(records, record)
		}
	}
	sortDelegationRecords(records)
	return records
}

func (s DelegationCapitalState) TotalDelegatedToValidator(validator string) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, record := range s.RecordsForValidator(validator) {
		total = total.Add(record.Amount)
	}
	return total
}

func CheckCommissionTolerance(params postypes.Params, record DelegationRecord, currentCommissionBps uint32, height uint64, emitRedelegationAlert bool) (DelegationRecord, *DelegationCommissionAlert, error) {
	if err := record.Validate(params); err != nil {
		return DelegationRecord{}, nil, err
	}
	if currentCommissionBps > params.MaxCommissionBps {
		return DelegationRecord{}, nil, fmt.Errorf("current commission must be <= %d bps", params.MaxCommissionBps)
	}
	updated := record
	updated.UpdatedHeight = height
	if currentCommissionBps <= record.CommissionTolerance {
		updated.Status = DelegationStatusActive
		return updated, nil, nil
	}
	previous := record.Status
	updated.Status = DelegationStatusCommissionExceeded
	alert := &DelegationCommissionAlert{
		Delegator:		record.Delegator,
		Validator:		record.Validator,
		PreviousStatus:		previous,
		NewStatus:		updated.Status,
		CommissionToleranceBps:	record.CommissionTolerance,
		CurrentCommissionBps:	currentCommissionBps,
		Height:			height,
		RedelegationAdvisory:	emitRedelegationAlert,
	}
	return updated, alert, nil
}

func EvaluateLockDurationPreference(params postypes.Params, record DelegationRecord, slashableWindowEpochs uint64) (LockDurationRewardEligibility, error) {
	if err := record.Validate(params); err != nil {
		return LockDurationRewardEligibility{}, err
	}
	if slashableWindowEpochs == 0 {
		return LockDurationRewardEligibility{}, errors.New("slashable window must be positive")
	}
	requiredWindow := params.EvidenceWindowEpochs
	effectiveUnbonding := params.UnbondingSeconds
	multiplier := uint32(postypes.BasisPoints)
	eligible := false
	switch record.LockDurationPreference {
	case LockDurationFlexible:
		effectiveUnbonding = params.UnbondingSeconds
	case LockDurationEpoch:
		effectiveUnbonding = params.UnbondingSeconds
	case LockDurationLongTerm:
		effectiveUnbonding = params.UnbondingSeconds * 2
		requiredWindow = params.EvidenceWindowEpochs * 2
		if slashableWindowEpochs >= requiredWindow {
			multiplier = 11_000
			eligible = true
		}
	}
	if effectiveUnbonding < postypes.MinUnbondingSeconds {
		return LockDurationRewardEligibility{}, errors.New("effective unbonding period cannot be below protocol minimum")
	}
	return LockDurationRewardEligibility{
		Delegator:			record.Delegator,
		Validator:			record.Validator,
		LockDurationPreference:		record.LockDurationPreference,
		ProtocolUnbondingSeconds:	params.UnbondingSeconds,
		EffectiveUnbondingSeconds:	effectiveUnbonding,
		SlashableWindowEpochs:		slashableWindowEpochs,
		RequiredSlashableWindowEpochs:	requiredWindow,
		RewardMultiplierBps:		multiplier,
		EligibleForRewardMultiplier:	eligible,
		RedelegationKeepsRiskHistory:	true,
	}, nil
}

func sortDelegationRecords(records []DelegationRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		left := records[i]
		right := records[j]
		if left.ActivationEpoch != right.ActivationEpoch {
			return left.ActivationEpoch < right.ActivationEpoch
		}
		if left.Validator != right.Validator {
			return left.Validator < right.Validator
		}
		return left.Delegator < right.Delegator
	})
}

func delegationRecordKey(delegator string, validator string, activationEpoch uint64) string {
	return fmt.Sprintf("%s/%s/%d", delegator, validator, activationEpoch)
}

func normalizeDefault(value string, defaultValue string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

func validateEconomyToken(fieldName string, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if strings.ContainsAny(value, "|\n\r\t") {
		return fmt.Errorf("%s contains unsupported separators", fieldName)
	}
	return nil
}

func isRiskAppetite(value string) bool {
	switch value {
	case RiskAppetiteConservative, RiskAppetiteBalanced, RiskAppetiteAggressive:
		return true
	default:
		return false
	}
}

func isDelegationStatus(value string) bool {
	switch value {
	case DelegationStatusActive, DelegationStatusCommissionExceeded:
		return true
	default:
		return false
	}
}

func isLockDurationPreference(value string) bool {
	switch value {
	case LockDurationFlexible, LockDurationEpoch, LockDurationLongTerm:
		return true
	default:
		return false
	}
}

func isRewardStrategy(value string) bool {
	switch value {
	case RewardStrategyLiquid, RewardStrategyCompound, RewardStrategyAutoRedelegate:
		return true
	default:
		return false
	}
}

func isRiskTranche(value string) bool {
	switch value {
	case RiskTrancheSenior, RiskTrancheMezzanine, RiskTrancheJunior, RiskTrancheFirstLoss:
		return true
	default:
		return false
	}
}
