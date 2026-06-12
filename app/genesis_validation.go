package app

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"

	"github.com/sovereign-l1/l1/app/genesisvalidation"
	appparams "github.com/sovereign-l1/l1/app/params"
)

func (app *L1App) validateAetraGenesis(genesisState GenesisState) error {
	return genesisvalidation.ValidateAetraGenesis(app.appCodec, genesisvalidation.State(genesisState), BondDenom)
}

func (app *L1App) validateAetraAuthGenesis(genesisState GenesisState) error {
	return genesisvalidation.ValidateAuthGenesis(app.appCodec, genesisvalidation.State(genesisState))
}

func (app *L1App) ensureCoreGenesisCollections(ctx sdk.Context) error {
	if err := ensureCollectionItem(ctx, app.MintKeeper.Params, appparams.AetraMintParams()); err != nil {
		return err
	}
	if err := ensureCollectionItem(ctx, app.MintKeeper.Minter, appparams.AetraInitialMinter()); err != nil {
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
	return genesisvalidation.EnsureCollectionItem(ctx, item, defaultValue)
}
