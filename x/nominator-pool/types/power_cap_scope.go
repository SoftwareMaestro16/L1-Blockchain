package types

import "errors"

type ValidatorPowerCapStage string

const (
	ValidatorPowerCapStageRewardsOnly	ValidatorPowerCapStage	= "stage1_rewards_allocation_only"
	ValidatorPowerCapStageCometBFT		ValidatorPowerCapStage	= "stage2_cometbft_voting_power"
)

type ValidatorPowerCapScope struct {
	Stage				ValidatorPowerCapStage	`json:"stage"`
	CapsPoolAllocationWeight	bool			`json:"caps_pool_allocation_weight"`
	CapsRewardEffectivePower	bool			`json:"caps_reward_effective_power"`
	CapsCometBFTVotingPower		bool			`json:"caps_cometbft_voting_power"`
	ConsensusVotingPowerOwner	string			`json:"consensus_voting_power_owner"`
	RequiresCometBFTCapEvidence	bool			`json:"requires_cometbft_cap_evidence"`
}

func ActiveValidatorPowerCapScope() ValidatorPowerCapScope {
	return ValidatorPowerCapScope{
		Stage:				ValidatorPowerCapStageRewardsOnly,
		CapsPoolAllocationWeight:	true,
		CapsRewardEffectivePower:	true,
		CapsCometBFTVotingPower:	false,
		ConsensusVotingPowerOwner:	"x/pos+x/staking+CometBFT",
		RequiresCometBFTCapEvidence:	true,
	}
}

func (s ValidatorPowerCapScope) Validate() error {
	switch s.Stage {
	case ValidatorPowerCapStageRewardsOnly:
		if !s.CapsPoolAllocationWeight || !s.CapsRewardEffectivePower {
			return errors.New("stage1 validator power cap must apply to pool allocation and reward effective power")
		}
		if s.CapsCometBFTVotingPower {
			return errors.New("stage1 validator power cap must not claim CometBFT voting power enforcement")
		}
		if s.ConsensusVotingPowerOwner == "" {
			return errors.New("stage1 validator power cap must declare consensus voting power owner")
		}
		if !s.RequiresCometBFTCapEvidence {
			return errors.New("stage1 validator power cap must require CometBFT cap evidence before stage2")
		}
	case ValidatorPowerCapStageCometBFT:
		if !s.CapsPoolAllocationWeight || !s.CapsRewardEffectivePower || !s.CapsCometBFTVotingPower {
			return errors.New("stage2 validator power cap must cover allocation, rewards, and CometBFT voting power")
		}
	default:
		return errors.New("unsupported validator power cap stage")
	}
	return nil
}
