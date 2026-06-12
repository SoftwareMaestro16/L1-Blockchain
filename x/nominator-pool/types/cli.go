package types

import "github.com/spf13/cobra"

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:				ModuleName,
		Short:				"Nominator pool transaction commands",
		DisableFlagParsing:		true,
		SuggestionsMinimumDistance:	2,
		RunE:				cobra.NoArgs,
	}
	for _, use := range []string{
		"create-pool",
		"deposit-to-pool",
		"request-withdrawal",
		"cancel-withdrawal",
		"deposit",
		"request-unbond",
		"withdraw",
		"claim-rewards",
		"sync-rewards",
		"claim-staking-rewards",
		"claim-reputation",
		"top-up-reserve",
		"update-pool-commission",
		"change-pool-validator",
		"register-validator",
		"update-validator",
		"update-staking-params",
		"create-official-pool",
	} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Build " + use + " transaction", RunE: cobra.NoArgs})
	}
	return cmd
}

func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:				ModuleName,
		Short:				"Nominator pool query commands",
		DisableFlagParsing:		true,
		SuggestionsMinimumDistance:	2,
		RunE:				cobra.NoArgs,
	}
	for _, use := range []string{
		"pool",
		"pools",
		"pool-delegator",
		"pool-rewards",
		"pool-share",
		"pool-allocations",
		"stake-reputation",
		"account-reputation",
		"staking-rewards",
		"staking-proof",
		"pool-unbonding-queue",
	} {
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Run " + use + " query", RunE: cobra.NoArgs})
	}
	return cmd
}
