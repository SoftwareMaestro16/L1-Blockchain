package app

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	loadkeeper "github.com/sovereign-l1/l1/x/load/keeper"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshkeeper "github.com/sovereign-l1/l1/x/mesh/keeper"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	routingkeeper "github.com/sovereign-l1/l1/x/routing/keeper"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	zoneskeeper "github.com/sovereign-l1/l1/x/zones/keeper"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestDefaultGenesisRejectsCorruptedPrototypeModuleState(t *testing.T) {
	app, baseGenesis := setup(true, 5)
	cdc := app.AppCodec()

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

	err = app.validateAetraAuthGenesis(genesis)
	require.ErrorContains(t, err, aetraaddress.ZeroRawAddress)

	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(t, err)
	_, err = app.InitChain(&abci.RequestInitChain{
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.ErrorContains(t, err, "must not be zero address")
}

func TestGenesisRejectsDuplicateAndMalformedAccounts(t *testing.T) {
	tests := map[string]struct {
		mutate		func(*L1App, GenesisState)
		errMatch	string
	}{
		"duplicate auth account": {
			mutate: func(app *L1App, genesis GenesisState) {
				authGenesis := authtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				require.NotEmpty(t, authGenesis.Accounts)
				authGenesis.Accounts = append(authGenesis.Accounts, authGenesis.Accounts[0])
				genesis[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&authGenesis)
			},
			errMatch:	"duplicate account",
		},
		"malformed auth account any": {
			mutate: func(app *L1App, genesis GenesisState) {
				var authRaw map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(genesis[authtypes.ModuleName], &authRaw))
				authRaw["accounts"] = json.RawMessage(`[{"@type":"/aetra.malformed.GenesisAccount"}]`)
				raw, err := json.Marshal(authRaw)
				require.NoError(t, err)
				genesis[authtypes.ModuleName] = raw
			},
			errMatch:	"unable to resolve type URL",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			app, _ := setup(true, 5)
			genesis := GenesisStateWithSingleValidator(t, app)
			tc.mutate(app, genesis)
			requireGenesisValidationError(t, app, genesis, tc.errMatch)
		})
	}
}

func TestGenesisRejectsInvalidCoreBankAndStakingState(t *testing.T) {
	tests := map[string]struct {
		mutate		func(*L1App, GenesisState)
		errMatch	string
	}{
		"duplicate bank balance": {
			mutate: func(app *L1App, genesis GenesisState) {
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				require.NotEmpty(t, bankGenesis.Balances)
				bankGenesis.Balances = append(bankGenesis.Balances, bankGenesis.Balances[0])
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch:	"duplicate balance",
		},
		"malformed bank balance address": {
			mutate: func(app *L1App, genesis GenesisState) {
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
					Address:	"not-an-aetra-address",
					Coins:		sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1)),
				})
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch:	"decoding bech32 failed",
		},
		"zero bank balance address": {
			mutate: func(app *L1App, genesis GenesisState) {
				zeroBech32, err := sdk.Bech32ifyAddressBytes(SDKBech32AccountPrefix, bytes.Repeat([]byte{0}, 20))
				require.NoError(t, err)
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
					Address:	zeroBech32,
					Coins:		sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1)),
				})
				bankGenesis.Supply = bankGenesis.Supply.Add(sdk.NewInt64Coin(appparams.BaseDenom, 1))
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch:	"must not be zero address",
		},
		"bank supply mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				bankGenesis := banktypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				bankGenesis.Supply = bankGenesis.Supply.Add(sdk.NewInt64Coin(appparams.BaseDenom, 1))
				genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)
			},
			errMatch:	"genesis supply is incorrect",
		},
		"staking denom mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				stakingGenesis := stakingtypes.GetGenesisStateFromAppState(app.AppCodec(), genesis)
				stakingGenesis.Params.BondDenom = fixtureTestAssetDenom
				genesis[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)
			},
			errMatch:	"invalid staking denom",
		},
		"mint denom mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				var mintGenesis minttypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis)
				mintGenesis.Params.MintDenom = fixtureTestAssetDenom
				genesis[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(&mintGenesis)
			},
			errMatch:	"invalid mint denom",
		},
		"mint inflation max mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				var mintGenesis minttypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis)
				mintGenesis.Params.InflationMax = minttypes.DefaultParams().InflationMax
				genesis[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(&mintGenesis)
			},
			errMatch:	"invalid mint max inflation",
		},
		"mint current inflation outside bounds": {
			mutate: func(app *L1App, genesis GenesisState) {
				var mintGenesis minttypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis)
				mintGenesis.Minter.Inflation = minttypes.DefaultInitialMinter().Inflation
				genesis[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(&mintGenesis)
			},
			errMatch:	"invalid mint current inflation",
		},
		"fees denom mismatch": {
			mutate: func(app *L1App, genesis GenesisState) {
				var feesGenesis feestypes.GenesisState
				app.AppCodec().MustUnmarshalJSON(genesis[feestypes.ModuleName], &feesGenesis)
				feesGenesis.Params.AllowedFeeDenoms = []string{fixtureTestAssetDenom}
				genesis[feestypes.ModuleName] = app.AppCodec().MustMarshalJSON(&feesGenesis)
			},
			errMatch:	"v1 only accepts fee denom naet",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			app, _ := setup(true, 5)
			genesis := GenesisStateWithSingleValidator(t, app)
			tc.mutate(app, genesis)
			requireGenesisValidationError(t, app, genesis, tc.errMatch)
		})
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
	appPolicyErr := app.validateAetraGenesis(genesis)
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
		Validators:		[]abci.ValidatorUpdate{},
		ConsensusParams:	sims.DefaultConsensusParams,
		AppStateBytes:		stateBytes,
	})
	require.Error(t, err)
}
