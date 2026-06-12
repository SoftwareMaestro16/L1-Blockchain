package genesisaccount

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestSimGenesisAccountRejectsNilBaseAccount(t *testing.T) {
	require.ErrorContains(t, SimGenesisAccount{}.Validate(), "base account")
}

func TestSimGenesisAccountRejectsZeroAddress(t *testing.T) {
	account := SimGenesisAccount{BaseAccount: authtypes.NewBaseAccountWithAddress(sdk.AccAddress(make([]byte, 20)))}
	require.ErrorContains(t, account.Validate(), "zero address")
}

func TestSimGenesisAccountRejectsInvalidVestingWindow(t *testing.T) {
	account := SimGenesisAccount{
		BaseAccount:		authtypes.NewBaseAccountWithAddress(testAddress(0x11)),
		OriginalVesting:	sdk.NewCoins(sdk.NewInt64Coin("naet", 1)),
		StartTime:		10,
		EndTime:		10,
	}

	require.ErrorContains(t, account.Validate(), "vesting start-time")
}

func testAddress(fill byte) sdk.AccAddress {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return sdk.AccAddress(out)
}
