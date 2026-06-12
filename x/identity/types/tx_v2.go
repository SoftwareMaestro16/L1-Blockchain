package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MaxIdentityTxChainIDBytesV2		= 64
	MaxIdentityTxBatchResolverUpdatesV2	= 32
	MaxIdentityTxBatchRenewDomainsV2	= 64
	MaxIdentityTxAuctionSaltBytesV2		= 128
	MaxIdentityTxRegistrationSaltBytesV2	= 128
)

type IdentitySignerScopeV2 string

const (
	IdentitySignerScopeOwner		IdentitySignerScopeV2	= "owner"
	IdentitySignerScopeRegistration		IdentitySignerScopeV2	= "registration"
	IdentitySignerScopeResolverUpdate	IdentitySignerScopeV2	= "resolver_update"
	IdentitySignerScopeReverseUpdate	IdentitySignerScopeV2	= "reverse_update"
	IdentitySignerScopeSubdomainAdmin	IdentitySignerScopeV2	= "subdomain_admin"
	IdentitySignerScopeDelegationAdmin	IdentitySignerScopeV2	= "delegation_admin"
	IdentitySignerScopeAuctionBidder	IdentitySignerScopeV2	= "auction_bidder"
	IdentitySignerScopeAuctionAdmin		IdentitySignerScopeV2	= "auction_admin"
	IdentitySignerScopeCacheAdmin		IdentitySignerScopeV2	= "cache_admin"
	IdentitySignerScopeBatchAdmin		IdentitySignerScopeV2	= "batch_admin"
)

type IdentityTxAuthV2 struct {
	ChainID				string
	Signer				sdk.AccAddress
	Scope				IdentitySignerScopeV2
	NameNormalizationVersion	uint64
	Nonce				uint64
	Fee				uint64
	StorageCost			uint64
}

type IdentityMsgV2 interface {
	IdentityMessageName() string
	SignerAddress() sdk.AccAddress
	SignerScope() IdentitySignerScopeV2
	ValidateBasic() error
}

type MsgCommitRegistrationV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	CommitmentHash		string
	CommitmentVersion	uint64
	SaltHashOptional	string
}

type MsgRevealRegistrationV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	CommitmentHash		string
	CommitmentVersion	uint64
	Salt			string
}

type MsgRegisterDirectV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	Owner			sdk.AccAddress
	ExpectedRecordVersion	uint64
}

type MsgRenewDomainV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	ExpectedRecordVersion	uint64
}

type MsgTransferDomainV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	NewOwner		sdk.AccAddress
	ExpectedRecordVersion	uint64
}

type MsgSetResolverV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	Resolver		sdk.AccAddress
	ExpectedRecordVersion	uint64
}

type MsgUpdateResolverRecordV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	Patch			ResolverPatch
	ExpectedRecordVersion	uint64
	RecordTTL		uint64
}

type MsgSetReverseRecordV2 struct {
	Auth			IdentityTxAuthV2
	Record			ReverseResolutionRecordV2
	ExpectedRecordVersion	uint64
}

type MsgVerifyReverseRecordV2 struct {
	Auth			IdentityTxAuthV2
	Record			ReverseResolutionRecordV2
	AuthorizedAliasKeys	[]string
	ExpectedRecordVersion	uint64
}

type MsgCreateSubdomainV2 struct {
	Auth			IdentityTxAuthV2
	ParentName		string
	ParentNameHash		string
	Label			string
	ChildOwner		sdk.AccAddress
	ParentControlsRecord	bool
	DelegationType		SubdomainDelegationTypeV2
	ChildExpiryHeight	uint64
	DetachedPaid		bool
	IndependentPayment	bool
	ParentAuthorization	bool
	Ephemeral		bool
	TimeLockedUntilHeight	uint64
	ExpectedParentVersion	uint64
}

type MsgDelegateSubdomainV2 struct {
	Auth			IdentityTxAuthV2
	Delegation		DelegationRecordV2
	ExpectedRecordVersion	uint64
}

type MsgRevokeDelegationV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	Delegate		sdk.AccAddress
	Scope			DelegationScopeV2
	ExpectedRecordVersion	uint64
}

type MsgStartAuctionV2 struct {
	Auth		IdentityTxAuthV2
	Name		string
	NameHash	string
	MinBid		uint64
	FeeSplitID	string
}

type MsgCommitBidV2 struct {
	Auth		IdentityTxAuthV2
	AuctionID	string
	NameHash	string
	CommitmentHash	string
}

type MsgRevealBidV2 struct {
	Auth		IdentityTxAuthV2
	AuctionID	string
	NameHash	string
	Bid		uint64
	Salt		string
	CommitmentHash	string
}

type MsgFinalizeAuctionV2 struct {
	Auth			IdentityTxAuthV2
	AuctionID		string
	NameHash		string
	ExpectedAuctionVersion	uint64
}

type MsgExpireDomainV2 struct {
	Auth			IdentityTxAuthV2
	Name			string
	NameHash		string
	ExpectedRecordVersion	uint64
}

type ResolverBatchUpdateV2 struct {
	Name			string
	NameHash		string
	Patch			ResolverPatch
	ExpectedRecordVersion	uint64
	RecordTTL		uint64
}

type MsgBatchUpdateResolversV2 struct {
	Auth	IdentityTxAuthV2
	Updates	[]ResolverBatchUpdateV2
}

type RenewDomainBatchItemV2 struct {
	Name			string
	NameHash		string
	ExpectedRecordVersion	uint64
}

type MsgBatchRenewDomainsV2 struct {
	Auth		IdentityTxAuthV2
	Renewals	[]RenewDomainBatchItemV2
}

type MsgInvalidateResolutionCacheV2 struct {
	Auth			IdentityTxAuthV2
	NameHash		string
	ResolutionPathHash	string
	SourceVersion		uint64
	ParentEpoch		uint64
	ChildEpoch		uint64
	ExpectedRecordVersion	uint64
}

func ValidateIdentityMsgV2(msg IdentityMsgV2) error {
	if msg == nil {
		return errors.New("identity v2 message is required")
	}
	return msg.ValidateBasic()
}

func (m MsgCommitRegistrationV2) IdentityMessageName() string	{ return "MsgCommitRegistration" }
func (m MsgCommitRegistrationV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgCommitRegistrationV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgCommitRegistrationV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeRegistration, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 tx commitment hash", m.CommitmentHash); err != nil {
		return err
	}
	if m.CommitmentVersion == 0 {
		return errors.New("identity v2 tx commitment_version is required")
	}
	if m.SaltHashOptional != "" {
		return validateHexHash("identity v2 tx salt hash", m.SaltHashOptional)
	}
	return nil
}

func (m MsgRevealRegistrationV2) IdentityMessageName() string	{ return "MsgRevealRegistration" }
func (m MsgRevealRegistrationV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgRevealRegistrationV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgRevealRegistrationV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeRegistration, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 tx commitment hash", m.CommitmentHash); err != nil {
		return err
	}
	if m.CommitmentVersion == 0 {
		return errors.New("identity v2 tx commitment_version is required")
	}
	return validateIdentityTxSaltV2("identity v2 tx registration salt", m.Salt, MaxIdentityTxRegistrationSaltBytesV2)
}

func (m MsgRegisterDirectV2) IdentityMessageName() string		{ return "MsgRegisterDirect" }
func (m MsgRegisterDirectV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgRegisterDirectV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgRegisterDirectV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeRegistration, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 tx direct registration owner", m.Owner); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgRenewDomainV2) IdentityMessageName() string		{ return "MsgRenewDomain" }
func (m MsgRenewDomainV2) SignerAddress() sdk.AccAddress	{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgRenewDomainV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgRenewDomainV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeOwner, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgTransferDomainV2) IdentityMessageName() string		{ return "MsgTransferDomain" }
func (m MsgTransferDomainV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgTransferDomainV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgTransferDomainV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeOwner, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 tx transfer new owner", m.NewOwner); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgSetResolverV2) IdentityMessageName() string		{ return "MsgSetResolver" }
func (m MsgSetResolverV2) SignerAddress() sdk.AccAddress	{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgSetResolverV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgSetResolverV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeResolverUpdate, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 tx resolver", m.Resolver); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgUpdateResolverRecordV2) IdentityMessageName() string	{ return "MsgUpdateResolverRecord" }
func (m MsgUpdateResolverRecordV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgUpdateResolverRecordV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgUpdateResolverRecordV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeResolverUpdate, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if _, err := ResolverPatchKeys(m.Patch); err != nil {
		return err
	}
	if m.RecordTTL == 0 {
		return errors.New("identity v2 tx resolver record_ttl is required")
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgSetReverseRecordV2) IdentityMessageName() string		{ return "MsgSetReverseRecord" }
func (m MsgSetReverseRecordV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgSetReverseRecordV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgSetReverseRecordV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeReverseUpdate, true); err != nil {
		return err
	}
	if err := ValidateReverseResolutionRecordV2Format(m.Record); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgVerifyReverseRecordV2) IdentityMessageName() string	{ return "MsgVerifyReverseRecord" }
func (m MsgVerifyReverseRecordV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgVerifyReverseRecordV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgVerifyReverseRecordV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeReverseUpdate, true); err != nil {
		return err
	}
	if err := ValidateReverseResolutionRecordV2Format(m.Record); err != nil {
		return err
	}
	if !m.Record.Verified {
		return errors.New("identity v2 tx reverse verify requires verified record")
	}
	for _, key := range m.AuthorizedAliasKeys {
		if err := ValidateResolverGrantKey(key); err != nil {
			return err
		}
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgCreateSubdomainV2) IdentityMessageName() string		{ return "MsgCreateSubdomain" }
func (m MsgCreateSubdomainV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgCreateSubdomainV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgCreateSubdomainV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeSubdomainAdmin, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.ParentName, m.ParentNameHash); err != nil {
		return err
	}
	if err := validateDomainLabel(m.Label); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 tx subdomain child owner", m.ChildOwner); err != nil {
		return err
	}
	if m.DelegationType != "" {
		if err := validateSubdomainDelegationTypeV2(m.DelegationType); err != nil {
			return err
		}
	}
	if m.DetachedPaid {
		if m.DelegationType != SubdomainDelegationDetachedPaidV2 {
			return errors.New("identity v2 tx detached subdomain requires detached_paid delegation type")
		}
		if !m.IndependentPayment || !m.ParentAuthorization {
			return errors.New("identity v2 tx detached subdomain requires independent payment and parent authorization")
		}
	}
	if m.Ephemeral && m.DelegationType != SubdomainDelegationEphemeralServiceV2 {
		return errors.New("identity v2 tx ephemeral subdomain requires ephemeral_service delegation type")
	}
	if m.TimeLockedUntilHeight != 0 && m.ChildExpiryHeight != 0 && m.TimeLockedUntilHeight >= m.ChildExpiryHeight {
		return errors.New("identity v2 tx subdomain time lock must end before child expiry")
	}
	return validateExpectedRecordVersionV2(m.ExpectedParentVersion)
}

func (m MsgDelegateSubdomainV2) IdentityMessageName() string	{ return "MsgDelegateSubdomain" }
func (m MsgDelegateSubdomainV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgDelegateSubdomainV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgDelegateSubdomainV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeDelegationAdmin, true); err != nil {
		return err
	}
	if err := ValidateDelegationRecordV2(m.Delegation); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgRevokeDelegationV2) IdentityMessageName() string		{ return "MsgRevokeDelegation" }
func (m MsgRevokeDelegationV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgRevokeDelegationV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgRevokeDelegationV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeDelegationAdmin, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 tx delegation delegate", m.Delegate); err != nil {
		return err
	}
	if err := validateDelegationScopeV2(m.Scope); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgStartAuctionV2) IdentityMessageName() string		{ return "MsgStartAuction" }
func (m MsgStartAuctionV2) SignerAddress() sdk.AccAddress	{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgStartAuctionV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgStartAuctionV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeAuctionAdmin, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	if m.MinBid == 0 {
		return errors.New("identity v2 tx auction min_bid is required")
	}
	return validateAuctionFeeSplitIDV2(m.FeeSplitID)
}

func (m MsgCommitBidV2) IdentityMessageName() string		{ return "MsgCommitBid" }
func (m MsgCommitBidV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgCommitBidV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgCommitBidV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeAuctionBidder, true); err != nil {
		return err
	}
	if err := validateIdentityTxAuctionIDOrNameHashV2(m.AuctionID, m.NameHash); err != nil {
		return err
	}
	return validateHexHash("identity v2 tx bid commitment hash", m.CommitmentHash)
}

func (m MsgRevealBidV2) IdentityMessageName() string		{ return "MsgRevealBid" }
func (m MsgRevealBidV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgRevealBidV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgRevealBidV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeAuctionBidder, true); err != nil {
		return err
	}
	if err := validateIdentityTxAuctionIDOrNameHashV2(m.AuctionID, m.NameHash); err != nil {
		return err
	}
	if m.Bid == 0 {
		return errors.New("identity v2 tx bid amount is required")
	}
	if err := validateIdentityTxSaltV2("identity v2 tx auction salt", m.Salt, MaxIdentityTxAuctionSaltBytesV2); err != nil {
		return err
	}
	return validateHexHash("identity v2 tx bid commitment hash", m.CommitmentHash)
}

func (m MsgFinalizeAuctionV2) IdentityMessageName() string		{ return "MsgFinalizeAuction" }
func (m MsgFinalizeAuctionV2) SignerAddress() sdk.AccAddress		{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgFinalizeAuctionV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgFinalizeAuctionV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeAuctionAdmin, true); err != nil {
		return err
	}
	if err := validateIdentityTxAuctionIDOrNameHashV2(m.AuctionID, m.NameHash); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedAuctionVersion)
}

func (m MsgExpireDomainV2) IdentityMessageName() string		{ return "MsgExpireDomain" }
func (m MsgExpireDomainV2) SignerAddress() sdk.AccAddress	{ return cloneSpecAddress(m.Auth.Signer) }
func (m MsgExpireDomainV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgExpireDomainV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeOwner, true); err != nil {
		return err
	}
	if _, err := validateIdentityTxNameOrHashV2(m.Name, m.NameHash); err != nil {
		return err
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func (m MsgBatchUpdateResolversV2) IdentityMessageName() string	{ return "MsgBatchUpdateResolvers" }
func (m MsgBatchUpdateResolversV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgBatchUpdateResolversV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgBatchUpdateResolversV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeBatchAdmin, true); err != nil {
		return err
	}
	if len(m.Updates) == 0 {
		return errors.New("identity v2 tx resolver batch updates are required")
	}
	if len(m.Updates) > MaxIdentityTxBatchResolverUpdatesV2 {
		return fmt.Errorf("identity v2 tx resolver batch updates must not exceed %d", MaxIdentityTxBatchResolverUpdatesV2)
	}
	seen := map[string]struct{}{}
	for _, update := range m.Updates {
		nameHash, err := validateIdentityTxNameOrHashV2(update.Name, update.NameHash)
		if err != nil {
			return err
		}
		if _, found := seen[nameHash]; found {
			return errors.New("identity v2 tx resolver batch contains duplicate domain")
		}
		seen[nameHash] = struct{}{}
		if _, err := ResolverPatchKeys(update.Patch); err != nil {
			return err
		}
		if update.RecordTTL == 0 {
			return errors.New("identity v2 tx resolver batch record_ttl is required")
		}
		if err := validateExpectedRecordVersionV2(update.ExpectedRecordVersion); err != nil {
			return err
		}
	}
	return nil
}

func (m MsgBatchRenewDomainsV2) IdentityMessageName() string	{ return "MsgBatchRenewDomains" }
func (m MsgBatchRenewDomainsV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgBatchRenewDomainsV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgBatchRenewDomainsV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeBatchAdmin, true); err != nil {
		return err
	}
	if len(m.Renewals) == 0 {
		return errors.New("identity v2 tx batch renewals are required")
	}
	if len(m.Renewals) > MaxIdentityTxBatchRenewDomainsV2 {
		return fmt.Errorf("identity v2 tx batch renewals must not exceed %d", MaxIdentityTxBatchRenewDomainsV2)
	}
	seen := map[string]struct{}{}
	for _, renewal := range m.Renewals {
		nameHash, err := validateIdentityTxNameOrHashV2(renewal.Name, renewal.NameHash)
		if err != nil {
			return err
		}
		if _, found := seen[nameHash]; found {
			return errors.New("identity v2 tx batch renewals contain duplicate domain")
		}
		seen[nameHash] = struct{}{}
		if err := validateExpectedRecordVersionV2(renewal.ExpectedRecordVersion); err != nil {
			return err
		}
	}
	return nil
}

func (m MsgInvalidateResolutionCacheV2) IdentityMessageName() string {
	return "MsgInvalidateResolutionCache"
}
func (m MsgInvalidateResolutionCacheV2) SignerAddress() sdk.AccAddress {
	return cloneSpecAddress(m.Auth.Signer)
}
func (m MsgInvalidateResolutionCacheV2) SignerScope() IdentitySignerScopeV2	{ return m.Auth.Scope }
func (m MsgInvalidateResolutionCacheV2) ValidateBasic() error {
	if err := validateIdentityTxAuthV2(m.Auth, IdentitySignerScopeCacheAdmin, true); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 tx cache name hash", m.NameHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 tx cache path hash", m.ResolutionPathHash); err != nil {
		return err
	}
	if m.SourceVersion == 0 {
		return errors.New("identity v2 tx cache source_version is required")
	}
	return validateExpectedRecordVersionV2(m.ExpectedRecordVersion)
}

func validateIdentityTxAuthV2(auth IdentityTxAuthV2, expectedScope IdentitySignerScopeV2, requireNonce bool) error {
	if err := validateIdentityTxChainIDV2(auth.ChainID); err != nil {
		return err
	}
	if err := validateSpecAddress("identity v2 tx signer", auth.Signer); err != nil {
		return err
	}
	if err := validateIdentitySignerScopeV2(auth.Scope); err != nil {
		return err
	}
	if auth.Scope != expectedScope {
		return fmt.Errorf("identity v2 tx signer scope must be %q", expectedScope)
	}
	if err := ValidateNameNormalizationVersionV2(auth.NameNormalizationVersion); err != nil {
		return err
	}
	if requireNonce && auth.Nonce == 0 {
		return errors.New("identity v2 tx nonce is required")
	}
	if auth.Fee == 0 {
		return errors.New("identity v2 tx fee payment is required")
	}
	if auth.StorageCost == 0 {
		return errors.New("identity v2 tx storage-cost payment is required")
	}
	return nil
}

func validateIdentityTxChainIDV2(chainID string) error {
	if strings.TrimSpace(chainID) == "" {
		return errors.New("identity v2 tx chain_id is required")
	}
	if strings.TrimSpace(chainID) != chainID {
		return errors.New("identity v2 tx chain_id must not have surrounding whitespace")
	}
	if len(chainID) > MaxIdentityTxChainIDBytesV2 {
		return fmt.Errorf("identity v2 tx chain_id must not exceed %d bytes", MaxIdentityTxChainIDBytesV2)
	}
	return validateStorePathSegmentV2("identity v2 tx chain_id", chainID)
}

func validateIdentitySignerScopeV2(scope IdentitySignerScopeV2) error {
	switch scope {
	case IdentitySignerScopeOwner,
		IdentitySignerScopeRegistration,
		IdentitySignerScopeResolverUpdate,
		IdentitySignerScopeReverseUpdate,
		IdentitySignerScopeSubdomainAdmin,
		IdentitySignerScopeDelegationAdmin,
		IdentitySignerScopeAuctionBidder,
		IdentitySignerScopeAuctionAdmin,
		IdentitySignerScopeCacheAdmin,
		IdentitySignerScopeBatchAdmin:
		return nil
	default:
		return fmt.Errorf("invalid identity v2 signer scope %q", scope)
	}
}

func validateIdentityTxNameOrHashV2(name string, nameHash string) (string, error) {
	if name == "" && nameHash == "" {
		return "", errors.New("identity v2 tx normalized name or name_hash is required")
	}
	if nameHash != "" {
		if err := validateHexHash("identity v2 tx name hash", nameHash); err != nil {
			return "", err
		}
	}
	if name == "" {
		return nameHash, nil
	}
	expectedHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return "", err
	}
	if nameHash != "" && nameHash != expectedHash {
		return "", errors.New("identity v2 tx name_hash mismatch")
	}
	return expectedHash, nil
}

func validateIdentityTxAuctionIDOrNameHashV2(auctionID string, nameHash string) error {
	if auctionID == "" && nameHash == "" {
		return errors.New("identity v2 tx auction_id or name_hash is required")
	}
	if auctionID != "" {
		if err := validateHexHash("identity v2 tx auction id", auctionID); err != nil {
			return err
		}
	}
	if nameHash != "" {
		return validateHexHash("identity v2 tx auction name hash", nameHash)
	}
	return nil
}

func validateExpectedRecordVersionV2(version uint64) error {
	if version == 0 {
		return errors.New("identity v2 tx expected record version is required")
	}
	return nil
}

func validateIdentityTxSaltV2(field string, salt string, maxBytes int) error {
	if strings.TrimSpace(salt) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if strings.TrimSpace(salt) != salt {
		return fmt.Errorf("%s must not have surrounding whitespace", field)
	}
	if len(salt) > maxBytes {
		return fmt.Errorf("%s must not exceed %d bytes", field, maxBytes)
	}
	for i := 0; i < len(salt); i++ {
		c := salt[i]
		if c < 0x21 || c > 0x7e {
			return fmt.Errorf("%s contains unsupported character %q", field, c)
		}
	}
	return nil
}
