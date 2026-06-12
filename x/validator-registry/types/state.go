package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	StatusCandidate		= "candidate"
	StatusActive		= "active"
	StatusJailed		= "jailed"
	StatusTombstoned	= "tombstoned"
	StatusRetired		= "retired"

	HistoryRegistered		= "registered"
	HistoryMetadataUpdated		= "metadata-updated"
	HistoryConsensusRotated		= "consensus-key-rotated"
	HistoryWithdrawalUpdated	= "withdrawal-updated"
	HistoryTreasuryUpdated		= "treasury-updated"
	HistoryRetired			= "retired"
	HistoryCapabilitiesSet		= "capabilities-set"
	HistoryStatusChanged		= "status-changed"

	MaxValidatorsV1		= uint32(10_000)
	MaxMetadataBytesV1	= uint32(4_096)
	MaxConsensusKeyBytesV1	= uint32(512)
	MaxCapabilitiesV1	= uint32(64)
	MaxCapabilityBytesV1	= uint32(128)
	MaxHistoryEntriesV1	= uint32(512)
	MaxAuditFlagsV1		= uint32(32)
	MaxAuditFlagBytesV1	= uint32(128)
	MaxSlashReasonBytesV1	= uint32(256)
	MaxBasisPoints		= uint32(10_000)
	DefaultRotationDelay	= uint64(100)
	DefaultAETBaseUnits	= uint64(appparams.BaseUnitsPerDisplay)

	DefaultMinValidatorStake			= uint64(1_000_000) * DefaultAETBaseUnits
	DefaultSoloValidatorMinSelfStake		= uint64(1_000_000) * DefaultAETBaseUnits
	DefaultPoolBackedValidatorMinSelfStake		= uint64(400_000) * DefaultAETBaseUnits
	DefaultPoolBackedValidatorMaxNominatorStake	= uint64(600_000) * DefaultAETBaseUnits
	DefaultMinActiveValidators			= uint32(100)
	DefaultMaxActiveValidators			= uint32(300)
	DefaultValidatorCommissionFloorBps		= uint32(500)
	DefaultValidatorCommissionBps			= uint32(1_000)
	DefaultValidatorCommissionCeilingBps		= uint32(2_000)
	DefaultValidatorCommissionMaxDailyChangeBps	= uint32(100)
)

type Params struct {
	Authority			string
	MaxValidators			uint32
	MinValidatorStake		uint64
	SoloMinSelfStake		uint64
	PoolBackedMinSelfStake		uint64
	PoolBackedMaxNominatorStake	uint64
	MinActiveValidators		uint32
	MaxActiveValidators		uint32
	CommissionFloorBps		uint32
	DefaultCommissionBps		uint32
	CommissionCeilingBps		uint32
	CommissionMaxDailyChangeBps	uint32
	PowerCapSchedule		[]ValidatorPowerCapPhase
	MaxMetadataBytes		uint32
	MaxConsensusKeyBytes		uint32
	MaxCapabilities			uint32
	MaxCapabilityBytes		uint32
	MaxHistoryEntries		uint32
	MaxAuditFlags			uint32
	MaxAuditFlagBytes		uint32
	ConsensusKeyRotationDelay	uint64
}

type CommissionPolicy struct {
	CurrentRateBps		uint32
	MaxRateBps		uint32
	MaxChangeRateBps	uint32
}

type ValidatorPowerCapPhase struct {
	MaxActiveValidators	uint32
	PowerCapBps		uint32
}

type ValidatorFundingMode string

const (
	ValidatorFundingSolo		ValidatorFundingMode	= "solo"
	ValidatorFundingPoolBacked	ValidatorFundingMode	= "pool_backed"
)

type ValidatorFunding struct {
	Mode		ValidatorFundingMode
	SelfStake	uint64
	NominatorBond	uint64
}

type UptimeSample struct {
	Height		uint64
	UptimeBps	uint32
}

type LatencySample struct {
	Height		uint64
	LatencyMs	uint64
}

type SlashingEvent struct {
	Height		uint64
	FractionBps	uint32
	Reason		string
}

type ValidatorHistoryEvent struct {
	Height	uint64
	Type	string
	Detail	string
}

type ValidatorRecord struct {
	OperatorAddress			string
	ConsensusPublicKey		string
	PendingConsensusPublicKey	string
	ConsensusKeyActivationHeight	uint64
	TreasuryAddress			string
	WithdrawalAddress		string
	EmergencyAddress		string
	Metadata			string
	CommissionPolicy		CommissionPolicy
	UptimeHistory			[]UptimeSample
	LatencyHistory			[]LatencySample
	MissedBlockCounter		uint64
	SlashingHistory			[]SlashingEvent
	ReputationScore			uint32
	PerformanceScore		uint32
	Status				string
	Capabilities			[]string
	SelfBond			uint64
	NominatorBond			uint64
	ExternalAuditFlags		[]string
	History				[]ValidatorHistoryEvent
}

type State struct {
	Validators []ValidatorRecord
}

type ValidatorKeys struct {
	OperatorAddress			string
	ConsensusPublicKey		string
	PendingConsensusPublicKey	string
	ConsensusKeyActivationHeight	uint64
	TreasuryAddress			string
	WithdrawalAddress		string
	EmergencyAddress		string
}

type ValidatorPerformance struct {
	OperatorAddress		string
	UptimeHistory		[]UptimeSample
	LatencyHistory		[]LatencySample
	MissedBlockCounter	uint64
	ReputationScore		uint32
	PerformanceScore	uint32
}

type ValidatorSecurityStatus struct {
	OperatorAddress		string
	Status			string
	SlashingHistory		[]SlashingEvent
	ExternalAuditFlags	[]string
}

type ValidatorAllocationQueryRequest struct {
	IncludeCandidates	bool
	IncludeJailed		bool
	EnforceActiveBounds	bool
	TestnetOverride		bool
}

type ValidatorAllocationEngineInput struct {
	OperatorAddress		string
	ConsensusPublicKey	string
	Status			string
	SelfBond		uint64
	NominatorBond		uint64
	CommissionBps		uint32
	PerformanceScore	uint32
	ReputationScore		uint32
	PowerCapBps		uint32
	Jailed			bool
	Tombstoned		bool
}

type MsgRegisterValidator struct {
	Authority	string
	Validator	ValidatorRecord
	Height		uint64
}

type MsgUpdateValidatorMetadata struct {
	Authority	string
	OperatorAddress	string
	Metadata	string
	Height		uint64
}

type MsgRotateConsensusKey struct {
	Authority		string
	OperatorAddress		string
	NewConsensusPublicKey	string
	ActivationHeight	uint64
	Height			uint64
}

type MsgUpdateWithdrawalAddress struct {
	Authority		string
	OperatorAddress		string
	WithdrawalAddress	string
	Height			uint64
}

type MsgUpdateTreasuryAddress struct {
	Authority	string
	OperatorAddress	string
	TreasuryAddress	string
	Height		uint64
}

type MsgRetireValidator struct {
	Authority	string
	OperatorAddress	string
	Height		uint64
}

type MsgSetValidatorCapabilities struct {
	Authority	string
	OperatorAddress	string
	Capabilities	[]string
	Height		uint64
}

type MsgUpdateValidatorCommission struct {
	Authority	string
	OperatorAddress	string
	NewRateBps	uint32
	Height		uint64
}

func DefaultParams() Params {
	return Params{
		Authority:			prototype.DefaultAuthority,
		MaxValidators:			MaxValidatorsV1,
		MinValidatorStake:		DefaultMinValidatorStake,
		SoloMinSelfStake:		DefaultSoloValidatorMinSelfStake,
		PoolBackedMinSelfStake:		DefaultPoolBackedValidatorMinSelfStake,
		PoolBackedMaxNominatorStake:	DefaultPoolBackedValidatorMaxNominatorStake,
		MinActiveValidators:		DefaultMinActiveValidators,
		MaxActiveValidators:		DefaultMaxActiveValidators,
		CommissionFloorBps:		DefaultValidatorCommissionFloorBps,
		DefaultCommissionBps:		DefaultValidatorCommissionBps,
		CommissionCeilingBps:		DefaultValidatorCommissionCeilingBps,
		CommissionMaxDailyChangeBps:	DefaultValidatorCommissionMaxDailyChangeBps,
		PowerCapSchedule: []ValidatorPowerCapPhase{
			{MaxActiveValidators: 150, PowerCapBps: 300},
			{MaxActiveValidators: 250, PowerCapBps: 250},
			{MaxActiveValidators: 0, PowerCapBps: 200},
		},
		MaxMetadataBytes:		MaxMetadataBytesV1,
		MaxConsensusKeyBytes:		MaxConsensusKeyBytesV1,
		MaxCapabilities:		MaxCapabilitiesV1,
		MaxCapabilityBytes:		MaxCapabilityBytesV1,
		MaxHistoryEntries:		MaxHistoryEntriesV1,
		MaxAuditFlags:			MaxAuditFlagsV1,
		MaxAuditFlagBytes:		MaxAuditFlagBytesV1,
		ConsensusKeyRotationDelay:	DefaultRotationDelay,
	}
}

func DefaultCommissionPolicy() CommissionPolicy {
	return CommissionPolicy{
		CurrentRateBps:		DefaultValidatorCommissionBps,
		MaxRateBps:		DefaultValidatorCommissionCeilingBps,
		MaxChangeRateBps:	DefaultValidatorCommissionMaxDailyChangeBps,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("validator registry authority", p.Authority); err != nil {
		return err
	}
	if p.MaxValidators == 0 || p.MaxValidators > MaxValidatorsV1 {
		return fmt.Errorf("validator registry max validators must be between 1 and %d", MaxValidatorsV1)
	}
	if err := p.ValidateValidatorFunding(ValidatorFunding{Mode: ValidatorFundingSolo, SelfStake: p.SoloMinSelfStake}); err != nil {
		return err
	}
	if err := p.ValidateValidatorFunding(ValidatorFunding{Mode: ValidatorFundingPoolBacked, SelfStake: p.PoolBackedMinSelfStake, NominatorBond: p.PoolBackedMaxNominatorStake}); err != nil {
		return err
	}
	if p.MinActiveValidators == 0 || p.MaxActiveValidators == 0 || p.MinActiveValidators > p.MaxActiveValidators || p.MaxActiveValidators > p.MaxValidators {
		return errors.New("validator registry active validator bounds are invalid")
	}
	if err := validateCommissionParams(p); err != nil {
		return err
	}
	if err := validatePowerCapSchedule(p.PowerCapSchedule); err != nil {
		return err
	}
	if p.MaxMetadataBytes == 0 || p.MaxMetadataBytes > MaxMetadataBytesV1 {
		return fmt.Errorf("validator registry max metadata bytes must be between 1 and %d", MaxMetadataBytesV1)
	}
	if p.MaxConsensusKeyBytes == 0 || p.MaxConsensusKeyBytes > MaxConsensusKeyBytesV1 {
		return fmt.Errorf("validator registry max consensus key bytes must be between 1 and %d", MaxConsensusKeyBytesV1)
	}
	if p.MaxCapabilities == 0 || p.MaxCapabilities > MaxCapabilitiesV1 {
		return fmt.Errorf("validator registry max capabilities must be between 1 and %d", MaxCapabilitiesV1)
	}
	if p.MaxCapabilityBytes == 0 || p.MaxCapabilityBytes > MaxCapabilityBytesV1 {
		return fmt.Errorf("validator registry max capability bytes must be between 1 and %d", MaxCapabilityBytesV1)
	}
	if p.MaxHistoryEntries == 0 || p.MaxHistoryEntries > MaxHistoryEntriesV1 {
		return fmt.Errorf("validator registry max history entries must be between 1 and %d", MaxHistoryEntriesV1)
	}
	if p.MaxAuditFlags > MaxAuditFlagsV1 {
		return fmt.Errorf("validator registry max audit flags must be <= %d", MaxAuditFlagsV1)
	}
	if p.MaxAuditFlagBytes == 0 || p.MaxAuditFlagBytes > MaxAuditFlagBytesV1 {
		return fmt.Errorf("validator registry max audit flag bytes must be between 1 and %d", MaxAuditFlagBytesV1)
	}
	if p.ConsensusKeyRotationDelay == 0 {
		return errors.New("validator registry consensus key rotation delay must be positive")
	}
	return nil
}

func (p Params) ValidateValidatorFunding(funding ValidatorFunding) error {
	totalStake, err := checkedAddUint64(funding.SelfStake, funding.NominatorBond)
	if err != nil {
		return err
	}
	switch funding.Mode {
	case ValidatorFundingSolo:
		if funding.NominatorBond != 0 {
			return errors.New("validator registry solo validator cannot use nominator stake")
		}
		if funding.SelfStake < p.SoloMinSelfStake {
			return errors.New("validator registry solo validator self-stake below configured minimum")
		}
	case ValidatorFundingPoolBacked:
		if funding.SelfStake < p.PoolBackedMinSelfStake {
			return errors.New("validator registry pool-backed validator self-stake below configured minimum")
		}
		if funding.NominatorBond > p.PoolBackedMaxNominatorStake {
			return errors.New("validator registry pool-backed validator nominator contribution exceeds configured maximum")
		}
		if p.MinValidatorStake > funding.SelfStake && p.MinValidatorStake-funding.SelfStake > p.PoolBackedMaxNominatorStake {
			return errors.New("validator registry pool-backed validator cannot satisfy minimum stake with configured nominator cap")
		}
	default:
		return fmt.Errorf("validator registry unsupported validator funding mode %q", funding.Mode)
	}
	if totalStake < p.MinValidatorStake {
		return errors.New("validator registry validator below minimum validator stake")
	}
	return nil
}

func (p Params) ValidateActiveValidatorCount(count uint32, testnetOverride bool) error {
	if count > p.MaxActiveValidators {
		return errors.New("validator registry active validator count exceeds configured maximum")
	}
	if !testnetOverride && count < p.MinActiveValidators {
		return errors.New("validator registry active validator count below configured minimum")
	}
	return nil
}

func (p Params) ValidateCommissionRate(rateBps uint32) error {
	if rateBps < p.CommissionFloorBps {
		return errors.New("validator registry commission below configured floor")
	}
	if rateBps > p.CommissionCeilingBps {
		return errors.New("validator registry commission above configured ceiling")
	}
	return nil
}

func (p Params) ValidateCommissionChange(previousRateBps, nextRateBps uint32) error {
	if err := p.ValidateCommissionRate(nextRateBps); err != nil {
		return err
	}
	if nextRateBps > previousRateBps && nextRateBps-previousRateBps > p.CommissionMaxDailyChangeBps {
		return errors.New("validator registry commission daily change exceeds configured maximum")
	}
	return nil
}

func (p Params) PowerCapBpsForValidatorCount(count uint32) (uint32, error) {
	if len(p.PowerCapSchedule) == 0 {
		return 0, errors.New("validator registry power cap schedule is required")
	}
	for _, phase := range p.PowerCapSchedule {
		if phase.MaxActiveValidators == 0 || count <= phase.MaxActiveValidators {
			return phase.PowerCapBps, nil
		}
	}
	return 0, errors.New("validator registry power cap schedule did not cover validator count")
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("validator registry update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("validator registry update requires governance authority")
	}
	return nil
}

func (p CommissionPolicy) Validate() error {
	if p.CurrentRateBps > MaxBasisPoints || p.MaxRateBps > MaxBasisPoints || p.MaxChangeRateBps > MaxBasisPoints {
		return fmt.Errorf("validator registry commission policy must be <= %d bps", MaxBasisPoints)
	}
	if p.CurrentRateBps > p.MaxRateBps {
		return errors.New("validator registry current commission exceeds max commission")
	}
	if p.MaxChangeRateBps > p.MaxRateBps {
		return errors.New("validator registry max commission change exceeds max commission")
	}
	return nil
}

func (p Params) ValidateCommissionPolicy(policy CommissionPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if err := p.ValidateCommissionRate(policy.CurrentRateBps); err != nil {
		return err
	}
	if policy.MaxRateBps > p.CommissionCeilingBps {
		return errors.New("validator registry commission policy max exceeds configured ceiling")
	}
	if policy.MaxChangeRateBps > p.CommissionMaxDailyChangeBps {
		return errors.New("validator registry commission policy daily change exceeds configured maximum")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Validators)) > params.MaxValidators {
		return errors.New("validator registry validator limit exceeded")
	}
	byOperator := map[string]struct{}{}
	byConsensus := map[string]string{}
	activeCount := uint32(0)
	for _, validator := range s.Validators {
		normalized := validator.Normalize(params)
		if err := normalized.Validate(params); err != nil {
			return err
		}
		if normalized.Status == StatusActive {
			activeCount++
		}
		if _, found := byOperator[normalized.OperatorAddress]; found {
			return fmt.Errorf("validator registry duplicate operator %q", normalized.OperatorAddress)
		}
		byOperator[normalized.OperatorAddress] = struct{}{}
		for _, key := range normalized.ConsensusKeys() {
			if other, found := byConsensus[key]; found {
				return fmt.Errorf("validator registry duplicate consensus key used by %s and %s", other, normalized.OperatorAddress)
			}
			byConsensus[key] = normalized.OperatorAddress
		}
	}
	if activeCount > params.MaxActiveValidators {
		return errors.New("validator registry active validator count exceeds configured maximum")
	}
	return nil
}

func (v ValidatorRecord) Validate(params Params) error {
	if err := addressing.ValidateAuthorityAddress("validator operator address", v.OperatorAddress); err != nil {
		return err
	}
	if err := validateConsensusKey("validator consensus public key", v.ConsensusPublicKey, params.MaxConsensusKeyBytes); err != nil {
		return err
	}
	if strings.TrimSpace(v.PendingConsensusPublicKey) != "" {
		if err := validateConsensusKey("validator pending consensus public key", v.PendingConsensusPublicKey, params.MaxConsensusKeyBytes); err != nil {
			return err
		}
		if v.PendingConsensusPublicKey == v.ConsensusPublicKey {
			return errors.New("validator registry pending consensus key must differ from current key")
		}
		if v.ConsensusKeyActivationHeight == 0 {
			return errors.New("validator registry pending consensus key requires activation height")
		}
	} else if v.ConsensusKeyActivationHeight != 0 {
		return errors.New("validator registry activation height requires pending consensus key")
	}
	if strings.TrimSpace(v.ConsensusPublicKey) == strings.TrimSpace(v.OperatorAddress) {
		return errors.New("validator registry operator and consensus keys must be distinct roles")
	}
	for label, value := range map[string]string{
		"validator treasury address":	v.TreasuryAddress,
		"validator withdrawal address":	v.WithdrawalAddress,
		"validator emergency address":	v.EmergencyAddress,
	} {
		if err := addressing.ValidateAuthorityAddress(label, value); err != nil {
			return err
		}
		if strings.TrimSpace(v.ConsensusPublicKey) == strings.TrimSpace(value) {
			return errors.New("validator registry consensus key must not reuse account address")
		}
	}
	if uint32(len(v.Metadata)) > params.MaxMetadataBytes {
		return fmt.Errorf("validator registry metadata exceeds %d bytes", params.MaxMetadataBytes)
	}
	if err := params.ValidateCommissionPolicy(v.CommissionPolicy); err != nil {
		return err
	}
	mode := ValidatorFundingPoolBacked
	if v.NominatorBond == 0 {
		mode = ValidatorFundingSolo
	}
	if err := params.ValidateValidatorFunding(ValidatorFunding{Mode: mode, SelfStake: v.SelfBond, NominatorBond: v.NominatorBond}); err != nil {
		return err
	}
	if !IsStatus(v.Status) {
		return fmt.Errorf("validator registry status %q is invalid", v.Status)
	}
	if uint32(len(v.Capabilities)) > params.MaxCapabilities {
		return errors.New("validator registry capability limit exceeded")
	}
	if err := validateSortedUniqueTokens("validator capability", v.Capabilities, params.MaxCapabilityBytes); err != nil {
		return err
	}
	if uint32(len(v.ExternalAuditFlags)) > params.MaxAuditFlags {
		return errors.New("validator registry audit flag limit exceeded")
	}
	if err := validateSortedUniqueTokens("validator audit flag", v.ExternalAuditFlags, params.MaxAuditFlagBytes); err != nil {
		return err
	}
	if err := validateUptimeHistory(v.UptimeHistory, params.MaxHistoryEntries); err != nil {
		return err
	}
	if err := validateLatencyHistory(v.LatencyHistory, params.MaxHistoryEntries); err != nil {
		return err
	}
	if err := validateSlashingHistory(v.SlashingHistory, params.MaxHistoryEntries); err != nil {
		return err
	}
	if err := validateValidatorHistory(v.History, params.MaxHistoryEntries); err != nil {
		return err
	}
	if v.ReputationScore > MaxBasisPoints || v.PerformanceScore > MaxBasisPoints {
		return fmt.Errorf("validator registry scores must be <= %d", MaxBasisPoints)
	}
	return nil
}

func (v ValidatorRecord) Normalize(params Params) ValidatorRecord {
	v.OperatorAddress = strings.TrimSpace(v.OperatorAddress)
	v.ConsensusPublicKey = strings.TrimSpace(v.ConsensusPublicKey)
	v.PendingConsensusPublicKey = strings.TrimSpace(v.PendingConsensusPublicKey)
	v.TreasuryAddress = strings.TrimSpace(v.TreasuryAddress)
	v.WithdrawalAddress = strings.TrimSpace(v.WithdrawalAddress)
	v.EmergencyAddress = strings.TrimSpace(v.EmergencyAddress)
	v.Status = strings.TrimSpace(v.Status)
	if v.Status == "" {
		v.Status = StatusCandidate
	}
	if v.CommissionPolicy == (CommissionPolicy{}) {
		v.CommissionPolicy = DefaultCommissionPolicy()
	}
	v.Capabilities = sortedUniqueTokens(v.Capabilities)
	v.ExternalAuditFlags = sortedUniqueTokens(v.ExternalAuditFlags)
	v.UptimeHistory = sortedUptime(v.UptimeHistory)
	v.LatencyHistory = sortedLatency(v.LatencyHistory)
	v.SlashingHistory = sortedSlashing(v.SlashingHistory)
	v.History = sortedHistory(v.History)
	if uint32(len(v.UptimeHistory)) > params.MaxHistoryEntries {
		v.UptimeHistory = v.UptimeHistory[len(v.UptimeHistory)-int(params.MaxHistoryEntries):]
	}
	if uint32(len(v.LatencyHistory)) > params.MaxHistoryEntries {
		v.LatencyHistory = v.LatencyHistory[len(v.LatencyHistory)-int(params.MaxHistoryEntries):]
	}
	if uint32(len(v.SlashingHistory)) > params.MaxHistoryEntries {
		v.SlashingHistory = v.SlashingHistory[len(v.SlashingHistory)-int(params.MaxHistoryEntries):]
	}
	if uint32(len(v.History)) > params.MaxHistoryEntries {
		v.History = v.History[len(v.History)-int(params.MaxHistoryEntries):]
	}
	return v
}

func (v ValidatorRecord) ConsensusKeys() []string {
	keys := []string{v.ConsensusPublicKey}
	if strings.TrimSpace(v.PendingConsensusPublicKey) != "" {
		keys = append(keys, v.PendingConsensusPublicKey)
	}
	return keys
}

func (v ValidatorRecord) Keys() ValidatorKeys {
	return ValidatorKeys{
		OperatorAddress:		v.OperatorAddress,
		ConsensusPublicKey:		v.ConsensusPublicKey,
		PendingConsensusPublicKey:	v.PendingConsensusPublicKey,
		ConsensusKeyActivationHeight:	v.ConsensusKeyActivationHeight,
		TreasuryAddress:		v.TreasuryAddress,
		WithdrawalAddress:		v.WithdrawalAddress,
		EmergencyAddress:		v.EmergencyAddress,
	}
}

func (v ValidatorRecord) Performance() ValidatorPerformance {
	return ValidatorPerformance{
		OperatorAddress:	v.OperatorAddress,
		UptimeHistory:		append([]UptimeSample(nil), v.UptimeHistory...),
		LatencyHistory:		append([]LatencySample(nil), v.LatencyHistory...),
		MissedBlockCounter:	v.MissedBlockCounter,
		ReputationScore:	v.ReputationScore,
		PerformanceScore:	v.PerformanceScore,
	}
}

func (v ValidatorRecord) SecurityStatus() ValidatorSecurityStatus {
	return ValidatorSecurityStatus{
		OperatorAddress:	v.OperatorAddress,
		Status:			v.Status,
		SlashingHistory:	append([]SlashingEvent(nil), v.SlashingHistory...),
		ExternalAuditFlags:	append([]string(nil), v.ExternalAuditFlags...),
	}
}

func (s State) Normalize(params Params) State {
	out := State{Validators: make([]ValidatorRecord, 0, len(s.Validators))}
	for _, validator := range s.Validators {
		out.Validators = append(out.Validators, validator.Normalize(params))
	}
	out.Validators = SortValidators(out.Validators)
	return out
}

func (s State) Validator(operator string) (ValidatorRecord, bool) {
	operator = strings.TrimSpace(operator)
	for _, validator := range SortValidators(s.Validators) {
		if validator.OperatorAddress == operator {
			return validator, true
		}
	}
	return ValidatorRecord{}, false
}

func (s State) ValidatorAllocationEngineInputs(params Params, req ValidatorAllocationQueryRequest) ([]ValidatorAllocationEngineInput, error) {
	normalized := s.Normalize(params)
	activeCount := uint32(0)
	for _, validator := range normalized.Validators {
		if validator.Status == StatusActive {
			activeCount++
		}
	}
	if req.EnforceActiveBounds {
		if err := params.ValidateActiveValidatorCount(activeCount, req.TestnetOverride); err != nil {
			return nil, err
		}
	}
	powerCap, err := params.PowerCapBpsForValidatorCount(activeCount)
	if err != nil {
		return nil, err
	}
	out := make([]ValidatorAllocationEngineInput, 0, len(normalized.Validators))
	for _, validator := range normalized.Validators {
		switch validator.Status {
		case StatusActive:
		case StatusCandidate:
			if !req.IncludeCandidates {
				continue
			}
		case StatusJailed:
			if !req.IncludeJailed {
				continue
			}
		default:
			continue
		}
		out = append(out, ValidatorAllocationEngineInput{
			OperatorAddress:	validator.OperatorAddress,
			ConsensusPublicKey:	validator.ConsensusPublicKey,
			Status:			validator.Status,
			SelfBond:		validator.SelfBond,
			NominatorBond:		validator.NominatorBond,
			CommissionBps:		validator.CommissionPolicy.CurrentRateBps,
			PerformanceScore:	validator.PerformanceScore,
			ReputationScore:	validator.ReputationScore,
			PowerCapBps:		powerCap,
			Jailed:			validator.Status == StatusJailed,
			Tombstoned:		validator.Status == StatusTombstoned,
		})
	}
	return out, nil
}

func SortValidators(validators []ValidatorRecord) []ValidatorRecord {
	out := make([]ValidatorRecord, len(validators))
	copy(out, validators)
	sort.Slice(out, func(i, j int) bool {
		return out[i].OperatorAddress < out[j].OperatorAddress
	})
	return out
}

func UpsertValidator(validators []ValidatorRecord, validator ValidatorRecord) []ValidatorRecord {
	next := make([]ValidatorRecord, 0, len(validators)+1)
	replaced := false
	for _, current := range validators {
		if current.OperatorAddress == validator.OperatorAddress {
			next = append(next, validator)
			replaced = true
			continue
		}
		next = append(next, current)
	}
	if !replaced {
		next = append(next, validator)
	}
	return SortValidators(next)
}

func ValidStatusTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "":
		return to == StatusCandidate
	case StatusCandidate:
		return to == StatusActive || to == StatusJailed || to == StatusRetired || to == StatusTombstoned
	case StatusActive:
		return to == StatusJailed || to == StatusRetired || to == StatusTombstoned
	case StatusJailed:
		return to == StatusCandidate || to == StatusRetired || to == StatusTombstoned
	case StatusRetired, StatusTombstoned:
		return false
	default:
		return false
	}
}

func IsStatus(status string) bool {
	switch status {
	case StatusCandidate, StatusActive, StatusJailed, StatusTombstoned, StatusRetired:
		return true
	default:
		return false
	}
}

func AddHistory(v ValidatorRecord, height uint64, eventType, detail string, params Params) ValidatorRecord {
	if height == 0 {
		return v
	}
	v.History = append(v.History, ValidatorHistoryEvent{Height: height, Type: eventType, Detail: detail})
	return v.Normalize(params)
}

func validateConsensusKey(label, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s must be non-empty", label)
	}
	if uint32(len(value)) > maxBytes {
		return fmt.Errorf("%s exceeds %d bytes", label, maxBytes)
	}
	for _, ch := range value {
		if unicode.IsControl(ch) || unicode.IsSpace(ch) {
			return fmt.Errorf("%s contains invalid whitespace/control character", label)
		}
	}
	return nil
}

func validateSortedUniqueTokens(label string, values []string, maxBytes uint32) error {
	normalized := sortedUniqueTokens(values)
	if len(normalized) != len(values) {
		return fmt.Errorf("%s values must be unique", label)
	}
	for i, value := range values {
		if value != normalized[i] {
			return fmt.Errorf("%s values must be sorted deterministically", label)
		}
		if err := validateToken(label, value, maxBytes); err != nil {
			return err
		}
	}
	return nil
}

func validateToken(label, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s must be non-empty", label)
	}
	if uint32(len(value)) > maxBytes {
		return fmt.Errorf("%s exceeds %d bytes", label, maxBytes)
	}
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' || ch == ':' {
			continue
		}
		return fmt.Errorf("%s contains invalid character %q", label, ch)
	}
	return nil
}

func sortedUniqueTokens(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func validateUptimeHistory(values []UptimeSample, maxEntries uint32) error {
	if uint32(len(values)) > maxEntries {
		return errors.New("validator registry uptime history limit exceeded")
	}
	previous := uint64(0)
	for _, value := range values {
		if value.Height == 0 || value.Height < previous {
			return errors.New("validator registry uptime history must be sorted by positive height")
		}
		if value.UptimeBps > MaxBasisPoints {
			return fmt.Errorf("validator registry uptime must be <= %d bps", MaxBasisPoints)
		}
		previous = value.Height
	}
	return nil
}

func validateLatencyHistory(values []LatencySample, maxEntries uint32) error {
	if uint32(len(values)) > maxEntries {
		return errors.New("validator registry latency history limit exceeded")
	}
	previous := uint64(0)
	for _, value := range values {
		if value.Height == 0 || value.Height < previous {
			return errors.New("validator registry latency history must be sorted by positive height")
		}
		previous = value.Height
	}
	return nil
}

func validateSlashingHistory(values []SlashingEvent, maxEntries uint32) error {
	if uint32(len(values)) > maxEntries {
		return errors.New("validator registry slashing history limit exceeded")
	}
	previous := uint64(0)
	for _, value := range values {
		if value.Height == 0 || value.Height < previous {
			return errors.New("validator registry slashing history must be sorted by positive height")
		}
		if value.FractionBps > MaxBasisPoints {
			return fmt.Errorf("validator registry slash fraction must be <= %d bps", MaxBasisPoints)
		}
		if uint32(len(value.Reason)) > MaxSlashReasonBytesV1 {
			return fmt.Errorf("validator registry slash reason exceeds %d bytes", MaxSlashReasonBytesV1)
		}
		previous = value.Height
	}
	return nil
}

func validateValidatorHistory(values []ValidatorHistoryEvent, maxEntries uint32) error {
	if uint32(len(values)) > maxEntries {
		return errors.New("validator registry history limit exceeded")
	}
	previousHeight := uint64(0)
	previousType := ""
	for _, value := range values {
		if value.Height == 0 || value.Height < previousHeight {
			return errors.New("validator registry history must be sorted by positive height")
		}
		if value.Height == previousHeight && value.Type < previousType {
			return errors.New("validator registry history must be sorted deterministically")
		}
		if err := validateToken("validator history type", value.Type, MaxCapabilityBytesV1); err != nil {
			return err
		}
		if uint32(len(value.Detail)) > MaxMetadataBytesV1 {
			return fmt.Errorf("validator registry history detail exceeds %d bytes", MaxMetadataBytesV1)
		}
		previousHeight = value.Height
		previousType = value.Type
	}
	return nil
}

func validateCommissionParams(p Params) error {
	if p.CommissionFloorBps > p.DefaultCommissionBps ||
		p.DefaultCommissionBps > p.CommissionCeilingBps ||
		p.CommissionCeilingBps > MaxBasisPoints {
		return errors.New("validator registry commission floor/default/ceiling are invalid")
	}
	if p.CommissionMaxDailyChangeBps == 0 || p.CommissionMaxDailyChangeBps > p.CommissionCeilingBps {
		return errors.New("validator registry commission daily change is invalid")
	}
	return nil
}

func validatePowerCapSchedule(schedule []ValidatorPowerCapPhase) error {
	if len(schedule) == 0 {
		return errors.New("validator registry power cap schedule is required")
	}
	previousMax := uint32(0)
	for idx, phase := range schedule {
		if phase.PowerCapBps == 0 || phase.PowerCapBps > MaxBasisPoints {
			return errors.New("validator registry power cap phase is invalid")
		}
		if idx < len(schedule)-1 {
			if phase.MaxActiveValidators <= previousMax {
				return errors.New("validator registry power cap schedule must be sorted")
			}
			previousMax = phase.MaxActiveValidators
			continue
		}
		if phase.MaxActiveValidators != 0 {
			return errors.New("validator registry final power cap phase must be open-ended")
		}
	}
	return nil
}

func checkedAddUint64(a, b uint64) (uint64, error) {
	if ^uint64(0)-a < b {
		return 0, errors.New("validator registry uint64 overflow")
	}
	return a + b, nil
}

func sortedUptime(values []UptimeSample) []UptimeSample {
	out := append([]UptimeSample(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].Height < out[j].Height })
	return out
}

func sortedLatency(values []LatencySample) []LatencySample {
	out := append([]LatencySample(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].Height < out[j].Height })
	return out
}

func sortedSlashing(values []SlashingEvent) []SlashingEvent {
	out := append([]SlashingEvent(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Height == out[j].Height {
			return out[i].Reason < out[j].Reason
		}
		return out[i].Height < out[j].Height
	})
	return out
}

func sortedHistory(values []ValidatorHistoryEvent) []ValidatorHistoryEvent {
	out := append([]ValidatorHistoryEvent(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Height == out[j].Height {
			return out[i].Type < out[j].Type
		}
		return out[i].Height < out[j].Height
	})
	return out
}
