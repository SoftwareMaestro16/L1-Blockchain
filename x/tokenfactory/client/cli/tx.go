package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/tokenfactory/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Token factory transactions",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(NewCreateDenomCmd(), NewMintCmd(), NewBurnCmd(), NewChangeAdminCmd())
	return cmd
}

func NewCreateDenomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-denom [subdenom]",
		Short: "Create a new factory denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := &types.MsgCreateDenom{Creator: orbitaladdress.FormatAccAddress(clientCtx.GetFromAddress()), Subdenom: args[0]}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [coin] [to-address]",
		Short: "Mint factory tokens",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}
			msg := &types.MsgMint{Sender: orbitaladdress.FormatAccAddress(clientCtx.GetFromAddress()), Amount: coin, MintToAddress: args[1]}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewBurnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [coin] [from-address]",
		Short: "Burn factory tokens",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}
			msg := &types.MsgBurn{Sender: orbitaladdress.FormatAccAddress(clientCtx.GetFromAddress()), Amount: coin, BurnFromAddress: args[1]}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewChangeAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "change-admin [denom] [new-admin]",
		Short: "Transfer factory denom admin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := &types.MsgChangeAdmin{Sender: orbitaladdress.FormatAccAddress(clientCtx.GetFromAddress()), Denom: args[0], NewAdmin: args[1]}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
