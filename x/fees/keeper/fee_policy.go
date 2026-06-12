package keeper

import (
	"context"
	"encoding/binary"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/observability"
	"github.com/sovereign-l1/l1/x/fees/types"
)

// stateCreatingMsgTypes contains message type URLs whose execution creates or
// expands chain state. Transactions carrying at least one such message should
// incur the additional StorageRentSideEffects fee.
var stateCreatingMsgTypes = map[string]bool{
	"/l1.contracts.v1.MsgStoreCode":		true,
	"/l1.contracts.v1.MsgDeployContract":		true,
	"/l1.contracts.v1.MsgExecuteExternal":		true,
	"/l1.contracts.v1.MsgExecuteInternal":		true,
	"/l1.contracts.v1.MsgSendInternalMessage":	true,
	"/l1.contracts.v1.MsgUpdateContractParams":	true,
}

func (k Keeper) ValidateTxFees(ctx context.Context, fees sdk.Coins) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	return types.ValidateFeeCoins(params, fees, true)
}

func (k Keeper) AdmitTx(ctx sdk.Context, tx sdk.FeeTx, sender sdk.AccAddress, simulate bool) (types.FeeQuote, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.FeeQuote{}, err
	}
	formulaParams, err := k.GetFeeFormulaParams(ctx)
	if err != nil {
		return types.FeeQuote{}, err
	}
	blockCount, err := k.getBlockTxCount(ctx)
	if err != nil {
		return types.FeeQuote{}, err
	}
	senderCount, err := k.getSenderTxCount(ctx, sender)
	if err != nil {
		return types.FeeQuote{}, err
	}
	gasConsumed := uint64(0)
	if ctx.BlockGasMeter() != nil {
		gasConsumed = ctx.BlockGasMeter().GasConsumed()
	}

	quote, err := types.ValidateAdmission(params, types.AdmissionInput{
		Fee:			tx.GetFee(),
		GasLimit:		tx.GetGas(),
		BlockGasConsumed:	gasConsumed,
		BlockTxCount:		blockCount,
		SenderTxCount:		senderCount,
		SenderStake:		sdkmath.ZeroInt(),
	})
	if err != nil {
		return types.FeeQuote{}, err
	}

	kvUtilizationBps := k.getKVCongestionBps(ctx, gasConsumed, tx.GetGas(), params.MaxBlockGas)

	reputationScore, reputationFound, err := k.GetReputationScore(ctx, sender)
	if err != nil {

		reputationScore = types.ReputationNeutralScore
		reputationFound = false
	}

	storageRentNaet := sdkmath.ZeroInt()
	if srDefault, parseErr := formulaParams.StorageRentSideEffectsInt(); parseErr == nil && srDefault.IsPositive() {
		for _, msg := range tx.GetMsgs() {
			if stateCreatingMsgTypes[sdk.MsgTypeURL(msg)] {
				storageRentNaet = srDefault
				break
			}
		}
	}

	// Compute the full deterministic fee per Requirement 1.1.
	var txSizeBytes uint64
	if txBytes := ctx.TxBytes(); len(txBytes) > 0 {
		txSizeBytes = uint64(len(txBytes))
	}
	msgCount := uint64(len(tx.GetMsgs()))

	requiredFull, err := types.ComputeFullTransferFee(
		params,
		formulaParams,
		tx.GetGas(),
		txSizeBytes,
		msgCount,
		kvUtilizationBps,
		reputationScore,
		reputationFound,
		storageRentNaet,
	)
	if err != nil {
		return types.FeeQuote{}, err
	}

	paidAmount := tx.GetFee().AmountOf(types.BondDenom)
	if paidAmount.LT(requiredFull) {
		return types.FeeQuote{}, types.ErrInvalidFee.Wrapf(
			"fee must be at least %s%s (full formula requirement), paid %s%s",
			requiredFull.String(), types.BondDenom,
			paidAmount.String(), types.BondDenom,
		)
	}

	maxFee, err := params.MaxFeeInt()
	if err != nil {
		return types.FeeQuote{}, err
	}
	if paidAmount.GT(maxFee) {
		return types.FeeQuote{}, types.ErrInvalidFee.Wrapf(
			"fee must not exceed hard cap %s%s", maxFee.String(), types.BondDenom,
		)
	}

	if !simulate {
		observability.RecordEconomicControl(
			quote.EconomicControl.InflationBps,
			quote.EconomicControl.BurnRatioBps,
			quote.EconomicControl.ValidatorFeeRatioBps,
			quote.EconomicControl.DeflationGuardActive,
			quote.EconomicControl.QueueLimited,
			quote.EconomicControl.RateLimited,
		)
		if flow, err := appparams.ComputeProtocolEconomicFlow(appparams.ProtocolEconomicFlowInput{
			Activity: appparams.ProtocolEconomicActivity{
				TxFeeNaet: quote.AcceptedFeeAmount,
			},
			BurnRatioBps:		quote.EconomicControl.BurnRatioBps,
			TreasuryRatioBps:	appparams.TreasuryFeeRatioBps,
		}); err == nil {
			observability.RecordEconomicFlow(
				flow.TotalChargesNaet.Int64(),
				flow.BurnNaet.Int64(),
				flow.TreasuryNaet.Int64(),
				flow.ValidatorRewardsNaet.Int64(),
			)
		}
		if err := k.setBlockTxCount(ctx, blockCount+1); err != nil {
			return types.FeeQuote{}, err
		}
		if err := k.setSenderTxCount(ctx, sender, senderCount+1); err != nil {
			return types.FeeQuote{}, err
		}
	}
	return quote, nil
}

func (k Keeper) TxPriority(params types.Params, paidFee sdk.Coin, requiredFee sdk.Coin, stake sdkmath.Int) (int64, error) {
	return types.PriorityScore(params, paidFee, requiredFee, stake)
}

func (k Keeper) getBlockTxCount(ctx sdk.Context) (uint64, error) {
	return k.getHeightCounter(ctx, types.BlockTxCountKey)
}

func (k Keeper) setBlockTxCount(ctx sdk.Context, count uint64) error {
	return k.setHeightCounter(ctx, types.BlockTxCountKey, count)
}

func (k Keeper) getSenderTxCount(ctx sdk.Context, sender sdk.AccAddress) (uint64, error) {
	return k.getHeightCounter(ctx, senderTxCountKey(sender))
}

func (k Keeper) setSenderTxCount(ctx sdk.Context, sender sdk.AccAddress, count uint64) error {
	return k.setHeightCounter(ctx, senderTxCountKey(sender), count)
}

func (k Keeper) getHeightCounter(ctx sdk.Context, key []byte) (uint64, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(key)
	if err != nil {
		return 0, err
	}
	if len(bz) != 16 {
		return 0, nil
	}
	height := int64(binary.BigEndian.Uint64(bz[:8]))
	if height != ctx.BlockHeight() {
		return 0, nil
	}
	return binary.BigEndian.Uint64(bz[8:]), nil
}

func (k Keeper) setHeightCounter(ctx sdk.Context, key []byte, count uint64) error {
	var bz [16]byte
	binary.BigEndian.PutUint64(bz[:8], uint64(ctx.BlockHeight()))
	binary.BigEndian.PutUint64(bz[8:], count)
	return k.storeService.OpenKVStore(ctx).Set(key, bz[:])
}

func senderTxCountKey(sender sdk.AccAddress) []byte {
	key := make([]byte, 0, len(types.SenderTxCountPrefix)+len(sender))
	key = append(key, types.SenderTxCountPrefix...)
	key = append(key, sender...)
	return key
}
