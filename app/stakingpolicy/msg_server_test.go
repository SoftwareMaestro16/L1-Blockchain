package stakingpolicy

import (
	"bytes"
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestPoolOnlyMsgServerRejectsDirectUserValidatorMessages(t *testing.T) {
	server := NewPoolOnlyMsgServer(nil)
	amount := sdk.NewInt64Coin("naet", 10)

	_, err := server.Delegate(context.Background(), stakingtypes.NewMsgDelegate("AE1", "AE2", amount))
	require.ErrorContains(t, err, DirectUserDelegationDisabledMessage)

	_, err = server.BeginRedelegate(context.Background(), stakingtypes.NewMsgBeginRedelegate("AE1", "AE2", "AE3", amount))
	require.ErrorContains(t, err, DirectUserDelegationDisabledMessage)

	_, err = server.Undelegate(context.Background(), stakingtypes.NewMsgUndelegate("AE1", "AE2", amount))
	require.ErrorContains(t, err, DirectUserDelegationDisabledMessage)
}

func TestPoolOnlyMsgServerAllowsValidatorSelfBondWhenDirectUserDelegationDisabled(t *testing.T) {
	inner := &recordingStakingMsgServer{}
	server := NewPoolOnlyMsgServer(inner)
	operator := aeFromBytesForPolicyTest(t, bytesOf(0x11))
	amount := sdk.NewInt64Coin("naet", 10)

	_, err := server.Delegate(context.Background(), stakingtypes.NewMsgDelegate(operator, operator, amount))
	require.NoError(t, err)
	require.Equal(t, operator, inner.delegate.DelegatorAddress)
	require.Equal(t, appparams.DirectUserDelegationDisabled, DefaultDirectDelegationPolicy().DirectUserValidatorDelegation)
}

func TestValidateDelegateRejectsOrdinaryUserWhenGovernanceParamDisabled(t *testing.T) {
	user := aeFromBytesForPolicyTest(t, bytesOf(0x22))
	validator := aeFromBytesForPolicyTest(t, bytesOf(0x33))
	msg := stakingtypes.NewMsgDelegate(user, validator, sdk.NewInt64Coin("naet", 10))

	err := ValidateDelegate(DefaultDirectDelegationPolicy(), msg)
	require.ErrorContains(t, err, DirectUserDelegationDisabledMessage)
	require.Equal(t, appparams.DirectUserDelegationDisabled, DefaultDirectDelegationPolicy().DirectUserValidatorDelegation)

	err = ValidateDelegate(DirectDelegationPolicy{}, msg)
	require.ErrorContains(t, err, DirectUserDelegationDisabledMessage)
}

type recordingStakingMsgServer struct {
	stakingtypes.UnimplementedMsgServer
	delegate	*stakingtypes.MsgDelegate
}

func (s *recordingStakingMsgServer) Delegate(_ context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	s.delegate = msg
	return &stakingtypes.MsgDelegateResponse{}, nil
}

func bytesOf(value byte) []byte {
	return bytes.Repeat([]byte{value}, 20)
}

func aeFromBytesForPolicyTest(t *testing.T, bz []byte) string {
	t.Helper()
	text, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return text
}
