package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/observability"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func (k Keeper) AnteHandlerDecorator(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			observability.RecordFeeRejected("missing_fee_tx")
			return ctx, types.ErrInvalidFee.Wrap("transaction must expose fees")
		}
		fees := feeTx.GetFee()
		enforceMin := !simulate && ctx.BlockHeight() > 0
		if err := k.ValidateFees(ctx, fees, enforceMin); err != nil {
			observability.RecordFeeRejected("invalid_fee")
			return ctx, err
		}
		newCtx, err := next(ctx, tx, simulate)
		if err != nil || simulate {
			if err != nil {
				observability.RecordModuleError(types.ModuleName, "ante", "next_error")
			}
			return newCtx, err
		}
		if err := k.RecordCollectedFees(newCtx, fees); err != nil {
			observability.RecordModuleError(types.ModuleName, "record_collected_fees", "error")
			return newCtx, err
		}
		observability.RecordFeeAccepted()
		return newCtx, nil
	}
}
