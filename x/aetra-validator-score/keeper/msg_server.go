package keeper

import "github.com/sovereign-l1/l1/x/aetra-validator-score/types"

type MsgServer struct {
	Keeper *Keeper
}

func NewMsgServerImpl(k *Keeper) MsgServer {
	return MsgServer{Keeper: k}
}

func (m MsgServer) UpdateValidatorScoreParams(msg types.MsgUpdateValidatorScoreParams) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	return m.Keeper.SetParams(msg.Params)
}

func (m MsgServer) UpdateValidatorScores(msg types.MsgUpdateValidatorScores) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	_, err := m.Keeper.UpdateScores(msg.Epoch, msg.Metrics)
	return err
}

func (m MsgServer) requireAuthority(authority string) error {
	if authority != m.Keeper.Authority() {
		return types.ErrUnauthorized.Wrap("invalid authority")
	}
	return nil
}
