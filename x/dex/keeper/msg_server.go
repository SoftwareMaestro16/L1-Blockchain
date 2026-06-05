package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/observability"
	"github.com/sovereign-l1/l1/x/dex/types"
	txutil "github.com/sovereign-l1/l1/x/internal/tx"
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
	creator, err := orbitaladdress.ParseAccAddress(msg.Creator)
	if err != nil {
		return nil, err
	}
	creatorText := orbitaladdress.FormatAccAddress(creator)
	token0, token1, err := canonicalPair(msg.TokenA, msg.TokenB)
	if err != nil {
		return nil, err
	}
	if existingID, found, err := m.GetPoolIDByPair(ctx, token0.Denom, token1.Denom); err != nil {
		return nil, err
	} else if found {
		return nil, types.ErrInvalidPool.Wrapf("pool already exists for pair %s/%s: %d", token0.Denom, token1.Denom, existingID)
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
		return nil, types.ErrInvalidLiquidity.Wrap("initial shares below minimum")
	}
	shareCoin := sdk.NewCoin(lp, shares)
	pool := types.Pool{
		Id:          id,
		Denom0:      token0.Denom,
		Denom1:      token1.Denom,
		Reserve0:    intString(token0.Amount),
		Reserve1:    intString(token1.Amount),
		TotalShares: intString(shares),
		LpDenom:     lp,
	}
	if err := txutil.AtomicStateChange(ctx, func(cacheCtx context.Context) error {
		if err := m.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, creator, types.ModuleName, sdk.NewCoins(token0, token1)); err != nil {
			return err
		}
		if err := m.bankKeeper.MintCoins(cacheCtx, types.ModuleName, sdk.NewCoins(shareCoin)); err != nil {
			return err
		}
		if err := m.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, creator, sdk.NewCoins(shareCoin)); err != nil {
			return err
		}
		if err := m.SetPool(cacheCtx, pool); err != nil {
			return err
		}
		if err := m.SetPoolPairIndex(cacheCtx, pool.Denom0, pool.Denom1, pool.Id); err != nil {
			return err
		}
		if err := m.SetNextPoolID(cacheCtx, id+1); err != nil {
			return err
		}
		sdk.UnwrapSDKContext(cacheCtx).EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeCreatePool,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(id, 10)),
			sdk.NewAttribute(types.AttributeKeyCreator, creatorText),
			sdk.NewAttribute(types.AttributeKeyDenom0, pool.Denom0),
			sdk.NewAttribute(types.AttributeKeyDenom1, pool.Denom1),
			sdk.NewAttribute(types.AttributeKeyAmount0, pool.Reserve0),
			sdk.NewAttribute(types.AttributeKeyAmount1, pool.Reserve1),
			sdk.NewAttribute(types.AttributeKeyLPDenom, lp),
			sdk.NewAttribute(types.AttributeKeyMintedShares, shares.String()),
		))
		return nil
	}); err != nil {
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
	depositor, err := orbitaladdress.ParseAccAddress(msg.Depositor)
	if err != nil {
		return nil, err
	}
	depositorText := orbitaladdress.FormatAccAddress(depositor)
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
	shareCoin := sdk.NewCoin(pool.LpDenom, shares)
	pool.Reserve0 = intString(reserve0.Add(token0.Amount))
	pool.Reserve1 = intString(reserve1.Add(token1.Amount))
	pool.TotalShares = intString(totalShares.Add(shares))
	if err := txutil.AtomicStateChange(ctx, func(cacheCtx context.Context) error {
		if err := m.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, depositor, types.ModuleName, sdk.NewCoins(token0, token1)); err != nil {
			return err
		}
		if err := m.bankKeeper.MintCoins(cacheCtx, types.ModuleName, sdk.NewCoins(shareCoin)); err != nil {
			return err
		}
		if err := m.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, depositor, sdk.NewCoins(shareCoin)); err != nil {
			return err
		}
		if err := m.SetPool(cacheCtx, pool); err != nil {
			return err
		}
		sdk.UnwrapSDKContext(cacheCtx).EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeAddLiquidity,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositorText),
			sdk.NewAttribute(types.AttributeKeyDenom0, pool.Denom0),
			sdk.NewAttribute(types.AttributeKeyDenom1, pool.Denom1),
			sdk.NewAttribute(types.AttributeKeyAmount0, token0.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyAmount1, token1.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyLPDenom, pool.LpDenom),
			sdk.NewAttribute(types.AttributeKeyMintedShares, shares.String()),
		))
		return nil
	}); err != nil {
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
	withdrawer, err := orbitaladdress.ParseAccAddress(msg.Withdrawer)
	if err != nil {
		return nil, err
	}
	withdrawerText := orbitaladdress.FormatAccAddress(withdrawer)
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
	out0, out1 := sdk.NewCoin(pool.Denom0, amount0), sdk.NewCoin(pool.Denom1, amount1)
	pool.Reserve0 = intString(reserve0.Sub(amount0))
	pool.Reserve1 = intString(reserve1.Sub(amount1))
	pool.TotalShares = intString(totalShares.Sub(msg.Shares.Amount))
	if err := txutil.AtomicStateChange(ctx, func(cacheCtx context.Context) error {
		if err := m.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, withdrawer, types.ModuleName, sdk.NewCoins(msg.Shares)); err != nil {
			return err
		}
		if err := m.bankKeeper.BurnCoins(cacheCtx, types.ModuleName, sdk.NewCoins(msg.Shares)); err != nil {
			return err
		}
		if err := m.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, withdrawer, sdk.NewCoins(out0, out1)); err != nil {
			return err
		}
		if err := m.SetPool(cacheCtx, pool); err != nil {
			return err
		}
		sdk.UnwrapSDKContext(cacheCtx).EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeRemoveLiquidity,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyWithdrawer, withdrawerText),
			sdk.NewAttribute(types.AttributeKeyLPDenom, msg.Shares.Denom),
			sdk.NewAttribute(types.AttributeKeyShares, msg.Shares.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyDenom0, out0.Denom),
			sdk.NewAttribute(types.AttributeKeyDenom1, out1.Denom),
			sdk.NewAttribute(types.AttributeKeyAmount0, out0.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyAmount1, out1.Amount.String()),
		))
		return nil
	}); err != nil {
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
	trader, err := orbitaladdress.ParseAccAddress(msg.Trader)
	if err != nil {
		return nil, err
	}
	traderText := orbitaladdress.FormatAccAddress(trader)
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
	if err := txutil.AtomicStateChange(ctx, func(cacheCtx context.Context) error {
		if err := m.bankKeeper.SendCoinsFromAccountToModule(cacheCtx, trader, types.ModuleName, sdk.NewCoins(msg.TokenIn)); err != nil {
			return err
		}
		if err := m.bankKeeper.SendCoinsFromModuleToAccount(cacheCtx, types.ModuleName, trader, sdk.NewCoins(out)); err != nil {
			return err
		}
		if err := m.SetPool(cacheCtx, pool); err != nil {
			return err
		}
		sdk.UnwrapSDKContext(cacheCtx).EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSwapExactAmountIn,
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(msg.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyTrader, traderText),
			sdk.NewAttribute(types.AttributeKeyTokenIn, msg.TokenIn.String()),
			sdk.NewAttribute(types.AttributeKeyTokenOut, out.String()),
		))
		return nil
	}); err != nil {
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
