package keeper

import (
	"context"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/stake-concentration/types"
)

type Keeper struct {
	cdc		codec.BinaryCodec
	storeService	corestore.KVStoreService
	authority	string
}

func NewKeeper(cdc codec.BinaryCodec, storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{cdc: cdc, storeService: storeService, authority: authority}
}

func (k Keeper) Authority() string	{ return k.authority }

func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	params = types.NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ParamsKey, bz)
}

func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ParamsKey)
	if err != nil || bz == nil {
		return types.DefaultParams(), err
	}
	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return types.Params{}, err
	}
	params = types.NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return types.Params{}, types.ErrInvalidParams.Wrap(err.Error())
	}
	return params, nil
}

func (k Keeper) RecomputeConcentration(ctx context.Context, epoch uint64, validatorSet []types.ValidatorPower) (types.NetworkConcentration, error) {
	if epoch == 0 {
		return types.NetworkConcentration{}, types.ErrInvalidConcentration.Wrap("epoch must be positive")
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.NetworkConcentration{}, err
	}
	network, err := types.ComputeNetworkConcentration(params, epoch, validatorSet, sdk.UnwrapSDKContext(ctx).BlockHeight())
	if err != nil {
		return types.NetworkConcentration{}, err
	}
	if err := k.SetNetworkConcentration(ctx, network); err != nil {
		return types.NetworkConcentration{}, err
	}
	return network, nil
}

func (k Keeper) SetNetworkConcentration(ctx context.Context, network types.NetworkConcentration) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := network.Validate(params); err != nil {
		return types.ErrInvalidConcentration.Wrap(err.Error())
	}
	bz, err := k.cdc.Marshal(&network)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.NetworkKey, bz)
}

func (k Keeper) GetNetworkConcentration(ctx context.Context) (types.NetworkConcentration, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.NetworkKey)
	if err != nil || bz == nil {
		return types.DefaultNetworkConcentration(), err
	}
	var network types.NetworkConcentration
	if err := k.cdc.Unmarshal(bz, &network); err != nil {
		return types.NetworkConcentration{}, err
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.NetworkConcentration{}, err
	}
	if err := network.Validate(params); err != nil {
		return types.NetworkConcentration{}, types.ErrInvalidConcentration.Wrap(err.Error())
	}
	return network, nil
}

func (k Keeper) GetValidatorConcentration(ctx context.Context, operatorAddress string) (types.ValidatorConcentration, bool, error) {
	network, err := k.GetNetworkConcentration(ctx)
	if err != nil {
		return types.ValidatorConcentration{}, false, err
	}
	for _, validator := range network.Validators {
		if validator.OperatorAddress == operatorAddress {
			return validator, true, nil
		}
	}
	return types.ValidatorConcentration{}, false, nil
}

func (k Keeper) CanAcceptDelegation(ctx context.Context, operatorAddress string) (bool, error) {
	validator, found, err := k.GetValidatorConcentration(ctx, operatorAddress)
	if err != nil || !found {
		return false, err
	}
	return validator.DelegationAllowed, nil
}

func (k Keeper) RewardModifierBps(ctx context.Context, operatorAddress string) (uint32, error) {
	validator, found, err := k.GetValidatorConcentration(ctx, operatorAddress)
	if err != nil || !found {
		return 0, err
	}
	return validator.RewardModifierBps, nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	gs.Params = types.NormalizeParams(gs.Params)
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	return k.SetNetworkConcentration(ctx, gs.Network)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	network, err := k.GetNetworkConcentration(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{Params: params, Network: network}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}
