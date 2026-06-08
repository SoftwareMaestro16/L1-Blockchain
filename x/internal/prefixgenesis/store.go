package prefixgenesis

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	corestore "cosmossdk.io/core/store"
)

var layoutKey = []byte("prefix_genesis/layout")

// Load reads a deterministic per-field genesis layout. If an old monolithic
// genesis blob exists under legacyKey, it is migrated into the prefix layout.
func Load[T any](ctx context.Context, storeService corestore.KVStoreService, legacyKey []byte, defaults T) (T, bool, error) {
	if storeService == nil {
		return defaults, false, nil
	}
	store := storeService.OpenKVStore(ctx)
	if legacy, err := store.Get(legacyKey); err != nil {
		return defaults, false, err
	} else if len(legacy) != 0 {
		var migrated T
		if err := json.Unmarshal(legacy, &migrated); err != nil {
			return defaults, false, err
		}
		if err := Save(ctx, storeService, legacyKey, migrated); err != nil {
			return defaults, false, err
		}
		return migrated, true, nil
	}
	if marker, err := store.Get(layoutKey); err != nil {
		return defaults, false, err
	} else if len(marker) == 0 {
		return defaults, false, nil
	}

	out := defaults
	target := reflect.ValueOf(&out).Elem()
	if target.Kind() != reflect.Struct {
		return defaults, false, fmt.Errorf("prefix genesis target must be a struct")
	}
	for i := 0; i < target.NumField(); i++ {
		field := target.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		bz, err := store.Get(fieldKey(field.Name))
		if err != nil {
			return defaults, false, err
		}
		if len(bz) == 0 {
			continue
		}
		if err := json.Unmarshal(bz, target.Field(i).Addr().Interface()); err != nil {
			return defaults, false, fmt.Errorf("read prefix genesis field %s: %w", field.Name, err)
		}
	}
	return out, true, nil
}

// Save writes each exported genesis struct field under a deterministic prefix key.
func Save[T any](ctx context.Context, storeService corestore.KVStoreService, legacyKey []byte, value T) error {
	if storeService == nil {
		return nil
	}
	store := storeService.OpenKVStore(ctx)
	if err := store.Set(layoutKey, []byte("v2")); err != nil {
		return err
	}
	source := reflect.ValueOf(value)
	if source.Kind() == reflect.Pointer {
		source = source.Elem()
	}
	if source.Kind() != reflect.Struct {
		return fmt.Errorf("prefix genesis value must be a struct")
	}
	for i := 0; i < source.NumField(); i++ {
		field := source.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		bz, err := json.Marshal(source.Field(i).Interface())
		if err != nil {
			return fmt.Errorf("write prefix genesis field %s: %w", field.Name, err)
		}
		if err := store.Set(fieldKey(field.Name), bz); err != nil {
			return err
		}
	}
	_ = store.Delete(legacyKey)
	return nil
}

func fieldKey(name string) []byte {
	return []byte("prefix_genesis/" + snake(name))
}

func snake(name string) string {
	var out strings.Builder
	for idx, r := range name {
		if unicode.IsUpper(r) {
			if idx > 0 {
				out.WriteByte('_')
			}
			out.WriteRune(unicode.ToLower(r))
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}
