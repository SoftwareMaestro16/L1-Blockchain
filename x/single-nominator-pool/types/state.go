package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

const (
	StatusActive	= "active"
	StatusClosed	= "closed"

	WithdrawalStatusNone		= ""
	WithdrawalStatusPending		= "pending"
	WithdrawalStatusCompleted	= "completed"

	MaxPoolsV1		= uint32(100_000)
	MaxPoolIDBytesV1	= uint32(96)
	DefaultUnbondingBlocks	= appparams.StakingUnbondingDefaultBlocks
)

type Params struct {
	Authority	string
	MaxPools	uint32
	MaxPoolIDBytes	uint32
	UnbondingBlocks	uint64
}

type State struct {
	Pools []SingleNominatorPool
}

type SingleNominatorPool struct {
	PoolAddress		string
	Owner			string
	Validator		string
	BondedStake		uint64
	PendingWithdrawal	PendingWithdrawal
	RewardBalance		uint64
	EmergencyLock		bool
	Status			string
}

type PendingWithdrawal struct {
	Amount		uint64
	RequestHeight	uint64
	CompleteHeight	uint64
	Status		string
}

type MsgCreateSingleNominatorPool struct {
	Authority	string
	PoolAddress	string
	Owner		string
	Validator	string
	ValidatorStatus	string
	Height		uint64
}

type MsgDepositSingleNominator struct {
	Authority	string
	PoolAddress	string
	Owner		string
	Amount		uint64
	Height		uint64
}

type MsgWithdrawSingleNominator struct {
	Authority	string
	PoolAddress	string
	Owner		string
	Amount		uint64
	Height		uint64
}

type MsgClaimSingleNominatorRewards struct {
	Authority	string
	PoolAddress	string
	Owner		string
	Height		uint64
}

type MsgEmergencyLockSingleNominator struct {
	Authority	string
	PoolAddress	string
	Owner		string
	Locked		bool
	Height		uint64
}

type MsgChangeSingleNominatorValidator struct {
	Authority	string
	PoolAddress	string
	Owner		string
	Validator	string
	ValidatorStatus	string
	Height		uint64
}

func DefaultParams() Params {
	return Params{
		Authority:		prototype.DefaultAuthority,
		MaxPools:		MaxPoolsV1,
		MaxPoolIDBytes:		MaxPoolIDBytesV1,
		UnbondingBlocks:	DefaultUnbondingBlocks,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("single nominator pool authority", p.Authority); err != nil {
		return err
	}
	if p.MaxPools == 0 || p.MaxPools > MaxPoolsV1 {
		return fmt.Errorf("single nominator pool max pools must be between 1 and %d", MaxPoolsV1)
	}
	if p.MaxPoolIDBytes == 0 || p.MaxPoolIDBytes > MaxPoolIDBytesV1 {
		return fmt.Errorf("single nominator pool max pool id bytes must be between 1 and %d", MaxPoolIDBytesV1)
	}
	if err := appparams.ValidateStakingUnbondingBlocks(p.UnbondingBlocks); err != nil {
		return fmt.Errorf("single nominator pool %w", err)
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("single nominator pool update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("single nominator pool update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Pools)) > params.MaxPools {
		return errors.New("single nominator pool count limit exceeded")
	}
	seen := map[string]struct{}{}
	for _, pool := range s.Pools {
		if err := pool.Validate(params); err != nil {
			return err
		}
		if _, found := seen[pool.PoolAddress]; found {
			return fmt.Errorf("duplicate single nominator pool %s", pool.PoolAddress)
		}
		seen[pool.PoolAddress] = struct{}{}
	}
	return nil
}

func (p SingleNominatorPool) Validate(params Params) error {
	if err := validatePoolAddress(p.PoolAddress, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("single nominator owner", p.Owner); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("single nominator validator", p.Validator); err != nil {
		return err
	}
	if !isStatus(p.Status) {
		return fmt.Errorf("unsupported single nominator pool status %q", p.Status)
	}
	if err := p.PendingWithdrawal.Validate(); err != nil {
		return err
	}
	return nil
}

func (w PendingWithdrawal) Validate() error {
	if w.Status == WithdrawalStatusNone {
		if w.Amount != 0 || w.RequestHeight != 0 || w.CompleteHeight != 0 {
			return errors.New("empty single nominator withdrawal must not carry values")
		}
		return nil
	}
	if w.Status != WithdrawalStatusPending && w.Status != WithdrawalStatusCompleted {
		return fmt.Errorf("unsupported single nominator withdrawal status %q", w.Status)
	}
	if w.Amount == 0 || w.RequestHeight == 0 || w.CompleteHeight <= w.RequestHeight {
		return errors.New("single nominator withdrawal amounts and heights are invalid")
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Pools = SortPools(s.Pools)
	return s
}

func SortPools(values []SingleNominatorPool) []SingleNominatorPool {
	out := append([]SingleNominatorPool(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolAddress < out[j].PoolAddress })
	return out
}

func IsJailedValidatorStatus(status string) bool {
	return status == validatorregistrytypes.StatusJailed || status == validatorregistrytypes.StatusTombstoned
}

func validatePoolAddress(value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("single nominator pool address is required")
	}
	if strings.HasPrefix(value, "4:") {
		return addressing.ValidateAuthorityAddress("single nominator pool address", value)
	}
	if uint32(len(value)) > maxBytes || strings.ContainsAny(value, " \t\r\n") {
		return errors.New("single nominator pool address must be address-like or bounded id")
	}
	return nil
}

func isStatus(status string) bool {
	return status == StatusActive || status == StatusClosed
}
