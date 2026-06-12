package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/identity-root/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version		uint64
	Params		prototype.Params
	IdentityParams	types.IdentityRootParams
	State		types.IdentityRootState
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	state := types.EmptyIdentityRootState()
	state.RootAuthorities = append(state.RootAuthorities, types.RootAuthority{Authority: prototype.DefaultAuthority, Role: "root"})
	return GenesisState{
		Version:	prototype.CurrentGenesisVersion,
		Params:		prototype.DefaultParams(),
		IdentityParams:	types.DefaultIdentityRootParams(),
		State:		state,
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("identity root prototype unsupported genesis version")
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return gs.State.Validate(gs.IdentityParams)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) RegisterName(msg types.MsgRegisterName) (types.NameRecord, error) {
	if err := k.requireEnabled(); err != nil {
		return types.NameRecord{}, err
	}
	if msg.Height == 0 {
		return types.NameRecord{}, errors.New("identity registration height must be positive")
	}
	name, err := types.NormalizeName(msg.Name, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return types.NameRecord{}, err
	}
	root, _ := types.NormalizeRootNamespace(k.genesis.IdentityParams.RootNamespace)
	if name == root {
		return types.NameRecord{}, errors.New("identity root namespace cannot be registered")
	}
	if err := types.ValidateUserFacingAEAddress("identity owner", msg.Owner); err != nil {
		return types.NameRecord{}, err
	}
	if _, _, found := recordIndex(k.genesis.State.Records, name); found {
		return types.NameRecord{}, errors.New("identity name already registered")
	}
	if isReserved(k.genesis.State.ReservedNames, name) && !isRootAuthority(k.genesis.State.RootAuthorities, msg.Owner) {
		return types.NameRecord{}, errors.New("identity reserved name cannot be registered by normal user")
	}
	expiry, err := addHeight(msg.Height, k.genesis.IdentityParams.RegistrationPeriod)
	if err != nil {
		return types.NameRecord{}, err
	}
	parent, err := types.ParentName(name, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return types.NameRecord{}, err
	}
	binding := prepareBinding(name, msg.Owner, msg.NFTBinding, k.genesis.IdentityParams)
	record := types.NameRecord{
		Name:				name,
		ParentName:			parent,
		Owner:				msg.Owner,
		ResolverRoot:			msg.ResolverRoot,
		ExpiryHeight:			expiry,
		RenewalHeight:			msg.Height,
		SubdomainPolicy:		msg.SubdomainPolicy,
		NFTBinding:			binding,
		LastStorageChargeHeight:	msg.Height,
		RentPayerPolicy:		nextDefaultRentPayerPolicy(k.genesis.IdentityParams),
		CreatedHeight:			msg.Height,
		UpdatedHeight:			msg.Height,
	}.Normalize(k.genesis.IdentityParams)
	next := cloneGenesis(k.genesis)
	next.State.Records = append(next.State.Records, record)
	if record.ResolverRoot != types.DefaultResolverRoot {
		next.State.Resolvers = upsertResolver(next.State.Resolvers, types.ResolverRecord{Name: name, ResolverRoot: record.ResolverRoot, UpdatedHeight: msg.Height}, next.IdentityParams)
	}
	if binding.Enabled {
		next.State.NFTBindings = upsertBinding(next.State.NFTBindings, binding, next.IdentityParams)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.NameRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) RenewName(msg types.MsgRenewName) (types.NameRecord, error) {
	if err := k.requireEnabled(); err != nil {
		return types.NameRecord{}, err
	}
	index, record, err := k.requireOwnedName(msg.Name, msg.Owner, msg.Height, false)
	if err != nil {
		return types.NameRecord{}, err
	}
	base := record.ExpiryHeight
	if msg.Height > base {
		base = msg.Height
	}
	expiry, err := addHeight(base, k.genesis.IdentityParams.RenewalPeriod)
	if err != nil {
		return types.NameRecord{}, err
	}
	record, err = accrueDomainRent(record, k.genesis.IdentityParams, msg.Height)
	if err != nil {
		return types.NameRecord{}, err
	}
	record.ExpiryHeight = expiry
	record.RenewalHeight = msg.Height
	record.UpdatedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Records[index] = record.Normalize(next.IdentityParams)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.NameRecord{}, err
	}
	k.genesis = next
	return record.Normalize(k.genesis.IdentityParams), nil
}

func (k *Keeper) TransferName(msg types.MsgTransferName) (types.NameRecord, error) {
	if err := k.requireEnabled(); err != nil {
		return types.NameRecord{}, err
	}
	index, record, err := k.requireOwnedName(msg.Name, msg.Owner, msg.Height, true)
	if err != nil {
		return types.NameRecord{}, err
	}
	if err := types.ValidateUserFacingAEAddress("identity new owner", msg.NewOwner); err != nil {
		return types.NameRecord{}, err
	}
	record, err = accrueDomainRent(record, k.genesis.IdentityParams, msg.Height)
	if err != nil {
		return types.NameRecord{}, err
	}
	binding := prepareBinding(record.Name, msg.NewOwner, msg.NewNFTBinding, k.genesis.IdentityParams)
	record.Owner = msg.NewOwner
	record.NFTBinding = binding
	record.UpdatedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Records[index] = record.Normalize(next.IdentityParams)
	next.State.ReverseRecords = removeReverseByName(next.State.ReverseRecords, record.Name)
	if next.IdentityParams.NFTBindingEnabled {
		next.State.NFTBindings = upsertBinding(next.State.NFTBindings, binding, next.IdentityParams)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.NameRecord{}, err
	}
	k.genesis = next
	return record.Normalize(k.genesis.IdentityParams), nil
}

func (k *Keeper) SetResolver(msg types.MsgSetResolver) (types.ResolverRecord, error) {
	if err := k.requireEnabled(); err != nil {
		return types.ResolverRecord{}, err
	}
	index, record, err := k.requireOwnedName(msg.Name, msg.Owner, msg.Height, true)
	if err != nil {
		return types.ResolverRecord{}, err
	}
	record, err = accrueDomainRent(record, k.genesis.IdentityParams, msg.Height)
	if err != nil {
		return types.ResolverRecord{}, err
	}
	record.ResolverRoot = msg.ResolverRoot
	record.UpdatedHeight = msg.Height
	resolver := types.ResolverRecord{Name: record.Name, ResolverRoot: msg.ResolverRoot, UpdatedHeight: msg.Height}.Normalize(k.genesis.IdentityParams)
	next := cloneGenesis(k.genesis)
	next.State.Records[index] = record.Normalize(next.IdentityParams)
	next.State.Resolvers = upsertResolver(next.State.Resolvers, resolver, next.IdentityParams)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ResolverRecord{}, err
	}
	k.genesis = next
	return resolver, nil
}

func (k *Keeper) SetReverseRecord(msg types.MsgSetReverseRecord) (types.ReverseRecord, error) {
	if err := k.requireEnabled(); err != nil {
		return types.ReverseRecord{}, err
	}
	_, record, err := k.requireOwnedName(msg.Name, msg.Owner, msg.Height, true)
	if err != nil {
		return types.ReverseRecord{}, err
	}
	if err := types.ValidateUserFacingAEAddress("identity reverse address", msg.Address); err != nil {
		return types.ReverseRecord{}, err
	}
	reverse := types.ReverseRecord{Address: msg.Address, Name: record.Name, Owner: record.Owner, UpdatedHeight: msg.Height}.Normalize(k.genesis.IdentityParams)
	next := cloneGenesis(k.genesis)
	next.State.ReverseRecords = upsertReverse(next.State.ReverseRecords, reverse, next.IdentityParams)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ReverseRecord{}, err
	}
	k.genesis = next
	return reverse, nil
}

func (k *Keeper) CreateSubdomain(msg types.MsgCreateSubdomain) (types.NameRecord, error) {
	if err := k.requireEnabled(); err != nil {
		return types.NameRecord{}, err
	}
	_, parent, err := k.requireOwnedName(msg.ParentName, msg.Owner, msg.Height, true)
	if err != nil {
		return types.NameRecord{}, err
	}
	if parent.SubdomainPolicy == types.SubdomainPolicyDisabled {
		return types.NameRecord{}, errors.New("identity parent disables subdomains")
	}
	subOwner := msg.SubdomainOwner
	if subOwner == "" {
		subOwner = msg.Owner
	}
	if err := types.ValidateUserFacingAEAddress("identity subdomain owner", subOwner); err != nil {
		return types.NameRecord{}, err
	}
	if parent.SubdomainPolicy == types.SubdomainPolicyOwnerOnly && subOwner != parent.Owner {
		return types.NameRecord{}, errors.New("identity subdomain ownership must follow parent policy")
	}
	if parent.SubdomainPolicy == types.SubdomainPolicyPublic && !k.genesis.IdentityParams.AllowPublicSubdomains && subOwner != parent.Owner {
		return types.NameRecord{}, errors.New("identity public subdomains are disabled")
	}
	name, err := types.ChildName(msg.Label, parent.Name, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return types.NameRecord{}, err
	}
	if _, _, found := recordIndex(k.genesis.State.Records, name); found {
		return types.NameRecord{}, errors.New("identity subdomain already registered")
	}
	binding := prepareBinding(name, subOwner, msg.NFTBinding, k.genesis.IdentityParams)
	record := types.NameRecord{
		Name:				name,
		ParentName:			parent.Name,
		Owner:				subOwner,
		ResolverRoot:			msg.ResolverRoot,
		ExpiryHeight:			parent.ExpiryHeight,
		RenewalHeight:			msg.Height,
		SubdomainPolicy:		msg.SubdomainPolicy,
		NFTBinding:			binding,
		LastStorageChargeHeight:	msg.Height,
		RentPayerPolicy:		nextDefaultRentPayerPolicy(k.genesis.IdentityParams),
		CreatedHeight:			msg.Height,
		UpdatedHeight:			msg.Height,
	}.Normalize(k.genesis.IdentityParams)
	next := cloneGenesis(k.genesis)
	next.State.Records = append(next.State.Records, record)
	if binding.Enabled {
		next.State.NFTBindings = upsertBinding(next.State.NFTBindings, binding, next.IdentityParams)
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.NameRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) ReserveName(msg types.MsgReserveName) (types.ReservedName, error) {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return types.ReservedName{}, err
	}
	reserved := types.ReservedName{Name: msg.Name, Authority: msg.Authority, Reason: msg.Reason}.Normalize(k.genesis.IdentityParams)
	if isReserved(k.genesis.State.ReservedNames, reserved.Name) {
		return types.ReservedName{}, errors.New("identity name already reserved")
	}
	next := cloneGenesis(k.genesis)
	next.State.ReservedNames = append(next.State.ReservedNames, reserved)
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return types.ReservedName{}, err
	}
	k.genesis = next
	return reserved, nil
}

func (k *Keeper) ReleaseReservedName(msg types.MsgReleaseReservedName) error {
	if err := k.requireAuthority(msg.Authority); err != nil {
		return err
	}
	name, err := types.NormalizeName(msg.Name, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return err
	}
	next := cloneGenesis(k.genesis)
	var removed bool
	next.State.ReservedNames, removed = removeReserved(next.State.ReservedNames, name)
	if !removed {
		return errors.New("identity reserved name not found")
	}
	next.State = next.State.Export()
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) NameRecord(name string) (types.NameRecord, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.NameRecord{}, false, err
	}
	name, err := types.NormalizeName(name, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return types.NameRecord{}, false, err
	}
	_, record, found := recordIndex(k.genesis.State.Records, name)
	return record, found, nil
}

func (k Keeper) ResolveName(name string, height uint64) (types.NameRecord, types.ResolverRecord, bool, error) {
	record, found, err := k.NameRecord(name)
	if err != nil || !found {
		return types.NameRecord{}, types.ResolverRecord{}, false, err
	}
	if !types.IsActive(record, height) {
		return types.NameRecord{}, types.ResolverRecord{}, false, nil
	}
	_, resolver, resolverFound := resolverIndex(k.genesis.State.Resolvers, record.Name)
	if !resolverFound {
		resolver = types.ResolverRecord{Name: record.Name, ResolverRoot: record.ResolverRoot, UpdatedHeight: record.UpdatedHeight}
	}
	return record, resolver.Normalize(k.genesis.IdentityParams), true, nil
}

func (k Keeper) ReverseRecord(address string) (types.ReverseRecord, bool, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return types.ReverseRecord{}, false, err
	}
	_, reverse, found := reverseIndex(k.genesis.State.ReverseRecords, address)
	return reverse, found, nil
}

func (k Keeper) Subdomains(parentName string) ([]types.NameRecord, error) {
	if err := k.genesis.Params.Validate(); err != nil {
		return nil, err
	}
	parentName, err := types.NormalizeName(parentName, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return nil, err
	}
	out := make([]types.NameRecord, 0)
	for _, record := range k.genesis.State.Export().Records {
		if record.ParentName == parentName {
			out = append(out, record)
		}
	}
	types.SortRecords(out)
	return out, nil
}

func (k Keeper) IdentityRootParams() (types.IdentityRootParams, error) {
	if err := k.genesis.IdentityParams.Validate(); err != nil {
		return types.IdentityRootParams{}, err
	}
	return k.genesis.IdentityParams, nil
}

type Migrator struct {
	keeper *Keeper
}

func NewMigrator(k *Keeper) Migrator {
	return Migrator{keeper: k}
}

func (m Migrator) Migrate1to2() error {
	return m.keeper.ExportGenesis().Validate()
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k Keeper) requireEnabled() error {
	return k.genesis.Params.RequireEnabled()
}

func (k Keeper) requireAuthority(authority string) error {
	if err := k.genesis.Params.RequireEnabled(); err != nil {
		return err
	}
	return k.genesis.Params.Authorize(authority)
}

func (k Keeper) requireOwnedName(name, owner string, height uint64, requireActive bool) (int, types.NameRecord, error) {
	if height == 0 {
		return -1, types.NameRecord{}, errors.New("identity message height must be positive")
	}
	name, err := types.NormalizeName(name, k.genesis.IdentityParams.RootNamespace)
	if err != nil {
		return -1, types.NameRecord{}, err
	}
	index, record, found := recordIndex(k.genesis.State.Records, name)
	if !found {
		return -1, types.NameRecord{}, errors.New("identity name not found")
	}
	if record.Owner != owner {
		return -1, types.NameRecord{}, errors.New("identity name operation requires owner")
	}
	if requireActive && !types.IsActive(record, height) {
		return -1, types.NameRecord{}, errors.New("identity expired name cannot be used as active")
	}
	return index, record, nil
}

func prepareBinding(name, owner string, binding types.IdentityNFTBindingReference, params types.IdentityRootParams) types.IdentityNFTBindingReference {
	if !params.NFTBindingEnabled {
		return types.IdentityNFTBindingReference{Name: name}
	}
	binding.Name = name
	binding.Owner = owner
	return binding.Normalize(params)
}

func addHeight(base, delta uint64) (uint64, error) {
	if base > math.MaxUint64-delta {
		return 0, errors.New("identity height overflow")
	}
	return base + delta, nil
}

func nextDefaultRentPayerPolicy(params types.IdentityRootParams) string {
	if types.IsDomainRentPayerPolicy(params.DefaultDomainRentPayerPolicy) {
		return params.DefaultDomainRentPayerPolicy
	}
	return types.DomainRentPayerOwner
}

func accrueDomainRent(record types.NameRecord, params types.IdentityRootParams, height uint64) (types.NameRecord, error) {
	record = record.Normalize(params)
	delta, err := types.DomainStorageRentDelta(record, params, height)
	if err != nil {
		return types.NameRecord{}, err
	}
	if record.RentPayerPolicy == types.DomainRentPayerOwner {
		if record.StorageRentDebt > math.MaxUint64-delta {
			return types.NameRecord{}, errors.New("identity domain storage rent overflow")
		}
		record.StorageRentDebt += delta
	}
	record.LastStorageChargeHeight = height
	return record, nil
}

func recordIndex(records []types.NameRecord, name string) (int, types.NameRecord, bool) {
	for i, record := range records {
		if record.Name == name {
			return i, record, true
		}
	}
	return -1, types.NameRecord{}, false
}

func resolverIndex(records []types.ResolverRecord, name string) (int, types.ResolverRecord, bool) {
	for i, record := range records {
		if record.Name == name {
			return i, record, true
		}
	}
	return -1, types.ResolverRecord{}, false
}

func reverseIndex(records []types.ReverseRecord, address string) (int, types.ReverseRecord, bool) {
	for i, record := range records {
		if record.Address == address {
			return i, record, true
		}
	}
	return -1, types.ReverseRecord{}, false
}

func upsertResolver(records []types.ResolverRecord, resolver types.ResolverRecord, params types.IdentityRootParams) []types.ResolverRecord {
	resolver = resolver.Normalize(params)
	out := append([]types.ResolverRecord(nil), records...)
	if i, _, found := resolverIndex(out, resolver.Name); found {
		out[i] = resolver
	} else {
		out = append(out, resolver)
	}
	types.SortResolvers(out)
	return out
}

func upsertReverse(records []types.ReverseRecord, reverse types.ReverseRecord, params types.IdentityRootParams) []types.ReverseRecord {
	reverse = reverse.Normalize(params)
	out := append([]types.ReverseRecord(nil), records...)
	if i, _, found := reverseIndex(out, reverse.Address); found {
		out[i] = reverse
	} else {
		out = append(out, reverse)
	}
	types.SortReverseRecords(out)
	return out
}

func upsertBinding(bindings []types.IdentityNFTBindingReference, binding types.IdentityNFTBindingReference, params types.IdentityRootParams) []types.IdentityNFTBindingReference {
	binding = binding.Normalize(params)
	out := append([]types.IdentityNFTBindingReference(nil), bindings...)
	for i := range out {
		if out[i].Name == binding.Name {
			out[i] = binding
			types.SortBindings(out)
			return out
		}
	}
	out = append(out, binding)
	types.SortBindings(out)
	return out
}

func removeReverseByName(records []types.ReverseRecord, name string) []types.ReverseRecord {
	out := records[:0]
	for _, record := range records {
		if record.Name != name {
			out = append(out, record)
		}
	}
	return append([]types.ReverseRecord(nil), out...)
}

func isReserved(names []types.ReservedName, name string) bool {
	for _, reserved := range names {
		if reserved.Name == name {
			return true
		}
	}
	return false
}

func isRootAuthority(authorities []types.RootAuthority, authority string) bool {
	for _, root := range authorities {
		if root.Authority == authority {
			return true
		}
	}
	return false
}

func removeReserved(names []types.ReservedName, name string) ([]types.ReservedName, bool) {
	out := make([]types.ReservedName, 0, len(names))
	var removed bool
	for _, reserved := range names {
		if reserved.Name == name {
			removed = true
			continue
		}
		out = append(out, reserved)
	}
	return out, removed
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Export()
	return gs
}
