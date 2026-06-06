package app

import (
	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

func unsafeSkipUpgradeHeights(appOpts servertypes.AppOptions) map[int64]bool {
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true // #nosec G115 -- unsafe skip upgrade heights are non-negative CLI config values.
	}
	return skipUpgradeHeights
}
