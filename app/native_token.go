package app

import (
	"encoding/json"

	appparams "github.com/sovereign-l1/l1/app/params"

	"github.com/cosmos/cosmos-sdk/codec"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func withNativeTokenMetadata(cdc codec.Codec, genesis map[string]json.RawMessage) map[string]json.RawMessage {
	var bankGenState banktypes.GenesisState
	cdc.MustUnmarshalJSON(genesis[banktypes.ModuleName], &bankGenState)
	bankGenState.DenomMetadata = appparams.EnsureNativeTokenMetadata(bankGenState.DenomMetadata)
	genesis[banktypes.ModuleName] = cdc.MustMarshalJSON(&bankGenState)
	return genesis
}
