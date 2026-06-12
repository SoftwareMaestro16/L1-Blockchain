package keeper

import (
	"context"
	"encoding/json"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
)

var genesisKey = []byte{0x01}

type Keeper struct {
	state		types.GenesisState
	storeService	corestore.KVStoreService
	runtimeCtx	context.Context
}

func NewKeeper(authority string) Keeper {
	return Keeper{state: types.DefaultGenesisState(authority)}
}

func NewPersistentKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{state: types.DefaultGenesisState(authority), storeService: storeService}
}

func (k Keeper) Authority() string {
	return k.state.Params.Authority
}

func (k Keeper) Params() types.Params {
	return k.state.Params
}

func (k *Keeper) SetParams(params types.Params) error {
	if err := params.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	next := k.state
	next.Params = params
	if err := next.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	k.state = next
	return k.save()
}

func (k *Keeper) RecomputePolicy(epoch uint64, validators []types.ValidatorStake) (types.NetworkPolicy, error) {
	network, err := types.ComputeNetworkPolicy(k.state.Params, epoch, validators, k.state.WarningAcknowledgements)
	if err != nil {
		return types.NetworkPolicy{}, err
	}
	k.state.Network = network
	return network, k.save()
}

func (k *Keeper) RegisterValidatorIdentity(identity types.ValidatorIdentityMetadata) error {
	if err := identity.Validate(); err != nil {
		return types.ErrInvalidPolicy.Wrap(err.Error())
	}
	next := k.state
	next.Identities = upsertIdentity(next.Identities, identity)
	if err := next.Validate(); err != nil {
		return types.ErrInvalidPolicy.Wrap(err.Error())
	}
	k.state = next
	return k.save()
}

func (k *Keeper) AcknowledgeConcentrationWarning(ack types.WarningAcknowledgement) error {
	if err := ack.Validate(); err != nil {
		return types.ErrInvalidPolicy.Wrap(err.Error())
	}
	next := k.state
	next.WarningAcknowledgements = upsertAcknowledgement(next.WarningAcknowledgements, ack)
	for i := range next.Network.Validators {
		if next.Network.Validators[i].OperatorAddress == ack.OperatorAddress && next.Network.Validators[i].DelegationWarning == ack.Warning {
			next.Network.Validators[i].WarningAcknowledged = true
		}
	}
	if err := next.Validate(); err != nil {
		return types.ErrInvalidPolicy.Wrap(err.Error())
	}
	k.state = next
	return k.save()
}

func (k Keeper) QueryParams(req types.QueryParamsRequest) (types.QueryParamsResponse, error) {
	return types.QueryParamsResponse{Params: k.state.Params}, nil
}

func (k Keeper) QueryValidatorEffectivePower(req types.QueryValidatorEffectivePowerRequest) (types.QueryValidatorEffectivePowerResponse, error) {
	validator, found := k.findValidator(req.OperatorAddress)
	if !found {
		return types.QueryValidatorEffectivePowerResponse{}, types.ErrNotFound
	}
	return types.QueryValidatorEffectivePowerResponse{
		OperatorAddress:	validator.OperatorAddress,
		EffectiveStake:		validator.EffectiveStake,
		EffectivePowerBps:	validator.EffectivePowerBps,
		PowerCapBps:		validator.PowerCapBps,
	}, nil
}

func (k Keeper) QueryValidatorStake(req types.QueryValidatorStakeRequest) (types.QueryValidatorStakeResponse, error) {
	validator, found := k.findValidator(req.OperatorAddress)
	if !found {
		return types.QueryValidatorStakeResponse{}, types.ErrNotFound
	}
	return types.QueryValidatorStakeResponse{
		OperatorAddress:	validator.OperatorAddress,
		RawStake:		validator.RawStake,
		EffectiveStake:		validator.EffectiveStake,
		OverflowStake:		validator.OverflowStake,
		RawPowerBps:		validator.RawPowerBps,
		EffectivePowerBps:	validator.EffectivePowerBps,
	}, nil
}

func (k Keeper) QueryTopNConcentration(req types.QueryTopNConcentrationRequest) (types.QueryTopNConcentrationResponse, error) {
	if req.N == 0 {
		return types.QueryTopNConcentrationResponse{}, types.ErrInvalidPolicy.Wrap("top-n must be positive")
	}
	values := make([]uint32, 0, len(k.state.Network.Validators))
	for _, validator := range k.state.Network.Validators {
		values = append(values, validator.RawPowerBps)
	}
	power := types.TopNBps(values, req.N)
	target := types.TopNTargetBps(req.N)
	return types.QueryTopNConcentrationResponse{N: req.N, PowerBps: power, TargetBps: target, Exceeded: power >= target}, nil
}

func (k Keeper) QueryValidatorRewardMultiplier(req types.QueryValidatorRewardMultiplierRequest) (types.QueryValidatorRewardMultiplierResponse, error) {
	validator, found := k.findValidator(req.OperatorAddress)
	if !found {
		return types.QueryValidatorRewardMultiplierResponse{}, types.ErrNotFound
	}
	return types.QueryValidatorRewardMultiplierResponse{OperatorAddress: validator.OperatorAddress, RewardMultiplierBps: validator.RewardMultiplierBps}, nil
}

func (k Keeper) QueryDelegationWarningStatus(req types.QueryDelegationWarningStatusRequest) (types.QueryDelegationWarningStatusResponse, error) {
	validator, found := k.findValidator(req.OperatorAddress)
	if !found {
		return types.QueryDelegationWarningStatusResponse{}, types.ErrNotFound
	}
	return types.QueryDelegationWarningStatusResponse{
		OperatorAddress:	validator.OperatorAddress,
		Warning:		validator.DelegationWarning,
		WarningAcknowledged:	validator.WarningAcknowledged,
	}, nil
}

func (k *Keeper) InitGenesis(gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.state = gs
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.state = gs
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return k.saveWithCtx(ctx)
}

func (k Keeper) ExportGenesis() (types.GenesisState, error) {
	if err := k.state.Validate(); err != nil {
		return types.GenesisState{}, err
	}
	return k.state, nil
}

func (k Keeper) ExportGenesisState(ctx context.Context) (types.GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis()
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return types.GenesisState{}, err
	}
	if len(bz) == 0 {
		return k.ExportGenesis()
	}
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return types.GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return types.GenesisState{}, err
	}
	return gs, nil
}

func (k Keeper) MarshalGenesis() ([]byte, error) {
	gs, err := k.ExportGenesis()
	if err != nil {
		return nil, err
	}
	return json.Marshal(gs)
}

func (k *Keeper) UnmarshalGenesis(bz []byte) error {
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return err
	}
	return k.InitGenesis(gs)
}

func (k Keeper) findValidator(operator string) (types.ValidatorPolicy, bool) {
	for _, validator := range k.state.Network.Validators {
		if validator.OperatorAddress == operator {
			return validator, true
		}
	}
	return types.ValidatorPolicy{}, false
}

func (k *Keeper) save() error {
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return k.saveWithCtx(k.runtimeCtx)
}

func (k *Keeper) saveWithCtx(ctx context.Context) error {
	bz, err := json.Marshal(k.state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func upsertIdentity(values []types.ValidatorIdentityMetadata, identity types.ValidatorIdentityMetadata) []types.ValidatorIdentityMetadata {
	next := append([]types.ValidatorIdentityMetadata(nil), values...)
	for i, current := range next {
		if current.OperatorAddress == identity.OperatorAddress {
			next[i] = identity
			return types.SortIdentities(next)
		}
	}
	return types.SortIdentities(append(next, identity))
}

func upsertAcknowledgement(values []types.WarningAcknowledgement, ack types.WarningAcknowledgement) []types.WarningAcknowledgement {
	next := append([]types.WarningAcknowledgement(nil), values...)
	for i, current := range next {
		if current.OperatorAddress == ack.OperatorAddress && current.Warning == ack.Warning {
			next[i] = ack
			return types.SortAcknowledgements(next)
		}
	}
	return types.SortAcknowledgements(append(next, ack))
}
