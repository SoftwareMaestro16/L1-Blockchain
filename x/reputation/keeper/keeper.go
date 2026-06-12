package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	corestore "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/reputation/types"
)

type Keeper struct {
	storeService	corestore.KVStoreService
	authority	string
}

func NewKeeper(storeService corestore.KVStoreService, authority string) Keeper {
	return Keeper{storeService: storeService, authority: authority}
}

func (k Keeper) Authority() string	{ return k.authority }

// StateKeyV2 is the store key for the consolidated reputation state.
var stateKeyV2 = []byte{0x02}

// DefaultGenesis returns a default consolidated genesis state.
func (k Keeper) DefaultGenesis() types.ConsolidatedReputationState {
	params := types.DefaultReputationParams()
	params.Authority = k.authority
	return types.NewConsolidatedReputationState(params)
}

// GetState reads the consolidated reputation state.
// If the old v1 state is found, it auto-migrates to v2 on read.
func (k Keeper) GetState(ctx context.Context) (types.ConsolidatedReputationState, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(stateKeyV2)
	if err != nil {
		return k.DefaultGenesis(), err
	}
	if bz != nil {
		var state types.ConsolidatedReputationState
		if err := json.Unmarshal(bz, &state); err != nil {
			return types.ConsolidatedReputationState{}, err
		}
		state = types.NormalizeConsolidatedState(state)
		if err := state.Validate(); err != nil {
			return types.ConsolidatedReputationState{}, err
		}
		return state, nil
	}

	v1Key := []byte{0x01}
	v1bz, err := store.Get(v1Key)
	if err != nil {
		return k.DefaultGenesis(), err
	}
	if v1bz != nil {
		var oldState types.ReputationState
		if err := json.Unmarshal(v1bz, &oldState); err != nil {
			return types.ConsolidatedReputationState{}, fmt.Errorf("failed to unmarshal old state: %w", err)
		}
		oldState = types.NormalizeReputationState(oldState)
		consolidated := types.MigrateFromReputationState(oldState)
		if err := k.SetState(ctx, consolidated); err != nil {
			return types.ConsolidatedReputationState{}, fmt.Errorf("failed to persist migrated state: %w", err)
		}

		_ = store.Delete(v1Key)
		return consolidated, nil
	}

	return k.DefaultGenesis(), nil
}

// SetState writes the consolidated reputation state.
func (k Keeper) SetState(ctx context.Context, state types.ConsolidatedReputationState) error {
	state = types.NormalizeConsolidatedState(state)
	if err := state.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(stateKeyV2, bz)
}

func (k Keeper) InitGenesis(ctx context.Context, state types.ConsolidatedReputationState) error {
	if state.Params.Authority == "" || state.Params.Authority == types.DefaultReputationAuthority {
		state.Params.Authority = k.authority
	}
	return k.SetState(ctx, state)
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.ConsolidatedReputationState, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	exported, err := types.ExportConsolidatedState(state)
	if err != nil {
		return nil, err
	}
	return &exported, nil
}

// GetIdentityReputation returns the identity reputation for the given account.
func (k Keeper) GetIdentityReputation(ctx context.Context, account string) (*types.IdentityReputation, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	id, found := types.FindIdentity(state, account)
	if !found {
		return types.NewIdentityReputation(account), nil
	}
	return &id, nil
}

// SetIdentityReputation updates the identity reputation for an account.
func (k Keeper) SetIdentityReputation(ctx context.Context, identity types.IdentityReputation) error {
	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}
	state = types.UpsertIdentity(state, identity)
	return k.SetState(ctx, state)
}

// GetValidatorReputation returns the validator score for the given validator address.
func (k Keeper) GetValidatorReputation(ctx context.Context, addr string) (*types.ValidatorScore, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	vs, found := types.FindValidatorScore(state, addr)
	if !found {
		return types.NewValidatorScore(addr), nil
	}
	return &vs, nil
}

// SetValidatorReputation updates the validator score.
func (k Keeper) SetValidatorReputation(ctx context.Context, vs types.ValidatorScore) error {
	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}
	state = types.UpsertValidatorScore(state, vs)
	return k.SetState(ctx, state)
}

// GetServiceTrustScore returns the service trust score.
func (k Keeper) GetServiceTrustScore(ctx context.Context, addr string) (*types.ServiceTrustScore, error) {
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}
	sts, found := types.FindServiceTrustScore(state, addr)
	if !found {
		return types.NewServiceTrustScore(addr), nil
	}
	return &sts, nil
}

// SetServiceTrustScore updates the service trust score.
func (k Keeper) SetServiceTrustScore(ctx context.Context, sts types.ServiceTrustScore) error {
	state, err := k.GetState(ctx)
	if err != nil {
		return err
	}
	state = types.UpsertServiceTrustScore(state, sts)
	return k.SetState(ctx, state)
}

// StakeReputation is disabled in the public API; use IdentityReputation instead.
func (k Keeper) StakeReputation(ctx context.Context, req types.QueryStakeReputationRequest) (types.QueryStakeReputationResponse, error) {
	return types.QueryStakeReputationResponse{}, fmt.Errorf("stake reputation query disabled; use QueryIdentityReputation or GetIdentityReputation")
}

// AccountReputation is disabled in the public API; use IdentityReputation instead.
func (k Keeper) AccountReputation(ctx context.Context, req types.QueryAccountReputationRequest) (types.QueryAccountReputationResponse, error) {
	return types.QueryAccountReputationResponse{}, fmt.Errorf("account reputation query disabled; use QueryIdentityReputation or GetIdentityReputation")
}

// GetIdentityReputationScore returns the uint32 score for fee module integration.
func (k Keeper) GetIdentityReputationScore(ctx context.Context, addr sdk.AccAddress) (uint32, bool, error) {

	if len(addr) == 0 {
		return types.IdentityScoreDefault, false, nil
	}
	if !(len(addr) == 20 || len(addr) == 32) {

		return types.IdentityScoreDefault, false, nil
	}
	state, err := k.GetState(ctx)
	if err != nil {
		return types.IdentityScoreDefault, false, nil
	}
	userAddr := aetraaddress.FormatAccAddress(addr)
	id, found := types.FindIdentity(state, userAddr)
	if !found {
		return types.IdentityScoreDefault, false, nil
	}
	return id.Score, true, nil
}

func parseAddress(text string) (sdk.AccAddress, error) {
	return aetraaddress.ParseAccAddress(text)
}

func mustJSON(v any) (string, error) {
	bz, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}
