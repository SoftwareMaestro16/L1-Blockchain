package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDefaultImplementationSequencingPhase0AndPhase1Ready(t *testing.T) {
	report := BuildImplementationSequencingReport(nil, nil)
	require.True(t, report.Passed, report.Failed)
	require.True(t, report.ReadyForPhase1)
	require.True(t, report.ReadyForIncentiveChanges)
	require.Len(t, report.Phases, 2)
	require.Len(t, report.Phases[0].Tasks, 7)
	require.Len(t, report.Phases[1].Tasks, 7)
	require.Equal(t, SequencingStatusReady, report.Phases[0].Status)
	require.Equal(t, SequencingStatusReady, report.Phases[1].Status)
	require.Contains(t, report.GovernanceSummary, "phase0=ready")
	require.Contains(t, report.GovernanceSummary, "phase1=ready")
}

func TestImplementationSequencingBlocksPhase1WhenMeasurementMissing(t *testing.T) {
	phase0 := DefaultPhase0SequencingTasks()
	phase0[0].Implemented = false
	phase0[0].Reconciled = false

	report := BuildImplementationSequencingReport(phase0, nil)
	require.False(t, report.Passed)
	require.False(t, report.ReadyForPhase1)
	require.False(t, report.ReadyForIncentiveChanges)
	require.Equal(t, SequencingStatusBlocked, report.Phases[0].Status)
	require.Contains(t, report.Failed, EconomicSequencePhase0+":"+SequencingTaskNetIssuanceAccounting+"_not_ready")
	require.Contains(t, report.Failed, EconomicSequencePhase0+":accounting_reconciles_per_epoch:accounting_task_not_reconciled")
}

func TestImplementationSequencingBlocksIncentiveChangesWhenSafetyPathDisconnected(t *testing.T) {
	phase1 := DefaultPhase1SequencingTasks()
	for i := range phase1 {
		if phase1[i].Name == SequencingTaskMempoolExecutionFeeValidationAlignment {
			phase1[i].Reconciled = false
		}
	}

	report := BuildImplementationSequencingReport(nil, phase1)
	require.False(t, report.Passed)
	require.True(t, report.ReadyForPhase1)
	require.False(t, report.ReadyForIncentiveChanges)
	require.Equal(t, SequencingStatusBlocked, report.Phases[1].Status)
	require.Contains(t, report.Failed, EconomicSequencePhase1+":"+SequencingTaskMempoolExecutionFeeValidationAlignment+"_not_ready")
	require.Contains(t, report.Failed, EconomicSequencePhase1+":controllers_bounded_and_aligned:controller_or_fee_validation_not_reconciled")
}

func TestImplementationSequencingRejectsDuplicateAndMissingTasks(t *testing.T) {
	phase0 := DefaultPhase0SequencingTasks()
	phase0 = append(phase0[:1], phase0[2:]...)
	phase0 = append(phase0, phase0[0])

	report := BuildImplementationSequencingReport(phase0, nil)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, EconomicSequencePhase0+":"+SequencingTaskCumulativeBurnAccounting+"_missing")
	require.Contains(t, report.Failed, EconomicSequencePhase0+":"+SequencingTaskNetIssuanceAccounting+"_duplicate")
}

func TestValidatorRewardPerVotingPowerTelemetry(t *testing.T) {
	out, err := ComputeValidatorRewardPerVotingPowerTelemetry([]ValidatorRewardTelemetrySample{
		{ValidatorID: "val-b", VotingPowerBps: 500, RewardNaet: sdkmath.NewInt(4_000)},
		{ValidatorID: "val-a", VotingPowerBps: 2_000, RewardNaet: sdkmath.NewInt(10_000)},
		{ValidatorID: "val-c", VotingPowerBps: 0, RewardNaet: sdkmath.NewInt(700)},
	})
	require.NoError(t, err)
	require.Len(t, out, 3)
	require.Equal(t, "val-a", out[0].ValidatorID)
	require.Equal(t, sdkmath.NewInt(5), out[0].RewardPerBpsNaet)
	require.Equal(t, "val-b", out[1].ValidatorID)
	require.Equal(t, sdkmath.NewInt(8), out[1].RewardPerBpsNaet)
	require.Equal(t, "val-c", out[2].ValidatorID)
	require.True(t, out[2].RewardPerBpsNaet.IsZero())

	_, err = ComputeValidatorRewardPerVotingPowerTelemetry([]ValidatorRewardTelemetrySample{{ValidatorID: "bad", VotingPowerBps: BasisPoints + 1, RewardNaet: sdkmath.OneInt()}})
	require.ErrorContains(t, err, "voting_power_bps")
	_, err = ComputeValidatorRewardPerVotingPowerTelemetry([]ValidatorRewardTelemetrySample{{ValidatorID: "bad", VotingPowerBps: 1, RewardNaet: sdkmath.NewInt(-1)}})
	require.ErrorContains(t, err, "reward_naet must not be negative")
	_, err = ComputeValidatorRewardPerVotingPowerTelemetry([]ValidatorRewardTelemetrySample{{ValidatorID: "dup", VotingPowerBps: 1, RewardNaet: sdkmath.OneInt()}, {ValidatorID: "dup", VotingPowerBps: 1, RewardNaet: sdkmath.OneInt()}})
	require.ErrorContains(t, err, "duplicate validator_id")
}

func TestValidateEconomicSequencingMetricsReconcilesPhase0AndPhase1(t *testing.T) {
	criterion := ValidateEconomicSequencingMetrics(EconomicSequencingMetricInput{
		EpochReport:              EpochEconomicReport{Reconciled: true},
		BurnSupply:               BurnSupplyQueryOutput{CumulativeBurnedNaet: sdkmath.NewInt(1_000), RecentBurnedNaet: sdkmath.NewInt(100)},
		FeeAllocation:            FeeAllocationBuckets{SumsExactly: true},
		SecurityReport:           EconomicSecurityEpochReport{Passed: true},
		StateGrowth:              StateGrowthTelemetryOutput{EpochID: 1, BlockHeight: 10},
		FeeOptimizer:             FeeMarketOptimizerOutput{Validation: FeeValidationResult{MempoolExecutionAligned: true}},
		BurnFeeInvariant:         BurnAccountingInvariantReport{Passed: true},
		BurnSlashingInvariant:    BurnAccountingInvariantReport{Passed: true},
		AdaptiveInflation:        AdaptiveInflationEpochReport{Reconciled: true},
		ValidatorRewardTelemetry: []ValidatorRewardTelemetrySample{{ValidatorID: "val-a", VotingPowerBps: 1_000, RewardNaet: sdkmath.NewInt(5_000)}},
	})
	require.True(t, criterion.Met, criterion.FailedReason)

	criterion = ValidateEconomicSequencingMetrics(EconomicSequencingMetricInput{
		EpochReport:              EpochEconomicReport{Reconciled: false},
		BurnSupply:               BurnSupplyQueryOutput{CumulativeBurnedNaet: sdkmath.NewInt(-1)},
		FeeAllocation:            FeeAllocationBuckets{SumsExactly: false},
		SecurityReport:           EconomicSecurityEpochReport{Failed: []string{"fund_movement_mismatch"}},
		StateGrowth:              StateGrowthTelemetryOutput{},
		FeeOptimizer:             FeeMarketOptimizerOutput{Validation: FeeValidationResult{MempoolExecutionAligned: false}},
		BurnFeeInvariant:         BurnAccountingInvariantReport{Passed: false},
		BurnSlashingInvariant:    BurnAccountingInvariantReport{Passed: false},
		AdaptiveInflation:        AdaptiveInflationEpochReport{Reconciled: false},
		ValidatorRewardTelemetry: []ValidatorRewardTelemetrySample{{ValidatorID: "", VotingPowerBps: 1, RewardNaet: sdkmath.OneInt()}},
	})
	require.False(t, criterion.Met)
	require.Contains(t, criterion.FailedReason, "epoch_economic_report_not_reconciled")
	require.Contains(t, criterion.FailedReason, "fee_buckets_do_not_sum")
	require.Contains(t, criterion.FailedReason, "validator_reward_telemetry_invalid")
}
