package types

import "errors"

const (
	RewardPolicyV1ID			= "nominator-pool/reward-policy/v1"
	RewardSourceFeesAndInflation		= "fees_and_inflation"
	RewardDistributionByPoolShares		= "pool_shares"
	RewardRoundingFloorWithRemainder	= "floor_with_pool_remainder"
	RewardValidatorCommissionBeforePool	= "validator_commission_before_pool_fee"
)

type RewardPolicyV1 struct {
	PolicyID		string
	RewardSource		string
	Distribution		string
	CommissionRule		string
	RoundingRule		string
	LazyRewardIndex		bool
	CapByCollectedRewards	bool
	ManualValidatorChoice	bool
}

func DefaultRewardPolicyV1() RewardPolicyV1 {
	return RewardPolicyV1{
		PolicyID:		RewardPolicyV1ID,
		RewardSource:		RewardSourceFeesAndInflation,
		Distribution:		RewardDistributionByPoolShares,
		CommissionRule:		RewardValidatorCommissionBeforePool,
		RoundingRule:		RewardRoundingFloorWithRemainder,
		LazyRewardIndex:	true,
		CapByCollectedRewards:	true,
		ManualValidatorChoice:	false,
	}
}

func (p RewardPolicyV1) Validate() error {
	if p.PolicyID != RewardPolicyV1ID {
		return errors.New("reward policy id must be v1")
	}
	if p.RewardSource != RewardSourceFeesAndInflation {
		return errors.New("reward policy source must be fees plus inflation")
	}
	if p.Distribution != RewardDistributionByPoolShares {
		return errors.New("reward policy must distribute by pool shares")
	}
	if p.CommissionRule != RewardValidatorCommissionBeforePool {
		return errors.New("reward policy must deduct validator commission before pool fee")
	}
	if p.RoundingRule != RewardRoundingFloorWithRemainder {
		return errors.New("reward policy rounding must floor and carry pool remainder")
	}
	if !p.LazyRewardIndex {
		return errors.New("reward policy must use lazy reward index accounting")
	}
	if !p.CapByCollectedRewards {
		return errors.New("reward policy must cap payouts by collected and emitted rewards")
	}
	if p.ManualValidatorChoice {
		return errors.New("reward policy must not use manual validator choice")
	}
	return nil
}
