package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/sovereign-l1/l1/app/wiring/storekeys"
)

func newKVStoreKeys() map[string]*storetypes.KVStoreKey {
	return storekeys.NewKVStoreKeys()
}
