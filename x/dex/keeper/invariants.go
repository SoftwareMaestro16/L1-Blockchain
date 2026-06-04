package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sovereign-l1/l1/x/dex/types"
)

type poolState struct {
	reserve0    sdkmath.Int
	reserve1    sdkmath.Int
	totalShares sdkmath.Int
}

func validatePoolState(pool types.Pool) (reserve0, reserve1, totalShares sdkmath.Int, err error) {
	if pool.Id == 0 {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("pool id must be positive")
	}
	if err := sdk.ValidateDenom(pool.Denom0); err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid denom0: %v", err)
	}
	if err := sdk.ValidateDenom(pool.Denom1); err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid denom1: %v", err)
	}
	if pool.Denom0 >= pool.Denom1 {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("pool denoms must be unique and canonical")
	}
	if pool.LpDenom != lpDenom(pool.Id) {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrap("invalid lp denom")
	}
	reserve0, err = parsePositiveInt("reserve0", pool.Reserve0)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	reserve1, err = parsePositiveInt("reserve1", pool.Reserve1)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	totalShares, err = parsePositiveInt("total_shares", pool.TotalShares)
	if err != nil {
		return sdkmath.Int{}, sdkmath.Int{}, sdkmath.Int{}, types.ErrInvalidPool.Wrapf("invalid pool state: %v", err)
	}
	return reserve0, reserve1, totalShares, nil
}

func (k Keeper) assertPoolAccounting(ctx context.Context, pool types.Pool) (poolState, error) {
	reserve0, reserve1, totalShares, err := validatePoolState(pool)
	if err != nil {
		return poolState{}, err
	}

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	if balance := k.bankKeeper.GetBalance(ctx, moduleAddr, pool.Denom0); !balance.Amount.Equal(reserve0) {
		return poolState{}, types.ErrInvariant.Wrapf("module balance mismatch for %s: reserve=%s balance=%s", pool.Denom0, reserve0, balance.Amount)
	}
	if balance := k.bankKeeper.GetBalance(ctx, moduleAddr, pool.Denom1); !balance.Amount.Equal(reserve1) {
		return poolState{}, types.ErrInvariant.Wrapf("module balance mismatch for %s: reserve=%s balance=%s", pool.Denom1, reserve1, balance.Amount)
	}
	if supply := k.bankKeeper.GetSupply(ctx, pool.LpDenom); !supply.Amount.Equal(totalShares) {
		return poolState{}, types.ErrInvariant.Wrapf("LP supply mismatch for %s: pool=%s supply=%s", pool.LpDenom, totalShares, supply.Amount)
	}

	return poolState{reserve0: reserve0, reserve1: reserve1, totalShares: totalShares}, nil
}
