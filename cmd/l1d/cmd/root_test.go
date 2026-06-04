package cmd_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/cmd/l1d/cmd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestInitCmd(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",       // Test the init cmd
		"l1app-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", l1app.DefaultNodeHome))
}

func TestRootCommandBranding(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	require.Equal(t, "orbitalisd", rootCmd.Use)
	require.Contains(t, rootCmd.Short, "Orbitalis")
}

func TestObservabilityFlagsRegistered(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	require.NotNil(t, rootCmd.PersistentFlags().Lookup("observability-metrics"))
	require.NotNil(t, rootCmd.PersistentFlags().Lookup("observability-metrics-addr"))
}

func TestHomeFlagRegistration(t *testing.T) {
	homeDir := "/tmp/foo"

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"query",
		fmt.Sprintf("--%s", flags.FlagHome),
		homeDir,
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", l1app.DefaultNodeHome))

	result, err := rootCmd.Flags().GetString(flags.FlagHome)
	require.NoError(t, err)
	require.Equal(t, result, homeDir)
}
