package genesisvalidation

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
)

func EnsureCollectionItem[T any](ctx context.Context, item collections.Item[T], defaultValue T) error {
	if _, err := item.Get(ctx); err == nil {
		return nil
	} else if !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	return item.Set(ctx, defaultValue)
}
