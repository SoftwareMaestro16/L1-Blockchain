package keeper

import (
	"context"
	"errors"

	"github.com/sovereign-l1/l1/x/aetra-validator-score/types"
)

var _ types.MsgServer = grpcMsgServer{}
var _ types.QueryServer = grpcQueryServer{}

type grpcMsgServer struct{ keeper *Keeper }
type grpcQueryServer struct{ keeper *Keeper }

func NewGRPCMsgServer(k *Keeper) types.MsgServer	{ return grpcMsgServer{keeper: k} }
func NewGRPCQueryServer(k *Keeper) types.QueryServer	{ return grpcQueryServer{keeper: k} }

func (s grpcMsgServer) UpdateValidatorScoreParams(_ context.Context, msg *types.MsgUpdateValidatorScoreParams) (*types.MsgUpdateValidatorScoreParamsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty validator score params update request")
	}
	if err := NewMsgServerImpl(s.keeper).UpdateValidatorScoreParams(*msg); err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorScoreParamsResponse{}, nil
}
func (s grpcMsgServer) UpdateValidatorScores(_ context.Context, msg *types.MsgUpdateValidatorScores) (*types.MsgUpdateValidatorScoresResponse, error) {
	if msg == nil {
		return nil, errors.New("empty validator score update request")
	}
	if err := NewMsgServerImpl(s.keeper).UpdateValidatorScores(*msg); err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorScoresResponse{}, nil
}

func (s grpcQueryServer) Params(_ context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	res, err := s.keeper.QueryParams(*req)
	return &res, err
}
func (s grpcQueryServer) ValidatorScore(_ context.Context, req *types.QueryValidatorScoreRequest) (*types.QueryValidatorScoreResponse, error) {
	res, err := s.keeper.QueryValidatorScore(*req)
	return &res, err
}
func (s grpcQueryServer) PublicValidatorMetrics(_ context.Context, req *types.QueryPublicValidatorMetricsRequest) (*types.QueryPublicValidatorMetricsResponse, error) {
	res, err := s.keeper.QueryPublicValidatorMetrics(*req)
	return &res, err
}
func (s grpcQueryServer) AllValidatorScores(_ context.Context, req *types.QueryAllValidatorScoresRequest) (*types.QueryAllValidatorScoresResponse, error) {
	res, err := s.keeper.QueryAllValidatorScores(*req)
	return &res, err
}
