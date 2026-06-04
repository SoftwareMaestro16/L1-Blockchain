package app

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func withCoreModuleGenesisDefaults(cdc codec.JSONCodec, genesis map[string]json.RawMessage) map[string]json.RawMessage {
	setDefaultGenesis(cdc, genesis, distrtypes.ModuleName, distrtypes.DefaultGenesisState())
	setDefaultGenesis(cdc, genesis, govtypes.ModuleName, govv1.DefaultGenesisState())
	setDefaultGenesis(cdc, genesis, minttypes.ModuleName, minttypes.DefaultGenesisState())
	setDefaultGenesis(cdc, genesis, protocolpooltypes.ModuleName, protocolpooltypes.DefaultGenesisState())
	return genesis
}

func setDefaultGenesis(cdc codec.JSONCodec, genesis map[string]json.RawMessage, moduleName string, state proto.Message) {
	genesis[moduleName] = cdc.MustMarshalJSON(state)
}
