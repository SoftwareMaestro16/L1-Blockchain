package contracts

import (
	"github.com/sovereign-l1/l1/x/contracts/keeper"
	"github.com/sovereign-l1/l1/x/contracts/types"
)

const ModuleName = types.ModuleName

type AppModule struct {
	Keeper keeper.Keeper
}

func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{Keeper: k}
}

func (AppModule) Name() string {
	return ModuleName
}
