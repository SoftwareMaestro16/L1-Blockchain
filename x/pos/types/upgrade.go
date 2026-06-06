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

type WorkloadTask struct {
	TaskID             string
	ZoneID             string
	ShardID            string
	WorkloadClass      string
	RequiredValidators uint32
	Roles              []ValidatorRole
}

type TaskAssignment struct {
	TaskID         string
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
	seed, err := DeriveEpochSeed(epochID, startHeight, previousSeed, validatorSetHash)
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

func (t WorkloadTask) Validate(params Params) error {
	if err := validatePosToken("task id", t.TaskID); err != nil {
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

func ComputeTaskAssignmentHash(epochID uint64, seed string, assignment TaskAssignment) string {
	return posHashRoot("aetheris-pos-task-assignment-v1", func(w posByteWriter) {
		posWriteUint64(w, epochID)
		posWritePart(w, seed)
		posWritePart(w, assignment.TaskID)
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
	if currentEpoch < e.EvidenceEpoch {
		return errors.New("current epoch cannot be before evidence epoch")
	}
	if currentEpoch-e.EvidenceEpoch > params.EvidenceWindowEpochs {
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

func compareWorkloadTasks(left, right WorkloadTask) int {
	if left.TaskID < right.TaskID {
		return -1
	}
	if left.TaskID > right.TaskID {
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
	return 0
}

func compareTaskAssignments(left, right TaskAssignment) int {
	if cmp := compareWorkloadTasks(
		WorkloadTask{TaskID: left.TaskID, ZoneID: left.ZoneID, ShardID: left.ShardID},
		WorkloadTask{TaskID: right.TaskID, ZoneID: right.ZoneID, ShardID: right.ShardID},
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
