package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	StateValidatorStatusActive	= "active"
	StateValidatorStatusJailed	= "jailed"
	StateValidatorStatusSlashed	= "slashed"
	StateValidatorStatusRetired	= "retired"

	RentPayerPolicyPoolReserve	= "pool_reserve"

	SlashingFaultDowntime	= "downtime"
	SlashingFaultDoubleSign	= "double_sign"
)

type Validator struct {
	Address			string
	SelfStake		uint64
	NominatorStake		uint64
	Status			string
	PerformanceScore	uint32
	CommissionBps		uint32
	SlashingRiskBps		uint32
	AllocationLimitBps	uint32
	UpdatedHeight		uint64
	Jailed			bool
	Tombstoned		bool
}

type ValidatorPerformanceScore struct {
	Validator	string
	Epoch		uint64
	ScoreBps	uint32
}

type ValidatorCommission struct {
	Validator	string
	RateBps		uint32
	Epoch		uint64
}

type ValidatorSlashingRisk struct {
	Validator	string
	RiskBps		uint32
	Epoch		uint64
}

type ValidatorAllocationLimit struct {
	Validator	string
	LimitBps	uint32
	Epoch		uint64
}

type LiquidStakingPool struct {
	PoolID			string
	ContractAddressUser	string
	ContractAddressRaw	string
	ReceiptToken		string
	TotalDeposited		uint64
	TotalActiveStake	uint64
	TotalUnbonding		uint64
	TotalShares		uint64
	RewardIndex		uint64
	AllocationEpoch		uint64
	LastStorageChargeHeight	uint64
	StorageRentDebt		uint64
	StorageRentReserve	uint64
	RentPayerPolicy		string
	Status			string
}

type PoolShare struct {
	Owner			string
	PoolID			string
	Shares			uint64
	PrincipalAmount		uint64
	CreatedHeight		uint64
	UpdatedHeight		uint64
	LastRewardIndex		uint64
	PendingRewards		uint64
	StakeWeightedSeconds	uint64
	LastReputationUpdate	uint64
}

type PoolValidatorAllocation struct {
	PoolID			string
	Validator		string
	TargetWeightBps		uint32
	ActiveStake		uint64
	PendingStake		uint64
	UnbondingStake		uint64
	PerformanceScore	uint32
	CommissionBps		uint32
	SlashingRiskBps		uint32
	UpdatedHeight		uint64
}

type PoolUnbondingRequest struct {
	PoolID		string
	Owner		string
	RequestID	string
	Shares		uint64
	Amount		uint64
	RequestHeight	uint64
	CompleteHeight	uint64
	Status		string
}

type PoolRewardIndex struct {
	PoolID		string
	RewardIndex	uint64
	Epoch		uint64
}

type RewardClaim struct {
	PoolID	string
	Owner	string
	Epoch	uint64
	Amount	uint64
}

type EpochStakingSnapshot struct {
	Epoch			uint64
	TotalActiveStake	uint64
	TotalPools		uint64
	ValidatorCount		uint32
	SnapshotHash		string
}

type ValidatorSetSnapshot struct {
	HeightOrEpoch	uint64
	Validators	[]string
	TotalPower	uint64
	SnapshotHash	string
}

type ValidatorSlashEvent struct {
	Height			uint64
	Validator		string
	PoolID			string
	Fault			string
	Epoch			uint64
	SlashingLoss		uint64
	ValidatorStatus		string
	Tombstoned		bool
	PoolSlashIndexAfter	uint64
}

func (v Validator) Validate(params Params) error {
	if err := ValidateUserFacingAEAddress("staking validator", v.Address); err != nil {
		return err
	}
	if !isStateValidatorStatus(v.Status) {
		return fmt.Errorf("unsupported staking validator status %q", v.Status)
	}
	if v.PerformanceScore > MaxBasisPoints || v.CommissionBps > MaxBasisPoints || v.SlashingRiskBps > MaxBasisPoints || v.AllocationLimitBps > MaxBasisPoints {
		return errors.New("staking validator bps fields must be <= 10000")
	}
	if v.Status == StateValidatorStatusActive {
		mode := ValidatorFundingPoolBacked
		if v.NominatorStake == 0 {
			mode = ValidatorFundingSolo
		}
		return params.ValidateValidatorFunding(ValidatorFunding{
			Mode:		mode,
			SelfStake:	v.SelfStake,
			NominatorStake:	v.NominatorStake,
		})
	}
	return nil
}

func (s ValidatorPerformanceScore) Validate() error {
	if err := ValidateUserFacingAEAddress("validator performance score validator", s.Validator); err != nil {
		return err
	}
	if s.Epoch == 0 {
		return errors.New("validator performance score epoch must be positive")
	}
	if s.ScoreBps > MaxBasisPoints {
		return errors.New("validator performance score exceeds basis points")
	}
	return nil
}

func (c ValidatorCommission) Validate() error {
	if err := ValidateUserFacingAEAddress("validator commission validator", c.Validator); err != nil {
		return err
	}
	if c.Epoch == 0 {
		return errors.New("validator commission epoch must be positive")
	}
	if c.RateBps > MaxBasisPoints {
		return errors.New("validator commission exceeds basis points")
	}
	return nil
}

func (r ValidatorSlashingRisk) Validate() error {
	if err := ValidateUserFacingAEAddress("validator slashing risk validator", r.Validator); err != nil {
		return err
	}
	if r.Epoch == 0 {
		return errors.New("validator slashing risk epoch must be positive")
	}
	if r.RiskBps > MaxBasisPoints {
		return errors.New("validator slashing risk exceeds basis points")
	}
	return nil
}

func (l ValidatorAllocationLimit) Validate() error {
	if err := ValidateUserFacingAEAddress("validator allocation limit validator", l.Validator); err != nil {
		return err
	}
	if l.Epoch == 0 {
		return errors.New("validator allocation limit epoch must be positive")
	}
	if l.LimitBps > MaxBasisPoints {
		return errors.New("validator allocation limit exceeds basis points")
	}
	return nil
}

func (p LiquidStakingPool) Validate(params Params) error {
	if err := validateID("liquid staking pool id", p.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("liquid staking pool contract user address", p.ContractAddressUser); err != nil {
		return err
	}
	if err := ValidateRawAddress("liquid staking pool contract raw address", p.ContractAddressRaw); err != nil {
		return err
	}
	if err := ValidateAddressPair("liquid staking pool contract address pair", p.ContractAddressUser, p.ContractAddressRaw); err != nil {
		return err
	}
	if strings.TrimSpace(p.ReceiptToken) == "" {
		return errors.New("liquid staking pool receipt token is required")
	}
	if p.TotalActiveStake+p.TotalUnbonding > p.TotalDeposited {
		return errors.New("liquid staking pool active plus unbonding stake exceeds deposits")
	}
	if p.TotalShares == 0 && p.TotalDeposited > 0 {
		return errors.New("liquid staking pool deposited funds require shares")
	}
	if p.RentPayerPolicy == "" {
		return errors.New("liquid staking pool rent payer policy is required")
	}
	if !isPoolStatus(p.Status) {
		return fmt.Errorf("unsupported liquid staking pool status %q", p.Status)
	}
	return nil
}

func (s PoolShare) Validate(params Params) error {
	if err := ValidateUserFacingAEAddress("pool share owner", s.Owner); err != nil {
		return err
	}
	if err := validateID("pool share pool id", s.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if s.Shares == 0 || s.PrincipalAmount == 0 || s.CreatedHeight == 0 || s.UpdatedHeight < s.CreatedHeight {
		return errors.New("pool share amounts and heights are invalid")
	}
	return nil
}

func (a PoolValidatorAllocation) Validate(params Params, validator Validator) error {
	if err := validateID("pool allocation pool id", a.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("pool allocation validator", a.Validator); err != nil {
		return err
	}
	if validator.Address != a.Validator {
		return errors.New("pool allocation validator address mismatch")
	}
	if validator.Status != StateValidatorStatusActive {
		if a.TargetWeightBps != 0 {
			return errors.New("pool allocation for non-active validator must have zero target weight")
		}
		return nil
	}
	if validator.SlashingRiskBps >= MaxBasisPoints {
		return errors.New("pool allocation validator slashing risk is not eligible")
	}
	if a.TargetWeightBps > MaxBasisPoints {
		return errors.New("pool allocation target weight exceeds basis points")
	}
	if validator.AllocationLimitBps == 0 {
		return errors.New("pool allocation validator allocation limit is not eligible")
	}
	if a.PerformanceScore > MaxBasisPoints || a.CommissionBps > MaxBasisPoints || a.SlashingRiskBps > MaxBasisPoints {
		return errors.New("pool allocation bps fields must be <= 10000")
	}
	if a.UpdatedHeight == 0 {
		return errors.New("pool allocation updated height must be positive")
	}
	return nil
}

func (u PoolUnbondingRequest) Validate(params Params) error {
	if err := validateID("pool unbonding pool id", u.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("pool unbonding owner", u.Owner); err != nil {
		return err
	}
	if err := validateID("pool unbonding request id", u.RequestID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if u.Shares == 0 || u.Amount == 0 || u.RequestHeight == 0 || u.CompleteHeight <= u.RequestHeight {
		return errors.New("pool unbonding request amounts and heights are invalid")
	}
	if !isWithdrawalStatus(u.Status) {
		return fmt.Errorf("unsupported pool unbonding status %q", u.Status)
	}
	return nil
}

func (r PoolRewardIndex) Validate(params Params) error {
	if err := validateID("pool reward index pool id", r.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if r.Epoch == 0 {
		return errors.New("pool reward index epoch must be positive")
	}
	return nil
}

func (r RewardClaim) Validate(params Params) error {
	if err := validateID("reward claim pool id", r.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("reward claim owner", r.Owner); err != nil {
		return err
	}
	if r.Epoch == 0 {
		return errors.New("reward claim epoch must be positive")
	}
	return nil
}

func (s EpochStakingSnapshot) Validate() error {
	if s.Epoch == 0 {
		return errors.New("epoch staking snapshot epoch must be positive")
	}
	return nil
}

func (s ValidatorSetSnapshot) Validate() error {
	if s.HeightOrEpoch == 0 {
		return errors.New("validator set snapshot height or epoch must be positive")
	}
	ordered := append([]string(nil), s.Validators...)
	sort.Strings(ordered)
	for i, validator := range s.Validators {
		if err := ValidateUserFacingAEAddress("validator set snapshot validator", validator); err != nil {
			return err
		}
		if validator != ordered[i] {
			return errors.New("validator set snapshot validators must be sorted deterministically")
		}
		if i > 0 && validator == s.Validators[i-1] {
			return errors.New("validator set snapshot validators must be unique")
		}
	}
	return nil
}

func isStateValidatorStatus(status string) bool {
	switch status {
	case StateValidatorStatusActive, StateValidatorStatusJailed, StateValidatorStatusSlashed, StateValidatorStatusRetired:
		return true
	default:
		return false
	}
}
