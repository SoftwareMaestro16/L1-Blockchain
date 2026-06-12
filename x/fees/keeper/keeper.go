package keeper

import (
	"context"
	"encoding/json"

	corestore "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/fees/types"
)

// FeeFormulaParamsKey is the KV key for extended formula governance params.
var FeeFormulaParamsKey = []byte{0x10}

type Keeper struct {
	cdc			codec.BinaryCodec
	storeService		corestore.KVStoreService
	accountKeeper		types.AccountKeeper
	bankKeeper		types.BankKeeper
	distributionKeeper	distrkeeper.Keeper
	authority		string
	// reputationReader is optional; nil → neutral reputation for all senders.
	reputationReader	types.ReputationReader
	// feeCollector is the fee-collector module for distributing collected fees.
	feeCollector	types.FeeCollectorKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestore.KVStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distributionKeeper distrkeeper.Keeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:			cdc,
		storeService:		storeService,
		accountKeeper:		accountKeeper,
		bankKeeper:		bankKeeper,
		distributionKeeper:	distributionKeeper,
		authority:		authority,
	}
}

// WithReputationReader returns a Keeper with the optional ReputationReader set.
// This is the AWCE-1 integration boundary — wired in app/keeperwiring, not in NewKeeper.
func (k Keeper) WithReputationReader(r types.ReputationReader) Keeper {
	k.reputationReader = r
	return k
}

// WithFeeCollector returns a Keeper with the FeeCollectorKeeper set.
// The fee-collector module records fees into pending distribution buckets.
func (k Keeper) WithFeeCollector(fc types.FeeCollectorKeeper) Keeper {
	k.feeCollector = fc
	return k
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if err := k.storeService.OpenKVStore(cacheCtx).Set(types.ParamsKey, bz); err != nil {
		return err
	}
	if err := k.syncDistributionParams(cacheCtx, params); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ParamsKey)
	if err != nil {
		return types.Params{}, err
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}
	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return types.Params{}, err
	}
	params = types.NormalizeParams(params)
	return params, nil
}

// SetFeeFormulaParams stores the extended formula governance params in KV.
func (k Keeper) SetFeeFormulaParams(ctx context.Context, p types.FeeFormulaParams) error {
	p = types.NormalizeFeeFormulaParams(p)
	if err := p.Validate(); err != nil {
		return types.ErrInvalidParams.Wrap(err.Error())
	}
	bz, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(FeeFormulaParamsKey, bz)
}

// GetFeeFormulaParams reads the extended formula governance params from KV.
func (k Keeper) GetFeeFormulaParams(ctx context.Context) (types.FeeFormulaParams, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(FeeFormulaParamsKey)
	if err != nil {
		return types.FeeFormulaParams{}, err
	}
	if bz == nil {
		return types.DefaultFeeFormulaParams(), nil
	}
	var p types.FeeFormulaParams
	if err := json.Unmarshal(bz, &p); err != nil {
		return types.FeeFormulaParams{}, err
	}
	return types.NormalizeFeeFormulaParams(p), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	if err := k.SetProtocolFeeState(ctx, gs.ProtocolFeeState); err != nil {
		return err
	}

	if err := k.SetFeeFormulaParams(ctx, types.DefaultFeeFormulaParams()); err != nil {
		return err
	}

	if gs.CongestionBps > uint32(types.BasisPoints) {
		gs.CongestionBps = 0
	}
	return k.SetCongestionState(sdk.UnwrapSDKContext(ctx), gs.CongestionBps)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	state, err := k.GetProtocolFeeState(ctx)
	if err != nil {
		return nil, err
	}
	gs := &types.GenesisState{
		Params:			params,
		ProtocolFeeState:	state,
		CongestionBps:		k.GetCongestionState(sdk.UnwrapSDKContext(ctx)),
	}
	if err := gs.Validate(); err != nil {
		return nil, err
	}
	return gs, nil
}

func (k Keeper) IsAllowedFeeDenom(ctx context.Context, denom string) (bool, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}
	for _, allowed := range params.AllowedFeeDenoms {
		if denom == allowed {
			return true, nil
		}
	}
	return false, nil
}

func (k Keeper) ValidateFees(ctx context.Context, fees sdk.Coins, enforceMin bool) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	return types.ValidateFeeCoins(params, fees, enforceMin)
}

func (k Keeper) SetProtocolFeeState(ctx context.Context, state types.ProtocolFeeState) error {
	if err := state.Validate(); err != nil {
		return err
	}
	bz, err := k.cdc.Marshal(&state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ProtocolFeeStateKey, bz)
}

func (k Keeper) GetProtocolFeeState(ctx context.Context) (types.ProtocolFeeState, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ProtocolFeeStateKey)
	if err != nil {
		return types.ProtocolFeeState{}, err
	}
	if bz == nil {
		return types.DefaultProtocolFeeState(), nil
	}
	var state types.ProtocolFeeState
	if err := k.cdc.Unmarshal(bz, &state); err != nil {
		return types.ProtocolFeeState{}, err
	}
	return state, nil
}

func (k Keeper) RecordCollectedFees(ctx context.Context, fees sdk.Coins) error {
	if fees.Empty() {
		return nil
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	validatorRewards, communityPool, err := types.SplitFees(params, fees)
	if err != nil {
		return err
	}
	state, err := k.GetProtocolFeeState(ctx)
	if err != nil {
		return err
	}
	state.TotalCollected = state.TotalCollected.Add(fees...)
	state.ValidatorRewards = state.ValidatorRewards.Add(validatorRewards...)
	state.CommunityPool = state.CommunityPool.Add(communityPool...)
	if err := k.SetProtocolFeeState(ctx, state); err != nil {
		return err
	}

	if k.feeCollector != nil {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.FeeCollectorModuleName, "feecollector", fees); err != nil {
			return err
		}
		return k.feeCollector.RecordCollectedFees(ctx, fees, "protocol")
	}
	return nil
}

func (k Keeper) GetModuleBalances(ctx context.Context) ([]types.ModuleBalance, error) {
	moduleNames := []string{
		types.FeeCollectorModuleName,
		types.ModuleName,
		types.DistributionModuleName,
		types.ProtocolPoolModuleName,
	}
	balances := make([]types.ModuleBalance, 0, len(moduleNames))
	for _, moduleName := range moduleNames {
		addr := k.accountKeeper.GetModuleAddress(moduleName)
		if addr == nil {
			return nil, types.ErrInvalidParams.Wrapf("module account %s is not configured", moduleName)
		}
		balances = append(balances, types.ModuleBalance{
			ModuleName:	moduleName,
			Address:	aetraaddress.FormatAccAddress(addr),
			Balance:	k.bankKeeper.GetAllBalances(ctx, addr),
		})
	}
	return balances, nil
}

func (k Keeper) syncDistributionParams(ctx context.Context, params types.Params) error {
	communityRatio, err := params.CommunityRatioDec()
	if err != nil {
		return err
	}
	distrParams, err := k.distributionKeeper.Params.Get(ctx)
	if err != nil {
		return err
	}
	distrParams.CommunityTax = communityRatio
	if err := distrParams.ValidateBasic(); err != nil {
		return err
	}
	return k.distributionKeeper.Params.Set(ctx, distrParams)
}

// GetReputationScore returns the on-chain identity reputation score for addr.
// If no ReputationReader is wired (nil), returns neutral score with found=false.
func (k Keeper) GetReputationScore(ctx context.Context, addr sdk.AccAddress) (score uint32, found bool, err error) {
	if k.reputationReader == nil {
		return types.ReputationNeutralScore, false, nil
	}
	return k.reputationReader.GetIdentityReputationScore(ctx, addr)
}
