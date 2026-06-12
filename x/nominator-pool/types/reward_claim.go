package types

import "errors"

type PoolRewardClaimInput struct {
	PoolID		string
	Share		DelegatorShare
	RewardIndex	uint64
	Epoch		uint64
	Height		uint64
}

func ClaimPoolRewardShare(input PoolRewardClaimInput) (DelegatorShare, PoolRewardClaimReceipt, error) {
	if err := validateID("pool reward claim pool id", input.PoolID, MaxPoolIDBytesV1); err != nil {
		return DelegatorShare{}, PoolRewardClaimReceipt{}, err
	}
	if err := ValidateUserFacingAEAddress("pool reward claim delegator", input.Share.Delegator); err != nil {
		return DelegatorShare{}, PoolRewardClaimReceipt{}, err
	}
	if input.Share.Shares == 0 {
		return DelegatorShare{}, PoolRewardClaimReceipt{}, errors.New("pool reward claim requires positive shares")
	}
	if input.Epoch == 0 || input.Height == 0 {
		return DelegatorShare{}, PoolRewardClaimReceipt{}, errors.New("pool reward claim epoch and height must be positive")
	}
	if input.RewardIndex < input.Share.RewardIndexCheckpoint {
		return DelegatorShare{}, PoolRewardClaimReceipt{}, errors.New("pool reward claim index cannot go backwards")
	}
	amount := AccruedReward(input.Share, input.RewardIndex)
	if amount == 0 {
		return DelegatorShare{}, PoolRewardClaimReceipt{}, errors.New("pool reward claim has no claimable rewards")
	}
	next := input.Share
	next.PendingRewards = 0
	next.RewardIndexCheckpoint = input.RewardIndex
	receipt := PoolRewardClaimReceipt{
		PoolID:		input.PoolID,
		OwnerAddress:	next.Delegator,
		Amount:		amount,
		Epoch:		input.Epoch,
		Height:		input.Height,
		InternalMetadata: PoolStateMetadata{
			TouchedKeys: []string{string(PoolShareKey(input.PoolID, next.Delegator))},
		},
	}
	return next, receipt, nil
}

func RecordPoolRewardClaim(params Params, existing []RewardClaim, receipt PoolRewardClaimReceipt) ([]RewardClaim, RewardClaim, error) {
	if err := params.Validate(); err != nil {
		return nil, RewardClaim{}, err
	}
	if err := validateID("pool reward claim record pool id", receipt.PoolID, params.MaxPoolIDBytes); err != nil {
		return nil, RewardClaim{}, err
	}
	if err := ValidateUserFacingAEAddress("pool reward claim record owner", receipt.OwnerAddress); err != nil {
		return nil, RewardClaim{}, err
	}
	if receipt.Epoch == 0 || receipt.Height == 0 {
		return nil, RewardClaim{}, errors.New("pool reward claim record epoch and height must be positive")
	}
	if receipt.Amount == 0 {
		return nil, RewardClaim{}, errors.New("pool reward claim record amount must be positive")
	}
	claim := RewardClaim{
		PoolID:	receipt.PoolID,
		Owner:	receipt.OwnerAddress,
		Epoch:	receipt.Epoch,
		Amount:	receipt.Amount,
	}
	if err := claim.Validate(params); err != nil {
		return nil, RewardClaim{}, err
	}
	next := append([]RewardClaim(nil), existing...)
	for _, current := range next {
		if current.PoolID == claim.PoolID && current.Owner == claim.Owner && current.Epoch == claim.Epoch {
			return nil, RewardClaim{}, errors.New("duplicate reward claim")
		}
	}
	next = append(next, claim)
	next = SortRewardClaims(next)
	return next, claim, nil
}
