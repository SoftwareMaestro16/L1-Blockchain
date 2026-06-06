package keeper

import paymentstypes "github.com/sovereign-l1/l1/x/payments/types"

func (k Keeper) FraudProofVerificationState() (paymentstypes.FraudProofVerificationState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.FraudProofVerificationState{}, err
	}
	return k.genesis.FraudProofs.Export(), nil
}

func (k *Keeper) HandleFraudProofVerificationMessage(msg interface{}) (paymentstypes.FraudProofVerificationState, error) {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return paymentstypes.FraudProofVerificationState{}, err
	}
	nextChain, nextModule, err := paymentstypes.ApplyFraudProofVerificationMessage(k.genesis.State, k.genesis.FraudProofs, msg)
	if err != nil {
		return paymentstypes.FraudProofVerificationState{}, err
	}
	k.genesis.State = nextChain
	k.genesis.FraudProofs = nextModule
	return nextModule.Export(), nil
}
