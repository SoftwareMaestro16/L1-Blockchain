package app

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sovereign-l1/l1/app/exporting"
)

// ExportAppStateAndValidators exports the state of the application for a genesis file.
func (app *L1App) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs, modulesToExport []string) (servertypes.ExportedApp, error) {
	return exporting.ExportAppStateAndValidators(app.exportDependencies(), forZeroHeight, jailAllowedAddrs, modulesToExport)
}

func (app *L1App) exportDependencies() exporting.Dependencies {
	return exporting.Dependencies{
		AppCodec:		app.appCodec,
		ModuleManager:		app.ModuleManager,
		AccountKeeper:		app.AccountKeeper,
		StakingKeeper:		app.StakingKeeper,
		DistrKeeper:		app.DistrKeeper,
		SlashingKeeper:		app.SlashingKeeper,
		StakingStoreKey:	app.GetKey(stakingtypes.StoreKey),
		Logger:			app.Logger(),
		NewContext: func(header cmtproto.Header) sdk.Context {
			return app.NewUncachedContext(false, header)
		},
		LastBlockHeight:	app.LastBlockHeight,
		ConsensusParams:	app.GetConsensusParams,
		EnsureCollections:	app.ensureCoreGenesisCollections,
	}
}
