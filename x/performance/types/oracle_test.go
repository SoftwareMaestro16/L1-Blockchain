package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

func TestPerformanceOracleAggregatesUptime(t *testing.T) {
	state := newPerformanceOracleState(t)
	state = submitReport(t, state, oracleReport(1, "val-a", "rep-a", 90, 100, 1_000, 2_000, 1, 100, 8_000, 10))
	state = submitReport(t, state, oracleReport(1, "val-a", "rep-b", 100, 100, 1_000, 2_000, 0, 100, 10_000, 11))

	finalized, err := ApplyFinalizePerformanceEpoch(state, MsgFinalizePerformanceEpoch{
		Authority:	state.Params.Authority,
		Epoch:		1,
	})
	require.NoError(t, err)
	res, err := QueryValidatorPerformanceOracle(finalized, QueryValidatorPerformanceRequest{Epoch: 1, ValidatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, uint32(9_500), res.Aggregate.UptimeScoreBps)
	require.Equal(t, uint32(9_950), res.Aggregate.MissedBlockScoreBps)
	require.Equal(t, uint32(2), res.Aggregate.ReportCount)
	require.LessOrEqual(t, res.Aggregate.PerformanceScoreBps, postypes.BasisPoints)
	require.NoError(t, CheckPerformanceOracleInvariants(finalized))
}

func TestPerformanceOracleAggregatesLatencyAndResponseTime(t *testing.T) {
	state := newPerformanceOracleState(t)
	state = submitReport(t, state, oracleReport(2, "val-a", "rep-a", 100, 100, 1_000, 2_000, 0, 100, 10_000, 10))
	state = submitReport(t, state, oracleReport(2, "val-a", "rep-b", 100, 100, 2_000, 4_000, 0, 100, 10_000, 11))

	finalized, err := ApplyFinalizePerformanceEpoch(state, MsgFinalizePerformanceEpoch{
		Authority:	state.Params.Authority,
		Epoch:		2,
	})
	require.NoError(t, err)
	res, err := QueryValidatorPerformanceOracle(finalized, QueryValidatorPerformanceRequest{Epoch: 2, ValidatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, uint32(7_000), res.Aggregate.LatencyScoreBps)
	require.Equal(t, uint32(7_000), res.Aggregate.ResponseTimeScoreBps)
	require.NoError(t, res.Aggregate.Validate(finalized.Params))
}

func TestPerformanceOracleRejectsMalformedReport(t *testing.T) {
	state := newPerformanceOracleState(t)
	_, err := ApplySubmitPerformanceReport(state, MsgSubmitPerformanceReport{
		Authority:	state.Params.Authority,
		Report:		oracleReport(1, "val-a", "rep-a", 101, 100, 1_000, 2_000, 0, 100, 9_000, 10),
	})
	require.ErrorContains(t, err, "signed blocks exceed")

	_, err = ApplySubmitPerformanceReport(state, MsgSubmitPerformanceReport{
		Authority:	state.Params.Authority,
		Report:		oracleReport(1, "val-a", "rep-a", 100, 100, state.Params.MaxLatencyMillis+1, 2_000, 0, 100, 9_000, 10),
	})
	require.ErrorContains(t, err, "latency exceeds")

	report := oracleReport(1, "val-a", "rep-a", 100, 100, 1_000, 2_000, 0, 100, 9_000, 10)
	report.Slashable = false
	report.ReportHash = ComputePerformanceReportHash(report)
	_, err = ApplySubmitPerformanceReport(state, MsgSubmitPerformanceReport{
		Authority:	state.Params.Authority,
		Report:		report,
	})
	require.ErrorContains(t, err, "slashable")

	_, err = ApplySubmitPerformanceReport(state, MsgSubmitPerformanceReport{
		Authority:	"wrong",
		Report:		oracleReport(1, "val-a", "rep-a", 100, 100, 1_000, 2_000, 0, 100, 9_000, 10),
	})
	require.ErrorContains(t, err, "requires authority")
}

func TestPerformanceOracleChallengeReport(t *testing.T) {
	state := newPerformanceOracleState(t)
	bad := oracleReport(3, "val-a", "rep-bad", 0, 100, 5_000, 10_000, 100, 100, 0, 10)
	good := oracleReport(3, "val-a", "rep-good", 100, 100, 1_000, 2_000, 0, 100, 10_000, 11)
	state = submitReport(t, state, bad)
	state = submitReport(t, state, good)

	challenged, err := ApplyChallengePerformanceReport(state, MsgChallengePerformanceReport{
		Authority:	state.Params.Authority,
		Epoch:		3,
		ReportID:	bad.ReportID,
		Challenger:	"watcher-1",
		Reason:		"invalid sample",
		Accepted:	true,
	})
	require.NoError(t, err)
	reports, err := QueryPerformanceReportsOracle(challenged, QueryPerformanceReportsRequest{Epoch: 3, ValidatorAddress: "val-a"})
	require.NoError(t, err)
	require.Len(t, reports.Reports, 2)
	require.True(t, reports.Reports[0].Challenged)

	finalized, err := ApplyFinalizePerformanceEpoch(challenged, MsgFinalizePerformanceEpoch{
		Authority:	challenged.Params.Authority,
		Epoch:		3,
	})
	require.NoError(t, err)
	res, err := QueryValidatorPerformanceOracle(finalized, QueryValidatorPerformanceRequest{Epoch: 3, ValidatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, uint32(10_000), res.Aggregate.UptimeScoreBps)
	require.Equal(t, uint32(8_000), res.Aggregate.LatencyScoreBps)
	epoch, err := QueryPerformanceEpochOracle(finalized, QueryPerformanceEpochRequest{Epoch: 3})
	require.NoError(t, err)
	require.Len(t, epoch.Epoch.Challenges, 1)
	require.Equal(t, ChallengeStatusAccepted, epoch.Epoch.Challenges[0].Status)
}

func TestPerformanceOracleDeterministicOrderForEqualReports(t *testing.T) {
	stateA := newPerformanceOracleState(t)
	stateB := newPerformanceOracleState(t)
	a := oracleReport(4, "val-a", "rep-a", 100, 100, 1_000, 2_000, 0, 100, 10_000, 10)
	b := oracleReport(4, "val-a", "rep-b", 100, 100, 1_000, 2_000, 0, 100, 10_000, 11)
	c := oracleReport(4, "val-b", "rep-a", 50, 100, 2_500, 5_000, 50, 100, 5_000, 12)

	for _, report := range []PerformanceReport{c, a, b} {
		stateA = submitReport(t, stateA, report)
	}
	for _, report := range []PerformanceReport{b, c, a} {
		stateB = submitReport(t, stateB, report)
	}

	finalA, err := ApplyFinalizePerformanceEpoch(stateA, MsgFinalizePerformanceEpoch{Authority: stateA.Params.Authority, Epoch: 4})
	require.NoError(t, err)
	finalB, err := ApplyFinalizePerformanceEpoch(stateB, MsgFinalizePerformanceEpoch{Authority: stateB.Params.Authority, Epoch: 4})
	require.NoError(t, err)
	require.Equal(t, finalA.Epochs, finalB.Epochs)
	require.NoError(t, CheckPerformanceOracleInvariants(finalA))
}

func TestPerformanceOracleExportImportDuringAggregation(t *testing.T) {
	state := newPerformanceOracleState(t)
	state = submitReport(t, state, oracleReport(5, "val-a", "rep-a", 100, 100, 1_000, 2_000, 0, 100, 10_000, 10))
	state = submitReport(t, state, oracleReport(5, "val-a", "rep-b", 80, 100, 2_000, 4_000, 20, 100, 8_000, 11))

	exported, err := ExportPerformanceOracleState(state)
	require.NoError(t, err)
	imported, err := ImportPerformanceOracleState(exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported)

	finalized, err := ApplyFinalizePerformanceEpoch(imported, MsgFinalizePerformanceEpoch{
		Authority:	imported.Params.Authority,
		Epoch:		5,
	})
	require.NoError(t, err)
	require.NoError(t, CheckPerformanceOracleInvariants(finalized))
	params := QueryPerformanceParamsOracle(finalized)
	require.Equal(t, finalized.Params, params.Params)
}

func newPerformanceOracleState(t *testing.T) PerformanceOracleState {
	t.Helper()
	params := DefaultPerformanceOracleParams()
	state, err := NewPerformanceOracleState(params)
	require.NoError(t, err)
	return state
}

func submitReport(t *testing.T, state PerformanceOracleState, report PerformanceReport) PerformanceOracleState {
	t.Helper()
	next, err := ApplySubmitPerformanceReport(state, MsgSubmitPerformanceReport{
		Authority:	state.Params.Authority,
		Report:		report,
	})
	require.NoError(t, err)
	return next
}

func oracleReport(epoch uint64, validator string, reporter string, signed uint64, total uint64, latency uint64, response uint64, missed uint64, missedWindow uint64, peer uint32, height uint64) PerformanceReport {
	report := PerformanceReport{
		Epoch:			epoch,
		ValidatorAddress:	validator,
		ReporterAddress:	reporter,
		Source:			ReportSourceObserver,
		UptimeSignedBlocks:	signed,
		UptimeTotalBlocks:	total,
		LatencyMillis:		latency,
		ResponseTimeMillis:	response,
		MissedBlocks:		missed,
		MissedWindowBlocks:	missedWindow,
		PeerScoreBps:		peer,
		SubmittedHeight:	height,
		Slashable:		true,
	}
	report.ReportID = ComputePerformanceReportID(report)
	report.ReportHash = ComputePerformanceReportHash(report)
	return report
}
