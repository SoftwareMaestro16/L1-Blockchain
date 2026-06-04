package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrbitalisChainConstants(t *testing.T) {
	require.Equal(t, "Orbitalis", appName)
	require.Equal(t, "orb", AccountAddressPrefix)
	require.Equal(t, "orbvaloper", ValidatorAddressPrefix)
	require.Equal(t, "orbvalcons", ConsensusAddressPrefix)
	require.Equal(t, "uorb", BondDenom)
	require.True(t, strings.HasSuffix(DefaultNodeHome, ".orbitalis"), DefaultNodeHome)
}
