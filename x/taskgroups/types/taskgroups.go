package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const ModuleName = "taskgroups"

type WorkloadRegistration struct {
	WorkloadID		string
	WorkloadType		postypes.WorkloadType
	ZoneID			string
	ShardID			string
	WorkloadClass		string
	MinimumGroupSize	uint32
	RoleRequirements	[]postypes.ValidatorRole
	ExcludedValidators	[]string
}

type WorkloadRegistry struct {
	Entries []WorkloadRegistration
}

type AssignmentInput struct {
	Params			postypes.Params
	Epoch			postypes.EpochRecord
	ValidatorSet		[]postypes.ScoredValidator
	Registry		WorkloadRegistry
	ActivationHeight	uint64
	ExpiryHeight		uint64
}

type AssignmentState struct {
	Registry		WorkloadRegistry
	GroupSet		postypes.TaskGroupSet
	ValidatorSetHash	string
}

type AssignmentProof struct {
	EpochID			uint64
	WorkloadID		string
	TaskGroupID		string
	TaskGroupRoot		string
	StakeWeightRoot		string
	AssignmentSeed		string
	ValidatorMembers	[]string
	ProposerOrder		[]string
	VerifierSet		[]string
}

func NewWorkloadRegistry(params postypes.Params, entries []WorkloadRegistration) (WorkloadRegistry, error) {
	out := make([]WorkloadRegistration, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for i, entry := range entries {
		entry = normalizeWorkloadRegistration(params, entry)
		if err := entry.Validate(params); err != nil {
			return WorkloadRegistry{}, err
		}
		if _, found := seen[entry.WorkloadID]; found {
			return WorkloadRegistry{}, fmt.Errorf("duplicate workload registration %q", entry.WorkloadID)
		}
		seen[entry.WorkloadID] = struct{}{}
		out[i] = entry
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].WorkloadID < out[j].WorkloadID
	})
	return WorkloadRegistry{Entries: out}, nil
}

func (r WorkloadRegistration) Validate(params postypes.Params) error {
	task := r.ToWorkloadTask()
	if err := task.Validate(params); err != nil {
		return err
	}
	if r.MinimumGroupSize == 0 {
		return errors.New("workload minimum group size is required")
	}
	return nil
}

func (r WorkloadRegistration) ToWorkloadTask() postypes.WorkloadTask {
	return postypes.WorkloadTask{
		TaskID:			r.WorkloadID,
		WorkloadID:		r.WorkloadID,
		WorkloadType:		r.WorkloadType,
		ZoneID:			r.ZoneID,
		ShardID:		r.ShardID,
		WorkloadClass:		r.WorkloadClass,
		RequiredValidators:	r.MinimumGroupSize,
		Roles:			cloneRoles(r.RoleRequirements),
		ExcludedValidators:	cloneStrings(r.ExcludedValidators),
	}
}

func (r WorkloadRegistry) ToWorkloadTasks() []postypes.WorkloadTask {
	tasks := make([]postypes.WorkloadTask, len(r.Entries))
	for i, entry := range r.Entries {
		tasks[i] = entry.ToWorkloadTask()
	}
	return tasks
}

func DeterministicAssignment(input AssignmentInput) (AssignmentState, error) {
	registry, err := NewWorkloadRegistry(input.Params, input.Registry.Entries)
	if err != nil {
		return AssignmentState{}, err
	}
	validatorSetHash, err := postypes.ComputeValidatorSetHash(input.ValidatorSet)
	if err != nil {
		return AssignmentState{}, err
	}
	groupSet, err := postypes.BuildTaskGroups(input.Params, input.Epoch, input.ValidatorSet, registry.ToWorkloadTasks(), input.ActivationHeight, input.ExpiryHeight)
	if err != nil {
		return AssignmentState{}, err
	}
	return AssignmentState{
		Registry:		registry,
		GroupSet:		groupSet,
		ValidatorSetHash:	validatorSetHash,
	}, nil
}

func (s AssignmentState) Validate() error {
	if err := s.GroupSet.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(s.ValidatorSetHash) == "" {
		return errors.New("assignment state validator set hash is required")
	}
	workloadIDs := make(map[string]struct{}, len(s.Registry.Entries))
	for _, entry := range s.Registry.Entries {
		workloadIDs[entry.WorkloadID] = struct{}{}
	}
	for _, group := range s.GroupSet.Groups {
		if _, found := workloadIDs[group.WorkloadID]; !found {
			return fmt.Errorf("task group %q is not registered", group.WorkloadID)
		}
	}
	return nil
}

func (s AssignmentState) QueryAssignmentProof(workloadID string) (AssignmentProof, bool, error) {
	workloadID = strings.TrimSpace(workloadID)
	for _, group := range s.GroupSet.Groups {
		if group.WorkloadID != workloadID {
			continue
		}
		if err := group.Validate(); err != nil {
			return AssignmentProof{}, false, err
		}
		if err := s.GroupSet.Validate(); err != nil {
			return AssignmentProof{}, false, err
		}
		return AssignmentProof{
			EpochID:		group.EpochID,
			WorkloadID:		group.WorkloadID,
			TaskGroupID:		group.TaskGroupID,
			TaskGroupRoot:		s.GroupSet.Root,
			StakeWeightRoot:	group.StakeWeightRoot,
			AssignmentSeed:		group.AssignmentSeed,
			ValidatorMembers:	cloneStrings(group.ValidatorMembers),
			ProposerOrder:		cloneStrings(group.ProposerOrder),
			VerifierSet:		cloneStrings(group.VerifierSet),
		}, true, nil
	}
	return AssignmentProof{}, false, nil
}

func normalizeWorkloadRegistration(params postypes.Params, entry WorkloadRegistration) WorkloadRegistration {
	entry.WorkloadID = strings.TrimSpace(entry.WorkloadID)
	entry.ZoneID = strings.TrimSpace(entry.ZoneID)
	entry.ShardID = strings.TrimSpace(entry.ShardID)
	entry.WorkloadClass = strings.TrimSpace(entry.WorkloadClass)
	if entry.WorkloadClass == "" {
		entry.WorkloadClass = postypes.DefaultWorkloadClass
	}
	if entry.MinimumGroupSize == 0 {
		entry.MinimumGroupSize = params.MinTaskGroupValidators
	}
	if len(entry.RoleRequirements) == 0 {
		entry.RoleRequirements = postypes.DefaultTaskRoles()
	}
	entry.ExcludedValidators = cloneStrings(entry.ExcludedValidators)
	sort.Strings(entry.ExcludedValidators)
	return entry
}

func cloneRoles(values []postypes.ValidatorRole) []postypes.ValidatorRole {
	out := make([]postypes.ValidatorRole, len(values))
	copy(out, values)
	return out
}

func cloneStrings(values []string) []string {
	out := make([]string, len(values))
	copy(out, values)
	return out
}
