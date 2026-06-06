package keeper

import (
	"errors"

	coretypes "github.com/sovereign-l1/l1/x/aethercore/types"
	"github.com/sovereign-l1/l1/x/contracts/types"
)

type Keeper struct {
	genesis types.GenesisState
}

func NewKeeper() Keeper {
	return Keeper{genesis: types.DefaultGenesis()}
}

func DefaultGenesis() types.GenesisState {
	return types.DefaultGenesis()
}

func (k *Keeper) InitGenesis(gs types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = gs
	return nil
}

func (k Keeper) ExportGenesis() types.GenesisState {
	return k.genesis
}

func (k Keeper) Params() types.Params {
	return k.genesis.Params
}

func (k Keeper) ValidateInvariants() error {
	return k.genesis.Validate()
}

func (k Keeper) RootContribution() (coretypes.RootContribution, error) {
	return types.RootContribution(k.genesis)
}

func (k *Keeper) StoreCode(msg types.MsgStoreCode) (types.StoreCodeResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.StoreCodeResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if msg.CodeBytes == 0 || msg.CodeBytes > k.genesis.Params.MaxCodeBytes {
		return types.StoreCodeResponse{}, errors.New(types.ErrInvalidBytecode + ": code size out of bounds")
	}
	if err := coretypes.ValidateHash("contracts code hash", msg.CodeHash); err != nil {
		return types.StoreCodeResponse{}, err
	}
	return types.StoreCodeResponse{CodeID: msg.CodeHash, StateRoot: k.genesis.StateRoot}, nil
}

func (k Keeper) Contract(req types.QueryContractRequest) (types.QueryContractResponse, error) {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return types.QueryContractResponse{}, err
	}
	return types.QueryContractResponse{ContractAddress: req.ContractAddress, StateRoot: k.genesis.StateRoot, Found: false}, nil
}
