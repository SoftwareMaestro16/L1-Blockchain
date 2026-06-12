package types

import (
	"fmt"
)

const (
	MaxFeePremiumBpsDefault			= uint32(15_000)
	MaxFeeDiscountBpsDefault		= uint32(5_000)
	MaxQueuePriorityBoostDefault		= uint32(5_000)
	MaxValidatorAllocBoostBpsDefault	= uint32(2_000)
	MaxValidatorAllocPenaltyBpsDefault	= uint32(3_000)
	MaxServiceTrustRoutingBoostDefault	= uint32(3_000)
	DecayRatePerEpochDefaultEffect		= uint8(1)
	ConfidenceGainRateDefault		= uint32(100)
	ConfidenceLossRateDefault		= uint32(200)
	PerEpochScoreCapDefault			= uint32(500)
	PerEpochConfidenceCapDefault		= uint32(300)

	NeutralScoreBps	= uint32(5_000)

	MinStakeAmountForPoolStakingDefault	= uint64(10_000_000_000)
)

type ReputationEffectParams struct {
	MaxFeePremiumBps		uint32	`json:"max_fee_premium_bps"`
	MaxFeeDiscountBps		uint32	`json:"max_fee_discount_bps"`
	MaxQueuePriorityBoostBps	uint32	`json:"max_queue_priority_boost_bps"`
	MaxValidatorAllocBoostBps	uint32	`json:"max_validator_allocation_boost_bps"`
	MaxValidatorAllocPenaltyBps	uint32	`json:"max_validator_allocation_penalty_bps"`
	MaxServiceTrustRoutingBps	uint32	`json:"max_service_trust_routing_boost_bps"`
	DecayRatePerEpoch		uint8	`json:"decay_rate_per_epoch"`
	ConfidenceGainRate		uint32	`json:"confidence_gain_rate"`
	ConfidenceLossRate		uint32	`json:"confidence_loss_rate"`
	PerEpochScoreCap		uint32	`json:"per_epoch_score_cap"`
	PerEpochConfidenceCap		uint32	`json:"per_epoch_confidence_cap"`
	MinStakeAmountForPool		uint64	`json:"min_stake_amount_for_pool"`
}

func DefaultReputationEffectParams() ReputationEffectParams {
	return ReputationEffectParams{
		MaxFeePremiumBps:		MaxFeePremiumBpsDefault,
		MaxFeeDiscountBps:		MaxFeeDiscountBpsDefault,
		MaxQueuePriorityBoostBps:	MaxQueuePriorityBoostDefault,
		MaxValidatorAllocBoostBps:	MaxValidatorAllocBoostBpsDefault,
		MaxValidatorAllocPenaltyBps:	MaxValidatorAllocPenaltyBpsDefault,
		MaxServiceTrustRoutingBps:	MaxServiceTrustRoutingBoostDefault,
		DecayRatePerEpoch:		DecayRatePerEpochDefaultEffect,
		ConfidenceGainRate:		ConfidenceGainRateDefault,
		ConfidenceLossRate:		ConfidenceLossRateDefault,
		PerEpochScoreCap:		PerEpochScoreCapDefault,
		PerEpochConfidenceCap:		PerEpochConfidenceCapDefault,
		MinStakeAmountForPool:		MinStakeAmountForPoolStakingDefault,
	}
}

func (p ReputationEffectParams) Validate() error {
	if p.MaxFeePremiumBps == 0 {
		return fmt.Errorf("reputation effect params: max_fee_premium_bps must be positive")
	}
	if p.MaxFeeDiscountBps > 10_000 {
		return fmt.Errorf("reputation effect params: max_fee_discount_bps %d exceeds 10000", p.MaxFeeDiscountBps)
	}
	if p.MaxQueuePriorityBoostBps > 10_000 {
		return fmt.Errorf("reputation effect params: max_queue_priority_boost_bps %d exceeds 10000", p.MaxQueuePriorityBoostBps)
	}
	if p.MaxValidatorAllocBoostBps > 10_000 {
		return fmt.Errorf("reputation effect params: max_validator_allocation_boost_bps %d exceeds 10000", p.MaxValidatorAllocBoostBps)
	}
	if p.MaxValidatorAllocPenaltyBps > 10_000 {
		return fmt.Errorf("reputation effect params: max_validator_allocation_penalty_bps %d exceeds 10000", p.MaxValidatorAllocPenaltyBps)
	}
	if p.MaxServiceTrustRoutingBps > 10_000 {
		return fmt.Errorf("reputation effect params: max_service_trust_routing_boost_bps %d exceeds 10000", p.MaxServiceTrustRoutingBps)
	}
	if p.DecayRatePerEpoch == 0 {
		return fmt.Errorf("reputation effect params: decay_rate_per_epoch must be positive")
	}
	if p.ConfidenceGainRate == 0 {
		return fmt.Errorf("reputation effect params: confidence_gain_rate must be positive")
	}
	if p.ConfidenceLossRate == 0 {
		return fmt.Errorf("reputation effect params: confidence_loss_rate must be positive")
	}
	if p.PerEpochScoreCap == 0 {
		return fmt.Errorf("reputation effect params: per_epoch_score_cap must be positive")
	}
	if p.PerEpochConfidenceCap == 0 {
		return fmt.Errorf("reputation effect params: per_epoch_confidence_cap must be positive")
	}
	return nil
}

func ComputeBoundedFeeAdjustment(score uint32, found bool, params ReputationEffectParams) (premiumNaet int64, discountNaet int64) {
	if !found || score == NeutralScoreBps {
		return 0, 0
	}
	if score < NeutralScoreBps {
		deficit := NeutralScoreBps - score
		premiumBps := uint32(uint64(deficit) * uint64(params.MaxFeePremiumBps) / uint64(NeutralScoreBps))
		if premiumBps > params.MaxFeePremiumBps {
			premiumBps = params.MaxFeePremiumBps
		}
		return int64(premiumBps), 0
	}
	surplus := score - NeutralScoreBps
	discountBps := uint32(uint64(surplus) * uint64(params.MaxFeeDiscountBps) / uint64(IdentityScoreMax-NeutralScoreBps))
	if discountBps > params.MaxFeeDiscountBps {
		discountBps = params.MaxFeeDiscountBps
	}
	return 0, int64(discountBps)
}

func ComputeQueuePriorityBoost(score uint32, params ReputationEffectParams) uint32 {
	if score >= NeutralScoreBps {
		surplus := score - NeutralScoreBps
		boost := uint32(uint64(surplus) * uint64(params.MaxQueuePriorityBoostBps) / uint64(IdentityScoreMax-NeutralScoreBps))
		if boost > params.MaxQueuePriorityBoostBps {
			boost = params.MaxQueuePriorityBoostBps
		}
		return boost
	}
	return 0
}

func ComputeQueuePriorityPenalty(score uint32, params ReputationEffectParams) uint32 {
	if score < NeutralScoreBps {
		deficit := NeutralScoreBps - score
		penalty := uint32(uint64(deficit) * uint64(params.MaxQueuePriorityBoostBps) / uint64(NeutralScoreBps))
		if penalty > params.MaxQueuePriorityBoostBps {
			penalty = params.MaxQueuePriorityBoostBps
		}
		return penalty
	}
	return 0
}

func ComputePoolAllocationWeight(vs *ValidatorScore, params ReputationEffectParams) uint32 {
	if vs == nil {
		return 0
	}
	if vs.IsJailed || vs.IsSlashed {
		return 0
	}
	rawScore := vs.TotalScore
	boostBps := uint32(uint64(rawScore) * uint64(params.MaxValidatorAllocBoostBps) / uint64(IdentityScoreMax))
	if boostBps > params.MaxValidatorAllocBoostBps {
		boostBps = params.MaxValidatorAllocBoostBps
	}
	return boostBps
}

func ComputePoolAllocationPenalty(vs *ValidatorScore, params ReputationEffectParams) uint32 {
	if vs == nil {
		return params.MaxValidatorAllocPenaltyBps
	}
	deficit := uint32(0)
	if vs.TotalScore < IdentityScoreMax {
		deficit = IdentityScoreMax - vs.TotalScore
	}
	penaltyBps := uint32(uint64(deficit) * uint64(params.MaxValidatorAllocPenaltyBps) / uint64(IdentityScoreMax))
	if penaltyBps > params.MaxValidatorAllocPenaltyBps {
		penaltyBps = params.MaxValidatorAllocPenaltyBps
	}
	if vs.IsJailed || vs.IsSlashed {
		penaltyBps = params.MaxValidatorAllocPenaltyBps
	}
	return penaltyBps
}

func ComputeServiceTrustRoutingBoost(sts *ServiceTrustScore, params ReputationEffectParams) uint32 {
	if sts == nil {
		return 0
	}
	surplus := int64(sts.Trust) - int64(50*100)
	if surplus <= 0 {
		return 0
	}
	boost := uint32(uint64(surplus) * uint64(params.MaxServiceTrustRoutingBps) / uint64(int64(IdentityScoreMax)-50*100))
	if boost > params.MaxServiceTrustRoutingBps {
		boost = params.MaxServiceTrustRoutingBps
	}
	return boost
}

func ApplyPerEpochScoreCap(oldScore uint32, newScore uint32, cap uint32) uint32 {
	delta := int32(newScore) - int32(oldScore)
	if delta > 0 && uint32(delta) > cap {
		return oldScore + cap
	}
	if delta < 0 && uint32(-delta) > cap {
		if oldScore < cap {
			return IdentityScoreDefault
		}
		return oldScore - cap
	}
	return newScore
}

func ApplyPerEpochConfidenceCap(oldConf uint32, newConf uint32, cap uint32) uint32 {
	delta := int32(newConf) - int32(oldConf)
	if delta > 0 && uint32(delta) > cap {
		return oldConf + cap
	}
	if delta < 0 && uint32(-delta) > cap {
		if oldConf < cap {
			return ConfidenceDefault
		}
		return oldConf - cap
	}
	return newConf
}

func LowReputationCanPerformBasicTransfer(rep *IdentityReputation) bool {
	return true
}

func LowReputationCanPerformContractDeployment(rep *IdentityReputation) bool {
	return true
}

func LowReputationCanPerformContractExecution(rep *IdentityReputation) bool {
	return true
}

func LowReputationCanPerformPoolStaking(rep *IdentityReputation, stakeAmount uint64) bool {
	return stakeAmount >= MinStakeAmountForReputation
}

func ServiceTrustCannotMoveFunds(sts *ServiceTrustScore) bool {
	return true
}

func ServiceTrustCannotBypassFees(sts *ServiceTrustScore) bool {
	return true
}

func ComputeBoundedFeeAdjustmentNaet(baseFeeNaet uint64, score uint32, found bool, params ReputationEffectParams) uint64 {
	premiumBps, discountBps := ComputeBoundedFeeAdjustment(score, found, params)
	if premiumBps > 0 {
		premium := uint64(baseFeeNaet) * uint64(premiumBps) / 10_000
		return baseFeeNaet + premium
	}
	if discountBps > 0 {
		discount := uint64(baseFeeNaet) * uint64(discountBps) / 10_000
		adjusted := baseFeeNaet - discount
		minFee := uint64(baseFeeNaet) / 10
		if adjusted < minFee {
			adjusted = minFee
		}
		return adjusted
	}
	return baseFeeNaet
}
