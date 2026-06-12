package exporting

import (
	"encoding/json"
	"fmt"
	"log"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdklog "cosmossdk.io/log/v2"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

type Dependencies struct {
	AppCodec		codec.Codec
	ModuleManager		*module.Manager
	AccountKeeper		authkeeper.AccountKeeper
	StakingKeeper		*stakingkeeper.Keeper
	DistrKeeper		distrkeeper.Keeper
	SlashingKeeper		slashingkeeper.Keeper
	StakingStoreKey		*storetypes.KVStoreKey
	Logger			sdklog.Logger
	NewContext		func(cmtproto.Header) sdk.Context
	LastBlockHeight		func() int64
	ConsensusParams		func(sdk.Context) cmtproto.ConsensusParams
	EnsureCollections	func(sdk.Context) error
}

func ExportAppStateAndValidators(deps Dependencies, forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	ctx := deps.NewContext(cmtproto.Header{Height: deps.LastBlockHeight()})

	height := deps.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
		PrepForZeroHeightGenesis(deps, ctx, jailAllowedAddrs)
	}
	if err := deps.EnsureCollections(ctx); err != nil {
		return servertypes.ExportedApp{}, err
	}

	genState, err := deps.ModuleManager.ExportGenesisForModules(ctx, deps.AppCodec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, deps.StakingKeeper)
	return servertypes.ExportedApp{
		AppState:		appState,
		Validators:		validators,
		Height:			height,
		ConsensusParams:	deps.ConsensusParams(ctx),
	}, err
}

func PrepForZeroHeightGenesis(deps Dependencies, ctx sdk.Context, jailAllowedAddrs []string) {
	applyAllowedAddrs := len(jailAllowedAddrs) > 0
	allowedAddrsMap := make(map[string]bool)

	for _, addr := range jailAllowedAddrs {
		if _, err := deps.StakingKeeper.ValidatorAddressCodec().StringToBytes(addr); err != nil {
			log.Fatal(err)
		}
		allowedAddrsMap[addr] = true
	}

	err := deps.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
		valBz, err := deps.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			panic(err)
		}
		_, _ = deps.DistrKeeper.WithdrawValidatorCommission(ctx, valBz)
		return false
	})
	if err != nil {
		panic(err)
	}

	dels, err := deps.StakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		panic(err)
	}

	for _, delegation := range dels {
		valBz, err := deps.StakingKeeper.ValidatorAddressCodec().StringToBytes(delegation.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		valAddr := sdk.ValAddress(valBz)

		delBz, err := deps.AccountKeeper.AddressCodec().StringToBytes(delegation.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr := sdk.AccAddress(delBz)

		_, _ = deps.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
	}

	deps.DistrKeeper.DeleteAllValidatorSlashEvents(ctx)
	deps.DistrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	err = deps.StakingKeeper.IterateValidators(ctx, func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
		valBz, err := deps.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			panic(err)
		}
		scraps, err := deps.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, valBz)
		if err != nil {
			panic(err)
		}
		feePool, err := deps.DistrKeeper.FeePool.Get(ctx)
		if err != nil {
			panic(err)
		}
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps...)
		if err := deps.DistrKeeper.FeePool.Set(ctx, feePool); err != nil {
			panic(err)
		}

		if err := deps.DistrKeeper.Hooks().AfterValidatorCreated(ctx, valBz); err != nil {
			panic(err)
		}
		return false
	})
	if err != nil {
		panic(err)
	}

	for _, del := range dels {
		valBz, err := deps.StakingKeeper.ValidatorAddressCodec().StringToBytes(del.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		valAddr := sdk.ValAddress(valBz)
		delBz, err := deps.AccountKeeper.AddressCodec().StringToBytes(del.DelegatorAddress)
		if err != nil {
			panic(err)
		}
		delAddr := sdk.AccAddress(delBz)

		if err := deps.DistrKeeper.Hooks().BeforeDelegationCreated(ctx, delAddr, valAddr); err != nil {
			panic(fmt.Errorf("error while incrementing period: %w", err))
		}
		if err := deps.DistrKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			panic(fmt.Errorf("error while creating a new delegation period record: %w", err))
		}
	}

	ctx = ctx.WithBlockHeight(height)

	err = deps.StakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingtypes.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		if err := deps.StakingKeeper.SetRedelegation(ctx, red); err != nil {
			panic(err)
		}
		return false
	})
	if err != nil {
		panic(fmt.Errorf("error while iterating redelegations: %w", err))
	}

	err = deps.StakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingtypes.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		if err := deps.StakingKeeper.SetUnbondingDelegation(ctx, ubd); err != nil {
			panic(err)
		}
		return false
	})
	if err != nil {
		panic(fmt.Errorf("error while iterating unbonding delegations: %w", err))
	}

	store := ctx.KVStore(deps.StakingStoreKey)
	iter := storetypes.KVStoreReversePrefixIterator(store, stakingtypes.ValidatorsKey)

	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(stakingtypes.AddressFromValidatorsKey(iter.Key()))
		validator, err := deps.StakingKeeper.GetValidator(ctx, addr)
		if err != nil {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0
		if applyAllowedAddrs && !allowedAddrsMap[aetraaddress.FormatValAddress(addr)] {
			validator.Jailed = true
		}

		if err := deps.StakingKeeper.SetValidator(ctx, validator); err != nil {
			panic(fmt.Errorf("unable to set validator: %w", err))
		}
	}

	if err := iter.Close(); err != nil {
		deps.Logger.Error("error while closing the key-value store reverse prefix iterator: ", err)
		return
	}

	if _, err := deps.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx); err != nil {
		log.Fatal(err)
	}

	err = deps.SlashingKeeper.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			if err := deps.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, info); err != nil {
				panic("unable to set validator signing info")
			}
			return false
		},
	)
	if err != nil {
		panic(fmt.Errorf("error while iterating validator signing info: %w", err))
	}
}
