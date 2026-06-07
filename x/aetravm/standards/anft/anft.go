package anft

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	NFTStandardName       = "ANFT-66"
	SBTStandardName       = "ASBT-67"
	MaxNameLength         = 96
	MaxContentRefLength   = 512
	MaxRevokeReasonLength = 256
	MaxBatchMintCount     = uint32(100)
	MaxRoyaltyBasisPoints = uint32(1_000)
	ItemCodeHashLength    = 32
	DefaultVersion        = uint32(1)
	itemDerivationDomain  = "aetra/ANFT-66/item/v1"
)

type ItemKind string

const (
	ItemKindNFT ItemKind = "nft"
	ItemKindSBT ItemKind = "sbt"
)

type Metadata struct {
	Name       string
	Symbol     string
	ContentRef string
}

type RoyaltyPolicy struct {
	Enabled     bool
	Recipient   sdk.AccAddress
	BasisPoints uint32
}

type CollectionState struct {
	Address         sdk.AccAddress
	Admin           sdk.AccAddress
	Metadata        Metadata
	NextItemIndex   uint64
	ItemCodeHash    []byte
	StandardVersion uint32
	MutableMetadata bool
	RoyaltyPolicy   RoyaltyPolicy
	SoulboundOnly   bool
}

type ItemState struct {
	Address           sdk.AccAddress
	CollectionAddress sdk.AccAddress
	Index             uint64
	Owner             sdk.AccAddress
	Metadata          Metadata
	Initialized       bool
	Transferable      bool
	Kind              ItemKind
	ImmutableOwner    sdk.AccAddress
	Authority         sdk.AccAddress
	RevokedAt         int64
	RevokeReason      string
	Destructible      bool
	Destroyed         bool
}

type OwnershipProof struct {
	ItemAddress       sdk.AccAddress
	CollectionAddress sdk.AccAddress
	Owner             sdk.AccAddress
	Index             uint64
	Kind              ItemKind
	Revoked           bool
}

type State struct {
	Collection CollectionState
	Items      map[string]ItemState
}

func NewState(collection CollectionState) (*State, error) {
	if err := collection.Validate(); err != nil {
		return nil, err
	}
	return &State{
		Collection: collection,
		Items:      make(map[string]ItemState),
	}, nil
}

func DeriveItemAddress(collection sdk.AccAddress, index uint64, itemCodeHash []byte) (sdk.AccAddress, error) {
	if err := aetraaddress.RejectZeroAddress("nft collection", collection); err != nil {
		return nil, err
	}
	if len(itemCodeHash) != ItemCodeHashLength {
		return nil, fmt.Errorf("item code hash must be %d bytes", ItemCodeHashLength)
	}
	var indexBytes [8]byte
	binary.BigEndian.PutUint64(indexBytes[:], index)

	h := sha256.New()
	h.Write([]byte(itemDerivationDomain))
	writePart(h.Write, collection)
	writePart(h.Write, indexBytes[:])
	writePart(h.Write, itemCodeHash)
	return sdk.AccAddress(h.Sum(nil)), nil
}

func (c CollectionState) Validate() error {
	if err := aetraaddress.RejectZeroAddress("nft collection", c.Address); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("collection admin", c.Admin); err != nil {
		return err
	}
	if len(c.ItemCodeHash) != ItemCodeHashLength {
		return fmt.Errorf("item code hash must be %d bytes", ItemCodeHashLength)
	}
	if c.StandardVersion == 0 {
		return errors.New("standard version must be positive")
	}
	if err := c.RoyaltyPolicy.Validate(); err != nil {
		return err
	}
	return ValidateMetadata(c.Metadata)
}

func (r RoyaltyPolicy) Validate() error {
	if !r.Enabled {
		if len(r.Recipient) != 0 {
			return errors.New("disabled royalty policy must not set recipient")
		}
		if r.BasisPoints != 0 {
			return errors.New("disabled royalty policy must not set basis points")
		}
		return nil
	}
	if err := aetraaddress.RejectZeroAddress("royalty recipient", r.Recipient); err != nil {
		return err
	}
	if r.BasisPoints > MaxRoyaltyBasisPoints {
		return fmt.Errorf("royalty basis points must be <= %d", MaxRoyaltyBasisPoints)
	}
	return nil
}

func (i ItemState) Validate() error {
	if err := aetraaddress.RejectZeroAddress("nft item", i.Address); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("nft collection", i.CollectionAddress); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("item owner", i.Owner); err != nil {
		return err
	}
	if !i.Initialized {
		return errors.New("item must be initialized")
	}
	if i.Kind != ItemKindNFT && i.Kind != ItemKindSBT {
		return errors.New("item kind must be nft or sbt")
	}
	if i.Kind == ItemKindSBT {
		if i.Transferable {
			return errors.New("SBT item must not be transferable")
		}
		if err := aetraaddress.RejectZeroAddress("SBT immutable owner", i.ImmutableOwner); err != nil {
			return err
		}
		if err := aetraaddress.RejectZeroAddress("SBT authority", i.Authority); err != nil {
			return err
		}
		if !i.Owner.Equals(i.ImmutableOwner) {
			return errors.New("SBT owner must remain immutable")
		}
		if len(strings.TrimSpace(i.RevokeReason)) > MaxRevokeReasonLength {
			return fmt.Errorf("SBT revoke reason length must be <= %d", MaxRevokeReasonLength)
		}
	}
	if i.Kind == ItemKindNFT && !i.Transferable {
		return errors.New("NFT item must be transferable")
	}
	return ValidateMetadata(i.Metadata)
}

func ValidateMetadata(metadata Metadata) error {
	name := strings.TrimSpace(metadata.Name)
	symbol := strings.TrimSpace(metadata.Symbol)
	contentRef := strings.TrimSpace(metadata.ContentRef)
	if name == "" {
		return errors.New("metadata name must be set")
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("metadata name length must be <= %d", MaxNameLength)
	}
	if len(contentRef) > MaxContentRefLength {
		return fmt.Errorf("metadata content reference length must be <= %d", MaxContentRefLength)
	}
	for _, value := range []string{name, symbol} {
		if spoofsNative(value) {
			return errors.New("ANFT-66/ASBT-67 metadata must not spoof native AET/naet metadata")
		}
	}
	return nil
}

func (s *State) MintNFT(caller, owner sdk.AccAddress, metadata Metadata) (ItemState, error) {
	if s.Collection.SoulboundOnly {
		return ItemState{}, errors.New("collection is soulbound-only")
	}
	return s.mint(caller, owner, nil, metadata, ItemKindNFT)
}

func (s *State) MintSBT(caller, owner, authority sdk.AccAddress, metadata Metadata) (ItemState, error) {
	if err := aetraaddress.RejectZeroAddress("SBT authority", authority); err != nil {
		return ItemState{}, err
	}
	return s.mint(caller, owner, authority, metadata, ItemKindSBT)
}

func (s *State) BatchMintNFT(caller sdk.AccAddress, owners []sdk.AccAddress, metadata []Metadata) ([]ItemState, error) {
	if len(owners) == 0 {
		return nil, errors.New("batch mint count must be positive")
	}
	if len(owners) > int(MaxBatchMintCount) {
		return nil, fmt.Errorf("batch mint count must be <= %d", MaxBatchMintCount)
	}
	if len(owners) != len(metadata) {
		return nil, errors.New("batch owners and metadata length mismatch")
	}
	checkpoint := s.snapshot()
	items := make([]ItemState, 0, len(owners))
	for i := range owners {
		item, err := s.MintNFT(caller, owners[i], metadata[i])
		if err != nil {
			s.restore(checkpoint)
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *State) TransferNFT(caller sdk.AccAddress, itemAddress sdk.AccAddress, newOwner sdk.AccAddress) error {
	item, err := s.requireItem(itemAddress)
	if err != nil {
		return err
	}
	if item.Kind != ItemKindNFT {
		return errors.New("SBT transfer must be rejected")
	}
	if !item.Owner.Equals(caller) {
		return errors.New("NFT transfer requires current owner authorization")
	}
	if err := aetraaddress.RejectZeroAddress("new item owner", newOwner); err != nil {
		return err
	}
	item.Owner = append(sdk.AccAddress(nil), newOwner...)
	s.Items[string(item.Address)] = item
	return s.ValidateCollectionMembership()
}

func (s *State) TransferSBT(caller sdk.AccAddress, itemAddress sdk.AccAddress, newOwner sdk.AccAddress) error {
	if _, err := s.requireItem(itemAddress); err != nil {
		return err
	}
	_ = caller
	_ = newOwner
	return errors.New("SBT transfer must be rejected")
}

func (s *State) ChangeCollectionMetadata(caller sdk.AccAddress, metadata Metadata) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	if !s.Collection.MutableMetadata {
		return errors.New("collection metadata is immutable")
	}
	if err := ValidateMetadata(metadata); err != nil {
		return err
	}
	s.Collection.Metadata = metadata
	return s.Collection.Validate()
}

func (s *State) ChangeAdmin(caller, newAdmin sdk.AccAddress) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("new collection admin", newAdmin); err != nil {
		return err
	}
	s.Collection.Admin = append(sdk.AccAddress(nil), newAdmin...)
	return s.Collection.Validate()
}

func (s *State) ItemAddress(index uint64) (sdk.AccAddress, error) {
	return DeriveItemAddress(s.Collection.Address, index, s.Collection.ItemCodeHash)
}

func (s *State) ProveItemBelongsToCollection(itemAddress sdk.AccAddress, index uint64) (bool, error) {
	expected, err := s.ItemAddress(index)
	if err != nil {
		return false, err
	}
	if !expected.Equals(itemAddress) {
		return false, nil
	}
	item, ok := s.Items[string(itemAddress)]
	if !ok {
		return false, errors.New("item is not minted")
	}
	if item.Destroyed {
		return false, errors.New("item is destroyed")
	}
	return item.CollectionAddress.Equals(s.Collection.Address) && item.Index == index, nil
}

func (s *State) ProveSBTOwnership(itemAddress, owner sdk.AccAddress) (OwnershipProof, error) {
	item, err := s.requireItem(itemAddress)
	if err != nil {
		return OwnershipProof{}, err
	}
	if item.Kind != ItemKindSBT {
		return OwnershipProof{}, errors.New("item is not SBT")
	}
	if !item.Owner.Equals(owner) {
		return OwnershipProof{}, errors.New("SBT ownership proof owner mismatch")
	}
	return OwnershipProof{
		ItemAddress:       append(sdk.AccAddress(nil), item.Address...),
		CollectionAddress: append(sdk.AccAddress(nil), item.CollectionAddress...),
		Owner:             append(sdk.AccAddress(nil), item.Owner...),
		Index:             item.Index,
		Kind:              item.Kind,
		Revoked:           item.RevokedAt != 0,
	}, nil
}

func (s *State) RequestCurrentOwner(itemAddress sdk.AccAddress) (sdk.AccAddress, error) {
	item, err := s.requireItem(itemAddress)
	if err != nil {
		return nil, err
	}
	return append(sdk.AccAddress(nil), item.Owner...), nil
}

func (s *State) RevokeSBT(caller sdk.AccAddress, itemAddress sdk.AccAddress, revokedAt int64, reason string) error {
	item, err := s.requireItem(itemAddress)
	if err != nil {
		return err
	}
	if item.Kind != ItemKindSBT {
		return errors.New("item is not SBT")
	}
	if !item.Authority.Equals(caller) {
		return errors.New("only SBT authority can revoke")
	}
	if revokedAt <= 0 {
		return errors.New("revoked_at must be positive")
	}
	if len(strings.TrimSpace(reason)) > MaxRevokeReasonLength {
		return fmt.Errorf("SBT revoke reason length must be <= %d", MaxRevokeReasonLength)
	}
	item.RevokedAt = revokedAt
	item.RevokeReason = strings.TrimSpace(reason)
	s.Items[string(item.Address)] = item
	return s.ValidateCollectionMembership()
}

func (s *State) DestroySBT(caller sdk.AccAddress, itemAddress sdk.AccAddress) error {
	item, err := s.requireItem(itemAddress)
	if err != nil {
		return err
	}
	if item.Kind != ItemKindSBT {
		return errors.New("item is not SBT")
	}
	if !item.Destructible {
		return errors.New("SBT destroy is disabled by policy")
	}
	if !item.Authority.Equals(caller) {
		return errors.New("only SBT authority can destroy")
	}
	item.Destroyed = true
	s.Items[string(item.Address)] = item
	return nil
}

func (s *State) ValidateCollectionMembership() error {
	if err := s.Collection.Validate(); err != nil {
		return err
	}
	for _, item := range s.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		expected, err := s.ItemAddress(item.Index)
		if err != nil {
			return err
		}
		if !expected.Equals(item.Address) {
			return errors.New("item address derivation mismatch")
		}
		if !item.CollectionAddress.Equals(s.Collection.Address) {
			return errors.New("item collection address mismatch")
		}
	}
	return nil
}

func (s *State) mint(caller, owner, authority sdk.AccAddress, metadata Metadata, kind ItemKind) (ItemState, error) {
	if err := s.requireAdmin(caller); err != nil {
		return ItemState{}, err
	}
	if err := aetraaddress.RejectZeroAddress("item owner", owner); err != nil {
		return ItemState{}, err
	}
	if err := ValidateMetadata(metadata); err != nil {
		return ItemState{}, err
	}
	index := s.Collection.NextItemIndex
	address, err := s.ItemAddress(index)
	if err != nil {
		return ItemState{}, err
	}
	if _, exists := s.Items[string(address)]; exists {
		return ItemState{}, errors.New("item index already minted")
	}
	item := ItemState{
		Address:           address,
		CollectionAddress: append(sdk.AccAddress(nil), s.Collection.Address...),
		Index:             index,
		Owner:             append(sdk.AccAddress(nil), owner...),
		Metadata:          metadata,
		Initialized:       true,
		Transferable:      kind == ItemKindNFT,
		Kind:              kind,
	}
	if kind == ItemKindSBT {
		item.ImmutableOwner = append(sdk.AccAddress(nil), owner...)
		item.Authority = append(sdk.AccAddress(nil), authority...)
	}
	if err := item.Validate(); err != nil {
		return ItemState{}, err
	}
	s.Collection.NextItemIndex++
	s.Items[string(address)] = item
	return item, s.ValidateCollectionMembership()
}

func (s *State) requireItem(address sdk.AccAddress) (ItemState, error) {
	if err := aetraaddress.RejectZeroAddress("nft item", address); err != nil {
		return ItemState{}, err
	}
	item, ok := s.Items[string(address)]
	if !ok {
		return ItemState{}, errors.New("item is not minted")
	}
	if item.Destroyed {
		return ItemState{}, errors.New("item is destroyed")
	}
	return item, nil
}

func (s *State) requireAdmin(caller sdk.AccAddress) error {
	if !s.Collection.Admin.Equals(caller) {
		return errors.New("only collection admin can perform this action")
	}
	return nil
}

func (s *State) snapshot() State {
	items := make(map[string]ItemState, len(s.Items))
	for key, value := range s.Items {
		items[key] = value
	}
	return State{
		Collection: s.Collection,
		Items:      items,
	}
}

func (s *State) restore(checkpoint State) {
	s.Collection = checkpoint.Collection
	s.Items = checkpoint.Items
}

func spoofsNative(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	return normalized == strings.ToLower(appparams.TokenName) ||
		normalized == strings.ToLower(appparams.TokenSymbol) ||
		normalized == strings.ToLower(appparams.DisplayDenom) ||
		normalized == strings.ToLower(appparams.BaseDenom)
}

func writePart(write func([]byte) (int, error), bz []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(bz)))
	_, _ = write(length[:])
	_, _ = write(bz)
}
