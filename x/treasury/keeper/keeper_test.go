package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	treasurykeeper "github.com/sovereign-l1/l1/x/treasury/keeper"
	"github.com/sovereign-l1/l1/x/treasury/types"
)

func TestFeeDistributionEntersTreasuryAccounting(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	user := l1app.AddTestAddrsWithCoins(t, app, ctx, 1, sdk.NewCoins(coin(1_000)))[0]

	require.NoError(t, app.FeeCollectorKeeper.CollectFeesFromAccount(ctx, user, sdk.NewCoins(coin(1_000)), feecollectortypes.FeeTypeProtocol))
	_, err := app.FeeCollectorKeeper.DistributeFees(ctx, 3)
	require.NoError(t, err)
	require.NoError(t, app.TreasuryKeeper.SyncIncomingFunds(ctx))

	moduleBalance := app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(types.TreasuryModuleName))
	require.Equal(t, sdk.NewCoins(coin(150)), moduleBalance)
	allocations, err := app.TreasuryKeeper.GetAllocations(ctx)
	require.NoError(t, err)
	require.Equal(t, moduleBalance, allocations.AccountingBalance())
	require.Equal(t, sdk.NewCoins(coin(76)), allocations.ReserveBalance)
	require.Equal(t, sdk.NewCoins(coin(45)), allocations.EcosystemBalance)
	require.Equal(t, sdk.NewCoins(coin(22)), allocations.ValidatorIncentiveBalance)
	require.Equal(t, sdk.NewCoins(coin(7)), allocations.BurnBalance)
	require.NoError(t, app.TreasuryKeeper.AssertTreasuryAccountingInvariant(ctx))
}

func TestSpendProposalLifecycle(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsWithCoins(t, app, ctx, 2, sdk.NewCoins())
	proposer, recipient := addrs[0], addrs[1]
	l1app.FundTestAddr(t, app, ctx, proposer, sdk.NewCoins(coin(10)))
	fundTreasury(t, app, ctx, 1_000)
	msgServer := treasurykeeper.NewMsgServerImpl(app.TreasuryKeeper)

	submit, err := msgServer.SubmitTreasurySpend(ctx, &types.MsgSubmitTreasurySpend{
		Proposer:	aetraaddress.FormatAccAddress(proposer),
		Recipient:	aetraaddress.FormatAccAddress(recipient),
		Amount:		sdk.NewCoins(coin(100)),
		Bucket:		types.BucketEcosystem,
		Epoch:		4,
		Metadata:	"grant",
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusPending, submit.Spend.Status)

	approved, err := msgServer.ApproveTreasurySpend(ctx, &types.MsgApproveTreasurySpend{
		Authority:	app.TreasuryKeeper.Authority(),
		SpendId:	submit.Spend.Id,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusApproved, approved.Spend.Status)

	before := app.BankKeeper.GetBalance(ctx, recipient, types.BaseDenom)
	executed, err := msgServer.ExecuteTreasurySpend(ctx, &types.MsgExecuteTreasurySpend{
		Authority:	app.TreasuryKeeper.Authority(),
		SpendId:	submit.Spend.Id,
		Epoch:		4,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusExecuted, executed.Spend.Status)
	require.Equal(t, before.Amount.Add(sdkmath.NewInt(100)), app.BankKeeper.GetBalance(ctx, recipient, types.BaseDenom).Amount)
	require.NoError(t, app.TreasuryKeeper.AssertTreasuryAccountingInvariant(ctx))

	_, err = msgServer.CancelTreasurySpend(ctx, &types.MsgCancelTreasurySpend{
		Actor:		aetraaddress.FormatAccAddress(proposer),
		SpendId:	submit.Spend.Id,
	})
	require.ErrorIs(t, err, types.ErrInvalidSpend)
}

func TestPerEpochSpendCapEnforced(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsWithCoins(t, app, ctx, 2, sdk.NewCoins())
	proposer, recipient := addrs[0], addrs[1]
	l1app.FundTestAddr(t, app, ctx, proposer, sdk.NewCoins(coin(10)))
	fundTreasury(t, app, ctx, 1_000)
	params, err := app.TreasuryKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.PerEpochSpendCap = coin(100)
	require.NoError(t, app.TreasuryKeeper.SetParams(ctx, params))
	msgServer := treasurykeeper.NewMsgServerImpl(app.TreasuryKeeper)

	first := submitAndApprove(t, ctx, msgServer, app.TreasuryKeeper.Authority(), proposer, recipient, 80, 8, 0)
	_, err = msgServer.ExecuteTreasurySpend(ctx, &types.MsgExecuteTreasurySpend{Authority: app.TreasuryKeeper.Authority(), SpendId: first.Id, Epoch: 8})
	require.NoError(t, err)

	second := submitAndApprove(t, ctx, msgServer, app.TreasuryKeeper.Authority(), proposer, recipient, 30, 8, 0)
	_, err = msgServer.ExecuteTreasurySpend(ctx, &types.MsgExecuteTreasurySpend{Authority: app.TreasuryKeeper.Authority(), SpendId: second.Id, Epoch: 8})
	require.ErrorIs(t, err, types.ErrSpendCapExceeded)
}

func TestVestingReleaseSchedule(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsWithCoins(t, app, ctx, 2, sdk.NewCoins())
	proposer, recipient := addrs[0], addrs[1]
	l1app.FundTestAddr(t, app, ctx, proposer, sdk.NewCoins(coin(10)))
	fundTreasury(t, app, ctx, 1_000)
	msgServer := treasurykeeper.NewMsgServerImpl(app.TreasuryKeeper)
	spend := submitAndApprove(t, ctx, msgServer, app.TreasuryKeeper.Authority(), proposer, recipient, 50, 9, 10)

	_, err := msgServer.ExecuteTreasurySpend(ctx, &types.MsgExecuteTreasurySpend{Authority: app.TreasuryKeeper.Authority(), SpendId: spend.Id, Epoch: 9})
	require.ErrorIs(t, err, types.ErrInvalidSpend)
	require.True(t, app.BankKeeper.GetBalance(ctx, recipient, types.BaseDenom).IsZero())

	_, err = msgServer.ExecuteTreasurySpend(ctx, &types.MsgExecuteTreasurySpend{Authority: app.TreasuryKeeper.Authority(), SpendId: spend.Id, Epoch: 10})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(50), app.BankKeeper.GetBalance(ctx, recipient, types.BaseDenom).Amount)
}

func TestInsufficientFundsRejected(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsWithCoins(t, app, ctx, 2, sdk.NewCoins())
	proposer, recipient := addrs[0], addrs[1]
	l1app.FundTestAddr(t, app, ctx, proposer, sdk.NewCoins(coin(10)))
	fundTreasury(t, app, ctx, 100)
	msgServer := treasurykeeper.NewMsgServerImpl(app.TreasuryKeeper)
	spend := submitAndApprove(t, ctx, msgServer, app.TreasuryKeeper.Authority(), proposer, recipient, 31, 2, 0)

	_, err := msgServer.ExecuteTreasurySpend(ctx, &types.MsgExecuteTreasurySpend{Authority: app.TreasuryKeeper.Authority(), SpendId: spend.Id, Epoch: 2})
	require.ErrorIs(t, err, types.ErrInsufficientFunds)
	require.True(t, app.BankKeeper.GetBalance(ctx, recipient, types.BaseDenom).IsZero())
}

func TestRecipientAllowlistEnforced(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addrs := l1app.AddTestAddrsWithCoins(t, app, ctx, 2, sdk.NewCoins())
	proposer, recipient := addrs[0], addrs[1]
	l1app.FundTestAddr(t, app, ctx, proposer, sdk.NewCoins(coin(10)))
	params, err := app.TreasuryKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.RecipientAllowlistEnabled = true
	params.RecipientAllowlist = []string{aetraaddress.FormatAccAddress(proposer)}
	require.NoError(t, app.TreasuryKeeper.SetParams(ctx, params))
	msgServer := treasurykeeper.NewMsgServerImpl(app.TreasuryKeeper)

	_, err = msgServer.SubmitTreasurySpend(ctx, &types.MsgSubmitTreasurySpend{
		Proposer:	aetraaddress.FormatAccAddress(proposer),
		Recipient:	aetraaddress.FormatAccAddress(recipient),
		Amount:		sdk.NewCoins(coin(1)),
		Bucket:		types.BucketEcosystem,
		Epoch:		1,
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestExportImportPreservesPendingSpends(t *testing.T) {
	source := l1app.Setup(t, false)
	sourceCtx := source.NewContext(false)
	addrs := l1app.AddTestAddrsWithCoins(t, source, sourceCtx, 2, sdk.NewCoins())
	proposer, recipient := addrs[0], addrs[1]
	l1app.FundTestAddr(t, source, sourceCtx, proposer, sdk.NewCoins(coin(10)))
	spend, err := source.TreasuryKeeper.SubmitSpend(
		sourceCtx,
		aetraaddress.FormatAccAddress(proposer),
		aetraaddress.FormatAccAddress(recipient),
		sdk.NewCoins(coin(77)),
		types.BucketEcosystem,
		12,
		0,
		0,
		"pending-export",
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), spend.Id)

	exported, err := source.TreasuryKeeper.ExportGenesis(sourceCtx)
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := l1app.Setup(t, false)
	targetCtx := target.NewContext(false)
	require.NoError(t, target.TreasuryKeeper.InitGenesis(targetCtx, *exported))
	imported, err := target.TreasuryKeeper.ExportGenesis(targetCtx)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func submitAndApprove(t *testing.T, ctx sdk.Context, msgServer types.MsgServer, authority string, proposer, recipient sdk.AccAddress, amount int64, epoch, vestingEnd uint64) types.TreasurySpend {
	t.Helper()
	submit, err := msgServer.SubmitTreasurySpend(ctx, &types.MsgSubmitTreasurySpend{
		Proposer:		aetraaddress.FormatAccAddress(proposer),
		Recipient:		aetraaddress.FormatAccAddress(recipient),
		Amount:			sdk.NewCoins(coin(amount)),
		Bucket:			types.BucketEcosystem,
		Epoch:			epoch,
		VestingEndEpoch:	vestingEnd,
	})
	require.NoError(t, err)
	approved, err := msgServer.ApproveTreasurySpend(ctx, &types.MsgApproveTreasurySpend{Authority: authority, SpendId: submit.Spend.Id})
	require.NoError(t, err)
	return approved.Spend
}

func fundTreasury(t *testing.T, app *l1app.L1App, ctx sdk.Context, amount int64) {
	t.Helper()
	require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(coin(amount))))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.TreasuryModuleName, sdk.NewCoins(coin(amount))))
	require.NoError(t, app.TreasuryKeeper.SyncIncomingFunds(ctx))
}

func coin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(types.BaseDenom, amount)
}
