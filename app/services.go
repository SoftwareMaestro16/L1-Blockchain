package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/collections"
	"cosmossdk.io/core/appmodule"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	orbitaladdress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/observability"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

func (app *L1App) Name() string { return app.BaseApp.Name() }

func (app *L1App) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

func (app *L1App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

func (app *L1App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (app *L1App) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	start := time.Now()
	res, err := app.BaseApp.FinalizeBlock(req)
	// Wall-clock timing is process telemetry only; it does not read or write consensus state.
	observability.RecordFinalizeBlock(req.Height, req.Time, len(req.Txs), time.Since(start))
	if err != nil {
		observability.RecordModuleError("app", "finalize_block", "error")
	}
	return res, err
}

func (a *L1App) Configurator() module.Configurator {
	return a.configurator
}

func (app *L1App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	if err != nil {
		return nil, err
	}
	if err := app.validateOrbitalisGenesis(genesisState); err != nil {
		return nil, err
	}
	res, err := app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
	if err != nil {
		return nil, err
	}
	if err := app.ensureCoreGenesisCollections(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func (app *L1App) validateOrbitalisGenesis(genesisState GenesisState) error {
	if err := app.validateOrbitalisAuthGenesis(genesisState); err != nil {
		return err
	}
	if err := app.validateOrbitalisBankGenesis(genesisState); err != nil {
		return err
	}
	if err := app.validateOrbitalisStakingGenesis(genesisState); err != nil {
		return err
	}
	return app.validateOrbitalisFeeGenesis(genesisState)
}

func (app *L1App) validateOrbitalisAuthGenesis(genesisState GenesisState) error {
	var authGenesis authtypes.GenesisState
	if genesisState[authtypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", authtypes.ModuleName)
	}
	if err := app.appCodec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", authtypes.ModuleName, err)
	}
	accounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		return err
	}
	seenAccounts := make(map[string]struct{}, len(accounts))
	for _, account := range accounts {
		addr := account.GetAddress()
		addrText := addr.String()
		if _, found := seenAccounts[addrText]; found {
			return fmt.Errorf("duplicate auth genesis account: %s", orbitaladdress.FormatAccAddress(addr))
		}
		seenAccounts[addrText] = struct{}{}
		if orbitaladdress.IsZeroAccAddress(addr) {
			return fmt.Errorf("auth genesis account %s must not be zero address", orbitaladdress.ZeroRawAddress)
		}
	}
	return nil
}

func (app *L1App) validateOrbitalisBankGenesis(genesisState GenesisState) error {
	var bankGenesis banktypes.GenesisState
	if genesisState[banktypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", banktypes.ModuleName)
	}
	if err := app.appCodec.UnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", banktypes.ModuleName, err)
	}
	if err := bankGenesis.Validate(); err != nil {
		return err
	}
	return nil
}

func (app *L1App) validateOrbitalisStakingGenesis(genesisState GenesisState) error {
	var stakingGenesis stakingtypes.GenesisState
	if genesisState[stakingtypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", stakingtypes.ModuleName)
	}
	if err := app.appCodec.UnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", stakingtypes.ModuleName, err)
	}
	if stakingGenesis.Params.BondDenom != BondDenom {
		return fmt.Errorf("invalid staking denom: expected %s, got %s", BondDenom, stakingGenesis.Params.BondDenom)
	}
	return nil
}

func (app *L1App) validateOrbitalisFeeGenesis(genesisState GenesisState) error {
	var feesGenesis feestypes.GenesisState
	if genesisState[feestypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", feestypes.ModuleName)
	}
	if err := app.appCodec.UnmarshalJSON(genesisState[feestypes.ModuleName], &feesGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", feestypes.ModuleName, err)
	}
	if err := feesGenesis.Validate(); err != nil {
		return err
	}
	if len(feesGenesis.Params.AllowedFeeDenoms) != 1 || feesGenesis.Params.AllowedFeeDenoms[0] != BondDenom {
		return fmt.Errorf("allowed fee denom must be exactly %s", BondDenom)
	}
	return nil
}

func (app *L1App) ensureCoreGenesisCollections(ctx sdk.Context) error {
	if err := ensureCollectionItem(ctx, app.MintKeeper.Params, minttypes.DefaultParams()); err != nil {
		return err
	}
	if err := ensureCollectionItem(ctx, app.MintKeeper.Minter, minttypes.DefaultInitialMinter()); err != nil {
		return err
	}
	if err := ensureCollectionItem(ctx, app.DistrKeeper.Params, distrtypes.DefaultParams()); err != nil {
		return err
	}
	if err := ensureCollectionItem(ctx, app.DistrKeeper.FeePool, distrtypes.InitialFeePool()); err != nil {
		return err
	}
	if _, err := app.DistrKeeper.GetPreviousProposerConsAddr(ctx); err != nil {
		if err.Error() != "previous proposer not set" {
			return err
		}
		if err := app.DistrKeeper.SetPreviousProposerConsAddr(ctx, sdk.ConsAddress{}); err != nil {
			return err
		}
	}
	if err := ensureCollectionItem(ctx, app.GovKeeper.Params, govv1.DefaultParams()); err != nil {
		return err
	}
	if err := ensureCollectionItem(ctx, app.GovKeeper.Constitution, ""); err != nil {
		return err
	}
	proposalID, err := app.GovKeeper.ProposalID.Peek(ctx)
	if err != nil {
		return err
	}
	if proposalID == 0 {
		if err := app.GovKeeper.ProposalID.Set(ctx, govv1.DefaultStartingProposalID); err != nil {
			return err
		}
	}
	return ensureCollectionItem(ctx, app.ProtocolPoolKeeper.Params, protocolpooltypes.DefaultParams())
}

func ensureCollectionItem[T any](ctx context.Context, item collections.Item[T], defaultValue T) error {
	if _, err := item.Get(ctx); err == nil {
		return nil
	} else if !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	return item.Set(ctx, defaultValue)
}

func (app *L1App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *L1App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *L1App) AppCodec() codec.Codec {
	return app.appCodec
}

func (app *L1App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *L1App) TxConfig() client.TxConfig {
	return app.txConfig
}

func (app *L1App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.ModuleManager.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.ModuleManager.Modules),
		AddressCodec:          orbitaladdress.Codec{},
		ValidatorAddressCodec: orbitaladdress.Codec{},
		ConsensusAddressCodec: orbitaladdress.Codec{},
	}
}

func (a *L1App) DefaultGenesis() map[string]json.RawMessage {
	return withNativeTokenMetadata(a.appCodec, withCoreModuleGenesisDefaults(a.appCodec, a.BasicModuleManager.DefaultGenesis(a.appCodec)))
}

func (app *L1App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

func (app *L1App) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, key := range app.keys {
		keys = append(keys, key)
	}

	return keys
}

func (app *L1App) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *L1App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func (app *L1App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

func (app *L1App) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

func (app *L1App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg, func() int64 {
		return app.CommitMultiStore().EarliestVersion()
	})
}

func GetMaccPerms() map[string][]string {
	return maps.Clone(maccPerms)
}

func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
