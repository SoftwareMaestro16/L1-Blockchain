package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DomainRecordV2FlagRestricted = uint64(1 << 0)

type DomainNFTBinding struct {
	NameHash		string
	NFTClassID		string
	NFTItemID		string
	Owner			sdk.AccAddress
	LastVerifiedHeight	uint64
	BindingVersion		uint64
}

type DomainNFTBindingContext struct {
	RegistryOwner	sdk.AccAddress
	NFTModuleOwner	sdk.AccAddress
	CurrentHeight	uint64
}

func NewDomainNFTBinding(name string, nftItemID string, owner sdk.AccAddress, lastVerifiedHeight uint64) (DomainNFTBinding, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return DomainNFTBinding{}, err
	}
	binding := DomainNFTBinding{
		NameHash:		nameHash,
		NFTClassID:		DomainNFTClassID,
		NFTItemID:		nftItemID,
		Owner:			cloneSpecAddress(owner),
		LastVerifiedHeight:	lastVerifiedHeight,
		BindingVersion:		1,
	}
	return binding, ValidateDomainNFTBinding(binding, DomainNFTBindingContext{
		RegistryOwner:	owner,
		NFTModuleOwner:	owner,
		CurrentHeight:	lastVerifiedHeight,
	})
}

func ValidateDomainNFTBinding(binding DomainNFTBinding, ctx DomainNFTBindingContext) error {
	if err := validateHexHash("identity v2 nft binding name hash", binding.NameHash); err != nil {
		return err
	}
	if binding.NFTClassID != DomainNFTClassID {
		return fmt.Errorf("identity v2 nft binding class id must be %q", DomainNFTClassID)
	}
	if binding.NFTItemID == "" {
		return errors.New("identity v2 nft binding item id is required")
	}
	if err := validateSpecAddress("identity v2 nft binding owner", binding.Owner); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 registry owner", ctx.RegistryOwner); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 nft module owner", ctx.NFTModuleOwner); err != nil {
		return err
	}
	if !addressesEqual(ctx.RegistryOwner, ctx.NFTModuleOwner) {
		return errors.New("identity v2 nft binding requires nft module owner to equal registry owner")
	}
	if !addressesEqual(binding.Owner, ctx.RegistryOwner) {
		return errors.New("identity v2 nft binding owner must equal registry owner")
	}
	if binding.LastVerifiedHeight == 0 {
		return errors.New("identity v2 nft binding last_verified_height is required")
	}
	if ctx.CurrentHeight != 0 && binding.LastVerifiedHeight > ctx.CurrentHeight {
		return errors.New("identity v2 nft binding last_verified_height is from the future")
	}
	if binding.BindingVersion == 0 {
		return errors.New("identity v2 nft binding version is required")
	}
	return nil
}

func TransferDomainNFTBindingAtomic(record DomainRecordV2, binding DomainNFTBinding, newOwner sdk.AccAddress, height uint64) (DomainRecordV2, DomainNFTBinding, error) {
	if err := validateSpecAddress("identity v2 nft binding new owner", newOwner); err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	if record.NameHash != binding.NameHash {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 nft binding name_hash mismatch")
	}
	if record.NFTClassID != binding.NFTClassID || record.NFTItemID != binding.NFTItemID {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 nft binding item mismatch")
	}
	nextRecord := record
	nextBinding := binding
	nextRecord.Owner = cloneSpecAddress(newOwner)
	nextRecord.UpdatedAtHeight = height
	nextRecord.Flags &^= DomainRecordV2FlagRestricted
	nextBinding.Owner = cloneSpecAddress(newOwner)
	nextBinding.LastVerifiedHeight = height
	if nextBinding.BindingVersion == 0 {
		nextBinding.BindingVersion = 1
	}
	if err := ValidateDomainNFTBinding(nextBinding, DomainNFTBindingContext{RegistryOwner: nextRecord.Owner, NFTModuleOwner: newOwner, CurrentHeight: height}); err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	return nextRecord, nextBinding, nil
}

func RestrictDomainRecordV2ForBrokenBinding(record DomainRecordV2, binding DomainNFTBinding, nftModuleOwner sdk.AccAddress, height uint64) (DomainRecordV2, error) {
	if record.NameHash != binding.NameHash {
		return DomainRecordV2{}, errors.New("identity v2 broken binding name_hash mismatch")
	}
	if addressesEqual(record.Owner, binding.Owner) && addressesEqual(record.Owner, nftModuleOwner) {
		return DomainRecordV2{}, errors.New("identity v2 nft binding is not broken")
	}
	record.Status = DomainRecordV2GraceLocked
	record.Flags |= DomainRecordV2FlagRestricted
	record.UpdatedAtHeight = height
	record.LifecycleEpoch = height
	return record, nil
}

func RepairDomainNFTBinding(record DomainRecordV2, binding DomainNFTBinding, actor sdk.AccAddress, nftModuleOwner sdk.AccAddress, height uint64) (DomainRecordV2, DomainNFTBinding, error) {
	if record.NameHash != binding.NameHash {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 repair binding name_hash mismatch")
	}
	if err := validateSpecAddress("identity v2 nft binding repair actor", actor); err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	if !addressesEqual(actor, record.Owner) && !addressesEqual(actor, nftModuleOwner) {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 nft binding repair requires registry owner or nft owner")
	}
	if !addressesEqual(record.Owner, nftModuleOwner) {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 nft binding cannot be repaired until registry and nft owner agree")
	}
	nextRecord := record
	nextBinding := binding
	nextRecord.Flags &^= DomainRecordV2FlagRestricted
	if nextRecord.Status == DomainRecordV2GraceLocked {
		nextRecord.Status = DomainRecordV2Active
	}
	nextRecord.UpdatedAtHeight = height
	nextRecord.LifecycleEpoch = height
	nextBinding.Owner = cloneSpecAddress(record.Owner)
	nextBinding.LastVerifiedHeight = height
	if nextBinding.BindingVersion == 0 {
		nextBinding.BindingVersion = 1
	}
	if err := ValidateDomainNFTBinding(nextBinding, DomainNFTBindingContext{RegistryOwner: nextRecord.Owner, NFTModuleOwner: nftModuleOwner, CurrentHeight: height}); err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	return nextRecord, nextBinding, nil
}
