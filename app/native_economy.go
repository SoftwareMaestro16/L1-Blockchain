package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
)

// FinalizeNativeEconomyEpoch connects emission accounting to bank supply and
// module balances. Rounding remainder is credited to treasury/community.
func (app *L1App) FinalizeNativeEconomyEpoch(ctx sdk.Context, epoch uint64, stakingRatioBps uint32) (emissionstypes.EmissionEpoch, error) {
	if ctx.BlockHeight() < 0 {
		return emissionstypes.EmissionEpoch{}, fmt.Errorf("native economy epoch height cannot be negative")
	}
	record, err := app.EmissionsKeeper.FinalizeEmissionEpoch(ctx, epoch, stakingRatioBps)
	if err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}
	if record.EmissionAmount.Amount.IsZero() {
		return record, nil
	}

	decision := mintauthoritytypes.EmissionDecision{
		Caller:		mintauthoritytypes.DefaultEmissionCaller,
		Denom:		record.EmissionAmount.Denom,
		Amount:		record.EmissionAmount.Amount,
		Epoch:		epoch,
		Height:		uint64(ctx.BlockHeight()),
		Approved:	true,
	}
	decision.DecisionHash = mintauthoritytypes.ComputeEmissionDecisionHash(decision)

	state, err := app.MintAuthorityKeeper.GetState(ctx)
	if err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}
	newState, _, err := mintauthoritytypes.ApplyMintProtocolCoins(state, mintauthoritytypes.MsgMintProtocolCoins{
		Caller:			mintauthoritytypes.DefaultEmissionCaller,
		Recipient:		aetraaddress.FormatAccAddress(app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)),
		Denom:			record.EmissionAmount.Denom,
		Amount:			record.EmissionAmount.Amount,
		Epoch:			epoch,
		Height:			uint64(ctx.BlockHeight()),
		EmissionsDecisionHash:	decision.DecisionHash,
	}, decision, mintauthoritytypes.ConstitutionEmergencyAuthorization{})
	if err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}
	if err := app.MintAuthorityKeeper.SetState(ctx, newState); err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}

	if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(record.EmissionAmount)); err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}
	if err := app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, sdk.NewCoins(record.EmissionAmount)); err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}
	if err := app.distributeNativeEmission(ctx, epoch, record); err != nil {
		return emissionstypes.EmissionEpoch{}, err
	}
	return record, nil
}

func (app *L1App) maybeFinalizeNativeEmissionEpoch(ctx sdk.Context) error {
	if ctx.BlockHeight() <= 0 {
		return nil
	}
	interval := uint64(nominatorpooltypes.DefaultRewardEpochDurationBlocks)
	height := uint64(ctx.BlockHeight())
	if interval == 0 || height%interval != 0 {
		return nil
	}
	epoch := height / interval
	if _, found, err := app.EmissionsKeeper.GetEmissionEpoch(ctx, epoch); err != nil {
		return err
	} else if found {
		return nil
	}
	params, err := app.EmissionsKeeper.GetParams(ctx)
	if err != nil {
		return err
	}
	_, err = app.FinalizeNativeEconomyEpoch(ctx, epoch, params.TargetStakingRatioBps)
	return err
}

func (app *L1App) distributeNativeEmission(ctx sdk.Context, epoch uint64, record emissionstypes.EmissionEpoch) error {

	treasury := record.Treasury
	if record.RoundingRemainder.Amount.IsPositive() {
		treasury = treasury.Add(record.RoundingRemainder)
	}
	if err := app.sendFromFeeCollector(ctx, feecollectortypes.TreasuryModuleName, treasury); err != nil {
		return err
	}
	if err := app.sendFromFeeCollector(ctx, feecollectortypes.ProtectionModuleName, record.ProtectionFund); err != nil {
		return err
	}
	if err := app.sendFromFeeCollector(ctx, feecollectortypes.EcosystemGrantsModuleName, record.Ecosystem); err != nil {
		return err
	}
	if record.Burn.Amount.IsPositive() {
		if _, err := app.BurnKeeper.BurnProtocolCoins(ctx, authtypes.FeeCollectorName, sdk.NewCoins(record.Burn), epoch, "emissions.distribute"); err != nil {
			return err
		}
	}
	return nil
}

func (app *L1App) sendFromFeeCollector(ctx sdk.Context, recipientModule string, coin sdk.Coin) error {
	if coin.Amount.IsNil() || !coin.Amount.IsPositive() {
		return nil
	}
	return app.BankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, recipientModule, sdk.NewCoins(coin))
}
