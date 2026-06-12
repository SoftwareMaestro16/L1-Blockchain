package cmd

import (
	"context"
	"os"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"cosmossdk.io/log/v2"
	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/app/params"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// NewRootCmd creates a new root command for aetrad. It is called once in the
// main function.
func NewRootCmd() *cobra.Command {
	extraVersionInfo := initVersionInfo()

	tempApp := l1app.NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(l1app.DefaultNodeHome))
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry:	tempApp.InterfaceRegistry(),
		Codec:			tempApp.AppCodec(),
		TxConfig:		tempApp.TxConfig(),
		Amino:			tempApp.LegacyAmino(),
	}

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(l1app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:		"aetrad",
		Short:		"Aetra sovereign L1 app",
		SilenceErrors:	true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {

			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if !initClientCtx.Offline {
				enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
				txConfigOpts := tx.ConfigOptions{
					EnabledSignModes:		enabledSignModes,
					SigningOptions:			&txsigning.Options{AddressCodec: aetraaddress.Codec{}, ValidatorAddressCodec: aetraaddress.Codec{}},
					TextualCoinMetadataQueryFn:	authtxconfig.NewGRPCCoinMetadataQueryFn(initClientCtx),
				}
				txConfig, err := tx.NewTxConfigWithOptions(
					initClientCtx.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			if err := server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig); err != nil {
				return err
			}
			return startObservabilityMetrics(cmd)
		},
	}
	rootCmd.PersistentFlags().Bool(flagObservabilityMetrics, false, "enable Aetra process Prometheus metrics endpoint")
	rootCmd.PersistentFlags().String(flagObservabilityMetricsAddr, "127.0.0.1:27660", "Aetra process metrics listen address")
	rootCmd.SetContext(context.WithValue(context.Background(), version.ContextKey{}, extraVersionInfo))

	initRootCmd(rootCmd, encodingConfig.TxConfig, tempApp.BasicModuleManager)
	if versionCmd, _, err := rootCmd.Find([]string{"version"}); err == nil && versionCmd != nil {
		versionCmd.SetContext(context.WithValue(context.Background(), version.ContextKey{}, extraVersionInfo))
	}

	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.ClientCtx = initClientCtx

	nodeCmds := nodeservice.NewNodeCommands()
	autoCliOpts.ModuleOptions[nodeCmds.Name()] = nodeCmds.AutoCLIOptions()

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}
