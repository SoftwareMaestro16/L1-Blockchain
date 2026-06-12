package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	JobTypePeriodic	= "periodic"
	JobTypeDelayed	= "delayed"
	JobTypeEpoch	= "epoch"

	HistoryStatusSuccess	= "success"
	HistoryStatusFailure	= "failure"
	HistoryStatusSkipped	= "skipped"
)

type SchedulerParams struct {
	MaxJobsPerBlock		uint32
	MaxSchedulerGas		uint64
	MaxGasPerJob		uint64
	AuthorizedModules	[]string
	HistoryRetention	uint32
}

type RetryPolicy struct {
	MaxRetries	uint32
	BackoffInterval	uint64
}

type ScheduledJob struct {
	ID			string
	OwnerModule		string
	Type			string
	NextExecutionHeight	uint64
	Interval		uint64
	MaxGas			uint64
	RetryPolicy		RetryPolicy
	FailureCount		uint32
	Paused			bool
	Cancelled		bool
	Payload			[]byte
	ExecutionCount		uint64
}

type SchedulerState struct {
	Jobs			[]ScheduledJob
	History			[]JobHistoryRecord
	LastExecutionHeight	uint64
}

type JobHistoryRecord struct {
	JobID		string
	OwnerModule	string
	Height		uint64
	Status		string
	GasUsed		uint64
	Error		string
	Attempt		uint32
}

type JobExecutionResult struct {
	Success	bool
	GasUsed	uint64
	Error	string
}

type ExecutionBatchResult struct {
	Height		uint64
	ExecutedJobs	uint32
	SkippedJobs	uint32
	GasReserved	uint64
	GasUsed		uint64
	History		[]JobHistoryRecord
	RemainingDue	[]ScheduledJob
}

type MsgRegisterScheduledJob struct {
	Authority	string
	Job		ScheduledJob
}

type MsgPauseScheduledJob struct {
	Authority	string
	OwnerModule	string
	JobID		string
}

type MsgResumeScheduledJob struct {
	Authority	string
	OwnerModule	string
	JobID		string
}

type MsgCancelScheduledJob struct {
	Authority	string
	OwnerModule	string
	JobID		string
}

type MsgExecuteDueJobs struct {
	Authority	string
	CurrentHeight	uint64
}

func DefaultSchedulerParams() SchedulerParams {
	return SchedulerParams{
		MaxJobsPerBlock:	25,
		MaxSchedulerGas:	2_000_000,
		MaxGasPerJob:		100_000,
		HistoryRetention:	200,
		AuthorizedModules: []string{
			"aetracore",
			"avm-dex-contract",
			"epoch",
			"fees",
			"load",
			"mesh",
			"networking",
			"payments",
			"routing",
			"contract-assets",
			"zones",
		},
	}
}

func EmptySchedulerState() SchedulerState {
	return SchedulerState{Jobs: []ScheduledJob{}, History: []JobHistoryRecord{}}
}

func (p SchedulerParams) Normalize() SchedulerParams {
	out := p
	out.AuthorizedModules = normalizeStrings(out.AuthorizedModules)
	return out
}

func (p SchedulerParams) Validate() error {
	p = p.Normalize()
	if p.MaxJobsPerBlock == 0 {
		return errors.New("scheduler max jobs per block must be positive")
	}
	if p.MaxSchedulerGas == 0 {
		return errors.New("scheduler max gas must be positive")
	}
	if p.MaxGasPerJob == 0 {
		return errors.New("scheduler max gas per job must be positive")
	}
	if p.MaxGasPerJob > p.MaxSchedulerGas {
		return errors.New("scheduler max gas per job must not exceed total scheduler gas")
	}
	if len(p.AuthorizedModules) == 0 {
		return errors.New("scheduler requires authorized modules")
	}
	previous := ""
	for i, module := range p.AuthorizedModules {
		if module == "" {
			return errors.New("scheduler authorized module must be non-empty")
		}
		if i > 0 && module <= previous {
			return errors.New("scheduler authorized modules must be sorted and unique")
		}
		previous = module
	}
	return nil
}

func (p SchedulerParams) IsAuthorizedModule(module string) bool {
	module = strings.TrimSpace(module)
	for _, allowed := range p.Normalize().AuthorizedModules {
		if allowed == module {
			return true
		}
	}
	return false
}

func (j ScheduledJob) Normalize() ScheduledJob {
	out := j
	out.ID = strings.TrimSpace(out.ID)
	out.OwnerModule = strings.TrimSpace(out.OwnerModule)
	out.Type = strings.TrimSpace(out.Type)
	out.Payload = append([]byte(nil), out.Payload...)
	return out
}

func (j ScheduledJob) Validate(params SchedulerParams) error {
	params = params.Normalize()
	if err := params.Validate(); err != nil {
		return err
	}
	j = j.Normalize()
	if j.ID == "" {
		return errors.New("scheduler job id is required")
	}
	if j.OwnerModule == "" {
		return errors.New("scheduler job owner module is required")
	}
	if !params.IsAuthorizedModule(j.OwnerModule) {
		return fmt.Errorf("scheduler owner module %q is not authorized", j.OwnerModule)
	}
	if !IsJobType(j.Type) {
		return fmt.Errorf("scheduler job type %q is invalid", j.Type)
	}
	if j.NextExecutionHeight == 0 && !j.Cancelled {
		return errors.New("scheduler next execution height must be positive")
	}
	if j.MaxGas == 0 {
		return errors.New("scheduler job max gas must be positive")
	}
	if j.MaxGas > params.MaxGasPerJob {
		return errors.New("scheduler job max gas exceeds per-job limit")
	}
	if (j.Type == JobTypePeriodic || j.Type == JobTypeEpoch) && j.Interval == 0 {
		return errors.New("scheduler recurring job interval must be positive")
	}
	if j.Type == JobTypeDelayed && j.Interval != 0 {
		return errors.New("scheduler delayed job interval must be zero")
	}
	if j.RetryPolicy.MaxRetries > 0 && j.RetryPolicy.BackoffInterval == 0 {
		return errors.New("scheduler retry backoff interval must be positive when retries are enabled")
	}
	return nil
}

func (s SchedulerState) Export() SchedulerState {
	out := SchedulerState{
		Jobs:			cloneJobs(s.Jobs),
		History:		cloneHistory(s.History),
		LastExecutionHeight:	s.LastExecutionHeight,
	}
	SortJobs(out.Jobs)
	return out
}

func (s SchedulerState) Validate(params SchedulerParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(s.Jobs))
	previousOrder := ""
	for _, job := range s.Export().Jobs {
		if err := job.Validate(params); err != nil {
			return err
		}
		key := JobKey(job)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate scheduler job %q", key)
		}
		seen[key] = struct{}{}
		orderKey := JobOrderKey(job)
		if previousOrder != "" && orderKey <= previousOrder {
			return errors.New("scheduler jobs must be sorted")
		}
		previousOrder = orderKey
	}
	for _, record := range s.History {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r JobHistoryRecord) Validate() error {
	r.JobID = strings.TrimSpace(r.JobID)
	r.OwnerModule = strings.TrimSpace(r.OwnerModule)
	r.Status = strings.TrimSpace(r.Status)
	if r.JobID == "" {
		return errors.New("scheduler history job id is required")
	}
	if r.OwnerModule == "" {
		return errors.New("scheduler history owner module is required")
	}
	if r.Height == 0 {
		return errors.New("scheduler history height must be positive")
	}
	if r.Status != HistoryStatusSuccess && r.Status != HistoryStatusFailure && r.Status != HistoryStatusSkipped {
		return errors.New("scheduler history status is invalid")
	}
	return nil
}

func IsJobType(value string) bool {
	switch value {
	case JobTypePeriodic, JobTypeDelayed, JobTypeEpoch:
		return true
	default:
		return false
	}
}

func JobKey(job ScheduledJob) string {
	job = job.Normalize()
	return job.OwnerModule + "/" + job.ID
}

func JobOrderKey(job ScheduledJob) string {
	job = job.Normalize()
	return fmt.Sprintf("%020d/%s/%s", job.NextExecutionHeight, job.OwnerModule, job.ID)
}

func SortJobs(jobs []ScheduledJob) {
	sort.SliceStable(jobs, func(i, j int) bool {
		left := jobs[i].Normalize()
		right := jobs[j].Normalize()
		if left.NextExecutionHeight != right.NextExecutionHeight {
			return left.NextExecutionHeight < right.NextExecutionHeight
		}
		if left.OwnerModule != right.OwnerModule {
			return left.OwnerModule < right.OwnerModule
		}
		return left.ID < right.ID
	})
}

func DueJobs(state SchedulerState, height uint64) []ScheduledJob {
	out := make([]ScheduledJob, 0)
	for _, job := range state.Export().Jobs {
		if job.Cancelled || job.Paused {
			continue
		}
		if job.NextExecutionHeight != 0 && job.NextExecutionHeight <= height {
			out = append(out, job)
		}
	}
	SortJobs(out)
	return out
}

func UpsertJob(jobs []ScheduledJob, next ScheduledJob) []ScheduledJob {
	next = next.Normalize()
	nextKey := JobKey(next)
	out := make([]ScheduledJob, 0, len(jobs)+1)
	replaced := false
	for _, existing := range jobs {
		existing = existing.Normalize()
		if JobKey(existing) == nextKey {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, existing)
	}
	if !replaced {
		out = append(out, next)
	}
	SortJobs(out)
	return out
}

func PruneHistory(history []JobHistoryRecord, retention uint32) []JobHistoryRecord {
	out := cloneHistory(history)
	if retention == 0 || uint32(len(out)) <= retention {
		return out
	}
	return out[len(out)-int(retention):]
}

func cloneJobs(jobs []ScheduledJob) []ScheduledJob {
	out := make([]ScheduledJob, len(jobs))
	for i, job := range jobs {
		out[i] = job.Normalize()
	}
	return out
}

func cloneHistory(history []JobHistoryRecord) []JobHistoryRecord {
	out := make([]JobHistoryRecord, len(history))
	copy(out, history)
	return out
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			out = append(out, value)
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
