package stakingpolicy

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
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
