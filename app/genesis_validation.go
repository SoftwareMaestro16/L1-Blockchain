package app

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

func (app *L1App) validateAetraGenesis(genesisState GenesisState) error {
	if err := app.validateAetraAuthGenesis(genesisState); err != nil {
		return err
	}
	if err := app.validateAetraBankGenesis(genesisState); err != nil {
		return err
	}
	if err := app.validateAetraStakingGenesis(genesisState); err != nil {
		return err
	}
	if err := app.validateAetraMintGenesis(genesisState); err != nil {
		return err
	}
	return app.validateAetraFeeGenesis(genesisState)
}

func (app *L1App) validateAetraAuthGenesis(genesisState GenesisState) error {
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
			return fmt.Errorf("duplicate auth genesis account: %s", aetherisaddress.FormatAccAddress(addr))
		}
		seenAccounts[addrText] = struct{}{}
		if aetherisaddress.IsZeroAccAddress(addr) {
			return fmt.Errorf("auth genesis account %s must not be zero address", aetherisaddress.ZeroRawAddress)
		}
	}
	return nil
}

func (app *L1App) validateAetraBankGenesis(genesisState GenesisState) error {
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
	for _, balance := range bankGenesis.Balances {
		addr, err := aetherisaddress.ParseAccAddress(balance.Address)
		if err != nil {
			return fmt.Errorf("invalid bank balance address %s: %w", balance.Address, err)
		}
		if aetherisaddress.IsZeroAccAddress(addr) {
			return fmt.Errorf("bank balance address %s must not be zero address", aetherisaddress.ZeroRawAddress)
		}
	}
	return nil
}

func (app *L1App) validateAetraStakingGenesis(genesisState GenesisState) error {
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

func (app *L1App) validateAetraMintGenesis(genesisState GenesisState) error {
	var mintGenesis minttypes.GenesisState
	if genesisState[minttypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", minttypes.ModuleName)
	}
	if err := app.appCodec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", minttypes.ModuleName, err)
	}
	if err := minttypes.ValidateGenesis(mintGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", minttypes.ModuleName, err)
	}
	expected := appparams.AetraMintParams()
	if mintGenesis.Params.MintDenom != expected.MintDenom {
		return fmt.Errorf("invalid mint denom: expected %s, got %s", expected.MintDenom, mintGenesis.Params.MintDenom)
	}
	if !mintGenesis.Params.InflationRateChange.Equal(expected.InflationRateChange) {
		return fmt.Errorf("invalid mint inflation rate change: expected %s, got %s", expected.InflationRateChange, mintGenesis.Params.InflationRateChange)
	}
	if !mintGenesis.Params.InflationMin.Equal(expected.InflationMin) {
		return fmt.Errorf("invalid mint min inflation: expected %s, got %s", expected.InflationMin, mintGenesis.Params.InflationMin)
	}
	if !mintGenesis.Params.InflationMax.Equal(expected.InflationMax) {
		return fmt.Errorf("invalid mint max inflation: expected %s, got %s", expected.InflationMax, mintGenesis.Params.InflationMax)
	}
	if !mintGenesis.Params.GoalBonded.Equal(expected.GoalBonded) {
		return fmt.Errorf("invalid mint goal bonded: expected %s, got %s", expected.GoalBonded, mintGenesis.Params.GoalBonded)
	}
	if !mintGenesis.Params.MaxSupply.Equal(expected.MaxSupply) {
		return fmt.Errorf("invalid mint max supply: expected %s, got %s", expected.MaxSupply, mintGenesis.Params.MaxSupply)
	}
	if mintGenesis.Minter.Inflation.LT(expected.InflationMin) || mintGenesis.Minter.Inflation.GT(expected.InflationMax) {
		return fmt.Errorf("invalid mint current inflation: expected within %s..%s, got %s", expected.InflationMin, expected.InflationMax, mintGenesis.Minter.Inflation)
	}
	return nil
}

func (app *L1App) validateAetraFeeGenesis(genesisState GenesisState) error {
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
	if err := appparams.ValidateNativeFeeDenomsV1(feesGenesis.Params.AllowedFeeDenoms, feestypes.MaxAllowedFeeDenomsV1); err != nil {
		return err
	}
	return nil
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
	if _, err := item.Get(ctx); err == nil {
		return nil
	} else if !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	return item.Set(ctx, defaultValue)
}
