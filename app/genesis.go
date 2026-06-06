package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sovereign-l1/l1/app/genesisconfig"
)

// GenesisState of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState = genesisconfig.State

func withCoreModuleGenesisDefaults(cdc codec.JSONCodec, genesis map[string]json.RawMessage) map[string]json.RawMessage {
	return genesisconfig.ApplyCoreModuleDefaults(cdc, genesis)
}

func withNativeTokenMetadata(cdc codec.Codec, genesis map[string]json.RawMessage) map[string]json.RawMessage {
	return genesisconfig.ApplyNativeTokenMetadata(cdc, genesis)
}
