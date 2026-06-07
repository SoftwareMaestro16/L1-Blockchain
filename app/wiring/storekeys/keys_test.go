package storekeys

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
)

func TestNewKVStoreKeysIncludesCoreAndNativeSystemModules(t *testing.T) {
	keys := NewKVStoreKeys()

	for _, storeKey := range []string{
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		aetracoretypes.StoreKey,
		feecollectortypes.StoreKey,
	} {
		require.Contains(t, keys, storeKey)
		require.NotNil(t, keys[storeKey])
	}
}
