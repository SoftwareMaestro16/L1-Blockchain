package keeper

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	servicestypes "github.com/sovereign-l1/l1/x/services/types"
)

type Keeper struct {
	genesis	servicestypes.GenesisState
	store	map[string][]byte
}

func NewKeeper() Keeper {
	keeper := Keeper{genesis: servicestypes.DefaultGenesis(), store: map[string][]byte{}}
	keeper.rebuildStore()
	return keeper
}

func (k *Keeper) InitGenesis(gs servicestypes.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.rebuildStore()
	return nil
}

func (k Keeper) ExportGenesis() servicestypes.GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) Params() coretypes.AetraCoreParams {
	return k.genesis.Params
}

func (k Keeper) RegistryState() servicestypes.ServiceRegistryState {
	return k.genesis.Registry
}

func (k Keeper) StoreKeys() []string {
	keys := make([]string, 0, len(k.store))
	for key := range k.store {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (k Keeper) StoreValue(key string) ([]byte, bool) {
	value, found := k.store[key]
	if !found {
		return nil, false
	}
	out := append([]byte(nil), value...)
	return out, true
}

func (k Keeper) ValidateInvariants() error {
	return servicestypes.ValidateRegistryInvariants(k.genesis.Registry)
}

func (k *Keeper) RegisterService(msg servicestypes.MsgRegisterService) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, found := k.genesis.Registry.ServiceDescriptorByID(msg.Descriptor.ServiceID); found {
		return fmt.Errorf("services descriptor %s already exists", msg.Descriptor.ServiceID)
	}
	descriptors := append([]servicestypes.ServiceDescriptor(nil), k.genesis.Registry.Descriptors...)
	descriptors = append(descriptors, coretypes.CanonicalServiceDescriptor(msg.Descriptor))
	return k.replaceRegistry(descriptors, k.genesis.Registry.Providers, k.genesis.Registry.Reputations, k.genesis.Registry.Receipts, msg.Descriptor.UpdatedHeight)
}

func (k *Keeper) UpdateService(msg servicestypes.MsgUpdateService) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	descriptors := append([]servicestypes.ServiceDescriptor(nil), k.genesis.Registry.Descriptors...)
	index, current, found := descriptorByID(descriptors, msg.Descriptor.ServiceID)
	if !found {
		return fmt.Errorf("services descriptor %s not found", msg.Descriptor.ServiceID)
	}
	if current.Owner != msg.Authority {
		return errors.New("services update requires current owner")
	}
	if current.Version != msg.ExpectedVersion {
		return errors.New("services update expected version mismatch")
	}
	next := coretypes.CanonicalServiceDescriptor(msg.Descriptor)
	if next.Version <= current.Version {
		next.Version = current.Version + 1
	}
	next.Interface.Version = next.Version
	next.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(next.Interface)
	if next.UpdatedHeight <= current.UpdatedHeight {
		next.UpdatedHeight = current.UpdatedHeight + 1
	}
	descriptors[index] = next
	return k.replaceRegistry(descriptors, k.genesis.Registry.Providers, k.genesis.Registry.Reputations, k.genesis.Registry.Receipts, next.UpdatedHeight)
}

func (k *Keeper) RegisterInterface(msg servicestypes.MsgRegisterInterface) error {
	next, err := servicestypes.RegisterInterfaceInState(k.genesis.Registry, msg, k.genesis.Registry.UpdatedHeight+1)
	if err != nil {
		return err
	}
	if err := servicestypes.ValidateRegistryInvariants(next); err != nil {
		return err
	}
	k.genesis.Registry = next
	k.rebuildStore()
	return nil
}

func (k *Keeper) UpdateInterface(msg servicestypes.MsgUpdateInterface) error {
	next, err := servicestypes.UpdateInterfaceInState(k.genesis.Registry, msg, k.genesis.Registry.UpdatedHeight+1)
	if err != nil {
		return err
	}
	if err := servicestypes.ValidateRegistryInvariants(next); err != nil {
		return err
	}
	k.genesis.Registry = next
	k.rebuildStore()
	return nil
}

func (k *Keeper) RenewService(msg servicestypes.MsgRenewService) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.mutateService(msg.ServiceID, msg.Authority, msg.ExpectedVersion, func(descriptor *servicestypes.ServiceDescriptor) error {
		if msg.ExpiryHeight <= descriptor.ExpiryHeight {
			return errors.New("services renewal must extend expiry")
		}
		descriptor.ExpiryHeight = msg.ExpiryHeight
		return nil
	})
}

func (k *Keeper) DisableService(msg servicestypes.MsgDisableService) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.mutateService(msg.ServiceID, msg.Authority, msg.ExpectedVersion, func(descriptor *servicestypes.ServiceDescriptor) error {
		descriptor.Enabled = false
		descriptor.Status = coretypes.ServiceStatusDisabled
		return nil
	})
}

func (k *Keeper) TransferService(msg servicestypes.MsgTransferService) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.mutateService(msg.ServiceID, msg.Authority, msg.ExpectedVersion, func(descriptor *servicestypes.ServiceDescriptor) error {
		descriptor.Owner = msg.NewOwner
		return nil
	})
}

func (k *Keeper) BindServiceIdentity(msg servicestypes.MsgBindServiceIdentity) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.mutateService(msg.ServiceID, msg.Authority, msg.ExpectedVersion, func(descriptor *servicestypes.ServiceDescriptor) error {
		descriptor.Discovery.IdentityName = msg.IdentityName
		return nil
	})
}

func (k *Keeper) UnbindServiceIdentity(msg servicestypes.MsgUnbindServiceIdentity) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.mutateService(msg.ServiceID, msg.Authority, msg.ExpectedVersion, func(descriptor *servicestypes.ServiceDescriptor) error {
		if descriptor.Discovery.IdentityName != msg.IdentityName {
			return errors.New("services identity binding mismatch")
		}
		descriptor.Discovery.IdentityName = ""
		return nil
	})
}

func (k *Keeper) RegisterProvider(msg servicestypes.MsgRegisterProvider) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, found := k.genesis.Registry.ServiceDescriptorByID(msg.ServiceID); !found {
		return fmt.Errorf("services provider service %s not found", msg.ServiceID)
	}
	providers := append([]servicestypes.ProviderRecord(nil), k.genesis.Registry.Providers...)
	for _, provider := range providers {
		if provider.ServiceID == msg.ServiceID && provider.Provider.ProviderID == msg.Provider.ProviderID {
			return fmt.Errorf("services provider %s already exists", msg.Provider.ProviderID)
		}
	}
	record, err := coretypes.NewProviderRecord(msg.ServiceID, msg.Provider)
	if err != nil {
		return err
	}
	providers = append(providers, record)
	return k.replaceRegistry(k.genesis.Registry.Descriptors, providers, k.genesis.Registry.Reputations, k.genesis.Registry.Receipts, msg.Provider.UpdatedHeight)
}

func (k *Keeper) UpdateProvider(msg servicestypes.MsgUpdateProvider) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	providers := append([]servicestypes.ProviderRecord(nil), k.genesis.Registry.Providers...)
	for i, provider := range providers {
		if provider.ServiceID == msg.ServiceID && provider.Provider.ProviderID == msg.Provider.ProviderID {
			record, err := coretypes.NewProviderRecord(msg.ServiceID, msg.Provider)
			if err != nil {
				return err
			}
			providers[i] = record
			return k.replaceRegistry(k.genesis.Registry.Descriptors, providers, k.genesis.Registry.Reputations, k.genesis.Registry.Receipts, msg.Provider.UpdatedHeight)
		}
	}
	return fmt.Errorf("services provider %s not found", msg.Provider.ProviderID)
}

func (k *Keeper) StakeProviderCollateral(msg servicestypes.MsgStakeProviderCollateral) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.adjustProviderCollateral(msg.ServiceID, msg.ProviderID, msg.Amount, msg.Height, true)
}

func (k *Keeper) UnstakeProviderCollateral(msg servicestypes.MsgUnstakeProviderCollateral) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	return k.adjustProviderCollateral(msg.ServiceID, msg.ProviderID, msg.Amount, msg.Height, false)
}

func (k *Keeper) AnchorServiceReceipt(msg servicestypes.MsgAnchorServiceReceipt) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	if _, found := k.genesis.Registry.ServiceDescriptorByID(msg.Receipt.ServiceID); !found {
		return fmt.Errorf("services receipt service %s not found", msg.Receipt.ServiceID)
	}
	receipts := append([]servicestypes.ServiceReceipt(nil), k.genesis.Registry.Receipts...)
	for _, receipt := range receipts {
		if receipt.ServiceID == msg.Receipt.ServiceID && receipt.CallID == msg.Receipt.CallID {
			return fmt.Errorf("services receipt %s already anchored", msg.Receipt.CallID)
		}
	}
	receipts = append(receipts, msg.Receipt)
	return k.replaceRegistry(k.genesis.Registry.Descriptors, k.genesis.Registry.Providers, k.genesis.Registry.Reputations, receipts, msg.Receipt.AnchoredHeight)
}

func (k *Keeper) SubmitServiceDispute(msg servicestypes.MsgSubmitServiceDispute) error {
	record, err := servicestypes.NewServiceDisputeRecord(msg)
	if err != nil {
		return err
	}
	for _, existing := range k.genesis.Disputes {
		if existing.DisputeID == record.DisputeID {
			return fmt.Errorf("services dispute %s already exists", record.DisputeID)
		}
	}
	k.genesis.Disputes = append(k.genesis.Disputes, record)
	servicestypes.SortDisputes(k.genesis.Disputes)
	return nil
}

func (k *Keeper) mutateService(serviceID, authority string, expectedVersion uint64, mutate func(*servicestypes.ServiceDescriptor) error) error {
	descriptors := append([]servicestypes.ServiceDescriptor(nil), k.genesis.Registry.Descriptors...)
	index, descriptor, found := descriptorByID(descriptors, serviceID)
	if !found {
		return fmt.Errorf("services descriptor %s not found", serviceID)
	}
	if descriptor.Owner != authority {
		return errors.New("services mutation requires current owner")
	}
	if descriptor.Version != expectedVersion {
		return errors.New("services mutation expected version mismatch")
	}
	if err := mutate(&descriptor); err != nil {
		return err
	}
	descriptor.Version++
	descriptor.Interface.Version = descriptor.Version
	descriptor.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(descriptor.Interface)
	descriptor.UpdatedHeight++
	descriptors[index] = coretypes.CanonicalServiceDescriptor(descriptor)
	return k.replaceRegistry(descriptors, k.genesis.Registry.Providers, k.genesis.Registry.Reputations, k.genesis.Registry.Receipts, descriptor.UpdatedHeight)
}

func (k *Keeper) adjustProviderCollateral(serviceID, providerID, amount string, height uint64, add bool) error {
	providers := append([]servicestypes.ProviderRecord(nil), k.genesis.Registry.Providers...)
	for i, record := range providers {
		if record.ServiceID != serviceID || record.Provider.ProviderID != providerID {
			continue
		}
		next := record.Provider
		value, err := adjustAmountString(next.CollateralAmount, amount, add)
		if err != nil {
			return err
		}
		next.CollateralAmount = value
		next.StakeAmount = value
		next.UpdatedHeight = height
		if next.ExpiryHeight <= next.UpdatedHeight {
			next.ExpiryHeight = next.UpdatedHeight + 1
		}
		next.ProviderHash = coretypes.ComputeFogProviderHash(next)
		providerRecord, err := coretypes.NewProviderRecord(serviceID, next)
		if err != nil {
			return err
		}
		providers[i] = providerRecord
		return k.replaceRegistry(k.genesis.Registry.Descriptors, providers, k.genesis.Registry.Reputations, k.genesis.Registry.Receipts, height)
	}
	return fmt.Errorf("services provider %s not found", providerID)
}

func (k *Keeper) replaceRegistry(descriptors []servicestypes.ServiceDescriptor, providers []servicestypes.ProviderRecord, reputations []servicestypes.ReputationRecord, receipts []servicestypes.ServiceReceipt, height uint64) error {
	if height == 0 {
		height = k.genesis.Registry.UpdatedHeight + 1
	}
	anchors := make([]servicestypes.ServiceAnchor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		anchor, err := coretypes.NewServiceAnchorFromDescriptor(descriptor)
		if err != nil {
			return err
		}
		anchors = append(anchors, anchor)
	}
	next, err := coretypes.NewServiceRegistryState(descriptors, anchors, nil, providers, reputations, receipts, height)
	if err != nil {
		return err
	}
	if err := servicestypes.ValidateRegistryInvariants(next); err != nil {
		return err
	}
	k.genesis.Registry = next
	k.rebuildStore()
	return nil
}

func (k *Keeper) rebuildStore() {
	k.store = map[string][]byte{}
	for _, entry := range k.genesis.Registry.Entries {
		if servicestypes.IsServiceStoreKey(entry.Key) {
			k.store[entry.Key] = []byte(entry.Value)
		}
	}
}

func descriptorByID(descriptors []servicestypes.ServiceDescriptor, serviceID string) (int, servicestypes.ServiceDescriptor, bool) {
	for i, descriptor := range descriptors {
		if descriptor.ServiceID == serviceID {
			return i, descriptor, true
		}
	}
	return 0, servicestypes.ServiceDescriptor{}, false
}

func adjustAmountString(current string, delta string, add bool) (string, error) {
	left, ok := new(big.Int).SetString(current, 10)
	if !ok {
		return "", fmt.Errorf("services invalid current amount %q", current)
	}
	right, ok := new(big.Int).SetString(delta, 10)
	if !ok || right.Sign() <= 0 {
		return "", fmt.Errorf("services invalid delta amount %q", delta)
	}
	if add {
		return new(big.Int).Add(left, right).String(), nil
	}
	if left.Cmp(right) < 0 {
		return "", errors.New("services collateral underflow")
	}
	return new(big.Int).Sub(left, right).String(), nil
}

func cloneGenesis(gs servicestypes.GenesisState) servicestypes.GenesisState {
	gs.Registry.Descriptors = append([]servicestypes.ServiceDescriptor(nil), gs.Registry.Descriptors...)
	gs.Registry.Anchors = append([]servicestypes.ServiceAnchor(nil), gs.Registry.Anchors...)
	gs.Registry.Interfaces = append([]servicestypes.ServiceInterface(nil), gs.Registry.Interfaces...)
	gs.Registry.OwnerIndex = append([]coretypes.ServiceRegistryStateEntry(nil), gs.Registry.OwnerIndex...)
	gs.Registry.NameIndex = append([]coretypes.ServiceRegistryStateEntry(nil), gs.Registry.NameIndex...)
	gs.Registry.IdentityBindings = append([]servicestypes.IdentityServiceBinding(nil), gs.Registry.IdentityBindings...)
	gs.Registry.Providers = append([]servicestypes.ProviderRecord(nil), gs.Registry.Providers...)
	gs.Registry.ExpiryIndex = append([]coretypes.ServiceRegistryStateEntry(nil), gs.Registry.ExpiryIndex...)
	gs.Registry.Reputations = append([]servicestypes.ReputationRecord(nil), gs.Registry.Reputations...)
	gs.Registry.Receipts = append([]servicestypes.ServiceReceipt(nil), gs.Registry.Receipts...)
	gs.Registry.Entries = append([]coretypes.ServiceRegistryStateEntry(nil), gs.Registry.Entries...)
	gs.Disputes = append([]servicestypes.ServiceDisputeRecord(nil), gs.Disputes...)
	return gs
}
