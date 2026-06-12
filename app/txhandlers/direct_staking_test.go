package txhandlers

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/app/stakingpolicy"
)

type msgTx struct {
	msgs []sdk.Msg
}

func (tx msgTx) GetMsgs() []sdk.Msg	{ return tx.msgs }

func (tx msgTx) GetMsgsV2() ([]protov2.Message, error)	{ return nil, nil }

func TestRejectDirectUserStakingDecorator(t *testing.T) {
	amount := sdk.NewInt64Coin("naet", 10)
	tests := []struct {
		name	string
		msg	sdk.Msg
	}{
		{
			name:	"delegate",
			msg:	stakingtypes.NewMsgDelegate("AE1", "AE2", amount),
		},
		{
			name:	"redelegate",
			msg:	stakingtypes.NewMsgBeginRedelegate("AE1", "AE2", "AE3", amount),
		},
		{
			name:	"undelegate",
			msg:	stakingtypes.NewMsgUndelegate("AE1", "AE2", amount),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				called = true
				return ctx, nil
			}
			_, err := RejectDirectUserStakingDecorator(next)(sdk.Context{}, msgTx{msgs: []sdk.Msg{tc.msg}}, false)
			require.ErrorContains(t, err, stakingpolicy.DirectUserDelegationDisabledMessage)
			require.False(t, called)
		})
	}
}

func TestRejectDirectUserStakingDecoratorAllowsCreateValidator(t *testing.T) {
	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}
	_, err := RejectDirectUserStakingDecorator(next)(sdk.Context{}, msgTx{msgs: []sdk.Msg{&stakingtypes.MsgCreateValidator{}}}, false)
	require.NoError(t, err)
	require.True(t, called)
}

func TestRejectDirectUserStakingDecoratorAllowsValidatorSelfBond(t *testing.T) {
	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}
	operator := aeFromBytesForAnteTest(t, bytes.Repeat([]byte{0x44}, 20))
	msg := stakingtypes.NewMsgDelegate(operator, operator, sdk.NewInt64Coin("naet", 10))

	_, err := RejectDirectUserStakingDecorator(next)(sdk.Context{}, msgTx{msgs: []sdk.Msg{msg}}, false)
	require.NoError(t, err)
	require.True(t, called)
}

func aeFromBytesForAnteTest(t *testing.T, bz []byte) string {
	t.Helper()
	text, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return text
}
