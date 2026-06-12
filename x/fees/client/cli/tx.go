package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/sovereign-l1/l1/x/fees/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:				types.ModuleName,
		Short:				"Fee policy transactions",
		DisableFlagParsing:		true,
		SuggestionsMinimumDistance:	2,
		RunE:				client.ValidateCmd,
	}
	return cmd
}
