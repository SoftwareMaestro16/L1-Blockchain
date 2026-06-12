package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const PaymentRoutingLogicTaskSpecVersion = uint64(1)

type PaymentRoutingLogicTaskID string

const (
	PaymentRoutingTaskRouteTableState		PaymentRoutingLogicTaskID	= "route_table_state"
	PaymentRoutingTaskEpochUpdateRules		PaymentRoutingLogicTaskID	= "routing_epoch_update_rules"
	PaymentRoutingTaskDeterministicPathScoring	PaymentRoutingLogicTaskID	= "deterministic_path_scoring"
	PaymentRoutingTaskDeliveryScheduler		PaymentRoutingLogicTaskID	= "message_delivery_scheduler"
	PaymentRoutingTaskReceiptCommitmentQuery	PaymentRoutingLogicTaskID	= "receipt_commitment_query"
	PaymentRoutingTaskCongestionEpochTests		PaymentRoutingLogicTaskID	= "congestion_epoch_tests"
	PaymentRoutingTaskRetryExpiryTests		PaymentRoutingLogicTaskID	= "retry_expiry_tests"
	PaymentRoutingTaskBounceConservationTests	PaymentRoutingLogicTaskID	= "bounce_value_conservation_tests"
)

type PaymentRoutingLogicTaskDescriptor struct {
	TaskID		PaymentRoutingLogicTaskID
	Task		string
	Target		string
	Enforcement	string
	Evidence	string
	DescriptorHash	string
}

type PaymentRoutingLogicTaskSpec struct {
	Version	uint64
	Tasks	[]PaymentRoutingLogicTaskDescriptor
	Root	string
}

func PaymentRoutingLogicTaskDescriptors() []PaymentRoutingLogicTaskDescriptor {
	return []PaymentRoutingLogicTaskDescriptor{
		paymentRoutingLogicTask(
			PaymentRoutingTaskRouteTableState,
			"Define route table state.",
			"PaymentRouteTableState",
			"Routes, scheduled delivery tasks, receipts, epoch, and root hash are normalized into one canonical table root.",
			"PaymentRouteTableState;ComputePaymentRouteTableRoot",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskEpochUpdateRules,
			"Define routing epoch update rules.",
			"ApplyPaymentRoutingEpochUpdate",
			"Epochs must strictly increase, use a positive committed height, reject expired routes, and verify the previous root.",
			"PaymentRoutingEpochUpdate;ApplyPaymentRoutingEpochUpdate",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskDeterministicPathScoring,
			"Implement deterministic path scoring.",
			"DeterministicPaymentRouteScore",
			"Scores use route fees, hop order, and committed congestion snapshots with deterministic filtering and no live mempool inputs.",
			"DeterministicPaymentRouteScore;PaymentRouteCongestionSnapshot",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskDeliveryScheduler,
			"Implement message delivery scheduler.",
			"SchedulePaymentRouteDelivery",
			"Delivery tasks are derived from route hops, retry policy, attempts, delivery height, expiry, and deterministic task hashes.",
			"SchedulePaymentRouteDelivery;PaymentRouteDeliveryTask",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskReceiptCommitmentQuery,
			"Implement receipt commitment and query.",
			"RecordPaymentRouteReceipt",
			"Route receipts are hash-validated, inserted or replaced by route and attempt, committed to table root, and queryable by route ID.",
			"PaymentRouteReceipt;RecordPaymentRouteReceipt;QueryPaymentRouteReceipt",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskCongestionEpochTests,
			"Add tests for congestion changes between epochs.",
			"TestPaymentRouteScoringChangesDeterministicallyBetweenEpochs",
			"Committed congestion input changes score and epoch root deterministically while preserving replayable route contents.",
			"TestPaymentRouteScoringChangesDeterministicallyBetweenEpochs",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskRetryExpiryTests,
			"Add tests for retry and expiry behavior.",
			"TestPaymentRouteSchedulerRetriesAndExpires",
			"Scheduler tests cover initial delivery, delayed retry, exhausted retry bounce, expired route receipt, and receipt query.",
			"TestPaymentRouteSchedulerRetriesAndExpires",
		),
		paymentRoutingLogicTask(
			PaymentRoutingTaskBounceConservationTests,
			"Add tests for bounce value conservation.",
			"TestPaymentRouteBounceValueConservation",
			"Bounce receipts return at most the original route amount plus max fee and reject leaking or value-creating receipts.",
			"ValidatePaymentRouteBounceConservation;TestPaymentRouteBounceValueConservation",
		),
	}
}

func DefaultPaymentRoutingLogicTaskSpec() (PaymentRoutingLogicTaskSpec, error) {
	return BuildPaymentRoutingLogicTaskSpec(PaymentRoutingLogicTaskDescriptors())
}

func BuildPaymentRoutingLogicTaskSpec(tasks []PaymentRoutingLogicTaskDescriptor) (PaymentRoutingLogicTaskSpec, error) {
	spec := PaymentRoutingLogicTaskSpec{
		Version:	PaymentRoutingLogicTaskSpecVersion,
		Tasks:		normalizePaymentRoutingLogicTaskDescriptors(tasks),
	}
	if err := spec.ValidateFormat(); err != nil {
		return PaymentRoutingLogicTaskSpec{}, err
	}
	spec.Root = ComputePaymentRoutingLogicTaskSpecRoot(spec.Tasks)
	return spec, spec.Validate()
}

func (s PaymentRoutingLogicTaskSpec) Normalize() PaymentRoutingLogicTaskSpec {
	if s.Version == 0 {
		s.Version = PaymentRoutingLogicTaskSpecVersion
	}
	s.Tasks = normalizePaymentRoutingLogicTaskDescriptors(s.Tasks)
	s.Root = normalizeOptionalHash(s.Root)
	return s
}

func (s PaymentRoutingLogicTaskSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != PaymentRoutingLogicTaskSpecVersion {
		return fmt.Errorf("payments routing task spec version must be %d", PaymentRoutingLogicTaskSpecVersion)
	}
	if len(s.Tasks) == 0 {
		return errors.New("payments routing task spec requires tasks")
	}
	seen := make(map[PaymentRoutingLogicTaskID]struct{}, len(s.Tasks))
	var previous PaymentRoutingLogicTaskID
	for i, task := range s.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		if _, found := seen[task.TaskID]; found {
			return fmt.Errorf("duplicate payments routing task %s", task.TaskID)
		}
		seen[task.TaskID] = struct{}{}
		if i > 0 && previous >= task.TaskID {
			return errors.New("payments routing tasks must be sorted canonically")
		}
		previous = task.TaskID
	}
	if s.Root != "" {
		if err := ValidateHash("payments routing task spec root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s PaymentRoutingLogicTaskSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("payments routing task spec root is required")
	}
	expected := ComputePaymentRoutingLogicTaskSpecRoot(s.Tasks)
	if s.Root != expected {
		return fmt.Errorf("payments routing task spec root mismatch: expected %s", expected)
	}
	return nil
}

func BuildPaymentRoutingLogicTaskDescriptor(desc PaymentRoutingLogicTaskDescriptor) (PaymentRoutingLogicTaskDescriptor, error) {
	desc = desc.Normalize()
	if desc.DescriptorHash != "" {
		return PaymentRoutingLogicTaskDescriptor{}, errors.New("payments routing task descriptor hash must be empty before construction")
	}
	if err := desc.ValidateFormat(); err != nil {
		return PaymentRoutingLogicTaskDescriptor{}, err
	}
	desc.DescriptorHash = ComputePaymentRoutingLogicTaskDescriptorHash(desc)
	return desc, desc.Validate()
}

func (d PaymentRoutingLogicTaskDescriptor) Normalize() PaymentRoutingLogicTaskDescriptor {
	d.Task = compactPaymentRoutingText(d.Task)
	d.Target = compactPaymentRoutingText(d.Target)
	d.Enforcement = compactPaymentRoutingText(d.Enforcement)
	d.Evidence = compactPaymentRoutingText(d.Evidence)
	d.DescriptorHash = normalizeOptionalHash(d.DescriptorHash)
	return d
}

func (d PaymentRoutingLogicTaskDescriptor) ValidateFormat() error {
	d = d.Normalize()
	if !IsPaymentRoutingLogicTaskID(d.TaskID) {
		return fmt.Errorf("unknown payments routing task %q", d.TaskID)
	}
	if d.Task == "" {
		return errors.New("payments routing task text is required")
	}
	if d.Target == "" {
		return errors.New("payments routing task target is required")
	}
	if d.Enforcement == "" {
		return errors.New("payments routing task enforcement is required")
	}
	if d.Evidence == "" {
		return errors.New("payments routing task evidence is required")
	}
	if d.DescriptorHash != "" {
		if err := ValidateHash("payments routing task descriptor hash", d.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (d PaymentRoutingLogicTaskDescriptor) Validate() error {
	d = d.Normalize()
	if err := d.ValidateFormat(); err != nil {
		return err
	}
	if d.DescriptorHash == "" {
		return errors.New("payments routing task descriptor hash is required")
	}
	expected := ComputePaymentRoutingLogicTaskDescriptorHash(d)
	if d.DescriptorHash != expected {
		return fmt.Errorf("payments routing task descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidatePaymentRoutingLogicTaskSpec() error {
	spec, err := DefaultPaymentRoutingLogicTaskSpec()
	if err != nil {
		return err
	}
	required := []PaymentRoutingLogicTaskID{
		PaymentRoutingTaskRouteTableState,
		PaymentRoutingTaskEpochUpdateRules,
		PaymentRoutingTaskDeterministicPathScoring,
		PaymentRoutingTaskDeliveryScheduler,
		PaymentRoutingTaskReceiptCommitmentQuery,
		PaymentRoutingTaskCongestionEpochTests,
		PaymentRoutingTaskRetryExpiryTests,
		PaymentRoutingTaskBounceConservationTests,
	}
	seen := make(map[PaymentRoutingLogicTaskID]struct{}, len(spec.Tasks))
	for _, task := range spec.Tasks {
		seen[task.TaskID] = struct{}{}
	}
	for _, taskID := range required {
		if _, found := seen[taskID]; !found {
			return fmt.Errorf("payments routing task spec missing %s", taskID)
		}
	}
	return nil
}

func IsPaymentRoutingLogicTaskID(taskID PaymentRoutingLogicTaskID) bool {
	switch taskID {
	case PaymentRoutingTaskRouteTableState,
		PaymentRoutingTaskEpochUpdateRules,
		PaymentRoutingTaskDeterministicPathScoring,
		PaymentRoutingTaskDeliveryScheduler,
		PaymentRoutingTaskReceiptCommitmentQuery,
		PaymentRoutingTaskCongestionEpochTests,
		PaymentRoutingTaskRetryExpiryTests,
		PaymentRoutingTaskBounceConservationTests:
		return true
	default:
		return false
	}
}

func ComputePaymentRoutingLogicTaskDescriptorHash(desc PaymentRoutingLogicTaskDescriptor) string {
	desc = desc.Normalize()
	return HashParts("payments-routing-logic-task-descriptor", string(desc.TaskID), desc.Task, desc.Target, desc.Enforcement, desc.Evidence)
}

func ComputePaymentRoutingLogicTaskSpecRoot(tasks []PaymentRoutingLogicTaskDescriptor) string {
	ordered := normalizePaymentRoutingLogicTaskDescriptors(tasks)
	parts := []string{"payments-routing-logic-task-spec", fmt.Sprintf("%020d", PaymentRoutingLogicTaskSpecVersion)}
	for _, task := range ordered {
		parts = append(parts, string(task.TaskID), task.DescriptorHash)
	}
	return HashParts(parts...)
}

func paymentRoutingLogicTask(taskID PaymentRoutingLogicTaskID, task string, target string, enforcement string, evidence string) PaymentRoutingLogicTaskDescriptor {
	desc, err := BuildPaymentRoutingLogicTaskDescriptor(PaymentRoutingLogicTaskDescriptor{
		TaskID:		taskID,
		Task:		task,
		Target:		target,
		Enforcement:	enforcement,
		Evidence:	evidence,
	})
	if err != nil {
		panic(err)
	}
	return desc
}

func normalizePaymentRoutingLogicTaskDescriptors(values []PaymentRoutingLogicTaskDescriptor) []PaymentRoutingLogicTaskDescriptor {
	out := make([]PaymentRoutingLogicTaskDescriptor, len(values))
	for i, value := range values {
		normalized := value.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputePaymentRoutingLogicTaskDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].TaskID < out[j].TaskID
	})
	return out
}

func compactPaymentRoutingText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
