package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/dex/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) CreatePool(ctx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidPool.Wrap("empty create pool request")
	}
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	token0, token1, err := canonicalPair(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}
	if existingID, found, err := m.GetPoolIDByPair(ctx, token0.Denom, token1.Denom); err != nil {
		return nil, err
	} else if found {
		return nil, types.ErrPoolExists.Wrapf("pool already exists for pair %s/%s at id %d", token0.Denom, token1.Denom, existingID)
	}
	id, err := m.GetNextPoolID(ctx)
	if err != nil {
		return nil, err
	}
	lp := lpDenom(id)
	shares := minInt(token0.Amount, token1.Amount)
	if !shares.IsPositive() {
		return nil, types.ErrInvalidLiquidity.Wrap("initial shares must be positive")
	}
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, sdk.NewCoins(token0, token1)); err != nil {
		return nil, err
	}
	shareCoin := sdk.NewCoin(lp, shares)
	if err := m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(shareCoin)); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creator, sdk.NewCoins(shareCoin)); err != nil {
		return nil, err
	}
	pool := types.Pool{
		Id:          id,
		Denom0:      token0.Denom,
		Denom1:      token1.Denom,
		Reserve0:    intString(token0.Amount),
		Reserve1:    intString(token1.Amount),
		TotalShares: intString(shares),
		LpDenom:     lp,
	}
	if err := m.SetPool(ctx, pool); err != nil {
		return nil, err
	}
	if err := m.SetPoolPairIndex(ctx, pool); err != nil {
		return nil, err
	}
	if err := m.SetNextPoolID(ctx, id+1); err != nil {
		return nil, err
	}
	if _, err := m.assertPoolAccounting(ctx, pool); err != nil {
		return nil, err
	}
	return &types.MsgCreatePoolResponse{PoolId: id, LpDenom: lp, MintedShares: shareCoin}, nil
}

func (m msgServer) AddLiquidity(ctx context.Context, msg *types.MsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidLiquidity.Wrap("empty add liquidity request")
	}
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, err
	}
	pool, found, err := m.GetPool(ctx, msg.PoolId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrPoolNotFound.Wrapf("%d", msg.PoolId)
	}
	token0, token1, err := coinsForPool(pool, msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}
	if !token0.IsPositive() || !token1.IsPositive() {
		return nil, types.ErrInvalidLiquidity.Wrap("tokens must be positive")
	}
	state, err := m.assertPoolAccounting(ctx, pool)
	if err != nil {
		return nil, err
	}
	shares, err := calcLiquidityShares(state.reserve0, state.reserve1, state.totalShares, token0.Amount, token1.Amount)
	if err != nil {
		return nil, err
	}
	minShares, err := parseNonNegativeInt("min_shares", msg.MinShares)
	if err != nil {
		return nil, err
	}
	if !shares.IsPositive() || shares.LT(minShares) {
		return nil, types.ErrSlippage.Wrap("minted shares below minimum")
	}
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, sdk.NewCoins(token0, token1)); err != nil {
		return nil, err
	}
	shareCoin := sdk.NewCoin(pool.LpDenom, shares)
	if err := m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(shareCoin)); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, sdk.NewCoins(shareCoin)); err != nil {
		return nil, err
	}
	pool.Reserve0 = intString(state.reserve0.Add(token0.Amount))
	pool.Reserve1 = intString(state.reserve1.Add(token1.Amount))
	pool.TotalShares = intString(state.totalShares.Add(shares))
	if err := m.SetPool(ctx, pool); err != nil {
		return nil, err
	}
	if _, err := m.assertPoolAccounting(ctx, pool); err != nil {
		return nil, err
	}
	return &types.MsgAddLiquidityResponse{MintedShares: shareCoin}, nil
}

func (m msgServer) RemoveLiquidity(ctx context.Context, msg *types.MsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidLiquidity.Wrap("empty remove liquidity request")
	}
	withdrawer, err := sdk.AccAddressFromBech32(msg.Withdrawer)
	if err != nil {
		return nil, err
	}
	pool, found, err := m.GetPool(ctx, msg.PoolId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrPoolNotFound.Wrapf("%d", msg.PoolId)
	}
	if msg.Shares.Denom != pool.LpDenom || !msg.Shares.IsPositive() {
		return nil, types.ErrInvalidLiquidity.Wrap("invalid LP shares")
	}
	state, err := m.assertPoolAccounting(ctx, pool)
	if err != nil {
		return nil, err
	}
	if msg.Shares.Amount.GT(state.totalShares) {
		return nil, types.ErrInvalidLiquidity.Wrap("shares exceed pool supply")
	}
	if msg.Shares.Amount.Equal(state.totalShares) {
		return nil, types.ErrInvalidLiquidity.Wrap("cannot remove all liquidity")
	}
	amount0 := state.reserve0.Mul(msg.Shares.Amount).Quo(state.totalShares)
	amount1 := state.reserve1.Mul(msg.Shares.Amount).Quo(state.totalShares)
	if !amount0.IsPositive() || !amount1.IsPositive() {
		return nil, types.ErrInvalidLiquidity.Wrap("withdrawal amount rounds to zero")
	}
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, withdrawer, types.ModuleName, sdk.NewCoins(msg.Shares)); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(msg.Shares)); err != nil {
		return nil, err
	}
	out0, out1 := sdk.NewCoin(pool.Denom0, amount0), sdk.NewCoin(pool.Denom1, amount1)
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawer, sdk.NewCoins(out0, out1)); err != nil {
		return nil, err
	}
	pool.Reserve0 = intString(state.reserve0.Sub(amount0))
	pool.Reserve1 = intString(state.reserve1.Sub(amount1))
	pool.TotalShares = intString(state.totalShares.Sub(msg.Shares.Amount))
	if err := m.SetPool(ctx, pool); err != nil {
		return nil, err
	}
	if _, err := m.assertPoolAccounting(ctx, pool); err != nil {
		return nil, err
	}
	return &types.MsgRemoveLiquidityResponse{TokenA: out0, TokenB: out1}, nil
}

func (m msgServer) SwapExactAmountIn(ctx context.Context, msg *types.MsgSwapExactAmountIn) (*types.MsgSwapExactAmountInResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidLiquidity.Wrap("empty swap request")
	}
	trader, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, err
	}
	pool, found, err := m.GetPool(ctx, msg.PoolId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.ErrPoolNotFound.Wrapf("%d", msg.PoolId)
	}
	if !msg.TokenIn.IsValid() || !msg.TokenIn.IsPositive() {
		return nil, types.ErrInvalidLiquidity.Wrap("token_in must be positive")
	}
	state, err := m.assertPoolAccounting(ctx, pool)
	if err != nil {
		return nil, err
	}
	var out sdk.Coin
	if msg.TokenIn.Denom == pool.Denom0 && msg.TokenOutDenom == pool.Denom1 {
		if !calcAmountInAfterFee(msg.TokenIn.Amount).IsPositive() {
			return nil, types.ErrInvalidLiquidity.Wrap("effective input after fee rounds to zero")
		}
		outAmt := calcSwapOut(state.reserve0, state.reserve1, msg.TokenIn.Amount)
		out = sdk.NewCoin(pool.Denom1, outAmt)
		if !outAmt.IsPositive() || !outAmt.LT(state.reserve1) {
			return nil, types.ErrInvalidLiquidity.Wrap("invalid swap output")
		}
		newReserve0 := state.reserve0.Add(msg.TokenIn.Amount)
		newReserve1 := state.reserve1.Sub(outAmt)
		if err := assertConstantProductNotDecreased(state.reserve0, state.reserve1, newReserve0, newReserve1); err != nil {
			return nil, err
		}
		pool.Reserve0 = intString(newReserve0)
		pool.Reserve1 = intString(newReserve1)
	} else if msg.TokenIn.Denom == pool.Denom1 && msg.TokenOutDenom == pool.Denom0 {
		if !calcAmountInAfterFee(msg.TokenIn.Amount).IsPositive() {
			return nil, types.ErrInvalidLiquidity.Wrap("effective input after fee rounds to zero")
		}
		outAmt := calcSwapOut(state.reserve1, state.reserve0, msg.TokenIn.Amount)
		out = sdk.NewCoin(pool.Denom0, outAmt)
		if !outAmt.IsPositive() || !outAmt.LT(state.reserve0) {
			return nil, types.ErrInvalidLiquidity.Wrap("invalid swap output")
		}
		newReserve1 := state.reserve1.Add(msg.TokenIn.Amount)
		newReserve0 := state.reserve0.Sub(outAmt)
		if err := assertConstantProductNotDecreased(state.reserve0, state.reserve1, newReserve0, newReserve1); err != nil {
			return nil, err
		}
		pool.Reserve1 = intString(newReserve1)
		pool.Reserve0 = intString(newReserve0)
	} else {
		return nil, types.ErrInvalidPool.Wrap("swap denoms do not match pool")
	}
	minOut, err := parseNonNegativeInt("min_amount_out", msg.MinAmountOut)
	if err != nil {
		return nil, err
	}
	if !out.Amount.IsPositive() || out.Amount.LT(minOut) {
		return nil, types.ErrSlippage.Wrap("amount out below minimum")
	}
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, sdk.NewCoins(msg.TokenIn)); err != nil {
		return nil, err
	}
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, sdk.NewCoins(out)); err != nil {
		return nil, err
	}
	if err := m.SetPool(ctx, pool); err != nil {
		return nil, err
	}
	if _, err := m.assertPoolAccounting(ctx, pool); err != nil {
		return nil, err
	}
	return &types.MsgSwapExactAmountInResponse{TokenOut: out}, nil
}
