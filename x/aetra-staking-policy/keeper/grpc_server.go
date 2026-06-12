package keeper

import (
	"context"
	"errors"

	"github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
)

var _ types.MsgServer = grpcMsgServer{}
var _ types.QueryServer = grpcQueryServer{}

type grpcMsgServer struct{ keeper *Keeper }
type grpcQueryServer struct{ keeper *Keeper }

func NewGRPCMsgServer(k *Keeper) types.MsgServer	{ return grpcMsgServer{keeper: k} }
func NewGRPCQueryServer(k *Keeper) types.QueryServer	{ return grpcQueryServer{keeper: k} }

func (s grpcMsgServer) UpdateStakingPolicyParams(_ context.Context, msg *types.MsgUpdateStakingPolicyParams) (*types.MsgUpdateStakingPolicyParamsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty staking policy params update request")
	}
	if err := NewMsgServerImpl(s.keeper).UpdateStakingPolicyParams(*msg); err != nil {
		return nil, err
	}
	return &types.MsgUpdateStakingPolicyParamsResponse{}, nil
}
func (s grpcMsgServer) RegisterValidatorIdentity(_ context.Context, msg *types.MsgRegisterValidatorIdentity) (*types.MsgRegisterValidatorIdentityResponse, error) {
	if msg == nil {
		return nil, errors.New("empty validator identity request")
	}
	if err := NewMsgServerImpl(s.keeper).RegisterValidatorIdentity(*msg); err != nil {
		return nil, err
	}
	return &types.MsgRegisterValidatorIdentityResponse{}, nil
}
func (s grpcMsgServer) AcknowledgeConcentrationWarning(_ context.Context, msg *types.MsgAcknowledgeConcentrationWarning) (*types.MsgAcknowledgeConcentrationWarningResponse, error) {
	if msg == nil {
		return nil, errors.New("empty concentration warning acknowledgement")
	}
	if err := NewMsgServerImpl(s.keeper).AcknowledgeConcentrationWarning(*msg); err != nil {
		return nil, err
	}
	return &types.MsgAcknowledgeConcentrationWarningResponse{}, nil
}

func (s grpcQueryServer) Params(_ context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	res, err := s.keeper.QueryParams(*req)
	return &res, err
}
func (s grpcQueryServer) ValidatorEffectivePower(_ context.Context, req *types.QueryValidatorEffectivePowerRequest) (*types.QueryValidatorEffectivePowerResponse, error) {
	res, err := s.keeper.QueryValidatorEffectivePower(*req)
	return &res, err
}
func (s grpcQueryServer) ValidatorStake(_ context.Context, req *types.QueryValidatorStakeRequest) (*types.QueryValidatorStakeResponse, error) {
	res, err := s.keeper.QueryValidatorStake(*req)
	return &res, err
}
func (s grpcQueryServer) TopNConcentration(_ context.Context, req *types.QueryTopNConcentrationRequest) (*types.QueryTopNConcentrationResponse, error) {
	res, err := s.keeper.QueryTopNConcentration(*req)
	return &res, err
}
func (s grpcQueryServer) ValidatorRewardMultiplier(_ context.Context, req *types.QueryValidatorRewardMultiplierRequest) (*types.QueryValidatorRewardMultiplierResponse, error) {
	res, err := s.keeper.QueryValidatorRewardMultiplier(*req)
	return &res, err
}
func (s grpcQueryServer) DelegationWarningStatus(_ context.Context, req *types.QueryDelegationWarningStatusRequest) (*types.QueryDelegationWarningStatusResponse, error) {
	res, err := s.keeper.QueryDelegationWarningStatus(*req)
	return &res, err
}
