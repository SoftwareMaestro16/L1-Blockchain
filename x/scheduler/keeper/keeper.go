package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
)

var genesisKey = []byte{0x01}

type JobHandler func(job schedulertypes.ScheduledJob, currentHeight uint64) schedulertypes.JobExecutionResult

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	SchedulerParams	schedulertypes.SchedulerParams
	State		schedulertypes.SchedulerState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
	handlers	map[string]JobHandler
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis(), handlers: map[string]JobHandler{}}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService, handlers: map[string]JobHandler{}}
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		Version:		prototype.CurrentGenesisVersion,
		Params:			prototype.DefaultParams(),
		SchedulerParams:	schedulertypes.DefaultSchedulerParams(),
		State:			schedulertypes.EmptySchedulerState(),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("scheduler prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.SchedulerParams.Normalize())
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) UpdateParams(authority string, params prototype.Params, schedulerParams schedulertypes.SchedulerParams) error {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	if err := params.Validate(); err != nil {
		return err
	}
	schedulerParams = schedulerParams.Normalize()
	if err := k.genesis.State.Validate(schedulerParams); err != nil {
		return err
	}
	k.genesis.Params = params
	k.genesis.SchedulerParams = schedulerParams
	return nil
}

func (k *Keeper) RegisterJobHandler(ownerModule string, handler JobHandler) error {
	if ownerModule == "" {
		return errors.New("scheduler handler owner module is required")
	}
	if handler == nil {
		return errors.New("scheduler handler is required")
	}
	if k.handlers == nil {
		k.handlers = map[string]JobHandler{}
	}
	k.handlers[ownerModule] = handler
	return nil
}

func (k *Keeper) RegisterScheduledJob(msg schedulertypes.MsgRegisterScheduledJob) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return err
	}
	job := msg.Job.Normalize()
	if err := job.Validate(k.genesis.SchedulerParams); err != nil {
		return err
	}
	if _, _, found := k.findJob(job.OwnerModule, job.ID); found {
		return errors.New("scheduler job already registered")
	}
	next := cloneGenesis(k.genesis)
	next.State.Jobs = schedulertypes.UpsertJob(next.State.Jobs, job)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k *Keeper) PauseScheduledJob(msg schedulertypes.MsgPauseScheduledJob) error {
	return k.setJobPaused(msg.Authority, msg.OwnerModule, msg.JobID, true)
}

func (k *Keeper) ResumeScheduledJob(msg schedulertypes.MsgResumeScheduledJob) error {
	return k.setJobPaused(msg.Authority, msg.OwnerModule, msg.JobID, false)
}

func (k *Keeper) CancelScheduledJob(msg schedulertypes.MsgCancelScheduledJob) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return err
	}
	index, job, found := k.findJob(msg.OwnerModule, msg.JobID)
	if !found {
		return errors.New("scheduler job not found")
	}
	job.Cancelled = true
	job.Paused = true
	job.NextExecutionHeight = 0
	next := cloneGenesis(k.genesis)
	next.State.Jobs[index] = job.Normalize()
	schedulertypes.SortJobs(next.State.Jobs)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k *Keeper) ExecuteDueJobs(msg schedulertypes.MsgExecuteDueJobs) (schedulertypes.ExecutionBatchResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return schedulertypes.ExecutionBatchResult{}, err
	}
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return schedulertypes.ExecutionBatchResult{}, err
	}
	if msg.CurrentHeight == 0 {
		return schedulertypes.ExecutionBatchResult{}, errors.New("scheduler execution height must be positive")
	}
	due := schedulertypes.DueJobs(k.genesis.State, msg.CurrentHeight)
	result := schedulertypes.ExecutionBatchResult{Height: msg.CurrentHeight}
	next := cloneGenesis(k.genesis)
	for _, job := range due {
		if result.ExecutedJobs >= next.SchedulerParams.MaxJobsPerBlock {
			result.SkippedJobs++
			continue
		}
		if result.GasReserved+job.MaxGas > next.SchedulerParams.MaxSchedulerGas {
			result.SkippedJobs++
			continue
		}
		result.GasReserved += job.MaxGas
		record, updated := k.executeJob(job, msg.CurrentHeight)
		result.ExecutedJobs++
		result.GasUsed += record.GasUsed
		result.History = append(result.History, record)
		next.State.Jobs = schedulertypes.UpsertJob(next.State.Jobs, updated)
		next.State.History = append(next.State.History, record)
		next.State.History = schedulertypes.PruneHistory(next.State.History, next.SchedulerParams.HistoryRetention)
	}
	next.State.LastExecutionHeight = msg.CurrentHeight
	if err := next.Validate(); err != nil {
		return schedulertypes.ExecutionBatchResult{}, err
	}
	k.genesis = next
	result.RemainingDue = schedulertypes.DueJobs(k.genesis.State, msg.CurrentHeight)
	return result, nil
}

func (k Keeper) ScheduledJob(ownerModule, jobID string) (schedulertypes.ScheduledJob, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return schedulertypes.ScheduledJob{}, false, err
	}
	_, job, found := k.findJob(ownerModule, jobID)
	return job.Normalize(), found, nil
}

func (k Keeper) ScheduledJobs(req *prototype.PageRequest) ([]schedulertypes.ScheduledJob, prototype.PageResponse, error) {
	jobs := k.genesis.State.Export().Jobs
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(jobs))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return append([]schedulertypes.ScheduledJob(nil), jobs[start:end]...), res, nil
}

func (k Keeper) DueJobs(height uint64, req *prototype.PageRequest) ([]schedulertypes.ScheduledJob, prototype.PageResponse, error) {
	jobs := schedulertypes.DueJobs(k.genesis.State, height)
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(jobs))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return append([]schedulertypes.ScheduledJob(nil), jobs[start:end]...), res, nil
}

func (k Keeper) JobHistory(jobID string, req *prototype.PageRequest) ([]schedulertypes.JobHistoryRecord, prototype.PageResponse, error) {
	history := make([]schedulertypes.JobHistoryRecord, 0)
	for _, record := range k.genesis.State.History {
		if jobID == "" || record.JobID == jobID {
			history = append(history, record)
		}
	}
	start, end, res, err := prototype.NormalizePage(req, k.genesis.Params, len(history))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	return append([]schedulertypes.JobHistoryRecord(nil), history[start:end]...), res, nil
}

func (k Keeper) SchedulerParams() schedulertypes.SchedulerParams {
	return k.genesis.SchedulerParams.Normalize()
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k *Keeper) setJobPaused(authority, ownerModule, jobID string, paused bool) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return err
	}
	index, job, found := k.findJob(ownerModule, jobID)
	if !found {
		return errors.New("scheduler job not found")
	}
	if job.Cancelled {
		return errors.New("scheduler cancelled job cannot be changed")
	}
	job.Paused = paused
	next := cloneGenesis(k.genesis)
	next.State.Jobs[index] = job.Normalize()
	schedulertypes.SortJobs(next.State.Jobs)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) executeJob(job schedulertypes.ScheduledJob, currentHeight uint64) (schedulertypes.JobHistoryRecord, schedulertypes.ScheduledJob) {
	job = job.Normalize()
	record := schedulertypes.JobHistoryRecord{
		JobID:		job.ID,
		OwnerModule:	job.OwnerModule,
		Height:		currentHeight,
		Status:		schedulertypes.HistoryStatusFailure,
		Attempt:	job.FailureCount + 1,
	}
	handler := k.handlers[job.OwnerModule]
	if handler == nil {
		return applyFailure(job, currentHeight, record, "scheduler handler not registered")
	}

	exec := safeExecute(handler, job, currentHeight)
	record.GasUsed = exec.GasUsed
	if record.GasUsed > job.MaxGas {
		record.GasUsed = job.MaxGas
		return applyFailure(job, currentHeight, record, "scheduler job exceeded max gas")
	}
	if exec.Success {
		record.Status = schedulertypes.HistoryStatusSuccess
		return record, applySuccess(job, currentHeight)
	}
	if exec.Error == "" {
		exec.Error = "scheduler job failed"
	}
	return applyFailure(job, currentHeight, record, exec.Error)
}

func safeExecute(handler JobHandler, job schedulertypes.ScheduledJob, currentHeight uint64) (result schedulertypes.JobExecutionResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = schedulertypes.JobExecutionResult{
				Success:	false,
				GasUsed:	job.MaxGas,
				Error:		fmt.Sprintf("scheduler job panicked: %v", recovered),
			}
		}
	}()
	return handler(job, currentHeight)
}

func applySuccess(job schedulertypes.ScheduledJob, currentHeight uint64) schedulertypes.ScheduledJob {
	job.FailureCount = 0
	job.ExecutionCount++
	if job.Type == schedulertypes.JobTypeDelayed {
		job.Cancelled = true
		job.Paused = true
		job.NextExecutionHeight = 0
		return job.Normalize()
	}
	job.NextExecutionHeight = currentHeight + job.Interval
	return job.Normalize()
}

func applyFailure(job schedulertypes.ScheduledJob, currentHeight uint64, record schedulertypes.JobHistoryRecord, errText string) (schedulertypes.JobHistoryRecord, schedulertypes.ScheduledJob) {
	record.Status = schedulertypes.HistoryStatusFailure
	record.Error = errText
	job.FailureCount++
	if job.RetryPolicy.MaxRetries > 0 && job.FailureCount <= job.RetryPolicy.MaxRetries {
		job.NextExecutionHeight = currentHeight + job.RetryPolicy.BackoffInterval
		return record, job.Normalize()
	}
	if job.Type == schedulertypes.JobTypeDelayed {
		job.Cancelled = true
		job.Paused = true
		job.NextExecutionHeight = 0
		return record, job.Normalize()
	}
	if job.Interval > 0 {
		job.NextExecutionHeight = currentHeight + job.Interval
	}
	return record, job.Normalize()
}

func (k Keeper) findJob(ownerModule, jobID string) (int, schedulertypes.ScheduledJob, bool) {
	key := schedulertypes.JobKey(schedulertypes.ScheduledJob{OwnerModule: ownerModule, ID: jobID})
	for i, job := range k.genesis.State.Export().Jobs {
		if schedulertypes.JobKey(job) == key {
			return i, job.Normalize(), true
		}
	}
	return -1, schedulertypes.ScheduledJob{}, false
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.SchedulerParams = gs.SchedulerParams.Normalize()
	gs.State = gs.State.Export()
	return gs
}
