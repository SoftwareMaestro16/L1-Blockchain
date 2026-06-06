package keeper

import (
	"errors"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
)

func (k *Keeper) HandlePaymentChannelMessage(msg paymentstypes.PaymentChannelModuleMessage) (paymentstypes.PaymentChannelMessageResult, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.PaymentChannelMessageResult{}, err
	}
	next, result, err := paymentstypes.ApplyPaymentChannelMessage(k.genesis.State, msg)
	if err != nil {
		return paymentstypes.PaymentChannelMessageResult{}, err
	}
	k.genesis.State = next
	return result, nil
}

func (k Keeper) ValidatePaymentChannelAnte(msg paymentstypes.PaymentChannelModuleMessage) (paymentstypes.PaymentChannelAnteFee, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.PaymentChannelAnteFee{}, err
	}
	return paymentstypes.ValidatePaymentChannelMessageFee(k.genesis.State, msg)
}

func (k Keeper) PaymentChannelAccessPlan(msg paymentstypes.PaymentChannelModuleMessage, blockHeight uint64) (paymentstypes.BlockSTMAccessPlan, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.BlockSTMAccessPlan{}, err
	}
	return paymentstypes.PaymentChannelMessageAccessPlan(msg, blockHeight)
}

func (k Keeper) PaymentChannelConflictProfile(messages []paymentstypes.PaymentChannelModuleMessage, blockHeight uint64) (paymentstypes.BlockSTMConflictProfile, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.BlockSTMConflictProfile{}, err
	}
	return paymentstypes.PaymentChannelMessagesConflictProfile(messages, blockHeight)
}

func (k Keeper) PaymentChannelModuleState(blockHeight uint64) (paymentstypes.PaymentChannelModuleState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.PaymentChannelModuleState{}, err
	}
	return paymentstypes.SnapshotPaymentChannelModuleState(k.genesis.State, blockHeight)
}

func (k Keeper) ChannelFeeAccumulators(req *prototype.PageRequest, blockHeight uint64) ([]paymentstypes.ChannelFeeAccumulator, prototype.PageResponse, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return nil, prototype.PageResponse{}, err
	}
	snapshot, err := paymentstypes.SnapshotPaymentChannelModuleState(k.genesis.State, blockHeight)
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	start, end, page, err := prototype.NormalizePage(req, k.genesis.Params, len(snapshot.FeeAccumulators))
	if err != nil {
		return nil, prototype.PageResponse{}, err
	}
	out := make([]paymentstypes.ChannelFeeAccumulator, end-start)
	copy(out, snapshot.FeeAccumulators[start:end])
	return out, page, nil
}

func (k Keeper) AssertPaymentChannelCollateralInvariant() error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	if err := paymentstypes.ValidateLockedCollateralForFinality(k.genesis.State); err != nil {
		return err
	}
	if err := k.genesis.State.Validate(); err != nil {
		return err
	}
	return nil
}

func (k Keeper) RejectEarlyTombstonePruning(channelID string, pruneHeight uint64) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	state := k.genesis.State.Export()
	channel, found := state.ChannelByID(channelID)
	if !found {
		return errors.New("payments tombstone prune channel not found")
	}
	for _, tombstone := range state.ClosedChannels {
		tombstone = tombstone.Normalize()
		if tombstone.ChannelID != channel.Normalize().ChannelID {
			continue
		}
		if pruneHeight < tombstone.ExpiresHeight {
			return errors.New("payments settlement tombstone pruning before replay horizon")
		}
		return nil
	}
	return errors.New("payments settlement tombstone not found")
}
