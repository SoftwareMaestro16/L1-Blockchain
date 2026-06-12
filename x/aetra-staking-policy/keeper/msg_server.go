package keeper

import (
	"github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
)

type MsgServer struct {
	Keeper *Keeper
}

func NewMsgServerImpl(k *Keeper) MsgServer {
	return MsgServer{Keeper: k}
}

func (m MsgServer) UpdateStakingPolicyParams(msg types.MsgUpdateStakingPolicyParams) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	return m.Keeper.SetParams(msg.Params)
}

func (m MsgServer) RegisterValidatorIdentity(msg types.MsgRegisterValidatorIdentity) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	return m.Keeper.RegisterValidatorIdentity(msg.Identity)
}

func (m MsgServer) AcknowledgeConcentrationWarning(msg types.MsgAcknowledgeConcentrationWarning) error {
	if err := m.requireAuthority(msg.Authority); err != nil {
		return err
	}
	return m.Keeper.AcknowledgeConcentrationWarning(types.WarningAcknowledgement{
		OperatorAddress:	msg.OperatorAddress,
		AcknowledgedAt:		msg.Height,
		Warning:		msg.Warning,
	})
}

func (m MsgServer) requireAuthority(authority string) error {
	if authority != m.Keeper.Authority() {
		return types.ErrUnauthorized.Wrap("invalid authority")
	}
	return nil
}
