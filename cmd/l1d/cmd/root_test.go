package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	"github.com/sovereign-l1/l1/cmd/l1d/cmd"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestInitCmd(t *testing.T) {
	homeDir := t.TempDir()
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",
		"l1app-test",
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir),
	})

	require.NoError(t, svrcmd.Execute(rootCmd, "", homeDir))
}

func TestRootCommandBranding(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	require.Equal(t, "aetrad", rootCmd.Use)
	require.Contains(t, rootCmd.Short, "Aetra")
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

func TestRootHelpShowsOperatorCommandSurface(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"--help"})

	require.NoError(t, rootCmd.Execute())

	help := out.String()
	require.Contains(t, help, "Aetra sovereign L1 app")
	require.Contains(t, help, "version")
	require.Contains(t, help, "testnet")
	require.Contains(t, help, "query")
	require.Contains(t, help, "tx")
	require.NotContains(t, strings.ToLower(help), "mnemonic")
}

func TestVersionCommandShowsOperatorMetadata(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"version", "--long", "--output", "json"})

	require.NoError(t, svrcmd.Execute(rootCmd, "", t.TempDir()))

	var info struct {
		Name			string			`json:"name"`
		ServerName		string			`json:"server_name"`
		Version			string			`json:"version"`
		Commit			string			`json:"commit"`
		CosmosSDKVersion	string			`json:"cosmos_sdk_version"`
		ExtraInfo		map[string]string	`json:"extra_info"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &info), out.String())
	require.Equal(t, "Aetra", info.Name)
	require.Equal(t, "aetrad", info.ServerName)
	require.NotEmpty(t, info.Version)
	require.NotEmpty(t, info.Commit)
	require.NotEmpty(t, firstNonEmpty(info.CosmosSDKVersion, info.ExtraInfo["cosmos_sdk_version"]))
	require.NotEmpty(t, info.ExtraInfo["build_date"])
	require.NotEmpty(t, info.ExtraInfo["dirty"])
	require.NotEmpty(t, info.ExtraInfo["cometbft_version"])
	require.NotEmpty(t, info.ExtraInfo["avm_version"])
	require.Equal(t, "1", info.ExtraInfo["avm_version"])
}

func TestPrototypeCommandsAreRegistered(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	for _, path := range [][]string{
		{"address", "convert"},
		{"execution-os", "profiles"},
		{"execution-os", "smoke"},
		{"execution-os", "diagnostics"},
		{"init-localnet"},
		{"add-genesis-account"},
		{"gentx"},
		{"collect-gentxs"},
		{"faucet", "send"},
		{"balances"},
		{"validators"},
		{"system-addresses"},
		{"invariants", "list"},
		{"invariants", "check"},
		{"query", "block"},
		{"query", "bank", "balance"},
		{"query", "staking", "validators"},
		{"query", "slashing", "params"},
		{"query", "fees", "params"},
		{"query", "system", "config", "params"},
		{"query", "system", "system-registry", "reserved-addresses"},
		{"tx", "bank", "send"},
		{"tx", "staking", "delegate"},
		{"tx", "system", "config", "submit-change"},
		{"tx", "system", "validator-registry", "register-validator"},
		{"testnet", "init-files"},
		{"testnet", "start"},
	} {
		found := requireCommand(t, rootCmd, path...)
		require.NotEmpty(t, found.Short, strings.Join(path, " "))
	}
}

func TestAddressConvertCommandOutputsRawAndUserFriendly(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	homeDir := t.TempDir()
	rootCmd.SetArgs([]string{"address", "convert", "4:00000000000000000000000020dbf996b75fdc4e208146e0ca920168148149ca", fmt.Sprintf("--%s=%s", flags.FlagHome, homeDir)})

	require.NoError(t, svrcmd.Execute(rootCmd, "", homeDir))

	var res struct {
		Raw		string	`json:"raw"`
		UserFriendly	string	`json:"user_friendly"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &res), out.String())
	require.Regexp(t, `^4:[0-9a-f]{64}$`, res.Raw)
	require.Regexp(t, `^AE[A-Za-z0-9_-]{46}$`, res.UserFriendly)
	require.Len(t, res.UserFriendly, 48)
}

func TestOperatorTxCommandsExposeCommonFlags(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	flagNames := []string{"from", "chain-id", "keyring-backend", "fees", "node", "output"}

	for _, path := range [][]string{
		{"tx", "bank", "send"},
		{"tx", "staking", "delegate"},
	} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			command := requireCommand(t, rootCmd, path...)
			for _, flagName := range flagNames {
				require.NotNil(t, command.Flag(flagName), "%s missing --%s", strings.Join(path, " "), flagName)
			}
		})
	}
}

func TestPrototypeCommandArgsValidation(t *testing.T) {
	rootCmd := cmd.NewRootCmd()

	tests := []struct {
		name	string
		path	[]string
		args	[]string
		wantErr	bool
	}{
		{
			name:	"fees params accepts no args",
			path:	[]string{"query", "fees", "params"},
		},
		{
			name:		"fees params rejects extra arg",
			path:		[]string{"query", "fees", "params"},
			args:		[]string{"extra"},
			wantErr:	true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			command := requireCommand(t, rootCmd, tc.path...)
			err := command.Args(command, tc.args)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTestnetStartDoesNotPrintMnemonicByDefault(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	startCmd := requireCommand(t, rootCmd, "testnet", "start")

	flag := startCmd.Flags().Lookup("print-mnemonic")
	require.NotNil(t, flag)
	require.Equal(t, "false", flag.DefValue)
}

func TestInitLocalnetUsesNaetDefaults(t *testing.T) {
	rootCmd := cmd.NewRootCmd()
	initCmd := requireCommand(t, rootCmd, "init-localnet")

	validatorCount := initCmd.Flags().Lookup("validator-count")
	require.NotNil(t, validatorCount)
	require.Equal(t, "1", validatorCount.DefValue)
	stakingDenom := initCmd.Flags().Lookup("staking-denom")
	require.NotNil(t, stakingDenom)
	require.Equal(t, "naet", stakingDenom.DefValue)
	minGas := initCmd.Flags().Lookup("minimum-gas-prices")
	require.NotNil(t, minGas)
	require.Equal(t, "0naet", minGas.DefValue)
}

func requireCommand(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()

	found, _, err := root.Find(path)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, path[len(path)-1], found.Name())
	return found
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
