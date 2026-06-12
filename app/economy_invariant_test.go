package app

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/sovereign-l1/l1/app/params"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
)

func TestEmissionCapInvariantFailsOnExcessMinting(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	emParams := emissionstypes.DefaultParams()
	require.NoError(t, app.EmissionsKeeper.SetParams(ctx, emParams))

	maxMintable := emParams.AnnualReferenceSupply.Amount.Mul(sdkmath.NewInt(int64(emParams.ConstitutionalMaxInflationBps))).Quo(sdkmath.NewInt(10_000))
	excess := maxMintable.Add(sdkmath.NewInt(1))
	excessCoin := sdk.NewCoin(emParams.BaseDenom, excess)
	fakeGenesis := emissionstypes.DefaultGenesisState()
	fakeGenesis.TotalMintedAccounting = excessCoin
	fakeGenesis.EpochHistory = []emissionstypes.EmissionEpoch{
		{
			Epoch:			1,
			EmissionAmount:		excessCoin,
			ValidatorReward:	excessCoin,
			Treasury:		sdk.NewCoin(emParams.BaseDenom, sdkmath.ZeroInt()),
			ProtectionFund:		sdk.NewCoin(emParams.BaseDenom, sdkmath.ZeroInt()),
			Burn:			sdk.NewCoin(emParams.BaseDenom, sdkmath.ZeroInt()),
			Ecosystem:		sdk.NewCoin(emParams.BaseDenom, sdkmath.ZeroInt()),
			RoundingRemainder:	sdk.NewCoin(emParams.BaseDenom, sdkmath.ZeroInt()),
		},
	}
	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.EmissionsKeeper.InitGenesis(importedCtx, *fakeGenesis))
	err := importedApp.assertEmissionCapInvariant(importedCtx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds constitutional max")
}

func TestBurnAccountingInvariantFailsOnMismatch(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	configureGoldenBurnParams(t, app, ctx)

	require.NoError(t, app.assertBurnAccountingInvariant(ctx))

	gs, err := app.BurnKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	gs.BurnedByDenom = []burntypes.BurnedByDenomEntry{
		{Denom: params.BaseDenom, Amount: sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 999_999))},
	}
	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.BurnKeeper.InitGenesis(importedCtx, *gs))
	err = importedApp.assertBurnAccountingInvariant(importedCtx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "burn accounting mismatch")
}

func TestTreasuryAccountingInvariantFailsOnMismatch(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.assertTreasuryAccountingInvariant(ctx))

	gs, err := app.TreasuryKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	gs.Allocations = treasurytypes.TreasuryAllocations{
		ReserveBalance: sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 1)),
	}
	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.TreasuryKeeper.InitGenesis(importedCtx, *gs))
	err = importedApp.assertTreasuryAccountingInvariant(importedCtx)
	require.Error(t, err)
}

func TestFeeCollectorAccountingInvariantFailsOnMismatch(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.assertFeeCollectorAccountingInvariant(ctx))

	gs, err := app.FeeCollectorKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	gs.Balances = feecollectortypes.FeeBalances{
		GasFees:	sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 999)),
		TotalCollected:	sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 999)),
	}
	gs.PendingDistribution = feecollectortypes.PendingDistribution{
		Treasury: sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 999)),
	}
	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.FeeCollectorKeeper.InitGenesis(importedCtx, *gs))
	err = importedApp.assertFeeCollectorAccountingInvariant(importedCtx)
	require.Error(t, err)
}

func TestRentReserveBalanceInvariantFailsOnAlertState(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.assertRentReserveBalanceInvariant(ctx))

	gs, err := app.StorageRentKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)
	gs.State.SystemReserve = storagerenttypes.SystemRentReserve{
		AvailableFunds:				0,
		ProjectedRentPerBlock:			100,
		WarningRunwayBlocks:			100,
		CriticalRunwayBlocks:			10,
		RequiredTopUp:				1_000,
		FeeCollectorBalance:			0,
		TreasuryBalance:			0,
		GovernanceConfiguredPayerBalance:	0,
		ProtocolCriticalExecutable:		false,
	}
	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.StorageRentKeeper.InitGenesisState(importedCtx, gs))
	err = importedApp.assertRentReserveBalanceInvariant(importedCtx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "storage rent system reserve is in invariant alert state")
}

func TestFeesModuleGenesisRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	original, err := app.FeesKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, original.Validate())

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.FeesKeeper.InitGenesis(importedCtx, *original))

	reexported, err := importedApp.FeesKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, original, reexported)
}

func TestFeeCollectorModuleGenesisRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	original, err := app.FeeCollectorKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, original.Validate())

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.FeeCollectorKeeper.InitGenesis(importedCtx, *original))

	reexported, err := importedApp.FeeCollectorKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, original, reexported)
}

func TestEmissionsModuleGenesisRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	original, err := app.EmissionsKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, original.Validate())

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.EmissionsKeeper.InitGenesis(importedCtx, *original))

	reexported, err := importedApp.EmissionsKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, original, reexported)
}

func TestBurnModuleGenesisRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	original, err := app.BurnKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, original.Validate())

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.BurnKeeper.InitGenesis(importedCtx, *original))

	reexported, err := importedApp.BurnKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, original, reexported)
}

func TestTreasuryModuleGenesisRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	require.NoError(t, app.TreasuryKeeper.SyncIncomingFunds(ctx))

	original, err := app.TreasuryKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, original.Validate())

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	if err := initTreasuryBankBalance(importedCtx, importedApp, original); err != nil {
		require.NoError(t, err)
	}
	require.NoError(t, importedApp.TreasuryKeeper.InitGenesis(importedCtx, *original))

	reexported, err := importedApp.TreasuryKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, original, reexported)
}

func TestStorageRentModuleGenesisRoundTrip(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)

	original, err := app.StorageRentKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.NoError(t, original.Validate())

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	require.NoError(t, importedApp.StorageRentKeeper.InitGenesisState(importedCtx, original))

	reexported, err := importedApp.StorageRentKeeper.ExportGenesisState(importedCtx)
	require.NoError(t, err)
	require.Equal(t, original, reexported)
}

func TestGoldenEconomyFullCycleWithStepByStepInvariants(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(42)

	require.Empty(t, app.RunAppInvariants(ctx))

	fees := sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 20_000_000))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, fees))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, fees))
	require.NoError(t, app.FeesKeeper.RecordCollectedFees(ctx, fees))
	require.Empty(t, app.RunAppInvariants(ctx))

	_, err := app.EndBlocker(ctx)
	require.NoError(t, err)
	require.Empty(t, app.RunAppInvariants(ctx))

	wrongDenomFee := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 100))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, wrongDenomFee))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, wrongDenomFee))
	err = app.FeesKeeper.RecordCollectedFees(ctx, wrongDenomFee)
	require.Error(t, err)
	require.Empty(t, app.RunAppInvariants(ctx))

	zeroFee := sdk.NewCoins()
	require.NoError(t, app.FeesKeeper.RecordCollectedFees(ctx, zeroFee))
	require.Empty(t, app.RunAppInvariants(ctx))

	configureGoldenBurnParams(t, app, ctx)
	configureGoldenEmissionParams(t, app, ctx)
	ctx = ctx.WithBlockHeight(43)
	emission, err := app.FinalizeNativeEconomyEpoch(ctx, 7, 5_000)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(params.BaseDenom, 1_000), emission.EmissionAmount)
	require.Empty(t, app.RunAppInvariants(ctx))

	rentPayer := AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 1_000)))[0]
	allocations, remainder, err := app.FeeCollectorKeeper.CollectAndDistributeProtocolIncomeFromAccount(ctx, rentPayer, sdk.NewCoins(sdk.NewInt64Coin(params.BaseDenom, 100)))
	require.NoError(t, err)
	require.True(t, remainder.Empty())
	require.NotEmpty(t, allocations)
	require.Empty(t, app.RunAppInvariants(ctx))

	feeCollectorGenesis, err := app.FeeCollectorKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	burnGenesis, err := app.BurnKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	emissionsGenesis, err := app.EmissionsKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	treasuryGenesis, err := app.TreasuryKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	storageRentGenesis, err := app.StorageRentKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)

	importedApp := Setup(t, false)
	importedCtx := importedApp.NewContext(false)
	treasuryAccounting := treasuryGenesis.Allocations.AccountingBalance()
	if !treasuryAccounting.Empty() {
		require.NoError(t, importedApp.BankKeeper.MintCoins(importedCtx, minttypes.ModuleName, treasuryAccounting))
		require.NoError(t, importedApp.BankKeeper.SendCoinsFromModuleToModule(importedCtx, minttypes.ModuleName, treasurytypes.TreasuryModuleName, treasuryAccounting))
	}
	require.NoError(t, importedApp.FeeCollectorKeeper.InitGenesis(importedCtx, *feeCollectorGenesis))
	require.NoError(t, importedApp.BurnKeeper.InitGenesis(importedCtx, *burnGenesis))
	require.NoError(t, importedApp.EmissionsKeeper.InitGenesis(importedCtx, *emissionsGenesis))
	require.NoError(t, importedApp.TreasuryKeeper.InitGenesis(importedCtx, *treasuryGenesis))
	require.NoError(t, importedApp.StorageRentKeeper.InitGenesisState(importedCtx, storageRentGenesis))
	require.Empty(t, importedApp.RunAppInvariants(importedCtx))
}

func initTreasuryBankBalance(ctx sdk.Context, app *L1App, gs *treasurytypes.GenesisState) error {
	accounting := gs.Allocations.AccountingBalance()
	if !accounting.Empty() {
		if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, accounting); err != nil {
			return err
		}
		return app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, treasurytypes.TreasuryModuleName, accounting)
	}
	return nil
}
