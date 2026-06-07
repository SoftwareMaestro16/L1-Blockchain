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
	PoolStatusActive = "active"
	PoolStatusPaused = "paused"
	PoolStatusClosed = "closed"

	WithdrawalStatusPending   = "pending"
	WithdrawalStatusCancelled = "cancelled"
	WithdrawalStatusCompleted = "completed"

	MaxPoolsV1                  = uint32(10_000)
	MaxDelegatorsV1             = uint32(1_000_000)
	MaxPendingDepositsV1        = uint32(1_000_000)
	MaxPendingWithdrawalsV1     = uint32(1_000_000)
	MaxUnbondingEntriesV1       = uint32(1_000_000)
	MaxPoolIDBytesV1            = uint32(96)
	MaxBasisPoints              = uint32(10_000)
	IndexScale                  = uint64(1_000_000_000)
	DefaultMaxCommissionBps     = uint32(2_000)
	DefaultUnbondingBlocks      = appparams.StakingUnbondingDefaultBlocks
	DefaultValidatorChangeDelay = uint64(100)
)

type Params struct {
	Authority             string
	MaxPools              uint32
	MaxDelegators         uint32
	MaxPendingDeposits    uint32
	MaxPendingWithdrawals uint32
	MaxUnbondingEntries   uint32
	MaxPoolIDBytes        uint32
	MaxCommissionBps      uint32
	UnbondingBlocks       uint64
	ValidatorChangeDelay  uint64
}

type State struct {
	Pools []NominatorPool
}

type NominatorPool struct {
	PoolID                 string
	PoolOperator           string
	ValidatorTarget        string
	PendingValidatorTarget string
	ValidatorChangeHeight  uint64
	TotalShares            uint64
	TotalBondedStake       uint64
	PendingDeposits        []PendingDeposit
	PendingWithdrawals     []PendingWithdrawal
	DelegatorShares        []DelegatorShare
	RewardIndex            uint64
	SlashIndex             uint64
	PoolCommissionBps      uint32
	Status                 string
	UnbondingQueue         []UnbondingEntry
}

type PendingDeposit struct {
	Delegator string
	Amount    uint64
	Height    uint64
}

type PendingWithdrawal struct {
	WithdrawalID   string
	Delegator      string
	Shares         uint64
	Amount         uint64
	RequestHeight  uint64
	CompleteHeight uint64
	Status         string
}

type DelegatorShare struct {
	Delegator             string
	Shares                uint64
	RewardIndexCheckpoint uint64
	PendingRewards        uint64
	SlashIndexCheckpoint  uint64
}

type UnbondingEntry struct {
	WithdrawalID   string
	Delegator      string
	Amount         uint64
	CompleteHeight uint64
	Status         string
}

type MsgCreateNominatorPool struct {
	Authority         string
	PoolID            string
	PoolOperator      string
	ValidatorTarget   string
	PoolCommissionBps uint32
	Height            uint64
	ValidatorStatus   string
}

type MsgDepositToPool struct {
	Authority string
	PoolID    string
	Delegator string
	Amount    uint64
	Height    uint64
}

type MsgRequestPoolWithdrawal struct {
	Authority    string
	PoolID       string
	WithdrawalID string
	Delegator    string
	Shares       uint64
	Height       uint64
}

type MsgCancelPoolWithdrawal struct {
	Authority    string
	PoolID       string
	WithdrawalID string
	Delegator    string
	Height       uint64
}

type MsgClaimPoolRewards struct {
	Authority string
	PoolID    string
	Delegator string
	Height    uint64
}

type MsgUpdatePoolCommission struct {
	Authority         string
	PoolID            string
	PoolOperator      string
	PoolCommissionBps uint32
	Height            uint64
}

type MsgChangePoolValidator struct {
	Authority       string
	PoolID          string
	PoolOperator    string
	ValidatorTarget string
	ValidatorStatus string
	Height          uint64
}

func DefaultParams() Params {
	return Params{
		Authority:             prototype.DefaultAuthority,
		MaxPools:              MaxPoolsV1,
		MaxDelegators:         MaxDelegatorsV1,
		MaxPendingDeposits:    MaxPendingDepositsV1,
		MaxPendingWithdrawals: MaxPendingWithdrawalsV1,
		MaxUnbondingEntries:   MaxUnbondingEntriesV1,
		MaxPoolIDBytes:        MaxPoolIDBytesV1,
		MaxCommissionBps:      DefaultMaxCommissionBps,
		UnbondingBlocks:       DefaultUnbondingBlocks,
		ValidatorChangeDelay:  DefaultValidatorChangeDelay,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool authority", p.Authority); err != nil {
		return err
	}
	if p.MaxPools == 0 || p.MaxPools > MaxPoolsV1 {
		return fmt.Errorf("nominator pool max pools must be between 1 and %d", MaxPoolsV1)
	}
	if p.MaxDelegators == 0 || p.MaxDelegators > MaxDelegatorsV1 {
		return fmt.Errorf("nominator pool max delegators must be between 1 and %d", MaxDelegatorsV1)
	}
	if p.MaxPendingDeposits == 0 || p.MaxPendingDeposits > MaxPendingDepositsV1 {
		return fmt.Errorf("nominator pool max pending deposits must be between 1 and %d", MaxPendingDepositsV1)
	}
	if p.MaxPendingWithdrawals == 0 || p.MaxPendingWithdrawals > MaxPendingWithdrawalsV1 {
		return fmt.Errorf("nominator pool max pending withdrawals must be between 1 and %d", MaxPendingWithdrawalsV1)
	}
	if p.MaxUnbondingEntries == 0 || p.MaxUnbondingEntries > MaxUnbondingEntriesV1 {
		return fmt.Errorf("nominator pool max unbonding entries must be between 1 and %d", MaxUnbondingEntriesV1)
	}
	if p.MaxPoolIDBytes == 0 || p.MaxPoolIDBytes > MaxPoolIDBytesV1 {
		return fmt.Errorf("nominator pool max pool id bytes must be between 1 and %d", MaxPoolIDBytesV1)
	}
	if p.MaxCommissionBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max commission must be <= %d", MaxBasisPoints)
	}
	if err := appparams.ValidateStakingUnbondingBlocks(p.UnbondingBlocks); err != nil {
		return fmt.Errorf("nominator pool %w", err)
	}
	if p.ValidatorChangeDelay == 0 {
		return errors.New("nominator pool validator change delay must be positive")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("nominator pool update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Pools)) > params.MaxPools {
		return errors.New("nominator pool count limit exceeded")
	}
	ids := map[string]struct{}{}
	for _, pool := range s.Pools {
		if err := pool.Validate(params); err != nil {
			return err
		}
		if _, found := ids[pool.PoolID]; found {
			return fmt.Errorf("duplicate nominator pool id %s", pool.PoolID)
		}
		ids[pool.PoolID] = struct{}{}
	}
	return nil
}

func (p NominatorPool) Validate(params Params) error {
	if err := validateID("nominator pool id", p.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool operator", p.PoolOperator); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool validator target", p.ValidatorTarget); err != nil {
		return err
	}
	if strings.TrimSpace(p.PendingValidatorTarget) != "" {
		if err := addressing.ValidateAuthorityAddress("nominator pool pending validator target", p.PendingValidatorTarget); err != nil {
			return err
		}
		if p.ValidatorChangeHeight == 0 {
			return errors.New("nominator pool pending validator change requires activation height")
		}
	}
	if p.PoolCommissionBps > params.MaxCommissionBps {
		return errors.New("nominator pool commission exceeds configured bound")
	}
	if !isPoolStatus(p.Status) {
		return fmt.Errorf("unsupported nominator pool status %q", p.Status)
	}
	if uint32(len(p.DelegatorShares)) > params.MaxDelegators {
		return errors.New("nominator pool delegator limit exceeded")
	}
	if uint32(len(p.PendingDeposits)) > params.MaxPendingDeposits {
		return errors.New("nominator pool pending deposit limit exceeded")
	}
	if uint32(len(p.PendingWithdrawals)) > params.MaxPendingWithdrawals {
		return errors.New("nominator pool pending withdrawal limit exceeded")
	}
	if uint32(len(p.UnbondingQueue)) > params.MaxUnbondingEntries {
		return errors.New("nominator pool unbonding queue limit exceeded")
	}
	if p.TotalShares != sumShares(p.DelegatorShares) {
		return errors.New("nominator pool total shares do not match delegator shares")
	}
	delegators := map[string]struct{}{}
	for _, delegator := range p.DelegatorShares {
		if err := delegator.Validate(); err != nil {
			return err
		}
		if _, found := delegators[delegator.Delegator]; found {
			return fmt.Errorf("duplicate pool delegator %s", delegator.Delegator)
		}
		delegators[delegator.Delegator] = struct{}{}
	}
	withdrawals := map[string]struct{}{}
	for _, withdrawal := range p.PendingWithdrawals {
		if err := withdrawal.Validate(); err != nil {
			return err
		}
		if _, found := withdrawals[withdrawal.WithdrawalID]; found {
			return fmt.Errorf("duplicate pool withdrawal %s", withdrawal.WithdrawalID)
		}
		withdrawals[withdrawal.WithdrawalID] = struct{}{}
		if withdrawal.Status == WithdrawalStatusPending {
			if _, found := delegators[withdrawal.Delegator]; !found {
				return fmt.Errorf("withdrawal %s references unknown delegator", withdrawal.WithdrawalID)
			}
		}
	}
	for _, deposit := range p.PendingDeposits {
		if err := deposit.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range p.UnbondingQueue {
		if err := entry.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d PendingDeposit) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool pending deposit delegator", d.Delegator); err != nil {
		return err
	}
	if d.Amount == 0 || d.Height == 0 {
		return errors.New("nominator pool pending deposit amount and height must be positive")
	}
	return nil
}

func (w PendingWithdrawal) Validate() error {
	if err := validateID("nominator pool withdrawal id", w.WithdrawalID, MaxPoolIDBytesV1); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool withdrawal delegator", w.Delegator); err != nil {
		return err
	}
	if w.Shares == 0 || w.Amount == 0 || w.RequestHeight == 0 || w.CompleteHeight <= w.RequestHeight {
		return errors.New("nominator pool withdrawal amounts and heights are invalid")
	}
	if !isWithdrawalStatus(w.Status) {
		return fmt.Errorf("unsupported nominator pool withdrawal status %q", w.Status)
	}
	return nil
}

func (d DelegatorShare) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool delegator", d.Delegator); err != nil {
		return err
	}
	if d.Shares == 0 {
		return errors.New("nominator pool delegator shares must be positive")
	}
	return nil
}

func (e UnbondingEntry) Validate() error {
	if err := validateID("nominator pool unbonding withdrawal id", e.WithdrawalID, MaxPoolIDBytesV1); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool unbonding delegator", e.Delegator); err != nil {
		return err
	}
	if e.Amount == 0 || e.CompleteHeight == 0 {
		return errors.New("nominator pool unbonding amount and completion height must be positive")
	}
	if !isWithdrawalStatus(e.Status) {
		return fmt.Errorf("unsupported nominator pool unbonding status %q", e.Status)
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Pools = SortPools(s.Pools)
	for idx := range s.Pools {
		s.Pools[idx].PendingDeposits = SortDeposits(s.Pools[idx].PendingDeposits)
		s.Pools[idx].PendingWithdrawals = SortWithdrawals(s.Pools[idx].PendingWithdrawals)
		s.Pools[idx].DelegatorShares = SortDelegators(s.Pools[idx].DelegatorShares)
		s.Pools[idx].UnbondingQueue = SortUnbonding(s.Pools[idx].UnbondingQueue)
	}
	return s
}

func SortPools(values []NominatorPool) []NominatorPool {
	out := append([]NominatorPool(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func SortDelegators(values []DelegatorShare) []DelegatorShare {
	out := append([]DelegatorShare(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Delegator < out[j].Delegator })
	return out
}

func SortWithdrawals(values []PendingWithdrawal) []PendingWithdrawal {
	out := append([]PendingWithdrawal(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].WithdrawalID < out[j].WithdrawalID })
	return out
}

func SortDeposits(values []PendingDeposit) []PendingDeposit {
	out := append([]PendingDeposit(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].Delegator < out[j].Delegator
	})
	return out
}

func SortUnbonding(values []UnbondingEntry) []UnbondingEntry {
	out := append([]UnbondingEntry(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CompleteHeight != out[j].CompleteHeight {
			return out[i].CompleteHeight < out[j].CompleteHeight
		}
		return out[i].WithdrawalID < out[j].WithdrawalID
	})
	return out
}

func ShareValue(pool NominatorPool, shares uint64) uint64 {
	if pool.TotalShares == 0 {
		return 0
	}
	return shares * pool.TotalBondedStake / pool.TotalShares
}

func SharesForDeposit(pool NominatorPool, amount uint64) uint64 {
	if pool.TotalShares == 0 || pool.TotalBondedStake == 0 {
		return amount
	}
	shares := amount * pool.TotalShares / pool.TotalBondedStake
	if shares == 0 && amount > 0 {
		return 1
	}
	return shares
}

func RewardDelta(amount uint64, totalShares uint64) uint64 {
	if amount == 0 || totalShares == 0 {
		return 0
	}
	return amount * IndexScale / totalShares
}

func AccruedReward(delegator DelegatorShare, rewardIndex uint64) uint64 {
	if rewardIndex <= delegator.RewardIndexCheckpoint {
		return delegator.PendingRewards
	}
	return delegator.PendingRewards + delegator.Shares*(rewardIndex-delegator.RewardIndexCheckpoint)/IndexScale
}

func IsJailedValidatorStatus(status string) bool {
	return status == validatorregistrytypes.StatusJailed || status == validatorregistrytypes.StatusTombstoned
}

func validateID(field, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if uint32(len(value)) > maxBytes || strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s must be non-blank, whitespace-free, and within configured length", field)
	}
	return nil
}

func isPoolStatus(status string) bool {
	return status == PoolStatusActive || status == PoolStatusPaused || status == PoolStatusClosed
}

func isWithdrawalStatus(status string) bool {
	return status == WithdrawalStatusPending || status == WithdrawalStatusCancelled || status == WithdrawalStatusCompleted
}

func sumShares(values []DelegatorShare) uint64 {
	total := uint64(0)
	for _, value := range values {
		total += value.Shares
	}
	return total
}
