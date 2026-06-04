package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/observability"
	"github.com/sovereign-l1/l1/x/dex/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

func (m msgServer) CreatePool(ctx context.Context, msg *types.MsgCreatePool) (res *types.MsgCreatePoolResponse, err error) {
	defer recordDexResult("create_pool", &err)
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.PoolCreationEnabled {
		return nil, types.ErrOperationDisabled.Wrap("pool creation is disabled")
	}
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}
	token0, token1, err := canonicalPair(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}
	if _, found, err := m.GetPoolIDByPair(ctx, token0.Denom, token1.Denom); err != nil {
		return nil, err
	} else if found {
		return nil, types.ErrInvalidPool.Wrap("pool pair already exists")
	}
	id, err := m.GetNextPoolID(ctx)
	if err != nil {
		return nil, err
	}
	lp := lpDenom(id)
	shares := minInt(token0.Amount, token1.Amount)
	minInitialLiquidity, err := parsePositiveInt("min_initial_liquidity", params.MinInitialLiquidity)
	if err != nil {
		return nil, err
	}
	if !shares.IsPositive() {
		return nil, types.ErrInvalidLiquidity.Wrap("initial shares must be positive")
	}
	if shares.LT(minInitialLiquidity) {
		return nil, types.ErrInvalidLiquidity.Wrap("initial shares below protocol minimum")
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
	if err := m.SetNextPoolID(ctx, id+1); err != nil {
		return nil, err
	}
	observability.RecordDexPoolCreated()
	recordNorbLiquidityDelta(token0, token1)
	return &types.MsgCreatePoolResponse{PoolId: id, LpDenom: lp, MintedShares: shareCoin}, nil
}

func (m msgServer) AddLiquidity(ctx context.Context, msg *types.MsgAddLiquidity) (res *types.MsgAddLiquidityResponse, err error) {
	defer recordDexResult("add_liquidity", &err)
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.LiquidityEnabled {
		return nil, types.ErrOperationDisabled.Wrap("liquidity operations are disabled")
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
	reserve0, reserve1, totalShares, err := validatePoolState(pool)
	if err != nil {
		return nil, err
	}
	shares0 := token0.Amount.Mul(totalShares).Quo(reserve0)
	shares1 := token1.Amount.Mul(totalShares).Quo(reserve1)
	shares := minInt(shares0, shares1)
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
	pool.Reserve0 = intString(reserve0.Add(token0.Amount))
	pool.Reserve1 = intString(reserve1.Add(token1.Amount))
	pool.TotalShares = intString(totalShares.Add(shares))
	if err := m.SetPool(ctx, pool); err != nil {
		return nil, err
	}
	recordNorbLiquidityDelta(token0, token1)
	return &types.MsgAddLiquidityResponse{MintedShares: shareCoin}, nil
}

func (m msgServer) RemoveLiquidity(ctx context.Context, msg *types.MsgRemoveLiquidity) (res *types.MsgRemoveLiquidityResponse, err error) {
	defer recordDexResult("remove_liquidity", &err)
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.LiquidityEnabled {
		return nil, types.ErrOperationDisabled.Wrap("liquidity operations are disabled")
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
	reserve0, reserve1, totalShares, err := validatePoolState(pool)
	if err != nil {
		return nil, err
	}
	if msg.Shares.Amount.GT(totalShares) {
		return nil, types.ErrInvalidLiquidity.Wrap("shares exceed pool supply")
	}
	amount0 := reserve0.Mul(msg.Shares.Amount).Quo(totalShares)
	amount1 := reserve1.Mul(msg.Shares.Amount).Quo(totalShares)
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
	pool.Reserve0 = intString(reserve0.Sub(amount0))
	pool.Reserve1 = intString(reserve1.Sub(amount1))
	pool.TotalShares = intString(totalShares.Sub(msg.Shares.Amount))
	if err := m.SetPool(ctx, pool); err != nil {
		return nil, err
	}
	recordNorbLiquidityDelta(negativeCoin(out0), negativeCoin(out1))
	return &types.MsgRemoveLiquidityResponse{TokenA: out0, TokenB: out1}, nil
}

func (m msgServer) SwapExactAmountIn(ctx context.Context, msg *types.MsgSwapExactAmountIn) (res *types.MsgSwapExactAmountInResponse, err error) {
	defer recordDexResult("swap_exact_amount_in", &err)
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if !params.SwapsEnabled {
		return nil, types.ErrOperationDisabled.Wrap("swaps are disabled")
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
	reserve0, reserve1, _, err := validatePoolState(pool)
	if err != nil {
		return nil, err
	}
	var out sdk.Coin
	if msg.TokenIn.Denom == pool.Denom0 && msg.TokenOutDenom == pool.Denom1 {
		outAmt := calcSwapOut(reserve0, reserve1, msg.TokenIn.Amount, params.SwapFeeBps)
		out = sdk.NewCoin(pool.Denom1, outAmt)
		pool.Reserve0 = intString(reserve0.Add(msg.TokenIn.Amount))
		pool.Reserve1 = intString(reserve1.Sub(outAmt))
	} else if msg.TokenIn.Denom == pool.Denom1 && msg.TokenOutDenom == pool.Denom0 {
		outAmt := calcSwapOut(reserve1, reserve0, msg.TokenIn.Amount, params.SwapFeeBps)
		out = sdk.NewCoin(pool.Denom0, outAmt)
		pool.Reserve1 = intString(reserve1.Add(msg.TokenIn.Amount))
		pool.Reserve0 = intString(reserve0.Sub(outAmt))
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
	observability.RecordDexSwap()
	recordNorbLiquidityDelta(msg.TokenIn, negativeCoin(out))
	return &types.MsgSwapExactAmountInResponse{TokenOut: out}, nil
}

func (m msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (res *types.MsgUpdateParamsResponse, err error) {
	defer recordDexResult("update_params", &err)
	if msg == nil {
		return nil, types.ErrInvalidParams.Wrap("empty request")
	}
	if msg.Authority != m.Authority() {
		return nil, types.ErrUnauthorized.Wrap("invalid authority")
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}

func recordDexResult(action string, err *error) {
	if err != nil && *err != nil {
		observability.RecordModuleError(types.ModuleName, action, "error")
	}
}

func recordNorbLiquidityDelta(coins ...sdk.Coin) {
	for _, coin := range coins {
		if coin.Denom != "norb" || !coin.Amount.IsInt64() {
			continue
		}
		observability.RecordDexLiquidityNorbDelta(coin.Amount.Int64())
	}
}

func negativeCoin(coin sdk.Coin) sdk.Coin {
	return sdk.Coin{Denom: coin.Denom, Amount: coin.Amount.Neg()}
}
