package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	MaxDomainLabels		= 8
	MaxDomainFullBytes	= 253

	DefaultRegistrationPeriodBlocks	= uint64(365 * 24 * 60 * 6)
	DefaultRenewalWindowBlocks	= uint64(30 * 24 * 60 * 6)
	DefaultCommitTTLBlocks		= uint64(24 * 60 * 6)
	DefaultAuctionCommitBlocks	= uint64(24 * 60 * 6)
	DefaultAuctionRevealBlocks	= uint64(24 * 60 * 6)

	ResolverKeyZoneEndpoint	= "zone_endpoint"
)

type DomainLifecycleStatus string

const (
	DomainLifecycleAvailable	DomainLifecycleStatus	= "available"
	DomainLifecycleCommitted	DomainLifecycleStatus	= "committed"
	DomainLifecycleActive		DomainLifecycleStatus	= "active"
	DomainLifecycleRenewalWindow	DomainLifecycleStatus	= "renewal_window"
	DomainLifecycleExpired		DomainLifecycleStatus	= "expired"
)

type DomainOwner struct {
	Address sdk.AccAddress
}

type Domain struct {
	Name			string
	Owner			sdk.AccAddress
	NFTID			string
	RegisteredHeight	uint64
	ExpiryHeight		uint64
	UpdatedHeight		uint64
	ParentName		string
	ParentControlsRecord	bool
}

type DomainNFT struct {
	ID		string
	Domain		string
	Owner		sdk.AccAddress
	MintHeight	uint64
	TransferHeight	uint64
}

type DomainCommit struct {
	Name		string
	Owner		sdk.AccAddress
	CommitmentHash	string
	CommitHeight	uint64
	ExpiresHeight	uint64
}

type UsedDomainCommitment struct {
	CommitmentHash		string
	Name			string
	Owner			sdk.AccAddress
	RevealedHeight		uint64
	ExpiresHeight		uint64
	ChainID			string
	ModuleName		string
	ModuleVersion		uint64
	RegistrationClass	string
	MaxPrice		string
}

type ReverseRecord = ReverseResolverRecord

type SubdomainRecord struct {
	ParentName		string
	Name			string
	Owner			sdk.AccAddress
	ParentControlsRecord	bool
	CreatedHeight		uint64
	DelegationType		SubdomainDelegationTypeV2
	Detached		bool
	Ephemeral		bool
	ExpiryHeight		uint64
	TimeLockedUntilHeight	uint64
	ParentAuthorized	bool
}

type ResolverUpdateIntent struct {
	Domain	string
	Actor	sdk.AccAddress
	Nonce	uint64
}

type IdentityParams struct {
	RegistrationPeriodBlocks	uint64
	RenewalWindowBlocks		uint64
	CommitTTLBlocks			uint64
	AuctionCommitBlocks		uint64
	AuctionRevealBlocks		uint64
}

type AuctionPhase string

const (
	AuctionPhaseCommit	AuctionPhase	= "commit"
	AuctionPhaseReveal	AuctionPhase	= "reveal"
	AuctionPhaseFinalized	AuctionPhase	= "finalized"
)

type AuctionCommitment struct {
	Name		string
	Bidder		sdk.AccAddress
	CommitmentHash	string
	CommitHeight	uint64
}

type AuctionReveal struct {
	Name		string
	Bidder		sdk.AccAddress
	Bid		uint64
	Salt		string
	RevealHeight	uint64
	CommitmentHash	string
}

type AuctionRefundReceipt struct {
	ReceiptID	string
	Name		string
	Bidder		sdk.AccAddress
	Amount		uint64
	CommitmentHash	string
	Reason		string
}

type Auction struct {
	Name			string
	CommitStartHeight	uint64
	RevealStartHeight	uint64
	RevealEndHeight		uint64
	Phase			AuctionPhase
	Commitments		[]AuctionCommitment
	Reveals			[]AuctionReveal
	Winner			sdk.AccAddress
	WinningBid		uint64
	WinningCommitment	string
	Refunds			[]AuctionRefundReceipt
}

type IdentityState struct {
	Params			IdentityParams
	Domains			[]Domain
	DomainNFTs		[]DomainNFT
	Commits			[]DomainCommit
	UsedCommitments		[]UsedDomainCommitment
	Resolvers		[]ResolverRecord
	ReverseRecords		[]ReverseRecord
	Subdomains		[]SubdomainRecord
	Auctions		[]Auction
	PendingResolverUpdates	[]ResolverUpdateIntent
}
