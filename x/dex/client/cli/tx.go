package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/dex/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "DEX transactions",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(NewCreatePoolCmd(), NewAddLiquidityCmd(), NewRemoveLiquidityCmd(), NewSwapCmd())
	return cmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [coin-a] [coin-b]",
		Short: "Create a constant-product pool",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			a, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}
			b, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := &types.MsgCreatePool{Creator: clientCtx.GetFromAddress().String(), TokenA: a, TokenB: b}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewAddLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [pool-id] [coin-a] [coin-b] [min-shares]",
		Short: "Add liquidity to a pool",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			poolID, err := parseUint(args[0])
			if err != nil {
				return err
			}
			a, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			b, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}
			msg := &types.MsgAddLiquidity{Depositor: clientCtx.GetFromAddress().String(), PoolId: poolID, TokenA: a, TokenB: b, MinShares: args[3]}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewRemoveLiquidityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [pool-id] [shares]",
		Short: "Remove liquidity from a pool",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			poolID, err := parseUint(args[0])
			if err != nil {
				return err
			}
			shares, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := &types.MsgRemoveLiquidity{Withdrawer: clientCtx.GetFromAddress().String(), PoolId: poolID, Shares: shares}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewSwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap-exact-in [pool-id] [coin-in] [out-denom] [min-out]",
		Short: "Swap an exact input amount",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			poolID, err := parseUint(args[0])
			if err != nil {
				return err
			}
			in, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}
			msg := &types.MsgSwapExactAmountIn{Trader: clientCtx.GetFromAddress().String(), PoolId: poolID, TokenIn: in, TokenOutDenom: args[2], MinAmountOut: args[3]}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
