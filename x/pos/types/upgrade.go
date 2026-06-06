package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	DefaultDelegationActivationEpochs = uint64(1)
	DefaultEvidenceWindowEpochs       = uint64(4)
	DefaultMinTaskGroupValidators     = uint32(3)
	DefaultMaxTaskGroupValidators     = uint32(21)
	DefaultReporterRewardBps          = uint32(500)
	MaxReporterRewardBps              = uint32(2_000)

	PosHashHexLength = 64
	PosEmptyRootHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	DefaultWorkloadClass = "general"
	maxPosTokenLength    = 128
)

type EpochPhase string

const (
	EpochPhaseDelegation EpochPhase = "delegation"
	EpochPhaseElection   EpochPhase = "election"
	EpochPhaseAssignment EpochPhase = "assignment"
	EpochPhaseActive     EpochPhase = "active"
	EpochPhaseSettlement EpochPhase = "settlement"
	EpochPhaseClosed     EpochPhase = "closed"
)

type SettlementStatus string

const (
	SettlementStatusPending   SettlementStatus = "pending"
	SettlementStatusFinalized SettlementStatus = "finalized"
)

type ValidatorRole string

const (
	ValidatorRoleBlockProducer    ValidatorRole = "block_producer"
	ValidatorRoleVerifier         ValidatorRole = "verifier"
	ValidatorRoleCollator         ValidatorRole = "collator"
	ValidatorRoleEvidenceReviewer ValidatorRole = "evidence_reviewer"
)

type EpochPhaseDurations struct {
	DelegationSeconds       uint64
	ElectionSeconds         uint64
	AssignmentSeconds       uint64
	ActiveValidationSeconds uint64
	SettlementSeconds       uint64
}

type EpochSeedSource string

const (
	EpochSeedSourcePreviousSeedValidatorSet EpochSeedSource = "previous_seed_validator_set"
	EpochSeedSourceCometBFTBlockID          EpochSeedSource = "cometbft_block_id"
	EpochSeedSourceExternalBeacon           EpochSeedSource = "external_beacon"
)

type EpochLifecycleStep struct {
	Phase       EpochPhase
	Name        string
	DurationKey string
}

type EpochRecord struct {
	EpochID          uint64
	StartHeight      uint64
	EndHeight        uint64
	Phase            EpochPhase
	Seed             string
	ValidatorSetHash string
	TaskGroupRoot    string
	PerformanceRoot  string
	RewardRoot       string
	SlashRoot        string
	SettlementStatus SettlementStatus
}

type EpochSettlementRoots struct {
	PerformanceRoot string
	RewardRoot      string
	SlashRoot       string
}

type WorkloadTask struct {
	TaskID             string
	WorkloadID         string
	WorkloadType       WorkloadType
	ZoneID             string
	ShardID            string
	WorkloadClass      string
	RequiredValidators uint32
	Roles              []ValidatorRole
}

type WorkloadType string

const (
	WorkloadTypeGlobalConsensus      WorkloadType = "global_consensus"
	WorkloadTypeZoneExecution        WorkloadType = "zone_execution"
	WorkloadTypeShardExecution       WorkloadType = "shard_execution"
	WorkloadTypeProofVerification    WorkloadType = "proof_verification"
	WorkloadTypeEvidenceVerification WorkloadType = "evidence_verification"
	WorkloadTypeDataAvailability     WorkloadType = "data_availability"
	WorkloadTypeServiceValidation    WorkloadType = "service_validation"
)

type TaskAssignment struct {
	TaskID         string
	WorkloadID     string
	WorkloadType   WorkloadType
	ZoneID         string
	ShardID        string
	WorkloadClass  string
	Role           ValidatorRole
	Validators     []string
	AssignmentHash string
}

type TaskAssignmentSet struct {
	EpochID     uint64
	Seed        string
	Assignments []TaskAssignment
	Root        string
}

type TaskGroup struct {
	EpochID          uint64
	TaskGroupID      string
	WorkloadID       string
	WorkloadType     WorkloadType
	ValidatorMembers []string
	ProposerOrder    []string
	VerifierSet      []string
	MinimumGroupSize uint32
	StakeWeightRoot  string
	AssignmentSeed   string
	ActivationHeight uint64
	ExpiryHeight     uint64
}

type TaskGroupSet struct {
	EpochID uint64
	Seed    string
	Groups  []TaskGroup
	Root    string
}

type DelegationIntent struct {
	NominatorID            string
	ValidatorID            string
	StakeNaet              sdkmath.Int
	RequestedEpoch         uint64
	MaxCommissionBps       uint32
	MinPerformanceScoreBps uint32
}

type DelegationActivation struct {
	ValidatorID   string
	Nominations   []Nomination
	ActivatedAt   uint64
	IntentCount   uint32
	TotalStake    sdkmath.Int
	ActivationKey string
}

type RejectedDelegationIntent struct {
	Intent DelegationIntent
	Reason string
}

type EvidenceCase struct {
	EvidenceID       string
	ReporterID       string
	ValidatorID      string
	Misbehavior      string
	SlashFractionBps uint32
	EvidenceHeight   int64
	EvidenceEpoch    uint64
	Finalized        bool
}

type EvidenceSettlement struct {
	EvidenceID         string
	ReporterID         string
	Slash              SlashDistribution
	ReporterRewardNaet sdkmath.Int
	BurnNaet           sdkmath.Int
	SettlementHash     string
}

type RoleRewardWeight struct {
	Role      ValidatorRole
	WeightBps uint32
}

type AssignmentOutcome struct {
	TaskID      string
	Role        ValidatorRole
	ValidatorID string
	Completed   bool
	Faulted     bool
	WorkUnits   uint64
}

type ValidatorWorkloadReward struct {
	ValidatorID string
	RewardNaet  sdkmath.Int
	WorkUnits   uint64
}

type WorkloadRewardInput struct {
	EpochID          uint64
	TotalRewardsNaet sdkmath.Int
	RoleWeights      []RoleRewardWeight
	Outcomes         []AssignmentOutcome
}

type WorkloadRewardSettlement struct {
	EpochID        uint64
	Rewards        []ValidatorWorkloadReward
	RemainderNaet  sdkmath.Int
	RewardRoot     string
	CompletedUnits uint64
}

type PerformanceFactorInput struct {
	CompletedTasks         uint64
	MissedTasks            uint64
	CorrectVerifications   uint64
	IncorrectVerifications uint64
	AvailableWindows       uint64
	CommittedWindows       uint64
}

type UptimeFactorInput struct {
	SignedBlocks             uint64
	TotalBlocks              uint64
	TaskParticipations       uint64
	MissedTaskParticipations uint64
}

type LatencyFactorInput struct {
	CommittedWindow bool
	AdvisoryOnly    bool
	TargetMillis    uint64
	P95Millis       uint64
}

type ReliabilityIndexInput struct {
	PriorIndexBps    uint32
	SlashEvents      uint64
	DowntimeEpochs   uint64
	MissedTasks      uint64
	RejectedEvidence uint64
	RecoveryEpochs   uint64
}

type PosLayer string

const (
	PosLayerEconomicConsensus  PosLayer = "economic_consensus"
	PosLayerTaskAssignment     PosLayer = "task_assignment"
	PosLayerValidatorExecution PosLayer = "validator_execution"
	PosLayerStakingCapital     PosLayer = "staking_capital"
	PosLayerBaseCometBFT       PosLayer = "base_cometbft"
)

type PosLayerSpec struct {
	Layer            PosLayer
	Responsibilities []string
	DependsOn        []PosLayer
}

type LayeredPosArchitecture struct {
	Layers []PosLayerSpec
	Root   string
}

func DefaultLayeredPosArchitecture() LayeredPosArchitecture {
	layers := []PosLayerSpec{
		{
			Layer: PosLayerEconomicConsensus,
			Responsibilities: []string{
				"validator scoring",
				"performance incentives",
				"stake saturation",
				"role-specific reward weights",
				"slashing severity",
				"reporter incentives",
				"treasury, burn, and stabilization routing",
			},
			DependsOn: []PosLayer{PosLayerTaskAssignment, PosLayerValidatorExecution, PosLayerStakingCapital, PosLayerBaseCometBFT},
		},
		{
			Layer: PosLayerTaskAssignment,
			Responsibilities: []string{
				"workload grouping",
				"shard validator groups",
				"zone validator groups",
				"evidence verification subsets",
				"collator and verifier assignments",
			},
			DependsOn: []PosLayer{PosLayerValidatorExecution, PosLayerStakingCapital, PosLayerBaseCometBFT},
		},
		{
			Layer: PosLayerValidatorExecution,
			Responsibilities: []string{
				"block production",
				"state transition verification",
				"cross-domain proof verification",
				"signature production",
				"fault rejection",
			},
			DependsOn: []PosLayer{PosLayerStakingCapital, PosLayerBaseCometBFT},
		},
		{
			Layer: PosLayerStakingCapital,
			Responsibilities: []string{
				"validators",
				"delegators",
				"bonded stake",
				"unbonding",
				"redelegation",
				"capital risk preferences",
				"commission and delegation market metadata",
			},
			DependsOn: []PosLayer{PosLayerBaseCometBFT},
		},
		{
			Layer: PosLayerBaseCometBFT,
			Responsibilities: []string{
				"finality",
				"proposal and vote protocol",
				"validator public key set",
				"consensus safety and liveness",
			},
		},
	}
	architecture := LayeredPosArchitecture{Layers: layers}
	architecture.Root = ComputeLayeredPosArchitectureRoot(layers)
	return architecture
}

func (a LayeredPosArchitecture) Validate() error {
	if len(a.Layers) != len(DefaultPosLayerOrder()) {
		return errors.New("layered pos architecture must define all layers")
	}
	expectedOrder := DefaultPosLayerOrder()
	seen := make(map[PosLayer]int, len(a.Layers))
	for i, layer := range a.Layers {
		if layer.Layer != expectedOrder[i] {
			return fmt.Errorf("pos layer %d must be %s", i, expectedOrder[i])
		}
		if _, found := seen[layer.Layer]; found {
			return fmt.Errorf("duplicate pos layer %s", layer.Layer)
		}
		seen[layer.Layer] = i
		if err := layer.Validate(); err != nil {
			return err
		}
	}
	for _, layer := range a.Layers {
		layerIndex := seen[layer.Layer]
		for _, dependency := range layer.DependsOn {
			dependencyIndex, found := seen[dependency]
			if !found {
				return fmt.Errorf("pos layer %s depends on unknown layer %s", layer.Layer, dependency)
			}
			if dependencyIndex <= layerIndex {
				return fmt.Errorf("pos layer %s must depend only on lower layers", layer.Layer)
			}
		}
	}
	if err := validatePosHash("layered pos architecture root", a.Root); err != nil {
		return err
	}
	if expected := ComputeLayeredPosArchitectureRoot(a.Layers); a.Root != expected {
		return errors.New("layered pos architecture root mismatch")
	}
	return nil
}

func (s PosLayerSpec) Validate() error {
	if err := validatePosLayer(s.Layer); err != nil {
		return err
	}
	if len(s.Responsibilities) == 0 {
		return fmt.Errorf("pos layer %s responsibilities are required", s.Layer)
	}
	for _, responsibility := range s.Responsibilities {
		if err := validatePosResponsibility("pos layer responsibility", responsibility); err != nil {
			return err
		}
	}
	seen := make(map[PosLayer]struct{}, len(s.DependsOn))
	for _, dependency := range s.DependsOn {
		if err := validatePosLayer(dependency); err != nil {
			return err
		}
		if dependency == s.Layer {
			return fmt.Errorf("pos layer %s cannot depend on itself", s.Layer)
		}
		if _, found := seen[dependency]; found {
			return fmt.Errorf("duplicate dependency %s for pos layer %s", dependency, s.Layer)
		}
		seen[dependency] = struct{}{}
	}
	return nil
}

func DefaultPosLayerOrder() []PosLayer {
	return []PosLayer{
		PosLayerEconomicConsensus,
		PosLayerTaskAssignment,
		PosLayerValidatorExecution,
		PosLayerStakingCapital,
		PosLayerBaseCometBFT,
	}
}

func ComputeLayeredPosArchitectureRoot(layers []PosLayerSpec) string {
	return posHashRoot("aetheris-pos-layered-architecture-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(layers)))
		for _, layer := range layers {
			posWritePart(w, string(layer.Layer))
			posWriteUint64(w, uint64(len(layer.Responsibilities)))
			for _, responsibility := range layer.Responsibilities {
				posWritePart(w, responsibility)
			}
			posWriteUint64(w, uint64(len(layer.DependsOn)))
			for _, dependency := range layer.DependsOn {
				posWritePart(w, string(dependency))
			}
		}
	})
}

func DefaultEpochLifecycle() []EpochLifecycleStep {
	return []EpochLifecycleStep{
		{Phase: EpochPhaseDelegation, Name: "delegation phase", DurationKey: "delegation_phase_duration"},
		{Phase: EpochPhaseElection, Name: "validator election", DurationKey: "election_phase_duration"},
		{Phase: EpochPhaseAssignment, Name: "task group assignment", DurationKey: "assignment_phase_duration"},
		{Phase: EpochPhaseActive, Name: "active validation", DurationKey: "active_validation_duration"},
		{Phase: EpochPhaseSettlement, Name: "settlement + reward + slash finality", DurationKey: "settlement_phase_duration"},
	}
}

func ValidateEpochLifecycle(lifecycle []EpochLifecycleStep) error {
	expected := DefaultEpochLifecycle()
	if len(lifecycle) != len(expected) {
		return errors.New("epoch lifecycle must define every active phase")
	}
	seen := make(map[EpochPhase]struct{}, len(lifecycle))
	for i, step := range lifecycle {
		if step.Phase != expected[i].Phase {
			return fmt.Errorf("epoch lifecycle step %d must be %s", i, expected[i].Phase)
		}
		if _, found := seen[step.Phase]; found {
			return fmt.Errorf("duplicate epoch lifecycle phase %s", step.Phase)
		}
		seen[step.Phase] = struct{}{}
		if strings.TrimSpace(step.Name) != step.Name || step.Name == "" {
			return fmt.Errorf("epoch lifecycle phase %s name is required", step.Phase)
		}
		if strings.TrimSpace(step.DurationKey) != step.DurationKey || step.DurationKey == "" {
			return fmt.Errorf("epoch lifecycle phase %s duration key is required", step.Phase)
		}
	}
	return nil
}

func NextEpochPhase(phase EpochPhase) (EpochPhase, bool, error) {
	switch phase {
	case EpochPhaseDelegation:
		return EpochPhaseElection, false, nil
	case EpochPhaseElection:
		return EpochPhaseAssignment, false, nil
	case EpochPhaseAssignment:
		return EpochPhaseActive, false, nil
	case EpochPhaseActive:
		return EpochPhaseSettlement, false, nil
	case EpochPhaseSettlement:
		return EpochPhaseClosed, true, nil
	case EpochPhaseClosed:
		return EpochPhaseClosed, true, nil
	default:
		return "", false, fmt.Errorf("unsupported epoch phase %q", phase)
	}
}

func ValidateEpochPhaseTransition(from EpochPhase, to EpochPhase) error {
	next, _, err := NextEpochPhase(from)
	if err != nil {
		return err
	}
	if next != to {
		return fmt.Errorf("invalid epoch phase transition from %s to %s", from, to)
	}
	return nil
}

func (s EpochSeedSource) Validate() error {
	switch s {
	case EpochSeedSourcePreviousSeedValidatorSet, EpochSeedSourceCometBFTBlockID, EpochSeedSourceExternalBeacon:
		return nil
	default:
		return fmt.Errorf("unsupported epoch seed source %q", s)
	}
}

func (p Params) EffectiveEpochSeedSource() EpochSeedSource {
	if p.EpochSeedSource == "" {
		return EpochSeedSourcePreviousSeedValidatorSet
	}
	return p.EpochSeedSource
}

func MaxValidatorSetChanges(params Params, activeValidatorCount uint32) (uint32, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if activeValidatorCount == 0 {
		return 0, errors.New("active validator count must be positive")
	}
	changes := (uint64(activeValidatorCount)*uint64(params.MaxValidatorSetChangeRateBps) + uint64(BasisPoints) - 1) / uint64(BasisPoints)
	if changes == 0 {
		return 1, nil
	}
	if changes > uint64(activeValidatorCount) {
		return activeValidatorCount, nil
	}
	return uint32(changes), nil
}

func BuildElectionCandidates(params Params, electionEpoch uint64, candidates []Candidate, intents []DelegationIntent) ([]Candidate, []RejectedDelegationIntent, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	out := make([]Candidate, len(candidates))
	indexByID := make(map[string]int, len(candidates))
	for i, candidate := range candidates {
		cloned := cloneCandidate(candidate)
		if err := cloned.Validate(params); err != nil {
			return nil, nil, err
		}
		if _, found := indexByID[cloned.ValidatorID]; found {
			return nil, nil, fmt.Errorf("duplicate candidate %q", cloned.ValidatorID)
		}
		indexByID[cloned.ValidatorID] = i
		out[i] = cloned
	}
	activations, rejected, err := ActivateDelegationIntents(params, electionEpoch, out, intents)
	if err != nil {
		return nil, nil, err
	}
	for _, activation := range activations {
		idx, found := indexByID[activation.ValidatorID]
		if !found {
			continue
		}
		out[idx].Nominations = mergeNominations(out[idx].Nominations, activation.Nominations)
		out[idx].DelegatedStakeNaet = sumNominations(out[idx].Nominations)
	}
	return out, rejected, nil
}

func ComputePerformanceFactor(input PerformanceFactorInput) (uint32, error) {
	completionDenom := input.CompletedTasks + input.MissedTasks
	if completionDenom < input.CompletedTasks {
		return 0, errors.New("performance task count overflow")
	}
	correctnessDenom := input.CorrectVerifications + input.IncorrectVerifications
	if correctnessDenom < input.CorrectVerifications {
		return 0, errors.New("performance verification count overflow")
	}
	if input.CommittedWindows < input.AvailableWindows {
		return 0, errors.New("available windows cannot exceed committed windows")
	}
	completion := ratioBps(input.CompletedTasks, completionDenom)
	correctness := ratioBps(input.CorrectVerifications, correctnessDenom)
	availability := ratioBps(input.AvailableWindows, input.CommittedWindows)
	score := uint64(4_000)*uint64(completion) +
		uint64(4_000)*uint64(correctness) +
		uint64(2_000)*uint64(availability)
	return uint32(score / uint64(BasisPoints)), nil
}

func ComputeUptimeFactor(input UptimeFactorInput) (uint32, error) {
	if input.TotalBlocks < input.SignedBlocks {
		return 0, errors.New("signed blocks cannot exceed total blocks")
	}
	totalTaskParticipations := input.TaskParticipations + input.MissedTaskParticipations
	if totalTaskParticipations < input.TaskParticipations {
		return 0, errors.New("task participation count overflow")
	}
	blocks := ratioBps(input.SignedBlocks, input.TotalBlocks)
	tasks := ratioBps(input.TaskParticipations, totalTaskParticipations)
	score := uint64(7_000)*uint64(blocks) + uint64(3_000)*uint64(tasks)
	return uint32(score / uint64(BasisPoints)), nil
}

func ComputeLatencyFactor(input LatencyFactorInput) (uint32, error) {
	if !input.CommittedWindow {
		return 0, errors.New("latency factor requires committed measurement window")
	}
	if input.AdvisoryOnly {
		return BasisPoints, nil
	}
	if input.TargetMillis == 0 {
		return 0, errors.New("latency target must be positive")
	}
	if input.P95Millis == 0 || input.P95Millis <= input.TargetMillis {
		return BasisPoints, nil
	}
	return uint32(sdkmath.NewIntFromUint64(input.TargetMillis).MulRaw(int64(BasisPoints)).Quo(sdkmath.NewIntFromUint64(input.P95Millis)).Uint64()), nil
}

func ComputeReliabilityIndex(input ReliabilityIndexInput) (uint32, error) {
	if input.PriorIndexBps > BasisPoints {
		return 0, fmt.Errorf("prior reliability index must be <= %d bps", BasisPoints)
	}
	index := input.PriorIndexBps
	if index == 0 {
		index = BasisPoints
	}
	penalty, err := reliabilityPenalty(input)
	if err != nil {
		return 0, err
	}
	if penalty >= uint64(index) {
		index = 0
	} else {
		index -= uint32(penalty)
	}
	recovery := input.RecoveryEpochs * 100
	if recovery > uint64(BasisPoints-index) {
		return BasisPoints, nil
	}
	return index + uint32(recovery), nil
}

func reliabilityPenalty(input ReliabilityIndexInput) (uint64, error) {
	penalty := sdkmath.NewIntFromUint64(input.SlashEvents).MulRaw(2_000)
	penalty = penalty.Add(sdkmath.NewIntFromUint64(input.DowntimeEpochs).MulRaw(500))
	penalty = penalty.Add(sdkmath.NewIntFromUint64(input.MissedTasks).MulRaw(100))
	penalty = penalty.Add(sdkmath.NewIntFromUint64(input.RejectedEvidence).MulRaw(250))
	if !penalty.LTE(sdkmath.NewIntFromUint64(uint64(BasisPoints))) {
		return uint64(BasisPoints), nil
	}
	return penalty.Uint64(), nil
}

func EpochRecordFieldNames() []string {
	return []string{
		"epoch_id",
		"start_height",
		"end_height",
		"phase",
		"seed",
		"validator_set_hash",
		"task_group_root",
		"performance_root",
		"reward_root",
		"slash_root",
		"settlement_status",
	}
}

func EpochPhaseValues() []EpochPhase {
	return []EpochPhase{
		EpochPhaseDelegation,
		EpochPhaseElection,
		EpochPhaseAssignment,
		EpochPhaseActive,
		EpochPhaseSettlement,
		EpochPhaseClosed,
	}
}

func ValidateValidatorSetChangeActivation(epoch EpochRecord, activationHeight uint64) error {
	if err := epoch.Validate(); err != nil {
		return err
	}
	if activationHeight != epoch.StartHeight {
		return fmt.Errorf("validator set changes must activate at epoch boundary height %d", epoch.StartHeight)
	}
	return nil
}

func ValidateConsecutiveEpochs(previous EpochRecord, next EpochRecord) error {
	if err := previous.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if next.EpochID != previous.EpochID+1 {
		return errors.New("next epoch id must increment by one")
	}
	if next.StartHeight != previous.EndHeight+1 {
		return errors.New("next epoch must start at previous end height plus one")
	}
	return nil
}

func DelegationEffectiveElectionEpoch(params Params, requestedEpoch uint64) (uint64, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	return requestedEpoch + params.DelegationActivationEpochs, nil
}

func DelegationAffectsElection(params Params, requestedEpoch uint64, electionEpoch uint64) (bool, error) {
	effectiveEpoch, err := DelegationEffectiveElectionEpoch(params, requestedEpoch)
	if err != nil {
		return false, err
	}
	return electionEpoch >= effectiveEpoch, nil
}

func EvidenceWithinSlashableWindow(params Params, evidenceEpoch uint64, currentEpoch uint64) (bool, error) {
	if err := params.Validate(); err != nil {
		return false, err
	}
	if currentEpoch < evidenceEpoch {
		return false, errors.New("current epoch cannot be before evidence epoch")
	}
	return currentEpoch-evidenceEpoch <= params.EvidenceWindowEpochs, nil
}

func DefaultEpochPhaseDurations(epochDurationSeconds uint64) EpochPhaseDurations {
	delegation := epochDurationSeconds / 4
	election := epochDurationSeconds / 12
	assignment := epochDurationSeconds / 12
	settlement := epochDurationSeconds / 12
	active := epochDurationSeconds - delegation - election - assignment - settlement
	return EpochPhaseDurations{
		DelegationSeconds:       delegation,
		ElectionSeconds:         election,
		AssignmentSeconds:       assignment,
		ActiveValidationSeconds: active,
		SettlementSeconds:       settlement,
	}
}

func (p Params) EffectivePhaseDurations() EpochPhaseDurations {
	baseDefault := DefaultEpochPhaseDurations(DefaultEpochDurationSeconds)
	if p.PhaseDurations.IsZero() ||
		(p.EpochDurationSeconds != DefaultEpochDurationSeconds && p.PhaseDurations == baseDefault) {
		return DefaultEpochPhaseDurations(p.EpochDurationSeconds)
	}
	return p.PhaseDurations
}

func (d EpochPhaseDurations) IsZero() bool {
	return d.DelegationSeconds == 0 &&
		d.ElectionSeconds == 0 &&
		d.AssignmentSeconds == 0 &&
		d.ActiveValidationSeconds == 0 &&
		d.SettlementSeconds == 0
}

func (d EpochPhaseDurations) TotalSeconds() uint64 {
	return d.DelegationSeconds +
		d.ElectionSeconds +
		d.AssignmentSeconds +
		d.ActiveValidationSeconds +
		d.SettlementSeconds
}

func (d EpochPhaseDurations) Validate(epochDurationSeconds uint64) error {
	if d.DelegationSeconds == 0 {
		return errors.New("delegation phase duration must be positive")
	}
	if d.ElectionSeconds == 0 {
		return errors.New("election phase duration must be positive")
	}
	if d.AssignmentSeconds == 0 {
		return errors.New("assignment phase duration must be positive")
	}
	if d.ActiveValidationSeconds == 0 {
		return errors.New("active validation phase duration must be positive")
	}
	if d.SettlementSeconds == 0 {
		return errors.New("settlement phase duration must be positive")
	}
	if d.TotalSeconds() != epochDurationSeconds {
		return fmt.Errorf("epoch phase durations must sum to %d seconds", epochDurationSeconds)
	}
	return nil
}

func EpochPhaseAt(params Params, epochStartUnixSeconds uint64, nowUnixSeconds uint64) (EpochPhase, error) {
	if err := params.Validate(); err != nil {
		return "", err
	}
	if nowUnixSeconds < epochStartUnixSeconds {
		return "", errors.New("epoch phase time cannot be before epoch start")
	}
	elapsed := nowUnixSeconds - epochStartUnixSeconds
	if elapsed >= params.EpochDurationSeconds {
		return EpochPhaseClosed, nil
	}
	durations := params.EffectivePhaseDurations()
	if elapsed < durations.DelegationSeconds {
		return EpochPhaseDelegation, nil
	}
	elapsed -= durations.DelegationSeconds
	if elapsed < durations.ElectionSeconds {
		return EpochPhaseElection, nil
	}
	elapsed -= durations.ElectionSeconds
	if elapsed < durations.AssignmentSeconds {
		return EpochPhaseAssignment, nil
	}
	elapsed -= durations.AssignmentSeconds
	if elapsed < durations.ActiveValidationSeconds {
		return EpochPhaseActive, nil
	}
	return EpochPhaseSettlement, nil
}

func NewEpochRecord(params Params, epochID uint64, startHeight uint64, endHeight uint64, phase EpochPhase, previousSeed string, validators []ScoredValidator) (EpochRecord, error) {
	if err := params.Validate(); err != nil {
		return EpochRecord{}, err
	}
	if startHeight == 0 || endHeight < startHeight {
		return EpochRecord{}, errors.New("epoch heights must be positive and ordered")
	}
	if err := validateEpochPhase(phase); err != nil {
		return EpochRecord{}, err
	}
	validatorSetHash, err := ComputeValidatorSetHash(validators)
	if err != nil {
		return EpochRecord{}, err
	}
	seed, err := DeriveEpochSeedWithSource(params.EffectiveEpochSeedSource(), epochID, startHeight, previousSeed, validatorSetHash)
	if err != nil {
		return EpochRecord{}, err
	}
	record := EpochRecord{
		EpochID:          epochID,
		StartHeight:      startHeight,
		EndHeight:        endHeight,
		Phase:            phase,
		Seed:             seed,
		ValidatorSetHash: validatorSetHash,
		TaskGroupRoot:    PosEmptyRootHash,
		PerformanceRoot:  PosEmptyRootHash,
		RewardRoot:       PosEmptyRootHash,
		SlashRoot:        PosEmptyRootHash,
		SettlementStatus: SettlementStatusPending,
	}
	if err := record.Validate(); err != nil {
		return EpochRecord{}, err
	}
	return record, nil
}

func CloseEpochRecord(record EpochRecord, performanceRoot string, rewardRoot string, slashRoot string) (EpochRecord, error) {
	if err := record.Validate(); err != nil {
		return EpochRecord{}, err
	}
	if record.Phase != EpochPhaseSettlement {
		return EpochRecord{}, errors.New("epoch must be in settlement phase before closing")
	}
	if err := validatePosHash("performance root", performanceRoot); err != nil {
		return EpochRecord{}, err
	}
	if err := validatePosHash("reward root", rewardRoot); err != nil {
		return EpochRecord{}, err
	}
	if err := validatePosHash("slash root", slashRoot); err != nil {
		return EpochRecord{}, err
	}
	record.Phase = EpochPhaseClosed
	record.PerformanceRoot = performanceRoot
	record.RewardRoot = rewardRoot
	record.SlashRoot = slashRoot
	record.SettlementStatus = SettlementStatusFinalized
	return record, record.Validate()
}

func (r EpochRecord) Validate() error {
	if r.StartHeight == 0 || r.EndHeight < r.StartHeight {
		return errors.New("epoch heights must be positive and ordered")
	}
	if err := validateEpochPhase(r.Phase); err != nil {
		return err
	}
	if err := validatePosHash("epoch seed", r.Seed); err != nil {
		return err
	}
	if err := validatePosHash("validator set hash", r.ValidatorSetHash); err != nil {
		return err
	}
	if err := validatePosHash("task group root", r.TaskGroupRoot); err != nil {
		return err
	}
	if err := validatePosHash("performance root", r.PerformanceRoot); err != nil {
		return err
	}
	if err := validatePosHash("reward root", r.RewardRoot); err != nil {
		return err
	}
	if err := validatePosHash("slash root", r.SlashRoot); err != nil {
		return err
	}
	switch r.SettlementStatus {
	case SettlementStatusPending, SettlementStatusFinalized:
	default:
		return errors.New("unsupported settlement status")
	}
	if r.Phase == EpochPhaseClosed && r.SettlementStatus != SettlementStatusFinalized {
		return errors.New("closed epoch must have finalized settlement")
	}
	if r.SettlementStatus == SettlementStatusFinalized && r.Phase != EpochPhaseClosed {
		return errors.New("finalized settlement must close the epoch")
	}
	return nil
}

func ComputeValidatorSetHash(validators []ScoredValidator) (string, error) {
	ordered := cloneScoredValidators(validators)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].ValidatorID < ordered[j].ValidatorID
	})
	seen := make(map[string]struct{}, len(ordered))
	for _, validator := range ordered {
		if err := validatePosToken("validator id", validator.ValidatorID); err != nil {
			return "", err
		}
		if validator.VotingPowerNaet.IsNegative() {
			return "", errors.New("validator voting power cannot be negative")
		}
		if validator.Score.IsNegative() {
			return "", errors.New("validator score cannot be negative")
		}
		if _, found := seen[validator.ValidatorID]; found {
			return "", fmt.Errorf("duplicate validator %q", validator.ValidatorID)
		}
		seen[validator.ValidatorID] = struct{}{}
		if err := validateValidatorRoles(validator.Roles); err != nil {
			return "", err
		}
	}
	return posHashRoot("aetheris-pos-validator-set-v1", func(w posByteWriter) {
		posWriteUint64(w, uint64(len(ordered)))
		for _, validator := range ordered {
			posWritePart(w, validator.ValidatorID)
			posWritePart(w, validator.VotingPowerNaet.String())
			posWritePart(w, validator.Score.String())
			for _, role := range normalizedRoles(validator.Roles, AllValidatorRoles()) {
				posWritePart(w, string(role))
			}
		}
	}), nil
}

func DeriveEpochSeed(epochID uint64, startHeight uint64, previousSeed string, validatorSetHash string) (string, error) {
	return DeriveEpochSeedWithSource(EpochSeedSourcePreviousSeedValidatorSet, epochID, startHeight, previousSeed, validatorSetHash)
}

func DeriveEpochSeedWithSource(source EpochSeedSource, epochID uint64, startHeight uint64, previousSeed string, validatorSetHash string) (string, error) {
	if err := source.Validate(); err != nil {
		return "", err
	}
	if startHeight == 0 {
		return "", errors.New("epoch seed start height must be positive")
	}
	if previousSeed == "" {
		previousSeed = PosEmptyRootHash
	}
	if err := validatePosHash("previous epoch seed", previousSeed); err != nil {
		return "", err
	}
	if err := validatePosHash("validator set hash", validatorSetHash); err != nil {
		return "", err
	}
	return posHashRoot("aetheris-pos-epoch-seed-v1", func(w posByteWriter) {
		posWritePart(w, string(source))
		posWriteUint64(w, epochID)
		posWriteUint64(w, startHeight)
		posWritePart(w, previousSeed)
		posWritePart(w, validatorSetHash)
	}), nil
}

func BuildTaskAssignments(params Params, epoch EpochRecord, validators []ScoredValidator, tasks []WorkloadTask) (TaskAssignmentSet, error) {
	if err := params.Validate(); err != nil {
		return TaskAssignmentSet{}, err
	}
	if err := epoch.Validate(); err != nil {
		return TaskAssignmentSet{}, err
	}
	if len(validators) == 0 {
		return TaskAssignmentSet{}, errors.New("task assignment requires active validators")
	}
	validatorSetHash, err := ComputeValidatorSetHash(validators)
	if err != nil {
		return TaskAssignmentSet{}, err
	}
	if validatorSetHash != epoch.ValidatorSetHash {
		return TaskAssignmentSet{}, errors.New("task assignments require committed validator set hash")
	}
	if len(tasks) == 0 {
		return TaskAssignmentSet{
			EpochID: epoch.EpochID,
			Seed:    epoch.Seed,
			Root:    PosEmptyRootHash,
		}, nil
	}

	orderedTasks := make([]WorkloadTask, len(tasks))
	for i, task := range tasks {
		normalized := normalizeWorkloadTask(params, task)
		if err := normalized.Validate(params); err != nil {
			return TaskAssignmentSet{}, err
		}
		orderedTasks[i] = normalized
	}
	sort.SliceStable(orderedTasks, func(i, j int) bool {
		return compareWorkloadTasks(orderedTasks[i], orderedTasks[j]) < 0
	})

	assignments := make([]TaskAssignment, 0)
	for _, task := range orderedTasks {
		for _, role := range normalizedRoles(task.Roles, DefaultTaskRoles()) {
			eligible := validatorsForRole(validators, role)
			if uint32(len(eligible)) < task.RequiredValidators {
				return TaskAssignmentSet{}, fmt.Errorf("insufficient validators for task %s role %s", task.TaskID, role)
			}
			selected := selectTaskValidatorIDs(epoch.Seed, task, role, eligible, task.RequiredValidators)
			assignment := TaskAssignment{
				TaskID:        task.TaskID,
				WorkloadID:    task.WorkloadID,
				WorkloadType:  task.WorkloadType,
				ZoneID:        task.ZoneID,
				ShardID:       task.ShardID,
				WorkloadClass: task.WorkloadClass,
				Role:          role,
				Validators:    selected,
			}
			assignment.AssignmentHash = ComputeTaskAssignmentHash(epoch.EpochID, epoch.Seed, assignment)
			assignments = append(assignments, assignment)
		}
	}
	sort.SliceStable(assignments, func(i, j int) bool {
		return compareTaskAssignments(assignments[i], assignments[j]) < 0
	})
	root := ComputeTaskAssignmentRoot(epoch.EpochID, epoch.Seed, assignments)
	out := TaskAssignmentSet{EpochID: epoch.EpochID, Seed: epoch.Seed, Assignments: assignments, Root: root}
	return out, out.Validate()
}

func BuildTaskGroups(params Params, epoch EpochRecord, validators []ScoredValidator, tasks []WorkloadTask, activationHeight uint64, expiryHeight uint64) (TaskGroupSet, error) {
	if activationHeight == 0 {
		return TaskGroupSet{}, errors.New("task group activation height is required")
	}
	if expiryHeight <= activationHeight {
		return TaskGroupSet{}, errors.New("task group expiry height must be after activation height")
	}
	assignments, err := BuildTaskAssignments(params, epoch, validators, tasks)
	if err != nil {
		return TaskGroupSet{}, err
	}
	if len(tasks) == 0 {
		return TaskGroupSet{EpochID: epoch.EpochID, Seed: epoch.Seed, Root: PosEmptyRootHash}, nil
	}
	validatorByID := make(map[string]ScoredValidator, len(validators))
	for _, validator := range validators {
		validatorByID[validator.ValidatorID] = validator
	}
	taskByID := make(map[string]WorkloadTask, len(tasks))
	for _, task := range tasks {
		normalized := normalizeWorkloadTask(params, task)
		taskByID[taskKey(normalized)] = normalized
	}
	assignmentsByTask := make(map[string][]TaskAssignment)
	for _, assignment := range assignments.Assignments {
		key := taskKey(WorkloadTask{
			TaskID:        assignment.TaskID,
			WorkloadID:    assignment.WorkloadID,
			WorkloadType:  assignment.WorkloadType,
			ZoneID:        assignment.ZoneID,
			ShardID:       assignment.ShardID,
			WorkloadClass: assignment.WorkloadClass,
		})
		assignmentsByTask[key] = append(assignmentsByTask[key], assignment)
	}
	groups := make([]TaskGroup, 0, len(taskByID))
	taskKeys := sortedStringKeys(taskByID)
	for _, key := range taskKeys {
		task := taskByID[key]
		taskAssignments := assignmentsByTask[key]
		members := taskGroupMembers(taskAssignments)
		verifiers := taskGroupVerifiers(taskAssignments, members)
		group := TaskGroup{
			EpochID:          epoch.EpochID,
			WorkloadID:       task.WorkloadID,
			WorkloadType:     task.WorkloadType,
			ValidatorMembers: members,
			ProposerOrder:    taskGroupProposerOrder(epoch.Seed, task, members),
			VerifierSet:      verifiers,
			MinimumGroupSize: task.RequiredValidators,
			StakeWeightRoot:  ComputeTaskGroupStakeWeightRoot(epoch.EpochID, task, members, validatorByID),
			AssignmentSeed:   epoch.Seed,
			ActivationHeight: activationHeight,
			ExpiryHeight:     expiryHeight,
		}
		group.TaskGroupID = ComputeTaskGroupID(group)
		groups = append(groups, group)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return compareTaskGroups(groups[i], groups[j]) < 0
	})
	out := TaskGroupSet{
		EpochID: epoch.EpochID,
		Seed:    epoch.Seed,
		Groups:  groups,
		Root:    ComputeTaskGroupRoot(epoch.EpochID, epoch.Seed, groups),
	}
	return out, out.Validate()
}

func (t WorkloadTask) Validate(params Params) error {
	if err := validatePosToken("task id", t.TaskID); err != nil {
		return err
	}
	if err := validatePosToken("task workload id", t.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(t.WorkloadType); err != nil {
		return err
	}
	if err := validatePosToken("task zone id", t.ZoneID); err != nil {
		return err
	}
	if err := validatePosToken("task shard id", t.ShardID); err != nil {
		return err
	}
	if err := validatePosToken("task workload class", t.WorkloadClass); err != nil {
		return err
	}
	if t.RequiredValidators < params.MinTaskGroupValidators {
		return fmt.Errorf("task validators must be at least %d", params.MinTaskGroupValidators)
	}
	if t.RequiredValidators > params.MaxTaskGroupValidators {
		return fmt.Errorf("task validators must be <= %d", params.MaxTaskGroupValidators)
	}
	return validateValidatorRoles(t.Roles)
}

func (a TaskAssignment) Validate() error {
	if err := validatePosToken("assignment task id", a.TaskID); err != nil {
		return err
	}
	if err := validatePosToken("assignment workload id", a.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(a.WorkloadType); err != nil {
		return err
	}
	if err := validatePosToken("assignment zone id", a.ZoneID); err != nil {
		return err
	}
	if err := validatePosToken("assignment shard id", a.ShardID); err != nil {
		return err
	}
	if err := validatePosToken("assignment workload class", a.WorkloadClass); err != nil {
		return err
	}
	if err := validateValidatorRole(a.Role); err != nil {
		return err
	}
	if len(a.Validators) == 0 {
		return errors.New("assignment validators are required")
	}
	seen := make(map[string]struct{}, len(a.Validators))
	var previous string
	for i, validatorID := range a.Validators {
		if err := validatePosToken("assignment validator id", validatorID); err != nil {
			return err
		}
		if _, found := seen[validatorID]; found {
			return fmt.Errorf("duplicate assignment validator %q", validatorID)
		}
		seen[validatorID] = struct{}{}
		if i > 0 && previous >= validatorID {
			return errors.New("assignment validators must be sorted canonically")
		}
		previous = validatorID
	}
	return validatePosHash("assignment hash", a.AssignmentHash)
}

func (s TaskAssignmentSet) Validate() error {
	if err := validatePosHash("assignment seed", s.Seed); err != nil {
		return err
	}
	if err := validatePosHash("assignment root", s.Root); err != nil {
		return err
	}
	for i, assignment := range s.Assignments {
		if err := assignment.Validate(); err != nil {
			return err
		}
		expectedHash := ComputeTaskAssignmentHash(s.EpochID, s.Seed, assignment)
		if assignment.AssignmentHash != expectedHash {
			return errors.New("assignment hash mismatch")
		}
		if i > 0 && compareTaskAssignments(s.Assignments[i-1], assignment) >= 0 {
			return errors.New("task assignments must be sorted canonically")
		}
	}
	expectedRoot := PosEmptyRootHash
	if len(s.Assignments) > 0 {
		expectedRoot = ComputeTaskAssignmentRoot(s.EpochID, s.Seed, s.Assignments)
	}
	if s.Root != expectedRoot {
		return errors.New("task assignment root mismatch")
	}
	return nil
}

func (g TaskGroup) Validate() error {
	if g.EpochID == 0 {
		return errors.New("task group epoch id is required")
	}
	if err := validatePosToken("task group id", g.TaskGroupID); err != nil {
		return err
	}
	if err := validatePosToken("task group workload id", g.WorkloadID); err != nil {
		return err
	}
	if err := validateWorkloadType(g.WorkloadType); err != nil {
		return err
	}
	if len(g.ValidatorMembers) < int(g.MinimumGroupSize) {
		return errors.New("task group members below minimum group size")
	}
	if err := validateCanonicalValidatorIDs("task group member", g.ValidatorMembers); err != nil {
		return err
	}
	if len(g.ProposerOrder) != len(g.ValidatorMembers) {
		return errors.New("task group proposer order must include every member")
	}
	if err := validateValidatorIDSet("task group proposer", g.ProposerOrder, g.ValidatorMembers); err != nil {
		return err
	}
	if len(g.VerifierSet) == 0 {
		return errors.New("task group verifier set is required")
	}
	if err := validateCanonicalValidatorIDs("task group verifier", g.VerifierSet); err != nil {
		return err
	}
	if err := validateValidatorIDSubset("task group verifier", g.VerifierSet, g.ValidatorMembers); err != nil {
		return err
	}
	if err := validatePosHash("task group stake weight root", g.StakeWeightRoot); err != nil {
		return err
	}
	if err := validatePosHash("task group assignment seed", g.AssignmentSeed); err != nil {
		return err
	}
	if g.ActivationHeight == 0 {
		return errors.New("task group activation height is required")
	}
	if g.ExpiryHeight <= g.ActivationHeight {
		return errors.New("task group expiry height must be after activation height")
	}
	if expected := ComputeTaskGroupID(g); g.TaskGroupID != expected {
		return errors.New("task group id mismatch")
	}
	return nil
}

func (s TaskGroupSet) Validate() error {
	if err := validatePosHash("task group set seed", s.Seed); err != nil {
		return err
	}
	if err := validatePosHash("task group set root", s.Root); err != nil {
		return err
	}
	for i, group := range s.Groups {
		if err := group.Validate(); err != nil {
			return err
		}
		if i > 0 && compareTaskGroups(s.Groups[i-1], group) >= 0 {
			return errors.New("task groups must be sorted canonically")
		}
	}
	expectedRoot := PosEmptyRootHash
	if len(s.Groups) > 0 {
		expectedRoot = ComputeTaskGroupRoot(s.EpochID, s.Seed, s.Groups)
	}
	if s.Root != expectedRoot {
		return errors.New("task group root mismatch")
	}
	return nil
}

func ComputeTaskAssignmentHash(epochID uint64, seed string, assignment TaskAssignment) string {
	return posHashRoot("aetheris-pos-task-assignment-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWritePart(w, assignment.TaskID)
		posWritePart(w, assignment.WorkloadID)
		posWritePart(w, string(assignment.WorkloadType))
		posWritePart(w, assignment.ZoneID)
		posWritePart(w, assignment.ShardID)
		posWritePart(w, assignment.WorkloadClass)
		posWritePart(w, string(assignment.Role))
		posWriteUint64(w, uint64(len(assignment.Validators)))
		for _, validatorID := range assignment.Validators {
			posWritePart(w, validatorID)
		}
	})
}

func ComputeTaskAssignmentRoot(epochID uint64, seed string, assignments []TaskAssignment) string {
	return posHashRoot("aetheris-pos-task-assignment-root-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWriteUint64(w, uint64(len(assignments)))
		for _, assignment := range assignments {
			posWritePart(w, assignment.AssignmentHash)
		}
	})
}

func ComputeTaskGroupID(group TaskGroup) string {
	return posHashRoot("aetheris-pos-task-group-id-v1", func(w posByteWriter) {
		posWriteUint64(w, group.EpochID)
		posWritePart(w, group.WorkloadID)
		posWritePart(w, string(group.WorkloadType))
		posWritePart(w, group.AssignmentSeed)
		posWriteUint64(w, group.ActivationHeight)
		posWriteUint64(w, group.ExpiryHeight)
	})
}

func ComputeTaskGroupStakeWeightRoot(epochID uint64, task WorkloadTask, members []string, validators map[string]ScoredValidator) string {
	return posHashRoot("aetheris-pos-task-group-stake-root-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, task.TaskID)
		posWritePart(w, task.WorkloadID)
		posWritePart(w, string(task.WorkloadType))
		posWriteUint64(w, uint64(len(members)))
		for _, validatorID := range members {
			validator := validators[validatorID]
			posWritePart(w, validatorID)
			posWritePart(w, validator.ScoreComponents.StakeWeightNaet.String())
			posWritePart(w, validator.VotingPowerNaet.String())
		}
	})
}

func ComputeTaskGroupRoot(epochID uint64, seed string, groups []TaskGroup) string {
	return posHashRoot("aetheris-pos-task-group-root-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWriteUint64(w, uint64(len(groups)))
		for _, group := range groups {
			posWritePart(w, group.TaskGroupID)
			posWritePart(w, group.StakeWeightRoot)
		}
	})
}

func ActivateDelegationIntents(params Params, electionEpoch uint64, candidates []Candidate, intents []DelegationIntent) ([]DelegationActivation, []RejectedDelegationIntent, error) {
	if err := params.Validate(); err != nil {
		return nil, nil, err
	}
	candidateByID := make(map[string]Candidate, len(candidates))
	for _, candidate := range candidates {
		id := strings.TrimSpace(candidate.ValidatorID)
		if err := validatePosToken("validator id", id); err != nil {
			return nil, nil, err
		}
		if _, found := candidateByID[id]; found {
			return nil, nil, fmt.Errorf("duplicate candidate %q", id)
		}
		candidate.ValidatorID = id
		candidateByID[id] = candidate
	}

	ordered := make([]DelegationIntent, len(intents))
	copy(ordered, intents)
	sort.SliceStable(ordered, func(i, j int) bool {
		return compareDelegationIntents(ordered[i], ordered[j]) < 0
	})

	nominationsByValidator := make(map[string][]Nomination)
	seenNomination := make(map[string]struct{}, len(ordered))
	rejected := make([]RejectedDelegationIntent, 0)
	for _, intent := range ordered {
		if err := intent.Validate(params); err != nil {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: err.Error()})
			continue
		}
		if electionEpoch < intent.RequestedEpoch+params.DelegationActivationEpochs {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "delegation activation delay has not elapsed"})
			continue
		}
		candidate, found := candidateByID[intent.ValidatorID]
		if !found {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator is not in election market"})
			continue
		}
		if candidate.Jailed || candidate.Tombstoned {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator is not eligible for delegation"})
			continue
		}
		if candidate.CommissionBps > intent.MaxCommissionBps {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator commission exceeds delegation risk profile"})
			continue
		}
		if candidate.PerformanceScoreBps < intent.MinPerformanceScoreBps {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "validator performance below delegation risk profile"})
			continue
		}
		nominationKey := intent.ValidatorID + "\x00" + intent.NominatorID
		if _, found := seenNomination[nominationKey]; found {
			rejected = append(rejected, RejectedDelegationIntent{Intent: intent, Reason: "duplicate delegation intent for validator"})
			continue
		}
		seenNomination[nominationKey] = struct{}{}
		nominationsByValidator[intent.ValidatorID] = append(nominationsByValidator[intent.ValidatorID], Nomination{
			NominatorID: intent.NominatorID,
			StakeNaet:   intent.StakeNaet,
		})
	}

	validatorIDs := make([]string, 0, len(nominationsByValidator))
	for validatorID := range nominationsByValidator {
		validatorIDs = append(validatorIDs, validatorID)
	}
	sort.Strings(validatorIDs)

	activations := make([]DelegationActivation, 0, len(validatorIDs))
	for _, validatorID := range validatorIDs {
		nominations := sortNominations(nominationsByValidator[validatorID])
		totalStake := sumNominations(nominations)
		activation := DelegationActivation{
			ValidatorID:   validatorID,
			Nominations:   nominations,
			ActivatedAt:   electionEpoch,
			IntentCount:   uint32(len(nominations)),
			TotalStake:    totalStake,
			ActivationKey: computeDelegationActivationKey(electionEpoch, validatorID, nominations),
		}
		activations = append(activations, activation)
	}
	return activations, rejected, nil
}

func (i DelegationIntent) Validate(params Params) error {
	if err := validatePosToken("nominator id", i.NominatorID); err != nil {
		return err
	}
	if err := validatePosToken("validator id", i.ValidatorID); err != nil {
		return err
	}
	if !i.StakeNaet.IsPositive() {
		return errors.New("delegation intent stake must be positive")
	}
	if i.MaxCommissionBps > params.MaxCommissionBps {
		return fmt.Errorf("delegation max commission must be <= %d bps", params.MaxCommissionBps)
	}
	if i.MinPerformanceScoreBps > BasisPoints {
		return fmt.Errorf("delegation minimum performance must be <= %d bps", BasisPoints)
	}
	return nil
}

func SettleEvidenceCase(params Params, currentEpoch uint64, evidence EvidenceCase, selfStake sdkmath.Int, nominations []Nomination) (EvidenceSettlement, error) {
	if err := evidence.Validate(params, currentEpoch); err != nil {
		return EvidenceSettlement{}, err
	}
	slash, err := ComputeSlash(SlashInput{
		ValidatorID:       evidence.ValidatorID,
		Misbehavior:       evidence.Misbehavior,
		SlashFractionBps:  evidence.SlashFractionBps,
		SelfStakeNaet:     selfStake,
		Nominations:       nominations,
		EvidenceHeight:    evidence.EvidenceHeight,
		EvidenceFinalized: true,
	})
	if err != nil {
		return EvidenceSettlement{}, err
	}
	reporterReward := mulIntBps(slash.TotalSlashedNaet, params.ReporterRewardBps)
	settlement := EvidenceSettlement{
		EvidenceID:         evidence.EvidenceID,
		ReporterID:         evidence.ReporterID,
		Slash:              slash,
		ReporterRewardNaet: reporterReward,
		BurnNaet:           slash.TotalSlashedNaet.Sub(reporterReward),
	}
	settlement.SettlementHash = computeEvidenceSettlementHash(settlement)
	return settlement, nil
}

func (e EvidenceCase) Validate(params Params, currentEpoch uint64) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validatePosToken("evidence id", e.EvidenceID); err != nil {
		return err
	}
	if err := validatePosToken("evidence reporter id", e.ReporterID); err != nil {
		return err
	}
	if err := validatePosToken("evidence validator id", e.ValidatorID); err != nil {
		return err
	}
	if !IsSlashableMisbehavior(e.Misbehavior) {
		return fmt.Errorf("unsupported misbehavior %q", e.Misbehavior)
	}
	if e.SlashFractionBps == 0 || e.SlashFractionBps > BasisPoints {
		return fmt.Errorf("slash fraction must be within 1..%d bps", BasisPoints)
	}
	if e.EvidenceHeight < 0 {
		return errors.New("evidence height cannot be negative")
	}
	if !e.Finalized {
		return errors.New("evidence must be finalized before settlement")
	}
	withinWindow, err := EvidenceWithinSlashableWindow(params, e.EvidenceEpoch, currentEpoch)
	if err != nil {
		return err
	}
	if !withinWindow {
		return errors.New("evidence is outside slashable window")
	}
	return nil
}

func DefaultRoleRewardWeights() []RoleRewardWeight {
	return []RoleRewardWeight{
		{Role: ValidatorRoleBlockProducer, WeightBps: 3_500},
		{Role: ValidatorRoleVerifier, WeightBps: 3_500},
		{Role: ValidatorRoleCollator, WeightBps: 1_500},
		{Role: ValidatorRoleEvidenceReviewer, WeightBps: 1_500},
	}
}

func SettleWorkloadRewards(input WorkloadRewardInput) (WorkloadRewardSettlement, error) {
	if input.TotalRewardsNaet.IsNegative() {
		return WorkloadRewardSettlement{}, errors.New("workload rewards cannot be negative")
	}
	weights := input.RoleWeights
	if len(weights) == 0 {
		weights = DefaultRoleRewardWeights()
	}
	if err := validateRoleRewardWeights(weights); err != nil {
		return WorkloadRewardSettlement{}, err
	}

	outcomesByRole := make(map[ValidatorRole][]AssignmentOutcome)
	for _, outcome := range input.Outcomes {
		if err := outcome.Validate(); err != nil {
			return WorkloadRewardSettlement{}, err
		}
		outcomesByRole[outcome.Role] = append(outcomesByRole[outcome.Role], outcome)
	}

	rewardByValidator := make(map[string]sdkmath.Int)
	workUnitsByValidator := make(map[string]uint64)
	remainder := sdkmath.ZeroInt()
	completedUnits := uint64(0)

	for _, weight := range weights {
		roleBudget := mulIntBps(input.TotalRewardsNaet, weight.WeightBps)
		roleUnitsByValidator := make(map[string]uint64)
		totalRoleUnits := uint64(0)
		for _, outcome := range outcomesByRole[weight.Role] {
			if !outcome.Completed || outcome.Faulted || outcome.WorkUnits == 0 {
				continue
			}
			roleUnitsByValidator[outcome.ValidatorID] += outcome.WorkUnits
			workUnitsByValidator[outcome.ValidatorID] += outcome.WorkUnits
			totalRoleUnits += outcome.WorkUnits
			completedUnits += outcome.WorkUnits
		}
		if totalRoleUnits == 0 {
			remainder = remainder.Add(roleBudget)
			continue
		}
		validatorIDs := sortedStringKeys(roleUnitsByValidator)
		distributed := sdkmath.ZeroInt()
		for _, validatorID := range validatorIDs {
			reward := roleBudget.MulRaw(int64(roleUnitsByValidator[validatorID])).QuoRaw(int64(totalRoleUnits))
			currentReward, found := rewardByValidator[validatorID]
			if !found {
				currentReward = sdkmath.ZeroInt()
			}
			rewardByValidator[validatorID] = currentReward.Add(reward)
			distributed = distributed.Add(reward)
		}
		remainder = remainder.Add(roleBudget.Sub(distributed))
	}

	validatorIDs := sortedStringKeys(rewardByValidator)
	rewards := make([]ValidatorWorkloadReward, 0, len(validatorIDs))
	for _, validatorID := range validatorIDs {
		rewards = append(rewards, ValidatorWorkloadReward{
			ValidatorID: validatorID,
			RewardNaet:  rewardByValidator[validatorID],
			WorkUnits:   workUnitsByValidator[validatorID],
		})
	}
	settlement := WorkloadRewardSettlement{
		EpochID:        input.EpochID,
		Rewards:        rewards,
		RemainderNaet:  remainder,
		CompletedUnits: completedUnits,
	}
	settlement.RewardRoot = computeWorkloadRewardRoot(settlement)
	return settlement, nil
}

func (o AssignmentOutcome) Validate() error {
	if err := validatePosToken("assignment outcome task id", o.TaskID); err != nil {
		return err
	}
	if err := validateValidatorRole(o.Role); err != nil {
		return err
	}
	if err := validatePosToken("assignment outcome validator id", o.ValidatorID); err != nil {
		return err
	}
	if o.Completed && o.Faulted {
		return errors.New("assignment outcome cannot be both completed and faulted")
	}
	return nil
}

func ValidatorSupportsRole(candidate Candidate, role ValidatorRole) bool {
	if err := validateValidatorRole(role); err != nil {
		return false
	}
	if len(candidate.Roles) == 0 {
		return true
	}
	for _, candidateRole := range candidate.Roles {
		if candidateRole == role {
			return true
		}
	}
	return false
}

func AllValidatorRoles() []ValidatorRole {
	return []ValidatorRole{
		ValidatorRoleBlockProducer,
		ValidatorRoleVerifier,
		ValidatorRoleCollator,
		ValidatorRoleEvidenceReviewer,
	}
}

func DefaultTaskRoles() []ValidatorRole {
	return []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier}
}

func validateEpochPhase(phase EpochPhase) error {
	switch phase {
	case EpochPhaseDelegation, EpochPhaseElection, EpochPhaseAssignment, EpochPhaseActive, EpochPhaseSettlement, EpochPhaseClosed:
		return nil
	default:
		return fmt.Errorf("unsupported epoch phase %q", phase)
	}
}

func validateValidatorRole(role ValidatorRole) error {
	switch role {
	case ValidatorRoleBlockProducer, ValidatorRoleVerifier, ValidatorRoleCollator, ValidatorRoleEvidenceReviewer:
		return nil
	default:
		return fmt.Errorf("unsupported validator role %q", role)
	}
}

func validatePosLayer(layer PosLayer) error {
	switch layer {
	case PosLayerEconomicConsensus, PosLayerTaskAssignment, PosLayerValidatorExecution, PosLayerStakingCapital, PosLayerBaseCometBFT:
		return nil
	default:
		return fmt.Errorf("unsupported pos layer %q", layer)
	}
}

func validateValidatorRoles(roles []ValidatorRole) error {
	seen := make(map[ValidatorRole]struct{}, len(roles))
	for _, role := range roles {
		if err := validateValidatorRole(role); err != nil {
			return err
		}
		if _, found := seen[role]; found {
			return fmt.Errorf("duplicate validator role %q", role)
		}
		seen[role] = struct{}{}
	}
	return nil
}

func normalizedRoles(roles []ValidatorRole, defaults []ValidatorRole) []ValidatorRole {
	if len(roles) == 0 {
		roles = defaults
	}
	out := make([]ValidatorRole, len(roles))
	copy(out, roles)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

func normalizeWorkloadTask(params Params, task WorkloadTask) WorkloadTask {
	out := task
	if out.WorkloadID == "" {
		out.WorkloadID = out.TaskID
	}
	if out.WorkloadType == "" {
		out.WorkloadType = WorkloadTypeServiceValidation
	}
	if out.WorkloadClass == "" {
		out.WorkloadClass = DefaultWorkloadClass
	}
	if out.RequiredValidators == 0 {
		out.RequiredValidators = params.MinTaskGroupValidators
	}
	if len(out.Roles) == 0 {
		out.Roles = DefaultTaskRoles()
	}
	out.Roles = normalizedRoles(out.Roles, DefaultTaskRoles())
	return out
}

func validateWorkloadType(workloadType WorkloadType) error {
	switch workloadType {
	case WorkloadTypeGlobalConsensus,
		WorkloadTypeZoneExecution,
		WorkloadTypeShardExecution,
		WorkloadTypeProofVerification,
		WorkloadTypeEvidenceVerification,
		WorkloadTypeDataAvailability,
		WorkloadTypeServiceValidation:
		return nil
	default:
		return fmt.Errorf("unsupported workload type %q", workloadType)
	}
}

func validatorsForRole(validators []ScoredValidator, role ValidatorRole) []ScoredValidator {
	out := make([]ScoredValidator, 0, len(validators))
	for _, validator := range validators {
		if ValidatorSupportsRole(validator.Candidate, role) {
			out = append(out, validator)
		}
	}
	return out
}

func selectTaskValidatorIDs(seed string, task WorkloadTask, role ValidatorRole, validators []ScoredValidator, required uint32) []string {
	type rankedValidator struct {
		validatorID string
		rankHash    string
		score       sdkmath.Int
	}
	ranked := make([]rankedValidator, len(validators))
	for i, validator := range validators {
		ranked[i] = rankedValidator{
			validatorID: validator.ValidatorID,
			score:       validator.Score,
			rankHash: posHashRoot("aetheris-pos-task-rank-v1", func(w posByteWriter) {
				posWritePart(w, seed)
				posWritePart(w, task.TaskID)
				posWritePart(w, task.WorkloadID)
				posWritePart(w, string(task.WorkloadType))
				posWritePart(w, task.ZoneID)
				posWritePart(w, task.ShardID)
				posWritePart(w, string(role))
				posWritePart(w, validator.ValidatorID)
				posWritePart(w, validator.Score.String())
				posWritePart(w, validator.VotingPowerNaet.String())
			}),
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].rankHash != ranked[j].rankHash {
			return ranked[i].rankHash < ranked[j].rankHash
		}
		if !ranked[i].score.Equal(ranked[j].score) {
			return ranked[i].score.GT(ranked[j].score)
		}
		return ranked[i].validatorID < ranked[j].validatorID
	})
	selected := make([]string, 0, required)
	for i := uint32(0); i < required; i++ {
		selected = append(selected, ranked[i].validatorID)
	}
	sort.Strings(selected)
	return selected
}

func taskKey(task WorkloadTask) string {
	return strings.Join([]string{task.TaskID, task.WorkloadID, string(task.WorkloadType), task.ZoneID, task.ShardID, task.WorkloadClass}, "|")
}

func taskGroupMembers(assignments []TaskAssignment) []string {
	seen := make(map[string]struct{})
	for _, assignment := range assignments {
		for _, validatorID := range assignment.Validators {
			seen[validatorID] = struct{}{}
		}
	}
	members := make([]string, 0, len(seen))
	for validatorID := range seen {
		members = append(members, validatorID)
	}
	sort.Strings(members)
	return members
}

func taskGroupVerifiers(assignments []TaskAssignment, members []string) []string {
	seen := make(map[string]struct{})
	for _, assignment := range assignments {
		if assignment.Role == ValidatorRoleVerifier || assignment.Role == ValidatorRoleEvidenceReviewer {
			for _, validatorID := range assignment.Validators {
				seen[validatorID] = struct{}{}
			}
		}
	}
	if len(seen) == 0 {
		for _, validatorID := range members {
			seen[validatorID] = struct{}{}
		}
	}
	verifiers := make([]string, 0, len(seen))
	for validatorID := range seen {
		verifiers = append(verifiers, validatorID)
	}
	sort.Strings(verifiers)
	return verifiers
}

func taskGroupProposerOrder(seed string, task WorkloadTask, members []string) []string {
	type proposerRank struct {
		validatorID string
		hash        string
	}
	ranks := make([]proposerRank, len(members))
	for i, validatorID := range members {
		ranks[i] = proposerRank{
			validatorID: validatorID,
			hash: posHashRoot("aetheris-pos-task-group-proposer-v1", func(w posByteWriter) {
				posWritePart(w, seed)
				posWritePart(w, task.TaskID)
				posWritePart(w, task.WorkloadID)
				posWritePart(w, string(task.WorkloadType))
				posWritePart(w, validatorID)
			}),
		}
	}
	sort.SliceStable(ranks, func(i, j int) bool {
		if ranks[i].hash != ranks[j].hash {
			return ranks[i].hash < ranks[j].hash
		}
		return ranks[i].validatorID < ranks[j].validatorID
	})
	out := make([]string, len(ranks))
	for i, rank := range ranks {
		out[i] = rank.validatorID
	}
	return out
}

func validateCanonicalValidatorIDs(fieldName string, validatorIDs []string) error {
	if len(validatorIDs) == 0 {
		return fmt.Errorf("%s ids are required", fieldName)
	}
	seen := make(map[string]struct{}, len(validatorIDs))
	var previous string
	for i, validatorID := range validatorIDs {
		if err := validatePosToken(fieldName+" id", validatorID); err != nil {
			return err
		}
		if _, found := seen[validatorID]; found {
			return fmt.Errorf("duplicate %s id %q", fieldName, validatorID)
		}
		seen[validatorID] = struct{}{}
		if i > 0 && previous >= validatorID {
			return fmt.Errorf("%s ids must be sorted canonically", fieldName)
		}
		previous = validatorID
	}
	return nil
}

func validateValidatorIDSet(fieldName string, values []string, expectedMembers []string) error {
	if len(values) != len(expectedMembers) {
		return fmt.Errorf("%s ids must include every task group member", fieldName)
	}
	return validateValidatorIDSubset(fieldName, values, expectedMembers)
}

func validateValidatorIDSubset(fieldName string, values []string, members []string) error {
	memberSet := make(map[string]struct{}, len(members))
	for _, member := range members {
		memberSet[member] = struct{}{}
	}
	for _, value := range values {
		if _, found := memberSet[value]; !found {
			return fmt.Errorf("%s id %q is not a task group member", fieldName, value)
		}
	}
	return nil
}

func compareWorkloadTasks(left, right WorkloadTask) int {
	if left.TaskID < right.TaskID {
		return -1
	}
	if left.TaskID > right.TaskID {
		return 1
	}
	if left.WorkloadID < right.WorkloadID {
		return -1
	}
	if left.WorkloadID > right.WorkloadID {
		return 1
	}
	if left.WorkloadType < right.WorkloadType {
		return -1
	}
	if left.WorkloadType > right.WorkloadType {
		return 1
	}
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.ShardID < right.ShardID {
		return -1
	}
	if left.ShardID > right.ShardID {
		return 1
	}
	if left.WorkloadClass < right.WorkloadClass {
		return -1
	}
	if left.WorkloadClass > right.WorkloadClass {
		return 1
	}
	return 0
}

func compareTaskAssignments(left, right TaskAssignment) int {
	if cmp := compareWorkloadTasks(
		WorkloadTask{TaskID: left.TaskID, WorkloadID: left.WorkloadID, WorkloadType: left.WorkloadType, ZoneID: left.ZoneID, ShardID: left.ShardID, WorkloadClass: left.WorkloadClass},
		WorkloadTask{TaskID: right.TaskID, WorkloadID: right.WorkloadID, WorkloadType: right.WorkloadType, ZoneID: right.ZoneID, ShardID: right.ShardID, WorkloadClass: right.WorkloadClass},
	); cmp != 0 {
		return cmp
	}
	if left.Role < right.Role {
		return -1
	}
	if left.Role > right.Role {
		return 1
	}
	return 0
}

func compareTaskGroups(left, right TaskGroup) int {
	if left.EpochID < right.EpochID {
		return -1
	}
	if left.EpochID > right.EpochID {
		return 1
	}
	if left.WorkloadID < right.WorkloadID {
		return -1
	}
	if left.WorkloadID > right.WorkloadID {
		return 1
	}
	if left.WorkloadType < right.WorkloadType {
		return -1
	}
	if left.WorkloadType > right.WorkloadType {
		return 1
	}
	if left.TaskGroupID < right.TaskGroupID {
		return -1
	}
	if left.TaskGroupID > right.TaskGroupID {
		return 1
	}
	return 0
}

func compareDelegationIntents(left, right DelegationIntent) int {
	if left.ValidatorID < right.ValidatorID {
		return -1
	}
	if left.ValidatorID > right.ValidatorID {
		return 1
	}
	if left.NominatorID < right.NominatorID {
		return -1
	}
	if left.NominatorID > right.NominatorID {
		return 1
	}
	if left.RequestedEpoch < right.RequestedEpoch {
		return -1
	}
	if left.RequestedEpoch > right.RequestedEpoch {
		return 1
	}
	return 0
}

func computeDelegationActivationKey(epoch uint64, validatorID string, nominations []Nomination) string {
	return posHashRoot("aetheris-pos-delegation-activation-v1", func(w posByteWriter) {
		posWriteUint64(w, epoch)
		posWritePart(w, validatorID)
		posWriteUint64(w, uint64(len(nominations)))
		for _, nomination := range nominations {
			posWritePart(w, nomination.NominatorID)
			posWritePart(w, nomination.StakeNaet.String())
		}
	})
}

func computeEvidenceSettlementHash(settlement EvidenceSettlement) string {
	return posHashRoot("aetheris-pos-evidence-settlement-v1", func(w posByteWriter) {
		posWritePart(w, settlement.EvidenceID)
		posWritePart(w, settlement.ReporterID)
		posWritePart(w, settlement.Slash.ValidatorID)
		posWritePart(w, settlement.Slash.Misbehavior)
		posWritePart(w, settlement.Slash.TotalSlashedNaet.String())
		posWritePart(w, settlement.ReporterRewardNaet.String())
		posWritePart(w, settlement.BurnNaet.String())
		posWriteUint64(w, uint64(settlement.Slash.EvidenceHeight))
	})
}

func computeWorkloadRewardRoot(settlement WorkloadRewardSettlement) string {
	return posHashRoot("aetheris-pos-workload-reward-root-v1", func(w posByteWriter) {
		posWriteUint64(w, settlement.EpochID)
		posWriteUint64(w, uint64(len(settlement.Rewards)))
		for _, reward := range settlement.Rewards {
			posWritePart(w, reward.ValidatorID)
			posWritePart(w, reward.RewardNaet.String())
			posWriteUint64(w, reward.WorkUnits)
		}
		posWritePart(w, settlement.RemainderNaet.String())
		posWriteUint64(w, settlement.CompletedUnits)
	})
}

func validateRoleRewardWeights(weights []RoleRewardWeight) error {
	if len(weights) == 0 {
		return errors.New("role reward weights are required")
	}
	total := uint64(0)
	seen := make(map[ValidatorRole]struct{}, len(weights))
	for _, weight := range weights {
		if err := validateValidatorRole(weight.Role); err != nil {
			return err
		}
		if _, found := seen[weight.Role]; found {
			return fmt.Errorf("duplicate role reward weight %q", weight.Role)
		}
		seen[weight.Role] = struct{}{}
		total += uint64(weight.WeightBps)
	}
	if total != uint64(BasisPoints) {
		return fmt.Errorf("role reward weights must sum to %d bps", BasisPoints)
	}
	return nil
}

func sortedStringKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func mergeNominations(existing []Nomination, activated []Nomination) []Nomination {
	byNominator := make(map[string]sdkmath.Int, len(existing)+len(activated))
	for _, nomination := range existing {
		current, found := byNominator[nomination.NominatorID]
		if !found {
			current = sdkmath.ZeroInt()
		}
		byNominator[nomination.NominatorID] = current.Add(nomination.StakeNaet)
	}
	for _, nomination := range activated {
		current, found := byNominator[nomination.NominatorID]
		if !found {
			current = sdkmath.ZeroInt()
		}
		byNominator[nomination.NominatorID] = current.Add(nomination.StakeNaet)
	}
	nominatorIDs := sortedStringKeys(byNominator)
	out := make([]Nomination, 0, len(nominatorIDs))
	for _, nominatorID := range nominatorIDs {
		out = append(out, Nomination{NominatorID: nominatorID, StakeNaet: byNominator[nominatorID]})
	}
	return out
}

func ratioBps(numerator uint64, denominator uint64) uint32 {
	if denominator == 0 {
		return BasisPoints
	}
	if numerator >= denominator {
		return BasisPoints
	}
	return uint32((uint64(BasisPoints) * numerator) / denominator)
}

func validatePosToken(fieldName string, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxPosTokenLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxPosTokenLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func validatePosResponsibility(fieldName string, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > maxPosTokenLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, maxPosTokenLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' || r == ' ' || r == '+' || r == ',' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func validatePosHash(fieldName string, value string) error {
	if len(value) != PosHashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, PosHashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, PosHashHexLength)
	}
	return nil
}

type posByteWriter interface {
	Write([]byte) (int, error)
}

func posHashRoot(domain string, write func(posByteWriter)) string {
	h := sha256.New()
	posWritePart(h, domain)
	write(h)
	return hex.EncodeToString(h.Sum(nil))
}

func posWritePart(w posByteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func posWriteUint64(w posByteWriter, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}
