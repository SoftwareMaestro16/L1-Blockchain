package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	"github.com/sovereign-l1/l1/x/reputation/types"
	reputationpb "github.com/sovereign-l1/l1/x/reputation/types/reputationpb"
)

func TestNativeReputationRewardPenaltyAndExport(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := reputationkeeper.NewMsgServerImpl(app.ReputationKeeper)
	subject := addr(0x11)

	_, err := msgServer.ApplyReputationReward(ctx, &reputationpb.MsgApplyReputationReward{
		Authority:	app.ReputationKeeper.Authority(),
		SubjectType:	types.SubjectAccount,
		Subject:	subject,
		Component:	types.ComponentUptime,
		Epoch:		1,
	})
	require.NoError(t, err)

	_, err = msgServer.ApplyReputationPenalty(ctx, &reputationpb.MsgApplyReputationPenalty{
		Authority:	addr(0x22),
		SubjectType:	types.SubjectAccount,
		Subject:	subject,
		Component:	types.ComponentSlashing,
		Epoch:		2,
	})
	require.Error(t, err)

	exported, err := app.ReputationKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	imported := l1app.Setup(t, false)
	importedCtx := imported.NewContext(false)
	require.NoError(t, imported.ReputationKeeper.InitGenesis(importedCtx, *exported))
	roundTrip, err := imported.ReputationKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}

func TestStakeReputationClaimQueryAndExport(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	account := addr(0x33)

	_, err := app.ReputationKeeper.ClaimStakeReputation(ctx, types.MsgClaimStakeReputation{
		Authority:		app.ReputationKeeper.Authority(),
		Account:		account,
		PoolID:			"pool-a",
		PoolShares:		100,
		PoolTotalShares:	100,
		PoolActiveStake:	10_000,
		TimestampUnix:		1,
	})
	require.NoError(t, err)

	claim, err := app.ReputationKeeper.ClaimStakeReputation(ctx, types.MsgClaimStakeReputation{
		Authority:		app.ReputationKeeper.Authority(),
		Account:		account,
		PoolID:			"pool-a",
		PoolShares:		100,
		PoolTotalShares:	100,
		PoolActiveStake:	10_000,
		TimestampUnix:		3_601,
	})
	require.NoError(t, err)
	require.Greater(t, uint64(claim.Score), uint64(0))

	id, err := app.ReputationKeeper.GetIdentityReputation(ctx, account)
	require.NoError(t, err)
	require.Contains(t, id.Account, "AE")
	require.Greater(t, id.SignalCounters.StakeTimeSeconds, uint64(0))

	exported, err := app.ReputationKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	imported := l1app.Setup(t, false)
	importedCtx := imported.NewContext(false)
	require.NoError(t, imported.ReputationKeeper.InitGenesis(importedCtx, *exported))
	roundTrip, err := imported.ReputationKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}

func TestGetIdentityReputationScore(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	addr := sdk.AccAddress(bytes20(0x99))
	score, found, err := app.ReputationKeeper.GetIdentityReputationScore(ctx, addr)
	require.NoError(t, err)
	require.False(t, found)
	require.Equal(t, uint32(100), score)
}

func bytes20(fill byte) []byte {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return bz
}

func addr(fill byte) string {
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bytes20(fill)))
}
