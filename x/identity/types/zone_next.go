package types

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type IdentityShardRoutingMode string
type IdentityLookupTargetType string
type IdentityResolutionStatus string

const (
	IdentityRouteNameHash	IdentityShardRoutingMode	= "name_hash"
	IdentityRouteAddress	IdentityShardRoutingMode	= "address_hash"
	IdentityRouteAuction	IdentityShardRoutingMode	= "auction_id"

	IdentityLookupTargetResolver	IdentityLookupTargetType	= "resolver"
	IdentityLookupTargetReverse	IdentityLookupTargetType	= "reverse"
	IdentityLookupTargetOwner	IdentityLookupTargetType	= "owner"
	IdentityLookupTargetAccount	IdentityLookupTargetType	= "account"
	IdentityLookupTargetContract	IdentityLookupTargetType	= "contract"
	IdentityLookupTargetService	IdentityLookupTargetType	= "service"
	IdentityLookupTargetPayment	IdentityLookupTargetType	= "payment_route"
	IdentityLookupTargetMetadata	IdentityLookupTargetType	= "metadata"

	IdentityResolutionStatusResolved	IdentityResolutionStatus	= "resolved"
	IdentityResolutionStatusNotFound	IdentityResolutionStatus	= "not_found"
	IdentityResolutionStatusExpired		IdentityResolutionStatus	= "expired"
	IdentityResolutionStatusUnauthorized	IdentityResolutionStatus	= "unauthorized"
	IdentityResolutionStatusFailed		IdentityResolutionStatus	= "failed"
	IdentityResolutionStatusRejected	IdentityResolutionStatus	= "rejected"
)

type IdentityShardRoute struct {
	ZoneID		string
	LayoutEpoch	uint64
	ShardCount	uint32
	ShardID		uint32
	RoutingMode	IdentityShardRoutingMode
	RouteKey	string
	StateKey	string
	RouteHash	string
}

type MsgResolveIdentity struct {
	RequestID	string
	Requester	string
	SourceZoneID	string
	TargetName	string
	TargetType	IdentityLookupTargetType
	ProofRequired	bool
	ReplyTo		string
	ExpiryHeight	uint64
	MessageHash	string
}

type MsgIdentityResolutionResult struct {
	RequestID		string
	Name			string
	TargetType		IdentityLookupTargetType
	ResolvedValue		string
	ResolverRecordVersion	uint64
	ProofHashOptional	string
	Status			IdentityResolutionStatus
	ExpiryHeight		uint64
	ResultHash		string
}

type IdentityReverseLookupProofResponse struct {
	Height		uint64
	AddressHash	string
	ProofIndex	IdentityZoneProofIndexEntry
	Proof		IdentityInclusionProof
	ReceiptHash	string
}

type IdentityStoreV2NameHashLayout struct {
	StorePrefix		string
	DomainPrefix		string
	ResolverPrefix		string
	ReversePrefix		string
	DelegationPrefix	string
	AuctionPrefix		string
	PrimaryShardKey		string
	LayoutHash		string
}

func BuildIdentityZoneProofRoots(height uint64, roots IdentityZoneRoots) ([]coretypes.ProofRoot, error) {
	if roots.Height != height {
		return nil, errors.New("identity zone proof root height mismatch")
	}
	if err := roots.Validate(); err != nil {
		return nil, err
	}
	items := []coretypes.ProofRoot{
		{Height: height, ZoneID: coretypes.ZoneIDIdentity, RootType: coretypes.IdentityProofRootType, RootHash: roots.StateRoot, Source: "identity.zone.state"},
		{Height: height, ZoneID: coretypes.ZoneIDIdentity, RootType: coretypes.ResolverProofRootType, RootHash: roots.ResolverRoot, Source: "identity.zone.resolver"},
		{Height: height, ZoneID: coretypes.ZoneIDIdentity, RootType: coretypes.RootType(IdentityProofReverse), RootHash: roots.ReverseRoot, Source: "identity.zone.reverse"},
		{Height: height, ZoneID: coretypes.ZoneIDIdentity, RootType: coretypes.RootType(IdentityProofNFTBinding), RootHash: roots.NFTBindingRoot, Source: "identity.zone.nft"},
		{Height: height, ZoneID: coretypes.ZoneIDIdentity, RootType: coretypes.RootType(IdentityProofAuction), RootHash: roots.AuctionRoot, Source: "identity.zone.auction"},
	}
	sort.SliceStable(items, func(i, j int) bool {
		return string(items[i].RootType) < string(items[j].RootType)
	})
	for _, root := range items {
		if err := root.Validate(); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func RouteIdentityDomainShard(name string, shardCount uint32, layoutEpoch uint64) (IdentityShardRoute, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return IdentityShardRoute{}, err
	}
	key, err := IdentityStoreV2SpecDomainKeyByHash(nameHash)
	if err != nil {
		return IdentityShardRoute{}, err
	}
	return routeIdentityStateKey(IdentityRouteNameHash, nameHash, key, shardCount, layoutEpoch)
}

func RouteIdentityResolverShard(name string, shardCount uint32, layoutEpoch uint64) (IdentityShardRoute, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return IdentityShardRoute{}, err
	}
	key, err := IdentityStoreV2SpecResolverKey(name)
	if err != nil {
		return IdentityShardRoute{}, err
	}
	return routeIdentityStateKey(IdentityRouteNameHash, nameHash, key, shardCount, layoutEpoch)
}

func RouteIdentityReverseShard(address sdk.AccAddress, shardCount uint32, layoutEpoch uint64) (IdentityShardRoute, error) {
	key, err := IdentityStoreV2SpecReverseKey(address)
	if err != nil {
		return IdentityShardRoute{}, err
	}
	addressHash := identityHash("identity-zone-reverse-route", hex.EncodeToString(address))
	return routeIdentityStateKey(IdentityRouteAddress, addressHash, key, shardCount, layoutEpoch)
}

func RouteIdentityAuctionShard(auctionID string, shardCount uint32, layoutEpoch uint64) (IdentityShardRoute, error) {
	key, err := IdentityStoreV2SpecAuctionKey(auctionID)
	if err != nil {
		return IdentityShardRoute{}, err
	}
	return routeIdentityStateKey(IdentityRouteAuction, auctionID, key, shardCount, layoutEpoch)
}

func NewMsgResolveIdentity(msg MsgResolveIdentity) (MsgResolveIdentity, error) {
	if msg.MessageHash != "" {
		return MsgResolveIdentity{}, errors.New("identity resolve message hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return MsgResolveIdentity{}, err
	}
	msg.MessageHash = ComputeMsgResolveIdentityHash(msg)
	return msg, msg.Validate()
}

func (m MsgResolveIdentity) ValidateFormat() error {
	if m.RequestID == "" || m.Requester == "" || m.ReplyTo == "" {
		return errors.New("identity resolve message request, requester, and reply route are required")
	}
	if err := coretypes.ValidateZoneID(coretypes.ZoneID(m.SourceZoneID)); err != nil {
		return err
	}
	if _, err := NormalizeAETDomain(m.TargetName); err != nil {
		return err
	}
	if !IsIdentityLookupTargetType(m.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", m.TargetType)
	}
	if m.ExpiryHeight == 0 {
		return errors.New("identity resolve message expiry height must be positive")
	}
	if m.MessageHash != "" {
		return validateHexHash("identity resolve message hash", m.MessageHash)
	}
	return nil
}

func (m MsgResolveIdentity) Validate() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.MessageHash != ComputeMsgResolveIdentityHash(m) {
		return errors.New("identity resolve message hash mismatch")
	}
	return nil
}

func NewMsgIdentityResolutionResult(msg MsgIdentityResolutionResult) (MsgIdentityResolutionResult, error) {
	if msg.ResultHash != "" {
		return MsgIdentityResolutionResult{}, errors.New("identity resolution result hash must be empty before construction")
	}
	if err := msg.ValidateFormat(); err != nil {
		return MsgIdentityResolutionResult{}, err
	}
	msg.ResultHash = ComputeMsgIdentityResolutionResultHash(msg)
	return msg, msg.Validate()
}

func (m MsgIdentityResolutionResult) ValidateFormat() error {
	if m.RequestID == "" {
		return errors.New("identity resolution result request id is required")
	}
	if _, err := NormalizeAETDomain(m.Name); err != nil {
		return err
	}
	if !IsIdentityLookupTargetType(m.TargetType) {
		return fmt.Errorf("unknown identity lookup target type %q", m.TargetType)
	}
	if !IsIdentityResolutionStatus(m.Status) {
		return fmt.Errorf("unknown identity resolution status %q", m.Status)
	}
	if m.Status == IdentityResolutionStatusResolved && m.ResolvedValue == "" {
		return errors.New("resolved identity result requires value")
	}
	if m.ProofHashOptional != "" {
		if err := validateHexHash("identity resolution result proof hash", m.ProofHashOptional); err != nil {
			return err
		}
	}
	if m.ExpiryHeight == 0 {
		return errors.New("identity resolution result expiry height must be positive")
	}
	if m.ResultHash != "" {
		return validateHexHash("identity resolution result hash", m.ResultHash)
	}
	return nil
}

func (m MsgIdentityResolutionResult) Validate() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.ResultHash != ComputeMsgIdentityResolutionResultHash(m) {
		return errors.New("identity resolution result hash mismatch")
	}
	return nil
}

func BuildProofBackedIdentityReverseLookup(state IdentityState, address sdk.AccAddress, height uint64) (IdentityReverseLookupProofResponse, error) {
	entry, proof, err := QueryIdentityZoneReverseLookupProof(state, address, height)
	if err != nil {
		return IdentityReverseLookupProofResponse{}, err
	}
	response := IdentityReverseLookupProofResponse{
		Height:		height,
		AddressHash:	identityHash("identity-reverse-proof-address", hex.EncodeToString(address)),
		ProofIndex:	entry,
		Proof:		proof,
	}
	response.ReceiptHash = ComputeIdentityReverseLookupProofResponseHash(response)
	return response, response.Validate()
}

func (r IdentityReverseLookupProofResponse) Validate() error {
	if r.Height == 0 {
		return errors.New("identity reverse proof response height must be positive")
	}
	if err := validateHexHash("identity reverse proof response address hash", r.AddressHash); err != nil {
		return err
	}
	if err := r.ProofIndex.Validate(); err != nil {
		return err
	}
	if err := VerifyIdentityProof(r.Proof); err != nil {
		return err
	}
	if r.ProofIndex.Height != r.Height || r.ProofIndex.ProofKind != IdentityProofReverse {
		return errors.New("identity reverse proof response index mismatch")
	}
	if r.ProofIndex.ProofHash != r.Proof.LeafHash {
		return errors.New("identity reverse proof response proof hash mismatch")
	}
	if r.ReceiptHash != ComputeIdentityReverseLookupProofResponseHash(r) {
		return errors.New("identity reverse proof response receipt hash mismatch")
	}
	return nil
}

func DefaultIdentityStoreV2NameHashLayout() (IdentityStoreV2NameHashLayout, error) {
	layout := IdentityStoreV2NameHashLayout{
		StorePrefix:		IdentityStoreV2Prefix,
		DomainPrefix:		IdentityStoreV2SpecDomainsPrefix,
		ResolverPrefix:		IdentityStoreV2SpecResolversPrefix,
		ReversePrefix:		IdentityStoreV2SpecReversePrefix,
		DelegationPrefix:	IdentityStoreV2SpecDelegationsPrefix,
		AuctionPrefix:		IdentityStoreV2SpecAuctionsPrefix,
		PrimaryShardKey:	IdentityZoneShardKey,
	}
	layout.LayoutHash = ComputeIdentityStoreV2NameHashLayoutHash(layout)
	return layout, layout.Validate()
}

func (l IdentityStoreV2NameHashLayout) Validate() error {
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "identity store v2 prefix", value: l.StorePrefix},
		{name: "identity store v2 domain prefix", value: l.DomainPrefix},
		{name: "identity store v2 resolver prefix", value: l.ResolverPrefix},
		{name: "identity store v2 reverse prefix", value: l.ReversePrefix},
		{name: "identity store v2 delegation prefix", value: l.DelegationPrefix},
		{name: "identity store v2 auction prefix", value: l.AuctionPrefix},
	} {
		if !hasIdentityPrefix(item.value, IdentityStoreV2Prefix) {
			return fmt.Errorf("%s must be below %s", item.name, IdentityStoreV2Prefix)
		}
	}
	if l.PrimaryShardKey != IdentityZoneShardKey {
		return fmt.Errorf("identity store v2 primary shard key must be %q", IdentityZoneShardKey)
	}
	if l.LayoutHash != ComputeIdentityStoreV2NameHashLayoutHash(l) {
		return errors.New("identity store v2 layout hash mismatch")
	}
	return nil
}

func (r IdentityShardRoute) ValidateHash() error {
	if r.ZoneID != IdentityZoneID {
		return errors.New("identity shard route must use IDENTITY_ZONE")
	}
	if r.LayoutEpoch == 0 {
		return errors.New("identity shard route layout epoch must be positive")
	}
	if r.ShardCount == 0 {
		return errors.New("identity shard route count must be positive")
	}
	if r.ShardID >= r.ShardCount {
		return errors.New("identity shard route shard id out of range")
	}
	if !IsIdentityShardRoutingMode(r.RoutingMode) {
		return fmt.Errorf("unknown identity shard routing mode %q", r.RoutingMode)
	}
	if err := validateHexHash("identity shard route key", r.RouteKey); err != nil {
		return err
	}
	if r.StateKey == "" {
		return errors.New("identity shard route state key is required")
	}
	if err := validateHexHash("identity shard route hash", r.RouteHash); err != nil {
		return err
	}
	if r.RouteHash != ComputeIdentityShardRouteHash(r) {
		return errors.New("identity shard route hash mismatch")
	}
	return nil
}

func ComputeIdentityShardRouteHash(route IdentityShardRoute) string {
	return identityHash(
		"identity-zone-shard-route-v1",
		route.ZoneID,
		fmt.Sprintf("%020d", route.LayoutEpoch),
		fmt.Sprintf("%010d", route.ShardCount),
		fmt.Sprintf("%010d", route.ShardID),
		string(route.RoutingMode),
		route.RouteKey,
		route.StateKey,
	)
}

func ComputeMsgResolveIdentityHash(msg MsgResolveIdentity) string {
	return identityHash(
		"identity-zone-msg-resolve-v1",
		msg.RequestID,
		msg.Requester,
		msg.SourceZoneID,
		msg.TargetName,
		string(msg.TargetType),
		fmt.Sprintf("%t", msg.ProofRequired),
		msg.ReplyTo,
		fmt.Sprintf("%020d", msg.ExpiryHeight),
	)
}

func ComputeMsgIdentityResolutionResultHash(msg MsgIdentityResolutionResult) string {
	return identityHash(
		"identity-zone-msg-resolution-result-v1",
		msg.RequestID,
		msg.Name,
		string(msg.TargetType),
		msg.ResolvedValue,
		fmt.Sprintf("%020d", msg.ResolverRecordVersion),
		msg.ProofHashOptional,
		string(msg.Status),
		fmt.Sprintf("%020d", msg.ExpiryHeight),
	)
}

func ComputeIdentityReverseLookupProofResponseHash(response IdentityReverseLookupProofResponse) string {
	return identityHash(
		"identity-zone-reverse-proof-response-v1",
		fmt.Sprintf("%020d", response.Height),
		response.AddressHash,
		response.ProofIndex.IndexHash,
		response.Proof.LeafHash,
	)
}

func ComputeIdentityStoreV2NameHashLayoutHash(layout IdentityStoreV2NameHashLayout) string {
	return identityHash(
		"identity-store-v2-name-hash-layout-v1",
		layout.StorePrefix,
		layout.DomainPrefix,
		layout.ResolverPrefix,
		layout.ReversePrefix,
		layout.DelegationPrefix,
		layout.AuctionPrefix,
		layout.PrimaryShardKey,
	)
}

func IsIdentityShardRoutingMode(mode IdentityShardRoutingMode) bool {
	switch mode {
	case IdentityRouteNameHash, IdentityRouteAddress, IdentityRouteAuction:
		return true
	default:
		return false
	}
}

func IsIdentityLookupTargetType(target IdentityLookupTargetType) bool {
	switch target {
	case IdentityLookupTargetResolver, IdentityLookupTargetReverse, IdentityLookupTargetOwner,
		IdentityLookupTargetAccount, IdentityLookupTargetContract, IdentityLookupTargetService,
		IdentityLookupTargetPayment, IdentityLookupTargetMetadata:
		return true
	default:
		return false
	}
}

func IsIdentityResolutionStatus(status IdentityResolutionStatus) bool {
	switch status {
	case IdentityResolutionStatusResolved, IdentityResolutionStatusNotFound, IdentityResolutionStatusExpired,
		IdentityResolutionStatusUnauthorized, IdentityResolutionStatusFailed, IdentityResolutionStatusRejected:
		return true
	default:
		return false
	}
}

func routeIdentityStateKey(mode IdentityShardRoutingMode, routeKey string, stateKey string, shardCount uint32, layoutEpoch uint64) (IdentityShardRoute, error) {
	if shardCount == 0 {
		return IdentityShardRoute{}, errors.New("identity shard count must be positive")
	}
	if layoutEpoch == 0 {
		return IdentityShardRoute{}, errors.New("identity shard layout epoch must be positive")
	}
	if err := validateHexHash("identity shard route key", routeKey); err != nil {
		return IdentityShardRoute{}, err
	}
	if stateKey == "" {
		return IdentityShardRoute{}, errors.New("identity shard state key is required")
	}
	hash := identityHash("identity-zone-route-key-v1", string(mode), routeKey, fmt.Sprintf("%020d", layoutEpoch))
	bytes, err := hex.DecodeString(hash[:16])
	if err != nil {
		return IdentityShardRoute{}, err
	}
	route := IdentityShardRoute{
		ZoneID:		IdentityZoneID,
		LayoutEpoch:	layoutEpoch,
		ShardCount:	shardCount,
		ShardID:	uint32(binary.BigEndian.Uint64(bytes) % uint64(shardCount)),
		RoutingMode:	mode,
		RouteKey:	routeKey,
		StateKey:	stateKey,
	}
	route.RouteHash = ComputeIdentityShardRouteHash(route)
	return route, route.ValidateHash()
}

func hasIdentityPrefix(value string, prefix string) bool {
	return value == prefix || len(value) > len(prefix) && value[:len(prefix)+1] == prefix+"/"
}
