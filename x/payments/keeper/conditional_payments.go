package keeper

import paymentstypes "github.com/sovereign-l1/l1/x/payments/types"

func (k Keeper) ConditionalPaymentsState() (paymentstypes.ConditionalPaymentsModuleState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ConditionalPaymentsModuleState{}, err
	}
	return paymentstypes.SnapshotConditionalPaymentsModuleState(k.genesis.State)
}

func (k *Keeper) HandleConditionalPaymentMessage(msg paymentstypes.ConditionalPaymentMessage) (paymentstypes.ConditionalPaymentsModuleState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.ConditionalPaymentsModuleState{}, err
	}
	next, snapshot, err := paymentstypes.ApplyConditionalPaymentMessage(k.genesis.State, msg)
	if err != nil {
		return paymentstypes.ConditionalPaymentsModuleState{}, err
	}
	k.genesis.State = next
	return snapshot, nil
}

func (k Keeper) AssertConditionalReserveInvariant() error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	state := k.genesis.State.Export()
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if err := paymentstypes.ValidateReservedBalancesForConditions(channel, channel.LatestState); err != nil {
			return err
		}
	}
	return nil
}
