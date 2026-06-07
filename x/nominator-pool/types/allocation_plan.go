package types

import (
	"errors"
	"sort"
)

type PoolAllocationPlanInput struct {
	PoolID                string
	Epoch                 uint64
	Height                uint64
	MaxTouchedAllocations uint32
	Weights               []AllocationWeight
}

func BuildPoolAllocationPlan(input PoolAllocationPlanInput) (PoolRebalanceReceipt, error) {
	if err := validateID("pool allocation plan pool id", input.PoolID, MaxPoolIDBytesV1); err != nil {
		return PoolRebalanceReceipt{}, err
	}
	if input.Epoch == 0 || input.Height == 0 {
		return PoolRebalanceReceipt{}, errors.New("pool allocation plan epoch and height must be positive")
	}
	if input.MaxTouchedAllocations == 0 {
		return PoolRebalanceReceipt{}, errors.New("pool allocation plan max touched allocations must be positive")
	}
	weights := append([]AllocationWeight(nil), input.Weights...)
	sort.SliceStable(weights, func(i, j int) bool { return weights[i].ValidatorAddress < weights[j].ValidatorAddress })

	receipt := PoolRebalanceReceipt{
		PoolID: input.PoolID,
		Epoch:  input.Epoch,
		Height: input.Height,
	}
	for _, weight := range weights {
		if weight.WeightBps == 0 {
			continue
		}
		if uint32(len(receipt.Allocations)) >= input.MaxTouchedAllocations {
			break
		}
		allocation := PoolValidatorAllocation{
			PoolID:          input.PoolID,
			Validator:       weight.ValidatorAddress,
			TargetWeightBps: weight.WeightBps,
			UpdatedHeight:   input.Height,
		}
		receipt.Allocations = append(receipt.Allocations, allocation)
		receipt.InternalMetadata.TouchedKeys = append(receipt.InternalMetadata.TouchedKeys, string(PoolAllocationKey(input.PoolID, weight.ValidatorAddress)))
	}
	if len(receipt.Allocations) == 0 {
		return PoolRebalanceReceipt{}, errors.New("pool allocation plan requires at least one positive allocation")
	}
	return receipt, nil
}
