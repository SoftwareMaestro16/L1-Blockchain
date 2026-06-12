package prefixgenesis

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
)

type testGenesis struct {
	Version	uint64
	Params	testParams
	State	testState
}

type testParams struct {
	Enabled bool
}

type testState struct {
	Records []string
}

func TestLoadMigratesLegacyGenesisBlobToPrefixLayout(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	legacyKey := []byte{0x01}
	legacy := testGenesis{
		Version:	2,
		Params:		testParams{Enabled: true},
		State:		testState{Records: []string{"b", "a"}},
	}
	bz, err := json.Marshal(legacy)
	require.NoError(t, err)
	require.NoError(t, service.RawStore().Set(legacyKey, bz))

	loaded, migrated, err := Load(ctx, service, legacyKey, testGenesis{})
	require.NoError(t, err)
	require.True(t, migrated)
	require.Equal(t, legacy, loaded)

	legacyValue, err := service.RawStore().Get(legacyKey)
	require.NoError(t, err)
	require.Empty(t, legacyValue)
	marker, err := service.RawStore().Get(layoutKey)
	require.NoError(t, err)
	require.Equal(t, []byte("v2"), marker)
	state, err := service.RawStore().Get([]byte("prefix_genesis/state"))
	require.NoError(t, err)
	require.NotEmpty(t, state)

	reloaded, found, err := Load(ctx, service, legacyKey, testGenesis{})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, legacy, reloaded)
}
