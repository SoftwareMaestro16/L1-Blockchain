package keeper

import (
	"context"
	"errors"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/validator-registry/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.State
}

type Keeper struct {
	genesis			GenesisState
	storeService		corestore.KVStoreService
	runtimeCtx		context.Context
	reputationKeeper	types.ReputationKeeper
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func (k Keeper) WithReputationKeeper(rk types.ReputationKeeper) Keeper {
	k.reputationKeeper = rk
	return k
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		params,
		State:		types.State{Validators: []types.ValidatorRecord{}},
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("validator registry unsupported genesis version")
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
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, k.genesis)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	gs, _, err := prefixgenesis.Load(ctx, k.storeService, genesisKey, DefaultGenesis())
	if err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) saveGenesis(next GenesisState) error {
	next = cloneGenesis(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return prefixgenesis.Save(k.runtimeCtx, k.storeService, genesisKey, next)
}

func (k *Keeper) RegisterValidator(msg types.MsgRegisterValidator) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	if msg.Height == 0 {
		return types.ValidatorRecord{}, errors.New("validator registry register height must be positive")
	}
	validator := msg.Validator.Normalize(k.genesis.Params)
	if _, found := k.genesis.State.Validator(validator.OperatorAddress); found {
		return types.ValidatorRecord{}, errors.New("validator registry validator already registered")
	}
	if validator.Status == types.StatusTombstoned {
		return types.ValidatorRecord{}, errors.New("validator registry cannot register tombstoned validator")
	}
	validator = types.AddHistory(validator, msg.Height, types.HistoryRegistered, "validator registered", k.genesis.Params)
	if err := validator.Validate(k.genesis.Params); err != nil {
		return types.ValidatorRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Validators = types.UpsertValidator(next.State.Validators, validator)
	if err := next.Validate(); err != nil {
		return types.ValidatorRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ValidatorRecord{}, err
	}
	return validator, nil
}

func (k *Keeper) UpdateValidatorMetadata(msg types.MsgUpdateValidatorMetadata) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.transition(msg.OperatorAddress, msg.Height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		if uint32(len(msg.Metadata)) > k.genesis.Params.MaxMetadataBytes {
			return types.ValidatorRecord{}, errors.New("validator registry metadata limit exceeded")
		}
		v.Metadata = msg.Metadata
		return types.AddHistory(v, msg.Height, types.HistoryMetadataUpdated, "metadata updated", k.genesis.Params), nil
	})
}

func (k *Keeper) RotateConsensusKey(msg types.MsgRotateConsensusKey) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	if msg.Height == 0 {
		return types.ValidatorRecord{}, errors.New("validator registry rotation height must be positive")
	}
	if msg.ActivationHeight < msg.Height+k.genesis.Params.ConsensusKeyRotationDelay {
		return types.ValidatorRecord{}, errors.New("validator registry consensus key rotation delay has not elapsed")
	}
	return k.transition(msg.OperatorAddress, msg.Height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		if v.Status == types.StatusTombstoned || v.Status == types.StatusRetired {
			return types.ValidatorRecord{}, errors.New("validator registry inactive validator cannot rotate consensus key")
		}
		v.PendingConsensusPublicKey = msg.NewConsensusPublicKey
		v.ConsensusKeyActivationHeight = msg.ActivationHeight
		return types.AddHistory(v, msg.Height, types.HistoryConsensusRotated, "consensus key rotation scheduled", k.genesis.Params), nil
	})
}

func (k *Keeper) ApplyConsensusKeyRotation(operator string, height uint64) (types.ValidatorRecord, bool, error) {
	validator, found := k.genesis.State.Validator(operator)
	if !found {
		return types.ValidatorRecord{}, false, errors.New("validator registry validator not found")
	}
	if validator.PendingConsensusPublicKey == "" {
		return validator, false, nil
	}
	if height < validator.ConsensusKeyActivationHeight {
		return validator, false, errors.New("validator registry consensus key rotation delay has not elapsed")
	}
	validator.ConsensusPublicKey = validator.PendingConsensusPublicKey
	validator.PendingConsensusPublicKey = ""
	validator.ConsensusKeyActivationHeight = 0
	validator = types.AddHistory(validator, height, types.HistoryConsensusRotated, "consensus key rotation applied", k.genesis.Params)
	next := cloneGenesis(k.genesis)
	next.State.Validators = types.UpsertValidator(next.State.Validators, validator.Normalize(k.genesis.Params))
	if err := next.Validate(); err != nil {
		return types.ValidatorRecord{}, false, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ValidatorRecord{}, false, err
	}
	return validator, true, nil
}

func (k *Keeper) UpdateWithdrawalAddress(msg types.MsgUpdateWithdrawalAddress) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.transition(msg.OperatorAddress, msg.Height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		v.WithdrawalAddress = msg.WithdrawalAddress
		return types.AddHistory(v, msg.Height, types.HistoryWithdrawalUpdated, "withdrawal address updated", k.genesis.Params), nil
	})
}

func (k *Keeper) UpdateTreasuryAddress(msg types.MsgUpdateTreasuryAddress) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.transition(msg.OperatorAddress, msg.Height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		v.TreasuryAddress = msg.TreasuryAddress
		return types.AddHistory(v, msg.Height, types.HistoryTreasuryUpdated, "treasury address updated", k.genesis.Params), nil
	})
}

func (k *Keeper) RetireValidator(msg types.MsgRetireValidator) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.SetValidatorStatus(msg.Authority, msg.OperatorAddress, types.StatusRetired, msg.Height)
}

func (k *Keeper) SetValidatorCapabilities(msg types.MsgSetValidatorCapabilities) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.transition(msg.OperatorAddress, msg.Height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		v.Capabilities = msg.Capabilities
		return types.AddHistory(v, msg.Height, types.HistoryCapabilitiesSet, "validator capabilities updated", k.genesis.Params), nil
	})
}

func (k *Keeper) UpdateValidatorCommission(msg types.MsgUpdateValidatorCommission) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.transition(msg.OperatorAddress, msg.Height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		if err := k.genesis.Params.ValidateCommissionChange(v.CommissionPolicy.CurrentRateBps, msg.NewRateBps); err != nil {
			return types.ValidatorRecord{}, err
		}
		v.CommissionPolicy.CurrentRateBps = msg.NewRateBps
		return types.AddHistory(v, msg.Height, "commission-updated", "validator commission updated", k.genesis.Params), nil
	})
}

func (k *Keeper) SetValidatorStatus(authority, operator, status string, height uint64) (types.ValidatorRecord, error) {
	if err := k.genesis.Params.Authorize(authority); err != nil {
		return types.ValidatorRecord{}, err
	}
	return k.transition(operator, height, func(v types.ValidatorRecord) (types.ValidatorRecord, error) {
		if !types.ValidStatusTransition(v.Status, status) {
			return types.ValidatorRecord{}, errors.New("validator registry invalid status transition")
		}
		v.Status = status
		return types.AddHistory(v, height, types.HistoryStatusChanged, status, k.genesis.Params), nil
	})
}

func (k Keeper) Validator(operator string) (types.ValidatorRecord, bool, error) {
	if err := k.genesis.Validate(); err != nil {
		return types.ValidatorRecord{}, false, err
	}
	validator, found := k.genesis.State.Validator(operator)
	return validator, found, nil
}

func (k Keeper) Validators() ([]types.ValidatorRecord, error) {
	if err := k.genesis.Validate(); err != nil {
		return nil, err
	}
	return types.SortValidators(k.genesis.State.Validators), nil
}

func (k Keeper) ValidatorKeys(operator string) (types.ValidatorKeys, bool, error) {
	validator, found, err := k.Validator(operator)
	if err != nil || !found {
		return types.ValidatorKeys{}, found, err
	}
	return validator.Keys(), true, nil
}

func (k Keeper) ValidatorPerformance(operator string) (types.ValidatorPerformance, bool, error) {
	validator, found, err := k.Validator(operator)
	if err != nil || !found {
		return types.ValidatorPerformance{}, found, err
	}
	perf := validator.Performance()
	if score := k.effectiveReputationScore(operator); score > 0 {
		perf.ReputationScore = score
	}
	return perf, true, nil
}

func (k Keeper) effectiveReputationScore(operator string) uint32 {
	if k.reputationKeeper == nil || k.runtimeCtx == nil {
		return 0
	}
	score, _, err := k.reputationKeeper.GetValidatorTotalScore(k.runtimeCtx, operator)
	if err != nil {
		return 0
	}
	return score
}

func (k Keeper) ValidatorSecurityStatus(operator string) (types.ValidatorSecurityStatus, bool, error) {
	validator, found, err := k.Validator(operator)
	if err != nil || !found {
		return types.ValidatorSecurityStatus{}, found, err
	}
	return validator.SecurityStatus(), true, nil
}

func (k Keeper) ValidatorHistory(operator string) ([]types.ValidatorHistoryEvent, bool, error) {
	validator, found, err := k.Validator(operator)
	if err != nil || !found {
		return nil, found, err
	}
	return append([]types.ValidatorHistoryEvent(nil), validator.History...), true, nil
}

func (k Keeper) ValidatorAllocationInputs(req types.ValidatorAllocationQueryRequest) ([]types.ValidatorAllocationEngineInput, error) {
	if err := k.genesis.Validate(); err != nil {
		return nil, err
	}
	return k.genesis.State.ValidatorAllocationEngineInputs(k.genesis.Params, req)
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k *Keeper) transition(operator string, height uint64, mutate func(types.ValidatorRecord) (types.ValidatorRecord, error)) (types.ValidatorRecord, error) {
	if height == 0 {
		return types.ValidatorRecord{}, errors.New("validator registry event height must be positive")
	}
	validator, found := k.genesis.State.Validator(operator)
	if !found {
		return types.ValidatorRecord{}, errors.New("validator registry validator not found")
	}
	updated, err := mutate(validator)
	if err != nil {
		return types.ValidatorRecord{}, err
	}
	updated = updated.Normalize(k.genesis.Params)
	if err := updated.Validate(k.genesis.Params); err != nil {
		return types.ValidatorRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Validators = types.UpsertValidator(next.State.Validators, updated)
	if err := next.Validate(); err != nil {
		return types.ValidatorRecord{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ValidatorRecord{}, err
	}
	return updated, nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}
