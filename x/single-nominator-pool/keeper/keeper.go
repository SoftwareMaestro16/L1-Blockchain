package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/single-nominator-pool/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.State
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
}

func NewKeeper() Keeper	{ return Keeper{genesis: DefaultGenesis()} }

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{Version: prototype.CurrentGenesisVersion, Params: params, State: types.State{}.Normalize(params)}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("single nominator pool unsupported genesis version")
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

func (k Keeper) ExportGenesis() GenesisState	{ return cloneGenesis(k.genesis) }

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

func (k *Keeper) CreateSingleNominatorPool(msg types.MsgCreateSingleNominatorPool) (types.SingleNominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SingleNominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.SingleNominatorPool{}, errors.New("single nominator pool creation height must be positive")
	}
	if types.IsJailedValidatorStatus(msg.ValidatorStatus) {
		return types.SingleNominatorPool{}, errors.New("single nominator pool cannot delegate to jailed validator")
	}
	if _, _, found := findPool(k.genesis.State.Pools, msg.PoolAddress); found {
		return types.SingleNominatorPool{}, errors.New("single nominator pool already exists")
	}
	pool := types.SingleNominatorPool{
		PoolAddress:	msg.PoolAddress,
		Owner:		msg.Owner,
		Validator:	msg.Validator,
		Status:		types.StatusActive,
	}
	if err := pool.Validate(k.genesis.Params); err != nil {
		return types.SingleNominatorPool{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Pools = append(next.State.Pools, pool)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.SingleNominatorPool{}, err
	}
	k.genesis = next
	return pool, nil
}

func (k *Keeper) DepositSingleNominator(msg types.MsgDepositSingleNominator) (types.SingleNominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SingleNominatorPool{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.SingleNominatorPool{}, errors.New("single nominator deposit amount and height must be positive")
	}
	idx, pool, err := k.ownerPool(msg.PoolAddress, msg.Owner)
	if err != nil {
		return types.SingleNominatorPool{}, err
	}
	if pool.Status != types.StatusActive {
		return types.SingleNominatorPool{}, errors.New("single nominator pool must be active")
	}
	if math.MaxUint64-pool.BondedStake < msg.Amount {
		return types.SingleNominatorPool{}, errors.New("single nominator deposit would overflow bonded stake")
	}
	pool.BondedStake += msg.Amount
	return k.savePool(idx, pool)
}

func (k *Keeper) WithdrawSingleNominator(msg types.MsgWithdrawSingleNominator) (types.PendingWithdrawal, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingWithdrawal{}, err
	}
	if msg.Height == 0 {
		return types.PendingWithdrawal{}, errors.New("single nominator withdrawal height must be positive")
	}
	idx, pool, err := k.ownerPool(msg.PoolAddress, msg.Owner)
	if err != nil {
		return types.PendingWithdrawal{}, err
	}
	if pool.EmergencyLock {
		return types.PendingWithdrawal{}, errors.New("single nominator emergency lock blocks withdrawals")
	}
	if pool.PendingWithdrawal.Status == types.WithdrawalStatusPending {
		if msg.Height < pool.PendingWithdrawal.CompleteHeight {
			return types.PendingWithdrawal{}, errors.New("single nominator withdrawal unbonding has not completed")
		}
		pool.PendingWithdrawal.Status = types.WithdrawalStatusCompleted
		completed := pool.PendingWithdrawal
		pool.PendingWithdrawal = types.PendingWithdrawal{}
		if _, err := k.savePool(idx, pool); err != nil {
			return types.PendingWithdrawal{}, err
		}
		return completed, nil
	}
	if msg.Amount == 0 || msg.Amount > pool.BondedStake {
		return types.PendingWithdrawal{}, errors.New("single nominator withdrawal amount exceeds bonded stake")
	}
	pool.BondedStake -= msg.Amount
	pool.PendingWithdrawal = types.PendingWithdrawal{
		Amount:		msg.Amount,
		RequestHeight:	msg.Height,
		CompleteHeight:	msg.Height + k.genesis.Params.UnbondingBlocks,
		Status:		types.WithdrawalStatusPending,
	}
	if _, err := k.savePool(idx, pool); err != nil {
		return types.PendingWithdrawal{}, err
	}
	return pool.PendingWithdrawal, nil
}

func (k *Keeper) ClaimSingleNominatorRewards(msg types.MsgClaimSingleNominatorRewards) (uint64, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return 0, err
	}
	if msg.Height == 0 {
		return 0, errors.New("single nominator reward claim height must be positive")
	}
	idx, pool, err := k.ownerPool(msg.PoolAddress, msg.Owner)
	if err != nil {
		return 0, err
	}
	reward := pool.RewardBalance
	pool.RewardBalance = 0
	if _, err := k.savePool(idx, pool); err != nil {
		return 0, err
	}
	return reward, nil
}

func (k *Keeper) EmergencyLockSingleNominator(msg types.MsgEmergencyLockSingleNominator) (types.SingleNominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SingleNominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.SingleNominatorPool{}, errors.New("single nominator emergency lock height must be positive")
	}
	idx, pool, err := k.ownerPool(msg.PoolAddress, msg.Owner)
	if err != nil {
		return types.SingleNominatorPool{}, err
	}
	pool.EmergencyLock = msg.Locked
	return k.savePool(idx, pool)
}

func (k *Keeper) ChangeSingleNominatorValidator(msg types.MsgChangeSingleNominatorValidator) (types.SingleNominatorPool, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.SingleNominatorPool{}, err
	}
	if msg.Height == 0 {
		return types.SingleNominatorPool{}, errors.New("single nominator validator change height must be positive")
	}
	if types.IsJailedValidatorStatus(msg.ValidatorStatus) {
		return types.SingleNominatorPool{}, errors.New("single nominator pool cannot delegate to jailed validator")
	}
	idx, pool, err := k.ownerPool(msg.PoolAddress, msg.Owner)
	if err != nil {
		return types.SingleNominatorPool{}, err
	}
	pool.Validator = msg.Validator
	return k.savePool(idx, pool)
}

func (k *Keeper) ApplySingleNominatorReward(poolAddress string, amount uint64) (types.SingleNominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolAddress)
	if !found {
		return types.SingleNominatorPool{}, errors.New("single nominator pool not found")
	}
	if math.MaxUint64-pool.RewardBalance < amount {
		return types.SingleNominatorPool{}, errors.New("single nominator reward would overflow balance")
	}
	pool.RewardBalance += amount
	return k.savePool(idx, pool)
}

func (k *Keeper) ApplySingleNominatorSlash(poolAddress string, amount uint64) (types.SingleNominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolAddress)
	if !found {
		return types.SingleNominatorPool{}, errors.New("single nominator pool not found")
	}
	if amount > pool.BondedStake {
		amount = pool.BondedStake
	}
	pool.BondedStake -= amount
	return k.savePool(idx, pool)
}

func (k Keeper) SingleNominatorPool(poolAddress string) (types.SingleNominatorPool, bool) {
	_, pool, found := findPool(k.genesis.State.Pools, poolAddress)
	return pool, found
}

func (k Keeper) SingleNominatorPools() []types.SingleNominatorPool {
	return types.SortPools(k.genesis.State.Pools)
}

func (k Keeper) SingleNominatorRewards(poolAddress string) (uint64, bool) {
	_, pool, found := findPool(k.genesis.State.Pools, poolAddress)
	if !found {
		return 0, false
	}
	return pool.RewardBalance, true
}

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator	{ return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error	{ return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k Keeper) ownerPool(poolAddress, owner string) (int, types.SingleNominatorPool, error) {
	idx, pool, found := findPool(k.genesis.State.Pools, poolAddress)
	if !found {
		return -1, types.SingleNominatorPool{}, errors.New("single nominator pool not found")
	}
	if pool.Owner != owner {
		return -1, types.SingleNominatorPool{}, errors.New("only single nominator owner can manage stake")
	}
	return idx, pool, nil
}

func (k *Keeper) savePool(idx int, pool types.SingleNominatorPool) (types.SingleNominatorPool, error) {
	next := cloneGenesis(k.genesis)
	next.State.Pools[idx] = pool
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.SingleNominatorPool{}, err
	}
	k.genesis = next
	return pool, nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func findPool(pools []types.SingleNominatorPool, poolAddress string) (int, types.SingleNominatorPool, bool) {
	for idx, pool := range pools {
		if pool.PoolAddress == poolAddress {
			return idx, pool, true
		}
	}
	return -1, types.SingleNominatorPool{}, false
}
