package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	nativeaccount "github.com/sovereign-l1/l1/x/native-account/types"
)

var nextAccountNumberKey = []byte("account/meta/next_account_number")

type Keeper struct {
	storeService	corestore.KVStoreService
	feePolicy	nativeaccount.ActivationFeePolicy
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{storeService: storeService}
}

func (k Keeper) WithActivationFeePolicy(policy nativeaccount.ActivationFeePolicy) Keeper {
	k.feePolicy = policy
	return k
}

func (k Keeper) ActivateAccount(ctx context.Context, msg nativeaccount.MsgActivateAccount) (nativeaccount.AccountActivationResult, error) {
	store, err := newActivationStore(ctx, k)
	if err != nil {
		return nativeaccount.AccountActivationResult{}, err
	}
	service, err := nativeaccount.NewAccountActivationService(store, k.feePolicy)
	if err != nil {
		return nativeaccount.AccountActivationResult{}, err
	}
	result, err := service.ActivateAccount(msg, activationHeight(ctx))
	if err != nil {
		return nativeaccount.AccountActivationResult{}, err
	}
	emitAccountActivated(ctx, result.Event)
	return result, nil
}

func (k Keeper) UpdateAuthPolicy(ctx context.Context, msg nativeaccount.MsgUpdateAuthPolicy) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationBlocked, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgUpdateAuthPolicy(account, msg)
	})
}

func (k Keeper) RotateKey(ctx context.Context, msg nativeaccount.MsgRotateKey) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationBlocked, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgRotateKey(account, msg)
	})
}

func (k Keeper) RecoverAccount(ctx context.Context, msg nativeaccount.MsgRecoverAccount) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationAllowed, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgRecoverAccount(account, msg)
	})
}

func (k Keeper) FreezeAccount(ctx context.Context, msg nativeaccount.MsgFreezeAccount) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationBlocked, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgFreezeAccount(account, msg)
	})
}

func (k Keeper) PayStorageDebt(ctx context.Context, msg nativeaccount.MsgPayStorageDebt) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationAllowed, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgPayStorageDebt(account, msg)
	})
}

func (k Keeper) UnfreezeAccount(ctx context.Context, msg nativeaccount.MsgUnfreezeAccount) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationAllowed, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgUnfreezeAccount(account, msg)
	})
}

func (k Keeper) UpdateAccountMetadata(ctx context.Context, msg nativeaccount.MsgUpdateAccountMetadata) (nativeaccount.Account, error) {
	return k.applyAccountMutation(ctx, msg.AccountUser, frozenMutationBlocked, func(account nativeaccount.Account) (nativeaccount.Account, error) {
		return nativeaccount.ApplyMsgUpdateAccountMetadata(account, msg)
	})
}

func (k Keeper) AccountByUser(ctx context.Context, userAddress string) (nativeaccount.Account, bool, error) {
	key, err := nativeaccount.AccountByUserKey(userAddress)
	if err != nil {
		return nativeaccount.Account{}, false, err
	}
	return k.accountByStorageKey(ctx, []byte(key))
}

func (k Keeper) AccountByRaw(ctx context.Context, rawAddress string) (nativeaccount.Account, bool, error) {
	key, err := nativeaccount.AccountByRawKey(rawAddress)
	if err != nil {
		return nativeaccount.Account{}, false, err
	}
	userAddress, found, err := k.userAddressByIndexKey(ctx, []byte(key))
	if err != nil || !found {
		return nativeaccount.Account{}, found, err
	}
	return k.AccountByUser(ctx, userAddress)
}

func (k Keeper) AccountByReputation(ctx context.Context, reputationID string) (nativeaccount.Account, bool, error) {
	key, err := nativeaccount.AccountByReputationKey(reputationID)
	if err != nil {
		return nativeaccount.Account{}, false, err
	}
	userAddress, found, err := k.userAddressByIndexKey(ctx, []byte(key))
	if err != nil || !found {
		return nativeaccount.Account{}, found, err
	}
	return k.AccountByUser(ctx, userAddress)
}

func (k Keeper) AccountStatus(ctx context.Context, userAddress string) (string, bool, error) {
	account, found, err := k.AccountByUser(ctx, userAddress)
	if err != nil || !found {
		return nativeaccount.AccountStatusInactive, found, err
	}
	return account.Status, true, nil
}

func (k Keeper) AccountsAfter(ctx context.Context, cursor string, limit uint64) ([]nativeaccount.Account, bool, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor != "" {
		pair, err := nativeaccountAddressPairFromUser(cursor)
		if err != nil {
			return nil, false, err
		}
		cursor = pair
	}
	store := k.storeService.OpenKVStore(ctx)
	start := []byte(nativeaccount.AccountByUserPrefix)
	iter, err := store.Iterator(start, types.PrefixEndBytes(start))
	if err != nil {
		return nil, false, err
	}
	defer iter.Close()

	accounts := make([]nativeaccount.Account, 0)
	for ; iter.Valid(); iter.Next() {
		userAddress := strings.TrimPrefix(string(iter.Key()), nativeaccount.AccountByUserPrefix)
		if cursor != "" && userAddress <= cursor {
			continue
		}
		if limit != 0 && uint64(len(accounts)) >= limit {
			return accounts, true, nil
		}
		account, err := decodeAccount(iter.Value())
		if err != nil {
			return nil, false, err
		}
		accounts = append(accounts, account)
	}
	return accounts, false, nil
}

func (k Keeper) SetAccount(ctx context.Context, account nativeaccount.Account) error {
	if err := nativeaccount.ValidateAccountInvariant(account); err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if existing, found, err := k.AccountByUser(ctx, account.AddressUser); err != nil {
		return err
	} else if found {
		if err := deleteAccountIndexes(store, existing); err != nil {
			return err
		}
	}
	if err := ensureUniqueIndex(ctx, k, nativeaccount.AccountByRawKey, account.AddressRaw, account.AddressUser, "raw address"); err != nil {
		return err
	}
	if err := ensureUniqueNumberIndex(ctx, k, account.AccountNumber, account.AddressUser); err != nil {
		return err
	}
	if strings.TrimSpace(account.ReputationID) != "" {
		if err := ensureUniqueIndex(ctx, k, nativeaccount.AccountByReputationKey, account.ReputationID, account.AddressUser, "reputation id"); err != nil {
			return err
		}
	}
	accountKey, err := nativeaccount.AccountByUserKey(account.AddressUser)
	if err != nil {
		return err
	}
	bz, err := json.Marshal(account)
	if err != nil {
		return err
	}
	if err := store.Set([]byte(accountKey), bz); err != nil {
		return err
	}
	if err := setIndex(store, nativeaccount.AccountByRawKey, account.AddressRaw, account.AddressUser); err != nil {
		return err
	}
	if err := store.Set([]byte(nativeaccount.AccountByNumberKey(account.AccountNumber)), []byte(account.AddressUser)); err != nil {
		return err
	}
	if strings.TrimSpace(account.ReputationID) != "" {
		if err := setIndex(store, nativeaccount.AccountByReputationKey, account.ReputationID, account.AddressUser); err != nil {
			return err
		}
	}
	next, err := k.NextAccountNumber(ctx)
	if err != nil {
		return err
	}
	if account.AccountNumber >= next {
		return setNextAccountNumber(store, account.AccountNumber+1)
	}
	return nil
}

type frozenMutationPolicy bool

const (
	frozenMutationBlocked	frozenMutationPolicy	= false
	frozenMutationAllowed	frozenMutationPolicy	= true
)

func (k Keeper) applyAccountMutation(ctx context.Context, userAddress string, frozenPolicy frozenMutationPolicy, apply func(nativeaccount.Account) (nativeaccount.Account, error)) (nativeaccount.Account, error) {
	if apply == nil {
		return nativeaccount.Account{}, fmt.Errorf("native account mutation handler is required")
	}
	account, found, err := k.AccountByUser(ctx, userAddress)
	if err != nil {
		return nativeaccount.Account{}, err
	}
	if !found {
		return nativeaccount.Account{}, fmt.Errorf("inactive account cannot send normal messages")
	}
	if err := validatePersistentMutationStatus(account, frozenPolicy); err != nil {
		return nativeaccount.Account{}, err
	}
	next, err := apply(account)
	if err != nil {
		return nativeaccount.Account{}, err
	}
	if err := k.SetAccount(ctx, next); err != nil {
		return nativeaccount.Account{}, err
	}
	return next, nil
}

func validatePersistentMutationStatus(account nativeaccount.Account, frozenPolicy frozenMutationPolicy) error {
	switch account.Status {
	case nativeaccount.AccountStatusActive, nativeaccount.AccountStatusRecovered:
		return nil
	case nativeaccount.AccountStatusFrozen:
		if frozenPolicy == frozenMutationAllowed {
			return nil
		}
		return fmt.Errorf("frozen account allows only recovery, storage debt payment, and unfreeze")
	case nativeaccount.AccountStatusArchived, nativeaccount.AccountStatusClosed:
		return fmt.Errorf("%s account cannot send normal messages", account.Status)
	default:
		return fmt.Errorf("%s account cannot send normal messages", account.Status)
	}
}

func (k Keeper) NextAccountNumber(ctx context.Context) (uint64, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(nextAccountNumberKey)
	if err != nil || len(bz) == 0 {
		return 1, err
	}
	if len(bz) != 8 {
		return 0, fmt.Errorf("native account next account number is malformed")
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k Keeper) InitGenesis(ctx context.Context, gs nativeaccount.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	cacheCtx, write := sdkCtx.CacheContext()
	if err := nativeaccount.ImportGenesis(accountStore{ctx: cacheCtx, keeper: k}, gs); err != nil {
		return err
	}
	write()
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (nativeaccount.GenesisState, error) {
	next, err := k.NextAccountNumber(ctx)
	if err != nil {
		return nativeaccount.GenesisState{}, err
	}
	accounts := make([]nativeaccount.Account, 0)
	for accountNumber := uint64(1); accountNumber < next; accountNumber++ {
		userAddress, found, err := k.userAddressByIndexKey(ctx, []byte(nativeaccount.AccountByNumberKey(accountNumber)))
		if err != nil {
			return nativeaccount.GenesisState{}, err
		}
		if !found {
			continue
		}
		account, found, err := k.AccountByUser(ctx, userAddress)
		if err != nil {
			return nativeaccount.GenesisState{}, err
		}
		if !found {
			return nativeaccount.GenesisState{}, fmt.Errorf("native account number %d references missing account %s", accountNumber, userAddress)
		}
		accounts = append(accounts, account)
	}
	gs := nativeaccount.GenesisState{Version: nativeaccount.DefaultGenesis().Version, Accounts: nativeaccount.SortAccounts(accounts)}
	if err := gs.Validate(); err != nil {
		return nativeaccount.GenesisState{}, err
	}
	return gs, nil
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	plan := nativeaccount.NativeAccountVersionUpgradePlan()
	if err := nativeaccount.ValidateNativeAccountVersionUpgradePlan(plan); err != nil {
		return err
	}
	_, err := k.NextAccountNumber(ctx)
	return err
}

func (k Keeper) LazyMigrateAccount(ctx context.Context, userAddress string) (nativeaccount.Account, bool, error) {
	return nativeaccount.LazyMigrateAccount(accountStore{ctx: ctx, keeper: k}, userAddress)
}

func (k Keeper) RunBatchedMigrationJob(ctx context.Context, cursor string, limit uint64) (nativeaccount.BatchMigrationResult, error) {
	return nativeaccount.RunBatchedMigrationJob(accountStore{ctx: ctx, keeper: k}, cursor, limit)
}

func (k Keeper) accountByStorageKey(ctx context.Context, key []byte) (nativeaccount.Account, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(key)
	if err != nil || len(bz) == 0 {
		return nativeaccount.Account{}, false, err
	}
	account, err := decodeAccount(bz)
	if err != nil {
		return nativeaccount.Account{}, false, err
	}
	return account, true, nil
}

func (k Keeper) userAddressByIndexKey(ctx context.Context, key []byte) (string, bool, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(key)
	if err != nil || len(bz) == 0 {
		return "", false, err
	}
	return string(bz), true, nil
}

type accountStore struct {
	ctx	context.Context
	keeper	Keeper
}

func newActivationStore(ctx context.Context, keeper Keeper) (accountStore, error) {
	if _, err := keeper.NextAccountNumber(ctx); err != nil {
		return accountStore{}, err
	}
	return accountStore{ctx: ctx, keeper: keeper}, nil
}

func (s accountStore) AccountByUser(userAddress string) (nativeaccount.Account, bool, error) {
	return s.keeper.AccountByUser(s.ctx, userAddress)
}

func (s accountStore) AccountByRaw(rawAddress string) (nativeaccount.Account, bool, error) {
	return s.keeper.AccountByRaw(s.ctx, rawAddress)
}

func (s accountStore) AccountByReputation(reputationID string) (nativeaccount.Account, bool, error) {
	return s.keeper.AccountByReputation(s.ctx, reputationID)
}

func (s accountStore) AccountsAfter(cursor string, limit uint64) ([]nativeaccount.Account, bool, error) {
	return s.keeper.AccountsAfter(s.ctx, cursor, limit)
}

func (s accountStore) SetAccount(account nativeaccount.Account) error {
	return s.keeper.SetAccount(s.ctx, account)
}

func (s accountStore) NextAccountNumber() uint64 {
	next, err := s.keeper.NextAccountNumber(s.ctx)
	if err != nil {
		panic(err)
	}
	return next
}

func decodeAccount(bz []byte) (nativeaccount.Account, error) {
	var account nativeaccount.Account
	if err := json.Unmarshal(bz, &account); err != nil {
		return nativeaccount.Account{}, err
	}
	if err := nativeaccount.ValidateAccountInvariant(account); err != nil {
		return nativeaccount.Account{}, err
	}
	return account, nil
}

func ensureUniqueIndex(ctx context.Context, keeper Keeper, keyFn func(string) (string, error), value, userAddress, label string) error {
	key, err := keyFn(value)
	if err != nil {
		return err
	}
	existing, found, err := keeper.userAddressByIndexKey(ctx, []byte(key))
	if err != nil || !found || existing == userAddress {
		return err
	}
	return fmt.Errorf("duplicate native account %s %s", label, value)
}

func ensureUniqueNumberIndex(ctx context.Context, keeper Keeper, accountNumber uint64, userAddress string) error {
	existing, found, err := keeper.userAddressByIndexKey(ctx, []byte(nativeaccount.AccountByNumberKey(accountNumber)))
	if err != nil || !found || existing == userAddress {
		return err
	}
	return fmt.Errorf("duplicate native account number %d", accountNumber)
}

func setIndex(store corestore.KVStore, keyFn func(string) (string, error), value, userAddress string) error {
	key, err := keyFn(value)
	if err != nil {
		return err
	}
	return store.Set([]byte(key), []byte(userAddress))
}

func deleteAccountIndexes(store corestore.KVStore, account nativeaccount.Account) error {
	rawKey, err := nativeaccount.AccountByRawKey(account.AddressRaw)
	if err != nil {
		return err
	}
	if err := store.Delete([]byte(rawKey)); err != nil {
		return err
	}
	if err := store.Delete([]byte(nativeaccount.AccountByNumberKey(account.AccountNumber))); err != nil {
		return err
	}
	if strings.TrimSpace(account.ReputationID) != "" {
		reputationKey, err := nativeaccount.AccountByReputationKey(account.ReputationID)
		if err != nil {
			return err
		}
		return store.Delete([]byte(reputationKey))
	}
	return nil
}

func setNextAccountNumber(store corestore.KVStore, next uint64) error {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, next)
	return store.Set(nextAccountNumberKey, bz)
}

func nativeaccountAddressPairFromUser(userAddress string) (string, error) {
	key, err := nativeaccount.AccountByUserKey(userAddress)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(key, nativeaccount.AccountByUserPrefix), nil
}

func activationHeight(ctx context.Context) uint64 {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()
	if height <= 0 {
		return 1
	}
	return uint64(height)
}

func emitAccountActivated(ctx context.Context, event nativeaccount.AccountActivatedEvent) {
	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(sdk.NewEvent(
		nativeaccount.EventTypeAccountActivated,
		sdk.NewAttribute("address_user", event.AddressUser),
		sdk.NewAttribute("address_raw", event.AddressRaw),
		sdk.NewAttribute("account_number", fmt.Sprintf("%d", event.AccountNumber)),
		sdk.NewAttribute("sequence", fmt.Sprintf("%d", event.Sequence)),
		sdk.NewAttribute("pubkey_hash", event.PubKeyHash),
		sdk.NewAttribute("fee_paid", fmt.Sprintf("%d", event.FeePaid)),
	))
}
