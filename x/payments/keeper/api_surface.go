package keeper

import paymentstypes "github.com/sovereign-l1/l1/x/payments/types"

func (k *Keeper) HandlePaymentAPIMessage(msg interface{}) (paymentstypes.PaymentAPISurfaceResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.PaymentAPISurfaceResult{}, err
	}
	nextState, nextFraud, result, err := paymentstypes.ApplyPaymentAPISurfaceMessage(k.genesis.State, k.genesis.FraudProofs, msg)
	if err != nil {
		return paymentstypes.PaymentAPISurfaceResult{}, err
	}
	k.genesis.State = nextState.Export()
	k.genesis.FraudProofs = nextFraud.Export()
	return result, nil
}

func (k Keeper) QueryChannel(channelID string) (paymentstypes.ChannelRecord, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ChannelRecord{}, false, err
	}
	return paymentstypes.QueryChannel(k.genesis.State, channelID)
}

func (k Keeper) QueryChannelsByParticipant(participant string) ([]paymentstypes.ChannelRecord, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	return paymentstypes.QueryChannelsByParticipant(k.genesis.State, participant)
}

func (k Keeper) QueryPendingClose(channelID string) (paymentstypes.PendingClose, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.PendingClose{}, false, err
	}
	return paymentstypes.QueryPendingClose(k.genesis.State, channelID)
}

func (k Keeper) QueryFinalizationHeight(channelID string) (uint64, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return 0, false, err
	}
	return paymentstypes.QueryFinalizationHeight(k.genesis.State, channelID)
}

func (k Keeper) QueryCondition(channelID, conditionID string) (paymentstypes.ConditionalPayment, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ConditionalPayment{}, false, err
	}
	return paymentstypes.QueryCondition(k.genesis.State, channelID, conditionID)
}

func (k Keeper) QueryConditionsByChannel(channelID string) ([]paymentstypes.ConditionalPayment, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	return paymentstypes.QueryConditionsByChannel(k.genesis.State, channelID)
}

func (k Keeper) QueryVirtualChannel(virtualChannelID string) (paymentstypes.VirtualChannel, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.VirtualChannel{}, false, err
	}
	return paymentstypes.QueryVirtualChannel(k.genesis.State, virtualChannelID)
}

func (k Keeper) QueryChannelCapacity(channelID string, currentHeight uint64) (paymentstypes.ChannelCapacity, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ChannelCapacity{}, err
	}
	return paymentstypes.QueryChannelCapacity(k.genesis.State, k.genesis.Liquidity, channelID, currentHeight)
}

func (k Keeper) QueryFeeSchedule() (paymentstypes.PaymentFeeSchedule, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.PaymentFeeSchedule{}, err
	}
	return paymentstypes.QueryFeeSchedule(k.genesis.State)
}

func (k Keeper) QuerySettlementTombstone(channelID string) (paymentstypes.ClosedChannelTombstone, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ClosedChannelTombstone{}, false, err
	}
	return paymentstypes.QuerySettlementTombstone(k.genesis.State, channelID)
}

func (k Keeper) QueryFraudProof(proofID string) (paymentstypes.FraudProofQueryResult, bool, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.FraudProofQueryResult{}, false, err
	}
	return paymentstypes.QueryFraudProof(k.genesis.State, k.genesis.FraudProofs, proofID)
}

func (k Keeper) QueryActiveDisputes(currentHeight uint64) ([]paymentstypes.AdaptiveSyncActiveDisputeIndex, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	return paymentstypes.QueryActiveDisputes(k.genesis.State, currentHeight)
}

func (k Keeper) QueryPendingFinalizations(currentHeight uint64) ([]paymentstypes.AdaptiveSyncPendingFinalizationIndex, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, err
	}
	return paymentstypes.QueryPendingFinalizations(k.genesis.State, currentHeight)
}
