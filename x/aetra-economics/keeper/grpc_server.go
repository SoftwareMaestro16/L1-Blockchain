package keeper

import (
	"context"
	"errors"

	"github.com/sovereign-l1/l1/x/aetra-economics/types"
)

var _ types.MsgServer = grpcMsgServer{}
var _ types.QueryServer = grpcQueryServer{}

type grpcMsgServer struct{ keeper *Keeper }
type grpcQueryServer struct{ keeper *Keeper }

func NewGRPCMsgServer(k *Keeper) types.MsgServer	{ return grpcMsgServer{keeper: k} }
func NewGRPCQueryServer(k *Keeper) types.QueryServer	{ return grpcQueryServer{keeper: k} }

func (s grpcMsgServer) UpdateEconomicsParams(_ context.Context, msg *types.MsgUpdateEconomicsParams) (*types.MsgUpdateEconomicsParamsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty economics params update request")
	}
	if err := NewMsgServerImpl(s.keeper).UpdateEconomicsParams(*msg); err != nil {
		return nil, err
	}
	return &types.MsgUpdateEconomicsParamsResponse{}, nil
}

func (s grpcMsgServer) ApplyEpochEconomics(_ context.Context, msg *types.MsgApplyEpochEconomics) (*types.MsgApplyEpochEconomicsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty economics epoch request")
	}
	if err := NewMsgServerImpl(s.keeper).ApplyEpochEconomics(*msg); err != nil {
		return nil, err
	}
	return &types.MsgApplyEpochEconomicsResponse{}, nil
}

func (s grpcQueryServer) CurrentInflation(_ context.Context, req *types.QueryCurrentInflationRequest) (*types.QueryCurrentInflationResponse, error) {
	res, err := s.keeper.QueryCurrentInflation(*req)
	return &res, err
}
func (s grpcQueryServer) CurrentBondedRatio(_ context.Context, req *types.QueryCurrentBondedRatioRequest) (*types.QueryCurrentBondedRatioResponse, error) {
	res, err := s.keeper.QueryCurrentBondedRatio(*req)
	return &res, err
}
func (s grpcQueryServer) EstimatedAPR(_ context.Context, req *types.QueryEstimatedAPRRequest) (*types.QueryEstimatedAPRResponse, error) {
	res, err := s.keeper.QueryEstimatedAPR(*req)
	return &res, err
}
func (s grpcQueryServer) FeeSplitParams(_ context.Context, req *types.QueryFeeSplitParamsRequest) (*types.QueryFeeSplitParamsResponse, error) {
	res, err := s.keeper.QueryFeeSplitParams(*req)
	return &res, err
}
func (s grpcQueryServer) BurnedSupply(_ context.Context, req *types.QueryBurnedSupplyRequest) (*types.QueryBurnedSupplyResponse, error) {
	res, err := s.keeper.QueryBurnedSupply(*req)
	return &res, err
}
func (s grpcQueryServer) TreasuryBalance(_ context.Context, req *types.QueryTreasuryBalanceRequest) (*types.QueryTreasuryBalanceResponse, error) {
	res, err := s.keeper.QueryTreasuryBalance(*req)
	return &res, err
}
func (s grpcQueryServer) EpochRewardSummary(_ context.Context, req *types.QueryEpochRewardSummaryRequest) (*types.QueryEpochRewardSummaryResponse, error) {
	res, err := s.keeper.QueryEpochRewardSummary(*req)
	return &res, err
}
