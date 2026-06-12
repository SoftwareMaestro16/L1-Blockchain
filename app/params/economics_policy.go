package params

import "fmt"

type FeeSplitPolicy struct {
	BurnBps			int64
	ValidatorDelegatorBps	int64
	TreasuryBps		int64
	GovernanceConfigurable	bool
}

func DefaultFeeSplitPolicy() FeeSplitPolicy {
	return FeeSplitPolicy{
		BurnBps:		5_000,
		ValidatorDelegatorBps:	3_500,
		TreasuryBps:		1_500,
		GovernanceConfigurable:	true,
	}
}

func (p FeeSplitPolicy) Validate() error {
	if err := validateBps("fee_burn_bps", p.BurnBps, AetraFeeBurnShareMinBps, AetraFeeBurnShareMaxBps); err != nil {
		return err
	}
	if err := validateBps("fee_validator_delegator_bps", p.ValidatorDelegatorBps, AetraFeeRewardShareMinBps, AetraFeeRewardShareMaxBps); err != nil {
		return err
	}
	if err := validateBps("fee_treasury_bps", p.TreasuryBps, AetraFeeTreasuryShareMinBps, AetraFeeTreasuryShareMaxBps); err != nil {
		return err
	}
	if p.BurnBps+p.ValidatorDelegatorBps+p.TreasuryBps != BasisPoints {
		return fmt.Errorf("fee split must sum to %d bps", BasisPoints)
	}
	if !p.GovernanceConfigurable {
		return fmt.Errorf("fee split must be governance configurable within safe bounds")
	}
	return nil
}

func ApproximateStakingAPRBps(inflationBps, bondedRatioBps int64) (int64, error) {
	if err := validateBps("inflation_bps", inflationBps, 0, BasisPoints); err != nil {
		return 0, err
	}
	if err := validateBps("bonded_ratio_bps", bondedRatioBps, 1, BasisPoints); err != nil {
		return 0, err
	}
	return (inflationBps*BasisPoints + bondedRatioBps/2) / bondedRatioBps, nil
}
