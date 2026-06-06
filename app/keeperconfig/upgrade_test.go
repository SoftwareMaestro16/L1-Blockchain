package keeperconfig

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/server"
	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
)

func TestUnsafeSkipUpgradeHeightsParsesAppOptions(t *testing.T) {
	heights := UnsafeSkipUpgradeHeights(sims.AppOptionsMap{
		server.FlagUnsafeSkipUpgrades: []int{7, 11},
	})

	require.Equal(t, map[int64]bool{7: true, 11: true}, heights)
}
