package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	l1app "github.com/sovereign-l1/l1/app"
	appparams "github.com/sovereign-l1/l1/app/params"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

func Test_TestnetCmd(t *testing.T) {
	home := t.TempDir()
	testApp := l1app.NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(home))
	moduleBasic := testApp.BasicModuleManager
	cdc := testApp.AppCodec()
	txConfig := testApp.TxConfig()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	err = genutiltest.ExecInitCmd(moduleBasic, home, cdc)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(cdc).
		WithHomeDir(home).
		WithTxConfig(txConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	cmd := testnetInitFilesCmd(moduleBasic, banktypes.GenesisBalancesIterator{})
	outputDir := filepath.Join(home, "localnet")
	const validatorCount = 3
	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=test", flags.FlagKeyringBackend),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, "aetra-local-1"),
		fmt.Sprintf("--%s=%d", flagNumValidators, validatorCount),
		fmt.Sprintf("--%s=%s", flagOutputDir, outputDir),
		fmt.Sprintf("--%s=%s", flagStakingDenom, appparams.BaseDenom),
		fmt.Sprintf("--%s=0%s", server.FlagMinGasPrices, appparams.BaseDenom),
		fmt.Sprintf("--%s", flagSingleHost),
	})
	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)

	var firstGenesisHash string
	var firstGenesisRaw []byte
	for i := 0; i < validatorCount; i++ {
		genFile := filepath.Join(outputDir, fmt.Sprintf("node%d", i), "aetrad", "config", "genesis.json")
		genesisRaw, err := os.ReadFile(genFile)
		require.NoError(t, err)
		require.NotRegexp(t, `(?i)\b(mnemonic|private[_-]?key|priv_validator|secret|seed|wallet)\b`, string(genesisRaw))

		hash := sha256.Sum256(genesisRaw)
		genesisHash := hex.EncodeToString(hash[:])
		if i == 0 {
			firstGenesisHash = genesisHash
			firstGenesisRaw = genesisRaw
		} else {
			require.Equal(t, firstGenesisHash, genesisHash)
			require.JSONEq(t, string(firstGenesisRaw), string(genesisRaw))
		}

		appState, appGenesis, err := genutiltypes.GenesisStateFromGenFile(genFile)
		require.NoError(t, err)
		require.NoError(t, appGenesis.ValidateAndComplete())
		require.Equal(t, "aetra-local-1", appGenesis.ChainID)
		require.Equal(t, int64(1), appGenesis.InitialHeight)
		require.NoError(t, moduleBasic.ValidateGenesis(cdc, txConfig, appState))

		if i == 0 {
			assertPrototypeGenesisProfile(t, cdc, txConfig, appState, validatorCount)
		}
	}
}

func TestTestnetInitRejectsMalformedChainIDBeforeGenesisWrite(t *testing.T) {
	home := t.TempDir()
	testApp := l1app.NewL1App(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(home))
	moduleBasic := testApp.BasicModuleManager
	cdc := testApp.AppCodec()
	txConfig := testApp.TxConfig()
	logger := log.NewNopLogger()
	cfg, err := genutiltest.CreateDefaultCometConfig(home)
	require.NoError(t, err)

	err = genutiltest.ExecInitCmd(moduleBasic, home, cdc)
	require.NoError(t, err)

	serverCtx := server.NewContext(viper.New(), cfg, logger)
	clientCtx := client.Context{}.
		WithCodec(cdc).
		WithHomeDir(home).
		WithTxConfig(txConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
	cmd := testnetInitFilesCmd(moduleBasic, banktypes.GenesisBalancesIterator{})
	outputDir := filepath.Join(home, "bad-localnet")
	cmd.SetArgs([]string{
		fmt.Sprintf("--%s=test", flags.FlagKeyringBackend),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, "cosmoshub-4"),
		fmt.Sprintf("--%s=1", flagNumValidators),
		fmt.Sprintf("--%s=%s", flagOutputDir, outputDir),
		fmt.Sprintf("--%s=%s", flagStakingDenom, appparams.BaseDenom),
		fmt.Sprintf("--%s=0%s", server.FlagMinGasPrices, appparams.BaseDenom),
	})

	require.ErrorContains(t, cmd.ExecuteContext(ctx), "chain-id must start with aetra-")
	_, statErr := os.Stat(outputDir)
	require.True(t, os.IsNotExist(statErr), "malformed chain-id must fail before writing localnet files")
}

func assertPrototypeGenesisProfile(
	t *testing.T,
	cdc codec.Codec,
	txConfig client.TxConfig,
	appState map[string]json.RawMessage,
	validatorCount int,
) {
	t.Helper()

	authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)
	require.Len(t, authGenState.Accounts, validatorCount)

	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
	require.NotEmpty(t, bankGenState.Supply.String())
	require.Contains(t, bankGenState.Supply.String(), appparams.BaseDenom)
	require.Len(t, bankGenState.Balances, validatorCount)

	var native banktypes.Metadata
	for _, metadata := range bankGenState.DenomMetadata {
		if metadata.Base == appparams.BaseDenom {
			native = metadata
			break
		}
	}
	requireNativeTokenMetadata(t, native)

	expectedAccountCoins := sdk.NewCoins(
		sdk.NewCoin(bootstrapTestAssetDenom, sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction)),
		sdk.NewCoin(appparams.BaseDenom, sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction)),
	).Sort()
	for _, balance := range bankGenState.Balances {
		_, err := sdk.AccAddressFromBech32(balance.Address)
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(balance.Address, l1app.SDKBech32AccountPrefix+"1"), balance.Address)
		require.Equal(t, expectedAccountCoins, balance.Coins)
	}

	stakingGenState := stakingtypes.GetGenesisStateFromAppState(cdc, appState)
	require.Equal(t, appparams.BaseDenom, stakingGenState.Params.BondDenom)

	var mintGenState minttypes.GenesisState
	cdc.MustUnmarshalJSON(appState[minttypes.ModuleName], &mintGenState)
	require.Equal(t, appparams.BaseDenom, mintGenState.Params.MintDenom)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DefaultTargetInflationBps), mintGenState.Minter.Inflation)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.MinInflationBps), mintGenState.Params.InflationMin)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.MaxInflationBps), mintGenState.Params.InflationMax)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DefaultTargetStakeBps), mintGenState.Params.GoalBonded)
	require.True(t, mintGenState.Params.MaxSupply.IsZero())

	var feesGenState feestypes.GenesisState
	cdc.MustUnmarshalJSON(appState[feestypes.ModuleName], &feesGenState)
	require.Equal(t, feestypes.DefaultGenesisState(), &feesGenState)

	genutilGenState := genutiltypes.GetGenesisStateFromAppState(cdc, appState)
	require.Len(t, genutilGenState.GenTxs, validatorCount)
	expectedSelfDelegation := sdk.NewCoin(appparams.BaseDenom, sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction))
	for _, genTx := range genutilGenState.GenTxs {
		tx, err := genutiltypes.ValidateAndGetGenTx(genTx, txConfig.TxJSONDecoder(), genutiltypes.DefaultMessageValidator)
		require.NoError(t, err)
		msgs := tx.GetMsgs()
		require.Len(t, msgs, 1)
		createValMsg, ok := msgs[0].(*stakingtypes.MsgCreateValidator)
		require.True(t, ok)
		require.Equal(t, expectedSelfDelegation, createValMsg.Value)
		require.True(t, createValMsg.MinSelfDelegation.IsPositive())
		require.True(t, strings.HasPrefix(createValMsg.ValidatorAddress, l1app.ValidatorAddressPrefix), createValMsg.ValidatorAddress)
		require.NotRegexp(t, `^[a-z]+1`, createValMsg.ValidatorAddress)
	}
}

func requireNativeTokenMetadata(t *testing.T, native banktypes.Metadata) {
	t.Helper()

	require.NoError(t, native.Validate())
	require.Equal(t, appparams.BaseDenom, native.Base)
	require.Equal(t, appparams.DisplayDenom, native.Display)
	require.Equal(t, appparams.TokenName, native.Name)
	require.Equal(t, appparams.TokenSymbol, native.Symbol)
	requireDenomUnit(t, native, appparams.BaseDenom, 0)
	requireDenomUnit(t, native, appparams.DisplayDenom, appparams.DisplayDenomExponent)
}

func requireDenomUnit(t *testing.T, metadata banktypes.Metadata, denom string, exponent uint32) {
	t.Helper()

	for _, unit := range metadata.DenomUnits {
		if unit.Denom == denom {
			require.Equal(t, exponent, unit.Exponent)
			return
		}
	}
	require.Failf(t, "missing denom unit", "denom %s", denom)
}

func TestTestnetStartDoesNotPrintMnemonicByDefault(t *testing.T) {
	cmd := testnetStartCmd()
	printMnemonic, err := cmd.Flags().GetBool(flagPrintMnemonic)
	require.NoError(t, err)
	require.False(t, printMnemonic)
	require.Equal(t, "false", cmd.Flags().Lookup(flagPrintMnemonic).DefValue)
}
