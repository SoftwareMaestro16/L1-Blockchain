package types

import (
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/x/internal/prototype"
)

type BatchMigrationResult struct {
	Processed	uint64
	Migrated	uint64
	NextCursor	string
	Complete	bool
	ProcessedUsers	[]string
	MigrationScans	uint64
	NormalPathScans	uint64
}

type AccountVersionUpgradePlan struct {
	ModuleName		string
	FromAccountVersion	uint64
	ToAccountVersion	uint64
	LazyMigrationEnabled	bool
	BatchedMigrationEnabled	bool
	RequiresFullBlockScan	bool
	RegisteredHandlers	[]string
}

const nativeAccountUpgradeModuleName = "nativeaccount"

func LazyMigrateAccount(store AccountStore, userAddress string) (Account, bool, error) {
	if store == nil {
		return Account{}, false, errors.New("native account lazy migration store is required")
	}
	account, found, err := store.AccountByUser(userAddress)
	if err != nil || !found {
		return account, found, err
	}
	migrated, err := MigrateAccountIfNeeded(account)
	if err != nil {
		return Account{}, true, err
	}
	if migrated.Version != account.Version {
		if err := store.SetAccount(migrated); err != nil {
			return Account{}, true, err
		}
	}
	return migrated, true, nil
}

func RunBatchedMigrationJob(store AccountStore, cursor string, limit uint64) (BatchMigrationResult, error) {
	if store == nil {
		return BatchMigrationResult{}, errors.New("native account batched migration store is required")
	}
	if limit == 0 || limit > prototype.MaxQueryLimit {
		return BatchMigrationResult{}, fmt.Errorf("native account batched migration limit must be between 1 and %d", prototype.MaxQueryLimit)
	}
	accounts, hasMore, err := store.AccountsAfter(cursor, limit)
	if err != nil {
		return BatchMigrationResult{}, err
	}
	result := BatchMigrationResult{
		Complete:	!hasMore,
		ProcessedUsers:	make([]string, 0, len(accounts)),
		MigrationScans:	1,
	}
	for _, account := range accounts {
		migrated, err := MigrateAccountIfNeeded(account)
		if err != nil {
			return BatchMigrationResult{}, err
		}
		if migrated.Version != account.Version {
			if err := store.SetAccount(migrated); err != nil {
				return BatchMigrationResult{}, err
			}
			result.Migrated++
		}
		result.Processed++
		result.NextCursor = account.AddressUser
		result.ProcessedUsers = append(result.ProcessedUsers, account.AddressUser)
	}
	if len(accounts) == 0 {
		result.NextCursor = cursor
		result.Complete = true
	}
	return result, nil
}

func NativeAccountVersionUpgradePlan() AccountVersionUpgradePlan {
	return AccountVersionUpgradePlan{
		ModuleName:			nativeAccountUpgradeModuleName,
		FromAccountVersion:		AccountVersionV1,
		ToAccountVersion:		AccountVersionV2,
		LazyMigrationEnabled:		true,
		BatchedMigrationEnabled:	true,
		RequiresFullBlockScan:		false,
		RegisteredHandlers: []string{
			"MigrateAccountIfNeeded",
			"MigrateAccountV1ToV2",
			"LazyMigrateAccount",
			"RunBatchedMigrationJob",
		},
	}
}

func ValidateNativeAccountVersionUpgradePlan(plan AccountVersionUpgradePlan) error {
	if plan.ModuleName != nativeAccountUpgradeModuleName {
		return fmt.Errorf("native account upgrade plan module must be %s", nativeAccountUpgradeModuleName)
	}
	if plan.FromAccountVersion != AccountVersionV1 || plan.ToAccountVersion != AccountVersionV2 {
		return errors.New("native account upgrade plan must migrate v1 accounts to v2")
	}
	if !plan.LazyMigrationEnabled {
		return errors.New("native account upgrade plan must enable lazy migration")
	}
	if !plan.BatchedMigrationEnabled {
		return errors.New("native account upgrade plan must expose optional batched migration")
	}
	if plan.RequiresFullBlockScan {
		return errors.New("native account normal block execution must not require a full account scan")
	}
	required := map[string]bool{
		"MigrateAccountIfNeeded":	false,
		"MigrateAccountV1ToV2":		false,
		"LazyMigrateAccount":		false,
		"RunBatchedMigrationJob":	false,
	}
	for _, handler := range plan.RegisteredHandlers {
		if _, found := required[handler]; found {
			required[handler] = true
		}
	}
	for handler, found := range required {
		if !found {
			return fmt.Errorf("native account upgrade plan missing handler %s", handler)
		}
	}
	return nil
}
