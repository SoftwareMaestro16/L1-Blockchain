package keeper

import (
	"context"
	"encoding/json"

	"github.com/sovereign-l1/l1/x/delegator-protection/types"
	delegatorprotectionpb "github.com/sovereign-l1/l1/x/delegator-protection/types/delegatorprotectionpb"
)

var _ delegatorprotectionpb.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) delegatorprotectionpb.MsgServer	{ return msgServer{Keeper: k} }

func (m msgServer) SubmitDelegatorProtectionClaim(ctx context.Context, msg *delegatorprotectionpb.MsgSubmitDelegatorProtectionClaim) (*delegatorprotectionpb.MsgSubmitDelegatorProtectionClaimResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	loss, err := parseInt(msg.LossAmount)
	if err != nil {
		return nil, err
	}
	requested, err := parseInt(msg.RequestedPayout)
	if err != nil {
		return nil, err
	}
	next, claim, err := types.ApplySubmitDelegatorProtectionClaim(state, types.MsgSubmitDelegatorProtectionClaim{Delegator: msg.Delegator, Validator: msg.Validator, LossAmount: loss, RequestedPayout: requested, EligibilityHash: msg.EligibilityHash, Reason: msg.Reason, Epoch: msg.Epoch, Height: msg.Height})
	if err != nil {
		return nil, err
	}
	out, err := mustJSON(claim)
	if err != nil {
		return nil, err
	}
	return &delegatorprotectionpb.MsgSubmitDelegatorProtectionClaimResponse{ClaimJson: out}, m.SetState(ctx, next)
}

func (m msgServer) ApproveDelegatorProtectionClaim(ctx context.Context, msg *delegatorprotectionpb.MsgApproveDelegatorProtectionClaim) (*delegatorprotectionpb.MsgApproveDelegatorProtectionClaimResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	approved, err := parseInt(msg.ApprovedPayout)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyApproveDelegatorProtectionClaim(state, types.MsgApproveDelegatorProtectionClaim{Authority: msg.Authority, ClaimID: msg.ClaimId, ApprovedPayout: approved, Height: msg.Height})
	if err != nil {
		return nil, err
	}
	return &delegatorprotectionpb.MsgApproveDelegatorProtectionClaimResponse{}, m.SetState(ctx, next)
}

func (m msgServer) RejectDelegatorProtectionClaim(ctx context.Context, msg *delegatorprotectionpb.MsgRejectDelegatorProtectionClaim) (*delegatorprotectionpb.MsgRejectDelegatorProtectionClaimResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	next, err := types.ApplyRejectDelegatorProtectionClaim(state, types.MsgRejectDelegatorProtectionClaim{Authority: msg.Authority, ClaimID: msg.ClaimId, Reason: msg.Reason, Height: msg.Height})
	if err != nil {
		return nil, err
	}
	return &delegatorprotectionpb.MsgRejectDelegatorProtectionClaimResponse{}, m.SetState(ctx, next)
}

func (m msgServer) ClaimDelegatorCompensation(ctx context.Context, msg *delegatorprotectionpb.MsgClaimDelegatorCompensation) (*delegatorprotectionpb.MsgClaimDelegatorCompensationResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	next, payout, err := types.ApplyClaimDelegatorCompensation(state, types.MsgClaimDelegatorCompensation{Delegator: msg.Delegator, ClaimID: msg.ClaimId, Epoch: msg.Epoch, Height: msg.Height})
	if err != nil {
		return nil, err
	}
	out, err := mustJSON(payout)
	if err != nil {
		return nil, err
	}
	return &delegatorprotectionpb.MsgClaimDelegatorCompensationResponse{PayoutJson: out}, m.SetState(ctx, next)
}

func (m msgServer) UpdateProtectionParams(ctx context.Context, msg *delegatorprotectionpb.MsgUpdateProtectionParams) (*delegatorprotectionpb.MsgUpdateProtectionParamsResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	var params types.ProtectionParams
	if err := json.Unmarshal([]byte(msg.ParamsJson), &params); err != nil {
		return nil, err
	}
	next, err := types.ApplyUpdateProtectionParams(state, types.MsgUpdateProtectionParams{Authority: msg.Authority, Params: params})
	if err != nil {
		return nil, err
	}
	return &delegatorprotectionpb.MsgUpdateProtectionParamsResponse{}, m.SetState(ctx, next)
}
