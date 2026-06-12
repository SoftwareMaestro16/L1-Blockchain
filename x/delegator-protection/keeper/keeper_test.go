package keeper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	delegatorprotectionkeeper "github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	delegatorprotectionpb "github.com/sovereign-l1/l1/x/delegator-protection/types/delegatorprotectionpb"
)

func TestNativeDelegatorProtectionClaimAndAuthority(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	msgServer := delegatorprotectionkeeper.NewMsgServerImpl(app.DelegatorProtectionKeeper)

	res, err := msgServer.SubmitDelegatorProtectionClaim(ctx, &delegatorprotectionpb.MsgSubmitDelegatorProtectionClaim{
		Delegator:		"delegator-a",
		Validator:		"validator-a",
		LossAmount:		"1000",
		RequestedPayout:	"500",
		EligibilityHash:	strings.Repeat("a", 64),
		Reason:			"slash",
		Epoch:			1,
		Height:			10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, res.ClaimJson)

	_, err = msgServer.UpdateProtectionParams(ctx, &delegatorprotectionpb.MsgUpdateProtectionParams{
		Authority:	"wrong",
		ParamsJson:	`{"Authority":"wrong"}`,
	})
	require.Error(t, err)

	exported, err := app.DelegatorProtectionKeeper.ExportGenesis(ctx)
	require.NoError(t, err)
	imported := l1app.Setup(t, false)
	importedCtx := imported.NewContext(false)
	require.NoError(t, imported.DelegatorProtectionKeeper.InitGenesis(importedCtx, *exported))
	roundTrip, err := imported.DelegatorProtectionKeeper.ExportGenesis(importedCtx)
	require.NoError(t, err)
	require.Equal(t, exported, roundTrip)
}
