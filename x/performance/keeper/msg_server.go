package keeper

import (
	"context"
	"encoding/json"

	"github.com/sovereign-l1/l1/x/performance/types"
	performancepb "github.com/sovereign-l1/l1/x/performance/types/performancepb"
)

var _ performancepb.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) performancepb.MsgServer	{ return msgServer{Keeper: k} }

func (m msgServer) SubmitPerformanceReport(ctx context.Context, msg *performancepb.MsgSubmitPerformanceReport) (*performancepb.MsgSubmitPerformanceReportResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	var report types.PerformanceReport
	if err := json.Unmarshal([]byte(msg.ReportJson), &report); err != nil {
		return nil, err
	}
	next, err := types.ApplySubmitPerformanceReport(state, types.MsgSubmitPerformanceReport{Authority: msg.Authority, Report: report})
	if err != nil {
		return nil, err
	}
	return &performancepb.MsgSubmitPerformanceReportResponse{}, m.SetState(ctx, next)
}

func (m msgServer) FinalizePerformanceEpoch(ctx context.Context, msg *performancepb.MsgFinalizePerformanceEpoch) (*performancepb.MsgFinalizePerformanceEpochResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyFinalizePerformanceEpoch(state, types.MsgFinalizePerformanceEpoch{Authority: msg.Authority, Epoch: msg.Epoch})
	if err != nil {
		return nil, err
	}
	epoch, err := types.QueryPerformanceEpochOracle(next, types.QueryPerformanceEpochRequest{Epoch: msg.Epoch})
	if err != nil {
		return nil, err
	}
	out, err := mustJSON(epoch.Epoch)
	if err != nil {
		return nil, err
	}
	return &performancepb.MsgFinalizePerformanceEpochResponse{EpochJson: out}, m.SetState(ctx, next)
}

func (m msgServer) ChallengePerformanceReport(ctx context.Context, msg *performancepb.MsgChallengePerformanceReport) (*performancepb.MsgChallengePerformanceReportResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyChallengePerformanceReport(state, types.MsgChallengePerformanceReport{Authority: msg.Authority, Epoch: msg.Epoch, ReportID: msg.ReportId, Challenger: msg.Challenger, Reason: msg.Reason, Accepted: msg.Accepted})
	if err != nil {
		return nil, err
	}
	return &performancepb.MsgChallengePerformanceReportResponse{}, m.SetState(ctx, next)
}
