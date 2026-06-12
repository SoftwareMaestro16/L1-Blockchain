package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const DomainNFTClassID = "anft66:domain"

type DomainRecordV2Status string

const (
	DomainRecordV2Available		DomainRecordV2Status	= "available"
	DomainRecordV2Committed		DomainRecordV2Status	= "committed"
	DomainRecordV2Auction		DomainRecordV2Status	= "auction"
	DomainRecordV2Active		DomainRecordV2Status	= "active"
	DomainRecordV2RenewalWindow	DomainRecordV2Status	= "renewal_window"
	DomainRecordV2Expired		DomainRecordV2Status	= "expired"
	DomainRecordV2GraceLocked	DomainRecordV2Status	= "grace_locked"
	DomainRecordV2Released		DomainRecordV2Status	= "released"
	DomainRecordV2Revoked		DomainRecordV2Status	= "revoked"
)

type DomainRecordV2 struct {
	Name			string
	NameHash		string
	NormalizedName		string
	ParentNameHash		string
	TLD			string
	Owner			sdk.AccAddress
	Resolver		sdk.AccAddress
	ExpiryHeight		uint64
	ExpiryTime		int64
	RenewalStartHeight	uint64
	NFTClassID		string
	NFTItemID		string
	Status			DomainRecordV2Status
	LifecycleEpoch		uint64
	CreatedAtHeight		uint64
	UpdatedAtHeight		uint64
	Version			uint64
	Flags			uint64
}

type DomainRecordV2ValidationContext struct {
	CurrentHeight		uint64
	NFTOwner		sdk.AccAddress
	ResolverDelegates	[]sdk.AccAddress
}

func NewDomainRecordV2FromDomain(domain Domain, status DomainRecordV2Status, expiryTime int64, currentHeight uint64) (DomainRecordV2, error) {
	normalized, err := NormalizeAETDomain(domain.Name)
	if err != nil {
		return DomainRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return DomainRecordV2{}, err
	}
	parentHash, err := DomainRecordV2ParentNameHash(normalized)
	if err != nil {
		return DomainRecordV2{}, err
	}
	renewalStart := uint64(0)
	params := DefaultIdentityParams()
	if domain.ExpiryHeight > params.RenewalWindowBlocks {
		renewalStart = domain.ExpiryHeight - params.RenewalWindowBlocks
	}
	record := DomainRecordV2{
		Name:			normalized,
		NameHash:		nameHash,
		NormalizedName:		normalized,
		ParentNameHash:		parentHash,
		TLD:			DomainTLD,
		Owner:			cloneSpecAddress(domain.Owner),
		ExpiryHeight:		domain.ExpiryHeight,
		ExpiryTime:		expiryTime,
		RenewalStartHeight:	renewalStart,
		NFTClassID:		DomainNFTClassID,
		NFTItemID:		domain.NFTID,
		Status:			status,
		LifecycleEpoch:		currentHeight,
		CreatedAtHeight:	domain.RegisteredHeight,
		UpdatedAtHeight:	domain.UpdatedHeight,
		Version:		1,
		Flags:			0,
	}
	return record, nil
}

func ValidateDomainRecordV2(record DomainRecordV2, ctx DomainRecordV2ValidationContext) error {
	normalized, err := NormalizeAETDomain(record.Name)
	if err != nil {
		return err
	}
	if record.Name != record.NormalizedName || record.Name != normalized {
		return errors.New("identity v2 domain name must equal normalized_name")
	}
	if record.TLD != DomainTLD {
		return fmt.Errorf("identity v2 domain tld must be %q", DomainTLD)
	}
	expectedNameHash, err := DomainRecordV2NameHash(record.NormalizedName)
	if err != nil {
		return err
	}
	if record.NameHash != expectedNameHash {
		return errors.New("identity v2 domain name_hash mismatch")
	}
	expectedParentHash, err := DomainRecordV2ParentNameHash(record.NormalizedName)
	if err != nil {
		return err
	}
	if record.ParentNameHash != expectedParentHash {
		return errors.New("identity v2 domain parent_name_hash mismatch")
	}
	if !IsDomainRecordV2Status(record.Status) {
		return fmt.Errorf("invalid identity v2 domain status %q", record.Status)
	}
	if requiresDomainRecordV2Owner(record.Status) {
		if err := validateSpecAddress("identity v2 domain owner", record.Owner); err != nil {
			return err
		}
	}
	if record.NFTClassID != "" && record.NFTClassID != DomainNFTClassID {
		return fmt.Errorf("identity v2 domain nft_class_id must be %q", DomainNFTClassID)
	}
	if requiresDomainRecordV2NFTAgreement(record.Status) {
		if strings.TrimSpace(record.NFTItemID) == "" {
			return errors.New("identity v2 active domain nft_item_id is required")
		}
		if len(ctx.NFTOwner) == 0 {
			return errors.New("identity v2 nft owner is required for active domain")
		}
		if !addressesEqual(record.Owner, ctx.NFTOwner) {
			return errors.New("identity v2 owner must match nft owner")
		}
	}
	if len(record.Resolver) > 0 {
		if err := validateSpecAddress("identity v2 domain resolver", record.Resolver); err != nil {
			return err
		}
		if !addressesEqual(record.Resolver, record.Owner) && !addressInSet(record.Resolver, ctx.ResolverDelegates) {
			return errors.New("identity v2 resolver must be controlled by owner or authorized delegate")
		}
	}
	if requiresUnexpiredDomainRecordV2(record.Status) && record.ExpiryHeight <= ctx.CurrentHeight {
		return errors.New("identity v2 active domain expiry_height must be greater than current height")
	}
	if record.ExpiryTime < 0 {
		return errors.New("identity v2 domain expiry_time must be non-negative")
	}
	if record.RenewalStartHeight > record.ExpiryHeight && record.ExpiryHeight != 0 {
		return errors.New("identity v2 renewal_start_height must not exceed expiry_height")
	}
	if record.CreatedAtHeight == 0 && requiresDomainRecordV2Owner(record.Status) {
		return errors.New("identity v2 created_at_height is required")
	}
	if record.UpdatedAtHeight < record.CreatedAtHeight {
		return errors.New("identity v2 updated_at_height must not precede created_at_height")
	}
	if record.Version == 0 {
		return errors.New("identity v2 domain version is required")
	}
	return nil
}

func DomainRecordV2NameHash(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return identityHash("identity-v2-domain-name", normalized), nil
}

func DomainRecordV2ParentNameHash(name string) (string, error) {
	parent, found, err := ImmediateParentAETDomain(name)
	if err != nil {
		return "", err
	}
	if !found {
		return "", nil
	}
	return DomainRecordV2NameHash(parent)
}

func ImmediateParentAETDomain(name string) (string, bool, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", false, err
	}
	labelsPart := strings.TrimSuffix(normalized, DomainTLD)
	labels := strings.Split(labelsPart, ".")
	if len(labels) <= 1 {
		return "", false, nil
	}
	return strings.Join(labels[1:], ".") + DomainTLD, true, nil
}

func IsDomainRecordV2Status(status DomainRecordV2Status) bool {
	switch status {
	case DomainRecordV2Available, DomainRecordV2Committed, DomainRecordV2Auction, DomainRecordV2Active, DomainRecordV2RenewalWindow, DomainRecordV2Expired, DomainRecordV2GraceLocked, DomainRecordV2Released, DomainRecordV2Revoked:
		return true
	default:
		return false
	}
}

func requiresDomainRecordV2Owner(status DomainRecordV2Status) bool {
	switch status {
	case DomainRecordV2Committed, DomainRecordV2Auction, DomainRecordV2Active, DomainRecordV2RenewalWindow, DomainRecordV2Expired, DomainRecordV2GraceLocked, DomainRecordV2Revoked:
		return true
	default:
		return false
	}
}

func requiresDomainRecordV2NFTAgreement(status DomainRecordV2Status) bool {
	return status == DomainRecordV2Active || status == DomainRecordV2RenewalWindow
}

func requiresUnexpiredDomainRecordV2(status DomainRecordV2Status) bool {
	return status == DomainRecordV2Active || status == DomainRecordV2RenewalWindow
}

func addressInSet(address sdk.AccAddress, values []sdk.AccAddress) bool {
	for _, value := range values {
		if addressesEqual(address, value) {
			return true
		}
	}
	return false
}
