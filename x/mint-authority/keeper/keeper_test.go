package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/mint-authority/types"
)

func TestNativeMintAuthorityMintsAndAccounts(t *testing.T) {
	app := l1app.Setup(t, false)
	ctx := app.NewContext(false)
	recipient := sdk.AccAddress(bytes20(0x44))
	recipientText := aetraaddress.FormatAccAddress(recipient)
	decision := types.EmissionDecision{
		Caller:		types.DefaultEmissionCaller,
		Denom:		types.DefaultBaseDenom,
		Amount:		sdkmath.NewInt(100),
		Epoch:		1,
		Height:		10,
		Approved:	true,
	}
	decision.DecisionHash = types.ComputeEmissionDecisionHash(decision)

	_, err := app.MintAuthorityKeeper.MintProtocolCoins(ctx, types.MsgMintProtocolCoins{
		Caller:			types.DefaultEmissionCaller,
		Recipient:		recipientText,
		Denom:			types.DefaultBaseDenom,
		Amount:			sdkmath.NewInt(100),
		Epoch:			1,
		Height:			10,
		EmissionsDecisionHash:	decision.DecisionHash,
	}, decision, types.ConstitutionEmergencyAuthorization{})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(100), app.BankKeeper.GetBalance(ctx, recipient, types.DefaultBaseDenom).Amount)

	_, err = app.MintAuthorityKeeper.MintProtocolCoins(ctx, types.MsgMintProtocolCoins{
		Caller:			"x/attacker",
		Recipient:		recipientText,
		Denom:			types.DefaultBaseDenom,
		Amount:			sdkmath.NewInt(1),
		Epoch:			1,
		Height:			11,
		EmissionsDecisionHash:	decision.DecisionHash,
	}, decision, types.ConstitutionEmergencyAuthorization{})
	require.Error(t, err)
}

func bytes20(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
