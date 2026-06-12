package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sovereign-l1/l1/x/fees/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:				types.ModuleName,
		Short:				"Fee policy queries",
		DisableFlagParsing:		true,
		SuggestionsMinimumDistance:	2,
		RunE:				client.ValidateCmd,
	}
	cmd.AddCommand(NewParamsCmd(), NewNetworkLoadCmd(), NewEstimateFeeCmd(), NewAccountingCmd(), NewModuleBalancesCmd())
	return cmd
}

func NewParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"params",
		Short:	"Query fee policy params",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewNetworkLoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"network-load",
		Short:	"Query current fee network load",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).NetworkLoad(cmd.Context(), &types.QueryNetworkLoadRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewEstimateFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"estimate-fee [gas-limit]",
		Short:	"Estimate required native fee for a gas limit",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gasLimit, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).EstimateFee(cmd.Context(), &types.QueryEstimateFeeRequest{GasLimit: gasLimit})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewAccountingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"accounting",
		Short:	"Query protocol fee accounting",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).Accounting(cmd.Context(), &types.QueryAccountingRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func NewModuleBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"module-balances",
		Short:	"Query protocol fee module account balances",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).ModuleBalances(cmd.Context(), &types.QueryModuleBalancesRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
