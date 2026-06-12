package cmd

import (
	"errors"
	"fmt"

	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log/v2"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	l1app "github.com/sovereign-l1/l1/app"
	appparams "github.com/sovereign-l1/l1/app/params"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// initCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {

	// CustomConfig defines an arbitrary custom config to extend app.toml.
	// If you don't need it, you can remove it.
	// If you wish to add fields that correspond to flags that aren't in the SDK server config,
	// this custom config can as well help.
	type CustomConfig struct {
		CustomField string `mapstructure:"custom-field"`
	}

	type CustomAppConfig struct {
		serverconfig.Config	`mapstructure:",squash"`

		Custom	CustomConfig	`mapstructure:"custom"`
	}

	srvCfg := serverconfig.DefaultConfig()

	srvCfg.MinGasPrices = fmt.Sprintf("0%s", appparams.BaseDenom)

	customAppConfig := CustomAppConfig{
		Config:	*srvCfg,
		Custom: CustomConfig{
			CustomField: "anything",
		},
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[custom]
# That field will be parsed by server.InterceptConfigsPreRunHandler and held by viper.
# Do not forget to add quotes around the value if it is a string.
custom-field = "{{ .Custom.CustomField }}"`

	return customAppTemplate, customAppConfig
}

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	basicManager module.BasicManager,
) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	debugCmd := debug.Cmd()
	debugCmd.AddCommand(NewAVMDebugCmd())

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, l1app.DefaultNodeHome),
		NewInitLocalnetCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		NewTestnetCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		NewFaucetCmd(),
		NewBalancesCmd(),
		NewValidatorsCmd(),
		NewSystemAddressesCmd(),
		NewInvariantsCmd(),
		debugCmd,
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, l1app.DefaultNodeHome),
		snapshot.Cmd(newApp),
		NewBankSpeedTest(),
		NewAddressCmd(),
		NewExecutionOSCmd(),
	)

	server.AddCommandsWithStartCmdOptions(rootCmd, l1app.DefaultNodeHome, newApp, appExport, server.StartCmdOptions{
		AddFlags: func(startCmd *cobra.Command) {
		},
	})

	rootCmd.AddCommand(
		server.StatusCommand(),
		topLevelGenesisAccountCmd(txConfig),
		topLevelGenTxCmd(txConfig, basicManager),
		topLevelCollectGenTxsCmd(txConfig, basicManager),
		genesisCommand(txConfig, basicManager),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)
}

// genesisCommand builds genesis-related `aetrad genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, basicManager, l1app.DefaultNodeHome)
	wrapAetraGenesisValidation(cmd)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:				"query",
		Aliases:			[]string{"q"},
		Short:				"Querying subcommands",
		DisableFlagParsing:		false,
		SuggestionsMinimumDistance:	2,
		RunE:				client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.WaitTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
		NewAVMQueryCmd(),
		NewSystemQueryCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:				"tx",
		Short:				"Transactions subcommands",
		DisableFlagParsing:		false,
		SuggestionsMinimumDistance:	2,
		RunE:				client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
		NewAVMTxCmd(),
		NewSystemTxCmd(),
	)

	return cmd
}

func topLevelGenesisAccountCmd(txConfig client.TxConfig) *cobra.Command {
	cmd := genutilcli.AddGenesisAccountCmd(l1app.DefaultNodeHome, txConfig.SigningContext().AddressCodec())
	cmd.Use = "add-genesis-account [address_or_key_name] [coin][,[coin]]"
	cmd.Short = "Add a genesis account using Aetra AE... addresses and naet balances"
	return cmd
}

func topLevelGenTxCmd(txConfig client.TxConfig, basicManager module.BasicManager) *cobra.Command {
	cmd := genutilcli.GenTxCmd(
		basicManager,
		txConfig,
		banktypes.GenesisBalancesIterator{},
		l1app.DefaultNodeHome,
		txConfig.SigningContext().ValidatorAddressCodec(),
	)
	cmd.Short = "Create a genesis validator transaction using naet self-delegation"
	return cmd
}

func topLevelCollectGenTxsCmd(txConfig client.TxConfig, basicManager module.BasicManager) *cobra.Command {
	gentxModule := basicManager[genutiltypes.ModuleName].(genutil.AppModuleBasic)
	cmd := genutilcli.CollectGenTxsCmd(
		banktypes.GenesisBalancesIterator{},
		l1app.DefaultNodeHome,
		gentxModule.GenTxValidator,
		txConfig.SigningContext().ValidatorAddressCodec(),
	)
	cmd.Short = "Collect gentxs into genesis for localnet/testnet startup"
	return cmd
}

// newApp creates the application
func newApp(
	logger log.Logger,
	db dbm.DB,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)
	return l1app.NewL1App(
		logger, db, true,
		appOpts,
		baseappOptions...,
	)
}

// appExport creates a new l1app (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	db dbm.DB,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	var l1App *l1app.L1App
	if height != -1 {
		l1App = l1app.NewL1App(logger, db, false, appOpts)

		if err := l1App.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		l1App = l1app.NewL1App(logger, db, true, appOpts)
	}

	return l1App.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
