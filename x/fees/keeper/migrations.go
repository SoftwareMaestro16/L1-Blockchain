package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

type Migrator struct {
	keeper Keeper
}

func NewMigrator(k Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	gs, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		return err
	}
	return gs.Validate()
}
