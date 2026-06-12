package params

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func TestNativeTokenMetadataContract(t *testing.T) {
	metadata := NativeTokenMetadata()

	require.Equal(t, BaseDenom, metadata.Base)
	require.Equal(t, DisplayDenom, metadata.Display)
	require.Equal(t, TokenName, metadata.Name)
	require.Equal(t, TokenSymbol, metadata.Symbol)
	require.Equal(t, int64(1_000_000_000), BaseUnitsPerDisplay)
	require.NoError(t, metadata.Validate())
	require.ElementsMatch(t, []*banktypes.DenomUnit{
		{Denom: BaseDenom, Exponent: 0},
		{Denom: DisplayDenom, Exponent: DisplayDenomExponent},
	}, metadata.DenomUnits)
}

func TestEnsureNativeTokenMetadataReplacesStaleEntry(t *testing.T) {
	const fixtureDenom = "fixturetoken"

	stale := banktypes.Metadata{
		Base:		BaseDenom,
		Display:	"old",
	}
	other := banktypes.Metadata{
		Base:		fixtureDenom,
		Display:	fixtureDenom,
		DenomUnits:	[]*banktypes.DenomUnit{{Denom: fixtureDenom, Exponent: 0}},
	}

	metadata := EnsureNativeTokenMetadata([]banktypes.Metadata{stale, other})

	require.Len(t, metadata, 2)
	require.Equal(t, NativeTokenMetadata(), metadata[0])
	require.Equal(t, other, metadata[1])
}
