package async

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

type RefundCalculation struct {
	Amount	sdkmath.Int
	Fee	sdkmath.Int
}

func CalculateRefund(msg MessageEnvelope, receipt ExecutionReceipt) (RefundCalculation, error) {
	if msg.Value.Amount.IsNil() || msg.Value.Amount.IsNegative() {
		return RefundCalculation{}, errors.New("refund source value must be non-negative")
	}
	fee := receipt.ForwardFeeNaet
	if fee.IsNil() {
		fee = sdkmath.ZeroInt()
	}
	if fee.IsNegative() {
		return RefundCalculation{}, errors.New("refund fee must be non-negative")
	}
	if fee.GT(msg.Value.Amount) {
		fee = msg.Value.Amount
	}
	return RefundCalculation{
		Amount:	msg.Value.Amount.Sub(fee),
		Fee:	fee,
	}, nil
}

func BuildBounceMessage(msg MessageEnvelope, refund RefundCalculation, forwardingFee sdkmath.Int) (MessageEnvelope, error) {
	if refund.Amount.IsNil() || refund.Amount.IsNegative() {
		return MessageEnvelope{}, errors.New("bounce refund amount must be non-negative")
	}
	if forwardingFee.IsNil() || forwardingFee.IsNegative() {
		return MessageEnvelope{}, errors.New("bounce forwarding fee must be non-negative")
	}
	return MessageEnvelope{
		Source:			append(sdk.AccAddress(nil), msg.Destination...),
		Destination:		append(sdk.AccAddress(nil), msg.Source...),
		Value:			sdk.NewCoin(appparams.BaseDenom, refund.Amount),
		Opcode:			BounceOpcode,
		QueryID:		msg.QueryID,
		Body:			append([]byte(nil), msg.Body...),
		Bounce:			false,
		Bounced:		true,
		CreatedLogicalTime:	msg.CreatedLogicalTime,
		DeadlineBlock:		msg.DeadlineBlock,
		GasLimit:		msg.GasLimit,
		ForwardFee:		sdk.NewCoin(appparams.BaseDenom, forwardingFee),
		Depth:			msg.Depth + 1,
	}, nil
}

func BuildRefundMessage(msg MessageEnvelope, refund RefundCalculation, forwardingFee sdkmath.Int) (MessageEnvelope, error) {
	if refund.Amount.IsNil() || refund.Amount.IsNegative() {
		return MessageEnvelope{}, errors.New("refund amount must be non-negative")
	}
	if forwardingFee.IsNil() || forwardingFee.IsNegative() {
		return MessageEnvelope{}, errors.New("refund forwarding fee must be non-negative")
	}
	return MessageEnvelope{
		Source:			append(sdk.AccAddress(nil), msg.Destination...),
		Destination:		append(sdk.AccAddress(nil), msg.Source...),
		Value:			sdk.NewCoin(appparams.BaseDenom, refund.Amount),
		Opcode:			RefundOpcode,
		QueryID:		msg.QueryID,
		Body:			[]byte("refund"),
		Bounce:			false,
		Bounced:		false,
		CreatedLogicalTime:	msg.CreatedLogicalTime,
		DeadlineBlock:		0,
		GasLimit:		msg.GasLimit,
		ForwardFee:		sdk.NewCoin(appparams.BaseDenom, forwardingFee),
		Depth:			msg.Depth + 1,
	}, nil
}

func MarkRefunded(receipt *ExecutionReceipt, refund RefundCalculation, reason string, refundSequence uint64) error {
	if receipt == nil {
		return errors.New("receipt is nil")
	}
	if receipt.Refunded {
		return errors.New("receipt already refunded")
	}
	if refund.Amount.IsNil() || refund.Amount.IsNegative() {
		return errors.New("refund amount must be non-negative")
	}
	if refund.Fee.IsNil() || refund.Fee.IsNegative() {
		return errors.New("refund fee must be non-negative")
	}
	receipt.Refunded = true
	receipt.RefundAmountNaet = refund.Amount
	receipt.RefundFeeNaet = refund.Fee
	receipt.RefundReason = reason
	receipt.RefundOfSequence = refundSequence
	return nil
}
