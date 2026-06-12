package params

import "fmt"

const (
	StakingPolicyDenom	= BaseDenom

	StakingUnbondingBlockTimeSeconds	= uint64(6)
	StakingUnbondingMinDays			= uint64(14)
	StakingUnbondingMaxDays			= uint64(21)
	StakingUnbondingDefaultDays		= StakingUnbondingMinDays

	StakingUnbondingMinBlocks	= StakingUnbondingMinDays * 24 * 60 * 60 / StakingUnbondingBlockTimeSeconds
	StakingUnbondingMaxBlocks	= StakingUnbondingMaxDays * 24 * 60 * 60 / StakingUnbondingBlockTimeSeconds
	StakingUnbondingDefaultBlocks	= StakingUnbondingDefaultDays * 24 * 60 * 60 / StakingUnbondingBlockTimeSeconds

	StakingMinSelfBondAET		= int64(10_000)
	StakingMinValidatorBondAET	= int64(50_000)
	StakingMinSelfBondNaet		= StakingMinSelfBondAET * BaseUnitsPerDisplay
	StakingMinValidatorBondNaet	= StakingMinValidatorBondAET * BaseUnitsPerDisplay
	StakingCommissionFloorBps	= int64(300)
	StakingCommissionCeilingBps	= MaxCommissionBps
	StakingMaxDailyCommissionBps	= MaxDailyCommissionChangeBps
)

type StakingDelegationPolicy struct {
	Denom				string
	MinSelfBondNaet			int64
	MinValidatorBondNaet		int64
	MinCommissionBps		int64
	MaxCommissionBps		int64
	MaxDailyCommissionBps		int64
	UnbondingMinBlocks		uint64
	UnbondingMaxBlocks		uint64
	UnbondingDefaultBlocks		uint64
	DelegationEnabled		bool
	RedelegationEnabled		bool
	NominationPoolsEnabled		bool
	SlashingInherited		bool
	RequireValidatorMetadata	bool
}

type AntiCartelPolicy struct {
	CommissionFloorBps		int64
	CommissionCeilingBps		int64
	MaxDailyCommissionChangeBps	int64
	ValidatorIdentityRegistry	bool
	MandatoryKYC			bool
	ValidatorMetadataTransparency	bool
	PublicConcentrationMetrics	bool
	SelfBondRatioVisibility		bool
	ObjectiveCorrelationWarnings	bool
	EconomicSignalsInsteadOfHalting	bool
}

func DefaultStakingDelegationPolicy() StakingDelegationPolicy {
	return StakingDelegationPolicy{
		Denom:				StakingPolicyDenom,
		MinSelfBondNaet:		StakingMinSelfBondNaet,
		MinValidatorBondNaet:		StakingMinValidatorBondNaet,
		MinCommissionBps:		StakingCommissionFloorBps,
		MaxCommissionBps:		StakingCommissionCeilingBps,
		MaxDailyCommissionBps:		StakingMaxDailyCommissionBps,
		UnbondingMinBlocks:		StakingUnbondingMinBlocks,
		UnbondingMaxBlocks:		StakingUnbondingMaxBlocks,
		UnbondingDefaultBlocks:		StakingUnbondingDefaultBlocks,
		DelegationEnabled:		true,
		RedelegationEnabled:		true,
		NominationPoolsEnabled:		true,
		SlashingInherited:		true,
		RequireValidatorMetadata:	true,
	}
}

func DefaultAntiCartelPolicy() AntiCartelPolicy {
	return AntiCartelPolicy{
		CommissionFloorBps:			StakingCommissionFloorBps,
		CommissionCeilingBps:			StakingCommissionCeilingBps,
		MaxDailyCommissionChangeBps:		StakingMaxDailyCommissionBps,
		ValidatorIdentityRegistry:		true,
		MandatoryKYC:				false,
		ValidatorMetadataTransparency:		true,
		PublicConcentrationMetrics:		true,
		SelfBondRatioVisibility:		true,
		ObjectiveCorrelationWarnings:		true,
		EconomicSignalsInsteadOfHalting:	true,
	}
}

func (p StakingDelegationPolicy) Validate() error {
	if p.Denom != BaseDenom {
		return fmt.Errorf("staking denom must be %s", BaseDenom)
	}
	if p.MinSelfBondNaet <= 0 {
		return fmt.Errorf("minimum self-bond must be positive")
	}
	if p.MinValidatorBondNaet <= p.MinSelfBondNaet {
		return fmt.Errorf("minimum validator bond must be higher than minimum self-bond")
	}
	if err := validateBpsRange("validator_commission", p.MinCommissionBps, p.MaxCommissionBps, StakingCommissionFloorBps, StakingCommissionCeilingBps); err != nil {
		return err
	}
	if p.MaxDailyCommissionBps <= 0 || p.MaxDailyCommissionBps > StakingMaxDailyCommissionBps {
		return fmt.Errorf("daily commission change must stay within configured commission bounds")
	}
	if err := ValidateStakingUnbondingBlocks(p.UnbondingDefaultBlocks); err != nil {
		return err
	}
	if p.UnbondingMinBlocks != StakingUnbondingMinBlocks || p.UnbondingMaxBlocks != StakingUnbondingMaxBlocks {
		return fmt.Errorf("staking unbonding bounds must stay within 14-21 days")
	}
	if !p.DelegationEnabled || !p.RedelegationEnabled || !p.NominationPoolsEnabled {
		return fmt.Errorf("delegation, redelegation, and nomination pools must be enabled")
	}
	if !p.SlashingInherited {
		return fmt.Errorf("delegators and nomination pools must inherit validator slashing risk")
	}
	if !p.RequireValidatorMetadata {
		return fmt.Errorf("validator metadata must be required")
	}
	return nil
}

func (p AntiCartelPolicy) Validate() error {
	if err := validateBpsRange("anti_cartel_commission", p.CommissionFloorBps, p.CommissionCeilingBps, 300, 2_000); err != nil {
		return err
	}
	if p.MaxDailyCommissionChangeBps < 50 || p.MaxDailyCommissionChangeBps > StakingMaxDailyCommissionBps {
		return fmt.Errorf("anti-cartel max daily commission change must stay within 0.5-1 percent")
	}
	if !p.ValidatorIdentityRegistry {
		return fmt.Errorf("anti-cartel policy requires optional validator identity registry")
	}
	if p.MandatoryKYC {
		return fmt.Errorf("anti-cartel policy must not require mandatory KYC")
	}
	if !p.ValidatorMetadataTransparency || !p.PublicConcentrationMetrics || !p.SelfBondRatioVisibility {
		return fmt.Errorf("anti-cartel policy requires public validator transparency metrics")
	}
	if !p.ObjectiveCorrelationWarnings {
		return fmt.Errorf("operator correlation warnings must be based on public objective evidence")
	}
	if !p.EconomicSignalsInsteadOfHalting {
		return fmt.Errorf("anti-cartel policy must use economic signals instead of halting staking")
	}
	return nil
}

func ValidateStakingUnbondingBlocks(blocks uint64) error {
	if blocks < StakingUnbondingMinBlocks || blocks > StakingUnbondingMaxBlocks {
		return fmt.Errorf("staking unbonding blocks must represent 14-21 days")
	}
	return nil
}

func (p StakingDelegationPolicy) ValidateValidatorBond(selfBondNaet, validatorBondNaet int64) error {
	if selfBondNaet < p.MinSelfBondNaet {
		return fmt.Errorf("validator self-bond below policy minimum")
	}
	if validatorBondNaet < p.MinValidatorBondNaet {
		return fmt.Errorf("validator bond below policy minimum")
	}
	if validatorBondNaet < selfBondNaet {
		return fmt.Errorf("validator bond cannot be lower than self-bond")
	}
	return nil
}
