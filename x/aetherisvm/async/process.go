package async

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func (e *Executor) processNext() (ExecutionReceipt, error) {
	queued := e.queue[0]
	e.queue = e.queue[1:]
	msg := queued.Envelope
	msg.ExecutionBlockHeight = e.blockHeight
	receipt := ExecutionReceipt{
		Sequence:       queued.Sequence,
		Source:         append(sdk.AccAddress(nil), msg.Source...),
		Destination:    append(sdk.AccAddress(nil), msg.Destination...),
		Opcode:         msg.Opcode,
		QueryID:        msg.QueryID,
		GasUsed:        e.params.ExecutionGasPerMessage,
		StorageFeeNaet: sdkmath.ZeroInt(),
		ForwardFeeNaet: msg.ForwardFee.Amount,
	}
	e.metrics.ProcessedMessages++
	e.metrics.GasUsed += e.params.ExecutionGasPerMessage

	if msg.DeadlineBlock != 0 && e.blockHeight > msg.DeadlineBlock {
		receipt.ResultCode = ResultExpired
		receipt.Error = "message expired"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}

	contract, ok := e.contracts[string(msg.Destination)]
	if !ok {
		receipt.ResultCode = ResultNoDestination
		receipt.Error = "destination contract not found"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}

	handler := e.handlers[string(msg.Destination)]
	if handler == nil {
		receipt.ResultCode = ResultExecutionFailed
		receipt.Error = "destination contract has no handler"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}

	working := cloneContract(contract)
	working.BalanceNaet = working.BalanceNaet.Add(msg.Value.Amount)
	working.LogicalTime++
	result := handler(working, cloneMessage(msg))
	if result.GasUsed > 0 {
		receipt.GasUsed = result.GasUsed
	}
	if !e.acceptExecutionResult(&receipt, msg, result) {
		return receipt, nil
	}

	working.State = append([]byte(nil), result.NewState...)
	receipt.StorageFeeNaet = e.params.StorageFeePerByte.MulRaw(int64(len(working.State)))
	working.BalanceNaet = working.BalanceNaet.Sub(receipt.StorageFeeNaet)
	if working.BalanceNaet.IsNegative() {
		receipt.ResultCode = ResultExecutionFailed
		receipt.Error = "insufficient naet for storage fee"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, receipt)
		e.receipts = append(e.receipts, receipt)
		return receipt, nil
	}
	outgoing := make([]MessageEnvelope, len(result.Outgoing))
	outgoingTxIndex := e.nextTxIndex
	for i, out := range result.Outgoing {
		out.Source = append(sdk.AccAddress(nil), working.Address...)
		out.CreatedLogicalTime = working.LogicalTime
		out.ExecutionBlockHeight = 0
		out.Depth = msg.Depth + 1
		if err := out.Validate(e.params); err != nil {
			receipt.ResultCode = ResultLimitExceeded
			receipt.Error = err.Error()
			e.metrics.FailedExecutions++
			e.finalizeFailure(msg, receipt)
			e.receipts = append(e.receipts, receipt)
			return receipt, nil
		}
		outgoing[i] = out
	}
	e.contracts[string(working.Address)] = working
	if len(outgoing) > 0 {
		e.nextTxIndex++
	}
	for i, out := range outgoing {
		if err := e.enqueueMessageWithOrder(out, outgoingTxIndex, uint32(i)); err != nil {
			receipt.ResultCode = ResultLimitExceeded
			receipt.Error = err.Error()
			e.metrics.FailedExecutions++
			e.finalizeFailure(msg, receipt)
			e.receipts = append(e.receipts, receipt)
			return receipt, nil
		}
	}
	e.receipts = append(e.receipts, receipt)
	return receipt, nil
}

func (e *Executor) acceptExecutionResult(receipt *ExecutionReceipt, msg MessageEnvelope, result ExecutionResult) bool {
	if receipt.GasUsed > msg.GasLimit {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "message gas limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, *receipt)
		e.receipts = append(e.receipts, *receipt)
		return false
	}
	receipt.ResultCode = result.ResultCode
	if result.ResultCode != ResultOK {
		if result.Error != "" {
			receipt.Error = result.Error
		} else {
			receipt.Error = "contract execution failed"
		}
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, *receipt)
		e.receipts = append(e.receipts, *receipt)
		return false
	}
	if len(result.NewState) > int(e.params.MaxStateSize) {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "contract state limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, *receipt)
		e.receipts = append(e.receipts, *receipt)
		return false
	}
	if len(result.Outgoing) > int(e.params.MaxEmittedMessagesPerExec) {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "emitted message limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, *receipt)
		e.receipts = append(e.receipts, *receipt)
		return false
	}
	if result.StorageWrites > e.params.MaxStorageWritesPerExec {
		receipt.ResultCode = ResultLimitExceeded
		receipt.Error = "storage write limit exceeded"
		e.metrics.FailedExecutions++
		e.finalizeFailure(msg, *receipt)
		e.receipts = append(e.receipts, *receipt)
		return false
	}
	return true
}

func (e *Executor) finalizeFailure(msg MessageEnvelope, receipt ExecutionReceipt) {
	if msg.Bounce && !msg.Bounced {
		bounce := MessageEnvelope{
			Source:             append(sdk.AccAddress(nil), msg.Destination...),
			Destination:        append(sdk.AccAddress(nil), msg.Source...),
			Value:              sdk.NewCoin(appparams.BaseDenom, msg.Value.Amount),
			Opcode:             BounceOpcode,
			QueryID:            msg.QueryID,
			Body:               append([]byte(nil), msg.Body...),
			Bounce:             false,
			Bounced:            true,
			CreatedLogicalTime: msg.CreatedLogicalTime,
			DeadlineBlock:      msg.DeadlineBlock,
			GasLimit:           msg.GasLimit,
			ForwardFee:         sdk.NewCoin(appparams.BaseDenom, e.params.ForwardingFee),
			Depth:              msg.Depth + 1,
		}
		if err := e.EnqueueMessage(bounce); err == nil {
			e.metrics.BouncedMessages++
		}
		return
	}
	if msg.Bounced || msg.Opcode == RefundOpcode {
		return
	}
	if msg.Value.Amount.IsPositive() {
		refund := MessageEnvelope{
			Source:             append(sdk.AccAddress(nil), msg.Destination...),
			Destination:        append(sdk.AccAddress(nil), msg.Source...),
			Value:              sdk.NewCoin(appparams.BaseDenom, msg.Value.Amount),
			Opcode:             RefundOpcode,
			QueryID:            msg.QueryID,
			Body:               []byte("refund"),
			Bounce:             false,
			Bounced:            false,
			CreatedLogicalTime: msg.CreatedLogicalTime,
			DeadlineBlock:      0,
			GasLimit:           msg.GasLimit,
			ForwardFee:         sdk.NewCoin(appparams.BaseDenom, e.params.ForwardingFee),
			Depth:              msg.Depth + 1,
		}
		if err := e.EnqueueMessage(refund); err == nil {
			e.metrics.RefundMessages++
		}
	}
}
