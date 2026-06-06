package genesisconfig

import (
	"encoding/json"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func ApplyCoreModuleDefaults(cdc codec.JSONCodec, genesis map[string]json.RawMessage) map[string]json.RawMessage {
	setDefaultGenesis(cdc, genesis, distrtypes.ModuleName, distrtypes.DefaultGenesisState())
	setDefaultGenesis(cdc, genesis, govtypes.ModuleName, govv1.DefaultGenesisState())
	setDefaultGenesis(cdc, genesis, minttypes.ModuleName, appparams.AetraMintGenesisState())
	setDefaultGenesis(cdc, genesis, protocolpooltypes.ModuleName, protocolpooltypes.DefaultGenesisState())
	return genesis
}

func ApplyNativeTokenMetadata(cdc codec.Codec, genesis map[string]json.RawMessage) map[string]json.RawMessage {
	var bankGenState banktypes.GenesisState
	cdc.MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)
	bankGenState.DenomMetadata = appparams.EnsureNativeTokenMetadata(bankGenState.DenomMetadata)
	genesis[banktypes.ModuleName] = cdc.MustMarshalJSON(&bankGenState)
	return genesis
}

func setDefaultGenesis(cdc codec.JSONCodec, genesis map[string]json.RawMessage, moduleName string, state proto.Message) {
	genesis[moduleName] = cdc.MustMarshalJSON(state)
}
