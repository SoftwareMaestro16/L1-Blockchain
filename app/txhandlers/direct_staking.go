package txhandlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/app/stakingpolicy"
)

func RejectDirectUserStakingDecorator(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		for _, msg := range tx.GetMsgs() {
			switch msg.(type) {
			case *stakingtypes.MsgDelegate, *stakingtypes.MsgBeginRedelegate, *stakingtypes.MsgUndelegate:
				return ctx, stakingpolicy.DirectUserDelegationDisabledError()
			}
		}
		return next(ctx, tx, simulate)
	}
}
