package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	performancekeeper "github.com/sovereign-l1/l1/x/performance/keeper"
	"github.com/sovereign-l1/l1/x/performance/types"
	performancepb "github.com/sovereign-l1/l1/x/performance/types/performancepb"
)

func TestNativePerformanceOracleSubmitFinalizeAndExport(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := performancekeeper.NewMsgServerImpl(app.PerformanceKeeper)
	report, err := types.NewPerformanceReport(types.PerformanceReport{
		Epoch:			1,
		ValidatorAddress:	"validator-a",
		ReporterAddress:	"reporter-a",
		Source:			types.ReportSourceObserver,
		UptimeSignedBlocks:	90,
		UptimeTotalBlocks:	100,
		LatencyMillis:		100,
		ResponseTimeMillis:	200,
		MissedBlocks:		1,
		MissedWindowBlocks:	100,
		PeerScoreBps:		9_000,
		SubmittedHeight:	10,
		Slashable:		true,
	})
	require.NoError(t, err)
	reportJSON, err := json.Marshal(report)
	require.NoError(t, err)

	_, err = msgServer.SubmitPerformanceReport(ctx, &performancepb.MsgSubmitPerformanceReport{
		Authority:	app.PerformanceKeeper.Authority(),
		ReportJson:	string(reportJSON),
	})
	require.NoError(t, err)
	finalized, err := msgServer.FinalizePerformanceEpoch(ctx, &performancepb.MsgFinalizePerformanceEpoch{
		Authority:	app.PerformanceKeeper.Authority(),
		Epoch:		1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, finalized.EpochJson)

	exported, err := app.PerformanceKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	imported := l1app.Setup(t, false)
	importedCtx := imported.NewContext(false)
	require.NoError(t, imported.PerformanceKeeper.InitGenesis(importedCtx, *exported))
	roundTrip, err := imported.PerformanceKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}
