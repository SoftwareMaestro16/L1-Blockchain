package genesisconfig

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestApplyCoreModuleDefaultsSetsExpectedModules(t *testing.T) {
	cdc := testCodec()
	genesis := ApplyCoreModuleDefaults(cdc, map[string]json.RawMessage{})

	for _, moduleName := range []string{
		distrtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		protocolpooltypes.ModuleName,
	} {
		require.NotEmpty(t, genesis[moduleName])
	}
}

func TestApplyNativeTokenMetadataIsIdempotent(t *testing.T) {
	cdc := testCodec()
	genesis := map[string]json.RawMessage{
		banktypes.ModuleName: cdc.MustMarshalJSON(banktypes.DefaultGenesisState()),
	}

	ApplyNativeTokenMetadata(cdc, genesis)
	ApplyNativeTokenMetadata(cdc, genesis)

	var bankGenesis banktypes.GenesisState
	cdc.MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenesis)
	require.Len(t, bankGenesis.DenomMetadata, 1)
	require.Equal(t, appparams.BaseDenom, bankGenesis.DenomMetadata[0].Base)
}

func testCodec() codec.Codec {
	return codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
}
