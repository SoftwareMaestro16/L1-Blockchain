package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	flagFaucetAmount	= "amount"
	flagFaucetFees		= "fees"
	flagFaucetFromKey	= "from-key"
	flagFaucetFromHome	= "from-home"
)

type operatorCommandPlan struct {
	Command		string		`json:"command"`
	Equivalent	[]string	`json:"equivalent_args"`
	RPCPath		string		`json:"rpc_path,omitempty"`
	Denom		string		`json:"denom,omitempty"`
	Notes		[]string	`json:"notes,omitempty"`
}

func NewFaucetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"faucet",
		Short:	"Localnet faucet helpers using genesis-funded test keys",
	}
	cmd.AddCommand(newFaucetSendCmd())
	return cmd
}

func newFaucetSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"send [recipient]",
		Short:	"Build a localnet faucet transfer command",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			recipient := strings.TrimSpace(args[0])
			if err := addressing.ValidateUserRecipientAddress(recipient); err != nil {
				return err
			}
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if !strings.Contains(chainID, "local") {
				return fmt.Errorf("faucet refuses non-local chain-id %q", chainID)
			}
			amount, _ := cmd.Flags().GetString(flagFaucetAmount)
			fees, _ := cmd.Flags().GetString(flagFaucetFees)
			if err := requireNaetCoin("faucet amount", amount); err != nil {
				return err
			}
			if err := requireNaetCoin("faucet fees", fees); err != nil {
				return err
			}
			node, _ := cmd.Flags().GetString(flags.FlagNode)
			fromKey, _ := cmd.Flags().GetString(flagFaucetFromKey)
			fromHome, _ := cmd.Flags().GetString(flagFaucetFromHome)
			return writeCommandJSON(cmd, operatorCommandPlan{
				Command:	"faucet send",
				Equivalent: []string{
					"scripts/localnet/fund.ps1",
					"-ChainId", chainID,
					"-RPCPort", rpcPortHint(node),
					"-FromHome", fromHome,
					"-FromKey", fromKey,
					"-Recipients", recipient,
					"-Amount", amount,
					"-Fees", fees,
				},
				Denom:	appparams.BaseDenom,
				Notes: []string{
					"local-only",
					"uses normal bank send from genesis-funded localnet key",
					"does not mint and does not edit genesis",
				},
			})
		},
	}
	cmd.Flags().String(flags.FlagChainID, "aetra-local-1", "local chain id; non-local chain ids are rejected")
	cmd.Flags().String(flags.FlagNode, "tcp://127.0.0.1:26657", "RPC node")
	cmd.Flags().String(flagFaucetAmount, "1000000"+appparams.BaseDenom, "faucet transfer amount")
	cmd.Flags().String(flagFaucetFees, "1000000"+appparams.BaseDenom, "bank send fees")
	cmd.Flags().String(flagFaucetFromKey, "node0", "localnet key name that funds faucet sends")
	cmd.Flags().String(flagFaucetFromHome, ".localnet/node0/aetrad", "localnet key home")
	return cmd
}

func NewBalancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"balances [AE-address]",
		Short:	"Convenience alias for querying naet account balances",
		Args:	cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := strings.TrimSpace(args[0])
			if err := addressing.ValidateUserAddress("balance address", address); err != nil {
				return err
			}
			node, _ := cmd.Flags().GetString(flags.FlagNode)
			return writeCommandJSON(cmd, operatorCommandPlan{
				Command:	"balances",
				Equivalent:	[]string{"query", "bank", "balances", address, "--denom", appparams.BaseDenom, "--node", node, "--output", "json"},
				RPCPath:	"/cosmos.bank.v1beta1.Query/AllBalances",
				Denom:		appparams.BaseDenom,
			})
		},
	}
	cmd.Flags().String(flags.FlagNode, "tcp://127.0.0.1:26657", "RPC node")
	return cmd
}

func NewValidatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"validators",
		Short:	"Convenience alias for querying active validators",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			node, _ := cmd.Flags().GetString(flags.FlagNode)
			return writeCommandJSON(cmd, operatorCommandPlan{
				Command:	"validators",
				Equivalent:	[]string{"query", "staking", "validators", "--node", node, "--output", "json"},
				RPCPath:	"/cosmos.staking.v1beta1.Query/Validators",
				Denom:		appparams.BaseDenom,
			})
		},
	}
	cmd.Flags().String(flags.FlagNode, "tcp://127.0.0.1:26657", "RPC node")
	return cmd
}

func NewSystemAddressesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:	"system-addresses",
		Short:	"List reserved Aetra system addresses",
		Args:	cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return writeCommandJSON(cmd, struct {
				Command		string				`json:"command"`
				Count		int				`json:"count"`
				Addresses	[]addressing.SystemAddress	`json:"addresses"`
			}{
				Command:	"system-addresses",
				Count:		len(addressing.AllSystemAddresses()),
				Addresses:	addressing.AllSystemAddresses(),
			})
		},
	}
	return cmd
}

func NewSystemQueryCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "system", Short: "System module query helpers"}
	for _, spec := range systemModuleSpecs() {
		moduleCmd := &cobra.Command{Use: spec.module, Short: spec.short + " query helpers"}
		for _, query := range spec.queries {
			moduleCmd.AddCommand(newSystemPlanLeaf(query, "query", spec.module))
		}
		cmd.AddCommand(moduleCmd)
	}
	cmd.AddCommand(NewSystemAddressesCmd())
	return cmd
}

func NewSystemTxCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "system", Short: "System module transaction helpers"}
	for _, spec := range systemModuleSpecs() {
		moduleCmd := &cobra.Command{Use: spec.module, Short: spec.short + " transaction helpers"}
		for _, txName := range spec.txs {
			moduleCmd.AddCommand(newSystemPlanLeaf(txName, "tx", spec.module))
		}
		cmd.AddCommand(moduleCmd)
	}
	return cmd
}

type systemModuleSpec struct {
	module	string
	short	string
	queries	[]string
	txs	[]string
}

func systemModuleSpecs() []systemModuleSpec {
	return []systemModuleSpec{
		{module: "config", short: "governance config", queries: []string{"params", "entries", "pending-changes"}, txs: []string{"submit-change", "approve-change", "execute-change"}},
		{module: "constitution", short: "constitution", queries: []string{"constitution", "pending-amendments", "protected-limits"}, txs: []string{"propose-amendment", "vote-amendment", "execute-amendment"}},
		{module: "system-registry", short: "system registry", queries: []string{"reserved-addresses", "entities", "dependency-graph"}, txs: []string{"register-entity", "update-entity", "pause-entity", "resume-entity"}},
		{module: "fees", short: "fee policy", queries: []string{"params"}, txs: []string{"update-params"}},
		{module: "native-account", short: "native account", queries: []string{"account", "status", "params"}, txs: []string{"activate-account", "pay-storage-debt", "unfreeze-account"}},
		{module: "storage-rent", short: "storage rent", queries: []string{"params", "account-rent", "system-reserve"}, txs: []string{"pay-debt", "top-up-reserve"}},
		{module: "validator-registry", short: "validator registry", queries: []string{"validators", "validator", "params"}, txs: []string{"register-validator", "update-validator", "set-commission"}},
	}
}

func newSystemPlanLeaf(name, kind, module string) *cobra.Command {
	return &cobra.Command{
		Use:	name,
		Short:	fmt.Sprintf("Build %s %s/%s request", kind, module, name),
		Args:	cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeCommandJSON(cmd, operatorCommandPlan{
				Command:	kind + " system " + module + " " + name,
				Equivalent:	append([]string{kind, module, name}, args...),
				Notes:		[]string{"system module command surface", "use module proto tx/query service when wired"},
			})
		},
	}
}

func requireNaetCoin(field, coin string) error {
	coin = strings.TrimSpace(coin)
	if coin == "" {
		return fmt.Errorf("%s is required", field)
	}
	if !strings.HasSuffix(coin, appparams.BaseDenom) {
		return fmt.Errorf("%s must use %s denom", field, appparams.BaseDenom)
	}
	amount := strings.TrimSuffix(coin, appparams.BaseDenom)
	if amount == "" || strings.HasPrefix(amount, "-") {
		return fmt.Errorf("%s amount must be positive", field)
	}
	for _, ch := range amount {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("%s amount must be an integer %s coin", field, appparams.BaseDenom)
		}
	}
	if strings.TrimLeft(amount, "0") == "" {
		return fmt.Errorf("%s amount must be positive", field)
	}
	return nil
}

func rpcPortHint(node string) string {
	node = strings.TrimSpace(node)
	if node == "" {
		return "26657"
	}
	idx := strings.LastIndex(node, ":")
	if idx < 0 || idx == len(node)-1 {
		return "26657"
	}
	return strings.TrimRight(node[idx+1:], "/")
}
