package keeper

import (
	"context"
	"encoding/json"

	"github.com/sovereign-l1/l1/x/mint-authority/types"
	mintauthoritypb "github.com/sovereign-l1/l1/x/mint-authority/types/mintauthoritypb"
)

var _ mintauthoritypb.MsgServer = msgServer{}

type msgServer struct{ Keeper }

func NewMsgServerImpl(k Keeper) mintauthoritypb.MsgServer	{ return msgServer{Keeper: k} }

func (m msgServer) MintProtocolCoins(ctx context.Context, msg *mintauthoritypb.MsgMintProtocolCoins) (*mintauthoritypb.MsgMintProtocolCoinsResponse, error) {
	amount, err := parseInt(msg.Amount)
	if err != nil {
		return nil, err
	}
	var decision types.EmissionDecision
	if msg.EmissionsDecisionJson != "" {
		if err := json.Unmarshal([]byte(msg.EmissionsDecisionJson), &decision); err != nil {
			return nil, err
		}
	}
	var auth types.ConstitutionEmergencyAuthorization
	if msg.EmergencyAuthorizationJson != "" {
		if err := json.Unmarshal([]byte(msg.EmergencyAuthorizationJson), &auth); err != nil {
			return nil, err
		}
	}
	event, err := m.Keeper.MintProtocolCoins(ctx, types.MsgMintProtocolCoins{Caller: msg.Caller, Recipient: msg.Recipient, Denom: msg.Denom, Amount: amount, Epoch: msg.Epoch, Height: msg.Height, EmissionsDecisionHash: decision.DecisionHash, Emergency: msg.Emergency, ConstitutionDecisionHash: auth.AuthorizationHash}, decision, auth)
	if err != nil {
		return nil, err
	}
	out, err := mustJSON(event)
	if err != nil {
		return nil, err
	}
	return &mintauthoritypb.MsgMintProtocolCoinsResponse{EventJson: out}, nil
}

func (m msgServer) UpdateMintAuthorityParams(ctx context.Context, msg *mintauthoritypb.MsgUpdateMintAuthorityParams) (*mintauthoritypb.MsgUpdateMintAuthorityParamsResponse, error) {
	state, err := m.GetState(ctx)
	if err != nil {
		return nil, err
	}
	var nextState types.MintAuthorityState
	if err := json.Unmarshal([]byte(msg.StateJson), &nextState); err != nil {
		return nil, err
	}
	next, err := types.ApplyUpdateMintAuthorityParams(state, types.MsgUpdateMintAuthorityParams{Authority: msg.Authority, Params: nextState.Params, Registration: nextState.Registration, AllowedCallers: nextState.AllowedCallers, Caps: nextState.Caps})
	if err != nil {
		return nil, err
	}
	return &mintauthoritypb.MsgUpdateMintAuthorityParamsResponse{}, m.SetState(ctx, next)
}
