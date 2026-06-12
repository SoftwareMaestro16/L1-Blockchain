package stakingpolicy

import (
	"bytes"
	"context"
	"errors"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

const DirectUserDelegationDisabledMessage = "direct user delegation to validators is disabled; use official liquid staking pool deposit"

const (
	nominatorPoolModule		= "nominator-pool"
	singleNominatorPoolModule	= "single-nominator-pool"
)

type DirectDelegationPolicy struct {
	DirectUserValidatorDelegation string
}

type PoolOnlyMsgServer struct {
	inner stakingtypes.MsgServer
}

func NewPoolOnlyMsgServer(inner stakingtypes.MsgServer) PoolOnlyMsgServer {
	return PoolOnlyMsgServer{inner: inner}
}

func DirectUserDelegationDisabledError() error {
	return errors.New(DirectUserDelegationDisabledMessage)
}

func DefaultDirectDelegationPolicy() DirectDelegationPolicy {
	return DirectDelegationPolicy{
		DirectUserValidatorDelegation: appparams.DirectUserDelegationDisabled,
	}
}

func ValidateDelegate(policy DirectDelegationPolicy, msg *stakingtypes.MsgDelegate) error {
	if msg == nil {
		return DirectUserDelegationDisabledError()
	}
	if !directUserDelegationDisabled(policy) {
		return nil
	}
	if IsValidatorSelfBond(msg.DelegatorAddress, msg.ValidatorAddress) {
		return nil
	}
	if IsNominatorPoolControlledDelegator(msg.DelegatorAddress) {
		return nil
	}
	return DirectUserDelegationDisabledError()
}

func ValidateBeginRedelegate(policy DirectDelegationPolicy, msg *stakingtypes.MsgBeginRedelegate) error {
	if msg == nil {
		return DirectUserDelegationDisabledError()
	}
	if !directUserDelegationDisabled(policy) {
		return nil
	}
	return DirectUserDelegationDisabledError()
}

func ValidateUndelegate(policy DirectDelegationPolicy, msg *stakingtypes.MsgUndelegate) error {
	if msg == nil {
		return DirectUserDelegationDisabledError()
	}
	if !directUserDelegationDisabled(policy) {
		return nil
	}
	return DirectUserDelegationDisabledError()
}

func IsValidatorSelfBond(delegatorAddress, validatorAddress string) bool {
	delegator, err := addressing.Parse(delegatorAddress)
	if err != nil {
		return false
	}
	validator, err := addressing.Parse(validatorAddress)
	if err != nil {
		return false
	}
	delegatorRaw, err := addressing.ToRawPayload(delegator)
	if err != nil {
		return false
	}
	validatorRaw, err := addressing.ToRawPayload(validator)
	if err != nil {
		return false
	}
	return bytes.Equal(delegatorRaw, validatorRaw)
}

func IsNominatorPoolControlledDelegator(delegatorAddress string) bool {
	systemAddress, found := addressing.SystemAddressByText(delegatorAddress)
	if !found {
		return false
	}
	return systemAddress.ModuleName == nominatorPoolModule || systemAddress.ModuleName == singleNominatorPoolModule
}

func directUserDelegationDisabled(policy DirectDelegationPolicy) bool {
	return policy.DirectUserValidatorDelegation == "" ||
		policy.DirectUserValidatorDelegation == appparams.DirectUserDelegationDisabled
}

func (s PoolOnlyMsgServer) CreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	return s.inner.CreateValidator(ctx, msg)
}

func (s PoolOnlyMsgServer) EditValidator(ctx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	return s.inner.EditValidator(ctx, msg)
}

func (s PoolOnlyMsgServer) Delegate(ctx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	if err := ValidateDelegate(DefaultDirectDelegationPolicy(), msg); err != nil {
		return nil, err
	}
	return s.inner.Delegate(ctx, msg)
}

func (s PoolOnlyMsgServer) BeginRedelegate(ctx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	if err := ValidateBeginRedelegate(DefaultDirectDelegationPolicy(), msg); err != nil {
		return nil, err
	}
	return s.inner.BeginRedelegate(ctx, msg)
}

func (s PoolOnlyMsgServer) Undelegate(ctx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	if err := ValidateUndelegate(DefaultDirectDelegationPolicy(), msg); err != nil {
		return nil, err
	}
	return s.inner.Undelegate(ctx, msg)
}

func (s PoolOnlyMsgServer) CancelUnbondingDelegation(ctx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	return s.inner.CancelUnbondingDelegation(ctx, msg)
}

func (s PoolOnlyMsgServer) UpdateParams(ctx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	return s.inner.UpdateParams(ctx, msg)
}
