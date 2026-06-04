package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{NextPoolId: DefaultNextPoolID, Pools: []Pool{}, Params: DefaultParams()}
}

func DefaultParams() Params {
	return Params{
		SwapFeeBps:          DefaultSwapFeeBps,
		MaxSwapFeeBps:       DefaultMaxSwapFeeBps,
		PoolCreationEnabled: true,
		SwapsEnabled:        true,
		LiquidityEnabled:    true,
		MinInitialLiquidity: "1",
	}
}

func NormalizeParams(params Params) Params {
	if isZeroParams(params) {
		return DefaultParams()
	}
	if params.MaxSwapFeeBps == 0 {
		params.MaxSwapFeeBps = DefaultMaxSwapFeeBps
	}
	if params.MinInitialLiquidity == "" {
		params.MinInitialLiquidity = "1"
	}
	return params
}

func isZeroParams(params Params) bool {
	return params.SwapFeeBps == 0 &&
		params.MaxSwapFeeBps == 0 &&
		!params.PoolCreationEnabled &&
		!params.SwapsEnabled &&
		!params.LiquidityEnabled &&
		params.MinInitialLiquidity == ""
}

func (p Params) Validate() error {
	if p.MaxSwapFeeBps == 0 || p.MaxSwapFeeBps > DefaultMaxSwapFeeBps {
		return fmt.Errorf("max_swap_fee_bps must be between 1 and %d", DefaultMaxSwapFeeBps)
	}
	if p.SwapFeeBps > p.MaxSwapFeeBps {
		return fmt.Errorf("swap_fee_bps must be <= max_swap_fee_bps")
	}
	if err := validatePositiveInt("min_initial_liquidity", p.MinInitialLiquidity); err != nil {
		return err
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	if gs.NextPoolId == 0 {
		return fmt.Errorf("next_pool_id must be positive")
	}
	seen := map[uint64]struct{}{}
	pairs := map[string]struct{}{}
	var maxID uint64
	for _, pool := range gs.Pools {
		if pool.Id == 0 {
			return fmt.Errorf("pool id must be positive")
		}
		if _, ok := seen[pool.Id]; ok {
			return fmt.Errorf("duplicate pool id %d", pool.Id)
		}
		seen[pool.Id] = struct{}{}
		if pool.Id > maxID {
			maxID = pool.Id
		}
		if err := sdk.ValidateDenom(pool.Denom0); err != nil {
			return fmt.Errorf("invalid denom0 for pool %d: %w", pool.Id, err)
		}
		if err := sdk.ValidateDenom(pool.Denom1); err != nil {
			return fmt.Errorf("invalid denom1 for pool %d: %w", pool.Id, err)
		}
		if pool.Denom0 >= pool.Denom1 {
			return fmt.Errorf("pool %d denoms must be unique and canonical", pool.Id)
		}
		pair := pool.Denom0 + "/" + pool.Denom1
		if _, ok := pairs[pair]; ok {
			return fmt.Errorf("duplicate pool pair %s", pair)
		}
		pairs[pair] = struct{}{}
		if pool.LpDenom != fmt.Sprintf("%s/%d", LPDenomPrefix, pool.Id) {
			return fmt.Errorf("invalid lp denom for pool %d", pool.Id)
		}
		if err := validatePositiveInt("reserve0", pool.Reserve0); err != nil {
			return fmt.Errorf("invalid reserve0 for pool %d: %w", pool.Id, err)
		}
		if err := validatePositiveInt("reserve1", pool.Reserve1); err != nil {
			return fmt.Errorf("invalid reserve1 for pool %d: %w", pool.Id, err)
		}
		if err := validatePositiveInt("total_shares", pool.TotalShares); err != nil {
			return fmt.Errorf("invalid total_shares for pool %d: %w", pool.Id, err)
		}
	}
	if maxID >= gs.NextPoolId {
		return fmt.Errorf("next_pool_id must be greater than existing pool ids")
	}
	return nil
}

func validatePositiveInt(field, value string) error {
	out, ok := sdkmath.NewIntFromString(value)
	if !ok || !out.IsPositive() {
		return fmt.Errorf("%s must be a positive integer", field)
	}
	return nil
}
