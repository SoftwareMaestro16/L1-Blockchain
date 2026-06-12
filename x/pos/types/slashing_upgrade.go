package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	SlashSeverityMinorLivenessFault		= "minor_liveness_fault"
	SlashSeverityMajorLivenessFault		= "major_liveness_fault"
	SlashSeverityRepeatedLivenessFault	= "repeated_liveness_fault"
	SlashSeverityInvalidTaskExecution	= "invalid_task_execution"
	SlashSeverityInvalidStateTransition	= "invalid_state_transition"
	SlashSeverityRepeatedInvalidProposal	= "repeated_invalid_proposal"
	SlashSeverityRepeatedTimestampViolation	= "repeated_timestamp_violation"
	SlashSeverityEquivocation		= "equivocation"
	SlashSeverityDoubleSign			= "double_sign"
	SlashSeverityEvidenceFraud		= "evidence_fraud"

	SlashSeverityLow	= "low"
	SlashSeverityMedium	= "medium"
	SlashSeverityHigh	= "high"
	SlashSeverityCritical	= "critical"

	DefaultSlashRepeatMultiplierBps	= uint32(BasisPoints)
	DefaultSlashImpactBps		= uint32(BasisPoints)

	DefaultPenaltyBurnBps		= uint32(4_000)
	DefaultPenaltyReporterBps	= uint32(1_000)
	DefaultPenaltyTreasuryBps	= uint32(4_000)
	DefaultPenaltyCompensationBps	= uint32(1_000)
)

type SlashingPenaltyInput struct {
	PenaltyID			string
	ValidatorID			string
	SeverityLevel			string
	SeverityBps			uint32
	StakeExposureNaet		sdkmath.Int
	RoleWeightBps			uint32
	RepeatOffenseMultiplierBps	uint32
	TaskImpactBps			uint32
	SafetyImpactBps			uint32
	LivenessImpactBps		uint32
	SelfStakeNaet			sdkmath.Int
	Nominations			[]Nomination
	RewardConfiscationNaet		sdkmath.Int
	TemporaryJailEpochs		uint64
	PermanentTombstone		bool
	IdentityInvalidation		bool
	RoleSuspensions			[]ValidatorRole
	FutureElectionScorePenaltyBps	uint32
	EvidenceHeight			int64
}

type SlashingPenalty struct {
	PenaltyID			string
	ValidatorID			string
	SeverityLevel			string
	SeverityBps			uint32
	StakeExposureNaet		sdkmath.Int
	RoleWeightBps			uint32
	RepeatOffenseMultiplierBps	uint32
	TaskImpactBps			uint32
	SafetyImpactBps			uint32
	LivenessImpactBps		uint32
	ScaledPenaltyBps		uint32
	StakeSlashNaet			sdkmath.Int
	ValidatorStakeSlashNaet		sdkmath.Int
	DelegatorSlashes		[]NominatorSlash
	DelegatorProportionalSlash	sdkmath.Int
	RewardConfiscationNaet		sdkmath.Int
	TemporaryJailEpochs		uint64
	PermanentTombstone		bool
	IdentityInvalidation		bool
	RoleSuspensions			[]ValidatorRole
	FutureElectionScorePenaltyBps	uint32
	EvidenceHeight			int64
	PenaltyHash			string
}

type SlashingPenaltyRoutingInput struct {
	Penalty			SlashingPenalty
	ReporterID		string
	AffectedPoolIDOptional	string
	BurnBps			uint32
	ReporterRewardBps	uint32
	ProtocolTreasuryBps	uint32
	CompensationBps		uint32
	ReporterRewardCapBps	uint32
}

type SlashingPenaltyRouting struct {
	PenaltyID		string
	ValidatorID		string
	ReporterID		string
	AffectedPoolIDOptional	string
	TotalPenaltyNaet	sdkmath.Int
	BurnNaet		sdkmath.Int
	ReporterRewardNaet	sdkmath.Int
	ProtocolTreasuryNaet	sdkmath.Int
	CompensationNaet	sdkmath.Int
	ResidualNaet		sdkmath.Int
	BurnBps			uint32
	ReporterRewardBps	uint32
	ProtocolTreasuryBps	uint32
	CompensationBps		uint32
	ReporterRewardCapBps	uint32
	RoutingHash		string
}

type SlashingRecord struct {
	PenaltyID		string
	ValidatorAddress	string
	EvidenceID		string
	Severity		string
	StakeExposure		sdkmath.Int
	RoleWeight		uint32
	SlashAmount		sdkmath.Int
	DelegatorSlashAmount	sdkmath.Int
	RewardConfiscation	sdkmath.Int
	JailUntilEpochOptional	uint64
	Tombstone		bool
	Routing			SlashingPenaltyRouting
	ExecutedHeight		int64
	RecordHash		string
}

type SlashingExecutionInput struct {
	EvidenceID		string
	ReporterID		string
	AffectedPoolIDOptional	string
	ExecutedHeight		int64
	CurrentEpoch		uint64
	Candidate		Candidate
	PenaltyInput		SlashingPenaltyInput
	RoutingInput		SlashingPenaltyRoutingInput
}

type SlashingExecutionResult struct {
	Record			SlashingRecord
	UpdatedCandidate	Candidate
	SeverityMatrix		map[string]uint32
	NonNegative		bool
	ExactRouting		bool
	DelegatorExact		bool
}

type SlashingEvidenceTimingInput struct {
	FaultHeight			uint64
	EvidenceHeight			uint64
	CurrentHeight			uint64
	MaxEvidenceAgeBlocks		uint64
	ValidatorBondedHeight		uint64
	ValidatorUnbondingHeight	uint64
	UnbondingEvidenceWindowBlocks	uint64
}

func SlashSeverityClasses() []string {
	return []string{
		SlashSeverityMinorLivenessFault,
		SlashSeverityMajorLivenessFault,
		SlashSeverityRepeatedLivenessFault,
		SlashSeverityInvalidTaskExecution,
		SlashSeverityInvalidStateTransition,
		SlashSeverityRepeatedInvalidProposal,
		SlashSeverityRepeatedTimestampViolation,
		SlashSeverityEquivocation,
		SlashSeverityDoubleSign,
		SlashSeverityEvidenceFraud,
	}
}

func ValidateSlashingEvidenceTiming(input SlashingEvidenceTimingInput) error {
	if input.FaultHeight == 0 || input.EvidenceHeight == 0 || input.CurrentHeight == 0 {
		return errors.New("slashing evidence timing heights must be positive")
	}
	if input.MaxEvidenceAgeBlocks == 0 {
		return errors.New("slashing evidence max age must be positive")
	}
	if input.ValidatorBondedHeight == 0 {
		return errors.New("slashing evidence validator bonded height is required")
	}
	if input.FaultHeight < input.ValidatorBondedHeight {
		return errors.New("slashing fault height precedes validator bonded height")
	}
	if input.EvidenceHeight < input.FaultHeight {
		return errors.New("slashing evidence height cannot precede fault height")
	}
	if input.CurrentHeight < input.EvidenceHeight {
		return errors.New("slashing current height cannot precede evidence height")
	}
	if input.EvidenceHeight-input.FaultHeight > input.MaxEvidenceAgeBlocks {
		return errors.New("slashing evidence expired by max age")
	}
	if input.ValidatorUnbondingHeight == 0 {
		return nil
	}
	if input.UnbondingEvidenceWindowBlocks == 0 {
		return errors.New("slashing unbonding evidence window is required")
	}
	if input.FaultHeight >= input.ValidatorUnbondingHeight {
		return errors.New("slashing fault height is after validator unbonding height")
	}
	if input.EvidenceHeight > input.ValidatorUnbondingHeight+input.UnbondingEvidenceWindowBlocks {
		return errors.New("slashing evidence arrived after unbonding risk window")
	}
	return nil
}

func DefaultSeverityBps(severityLevel string) (uint32, error) {
	switch severityLevel {
	case SlashSeverityMinorLivenessFault:
		return 5, nil
	case SlashSeverityMajorLivenessFault:
		return 25, nil
	case SlashSeverityRepeatedLivenessFault:
		return 100, nil
	case SlashSeverityInvalidTaskExecution:
		return 750, nil
	case SlashSeverityInvalidStateTransition:
		return 1_500, nil
	case SlashSeverityRepeatedInvalidProposal:
		return 25, nil
	case SlashSeverityRepeatedTimestampViolation:
		return 25, nil
	case SlashSeverityEquivocation:
		return 2_000, nil
	case SlashSeverityDoubleSign:
		return 500, nil
	case SlashSeverityEvidenceFraud:
		return 7_500, nil
	case "low":
		return 250, nil
	case "medium":
		return 1_000, nil
	case "high":
		return 3_000, nil
	case "critical":
		return 7_500, nil
	default:
		return 0, fmt.Errorf("unsupported slash severity level %q", severityLevel)
	}
}

func ComputeSlashingPenalty(input SlashingPenaltyInput) (SlashingPenalty, error) {
	input = normalizeSlashingPenaltyInput(input)
	if err := input.Validate(); err != nil {
		return SlashingPenalty{}, err
	}
	scaledBps := scaledPenaltyBps(input)
	stakeSlash := mulIntBps(input.StakeExposureNaet, scaledBps)
	if stakeSlash.GT(input.StakeExposureNaet) {
		stakeSlash = input.StakeExposureNaet
	}
	totalStake := input.SelfStakeNaet.Add(sumNominations(input.Nominations))
	if stakeSlash.GT(totalStake) {
		stakeSlash = totalStake
	}
	validatorSlash := shareByStake(stakeSlash, input.SelfStakeNaet, totalStake)
	delegatorTotal := sdkmath.ZeroInt()
	delegatorSlashes := make([]NominatorSlash, 0, len(input.Nominations))
	for _, nomination := range sortNominations(input.Nominations) {
		slashed := shareByStake(stakeSlash, nomination.StakeNaet, totalStake)
		delegatorSlashes = append(delegatorSlashes, NominatorSlash{NominatorID: nomination.NominatorID, SlashedNaet: slashed})
		delegatorTotal = delegatorTotal.Add(slashed)
	}
	remainder := stakeSlash.Sub(validatorSlash).Sub(delegatorTotal)
	if remainder.IsPositive() {
		validatorSlash = validatorSlash.Add(remainder)
	}
	roles := cloneValidatorRoles(input.RoleSuspensions)
	sort.SliceStable(roles, func(i, j int) bool { return roles[i] < roles[j] })
	penalty := SlashingPenalty{
		PenaltyID:			input.PenaltyID,
		ValidatorID:			input.ValidatorID,
		SeverityLevel:			input.SeverityLevel,
		SeverityBps:			input.SeverityBps,
		StakeExposureNaet:		input.StakeExposureNaet,
		RoleWeightBps:			input.RoleWeightBps,
		RepeatOffenseMultiplierBps:	input.RepeatOffenseMultiplierBps,
		TaskImpactBps:			input.TaskImpactBps,
		SafetyImpactBps:		input.SafetyImpactBps,
		LivenessImpactBps:		input.LivenessImpactBps,
		ScaledPenaltyBps:		scaledBps,
		StakeSlashNaet:			stakeSlash,
		ValidatorStakeSlashNaet:	validatorSlash,
		DelegatorSlashes:		delegatorSlashes,
		DelegatorProportionalSlash:	delegatorTotal,
		RewardConfiscationNaet:		input.RewardConfiscationNaet,
		TemporaryJailEpochs:		input.TemporaryJailEpochs,
		PermanentTombstone:		input.PermanentTombstone,
		IdentityInvalidation:		input.IdentityInvalidation,
		RoleSuspensions:		roles,
		FutureElectionScorePenaltyBps:	input.FutureElectionScorePenaltyBps,
		EvidenceHeight:			input.EvidenceHeight,
	}
	penalty.PenaltyHash = computeSlashingPenaltyHash(penalty)
	return penalty, penalty.Validate()
}

func (i SlashingPenaltyInput) Validate() error {
	if err := validatePosToken("slashing penalty id", i.PenaltyID); err != nil {
		return err
	}
	if err := validatePosToken("slashing validator id", i.ValidatorID); err != nil {
		return err
	}
	if _, err := DefaultSeverityBps(i.SeverityLevel); err != nil {
		return err
	}
	if i.SeverityBps == 0 || i.SeverityBps > BasisPoints {
		return fmt.Errorf("slash severity must be within 1..%d bps", BasisPoints)
	}
	if i.StakeExposureNaet.IsNil() || !i.StakeExposureNaet.IsPositive() {
		return errors.New("slash stake exposure must be positive")
	}
	if err := validateBps("slash role weight", i.RoleWeightBps); err != nil {
		return err
	}
	if err := validateBps("slash repeat offense multiplier", i.RepeatOffenseMultiplierBps); err != nil {
		return err
	}
	if err := validateBps("slash task impact", i.TaskImpactBps); err != nil {
		return err
	}
	if err := validateBps("slash safety impact", i.SafetyImpactBps); err != nil {
		return err
	}
	if err := validateBps("slash liveness impact", i.LivenessImpactBps); err != nil {
		return err
	}
	if i.SelfStakeNaet.IsNil() || i.SelfStakeNaet.IsNegative() {
		return errors.New("slash self stake cannot be negative")
	}
	if err := validateNominations(i.Nominations); err != nil {
		return err
	}
	if !i.SelfStakeNaet.Add(sumNominations(i.Nominations)).IsPositive() {
		return errors.New("slash total stake must be positive")
	}
	if i.RewardConfiscationNaet.IsNil() || i.RewardConfiscationNaet.IsNegative() {
		return errors.New("slash reward confiscation cannot be negative")
	}
	if i.FutureElectionScorePenaltyBps > BasisPoints {
		return fmt.Errorf("future election score penalty must be <= %d bps", BasisPoints)
	}
	if i.EvidenceHeight < 0 {
		return errors.New("slash evidence height cannot be negative")
	}
	return validateValidatorRoles(i.RoleSuspensions)
}

func (p SlashingPenalty) Validate() error {
	if err := validatePosToken("slashing penalty id", p.PenaltyID); err != nil {
		return err
	}
	if err := validatePosToken("slashing validator id", p.ValidatorID); err != nil {
		return err
	}
	if p.StakeSlashNaet.IsNil() || p.StakeSlashNaet.IsNegative() {
		return errors.New("stake slash cannot be negative")
	}
	if p.ValidatorStakeSlashNaet.IsNil() || p.ValidatorStakeSlashNaet.IsNegative() {
		return errors.New("validator stake slash cannot be negative")
	}
	if p.DelegatorProportionalSlash.IsNil() || p.DelegatorProportionalSlash.IsNegative() {
		return errors.New("delegator proportional slash cannot be negative")
	}
	if p.ValidatorStakeSlashNaet.Add(p.DelegatorProportionalSlash).Equal(p.StakeSlashNaet) == false {
		return errors.New("stake slash components must sum to total stake slash")
	}
	if p.RewardConfiscationNaet.IsNil() || p.RewardConfiscationNaet.IsNegative() {
		return errors.New("reward confiscation cannot be negative")
	}
	if p.FutureElectionScorePenaltyBps > BasisPoints {
		return fmt.Errorf("future election score penalty must be <= %d bps", BasisPoints)
	}
	if err := validateValidatorRoles(p.RoleSuspensions); err != nil {
		return err
	}
	if p.PenaltyHash != computeSlashingPenaltyHash(p) {
		return errors.New("slashing penalty hash mismatch")
	}
	return nil
}

func ApplySlashingPenaltyToCandidate(candidate Candidate, penalty SlashingPenalty) (Candidate, error) {
	if strings.TrimSpace(candidate.ValidatorID) != penalty.ValidatorID {
		return Candidate{}, errors.New("slashing penalty validator mismatch")
	}
	if err := penalty.Validate(); err != nil {
		return Candidate{}, err
	}
	next := cloneCandidate(candidate)
	next.SelfStakeNaet = maxInt(next.SelfStakeNaet.Sub(penalty.ValidatorStakeSlashNaet), sdkmath.ZeroInt())
	next.DelegatedStakeNaet = maxInt(next.DelegatedStakeNaet.Sub(penalty.DelegatorProportionalSlash), sdkmath.ZeroInt())
	if penalty.TemporaryJailEpochs > 0 {
		next.Jailed = true
	}
	if penalty.PermanentTombstone {
		next.Tombstoned = true
	}
	if penalty.FutureElectionScorePenaltyBps > 0 {
		next.ReliabilityIndexBps = reduceBps(normalizeOptionalFactorBps(next.ReliabilityIndexBps), penalty.FutureElectionScorePenaltyBps)
	}
	if len(penalty.RoleSuspensions) > 0 {
		next.Roles = removeSuspendedRoles(next.Roles, penalty.RoleSuspensions)
	}
	return next, nil
}

func RouteSlashingPenalty(input SlashingPenaltyRoutingInput) (SlashingPenaltyRouting, error) {
	input = normalizeSlashingPenaltyRoutingInput(input)
	if err := input.Validate(); err != nil {
		return SlashingPenaltyRouting{}, err
	}
	total := input.Penalty.StakeSlashNaet.Add(input.Penalty.RewardConfiscationNaet)
	reporterReward := mulIntBps(total, input.ReporterRewardBps)
	capAmount := mulIntBps(total, input.ReporterRewardCapBps)
	if reporterReward.GT(capAmount) {
		reporterReward = capAmount
	}
	burn := mulIntBps(total, input.BurnBps)
	treasury := mulIntBps(total, input.ProtocolTreasuryBps)
	compensation := mulIntBps(total, input.CompensationBps)
	residual := total.Sub(burn).Sub(reporterReward).Sub(treasury).Sub(compensation)
	if residual.IsNegative() {
		return SlashingPenaltyRouting{}, errors.New("slashing penalty routing exceeds total penalty")
	}
	routing := SlashingPenaltyRouting{
		PenaltyID:		input.Penalty.PenaltyID,
		ValidatorID:		input.Penalty.ValidatorID,
		ReporterID:		input.ReporterID,
		AffectedPoolIDOptional:	input.AffectedPoolIDOptional,
		TotalPenaltyNaet:	total,
		BurnNaet:		burn,
		ReporterRewardNaet:	reporterReward,
		ProtocolTreasuryNaet:	treasury,
		CompensationNaet:	compensation,
		ResidualNaet:		residual,
		BurnBps:		input.BurnBps,
		ReporterRewardBps:	input.ReporterRewardBps,
		ProtocolTreasuryBps:	input.ProtocolTreasuryBps,
		CompensationBps:	input.CompensationBps,
		ReporterRewardCapBps:	input.ReporterRewardCapBps,
	}
	routing.RoutingHash = computeSlashingPenaltyRoutingHash(routing)
	return routing, routing.Validate()
}

func ExecuteSlashing(input SlashingExecutionInput) (SlashingExecutionResult, error) {
	input.EvidenceID = strings.TrimSpace(input.EvidenceID)
	input.ReporterID = strings.TrimSpace(input.ReporterID)
	input.AffectedPoolIDOptional = strings.TrimSpace(input.AffectedPoolIDOptional)
	if err := validatePosToken("slashing execution evidence id", input.EvidenceID); err != nil {
		return SlashingExecutionResult{}, err
	}
	if input.ExecutedHeight < 0 {
		return SlashingExecutionResult{}, errors.New("slashing executed height cannot be negative")
	}
	if input.CurrentEpoch == 0 {
		return SlashingExecutionResult{}, errors.New("slashing current epoch is required")
	}
	penaltyInput := input.PenaltyInput
	penaltyInput.ValidatorID = input.Candidate.ValidatorID
	penalty, err := ComputeSlashingPenalty(penaltyInput)
	if err != nil {
		return SlashingExecutionResult{}, err
	}
	routingInput := input.RoutingInput
	routingInput.Penalty = penalty
	if routingInput.ReporterID == "" {
		routingInput.ReporterID = input.ReporterID
	}
	if routingInput.AffectedPoolIDOptional == "" {
		routingInput.AffectedPoolIDOptional = input.AffectedPoolIDOptional
	}
	routing, err := RouteSlashingPenalty(routingInput)
	if err != nil {
		return SlashingExecutionResult{}, err
	}
	updated, err := ApplySlashingPenaltyToCandidate(input.Candidate, penalty)
	if err != nil {
		return SlashingExecutionResult{}, err
	}
	record := NewSlashingRecord(input.EvidenceID, penalty, routing, input.CurrentEpoch, input.ExecutedHeight)
	result := SlashingExecutionResult{
		Record:			record,
		UpdatedCandidate:	updated,
		SeverityMatrix:		SeverityMatrix(),
		NonNegative:		!updated.SelfStakeNaet.IsNegative() && !updated.DelegatedStakeNaet.IsNegative(),
		ExactRouting:		routing.BurnNaet.Add(routing.ReporterRewardNaet).Add(routing.ProtocolTreasuryNaet).Add(routing.CompensationNaet).Add(routing.ResidualNaet).Equal(routing.TotalPenaltyNaet),
		DelegatorExact:		record.DelegatorSlashAmount.Equal(penalty.DelegatorProportionalSlash),
	}
	if err := result.Validate(); err != nil {
		return SlashingExecutionResult{}, err
	}
	return result, nil
}

func NewSlashingRecord(evidenceID string, penalty SlashingPenalty, routing SlashingPenaltyRouting, currentEpoch uint64, executedHeight int64) SlashingRecord {
	jailUntil := uint64(0)
	if penalty.TemporaryJailEpochs > 0 {
		jailUntil = currentEpoch + penalty.TemporaryJailEpochs
	}
	record := SlashingRecord{
		PenaltyID:		penalty.PenaltyID,
		ValidatorAddress:	penalty.ValidatorID,
		EvidenceID:		strings.TrimSpace(evidenceID),
		Severity:		penalty.SeverityLevel,
		StakeExposure:		penalty.StakeExposureNaet,
		RoleWeight:		penalty.RoleWeightBps,
		SlashAmount:		penalty.StakeSlashNaet,
		DelegatorSlashAmount:	penalty.DelegatorProportionalSlash,
		RewardConfiscation:	penalty.RewardConfiscationNaet,
		JailUntilEpochOptional:	jailUntil,
		Tombstone:		penalty.PermanentTombstone,
		Routing:		routing,
		ExecutedHeight:		executedHeight,
	}
	record.RecordHash = computeSlashingRecordHash(record)
	return record
}

func SlashingRecordFieldNames() []string {
	return []string{
		"penalty_id",
		"validator_address",
		"evidence_id",
		"severity",
		"stake_exposure",
		"role_weight",
		"slash_amount",
		"delegator_slash_amount",
		"reward_confiscation",
		"jail_until_epoch_optional",
		"tombstone",
		"routing",
		"executed_height",
	}
}

func SeverityMatrix() map[string]uint32 {
	matrix := make(map[string]uint32, len(SlashSeverityClasses()))
	for _, severity := range SlashSeverityClasses() {
		bps, _ := DefaultSeverityBps(severity)
		matrix[severity] = bps
	}
	return matrix
}

func (r SlashingRecord) Validate() error {
	if err := validatePosToken("slashing record penalty id", r.PenaltyID); err != nil {
		return err
	}
	if err := validatePosToken("slashing record validator address", r.ValidatorAddress); err != nil {
		return err
	}
	if err := validatePosToken("slashing record evidence id", r.EvidenceID); err != nil {
		return err
	}
	if _, err := DefaultSeverityBps(r.Severity); err != nil {
		return err
	}
	if r.StakeExposure.IsNil() || r.StakeExposure.IsNegative() {
		return errors.New("slashing record stake exposure cannot be negative")
	}
	if r.RoleWeight == 0 || r.RoleWeight > BasisPoints {
		return fmt.Errorf("slashing record role weight must be within 1..%d bps", BasisPoints)
	}
	if r.SlashAmount.IsNil() || r.SlashAmount.IsNegative() {
		return errors.New("slashing record slash amount cannot be negative")
	}
	if r.DelegatorSlashAmount.IsNil() || r.DelegatorSlashAmount.IsNegative() {
		return errors.New("slashing record delegator slash amount cannot be negative")
	}
	if r.RewardConfiscation.IsNil() || r.RewardConfiscation.IsNegative() {
		return errors.New("slashing record reward confiscation cannot be negative")
	}
	if r.ExecutedHeight < 0 {
		return errors.New("slashing record executed height cannot be negative")
	}
	if err := r.Routing.Validate(); err != nil {
		return err
	}
	if !r.Routing.TotalPenaltyNaet.Equal(r.SlashAmount.Add(r.RewardConfiscation)) {
		return errors.New("slashing record routing total must equal slash plus confiscated reward")
	}
	if r.RecordHash != computeSlashingRecordHash(r) {
		return errors.New("slashing record hash mismatch")
	}
	return nil
}

func (r SlashingExecutionResult) Validate() error {
	if err := r.Record.Validate(); err != nil {
		return err
	}
	if !r.NonNegative {
		return errors.New("slashing execution produced negative balances")
	}
	if !r.ExactRouting {
		return errors.New("slashing execution routing is not exact")
	}
	if !r.DelegatorExact {
		return errors.New("slashing execution delegator propagation is not exact")
	}
	return nil
}

func (i SlashingPenaltyRoutingInput) Validate() error {
	if err := i.Penalty.Validate(); err != nil {
		return err
	}
	if i.ReporterID != "" {
		if err := validatePosToken("slashing penalty reporter id", i.ReporterID); err != nil {
			return err
		}
	}
	if i.AffectedPoolIDOptional != "" {
		if err := validatePosToken("slashing affected pool id", i.AffectedPoolIDOptional); err != nil {
			return err
		}
	}
	if i.ReporterRewardBps > 0 && i.ReporterID == "" {
		return errors.New("slashing reporter reward requires reporter id")
	}
	if i.CompensationBps > 0 && i.AffectedPoolIDOptional == "" {
		return errors.New("slashing compensation requires affected pool id")
	}
	for _, item := range []struct {
		name	string
		value	uint32
	}{
		{name: "burn bps", value: i.BurnBps},
		{name: "reporter reward bps", value: i.ReporterRewardBps},
		{name: "protocol treasury bps", value: i.ProtocolTreasuryBps},
		{name: "compensation bps", value: i.CompensationBps},
		{name: "reporter reward cap bps", value: i.ReporterRewardCapBps},
	} {
		if item.value > BasisPoints {
			return fmt.Errorf("slashing %s must be <= %d", item.name, BasisPoints)
		}
	}
	totalBps := uint64(i.BurnBps) + uint64(i.ReporterRewardBps) + uint64(i.ProtocolTreasuryBps) + uint64(i.CompensationBps)
	if totalBps > uint64(BasisPoints) {
		return fmt.Errorf("slashing penalty routing bps must be <= %d", BasisPoints)
	}
	return nil
}

func (r SlashingPenaltyRouting) Validate() error {
	if err := validatePosToken("slashing routing penalty id", r.PenaltyID); err != nil {
		return err
	}
	if err := validatePosToken("slashing routing validator id", r.ValidatorID); err != nil {
		return err
	}
	if r.TotalPenaltyNaet.IsNil() || r.TotalPenaltyNaet.IsNegative() {
		return errors.New("slashing routing total penalty cannot be negative")
	}
	if r.BurnNaet.IsNil() || r.ReporterRewardNaet.IsNil() || r.ProtocolTreasuryNaet.IsNil() || r.CompensationNaet.IsNil() || r.ResidualNaet.IsNil() {
		return errors.New("slashing routing amounts must not be nil")
	}
	sum := r.BurnNaet.Add(r.ReporterRewardNaet).Add(r.ProtocolTreasuryNaet).Add(r.CompensationNaet).Add(r.ResidualNaet)
	if !sum.Equal(r.TotalPenaltyNaet) {
		return errors.New("slashing routing amounts must sum to total penalty")
	}
	if r.RoutingHash != computeSlashingPenaltyRoutingHash(r) {
		return errors.New("slashing routing hash mismatch")
	}
	return nil
}

func normalizeSlashingPenaltyInput(input SlashingPenaltyInput) SlashingPenaltyInput {
	input.PenaltyID = strings.TrimSpace(input.PenaltyID)
	input.ValidatorID = strings.TrimSpace(input.ValidatorID)
	input.SeverityLevel = strings.TrimSpace(input.SeverityLevel)
	if input.SeverityBps == 0 {
		severity, err := DefaultSeverityBps(input.SeverityLevel)
		if err == nil {
			input.SeverityBps = severity
		}
	}
	if input.RoleWeightBps == 0 {
		input.RoleWeightBps = BasisPoints
	}
	if input.RepeatOffenseMultiplierBps == 0 {
		input.RepeatOffenseMultiplierBps = DefaultSlashRepeatMultiplierBps
	}
	if input.TaskImpactBps == 0 {
		input.TaskImpactBps = DefaultSlashImpactBps
	}
	if input.SafetyImpactBps == 0 {
		input.SafetyImpactBps = DefaultSlashImpactBps
	}
	if input.LivenessImpactBps == 0 {
		input.LivenessImpactBps = DefaultSlashImpactBps
	}
	if input.RewardConfiscationNaet.IsNil() {
		input.RewardConfiscationNaet = sdkmath.ZeroInt()
	}
	return input
}

func normalizeSlashingPenaltyRoutingInput(input SlashingPenaltyRoutingInput) SlashingPenaltyRoutingInput {
	input.ReporterID = strings.TrimSpace(input.ReporterID)
	input.AffectedPoolIDOptional = strings.TrimSpace(input.AffectedPoolIDOptional)
	if input.BurnBps == 0 && input.ReporterRewardBps == 0 && input.ProtocolTreasuryBps == 0 && input.CompensationBps == 0 {
		input.BurnBps = DefaultPenaltyBurnBps
		input.ReporterRewardBps = DefaultPenaltyReporterBps
		input.ProtocolTreasuryBps = DefaultPenaltyTreasuryBps
		input.CompensationBps = DefaultPenaltyCompensationBps
	}
	if input.ReporterRewardCapBps == 0 {
		input.ReporterRewardCapBps = input.ReporterRewardBps
	}
	return input
}

func scaledPenaltyBps(input SlashingPenaltyInput) uint32 {
	value := uint64(input.SeverityBps)
	for _, factor := range []uint32{
		input.RoleWeightBps,
		input.RepeatOffenseMultiplierBps,
		input.TaskImpactBps,
		input.SafetyImpactBps,
		input.LivenessImpactBps,
	} {
		value = value * uint64(factor) / uint64(BasisPoints)
		if value >= uint64(BasisPoints) {
			return BasisPoints
		}
	}
	return uint32(value)
}

func validateBps(fieldName string, value uint32) error {
	if value == 0 || value > BasisPoints {
		return fmt.Errorf("%s must be within 1..%d bps", fieldName, BasisPoints)
	}
	return nil
}

func computeSlashingPenaltyHash(penalty SlashingPenalty) string {
	return posHashRoot("aetra-pos-slashing-penalty-v1", func(w posByteWriter) {
		posWritePart(w, penalty.PenaltyID)
		posWritePart(w, penalty.ValidatorID)
		posWritePart(w, penalty.SeverityLevel)
		posWriteUint64(w, uint64(penalty.SeverityBps))
		posWritePart(w, penalty.StakeExposureNaet.String())
		posWriteUint64(w, uint64(penalty.RoleWeightBps))
		posWriteUint64(w, uint64(penalty.RepeatOffenseMultiplierBps))
		posWriteUint64(w, uint64(penalty.TaskImpactBps))
		posWriteUint64(w, uint64(penalty.SafetyImpactBps))
		posWriteUint64(w, uint64(penalty.LivenessImpactBps))
		posWriteUint64(w, uint64(penalty.ScaledPenaltyBps))
		posWritePart(w, penalty.StakeSlashNaet.String())
		posWritePart(w, penalty.ValidatorStakeSlashNaet.String())
		posWritePart(w, penalty.DelegatorProportionalSlash.String())
		posWritePart(w, penalty.RewardConfiscationNaet.String())
		posWriteUint64(w, penalty.TemporaryJailEpochs)
		posWritePart(w, fmt.Sprintf("%t", penalty.PermanentTombstone))
		posWritePart(w, fmt.Sprintf("%t", penalty.IdentityInvalidation))
		posWriteUint64(w, uint64(penalty.FutureElectionScorePenaltyBps))
		posWriteUint64(w, uint64(penalty.EvidenceHeight))
		posWriteUint64(w, uint64(len(penalty.RoleSuspensions)))
		for _, role := range penalty.RoleSuspensions {
			posWritePart(w, string(role))
		}
		posWriteUint64(w, uint64(len(penalty.DelegatorSlashes)))
		for _, slash := range penalty.DelegatorSlashes {
			posWritePart(w, slash.NominatorID)
			posWritePart(w, slash.SlashedNaet.String())
		}
	})
}

func computeSlashingPenaltyRoutingHash(routing SlashingPenaltyRouting) string {
	return posHashRoot("aetra-pos-slashing-penalty-routing-v1", func(w posByteWriter) {
		posWritePart(w, routing.PenaltyID)
		posWritePart(w, routing.ValidatorID)
		posWritePart(w, routing.ReporterID)
		posWritePart(w, routing.AffectedPoolIDOptional)
		posWritePart(w, routing.TotalPenaltyNaet.String())
		posWritePart(w, routing.BurnNaet.String())
		posWritePart(w, routing.ReporterRewardNaet.String())
		posWritePart(w, routing.ProtocolTreasuryNaet.String())
		posWritePart(w, routing.CompensationNaet.String())
		posWritePart(w, routing.ResidualNaet.String())
		posWriteUint64(w, uint64(routing.BurnBps))
		posWriteUint64(w, uint64(routing.ReporterRewardBps))
		posWriteUint64(w, uint64(routing.ProtocolTreasuryBps))
		posWriteUint64(w, uint64(routing.CompensationBps))
		posWriteUint64(w, uint64(routing.ReporterRewardCapBps))
	})
}

func computeSlashingRecordHash(record SlashingRecord) string {
	return posHashRoot("aetra-pos-slashing-record-v1", func(w posByteWriter) {
		posWritePart(w, record.PenaltyID)
		posWritePart(w, record.ValidatorAddress)
		posWritePart(w, record.EvidenceID)
		posWritePart(w, record.Severity)
		posWritePart(w, record.StakeExposure.String())
		posWriteUint64(w, uint64(record.RoleWeight))
		posWritePart(w, record.SlashAmount.String())
		posWritePart(w, record.DelegatorSlashAmount.String())
		posWritePart(w, record.RewardConfiscation.String())
		posWriteUint64(w, record.JailUntilEpochOptional)
		posWritePart(w, fmt.Sprintf("%t", record.Tombstone))
		posWritePart(w, record.Routing.RoutingHash)
		posWritePart(w, fmt.Sprintf("%d", record.ExecutedHeight))
	})
}

func cloneValidatorRoles(values []ValidatorRole) []ValidatorRole {
	out := make([]ValidatorRole, len(values))
	copy(out, values)
	return out
}

func removeSuspendedRoles(existing []ValidatorRole, suspended []ValidatorRole) []ValidatorRole {
	suspendedSet := make(map[ValidatorRole]struct{}, len(suspended))
	for _, role := range suspended {
		suspendedSet[role] = struct{}{}
	}
	out := make([]ValidatorRole, 0, len(existing))
	for _, role := range existing {
		if _, found := suspendedSet[role]; !found {
			out = append(out, role)
		}
	}
	return out
}

func reduceBps(value uint32, penaltyBps uint32) uint32 {
	if penaltyBps >= BasisPoints {
		return 0
	}
	return uint32(uint64(value) * uint64(BasisPoints-penaltyBps) / uint64(BasisPoints))
}

func maxInt(left sdkmath.Int, right sdkmath.Int) sdkmath.Int {
	if left.GTE(right) {
		return left
	}
	return right
}
