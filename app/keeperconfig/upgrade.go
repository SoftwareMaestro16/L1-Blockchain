package keeperconfig

import (
	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

func UnsafeSkipUpgradeHeights(appOpts servertypes.AppOptions) map[int64]bool {
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	return skipUpgradeHeights
}
