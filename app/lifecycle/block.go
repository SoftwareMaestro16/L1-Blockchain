package lifecycle

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sovereign-l1/l1/app/genesisconfig"
	"github.com/sovereign-l1/l1/observability"
)

func FinalizeBlock(req *abci.RequestFinalizeBlock, finalize func(*abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error)) (*abci.ResponseFinalizeBlock, error) {
	res, err := finalize(req)

	observability.RecordFinalizeBlock(req.Height, req.Time, len(req.Txs), -1)
	if err != nil {
		observability.RecordModuleError("app", "finalize_block", "error")
	}
	return res, err
}

type InitChainDependencies struct {
	AppCodec			codec.Codec
	ModuleManager			*module.Manager
	SetModuleVersionMap		func(sdk.Context, module.VersionMap) error
	ValidateGenesis			func(genesisconfig.State) error
	EnsureCoreGenesisCollections	func(sdk.Context) error
}

func InitChain(ctx sdk.Context, req *abci.RequestInitChain, deps InitChainDependencies) (*abci.ResponseInitChain, error) {
	var genesisState genesisconfig.State
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	if err := deps.SetModuleVersionMap(ctx, deps.ModuleManager.GetVersionMap()); err != nil {
		return nil, err
	}
	if err := deps.ValidateGenesis(genesisState); err != nil {
		return nil, err
	}
	res, err := deps.ModuleManager.InitGenesis(ctx, deps.AppCodec, genesisState)
	if err != nil {
		return nil, err
	}
	if err := deps.EnsureCoreGenesisCollections(ctx); err != nil {
		return nil, err
	}
	return res, nil
}
