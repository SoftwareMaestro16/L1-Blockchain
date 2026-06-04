package app

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestOrbitalisChainConstants(t *testing.T) {
	require.Equal(t, "Orbitalis", appName)
	require.Equal(t, "orb", AccountAddressPrefix)
	require.Equal(t, "orbvaloper", ValidatorAddressPrefix)
	require.Equal(t, "orbvalcons", ConsensusAddressPrefix)
	require.Equal(t, "norb", BondDenom)
	require.Equal(t, "norb", sdk.DefaultBondDenom)
	require.True(t, strings.HasSuffix(DefaultNodeHome, ".orbitalis"), DefaultNodeHome)
}

func TestDefaultGenesisIncludesNativeTokenMetadata(t *testing.T) {
	app, genesis := setup(true, 5)

	var bankGenState banktypes.GenesisState
	app.AppCodec().MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)

	var native banktypes.Metadata
	for _, metadata := range bankGenState.DenomMetadata {
		if metadata.Base == appparams.BaseDenom {
			native = metadata
			break
		}
	}

	require.Equal(t, appparams.NativeTokenMetadata(), native)
	require.NoError(t, native.Validate())
}
