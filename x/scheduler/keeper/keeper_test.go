package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
)

func TestDefaultGenesisIsDisabledAndValid(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Jobs)
	require.NotZero(t, gs.SchedulerParams.MaxJobsPerBlock)
}

func TestOnlyAuthorizedModulesCanRegisterJobs(t *testing.T) {
	k := enabledKeeper(t)
	job := testJob("unauthorized", "job", schedulertypes.JobTypeDelayed, 10)

	err := k.RegisterScheduledJob(schedulertypes.MsgRegisterScheduledJob{
		Authority:	prototype.DefaultAuthority,
		Job:		job,
	})
	require.ErrorContains(t, err, "not authorized")
}

func TestPeriodicJobExecutesAtCorrectHeight(t *testing.T) {
	k := enabledKeeper(t)
	executed := make([]uint64, 0)
	require.NoError(t, k.RegisterJobHandler("aetracore", func(job schedulertypes.ScheduledJob, height uint64) schedulertypes.JobExecutionResult {
		executed = append(executed, height)
		return schedulertypes.JobExecutionResult{Success: true, GasUsed: 20}
	}))
	require.NoError(t, k.RegisterScheduledJob(register(testJob("aetracore", "periodic", schedulertypes.JobTypePeriodic, 10))))

	result, err := k.ExecuteDueJobs(execute(9))
	require.NoError(t, err)
	require.Zero(t, result.ExecutedJobs)

	result, err = k.ExecuteDueJobs(execute(10))
	require.NoError(t, err)
	require.Equal(t, uint32(1), result.ExecutedJobs)
	require.Equal(t, []uint64{10}, executed)
	job, found, err := k.ScheduledJob("aetracore", "periodic")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(15), job.NextExecutionHeight)
	require.Equal(t, uint64(1), job.ExecutionCount)
}

func TestDelayedJobExecutesOnce(t *testing.T) {
	k := enabledKeeper(t)
	calls := 0
	require.NoError(t, k.RegisterJobHandler("aetracore", func(job schedulertypes.ScheduledJob, height uint64) schedulertypes.JobExecutionResult {
		calls++
		return schedulertypes.JobExecutionResult{Success: true, GasUsed: 10}
	}))
	job := testJob("aetracore", "delayed", schedulertypes.JobTypeDelayed, 7)
	job.Interval = 0
	require.NoError(t, k.RegisterScheduledJob(register(job)))

	result, err := k.ExecuteDueJobs(execute(7))
	require.NoError(t, err)
	require.Equal(t, uint32(1), result.ExecutedJobs)
	result, err = k.ExecuteDueJobs(execute(8))
	require.NoError(t, err)
	require.Zero(t, result.ExecutedJobs)
	require.Equal(t, 1, calls)
	stored, found, err := k.ScheduledJob("aetracore", "delayed")
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, stored.Cancelled)
}

func TestJobFailureIncrementsFailureCountAndDoesNotPanicBlock(t *testing.T) {
	k := enabledKeeper(t)
	require.NoError(t, k.RegisterJobHandler("aetracore", func(job schedulertypes.ScheduledJob, height uint64) schedulertypes.JobExecutionResult {
		panic("boom")
	}))
	require.NoError(t, k.RegisterScheduledJob(register(testJob("aetracore", "panic", schedulertypes.JobTypePeriodic, 3))))

	result, err := k.ExecuteDueJobs(execute(3))
	require.NoError(t, err)
	require.Equal(t, uint32(1), result.ExecutedJobs)
	require.Len(t, result.History, 1)
	require.Equal(t, schedulertypes.HistoryStatusFailure, result.History[0].Status)
	require.Contains(t, result.History[0].Error, "panicked")
	job, found, err := k.ScheduledJob("aetracore", "panic")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint32(1), job.FailureCount)
	require.Equal(t, uint64(5), job.NextExecutionHeight)
}

func TestJobOrderIsDeterministic(t *testing.T) {
	k := enabledKeeper(t)
	order := make([]string, 0)
	for _, module := range []string{"aetracore", "avm-dex-contract"} {
		require.NoError(t, k.RegisterJobHandler(module, func(job schedulertypes.ScheduledJob, height uint64) schedulertypes.JobExecutionResult {
			order = append(order, schedulertypes.JobKey(job))
			return schedulertypes.JobExecutionResult{Success: true, GasUsed: 1}
		}))
	}
	require.NoError(t, k.RegisterScheduledJob(register(testJob("avm-dex-contract", "b", schedulertypes.JobTypePeriodic, 5))))
	require.NoError(t, k.RegisterScheduledJob(register(testJob("aetracore", "z", schedulertypes.JobTypePeriodic, 5))))
	require.NoError(t, k.RegisterScheduledJob(register(testJob("aetracore", "a", schedulertypes.JobTypePeriodic, 5))))

	_, err := k.ExecuteDueJobs(execute(5))
	require.NoError(t, err)
	require.Equal(t, []string{"aetracore/a", "aetracore/z", "avm-dex-contract/b"}, order)
}

func TestBlockGasLimitAndJobCountAreRespected(t *testing.T) {
	k := enabledKeeper(t)
	params := prototype.TestnetParams()
	schedulerParams := schedulertypes.DefaultSchedulerParams()
	schedulerParams.MaxJobsPerBlock = 2
	schedulerParams.MaxSchedulerGas = 150
	schedulerParams.MaxGasPerJob = 100
	require.NoError(t, k.UpdateParams(prototype.DefaultAuthority, params, schedulerParams))
	require.NoError(t, k.RegisterJobHandler("aetracore", func(job schedulertypes.ScheduledJob, height uint64) schedulertypes.JobExecutionResult {
		return schedulertypes.JobExecutionResult{Success: true, GasUsed: 75}
	}))
	for _, id := range []string{"a", "b", "c"} {
		job := testJob("aetracore", id, schedulertypes.JobTypePeriodic, 10)
		job.MaxGas = 75
		require.NoError(t, k.RegisterScheduledJob(register(job)))
	}

	result, err := k.ExecuteDueJobs(execute(10))
	require.NoError(t, err)
	require.Equal(t, uint32(2), result.ExecutedJobs)
	require.Equal(t, uint32(1), result.SkippedJobs)
	require.Equal(t, uint64(150), result.GasReserved)
	require.LessOrEqual(t, result.GasReserved, schedulerParams.MaxSchedulerGas)
	require.Len(t, result.RemainingDue, 1)
}

func TestExportImportPreservesJobQueueAndHistory(t *testing.T) {
	source := enabledKeeper(t)
	require.NoError(t, source.RegisterJobHandler("aetracore", func(job schedulertypes.ScheduledJob, height uint64) schedulertypes.JobExecutionResult {
		return schedulertypes.JobExecutionResult{Success: true, GasUsed: 9}
	}))
	require.NoError(t, source.RegisterScheduledJob(register(testJob("aetracore", "periodic", schedulertypes.JobTypePeriodic, 4))))
	_, err := source.ExecuteDueJobs(execute(4))
	require.NoError(t, err)

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
}

func enabledKeeper(t *testing.T) Keeper {
	t.Helper()
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))
	return k
}

func register(job schedulertypes.ScheduledJob) schedulertypes.MsgRegisterScheduledJob {
	return schedulertypes.MsgRegisterScheduledJob{Authority: prototype.DefaultAuthority, Job: job}
}

func execute(height uint64) schedulertypes.MsgExecuteDueJobs {
	return schedulertypes.MsgExecuteDueJobs{Authority: prototype.DefaultAuthority, CurrentHeight: height}
}

func testJob(owner, id, jobType string, nextHeight uint64) schedulertypes.ScheduledJob {
	return schedulertypes.ScheduledJob{
		ID:			id,
		OwnerModule:		owner,
		Type:			jobType,
		NextExecutionHeight:	nextHeight,
		Interval:		5,
		MaxGas:			100,
		RetryPolicy:		schedulertypes.RetryPolicy{MaxRetries: 1, BackoffInterval: 2},
	}
}
