package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/observability"
	"github.com/sovereign-l1/l1/x/fees/types"
)

func (k Keeper) AnteHandlerDecorator(next sdk.AnteHandler) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		if isGenesisCreateValidatorTx(ctx, tx) {
			return next(ctx, tx, simulate)
		}
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			observability.RecordFeeRejected("missing_fee_tx")
			return ctx, types.ErrInvalidFee.Wrap("transaction must expose fees")
		}
		fees := feeTx.GetFee()
		if err := k.ValidateTxFees(ctx, fees); err != nil {
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

func isGenesisCreateValidatorTx(ctx sdk.Context, tx sdk.Tx) bool {
	if ctx.BlockHeight() != 0 {
		return false
	}
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*stakingtypes.MsgCreateValidator); !ok {
			return false
		}
	}
	return true
}
