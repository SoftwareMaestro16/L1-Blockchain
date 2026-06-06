package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MaxDelegationPermissionsV2       = 16
	MaxDelegationPermissionBytesV2   = 48
	MaxDelegationRecordPrefixBytesV2 = 48
	MaxAuctionFeeSplitIDBytesV2      = 64
)

type DelegationScopeV2 string

const (
	DelegationScopeResolverUpdate        DelegationScopeV2 = "resolver_update"
	DelegationScopeSubdomainCreate       DelegationScopeV2 = "subdomain_create"
	DelegationScopeSubdomainTransfer     DelegationScopeV2 = "subdomain_transfer"
	DelegationScopeServiceRecordUpdate   DelegationScopeV2 = "service_record_update"
	DelegationScopeInterfaceRecordUpdate DelegationScopeV2 = "interface_record_update"
	DelegationScopeRoutingRecordUpdate   DelegationScopeV2 = "routing_record_update"
	DelegationScopeZoneAdmin             DelegationScopeV2 = "zone_admin"
)

type DelegationRecordV2 struct {
	NameHash          string
	Delegate          sdk.AccAddress
	Scope             DelegationScopeV2
	Permissions       []string
	ExpiresAtHeight   uint64
	SubtreeLimit      uint8
	RecordPrefixLimit string
	CreatedAtHeight   uint64
}

type AuctionRecordV2Status string

const (
	AuctionRecordV2Commit    AuctionRecordV2Status = "commit"
	AuctionRecordV2Reveal    AuctionRecordV2Status = "reveal"
	AuctionRecordV2Finalized AuctionRecordV2Status = "finalized"
)

type AuctionRecordV2 struct {
	AuctionID             string
	NameHash              string
	Status                AuctionRecordV2Status
	CommitStartHeight     uint64
	CommitEndHeight       uint64
	RevealStartHeight     uint64
	RevealEndHeight       uint64
	MinBid                uint64
	WinningBid            uint64
	Winner                sdk.AccAddress
	SealedCommitmentsRoot string
	RevealedBidsCount     uint64
	FeeSplitID            string
}

func NewDelegationRecordV2(name string, delegate sdk.AccAddress, scope DelegationScopeV2, permissions []string, expiresAtHeight uint64, subtreeLimit uint8, recordPrefixLimit string, createdAtHeight uint64) (DelegationRecordV2, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return DelegationRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(normalized)
	if err != nil {
		return DelegationRecordV2{}, err
	}
	record := DelegationRecordV2{
		NameHash:          nameHash,
		Delegate:          cloneSpecAddress(delegate),
		Scope:             scope,
		Permissions:       sortStringSet(permissions),
		ExpiresAtHeight:   expiresAtHeight,
		SubtreeLimit:      subtreeLimit,
		RecordPrefixLimit: recordPrefixLimit,
		CreatedAtHeight:   createdAtHeight,
	}
	return record, ValidateDelegationRecordV2(record)
}

func ValidateDelegationRecordV2(record DelegationRecordV2) error {
	if err := validateHexHash("identity v2 delegation name hash", record.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 delegation delegate", record.Delegate); err != nil {
		return err
	}
	if err := validateDelegationScopeV2(record.Scope); err != nil {
		return err
	}
	if record.CreatedAtHeight == 0 {
		return errors.New("identity v2 delegation created_at_height is required")
	}
	if record.ExpiresAtHeight <= record.CreatedAtHeight {
		return errors.New("identity v2 delegation expires_at_height must be after created_at_height")
	}
	if record.SubtreeLimit > MaxResolverLabels {
		return fmt.Errorf("identity v2 delegation subtree_limit must not exceed %d", MaxResolverLabels)
	}
	if err := validateDelegationRecordPrefixLimitV2(record.RecordPrefixLimit); err != nil {
		return err
	}
	return validateDelegationPermissionsV2(record.Permissions)
}

func ValidateDelegationRecordV2Use(record DelegationRecordV2, scope DelegationScopeV2, permission string, recordKey string, subtreeDepth uint8, height uint64) error {
	if err := ValidateDelegationRecordV2(record); err != nil {
		return err
	}
	if height >= record.ExpiresAtHeight {
		return errors.New("identity v2 delegation is expired")
	}
	if record.Scope != scope {
		return errors.New("identity v2 delegation scope mismatch")
	}
	if subtreeDepth > record.SubtreeLimit {
		return errors.New("identity v2 delegation subtree limit exceeded")
	}
	if !delegationPermissionsContainV2(record.Permissions, permission) {
		return fmt.Errorf("identity v2 delegation does not allow permission %q", permission)
	}
	if record.RecordPrefixLimit != "" {
		if err := validateDelegationPermissionV2("identity v2 delegation record key", recordKey); err != nil {
			return err
		}
		if !strings.HasPrefix(recordKey, record.RecordPrefixLimit) {
			return errors.New("identity v2 delegation record prefix limit exceeded")
		}
	}
	return nil
}

func BuildAuctionRecordV2(auction Auction, minBid uint64, feeSplitID string) (AuctionRecordV2, error) {
	if err := validateAuction(auction); err != nil {
		return AuctionRecordV2{}, err
	}
	nameHash, err := DomainRecordV2NameHash(auction.Name)
	if err != nil {
		return AuctionRecordV2{}, err
	}
	record := AuctionRecordV2{
		AuctionID:             identityHash("identity-v2-auction", auction.Name, fmt.Sprintf("%020d", auction.CommitStartHeight)),
		NameHash:              nameHash,
		Status:                auctionRecordStatusFromPhaseV2(auction.Phase),
		CommitStartHeight:     auction.CommitStartHeight,
		CommitEndHeight:       auction.RevealStartHeight,
		RevealStartHeight:     auction.RevealStartHeight,
		RevealEndHeight:       auction.RevealEndHeight,
		MinBid:                minBid,
		WinningBid:            auction.WinningBid,
		Winner:                cloneSpecAddress(auction.Winner),
		SealedCommitmentsRoot: ComputeAuctionSealedCommitmentsRootV2(auction.Commitments),
		RevealedBidsCount:     uint64(len(auction.Reveals)),
		FeeSplitID:            feeSplitID,
	}
	return record, ValidateAuctionRecordV2(record)
}

func ValidateAuctionRecordV2(record AuctionRecordV2) error {
	if err := validateHexHash("identity v2 auction id", record.AuctionID); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 auction name hash", record.NameHash); err != nil {
		return err
	}
	if err := validateAuctionRecordStatusV2(record.Status); err != nil {
		return err
	}
	if record.CommitStartHeight == 0 {
		return errors.New("identity v2 auction commit_start_height is required")
	}
	if record.CommitEndHeight <= record.CommitStartHeight {
		return errors.New("identity v2 auction commit_end_height must be after commit_start_height")
	}
	if record.RevealStartHeight != record.CommitEndHeight {
		return errors.New("identity v2 auction reveal_start_height must equal commit_end_height")
	}
	if record.RevealEndHeight <= record.RevealStartHeight {
		return errors.New("identity v2 auction reveal_end_height must be after reveal_start_height")
	}
	if record.MinBid == 0 {
		return errors.New("identity v2 auction min_bid is required")
	}
	if err := validateHexHash("identity v2 auction sealed commitments root", record.SealedCommitmentsRoot); err != nil {
		return err
	}
	if err := validateAuctionFeeSplitIDV2(record.FeeSplitID); err != nil {
		return err
	}
	if record.Status == AuctionRecordV2Finalized {
		if record.RevealedBidsCount == 0 {
			return errors.New("identity v2 finalized auction requires revealed bids")
		}
		if err := validateSpecAddress("identity v2 auction winner", record.Winner); err != nil {
			return err
		}
		if record.WinningBid < record.MinBid {
			return errors.New("identity v2 auction winning_bid must be at least min_bid")
		}
		return nil
	}
	if len(record.Winner) != 0 {
		return errors.New("identity v2 unfinished auction must not set winner")
	}
	if record.WinningBid != 0 {
		return errors.New("identity v2 unfinished auction must not set winning_bid")
	}
	return nil
}

func ComputeAuctionSealedCommitmentsRootV2(commitments []AuctionCommitment) string {
	if len(commitments) == 0 {
		return identityHash("identity-v2-auction-commitments-empty")
	}
	ordered := cloneAuctionCommitments(commitments)
	sortAuctionCommitments(ordered)
	parts := []string{"identity-v2-auction-commitments-root", fmt.Sprintf("%020d", len(ordered))}
	for _, commitment := range ordered {
		parts = append(parts,
			commitment.Name,
			string(commitment.Bidder),
			commitment.CommitmentHash,
			fmt.Sprintf("%020d", commitment.CommitHeight),
		)
	}
	return identityHash(parts...)
}

func auctionRecordStatusFromPhaseV2(phase AuctionPhase) AuctionRecordV2Status {
	switch phase {
	case AuctionPhaseCommit:
		return AuctionRecordV2Commit
	case AuctionPhaseReveal:
		return AuctionRecordV2Reveal
	case AuctionPhaseFinalized:
		return AuctionRecordV2Finalized
	default:
		return AuctionRecordV2Status(phase)
	}
}

func validateDelegationScopeV2(scope DelegationScopeV2) error {
	switch scope {
	case DelegationScopeResolverUpdate,
		DelegationScopeSubdomainCreate,
		DelegationScopeSubdomainTransfer,
		DelegationScopeServiceRecordUpdate,
		DelegationScopeInterfaceRecordUpdate,
		DelegationScopeRoutingRecordUpdate,
		DelegationScopeZoneAdmin:
		return nil
	default:
		return fmt.Errorf("invalid identity v2 delegation scope %q", scope)
	}
}

func validateDelegationPermissionsV2(permissions []string) error {
	if len(permissions) == 0 {
		return errors.New("identity v2 delegation permissions are required")
	}
	if len(permissions) > MaxDelegationPermissionsV2 {
		return fmt.Errorf("identity v2 delegation permissions must not exceed %d", MaxDelegationPermissionsV2)
	}
	seen := map[string]struct{}{}
	for i, permission := range permissions {
		if err := validateDelegationPermissionV2("identity v2 delegation permission", permission); err != nil {
			return err
		}
		if _, found := seen[permission]; found {
			return fmt.Errorf("duplicate identity v2 delegation permission %q", permission)
		}
		seen[permission] = struct{}{}
		if i > 0 && permissions[i-1] >= permission {
			return errors.New("identity v2 delegation permissions must be sorted canonically")
		}
	}
	return nil
}

func validateDelegationPermissionV2(field string, value string) error {
	if value == "*" {
		return nil
	}
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > MaxDelegationPermissionBytesV2 {
		return fmt.Errorf("%s must not exceed %d bytes", field, MaxDelegationPermissionBytesV2)
	}
	return ValidateResolverMetadataKey(value)
}

func validateDelegationRecordPrefixLimitV2(value string) error {
	if value == "" {
		return nil
	}
	if len(value) > MaxDelegationRecordPrefixBytesV2 {
		return fmt.Errorf("identity v2 delegation record_prefix_limit must not exceed %d bytes", MaxDelegationRecordPrefixBytesV2)
	}
	if value == "*" {
		return errors.New("identity v2 delegation record_prefix_limit must not be wildcard")
	}
	return ValidateResolverMetadataKey(value)
}

func delegationPermissionsContainV2(permissions []string, permission string) bool {
	if err := validateDelegationPermissionV2("identity v2 delegation requested permission", permission); err != nil {
		return false
	}
	for _, candidate := range permissions {
		if candidate == "*" || candidate == permission {
			return true
		}
	}
	return false
}

func validateAuctionRecordStatusV2(status AuctionRecordV2Status) error {
	switch status {
	case AuctionRecordV2Commit, AuctionRecordV2Reveal, AuctionRecordV2Finalized:
		return nil
	default:
		return fmt.Errorf("invalid identity v2 auction status %q", status)
	}
}

func validateAuctionFeeSplitIDV2(value string) error {
	if value == "" {
		return errors.New("identity v2 auction fee_split_id is required")
	}
	if len(value) > MaxAuctionFeeSplitIDBytesV2 {
		return fmt.Errorf("identity v2 auction fee_split_id must not exceed %d bytes", MaxAuctionFeeSplitIDBytesV2)
	}
	return ValidateResolverMetadataKey(value)
}
