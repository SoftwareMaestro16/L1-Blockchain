package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/observability"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const fixtureTestAssetDenom = "testtoken"

func TestAetherisChainConstants(t *testing.T) {
	require.Equal(t, "Aetheris", appName)
	require.Equal(t, "ae", AccountAddressPrefix)
	require.Equal(t, "aevaloper", ValidatorAddressPrefix)
	require.Equal(t, "aevalcons", ConsensusAddressPrefix)
	require.Equal(t, appparams.BaseDenom, BondDenom)
	require.Equal(t, appparams.BaseDenom, sdk.DefaultBondDenom)
	require.Equal(t, int64(1_000_000_000), appparams.BaseUnitsPerDisplay)
	require.True(t, strings.HasSuffix(DefaultNodeHome, ".aetheris"), DefaultNodeHome)
}

func TestDefaultGenesisIncludesNativeTokenMetadata(t *testing.T) {
	app, genesis := setup(true, 5)

	var bankGenState banktypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)

	var native banktypes.Metadata
	for _, metadata := range bankGenState.DenomMetadata {
		if metadata.Base == appparams.BaseDenom {
			native = metadata
			break
		}
	}

	requireNativeTokenMetadata(t, native)
}

func requireNativeTokenMetadata(t *testing.T, native banktypes.Metadata) {
	t.Helper()

	require.NoError(t, native.Validate())
	require.Equal(t, appparams.BaseDenom, native.Base)
	require.Equal(t, appparams.DisplayDenom, native.Display)
	require.Equal(t, appparams.TokenSymbol, native.Symbol)
	require.Equal(t, appparams.TokenName, native.Name)
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

func TestDefaultGenesisValidatesAndSetsCustomModuleDefaults(t *testing.T) {
	app, genesis := setup(true, 5)

	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, []string{appparams.BaseDenom}, feesGenState.Params.AllowedFeeDenoms)
	require.Equal(t, "0.98", feesGenState.Params.ValidatorRewardsRatio)
	require.Equal(t, "0.02", feesGenState.Params.CommunityPoolRatio)

	stakingGenState := stakingtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
	require.Equal(t, appparams.BaseDenom, stakingGenState.Params.BondDenom)

	var mintGenState minttypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenState)
	require.Equal(t, appparams.BaseDenom, mintGenState.Params.MintDenom)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DefaultTargetInflationBps), mintGenState.Minter.Inflation)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DefaultResponsivenessBps), mintGenState.Params.InflationRateChange)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.MinInflationBps), mintGenState.Params.InflationMin)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.MaxInflationBps), mintGenState.Params.InflationMax)
	require.Equal(t, appparams.BpsToLegacyDec(appparams.DefaultTargetStakeBps), mintGenState.Params.GoalBonded)
	require.True(t, mintGenState.Params.MaxSupply.IsZero())

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)

	var loadGenState loadkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[loadtypes.ModuleName], &loadGenState))
	require.False(t, loadGenState.Params.Enabled)
	require.Empty(t, loadGenState.History)

	var routingGenState routingkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[routingtypes.ModuleName], &routingGenState))
	require.False(t, routingGenState.Params.Enabled)
	require.Empty(t, routingGenState.Shards)

	var zonesGenState zoneskeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[zonestypes.ModuleName], &zonesGenState))
	require.False(t, zonesGenState.Params.Enabled)
	require.Empty(t, zonesGenState.State.Zones)

	var meshGenState meshkeeper.GenesisState
	require.NoError(t, json.Unmarshal(genesis[meshtypes.ModuleName], &meshGenState))
	require.False(t, meshGenState.Params.Enabled)
	require.Empty(t, meshGenState.State.Destinations)
}

func TestNativeTokenRuntimeMetadataSendAndSupply(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	metadata, found := app.BankKeeper.GetDenomMetaData(ctx, appparams.BaseDenom)
	require.True(t, found)
	requireNativeTokenMetadata(t, metadata)

	addrs := AddTestAddrsIncremental(app, ctx, 2, sdkmath.NewInt(1_000_000))
	sender, recipient := addrs[0], addrs[1]
	beforeSupply := app.BankKeeper.GetSupply(ctx, appparams.BaseDenom)
	sendAmount := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 123_456))

	require.NoError(t, app.BankKeeper.SendCoins(ctx, sender, recipient, sendAmount))

	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 876_544), app.BankKeeper.GetBalance(ctx, sender, appparams.BaseDenom))
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 1_123_456), app.BankKeeper.GetBalance(ctx, recipient, appparams.BaseDenom))
	require.Equal(t, beforeSupply, app.BankKeeper.GetSupply(ctx, appparams.BaseDenom))
}

func TestCustomModuleGenesisInitExportRoundTrip(t *testing.T) {
	app := Setup(t, false)

	exported, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	var exportedGenesis GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &exportedGenesis))
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))

	var feesGenState feestypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[feestypes.ModuleName], &feesGenState)
	require.Equal(t, feestypes.DefaultGenesisState(), &feesGenState)

	var tokenfactoryGenState tokenfactorytypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[tokenfactorytypes.ModuleName], &tokenfactoryGenState)
	require.Empty(t, tokenfactoryGenState.Denoms)
	require.NoError(t, tokenfactoryGenState.Validate())

	var dexGenState dextypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(exportedGenesis[dextypes.ModuleName], &dexGenState)
	require.Equal(t, dextypes.DefaultNextPoolID, dexGenState.NextPoolId)
	require.Empty(t, dexGenState.Pools)
	require.NoError(t, dexGenState.Validate())

	var loadGenState loadkeeper.GenesisState
	require.NoError(t, json.Unmarshal(exportedGenesis[loadtypes.ModuleName], &loadGenState))
	require.NoError(t, loadGenState.Validate())
	require.False(t, loadGenState.Params.Enabled)

	var routingGenState routingkeeper.GenesisState
	require.NoError(t, json.Unmarshal(exportedGenesis[routingtypes.ModuleName], &routingGenState))
	require.NoError(t, routingGenState.Validate())
	require.False(t, routingGenState.Params.Enabled)

	var zonesGenState zoneskeeper.GenesisState
	require.NoError(t, json.Unmarshal(exportedGenesis[zonestypes.ModuleName], &zonesGenState))
	require.NoError(t, zonesGenState.Validate())
	require.False(t, zonesGenState.Params.Enabled)

	var meshGenState meshkeeper.GenesisState
	require.NoError(t, json.Unmarshal(exportedGenesis[meshtypes.ModuleName], &meshGenState))
	require.NoError(t, meshGenState.Validate())
	require.False(t, meshGenState.Params.Enabled)
}

func TestDefaultGenesisRejectsCorruptedPrototypeModuleState(t *testing.T) {
	app, baseGenesis := setup(true, 5)
	cdc := app.AppCodec()
	admin := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()

	tests := map[string]func(GenesisState){
		"invalid native metadata": func(genesis GenesisState) {
			var bankGenState banktypes.GenesisState
			cdc.MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)
			bankGenState.DenomMetadata = []banktypes.Metadata{{Base: "bad denom", Display: appparams.DisplayDenom}}
			genesis[banktypes.ModuleName] = cdc.MustMarshalJSON(&bankGenState)
		},
		"invalid staking denom": func(genesis GenesisState) {
			stakingGenState := stakingtypes.GetGenesisStateFromAppState(cdc, genesis)
			stakingGenState.Params.BondDenom = "bad denom"
			genesis[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenState)
		},
		"invalid fees params": func(genesis GenesisState) {
			var feesGenState feestypes.GenesisState
			cdc.MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenState)
			feesGenState.Params.AllowedFeeDenoms = []string{fixtureTestAssetDenom}
			genesis[feestypes.ModuleName] = cdc.MustMarshalJSON(&feesGenState)
		},
		"invalid tokenfactory metadata": func(genesis GenesisState) {
			tokenfactoryGenState := tokenfactorytypes.GenesisState{Denoms: []tokenfactorytypes.DenomAuthorityMetadata{{
				Denom: "factory/" + admin + "/" + appparams.BaseDenom,
				Admin: admin,
			}}}
			genesis[tokenfactorytypes.ModuleName] = cdc.MustMarshalJSON(&tokenfactoryGenState)
		},
		"duplicate dex pool pair": func(genesis GenesisState) {
			dexGenState := dextypes.GenesisState{NextPoolId: 3, Pools: []dextypes.Pool{
				{Id: 1, Denom0: "aaa", Denom1: appparams.BaseDenom, Reserve0: "1", Reserve1: "1", TotalShares: "1", LpDenom: "lp/1"},
				{Id: 2, Denom0: "aaa", Denom1: appparams.BaseDenom, Reserve0: "1", Reserve1: "1", TotalShares: "1", LpDenom: "lp/2"},
			}}
			genesis[dextypes.ModuleName] = cdc.MustMarshalJSON(&dexGenState)
		},
		"invalid load history ordering": func(genesis GenesisState) {
			loadGenState := loadkeeper.DefaultGenesis()
			loadGenState.History = []loadtypes.Result{
				{EMA: loadtypes.EMAState{WindowHeight: 2}},
				{EMA: loadtypes.EMAState{WindowHeight: 1}},
			}
			raw, err := json.Marshal(loadGenState)
			require.NoError(t, err)
			genesis[loadtypes.ModuleName] = raw
		},
		"duplicate routing shard config": func(genesis GenesisState) {
			routingGenState := routingkeeper.DefaultGenesis()
			routingGenState.Shards = []routingkeeper.ShardConfig{
				{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 1},
				{ZoneID: routingtypes.ZoneFinancial, ActiveShards: 2},
			}
			raw, err := json.Marshal(routingGenState)
			require.NoError(t, err)
			genesis[routingtypes.ModuleName] = raw
		},
		"duplicate zone id": func(genesis GenesisState) {
			zonesGenState := zoneskeeper.DefaultGenesis()
			zone := zonestypes.Zone{ID: zonestypes.ZoneIDFinancial}
			zonesGenState.State.Zones = []zonestypes.Zone{zone, zone}
			raw, err := json.Marshal(zonesGenState)
			require.NoError(t, err)
			genesis[zonestypes.ModuleName] = raw
		},
		"duplicate mesh destination": func(genesis GenesisState) {
			meshGenState := meshkeeper.DefaultGenesis()
			destination := meshtypes.MeshDestination{ZoneID: "FINANCIAL_ZONE", ShardID: "0:0", Active: true}
			meshGenState.State.Destinations = []meshtypes.MeshDestination{destination, destination}
			raw, err := json.Marshal(meshGenState)
			require.NoError(t, err)
			genesis[meshtypes.ModuleName] = raw
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			genesis := cloneGenesisState(baseGenesis)
			mutate(genesis)
			err := app.BasicModuleManager.ValidateGenesis(cdc, app.TxConfig(), genesis)
			require.Error(t, err)
		})
	}
}

func TestInitChainRejectsZeroGenesisAccount(t *testing.T) {
	app, genesis := setup(true, 5)
	zeroAccount := authtypes.NewBaseAccount(sdk.AccAddress(bytes.Repeat([]byte{0}, 20)), nil, 0, 0)
	zeroAny, err := codectypes.NewAnyWithValue(zeroAccount)
	require.NoError(t, err)

	authGenesis := authtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
	authGenesis.Accounts = append(authGenesis.Accounts, zeroAny)
	genesis[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&authGenesis)

	err = app.validateAetherisAuthGenesis(genesis)
	require.ErrorContains(t, err, aetherisaddress.ZeroRawAddress)

	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.ErrorContains(t, err, "must not be zero address")
}

func TestGenesisRejectsDuplicateAndMalformedAccounts(t *testing.T) {
	tests := map[string]struct {
		mutate   func(*L1App, GenesisState)
		errMatch string
	}{
		"duplicate auth account": {
			mutate: func(app *L1App, genesis GenesisState) {
				authGenesis := authtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				require.NotEmpty(t, authGenesis.Accounts)
				authGenesis.Accounts = append(authGenesis.Accounts, authGenesis.Accounts[0])
				genesis[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&authGenesis)
			},
			errMatch: "duplicate account",
		},
		"malformed auth account any": {
			mutate: func(app *L1App, genesis GenesisState) {
				var authRaw map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(genesis[authtypes.ModuleName], &authRaw))
				authRaw["accounts"] = json.RawMessage(`[{"@type":"/aetheris.malformed.GenesisAccount"}]`)
				raw, err := json.Marshal(authRaw)
				require.NoError(t, err)
				genesis[authtypes.ModuleName] = raw
			},
			errMatch: "unable to resolve type URL",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			app, genesis := setup(true, 5)
			genesis = GenesisStateWithSingleValidator(t, app)
			tc.mutate(app, genesis)
			requireGenesisValidationError(t, app, genesis, tc.errMatch)
		})
	}
}

func TestGenesisRejectsInvalidCoreBankAndStakingState(t *testing.T) {
	tests := map[string]struct {
		mutate   func(*L1App, GenesisState)
		errMatch string
	}{
		"duplicate bank balance": {
			mutate: func(app *L1App, genesis GenesisState) {
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				require.NotEmpty(t, bankGenesis.Balances)
				bankGenesis.Balances = append(bankGenesis.Balances, bankGenesis.Balances[0])
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch: "duplicate balance",
		},
		"malformed bank balance address": {
			mutate: func(app *L1App, genesis GenesisState) {
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
					Address: "not-an-aetheris-address",
					Coins:   sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1)),
				})
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch: "decoding bech32 failed",
		},
		"zero bank balance address": {
			mutate: func(app *L1App, genesis GenesisState) {
				zeroBech32, err := sdk.Bech32ifyAddressBytes(AccountAddressPrefix, bytes.Repeat([]byte{0}, 20))
				require.NoError(t, err)
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
					Address: zeroBech32,
					Coins:   sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1)),
				})
				bankGenesis.Supply = bankGenesis.Supply.Add(sdk.NewInt64Coin(appparams.BaseDenom, 1))
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch: "must not be zero address",
		},
		"bank supply mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Supply = bankGenesis.Supply.Add(sdk.NewInt64Coin(appparams.BaseDenom, 1))
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch: "genesis supply is incorrect",
		},
		"staking denom mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				stakingGenesis := stakingtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				stakingGenesis.Params.BondDenom = fixtureTestAssetDenom
				genesis[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)
			},
			errMatch: "invalid staking denom",
		},
		"mint denom mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				var mintGenesis minttypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis)
				mintGenesis.Params.MintDenom = fixtureTestAssetDenom
				genesis[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(&mintGenesis)
			},
			errMatch: "invalid mint denom",
		},
		"mint inflation max mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				var mintGenesis minttypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis)
				mintGenesis.Params.InflationMax = minttypes.DefaultParams().InflationMax
				genesis[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(&mintGenesis)
			},
			errMatch: "invalid mint max inflation",
		},
		"mint current inflation outside bounds": {
			mutate: func(app *L1App, genesis GenesisState) {
				var mintGenesis minttypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis)
				mintGenesis.Minter.Inflation = minttypes.DefaultInitialMinter().Inflation
				genesis[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(&mintGenesis)
			},
			errMatch: "invalid mint current inflation",
		},
		"fees denom mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				var feesGenesis feestypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenesis)
				feesGenesis.Params.AllowedFeeDenoms = []string{fixtureTestAssetDenom}
				genesis[feestypes.ModuleName] = app.AppCodec().MustMarshalJSON(&feesGenesis)
			},
			errMatch: "v1 only accepts fee denom naet",
		},
		"dex reserve module balance mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				dexGenesis := dextypes.GenesisState{
					NextPoolId: 2,
					Params:     dextypes.DefaultParams(),
					Pools: []dextypes.Pool{{
						Id:          1,
						Denom0:      appparams.BaseDenom,
						Denom1:      "uatom",
						Reserve0:    "100",
						Reserve1:    "200",
						TotalShares: "100",
						LpDenom:     "lp/1",
					}},
				}
				genesis[dextypes.ModuleName] = app.AppCodec().MustMarshalJSON(&dexGenesis)
			},
			errMatch: "dex genesis reserve mismatch",
		},
		"dex LP supply mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				dexGenesis := dextypes.GenesisState{
					NextPoolId: 2,
					Params:     dextypes.DefaultParams(),
					Pools: []dextypes.Pool{{
						Id:          1,
						Denom0:      appparams.BaseDenom,
						Denom1:      "uatom",
						Reserve0:    "100",
						Reserve1:    "200",
						TotalShares: "100",
						LpDenom:     "lp/1",
					}},
				}
				genesis[dextypes.ModuleName] = app.AppCodec().MustMarshalJSON(&dexGenesis)

				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
					Address: authtypes.NewModuleAddress(dextypes.ModuleName).String(),
					Coins: sdk.NewCoins(
						sdk.NewInt64Coin(appparams.BaseDenom, 100),
						sdk.NewInt64Coin("uatom", 200),
					),
				})
				bankGenesis.Supply = bankGenesis.Supply.Add(
					sdk.NewInt64Coin(appparams.BaseDenom, 100),
					sdk.NewInt64Coin("uatom", 200),
				)
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch: "dex genesis LP supply mismatch",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			app, genesis := setup(true, 5)
			genesis = GenesisStateWithSingleValidator(t, app)
			tc.mutate(app, genesis)
			requireGenesisValidationError(t, app, genesis, tc.errMatch)
		})
	}
}

func TestDefaultGenesisInitExportValidateAcceptanceChain(t *testing.T) {
	app, genesis := setup(true, 5)
	genesis = GenesisStateWithSingleValidator(t, app)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)

	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	exportedA, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	exportedB, err := app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	require.Equal(t, exportedA.AppState, exportedB.AppState)

	var exportedGenesis GenesisState
	require.NoError(t, json.Unmarshal(exportedA.AppState, &exportedGenesis))
	require.NoError(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), exportedGenesis))
}

func TestPrototypeModuleAccountPermissionsAreNarrow(t *testing.T) {
	expected := map[string][]string{
		authtypes.FeeCollectorName:                     nil,
		distrtypes.ModuleName:                          nil,
		minttypes.ModuleName:                           {authtypes.Minter},
		stakingtypes.BondedPoolName:                    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:                 {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                            {authtypes.Burner},
		protocolpooltypes.ModuleName:                   nil,
		protocolpooltypes.ProtocolPoolEscrowAccount:    nil,
		tokenfactorytypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		dextypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
		burntypes.ModuleName:                           {authtypes.Burner},
		feecollectortypes.CollectorModuleName:          {authtypes.Burner},
		feecollectortypes.TreasuryModuleName:           nil,
		feecollectortypes.ProtectionModuleName:         nil,
		feecollectortypes.ValidatorInsuranceModuleName: nil,
		feecollectortypes.EcosystemGrantsModuleName:    nil,
		feecollectortypes.StorageRentReserveModuleName: nil,
		feecollectortypes.BurnModuleName:               nil,
		feecollectortypes.ReporterRewardsModuleName:    nil,
		feestypes.ModuleName:                           nil,
	}
	require.Equal(t, expected, GetMaccPerms())

	blocked := BlockedAddresses()
	for moduleName := range expected {
		addr := authtypes.NewModuleAddress(moduleName).String()
		if moduleName == govtypes.ModuleName {
			require.False(t, blocked[addr])
			continue
		}
		require.True(t, blocked[addr], moduleName)
	}
}

func cloneGenesisState(genesis GenesisState) GenesisState {
	clone := make(GenesisState, len(genesis))
	for moduleName, raw := range genesis {
		clone[moduleName] = append(json.RawMessage(nil), raw...)
	}
	return clone
}

func requireGenesisValidationError(t *testing.T, app *L1App, genesis GenesisState, errMatch string) {
	t.Helper()
	appPolicyErr := app.validateAetherisGenesis(genesis)
	moduleErr := app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis)
	require.True(t, appPolicyErr != nil || moduleErr != nil, "expected app policy or module genesis validation error")
	matched := appPolicyErr != nil && strings.Contains(appPolicyErr.Error(), errMatch)
	matched = matched || moduleErr != nil && strings.Contains(moduleErr.Error(), errMatch)
	require.Truef(
		t,
		matched,
		"expected error containing %q, got app policy error %v and module error %v",
		errMatch,
		appPolicyErr,
		moduleErr,
	)

	stateBytes, marshalErr := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, marshalErr)
	_, err := app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.Error(t, err)
}

func TestAppGenesisExportImportRoundTripAndDeterminism(t *testing.T) {
	source, _ := setup(true, 5)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = source.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = source.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   source.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = source.Commit()
	require.NoError(t, err)

	exportedA, err := source.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	exportedB, err := source.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	require.Equal(t, exportedA.AppState, exportedB.AppState)

	var exportedState GenesisState
	require.NoError(t, json.Unmarshal(exportedA.AppState, &exportedState))
	require.NoError(t, source.BasicModuleManager.ValidateGenesis(source.AppCodec(), source.TxConfig(), exportedState))

	target := NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome},
	)
	_, err = target.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: &exportedA.ConsensusParams,
		AppStateBytes:   exportedA.AppState,
	})
	require.NoError(t, err)
	_, err = target.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   target.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = target.Commit()
	require.NoError(t, err)

	reexported, err := target.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	var reexportedState GenesisState
	require.NoError(t, json.Unmarshal(reexported.AppState, &reexportedState))
	require.NoError(t, target.BasicModuleManager.ValidateGenesis(target.AppCodec(), target.TxConfig(), reexportedState))
}

func TestTelemetryDoesNotChangeAppHash(t *testing.T) {
	source, _ := setup(true, 5)
	genesis := GenesisStateWithSingleValidator(t, source)
	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	t.Cleanup(func() {
		observability.ResetForTesting()
	})

	enabledHash := runSingleBlockForTelemetryTest(t, stateBytes, true)
	disabledHash := runSingleBlockForTelemetryTest(t, stateBytes, false)

	require.Equal(t, enabledHash, disabledHash)
}

func TestCustomModuleMigrationsFromV1ToCurrent(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	fromVM := app.ModuleManager.GetVersionMap()
	fromVM[feestypes.ModuleName] = 1
	fromVM[tokenfactorytypes.ModuleName] = 1
	fromVM[dextypes.ModuleName] = 1
	fromVM[loadtypes.ModuleName] = 1
	fromVM[routingtypes.ModuleName] = 1
	fromVM[zonestypes.ModuleName] = 1
	fromVM[meshtypes.ModuleName] = 1

	updated, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
	require.NoError(t, err)
	require.Equal(t, uint64(2), updated[feestypes.ModuleName])
	require.Equal(t, uint64(2), updated[tokenfactorytypes.ModuleName])
	require.Equal(t, uint64(2), updated[dextypes.ModuleName])
	require.Equal(t, uint64(2), updated[loadtypes.ModuleName])
	require.Equal(t, uint64(2), updated[routingtypes.ModuleName])
	require.Equal(t, uint64(2), updated[zonestypes.ModuleName])
	require.Equal(t, uint64(2), updated[meshtypes.ModuleName])
}

func runSingleBlockForTelemetryTest(t *testing.T, stateBytes []byte, telemetryEnabled bool) []byte {
	t.Helper()
	observability.ResetForTesting()
	observability.SetEnabled(telemetryEnabled)
	app := NewL1App(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		true,
		sims.AppOptionsMap{flags.FlagHome: DefaultNodeHome},
	)
	_, err := app.InitChain(&abci.RequestInitChain{
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: sims.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Hash:   app.LastCommitID().Hash,
		Time:   time.Unix(1_700_000_000, 0).UTC(),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)
	return app.LastCommitID().Hash
}
