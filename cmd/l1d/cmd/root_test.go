package cmd_test

import (
	"fmt"
	"testing"

	cmtversion "github.com/cometbft/cometbft/version"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/cmd/l1d/cmd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdkversion "github.com/cosmos/cosmos-sdk/version"
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

func TestVersionMetadataDefaults(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	require.Equal(t, "Orbitalis", sdkversion.Name)
	require.Equal(t, "orbitalisd", sdkversion.AppName)
	require.NotEmpty(t, sdkversion.Version)
	require.NotEmpty(t, sdkversion.Commit)
	require.NotEmpty(t, sdkversion.BuildTags)

	extraInfo, ok := rootCmd.Context().Value(sdkversion.ContextKey{}).(sdkversion.ExtraInfo)
	require.True(t, ok)
	require.NotEmpty(t, extraInfo["build_date"])
	require.NotEmpty(t, extraInfo["dirty"])
	require.Equal(t, cmtversion.TMCoreSemVer, extraInfo["cometbft_version"])
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
