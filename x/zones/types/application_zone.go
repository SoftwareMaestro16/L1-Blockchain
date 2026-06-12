package types

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

const (
	ApplicationZonePrefix		= "apps"
	ApplicationAppPrefix		= ApplicationZonePrefix + "/app"
	ApplicationWorkflowPrefix	= ApplicationZonePrefix + "/workflow"
	ApplicationSchedulerPrefix	= ApplicationZonePrefix + "/scheduler"
	ApplicationAutomationPrefix	= ApplicationZonePrefix + "/automation"
	ApplicationPermissionsPrefix	= ApplicationZonePrefix + "/permissions"
	ApplicationReceiptPrefix	= ApplicationZonePrefix + "/receipts"
	ApplicationOutboxPrefix		= ApplicationZonePrefix + "/outbox"
)

type ApplicationRuntime interface {
	RuntimeID() string
	CreateApplication(context.Context, ApplicationRecord) (ApplicationExecutionReceipt, error)
	UpdateApplication(context.Context, ApplicationRecord) (ApplicationExecutionReceipt, error)
	StartWorkflow(context.Context, ApplicationWorkflowState) (ApplicationExecutionReceipt, error)
	AdvanceWorkflow(context.Context, ApplicationWorkflowState) (ApplicationExecutionReceipt, error)
	ExecuteAutomation(context.Context, ApplicationAutomation) (ApplicationExecutionReceipt, []ZoneMessage, error)
	ExecuteApplicationTask(context.Context, ApplicationScheduledTask) (ApplicationExecutionReceipt, []ZoneMessage, error)
	ApplyApplicationMessage(context.Context, ZoneMessage) (ApplicationExecutionReceipt, []ZoneMessage, error)
	ComputeApplicationStateRoot(context.Context) (string, error)
}

type ApplicationWorkflowStatus string
type ApplicationTaskStatus string
type ApplicationProofKind string
type ApplicationMessageKind string
type ApplicationPermissionScope string
type ApplicationShardRoutingMode string
type ApplicationProofRootType string

const (
	ApplicationWorkflowPending	ApplicationWorkflowStatus	= "PENDING"
	ApplicationWorkflowRunning	ApplicationWorkflowStatus	= "RUNNING"
	ApplicationWorkflowSucceeded	ApplicationWorkflowStatus	= "SUCCEEDED"
	ApplicationWorkflowFailed	ApplicationWorkflowStatus	= "FAILED"
	ApplicationWorkflowCanceled	ApplicationWorkflowStatus	= "CANCELED"

	ApplicationTaskPending	ApplicationTaskStatus	= "PENDING"
	ApplicationTaskExecuted	ApplicationTaskStatus	= "EXECUTED"
	ApplicationTaskFailed	ApplicationTaskStatus	= "FAILED"
	ApplicationTaskDeferred	ApplicationTaskStatus	= "DEFERRED"
	ApplicationTaskSkipped	ApplicationTaskStatus	= "SKIPPED"
	ApplicationTaskCanceled	ApplicationTaskStatus	= "CANCELED"

	ApplicationMessageCreateApp		ApplicationMessageKind	= "MsgCreateApp"
	ApplicationMessageUpdateApp		ApplicationMessageKind	= "MsgUpdateApp"
	ApplicationMessageStartWorkflow		ApplicationMessageKind	= "MsgStartWorkflow"
	ApplicationMessageAdvanceWorkflow	ApplicationMessageKind	= "MsgAdvanceWorkflow"
	ApplicationMessageScheduleTask		ApplicationMessageKind	= "MsgScheduleTask"
	ApplicationMessageCancelTask		ApplicationMessageKind	= "MsgCancelTask"
	ApplicationMessageExecuteAutomation	ApplicationMessageKind	= "MsgExecuteAutomation"

	ApplicationProofApp		ApplicationProofKind	= "QueryApp"
	ApplicationProofWorkflow	ApplicationProofKind	= "QueryWorkflow"
	ApplicationProofScheduledTask	ApplicationProofKind	= "QueryScheduledTask"
	ApplicationProofAutomation	ApplicationProofKind	= "QueryAutomation"
	ApplicationProofAppReceipts	ApplicationProofKind	= "QueryAppReceipts"
	ApplicationProofAppQueue	ApplicationProofKind	= "QueryAppQueue"
	ApplicationProofPermissions	ApplicationProofKind	= "QueryServicePermissions"

	ApplicationPermissionAdmin	ApplicationPermissionScope	= "admin"
	ApplicationPermissionExecute	ApplicationPermissionScope	= "execute"
	ApplicationPermissionSchedule	ApplicationPermissionScope	= "schedule"

	ApplicationRouteAppID		ApplicationShardRoutingMode	= "app_id"
	ApplicationRouteWorkflowID	ApplicationShardRoutingMode	= "workflow_id"
	ApplicationRouteSchedulerBucket	ApplicationShardRoutingMode	= "scheduler_bucket"

	ApplicationProofRootApp		ApplicationProofRootType	= "app"
	ApplicationProofRootWorkflow	ApplicationProofRootType	= "workflow"
	ApplicationProofRootScheduler	ApplicationProofRootType	= "scheduler"
	ApplicationProofRootQueue	ApplicationProofRootType	= "app_queue"
	ApplicationProofRootPermission	ApplicationProofRootType	= "service_permission"
)

type ApplicationZoneBoundary struct {
	ZoneID		ZoneID
	OwnsPrefixes	[]string
	Messages	[]ApplicationMessageKind
	ProofKinds	[]ApplicationProofKind
}

type ApplicationRecord struct {
	AppID		string
	Owner		string
	RuntimeID	string
	Version		uint64
	Enabled		bool
	ConfigHash	string
	UpdatedHeight	uint64
}

type ApplicationWorkflowState struct {
	WorkflowID	string
	AppID		string
	Owner		string
	Status		ApplicationWorkflowStatus
	CurrentStep	uint32
	TotalSteps	uint32
	PayloadHash	string
	UpdatedHeight	uint64
}

type ApplicationScheduledTask struct {
	Bucket		string
	TaskID		string
	WorkflowID	string
	AppID		string
	ScheduledHeight	uint64
	Priority	uint32
	Sequence	uint64
	GasLimit	uint64
	PayloadHash	string
	Status		ApplicationTaskStatus
}

type ApplicationAutomation struct {
	AutomationID	string
	AppID		string
	WorkflowID	string
	Enabled		bool
	TriggerHash	string
	NextRunHeight	uint64
	UpdatedHeight	uint64
}

type ApplicationPermission struct {
	AppID		string
	Address		string
	Scope		ApplicationPermissionScope
	ExpiresHeight	uint64
	GrantHash	string
}

type ApplicationSchedulerQueue struct {
	ZoneID		ZoneID
	Height		uint64
	MaxWorkPerBlock	uint32
	Tasks		[]ApplicationScheduledTask
}

type ApplicationWorkLimit struct {
	MaxTasksPerBlock	uint32
	MaxMessagesPerBlock	uint32
	MaxGasPerBlock		uint64
}

type ApplicationRuntimeBoundary struct {
	ZoneID				ZoneID
	RuntimeID			string
	AllowedStatePrefixes		[]string
	CrossZoneEffectMechanism	string
	BoundaryHash			string
}

type ApplicationShardRoute struct {
	ZoneID		ZoneID
	LayoutEpoch	uint64
	ShardCount	uint32
	ShardID		uint32
	RoutingMode	ApplicationShardRoutingMode
	RouteKey	string
	StateKey	string
	RouteHash	string
}

type ApplicationAsyncOutput struct {
	OutputID	string
	AppID		string
	WorkflowID	string
	DestinationZone	ZoneID
	Destination	string
	PayloadHash	string
	GasLimit	uint64
	RetryNonce	uint64
	CreatedHeight	uint64
	MessageHash	string
}

type ApplicationProofRootExport struct {
	ZoneID		ZoneID
	Height		uint64
	RootType	ApplicationProofRootType
	RootHash	string
	Source		string
}

type ApplicationExecutionReceipt struct {
	ZoneID		ZoneID
	Height		uint64
	ExecutionID	string
	TaskID		string
	WorkflowID	string
	AppID		string
	Status		ApplicationTaskStatus
	GasUsed		uint64
	OutputHash	string
	OutboxMessages	uint32
	Sequence	uint64
	ReceiptHash	string
}

type ApplicationZoneState struct {
	Height		uint64
	Apps		[]ApplicationRecord
	Workflows	[]ApplicationWorkflowState
	Tasks		[]ApplicationScheduledTask
	Automations	[]ApplicationAutomation
	Permissions	[]ApplicationPermission
	Receipts	[]ApplicationExecutionReceipt
}

type ApplicationZoneRoots struct {
	Height		uint64
	AppRoot		string
	WorkflowRoot	string
	SchedulerRoot	string
	AutomationRoot	string
	PermissionRoot	string
	ReceiptRoot	string
	QueueRoot	string
	InboxRoot	string
	OutboxRoot	string
	ExecutionRoot	string
	ProofRoot	string
	StateRoot	string
}

func DefaultApplicationZoneBoundary() ApplicationZoneBoundary {
	return ApplicationZoneBoundary{
		ZoneID:	ZoneIDApplication,
		OwnsPrefixes: []string{
			ApplicationAppPrefix,
			ApplicationAutomationPrefix,
			ApplicationOutboxPrefix,
			ApplicationPermissionsPrefix,
			ApplicationReceiptPrefix,
			ApplicationSchedulerPrefix,
			ApplicationWorkflowPrefix,
		},
		Messages: []ApplicationMessageKind{
			ApplicationMessageCreateApp,
			ApplicationMessageUpdateApp,
			ApplicationMessageStartWorkflow,
			ApplicationMessageAdvanceWorkflow,
			ApplicationMessageScheduleTask,
			ApplicationMessageCancelTask,
			ApplicationMessageExecuteAutomation,
		},
		ProofKinds: []ApplicationProofKind{
			ApplicationProofApp,
			ApplicationProofWorkflow,
			ApplicationProofScheduledTask,
			ApplicationProofAutomation,
			ApplicationProofAppReceipts,
			ApplicationProofAppQueue,
			ApplicationProofPermissions,
		},
	}
}

func DefaultApplicationWorkLimit() ApplicationWorkLimit {
	return ApplicationWorkLimit{
		MaxTasksPerBlock:	1024,
		MaxMessagesPerBlock:	4096,
		MaxGasPerBlock:		50_000_000,
	}
}

func (b ApplicationZoneBoundary) Validate() error {
	if b.ZoneID != ZoneIDApplication {
		return errors.New("application zone boundary must use APPLICATION_ZONE")
	}
	if len(b.OwnsPrefixes) == 0 || len(b.Messages) == 0 || len(b.ProofKinds) == 0 {
		return errors.New("application zone boundary requires prefixes, messages, and proof kinds")
	}
	for i, prefix := range b.OwnsPrefixes {
		if err := validateRuntimeToken("application zone prefix", prefix, MaxZoneNamespaceLength); err != nil {
			return err
		}
		if i > 0 && b.OwnsPrefixes[i-1] >= prefix {
			return errors.New("application zone prefixes must be sorted canonically")
		}
	}
	for _, msg := range b.Messages {
		if !IsApplicationMessageKind(msg) {
			return fmt.Errorf("unknown application message kind %q", msg)
		}
	}
	for _, kind := range b.ProofKinds {
		if !IsApplicationProofKind(kind) {
			return fmt.Errorf("unknown application proof kind %q", kind)
		}
	}
	return nil
}

func ApplicationAppKey(appID string) (string, error) {
	if err := validateRuntimeToken("application app id", appID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ApplicationAppPrefix + "/" + appID, nil
}

func ApplicationWorkflowKey(workflowID string) (string, error) {
	if err := validateRuntimeToken("application workflow id", workflowID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ApplicationWorkflowPrefix + "/" + workflowID, nil
}

func ApplicationSchedulerTaskKey(task ApplicationScheduledTask) (string, error) {
	if err := task.Validate(); err != nil {
		return "", err
	}
	return ApplicationSchedulerPrefix + "/" + task.Bucket + "/" + task.TaskID, nil
}

func ApplicationAutomationKey(automationID string) (string, error) {
	if err := validateRuntimeToken("application automation id", automationID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ApplicationAutomationPrefix + "/" + automationID, nil
}

func ApplicationPermissionKey(appID, address string) (string, error) {
	if err := validateRuntimeToken("application permission app id", appID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	if err := validateRuntimeToken("application permission address", address, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ApplicationPermissionsPrefix + "/" + appID + "/" + address, nil
}

func ApplicationReceiptKey(executionID string) (string, error) {
	if err := validateRuntimeToken("application execution id", executionID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ApplicationReceiptPrefix + "/" + executionID, nil
}

func ApplicationOutboxKey(outputID string) (string, error) {
	if err := validateRuntimeToken("application async output id", outputID, MaxZoneEndpointLength); err != nil {
		return "", err
	}
	return ApplicationOutboxPrefix + "/" + outputID, nil
}

func DefaultApplicationRuntimeBoundary(runtimeID string) (ApplicationRuntimeBoundary, error) {
	boundary := ApplicationRuntimeBoundary{
		ZoneID:				ZoneIDApplication,
		RuntimeID:			runtimeID,
		AllowedStatePrefixes:		DefaultApplicationZoneBoundary().OwnsPrefixes,
		CrossZoneEffectMechanism:	"zone-messages-only",
	}
	boundary.BoundaryHash = ComputeApplicationRuntimeBoundaryHash(boundary)
	return boundary, boundary.ValidateHash()
}

func RouteApplicationAppShard(appID string, shardCount uint32, layoutEpoch uint64) (ApplicationShardRoute, error) {
	key, err := ApplicationAppKey(appID)
	if err != nil {
		return ApplicationShardRoute{}, err
	}
	return routeApplicationStateKey(ApplicationRouteAppID, appID, key, shardCount, layoutEpoch)
}

func RouteApplicationWorkflowShard(workflowID string, shardCount uint32, layoutEpoch uint64) (ApplicationShardRoute, error) {
	key, err := ApplicationWorkflowKey(workflowID)
	if err != nil {
		return ApplicationShardRoute{}, err
	}
	return routeApplicationStateKey(ApplicationRouteWorkflowID, workflowID, key, shardCount, layoutEpoch)
}

func RouteApplicationSchedulerShard(bucket string, shardCount uint32, layoutEpoch uint64) (ApplicationShardRoute, error) {
	if err := validateRuntimeToken("application scheduler bucket", bucket, MaxZoneEndpointLength); err != nil {
		return ApplicationShardRoute{}, err
	}
	return routeApplicationStateKey(ApplicationRouteSchedulerBucket, bucket, ApplicationSchedulerPrefix+"/"+bucket, shardCount, layoutEpoch)
}

func NewApplicationSchedulerQueue(height uint64, maxWorkPerBlock uint32, tasks []ApplicationScheduledTask) (ApplicationSchedulerQueue, error) {
	queue := ApplicationSchedulerQueue{
		ZoneID:			ZoneIDApplication,
		Height:			height,
		MaxWorkPerBlock:	maxWorkPerBlock,
		Tasks:			cloneApplicationScheduledTasks(tasks),
	}
	return queue, queue.Validate()
}

func EmitApplicationAsyncOutput(queues ZoneMessageQueues, output ApplicationAsyncOutput) (ZoneMessageQueues, ApplicationExecutionReceipt, error) {
	if err := output.ValidateFormat(); err != nil {
		return ZoneMessageQueues{}, ApplicationExecutionReceipt{}, err
	}
	if queues.ZoneID != ZoneIDApplication {
		return ZoneMessageQueues{}, ApplicationExecutionReceipt{}, errors.New("application async output queue route mismatch")
	}
	output.OutputID = ComputeApplicationAsyncOutputID(output)
	output.MessageHash = ComputeApplicationAsyncOutputHash(output)
	if err := output.ValidateHash(); err != nil {
		return ZoneMessageQueues{}, ApplicationExecutionReceipt{}, err
	}
	msg := ZoneMessage{
		ZoneID:		ZoneIDApplication,
		MessageType:	"application.async_output",
		Source:		ApplicationOutboxPrefix,
		Destination:	string(output.DestinationZone) + ":" + output.Destination,
		GasLimit:	output.GasLimit,
		PayloadHash:	output.MessageHash,
		Sequence:	output.RetryNonce,
	}
	next, err := queues.EnqueueOutbox(msg)
	if err != nil {
		return ZoneMessageQueues{}, ApplicationExecutionReceipt{}, err
	}
	receipt, err := NewApplicationExecutionReceipt(ApplicationExecutionReceipt{
		ZoneID:		ZoneIDApplication,
		Height:		output.CreatedHeight,
		ExecutionID:	output.OutputID,
		TaskID:		"async-output",
		WorkflowID:	output.WorkflowID,
		AppID:		output.AppID,
		Status:		ApplicationTaskExecuted,
		GasUsed:	output.GasLimit,
		OutputHash:	output.MessageHash,
		OutboxMessages:	1,
		Sequence:	output.RetryNonce,
	})
	if err != nil {
		return ZoneMessageQueues{}, ApplicationExecutionReceipt{}, err
	}
	return next, receipt, nil
}

func ExecuteApplicationScheduledTasks(queue ApplicationSchedulerQueue, height uint64, limit ApplicationWorkLimit) (ApplicationSchedulerQueue, []ApplicationScheduledTask, error) {
	if err := queue.Validate(); err != nil {
		return ApplicationSchedulerQueue{}, nil, err
	}
	if height == 0 {
		return ApplicationSchedulerQueue{}, nil, errors.New("application execution height must be positive")
	}
	if err := limit.Validate(); err != nil {
		return ApplicationSchedulerQueue{}, nil, err
	}
	maxTasks := limit.MaxTasksPerBlock
	if queue.MaxWorkPerBlock < maxTasks {
		maxTasks = queue.MaxWorkPerBlock
	}
	ready := make([]ApplicationScheduledTask, 0, maxTasks)
	remaining := make([]ApplicationScheduledTask, 0, len(queue.Tasks))
	var gasUsed uint64
	var messageCount uint32
	for _, task := range queue.Tasks {
		if task.Status == ApplicationTaskCanceled || task.Status == ApplicationTaskExecuted {
			continue
		}
		if task.ScheduledHeight > height || uint32(len(ready)) >= maxTasks {
			remaining = append(remaining, task)
			continue
		}
		nextGas, err := addZoneGas(gasUsed, task.GasLimit)
		if err != nil {
			return ApplicationSchedulerQueue{}, nil, err
		}
		if nextGas > limit.MaxGasPerBlock || messageCount+1 > limit.MaxMessagesPerBlock {
			task.Status = ApplicationTaskDeferred
			remaining = append(remaining, task)
			continue
		}
		task.Status = ApplicationTaskExecuted
		ready = append(ready, task)
		gasUsed = nextGas
		messageCount++
	}
	next := queue
	next.Height = height
	next.Tasks = cloneApplicationScheduledTasks(remaining)
	return next, ready, next.Validate()
}

func CancelApplicationScheduledTask(queue ApplicationSchedulerQueue, bucket, taskID string) (ApplicationSchedulerQueue, ApplicationScheduledTask, error) {
	if err := queue.Validate(); err != nil {
		return ApplicationSchedulerQueue{}, ApplicationScheduledTask{}, err
	}
	if err := validateRuntimeToken("application scheduler bucket", bucket, MaxZoneEndpointLength); err != nil {
		return ApplicationSchedulerQueue{}, ApplicationScheduledTask{}, err
	}
	if err := validateRuntimeToken("application task id", taskID, MaxZoneEndpointLength); err != nil {
		return ApplicationSchedulerQueue{}, ApplicationScheduledTask{}, err
	}
	next := queue
	next.Tasks = make([]ApplicationScheduledTask, 0, len(queue.Tasks))
	for _, task := range queue.Tasks {
		if task.Bucket == bucket && task.TaskID == taskID {
			task.Status = ApplicationTaskCanceled
			return next, task, nil
		}
		next.Tasks = append(next.Tasks, task)
	}
	return ApplicationSchedulerQueue{}, ApplicationScheduledTask{}, errors.New("application scheduled task not found")
}

func StartApplicationWorkflow(state ApplicationZoneState, workflow ApplicationWorkflowState, height uint64) (ApplicationZoneState, ApplicationExecutionReceipt, error) {
	if height == 0 {
		return ApplicationZoneState{}, ApplicationExecutionReceipt{}, errors.New("application workflow start height must be positive")
	}
	workflow.Status = ApplicationWorkflowRunning
	workflow.UpdatedHeight = height
	if workflow.CurrentStep == 0 {
		workflow.CurrentStep = 1
	}
	if err := workflow.Validate(); err != nil {
		return ApplicationZoneState{}, ApplicationExecutionReceipt{}, err
	}
	next := state.Normalize()
	next.Height = height
	next.Workflows = upsertApplicationWorkflow(next.Workflows, workflow)
	receipt, err := NewApplicationExecutionReceipt(ApplicationExecutionReceipt{
		ZoneID:		ZoneIDApplication,
		Height:		height,
		ExecutionID:	hashRuntimeParts("application-workflow-start", workflow.WorkflowID, fmt.Sprint(height)),
		TaskID:		"workflow-start",
		WorkflowID:	workflow.WorkflowID,
		AppID:		workflow.AppID,
		Status:		ApplicationTaskExecuted,
		GasUsed:	1,
		OutputHash:	ComputeApplicationWorkflowRoot([]ApplicationWorkflowState{workflow}),
		Sequence:	uint64(len(next.Receipts) + 1),
	})
	if err != nil {
		return ApplicationZoneState{}, ApplicationExecutionReceipt{}, err
	}
	next.Receipts = append(next.Receipts, receipt)
	return next.Normalize(), receipt, nil
}

func AdvanceApplicationWorkflow(state ApplicationZoneState, workflowID string, height uint64) (ApplicationZoneState, ApplicationExecutionReceipt, error) {
	if height == 0 {
		return ApplicationZoneState{}, ApplicationExecutionReceipt{}, errors.New("application workflow advance height must be positive")
	}
	next := state.Normalize()
	for i, workflow := range next.Workflows {
		if workflow.WorkflowID != workflowID {
			continue
		}
		if workflow.Status != ApplicationWorkflowRunning {
			return ApplicationZoneState{}, ApplicationExecutionReceipt{}, errors.New("application workflow is not running")
		}
		if workflow.CurrentStep < workflow.TotalSteps {
			workflow.CurrentStep++
		}
		if workflow.CurrentStep == workflow.TotalSteps {
			workflow.Status = ApplicationWorkflowSucceeded
		}
		workflow.UpdatedHeight = height
		next.Workflows[i] = workflow
		receipt, err := NewApplicationExecutionReceipt(ApplicationExecutionReceipt{
			ZoneID:		ZoneIDApplication,
			Height:		height,
			ExecutionID:	hashRuntimeParts("application-workflow-advance", workflow.WorkflowID, fmt.Sprint(height), fmt.Sprint(workflow.CurrentStep)),
			TaskID:		"workflow-advance",
			WorkflowID:	workflow.WorkflowID,
			AppID:		workflow.AppID,
			Status:		ApplicationTaskExecuted,
			GasUsed:	1,
			OutputHash:	ComputeApplicationWorkflowRoot([]ApplicationWorkflowState{workflow}),
			Sequence:	uint64(len(next.Receipts) + 1),
		})
		if err != nil {
			return ApplicationZoneState{}, ApplicationExecutionReceipt{}, err
		}
		next.Receipts = append(next.Receipts, receipt)
		next.Height = height
		return next.Normalize(), receipt, nil
	}
	return ApplicationZoneState{}, ApplicationExecutionReceipt{}, errors.New("application workflow not found")
}

func NewApplicationMessageQueues(inbox []ZoneMessage, outbox []ZoneMessage) (ZoneMessageQueues, error) {
	return NewZoneMessageQueues(ZoneIDApplication, inbox, outbox)
}

func NewApplicationExecutionReceipt(receipt ApplicationExecutionReceipt) (ApplicationExecutionReceipt, error) {
	if receipt.ReceiptHash != "" {
		return ApplicationExecutionReceipt{}, errors.New("application receipt hash must be empty before construction")
	}
	if receipt.ExecutionID == "" {
		receipt.ExecutionID = hashRuntimeParts("application-execution-id", receipt.TaskID, receipt.WorkflowID, receipt.AppID, fmt.Sprint(receipt.Sequence))
	}
	if err := receipt.ValidateFormat(); err != nil {
		return ApplicationExecutionReceipt{}, err
	}
	receipt.ReceiptHash = ComputeApplicationReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func ApplicationProofRequest(kind ApplicationProofKind, key string, height uint64, root string, limit uint32) (ZoneProofRequest, error) {
	if !IsApplicationProofKind(kind) {
		return ZoneProofRequest{}, fmt.Errorf("unknown application proof kind %q", kind)
	}
	if err := validateRuntimeToken("application proof key", key, MaxZoneProofKeyLength); err != nil {
		return ZoneProofRequest{}, err
	}
	req := ZoneProofRequest{
		ZoneID:	ZoneIDApplication,
		Height:	height,
		Kind:	ZoneProofKindState,
		Key:	string(kind) + "/" + key,
		Root:	root,
		Limit:	limit,
	}
	return req, req.Validate()
}

func BuildApplicationZoneRoot(roots ApplicationZoneRoots) (ZoneRoot, error) {
	if err := roots.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	stateRoot := roots.StateRoot
	if stateRoot == "" {
		stateRoot = hashRuntimeParts(
			"aetra-application-zone-state-v2",
			roots.AppRoot,
			roots.WorkflowRoot,
			roots.SchedulerRoot,
			roots.AutomationRoot,
			roots.PermissionRoot,
			roots.ReceiptRoot,
		)
	}
	root := ZoneRoot{
		ZoneID:			ZoneIDApplication,
		Height:			roots.Height,
		ZoneStateRoot:		stateRoot,
		InboxRoot:		roots.InboxRoot,
		OutboxRoot:		roots.OutboxRoot,
		ReceiptRoot:		roots.ReceiptRoot,
		EventRoot:		EmptyRootHash(),
		ExecutionResultRoot:	roots.ExecutionRoot,
		ProofRoot:		roots.ProofRoot,
	}
	root.RootHash = ComputeZoneRootHash(root)
	return root, root.Validate()
}

func BuildApplicationZoneRootFromState(height uint64, state ApplicationZoneState, queues ZoneMessageQueues, proofRoot string) (ZoneRoot, error) {
	if err := queues.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	if queues.ZoneID != ZoneIDApplication {
		return ZoneRoot{}, errors.New("application zone root queue route mismatch")
	}
	normalized := state.Normalize()
	roots := ApplicationZoneRoots{
		Height:		height,
		AppRoot:	ComputeApplicationAppRoot(normalized.Apps),
		WorkflowRoot:	ComputeApplicationWorkflowRoot(normalized.Workflows),
		SchedulerRoot:	ComputeApplicationTaskRoot(normalized.Tasks),
		AutomationRoot:	ComputeApplicationAutomationRoot(normalized.Automations),
		PermissionRoot:	ComputeApplicationPermissionRoot(normalized.Permissions),
		ReceiptRoot:	ComputeApplicationReceiptRoot(normalized.Receipts),
		InboxRoot:	queues.InboxRoot(),
		OutboxRoot:	queues.OutboxRoot(),
		ExecutionRoot:	ComputeApplicationExecutionRoot(normalized.Receipts),
		ProofRoot:	proofRoot,
		StateRoot:	ComputeApplicationZoneStateRoot(normalized),
	}
	roots.QueueRoot = ComputeApplicationQueueRoot(queues)
	return BuildApplicationZoneRoot(roots)
}

func (s ApplicationZoneState) Normalize() ApplicationZoneState {
	s.Apps = normalizeApplicationRecords(s.Apps)
	s.Workflows = normalizeApplicationWorkflows(s.Workflows)
	s.Tasks = cloneApplicationScheduledTasks(s.Tasks)
	s.Automations = normalizeApplicationAutomations(s.Automations)
	s.Permissions = normalizeApplicationPermissions(s.Permissions)
	s.Receipts = normalizeApplicationReceipts(s.Receipts)
	return s
}

func (s ApplicationZoneState) Validate() error {
	normalized := s.Normalize()
	for _, app := range normalized.Apps {
		if err := app.Validate(); err != nil {
			return err
		}
	}
	for _, workflow := range normalized.Workflows {
		if err := workflow.Validate(); err != nil {
			return err
		}
	}
	for _, task := range normalized.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
	}
	for _, automation := range normalized.Automations {
		if err := automation.Validate(); err != nil {
			return err
		}
	}
	for _, permission := range normalized.Permissions {
		if err := permission.Validate(); err != nil {
			return err
		}
	}
	for _, receipt := range normalized.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a ApplicationRecord) Validate() error {
	if _, err := ApplicationAppKey(a.AppID); err != nil {
		return err
	}
	if err := validateRuntimeToken("application owner", a.Owner, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("application runtime id", a.RuntimeID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if a.Version == 0 {
		return errors.New("application version must be positive")
	}
	if err := ValidateHash("application config hash", a.ConfigHash); err != nil {
		return err
	}
	if a.UpdatedHeight == 0 {
		return errors.New("application updated height must be positive")
	}
	return nil
}

func (s ApplicationWorkflowState) Validate() error {
	if _, err := ApplicationWorkflowKey(s.WorkflowID); err != nil {
		return err
	}
	if _, err := ApplicationAppKey(s.AppID); err != nil {
		return err
	}
	if err := validateRuntimeToken("application workflow owner", s.Owner, MaxZoneEndpointLength); err != nil {
		return err
	}
	if !IsApplicationWorkflowStatus(s.Status) {
		return fmt.Errorf("unknown application workflow status %q", s.Status)
	}
	if s.TotalSteps == 0 {
		return errors.New("application workflow total steps must be positive")
	}
	if s.CurrentStep > s.TotalSteps {
		return errors.New("application workflow current step exceeds total steps")
	}
	if err := ValidateHash("application workflow payload hash", s.PayloadHash); err != nil {
		return err
	}
	if s.UpdatedHeight == 0 {
		return errors.New("application workflow updated height must be positive")
	}
	return nil
}

func (t ApplicationScheduledTask) Validate() error {
	if err := validateRuntimeToken("application scheduler bucket", t.Bucket, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("application task id", t.TaskID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if _, err := ApplicationWorkflowKey(t.WorkflowID); err != nil {
		return err
	}
	if _, err := ApplicationAppKey(t.AppID); err != nil {
		return err
	}
	if t.ScheduledHeight == 0 {
		return errors.New("application task scheduled height must be positive")
	}
	if t.Sequence == 0 {
		return errors.New("application task sequence must be positive")
	}
	if t.GasLimit == 0 {
		return errors.New("application task gas limit must be positive")
	}
	if err := ValidateHash("application task payload hash", t.PayloadHash); err != nil {
		return err
	}
	if !IsApplicationTaskStatus(t.Status) {
		return fmt.Errorf("unknown application task status %q", t.Status)
	}
	return nil
}

func (a ApplicationAutomation) Validate() error {
	if _, err := ApplicationAutomationKey(a.AutomationID); err != nil {
		return err
	}
	if _, err := ApplicationAppKey(a.AppID); err != nil {
		return err
	}
	if _, err := ApplicationWorkflowKey(a.WorkflowID); err != nil {
		return err
	}
	if err := ValidateHash("application automation trigger hash", a.TriggerHash); err != nil {
		return err
	}
	if a.NextRunHeight == 0 {
		return errors.New("application automation next run height must be positive")
	}
	if a.UpdatedHeight == 0 {
		return errors.New("application automation updated height must be positive")
	}
	return nil
}

func (p ApplicationPermission) Validate() error {
	if _, err := ApplicationPermissionKey(p.AppID, p.Address); err != nil {
		return err
	}
	if !IsApplicationPermissionScope(p.Scope) {
		return fmt.Errorf("unknown application permission scope %q", p.Scope)
	}
	if err := ValidateHash("application permission grant hash", p.GrantHash); err != nil {
		return err
	}
	return nil
}

func (q ApplicationSchedulerQueue) Validate() error {
	if q.ZoneID != ZoneIDApplication {
		return errors.New("application scheduler queue must use APPLICATION_ZONE")
	}
	if q.Height == 0 {
		return errors.New("application scheduler queue height must be positive")
	}
	if q.MaxWorkPerBlock == 0 {
		return errors.New("application scheduler max work per block must be positive")
	}
	for i, task := range q.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		if i > 0 && compareApplicationScheduledTasks(q.Tasks[i-1], task) >= 0 {
			return errors.New("application scheduler tasks must be sorted canonically")
		}
	}
	return nil
}

func (l ApplicationWorkLimit) Validate() error {
	if l.MaxTasksPerBlock == 0 {
		return errors.New("application work max tasks per block must be positive")
	}
	if l.MaxMessagesPerBlock == 0 {
		return errors.New("application work max messages per block must be positive")
	}
	if l.MaxGasPerBlock == 0 {
		return errors.New("application work max gas per block must be positive")
	}
	return nil
}

func (r ApplicationExecutionReceipt) ValidateFormat() error {
	if r.ZoneID != ZoneIDApplication {
		return errors.New("application receipt must use APPLICATION_ZONE")
	}
	if r.Height == 0 {
		return errors.New("application receipt height must be positive")
	}
	if _, err := ApplicationReceiptKey(r.ExecutionID); err != nil {
		return err
	}
	if err := validateRuntimeToken("application receipt task id", r.TaskID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if _, err := ApplicationWorkflowKey(r.WorkflowID); err != nil {
		return err
	}
	if _, err := ApplicationAppKey(r.AppID); err != nil {
		return err
	}
	if !IsApplicationTaskStatus(r.Status) {
		return fmt.Errorf("unknown application receipt status %q", r.Status)
	}
	if err := ValidateHash("application receipt output hash", r.OutputHash); err != nil {
		return err
	}
	if r.Sequence == 0 {
		return errors.New("application receipt sequence must be positive")
	}
	if r.ReceiptHash != "" {
		return ValidateHash("application receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r ApplicationExecutionReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("application receipt hash is required")
	}
	if r.ReceiptHash != ComputeApplicationReceiptHash(r) {
		return errors.New("application receipt hash mismatch")
	}
	return nil
}

func (r ApplicationExecutionReceipt) ZoneReceipt() (ZoneReceipt, error) {
	status := ZoneReceiptStatusSuccess
	if r.Status == ApplicationTaskFailed {
		status = ZoneReceiptStatusFailed
	}
	return NewZoneReceipt(ZoneReceipt{
		ZoneID:		ZoneIDApplication,
		Height:		r.Height,
		ItemHash:	hashRuntimeParts("application-receipt-item-v1", r.ExecutionID, r.TaskID, r.WorkflowID, r.AppID),
		Status:		status,
		GasUsed:	r.GasUsed,
		ResultHash:	r.OutputHash,
		Sequence:	r.Sequence,
	})
}

func (b ApplicationRuntimeBoundary) ValidateHash() error {
	if b.ZoneID != ZoneIDApplication {
		return errors.New("application runtime boundary must use APPLICATION_ZONE")
	}
	if err := validateRuntimeToken("application runtime id", b.RuntimeID, MaxZoneEndpointLength); err != nil {
		return err
	}
	if b.CrossZoneEffectMechanism != "zone-messages-only" {
		return errors.New("application runtime boundary must use zone messages for cross-zone effects")
	}
	if len(b.AllowedStatePrefixes) == 0 {
		return errors.New("application runtime boundary requires state prefixes")
	}
	var previous string
	for i, prefix := range b.AllowedStatePrefixes {
		if err := validateRuntimeToken("application runtime boundary prefix", prefix, MaxZoneNamespaceLength); err != nil {
			return err
		}
		if i > 0 && previous >= prefix {
			return errors.New("application runtime boundary prefixes must be sorted canonically")
		}
		previous = prefix
	}
	if err := ValidateHash("application runtime boundary hash", b.BoundaryHash); err != nil {
		return err
	}
	if b.BoundaryHash != ComputeApplicationRuntimeBoundaryHash(b) {
		return errors.New("application runtime boundary hash mismatch")
	}
	return nil
}

func (r ApplicationShardRoute) ValidateHash() error {
	if r.ZoneID != ZoneIDApplication {
		return errors.New("application shard route must use APPLICATION_ZONE")
	}
	if r.LayoutEpoch == 0 || r.ShardCount == 0 {
		return errors.New("application shard route requires layout epoch and shard count")
	}
	if r.ShardID >= r.ShardCount {
		return errors.New("application shard route shard id out of range")
	}
	if !IsApplicationShardRoutingMode(r.RoutingMode) {
		return fmt.Errorf("unknown application shard routing mode %q", r.RoutingMode)
	}
	if err := validateRuntimeToken("application shard route key", r.RouteKey, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := validateRuntimeToken("application shard state key", r.StateKey, MaxZoneNamespaceLength); err != nil {
		return err
	}
	if err := ValidateHash("application shard route hash", r.RouteHash); err != nil {
		return err
	}
	if r.RouteHash != ComputeApplicationShardRouteHash(r) {
		return errors.New("application shard route hash mismatch")
	}
	return nil
}

func (o ApplicationAsyncOutput) ValidateFormat() error {
	if o.OutputID != "" {
		if _, err := ApplicationOutboxKey(o.OutputID); err != nil {
			return err
		}
	}
	if _, err := ApplicationAppKey(o.AppID); err != nil {
		return err
	}
	if _, err := ApplicationWorkflowKey(o.WorkflowID); err != nil {
		return err
	}
	if err := ValidateZoneID(o.DestinationZone); err != nil {
		return err
	}
	if err := validateRuntimeToken("application async output destination", o.Destination, MaxZoneEndpointLength); err != nil {
		return err
	}
	if err := ValidateHash("application async output payload hash", o.PayloadHash); err != nil {
		return err
	}
	if o.GasLimit == 0 || o.RetryNonce == 0 || o.CreatedHeight == 0 {
		return errors.New("application async output gas, retry nonce, and height must be positive")
	}
	if o.MessageHash != "" {
		return ValidateHash("application async output hash", o.MessageHash)
	}
	return nil
}

func (o ApplicationAsyncOutput) ValidateHash() error {
	if err := o.ValidateFormat(); err != nil {
		return err
	}
	if o.OutputID != ComputeApplicationAsyncOutputID(o) {
		return errors.New("application async output id mismatch")
	}
	if o.MessageHash != ComputeApplicationAsyncOutputHash(o) {
		return errors.New("application async output hash mismatch")
	}
	return nil
}

func (p ApplicationProofRootExport) Validate() error {
	if p.ZoneID != ZoneIDApplication {
		return errors.New("application proof root export must use APPLICATION_ZONE")
	}
	if p.Height == 0 {
		return errors.New("application proof root export height must be positive")
	}
	if !IsApplicationProofRootType(p.RootType) {
		return fmt.Errorf("unknown application proof root type %q", p.RootType)
	}
	if err := ValidateHash("application proof root export hash", p.RootHash); err != nil {
		return err
	}
	return validateRuntimeToken("application proof root export source", p.Source, MaxZoneNamespaceLength)
}

func (r ApplicationZoneRoots) Validate() error {
	if r.Height == 0 {
		return errors.New("application zone root height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "application app root", value: r.AppRoot},
		{name: "application workflow root", value: r.WorkflowRoot},
		{name: "application scheduler root", value: r.SchedulerRoot},
		{name: "application automation root", value: r.AutomationRoot},
		{name: "application permission root", value: r.PermissionRoot},
		{name: "application receipt root", value: r.ReceiptRoot},
		{name: "application queue root", value: r.QueueRoot},
		{name: "application inbox root", value: r.InboxRoot},
		{name: "application outbox root", value: r.OutboxRoot},
		{name: "application execution root", value: r.ExecutionRoot},
		{name: "application proof root", value: r.ProofRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.StateRoot != "" {
		return ValidateHash("application state root", r.StateRoot)
	}
	return nil
}

func ComputeApplicationRuntimeBoundaryHash(boundary ApplicationRuntimeBoundary) string {
	prefixes := append([]string(nil), boundary.AllowedStatePrefixes...)
	sort.Strings(prefixes)
	parts := []string{"aetra-application-runtime-boundary-v1", string(boundary.ZoneID), boundary.RuntimeID, boundary.CrossZoneEffectMechanism, fmt.Sprint(len(prefixes))}
	parts = append(parts, prefixes...)
	return hashRuntimeParts(parts...)
}

func ComputeApplicationShardRouteHash(route ApplicationShardRoute) string {
	return hashRuntimeParts(
		"aetra-application-shard-route-v1",
		string(route.ZoneID),
		fmt.Sprint(route.LayoutEpoch),
		fmt.Sprint(route.ShardCount),
		fmt.Sprint(route.ShardID),
		string(route.RoutingMode),
		route.RouteKey,
		route.StateKey,
	)
}

func ComputeApplicationAsyncOutputID(output ApplicationAsyncOutput) string {
	return hashRuntimeParts("aetra-application-async-output-id-v1", output.AppID, output.WorkflowID, string(output.DestinationZone), output.Destination, fmt.Sprint(output.CreatedHeight), fmt.Sprint(output.RetryNonce))
}

func ComputeApplicationAsyncOutputHash(output ApplicationAsyncOutput) string {
	outputID := output.OutputID
	if outputID == "" {
		outputID = ComputeApplicationAsyncOutputID(output)
	}
	return hashRuntimeParts("aetra-application-async-output-v1", outputID, output.AppID, output.WorkflowID, string(output.DestinationZone), output.Destination, output.PayloadHash, fmt.Sprint(output.GasLimit), fmt.Sprint(output.RetryNonce), fmt.Sprint(output.CreatedHeight))
}

func ComputeApplicationReceiptHash(receipt ApplicationExecutionReceipt) string {
	return hashRuntimeParts(
		"aetra-application-receipt-v2",
		string(receipt.ZoneID),
		fmt.Sprintf("%020d", receipt.Height),
		receipt.ExecutionID,
		receipt.TaskID,
		receipt.WorkflowID,
		receipt.AppID,
		string(receipt.Status),
		fmt.Sprintf("%020d", receipt.GasUsed),
		receipt.OutputHash,
		fmt.Sprintf("%010d", receipt.OutboxMessages),
		fmt.Sprintf("%020d", receipt.Sequence),
	)
}

func ComputeApplicationZoneStateRoot(state ApplicationZoneState) string {
	normalized := state.Normalize()
	return hashRuntimeParts(
		"aetra-application-zone-state-v2",
		ComputeApplicationAppRoot(normalized.Apps),
		ComputeApplicationWorkflowRoot(normalized.Workflows),
		ComputeApplicationTaskRoot(normalized.Tasks),
		ComputeApplicationAutomationRoot(normalized.Automations),
		ComputeApplicationPermissionRoot(normalized.Permissions),
		ComputeApplicationReceiptRoot(normalized.Receipts),
	)
}

func ComputeApplicationQueueRoot(queues ZoneMessageQueues) string {
	return queues.QueueRoot()
}

func BuildApplicationProofRootExports(height uint64, roots ApplicationZoneRoots) ([]ApplicationProofRootExport, error) {
	if roots.Height != height {
		return nil, errors.New("application proof root export height mismatch")
	}
	if err := roots.Validate(); err != nil {
		return nil, err
	}
	exports := []ApplicationProofRootExport{
		{ZoneID: ZoneIDApplication, Height: height, RootType: ApplicationProofRootApp, RootHash: roots.AppRoot, Source: "application.zone.app"},
		{ZoneID: ZoneIDApplication, Height: height, RootType: ApplicationProofRootWorkflow, RootHash: roots.WorkflowRoot, Source: "application.zone.workflow"},
		{ZoneID: ZoneIDApplication, Height: height, RootType: ApplicationProofRootScheduler, RootHash: roots.SchedulerRoot, Source: "application.zone.scheduler"},
		{ZoneID: ZoneIDApplication, Height: height, RootType: ApplicationProofRootQueue, RootHash: roots.QueueRoot, Source: "application.zone.queue"},
		{ZoneID: ZoneIDApplication, Height: height, RootType: ApplicationProofRootPermission, RootHash: roots.PermissionRoot, Source: "application.zone.permission"},
	}
	sort.SliceStable(exports, func(i, j int) bool { return exports[i].RootType < exports[j].RootType })
	for _, export := range exports {
		if err := export.Validate(); err != nil {
			return nil, err
		}
	}
	return exports, nil
}

func ComputeApplicationAppRoot(apps []ApplicationRecord) string {
	ordered := normalizeApplicationRecords(apps)
	parts := []string{"aetra-application-app-root-v1", fmt.Sprint(len(ordered))}
	for _, app := range ordered {
		parts = append(parts, app.AppID, app.Owner, app.RuntimeID, fmt.Sprint(app.Version), fmt.Sprint(app.Enabled), app.ConfigHash, fmt.Sprint(app.UpdatedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationWorkflowRoot(workflows []ApplicationWorkflowState) string {
	ordered := normalizeApplicationWorkflows(workflows)
	parts := []string{"aetra-application-workflow-root-v1", fmt.Sprint(len(ordered))}
	for _, workflow := range ordered {
		parts = append(parts, workflow.WorkflowID, workflow.AppID, workflow.Owner, string(workflow.Status), fmt.Sprint(workflow.CurrentStep), fmt.Sprint(workflow.TotalSteps), workflow.PayloadHash, fmt.Sprint(workflow.UpdatedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationTaskRoot(tasks []ApplicationScheduledTask) string {
	ordered := cloneApplicationScheduledTasks(tasks)
	parts := []string{"aetra-application-task-root-v1", fmt.Sprint(len(ordered))}
	for _, task := range ordered {
		key, _ := ApplicationSchedulerTaskKey(task)
		parts = append(parts, key, task.WorkflowID, task.AppID, fmt.Sprint(task.ScheduledHeight), fmt.Sprint(task.Priority), fmt.Sprint(task.Sequence), fmt.Sprint(task.GasLimit), task.PayloadHash, string(task.Status))
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationAutomationRoot(automations []ApplicationAutomation) string {
	ordered := normalizeApplicationAutomations(automations)
	parts := []string{"aetra-application-automation-root-v1", fmt.Sprint(len(ordered))}
	for _, automation := range ordered {
		parts = append(parts, automation.AutomationID, automation.AppID, automation.WorkflowID, fmt.Sprint(automation.Enabled), automation.TriggerHash, fmt.Sprint(automation.NextRunHeight), fmt.Sprint(automation.UpdatedHeight))
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationPermissionRoot(permissions []ApplicationPermission) string {
	ordered := normalizeApplicationPermissions(permissions)
	parts := []string{"aetra-application-permission-root-v1", fmt.Sprint(len(ordered))}
	for _, permission := range ordered {
		parts = append(parts, permission.AppID, permission.Address, string(permission.Scope), fmt.Sprint(permission.ExpiresHeight), permission.GrantHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationReceiptRoot(receipts []ApplicationExecutionReceipt) string {
	ordered := normalizeApplicationReceipts(receipts)
	parts := []string{"aetra-application-receipt-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ReceiptHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationExecutionRoot(receipts []ApplicationExecutionReceipt) string {
	ordered := normalizeApplicationReceipts(receipts)
	parts := []string{"aetra-application-execution-root-v1", fmt.Sprint(len(ordered))}
	for _, receipt := range ordered {
		parts = append(parts, receipt.ExecutionID, receipt.OutputHash, string(receipt.Status), fmt.Sprint(receipt.GasUsed))
	}
	return hashRuntimeParts(parts...)
}

func ComputeApplicationSchedulerRoot(queue ApplicationSchedulerQueue) (string, error) {
	if err := queue.Validate(); err != nil {
		return "", err
	}
	parts := []string{
		"aetra-application-scheduler-root-v1",
		fmt.Sprintf("%020d", queue.Height),
		fmt.Sprintf("%010d", queue.MaxWorkPerBlock),
	}
	for _, task := range queue.Tasks {
		key, err := ApplicationSchedulerTaskKey(task)
		if err != nil {
			return "", err
		}
		parts = append(parts, key, task.PayloadHash, string(task.Status), fmt.Sprintf("%020d", task.GasLimit))
	}
	return hashRuntimeParts(parts...), nil
}

func IsApplicationWorkflowStatus(status ApplicationWorkflowStatus) bool {
	switch status {
	case ApplicationWorkflowPending, ApplicationWorkflowRunning, ApplicationWorkflowSucceeded, ApplicationWorkflowFailed, ApplicationWorkflowCanceled:
		return true
	default:
		return false
	}
}

func IsApplicationTaskStatus(status ApplicationTaskStatus) bool {
	switch status {
	case ApplicationTaskPending, ApplicationTaskExecuted, ApplicationTaskFailed, ApplicationTaskDeferred, ApplicationTaskSkipped, ApplicationTaskCanceled:
		return true
	default:
		return false
	}
}

func IsApplicationMessageKind(kind ApplicationMessageKind) bool {
	switch kind {
	case ApplicationMessageCreateApp,
		ApplicationMessageUpdateApp,
		ApplicationMessageStartWorkflow,
		ApplicationMessageAdvanceWorkflow,
		ApplicationMessageScheduleTask,
		ApplicationMessageCancelTask,
		ApplicationMessageExecuteAutomation:
		return true
	default:
		return false
	}
}

func IsApplicationProofKind(kind ApplicationProofKind) bool {
	switch kind {
	case ApplicationProofApp, ApplicationProofWorkflow, ApplicationProofScheduledTask, ApplicationProofAutomation, ApplicationProofAppReceipts, ApplicationProofAppQueue, ApplicationProofPermissions:
		return true
	default:
		return false
	}
}

func IsApplicationShardRoutingMode(mode ApplicationShardRoutingMode) bool {
	switch mode {
	case ApplicationRouteAppID, ApplicationRouteWorkflowID, ApplicationRouteSchedulerBucket:
		return true
	default:
		return false
	}
}

func IsApplicationProofRootType(rootType ApplicationProofRootType) bool {
	switch rootType {
	case ApplicationProofRootApp, ApplicationProofRootWorkflow, ApplicationProofRootScheduler, ApplicationProofRootQueue, ApplicationProofRootPermission:
		return true
	default:
		return false
	}
}

func IsApplicationPermissionScope(scope ApplicationPermissionScope) bool {
	switch scope {
	case ApplicationPermissionAdmin, ApplicationPermissionExecute, ApplicationPermissionSchedule:
		return true
	default:
		return false
	}
}

func normalizeApplicationRecords(apps []ApplicationRecord) []ApplicationRecord {
	out := append([]ApplicationRecord(nil), apps...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].AppID < out[j].AppID })
	return out
}

func normalizeApplicationWorkflows(workflows []ApplicationWorkflowState) []ApplicationWorkflowState {
	out := append([]ApplicationWorkflowState(nil), workflows...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].WorkflowID < out[j].WorkflowID })
	return out
}

func normalizeApplicationAutomations(automations []ApplicationAutomation) []ApplicationAutomation {
	out := append([]ApplicationAutomation(nil), automations...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].AutomationID < out[j].AutomationID })
	return out
}

func normalizeApplicationPermissions(permissions []ApplicationPermission) []ApplicationPermission {
	out := append([]ApplicationPermission(nil), permissions...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].AppID != out[j].AppID {
			return out[i].AppID < out[j].AppID
		}
		if out[i].Address != out[j].Address {
			return out[i].Address < out[j].Address
		}
		return out[i].Scope < out[j].Scope
	})
	return out
}

func normalizeApplicationReceipts(receipts []ApplicationExecutionReceipt) []ApplicationExecutionReceipt {
	out := append([]ApplicationExecutionReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].ExecutionID < out[j].ExecutionID
	})
	return out
}

func cloneApplicationScheduledTasks(tasks []ApplicationScheduledTask) []ApplicationScheduledTask {
	out := append([]ApplicationScheduledTask(nil), tasks...)
	sort.SliceStable(out, func(i, j int) bool {
		return compareApplicationScheduledTasks(out[i], out[j]) < 0
	})
	return out
}

func upsertApplicationWorkflow(workflows []ApplicationWorkflowState, workflow ApplicationWorkflowState) []ApplicationWorkflowState {
	out := append([]ApplicationWorkflowState(nil), workflows...)
	for i := range out {
		if out[i].WorkflowID == workflow.WorkflowID {
			out[i] = workflow
			return normalizeApplicationWorkflows(out)
		}
	}
	out = append(out, workflow)
	return normalizeApplicationWorkflows(out)
}

func compareApplicationScheduledTasks(left, right ApplicationScheduledTask) int {
	if left.Bucket < right.Bucket {
		return -1
	}
	if left.Bucket > right.Bucket {
		return 1
	}
	if left.ScheduledHeight < right.ScheduledHeight {
		return -1
	}
	if left.ScheduledHeight > right.ScheduledHeight {
		return 1
	}
	if left.Priority > right.Priority {
		return -1
	}
	if left.Priority < right.Priority {
		return 1
	}
	if left.TaskID < right.TaskID {
		return -1
	}
	if left.TaskID > right.TaskID {
		return 1
	}
	if left.WorkflowID < right.WorkflowID {
		return -1
	}
	if left.WorkflowID > right.WorkflowID {
		return 1
	}
	if left.AppID < right.AppID {
		return -1
	}
	if left.AppID > right.AppID {
		return 1
	}
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	return 0
}

func routeApplicationStateKey(mode ApplicationShardRoutingMode, routeKey string, stateKey string, shardCount uint32, layoutEpoch uint64) (ApplicationShardRoute, error) {
	if shardCount == 0 {
		return ApplicationShardRoute{}, errors.New("application shard count must be positive")
	}
	if layoutEpoch == 0 {
		return ApplicationShardRoute{}, errors.New("application shard layout epoch must be positive")
	}
	if err := validateRuntimeToken("application shard route key", routeKey, MaxZoneEndpointLength); err != nil {
		return ApplicationShardRoute{}, err
	}
	if err := validateRuntimeToken("application shard state key", stateKey, MaxZoneNamespaceLength); err != nil {
		return ApplicationShardRoute{}, err
	}
	hash := hashRuntimeParts("aetra-application-route-key-v1", string(mode), routeKey, fmt.Sprint(layoutEpoch))
	bytes, err := hex.DecodeString(hash[:16])
	if err != nil {
		return ApplicationShardRoute{}, err
	}
	route := ApplicationShardRoute{
		ZoneID:		ZoneIDApplication,
		LayoutEpoch:	layoutEpoch,
		ShardCount:	shardCount,
		ShardID:	uint32(binary.BigEndian.Uint64(bytes) % uint64(shardCount)),
		RoutingMode:	mode,
		RouteKey:	routeKey,
		StateKey:	stateKey,
	}
	route.RouteHash = ComputeApplicationShardRouteHash(route)
	return route, route.ValidateHash()
}
