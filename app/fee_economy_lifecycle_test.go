package app

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
)

func TestEndBlockerDistributesNativeTransactionFees(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(42)
	fees := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 20_000_000))

	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, fees))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, fees))
	supplyBefore := app.BankKeeper.GetSupply(ctx, appparams.BaseDenom)
	treasuryBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.TreasuryModuleName), appparams.BaseDenom)
	protectionBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.ProtectionModuleName), appparams.BaseDenom)

	require.NoError(t, app.FeesKeeper.RecordCollectedFees(ctx, fees))
	validatorsBeforeDistribution := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), appparams.BaseDenom)
	pending, err := app.FeeCollectorKeeper.GetPendingDistribution(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 10_000_000)), pending.Burn)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 3_000_000)), pending.Treasury)
	require.True(t, pending.Protection.Empty())
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 7_000_000)), pending.Validators)

	_, err = app.EndBlocker(ctx)
	require.NoError(t, err)

	pending, err = app.FeeCollectorKeeper.GetPendingDistribution(ctx)
	require.NoError(t, err)
	require.True(t, pending.Total().Empty())
	require.Equal(t, sdk.NewCoins(), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.CollectorModuleName)))
	require.Equal(t, treasuryBefore.Add(sdk.NewInt64Coin(appparams.BaseDenom, 3_000_000)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.TreasuryModuleName), appparams.BaseDenom))
	require.Equal(t, protectionBefore, app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.ProtectionModuleName), appparams.BaseDenom))
	require.Equal(t, validatorsBeforeDistribution.Add(sdk.NewInt64Coin(appparams.BaseDenom, 7_000_000)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), appparams.BaseDenom))
	require.Equal(t, supplyBefore.Amount.Sub(sdkmath.NewInt(10_000_000)), app.BankKeeper.GetSupply(ctx, appparams.BaseDenom).Amount)

	history, found, err := app.FeeCollectorKeeper.GetFeeHistory(ctx, 42)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 20_000_000)), history.Collected)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 10_000_000)), history.Burn)
	require.NoError(t, app.FeeCollectorKeeper.AssertModuleAccountingInvariant(ctx))
}

func TestGoldenNativeEconomyLoopCoversFeesEmissionsPoolRewardsAndStorageRent(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(42)
	rentPayer := AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 1_000)))[0]

	fees := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 20_000_000))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, fees))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, fees))
	supplyBefore := app.BankKeeper.GetSupply(ctx, appparams.BaseDenom)
	treasuryBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.TreasuryModuleName), appparams.BaseDenom)
	protectionBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.ProtectionModuleName), appparams.BaseDenom)
	ecosystemBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.EcosystemGrantsModuleName), appparams.BaseDenom)
	storageReserveBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.StorageRentReserveModuleName), appparams.BaseDenom)

	require.NoError(t, app.FeesKeeper.RecordCollectedFees(ctx, fees))

	validatorsBefore := app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), appparams.BaseDenom)
	_, err := app.EndBlocker(ctx)
	require.NoError(t, err)

	configureGoldenBurnParams(t, app, ctx)
	configureGoldenEmissionParams(t, app, ctx)
	ctx = ctx.WithBlockHeight(43)
	emission, err := app.FinalizeNativeEconomyEpoch(ctx, 7, 5_000)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 1_000), emission.EmissionAmount)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 700), emission.ValidatorReward)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 100), emission.Treasury)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 100), emission.ProtectionFund)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 50), emission.Burn)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 50), emission.Ecosystem)
	require.Equal(t, appparams.BaseDenom, emission.RoundingRemainder.Denom)
	require.True(t, emission.RoundingRemainder.Amount.IsZero())

	nextPool, rewardSummary, err := nominatorpooltypes.SyncPoolRewards(nominatorpooltypes.DefaultParams(), nominatorpooltypes.NominatorPool{
		PoolID:			"golden-pool",
		TotalBondedStake:	1_000,
		TotalShares:		1_000,
		PoolCommissionBps:	100,
	}, nominatorpooltypes.MsgSyncPoolRewards{
		Authority:		nominatorpooltypes.DefaultParams().Authority,
		PoolID:			"golden-pool",
		Epoch:			7,
		RewardRateBps:		1_000,
		EmissionsAllocated:	uint64(emission.ValidatorReward.Amount.Uint64()),
		FeesAllocated:		7_000_000,
		Height:			43,
		Allocations: []nominatorpooltypes.ValidatorRewardAllocation{{
			Validator:		testAEAddress(0x51),
			PoolAllocatedStake:	1_000,
			ValidatorSelfStake:	500,
			PerformanceBps:		10_000,
			CommissionBps:		500,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(100), rewardSummary.GrossPoolRewards)
	require.Equal(t, uint64(5), rewardSummary.ValidatorCommission)
	require.Equal(t, uint64(0), rewardSummary.PoolProtocolFee)
	require.Equal(t, uint64(95), rewardSummary.PoolUserRewards)
	require.Equal(t, uint64(50), rewardSummary.ValidatorSelfStakeRewards)
	require.LessOrEqual(t, rewardSummary.PoolUserRewards, rewardSummary.EmissionsAllocated+rewardSummary.FeesAllocated)
	require.Equal(t, uint64(95_000_000), nextPool.RewardIndex)
	require.Equal(t, uint64(5), nextPool.ValidatorCommissionAccrued)
	require.Equal(t, uint64(1_095), nextPool.TotalBondedStake)

	allocations, remainder, err := app.FeeCollectorKeeper.CollectAndDistributeProtocolIncomeFromAccount(ctx, rentPayer, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 100)))
	require.NoError(t, err)
	require.True(t, remainder.Empty())
	byBucket := map[string]sdk.Coins{}
	for _, allocation := range allocations {
		byBucket[allocation.Bucket] = allocation.Amount
	}
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 5)), byBucket[feecollectortypes.BucketStorageRentReserve])
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 2)), byBucket[feecollectortypes.BucketBurn])
	require.Equal(t, sdk.NewCoins(), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.CollectorModuleName)))

	require.Equal(t, validatorsBefore.Add(sdk.NewInt64Coin(appparams.BaseDenom, 7_000_738)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), appparams.BaseDenom))
	require.Equal(t, treasuryBefore.Add(sdk.NewInt64Coin(appparams.BaseDenom, 3_000_125)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.TreasuryModuleName), appparams.BaseDenom))
	require.Equal(t, protectionBefore.Add(sdk.NewInt64Coin(appparams.BaseDenom, 110)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.ProtectionModuleName), appparams.BaseDenom))
	require.Equal(t, ecosystemBefore.Add(sdk.NewInt64Coin(appparams.BaseDenom, 62)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.EcosystemGrantsModuleName), appparams.BaseDenom))
	require.Equal(t, storageReserveBefore.Add(sdk.NewInt64Coin(appparams.BaseDenom, 5)), app.BankKeeper.GetBalance(ctx, app.AccountKeeper.GetModuleAddress(feecollectortypes.StorageRentReserveModuleName), appparams.BaseDenom))
	require.Equal(t, supplyBefore.Amount.Sub(sdkmath.NewInt(10_000_000)).Add(sdkmath.NewInt(950)).Sub(sdkmath.NewInt(2)), app.BankKeeper.GetSupply(ctx, appparams.BaseDenom).Amount)

	burned, found, err := app.BurnKeeper.GetBurnedDenomEntry(ctx, burntypes.BaseDenom)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 50)), burned.Amount)
	emissionsGenesis, err := app.EmissionsKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 1_000), emissionsGenesis.TotalMintedAccounting)
	require.NoError(t, emissionsGenesis.Validate())
	feeCollectorGenesis, err := app.FeeCollectorKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, feeCollectorGenesis.Validate())
	burnGenesis, err := app.BurnKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, burnGenesis.Validate())
	mintAuthorityGenesis, err := app.MintAuthorityKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, mintAuthorityGenesis.Validate())
	require.NoError(t, app.TreasuryKeeper.SyncIncomingFunds(ctx))
	treasuryGenesis, err := app.TreasuryKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.NoError(t, treasuryGenesis.Validate())
	require.NoError(t, app.FeeCollectorKeeper.AssertModuleAccountingInvariant(ctx))
	require.NoError(t, app.TreasuryKeeper.AssertTreasuryAccountingInvariant(ctx))

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
	require.NoError(t, importedApp.MintAuthorityKeeper.InitGenesis(importedCtx, *mintAuthorityGenesis))
	require.NoError(t, importedApp.TreasuryKeeper.InitGenesis(importedCtx, *treasuryGenesis))
	reexportedFees, err := importedApp.FeeCollectorKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, feeCollectorGenesis, reexportedFees)
	reexportedBurn, err := importedApp.BurnKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, burnGenesis, reexportedBurn)
	reexportedEmissions, err := importedApp.EmissionsKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, emissionsGenesis, reexportedEmissions)
	reexportedMintAuthority, err := importedApp.MintAuthorityKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, mintAuthorityGenesis, reexportedMintAuthority)
	reexportedTreasury, err := importedApp.TreasuryKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, treasuryGenesis, reexportedTreasury)
	require.NoError(t, importedApp.FeeCollectorKeeper.AssertModuleAccountingInvariant(importedCtx))
	require.NoError(t, importedApp.TreasuryKeeper.AssertTreasuryAccountingInvariant(importedCtx))
}

func configureGoldenBurnParams(t *testing.T, app *L1App, ctx sdk.Context) {
	t.Helper()
	params := burntypes.DefaultParams()
	params.ProtocolBurnPermissions = append(params.ProtocolBurnPermissions, burntypes.BurnPermission{
		ModuleName:	authtypes.FeeCollectorName,
		AllowedDenoms:	[]string{appparams.BaseDenom},
	})
	require.NoError(t, app.BurnKeeper.SetParams(ctx, params))
}

func configureGoldenEmissionParams(t *testing.T, app *L1App, ctx sdk.Context) {
	t.Helper()
	params := emissionstypes.DefaultParams()
	params.CurrentInflationBps = 1_000
	params.MinAnnualInflationBps = 1_000
	params.MaxAnnualInflationBps = 1_000
	params.ConstitutionalMaxInflationBps = 1_000
	params.TargetStakingRatioBps = 5_000
	params.ResponsivenessBps = 1
	params.AnnualReferenceSupply = sdk.NewInt64Coin(appparams.BaseDenom, 10_000)
	params.EpochsPerYear = 1
	params.DistributionWeights = emissionstypes.DefaultDistributionWeights()
	require.NoError(t, app.EmissionsKeeper.SetParams(ctx, params))
}

func testAEAddress(fill byte) string {
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bytes.Repeat([]byte{fill}, 20)))
}

func TestEmissionEpochDuplicateRejected(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(42)
	configureGoldenBurnParams(t, app, ctx)
	configureGoldenEmissionParams(t, app, ctx)

	_, err := app.FinalizeNativeEconomyEpoch(ctx, 7, 5_000)
	require.NoError(t, err)

	_, err = app.FinalizeNativeEconomyEpoch(ctx, 7, 5_000)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already finalized")
}

func TestEmissionCapInvariant(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(42)
	configureGoldenBurnParams(t, app, ctx)
	configureGoldenEmissionParams(t, app, ctx)

	emission, err := app.FinalizeNativeEconomyEpoch(ctx, 7, 5_000)
	require.NoError(t, err)

	totalDistributed := emission.ValidatorReward.Amount.
		Add(emission.Treasury.Amount).
		Add(emission.ProtectionFund.Amount).
		Add(emission.Burn.Amount).
		Add(emission.Ecosystem.Amount).
		Add(emission.RoundingRemainder.Amount)
	require.Equal(t, emission.EmissionAmount.Amount, totalDistributed,
		"total distributed rewards must equal total emission amount")
}

func TestJailedValidatorProducesZeroPoolRewards(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(42)
	configureGoldenBurnParams(t, app, ctx)
	configureGoldenEmissionParams(t, app, ctx)

	emission, err := app.FinalizeNativeEconomyEpoch(ctx, 7, 5_000)
	require.NoError(t, err)

	_, rewardSummary, err := nominatorpooltypes.SyncPoolRewards(nominatorpooltypes.DefaultParams(), nominatorpooltypes.NominatorPool{
		PoolID:			"jail-test-pool",
		TotalBondedStake:	1_000,
		TotalShares:		1_000,
		PoolCommissionBps:	100,
	}, nominatorpooltypes.MsgSyncPoolRewards{
		Authority:		nominatorpooltypes.DefaultParams().Authority,
		PoolID:			"jail-test-pool",
		Epoch:			7,
		RewardRateBps:		1_000,
		EmissionsAllocated:	uint64(emission.ValidatorReward.Amount.Uint64()),
		FeesAllocated:		0,
		Height:			42,
		Allocations: []nominatorpooltypes.ValidatorRewardAllocation{{
			Validator:		testAEAddress(0x51),
			PoolAllocatedStake:	1_000,
			ValidatorSelfStake:	500,
			PerformanceBps:		10_000,
			CommissionBps:		500,
			Jailed:			true,
		}},
	})
	require.NoError(t, err)

	require.Equal(t, uint64(0), rewardSummary.GrossPoolRewards)
	require.Equal(t, uint64(0), rewardSummary.ValidatorCommission)
	require.Equal(t, uint64(0), rewardSummary.PoolProtocolFee)
	require.Equal(t, uint64(0), rewardSummary.PoolUserRewards)
	require.Equal(t, uint64(0), rewardSummary.ValidatorSelfStakeRewards)
	require.Equal(t, uint64(0), rewardSummary.ValidatorGrossIncome)
}

func TestMaybeFinalizeNativeEmissionEpochAtEpochBoundary(t *testing.T) {
	app := Setup(t, false)

	epochBoundary := int64(nominatorpooltypes.DefaultRewardEpochDurationBlocks)
	ctx := app.NewContext(false).WithBlockHeight(epochBoundary)
	configureGoldenBurnParams(t, app, ctx)
	configureGoldenEmissionParams(t, app, ctx)

	err := app.maybeFinalizeNativeEmissionEpoch(ctx)
	require.NoError(t, err)

	epoch1, found, err := app.EmissionsKeeper.GetEmissionEpoch(ctx, 1)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(1), epoch1.Epoch)

	err = app.maybeFinalizeNativeEmissionEpoch(ctx)
	require.NoError(t, err)
}
