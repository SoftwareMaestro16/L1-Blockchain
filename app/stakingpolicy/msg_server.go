package stakingpolicy

import (
	"context"
	"errors"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const DirectUserDelegationDisabledMessage = "direct user delegation to validators is disabled; use official liquid staking pool deposit"

type PoolOnlyMsgServer struct {
	inner stakingtypes.MsgServer
}

func NewPoolOnlyMsgServer(inner stakingtypes.MsgServer) PoolOnlyMsgServer {
	return PoolOnlyMsgServer{inner: inner}
}

func DirectUserDelegationDisabledError() error {
	return errors.New(DirectUserDelegationDisabledMessage)
}

func (s PoolOnlyMsgServer) CreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	return s.inner.CreateValidator(ctx, msg)
}

func (s PoolOnlyMsgServer) EditValidator(ctx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	return s.inner.EditValidator(ctx, msg)
}

func (s PoolOnlyMsgServer) Delegate(context.Context, *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	return nil, DirectUserDelegationDisabledError()
}

func (s PoolOnlyMsgServer) BeginRedelegate(context.Context, *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	return nil, DirectUserDelegationDisabledError()
}

func (s PoolOnlyMsgServer) Undelegate(context.Context, *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	return nil, DirectUserDelegationDisabledError()
}

func (s PoolOnlyMsgServer) CancelUnbondingDelegation(ctx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	return s.inner.CancelUnbondingDelegation(ctx, msg)
}

func (s PoolOnlyMsgServer) UpdateParams(ctx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	return s.inner.UpdateParams(ctx, msg)
}
