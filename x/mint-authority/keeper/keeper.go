package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/mint-authority/types"
)

type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type Keeper struct {
	storeService	corestore.KVStoreService
	bankKeeper	BankKeeper
	authority	string
}

func NewKeeper(storeService corestore.KVStoreService, bankKeeper BankKeeper, authority string) Keeper {
	return Keeper{storeService: storeService, bankKeeper: bankKeeper, authority: authority}
}

func (k Keeper) Authority() string	{ return k.authority }

func (k Keeper) DefaultGenesis() types.MintAuthorityState {
	params := types.DefaultMintAuthorityParams()
	params.Authority = k.authority
	state, err := types.NewMintAuthorityState(params, types.DefaultMintAuthorityRegistration(), nil)
	if err != nil {
		panic(err)
	}
	return state
}

func (k Keeper) GetState(ctx context.Context) (types.MintAuthorityState, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.StateKey)
	if err != nil || bz == nil {
		return k.DefaultGenesis(), err
	}
	var state types.MintAuthorityState
	if err := json.Unmarshal(bz, &state); err != nil {
		return types.MintAuthorityState{}, err
	}
	state = types.NormalizeMintAuthorityState(state)
	if err := state.Validate(); err != nil {
		return types.MintAuthorityState{}, err
	}
	return state, nil
}

func (k Keeper) SetState(ctx context.Context, state types.MintAuthorityState) error {
	state = types.NormalizeMintAuthorityState(state)
	if err := state.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.StateKey, bz)
}

func (k Keeper) MintProtocolCoins(ctx context.Context, msg types.MsgMintProtocolCoins, decision types.EmissionDecision, auth types.ConstitutionEmergencyAuthorization) (types.MintEvent, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return types.MintEvent{}, err
	}
	next, event, err := types.ApplyMintProtocolCoins(state, msg, decision, auth)
	if err != nil {
		return types.MintEvent{}, err
	}
	if k.bankKeeper != nil {
		recipient, err := aetraaddress.ParseAccAddress(msg.Recipient)
		if err != nil {
			return types.MintEvent{}, err
		}
		coin := sdk.NewCoin(msg.Denom, msg.Amount)
		if err := k.bankKeeper.MintCoins(ctx, types.DefaultMintAuthorityModuleAccount, sdk.NewCoins(coin)); err != nil {
			return types.MintEvent{}, err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.DefaultMintAuthorityModuleAccount, recipient, sdk.NewCoins(coin)); err != nil {
			return types.MintEvent{}, err
		}
	}
	if err := k.SetState(ctx, next); err != nil {
		return types.MintEvent{}, err
	}
	return event, nil
}

func (k Keeper) InitGenesis(ctx context.Context, state types.MintAuthorityState) error {
	if state.Params.Authority == "" || state.Params.Authority == types.DefaultMintAuthorityParamsAuthority {
		state.Params.Authority = k.authority
	}
	return k.SetState(ctx, state)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.MintAuthorityState, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	exported, err := types.ExportMintAuthorityState(state)
	if err != nil {
		return nil, err
	}
	return &exported, nil
}

func parseInt(value string) (sdkmath.Int, error) {
	if value == "" {
		return sdkmath.ZeroInt(), nil
	}
	out, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid integer %q", value)
	}
	return out, nil
}

func mustJSON(v any) (string, error) {
	bz, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
