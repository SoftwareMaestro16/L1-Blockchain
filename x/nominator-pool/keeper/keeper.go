package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version uint64
	Params  types.Params
	State   types.State
}

type Keeper struct {
	genesis      GenesisState
	storeService corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{Version: prototype.CurrentGenesisVersion, Params: params, State: types.State{}.Normalize(params)}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("nominator pool unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) CreateNominatorPool(msg types.MsgCreateNominatorPool) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("nominator pool creation height must be positive")
	}
	if types.IsJailedValidatorStatus(msg.ValidatorStatus) {
		return types.NominatorPool{}, errors.New("nominator pool cannot delegate to jailed validator")
	}
	if _, _, found := findPool(k.genesis.State.Pools, msg.PoolID); found {
		return types.NominatorPool{}, errors.New("nominator pool already exists")
	}
	pool := types.NominatorPool{
		PoolID:            msg.PoolID,
		PoolOperator:      msg.PoolOperator,
		ValidatorTarget:   msg.ValidatorTarget,
		PoolCommissionBps: msg.PoolCommissionBps,
		Status:            types.PoolStatusActive,
	}
	if err := pool.Validate(k.genesis.Params); err != nil {
		return types.NominatorPool{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools = append(next.State.Pools, pool)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	k.genesis = next
	return pool, nil
}

func (k *Keeper) CreateOfficialLiquidStakingPool(msg types.MsgCreateOfficialLiquidStakingPool) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("official liquid staking pool creation height must be positive")
	}
	if _, _, found := findPool(k.genesis.State.Pools, msg.PoolID); found {
		return types.NominatorPool{}, errors.New("official liquid staking pool already exists")
	}
	pool := types.NominatorPool{
		PoolID:                msg.PoolID,
		ContractAddressUser:   msg.ContractAddressUser,
		ContractAddressRaw:    msg.ContractAddressRaw,
		OfficialLiquidStaking: true,
		PoolOperator:          msg.PoolOperator,
		PoolCommissionBps:     msg.PoolCommissionBps,
		Status:                types.PoolStatusActive,
	}
	if err := pool.Validate(k.genesis.Params); err != nil {
		return types.NominatorPool{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools = append(next.State.Pools, pool)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	k.genesis = next
	return pool, nil
}

func (k *Keeper) DepositToPool(msg types.MsgDepositToPool) (types.DelegatorShare, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.DelegatorShare{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.DelegatorShare{}, errors.New("nominator pool deposit amount and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.DelegatorShare{}, errors.New("nominator pool not found")
	}
	if pool.Status != types.PoolStatusActive {
		return types.DelegatorShare{}, errors.New("nominator pool must be active for deposits")
	}
	shareAmount := types.SharesForDeposit(pool, msg.Amount)
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if found {
		delegator.PendingRewards = types.AccruedReward(delegator, pool.RewardIndex)
		delegator.Shares += shareAmount
		delegator.RewardIndexCheckpoint = pool.RewardIndex
		delegator.SlashIndexCheckpoint = pool.SlashIndex
		pool.DelegatorShares[delegatorIdx] = delegator
	} else {
		delegator = types.DelegatorShare{
			Delegator:             msg.Delegator,
			Shares:                shareAmount,
			RewardIndexCheckpoint: pool.RewardIndex,
			SlashIndexCheckpoint:  pool.SlashIndex,
		}
		pool.DelegatorShares = append(pool.DelegatorShares, delegator)
	}
	pool.TotalShares += shareAmount
	pool.TotalBondedStake += msg.Amount
	pool.PendingDeposits = append(pool.PendingDeposits, types.PendingDeposit{Delegator: msg.Delegator, Amount: msg.Amount, Height: msg.Height})
	return k.savePool(idx, pool, delegator)
}

func (k *Keeper) DepositToOfficialLiquidStaking(msg types.MsgDepositToOfficialLiquidStaking) (types.DelegatorShare, error) {
	if err := types.ValidateOfficialLiquidStakingDeposit(msg, k.genesis.Params); err != nil {
		return types.DelegatorShare{}, err
	}
	rawUserAddress, err := types.RawAddressForUserAddress(msg.UserAddress)
	if err != nil {
		return types.DelegatorShare{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.DelegatorShare{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking {
		return types.DelegatorShare{}, errors.New("pool is not an official liquid staking pool")
	}
	if pool.Status != types.PoolStatusActive {
		return types.DelegatorShare{}, errors.New("official liquid staking pool must be active for deposits")
	}
	shareAmount := types.SharesForDeposit(pool, msg.Amount)
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, rawUserAddress)
	if found {
		delegator.PendingRewards = types.AccruedReward(delegator, pool.RewardIndex)
		delegator.Shares += shareAmount
		delegator.RewardIndexCheckpoint = pool.RewardIndex
		delegator.SlashIndexCheckpoint = pool.SlashIndex
		pool.DelegatorShares[delegatorIdx] = delegator
	} else {
		delegator = types.DelegatorShare{
			Delegator:             rawUserAddress,
			Shares:                shareAmount,
			RewardIndexCheckpoint: pool.RewardIndex,
			SlashIndexCheckpoint:  pool.SlashIndex,
		}
		pool.DelegatorShares = append(pool.DelegatorShares, delegator)
	}
	pool.TotalShares += shareAmount
	pool.TotalBondedStake += msg.Amount
	pool.PendingDeposits = append(pool.PendingDeposits, types.PendingDeposit{Delegator: rawUserAddress, Amount: msg.Amount, Height: msg.Height})
	return k.savePool(idx, pool, delegator)
}

func (k *Keeper) DelegateUserToValidator(msg types.MsgDelegateToValidator) error {
	return types.ValidateDirectUserDelegation(msg, k.genesis.Params)
}

func (k *Keeper) InjectPooledStake(msg types.MsgInjectPooledStake) (types.NominatorPool, error) {
	if err := types.ValidateUserFacingAEAddress("pooled stake caller contract", msg.CallerContractUser); err != nil {
		return types.NominatorPool{}, err
	}
	if err := types.ValidateUserFacingAEAddress("pooled stake validator address", msg.ValidatorAddress); err != nil {
		return types.NominatorPool{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.NominatorPool{}, errors.New("pooled stake injection amount and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("official liquid staking pool not found")
	}
	if !pool.OfficialLiquidStaking || pool.ContractAddressUser != msg.CallerContractUser {
		return types.NominatorPool{}, errors.New("pooled stake injection requires official liquid staking contract")
	}
	if pool.Status != types.PoolStatusActive {
		return types.NominatorPool{}, errors.New("official liquid staking pool must be active for stake injection")
	}
	currentAllocated := totalAllocated(pool.Allocations)
	if msg.Amount > pool.TotalBondedStake-currentAllocated {
		return types.NominatorPool{}, errors.New("pooled stake injection exceeds unallocated pool stake")
	}
	allocationIdx, allocation, found := findAllocation(pool.Allocations, msg.ValidatorAddress)
	if found {
		allocation.Amount += msg.Amount
		allocation.Height = msg.Height
		pool.Allocations[allocationIdx] = allocation
	} else {
		pool.Allocations = append(pool.Allocations, types.PoolAllocation{
			ValidatorAddress: msg.ValidatorAddress,
			Amount:           msg.Amount,
			Height:           msg.Height,
		})
	}
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) RequestPoolWithdrawal(msg types.MsgRequestPoolWithdrawal) (types.PendingWithdrawal, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingWithdrawal{}, err
	}
	if msg.Shares == 0 || msg.Height == 0 {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal shares and height must be positive")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool not found")
	}
	if _, _, found := findWithdrawal(pool.PendingWithdrawals, msg.WithdrawalID); found {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal already exists")
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool delegator not found")
	}
	if msg.Shares > delegator.Shares || msg.Shares > pool.TotalShares {
		return types.PendingWithdrawal{}, errors.New("nominator pool cannot withdraw more than total stake")
	}
	reward := types.AccruedReward(delegator, pool.RewardIndex)
	amount := types.ShareValue(pool, msg.Shares)
	if amount == 0 || amount > pool.TotalBondedStake {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal amount exceeds bonded stake")
	}
	delegator.Shares -= msg.Shares
	delegator.PendingRewards = reward
	delegator.RewardIndexCheckpoint = pool.RewardIndex
	delegator.SlashIndexCheckpoint = pool.SlashIndex
	pool.TotalShares -= msg.Shares
	pool.TotalBondedStake -= amount
	if delegator.Shares == 0 && delegator.PendingRewards == 0 {
		pool.DelegatorShares = append(pool.DelegatorShares[:delegatorIdx], pool.DelegatorShares[delegatorIdx+1:]...)
	} else {
		pool.DelegatorShares[delegatorIdx] = delegator
	}
	withdrawal := types.PendingWithdrawal{
		WithdrawalID:   msg.WithdrawalID,
		Delegator:      msg.Delegator,
		Shares:         msg.Shares,
		Amount:         amount,
		RequestHeight:  msg.Height,
		CompleteHeight: msg.Height + k.genesis.Params.UnbondingBlocks,
		Status:         types.WithdrawalStatusPending,
	}
	pool.PendingWithdrawals = append(pool.PendingWithdrawals, withdrawal)
	pool.UnbondingQueue = append(pool.UnbondingQueue, types.UnbondingEntry{
		WithdrawalID:   withdrawal.WithdrawalID,
		Delegator:      withdrawal.Delegator,
		Amount:         withdrawal.Amount,
		CompleteHeight: withdrawal.CompleteHeight,
		Status:         withdrawal.Status,
	})
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PendingWithdrawal{}, err
	}
	k.genesis = next
	return withdrawal, nil
}

func (k *Keeper) CancelPoolWithdrawal(msg types.MsgCancelPoolWithdrawal) (types.PendingWithdrawal, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingWithdrawal{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool not found")
	}
	withdrawalIdx, withdrawal, found := findWithdrawal(pool.PendingWithdrawals, msg.WithdrawalID)
	if !found {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal not found")
	}
	if withdrawal.Delegator != msg.Delegator {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal delegator mismatch")
	}
	if withdrawal.Status != types.WithdrawalStatusPending {
		return types.PendingWithdrawal{}, errors.New("nominator pool withdrawal is not pending")
	}
	withdrawal.Status = types.WithdrawalStatusCancelled
	pool.PendingWithdrawals[withdrawalIdx] = withdrawal
	for entryIdx, entry := range pool.UnbondingQueue {
		if entry.WithdrawalID == msg.WithdrawalID {
			entry.Status = types.WithdrawalStatusCancelled
			pool.UnbondingQueue[entryIdx] = entry
		}
	}
	shares := types.SharesForDeposit(pool, withdrawal.Amount)
	if shares < withdrawal.Shares {
		shares = withdrawal.Shares
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if found {
		delegator.PendingRewards = types.AccruedReward(delegator, pool.RewardIndex)
		delegator.Shares += shares
		delegator.RewardIndexCheckpoint = pool.RewardIndex
		pool.DelegatorShares[delegatorIdx] = delegator
	} else {
		pool.DelegatorShares = append(pool.DelegatorShares, types.DelegatorShare{
			Delegator:             msg.Delegator,
			Shares:                shares,
			RewardIndexCheckpoint: pool.RewardIndex,
			SlashIndexCheckpoint:  pool.SlashIndex,
		})
	}
	pool.TotalShares += shares
	pool.TotalBondedStake += withdrawal.Amount
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PendingWithdrawal{}, err
	}
	k.genesis = next
	return withdrawal, nil
}

func (k *Keeper) ClaimPoolRewards(msg types.MsgClaimPoolRewards) (uint64, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return 0, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return 0, errors.New("nominator pool not found")
	}
	delegatorIdx, delegator, found := findDelegator(pool.DelegatorShares, msg.Delegator)
	if !found {
		return 0, errors.New("nominator pool delegator not found")
	}
	reward := types.AccruedReward(delegator, pool.RewardIndex)
	delegator.PendingRewards = 0
	delegator.RewardIndexCheckpoint = pool.RewardIndex
	pool.DelegatorShares[delegatorIdx] = delegator
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return 0, err
	}
	k.genesis = next
	return reward, nil
}

func (k *Keeper) UpdatePoolCommission(msg types.MsgUpdatePoolCommission) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if pool.PoolOperator != msg.PoolOperator {
		return types.NominatorPool{}, errors.New("nominator pool operator mismatch")
	}
	pool.PoolCommissionBps = msg.PoolCommissionBps
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ChangePoolValidator(msg types.MsgChangePoolValidator) (types.NominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.NominatorPool{}, err
	}
	if types.IsJailedValidatorStatus(msg.ValidatorStatus) {
		return types.NominatorPool{}, errors.New("nominator pool cannot delegate to jailed validator")
	}
	idx, pool, found := findPool(k.genesis.State.Pools, msg.PoolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if pool.PoolOperator != msg.PoolOperator {
		return types.NominatorPool{}, errors.New("nominator pool operator mismatch")
	}
	if msg.Height == 0 {
		return types.NominatorPool{}, errors.New("nominator pool validator change height must be positive")
	}
	if pool.PendingValidatorTarget == msg.ValidatorTarget && msg.Height >= pool.ValidatorChangeHeight {
		pool.ValidatorTarget = msg.ValidatorTarget
		pool.PendingValidatorTarget = ""
		pool.ValidatorChangeHeight = 0
	} else if pool.PendingValidatorTarget == msg.ValidatorTarget {
		// Keep the original activation height; repeated calls before the delay
		// elapses must not move the goalpost.
	} else {
		pool.PendingValidatorTarget = msg.ValidatorTarget
		pool.ValidatorChangeHeight = msg.Height + k.genesis.Params.ValidatorChangeDelay
	}
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ApplyPoolReward(poolID string, rewardAmount uint64) (types.NominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if rewardAmount == 0 {
		return pool, nil
	}
	commission := rewardAmount * uint64(pool.PoolCommissionBps) / uint64(types.MaxBasisPoints)
	netReward := rewardAmount - commission
	pool.TotalBondedStake += netReward
	pool.RewardIndex += types.RewardDelta(netReward, pool.TotalShares)
	return k.savePoolOnly(idx, pool)
}

func (k *Keeper) ApplyPoolSlash(poolID string, slashAmount uint64) (types.NominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return types.NominatorPool{}, errors.New("nominator pool not found")
	}
	if slashAmount == 0 {
		return pool, nil
	}
	if slashAmount > pool.TotalBondedStake {
		slashAmount = pool.TotalBondedStake
	}
	pool.TotalBondedStake -= slashAmount
	pool.SlashIndex += types.RewardDelta(slashAmount, pool.TotalShares)
	return k.savePoolOnly(idx, pool)
}

func (k Keeper) NominatorPool(poolID string) (types.NominatorPool, bool) {
	_, pool, found := findPool(k.genesis.State.Pools, poolID)
	return pool, found
}

func (k Keeper) NominatorPools() []types.NominatorPool {
	return types.SortPools(k.genesis.State.Pools)
}

func (k Keeper) PoolDelegator(poolID string, delegator string) (types.DelegatorShare, bool) {
	_, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return types.DelegatorShare{}, false
	}
	_, share, found := findDelegator(pool.DelegatorShares, delegator)
	return share, found
}

func (k Keeper) PoolRewards(poolID string, delegator string) (uint64, bool) {
	_, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return 0, false
	}
	_, share, found := findDelegator(pool.DelegatorShares, delegator)
	if !found {
		return 0, false
	}
	return types.AccruedReward(share, pool.RewardIndex), true
}

func (k Keeper) PoolUnbondingQueue(poolID string) []types.UnbondingEntry {
	_, pool, found := findPool(k.genesis.State.Pools, poolID)
	if !found {
		return []types.UnbondingEntry{}
	}
	return types.SortUnbonding(pool.UnbondingQueue)
}

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator  { return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error { return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k *Keeper) savePool(idx int, pool types.NominatorPool, delegator types.DelegatorShare) (types.DelegatorShare, error) {
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.DelegatorShare{}, err
	}
	k.genesis = next
	return delegator, nil
}

func (k *Keeper) savePoolOnly(idx int, pool types.NominatorPool) (types.NominatorPool, error) {
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.NominatorPool{}, err
	}
	k.genesis = next
	return pool, nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func findPool(pools []types.NominatorPool, poolID string) (int, types.NominatorPool, bool) {
	for idx, pool := range pools {
		if pool.PoolID == poolID {
			return idx, pool, true
		}
	}
	return -1, types.NominatorPool{}, false
}

func findDelegator(delegators []types.DelegatorShare, delegator string) (int, types.DelegatorShare, bool) {
	for idx, share := range delegators {
		if share.Delegator == delegator {
			return idx, share, true
		}
	}
	return -1, types.DelegatorShare{}, false
}

func findWithdrawal(withdrawals []types.PendingWithdrawal, withdrawalID string) (int, types.PendingWithdrawal, bool) {
	for idx, withdrawal := range withdrawals {
		if withdrawal.WithdrawalID == withdrawalID {
			return idx, withdrawal, true
		}
	}
	return -1, types.PendingWithdrawal{}, false
}

func findAllocation(allocations []types.PoolAllocation, validatorAddress string) (int, types.PoolAllocation, bool) {
	for idx, allocation := range allocations {
		if allocation.ValidatorAddress == validatorAddress {
			return idx, allocation, true
		}
	}
	return -1, types.PoolAllocation{}, false
}

func totalAllocated(allocations []types.PoolAllocation) uint64 {
	total := uint64(0)
	for _, allocation := range allocations {
		total += allocation.Amount
	}
	return total
}
