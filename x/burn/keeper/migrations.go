package keeper

import "context"

type Migrator struct{ keeper Keeper }

func NewMigrator(k Keeper) Migrator	{ return Migrator{keeper: k} }

func (m Migrator) Migrate1to2(ctx context.Context) error {
	_, err := m.keeper.ExportGenesis(ctx)
	return err
}
